---
title: "Database API"
description: "DatabaseManager interface for database operations and connection management"
category: "api"
tags: ["api", "database", "sql", "transactions", "queries"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "../guides/database.md"
---

# Database API

## Overview

The `DatabaseManager` interface provides a comprehensive abstraction for database operations in the Rockstar Web Framework. It supports multiple database engines (MySQL, PostgreSQL, MSSQL, SQLite) with a unified API, connection pooling, transaction management, and framework-specific model operations.

The DatabaseManager is accessed through the Context interface in handlers, eliminating the need for global database instances and enabling proper dependency injection.

**Primary Use Cases:**
- Executing SQL queries and commands
- Managing database connections and connection pools
- Handling transactions with commit/rollback support
- Storing and retrieving framework models (sessions, tokens, tenants)
- Rate limiting and metrics storage
- Database migrations and schema management

## Type Definition

### DatabaseManager Interface

```go
type DatabaseManager interface {
    // Connection management
    Connect(config DatabaseConfig) error
    Close() error
    Ping() error
    Stats() DatabaseStats
    IsConnected() bool

    // Query execution
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)

    // Prepared statements
    Prepare(query string) (*sql.Stmt, error)

    // Transaction support
    Begin() (Transaction, error)
    BeginTx(opts *sql.TxOptions) (Transaction, error)

    // Framework-specific model operations
    SaveSession(session *Session) error
    LoadSession(sessionID string) (*Session, error)
    DeleteSession(sessionID string) error
    CleanupExpiredSessions() error

    SaveAccessToken(token *AccessToken) error
    LoadAccessToken(tokenValue string) (*AccessToken, error)
    ValidateAccessToken(tokenValue string) (*AccessToken, error)
    DeleteAccessToken(tokenValue string) error
    CleanupExpiredTokens() error

    SaveTenant(tenant *Tenant) error
    LoadTenant(tenantID string) (*Tenant, error)
    LoadTenantByHost(hostname string) (*Tenant, error)

    SaveWorkloadMetrics(metrics *WorkloadMetrics) error
    GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)

    // Rate limiting support
    CheckRateLimit(key string, limit int, window time.Duration) (bool, error)
    IncrementRateLimit(key string, window time.Duration) error

    // Migration support
    Migrate() error
    CreateTables() error
    DropTables() error

    // Plugin system support
    InitializePluginTables() error

    // SQL loader support
    GetQuery(name string) (string, error)
}
```

### Transaction Interface

```go
type Transaction interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Exec(query string, args ...interface{}) (sql.Result, error)
    Prepare(query string) (*sql.Stmt, error)
    Commit() error
    Rollback() error
}
```


## Connection Management

### Connect

```go
func Connect(config DatabaseConfig) error
```

**Description**: Establishes a database connection using the provided configuration. Supports MySQL, PostgreSQL, MSSQL, and SQLite databases with connection pooling.

**Parameters**:
- `config` (DatabaseConfig): Database configuration including driver, host, credentials, and pool settings

**Returns**:
- `error`: Error if connection fails

**Example**:
```go
config := pkg.DatabaseConfig{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "myapp",
    Username:        "dbuser",
    Password:        "dbpass",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}

db := pkg.NewDatabaseManager()
if err := db.Connect(config); err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}
defer db.Close()
```

**See Also**:
- [DatabaseConfig](#databaseconfig)
- [Database Guide](../guides/database.md#connection-configuration)

### Close

```go
func Close() error
```

**Description**: Closes the database connection and releases all resources. Should be called when the application shuts down.

**Returns**:
- `error`: Error if closing fails

**Example**:
```go
db := ctx.DB()
defer db.Close()
```


### Ping

```go
func Ping() error
```

**Description**: Tests the database connection to verify it's alive and responsive.

**Returns**:
- `error`: Error if ping fails or connection is not established

**Example**:
```go
func healthCheckHandler(ctx pkg.Context) error {
    db := ctx.DB()
    if err := db.Ping(); err != nil {
        return ctx.JSON(503, map[string]string{
            "status": "unhealthy",
            "error": "Database connection failed",
        })
    }
    
    return ctx.JSON(200, map[string]string{"status": "healthy"})
}
```

### Stats

```go
func Stats() DatabaseStats
```

**Description**: Returns current database connection pool statistics for monitoring and debugging.

**Returns**:
- `DatabaseStats`: Connection pool statistics

**Example**:
```go
func statsHandler(ctx pkg.Context) error {
    db := ctx.DB()
    stats := db.Stats()
    
    return ctx.JSON(200, map[string]interface{}{
        "open_connections": stats.OpenConnections,
        "in_use": stats.InUse,
        "idle": stats.Idle,
        "wait_count": stats.WaitCount,
        "wait_duration": stats.WaitDuration.String(),
    })
}
```

**See Also**:
- [DatabaseStats](#databasestats)

### IsConnected

```go
func IsConnected() bool
```

**Description**: Checks if a database connection is currently established.

**Returns**:
- `bool`: true if connected, false otherwise

**Example**:
```go
func handler(ctx pkg.Context) error {
    db := ctx.DB()
    if !db.IsConnected() {
        return ctx.JSON(503, map[string]string{"error": "Database unavailable"})
    }
    
    // Proceed with database operations...
    return ctx.JSON(200, result)
}
```


## Query Execution

### Query

```go
func Query(query string, args ...interface{}) (*sql.Rows, error)
```

**Description**: Executes a SQL query that returns multiple rows. Use for SELECT statements.

**Parameters**:
- `query` (string): SQL query with placeholders
- `args` (...interface{}): Query parameters for placeholders

**Returns**:
- `*sql.Rows`: Result rows (must be closed after use)
- `error`: Error if query execution fails

**Example**:
```go
func listUsersHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    rows, err := db.Query("SELECT id, name, email FROM users WHERE active = ?", true)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query failed"})
    }
    defer rows.Close()
    
    var users []map[string]interface{}
    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            return ctx.JSON(500, map[string]string{"error": "Scan failed"})
        }
        users = append(users, map[string]interface{}{
            "id": id,
            "name": name,
            "email": email,
        })
    }
    
    return ctx.JSON(200, users)
}
```

**Note**: Always close rows with `defer rows.Close()` to prevent resource leaks.

### QueryRow

```go
func QueryRow(query string, args ...interface{}) *sql.Row
```

**Description**: Executes a SQL query that returns at most one row. Use for SELECT statements that return a single result.

**Parameters**:
- `query` (string): SQL query with placeholders
- `args` (...interface{}): Query parameters for placeholders

**Returns**:
- `*sql.Row`: Single result row

**Example**:
```go
func getUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Param("id")
    
    var id int
    var name, email string
    
    row := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userID)
    err := row.Scan(&id, &name, &email)
    
    if err == sql.ErrNoRows {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query failed"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "id": id,
        "name": name,
        "email": email,
    })
}
```


### Exec

```go
func Exec(query string, args ...interface{}) (sql.Result, error)
```

**Description**: Executes a SQL command that doesn't return rows. Use for INSERT, UPDATE, DELETE, and DDL statements.

**Parameters**:
- `query` (string): SQL command with placeholders
- `args` (...interface{}): Command parameters for placeholders

**Returns**:
- `sql.Result`: Result with LastInsertId() and RowsAffected()
- `error`: Error if execution fails

**Example**:
```go
func createUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    name := ctx.FormValue("name")
    email := ctx.FormValue("email")
    
    result, err := db.Exec(
        "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
        name, email, time.Now(),
    )
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Insert failed"})
    }
    
    id, _ := result.LastInsertId()
    
    return ctx.JSON(201, map[string]interface{}{
        "id": id,
        "name": name,
        "email": email,
    })
}

func updateUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Param("id")
    name := ctx.FormValue("name")
    
    result, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", name, userID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Update failed"})
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "User updated"})
}
```

### Prepare

```go
func Prepare(query string) (*sql.Stmt, error)
```

**Description**: Creates a prepared statement for repeated execution with different parameters. Improves performance for frequently executed queries.

**Parameters**:
- `query` (string): SQL query with placeholders

**Returns**:
- `*sql.Stmt`: Prepared statement (must be closed after use)
- `error`: Error if preparation fails

**Example**:
```go
func batchInsertHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    stmt, err := db.Prepare("INSERT INTO logs (message, level, timestamp) VALUES (?, ?, ?)")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Prepare failed"})
    }
    defer stmt.Close()
    
    logs := []struct{ Message, Level string }{
        {"User logged in", "INFO"},
        {"Data processed", "INFO"},
        {"Cache cleared", "DEBUG"},
    }
    
    for _, log := range logs {
        _, err := stmt.Exec(log.Message, log.Level, time.Now())
        if err != nil {
            return ctx.JSON(500, map[string]string{"error": "Insert failed"})
        }
    }
    
    return ctx.JSON(200, map[string]string{"message": "Logs inserted"})
}
```


## Transaction Support

### Begin

```go
func Begin() (Transaction, error)
```

**Description**: Starts a new database transaction with default isolation level.

**Returns**:
- `Transaction`: Transaction interface for executing queries
- `error`: Error if transaction creation fails

**Example**:
```go
func transferFundsHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    tx, err := db.Begin()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Transaction failed"})
    }
    defer tx.Rollback() // Rollback if not committed
    
    // Deduct from source account
    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Debit failed"})
    }
    
    // Add to destination account
    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Credit failed"})
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Commit failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Transfer completed"})
}
```

**Best Practice**: Always use `defer tx.Rollback()` after creating a transaction. It's safe to call even after Commit().

### BeginTx

```go
func BeginTx(opts *sql.TxOptions) (Transaction, error)
```

**Description**: Starts a new database transaction with custom options (isolation level, read-only mode).

**Parameters**:
- `opts` (*sql.TxOptions): Transaction options including isolation level

**Returns**:
- `Transaction`: Transaction interface for executing queries
- `error`: Error if transaction creation fails

**Example**:
```go
func readOnlyReportHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Start read-only transaction with serializable isolation
    tx, err := db.BeginTx(&sql.TxOptions{
        Isolation: sql.LevelSerializable,
        ReadOnly:  true,
    })
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Transaction failed"})
    }
    defer tx.Rollback()
    
    // Execute read-only queries
    rows, err := tx.Query("SELECT * FROM reports WHERE date = ?", date)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query failed"})
    }
    defer rows.Close()
    
    // Process results...
    
    tx.Commit()
    return ctx.JSON(200, results)
}
```


## Transaction Methods

### Transaction.Query

```go
func Query(query string, args ...interface{}) (*sql.Rows, error)
```

**Description**: Executes a query within the transaction context.

**Example**: See [Begin](#begin) example.

### Transaction.QueryRow

```go
func QueryRow(query string, args ...interface{}) *sql.Row
```

**Description**: Executes a single-row query within the transaction context.

### Transaction.Exec

```go
func Exec(query string, args ...interface{}) (sql.Result, error)
```

**Description**: Executes a command within the transaction context.

### Transaction.Prepare

```go
func Prepare(query string) (*sql.Stmt, error)
```

**Description**: Creates a prepared statement within the transaction context.

### Transaction.Commit

```go
func Commit() error
```

**Description**: Commits the transaction, making all changes permanent.

**Returns**:
- `error`: Error if commit fails

### Transaction.Rollback

```go
func Rollback() error
```

**Description**: Rolls back the transaction, discarding all changes.

**Returns**:
- `error`: Error if rollback fails (safe to ignore if already committed)


## Session Management

### SaveSession

```go
func SaveSession(session *Session) error
```

**Description**: Saves a session to the database. Creates a new session or updates an existing one.

**Parameters**:
- `session` (*Session): Session object to save

**Returns**:
- `error`: Error if save fails

**Example**:
```go
func loginHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Create new session
    session := &pkg.Session{
        ID:        generateSessionID(),
        UserID:    user.ID,
        TenantID:  ctx.Tenant().ID,
        Data:      map[string]interface{}{"login_time": time.Now()},
        ExpiresAt: time.Now().Add(24 * time.Hour),
        IPAddress: ctx.Request().RemoteAddr,
        UserAgent: ctx.GetHeader("User-Agent"),
    }
    
    if err := db.SaveSession(session); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Session save failed"})
    }
    
    return ctx.JSON(200, map[string]string{"session_id": session.ID})
}
```

**See Also**:
- [Session Structure](#session)

### LoadSession

```go
func LoadSession(sessionID string) (*Session, error)
```

**Description**: Loads a session from the database by ID. Returns error if session not found or expired.

**Parameters**:
- `sessionID` (string): Session identifier

**Returns**:
- `*Session`: Session object
- `error`: Error if session not found or expired

**Example**:
```go
func validateSessionHandler(ctx pkg.Context) error {
    db := ctx.DB()
    sessionID := ctx.GetHeader("X-Session-ID")
    
    session, err := db.LoadSession(sessionID)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid session"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": session.UserID,
        "expires_at": session.ExpiresAt,
    })
}
```

### DeleteSession

```go
func DeleteSession(sessionID string) error
```

**Description**: Deletes a session from the database.

**Parameters**:
- `sessionID` (string): Session identifier

**Returns**:
- `error`: Error if deletion fails

**Example**:
```go
func logoutHandler(ctx pkg.Context) error {
    db := ctx.DB()
    sessionID := ctx.GetHeader("X-Session-ID")
    
    if err := db.DeleteSession(sessionID); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Logout failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Logged out"})
}
```

### CleanupExpiredSessions

```go
func CleanupExpiredSessions() error
```

**Description**: Removes all expired sessions from the database. Should be called periodically.

**Returns**:
- `error`: Error if cleanup fails

**Example**:
```go
// Run as a background task
func sessionCleanupTask(db pkg.DatabaseManager) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := db.CleanupExpiredSessions(); err != nil {
            log.Printf("Session cleanup failed: %v", err)
        }
    }
}
```


## Access Token Management

### SaveAccessToken

```go
func SaveAccessToken(token *AccessToken) error
```

**Description**: Saves an access token to the database for API authentication.

**Parameters**:
- `token` (*AccessToken): Access token object to save

**Returns**:
- `error`: Error if save fails

**Example**:
```go
func createTokenHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    token := &pkg.AccessToken{
        Token:     generateToken(),
        UserID:    ctx.User().ID,
        TenantID:  ctx.Tenant().ID,
        Scopes:    []string{"read", "write"},
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }
    
    if err := db.SaveAccessToken(token); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Token creation failed"})
    }
    
    return ctx.JSON(201, map[string]interface{}{
        "token": token.Token,
        "expires_at": token.ExpiresAt,
    })
}
```

### LoadAccessToken

```go
func LoadAccessToken(tokenValue string) (*AccessToken, error)
```

**Description**: Loads an access token from the database by token value.

**Parameters**:
- `tokenValue` (string): Token string

**Returns**:
- `*AccessToken`: Access token object
- `error`: Error if token not found

**Example**:
```go
func getTokenInfoHandler(ctx pkg.Context) error {
    db := ctx.DB()
    tokenValue := ctx.GetHeader("Authorization")
    
    token, err := db.LoadAccessToken(tokenValue)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Token not found"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": token.UserID,
        "scopes": token.Scopes,
        "expires_at": token.ExpiresAt,
    })
}
```

### ValidateAccessToken

```go
func ValidateAccessToken(tokenValue string) (*AccessToken, error)
```

**Description**: Validates an access token and returns it if valid and not expired.

**Parameters**:
- `tokenValue` (string): Token string

**Returns**:
- `*AccessToken`: Valid access token object
- `error`: Error if token invalid or expired

**Example**:
```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    db := ctx.DB()
    tokenValue := ctx.GetHeader("Authorization")
    
    token, err := db.ValidateAccessToken(tokenValue)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid or expired token"})
    }
    
    // Store token info in context for handler use
    ctx.Set("token", token)
    ctx.Set("user_id", token.UserID)
    
    return next(ctx)
}
```

### DeleteAccessToken

```go
func DeleteAccessToken(tokenValue string) error
```

**Description**: Deletes an access token from the database.

**Parameters**:
- `tokenValue` (string): Token string

**Returns**:
- `error`: Error if deletion fails

**Example**:
```go
func revokeTokenHandler(ctx pkg.Context) error {
    db := ctx.DB()
    tokenValue := ctx.GetHeader("Authorization")
    
    if err := db.DeleteAccessToken(tokenValue); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Token revocation failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Token revoked"})
}
```

### CleanupExpiredTokens

```go
func CleanupExpiredTokens() error
```

**Description**: Removes all expired access tokens from the database.

**Returns**:
- `error`: Error if cleanup fails

**Example**:
```go
// Run as a background task
func tokenCleanupTask(db pkg.DatabaseManager) {
    ticker := time.NewTicker(6 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := db.CleanupExpiredTokens(); err != nil {
            log.Printf("Token cleanup failed: %v", err)
        }
    }
}
```


## Tenant Management

### SaveTenant

```go
func SaveTenant(tenant *Tenant) error
```

**Description**: Saves a tenant to the database for multi-tenant applications.

**Parameters**:
- `tenant` (*Tenant): Tenant object to save

**Returns**:
- `error`: Error if save fails

**Example**:
```go
func createTenantHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    tenant := &pkg.Tenant{
        ID:          generateTenantID(),
        Name:        ctx.FormValue("name"),
        Hosts:       []string{"tenant1.example.com", "tenant1.app"},
        Config:      map[string]interface{}{"theme": "blue"},
        IsActive:    true,
        MaxUsers:    100,
        MaxStorage:  10737418240, // 10GB
        MaxRequests: 1000000,
    }
    
    if err := db.SaveTenant(tenant); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Tenant creation failed"})
    }
    
    return ctx.JSON(201, tenant)
}
```

### LoadTenant

```go
func LoadTenant(tenantID string) (*Tenant, error)
```

**Description**: Loads a tenant from the database by tenant ID.

**Parameters**:
- `tenantID` (string): Tenant identifier

**Returns**:
- `*Tenant`: Tenant object
- `error`: Error if tenant not found

**Example**:
```go
func getTenantHandler(ctx pkg.Context) error {
    db := ctx.DB()
    tenantID := ctx.Param("id")
    
    tenant, err := db.LoadTenant(tenantID)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Tenant not found"})
    }
    
    return ctx.JSON(200, tenant)
}
```

### LoadTenantByHost

```go
func LoadTenantByHost(hostname string) (*Tenant, error)
```

**Description**: Loads a tenant from the database by hostname. Used for host-based tenant routing.

**Parameters**:
- `hostname` (string): Hostname to match against tenant hosts

**Returns**:
- `*Tenant`: Tenant object
- `error`: Error if no tenant found for hostname

**Example**:
```go
func tenantMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    db := ctx.DB()
    hostname := ctx.Request().Host
    
    tenant, err := db.LoadTenantByHost(hostname)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Tenant not found"})
    }
    
    // Store tenant in context
    ctx.Set("tenant", tenant)
    
    return next(ctx)
}
```

**See Also**:
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)


## Metrics and Monitoring

### SaveWorkloadMetrics

```go
func SaveWorkloadMetrics(metrics *WorkloadMetrics) error
```

**Description**: Saves workload metrics to the database for performance monitoring and analysis.

**Parameters**:
- `metrics` (*WorkloadMetrics): Metrics object to save

**Returns**:
- `error`: Error if save fails

**Example**:
```go
func metricsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Execute handler
    err := next(ctx)
    
    // Record metrics
    db := ctx.DB()
    metrics := &pkg.WorkloadMetrics{
        Timestamp:    time.Now(),
        TenantID:     ctx.Tenant().ID,
        UserID:       ctx.User().ID,
        RequestID:    ctx.Request().ID,
        Duration:     time.Since(start).Milliseconds(),
        Path:         ctx.Request().RequestURI,
        Method:       ctx.Request().Method,
        StatusCode:   ctx.Response().StatusCode(),
        ResponseSize: ctx.Response().Size(),
    }
    
    db.SaveWorkloadMetrics(metrics)
    
    return err
}
```

### GetWorkloadMetrics

```go
func GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)
```

**Description**: Retrieves workload metrics for a tenant within a time range.

**Parameters**:
- `tenantID` (string): Tenant identifier
- `from` (time.Time): Start time for metrics range
- `to` (time.Time): End time for metrics range

**Returns**:
- `[]*WorkloadMetrics`: Array of metrics
- `error`: Error if query fails

**Example**:
```go
func metricsReportHandler(ctx pkg.Context) error {
    db := ctx.DB()
    tenantID := ctx.Tenant().ID
    
    // Get metrics for last 24 hours
    from := time.Now().Add(-24 * time.Hour)
    to := time.Now()
    
    metrics, err := db.GetWorkloadMetrics(tenantID, from, to)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Metrics query failed"})
    }
    
    // Calculate statistics
    var totalDuration int64
    for _, m := range metrics {
        totalDuration += m.Duration
    }
    avgDuration := totalDuration / int64(len(metrics))
    
    return ctx.JSON(200, map[string]interface{}{
        "count": len(metrics),
        "avg_duration_ms": avgDuration,
        "metrics": metrics,
    })
}
```

**See Also**:
- [Monitoring Guide](../guides/monitoring.md)


## Rate Limiting

### CheckRateLimit

```go
func CheckRateLimit(key string, limit int, window time.Duration) (bool, error)
```

**Description**: Checks if a rate limit has been exceeded for a given key within a time window.

**Parameters**:
- `key` (string): Rate limit key (e.g., user ID, IP address)
- `limit` (int): Maximum number of requests allowed
- `window` (time.Duration): Time window for the limit

**Returns**:
- `bool`: true if limit exceeded, false otherwise
- `error`: Error if check fails

**Example**:
```go
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    db := ctx.DB()
    
    // Use IP address as rate limit key
    key := ctx.Request().RemoteAddr
    limit := 100
    window := 1 * time.Minute
    
    exceeded, err := db.CheckRateLimit(key, limit, window)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Rate limit check failed"})
    }
    
    if exceeded {
        return ctx.JSON(429, map[string]string{"error": "Rate limit exceeded"})
    }
    
    // Increment counter
    db.IncrementRateLimit(key, window)
    
    return next(ctx)
}
```

### IncrementRateLimit

```go
func IncrementRateLimit(key string, window time.Duration) error
```

**Description**: Increments the rate limit counter for a key and automatically cleans up old entries.

**Parameters**:
- `key` (string): Rate limit key
- `window` (time.Duration): Time window for the limit

**Returns**:
- `error`: Error if increment fails

**Example**:
```go
func apiHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.User().ID
    
    // Increment user's API call counter
    if err := db.IncrementRateLimit(userID, 1*time.Hour); err != nil {
        log.Printf("Failed to increment rate limit: %v", err)
    }
    
    // Process API request...
    return ctx.JSON(200, result)
}
```

**See Also**:
- [Security Guide](../guides/security.md#rate-limiting)


## Migration and Schema Management

### Migrate

```go
func Migrate() error
```

**Description**: Runs database migrations to create or update schema. Currently equivalent to CreateTables().

**Returns**:
- `error`: Error if migration fails

**Example**:
```go
func main() {
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Run migrations on startup
    db := app.DB()
    if err := db.Migrate(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    app.Start()
}
```

### CreateTables

```go
func CreateTables() error
```

**Description**: Creates all framework tables (sessions, tokens, tenants, metrics, rate limits, plugins) in the database.

**Returns**:
- `error`: Error if table creation fails

**Example**:
```go
func setupDatabase(db pkg.DatabaseManager) error {
    if err := db.CreateTables(); err != nil {
        return fmt.Errorf("failed to create tables: %w", err)
    }
    
    log.Println("Database tables created successfully")
    return nil
}
```

**Note**: This method is idempotent and safe to call multiple times. Existing tables are not modified.

### DropTables

```go
func DropTables() error
```

**Description**: Drops all framework tables from the database. Use with caution - this deletes all data.

**Returns**:
- `error`: Error if table dropping fails

**Example**:
```go
func resetDatabase(db pkg.DatabaseManager) error {
    if err := db.DropTables(); err != nil {
        return fmt.Errorf("failed to drop tables: %w", err)
    }
    
    if err := db.CreateTables(); err != nil {
        return fmt.Errorf("failed to create tables: %w", err)
    }
    
    log.Println("Database reset successfully")
    return nil
}
```

**Warning**: This permanently deletes all data in framework tables. Only use in development or testing.

### InitializePluginTables

```go
func InitializePluginTables() error
```

**Description**: Creates plugin system tables (plugins, plugin_hooks, plugin_events, plugin_storage, plugin_metrics).

**Returns**:
- `error`: Error if table creation fails

**Example**:
```go
func setupPluginSystem(db pkg.DatabaseManager) error {
    if err := db.InitializePluginTables(); err != nil {
        return fmt.Errorf("failed to initialize plugin tables: %w", err)
    }
    
    log.Println("Plugin tables initialized")
    return nil
}
```

**See Also**:
- [Plugin Guide](../guides/plugins.md)


## SQL Loader

### GetQuery

```go
func GetQuery(name string) (string, error)
```

**Description**: Retrieves a SQL query by name from the SQL loader. Queries are organized by database driver in the `sql/` directory.

**Parameters**:
- `name` (string): Query name (e.g., "save_session", "load_tenant")

**Returns**:
- `string`: SQL query text
- `error`: Error if query not found

**Example**:
```go
func customQueryHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Get custom query from SQL loader
    query, err := db.GetQuery("custom_report")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query not found"})
    }
    
    // Execute query
    rows, err := db.Query(query, param1, param2)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query failed"})
    }
    defer rows.Close()
    
    // Process results...
    return ctx.JSON(200, results)
}
```

**See Also**:
- [Database Guide](../guides/database.md#sql-loader)


## Data Structures

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver          string            // Database driver: mysql, postgres, mssql, sqlite
    Host            string            // Database server hostname (default: "localhost")
    Port            int               // Database server port (driver-specific defaults)
    Database        string            // Database name (required)
    Username        string            // Database username (required for non-SQLite)
    Password        string            // Database password (required for non-SQLite)
    SSLMode         string            // SSL/TLS mode (optional)
    Charset         string            // Character set (optional)
    Timezone        string            // Timezone (optional)
    ConnMaxLifetime time.Duration     // Max connection lifetime (default: 5 minutes)
    MaxOpenConns    int               // Max open connections (default: 25)
    MaxIdleConns    int               // Max idle connections (default: 5)
    Options         map[string]string // Driver-specific options (optional)
}
```

**Example**:
```go
// PostgreSQL configuration
postgresConfig := pkg.DatabaseConfig{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "myapp",
    Username:        "dbuser",
    Password:        "dbpass",
    SSLMode:         "require",
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 10 * time.Minute,
}

// MySQL configuration
mysqlConfig := pkg.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    Charset:  "utf8mb4",
    Timezone: "UTC",
}

// SQLite configuration
sqliteConfig := pkg.DatabaseConfig{
    Driver:   "sqlite",
    Database: "./myapp.db",
}
```

### DatabaseStats

```go
type DatabaseStats struct {
    OpenConnections   int           // Number of open connections
    InUse             int           // Connections currently in use
    Idle              int           // Idle connections in pool
    WaitCount         int64         // Total number of connection waits
    WaitDuration      time.Duration // Total time waited for connections
    MaxIdleClosed     int64         // Connections closed due to max idle
    MaxLifetimeClosed int64         // Connections closed due to max lifetime
}
```

**Example**:
```go
stats := db.Stats()
log.Printf("Database connections: %d open, %d in use, %d idle", 
    stats.OpenConnections, stats.InUse, stats.Idle)
```


### Session

```go
type Session struct {
    ID        string                 // Unique session identifier
    UserID    string                 // Associated user ID
    TenantID  string                 // Associated tenant ID
    Data      map[string]interface{} // Session data (JSON)
    ExpiresAt time.Time              // Expiration timestamp
    CreatedAt time.Time              // Creation timestamp
    UpdatedAt time.Time              // Last update timestamp
    IPAddress string                 // Client IP address
    UserAgent string                 // Client user agent
}
```

### AccessToken

```go
type AccessToken struct {
    Token     string    // Token value
    UserID    string    // Associated user ID
    TenantID  string    // Associated tenant ID
    Scopes    []string  // Token scopes/permissions
    ExpiresAt time.Time // Expiration timestamp
    CreatedAt time.Time // Creation timestamp
}
```

### Tenant

```go
type Tenant struct {
    ID          string                 // Unique tenant identifier
    Name        string                 // Tenant name
    Hosts       []string               // Associated hostnames
    Config      map[string]interface{} // Tenant configuration (JSON)
    IsActive    bool                   // Active status
    CreatedAt   time.Time              // Creation timestamp
    UpdatedAt   time.Time              // Last update timestamp
    MaxUsers    int                    // Maximum users allowed
    MaxStorage  int64                  // Maximum storage in bytes
    MaxRequests int64                  // Maximum requests allowed
}
```

### WorkloadMetrics

```go
type WorkloadMetrics struct {
    ID           int64     // Unique metric ID
    Timestamp    time.Time // Metric timestamp
    TenantID     string    // Associated tenant ID
    UserID       string    // Associated user ID
    RequestID    string    // Request identifier
    Duration     int64     // Request duration in milliseconds
    ContextSize  int64     // Context size in bytes
    MemoryUsage  int64     // Memory usage in bytes
    CPUUsage     float64   // CPU usage percentage
    Path         string    // Request path
    Method       string    // HTTP method
    StatusCode   int       // HTTP status code
    ResponseSize int64     // Response size in bytes
    ErrorMessage string    // Error message if any
}
```


## Usage Examples

### Complete CRUD Example

```go
package main

import (
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Configure database
    config := pkg.DatabaseConfig{
        Driver:   "postgres",
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        Username: "dbuser",
        Password: "dbpass",
    }
    
    // Create and connect
    db := pkg.NewDatabaseManager()
    if err := db.Connect(config); err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create tables
    if err := db.CreateTables(); err != nil {
        log.Fatal(err)
    }
    
    // Create (INSERT)
    result, err := db.Exec(
        "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
        "John Doe", "john@example.com", time.Now(),
    )
    if err != nil {
        log.Fatal(err)
    }
    userID, _ := result.LastInsertId()
    log.Printf("Created user with ID: %d", userID)
    
    // Read (SELECT)
    row := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userID)
    var id int
    var name, email string
    if err := row.Scan(&id, &name, &email); err != nil {
        log.Fatal(err)
    }
    log.Printf("User: %d - %s (%s)", id, name, email)
    
    // Update
    _, err = db.Exec("UPDATE users SET name = ? WHERE id = ?", "Jane Doe", userID)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("User updated")
    
    // Delete
    _, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("User deleted")
}
```

### Transaction Example

```go
func transferFunds(db pkg.DatabaseManager, fromID, toID string, amount float64) error {
    // Start transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed
    
    // Check source balance
    var balance float64
    row := tx.QueryRow("SELECT balance FROM accounts WHERE id = ?", fromID)
    if err := row.Scan(&balance); err != nil {
        return err
    }
    
    if balance < amount {
        return fmt.Errorf("insufficient funds")
    }
    
    // Deduct from source
    _, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil {
        return err
    }
    
    // Add to destination
    _, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    if err != nil {
        return err
    }
    
    // Record transaction
    _, err = tx.Exec(
        "INSERT INTO transactions (from_id, to_id, amount, timestamp) VALUES (?, ?, ?, ?)",
        fromID, toID, amount, time.Now(),
    )
    if err != nil {
        return err
    }
    
    // Commit transaction
    return tx.Commit()
}
```


### Connection Pool Monitoring

```go
func monitorConnectionPool(db pkg.DatabaseManager) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := db.Stats()
        
        log.Printf("Database Pool Stats:")
        log.Printf("  Open Connections: %d", stats.OpenConnections)
        log.Printf("  In Use: %d", stats.InUse)
        log.Printf("  Idle: %d", stats.Idle)
        log.Printf("  Wait Count: %d", stats.WaitCount)
        log.Printf("  Wait Duration: %s", stats.WaitDuration)
        
        // Alert if pool is exhausted
        if stats.InUse == stats.OpenConnections {
            log.Println("WARNING: Connection pool exhausted!")
        }
        
        // Alert if high wait times
        if stats.WaitDuration > 1*time.Second {
            log.Println("WARNING: High connection wait times!")
        }
    }
}
```

### Multi-Tenant Query Example

```go
func getTenantUsers(ctx pkg.Context) error {
    db := ctx.DB()
    tenant := ctx.Tenant()
    
    if tenant == nil {
        return ctx.JSON(400, map[string]string{"error": "No tenant context"})
    }
    
    // Query users for specific tenant
    rows, err := db.Query(
        "SELECT id, name, email FROM users WHERE tenant_id = ? AND active = ?",
        tenant.ID, true,
    )
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Query failed"})
    }
    defer rows.Close()
    
    var users []map[string]interface{}
    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            continue
        }
        users = append(users, map[string]interface{}{
            "id":    id,
            "name":  name,
            "email": email,
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "tenant": tenant.Name,
        "users":  users,
        "count":  len(users),
    })
}
```


## Best Practices

### Connection Management

1. **Use Connection Pooling**: Configure appropriate pool sizes based on your workload
   ```go
   config.MaxOpenConns = 25  // Adjust based on database server capacity
   config.MaxIdleConns = 5   // Keep some connections ready
   config.ConnMaxLifetime = 5 * time.Minute
   ```

2. **Always Close Resources**: Use `defer` to ensure resources are released
   ```go
   rows, err := db.Query(query)
   if err != nil {
       return err
   }
   defer rows.Close() // Always close rows
   ```

3. **Monitor Connection Pool**: Track pool statistics to identify bottlenecks
   ```go
   stats := db.Stats()
   if stats.WaitCount > threshold {
       // Consider increasing MaxOpenConns
   }
   ```

### Query Execution

1. **Use Parameterized Queries**: Always use placeholders to prevent SQL injection
   ```go
   // Good
   db.Query("SELECT * FROM users WHERE id = ?", userID)
   
   // Bad - vulnerable to SQL injection
   db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))
   ```

2. **Handle Errors Properly**: Check for specific errors like `sql.ErrNoRows`
   ```go
   row := db.QueryRow(query, id)
   err := row.Scan(&data)
   if err == sql.ErrNoRows {
       return ctx.JSON(404, map[string]string{"error": "Not found"})
   }
   if err != nil {
       return ctx.JSON(500, map[string]string{"error": "Database error"})
   }
   ```

3. **Use Prepared Statements for Repeated Queries**: Improves performance
   ```go
   stmt, err := db.Prepare("INSERT INTO logs (message) VALUES (?)")
   defer stmt.Close()
   
   for _, log := range logs {
       stmt.Exec(log.Message)
   }
   ```

### Transaction Management

1. **Always Use defer Rollback**: Ensures cleanup even if commit fails
   ```go
   tx, err := db.Begin()
   if err != nil {
       return err
   }
   defer tx.Rollback() // Safe to call even after Commit()
   
   // ... transaction operations ...
   
   return tx.Commit()
   ```

2. **Keep Transactions Short**: Long transactions can cause lock contention
   ```go
   // Good - short transaction
   tx, _ := db.Begin()
   tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, id)
   tx.Commit()
   
   // Bad - long transaction with external calls
   tx, _ := db.Begin()
   tx.Exec("UPDATE accounts ...")
   callExternalAPI() // Don't do this in a transaction!
   tx.Commit()
   ```

3. **Use Appropriate Isolation Levels**: Choose based on your consistency needs
   ```go
   // Read-only transaction with serializable isolation
   tx, err := db.BeginTx(&sql.TxOptions{
       Isolation: sql.LevelSerializable,
       ReadOnly:  true,
   })
   ```

### Performance Optimization

1. **Use Indexes**: Ensure frequently queried columns are indexed
2. **Batch Operations**: Use transactions for multiple related operations
3. **Limit Result Sets**: Use LIMIT clauses to avoid loading too much data
4. **Use Connection Pooling**: Reuse connections instead of creating new ones
5. **Monitor Slow Queries**: Track query execution times and optimize slow queries

### Security

1. **Never Trust User Input**: Always use parameterized queries
2. **Use Least Privilege**: Database users should have minimal required permissions
3. **Encrypt Sensitive Data**: Use database encryption for sensitive fields
4. **Audit Database Access**: Log all database operations for security auditing
5. **Rotate Credentials**: Regularly update database passwords

## See Also

- [Database Guide](../guides/database.md) - Comprehensive database usage guide
- [Context API](context.md) - Accessing database through context
- [Framework API](framework.md) - Framework initialization and configuration
- [Configuration Guide](../guides/configuration.md) - Database configuration options
- [Performance Guide](../guides/performance.md) - Database performance optimization

