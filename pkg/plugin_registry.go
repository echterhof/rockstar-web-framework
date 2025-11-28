package pkg

import (
	"fmt"
	"sync"
)

// PluginFactory is a function that creates a new instance of a plugin
type PluginFactory func() Plugin

// PluginRegistry is the runtime registry interface for managing loaded plugins
type PluginRegistry interface {
	Register(plugin Plugin, ctx PluginContext) error
	Unregister(name string) error
	Get(name string) (Plugin, PluginContext, error)
	GetContext(name string) (PluginContext, error)
	List() []PluginInfo
}

// CompileTimeRegistry manages compile-time plugin registration
type CompileTimeRegistry struct {
	factories map[string]PluginFactory
	mu        sync.RWMutex
}

// newCompileTimeRegistry creates a new compile-time registry
func newCompileTimeRegistry() *CompileTimeRegistry {
	return &CompileTimeRegistry{
		factories: make(map[string]PluginFactory),
	}
}

// Register registers a plugin factory with the given name
func (r *CompileTimeRegistry) Register(name string, factory PluginFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// GetFactory retrieves a plugin factory by name
func (r *CompileTimeRegistry) GetFactory(name string) (PluginFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.factories[name]
	return factory, ok
}

// ListPlugins returns a list of all registered plugin names
func (r *CompileTimeRegistry) ListPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Unregister removes a plugin factory from the registry
// This is primarily for testing purposes
func (r *CompileTimeRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, name)
}

// Global registry for compile-time plugin registration
var globalRegistry = newCompileTimeRegistry()

// RegisterPlugin registers a plugin factory at compile time
// This function should be called from plugin init() functions
func RegisterPlugin(name string, factory PluginFactory) {
	globalRegistry.Register(name, factory)
}

// GetRegisteredPlugins returns a list of all registered plugin names
func GetRegisteredPlugins() []string {
	return globalRegistry.ListPlugins()
}

// CreatePlugin creates a new instance of a plugin by name
func CreatePlugin(name string) (Plugin, error) {
	factory, ok := globalRegistry.GetFactory(name)
	if !ok {
		return nil, fmt.Errorf("plugin %s not registered", name)
	}
	return factory(), nil
}

// pluginRegistryImpl is the default implementation of PluginRegistry
type pluginRegistryImpl struct {
	mu      sync.RWMutex
	plugins map[string]*registryEntry
}

// registryEntry stores a plugin and its context
type registryEntry struct {
	plugin Plugin
	ctx    PluginContext
}

// NewPluginRegistry creates a new runtime plugin registry
func NewPluginRegistry() PluginRegistry {
	return &pluginRegistryImpl{
		plugins: make(map[string]*registryEntry),
	}
}

// Register registers a plugin with its context
func (r *pluginRegistryImpl) Register(plugin Plugin, ctx PluginContext) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already registered", name)
	}

	r.plugins[name] = &registryEntry{
		plugin: plugin,
		ctx:    ctx,
	}

	return nil
}

// Unregister removes a plugin from the registry
func (r *pluginRegistryImpl) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s is not registered", name)
	}

	delete(r.plugins, name)
	return nil
}

// Get retrieves a plugin and its context by name
func (r *pluginRegistryImpl) Get(name string) (Plugin, PluginContext, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return nil, nil, fmt.Errorf("plugin %s not found", name)
	}

	return entry.plugin, entry.ctx, nil
}

// GetContext retrieves a plugin's context by name
func (r *pluginRegistryImpl) GetContext(name string) (PluginContext, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return entry.ctx, nil
}

// List returns information about all registered plugins
func (r *pluginRegistryImpl) List() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(r.plugins))
	for _, entry := range r.plugins {
		infos = append(infos, PluginInfo{
			Name:        entry.plugin.Name(),
			Version:     entry.plugin.Version(),
			Description: entry.plugin.Description(),
			Author:      entry.plugin.Author(),
			Loaded:      true,
			Enabled:     true,
			Status:      PluginStatusRunning,
		})
	}

	return infos
}
