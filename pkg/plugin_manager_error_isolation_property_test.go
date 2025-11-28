package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestErrorIsolationProperty tests Property 14: Error Isolation
// Feature: compile-time-plugins, Property 14: Error Isolation
// Validates: Requirements 13.1, 13.2, 13.4
//
// For any plugin that panics or returns an error during hook execution,
// the framework should recover, disable the plugin, and continue processing other plugins.
func TestErrorIsolationProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("plugin errors don't affect other plugins", prop.ForAll(
		func(numPlugins int, errorPluginIndex int) bool {
			if numPlugins < 2 {
				return true // Need at least 2 plugins to test isolation
			}

			// Ensure errorPluginIndex is within bounds
			errorPluginIndex = errorPluginIndex % numPlugins
			if errorPluginIndex < 0 {
				errorPluginIndex = -errorPluginIndex
			}

			// Create test framework components
			logger := NewTestLogger()
			metrics := NewTestMetrics()
			registry := NewPluginRegistry()
			hookSystem := NewHookSystem(logger, metrics)
			eventBus := NewEventBus(logger)
			permissionChecker := NewPermissionChecker(logger)

			// Create plugin manager
			manager := NewPluginManager(
				registry,
				hookSystem,
				eventBus,
				permissionChecker,
				logger,
				metrics,
				nil, // router
				nil, // database
				nil, // cache
				nil, // config
				nil, // fileSystem
				nil, // network
			)

			// Register test plugins
			pluginNames := make([]string, numPlugins)
			for i := 0; i < numPlugins; i++ {
				pluginName := fmt.Sprintf("test-plugin-%d", i)
				pluginNames[i] = pluginName

				var plugin Plugin
				if i == errorPluginIndex {
					// This plugin will panic during initialization
					plugin = &PanicPlugin{name: pluginName}
				} else {
					// Normal plugin
					plugin = &TestPlugin{name: pluginName}
				}

				// Register the plugin in the compile-time registry
				RegisterPlugin(pluginName, func(p Plugin) func() Plugin {
					return func() Plugin { return p }
				}(plugin))

				// Set plugin config
				if pm, ok := manager.(*pluginManagerImpl); ok {
					pm.SetPluginConfig(pluginName, PluginConfig{
						Enabled: true,
						Config:  make(map[string]interface{}),
					})
				}
			}

			// Discover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("Failed to discover plugins: %v", err)
				return false
			}

			// Initialize all plugins (error plugin should panic and be disabled)
			_ = manager.InitializeAll()

			// Verify error plugin is disabled or in error state
			errorPluginName := pluginNames[errorPluginIndex]
			errorHealth := manager.GetPluginHealth(errorPluginName)
			if errorHealth.Status != PluginStatusError {
				t.Logf("Error plugin %s should be in error status, got: %s", errorPluginName, errorHealth.Status)
				return false
			}

			// Verify other plugins are initialized or running
			for i, pluginName := range pluginNames {
				if i == errorPluginIndex {
					continue // Skip the error plugin
				}

				health := manager.GetPluginHealth(pluginName)
				if health.Status != PluginStatusInitialized && health.Status != PluginStatusRunning {
					t.Logf("Plugin %s should be initialized or running, got: %s", pluginName, health.Status)
					return false
				}
			}

			// Clean up: unregister plugins from global registry
			for _, pluginName := range pluginNames {
				globalRegistry.Unregister(pluginName)
			}

			return true
		},
		gen.IntRange(2, 10),  // Number of plugins
		gen.IntRange(0, 100), // Error plugin index
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestErrorIsolationInStartProperty tests error isolation during Start phase
func TestErrorIsolationInStartProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("plugin start errors don't affect other plugins", prop.ForAll(
		func(numPlugins int, errorPluginIndex int) bool {
			if numPlugins < 2 {
				return true // Need at least 2 plugins to test isolation
			}

			// Ensure errorPluginIndex is within bounds
			errorPluginIndex = errorPluginIndex % numPlugins
			if errorPluginIndex < 0 {
				errorPluginIndex = -errorPluginIndex
			}

			// Create test framework components
			logger := NewTestLogger()
			metrics := NewTestMetrics()
			registry := NewPluginRegistry()
			hookSystem := NewHookSystem(logger, metrics)
			eventBus := NewEventBus(logger)
			permissionChecker := NewPermissionChecker(logger)

			// Create plugin manager
			manager := NewPluginManager(
				registry,
				hookSystem,
				eventBus,
				permissionChecker,
				logger,
				metrics,
				nil, // router
				nil, // database
				nil, // cache
				nil, // config
				nil, // fileSystem
				nil, // network
			)

			// Register test plugins
			pluginNames := make([]string, numPlugins)
			for i := 0; i < numPlugins; i++ {
				pluginName := fmt.Sprintf("test-start-plugin-%d", i)
				pluginNames[i] = pluginName

				plugin := &TestPlugin{name: pluginName}
				if i == errorPluginIndex {
					// This plugin will panic during start
					plugin.startPanic = true
				}
				RegisterPlugin(pluginName, func(p *TestPlugin) func() Plugin {
					return func() Plugin { return p }
				}(plugin))

				// Set plugin config
				if pm, ok := manager.(*pluginManagerImpl); ok {
					pm.SetPluginConfig(pluginName, PluginConfig{
						Enabled: true,
						Config:  make(map[string]interface{}),
					})
				}
			}

			// Discover and initialize plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("Failed to discover plugins: %v", err)
				return false
			}

			_ = manager.InitializeAll()

			// Start all plugins (error plugin should panic and be disabled)
			_ = manager.StartAll()

			// Verify error plugin is disabled or in error state
			errorPluginName := pluginNames[errorPluginIndex]
			errorHealth := manager.GetPluginHealth(errorPluginName)
			if errorHealth.Status != PluginStatusError {
				t.Logf("Error plugin %s should be in error status, got: %s", errorPluginName, errorHealth.Status)
				return false
			}

			// Verify other plugins are running
			for i, pluginName := range pluginNames {
				if i == errorPluginIndex {
					continue // Skip the error plugin
				}

				health := manager.GetPluginHealth(pluginName)
				if health.Status != PluginStatusRunning {
					t.Logf("Plugin %s should be running, got: %s", pluginName, health.Status)
					return false
				}
			}

			// Clean up: unregister plugins from global registry
			for _, pluginName := range pluginNames {
				globalRegistry.Unregister(pluginName)
			}

			return true
		},
		gen.IntRange(2, 10),  // Number of plugins
		gen.IntRange(0, 100), // Error plugin index
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestErrorIsolationInStopProperty tests error isolation during Stop phase
func TestErrorIsolationInStopProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("plugin stop errors don't affect other plugins", prop.ForAll(
		func(numPlugins int, errorPluginIndex int) bool {
			if numPlugins < 2 {
				return true // Need at least 2 plugins to test isolation
			}

			// Ensure errorPluginIndex is within bounds
			errorPluginIndex = errorPluginIndex % numPlugins
			if errorPluginIndex < 0 {
				errorPluginIndex = -errorPluginIndex
			}

			// Create test framework components
			logger := NewTestLogger()
			metrics := NewTestMetrics()
			registry := NewPluginRegistry()
			hookSystem := NewHookSystem(logger, metrics)
			eventBus := NewEventBus(logger)
			permissionChecker := NewPermissionChecker(logger)

			// Create plugin manager
			manager := NewPluginManager(
				registry,
				hookSystem,
				eventBus,
				permissionChecker,
				logger,
				metrics,
				nil, // router
				nil, // database
				nil, // cache
				nil, // config
				nil, // fileSystem
				nil, // network
			)

			// Register test plugins
			pluginNames := make([]string, numPlugins)
			for i := 0; i < numPlugins; i++ {
				pluginName := fmt.Sprintf("test-stop-plugin-%d", i)
				pluginNames[i] = pluginName

				plugin := &TestPlugin{name: pluginName}
				if i == errorPluginIndex {
					// This plugin will panic during stop
					plugin.stopPanic = true
				}
				RegisterPlugin(pluginName, func(p *TestPlugin) func() Plugin {
					return func() Plugin { return p }
				}(plugin))

				// Set plugin config
				if pm, ok := manager.(*pluginManagerImpl); ok {
					pm.SetPluginConfig(pluginName, PluginConfig{
						Enabled: true,
						Config:  make(map[string]interface{}),
					})
				}
			}

			// Discover, initialize, and start plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("Failed to discover plugins: %v", err)
				return false
			}

			_ = manager.InitializeAll()
			_ = manager.StartAll()

			// Stop all plugins (error plugin should panic)
			_ = manager.StopAll()

			// Verify error plugin is in error state
			errorPluginName := pluginNames[errorPluginIndex]
			errorHealth := manager.GetPluginHealth(errorPluginName)
			if errorHealth.Status != PluginStatusError {
				t.Logf("Error plugin %s should be in error status, got: %s", errorPluginName, errorHealth.Status)
				return false
			}

			// Verify other plugins are stopped
			for i, pluginName := range pluginNames {
				if i == errorPluginIndex {
					continue // Skip the error plugin
				}

				health := manager.GetPluginHealth(pluginName)
				if health.Status != PluginStatusStopped {
					t.Logf("Plugin %s should be stopped, got: %s", pluginName, health.Status)
					return false
				}
			}

			// Clean up: unregister plugins from global registry
			for _, pluginName := range pluginNames {
				globalRegistry.Unregister(pluginName)
			}

			return true
		},
		gen.IntRange(2, 10),  // Number of plugins
		gen.IntRange(0, 100), // Error plugin index
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// PanicPlugin is a test plugin that panics during initialization
type PanicPlugin struct {
	name string
}

func (p *PanicPlugin) Name() string        { return p.name }
func (p *PanicPlugin) Version() string     { return "1.0.0" }
func (p *PanicPlugin) Description() string { return "Test plugin that panics" }
func (p *PanicPlugin) Author() string      { return "Test" }

func (p *PanicPlugin) Dependencies() []PluginDependency {
	return []PluginDependency{}
}

func (p *PanicPlugin) Initialize(ctx PluginContext) error {
	panic("intentional panic during initialization")
}

func (p *PanicPlugin) Start() error {
	return nil
}

func (p *PanicPlugin) Stop() error {
	return nil
}

func (p *PanicPlugin) Cleanup() error {
	return nil
}

func (p *PanicPlugin) ConfigSchema() map[string]interface{} {
	return make(map[string]interface{})
}

func (p *PanicPlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}
