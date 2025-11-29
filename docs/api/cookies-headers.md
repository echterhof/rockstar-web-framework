# Cookies and Headers API Reference

## Overview

The Rockstar Web Framework provides comprehensive cookie and header management with built-in encryption support for cookies, security defaults, and convenient helper methods for common operations.

## Cookie Management

### CookieManager Interface

Interface for managing cookies with encryption support.

```go
type CookieManager interface {
    // Cookie operations
    SetCookie(ctx Context, cookie *Cookie) error
    GetCookie(ctx Context, name string) (*Cookie, error)
    GetAllCookies(ctx Context) ([]*Cookie, error)
    DeleteCookie(ctx Context, name string) error
    
    // Encrypted cookie operations
    SetEncryptedCookie(ctx Context, cookie *Cookie) error
    GetEncryptedCookie(ctx Context, name string) (*Cookie, error)
    
    // Cookie value encryption/decryption
    EncryptValue(value string) (string, error)
    DecryptValue(encryptedValue string) (string, error)
}
```

### CookieConfig

Configuration for cookie manager.

```go
type CookieConfig struct {
    EncryptionKey   []byte        // AES-256 key (32 bytes)
    DefaultPath     string        // Default cookie path
    DefaultDomain   string        // Default cookie domain
    DefaultSecure   bool          // Default secure flag
    DefaultHTTPOnly bool          // Default HTTP-only flag
    DefaultSameSite http.SameSite // Default SameSite policy
    DefaultMaxAge   int           // Default max age in seconds
}
```

### DefaultCookieConfig()

Returns default cookie configuration with secure defaults.

**Signature:**
```go
func DefaultCookieConfig() *CookieConfig
```

**Returns:**
- `*CookieConfig` - Default configuration

**Defaults:**
- Path: `/`
- Secure: `true`
- HttpOnly: `true`
- SameSite: `http.SameSiteLaxMode`
- MaxAge: `86400` (24 hours)

**Example:**
```go
config := pkg.DefaultCookieConfig()
config.EncryptionKey = []byte("your-32-byte-encryption-key-here")
config.DefaultMaxAge = 3600 // 1 hour

manager, err := pkg.NewCookieManager(config)
```

### NewCookieManager()

Creates a new cookie manager instance.

**Signature:**
```go
func NewCookieManager(config *CookieConfig) (CookieManager, error)
```

**Parameters:**
- `config` - Cookie configuration (uses defaults if nil)

**Returns:**
- `CookieManager` - Cookie manager instance
- `error` - Error if encryption key is invalid

**Example:**
```go
config := pkg.DefaultCookieConfig()
config.EncryptionKey, _ = pkg.GenerateEncryptionKey(32)

manager, err := pkg.NewCookieManager(config)
if err != nil {
    log.Fatal(err)
}
```

### SetCookie()

Sets a cookie in the response with default values applied.

**Signature:**
```go
SetCookie(ctx Context, cookie *Cookie) error
```

**Parameters:**
- `ctx` - Request context
- `cookie` - Cookie to set

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.POST("/login", func(ctx pkg.Context) error {
    manager, _ := pkg.NewCookieManager(nil)
    
    cookie := &pkg.Cookie{
        Name:  "session_id",
        Value: "abc123xyz",
        MaxAge: 3600, // 1 hour
    }
    
    err := manager.SetCookie(ctx, cookie)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "logged in"})
})
```

### GetCookie()

Retrieves a cookie from the request.

**Signature:**
```go
GetCookie(ctx Context, name string) (*Cookie, error)
```

**Parameters:**
- `ctx` - Request context
- `name` - Cookie name

**Returns:**
- `*Cookie` - Cookie object
- `error` - Error if cookie not found

**Example:**
```go
router.GET("/profile", func(ctx pkg.Context) error {
    manager, _ := pkg.NewCookieManager(nil)
    
    cookie, err := manager.GetCookie(ctx, "session_id")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    sessionID := cookie.Value
    // Validate session
    
    return ctx.JSON(200, map[string]interface{}{"session": sessionID})
})
```

### GetAllCookies()

Retrieves all cookies from the request.

**Signature:**
```go
GetAllCookies(ctx Context) ([]*Cookie, error)
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `[]*Cookie` - Array of cookies
- `error` - Error if operation fails

