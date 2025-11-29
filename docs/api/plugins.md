---
title: "Plugins API"
description: "Plugin system interfaces for extending framework functionality"
category: "api"
tags: ["api", "plugins", "extensibility", "architecture"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "../guides/plugins.md"
---

# Plugins API

## Overview

The Plugin system provides a comprehensive architecture for extending the Rockstar Web Framework with custom functionality. Plugins can register hooks, middleware, services, and event handlers while maintaining isolation and security through a permission-based system.

**Primary Use Cases:**
- Adding custom authentication providers
- Implementing custom caching strategies
- Integrating third-party services
- Adding monitoring and analytics
- Extending routing capabilities
- Custom data validation and transformation

## Core Interfaces

### Plugin

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

**Description**: Core interface that all plugins must implement.

**Example**:
```go
type MyPlugin struct {
    config map[string]interface{}
    ctx    pkg.PluginContext
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Description() string {
    return "My custom plugin"
}

func (p *MyPlugin) Author() string {
    return "Your Name"
}

func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{
        {
            Name:             "core-plugin",
            Version:          ">=1.0.0",
            Optional:         false,
            FrameworkVersion: ">=1.0.0",
        },
    }
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register hooks
    ctx.RegisterHook(pkg.HookTypePreRequest, 10, p.preRequestHook)
    
    // Register middleware
    ctx.RegisterMiddleware("my-middleware", p.middleware, 10, []string{"/api/*"})
    
    return nil
}

func (p *MyPlugin) Start() error {
    // Start plugin services
    return nil
}

func (p *MyPlugin) Stop() error {
    // Stop plugin services
    return nil
}

func (p *MyPlugin) Cleanup() error {
    // Cleanup resources
    return nil
}

func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "enabled": map[string]interface{}{
            "type":    "boolean",
            "default": true,
        },
        "api_key": map[string]interface{}{
            "type":     "string",
            "required": true,
        },
    }
}

func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *MyPlugin) preRequestHook(ctx pkg.HookContext) error {
    // Pre-request logic
    return nil
}

func (p *MyPlugin) middleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Middleware logic
    return next(ctx)
}
```

### PluginContext

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

**Description**: Provides isolated access to framework services for plugins.

**Example**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Access framework services
    logger := ctx.Logger()
    logger.Info("Plugin initializing")

    // Get plugin configuration
    config := ctx.PluginConfig()
    apiKey := config["api_key"].(string)

    // Access plugin storage
    storage := ctx.PluginStorage()
    storage.Set("api_key", apiKey)

    // Register event handler
    ctx.SubscribeEvent("user.created", p.onUserCreated)

    // Export service for other plugins
    ctx.ExportService("my-service", &MyService{})

    return nil
}
```

## Hook System

### HookType

```go
const (
    HookTypeStartup      HookType = "startup"
    HookTypeShutdown     HookType = "shutdown"
    HookTypePreRequest   HookType = "pre_request"
    HookTypePostRequest  HookType = "post_request"
    HookTypePreResponse  HookType = "pre_response"
    HookTypePostResponse HookType = "post_response"
    HookTypeError        HookType = "error"
)
```

**Description**: Defines lifecycle points where plugins can hook into framework execution.

### RegisterHook

```go
func RegisterHook(hookType HookType, priority int, handler HookHandler) error
```

**Description**: Registers a hook handler at a specific lifecycle point.

**Parameters**:
- `hookType` (HookType): Type of hook
- `priority` (int): Execution priority (lower = earlier)
- `handler` (HookHandler): Hook handler function

**Returns**:
- `error`: Error if registration fails

**Example**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Register pre-request hook
    ctx.RegisterHook(pkg.HookTypePreRequest, 10, func(hctx pkg.HookContext) error {
        // Access request context
        reqCtx := hctx.Context()
        
        // Add custom header
        reqCtx.SetHeader("X-Plugin", "my-plugin")
        
        return nil
    })

    // Register error hook
    ctx.RegisterHook(pkg.HookTypeError, 20, func(hctx pkg.HookContext) error {
        // Log errors
        logger := ctx.Logger()
        logger.Error("Request error occurred")
        
        return nil
    })

    return nil
}
```

### HookContext

```go
type HookContext interface {
    // Request context (nil for non-request hooks)
    Context() Context

    // Hook metadata
    HookType() HookType
    PluginName() string

    // Data passing between hooks
    Set(key string, value interface{})
    Get(key string) interface{}

    // Control flow
    Skip() // Skip remaining hooks
    IsSkipped() bool
}
```

**Description**: Provides context for hook execution.

