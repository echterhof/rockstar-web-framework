# Error Codes Reference

Complete catalog of all error codes in the Rockstar Web Framework with recovery strategies and examples.

## Overview

The framework uses structured errors via the `FrameworkError` type, providing consistent error handling with internationalization support, HTTP status codes, and detailed context.

## Error Structure

```go
type FrameworkError struct {
    Code       string                 // Error code (e.g., "AUTH_FAILED")
    Message    string                 // Human-readable message
    StatusCode int                    // HTTP status code
    Details    map[string]interface{} // Additional details
    Cause      error                  // Underlying error
    I18nKey    string                 // Translation key
    I18nParams map[string]interface{} // Translation parameters
    RequestID  string                 // Request identifier
    TenantID   string                 // Tenant identifier
    UserID     string                 // User identifier
    Path       string                 // Request path
    Method     string                 // HTTP method
}
```

## Error Categories

- [Authentication Errors](#authentication-errors)
- [Authorization Errors](#authorization-errors)
- [Validation Errors](#validation-errors)
- [Request Errors](#request-errors)
- [Security Errors](#security-errors)
- [Database Errors](#database-errors)
- [Session Errors](#session-errors)
- [Configuration Errors](#configuration-errors)
- [Server Errors](#server-errors)
- [Multi-Tenancy Errors](#multi-tenancy-errors)
- [WebSocket Errors](#websocket-errors)
- [Routing Errors](#routing-errors)

---

## Authentication Errors

### AUTH_FAILED

**Code**: `AUTH_FAILED`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.authentication.failed`

**Description**: Authentication attempt failed due to invalid credentials or missing authentication data.

**Common Causes**:
- Invalid username/password
- Missing authentication token
- Expired credentials
- Malformed authentication header

**Recovery Strategy**:
```go
// Client-side: Prompt for re-authentication
if err, ok := pkg.GetFrameworkError(err); ok && err.Code == pkg.ErrCodeAuthenticationFailed {
    // Redirect to login page
    return ctx.Redirect(302, "/login")
}

// API: Return clear error message
return pkg.NewAuthenticationError("Invalid credentials").
    WithDetails(map[string]interface{}{
        "hint": "Check username and password",
    })
```

**Prevention**:
- Validate credentials before authentication
- Use secure password hashing
- Implement rate limiting on auth endpoints
- Log failed authentication attempts

---

### INVALID_TOKEN

**Code**: `INVALID_TOKEN`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.authentication.invalid_token`

**Description**: Provided authentication token is invalid or malformed.

**Common Causes**:
- Token signature verification failed
- Token format is incorrect
- Token has been revoked
- Wrong token type (JWT vs API token)

**Recovery Strategy**:
```go
// Validate token format before use
token := ctx.GetHeader("Authorization")
if !strings.HasPrefix(token, "Bearer ") {
    return pkg.NewAuthenticationError("Invalid token format")
}

// Handle invalid token
if err.Code == pkg.ErrCodeInvalidToken {
    // Clear invalid token and request new one
    ctx.SetCookie(&pkg.Cookie{
        Name:   "auth_token",
        Value:  "",
        MaxAge: -1,
    })
    return ctx.Redirect(302, "/auth/refresh")
}
```

**Prevention**:
- Use standard token formats (JWT)
- Implement token validation middleware
- Set appropriate token expiration
- Use secure token storage

---

### TOKEN_EXPIRED

**Code**: `TOKEN_EXPIRED`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.authentication.token_expired`

**Description**: Authentication token has expired and needs renewal.

**Common Causes**:
- Token TTL exceeded
- System time mismatch
- Token not refreshed in time

**Recovery Strategy**:
```go
// Automatic token refresh
if err.Code == pkg.ErrCodeTokenExpired {
    // Try to refresh token
    newToken, err := ctx.Security().RefreshToken(oldToken)
    if err != nil {
        // Refresh failed, require re-authentication
        return ctx.Redirect(302, "/login")
    }
    
    // Update token and retry request
    ctx.SetHeader("Authorization", "Bearer "+newToken)
    return next(ctx)
}
```

**Prevention**:
- Implement token refresh mechanism
- Set reasonable token expiration times
- Use refresh tokens for long-lived sessions
- Monitor token expiration proactively

---

### UNAUTHORIZED

**Code**: `UNAUTHORIZED`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.authentication.unauthorized`

**Description**: Request requires authentication but none was provided.

**Common Causes**:
- Missing Authorization header
- No session cookie present
- Anonymous access to protected resource

**Recovery Strategy**:
```go
// Authentication middleware
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    if !ctx.IsAuthenticated() {
        return pkg.NewAuthenticationError("Authentication required").
            WithDetails(map[string]interface{}{
                "login_url": "/login",
            })
    }
    return next(ctx)
}
```

**Prevention**:
- Apply authentication middleware to protected routes
- Provide clear authentication requirements in API docs
- Return WWW-Authenticate header with challenge


---

## Authorization Errors

### FORBIDDEN

**Code**: `FORBIDDEN`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.authorization.forbidden`

**Description**: User is authenticated but lacks permission to access the resource.

**Common Causes**:
- Insufficient role privileges
- Missing required permissions
- Resource access denied by policy
- Tenant isolation violation

**Recovery Strategy**:
```go
// Check permissions before action
if !ctx.IsAuthorized("users", "write") {
    return pkg.NewAuthorizationError("Insufficient permissions").
        WithDetails(map[string]interface{}{
            "required_permission": "users:write",
            "user_roles": ctx.User().Roles,
        })
}

// Handle authorization error
if err.Code == pkg.ErrCodeForbidden {
    // Log unauthorized access attempt
    ctx.Logger().Warn("unauthorized access attempt",
        "user_id", ctx.User().ID,
        "resource", ctx.Request().Path,
    )
    
    // Return user-friendly message
    return ctx.JSON(403, map[string]interface{}{
        "error": "You don't have permission to access this resource",
        "contact": "admin@example.com",
    })
}
```

**Prevention**:
- Implement RBAC (Role-Based Access Control)
- Use authorization middleware
- Document required permissions for each endpoint
- Audit permission checks regularly

---

### INSUFFICIENT_ROLES

**Code**: `INSUFFICIENT_ROLES`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.authorization.insufficient_roles`

**Description**: User's roles are insufficient for the requested action.

**Common Causes**:
- User lacks required role (e.g., "admin")
- Role hierarchy not satisfied
- Role assignment not synced

**Recovery Strategy**:
```go
// Role-based middleware
func requireRole(role string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        user := ctx.User()
        if user == nil {
            return pkg.NewAuthenticationError("Authentication required")
        }
        
        hasRole := false
        for _, r := range user.Roles {
            if r == role {
                hasRole = true
                break
            }
        }
        
        if !hasRole {
            return &pkg.FrameworkError{
                Code: pkg.ErrCodeInsufficientRoles,
                Message: fmt.Sprintf("Role '%s' required", role),
                StatusCode: 403,
                Details: map[string]interface{}{
                    "required_role": role,
                    "user_roles": user.Roles,
                },
            }
        }
        
        return next(ctx)
    }
}
```

**Prevention**:
- Assign roles during user creation
- Implement role hierarchy
- Cache role checks for performance
- Provide role management UI

---

### INSUFFICIENT_ACTIONS

**Code**: `INSUFFICIENT_ACTIONS`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.authorization.insufficient_actions`

**Description**: User lacks specific action permission on a resource.

**Common Causes**:
- Fine-grained permission check failed
- Action not in user's permission set
- Resource-specific permission denied

**Recovery Strategy**:
```go
// Action-based authorization
requiredActions := []string{"read", "write"}
for _, action := range requiredActions {
    if !ctx.IsAuthorized("documents", action) {
        return &pkg.FrameworkError{
            Code: pkg.ErrCodeInsufficientActions,
            Message: "Missing required action permission",
            StatusCode: 403,
            Details: map[string]interface{}{
                "resource": "documents",
                "required_action": action,
            },
        }
    }
}
```

**Prevention**:
- Use granular permission model
- Document required actions per endpoint
- Implement permission inheritance
- Provide permission audit logs

---

### INSUFFICIENT_SCOPES

**Code**: `INSUFFICIENT_SCOPES`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.authorization.insufficient_scopes`

**Description**: OAuth2/API token lacks required scopes.

**Common Causes**:
- Token issued with limited scopes
- Scope not requested during authorization
- Scope revoked after token issuance

**Recovery Strategy**:
```go
// Scope validation for API tokens
func requireScopes(scopes ...string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        token := ctx.Get("access_token")
        if token == nil {
            return pkg.NewAuthenticationError("Access token required")
        }
        
        accessToken := token.(*pkg.AccessToken)
        for _, required := range scopes {
            hasScope := false
            for _, scope := range accessToken.Scopes {
                if scope == required {
                    hasScope = true
                    break
                }
            }
            
            if !hasScope {
                return &pkg.FrameworkError{
                    Code: pkg.ErrCodeInsufficientScopes,
                    Message: "Missing required scope",
                    StatusCode: 403,
                    Details: map[string]interface{}{
                        "required_scopes": scopes,
                        "token_scopes": accessToken.Scopes,
                    },
                }
            }
        }
        
        return next(ctx)
    }
}
```

**Prevention**:
- Request appropriate scopes during OAuth2 flow
- Document required scopes in API documentation
- Implement scope validation middleware
- Allow scope upgrade without re-authentication


---

## Validation Errors

### VALIDATION_FAILED

**Code**: `VALIDATION_FAILED`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.validation.failed`

**Description**: Input validation failed for one or more fields.

**Common Causes**:
- Invalid field format
- Value out of range
- Failed custom validation rule
- Type mismatch

**Recovery Strategy**:
```go
// Comprehensive validation with detailed errors
validator := pkg.NewFormValidator(parser)
rules := pkg.FormValidationRules{
    RequiredFields: []string{"email", "password"},
    CustomRules: map[string]func(string) error{
        "email": func(value string) error {
            if !strings.Contains(value, "@") {
                return pkg.NewValidationError("Invalid email format", "email")
            }
            return nil
        },
    },
}

if err := validator.ValidateAll(ctx.Request(), rules); err != nil {
    // Return detailed validation errors
    return ctx.JSON(400, map[string]interface{}{
        "error": "Validation failed",
        "details": err.Error(),
    })
}
```

**Prevention**:
- Validate input on client-side first
- Use schema validation (JSON Schema, etc.)
- Provide clear validation error messages
- Document validation rules in API docs

---

### INVALID_INPUT

**Code**: `INVALID_INPUT`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.validation.invalid_input`

**Description**: Input data is invalid or malformed.

**Common Causes**:
- Invalid JSON/XML format
- Unexpected data type
- Malformed request body
- Invalid encoding

**Recovery Strategy**:
```go
// Safe JSON parsing with error handling
var data map[string]interface{}
if err := json.Unmarshal(ctx.Body(), &data); err != nil {
    return pkg.NewFrameworkError(
        pkg.ErrCodeInvalidInput,
        "Invalid JSON format",
        400,
    ).WithCause(err).WithDetails(map[string]interface{}{
        "hint": "Check JSON syntax",
    })
}
```

**Prevention**:
- Validate content-type header
- Use strict JSON/XML parsing
- Implement request size limits
- Sanitize input data

---

### MISSING_FIELD

**Code**: `MISSING_FIELD`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.validation.missing_field`

**Description**: Required field is missing from request.

**Common Causes**:
- Field not included in request
- Field name typo
- Empty field value
- Null value for required field

**Recovery Strategy**:
```go
// Helper for required field validation
func requireField(ctx pkg.Context, field string) error {
    value := ctx.FormValue(field)
    if value == "" {
        return pkg.NewMissingFieldError(field)
    }
    return nil
}

// Usage
if err := requireField(ctx, "username"); err != nil {
    return err
}
```

**Prevention**:
- Mark required fields in API documentation
- Use schema validation
- Provide default values where appropriate
- Return clear field names in errors

---

### INVALID_FORMAT

**Code**: `INVALID_FORMAT`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.validation.invalid_format`

**Description**: Field value doesn't match expected format.

**Common Causes**:
- Invalid email format
- Invalid date/time format
- Invalid phone number format
- Invalid URL format

**Recovery Strategy**:
```go
// Format validation with regex
import "regexp"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return pkg.NewInvalidFormatError("email", "user@example.com")
    }
    return nil
}
```

**Prevention**:
- Use standard format validators
- Document expected formats
- Provide format examples
- Use input masks on client-side

---

### FILE_TOO_LARGE

**Code**: `FILE_TOO_LARGE`  
**HTTP Status**: 413 Payload Too Large  
**I18n Key**: `error.validation.file_too_large`

**Description**: Uploaded file exceeds size limit.

**Common Causes**:
- File size exceeds configured limit
- Multiple files exceed total limit
- Memory limit reached during upload

**Recovery Strategy**:
```go
// File size validation
const maxFileSize = 10 * 1024 * 1024 // 10MB

file, err := ctx.FormFile("upload")
if err != nil {
    return err
}

if file.Size > maxFileSize {
    return &pkg.FrameworkError{
        Code: pkg.ErrCodeFileTooLarge,
        Message: "File size exceeds limit",
        StatusCode: 413,
        Details: map[string]interface{}{
            "file_size": file.Size,
            "max_size": maxFileSize,
            "max_size_mb": maxFileSize / (1024 * 1024),
        },
    }
}
```

**Prevention**:
- Set appropriate file size limits
- Validate file size on client-side
- Use chunked uploads for large files
- Provide progress feedback

---

### INVALID_FILE_TYPE

**Code**: `INVALID_FILE_TYPE`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.validation.invalid_file_type`

**Description**: Uploaded file type is not allowed.

**Common Causes**:
- File extension not in whitelist
- MIME type mismatch
- File content doesn't match extension
- Malicious file upload attempt

**Recovery Strategy**:
```go
// File type validation
allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}

file, err := ctx.FormFile("avatar")
if err != nil {
    return err
}

contentType := file.Header["Content-Type"][0]
allowed := false
for _, t := range allowedTypes {
    if contentType == t {
        allowed = true
        break
    }
}

if !allowed {
    return &pkg.FrameworkError{
        Code: pkg.ErrCodeInvalidFileType,
        Message: "File type not allowed",
        StatusCode: 400,
        Details: map[string]interface{}{
            "file_type": contentType,
            "allowed_types": allowedTypes,
        },
    }
}
```

**Prevention**:
- Whitelist allowed file types
- Validate MIME type and extension
- Scan files for malware
- Use content-based type detection

---

## Request Errors

### REQUEST_TOO_LARGE

**Code**: `REQUEST_TOO_LARGE`  
**HTTP Status**: 413 Payload Too Large  
**I18n Key**: `error.request.too_large`

**Description**: Request body exceeds maximum allowed size.

**Common Causes**:
- Large JSON/XML payload
- Multiple file uploads
- Memory limit exceeded
- Malicious large request

**Recovery Strategy**:
```go
// Request size middleware
func requestSizeLimit(maxSize int64) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if ctx.Request().ContentLength > maxSize {
            return pkg.NewRequestTooLargeError(maxSize)
        }
        return next(ctx)
    }
}

// Apply to routes
app.Router().POST("/upload", handler, requestSizeLimit(50*1024*1024))
```

**Prevention**:
- Set appropriate request size limits
- Use streaming for large uploads
- Implement chunked transfer encoding
- Validate size on client-side

---

### REQUEST_TIMEOUT

**Code**: `REQUEST_TIMEOUT`  
**HTTP Status**: 408 Request Timeout  
**I18n Key**: `error.request.timeout`

**Description**: Request processing exceeded timeout limit.

**Common Causes**:
- Slow database query
- External API call timeout
- Long-running computation
- Network latency

**Recovery Strategy**:
```go
// Request timeout with context
func handlerWithTimeout(ctx pkg.Context) error {
    // Set 10 second timeout
    timeoutCtx := ctx.WithTimeout(10 * time.Second)
    
    // Use timeout context for operations
    done := make(chan error, 1)
    go func() {
        done <- processRequest(timeoutCtx)
    }()
    
    select {
    case err := <-done:
        return err
    case <-timeoutCtx.Context().Done():
        return pkg.NewRequestTimeoutError(10 * time.Second)
    }
}
```

**Prevention**:
- Set reasonable timeouts
- Optimize slow operations
- Use async processing for long tasks
- Implement request cancellation

---

### BOGUS_DATA

**Code**: `BOGUS_DATA`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.request.bogus_data`

**Description**: Request contains suspicious or malformed data.

**Common Causes**:
- SQL injection attempt detected
- XSS payload detected
- Path traversal attempt
- Malicious input patterns

**Recovery Strategy**:
```go
// Input sanitization
func sanitizeInput(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check for SQL injection patterns
    body := string(ctx.Body())
    if strings.Contains(strings.ToLower(body), "drop table") ||
       strings.Contains(strings.ToLower(body), "union select") {
        return pkg.NewBogusDataError("SQL injection attempt detected")
    }
    
    // Check for XSS patterns
    if strings.Contains(body, "<script>") {
        return pkg.NewBogusDataError("XSS attempt detected")
    }
    
    return next(ctx)
}
```

**Prevention**:
- Use parameterized queries
- Sanitize all user input
- Implement WAF (Web Application Firewall)
- Log suspicious requests

---

### RATE_LIMIT_EXCEEDED

**Code**: `RATE_LIMIT_EXCEEDED`  
**HTTP Status**: 429 Too Many Requests  
**I18n Key**: `error.rate_limit.exceeded`

**Description**: Client has exceeded rate limit for the endpoint.

**Common Causes**:
- Too many requests in time window
- Aggressive polling
- DDoS attack
- Misconfigured client

**Recovery Strategy**:
```go
// Rate limiting middleware
func rateLimitMiddleware(limit int, window time.Duration) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        key := fmt.Sprintf("ratelimit:%s:%s", 
            ctx.Request().RemoteAddr, 
            ctx.Request().Path)
        
        allowed, err := ctx.DB().CheckRateLimit(key, limit, window)
        if err != nil {
            return err
        }
        
        if !allowed {
            return pkg.NewRateLimitError(
                "Rate limit exceeded",
                limit,
                window.String(),
            ).WithDetails(map[string]interface{}{
                "retry_after": window.Seconds(),
            })
        }
        
        ctx.DB().IncrementRateLimit(key, window)
        return next(ctx)
    }
}
```

**Prevention**:
- Implement rate limiting per user/IP
- Use exponential backoff on client
- Cache responses to reduce requests
- Monitor rate limit metrics

---

## Security Errors

### CSRF_TOKEN_INVALID

**Code**: `CSRF_TOKEN_INVALID`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.security.csrf_invalid`

**Description**: CSRF token validation failed.

**Common Causes**:
- Missing CSRF token
- Token mismatch
- Token expired
- Token from different session

**Recovery Strategy**:
```go
// CSRF protection middleware
func csrfMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    if ctx.Request().Method == "POST" || 
       ctx.Request().Method == "PUT" || 
       ctx.Request().Method == "DELETE" {
        
        token := ctx.FormValue("csrf_token")
        if token == "" {
            token = ctx.GetHeader("X-CSRF-Token")
        }
        
        valid, err := ctx.Security().ValidateCSRFToken(token)
        if err != nil || !valid {
            return &pkg.FrameworkError{
                Code: pkg.ErrCodeCSRFTokenInvalid,
                Message: "CSRF token validation failed",
                StatusCode: 403,
            }
        }
    }
    
    return next(ctx)
}
```

**Prevention**:
- Generate CSRF token per session
- Include token in forms and AJAX requests
- Validate token on state-changing operations
- Use SameSite cookie attribute

---

### XSS_DETECTED

**Code**: `XSS_DETECTED`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.security.xss_detected`

**Description**: Cross-site scripting attempt detected in input.

**Common Causes**:
- Script tags in input
- Event handlers in HTML
- JavaScript URLs
- Encoded XSS payloads

**Recovery Strategy**:
```go
// XSS detection and sanitization
import "html"

func sanitizeHTML(input string) (string, error) {
    dangerous := []string{
        "<script", "</script>",
        "javascript:", "onerror=", "onload=",
    }
    
    lower := strings.ToLower(input)
    for _, pattern := range dangerous {
        if strings.Contains(lower, pattern) {
            return "", &pkg.FrameworkError{
                Code: pkg.ErrCodeXSSDetected,
                Message: "XSS attempt detected",
                StatusCode: 400,
            }
        }
    }
    
    return html.EscapeString(input), nil
}
```

**Prevention**:
- Escape all user input in HTML output
- Use Content Security Policy headers
- Sanitize input on server-side
- Use templating engines with auto-escaping

---

### SQL_INJECTION_DETECTED

**Code**: `SQL_INJECTION_DETECTED`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.security.sql_injection_detected`

**Description**: SQL injection attempt detected in input.

**Common Causes**:
- SQL keywords in input
- Quote escaping attempts
- UNION queries
- Comment sequences

**Recovery Strategy**:
```go
// Always use parameterized queries
func getUserByEmail(ctx pkg.Context, email string) (*pkg.User, error) {
    // GOOD: Parameterized query
    query := ctx.DB().GetQuery("select_user_by_email")
    row := ctx.DB().QueryRow(query, email)
    
    // BAD: String concatenation (vulnerable)
    // query := "SELECT * FROM users WHERE email = '" + email + "'"
    
    var user pkg.User
    err := row.Scan(&user.ID, &user.Email, &user.Username)
    return &user, err
}
```

**Prevention**:
- Always use parameterized queries
- Never concatenate SQL strings
- Use ORM with prepared statements
- Validate and sanitize input

---

### PATH_TRAVERSAL

**Code**: `PATH_TRAVERSAL`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.security.path_traversal`

**Description**: Path traversal attempt detected in file path.

**Common Causes**:
- "../" sequences in path
- Absolute path attempts
- Encoded path traversal
- Symlink exploitation

**Recovery Strategy**:
```go
// Safe file path validation
import "path/filepath"

func validateFilePath(basePath, userPath string) (string, error) {
    // Clean the path
    cleanPath := filepath.Clean(userPath)
    
    // Check for path traversal
    if strings.Contains(cleanPath, "..") {
        return "", &pkg.FrameworkError{
            Code: pkg.ErrCodePathTraversal,
            Message: "Path traversal attempt detected",
            StatusCode: 400,
        }
    }
    
    // Build full path
    fullPath := filepath.Join(basePath, cleanPath)
    
    // Ensure path is within base directory
    if !strings.HasPrefix(fullPath, basePath) {
        return "", &pkg.FrameworkError{
            Code: pkg.ErrCodePathTraversal,
            Message: "Invalid file path",
            StatusCode: 400,
        }
    }
    
    return fullPath, nil
}
```

**Prevention**:
- Validate all file paths
- Use whitelist of allowed paths
- Resolve symlinks before validation
- Restrict file system access

---

### REGEX_TIMEOUT

**Code**: `REGEX_TIMEOUT`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.security.regex_timeout`

**Description**: Regular expression processing exceeded timeout (ReDoS protection).

**Common Causes**:
- Complex regex pattern
- Large input string
- Catastrophic backtracking
- ReDoS attack attempt

**Recovery Strategy**:
```go
// Regex with timeout protection
import "context"
import "time"

func matchWithTimeout(pattern, input string, timeout time.Duration) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    done := make(chan bool, 1)
    go func() {
        re := regexp.MustCompile(pattern)
        done <- re.MatchString(input)
    }()
    
    select {
    case result := <-done:
        return result, nil
    case <-ctx.Done():
        return false, &pkg.FrameworkError{
            Code: pkg.ErrCodeRegexTimeout,
            Message: "Regex processing timeout",
            StatusCode: 400,
        }
    }
}
```

**Prevention**:
- Use simple regex patterns
- Set regex timeout limits
- Validate input length before regex
- Use alternative parsing methods


---

## Database Errors

### DATABASE_CONNECTION

**Code**: `DATABASE_CONNECTION`  
**HTTP Status**: 503 Service Unavailable  
**I18n Key**: `error.database.connection`

**Description**: Failed to establish database connection.

**Common Causes**:
- Database server down
- Network connectivity issues
- Invalid credentials
- Connection pool exhausted

**Recovery Strategy**:
```go
// Connection retry with exponential backoff
func connectWithRetry(config pkg.DatabaseConfig, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = app.DB().Connect(config)
        if err == nil {
            return nil
        }
        
        // Exponential backoff
        wait := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(wait)
    }
    
    return pkg.NewDatabaseError("Failed to connect to database", "connect").
        WithCause(err)
}

// Graceful degradation
func handler(ctx pkg.Context) error {
    if !ctx.DB().IsConnected() {
        // Return cached data or error
        return ctx.JSON(503, map[string]interface{}{
            "error": "Database temporarily unavailable",
            "retry_after": 60,
        })
    }
    // Normal processing
}
```

**Prevention**:
- Monitor database health
- Use connection pooling
- Implement circuit breaker
- Set appropriate timeouts

---

### DATABASE_QUERY

**Code**: `DATABASE_QUERY`  
**HTTP Status**: 500 Internal Server Error  
**I18n Key**: `error.database.query`

**Description**: Database query execution failed.

**Common Causes**:
- SQL syntax error
- Table/column doesn't exist
- Constraint violation
- Query timeout

**Recovery Strategy**:
```go
// Safe query execution with error handling
func executeQuery(ctx pkg.Context, query string, args ...interface{}) error {
    rows, err := ctx.DB().Query(query, args...)
    if err != nil {
        // Log the error with context
        ctx.Logger().Error("query failed",
            "query", query,
            "error", err,
        )
        
        // Return user-friendly error
        return pkg.NewDatabaseError(
            "Failed to retrieve data",
            "query",
        ).WithCause(err)
    }
    defer rows.Close()
    
    // Process rows...
    return nil
}
```

**Prevention**:
- Test queries before deployment
- Use query builder or ORM
- Set query timeouts
- Monitor slow queries

---

### DATABASE_TRANSACTION

**Code**: `DATABASE_TRANSACTION`  
**HTTP Status**: 500 Internal Server Error  
**I18n Key**: `error.database.transaction`

**Description**: Database transaction failed.

**Common Causes**:
- Deadlock detected
- Transaction timeout
- Constraint violation
- Connection lost during transaction

**Recovery Strategy**:
```go
// Safe transaction handling
func performTransaction(ctx pkg.Context) error {
    tx, err := ctx.DB().Begin()
    if err != nil {
        return pkg.NewDatabaseError("Failed to start transaction", "begin")
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()
    
    // Perform operations
    if err := doOperation1(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    if err := doOperation2(tx); err != nil {
        tx.Rollback()
        return err
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        return pkg.NewDatabaseError("Failed to commit transaction", "commit").
            WithCause(err)
    }
    
    return nil
}
```

**Prevention**:
- Keep transactions short
- Handle deadlocks with retry
- Use appropriate isolation levels
- Monitor transaction duration


---

### RECORD_NOT_FOUND

**Code**: `RECORD_NOT_FOUND`  
**HTTP Status**: 404 Not Found  
**I18n Key**: `error.database.not_found`

**Description**: Requested database record doesn't exist.

**Common Causes**:
- Invalid ID provided
- Record deleted
- Wrong database/tenant
- Query returned no rows

**Recovery Strategy**:
```go
// Handle not found gracefully
func getUser(ctx pkg.Context, userID string) (*pkg.User, error) {
    query := ctx.DB().GetQuery("select_user_by_id")
    row := ctx.DB().QueryRow(query, userID)
    
    var user pkg.User
    err := row.Scan(&user.ID, &user.Email, &user.Username)
    if err == sql.ErrNoRows {
        return nil, pkg.NewNotFoundError("User")
    }
    if err != nil {
        return nil, pkg.NewDatabaseError("Failed to get user", "query")
    }
    
    return &user, nil
}
```

**Prevention**:
- Validate IDs before querying
- Use EXISTS queries for checks
- Implement soft deletes
- Cache frequently accessed records

---

### DUPLICATE_RECORD

**Code**: `DUPLICATE_RECORD`  
**HTTP Status**: 409 Conflict  
**I18n Key**: `error.database.duplicate`

**Description**: Attempted to create duplicate record (unique constraint violation).

**Common Causes**:
- Duplicate email/username
- Unique constraint violation
- Race condition in creation
- Retry of failed request

**Recovery Strategy**:
```go
// Handle duplicate gracefully
func createUser(ctx pkg.Context, user *pkg.User) error {
    query := ctx.DB().GetQuery("insert_user")
    _, err := ctx.DB().Exec(query, user.ID, user.Email, user.Username)
    
    if err != nil {
        // Check for duplicate key error
        if strings.Contains(err.Error(), "duplicate") ||
           strings.Contains(err.Error(), "unique constraint") {
            return &pkg.FrameworkError{
                Code: pkg.ErrCodeDuplicateRecord,
                Message: "User already exists",
                StatusCode: 409,
                Details: map[string]interface{}{
                    "field": "email",
                },
            }
        }
        return pkg.NewDatabaseError("Failed to create user", "insert")
    }
    
    return nil
}
```

**Prevention**:
- Check existence before insert
- Use UPSERT operations
- Implement idempotency keys
- Handle race conditions

---

### NO_DATABASE_CONFIGURED

**Code**: `NO_DATABASE_CONFIGURED`  
**HTTP Status**: 503 Service Unavailable  
**I18n Key**: `error.database.not_configured`

**Description**: Database operations attempted without configured database.

**Common Causes**:
- Missing database configuration
- Database disabled in config
- Optional database feature used
- Configuration not loaded

**Recovery Strategy**:
```go
// Check database availability
func handler(ctx pkg.Context) error {
    if !ctx.DB().IsConnected() {
        return pkg.ErrNoDatabaseConfigured
    }
    
    // Proceed with database operations
    return nil
}

// Optional database pattern
func handlerWithOptionalDB(ctx pkg.Context) error {
    if ctx.DB().IsConnected() {
        // Use database
        return fetchFromDB(ctx)
    } else {
        // Fallback to in-memory or external service
        return fetchFromCache(ctx)
    }
}
```

**Prevention**:
- Validate configuration at startup
- Document database requirements
- Provide clear setup instructions
- Support optional database mode

---

## Session Errors

### SESSION_NOT_FOUND

**Code**: `SESSION_NOT_FOUND`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.session.not_found`

**Description**: Session ID not found in storage.

**Common Causes**:
- Session expired and cleaned up
- Invalid session ID
- Session storage cleared
- Wrong session storage backend

**Recovery Strategy**:
```go
// Session validation middleware
func sessionMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    sessionID := ctx.GetCookie("session_id")
    if sessionID == nil {
        return pkg.NewSessionError(pkg.ErrCodeSessionNotFound, "No session found")
    }
    
    session, err := ctx.Session().Load(ctx, sessionID.Value)
    if err != nil {
        // Create new session
        newSession, err := ctx.Session().Create(ctx)
        if err != nil {
            return err
        }
        
        ctx.SetCookie(&pkg.Cookie{
            Name:     "session_id",
            Value:    newSession.ID,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
        })
    }
    
    return next(ctx)
}
```

**Prevention**:
- Set appropriate session lifetime
- Implement session renewal
- Handle expired sessions gracefully
- Use persistent session storage

---

### SESSION_EXPIRED

**Code**: `SESSION_EXPIRED`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.session.expired`

**Description**: Session has expired and is no longer valid.

**Common Causes**:
- Session TTL exceeded
- Inactivity timeout
- Explicit session termination
- Server restart with memory storage

**Recovery Strategy**:
```go
// Session expiration handling
func checkSession(ctx pkg.Context) error {
    session, err := ctx.Session().Load(ctx, sessionID)
    if err != nil {
        return err
    }
    
    if time.Now().After(session.ExpiresAt) {
        // Clean up expired session
        ctx.Session().Delete(ctx, session.ID)
        
        return pkg.NewSessionError(
            pkg.ErrCodeSessionExpired,
            "Session has expired",
        )
    }
    
    // Extend session on activity
    session.ExpiresAt = time.Now().Add(24 * time.Hour)
    ctx.Session().Save(ctx, session)
    
    return nil
}
```

**Prevention**:
- Implement sliding expiration
- Warn users before expiration
- Auto-save user work
- Use refresh tokens

---

### SESSION_INVALID

**Code**: `SESSION_INVALID`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.session.invalid`

**Description**: Session data is corrupted or invalid.

**Common Causes**:
- Tampered session cookie
- Decryption failure
- Data corruption
- Version mismatch

**Recovery Strategy**:
```go
// Session validation
func validateSession(ctx pkg.Context, session *pkg.Session) error {
    // Validate session structure
    if session.ID == "" || session.CreatedAt.IsZero() {
        return pkg.NewSessionError(
            pkg.ErrCodeSessionInvalid,
            "Invalid session data",
        )
    }
    
    // Validate session signature (if using signed cookies)
    if !ctx.Security().ValidateSessionSignature(session) {
        return pkg.NewSessionError(
            pkg.ErrCodeSessionInvalid,
            "Session signature invalid",
        )
    }
    
    return nil
}
```

**Prevention**:
- Use encrypted session cookies
- Sign session data
- Validate session structure
- Version session format

---

## Configuration Errors

### CONFIGURATION_ERROR

**Code**: `CONFIGURATION_ERROR`  
**HTTP Status**: 500 Internal Server Error  
**I18n Key**: `error.configuration.error`

**Description**: Configuration is invalid or malformed.

**Common Causes**:
- Invalid YAML/JSON syntax
- Missing required fields
- Invalid value types
- Conflicting settings

**Recovery Strategy**:
```go
// Configuration validation
func validateConfig(config *pkg.FrameworkConfig) error {
    if config.ServerConfig.Address == "" {
        return pkg.NewConfigurationError(
            "ServerConfig.Address",
            "address is required",
        )
    }
    
    if config.DatabaseConfig.Driver != "" {
        validDrivers := []string{"mysql", "postgres", "sqlite", "mssql"}
        valid := false
        for _, d := range validDrivers {
            if config.DatabaseConfig.Driver == d {
                valid = true
                break
            }
        }
        if !valid {
            return pkg.NewConfigurationError(
                "DatabaseConfig.Driver",
                fmt.Sprintf("must be one of: %v", validDrivers),
            )
        }
    }
    
    return nil
}
```

**Prevention**:
- Validate config at startup
- Use schema validation
- Provide config examples
- Document all config options

---

### MISSING_CONFIGURATION

**Code**: `MISSING_CONFIGURATION`  
**HTTP Status**: 500 Internal Server Error  
**I18n Key**: `error.configuration.missing`

**Description**: Required configuration is missing.

**Common Causes**:
- Config file not found
- Environment variable not set
- Required field omitted
- Config not loaded

**Recovery Strategy**:
```go
// Configuration with defaults
func loadConfig() (*pkg.FrameworkConfig, error) {
    config := &pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            Address: ":8080", // Default
        },
    }
    
    // Try to load from file
    if err := config.LoadFromFile("config.yaml"); err != nil {
        // Try environment variables
        if err := config.LoadFromEnv(); err != nil {
            return nil, &pkg.FrameworkError{
                Code: pkg.ErrCodeMissingConfiguration,
                Message: "No configuration found",
                StatusCode: 500,
                Details: map[string]interface{}{
                    "hint": "Create config.yaml or set environment variables",
                },
            }
        }
    }
    
    return config, nil
}
```

**Prevention**:
- Provide default configuration
- Document required settings
- Validate config at startup
- Support multiple config sources

---

## Server Errors

### INTERNAL_ERROR

**Code**: `INTERNAL_ERROR`  
**HTTP Status**: 500 Internal Server Error  
**I18n Key**: `error.internal.server`

**Description**: Unexpected internal server error occurred.

**Common Causes**:
- Unhandled exception
- Null pointer dereference
- Resource exhaustion
- Bug in application code

**Recovery Strategy**:
```go
// Global error handler with recovery
func errorHandler(ctx pkg.Context, next pkg.HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            ctx.Logger().Error("panic recovered",
                "panic", r,
                "stack", string(debug.Stack()),
            )
            
            ctx.JSON(500, map[string]interface{}{
                "error": "Internal server error",
                "request_id": ctx.Request().ID,
            })
        }
    }()
    
    err := next(ctx)
    if err != nil {
        // Log error with context
        ctx.Logger().Error("request failed",
            "error", err,
            "path", ctx.Request().Path,
            "method", ctx.Request().Method,
        )
        
        // Return generic error to client
        if !pkg.IsFrameworkError(err) {
            return pkg.NewInternalError("An unexpected error occurred")
        }
    }
    
    return err
}
```

**Prevention**:
- Implement comprehensive error handling
- Use panic recovery middleware
- Monitor error rates
- Log errors with context

---

### SERVICE_UNAVAILABLE

**Code**: `SERVICE_UNAVAILABLE`  
**HTTP Status**: 503 Service Unavailable  
**I18n Key**: `error.service.unavailable`

**Description**: Service is temporarily unavailable.

**Common Causes**:
- Server overloaded
- Maintenance mode
- Dependency failure
- Resource exhaustion

**Recovery Strategy**:
```go
// Health check and circuit breaker
type healthChecker struct {
    healthy bool
    mu      sync.RWMutex
}

func (h *healthChecker) isHealthy() bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.healthy
}

func healthCheckMiddleware(checker *healthChecker) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if !checker.isHealthy() {
            return &pkg.FrameworkError{
                Code: pkg.ErrCodeServiceUnavailable,
                Message: "Service temporarily unavailable",
                StatusCode: 503,
                Details: map[string]interface{}{
                    "retry_after": 60,
                },
            }
        }
        return next(ctx)
    }
}
```

**Prevention**:
- Implement health checks
- Use circuit breakers
- Monitor resource usage
- Implement graceful degradation

---

### NOT_IMPLEMENTED

**Code**: `NOT_IMPLEMENTED`  
**HTTP Status**: 501 Not Implemented  
**I18n Key**: `error.server.not_implemented`

**Description**: Requested feature is not implemented.

**Common Causes**:
- Feature under development
- Deprecated endpoint
- Unsupported operation
- Platform limitation

**Recovery Strategy**:
```go
// Feature flag pattern
func featureHandler(ctx pkg.Context) error {
    if !ctx.Config().IsFeatureEnabled("new_feature") {
        return &pkg.FrameworkError{
            Code: pkg.ErrCodeNotImplemented,
            Message: "Feature not yet available",
            StatusCode: 501,
            Details: map[string]interface{}{
                "feature": "new_feature",
                "status": "coming_soon",
            },
        }
    }
    
    // Feature implementation
    return nil
}
```

**Prevention**:
- Use feature flags
- Document API roadmap
- Version APIs properly
- Communicate deprecations

---

## Multi-Tenancy Errors

### TENANT_NOT_FOUND

**Code**: `TENANT_NOT_FOUND`  
**HTTP Status**: 404 Not Found  
**I18n Key**: `error.tenant.not_found`

**Description**: Requested tenant doesn't exist.

**Common Causes**:
- Invalid tenant ID
- Tenant deleted
- Wrong hostname
- Tenant not configured

**Recovery Strategy**:
```go
// Tenant resolution middleware
func tenantMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    hostname := ctx.Request().Host
    
    tenant, err := ctx.DB().LoadTenantByHost(hostname)
    if err != nil {
        return pkg.NewTenantError(
            pkg.ErrCodeTenantNotFound,
            "Tenant not found for this hostname",
            404,
        ).WithDetails(map[string]interface{}{
            "hostname": hostname,
        })
    }
    
    // Store tenant in context
    ctx.Set("tenant", tenant)
    
    return next(ctx)
}
```

**Prevention**:
- Validate tenant on each request
- Cache tenant lookups
- Provide default tenant
- Document tenant setup

---

### TENANT_INACTIVE

**Code**: `TENANT_INACTIVE`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.tenant.inactive`

**Description**: Tenant account is inactive or suspended.

**Common Causes**:
- Subscription expired
- Account suspended
- Maintenance mode
- Terms violation

**Recovery Strategy**:
```go
// Tenant status check
func checkTenantStatus(ctx pkg.Context, next pkg.HandlerFunc) error {
    tenant := ctx.Tenant()
    if tenant == nil {
        return pkg.NewTenantError(
            pkg.ErrCodeTenantNotFound,
            "No tenant context",
            404,
        )
    }
    
    if !tenant.IsActive {
        return pkg.NewTenantError(
            pkg.ErrCodeTenantInactive,
            "Tenant account is inactive",
            403,
        ).WithDetails(map[string]interface{}{
            "tenant_id": tenant.ID,
            "contact": "support@example.com",
        })
    }
    
    return next(ctx)
}
```

**Prevention**:
- Monitor tenant status
- Send expiration warnings
- Provide reactivation flow
- Grace period for expired accounts

---

### TENANT_LIMIT_EXCEEDED

**Code**: `TENANT_LIMIT_EXCEEDED`  
**HTTP Status**: 429 Too Many Requests  
**I18n Key**: `error.tenant.limit_exceeded`

**Description**: Tenant has exceeded resource limits.

**Common Causes**:
- User limit reached
- Storage quota exceeded
- Request quota exceeded
- Feature limit reached

**Recovery Strategy**:
```go
// Tenant quota enforcement
func checkTenantQuota(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    
    // Check user limit
    userCount, err := getUserCount(ctx, tenant.ID)
    if err != nil {
        return err
    }
    
    if userCount >= tenant.MaxUsers {
        return pkg.NewTenantError(
            pkg.ErrCodeTenantLimitExceeded,
            "User limit exceeded",
            429,
        ).WithDetails(map[string]interface{}{
            "limit": tenant.MaxUsers,
            "current": userCount,
            "upgrade_url": "/billing/upgrade",
        })
    }
    
    return nil
}
```

**Prevention**:
- Monitor quota usage
- Send usage alerts
- Provide upgrade path
- Implement soft limits

---

## WebSocket Errors

### WEBSOCKET_UPGRADE_FAILED

**Code**: `WEBSOCKET_UPGRADE_FAILED`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.websocket.upgrade_failed`

**Description**: Failed to upgrade HTTP connection to WebSocket.

**Common Causes**:
- Missing Upgrade header
- Invalid WebSocket key
- Protocol mismatch
- Proxy interference

**Recovery Strategy**:
```go
// WebSocket upgrade with validation
func wsHandler(ctx pkg.Context) error {
    // Validate upgrade headers
    if ctx.GetHeader("Upgrade") != "websocket" {
        return pkg.NewWebSocketError(
            pkg.ErrCodeWebSocketUpgradeFailed,
            "Invalid upgrade header",
        )
    }
    
    // Upgrade connection
    conn, err := ctx.Router().UpgradeWebSocket(ctx, nil)
    if err != nil {
        return pkg.NewWebSocketError(
            pkg.ErrCodeWebSocketUpgradeFailed,
            "Failed to upgrade connection",
        ).WithCause(err)
    }
    
    // Handle WebSocket connection
    return handleWebSocket(conn, ctx)
}
```

**Prevention**:
- Validate upgrade headers
- Check WebSocket support
- Configure proxy correctly
- Use standard WebSocket protocol

---

### WEBSOCKET_CONNECTION_CLOSED

**Code**: `WEBSOCKET_CONNECTION_CLOSED`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.websocket.connection_closed`

**Description**: WebSocket connection was closed unexpectedly.

**Common Causes**:
- Client disconnected
- Network timeout
- Server shutdown
- Protocol error

**Recovery Strategy**:
```go
// WebSocket with reconnection
func handleWebSocket(conn pkg.WebSocketConnection, ctx pkg.Context) error {
    defer conn.Close()
    
    for {
        msgType, data, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                // Normal closure
                return nil
            }
            
            return pkg.NewWebSocketError(
                pkg.ErrCodeWebSocketConnectionClosed,
                "Connection closed unexpectedly",
            ).WithCause(err)
        }
        
        // Process message
        if err := processMessage(msgType, data); err != nil {
            conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
        }
    }
}
```

**Prevention**:
- Implement ping/pong heartbeat
- Handle reconnection on client
- Set appropriate timeouts
- Log disconnection reasons

---

### WEBSOCKET_AUTH_REQUIRED

**Code**: `WEBSOCKET_AUTH_REQUIRED`  
**HTTP Status**: 401 Unauthorized  
**I18n Key**: `error.websocket.auth_required`

**Description**: WebSocket connection requires authentication.

**Common Causes**:
- Missing auth token
- Token not sent in handshake
- Session expired
- Invalid credentials

**Recovery Strategy**:
```go
// WebSocket authentication
func authenticatedWSHandler(ctx pkg.Context) error {
    // Check authentication before upgrade
    if !ctx.IsAuthenticated() {
        return pkg.NewWebSocketError(
            pkg.ErrCodeWebSocketAuthRequired,
            "Authentication required for WebSocket",
        )
    }
    
    // Upgrade with authenticated context
    conn, err := ctx.Router().UpgradeWebSocket(ctx, nil)
    if err != nil {
        return err
    }
    
    // Store user context
    userID := ctx.User().ID
    
    return handleAuthenticatedWS(conn, userID)
}
```

**Prevention**:
- Authenticate before upgrade
- Pass token in query string or header
- Validate session on connect
- Implement token refresh

---

### WEBSOCKET_INVALID_MESSAGE

**Code**: `WEBSOCKET_INVALID_MESSAGE`  
**HTTP Status**: 400 Bad Request  
**I18n Key**: `error.websocket.invalid_message`

**Description**: Received invalid WebSocket message.

**Common Causes**:
- Malformed JSON
- Invalid message type
- Protocol violation
- Oversized message

**Recovery Strategy**:
```go
// WebSocket message validation
func processWSMessage(conn pkg.WebSocketConnection, data []byte) error {
    var msg struct {
        Type    string          `json:"type"`
        Payload json.RawMessage `json:"payload"`
    }
    
    if err := json.Unmarshal(data, &msg); err != nil {
        errMsg := pkg.NewWebSocketError(
            pkg.ErrCodeWebSocketInvalidMessage,
            "Invalid message format",
        )
        
        // Send error back to client
        conn.WriteMessage(websocket.TextMessage, []byte(errMsg.Error()))
        return errMsg
    }
    
    // Validate message type
    validTypes := []string{"chat", "ping", "subscribe"}
    valid := false
    for _, t := range validTypes {
        if msg.Type == t {
            valid = true
            break
        }
    }
    
    if !valid {
        return pkg.NewWebSocketError(
            pkg.ErrCodeWebSocketInvalidMessage,
            "Unknown message type",
        )
    }
    
    return nil
}
```

**Prevention**:
- Define message schema
- Validate message structure
- Set message size limits
- Document message protocol

---

## Routing Errors

### NOT_FOUND

**Code**: `NOT_FOUND`  
**HTTP Status**: 404 Not Found  
**I18n Key**: `error.not_found`

**Description**: Requested resource or endpoint not found.

**Common Causes**:
- Invalid URL path
- Route not registered
- Resource deleted
- Typo in URL

**Recovery Strategy**:
```go
// Custom 404 handler
func notFoundHandler(ctx pkg.Context) error {
    return ctx.JSON(404, map[string]interface{}{
        "error": "Resource not found",
        "path": ctx.Request().Path,
        "suggestions": getSimilarPaths(ctx.Request().Path),
    })
}

// Register 404 handler
app.Router().NotFound(notFoundHandler)
```

**Prevention**:
- Document all endpoints
- Implement route suggestions
- Use consistent URL patterns
- Log 404 errors for analysis

---

### METHOD_NOT_ALLOWED

**Code**: `METHOD_NOT_ALLOWED`  
**HTTP Status**: 405 Method Not Allowed  
**I18n Key**: `error.method_not_allowed`

**Description**: HTTP method not allowed for this endpoint.

**Common Causes**:
- Wrong HTTP method used
- Method not implemented
- Route registered for different method
- API version mismatch

**Recovery Strategy**:
```go
// Method not allowed handler
func methodNotAllowedHandler(ctx pkg.Context) error {
    allowedMethods := []string{"GET", "POST"}
    
    return pkg.NewMethodNotAllowedError(
        ctx.Request().Method,
        allowedMethods,
    )
}

// Register handler
app.Router().MethodNotAllowed(methodNotAllowedHandler)
```

**Prevention**:
- Document allowed methods
- Return Allow header
- Use OPTIONS for discovery
- Implement CORS properly

---

### AUTHORIZATION_FAILED

**Code**: `AUTHORIZATION_FAILED`  
**HTTP Status**: 403 Forbidden  
**I18n Key**: `error.authorization.failed`

**Description**: Authorization check failed for route.

**Common Causes**:
- Insufficient permissions
- Role check failed
- Resource ownership violation
- Policy evaluation failed

**Recovery Strategy**:
```go
// Authorization middleware
func authorize(resource, action string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if !ctx.IsAuthorized(resource, action) {
            return &pkg.FrameworkError{
                Code: pkg.ErrCodeAuthorizationFailed,
                Message: "Authorization failed",
                StatusCode: 403,
                Details: map[string]interface{}{
                    "resource": resource,
                    "action": action,
                },
            }
        }
        return next(ctx)
    }
}
```

**Prevention**:
- Implement consistent authorization
- Use middleware for protection
- Document permission requirements
- Audit authorization failures

---

## Error Handling Best Practices

### 1. Consistent Error Responses

Always return errors in consistent format:

```go
{
    "error": {
        "code": "VALIDATION_FAILED",
        "message": "Email is required",
        "details": {
            "field": "email"
        }
    }
}
```

### 2. Error Logging

Log errors with context:

```go
ctx.Logger().Error("operation failed",
    "error", err,
    "user_id", ctx.User().ID,
    "tenant_id", ctx.Tenant().ID,
    "request_id", ctx.Request().ID,
)
```

### 3. Error Recovery

Implement graceful degradation:

```go
func handler(ctx pkg.Context) error {
    data, err := fetchFromDB(ctx)
    if err != nil {
        // Fallback to cache
        data, err = fetchFromCache(ctx)
        if err != nil {
            // Return error
            return err
        }
    }
    return ctx.JSON(200, data)
}
```

### 4. Error Monitoring

Track error rates and patterns:

```go
func errorMetrics(ctx pkg.Context, next pkg.HandlerFunc) error {
    err := next(ctx)
    if err != nil {
        if fwErr, ok := pkg.GetFrameworkError(err); ok {
            ctx.Metrics().IncrementCounter(
                "errors",
                "code", fwErr.Code,
                "status", fwErr.StatusCode,
            )
        }
    }
    return err
}
```

### 5. User-Friendly Messages

Provide helpful error messages:

```go
return pkg.NewValidationError("Email format is invalid", "email").
    WithDetails(map[string]interface{}{
        "example": "user@example.com",
        "pattern": "^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$",
    })
```

---

## Error Code Quick Reference

| Code | Status | Category | Severity |
|------|--------|----------|----------|
| AUTH_FAILED | 401 | Authentication | High |
| INVALID_TOKEN | 401 | Authentication | High |
| TOKEN_EXPIRED | 401 | Authentication | Medium |
| UNAUTHORIZED | 401 | Authentication | High |
| FORBIDDEN | 403 | Authorization | High |
| INSUFFICIENT_ROLES | 403 | Authorization | High |
| INSUFFICIENT_ACTIONS | 403 | Authorization | High |
| INSUFFICIENT_SCOPES | 403 | Authorization | High |
| VALIDATION_FAILED | 400 | Validation | Low |
| INVALID_INPUT | 400 | Validation | Low |
| MISSING_FIELD | 400 | Validation | Low |
| INVALID_FORMAT | 400 | Validation | Low |
| FILE_TOO_LARGE | 413 | Validation | Medium |
| INVALID_FILE_TYPE | 400 | Validation | Medium |
| REQUEST_TOO_LARGE | 413 | Request | Medium |
| REQUEST_TIMEOUT | 408 | Request | Medium |
| BOGUS_DATA | 400 | Security | High |
| RATE_LIMIT_EXCEEDED | 429 | Request | Medium |
| CSRF_TOKEN_INVALID | 403 | Security | High |
| XSS_DETECTED | 400 | Security | High |
| SQL_INJECTION_DETECTED | 400 | Security | Critical |
| PATH_TRAVERSAL | 400 | Security | Critical |
| REGEX_TIMEOUT | 400 | Security | Medium |
| DATABASE_CONNECTION | 503 | Database | Critical |
| DATABASE_QUERY | 500 | Database | High |
| DATABASE_TRANSACTION | 500 | Database | High |
| RECORD_NOT_FOUND | 404 | Database | Low |
| DUPLICATE_RECORD | 409 | Database | Low |
| NO_DATABASE_CONFIGURED | 503 | Configuration | Critical |
| SESSION_NOT_FOUND | 401 | Session | Low |
| SESSION_EXPIRED | 401 | Session | Low |
| SESSION_INVALID | 401 | Session | Medium |
| CONFIGURATION_ERROR | 500 | Configuration | Critical |
| MISSING_CONFIGURATION | 500 | Configuration | Critical |
| INTERNAL_ERROR | 500 | Server | Critical |
| SERVICE_UNAVAILABLE | 503 | Server | Critical |
| NOT_IMPLEMENTED | 501 | Server | Low |
| TENANT_NOT_FOUND | 404 | Multi-Tenancy | Medium |
| TENANT_INACTIVE | 403 | Multi-Tenancy | High |
| TENANT_LIMIT_EXCEEDED | 429 | Multi-Tenancy | Medium |
| WEBSOCKET_UPGRADE_FAILED | 400 | WebSocket | Medium |
| WEBSOCKET_CONNECTION_CLOSED | 400 | WebSocket | Low |
| WEBSOCKET_AUTH_REQUIRED | 401 | WebSocket | High |
| WEBSOCKET_INVALID_MESSAGE | 400 | WebSocket | Low |
| NOT_FOUND | 404 | Routing | Low |
| METHOD_NOT_ALLOWED | 405 | Routing | Low |
| AUTHORIZATION_FAILED | 403 | Routing | High |

---

## See Also

- [Error Handling API](errors-recovery.md) - Error handler interfaces
- [Security API](security.md) - Security-related errors
- [Database API](database.md) - Database error handling
- [Context API](context.md) - Request context and error propagation

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0  
**Total Error Codes**: 47
