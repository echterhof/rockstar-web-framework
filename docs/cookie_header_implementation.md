# Cookie and Header Management Implementation

## Overview

The Rockstar Web Framework provides comprehensive cookie and header management functionality with built-in encryption support for sensitive data. This implementation satisfies Requirements 20.1-20.5 from the requirements document.

## Features

### Cookie Management

- **Basic Cookie Operations**: Set, get, and delete cookies through the context
- **Cookie Encryption**: AES-256-GCM encryption for sensitive cookie values
- **Default Configuration**: Configurable defaults for path, domain, security flags
- **Multiple Cookie Support**: Retrieve all cookies from a request
- **Security Flags**: Support for Secure, HttpOnly, and SameSite attributes

### Header Management

- **Request Headers**: Access all request headers through the context
- **Response Headers**: Set, modify, and delete response headers
- **Common Helpers**: Convenience methods for common headers (Content-Type, Authorization, etc.)
- **Multiple Headers**: Set multiple headers at once
- **Header Validation**: Check for header existence

## Architecture

### Cookie Manager

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

### Header Manager

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

## Usage Examples

### Basic Cookie Management

```go
// Create cookie manager
config := pkg.DefaultCookieConfig()
cookieManager, err := pkg.NewCookieManager(config)
if err != nil {
    log.Fatal(err)
}

// Set a cookie
cookie := &pkg.Cookie{
    Name:  "user_preference",
    Value: "dark_mode",
    Path:  "/",
    MaxAge: 3600,
}

err = cookieManager.SetCookie(ctx, cookie)
if err != nil {
    log.Fatal(err)
}

// Get a cookie
cookie, err = cookieManager.GetCookie(ctx, "user_preference")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User preference: %s\n", cookie.Value)

// Delete a cookie
err = cookieManager.DeleteCookie(ctx, "user_preference")
if err != nil {
    log.Fatal(err)
}
```

### Encrypted Cookie Management

```go
// Generate encryption key (32 bytes for AES-256)
encryptionKey := make([]byte, 32)
rand.Read(encryptionKey)

// Create cookie manager with encryption
config := &pkg.CookieConfig{
    EncryptionKey:   encryptionKey,
    DefaultPath:     "/",
    DefaultSecure:   true,
    DefaultHTTPOnly: true,
}

cookieManager, err := pkg.NewCookieManager(config)
if err != nil {
    log.Fatal(err)
}

// Set an encrypted cookie
cookie := &pkg.Cookie{
    Name:  "session_token",
    Value: "sensitive_session_data",
}

err = cookieManager.SetEncryptedCookie(ctx, cookie)
if err != nil {
    log.Fatal(err)
}

// Get and decrypt a cookie
cookie, err = cookieManager.GetEncryptedCookie(ctx, "session_token")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Decrypted session token: %s\n", cookie.Value)
```

### Header Management

```go
// Create header manager
headerManager := pkg.NewHeaderManager()

// Get request headers
userAgent := headerManager.GetUserAgent(ctx)
authorization := headerManager.GetAuthorization(ctx)
contentType := headerManager.GetContentType(ctx)

// Set response headers
err := headerManager.SetContentType(ctx, "application/json")
err = headerManager.SetCacheControl(ctx, "no-cache")
err = headerManager.SetHeader(ctx, "X-Request-ID", "req-123")

// Set multiple headers
headers := map[string]string{
    "X-API-Version": "v1",
    "X-Server":      "Rockstar",
}
err = headerManager.SetHeaders(ctx, headers)

// Check if header exists
if headerManager.HasHeader(ctx, "Authorization") {
    // Process authenticated request
}

// Get all headers
allHeaders := headerManager.GetAllHeaders(ctx)
for key, value := range allHeaders {
    fmt.Printf("%s: %s\n", key, value)
}
```

### Context Integration

The cookie and header management functionality is integrated directly into the Context interface:

```go
// Cookie operations through context
cookie := &pkg.Cookie{
    Name:  "session_id",
    Value: "abc123",
}
ctx.SetCookie(cookie)

cookie, err := ctx.GetCookie("session_id")

// Header operations through context
ctx.SetHeader("X-Custom-Header", "value")
value := ctx.GetHeader("X-Custom-Header")
```

## Configuration

### Cookie Configuration

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

