# Plugin System API Reference

## Table of Contents

- [Overview](#overview)
- [PluginManager API](#pluginmanager-api)
- [Plugin Interface](#plugin-interface)
- [PluginContext API](#plugincontext-api)
- [Configuration Format](#configuration-format)
- [Hot Reload Process](#hot-reload-process)
- [Monitoring and Metrics](#monitoring-and-metrics)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Best Practices](#best-practices)

## Overview

The Plugin System provides a robust, secure, and performant architecture for extending the Rockstar Web Framework. This document serves as the complete API reference for plugin system integration and management.

### Key Features

- **Dynamic Loading**: Load and unload plugins at runtime without restarting the application
- **Hot Reload**: Update plugins without downtime (Requirement 7.1-7.6)
- **Dependency Management**: Automatic dependency resolution and load ordering (Requirement 5.1-5.5)
- **Security**: Granular permission system for controlling plugin access (Requirement 11.1-11.5)
- **Isolation**: Plugin storage, configuration, and error isolation (Requirement 8.1-8.3, 10.5)
- **Inter-Plugin Communication**: Event bus and service export/import (Requirement 10.1-10.5)
- **Monitoring**: Comprehensive metrics and health tracking (Requirement 9.1-9.5)

### Architecture

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
```


## PluginManager API

The `PluginManager` is the central component for managing plugin lifecycle, dependencies, and health.

### Interface Definition

```go
type PluginManager interface {
    // Loading (Requirement 3.1-3.5)
    LoadPlugin(path string, config PluginConfig) error
    LoadPluginsFromConfig(configPath string) error
    UnloadPlugin(name string) error
    
    // Lifecycle (Requirement 1.1)
    InitializeAll() error
    StartAll() error
    StopAll() error
    
    // Hot reload (Requirement 7.1-7.6)
    ReloadPlugin(name string) error
    
    // Query
    GetPlugin(name string) (Plugin, error)
    ListPlugins() []PluginInfo
    IsLoaded(name string) bool
    
    // Dependency management (Requirement 5.1-5.5)
    ResolveDependencies() error
    GetDependencyGraph() map[string][]string
    
    // Health (Requirement 9.1-9.5)
    GetPluginHealth(name string) PluginHealth
    GetAllHealth() map[string]PluginHealth
}
```

### LoadPlugin

Loads a single plugin from the specified path.

**Signature:**
```go
func (pm *PluginManager) LoadPlugin(path string, config PluginConfig) error
```

**Parameters:**
- `path` (string): Absolute or relative path to the plugin directory or binary
- `config` (PluginConfig): Plugin configuration including permissions and initialization parameters

**Returns:**
- `error`: Error if loading fails, nil on success

**Requirements:** 3.2, 3.3, 3.4

**Example:**
```go
err := pluginManager.LoadPlugin("./plugins/auth-plugin", pkg.PluginConfig{
    Enabled: true,
    Path:    "./plugins/auth-plugin",
    Config: map[string]interface{}{
        "jwt_secret": "my-secret-key",
        "token_ttl":  "24h",
    },
    Permissions: pkg.PluginPermissions{
        AllowDatabase: true,
        AllowCache:    true,
        AllowRouter:   true,
    },
    Priority: 100,
})
if err != nil {
    log.Fatalf("Failed to load plugin: %v", err)
}
```

**Error Conditions:**
- Plugin file not found
- Invalid plugin manifest
- Missing dependencies (Requirement 5.3)
- Framework version mismatch (Requirement 5.1)
- Circular dependencies (Requirement 5.5)
- Plugin interface not implemented (Requirement 1.2)


### LoadPluginsFromConfig

Loads multiple plugins from a configuration file.

**Signature:**
```go
func (pm *PluginManager) LoadPluginsFromConfig(configPath string) error
```

**Parameters:**
- `configPath` (string): Path to configuration file (YAML, JSON, or TOML)

**Returns:**
- `error`: Error if loading fails, nil on success

**Requirements:** 3.1, 3.3, 3.5

**Example:**
```go
// Load from YAML config
err := pluginManager.LoadPluginsFromConfig("./config/plugins.yaml")
if err != nil {
    log.Fatalf("Failed to load plugins: %v", err)
}
```

**Configuration File Example:**
```yaml
plugins:
  enabled: true
  directory: ./plugins
  
  plugins:
    - name: auth-plugin
      enabled: true
      path: ./plugins/auth-plugin
      priority: 200
      config:
        jwt_secret: "secret-key"
      permissions:
        database: true
        cache: true
        router: true
    
    - name: logging-plugin
      enabled: true
      path: ./plugins/logging-plugin
      priority: 100
```

**Load Order:** Plugins are loaded in the order specified in the configuration file, respecting dependencies (Requirement 3.5, 5.2).

### UnloadPlugin

Unloads a plugin and cleans up its resources.

**Signature:**
```go
func (pm *PluginManager) UnloadPlugin(name string) error
```

**Parameters:**
- `name` (string): Plugin name to unload

**Returns:**
- `error`: Error if unloading fails, nil on success

**Requirements:** 1.1, 6.5

**Example:**
```go
err := pluginManager.UnloadPlugin("auth-plugin")
if err != nil {
    log.Printf("Failed to unload plugin: %v", err)
}
```

**Unload Process:**
1. Call plugin's `Stop()` method
2. Remove all registered hooks
3. Unsubscribe from all events
4. Remove all registered middleware (Requirement 6.5)
5. Call plugin's `Cleanup()` method
6. Remove from registry

**Note:** Dependent plugins are NOT automatically unloaded. You must unload them separately or handle dependency checks.


### InitializeAll

Initializes all loaded plugins in dependency order.

**Signature:**
```go
func (pm *PluginManager) InitializeAll() error
```

**Returns:**
- `error`: Error if any plugin initialization fails

**Requirements:** 1.1, 5.2

**Example:**
```go
if err := pluginManager.InitializeAll(); err != nil {
    log.Fatalf("Plugin initialization failed: %v", err)
}
```

**Initialization Order:**
1. Resolve dependencies (Requirement 5.2)
2. Sort plugins by dependency graph
3. Call `Initialize(ctx)` on each plugin in order
4. Register hooks and event subscriptions

**Error Handling:** If any plugin fails to initialize, the process stops and returns an error. Successfully initialized plugins remain initialized.

### StartAll

Starts all initialized plugins.

**Signature:**
```go
func (pm *PluginManager) StartAll() error
```

**Returns:**
- `error`: Error if any plugin fails to start

**Requirements:** 1.1, 2.1

**Example:**
```go
if err := pluginManager.StartAll(); err != nil {
    log.Fatalf("Plugin startup failed: %v", err)
}
```

**Startup Process:**
1. Execute startup hooks (Requirement 2.1)
2. Call `Start()` on each plugin
3. Mark plugins as running
4. Record startup metrics (Requirement 9.1)

### StopAll

Stops all running plugins in reverse dependency order.

**Signature:**
```go
func (pm *PluginManager) StopAll() error
```

**Returns:**
- `error`: Error if any plugin fails to stop (non-fatal)

**Requirements:** 1.1, 2.2

**Example:**
```go
if err := pluginManager.StopAll(); err != nil {
    log.Printf("Some plugins failed to stop: %v", err)
}
```

**Shutdown Process:**
1. Execute shutdown hooks (Requirement 2.2)
2. Call `Stop()` on each plugin in reverse order
3. Wait for graceful shutdown
4. Force stop after timeout


### ReloadPlugin

Hot reloads a plugin without stopping the application.

**Signature:**
```go
func (pm *PluginManager) ReloadPlugin(name string) error
```

**Parameters:**
- `name` (string): Plugin name to reload

**Returns:**
- `error`: Error if reload fails

**Requirements:** 7.1, 7.2, 7.3, 7.4, 7.5, 7.6

**Example:**
```go
err := pluginManager.ReloadPlugin("auth-plugin")
if err != nil {
    log.Printf("Hot reload failed: %v", err)
    // Previous version is restored
}
```

**Reload Process:**
1. Queue incoming requests for the plugin (Requirement 7.6)
2. Call `Stop()` on current version (Requirement 7.1)
3. Unload current version from memory (Requirement 7.2)
4. Load new version from disk (Requirement 7.3)
5. Call `Initialize()` and `Start()` on new version (Requirement 7.4)
6. Process queued requests
7. On failure: Rollback to previous version (Requirement 7.5)

**Rollback Behavior:**
If the new version fails to initialize or start, the system automatically:
1. Unloads the failed version
2. Reloads the previous version
3. Restarts the previous version
4. Processes queued requests

**Request Queuing:**
During reload, requests that would be handled by the plugin are queued with a configurable timeout (default: 30 seconds). If reload takes longer than the timeout, requests fail with a 503 Service Unavailable error.

### GetPlugin

Retrieves a loaded plugin by name.

**Signature:**
```go
func (pm *PluginManager) GetPlugin(name string) (Plugin, error)
```

**Parameters:**
- `name` (string): Plugin name

**Returns:**
- `Plugin`: Plugin instance
- `error`: Error if plugin not found

**Example:**
```go
plugin, err := pluginManager.GetPlugin("auth-plugin")
if err != nil {
    log.Printf("Plugin not found: %v", err)
    return
}
fmt.Printf("Plugin version: %s\n", plugin.Version())
```


### ListPlugins

Returns information about all loaded plugins.

**Signature:**
```go
func (pm *PluginManager) ListPlugins() []PluginInfo
```

**Returns:**
- `[]PluginInfo`: Array of plugin information

**Example:**
```go
plugins := pluginManager.ListPlugins()
for _, info := range plugins {
    fmt.Printf("Plugin: %s v%s - Status: %s\n", 
        info.Name, info.Version, info.Status)
}
```

**PluginInfo Structure:**
```go
type PluginInfo struct {
    Name        string       // Plugin name
    Version     string       // Plugin version
    Description string       // Plugin description
    Author      string       // Plugin author
    Loaded      bool         // Whether plugin is loaded
    Enabled     bool         // Whether plugin is enabled
    LoadTime    time.Time    // When plugin was loaded
    Status      PluginStatus // Current status
}

type PluginStatus string

const (
    PluginStatusUnloaded    PluginStatus = "unloaded"
    PluginStatusLoading     PluginStatus = "loading"
    PluginStatusInitialized PluginStatus = "initialized"
    PluginStatusRunning     PluginStatus = "running"
    PluginStatusStopped     PluginStatus = "stopped"
    PluginStatusError       PluginStatus = "error"
)
```

### IsLoaded

Checks if a plugin is currently loaded.

**Signature:**
```go
func (pm *PluginManager) IsLoaded(name string) bool
```

**Parameters:**
- `name` (string): Plugin name

**Returns:**
- `bool`: true if loaded, false otherwise

**Example:**
```go
if pluginManager.IsLoaded("auth-plugin") {
    fmt.Println("Auth plugin is loaded")
}
```

### ResolveDependencies

Resolves and validates all plugin dependencies.

**Signature:**
```go
func (pm *PluginManager) ResolveDependencies() error
```

**Returns:**
- `error`: Error if dependencies cannot be resolved

**Requirements:** 5.1, 5.2, 5.3, 5.4, 5.5

**Example:**
```go
if err := pluginManager.ResolveDependencies(); err != nil {
    log.Fatalf("Dependency resolution failed: %v", err)
}
```

**Validation Checks:**
1. Framework version compatibility (Requirement 5.1)
2. All required dependencies are present (Requirement 5.3)
3. Dependency versions are compatible (Requirement 5.4)
4. No circular dependencies exist (Requirement 5.5)
5. Correct load order can be determined (Requirement 5.2)

**Error Messages:**
- `"missing dependency: plugin 'X' requires 'Y' version 'Z'"`
- `"version mismatch: plugin 'X' requires 'Y' >=1.0.0, found 0.9.0"`
- `"circular dependency detected: A -> B -> C -> A"`
- `"framework version mismatch: plugin requires >=2.0.0, found 1.5.0"`


### GetDependencyGraph

Returns the plugin dependency graph.

**Signature:**
```go
func (pm *PluginManager) GetDependencyGraph() map[string][]string
```

**Returns:**
- `map[string][]string`: Map of plugin names to their dependencies

**Example:**
```go
graph := pluginManager.GetDependencyGraph()
for plugin, deps := range graph {
    fmt.Printf("%s depends on: %v\n", plugin, deps)
}

// Output:
// auth-plugin depends on: []
// user-plugin depends on: [auth-plugin]
// admin-plugin depends on: [auth-plugin, user-plugin]
```

### GetPluginHealth

Returns health information for a specific plugin.

**Signature:**
```go
func (pm *PluginManager) GetPluginHealth(name string) PluginHealth
```

**Parameters:**
- `name` (string): Plugin name

**Returns:**
- `PluginHealth`: Health information

**Requirements:** 9.1, 9.2, 9.3, 9.4, 9.5

**Example:**
```go
health := pluginManager.GetPluginHealth("auth-plugin")
fmt.Printf("Status: %s\n", health.Status)
fmt.Printf("Error Count: %d\n", health.ErrorCount)
if health.LastError != nil {
    fmt.Printf("Last Error: %v at %s\n", 
        health.LastError, health.LastErrorAt)
}
```

**PluginHealth Structure:**
```go
type PluginHealth struct {
    Status       PluginStatus           // Current status
    ErrorCount   int64                  // Total error count (Requirement 9.3)
    LastError    error                  // Last error encountered
    LastErrorAt  time.Time              // When last error occurred
    HookMetrics  map[string]HookMetrics // Per-hook metrics (Requirement 9.2)
}

type HookMetrics struct {
    ExecutionCount  int64         // Number of executions
    TotalDuration   time.Duration // Total execution time
    AverageDuration time.Duration // Average execution time
    ErrorCount      int64         // Hook-specific errors
}
```

### GetAllHealth

Returns health information for all plugins.

**Signature:**
```go
func (pm *PluginManager) GetAllHealth() map[string]PluginHealth
```

**Returns:**
- `map[string]PluginHealth`: Map of plugin names to health information

**Example:**
```go
allHealth := pluginManager.GetAllHealth()
for name, health := range allHealth {
    if health.ErrorCount > 10 {
        log.Printf("WARNING: Plugin %s has %d errors", 
            name, health.ErrorCount)
    }
}
```


## Plugin Interface

Every plugin must implement the `Plugin` interface.

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
    
    // Configuration (Requirement 8.1, 8.2, 8.5)
    ConfigSchema() map[string]interface{}
    OnConfigChange(config map[string]interface{}) error
}
```

### Lifecycle Methods

#### Initialize(ctx PluginContext) error

Called once when the plugin is loaded. Use this to:
- Store the PluginContext
- Register hooks (Requirement 2.1-2.5)
- Subscribe to events (Requirement 10.2)
- Validate configuration
- Initialize internal state

**Requirements:** 1.1, 1.5

**Example:**
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Validate required configuration
    config := ctx.PluginConfig()
    if _, ok := config["api_key"]; !ok {
        return fmt.Errorf("api_key is required")
    }
    
    // Register pre-request hook
    if err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, 
        p.handlePreRequest); err != nil {
        return err
    }
    
    // Subscribe to events
    if err := ctx.SubscribeEvent("user.created", 
        p.handleUserCreated); err != nil {
        return err
    }
    
    return nil
}
```

**Important:** Do not start background goroutines or open connections in Initialize(). Use Start() for that.

#### Start() error

Called after all plugins are initialized. Use this to:
- Start background goroutines
- Open connections (database, network)
- Begin processing

**Requirements:** 1.1

**Example:**
```go
func (p *MyPlugin) Start() error {
    // Start background worker
    p.stopChan = make(chan struct{})
    p.wg.Add(1)
    go p.backgroundWorker()
    
    // Open external connection
    conn, err := grpc.Dial(p.serviceAddr)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    p.conn = conn
    
    p.ctx.Logger().Info("Plugin started")
    return nil
}
```


#### Stop() error

Called when the plugin is being stopped (shutdown or hot reload). Use this to:
- Stop background goroutines
- Close connections
- Flush buffers
- Save state

**Requirements:** 1.1, 7.1

**Example:**
```go
func (p *MyPlugin) Stop() error {
    // Signal workers to stop
    close(p.stopChan)
    
    // Wait for workers to finish
    p.wg.Wait()
    
    // Close connections
    if p.conn != nil {
        p.conn.Close()
    }
    
    // Flush buffers
    p.flushBuffers()
    
    p.ctx.Logger().Info("Plugin stopped")
    return nil
}
```

**Important:** Stop() should be idempotent and safe to call multiple times.

#### Cleanup() error

Called when the plugin is being unloaded. Use this for final cleanup:
- Release resources
- Delete temporary files
- Clear caches

**Requirements:** 1.1

**Example:**
```go
func (p *MyPlugin) Cleanup() error {
    // Remove temporary files
    if err := os.RemoveAll(p.tempDir); err != nil {
        p.ctx.Logger().Error("Failed to remove temp dir: " + err.Error())
    }
    
    // Clear caches
    p.cache = nil
    
    p.ctx.Logger().Info("Plugin cleaned up")
    return nil
}
```

### Configuration Methods

#### ConfigSchema() map[string]interface{}

Defines the configuration schema with default values.

**Requirements:** 8.5

**Example:**
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
            "type":    "integer",
            "default": 3,
        },
    }
}
```

#### OnConfigChange(config map[string]interface{}) error

Called when plugin configuration is updated at runtime.

**Requirements:** 8.2

**Example:**
```go
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    if timeout, ok := config["timeout"].(string); ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout: %w", err)
        }
        p.timeout = duration
    }
    
    p.ctx.Logger().Info("Configuration updated")
    return nil
}
```


## PluginContext API

The `PluginContext` provides isolated access to framework services.

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

Access the framework router to register routes.

**Requirements:** 4.1
**Permission Required:** `router: true`

**Example:**
```go
router := ctx.Router()
router.GET("/api/myplugin/status", func(c pkg.Context) error {
    return c.JSON(200, map[string]string{"status": "ok"})
})
```

#### Logger() Logger

Access the framework logger.

**Requirements:** 4.5
**Permission Required:** None (always available)

**Example:**
```go
logger := ctx.Logger()
logger.Info("Plugin operation completed")
logger.WithFields(map[string]interface{}{
    "user_id": 123,
    "action":  "login",
}).Info("User logged in")
```

#### Metrics() MetricsCollector

Access the metrics collector.

**Requirements:** 4.6
**Permission Required:** None (always available)

**Example:**
```go
metrics := ctx.Metrics()
metrics.IncrementCounter("myplugin.requests.total")
metrics.RecordGauge("myplugin.queue.size", float64(queueSize))
metrics.RecordHistogram("myplugin.request.duration", duration.Seconds())
```

#### Database() DatabaseManager

Access the database manager.

**Requirements:** 4.2
**Permission Required:** `database: true`

**Example:**
```go
db := ctx.Database()
result, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
if err != nil {
    return err
}
```

#### Cache() CacheManager

Access the cache manager.

**Requirements:** 4.3
**Permission Required:** `cache: true`

**Example:**
```go
cache := ctx.Cache()
value, found := cache.Get("user:123")
if !found {
    value = fetchUser(123)
    cache.Set("user:123", value, 5*time.Minute)
}
```

#### Config() ConfigManager

Access the framework configuration.

**Requirements:** 4.4
**Permission Required:** `config: true`

**Example:**
```go
config := ctx.Config()
dbHost := config.GetString("database.host")
```


### Plugin-Specific Methods

#### PluginConfig() map[string]interface{}

Get plugin-specific configuration.

**Requirements:** 8.1

**Example:**
```go
config := ctx.PluginConfig()
apiKey := config["api_key"].(string)
timeout := config["timeout"].(string)
```

**Configuration Isolation:** Each plugin only sees its own configuration (Requirement 8.1).

#### PluginStorage() PluginStorage

Access isolated plugin storage.

**Requirements:** 8.3

**Storage Interface:**
```go
type PluginStorage interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    List() ([]string, error)
    Clear() error
}
```

**Example:**
```go
storage := ctx.PluginStorage()

// Save data
storage.Set("last_sync", time.Now())

// Load data
lastSync, err := storage.Get("last_sync")
if err != nil {
    return err
}
```

**Storage Isolation:** Each plugin has isolated storage that other plugins cannot access (Requirement 8.3).

### Hook Registration

#### RegisterHook

Register a hook at a framework lifecycle point.

**Signature:**
```go
func (ctx PluginContext) RegisterHook(hookType HookType, priority int, handler HookHandler) error
```

**Parameters:**
- `hookType` (HookType): Type of hook (startup, shutdown, pre_request, etc.)
- `priority` (int): Execution priority (higher = earlier)
- `handler` (HookHandler): Hook handler function

**Requirements:** 2.1, 2.2, 2.3, 2.4, 2.5, 2.6

**Hook Types:**
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

**Example:**
```go
// Register pre-request hook
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    reqCtx := hookCtx.Context()
    ctx.Logger().Info("Processing request: " + reqCtx.Path())
    return nil
})
```

**Priority Ranges:**
- 300-1000: Critical (authentication, security)
- 200-299: High (rate limiting, validation)
- 100-199: Normal (default)
- 50-99: Low (logging, metrics)
- 0-49: Lowest (cleanup)

**Hook Execution Order:** Hooks execute in descending priority order (Requirement 2.6). Errors in one hook don't prevent other hooks from executing (Requirement 2.7).


### Event System

#### PublishEvent

Publish an event to all subscribers.

**Signature:**
```go
func (ctx PluginContext) PublishEvent(event string, data interface{}) error
```

**Parameters:**
- `event` (string): Event name (use dot notation: "user.created")
- `data` (interface{}): Event payload

**Requirements:** 10.1

**Example:**
```go
err := ctx.PublishEvent("user.created", map[string]interface{}{
    "user_id":   123,
    "username":  "john",
    "timestamp": time.Now(),
})
```

**Event Delivery:** Events are delivered asynchronously to all subscribers (Requirement 10.1).

#### SubscribeEvent

Subscribe to an event.

**Signature:**
```go
func (ctx PluginContext) SubscribeEvent(event string, handler EventHandler) error
```

**Parameters:**
- `event` (string): Event name to subscribe to
- `handler` (EventHandler): Event handler function

**Requirements:** 10.2

**Example:**
```go
ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    data := event.Data.(map[string]interface{})
    userID := data["user_id"].(int)
    
    // Handle event
    return sendWelcomeEmail(userID)
})
```

**Error Isolation:** Errors in event handlers don't crash the publishing plugin or other subscribers (Requirement 10.5).

### Inter-Plugin Communication

#### ExportService

Export a service for other plugins to use.

**Signature:**
```go
func (ctx PluginContext) ExportService(name string, service interface{}) error
```

**Parameters:**
- `name` (string): Service name
- `service` (interface{}): Service implementation

**Requirements:** 10.4

**Example:**
```go
type UserService interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}

// Export service
ctx.ExportService("UserService", &myUserService)
```

#### ImportService

Import a service from another plugin.

**Signature:**
```go
func (ctx PluginContext) ImportService(pluginName, serviceName string) (interface{}, error)
```

**Parameters:**
- `pluginName` (string): Plugin that exports the service
- `serviceName` (string): Service name

**Returns:**
- `interface{}`: Service implementation
- `error`: Error if service not found

**Requirements:** 10.3, 10.4

**Example:**
```go
// Import service
svc, err := ctx.ImportService("auth-plugin", "UserService")
if err != nil {
    return err
}

userService := svc.(UserService)
user, err := userService.GetUser(123)
```

**Service Discovery:** All exported services are discoverable by other plugins (Requirement 10.4). Calls are routed through the Plugin Registry (Requirement 10.3).


### Middleware Registration

#### RegisterMiddleware

Register middleware for request processing.

**Signature:**
```go
func (ctx PluginContext) RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error
```

**Parameters:**
- `name` (string): Middleware name (unique per plugin)
- `handler` (MiddlewareFunc): Middleware function
- `priority` (int): Execution priority
- `routes` ([]string): Routes to apply middleware to (empty = global)

**Requirements:** 6.1, 6.2, 6.3

**Example:**
```go
// Global middleware
ctx.RegisterMiddleware("auth", authMiddleware, 200, nil)

// Route-specific middleware
ctx.RegisterMiddleware("admin-auth", adminAuthMiddleware, 200, []string{
    "/admin/*",
    "/api/admin/*",
})
```

**Middleware Function Signature:**
```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

**Example Middleware:**
```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return ctx.JSON(401, map[string]string{"error": "unauthorized"})
    }
    
    // Validate token
    user, err := validateToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "invalid token"})
    }
    
    // Set user in context
    ctx.Set("user", user)
    
    // Call next handler
    return next(ctx)
}
```

**Middleware Cleanup:** When a plugin is unloaded, all its middleware is automatically removed (Requirement 6.5).

#### UnregisterMiddleware

Remove registered middleware.

**Signature:**
```go
func (ctx PluginContext) UnregisterMiddleware(name string) error
```

**Parameters:**
- `name` (string): Middleware name to remove

**Example:**
```go
ctx.UnregisterMiddleware("auth")
```


## Configuration Format

Plugins can be configured using YAML, JSON, or TOML formats.

### Framework Configuration

**Requirements:** 3.1, 3.3, 3.4, 3.5

#### YAML Format

```yaml
plugins:
  enabled: true
  directory: ./plugins
  
  plugins:
    - name: auth-plugin
      enabled: true
      path: ./plugins/auth-plugin
      priority: 200
      config:
        jwt_secret: "my-secret-key"
        token_ttl: "24h"
        refresh_enabled: true
      permissions:
        database: true
        cache: true
        router: true
        config: true
        filesystem: false
        network: false
        exec: false
    
    - name: logging-plugin
      enabled: true
      path: ./plugins/logging-plugin
      priority: 100
      config:
        log_level: "info"
        output: "stdout"
      permissions:
        database: false
        cache: false
        router: false
        config: true
        filesystem: true
        network: false
        exec: false
