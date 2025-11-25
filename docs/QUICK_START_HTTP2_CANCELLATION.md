# Quick Start: HTTP/2 Stream Cancellation

## 1-Minute Setup

### Enable Cancellation Detection

```go
// Create server with HTTP/2
server := pkg.NewServer(pkg.ServerConfig{
    EnableHTTP2: true,
})
server.EnableHTTP2()

// Add cancellation middleware
server.SetMiddleware(pkg.CancellationMiddleware())
```

### Check for Cancellation in Handlers

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Check if cancelled
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    
    // Your code here
    return ctx.JSON(200, data)
})
```

## Common Patterns

### Pattern 1: Long-Running Loop

```go
for i := 0; i < 1000; i++ {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    processItem(i)
}
```

### Pattern 2: Database Query

```go
queryCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
defer cancel()

result, err := db.QueryContext(queryCtx, query)
if pkg.IsCancellationError(err) {
    return ctx.Context().Err()
}
```

### Pattern 3: Streaming Response

```go
for {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    case data := <-dataChan:
        ctx.Response().Write(data)
        ctx.Response().Flush()
    }
}
```

## Testing

```bash
# Run tests
go test -v ./pkg -run TestCancellation

# Run example
go run examples/http2_cancellation_example.go

# Test with curl (Ctrl+C to cancel)
curl -k https://localhost:8443/api/long-running
```

## See Full Documentation

- [Complete Guide](http2_cancellation.md)
- [Implementation Details](HTTP2_CANCELLATION_IMPLEMENTATION.md)
- [Example Code](../examples/http2_cancellation_example.go)
