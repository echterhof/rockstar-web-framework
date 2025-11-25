# HTTP/2 Stream Cancellation Implementation

## Summary

Implemented early cancellation detection for HTTP/2 streams in the Rockstar web framework. When a client cancels an HTTP/2 stream (e.g., user navigates away, timeout, or explicit cancellation), the server now detects this immediately and stops processing the request, saving server resources.

## Changes Made

### 1. Server Implementation (`pkg/server_impl.go`)

#### Modified `createHandler()` function:
- Added early cancellation check before request processing
- Implemented concurrent monitoring of handler execution and request context
- Handler execution now runs in a goroutine with cancellation detection
- Server stops processing immediately when HTTP/2 stream is cancelled

#### Modified `executeHandler()` function:
- Added cancellation check before starting handler execution
- Added cancellation checks between each middleware execution
- Ensures middleware chain respects context cancellation

### 2. Cancellation Middleware (`pkg/middleware_cancellation.go`)

Created new middleware and utilities for handling request cancellation:

#### `CancellationMiddleware()`
- Monitors for request cancellation throughout the request lifecycle
- Can be applied globally or per-route
- Executes handler in goroutine with cancellation monitoring

#### `WithCancellationCheck(handler HandlerFunc)`
- Wraps handlers to check for cancellation before and after execution
- Lightweight wrapper for simple handlers

#### `CancellationAwareHandler(handler, checkInterval)`
- Creates handlers that monitor for cancellation during execution
- Useful for long-running operations

#### `IsCancellationError(err error) bool`
- Utility function to check if an error is due to cancellation
- Checks for `context.Canceled` and `context.DeadlineExceeded`

### 3. Tests (`pkg/middleware_cancellation_test.go`)

Comprehensive test suite covering:
- Middleware behavior with and without cancellation
- Cancellation detection before handler execution
- Cancellation detection during handler execution
- Wrapped handler behavior
- Error type detection

### 4. Integration Tests (`tests/http2_cancellation_test.go`)

Integration tests demonstrating:
- HTTP/2 server configuration with cancellation middleware
- Long-running handler with cancellation checks
- Client-side cancellation scenarios

### 5. Example Application (`examples/http2_cancellation_example.go`)

Complete working example showing:
- Simple handler with cancellation checks
- Long-running operations with periodic cancellation checks
- Database queries with timeout and cancellation
- Streaming responses with cancellation
- Batch processing with cancellation
- Custom error handler for cancellation errors

### 6. Documentation (`docs/http2_cancellation.md`)

Comprehensive documentation including:
- Overview of HTTP/2 stream cancellation
- How the framework detects cancellation
- Usage examples for various scenarios
- Best practices
- API reference
- Troubleshooting guide

## How It Works

### Automatic Detection

The server monitors for cancellation at multiple points:

1. **Before request processing**: Checks if stream is already cancelled
2. **During middleware execution**: Checks between each middleware
3. **During handler execution**: Monitors while handler runs

### Request Context

Every HTTP request has an associated `context.Context` that is cancelled when:
- HTTP/2 stream is reset by client
- Connection is closed
- Timeout occurs

### Cancellation Flow

```
Client cancels stream
    ↓
HTTP/2 layer cancels request context
    ↓
Server detects ctx.Done() signal
    ↓
Handler execution stops
    ↓
Resources cleaned up
    ↓
No response sent (stream already closed)
```

## Usage Examples

### Basic Usage

```go
// Apply middleware globally
server.SetMiddleware(pkg.CancellationMiddleware())

// Check for cancellation in handler
router.GET("/api/data", func(ctx pkg.Context) error {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    // Process request
    return ctx.JSON(200, data)
})
```

### Long-Running Operations

```go
router.GET("/api/process", func(ctx pkg.Context) error {
    for i := 0; i < 100; i++ {
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        default:
        }
        processItem(i)
    }
    return ctx.JSON(200, result)
})
```

### Database Queries

```go
router.GET("/api/users", func(ctx pkg.Context) error {
    queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
    defer cancel()
    
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

## Benefits

1. **Resource Efficiency**: Stop processing cancelled requests immediately
2. **Better Throughput**: Free up resources for active requests
3. **Improved Responsiveness**: Handle more concurrent requests
4. **Graceful Degradation**: Clean handling of client disconnections

## Testing

Run the cancellation tests:

```bash
go test -v ./pkg -run TestCancellation
go test -v ./tests -run TestHTTP2
```

Run the example:

```bash
# Generate test certificates
openssl req -x509 -newkey rsa:4096 -keyout cert.pem -out key.pem -days 365 -nodes

# Run example
go run examples/http2_cancellation_example.go

# Test with curl (press Ctrl+C to cancel)
curl -k https://localhost:8443/api/long-running
```

## Performance Impact

- **Minimal overhead**: Context checking is very fast (nanoseconds)
- **Goroutine per request**: One additional goroutine for concurrent monitoring
- **Recommended for production**: Benefits outweigh minimal overhead

## Compatibility

- Works with HTTP/1.1, HTTP/2, and HTTP/3 (QUIC)
- Backward compatible with existing handlers
- Optional middleware - can be applied selectively

## Future Enhancements

Potential improvements:
- Metrics for cancelled requests
- Configurable cancellation behavior
- Automatic retry logic for transient cancellations
- Integration with distributed tracing

## Files Modified

- `pkg/server_impl.go` - Core server cancellation detection
- `pkg/middleware_cancellation.go` - New middleware and utilities
- `pkg/middleware_cancellation_test.go` - Unit tests
- `tests/http2_cancellation_test.go` - Integration tests
- `examples/http2_cancellation_example.go` - Example application
- `docs/http2_cancellation.md` - Documentation

## References

- [HTTP/2 Specification - Stream Cancellation](https://httpwg.org/specs/rfc7540.html#StreamStates)
- [Go Context Package](https://pkg.go.dev/context)
- [HTTP/2 in Go](https://pkg.go.dev/golang.org/x/net/http2)
