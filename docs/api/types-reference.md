# Types and Data Models Reference

Complete reference for all data types, structs, and models in the Rockstar Web Framework.

## Overview

This document provides detailed documentation for all public types, structs, and data models used throughout the framework. These types are used for configuration, data persistence, and inter-component communication.

## Table of Contents

- [Core Types](#core-types)
- [Configuration Types](#configuration-types)
- [Data Models](#data-models)
- [Request/Response Types](#requestresponse-types)
- [Security Types](#security-types)
- [Monitoring Types](#monitoring-types)
- [Plugin Types](#plugin-types)

---

## Core Types

### HandlerFunc

```go
type HandlerFunc func(ctx Context) error
```

The primary handler function type for route handlers.

**Parameters:**
- `ctx`: Request context with access to all framework features

**Returns:**
- `error`: Error if handler fails (automatically converted to HTTP error response)

**Example:**
```go
func myHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{
        "message": "Success",
    })
}

app.Router().GET("/", myHandler)
```

### MiddlewareFunc

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

Middleware function type for request processing.

**Parameters:**
- `ctx`: Request context
- `next`: Next handler in the chain

**Returns:**
- `error`: Error if middleware fails

**Example:**
```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return pkg.NewAuthenticationError("Missing token")
    }
    
    // Validate token...
    
    return next(ctx)
}
```

### ValidationRule

```go
type ValidationRule interface {
    Validate(value interface{}) error
    ErrorMessage() string
}
```

Interface for form validation rules.

**Methods:**
- `Validate(value)`: Validates a value, returns error if invalid
- `ErrorMessage()`: Returns human-readable error message

**Example:**
```go
type MinLengthRule struct {
    Min int
}

func (r MinLengthRule) Validate(value interface{}) error {
    str, ok := value.(string)
    if !ok || len(str) < r.Min {
        return fmt.Errorf("minimum length is %d", r.Min)
    }
    return nil
}

func (r MinLengthRule) ErrorMessage() string {
    return fmt.Sprintf("must be at least %d characters", r.Min)
}
```

---

## Configuration Types

### FrameworkConfig

```go
type FrameworkConfig struct {
    ServerConfig       ServerConfig
    DatabaseConfig     DatabaseConfig
    CacheConfig        CacheConfig
    SessionConfig      SessionConfig
    ConfigFiles        []string
    I18nConfig         I18nConfig
    SecurityConfig     SecurityConfig
    MonitoringConfig   MonitoringConfig
    ProxyConfig        ProxyConfig
    PluginConfigPath   string
    EnablePlugins      bool
    FileSystemRoot     string
}
```

Complete framework configuration.

**Fields:**
- `ServerConfig`: HTTP server configuration
- `DatabaseConfig`: Database connection settings
- `CacheConfig`: Caching system configuration
- `SessionConfig`: Session management settings
- `ConfigFiles`: Paths to YAML config files to load
- `I18nConfig`: Internationalization settings
- `SecurityConfig`: Security and authentication settings
- `MonitoringConfig`: Monitoring and profiling settings
- `ProxyConfig`: Proxy and load balancing settings
- `PluginConfigPath`: Path to plugin configuration file
- `EnablePlugins`: Whether to enable the plugin system
- `FileSystemRoot`: Root directory for file operations

**Example:**
```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        Address: ":8080",
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver: "sqlite",
        Database: "app.db",
    },
    EnablePlugins: true,
}

app, err := pkg.New(config)
```

### ServerConfig

```go
type ServerConfig struct {
    Address            string
    ReadTimeout        time.Duration
    WriteTimeout       time.Duration
    IdleTimeout        time.Duration
    MaxHeaderBytes     int
    EnableHTTP2        bool
    EnableQUIC         bool
    TLSCertFile        string
    TLSKeyFile         string
    EnablePrefork      bool
    PreforkWorkers     int
    EnableKeepAlive    bool
    KeepAlivePeriod    time.Duration
    ReadBufferSize     int
    WriteBufferSize    int
}
```

HTTP server configuration with defaults.

**Defaults:**
- `ReadTimeout`: 10 seconds
- `WriteTimeout`: 10 seconds
- `IdleTimeout`: 120 seconds
- `MaxHeaderBytes`: 1MB
- `EnableHTTP2`: false
- `EnableQUIC`: false
- `EnablePrefork`: false (Unix only)
- `PreforkWorkers`: runtime.NumCPU()
- `EnableKeepAlive`: true
- `KeepAlivePeriod`: 3 minutes
- `ReadBufferSize`: 4096 bytes
- `WriteBufferSize`: 4096 bytes

**Example:**
```go
config := pkg.ServerConfig{
    Address: ":8080",
    ReadTimeout: 15 * time.Second,
    WriteTimeout: 15 * time.Second,
    EnableHTTP2: true,
    EnablePrefork: true,
    PreforkWorkers: 4,
}
```

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver          string
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    SSLMode         string
    Charset         string
    Timezone        string
    ConnMaxLifetime time.Duration
    MaxOpenConns    int
    MaxIdleConns    int
    Options         map[string]string
}
```

Database connection configuration.

**Defaults:**
- `Host`: "localhost"
- `Port`: Driver-specific (postgres=5432, mysql=3306, mssql=1433, sqlite=0)
- `ConnMaxLifetime`: 5 minutes
- `MaxOpenConns`: 25
- `MaxIdleConns`: 5

**Example:**
```go
config := pkg.DatabaseConfig{
    Driver: "postgres",
    Host: "localhost",
    Port: 5432,
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    SSLMode: "disable",
    MaxOpenConns: 50,
    MaxIdleConns: 10,
}
```

### CacheConfig

```go
type CacheConfig struct {
    Driver          string
    DefaultTTL      time.Duration
    CleanupInterval time.Duration
    MaxSize         int64
    RedisAddr       string
    RedisPassword   string
    RedisDB         int
    MemcachedAddrs  []string
}
```

Cache system configuration.

**Defaults:**
- `Driver`: "memory"
- `DefaultTTL`: 5 minutes
- `CleanupInterval`: 10 minutes
- `MaxSize`: 100MB

**Supported Drivers:**
- `memory`: In-memory cache (default)
- `redis`: Redis cache
- `memcached`: Memcached cache

**Example:**
```go
config := pkg.CacheConfig{
    Driver: "redis",
    DefaultTTL: 10 * time.Minute,
    RedisAddr: "localhost:6379",
    RedisPassword: "",
    RedisDB: 0,
}
```

### SessionConfig

```go
type SessionConfig struct {
    CookieName     string
    CookiePath     string
    CookieDomain   string
    CookieSecure   bool
    CookieHTTPOnly bool
    CookieSameSite string
    Lifetime       time.Duration
    IDLength       int
    Storage        string
}
```

Session management configuration.

**Defaults:**
- `CookieName`: "session_id"
- `CookiePath`: "/"
- `CookieSecure`: false
- `CookieHTTPOnly`: true
- `CookieSameSite`: "Lax"
- `Lifetime`: 24 hours
- `IDLength`: 32 bytes
- `Storage`: "memory"

**Storage Options:**
- `memory`: In-memory storage
- `database`: Database storage
- `cache`: Cache storage

**Example:**
```go
config := pkg.SessionConfig{
    CookieName: "app_session",
    CookieSecure: true,
    CookieHTTPOnly: true,
    CookieSameSite: "Strict",
    Lifetime: 7 * 24 * time.Hour, // 7 days
    Storage: "database",
}
```

---

## Data Models

### Session

```go
type Session struct {
    ID        string
    UserID    string
    TenantID  string
    Data      map[string]interface{}
    ExpiresAt time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
    IPAddress string
    UserAgent string
}
```

User session model stored in database or cache.

**Fields:**
- `ID`: Unique session identifier (UUID or secure random string)
- `UserID`: Associated user ID
- `TenantID`: Associated tenant ID (for multi-tenancy)
- `Data`: Arbitrary session data as key-value pairs
- `ExpiresAt`: Session expiration timestamp
- `CreatedAt`: Session creation timestamp
- `UpdatedAt`: Last update timestamp
- `IPAddress`: Client IP address
- `UserAgent`: Client user agent string

**Example:**
```go
session, err := ctx.Session().Create(ctx)
session.Set("user_id", "user_123")
session.Set("cart", []string{"item1", "item2"})
ctx.Session().Save(ctx, session)
```

### User

```go
type User struct {
    ID        string
    Username  string
    Email     string
    Password  string
    Roles     []string
    TenantID  string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

User model for authentication and authorization.

**Fields:**
- `ID`: Unique user identifier
- `Username`: User's username
- `Email`: User's email address
- `Password`: Hashed password
- `Roles`: User's roles for RBAC
- `TenantID`: Associated tenant ID
- `IsActive`: Whether user account is active
- `CreatedAt`: Account creation timestamp
- `UpdatedAt`: Last update timestamp

**Example:**
```go
user := ctx.User()
if user != nil {
    fmt.Printf("Logged in as: %s\n", user.Username)
}
```

### Tenant

```go
type Tenant struct {
    ID          string
    Name        string
    Hosts       []string
    Config      map[string]interface{}
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    MaxUsers    int
    MaxStorage  int64
    MaxRequests int64
}
```

Tenant model for multi-tenancy support.

**Fields:**
- `ID`: Unique tenant identifier
- `Name`: Human-readable tenant name
- `Hosts`: Hostnames associated with this tenant
- `Config`: Tenant-specific configuration
- `IsActive`: Whether tenant is active
- `CreatedAt`: Tenant creation timestamp
- `UpdatedAt`: Last update timestamp
- `MaxUsers`: Maximum users allowed
- `MaxStorage`: Maximum storage in bytes
- `MaxRequests`: Maximum requests per period

**Example:**
```go
tenant := ctx.Tenant()
if tenant != nil {
    fmt.Printf("Tenant: %s\n", tenant.Name)
}
```

### AccessToken

```go
type AccessToken struct {
    Token     string
    UserID    string
    TenantID  string
    Scopes    []string
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

API access token for authentication.

**Fields:**
- `Token`: The actual token string (cryptographically secure)
- `UserID`: Associated user ID
- `TenantID`: Associated tenant ID
- `Scopes`: Permission scopes granted to this token
- `ExpiresAt`: Token expiration timestamp
- `CreatedAt`: Token creation timestamp

**Example:**
```go
token, err := ctx.Security().GenerateAccessToken(userID, scopes, 1*time.Hour)
```

### WorkloadMetrics

```go
type WorkloadMetrics struct {
    ID           int64
    Timestamp    time.Time
    TenantID     string
    UserID       string
    RequestID    string
    Duration     int64
    ContextSize  int64
    MemoryUsage  int64
    CPUUsage     float64
    Path         string
    Method       string
    StatusCode   int
    ResponseSize int64
    ErrorMessage string
}
```

Performance and usage metrics for monitoring.

**Fields:**
- `ID`: Auto-incrementing unique identifier
- `Timestamp`: When metric was recorded
- `TenantID`: Associated tenant ID
- `UserID`: Associated user ID (optional)
- `RequestID`: Unique request identifier
- `Duration`: Request duration in milliseconds
- `ContextSize`: Request context size in bytes
- `MemoryUsage`: Memory used in bytes
- `CPUUsage`: CPU usage percentage
- `Path`: Request path
- `Method`: HTTP method
- `StatusCode`: HTTP status code
- `ResponseSize`: Response size in bytes
- `ErrorMessage`: Error message if failed

**Example:**
```go
metrics, err := ctx.DB().GetWorkloadMetrics(tenantID, from, to)
```

---

## Request/Response Types

### Request

```go
type Request struct {
    // Standard HTTP request fields
    Method     string
    URL        *url.URL
    Proto      string
    Header     http.Header
    Body       io.ReadCloser
    Host       string
    RemoteAddr string
    RequestURI string

    // Framework-specific fields
    ID        string               // Unique request ID
    TenantID  string               // Multi-tenancy support
    UserID    string               // Authenticated user ID
    StartTime time.Time            // Request start time
    Params    map[string]string    // Route parameters
    Query     map[string]string    // Query parameters
    Form      map[string]string    // Form data
    Files     map[string]*FormFile // Uploaded files

    // Security context
    AccessToken string // API access token
    SessionID   string // Session identifier

    // Protocol information
    IsWebSocket bool   // WebSocket upgrade request
    Protocol    string // HTTP/1, HTTP/2, QUIC, WebSocket

    // Raw request data
    RawBody []byte // Cached body content
}
```

Enhanced HTTP request with framework-specific features.

**Standard HTTP Fields:**
- `Method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `URL`: Parsed URL with path, query, and fragment
- `Proto`: Protocol version (HTTP/1.1, HTTP/2.0, etc.)
- `Header`: HTTP headers as key-value pairs
- `Body`: Request body as a readable stream
- `Host`: Host header value
- `RemoteAddr`: Client IP address and port
- `RequestURI`: Unparsed request URI

**Framework Fields:**
- `ID`: Unique identifier for request tracking and logging
- `TenantID`: Tenant identifier for multi-tenant applications
- `UserID`: Authenticated user identifier (set by auth middleware)
- `StartTime`: Timestamp when request processing began
- `Params`: Route parameters extracted from URL patterns (e.g., `/users/:id`)
- `Query`: Parsed query string parameters
- `Form`: Parsed form data from POST requests
- `Files`: Uploaded files from multipart forms

**Security Fields:**
- `AccessToken`: Bearer token or API key for authentication
- `SessionID`: Session identifier for stateful sessions

**Protocol Fields:**
- `IsWebSocket`: True if this is a WebSocket upgrade request
- `Protocol`: Protocol identifier (HTTP/1, HTTP/2, QUIC, WebSocket)

**Raw Data:**
- `RawBody`: Cached request body content (populated when body is read)

**Usage Examples:**

**Accessing Route Parameters:**
```go
func getUserHandler(ctx pkg.Context) error {
    req := ctx.Request()
    userID := req.Params["id"]  // From route /users/:id
    
    return ctx.JSON(200, map[string]string{
        "user_id": userID,
    })
}
```

**Reading Query Parameters:**
```go
func searchHandler(ctx pkg.Context) error {
    req := ctx.Request()
    query := req.Query["q"]
    page := req.Query["page"]
    
    // Perform search...
    return ctx.JSON(200, results)
}
```

**Accessing Request Headers:**
```go
func apiHandler(ctx pkg.Context) error {
    req := ctx.Request()
    authHeader := req.Header.Get("Authorization")
    contentType := req.Header.Get("Content-Type")
    
    // Process request...
    return ctx.JSON(200, data)
}
```

**Protocol-Specific Behavior:**
```go
func protocolAwareHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    if req.IsWebSocket {
        // Handle WebSocket upgrade
        return ctx.UpgradeWebSocket(wsHandler)
    }
    
    if req.Protocol == "HTTP/2" {
        // Use HTTP/2 server push
        ctx.Response().Push("/static/app.css", nil)
    }
    
    return ctx.JSON(200, data)
}
```

**Multi-Tenant Request Handling:**
```go
func tenantHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    // TenantID is automatically populated by framework
    if req.TenantID == "" {
        return pkg.NewAuthorizationError("No tenant context")
    }
    
    // Load tenant-specific data
    data := loadDataForTenant(req.TenantID)
    return ctx.JSON(200, data)
}
```

**Request Tracking:**
```go
func trackedHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    // Log request with unique ID
    ctx.Logger().Info("Processing request", 
        "request_id", req.ID,
        "method", req.Method,
        "path", req.URL.Path,
        "duration", time.Since(req.StartTime))
    
    return ctx.JSON(200, data)
}
```

**See Also:**
- [Context API](context.md) - Accessing request through context
- [Router API](router.md) - Route parameter extraction
- [Security API](security.md) - Authentication and authorization

### Cookie

```go
type Cookie struct {
    Name     string
    Value    string
    Path     string
    Domain   string
    Expires  time.Time
    MaxAge   int
    Secure   bool
    HttpOnly bool
    SameSite http.SameSite

    // Framework extensions
    Encrypted bool // Whether cookie value is encrypted
}
```

HTTP cookie with framework-specific encryption support.

**Standard Fields:**
- `Name`: Cookie name (must be unique per domain/path)
- `Value`: Cookie value (plain text or encrypted)
- `Path`: URL path where cookie is valid (default: "/")
- `Domain`: Domain where cookie is valid (default: current domain)
- `Expires`: Absolute expiration time (alternative to MaxAge)
- `MaxAge`: Relative expiration in seconds (0 = session cookie, -1 = delete)
- `Secure`: Only send over HTTPS connections
- `HttpOnly`: Prevent JavaScript access (XSS protection)
- `SameSite`: CSRF protection policy

**Framework Field:**
- `Encrypted`: Indicates if the cookie value is encrypted (set automatically by framework)

**SameSite Values:**
- `http.SameSiteDefaultMode` (0): Browser default behavior
- `http.SameSiteLaxMode` (1): Sent with top-level navigation and same-site requests
- `http.SameSiteStrictMode` (2): Only sent with same-site requests
- `http.SameSiteNoneMode` (3): Sent with all requests (requires Secure=true)

**Basic Cookie Usage:**

```go
func setPreferenceCookie(ctx pkg.Context) error {
    cookie := &pkg.Cookie{
        Name:     "theme",
        Value:    "dark",
        Path:     "/",
        MaxAge:   86400 * 30, // 30 days
        HttpOnly: false,      // Allow JavaScript access
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    
    return ctx.SetCookie(cookie)
}

func getPreferenceCookie(ctx pkg.Context) error {
    cookie, err := ctx.GetCookie("theme")
    if err != nil {
        return err
    }
    
    theme := cookie.Value // "dark"
    return ctx.JSON(200, map[string]string{"theme": theme})
}
```

**Encrypted Cookies:**

The framework provides automatic cookie encryption using AES-256-GCM for sensitive data.

**Setting Encrypted Cookies:**
```go
func setSessionCookie(ctx pkg.Context) error {
    // Configure encryption key (32 bytes for AES-256)
    cookieManager, err := pkg.NewCookieManager(&pkg.CookieConfig{
        EncryptionKey: []byte("your-32-byte-encryption-key!!"),
    })
    if err != nil {
        return err
    }
    
    cookie := &pkg.Cookie{
        Name:     "session_data",
        Value:    "sensitive-user-data",
        Path:     "/",
        MaxAge:   3600, // 1 hour
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }
    
    // Encrypt and set cookie
    return cookieManager.SetEncryptedCookie(ctx, cookie)
}
```

**Reading Encrypted Cookies:**
```go
func getSessionCookie(ctx pkg.Context) error {
    cookieManager, err := pkg.NewCookieManager(&pkg.CookieConfig{
        EncryptionKey: []byte("your-32-byte-encryption-key!!"),
    })
    if err != nil {
        return err
    }
    
    // Automatically decrypts the cookie value
    cookie, err := cookieManager.GetEncryptedCookie(ctx, "session_data")
    if err != nil {
        return err
    }
    
    // Value is decrypted
    sessionData := cookie.Value
    return ctx.JSON(200, map[string]string{"data": sessionData})
}
```

**Manual Encryption/Decryption:**
```go
func manualEncryption(ctx pkg.Context) error {
    cookieManager, err := pkg.NewCookieManager(&pkg.CookieConfig{
        EncryptionKey: []byte("your-32-byte-encryption-key!!"),
    })
    if err != nil {
        return err
    }
    
    // Encrypt a value
    plaintext := "secret-data"
    encrypted, err := cookieManager.EncryptValue(plaintext)
    if err != nil {
        return err
    }
    
    // Decrypt a value
    decrypted, err := cookieManager.DecryptValue(encrypted)
    if err != nil {
        return err
    }
    
    // decrypted == plaintext
    return ctx.JSON(200, map[string]string{
        "encrypted": encrypted,
        "decrypted": decrypted,
    })
}
```

**Secure Cookie Patterns:**

**1. Authentication Tokens:**
```go
func setAuthCookie(ctx pkg.Context, token string) error {
    cookie := &pkg.Cookie{
        Name:     "auth_token",
        Value:    token,
        Path:     "/",
        MaxAge:   3600 * 24 * 7, // 7 days
        HttpOnly: true,           // Prevent XSS
        Secure:   true,           // HTTPS only
        SameSite: http.SameSiteStrictMode, // Prevent CSRF
        Encrypted: true,          // Encrypt token
    }
    
    return ctx.SetCookie(cookie)
}
```

**2. Remember Me Functionality:**
```go
func setRememberMeCookie(ctx pkg.Context, userID string) error {
    // Generate secure token
    token := generateSecureToken()
    
    cookie := &pkg.Cookie{
        Name:     "remember_me",
        Value:    fmt.Sprintf("%s:%s", userID, token),
        Path:     "/",
        MaxAge:   3600 * 24 * 30, // 30 days
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Encrypted: true,
    }
    
    return ctx.SetCookie(cookie)
}
```

**3. CSRF Token Cookies:**
```go
func setCSRFCookie(ctx pkg.Context) error {
    csrfToken := generateCSRFToken()
    
    cookie := &pkg.Cookie{
        Name:     "csrf_token",
        Value:    csrfToken,
        Path:     "/",
        MaxAge:   3600, // 1 hour
        HttpOnly: false, // JavaScript needs to read this
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }
    
    return ctx.SetCookie(cookie)
}
```

**4. Session Cookies:**
```go
func setSessionCookie(ctx pkg.Context, sessionID string) error {
    cookie := &pkg.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        MaxAge:   0, // Session cookie (deleted when browser closes)
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        Encrypted: true,
    }
    
    return ctx.SetCookie(cookie)
}
```

**Deleting Cookies:**
```go
func deleteCookie(ctx pkg.Context, name string) error {
    cookie := &pkg.Cookie{
        Name:   name,
        Value:  "",
        Path:   "/",
        MaxAge: -1, // Delete immediately
    }
    
    return ctx.SetCookie(cookie)
}
```

**Security Best Practices:**

1. **Always use Secure flag in production:**
```go
cookie.Secure = ctx.Config().IsProduction()
```

2. **Use HttpOnly for sensitive cookies:**
```go
cookie.HttpOnly = true // Prevents XSS attacks
```

3. **Set appropriate SameSite policy:**
```go
// Strict: Maximum protection, may break some flows
cookie.SameSite = http.SameSiteStrictMode

// Lax: Good balance (default recommendation)
cookie.SameSite = http.SameSiteLaxMode

// None: Only if you need cross-site cookies (requires Secure=true)
cookie.SameSite = http.SameSiteNoneMode
```

4. **Encrypt sensitive data:**
```go
// Always encrypt authentication tokens, session data, PII
cookie.Encrypted = true
```

5. **Set minimal Path and Domain:**
```go
// Limit cookie scope to specific paths
cookie.Path = "/api"
cookie.Domain = "api.example.com"
```

6. **Use short expiration for sensitive cookies:**
```go
// Auth cookies: 1-24 hours
cookie.MaxAge = 3600

// Remember me: 7-30 days
cookie.MaxAge = 3600 * 24 * 7

// Preferences: 30-365 days
cookie.MaxAge = 3600 * 24 * 30
```

**Cookie Size Limits:**

Browsers typically limit cookies to 4KB. For larger data, use sessions:

```go
// Bad: Cookie too large
cookie.Value = largeJSONString // May exceed 4KB

// Good: Store in session
session, _ := ctx.Session().Create(ctx)
session.Set("large_data", largeObject)
ctx.Session().Save(ctx, session)
```

**See Also:**
- [Session API](session.md) - Session management
- [Security API](security.md) - Authentication and CSRF protection
- [Context API](context.md) - Cookie operations through context

### FormFile

```go
type FormFile struct {
    Filename string
    Header   map[string][]string
    Size     int64
    Content  []byte
}
```

Represents an uploaded file from a multipart form submission.

**Fields:**
- `Filename`: Original filename as provided by the client
- `Header`: MIME headers from the multipart section (Content-Type, Content-Disposition, etc.)
- `Size`: File size in bytes
- `Content`: Complete file content loaded into memory

**Header Structure:**

The `Header` field contains MIME headers from the multipart form section:

```go
file.Header["Content-Type"]        // ["image/jpeg"]
file.Header["Content-Disposition"] // ["form-data; name=\"file\"; filename=\"photo.jpg\""]
```

Common headers:
- `Content-Type`: MIME type of the uploaded file
- `Content-Disposition`: Form field name and original filename
- `Content-Transfer-Encoding`: Encoding used for transfer

**Memory Management:**

The framework loads uploaded files into memory by default. For large files, this can cause memory pressure. Consider these strategies:

**Small Files (< 10MB):**
```go
// Direct memory access is fine
file, err := ctx.FormFile("avatar")
if err != nil {
    return err
}

// File content is already in memory
data := file.Content
```

**Large Files (> 10MB):**

For large files, use streaming to avoid loading the entire file into memory:

```go
func uploadLargeFile(ctx pkg.Context) error {
    // Get the multipart form
    err := ctx.Request().Body.ParseMultipartForm(32 << 20) // 32MB max memory
    if err != nil {
        return err
    }
    
    // Get file header (not content)
    file, header, err := ctx.Request().FormFile("large_file")
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Stream to disk without loading into memory
    dst, err := os.Create("uploads/" + header.Filename)
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // Copy in chunks
    _, err = io.Copy(dst, file)
    return err
}
```

**Streaming Upload Example:**

```go
func streamingUploadHandler(ctx pkg.Context) error {
    reader, err := ctx.Request().Body.MultipartReader()
    if err != nil {
        return err
    }
    
    for {
        part, err := reader.NextPart()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        // Process each part as a stream
        filename := part.FileName()
        if filename == "" {
            continue // Skip non-file parts
        }
        
        // Stream directly to storage
        dst, err := os.Create("uploads/" + filename)
        if err != nil {
            return err
        }
        
        written, err := io.Copy(dst, part)
        dst.Close()
        
        if err != nil {
            return err
        }
        
        ctx.Logger().Info("File uploaded",
            "filename", filename,
            "size", written)
    }
    
    return ctx.JSON(200, map[string]string{
        "status": "uploaded",
    })
}
```

**Best Practices:**

1. **Validate File Size:**
```go
file, err := ctx.FormFile("upload")
if err != nil {
    return err
}

maxSize := int64(10 * 1024 * 1024) // 10MB
if file.Size > maxSize {
    return pkg.NewValidationError("File too large")
}
```

2. **Validate File Type:**
```go
file, err := ctx.FormFile("image")
if err != nil {
    return err
}

contentType := file.Header["Content-Type"][0]
allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}

if !contains(allowedTypes, contentType) {
    return pkg.NewValidationError("Invalid file type")
}
```

3. **Sanitize Filenames:**
```go
import (
    "path/filepath"
    "strings"
)

func sanitizeFilename(filename string) string {
    // Remove path components
    filename = filepath.Base(filename)
    
    // Remove dangerous characters
    filename = strings.ReplaceAll(filename, "..", "")
    filename = strings.ReplaceAll(filename, "/", "")
    filename = strings.ReplaceAll(filename, "\\", "")
    
    return filename
}

file, err := ctx.FormFile("upload")
if err != nil {
    return err
}

safeFilename := sanitizeFilename(file.Filename)
```

4. **Use Temporary Storage:**
```go
func processUpload(ctx pkg.Context) error {
    file, err := ctx.FormFile("data")
    if err != nil {
        return err
    }
    
    // Write to temporary file
    tmpFile, err := os.CreateTemp("", "upload-*.tmp")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name()) // Clean up
    defer tmpFile.Close()
    
    _, err = tmpFile.Write(file.Content)
    if err != nil {
        return err
    }
    
    // Process the temporary file
    result, err := processFile(tmpFile.Name())
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, result)
}
```

5. **Limit Memory Usage:**
```go
// Configure maximum memory for multipart forms
func configureServer() *pkg.FrameworkConfig {
    return &pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            MaxHeaderBytes: 10 << 20, // 10MB max for headers + small files
        },
    }
}
```

**Multiple File Upload:**
```go
func multipleFilesHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    // Access all uploaded files
    for fieldName, file := range req.Files {
        ctx.Logger().Info("Processing file",
            "field", fieldName,
            "filename", file.Filename,
            "size", file.Size)
        
        // Save each file
        path := fmt.Sprintf("uploads/%s", file.Filename)
        err := ctx.Files().Write(path, file.Content)
        if err != nil {
            return err
        }
    }
    
    return ctx.JSON(200, map[string]string{
        "status": "all files uploaded",
        "count": fmt.Sprintf("%d", len(req.Files)),
    })
}
```

**See Also:**
- [Forms API](forms.md) - Form handling and validation
- [Context API](context.md) - Request context methods
- [File System API](utilities.md#filesystem) - File operations

### HostConfig

```go
type HostConfig struct {
    Hostname       string
    TenantID       string
    VirtualFS      VirtualFS
    Middleware     []MiddlewareFunc
    RateLimits     *RateLimitConfig
    SecurityConfig *ServerSecurityConfig
}
```

Host-specific configuration for multi-tenant and virtual host setups.

**Fields:**
- `Hostname`: Domain or hostname for this configuration (e.g., "api.example.com")
- `TenantID`: Tenant identifier for multi-tenancy isolation
- `VirtualFS`: Virtual file system for host-specific static files
- `Middleware`: Host-specific middleware stack
- `RateLimits`: Rate limiting configuration for this host
- `SecurityConfig`: Security settings specific to this host

**Use Cases:**

HostConfig enables several advanced patterns:
- Multi-tenant applications with host-based routing
- Virtual hosts with different configurations
- Per-domain middleware and security policies
- Tenant-specific static file serving

**Basic Multi-Host Configuration:**

```go
func configureMultiHost() *pkg.FrameworkConfig {
    return &pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            Address: ":8080",
            Hosts: []*pkg.HostConfig{
                {
                    Hostname: "api.example.com",
                    TenantID: "tenant-1",
                    RateLimits: &pkg.RateLimitConfig{
                        Enabled:           true,
                        RequestsPerSecond: 100,
                        BurstSize:         20,
                    },
                },
                {
                    Hostname: "admin.example.com",
                    TenantID: "tenant-1",
                    RateLimits: &pkg.RateLimitConfig{
                        Enabled:           true,
                        RequestsPerSecond: 50,
                        BurstSize:         10,
                    },
                },
            },
        },
    }
}
```

**Multi-Tenant Application:**

```go
func setupMultiTenant(app *pkg.Framework) error {
    // Configure tenant 1
    tenant1Config := &pkg.HostConfig{
        Hostname: "tenant1.myapp.com",
        TenantID: "tenant-1",
        Middleware: []pkg.MiddlewareFunc{
            tenant1AuthMiddleware,
            tenant1LoggingMiddleware,
        },
        RateLimits: &pkg.RateLimitConfig{
            Enabled:           true,
            RequestsPerSecond: 1000,
            BurstSize:         100,
        },
    }
    
    // Configure tenant 2
    tenant2Config := &pkg.HostConfig{
        Hostname: "tenant2.myapp.com",
        TenantID: "tenant-2",
        Middleware: []pkg.MiddlewareFunc{
            tenant2AuthMiddleware,
            tenant2LoggingMiddleware,
        },
        RateLimits: &pkg.RateLimitConfig{
            Enabled:           true,
            RequestsPerSecond: 500,
            BurstSize:         50,
        },
    }
    
    // Register host configurations
    app.Server().AddHost(tenant1Config)
    app.Server().AddHost(tenant2Config)
    
    return nil
}
```

**Virtual Host Pattern:**

```go
func setupVirtualHosts(app *pkg.Framework) error {
    // Main website
    mainSite := &pkg.HostConfig{
        Hostname: "www.example.com",
        VirtualFS: pkg.NewVirtualFS("./public/main"),
        Middleware: []pkg.MiddlewareFunc{
            corsMiddleware,
            compressionMiddleware,
        },
    }
    
    // Blog subdomain
    blogSite := &pkg.HostConfig{
        Hostname: "blog.example.com",
        VirtualFS: pkg.NewVirtualFS("./public/blog"),
        Middleware: []pkg.MiddlewareFunc{
            blogAuthMiddleware,
        },
    }
    
    // API subdomain
    apiSite := &pkg.HostConfig{
        Hostname: "api.example.com",
        TenantID: "api",
        Middleware: []pkg.MiddlewareFunc{
            apiAuthMiddleware,
            apiRateLimitMiddleware,
        },
        RateLimits: &pkg.RateLimitConfig{
            Enabled:           true,
            RequestsPerSecond: 10000,
            BurstSize:         1000,
        },
    }
    
    app.Server().AddHost(mainSite)
    app.Server().AddHost(blogSite)
    app.Server().AddHost(apiSite)
    
    return nil
}
```

**Host-Based Routing:**

```go
func setupHostRouting(app *pkg.Framework) {
    router := app.Router()
    
    // Routes for api.example.com
    router.Host("api.example.com").Group(func(api *pkg.RouterGroup) {
        api.GET("/users", listUsers)
        api.POST("/users", createUser)
        api.GET("/users/:id", getUser)
    })
    
    // Routes for admin.example.com
    router.Host("admin.example.com").Group(func(admin *pkg.RouterGroup) {
        admin.GET("/dashboard", adminDashboard)
        admin.GET("/users", adminListUsers)
        admin.POST("/settings", updateSettings)
    })
    
    // Routes for www.example.com
    router.Host("www.example.com").Group(func(www *pkg.RouterGroup) {
        www.GET("/", homepage)
        www.GET("/about", aboutPage)
        www.GET("/contact", contactPage)
    })
}
```

**Per-Host Middleware:**

```go
func hostSpecificMiddleware(hostname string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        // Add host-specific headers
        ctx.SetHeader("X-Served-By", hostname)
        
        // Host-specific logging
        ctx.Logger().Info("Request received",
            "host", hostname,
            "path", ctx.Request().URL.Path)
        
        return next(ctx)
    }
}

func setupHostMiddleware() *pkg.HostConfig {
    return &pkg.HostConfig{
        Hostname: "api.example.com",
        Middleware: []pkg.MiddlewareFunc{
            hostSpecificMiddleware("api.example.com"),
            apiAuthMiddleware,
            apiMetricsMiddleware,
        },
    }
}
```

**Tenant Isolation:**

```go
func tenantIsolationHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    // TenantID is automatically set from HostConfig
    if req.TenantID == "" {
        return pkg.NewAuthorizationError("No tenant context")
    }
    
    // Load tenant-specific data
    tenant := ctx.Tenant()
    if tenant == nil || !tenant.IsActive {
        return pkg.NewAuthorizationError("Tenant not active")
    }
    
    // All database queries are automatically scoped to tenant
    users, err := ctx.DB().Query(
        "SELECT * FROM users WHERE tenant_id = ?",
        req.TenantID,
    )
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, users)
}
```

**Host-Specific Security:**

```go
func setupHostSecurity() []*pkg.HostConfig {
    return []*pkg.HostConfig{
        {
            Hostname: "public.example.com",
            SecurityConfig: &pkg.ServerSecurityConfig{
                EnableCORS: true,
                CORSConfig: &pkg.CORSConfig{
                    AllowOrigins: []string{"*"},
                    AllowMethods: []string{"GET", "POST"},
                },
                MaxRequestSize: 1 << 20, // 1MB
            },
        },
        {
            Hostname: "admin.example.com",
            SecurityConfig: &pkg.ServerSecurityConfig{
                EnableCORS: false,
                EnableCSRF: true,
                EnableXSS:  true,
                MaxRequestSize: 10 << 20, // 10MB
                RequestTimeout: 30 * time.Second,
            },
        },
    }
}
```

**Dynamic Host Configuration:**

```go
func loadHostConfigsFromDatabase(db *sql.DB) ([]*pkg.HostConfig, error) {
    rows, err := db.Query(`
        SELECT hostname, tenant_id, rate_limit_rps, rate_limit_burst
        FROM host_configs
        WHERE active = true
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var configs []*pkg.HostConfig
    
    for rows.Next() {
        var hostname, tenantID string
        var rps, burst int
        
        err := rows.Scan(&hostname, &tenantID, &rps, &burst)
        if err != nil {
            return nil, err
        }
        
        config := &pkg.HostConfig{
            Hostname: hostname,
            TenantID: tenantID,
            RateLimits: &pkg.RateLimitConfig{
                Enabled:           true,
                RequestsPerSecond: rps,
                BurstSize:         burst,
            },
        }
        
        configs = append(configs, config)
    }
    
    return configs, nil
}
```

**Wildcard Host Matching:**

```go
func setupWildcardHosts(app *pkg.Framework) {
    // Match any subdomain of example.com
    wildcardConfig := &pkg.HostConfig{
        Hostname: "*.example.com",
        Middleware: []pkg.MiddlewareFunc{
            extractSubdomainMiddleware,
        },
    }
    
    app.Server().AddHost(wildcardConfig)
}

func extractSubdomainMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    host := ctx.Request().Host
    parts := strings.Split(host, ".")
    
    if len(parts) >= 3 {
        subdomain := parts[0]
        ctx.Set("subdomain", subdomain)
        
        // Use subdomain as tenant ID
        ctx.Request().TenantID = subdomain
    }
    
    return next(ctx)
}
```

**See Also:**
- [Multi-Tenancy Guide](../guides/multi-tenancy.md) - Multi-tenant architecture
- [Server API](server.md) - Server configuration
- [Router API](router.md) - Host-based routing
- [Middleware API](middleware.md) - Middleware configuration

### RateLimitConfig

```go
type RateLimitConfig struct {
    Enabled           bool
    RequestsPerSecond int
    BurstSize         int
    Storage           string // "memory", "database", "redis"
}
```

Rate limiting configuration for controlling request throughput.

**Fields:**
- `Enabled`: Whether rate limiting is active
- `RequestsPerSecond`: Maximum sustained requests per second
- `BurstSize`: Maximum burst of requests allowed above the rate
- `Storage`: Backend storage for rate limit counters

**Rate Limiting Algorithm:**

The framework uses a token bucket algorithm:
- Tokens are added at `RequestsPerSecond` rate
- Each request consumes one token
- `BurstSize` determines the maximum token accumulation
- Requests are rejected when no tokens are available

**Storage Options:**
- `memory`: In-memory storage (fast, but not shared across instances)
- `database`: Database storage (shared, persistent, slower)
- `redis`: Redis storage (shared, fast, recommended for production)

**Basic Rate Limiting:**

```go
func configureRateLimit() *pkg.FrameworkConfig {
    return &pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            Address: ":8080",
            RateLimits: &pkg.RateLimitConfig{
                Enabled:           true,
                RequestsPerSecond: 100,  // 100 requests/second sustained
                BurstSize:         20,   // Allow bursts up to 120 req/s
                Storage:           "memory",
            },
        },
    }
}
```

**Per-Host Rate Limiting:**

```go
func setupPerHostRateLimits() []*pkg.HostConfig {
    return []*pkg.HostConfig{
        {
            Hostname: "api.example.com",
            RateLimits: &pkg.RateLimitConfig{
                Enabled:           true,
                RequestsPerSecond: 1000,
                BurstSize:         200,
                Storage:           "redis",
            },
        },
        {
            Hostname: "public.example.com",
            RateLimits: &pkg.RateLimitConfig{
                Enabled:           true,
                RequestsPerSecond: 100,
                BurstSize:         20,
                Storage:           "memory",
            },
        },
    }
}
```

**Per-User Rate Limiting:**

```go
func perUserRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    user := ctx.User()
    if user == nil {
        return pkg.NewAuthenticationError("Authentication required")
    }
    
    // Check user-specific rate limit
    key := fmt.Sprintf("rate_limit:user:%s", user.ID)
    
    allowed, err := ctx.Cache().CheckRateLimit(key, 100, time.Minute)
    if err != nil {
        return err
    }
    
    if !allowed {
        return pkg.NewRateLimitError(60 * time.Second)
    }
    
    return next(ctx)
}
```

**Per-IP Rate Limiting:**

```go
func perIPRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    req := ctx.Request()
    
    // Extract client IP
    clientIP := req.RemoteAddr
    if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
        clientIP = strings.Split(forwardedFor, ",")[0]
    }
    
    // Check IP-specific rate limit
    key := fmt.Sprintf("rate_limit:ip:%s", clientIP)
    
    allowed, err := ctx.Cache().CheckRateLimit(key, 1000, time.Hour)
    if err != nil {
        return err
    }
    
    if !allowed {
        // Return 429 Too Many Requests
        ctx.SetHeader("Retry-After", "3600")
        return pkg.NewRateLimitError(time.Hour)
    }
    
    return next(ctx)
}
```

**Per-Endpoint Rate Limiting:**

```go
func setupEndpointRateLimits(router *pkg.Router) {
    // Expensive endpoint: 10 req/min
    router.GET("/api/expensive", rateLimitMiddleware(10, time.Minute), expensiveHandler)
    
    // Normal endpoint: 100 req/min
    router.GET("/api/normal", rateLimitMiddleware(100, time.Minute), normalHandler)
    
    // Public endpoint: 1000 req/min
    router.GET("/api/public", rateLimitMiddleware(1000, time.Minute), publicHandler)
}

func rateLimitMiddleware(requests int, window time.Duration) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        key := fmt.Sprintf("rate_limit:endpoint:%s", ctx.Request().URL.Path)
        
        allowed, err := ctx.Cache().CheckRateLimit(key, requests, window)
        if err != nil {
            return err
        }
        
        if !allowed {
            return pkg.NewRateLimitError(window)
        }
        
        return next(ctx)
    }
}
```

**Tiered Rate Limiting:**

```go
func tieredRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    user := ctx.User()
    if user == nil {
        return pkg.NewAuthenticationError("Authentication required")
    }
    
    // Determine rate limit based on user tier
    var limit int
    var window time.Duration
    
    switch user.Tier {
    case "free":
        limit = 100
        window = time.Hour
    case "pro":
        limit = 1000
        window = time.Hour
    case "enterprise":
        limit = 10000
        window = time.Hour
    default:
        limit = 10
        window = time.Hour
    }
    
    key := fmt.Sprintf("rate_limit:user:%s", user.ID)
    allowed, err := ctx.Cache().CheckRateLimit(key, limit, window)
    if err != nil {
        return err
    }
    
    if !allowed {
        return pkg.NewRateLimitError(window)
    }
    
    // Add rate limit headers
    ctx.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
    ctx.SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", getRemainingRequests(key)))
    ctx.SetHeader("X-RateLimit-Reset", fmt.Sprintf("%d", getResetTime(key).Unix()))
    
    return next(ctx)
}
```

**Distributed Rate Limiting with Redis:**

```go
func configureRedisRateLimit() *pkg.FrameworkConfig {
    return &pkg.FrameworkConfig{
        CacheConfig: pkg.CacheConfig{
            Driver:    "redis",
            RedisAddr: "localhost:6379",
        },
        ServerConfig: pkg.ServerConfig{
            RateLimits: &pkg.RateLimitConfig{
                Enabled:           true,
                RequestsPerSecond: 1000,
                BurstSize:         200,
                Storage:           "redis", // Shared across all instances
            },
        },
    }
}
```

**Rate Limit Bypass:**

```go
func rateLimitBypassMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check for bypass token
    bypassToken := ctx.GetHeader("X-RateLimit-Bypass")
    
    if bypassToken != "" && validateBypassToken(bypassToken) {
        // Skip rate limiting
        ctx.Set("rate_limit_bypassed", true)
        return next(ctx)
    }
    
    // Apply normal rate limiting
    return applyRateLimit(ctx, next)
}

