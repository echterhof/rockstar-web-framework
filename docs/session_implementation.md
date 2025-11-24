# Session Implementation Architecture

This document describes the session management implementation architecture in the Rockstar framework.

## File Structure

The session management system is split across multiple files for better organization and maintainability:

```
pkg/
├── session.go              # SessionManager interface and core logic
├── session_impl.go         # Database storage implementation (production)
├── session_impl_test.go    # Mock storage implementation (testing)
├── database.go             # DatabaseManager interface
└── database_impl.go        # DatabaseManager implementation
```

## Components

### 1. Session Manager (`session.go`)

The main session manager implementation that handles:
- Session lifecycle (create, load, save, destroy)
- Multiple storage backends (database, cache, filesystem)
- Cookie encryption/decryption
- Session validation and expiration
- Automatic cleanup

**Key Types:**
```go
type SessionManager interface {
    Create(ctx Context) (*Session, error)
    Load(ctx Context, sessionID string) (*Session, error)
    Save(ctx Context, session *Session) error
    Destroy(ctx Context, sessionID string) error
    // ... more methods
}

type sessionManager struct {
    config      *SessionConfig
    db          DatabaseManager
    cache       CacheManager
    cipher      cipher.Block
    mu          sync.RWMutex
    sessions    map[string]*Session
    stopCleanup chan struct{}
}
```

### 2. Session Storage Implementation (`session_impl.go`)

Provides database-specific session storage operations:

**Key Type:**
```go
type sessionStorage struct {
    db DatabaseManager
}
```

**Methods:**
- `SaveSession(session *Session) error` - Persists session to database
- `LoadSession(sessionID string) (*Session, error)` - Retrieves session from database
- `DeleteSession(sessionID string) error` - Removes session from database
- `CleanupExpiredSessions() error` - Removes expired sessions

**Integration with DatabaseManager:**

The `databaseManager` delegates session operations to `sessionStorage`:

```go
func (dm *databaseManager) SaveSession(session *Session) error {
    ss := newSessionStorage(dm)
    return ss.SaveSession(session)
}
```

This design allows:
- Clean separation of concerns
- Easy testing with mock implementations
- Flexibility to swap storage backends

### 3. Test Implementation (`session_impl_test.go`)

Provides in-memory mock storage for testing:

**Key Type:**
```go
type mockSessionStorage struct {
    sessions map[string]*Session
}
```

This mock implementation:
- Stores sessions in memory
- Simulates database behavior
- Enables fast unit testing
- No external dependencies required

**Build Tags:**

The implementation uses Go build tags to separate production and test code:

- `session_impl.go`: `//go:build !test`
- `session_impl_test.go`: `//go:build test`

## Storage Backends

### Database Storage

Uses the `sessionStorage` implementation to persist sessions in a relational database.

**Advantages:**
- Persistent across restarts
- ACID compliance
- Multi-server support
- Complex query support

**Schema:**
```sql
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    data JSON,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    ip_address VARCHAR(45),
    user_agent TEXT,
    INDEX idx_sessions_expires (expires_at),
    INDEX idx_sessions_user (user_id),
    INDEX idx_sessions_tenant (tenant_id)
);
```

### Cache Storage

Uses the `CacheManager` interface for high-performance session storage.

**Advantages:**
- Extremely fast access
- Automatic expiration
- Distributed caching
- Low latency

### Filesystem Storage

Stores sessions as JSON files on disk.

**Advantages:**
- No external dependencies
- Simple debugging
- Easy backup
- Good for development

## Data Flow

### Session Creation

```
Client Request
    ↓
SessionManager.Create(ctx)
    ↓
Generate Session ID
    ↓
Create Session Object
    ↓
SessionManager.saveToStorage()
    ↓
[Database Storage]
    ↓
DatabaseManager.SaveSession()
    ↓
sessionStorage.SaveSession()
    ↓
SQL INSERT/UPDATE
    ↓
Session Saved
```

### Session Loading

```
Client Request with Cookie
    ↓
SessionManager.GetSessionFromCookie(ctx)
    ↓
Decrypt Session ID
    ↓
SessionManager.Load(ctx, sessionID)
    ↓
SessionManager.loadFromStorage()
    ↓
[Database Storage]
    ↓
DatabaseManager.LoadSession()
    ↓
sessionStorage.LoadSession()
    ↓
SQL SELECT
    ↓
Validate Expiration
    ↓
Return Session
```

## Multi-Tenancy Support

Sessions are automatically isolated by tenant:

1. **Session Creation**: Tenant ID is extracted from context and stored in session
2. **Session Storage**: Tenant ID is indexed in database for efficient queries
3. **Session Loading**: Sessions are filtered by tenant ID
4. **Session Cleanup**: Expired sessions are cleaned up per tenant

## Security Features

### 1. Cookie Encryption

Session IDs in cookies are encrypted using AES-256-GCM:

```go
func (sm *sessionManager) encryptSessionID(sessionID string) (string, error) {
    gcm, _ := cipher.NewGCM(sm.cipher)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    ciphertext := gcm.Seal(nonce, nonce, []byte(sessionID), nil)
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}
```

### 2. Secure Cookie Settings

```go
cookie := &Cookie{
    Name:      "app_session",
    Value:     encryptedID,
    Secure:    true,  // HTTPS only
    HttpOnly:  true,  // No JavaScript access
    SameSite:  "Lax", // CSRF protection
}
```

### 3. Session Expiration

- Automatic expiration based on `SessionLifetime`
- Database queries filter expired sessions
- Background cleanup removes expired sessions
- Refresh mechanism extends session lifetime

## Performance Considerations

### 1. Connection Pooling

Database connections are pooled for efficiency:

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)
```

### 2. Indexing

Database indexes on frequently queried columns:
- `id` (PRIMARY KEY)
- `expires_at` (for cleanup queries)
- `user_id` (for user session queries)
- `tenant_id` (for multi-tenant isolation)

### 3. Cleanup Strategy

Background goroutine periodically removes expired sessions:

```go
func (sm *sessionManager) cleanupLoop() {
    ticker := time.NewTicker(sm.config.CleanupInterval)
    for {
        select {
        case <-ticker.C:
            sm.CleanupExpired()
        case <-sm.stopCleanup:
            return
        }
    }
}
```

## Testing

### Unit Tests

Use the mock implementation in `session_impl_test.go`:

```go
func TestSessionStorage(t *testing.T) {
    db := NewDatabaseManager()
    
    session := &Session{
        ID:        "test-session-id",
        UserID:    "user-123",
        TenantID:  "tenant-1",
        Data:      map[string]interface{}{"key": "value"},
        ExpiresAt: time.Now().Add(time.Hour),
    }
    
    // Save session
    err := db.SaveSession(session)
    if err != nil {
        t.Errorf("Failed to save session: %v", err)
    }
    
    // Load session
    loaded, err := db.LoadSession(session.ID)
    if err != nil {
        t.Errorf("Failed to load session: %v", err)
    }
    
    // Verify data
    if loaded.Data["key"] != "value" {
        t.Errorf("Session data mismatch")
    }
}
```

### Integration Tests

Use real database for integration testing:

```go
func TestSessionIntegration(t *testing.T) {
    db := NewDatabaseManager()
    db.Connect(DatabaseConfig{
        Driver:   "sqlite3",
        Database: ":memory:",
    })
    defer db.Close()
    
    db.CreateTables()
    
    // Test with real database operations
    // ...
}
```

## Migration Guide

If you're upgrading from an older version where session methods were in `database_impl.go`:

### Before
```go
// Session methods were directly in database_impl.go
func (dm *databaseManager) SaveSession(session *Session) error {
    // Implementation here
}
```

### After
```go
// Session methods delegate to sessionStorage
func (dm *databaseManager) SaveSession(session *Session) error {
    ss := newSessionStorage(dm)
    return ss.SaveSession(session)
}
```

**No API changes** - The public interface remains the same, only the internal implementation is reorganized.

## Best Practices

1. **Always use encryption keys from secure sources**
   ```go
   encryptionKey := []byte(os.Getenv("SESSION_ENCRYPTION_KEY"))
   ```

2. **Configure appropriate session lifetime**
   ```go
   SessionLifetime: 30 * time.Minute, // For sensitive operations
   ```

3. **Enable cleanup for production**
   ```go
   CleanupInterval: 1 * time.Hour,
   ```

4. **Use secure cookie settings**
   ```go
   CookieSecure:   true,
   CookieHTTPOnly: true,
   CookieSameSite: "Strict",
   ```

5. **Monitor session storage size**
   - Implement metrics for session count
   - Alert on excessive session growth
   - Regular cleanup verification

## Troubleshooting

### Sessions not persisting

Check database connection and table creation:
```go
if err := db.Ping(); err != nil {
    log.Fatal("Database not connected")
}
db.CreateTables()
```

### Sessions expiring too quickly

Adjust session lifetime:
```go
config.SessionLifetime = 24 * time.Hour
```

### Cookie not being set

Verify cookie configuration:
```go
// For local development
config.CookieSecure = false // Only for HTTP
config.CookieDomain = ""    // Empty for localhost
```

### Cleanup not running

Check cleanup interval and ensure session manager is not stopped:
```go
config.CleanupInterval = 15 * time.Minute
// Don't call sessionManager.Stop() prematurely
```

## Future Enhancements

Potential improvements to the session implementation:

1. **Redis Backend**: Native Redis support for distributed sessions
2. **Session Replication**: Automatic replication across multiple databases
3. **Session Analytics**: Built-in analytics for session usage patterns
4. **Custom Serialization**: Support for custom session data serializers
5. **Session Versioning**: Track session schema versions for migrations

## References

- [Session Management Documentation](./session_management.md)
- [Database Implementation](../pkg/database_impl.go)
- [Session Example](../examples/session_example.go)
- [API Reference](./API_REFERENCE.md)
