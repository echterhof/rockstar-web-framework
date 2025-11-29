# Security Guide

## Overview

The Rockstar Web Framework provides comprehensive, enterprise-grade security features designed to protect your applications from common vulnerabilities and attacks. This guide covers authentication, authorization, CSRF/XSS protection, session security, input validation, and security best practices.

Security is built into the framework's core architecture, with features including:

- **Authentication**: OAuth2, JWT, and access token support
- **Authorization**: Role-Based Access Control (RBAC) and action-based permissions
- **Protection**: CSRF, XSS, SQL injection, and other attack prevention
- **Session Security**: Encrypted sessions with secure cookie handling
- **Input Validation**: Comprehensive validation and sanitization
- **Rate Limiting**: Request throttling to prevent abuse
- **Security Headers**: Automatic security header management

## Quick Start

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
        EnableCSRF:       true,
        EnableXSSProtect: true,
        EnableHSTS:       true,
        JWTSecret:        "your-secret-key-here",
        EncryptionKey:    "your-32-byte-encryption-key-hex",
    }
    
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Protected route with authentication
    router.GET("/api/protected", protectedHandler, authMiddleware)
    
    app.Start()
}

func authMiddleware(ctx pkg.Context) error {
    // Extract JWT token from Authorization header
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return ctx.JSON(401, map[string]string{"error": "unauthorized"})
    }
    
    // Authenticate user
    security := ctx.Security()
    user, err := security.AuthenticateJWT(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "invalid token"})
    }
    
    // Set user in context
    ctx.SetUser(user)
    return ctx.Next()
}

func protectedHandler(ctx pkg.Context) error {
    user := ctx.User()
    return ctx.JSON(200, map[string]interface{}{
        "message": "Welcome to protected resource",
        "user":    user.Username,
    })
}
```

## Authentication

The framework supports multiple authentication mechanisms to suit different application needs.

### OAuth2 Authentication

OAuth2 is an industry-standard protocol for authorization, commonly used for third-party authentication.

```go
// Configure OAuth2
config := pkg.OAuth2Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    TokenURL:     "https://oauth-provider.com/token",
    AuthURL:      "https://oauth-provider.com/authorize",
    RedirectURL:  "https://yourapp.com/callback",
    Scopes:       []string{"read", "write"},
}

authManager := pkg.NewAuthManager(app.Database(), "jwt-secret", config)

// Authenticate with OAuth2 token
router.GET("/api/oauth/callback", func(ctx pkg.Context) error {
    token := ctx.Query("token")
    
    user, err := authManager.AuthenticateOAuth2(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "authentication failed"})
    }
    
    // User is authenticated
    ctx.SetUser(user)
    return ctx.JSON(200, map[string]interface{}{
        "message": "authenticated",
        "user":    user,
    })
})
```

**OAuth2 Features:**
- Token validation with expiration checking
- Scope-based access control
- Automatic token refresh support
- Integration with popular OAuth2 providers

### JWT Authentication

JSON Web Tokens (JWT) provide stateless authentication with cryptographic signatures.

```go
// Authenticate with JWT
router.POST("/api/login", func(ctx pkg.Context) error {
    var credentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := ctx.BindJSON(&credentials); err != nil {
        return ctx.JSON(400, map[string]string{"error": "invalid request"})
    }
    
    // Validate credentials (implement your logic)
    user := validateCredentials(credentials.Username, credentials.Password)
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "invalid credentials"})
    }
    
    // Generate JWT token
    authManager := pkg.NewAuthManager(app.Database(), "jwt-secret", pkg.OAuth2Config{})
    token, err := authManager.GenerateJWT(user, 24*time.Hour)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "token generation failed"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "token": token,
        "user":  user,
    })
})

// Verify JWT in middleware
func jwtMiddleware(ctx pkg.Context) error {
    authHeader := ctx.GetHeader("Authorization")
    if authHeader == "" {
        return ctx.JSON(401, map[string]string{"error": "missing authorization header"})
    }
    
    // Extract token (format: "Bearer <token>")
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    security := ctx.Security()
    user, err := security.AuthenticateJWT(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "invalid token"})
    }
    
    ctx.SetUser(user)
    return ctx.Next()
}
```

**JWT Features:**
- HMAC-SHA256 signature verification
- Automatic expiration checking
- Custom claims support (roles, actions, scopes)
- Tenant-aware authentication
- Metadata storage in token

**JWT Claims Structure:**
```go
type JWTClaims struct {
    UserID    string                 `json:"user_id"`
    Username  string                 `json:"username"`
    Email     string                 `json:"email"`
    Roles     []string               `json:"roles"`
    Actions   []string               `json:"actions"`
    Scopes    []string               `json:"scopes"`
    TenantID  string                 `json:"tenant_id"`
    Metadata  map[string]interface{} `json:"metadata"`
    IssuedAt  int64                  `json:"iat"`
    ExpiresAt int64                  `json:"exp"`
    Issuer    string                 `json:"iss"`
    Subject   string                 `json:"sub"`
}
```

### Access Token Authentication

Access tokens provide a simple, database-backed authentication mechanism.

```go
// Create access token
authManager := pkg.NewAuthManager(app.Database(), "jwt-secret", pkg.OAuth2Config{})