```

#### JSON Format

```json
{
  "plugins": {
    "enabled": true,
    "directory": "./plugins",
    "plugins": [
      {
        "name": "auth-plugin",
        "enabled": true,
        "path": "./plugins/auth-plugin",
        "priority": 200,
        "config": {
          "jwt_secret": "my-secret-key",
          "token_ttl": "24h"
        },
        "permissions": {
          "database": true,
          "cache": true,
          "router": true
        }
      }
    ]
  }
}
```

#### TOML Format

```toml
[plugins]
enabled = true
directory = "./plugins"

[[plugins.plugins]]
name = "auth-plugin"
enabled = true
path = "./plugins/auth-plugin"
priority = 200

[plugins.plugins.config]
jwt_secret = "my-secret-key"
token_ttl = "24h"

[plugins.plugins.permissions]
database = true
cache = true
router = true
```

### Configuration Fields

#### Top-Level Fields

- `enabled` (bool): Whether plugin system is enabled
- `directory` (string): Default directory for plugins
- `plugins` (array): Array of plugin configurations

#### Plugin Configuration Fields

- `name` (string): Plugin name (must match plugin.yaml)
- `enabled` (bool): Whether plugin is enabled (Requirement 3.3)
- `path` (string): Path to plugin (absolute or relative) (Requirement 3.2)
- `priority` (int): Load priority (higher = loaded first)
- `config` (object): Plugin-specific configuration (Requirement 3.4)
- `permissions` (object): Plugin permissions (Requirement 11.1)


### Plugin Manifest Format

**Requirements:** 12.1, 12.2, 12.3, 12.4

Each plugin must have a `plugin.yaml` or `plugin.json` manifest file.

#### YAML Manifest

```yaml
name: auth-plugin
version: 1.2.0
description: Authentication and authorization plugin
author: John Doe <john@example.com>