func validateBypassToken(token string) bool {
    // Validate against known bypass tokens
    validTokens := []string{
        "admin-bypass-token",
        "monitoring-bypass-token",
    }
    
    for _, valid := range validTokens {
        if token == valid {
            return true
        }
    }
    
    return false
}
```

**Adaptive Rate Limiting:**

```go
func adaptiveRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Get current system load
    load := getSystemLoad()
    
    // Adjust rate limit based on load
    var limit int
    switch {
    case load < 0.5:
        limit = 1000 // Low load: generous limit
    case load < 0.8:
        limit = 500  // Medium load: moderate limit
    default:
        limit = 100  // High load: strict limit
    }
    
    key := fmt.Sprintf("rate_limit:adaptive:%s", ctx.Request().RemoteAddr)
    allowed, err := ctx.Cache().CheckRateLimit(key, limit, time.Minute)
    if err != nil {
        return err
    }
    
    if !allowed {
        return pkg.NewRateLimitError(time.Minute)
    }
    
    return next(ctx)
}
```

**Rate Limit Response Headers:**

```go
func addRateLimitHeaders(ctx pkg.Context, limit, remaining int, reset time.Time) {
    ctx.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
    ctx.SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
    ctx.SetHeader("X-RateLimit-Reset", fmt.Sprintf("%d", reset.Unix()))
    
    if remaining == 0 {
        ctx.SetHeader("Retry-After", fmt.Sprintf("%d", int(time.Until(reset).Seconds())))
    }
}
```

**Custom Rate Limit Error Response:**

```go
func handleRateLimitError(ctx pkg.Context, err error) error {
    if rateLimitErr, ok := err.(*pkg.RateLimitError); ok {
        return ctx.JSON(429, map[string]interface{}{
            "error": "Rate limit exceeded",
            "retry_after": rateLimitErr.RetryAfter.Seconds(),
            "message": "Too many requests. Please try again later.",
        })
    }
    
    return err
}
```

**Monitoring Rate Limits:**

```go
func monitorRateLimits(ctx pkg.Context) error {
    // Get rate limit statistics
    stats := ctx.Cache().GetRateLimitStats()
    
    return ctx.JSON(200, map[string]interface{}{
        "total_requests": stats.TotalRequests,
        "blocked_requests": stats.BlockedRequests,
        "block_rate": float64(stats.BlockedRequests) / float64(stats.TotalRequests),
        "top_blocked_ips": stats.TopBlockedIPs,
    })
}
```

**Best Practices:**

1. **Use Redis for distributed systems:**
```go
// Ensures rate limits work across multiple server instances
config.Storage = "redis"
```

2. **Set appropriate burst sizes:**
```go
// Burst should be 10-20% of sustained rate
config.BurstSize = config.RequestsPerSecond / 5
```

3. **Implement graceful degradation:**
```go
// If rate limit storage fails, allow requests through
if err != nil {
    ctx.Logger().Error("Rate limit check failed", "error", err)
    return next(ctx) // Fail open
}
```

4. **Add informative headers:**
```go
// Help clients understand rate limits
ctx.SetHeader("X-RateLimit-Limit", "1000")
ctx.SetHeader("X-RateLimit-Remaining", "847")
ctx.SetHeader("X-RateLimit-Reset", "1640000000")
```

5. **Consider different limits for different operations:**
```go
// Read operations: higher limit
GET /api/users -> 1000 req/min

