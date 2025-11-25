# Plugin Development Guide

## Table of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [Plugin Interface](#plugin-interface)
- [Plugin Context](#plugin-context)
- [Hook System](#hook-system)
- [Event System](#event-system)
- [Permission System](#permission-system)
- [Plugin Manifest](#plugin-manifest)
- [Configuration Management](#configuration-management)
- [Inter-Plugin Communication](#inter-plugin-communication)
- [Middleware Registration](#middleware-registration)
- [Storage and Persistence](#storage-and-persistence)
- [Testing Your Plugin](#testing-your-plugin)
- [Best Practices](#best-practices)
- [Advanced Topics](#advanced-topics)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)

## Introduction

The Rockstar Web Framework plugin system enables you to extend framework functionality through dynamically loadable modules. Plugins can hook into lifecycle events, register middleware, access framework services, and communicate with other plugins—all without modifying the core framework code.

### What Can Plugins Do?

- **Lifecycle Hooks**: Execute code at specific framework lifecycle points (startup, shutdown, request processing)
- **Middleware**: Register custom middleware for request/response processing
- **Route Registration**: Add new routes and handlers
- **Service Access**: Use framework services (database, cache, router, logger, metrics)
- **Inter-Plugin Communication**: Publish/subscribe to events and export/import services
- **Configuration**: Define and access plugin-specific configuration
- **Storage**: Store persistent plugin data in isolated storage

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Framework Core                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Plugin Manager                          │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │  │
│  │  │  Registry  │  │   Loader   │  │  Lifecycle │    │  │
│  │  └────────────┘  └────────────┘  └────────────┘    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Hook System                             │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │  │
│  │  │  Startup   │  │  Request   │  │  Shutdown  │    │  │
│  │  └────────────┘  └────────────┘  └────────────┘    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Event Bus                               │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │  │
│  │  │  Publish   │  │ Subscribe  │  │  Dispatch  │    │  │
│  │  └────────────┘  └────────────┘  └────────────┘    │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Your Plugin                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │    Hooks     │  │  Middleware  │  │    Events    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

**Requirements Addressed**: This guide covers all requirements from the plugin system specification (Requirements 1.1-12.5).

## Getting Started

### Prerequisites

- Go 1.21 or later
- Rockstar Web Framework v1.0.0 or later
- Basic understanding of Go interfaces and error handling

### Quick Start: Your First Plugin

Let's create a simple "Hello World" plugin in 5 minutes.

#### Step 1: Create Plugin Directory

```bash
mkdir -p my-hello-plugin
cd my-hello-plugin
```

#### Step 2: Initialize Go Module

```bash
go mod init github.com/yourusername/my-hello-plugin
go mod edit -require github.com/yourusername/rockstar@v1.0.0
```

#### Step 3: Create plugin.go

```go
package main

import (
    "github.com/yourusername/rockstar/pkg"
)

type HelloPlugin struct {
    ctx pkg.PluginContext
}

// Metadata methods (Requirement 1.3)
func (p *HelloPlugin) Name() string        { return "hello-plugin" }
func (p *HelloPlugin) Version() string     { return "1.0.0" }
func (p *HelloPlugin) Description() string { return "A simple hello world plugin" }
func (p *HelloPlugin) Author() string      { return "Your Name" }

// Dependencies (Requirement 1.4)
func (p *HelloPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{}
}

// Lifecycle methods (Requirement 1.1)
func (p *HelloPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    ctx.Logger().Info("Hello plugin initialized!")
    
    // Register a startup hook (Requirement 2.1)
    return ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hookCtx pkg.HookContext) error {
        ctx.Logger().Info("Hello from startup hook!")
        return nil
    })
}

func (p *HelloPlugin) Start() error {
    p.ctx.Logger().Info("Hello plugin started!")
    return nil
}

func (p *HelloPlugin) Stop() error {
    p.ctx.Logger().Info("Hello plugin stopped!")
    return nil
}

func (p *HelloPlugin) Cleanup() error {
    p.ctx.Logger().Info("Hello plugin cleaned up!")
    return nil
}

// Configuration (Requirement 8.1, 8.5)
func (p *HelloPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "greeting": map[string]interface{}{
            "type":        "string",
            "default":     "Hello, World!",
            "description": "The greeting message",
        },
    }
}

func (p *HelloPlugin) OnConfigChange(config map[string]interface{}) error {
    if greeting, ok := config["greeting"].(string); ok {
        p.ctx.Logger().Info("Greeting changed to: " + greeting)
    }
    return nil
}

// Factory function
func NewPlugin() pkg.Plugin {
    return &HelloPlugin{}
}
```

#### Step 4: Create plugin.yaml

```yaml
name: hello-plugin
version: 1.0.0
description: A simple hello world plugin
author: Your Name

framework:
  version: ">=1.0.0"

dependencies: []

permissions:
  database: false
  cache: false
  router: false
  config: true
  filesystem: false
  network: false
  exec: false

config:
  greeting:
    type: string
    default: "Hello, World!"
    description: The greeting message

hooks:
  - type: startup
    priority: 100

events:
  publishes: []
  subscribes: []

exports: []
```

#### Step 5: Load Your Plugin

```go
// In your application
framework.PluginManager().LoadPlugin("./my-hello-plugin", pkg.PluginConfig{
    Enabled: true,
    Config: map[string]interface{}{
        "greeting": "Hello from my plugin!",
    },
})
```

Congratulations! You've created your first plugin. 🎉

## Plugin Interface

Every plugin must implement the `Plugin` interface defined in `pkg/plugin.go`.

### Interface Definition

```go
type Plugin interface {
    // Metadata (Requirement 1.3)
    Name() string
    Version() string
    Description() string
    Author() string
    
    // Dependencies (Requirement 1.4)
    Dependencies() []PluginDependency
    
    // Lifecycle (Requirement 1.1)
    Initialize(ctx PluginContext) error
    Start() error
    Stop() error
    Cleanup() error
    
    // Configuration (Requirement 8.1, 8.2)
    ConfigSchema() map[string]interface{}
    OnConfigChange(config map[string]interface{}) error
}
```

### Metadata Methods

#### Name() string

**Requirement**: 1.3

Returns the unique identifier for your plugin. This name is used for:
- Plugin registry lookups
- Dependency resolution
- Inter-plugin communication
- Logging and metrics

**Rules**:
- Must be unique across all plugins
- Use lowercase with hyphens (kebab-case)
- Should be descriptive and concise

```go
func (p *MyPlugin) Name() string {
    return "my-awesome-plugin"
}
```

#### Version() string

**Requirement**: 1.3

Returns the plugin version using semantic versioning (MAJOR.MINOR.PATCH).

```go
func (p *MyPlugin) Version() string {
    return "1.2.3"
}
```

**Versioning Guidelines**:
- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes

#### Description() string

**Requirement**: 1.3

Returns a brief description of what your plugin does.

```go
func (p *MyPlugin) Description() string {
    return "Provides advanced caching with Redis backend"
}
```

#### Author() string

**Requirement**: 1.3

Returns the plugin author information.

```go
func (p *MyPlugin) Author() string {
    return "John Doe <john@example.com>"
}
```

### Dependencies

#### Dependencies() []PluginDependency

**Requirement**: 1.4, 5.1-5.5

Declares plugin dependencies on other plugins or framework versions.

```go
type PluginDependency struct {
    Name             string // Plugin name
    Version          string // Semantic version constraint
    Optional         bool   // Whether dependency is optional
    FrameworkVersion string // Framework version requirement
}
```

**Example**:

```go
func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{
        {
            Name:    "auth-plugin",
            Version: ">=1.0.0,<2.0.0",
            Optional: false,
        },
        {
            Name:    "cache-plugin",
            Version: ">=2.0.0",
            Optional: true,
        },
        {
            FrameworkVersion: ">=1.0.0,<2.0.0",
        },
    }
}
```

**Version Constraint Syntax**:
- `>=1.0.0`: Greater than or equal to 1.0.0
- `<2.0.0`: Less than 2.0.0
- `>=1.0.0,<2.0.0`: Between 1.0.0 and 2.0.0
- `^1.0.0`: Compatible with 1.0.0 (>=1.0.0, <2.0.0)
- `~1.2.0`: Approximately 1.2.0 (>=1.2.0, <1.3.0)

**Dependency Resolution** (Requirement 5.2):
- Dependencies are loaded before dependent plugins
- Circular dependencies are detected and prevented (Requirement 5.5)
- Missing dependencies cause load failure (Requirement 5.3)
- Version mismatches are detected (Requirement 5.4)

### Lifecycle Methods

#### Initialize(ctx PluginContext) error

**Requirement**: 1.1, 1.5

Called once when the plugin is loaded. Use this method to:
- Store the PluginContext
- Register hooks
- Subscribe to events
- Validate configuration
- Initialize internal state

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Validate configuration
    config := ctx.PluginConfig()
    if _, ok := config["api_key"]; !ok {
        return fmt.Errorf("api_key is required")
    }
    
    // Register hooks (Requirement 2.1-2.5)
    if err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, p.handlePreRequest); err != nil {
        return fmt.Errorf("failed to register hook: %w", err)
    }
    
    // Subscribe to events (Requirement 10.2)
    if err := ctx.SubscribeEvent("user.created", p.handleUserCreated); err != nil {
        return fmt.Errorf("failed to subscribe to event: %w", err)
    }
    
    ctx.Logger().Info("Plugin initialized successfully")
    return nil
}
```

**Important**: Do not start background goroutines or open connections in Initialize(). Use Start() for that.

#### Start() error

**Requirement**: 1.1

Called after all plugins are initialized. Use this method to:
- Start background goroutines
- Open connections (database, network)
- Begin processing

```go
func (p *MyPlugin) Start() error {
    // Start background worker
    go p.backgroundWorker()
    
    // Open connections
    if err := p.connectToExternalService(); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    
    p.ctx.Logger().Info("Plugin started successfully")
    return nil
}
```

#### Stop() error

**Requirement**: 1.1

Called when the plugin is being stopped (shutdown or hot reload). Use this method to:
- Stop background goroutines
- Close connections
- Flush buffers
- Save state

```go
func (p *MyPlugin) Stop() error {
    // Signal background workers to stop
    close(p.stopChan)
    
    // Wait for workers to finish
    p.wg.Wait()
    
    // Close connections
    if p.conn != nil {
        p.conn.Close()
    }
    
    p.ctx.Logger().Info("Plugin stopped successfully")
    return nil
}
```

**Important**: Stop() should be idempotent and safe to call multiple times.

#### Cleanup() error

**Requirement**: 1.1

Called when the plugin is being unloaded. Use this method to:
- Release resources
- Delete temporary files
- Final cleanup

```go
func (p *MyPlugin) Cleanup() error {
    // Remove temporary files
    if err := os.RemoveAll(p.tempDir); err != nil {
        p.ctx.Logger().Error("Failed to remove temp dir: " + err.Error())
    }
    
    // Clear caches
    p.cache = nil
    
    p.ctx.Logger().Info("Plugin cleaned up successfully")
    return nil
}
```

### Configuration Methods

#### ConfigSchema() map[string]interface{}

**Requirement**: 8.5

Defines the configuration schema with default values and descriptions.

```go
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
        "max_retries": map[string]interface{}{
            "type":        "integer",
            "default":     3,
            "description": "Maximum number of retries",
        },
        "enabled_features": map[string]interface{}{
            "type":        "array",
            "items":       "string",
            "default":     []string{},
            "description": "List of enabled features",
        },
    }
}
```

**Supported Types**:
- `string`: Text values
- `integer`: Whole numbers
- `float`: Decimal numbers
- `boolean`: true/false
- `duration`: Time durations (e.g., "30s", "5m")
- `array`: Lists of values
- `object`: Nested structures

#### OnConfigChange(config map[string]interface{}) error

**Requirement**: 8.2

Called when plugin configuration is updated at runtime.

```go
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    // Update timeout
    if timeout, ok := config["timeout"].(string); ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout: %w", err)
        }
        p.timeout = duration
    }
    
    // Update enabled features
    if features, ok := config["enabled_features"].([]interface{}); ok {
        p.enabledFeatures = make([]string, len(features))
        for i, f := range features {
            p.enabledFeatures[i] = f.(string)
        }
    }
    
    p.ctx.Logger().Info("Configuration updated")
    return nil
}
```

## Plugin Context

The `PluginContext` provides access to framework services and plugin-specific functionality.

### Interface Definition

```go
type PluginContext interface {
    // Core framework access (Requirement 4.1-4.6)
    Router() RouterEngine
    Logger() Logger
    Metrics() MetricsCollector
    Database() DatabaseManager
    Cache() CacheManager
    Config() ConfigManager
    
    // Plugin-specific (Requirement 8.1, 8.3)
    PluginConfig() map[string]interface{}
    PluginStorage() PluginStorage
    
    // Hook registration (Requirement 2.1-2.5)
    RegisterHook(hookType HookType, priority int, handler HookHandler) error
    
    // Event system (Requirement 10.1, 10.2)
    PublishEvent(event string, data interface{}) error
    SubscribeEvent(event string, handler EventHandler) error
    
    // Service export/import (Requirement 10.3, 10.4)
    ExportService(name string, service interface{}) error
    ImportService(pluginName, serviceName string) (interface{}, error)
    
    // Middleware registration (Requirement 6.1, 6.2)
    RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error
    UnregisterMiddleware(name string) error
}
```

### Framework Services

#### Router() RouterEngine

**Requirement**: 4.1

Access the framework router to register routes and handlers.

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register a route
    router := ctx.Router()
    router.GET("/api/myplugin/status", func(c pkg.Context) error {
        return c.JSON(200, map[string]string{
            "status": "ok",
            "plugin": p.Name(),
        })
    })
    
    return nil
}
```

**Permission Required**: `router: true`

#### Logger() Logger

**Requirement**: 4.5

Access the framework logger for consistent logging.

```go
func (p *MyPlugin) someMethod() {
    logger := p.ctx.Logger()
    
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
    
    // Structured logging
    logger.WithFields(map[string]interface{}{
        "user_id": 123,
        "action":  "login",
    }).Info("User logged in")
}
```

**Permission Required**: None (always available)

#### Metrics() MetricsCollector

**Requirement**: 4.6

Access the metrics collector to record plugin metrics.

```go
func (p *MyPlugin) recordMetrics() {
    metrics := p.ctx.Metrics()
    
    // Increment counter
    metrics.IncrementCounter("myplugin.requests.total")
    
    // Record gauge
    metrics.RecordGauge("myplugin.queue.size", float64(len(p.queue)))
    
    // Record histogram
    metrics.RecordHistogram("myplugin.request.duration", duration.Seconds())
}
```

**Permission Required**: None (always available)

#### Database() DatabaseManager

**Requirement**: 4.2

Access the database manager for data persistence.

```go
func (p *MyPlugin) saveData(data interface{}) error {
    db := p.ctx.Database()
    
    // Execute query
    result, err := db.Exec("INSERT INTO plugin_data (key, value) VALUES (?, ?)", 
        "mykey", data)
    if err != nil {
        return err
    }
    
    return nil
}
```

**Permission Required**: `database: true`

#### Cache() CacheManager

**Requirement**: 4.3

Access the cache manager for caching operations.

```go
func (p *MyPlugin) getCachedData(key string) (interface{}, error) {
    cache := p.ctx.Cache()
    
    // Get from cache
    value, err := cache.Get(key)
    if err == nil {
        return value, nil
    }
    
    // Cache miss - fetch and cache
    data := p.fetchData(key)
    cache.Set(key, data, 5*time.Minute)
    
    return data, nil
}
```

**Permission Required**: `cache: true`

#### Config() ConfigManager

**Requirement**: 4.4

Access the framework configuration.

```go
func (p *MyPlugin) readConfig() {
    config := p.ctx.Config()
    
    // Get configuration value
    dbHost := config.GetString("database.host")
    dbPort := config.GetInt("database.port")
}
```

**Permission Required**: `config: true`

### Plugin-Specific Methods

#### PluginConfig() map[string]interface{}

**Requirement**: 8.1

Get plugin-specific configuration.

```go
func (p *MyPlugin) getConfig() {
    config := p.ctx.PluginConfig()
    
    // Access configuration values
    apiKey := config["api_key"].(string)
    timeout := config["timeout"].(string)
    
    // With type assertion and error checking
    if apiKey, ok := config["api_key"].(string); ok {
        p.apiKey = apiKey
    }
}
```

**Configuration Isolation** (Requirement 8.1): Each plugin only sees its own configuration.

#### PluginStorage() PluginStorage

**Requirement**: 8.3

Access isolated plugin storage.

```go
type PluginStorage interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    List() ([]string, error)
    Clear() error
}
```

**Example**:

```go
func (p *MyPlugin) saveState() error {
    storage := p.ctx.PluginStorage()
    
    // Save data
    return storage.Set("last_sync", time.Now())
}

func (p *MyPlugin) loadState() error {
    storage := p.ctx.PluginStorage()
    
    // Load data
    lastSync, err := storage.Get("last_sync")
    if err != nil {
        return err
    }
    
    p.lastSync = lastSync.(time.Time)
    return nil
}
```

**Storage Isolation** (Requirement 8.3): Each plugin has isolated storage that other plugins cannot access.


## Hook System

Hooks allow plugins to execute code at specific framework lifecycle points.

### Hook Types

**Requirement**: 2.1-2.5

```go
const (
    HookTypeStartup      HookType = "startup"       // Framework initialization
    HookTypeShutdown     HookType = "shutdown"      // Graceful shutdown
    HookTypePreRequest   HookType = "pre_request"   // Before routing
    HookTypePostRequest  HookType = "post_request"  // After handler execution
    HookTypePreResponse  HookType = "pre_response"  // Before sending response
    HookTypePostResponse HookType = "post_response" // After sending response
    HookTypeError        HookType = "error"         // Error handling
)
```

### Registering Hooks

**Requirement**: 2.1-2.5

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register startup hook (Requirement 2.1)
    ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hookCtx pkg.HookContext) error {
        p.ctx.Logger().Info("Startup hook executed")
        return nil
    })
    
    // Register pre-request hook (Requirement 2.3)
    ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
        reqCtx := hookCtx.Context()
        p.ctx.Logger().Info("Processing request: " + reqCtx.Path())
        return nil
    })
    
    // Register post-request hook (Requirement 2.4)
    ctx.RegisterHook(pkg.HookTypePostRequest, 100, func(hookCtx pkg.HookContext) error {
        reqCtx := hookCtx.Context()
        p.ctx.Logger().Info("Request completed: " + reqCtx.Path())
        return nil
    })
    
    // Register shutdown hook (Requirement 2.2)
    ctx.RegisterHook(pkg.HookTypeShutdown, 100, func(hookCtx pkg.HookContext) error {
        p.ctx.Logger().Info("Shutdown hook executed")
        return nil
    })
    
    return nil
}
```

### Hook Priority

**Requirement**: 2.6

Hooks are executed in priority order (highest priority first).

```go
// High priority - executes first
ctx.RegisterHook(pkg.HookTypePreRequest, 200, authHook)