token, err := authManager.CreateAccessToken(
    userID,
    tenantID,
    []string{"read", "write"},
    7*24*time.Hour, // 7 days
)
if err != nil {
    return err
}

// Authenticate with access token
router.GET("/api/data", func(ctx pkg.Context) error {
    token := ctx.GetHeader("X-API-Token")
    
    security := ctx.Security()
    accessToken, err := security.AuthenticateAccessToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "invalid token"})
    }
    
    // Token is valid
    return ctx.JSON(200, map[string]interface{}{
        "data": "sensitive data",
    })
})


// Refresh access token
newToken, err := authManager.RefreshAccessToken(oldToken, 7*24*time.Hour)

// Revoke access token
err = authManager.RevokeAccessToken(token.Token)
```

**Access Token Features:**
- Database-backed storage
- Automatic expiration
- Scope-based permissions
- Token refresh and revocation
- Tenant isolation

## Authorization

The framework provides flexible authorization mechanisms including Role-Based Access Control (RBAC) and action-based permissions.

### Role-Based Access Control (RBAC)

RBAC allows you to control access based on user roles.

```go
// Check single role
authManager := pkg.NewAuthManager(app.Database(), "jwt-secret", pkg.OAuth2Config{})

err := authManager.AuthorizeRole(user, "admin")
if err != nil {
    return ctx.JSON(403, map[string]string{"error": "insufficient permissions"})
}

// Check any of multiple roles (OR logic)
err = authManager.AuthorizeRoles(user, []string{"admin", "moderator"})

// Check all roles required (AND logic)
err = authManager.AuthorizeAllRoles(user, []string{"admin", "superuser"})

// Middleware example
func requireRole(role string) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        user := ctx.User()
        if user == nil {
            return ctx.JSON(401, map[string]string{"error": "unauthorized"})
        }
        
        authManager := pkg.NewAuthManager(ctx.Database(), "jwt-secret", pkg.OAuth2Config{})
        if err := authManager.AuthorizeRole(user, role); err != nil {
            return ctx.JSON(403, map[string]string{"error": "forbidden"})
        }
        
        return ctx.Next()
    }
}

// Use in routes
router.GET("/admin/dashboard", adminHandler, requireRole("admin"))
router.GET("/moderator/panel", modHandler, requireRole("moderator"))
```

### Action-Based Authorization

Action-based authorization provides fine-grained control over specific operations.

```go
// Check single action
err := authManager.AuthorizeAction(user, "posts:delete")

// Check any of multiple actions (OR logic)
err = authManager.AuthorizeActions(user, []string{"posts:edit", "posts:delete"})

// Check all actions required (AND logic)
err = authManager.AuthorizeAllActions(user, []string{"posts:create", "posts:publish"})

// Middleware example
func requireAction(action string) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        user := ctx.User()
        if user == nil {
            return ctx.JSON(401, map[string]string{"error": "unauthorized"})
        }
        
        authManager := pkg.NewAuthManager(ctx.Database(), "jwt-secret", pkg.OAuth2Config{})
        if err := authManager.AuthorizeAction(user, action); err != nil {
            return ctx.JSON(403, map[string]string{"error": "forbidden"})
        }
        
        return ctx.Next()
    }
}

// Use in routes
router.DELETE("/api/posts/:id", deletePostHandler, requireAction("posts:delete"))
router.POST("/api/posts", createPostHandler, requireAction("posts:create"))
```


### Scope-Based Authorization

Scopes provide hierarchical access control, commonly used with OAuth2.

```go
// Check single scope
err := authManager.AuthorizeScope(user, "api:read")

// Check any of multiple scopes (OR logic)
err = authManager.AuthorizeScopes(user, []string{"api:read", "api:write"})

// Check all scopes required (AND logic)
err = authManager.AuthorizeAllScopes(user, []string{"api:read", "api:write"})

// Hierarchical scope matching
// User with "admin" scope can access "admin:read" and "admin:write"
user.Scopes = []string{"admin"}
err := authManager.AuthorizeScope(user, "admin:read") // ✓ Allowed

// Wildcard scope
user.Scopes = []string{"*"}
err := authManager.AuthorizeScope(user, "any:scope") // ✓ Allowed
```

### Combined Authorization

Combine roles and actions for comprehensive access control.

```go
// Check both roles and actions
err := authManager.Authorize(
    user,
    []string{"admin", "moderator"}, // Required roles (any)
    []string{"posts:delete"},        // Required actions (any)
)

// Custom authorization logic
func authorizePostAccess(ctx pkg.Context, postID string) error {
    user := ctx.User()
    if user == nil {
        return errors.New("unauthorized")
    }
    
    authManager := pkg.NewAuthManager(ctx.Database(), "jwt-secret", pkg.OAuth2Config{})
    
    // Check if user is admin (can access any post)
    if err := authManager.AuthorizeRole(user, "admin"); err == nil {
        return nil
    }
    
    // Check if user owns the post
    post := getPost(postID)
    if post.AuthorID == user.ID {
        return nil
    }
    
    // Check if user has moderator role
    if err := authManager.AuthorizeRole(user, "moderator"); err == nil {
        return nil
    }
    
    return errors.New("forbidden")
}
```

## CSRF Protection

Cross-Site Request Forgery (CSRF) protection prevents unauthorized commands from being transmitted from a user that the web application trusts.

### Enabling CSRF Protection

```go
// Enable CSRF in configuration
config := pkg.DefaultConfig()
config.Security.EnableCSRF = true
config.Security.CSRFTokenExpiry = 24 * time.Hour

