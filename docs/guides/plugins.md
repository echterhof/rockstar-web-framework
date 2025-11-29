# Plugin System

## Overview

The Rockstar Web Framework features a powerful plugin system that allows you to extend the framework's functionality without modifying its core code. Plugins can register hooks, middleware, subscribe to events, export services, and access framework resources with fine-grained permission control.

**When to use plugins:**
- Adding custom functionality to the framework
- Integrating third-party services
- Implementing reusable components across projects
- Extending framework capabilities without forking
- Creating modular, maintainable applications

**Key benefits:**
- **Compile-Time Integration**: Plugins are compiled into your application for maximum performance
- **Lifecycle Management**: Automatic initialization, startup, and shutdown
- **Dependency Resolution**: Automatic dependency ordering and validation
- **Permission System**: Fine-grained control over framework resource access
- **Hook System**: Intercept and modify framework behavior at key points
- **Event Bus**: Inter-plugin communication through events
- **Service Registry**: Share functionality between plugins
- **Hot Configuration**: Update plugin configuration without restart

## Quick Start

Here's a minimal plugin example:

```go
package myplugin

import (
    "fmt"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

// Register plugin at compile time
func init() {
    pkg.RegisterPlugin("my-plugin", func() pkg.Plugin {
        return &MyPlugin{}
    })
}

// MyPlugin implements the Plugin interface
type MyPlugin struct {
    ctx    pkg.PluginContext
    config map[string]interface{}
}

// Metadata methods
func (p *MyPlugin) Name() string        { return "my-plugin" }
func (p *MyPlugin) Version() string     { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }
func (p *MyPlugin) Author() string      { return "Your Name <email@example.com>" }

// Dependencies
func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{}
}

// Lifecycle methods
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    p.config = ctx.PluginConfig()
    
    // Register a hook
    return ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
        fmt.Println("Pre-request hook executed")
        return nil
    })
}

func (p *MyPlugin) Start() error {
    fmt.Println("Plugin started")
    return nil
}

func (p *MyPlugin) Stop() error {
    fmt.Println("Plugin stopped")
    return nil
}

func (p *MyPlugin) Cleanup() error {
    p.ctx = nil
    return nil
}

// Configuration
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{}
}

func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    p.config = config
    return nil
}
```

## Plugin Architecture

### Plugin Interface

All plugins must implement the `Plugin` interface:

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

### Plugin Lifecycle

Plugins go through several lifecycle stages:

1. **Registration**: Plugin is registered via `init()` function
2. **Discovery**: Framework discovers registered plugins
3. **Dependency Resolution**: Dependencies are validated and ordered
4. **Initialization**: `Initialize()` is called with plugin context
5. **Startup**: `Start()` is called after all plugins initialized
6. **Running**: Plugin is active and processing events/hooks
7. **Shutdown**: `Stop()` is called during graceful shutdown
8. **Cleanup**: `Cleanup()` is called to release resources

```
Registration → Discovery → Dependency Resolution → Initialize → Start → Running → Stop → Cleanup
```

### Plugin Context

The `PluginContext` provides isolated access to framework services:

```go
type PluginContext interface {
    // Core framework access
    Router() RouterEngine
    Logger() Logger
    Metrics() MetricsCollector
    
    // Data access (permission-controlled)
    Database() DatabaseManager
    Cache() CacheManager
    Config() ConfigManager
    FileSystem() FileManager
    Network() NetworkClient
    
    // Plugin-specific
    PluginConfig() map[string]interface{}
    PluginStorage() PluginStorage
    
    // Hook registration
    RegisterHook(hookType HookType, priority int, handler HookHandler) error
    
    // Event system
    PublishEvent(event string, data interface{}) error
    SubscribeEvent(event string, handler EventHandler) error
    
    // Service export/import
    ExportService(name string, service interface{}) error
    ImportService(pluginName, serviceName string) (interface{}, error)
    
    // Middleware registration
    RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error
    UnregisterMiddleware(name string) error
}
```

## Creating a Plugin

### Step 1: Create Plugin Directory

```bash
mkdir -p plugins/my-plugin
cd plugins/my-plugin
```

### Step 2: Create Plugin Manifest

Create `plugin.yaml`:

```yaml
name: my-plugin
version: 1.0.0
description: My custom plugin
author: Your Name <your.email@example.com>

framework:
  version: ">=1.0.0"

dependencies: []

permissions:
  database: false
  cache: true
  router: true
  config: false
  filesystem: false
  network: false
  exec: false

config:
  api_key:
    type: string
    required: true
    description: API key for external service
  timeout:
    type: duration
    default: "30s"
    description: Request timeout
  enabled:
    type: bool
    default: true
    description: Enable plugin functionality

hooks:
  - type: pre_request
    priority: 100
  - type: post_request
    priority: 100

events:
  publishes:
    - my-plugin.started
    - my-plugin.action_completed
  subscribes:
    - user.created
    - cache.invalidate

exports:
  - name: MyService
    description: Service for doing X
```

