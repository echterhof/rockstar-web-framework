# Rockstar Framework Plugin Examples

This directory contains example plugins demonstrating various features of the Rockstar Web Framework plugin system.

## Overview

The plugin system allows you to extend the framework with custom functionality through dynamically loadable modules. Each plugin can:

- Hook into framework lifecycle events (startup, shutdown, request processing)
- Register custom middleware
- Access framework services (router, database, cache, etc.)
- Communicate with other plugins through events
- Export services for other plugins to use

## Example Plugins

### 1. Minimal Plugin

**Location:** `minimal-plugin/`

A minimal example demonstrating the basic plugin interface. This is the simplest possible plugin that implements all required methods.

**Features:**
- Basic lifecycle methods (Initialize, Start, Stop, Cleanup)
- Simple configuration schema
- Startup and shutdown hooks
- Logging

**Use Case:** Starting point for creating new plugins

**Configuration Example:**
```yaml
plugins:
  - name: minimal-plugin
    enabled: true
    path: ./examples/plugins/minimal-plugin
    config:
      message: "Hello from minimal plugin!"
```

### 2. Auth Plugin

**Location:** `auth-plugin/`

Demonstrates authentication hooks and middleware for securing your application.

**Features:**
- Pre-request hook for authentication
- Token-based authentication
- Authentication middleware
- Event publishing and subscription
- Configurable excluded paths
- Token duration management

**Use Case:** Adding authentication to your application without modifying core code

**Configuration Example:**
```yaml
plugins:
  - name: auth-plugin
    enabled: true
    path: ./examples/plugins/auth-plugin
    permissions:
      router: true
    config:
      token_duration: "2h"
      require_auth: true
      excluded_paths:
        - "/health"
        - "/public"
        - "/api/docs"
```

**Events:**
- **Publishes:** `auth.started`, `auth.stopped`, `auth.authenticated`
- **Subscribes:** `auth.login`

### 3. Logging Plugin

**Location:** `logging-plugin/`

Demonstrates comprehensive request/response logging with metrics.

**Features:**
- Pre-request hook for logging incoming requests
- Post-request hook for logging responses
- Request duration tracking
- Configurable logging levels
- Metrics collection (request count, response count, duration)
- Exported service for accessing statistics

**Use Case:** Monitoring and debugging request/response flow

**Configuration Example:**
```yaml
plugins:
  - name: logging-plugin
    enabled: true
    path: ./examples/plugins/logging-plugin
    config:
      log_requests: true
      log_responses: true
      log_headers: false
      log_body: false
```

**Exported Services:**
- **LoggingService:** Provides access to request/response counts

### 4. Cache Plugin

**Location:** `cache-plugin/`

Demonstrates response caching with configurable duration and exclusions.

**Features:**
- Caching middleware for GET requests
- Pre-response hook for caching responses
- Configurable cache duration
- Path exclusion support
- Cache hit/miss tracking
- Cache invalidation event handling
- Exported service for cache statistics

**Use Case:** Improving performance by caching responses

**Configuration Example:**
```yaml
plugins:
  - name: cache-plugin
    enabled: true
    path: ./examples/plugins/cache-plugin
    permissions:
      cache: true
      router: true
    config:
      cache_duration: "10m"
      enabled: true
      cache_methods: ["GET"]
      excluded_paths:
        - "/admin"
        - "/api/auth"
```

**Events:**
- **Subscribes:** `cache.invalidate`

**Exported Services:**
- **CacheService:** Provides cache statistics (hits, misses, hit rate)

## Plugin Structure

Each plugin follows this structure:

```
plugin-name/
├── plugin.go       # Plugin implementation
└── plugin.yaml     # Plugin manifest
```

### Plugin Implementation (plugin.go)

Every plugin must implement the `pkg.Plugin` interface:

```go
type Plugin interface {
    // Metadata
    Name() string
    Version() string
    Description() string
    Author() string
    
    // Dependencies
    Dependencies() []PluginDependency
    
    // Lifecycle
    Initialize(ctx PluginContext) error
    Start() error
    Stop() error
    Cleanup() error
    
    // Configuration
    ConfigSchema() map[string]interface{}
    OnConfigChange(config map[string]interface{}) error
}
```

### Plugin Manifest (plugin.yaml)

The manifest describes the plugin's metadata, dependencies, permissions, and configuration:

```yaml
name: plugin-name
version: 1.0.0
description: Plugin description
author: Author Name

framework:
  version: ">=1.0.0"

dependencies: []

permissions:
  database: false
  cache: false
  router: false
  config: false
  filesystem: false
  network: false
  exec: false

config:
  setting_name:
    type: string
    default: "value"
    description: Setting description

hooks:
  - type: startup
    priority: 100

events:
  publishes: []
  subscribes: []

exports: []
```

## Using Plugins

### 1. Loading Plugins from Configuration