app, _ := pkg.New(config)

// Generate CSRF token for forms
router.GET("/form", func(ctx pkg.Context) error {
    security := ctx.Security()
    
    // Generate CSRF token
    csrfToken, err := security.EnableCSRFProtection(ctx)
    if err != nil {
        return err
    }
    
    // Render form with token
    return ctx.HTML(200, `
        <form method="POST" action="/submit">
            <input type="hidden" name="csrf_token" value="`+csrfToken+`">
            <input type="text" name="data">
            <button type="submit">Submit</button>
        </form>
    `)
})


// Validate CSRF token on form submission
router.POST("/submit", func(ctx pkg.Context) error {
    security := ctx.Security()
    
    // Get token from form
    csrfToken := ctx.FormValue("csrf_token")
    
    // Validate token
    if err := security.ValidateCSRFToken(ctx, csrfToken); err != nil {
        return ctx.JSON(403, map[string]string{"error": "invalid CSRF token"})
    }
    
    // Process form
    return ctx.JSON(200, map[string]string{"message": "success"})
})
```

### CSRF Middleware

Create reusable middleware for CSRF protection:

```go
func csrfMiddleware(ctx pkg.Context) error {
    // Skip CSRF check for GET, HEAD, OPTIONS
    if ctx.Request().Method == "GET" || 
       ctx.Request().Method == "HEAD" || 
       ctx.Request().Method == "OPTIONS" {
        return ctx.Next()
    }
    
    security := ctx.Security()
    
    // Get token from header or form
    token := ctx.GetHeader("X-CSRF-Token")
    if token == "" {
        token = ctx.FormValue("csrf_token")
    }
    
    // Validate token
    if err := security.ValidateCSRFToken(ctx, token); err != nil {
        return ctx.JSON(403, map[string]string{"error": "CSRF validation failed"})
    }
    
    return ctx.Next()
}

// Apply to routes
router.POST("/api/data", dataHandler, csrfMiddleware)
```

**CSRF Features:**
- Automatic token generation and validation
- Secure cookie storage with HttpOnly flag
- Constant-time comparison to prevent timing attacks
- Configurable token expiration
- Support for both form and header tokens

## XSS Protection

Cross-Site Scripting (XSS) protection prevents injection of malicious scripts into web pages.

### Enabling XSS Protection

```go
// Enable XSS protection in configuration
config := pkg.DefaultConfig()
config.Security.EnableXSSProtect = true

app, _ := pkg.New(config)

// XSS protection headers are automatically set
// X-XSS-Protection: 1; mode=block
// Content-Security-Policy: default-src 'self'
```

### Input Sanitization

```go
// Sanitize user input
security := ctx.Security()

userInput := ctx.FormValue("comment")
sanitized := security.SanitizeInput(userInput)

// Sanitization includes:
// - HTML entity escaping
// - Null byte removal
// - Whitespace trimming

// Validate input with rules
rules := pkg.InputValidationRules{
    AllowHTML:      false,
    MaxLength:      1000,
    EscapeHTML:     true,
    TrimWhitespace: true,
}

if err := security.ValidateInput(userInput, rules); err != nil {
    return ctx.JSON(400, map[string]string{"error": "invalid input"})
}
```

### Content Security Policy (CSP)

```go
// Set custom CSP header
router.Use(func(ctx pkg.Context) error {
    ctx.SetHeader("Content-Security-Policy", 
        "default-src 'self'; "+
        "script-src 'self' 'unsafe-inline' https://cdn.example.com; "+
        "style-src 'self' 'unsafe-inline'; "+
        "img-src 'self' data: https:; "+
        "font-src 'self' data:; "+
        "connect-src 'self' https://api.example.com")
    return ctx.Next()
})
```


## Session Security

The framework provides encrypted session management with secure cookie handling.

### Session Configuration

```go
// Configure secure sessions
sessionConfig := pkg.DefaultSessionConfig()
sessionConfig.CookieSecure = true      // HTTPS only
sessionConfig.CookieHTTPOnly = true    // No JavaScript access
sessionConfig.CookieSameSite = "Strict" // CSRF protection
sessionConfig.SessionLifetime = 24 * time.Hour
sessionConfig.EncryptionKey = []byte("your-32-byte-encryption-key-here")

sessionManager, err := pkg.NewSessionManager(sessionConfig, app.Database(), app.Cache())
if err != nil {
    log.Fatal(err)
}

// Create session
router.POST("/login", func(ctx pkg.Context) error {
    // Authenticate user...
    
    // Create new session
    session, err := sessionManager.Create(ctx)
    if err != nil {
        return err
    }
    
    // Store user data in session
    session.Data["user_id"] = user.ID
    session.Data["username"] = user.Username
    
    // Save session
    if err := sessionManager.Save(ctx, session); err != nil {
        return err
    }
    
    // Set encrypted session cookie
    if err := sessionManager.SetCookie(ctx, session); err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "logged in"})
})

