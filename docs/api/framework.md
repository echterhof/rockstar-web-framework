---
title: "Framework API"
description: "Framework struct and core initialization methods"
category: "api"
tags: ["api", "framework", "initialization", "lifecycle"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "context.md"
  - "router.md"
  - "server.md"
  - "../guides/configuration.md"
---

# Framework API

## Overview

The `Framework` struct is the main entry point for the Rockstar Web Framework. It wires together all framework components (router, database, cache, security, etc.) and provides a unified interface for application initialization, configuration, and lifecycle management.

The Framework follows a manager-based architecture where each major feature area is exposed through a dedicated manager interface, eliminating global state and enabling dependency injection through the Context.

**Primary Use Cases:**
- Application initialization and configuration
- Component access and management
- Server lifecycle control (startup, shutdown)
- Global middleware registration
- Lifecycle hook registration

## Type Definitions

### Framework

```go
type Framework struct {
    // Core components
    serverManager ServerManager
    router        RouterEngine
    security      SecurityManager

    // Data layer
    database DatabaseManager
    cache    CacheManager
    session  SessionManager

    // Configuration and i18n
    config ConfigManager
    i18n   I18nManager

    // Monitoring and metrics
    metrics    MetricsCollector
    monitoring MonitoringManager

    // Proxy
    proxy ProxyManager

    // Plugin system
    pluginManager PluginManager

    // File and network
    fileManager   FileManager
    networkClient NetworkClient

    // Middleware
    globalMiddleware []MiddlewareFunc

    // Error handling
    errorHandler func(ctx Context, err error) error

    // Lifecycle hooks
    shutdownHooks []func(ctx context.Context) error
    startupHooks  []func(ctx context.Context) error

    // State
    isRunning bool
}
```

The Framework struct is the central orchestrator that manages all framework components. All fields are private to enforce proper initialization through the `New()` function.

### FrameworkConfig

```go
type FrameworkConfig struct {
    // Server configuration
    ServerConfig ServerConfig

    // Database configuration
    DatabaseConfig DatabaseConfig

    // Cache configuration
    CacheConfig CacheConfig

    // Session configuration
    SessionConfig SessionConfig

    // Configuration file paths
    ConfigFiles []string

    // i18n configuration
    I18nConfig I18nConfig

    // Security configuration
    SecurityConfig SecurityConfig

    // Monitoring configuration
    MonitoringConfig MonitoringConfig

    // Proxy configuration
    ProxyConfig ProxyConfig

    // Plugin configuration
    PluginConfigPath string
    EnablePlugins    bool

    // File system configuration
    FileSystemRoot string // Root directory for file operations (defaults to current directory)
}
```

FrameworkConfig holds the complete framework configuration. All nested configuration structs have sensible defaults applied automatically via `ApplyDefaults()` methods.

## Initialization

### New

```go
func New(config FrameworkConfig) (*Framework, error)
```

**Description**: Creates a new Framework instance with the given configuration. This is the primary initialization function that sets up all framework components.

**Parameters**:
- `config` (FrameworkConfig): Complete framework configuration including server, database, cache, session, and other component settings

**Returns**:
- `*Framework`: Initialized framework instance ready for route registration and server startup
- `error`: Error if initialization fails (e.g., database connection failure, invalid configuration)

**Behavior**:
1. Applies defaults to all configuration structures
2. Initializes configuration manager and loads config files
3. Initializes i18n manager with locale files
4. Conditionally initializes database (or uses no-op implementation)
5. Initializes cache, session, security, metrics, and monitoring managers
6. Initializes proxy, router, and server managers
7. Initializes file and network managers
8. Optionally discovers and initializes plugins if enabled

**Example**:
```go
package main

import (
    "log"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Create configuration with minimal settings
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1: true,
            EnableHTTP2: true,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
        SessionConfig: pkg.SessionConfig{
            EncryptionKey: []byte("your-32-byte-encryption-key-here"),
            CookieSecure:  false, // Set true in production
        },
    }

    // Initialize framework
    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create framework: %v", err)
    }

    // Framework is ready for route registration
    router := app.Router()
    router.GET("/", homeHandler)

    // Start server
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

**See Also**:
- [Configuration Guide](../guides/configuration.md)
- [Getting Started](../GETTING_STARTED.md)

## Component Access Methods

The Framework provides getter methods for accessing all managed components. These methods return the respective manager interfaces.

### Router

```go
func (f *Framework) Router() RouterEngine
```

**Description**: Returns the framework's router for route registration.

**Returns**:
- `RouterEngine`: Router interface for registering routes and route groups

**Example**:
```go
router := app.Router()
router.GET("/api/users", listUsersHandler)
router.POST("/api/users", createUserHandler)
router.GET("/api/users/:id", getUserHandler)
```

**See Also**:
- [Router API](router.md)
- [Routing Guide](../guides/routing.md)

### Database

```go
func (f *Framework) Database() DatabaseManager
```

**Description**: Returns the framework's database manager for direct database access outside of request handlers.

**Returns**:
- `DatabaseManager`: Database interface for executing queries and managing connections

**Example**:
```go
db := app.Database()
if db.IsConnected() {
    // Perform database operations
    result, err := db.Query("SELECT * FROM users")
    if err != nil {
        log.Printf("Query failed: %v", err)
    }
}
```

**See Also**:
- [Database API](database.md)
- [Database Guide](../guides/database.md)

### Cache

```go
func (f *Framework) Cache() CacheManager
```

**Description**: Returns the framework's cache manager for caching operations.

**Returns**:
- `CacheManager`: Cache interface for storing and retrieving cached data

**Example**:
```go
cache := app.Cache()
cache.Set("user:123", userData, 5*time.Minute)
```

**See Also**:
- [Cache API](cache.md)
- [Caching Guide](../guides/caching.md)

### Session

```go
func (f *Framework) Session() SessionManager
```

**Description**: Returns the framework's session manager for session operations.

**Returns**:
- `SessionManager`: Session interface for managing user sessions

**Example**:
```go
session := app.Session()
session.CleanupExpired() // Manually trigger cleanup
```

**See Also**:
- [Session API](session.md)
- [Sessions Guide](../guides/sessions.md)

### Security

```go
func (f *Framework) Security() SecurityManager
```

**Description**: Returns the framework's security manager for authentication and authorization.

**Returns**:
- `SecurityManager`: Security interface for managing authentication, authorization, and tokens

**Example**:
```go
security := app.Security()
token, err := security.GenerateToken(userID, 24*time.Hour)
```

**See Also**:
- [Security API](security.md)
- [Security Guide](../guides/security.md)

### Config

```go
func (f *Framework) Config() ConfigManager
```

**Description**: Returns the framework's configuration manager for accessing configuration values.

**Returns**:
- `ConfigManager`: Configuration interface for reading and managing configuration

**Example**:
```go
config := app.Config()
apiKey := config.GetString("api.key")
maxRetries := config.GetInt("api.max_retries")
```

**See Also**:
- [Configuration Guide](../guides/configuration.md)

### I18n

```go
func (f *Framework) I18n() I18nManager
```

**Description**: Returns the framework's internationalization manager for translations.

**Returns**:
- `I18nManager`: I18n interface for managing translations and locales

**Example**:
```go
i18n := app.I18n()
message := i18n.Translate("en", "welcome.message")
```

**See Also**:
- [I18n API](i18n.md)
- [Internationalization Guide](../guides/i18n.md)

### Metrics

```go
func (f *Framework) Metrics() MetricsCollector
```

**Description**: Returns the framework's metrics collector for recording application metrics.

**Returns**:
- `MetricsCollector`: Metrics interface for collecting and exporting metrics

**Example**:
```go
metrics := app.Metrics()
metrics.IncrementCounter("api.requests")
metrics.RecordHistogram("api.latency", duration)
```

**See Also**:
- [Metrics API](metrics.md)
- [Monitoring Guide](../guides/monitoring.md)

### Monitoring

```go
func (f *Framework) Monitoring() MonitoringManager
```

**Description**: Returns the framework's monitoring manager for health checks and monitoring.

**Returns**:
- `MonitoringManager`: Monitoring interface for health checks and system monitoring

**Example**:
```go
monitoring := app.Monitoring()
monitoring.Start() // Start monitoring
```

**See Also**:
- [Monitoring API](monitoring.md)
- [Monitoring Guide](../guides/monitoring.md)

### Proxy

```go
func (f *Framework) Proxy() ProxyManager
```

**Description**: Returns the framework's proxy manager for reverse proxy operations.

**Returns**:
- `ProxyManager`: Proxy interface for configuring and managing reverse proxies

**Example**:
```go
proxy := app.Proxy()
// Configure proxy rules
```

**See Also**:
- [Proxy API](proxy.md)

### PluginManager

```go
func (f *Framework) PluginManager() PluginManager
```

**Description**: Returns the framework's plugin manager for plugin lifecycle management.

**Returns**:
- `PluginManager`: Plugin interface for managing plugins (nil if plugins not enabled)

**Example**:
```go
if pluginMgr := app.PluginManager(); pluginMgr != nil {
    plugins := pluginMgr.ListPlugins()
    for _, plugin := range plugins {
        log.Printf("Plugin: %s v%s", plugin.Name, plugin.Version)
    }
}
```

**See Also**:
- [Plugins API](plugins.md)
- [Plugin Development Guide](../guides/plugins.md)

### ServerManager

```go
func (f *Framework) ServerManager() ServerManager
```

**Description**: Returns the framework's server manager for advanced server operations.

**Returns**:
- `ServerManager`: Server interface for managing multiple servers

**Example**:
```go
serverMgr := app.ServerManager()
// Advanced server management
```

**See Also**:
- [Server API](server.md)

### FileManager

```go
func (f *Framework) FileManager() FileManager
```

**Description**: Returns the framework's file manager for file system operations.

**Returns**:
- `FileManager`: File interface for file operations

**Example**:
```go
files := app.FileManager()
content, err := files.Read("config.json")
```

### NetworkClient

```go
func (f *Framework) NetworkClient() NetworkClient
```

**Description**: Returns the framework's network client for making HTTP requests.

**Returns**:
- `NetworkClient`: Network interface for HTTP client operations

**Example**:
```go
client := app.NetworkClient()
// Make HTTP requests
```

## Middleware Management

### Use

```go
func (f *Framework) Use(middleware ...MiddlewareFunc)
```

**Description**: Adds global middleware to the framework. Global middleware is executed for every request before route-specific handlers.

**Parameters**:
- `middleware` (...MiddlewareFunc): One or more middleware functions to add

**Execution Order**: Middleware is executed in the order it is registered (first registered, first executed).

**Example**:
```go
// Add logging middleware
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    log.Printf("[%s] %s", ctx.Request().Method, ctx.Request().RequestURI)
    err := next(ctx)
    log.Printf("Completed in %v", time.Since(start))
    return err
})

