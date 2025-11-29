---
title: "Common Errors"
description: "Common error messages and their solutions"
category: "troubleshooting"
tags: ["troubleshooting", "errors", "solutions"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "debugging.md"
  - "faq.md"
---

# Common Errors

This guide documents common error messages you may encounter when using the Rockstar Web Framework, along with their causes and solutions.

## Table of Contents

- [Configuration Errors](#configuration-errors)
- [Database Errors](#database-errors)
- [Authentication Errors](#authentication-errors)
- [Authorization Errors](#authorization-errors)
- [Validation Errors](#validation-errors)
- [Security Errors](#security-errors)
- [Session Errors](#session-errors)
- [Routing Errors](#routing-errors)
- [Plugin Errors](#plugin-errors)
- [Protocol Errors](#protocol-errors)
- [Multi-Tenancy Errors](#multi-tenancy-errors)
- [WebSocket Errors](#websocket-errors)
- [Request Errors](#request-errors)
- [Server Errors](#server-errors)

## Configuration Errors

### CONFIGURATION_ERROR

**Error Message:**
```
configuration error for '{key}': {reason}
```

**Causes:**
- Invalid configuration value
- Missing required configuration field
- Configuration file syntax error
- Type mismatch in configuration

**Solutions:**

1. **Validate your configuration:**
```go
config := pkg.FrameworkConfig{
    // ... your configuration
}

// Validate before use
if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

2. **Check configuration file syntax:**
```bash
# For YAML files
yamllint config.yaml

# For JSON files
jq . config.json
```

3. **Use configuration defaults:**
```go
config := pkg.FrameworkConfig{}
config.ApplyDefaults()
```

4. **Enable debug logging to see configuration loading:**
```go
config := pkg.FrameworkConfig{
    LogLevel: "debug",
}
```

**See Also:**
- [Configuration Guide](../guides/configuration.md)
- [FAQ: Configuration](faq.md#configuration)

### MISSING_CONFIGURATION

**Error Message:**
```
required configuration '{key}' is missing
```

**Causes:**
- Required configuration field not provided
- Environment variable not set
- Configuration file not found

**Solutions:**

1. **Provide required configuration:**
```go
config := pkg.FrameworkConfig{
    Port: 8080,  // Required
    Host: "0.0.0.0",  // Required
}
```

2. **Set environment variables:**
```bash
export ROCKSTAR_PORT=8080
export ROCKSTAR_HOST=0.0.0.0
```

3. **Check configuration file path:**
```go
config, err := pkg.LoadConfig("config.yaml")
if err != nil {
    log.Fatal("Failed to load config:", err)
}
```

## Database Errors

### NO_DATABASE_CONFIGURED

**Error Message:**
```
No database is configured. To use database features, provide a DatabaseConfig with Driver, Database, and credentials.
```

**Causes:**
- Attempting to use database features without configuring a database
- Database configuration is nil or empty

**Solutions:**

1. **Configure a database:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "app.db",
    },
}
```

2. **For production, use environment variables:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        Driver:   os.Getenv("DB_DRIVER"),
        Host:     os.Getenv("DB_HOST"),
        Port:     5432,
        Database: os.Getenv("DB_NAME"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
    },
}
```

3. **Check if database is optional for your use case:**
```go
// If database is optional, check before using
if app.Database() != nil {
    db := app.Database()
    // Use database
}
```

**See Also:**
- [Database Guide](../guides/database.md)
- [Database API Reference](../api/database.md)

### DATABASE_CONNECTION

**Error Message:**
```
failed to connect to database: {details}
```

**Causes:**
- Database server not running
- Incorrect connection credentials
- Network connectivity issues
- Firewall blocking connection
- Wrong host or port

**Solutions:**

1. **Verify database server is running:**
```bash
# For PostgreSQL
pg_isready -h localhost -p 5432

# For MySQL
mysqladmin ping -h localhost

# For SQLite (check file exists)
ls -l app.db
```

2. **Test connection manually:**
```bash
# PostgreSQL
psql -h localhost -U myuser -d mydb

# MySQL
mysql -h localhost -u myuser -p mydb
```

3. **Check connection string:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        Driver:   "postgres",
        Host:     "localhost",  // Verify host
        Port:     5432,         // Verify port
        Database: "mydb",       // Verify database name
        User:     "myuser",     // Verify username
        Password: "mypass",     // Verify password
    },
}
```

4. **Enable connection logging:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        // ... connection details
        LogQueries: true,
    },
    LogLevel: "debug",
}
```

5. **Check firewall rules:**
```bash
# Linux
sudo ufw status
sudo iptables -L

# Check if port is open
telnet localhost 5432
```

### DATABASE_QUERY

**Error Message:**
```
database query failed: {details}
```

**Causes:**
- SQL syntax error
- Table or column doesn't exist
- Constraint violation
- Connection lost during query
- Query timeout

**Solutions:**

1. **Check SQL syntax:**
```go
// Use parameterized queries
result, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
if err != nil {
    log.Printf("Query error: %v", err)
}
```

2. **Verify table exists:**
```go
// Check if table exists
exists, err := db.TableExists("users")
if !exists {
    // Run migrations
}
```

3. **Handle constraint violations:**
```go
result, err := db.Exec("INSERT INTO users (email) VALUES (?)", email)
if err != nil {
    if strings.Contains(err.Error(), "UNIQUE constraint") {
        return pkg.NewDatabaseError("Email already exists", "insert")
    }
    return err
}
```

4. **Set query timeout:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        // ... connection details
        QueryTimeout: 30 * time.Second,
    },
}
```

### RECORD_NOT_FOUND

**Error Message:**
```
record not found
```

**Causes:**
- Querying for non-existent record
- Record was deleted
- Wrong query parameters

**Solutions:**

1. **Handle not found gracefully:**
```go
user, err := db.FindUserByID(userID)
if err != nil {
    if pkg.IsFrameworkError(err) {
        if fe, _ := pkg.GetFrameworkError(err); fe.Code == pkg.ErrCodeRecordNotFound {
            return ctx.JSON(404, map[string]string{"error": "User not found"})
        }
    }
    return err
}
```

2. **Use existence checks:**
```go
exists, err := db.Exists("users", "id = ?", userID)
if !exists {
    return pkg.NewNotFoundError("user")
}
```

### DUPLICATE_RECORD

**Error Message:**
```
duplicate record: unique constraint violation
```

**Causes:**
- Attempting to insert duplicate unique value
- Race condition in concurrent inserts

**Solutions:**

1. **Check before insert:**
```go
exists, err := db.Exists("users", "email = ?", email)
if exists {
    return pkg.NewValidationError("Email already exists", "email")
}
```

2. **Use upsert operations:**
```go
// Insert or update
err := db.Upsert("users", user, "email")
```

3. **Handle duplicate errors:**
```go
err := db.Insert("users", user)
if err != nil {
    if strings.Contains(err.Error(), "duplicate") {
        return pkg.NewValidationError("Record already exists", "")
    }
    return err
}
```

## Authentication Errors

### AUTH_FAILED

**Error Message:**
```
authentication failed: {reason}
```

**Causes:**
- Invalid credentials
- User not found
- Password mismatch
- Account locked or disabled

**Solutions:**

1. **Verify credentials:**
```go
user, err := security.AuthenticateUser(username, password)
if err != nil {
    return pkg.NewAuthenticationError("Invalid username or password")
}
```

2. **Check account status:**
```go
if !user.IsActive {
    return pkg.NewAuthenticationError("Account is disabled")
}
```

3. **Implement rate limiting:**
```go
// Prevent brute force attacks
if security.IsRateLimited(username) {
    return pkg.NewRateLimitError("Too many login attempts", 5, "15 minutes")
}
```

### INVALID_TOKEN

**Error Message:**
```
invalid or malformed token
```

**Causes:**
- Token format is incorrect
- Token signature invalid
- Token has been tampered with

**Solutions:**

1. **Verify token format:**
```go
token := ctx.Header("Authorization")
if !strings.HasPrefix(token, "Bearer ") {
    return pkg.NewAuthenticationError("Invalid token format")
}
token = strings.TrimPrefix(token, "Bearer ")
```

2. **Validate token:**
```go
claims, err := security.ValidateJWT(token)
if err != nil {
    return pkg.NewAuthenticationError("Invalid token")
}
```

3. **Check token signing key:**
```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        JWTSecret: os.Getenv("JWT_SECRET"),  // Must match signing key
    },
}
```

### TOKEN_EXPIRED

**Error Message:**
```
token has expired
```

**Causes:**
- Token lifetime exceeded
- System clock skew

**Solutions:**

1. **Implement token refresh:**
```go
if err == pkg.ErrCodeTokenExpired {
    // Redirect to refresh token endpoint
    return ctx.Redirect(302, "/auth/refresh")
}
```

2. **Adjust token lifetime:**
```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        JWTExpiration: 24 * time.Hour,  // Adjust as needed
    },
}
```

3. **Handle gracefully in client:**
```javascript
// Client-side handling
if (error.code === 'TOKEN_EXPIRED') {
    // Refresh token or redirect to login
    refreshToken();
}
```

## Authorization Errors

### FORBIDDEN

**Error Message:**
```
access forbidden: insufficient permissions
```

**Causes:**
- User lacks required role
- User lacks required permission
- Resource access denied

**Solutions:**

1. **Check user roles:**
```go
if !security.HasRole(ctx, "admin") {
    return pkg.NewAuthorizationError("Admin role required")
}
```

2. **Check permissions:**
```go
if !security.HasPermission(ctx, "users:write") {
    return pkg.NewAuthorizationError("Write permission required")
}
```

3. **Implement RBAC properly:**
```go
// Define roles and permissions
security.DefineRole("admin", []string{
    "users:read",
    "users:write",
    "users:delete",
})
```

### INSUFFICIENT_ROLES

**Error Message:**
```
user does not have required roles: {roles}
```

**Causes:**
- User not assigned required role
- Role hierarchy not configured

**Solutions:**

1. **Assign roles to user:**
```go
err := security.AssignRole(userID, "editor")
```

2. **Use role middleware:**
```go
router.GET("/admin", handler, pkg.RequireRole("admin"))
```

3. **Check multiple roles:**
```go
if !security.HasAnyRole(ctx, "admin", "moderator") {
    return pkg.NewAuthorizationError("Admin or moderator role required")
}
```

## Validation Errors

### VALIDATION_FAILED

**Error Message:**
```
validation failed: {details}
```

**Causes:**
- Input doesn't meet validation rules
- Required field missing
- Invalid format
- Value out of range

**Solutions:**

1. **Implement input validation:**
```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
}

var req CreateUserRequest
if err := ctx.BindJSON(&req); err != nil {
    return pkg.NewValidationError("Invalid request body", "")
}

if err := ctx.Validate(&req); err != nil {
    return err
}
```

2. **Provide clear error messages:**
```go
if len(password) < 8 {
    return pkg.NewValidationError("Password must be at least 8 characters", "password")
}
```

3. **Validate before processing:**
```go
// Validate email format
if !isValidEmail(email) {
    return pkg.NewInvalidFormatError("email", "valid email address")
}
```

### MISSING_FIELD

**Error Message:**
```
required field '{field}' is missing
```

**Causes:**
- Required field not provided in request
- Field is null or empty

**Solutions:**

1. **Check required fields:**
```go
if email == "" {
    return pkg.NewMissingFieldError("email")
}
```

2. **Use struct tags:**
```go
type User struct {
    Email string `json:"email" binding:"required"`
    Name  string `json:"name" binding:"required"`
}
```

### INVALID_FORMAT

**Error Message:**
```
field '{field}' has invalid format, expected: {format}
```

**Causes:**
- Value doesn't match expected format
- Invalid date/time format
- Invalid email/URL format

**Solutions:**

1. **Validate formats:**
```go
// Email validation
emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
if !emailRegex.MatchString(email) {
    return pkg.NewInvalidFormatError("email", "valid email address")
}

// Date validation
_, err := time.Parse("2006-01-02", dateStr)
if err != nil {
    return pkg.NewInvalidFormatError("date", "YYYY-MM-DD")
}
```

## Security Errors

### CSRF_TOKEN_INVALID

**Error Message:**
```
CSRF token is invalid or missing
```

**Causes:**
- CSRF token not included in request
- Token doesn't match session
- Token expired

**Solutions:**

1. **Include CSRF token in forms:**
```html
<form method="POST">
    <input type="hidden" name="_csrf" value="{{ .CSRFToken }}">
    <!-- form fields -->
</form>
```

2. **Include in AJAX requests:**
```javascript
fetch('/api/endpoint', {
    method: 'POST',
    headers: {
        'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]').content
    },
    body: JSON.stringify(data)
});
```

3. **Configure CSRF protection:**
```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        EnableCSRF: true,
        CSRFTokenLength: 32,
    },
}
```

### XSS_DETECTED

**Error Message:**
```
potential XSS attack detected in input
```

**Causes:**
- Input contains script tags or JavaScript
- Malicious HTML in user input

**Solutions:**

1. **Sanitize user input:**
```go
import "html"

sanitized := html.EscapeString(userInput)
```

2. **Use template auto-escaping:**
```go
// Templates automatically escape by default
tmpl.Execute(w, data)
```

3. **Validate input:**
```go
if security.ContainsXSS(input) {
    return pkg.NewValidationError("Input contains invalid characters", "content")
}
```

### RATE_LIMIT_EXCEEDED

**Error Message:**
```
rate limit exceeded: {limit} requests per {window}
```

**Causes:**
- Too many requests from same IP/user
- Brute force attack attempt
- Misconfigured rate limits

**Solutions:**

1. **Wait and retry:**
```go
// Client-side: implement exponential backoff
time.Sleep(time.Second * time.Duration(math.Pow(2, retryCount)))
```

2. **Adjust rate limits:**
```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        RateLimit: &pkg.RateLimitConfig{
            RequestsPerMinute: 100,  // Adjust as needed
            BurstSize: 20,
        },
    },
}
```

3. **Use authentication:**
```go
// Authenticated users may have higher limits
if ctx.IsAuthenticated() {
    // Higher limit for authenticated users
}
```

## Session Errors

### SESSION_NOT_FOUND

**Error Message:**
```
session not found
```

**Causes:**
- Session expired
- Session ID invalid
- Session storage cleared

**Solutions:**

1. **Handle missing sessions:**
```go
session, err := ctx.Session()
if err != nil {
    // Redirect to login
    return ctx.Redirect(302, "/login")
}
```

2. **Extend session lifetime:**
```go
config := pkg.FrameworkConfig{
    Session: &pkg.SessionConfig{
        MaxAge: 24 * time.Hour,  // Adjust as needed
    },
}
```

### SESSION_EXPIRED

**Error Message:**
```
session has expired
```

**Causes:**
- Session lifetime exceeded
- Idle timeout reached

**Solutions:**

1. **Implement session refresh:**
```go
// Refresh session on activity
session.Touch()
```

2. **Configure session timeouts:**
```go
config := pkg.FrameworkConfig{
    Session: &pkg.SessionConfig{
        MaxAge: 24 * time.Hour,
        IdleTimeout: 30 * time.Minute,
    },
}
```

## Routing Errors

### NOT_FOUND

**Error Message:**
```
{resource} not found
```

**Causes:**
- Route not registered
- Wrong HTTP method
- Typo in URL path

**Solutions:**

1. **Verify route registration:**
```go
router := app.Router()
router.GET("/api/users/:id", getUserHandler)
```

2. **Check route patterns:**
```go
// List all registered routes
routes := router.Routes()
for _, route := range routes {
    log.Printf("%s %s", route.Method, route.Path)
}
```

3. **Implement custom 404 handler:**
```go
router.NotFound(func(ctx pkg.Context) error {
    return ctx.JSON(404, map[string]string{
        "error": "Route not found",
    })
})
```

### METHOD_NOT_ALLOWED

**Error Message:**
```
method {method} not allowed
```

**Causes:**
- Using wrong HTTP method
- Route registered for different method

**Solutions:**

1. **Check HTTP method:**
```go
// Ensure you're using the correct method
router.POST("/api/users", createUserHandler)  // Not GET
```

2. **Support multiple methods:**
```go
router.Match([]string{"GET", "POST"}, "/api/users", handler)
```

## Plugin Errors

### Plugin Loading Failures

**Error Message:**
```
failed to load plugin: {details}
```

**Causes:**
- Plugin file not found
- Plugin compiled with wrong Go version
- Missing plugin dependencies
- Invalid plugin manifest

**Solutions:**

1. **Verify plugin file exists:**
```bash
ls -l plugins/my-plugin.so
```

2. **Check plugin manifest:**
```yaml
# plugin.yaml
name: my-plugin
version: 1.0.0
dependencies: []
```

3. **Rebuild plugin with correct Go version:**
```bash
go build -buildmode=plugin -o my-plugin.so plugin.go
```

4. **Check plugin permissions:**
```go
config := pkg.FrameworkConfig{
    Plugins: &pkg.PluginConfig{
        Directory: "./plugins",
        Enabled: true,
    },
}
```

**See Also:**
- [Plugin Development Guide](../guides/plugins.md)
- [Plugin API Reference](../api/plugins.md)

## Protocol Errors

### HTTP/2 Issues

**Error Message:**
```
HTTP/2 connection failed
```

**Causes:**
- TLS not configured
- Client doesn't support HTTP/2
- ALPN negotiation failed

**Solutions:**

1. **Configure TLS:**
```go
config := pkg.FrameworkConfig{
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile:  "key.pem",
    },
    EnableHTTP2: true,
}
```

2. **Generate certificates:**
```bash
# For development
go run examples/generate_keys.go

# For production, use Let's Encrypt
certbot certonly --standalone -d yourdomain.com
```

### QUIC Connection Failures

**Error Message:**
```
QUIC connection failed
```

**Causes:**
- UDP port blocked
- Firewall rules
- QUIC not supported by client

**Solutions:**

1. **Check UDP port:**
```bash
# Ensure UDP port is open
sudo ufw allow 443/udp
```

2. **Configure QUIC:**
```go
config := pkg.FrameworkConfig{
    EnableQUIC: true,
    QUICPort: 443,
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile:  "key.pem",
    },
}
```

## Multi-Tenancy Errors

### TENANT_NOT_FOUND

**Error Message:**
```
tenant not found for host: {host}
```

**Causes:**
- Tenant not configured for hostname
- DNS not pointing to application
- Tenant disabled

**Solutions:**

1. **Register tenant:**
```go
tenant := &pkg.Tenant{
    ID:     "tenant1",
    Name:   "Tenant 1",
    Host:   "tenant1.example.com",
    Active: true,
}
err := app.RegisterTenant(tenant)
```

2. **Configure wildcard tenant:**
```go
config := pkg.FrameworkConfig{
    MultiTenancy: &pkg.MultiTenancyConfig{
        Enabled: true,
        DefaultTenant: "default",
    },
}
```

### TENANT_INACTIVE

**Error Message:**
```
tenant is inactive
```

**Causes:**
- Tenant disabled in configuration
- Subscription expired
- Maintenance mode

**Solutions:**

1. **Activate tenant:**
```go
err := app.ActivateTenant(tenantID)
```

2. **Check tenant status:**
```go
tenant, err := app.GetTenant(tenantID)
if !tenant.Active {
    return pkg.NewTenantError(pkg.ErrCodeTenantInactive, "Tenant is inactive", 503)
}
```

## WebSocket Errors

### WEBSOCKET_UPGRADE_FAILED

**Error Message:**
```
WebSocket upgrade failed
```

**Causes:**
- Missing Upgrade header
- Wrong protocol version
- Origin check failed

**Solutions:**

1. **Configure WebSocket properly:**
```go
router.WebSocket("/ws", func(ctx pkg.Context) error {
    conn, err := ctx.UpgradeWebSocket()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Handle WebSocket connection
    return nil
})
```

2. **Check client request:**
```javascript
// Client-side
const ws = new WebSocket('ws://localhost:8080/ws');
```

3. **Configure CORS for WebSocket:**
```go
config := pkg.FrameworkConfig{
    CORS: &pkg.CORSConfig{
        AllowOrigins: []string{"http://localhost:3000"},
        AllowWebSockets: true,
    },
}
```

## Request Errors

### REQUEST_TOO_LARGE

**Error Message:**
```
request size exceeds maximum allowed: {max_size}
```

**Causes:**
- Request body too large
- File upload too large

**Solutions:**

1. **Adjust max request size:**
```go
config := pkg.FrameworkConfig{
    MaxRequestSize: 10 * 1024 * 1024,  // 10MB
}
```

2. **Handle large files differently:**
```go
// Stream large files instead of loading into memory
file, err := ctx.FormFile("upload")
if err != nil {
    return err
}
defer file.Close()

// Stream to storage
```

### REQUEST_TIMEOUT

**Error Message:**
```
request timeout exceeded: {timeout}
```

**Causes:**
- Request taking too long
- Slow database query
- External API timeout

**Solutions:**

1. **Adjust timeout:**
```go
config := pkg.FrameworkConfig{
    RequestTimeout: 60 * time.Second,
}
```

2. **Optimize slow operations:**
```go
// Use context with timeout
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

result, err := db.QueryContext(ctx, query)
```

## Server Errors

### INTERNAL_ERROR

**Error Message:**
```
internal server error
```

**Causes:**
- Unhandled exception
- Panic in handler
- Unexpected error condition

**Solutions:**

1. **Enable error logging:**
```go
config := pkg.FrameworkConfig{
    LogLevel: "debug",
    LogErrors: true,
}
```

2. **Implement error recovery:**
```go
app.Use(pkg.RecoveryMiddleware(errorHandler))
```

3. **Check application logs:**
```bash
tail -f app.log
```

### SERVICE_UNAVAILABLE

**Error Message:**
```
service temporarily unavailable
```

**Causes:**
- Server overloaded
- Maintenance mode
- Dependency unavailable

**Solutions:**

1. **Implement health checks:**
```go
router.GET("/health", func(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

2. **Enable graceful shutdown:**
```go
// Handle shutdown signals
app.GracefulShutdown(30 * time.Second)
```

## Getting More Help

If you encounter an error not listed here:

1. **Enable debug logging:**
```go
config := pkg.FrameworkConfig{
    LogLevel: "debug",
    LogFormat: "json",
}
```

2. **Check the logs for details:**
```bash
tail -f app.log | jq .
```

3. **Search GitHub issues:**
- [Existing Issues](https://github.com/echterhof/rockstar-web-framework/issues)

4. **Create a new issue with:**
- Complete error message
- Stack trace
- Minimal reproduction example
- Framework and Go versions

## Navigation

- [← Back to Troubleshooting](README.md)
- [Next: Debugging →](debugging.md)
- [FAQ →](faq.md)
