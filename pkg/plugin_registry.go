package pkg

import (
	"fmt"
	"sync"
)

// PluginRegistry tracks loaded plugins and their contexts
type PluginRegistry interface {
	Register(plugin Plugin, context PluginContext) error
	Unregister(name string) error
	Get(name string) (Plugin, PluginContext, error)
	List() []string
	Exists(name string) bool
	GetContext(name string) (PluginContext, error)
}

// pluginRegistryImpl is the default implementation of PluginRegistry
type pluginRegistryImpl struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	contexts map[string]PluginContext
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() PluginRegistry {
	return &pluginRegistryImpl{
		plugins:  make(map[string]Plugin),
		contexts: make(map[string]PluginContext),
	}
}

// Register adds a plugin and its context to the registry
func (r *pluginRegistryImpl) Register(plugin Plugin, context PluginContext) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}
	if context == nil {
		return fmt.Errorf("plugin context cannot be nil")
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

	r.plugins[name] = plugin
	r.contexts[name] = context

	return nil
}

// Unregister removes a plugin from the registry
func (r *pluginRegistryImpl) Unregister(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s is not registered", name)
	}

	delete(r.plugins, name)
	delete(r.contexts, name)

	return nil
}

// Get retrieves a plugin and its context from the registry
func (r *pluginRegistryImpl) Get(name string) (Plugin, PluginContext, error) {
	if name == "" {
		return nil, nil, fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, nil, fmt.Errorf("plugin %s not found", name)
	}

	context := r.contexts[name]
	return plugin, context, nil
}

// List returns the names of all registered plugins
func (r *pluginRegistryImpl) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}

	return names
}

// Exists checks if a plugin is registered
func (r *pluginRegistryImpl) Exists(name string) bool {
	if name == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[name]
	return exists
}

// GetContext retrieves the context for a registered plugin
func (r *pluginRegistryImpl) GetContext(name string) (PluginContext, error) {
	if name == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	context, exists := r.contexts[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return context, nil
}
