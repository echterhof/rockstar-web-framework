package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// Example demonstrating QUIC/HTTP3 server usage
func main() {
	// Create server configuration with QUIC enabled
	config := pkg.ServerConfig{
		EnableQUIC:   true,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create new server
	server := pkg.NewServer(config)
	server.EnableQUIC()

	// Create router
	router := pkg.NewRouter()

	// Add routes
	router.GET("/", func(ctx pkg.Context) error {
		return ctx.String(http.StatusOK, "Welcome to QUIC/HTTP3 Server!")
	})

	router.GET("/hello/:name", func(ctx pkg.Context) error {
		name := ctx.Params()["name"]
		return ctx.String(http.StatusOK, fmt.Sprintf("Hello, %s via QUIC!", name))
	})

	router.POST("/echo", func(ctx pkg.Context) error {
		body := ctx.Body()
		return ctx.String(http.StatusOK, string(body))
	})

	// Set router
	server.SetRouter(router)

	// Start QUIC server
	// Note: You need to provide valid TLS certificates
	addr := "0.0.0.0:4433"
	certFile := "cert.pem"
	keyFile := "key.pem"

	fmt.Printf("Starting QUIC/HTTP3 server on %s\n", addr)
	fmt.Println("Protocol: HTTP/3 over QUIC")
	fmt.Println("Make sure you have valid TLS certificates!")

	if err := server.ListenQUIC(addr, certFile, keyFile); err != nil {
		log.Fatalf("Failed to start QUIC server: %v", err)
	}

	// Keep server running
	select {}
}
