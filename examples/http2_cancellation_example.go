package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	// Create framework configuration with HTTP/2 enabled for cancellation support
	config := pkg.FrameworkConfig{
		// Server configuration - HTTP/2 is required for proper cancellation support
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     30 * time.Second, // Longer timeout for long-running operations
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1 MB
			EnableHTTP1:     true,    // Enable HTTP/1.1 for compatibility
			EnableHTTP2:     true,    // Enable HTTP/2 for cancellation features
			ShutdownTimeout: 10 * time.Second,
		},
		// Database configuration - for database cancellation examples
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "http2_cancellation.db",
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
	// Global Middleware
	// ============================================================================
	// Add cancellation middleware globally - monitors all requests for cancellation
	// This middleware checks if the client has cancelled the request and stops processing
	app.Use(pkg.CancellationMiddleware())

	// ============================================================================
	// Custom Error Handler
	// ============================================================================
	// Set error handler that properly handles cancellation errors
	app.SetErrorHandler(func(ctx pkg.Context, err error) error {
		// Check if the error is due to request cancellation
		if pkg.IsCancellationError(err) {
			log.Printf("🚫 Request cancelled: %s %s", ctx.Request().Method, ctx.Request().RequestURI)
			// Don't send response for cancelled requests - client is already gone
			return nil
		}

		// Handle other errors normally
		log.Printf("❌ Error: %v", err)
		return ctx.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})
	})

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Example 1: Simple handler that respects cancellation
	// This demonstrates basic cancellation checking before processing
	router.GET("/api/simple", simpleHandler)

	// Example 2: Long-running operation with periodic cancellation checks
	// This demonstrates how to check for cancellation during lengthy operations
	router.GET("/api/long-running", longRunningHandler)

	// Example 3: Database query with timeout and cancellation
	// This demonstrates combining timeouts with cancellation for database operations
	router.GET("/api/database", databaseHandler)

	// Example 4: Streaming response with cancellation
	// This demonstrates server-sent events that respect client cancellation
	router.GET("/api/stream", streamHandler)

	// Example 5: Batch processing with cancellation
	// This demonstrates processing multiple items with cancellation support
	router.POST("/api/batch", batchHandler)

	// Example 6: Cleanup on cancellation
	// This demonstrates proper resource cleanup when a request is cancelled
	router.GET("/api/cleanup", cleanupHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - HTTP/2 Request Cancellation Example")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("This example demonstrates HTTP/2 request cancellation handling.")
	fmt.Println("When a client cancels a request (e.g., by pressing Ctrl+C), the")
	fmt.Println("server detects it and stops processing to save resources.")
	fmt.Println()
	fmt.Println("Server listening on: https://localhost:8443")
	fmt.Println()
	fmt.Println("⚠️  TLS Certificate Required:")
	fmt.Println("   Generate self-signed certificates with:")
	fmt.Println("   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj '/CN=localhost'")
	fmt.Println()
	fmt.Println("Try these commands (press Ctrl+C to cancel):")
	fmt.Println("  curl -k https://localhost:8443/api/simple")
	fmt.Println("  curl -k https://localhost:8443/api/long-running")
	fmt.Println("  curl -k https://localhost:8443/api/database")
	fmt.Println("  curl -k https://localhost:8443/api/stream")
	fmt.Println("  curl -k -X POST https://localhost:8443/api/batch")
	fmt.Println("  curl -k https://localhost:8443/api/cleanup")
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println()

	// Start the server with TLS (required for HTTP/2)
	if err := app.ListenTLS(":8443", "cert.pem", "key.pem"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions
// ============================================================================

// simpleHandler demonstrates basic cancellation checking
func simpleHandler(ctx pkg.Context) error {
	// Check if request is already cancelled before processing
	select {
	case <-ctx.Context().Done():
		log.Println("⚠️  Request was already cancelled")
		return ctx.Context().Err()
	default:
	}

	log.Println("✅ Processing simple request")
	return ctx.JSON(200, map[string]interface{}{
		"message": "Simple response - request completed successfully",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// longRunningHandler demonstrates periodic cancellation checks during long operations
func longRunningHandler(ctx pkg.Context) error {
	log.Println("🔄 Starting long-running operation...")

	// Simulate a long-running operation with 10 iterations
	for i := 0; i < 10; i++ {
		// Check for cancellation before each iteration
		select {
		case <-ctx.Context().Done():
			log.Printf("🚫 Long-running operation cancelled at iteration %d", i)
			return ctx.Context().Err()
		default:
		}

		// Simulate work
		time.Sleep(500 * time.Millisecond)
		log.Printf("   Processing iteration %d/10", i+1)
	}

	log.Println("✅ Long-running operation completed")
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Long operation completed successfully",
		"iterations": 10,
		"duration":   "5 seconds",
	})
}

// databaseHandler demonstrates database queries with timeout and cancellation
func databaseHandler(ctx pkg.Context) error {
	log.Println("🗄️  Starting database query...")

	// Create a timeout context for the database query
	// This ensures the query doesn't run forever
	queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	// Simulate database query in a goroutine
	resultChan := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		// Simulate slow database query
		time.Sleep(2 * time.Second)

		// Check if context is cancelled before returning result
		select {
		case <-queryCtx.Done():
			errChan <- queryCtx.Err()
			return
		default:
		}

		// Return query result
		resultChan <- map[string]interface{}{
			"id":        123,
			"name":      "Example Record",
			"timestamp": time.Now().Format(time.RFC3339),
		}
	}()

	// Wait for result, cancellation, or timeout
	select {
	case result := <-resultChan:
		log.Println("✅ Database query completed")
		return ctx.JSON(200, result)

	case err := <-errChan:
		if pkg.IsCancellationError(err) {
			log.Println("🚫 Database query cancelled")
			return ctx.JSON(499, map[string]interface{}{
				"error": "Request cancelled by client",
			})
		}
		log.Printf("❌ Database query error: %v", err)
		return ctx.JSON(500, map[string]interface{}{
			"error": err.Error(),
		})

	case <-queryCtx.Done():
		log.Println("⏱️  Database query timeout")
		return ctx.JSON(408, map[string]interface{}{
			"error": "Request timeout - query took too long",
		})
	}
}

// streamHandler demonstrates server-sent events with cancellation
func streamHandler(ctx pkg.Context) error {
	log.Println("📡 Starting event stream...")

	// Set headers for server-sent events
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("Connection", "keep-alive")

	// Send events periodically
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Context().Done():
			log.Printf("🚫 Stream cancelled after %d events", i)
			return ctx.Context().Err()

		case <-ticker.C:
			// Format as server-sent event
			event := fmt.Sprintf("data: {\"event\": %d, \"message\": \"Event %d\", \"time\": \"%s\"}\n\n",
				i, i, time.Now().Format(time.RFC3339))

			// Write event to response
			if _, err := ctx.Response().Write([]byte(event)); err != nil {
				log.Printf("❌ Error writing event: %v", err)
				return err
			}

			// Flush to ensure event is sent immediately
			ctx.Response().Flush()
			log.Printf("   Sent event %d/30", i+1)
		}
	}

	log.Println("✅ Stream completed")
	return nil
}

