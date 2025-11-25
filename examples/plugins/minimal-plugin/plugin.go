package main

import (
	"github.com/yourusername/rockstar/pkg"
)

// MinimalPlugin is a minimal example plugin that demonstrates the basic plugin interface
type MinimalPlugin struct {
	ctx pkg.PluginContext
}

// Name returns the plugin name
func (p *MinimalPlugin) Name() string {
	return "minimal-plugin"
}

// Version returns the plugin version
func (p *MinimalPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *MinimalPlugin) Description() string {
	return "A minimal example plugin demonstrating the basic plugin interface"
}

// Author returns the plugin author
func (p *MinimalPlugin) Author() string {
	return "Rockstar Framework Team"
}

// Dependencies returns the plugin dependencies
func (p *MinimalPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

// Initialize initializes the plugin with the provided context
func (p *MinimalPlugin) Initialize(ctx pkg.PluginContext) error {
	p.ctx = ctx
	
	// Log initialization
	if logger := ctx.Logger(); logger != nil {
		logger.Info("Minimal plugin initialized")
	}
	
	return nil
}

// Start starts the plugin
func (p *MinimalPlugin) Start() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Minimal plugin started")
	}
	return nil
}

// Stop stops the plugin
func (p *MinimalPlugin) Stop() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Minimal plugin stopped")
	}
	return nil
}

// Cleanup cleans up plugin resources
func (p *MinimalPlugin) Cleanup() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Minimal plugin cleaned up")
	}
	return nil
}

// ConfigSchema returns the configuration schema
func (p *MinimalPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"message": map[string]interface{}{
			"type":        "string",
			"default":     "Hello from minimal plugin!",
			"description": "A simple message to log",
		},
	}
}

// OnConfigChange handles configuration changes
func (p *MinimalPlugin) OnConfigChange(config map[string]interface{}) error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Minimal plugin configuration changed")
	}
	return nil
}

// NewPlugin creates a new instance of the plugin
func NewPlugin() pkg.Plugin {
	return &MinimalPlugin{}
}