// Medium priority
ctx.RegisterHook(pkg.HookTypePreRequest, 100, loggingHook)

// Low priority - executes last
ctx.RegisterHook(pkg.HookTypePreRequest, 50, metricsHook)
```

**Priority Ranges**:
- **300-1000**: Critical (authentication, security)
- **200-299**: High (rate limiting, validation)
- **100-199**: Normal (default)
- **50-99**: Low (logging, metrics)
- **0-49**: Lowest (cleanup, finalization)

### Hook Context

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

**Example: Passing Data Between Hooks**

```go
// Plugin A sets data
ctx.RegisterHook(pkg.HookTypePreRequest, 200, func(hookCtx pkg.HookContext) error {
    hookCtx.Set("user_id", 123)
    return nil
})

// Plugin B reads data
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    if userID := hookCtx.Get("user_id"); userID != nil {
        p.ctx.Logger().Info("User ID: " + fmt.Sprint(userID))
    }
    return nil
})
```

### Error Handling

**Requirement**: 2.7

Hook errors are logged but don't stop other hooks from executing.

```go
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    if err := p.validateRequest(hookCtx.Context()); err != nil {
        // Error is logged, but other hooks continue
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
})
```

### Hook Execution Order

**Requirement**: 2.6

For multiple plugins with hooks at the same point:

1. Hooks are sorted by priority (descending)
2. Hooks with the same priority execute in plugin load order
3. Errors in one hook don't prevent other hooks from executing

```
Request arrives
    ↓
