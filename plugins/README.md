# Rockstar Framework Plugins

This directory contains compile-time plugins for the Rockstar Web Framework. Plugins extend the framework with additional functionality and are compiled directly into the application binary.

## Available Plugins

### Template Plugin
**Location:** `template/`

A template for creating new plugins. Copy this directory to start building your own plugin.

**Features:**
- Complete plugin structure
- Documented interface methods
- Example configuration schema
- Test template

### Auth Plugin
**Location:** `auth-plugin/`

Authentication and authorization plugin with JWT and session support.

**Features:**
- JWT token authentication
- Session-based authentication
- CSRF protection
- Rate limiting for login attempts
- Configurable excluded paths
- Authentication middleware

**Use Cases:**
- Securing API endpoints
- User authentication
- Session management

### Cache Plugin
**Location:** `cache-plugin/`

HTTP response caching plugin with intelligent cache key generation.

**Features:**
- Response caching for GET/HEAD requests
- Configurable cache duration
- Path exclusion support
- Cache invalidation on mutations
- Cache hit/miss tracking
- Compression support

**Use Cases:**
- Improving API performance
- Reducing database load
- Caching static content

### Logging Plugin
**Location:** `logging-plugin/`

Request and response logging plugin with sensitive data masking.

**Features:**
- Request/response logging
- Sensitive data masking
- Async logging with buffering
- Multiple output formats (JSON, text)
- Configurable log levels
- Path exclusion

**Use Cases:**
- Debugging and monitoring
- Audit trails
- Performance analysis

### CAPTCHA Plugin
**Location:** `captcha-plugin/`

CAPTCHA generation and validation plugin for bot protection.

**Features:**
- Text-based CAPTCHA generation
- Validation with attempt limiting
- Automatic expiration
- Path-based protection
- Exported service for other plugins
- Configurable difficulty

**Use Cases:**
- Protecting login forms
- Preventing bot submissions
- Rate limiting human verification

### Storage Plugin
**Location:** `storage-plugin/`

File storage plugin with local filesystem and S3-compatible backends.

**Features:**
- Multiple storage backends (local, S3)
- File upload/download
- Extension filtering
- Size limiting
- Public URL generation
- Exported service for other plugins

**Use Cases:**
- User file uploads
- Document storage
- Image hosting
- Backup storage

## Quick Start

### Using an Existing Plugin

1. **Add plugin import** to `cmd/rockstar/main.go`:
   ```go
   import _ "github.com/echterhof/rockstar-web-framework/plugins/auth-plugin"
   ```

2. **Update root go.mod**:
   ```bash
   go mod edit -replace github.com/echterhof/rockstar-web-framework/plugins/auth-plugin=./plugins/auth-plugin
   go mod tidy
   ```

3. **Configure the plugin** in your config file:
   ```yaml
   plugins:
     auth-plugin:
       enabled: true
       config:
         jwt_secret: "your-secret-key"
         token_duration: "2h"
       permissions:
         database: true
         cache: true
         router: true
   ```

4. **Build and run**:
   ```bash
   make build
   ./rockstar --config config.yaml
   ```

### Creating a New Plugin

1. **Copy the template**:
   ```bash
   cp -r plugins/template plugins/my-plugin
   cd plugins/my-plugin
   ```

2. **Update the code**:
   - Change package name
   - Update plugin registration in `init()`
   - Rename structs and implement functionality
   - Update `plugin.yaml` with metadata

3. **Initialize Go module**:
   ```bash
   go mod init github.com/echterhof/rockstar-web-framework/plugins/my-plugin
   go mod edit -require github.com/echterhof/rockstar-web-framework@latest
   go mod edit -replace github.com/echterhof/rockstar-web-framework=../..
   go mod tidy
   ```

4. **Add to main application**:
   ```go
   // cmd/rockstar/main.go
   import _ "github.com/echterhof/rockstar-web-framework/plugins/my-plugin"
   ```

5. **Update root go.mod**:
   ```bash
   cd ../..
   go mod edit -replace github.com/echterhof/rockstar-web-framework/plugins/my-plugin=./plugins/my-plugin
   go mod tidy
   ```

6. **Build and test**:
   ```bash
   make build
   ./rockstar --config config.yaml
   ```

## Plugin Structure

Each plugin follows this structure:

```
plugins/my-plugin/
├── plugin.go       # Plugin implementation
├── plugin.yaml     # Plugin manifest
├── plugin_test.go  # Plugin tests (optional)
├── go.mod          # Go module definition
└── README.md       # Plugin documentation
```

### plugin.go

The main plugin implementation file:

```go
package myplugin

import "github.com/echterhof/rockstar-web-framework/pkg"

// Register plugin at compile time
func init() {
    pkg.RegisterPlugin("my-plugin", func() pkg.Plugin {
        return &MyPlugin{}
    })
}

type MyPlugin struct {
    ctx pkg.PluginContext
}

// Implement Plugin interface methods...
```

### plugin.yaml

The plugin manifest file:

```yaml
name: my-plugin
version: 1.0.0
description: My awesome plugin
author: Your Name <your.email@example.com>

framework:
  version: ">=1.0.0"

dependencies: []

permissions:
  database: false
  cache: false
  router: true
  config: true
  filesystem: false
  network: false
  exec: false

config:
  setting:
    type: string
    default: "value"
    description: Setting description

hooks: []
events:
  publishes: []
  subscribes: []
exports: []
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

## Permissions

Plugins must declare required permissions in their manifest:

- `database` - Access to database operations
- `cache` - Access to cache operations
- `router` - Access to router (register routes, middleware)
- `config` - Access to framework configuration
- `filesystem` - Access to filesystem operations
- `network` - Access to network operations (HTTP clients, etc.)
- `exec` - Access to execute external commands

## Hooks

Plugins can register hooks at various lifecycle points:

- `HookTypeStartup` - Framework startup
- `HookTypeShutdown` - Framework shutdown
- `HookTypePreRequest` - Before request routing
- `HookTypePostRequest` - After handler execution
- `HookTypePreResponse` - Before sending response
- `HookTypePostResponse` - After sending response
- `HookTypeError` - On error

## Events

Plugins can publish and subscribe to events for inter-plugin communication:

```go
// Publish an event
ctx.PublishEvent("user.created", map[string]interface{}{
    "user_id": 123,
})

// Subscribe to an event
ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
    // Handle event
    return nil
})
```

## Services

Plugins can export services for other plugins to use:

```go
// Export a service
type MyService struct {
    plugin *MyPlugin
}

ctx.ExportService("MyService", &MyService{plugin: p})

// Import a service
service, err := ctx.ImportService("other-plugin", "MyService")
myService := service.(*MyService)
```

## Testing

Run plugin tests:

```bash
cd plugins/my-plugin
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
9. **Export services** - Make functionality reusable
10. **Subscribe to events** - Enable inter-plugin communication

## Contributing

To contribute a new plugin:

1. Create your plugin following the structure above
2. Write comprehensive tests
3. Document configuration and usage
4. Submit a pull request