// Write operations: lower limit
POST /api/users -> 100 req/min

// Expensive operations: very low limit
POST /api/reports/generate -> 10 req/hour
```

**See Also:**
- [Security API](security.md) - Security configuration
- [Cache API](cache.md) - Rate limit storage
- [Middleware API](middleware.md) - Rate limit middleware
- [Performance Guide](../guides/performance.md) - Performance tuning

---

## Security Types

### SecurityConfig

```go
type SecurityConfig struct {
    JWTSecret          string
    JWTExpiration      time.Duration
    OAuth2Providers    map[string]OAuth2Config
    EnableCSRF         bool
    CSRFTokenLength    int
    EnableRateLimiting bool
    RateLimitRequests  int
    RateLimitWindow    time.Duration
    PasswordMinLength  int
    PasswordRequireUpper bool
    PasswordRequireLower bool
    PasswordRequireDigit bool
    PasswordRequireSpecial bool
}
```

Security and authentication configuration.

**Defaults:**
- `JWTExpiration`: 1 hour
- `EnableCSRF`: true
- `CSRFTokenLength`: 32 bytes
- `EnableRateLimiting`: false
- `RateLimitRequests`: 100
- `RateLimitWindow`: 1 minute
- `PasswordMinLength`: 8
- `PasswordRequireUpper`: true
- `PasswordRequireLower`: true
- `PasswordRequireDigit`: true
- `PasswordRequireSpecial`: false

**Example:**
```go
config := pkg.SecurityConfig{
    JWTSecret: "your-secret-key",
    JWTExpiration: 24 * time.Hour,
    EnableCSRF: true,
    EnableRateLimiting: true,
    RateLimitRequests: 1000,
    RateLimitWindow: 1 * time.Minute,
}
```

### OAuth2Config

```go
type OAuth2Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
    AuthURL      string
    TokenURL     string
}
```

OAuth2 provider configuration.

**Example:**
```go
oauth2Config := pkg.OAuth2Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURL:  "https://yourapp.com/auth/callback",
    Scopes:       []string{"email", "profile"},
    AuthURL:      "https://provider.com/oauth/authorize",
    TokenURL:     "https://provider.com/oauth/token",
}
```

---

## Monitoring Types

### MonitoringConfig

```go
type MonitoringConfig struct {
    EnableMetrics        bool
    MetricsPath          string
    MetricsPort          int
    EnablePprof          bool
    PprofPath            string
    PprofPort            int
    EnableSNMP           bool
    SNMPPort             int
    SNMPCommunity        string
    EnableOptimization   bool
    OptimizationInterval time.Duration
    RequireAuth          bool
    AuthToken            string
}
```

Monitoring and profiling configuration.

**Defaults:**
- `EnableMetrics`: false
- `MetricsPath`: "/metrics"
- `MetricsPort`: 9090
- `EnablePprof`: false
- `PprofPath`: "/debug/pprof"
- `PprofPort`: 6060
- `EnableSNMP`: false
- `SNMPPort`: 161
- `SNMPCommunity`: "public"
- `EnableOptimization`: false
- `OptimizationInterval`: 5 minutes
- `RequireAuth`: false

**Example:**
```go
config := pkg.MonitoringConfig{
    EnableMetrics: true,
    MetricsPort: 9090,
    EnablePprof: true,
    PprofPort: 6060,
    RequireAuth: true,
    AuthToken: "secret-token",
}
```

### SystemInfo

```go
type SystemInfo struct {
    NumCPU           int
    NumGoroutine     int
    MemoryAlloc      uint64
    MemoryTotalAlloc uint64
    MemorySys        uint64
    NumGC            uint32
    GCPauseTotal     uint64
    LastGCTime       time.Time
}
```

System-level information for monitoring.

**Example:**
```go
data, err := app.Monitoring().GetSNMPData()
fmt.Printf("CPUs: %d, Goroutines: %d\n", 
    data.SystemInfo.NumCPU, 
    data.SystemInfo.NumGoroutine)
