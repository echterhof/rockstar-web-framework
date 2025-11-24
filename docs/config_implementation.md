# Configuration Management Implementation

## Overview

The Rockstar Web Framework provides a comprehensive configuration management system that supports multiple file formats (INI, TOML, YAML) and environment variables. The configuration system is thread-safe, supports nested configuration, type conversion, and file watching.

## Features

- **Multiple Format Support**: INI, TOML, and YAML configuration files
- **Environment Variables**: Load configuration from environment variables with `ROCKSTAR_` prefix
- **Type Conversion**: Automatic conversion to string, int, float, bool, duration, and slices
- **Nested Configuration**: Support for nested configuration sections
- **Default Values**: Fallback to default values when keys are missing
- **File Watching**: Automatic reload when configuration files change
- **Thread-Safe**: Concurrent read access with mutex protection
- **Environment Detection**: Built-in support for production, development, and test environments

## Requirements Satisfied

This implementation satisfies the following requirements:

- **9.1**: Support for INI file format
- **9.2**: Support for TOML file format
- **9.3**: Support for YAML file format
- **9.4**: Support for environment variables
- **9.5**: Configuration access through context

## Usage

### Basic Configuration Loading

```go
// Create a new configuration manager
config := pkg.NewConfigManager()

// Load from INI file
err := config.Load("config.ini")
if err != nil {
    log.Fatal(err)
}

// Access values
appName := config.GetString("app_name")
port := config.GetInt("port")
debug := config.GetBool("debug")
timeout := config.GetDuration("timeout")
```

### Loading from Different Formats

#### INI Format

```ini
# config.ini
app_name = MyApp
port = 8080
debug = true

[database]
host = localhost
port = 5432
```

```go
config := pkg.NewConfigManager()
config.Load("config.ini")

// Access root level
appName := config.GetString("app_name")

// Access nested section
dbConfig := config.Sub("database")
host := dbConfig.GetString("host")
```

#### TOML Format

```toml
# config.toml
app_name = "MyApp"
port = 8080
features = ["auth", "api"]

[database]
host = "localhost"
port = 5432

[server.tls]
enabled = true
cert_file = "/path/to/cert.pem"
```

```go
config := pkg.NewConfigManager()
config.Load("config.toml")

// Access arrays
features := config.GetStringSlice("features")

// Access deeply nested
serverConfig := config.Sub("server")
tlsConfig := serverConfig.Sub("tls")
enabled := tlsConfig.GetBool("enabled")
```

#### YAML Format

```yaml
# config.yaml
app_name: MyApp
port: 8080
debug: true

database:
  host: localhost
  port: 5432
  name: mydb
```

```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Access nested values
dbConfig := config.Sub("database")
dbName := dbConfig.GetString("name")
```

### Environment Variables

Environment variables with the `ROCKSTAR_` prefix are automatically loaded:

```bash
export ROCKSTAR_APP_NAME=MyApp
export ROCKSTAR_PORT=8080
export ROCKSTAR_DATABASE_HOST=localhost
export ROCKSTAR_DATABASE_PORT=5432
```

```go
config := pkg.NewConfigManager()
config.LoadFromEnv()

// Access with dot notation
appName := config.GetString("app.name")
port := config.GetInt("port")
dbHost := config.GetString("database.host")
```

### Default Values

```go
// Get with default fallback
appName := config.GetStringWithDefault("app_name", "DefaultApp")
port := config.GetIntWithDefault("port", 8080)
debug := config.GetBoolWithDefault("debug", false)
```

### Type Conversions

The configuration manager supports automatic type conversion:

```go
// String
name := config.GetString("name")

// Integer
port := config.GetInt("port")
count := config.GetInt64("count")

// Float
ratio := config.GetFloat64("ratio")

// Boolean
enabled := config.GetBool("enabled")

// Duration
timeout := config.GetDuration("timeout") // e.g., "30s", "5m"

// String slice
features := config.GetStringSlice("features")
```

### Nested Configuration

```go
// Access nested configuration
dbConfig := config.Sub("database")
host := dbConfig.GetString("host")
port := dbConfig.GetInt("port")

// Check if key exists
if config.IsSet("cache") {
    cacheConfig := config.Sub("cache")
    // ...
}
```

### Configuration Watching

Monitor configuration files for changes:

```go
config := pkg.NewConfigManager()
config.Load("config.ini")

// Watch for changes
err := config.Watch(func() {
    log.Println("Configuration reloaded!")
    // Handle configuration changes
})

// Stop watching when done
defer config.StopWatching()
```

### Manual Reload

```go
// Reload configuration from file
err := config.Reload()
if err != nil {
    log.Printf("Failed to reload config: %v", err)
}
```

### Environment Detection

