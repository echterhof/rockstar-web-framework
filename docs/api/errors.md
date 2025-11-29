# Error Handling API Reference

## Overview

The Rockstar Web Framework provides a comprehensive error handling system with internationalization support, structured error responses, panic recovery, and detailed error tracking. All errors implement standard Go error interfaces while providing additional framework-specific features.

## FrameworkError

The main error type used throughout the framework.

### Type Definition

```go
type FrameworkError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    StatusCode int                    `json:"-"`
    Cause      error                  `json:"-"`
    
    // Internationalization support
    I18nKey    string                 `json:"-"`
    I18nParams map[string]interface{} `json:"-"`
    
    // Request context
    RequestID string `json:"request_id,omitempty"`
    TenantID  string `json:"tenant_id,omitempty"`
    UserID    string `json:"user_id,omitempty"`
    Path      string `json:"path,omitempty"`
    Method    string `json:"method,omitempty"`
}
```

### Methods

#### Error()

Implements the error interface. Returns a formatted error string.

**Signature:**
```go
func (e *FrameworkError) Error() string
```

**Example:**
```go
err := pkg.NewValidationError("Invalid email", "email")
fmt.Println(err.Error())
// Output: "VALIDATION_FAILED: Invalid email"
```

#### Unwrap()

Returns the underlying cause error for error chain unwrapping.

**Signature:**
```go
func (e *FrameworkError) Unwrap() error
```

**Example:**
```go
dbErr := errors.New("connection failed")
err := pkg.NewDatabaseError("Query failed", "SELECT").WithCause(dbErr)
cause := errors.Unwrap(err)  // Returns dbErr
```

#### WithCause()

Adds a cause error to the error chain.

**Signature:**
```go
func (e *FrameworkError) WithCause(cause error) *FrameworkError
```

**Parameters:**
- `cause` - Underlying error that caused this error

**Returns:**
- `*FrameworkError` - The error instance (for chaining)

**Example:**
```go
dbErr := sql.ErrNoRows
err := pkg.NewDatabaseError("User not found", "SELECT").
    WithCause(dbErr)
```

#### WithDetails()

Adds additional details to the error.

**Signature:**
```go
func (e *FrameworkError) WithDetails(details map[string]interface{}) *FrameworkError
```

**Parameters:**
- `details` - Map of additional error details

**Returns:**
- `*FrameworkError` - The error instance (for chaining)

**Example:**
```go
err := pkg.NewValidationError("Invalid input", "age").
    WithDetails(map[string]interface{}{
        "min": 18,
        "max": 100,
        "provided": 150,
    })
```

#### WithContext()

Adds request context information to the error.

**Signature:**
```go
func (e *FrameworkError) WithContext(requestID, tenantID, userID, path, method string) *FrameworkError
```

**Parameters:**
- `requestID` - Request ID
- `tenantID` - Tenant ID
- `userID` - User ID
- `path` - Request path
- `method` - HTTP method

**Returns:**
- `*FrameworkError` - The error instance (for chaining)

**Example:**
```go
err := pkg.NewAuthenticationError("Invalid credentials").
    WithContext(req.ID, req.TenantID, req.UserID, req.Path, req.Method)
```

## Error Codes

### Authentication Errors

```go
const (
    ErrCodeAuthenticationFailed = "AUTH_FAILED"
    ErrCodeInvalidToken         = "INVALID_TOKEN"
    ErrCodeTokenExpired         = "TOKEN_EXPIRED"
    ErrCodeUnauthorized         = "UNAUTHORIZED"
)
```

### Authorization Errors

```go
const (
    ErrCodeForbidden           = "FORBIDDEN"
    ErrCodeInsufficientRoles   = "INSUFFICIENT_ROLES"
    ErrCodeInsufficientActions = "INSUFFICIENT_ACTIONS"
    ErrCodeInsufficientScopes  = "INSUFFICIENT_SCOPES"
)
```

### Validation Errors

