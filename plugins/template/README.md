# Plugin Template

This is a template for creating compile-time plugins for the Rockstar Web Framework.

## Quick Start

1. **Copy this template**:
   ```bash
   cp -r plugins/template plugins/my-plugin
   cd plugins/my-plugin
   ```

2. **Update the package name** in `plugin.go`:
   ```go
   package myplugin  // Change from 'template'
   ```

3. **Update the plugin registration** in `plugin.go`:
   ```go
   func init() {
       pkg.RegisterPlugin("my-plugin", func() pkg.Plugin {
           return &MyPlugin{}  // Change from TemplatePlugin
       })
   }
   ```

4. **Rename the struct** and update all methods:
   ```go
   type MyPlugin struct {  // Change from TemplatePlugin
       ctx pkg.PluginContext
       // Add your fields
   }
   ```

5. **Update `plugin.yaml`** with your plugin's metadata:
   ```yaml
   name: my-plugin
   version: 1.0.0
   description: My awesome plugin
   author: Your Name <your.email@example.com>
   ```

6. **Initialize Go module**:
   ```bash
   go mod init github.com/echterhof/rockstar-web-framework/plugins/my-plugin
   go mod edit -require github.com/echterhof/rockstar-web-framework@latest
   go mod edit -replace github.com/echterhof/rockstar-web-framework=../..
   go mod tidy
   ```

7. **Implement your plugin logic** in the interface methods

8. **Write tests** in `plugin_test.go`

9. **Add to main application** in `cmd/rockstar/main.go`:
   ```go
   import _ "github.com/echterhof/rockstar-web-framework/plugins/my-plugin"
   ```

10. **Update root go.mod**:
    ```bash
    cd ../..
    go mod edit -replace github.com/echterhof/rockstar-web-framework/plugins/my-plugin=./plugins/my-plugin
    go mod tidy
    ```

11. **Build and test**:
    ```bash
    make build
    ./rockstar --config config.yaml
    ```

## Plugin Structure

```
plugins/my-plugin/
├── plugin.go       # Plugin implementation
├── plugin.yaml     # Plugin manifest
├── plugin_test.go  # Plugin tests
├── go.mod          # Go module definition
└── README.md       # Plugin documentation
```

## Plugin Interface

All plugins must implement the `pkg.Plugin` interface:

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

## Plugin Context

The `PluginContext` provides access to framework services:

```go
type PluginContext interface {
    // Service access
    Database() DatabaseManager
    Cache() CacheManager
    Router() RouterEngine
    Config() *Config
    Logger() Logger
    
    // Plugin functionality
    RegisterHook(hookType HookType, priority int, handler HookHandler) error
    RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error
    PublishEvent(name string, data interface{}) error
    SubscribeEvent(name string, handler EventHandler) error
    ExportService(name string, service interface{}) error
    ImportService(pluginName, serviceName string) (interface{}, error)
    
    // Plugin information
    PluginName() string
    PluginConfig() map[string]interface{}
    PluginStorage() PluginStorage
}
```

## Configuration

Define your configuration schema in `ConfigSchema()`:

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
    }
}
```

Users configure your plugin in their config file:

```yaml
plugins:
  my-plugin:
    enabled: true
    config:
      api_key: "your-api-key"
      timeout: "60s"
    permissions:
      network: true
```

## Permissions

Declare required permissions in `plugin.yaml`:

```yaml
permissions:
  database: true     # Need database access
  cache: true        # Need cache access
  router: true       # Need to register routes/middleware
  network: true      # Need to make HTTP requests
  filesystem: false  # Don't need filesystem access
  exec: false        # Don't need to execute commands
```

## Hooks

Register hooks in `Initialize()`:

```go
// Pre-request hook
err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
    reqCtx := hookCtx.Context()
    // Process request
    return nil
})

// Post-request hook
err = ctx.RegisterHook(pkg.HookTypePostRequest, 100, func(hookCtx pkg.HookContext) error {
    reqCtx := hookCtx.Context()
    // Process response
    return nil
})
```

Available hook types:
- `HookTypeStartup` - Framework startup
- `HookTypeShutdown` - Framework shutdown
- `HookTypePreRequest` - Before request routing
- `HookTypePostRequest` - After handler execution
- `HookTypePreResponse` - Before sending response
- `HookTypePostResponse` - After sending response
- `HookTypeError` - On error

## Middleware

Register middleware in `Initialize()`:

```go
middleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Pre-processing
    err := next(ctx)
    // Post-processing
    return err
}

err := ctx.RegisterMiddleware("my-middleware", middleware, 100, []string{})
```

## Events

Publish events:

```go
err := p.ctx.PublishEvent("my-plugin.action", map[string]interface{}{
    "timestamp": time.Now(),
    "data": "some data",
})
```

Subscribe to events:

```go
err := ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    // Handle event
    return nil
})
```

## Services

Export a service:

```go
type MyService struct {
    plugin *MyPlugin
}

func (s *MyService) DoSomething() string {
    return "result"
}

// In Initialize()
err := ctx.ExportService("MyService", &MyService{plugin: p})
```

Import a service:

```go
// In Start()
service, err := p.ctx.ImportService("other-plugin", "OtherService")
if err != nil {
    return err
}
otherService := service.(*OtherService)
```

## Testing

Run plugin tests:

```bash
go test ./...
```

Use the mock plugin context for testing:

```go
func TestMyPlugin(t *testing.T) {
    plugin := &MyPlugin{}
    ctx := pkg.NewMockPluginContext()
    
    err := plugin.Initialize(ctx)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
}
```

## Best Practices

1. **Keep plugins focused** - One responsibility per plugin
2. **Declare all permissions** - Be explicit about what you need
3. **Provide defaults** - Make plugins work out-of-the-box
4. **Write tests** - Test plugins independently
5. **Document configuration** - Clear schema and examples
6. **Handle errors gracefully** - Don't crash the framework
7. **Use semantic versioning** - Version plugins properly
8. **Minimize dependencies** - Fewer dependencies = better security

## Examples

See the example plugins for reference:
- `plugins/auth-plugin` - Authentication and authorization
- `plugins/cache-plugin` - Response caching
- `plugins/logging-plugin` - Request/response logging
- `plugins/captcha-plugin` - CAPTCHA validation
- `plugins/storage-plugin` - File storage (S3/filesystem)

## Documentation

- [Plugin System Overview](../../docs/PLUGIN_SYSTEM.md)
- [Plugin Development Guide](../../docs/PLUGIN_DEVELOPMENT.md)
- [API Reference](../../docs/API_REFERENCE.md)

## License

This template is part of the Rockstar Web Framework and is provided under the same license.
