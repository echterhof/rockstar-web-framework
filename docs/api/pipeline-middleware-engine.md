# Pipeline and Middleware Engine API

Complete API reference for advanced middleware management, request pipelines, and execution control in the Rockstar Web Framework.

## Overview

The framework provides two powerful systems for request processing:
- **MiddlewareEngine**: Advanced middleware management with priorities and conditional execution
- **PipelineEngine**: Request pipeline orchestration with chaining and multiplexing

## MiddlewareEngine Interface

Manages middleware execution with configurable ordering and conditional logic.

```go
type MiddlewareEngine interface {
    Register(config MiddlewareConfig) error
    Unregister(name string) error
    Enable(name string) error
    Disable(name string) error
    SetPriority(name string, priority int) error
    Get(name string) (*MiddlewareConfig, error)
    List() []MiddlewareConfig
    Execute(ctx Context, handler HandlerFunc) error
    Clear() error
}
```

### Methods

#### Register

```go
Register(config MiddlewareConfig) error
```

Registers a middleware with configuration.

**Parameters:**
- `config`: Middleware configuration

**Returns:**
- `error`: Error if registration fails

**Example:**
```go
engine := app.MiddlewareEngine()

config := pkg.MiddlewareConfig{
    Name: "auth",
    Handler: authMiddleware,
    Priority: 100,
    Enabled: true,
    Routes: []string{"/api/*"},
}

if err := engine.Register(config); err != nil {
    return err
}
```

#### Unregister

```go
Unregister(name string) error
```

Removes a middleware by name.

**Parameters:**
- `name`: Middleware name

**Returns:**
- `error`: Error if middleware not found

#### Enable

```go
Enable(name string) error
```

Enables a disabled middleware.

**Parameters:**
- `name`: Middleware name

**Returns:**
- `error`: Error if middleware not found

#### Disable

```go
Disable(name string) error
```

Disables a middleware without removing it.

**Parameters:**
- `name`: Middleware name

**Returns:**
- `error`: Error if middleware not found

**Example:**
```go
// Temporarily disable rate limiting
engine.Disable("rate-limit")

// Re-enable later
engine.Enable("rate-limit")
```

#### SetPriority

```go
SetPriority(name string, priority int) error
```

Changes middleware execution priority (lower = earlier).

**Parameters:**
- `name`: Middleware name
- `priority`: New priority value

**Returns:**
- `error`: Error if middleware not found

**Example:**
```go
// Ensure CORS runs before authentication
engine.SetPriority("cors", 10)
engine.SetPriority("auth", 20)
```

#### Get

```go
Get(name string) (*MiddlewareConfig, error)
```

Retrieves middleware configuration by name.

**Parameters:**
- `name`: Middleware name

**Returns:**
- `*MiddlewareConfig`: Middleware configuration
- `error`: Error if not found

#### List

```go
List() []MiddlewareConfig
```

Returns all registered middleware configurations.

**Returns:**
- `[]MiddlewareConfig`: List of middleware configs

#### Execute

```go
Execute(ctx Context, handler HandlerFunc) error
```

Executes middleware chain for a request.

**Parameters:**
- `ctx`: Request context
- `handler`: Final handler to execute

**Returns:**
- `error`: Error from middleware or handler

#### Clear

```go
Clear() error
```

Removes all registered middleware.

**Returns:**
- `error`: Error if clear fails

## MiddlewareConfig Type

Configuration for middleware registration.

```go
type MiddlewareConfig struct {
    Name        string          // Unique middleware name
    Handler     MiddlewareFunc  // Middleware function
    Priority    int             // Execution priority (lower = earlier)
    Enabled     bool            // Whether middleware is active
    Routes      []string        // Route patterns to apply to
    ExcludeRoutes []string      // Route patterns to exclude
    Methods     []string        // HTTP methods to apply to
    Condition   ConditionFunc   // Custom condition function
}
```

### Fields

#### Name

Unique identifier for the middleware.

#### Handler

The middleware function to execute.

#### Priority

Execution order (lower values execute first):
- 0-99: Infrastructure (CORS, security headers)
- 100-199: Authentication
- 200-299: Authorization
- 300-399: Request processing
- 400-499: Business logic
- 500+: Response processing

#### Enabled

Whether the middleware is currently active.

#### Routes

Route patterns where middleware applies (supports wildcards):
- `/api/*` - All API routes
- `/admin/*` - All admin routes
- `/users/:id` - Specific route pattern

