package template

import (
	"fmt"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Plugin Template
// ============================================================================
//
// This is a template for creating compile-time plugins for the Rockstar
// Web Framework. Copy this directory and modify it to create your own plugin.
//
// Steps to create a new plugin:
// 1. Copy this template directory to plugins/your-plugin-name/
// 2. Update the package name
// 3. Update the plugin registration in init()
// 4. Implement the Plugin interface methods
// 5. Update plugin.yaml with your plugin's metadata
// 6. Add your plugin import to cmd/rockstar/main.go
//
// ============================================================================

// init registers the plugin at compile time
// This function is automatically called when the package is imported
func init() {
	pkg.RegisterPlugin("template-plugin", func() pkg.Plugin {
		return &TemplatePlugin{}
	})
}

// TemplatePlugin is a basic plugin implementation
type TemplatePlugin struct {
	// Plugin context provides access to framework services
	ctx pkg.PluginContext
	
	// Plugin configuration
	config map[string]interface{}
	
	// Add your plugin-specific fields here
	// Example:
	// apiKey string
	// timeout time.Duration
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

// Name returns the plugin name
// This must match the name used in RegisterPlugin() and plugin.yaml
func (p *TemplatePlugin) Name() string {
	return "template-plugin"
}

// Version returns the plugin version using semantic versioning
func (p *TemplatePlugin) Version() string {
	return "1.0.0"
}

// Description returns a brief description of the plugin
func (p *TemplatePlugin) Description() string {
	return "A template plugin for creating new plugins"
}

// Author returns the plugin author information
func (p *TemplatePlugin) Author() string {
	return "Your Name <your.email@example.com>"
}

// Dependencies returns the list of plugin dependencies
// Return an empty slice if the plugin has no dependencies
func (p *TemplatePlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
	
	// Example with dependencies:
	// return []pkg.PluginDependency{
	// 	{
	// 		Name:     "other-plugin",
	// 		Version:  ">=1.0.0",
	// 		Optional: false,
	// 	},
	// }
}

// Initialize is called when the plugin is loaded
// This is where you should:
// - Store the plugin context
// - Parse and validate configuration
// - Register hooks
// - Subscribe to events
// - Initialize resources
func (p *TemplatePlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing plugin...\n", p.Name())
	
	// Store the plugin context for later use
	p.ctx = ctx
	
	// Get plugin configuration
	p.config = ctx.PluginConfig()
	
	// Parse configuration
	// Example:
	// if apiKey, ok := p.config["api_key"].(string); ok {
	// 	p.apiKey = apiKey
	// } else {
	// 	return fmt.Errorf("api_key is required")
	// }
	
	// Register hooks (optional)
	// Example: Register a startup hook
	// err := ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hookCtx pkg.HookContext) error {
	// 	fmt.Printf("[%s] Startup hook executed\n", p.Name())
	// 	return nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to register startup hook: %w", err)
	// }
	
	// Register middleware (optional)
	// Example: Register a middleware
	// err = ctx.RegisterMiddleware("template", func(c pkg.Context, next pkg.HandlerFunc) error {
	// 	// Middleware logic here
	// 	return next(c)
	// }, 100, []string{})
	// if err != nil {
	// 	return fmt.Errorf("failed to register middleware: %w", err)
	// }
	
	// Subscribe to events (optional)
	// Example: Subscribe to an event
	// err = ctx.SubscribeEvent("some.event", func(event pkg.Event) error {
	// 	fmt.Printf("[%s] Received event: %s\n", p.Name(), event.Name)
	// 	return nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to subscribe to event: %w", err)
	// }
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

// Start is called after all plugins are initialized
// This is where you should start background tasks or services
func (p *TemplatePlugin) Start() error {
	fmt.Printf("[%s] Starting plugin...\n", p.Name())
	
	// Start background tasks (optional)
	// Example:
	// go p.backgroundTask()
	
	// Publish events (optional)
	// Example:
	// err := p.ctx.PublishEvent("template.started", map[string]interface{}{
	// 	"timestamp": time.Now(),
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to publish event: %w", err)
	// }
	
	fmt.Printf("[%s] Plugin started successfully\n", p.Name())
	return nil
}

// Stop is called when the plugin should stop
// This is where you should stop background tasks and prepare for shutdown
func (p *TemplatePlugin) Stop() error {
	fmt.Printf("[%s] Stopping plugin...\n", p.Name())
	
	// Stop background tasks
	// Example:
	// p.stopBackgroundTasks()
	
	// Publish events (optional)
	// Example:
	// p.ctx.PublishEvent("template.stopped", nil)
	
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

// Cleanup is called when the plugin is being unloaded
// This is where you should release all resources
func (p *TemplatePlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up plugin...\n", p.Name())
	
	// Release resources
	p.ctx = nil
	p.config = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

// ConfigSchema returns the configuration schema for this plugin
// This defines what configuration options are available and their defaults
func (p *TemplatePlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		// Example configuration fields:
		// "api_key": map[string]interface{}{
		// 	"type":        "string",
		// 	"required":    true,
		// 	"description": "API key for external service",
		// },
		// "timeout": map[string]interface{}{
		// 	"type":        "duration",
		// 	"default":     "30s",
		// 	"description": "Request timeout",
		// },
		// "enabled": map[string]interface{}{
		// 	"type":        "bool",
		// 	"default":     true,
		// 	"description": "Enable plugin functionality",
		// },
	}
}

// OnConfigChange is called when the plugin configuration is updated
// This allows plugins to react to configuration changes without reloading
func (p *TemplatePlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	
	// Update configuration
	p.config = config
	
	// Re-parse configuration
	// Example:
	// if apiKey, ok := config["api_key"].(string); ok {
	// 	p.apiKey = apiKey
	// }
	
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// Add your plugin-specific helper methods here
// Example:
// func (p *TemplatePlugin) backgroundTask() {
// 	ticker := time.NewTicker(1 * time.Minute)
// 	defer ticker.Stop()
// 	
// 	for range ticker.C {
// 		// Perform periodic task
// 	}
// }
