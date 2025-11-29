# Configuration API Reference

## Overview

The Rockstar Web Framework provides a flexible configuration system supporting multiple formats (JSON, YAML, TOML, INI), environment variables, hot-reloading, and nested configuration access.

## ConfigManager Interface

Interface for managing application configuration.

```go
type ConfigManager interface {
    // Loading
    Load(configPath string) error
    LoadFromEnv() error
    Reload() error
    
    // Getting values
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetInt64(key string) int64
    GetFloat64(key string) float64
    GetBool(key string) bool
    GetDuration(key string) time.Duration
    GetStringSlice(key string) []string
    
    // Getting with defaults
    GetWithDefault(key string, defaultValue interface{}) interface{}
    GetStringWithDefault(key string, defaultValue string) string
    GetIntWithDefault(key string, defaultValue int) int
    GetBoolWithDefault(key string, defaultValue bool) bool
    
    // Validation and checking
    Validate() error
    IsSet(key string) bool
    
    // Environment
    GetEnv() string
    IsProduction() bool
    IsDevelopment() bool
    IsTest() bool
    
    // Sub-configuration
    Sub(key string) ConfigManager
    
    // Watching
    Watch(callback func()) error
    StopWatching() error
}
```

## Factory Function

### NewConfigManager()

Creates a new configuration manager.

**Signature:**
```go
func NewConfigManager() ConfigManager
```

**Returns:**
- `ConfigManager` - Configuration manager instance

**Example:**
```go
config := pkg.NewConfigManager()
```

## Loading Configuration

### Load()

Loads configuration from a file. Supports JSON, YAML, TOML, and INI formats.

**Signature:**
```go
Load(configPath string) error
```

**Parameters:**
- `configPath` - Path to configuration file

**Returns:**
- `error` - Error if loading fails

**Example:**
```go
config := pkg.NewConfigManager()

// Load from YAML
err := config.Load("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Load from JSON
err = config.Load("config.json")

// Load from TOML
err = config.Load("config.toml")

// Load from INI
err = config.Load("config.ini")
```

### LoadFromEnv()

Loads configuration from environment variables with `ROCKSTAR_` prefix.

**Signature:**
```go
LoadFromEnv() error
```

**Returns:**
- `error` - Error if loading fails

**Environment Variable Format:**
- Prefix: `ROCKSTAR_`
- Separator: `_` (converted to `.` in config keys)
- Example: `ROCKSTAR_DATABASE_HOST` → `database.host`

**Example:**
```go
// Set environment variables
// ROCKSTAR_DATABASE_HOST=localhost
// ROCKSTAR_DATABASE_PORT=5432
// ROCKSTAR_DEBUG=true

config := pkg.NewConfigManager()
err := config.LoadFromEnv()

host := config.GetString("database.host")     // "localhost"
port := config.GetInt("database.port")        // 5432
debug := config.GetBool("debug")              // true
```

### Reload()

Reloads configuration from the original source.

**Signature:**
```go
Reload() error
```

**Returns:**
- `error` - Error if reloading fails

**Example:**
```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Later, reload configuration
err := config.Reload()
if err != nil {
    log.Printf("Failed to reload config: %v", err)
}
```

## Getting Values

### Get()

Retrieves a configuration value as interface{}.

**Signature:**
```go
Get(key string) interface{}
```

**Parameters:**
- `key` - Configuration key (supports nested keys with dots)

**Returns:**
- `interface{}` - Configuration value, or nil if not found

**Example:**
```go
value := config.Get("database.host")
value := config.Get("server.port")
value := config.Get("nested.key.value")
```

### GetString()

Retrieves a string configuration value.

**Signature:**
```go
GetString(key string) string
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `string` - String value, or empty string if not found

**Example:**
```go
host := config.GetString("database.host")
name := config.GetString("app.name")
```

### GetInt()

Retrieves an integer configuration value.

**Signature:**
```go
GetInt(key string) int
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `int` - Integer value, or 0 if not found

**Example:**
```go
port := config.GetInt("server.port")
maxConnections := config.GetInt("database.max_connections")
```

### GetInt64()

Retrieves an int64 configuration value.

**Signature:**
```go
GetInt64(key string) int64
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `int64` - Int64 value, or 0 if not found

**Example:**
```go
maxSize := config.GetInt64("upload.max_size")
```

### GetFloat64()

Retrieves a float64 configuration value.

**Signature:**
```go
GetFloat64(key string) float64
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `float64` - Float64 value, or 0.0 if not found

**Example:**
```go
timeout := config.GetFloat64("request.timeout")
rate := config.GetFloat64("rate_limit.per_second")
```

### GetBool()

Retrieves a boolean configuration value.

