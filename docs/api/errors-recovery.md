# Error Handling and Recovery API

Complete API reference for error handling, panic recovery, and error formatting in the Rockstar Web Framework.

## Overview

The framework provides comprehensive error handling with:
- Structured error types with i18n support
- Automatic panic recovery
- Custom error handlers
- HTTP status code mapping
- Error logging and tracking

## ErrorHandler Interface

Defines error handling behavior for the framework.

```go
type ErrorHandler interface {
    HandleError(ctx Context, err error) error
    HandlePanic(ctx Context, recovered interface{}) error
    FormatError(err error) map[string]interface{}
}
```

### Methods

#### HandleError

```go
HandleError(ctx Context, err error) error
```

Handles errors returned from handlers and middleware.

**Parameters:**
- `ctx`: Request context
- `err`: Error to handle

**Returns:**
- `error`: Processed error (may be transformed)

**Example:**
```go
type CustomErrorHandler struct{}

func (h *CustomErrorHandler) HandleError(ctx pkg.Context, err error) error {
    // Log error
    ctx.Logger().Error("Request error", "error", err, "path", ctx.Request().URL.Path)
    
    // Check if it's a framework error
    if fwErr, ok := err.(*pkg.FrameworkError); ok {
        return ctx.JSON(fwErr.StatusCode, map[string]interface{}{
            "error": fwErr.Message,
            "code": fwErr.Code,
        })
    }
    
    // Generic error
    return ctx.JSON(500, map[string]string{
        "error": "Internal server error",
    })
}

// Register custom error handler
app.SetErrorHandler(&CustomErrorHandler{})
```

#### HandlePanic

```go
HandlePanic(ctx Context, recovered interface{}) error
```

Handles panics that occur during request processing.

**Parameters:**
- `ctx`: Request context
- `recovered`: Value recovered from panic

**Returns:**
- `error`: Error to return to client

**Example:**
```go
func (h *CustomErrorHandler) HandlePanic(ctx pkg.Context, recovered interface{}) error {
    // Log panic with stack trace
    ctx.Logger().Error("Panic recovered", 
        "panic", recovered,
        "path", ctx.Request().URL.Path,
        "stack", string(debug.Stack()),
    )
    
    // Return generic error to client
    return ctx.JSON(500, map[string]string{
        "error": "An unexpected error occurred",
    })
}
```

#### FormatError

```go
FormatError(err error) map[string]interface{}
```

Formats an error for JSON response.

**Parameters:**
- `err`: Error to format

**Returns:**
- `map[string]interface{}`: Formatted error data

**Example:**
```go
func (h *CustomErrorHandler) FormatError(err error) map[string]interface{} {
    if fwErr, ok := err.(*pkg.FrameworkError); ok {
        return map[string]interface{}{
            "error": fwErr.Message,
            "code": fwErr.Code,
            "status": fwErr.StatusCode,
            "i18n_key": fwErr.I18nKey,
        }
    }
    
    return map[string]interface{}{
        "error": err.Error(),
        "status": 500,
    }
}
```

## RecoveryHandler Interface

Handles panic recovery with customizable behavior.

```go
type RecoveryHandler interface {
    Recover(ctx Context, recovered interface{}) error
    ShouldRecover(recovered interface{}) bool
}
```

### Methods

#### Recover

```go
Recover(ctx Context, recovered interface{}) error
```

Recovers from a panic and returns an appropriate error.

**Parameters:**
- `ctx`: Request context
- `recovered`: Value recovered from panic

**Returns:**
- `error`: Error to return

#### ShouldRecover

```go
ShouldRecover(recovered interface{}) bool
```

Determines if a panic should be recovered or re-panicked.

**Parameters:**
- `recovered`: Value recovered from panic

**Returns:**
- `bool`: True if panic should be recovered

**Example:**
```go
type CustomRecoveryHandler struct{}

func (h *CustomRecoveryHandler) ShouldRecover(recovered interface{}) bool {
    // Don't recover from specific panic types
    if msg, ok := recovered.(string); ok {
        if strings.Contains(msg, "FATAL") {
            return false // Re-panic
        }
    }
    return true
}

func (h *CustomRecoveryHandler) Recover(ctx pkg.Context, recovered interface{}) error {
    // Log with context
    ctx.Logger().Error("Panic recovered",
        "panic", recovered,
        "user_id", ctx.User().ID,
        "tenant_id", ctx.Tenant().ID,
    )
    
    return pkg.NewFrameworkError(
        "PANIC_RECOVERED",
        "An unexpected error occurred",
        500,
    )
}
```

