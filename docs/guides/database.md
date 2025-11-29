# Database Integration

## Overview

The Rockstar Web Framework provides a powerful database abstraction layer that supports multiple database systems with a unified interface. The framework handles connection pooling, transaction management, and provides a SQL loader system for organizing queries.

**Key Features:**
- Support for MySQL, PostgreSQL, MSSQL, and SQLite
- Connection pooling with configurable limits
- Transaction support with isolation levels
- SQL query organization through external files
- Built-in session, token, and tenant management
- Rate limiting support
- Workload metrics tracking

## Supported Database Systems

### SQLite
Lightweight, file-based database ideal for development and small applications.

**Use Cases:**
- Development and testing
- Small to medium applications
- Embedded applications
- Single-server deployments

**Connection String:** File path to database file

### MySQL
Popular open-source relational database with excellent performance.

**Use Cases:**
- Web applications
- Content management systems
- E-commerce platforms
- High-read workloads

**Default Port:** 3306

### PostgreSQL
Advanced open-source database with strong standards compliance.

**Use Cases:**
- Complex queries and analytics
- Applications requiring advanced features
- Data integrity critical applications
- Geographic data (PostGIS)

**Default Port:** 5432

### Microsoft SQL Server (MSSQL)
Enterprise database system with Windows integration.

**Use Cases:**
- Enterprise applications
- Windows-based infrastructure
- .NET application integration
- Business intelligence

**Default Port:** 1433

## Configuration

### Basic Configuration

Configure the database in your `FrameworkConfig`:

```go
config := pkg.FrameworkConfig{
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "myapp.db",
    },
}
```

### SQLite Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "sqlite",
    Database: "./data/myapp.db",  // File path
    
    // Connection pool settings
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    
    // SQLite-specific options
    Options: map[string]string{
        "_journal_mode": "WAL",      // Write-Ahead Logging
        "_foreign_keys": "ON",       // Enable foreign keys
        "_busy_timeout": "5000",     // Wait 5s when locked
    },
}
```

### MySQL Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    
    // Connection pool settings
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    
    // MySQL-specific options
    Charset:  "utf8mb4",
    Timezone: "UTC",
    Options: map[string]string{
        "parseTime": "true",
        "loc":       "UTC",
    },
}
```

### PostgreSQL Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    
    // Connection pool settings
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    
    // PostgreSQL-specific options
    SSLMode:  "disable",  // Use "require" in production
    Timezone: "UTC",
}
```

### MSSQL Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "mssql",
    Host:     "localhost",
    Port:     1433,
    Database: "myapp",
    Username: "sa",
    Password: "YourStrong!Passw0rd",
    
    // Connection pool settings
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Driver` | string | *required* | Database driver: "sqlite", "mysql", "postgres", "mssql" |
| `Database` | string | *required* | Database name or file path (SQLite) |
| `Host` | string | "localhost" | Database server hostname |
| `Port` | int | driver-specific | Database server port |
| `Username` | string | "" | Database username (not needed for SQLite) |
| `Password` | string | "" | Database password (not needed for SQLite) |
| `SSLMode` | string | "" | SSL/TLS mode (PostgreSQL) |
| `Charset` | string | "" | Character set (MySQL) |
| `Timezone` | string | "" | Timezone for connections |
| `MaxOpenConns` | int | 25 | Maximum open connections |
| `MaxIdleConns` | int | 5 | Maximum idle connections in pool |
| `ConnMaxLifetime` | duration | 5m | Maximum connection lifetime |
| `Options` | map[string]string | nil | Driver-specific options |

## Accessing the Database

Access the database manager through the context in your handlers:

```go
func myHandler(ctx pkg.Context) error {
    // Get database manager
    db := ctx.DB()
    
    // Use database operations
    // ...
    
    return nil
}
```

## CRUD Operations

### Query (SELECT)

Execute queries that return multiple rows:

```go
func listUsersHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Execute query
    rows, err := db.Query("SELECT id, name, email FROM users WHERE active = ?", true)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    // Scan results
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            return err
        }
        users = append(users, user)
    }
    
    // Check for errors during iteration
    if err := rows.Err(); err != nil {
        return err
    }
    
    return ctx.JSON(200, users)
}
```

### QueryRow (SELECT single row)

Execute queries that return at most one row:

```go
func getUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    // Execute query for single row
    row := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userID)
    
    // Scan result
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err == sql.ErrNoRows {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, user)
}
```

### Exec (INSERT, UPDATE, DELETE)

Execute queries that modify data:

```go
func createUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Parse request body
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return err
    }
    
    // Execute insert
    result, err := db.Exec(
        "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
        user.Name, user.Email, time.Now(),
    )
    if err != nil {
        return err
    }
    
    // Get inserted ID
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    user.ID = id
    
    return ctx.JSON(201, user)
}

func updateUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    // Parse request body
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return err
    }
    
    // Execute update
    result, err := db.Exec(
        "UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?",
        user.Name, user.Email, time.Now(), userID,
    )
    if err != nil {
        return err
    }
    
    // Check if row was updated
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return ctx.JSON(200, user)
}

func deleteUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    // Execute delete
    result, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        return err
    }
    
    // Check if row was deleted
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "User deleted"})
}
```

