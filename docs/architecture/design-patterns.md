---
title: "Design Patterns"
description: "Key design patterns used throughout the framework"
category: "architecture"
tags: ["architecture", "patterns", "design"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "overview.md"
  - "extension-points.md"
---

# Design Patterns

The Rockstar Web Framework employs several well-established design patterns to achieve clean architecture, maintainability, and extensibility. This document explains the key patterns used throughout the framework and how they work together.

## Manager Pattern

### Overview

The Manager pattern is the primary organizational pattern in the framework. Each major area of functionality is encapsulated in a "Manager" that provides a cohesive set of related operations.

### Structure

```go
// Manager interface defines the contract
type DatabaseManager interface {
    Connect(config DatabaseConfig) error
    Query(query string, args ...interface{}) (*sql.Rows, error)
    Close() error
    // ... other methods
}

// Implementation provides the concrete behavior
type databaseManagerImpl struct {
    db     *sql.DB
    config DatabaseConfig
    // ... internal state
}

// Constructor creates and initializes the manager
func NewDatabaseManager() DatabaseManager {
    return &databaseManagerImpl{
        // ... initialization
    }
}
```

### Framework Managers

The framework includes these core managers:

- **ServerManager**: HTTP server lifecycle and protocol handling
- **RouterEngine**: Request routing and handler registration
- **DatabaseManager**: Database operations and connection management
- **CacheManager**: Application and request-level caching
- **SessionManager**: User session management
- **SecurityManager**: Authentication, authorization, and security features
- **ConfigManager**: Configuration loading and access
- **I18nManager**: Internationalization and localization
- **MetricsCollector**: Performance metrics collection
- **MonitoringManager**: Application monitoring and health checks
- **ProxyManager**: Reverse proxy functionality
- **PluginManager**: Plugin lifecycle and coordination
- **FileManager**: File system operations

### Benefits

1. **Clear Responsibilities**: Each manager has a well-defined purpose
2. **Encapsulation**: Internal implementation details are hidden
3. **Testability**: Managers can be mocked for testing
4. **Replaceability**: Custom implementations can replace defaults
5. **Discoverability**: Consistent naming makes the API intuitive

### Example Usage

```go
// Framework provides access to all managers
app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

// Access managers through the framework
db := app.Database()
cache := app.Cache()
security := app.Security()

// Or through the request context
app.Router().GET("/users/:id", func(ctx pkg.Context) error {
    // Managers available through context
    user, err := ctx.DB().QueryRow("SELECT * FROM users WHERE id = ?", ctx.Param("id"))
    // ...
})
```

## Interface + Implementation Pattern

### Overview

The framework separates interface definitions from their implementations, providing clean contracts and enabling flexibility.

### Structure

```
component.go       - Interface definition and public types
component_impl.go  - Default implementation
component_test.go  - Unit tests
```

### Example: Router

**router.go** - Interface definition:
```go
type RouterEngine interface {
    GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    // ... other methods
}

type Route struct {
    Method  string
    Path    string
    Handler HandlerFunc
    // ... other fields
}
```

**router_impl.go** - Implementation:
```go
type routerImpl struct {
    routes     []*Route
    groups     map[string]*routeGroup
    middleware []MiddlewareFunc
}

func NewRouter() RouterEngine {
    return &routerImpl{
        routes: make([]*Route, 0),
        groups: make(map[string]*routeGroup),
    }
}

func (r *routerImpl) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
    // Implementation details
}
```

### Benefits

1. **Contract Clarity**: Interfaces define clear contracts
2. **Implementation Hiding**: Internal details are encapsulated
3. **Testing**: Easy to create mock implementations
4. **Flexibility**: Multiple implementations can coexist
5. **Documentation**: Interfaces serve as API documentation

### Custom Implementations

Users can provide custom implementations:

```go
// Custom router implementation
type myRouter struct {
    // Custom fields
}

func (r *myRouter) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
    // Custom routing logic
}

// Use custom router
app, _ := pkg.New(config)
customRouter := &myRouter{}
// Framework can use custom router through the interface
```

## Context-Driven Architecture

### Overview

The Context-Driven pattern eliminates global state by providing all framework services through a request-scoped Context interface. This is the most distinctive pattern in the framework.

### Structure

```go
// Context interface provides access to everything
type Context interface {
    // Request/Response
    Request() *Request
    Response() ResponseWriter
    
    // Routing
    Param(name string) string
    Query() map[string]string
    
    // Services (all managers accessible)
    DB() DatabaseManager
    Cache() CacheManager
    Session() SessionManager
    Security() SecurityManager
    Config() ConfigManager
    I18n() I18nManager
    Files() FileManager
    Logger() Logger
    Metrics() MetricsCollector
    
    // Response helpers
    JSON(statusCode int, data interface{}) error
    HTML(statusCode int, template string, data interface{}) error
    // ... other methods
}
```

### Implementation

```go
type contextImpl struct {
    request  *Request
    response ResponseWriter
    params   map[string]string
    
    // Manager references
    db       DatabaseManager
    cache    CacheManager
    session  SessionManager
    security SecurityManager
    config   ConfigManager
    i18n     I18nManager
    files    FileManager
    logger   Logger
    metrics  MetricsCollector
    
    // Context control
    ctx      context.Context
    cancel   context.CancelFunc
}
```

### Benefits

1. **No Global State**: Everything is request-scoped
2. **Explicit Dependencies**: Clear what each handler needs
3. **Testability**: Easy to create test contexts with mocks
4. **Isolation**: Requests don't interfere with each other
5. **Lifecycle Management**: Resources tied to request lifecycle

### Example Usage

```go
app.Router().GET("/users/:id", func(ctx pkg.Context) error {
    // All services available through context
    userID := ctx.Param("id")
    
    // Database access
    var user User
    err := ctx.DB().QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    if err != nil {
        return err
    }
    
    // Cache access
    ctx.Cache().Set("user:"+userID, user, 5*time.Minute)
    
    // I18n
    greeting := ctx.I18n().Translate("greeting", map[string]interface{}{
        "name": user.Name,
    })
    
    // Response
    return ctx.JSON(200, map[string]interface{}{
        "user":     user,
        "greeting": greeting,
    })
})
```

### Testing with Context

```go
func TestUserHandler(t *testing.T) {
    // Create mock context
    mockDB := &mockDatabaseManager{}
    mockCache := &mockCacheManager{}
    
    ctx := &testContext{
        db:    mockDB,
        cache: mockCache,
        params: map[string]string{"id": "123"},
    }
    
    // Test handler
    err := userHandler(ctx)
    assert.NoError(t, err)
}
```

## Dependency Injection

### Overview

The framework uses constructor-based dependency injection to wire components together without requiring a DI framework.

### Framework Initialization

```go
func New(config FrameworkConfig) (*Framework, error) {
    f := &Framework{}
    
    // Initialize managers in dependency order
    f.config = NewConfigManager()
    
    f.i18n, err = NewI18nManager(config.I18nConfig)
    if err != nil {
        return nil, err
    }
    
    f.database = NewDatabaseManager()
    if err := f.database.Connect(config.DatabaseConfig); err != nil {
        return nil, err
    }
    
    // Session manager depends on database and cache
    f.cache = NewCacheManager(config.CacheConfig)
    f.session, err = NewSessionManager(&config.SessionConfig, f.database, f.cache)
    if err != nil {
        return nil, err
    }
    
    // Security manager depends on database
    f.security, err = NewSecurityManager(f.database, config.SecurityConfig)
    if err != nil {
        return nil, err
    }
    
    // ... initialize other managers
    
    return f, nil
}
```

### Context Creation

```go
func (s *httpServer) createContext(w http.ResponseWriter, r *http.Request) Context {
    return &contextImpl{
        request:  newRequest(r),
        response: newResponseWriter(w),
        
        // Inject all managers
        db:       s.database,
        cache:    s.cache,
        session:  s.session,
        security: s.security,
        config:   s.config,
        i18n:     s.i18n,
        files:    s.files,
        logger:   s.logger,
        metrics:  s.metrics,
        
        ctx: r.Context(),
    }
}
```

### Benefits

1. **Explicit Dependencies**: Clear what each component needs
2. **No Magic**: No reflection or code generation
3. **Compile-Time Safety**: Dependency errors caught at compile time
4. **Testability**: Easy to inject mocks
5. **Simplicity**: No DI framework to learn

## Middleware Chain Pattern

### Overview

The Middleware Chain pattern allows request processing to be composed from reusable middleware functions.

### Structure

```go
// Middleware function signature
type MiddlewareFunc func(ctx Context, next HandlerFunc) error

// Handler function signature
type HandlerFunc func(ctx Context) error
```

### Middleware Composition

```go
// Logging middleware
func LoggingMiddleware(ctx Context, next HandlerFunc) error {
    start := time.Now()
    
    // Before handler
    ctx.Logger().Info("Request started", "path", ctx.Request().Path)
    
    // Call next middleware/handler
    err := next(ctx)
    
    // After handler
    duration := time.Since(start)
    ctx.Logger().Info("Request completed", "duration", duration)
    
    return err
}

// Authentication middleware
func AuthMiddleware(ctx Context, next HandlerFunc) error {
    // Check authentication
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Continue to next middleware/handler
    return next(ctx)
}

// Usage
app.Router().GET("/protected", handler, AuthMiddleware, LoggingMiddleware)
```

### Middleware Execution Order

Middleware executes in the order specified:

```go
app.Use(GlobalMiddleware1, GlobalMiddleware2)
app.Router().GET("/path", handler, RouteMiddleware1, RouteMiddleware2)

// Execution order:
// 1. GlobalMiddleware1
// 2. GlobalMiddleware2
// 3. RouteMiddleware1
// 4. RouteMiddleware2
// 5. handler
// 6. RouteMiddleware2 (after)
// 7. RouteMiddleware1 (after)
// 8. GlobalMiddleware2 (after)
// 9. GlobalMiddleware1 (after)
```

### Benefits

1. **Composability**: Build complex behavior from simple pieces
2. **Reusability**: Middleware can be used across routes
3. **Separation of Concerns**: Cross-cutting concerns in middleware
4. **Order Control**: Explicit execution order
5. **Error Handling**: Centralized error handling in middleware

## Plugin Architecture Pattern

### Overview

The Plugin Architecture pattern enables extensibility through dynamically loaded plugins with lifecycle management and permissions.

### Plugin Interface

```go
type Plugin interface {
    // Metadata
    Name() string
    Version() string
    Description() string
    
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

### Plugin Context

Plugins receive a restricted context with permission-controlled access:

```go
type PluginContext interface {
    // Core access
    Router() RouterEngine
    Logger() Logger
    Metrics() MetricsCollector
    
    // Permission-controlled access
    Database() DatabaseManager      // Requires AllowDatabase
    Cache() CacheManager           // Requires AllowCache
    Config() ConfigManager         // Requires AllowConfig
    FileSystem() FileManager       // Requires AllowFileSystem
    Network() NetworkClient        // Requires AllowNetwork
    
    // Plugin-specific
    PluginConfig() map[string]interface{}
    PluginStorage() PluginStorage
    
    // Extension points
    RegisterHook(hookType HookType, priority int, handler HookHandler) error
    PublishEvent(event string, data interface{}) error
    SubscribeEvent(event string, handler EventHandler) error
}
```

### Hook System

Plugins can register hooks to intercept framework events:

```go
// Plugin registers a pre-request hook
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    return ctx.RegisterHook(HookTypePreRequest, 100, func(hookCtx HookContext) error {
        // Access request context
        reqCtx := hookCtx.Context()
        
        // Perform custom logic
        reqCtx.Logger().Info("Plugin processing request")
        
        return nil
    })
}
```

### Event Bus

Plugins can communicate through events:

```go
// Plugin A publishes an event
ctx.PublishEvent("user.created", map[string]interface{}{
    "user_id": userID,
    "email":   email,
})