```go
const (
    ErrCodeValidationFailed = "VALIDATION_FAILED"
    ErrCodeInvalidInput     = "INVALID_INPUT"
    ErrCodeMissingField     = "MISSING_FIELD"
    ErrCodeInvalidFormat    = "INVALID_FORMAT"
    ErrCodeFileTooLarge     = "FILE_TOO_LARGE"
    ErrCodeInvalidFileType  = "INVALID_FILE_TYPE"
)
```

### Request Errors

```go
const (
    ErrCodeRequestTooLarge   = "REQUEST_TOO_LARGE"
    ErrCodeRequestTimeout    = "REQUEST_TIMEOUT"
    ErrCodeBogusData         = "BOGUS_DATA"
    ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)
```

### Security Errors

```go
const (
    ErrCodeCSRFTokenInvalid     = "CSRF_TOKEN_INVALID"
    ErrCodeXSSDetected          = "XSS_DETECTED"
    ErrCodeSQLInjectionDetected = "SQL_INJECTION_DETECTED"
    ErrCodePathTraversal        = "PATH_TRAVERSAL"
    ErrCodeRegexTimeout         = "REGEX_TIMEOUT"
)
```

### Database Errors

```go
const (
    ErrCodeDatabaseConnection   = "DATABASE_CONNECTION"
    ErrCodeDatabaseQuery        = "DATABASE_QUERY"
    ErrCodeDatabaseTransaction  = "DATABASE_TRANSACTION"
    ErrCodeRecordNotFound       = "RECORD_NOT_FOUND"
    ErrCodeDuplicateRecord      = "DUPLICATE_RECORD"
    ErrCodeNoDatabaseConfigured = "NO_DATABASE_CONFIGURED"
)
```

### Session Errors

```go
const (
    ErrCodeSessionNotFound = "SESSION_NOT_FOUND"
    ErrCodeSessionExpired  = "SESSION_EXPIRED"
    ErrCodeSessionInvalid  = "SESSION_INVALID"
)
```

### Configuration Errors

```go
const (
    ErrCodeConfigurationError   = "CONFIGURATION_ERROR"
    ErrCodeMissingConfiguration = "MISSING_CONFIGURATION"
)
```

### Server Errors

```go
const (
    ErrCodeInternalError      = "INTERNAL_ERROR"
    ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
    ErrCodeNotImplemented     = "NOT_IMPLEMENTED"
)
```

### Multi-Tenancy Errors

```go
const (
    ErrCodeTenantNotFound      = "TENANT_NOT_FOUND"
    ErrCodeTenantInactive      = "TENANT_INACTIVE"
    ErrCodeTenantLimitExceeded = "TENANT_LIMIT_EXCEEDED"
)
```

### WebSocket Errors

```go
const (
    ErrCodeWebSocketUpgradeFailed    = "WEBSOCKET_UPGRADE_FAILED"
    ErrCodeWebSocketConnectionClosed = "WEBSOCKET_CONNECTION_CLOSED"
    ErrCodeWebSocketAuthRequired     = "WEBSOCKET_AUTH_REQUIRED"
    ErrCodeWebSocketInvalidMessage   = "WEBSOCKET_INVALID_MESSAGE"
)
```

### Routing Errors

```go
const (
    ErrCodeNotFound            = "NOT_FOUND"
    ErrCodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
    ErrCodeAuthorizationFailed = "AUTHORIZATION_FAILED"
)
```

## Error Constructors

### NewAuthenticationError()

Creates an authentication error (401 Unauthorized).

**Signature:**
```go
func NewAuthenticationError(message string) *FrameworkError
```

**Example:**
```go
return pkg.NewAuthenticationError("Invalid credentials")
```

### NewAuthorizationError()

Creates an authorization error (403 Forbidden).

**Signature:**
```go
func NewAuthorizationError(message string) *FrameworkError
```

**Example:**
```go
return pkg.NewAuthorizationError("Insufficient permissions")
```

### NewValidationError()

Creates a validation error (400 Bad Request).

**Signature:**
```go
func NewValidationError(message string, field string) *FrameworkError
```

**Parameters:**
- `message` - Error message
- `field` - Field name that failed validation

