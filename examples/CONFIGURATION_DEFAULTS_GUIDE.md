# Configuration Defaults Guide

This guide explains the default values for all configuration structures in the Rockstar Web Framework. Understanding these defaults allows you to write minimal configuration files that only specify what you need to change.

## Philosophy

The Rockstar Web Framework follows the principle of **sensible defaults**. Every configuration parameter has a carefully chosen default value that works well for most applications. You only need to specify values when you want to override the defaults.

## Quick Start

Here's the absolute minimal configuration needed to start the framework:

```go
config := pkg.FrameworkConfig{
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "app.db",
    },
    SessionConfig: pkg.SessionConfig{
        EncryptionKey: []byte("your-32-byte-encryption-key-here"),
    },
}

app, err := pkg.New(config)
```

Everything else uses defaults!

## Configuration Structures

### ServerConfig

Controls HTTP server behavior.

| Field | Default | Description |
|-------|---------|-------------|
| `ReadTimeout` | 30 seconds | Maximum duration for reading the entire request |
| `WriteTimeout` | 30 seconds | Maximum duration before timing out writes |
| `IdleTimeout` | 120 seconds | Maximum idle time for keep-alive connections |
| `MaxHeaderBytes` | 1048576 (1 MB) | Maximum bytes for request headers |
| `MaxConnections` | 10000 | Maximum number of concurrent connections |
| `MaxRequestSize` | 10485760 (10 MB) | Maximum size of request body |
| `ShutdownTimeout` | 30 seconds | Maximum duration for graceful shutdown |
| `ReadBufferSize` | 4096 | Size of read buffer in bytes |
| `WriteBufferSize` | 4096 | Size of write buffer in bytes |
| `EnableHTTP1` | false | Enable HTTP/1.1 protocol |
| `EnableHTTP2` | false | Enable HTTP/2 protocol |
| `EnableQUIC` | false | Enable QUIC protocol |

**Example - Override only what you need:**
```go
ServerConfig: pkg.ServerConfig{
    EnableHTTP1: true,
    EnableHTTP2: true,
    MaxConnections: 5000, // Override default
    // All other values use defaults
}
```

### DatabaseConfig

Configures database connections.

| Field | Default | Description |
|-------|---------|-------------|
| `Host` | "localhost" | Database server hostname |
| `Port` | Driver-specific | Port number (postgres=5432, mysql=3306, mssql=1433, sqlite=0) |
| `MaxOpenConns` | 25 | Maximum open connections to database |
| `MaxIdleConns` | 5 | Maximum idle connections in pool |
| `ConnMaxLifetime` | 5 minutes | Maximum time a connection may be reused |
| `Driver` | (required) | Database driver name |
| `Database` | (required) | Database name |
| `Username` | (required) | Database username |
| `Password` | (required) | Database password |

**Example - Minimal configuration:**
```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "postgres",
    Database: "myapp",
    Username: "user",
    Password: "pass",
    // Host defaults to "localhost"
    // Port defaults to 5432 for postgres
    // Connection pool settings use defaults
}
```

### CacheConfig

Configures caching behavior.

| Field | Default | Description |
|-------|---------|-------------|
| `Type` | "memory" | Cache backend type |
| `MaxSize` | 0 (unlimited) | Maximum cache size in bytes |
| `DefaultTTL` | 0 (no expiration) | Default time-to-live for entries |

**Note:** Negative values for `MaxSize` and `DefaultTTL` are normalized to 0.

**Example - All defaults work great:**
```go
CacheConfig: pkg.CacheConfig{
    // Type defaults to "memory"
    // MaxSize defaults to unlimited
    // DefaultTTL defaults to no expiration
}
```

### SessionConfig

Manages user sessions.

| Field | Default | Description |
|-------|---------|-------------|
| `CookieName` | "rockstar_session" | Name of session cookie |
| `CookiePath` | "/" | Path scope for cookie |
| `SessionLifetime` | 24 hours | Duration before session expires |
| `CleanupInterval` | 1 hour | Interval for cleaning expired sessions |
| `FilesystemPath` | "./sessions" | Directory for filesystem storage |
| `CookieHTTPOnly` | true | Prevent JavaScript access to cookie |
| `CookieSecure` | true | Only send cookie over HTTPS |
| `CookieSameSite` | "Lax" | SameSite attribute value |
| `EncryptionKey` | (required) | 32-byte AES-256 encryption key |

**Example - Only encryption key required:**
```go
SessionConfig: pkg.SessionConfig{
    EncryptionKey: []byte("your-32-byte-key-here"),
    CookieSecure: false, // Override for development
    // All other values use secure defaults
}
```

### MonitoringConfig

Configures monitoring and profiling.