## FrameworkError Type

Structured error type with HTTP status codes and i18n support.

```go
type FrameworkError struct {
    Code       string                 // Error code (e.g., "VALIDATION_ERROR")
    Message    string                 // Human-readable message
    StatusCode int                    // HTTP status code
    I18nKey    string                 // i18n translation key
    I18nParams map[string]interface{} // i18n parameters
    Cause      error                  // Underlying error
}
```

### Constructor

```go
func NewFrameworkError(code, message string, statusCode int) *FrameworkError
```

Creates a new framework error.

**Example:**
```go
err := pkg.NewFrameworkError(
    "USER_NOT_FOUND",
    "User not found",
    404,
)
```

### Methods

#### Error

```go
func (e *FrameworkError) Error() string
```

Returns the error message (implements error interface).

#### WithCause

```go
func (e *FrameworkError) WithCause(cause error) *FrameworkError
```

Adds an underlying cause to the error.

**Example:**
```go
err := pkg.NewFrameworkError("DB_ERROR", "Database error", 500).
    WithCause(dbErr)
```

#### WithI18n

```go
func (e *FrameworkError) WithI18n(key string, params map[string]interface{}) *FrameworkError
```

Adds i18n translation information.

**Example:**
```go
err := pkg.NewFrameworkError("VALIDATION_ERROR", "Validation failed", 400).
    WithI18n("errors.validation.required", map[string]interface{}{
        "field": "email",
    })
```

#### Unwrap

```go
func (e *FrameworkError) Unwrap() error
```

Returns the underlying cause (for error wrapping).

## Common Error Codes

The framework defines standard error codes:

```go
const (
    ErrCodeValidation      = "VALIDATION_ERROR"
    ErrCodeAuthentication  = "AUTHENTICATION_ERROR"
    ErrCodeAuthorization   = "AUTHORIZATION_ERROR"
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeConflict        = "CONFLICT"
    ErrCodeRateLimit       = "RATE_LIMIT_EXCEEDED"
    ErrCodeInternal        = "INTERNAL_ERROR"
    ErrCodeBadRequest      = "BAD_REQUEST"
    ErrCodeTimeout         = "TIMEOUT"
    ErrCodeUnavailable     = "SERVICE_UNAVAILABLE"
)
```

## Error Creation Helpers

### Validation Errors

```go
func NewValidationError(message string) *FrameworkError
```

Creates a validation error (400).

**Example:**
```go
if email == "" {
    return pkg.NewValidationError("Email is required")
}
```

### Authentication Errors

```go
func NewAuthenticationError(message string) *FrameworkError
```

Creates an authentication error (401).

**Example:**
```go
if !validToken {
    return pkg.NewAuthenticationError("Invalid token")
}
```

### Authorization Errors

```go
func NewAuthorizationError(message string) *FrameworkError
```

Creates an authorization error (403).

**Example:**
```go
if !ctx.IsAuthorized("users", "delete") {
    return pkg.NewAuthorizationError("Insufficient permissions")
}
```

### Not Found Errors

```go
func NewNotFoundError(resource string) *FrameworkError
```

Creates a not found error (404).

**Example:**
```go
user, err := db.FindUser(id)
if err != nil {
    return pkg.NewNotFoundError("User")
}
```

### Rate Limit Errors

```go
func NewRateLimitError(retryAfter time.Duration) *FrameworkError
```

Creates a rate limit error (429).

**Example:**
```go
if !allowed {
    return pkg.NewRateLimitError(60 * time.Second)
}
```

## Error Middleware

Built-in middleware for error handling:

```go
func ErrorMiddleware(ctx Context, next HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            // Handle panic
            ctx.Logger().Error("Panic", "recovered", r)
            ctx.JSON(500, map[string]string{
                "error": "Internal server error",
            })
        }
    }()
    
    err := next(ctx)
    if err != nil {
        // Handle error
        if fwErr, ok := err.(*pkg.FrameworkError); ok {
            return ctx.JSON(fwErr.StatusCode, map[string]interface{}{
                "error": fwErr.Message,
                "code": fwErr.Code,
            })
        }
        return ctx.JSON(500, map[string]string{
            "error": err.Error(),
        })
    }
    
    return nil
}

// Register globally
app.Use(ErrorMiddleware)
```