// Access session
router.GET("/profile", func(ctx pkg.Context) error {
    // Get session from cookie
    session, err := sessionManager.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "not authenticated"})
    }
    
    // Access session data
    userID := session.Data["user_id"].(string)
    username := session.Data["username"].(string)
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id":  userID,
        "username": username,
    })
})

// Logout
router.POST("/logout", func(ctx pkg.Context) error {
    session, err := sessionManager.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "not authenticated"})
    }
    
    // Destroy session
    if err := sessionManager.Destroy(ctx, session.ID); err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "logged out"})
})
```

### Session Encryption

Sessions are automatically encrypted using AES-256-GCM:

```go
// Session IDs are encrypted in cookies
// Data is encrypted at rest (database/filesystem storage)
// Encryption key must be 32 bytes for AES-256

// Generate encryption key
key := make([]byte, 32)
if _, err := rand.Read(key); err != nil {
    log.Fatal(err)
}

sessionConfig.EncryptionKey = key
```

### Session Storage Options

```go
// Database storage (default)
sessionConfig.StorageType = pkg.SessionStorageDatabase

// Cache storage (Redis, Memcached)
sessionConfig.StorageType = pkg.SessionStorageCache

// Filesystem storage
sessionConfig.StorageType = pkg.SessionStorageFilesystem
sessionConfig.FilesystemPath = "./sessions"
```

**Session Security Features:**
- AES-256-GCM encryption for session IDs and data
- Secure cookie attributes (HttpOnly, Secure, SameSite)
- Automatic session expiration
- Session refresh on activity
- IP address and User-Agent tracking
- Tenant isolation


## Input Validation

Comprehensive input validation prevents injection attacks and ensures data integrity.

### Form Validation

```go
// Define validation rules
rules := pkg.ValidationRules{
    Required: []string{"username", "email", "password"},
    Types: map[string]string{
        "username": "string",
        "email":    "email",
        "age":      "int",
        "website":  "url",
    },
    Lengths: map[string]pkg.LengthRule{
        "username": {Min: 3, Max: 20},
        "password": {Min: 8, Max: 100},
    },
    Patterns: map[string]string{
        "username": "^[a-zA-Z0-9_]+$",
        "phone":    "^\\+?[1-9]\\d{1,14}$",
    },
}

// Validate form data
router.POST("/register", func(ctx pkg.Context) error {
    security := ctx.Security()
    
    if err := security.ValidateFormData(ctx, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Form data is valid
    username := ctx.FormValue("username")
    email := ctx.FormValue("email")
    
    // Process registration...
    return ctx.JSON(200, map[string]string{"message": "registered"})
})
```

### Custom Validators

```go
// Add custom validation logic
rules := pkg.ValidationRules{
    Required: []string{"username"},
    Custom: map[string]pkg.CustomValidator{
        "username": func(value interface{}) error {
            username := value.(string)
            
            // Check if username is already taken
            if isUsernameTaken(username) {
                return errors.New("username already exists")
            }
            
            return nil
        },
    },
}
```

### File Upload Validation

```go
// Define file validation rules
fileRules := pkg.FileValidationRules{
    MaxSize:      10 * 1024 * 1024, // 10 MB
    AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
    AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif"},
    Required:     []string{"avatar"},
    MaxFiles:     5,
}

router.POST("/upload", func(ctx pkg.Context) error {
    security := ctx.Security()
    
    if err := security.ValidateFileUpload(ctx, fileRules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Files are valid
    file := ctx.File("avatar")
    
    // Process upload...
    return ctx.JSON(200, map[string]string{"message": "uploaded"})
})
```

### SQL Injection Prevention

```go
// The framework automatically detects SQL injection patterns
security := ctx.Security()

userInput := ctx.Query("search")

// Validate input
rules := pkg.InputValidationRules{
    MaxLength: 100,
}

if err := security.ValidateInput(userInput, rules); err != nil {
    return ctx.JSON(400, map[string]string{"error": "invalid input"})
}

// Use parameterized queries (always!)
db := ctx.Database()
results, err := db.Query(
    "SELECT * FROM posts WHERE title LIKE ?",
    "%"+userInput+"%",
)
```

**Detected SQL Injection Patterns:**
- `' or '`, `" or "`
- `' or 1=1`, `" or 1=1`
- `; drop`, `; delete`, `; update`, `; insert`
- `exec(`, `execute(`
- `union select`, `union all select`


## Security Headers

The framework automatically sets security headers to protect against common attacks.

### Automatic Security Headers

```go
// Enable automatic security headers
config := pkg.DefaultConfig()
config.Security.EnableXSSProtect = true
config.Security.EnableHSTS = true
config.Security.XFrameOptions = "SAMEORIGIN"

app, _ := pkg.New(config)

// Headers automatically set:
// X-Frame-Options: SAMEORIGIN
// X-XSS-Protection: 1; mode=block
// X-Content-Type-Options: nosniff
// Strict-Transport-Security: max-age=31536000; includeSubDomains
// Referrer-Policy: strict-origin-when-cross-origin
// Permissions-Policy: geolocation=(), microphone=(), camera=()
// Content-Security-Policy: default-src 'self'
```

### HSTS Configuration

HTTP Strict Transport Security (HSTS) forces browsers to use HTTPS.

```go
config.Security.EnableHSTS = true
config.Security.HSTSMaxAge = 31536000 // 1 year in seconds
config.Security.HSTSIncludeSubdomains = true
config.Security.HSTSPreload = false // Enable for preload list submission

// Resulting header:
// Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### CORS Configuration

Cross-Origin Resource Sharing (CORS) controls which domains can access your API.

```go
// Configure CORS
corsConfig := pkg.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    ExposeHeaders:    []string{"X-Total-Count"},
    AllowCredentials: true,
    MaxAge:           3600,
}

// Enable CORS for specific routes
router.OPTIONS("/api/*", func(ctx pkg.Context) error {
    security := ctx.Security()
    if err := security.EnableCORS(ctx, corsConfig); err != nil {
        return err
    }
    return ctx.Status(204)
})

router.GET("/api/data", func(ctx pkg.Context) error {
    security := ctx.Security()
    if err := security.EnableCORS(ctx, corsConfig); err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"data": "value"})
})

