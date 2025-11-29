# Secure Application Example

The Secure Application example demonstrates security best practices and patterns using the Rockstar Web Framework. This guide shows how to implement authentication, authorization, CSRF protection, XSS prevention, and other security features demonstrated in `full_featured_app.go`.

## What This Example Demonstrates

- **Authentication** with JWT tokens
- **Authorization** with role-based access control (RBAC)
- **Session security** with encryption
- **CSRF protection** for state-changing operations
- **XSS prevention** with secure headers
- **Rate limiting** to prevent abuse
- **Secure password handling** with bcrypt
- **Security headers** configuration
- **Input validation** and sanitization

## Prerequisites

- Go 1.25 or higher
- Understanding of security concepts

## Security Configuration

### Comprehensive Security Setup

```go
config := pkg.FrameworkConfig{
    SecurityConfig: pkg.SecurityConfig{
        // Security headers
        XFrameOptions:    "SAMEORIGIN",
        EnableXSSProtect: true,
        EnableCSRF:       true,
        
        // Request limits
        MaxRequestSize:   20 * 1024 * 1024, // 20 MB
        RequestTimeout:   60 * time.Second,
        
        // CORS
        AllowedOrigins:   []string{"https://example.com"},
        
        // Encryption
        EncryptionKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
        JWTSecret:        "your-secret-jwt-key",
        CSRFTokenExpiry:  24 * time.Hour,
    },
    
    SessionConfig: pkg.SessionConfig{
        StorageType:     pkg.SessionStorageDatabase,
        CookieName:      "rockstar_session_id",
        SessionLifetime: 24 * time.Hour,
        CookieSecure:    true,  // Require HTTPS
        CookieHTTPOnly:  true,  // Prevent JavaScript access
        CookieSameSite:  "Strict",
        EncryptionKey:   []byte("..."), // 32 bytes for AES-256
    },
}
```

## Authentication

### JWT Authentication Middleware

```go
func authenticationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Authentication required",
        })
    }
    
    // Remove "Bearer " prefix
    if len(token) > 7 && token[:7] == "Bearer " {
        token = token[7:]
    }
    
    // Validate JWT token
    security := ctx.Security()
    user, err := security.AuthenticateJWT(token)
    if err != nil {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Invalid token",
        })
    }
    
    // Store user in context
    ctx.SetUser(user)
    
    return next(ctx)
}
```

### Login Handler

```go
func loginHandler(ctx pkg.Context) error {
    var credentials struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := ctx.BindJSON(&credentials); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    // Validate credentials
    user, err := validateCredentials(credentials.Email, credentials.Password)
    if err != nil {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Invalid credentials",
        })
    }
    
    // Generate JWT token
    security := ctx.Security()
    token, err := security.GenerateJWT(user.ID, user.Role, 24*time.Hour)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Token generation failed",
        })
    }
    
    // Create session
    session, err := ctx.Session().Create(ctx, user.ID)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Session creation failed",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "token":      token,
        "session_id": session.ID,
        "user":       user,
    })
}
```

### Password Hashing

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func registerHandler(ctx pkg.Context) error {
    var input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Name     string `json:"name"`
    }
    
    if err := ctx.BindJSON(&input); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    // Validate password strength
    if len(input.Password) < 8 {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Password must be at least 8 characters",
        })
    }
    
    // Hash password
    hashedPassword, err := hashPassword(input.Password)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Password hashing failed",
        })
    }
    
    // Create user
    user := &User{
        Email:    input.Email,
        Password: hashedPassword,
        Name:     input.Name,
        Role:     "user",
    }
    
    if err := saveUser(user); err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "User creation failed",
        })
    }
    
    return ctx.JSON(201, map[string]interface{}{
        "message": "User created successfully",
        "user":    user,
    })
}
```

## Authorization

### Role-Based Access Control (RBAC)

```go
func adminAuthorizationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Authentication required",
        })
    }
    
    // Check if user has admin role
    if !ctx.IsAuthorized("admin", "access") {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Insufficient permissions",
        })
    }
    
    return next(ctx)
}

// Apply to admin routes
admin := router.Group("/admin", authenticationMiddleware, adminAuthorizationMiddleware)
admin.GET("/users", listUsersHandler)
admin.POST("/users", createUserHandler)
admin.DELETE("/users/:id", deleteUserHandler)
```

### Permission Checking

```go
type Permission struct {
    Resource string
    Action   string
}

