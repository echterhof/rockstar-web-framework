package pkg

import (
	"fmt"
	"sync"
	"time"
)

// PluginManager manages the lifecycle of plugins
type PluginManager interface {
	// Discovery
	DiscoverPlugins() error
	LoadPluginsFromConfig(configPath string) error

	// Lifecycle
	InitializeAll() error
	StartAll() error
	StopAll() error

	// Individual plugin management
	InitializePlugin(name string) error
	StartPlugin(name string) error
	StopPlugin(name string) error
	DisablePlugin(name string) error

	// Query
	GetPlugin(name string) (Plugin, error)
	ListPlugins() []PluginInfo
	IsLoaded(name string) bool

	// Dependency management
	ResolveDependencies() error
	GetDependencyGraph() map[string][]string

	// Health
	GetPluginHealth(name string) PluginHealth
	GetAllHealth() map[string]PluginHealth

	// Configuration management
	UpdatePluginConfig(name string, config map[string]interface{}) error
	GetPluginConfig(name string) (map[string]interface{}, error)

	// Database initialization
	InitializeDatabase() error

	// Metrics and monitoring
	GetPluginMetrics(name string) *PluginMetrics
	GetAllPluginMetrics() map[string]*PluginMetrics
	ExportPrometheusMetrics() string
}

// pluginManagerImpl is the default implementation of PluginManager
type pluginManagerImpl struct {
	mu                 sync.RWMutex
	registry           PluginRegistry
	manifestParser     *ManifestParser
	dependencyResolver *DependencyResolver
	hookSystem         HookSystem
	eventBus           EventBus
	permissionChecker  PermissionChecker
	logger             Logger
	metrics            MetricsCollector

	// Plugin tracking
	plugins   map[string]*pluginEntry
	loadOrder []string

	// Framework services for creating plugin contexts
	router     RouterEngine
	database   DatabaseManager
	cache      CacheManager
	config     ConfigManager
	fileSystem FileManager
	network    NetworkClient

	// Shared registries
	serviceRegistry    ServiceRegistry
	middlewareRegistry MiddlewareRegistry

	// Error threshold configuration
	errorThreshold int64 // Number of errors before warning/disabling
	autoDisable    bool  // Whether to auto-disable plugins exceeding threshold

	// Plugin configurations from framework config
	pluginConfigs map[string]PluginConfig

	// Plugin metrics collection
	pluginMetrics *PluginMetricsCollector
}

