# Middleware System Implementation

## Overview

The Rockstar Web Framework includes a powerful and flexible middleware system that allows you to process requests before and after they reach your handlers. The middleware system supports configurable ordering, dynamic management, and both pre-processing and post-processing capabilities.

## Key Features

- **Configurable Ordering**: Middleware execution order is not static and can be configured using priorities
- **Pre and Post Processing**: Middleware can execute before or after the main handler
- **Dynamic Management**: Enable, disable, or change middleware configuration at runtime
- **Thread-Safe**: All operations are protected by mutexes for concurrent access
- **Helper Functions**: Built-in helpers for common middleware patterns

## Architecture

### Middleware Positions

Middleware can be positioned in two ways:

1. **Before Position** (`MiddlewarePositionBefore`): Executes before the handler
   - Higher priority middleware executes first
   - Useful for authentication, logging, validation

2. **After Position** (`MiddlewarePositionAfter`): Executes after the handler
   - Higher priority middleware executes last
   - Useful for response modification, headers, cleanup

### Execution Flow

```
Request → Before MW (High Priority) → Before MW (Low Priority) → 
After MW (Low Priority) → After MW (High Priority) → Handler → 
After MW (High Priority) completes → After MW (Low Priority) completes →
Before MW (Low Priority) completes → Before MW (High Priority) completes → Response
```

## Core Interfaces

### MiddlewareEngine

The `MiddlewareEngine` interface provides methods for managing middleware:

```go
type MiddlewareEngine interface {
    Register(config MiddlewareConfig) error
    Unregister(name string) error
    Enable(name string) error
    Disable(name string) error
    SetPriority(name string, priority int) error
    SetPosition(name string, position MiddlewarePosition) error
    Execute(ctx Context, handler HandlerFunc) error
    List() []MiddlewareConfig
    Clear()
}
```

### MiddlewareConfig

Configuration for a middleware:

```go
type MiddlewareConfig struct {
    Name     string              // Unique identifier
    Handler  MiddlewareFunc      // The middleware function
    Position MiddlewarePosition  // Before or After
    Priority int                 // Execution priority
    Enabled  bool                // Whether middleware is active
}
```

## Usage Examples

### Basic Middleware Registration

```go
engine := pkg.NewMiddlewareEngine()

// Register a logging middleware
err := engine.Register(pkg.MiddlewareConfig{
    Name:     "logger",
    Position: pkg.MiddlewarePositionBefore,
    Priority: 100,
    Enabled:  true,
    Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
        start := time.Now()
        fmt.Printf("Request: %s %s\n", ctx.Request().Method, ctx.Request().URL.Path)
        
        err := next(ctx)
        
        fmt.Printf("Completed in %v\n", time.Since(start))
        return err
    },
})
```

### Authentication Middleware

```go
engine.Register(pkg.MiddlewareConfig{
    Name:     "auth",
    Position: pkg.MiddlewarePositionBefore,
    Priority: 50,
    Enabled:  true,
    Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
        token := ctx.GetHeader("Authorization")
        if token == "" {
            return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
        }
        
        // Validate token...
        return next(ctx)
    },
})
```

### Response Modification Middleware

```go
engine.Register(pkg.MiddlewareConfig{
    Name:     "cors",
    Position: pkg.MiddlewarePositionAfter,
    Priority: 50,
    Enabled:  true,
    Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
        err := next(ctx)
        
        // Add CORS headers after handler executes
        ctx.SetHeader("Access-Control-Allow-Origin", "*")
        ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        
        return err
    },
})
```

### Dynamic Middleware Management

```go
// Disable middleware
engine.Disable("auth")

// Change priority
engine.SetPriority("logger", 200)

// Change position
engine.SetPosition("cors", pkg.MiddlewarePositionBefore)

// Re-enable middleware
engine.Enable("auth")

// Remove middleware
engine.Unregister("logger")
```

## Middleware Helpers

### ChainMiddleware

Combine multiple middleware into a single middleware:

```go
mw1 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
    fmt.Println("MW1 before")
    err := next(ctx)
    fmt.Println("MW1 after")
    return err
}

mw2 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
    fmt.Println("MW2 before")
    err := next(ctx)
    fmt.Println("MW2 after")
    return err
}

chained := pkg.ChainMiddleware(mw1, mw2)
```

### SkipMiddleware

Conditionally skip middleware execution:

```go
skipMw := pkg.SkipMiddleware(
    func(ctx pkg.Context) bool {
        // Skip for health check endpoints
        return ctx.Request().URL.Path == "/health"
    },
    authMiddleware,
)
```

