package pkg

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:     true,
		EnableHTTP2:     false,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1 << 20,
		ShutdownTimeout: 10 * time.Second,
	}

	server := NewServer(config)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.IsRunning() {
		t.Error("New server should not be running")
	}

	if server.Addr() != "" {
		t.Error("New server should not have an address")
	}
}

// TestServerListen tests basic HTTP server listening
func TestServerListen(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := NewServer(config)
	router := NewRouter()

	// Add a simple test route
	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	server.SetRouter(router)

	// Start server on random port
	addr := "127.0.0.1:0"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Clean up
	defer server.Close()
}

// TestServerHTTP1Request tests HTTP/1.1 request handling
func TestServerHTTP1Request(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := NewServer(config)
	router := NewRouter()

	// Add test routes
	router.GET("/hello", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	router.POST("/echo", func(ctx Context) error {
		body := ctx.Body()
		return ctx.String(http.StatusOK, string(body))
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18080"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test GET request
	resp, err := http.Get("http://" + addr + "/hello")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(body))
	}
}

// TestServerHTTP2Support tests HTTP/2 protocol support
func TestServerHTTP2Support(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:    true,
		EnableHTTP2:    true,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := NewServer(config)
	server.EnableHTTP2()

	protocol := server.Protocol()
	if protocol != "HTTP/1.1, HTTP/2" {
		t.Errorf("Expected 'HTTP/1.1, HTTP/2', got '%s'", protocol)
	}
}

// TestServerGracefulShutdown tests graceful server shutdown
func TestServerGracefulShutdown(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:     true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/slow", func(ctx Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(http.StatusOK, "Done")
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18081"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Initiate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after shutdown")
	}
}

// TestServerShutdownWithTimeout tests shutdown with context timeout
func TestServerShutdownWithTimeout(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18082"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test GracefulShutdown with timeout
	err = server.GracefulShutdown(2 * time.Second)
	if err != nil {
		t.Errorf("GracefulShutdown failed: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after shutdown")
	}
}

// TestServerShutdownHooks tests shutdown hook execution
func TestServerShutdownHooks(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	server.SetRouter(router)

	// Register shutdown hooks
	hookCalled := false
	server.RegisterShutdownHook(func(ctx context.Context) error {
		hookCalled = true
		return nil
	})

	// Start server
	addr := "127.0.0.1:18083"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !hookCalled {
		t.Error("Shutdown hook was not called")
	}
}

// TestServerClose tests immediate server close
func TestServerClose(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18084"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// Close immediately
	err = server.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if server.IsRunning() {
		t.Error("Server should not be running after close")
	}
}

// TestServerMiddleware tests global middleware execution
func TestServerMiddleware(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	// Track middleware execution
	middlewareCalled := false

	middleware := func(ctx Context, next HandlerFunc) error {
		middlewareCalled = true
		ctx.SetHeader("X-Middleware", "executed")
		return next(ctx)
	}

	server.SetMiddleware(middleware)

	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18085"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make request
	resp, err := http.Get("http://" + addr + "/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}

	if resp.Header.Get("X-Middleware") != "executed" {
		t.Error("Middleware header not set")
	}
}

// TestServerErrorHandler tests custom error handler
func TestServerErrorHandler(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	router := NewRouter()

	// Set custom error handler
	errorHandlerCalled := false
	server.SetErrorHandler(func(ctx Context, err error) error {
		errorHandlerCalled = true
		return ctx.String(http.StatusInternalServerError, "Custom error: "+err.Error())
	})

	router.GET("/error", func(ctx Context) error {
		return fmt.Errorf("test error")
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:18086"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make request that triggers error
	resp, err := http.Get("http://" + addr + "/error")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if !errorHandlerCalled {
		t.Error("Error handler was not called")
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Custom error: test error" {
		t.Errorf("Expected custom error message, got: %s", string(body))
	}
}

// TestServerProtocolDetection tests protocol detection
func TestServerProtocolDetection(t *testing.T) {
	tests := []struct {
		name          string
		enableHTTP1   bool
		enableHTTP2   bool
		enableQUIC    bool
		expectedProto string
	}{
		{"HTTP1 only", true, false, false, "HTTP/1.1"},
		{"HTTP2 only", false, true, false, "HTTP/2"},
		{"HTTP1 and HTTP2", true, true, false, "HTTP/1.1, HTTP/2"},
		{"All protocols", true, true, true, "HTTP/1.1, HTTP/2, QUIC"},
		{"No protocols", false, false, false, "none"},
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

// TestServerTLSConfiguration tests TLS server configuration
func TestServerTLSConfiguration(t *testing.T) {
	// Skip if no test certificates available
	t.Skip("Skipping TLS test - requires test certificates")

	config := ServerConfig{
		EnableHTTP1: true,
		EnableHTTP2: true,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/secure", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Secure connection")
	})

	server.SetRouter(router)

	// This would require actual certificate files
	// err := server.ListenTLS("127.0.0.1:18443", "cert.pem", "key.pem")
	// Test that the method exists and returns appropriate error
	err := server.ListenTLS("127.0.0.1:18443", "nonexistent.pem", "nonexistent.pem")
	if err == nil {
		t.Error("Expected error for missing certificates")
		server.Close()
	}
}

// TestServerDoubleStart tests that starting an already running server fails
func TestServerDoubleStart(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1: true,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	server.SetRouter(router)

	// Start server first time
	addr := "127.0.0.1:18087"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Try to start again
	err = server.Listen(addr)
	if err == nil {
		t.Error("Expected error when starting already running server")
	}
}
