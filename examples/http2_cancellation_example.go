package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create server with HTTP/2 enabled
	config := pkg.ServerConfig{
		EnableHTTP1:     true,
		EnableHTTP2:     true,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1 << 20,
		ShutdownTimeout: 10 * time.Second,
	}

	server := pkg.NewServer(config)
	server.EnableHTTP2()

	// Create router
	router := pkg.NewRouter()

	// Add cancellation middleware globally
	server.SetMiddleware(pkg.CancellationMiddleware())

	// Example 1: Simple handler that respects cancellation
	router.GET("/api/simple", func(ctx pkg.Context) error {
		// Check if request is cancelled
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
		}

		return ctx.JSON(200, map[string]string{
			"message": "Simple response",
		})
	})

	// Example 2: Long-running operation with cancellation checks
	router.GET("/api/long-running", pkg.WithCancellationCheck(func(ctx pkg.Context) error {
		// Simulate a long-running operation
		for i := 0; i < 10; i++ {
			// Check for cancellation periodically
			select {
			case <-ctx.Context().Done():
				log.Printf("Request cancelled at iteration %d", i)
				return ctx.Context().Err()
			default:
			}

			// Simulate work
			time.Sleep(500 * time.Millisecond)
			log.Printf("Processing iteration %d", i)
		}

		return ctx.JSON(200, map[string]string{
			"message": "Long operation completed",
		})
	}))

	// Example 3: Database query with timeout and cancellation
	router.GET("/api/database", func(ctx pkg.Context) error {
		// Create a timeout context
		queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
		defer cancel()

		// Simulate database query
		resultChan := make(chan map[string]interface{}, 1)
		errChan := make(chan error, 1)

		go func() {
			// Simulate slow database query
			time.Sleep(2 * time.Second)

			// Check if context is cancelled
			select {
			case <-queryCtx.Done():
				errChan <- queryCtx.Err()
				return
			default:
			}

			resultChan <- map[string]interface{}{
				"id":   123,
				"name": "Example Record",
			}
		}()

		// Wait for result or cancellation
		select {
		case result := <-resultChan:
			return ctx.JSON(200, result)
		case err := <-errChan:
			if pkg.IsCancellationError(err) {
				log.Printf("Database query cancelled: %v", err)
				return ctx.JSON(499, map[string]string{
					"error": "Request cancelled",
				})
			}
			return ctx.JSON(500, map[string]string{
				"error": err.Error(),
			})
		case <-queryCtx.Done():
			return ctx.JSON(408, map[string]string{
				"error": "Request timeout",
			})
		}
	})

	// Example 4: Streaming response with cancellation
	router.GET("/api/stream", func(ctx pkg.Context) error {
		ctx.SetHeader("Content-Type", "text/event-stream")
		ctx.SetHeader("Cache-Control", "no-cache")
		ctx.SetHeader("Connection", "keep-alive")

		// Send events periodically
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for i := 0; i < 30; i++ {
			select {
			case <-ctx.Context().Done():
				log.Printf("Stream cancelled after %d events", i)
				return ctx.Context().Err()
			case <-ticker.C:
				event := fmt.Sprintf("data: Event %d\n\n", i)
				if _, err := ctx.Response().Write([]byte(event)); err != nil {
					log.Printf("Error writing event: %v", err)
					return err
				}
				ctx.Response().Flush()
			}
		}

		return nil
	})

	// Example 5: Batch processing with cancellation
	router.POST("/api/batch", pkg.CancellationAwareHandler(func(ctx pkg.Context) error {
		// Simulate batch processing
		items := []string{"item1", "item2", "item3", "item4", "item5"}
		processed := []string{}

		for i, item := range items {
			// Check for cancellation
			select {
			case <-ctx.Context().Done():
				log.Printf("Batch processing cancelled after %d items", i)
				return ctx.JSON(499, map[string]interface{}{
					"error":     "Request cancelled",
					"processed": processed,
				})
			default:
			}

			// Simulate processing
			time.Sleep(500 * time.Millisecond)
			processed = append(processed, item)
			log.Printf("Processed: %s", item)
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":   "Batch completed",
			"processed": processed,
		})
	}, 0))

	// Example 6: Error handler that handles cancellation
	server.SetErrorHandler(func(ctx pkg.Context, err error) error {
		if pkg.IsCancellationError(err) {
			log.Printf("Request cancelled: %s %s", ctx.Request().Method, ctx.Request().URL.Path)
			// Don't send response for cancelled requests
			return nil
		}

		// Handle other errors normally
		return ctx.JSON(500, map[string]string{
			"error": err.Error(),
		})
	})

	server.SetRouter(router)

	// Start server
	log.Println("Starting HTTP/2 server on :8443")
	log.Println("Test cancellation with:")
	log.Println("  curl -k https://localhost:8443/api/long-running")
	log.Println("  (Press Ctrl+C to cancel)")
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  /api/simple         - Simple response")
	log.Println("  GET  /api/long-running   - Long operation with cancellation")
	log.Println("  GET  /api/database       - Database query with timeout")
	log.Println("  GET  /api/stream         - Server-sent events stream")
	log.Println("  POST /api/batch          - Batch processing")

	// For this example, you'll need to generate certificates:
	// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
	if err := server.ListenTLS(":8443", "cert.pem", "key.pem"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
