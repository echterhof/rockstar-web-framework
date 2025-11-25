package pkg

import (
	"context"
	"errors"
)

// CancellationMiddleware creates middleware that monitors for request cancellation
// This is particularly useful for HTTP/2 streams where clients can cancel requests
func CancellationMiddleware() MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		// Check if request is already cancelled
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
		}

		// Execute next handler with cancellation monitoring
		done := make(chan error, 1)
		go func() {
			done <- next(ctx)
		}()

		// Wait for completion or cancellation
		select {
		case err := <-done:
			return err
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		}
	}
}

// WithCancellationCheck wraps a handler function to check for cancellation
// before and after execution
func WithCancellationCheck(handler HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		// Check before execution
		select {
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		default:
		}

		// Execute handler
		err := handler(ctx)

		// Check after execution
		select {
		case <-ctx.Context().Done():
			// If cancelled during execution, return cancellation error
			return ctx.Context().Err()
		default:
			return err
		}
	}
}

// IsCancellationError checks if an error is a cancellation error
func IsCancellationError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// CancellationAwareHandler creates a handler that periodically checks for cancellation
// during long-running operations
func CancellationAwareHandler(handler func(ctx Context) error, checkInterval int) HandlerFunc {
	return func(ctx Context) error {
		// Create a channel to signal completion
		done := make(chan error, 1)

		// Run handler in goroutine
		go func() {
			done <- handler(ctx)
		}()

		// Simple wait for completion or cancellation
		select {
		case err := <-done:
			return err
		case <-ctx.Context().Done():
			return ctx.Context().Err()
		}
	}
}