**Example:**
```go
router.GET("/debug/cookies", func(ctx pkg.Context) error {
    manager, _ := pkg.NewCookieManager(nil)
    
    cookies, err := manager.GetAllCookies(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to get cookies"})
    }
    
    cookieNames := []string{}
    for _, cookie := range cookies {
        cookieNames = append(cookieNames, cookie.Name)
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "count":   len(cookies),
        "cookies": cookieNames,
    })
})
```

### DeleteCookie()

Deletes a cookie by setting it to expire immediately.

**Signature:**
```go
DeleteCookie(ctx Context, name string) error
```

**Parameters:**
- `ctx` - Request context
- `name` - Cookie name to delete

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.POST("/logout", func(ctx pkg.Context) error {
    manager, _ := pkg.NewCookieManager(nil)
    
    err := manager.DeleteCookie(ctx, "session_id")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to delete cookie"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "logged out"})
})
```

### SetEncryptedCookie()

Sets an encrypted cookie using AES-256-GCM encryption.

**Signature:**
```go
SetEncryptedCookie(ctx Context, cookie *Cookie) error
```

**Parameters:**
- `ctx` - Request context
- `cookie` - Cookie to encrypt and set

**Returns:**
- `error` - Error if encryption not configured or operation fails

**Example:**
```go
router.POST("/login", func(ctx pkg.Context) error {
    config := pkg.DefaultCookieConfig()
    config.EncryptionKey, _ = pkg.GenerateEncryptionKey(32)
    manager, _ := pkg.NewCookieManager(config)
    
    // Sensitive data will be encrypted
    cookie := &pkg.Cookie{
        Name:  "user_data",
        Value: `{"id":123,"email":"user@example.com"}`,
        MaxAge: 3600,
    }
    
    err := manager.SetEncryptedCookie(ctx, cookie)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "logged in"})
})
```

### GetEncryptedCookie()

Retrieves and decrypts an encrypted cookie.

**Signature:**
```go
GetEncryptedCookie(ctx Context, name string) (*Cookie, error)
```

**Parameters:**
- `ctx` - Request context
- `name` - Cookie name

**Returns:**
- `*Cookie` - Cookie with decrypted value
- `error` - Error if decryption fails

**Example:**
```go
router.GET("/profile", func(ctx pkg.Context) error {
    config := pkg.DefaultCookieConfig()
    config.EncryptionKey = []byte("your-32-byte-key")
    manager, _ := pkg.NewCookieManager(config)
    
    cookie, err := manager.GetEncryptedCookie(ctx, "user_data")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid session"})
    }
    
    // cookie.Value is now decrypted
    userData := cookie.Value
    
    return ctx.JSON(200, map[string]interface{}{"data": userData})
})
```

### EncryptValue()

Encrypts a cookie value using AES-256-GCM.

**Signature:**
```go
EncryptValue(value string) (string, error)
```

**Parameters:**
- `value` - Plain text value to encrypt

**Returns:**
- `string` - Base64-encoded encrypted value
- `error` - Error if encryption fails

**Example:**
```go
config := pkg.DefaultCookieConfig()
config.EncryptionKey, _ = pkg.GenerateEncryptionKey(32)
manager, _ := pkg.NewCookieManager(config)

encrypted, err := manager.EncryptValue("sensitive-data")
if err != nil {
    log.Fatal(err)
}
// Store encrypted value
```

### DecryptValue()

Decrypts an encrypted cookie value.

**Signature:**
```go
DecryptValue(encryptedValue string) (string, error)
```

**Parameters:**
- `encryptedValue` - Base64-encoded encrypted value

**Returns:**
- `string` - Decrypted plain text value
- `error` - Error if decryption fails

**Example:**
```go
config := pkg.DefaultCookieConfig()
config.EncryptionKey = []byte("your-32-byte-key")
manager, _ := pkg.NewCookieManager(config)