// pluginEntry tracks a loaded plugin and its metadata
type pluginEntry struct {
	plugin      Plugin
	context     PluginContext
	config      PluginConfig
	manifest    *PluginManifest
	status      PluginStatus
	loadTime    time.Time
	errorCount  int64
	lastError   error
	lastErrorAt time.Time
	enabled     bool
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(
	registry PluginRegistry,
	hookSystem HookSystem,
	eventBus EventBus,
	permissionChecker PermissionChecker,
	logger Logger,
	metrics MetricsCollector,
	router RouterEngine,
	database DatabaseManager,
	cache CacheManager,
	config ConfigManager,
	fileSystem FileManager,
	network NetworkClient,
) PluginManager {
	serviceRegistry := NewServiceRegistry()

	pm := &pluginManagerImpl{
		registry:           registry,
		manifestParser:     NewManifestParser(),
		dependencyResolver: NewDependencyResolver(),
		hookSystem:         hookSystem,
		eventBus:           eventBus,
		permissionChecker:  permissionChecker,
		logger:             logger,
		metrics:            metrics,
		router:             router,
		database:           database,
		cache:              cache,
		config:             config,
		fileSystem:         fileSystem,
		network:            network,
		plugins:            make(map[string]*pluginEntry),
		loadOrder:          []string{},
		serviceRegistry:    serviceRegistry,
		middlewareRegistry: NewMiddlewareRegistry(),
		errorThreshold:     100,   // Default: 100 errors before action
		autoDisable:        false, // Default: don't auto-disable, just warn
		pluginConfigs:      make(map[string]PluginConfig),
		pluginMetrics:      NewPluginMetricsCollector(),
	}

	// Set the plugin manager as the status checker for the service registry
	serviceRegistry.SetPluginStatusChecker(pm)

	return pm
}

// SetErrorThreshold sets the error threshold for plugin monitoring
func (m *pluginManagerImpl) SetErrorThreshold(threshold int64, autoDisable bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorThreshold = threshold
	m.autoDisable = autoDisable
}

// DiscoverPlugins discovers all registered compile-time plugins
func (m *pluginManagerImpl) DiscoverPlugins() error {
	startTime := time.Now()

	// Get all registered plugin names from the compile-time registry
	pluginNames := GetRegisteredPlugins()

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Discovered %d registered plugins", len(pluginNames)))
	}

	// Create plugin instances for each registered plugin
	for _, name := range pluginNames {
		// Check if we have a config for this plugin
		config, hasConfig := m.pluginConfigs[name]
		if !hasConfig {
			// Use default config if not specified
			config = PluginConfig{
				Enabled: true,
				Config:  make(map[string]interface{}),
			}
			config.ApplyDefaults()
		}

		// Skip disabled plugins
		if !config.Enabled {
			if m.logger != nil {
				m.logger.Info(fmt.Sprintf("Skipping disabled plugin %s", name))
			}
			continue
		}

		// Create plugin instance
		plugin, err := CreatePlugin(name)
		if err != nil {
			m.recordLoadMetrics(name, false, time.Since(startTime))
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Failed to create plugin %s: %v", name, err))
			}
			continue
		}

		// Check if already discovered
		m.mu.Lock()
		if _, exists := m.plugins[name]; exists {
			m.mu.Unlock()
			m.recordLoadMetrics(name, false, time.Since(startTime))
			if m.logger != nil {
				m.logger.Warn(fmt.Sprintf("Plugin %s already discovered", name))
			}
			continue
		}
		m.mu.Unlock()

		// Try to load manifest if it exists
		var manifest *PluginManifest
		// Manifest loading would be implemented here if needed

		// Create plugin entry
		entry := &pluginEntry{
			plugin:   plugin,
			config:   config,
			manifest: manifest,
			status:   PluginStatusLoading,
			loadTime: time.Now(),
			enabled:  config.Enabled,
		}

		// Add to dependency resolver
		m.dependencyResolver.AddPlugin(
			name,
			plugin.Version(),
			plugin.Dependencies(),
			manifest,
		)

		// Store the entry
		m.mu.Lock()
		m.plugins[name] = entry
		m.mu.Unlock()

		// Record successful discovery metrics
		m.recordLoadMetrics(name, true, time.Since(startTime))

		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Discovered plugin %s version %s", name, plugin.Version()))
		}
	}

	return nil
}

// SetPluginConfig sets the configuration for a plugin before discovery
func (m *pluginManagerImpl) SetPluginConfig(name string, config PluginConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginConfigs[name] = config
}

// InitializeAll initializes all loaded plugins in dependency order
func (m *pluginManagerImpl) InitializeAll() error {
	// Resolve dependencies first
	if err := m.ResolveDependencies(); err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Initialize plugins in load order
	for _, name := range m.loadOrder {
		if err := m.InitializePlugin(name); err != nil {
			// Error already logged and handled in InitializePlugin
			continue
		}
	}

	return nil
}

// InitializePlugin initializes a single plugin by name
func (m *pluginManagerImpl) InitializePlugin(name string) error {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !entry.enabled {
		return fmt.Errorf("plugin %s is disabled", name)
	}

	return m.safeInitializePlugin(entry)
}

// safeInitializePlugin initializes a plugin with panic recovery
func (m *pluginManagerImpl) safeInitializePlugin(entry *pluginEntry) (err error) {
	pluginName := entry.plugin.Name()
	startTime := time.Now()

	// Get or create metrics for this plugin
	pluginMetrics := m.pluginMetrics.GetOrCreate(pluginName)

	// Recover from panics
	defer func() {
		duration := time.Since(startTime)
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked during initialization: %v", r)
			entry.status = PluginStatusError
			m.incrementErrorCount(pluginName, err)
			pluginMetrics.RecordInit(duration, err)
			m.DisablePlugin(pluginName)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Plugin %s panicked during initialization: %v", pluginName, r))
			}
		} else {
			pluginMetrics.RecordInit(duration, err)
		}
	}()

	// Merge config with defaults from schema before creating context
	schema := entry.plugin.ConfigSchema()
	mergedConfig := mergeConfigWithDefaults(entry.config.Config, schema)
	entry.config.Config = mergedConfig

	// Create plugin context
	ctx := m.createPluginContext(pluginName, entry.config)

	// Initialize the plugin
	if err := entry.plugin.Initialize(ctx); err != nil {
		entry.status = PluginStatusError
		m.incrementErrorCount(pluginName, err)
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Failed to initialize plugin %s: %v", pluginName, err))
		}
		return err
	}

	// Register with registry
	if err := m.registry.Register(entry.plugin, ctx); err != nil {
		entry.status = PluginStatusError
		m.incrementErrorCount(pluginName, err)
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Failed to register plugin %s: %v", pluginName, err))
		}
		return err
	}

	// Store context
	entry.context = ctx
	entry.status = PluginStatusInitialized

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Initialized plugin %s in %v", pluginName, time.Since(startTime)))
	}

	return nil
}