func checkPermission(user *User, resource, action string) bool {
    // Check user role permissions
    permissions := getRolePermissions(user.Role)
    
    for _, perm := range permissions {
        if perm.Resource == resource && perm.Action == action {
            return true
        }
    }
    
    return false
}

func protectedResourceHandler(ctx pkg.Context) error {
    user := ctx.User()
    
    if !checkPermission(user, "documents", "read") {
        return ctx.JSON(403, map[string]interface{}{
            "error": "You don't have permission to read documents",
        })
    }
    
    // User has permission, proceed
    documents := getDocuments()
    return ctx.JSON(200, documents)
}
```

## CSRF Protection

### CSRF Token Generation

```go
func getCSRFToken(ctx pkg.Context) (string, error) {
    security := ctx.Security()
    return security.GenerateCSRFToken(ctx)
}

func csrfTokenHandler(ctx pkg.Context) error {
    token, err := getCSRFToken(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Token generation failed",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "csrf_token": token,
    })
}
```

### CSRF Validation Middleware

```go
func csrfMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Skip CSRF check for safe methods
    if ctx.Request().Method == "GET" || ctx.Request().Method == "HEAD" || ctx.Request().Method == "OPTIONS" {
        return next(ctx)
    }
    
    // Get CSRF token from header or form
    token := ctx.GetHeader("X-CSRF-Token")
    if token == "" {
        token = ctx.FormValue("csrf_token")
    }
    
    if token == "" {
        return ctx.JSON(403, map[string]interface{}{
            "error": "CSRF token required",
        })
    }
    
    // Validate CSRF token
    security := ctx.Security()
    if !security.ValidateCSRFToken(ctx, token) {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Invalid CSRF token",
        })
    }
    
    return next(ctx)
}

// Apply to state-changing routes
router.POST("/api/users", csrfMiddleware, createUserHandler)
router.PUT("/api/users/:id", csrfMiddleware, updateUserHandler)
router.DELETE("/api/users/:id", csrfMiddleware, deleteUserHandler)
```

## XSS Prevention

### Content Security Policy

```go
func securityHeadersMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Content Security Policy
    ctx.SetHeader("Content-Security-Policy", 
        "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
    
    // X-Frame-Options
    ctx.SetHeader("X-Frame-Options", "SAMEORIGIN")
    
    // X-Content-Type-Options
    ctx.SetHeader("X-Content-Type-Options", "nosniff")
    
    // X-XSS-Protection
    ctx.SetHeader("X-XSS-Protection", "1; mode=block")
    
    // Referrer-Policy
    ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
    
    return next(ctx)
}

app.Use(securityHeadersMiddleware)
```

### Input Sanitization

```go
import "html"

func sanitizeInput(input string) string {
    // Escape HTML special characters
    return html.EscapeString(input)
}

func createPostHandler(ctx pkg.Context) error {
    var input struct {
        Title   string `json:"title"`
        Content string `json:"content"`
    }
    
    if err := ctx.BindJSON(&input); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    // Sanitize user input
    post := &Post{
        Title:   sanitizeInput(input.Title),
        Content: sanitizeInput(input.Content),
    }
    
    if err := savePost(post); err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Post creation failed",
        })
    }
    
    return ctx.JSON(201, post)
}
```

## Rate Limiting

### IP-Based Rate Limiting

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        HostConfigs: map[string]*pkg.HostConfig{
            "api.example.com": {
                RateLimits: &pkg.RateLimitConfig{
                    Enabled:           true,
                    RequestsPerSecond: 100,
                    BurstSize:         20,
                    Storage:           "database",
                },
            },
        },
    },
}
```

### User-Based Rate Limiting

```go
func userRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    user := ctx.User()
    if user == nil {
        return next(ctx)
    }
    
    // Check user-specific rate limit
    limiter := getUserRateLimiter(user.ID)
    if !limiter.Allow() {
        return ctx.JSON(429, map[string]interface{}{
            "error": "Rate limit exceeded",
            "retry_after": limiter.RetryAfter(),
        })
    }
    
    return next(ctx)
}
```

## Session Security

### Secure Session Configuration