### Step 3: Implement Plugin

Create `plugin.go`:

```go
package myplugin

import (
    "fmt"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func init() {
    pkg.RegisterPlugin("my-plugin", func() pkg.Plugin {
        return &MyPlugin{}
    })
}

type MyPlugin struct {
    ctx     pkg.PluginContext
    config  map[string]interface{}
    apiKey  string
    timeout time.Duration
    enabled bool
}

func (p *MyPlugin) Name() string        { return "my-plugin" }
func (p *MyPlugin) Version() string     { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }
func (p *MyPlugin) Author() string      { return "Your Name <your.email@example.com>" }

func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{}
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    p.config = ctx.PluginConfig()
    
    // Parse configuration
    if apiKey, ok := p.config["api_key"].(string); ok {
        p.apiKey = apiKey
    } else {
        return fmt.Errorf("api_key is required")
    }
    
    if timeout, ok := p.config["timeout"].(time.Duration); ok {
        p.timeout = timeout
    } else {
        p.timeout = 30 * time.Second
    }
    
    if enabled, ok := p.config["enabled"].(bool); ok {
        p.enabled = enabled
    } else {
        p.enabled = true
    }
    
    // Register hooks
    err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, p.preRequestHook)
    if err != nil {
        return fmt.Errorf("failed to register pre-request hook: %w", err)
    }
    
    err = ctx.RegisterHook(pkg.HookTypePostRequest, 100, p.postRequestHook)
    if err != nil {
        return fmt.Errorf("failed to register post-request hook: %w", err)
    }
    
    // Subscribe to events
    err = ctx.SubscribeEvent("user.created", p.onUserCreated)
    if err != nil {
        return fmt.Errorf("failed to subscribe to user.created: %w", err)
    }
    
    // Export service
    service := &MyService{plugin: p}
    err = ctx.ExportService("MyService", service)
    if err != nil {
        return fmt.Errorf("failed to export service: %w", err)
    }
    
    return nil
}

func (p *MyPlugin) Start() error {
    if !p.enabled {
        return nil
    }
    
    // Publish startup event
    return p.ctx.PublishEvent("my-plugin.started", map[string]interface{}{
        "timestamp": time.Now(),
    })
}

func (p *MyPlugin) Stop() error {
    return nil
}

func (p *MyPlugin) Cleanup() error {
    p.ctx = nil
    return nil
}

func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "api_key": map[string]interface{}{
            "type":        "string",
            "required":    true,
            "description": "API key for external service",
        },
        "timeout": map[string]interface{}{
            "type":        "duration",
            "default":     "30s",
            "description": "Request timeout",
        },
        "enabled": map[string]interface{}{
            "type":        "bool",
            "default":     true,
            "description": "Enable plugin functionality",
        },
    }
}

func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    p.config = config
    // Re-parse configuration
    if apiKey, ok := config["api_key"].(string); ok {
        p.apiKey = apiKey
    }
    return nil
}

// Hook handlers
func (p *MyPlugin) preRequestHook(hookCtx pkg.HookContext) error {
    ctx := hookCtx.Context()
    if ctx != nil {
        // Add custom header
        ctx.Response().SetHeader("X-My-Plugin", "active")
    }
    return nil
}

func (p *MyPlugin) postRequestHook(hookCtx pkg.HookContext) error {
    // Log request completion
    return nil
}

// Event handlers
func (p *MyPlugin) onUserCreated(event pkg.Event) error {
    fmt.Printf("User created: %v\n", event.Data)
    return nil
}

// Exported service
type MyService struct {
    plugin *MyPlugin
}

func (s *MyService) DoSomething() string {
    return "Service method called"
}
```

### Step 4: Create Go Module

Create `go.mod`:

```go
module github.com/yourusername/my-plugin

go 1.21

require github.com/echterhof/rockstar-web-framework v1.0.0
```

### Step 5: Import Plugin

In your main application, import the plugin:

```go
package main

import (
    "log"
    "github.com/echterhof/rockstar-web-framework/pkg"
    
    // Import your plugin
    _ "github.com/yourusername/my-plugin"
)

func main() {
    config := pkg.FrameworkConfig{
        // ... framework configuration
        
        PluginConfigs: map[string]pkg.PluginConfig{
            "my-plugin": {
                Enabled: true,
                Config: map[string]interface{}{
                    "api_key": "your-api-key",
                    "timeout": "30s",
                    "enabled": true,
                },
            },
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Fatal(app.Listen(":8080"))
}
```