### RecoverMiddleware

Recover from panics in handlers:

```go
recoverMw := pkg.RecoverMiddleware(func(ctx pkg.Context, recovered interface{}) error {
    log.Printf("Panic recovered: %v", recovered)
    return ctx.JSON(500, map[string]string{
        "error": "Internal server error",
    })
})
```

## Integration with Server

The middleware engine integrates seamlessly with the server:

```go
// Create server
server := pkg.NewServer(pkg.ServerConfig{})

// Create middleware engine
middlewareEngine := pkg.NewMiddlewareEngine()

// Register middleware
middlewareEngine.Register(loggingMiddleware)
middlewareEngine.Register(authMiddleware)

// The server will use the middleware engine during request processing
```

## Integration with Router

Middleware can be applied at different levels:

### Global Middleware

```go
router := pkg.NewRouter()
router.Use(loggingMiddleware, authMiddleware)
```

### Route Group Middleware

```go
api := router.Group("/api", authMiddleware, rateLimitMiddleware)
api.GET("/users", getUsersHandler)
api.POST("/users", createUserHandler)
```

### Per-Route Middleware

```go
router.GET("/admin", adminHandler, adminAuthMiddleware, auditMiddleware)
```

## Best Practices

1. **Use Descriptive Names**: Give middleware clear, descriptive names for easier management
2. **Set Appropriate Priorities**: Use priority ranges (e.g., 0-50 for low, 51-100 for medium, 101+ for high)
3. **Handle Errors Properly**: Always check and propagate errors from `next(ctx)`
4. **Keep Middleware Focused**: Each middleware should have a single responsibility
5. **Use Position Wisely**: 
   - Use `Before` for validation, authentication, logging
   - Use `After` for response modification, cleanup, metrics
6. **Consider Performance**: Middleware executes on every request, keep it efficient
7. **Document Dependencies**: If middleware depends on other middleware, document it

## Common Middleware Patterns

### Request Logging

```go
func LoggingMiddleware() pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        start := time.Now()
        req := ctx.Request()
        
        log.Printf("[%s] %s %s", req.ID, req.Method, req.URL.Path)
        
        err := next(ctx)
        
        duration := time.Since(start)
        log.Printf("[%s] Completed in %v", req.ID, duration)
        
        return err
    }
}
```

### Rate Limiting

```go
func RateLimitMiddleware(limit int) pkg.MiddlewareFunc {
    limiter := rate.NewLimiter(rate.Limit(limit), limit)
    
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if !limiter.Allow() {
            return ctx.JSON(429, map[string]string{
                "error": "Rate limit exceeded",
            })
        }
        return next(ctx)
    }
}
```

### Request Timeout

```go
func TimeoutMiddleware(timeout time.Duration) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        timeoutCtx := ctx.WithTimeout(timeout)
        
        done := make(chan error, 1)
        go func() {
            done <- next(timeoutCtx)
        }()
        
        select {
        case err := <-done:
            return err
        case <-timeoutCtx.Context().Done():
            return ctx.JSON(408, map[string]string{
                "error": "Request timeout",
            })
        }
    }
}
```

### CORS

```go
func CORSMiddleware(origins []string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        err := next(ctx)
        
        origin := ctx.GetHeader("Origin")
        if contains(origins, origin) {
            ctx.SetHeader("Access-Control-Allow-Origin", origin)
            ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
        }
        
        return err
    }
}
```

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 7.1**: Middleware executes before pipelines or views ✓
- **Requirement 7.2**: Middleware executes after pipelines or views ✓
- **Requirement 7.3**: Middleware execution order is configurable (not static) ✓

## Testing

The middleware system includes comprehensive unit tests covering:

- Middleware registration and unregistration
- Enable/disable functionality
- Priority and position management
- Execution order verification
- Error propagation
- Helper functions (chain, skip, recover)

Run tests with:

```bash
go test -v -run TestMiddleware ./pkg
```

## Performance Considerations

- Middleware execution is optimized with minimal allocations
- Thread-safe operations use read-write locks for better concurrency
- Middleware sorting uses simple bubble sort (sufficient for typical middleware counts)
- Disabled middleware is filtered out before execution

## Future Enhancements

Potential future improvements:

1. Middleware groups for easier management
2. Conditional middleware based on path patterns
3. Middleware metrics and monitoring
4. Middleware composition DSL
5. Async middleware support
