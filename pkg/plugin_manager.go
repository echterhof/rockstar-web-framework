package pkg

import (
	"fmt"
	"sync"
	"time"
)

// PluginManager manages the lifecycle of plugins
type PluginManager interface {
	// Loading
	LoadPlugin(path string, config PluginConfig) error
	LoadPluginsFromConfig(configPath string) error
	UnloadPlugin(name string) error

	// Lifecycle
	InitializeAll() error
	StartAll() error
	StopAll() error

	// Hot reload
	ReloadPlugin(name string) error

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
}

// pluginManagerImpl is the default implementation of PluginManager
type pluginManagerImpl struct {
	mu                 sync.RWMutex
	registry           PluginRegistry
	loader             PluginLoader
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

	// Hot reload support
	requestQueues map[string]*pluginRequestQueue

	// Framework services for creating plugin contexts
	router   RouterEngine
	database DatabaseManager
	cache    CacheManager
	config   ConfigManager

	// Shared registries
	serviceRegistry    ServiceRegistry
	middlewareRegistry MiddlewareRegistry

	// Error threshold configuration
	errorThreshold int64 // Number of errors before warning/disabling
	autoDisable    bool  // Whether to auto-disable plugins exceeding threshold
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
	loader PluginLoader,
	hookSystem HookSystem,
	eventBus EventBus,
	permissionChecker PermissionChecker,
	logger Logger,
	metrics MetricsCollector,
	router RouterEngine,
	database DatabaseManager,
	cache CacheManager,
	config ConfigManager,
) PluginManager {
	return &pluginManagerImpl{
		registry:           registry,
		loader:             loader,
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
		plugins:            make(map[string]*pluginEntry),
		loadOrder:          []string{},
		serviceRegistry:    NewServiceRegistry(),
		middlewareRegistry: NewMiddlewareRegistry(),
		errorThreshold:     100,   // Default: 100 errors before action
		autoDisable:        false, // Default: don't auto-disable, just warn
	}
}

// SetErrorThreshold sets the error threshold for plugin monitoring
func (m *pluginManagerImpl) SetErrorThreshold(threshold int64, autoDisable bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorThreshold = threshold
	m.autoDisable = autoDisable
}

// LoadPlugin loads a single plugin from the specified path
func (m *pluginManagerImpl) LoadPlugin(path string, config PluginConfig) error {
	startTime := time.Now()

	// Check if enabled flag is set
	if !config.Enabled {
		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Skipping disabled plugin at %s", path))
		}
		return nil
	}

	// Load the plugin binary
	plugin, err := m.loader.Load(path, config)
	if err != nil {
		m.recordLoadMetrics(path, false, time.Since(startTime))
		return fmt.Errorf("failed to load plugin from %s: %w", path, err)
	}

	pluginName := plugin.Name()

	// Check if already loaded
	m.mu.Lock()
	if _, exists := m.plugins[pluginName]; exists {
		m.mu.Unlock()
		m.recordLoadMetrics(pluginName, false, time.Since(startTime))
		return fmt.Errorf("plugin %s is already loaded", pluginName)
	}
	m.mu.Unlock()

	// Try to load manifest if it exists
	var manifest *PluginManifest
	// Manifest would typically be in the same directory as the plugin
	// For now, we'll skip manifest loading in the basic implementation

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
		pluginName,
		plugin.Version(),
		plugin.Dependencies(),
		manifest,
	)

	// Store the entry
	m.mu.Lock()
	m.plugins[pluginName] = entry
	m.mu.Unlock()

	// Record successful load metrics
	m.recordLoadMetrics(pluginName, true, time.Since(startTime))

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Loaded plugin %s version %s", pluginName, plugin.Version()))
	}

	return nil
}