// StartAll starts all initialized plugins in dependency order
func (m *pluginManagerImpl) StartAll() error {
	// Start plugins in load order (dependency order)
	for _, name := range m.loadOrder {
		if err := m.StartPlugin(name); err != nil {
			// Error already logged and handled in StartPlugin
			continue
		}
	}

	return nil
}

// StartPlugin starts a single plugin by name
func (m *pluginManagerImpl) StartPlugin(name string) error {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !entry.enabled {
		return fmt.Errorf("plugin %s is disabled", name)
	}

	if entry.status != PluginStatusInitialized {
		return fmt.Errorf("plugin %s is not initialized (status: %s)", name, entry.status)
	}

	return m.safeStartPlugin(entry)
}

// safeStartPlugin starts a plugin with panic recovery
func (m *pluginManagerImpl) safeStartPlugin(entry *pluginEntry) (err error) {
	pluginName := entry.plugin.Name()
	startTime := time.Now()

	// Get or create metrics for this plugin
	pluginMetrics := m.pluginMetrics.GetOrCreate(pluginName)

	// Recover from panics
	defer func() {
		duration := time.Since(startTime)
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked during start: %v", r)
			entry.status = PluginStatusError
			m.incrementErrorCount(pluginName, err)
			pluginMetrics.RecordStart(duration, err)
			m.DisablePlugin(pluginName)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Plugin %s panicked during start: %v", pluginName, r))
			}
		} else {
			pluginMetrics.RecordStart(duration, err)
		}
	}()

	if err := entry.plugin.Start(); err != nil {
		m.incrementErrorCount(pluginName, err)
		entry.status = PluginStatusError
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Failed to start plugin %s: %v", pluginName, err))
		}
		return err
	}

	entry.status = PluginStatusRunning

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Started plugin %s in %v", pluginName, time.Since(startTime)))
	}

	return nil
}

// StopAll stops all running plugins in reverse dependency order
func (m *pluginManagerImpl) StopAll() error {
	// Stop plugins in reverse load order (reverse dependency order)
	for i := len(m.loadOrder) - 1; i >= 0; i-- {
		name := m.loadOrder[i]
		if err := m.StopPlugin(name); err != nil {
			// Error already logged and handled in StopPlugin
			continue
		}
	}

	return nil
}

// StopPlugin stops a single plugin by name
func (m *pluginManagerImpl) StopPlugin(name string) error {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if entry.status != PluginStatusRunning {
		// Already stopped or not running
		return nil
	}

	return m.safeStopPlugin(entry)
}

// safeStopPlugin stops a plugin with panic recovery
func (m *pluginManagerImpl) safeStopPlugin(entry *pluginEntry) (err error) {
	pluginName := entry.plugin.Name()

	// Get or create metrics for this plugin
	pluginMetrics := m.pluginMetrics.GetOrCreate(pluginName)

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked during stop: %v", r)
			entry.status = PluginStatusError
			m.incrementErrorCount(pluginName, err)
			pluginMetrics.RecordStop(err)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Plugin %s panicked during stop: %v", pluginName, r))
			}
		} else {
			pluginMetrics.RecordStop(err)
		}
	}()

	if err := entry.plugin.Stop(); err != nil {
		m.incrementErrorCount(pluginName, err)
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Failed to stop plugin %s: %v", pluginName, err))
		}
		return err
	}

	entry.status = PluginStatusStopped

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Stopped plugin %s", pluginName))
	}

	return nil
}

