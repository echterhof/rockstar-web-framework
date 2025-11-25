package pkg

import (
	"fmt"
	"sync"
)

// pluginContextImpl is the default implementation of PluginContext
type pluginContextImpl struct {
	pluginName string

	// Framework services
	router   RouterEngine
	logger   Logger
	metrics  MetricsCollector
	database DatabaseManager
	cache    CacheManager
	config   ConfigManager

	// Plugin-specific data
	pluginConfig  map[string]interface{}
	pluginStorage PluginStorage

	// Systems
	hookSystem         HookSystem
	eventBus           EventBus
	serviceRegistry    ServiceRegistry
	middlewareRegistry MiddlewareRegistry

	// Permissions
	permissions       PluginPermissions
	permissionChecker PermissionChecker
}

// NewPluginContext creates a new plugin context
func NewPluginContext(
	pluginName string,
	router RouterEngine,
	logger Logger,
	metrics MetricsCollector,
	database DatabaseManager,
	cache CacheManager,
	config ConfigManager,
	pluginConfig map[string]interface{},
	permissions PluginPermissions,
	hookSystem HookSystem,
	eventBus EventBus,
	serviceRegistry ServiceRegistry,
	middlewareRegistry MiddlewareRegistry,
	permissionChecker PermissionChecker,
) PluginContext {
	storage := NewPluginStorage(pluginName, database)

	return &pluginContextImpl{
		pluginName:         pluginName,
		router:             router,
		logger:             logger,
		metrics:            metrics,
		database:           database,
		cache:              cache,
		config:             config,
		pluginConfig:       pluginConfig,
		pluginStorage:      storage,
		hookSystem:         hookSystem,
		eventBus:           eventBus,
		serviceRegistry:    serviceRegistry,
		middlewareRegistry: middlewareRegistry,
		permissions:        permissions,
		permissionChecker:  permissionChecker,
	}
}

// Router returns the framework's router (permission-controlled)
func (c *pluginContextImpl) Router() RouterEngine {
	if c.permissionChecker != nil {
		if err := c.permissionChecker.CheckPermission(c.pluginName, "router"); err != nil {
			c.logger.Error(fmt.Sprintf("Plugin %s denied router access: %v", c.pluginName, err))
			return nil
		}
	}
	return c.router
}

// Logger returns the framework's logger
func (c *pluginContextImpl) Logger() Logger {
	return c.logger
}

// Metrics returns the framework's metrics collector
func (c *pluginContextImpl) Metrics() MetricsCollector {
	return c.metrics
}

// Database returns the framework's database manager (permission-controlled)
func (c *pluginContextImpl) Database() DatabaseManager {
	if c.permissionChecker != nil {
		if err := c.permissionChecker.CheckPermission(c.pluginName, "database"); err != nil {
			c.logger.Error(fmt.Sprintf("Plugin %s denied database access: %v", c.pluginName, err))
			return nil
		}
	}
	return c.database
}

// Cache returns the framework's cache manager (permission-controlled)
func (c *pluginContextImpl) Cache() CacheManager {
	if c.permissionChecker != nil {
		if err := c.permissionChecker.CheckPermission(c.pluginName, "cache"); err != nil {
			c.logger.Error(fmt.Sprintf("Plugin %s denied cache access: %v", c.pluginName, err))
			return nil
		}
	}
	return c.cache
}

// Config returns the framework's config manager (permission-controlled)
func (c *pluginContextImpl) Config() ConfigManager {
	if c.permissionChecker != nil {
		if err := c.permissionChecker.CheckPermission(c.pluginName, "config"); err != nil {
			c.logger.Error(fmt.Sprintf("Plugin %s denied config access: %v", c.pluginName, err))
			return nil
		}
	}
	return c.config
}

// PluginConfig returns the plugin-specific configuration
func (c *pluginContextImpl) PluginConfig() map[string]interface{} {
	return c.pluginConfig
}