**Signature:**
```go
GetBool(key string) bool
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `bool` - Boolean value, or false if not found

**Supports:** `true`, `false`, `yes`, `no`, `on`, `off`, `1`, `0`, `t`, `f`, `y`, `n`

**Example:**
```go
debug := config.GetBool("debug")
enabled := config.GetBool("feature.enabled")
```

### GetDuration()

Retrieves a duration configuration value.

**Signature:**
```go
GetDuration(key string) time.Duration
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `time.Duration` - Duration value, or 0 if not found

**Supports:** Go duration strings (e.g., "5s", "10m", "1h")

**Example:**
```go
timeout := config.GetDuration("request.timeout")  // "30s"
interval := config.GetDuration("poll.interval")   // "5m"
```

### GetStringSlice()

Retrieves a string slice configuration value.

**Signature:**
```go
GetStringSlice(key string) []string
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `[]string` - String slice, or nil if not found

**Example:**
```go
// YAML: allowed_hosts: ["host1.com", "host2.com"]
hosts := config.GetStringSlice("allowed_hosts")

// Or comma-separated string: "host1.com,host2.com"
hosts := config.GetStringSlice("allowed_hosts")
```

## Getting with Defaults

### GetWithDefault()

Retrieves a value with a default fallback.

**Signature:**
```go
GetWithDefault(key string, defaultValue interface{}) interface{}
```

**Example:**
```go
value := config.GetWithDefault("optional.key", "default")
```

### GetStringWithDefault()

Retrieves a string with a default fallback.

**Signature:**
```go
GetStringWithDefault(key string, defaultValue string) string
```

**Example:**
```go
host := config.GetStringWithDefault("database.host", "localhost")
name := config.GetStringWithDefault("app.name", "MyApp")
```

### GetIntWithDefault()

Retrieves an int with a default fallback.

**Signature:**
```go
GetIntWithDefault(key string, defaultValue int) int
```

**Example:**
```go
port := config.GetIntWithDefault("server.port", 8080)
maxConns := config.GetIntWithDefault("database.max_connections", 10)
```

### GetBoolWithDefault()

Retrieves a bool with a default fallback.

**Signature:**
```go
GetBoolWithDefault(key string, defaultValue bool) bool
```

**Example:**
```go
debug := config.GetBoolWithDefault("debug", false)
enabled := config.GetBoolWithDefault("feature.enabled", true)
```

## Validation and Checking

### Validate()

Validates the configuration.

**Signature:**
```go
Validate() error
```

**Returns:**
- `error` - Error if validation fails

**Example:**
```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

### IsSet()

Checks if a configuration key is set.

**Signature:**
```go
IsSet(key string) bool
```

**Parameters:**
- `key` - Configuration key

**Returns:**
- `bool` - true if key exists, false otherwise

**Example:**
```go
if config.IsSet("database.host") {
    host := config.GetString("database.host")
} else {
    // Use default
}
```

## Environment Methods

### GetEnv()

Returns the current environment.

**Signature:**
```go
GetEnv() string
```

**Returns:**
- `string` - Environment name (from `ROCKSTAR_ENV` or `ENV` variable)

**Example:**
```go
env := config.GetEnv()  // "development", "production", "test"
```

### IsProduction()

Checks if running in production.

**Signature:**
```go
IsProduction() bool
```

**Returns:**
- `bool` - true if environment is "production"

**Example:**
```go
if config.IsProduction() {
    // Production-specific logic
}
```

### IsDevelopment()

Checks if running in development.

**Signature:**
```go
IsDevelopment() bool
```

**Returns:**
- `bool` - true if environment is "development"

**Example:**
```go
if config.IsDevelopment() {
    // Enable debug logging
}
```

### IsTest()

Checks if running in test mode.

**Signature:**
```go
IsTest() bool
```

**Returns:**
- `bool` - true if environment is "test"

**Example:**
```go
if config.IsTest() {
    // Use test database
}
```

## Sub-Configuration

### Sub()

Returns a sub-configuration manager for a nested key.

**Signature:**
```go
Sub(key string) ConfigManager
```

**Parameters:**
- `key` - Configuration key prefix

**Returns:**
- `ConfigManager` - Sub-configuration manager

**Example:**
```go
// config.yaml:
// database:
//   host: localhost
//   port: 5432
//   credentials:
//     username: admin
//     password: secret

dbConfig := config.Sub("database")
host := dbConfig.GetString("host")           // "localhost"
port := dbConfig.GetInt("port")              // 5432

credsConfig := dbConfig.Sub("credentials")
username := credsConfig.GetString("username") // "admin"
```

## Configuration Watching

### Watch()

Watches for configuration file changes and calls callback on change.

**Signature:**
```go
Watch(callback func()) error
```

**Parameters:**
- `callback` - Function to call when configuration changes

**Returns:**
- `error` - Error if watching fails

**Example:**
```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Watch for changes
config.Watch(func() {
    log.Println("Configuration reloaded")
    // Update application settings
})
```

### StopWatching()

Stops watching for configuration changes.

**Signature:**
```go
StopWatching() error
```