```

### DatabaseStats

```go
type DatabaseStats struct {
    OpenConnections   int
    InUse             int
    Idle              int
    WaitCount         int64
    WaitDuration      time.Duration
    MaxIdleClosed     int64
    MaxLifetimeClosed int64
}
```

Database connection pool statistics.

**Example:**
```go
stats := ctx.DB().Stats()
fmt.Printf("Active connections: %d/%d\n", stats.InUse, stats.OpenConnections)
```

---

## Plugin Types

### PluginManifest

```go
type PluginManifest struct {
    Name         string
    Version      string
    Description  string
    Author       string
    License      string
    Dependencies []string
    Permissions  []string
    Config       map[string]interface{}
}
```

Plugin metadata and configuration.

**Example:**
```go
manifest := pkg.PluginManifest{
    Name:        "my-plugin",
    Version:     "1.0.0",
    Description: "My awesome plugin",
    Author:      "John Doe",
    License:     "MIT",
    Dependencies: []string{"other-plugin"},
    Permissions: []string{"database:read", "network:http"},
}
```

### HookType

```go
type HookType string

const (
    HookTypeStartup    HookType = "startup"
    HookTypeShutdown   HookType = "shutdown"
    HookTypePreRequest HookType = "pre_request"
    HookTypePostRequest HookType = "post_request"
    HookTypePreResponse HookType = "pre_response"
    HookTypePostResponse HookType = "post_response"
)
```

Plugin lifecycle hook types.

**Example:**
```go
func (p *MyPlugin) RegisterHooks(ctx pkg.PluginContext) error {
    return ctx.HookSystem().RegisterHook(
        p.Name(),
        pkg.HookTypeStartup,
        100,
        p.onStartup,
    )
}
```

---

## Best Practices

### Type Safety

Always use the provided types instead of primitives:

```go
// Good
config := pkg.ServerConfig{
    ReadTimeout: 10 * time.Second,
}