# Framework compatibility (Requirement 5.1)
framework:
  version: ">=1.0.0,<2.0.0"

# Dependencies (Requirement 5.2-5.5)
dependencies:
  - name: cache-plugin
    version: ">=2.0.0"
    optional: false
  - name: logging-plugin
    version: ">=1.0.0"
    optional: true

# Permissions (Requirement 11.1)
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: false
  network: false
  exec: false

# Configuration schema (Requirement 8.5)
config:
  jwt_secret:
    type: string
    required: true
    description: JWT signing secret
  token_ttl:
    type: duration
    default: "24h"
    description: Token time-to-live
  refresh_enabled:
    type: boolean
    default: true
    description: Enable refresh tokens

# Hooks (Requirement 2.1-2.5)
hooks:
  - type: startup
    priority: 200
  - type: pre_request
    priority: 200
  - type: shutdown
    priority: 100

# Events (Requirement 10.1, 10.2)
events:
  publishes:
    - user.login
    - user.logout
    - token.expired
  subscribes:
    - user.created
    - user.deleted

# Exported services (Requirement 10.4)
exports:
  - name: AuthService
    description: Authentication service
  - name: TokenService
    description: Token management service
```

#### JSON Manifest

```json
{
  "name": "auth-plugin",
  "version": "1.2.0",
  "description": "Authentication and authorization plugin",
  "author": "John Doe <john@example.com>",
  "framework": {
    "version": ">=1.0.0,<2.0.0"
  },
  "dependencies": [
    {
      "name": "cache-plugin",
      "version": ">=2.0.0",
      "optional": false
    }
  ],
  "permissions": {
    "database": true,
    "cache": true,
    "router": true
  },
  "config": {
    "jwt_secret": {
      "type": "string",
      "required": true,
      "description": "JWT signing secret"
    }
  }
}
```

### Manifest Validation

**Requirements:** 12.5

The system validates manifests and rejects invalid ones with descriptive errors:

**Required Fields:**
- `name`: Plugin name
- `version`: Semantic version
- `description`: Plugin description
- `author`: Author information

**Validation Errors:**
- `"missing required field: name"`
- `"invalid version format: must be semantic version (X.Y.Z)"`
- `"invalid dependency version constraint: >=1.0.0,<2.0.0"`
- `"unknown permission: custom_permission"`
- `"invalid config schema: type must be one of [string, integer, float, boolean, duration, array, object]"`


## Hot Reload Process

Hot reload allows updating plugins without stopping the application.

**Requirements:** 7.1, 7.2, 7.3, 7.4, 7.5, 7.6

### Reload Sequence

```
1. Trigger Reload
   ↓