// PluginStorage returns the plugin's isolated storage
func (c *pluginContextImpl) PluginStorage() PluginStorage {
	return c.pluginStorage
}

// RegisterHook registers a hook with the hook system
func (c *pluginContextImpl) RegisterHook(hookType HookType, priority int, handler HookHandler) error {
	if c.hookSystem == nil {
		return fmt.Errorf("hook system not available")
	}
	return c.hookSystem.RegisterHook(c.pluginName, hookType, priority, handler)
}

// PublishEvent publishes an event to the event bus
func (c *pluginContextImpl) PublishEvent(event string, data interface{}) error {
	if c.eventBus == nil {
		return fmt.Errorf("event bus not available")
	}
	return c.eventBus.Publish(event, data)
}

// SubscribeEvent subscribes to an event on the event bus
func (c *pluginContextImpl) SubscribeEvent(event string, handler EventHandler) error {
	if c.eventBus == nil {
		return fmt.Errorf("event bus not available")
	}
	return c.eventBus.Subscribe(c.pluginName, event, handler)
}

// ExportService exports a service for other plugins to use
func (c *pluginContextImpl) ExportService(name string, service interface{}) error {
	if c.serviceRegistry == nil {
		return fmt.Errorf("service registry not available")
	}
	return c.serviceRegistry.Export(c.pluginName, name, service)
}

// ImportService imports a service from another plugin
func (c *pluginContextImpl) ImportService(pluginName, serviceName string) (interface{}, error) {
	if c.serviceRegistry == nil {
		return nil, fmt.Errorf("service registry not available")
	}
	return c.serviceRegistry.Import(pluginName, serviceName)
}

// RegisterMiddleware registers middleware with the framework
func (c *pluginContextImpl) RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error {
	if c.middlewareRegistry == nil {
		return fmt.Errorf("middleware registry not available")
	}

	// Check router permission for middleware registration
	if c.permissionChecker != nil {
		if err := c.permissionChecker.CheckPermission(c.pluginName, "router"); err != nil {
			return fmt.Errorf("permission denied: %w", err)
		}
	}

	return c.middlewareRegistry.Register(c.pluginName, name, handler, priority, routes)
}

// UnregisterMiddleware unregisters middleware from the framework
func (c *pluginContextImpl) UnregisterMiddleware(name string) error {
	if c.middlewareRegistry == nil {
		return fmt.Errorf("middleware registry not available")
	}
	return c.middlewareRegistry.Unregister(c.pluginName, name)
}

// Note: PluginStorage implementation has been moved to plugin_storage.go
// to provide database-backed persistent storage with proper isolation

// ServiceRegistry manages exported services from plugins
type ServiceRegistry interface {
	Export(pluginName, serviceName string, service interface{}) error
	Import(pluginName, serviceName string) (interface{}, error)
	List(pluginName string) []string
	Unregister(pluginName, serviceName string) error
}

// serviceRegistryImpl is the default implementation of ServiceRegistry
type serviceRegistryImpl struct {
	mu       sync.RWMutex
	services map[string]map[string]interface{} // pluginName -> serviceName -> service
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() ServiceRegistry {
	return &serviceRegistryImpl{
		services: make(map[string]map[string]interface{}),
	}
}

// Export registers a service from a plugin
func (r *serviceRegistryImpl) Export(pluginName, serviceName string, service interface{}) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[pluginName]; !exists {
		r.services[pluginName] = make(map[string]interface{})
	}

	if _, exists := r.services[pluginName][serviceName]; exists {
		return fmt.Errorf("service %s already exported by plugin %s", serviceName, pluginName)
	}

	r.services[pluginName][serviceName] = service
	return nil
}

// Import retrieves a service from a plugin
func (r *serviceRegistryImpl) Import(pluginName, serviceName string) (interface{}, error) {
	if pluginName == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	pluginServices, exists := r.services[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s has no exported services", pluginName)
	}

	service, exists := pluginServices[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found in plugin %s", serviceName, pluginName)
	}

	return service, nil
}

