# Session Management

## Overview

The Rockstar Web Framework provides a comprehensive session management system with support for multiple storage backends, encrypted cookies, and automatic cleanup of expired sessions. Sessions enable you to maintain state across multiple requests from the same user.

**Key Features:**
- Multiple storage backends (database, cache, filesystem)
- AES-256 encrypted session cookies
- Automatic session expiration and cleanup
- Session data encryption
- Multi-tenant session support
- Configurable session lifetime
- IP address and user agent tracking

## Configuration

### Basic Configuration

Configure sessions in your `FrameworkConfig`:

```go
config := pkg.FrameworkConfig{
    SessionConfig: pkg.SessionConfig{
        StorageType:     pkg.SessionStorageDatabase,
        CookieName:      "rockstar_session",
        CookieSecure:    true,  // HTTPS only
        CookieHTTPOnly:  true,  // No JavaScript access
        SessionLifetime: 24 * time.Hour,
        EncryptionKey:   encryptionKey,  // 32 bytes for AES-256
    },
}
```

### Encryption Key

**IMPORTANT:** Never hardcode encryption keys in your code!

Generate a secure encryption key:

```bash
# Using the framework's key generator
go run examples/generate_keys.go

# Or using OpenSSL
openssl rand -hex 32
```

Load from environment variable:

```go
import (
    "encoding/hex"
    "os"
)

// Load encryption key from environment
encryptionKeyHex := os.Getenv("SESSION_ENCRYPTION_KEY")
if encryptionKeyHex == "" {
    log.Fatal("SESSION_ENCRYPTION_KEY not set")
}

encryptionKey, err := hex.DecodeString(encryptionKeyHex)
if err != nil {
    log.Fatal("Invalid encryption key format")
}

if len(encryptionKey) != 32 {
    log.Fatal("Encryption key must be 32 bytes for AES-256")
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `StorageType` | SessionStorageType | "database" | Storage backend: "database", "cache", or "filesystem" |
| `CookieName` | string | "rockstar_session" | Name of the session cookie |
| `CookiePath` | string | "/" | Path scope for the cookie |
| `CookieDomain` | string | "" | Domain scope for the cookie |
| `CookieSecure` | bool | true | Send cookie only over HTTPS |
| `CookieHTTPOnly` | bool | true | Prevent JavaScript access to cookie |
| `CookieSameSite` | string | "Lax" | SameSite attribute: "Strict", "Lax", or "None" |
| `SessionLifetime` | duration | 24h | Session expiration time |
| `EncryptionKey` | []byte | *required* | 32-byte AES-256 encryption key |
| `FilesystemPath` | string | "./sessions" | Directory for filesystem storage |
| `CleanupInterval` | duration | 1h | Interval for cleaning expired sessions |

### Storage Backend Options

#### Database Storage (Recommended for Production)

```go
SessionConfig: pkg.SessionConfig{
    StorageType:     pkg.SessionStorageDatabase,
    SessionLifetime: 24 * time.Hour,
    EncryptionKey:   encryptionKey,
    CookieSecure:    true,
    CookieHTTPOnly:  true,
}
```

**Pros:**
- Persistent across server restarts
- Scalable across multiple servers
- Supports complex queries
- Built-in cleanup

**Cons:**
- Requires database connection
- Slightly slower than memory

#### Cache Storage (Fast, Ephemeral)

```go
SessionConfig: pkg.SessionConfig{
    StorageType:     pkg.SessionStorageCache,
    SessionLifetime: 1 * time.Hour,
    EncryptionKey:   encryptionKey,
    CookieSecure:    true,
    CookieHTTPOnly:  true,
}
```

**Pros:**
- Very fast access
- Automatic expiration
- Low overhead

**Cons:**
- Lost on server restart
- Limited by cache size
- Not suitable for long sessions

#### Filesystem Storage (Development)

```go
SessionConfig: pkg.SessionConfig{
    StorageType:     pkg.SessionStorageFilesystem,
    FilesystemPath:  "./sessions",
    SessionLifetime: 24 * time.Hour,
    EncryptionKey:   encryptionKey,
    CookieSecure:    false,  // OK for development
    CookieHTTPOnly:  true,
}
```

**Pros:**
- Simple setup
- Persistent across restarts
- Good for development

**Cons:**
- Not scalable
- File I/O overhead
- Not suitable for production

## Accessing Sessions

Access the session manager through the context in your handlers:

```go
func myHandler(ctx pkg.Context) error {
    // Get session manager
    sessionMgr := ctx.Session()
    
    // Use session operations
    // ...
    
    return nil
}
```

## Session Lifecycle

### Create Session

Create a new session for a user:

```go
func loginHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Authenticate user
    user, err := authenticateUser(username, password)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid credentials"})
    }
    
    // Create new session
    session, err := sessionMgr.Create(ctx)
    if err != nil {
        return err
    }
    
    // Store user data in session
    session.Data["user_id"] = user.ID
    session.Data["username"] = user.Username
    session.Data["roles"] = user.Roles
    session.Data["login_time"] = time.Now()
    
    // Save session
    err = sessionMgr.Save(ctx, session)
    if err != nil {
        return err
    }
    
    // Set encrypted session cookie
    err = sessionMgr.SetCookie(ctx, session)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Login successful",
        "user":    user,
    })
}
```

### Load Session

Load an existing session from a cookie:

```go
func protectedHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Get session from cookie
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Access session data
    userID := session.Data["user_id"]
    username := session.Data["username"]
    
    return ctx.JSON(200, map[string]interface{}{
        "message":  "Access granted",
        "user_id":  userID,
        "username": username,
    })
}
```

### Destroy Session

Destroy a session (logout):

```go
func logoutHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Get session from cookie
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        // No session to destroy
        return ctx.JSON(200, map[string]string{"message": "Logged out"})
    }
    
    // Destroy session
    err = sessionMgr.Destroy(ctx, session.ID)
    if err != nil {
        return err
    }
    
    // Clear cookie
    ctx.SetCookie(&pkg.Cookie{
        Name:   "rockstar_session",
        Value:  "",
        MaxAge: -1,  // Delete cookie
    })
    
    return ctx.JSON(200, map[string]string{"message": "Logged out successfully"})
}
```

### Refresh Session

Extend session expiration:

```go
func refreshSessionHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Get session from cookie
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Refresh session (extends expiration)
    err = sessionMgr.Refresh(ctx, session.ID)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session refreshed"})
}
```

## Session Data Operations

### Set Data

Store data in a session:

```go
func setSessionDataHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Set individual values
    session.Data["preferences"] = map[string]interface{}{
        "theme":    "dark",
        "language": "en",
    }
    session.Data["last_activity"] = time.Now()
    
    // Save session
    err = sessionMgr.Save(ctx, session)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Data saved"})
}
```

**Alternative: Direct Set**
```go
// Set data directly without loading full session
err := sessionMgr.Set(session.ID, "key", "value")
```

### Get Data

Retrieve data from a session:

```go
func getSessionDataHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Get values
    preferences := session.Data["preferences"]
    lastActivity := session.Data["last_activity"]
    
    return ctx.JSON(200, map[string]interface{}{
        "preferences":   preferences,
        "last_activity": lastActivity,
    })
}
```

**Alternative: Direct Get**
```go
// Get data directly without loading full session
value, err := sessionMgr.Get(session.ID, "key")
```

### Delete Data

Remove specific data from a session:

```go
func deleteSessionDataHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Delete specific key
    delete(session.Data, "temporary_data")
    
    // Save session
    err = sessionMgr.Save(ctx, session)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Data deleted"})
}
```

**Alternative: Direct Delete**
```go
// Delete data directly without loading full session
err := sessionMgr.Delete(session.ID, "key")
```

### Clear Data

Clear all data from a session:

```go
func clearSessionDataHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Clear all data
    err = sessionMgr.Clear(session.ID)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session data cleared"})
}
```

## Session Validation

### Check if Valid

Check if a session is valid:

```go
func validateSessionHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]interface{}{
            "valid": false,
            "error": "No session found",
        })
    }
    
    // Check if session is valid
    isValid := sessionMgr.IsValid(session.ID)
    
    return ctx.JSON(200, map[string]interface{}{
        "valid":      isValid,
        "session_id": session.ID,
        "expires_at": session.ExpiresAt,
    })
}
```

### Check if Expired

Check if a session has expired:

```go
func checkExpirationHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "No session found"})
    }
    
    // Check if expired
    isExpired := sessionMgr.IsExpired(session.ID)
    
    if isExpired {
        return ctx.JSON(401, map[string]string{"error": "Session expired"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session is active"})
}
```

## Authentication Middleware

Create middleware to protect routes:

```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    sessionMgr := ctx.Session()
    
    // Get session from cookie
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{
            "error": "Authentication required",
        })
    }
    
    // Check if session is valid
    if !sessionMgr.IsValid(session.ID) {
        return ctx.JSON(401, map[string]string{
            "error": "Session expired",
        })
    }
    
    // Store user info in context for handlers
    if userID, ok := session.Data["user_id"]; ok {
        ctx.Set("user_id", userID)
    }
    
    // Continue to next handler
    return next(ctx)
}