// Or use middleware
func corsMiddleware(config pkg.CORSConfig) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        security := ctx.Security()
        if err := security.EnableCORS(ctx, config); err != nil {
            return err
        }
        
        // Handle preflight requests
        if ctx.Request().Method == "OPTIONS" {
            return ctx.Status(204)
        }
        
        return ctx.Next()
    }
}

router.Use(corsMiddleware(corsConfig))
```

**CORS Security Notes:**
- Never use wildcard `*` origin in production with credentials
- Specify exact origins instead of wildcards
- Limit allowed methods to only what's needed
- Be cautious with `AllowCredentials: true`

### Custom Security Headers

```go
// Set custom security headers
router.Use(func(ctx pkg.Context) error {
    // Additional security headers
    ctx.SetHeader("X-Permitted-Cross-Domain-Policies", "none")
    ctx.SetHeader("X-Download-Options", "noopen")
    ctx.SetHeader("X-DNS-Prefetch-Control", "off")
    
    return ctx.Next()
})
```


## Rate Limiting

Rate limiting prevents abuse by limiting the number of requests from a client.

### Resource-Specific Rate Limiting

```go
// Check rate limit for specific resource
router.POST("/api/posts", func(ctx pkg.Context) error {
    security := ctx.Security()
    
    // Check rate limit (default: 100 requests per minute)
    if err := security.CheckRateLimit(ctx, "posts:create"); err != nil {
        return ctx.JSON(429, map[string]string{
            "error": "rate limit exceeded",
            "retry_after": "60",
        })
    }
    
    // Process request...
    return ctx.JSON(200, map[string]string{"message": "created"})
})
```

### Global Rate Limiting

```go
// Check global rate limit
router.Use(func(ctx pkg.Context) error {
    security := ctx.Security()
    
    // Check global rate limit (default: 1000 requests per hour)
    if err := security.CheckGlobalRateLimit(ctx); err != nil {
        return ctx.JSON(429, map[string]string{
            "error": "global rate limit exceeded",
        })
    }
    
    return ctx.Next()
})
```

### Rate Limiting Middleware

```go
func rateLimitMiddleware(resource string) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        security := ctx.Security()
        
        if err := security.CheckRateLimit(ctx, resource); err != nil {
            // Set Retry-After header
            ctx.SetHeader("Retry-After", "60")
            
            return ctx.JSON(429, map[string]interface{}{
                "error": "rate limit exceeded",
                "message": "Too many requests. Please try again later.",
                "retry_after": 60,
            })
        }
        
        return ctx.Next()
    }
}

// Apply to routes
router.POST("/api/login", loginHandler, rateLimitMiddleware("auth:login"))
router.POST("/api/register", registerHandler, rateLimitMiddleware("auth:register"))
```

**Rate Limiting Features:**
- Per-user and per-IP rate limiting
- Resource-specific limits
- Global rate limits
- Automatic cleanup of expired limits
- Database or in-memory storage

## Cookie Encryption

Encrypt sensitive cookie values to prevent tampering.

```go
// Encrypt cookie value
security := ctx.Security()

encryptedValue, err := security.EncryptCookie("sensitive-data")
if err != nil {
    return err
}

// Set encrypted cookie
cookie := &pkg.Cookie{
    Name:     "user_pref",
    Value:    encryptedValue,
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    MaxAge:   3600,
}
ctx.SetCookie(cookie)

// Decrypt cookie value
cookie, err := ctx.GetCookie("user_pref")
if err != nil {
    return err
}

decryptedValue, err := security.DecryptCookie(cookie.Value)
if err != nil {
    return err
}

// Use decrypted value
fmt.Println("User preference:", decryptedValue)
```

**Cookie Encryption Features:**
- AES-256-GCM encryption
- Automatic nonce generation
- Base64 encoding for safe storage
- Tamper detection


## Security Best Practices

### 1. Use HTTPS in Production

Always use HTTPS in production to encrypt data in transit.

```go
// Enable HTTPS
config := pkg.DefaultConfig()
config.Server.TLSEnabled = true
config.Server.TLSCertFile = "/path/to/cert.pem"
config.Server.TLSKeyFile = "/path/to/key.pem"

// Enable HSTS
config.Security.EnableHSTS = true
config.Security.HSTSMaxAge = 31536000
config.Security.HSTSIncludeSubdomains = true

// Force secure cookies
config.Session.CookieSecure = true
```

### 2. Secure Secret Management

Never hardcode secrets in your application code.

```go
// ❌ Bad: Hardcoded secrets
config.Security.JWTSecret = "my-secret-key"
config.Security.EncryptionKey = "0123456789abcdef0123456789abcdef"

