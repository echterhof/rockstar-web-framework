# API Reference

Complete API reference for the Rockstar Web Framework.

## Table of Contents

1. [Framework](#framework)
2. [Context](#context)
3. [Router](#router)
4. [Server](#server)
5. [Database](#database)
6. [Cache](#cache)
7. [Session](#session)
8. [Security](#security)
9. [Configuration](#configuration)
10. [Internationalization](#internationalization)
11. [Plugin System](#plugin-system)

## Framework

The main framework struct that wires all components together.

### Creating a Framework Instance

```go
func New(config FrameworkConfig) (*Framework, error)
```

Creates a new framework instance with the given configuration.

**Parameters:**
- `config`: FrameworkConfig - Complete framework configuration

**Returns:**
- `*Framework`: Framework instance
- `error`: Error if initialization fails

**Example:**
```go
app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
```

### Framework Methods

#### Router()

```go
func (f *Framework) Router() RouterEngine
```

Returns the framework's router for route registration.

#### Use()

```go
func (f *Framework) Use(middleware ...MiddlewareFunc)
```

Adds global middleware to the framework.

**Parameters:**
- `middleware`: One or more middleware functions

**Example:**
```go
app.Use(loggingMiddleware, recoveryMiddleware)
```

#### Listen()

```go
func (f *Framework) Listen(addr string) error
```

Starts the framework server on the specified address.

**Parameters:**
- `addr`: Server address (e.g., ":8080")

**Returns:**
- `error`: Error if server fails to start

#### ListenTLS()

```go
func (f *Framework) ListenTLS(addr, certFile, keyFile string) error
```

Starts the framework server with TLS.

**Parameters:**
- `addr`: Server address
- `certFile`: Path to TLS certificate file
- `keyFile`: Path to TLS key file

#### ListenQUIC()

```go
func (f *Framework) ListenQUIC(addr, certFile, keyFile string) error
```

Starts the framework server with QUIC protocol.

#### Shutdown()

```go
func (f *Framework) Shutdown(timeout time.Duration) error
```

Gracefully shuts down the framework.

**Parameters:**
- `timeout`: Maximum time to wait for shutdown

#### RegisterStartupHook()

```go
func (f *Framework) RegisterStartupHook(hook func(ctx context.Context) error)
```

Registers a function to be called during startup.

#### RegisterShutdownHook()

```go
func (f *Framework) RegisterShutdownHook(hook func(ctx context.Context) error)
```

Registers a function to be called during graceful shutdown.

#### PluginManager()

```go
func (f *Framework) PluginManager() PluginManager
```

Returns the plugin manager for loading and managing plugins.

**Example:**
```go
pluginManager := app.PluginManager()
err := pluginManager.LoadPlugin("./plugins/my-plugin", config)
```

## Context

The unified request context providing access to all framework features.

### Request Data Methods

#### Request()

```go
func (ctx Context) Request() *Request
```

Returns the HTTP request object.

#### Response()

```go
func (ctx Context) Response() ResponseWriter
```

Returns the response writer.

#### Params()

```go
func (ctx Context) Params() map[string]string
```

Returns URL parameters (e.g., `/users/:id`).

**Example:**
```go
id := ctx.Params()["id"]
```

#### Query()

```go
func (ctx Context) Query() map[string]string
```

Returns query string parameters.

**Example:**
```go
search := ctx.Query()["q"]
```

#### Headers()

```go
func (ctx Context) Headers() map[string]string
```

Returns request headers.

#### Body()

```go
func (ctx Context) Body() []byte
```

Returns the request body as bytes.

### Framework Services

#### DB()

```go
func (ctx Context) DB() DatabaseManager
```

Returns the database manager.

#### Cache()

```go
func (ctx Context) Cache() CacheManager
```

Returns the cache manager.

#### Session()

```go
func (ctx Context) Session() SessionManager
```

Returns the session manager.

#### Config()

```go
func (ctx Context) Config() ConfigManager
```

Returns the configuration manager.

#### I18n()

```go
func (ctx Context) I18n() I18nManager
```

Returns the internationalization manager.

#### Logger()

```go
func (ctx Context) Logger() Logger
```

Returns the logger instance.

#### Metrics()

```go
func (ctx Context) Metrics() MetricsCollector
```

Returns the metrics collector.

### Response Methods

#### JSON()

```go
func (ctx Context) JSON(statusCode int, data interface{}) error
```

Sends a JSON response.

**Example:**
```go
return ctx.JSON(200, map[string]string{
    "message": "Success",
})
```

#### XML()

```go
func (ctx Context) XML(statusCode int, data interface{}) error
```

Sends an XML response.

#### HTML()

```go
func (ctx Context) HTML(statusCode int, template string, data interface{}) error
```

Renders an HTML template.

#### String()

```go
func (ctx Context) String(statusCode int, message string) error
```

Sends a plain text response.

#### Redirect()

```go
func (ctx Context) Redirect(statusCode int, url string) error
```

Redirects to another URL.

### Cookie Methods

#### SetCookie()

```go
func (ctx Context) SetCookie(cookie *Cookie) error
```

Sets a cookie in the response.

#### GetCookie()

```go
func (ctx Context) GetCookie(name string) (*Cookie, error)
```

Gets a cookie from the request.

### Header Methods

#### SetHeader()

```go
func (ctx Context) SetHeader(key, value string)
```

Sets a response header.

#### GetHeader()

```go
func (ctx Context) GetHeader(key string) string
```

Gets a request header.

### Form Methods

#### FormValue()

```go
func (ctx Context) FormValue(key string) string
```

Gets a form value from POST data.

#### FormFile()

```go
func (ctx Context) FormFile(key string) (*FormFile, error)
```

Gets an uploaded file from the request.

### Authentication Methods

#### User()

```go
func (ctx Context) User() *User
```

Returns the authenticated user.

#### Tenant()

```go
func (ctx Context) Tenant() *Tenant
```

Returns the current tenant (multi-tenancy).

#### IsAuthenticated()

```go
func (ctx Context) IsAuthenticated() bool
```

Checks if the user is authenticated.

#### IsAuthorized()

```go
func (ctx Context) IsAuthorized(resource, action string) bool
```

Checks if the user is authorized for an action.

## Router

The routing engine for mapping URLs to handlers.

### HTTP Method Routes

```go
func (r RouterEngine) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

### Route Groups

```go
func (r RouterEngine) Group(prefix string, middleware ...MiddlewareFunc) RouterEngine
```

Creates a route group with a common prefix and middleware.

**Example:**
```go
api := router.Group("/api/v1", authMiddleware)
api.GET("/profile", getProfile)
api.POST("/logout", logout)
```

### Host-Based Routing

```go
func (r RouterEngine) Host(hostname string) RouterEngine
```

Creates routes for a specific hostname (multi-tenancy).

**Example:**
```go
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHome)
```

### Static Files

```go
func (r RouterEngine) Static(prefix string, filesystem VirtualFS) RouterEngine
```

Serves static files from a filesystem.

**Example:**
```go
router.Static("/static", fileSystem)
```

### WebSocket

```go
func (r RouterEngine) WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine
```

Registers a WebSocket handler.

**Example:**
```go
router.WebSocket("/ws", wsHandler)
```

### Multi-Protocol APIs

```go
func (r RouterEngine) GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine
```

## Database

Database management interface.

### Connection Methods

```go
func (db DatabaseManager) Connect(config DatabaseConfig) error
func (db DatabaseManager) Close() error
```

### Query Methods

```go
func (db DatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error)
func (db DatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error)
```

**Example:**
```go
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
```

### Transaction Methods

```go
func (db DatabaseManager) Begin() (Transaction, error)
```

**Example:**
```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Perform operations
tx.Exec("INSERT INTO users ...")

// Commit
tx.Commit()
```

## Cache

Cache management interface.

### Cache Methods

```go
func (c CacheManager) Get(key string) (interface{}, bool)
func (c CacheManager) Set(key string, value interface{}, ttl time.Duration) error
func (c CacheManager) Delete(key string) error
func (c CacheManager) Clear() error
```

**Example:**
```go
// Get from cache
data, found := cache.Get("user:123")
if !found {
    // Fetch from database
    data = fetchUser(123)
    // Store in cache
    cache.Set("user:123", data, 5*time.Minute)
}
```

## Session

Session management interface.

### Session Methods

```go
func (s SessionManager) Create(ctx Context) (*Session, error)
func (s SessionManager) Load(ctx Context, sessionID string) (*Session, error)
func (s SessionManager) Save(ctx Context, session *Session) error
func (s SessionManager) Destroy(ctx Context, sessionID string) error
```

### Session Data Methods

```go
func (s SessionManager) Get(key string) interface{}
func (s SessionManager) Set(key string, value interface{})
func (s SessionManager) Delete(key string)
```

**Example:**
```go
session := ctx.Session()

// Get data
userID := session.Get("user_id")

// Set data
session.Set("user_id", "123")

// Save session
session.Save(ctx)
```

## Security

Security management interface.

### Authentication Methods

```go
func (s SecurityManager) AuthenticateOAuth2(token string) (*User, error)
func (s SecurityManager) AuthenticateJWT(token string) (*User, error)
```

### Authorization Methods

```go
func (s SecurityManager) Authorize(user *User, resource string, action string) bool
```

### Validation Methods

```go
func (s SecurityManager) ValidateRequest(ctx Context) error
func (s SecurityManager) ValidateFormData(ctx Context, rules ValidationRules) error
```

## Configuration

Configuration management interface.

### Configuration Methods

```go
func (c ConfigManager) LoadFile(path string) error
func (c ConfigManager) Get(key string) interface{}
func (c ConfigManager) GetString(key string) string
func (c ConfigManager) GetInt(key string) int
func (c ConfigManager) GetBool(key string) bool
```

**Example:**
```go
config := ctx.Config()
dbHost := config.GetString("database.host")
dbPort := config.GetInt("database.port")
```

## Internationalization

Internationalization (i18n) interface.

### I18n Methods

```go
func (i I18nManager) Translate(key string, params ...interface{}) string
func (i I18nManager) SetLocale(locale string) error
func (i I18nManager) GetLocale() string
```

**Example:**
```go
i18n := ctx.I18n()
message := i18n.Translate("welcome_message", userName)
```

## Plugin System

The plugin system enables extending framework functionality through compile-time plugins that are compiled directly into the application binary.

### PluginManager Interface

The main interface for managing plugins.

#### DiscoverPlugins()

```go
func (pm PluginManager) DiscoverPlugins() error
```

Discovers all plugins registered via `init()` functions.

**Example:**
```go
err := pluginManager.DiscoverPlugins()
```

#### InitializeAll()

```go
func (pm PluginManager) InitializeAll() error
```

Initializes all discovered plugins in dependency order.

**Example:**
```go
err := pluginManager.InitializeAll()
```

#### StartAll()

```go
func (pm PluginManager) StartAll() error
```

Starts all initialized plugins.

**Example:**
```go
err := pluginManager.StartAll()
```

#### StopAll()

```go
func (pm PluginManager) StopAll() error
```

Stops all running plugins in reverse dependency order.

**Example:**
```go
err := pluginManager.StopAll()
```

#### GetPlugin()

```go
func (pm PluginManager) GetPlugin(name string) (Plugin, error)
```

Retrieves a plugin by name.

**Example:**
```go
plugin, err := pluginManager.GetPlugin("auth-plugin")
```

#### ListPlugins()

```go
func (pm PluginManager) ListPlugins() []PluginInfo
```

Returns information about all plugins.

**Example:**
```go
plugins := pluginManager.ListPlugins()
for _, info := range plugins {
    fmt.Printf("%s v%s - %s\n", info.Name, info.Version, info.Status)
}
```

#### IsLoaded()

```go
func (pm PluginManager) IsLoaded(name string) bool
```

Checks if a plugin is loaded.

**Example:**
```go
if pluginManager.IsLoaded("auth-plugin") {
    fmt.Println("Auth plugin is loaded")
}
```

#### GetPluginHealth()

```go
func (pm PluginManager) GetPluginHealth(name string) PluginHealth
```

Returns health information for a specific plugin.

**Example:**
```go
health := pluginManager.GetPluginHealth("auth-plugin")
fmt.Printf("Status: %s, Errors: %d\n", health.Status, health.ErrorCount)
```

#### GetAllHealth()

```go
func (pm PluginManager) GetAllHealth() map[string]PluginHealth
```

Returns health information for all plugins.

#### ResolveDependencies()

```go
func (pm PluginManager) ResolveDependencies() error
```

Resolves plugin dependencies and determines load order.

**Example:**
```go
err := pluginManager.ResolveDependencies()
```

#### GetDependencyGraph()

```go
func (pm PluginManager) GetDependencyGraph() map[string][]string
```

Returns the plugin dependency graph.

**Example:**
```go
graph := pluginManager.GetDependencyGraph()
for plugin, deps := range graph {
    fmt.Printf("%s depends on: %v\n", plugin, deps)
}
```

### Plugin Interface

Interface that all plugins must implement.

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

### PluginContext Interface

Provides plugins with access to framework services.

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
    RegisterMiddleware(config MiddlewareConfig) error
    UnregisterMiddleware(name string) error
}
```

### Hook Types

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
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    // Register a pre-request hook
    ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hctx pkg.HookContext) error {
        // Process request before routing
        return nil
    })
    
    return nil
}
```

### Plugin Registration

Plugins register themselves at compile time via `init()` functions:

```go
package myplugin

import "github.com/echterhof/rockstar-web-framework/pkg"

func init() {
    pkg.RegisterPlugin("my-plugin", func() pkg.Plugin {
        return &MyPlugin{}
    })
}
```

### Plugin Configuration

Plugins can be configured via YAML, JSON, or TOML:

```yaml
plugins:
  enabled: true
  
  plugins:
    - name: auth-plugin
      enabled: true
      config:
        jwt_secret: "secret"
        token_expiry: 3600
      permissions:
        database: true
        cache: true
        router: true
```

### Plugin Manifest

Each plugin includes a manifest file (plugin.yaml):

```yaml
name: auth-plugin
version: 1.0.0
description: Authentication plugin with JWT support
author: Your Name <email@example.com>

framework:
  version: ">=1.0.0"

dependencies:
  - name: cache-plugin
    version: ">=1.0.0"
    optional: false

permissions:
  database: true
  cache: true
  router: true

config:
  jwt_secret:
    type: string
    required: true
    description: JWT signing secret
  token_expiry:
    type: int
    default: 3600
    description: Token expiration in seconds
```

### Main Application Integration

Import plugins in your main application:

```go
// cmd/rockstar/main.go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    
    // Import plugins to trigger init()
    _ "github.com/echterhof/rockstar-web-framework/plugins/auth-plugin"
    _ "github.com/echterhof/rockstar-web-framework/plugins/cache-plugin"
)

func main() {
    framework, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    framework.Start()
}
```

For detailed plugin development information, see:
- [Plugin System Documentation](PLUGIN_SYSTEM.md)
- [Plugin Development Guide](PLUGIN_DEVELOPMENT.md)

## Types

### FrameworkConfig

```go
type FrameworkConfig struct {
    ServerConfig     ServerConfig
    DatabaseConfig   DatabaseConfig
    CacheConfig      CacheConfig
    SessionConfig    SessionConfig
    ConfigFiles      []string
    I18nConfig       I18nConfig
    SecurityConfig   SecurityConfig
    MonitoringConfig MonitoringConfig
    ProxyConfig      ProxyConfig
    PluginConfig     PluginConfig
}
```

### PluginConfig

```go
type PluginConfig struct {
    Enabled   bool
    Directory string
    Plugins   []PluginLoadConfig
}
```

### PluginLoadConfig

```go
type PluginLoadConfig struct {
    Name        string
    Enabled     bool
    Path        string
    Priority    int
    Config      map[string]interface{}
    Permissions PluginPermissions
}
```

### PluginPermissions

```go
type PluginPermissions struct {
    AllowDatabase    bool
    AllowCache       bool
    AllowConfig      bool
    AllowRouter      bool
    AllowFileSystem  bool
    AllowNetwork     bool
    AllowExec        bool
    CustomPermissions map[string]bool
}
```

### PluginInfo

```go
type PluginInfo struct {
    Name        string
    Version     string
    Description string
    Author      string
    Loaded      bool
    Enabled     bool
    LoadTime    time.Time
    Status      PluginStatus
}
```

### PluginStatus

```go
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

### PluginHealth

```go
type PluginHealth struct {
    Status       PluginStatus
    ErrorCount   int64
    LastError    error
    LastErrorAt  time.Time
    HookMetrics  map[string]HookMetrics
}
```

### PluginDependency

```go
type PluginDependency struct {
    Name             string
    Version          string // Semantic version constraint
    Optional         bool
    FrameworkVersion string
}
```

### ServerConfig

```go
type ServerConfig struct {
    ReadTimeout      time.Duration
    WriteTimeout     time.Duration
    IdleTimeout      time.Duration
    MaxHeaderBytes   int
    EnableHTTP1      bool
    EnableHTTP2      bool
    EnableQUIC       bool
    EnableMetrics    bool
    MetricsPath      string
    EnablePprof      bool
    PprofPath        string
    ShutdownTimeout  time.Duration
}
```

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver          string
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}
```

### HandlerFunc

```go
type HandlerFunc func(ctx Context) error
```

Handler function signature for route handlers.

### MiddlewareFunc

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

Middleware function signature.

---

For more detailed examples, see the [examples/](../examples/) directory.
