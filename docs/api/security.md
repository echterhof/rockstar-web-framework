---
title: "Security API"
description: "SecurityManager interface for authentication, authorization, and security features"
category: "api"
tags: ["api", "security", "authentication", "authorization", "validation"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "../guides/security.md"
---

# Security API

## Overview

The `SecurityManager` interface provides comprehensive security features for the Rockstar Web Framework, including authentication, authorization, request validation, input sanitization, CSRF protection, rate limiting, and cookie encryption. It serves as the central security layer for protecting applications against common vulnerabilities.

**Primary Use Cases:**
- Authenticating users via OAuth2, JWT, or access tokens
- Authorizing user actions based on roles and permissions
- Validating and sanitizing user input
- Protecting against CSRF, XSS, and SQL injection attacks
- Rate limiting to prevent abuse
- Encrypting sensitive cookie data
- Setting security headers

## Type Definition

### SecurityManager Interface

```go
type SecurityManager interface {
    // Authentication methods
    AuthenticateOAuth2(token string) (*User, error)
    AuthenticateJWT(token string) (*User, error)
    AuthenticateAccessToken(token string) (*AccessToken, error)

    // Authorization methods
    Authorize(user *User, resource string, action string) bool
    AuthorizeRole(user *User, role string) bool
    AuthorizeAction(user *User, action string) bool

    // Request validation
    ValidateRequest(ctx Context) error
    ValidateRequestSize(ctx Context, maxSize int64) error
    ValidateRequestTimeout(ctx Context, timeout time.Duration) error
    ValidateBogusData(ctx Context) error

    // Form and file validation
    ValidateFormData(ctx Context, rules ValidationRules) error
    ValidateFileUpload(ctx Context, rules FileValidationRules) error
    ValidateExpectedFormValues(ctx Context, expectedFields []string) error
    ValidateExpectedFiles(ctx Context, expectedFiles []string) error

    // Security headers and protection
    SetSecurityHeaders(ctx Context) error
    SetXFrameOptions(ctx Context, option string) error
    EnableCORS(ctx Context, config CORSConfig) error
    EnableXSSProtection(ctx Context) error
    EnableCSRFProtection(ctx Context) (string, error)
    ValidateCSRFToken(ctx Context, token string) error

    // Input validation
    ValidateInput(input string, rules InputValidationRules) error
    SanitizeInput(input string) string

    // Rate limiting
    CheckRateLimit(ctx Context, resource string) error
    CheckGlobalRateLimit(ctx Context) error

    // Cookie encryption
    EncryptCookie(value string) (string, error)
    DecryptCookie(encryptedValue string) (string, error)
}
```

The SecurityManager interface provides a comprehensive security API. Access it through the framework or context:

```go
// Through framework
security := app.Security()

// Through context (in handlers)
func handler(ctx pkg.Context) error {
    // Security is accessed through context methods
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    return nil
}
```


## Authentication Methods

### AuthenticateOAuth2

```go
func AuthenticateOAuth2(token string) (*User, error)
```

**Description**: Authenticates a user using an OAuth2 access token. Validates the token and returns the associated user.

**Parameters**:
- `token` (string): OAuth2 access token

**Returns**:
- `*User`: Authenticated user object
- `error`: Error if authentication fails

**Example**:
```go
func oauthHandler(ctx pkg.Context) error {
    // Extract token from Authorization header
    authHeader := ctx.GetHeader("Authorization")
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    security := ctx.Framework().Security()
    user, err := security.AuthenticateOAuth2(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid OAuth2 token"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": user.ID,
        "username": user.Username,
    })
}
```

**See Also**:
- [AuthenticateJWT](#authenticatejwt)
- [Security Guide](../guides/security.md#oauth2-authentication)

### AuthenticateJWT

```go
func AuthenticateJWT(token string) (*User, error)
```

**Description**: Authenticates a user using a JWT (JSON Web Token). Validates the token signature and expiration, then returns the associated user.

**Parameters**:
- `token` (string): JWT token string

**Returns**:
- `*User`: Authenticated user object
- `error`: Error if token is invalid or expired

**Example**:
```go
func jwtHandler(ctx pkg.Context) error {
    // Extract JWT from Authorization header
    authHeader := ctx.GetHeader("Authorization")
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    security := ctx.Framework().Security()
    user, err := security.AuthenticateJWT(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid or expired JWT"})
    }
    
    // Store user in context for subsequent handlers
    ctx.Set("user", user)
    
    return ctx.JSON(200, map[string]interface{}{
        "message": fmt.Sprintf("Welcome, %s!", user.Username),
        "user_id": user.ID,
    })
}
```

**See Also**:
- [Security Guide](../guides/security.md#jwt-authentication)


### AuthenticateAccessToken

```go
func AuthenticateAccessToken(token string) (*AccessToken, error)
```

**Description**: Authenticates using an API access token. Returns the access token object containing user information and permissions.

**Parameters**:
- `token` (string): API access token

**Returns**:
- `*AccessToken`: Access token object with user ID and metadata
- `error`: Error if token is invalid or expired

**Example**:
```go
func apiHandler(ctx pkg.Context) error {
    // Extract token from header or query parameter
    token := ctx.GetHeader("X-API-Key")
    if token == "" {
        token = ctx.Query()["api_key"]
    }
    
    security := ctx.Framework().Security()
    accessToken, err := security.AuthenticateAccessToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid API key"})
    }
    
    // Check token expiration
    if time.Now().After(accessToken.ExpiresAt) {
        return ctx.JSON(401, map[string]string{"error": "API key expired"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": accessToken.UserID,
        "token_type": accessToken.TokenType,
    })
}
```

**See Also**:
- [Security Guide](../guides/security.md#access-tokens)

## Authorization Methods

### Authorize

```go
func Authorize(user *User, resource string, action string) bool
```

**Description**: Checks if a user is authorized to perform a specific action on a resource. Uses role-based access control (RBAC) to determine permissions.

**Parameters**:
- `user` (*User): User to check authorization for
- `resource` (string): Resource identifier (e.g., "users", "posts", "admin")
- `action` (string): Action to perform (e.g., "read", "write", "delete")

**Returns**:
- `bool`: true if user is authorized, false otherwise

**Example**:
```go
func deletePostHandler(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    postID := ctx.Param("id")
    
    // Check if user can delete posts
    security := ctx.Framework().Security()
    if !security.Authorize(user, "posts", "delete") {
        return ctx.JSON(403, map[string]string{"error": "Forbidden: insufficient permissions"})
    }
    
    // Delete the post
    db := ctx.DB()
    _, err := db.Exec("DELETE FROM posts WHERE id = ?", postID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to delete post"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Post deleted"})
}
```

**See Also**:
- [AuthorizeRole](#authorizerole)
- [AuthorizeAction](#authorizeaction)
- [Security Guide](../guides/security.md#authorization)


### AuthorizeRole

```go
func AuthorizeRole(user *User, role string) bool
```

**Description**: Checks if a user has a specific role.

**Parameters**:
- `user` (*User): User to check
- `role` (string): Role name (e.g., "admin", "moderator", "user")

**Returns**:
- `bool`: true if user has the role, false otherwise

**Example**:
```go
func adminPanelHandler(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    security := ctx.Framework().Security()
    if !security.AuthorizeRole(user, "admin") {
        return ctx.JSON(403, map[string]string{"error": "Admin access required"})
    }
    
    // Show admin panel
    return ctx.JSON(200, map[string]string{"message": "Welcome to admin panel"})
}

// Middleware example
func requireRole(role string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        user := ctx.User()
        if user == nil {
            return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
        }
        
        security := ctx.Framework().Security()
        if !security.AuthorizeRole(user, role) {
            return ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
        }
        
        return next(ctx)
    }
}
```

### AuthorizeAction

```go
func AuthorizeAction(user *User, action string) bool
```

**Description**: Checks if a user is authorized to perform a specific action, regardless of resource.

**Parameters**:
- `user` (*User): User to check
- `action` (string): Action name (e.g., "create_user", "delete_post", "view_analytics")

**Returns**:
- `bool`: true if user can perform the action, false otherwise

**Example**:
```go
func createUserHandler(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    security := ctx.Framework().Security()
    if !security.AuthorizeAction(user, "create_user") {
        return ctx.JSON(403, map[string]string{"error": "Cannot create users"})
    }
    
    // Create user...
    return ctx.JSON(201, map[string]string{"message": "User created"})
}
```


## Request Validation Methods

### ValidateRequest

```go
func ValidateRequest(ctx Context) error
```

**Description**: Performs comprehensive request validation including size limits, timeout checks, bogus data detection, and input length validation.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `error`: Error if validation fails, nil if request is valid

**Example**:
```go
// Use as middleware
func validationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    if err := security.ValidateRequest(ctx); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    return next(ctx)
}

// Use in handler
func handler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    if err := security.ValidateRequest(ctx); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid request"})
    }
    
    // Process valid request...
    return ctx.JSON(200, map[string]string{"message": "Success"})
}
```

**See Also**:
- [ValidateRequestSize](#validaterequestsize)
- [ValidateRequestTimeout](#validaterequesttimeout)
- [ValidateBogusData](#validatebogusdata)

### ValidateRequestSize

```go
func ValidateRequestSize(ctx Context, maxSize int64) error
```

**Description**: Validates that the request size does not exceed the specified maximum.

**Parameters**:
- `ctx` (Context): Request context
- `maxSize` (int64): Maximum allowed request size in bytes

**Returns**:
- `error`: Error if request exceeds size limit

**Example**:
```go
func uploadHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Limit request to 10MB
    maxSize := int64(10 * 1024 * 1024)
    if err := security.ValidateRequestSize(ctx, maxSize); err != nil {
        return ctx.JSON(413, map[string]string{"error": "Request too large"})
    }
    
    // Process upload...
    return ctx.JSON(200, map[string]string{"message": "Upload successful"})
}
```

### ValidateRequestTimeout

```go
func ValidateRequestTimeout(ctx Context, timeout time.Duration) error
```

**Description**: Validates that the request has not exceeded the specified timeout duration.

**Parameters**:
- `ctx` (Context): Request context
- `timeout` (time.Duration): Maximum allowed request duration

**Returns**:
- `error`: Error if request has timed out

**Example**:
```go
func longRunningHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Check if request has exceeded 30 second timeout
    if err := security.ValidateRequestTimeout(ctx, 30*time.Second); err != nil {
        return ctx.JSON(408, map[string]string{"error": "Request timeout"})
    }
    
    // Continue processing...
    return ctx.JSON(200, result)
}
```


### ValidateBogusData

```go
func ValidateBogusData(ctx Context) error
```

**Description**: Detects and validates against bogus or malformed data such as null bytes in URLs, headers, or request body.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `error`: Error if bogus data is detected

**Example**:
```go
func secureHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Check for malicious data patterns
    if err := security.ValidateBogusData(ctx); err != nil {
        ctx.Logger().Warn("Bogus data detected", map[string]interface{}{
            "error": err.Error(),
            "ip": ctx.Request().RemoteAddr,
        })
        return ctx.JSON(400, map[string]string{"error": "Invalid request data"})
    }
    
    // Process request...
    return ctx.JSON(200, map[string]string{"message": "Success"})
}
```

## Form and File Validation Methods

### ValidateFormData

```go
func ValidateFormData(ctx Context, rules ValidationRules) error
```

**Description**: Validates form data against specified validation rules including required fields, types, lengths, and patterns.

**Parameters**:
- `ctx` (Context): Request context
- `rules` (ValidationRules): Validation rules to apply

**Returns**:
- `error`: Error if validation fails

**Example**:
```go
func registerHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Define validation rules
    rules := pkg.ValidationRules{
        Required: []string{"username", "email", "password"},
        Types: map[string]string{
            "email": "email",
            "age":   "int",
        },
        Lengths: map[string]pkg.LengthRule{
            "username": {Min: 3, Max: 20},
            "password": {Min: 8, Max: 100},
        },
        Patterns: map[string]string{
            "username": "^[a-zA-Z0-9_]+$",
        },
    }
    
    // Validate form data
    if err := security.ValidateFormData(ctx, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Process registration...
    return ctx.JSON(201, map[string]string{"message": "User registered"})
}
```

**See Also**:
- [ValidationRules Type](#validationrules-type)


### ValidateFileUpload

```go
func ValidateFileUpload(ctx Context, rules FileValidationRules) error
```

**Description**: Validates uploaded files against specified rules including size limits, allowed types, and extensions.

**Parameters**:
- `ctx` (Context): Request context
- `rules` (FileValidationRules): File validation rules

**Returns**:
- `error`: Error if validation fails

**Example**:
```go
func uploadAvatarHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Define file validation rules
    rules := pkg.FileValidationRules{
        MaxSize:      5 * 1024 * 1024, // 5MB
        AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
        AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif"},
        Required:     []string{"avatar"},
        MaxFiles:     1,
    }
    
    // Validate uploaded files
    if err := security.ValidateFileUpload(ctx, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Process file upload...
    file, _ := ctx.FormFile("avatar")
    return ctx.JSON(200, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
    })
}
```

**See Also**:
- [FileValidationRules Type](#filevalidationrules-type)

### ValidateExpectedFormValues

```go
func ValidateExpectedFormValues(ctx Context, expectedFields []string) error
```

**Description**: Validates that all expected form fields are present in the request.

**Parameters**:
- `ctx` (Context): Request context
- `expectedFields` ([]string): List of required field names

**Returns**:
- `error`: Error if any expected field is missing

**Example**:
```go
func submitFormHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Ensure required fields are present
    expectedFields := []string{"name", "email", "message"}
    if err := security.ValidateExpectedFormValues(ctx, expectedFields); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Missing required fields"})
    }
    
    // Process form...
    return ctx.JSON(200, map[string]string{"message": "Form submitted"})
}
```

### ValidateExpectedFiles

```go
func ValidateExpectedFiles(ctx Context, expectedFiles []string) error
```

**Description**: Validates that all expected file uploads are present in the request.

**Parameters**:
- `ctx` (Context): Request context
- `expectedFiles` ([]string): List of required file field names

**Returns**:
- `error`: Error if any expected file is missing

**Example**:
```go
func uploadDocumentsHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Ensure required files are uploaded
    expectedFiles := []string{"resume", "cover_letter"}
    if err := security.ValidateExpectedFiles(ctx, expectedFiles); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Missing required files"})
    }
    
    // Process uploads...
    return ctx.JSON(200, map[string]string{"message": "Documents uploaded"})
}
```


## Security Headers and Protection Methods

### SetSecurityHeaders

```go
func SetSecurityHeaders(ctx Context) error
```

**Description**: Sets all recommended security headers including X-Frame-Options, X-Content-Type-Options, XSS protection, HSTS, and more.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `error`: Error if setting headers fails

**Example**:
```go
// Use as middleware
func securityHeadersMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    if err := security.SetSecurityHeaders(ctx); err != nil {
        return err
    }
    return next(ctx)
}

// Use in handler
func handler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    security.SetSecurityHeaders(ctx)
    
    return ctx.JSON(200, map[string]string{"message": "Success"})
}
```

**Headers Set**:
- `X-Frame-Options`: Prevents clickjacking
- `X-Content-Type-Options`: Prevents MIME sniffing
- `X-XSS-Protection`: Enables XSS filtering
- `Strict-Transport-Security`: Enforces HTTPS
- `Referrer-Policy`: Controls referrer information
- `Permissions-Policy`: Controls browser features

**See Also**:
- [Security Guide](../guides/security.md#security-headers)

### SetXFrameOptions

```go
func SetXFrameOptions(ctx Context, option string) error
```

**Description**: Sets the X-Frame-Options header to prevent clickjacking attacks.

**Parameters**:
- `ctx` (Context): Request context
- `option` (string): Frame option value ("DENY", "SAMEORIGIN", or "ALLOW-FROM uri")

**Returns**:
- `error`: Error if option is invalid

**Example**:
```go
func handler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Prevent page from being framed
    if err := security.SetXFrameOptions(ctx, "DENY"); err != nil {
        return err
    }
    
    return ctx.HTML(200, "page.html", data)
}

func embedHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Allow framing from same origin
    security.SetXFrameOptions(ctx, "SAMEORIGIN")
    
    return ctx.HTML(200, "embed.html", data)
}
```


### EnableCORS

```go
func EnableCORS(ctx Context, config CORSConfig) error
```

**Description**: Enables Cross-Origin Resource Sharing (CORS) with the specified configuration.

**Parameters**:
- `ctx` (Context): Request context
- `config` (CORSConfig): CORS configuration

**Returns**:
- `error`: Error if origin is not allowed

**Example**:
```go
func apiHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Configure CORS
    corsConfig := pkg.CORSConfig{
        AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           3600,
    }
    
    if err := security.EnableCORS(ctx, corsConfig); err != nil {
        return ctx.JSON(403, map[string]string{"error": "Origin not allowed"})
    }
    
    // Handle preflight request
    if ctx.Request().Method == "OPTIONS" {
        return ctx.String(204, "")
    }
    
    return ctx.JSON(200, data)
}
```

**Security Warning**: Never use wildcard (`*`) for `AllowOrigins` in production with `AllowCredentials: true`.

**See Also**:
- [CORSConfig Type](#corsconfig-type)
- [Security Guide](../guides/security.md#cors)

### EnableXSSProtection

```go
func EnableXSSProtection(ctx Context) error
```

**Description**: Enables XSS (Cross-Site Scripting) protection headers.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `error`: Error if setting headers fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Enable XSS protection
    if err := security.EnableXSSProtection(ctx); err != nil {
        return err
    }
    
    return ctx.HTML(200, "page.html", data)
}
```

**Headers Set**:
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`


### EnableCSRFProtection

```go
func EnableCSRFProtection(ctx Context) (string, error)
```

**Description**: Generates a CSRF token and sets it in a secure cookie. Returns the token for inclusion in forms.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `string`: Generated CSRF token
- `error`: Error if token generation fails

**Example**:
```go
func showFormHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Generate CSRF token
    csrfToken, err := security.EnableCSRFProtection(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to generate CSRF token"})
    }
    
    // Pass token to template
    data := map[string]interface{}{
        "csrf_token": csrfToken,
    }
    
    return ctx.HTML(200, "form.html", data)
}

// In HTML template:
// <form method="POST">
//   <input type="hidden" name="csrf_token" value="{{.csrf_token}}">
//   <!-- form fields -->
// </form>
```

**See Also**:
- [ValidateCSRFToken](#validatecsrftoken)
- [Security Guide](../guides/security.md#csrf-protection)

### ValidateCSRFToken

```go
func ValidateCSRFToken(ctx Context, token string) error
```

**Description**: Validates a CSRF token against the token stored in the cookie.

**Parameters**:
- `ctx` (Context): Request context
- `token` (string): CSRF token from form submission

**Returns**:
- `error`: Error if token is invalid, expired, or doesn't match

**Example**:
```go
func submitFormHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Get CSRF token from form
    csrfToken := ctx.FormValue("csrf_token")
    
    // Validate token
    if err := security.ValidateCSRFToken(ctx, csrfToken); err != nil {
        return ctx.JSON(403, map[string]string{"error": "Invalid CSRF token"})
    }
    
    // Process form submission...
    return ctx.JSON(200, map[string]string{"message": "Form submitted"})
}

// Middleware example
func csrfMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Skip CSRF check for GET, HEAD, OPTIONS
    if ctx.Request().Method == "GET" || ctx.Request().Method == "HEAD" || ctx.Request().Method == "OPTIONS" {
        return next(ctx)
    }
    
    security := ctx.Framework().Security()
    csrfToken := ctx.FormValue("csrf_token")
    
    if err := security.ValidateCSRFToken(ctx, csrfToken); err != nil {
        return ctx.JSON(403, map[string]string{"error": "CSRF validation failed"})
    }
    
    return next(ctx)
}
```


## Input Validation and Sanitization Methods

### ValidateInput

```go
func ValidateInput(input string, rules InputValidationRules) error
```

**Description**: Validates input string against specified rules including length, pattern matching, and content restrictions.

**Parameters**:
- `input` (string): Input string to validate
- `rules` (InputValidationRules): Validation rules

**Returns**:
- `error`: Error if validation fails

**Example**:
```go
func commentHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    comment := ctx.FormValue("comment")
    
    // Define validation rules
    rules := pkg.InputValidationRules{
        MaxLength:      500,
        AllowHTML:      false,
        TrimWhitespace: true,
        Pattern:        "^[a-zA-Z0-9\\s.,!?'-]+$",
    }
    
    // Validate input
    if err := security.ValidateInput(comment, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid comment"})
    }
    
    // Save comment...
    return ctx.JSON(201, map[string]string{"message": "Comment posted"})
}
```

**See Also**:
- [InputValidationRules Type](#inputvalidationrules-type)
- [SanitizeInput](#sanitizeinput)

### SanitizeInput

```go
func SanitizeInput(input string) string
```

**Description**: Sanitizes input by removing or escaping dangerous content including HTML entities and null bytes.

**Parameters**:
- `input` (string): Input string to sanitize

**Returns**:
- `string`: Sanitized input string

**Example**:
```go
func handler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Get user input
    userInput := ctx.FormValue("content")
    
    // Sanitize input
    sanitized := security.SanitizeInput(userInput)
    
    // Store sanitized input
    db := ctx.DB()
    _, err := db.Exec("INSERT INTO content (text) VALUES (?)", sanitized)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "original":  userInput,
        "sanitized": sanitized,
    })
}
```

**Sanitization Actions**:
- Trims whitespace
- Escapes HTML entities
- Removes null bytes
- Prevents XSS attacks


## Rate Limiting Methods

### CheckRateLimit

```go
func CheckRateLimit(ctx Context, resource string) error
```

**Description**: Checks if the client has exceeded the rate limit for a specific resource.

**Parameters**:
- `ctx` (Context): Request context
- `resource` (string): Resource identifier (e.g., "api", "login", "upload")

**Returns**:
- `error`: Error if rate limit is exceeded

**Example**:
```go
func loginHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Check rate limit for login attempts
    if err := security.CheckRateLimit(ctx, "login"); err != nil {
        return ctx.JSON(429, map[string]string{
            "error": "Too many login attempts. Please try again later.",
        })
    }
    
    // Process login...
    return ctx.JSON(200, map[string]string{"message": "Login successful"})
}

func apiHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Check API rate limit
    if err := security.CheckRateLimit(ctx, "api"); err != nil {
        ctx.SetHeader("Retry-After", "60")
        return ctx.JSON(429, map[string]string{
            "error": "Rate limit exceeded",
        })
    }
    
    // Process API request...
    return ctx.JSON(200, data)
}
```

**See Also**:
- [CheckGlobalRateLimit](#checkglobalratelimit)
- [Security Guide](../guides/security.md#rate-limiting)

### CheckGlobalRateLimit

```go
func CheckGlobalRateLimit(ctx Context) error
```

**Description**: Checks if the client has exceeded the global rate limit across all resources.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `error`: Error if global rate limit is exceeded

**Example**:
```go
// Use as middleware
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    
    // Check global rate limit
    if err := security.CheckGlobalRateLimit(ctx); err != nil {
        ctx.SetHeader("Retry-After", "60")
        return ctx.JSON(429, map[string]string{
            "error": "Too many requests. Please slow down.",
        })
    }
    
    return next(ctx)
}
```


## Cookie Encryption Methods

### EncryptCookie

```go
func EncryptCookie(value string) (string, error)
```

**Description**: Encrypts a cookie value using AES-GCM encryption for secure storage of sensitive data.

**Parameters**:
- `value` (string): Plain text value to encrypt

**Returns**:
- `string`: Base64-encoded encrypted value
- `error`: Error if encryption fails

**Example**:
```go
func setSecureCookieHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Sensitive data to store in cookie
    sessionData := fmt.Sprintf("user_id=%s&role=%s", userID, role)
    
    // Encrypt the value
    encrypted, err := security.EncryptCookie(sessionData)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Encryption failed"})
    }
    
    // Set encrypted cookie
    cookie := &pkg.Cookie{
        Name:     "session_data",
        Value:    encrypted,
        Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }
    
    ctx.SetCookie(cookie)
    return ctx.JSON(200, map[string]string{"message": "Secure cookie set"})
}
```

**See Also**:
- [DecryptCookie](#decryptcookie)
- [Security Guide](../guides/security.md#cookie-encryption)

### DecryptCookie

```go
func DecryptCookie(encryptedValue string) (string, error)
```

**Description**: Decrypts an encrypted cookie value.

**Parameters**:
- `encryptedValue` (string): Base64-encoded encrypted value

**Returns**:
- `string`: Decrypted plain text value
- `error`: Error if decryption fails

**Example**:
```go
func getSecureCookieHandler(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Get encrypted cookie
    cookie, err := ctx.GetCookie("session_data")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "No session cookie"})
    }
    
    // Decrypt the value
    decrypted, err := security.DecryptCookie(cookie.Value)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid session cookie"})
    }
    
    // Parse decrypted data
    // decrypted = "user_id=123&role=admin"
    
    return ctx.JSON(200, map[string]string{
        "session_data": decrypted,
    })
}
```

**Security Note**: Cookie encryption uses AES-256-GCM with random nonces for each encryption operation.


## Supporting Types

### ValidationRules Type

```go
type ValidationRules struct {
    Required []string                   // Required field names
    Types    map[string]string          // field -> type (string, int, email, url, bool, float)
    Lengths  map[string]LengthRule      // field -> length constraints
    Patterns map[string]string          // field -> regex pattern
    Custom   map[string]CustomValidator // field -> custom validator function
}

type LengthRule struct {
    Min int // Minimum length
    Max int // Maximum length
}

type CustomValidator func(value interface{}) error
```

**Example**:
```go
rules := pkg.ValidationRules{
    Required: []string{"username", "email", "password"},
    Types: map[string]string{
        "email": "email",
        "age":   "int",
        "website": "url",
    },
    Lengths: map[string]pkg.LengthRule{
        "username": {Min: 3, Max: 20},
        "password": {Min: 8, Max: 100},
    },
    Patterns: map[string]string{
        "username": "^[a-zA-Z0-9_]+$",
        "phone":    "^\\+?[1-9]\\d{1,14}$",
    },
    Custom: map[string]pkg.CustomValidator{
        "password": func(value interface{}) error {
            pwd := value.(string)
            if !hasUpperCase(pwd) || !hasLowerCase(pwd) || !hasDigit(pwd) {
                return errors.New("password must contain uppercase, lowercase, and digit")
            }
            return nil
        },
    },
}
```

### FileValidationRules Type

```go
type FileValidationRules struct {
    MaxSize      int64    // Maximum file size in bytes
    AllowedTypes []string // Allowed MIME types
    AllowedExts  []string // Allowed file extensions
    Required     []string // Required file fields
    MaxFiles     int      // Maximum number of files
}
```

**Example**:
```go
rules := pkg.FileValidationRules{
    MaxSize:      10 * 1024 * 1024, // 10MB
    AllowedTypes: []string{"image/jpeg", "image/png", "application/pdf"},
    AllowedExts:  []string{".jpg", ".jpeg", ".png", ".pdf"},
    Required:     []string{"document"},
    MaxFiles:     5,
}
```


### InputValidationRules Type

```go
type InputValidationRules struct {
    AllowHTML      bool     // Allow HTML tags
    AllowedTags    []string // Allowed HTML tags (if AllowHTML is true)
    MaxLength      int      // Maximum input length
    Pattern        string   // Regex pattern to match
    StripTags      bool     // Strip HTML tags
    EscapeHTML     bool     // Escape HTML entities
    TrimWhitespace bool     // Trim leading/trailing whitespace
}
```

**Example**:
```go
// Strict validation - no HTML
strictRules := pkg.InputValidationRules{
    MaxLength:      1000,
    AllowHTML:      false,
    TrimWhitespace: true,
    Pattern:        "^[a-zA-Z0-9\\s.,!?'-]+$",
}

// Allow limited HTML
htmlRules := pkg.InputValidationRules{
    MaxLength:   5000,
    AllowHTML:   true,
    AllowedTags: []string{"p", "br", "strong", "em", "a"},
    StripTags:   true, // Strip tags not in AllowedTags
}

// Sanitize and escape
sanitizeRules := pkg.InputValidationRules{
    MaxLength:      2000,
    EscapeHTML:     true,
    TrimWhitespace: true,
}
```

### CORSConfig Type

```go
type CORSConfig struct {
    AllowOrigins     []string // Allowed origins (e.g., ["https://example.com"])
    AllowMethods     []string // Allowed HTTP methods
    AllowHeaders     []string // Allowed request headers
    ExposeHeaders    []string // Headers exposed to client
    AllowCredentials bool     // Allow credentials (cookies, auth headers)
    MaxAge           int      // Preflight cache duration in seconds
}
```

**Example**:
```go
// Production CORS config
corsConfig := pkg.CORSConfig{
    AllowOrigins:     []string{"https://app.example.com", "https://admin.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
    ExposeHeaders:    []string{"X-Request-ID", "X-Response-Time"},
    AllowCredentials: true,
    MaxAge:           3600, // 1 hour
}

// Development CORS config (less restrictive)
devCorsConfig := pkg.CORSConfig{
    AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
    AllowHeaders:     []string{"*"},
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
```

**Security Warning**: Never use `AllowOrigins: []string{"*"}` with `AllowCredentials: true` in production.


## Complete Usage Example

Here's a comprehensive example demonstrating multiple security features:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

func main() {
    // Create framework with security configuration
    config := pkg.DefaultConfig()
    config.Security = pkg.SecurityConfig{
        MaxRequestSize:   10 * 1024 * 1024, // 10MB
        RequestTimeout:   30 * time.Second,
        CSRFTokenExpiry:  24 * time.Hour,
        EncryptionKey:    "your-32-byte-hex-key-here",
        JWTSecret:        "your-jwt-secret",
        EnableXSSProtect: true,
        EnableCSRF:       true,
        EnableHSTS:       true,
    }
    
    app, err := pkg.New(config)
    if err != nil {
        panic(err)
    }
    
    // Security middleware
    app.Use(securityMiddleware)
    app.Use(rateLimitMiddleware)
    
    // Public routes
    app.GET("/login", showLoginForm)
    app.POST("/login", handleLogin)
    
    // Protected routes
    api := app.Group("/api")
    api.Use(authMiddleware)
    api.GET("/users", listUsers)
    api.POST("/users", createUser)
    api.DELETE("/users/:id", deleteUser)
    
    app.Start(":8080")
}

// Security headers middleware
func securityMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    
    // Set all security headers
    if err := security.SetSecurityHeaders(ctx); err != nil {
        return err
    }
    
    // Validate request
    if err := security.ValidateRequest(ctx); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid request"})
    }
    
    return next(ctx)
}

// Rate limiting middleware
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    
    if err := security.CheckGlobalRateLimit(ctx); err != nil {
        return ctx.JSON(429, map[string]string{"error": "Too many requests"})
    }
    
    return next(ctx)
}

// Authentication middleware
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    security := ctx.Framework().Security()
    
    // Extract JWT from Authorization header
    authHeader := ctx.GetHeader("Authorization")
    if authHeader == "" {
        return ctx.JSON(401, map[string]string{"error": "Missing authorization"})
    }
    
    token := strings.TrimPrefix(authHeader, "Bearer ")
    user, err := security.AuthenticateJWT(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid token"})
    }
    
    // Store user in context
    ctx.Set("user", user)
    
    return next(ctx)
}

// Show login form with CSRF protection
func showLoginForm(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    csrfToken, err := security.EnableCSRFProtection(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to generate CSRF token"})
    }
    
    return ctx.HTML(200, "login.html", map[string]interface{}{
        "csrf_token": csrfToken,
    })
}

// Handle login with validation
func handleLogin(ctx pkg.Context) error {
    security := ctx.Framework().Security()
    
    // Check rate limit for login attempts
    if err := security.CheckRateLimit(ctx, "login"); err != nil {
        return ctx.JSON(429, map[string]string{"error": "Too many login attempts"})
    }
    
    // Validate CSRF token
    csrfToken := ctx.FormValue("csrf_token")
    if err := security.ValidateCSRFToken(ctx, csrfToken); err != nil {
        return ctx.JSON(403, map[string]string{"error": "Invalid CSRF token"})
    }
    
    // Validate form data
    rules := pkg.ValidationRules{
        Required: []string{"username", "password"},
        Lengths: map[string]pkg.LengthRule{
            "username": {Min: 3, Max: 50},
            "password": {Min: 8, Max: 100},
        },
    }
    
    if err := security.ValidateFormData(ctx, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Authenticate user...
    // Generate JWT and return
    
    return ctx.JSON(200, map[string]string{"message": "Login successful"})
}

// List users (requires authentication)
func listUsers(ctx pkg.Context) error {
    user := ctx.Get("user").(*pkg.User)
    security := ctx.Framework().Security()
    
    // Check authorization
    if !security.Authorize(user, "users", "read") {
        return ctx.JSON(403, map[string]string{"error": "Forbidden"})
    }
    
    // Fetch users from database
    db := ctx.DB()
    users, err := db.Query("SELECT id, username, email FROM users")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Database error"})
    }
    
    return ctx.JSON(200, users)
}

// Create user (requires admin role)
func createUser(ctx pkg.Context) error {
    user := ctx.Get("user").(*pkg.User)
    security := ctx.Framework().Security()
    
    // Check if user has admin role
    if !security.AuthorizeRole(user, "admin") {
        return ctx.JSON(403, map[string]string{"error": "Admin access required"})
    }
    
    // Validate input
    rules := pkg.ValidationRules{
        Required: []string{"username", "email", "password"},
        Types: map[string]string{
            "email": "email",
        },
        Lengths: map[string]pkg.LengthRule{
            "username": {Min: 3, Max: 50},
            "password": {Min: 8, Max: 100},
        },
    }
    
    if err := security.ValidateFormData(ctx, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Create user...
    
    return ctx.JSON(201, map[string]string{"message": "User created"})
}

// Delete user (requires delete permission)
func deleteUser(ctx pkg.Context) error {
    user := ctx.Get("user").(*pkg.User)
    security := ctx.Framework().Security()
    
    // Check authorization
    if !security.Authorize(user, "users", "delete") {
        return ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
    }
    
    userID := ctx.Param("id")
    
    // Delete user from database
    db := ctx.DB()
    _, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to delete user"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "User deleted"})
}
```

## Best Practices

1. **Always validate user input**: Use `ValidateFormData` and `ValidateInput` for all user-provided data
2. **Sanitize output**: Use `SanitizeInput` before storing or displaying user content
3. **Enable security headers**: Call `SetSecurityHeaders` in middleware for all responses
4. **Use CSRF protection**: Enable CSRF tokens for all state-changing operations
5. **Implement rate limiting**: Protect sensitive endpoints like login and API routes
6. **Encrypt sensitive cookies**: Use `EncryptCookie` for session data and sensitive information
7. **Validate file uploads**: Always validate file size, type, and extension
8. **Use HTTPS in production**: Enable HSTS and set `Secure` flag on cookies
9. **Implement proper authorization**: Check both authentication and authorization for protected resources
10. **Log security events**: Log failed authentication attempts, rate limit violations, and suspicious activity

## See Also

- [Security Guide](../guides/security.md) - Comprehensive security implementation guide
- [Context API](context.md) - Request context and handler interface
- [Framework API](framework.md) - Framework initialization and configuration
- [Authentication Example](../examples/secure-app.md) - Complete authentication example