```go
config := pkg.NewConfigManager()

// Check environment
env := config.GetEnv() // Returns "production", "development", or "test"

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

### Context Integration

The configuration manager is accessible through the request context:

```go
func handler(ctx pkg.Context) error {
    // Access configuration through context
    config := ctx.Config()
    
    // Use configuration
    timeout := config.GetDuration("request.timeout")
    maxSize := config.GetInt("request.max_size")
    
    // ...
    return nil
}
```

## Configuration File Examples

### Complete INI Example

```ini
# Application Configuration
app_name = MyRockstarApp
port = 8080
debug = true
timeout = 30s

[database]
host = localhost
port = 5432
name = mydb
username = admin
password = secret
max_connections = 100

[cache]
enabled = true
ttl = 3600
max_size = 1000

[server]
read_timeout = 30s
write_timeout = 30s
max_body_size = 10485760

[logging]
level = info
format = json
output = stdout
```

### Complete TOML Example

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
username = "admin"
password = "secret"
max_connections = 100

[cache]
enabled = true
ttl = 3600
max_size = 1000

[server]
timeout = "30s"
max_body_size = 10485760

[server.tls]
enabled = true
cert_file = "/path/to/cert.pem"
key_file = "/path/to/key.pem"

[logging]
level = "info"
format = "json"
output = "stdout"
```

### Complete YAML Example

```yaml
# Application Configuration
app_name: MyRockstarApp
port: 8080
debug: true
timeout: 30s

database:
  host: localhost
  port: 5432
  name: mydb
  username: admin
  password: secret
  max_connections: 100

cache:
  enabled: yes
  ttl: 3600
  max_size: 1000

server:
  timeout: 30s
  max_body_size: 10485760
  tls:
    enabled: yes
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem

logging:
  level: info
  format: json
  output: stdout
```

## Best Practices

1. **Use Environment-Specific Files**: Create separate config files for different environments
   ```
   config.development.yaml
   config.production.yaml
   config.test.yaml
   ```

2. **Sensitive Data**: Use environment variables for sensitive data like passwords and API keys
   ```bash
   export ROCKSTAR_DATABASE_PASSWORD=secret
   export ROCKSTAR_API_KEY=xyz123
   ```

3. **Default Values**: Always provide sensible defaults
   ```go
   timeout := config.GetDurationWithDefault("timeout", 30*time.Second)
   ```

4. **Validation**: Validate configuration after loading
   ```go
   if err := config.Validate(); err != nil {
       log.Fatal("Invalid configuration:", err)
   }
   ```

5. **Type Safety**: Use typed getters instead of generic Get()
   ```go
   // Good
   port := config.GetInt("port")
   
   // Avoid
   port := config.Get("port").(int) // Can panic
   ```

## Thread Safety

The configuration manager is thread-safe for concurrent reads:

```go
// Safe to call from multiple goroutines
go func() {
    value := config.GetString("key")
}()

go func() {
    value := config.GetInt("port")
}()
```

## Performance Considerations

- Configuration values are cached in memory for fast access
- File watching uses a 5-second polling interval (configurable)
- Nested configuration creates sub-managers that share the same data
- Type conversions are performed on each access (consider caching frequently used values)

## Error Handling

```go
// Check if file exists before loading
if _, err := os.Stat(configPath); os.IsNotExist(err) {
    log.Fatal("Config file not found:", configPath)
}

// Handle load errors
if err := config.Load(configPath); err != nil {
    log.Fatal("Failed to load config:", err)
}

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}

// Check if keys exist
if !config.IsSet("required_key") {
    log.Fatal("Required configuration key missing")
}
```

## Testing

The configuration manager includes comprehensive unit tests covering:

- Loading from INI, TOML, and YAML formats
- Environment variable loading
- Type conversions
- Default values
- Nested configuration
- File watching and reloading
- Concurrent access
- Validation

Run tests with:
```bash
go test -v -run TestConfigManager ./pkg
```

## Implementation Details

### File Format Detection

The configuration manager automatically detects the file format based on the file extension:
- `.ini` → INI parser
- `.toml` → TOML parser
- `.yaml`, `.yml` → YAML parser

### Environment Variable Mapping

Environment variables are mapped to configuration keys:
- `ROCKSTAR_APP_NAME` → `app.name`
- `ROCKSTAR_DATABASE_HOST` → `database.host`
- Underscores are converted to dots for nested keys

### Value Parsing

Values are automatically parsed to appropriate types:
- `"true"`, `"false"` → boolean
- `"123"` → integer
- `"3.14"` → float
- `"30s"`, `"5m"` → duration
- Other values → string

## Future Enhancements

Potential future improvements:
- Support for JSON configuration files
- Remote configuration sources (etcd, Consul)
- Configuration encryption
- Schema validation
- Configuration merging from multiple sources
- Hot reload without restart
