package main

import (
	"fmt"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Minimal Plugin Example
// ============================================================================
//
// This is a minimal plugin implementation demonstrating the basic structure
// and required methods for a Rockstar Web Framework plugin.
//
// Requirements: 1.1, 1.2, 1.3, 2.1, 7.2
// ============================================================================

// MinimalPlugin is a basic plugin implementation
type MinimalPlugin struct {
	// Plugin context provides access to framework services
	ctx pkg.PluginContext
	
	// Plugin configuration
	config map[string]interface{}
	
	// Plugin state
	message string
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================
// All plugins must implement the pkg.Plugin interface

// Name returns the plugin name
// This is used to identify the plugin in the system
func (p *MinimalPlugin) Name() string {
	return "minimal-plugin"
}

// Version returns the plugin version
// Use semantic versioning (e.g., "1.0.0")
func (p *MinimalPlugin) Version() string {
	return "1.0.0"
}

// Description returns a brief description of the plugin
func (p *MinimalPlugin) Description() string {
	return "A minimal plugin demonstrating the basic plugin structure"
}

// Author returns the plugin author information
func (p *MinimalPlugin) Author() string {
	return "Rockstar Framework Team"
}

// Dependencies returns the list of plugin dependencies
// Return an empty slice if the plugin has no dependencies
func (p *MinimalPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

// Initialize is called when the plugin is loaded
// This is where you should:
// - Store the plugin context
// - Parse configuration
// - Register hooks
// - Subscribe to events
// - Initialize resources
func (p *MinimalPlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing plugin...\n", p.Name())
	
	// Store the plugin context for later use
	p.ctx = ctx
	
	// Get plugin configuration
	p.config = ctx.PluginConfig()
	
	// Parse configuration with defaults
	if msg, ok := p.config["message"].(string); ok {
		p.message = msg
	} else {
		p.message = "Hello from minimal plugin!"
	}
	
	// Register a startup hook
	// Hooks allow plugins to execute code at specific lifecycle points
	err := ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hookCtx pkg.HookContext) error {
		fmt.Printf("[%s] Startup hook executed: %s\n", p.Name(), p.message)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register startup hook: %w", err)
	}
	
	// Register a pre-request hook
	// This hook runs before each request is processed
	err = ctx.RegisterHook(pkg.HookTypePreRequest, 100, func(hookCtx pkg.HookContext) error {
		// Access the request context
		reqCtx := hookCtx.Context()
		if reqCtx != nil {
			req := reqCtx.Request()
			if req != nil && req.URL != nil {
				fmt.Printf("[%s] Pre-request hook: %s\n", p.Name(), req.URL.Path)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register pre-request hook: %w", err)
	}
	
	// Register a shutdown hook
	err = ctx.RegisterHook(pkg.HookTypeShutdown, 100, func(hookCtx pkg.HookContext) error {
		fmt.Printf("[%s] Shutdown hook executed\n", p.Name())
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register shutdown hook: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

// Start is called after all plugins are initialized
// This is where you should start background tasks or services
func (p *MinimalPlugin) Start() error {
	fmt.Printf("[%s] Starting plugin...\n", p.Name())
	
	// Example: Start a background task (commented out)
	// go p.backgroundTask()
	
	fmt.Printf("[%s] Plugin started successfully\n", p.Name())
	return nil
}

// Stop is called when the plugin should stop
// This is where you should stop background tasks and prepare for shutdown
func (p *MinimalPlugin) Stop() error {
	fmt.Printf("[%s] Stopping plugin...\n", p.Name())
	
	// Example: Stop background tasks
	// p.stopBackgroundTasks()
	
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

// Cleanup is called when the plugin is being unloaded
// This is where you should release all resources
func (p *MinimalPlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up plugin...\n", p.Name())
	
	// Release resources
	p.ctx = nil
	p.config = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

// ConfigSchema returns the configuration schema for this plugin
// This defines what configuration options are available and their defaults
func (p *MinimalPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"message": map[string]interface{}{
			"type":        "string",
			"default":     "Hello from minimal plugin!",
			"description": "Message to display on startup",
		},
	}
}

// OnConfigChange is called when the plugin configuration is updated
// This allows plugins to react to configuration changes without reloading
func (p *MinimalPlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	
	// Update configuration
	p.config = config
	
	// Parse new configuration
	if msg, ok := config["message"].(string); ok {
		p.message = msg
		fmt.Printf("[%s] New message: %s\n", p.Name(), p.message)
	}
	
	return nil
}

// ============================================================================
// Plugin Entry Point
// ============================================================================
// This is required for the plugin to be loadable

// NewPlugin creates a new instance of the plugin
// This function is called by the plugin loader
func NewPlugin() pkg.Plugin {
	return &MinimalPlugin{}
}

// main is required for building the plugin as a standalone binary
func main() {
	// This is typically empty for plugins
	// The plugin is loaded and initialized by the framework
}