```go
SessionConfig: pkg.SessionConfig{
    StorageType:     pkg.SessionStorageDatabase,
    CookieName:      "rockstar_session_id",
    SessionLifetime: 24 * time.Hour,
    CookieSecure:    true,  // HTTPS only
    CookieHTTPOnly:  true,  // No JavaScript access
    CookieSameSite:  "Strict",
    EncryptionKey:   []byte("..."), // 32 bytes
    CleanupInterval: 10 * time.Minute,
}
```

### Session Validation

```go
func validateSession(ctx pkg.Context) (*Session, error) {
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return nil, err
    }
    
    // Check if session is expired
    if session.ExpiresAt.Before(time.Now()) {
        ctx.Session().Destroy(ctx, session.ID)
        return nil, fmt.Errorf("session expired")
    }
    
    // Refresh session
    session.ExpiresAt = time.Now().Add(24 * time.Hour)
    ctx.Session().Save(ctx, session)
    
    return session, nil
}
```

## Input Validation

### Request Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

func validateRequest(input interface{}) error {
    validate := validator.New()
    return validate.Struct(input)
}

func createUserHandler(ctx pkg.Context) error {
    var input CreateUserRequest
    
    if err := ctx.BindJSON(&input); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid JSON",
        })
    }
    
    if err := validateRequest(input); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Validation failed",
            "details": err.Error(),
        })
    }
    
    // Input is valid, proceed
    // ...
}
```

## Security Best Practices

### 1. Use HTTPS in Production

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        TLSCertFile: "/path/to/cert.pem",
        TLSKeyFile:  "/path/to/key.pem",
    },
}
```

### 2. Store Secrets Securely

```go
// Use environment variables
jwtSecret := os.Getenv("JWT_SECRET")
encryptionKey := os.Getenv("ENCRYPTION_KEY")

// Or use a secrets manager
secrets := loadSecretsFromVault()
```

### 3. Implement Audit Logging

```go
func auditLog(ctx pkg.Context, action string, resource string) {
    user := ctx.User()
    
    log := AuditLog{
        UserID:    user.ID,
        Action:    action,
        Resource:  resource,
        IP:        ctx.Request().RemoteAddr,
        Timestamp: time.Now(),
    }
    
    saveAuditLog(log)
}

func deleteUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // Perform deletion
    if err := deleteUser(userID); err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Deletion failed",
        })
    }
    
    // Log the action
    auditLog(ctx, "delete", fmt.Sprintf("user:%s", userID))
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "User deleted",
    })
}
```

### 4. Implement Account Lockout

```go
func loginHandler(ctx pkg.Context) error {
    var credentials struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    
    if err := ctx.BindJSON(&credentials); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    // Check if account is locked
    if isAccountLocked(credentials.Email) {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Account locked due to too many failed attempts",
        })
    }
    
    // Validate credentials
    user, err := validateCredentials(credentials.Email, credentials.Password)
    if err != nil {
        // Increment failed attempts
        incrementFailedAttempts(credentials.Email)
        
        return ctx.JSON(401, map[string]interface{}{
            "error": "Invalid credentials",
        })
    }
    
    // Reset failed attempts on successful login
    resetFailedAttempts(credentials.Email)
    
    // Generate token and return
    // ...
}
```

## Common Security Issues

### SQL Injection

**Bad**:
```go
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
```

**Good**:
```go
query := "SELECT * FROM users WHERE email = ?"
db.Query(query, email)
```

### XSS

**Bad**:
```go
html := fmt.Sprintf("<div>%s</div>", userInput)
```

**Good**:
```go
html := fmt.Sprintf("<div>%s</div>", html.EscapeString(userInput))
```

### Insecure Direct Object Reference

**Bad**:
```go
// No authorization check
userID := ctx.Params()["id"]
user := getUser(userID)
```

**Good**:
```go
// Check authorization
userID := ctx.Params()["id"]
currentUser := ctx.User()
if currentUser.ID != userID && !currentUser.IsAdmin {
    return ctx.JSON(403, map[string]interface{}{
        "error": "Unauthorized",
    })
}
user := getUser(userID)
```

## Related Documentation

- [Security Guide](../guides/security.md) - Security best practices
- [Session Guide](../guides/sessions.md) - Session management
- [Configuration Guide](../guides/configuration.md) - Security configuration
- [Full Featured App](full-featured-app.md) - Complete example

## Source Code

Security examples are available in `examples/full_featured_app.go` in the repository.