// Plugin B subscribes to the event
ctx.SubscribeEvent("user.created", func(event Event) error {
    data := event.Data.(map[string]interface{})
    userID := data["user_id"].(string)
    
    // Handle event
    return nil
})
```

### Benefits

1. **Extensibility**: Add features without modifying core
2. **Isolation**: Plugin errors don't crash the application
3. **Security**: Permission system prevents unauthorized access
4. **Lifecycle**: Proper initialization and cleanup
5. **Communication**: Event bus enables plugin interaction

## Factory Pattern

### Overview

The Factory pattern is used for creating complex objects with proper initialization.

### Server Factory

```go
type ServerManager interface {
    NewServer(config ServerConfig) Server
}

func (sm *serverManagerImpl) NewServer(config ServerConfig) Server {
    // Apply defaults
    config.ApplyDefaults()
    
    // Create appropriate server based on config
    if config.EnableQUIC {
        return newQUICServer(config)
    }
    
    return newHTTPServer(config)
}
```

### Context Factory

```go
func (s *httpServer) createContext(w http.ResponseWriter, r *http.Request) Context {
    ctx := &contextImpl{
        request:  newRequest(r),
        response: newResponseWriter(w),
        params:   make(map[string]string),
        
        // Inject dependencies
        db:       s.database,
        cache:    s.cache,
        session:  s.session,
        // ... other managers
        
        ctx: r.Context(),
    }
    
    // Initialize request-specific state
    ctx.initializeSession()
    ctx.loadTenant()
    
    return ctx
}
```

### Benefits

1. **Encapsulation**: Complex creation logic is hidden
2. **Consistency**: Objects are always properly initialized
3. **Flexibility**: Can return different implementations
4. **Defaults**: Centralized default value application

## Strategy Pattern

### Overview

The Strategy pattern is used to provide interchangeable algorithms, particularly for storage backends.

### Session Storage Strategy

```go
type SessionStorage interface {
    Save(session *Session) error
    Load(sessionID string) (*Session, error)
    Delete(sessionID string) error
}