// ✅ Good: Load from environment variables
config.Security.JWTSecret = os.Getenv("JWT_SECRET")
encKeyHex := os.Getenv("ENCRYPTION_KEY")
encKey, _ := hex.DecodeString(encKeyHex)
config.Security.EncryptionKey = encKeyHex

// ✅ Better: Use a secrets management service
// AWS Secrets Manager, HashiCorp Vault, etc.
```

### 3. Validate All Input

Always validate and sanitize user input.

```go
// Validate before processing
security := ctx.Security()

userInput := ctx.FormValue("comment")

// Validate
rules := pkg.InputValidationRules{
    AllowHTML:      false,
    MaxLength:      1000,
    EscapeHTML:     true,
    TrimWhitespace: true,
}

if err := security.ValidateInput(userInput, rules); err != nil {
    return ctx.JSON(400, map[string]string{"error": "invalid input"})
}

// Sanitize
sanitized := security.SanitizeInput(userInput)
```

### 4. Use Parameterized Queries

Always use parameterized queries to prevent SQL injection.

```go
// ❌ Bad: String concatenation
query := "SELECT * FROM users WHERE username = '" + username + "'"
db.Query(query)

// ✅ Good: Parameterized query
db.Query("SELECT * FROM users WHERE username = ?", username)
```

### 5. Implement Proper Error Handling

Don't expose sensitive information in error messages.

```go
// ❌ Bad: Exposing internal details
if err != nil {
    return ctx.JSON(500, map[string]string{
        "error": err.Error(), // May expose database structure, file paths, etc.
    })
}

// ✅ Good: Generic error message
if err != nil {
    // Log detailed error internally
    ctx.Logger().Error("Database error", "error", err)
    
    // Return generic message to client
    return ctx.JSON(500, map[string]string{
        "error": "An internal error occurred",
    })
}

// ✅ Better: Use production mode
config.Security.ProductionMode = true // Hides sensitive error details
```

### 6. Set Appropriate Session Timeouts

Configure session timeouts based on your security requirements.

```go
// Short timeout for sensitive operations
sessionConfig.SessionLifetime = 15 * time.Minute // Banking, admin panels

// Medium timeout for general use
sessionConfig.SessionLifetime = 24 * time.Hour // Standard web apps

// Implement idle timeout
router.Use(func(ctx pkg.Context) error {
    session, err := sessionManager.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.Next()
    }
    
    // Check last activity
    lastActivity := session.UpdatedAt
    idleTimeout := 30 * time.Minute
    
    if time.Since(lastActivity) > idleTimeout {
        sessionManager.Destroy(ctx, session.ID)
        return ctx.JSON(401, map[string]string{"error": "session expired"})
    }
    
    // Refresh session
    sessionManager.Refresh(ctx, session.ID)
    return ctx.Next()
})
```


### 7. Implement Account Security

Protect user accounts with proper security measures.

```go
// Password hashing (use bcrypt or similar)
import "golang.org/x/crypto/bcrypt"

// Hash password
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verify password
err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

// Implement account lockout after failed attempts
var failedAttempts int
maxAttempts := 5
lockoutDuration := 15 * time.Minute

if failedAttempts >= maxAttempts {
    return ctx.JSON(429, map[string]string{
        "error": "account locked due to too many failed attempts",
        "retry_after": lockoutDuration.String(),
    })
}

// Implement password strength requirements
func validatePasswordStrength(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
    
    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number, and special character")
    }
    
    return nil
}
```

### 8. Audit and Logging

Log security-relevant events for audit trails.

```go
// Log authentication events
router.POST("/login", func(ctx pkg.Context) error {
    username := ctx.FormValue("username")
    
    user, err := authenticateUser(username, password)
    if err != nil {
        // Log failed login attempt
        ctx.Logger().Warn("Failed login attempt",
            "username", username,
            "ip", ctx.Request().RemoteAddr,
            "user_agent", ctx.GetHeader("User-Agent"),
        )
        return ctx.JSON(401, map[string]string{"error": "invalid credentials"})
    }
    
    // Log successful login
    ctx.Logger().Info("Successful login",
        "user_id", user.ID,
        "username", user.Username,
        "ip", ctx.Request().RemoteAddr,
    )
    
    return ctx.JSON(200, map[string]interface{}{"user": user})
})

// Log authorization failures
func requireRole(role string) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        user := ctx.User()
        
        authManager := pkg.NewAuthManager(ctx.Database(), "jwt-secret", pkg.OAuth2Config{})
        if err := authManager.AuthorizeRole(user, role); err != nil {
            // Log authorization failure
            ctx.Logger().Warn("Authorization failed",
                "user_id", user.ID,
                "required_role", role,
                "user_roles", user.Roles,
                "path", ctx.Request().URL.Path,
            )
            return ctx.JSON(403, map[string]string{"error": "forbidden"})
        }
        
        return ctx.Next()
    }
}
```

### 9. Regular Security Updates

Keep dependencies and the framework up to date.

```bash
# Update dependencies regularly
go get -u ./...
go mod tidy

# Check for known vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### 10. Security Testing

Implement security testing in your development process.