**Example:**
```go
return pkg.NewValidationError("Email format is invalid", "email")
```

### NewRateLimitError()

Creates a rate limit error (429 Too Many Requests).

**Signature:**
```go
func NewRateLimitError(message string, limit int, window string) *FrameworkError
```

**Parameters:**
- `message` - Error message
- `limit` - Rate limit threshold
- `window` - Time window (e.g., "1m", "1h")

**Example:**
```go
return pkg.NewRateLimitError("Rate limit exceeded", 100, "1m")
```

### NewDatabaseError()

Creates a database error (500 Internal Server Error).

**Signature:**
```go
func NewDatabaseError(message string, operation string) *FrameworkError
```

**Parameters:**
- `message` - Error message
- `operation` - Database operation (e.g., "SELECT", "INSERT")

**Example:**
```go
return pkg.NewDatabaseError("Query failed", "SELECT").WithCause(err)
```

### NewInternalError()

Creates an internal server error (500 Internal Server Error).

**Signature:**
```go
func NewInternalError(message string) *FrameworkError
```

**Example:**
```go
return pkg.NewInternalError("Unexpected error occurred")
```

### NewTenantError()

Creates a tenant-related error.

**Signature:**
```go
func NewTenantError(code, message string, statusCode int) *FrameworkError
```

**Example:**
```go
return pkg.NewTenantError(pkg.ErrCodeTenantNotFound, "Tenant not found", 404)
```

### NewRequestTooLargeError()

Creates a request too large error (413 Request Entity Too Large).

**Signature:**
```go
func NewRequestTooLargeError(maxSize int64) *FrameworkError
```

**Example:**
```go
return pkg.NewRequestTooLargeError(10 * 1024 * 1024) // 10MB
```

### NewRequestTimeoutError()

Creates a request timeout error (408 Request Timeout).

**Signature:**
```go
func NewRequestTimeoutError(timeout time.Duration) *FrameworkError
```

**Example:**
```go
return pkg.NewRequestTimeoutError(30 * time.Second)
```

### NewBogusDataError()

Creates a bogus data error (400 Bad Request).

**Signature:**
```go
func NewBogusDataError(reason string) *FrameworkError
```

**Example:**
```go
return pkg.NewBogusDataError("Malformed JSON")
```

### NewMissingFieldError()

Creates a missing field error (400 Bad Request).

**Signature:**
```go
func NewMissingFieldError(field string) *FrameworkError
```

**Example:**
```go
return pkg.NewMissingFieldError("email")
```

### NewInvalidFormatError()

Creates an invalid format error (400 Bad Request).

**Signature:**
```go
func NewInvalidFormatError(field, expectedFormat string) *FrameworkError
```

**Example:**
```go
return pkg.NewInvalidFormatError("date", "YYYY-MM-DD")
```

### NewSessionError()

Creates a session-related error (401 Unauthorized).

**Signature:**
```go
func NewSessionError(code, message string) *FrameworkError
```

**Example:**
```go
return pkg.NewSessionError(pkg.ErrCodeSessionExpired, "Session has expired")
```

### NewWebSocketError()

Creates a WebSocket-related error (400 Bad Request).

**Signature:**
```go
func NewWebSocketError(code, message string) *FrameworkError
```

**Example:**
```go
return pkg.NewWebSocketError(pkg.ErrCodeWebSocketUpgradeFailed, "Failed to upgrade connection")
```

### NewNotFoundError()

Creates a not found error (404 Not Found).

**Signature:**
```go
func NewNotFoundError(resource string) *FrameworkError
```

**Example:**
```go
return pkg.NewNotFoundError("User")
```

### NewMethodNotAllowedError()

Creates a method not allowed error (405 Method Not Allowed).

**Signature:**
```go
func NewMethodNotAllowedError(method string, allowedMethods []string) *FrameworkError
```

**Example:**
```go
return pkg.NewMethodNotAllowedError("DELETE", []string{"GET", "POST"})
```

### NewConfigurationError()

Creates a configuration error (500 Internal Server Error).

