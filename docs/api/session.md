---
title: "Session API"
description: "Session manager interface for user session management"
category: "api"
tags: ["api", "session", "authentication", "security"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "security.md"
  - "../guides/sessions.md"
---

# Session API

## Overview

The `SessionManager` interface provides secure session management for the Rockstar Web Framework. It supports multiple storage backends (database, cache, filesystem), encrypted session cookies, automatic expiration, and cleanup.

Sessions are accessible through the Context in request handlers and provide a secure way to maintain user state across requests.

**Primary Use Cases:**
- User authentication state management
- Shopping cart data storage
- User preferences and settings
- Multi-step form data
- Temporary user-specific data

## Interface Definition

```go
type SessionManager interface {
    // Lifecycle
    Create(ctx Context) (*Session, error)
    Load(ctx Context, sessionID string) (*Session, error)
    Save(ctx Context, session *Session) error
    Destroy(ctx Context, sessionID string) error
    Refresh(ctx Context, sessionID string) error

    // Data access
    Get(sessionID, key string) (interface{}, error)
    Set(sessionID, key string, value interface{}) error
    Delete(sessionID, key string) error
    Clear(sessionID string) error

    // Cookie management
    SetCookie(ctx Context, session *Session) error
    GetSessionFromCookie(ctx Context) (*Session, error)

    // Validation
    IsValid(sessionID string) bool
    IsExpired(sessionID string) bool

    // Maintenance
    CleanupExpired() error
}
```

## Configuration

### SessionConfig

```go
type SessionConfig struct {
    // StorageType specifies the session storage backend.
    // Options: "database", "cache", "filesystem"
    // Default: "database"
    StorageType SessionStorageType

    // CookieName is the name of the session cookie.
    // Default: "rockstar_session"
    CookieName string

    // CookiePath is the path scope for the session cookie.
    // Default: "/"
    CookiePath string

    // CookieDomain is the domain scope for the session cookie.
    // Default: ""
    CookieDomain string

    // CookieSecure indicates if the cookie should only be sent over HTTPS.
    // Default: true
    CookieSecure bool

    // CookieHTTPOnly indicates if the cookie should be inaccessible to JavaScript.
    // Default: true
    CookieHTTPOnly bool

    // CookieSameSite specifies the SameSite attribute for the cookie.
    // Options: "Strict", "Lax", "None"
    // Default: "Lax"
    CookieSameSite string

    // SessionLifetime is the duration before a session expires.
    // Default: 24 hours
    SessionLifetime time.Duration

    // EncryptionKey is the AES-256 key (32 bytes) for encrypting session data.
    // Required, no default
    EncryptionKey []byte

    // FilesystemPath is the directory path for filesystem-based session storage.
    // Default: "./sessions"
    FilesystemPath string

    // CleanupInterval is the interval for cleaning up expired sessions.
    // Default: 1 hour
    CleanupInterval time.Duration
}
```

**Example**:
```go
config := pkg.SessionConfig{
    StorageType:     pkg.SessionStorageDatabase,
    CookieName:      "my_app_session",
    CookieSecure:    true,
    CookieHTTPOnly:  true,
    SessionLifetime: 24 * time.Hour,
    EncryptionKey:   []byte("your-32-byte-encryption-key-here"),
}
```

### Session Structure

```go
type Session struct {
    ID        string
    Data      map[string]interface{}
    UserID    string
    TenantID  string
    IPAddress string
    UserAgent string
    ExpiresAt time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Lifecycle Management

### Create

```go
func Create(ctx Context) (*Session, error)
```

**Description**: Creates a new session with a unique ID and stores it in the configured backend.

**Parameters**:
- `ctx` (Context): Request context (used to extract user, tenant, IP, and user agent)

**Returns**:
- `*Session`: Newly created session
- `error`: Error if session creation fails

**Example**:
```go
func loginHandler(ctx pkg.Context) error {
    // Authenticate user...
    
    // Create new session
    session, err := ctx.Session().Create(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
    }
    
    // Store user data in session
    session.Data["user_id"] = userID
    session.Data["username"] = username
    session.Data["login_time"] = time.Now()
    
    // Save session
    if err := ctx.Session().Save(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save session"})
    }
    
    // Set session cookie
    if err := ctx.Session().SetCookie(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Logged in successfully"})
}
```

### Load

```go
func Load(ctx Context, sessionID string) (*Session, error)
```

**Description**: Loads a session from storage by ID. Returns an error if the session doesn't exist or has expired.

**Parameters**:
- `ctx` (Context): Request context
- `sessionID` (string): Session identifier

**Returns**:
- `*Session`: Loaded session
- `error`: Error if session not found or expired

**Example**:
```go
func protectedHandler(ctx pkg.Context) error {
    // Get session from cookie
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Access session data
    userID := session.Data["user_id"]
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
        "message": "Welcome back!",
    })
}
```

### Save

```go
func Save(ctx Context, session *Session) error
```

**Description**: Saves session changes to storage. Updates the UpdatedAt timestamp.

**Parameters**:
- `ctx` (Context): Request context
- `session` (*Session): Session to save

**Returns**:
- `error`: Error if save fails

**Example**:
```go
func updatePreferencesHandler(ctx pkg.Context) error {
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Update session data
    session.Data["theme"] = "dark"
    session.Data["language"] = "en"
    
    // Save changes
    if err := ctx.Session().Save(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save preferences"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Preferences updated"})
}
```

### Destroy

```go
func Destroy(ctx Context, sessionID string) error
```

**Description**: Destroys a session, removing it from storage.

**Parameters**:
- `ctx` (Context): Request context
- `sessionID` (string): Session identifier

**Returns**:
- `error`: Error if destruction fails

**Example**:
```go
func logoutHandler(ctx pkg.Context) error {
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Destroy session
    if err := ctx.Session().Destroy(ctx, session.ID); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to logout"})
    }
    
    // Clear session cookie
    ctx.SetCookie(&pkg.Cookie{
        Name:   "rockstar_session",
        Value:  "",
        MaxAge: -1, // Delete cookie
    })
    
    return ctx.JSON(200, map[string]string{"message": "Logged out successfully"})
}
```

### Refresh

```go
func Refresh(ctx Context, sessionID string) error
```

**Description**: Refreshes a session's expiration time, extending its lifetime.

**Parameters**:
- `ctx` (Context): Request context
- `sessionID` (string): Session identifier

**Returns**:
- `error`: Error if refresh fails

**Example**:
```go
func activityMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err == nil {
        // Refresh session on each request
        ctx.Session().Refresh(ctx, session.ID)
    }
    
    return next(ctx)
}
```

## Data Access

### Get

```go
func Get(sessionID, key string) (interface{}, error)
```

**Description**: Retrieves a value from session data by key.

**Parameters**:
- `sessionID` (string): Session identifier
- `key` (string): Data key

**Returns**:
- `interface{}`: Value if found
- `error`: Error if session not found or key doesn't exist

**Example**:
```go
func handler(ctx pkg.Context) error {
    session, _ := ctx.Session().GetSessionFromCookie(ctx)
    
    value, err := ctx.Session().Get(session.ID, "cart_items")
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Cart not found"})
    }
    
    return ctx.JSON(200, value)
}
```

### Set

```go
func Set(sessionID, key string, value interface{}) error
```

**Description**: Sets a value in session data.

**Parameters**:
- `sessionID` (string): Session identifier
- `key` (string): Data key
- `value` (interface{}): Value to store

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func addToCartHandler(ctx pkg.Context) error {
    session, _ := ctx.Session().GetSessionFromCookie(ctx)
    
    // Get current cart
    cart, _ := ctx.Session().Get(session.ID, "cart_items")
    if cart == nil {
        cart = []string{}
    }
    
    // Add item
    cartItems := cart.([]string)
    cartItems = append(cartItems, productID)
    
    // Save back to session
    err := ctx.Session().Set(session.ID, "cart_items", cartItems)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to update cart"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Item added to cart"})
}
```