decrypted, err := manager.DecryptValue(encryptedValue)
if err != nil {
    log.Fatal(err)
}
```

## Header Management

### HeaderManager Interface

Interface for managing HTTP headers.

```go
type HeaderManager interface {
    // Request header operations
    GetHeader(ctx Context, key string) string
    GetAllHeaders(ctx Context) map[string]string
    GetHeaderValues(ctx Context, key string) []string
    HasHeader(ctx Context, key string) bool
    
    // Response header operations
    SetHeader(ctx Context, key, value string) error
    SetHeaders(ctx Context, headers map[string]string) error
    AddHeader(ctx Context, key, value string) error
    DeleteHeader(ctx Context, key string) error
    
    // Common header helpers
    SetContentType(ctx Context, contentType string) error
    SetCacheControl(ctx Context, directive string) error
    SetLocation(ctx Context, url string) error
    SetAuthorization(ctx Context, token string) error
    GetAuthorization(ctx Context) string
    GetUserAgent(ctx Context) string
    GetReferer(ctx Context) string
    GetContentType(ctx Context) string
}
```

### NewHeaderManager()

Creates a new header manager instance.

**Signature:**
```go
func NewHeaderManager() HeaderManager
```

**Returns:**
- `HeaderManager` - Header manager instance

**Example:**
```go
manager := pkg.NewHeaderManager()
```

### GetHeader()

Retrieves a request header value (case-insensitive).

**Signature:**
```go
GetHeader(ctx Context, key string) string
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name

**Returns:**
- `string` - Header value, or empty string if not found

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    auth := manager.GetHeader(ctx, "Authorization")
    userAgent := manager.GetHeader(ctx, "User-Agent")
    
    return ctx.JSON(200, map[string]interface{}{
        "has_auth": auth != "",
        "user_agent": userAgent,
    })
})
```

### GetAllHeaders()

Retrieves all request headers as a map.

**Signature:**
```go
GetAllHeaders(ctx Context) map[string]string
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `map[string]string` - Map of header names to values

**Example:**
```go
router.GET("/debug/headers", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    headers := manager.GetAllHeaders(ctx)
    
    return ctx.JSON(200, headers)
})
```

### GetHeaderValues()

Retrieves all values for a request header (for headers with multiple values).

**Signature:**
```go
GetHeaderValues(ctx Context, key string) []string
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name

**Returns:**
- `[]string` - Array of header values

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    acceptTypes := manager.GetHeaderValues(ctx, "Accept")
    for _, acceptType := range acceptTypes {
        fmt.Println(acceptType)
    }
    
    return ctx.JSON(200, map[string]interface{}{"accept": acceptTypes})
})
```

### HasHeader()

Checks if a request header exists.

**Signature:**
```go
HasHeader(ctx Context, key string) bool
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name

**Returns:**
- `bool` - true if header exists, false otherwise

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    if !manager.HasHeader(ctx, "Authorization") {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### SetHeader()

Sets a response header.

**Signature:**
```go
SetHeader(ctx Context, key, value string) error
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name
- `value` - Header value

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    manager.SetHeader(ctx, "X-Custom-Header", "CustomValue")
    manager.SetHeader(ctx, "X-Request-ID", "req-123")
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### SetHeaders()

Sets multiple response headers at once.

**Signature:**
```go
SetHeaders(ctx Context, headers map[string]string) error
```

**Parameters:**
- `ctx` - Request context
- `headers` - Map of header names to values

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    headers := map[string]string{
        "X-Custom-Header": "CustomValue",
        "X-Request-ID":    "req-123",
        "X-Version":       "1.0",
    }
    
    manager.SetHeaders(ctx, headers)
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### AddHeader()

