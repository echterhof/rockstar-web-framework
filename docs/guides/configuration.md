# Configuration

## Overview

The Rockstar Web Framework provides a flexible and powerful configuration system that supports multiple file formats, environment variables, and sensible defaults. Configuration can be loaded from JSON, YAML, TOML, or INI files, and can be overridden using environment variables for containerized deployments.

The framework applies intelligent defaults to all configuration values, so you only need to specify what you want to change. This makes it easy to get started quickly while still providing full control when needed.

## Quick Start

The simplest way to configure the framework is to use the built-in defaults:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Minimal configuration - everything else uses defaults
    config := pkg.FrameworkConfig{
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        panic(err)
    }
    
    // Your application code here
}
```

## Configuration File Formats

The framework supports four configuration file formats: JSON, YAML, TOML, and INI. Choose the format that best fits your workflow.

### JSON Configuration

JSON is a widely-supported format that's easy to parse and validate:

```json
{
  "app_name": "MyRockstarApp",
  "port": 8080,
  "debug": true,
  "timeout": "30s",
  "features": ["auth", "api", "websocket"],
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mydb",
    "username": "admin",
    "max_connections": 100
  },
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}
```

### YAML Configuration

YAML provides a clean, human-readable format with support for comments:

```yaml
# Application Configuration
app_name: MyRockstarApp
port: 8080
debug: true
timeout: 30s
features:
  - auth
  - api
  - websocket

database:
  host: localhost
  port: 5432
  name: mydb
  username: admin
  max_connections: 100

cache:
  enabled: yes
  ttl: 3600
  max_size: 1000
```

### TOML Configuration

TOML offers a balance between readability and structure:

```toml
# Application Configuration
app_name = "MyRockstarApp"
port = 8080
debug = true
features = ["auth", "api", "websocket"]

[database]
host = "localhost"
port = 5432
name = "mydb"
max_connections = 100

[cache]
enabled = true
ttl = 3600
```

### INI Configuration

INI format is simple and widely recognized:

```ini
; Application Configuration
app_name = MyRockstarApp
port = 8080
debug = true

[database]
host = localhost
port = 5432
name = mydb
max_connections = 100

[cache]
enabled = true
ttl = 3600
```

## Loading Configuration

### From Files

Use the `ConfigManager` to load configuration from files:

```go
config := pkg.NewConfigManager()

// Load from file (format detected by extension)
err := config.Load("config.yaml")
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Access configuration values
appName := config.GetString("app_name")
port := config.GetInt("port")
debug := config.GetBool("debug")
```

### From Environment Variables

Environment variables provide a way to override configuration without modifying files:

```go
// Set environment variables with ROCKSTAR_ prefix
// ROCKSTAR_APP_NAME=MyApp
// ROCKSTAR_PORT=9090
// ROCKSTAR_DATABASE_HOST=dbhost

config := pkg.NewConfigManager()
err := config.LoadFromEnv()
if err != nil {
    log.Fatalf("Failed to load from env: %v", err)
}

// Access values (underscores converted to dots)
appName := config.GetString("app.name")  // from ROCKSTAR_APP_NAME
port := config.GetInt("port")            // from ROCKSTAR_PORT
dbHost := config.GetString("database.host") // from ROCKSTAR_DATABASE_HOST
```

Environment variable naming convention:
- Prefix: `ROCKSTAR_`
- Nested keys: Use underscores (e.g., `ROCKSTAR_DATABASE_HOST` → `database.host`)
- Case: Converted to lowercase

## Configuration Options

### ServerConfig

Controls HTTP server behavior and resource limits.

```go
type ServerConfig struct {
    EnableHTTP1       bool          // Enable HTTP/1.1 support (default: false)
    EnableHTTP2       bool          // Enable HTTP/2 support (default: false)
    EnableQUIC        bool          // Enable QUIC support (default: false)
    ReadTimeout       time.Duration // Maximum duration for reading request (default: 30s)
    WriteTimeout      time.Duration // Maximum duration for writing response (default: 30s)
    IdleTimeout       time.Duration // Maximum idle time for keep-alive (default: 120s)
    MaxHeaderBytes    int           // Maximum size of request headers (default: 1MB)
    MaxConnections    int           // Maximum concurrent connections (default: 10000)
    MaxRequestSize    int64         // Maximum request body size (default: 10MB)
    ShutdownTimeout   time.Duration // Graceful shutdown timeout (default: 30s)
    ReadBufferSize    int           // TCP read buffer size (default: 4096)
    WriteBufferSize   int           // TCP write buffer size (default: 4096)
}
```

**Example:**

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1:    true,
        EnableHTTP2:    true,
        ReadTimeout:    15 * time.Second,
        WriteTimeout:   15 * time.Second,
        MaxConnections: 5000,
    },
}
```

