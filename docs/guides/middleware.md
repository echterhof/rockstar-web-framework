# Middleware

## Overview

Middleware functions are the backbone of request processing in the Rockstar Web Framework. They sit between the incoming request and your handler functions, allowing you to execute code before and after handlers run. Middleware can modify requests, responses, perform authentication, log activity, handle errors, and much more.

Middleware functions follow a simple pattern: they receive a context and a "next" function, execute their logic, and call the next function to continue the chain. This design enables powerful composition and reusability.

## Quick Start

Here's a simple middleware example:

```go
package main

import (
    "fmt"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
    }
    
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Add global middleware
    router.Use(loggingMiddleware)
    
    // Add route-specific middleware
    router.GET("/protected", protectedHandler, authMiddleware)
    
    app.Listen(":8080")
}

// Logging middleware
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    fmt.Printf("Request: %s %s\n", ctx.Request().Method, ctx.Request().URL.Path)
    
    // Call next middleware/handler
    err := next(ctx)
    
    duration := time.Since(start)
    fmt.Printf("Completed in %v\n", duration)
    
    return err
}

// Authentication middleware
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    
    if token == "" {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Validate token and continue
    return next(ctx)
}

func protectedHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{"message": "Protected resource"})
}
```

## Middleware Concept

### How Middleware Works

Middleware functions form a chain where each middleware can:

1. **Execute code before the handler** - Validate, authenticate, log, etc.
2. **Call the next middleware/handler** - Continue the chain
3. **Execute code after the handler** - Log results, modify responses, clean up
4. **Short-circuit the chain** - Return early without calling next

```
Request → Middleware 1 → Middleware 2 → Handler → Middleware 2 → Middleware 1 → Response
          ↓ before      ↓ before       ↓ execute  ↑ after       ↑ after
```

### Middleware Signature

All middleware functions follow this signature:

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

- `ctx`: The request context containing request/response data
- `next`: The next middleware or handler in the chain
- Returns an `error` if something goes wrong

## Execution Order

Middleware executes in the order it's registered:

```go
router.Use(middleware1)  // Executes first
router.Use(middleware2)  // Executes second
router.Use(middleware3)  // Executes third

router.GET("/", handler, middleware4, middleware5)
// Execution order: middleware1 → middleware2 → middleware3 → middleware4 → middleware5 → handler
```

### Before and After Handler

Middleware can execute code both before and after the handler:

```go
func timingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // BEFORE handler
    start := time.Now()
    fmt.Println("Handler starting...")
    
    // Call next middleware/handler
    err := next(ctx)
    
    // AFTER handler
    duration := time.Since(start)
    fmt.Printf("Handler completed in %v\n", duration)
    
    return err
}
```

### Short-Circuiting

Middleware can stop the chain by not calling `next()`:

```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    
    if token == "" {
        // Stop here - don't call next()
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Continue the chain
    return next(ctx)
}
```

## Global Middleware

Global middleware applies to all routes:

```go
app := pkg.New(config)
router := app.Router()

// Add global middleware
router.Use(loggingMiddleware)
router.Use(recoveryMiddleware)
router.Use(corsMiddleware)

// All routes will use these middleware
router.GET("/", homeHandler)
router.GET("/api/users", usersHandler)
router.POST("/api/posts", createPostHandler)
```

## Route-Specific Middleware

Apply middleware to individual routes:

```go
// Single middleware
router.GET("/protected", protectedHandler, authMiddleware)

// Multiple middleware (executed in order)
router.GET("/admin", adminHandler, authMiddleware, adminCheckMiddleware, rateLimitMiddleware)

// Different middleware for different routes
router.GET("/public", publicHandler)  // No middleware
router.GET("/user", userHandler, authMiddleware)  // Auth only
router.GET("/admin", adminHandler, authMiddleware, adminMiddleware)  // Auth + Admin
```

## Route Group Middleware

Apply middleware to groups of routes:

```go
// Create group with shared middleware
api := router.Group("/api", loggingMiddleware, corsMiddleware)

// All routes in this group use the middleware
api.GET("/users", usersHandler)
api.GET("/posts", postsHandler)
api.POST("/comments", createCommentHandler)

// Nested groups inherit parent middleware
v1 := api.Group("/v1", versionMiddleware("v1"))
v1.GET("/users", usersV1Handler)  // Uses: logging, cors, version middleware

// Admin group with additional middleware
admin := api.Group("/admin", authMiddleware, adminMiddleware)
admin.GET("/dashboard", dashboardHandler)  // Uses: logging, cors, auth, admin middleware
```

## Built-in Middleware

The framework provides several built-in middleware functions:

### ChainMiddleware

Combine multiple middleware into a single middleware:

```go
// Chain multiple middleware
combined := pkg.ChainMiddleware(
    loggingMiddleware,
    authMiddleware,
    rateLimitMiddleware,
)

router.GET("/api/data", dataHandler, combined)
```

### SkipMiddleware

Conditionally skip middleware execution:

```go
// Skip authentication for health check
skipAuth := pkg.SkipMiddleware(
    func(ctx pkg.Context) bool {
        return ctx.Request().URL.Path == "/health"
    },
    authMiddleware,
)

router.GET("/health", healthHandler, skipAuth)
router.GET("/api/data", dataHandler, skipAuth)  // Auth required for this route
```

### RecoverMiddleware

Recover from panics in handlers:

```go
recovery := pkg.RecoverMiddleware(func(ctx pkg.Context, recovered interface{}) error {
    log.Printf("Panic recovered: %v", recovered)
    return ctx.JSON(500, map[string]string{
        "error": "Internal server error",
    })
})

router.Use(recovery)
```

### CancellationMiddleware

Monitor for request cancellation (useful for HTTP/2):

```go
router.Use(pkg.CancellationMiddleware())

// Handler can check for cancellation
router.GET("/long-task", func(ctx pkg.Context) error {
    // Long-running operation
    for i := 0; i < 100; i++ {
        // Check if request was cancelled
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        default:
            // Continue processing
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    return ctx.JSON(200, map[string]string{"status": "completed"})
})
```

## Common Middleware Patterns

### Logging Middleware

Log all requests and responses:

```go
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Log request
    log.Printf("[%s] %s %s",
        ctx.Request().Method,
        ctx.Request().URL.Path,
        ctx.ClientIP(),
    )
    
    // Execute handler
    err := next(ctx)
    
    // Log response
    duration := time.Since(start)
    status := ctx.Response().Status()
    log.Printf("[%d] Completed in %v", status, duration)
    
    return err
}
```

### Authentication Middleware

Verify user authentication:

```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Get token from header
    token := ctx.GetHeader("Authorization")
    
    if token == "" {
        return ctx.JSON(401, map[string]string{
            "error": "Missing authorization token",
        })
    }
    
    // Validate token
    userID, err := validateToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{
            "error": "Invalid token",
        })
    }
    
    // Store user ID in context for handlers to use
    ctx.Set("user_id", userID)
    
    // Continue to next middleware/handler
    return next(ctx)
}

func validateToken(token string) (string, error) {
    // Implement token validation logic
    // Return user ID if valid
    return "user123", nil
}
```

### Authorization Middleware

Check user permissions:

```go
func requireRole(role string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        // Get user from context (set by auth middleware)
        userID, _ := ctx.Get("user_id").(string)
        
        // Check if user has required role
        hasRole, err := checkUserRole(userID, role)
        if err != nil || !hasRole {
            return ctx.JSON(403, map[string]string{
                "error": "Insufficient permissions",
            })
        }
        
        return next(ctx)
    }
}

// Usage
router.GET("/admin/users", adminUsersHandler, 
    authMiddleware, 
    requireRole("admin"),
)
```

### CORS Middleware

Handle Cross-Origin Resource Sharing:

```go
func corsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Set CORS headers
    ctx.SetHeader("Access-Control-Allow-Origin", "*")
    ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
    ctx.SetHeader("Access-Control-Max-Age", "3600")
    
    // Handle preflight requests
    if ctx.Request().Method == "OPTIONS" {
        return ctx.NoContent(204)
    }
    
    return next(ctx)
}
```

### Rate Limiting Middleware

Limit request rates per client:

```go
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    clientIP := ctx.ClientIP()
    
    // Check rate limit (implementation depends on your rate limiter)
    allowed, err := checkRateLimit(clientIP)
    if err != nil {
        return ctx.JSON(500, map[string]string{
            "error": "Rate limit check failed",
        })
    }
    
    if !allowed {
        ctx.SetHeader("Retry-After", "60")
        return ctx.JSON(429, map[string]string{
            "error": "Too many requests",
        })
    }
    
    return next(ctx)
}
```

### Request Validation Middleware

Validate request data:

```go
func validateJSONMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check Content-Type
    contentType := ctx.GetHeader("Content-Type")
    if contentType != "application/json" {
        return ctx.JSON(400, map[string]string{
            "error": "Content-Type must be application/json",
        })
    }
    
    // Check request body size
    if ctx.Request().ContentLength > 1024*1024 { // 1MB
        return ctx.JSON(413, map[string]string{
            "error": "Request body too large",
        })
    }
    
    return next(ctx)
}
```

### Caching Middleware

Cache responses:

```go
func cacheMiddleware(ttl time.Duration) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        // Generate cache key
        cacheKey := ctx.Request().Method + ":" + ctx.Request().URL.Path
        
        // Check cache
        cache := ctx.Cache()
        if cached, err := cache.Get(cacheKey); err == nil {
            // Return cached response
            return ctx.JSON(200, cached)
        }
        
        // Execute handler
        err := next(ctx)
        
        // Cache successful responses
        if err == nil && ctx.Response().Status() == 200 {
            // Store response in cache
            // (implementation depends on how you capture response)
        }
        
        return err
    }
}
```

### Timing Middleware

Measure handler execution time:

```go
func timingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Execute handler
    err := next(ctx)
    
    // Calculate duration
    duration := time.Since(start)
    
    // Add timing header
    ctx.SetHeader("X-Response-Time", duration.String())
    
    // Log slow requests
    if duration > 1*time.Second {
        log.Printf("Slow request: %s %s took %v",
            ctx.Request().Method,
            ctx.Request().URL.Path,
            duration,
        )
    }
    
    return err
}
```

### Request ID Middleware

Add unique request IDs for tracing:

```go
func requestIDMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Generate or get request ID
    requestID := ctx.GetHeader("X-Request-ID")
    if requestID == "" {
        requestID = generateRequestID()
    }
    
    // Store in context
    ctx.Set("request_id", requestID)
    
    // Add to response headers
    ctx.SetHeader("X-Request-ID", requestID)
    
    return next(ctx)
}

func generateRequestID() string {
    // Generate unique ID (UUID, etc.)
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

### Compression Middleware

Compress responses:

```go
func compressionMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check if client accepts compression
    acceptEncoding := ctx.GetHeader("Accept-Encoding")
    
    if !strings.Contains(acceptEncoding, "gzip") {
        return next(ctx)
    }
    
    // Set compression header
    ctx.SetHeader("Content-Encoding", "gzip")
    
    // Execute handler with compression
    // (implementation depends on response writer wrapping)
    return next(ctx)
}
```

## Advanced Middleware Techniques

### Parameterized Middleware

Create middleware factories that accept parameters:

```go
func requirePermission(permission string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        userID, _ := ctx.Get("user_id").(string)
        
        if !hasPermission(userID, permission) {
            return ctx.JSON(403, map[string]string{
                "error": "Permission denied",
            })
        }
        
        return next(ctx)
    }
}

// Usage
router.GET("/posts/:id/edit", editPostHandler, 
    authMiddleware,
    requirePermission("posts.edit"),
)