Adds a response header (allows multiple values for the same header).

**Signature:**
```go
AddHeader(ctx Context, key, value string) error
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name
- `value` - Header value to add

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    // Add multiple Set-Cookie headers
    manager.AddHeader(ctx, "Set-Cookie", "session=abc; Path=/")
    manager.AddHeader(ctx, "Set-Cookie", "user=123; Path=/")
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### DeleteHeader()

Deletes a response header.

**Signature:**
```go
DeleteHeader(ctx Context, key string) error
```

**Parameters:**
- `ctx` - Request context
- `key` - Header name to delete

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    manager.DeleteHeader(ctx, "X-Powered-By")
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### SetContentType()

Sets the Content-Type header.

**Signature:**
```go
SetContentType(ctx Context, contentType string) error
```

**Parameters:**
- `ctx` - Request context
- `contentType` - Content type value

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/data.xml", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    manager.SetContentType(ctx, "application/xml")
    
    return ctx.String(200, "<data>value</data>")
})
```

### SetCacheControl()

Sets the Cache-Control header.

**Signature:**
```go
SetCacheControl(ctx Context, directive string) error
```

**Parameters:**
- `ctx` - Request context
- `directive` - Cache control directive

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    // No caching
    manager.SetCacheControl(ctx, "no-cache, no-store, must-revalidate")
    
    // Or cache for 1 hour
    // manager.SetCacheControl(ctx, "public, max-age=3600")
    
    return ctx.JSON(200, map[string]string{"data": "value"})
})
```

### SetLocation()

Sets the Location header (used for redirects).

**Signature:**
```go
SetLocation(ctx Context, url string) error
```

**Parameters:**
- `ctx` - Request context
- `url` - Redirect URL

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
router.GET("/old-path", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    manager.SetLocation(ctx, "/new-path")
    
    return ctx.String(301, "")
})
```

### SetAuthorization()

Sets the Authorization header (for outgoing requests).

**Signature:**
```go
SetAuthorization(ctx Context, token string) error
```

**Parameters:**
- `ctx` - Request context
- `token` - Authorization token

**Returns:**
- `error` - Error if operation fails

**Example:**
```go
manager := pkg.NewHeaderManager()
manager.SetAuthorization(ctx, "Bearer abc123xyz")
```

### GetAuthorization()

Retrieves the Authorization header from the request.

**Signature:**
```go
GetAuthorization(ctx Context) string
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `string` - Authorization header value

**Example:**
```go
router.GET("/api/protected", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    auth := manager.GetAuthorization(ctx)
    if auth == "" {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Validate token
    token := strings.TrimPrefix(auth, "Bearer ")
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### GetUserAgent()

Retrieves the User-Agent header.

**Signature:**
```go
GetUserAgent(ctx Context) string
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `string` - User-Agent header value

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    userAgent := manager.GetUserAgent(ctx)
    
    // Log user agent for analytics
    log.Printf("User-Agent: %s", userAgent)
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### GetReferer()

Retrieves the Referer header.

**Signature:**
```go
GetReferer(ctx Context) string
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `string` - Referer header value

**Example:**
```go
router.POST("/api/action", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    referer := manager.GetReferer(ctx)
    
    // Validate referer for CSRF protection
    if !strings.HasPrefix(referer, "https://yourdomain.com") {
        return ctx.JSON(403, map[string]string{"error": "Invalid referer"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### GetContentType()

Retrieves the Content-Type header.

**Signature:**
```go
GetContentType(ctx Context) string
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `string` - Content-Type header value

**Example:**
```go
router.POST("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    contentType := manager.GetContentType(ctx)
    
    if !strings.Contains(contentType, "application/json") {
        return ctx.JSON(400, map[string]string{"error": "JSON required"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

## Complete Examples

### Secure Session Management

```go
router.POST("/login", func(ctx pkg.Context) error {
    // Configure cookie manager with encryption
    config := pkg.DefaultCookieConfig()
    config.EncryptionKey, _ = pkg.GenerateEncryptionKey(32)
    config.DefaultMaxAge = 3600 // 1 hour
    config.DefaultSecure = true
    config.DefaultHTTPOnly = true
    config.DefaultSameSite = http.SameSiteStrictMode
    
    manager, _ := pkg.NewCookieManager(config)
    
    // Authenticate user
    // ...
    
    // Set encrypted session cookie
    sessionData := fmt.Sprintf(`{"user_id":%d,"email":"%s"}`, userID, email)
    cookie := &pkg.Cookie{
        Name:  "session",
        Value: sessionData,
    }
    
    err := manager.SetEncryptedCookie(ctx, cookie)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "logged in"})
})