Add plugins to your framework configuration file. The plugin system supports YAML, JSON, and TOML formats.

**See comprehensive configuration examples:**
- [plugin-config.yaml](../plugin-config.yaml) - YAML format with detailed documentation
- [plugin-config.json](../plugin-config.json) - JSON format example
- [plugin-config.toml](../plugin-config.toml) - TOML format example
- [PLUGIN_CONFIG_REFERENCE.md](../PLUGIN_CONFIG_REFERENCE.md) - Complete configuration reference

**Quick Example (YAML):**

```yaml
# config.yaml
plugins:
  enabled: true
  directory: ./plugins
  
  plugins:
    - name: minimal-plugin
      enabled: true
      path: ./examples/plugins/minimal-plugin
      priority: 100
      config:
        message: "Hello!"
    
    - name: auth-plugin
      enabled: true
      path: ./examples/plugins/auth-plugin
      priority: 200
      permissions:
        router: true
      config:
        token_duration: "1h"
```

### 2. Loading Plugins Programmatically

```go
import "github.com/yourusername/rockstar/pkg"

// Create framework
framework := pkg.NewFramework(config)

// Load plugin
err := framework.PluginManager().LoadPlugin(
    "./examples/plugins/minimal-plugin",
    pkg.PluginConfig{
        Enabled: true,
        Config: map[string]interface{}{
            "message": "Hello!",
        },
    },
)
```

### 3. Hot Reloading Plugins

```go
// Reload a plugin without restarting the application
err := framework.PluginManager().ReloadPlugin("minimal-plugin")
```

## Hook Types

Plugins can register hooks at various lifecycle points:

- **startup**: Executed during framework initialization
- **shutdown**: Executed during graceful shutdown
- **pre_request**: Executed before routing each request
- **post_request**: Executed after handler execution
- **pre_response**: Executed before sending responses
- **post_response**: Executed after sending responses
- **error**: Executed when errors occur

## Permissions

Plugins must declare required permissions in their manifest:

- **database**: Access to database operations
- **cache**: Access to cache operations
- **router**: Access to router (register routes, middleware)
- **config**: Access to framework configuration
- **filesystem**: Access to filesystem operations
- **network**: Access to network operations
- **exec**: Access to execute external commands

## Inter-Plugin Communication

### Events

Plugins can publish and subscribe to events:

```go
// Publish an event
ctx.PublishEvent("user.created", map[string]interface{}{
    "user_id": 123,
    "username": "john",
})

// Subscribe to an event
ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    // Handle event
    return nil
})
```

### Service Export/Import

Plugins can export services for other plugins:

```go
// Export a service
type MyService struct {
    // ...
}

ctx.ExportService("MyService", &MyService{})

// Import a service from another plugin
service, err := ctx.ImportService("other-plugin", "MyService")
if err == nil {
    myService := service.(*MyService)
    // Use the service
}
```

## Best Practices

1. **Error Handling**: Always handle errors gracefully and log them
2. **Resource Cleanup**: Clean up resources in the Cleanup() method
3. **Configuration Validation**: Validate configuration in Initialize()
4. **Permissions**: Request only the permissions you need
5. **Dependencies**: Declare all plugin dependencies in the manifest
6. **Versioning**: Use semantic versioning for your plugins
7. **Documentation**: Document your plugin's configuration and exported services
8. **Testing**: Write tests for your plugin functionality
9. **Logging**: Use the provided logger for consistent logging
10. **Metrics**: Record metrics for monitoring plugin performance

## Building Plugins

### As Go Packages

Plugins can be built as Go packages and loaded directly:

```bash
cd examples/plugins/minimal-plugin
go build -o minimal-plugin.so -buildmode=plugin plugin.go
```

### As Separate Processes

For better isolation, plugins can run as separate processes communicating via gRPC (advanced usage).

## Troubleshooting

### Plugin Not Loading

1. Check that the plugin path is correct
2. Verify the manifest is valid YAML/JSON
3. Ensure all required fields are present in the manifest
4. Check that dependencies are loaded first
5. Verify permissions are granted in configuration

### Hook Not Executing

1. Verify the hook is registered in Initialize()
2. Check the hook priority (higher priority = earlier execution)
3. Ensure the plugin is started (Start() called)
4. Check logs for hook execution errors

### Permission Denied

1. Verify permissions are declared in the manifest
2. Ensure permissions are granted in the framework configuration
3. Check that you're accessing services through the PluginContext

## Further Reading

- [Framework Documentation](../../docs/PLUGIN_SYSTEM.md)

## Contributing

To contribute a new example plugin:

1. Create a new directory under `examples/plugins/`
2. Implement the plugin following the structure above
3. Create a manifest file (plugin.yaml)
4. Add documentation to this README
5. Test your plugin thoroughly
6. Submit a pull request

## License

These examples are part of the Rockstar Web Framework and are provided under the same license.