Pre-Request Hooks (priority order)
    ↓
Router matches route
    ↓
Middleware chain
    ↓
Handler executes
    ↓
Post-Request Hooks (priority order)
    ↓
Pre-Response Hooks (priority order)
    ↓
Response sent
    ↓
Post-Response Hooks (priority order)
```

## Event System

The event system enables inter-plugin communication through publish/subscribe.

### Publishing Events

**Requirement**: 10.1

```go
func (p *MyPlugin) createUser(user User) error {
    // Create user
    if err := p.saveUser(user); err != nil {
        return err
    }
    
    // Publish event
    return p.ctx.PublishEvent("user.created", map[string]interface{}{
        "user_id":  user.ID,
        "username": user.Username,
        "email":    user.Email,
        "timestamp": time.Now(),
    })
}
```

### Subscribing to Events

**Requirement**: 10.2

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Subscribe to event
    return ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
        // Extract event data
        data := event.Data.(map[string]interface{})
        userID := data["user_id"].(int)
        
        // Handle event
        p.ctx.Logger().Info(fmt.Sprintf("User created: %d", userID))
        
        // Send welcome email
        return p.sendWelcomeEmail(userID)
    })
}
```

### Event Structure

```go
type Event struct {
    Name      string      // Event name (e.g., "user.created")
    Data      interface{} // Event payload
    Source    string      // Publishing plugin name
    Timestamp time.Time   // Event timestamp
}
```

### Event Naming Conventions

Use dot notation for hierarchical event names:

```go
// Good
"user.created"
"user.updated"
"user.deleted"
"order.placed"
"order.shipped"
"payment.processed"

// Avoid
"UserCreated"
"user_created"
"CREATE_USER"
```