// DisablePlugin disables a plugin and cleans up its resources
func (m *pluginManagerImpl) DisablePlugin(name string) error {
	m.mu.Lock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %s not found", name)
	}

	// Mark as disabled
	entry.enabled = false
	entry.status = PluginStatusError
	m.mu.Unlock()

	// Stop the plugin if running
	if entry.status == PluginStatusRunning {
		if err := m.safeStopPlugin(entry); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error stopping plugin %s during disable: %v", name, err))
			}
		}
	}

	// Cleanup the plugin
	if err := m.safeCleanupPlugin(entry); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error cleaning up plugin %s during disable: %v", name, err))
		}
	}

	// Clean up middleware registered by this plugin
	if m.middlewareRegistry != nil {
		if err := m.middlewareRegistry.UnregisterAll(name); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error unregistering middleware for plugin %s: %v", name, err))
			}
		}
	}

	// Clean up services exported by this plugin
	if m.serviceRegistry != nil {
		if err := m.serviceRegistry.UnregisterAll(name); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error unregistering services for plugin %s: %v", name, err))
			}
		}
	}

	// Unregister hooks
	if m.hookSystem != nil {
		if err := m.hookSystem.UnregisterAll(name); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error unregistering hooks for plugin %s: %v", name, err))
			}
		}
	}

	// Unsubscribe from events
	if m.eventBus != nil {
		if err := m.eventBus.UnregisterAll(name); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error unregistering events for plugin %s: %v", name, err))
			}
		}
	}

	// Unregister from registry
	if err := m.registry.Unregister(name); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error unregistering plugin %s: %v", name, err))
		}
	}

	if m.logger != nil {
		m.logger.Warn(fmt.Sprintf("Disabled plugin %s", name))
	}

	// Record disable metric
	if m.metrics != nil {
		m.metrics.IncrementCounter(
			"plugin.disabled",
			map[string]string{
				"plugin": name,
			},
		)
	}

	return nil
}

// safeCleanupPlugin cleans up a plugin with panic recovery
func (m *pluginManagerImpl) safeCleanupPlugin(entry *pluginEntry) (err error) {
	pluginName := entry.plugin.Name()

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked during cleanup: %v", r)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Plugin %s panicked during cleanup: %v", pluginName, r))
			}
		}
	}()

	if err := entry.plugin.Cleanup(); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error cleaning up plugin %s: %v", pluginName, err))
		}
		return err
	}

	return nil
}

// GetPlugin retrieves a loaded plugin
func (m *pluginManagerImpl) GetPlugin(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return entry.plugin, nil
}

// ListPlugins returns information about all loaded plugins
func (m *pluginManagerImpl) ListPlugins() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.plugins))
	for _, entry := range m.plugins {
		infos = append(infos, PluginInfo{
			Name:        entry.plugin.Name(),
			Version:     entry.plugin.Version(),
			Description: entry.plugin.Description(),
			Author:      entry.plugin.Author(),
			Loaded:      true,
			Enabled:     entry.enabled,
			LoadTime:    entry.loadTime,
			Status:      entry.status,
		})
	}

	return infos
}

// IsLoaded checks if a plugin is loaded
func (m *pluginManagerImpl) IsLoaded(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.plugins[name]
	return exists
}

// IsPluginRunning checks if a plugin is loaded and running (implements PluginStatusChecker)
func (m *pluginManagerImpl) IsPluginRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.plugins[name]
	if !exists {
		return false
	}

	return entry.enabled && entry.status == PluginStatusRunning
}

// ResolveDependencies resolves plugin dependencies and determines load order
func (m *pluginManagerImpl) ResolveDependencies() error {
	loadOrder, err := m.dependencyResolver.ResolveDependencies()
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.loadOrder = loadOrder
	m.mu.Unlock()

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Resolved plugin load order: %v", loadOrder))
	}

	return nil
}

// GetDependencyGraph returns the dependency graph
func (m *pluginManagerImpl) GetDependencyGraph() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	graph := make(map[string][]string)
	for name, entry := range m.plugins {
		deps := make([]string, 0)
		for _, dep := range entry.plugin.Dependencies() {
			deps = append(deps, dep.Name)
		}
		graph[name] = deps
	}

	return graph
}

// GetPluginHealth returns health information for a plugin
func (m *pluginManagerImpl) GetPluginHealth(name string) PluginHealth {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return PluginHealth{
			Status: PluginStatusUnloaded,
		}
	}

	// Get hook metrics if hook system supports it
	var hookMetrics map[string]HookMetrics
	if hs, ok := m.hookSystem.(*hookSystemImpl); ok {
		hookMetrics = hs.GetHookMetrics(name)
	}

	return PluginHealth{
		Status:      entry.status,
		ErrorCount:  entry.errorCount,
		LastError:   entry.lastError,
		LastErrorAt: entry.lastErrorAt,
		HookMetrics: hookMetrics,
	}
}