// Database storage strategy
type databaseSessionStorage struct {
    db DatabaseManager
}

// Cache storage strategy
type cacheSessionStorage struct {
    cache CacheManager
}

// Memory storage strategy
type memorySessionStorage struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

// Session manager uses the strategy
type sessionManagerImpl struct {
    storage SessionStorage
}
```

### Cache Eviction Strategy

```go
type EvictionStrategy interface {
    Evict(cache *Cache) error
}

type LRUEviction struct{}
type LFUEviction struct{}
type TTLEviction struct{}
```

### Benefits

1. **Flexibility**: Swap algorithms at runtime
2. **Testability**: Easy to test different strategies
3. **Extensibility**: Add new strategies without changing existing code
4. **Separation**: Algorithm logic separated from usage

## Observer Pattern

### Overview

The Observer pattern is used in the event bus and hook system for decoupled communication.

### Event Bus Implementation

```go
type EventBus interface {
    Publish(event string, data interface{}) error
    Subscribe(pluginName, event string, handler EventHandler) error
}

type eventBusImpl struct {
    subscribers map[string][]subscription
    mu          sync.RWMutex
}

type subscription struct {
    pluginName string
    handler    EventHandler
}

func (eb *eventBusImpl) Publish(event string, data interface{}) error {
    eb.mu.RLock()
    subs := eb.subscribers[event]
    eb.mu.RUnlock()
    
    // Notify all subscribers
    for _, sub := range subs {
        if err := sub.handler(Event{
            Name:   event,
            Data:   data,
            Source: "framework",
        }); err != nil {
            // Log error but continue
        }
    }
    
    return nil
}
```

### Benefits

1. **Decoupling**: Publishers don't know about subscribers
2. **Extensibility**: Add subscribers without changing publishers
3. **Flexibility**: Dynamic subscription management
4. **Scalability**: Supports many-to-many relationships

## Adapter Pattern

### Overview

The Adapter pattern is used to integrate external libraries and provide consistent interfaces.

### Database Driver Adapter

```go
// Framework interface
type DatabaseManager interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
    // ...
}

