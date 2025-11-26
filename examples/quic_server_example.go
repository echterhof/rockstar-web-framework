package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// TLS Certificate Setup
	// ============================================================================
	// Ensure TLS certificates exist (required for QUIC)
	certFile := "cert.pem"
	keyFile := "key.pem"

	if err := ensureCertificates(certFile, keyFile); err != nil {
		log.Fatalf("Failed to setup TLS certificates: %v", err)
	}

	// ============================================================================
	// Configuration Setup
	// ============================================================================
	// Create framework configuration with QUIC/HTTP3 enabled
	// QUIC is a modern transport protocol that provides improved performance
	// over traditional TCP, with built-in encryption and multiplexing
	config := pkg.FrameworkConfig{
		// Server configuration - QUIC requires specific settings
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second, // QUIC connections can be idle longer
			EnableHTTP2:  true,              // Enable HTTP/2 for initial connection
			EnableQUIC:   true,              // Enable QUIC/HTTP3 protocol
		},
		// Database configuration - for demonstration purposes
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "quic_server.db",
		},
		// Cache configuration
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    50 * 1024 * 1024, // 50 MB
			DefaultTTL: 5 * time.Minute,
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	// Create new framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ============================================================================
	// Lifecycle Hooks
	// ============================================================================
	// Register startup hook
	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 QUIC server starting up...")
		fmt.Println("   Protocol: HTTP/3 over QUIC")
		fmt.Println("   Transport: UDP (not TCP)")
		return nil
	})

	// Register shutdown hook
	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 QUIC server shutting down...")
		return nil
	})

	// ============================================================================
	// Global Middleware
	// ============================================================================
	// Add Alt-Svc header to advertise HTTP/3 availability
	app.Use(altSvcMiddleware)

	// Add logging middleware to track requests
	app.Use(loggingMiddleware)

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Root endpoint - welcome message
	router.GET("/", homeHandler)

	// Parameterized route - demonstrates path parameters
	router.GET("/hello/:name", helloHandler)

	// Echo endpoint - demonstrates request body handling
	router.POST("/echo", echoHandler)

	// JSON API endpoint - demonstrates JSON responses
	router.GET("/api/data", dataHandler)

	// Performance test endpoint - demonstrates QUIC's speed advantages
	router.GET("/api/performance", performanceHandler)

	// Stream endpoint - demonstrates QUIC's multiplexing capabilities
	router.GET("/api/stream", streamHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - QUIC/HTTP3 Server Example")
	fmt.Println("======================================================")
	fmt.Println()
	fmt.Println("This example demonstrates QUIC/HTTP3 protocol support.")
	fmt.Println("QUIC provides:")
	fmt.Println("  • Faster connection establishment (0-RTT)")
	fmt.Println("  • Better performance on lossy networks")
	fmt.Println("  • Built-in encryption (TLS 1.3)")
	fmt.Println("  • Improved multiplexing (no head-of-line blocking)")
	fmt.Println()
	fmt.Println("Servers running:")
	fmt.Println("  HTTP/2: https://localhost:8443 (TCP)")
	fmt.Println("  HTTP/3: https://localhost:4433 (UDP/QUIC)")
	fmt.Println()
	fmt.Println("🔐 TLS Certificates: cert.pem, key.pem")
	fmt.Println("   (Auto-generated self-signed certificates for development)")
	fmt.Println()
	fmt.Println("📝 How to test:")
	fmt.Println("   Browser: https://localhost:8443/")
	fmt.Println("   (Browser will auto-upgrade to HTTP/3 via Alt-Svc header)")
	fmt.Println()
	fmt.Println("   curl (HTTP/2): curl -k https://localhost:8443/")
	fmt.Println("   curl (HTTP/3): curl --http3 -k https://localhost:4433/")
	fmt.Println("   (HTTP/3 requires curl 7.66+ with QUIC support)")
	fmt.Println()
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  /                    - Welcome message")
	fmt.Println("  GET  /hello/:name         - Personalized greeting")
	fmt.Println("  POST /echo                - Echo request body")
	fmt.Println("  GET  /api/data            - JSON data response")
	fmt.Println("  GET  /api/performance     - Performance test")
	fmt.Println("  GET  /api/stream          - Streaming response")
	fmt.Println()
	fmt.Println("======================================================")
	fmt.Println()

	// Start HTTP/2 server in a goroutine
	// This allows browsers to discover HTTP/3 via Alt-Svc header
	go func() {
		if err := app.ListenTLS(":8443", certFile, keyFile); err != nil {
			log.Printf("HTTP/2 server error: %v", err)
		}
	}()

	// Give HTTP/2 server time to start
	time.Sleep(500 * time.Millisecond)

	// Start the QUIC server with TLS (blocking)
	// QUIC always uses TLS 1.3 for encryption
	if err := app.ListenQUIC(":4433", certFile, keyFile); err != nil {
		log.Fatalf("QUIC server error: %v", err)
	}
}

// ============================================================================
// Middleware Functions
// ============================================================================