// Bad
config := map[string]interface{}{
    "read_timeout": "10s",
}
```

### Configuration Validation

Validate configuration before use:

```go
config := pkg.DatabaseConfig{
    Driver: "postgres",
    Database: "myapp",
}

if err := config.Validate(); err != nil {
    return fmt.Errorf("invalid config: %w", err)
}
```

### Struct Embedding

Use struct embedding for extending types:

```go
type CustomSession struct {
    pkg.Session
    CustomField string
}
```

### JSON Serialization

All models support JSON serialization:

```go
session := &pkg.Session{
    ID: "sess_123",
    UserID: "user_456",
}

data, err := json.Marshal(session)
```

## See Also

- [Framework API](framework.md) - Framework initialization
- [Configuration API](configuration.md) - Configuration management
- [Database API](database.md) - Data models and persistence
- [Security API](security.md) - Security types and authentication

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0


## Middleware Types

### MiddlewareConfig

```go
type MiddlewareConfig struct {
    Name          string
    Handler       MiddlewareFunc
    Priority      int
    Enabled       bool
    Routes        []string
    ExcludeRoutes []string
    Methods       []string
    Condition     ConditionFunc
}
```

Configuration for middleware registration.

**Fields:**
- `Name`: Unique middleware identifier
- `Handler`: Middleware function to execute
- `Priority`: Execution order (lower = earlier)
- `Enabled`: Whether middleware is active
- `Routes`: Route patterns to apply to (supports wildcards)
- `ExcludeRoutes`: Route patterns to exclude
- `Methods`: HTTP methods to apply to (empty = all)
- `Condition`: Custom condition function

**Priority Guidelines:**
- 0-99: Infrastructure (CORS, security headers)
- 100-199: Authentication
- 200-299: Authorization
- 300-399: Request processing
- 400-499: Business logic
- 500+: Response processing

**Example:**
```go
config := pkg.MiddlewareConfig{
    Name: "auth",
    Handler: authMiddleware,
    Priority: 100,
    Enabled: true,
    Routes: []string{"/api/*"},
    ExcludeRoutes: []string{"/api/public/*"},
    Methods: []string{"GET", "POST"},
    Condition: func(ctx pkg.Context) bool {
        return ctx.GetHeader("X-Require-Auth") == "true"
    },
}

