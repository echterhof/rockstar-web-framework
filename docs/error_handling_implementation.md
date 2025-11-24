# Error Handling Implementation

## Overview

The Rockstar Web Framework provides comprehensive error handling with internationalization support, graceful recovery mechanisms, and integration across all components. The error handling system is designed to provide consistent, user-friendly error messages while maintaining detailed error information for debugging.

## Features

- **Internationalized Error Messages**: All error messages support i18n with parameter interpolation
- **Error Codes**: Standardized error codes for programmatic error handling
- **Error Chaining**: Support for wrapping and unwrapping errors
- **Context Information**: Automatic capture of request context (request ID, tenant, user, path, method)
- **Graceful Recovery**: Built-in panic recovery with proper error conversion
- **Structured Error Responses**: Consistent JSON error response format
- **Development vs Production**: Different error detail levels based on environment

## Core Components

### FrameworkError

The `FrameworkError` struct is the foundation of the error handling system:

```go
type FrameworkError struct {
    Code       string                 // Error code for programmatic handling
    Message    string                 // Human-readable error message
    Details    map[string]interface{} // Additional error details
    StatusCode int                    // HTTP status code
    Cause      error                  // Underlying cause error
    
    // Internationalization support
    I18nKey    string                 // Translation key
    I18nParams map[string]interface{} // Translation parameters
    
    // Request context
    RequestID  string
    TenantID   string
    UserID     string
    Path       string
    Method     string
}
```

### ErrorHandler

The `ErrorHandler` interface provides methods for handling errors:

```go
type ErrorHandler interface {
    HandleError(ctx Context, err error) error
    HandlePanic(ctx Context, recovered interface{}) error
    FormatError(ctx Context, err *FrameworkError) interface{}
}
```

### RecoveryHandler

The `RecoveryHandler` interface provides panic recovery:

```go
type RecoveryHandler interface {
    Recover(ctx Context, recovered interface{}) error
    ShouldRecover(recovered interface{}) bool
}
```

## Error Codes

The framework defines standardized error codes for common scenarios:

### Authentication Errors
- `AUTH_FAILED`: Authentication failed
- `INVALID_TOKEN`: Invalid authentication token
- `TOKEN_EXPIRED`: Authentication token expired
- `UNAUTHORIZED`: Unauthorized access

### Authorization Errors
- `FORBIDDEN`: Access forbidden
- `INSUFFICIENT_ROLES`: Insufficient roles
- `INSUFFICIENT_ACTIONS`: Insufficient permissions

### Validation Errors
- `VALIDATION_FAILED`: Validation failed
- `INVALID_INPUT`: Invalid input
- `MISSING_FIELD`: Required field missing
- `INVALID_FORMAT`: Invalid format
- `FILE_TOO_LARGE`: File too large
- `INVALID_FILE_TYPE`: Invalid file type

### Request Errors
- `REQUEST_TOO_LARGE`: Request size exceeds limit
- `REQUEST_TIMEOUT`: Request timeout exceeded
- `BOGUS_DATA`: Invalid or malformed data
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded

### Security Errors
- `CSRF_TOKEN_INVALID`: CSRF token invalid
- `XSS_DETECTED`: XSS attack detected
- `SQL_INJECTION_DETECTED`: SQL injection detected

### Database Errors
- `DATABASE_CONNECTION`: Database connection failed
- `DATABASE_QUERY`: Database query failed
- `DATABASE_TRANSACTION`: Transaction failed
- `RECORD_NOT_FOUND`: Record not found
- `DUPLICATE_RECORD`: Duplicate record

### Session Errors
- `SESSION_NOT_FOUND`: Session not found
- `SESSION_EXPIRED`: Session expired
- `SESSION_INVALID`: Invalid session

### Server Errors
- `INTERNAL_ERROR`: Internal server error
- `SERVICE_UNAVAILABLE`: Service unavailable
- `NOT_IMPLEMENTED`: Not implemented

## Usage

### Creating Errors

Use the provided constructor functions to create errors:

```go
// Authentication error
err := pkg.NewAuthenticationError("Invalid credentials")

// Validation error with field
err := pkg.NewValidationError("Invalid email", "email")

// Rate limit error with parameters
err := pkg.NewRateLimitError("Too many requests", 100, "1m")

// Request timeout error
err := pkg.NewRequestTimeoutError(30 * time.Second)

// Missing field error
err := pkg.NewMissingFieldError("username")

// Not found error
err := pkg.NewNotFoundError("user")
```

### Adding Context

Add request context to errors:

```go
err := pkg.NewAuthenticationError("Invalid token")
err.WithContext(requestID, tenantID, userID, path, method)
```

### Adding Details

Add additional details to errors:

```go
err := pkg.NewValidationError("Invalid input", "email")
err.WithDetails(map[string]interface{}{
    "provided": "invalid-email",
    "expected": "user@example.com",
})
```

### Error Chaining

Chain errors to preserve the original cause:

```go
baseErr := errors.New("connection refused")
err := pkg.NewDatabaseError("Failed to connect", "CONNECT").WithCause(baseErr)

// Unwrap to get the cause
cause := err.Unwrap() // Returns baseErr
```