router.DELETE("/posts/:id", deletePostHandler,
    authMiddleware,
    requirePermission("posts.delete"),
)
```

### Context-Aware Middleware

Middleware that modifies context for downstream handlers:

```go
func tenantMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Extract tenant from subdomain or header
    host := ctx.Request().Host
    tenant := extractTenant(host)
    
    // Store tenant in context
    ctx.Set("tenant_id", tenant)
    
    // Set tenant-specific database connection
    // (implementation specific)
    
    return next(ctx)
}
```

### Conditional Middleware

Execute middleware based on conditions:

```go
func conditionalMiddleware(condition func(ctx pkg.Context) bool, mw pkg.MiddlewareFunc) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if condition(ctx) {
            return mw(ctx, next)
        }
        return next(ctx)
    }
}

// Usage
router.Use(conditionalMiddleware(
    func(ctx pkg.Context) bool {
        return strings.HasPrefix(ctx.Request().URL.Path, "/api/")
    },
    rateLimitMiddleware,
))
```

### Error Handling Middleware

Centralized error handling:

```go
func errorHandlerMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    err := next(ctx)
    
    if err != nil {
        // Log error
        log.Printf("Error: %v", err)
        
        // Return appropriate error response
        if apiErr, ok := err.(*APIError); ok {
            return ctx.JSON(apiErr.StatusCode, map[string]interface{}{
                "error":   apiErr.Message,
                "code":    apiErr.Code,
                "details": apiErr.Details,
            })
        }
        
        // Generic error response
        return ctx.JSON(500, map[string]string{
            "error": "Internal server error",
        })
    }
    
    return nil
}

type APIError struct {
    StatusCode int
    Code       string
    Message    string
    Details    interface{}
}

func (e *APIError) Error() string {
    return e.Message
}
```

## Best Practices

1. **Keep Middleware Focused:** Each middleware should do one thing well. Don't combine unrelated functionality.

2. **Order Matters:** Place middleware in logical order. For example, logging should come first, authentication before authorization.

3. **Always Call Next:** Unless you're intentionally short-circuiting, always call `next(ctx)`.

4. **Handle Errors Properly:** Return errors from middleware and let error handling middleware deal with them.

5. **Use Context for Data Sharing:** Store data in context using `ctx.Set()` for downstream middleware and handlers.

6. **Make Middleware Reusable:** Write middleware that can be used across different routes and applications.

7. **Document Middleware Behavior:** Clearly document what each middleware does and any context values it sets.

8. **Test Middleware Independently:** Write unit tests for middleware functions.

9. **Avoid Heavy Processing:** Keep middleware lightweight. Move heavy processing to handlers or background jobs.

10. **Use Middleware Groups:** Group related routes and apply shared middleware at the group level.

## Troubleshooting

### Middleware Not Executing

**Problem:** Middleware doesn't seem to run.

**Solutions:**
- Ensure middleware is registered before routes
- Check that `next(ctx)` is being called
- Verify middleware is added to the correct router/group
- Check for errors that might prevent execution

### Wrong Execution Order

**Problem:** Middleware executes in unexpected order.

**Solutions:**
- Remember: middleware executes in registration order
- Global middleware runs before route-specific middleware
- Group middleware runs before route middleware
- Check middleware registration order

### Context Values Not Available

**Problem:** Values set in middleware aren't available in handlers.

**Solutions:**
- Ensure middleware runs before the handler
- Use `ctx.Set()` to store values
- Use `ctx.Get()` to retrieve values
- Check for typos in context keys

### Response Already Written

**Problem:** Error about response already being written.

**Solutions:**
- Don't write responses in multiple places
- Check that middleware doesn't write response before calling `next()`
- Ensure only one response is sent per request

## See Also

- [Routing Guide](routing.md) - Route definition and groups
- [Context Guide](context.md) - Request context and data sharing
- [Security Guide](security.md) - Authentication and authorization
- [API Reference: Middleware](../api/framework.md) - Complete middleware API
- [Full Featured App Example](../examples/full-featured-app.md) - Complete working example with middleware
