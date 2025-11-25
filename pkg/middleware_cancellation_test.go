package pkg

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCancellationMiddleware(t *testing.T) {
	t.Run("passes through when not cancelled", func(t *testing.T) {
		called := false
		handler := func(ctx Context) error {
			called = true
			return nil
		}

		middleware := CancellationMiddleware()

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, req.Context())

		err := middleware(ctx, handler)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !called {
			t.Error("Handler was not called")
		}
	})

	t.Run("detects cancellation before handler", func(t *testing.T) {
		called := false
		handler := func(ctx Context) error {
			called = true
			return nil
		}

		middleware := CancellationMiddleware()

		// Create cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(cancelledCtx)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, cancelledCtx)

		err := middleware(ctx, handler)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
		if called {
			t.Error("Handler should not have been called")
		}
	})

	t.Run("detects cancellation during handler execution", func(t *testing.T) {
		handlerStarted := make(chan struct{})

		handler := func(ctx Context) error {
			close(handlerStarted)
			// Simulate long-running operation
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		middleware := CancellationMiddleware()

		// Create context that will be cancelled
		cancelCtx, cancel := context.WithCancel(context.Background())

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(cancelCtx)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, cancelCtx)

		// Run middleware in goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- middleware(ctx, handler)
		}()

		// Wait for handler to start
		<-handlerStarted

		// Cancel the context
		cancel()

		// Wait for middleware to return
		err := <-errChan

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestWithCancellationCheck(t *testing.T) {
	t.Run("executes handler when not cancelled", func(t *testing.T) {
		called := false
		handler := func(ctx Context) error {
			called = true
			return nil
		}

		wrappedHandler := WithCancellationCheck(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, req.Context())

		err := wrappedHandler(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !called {
			t.Error("Handler was not called")
		}
	})

	t.Run("returns error when cancelled before execution", func(t *testing.T) {
		handler := func(ctx Context) error {
			return nil
		}

		wrappedHandler := WithCancellationCheck(handler)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(cancelledCtx)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, cancelledCtx)

		err := wrappedHandler(ctx)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestIsCancellationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "context.Canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "context.DeadlineExceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "wrapped context.Canceled",
			err:      errors.New("wrapped: " + context.Canceled.Error()),
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCancellationError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCancellationAwareHandler(t *testing.T) {
	t.Run("completes normally when not cancelled", func(t *testing.T) {
		called := false
		handler := func(ctx Context) error {
			called = true
			return nil
		}

		wrappedHandler := CancellationAwareHandler(handler, 0)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, req.Context())

		err := wrappedHandler(ctx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !called {
			t.Error("Handler was not called")
		}
	})

	t.Run("detects cancellation during execution", func(t *testing.T) {
		handlerStarted := make(chan struct{})

		handler := func(ctx Context) error {
			close(handlerStarted)
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		wrappedHandler := CancellationAwareHandler(handler, 0)

		cancelCtx, cancel := context.WithCancel(context.Background())

		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(cancelCtx)
		w := httptest.NewRecorder()
		respWriter := NewResponseWriter(w)

		ctx := NewContext(&Request{}, respWriter, cancelCtx)

		errChan := make(chan error, 1)
		go func() {
			errChan <- wrappedHandler(ctx)
		}()

		<-handlerStarted
		cancel()

		err := <-errChan

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}