// Apply middleware to routes
router.GET("/api/protected", protectedHandler, authMiddleware)
```

## Session Security

### Secure Cookie Configuration

Always use secure cookies in production:

```go
SessionConfig: pkg.SessionConfig{
    CookieSecure:   true,   // HTTPS only
    CookieHTTPOnly: true,   // No JavaScript access
    CookieSameSite: "Strict", // CSRF protection
}
```

### Session Fixation Prevention

Regenerate session ID after login:

```go
func loginHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Destroy old session if exists
    if oldSession, err := sessionMgr.GetSessionFromCookie(ctx); err == nil {
        sessionMgr.Destroy(ctx, oldSession.ID)
    }
    
    // Create new session after authentication
    session, err := sessionMgr.Create(ctx)
    if err != nil {
        return err
    }
    
    // Store user data
    session.Data["user_id"] = user.ID
    sessionMgr.Save(ctx, session)
    sessionMgr.SetCookie(ctx, session)
    
    return ctx.JSON(200, map[string]string{"message": "Login successful"})
}
```

### IP Address Validation

Validate session IP address:

```go
func validateIPHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Check if IP address matches
    currentIP := ctx.Request().RemoteAddr
    if session.IPAddress != currentIP {
        // IP changed - potential session hijacking
        sessionMgr.Destroy(ctx, session.ID)
        return ctx.JSON(401, map[string]string{
            "error": "Session invalid - IP address changed",
        })
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session valid"})
}
```

### User Agent Validation

Validate session user agent:

```go
func validateUserAgentHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Check if user agent matches
    currentUA := ctx.GetHeader("User-Agent")
    if session.UserAgent != currentUA {
        // User agent changed - potential session hijacking
        sessionMgr.Destroy(ctx, session.ID)
        return ctx.JSON(401, map[string]string{
            "error": "Session invalid - user agent changed",
        })
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session valid"})
}
```

## Multi-Tenant Sessions

Sessions automatically support multi-tenancy:

```go
func createTenantSessionHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Create session - tenant ID is automatically set from context
    session, err := sessionMgr.Create(ctx)
    if err != nil {
        return err
    }
    
    // Session includes tenant ID
    fmt.Printf("Session for tenant: %s\n", session.TenantID)
    
    // Store tenant-specific data
    session.Data["tenant_settings"] = getTenantSettings(session.TenantID)
    
    sessionMgr.Save(ctx, session)
    sessionMgr.SetCookie(ctx, session)
    
    return ctx.JSON(200, map[string]interface{}{
        "message":   "Session created",
        "tenant_id": session.TenantID,
    })
}
```

## Session Cleanup

### Automatic Cleanup

Sessions are automatically cleaned up based on `CleanupInterval`:

```go
SessionConfig: pkg.SessionConfig{
    CleanupInterval: 15 * time.Minute,  // Clean every 15 minutes
}
```

### Manual Cleanup

Trigger cleanup manually:

```go
func cleanupSessionsHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    err := sessionMgr.CleanupExpired()
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{
        "message": "Expired sessions cleaned up",
    })
}
```

## Best Practices

### 1. Use Secure Configuration in Production

```go
SessionConfig: pkg.SessionConfig{
    CookieSecure:   true,      // HTTPS only
    CookieHTTPOnly: true,      // No JavaScript
    CookieSameSite: "Strict",  // CSRF protection
    SessionLifetime: 2 * time.Hour,  // Short lifetime
}
```

### 2. Store Minimal Data in Sessions

```go
// Good: Store only IDs
session.Data["user_id"] = user.ID