**Example**:
```go
func myHook(hctx pkg.HookContext) error {
    // Check if this is a request hook
    if hctx.Context() != nil {
        reqCtx := hctx.Context()
        
        // Store data for other hooks
        hctx.Set("start_time", time.Now())
        
        // Access data from previous hooks
        if value := hctx.Get("auth_token"); value != nil {
            token := value.(string)
            // Use token
        }
    }

    // Skip remaining hooks if needed
    if someCondition {
        hctx.Skip()
    }

    return nil
}
```

## Event System

### PublishEvent

```go
func PublishEvent(event string, data interface{}) error
```

**Description**: Publishes an event that other plugins can subscribe to.

**Parameters**:
- `event` (string): Event name
- `data` (interface{}): Event data

**Returns**:
- `error`: Error if publishing fails

**Example**:
```go
func (p *MyPlugin) onUserAction(ctx pkg.Context) error {
    // Publish event
    p.ctx.PublishEvent("user.action", map[string]interface{}{
        "user_id": ctx.User().ID,
        "action":  "login",
        "time":    time.Now(),
    })

    return nil
}
```

### SubscribeEvent

```go
func SubscribeEvent(event string, handler EventHandler) error
```

**Description**: Subscribes to an event published by other plugins.

**Parameters**:
- `event` (string): Event name to subscribe to
- `handler` (EventHandler): Event handler function

**Returns**:
- `error`: Error if subscription fails

**Example**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Subscribe to user events
    ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
        data := event.Data.(map[string]interface{})
        userID := data["user_id"].(string)
        
        // Handle event
        p.ctx.Logger().Info("User created", "user_id", userID)
        
        return nil
    })

    return nil
}
```

## Service Registry

### ExportService

```go
func ExportService(name string, service interface{}) error
```

**Description**: Exports a service that other plugins can import and use.

**Parameters**:
- `name` (string): Service name
- `service` (interface{}): Service implementation

**Returns**:
- `error`: Error if export fails

**Example**:
```go
type MyService struct {
    config map[string]interface{}
}

func (s *MyService) DoSomething(input string) (string, error) {
    return "processed: " + input, nil
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Export service
    service := &MyService{config: p.config}
    ctx.ExportService("my-service", service)

    return nil
}
```

### ImportService

```go
func ImportService(pluginName, serviceName string) (interface{}, error)
```

**Description**: Imports a service exported by another plugin.

**Parameters**:
- `pluginName` (string): Name of plugin that exported the service
- `serviceName` (string): Name of service to import

**Returns**:
- `interface{}`: Service implementation
- `error`: Error if import fails

**Example**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Import service from another plugin
    service, err := ctx.ImportService("other-plugin", "their-service")
    if err != nil {
        return err
    }

    // Use imported service
    myService := service.(*TheirService)
    result, _ := myService.DoSomething("input")

    return nil
}
```

## Plugin Storage

### PluginStorage

```go
type PluginStorage interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    List() ([]string, error)
    Clear() error
}
```

**Description**: Provides isolated key-value storage for plugins.

**Example**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    storage := ctx.PluginStorage()

    // Store data
    storage.Set("api_key", "secret-key")
    storage.Set("last_sync", time.Now())

    // Retrieve data
    apiKey, _ := storage.Get("api_key")
    
    // List all keys
    keys, _ := storage.List()

    // Delete data
    storage.Delete("old_key")

    // Clear all data
    storage.Clear()

    return nil
}
```

## Permissions

### PluginPermissions

```go
type PluginPermissions struct {
    AllowDatabase       bool
    AllowCache          bool
    AllowConfig         bool
    AllowRouter         bool
    AllowFileSystem     bool
    AllowNetwork        bool
    AllowExec           bool
    CustomPermissions   map[string]bool
}
```

**Description**: Defines what operations a plugin is allowed to perform.

**Example**:
```yaml
# plugin.yaml
name: my-plugin
version: 1.0.0
permissions:
  allow_database: true
  allow_cache: true
  allow_network: true
  allow_filesystem: false
  allow_exec: false
```

## Plugin Manifest

### Manifest Format

```yaml
name: my-plugin
version: 1.0.0
description: My custom plugin
author: Your Name
framework_version: ">=1.0.0"

dependencies:
  - name: core-plugin
    version: ">=1.0.0"
    optional: false

permissions:
  allow_database: true
  allow_cache: true
  allow_network: true

config_schema:
  enabled:
    type: boolean
    default: true
  api_key:
    type: string
    required: true
  timeout:
    type: integer
    default: 30

hooks:
  - type: pre_request
    priority: 10
  - type: post_response
    priority: 20