app.MiddlewareEngine().Register(config)
```

### ConditionFunc

```go
type ConditionFunc func(ctx Context) bool
```

Function type for conditional middleware execution.

**Example:**
```go
condition := func(ctx pkg.Context) bool {
    // Only apply to non-admin users
    user := ctx.User()
    return user != nil && !user.IsAdmin
}
```

---

## Pipeline Types

### PipelineConfig

```go
type PipelineConfig struct {
    Name       string
    Stages     []PipelineStage
    Priority   int
    Enabled    bool
    Timeout    time.Duration
    OnError    ErrorHandler
    OnComplete CompleteHandler
}
```

Configuration for pipeline registration.

**Fields:**
- `Name`: Unique pipeline identifier
- `Stages`: Pipeline stages to execute
- `Priority`: Execution priority
- `Enabled`: Whether pipeline is active
- `Timeout`: Maximum execution time
- `OnError`: Error handler function
- `OnComplete`: Completion handler function

**Example:**
```go
config := pkg.PipelineConfig{
    Name: "order-processing",
    Timeout: 30 * time.Second,
    Stages: []pkg.PipelineStage{
        {Name: "validate", Handler: validateOrder, Required: true},
        {Name: "payment", Handler: processPayment, Required: true},
        {Name: "notify", Handler: sendNotification, Required: false},
    },
    OnError: func(ctx pkg.Context, err error) error {
        ctx.Logger().Error("Pipeline failed", "error", err)
        return err
    },
}
```

### PipelineStage

```go
type PipelineStage struct {
    Name       string
    Handler    HandlerFunc
    Timeout    time.Duration
    Required   bool
    Retry      int
    RetryDelay time.Duration
}
```

Individual stage in a pipeline.

**Fields:**
- `Name`: Stage identifier
- `Handler`: Stage handler function
- `Timeout`: Stage timeout
- `Required`: Whether stage must succeed
- `Retry`: Number of retries on failure
- `RetryDelay`: Delay between retries

**Example:**
```go
stage := pkg.PipelineStage{
    Name: "payment",
    Handler: processPayment,
    Timeout: 10 * time.Second,
    Required: true,
    Retry: 3,
    RetryDelay: 1 * time.Second,
}
```

---

## Error Types

### FrameworkError

```go
type FrameworkError struct {
    Code       string
    Message    string
    StatusCode int
    Cause      error
    I18nKey    string
    I18nParams map[string]interface{}
}
```

Structured error type for framework errors.

**Fields:**
- `Code`: Error code (e.g., "VALIDATION_ERROR")
- `Message`: Human-readable error message
- `StatusCode`: HTTP status code
- `Cause`: Underlying error (optional)
- `I18nKey`: Internationalization key
- `I18nParams`: I18n parameters

**Methods:**
```go
func (e *FrameworkError) Error() string
func (e *FrameworkError) WithCause(cause error) *FrameworkError
func (e *FrameworkError) WithI18n(key string, params map[string]interface{}) *FrameworkError
```

**Example:**
```go
err := pkg.NewFrameworkError("INVALID_INPUT", "Invalid email format", 400)
err.WithCause(originalErr)
err.WithI18n("errors.invalid_email", map[string]interface{}{
    "field": "email",
})