2. Queue Incoming Requests (Requirement 7.6)
   ↓
3. Stop Current Version (Requirement 7.1)
   ↓
4. Unload from Memory (Requirement 7.2)
   ↓
5. Load New Version (Requirement 7.3)
   ↓
6. Initialize & Start (Requirement 7.4)
   ↓
7. Process Queued Requests
   ↓
8. Success or Rollback (Requirement 7.5)
```

### Triggering Hot Reload

```go
// Reload a specific plugin
err := pluginManager.ReloadPlugin("auth-plugin")
if err != nil {
    log.Printf("Reload failed: %v", err)
    // Previous version is automatically restored
}
```

### Request Queuing

During reload, requests that would be handled by the plugin are queued:

**Queue Configuration:**
```go
type ReloadConfig struct {
    QueueTimeout    time.Duration // Default: 30s
    MaxQueueSize    int          // Default: 1000
    QueueStrategy   string       // "block" or "reject"
}
```

**Queue Strategies:**
- `block`: Queue requests until reload completes or timeout
- `reject`: Immediately reject requests with 503 error

**Example:**
```go
config := pkg.ReloadConfig{
    QueueTimeout:  60 * time.Second,
    MaxQueueSize:  5000,
    QueueStrategy: "block",
}
pluginManager.ReloadPluginWithConfig("auth-plugin", config)
```

### Rollback Mechanism

If reload fails, the system automatically rolls back:

**Rollback Triggers:**
- New version fails to load
- New version fails to initialize
- New version fails to start
- New version crashes within grace period (default: 30s)

**Rollback Process:**
```
1. Detect Failure
   ↓