middleware:
  - name: auth-middleware
    priority: 10
    routes:
      - /api/*
      - /admin/*

events:
  publishes:
    - user.action
    - data.processed
  subscribes:
    - user.created
    - user.deleted

services:
  exports:
    - name: my-service
      description: Custom service
  imports:
    - plugin: other-plugin
      service: their-service
```

## Complete Plugin Example

Here's a complete example of a functional plugin:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

type AnalyticsPlugin struct {
    config  map[string]interface{}
    ctx     pkg.PluginContext
    storage pkg.PluginStorage
}

func (p *AnalyticsPlugin) Name() string {
    return "analytics-plugin"
}

func (p *AnalyticsPlugin) Version() string {
    return "1.0.0"
}

func (p *AnalyticsPlugin) Description() string {
    return "Analytics and tracking plugin"
}

func (p *AnalyticsPlugin) Author() string {
    return "Your Name"
}

func (p *AnalyticsPlugin) Dependencies() []pkg.PluginDependency {
    return nil
}

func (p *AnalyticsPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    p.storage = ctx.PluginStorage()
    p.config = ctx.PluginConfig()

    // Register hooks
    ctx.RegisterHook(pkg.HookTypePreRequest, 10, p.trackRequest)
    ctx.RegisterHook(pkg.HookTypePostResponse, 20, p.recordMetrics)

    // Subscribe to events
    ctx.SubscribeEvent("user.created", p.onUserCreated)

    // Export service
    service := &AnalyticsService{plugin: p}
    ctx.ExportService("analytics", service)

    return nil
}

func (p *AnalyticsPlugin) Start() error {
    p.ctx.Logger().Info("Analytics plugin started")
    return nil
}

func (p *AnalyticsPlugin) Stop() error {
    p.ctx.Logger().Info("Analytics plugin stopped")
    return nil
}

func (p *AnalyticsPlugin) Cleanup() error {
    return nil
}

func (p *AnalyticsPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "enabled": map[string]interface{}{
            "type":    "boolean",
            "default": true,
        },
        "api_key": map[string]interface{}{
            "type":     "string",
            "required": true,
        },
    }
}

func (p *AnalyticsPlugin) OnConfigChange(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *AnalyticsPlugin) trackRequest(hctx pkg.HookContext) error {
    ctx := hctx.Context()
    if ctx == nil {
        return nil
    }

    // Track request
    p.storage.Set("last_request", time.Now())
    
    // Increment counter
    count, _ := p.storage.Get("request_count")
    if count == nil {
        count = 0
    }
    p.storage.Set("request_count", count.(int)+1)

    return nil
}

func (p *AnalyticsPlugin) recordMetrics(hctx pkg.HookContext) error {
    ctx := hctx.Context()
    if ctx == nil {
        return nil
    }

    // Record metrics
    metrics := p.ctx.Metrics()
    metrics.IncrementCounter("analytics.requests", map[string]string{
        "path": ctx.Request().RequestURI,
    })

    return nil
}

func (p *AnalyticsPlugin) onUserCreated(event pkg.Event) error {
    data := event.Data.(map[string]interface{})
    userID := data["user_id"].(string)

    p.ctx.Logger().Info("User created", "user_id", userID)

    // Publish analytics event
    p.ctx.PublishEvent("analytics.user_created", map[string]interface{}{
        "user_id":   userID,
        "timestamp": time.Now(),
    })

    return nil
}

type AnalyticsService struct {
    plugin *AnalyticsPlugin
}

func (s *AnalyticsService) GetStats() map[string]interface{} {
    count, _ := s.plugin.storage.Get("request_count")
    lastRequest, _ := s.plugin.storage.Get("last_request")

    return map[string]interface{}{
        "request_count": count,
        "last_request":  lastRequest,
    }
}

// Plugin registration
func NewPlugin() pkg.Plugin {
    return &AnalyticsPlugin{}
}
```

## Best Practices

### Plugin Design

1. **Single responsibility**: Each plugin should have one clear purpose
2. **Minimal dependencies**: Reduce coupling with other plugins
3. **Graceful degradation**: Handle missing dependencies gracefully
4. **Resource cleanup**: Always clean up resources in Cleanup()

### Security

1. **Request minimal permissions**: Only request necessary permissions
2. **Validate input**: Always validate data from events and services
3. **Secure storage**: Don't store sensitive data in plugin storage
4. **Error handling**: Handle errors gracefully without exposing internals

### Performance

1. **Async operations**: Use goroutines for long-running tasks
2. **Efficient hooks**: Keep hook handlers fast and lightweight
3. **Caching**: Cache expensive computations
4. **Resource limits**: Implement timeouts and limits

### Testing

1. **Unit tests**: Test plugin logic independently
2. **Integration tests**: Test with framework
3. **Mock dependencies**: Use mocks for testing
4. **Error scenarios**: Test error handling

## See Also

- [Framework API](framework.md)
- [Context API](context.md)
- [Plugin Development Guide](../guides/plugins.md)
- [Architecture Overview](../architecture/overview.md)