// Bad: Store entire objects
session.Data["user"] = user  // Avoid large objects
```

### 3. Regenerate Session on Privilege Change

```go
func promoteUserHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Get old session
    oldSession, _ := sessionMgr.GetSessionFromCookie(ctx)
    
    // Promote user
    promoteUser(userID)
    
    // Destroy old session
    sessionMgr.Destroy(ctx, oldSession.ID)
    
    // Create new session with updated privileges
    newSession, _ := sessionMgr.Create(ctx)
    newSession.Data["user_id"] = userID
    newSession.Data["roles"] = getUpdatedRoles(userID)
    
    sessionMgr.Save(ctx, newSession)
    sessionMgr.SetCookie(ctx, newSession)
    
    return ctx.JSON(200, map[string]string{"message": "User promoted"})
}
```

### 4. Implement Session Timeout

```go
func activityCheckMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    sessionMgr := ctx.Session()
    
    session, err := sessionMgr.GetSessionFromCookie(ctx)
    if err != nil {
        return next(ctx)
    }
    
    // Check last activity
    if lastActivity, ok := session.Data["last_activity"].(time.Time); ok {
        if time.Since(lastActivity) > 30*time.Minute {
            // Inactive for 30 minutes - destroy session
            sessionMgr.Destroy(ctx, session.ID)
            return ctx.JSON(401, map[string]string{
                "error": "Session timeout due to inactivity",
            })
        }
    }
    
    // Update last activity
    session.Data["last_activity"] = time.Now()
    sessionMgr.Save(ctx, session)
    
    return next(ctx)
}
```

### 5. Use Database Storage for Production

```go
// Production configuration
SessionConfig: pkg.SessionConfig{
    StorageType: pkg.SessionStorageDatabase,  // Persistent
    // ...
}
```

### 6. Implement Remember Me Functionality

```go
func loginWithRememberMeHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    // Check remember me option
    rememberMe := ctx.QueryParam("remember_me") == "true"
    
    // Create session
    session, _ := sessionMgr.Create(ctx)
    session.Data["user_id"] = user.ID
    
    // Extend lifetime if remember me is checked
    if rememberMe {
        session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)  // 30 days
    }
    
    sessionMgr.Save(ctx, session)
    sessionMgr.SetCookie(ctx, session)
    
    return ctx.JSON(200, map[string]string{"message": "Login successful"})
}
```

### 7. Log Session Events

```go
func loginHandler(ctx pkg.Context) error {
    sessionMgr := ctx.Session()
    
    session, _ := sessionMgr.Create(ctx)
    session.Data["user_id"] = user.ID
    
    // Log session creation
    log.Printf("Session created: user=%s, ip=%s, session=%s",
        user.ID, session.IPAddress, session.ID)
    
    sessionMgr.Save(ctx, session)
    sessionMgr.SetCookie(ctx, session)
    
    return ctx.JSON(200, map[string]string{"message": "Login successful"})
}
```

## Troubleshooting

### Session Not Persisting

**Problem:** Session data is lost between requests

**Solutions:**
- Verify `SetCookie` is called after creating session
- Check cookie settings (Secure flag with HTTP)
- Verify encryption key is consistent
- Check browser cookie settings

### Session Expired Immediately

**Problem:** Session expires right after creation

**Solutions:**
- Check `SessionLifetime` configuration
- Verify system time is correct
- Check database time zone settings
- Ensure cleanup interval isn't too aggressive

### Cannot Decrypt Session Cookie

**Problem:** "failed to decrypt session ID" error

**Solutions:**
- Verify encryption key is 32 bytes
- Ensure same encryption key across restarts
- Check for cookie corruption
- Verify base64 encoding is correct

### High Memory Usage

**Problem:** Session storage consuming too much memory

**Solutions:**
- Use database storage instead of cache
- Reduce `SessionLifetime`
- Implement more aggressive cleanup
- Store less data in sessions

## See Also

- [Configuration Guide](configuration.md) - Session configuration options
- [Security Guide](security.md) - Security best practices
- [Cookie Guide](context.md#cookies) - Cookie management
- [API Reference: Session](../api/session.md) - Complete API documentation
