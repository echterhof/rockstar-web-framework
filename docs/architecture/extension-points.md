---
title: "Extension Points"
description: "How to extend and customize the framework"
category: "architecture"
tags: ["architecture", "extensibility", "customization"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "design-patterns.md"
  - "../guides/plugins.md"
---

# Extension Points

The Rockstar Web Framework is designed to be highly extensible without requiring modifications to the core codebase. This document describes all the extension points available for customizing and extending the framework's functionality.

## Overview

The framework provides multiple levels of extensibility:

1. **Middleware**: Intercept and modify request/response processing
2. **Plugin System**: Add complete features with lifecycle management
3. **Custom Managers**: Replace default implementations with custom ones
4. **Virtual File Systems**: Provide custom file system implementations
5. **Custom Routers**: Implement alternative routing strategies
6. **Protocol Handlers**: Add support for custom protocols
7. **Storage Backends**: Implement custom storage strategies
8. **Event Handlers**: React to framework and plugin events

## Middleware Extension Point

### Overview

Middleware is the simplest and most common extension point. Middleware functions can intercept requests before they reach handlers and responses before they're sent to clients.

### Middleware Signature

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

### Global Middleware

Applied to all routes:

```go
func LoggingMiddleware(ctx Context, next HandlerFunc) error {
    start := time.Now()
    path := ctx.Request().Path
    
    // Before handler
    ctx.Logger().Info("Request started", "path", path)
    
    // Call next middleware/handler
    err := next(ctx)
    
    // After handler
    duration := time.Since(start)
    ctx.Logger().Info("Request completed", "path", path, "duration", duration)
    
    return err
}

// Register globally
app.Use(LoggingMiddleware)
```

### Route-Specific Middleware

Applied to specific routes or groups:

```go
func AdminAuthMiddleware(ctx Context, next HandlerFunc) error {
    user := ctx.User()
    if user == nil || !user.IsAdmin {
        return ctx.JSON(403, map[string]string{"error": "Forbidden"})
    }
    return next(ctx)
}

// Apply to specific route
app.Router().GET("/admin/users", handler, AdminAuthMiddleware)

// Apply to route group
admin := app.Router().Group("/admin", AdminAuthMiddleware)
admin.GET("/users", listUsers)
admin.POST("/users", createUser)
```

### Middleware Use Cases

- **Authentication**: Verify user identity
- **Authorization**: Check permissions
- **Logging**: Record request/response information
- **Metrics**: Collect performance data
- **Rate Limiting**: Throttle requests
- **CORS**: Handle cross-origin requests
- **Compression**: Compress responses
- **Caching**: Cache responses
- **Request Validation**: Validate input data
- **Error Handling**: Centralized error handling

### Example: Rate Limiting Middleware

```go
func RateLimitMiddleware(limit int, window time.Duration) MiddlewareFunc {
    return func(ctx Context, next HandlerFunc) error {
        // Get client identifier
        clientIP := ctx.Request().RemoteAddr
        key := "ratelimit:" + clientIP
        
        // Check rate limit
        allowed, err := ctx.DB().CheckRateLimit(key, limit, window)
        if err != nil {
            return err
        }
        
        if !allowed {
            return ctx.JSON(429, map[string]string{
                "error": "Rate limit exceeded",
            })
        }
        
        // Increment counter
        if err := ctx.DB().IncrementRateLimit(key, window); err != nil {
            ctx.Logger().Error("Failed to increment rate limit", "error", err)
        }
        
        return next(ctx)
    }
}

// Usage
app.Router().GET("/api/data", handler, RateLimitMiddleware(100, time.Minute))
```

## Plugin System Extension Point

### Overview

The plugin system is the most powerful extension point, allowing you to add complete features with their own lifecycle, configuration, and permissions.

### Plugin Interface

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

### Basic Plugin Example

```go
type MyPlugin struct {
    config map[string]interface{}
    logger Logger
}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }
func (p *MyPlugin) Author() string { return "Your Name" }

func (p *MyPlugin) Dependencies() []PluginDependency {
    return []PluginDependency{
        {
            Name:             "core-plugin",
            Version:          ">=1.0.0",
            Optional:         false,
            FrameworkVersion: ">=1.0.0",
        },
    }
}

func (p *MyPlugin) Initialize(ctx PluginContext) error {
    p.logger = ctx.Logger()
    p.config = ctx.PluginConfig()
    
    // Register routes
    ctx.Router().GET("/plugin/status", p.handleStatus)
    
    // Register hooks
    err := ctx.RegisterHook(HookTypePreRequest, 100, p.preRequestHook)
    if err != nil {
        return err
    }
    
    // Subscribe to events
    err = ctx.SubscribeEvent("user.created", p.onUserCreated)
    if err != nil {
        return err
    }
    
    return nil
}

func (p *MyPlugin) Start() error {
    p.logger.Info("Plugin started")
    return nil
}

func (p *MyPlugin) Stop() error {
    p.logger.Info("Plugin stopped")
    return nil
}

func (p *MyPlugin) Cleanup() error {
    p.logger.Info("Plugin cleaned up")
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
    p.logger.Info("Configuration updated")
    return nil
}
```

### Hook Registration

Plugins can register hooks to intercept framework events:

```go
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    // Pre-request hook
    ctx.RegisterHook(HookTypePreRequest, 100, func(hookCtx HookContext) error {
        reqCtx := hookCtx.Context()
        reqCtx.Logger().Info("Plugin processing request")
        
        // Add custom header
        reqCtx.SetHeader("X-Plugin", "my-plugin")
        
        return nil
    })
    
    // Post-request hook
    ctx.RegisterHook(HookTypePostRequest, 100, func(hookCtx HookContext) error {
        reqCtx := hookCtx.Context()
        
        // Log request completion
        reqCtx.Logger().Info("Request completed by plugin")
        
        return nil
    })
    
    // Error hook
    ctx.RegisterHook(HookTypeError, 100, func(hookCtx HookContext) error {
        reqCtx := hookCtx.Context()
        
        // Custom error handling
        reqCtx.Logger().Error("Error occurred in request")
        
        return nil
    })
    
    return nil
}
```

### Event Publishing and Subscription

Plugins can communicate through events:

```go
// Publishing events
func (p *MyPlugin) handleCreateUser(ctx Context) error {
    // Create user logic...
    
    // Publish event
    pluginCtx := p.getPluginContext()
    pluginCtx.PublishEvent("user.created", map[string]interface{}{
        "user_id": userID,
        "email":   email,
    })
    
    return ctx.JSON(201, user)
}

// Subscribing to events
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    return ctx.SubscribeEvent("user.created", func(event Event) error {
        data := event.Data.(map[string]interface{})
        userID := data["user_id"].(string)
        
        // Handle user creation
        p.logger.Info("User created", "user_id", userID)
        
        return nil
    })
}
```

### Service Export/Import

Plugins can export services for other plugins to use:

```go
// Export a service
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    service := &MyService{
        // Service implementation
    }
    
    return ctx.ExportService("my-service", service)
}

// Import a service from another plugin
func (p *AnotherPlugin) Initialize(ctx PluginContext) error {
    service, err := ctx.ImportService("my-plugin", "my-service")
    if err != nil {
        return err
    }
    
    myService := service.(*MyService)
    // Use the service
    
    return nil
}
```

### Plugin Storage

Plugins have isolated key-value storage:

```go
func (p *MyPlugin) Initialize(ctx PluginContext) error {
    storage := ctx.PluginStorage()
    
    // Store data
    storage.Set("api_key", "secret-key")
    storage.Set("counter", 0)
    
    // Retrieve data
    apiKey, _ := storage.Get("api_key")
    counter, _ := storage.Get("counter")
    
    // List keys
    keys, _ := storage.List()
    
    // Delete data
    storage.Delete("old_key")
    
    // Clear all data
    storage.Clear()
    
    return nil
}
```

### Plugin Permissions

Plugins must declare required permissions in their manifest:

```yaml
# plugin.yaml
name: my-plugin
version: 1.0.0
description: My custom plugin
author: Your Name

permissions:
  allow_database: true
  allow_cache: true
  allow_config: false
  allow_router: true
  allow_filesystem: false
  allow_network: true
  allow_exec: false

dependencies:
  - name: core-plugin
    version: ">=1.0.0"
    optional: false
```

## Custom Manager Implementations

### Overview

You can replace default manager implementations with custom ones that implement the same interfaces.

### Custom Database Manager

```go
type myDatabaseManager struct {
    // Custom implementation
}

func (m *myDatabaseManager) Connect(config DatabaseConfig) error {
    // Custom connection logic
}

func (m *myDatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
    // Custom query logic
}

// Implement all DatabaseManager methods...

// Use custom manager
func main() {
    // Create framework without database
    config := pkg.FrameworkConfig{
        DatabaseConfig: pkg.DatabaseConfig{
            Driver: "", // Empty to skip default initialization
        },
    }
    
    app, _ := pkg.New(config)
    
    // Replace with custom manager
    customDB := &myDatabaseManager{}
    customDB.Connect(myConfig)
    
    // Note: Direct replacement requires framework modification
    // Better approach: Use plugin system to provide custom implementation
}
```

### Custom Cache Manager

```go
type redisCacheManager struct {
    client *redis.Client
}

func (c *redisCacheManager) Get(key string) (interface{}, bool) {
    val, err := c.client.Get(context.Background(), key).Result()
    if err != nil {
        return nil, false
    }
    return val, true
}

func (c *redisCacheManager) Set(key string, value interface{}, ttl time.Duration) {
    c.client.Set(context.Background(), key, value, ttl)
}

// Implement all CacheManager methods...
```

### Custom Session Storage

```go
type redisSessionStorage struct {
    client *redis.Client
}

func (s *redisSessionStorage) Save(session *Session) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    ttl := time.Until(session.ExpiresAt)
    return s.client.Set(context.Background(), session.ID, data, ttl).Err()
}

func (s *redisSessionStorage) Load(sessionID string) (*Session, error) {
    data, err := s.client.Get(context.Background(), sessionID).Bytes()
    if err != nil {
        return nil, err
    }
    
    var session Session
    err = json.Unmarshal(data, &session)
    return &session, err
}

// Implement all SessionStorage methods...
```

## Virtual File System Extension Point

### Overview

The framework uses a virtual file system interface, allowing you to provide custom file system implementations.

### VirtualFS Interface

```go
type VirtualFS interface {
    Open(name string) (File, error)
    Stat(name string) (FileInfo, error)
}

type File interface {
    io.ReadCloser
    Stat() (FileInfo, error)
}

type FileInfo interface {
    Name() string
    Size() int64
    Mode() os.FileMode
    ModTime() time.Time
    IsDir() bool
}
```

### Custom File System Example

```go
// S3-backed file system
type s3FileSystem struct {
    client *s3.Client
    bucket string
}

func (fs *s3FileSystem) Open(name string) (File, error) {
    result, err := fs.client.GetObject(context.Background(), &s3.GetObjectInput{
        Bucket: aws.String(fs.bucket),
        Key:    aws.String(name),
    })
    if err != nil {
        return nil, err
    }
    
    return &s3File{
        body: result.Body,
        name: name,
    }, nil
}

func (fs *s3FileSystem) Stat(name string) (FileInfo, error) {
    result, err := fs.client.HeadObject(context.Background(), &s3.HeadObjectInput{
        Bucket: aws.String(fs.bucket),
        Key:    aws.String(name),
    })
    if err != nil {
        return nil, err
    }
    
    return &s3FileInfo{
        name:    name,
        size:    *result.ContentLength,
        modTime: *result.LastModified,
    }, nil
}

// Usage
s3FS := &s3FileSystem{
    client: s3Client,
    bucket: "my-bucket",
}

app.Router().Static("/files", s3FS)
```

### In-Memory File System

```go
type memoryFileSystem struct {
    files map[string][]byte
    mu    sync.RWMutex
}

func (fs *memoryFileSystem) Open(name string) (File, error) {
    fs.mu.RLock()
    data, exists := fs.files[name]
    fs.mu.RUnlock()
    
    if !exists {
        return nil, os.ErrNotExist
    }
    
    return &memoryFile{
        data:   data,
        name:   name,
        reader: bytes.NewReader(data),
    }, nil
}

// Usage for testing
testFS := &memoryFileSystem{
    files: map[string][]byte{
        "index.html": []byte("<html>...</html>"),
        "style.css":  []byte("body { ... }"),
    },
}

app.Router().Static("/", testFS)
```

## Custom Router Implementation

### Overview

You can implement a custom router with different routing strategies.

### RouterEngine Interface

```go
type RouterEngine interface {
    GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    // ... other HTTP methods
    
    Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    Group(prefix string, middleware ...MiddlewareFunc) RouterEngine
    Host(hostname string) RouterEngine
    
    Match(method, path, host string) (*Route, map[string]string, bool)
    Routes() []*Route
}
```

### Custom Router Example

```go
type regexRouter struct {
    routes []*regexRoute
}

type regexRoute struct {
    method  string
    pattern *regexp.Regexp
    handler HandlerFunc
}

func (r *regexRouter) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
    pattern := regexp.MustCompile(path)
    r.routes = append(r.routes, &regexRoute{
        method:  "GET",
        pattern: pattern,
        handler: handler,
    })
    return r
}

func (r *regexRouter) Match(method, path, host string) (*Route, map[string]string, bool) {
    for _, route := range r.routes {
        if route.method == method && route.pattern.MatchString(path) {
            // Extract parameters from regex groups
            matches := route.pattern.FindStringSubmatch(path)
            params := make(map[string]string)
            
            // Map named groups to parameters
            for i, name := range route.pattern.SubexpNames() {
                if i > 0 && name != "" {
                    params[name] = matches[i]
                }
            }
            
            return &Route{
                Method:  method,
                Path:    path,
                Handler: route.handler,
            }, params, true
        }
    }
    
    return nil, nil, false
}
```

## Protocol Handler Extension Point

### Overview

Add support for custom protocols beyond HTTP/HTTPS.

### Custom Protocol Example

```go
// WebSocket protocol handler
type websocketHandler struct {
    upgrader websocket.Upgrader
}

func (h *websocketHandler) Handle(ctx Context) error {
    conn, err := h.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Handle WebSocket messages
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Process message
        response := h.processMessage(data)
        
        // Send response
        err = conn.WriteMessage(messageType, response)
        if err != nil {
            break
        }
    }
    
    return nil
}

// Register WebSocket handler
app.Router().WebSocket("/ws", func(ctx Context, conn WebSocketConnection) error {
    // Handle WebSocket connection
    return nil
})
```

## Storage Backend Extension Point

### Overview

Implement custom storage backends for sessions, cache, or other data.

### Custom Session Storage

```go
type mongoSessionStorage struct {
    collection *mongo.Collection
}

func (s *mongoSessionStorage) Save(session *Session) error {
    ctx := context.Background()
    _, err := s.collection.ReplaceOne(
        ctx,
        bson.M{"_id": session.ID},
        session,
        options.Replace().SetUpsert(true),
    )
    return err
}

func (s *mongoSessionStorage) Load(sessionID string) (*Session, error) {
    ctx := context.Background()
    var session Session
    err := s.collection.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session)
    return &session, err
}

func (s *mongoSessionStorage) Delete(sessionID string) error {
    ctx := context.Background()
    _, err := s.collection.DeleteOne(ctx, bson.M{"_id": sessionID})
    return err
}
```

## Event Handler Extension Point

### Overview

React to framework and plugin events through the event bus.

### Event Handler Example

```go
// Subscribe to framework events
func setupEventHandlers(app *Framework) {
    pluginMgr := app.PluginManager()
    
    // Get event bus (through plugin context)
    // Note: This is typically done within a plugin
    
    // Handle user events
    eventBus.Subscribe("my-plugin", "user.created", func(event Event) error {
        data := event.Data.(map[string]interface{})
        userID := data["user_id"].(string)
        
        // Send welcome email
        sendWelcomeEmail(userID)
        
        return nil
    })
    
    // Handle system events
    eventBus.Subscribe("my-plugin", "system.shutdown", func(event Event) error {
        // Cleanup before shutdown
        cleanup()
        
        return nil
    })
}
```

## Customization Examples

### Example 1: Custom Authentication Middleware

```go
func JWTAuthMiddleware(secretKey string) MiddlewareFunc {
    return func(ctx Context, next HandlerFunc) error {
        // Extract token from header
        authHeader := ctx.GetHeader("Authorization")
        if authHeader == "" {
            return ctx.JSON(401, map[string]string{"error": "Missing authorization"})
        }
        
        // Parse JWT token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := parseJWT(token, secretKey)
        if err != nil {
            return ctx.JSON(401, map[string]string{"error": "Invalid token"})
        }
        
        // Store user in context
        ctx.Set("user_id", claims["user_id"])
        ctx.Set("roles", claims["roles"])
        
        return next(ctx)
    }
}
```

### Example 2: Metrics Collection Plugin

```go
type MetricsPlugin struct {
    collector MetricsCollector
}

func (p *MetricsPlugin) Initialize(ctx PluginContext) error {
    p.collector = ctx.Metrics()
    
    // Register pre-request hook
    return ctx.RegisterHook(HookTypePreRequest, 100, func(hookCtx HookContext) error {
        reqCtx := hookCtx.Context()
        
        // Store start time
        reqCtx.Set("request_start", time.Now())
        
        return nil
    })
}

func (p *MetricsPlugin) postRequestHook(hookCtx HookContext) error {
    reqCtx := hookCtx.Context()
    
    // Calculate duration
    start, _ := reqCtx.Get("request_start")
    duration := time.Since(start.(time.Time))
    
    // Record metrics
    p.collector.RecordRequest(reqCtx.Request().Path, duration)
    
    return nil
}
```

### Example 3: Multi-Tenant Plugin

```go
type MultiTenantPlugin struct {
    db DatabaseManager
}

func (p *MultiTenantPlugin) Initialize(ctx PluginContext) error {
    p.db = ctx.Database()
    
    // Register pre-request hook for tenant resolution
    return ctx.RegisterHook(HookTypePreRequest, 50, func(hookCtx HookContext) error {
        reqCtx := hookCtx.Context()
        
        // Extract hostname
        host := reqCtx.Request().Host
        
        // Load tenant
        tenant, err := p.db.LoadTenantByHost(host)
        if err != nil {
            return reqCtx.JSON(404, map[string]string{"error": "Tenant not found"})
        }
        
        // Store tenant in context
        reqCtx.Set("tenant", tenant)
        
        return nil
    })
}
```

## Best Practices

### Extension Point Selection

Choose the appropriate extension point:

- **Middleware**: For request/response interception
- **Plugins**: For complete features with lifecycle
- **Custom Managers**: For alternative implementations
- **Virtual FS**: For custom file storage
- **Event Handlers**: For reactive behavior

### Plugin Development

1. **Keep plugins focused**: One plugin, one responsibility
2. **Declare dependencies**: Specify required plugins and versions
3. **Request minimal permissions**: Only request needed permissions
4. **Handle errors gracefully**: Don't crash the application
5. **Clean up resources**: Implement proper cleanup in Stop/Cleanup
6. **Document configuration**: Provide clear config schema
7. **Test thoroughly**: Test initialization, lifecycle, and error cases

### Middleware Development

1. **Call next()**: Always call next unless intentionally stopping
2. **Handle errors**: Properly handle and return errors
3. **Be efficient**: Minimize overhead in hot path
4. **Document behavior**: Explain what the middleware does
5. **Make it reusable**: Design for multiple use cases

### Custom Manager Development

1. **Implement full interface**: Implement all required methods
2. **Maintain compatibility**: Follow interface contracts
3. **Handle edge cases**: Test error conditions
4. **Document differences**: Explain deviations from default
5. **Provide examples**: Show how to use custom manager

## Summary

The Rockstar Web Framework provides extensive extension points:

- **Middleware**: Simple request/response interception
- **Plugin System**: Complete feature addition with lifecycle
- **Custom Managers**: Alternative service implementations
- **Virtual File Systems**: Custom file storage backends
- **Custom Routers**: Alternative routing strategies
- **Protocol Handlers**: Support for custom protocols
- **Storage Backends**: Custom data storage implementations
- **Event Handlers**: Reactive event-driven behavior

These extension points enable you to:
- Add features without modifying core code
- Customize behavior for specific needs
- Integrate with external systems
- Build reusable components
- Create plugin ecosystems

For detailed plugin development guidance, see the [Plugin Development Guide](../guides/plugins.md).

## Navigation

- [← Back to Design Patterns](design-patterns.md)
- [← Back to Architecture](README.md)