return err
```

### Common Error Constructors

```go
func NewValidationError(message string) *FrameworkError
func NewAuthenticationError(message string) *FrameworkError
func NewAuthorizationError(message string) *FrameworkError
func NewNotFoundError(resource string) *FrameworkError
func NewConflictError(message string) *FrameworkError
func NewRateLimitError(retryAfter time.Duration) *FrameworkError
func NewInternalError(message string) *FrameworkError
```

**Example:**
```go
// Validation error
return pkg.NewValidationError("Email is required")

// Authentication error
return pkg.NewAuthenticationError("Invalid credentials")

// Authorization error
return pkg.NewAuthorizationError("Insufficient permissions")

// Not found error
return pkg.NewNotFoundError("User")

// Rate limit error
return pkg.NewRateLimitError(60 * time.Second)
```

---

## Proxy Types

### ProxyConfig

```go
type ProxyConfig struct {
    Backends           []*Backend
    LoadBalancer       string
    HealthCheckPath    string
    HealthCheckInterval time.Duration
    CircuitBreakerThreshold int
    CircuitBreakerTimeout   time.Duration
    ConnectionPoolSize      int
    RequestTimeout          time.Duration
}
```

Proxy and load balancing configuration.

**Defaults:**
- `LoadBalancer`: "round-robin"
- `HealthCheckPath`: "/health"
- `HealthCheckInterval`: 30 seconds
- `CircuitBreakerThreshold`: 5
- `CircuitBreakerTimeout`: 60 seconds
- `ConnectionPoolSize`: 100
- `RequestTimeout`: 30 seconds

**Load Balancer Strategies:**
- `round-robin`: Distribute requests evenly
- `least-connections`: Route to backend with fewest connections
- `ip-hash`: Route based on client IP
- `random`: Random backend selection

**Example:**
```go
config := pkg.ProxyConfig{
    Backends: []*pkg.Backend{
        {URL: "http://backend1:8080", Weight: 1},
        {URL: "http://backend2:8080", Weight: 2},
    },
    LoadBalancer: "round-robin",
    HealthCheckInterval: 30 * time.Second,
    CircuitBreakerThreshold: 5,
}
```

### Backend

```go
type Backend struct {
    ID       string
    URL      string
    Weight   int
    Healthy  bool
    LastCheck time.Time
}
```

Backend server configuration.

**Fields:**
- `ID`: Unique backend identifier
- `URL`: Backend server URL
- `Weight`: Load balancing weight (higher = more traffic)
- `Healthy`: Whether backend is healthy
- `LastCheck`: Last health check timestamp

**Example:**
```go
backend := &pkg.Backend{
    ID: "backend-1",
    URL: "http://localhost:8081",
    Weight: 1,
    Healthy: true,
}