## Plugin Features

### Hook System

Hooks allow plugins to intercept framework operations:

```go
// Available hook types
const (
    HookTypeStartup      HookType = "startup"
    HookTypeShutdown     HookType = "shutdown"
    HookTypePreRequest   HookType = "pre_request"
    HookTypePostRequest  HookType = "post_request"
    HookTypePreResponse  HookType = "pre_response"
    HookTypePostResponse HookType = "post_response"
    HookTypeError        HookType = "error"
)

// Register a hook
err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    // Access request context
    reqCtx := hookCtx.Context()
    
    // Modify request
    reqCtx.Request().Headers["X-Custom"] = "value"
    
    // Share data between hooks
    hookCtx.Set("start_time", time.Now())
    
    return nil
})
```

Hook priorities determine execution order (lower numbers execute first).

### Event System

Plugins can publish and subscribe to events:

```go
// Subscribe to events
err := ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    fmt.Printf("Event: %s from %s\n", event.Name, event.Source)
    fmt.Printf("Data: %v\n", event.Data)
    return nil
})

// Publish events
err = ctx.PublishEvent("my-plugin.action", map[string]interface{}{
    "action": "completed",
    "timestamp": time.Now(),
})
```

### Service Registry

Export and import services between plugins:

```go
// Export a service
type MyService struct {
    // ...
}

func (s *MyService) DoSomething() string {
    return "result"
}

err := ctx.ExportService("MyService", &MyService{})

// Import a service from another plugin
service, err := ctx.ImportService("other-plugin", "OtherService")
if err != nil {
    return err
}

// Type assert and use
if otherService, ok := service.(*OtherService); ok {
    result := otherService.DoSomething()
}
```

### Middleware Registration

Plugins can register middleware:

```go
err := ctx.RegisterMiddleware("my-middleware", func(c pkg.Context, next pkg.HandlerFunc) error {
    // Before request
    start := time.Now()
    
    // Call next handler
    err := next(c)
    
    // After request
    duration := time.Since(start)
    fmt.Printf("Request took %v\n", duration)
    
    return err
}, 100, []string{"/api/*"}) // Priority 100, apply to /api/* routes
```

### Plugin Storage

Each plugin has isolated key-value storage:

```go
storage := ctx.PluginStorage()

// Store data
err := storage.Set("key", "value")

// Retrieve data
value, err := storage.Get("key")

// List keys
keys, err := storage.List()

// Delete data
err = storage.Delete("key")

// Clear all data
err = storage.Clear()
```

### Permission System

Control what framework resources plugins can access:

```yaml
# In plugin.yaml
permissions:
  database: true    # Access database
  cache: true       # Access cache
  router: true      # Modify routes
  config: false     # Read configuration
  filesystem: false # File system access
  network: false    # Network requests
  exec: false       # Execute commands
```

Attempting to access unauthorized resources returns an error:

```go
// If database permission is false
db := ctx.Database() // Returns error: permission denied
```

## Plugin Configuration

### Framework Configuration

Configure plugins in your application:

```go
config := pkg.FrameworkConfig{
    PluginConfigs: map[string]pkg.PluginConfig{
        "my-plugin": {
            Enabled: true,
            Config: map[string]interface{}{
                "api_key": "secret-key",
                "timeout": "30s",
            },
        },
        "other-plugin": {
            Enabled: false, // Disable this plugin
        },
    },
}
```

### Dynamic Configuration Updates

Update plugin configuration at runtime:

```go
pluginMgr := app.PluginManager()

newConfig := map[string]interface{}{
    "api_key": "new-key",
    "timeout": "60s",
}

err := pluginMgr.UpdatePluginConfig("my-plugin", newConfig)
```

The plugin's `OnConfigChange()` method will be called.

## Plugin Management

### Listing Plugins

```go
pluginMgr := app.PluginManager()

// List all plugins
plugins := pluginMgr.ListPlugins()
for _, info := range plugins {
    fmt.Printf("Plugin: %s v%s - %s\n", 
        info.Name, info.Version, info.Status)
}
```

### Plugin Health

Monitor plugin health:

```go
// Get health for specific plugin
health := pluginMgr.GetPluginHealth("my-plugin")
fmt.Printf("Status: %s\n", health.Status)
fmt.Printf("Errors: %d\n", health.ErrorCount)
if health.LastError != nil {
    fmt.Printf("Last Error: %v at %s\n", 
        health.LastError, health.LastErrorAt)
}

// Get health for all plugins
allHealth := pluginMgr.GetAllHealth()
```

### Plugin Metrics

Access plugin metrics:

```go
// Get metrics for specific plugin
metrics := pluginMgr.GetPluginMetrics("my-plugin")
fmt.Printf("Init Duration: %v\n", metrics.InitDuration)
fmt.Printf("Start Duration: %v\n", metrics.StartDuration)
fmt.Printf("Hook Executions: %v\n", metrics.HookExecutions)

// Export Prometheus metrics
prometheusMetrics := pluginMgr.ExportPrometheusMetrics()
```

### Disabling Plugins

Disable a plugin at runtime:

```go
err := pluginMgr.DisablePlugin("my-plugin")
```

## Testing Plugins

### Unit Testing

```go
package myplugin

import (
    "testing"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func TestPluginInitialize(t *testing.T) {
    plugin := &MyPlugin{}
    
    // Create mock context
    ctx := pkg.NewMockPluginContext()
    ctx.SetConfig(map[string]interface{}{
        "api_key": "test-key",
    })
    
    err := plugin.Initialize(ctx)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
    
    if plugin.apiKey != "test-key" {
        t.Errorf("Expected api_key to be 'test-key', got '%s'", plugin.apiKey)
    }
}
```

### Integration Testing

```go
func TestPluginIntegration(t *testing.T) {
    config := pkg.FrameworkConfig{
        PluginConfigs: map[string]pkg.PluginConfig{
            "my-plugin": {
                Enabled: true,
                Config: map[string]interface{}{
                    "api_key": "test-key",
                },
            },
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        t.Fatal(err)
    }
    
    pluginMgr := app.PluginManager()
    
    // Verify plugin loaded
    if !pluginMgr.IsLoaded("my-plugin") {
        t.Error("Plugin not loaded")
    }
    
    // Test plugin functionality
    // ...
}
```

## Best Practices

### Error Handling

Always handle errors gracefully:

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Validate configuration
    if apiKey, ok := p.config["api_key"].(string); !ok || apiKey == "" {
        return fmt.Errorf("api_key is required and cannot be empty")
    }
    
    // Handle hook registration errors
    if err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, p.hook); err != nil {
        return fmt.Errorf("failed to register hook: %w", err)
    }
    
    return nil
}
```

### Resource Cleanup

Always clean up resources:

```go
func (p *MyPlugin) Cleanup() error {
    // Close connections
    if p.client != nil {
        p.client.Close()
    }
    
    // Clear references
    p.ctx = nil
    p.config = nil
    
    return nil
}
```

### Logging

Use the provided logger:

```go
func (p *MyPlugin) Start() error {
    logger := p.ctx.Logger()
    logger.Info("Plugin starting", "plugin", p.Name())
    
    // ... startup logic
    
    logger.Info("Plugin started successfully", "plugin", p.Name())
    return nil
}
```

### Metrics

Record plugin metrics:

```go
func (p *MyPlugin) processRequest() error {
    metrics := p.ctx.Metrics()
    
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RecordTiming("plugin.request.duration", duration, map[string]string{
            "plugin": p.Name(),
        })
    }()
    
    // ... process request
    
    return nil
}
```

### Dependency Management

Declare dependencies explicitly:

```go
func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{
        {
            Name:     "auth-plugin",
            Version:  ">=1.0.0",
            Optional: false,
        },
        {
            Name:     "cache-plugin",
            Version:  ">=2.0.0",
            Optional: true,
        },
    }
}
```

## API Reference

See [Plugin API](../api/plugins.md) for complete plugin API documentation.

## Examples

See the `plugins/template/` directory in the repository for a complete plugin template and the `examples/plugin_monitoring_example.go` file for plugin usage examples.

## Troubleshooting

### Plugin Not Loading

**Problem**: Plugin doesn't appear in plugin list

**Solutions**:
- Verify `init()` function calls `RegisterPlugin()`
- Check plugin is imported in main application
- Ensure plugin name matches in code and manifest
- Check for initialization errors in logs

### Permission Denied Errors

**Problem**: Plugin gets permission denied when accessing resources

**Solutions**:
- Check `permissions` section in `plugin.yaml`
- Verify required permissions are set to `true`
- Review framework logs for permission errors

### Dependency Resolution Failures

**Problem**: Plugin fails to load due to dependency issues

**Solutions**:
- Verify all dependencies are registered
- Check version constraints are satisfiable
- Ensure dependency plugins are enabled
- Review dependency graph: `pluginMgr.GetDependencyGraph()`

### Hook Not Executing

**Problem**: Registered hooks don't execute

**Solutions**:
- Verify hook type is correct
- Check hook priority (lower executes first)
- Ensure plugin is started successfully
- Review hook registration errors in logs

## Related Documentation

- [Architecture Overview](../architecture/overview.md) - Framework architecture
- [Extension Points](../architecture/extension-points.md) - Customization options
- [Middleware Guide](middleware.md) - Creating middleware
- [Configuration Guide](configuration.md) - Plugin configuration
