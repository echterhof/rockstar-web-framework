# HTTP/2 Stream Cancellation

This document describes how the Rockstar framework handles HTTP/2 stream cancellation, allowing servers to detect and respond to client-initiated request cancellations early.

## Overview

HTTP/2 allows clients to cancel individual streams (requests) without closing the entire connection. This is useful when:
- A user navigates away from a page before it loads
- A client times out waiting for a response
- A client decides it no longer needs the data

The framework automatically detects stream cancellations and stops processing requests early, saving server resources.

## How It Works

### Automatic Detection

The server automatically monitors the request context for cancellation at multiple points:

1. **Before request processing** - Checks if the stream is already cancelled
2. **During middleware execution** - Checks between each middleware
3. **During handler execution** - Monitors for cancellation while the handler runs

When a cancellation is detected, the server:
- Stops processing immediately
- Cleans up resources
- Does not send a response (the stream is already closed)

### Request Context

Every request has an associated `context.Context` that is cancelled when:
- The HTTP/2 stream is reset by the client
- The connection is closed
- A timeout occurs

Access the context via `ctx.Context()` in your handlers.

## Usage

### Basic Cancellation Checking

Check for cancellation in your handlers:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Check if request is cancelled
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    
    // Process request
    return ctx.JSON(200, data)
})
```

### Cancellation Middleware

Use the built-in cancellation middleware to automatically handle cancellations:

```go
// Apply globally
server.SetMiddleware(pkg.CancellationMiddleware())

// Or per-route
router.GET("/api/data", handler, pkg.CancellationMiddleware())
```

### Long-Running Operations

For operations that take time, check for cancellation periodically:

```go
router.GET("/api/process", func(ctx pkg.Context) error {
    for i := 0; i < 100; i++ {
        // Check for cancellation
        select {
        case <-ctx.Context().Done():
            log.Printf("Processing cancelled at step %d", i)
            return ctx.Context().Err()
        default:
        }
        
        // Do work
        processItem(i)
    }
    
    return ctx.JSON(200, result)
})
```

### Cancellation-Aware Handler

Wrap handlers to automatically monitor for cancellation:

```go
router.GET("/api/data", pkg.WithCancellationCheck(func(ctx pkg.Context) error {
    // Your handler code
    // Cancellation is checked before and after execution
    return ctx.JSON(200, data)
}))
```

### Database Queries

Pass the request context to database operations:

```go
router.GET("/api/users", func(ctx pkg.Context) error {
    // Create query context with timeout
    queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
    defer cancel()
    
    // Execute query with context
    users, err := db.QueryContext(queryCtx, "SELECT * FROM users")
    if err != nil {
        if pkg.IsCancellationError(err) {
            return ctx.Context().Err()
        }
        return err
    }
    
    return ctx.JSON(200, users)
})
```

### Streaming Responses

Monitor for cancellation when streaming data:

```go
router.GET("/api/stream", func(ctx pkg.Context) error {
    ctx.SetHeader("Content-Type", "text/event-stream")
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Context().Done():
            // Client disconnected
            return ctx.Context().Err()
        case <-ticker.C:
            // Send event
            ctx.Response().Write([]byte("data: event\n\n"))
            ctx.Response().Flush()
        }
    }
})
```

## Error Handling

### Detecting Cancellation Errors

Use `IsCancellationError()` to check if an error is due to cancellation:

```go
if err != nil {
    if pkg.IsCancellationError(err) {
        // Request was cancelled
        log.Printf("Request cancelled")
        return nil
    }
    // Handle other errors
    return err
}
```

### Custom Error Handler

Handle cancellation errors in your error handler:

```go
server.SetErrorHandler(func(ctx pkg.Context, err error) error {
    if pkg.IsCancellationError(err) {
        // Don't send response for cancelled requests
        log.Printf("Request cancelled: %s", ctx.Request().URL.Path)
        return nil
    }
    
    // Handle other errors
    return ctx.JSON(500, map[string]string{
        "error": err.Error(),
    })
})
```

## Best Practices

### 1. Check Cancellation in Loops

Always check for cancellation in loops that may run for a long time:

```go
for i := 0; i < len(items); i++ {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    processItem(items[i])
}
```

### 2. Pass Context to Downstream Operations

Pass the request context to database queries, HTTP clients, and other operations:

```go
// Database
rows, err := db.QueryContext(ctx.Context(), query)

// HTTP client
req, _ := http.NewRequestWithContext(ctx.Context(), "GET", url, nil)
resp, err := client.Do(req)
```

### 3. Use Timeouts

Combine cancellation with timeouts for better resource management:

```go
queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
defer cancel()