// Adapts sql.DB to DatabaseManager
type databaseManagerImpl struct {
    db *sql.DB  // External library
}

func (dm *databaseManagerImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
    // Adapt to sql.DB
    return dm.db.Query(query, args...)
}
```

### Virtual File System Adapter

```go
type VirtualFS interface {
    Open(name string) (File, error)
    Stat(name string) (FileInfo, error)
}

// Adapts os package to VirtualFS
type osFileSystem struct {
    root string
}

func (fs *osFileSystem) Open(name string) (File, error) {
    return os.Open(filepath.Join(fs.root, name))
}
```

### Benefits

1. **Integration**: Seamlessly integrate external libraries
2. **Consistency**: Provide consistent interfaces
3. **Testability**: Easy to create test adapters
4. **Flexibility**: Swap implementations without changing code

## Summary

The Rockstar Web Framework combines these design patterns to create a cohesive, maintainable architecture:

- **Manager Pattern**: Organizes functionality into cohesive services
- **Interface + Implementation**: Separates contracts from implementations
- **Context-Driven**: Eliminates global state through request contexts
- **Dependency Injection**: Wires components without a DI framework
- **Middleware Chain**: Composes request processing from reusable pieces
- **Plugin Architecture**: Enables safe, controlled extensibility
- **Factory**: Creates complex objects with proper initialization
- **Strategy**: Provides interchangeable algorithms
- **Observer**: Enables decoupled event-driven communication
- **Adapter**: Integrates external libraries consistently

These patterns work together to provide a framework that is:
- **Maintainable**: Clear structure and separation of concerns
- **Testable**: Easy to mock and test components
- **Extensible**: Multiple extension points for customization
- **Performant**: Efficient patterns with minimal overhead
- **Secure**: Controlled access through permissions and interfaces

## Navigation

- [← Back to Overview](overview.md)
- [Next: Extension Points →](extension-points.md)