### Event Error Handling

**Requirement**: 10.5

Errors in event handlers don't crash the publishing plugin or other subscribers.

```go
ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    if err := p.processEvent(event); err != nil {
        // Error is logged, but other subscribers continue
        return fmt.Errorf("failed to process event: %w", err)
    }
    return nil
})
```

### Asynchronous Event Delivery

Events are delivered asynchronously to subscribers:

```go
// Publish returns immediately
p.ctx.PublishEvent("user.created", userData)

// Subscribers are called asynchronously
// Your code continues without waiting
```

## Permission System

The permission system controls plugin access to framework services.

### Permission Types

**Requirement**: 11.4

```go
type PluginPermissions struct {
    AllowDatabase    bool // Database access
    AllowCache       bool // Cache access
    AllowRouter      bool // Router access (routes, middleware)
    AllowConfig      bool // Configuration access
    AllowFileSystem  bool // Filesystem access
    AllowNetwork     bool // Network access
    AllowExec        bool // Execute external commands
    CustomPermissions map[string]bool // Custom permissions
}
```

### Declaring Permissions

**Requirement**: 11.1

Permissions are declared in the plugin manifest:

```yaml
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: false
  network: false
  exec: false
```

### Permission Enforcement

**Requirement**: 11.2, 11.3

The framework verifies permissions before allowing access:

```go
func (p *MyPlugin) accessDatabase() error {
    // This will fail if database permission is not granted
    db := p.ctx.Database()
    
    // Permission denied error is logged as security violation
    result, err := db.Query("SELECT * FROM users")
    if err != nil {
        return err // "permission denied: database access not allowed"
    }
    
    return nil
}
```

### Permission Best Practices

1. **Principle of Least Privilege**: Request only permissions you need

```yaml
# ❌ Bad: Requesting unnecessary permissions
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: true  # Not needed
  network: true     # Not needed
  exec: true        # Dangerous!

# ✅ Good: Minimal permissions
permissions:
  database: true
  cache: true
  router: false
  config: true
  filesystem: false
  network: false
  exec: false
```

2. **Document Why**: Explain why each permission is needed

```yaml
# In your plugin documentation
## Required Permissions

- **database**: Store user preferences and plugin state
- **cache**: Cache API responses for performance
- **config**: Read API endpoint configuration
```

3. **Graceful Degradation**: Handle missing permissions gracefully

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Try to use cache if available
    if cache := ctx.Cache(); cache != nil {
        p.cache = cache
        p.ctx.Logger().Info("Cache enabled")
    } else {
        p.ctx.Logger().Warn("Cache not available, running without cache")
    }
    
    return nil
}
```

### Custom Permissions

**Requirement**: 11.4

Define custom permissions for plugin-specific access control:

```yaml
permissions:
  database: true
  custom:
    user_management: true
    role_management: true
    billing_access: false
```

```go
func (p *MyPlugin) checkCustomPermission(permission string) bool {
    // Check custom permission
    // Implementation depends on your permission checker
    return p.hasPermission(permission)
}
```

## Plugin Manifest

The manifest file describes plugin metadata, dependencies, and requirements.

### Manifest Format

**Requirement**: 12.1

Manifests can be in YAML or JSON format.

#### YAML Format (plugin.yaml)

```yaml
# Basic metadata (Requirement 12.2)
name: my-plugin
version: 1.0.0
description: A comprehensive example plugin
author: John Doe <john@example.com>

# Framework compatibility (Requirement 5.1)
framework:
  version: ">=1.0.0,<2.0.0"

# Dependencies (Requirement 12.3)
dependencies:
  - name: auth-plugin
    version: ">=1.0.0"
    optional: false
  - name: cache-plugin
    version: ">=2.0.0"
    optional: true

# Permissions (Requirement 12.4)
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: false
  network: false
  exec: false

# Configuration schema
config:
  api_key:
    type: string
    required: true
    description: API key for external service
  timeout:
    type: duration
    default: 30s
    description: Request timeout
  max_retries:
    type: integer
    default: 3
    description: Maximum number of retries
  enabled_features:
    type: array
    items: string
    default: []
    description: List of enabled features

# Hooks
hooks:
  - type: startup
    priority: 100
  - type: pre_request
    priority: 50
  - type: shutdown
    priority: 100

# Events
events:
  publishes:
    - user.created
    - user.updated
    - user.deleted
  subscribes:
    - auth.login
    - auth.logout

# Exported services
exports:
  - name: UserService
    description: User management service
  - name: NotificationService
    description: Notification service
```

#### JSON Format (plugin.json)

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "A comprehensive example plugin",
  "author": "John Doe <john@example.com>",
  "framework": {
    "version": ">=1.0.0,<2.0.0"
  },
  "dependencies": [
    {
      "name": "auth-plugin",
      "version": ">=1.0.0",
      "optional": false
    }
  ],
  "permissions": {
    "database": true,
    "cache": true,
    "router": true,
    "config": true,
    "filesystem": false,
    "network": false,
    "exec": false
  },
  "config": {
    "api_key": {
      "type": "string",
      "required": true,
      "description": "API key for external service"
    }
  },
  "hooks": [
    {
      "type": "startup",
      "priority": 100
    }
  ],
  "events": {
    "publishes": ["user.created"],
    "subscribes": ["auth.login"]
  },
  "exports": [
    {
      "name": "UserService",
      "description": "User management service"
    }
  ]
}
```

### Manifest Validation

**Requirement**: 12.5

Invalid manifests prevent plugin loading:

```yaml
# ❌ Invalid: Missing required fields
name: my-plugin
# Missing version, description, author

# ❌ Invalid: Invalid version format
version: 1.0  # Should be 1.0.0

# ❌ Invalid: Invalid dependency version constraint
dependencies:
  - name: other-plugin
    version: "invalid"  # Should be semantic version constraint

# ✅ Valid: All required fields present
name: my-plugin
version: 1.0.0
description: My plugin
author: John Doe
```

## Configuration Management

### Accessing Configuration

**Requirement**: 8.1, 8.4

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Get plugin configuration
    config := ctx.PluginConfig()
    
    // Simple values
    apiKey := config["api_key"].(string)
    timeout := config["timeout"].(string)
    
    // Nested configuration (Requirement 8.4)
    if nested, ok := config["database"].(map[string]interface{}); ok {
        host := nested["host"].(string)
        port := nested["port"].(int)
    }
    
    // Arrays
    if features, ok := config["enabled_features"].([]interface{}); ok {
        for _, f := range features {
            feature := f.(string)
            p.enableFeature(feature)
        }
    }
    
    return nil
}
```

### Configuration Changes

**Requirement**: 8.2

Handle runtime configuration updates:

```go
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Update timeout
    if timeout, ok := config["timeout"].(string); ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout: %w", err)
        }
        p.timeout = duration
        p.ctx.Logger().Info("Timeout updated to: " + timeout)
    }
    
    // Update features
    if features, ok := config["enabled_features"].([]interface{}); ok {
        p.enabledFeatures = make([]string, len(features))
        for i, f := range features {
            p.enabledFeatures[i] = f.(string)
        }
        p.ctx.Logger().Info("Features updated")
    }
    
    return nil
}
```

### Default Values

**Requirement**: 8.5

Provide defaults in ConfigSchema():

```go
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "timeout": map[string]interface{}{
            "type":        "duration",
            "default":     "30s",  // Default value
            "description": "Request timeout",
        },
        "max_retries": map[string]interface{}{
            "type":        "integer",
            "default":     3,  // Default value
            "description": "Maximum retries",
        },
    }
}
```

If a configuration key is missing, the default value is used.

## Inter-Plugin Communication

### Exporting Services

**Requirement**: 10.3, 10.4

```go
// Define service interface
type UserService interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
    DeleteUser(id int) error
}