### Delete

```go
func Delete(sessionID, key string) error
```

**Description**: Deletes a key from session data.

**Parameters**:
- `sessionID` (string): Session identifier
- `key` (string): Data key to delete

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func clearCartHandler(ctx pkg.Context) error {
    session, _ := ctx.Session().GetSessionFromCookie(ctx)
    
    err := ctx.Session().Delete(session.ID, "cart_items")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to clear cart"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cart cleared"})
}
```

### Clear

```go
func Clear(sessionID string) error
```

**Description**: Clears all data from a session while keeping the session itself active.

**Parameters**:
- `sessionID` (string): Session identifier

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func resetSessionHandler(ctx pkg.Context) error {
    session, _ := ctx.Session().GetSessionFromCookie(ctx)
    
    err := ctx.Session().Clear(session.ID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to clear session"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session cleared"})
}
```

## Cookie Management

### SetCookie

```go
func SetCookie(ctx Context, session *Session) error
```

**Description**: Sets an encrypted session cookie in the response.

**Parameters**:
- `ctx` (Context): Request context
- `session` (*Session): Session to store in cookie

**Returns**:
- `error`: Error if cookie setting fails

**Example**:
```go
func loginHandler(ctx pkg.Context) error {
    session, _ := ctx.Session().Create(ctx)
    session.Data["user_id"] = userID
    ctx.Session().Save(ctx, session)
    
    // Set encrypted session cookie
    if err := ctx.Session().SetCookie(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Logged in"})
}
```