```go
// Test authentication
func TestAuthentication(t *testing.T) {
    // Test with invalid token
    req := httptest.NewRequest("GET", "/api/protected", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    
    resp := performRequest(app, req)
    assert.Equal(t, 401, resp.StatusCode)
    
    // Test with valid token
    token := generateValidToken()
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp = performRequest(app, req)
    assert.Equal(t, 200, resp.StatusCode)
}

// Test authorization
func TestAuthorization(t *testing.T) {
    // Test with insufficient permissions
    user := &pkg.User{Roles: []string{"user"}}
    err := authManager.AuthorizeRole(user, "admin")
    assert.Error(t, err)
    
    // Test with sufficient permissions
    admin := &pkg.User{Roles: []string{"admin"}}
    err = authManager.AuthorizeRole(admin, "admin")
    assert.NoError(t, err)
}

// Test CSRF protection
func TestCSRFProtection(t *testing.T) {
    // Test without CSRF token
    req := httptest.NewRequest("POST", "/submit", nil)
    resp := performRequest(app, req)
    assert.Equal(t, 403, resp.StatusCode)
    
    // Test with valid CSRF token
    token := generateCSRFToken()
    req.Header.Set("X-CSRF-Token", token)
    resp = performRequest(app, req)
    assert.Equal(t, 200, resp.StatusCode)
}
```


## Common Vulnerabilities Prevention

### SQL Injection

**Prevention:**
- Always use parameterized queries
- Never concatenate user input into SQL strings
- Use the framework's built-in SQL injection detection

```go
// ✅ Safe
db.Query("SELECT * FROM users WHERE id = ?", userID)

// ❌ Unsafe
db.Query("SELECT * FROM users WHERE id = " + userID)
```

### Cross-Site Scripting (XSS)

**Prevention:**
- Enable XSS protection headers
- Sanitize user input
- Escape output in templates
- Use Content Security Policy

```go
// Enable XSS protection
config.Security.EnableXSSProtect = true

// Sanitize input
sanitized := security.SanitizeInput(userInput)

// Escape in templates
ctx.HTML(200, "<div>"+html.EscapeString(userInput)+"</div>")
```

### Cross-Site Request Forgery (CSRF)

**Prevention:**
- Enable CSRF protection
- Use CSRF tokens in forms
- Validate tokens on state-changing operations
- Use SameSite cookie attribute

```go
// Enable CSRF
config.Security.EnableCSRF = true

// Generate and validate tokens
csrfToken, _ := security.EnableCSRFProtection(ctx)
err := security.ValidateCSRFToken(ctx, token)
```

### Insecure Direct Object References (IDOR)

**Prevention:**
- Always verify ownership before accessing resources
- Implement proper authorization checks
- Use UUIDs instead of sequential IDs

```go
// Verify ownership
func getPost(ctx pkg.Context) error {
    postID := ctx.Param("id")
    user := ctx.User()
    
    post := loadPost(postID)
    
    // Check ownership or permissions
    if post.AuthorID != user.ID {
        authManager := pkg.NewAuthManager(ctx.Database(), "jwt-secret", pkg.OAuth2Config{})
        if err := authManager.AuthorizeRole(user, "admin"); err != nil {
            return ctx.JSON(403, map[string]string{"error": "forbidden"})
        }
    }
    
    return ctx.JSON(200, post)
}
```

### Session Hijacking

**Prevention:**
- Use encrypted sessions
- Enable secure cookie attributes
- Implement session timeout
- Regenerate session ID after login
- Track IP address and User-Agent

```go
// Secure session configuration
sessionConfig.CookieSecure = true
sessionConfig.CookieHTTPOnly = true
sessionConfig.CookieSameSite = "Strict"
sessionConfig.EncryptionKey = []byte("32-byte-key")

// Regenerate session after login
oldSession, _ := sessionManager.GetSessionFromCookie(ctx)
sessionManager.Destroy(ctx, oldSession.ID)

newSession, _ := sessionManager.Create(ctx)
sessionManager.SetCookie(ctx, newSession)
```

### Brute Force Attacks

**Prevention:**
- Implement rate limiting
- Add account lockout after failed attempts
- Use CAPTCHA for sensitive operations
- Monitor and alert on suspicious activity

```go
// Rate limit login attempts
router.POST("/login", loginHandler, rateLimitMiddleware("auth:login"))

// Account lockout
if failedAttempts >= 5 {
    lockUntil := time.Now().Add(15 * time.Minute)
    return ctx.JSON(429, map[string]interface{}{
        "error": "account locked",
        "retry_after": lockUntil.Unix(),
    })
}
```

### Information Disclosure

**Prevention:**
- Use production mode to hide error details
- Don't expose stack traces
- Remove debug information
- Use generic error messages

```go
// Enable production mode
config.Security.ProductionMode = true

// Generic error handling
if err != nil {
    ctx.Logger().Error("Internal error", "error", err)
    return ctx.JSON(500, map[string]string{
        "error": "An error occurred",
    })
}
```


## Security Configuration Reference

### SecurityConfig

Complete security configuration options:

```go
type SecurityConfig struct {
    // Request validation
    MaxRequestSize   int64         // Maximum request size (default: 10 MB)
    RequestTimeout   time.Duration // Request timeout (default: 30s)
    
    // CSRF protection
    EnableCSRF      bool          // Enable CSRF protection (default: true)
    CSRFTokenExpiry time.Duration // CSRF token expiry (default: 24h)
    
    // XSS protection
    EnableXSSProtect bool   // Enable XSS protection (default: true)
    XFrameOptions    string // X-Frame-Options header (default: "SAMEORIGIN")
    
    // HSTS (HTTP Strict Transport Security)
    EnableHSTS            bool // Enable HSTS (default: true)
    HSTSMaxAge            int  // HSTS max-age in seconds (default: 31536000)
    HSTSIncludeSubdomains bool // Include subdomains (default: true)
    HSTSPreload           bool // Enable preload (default: false)
    
    // Encryption
    EncryptionKey string // AES-256 encryption key (32 bytes, hex-encoded)
    JWTSecret     string // JWT signing secret
    
    // CORS
    AllowedOrigins []string // Allowed CORS origins (default: empty)
    
    // Production mode
    ProductionMode bool // Hide sensitive error details (default: false)
    
    // Input length limits
    MaxHeaderSize     int // Max header size (default: 8 KB)
    MaxURLLength      int // Max URL length (default: 2048)
    MaxFormFieldSize  int // Max form field size (default: 1 MB)
    MaxFormFields     int // Max form fields (default: 1000)
    MaxFileNameLength int // Max filename length (default: 255)
    MaxCookieSize     int // Max cookie size (default: 4 KB)
    MaxQueryParams    int // Max query parameters (default: 100)
}

// Default configuration
config := pkg.DefaultSecurityConfig()
```

### SessionConfig

Complete session configuration options:

```go
type SessionConfig struct {
    // Storage
    StorageType     SessionStorageType // database, cache, or filesystem
    FilesystemPath  string             // Path for filesystem storage
    
    // Cookie settings
    CookieName      string // Cookie name (default: "rockstar_session")
    CookiePath      string // Cookie path (default: "/")
    CookieDomain    string // Cookie domain (default: "")
    CookieSecure    bool   // HTTPS only (default: true)
    CookieHTTPOnly  bool   // No JavaScript access (default: true)
    CookieSameSite  string // SameSite attribute (default: "Lax")
    
    // Session lifetime
    SessionLifetime time.Duration // Session expiry (default: 24h)
    CleanupInterval time.Duration // Cleanup interval (default: 1h)
    
    // Encryption
    EncryptionKey []byte // AES-256 key (32 bytes)
}

// Default configuration
sessionConfig := pkg.DefaultSessionConfig()
```

## API Reference

For detailed API documentation, see:
- [Security API Reference](../api/security.md)
- [Session API Reference](../api/session.md)
- [Context API Reference](../api/context.md)

## Examples

Complete working examples:
- [Secure Application Example](../examples/secure-app.md)
- [Full Featured App Example](../examples/full-featured-app.md)

## Troubleshooting

### Common Issues

**Issue: "Invalid encryption key" error**

Solution: Ensure encryption key is exactly 32 bytes for AES-256:
```go
// Generate key
key := make([]byte, 32)
rand.Read(key)
keyHex := hex.EncodeToString(key)

// Use hex-encoded key
config.Security.EncryptionKey = keyHex
```

**Issue: CSRF validation fails**

Solution: Ensure CSRF token is included in requests:
```go
// In form
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">

// In AJAX
headers: {
    'X-CSRF-Token': csrfToken
}
```

**Issue: Session not persisting**

Solution: Check session configuration:
```go
// Ensure cookie is secure only on HTTPS
sessionConfig.CookieSecure = false // For development

// Check session lifetime
sessionConfig.SessionLifetime = 24 * time.Hour

// Verify encryption key is set
sessionConfig.EncryptionKey = []byte("your-32-byte-key-here")
```

**Issue: Rate limiting too aggressive**

Solution: The default rate limits may need adjustment for your use case. Currently, rate limits are hardcoded (100/min per resource, 1000/hour global). Consider implementing custom rate limiting logic if needed.

## Security Checklist

Before deploying to production:

- [ ] Enable HTTPS with valid TLS certificates
- [ ] Enable HSTS with appropriate max-age
- [ ] Set secure cookie attributes (Secure, HttpOnly, SameSite)
- [ ] Enable CSRF protection for state-changing operations
- [ ] Enable XSS protection headers
- [ ] Configure Content Security Policy
- [ ] Implement rate limiting on sensitive endpoints
- [ ] Use strong encryption keys (32 bytes for AES-256)
- [ ] Use strong JWT secrets (at least 32 characters)
- [ ] Enable production mode to hide error details
- [ ] Implement proper input validation and sanitization
- [ ] Use parameterized queries for all database operations
- [ ] Implement proper authorization checks
- [ ] Set appropriate session timeouts
- [ ] Implement account lockout after failed login attempts
- [ ] Log security-relevant events
- [ ] Regularly update dependencies
- [ ] Perform security testing
- [ ] Review and audit security configurations

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
- [CWE/SANS Top 25](https://www.sans.org/top25-software-errors/)
- [Mozilla Web Security Guidelines](https://infosec.mozilla.org/guidelines/web_security)

## See Also

- [Configuration Guide](configuration.md)
- [Middleware Guide](middleware.md)
- [Database Guide](database.md)
- [Deployment Guide](deployment.md)