// altSvcMiddleware adds Alt-Svc header to advertise HTTP/3 availability
func altSvcMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// Advertise HTTP/3 on port 4433
	ctx.SetHeader("Alt-Svc", `h3=":4433"; ma=86400`)
	return next(ctx)
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()

	// Detect protocol
	protocol := ctx.Request().Protocol
	if protocol == "" {
		protocol = ctx.Request().Proto
	}

	// Log request
	fmt.Printf("[%s] %s %s (%s)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		ctx.Request().Method,
		ctx.Request().RequestURI,
		protocol,
	)

	// Call next handler
	err := next(ctx)

	// Log completion time
	duration := time.Since(start)
	fmt.Printf("  ⏱️  Completed in %v\n", duration)

	return err
}

// ============================================================================
// Handler Functions
// ============================================================================

// homeHandler handles the root endpoint
func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":  "Welcome to QUIC/HTTP3 Server! 🚀",
		"protocol": "HTTP/3 over QUIC",
		"features": []string{
			"0-RTT connection establishment",
			"No head-of-line blocking",
			"Built-in TLS 1.3 encryption",
			"Better performance on lossy networks",
			"Connection migration support",
		},
		"endpoints": []string{
			"GET /",
			"GET /hello/:name",
			"POST /echo",
			"GET /api/data",
			"GET /api/performance",
			"GET /api/stream",
		},
	})
}

// helloHandler demonstrates path parameters with QUIC
func helloHandler(ctx pkg.Context) error {
	// Extract name parameter from URL
	name := ctx.Params()["name"]

	return ctx.JSON(200, map[string]interface{}{
		"message":  fmt.Sprintf("Hello, %s! 👋", name),
		"protocol": "HTTP/3 over QUIC",
		"note":     "This response was delivered over QUIC, which is faster than traditional HTTP!",
	})
}

// echoHandler demonstrates request body handling
func echoHandler(ctx pkg.Context) error {
	// Get request body
	body := ctx.Body()

	// Echo it back
	return ctx.JSON(200, map[string]interface{}{
		"message":      "Echo response",
		"received":     string(body),
		"content_type": ctx.GetHeader("Content-Type"),
		"protocol":     "HTTP/3 over QUIC",
	})
}

// dataHandler demonstrates JSON API responses
func dataHandler(ctx pkg.Context) error {
	// Simulate fetching data
	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"},
			{"id": 3, "name": "Charlie", "role": "user"},
		},
		"total":     3,
		"timestamp": time.Now().Format(time.RFC3339),
		"protocol":  "HTTP/3 over QUIC",
	}

	return ctx.JSON(200, data)
}

// performanceHandler demonstrates QUIC's performance advantages
func performanceHandler(ctx pkg.Context) error {
	start := time.Now()

	// Simulate some processing
	time.Sleep(10 * time.Millisecond)

	// Calculate response time
	duration := time.Since(start)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Performance test completed",
		"metrics": map[string]interface{}{
			"processing_time_ms": duration.Milliseconds(),
			"protocol":           "HTTP/3 over QUIC",
			"advantages": []string{
				"Faster connection establishment (0-RTT)",
				"No TCP handshake overhead",
				"Better handling of packet loss",
				"Connection migration (switch networks seamlessly)",
			},
		},
		"note": "QUIC typically shows 2-3x faster connection times compared to TCP",
	})
}

// streamHandler demonstrates QUIC's multiplexing capabilities
func streamHandler(ctx pkg.Context) error {
	// Set headers for streaming response
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("Connection", "keep-alive")

	// Send multiple events
	// QUIC's multiplexing means these won't block each other
	for i := 0; i < 10; i++ {
		// Check for cancellation
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
		}

		// Format event
		event := fmt.Sprintf("data: {\"event\": %d, \"message\": \"Stream event %d\", \"protocol\": \"QUIC\", \"time\": \"%s\"}\n\n",
			i, i, time.Now().Format(time.RFC3339))

		// Write event
		if _, err := ctx.Response().Write([]byte(event)); err != nil {
			return err
		}

		// Flush immediately - QUIC handles this efficiently
		ctx.Response().Flush()

		// Small delay between events
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

// ============================================================================
// Certificate Management
// ============================================================================

// ensureCertificates checks if TLS certificates exist, and generates them if missing
func ensureCertificates(certFile, keyFile string) error {
	// Check if both files exist
	if fileExists(certFile) && fileExists(keyFile) {
		return nil
	}

	fmt.Println("🔐 Generating self-signed TLS certificates...")
	fmt.Println("   (Required for QUIC/HTTP3)")
	fmt.Println()

	// Generate certificates
	if err := generateSelfSignedCert(certFile, keyFile); err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	fmt.Println("✅ Certificates generated successfully!")
	fmt.Println("   Certificate: " + certFile)
	fmt.Println("   Private Key: " + keyFile)
	fmt.Println()
	fmt.Println("⚠️  These are self-signed certificates for development only.")
	fmt.Println("   Do NOT use in production!")
	fmt.Println()

	return nil
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// generateSelfSignedCert creates a self-signed certificate for localhost
func generateSelfSignedCert(certFile, keyFile string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key to file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