// Implement service
type userServiceImpl struct {
    plugin *MyPlugin
}

func (s *userServiceImpl) GetUser(id int) (*User, error) {
    return s.plugin.getUser(id)
}

// Export service in Initialize
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Export service
    service := &userServiceImpl{plugin: p}
    return ctx.ExportService("UserService", service)
}
```

### Importing Services

**Requirement**: 10.3, 10.4

```go
func (p *MyPlugin) Start() error {
    // Import service from another plugin
    service, err := p.ctx.ImportService("user-plugin", "UserService")
    if err != nil {
        return fmt.Errorf("failed to import UserService: %w", err)
    }
    
    // Type assert to interface
    userService := service.(UserService)
    
    // Use the service
    user, err := userService.GetUser(123)
    if err != nil {
        return err
    }
    
    p.ctx.Logger().Info("Got user: " + user.Username)
    return nil
}
```

### Service Discovery

**Requirement**: 10.4

```go
// Check if a service is available
service, err := p.ctx.ImportService("other-plugin", "SomeService")
if err != nil {
    p.ctx.Logger().Warn("SomeService not available, using fallback")
    // Use fallback implementation
} else {
    // Use imported service
}
```

### Error Isolation

**Requirement**: 10.5

Errors in inter-plugin calls don't crash either plugin:

```go
func (p *MyPlugin) callOtherPlugin() error {
    service, err := p.ctx.ImportService("other-plugin", "Service")
    if err != nil {
        // Service not available - handle gracefully
        return fmt.Errorf("service not available: %w", err)
    }
    
    // Call may fail - handle error
    if err := service.DoSomething(); err != nil {
        // Error is returned, plugins continue running
        return fmt.Errorf("service call failed: %w", err)
    }
    
    return nil
}
```

## Middleware Registration

### Global Middleware

**Requirement**: 6.1

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register global middleware
    return ctx.RegisterMiddleware(
        "my-middleware",
        func(c pkg.Context) error {
            // Pre-processing
            start := time.Now()
            
            // Continue to next middleware/handler
            err := c.Next()
            
            // Post-processing
            duration := time.Since(start)
            p.ctx.Logger().Info(fmt.Sprintf("Request took %v", duration))
            
            return err
        },
        100,  // Priority
        []string{},  // Empty = global
    )
}
```

### Route-Specific Middleware

**Requirement**: 6.2

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register middleware for specific routes
    return ctx.RegisterMiddleware(
        "auth-middleware",
        func(c pkg.Context) error {
            // Check authentication
            if !p.isAuthenticated(c) {
                return c.JSON(401, map[string]string{
                    "error": "unauthorized",
                })
            }
            return c.Next()
        },
        200,  // Priority
        []string{"/api/*", "/admin/*"},  // Specific routes
    )
}
```

### Middleware Priority

**Requirement**: 6.3

```go
// High priority - executes first
ctx.RegisterMiddleware("auth", authMiddleware, 200, []string{})

// Medium priority
ctx.RegisterMiddleware("logging", loggingMiddleware, 100, []string{})

// Low priority - executes last
ctx.RegisterMiddleware("metrics", metricsMiddleware, 50, []string{})
```

### Middleware Cleanup

**Requirement**: 6.5

Middleware is automatically removed when the plugin is unloaded:

```go
func (p *MyPlugin) Stop() error {
    // Middleware is automatically unregistered
    // No manual cleanup needed
    return nil
}

// Or manually unregister
func (p *MyPlugin) disableMiddleware() error {
    return p.ctx.UnregisterMiddleware("my-middleware")
}
```

## Storage and Persistence

### Plugin Storage

**Requirement**: 8.3

Each plugin has isolated storage:

```go
func (p *MyPlugin) saveData() error {
    storage := p.ctx.PluginStorage()
    
    // Save simple value
    if err := storage.Set("counter", 42); err != nil {
        return err
    }
    
    // Save complex data
    data := map[string]interface{}{
        "last_sync": time.Now(),
        "status":    "active",
        "count":     100,
    }
    return storage.Set("state", data)
}

func (p *MyPlugin) loadData() error {
    storage := p.ctx.PluginStorage()
    
    // Load value
    value, err := storage.Get("counter")
    if err != nil {
        return err
    }
    counter := value.(int)
    
    // Load complex data
    value, err = storage.Get("state")
    if err != nil {
        return err
    }
    state := value.(map[string]interface{})
    
    return nil
}
```

### Storage Operations

```go
type PluginStorage interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    List() ([]string, error)
    Clear() error
}
```

**Example**:

```go
func (p *MyPlugin) manageStorage() error {
    storage := p.ctx.PluginStorage()
    
    // Set value
    storage.Set("key1", "value1")
    storage.Set("key2", 123)
    
    // Get value
    val, err := storage.Get("key1")
    if err != nil {
        return err
    }
    
    // List all keys
    keys, err := storage.List()
    for _, key := range keys {
        p.ctx.Logger().Info("Key: " + key)
    }
    
    // Delete key
    storage.Delete("key1")
    
    // Clear all
    storage.Clear()
    
    return nil
}
```

### Storage Isolation

**Requirement**: 8.3

Each plugin's storage is isolated:

```go
// Plugin A
storageA := pluginA.ctx.PluginStorage()
storageA.Set("key", "value from A")

// Plugin B
storageB := pluginB.ctx.PluginStorage()
storageB.Set("key", "value from B")

// Each plugin sees only its own data
valueA, _ := storageA.Get("key")  // "value from A"
valueB, _ := storageB.Get("key")  // "value from B"
```


## Testing Your Plugin

### Unit Testing

Create unit tests for your plugin logic:

```go
// plugin_test.go
package main

import (
    "testing"
    "github.com/yourusername/rockstar/pkg"
)

func TestPluginMetadata(t *testing.T) {
    plugin := NewPlugin()
    
    if plugin.Name() != "my-plugin" {
        t.Errorf("Expected name 'my-plugin', got '%s'", plugin.Name())
    }
    
    if plugin.Version() != "1.0.0" {
        t.Errorf("Expected version '1.0.0', got '%s'", plugin.Version())
    }
}

func TestPluginInitialize(t *testing.T) {
    plugin := NewPlugin().(*MyPlugin)
    
    // Create mock context
    ctx := pkg.NewMockPluginContext()
    
    // Initialize plugin
    err := plugin.Initialize(ctx)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
    
    // Verify initialization
    if plugin.ctx == nil {
        t.Error("Plugin context not set")
    }
}