// UnloadPlugin unloads a plugin
func (m *pluginManagerImpl) UnloadPlugin(name string) error {
	m.mu.Lock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %s is not loaded", name)
	}
	m.mu.Unlock()

	// Stop the plugin if running
	if entry.status == PluginStatusRunning {
		if err := entry.plugin.Stop(); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Error stopping plugin %s: %v", name, err))
			}
		}
	}

	// Cleanup the plugin
	if err := entry.plugin.Cleanup(); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error cleaning up plugin %s: %v", name, err))
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

	// Unload from loader
	if err := m.loader.Unload(entry.plugin); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error unloading plugin %s: %v", name, err))
		}
	}

	// Unregister from registry
	if err := m.registry.Unregister(name); err != nil {
		if m.logger != nil {
			m.logger.Error(fmt.Sprintf("Error unregistering plugin %s: %v", name, err))
		}
	}

	// Remove from tracking
	m.mu.Lock()
	delete(m.plugins, name)
	m.mu.Unlock()

	// Update load order
	m.updateLoadOrder()

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Unloaded plugin %s", name))
	}

	return nil
}

// InitializeAll initializes all loaded plugins in dependency order
func (m *pluginManagerImpl) InitializeAll() error {
	// Resolve dependencies first
	if err := m.ResolveDependencies(); err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Initialize plugins in load order
	for _, name := range m.loadOrder {
		m.mu.RLock()
		entry, exists := m.plugins[name]
		m.mu.RUnlock()

		if !exists || !entry.enabled {
			continue
		}

		if err := m.initializePlugin(entry); err != nil {
			m.incrementErrorCount(name, err)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Failed to initialize plugin %s: %v", name, err))
			}
			// Continue with other plugins
			continue
		}
	}

	return nil
}

// initializePlugin initializes a single plugin
func (m *pluginManagerImpl) initializePlugin(entry *pluginEntry) error {
	pluginName := entry.plugin.Name()

	// Merge config with defaults from schema before creating context
	schema := entry.plugin.ConfigSchema()
	mergedConfig := mergeConfigWithDefaults(entry.config.Config, schema)
	entry.config.Config = mergedConfig

	// Create plugin context
	ctx := m.createPluginContext(pluginName, entry.config)

	// Initialize the plugin
	if err := entry.plugin.Initialize(ctx); err != nil {
		entry.status = PluginStatusError
		return err
	}

	// Register with registry
	if err := m.registry.Register(entry.plugin, ctx); err != nil {
		entry.status = PluginStatusError
		return err
	}

	// Store context
	entry.context = ctx
	entry.status = PluginStatusInitialized

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Initialized plugin %s", pluginName))
	}

	return nil
}

// StartAll starts all initialized plugins
func (m *pluginManagerImpl) StartAll() error {
	m.mu.RLock()
	plugins := make([]*pluginEntry, 0, len(m.plugins))
	for _, entry := range m.plugins {
		if entry.enabled && entry.status == PluginStatusInitialized {
			plugins = append(plugins, entry)
		}
	}
	m.mu.RUnlock()

	for _, entry := range plugins {
		if err := entry.plugin.Start(); err != nil {
			pluginName := entry.plugin.Name()
			m.incrementErrorCount(pluginName, err)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Failed to start plugin %s: %v", pluginName, err))
			}
			entry.status = PluginStatusError
			continue
		}

		entry.status = PluginStatusRunning

		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Started plugin %s", entry.plugin.Name()))
		}
	}

	return nil
}

// StopAll stops all running plugins
func (m *pluginManagerImpl) StopAll() error {
	m.mu.RLock()
	plugins := make([]*pluginEntry, 0, len(m.plugins))
	for _, entry := range m.plugins {
		if entry.status == PluginStatusRunning {
			plugins = append(plugins, entry)
		}
	}
	m.mu.RUnlock()

	// Stop in reverse order
	for i := len(plugins) - 1; i >= 0; i-- {
		entry := plugins[i]
		if err := entry.plugin.Stop(); err != nil {
			pluginName := entry.plugin.Name()
			m.incrementErrorCount(pluginName, err)
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Failed to stop plugin %s: %v", pluginName, err))
			}
			continue
		}

		entry.status = PluginStatusStopped

		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Stopped plugin %s", entry.plugin.Name()))
		}
	}

	return nil
}

