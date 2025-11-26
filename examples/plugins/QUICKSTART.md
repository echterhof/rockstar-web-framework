# Plugin Quick Start Guide

This guide will help you get started with the Rockstar Framework plugin system in 5 minutes.

## Prerequisites

- Go 1.21 or later
- Rockstar Web Framework installed

## Quick Start

### 1. Run the Example

The fastest way to see plugins in action:

```bash
# Run the plugin usage example
go run examples/plugin_usage_example.go
```

This will:
- Load all four example plugins
- Initialize and start them
- Demonstrate plugin features
- Show plugin health status
- Perform a hot reload
- Clean up and stop all plugins

### 2. Use Plugins with Configuration

Create a configuration file:

```yaml
# my-config.yaml
plugins:
  enabled: true
  
  plugins:
    - name: minimal-plugin
      enabled: true
      path: ./examples/plugins/minimal-plugin
      config:
        message: "Hello World!"
```

Load it in your application:

```go
import "github.com/echterhof/rockstar-web-framework/pkg"

framework, _ := pkg.NewFramework(config)
framework.PluginManager().LoadPluginsFromConfig("my-config.yaml")
framework.PluginManager().InitializeAll()
framework.PluginManager().StartAll()
```

### 3. Create Your First Plugin

#### Step 1: Create the directory structure

```bash
mkdir -p my-plugin
cd my-plugin
```

#### Step 2: Create plugin.go

```go
package main

import "github.com/echterhof/rockstar-web-framework/pkg"

type MyPlugin struct {
    ctx pkg.PluginContext
}

func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My first plugin" }
func (p *MyPlugin) Author() string { return "Your Name" }
func (p *MyPlugin) Dependencies() []pkg.PluginDependency { return nil }

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    ctx.Logger().Info("My plugin initialized!")
    return nil
}

func (p *MyPlugin) Start() error {
    p.ctx.Logger().Info("My plugin started!")
    return nil
}

func (p *MyPlugin) Stop() error {
    p.ctx.Logger().Info("My plugin stopped!")
    return nil
}

func (p *MyPlugin) Cleanup() error { return nil }
func (p *MyPlugin) ConfigSchema() map[string]interface{} { return nil }
func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error { return nil }

func NewPlugin() pkg.Plugin {
    return &MyPlugin{}
}
```

#### Step 3: Create plugin.yaml

```yaml
name: my-plugin
version: 1.0.0
description: My first plugin
author: Your Name

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

config: {}
hooks: []
events:
  publishes: []
  subscribes: []
exports: []
```

#### Step 4: Create go.mod

```bash
go mod init github.com/yourusername/my-plugin
go mod edit -require github.com/echterhof/rockstar-web-framework@v1.0.0
go mod edit -replace github.com/echterhof/rockstar-web-framework=../path/to/rockstar-web-framework
```

#### Step 5: Load your plugin

```go
framework.PluginManager().LoadPlugin("./my-plugin", pkg.PluginConfig{
    Enabled: true,
})
```

## Common Use Cases

### Add a Request Hook

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register a pre-request hook
    return ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
        reqCtx := hookCtx.Context()
        ctx.Logger().Info("Processing request to: " + reqCtx.Path())
        return nil
    })
}
```

### Add Middleware

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register middleware
    return ctx.RegisterMiddleware("my-middleware", func(c pkg.Context) error {
        ctx.Logger().Info("Middleware executing")
        return nil
    }, 100, []string{}) // Empty routes = global
}
```

### Publish Events

```go
func (p *MyPlugin) Start() error {
    // Publish an event
    return p.ctx.PublishEvent("my-plugin.started", map[string]interface{}{
        "timestamp": time.Now(),
    })
}
```

### Subscribe to Events

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Subscribe to events
    return ctx.SubscribeEvent("user.created", func(event pkg.Event) error {
        ctx.Logger().Info("User created event received")
        return nil
    })
}
```

### Export a Service

```go
type MyService struct {
    plugin *MyPlugin
}

func (s *MyService) DoSomething() string {
    return "Hello from service!"
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Export service
    return ctx.ExportService("MyService", &MyService{plugin: p})
}
```

### Import a Service

```go
func (p *MyPlugin) Start() error {
    // Import service from another plugin
    service, err := p.ctx.ImportService("other-plugin", "OtherService")
    if err != nil {
        return err
    }
    
    otherService := service.(*OtherService)
    result := otherService.DoSomething()
    p.ctx.Logger().Info("Got result: " + result)
    
    return nil
}
```

## Testing Your Plugin

Create a test file `plugin_test.go`:

```go
package main

import (
    "testing"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func TestPluginInitialize(t *testing.T) {
    plugin := NewPlugin()
    
    // Create a mock context
    ctx := pkg.NewMockPluginContext()
    
    err := plugin.Initialize(ctx)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
    
    if plugin.Name() != "my-plugin" {
        t.Errorf("Expected name 'my-plugin', got '%s'", plugin.Name())
    }
}
```

Run tests:

```bash
go test ./...
```

## Next Steps

1. **Read the full documentation**: See [README.md](README.md) for detailed information
2. **Study the examples**: Look at the four example plugins for different patterns

## Troubleshooting

### Plugin won't load

- Check the manifest is valid YAML
- Verify all required fields are present
- Ensure the plugin path is correct

### Permission denied errors

- Add required permissions to the manifest
- Grant permissions in the framework configuration

### Hook not executing

- Verify the hook is registered in Initialize()
- Check the plugin is started
- Look for errors in the logs

## Resources

- [Full README](README.md)
- [Example Configuration](example-config.yaml)
- [Usage Example](../plugin_usage_example.go)

## Getting Help

If you run into issues:

1. Check the logs for error messages
2. Review the example plugins
3. Read the documentation
4. Ask in the community forums
5. Open an issue on GitHub

Happy plugin development! 🚀