2. Unload Failed Version
   ↓
3. Reload Previous Version
   ↓
4. Initialize & Start Previous Version
   ↓
5. Process Queued Requests
   ↓
6. Log Rollback Event
```

**Example Rollback Log:**
```
[ERROR] Plugin reload failed: auth-plugin v1.3.0: initialization error
[INFO] Rolling back to previous version: auth-plugin v1.2.0
[INFO] Previous version restored successfully
[INFO] Processing 47 queued requests
```

### Reload Best Practices

1. **Test Before Reload**: Test new plugin versions in staging first
2. **Monitor Metrics**: Watch error rates and latency during reload
3. **Gradual Rollout**: Reload on one server at a time in production
4. **Backup State**: Ensure plugin state is persisted before reload
5. **Version Compatibility**: Ensure new version is backward compatible with stored data

### Reload Limitations

- **Memory**: Old version memory may not be immediately freed (Go GC)
- **Connections**: Long-lived connections may need manual handling
- **State**: In-memory state is lost (use PluginStorage for persistence)
- **Dependencies**: Dependent plugins are not automatically reloaded


## Monitoring and Metrics

The plugin system provides comprehensive monitoring and metrics.

**Requirements:** 9.1, 9.2, 9.3, 9.4, 9.5

### Plugin Metrics

#### Load Metrics

**Requirement:** 9.1

Recorded when a plugin is loaded:

```go
// Metrics recorded
plugin.load.duration_ms      // Load time in milliseconds
plugin.load.success          // 1 for success, 0 for failure
plugin.load.timestamp        // Unix timestamp
```

**Example Query:**
```go
health := pluginManager.GetPluginHealth("auth-plugin")
fmt.Printf("Load time: %v\n", health.LoadTime)
```

#### Hook Execution Metrics

**Requirement:** 9.2

Recorded for each hook execution:

```go
// Metrics per hook type
plugin.hook.execution_count      // Number of executions
plugin.hook.duration_ms          // Execution duration
plugin.hook.error_count          // Number of errors
plugin.hook.average_duration_ms  // Average duration
```

**Example:**
```go
health := pluginManager.GetPluginHealth("auth-plugin")
for hookType, metrics := range health.HookMetrics {
    fmt.Printf("Hook %s:\n", hookType)
    fmt.Printf("  Executions: %d\n", metrics.ExecutionCount)
    fmt.Printf("  Avg Duration: %v\n", metrics.AverageDuration)
    fmt.Printf("  Errors: %d\n", metrics.ErrorCount)
}
```

#### Error Metrics

**Requirement:** 9.3

Tracked for each plugin:

```go
// Error metrics
plugin.error.count           // Total error count
plugin.error.last_timestamp  // Last error timestamp
plugin.error.last_message    // Last error message
```

**Example:**
```go
health := pluginManager.GetPluginHealth("auth-plugin")
if health.ErrorCount > 0 {
    fmt.Printf("Errors: %d\n", health.ErrorCount)
    fmt.Printf("Last error: %v at %s\n", 
        health.LastError, health.LastErrorAt)
}
```

### Metrics Exposure

**Requirement:** 9.4

Plugin metrics are exposed through the framework's MetricsCollector:

```go
// Access metrics in handlers
func metricsHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Get plugin-specific metrics
    pluginMetrics := metrics.GetPluginMetrics("auth-plugin")
    
    return ctx.JSON(200, pluginMetrics)
}
```

**Prometheus Format:**
```
# Plugin load metrics
rockstar_plugin_load_duration_ms{plugin="auth-plugin"} 45.2
rockstar_plugin_load_success{plugin="auth-plugin"} 1

# Hook execution metrics
rockstar_plugin_hook_execution_count{plugin="auth-plugin",hook="pre_request"} 1523
rockstar_plugin_hook_duration_ms{plugin="auth-plugin",hook="pre_request"} 2.3
rockstar_plugin_hook_error_count{plugin="auth-plugin",hook="pre_request"} 0

# Error metrics
rockstar_plugin_error_count{plugin="auth-plugin"} 5
```

### Error Threshold Handling

**Requirement:** 9.5

Plugins can be automatically disabled when error thresholds are exceeded:

**Configuration:**
```yaml
plugins:
  error_threshold:
    enabled: true
    max_errors: 100           # Maximum errors before action
    time_window: "5m"         # Time window for counting errors
    action: "disable"         # Action: "disable", "log", or "alert"
    reset_on_success: true    # Reset counter on successful operation
```

**Example:**
```go
// Configure error threshold
pluginManager.SetErrorThreshold("auth-plugin", pkg.ErrorThreshold{
    MaxErrors:       100,
    TimeWindow:      5 * time.Minute,
    Action:          "disable",
    ResetOnSuccess:  true,
})
```

**Actions:**
- `log`: Log warning message
- `disable`: Automatically disable plugin
- `alert`: Send alert to monitoring system

**Threshold Events:**
```
[WARN] Plugin auth-plugin approaching error threshold: 95/100 errors in 5m
[ERROR] Plugin auth-plugin exceeded error threshold: 100/100 errors in 5m
[INFO] Plugin auth-plugin automatically disabled due to error threshold
```


### Health Checks

Monitor plugin health in real-time:

```go
// Check single plugin health
health := pluginManager.GetPluginHealth("auth-plugin")

if health.Status != pkg.PluginStatusRunning {
    log.Printf("Plugin not running: %s", health.Status)
}

if health.ErrorCount > 10 {
    log.Printf("High error count: %d", health.ErrorCount)
}

// Check all plugins
allHealth := pluginManager.GetAllHealth()
for name, health := range allHealth {
    if health.Status == pkg.PluginStatusError {
        log.Printf("Plugin %s in error state: %v", name, health.LastError)
    }
}
```

### Monitoring Dashboard

Example monitoring dashboard queries:

**Plugin Status:**
```sql
SELECT name, status, error_count, last_error_at
FROM plugins
WHERE status != 'running'
ORDER BY error_count DESC
```

**Hook Performance:**
```sql
SELECT plugin_name, hook_type, 
       execution_count, 
       total_duration_ms / execution_count as avg_duration_ms,
       error_count
FROM plugin_hooks
WHERE execution_count > 0
ORDER BY avg_duration_ms DESC
```

**Error Trends:**
```sql
SELECT plugin_name, 
       COUNT(*) as error_count,
       MAX(recorded_at) as last_error
FROM plugin_metrics
WHERE metric_name = 'error'
  AND recorded_at > NOW() - INTERVAL '1 hour'
GROUP BY plugin_name
ORDER BY error_count DESC
```

### Alerting Rules

Example alerting rules:

**High Error Rate:**
```yaml
alert: PluginHighErrorRate
expr: rate(rockstar_plugin_error_count[5m]) > 10
labels:
  severity: warning
annotations:
  summary: "Plugin {{ $labels.plugin }} has high error rate"
```

**Slow Hook Execution:**
```yaml
alert: PluginSlowHook
expr: rockstar_plugin_hook_duration_ms > 100
labels:
  severity: warning
annotations:
  summary: "Plugin {{ $labels.plugin }} hook {{ $labels.hook }} is slow"
```

**Plugin Down:**
```yaml
alert: PluginDown
expr: rockstar_plugin_status{status="error"} == 1
labels:
  severity: critical