func TestPluginConfiguration(t *testing.T) {
    plugin := NewPlugin().(*MyPlugin)
    
    // Test configuration schema
    schema := plugin.ConfigSchema()
    if schema == nil {
        t.Error("ConfigSchema returned nil")
    }
    
    // Test configuration change
    config := map[string]interface{}{
        "timeout": "60s",
    }
    err := plugin.OnConfigChange(config)
    if err != nil {
        t.Errorf("OnConfigChange failed: %v", err)
    }
}
```

### Integration Testing

Test plugin integration with the framework:

```go
func TestPluginIntegration(t *testing.T) {
    // Create framework
    config := &pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            Host: "localhost",
            Port: 8080,
        },
    }
    framework, err := pkg.NewFramework(config)
    if err != nil {
        t.Fatalf("Failed to create framework: %v", err)
    }
    
    // Load plugin
    err = framework.PluginManager().LoadPlugin(
        "./my-plugin",
        pkg.PluginConfig{
            Enabled: true,
            Config: map[string]interface{}{
                "api_key": "test-key",
            },
        },
    )
    if err != nil {
        t.Fatalf("Failed to load plugin: %v", err)
    }
    
    // Initialize and start
    framework.PluginManager().InitializeAll()
    framework.PluginManager().StartAll()
    
    // Test plugin functionality
    plugin, err := framework.PluginManager().GetPlugin("my-plugin")
    if err != nil {
        t.Fatalf("Failed to get plugin: %v", err)
    }
    
    if plugin.Name() != "my-plugin" {
        t.Errorf("Expected plugin name 'my-plugin', got '%s'", plugin.Name())
    }
    
    // Cleanup
    framework.PluginManager().StopAll()
}
```

### Mock Plugin Context

Create a mock context for testing:

```go
type MockPluginContext struct {
    config  map[string]interface{}
    storage map[string]interface{}
    logger  pkg.Logger
}

func NewMockPluginContext() *MockPluginContext {
    return &MockPluginContext{
        config:  make(map[string]interface{}),
        storage: make(map[string]interface{}),
        logger:  pkg.NewLogger(),
    }
}

func (m *MockPluginContext) PluginConfig() map[string]interface{} {
    return m.config
}

func (m *MockPluginContext) Logger() pkg.Logger {
    return m.logger
}

// Implement other methods...
```

### Testing Hooks

```go
func TestHookExecution(t *testing.T) {
    plugin := NewPlugin().(*MyPlugin)
    ctx := NewMockPluginContext()
    
    // Initialize plugin (registers hooks)
    plugin.Initialize(ctx)
    
    // Verify hook was registered
    if len(ctx.hooks) == 0 {
        t.Error("No hooks registered")
    }
    
    // Execute hook
    hookCtx := pkg.NewMockHookContext()
    err := ctx.hooks[0].Handler(hookCtx)
    if err != nil {
        t.Errorf("Hook execution failed: %v", err)
    }
}
```

### Testing Events

```go
func TestEventPublishing(t *testing.T) {
    plugin := NewPlugin().(*MyPlugin)
    ctx := NewMockPluginContext()
    
    plugin.Initialize(ctx)
    
    // Publish event
    err := plugin.publishTestEvent()
    if err != nil {
        t.Errorf("Failed to publish event: %v", err)
    }
    
    // Verify event was published
    if len(ctx.publishedEvents) == 0 {
        t.Error("No events published")
    }
}

func TestEventSubscription(t *testing.T) {
    plugin := NewPlugin().(*MyPlugin)
    ctx := NewMockPluginContext()
    
    plugin.Initialize(ctx)
    
    // Verify subscription
    if len(ctx.subscriptions) == 0 {
        t.Error("No event subscriptions")
    }
    
    // Trigger event
    event := pkg.Event{
        Name: "test.event",
        Data: map[string]interface{}{"key": "value"},
    }
    
    err := ctx.subscriptions[0].Handler(event)
    if err != nil {
        t.Errorf("Event handler failed: %v", err)
    }
}
```

## Best Practices

### 1. Error Handling

Always handle errors gracefully:

```go
// ❌ Bad: Ignoring errors
func (p *MyPlugin) badMethod() {
    p.ctx.Database().Exec("INSERT INTO ...")
}

// ✅ Good: Proper error handling
func (p *MyPlugin) goodMethod() error {
    _, err := p.ctx.Database().Exec("INSERT INTO ...")
    if err != nil {
        p.ctx.Logger().Error("Database error: " + err.Error())
        return fmt.Errorf("failed to insert: %w", err)
    }
    return nil
}
```

### 2. Resource Cleanup

Always clean up resources:

```go
func (p *MyPlugin) Start() error {
    // Open connection
    conn, err := net.Dial("tcp", "example.com:80")
    if err != nil {
        return err
    }
    p.conn = conn
    
    return nil
}

func (p *MyPlugin) Stop() error {
    // Close connection
    if p.conn != nil {
        p.conn.Close()
        p.conn = nil
    }
    return nil
}
```

### 3. Thread Safety

Use mutexes for concurrent access:

```go
type MyPlugin struct {
    ctx   pkg.PluginContext
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (p *MyPlugin) Get(key string) interface{} {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.cache[key]
}

func (p *MyPlugin) Set(key string, value interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.cache[key] = value
}
```

### 4. Configuration Validation

Validate configuration early:

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Validate required configuration
    config := ctx.PluginConfig()
    
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    timeout, ok := config["timeout"].(string)
    if ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout format: %w", err)
        }
        if duration < 0 {
            return fmt.Errorf("timeout must be positive")
        }
    }
    
    return nil
}
```

### 5. Logging Best Practices

Use structured logging:

```go
// ❌ Bad: String concatenation
p.ctx.Logger().Info("User " + username + " logged in from " + ip)

// ✅ Good: Structured logging
p.ctx.Logger().WithFields(map[string]interface{}{
    "username": username,
    "ip":       ip,
    "action":   "login",
}).Info("User logged in")
```

### 6. Metrics Recording

Record meaningful metrics:

```go
func (p *MyPlugin) processRequest(req Request) error {
    start := time.Now()
    
    // Process request
    err := p.doProcess(req)
    
    // Record metrics
    duration := time.Since(start)
    p.ctx.Metrics().RecordHistogram("myplugin.request.duration", duration.Seconds())
    
    if err != nil {
        p.ctx.Metrics().IncrementCounter("myplugin.request.errors")
        return err
    }
    
    p.ctx.Metrics().IncrementCounter("myplugin.request.success")
    return nil
}
```

### 7. Graceful Degradation

Handle missing dependencies gracefully:

```go
func (p *MyPlugin) Start() error {
    // Try to import optional service
    service, err := p.ctx.ImportService("cache-plugin", "CacheService")
    if err != nil {
        p.ctx.Logger().Warn("Cache service not available, using in-memory cache")
        p.cache = NewInMemoryCache()
    } else {
        p.cache = service.(CacheService)
    }
    
    return nil
}
```

### 8. Idempotent Operations

Make operations idempotent:

```go
func (p *MyPlugin) Stop() error {
    // Safe to call multiple times
    if p.conn != nil {
        p.conn.Close()
        p.conn = nil
    }
    
    if p.worker != nil {
        close(p.stopChan)
        p.worker = nil
    }
    
    return nil
}
```

### 9. Documentation

Document your plugin thoroughly:

```go
// MyPlugin provides advanced caching functionality with Redis backend.
//
// Features:
// - Automatic cache invalidation
// - Distributed caching
// - Cache warming
//
// Required Permissions:
// - cache: Access to cache manager
// - network: Connect to Redis
//
// Configuration:
// - redis_url: Redis connection URL (required)
// - cache_ttl: Default cache TTL (default: 5m)
//
// Events:
// - Publishes: cache.hit, cache.miss, cache.invalidate
// - Subscribes: data.updated
type MyPlugin struct {
    ctx pkg.PluginContext
}
```

### 10. Version Compatibility

Handle version compatibility:

```go
func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{
        {
            // Require framework 1.x
            FrameworkVersion: ">=1.0.0,<2.0.0",
        },
        {
            // Require auth plugin 1.x or 2.x
            Name:    "auth-plugin",
            Version: ">=1.0.0,<3.0.0",
        },
    }
}
```

## Advanced Topics

### Hot Reload Support

**Requirement**: 7.1-7.6

Plugins can be reloaded without restarting the application:

```go
// Reload plugin
err := framework.PluginManager().ReloadPlugin("my-plugin")
```

**Hot Reload Sequence**:
1. Stop() is called on the old plugin
2. Plugin is unloaded from memory
3. New plugin version is loaded
4. Initialize() is called on the new plugin
5. Start() is called on the new plugin

**Handling Hot Reload in Your Plugin**:

```go
func (p *MyPlugin) Stop() error {
    // Save state before reload
    state := map[string]interface{}{
        "counter":   p.counter,
        "last_sync": p.lastSync,
    }
    p.ctx.PluginStorage().Set("reload_state", state)
    
    // Stop operations
    close(p.stopChan)
    p.wg.Wait()
    
    return nil
}

func (p *MyPlugin) Start() error {
    // Restore state after reload
    if state, err := p.ctx.PluginStorage().Get("reload_state"); err == nil {
        stateMap := state.(map[string]interface{})
        p.counter = stateMap["counter"].(int)
        p.lastSync = stateMap["last_sync"].(time.Time)
    }
    
    // Start operations
    go p.backgroundWorker()
    
    return nil
}
```

### Background Workers

Run background tasks safely:

```go
type MyPlugin struct {
    ctx      pkg.PluginContext
    stopChan chan struct{}
    wg       sync.WaitGroup
}

func (p *MyPlugin) Start() error {
    p.stopChan = make(chan struct{})
    
    // Start background worker
    p.wg.Add(1)
    go p.backgroundWorker()
    
    return nil
}

func (p *MyPlugin) backgroundWorker() {
    defer p.wg.Done()
    
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Do periodic work
            p.doPeriodicWork()
            
        case <-p.stopChan:
            // Stop signal received
            p.ctx.Logger().Info("Background worker stopping")
            return
        }
    }
}

func (p *MyPlugin) Stop() error {
    // Signal workers to stop
    close(p.stopChan)
    
    // Wait for workers to finish
    p.wg.Wait()
    
    return nil
}
```

### Custom Middleware Chains

Create complex middleware chains:

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Authentication middleware (high priority)
    ctx.RegisterMiddleware("auth", p.authMiddleware, 200, []string{"/api/*"})
    
    // Rate limiting middleware (medium priority)
    ctx.RegisterMiddleware("ratelimit", p.rateLimitMiddleware, 150, []string{"/api/*"})
    
    // Logging middleware (low priority)
    ctx.RegisterMiddleware("logging", p.loggingMiddleware, 50, []string{})
    
    return nil
}