// Add authentication middleware
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    return next(ctx)
})

// Add recovery middleware
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic recovered: %v", r)
            ctx.JSON(500, map[string]string{"error": "Internal server error"})
        }
    }()
    return next(ctx)
})
```

**See Also**:
- [Middleware Guide](../guides/middleware.md)

## Error Handling

### SetErrorHandler

```go
func (f *Framework) SetErrorHandler(handler func(ctx Context, err error) error)
```

**Description**: Sets a custom error handler for all errors that occur during request processing. The error handler receives the context and error, and can return a new error or nil.

**Parameters**:
- `handler` (func(ctx Context, err error) error): Custom error handler function

**Example**:
```go
app.SetErrorHandler(func(ctx pkg.Context, err error) error {
    // Log the error
    log.Printf("Error: %v", err)
    
    // Return custom error response
    return ctx.JSON(500, map[string]interface{}{
        "error": "Internal server error",
        "message": err.Error(),
        "timestamp": time.Now().Unix(),
    })
})
```

**See Also**:
- [Error Handling Guide](../guides/middleware.md#error-handling)

## Lifecycle Management

### RegisterStartupHook

```go
func (f *Framework) RegisterStartupHook(hook func(ctx context.Context) error)
```

**Description**: Registers a function to be called during server startup, before the server begins accepting requests. Startup hooks are executed in registration order.

**Parameters**:
- `hook` (func(ctx context.Context) error): Function to execute on startup

**Returns**: If any startup hook returns an error, server startup is aborted.

**Example**:
```go
app.RegisterStartupHook(func(ctx context.Context) error {
    log.Println("Initializing database schema...")
    return app.Database().Exec("CREATE TABLE IF NOT EXISTS users (...)")
})