annotations:
  summary: "Plugin {{ $labels.plugin }} is down"
```


## Troubleshooting Guide

Common issues and solutions for the plugin system.

### Plugin Load Failures

#### Issue: Plugin file not found

**Error:**
```
Error: failed to load plugin: plugin file not found: ./plugins/auth-plugin
```

**Solutions:**
1. Verify the path is correct (absolute or relative to application directory)
2. Check file permissions
3. Ensure plugin binary exists and is executable

**Debug:**
```go
// Check if file exists
if _, err := os.Stat("./plugins/auth-plugin"); os.IsNotExist(err) {
    log.Printf("Plugin file does not exist")
}
```

#### Issue: Invalid plugin manifest

**Error:**
```
Error: failed to parse manifest: missing required field: name
```

**Solutions:**
1. Validate manifest YAML/JSON syntax
2. Ensure all required fields are present (name, version, description, author)
3. Check for typos in field names

**Validation:**
```bash
# Validate YAML syntax
yamllint plugin.yaml

# Validate JSON syntax
jq . plugin.json
```

#### Issue: Plugin interface not implemented

**Error:**
```
Error: plugin does not implement required interface: missing method Initialize
```

**Solutions:**
1. Ensure plugin implements all Plugin interface methods
2. Check method signatures match exactly
3. Verify plugin exports NewPlugin() function

**Required Methods:**
```go
Name() string
Version() string
Description() string
Author() string
Dependencies() []PluginDependency
Initialize(ctx PluginContext) error
Start() error
Stop() error
Cleanup() error
ConfigSchema() map[string]interface{}
OnConfigChange(config map[string]interface{}) error
```

### Dependency Issues

#### Issue: Missing dependency

**Error:**
```
Error: missing dependency: plugin 'auth-plugin' requires 'cache-plugin' version '>=2.0.0'
```

**Solutions:**
1. Install the required dependency plugin
2. Mark dependency as optional if not critical
3. Update plugin to remove dependency

**Check Dependencies:**
```go
graph := pluginManager.GetDependencyGraph()
for plugin, deps := range graph {
    fmt.Printf("%s depends on: %v\n", plugin, deps)
}
```

#### Issue: Version mismatch

**Error:**
```
Error: version mismatch: plugin 'auth-plugin' requires 'cache-plugin' >=2.0.0, found 1.5.0
```

**Solutions:**
1. Update dependency plugin to compatible version
2. Downgrade dependent plugin to compatible version
3. Adjust version constraint in manifest

**Version Constraints:**
```yaml
dependencies:
  - name: cache-plugin
    version: ">=1.5.0,<3.0.0"  # More flexible constraint
```

#### Issue: Circular dependency

**Error:**
```
Error: circular dependency detected: plugin-a -> plugin-b -> plugin-c -> plugin-a
```

**Solutions:**
1. Refactor plugins to remove circular dependency
2. Extract shared functionality into a separate plugin
3. Use event system instead of direct dependencies

**Detect Cycles:**
```go
if err := pluginManager.ResolveDependencies(); err != nil {
    if strings.Contains(err.Error(), "circular") {
        log.Printf("Circular dependency detected")
        // Analyze dependency graph
        graph := pluginManager.GetDependencyGraph()
        // ... find cycle
    }
}
```


### Permission Issues

#### Issue: Permission denied

**Error:**
```
Error: permission denied: plugin 'logging-plugin' attempted database access without permission
```

**Solutions:**
1. Add required permission to plugin configuration
2. Update plugin manifest to declare permission
3. Redesign plugin to not require the permission

**Grant Permission:**
```yaml
plugins:
  - name: logging-plugin
    permissions:
      database: true  # Add this permission
```

**Security Note:** Only grant permissions that are absolutely necessary (principle of least privilege).

#### Issue: Permission not declared in manifest

**Error:**
```
Error: permission mismatch: plugin requests 'database' permission but manifest does not declare it
```

**Solutions:**
1. Add permission to plugin.yaml manifest
2. Remove permission request from configuration

**Update Manifest:**
```yaml
permissions:
  database: true
  cache: true
  router: true
```

### Runtime Issues

#### Issue: Plugin crashes during initialization

**Error:**
```
Error: plugin initialization failed: panic: runtime error: invalid memory address
```

**Solutions:**
1. Check for nil pointer dereferences
2. Validate configuration before use
3. Add error handling for external dependencies

**Safe Initialization:**
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Validate configuration
    config := ctx.PluginConfig()
    if config == nil {
        return fmt.Errorf("configuration is nil")
    }
    
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    // Safe initialization
    p.client = &http.Client{Timeout: 30 * time.Second}
    
    return nil
}
```

#### Issue: Hook execution timeout

**Error:**
```
Error: hook execution timeout: pre_request hook exceeded 5s timeout
```

**Solutions:**
1. Optimize hook logic to execute faster
2. Move slow operations to background goroutines
3. Increase hook timeout (if appropriate)

**Async Processing:**
```go
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    // Quick validation
    if !isValid(hookCtx.Context()) {
        return fmt.Errorf("invalid request")
    }
    
    // Slow operation in background
    go func() {
        processRequest(hookCtx.Context())
    }()
    
    return nil
})
```

#### Issue: Memory leak

**Symptoms:**
- Memory usage grows over time
- Plugin performance degrades
- Out of memory errors

**Solutions:**
1. Ensure goroutines are properly stopped in Stop()
2. Close connections and release resources in Cleanup()
3. Use context cancellation for background tasks

**Proper Cleanup:**
```go
type MyPlugin struct {
    ctx      pkg.PluginContext
    stopChan chan struct{}
    wg       sync.WaitGroup
    conn     *grpc.ClientConn
}

func (p *MyPlugin) Start() error {
    p.stopChan = make(chan struct{})
    
    // Start background worker
    p.wg.Add(1)
    go func() {
        defer p.wg.Done()
        for {
            select {
            case <-p.stopChan:
                return
            case <-time.After(1 * time.Second):
                p.doWork()
            }
        }
    }()
    
    return nil
}

func (p *MyPlugin) Stop() error {
    // Signal workers to stop
    close(p.stopChan)
    
    // Wait for workers to finish
    p.wg.Wait()
    
    // Close connections
    if p.conn != nil {
        p.conn.Close()
    }
    
    return nil
}
```


### Hot Reload Issues

#### Issue: Reload fails with rollback

**Error:**
```
Error: plugin reload failed: new version initialization error
Info: rolling back to previous version
```

**Solutions:**
1. Test new version in development first
2. Check logs for specific initialization error
3. Verify configuration compatibility
4. Ensure database schema is compatible

**Safe Reload Process:**
```bash
# 1. Test in development
go test ./plugins/auth-plugin/...

# 2. Build new version
go build -o auth-plugin-v1.3.0 ./plugins/auth-plugin

# 3. Backup current version
cp plugins/auth-plugin plugins/auth-plugin.backup

# 4. Deploy new version
cp auth-plugin-v1.3.0 plugins/auth-plugin

# 5. Trigger reload
curl -X POST http://localhost:8080/admin/plugins/auth-plugin/reload
```

#### Issue: Requests timeout during reload

**Error:**
```
Error: request timeout: plugin reload exceeded queue timeout (30s)
```

**Solutions:**
1. Increase queue timeout for slow-loading plugins
2. Optimize plugin initialization
3. Use gradual rollout (reload one server at a time)

**Increase Timeout:**
```go
config := pkg.ReloadConfig{
    QueueTimeout: 60 * time.Second,  // Increase timeout
    MaxQueueSize: 5000,
}
pluginManager.ReloadPluginWithConfig("auth-plugin", config)
```

### Configuration Issues

#### Issue: Configuration not updating

**Error:**
```
Plugin configuration changed but OnConfigChange not called
```

**Solutions:**
1. Ensure OnConfigChange is implemented
2. Verify configuration file is being reloaded
3. Check for configuration parsing errors