### DatabaseConfig

Configures database connections and connection pooling.

```go
type DatabaseConfig struct {
    Driver          string        // Database driver: "postgres", "mysql", "mssql", "sqlite"
    Host            string        // Database host (default: "localhost")
    Port            int           // Database port (driver-specific defaults)
    Database        string        // Database name
    Username        string        // Database username
    Password        string        // Database password
    SSLMode         string        // SSL mode for connection
    MaxOpenConns    int           // Maximum open connections (default: 25)
    MaxIdleConns    int           // Maximum idle connections (default: 5)
    ConnMaxLifetime time.Duration // Maximum connection lifetime (default: 5m)
}
```

**Driver-Specific Port Defaults:**
- PostgreSQL: 5432
- MySQL: 3306
- MSSQL: 1433
- SQLite: 0 (not used)

**Example:**

```go
config := pkg.FrameworkConfig{
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:          "postgres",
        Host:            "db.example.com",
        Port:            5432,
        Database:        "myapp",
        Username:        "dbuser",
        Password:        "dbpass",
        SSLMode:         "require",
        MaxOpenConns:    50,
        MaxIdleConns:    10,
        ConnMaxLifetime: 10 * time.Minute,
    },
}
```

### CacheConfig

Configures caching behavior and limits.

```go
type CacheConfig struct {
    Type       string        // Cache type: "memory" (default)
    MaxSize    int64         // Maximum cache size in bytes (default: 0 = unlimited)
    DefaultTTL time.Duration // Default time-to-live (default: 0 = no expiration)
}
```

**Example:**

```go
config := pkg.FrameworkConfig{
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    100 * 1024 * 1024, // 100MB
        DefaultTTL: 1 * time.Hour,
    },
}
```

### SessionConfig

Configures session management and security.

```go
type SessionConfig struct {
    EncryptionKey   []byte        // AES-256 encryption key (default: random 32 bytes)
    CookieName      string        // Session cookie name (default: "rockstar_session")
    CookiePath      string        // Cookie path (default: "/")
    CookieDomain    string        // Cookie domain
    SessionLifetime time.Duration // Session lifetime (default: 24h)
    CleanupInterval time.Duration // Cleanup interval (default: 1h)
    CookieHTTPOnly  bool          // HTTPOnly flag (default: true)
    CookieSecure    bool          // Secure flag (default: true)
    CookieSameSite  string        // SameSite attribute (default: "Lax")
    FilesystemPath  string        // Filesystem storage path (default: "./sessions")
}
```

**Important:** The default `EncryptionKey` is randomly generated, which means sessions won't persist across application restarts. For production, provide a fixed encryption key.

**Example:**

```go
config := pkg.FrameworkConfig{
    SessionConfig: pkg.SessionConfig{
        EncryptionKey:   []byte("your-32-byte-encryption-key!!"),
        CookieName:      "myapp_session",
        SessionLifetime: 48 * time.Hour,
        CookieSecure:    true,
        CookieHTTPOnly:  true,
        CookieSameSite:  "Strict",
    },
}
```

### MonitoringConfig

Configures monitoring, metrics, and profiling endpoints.

```go
type MonitoringConfig struct {
    EnableMetrics        bool          // Enable metrics collection
    EnablePprof          bool          // Enable pprof profiling
    EnableSNMP           bool          // Enable SNMP monitoring
    MetricsPort          int           // Metrics endpoint port (default: 9090)
    PprofPort            int           // Pprof endpoint port (default: 6060)
    SNMPPort             int           // SNMP port (default: 161)
    SNMPCommunity        string        // SNMP community string (default: "public")
    OptimizationInterval time.Duration // Optimization interval (default: 5m)
}
```