app.RegisterStartupHook(func(ctx context.Context) error {
    log.Println("Loading configuration...")
    return app.Config().Load("config.yaml")
})

app.RegisterStartupHook(func(ctx context.Context) error {
    log.Println("Server ready!")
    return nil
})
```

**See Also**:
- [Lifecycle Hooks](../guides/deployment.md#lifecycle-hooks)

### RegisterShutdownHook

```go
func (f *Framework) RegisterShutdownHook(hook func(ctx context.Context) error)
```

**Description**: Registers a function to be called during graceful shutdown. Shutdown hooks are executed in registration order after the server stops accepting new requests.

**Parameters**:
- `hook` (func(ctx context.Context) error): Function to execute on shutdown

**Example**:
```go
app.RegisterShutdownHook(func(ctx context.Context) error {
    log.Println("Flushing metrics...")
    return app.Metrics().Flush()
})

app.RegisterShutdownHook(func(ctx context.Context) error {
    log.Println("Closing external connections...")
    // Close external service connections
    return nil
})

app.RegisterShutdownHook(func(ctx context.Context) error {
    log.Println("Cleanup complete!")
    return nil
})
```

**See Also**:
- [Graceful Shutdown](../guides/deployment.md#graceful-shutdown)

## Server Lifecycle

### Listen

```go
func (f *Framework) Listen(addr string) error
```

**Description**: Starts the framework server on the specified address with default configuration. This is a blocking call that runs until the server is shut down.

**Parameters**:
- `addr` (string): Address to listen on (e.g., ":8080", "localhost:3000", "0.0.0.0:8080")

**Returns**:
- `error`: Error if server fails to start or encounters a fatal error

**Behavior**:
1. Creates server with default configuration
2. Executes all registered startup hooks
3. Starts accepting HTTP/1.1 and HTTP/2 requests
4. Blocks until shutdown

**Example**:
```go
// Listen on port 8080
if err := app.Listen(":8080"); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

