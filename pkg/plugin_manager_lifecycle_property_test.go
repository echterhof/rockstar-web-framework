package pkg

import (
	"fmt"
	"sync"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_LifecycleOrder tests Property 12:
// Lifecycle Order
// **Feature: compile-time-plugins, Property 12: Lifecycle Order**
// **Validates: Requirements 8.1, 8.2, 8.3**
func TestProperty_LifecycleOrder(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Initialize called in dependency order", prop.ForAll(
		func(pluginNames []string) bool {
			// Skip empty or single plugin cases
			if len(pluginNames) < 2 {
				return true
			}

			// Create test plugins with dependencies
			testPlugins := createTestPluginsWithDeps(pluginNames)

			// Create plugin manager
			manager := createTestPluginManager()

			// Register plugins in compile-time registry
			for _, tp := range testPlugins {
				RegisterPlugin(tp.plugin.Name(), func(p *testLifecyclePlugin) func() Plugin {
					return func() Plugin { return p }
				}(tp.plugin))
			}

			// Set plugin configs
			for _, tp := range testPlugins {
				manager.(*pluginManagerImpl).SetPluginConfig(tp.plugin.Name(), PluginConfig{
					Enabled: true,
					Config:  make(map[string]interface{}),
				})
			}

			// Discover plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("DiscoverPlugins failed: %v", err)
				return false
			}

			// Initialize all plugins
			if err := manager.InitializeAll(); err != nil {
				t.Logf("InitializeAll failed: %v", err)
				return false
			}

			// Verify initialization order
			for _, tp := range testPlugins {
				if !tp.plugin.initialized {
					continue // Skip if not initialized (might be disabled due to error)
				}

				// Check that all dependencies were initialized before this plugin
				for _, dep := range tp.plugin.Dependencies() {
					if dep.Optional {
						continue
					}

					// Find the dependency plugin
					var depPlugin *testLifecyclePlugin
					for _, other := range testPlugins {
						if other.plugin.Name() == dep.Name {
							depPlugin = other.plugin
							break
						}
					}

					if depPlugin == nil {
						continue
					}

					// Verify dependency was initialized before dependent
					if depPlugin.initOrder >= tp.plugin.initOrder {
						t.Logf("Dependency %s (order %d) should be initialized before %s (order %d)",
							dep.Name, depPlugin.initOrder, tp.plugin.Name(), tp.plugin.initOrder)
						return false
					}
				}
			}

			// Cleanup
			cleanupTestPlugins(pluginNames)

			return true
		},
		genUniquePluginNames(),
	))

	properties.Property("Start called in dependency order", prop.ForAll(
		func(pluginNames []string) bool {
			// Skip empty or single plugin cases
			if len(pluginNames) < 2 {
				return true
			}

			// Create test plugins with dependencies
			testPlugins := createTestPluginsWithDeps(pluginNames)

			// Create plugin manager
			manager := createTestPluginManager()

			// Register plugins in compile-time registry
			for _, tp := range testPlugins {
				RegisterPlugin(tp.plugin.Name(), func(p *testLifecyclePlugin) func() Plugin {
					return func() Plugin { return p }
				}(tp.plugin))
			}

			// Set plugin configs
			for _, tp := range testPlugins {
				manager.(*pluginManagerImpl).SetPluginConfig(tp.plugin.Name(), PluginConfig{
					Enabled: true,
					Config:  make(map[string]interface{}),
				})
			}

			// Discover and initialize plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("DiscoverPlugins failed: %v", err)
				return false
			}

			if err := manager.InitializeAll(); err != nil {
				t.Logf("InitializeAll failed: %v", err)
				return false
			}

			// Start all plugins
			if err := manager.StartAll(); err != nil {
				t.Logf("StartAll failed: %v", err)
				return false
			}

			// Verify start order
			for _, tp := range testPlugins {
				if !tp.plugin.started {
					continue // Skip if not started
				}

				// Check that all dependencies were started before this plugin
				for _, dep := range tp.plugin.Dependencies() {
					if dep.Optional {
						continue
					}

					// Find the dependency plugin
					var depPlugin *testLifecyclePlugin
					for _, other := range testPlugins {
						if other.plugin.Name() == dep.Name {
							depPlugin = other.plugin
							break
						}
					}

					if depPlugin == nil {
						continue
					}

					// Verify dependency was started before dependent
					if depPlugin.startOrder >= tp.plugin.startOrder {
						t.Logf("Dependency %s (order %d) should be started before %s (order %d)",
							dep.Name, depPlugin.startOrder, tp.plugin.Name(), tp.plugin.startOrder)
						return false
					}
				}
			}

			// Cleanup
			cleanupTestPlugins(pluginNames)

			return true
		},
		genUniquePluginNames(),
	))

	properties.Property("Stop called in reverse dependency order", prop.ForAll(
		func(pluginNames []string) bool {
			// Skip empty or single plugin cases
			if len(pluginNames) < 2 {
				return true
			}

			// Create test plugins with dependencies
			testPlugins := createTestPluginsWithDeps(pluginNames)

			// Create plugin manager
			manager := createTestPluginManager()

			// Register plugins in compile-time registry
			for _, tp := range testPlugins {
				RegisterPlugin(tp.plugin.Name(), func(p *testLifecyclePlugin) func() Plugin {
					return func() Plugin { return p }
				}(tp.plugin))
			}

			// Set plugin configs
			for _, tp := range testPlugins {
				manager.(*pluginManagerImpl).SetPluginConfig(tp.plugin.Name(), PluginConfig{
					Enabled: true,
					Config:  make(map[string]interface{}),
				})
			}

			// Discover, initialize, and start plugins
			if err := manager.DiscoverPlugins(); err != nil {
				t.Logf("DiscoverPlugins failed: %v", err)
				return false
			}

			if err := manager.InitializeAll(); err != nil {
				t.Logf("InitializeAll failed: %v", err)
				return false
			}

			if err := manager.StartAll(); err != nil {
				t.Logf("StartAll failed: %v", err)
				return false
			}

			// Stop all plugins
			if err := manager.StopAll(); err != nil {
				t.Logf("StopAll failed: %v", err)
				return false
			}

			// Verify stop order (reverse of start order)
			for _, tp := range testPlugins {
				if !tp.plugin.stopped {
					continue // Skip if not stopped
				}

				// Check that all dependents were stopped before this plugin
				for _, other := range testPlugins {
					if other.plugin.Name() == tp.plugin.Name() {
						continue
					}

					// Check if other depends on tp
					for _, dep := range other.plugin.Dependencies() {
						if dep.Optional {
							continue
						}

						if dep.Name == tp.plugin.Name() {
							// other depends on tp, so other should be stopped first
							if other.plugin.stopOrder >= tp.plugin.stopOrder {
								t.Logf("Dependent %s (order %d) should be stopped before dependency %s (order %d)",
									other.plugin.Name(), other.plugin.stopOrder, tp.plugin.Name(), tp.plugin.stopOrder)
								return false
							}
						}
					}
				}
			}

			// Cleanup
			cleanupTestPlugins(pluginNames)

			return true
		},
		genUniquePluginNames(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// Test helper types and functions

type testPluginWithDeps struct {
	plugin *testLifecyclePlugin
}

var (
	lifecycleOrderCounter int
	lifecycleOrderMu      sync.Mutex
)

func getNextLifecycleOrder() int {
	lifecycleOrderMu.Lock()
	defer lifecycleOrderMu.Unlock()
	lifecycleOrderCounter++
	return lifecycleOrderCounter
}

type testLifecyclePlugin struct {
	name         string
	version      string
	dependencies []PluginDependency
	initialized  bool
	started      bool
	stopped      bool
	initOrder    int
	startOrder   int
	stopOrder    int
}

func (p *testLifecyclePlugin) Name() string        { return p.name }
func (p *testLifecyclePlugin) Version() string     { return p.version }
func (p *testLifecyclePlugin) Description() string { return "Test plugin for lifecycle order" }
func (p *testLifecyclePlugin) Author() string      { return "Test" }
func (p *testLifecyclePlugin) Dependencies() []PluginDependency {
	return p.dependencies
}

func (p *testLifecyclePlugin) Initialize(ctx PluginContext) error {
	p.initialized = true
	p.initOrder = getNextLifecycleOrder()
	return nil
}

func (p *testLifecyclePlugin) Start() error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}
	p.started = true
	p.startOrder = getNextLifecycleOrder()
	return nil
}

func (p *testLifecyclePlugin) Stop() error {
	p.stopped = true
	p.stopOrder = getNextLifecycleOrder()
	return nil
}

func (p *testLifecyclePlugin) Cleanup() error {
	return nil
}

func (p *testLifecyclePlugin) ConfigSchema() map[string]interface{} {
	return make(map[string]interface{})
}

func (p *testLifecyclePlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}

// createTestPluginsWithDeps creates test plugins with dependencies
func createTestPluginsWithDeps(names []string) []*testPluginWithDeps {
	plugins := make([]*testPluginWithDeps, len(names))

	for i, name := range names {
		// Create dependencies on earlier plugins to avoid cycles
		var deps []PluginDependency
		if i > 0 {
			// Depend on previous plugin
			deps = append(deps, PluginDependency{
				Name:     names[i-1],
				Version:  ">=1.0.0",
				Optional: false,
			})
		}

		plugins[i] = &testPluginWithDeps{
			plugin: &testLifecyclePlugin{
				name:         name,
				version:      "1.0.0",
				dependencies: deps,
			},
		}
	}

	return plugins
}

// createTestPluginManager creates a plugin manager for testing
func createTestPluginManager() PluginManager {
	registry := NewPluginRegistry()
	logger := NewLogger(nil)
	metrics := NewMetricsCollector(nil) // Pass nil database
	hookSystem := NewHookSystem(logger, metrics)
	eventBus := NewEventBus(logger)
	permissionChecker := NewPermissionChecker(logger)

	return NewPluginManager(
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
}

// cleanupTestPlugins removes test plugins from the global registry
func cleanupTestPlugins(names []string) {
	// Reset the global registry for the next test
	// This is a bit hacky but necessary for property testing
	globalRegistry.mu.Lock()
	for _, name := range names {
		delete(globalRegistry.factories, name)
	}
	globalRegistry.mu.Unlock()

	// Reset lifecycle order counter
	lifecycleOrderMu.Lock()
	lifecycleOrderCounter = 0
	lifecycleOrderMu.Unlock()
}

// genUniquePluginNames generates unique plugin names
func genUniquePluginNames() gopter.Gen {
	return gen.SliceOfN(5, gen.Identifier()).
		SuchThat(func(v interface{}) bool {
			names := v.([]string)
			// Ensure unique names
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) >= 2
		})
}