**Returns:**
- `error` - Error if stopping fails

**Example:**
```go
config.StopWatching()
```

## Configuration File Formats

### YAML Example

```yaml
# config.yaml
app:
  name: MyApp
  version: 1.0.0

server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s

database:
  driver: postgres
  host: localhost
  port: 5432
  database: myapp
  username: admin
  password: secret
  max_connections: 10

features:
  enabled:
    - feature1
    - feature2
  debug: true
```

### JSON Example

```json
{
  "app": {
    "name": "MyApp",
    "version": "1.0.0"
  },
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "timeout": "30s"
  },
  "database": {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432
  }
}
```

### TOML Example

```toml
[app]
name = "MyApp"
version = "1.0.0"

[server]
host = "0.0.0.0"
port = 8080
timeout = "30s"

[database]
driver = "postgres"
host = "localhost"
port = 5432
```

### INI Example

```ini
[app]
name = MyApp
version = 1.0.0

[server]
host = 0.0.0.0
port = 8080

[database]
driver = postgres
host = localhost
port = 5432
```

## Complete Examples

### Basic Configuration

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
)

func main() {
    config := pkg.NewConfigManager()
    
    // Load configuration
    err := config.Load("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get values
    appName := config.GetString("app.name")
    port := config.GetInt("server.port")
    debug := config.GetBool("debug")
    
    log.Printf("Starting %s on port %d (debug: %v)", appName, port, debug)
}
```

### Environment-Based Configuration

```go
func main() {
    config := pkg.NewConfigManager()
    
    // Load base configuration
    config.Load("config.yaml")
    
    // Override with environment variables
    config.LoadFromEnv()
    
    // Use environment-specific settings
    if config.IsProduction() {
        // Production settings
        log.SetLevel(log.InfoLevel)
    } else {
        // Development settings
        log.SetLevel(log.DebugLevel)
    }
}
```

### Configuration with Defaults

```go
func main() {
    config := pkg.NewConfigManager()
    config.Load("config.yaml")
    
    // Get with defaults
    host := config.GetStringWithDefault("database.host", "localhost")
    port := config.GetIntWithDefault("database.port", 5432)
    maxConns := config.GetIntWithDefault("database.max_connections", 10)
    timeout := config.GetDuration("database.timeout")
    
    dbConfig := DatabaseConfig{
        Host:           host,
        Port:           port,
        MaxConnections: maxConns,
        Timeout:        timeout,
    }
}
```

### Hot-Reloading Configuration

```go
func main() {
    config := pkg.NewConfigManager()
    config.Load("config.yaml")
    
    // Watch for changes
    config.Watch(func() {
        log.Println("Configuration changed, reloading...")
        
        // Update application settings
        newPort := config.GetInt("server.port")
        newDebug := config.GetBool("debug")
        
        // Apply new settings
        updateSettings(newPort, newDebug)
    })
    
    // Application continues running
    // Configuration will auto-reload on file changes
}
```

### Sub-Configuration

```go
func main() {
    config := pkg.NewConfigManager()
    config.Load("config.yaml")
    
    // Get database sub-configuration
    dbConfig := config.Sub("database")
    
    db := ConnectDatabase(
        dbConfig.GetString("host"),
        dbConfig.GetInt("port"),
        dbConfig.GetString("username"),
        dbConfig.GetString("password"),
    )
    
    // Get cache sub-configuration
    cacheConfig := config.Sub("cache")
    
    cache := ConnectCache(
        cacheConfig.GetString("host"),
        cacheConfig.GetInt("port"),
    )
}
```

## Best Practices

1. **Use Environment Variables:** Override config with environment variables for deployment
2. **Set Defaults:** Always provide sensible defaults
3. **Validate Early:** Validate configuration at startup
4. **Use Sub-Configs:** Use Sub() for organizing related settings
5. **Environment Detection:** Use IsProduction(), IsDevelopment() for environment-specific logic
6. **Hot-Reload Carefully:** Test hot-reload thoroughly before using in production
7. **Secure Secrets:** Don't commit secrets to version control
8. **Document Config:** Document all configuration options
9. **Type Safety:** Use typed getters (GetInt, GetBool) instead of Get()
10. **Check IsSet:** Check IsSet() before using optional configuration

## Security Considerations

1. **Secrets Management:** Use environment variables or secret managers for sensitive data
2. **File Permissions:** Restrict configuration file permissions
3. **Validation:** Validate all configuration values
4. **Default Security:** Use secure defaults (e.g., HTTPS enabled)
5. **Environment Separation:** Use different configs for different environments
6. **Audit Logging:** Log configuration changes
7. **Encryption:** Encrypt sensitive configuration values

## See Also

- [Framework API](framework.md) - Framework configuration
- [Database API](database.md) - Database configuration
- [Security API](security.md) - Security configuration
- [Configuration Guide](../guides/configuration.md) - Configuration patterns