**Signature:**
```go
func NewConfigurationError(key, reason string) *FrameworkError
```

**Example:**
```go
return pkg.NewConfigurationError("database.host", "Invalid hostname")
```

## Error Handler Interface

### ErrorHandler

Interface for handling framework errors.

```go
type ErrorHandler interface {
    HandleError(ctx Context, err error) error
    HandlePanic(ctx Context, recovered interface{}) error
    FormatError(ctx Context, err *FrameworkError) interface{}
}
```

### NewErrorHandler()

Creates a new error handler.

**Signature:**
```go
func NewErrorHandler(config ErrorHandlerConfig) ErrorHandler
```

**Parameters:**
```go
type ErrorHandlerConfig struct {
    I18n         I18nManager
    Logger       Logger
    IncludeStack bool
    LogErrors    bool
    Recovery     RecoveryHandler
}
```

**Example:**
```go
handler := pkg.NewErrorHandler(pkg.ErrorHandlerConfig{
    I18n:         app.I18n(),
    Logger:       app.Logger(),
    IncludeStack: false,
    LogErrors:    true,
})
```

### ErrorHandler.HandleError()

Handles framework errors with internationalization and logging.

**Signature:**
```go
HandleError(ctx Context, err error) error
```

**Example:**
```go
if err != nil {
    return errorHandler.HandleError(ctx, err)
}
```

### ErrorHandler.HandlePanic()

Handles panic recovery.

**Signature:**
```go
HandlePanic(ctx Context, recovered interface{}) error
```

### ErrorHandler.FormatError()

Formats an error for response.

**Signature:**
```go
FormatError(ctx Context, err *FrameworkError) interface{}
```

## Recovery Handler Interface

### RecoveryHandler

Interface for panic recovery.

```go
type RecoveryHandler interface {
    Recover(ctx Context, recovered interface{}) error
    ShouldRecover(recovered interface{}) bool
}
```

### NewRecoveryHandler()

Creates a new recovery handler.

**Signature:**
```go
func NewRecoveryHandler(logger Logger) RecoveryHandler
```

**Example:**
```go
recovery := pkg.NewRecoveryHandler(app.Logger())
```

## Middleware Functions

### RecoveryMiddleware()

Creates a middleware that recovers from panics.

**Signature:**
```go
func RecoveryMiddleware(handler ErrorHandler) MiddlewareFunc
```

**Example:**
```go
errorHandler := pkg.NewErrorHandler(config)
app.Use(pkg.RecoveryMiddleware(errorHandler))
```

### ErrorMiddleware()

Creates a middleware that handles errors.

**Signature:**
```go
func ErrorMiddleware(handler ErrorHandler) MiddlewareFunc
```

**Example:**
```go
errorHandler := pkg.NewErrorHandler(config)
app.Use(pkg.ErrorMiddleware(errorHandler))
```

## Utility Functions

### ValidationError()

Creates a validation error with field information (alias for NewValidationError).

**Signature:**
```go
func ValidationError(field, message string) *FrameworkError
```

### WrapError()

Wraps a generic error into a FrameworkError.

**Signature:**
```go
func WrapError(err error, code string, statusCode int) *FrameworkError
```

**Example:**
```go
if err != nil {
    return pkg.WrapError(err, "CUSTOM_ERROR", 500)
}
```

### IsFrameworkError()

Checks if an error is a FrameworkError.

**Signature:**
```go
func IsFrameworkError(err error) bool
```

**Example:**
```go
if pkg.IsFrameworkError(err) {
    // Handle framework error
}
```

### GetFrameworkError()

Extracts a FrameworkError from an error.

**Signature:**
```go
func GetFrameworkError(err error) (*FrameworkError, bool)
```

**Example:**
```go
if fwErr, ok := pkg.GetFrameworkError(err); ok {
    log.Printf("Error code: %s", fwErr.Code)
}
```

### ErrorResponse()

Creates a standardized error response.

**Signature:**
```go
func ErrorResponse(code, message string, details map[string]interface{}) map[string]interface{}
```

