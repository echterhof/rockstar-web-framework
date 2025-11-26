package pkg

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// generateTestCertificates generates self-signed certificates for testing
func generateTestCertificates(t *testing.T) (certFile, keyFile string) {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Rockstar Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "127.0.0.1"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Write certificate to file
	certFile = "test_cert.pem"
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("Failed to write cert: %v", err)
	}

	// Write private key to file
	keyFile = "test_key.pem"
	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatalf("Failed to write key: %v", err)
	}

	return certFile, keyFile
}

// cleanupTestCertificates removes test certificate files
func cleanupTestCertificates(certFile, keyFile string) {
	os.Remove(certFile)
	os.Remove(keyFile)
}

// TestServerQUICBasic tests basic QUIC server creation and startup
func TestServerQUICBasic(t *testing.T) {
	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	if server.Protocol() != "QUIC" {
		t.Errorf("Expected QUIC protocol to be enabled, got: %s", server.Protocol())
	}
}

// TestServerQUICListen tests QUIC server listening
func TestServerQUICListen(t *testing.T) {
	// Generate test certificates
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "QUIC Hello!")
	})
	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14433"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	if !server.IsRunning() {
		t.Error("QUIC server should be running")
	}

	if server.Addr() != addr {
		t.Errorf("Expected address %s, got %s", addr, server.Addr())
	}
}

// TestServerQUICRequest tests making HTTP/3 requests to QUIC server
func TestServerQUICRequest(t *testing.T) {
	// Generate test certificates
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/hello", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Hello from QUIC!")
	})

	router.POST("/echo", func(ctx Context) error {
		body := ctx.Body()
		return ctx.String(http.StatusOK, string(body))
	})

	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14434"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	// Create HTTP/3 client with custom TLS config to accept self-signed cert
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	roundTripper := &http3.Transport{
		TLSClientConfig: tlsConfig,
	}
	defer roundTripper.Close()

	client := &http.Client{
		Transport: roundTripper,
		Timeout:   3 * time.Second,
	}

	// Test GET request
	resp, err := client.Get("https://" + addr + "/hello")
	if err != nil {
		t.Fatalf("Failed to make QUIC request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Hello from QUIC!" {
		t.Errorf("Expected 'Hello from QUIC!', got '%s'", string(body))
	}

	// Verify HTTP/3 protocol
	if resp.Proto != "HTTP/3.0" {
		t.Errorf("Expected HTTP/3.0 protocol, got %s", resp.Proto)
	}
}

// TestServerQUICWithoutEnabling tests that QUIC fails when not enabled
func TestServerQUICWithoutEnabling(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC: false,
	}

	server := NewServer(config)
	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Try to start QUIC server without enabling it
	err := server.ListenQUIC("127.0.0.1:14435", certFile, keyFile)
	if err == nil {
		server.Close()
		t.Error("Expected error when starting QUIC without enabling it")
	}

	if err.Error() != "QUIC protocol is not enabled" {
		t.Errorf("Expected 'QUIC protocol is not enabled' error, got: %v", err)
	}
}

// TestServerQUICInvalidCertificates tests QUIC with invalid certificates
func TestServerQUICInvalidCertificates(t *testing.T) {
	config := ServerConfig{
		EnableQUIC: true,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Try to start with non-existent certificates
	err := server.ListenQUIC("127.0.0.1:14436", "nonexistent.pem", "nonexistent.pem")
	if err == nil {
		server.Close()
		t.Error("Expected error for invalid certificates")
	}
}

// TestServerQUICShutdown tests graceful QUIC server shutdown
func TestServerQUICShutdown(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:      true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14437"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after shutdown")
	}
}

// TestServerQUICClose tests immediate QUIC server close
func TestServerQUICClose(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14438"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Close immediately
	err := server.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after close")
	}
}

// TestServerQUICTLSConfig tests QUIC with custom TLS configuration
func TestServerQUICTLSConfig(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	// Load certificate for custom TLS config
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to load certificates: %v", err)
	}

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
			NextProtos:   []string{"h3", "h3-29"},
		},
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Custom TLS Config")
	})
	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14439"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	if !server.IsRunning() {
		t.Error("Server should be running")
	}
}

