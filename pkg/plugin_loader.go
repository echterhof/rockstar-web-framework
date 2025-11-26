package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// PluginLoader handles loading and managing plugin processes
type PluginLoader interface {
	// Load loads a plugin from the specified path
	Load(path string, config PluginConfig) (Plugin, error)

	// Unload stops and cleans up a plugin process
	Unload(plugin Plugin) error

	// ResolvePath resolves a plugin path (absolute or relative)
	ResolvePath(path string) (string, error)
}

// pluginLoaderImpl is the default implementation of PluginLoader
type pluginLoaderImpl struct {
	mu        sync.RWMutex
	processes map[string]*pluginProcess
	baseDir   string
	logger    Logger
}

// pluginProcess represents a running plugin process
type pluginProcess struct {
	plugin    Plugin
	cmd       *exec.Cmd
	path      string
	config    PluginConfig
	startTime time.Time
	stopChan  chan struct{}
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(baseDir string, logger Logger) PluginLoader {
	if baseDir == "" {
		baseDir = "."
	}

	return &pluginLoaderImpl{
		processes: make(map[string]*pluginProcess),
		baseDir:   baseDir,
		logger:    logger,
	}
}

// Load loads a plugin from the specified path
func (l *pluginLoaderImpl) Load(path string, config PluginConfig) (Plugin, error) {
	// Resolve the plugin path
	resolvedPath, err := l.ResolvePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve plugin path: %w", err)
	}

	// Check if the plugin binary exists
	if _, err := os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin binary not found at %s", resolvedPath)
		}
		return nil, fmt.Errorf("failed to access plugin binary: %w", err)
	}

	// Check if the file is executable
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat plugin binary: %w", err)
	}

	// On Unix-like systems, check execute permission
	if fileInfo.Mode()&0111 == 0 {
		l.logger.Warn(fmt.Sprintf("Plugin binary at %s may not be executable", resolvedPath))
	}

	// For now, we'll create a proxy plugin that will communicate with the process
	// In a full implementation, this would spawn the process and establish gRPC connection
	plugin := &processPlugin{
		path:   resolvedPath,
		config: config,
		loader: l,
	}

	// Store the process information
	l.mu.Lock()
	l.processes[resolvedPath] = &pluginProcess{
		plugin:    plugin,
		path:      resolvedPath,
		config:    config,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
	}
	l.mu.Unlock()

	return plugin, nil
}

// Unload stops and cleans up a plugin process
func (l *pluginLoaderImpl) Unload(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	// Find the process for this plugin
	l.mu.Lock()
	defer l.mu.Unlock()

	var processPath string
	for path, proc := range l.processes {
		if proc.plugin == plugin {
			processPath = path
			break
		}
	}

	if processPath == "" {
		return fmt.Errorf("plugin process not found")
	}

	proc := l.processes[processPath]

	// Signal the process to stop
	close(proc.stopChan)

	// If there's a running command, terminate it
	if proc.cmd != nil && proc.cmd.Process != nil {
		if err := proc.cmd.Process.Kill(); err != nil {
			l.logger.Error(fmt.Sprintf("Failed to kill plugin process: %v", err))
		}
	}

	// Remove from tracking
	delete(l.processes, processPath)

	return nil
}

// ResolvePath resolves a plugin path (absolute or relative)
func (l *pluginLoaderImpl) ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("plugin path cannot be empty")
	}

	// If the path is already absolute, return it
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// Otherwise, resolve it relative to the base directory
	absPath := filepath.Join(l.baseDir, path)
	return filepath.Clean(absPath), nil
}

// processPlugin is a Plugin implementation that communicates with a separate process
type processPlugin struct {
	path   string
	config PluginConfig
	loader *pluginLoaderImpl

	// Plugin metadata (would be retrieved from the process via gRPC)
	name         string
	version      string
	description  string
	author       string
	dependencies []PluginDependency
	configSchema map[string]interface{}

	// Plugin context
	ctx PluginContext
}

// Name returns the plugin name
func (p *processPlugin) Name() string {
	if p.name == "" {
		// Default to the binary filename without extension
		return filepath.Base(p.path)
	}
	return p.name
}

// Version returns the plugin version
func (p *processPlugin) Version() string {
	if p.version == "" {
		return "0.0.0"
	}
	return p.version
}

// Description returns the plugin description
func (p *processPlugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *processPlugin) Author() string {
	return p.author
}

// Dependencies returns the plugin dependencies
func (p *processPlugin) Dependencies() []PluginDependency {
	return p.dependencies
}

// Initialize initializes the plugin
func (p *processPlugin) Initialize(ctx PluginContext) error {
	p.ctx = ctx

	// In a full implementation, this would:
	// 1. Spawn the plugin process
	// 2. Establish gRPC connection
	// 3. Send Initialize RPC with context and config
	// 4. Wait for response

	if p.loader.logger != nil {
		p.loader.logger.Info(fmt.Sprintf("Initializing plugin from %s", p.path))
	}

	return nil
}

// Start starts the plugin
func (p *processPlugin) Start() error {
	// In a full implementation, this would send Start RPC to the plugin process

	if p.loader.logger != nil {
		p.loader.logger.Info(fmt.Sprintf("Starting plugin %s", p.Name()))
	}

	return nil
}

// Stop stops the plugin
func (p *processPlugin) Stop() error {
	// In a full implementation, this would send Stop RPC to the plugin process

	if p.loader.logger != nil {
		p.loader.logger.Info(fmt.Sprintf("Stopping plugin %s", p.Name()))
	}

	return nil
}

// Cleanup cleans up plugin resources
func (p *processPlugin) Cleanup() error {
	// In a full implementation, this would:
	// 1. Send Cleanup RPC to the plugin process
	// 2. Close gRPC connection
	// 3. Terminate the process if still running

	if p.loader.logger != nil {
		p.loader.logger.Info(fmt.Sprintf("Cleaning up plugin %s", p.Name()))
	}

	return nil
}

// ConfigSchema returns the plugin's configuration schema
func (p *processPlugin) ConfigSchema() map[string]interface{} {
	return p.configSchema
}

// OnConfigChange handles configuration changes
func (p *processPlugin) OnConfigChange(config map[string]interface{}) error {
	// In a full implementation, this would send OnConfigChange RPC to the plugin process

	if p.loader.logger != nil {
		p.loader.logger.Info(fmt.Sprintf("Config changed for plugin %s", p.Name()))
	}

	return nil
}

// PluginConfig represents configuration for a plugin
type PluginConfig struct {
	// Enabled indicates whether the plugin is enabled.
	// Default: true
	Enabled bool

	// Path is the file path to the plugin binary.
	// Required, no default
	Path string

	// Config provides plugin-specific configuration as key-value pairs.
	// Default: empty map
	Config map[string]interface{}

	// Permissions defines what operations the plugin is allowed to perform.
	// Default: all false (secure by default)
	Permissions PluginPermissions

	// Priority determines the plugin's execution order (lower values execute first).
	// Default: 0
	Priority int
}

// ApplyDefaults applies default values to PluginConfig
func (c *PluginConfig) ApplyDefaults() {
	// Default Enabled to true if not explicitly set
	// Note: We can't distinguish between false and unset for bool,
	// so we assume false means explicitly disabled

	// Initialize Config to empty map if nil
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}

	// Initialize Permissions with all false if not set
	// This ensures secure by default behavior
	if c.Permissions.CustomPermissions == nil {
		c.Permissions.CustomPermissions = make(map[string]bool)
	}

	// Priority defaults to 0 (already the zero value, no action needed)
}