**Default Configuration:**
- Path: `/`
- Secure: `true`
- HttpOnly: `true`
- SameSite: `Lax`
- MaxAge: `86400` (24 hours)

## Security Considerations

### Cookie Encryption

- Uses AES-256-GCM for authenticated encryption
- Generates unique nonce for each encryption operation
- Base64-URL encoding for safe cookie storage
- Prevents tampering through authenticated encryption

### Cookie Security Flags

- **Secure**: Ensures cookies are only sent over HTTPS
- **HttpOnly**: Prevents JavaScript access to cookies
- **SameSite**: Protects against CSRF attacks

### Header Security

- Validates header keys and values
- Supports security headers (X-Frame-Options, CSP, etc.)
- Integrates with SecurityManager for comprehensive protection

## Testing

The implementation includes comprehensive unit tests covering:

- Basic cookie operations (set, get, delete)
- Encrypted cookie operations
- Cookie encryption/decryption
- Default configuration application
- Header operations (set, get, delete)
- Multiple header operations
- Common header helpers
- Error handling
- Context integration

Run tests with:
```bash
go test -v -run "TestCookie|TestHeader" ./pkg
```

## Performance Considerations

### Cookie Operations

- Lazy header parsing for efficient memory usage
- Cookie caching in context for repeated access
- Minimal allocations for cookie operations

### Encryption Performance

- AES-GCM provides fast authenticated encryption
- Single-pass encryption/decryption
- Efficient nonce generation

### Header Operations

- Case-insensitive header access
- Header map caching in context
- Efficient multi-header operations

## Integration with Other Components

### Session Management

The cookie manager integrates with the session management system:

```go
// Session manager uses cookie manager for encrypted session cookies
sessionManager.SetCookie(ctx, session)
session, err := sessionManager.GetSessionFromCookie(ctx)
```

### Security Manager

The security manager can encrypt/decrypt cookies:

```go
// Security manager provides cookie encryption
encryptedValue, err := securityManager.EncryptCookie(value)
decryptedValue, err := securityManager.DecryptCookie(encryptedValue)
```

### Middleware

Cookie and header management can be used in middleware:

```go
func AuthMiddleware(next HandlerFunc) HandlerFunc {
    return func(ctx Context) error {
        // Get authorization header
        auth := ctx.GetHeader("Authorization")
        
        // Validate and set cookie
        if valid {
            cookie := &Cookie{Name: "auth_token", Value: token}
            ctx.SetCookie(cookie)
        }
        
        return next(ctx)
    }
}
```

## Requirements Validation

This implementation satisfies the following requirements:

- **20.1**: Request cookie access through context ✓
- **20.2**: Cookie creation and sending via context ✓
- **20.3**: Cookie encryption support ✓
- **20.4**: Header data access through context ✓
- **20.5**: Header data sending through context ✓

## Future Enhancements

Potential improvements for future versions:

1. **Cookie Signing**: Add HMAC-based cookie signing for integrity
2. **Cookie Compression**: Compress large cookie values
3. **Cookie Prefixes**: Support __Secure- and __Host- prefixes
4. **Header Validation**: More comprehensive header validation
5. **Header Compression**: Support for header compression (HPACK)
6. **Cookie Store**: Persistent cookie storage for testing
7. **Header Middleware**: Built-in middleware for common headers

## Troubleshooting

### Common Issues

**Issue**: Encrypted cookies fail to decrypt
- **Solution**: Ensure the same encryption key is used for encryption and decryption
- **Solution**: Verify the key is exactly 32 bytes for AES-256

**Issue**: Cookies not being set
- **Solution**: Check that response headers haven't been written yet
- **Solution**: Verify cookie configuration (path, domain, etc.)

**Issue**: Headers not accessible
- **Solution**: Ensure headers are accessed before response is written
- **Solution**: Check for case-sensitivity in header names

## References

- [RFC 6265 - HTTP State Management Mechanism (Cookies)](https://tools.ietf.org/html/rfc6265)
- [RFC 7231 - HTTP/1.1 Semantics and Content (Headers)](https://tools.ietf.org/html/rfc7231)
- [OWASP - Secure Cookie Attribute](https://owasp.org/www-community/controls/SecureCookieAttribute)
- [SameSite Cookie Attribute](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite)