// TestServerQUICDoubleStart tests that starting an already running QUIC server fails
func TestServerQUICDoubleStart(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC: true,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Start server first time in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14440"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start server: %v", err)
	default:
		// Server started successfully
	}

	// Try to start again
	err := server.ListenQUIC(addr, certFile, keyFile)
	if err == nil {
		t.Error("Expected error when starting already running server")
	}

	if err.Error() != "server is already running" {
		t.Errorf("Expected 'server is already running' error, got: %v", err)
	}
}

// TestServerQUICProtocolDetection tests protocol detection with QUIC enabled
func TestServerQUICProtocolDetection(t *testing.T) {
	tests := []struct {
		name          string
		enableHTTP1   bool
		enableHTTP2   bool
		enableQUIC    bool
		expectedProto string
	}{
		{"QUIC only", false, false, true, "QUIC"},
		{"HTTP1 and QUIC", true, false, true, "HTTP/1.1, QUIC"},
		{"HTTP2 and QUIC", false, true, true, "HTTP/2, QUIC"},
		{"All protocols", true, true, true, "HTTP/1.1, HTTP/2, QUIC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ServerConfig{
				EnableHTTP1: tt.enableHTTP1,
				EnableHTTP2: tt.enableHTTP2,
				EnableQUIC:  tt.enableQUIC,
			}

			server := NewServer(config)

			if tt.enableHTTP1 {
				server.EnableHTTP1()
			}
			if tt.enableHTTP2 {
				server.EnableHTTP2()
			}
			if tt.enableQUIC {
				server.EnableQUIC()
			}

			protocol := server.Protocol()
			if protocol != tt.expectedProto {
				t.Errorf("Expected protocol '%s', got '%s'", tt.expectedProto, protocol)
			}
		})
	}
}

// TestServerQUICShutdownHooks tests shutdown hook execution with QUIC
func TestServerQUICShutdownHooks(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})
	server.SetRouter(router)

	// Register shutdown hook
	hookCalled := false
	server.RegisterShutdownHook(func(ctx context.Context) error {
		hookCalled = true
		return nil
	})

	// Start server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14441"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start server: %v", err)
	default:
		// Server started successfully
	}

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !hookCalled {
		t.Error("Shutdown hook was not called")
	}
}

// TestServerQUICConcurrentRequests tests handling multiple concurrent QUIC requests
func TestServerQUICConcurrentRequests(t *testing.T) {
	certFile, keyFile := generateTestCertificates(t)
	defer cleanupTestCertificates(certFile, keyFile)

	config := ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := NewServer(config)
	server.EnableQUIC()

	router := NewRouter()
	router.GET("/concurrent", func(ctx Context) error {
		time.Sleep(50 * time.Millisecond)
		return ctx.String(http.StatusOK, "Concurrent response")
	})
	server.SetRouter(router)

	// Start QUIC server in goroutine (ListenQUIC blocks)
	addr := "127.0.0.1:14442"
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenQUIC(addr, certFile, keyFile)
	}()
	defer server.Close()

	time.Sleep(200 * time.Millisecond)

	// Check if server failed to start
	select {
	case err := <-errChan:
		t.Fatalf("Failed to start QUIC server: %v", err)
	default:
		// Server started successfully
	}

	// Create HTTP/3 client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	roundTripper := &http3.Transport{
		TLSClientConfig: tlsConfig,
	}
	defer roundTripper.Close()

	client := &http.Client{
		Transport: roundTripper,
		Timeout:   3 * time.Second,
	}

	// Make concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("panic in request %d: %v", id, r)
					done <- false
				}
			}()

			resp, err := client.Get(fmt.Sprintf("https://%s/concurrent", addr))
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %v", id, err)
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete with timeout
	successCount := 0
	timeout := time.After(10 * time.Second)
	for i := 0; i < numRequests; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			}
		case <-timeout:
			t.Fatalf("Test timed out waiting for requests to complete (%d/%d completed)", i, numRequests)
		}
	}

	// Report any errors
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}