## Transaction Management

### Basic Transactions

```go
func transferFundsHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Begin transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    // Ensure transaction is rolled back on error
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // Deduct from source account
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance - ? WHERE id = ?",
        amount, fromAccountID,
    )
    if err != nil {
        return err
    }
    
    // Add to destination account
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance + ? WHERE id = ?",
        amount, toAccountID,
    )
    if err != nil {
        return err
    }
    
    // Commit transaction
    if err = tx.Commit(); err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Transfer successful"})
}
```

### Transactions with Isolation Levels

```go
func criticalOperationHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Begin transaction with specific isolation level
    tx, err := db.BeginTx(&sql.TxOptions{
        Isolation: sql.LevelSerializable,
        ReadOnly:  false,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Perform operations
    // ...
    
    // Commit
    return tx.Commit()
}
```

### Isolation Levels

- `sql.LevelDefault` - Use database default
- `sql.LevelReadUncommitted` - Lowest isolation, highest performance
- `sql.LevelReadCommitted` - Prevents dirty reads
- `sql.LevelRepeatableRead` - Prevents non-repeatable reads
- `sql.LevelSerializable` - Highest isolation, prevents phantom reads

## Prepared Statements

Use prepared statements for repeated queries with different parameters:

```go
func batchInsertHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Prepare statement
    stmt, err := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    // Execute multiple times
    for _, user := range users {
        _, err := stmt.Exec(user.Name, user.Email)
        if err != nil {
            return err
        }
    }
    
    return ctx.JSON(200, map[string]string{"message": "Users created"})
}
```

## SQL Loader System

The framework provides a SQL loader system for organizing queries in external files.

### Directory Structure

```
sql/
├── sqlite/
│   ├── create_users_table.sql
│   ├── save_user.sql
│   ├── load_user.sql
│   └── delete_user.sql
├── mysql/
│   ├── create_users_table.sql
│   ├── save_user.sql
│   ├── load_user.sql
│   └── delete_user.sql
├── postgres/
│   └── ...
└── mssql/
    └── ...
```

### Query Files

Each database driver has its own directory with SQL files. The filename (without `.sql`) becomes the query name.

**Example: `sql/sqlite/save_user.sql`**
```sql
-- Save or update a user
INSERT INTO users (id, name, email, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    email = excluded.email,
    updated_at = excluded.updated_at;
```

### Using SQL Loader

```go
func saveUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Load query by name
    query, err := db.GetQuery("save_user")
    if err != nil {
        return err
    }
    
    // Execute query
    _, err = db.Exec(query, user.ID, user.Name, user.Email, 
        time.Now(), time.Now())
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, user)
}
```

### Custom SQL Directory

Configure a custom SQL directory:

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "sqlite",
    Database: "myapp.db",
    Options: map[string]string{
        "sql_dir": "./custom/sql/path",
    },
}
```

## Built-in Framework Operations

The framework provides built-in database operations for sessions, tokens, and tenants.

### Session Management

```go
// Save session
session := &pkg.Session{
    ID:        "session-123",
    UserID:    "user-456",
    TenantID:  "tenant-789",
    Data:      map[string]interface{}{"key": "value"},
    ExpiresAt: time.Now().Add(24 * time.Hour),
}
err := db.SaveSession(session)

// Load session
session, err := db.LoadSession("session-123")

// Delete session
err := db.DeleteSession("session-123")

// Cleanup expired sessions
err := db.CleanupExpiredSessions()
```

### Token Management

```go
// Save access token
token := &pkg.AccessToken{
    Token:     "token-abc",
    UserID:    "user-456",
    TenantID:  "tenant-789",
    Scopes:    []string{"read", "write"},
    ExpiresAt: time.Now().Add(1 * time.Hour),
}
err := db.SaveAccessToken(token)

// Load token
token, err := db.LoadAccessToken("token-abc")

// Validate token (checks expiration)
token, err := db.ValidateAccessToken("token-abc")

// Delete token
err := db.DeleteAccessToken("token-abc")

// Cleanup expired tokens
err := db.CleanupExpiredTokens()
```

### Tenant Management

```go
// Save tenant
tenant := &pkg.Tenant{
    ID:       "tenant-789",
    Name:     "Acme Corp",
    Hosts:    []string{"acme.example.com"},
    IsActive: true,
    Config:   map[string]interface{}{"theme": "dark"},
}
err := db.SaveTenant(tenant)

// Load tenant by ID
tenant, err := db.LoadTenant("tenant-789")

// Load tenant by hostname
tenant, err := db.LoadTenantByHost("acme.example.com")
```

### Rate Limiting

```go
// Check rate limit
exceeded, err := db.CheckRateLimit("user:123", 100, 1*time.Hour)
if exceeded {
    return ctx.JSON(429, map[string]string{"error": "Rate limit exceeded"})
}