#### ExcludeRoutes

Route patterns to exclude from middleware.

#### Methods

HTTP methods to apply middleware to (empty = all methods).

#### Condition

Custom function to determine if middleware should run:

```go
type ConditionFunc func(ctx Context) bool
```

## PipelineEngine Interface

Manages request pipelines with support for chaining and multiplexing.

```go
type PipelineEngine interface {
    Register(config PipelineConfig) error
    Unregister(name string) error
    Enable(name string) error
    Disable(name string) error
    SetPriority(name string, priority int) error
    Get(name string) (*PipelineConfig, error)
    List() []PipelineConfig
    Execute(ctx Context, pipelineName string) error
    ExecuteChain(ctx Context, pipelineNames []string) error
    ExecuteParallel(ctx Context, pipelineNames []string) error
    Clear() error
    Stats(name string) (*PipelineStats, error)
}
```

### Methods

#### Register

```go
Register(config PipelineConfig) error
```

Registers a pipeline with configuration.

**Parameters:**
- `config`: Pipeline configuration

**Returns:**
- `error`: Error if registration fails

**Example:**
```go
engine := app.PipelineEngine()

config := pkg.PipelineConfig{
    Name: "user-registration",
    Stages: []pkg.PipelineStage{
        {Name: "validate", Handler: validateUser},
        {Name: "create", Handler: createUser},
        {Name: "notify", Handler: sendWelcomeEmail},
    },
    OnError: handleRegistrationError,
}

engine.Register(config)
```

#### Execute

```go
Execute(ctx Context, pipelineName string) error
```

Executes a single pipeline.

**Parameters:**
- `ctx`: Request context
- `pipelineName`: Pipeline name

**Returns:**
- `error`: Error from pipeline execution

#### ExecuteChain

```go
ExecuteChain(ctx Context, pipelineNames []string) error
```

Executes multiple pipelines in sequence.

**Parameters:**
- `ctx`: Request context
- `pipelineNames`: Pipeline names to execute in order

**Returns:**
- `error`: Error from any pipeline

**Example:**
```go
// Execute pipelines in order
err := engine.ExecuteChain(ctx, []string{
    "validate-input",
    "process-payment",
    "create-order",
    "send-confirmation",
})
```

#### ExecuteParallel

```go
ExecuteParallel(ctx Context, pipelineNames []string) error
```

Executes multiple pipelines concurrently.

**Parameters:**
- `ctx`: Request context
- `pipelineNames`: Pipeline names to execute in parallel

**Returns:**
- `error`: Error from any pipeline

**Example:**
```go
// Execute independent pipelines concurrently
err := engine.ExecuteParallel(ctx, []string{
    "update-inventory",
    "send-notification",
    "log-analytics",
})
```

#### Stats

```go
Stats(name string) (*PipelineStats, error)
```

Returns execution statistics for a pipeline.

**Parameters:**
- `name`: Pipeline name

**Returns:**
- `*PipelineStats`: Pipeline statistics
- `error`: Error if pipeline not found

## PipelineConfig Type

Configuration for pipeline registration.

```go
type PipelineConfig struct {
    Name        string           // Unique pipeline name
    Stages      []PipelineStage  // Pipeline stages
    Priority    int              // Execution priority
    Enabled     bool             // Whether pipeline is active
    Timeout     time.Duration    // Execution timeout
    OnError     ErrorHandler     // Error handler
    OnComplete  CompleteHandler  // Completion handler
}
```

## PipelineStage Type

Represents a single stage in a pipeline.

```go
type PipelineStage struct {
    Name        string        // Stage name
    Handler     HandlerFunc   // Stage handler
    Timeout     time.Duration // Stage timeout
    Required    bool          // Whether stage must succeed
    Retry       int           // Number of retries on failure
    RetryDelay  time.Duration // Delay between retries
}
```

## PipelineStats Type

Statistics for pipeline execution.

```go
type PipelineStats struct {
    Name            string
    ExecutionCount  int64
    SuccessCount    int64
    ErrorCount      int64
    AverageDuration time.Duration
    LastExecution   time.Time
    LastError       error
}
```