// batchHandler demonstrates batch processing with cancellation
func batchHandler(ctx pkg.Context) error {
	log.Println("📦 Starting batch processing...")

	// Simulate batch of items to process
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	processed := []string{}

	for i, item := range items {
		// Check for cancellation before processing each item
		select {
		case <-ctx.Context().Done():
			log.Printf("🚫 Batch processing cancelled after %d items", i)
			return ctx.JSON(499, map[string]interface{}{
				"error":     "Request cancelled by client",
				"processed": processed,
				"remaining": len(items) - i,
			})
		default:
		}

		// Simulate processing time
		time.Sleep(500 * time.Millisecond)
		processed = append(processed, item)
		log.Printf("   Processed: %s (%d/%d)", item, i+1, len(items))
	}

	log.Println("✅ Batch processing completed")
	return ctx.JSON(200, map[string]interface{}{
		"message":   "Batch processing completed successfully",
		"processed": processed,
		"total":     len(items),
	})
}

// cleanupHandler demonstrates proper resource cleanup on cancellation
func cleanupHandler(ctx pkg.Context) error {
	log.Println("🧹 Starting operation with cleanup...")

	// Simulate acquiring a resource
	resource := acquireResource()
	log.Println("   Resource acquired")

	// Ensure cleanup happens even if cancelled
	defer func() {
		releaseResource(resource)
		log.Println("   Resource released (cleanup completed)")
	}()

	// Simulate long operation
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Context().Done():
			log.Printf("🚫 Operation cancelled at iteration %d - cleanup will still run", i)
			return ctx.Context().Err()
		default:
		}

		time.Sleep(500 * time.Millisecond)
		log.Printf("   Working with resource: iteration %d/10", i+1)
	}

	log.Println("✅ Operation completed successfully")
	return ctx.JSON(200, map[string]interface{}{
		"message":  "Operation completed with proper cleanup",
		"resource": resource,
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

// acquireResource simulates acquiring a resource (e.g., file handle, connection)
func acquireResource() string {
	return fmt.Sprintf("resource-%d", time.Now().Unix())
}

// releaseResource simulates releasing a resource
func releaseResource(resource string) {
	// In a real application, this would close files, connections, etc.
	log.Printf("   Releasing %s", resource)
}