| Field | Default | Description |
|-------|---------|-------------|
| `MetricsPort` | 9090 | Port for metrics HTTP server |
| `PprofPort` | 6060 | Port for pprof HTTP server |
| `SNMPPort` | 161 | Port for SNMP server |
| `SNMPCommunity` | "public" | SNMP community string |
| `OptimizationInterval` | 5 minutes | Interval between optimizations |
| `EnableMetrics` | false | Enable metrics endpoint |
| `EnablePprof` | false | Enable pprof profiling |
| `EnableSNMP` | false | Enable SNMP monitoring |
| `EnableOptimization` | false | Enable automatic optimization |

**Example - Enable features, use default ports:**
```go
MonitoringConfig: pkg.MonitoringConfig{
    EnableMetrics: true,
    EnablePprof: true,
    // MetricsPort defaults to 9090
    // PprofPort defaults to 6060
    // OptimizationInterval defaults to 5 minutes
}
```

### PluginConfig

Configures individual plugins.

| Field | Default | Description |
|-------|---------|-------------|
| `Enabled` | true | Whether plugin is enabled |
| `Priority` | 0 | Execution order (lower first) |
| `Config` | {} (empty map) | Plugin-specific configuration |
| `Permissions` | All false | Operations plugin can perform |
| `Path` | (required) | Path to plugin binary |

**Example - Minimal plugin configuration:**
```go
PluginConfig: pkg.PluginConfig{
    Path: "./plugins/my-plugin",
    // Enabled defaults to true
    // Priority defaults to 0
    // Config defaults to empty map
    // Permissions default to all false (secure)
}
```

### PluginsConfig

Configures the plugin system.

| Field | Default | Description |
|-------|---------|-------------|
| `Enabled` | true | Enable plugin system globally |
| `Directory` | "./plugins" | Base directory for plugins |
| `Plugins` | [] (empty slice) | List of plugin configurations |

**Example - Minimal plugins configuration:**
```yaml
plugins:
  # enabled defaults to true
  # directory defaults to "./plugins"
  plugins:
    - name: my-plugin
      path: ./plugins/my-plugin
      # All other fields use defaults
```

## Best Practices

### 1. Start Minimal

Begin with the smallest configuration possible:

```go
config := pkg.FrameworkConfig{
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "app.db",
    },
    SessionConfig: pkg.SessionConfig{
        EncryptionKey: []byte("your-32-byte-key"),
    },
}
```

### 2. Override Only What You Need

Don't specify values that match the defaults:

```go
// ❌ Bad - specifying defaults unnecessarily
ServerConfig: pkg.ServerConfig{
    ReadTimeout:  30 * time.Second,  // This is the default!
    WriteTimeout: 30 * time.Second,  // This is the default!
    IdleTimeout:  120 * time.Second, // This is the default!
}

// ✅ Good - only override what you need
ServerConfig: pkg.ServerConfig{
    MaxConnections: 5000, // Override default of 10000
    // All other values use defaults
}
```

### 3. Document Your Overrides

When you override a default, add a comment explaining why:

```go
ServerConfig: pkg.ServerConfig{
    MaxConnections: 5000, // Reduced for resource-constrained environment
    ReadTimeout: 60 * time.Second, // Longer timeout for slow clients
}
```

### 4. Use Defaults in Development

Defaults are production-ready, but you may want to override some for development:

```go
SessionConfig: pkg.SessionConfig{
    EncryptionKey: []byte("dev-key-not-for-production-use!"),
    CookieSecure: false, // Allow HTTP in development
}
```

## Examples

See these example files for complete demonstrations:

- `examples/minimal_config_example.go` - Minimal configuration examples
- `examples/getting_started.go` - Updated with default comments
- `examples/monitoring_example.go` - Monitoring with defaults
- `examples/plugin-config-minimal.yaml` - Minimal plugin configuration

## Default Values Reference

For a complete list of all default values with descriptions, see the struct field comments in:

- `pkg/server.go` - ServerConfig
- `pkg/database.go` - DatabaseConfig
- `pkg/cache.go` - CacheConfig
- `pkg/session.go` - SessionConfig
- `pkg/monitoring.go` - MonitoringConfig
- `pkg/plugin_loader.go` - PluginConfig
- `pkg/plugin_config.go` - PluginsConfig
- `pkg/plugin.go` - PluginPermissions

Each field has a comment in the format:
```go
// FieldName description.
// Default: value
```

## Summary

The Rockstar Web Framework is designed to work out of the box with minimal configuration. Every configuration parameter has a sensible default value chosen for production use. This means:

- ✅ Less boilerplate in your configuration files
- ✅ Fewer opportunities for misconfiguration
- ✅ Faster development and deployment
- ✅ Production-ready defaults
- ✅ Easy to understand what you're changing

Start with the minimal configuration and add overrides only when needed!