// List returns all service names exported by a plugin
func (r *serviceRegistryImpl) List(pluginName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pluginServices, exists := r.services[pluginName]
	if !exists {
		return []string{}
	}

	names := make([]string, 0, len(pluginServices))
	for name := range pluginServices {
		names = append(names, name)
	}

	return names
}

// Unregister removes a service from a plugin
func (r *serviceRegistryImpl) Unregister(pluginName, serviceName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	pluginServices, exists := r.services[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s has no exported services", pluginName)
	}

	delete(pluginServices, serviceName)

	if len(pluginServices) == 0 {
		delete(r.services, pluginName)
	}

	return nil
}

// MiddlewareRegistry manages middleware registered by plugins
type MiddlewareRegistry interface {
	Register(pluginName, name string, handler MiddlewareFunc, priority int, routes []string) error
	Unregister(pluginName, name string) error
	List(pluginName string) []MiddlewareRegistration
	UnregisterAll(pluginName string) error
}

// MiddlewareRegistration represents a registered middleware
type MiddlewareRegistration struct {
	PluginName string
	Name       string
	Handler    MiddlewareFunc
	Priority   int
	Routes     []string
}

// middlewareRegistryImpl is the default implementation of MiddlewareRegistry
type middlewareRegistryImpl struct {
	mu         sync.RWMutex
	middleware map[string][]MiddlewareRegistration // pluginName -> registrations
}

// NewMiddlewareRegistry creates a new middleware registry
func NewMiddlewareRegistry() MiddlewareRegistry {
	return &middlewareRegistryImpl{
		middleware: make(map[string][]MiddlewareRegistration),
	}
}

// Register adds middleware from a plugin
func (r *middlewareRegistryImpl) Register(pluginName, name string, handler MiddlewareFunc, priority int, routes []string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("middleware handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registration := MiddlewareRegistration{
		PluginName: pluginName,
		Name:       name,
		Handler:    handler,
		Priority:   priority,
		Routes:     routes,
	}

	r.middleware[pluginName] = append(r.middleware[pluginName], registration)
	return nil
}

// Unregister removes a specific middleware from a plugin
func (r *middlewareRegistryImpl) Unregister(pluginName, name string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	registrations, exists := r.middleware[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s has no registered middleware", pluginName)
	}

	filtered := make([]MiddlewareRegistration, 0, len(registrations))
	found := false
	for _, reg := range registrations {
		if reg.Name != name {
			filtered = append(filtered, reg)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("middleware %s not found for plugin %s", name, pluginName)
	}

	r.middleware[pluginName] = filtered

	if len(filtered) == 0 {
		delete(r.middleware, pluginName)
	}

	return nil
}

// List returns all middleware registered by a plugin
func (r *middlewareRegistryImpl) List(pluginName string) []MiddlewareRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registrations, exists := r.middleware[pluginName]
	if !exists {
		return []MiddlewareRegistration{}
	}

	// Return a copy to prevent external modification
	result := make([]MiddlewareRegistration, len(registrations))
	copy(result, registrations)
	return result
}

// UnregisterAll removes all middleware from a plugin
func (r *middlewareRegistryImpl) UnregisterAll(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.middleware, pluginName)
	return nil
}

// HookSystem manages lifecycle hooks for plugins
type HookSystem interface {
	RegisterHook(pluginName string, hookType HookType, priority int, handler HookHandler) error
	UnregisterHook(pluginName string, hookType HookType) error
	ExecuteHooks(hookType HookType, ctx Context) error
	ListHooks(hookType HookType) []HookRegistration
}

// HookRegistration represents a registered hook
type HookRegistration struct {
	PluginName string
	HookType   HookType
	Priority   int
	Handler    HookHandler
}