**Example:**

```go
monitoringConfig := pkg.MonitoringConfig{
    EnableMetrics:        true,
    EnablePprof:          true,
    MetricsPort:          9090,
    PprofPort:            6060,
    OptimizationInterval: 10 * time.Minute,
}
```

## Accessing Configuration Values

The `ConfigManager` provides type-safe methods for accessing configuration:

### Basic Types

```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// String values
appName := config.GetString("app_name")

// Integer values
port := config.GetInt("port")
maxConns := config.GetInt64("max_connections")

// Float values
ratio := config.GetFloat64("ratio")

// Boolean values
debug := config.GetBool("debug")

// Duration values
timeout := config.GetDuration("timeout")

// String slices
features := config.GetStringSlice("features")
```

### With Default Values

Provide fallback values when configuration keys are not set:

```go
// Get with defaults
appName := config.GetStringWithDefault("app_name", "DefaultApp")
port := config.GetIntWithDefault("port", 8080)
debug := config.GetBoolWithDefault("debug", false)
timeout := config.GetWithDefault("timeout", 30*time.Second)
```

### Nested Configuration

Access nested configuration using dot notation or the `Sub()` method:

```go
// Using dot notation
dbHost := config.GetString("database.host")
dbPort := config.GetInt("database.port")

// Using Sub() for cleaner code
dbConfig := config.Sub("database")
host := dbConfig.GetString("host")
port := dbConfig.GetInt("port")
username := dbConfig.GetString("username")

// Deeply nested
poolConfig := dbConfig.Sub("pool")
minSize := poolConfig.GetInt("min_size")
maxSize := poolConfig.GetInt("max_size")
```

### Checking if Keys Exist

```go
if config.IsSet("database") {
    // Database configuration is present
}

if config.IsSet("cache.redis.host") {
    // Redis cache is configured
}
```

## Environment-Specific Configuration

The framework supports environment-specific configuration through the `ROCKSTAR_ENV` or `ENV` environment variable:

```go
config := pkg.NewConfigManager()

// Check current environment
env := config.GetEnv() // "development", "production", "test"

if config.IsProduction() {
    // Production-specific logic
}

if config.IsDevelopment() {
    // Development-specific logic
}

if config.IsTest() {
    // Test-specific logic
}
```

**Best Practice:** Load different configuration files based on environment:

```go
config := pkg.NewConfigManager()

env := config.GetEnv()
configFile := fmt.Sprintf("config.%s.yaml", env)

if err := config.Load(configFile); err != nil {
    // Fallback to default config
    config.Load("config.yaml")
}
```

## Configuration Validation

Validate configuration before starting your application:

```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}

// Check required keys
if !config.IsSet("database.host") {
    log.Fatal("Database host is required")
}

if !config.IsSet("database.name") {
    log.Fatal("Database name is required")
}
```

## Configuration Watching

Automatically reload configuration when files change:

```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Watch for changes
err := config.Watch(func() {
    log.Println("Configuration changed, reloading...")
    
    // Access updated values
    newPort := config.GetInt("port")
    log.Printf("New port: %d", newPort)
    
    // Trigger application reconfiguration
    // (implementation depends on your application)
})

if err != nil {
    log.Fatalf("Failed to watch config: %v", err)
}

// Stop watching when done
defer config.StopWatching()
```

The watcher checks for file modifications every 5 seconds and calls the callback function when changes are detected.

## Complete Configuration Example

Here's a comprehensive example showing all configuration options:

```yaml
# config.yaml - Complete configuration example

# Application settings
app_name: MyRockstarApp
environment: production
debug: false

# Server configuration
server:
  http1_enabled: true
  http2_enabled: true
  quic_enabled: false
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576
  max_connections: 10000
  max_request_size: 10485760
  shutdown_timeout: 30s
  read_buffer_size: 4096
  write_buffer_size: 4096

# Database configuration
database:
  driver: postgres
  host: db.example.com
  port: 5432
  database: myapp
  username: dbuser
  password: ${DB_PASSWORD}  # Use environment variable
  ssl_mode: require
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 10m

# Cache configuration
cache:
  type: memory
  max_size: 104857600  # 100MB
  default_ttl: 1h

# Session configuration
session:
  encryption_key: ${SESSION_KEY}  # 32-byte key from environment
  cookie_name: myapp_session
  cookie_path: /
  cookie_domain: example.com
  session_lifetime: 24h
  cleanup_interval: 1h
  cookie_http_only: true
  cookie_secure: true
  cookie_same_site: Strict

# Monitoring configuration
monitoring:
  enable_metrics: true
  enable_pprof: true
  enable_snmp: false
  metrics_port: 9090
  pprof_port: 6060
  optimization_interval: 5m

# Plugin configuration
plugins:
  enabled: true
  directory: ./plugins
  plugins:
    - name: auth-plugin
      enabled: true
      priority: 10
    - name: cache-plugin
      enabled: true
      priority: 20
```

## Troubleshooting

### Configuration Not Loading

**Problem:** Configuration file fails to load.

**Solutions:**
- Verify the file path is correct
- Check file permissions
- Ensure the file format matches the extension
- Validate JSON/YAML/TOML syntax

```go
config := pkg.NewConfigManager()
if err := config.Load("config.yaml"); err != nil {
    log.Printf("Failed to load config: %v", err)
    // Check if file exists
    if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
        log.Fatal("Config file does not exist")
    }
}
```

### Environment Variables Not Working

**Problem:** Environment variables are not being loaded.

**Solutions:**
- Ensure variables have the `ROCKSTAR_` prefix
- Check variable names (underscores convert to dots)
- Verify environment variables are set before loading

```bash
# Correct format
export ROCKSTAR_DATABASE_HOST=localhost
export ROCKSTAR_DATABASE_PORT=5432

# Incorrect (missing prefix)
export DATABASE_HOST=localhost  # Won't be loaded
```

### Type Conversion Errors

**Problem:** Configuration values have wrong types.

**Solutions:**
- Use appropriate getter methods (`GetInt`, `GetBool`, etc.)
- Provide default values with `GetWithDefault` methods
- Validate configuration after loading

```go
// Wrong: returns 0 if not an integer
port := config.GetInt("port_string")

// Better: provide default
port := config.GetIntWithDefault("port", 8080)

// Best: check if set first
if config.IsSet("port") {
    port := config.GetInt("port")
} else {
    port := 8080
}
```

### Nested Configuration Access

**Problem:** Cannot access deeply nested values.

**Solutions:**
- Use dot notation for nested keys
- Use `Sub()` method for cleaner code
- Check if parent keys exist first

```go
// Dot notation
host := config.GetString("database.pool.host")

// Using Sub()
dbConfig := config.Sub("database")
poolConfig := dbConfig.Sub("pool")
host := poolConfig.GetString("host")

// Check existence
if config.IsSet("database.pool") {
    poolConfig := config.Sub("database").Sub("pool")
    // Access pool configuration
}
```

## Best Practices

1. **Use Environment Variables for Secrets:** Never commit sensitive data like passwords or API keys to configuration files. Use environment variables instead.

2. **Provide Sensible Defaults:** Use the framework's default values and only override what you need to change.

3. **Validate Configuration Early:** Validate configuration at application startup to catch errors before they cause runtime issues.

4. **Use Environment-Specific Files:** Maintain separate configuration files for development, staging, and production environments.

5. **Document Custom Configuration:** If you add custom configuration keys, document them clearly for your team.

6. **Use Type-Safe Getters:** Always use the appropriate getter method (`GetInt`, `GetBool`, etc.) to avoid type conversion issues.

7. **Handle Missing Configuration Gracefully:** Use default values and check if keys are set before accessing them.

8. **Keep Configuration DRY:** Use nested configuration and the `Sub()` method to avoid repeating prefixes.

## See Also

- [Getting Started Guide](../GETTING_STARTED.md) - Basic framework setup
- [Database Guide](database.md) - Database configuration details
- [Security Guide](security.md) - Security configuration
- [Deployment Guide](deployment.md) - Production configuration
- [API Reference: ConfigManager](../api/framework.md) - Complete API documentation