### GetSessionFromCookie

```go
func GetSessionFromCookie(ctx Context) (*Session, error)
```

**Description**: Retrieves and decrypts a session from the request cookie.

**Parameters**:
- `ctx` (Context): Request context

**Returns**:
- `*Session`: Session from cookie
- `error`: Error if cookie not found, decryption fails, or session expired

**Example**:
```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    // Store session in context for handlers
    ctx.Set("session", session)
    
    return next(ctx)
}
```

## Validation

### IsValid

```go
func IsValid(sessionID string) bool
```

**Description**: Checks if a session exists and is not expired.

**Parameters**:
- `sessionID` (string): Session identifier

**Returns**:
- `bool`: true if session is valid, false otherwise

**Example**:
```go
func handler(ctx pkg.Context) error {
    sessionID := ctx.Query()["session_id"]
    
    if !ctx.Session().IsValid(sessionID) {
        return ctx.JSON(401, map[string]string{"error": "Invalid session"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session is valid"})
}
```

### IsExpired

```go
func IsExpired(sessionID string) bool
```

**Description**: Checks if a session has expired.

**Parameters**:
- `sessionID` (string): Session identifier

**Returns**:
- `bool`: true if session is expired, false otherwise

**Example**:
```go
func handler(ctx pkg.Context) error {
    session, _ := ctx.Session().GetSessionFromCookie(ctx)
    
    if ctx.Session().IsExpired(session.ID) {
        return ctx.JSON(401, map[string]string{"error": "Session expired"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Session active"})
}
```

## Maintenance

### CleanupExpired

```go
func CleanupExpired() error
```

**Description**: Removes all expired sessions from storage. This is called automatically at the configured cleanup interval.

**Returns**:
- `error`: Error if cleanup fails

**Example**:
```go
func cleanupHandler(ctx pkg.Context) error {
    if err := ctx.Session().CleanupExpired(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Cleanup failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cleanup completed"})
}
```

## Complete Example

Here's a complete example demonstrating session management in a login/logout flow:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