**Example:**
```go
response := pkg.ErrorResponse("CUSTOM_ERROR", "Something went wrong", map[string]interface{}{
    "field": "email",
})
return ctx.JSON(400, response)
```

## Predefined Errors

### ErrNoDatabaseConfigured

Error returned when database operations are attempted without a configured database.

```go
var ErrNoDatabaseConfigured = &FrameworkError{
    Code:       ErrCodeNoDatabaseConfigured,
    Message:    "No database is configured...",
    StatusCode: http.StatusServiceUnavailable,
    I18nKey:    "error.database.not_configured",
}
```

## Usage Examples

### Basic Error Handling

```go
router.POST("/users", func(ctx pkg.Context) error {
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return pkg.NewValidationError("Invalid JSON", "body")
    }
    
    if user.Email == "" {
        return pkg.NewMissingFieldError("email")
    }
    
    // Create user
    if err := db.Create(&user); err != nil {
        return pkg.NewDatabaseError("Failed to create user", "INSERT").
            WithCause(err)
    }
    
    return ctx.JSON(201, user)
})
```

### Error with Details

```go
router.POST("/upload", func(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return pkg.NewValidationError("No file uploaded", "file")
    }
    
    maxSize := int64(10 * 1024 * 1024) // 10MB
    if file.Size > maxSize {
        return pkg.NewRequestTooLargeError(maxSize).
            WithDetails(map[string]interface{}{
                "file_size": file.Size,
                "max_size":  maxSize,
            })
    }
    
    return ctx.JSON(200, map[string]string{"status": "uploaded"})
})
```

### Error with Context

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    if err == sql.ErrNoRows {
        return pkg.NewNotFoundError("User").
            WithContext(
                ctx.Request().ID,
                ctx.Request().TenantID,
                ctx.Request().UserID,
                ctx.Request().Path,
                ctx.Request().Method,
            )
    }
    
    return ctx.JSON(200, user)
})
```

### Custom Error Handler

```go
config := pkg.FrameworkConfig{
    // ... other config
}

app, _ := pkg.New(config)

// Create custom error handler
errorHandler := pkg.NewErrorHandler(pkg.ErrorHandlerConfig{
    I18n:         app.I18n(),
    Logger:       app.Logger(),
    IncludeStack: app.Config().IsDevelopment(),
    LogErrors:    true,
})

// Use error middleware
app.Use(pkg.RecoveryMiddleware(errorHandler))
app.Use(pkg.ErrorMiddleware(errorHandler))
```

### Panic Recovery

```go
router.GET("/panic", func(ctx pkg.Context) error {
    // This panic will be recovered by RecoveryMiddleware
    panic("something went wrong")
})
```

## Error Response Format

### Standard Format

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Email format is invalid",
    "details": {
      "field": "email"
    }
  }
}
```

### Development Mode (includes additional context)

```json
{
  "error": {
    "code": "DATABASE_QUERY",
    "message": "Failed to fetch user",
    "details": {
      "operation": "SELECT"
    },
    "request_id": "req-123",
    "path": "/users/123",
    "method": "GET",
    "cause": "sql: no rows in result set"
  }
}
```

## Best Practices

1. **Use Specific Error Constructors:** Use the appropriate constructor for each error type
2. **Add Context:** Use `WithContext()` to add request information
3. **Chain Errors:** Use `WithCause()` to preserve the error chain
4. **Add Details:** Use `WithDetails()` to provide additional information
5. **Log Errors:** Enable error logging in production
6. **Internationalize:** Use i18n keys for user-facing messages
7. **Sanitize Errors:** Don't expose sensitive information in error messages
8. **Use Middleware:** Use RecoveryMiddleware and ErrorMiddleware for consistent handling
9. **Check Error Types:** Use `IsFrameworkError()` and `GetFrameworkError()` for type checking
10. **Return Early:** Return errors immediately instead of continuing execution

## See Also

- [Context API](context.md) - Context interface reference
- [Security Guide](../guides/security.md) - Security best practices
- [Middleware Guide](../guides/middleware.md) - Middleware patterns
- [I18n API](i18n.md) - Internationalization reference