// Increment rate limit counter
err := db.IncrementRateLimit("user:123", 1*time.Hour)
```

### Workload Metrics

```go
// Save metrics
metrics := &pkg.WorkloadMetrics{
    Timestamp:    time.Now(),
    TenantID:     "tenant-789",
    UserID:       "user-456",
    RequestID:    "req-123",
    Duration:     150,  // milliseconds
    Path:         "/api/users",
    Method:       "GET",
    StatusCode:   200,
    ResponseSize: 1024,
}
err := db.SaveWorkloadMetrics(metrics)

// Get metrics for time range
from := time.Now().Add(-24 * time.Hour)
to := time.Now()
metrics, err := db.GetWorkloadMetrics("tenant-789", from, to)
```

## Connection Management

### Connection Statistics

```go
func healthCheckHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Get connection statistics
    stats := db.Stats()
    
    return ctx.JSON(200, map[string]interface{}{
        "open_connections":    stats.OpenConnections,
        "in_use":             stats.InUse,
        "idle":               stats.Idle,
        "wait_count":         stats.WaitCount,
        "wait_duration":      stats.WaitDuration.String(),
        "max_idle_closed":    stats.MaxIdleClosed,
        "max_lifetime_closed": stats.MaxLifetimeClosed,
    })
}
```

### Connection Health

```go
func databaseHealthHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Ping database
    if err := db.Ping(); err != nil {
        return ctx.JSON(503, map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        })
    }
    
    return ctx.JSON(200, map[string]string{
        "status": "healthy",
    })
}
```

## Migration and Schema Management

### Creating Tables

```go
// Create all framework tables
err := db.CreateTables()

// Or run migrations
err := db.Migrate()
```

### Custom Migrations

Create migration files in your SQL directory:

**`sql/sqlite/create_users_table.sql`**
```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Load and execute:

```go
func runMigrations(db pkg.DatabaseManager) error {
    // Load migration query
    query, err := db.GetQuery("create_users_table")
    if err != nil {
        return err
    }
    
    // Execute migration
    _, err = db.Exec(query)
    return err
}
```

### Dropping Tables

```go
// Drop all framework tables
err := db.DropTables()
```

## Best Practices

### 1. Use Connection Pooling

Configure appropriate pool sizes based on your workload:

```go
DatabaseConfig: pkg.DatabaseConfig{
    MaxOpenConns:    25,  // Limit total connections
    MaxIdleConns:    5,   // Keep some connections ready
    ConnMaxLifetime: 5 * time.Minute,  // Recycle connections
}
```

### 2. Always Close Resources

```go
rows, err := db.Query("SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()  // Always close rows
```

### 3. Use Transactions for Multiple Operations

```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()  // Rollback if not committed

// Multiple operations
// ...

return tx.Commit()
```

### 4. Handle sql.ErrNoRows

```go
row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
err := row.Scan(&user)
if err == sql.ErrNoRows {
    return ctx.JSON(404, map[string]string{"error": "Not found"})
}
if err != nil {
    return err
}
```

### 5. Use Prepared Statements for Repeated Queries

```go
stmt, err := db.Prepare("INSERT INTO logs (message) VALUES (?)")
if err != nil {
    return err
}
defer stmt.Close()

for _, msg := range messages {
    stmt.Exec(msg)
}
```

### 6. Organize Queries in SQL Files

Keep SQL separate from Go code for better maintainability:

```
sql/
├── sqlite/
│   ├── users/
│   │   ├── create_table.sql
│   │   ├── insert.sql
│   │   ├── update.sql
│   │   └── delete.sql
│   └── posts/
│       └── ...
```

### 7. Use Context for Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.QueryContext(ctx, "SELECT * FROM large_table")
```

### 8. Monitor Connection Pool

```go
stats := db.Stats()
if stats.WaitCount > 1000 {
    // Consider increasing MaxOpenConns
}
```

## Troubleshooting

### Connection Refused

**Problem:** Cannot connect to database server

**Solutions:**
- Verify database server is running
- Check host and port configuration
- Verify firewall rules
- Check network connectivity

### Too Many Connections

**Problem:** "too many connections" error

**Solutions:**
- Reduce `MaxOpenConns` in configuration
- Ensure connections are properly closed
- Check for connection leaks
- Increase database server connection limit

### Slow Queries

**Problem:** Database operations are slow

**Solutions:**
- Add appropriate indexes
- Optimize query structure
- Use prepared statements
- Enable query logging
- Monitor connection pool statistics

### Deadlocks

**Problem:** Transaction deadlocks

**Solutions:**
- Use consistent lock ordering
- Keep transactions short
- Use appropriate isolation levels
- Implement retry logic

### Connection Pool Exhaustion

**Problem:** All connections in use

**Solutions:**
- Increase `MaxOpenConns`
- Reduce `ConnMaxLifetime`
- Check for long-running queries
- Ensure proper resource cleanup

## See Also

- [Configuration Guide](configuration.md) - Database configuration options
- [Sessions Guide](sessions.md) - Session management with database storage
- [Multi-Tenancy Guide](multi-tenancy.md) - Tenant-specific database access
- [API Reference: Database](../api/database.md) - Complete API documentation