// ReloadPlugin hot reloads a plugin with rollback on failure
func (m *pluginManagerImpl) ReloadPlugin(name string) error {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	if !exists {
		m.mu.RUnlock()
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	// Store old plugin and config for potential rollback
	oldPlugin := entry.plugin
	oldConfig := entry.config
	oldStatus := entry.status
	m.mu.RUnlock()

	// Create a request queue for this plugin during reload
	requestQueue := &pluginRequestQueue{
		pluginName: name,
		requests:   make([]interface{}, 0),
		mu:         &sync.Mutex{},
		done:       make(chan struct{}),
	}

	// Register the queue (in a real implementation, this would be used by the request handler)
	m.mu.Lock()
	if m.requestQueues == nil {
		m.requestQueues = make(map[string]*pluginRequestQueue)
	}
	m.requestQueues[name] = requestQueue
	m.mu.Unlock()

	// Ensure we always signal completion and cleanup, even on error
	defer func() {
		select {
		case <-requestQueue.done:
			// Already closed
		default:
			close(requestQueue.done)
		}
		m.cleanupRequestQueue(name)
	}()

	// Track reload status
	reloadStartTime := time.Now()

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Starting hot reload for plugin %s", name))
	}

	// Step 1: Stop the plugin
	if oldStatus == PluginStatusRunning {
		if err := oldPlugin.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin for reload: %w", err)
		}
		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Stopped plugin %s for reload", name))
		}
	}

	// Step 2: Unload the plugin
	if err := m.UnloadPlugin(name); err != nil {
		// Attempt rollback: restore the old plugin
		if rollbackErr := m.rollbackPlugin(name, oldPlugin, oldConfig, oldStatus); rollbackErr != nil {
			return fmt.Errorf("failed to unload plugin for reload and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to unload plugin for reload (rolled back): %w", err)
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Unloaded plugin %s for reload", name))
	}

	// Step 3: Load the new version
	if err := m.LoadPlugin(oldConfig.Path, oldConfig); err != nil {
		// Attempt rollback: restore the old plugin
		if rollbackErr := m.rollbackPlugin(name, oldPlugin, oldConfig, oldStatus); rollbackErr != nil {
			return fmt.Errorf("failed to load new plugin version and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to load new plugin version (rolled back): %w", err)
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Loaded new version of plugin %s", name))
	}

	// Step 4: Initialize the new plugin
	m.mu.RLock()
	newEntry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		// Attempt rollback
		if rollbackErr := m.rollbackPlugin(name, oldPlugin, oldConfig, oldStatus); rollbackErr != nil {
			return fmt.Errorf("plugin %s not found after reload and rollback failed (rollback error: %v)", name, rollbackErr)
		}
		return fmt.Errorf("plugin %s not found after reload (rolled back)", name)
	}

	if err := m.initializePlugin(newEntry); err != nil {
		// Attempt rollback
		if rollbackErr := m.rollbackPlugin(name, oldPlugin, oldConfig, oldStatus); rollbackErr != nil {
			return fmt.Errorf("failed to initialize reloaded plugin and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to initialize reloaded plugin (rolled back): %w", err)
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Initialized new version of plugin %s", name))
	}

	// Step 5: Start the new plugin
	if err := newEntry.plugin.Start(); err != nil {
		// Attempt rollback
		if rollbackErr := m.rollbackPlugin(name, oldPlugin, oldConfig, oldStatus); rollbackErr != nil {
			return fmt.Errorf("failed to start reloaded plugin and rollback failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("failed to start reloaded plugin (rolled back): %w", err)
	}

	newEntry.status = PluginStatusRunning

	// Record reload metrics
	reloadDuration := time.Since(reloadStartTime)
	if m.metrics != nil {
		m.metrics.RecordHistogram(
			"plugin.reload.duration",
			float64(reloadDuration.Milliseconds()),
			map[string]string{
				"plugin":  name,
				"success": "true",
			},
		)
		m.metrics.IncrementCounter(
			"plugin.reload.success",
			map[string]string{
				"plugin": name,
			},
		)
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Successfully reloaded plugin %s in %v", name, reloadDuration))
	}

	// Cleanup and channel close will be handled by defer
	return nil
}

// rollbackPlugin attempts to restore a plugin to its previous state after a failed reload
func (m *pluginManagerImpl) rollbackPlugin(name string, oldPlugin Plugin, oldConfig PluginConfig, oldStatus PluginStatus) error {
	if m.logger != nil {
		m.logger.Warn(fmt.Sprintf("Attempting to rollback plugin %s to previous version", name))
	}

	// Remove the failed new plugin if it exists
	m.mu.Lock()
	delete(m.plugins, name)
	m.mu.Unlock()

	// Recreate the old plugin entry
	entry := &pluginEntry{
		plugin:   oldPlugin,
		config:   oldConfig,
		status:   PluginStatusLoading,
		loadTime: time.Now(),
		enabled:  oldConfig.Enabled,
	}

	// Store the entry
	m.mu.Lock()
	m.plugins[name] = entry
	m.mu.Unlock()

	// Initialize the old plugin
	if err := m.initializePlugin(entry); err != nil {
		return fmt.Errorf("failed to initialize plugin during rollback: %w", err)
	}

	// If it was running before, start it again
	if oldStatus == PluginStatusRunning {
		if err := oldPlugin.Start(); err != nil {
			return fmt.Errorf("failed to start plugin during rollback: %w", err)
		}
		entry.status = PluginStatusRunning
	}

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Successfully rolled back plugin %s to previous version", name))
	}

	// Record rollback metrics
	if m.metrics != nil {
		m.metrics.IncrementCounter(
			"plugin.reload.rollback",
			map[string]string{
				"plugin": name,
			},
		)
	}

	return nil
}

// cleanupRequestQueue removes the request queue for a plugin
func (m *pluginManagerImpl) cleanupRequestQueue(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.requestQueues != nil {
		delete(m.requestQueues, name)
	}
}

// pluginRequestQueue holds requests for a plugin during reload
type pluginRequestQueue struct {
	pluginName string
	requests   []interface{}
	mu         *sync.Mutex
	done       chan struct{}
}

// IsReloading checks if a plugin is currently being reloaded
func (m *pluginManagerImpl) IsReloading(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.requestQueues == nil {
		return false
	}

	_, exists := m.requestQueues[name]
	return exists
}

// WaitForReload waits for a plugin reload to complete
func (m *pluginManagerImpl) WaitForReload(name string) error {
	m.mu.RLock()
	queue, exists := m.requestQueues[name]
	m.mu.RUnlock()

	if !exists {
		return nil // Not reloading
	}

	// Wait for reload to complete
	<-queue.done
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

			// Disable the plugin
			m.mu.Lock()
			if entry, exists := m.plugins[pluginName]; exists {
				entry.enabled = false
				entry.status = PluginStatusError
			}
			m.mu.Unlock()

			// Stop the plugin if it's running
			go func() {
				if err := m.stopPlugin(pluginName); err != nil {
					if m.logger != nil {
						m.logger.Error(fmt.Sprintf("Failed to stop plugin %s after threshold exceeded: %v", pluginName, err))
					}
				}
			}()
		}
	}
}

// stopPlugin stops a single plugin
func (m *pluginManagerImpl) stopPlugin(name string) error {
	m.mu.RLock()
	entry, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if entry.status != PluginStatusRunning {
		return nil // Already stopped
	}

	if err := entry.plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	m.mu.Lock()
	entry.status = PluginStatusStopped
	m.mu.Unlock()

	if m.logger != nil {
		m.logger.Info(fmt.Sprintf("Stopped plugin %s", name))
	}

	return nil
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