func (p *MyPlugin) authMiddleware(c pkg.Context) error {
    token := c.GetHeader("Authorization")
    if !p.validateToken(token) {
        return c.JSON(401, map[string]string{"error": "unauthorized"})
    }
    
    // Set user in context
    c.Set("user_id", p.getUserID(token))
    
    return c.Next()
}

func (p *MyPlugin) rateLimitMiddleware(c pkg.Context) error {
    userID := c.Get("user_id").(int)
    
    if !p.checkRateLimit(userID) {
        return c.JSON(429, map[string]string{"error": "rate limit exceeded"})
    }
    
    return c.Next()
}
```

### Database Migrations

Handle database schema changes:

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Run migrations
    if err := p.runMigrations(); err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    return nil
}

func (p *MyPlugin) runMigrations() error {
    db := p.ctx.Database()
    
    // Check current version
    var version int
    err := db.QueryRow("SELECT version FROM plugin_migrations WHERE plugin = ?", p.Name()).Scan(&version)
    if err != nil {
        version = 0
    }
    
    // Run migrations
    migrations := []func() error{
        p.migration1,
        p.migration2,
        p.migration3,
    }
    
    for i := version; i < len(migrations); i++ {
        if err := migrations[i](); err != nil {
            return fmt.Errorf("migration %d failed: %w", i+1, err)
        }
        
        // Update version
        _, err := db.Exec("INSERT OR REPLACE INTO plugin_migrations (plugin, version) VALUES (?, ?)",
            p.Name(), i+1)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (p *MyPlugin) migration1() error {
    _, err := p.ctx.Database().Exec(`
        CREATE TABLE IF NOT EXISTS my_plugin_data (
            id INTEGER PRIMARY KEY,
            key TEXT NOT NULL,
            value TEXT NOT NULL
        )
    `)
    return err
}
```

### Performance Optimization

Optimize plugin performance:

```go
// Use connection pooling
type MyPlugin struct {
    ctx        pkg.PluginContext
    httpClient *http.Client
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Create HTTP client with connection pooling
    p.httpClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    return nil
}

// Use caching
func (p *MyPlugin) getData(key string) (interface{}, error) {
    // Check cache first
    if cached, err := p.ctx.Cache().Get(key); err == nil {
        return cached, nil
    }
    
    // Cache miss - fetch data
    data, err := p.fetchData(key)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    p.ctx.Cache().Set(key, data, 5*time.Minute)
    
    return data, nil
}

// Use batching
func (p *MyPlugin) processBatch(items []Item) error {
    // Process items in batches
    batchSize := 100
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        if err := p.processBatchItems(batch); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Troubleshooting

### Plugin Not Loading

**Problem**: Plugin fails to load

**Solutions**:

1. **Check manifest validity**:
```bash
# Validate YAML
yamllint plugin.yaml

# Check for required fields
grep -E "name|version|description|author" plugin.yaml
```

2. **Verify dependencies**:
```go
// Check dependency versions
func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{
        {
            Name:    "auth-plugin",
            Version: ">=1.0.0",  // Make sure this matches installed version
        },
    }
}
```

3. **Check framework version**:
```yaml
framework:
  version: ">=1.0.0,<2.0.0"  # Ensure framework version is compatible
```

4. **Review logs**:
```go
// Enable debug logging
framework.SetLogLevel("debug")
```

### Permission Denied Errors

**Problem**: Plugin cannot access framework services

**Solutions**:

1. **Grant required permissions**:
```yaml
# In plugin.yaml
permissions:
  database: true  # Add missing permission
  cache: true
```

2. **Check configuration**:
```yaml
# In framework config
plugins:
  - name: my-plugin
    permissions:
      database: true  # Grant permission
```

3. **Handle permission errors gracefully**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Check if permission is available
    if db := ctx.Database(); db != nil {
        p.db = db
    } else {
        p.ctx.Logger().Warn("Database access not available")
    }
    return nil
}
```

### Hook Not Executing

**Problem**: Registered hooks are not being called

**Solutions**:

1. **Verify hook registration**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Make sure to return error if registration fails
    if err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, p.myHook); err != nil {
        return fmt.Errorf("failed to register hook: %w", err)
    }
    return nil
}
```

2. **Check plugin is started**:
```go
// Hooks only execute after Start() is called
framework.PluginManager().StartAll()
```

3. **Verify hook type**:
```go
// Use correct hook type constant
ctx.RegisterHook(pkg.HookTypePreRequest, 100, handler)  // ✅
ctx.RegisterHook("pre_request", 100, handler)           // ❌
```

### Memory Leaks

**Problem**: Plugin memory usage grows over time

**Solutions**:

1. **Close resources**:
```go
func (p *MyPlugin) Stop() error {
    // Close all connections
    if p.conn != nil {
        p.conn.Close()
    }
    
    // Stop goroutines
    close(p.stopChan)
    p.wg.Wait()
    
    // Clear caches
    p.cache = nil
    
    return nil
}
```

2. **Use context cancellation**:
```go
func (p *MyPlugin) Start() error {
    ctx, cancel := context.WithCancel(context.Background())
    p.cancel = cancel
    
    go p.worker(ctx)
    return nil
}