app.Proxy().AddBackend(backend)
```

---

## REST API Types

### RESTRouteConfig

```go
type RESTRouteConfig struct {
    RateLimit      int
    RateLimitWindow time.Duration
    RequireAuth    bool
    RequiredScopes []string
    CacheEnabled   bool
    CacheTTL       time.Duration
    Timeout        time.Duration
}
```

REST API route configuration.

**Example:**
```go
config := pkg.RESTRouteConfig{
    RateLimit: 100,
    RateLimitWindow: 1 * time.Minute,
    RequireAuth: true,
    RequiredScopes: []string{"read:users"},
    CacheEnabled: true,
    CacheTTL: 5 * time.Minute,
    Timeout: 10 * time.Second,
}
```

### RESTHandler

```go
type RESTHandler func(ctx Context) (interface{}, error)
```

REST API handler function type.

**Returns:**
- `interface{}`: Response data (automatically serialized to JSON)
- `error`: Error if handler fails

**Example:**
```go
func listUsers(ctx pkg.Context) (interface{}, error) {
    users, err := getUsersFromDB(ctx)
    if err != nil {
        return nil, err
    }
    
    return users, nil
}
```

---

## GraphQL Types

### GraphQLConfig

```go
type GraphQLConfig struct {
    Path           string
    PlaygroundPath string
    EnablePlayground bool
    MaxDepth       int
    MaxComplexity  int
    Timeout        time.Duration
}
```

GraphQL configuration.

**Defaults:**
- `Path`: "/graphql"
- `PlaygroundPath`: "/playground"
- `EnablePlayground`: false
- `MaxDepth`: 10
- `MaxComplexity`: 1000
- `Timeout`: 30 seconds

**Example:**
```go
config := pkg.GraphQLConfig{
    Path: "/graphql",
    EnablePlayground: true,
    MaxDepth: 15,
    MaxComplexity: 2000,
}
```

---

## gRPC Types

### GRPCConfig

```go
type GRPCConfig struct {
    Port              int
    MaxRecvMsgSize    int
    MaxSendMsgSize    int
    ConnectionTimeout time.Duration
    EnableReflection  bool
    EnableTLS         bool
    TLSCertFile       string
    TLSKeyFile        string
}
```

gRPC server configuration.

**Defaults:**
- `Port`: 50051
- `MaxRecvMsgSize`: 4MB
- `MaxSendMsgSize`: 4MB
- `ConnectionTimeout`: 120 seconds
- `EnableReflection`: false
- `EnableTLS`: false

**Example:**
```go
config := pkg.GRPCConfig{
    Port: 50051,
    MaxRecvMsgSize: 10 * 1024 * 1024, // 10MB
    EnableReflection: true,
    EnableTLS: true,
    TLSCertFile: "cert.pem",
    TLSKeyFile: "key.pem",
}
```

---

## WebSocket Types

### WebSocketConfig

```go
type WebSocketConfig struct {
    ReadBufferSize  int
    WriteBufferSize int
    HandshakeTimeout time.Duration
    EnableCompression bool
    CheckOrigin      func(r *http.Request) bool
}
```

WebSocket configuration.

**Defaults:**
- `ReadBufferSize`: 1024 bytes
- `WriteBufferSize`: 1024 bytes
- `HandshakeTimeout`: 10 seconds
- `EnableCompression`: false
- `CheckOrigin`: Allow all origins

**Example:**
```go
config := pkg.WebSocketConfig{
    ReadBufferSize: 4096,
    WriteBufferSize: 4096,
    EnableCompression: true,
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://myapp.com"
    },
}
```

---

## I18n Types

### I18nConfig

```go
type I18nConfig struct {
    DefaultLanguage string
    LocalesPath     string
    SupportedLanguages []string
    FallbackLanguage string
}
```

Internationalization configuration.

**Defaults:**
- `DefaultLanguage`: "en"
- `LocalesPath`: "./locales"
- `FallbackLanguage`: "en"

**Example:**
```go
config := pkg.I18nConfig{
    DefaultLanguage: "en",
    LocalesPath: "./locales",
    SupportedLanguages: []string{"en", "de", "fr", "es"},
    FallbackLanguage: "en",
}
```

---

## Metrics Types

### RequestMetrics

```go
type RequestMetrics struct {
    RequestID    string
    StartTime    time.Time
    EndTime      time.Time
    Duration     time.Duration
    StatusCode   int
    Method       string
    Path         string
    UserID       string
    TenantID     string
    ErrorMessage string
}
```

Request-level metrics.

**Example:**
```go
metrics := ctx.Metrics().Start(requestID)
// ... handle request ...
metrics.EndTime = time.Now()
metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
ctx.Metrics().Record(metrics)
```

### AggregatedMetrics

```go
type AggregatedMetrics struct {
    TotalRequests    int64
    SuccessfulRequests int64
    FailedRequests   int64
    AverageDuration  time.Duration
    MinDuration      time.Duration
    MaxDuration      time.Duration
    RequestsPerSecond float64
}
```

Aggregated metrics for reporting.

**Example:**
```go
metrics, err := ctx.Metrics().GetAggregatedMetrics(tenantID, from, to)
fmt.Printf("RPS: %.2f\n", metrics.RequestsPerSecond)
```

---

## Type Conversion Utilities

### String to Duration

```go
// Parse duration from string
duration, err := time.ParseDuration("10s")
duration, err := time.ParseDuration("5m")
duration, err := time.ParseDuration("1h")
```

### JSON Marshaling

```go
// Marshal to JSON
data, err := json.Marshal(session)

// Unmarshal from JSON
var session pkg.Session
err := json.Unmarshal(data, &session)
```

### Type Assertions

```go
// Safe type assertion
if user, ok := ctx.Get("user"); ok {
    if u, ok := user.(*pkg.User); ok {
        fmt.Println(u.Username)
    }
}
```

---

## See Also

- [Framework API](framework.md) - Framework initialization
- [Context API](context.md) - Request context
- [Configuration API](configuration.md) - Configuration management
- [Database API](database.md) - Data persistence
- [Security API](security.md) - Authentication and authorization

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0  
**Total Types Documented**: 50+