**See Also**:
- [Server Configuration](../guides/configuration.md#server-configuration)

### ListenTLS

```go
func (f *Framework) ListenTLS(addr, certFile, keyFile string) error
```

**Description**: Starts the framework server with TLS/HTTPS on the specified address.

**Parameters**:
- `addr` (string): Address to listen on
- `certFile` (string): Path to TLS certificate file
- `keyFile` (string): Path to TLS private key file

**Returns**:
- `error`: Error if server fails to start or TLS configuration is invalid

**Example**:
```go
// Listen with TLS on port 443
if err := app.ListenTLS(":443", "cert.pem", "key.pem"); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

**See Also**:
- [TLS Configuration](../guides/security.md#tls-configuration)
- [Deployment Guide](../guides/deployment.md)

### ListenQUIC

```go
func (f *Framework) ListenQUIC(addr, certFile, keyFile string) error
```

**Description**: Starts the framework server with QUIC protocol on the specified address. QUIC provides improved performance over traditional TCP-based protocols.

**Parameters**:
- `addr` (string): Address to listen on
- `certFile` (string): Path to TLS certificate file (required for QUIC)
- `keyFile` (string): Path to TLS private key file (required for QUIC)

**Returns**:
- `error`: Error if server fails to start or QUIC configuration is invalid

**Example**:
```go
// Listen with QUIC on port 443
if err := app.ListenQUIC(":443", "cert.pem", "key.pem"); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

**See Also**:
- [QUIC Protocol Guide](../guides/protocols.md#quic)

### ListenWithConfig

```go
func (f *Framework) ListenWithConfig(addr string, config ServerConfig) error
```

**Description**: Starts the framework server with custom server configuration. This provides full control over server behavior including timeouts, buffer sizes, and protocol support.

**Parameters**:
- `addr` (string): Address to listen on
- `config` (ServerConfig): Custom server configuration

**Returns**:
- `error`: Error if server fails to start

**Example**:
```go
config := pkg.ServerConfig{
    EnableHTTP1:      true,
    EnableHTTP2:      true,
    ReadTimeout:      30 * time.Second,
    WriteTimeout:     30 * time.Second,
    IdleTimeout:      120 * time.Second,
    MaxHeaderBytes:   1 << 20, // 1 MB
    MaxConnections:   10000,
    MaxRequestSize:   10 << 20, // 10 MB
    ShutdownTimeout:  30 * time.Second,
}

if err := app.ListenWithConfig(":8080", config); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

**See Also**:
- [Server Configuration](../guides/configuration.md#server-configuration)
- [Performance Tuning](../guides/performance.md)

### Shutdown

```go
func (f *Framework) Shutdown(timeout time.Duration) error
```

**Description**: Gracefully shuts down the framework. This stops accepting new requests, waits for active requests to complete (up to timeout), executes shutdown hooks, and cleans up resources.

**Parameters**:
- `timeout` (time.Duration): Maximum time to wait for graceful shutdown

**Returns**:
- `error`: Error if shutdown fails or times out

**Behavior**:
1. Executes all registered shutdown hooks (including plugin hooks)
2. Stops accepting new requests
3. Waits for active requests to complete (up to timeout)
4. Closes database connections
5. Clears cache
6. Cleans up expired sessions

**Example**:
```go
// In a signal handler
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down gracefully...")
    if err := app.Shutdown(30 * time.Second); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
    os.Exit(0)
}()

// Start server
app.Listen(":8080")
```

**See Also**:
- [Graceful Shutdown](../guides/deployment.md#graceful-shutdown)

### IsRunning

```go
func (f *Framework) IsRunning() bool
```

**Description**: Returns whether the framework server is currently running.

**Returns**:
- `bool`: true if server is running, false otherwise

**Example**:
```go
if app.IsRunning() {
    log.Println("Server is running")
} else {
    log.Println("Server is stopped")
}
```

## Complete Example

Here's a complete example demonstrating Framework initialization, configuration, middleware, routes, and lifecycle management:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Configuration
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1: true,
            EnableHTTP2: true,
            ReadTimeout: 30 * time.Second,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
        SessionConfig: pkg.SessionConfig{
            EncryptionKey: []byte("your-32-byte-encryption-key-here"),
        },
        EnablePlugins: true,
    }

    // Initialize framework
    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create framework: %v", err)
    }

    // Register lifecycle hooks
    app.RegisterStartupHook(func(ctx context.Context) error {
        log.Println("Server starting...")
        return nil
    })

    app.RegisterShutdownHook(func(ctx context.Context) error {
        log.Println("Server shutting down...")
        return nil
    })

    // Add global middleware
    app.Use(loggingMiddleware)
    app.Use(recoveryMiddleware)

    // Set error handler
    app.SetErrorHandler(errorHandler)

    // Register routes
    router := app.Router()
    router.GET("/", homeHandler)
    router.GET("/api/users", listUsersHandler)
    router.POST("/api/users", createUserHandler)

    // Setup graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        if err := app.Shutdown(30 * time.Second); err != nil {
            log.Printf("Shutdown error: %v", err)
        }
        os.Exit(0)
    }()

    // Start server
    log.Println("Server listening on :8080")
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    log.Printf("[%s] %s", ctx.Request().Method, ctx.Request().RequestURI)
    err := next(ctx)
    log.Printf("Completed in %v", time.Since(start))
    return err
}

func recoveryMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic: %v", r)
            ctx.JSON(500, map[string]string{"error": "Internal server error"})
        }
    }()
    return next(ctx)
}

func errorHandler(ctx pkg.Context, err error) error {
    log.Printf("Error: %v", err)
    return ctx.JSON(500, map[string]interface{}{
        "error": err.Error(),
    })
}

func homeHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{
        "message": "Welcome to Rockstar!",
    })
}

func listUsersHandler(ctx pkg.Context) error {
    // Use framework components through context
    db := ctx.DB()
    cache := ctx.Cache()
    
    // Implementation...
    return ctx.JSON(200, []string{"user1", "user2"})
}

func createUserHandler(ctx pkg.Context) error {
    // Implementation...
    return ctx.JSON(201, map[string]string{
        "message": "User created",
    })
}
```

## Best Practices

### Configuration Management

1. **Use Environment Variables**: Load sensitive configuration from environment variables, not hardcoded values
2. **Apply Defaults**: Let the framework apply sensible defaults - only override what you need
3. **Validate Early**: Validate configuration during initialization, not at runtime
4. **Separate Environments**: Use different configurations for development, staging, and production

### Lifecycle Management

1. **Register Hooks Early**: Register startup and shutdown hooks before calling `Listen()`
2. **Handle Errors**: Always check errors from startup hooks - they prevent server startup
3. **Graceful Shutdown**: Always implement graceful shutdown with appropriate timeout
4. **Cleanup Resources**: Use shutdown hooks to clean up external resources (connections, files, etc.)

### Middleware Organization

1. **Order Matters**: Register middleware in the correct order (logging → recovery → auth → business logic)
2. **Keep It Simple**: Middleware should be focused and single-purpose
3. **Error Handling**: Always handle errors in middleware appropriately
4. **Performance**: Be mindful of middleware performance - it runs on every request

### Component Access

1. **Use Context**: Access framework components through the Context in handlers, not directly from Framework
2. **Check Availability**: Check if optional components (like plugins) are enabled before using them
3. **Avoid Globals**: Never store Framework instance in global variables - pass it through your application

## Navigation

- [← Back to API Reference](README.md)
- [Next: Context API →](context.md)
- [Router API →](router.md)
- [Configuration Guide →](../guides/configuration.md)
- [Getting Started →](../GETTING_STARTED.md)