router.GET("/profile", func(ctx pkg.Context) error {
    config := pkg.DefaultCookieConfig()
    config.EncryptionKey = []byte("your-32-byte-key")
    manager, _ := pkg.NewCookieManager(config)
    
    // Get encrypted session
    cookie, err := manager.GetEncryptedCookie(ctx, "session")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Parse session data
    sessionData := cookie.Value
    
    return ctx.JSON(200, map[string]interface{}{"session": sessionData})
})
```

### Custom Headers for API

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    // Set custom headers
    headers := map[string]string{
        "X-API-Version":    "1.0",
        "X-Request-ID":     generateRequestID(),
        "X-Rate-Limit":     "100",
        "X-Rate-Remaining": "95",
    }
    manager.SetHeaders(ctx, headers)
    
    // Set cache control
    manager.SetCacheControl(ctx, "public, max-age=300")
    
    return ctx.JSON(200, map[string]string{"data": "value"})
})
```

### CORS Headers

```go
router.OPTIONS("/api/*", func(ctx pkg.Context) error {
    manager := pkg.NewHeaderManager()
    
    // Set CORS headers
    manager.SetHeader(ctx, "Access-Control-Allow-Origin", "*")
    manager.SetHeader(ctx, "Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    manager.SetHeader(ctx, "Access-Control-Allow-Headers", "Content-Type, Authorization")
    manager.SetHeader(ctx, "Access-Control-Max-Age", "86400")
    
    return ctx.String(204, "")
})
```

## Best Practices

### Cookies

1. **Use Encryption:** Always encrypt sensitive cookie data
2. **Set Secure Flags:** Use Secure, HttpOnly, and SameSite flags
3. **Limit Lifetime:** Set appropriate MaxAge for cookies
4. **Validate Input:** Validate cookie values before use
5. **Use HTTPS:** Always use HTTPS in production for secure cookies
6. **Rotate Keys:** Rotate encryption keys periodically
7. **Delete on Logout:** Always delete session cookies on logout

### Headers

1. **Set Security Headers:** Use security headers (CSP, X-Frame-Options, etc.)
2. **Cache Control:** Set appropriate cache headers
3. **Content Type:** Always set correct Content-Type
4. **CORS:** Configure CORS headers properly
5. **Rate Limiting:** Include rate limit headers
6. **Request ID:** Add request ID for tracing
7. **Remove Sensitive:** Remove headers that expose server info

## Security Considerations

1. **Cookie Encryption:** Use AES-256 for cookie encryption
2. **Key Management:** Store encryption keys securely
3. **HTTPS Only:** Set Secure flag for production cookies
4. **SameSite:** Use SameSite=Strict or Lax to prevent CSRF
5. **HttpOnly:** Set HttpOnly to prevent XSS attacks
6. **Path Scope:** Limit cookie path to minimum required
7. **Domain Scope:** Set appropriate domain for cookies
8. **Header Injection:** Validate header values to prevent injection
9. **CORS:** Configure CORS carefully to prevent unauthorized access

## See Also

- [Context API](context.md) - Context cookie and header methods
- [Security Guide](../guides/security.md) - Security best practices
- [Session API](session.md) - Session management
- [Utilities API](utilities.md) - Encryption utilities