// GetAllHealth returns health information for all plugins
func (m *pluginManagerImpl) GetAllHealth() map[string]PluginHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]PluginHealth)
	for name := range m.plugins {
		health[name] = m.GetPluginHealth(name)
	}

	return health
}

// Helper methods

// createPluginContext creates a plugin context for a plugin
func (m *pluginManagerImpl) createPluginContext(pluginName string, config PluginConfig) PluginContext {
	return NewPluginContext(
		pluginName,
		m.router,
		m.logger,
		m.metrics,
		m.database,
		m.cache,
		m.config,
		m.fileSystem,
		m.network,
		config.Config,
		config.Permissions,
		m.hookSystem,
		m.eventBus,
		m.serviceRegistry,
		m.middlewareRegistry,
		m.permissionChecker,
	)
}

// recordLoadMetrics records metrics for plugin loading
func (m *pluginManagerImpl) recordLoadMetrics(pluginName string, success bool, duration time.Duration) {
	if m.metrics == nil {
		return
	}

	m.metrics.RecordHistogram(
		"plugin.load.duration",
		float64(duration.Milliseconds()),
		map[string]string{
			"plugin":  pluginName,
			"success": fmt.Sprintf("%t", success),
		},
	)

	if success {
		m.metrics.IncrementCounter(
			"plugin.load.success",
			map[string]string{
				"plugin": pluginName,
			},
		)
	} else {
		m.metrics.IncrementCounter(
			"plugin.load.failure",
			map[string]string{
				"plugin": pluginName,
			},
		)
	}
}

// incrementErrorCount increments the error count for a plugin
func (m *pluginManagerImpl) incrementErrorCount(pluginName string, err error) {
	m.mu.Lock()

	entry, exists := m.plugins[pluginName]
	if !exists {
		m.mu.Unlock()
		return
	}

	entry.errorCount++
	entry.lastError = err
	entry.lastErrorAt = time.Now()

	errorCount := entry.errorCount
	threshold := m.errorThreshold
	autoDisable := m.autoDisable

	m.mu.Unlock()

	// Record error in plugin metrics
	pluginMetrics := m.pluginMetrics.GetOrCreate(pluginName)
	pluginMetrics.RecordError(err)

	// Record error metric
	if m.metrics != nil {
		m.metrics.IncrementCounter(
			"plugin.errors",
			map[string]string{
				"plugin": pluginName,
			},
		)
	}

	// Check error threshold
	if threshold > 0 && errorCount >= threshold {
		if m.logger != nil {
			m.logger.Warn(fmt.Sprintf(
				"Plugin %s has exceeded error threshold (%d errors, threshold: %d)",
				pluginName, errorCount, threshold,
			))
		}

		// Record threshold exceeded metric
		if m.metrics != nil {
			m.metrics.IncrementCounter(
				"plugin.threshold.exceeded",
				map[string]string{
					"plugin": pluginName,
				},
			)
		}

		// Optionally disable the plugin
		if autoDisable {
			if m.logger != nil {
				m.logger.Warn(fmt.Sprintf("Auto-disabling plugin %s due to error threshold", pluginName))
			}

			// Disable the plugin asynchronously
			go func() {
				if err := m.DisablePlugin(pluginName); err != nil {
					if m.logger != nil {
						m.logger.Error(fmt.Sprintf("Failed to disable plugin %s after threshold exceeded: %v", pluginName, err))
					}
				}
			}()
		}
	}
}

// updateLoadOrder updates the load order after plugin changes
func (m *pluginManagerImpl) updateLoadOrder() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove unloaded plugins from load order
	newOrder := make([]string, 0, len(m.loadOrder))
	for _, name := range m.loadOrder {
		if _, exists := m.plugins[name]; exists {
			newOrder = append(newOrder, name)
		}
	}
	m.loadOrder = newOrder
}