**Implement OnConfigChange:**
```go
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    p.ctx.Logger().Info("Configuration changed")
    
    // Update internal state
    if timeout, ok := config["timeout"].(string); ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout: %w", err)
        }
        p.timeout = duration
    }
    
    return nil
}
```

#### Issue: Default values not applied

**Error:**
```
Configuration key missing and no default value provided
```

**Solutions:**
1. Define defaults in ConfigSchema()
2. Check for missing keys before use
3. Provide fallback values

**Define Defaults:**
```go
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "timeout": map[string]interface{}{
            "type":    "duration",
            "default": "30s",  // Default value
        },
        "max_retries": map[string]interface{}{
            "type":    "integer",
            "default": 3,  // Default value
        },
    }
}
```

### Inter-Plugin Communication Issues

#### Issue: Event not delivered

**Symptoms:**
- Event published but subscribers not called
- No errors logged

**Solutions:**
1. Verify subscription is registered
2. Check event name matches exactly (case-sensitive)
3. Ensure subscriber plugin is running

**Debug Events:**
```go
// Publisher
p.ctx.Logger().Info("Publishing event: user.created")
err := p.ctx.PublishEvent("user.created", data)
if err != nil {
    p.ctx.Logger().Error("Failed to publish: " + err.Error())
}

// Subscriber
p.ctx.Logger().Info("Subscribing to: user.created")
err := p.ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    p.ctx.Logger().Info("Event received: " + event.Name)
    return nil
})
```

#### Issue: Service not found

**Error:**
```
Error: service not found: plugin 'auth-plugin' does not export 'UserService'
```

**Solutions:**
1. Verify service is exported with correct name
2. Ensure exporting plugin is loaded and running
3. Check for typos in service name

**Export Service:**
```go
// In auth-plugin Initialize()
err := ctx.ExportService("UserService", &myUserService)
if err != nil {
    return fmt.Errorf("failed to export service: %w", err)
}
```

**Import Service:**
```go
// In another plugin
svc, err := ctx.ImportService("auth-plugin", "UserService")
if err != nil {
    return fmt.Errorf("failed to import service: %w", err)
}
```


### Debugging Tips

#### Enable Debug Logging

```go
// In plugin
p.ctx.Logger().SetLevel("debug")
p.ctx.Logger().Debug("Detailed debug information")
```

#### Inspect Plugin State

```go
// Get plugin info
plugin, err := pluginManager.GetPlugin("auth-plugin")
if err == nil {
    fmt.Printf("Name: %s\n", plugin.Name())
    fmt.Printf("Version: %s\n", plugin.Version())
}

// Get health
health := pluginManager.GetPluginHealth("auth-plugin")
fmt.Printf("Status: %s\n", health.Status)
fmt.Printf("Errors: %d\n", health.ErrorCount)

// Get dependencies
graph := pluginManager.GetDependencyGraph()
fmt.Printf("Dependencies: %v\n", graph["auth-plugin"])
```

#### Monitor Metrics

```go
// Watch metrics in real-time
ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
    health := pluginManager.GetPluginHealth("auth-plugin")
    fmt.Printf("Errors: %d, Status: %s\n", 
        health.ErrorCount, health.Status)
}
```

#### Check Database State

```sql
-- View plugin status
SELECT name, status, error_count, last_error 
FROM plugins 
WHERE name = 'auth-plugin';

-- View hook metrics
SELECT hook_type, execution_count, total_duration_ms, error_count
FROM plugin_hooks
WHERE plugin_name = 'auth-plugin';

-- View recent errors
SELECT metric_name, metric_value, recorded_at
FROM plugin_metrics
WHERE plugin_name = 'auth-plugin'
  AND metric_name = 'error'
ORDER BY recorded_at DESC
LIMIT 10;
```

### Performance Troubleshooting

#### Issue: Slow plugin load time

**Symptoms:**
- Plugin takes > 1 second to load
- Application startup is slow

**Solutions:**
1. Optimize plugin initialization
2. Lazy-load heavy dependencies
3. Use connection pooling
4. Cache expensive computations

**Measure Load Time:**
```go
start := time.Now()
err := pluginManager.LoadPlugin("./plugins/auth-plugin", config)
duration := time.Since(start)
fmt.Printf("Load time: %v\n", duration)
```

#### Issue: High hook execution overhead

**Symptoms:**
- Request latency increased after plugin installation
- Hook execution > 10ms

**Solutions:**
1. Optimize hook logic
2. Reduce hook priority (execute later)
3. Move processing to background
4. Cache results

**Profile Hook:**
```go
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 10*time.Millisecond {
            p.ctx.Logger().Warn(fmt.Sprintf("Slow hook: %v", duration))
        }
    }()
    
    // Hook logic
    return p.processRequest(hookCtx)
})
```

#### Issue: High memory usage

**Symptoms:**
- Memory usage > 100MB per plugin
- Memory grows over time

**Solutions:**
1. Profile memory usage
2. Reduce cache sizes
3. Implement proper cleanup
4. Use memory pooling

**Profile Memory:**
```go
import "runtime"

func (p *MyPlugin) logMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    p.ctx.Logger().Info(fmt.Sprintf("Memory: Alloc=%vMB, TotalAlloc=%vMB", 
        m.Alloc/1024/1024, m.TotalAlloc/1024/1024))
}
```


## Best Practices

### Plugin Development

#### 1. Follow the Single Responsibility Principle

Each plugin should have one clear purpose:

```go
// ✅ Good: Focused plugin
type AuthPlugin struct {
    // Handles authentication only
}

// ❌ Bad: Does too much
type MegaPlugin struct {
    // Handles auth, logging, caching, metrics, etc.
}
```

#### 2. Declare All Dependencies

Be explicit about dependencies:

```yaml
dependencies:
  - name: cache-plugin
    version: ">=2.0.0"
    optional: false  # Required dependency
  - name: logging-plugin
    version: ">=1.0.0"
    optional: true   # Optional dependency
```

#### 3. Request Minimal Permissions

Only request permissions you actually need:

```yaml
# ✅ Good: Minimal permissions
permissions:
  database: true
  cache: true
  router: false
  filesystem: false
  network: false
  exec: false

# ❌ Bad: Excessive permissions
permissions:
  database: true
  cache: true
  router: true
  filesystem: true
  network: true
  exec: true  # Dangerous!
```

#### 4. Handle Errors Gracefully

Don't crash on errors:

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // ✅ Good: Validate and return error
    config := ctx.PluginConfig()
    apiKey, ok := config["api_key"].(string)
    if !ok || apiKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    // ❌ Bad: Panic
    // apiKey := config["api_key"].(string)  // Panics if not string
}
```

#### 5. Clean Up Resources

Always clean up in Stop() and Cleanup():

```go
func (p *MyPlugin) Stop() error {
    // Stop goroutines
    close(p.stopChan)
    p.wg.Wait()
    
    // Close connections
    if p.conn != nil {
        p.conn.Close()
    }
    
    // Flush buffers
    p.flushBuffers()
    
    return nil
}
```

#### 6. Use Semantic Versioning

Follow semantic versioning strictly:

```
1.0.0 → 1.0.1  (Bug fix)
1.0.1 → 1.1.0  (New feature, backward compatible)
1.1.0 → 2.0.0  (Breaking change)
```

#### 7. Document Your Plugin

Provide clear documentation:

```yaml
# plugin.yaml
name: auth-plugin
version: 1.0.0
description: |
  Provides JWT-based authentication and authorization.
  Supports OAuth2, API keys, and session tokens.
  
  Features:
  - JWT token generation and validation
  - Refresh token support
  - Role-based access control
  - API key management