func (p *MyPlugin) Stop() error {
    p.cancel()  // Cancel all operations
    return nil
}
```

3. **Profile memory usage**:
```go
import _ "net/http/pprof"

// Access profiling at http://localhost:6060/debug/pprof/
go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

### Event Not Received

**Problem**: Subscribed events are not being received

**Solutions**:

1. **Verify subscription**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Subscribe in Initialize, not Start
    return ctx.SubscribeEvent("user.created", p.handleUserCreated)
}
```

2. **Check event name**:
```go
// Publisher
ctx.PublishEvent("user.created", data)

// Subscriber - must match exactly
ctx.SubscribeEvent("user.created", handler)  // ✅
ctx.SubscribeEvent("user_created", handler)  // ❌ Wrong name
```

3. **Handle errors in event handler**:
```go
func (p *MyPlugin) handleEvent(event pkg.Event) error {
    // Log errors for debugging
    if err := p.processEvent(event); err != nil {
        p.ctx.Logger().Error("Event processing failed: " + err.Error())
        return err
    }
    return nil
}
```

## Examples

### Example 1: Authentication Plugin

Complete authentication plugin with JWT tokens:

```go
package main

import (
    "fmt"
    "time"
    "github.com/yourusername/rockstar/pkg"
)

type AuthPlugin struct {
    ctx           pkg.PluginContext
    tokenDuration time.Duration
    excludedPaths []string
}

func (p *AuthPlugin) Name() string        { return "auth-plugin" }
func (p *AuthPlugin) Version() string     { return "1.0.0" }
func (p *AuthPlugin) Description() string { return "JWT authentication plugin" }
func (p *AuthPlugin) Author() string      { return "Your Name" }
func (p *AuthPlugin) Dependencies() []pkg.PluginDependency { return nil }

func (p *AuthPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Load configuration
    config := ctx.PluginConfig()
    if duration, ok := config["token_duration"].(string); ok {
        d, _ := time.ParseDuration(duration)
        p.tokenDuration = d
    }
    
    if paths, ok := config["excluded_paths"].([]interface{}); ok {
        for _, path := range paths {
            p.excludedPaths = append(p.excludedPaths, path.(string))
        }
    }
    
    // Register authentication middleware
    return ctx.RegisterMiddleware("auth", p.authMiddleware, 200, []string{"/api/*"})
}

func (p *AuthPlugin) authMiddleware(c pkg.Context) error {
    // Check if path is excluded
    for _, path := range p.excludedPaths {
        if c.Path() == path {
            return c.Next()
        }
    }
    
    // Validate token
    token := c.GetHeader("Authorization")
    if !p.validateToken(token) {
        return c.JSON(401, map[string]string{"error": "unauthorized"})
    }
    
    // Set user in context
    userID := p.getUserIDFromToken(token)
    c.Set("user_id", userID)
    
    return c.Next()
}

func (p *AuthPlugin) validateToken(token string) bool {
    // Token validation logic
    return token != ""
}

func (p *AuthPlugin) getUserIDFromToken(token string) int {
    // Extract user ID from token
    return 123
}

func (p *AuthPlugin) Start() error {
    p.ctx.Logger().Info("Auth plugin started")
    return nil
}

func (p *AuthPlugin) Stop() error {
    p.ctx.Logger().Info("Auth plugin stopped")
    return nil
}

func (p *AuthPlugin) Cleanup() error { return nil }
func (p *AuthPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "token_duration": map[string]interface{}{
            "type":    "duration",
            "default": "2h",
        },
        "excluded_paths": map[string]interface{}{
            "type":    "array",
            "default": []string{"/health"},
        },
    }
}
func (p *AuthPlugin) OnConfigChange(config map[string]interface{}) error { return nil }

func NewPlugin() pkg.Plugin {
    return &AuthPlugin{}
}
```

### Example 2: Caching Plugin

Response caching plugin:

```go
package main

import (
    "time"
    "github.com/yourusername/rockstar/pkg"
)

type CachePlugin struct {
    ctx           pkg.PluginContext
    cacheDuration time.Duration
    hits          int64
    misses        int64
}

func (p *CachePlugin) Name() string        { return "cache-plugin" }
func (p *CachePlugin) Version() string     { return "1.0.0" }
func (p *CachePlugin) Description() string { return "Response caching plugin" }
func (p *CachePlugin) Author() string      { return "Your Name" }
func (p *CachePlugin) Dependencies() []pkg.PluginDependency { return nil }

func (p *CachePlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Load configuration
    config := ctx.PluginConfig()
    if duration, ok := config["cache_duration"].(string); ok {
        d, _ := time.ParseDuration(duration)
        p.cacheDuration = d
    }
    
    // Register caching middleware
    ctx.RegisterMiddleware("cache", p.cacheMiddleware, 100, []string{})
    
    // Export cache service
    return ctx.ExportService("CacheService", &CacheService{plugin: p})
}

func (p *CachePlugin) cacheMiddleware(c pkg.Context) error {
    // Only cache GET requests
    if c.Method() != "GET" {
        return c.Next()
    }
    
    // Check cache
    cacheKey := "response:" + c.Path()
    if cached, err := p.ctx.Cache().Get(cacheKey); err == nil {
        p.hits++
        return c.JSON(200, cached)
    }
    
    // Cache miss
    p.misses++
    
    // Continue to handler
    if err := c.Next(); err != nil {
        return err
    }
    
    // Cache response
    // (In real implementation, capture response body)
    
    return nil
}

func (p *CachePlugin) Start() error {
    p.ctx.Logger().Info("Cache plugin started")
    return nil
}

func (p *CachePlugin) Stop() error {
    p.ctx.Logger().Info(fmt.Sprintf("Cache stats - Hits: %d, Misses: %d", p.hits, p.misses))
    return nil
}

func (p *CachePlugin) Cleanup() error { return nil }
func (p *CachePlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "cache_duration": map[string]interface{}{
            "type":    "duration",
            "default": "10m",
        },
    }
}
func (p *CachePlugin) OnConfigChange(config map[string]interface{}) error { return nil }

type CacheService struct {
    plugin *CachePlugin
}

func (s *CacheService) GetStats() map[string]int64 {
    return map[string]int64{
        "hits":   s.plugin.hits,
        "misses": s.plugin.misses,
    }
}

func NewPlugin() pkg.Plugin {
    return &CachePlugin{}
}
```

## Additional Resources

- **Example Plugins**: See `examples/plugins/` for complete working examples
- **Plugin System Design**: `.kiro/specs/plugin-system/design.md`
- **Plugin System Requirements**: `.kiro/specs/plugin-system/requirements.md`
- **Configuration Reference**: `examples/PLUGIN_CONFIG_REFERENCE.md`
- **Quick Start Guide**: `examples/plugins/QUICKSTART.md`
- **API Documentation**: `docs/PLUGIN_SYSTEM.md`

## Contributing

To contribute to the plugin system:

1. Read the design document
2. Follow the coding standards
3. Write tests for your changes
4. Update documentation
5. Submit a pull request

## License

This documentation is part of the Rockstar Web Framework and is provided under the same license.

---

**Happy Plugin Development!** 🚀

For questions or support, please open an issue on GitHub or join our community forum.