### Wrapping Generic Errors

Convert generic errors to FrameworkErrors:

```go
genericErr := errors.New("something went wrong")
fwErr := pkg.WrapError(genericErr, "CUSTOM_CODE", http.StatusInternalServerError)
```

### Error Handler Setup

Initialize the error handler with i18n support:

```go
// Create i18n manager
i18nManager, err := pkg.NewI18nManager(pkg.I18nConfig{
    DefaultLocale:     "en",
    LocalesDir:        "./locales",
    SupportedLocales:  []string{"en", "de"},
    FallbackToDefault: true,
})

// Create error handler
errorHandler := pkg.NewErrorHandler(pkg.ErrorHandlerConfig{
    I18n:         i18nManager,
    Logger:       logger,
    IncludeStack: true,
    LogErrors:    true,
})
```

### Handling Errors

Handle errors in request handlers:

```go
func handler(ctx Context) error {
    // Business logic
    if err := someOperation(); err != nil {
        return errorHandler.HandleError(ctx, err)
    }
    
    return nil
}
```

### Recovery Middleware

Use recovery middleware to catch panics:

```go
// Create recovery middleware
recoveryMW := pkg.RecoveryMiddleware(errorHandler)

// Apply to routes
router.Use(recoveryMW)
```

### Error Middleware

Use error middleware to handle errors:

```go
// Create error middleware
errorMW := pkg.ErrorMiddleware(errorHandler)

// Apply to routes
router.Use(errorMW)
```

## Internationalization

### Locale Files

Create locale files with error translations:

**locales.error.en.yaml:**
```yaml
error:
  authentication:
    failed: "Authentication failed"
    invalid_token: "Invalid authentication token"
  
  validation:
    missing_field: "Required field '{{field}}' is missing"
    invalid_format: "Field '{{field}}' has invalid format, expected: {{format}}"
```

**locales.error.de.yaml:**
```yaml
error:
  authentication:
    failed: "Authentifizierung fehlgeschlagen"
    invalid_token: "Ungültiges Authentifizierungstoken"
  
  validation:
    missing_field: "Erforderliches Feld '{{field}}' fehlt"
    invalid_format: "Feld '{{field}}' hat ungültiges Format, erwartet: {{format}}"
```

### Translating Errors

Errors are automatically translated when handled:

```go
// Create error with i18n key
err := pkg.NewMissingFieldError("username")

// Error handler translates based on context language
errorHandler.HandleError(ctx, err)
```

## Error Response Format

### Standard Response

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Required field 'username' is missing",
    "details": {
      "field": "username"
    }
  }
}
```

### Development Mode Response

In development mode, additional information is included:

```json
{
  "error": {
    "code": "DATABASE_QUERY",
    "message": "Database query failed",
    "details": {
      "operation": "SELECT"
    },
    "request_id": "req-123",
    "path": "/api/users",
    "method": "GET",
    "cause": "connection refused"
  }
}
```

## Best Practices

1. **Use Specific Error Constructors**: Use the provided constructor functions instead of creating errors manually
2. **Add Context**: Always add request context to errors for better debugging
3. **Chain Errors**: Preserve the original error cause when wrapping errors
4. **Use Error Codes**: Use standardized error codes for programmatic error handling
5. **Provide Details**: Add relevant details to help with debugging
6. **Translate Messages**: Use i18n keys for all user-facing error messages
7. **Log Errors**: Enable error logging in production for monitoring
8. **Handle Panics**: Use recovery middleware to catch and handle panics gracefully

## Integration with Components

### Router Integration

```go
router.Use(pkg.RecoveryMiddleware(errorHandler))
router.Use(pkg.ErrorMiddleware(errorHandler))
```

### Middleware Integration

```go
func authMiddleware(errorHandler ErrorHandler) MiddlewareFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx Context) error {
            token := ctx.GetHeader("Authorization")
            if token == "" {
                return errorHandler.HandleError(ctx, 
                    pkg.NewAuthenticationError("Missing token"))
            }
            
            return next(ctx)
        }
    }
}
```

### Database Integration

```go
func getUser(ctx Context, id string) (*User, error) {
    user, err := db.FindByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, pkg.NewNotFoundError("user")
        }
        return nil, pkg.NewDatabaseError("Failed to query user", "SELECT").
            WithCause(err)
    }
    
    return user, nil
}
```

## Testing

The error handling system includes comprehensive unit tests:

```bash
go test -v ./pkg -run TestFrameworkError
go test -v ./pkg -run TestErrorHandler
go test -v ./pkg -run TestRecoveryHandler
```

## Example

See `examples/error_handling_example.go` for a complete example demonstrating all error handling features.

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 18.3**: Built-in recovery mechanisms for graceful error handling
- **Requirement 18.4**: Graceful error handling with internationalized messages
- **Requirement 5.1-5.5**: Internationalization support for error messages
- **Requirement 6.1-6.10**: Security validation with proper error responses

## Performance Considerations

- Error creation is lightweight with minimal allocations
- Error translation is cached by the i18n manager
- Context information is added lazily
- Error formatting is optimized for common cases