## Complete Example

```go
package main

import (
    "database/sql"
    "errors"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

// Custom error handler
type AppErrorHandler struct {
    logger pkg.Logger
}

func (h *AppErrorHandler) HandleError(ctx pkg.Context, err error) error {
    // Log all errors
    h.logger.Error("Request error",
        "error", err,
        "path", ctx.Request().URL.Path,
        "method", ctx.Request().Method,
        "user_id", ctx.User().ID,
    )
    
    // Handle framework errors
    if fwErr, ok := err.(*pkg.FrameworkError); ok {
        response := map[string]interface{}{
            "success": false,
            "error": fwErr.Message,
            "code": fwErr.Code,
        }
        
        // Add i18n translation if available
        if fwErr.I18nKey != "" {
            response["message"] = ctx.I18n().Translate(fwErr.I18nKey, fwErr.I18nParams)
        }
        
        return ctx.JSON(fwErr.StatusCode, response)
    }
    
    // Handle database errors
    if errors.Is(err, sql.ErrNoRows) {
        return ctx.JSON(404, map[string]string{
            "error": "Resource not found",
        })
    }
    
    // Generic error
    return ctx.JSON(500, map[string]string{
        "error": "An unexpected error occurred",
    })
}

func (h *AppErrorHandler) HandlePanic(ctx pkg.Context, recovered interface{}) error {
    h.logger.Error("Panic recovered",
        "panic", recovered,
        "path", ctx.Request().URL.Path,
    )
    
    return ctx.JSON(500, map[string]string{
        "error": "Internal server error",
    })
}

func (h *AppErrorHandler) FormatError(err error) map[string]interface{} {
    if fwErr, ok := err.(*pkg.FrameworkError); ok {
        return map[string]interface{}{
            "error": fwErr.Message,
            "code": fwErr.Code,
            "status": fwErr.StatusCode,
        }
    }
    
    return map[string]interface{}{
        "error": err.Error(),
        "status": 500,
    }
}

// Handler with error handling
func getUserHandler(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Validate input
    if userID == "" {
        return pkg.NewValidationError("User ID is required")
    }
    
    // Check authentication
    if !ctx.IsAuthenticated() {
        return pkg.NewAuthenticationError("Authentication required")
    }
    
    // Check authorization
    if !ctx.IsAuthorized("users", "read") {
        return pkg.NewAuthorizationError("Insufficient permissions")
    }
    
    // Query database
    user, err := ctx.DB().QueryRow("SELECT * FROM users WHERE id = ?", userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return pkg.NewNotFoundError("User")
        }
        return pkg.NewFrameworkError("DB_ERROR", "Database error", 500).
            WithCause(err)
    }
    
    return ctx.JSON(200, user)
}

func main() {
    config := pkg.FrameworkConfig{}
    app, err := pkg.New(config)
    if err != nil {
        panic(err)
    }
    
    // Register custom error handler
    app.SetErrorHandler(&AppErrorHandler{
        logger: app.Logger(),
    })
    
    // Register routes
    app.Router().GET("/users/:id", getUserHandler)
    
    // Start server
    app.Start(":8080")
}
```

## Error Logging

Errors are automatically logged with context:

```go
// Framework logs errors with:
// - Error message and code
// - Request path and method
// - User and tenant IDs (if available)
// - Stack trace (for panics)
// - Timestamp and request ID
```

## Best Practices

1. **Use framework errors**: Create `FrameworkError` instances for structured errors
2. **Add context**: Include relevant information in error messages
3. **Log appropriately**: Log errors at appropriate levels (error, warn, info)
4. **Don't expose internals**: Don't leak internal details to clients
5. **Use error codes**: Use consistent error codes for client handling
6. **Wrap errors**: Use `WithCause()` to preserve error chains
7. **Handle panics**: Always recover from panics in production
8. **Translate errors**: Use i18n for user-facing error messages

## See Also

- [Context API](context.md) - Request context
- [Logger API](utilities.md#logger) - Logging interface
- [Security Guide](../guides/security.md) - Error handling security
- [Monitoring API](monitoring.md) - Error tracking and metrics