func main() {
    config := pkg.FrameworkConfig{
        SessionConfig: pkg.SessionConfig{
            StorageType:     pkg.SessionStorageDatabase,
            CookieName:      "app_session",
            CookieSecure:    true,
            CookieHTTPOnly:  true,
            SessionLifetime: 24 * time.Hour,
            EncryptionKey:   []byte("your-32-byte-encryption-key-here"),
        },
    }

    app, _ := pkg.New(config)
    router := app.Router()

    // Login endpoint
    router.POST("/login", loginHandler)
    
    // Protected endpoints
    router.GET("/profile", authMiddleware, profileHandler)
    router.POST("/logout", authMiddleware, logoutHandler)

    app.Listen(":8080")
}

func loginHandler(ctx pkg.Context) error {
    // Validate credentials
    username := ctx.FormValue("username")
    password := ctx.FormValue("password")
    
    user, err := authenticateUser(username, password)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid credentials"})
    }

    // Create session
    session, err := ctx.Session().Create(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
    }

    // Store user data
    session.Data["user_id"] = user.ID
    session.Data["username"] = user.Username
    session.Data["roles"] = user.Roles
    session.Data["login_time"] = time.Now()

    // Save session
    if err := ctx.Session().Save(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save session"})
    }

    // Set cookie
    if err := ctx.Session().SetCookie(ctx, session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
    }

    return ctx.JSON(200, map[string]interface{}{
        "message": "Logged in successfully",
        "user":    user,
    })
}

func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Get session from cookie
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }

    // Check if session is valid
    if !ctx.Session().IsValid(session.ID) {
        return ctx.JSON(401, map[string]string{"error": "Invalid session"})
    }

    // Refresh session on activity
    ctx.Session().Refresh(ctx, session.ID)

    // Store session in context
    ctx.Set("session", session)
    ctx.Set("user_id", session.Data["user_id"])

    return next(ctx)
}

func profileHandler(ctx pkg.Context) error {
    session := ctx.Get("session").(*pkg.Session)
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id":    session.Data["user_id"],
        "username":   session.Data["username"],
        "login_time": session.Data["login_time"],
    })
}

func logoutHandler(ctx pkg.Context) error {
    session := ctx.Get("session").(*pkg.Session)

    // Destroy session
    if err := ctx.Session().Destroy(ctx, session.ID); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to logout"})
    }

    // Clear cookie
    ctx.SetCookie(&pkg.Cookie{
        Name:   "app_session",
        Value:  "",
        MaxAge: -1,
    })

    return ctx.JSON(200, map[string]string{"message": "Logged out successfully"})
}

func authenticateUser(username, password string) (*User, error) {
    // Authentication logic here
    return &User{ID: "123", Username: username}, nil
}

type User struct {
    ID       string
    Username string
    Roles    []string
}
```

## Best Practices

### Security

1. **Always use HTTPS in production**: Set `CookieSecure: true`
2. **Use strong encryption keys**: Generate random 32-byte keys
3. **Set HTTPOnly cookies**: Prevent JavaScript access with `CookieHTTPOnly: true`
4. **Use appropriate SameSite**: Set to "Strict" or "Lax" to prevent CSRF
5. **Validate session on every request**: Check expiration and validity

### Performance

1. **Choose appropriate storage**: Use cache for high-traffic, database for persistence
2. **Set reasonable lifetimes**: Balance security and user experience
3. **Clean up regularly**: Configure appropriate cleanup intervals
4. **Use request-scoped data**: Store temporary data in request cache, not session

### Data Management

1. **Store minimal data**: Only store essential user information
2. **Don't store sensitive data**: Never store passwords or credit cards
3. **Use typed data**: Convert session data to proper types when retrieving
4. **Handle missing data**: Always check if keys exist before accessing

## See Also

- [Framework API](framework.md)
- [Context API](context.md)
- [Security API](security.md)
- [Sessions Guide](../guides/sessions.md)
- [Security Guide](../guides/security.md)