## Complete Middleware Example

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    engine := app.MiddlewareEngine()
    
    // CORS middleware (runs first)
    engine.Register(pkg.MiddlewareConfig{
        Name: "cors",
        Priority: 10,
        Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
            ctx.SetHeader("Access-Control-Allow-Origin", "*")
            ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
            return next(ctx)
        },
    })
    
    // Authentication middleware
    engine.Register(pkg.MiddlewareConfig{
        Name: "auth",
        Priority: 100,
        Routes: []string{"/api/*"},
        ExcludeRoutes: []string{"/api/public/*"},
        Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
            token := ctx.GetHeader("Authorization")
            if token == "" {
                return pkg.NewAuthenticationError("Missing token")
            }
            
            user, err := validateToken(token)
            if err != nil {
                return pkg.NewAuthenticationError("Invalid token")
            }
            
            ctx.Set("user", user)
            return next(ctx)
        },
    })
    
    // Rate limiting middleware
    engine.Register(pkg.MiddlewareConfig{
        Name: "rate-limit",
        Priority: 50,
        Routes: []string{"/api/*"},
        Condition: func(ctx pkg.Context) bool {
            // Only rate limit non-admin users
            user, _ := ctx.Get("user")
            if u, ok := user.(*User); ok {
                return !u.IsAdmin
            }
            return true
        },
        Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
            allowed, err := checkRateLimit(ctx)
            if err != nil {
                return err
            }
            if !allowed {
                return pkg.NewRateLimitError(60 * time.Second)
            }
            return next(ctx)
        },
    })
    
    // Logging middleware
    engine.Register(pkg.MiddlewareConfig{
        Name: "logging",
        Priority: 1,
        Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
            start := time.Now()
            
            err := next(ctx)
            
            duration := time.Since(start)
            ctx.Logger().Info("Request completed",
                "path", ctx.Request().URL.Path,
                "method", ctx.Request().Method,
                "duration", duration,
                "error", err,
            )
            
            return err
        },
    })
    
    app.Start(":8080")
}
```

## Complete Pipeline Example

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    engine := app.PipelineEngine()
    
    // Order processing pipeline
    engine.Register(pkg.PipelineConfig{
        Name: "process-order",
        Timeout: 30 * time.Second,
        Stages: []pkg.PipelineStage{
            {
                Name: "validate",
                Handler: func(ctx pkg.Context) error {
                    // Validate order data
                    return validateOrder(ctx)
                },
                Required: true,
            },
            {
                Name: "check-inventory",
                Handler: func(ctx pkg.Context) error {
                    // Check product availability
                    return checkInventory(ctx)
                },
                Required: true,
                Retry: 3,
                RetryDelay: 1 * time.Second,
            },
            {
                Name: "process-payment",
                Handler: func(ctx pkg.Context) error {
                    // Process payment
                    return processPayment(ctx)
                },
                Required: true,
                Timeout: 10 * time.Second,
            },
            {
                Name: "create-order",
                Handler: func(ctx pkg.Context) error {
                    // Create order in database
                    return createOrder(ctx)
                },
                Required: true,
            },
            {
                Name: "send-confirmation",
                Handler: func(ctx pkg.Context) error {
                    // Send confirmation email
                    return sendConfirmation(ctx)
                },
                Required: false, // Non-critical
            },
        },
        OnError: func(ctx pkg.Context, err error) error {
            // Rollback on error
            rollbackOrder(ctx)
            return err
        },
        OnComplete: func(ctx pkg.Context) error {
            // Log completion
            ctx.Logger().Info("Order processed successfully")
            return nil
        },
    })
    
    // Route handler
    app.Router().POST("/orders", func(ctx pkg.Context) error {
        // Execute order processing pipeline
        if err := engine.Execute(ctx, "process-order"); err != nil {
            return err
        }
        
        return ctx.JSON(201, map[string]string{
            "message": "Order created successfully",
        })
    })
    
    app.Start(":8080")
}
```

## Best Practices

1. **Use priorities wisely**: Order middleware logically (security → auth → business logic)
2. **Enable/disable dynamically**: Toggle middleware based on environment or feature flags
3. **Use conditions**: Apply middleware conditionally to reduce overhead
4. **Set timeouts**: Prevent pipelines from running indefinitely
5. **Handle errors**: Always implement error handlers for pipelines
6. **Monitor performance**: Track pipeline statistics for optimization
7. **Keep stages focused**: Each pipeline stage should have a single responsibility
8. **Use parallel execution**: Run independent pipelines concurrently for better performance

## See Also

- [Middleware Guide](../guides/middleware.md) - Middleware concepts and patterns
- [Router API](router.md) - Route-level middleware
- [Context API](context.md) - Request context
- [Performance Guide](../guides/performance.md) - Performance optimization
