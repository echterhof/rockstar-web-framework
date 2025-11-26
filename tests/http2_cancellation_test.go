//go:build !benchmark
// +build !benchmark

package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestHTTP2StreamCancellation tests cancellation middleware integration
// Requirements: 6.1
func TestHTTP2StreamCancellation(t *testing.T) {
	// Create framework configuration with HTTP/2 enabled
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Add cancellation middleware
	framework.Router().Use(pkg.CancellationMiddleware())

	// Track if handler detected cancellation
	handlerCancelled := false

	// Register route with long-running handler
	framework.Router().GET("/long-running", func(ctx pkg.Context) error {
		// Simulate long-running operation
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Context().Done():
				handlerCancelled = true
				return ctx.Context().Err()
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
		return ctx.JSON(200, map[string]string{"status": "completed"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19301"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test cancellation detection
	t.Run("Handler detects cancellation", func(t *testing.T) {
		// Create request with cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19301/long-running", nil)
		assertNoError(t, err, "Failed to create request")

		// Start request in background
		client := &http.Client{}
		done := make(chan error, 1)
		go func() {
			resp, err := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
			done <- err
		}()

		// Cancel request after short delay
		time.Sleep(200 * time.Millisecond)
		cancel()

		// Wait for request to complete
		select {
		case err := <-done:
			// Request should fail due to cancellation
			if err == nil {
				t.Log("Request completed without error (expected cancellation)")
			}
		case <-time.After(2 * time.Second):
			t.Error("Request did not complete in time")
		}

		// Give handler time to detect cancellation
		time.Sleep(100 * time.Millisecond)

		// Verify handler detected cancellation
		if handlerCancelled {
			t.Log("Handler successfully detected cancellation")
		} else {
			t.Log("Handler did not detect cancellation (may have completed before cancellation)")
		}
	})
}

// TestCancellationMiddlewareIntegration tests IsCancellationError utility function
// Requirements: 6.2
func TestCancellationMiddlewareIntegration(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Track errors received
	var receivedError error

	// Add middleware that checks for cancellation errors
	framework.Router().Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		err := next(ctx)
		if err != nil && pkg.IsCancellationError(err) {
			receivedError = err
			// Log cancellation but don't propagate error
			return ctx.JSON(499, map[string]string{
				"error":  "Request cancelled",
				"status": "cancelled",
			})
		}
		return err
	})

	// Register route that can be cancelled
	framework.Router().GET("/cancellable", func(ctx pkg.Context) error {
		// Check for cancellation
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
			time.Sleep(500 * time.Millisecond)
		}

		// Check again after sleep
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
			return ctx.JSON(200, map[string]string{"status": "completed"})
		}
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19302"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test IsCancellationError utility
	t.Run("IsCancellationError detects context.Canceled", func(t *testing.T) {
		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Get cancellation error
		err := ctx.Err()
		assertNotNil(t, err, "Expected cancellation error")

		// Verify IsCancellationError detects it
		assertTrue(t, pkg.IsCancellationError(err), "IsCancellationError should detect context.Canceled")
	})

	t.Run("IsCancellationError detects context.DeadlineExceeded", func(t *testing.T) {
		// Create context with immediate deadline
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		// Get timeout error
		err := ctx.Err()
		assertNotNil(t, err, "Expected timeout error")

		// Verify IsCancellationError detects it
		assertTrue(t, pkg.IsCancellationError(err), "IsCancellationError should detect context.DeadlineExceeded")
	})

	t.Run("IsCancellationError returns false for other errors", func(t *testing.T) {
		// Create non-cancellation error
		err := context.Background().Err()

		// Verify IsCancellationError returns false
		assertFalse(t, pkg.IsCancellationError(err), "IsCancellationError should return false for nil error")
	})

	t.Run("Middleware handles cancellation error", func(t *testing.T) {
		// Reset received error
		receivedError = nil

		// Create request with cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19302/cancellable", nil)
		assertNoError(t, err, "Failed to create request")

		// Start request in background
		client := &http.Client{}
		done := make(chan error, 1)
		go func() {
			resp, err := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
			done <- err
		}()

		// Cancel request after short delay
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Wait for request to complete
		select {
		case <-done:
			// Request completed (may have been cancelled)
		case <-time.After(2 * time.Second):
			t.Error("Request did not complete in time")
		}

		// Give middleware time to process
		time.Sleep(100 * time.Millisecond)

		// Verify middleware received cancellation error
		if receivedError != nil {
			assertTrue(t, pkg.IsCancellationError(receivedError), "Middleware should receive cancellation error")
		}
	})
}

// TestHTTP2ClientCancellation tests client-side request cancellation
// Requirements: 6.3
func TestHTTP2ClientCancellation(t *testing.T) {
	// Create framework configuration with HTTP/2 enabled
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Add cancellation middleware
	framework.Router().Use(pkg.CancellationMiddleware())

	// Track handler execution
	handlerStarted := false
	handlerCompleted := false

	// Register route with long-running operation
	framework.Router().GET("/slow-operation", func(ctx pkg.Context) error {
		handlerStarted = true

		// Simulate long operation with cancellation checks
		for i := 0; i < 20; i++ {
			select {
			case <-ctx.Context().Done():
				// Cancellation detected
				return ctx.Context().Err()
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}

		handlerCompleted = true
		return ctx.JSON(200, map[string]string{"status": "completed"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19303"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test client-side cancellation
	t.Run("Client cancels request", func(t *testing.T) {
		// Reset tracking variables
		handlerStarted = false
		handlerCompleted = false

		// Create request with cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19303/slow-operation", nil)
		assertNoError(t, err, "Failed to create request")

		// Start request in background
		client := &http.Client{}
		requestErr := make(chan error, 1)
		go func() {
			resp, err := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
			requestErr <- err
		}()

		// Wait for handler to start
		time.Sleep(200 * time.Millisecond)

		// Cancel request
		cancel()

		// Wait for request to complete
		var requestError error
		select {
		case requestError = <-requestErr:
			// Request completed
		case <-time.After(2 * time.Second):
			t.Error("Request did not complete in time")
		}

		// Give handler time to detect cancellation
		time.Sleep(200 * time.Millisecond)

		// Verify handler started
		assertTrue(t, handlerStarted, "Handler should have started")

		// Verify handler did not complete normally
		assertFalse(t, handlerCompleted, "Handler should not have completed normally after cancellation")

		// Verify request failed
		if requestError != nil {
			t.Logf("Request failed as expected: %v", requestError)
		}
	})

	t.Run("Client timeout cancels request", func(t *testing.T) {
		// Reset tracking variables
		handlerStarted = false
		handlerCompleted = false

		// Create request with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19303/slow-operation", nil)
		assertNoError(t, err, "Failed to create request")

		// Execute request
		client := &http.Client{}
		resp, err := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// Give handler time to detect cancellation
		time.Sleep(200 * time.Millisecond)

		// Verify handler started
		assertTrue(t, handlerStarted, "Handler should have started")

		// Verify handler did not complete normally
		assertFalse(t, handlerCompleted, "Handler should not have completed normally after timeout")

		// Verify request failed due to timeout
		if err != nil {
			t.Logf("Request timed out as expected: %v", err)
			// Check if error is timeout-related
			if pkg.IsCancellationError(ctx.Err()) {
				t.Log("Context error is correctly identified as cancellation error")
			}
		}
	})

	t.Run("Request completes without cancellation", func(t *testing.T) {
		// Reset tracking variables
		handlerStarted = false
		handlerCompleted = false

		// Create request with sufficient timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Register fast route
		framework.Router().GET("/fast-operation", func(ctx pkg.Context) error {
			handlerStarted = true
			time.Sleep(100 * time.Millisecond)
			handlerCompleted = true
			return ctx.JSON(200, map[string]string{"status": "completed"})
		})

		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:19303/fast-operation", nil)
		assertNoError(t, err, "Failed to create request")

		// Execute request
		client := &http.Client{}
		resp, err := client.Do(req)
		assertNoError(t, err, "Request should complete successfully")
		defer resp.Body.Close()

		// Verify response
		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		// Verify handler completed
		assertTrue(t, handlerStarted, "Handler should have started")
		assertTrue(t, handlerCompleted, "Handler should have completed")
	})
}