// UpdatePluginConfig updates a plugin's configuration and notifies the plugin
func (m *pluginManagerImpl) UpdatePluginConfig(name string, config map[string]interface{}) error {
	m.mu.Lock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %s not found", name)
	}
	m.mu.Unlock()

	// Get the plugin's config schema for defaults
	schema := entry.plugin.ConfigSchema()

	// Merge new config with defaults from schema
	mergedConfig := mergeConfigWithDefaults(config, schema)

	// Update the stored config
	entry.config.Config = mergedConfig

	// Update the plugin context's config
	if ctx, ok := entry.context.(*pluginContextImpl); ok {
		ctx.pluginConfig = mergedConfig
	}

	// Notify the plugin of the configuration change
	if err := entry.plugin.OnConfigChange(mergedConfig); err != nil {
		m.incrementErrorCount(name, err)
		return fmt.Errorf("plugin %s rejected config change: %w", name, err)
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Updated configuration for plugin %s", name))
	}

	return nil
}

// GetPluginConfig retrieves a plugin's current configuration
func (m *pluginManagerImpl) GetPluginConfig(name string) (map[string]interface{}, error) {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	// Return a copy to prevent external modification
	return copyConfig(entry.config.Config), nil
}

// mergeConfigWithDefaults merges user config with defaults from schema
// Supports nested configuration structures
func mergeConfigWithDefaults(config map[string]interface{}, schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		return config
	}

	result := make(map[string]interface{})

	// First, apply all defaults from schema
	for key, schemaValue := range schema {
		if schemaMap, ok := schemaValue.(map[string]interface{}); ok {
			// Check if this is a schema definition with a default value
			defaultValue, hasDefault := schemaMap["default"]
			_, hasType := schemaMap["type"]

			if hasDefault {
				result[key] = defaultValue
			}
			// Check if this is a nested object (no "type" field means it's a nested structure)
			if !hasType && !hasDefault {
				// This is a nested configuration object
				var nestedConfig map[string]interface{}
				if config != nil {
					if nc, ok := config[key].(map[string]interface{}); ok {
						nestedConfig = nc
					}
				}
				result[key] = mergeConfigWithDefaults(nestedConfig, schemaMap)
			}
		}
	}

	// Then, override with user-provided values
	if config != nil {
		for key, value := range config {
			// Handle nested structures
			if valueMap, ok := value.(map[string]interface{}); ok {
				if schemaValue, hasSchema := schema[key]; hasSchema {
					if schemaMap, ok := schemaValue.(map[string]interface{}); ok {
						// Recursively merge nested config
						result[key] = mergeConfigWithDefaults(valueMap, schemaMap)
						continue
					}
				}
			}
			// For non-nested values or values without schema, just use the provided value
			result[key] = value
		}
	}

	return result
}

// copyConfig creates a deep copy of a configuration map
func copyConfig(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := make(map[string]interface{})
	for key, value := range config {
		// Handle nested maps
		if valueMap, ok := value.(map[string]interface{}); ok {
			result[key] = copyConfig(valueMap)
		} else if valueSlice, ok := value.([]interface{}); ok {
			// Handle slices
			copiedSlice := make([]interface{}, len(valueSlice))
			copy(copiedSlice, valueSlice)
			result[key] = copiedSlice
		} else {
			// For primitive types, direct assignment is fine
			result[key] = value
		}
	}

	return result
}

// InitializeDatabase creates the plugin system database tables
func (m *pluginManagerImpl) InitializeDatabase() error {
	if m.database == nil {
		return fmt.Errorf("database manager not configured")
	}

	if m.logger != nil {
		m.logger.Info("Initializing plugin system database tables")
	}

	if err := m.database.InitializePluginTables(); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Failed to initialize plugin tables: %v", err))
		}
		return fmt.Errorf("failed to initialize plugin tables: %w", err)
	}

	if m.logger != nil {
		m.logger.Info("Plugin system database tables initialized successfully")
	}

	return nil
}

// GetPluginMetrics returns metrics for a specific plugin
func (m *pluginManagerImpl) GetPluginMetrics(name string) *PluginMetrics {
	return m.pluginMetrics.Get(name)
}

// GetAllPluginMetrics returns metrics for all plugins
func (m *pluginManagerImpl) GetAllPluginMetrics() map[string]*PluginMetrics {
	return m.pluginMetrics.GetAll()
}

// ExportPrometheusMetrics exports all plugin metrics in Prometheus format
func (m *pluginManagerImpl) ExportPrometheusMetrics() string {
	return m.pluginMetrics.ExportPrometheus()
}