result, err := performQuery(queryCtx)
```

### 4. Clean Up Resources

Use defer to ensure resources are cleaned up even if cancelled:

```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()

// Process file with cancellation checks
```

### 5. Don't Ignore Cancellation

Always respect cancellation signals to avoid wasting resources:

```go
// Bad - ignores cancellation
for i := 0; i < 1000; i++ {
    processItem(i)
}

// Good - checks for cancellation
for i := 0; i < 1000; i++ {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    processItem(i)
}
```

## API Reference

### Middleware Functions

#### `CancellationMiddleware()`
Creates middleware that monitors for request cancellation throughout the request lifecycle.

```go
server.SetMiddleware(pkg.CancellationMiddleware())
```

#### `WithCancellationCheck(handler HandlerFunc)`
Wraps a handler to check for cancellation before and after execution.

```go
router.GET("/api/data", pkg.WithCancellationCheck(handler))
```

#### `CancellationAwareHandler(handler func(Context) error, checkInterval int)`
Creates a handler that monitors for cancellation during execution.

```go
router.GET("/api/data", pkg.CancellationAwareHandler(handler, 0))
```

### Utility Functions

#### `IsCancellationError(err error) bool`
Checks if an error is a cancellation error (context.Canceled or context.DeadlineExceeded).

```go
if pkg.IsCancellationError(err) {
    // Handle cancellation
}
```

### Context Methods

#### `Context() context.Context`
Returns the underlying Go context for the request.

```go
ctx.Context().Done() // Channel closed when cancelled
ctx.Context().Err()  // Returns cancellation error
```

#### `WithTimeout(timeout time.Duration) Context`
Creates a new context with a timeout.

```go
timeoutCtx := ctx.WithTimeout(5 * time.Second)
```

#### `WithCancel() (Context, context.CancelFunc)`
Creates a new context with a cancel function.

```go
cancelCtx, cancel := ctx.WithCancel()
defer cancel()
```

## Testing

Test cancellation handling in your handlers:

```go
func TestHandlerCancellation(t *testing.T) {
    // Create cancelled context
    cancelledCtx, cancel := context.WithCancel(context.Background())
    cancel()
    
    // Create request with cancelled context
    req := httptest.NewRequest("GET", "/api/data", nil)
    req = req.WithContext(cancelledCtx)
    
    // Create context
    ctx := pkg.NewContext(&pkg.Request{}, respWriter, cancelledCtx)
    
    // Call handler
    err := handler(ctx)
    
    // Verify cancellation was detected
    if !pkg.IsCancellationError(err) {
        t.Error("Expected cancellation error")
    }
}
```

## Performance Considerations

### Benefits
- **Resource savings**: Stop processing cancelled requests immediately
- **Better throughput**: Free up resources for active requests
- **Improved responsiveness**: Server can handle more concurrent requests

### Overhead
- Minimal overhead from context checking
- Goroutine creation for concurrent monitoring (only in createHandler)
- Recommended for all production deployments

### When to Use
- ✅ Long-running operations (> 1 second)
- ✅ Database queries
- ✅ External API calls
- ✅ File processing
- ✅ Streaming responses
- ⚠️ Very fast operations (< 10ms) - overhead may not be worth it

## Examples

See `examples/http2_cancellation_example.go` for complete working examples including:
- Simple cancellation checking
- Long-running operations
- Database queries with timeout
- Streaming responses
- Batch processing
- Custom error handling

## Troubleshooting

### Cancellation Not Detected

If cancellation isn't being detected:

1. Ensure HTTP/2 is enabled:
```go
config.EnableHTTP2 = true
server.EnableHTTP2()
```

2. Check that you're using HTTPS (required for HTTP/2):
```go
server.ListenTLS(":8443", "cert.pem", "key.pem")
```

3. Verify you're checking the context:
```go
select {
case <-ctx.Context().Done():
    return ctx.Context().Err()
default:
}
```

### Goroutine Leaks

If you see goroutine leaks:

1. Always use defer to clean up resources
2. Pass context to all blocking operations
3. Use timeouts for operations that might hang

### False Positives

If you're seeing cancellation errors when requests complete normally:

1. Check that you're not cancelling the context yourself
2. Verify timeout values are appropriate
3. Ensure you're not closing connections prematurely

## See Also

- [Server Configuration](./ARCHITECTURE.md#server-configuration)
- [Middleware](./middleware_implementation.md)
- [Error Handling](./error_handling_implementation.md)
- [Context Management](./API_REFERENCE.md#context)