author: John Doe <john@example.com>
```

### Configuration Management

#### 1. Provide Sensible Defaults

```go
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "timeout": map[string]interface{}{
            "type":    "duration",
            "default": "30s",  // Sensible default
        },
        "max_retries": map[string]interface{}{
            "type":    "integer",
            "default": 3,  // Sensible default
        },
    }
}
```

#### 2. Validate Configuration

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    config := ctx.PluginConfig()
    
    // Validate required fields
    if _, ok := config["api_key"]; !ok {
        return fmt.Errorf("api_key is required")
    }
    
    // Validate types
    timeout, ok := config["timeout"].(string)
    if !ok {
        return fmt.Errorf("timeout must be a string")
    }
    
    // Validate values
    duration, err := time.ParseDuration(timeout)
    if err != nil {
        return fmt.Errorf("invalid timeout format: %w", err)
    }
    if duration < 0 {
        return fmt.Errorf("timeout must be positive")
    }
    
    return nil
}
```

#### 3. Support Configuration Updates

```go
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    // Update configuration atomically
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if timeout, ok := config["timeout"].(string); ok {
        duration, err := time.ParseDuration(timeout)
        if err != nil {
            return err
        }
        p.timeout = duration
    }
    
    p.ctx.Logger().Info("Configuration updated")
    return nil
}
```


### Performance Optimization

#### 1. Minimize Hook Overhead

Keep hooks fast (< 1ms):

```go
// ✅ Good: Fast hook
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    // Quick validation
    if !isValid(hookCtx.Context()) {
        return fmt.Errorf("invalid")
    }
    return nil
})

// ❌ Bad: Slow hook
ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    // Slow database query
    user, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
    // ...
})
```

#### 2. Use Caching

Cache expensive operations:

```go
func (p *MyPlugin) getUser(id int) (*User, error) {
    // Check cache first
    cache := p.ctx.Cache()
    key := fmt.Sprintf("user:%d", id)
    
    if cached, found := cache.Get(key); found {
        return cached.(*User), nil
    }
    
    // Cache miss - fetch from database
    user, err := p.fetchUser(id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    cache.Set(key, user, 5*time.Minute)
    return user, nil
}
```

#### 3. Use Connection Pooling

Reuse connections:

```go
type MyPlugin struct {
    clientPool *sync.Pool
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.clientPool = &sync.Pool{
        New: func() interface{} {
            return &http.Client{
                Timeout: 30 * time.Second,
            }
        },
    }
    return nil
}

func (p *MyPlugin) makeRequest(url string) error {
    client := p.clientPool.Get().(*http.Client)
    defer p.clientPool.Put(client)
    
    resp, err := client.Get(url)
    // ...
}
```

#### 4. Batch Operations

Batch database operations:

```go
// ✅ Good: Batch insert
func (p *MyPlugin) saveUsers(users []*User) error {
    tx, err := p.ctx.Database().Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    for _, user := range users {
        tx.Exec("INSERT INTO users ...", user)
    }
    
    return tx.Commit()
}

// ❌ Bad: Individual inserts
func (p *MyPlugin) saveUsers(users []*User) error {
    for _, user := range users {
        p.ctx.Database().Exec("INSERT INTO users ...", user)
    }
}
```

### Security Best Practices

#### 1. Validate All Input

```go
func (p *MyPlugin) handleRequest(ctx pkg.Context) error {
    // Validate input
    userID := ctx.Query()["user_id"]
    if userID == "" {
        return ctx.JSON(400, map[string]string{"error": "user_id required"})
    }
    
    // Validate format
    id, err := strconv.Atoi(userID)
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "invalid user_id"})
    }
    
    // Validate range
    if id < 1 {
        return ctx.JSON(400, map[string]string{"error": "user_id must be positive"})
    }
    
    // Process request
    return p.processUser(id)
}
```

#### 2. Use Prepared Statements

Prevent SQL injection:

```go
// ✅ Good: Prepared statement
func (p *MyPlugin) getUser(id int) (*User, error) {
    row := p.ctx.Database().Query("SELECT * FROM users WHERE id = ?", id)
    // ...
}

// ❌ Bad: String concatenation
func (p *MyPlugin) getUser(id int) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", id)
    row := p.ctx.Database().Query(query)
    // ...
}
```

#### 3. Sanitize Output

Prevent XSS attacks:

```go
import "html"

func (p *MyPlugin) renderUser(user *User) string {
    // Sanitize user input
    safeName := html.EscapeString(user.Name)
    safeEmail := html.EscapeString(user.Email)
    
    return fmt.Sprintf("<div>%s (%s)</div>", safeName, safeEmail)
}
```

#### 4. Use HTTPS for External Calls

```go
func (p *MyPlugin) callExternalAPI(url string) error {
    // ✅ Good: HTTPS
    if !strings.HasPrefix(url, "https://") {
        return fmt.Errorf("only HTTPS URLs are allowed")
    }
    
    resp, err := http.Get(url)
    // ...
}
```

#### 5. Store Secrets Securely

```go
// ✅ Good: Use environment variables or secret management
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    apiKey := os.Getenv("API_KEY")
    if apiKey == "" {
        return fmt.Errorf("API_KEY environment variable not set")
    }
    p.apiKey = apiKey
    return nil
}

// ❌ Bad: Hardcoded secrets
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.apiKey = "hardcoded-secret-key"  // Never do this!
    return nil
}
```

### Testing Best Practices

#### 1. Write Unit Tests

```go
func TestPluginInitialize(t *testing.T) {
    // Create mock context
    mockCtx := &MockPluginContext{
        config: map[string]interface{}{
            "api_key": "test-key",
        },
    }
    
    // Test initialization
    plugin := &MyPlugin{}
    err := plugin.Initialize(mockCtx)
    
    if err != nil {
        t.Errorf("Initialize failed: %v", err)
    }
}
```

#### 2. Test Error Conditions

```go
func TestPluginInitializeWithoutAPIKey(t *testing.T) {
    mockCtx := &MockPluginContext{
        config: map[string]interface{}{},
    }
    
    plugin := &MyPlugin{}
    err := plugin.Initialize(mockCtx)
    
    if err == nil {
        t.Error("Expected error for missing API key")
    }
}
```

#### 3. Test Cleanup

```go
func TestPluginCleanup(t *testing.T) {
    plugin := setupPlugin(t)
    
    // Start plugin
    plugin.Start()
    
    // Stop plugin
    err := plugin.Stop()
    if err != nil {
        t.Errorf("Stop failed: %v", err)
    }
    
    // Verify cleanup
    if plugin.conn != nil {
        t.Error("Connection not closed")
    }
}
```

### Deployment Best Practices

#### 1. Version Your Plugins

Use semantic versioning and tag releases:

```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

#### 2. Test Before Deploying

```bash
# Run tests
go test ./...

# Build plugin
go build -o auth-plugin

# Test in staging
./deploy-to-staging.sh

# Monitor for errors
./monitor-plugin.sh auth-plugin

# Deploy to production
./deploy-to-production.sh
```

#### 3. Use Gradual Rollout

Deploy to one server at a time:

```bash
# Deploy to server 1
ssh server1 "systemctl stop app && cp auth-plugin /plugins/ && systemctl start app"

# Monitor for 5 minutes
sleep 300

# If successful, deploy to server 2
ssh server2 "systemctl stop app && cp auth-plugin /plugins/ && systemctl start app"
```

#### 4. Monitor After Deployment

```bash
# Watch error rates
watch -n 5 'curl -s http://localhost:8080/metrics | grep plugin_error_count'

# Watch performance
watch -n 5 'curl -s http://localhost:8080/metrics | grep plugin_hook_duration'
```

## Related Documentation

- [Plugin Development Guide](PLUGIN_DEVELOPMENT.md) - Comprehensive guide for developing plugins
- [API Reference](API_REFERENCE.md) - Complete framework API reference
- [Architecture](ARCHITECTURE.md) - Framework architecture overview
- [Examples](../examples/plugins/) - Example plugin implementations

## Support

For issues, questions, or contributions:

- GitHub Issues: https://github.com/yourusername/rockstar/issues
- Documentation: https://rockstar-framework.dev/docs
- Community: https://discord.gg/rockstar-framework

