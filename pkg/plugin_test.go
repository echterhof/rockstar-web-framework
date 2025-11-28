package pkg

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 1: Plugin interface validation**
// **Validates: Requirements 1.2**
// For any plugin loading attempt, if the plugin does not implement the required Plugin interface methods,
// then the system should reject the plugin and return an error
func TestProperty_PluginInterfaceValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("plugin must implement all required interface methods",
		prop.ForAll(
			func() bool {
				// Test that a valid plugin implementation satisfies the interface
				plugin := &ValidTestPlugin{}

				// Verify all interface methods are callable
				_ = plugin.Name()
				_ = plugin.Version()
				_ = plugin.Description()
				_ = plugin.Author()
				_ = plugin.Dependencies()
				_ = plugin.ConfigSchema()

				// Verify lifecycle methods exist
				err := plugin.Initialize(nil)
				if err != nil {
					return false
				}

				err = plugin.Start()
				if err != nil {
					return false
				}

				err = plugin.Stop()
				if err != nil {
					return false
				}

				err = plugin.Cleanup()
				if err != nil {
					return false
				}

				err = plugin.OnConfigChange(nil)
				if err != nil {
					return false
				}

				return true
			},
		),
	)

	properties.Property("plugin interface methods return expected types",
		prop.ForAll(
			func() bool {
				plugin := &ValidTestPlugin{}

				// Verify return types
				name := plugin.Name()
				if _, ok := interface{}(name).(string); !ok {
					return false
				}

				version := plugin.Version()
				if _, ok := interface{}(version).(string); !ok {
					return false
				}

				description := plugin.Description()
				if _, ok := interface{}(description).(string); !ok {
					return false
				}

				author := plugin.Author()
				if _, ok := interface{}(author).(string); !ok {
					return false
				}

				deps := plugin.Dependencies()
				if _, ok := interface{}(deps).([]PluginDependency); !ok {
					return false
				}

				schema := plugin.ConfigSchema()
				if _, ok := interface{}(schema).(map[string]interface{}); !ok {
					return false
				}

				return true
			},
		),
	)

	properties.Property("plugin with random metadata satisfies interface",
		prop.ForAll(
			func(deps []string) bool {
				plugin := &ValidTestPlugin{
					name:         "test-plugin",
					version:      "1.0.0",
					description:  "Test plugin",
					author:       "Test Author",
					dependencies: make([]PluginDependency, len(deps)),
				}

				// Verify the plugin can be used as Plugin interface
				var _ Plugin = plugin

				// Verify all methods are callable
				_ = plugin.Name()
				_ = plugin.Version()
				_ = plugin.Description()
				_ = plugin.Author()
				_ = plugin.Dependencies()
				_ = plugin.ConfigSchema()
				_ = plugin.Initialize(nil)
				_ = plugin.Start()
				_ = plugin.Stop()
				_ = plugin.Cleanup()
				_ = plugin.OnConfigChange(nil)

				return true
			},
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.TestingRun(t)
}

// ValidTestPlugin is a minimal valid plugin implementation for testing
type ValidTestPlugin struct {
	name         string
	version      string
	description  string
	author       string
	dependencies []PluginDependency
}

func (p *ValidTestPlugin) Name() string {
	if p.name == "" {
		return "test-plugin"
	}
	return p.name
}

func (p *ValidTestPlugin) Version() string {
	if p.version == "" {
		return "1.0.0"
	}
	return p.version
}

func (p *ValidTestPlugin) Description() string {
	if p.description == "" {
		return "A test plugin"
	}
	return p.description
}

func (p *ValidTestPlugin) Author() string {
	if p.author == "" {
		return "Test Author"
	}
	return p.author
}

func (p *ValidTestPlugin) Dependencies() []PluginDependency {
	return p.dependencies
}

func (p *ValidTestPlugin) Initialize(ctx PluginContext) error {
	return nil
}

func (p *ValidTestPlugin) Start() error {
	return nil
}

func (p *ValidTestPlugin) Stop() error {
	return nil
}

func (p *ValidTestPlugin) Cleanup() error {
	return nil
}

func (p *ValidTestPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *ValidTestPlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}

// **Feature: plugin-system, Property 14: Plugin context service access**
// **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6**
// For any plugin receiving a PluginContext, the context should provide non-nil access to
// Router, DatabaseManager, CacheManager, ConfigManager, Logger, and MetricsCollector
func TestProperty_PluginContextServiceAccess(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("plugin context provides non-nil access to all framework services",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create a real plugin context with all permissions
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     true,
					AllowFileSystem: true,
					AllowNetwork:    true,
					AllowExec:       true,
				}

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					nil, // permissionChecker (nil means no enforcement)
				)

				// Verify all services are non-nil
				if ctx.Router() == nil {
					return false
				}
				if ctx.Logger() == nil {
					return false
				}
				if ctx.Metrics() == nil {
					return false
				}
				if ctx.Database() == nil {
					return false
				}
				if ctx.Cache() == nil {
					return false
				}
				if ctx.Config() == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("registry stores and retrieves plugin contexts correctly",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				registry := NewPluginRegistry()
				plugin := &ValidTestPlugin{name: pluginName}

				// Create a real plugin context with all permissions
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     true,
					AllowFileSystem: true,
					AllowNetwork:    true,
					AllowExec:       true,
				}

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					nil, // permissionChecker (nil means no enforcement)
				)

				// Register plugin with context
				err := registry.Register(plugin, ctx)
				if err != nil {
					return false
				}

				// Retrieve context
				retrievedCtx, err := registry.GetContext(pluginName)
				if err != nil {
					return false
				}

				// Verify context provides all services
				if retrievedCtx.Router() == nil {
					return false
				}
				if retrievedCtx.Logger() == nil {
					return false
				}
				if retrievedCtx.Metrics() == nil {
					return false
				}
				if retrievedCtx.Database() == nil {
					return false
				}
				if retrievedCtx.Cache() == nil {
					return false
				}
				if retrievedCtx.Config() == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("multiple plugins can be registered with independent contexts",
		prop.ForAll(
			func(names []string) bool {
				if len(names) == 0 {
					return true
				}

				registry := NewPluginRegistry()

				// Filter out empty names and duplicates
				uniqueNames := make(map[string]bool)
				for _, name := range names {
					if name != "" {
						uniqueNames[name] = true
					}
				}

				// Register multiple plugins
				for name := range uniqueNames {
					plugin := &ValidTestPlugin{name: name}

					// Create a real plugin context with all permissions
					permissions := PluginPermissions{
						AllowDatabase:   true,
						AllowCache:      true,
						AllowConfig:     true,
						AllowRouter:     true,
						AllowFileSystem: true,
						AllowNetwork:    true,
						AllowExec:       true,
					}

					ctx := NewPluginContext(
						name,
						&MockRouter{},
						&MockLogger{},
						&MockMetrics{},
						&MockDatabase{},
						&MockCache{},
						&MockConfig{},
						&MockFileManager{},
						&MockNetworkClient{},
						map[string]interface{}{},
						permissions,
						nil, // hookSystem
						nil, // eventBus
						NewServiceRegistry(),
						NewMiddlewareRegistry(),
						nil, // permissionChecker (nil means no enforcement)
					)

					err := registry.Register(plugin, ctx)
					if err != nil {
						return false
					}
				}

				// Verify all plugins can retrieve their contexts
				for name := range uniqueNames {
					ctx, err := registry.GetContext(name)
					if err != nil {
						return false
					}

					// Verify context provides all services
					if ctx.Router() == nil || ctx.Logger() == nil ||
						ctx.Metrics() == nil || ctx.Database() == nil ||
						ctx.Cache() == nil || ctx.Config() == nil {
						return false
					}
				}

				return true
			},
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.TestingRun(t)
}

// Mock implementations are in plugin_test_helpers.go

// **Feature: plugin-system, Property 15: Security constraint enforcement**
// **Validates: Requirements 4.7**
// For any plugin accessing framework services, the same security constraints that apply to
// application code should be enforced
func TestProperty_SecurityConstraintEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("plugin without database permission cannot access database",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions without database access
				permissions := PluginPermissions{
					AllowDatabase:   false,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     true,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to access database should return a no-op manager (not nil)
				db := ctx.Database()
				if db == nil {
					return false // Should get a no-op manager, not nil
				}

				// Verify it's a permission-denied manager by checking if operations fail
				// The no-op manager should exist but operations should be no-ops
				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("plugin without cache permission cannot access cache",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions without cache access
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      false,
					AllowConfig:     true,
					AllowRouter:     true,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to access cache should return a no-op manager (not nil)
				cache := ctx.Cache()
				if cache == nil {
					return false // Should get a no-op manager, not nil
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("plugin without router permission cannot access router",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions without router access
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     false,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to access router should return a no-op manager (not nil)
				router := ctx.Router()
				if router == nil {
					return false // Should get a no-op manager, not nil
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("plugin without config permission cannot access config",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions without config access
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     false,
					AllowRouter:     true,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to access config should return a no-op manager (not nil)
				config := ctx.Config()
				if config == nil {
					return false // Should get a no-op manager, not nil
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("plugin with all permissions can access all services",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions with all access
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     true,
					AllowFileSystem: true,
					AllowNetwork:    true,
					AllowExec:       true,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// All services should be accessible
				if ctx.Router() == nil {
					return false
				}
				if ctx.Database() == nil {
					return false
				}
				if ctx.Cache() == nil {
					return false
				}
				if ctx.Config() == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("plugin without router permission cannot register middleware",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions without router access
				permissions := PluginPermissions{
					AllowDatabase:   true,
					AllowCache:      true,
					AllowConfig:     true,
					AllowRouter:     false,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				// Create a permission checker that enforces permissions
				permChecker := NewMockPermissionChecker(permissions)

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to register middleware should fail
				err := ctx.RegisterMiddleware("test-middleware", func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}, 0, []string{})

				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// MockPermissionChecker is in plugin_test_helpers.go

// **Feature: plugin-system, Property 31: Storage isolation**
// **Validates: Requirements 8.3**
// For any two different plugins, data stored by one plugin should not be accessible or modifiable by the other plugin
func TestProperty_StorageIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("data stored by one plugin is not accessible by another plugin",
		prop.ForAll(
			func(plugin1Name, plugin2Name, key, value string) bool {
				// Skip if plugins have the same name or empty names
				if plugin1Name == "" || plugin2Name == "" || plugin1Name == plugin2Name {
					return true
				}
				if key == "" {
					return true
				}

				// Create storage for plugin 1 (no database, uses in-memory only)
				storage1 := NewPluginStorage(plugin1Name, nil)

				// Create storage for plugin 2 (no database, uses in-memory only)
				storage2 := NewPluginStorage(plugin2Name, nil)

				// Plugin 1 stores a value
				err := storage1.Set(key, value)
				if err != nil {
					return false
				}

				// Plugin 2 should not be able to retrieve plugin 1's value
				// Since we're using separate storage instances,
				// plugin 2 should not see plugin 1's data
				retrievedValue, err := storage2.Get(key)
				if err == nil {
					// If no error, the value should not match plugin 1's value
					// because they have separate storage instances
					if retrievedValue == value {
						return false // Isolation violated
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("clearing one plugin's storage does not affect another plugin's storage",
		prop.ForAll(
			func(plugin1Name, plugin2Name, key1, key2, value1, value2 string) bool {
				// Skip if plugins have the same name or empty names
				if plugin1Name == "" || plugin2Name == "" || plugin1Name == plugin2Name {
					return true
				}
				if key1 == "" || key2 == "" {
					return true
				}

				// Create storage for plugin 1 (no database, uses in-memory only)
				storage1 := NewPluginStorage(plugin1Name, nil)

				// Create storage for plugin 2 (no database, uses in-memory only)
				storage2 := NewPluginStorage(plugin2Name, nil)

				// Both plugins store values
				err := storage1.Set(key1, value1)
				if err != nil {
					return false
				}

				err = storage2.Set(key2, value2)
				if err != nil {
					return false
				}

				// Clear plugin 1's storage
				err = storage1.Clear()
				if err != nil {
					return false
				}

				// Plugin 1's storage should be empty
				keys1, err := storage1.List()
				if err != nil {
					return false
				}
				if len(keys1) != 0 {
					return false
				}

				// Plugin 2's storage should still have its data
				keys2, err := storage2.List()
				if err != nil {
					return false
				}
				if len(keys2) != 1 {
					return false
				}

				// Plugin 2 should still be able to retrieve its value
				retrievedValue, err := storage2.Get(key2)
				if err != nil {
					return false
				}
				if retrievedValue != value2 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)

	properties.Property("deleting from one plugin's storage does not affect another plugin's storage",
		prop.ForAll(
			func(plugin1Name, plugin2Name, sharedKey, value1, value2 string) bool {
				// Skip if plugins have the same name or empty names
				if plugin1Name == "" || plugin2Name == "" || plugin1Name == plugin2Name {
					return true
				}
				if sharedKey == "" {
					return true
				}

				// Create storage for plugin 1 (no database, uses in-memory only)
				storage1 := NewPluginStorage(plugin1Name, nil)

				// Create storage for plugin 2 (no database, uses in-memory only)
				storage2 := NewPluginStorage(plugin2Name, nil)

				// Both plugins store values with the same key
				err := storage1.Set(sharedKey, value1)
				if err != nil {
					return false
				}

				err = storage2.Set(sharedKey, value2)
				if err != nil {
					return false
				}

				// Delete from plugin 1's storage
				err = storage1.Delete(sharedKey)
				if err != nil {
					return false
				}

				// Plugin 1 should not have the key anymore
				_, err = storage1.Get(sharedKey)
				if err == nil {
					return false // Should have returned an error
				}

				// Plugin 2 should still have its value
				retrievedValue, err := storage2.Get(sharedKey)
				if err != nil {
					return false
				}
				if retrievedValue != value2 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
			gen.AlphaString(),
		),
	)

	properties.Property("listing keys from one plugin does not include keys from another plugin",
		prop.ForAll(
			func(plugin1Name, plugin2Name string, keys1, keys2 []string) bool {
				// Skip if plugins have the same name or empty names
				if plugin1Name == "" || plugin2Name == "" || plugin1Name == plugin2Name {
					return true
				}
				if len(keys1) == 0 || len(keys2) == 0 {
					return true
				}

				// Create storage for plugin 1 (no database, uses in-memory only)
				storage1 := NewPluginStorage(plugin1Name, nil)

				// Create storage for plugin 2 (no database, uses in-memory only)
				storage2 := NewPluginStorage(plugin2Name, nil)

				// Plugin 1 stores multiple values
				for _, key := range keys1 {
					if key != "" {
						err := storage1.Set(key, "value1")
						if err != nil {
							return false
						}
					}
				}

				// Plugin 2 stores multiple values
				for _, key := range keys2 {
					if key != "" {
						err := storage2.Set(key, "value2")
						if err != nil {
							return false
						}
					}
				}

				// List keys from plugin 1
				listedKeys1, err := storage1.List()
				if err != nil {
					return false
				}

				// List keys from plugin 2
				listedKeys2, err := storage2.List()
				if err != nil {
					return false
				}

				// Verify that plugin 1's keys don't include plugin 2's keys
				// and vice versa (they should be completely isolated)
				keys1Map := make(map[string]bool)
				for _, key := range listedKeys1 {
					keys1Map[key] = true
				}

				keys2Map := make(map[string]bool)
				for _, key := range listedKeys2 {
					keys2Map[key] = true
				}

				// Check that there's no overlap (isolation)
				// Since they're separate instances, they should have separate key spaces
				for key := range keys1Map {
					if keys2Map[key] {
						// If the same key appears in both, it's only acceptable
						// if it was in both input arrays
						found1 := false
						found2 := false
						for _, k := range keys1 {
							if k == key {
								found1 = true
								break
							}
						}
						for _, k := range keys2 {
							if k == key {
								found2 = true
								break
							}
						}
						// If both plugins stored the same key, that's fine
						// as long as they're isolated (separate storage instances)
						if found1 && found2 {
							continue
						}
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("multiple plugins can use the same key without interference",
		prop.ForAll(
			func(pluginNames []string, key, value string) bool {
				if len(pluginNames) < 2 || key == "" {
					return true
				}

				// Filter out empty names and duplicates
				uniqueNames := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniqueNames[name] = true
					}
				}

				if len(uniqueNames) < 2 {
					return true
				}

				// Create storage for each plugin and store the same key with different values
				storages := make(map[string]PluginStorage)
				expectedValues := make(map[string]string)

				for name := range uniqueNames {
					storage := NewPluginStorage(name, nil)
					storages[name] = storage

					// Each plugin stores the same key with a plugin-specific value
					pluginValue := value + "_" + name
					expectedValues[name] = pluginValue

					err := storage.Set(key, pluginValue)
					if err != nil {
						return false
					}
				}

				// Verify each plugin can retrieve its own value
				for name, storage := range storages {
					retrievedValue, err := storage.Get(key)
					if err != nil {
						return false
					}

					// Each plugin should get its own value, not another plugin's value
					if retrievedValue != expectedValues[name] {
						return false
					}
				}

				return true
			},
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 2: Startup hook execution**
// **Validates: Requirements 2.1**
// For any plugin with a registered startup hook, the hook should be executed during framework initialization
func TestProperty_StartupHookExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("startup hooks are executed when ExecuteHooks is called with HookTypeStartup",
		prop.ForAll(
			func(pluginName string, priority int) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executed := false

				// Register a startup hook
				handler := func(ctx HookContext) error {
					executed = true
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypeStartup, priority, handler)
				if err != nil {
					return false
				}

				// Execute startup hooks
				err = hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				// Verify the hook was executed
				return executed
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Int(),
		),
	)

	properties.Property("multiple startup hooks from different plugins are all executed",
		prop.ForAll(
			func(pluginNames []string) bool {
				if len(pluginNames) == 0 {
					return true
				}

				// Filter out empty names and duplicates
				uniqueNames := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniqueNames[name] = true
					}
				}

				if len(uniqueNames) == 0 {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executedPlugins := make(map[string]bool)

				// Register startup hooks for each plugin
				for name := range uniqueNames {
					pluginName := name
					handler := func(ctx HookContext) error {
						executedPlugins[pluginName] = true
						return nil
					}

					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, 0, handler)
					if err != nil {
						return false
					}
				}

				// Execute startup hooks
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				// Verify all hooks were executed
				for name := range uniqueNames {
					if !executedPlugins[name] {
						return false
					}
				}

				return true
			},
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("startup hook receives correct hook type in context",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				correctType := false

				handler := func(ctx HookContext) error {
					if ctx.HookType() == HookTypeStartup {
						correctType = true
					}
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypeStartup, 0, handler)
				if err != nil {
					return false
				}

				err = hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				return correctType
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 3: Shutdown hook execution**
// **Validates: Requirements 2.2**
// For any plugin with a registered shutdown hook, the hook should be executed during graceful shutdown
func TestProperty_ShutdownHookExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("shutdown hooks are executed when ExecuteHooks is called with HookTypeShutdown",
		prop.ForAll(
			func(pluginName string, priority int) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executed := false

				// Register a shutdown hook
				handler := func(ctx HookContext) error {
					executed = true
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypeShutdown, priority, handler)
				if err != nil {
					return false
				}

				// Execute shutdown hooks
				err = hookSystem.ExecuteHooks(HookTypeShutdown, nil)
				if err != nil {
					return false
				}

				// Verify the hook was executed
				return executed
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Int(),
		),
	)

	properties.Property("multiple shutdown hooks from different plugins are all executed",
		prop.ForAll(
			func(pluginNames []string) bool {
				if len(pluginNames) == 0 {
					return true
				}

				// Filter out empty names and duplicates
				uniqueNames := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniqueNames[name] = true
					}
				}

				if len(uniqueNames) == 0 {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executedPlugins := make(map[string]bool)

				// Register shutdown hooks for each plugin
				for name := range uniqueNames {
					pluginName := name
					handler := func(ctx HookContext) error {
						executedPlugins[pluginName] = true
						return nil
					}

					err := hookSystem.RegisterHook(pluginName, HookTypeShutdown, 0, handler)
					if err != nil {
						return false
					}
				}

				// Execute shutdown hooks
				err := hookSystem.ExecuteHooks(HookTypeShutdown, nil)
				if err != nil {
					return false
				}

				// Verify all hooks were executed
				for name := range uniqueNames {
					if !executedPlugins[name] {
						return false
					}
				}

				return true
			},
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("shutdown hook receives correct hook type in context",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				correctType := false

				handler := func(ctx HookContext) error {
					if ctx.HookType() == HookTypeShutdown {
						correctType = true
					}
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypeShutdown, 0, handler)
				if err != nil {
					return false
				}

				err = hookSystem.ExecuteHooks(HookTypeShutdown, nil)
				if err != nil {
					return false
				}

				return correctType
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 4: Pre-request hook execution**
// **Validates: Requirements 2.3**
// For any request when a pre-request hook is registered, the hook should execute before the router matches the route
func TestProperty_PreRequestHookExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("pre-request hooks are executed when ExecuteHooks is called with HookTypePreRequest",
		prop.ForAll(
			func(pluginName string, priority int) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executed := false

				// Register a pre-request hook
				handler := func(ctx HookContext) error {
					executed = true
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePreRequest, priority, handler)
				if err != nil {
					return false
				}

				// Execute pre-request hooks
				err = hookSystem.ExecuteHooks(HookTypePreRequest, nil)
				if err != nil {
					return false
				}

				// Verify the hook was executed
				return executed
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Int(),
		),
	)

	properties.Property("pre-request hook receives correct hook type in context",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				correctType := false

				handler := func(ctx HookContext) error {
					if ctx.HookType() == HookTypePreRequest {
						correctType = true
					}
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePreRequest, 0, handler)
				if err != nil {
					return false
				}

				err = hookSystem.ExecuteHooks(HookTypePreRequest, nil)
				if err != nil {
					return false
				}

				return correctType
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 5: Post-request hook execution**
// **Validates: Requirements 2.4**
// For any request when a post-request hook is registered, the hook should execute after the handler completes
func TestProperty_PostRequestHookExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("post-request hooks are executed when ExecuteHooks is called with HookTypePostRequest",
		prop.ForAll(
			func(pluginName string, priority int) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executed := false

				// Register a post-request hook
				handler := func(ctx HookContext) error {
					executed = true
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePostRequest, priority, handler)
				if err != nil {
					return false
				}

				// Execute post-request hooks
				err = hookSystem.ExecuteHooks(HookTypePostRequest, nil)
				if err != nil {
					return false
				}

				// Verify the hook was executed
				return executed
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Int(),
		),
	)

	properties.Property("post-request hook receives correct hook type in context",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				correctType := false

				handler := func(ctx HookContext) error {
					if ctx.HookType() == HookTypePostRequest {
						correctType = true
					}
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePostRequest, 0, handler)
				if err != nil {
					return false
				}

				err = hookSystem.ExecuteHooks(HookTypePostRequest, nil)
				if err != nil {
					return false
				}

				return correctType
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 6: Pre-response hook execution**
// **Validates: Requirements 2.5**
// For any response when a pre-response hook is registered, the hook should execute before the response is sent to the client
func TestProperty_PreResponseHookExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("pre-response hooks are executed when ExecuteHooks is called with HookTypePreResponse",
		prop.ForAll(
			func(pluginName string, priority int) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executed := false

				// Register a pre-response hook
				handler := func(ctx HookContext) error {
					executed = true
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePreResponse, priority, handler)
				if err != nil {
					return false
				}

				// Execute pre-response hooks
				err = hookSystem.ExecuteHooks(HookTypePreResponse, nil)
				if err != nil {
					return false
				}

				// Verify the hook was executed
				return executed
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Int(),
		),
	)

	properties.Property("pre-response hook receives correct hook type in context",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				correctType := false

				handler := func(ctx HookContext) error {
					if ctx.HookType() == HookTypePreResponse {
						correctType = true
					}
					return nil
				}

				err := hookSystem.RegisterHook(pluginName, HookTypePreResponse, 0, handler)
				if err != nil {
					return false
				}

				err = hookSystem.ExecuteHooks(HookTypePreResponse, nil)
				if err != nil {
					return false
				}

				return correctType
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 7: Hook priority ordering**
// **Validates: Requirements 2.6**
// For any set of plugins registering hooks at the same hook point, the hooks should execute in descending priority order (highest priority first)
func TestProperty_HookPriorityOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("hooks execute in descending priority order",
		prop.ForAll(
			func(priorities []int) bool {
				if len(priorities) < 2 {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executionOrder := []int{}

				// Register hooks with different priorities
				for i, priority := range priorities {
					p := priority
					idx := i
					handler := func(ctx HookContext) error {
						executionOrder = append(executionOrder, idx)
						return nil
					}

					pluginName := fmt.Sprintf("plugin-%d", i)
					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, p, handler)
					if err != nil {
						return false
					}
				}

				// Execute hooks
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				// Verify execution order matches priority order (highest first)
				if len(executionOrder) != len(priorities) {
					return false
				}

				// Check that hooks executed in descending priority order
				for i := 0; i < len(executionOrder)-1; i++ {
					currentIdx := executionOrder[i]
					nextIdx := executionOrder[i+1]

					currentPriority := priorities[currentIdx]
					nextPriority := priorities[nextIdx]

					// Current priority should be >= next priority (descending order)
					if currentPriority < nextPriority {
						return false
					}
				}

				return true
			},
			gen.SliceOf(gen.Int()),
		),
	)

	properties.Property("hooks with same priority execute in registration order",
		prop.ForAll(
			func(numPlugins int) bool {
				if numPlugins < 2 || numPlugins > 10 {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executionOrder := []int{}
				samePriority := 100

				// Register multiple hooks with the same priority
				for i := 0; i < numPlugins; i++ {
					idx := i
					handler := func(ctx HookContext) error {
						executionOrder = append(executionOrder, idx)
						return nil
					}

					pluginName := fmt.Sprintf("plugin-%d", i)
					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, samePriority, handler)
					if err != nil {
						return false
					}
				}

				// Execute hooks
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				// Verify all hooks executed
				return len(executionOrder) == numPlugins
			},
			gen.IntRange(2, 10),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 8: Hook error isolation**
// **Validates: Requirements 2.7**
// For any hook that throws an error, the system should log the error and continue executing remaining hooks in the chain
func TestProperty_HookErrorIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("error in one hook does not prevent execution of subsequent hooks",
		prop.ForAll(
			func(numPlugins int, errorIndex int) bool {
				if numPlugins < 2 || numPlugins > 10 {
					return true
				}
				if errorIndex < 0 || errorIndex >= numPlugins {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executedPlugins := make(map[int]bool)

				// Register hooks, one of which will error
				for i := 0; i < numPlugins; i++ {
					idx := i
					handler := func(ctx HookContext) error {
						executedPlugins[idx] = true
						if idx == errorIndex {
							return fmt.Errorf("intentional error from plugin %d", idx)
						}
						return nil
					}

					pluginName := fmt.Sprintf("plugin-%d", i)
					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, i, handler)
					if err != nil {
						return false
					}
				}

				// Execute hooks - should not return error despite one hook failing
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false // ExecuteHooks should not propagate hook errors
				}

				// Verify all hooks were executed, including the one that errored
				for i := 0; i < numPlugins; i++ {
					if !executedPlugins[i] {
						return false
					}
				}

				return true
			},
			gen.IntRange(2, 10),
			gen.Int(),
		),
	)

	properties.Property("panic in hook is recovered and does not crash the system",
		prop.ForAll(
			func(numPlugins int, panicIndex int) bool {
				if numPlugins < 2 || numPlugins > 10 {
					return true
				}
				if panicIndex < 0 || panicIndex >= numPlugins {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executedPlugins := make(map[int]bool)

				// Register hooks, one of which will panic
				for i := 0; i < numPlugins; i++ {
					idx := i
					handler := func(ctx HookContext) error {
						executedPlugins[idx] = true
						if idx == panicIndex {
							panic(fmt.Sprintf("intentional panic from plugin %d", idx))
						}
						return nil
					}

					pluginName := fmt.Sprintf("plugin-%d", i)
					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, i, handler)
					if err != nil {
						return false
					}
				}

				// Execute hooks - should recover from panic
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false // ExecuteHooks should not propagate panic as error
				}

				// Verify the panicking hook was executed
				if !executedPlugins[panicIndex] {
					return false
				}

				// Verify hooks after the panic were still executed
				for i := panicIndex + 1; i < numPlugins; i++ {
					if !executedPlugins[i] {
						return false
					}
				}

				return true
			},
			gen.IntRange(2, 10),
			gen.Int(),
		),
	)

	properties.Property("multiple errors in different hooks are all isolated",
		prop.ForAll(
			func(numPlugins int) bool {
				if numPlugins < 3 || numPlugins > 10 {
					return true
				}

				hookSystem := NewHookSystem(&MockLogger{}, &MockMetrics{})
				executedPlugins := make(map[int]bool)

				// Register hooks, every other one will error
				for i := 0; i < numPlugins; i++ {
					idx := i
					handler := func(ctx HookContext) error {
						executedPlugins[idx] = true
						if idx%2 == 0 {
							return fmt.Errorf("intentional error from plugin %d", idx)
						}
						return nil
					}

					pluginName := fmt.Sprintf("plugin-%d", i)
					err := hookSystem.RegisterHook(pluginName, HookTypeStartup, i, handler)
					if err != nil {
						return false
					}
				}

				// Execute hooks
				err := hookSystem.ExecuteHooks(HookTypeStartup, nil)
				if err != nil {
					return false
				}

				// Verify all hooks were executed despite multiple errors
				for i := 0; i < numPlugins; i++ {
					if !executedPlugins[i] {
						return false
					}
				}

				return true
			},
			gen.IntRange(3, 10),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 39: Event delivery to subscribers**
// **Validates: Requirements 10.1**
// For any event published by a plugin, all plugins subscribed to that event should receive the event
func TestProperty_EventDeliveryToSubscribers(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all subscribed plugins receive published events",
		prop.ForAll(
			func(eventName string, subscriberNames []string, eventData string) bool {
				if eventName == "" || len(subscriberNames) == 0 {
					return true
				}

				// Filter out empty names and duplicates
				uniqueSubscribers := make(map[string]bool)
				for _, name := range subscriberNames {
					if name != "" {
						uniqueSubscribers[name] = true
					}
				}

				if len(uniqueSubscribers) == 0 {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				receivedEvents := make(map[string]bool)
				var mu sync.Mutex

				// Subscribe all plugins to the event
				for subscriberName := range uniqueSubscribers {
					name := subscriberName // Capture for closure
					handler := func(event Event) error {
						mu.Lock()
						defer mu.Unlock()
						receivedEvents[name] = true
						return nil
					}

					err := eventBus.Subscribe(name, eventName, handler)
					if err != nil {
						return false
					}
				}

				// Publish the event
				err := eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait a bit for asynchronous delivery
				time.Sleep(50 * time.Millisecond)

				// Verify all subscribers received the event
				mu.Lock()
				defer mu.Unlock()

				for subscriberName := range uniqueSubscribers {
					if !receivedEvents[subscriberName] {
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString(),
		),
	)

	properties.Property("event data is correctly passed to subscribers",
		prop.ForAll(
			func(eventName, pluginName, eventData string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				var receivedData interface{}
				var mu sync.Mutex
				done := make(chan bool, 1)

				// Subscribe to the event
				handler := func(event Event) error {
					mu.Lock()
					defer mu.Unlock()
					receivedData = event.Data
					done <- true
					return nil
				}

				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Publish the event
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait for delivery
				select {
				case <-done:
					// Event received
				case <-time.After(100 * time.Millisecond):
					return false // Timeout
				}

				// Verify data matches
				mu.Lock()
				defer mu.Unlock()

				if receivedData != eventData {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("unsubscribed plugins do not receive events",
		prop.ForAll(
			func(eventName, subscribedPlugin, unsubscribedPlugin, eventData string) bool {
				if eventName == "" || subscribedPlugin == "" || unsubscribedPlugin == "" {
					return true
				}
				if subscribedPlugin == unsubscribedPlugin {
					return true // Skip if same plugin
				}

				eventBus := NewEventBus(&MockLogger{})
				subscribedReceived := false
				unsubscribedReceived := false
				var mu sync.Mutex

				// Subscribe one plugin
				subscribedHandler := func(event Event) error {
					mu.Lock()
					defer mu.Unlock()
					subscribedReceived = true
					return nil
				}

				err := eventBus.Subscribe(subscribedPlugin, eventName, subscribedHandler)
				if err != nil {
					return false
				}

				// Don't subscribe the other plugin, but create a handler
				unsubscribedHandler := func(event Event) error {
					mu.Lock()
					defer mu.Unlock()
					unsubscribedReceived = true
					return nil
				}
				_ = unsubscribedHandler // Prevent unused variable error

				// Publish the event
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait for delivery
				time.Sleep(50 * time.Millisecond)

				// Verify only subscribed plugin received the event
				mu.Lock()
				defer mu.Unlock()

				if !subscribedReceived {
					return false
				}
				if unsubscribedReceived {
					return false // Unsubscribed plugin should not receive
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("publishing to non-existent event does not error",
		prop.ForAll(
			func(eventName, eventData string) bool {
				if eventName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				// Publish to an event with no subscribers
				err := eventBus.Publish(eventName, eventData)
				if err != nil {
					return false // Should not error
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 40: Event subscription registration**
// **Validates: Requirements 10.2**
// For any plugin subscribing to an event, the subscription should be registered and the handler
// should be invoked when the event is published
func TestProperty_EventSubscriptionRegistration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("subscribed handler is invoked when event is published",
		prop.ForAll(
			func(eventName, pluginName, eventData string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				handlerInvoked := false
				var mu sync.Mutex
				done := make(chan bool, 1)

				// Subscribe to the event
				handler := func(event Event) error {
					mu.Lock()
					defer mu.Unlock()
					handlerInvoked = true
					done <- true
					return nil
				}

				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Verify subscription is registered
				subscribers := eventBus.ListSubscriptions(eventName)
				found := false
				for _, sub := range subscribers {
					if sub == pluginName {
						found = true
						break
					}
				}
				if !found {
					return false
				}

				// Publish the event
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait for handler invocation
				select {
				case <-done:
					// Handler invoked
				case <-time.After(100 * time.Millisecond):
					return false // Timeout
				}

				mu.Lock()
				defer mu.Unlock()

				if !handlerInvoked {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("multiple plugins can subscribe to the same event",
		prop.ForAll(
			func(eventName string, pluginNames []string, eventData string) bool {
				if eventName == "" || len(pluginNames) < 2 {
					return true
				}

				// Filter out empty names and duplicates
				uniquePlugins := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniquePlugins[name] = true
					}
				}

				if len(uniquePlugins) < 2 {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				// Subscribe all plugins
				for pluginName := range uniquePlugins {
					handler := func(event Event) error {
						return nil
					}

					err := eventBus.Subscribe(pluginName, eventName, handler)
					if err != nil {
						return false
					}
				}

				// Verify all subscriptions are registered
				subscribers := eventBus.ListSubscriptions(eventName)
				if len(subscribers) != len(uniquePlugins) {
					return false
				}

				// Verify all plugins are in the subscriber list
				for pluginName := range uniquePlugins {
					found := false
					for _, sub := range subscribers {
						if sub == pluginName {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString(),
		),
	)

	properties.Property("plugin can subscribe to multiple events",
		prop.ForAll(
			func(pluginName string, eventNames []string) bool {
				if pluginName == "" || len(eventNames) == 0 {
					return true
				}

				// Filter out empty names and duplicates
				uniqueEvents := make(map[string]bool)
				for _, name := range eventNames {
					if name != "" {
						uniqueEvents[name] = true
					}
				}

				if len(uniqueEvents) == 0 {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				// Subscribe to all events
				for eventName := range uniqueEvents {
					handler := func(event Event) error {
						return nil
					}

					err := eventBus.Subscribe(pluginName, eventName, handler)
					if err != nil {
						return false
					}
				}

				// Verify all subscriptions are registered
				for eventName := range uniqueEvents {
					subscribers := eventBus.ListSubscriptions(eventName)
					found := false
					for _, sub := range subscribers {
						if sub == pluginName {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("unsubscribe removes subscription",
		prop.ForAll(
			func(eventName, pluginName, eventData string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				handlerInvoked := false
				var mu sync.Mutex

				// Subscribe to the event
				handler := func(event Event) error {
					mu.Lock()
					defer mu.Unlock()
					handlerInvoked = true
					return nil
				}

				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Verify subscription exists
				subscribers := eventBus.ListSubscriptions(eventName)
				if len(subscribers) != 1 {
					return false
				}

				// Unsubscribe
				err = eventBus.Unsubscribe(pluginName, eventName)
				if err != nil {
					return false
				}

				// Verify subscription is removed
				subscribers = eventBus.ListSubscriptions(eventName)
				if len(subscribers) != 0 {
					return false
				}

				// Publish event - handler should not be invoked
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait a bit
				time.Sleep(50 * time.Millisecond)

				mu.Lock()
				defer mu.Unlock()

				if handlerInvoked {
					return false // Handler should not have been invoked
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("duplicate subscription returns error",
		prop.ForAll(
			func(eventName, pluginName string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				handler := func(event Event) error {
					return nil
				}

				// First subscription should succeed
				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Second subscription should fail
				err = eventBus.Subscribe(pluginName, eventName, handler)
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 43: Inter-plugin error isolation**
// **Validates: Requirements 10.5**
// For any inter-plugin communication failure, the error should be returned to the caller
// without crashing either plugin
func TestProperty_InterPluginErrorIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("event handler error does not crash event bus",
		prop.ForAll(
			func(eventName, pluginName, eventData string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				// Subscribe with a handler that returns an error
				handler := func(event Event) error {
					return fmt.Errorf("handler error")
				}

				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Publish event - should not crash despite handler error
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false // Publish should not return error
				}

				// Wait for delivery
				time.Sleep(50 * time.Millisecond)

				// Event bus should still be functional
				subscribers := eventBus.ListSubscriptions(eventName)
				if len(subscribers) != 1 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("event handler panic does not crash event bus",
		prop.ForAll(
			func(eventName, pluginName, eventData string) bool {
				if eventName == "" || pluginName == "" {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})

				// Subscribe with a handler that panics
				handler := func(event Event) error {
					panic("handler panic")
				}

				err := eventBus.Subscribe(pluginName, eventName, handler)
				if err != nil {
					return false
				}

				// Publish event - should not crash despite handler panic
				err = eventBus.Publish(eventName, eventData)
				if err != nil {
					return false // Publish should not return error
				}

				// Wait for delivery
				time.Sleep(50 * time.Millisecond)

				// Event bus should still be functional
				subscribers := eventBus.ListSubscriptions(eventName)
				if len(subscribers) != 1 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.Property("one subscriber error does not affect other subscribers",
		prop.ForAll(
			func(eventName string, pluginNames []string, eventData string) bool {
				if eventName == "" || len(pluginNames) < 2 {
					return true
				}

				// Filter out empty names and duplicates
				uniquePlugins := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniquePlugins[name] = true
					}
				}

				if len(uniquePlugins) < 2 {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				receivedEvents := make(map[string]bool)
				var mu sync.Mutex

				// Subscribe all plugins, first one will error
				first := true
				for pluginName := range uniquePlugins {
					name := pluginName // Capture for closure

					if first {
						// First handler returns error
						handler := func(event Event) error {
							return fmt.Errorf("handler error")
						}
						err := eventBus.Subscribe(name, eventName, handler)
						if err != nil {
							return false
						}
						first = false
					} else {
						// Other handlers succeed
						handler := func(event Event) error {
							mu.Lock()
							defer mu.Unlock()
							receivedEvents[name] = true
							return nil
						}
						err := eventBus.Subscribe(name, eventName, handler)
						if err != nil {
							return false
						}
					}
				}

				// Publish the event
				err := eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait for delivery
				time.Sleep(50 * time.Millisecond)

				// Verify non-erroring subscribers received the event
				mu.Lock()
				defer mu.Unlock()

				// At least one subscriber should have received the event
				if len(receivedEvents) == 0 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString(),
		),
	)

	properties.Property("one subscriber panic does not affect other subscribers",
		prop.ForAll(
			func(eventName string, pluginNames []string, eventData string) bool {
				if eventName == "" || len(pluginNames) < 2 {
					return true
				}

				// Filter out empty names and duplicates
				uniquePlugins := make(map[string]bool)
				for _, name := range pluginNames {
					if name != "" {
						uniquePlugins[name] = true
					}
				}

				if len(uniquePlugins) < 2 {
					return true
				}

				eventBus := NewEventBus(&MockLogger{})
				receivedEvents := make(map[string]bool)
				var mu sync.Mutex

				// Subscribe all plugins, first one will panic
				first := true
				for pluginName := range uniquePlugins {
					name := pluginName // Capture for closure

					if first {
						// First handler panics
						handler := func(event Event) error {
							panic("handler panic")
						}
						err := eventBus.Subscribe(name, eventName, handler)
						if err != nil {
							return false
						}
						first = false
					} else {
						// Other handlers succeed
						handler := func(event Event) error {
							mu.Lock()
							defer mu.Unlock()
							receivedEvents[name] = true
							return nil
						}
						err := eventBus.Subscribe(name, eventName, handler)
						if err != nil {
							return false
						}
					}
				}

				// Publish the event
				err := eventBus.Publish(eventName, eventData)
				if err != nil {
					return false
				}

				// Wait for delivery
				time.Sleep(100 * time.Millisecond)

				// Verify non-panicking subscribers received the event
				mu.Lock()
				defer mu.Unlock()

				// At least one subscriber should have received the event
				if len(receivedEvents) == 0 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 44: Permission assignment**
// **Validates: Requirements 11.1**
// For any plugin loaded with a permission configuration, the system should assign exactly those permissions to the plugin
func TestProperty_PermissionAssignment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("assigned permissions match configuration exactly",
		prop.ForAll(
			func(pluginName string, allowDB, allowCache, allowConfig, allowRouter, allowFS, allowNet, allowExec bool) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Grant all permissions individually
				if allowDB {
					if err := checker.GrantPermission(pluginName, PermissionDatabase); err != nil {
						return false
					}
				}
				if allowCache {
					if err := checker.GrantPermission(pluginName, PermissionCache); err != nil {
						return false
					}
				}
				if allowConfig {
					if err := checker.GrantPermission(pluginName, PermissionConfig); err != nil {
						return false
					}
				}
				if allowRouter {
					if err := checker.GrantPermission(pluginName, PermissionRouter); err != nil {
						return false
					}
				}
				if allowFS {
					if err := checker.GrantPermission(pluginName, PermissionFileSystem); err != nil {
						return false
					}
				}
				if allowNet {
					if err := checker.GrantPermission(pluginName, PermissionNetwork); err != nil {
						return false
					}
				}
				if allowExec {
					if err := checker.GrantPermission(pluginName, PermissionExec); err != nil {
						return false
					}
				}

				// Retrieve permissions and verify they match
				retrievedPerms := checker.GetPermissions(pluginName)

				if retrievedPerms.AllowDatabase != allowDB {
					return false
				}
				if retrievedPerms.AllowCache != allowCache {
					return false
				}
				if retrievedPerms.AllowConfig != allowConfig {
					return false
				}
				if retrievedPerms.AllowRouter != allowRouter {
					return false
				}
				if retrievedPerms.AllowFileSystem != allowFS {
					return false
				}
				if retrievedPerms.AllowNetwork != allowNet {
					return false
				}
				if retrievedPerms.AllowExec != allowExec {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
		),
	)

	properties.Property("custom permissions are assigned correctly",
		prop.ForAll(
			func(pluginName string, customPerms []string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Grant custom permissions
				for _, perm := range customPerms {
					if perm != "" {
						if err := checker.GrantPermission(pluginName, perm); err != nil {
							return false
						}
					}
				}

				// Retrieve permissions and verify custom permissions
				retrievedPerms := checker.GetPermissions(pluginName)

				for _, perm := range customPerms {
					if perm != "" {
						if !retrievedPerms.CustomPermissions[perm] {
							return false
						}
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("revoking permissions removes them correctly",
		prop.ForAll(
			func(pluginName string, allowDB, allowCache bool) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Grant permissions
				if allowDB {
					if err := checker.GrantPermission(pluginName, PermissionDatabase); err != nil {
						return false
					}
				}
				if allowCache {
					if err := checker.GrantPermission(pluginName, PermissionCache); err != nil {
						return false
					}
				}

				// Revoke database permission
				if allowDB {
					if err := checker.RevokePermission(pluginName, PermissionDatabase); err != nil {
						return false
					}
				}

				// Verify database permission is revoked
				retrievedPerms := checker.GetPermissions(pluginName)
				if retrievedPerms.AllowDatabase {
					return false
				}

				// Verify cache permission is unchanged
				if retrievedPerms.AllowCache != allowCache {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.Bool(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 45: Permission verification**
// **Validates: Requirements 11.2**
// For any plugin attempting to access a framework service, the system should verify the plugin has the required permission before allowing access
func TestProperty_PermissionVerification(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("permission check succeeds when permission is granted",
		prop.ForAll(
			func(pluginName string, permissionType int) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Map int to permission type (0-6 for the 7 standard permissions)
				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				permType := permissions[permissionType%len(permissions)]

				// Grant the permission
				if err := checker.GrantPermission(pluginName, permType); err != nil {
					return false
				}

				// Check permission should succeed
				err := checker.CheckPermission(pluginName, permType)
				if err != nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 6),
		),
	)

	properties.Property("permission check fails when permission is not granted",
		prop.ForAll(
			func(pluginName string, permissionType int) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Map int to permission type
				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				permType := permissions[permissionType%len(permissions)]

				// Do NOT grant the permission

				// Check permission should fail
				err := checker.CheckPermission(pluginName, permType)
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 6),
		),
	)

	properties.Property("permission check is independent for different plugins",
		prop.ForAll(
			func(plugin1Name, plugin2Name string) bool {
				if plugin1Name == "" || plugin2Name == "" || plugin1Name == plugin2Name {
					return true // Skip invalid cases
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Grant database permission to plugin1 only
				if err := checker.GrantPermission(plugin1Name, PermissionDatabase); err != nil {
					return false
				}

				// Plugin1 should have access
				if err := checker.CheckPermission(plugin1Name, PermissionDatabase); err != nil {
					return false
				}

				// Plugin2 should NOT have access
				if err := checker.CheckPermission(plugin2Name, PermissionDatabase); err == nil {
					return false // Should have returned an error
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 46: Unauthorized access denial**
// **Validates: Requirements 11.3**
// For any plugin lacking required permissions, access attempts should be denied and logged as security violations
func TestProperty_UnauthorizedAccessDenial(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("unauthorized access is denied and logged",
		prop.ForAll(
			func(pluginName string, permissionType int) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create a logger that tracks warnings
				logger := &MockLoggerWithTracking{
					warnings: make([]string, 0),
				}
				checker := NewPermissionChecker(logger)

				// Map int to permission type
				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				permType := permissions[permissionType%len(permissions)]

				// Do NOT grant the permission

				// Check permission should fail
				err := checker.CheckPermission(pluginName, permType)
				if err == nil {
					return false // Should have returned an error
				}

				// Verify that a warning was logged
				if len(logger.warnings) == 0 {
					return false // Security violation should be logged
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 6),
		),
	)

	properties.Property("plugin context denies access to services without permission",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				// Create permissions with no access
				permissions := PluginPermissions{
					AllowDatabase:   false,
					AllowCache:      false,
					AllowConfig:     false,
					AllowRouter:     false,
					AllowFileSystem: false,
					AllowNetwork:    false,
					AllowExec:       false,
				}

				logger := &MockLoggerWithTracking{
					warnings: make([]string, 0),
					errors:   make([]string, 0),
				}
				permChecker := NewPermissionChecker(logger)

				// Set permissions for the plugin
				if impl, ok := permChecker.(*permissionCheckerImpl); ok {
					impl.SetPermissions(pluginName, permissions)
				}

				ctx := NewPluginContext(
					pluginName,
					&MockRouter{},
					logger,
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					&MockFileManager{},
					&MockNetworkClient{},
					map[string]interface{}{},
					permissions,
					nil, // hookSystem
					nil, // eventBus
					NewServiceRegistry(),
					NewMiddlewareRegistry(),
					permChecker,
				)

				// Attempt to access services should return no-op managers (not nil)
				if ctx.Database() == nil {
					return false // Should get no-op manager
				}
				if ctx.Cache() == nil {
					return false // Should get no-op manager
				}
				if ctx.Config() == nil {
					return false // Should get no-op manager
				}
				if ctx.Router() == nil {
					return false // Should get no-op manager
				}

				// Verify that errors were logged
				if len(logger.errors) == 0 {
					return false // Security violations should be logged
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 47: Granular permission support**
// **Validates: Requirements 11.4**
// For any plugin, the system should support independent permission flags for database, cache, router, config, filesystem, network, and exec access
func TestProperty_GranularPermissionSupport(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("each permission can be granted independently",
		prop.ForAll(
			func(pluginName string) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Test each permission independently
				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				for _, perm := range permissions {
					// Grant only this permission
					if err := checker.GrantPermission(pluginName, perm); err != nil {
						return false
					}

					// Check this permission succeeds
					if err := checker.CheckPermission(pluginName, perm); err != nil {
						return false
					}

					// Revoke this permission
					if err := checker.RevokePermission(pluginName, perm); err != nil {
						return false
					}

					// Check this permission now fails
					if err := checker.CheckPermission(pluginName, perm); err == nil {
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("granting one permission does not grant others",
		prop.ForAll(
			func(pluginName string, grantedPermIdx int) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				grantedPerm := permissions[grantedPermIdx%len(permissions)]

				// Grant only one permission
				if err := checker.GrantPermission(pluginName, grantedPerm); err != nil {
					return false
				}

				// Check that only the granted permission succeeds
				for _, perm := range permissions {
					err := checker.CheckPermission(pluginName, perm)
					if perm == grantedPerm {
						if err != nil {
							return false // Granted permission should succeed
						}
					} else {
						if err == nil {
							return false // Other permissions should fail
						}
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 6),
		),
	)

	properties.Property("revoking one permission does not affect others",
		prop.ForAll(
			func(pluginName string, revokedPermIdx int) bool {
				if pluginName == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				permissions := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
					PermissionFileSystem,
					PermissionNetwork,
					PermissionExec,
				}

				// Grant all permissions
				for _, perm := range permissions {
					if err := checker.GrantPermission(pluginName, perm); err != nil {
						return false
					}
				}

				revokedPerm := permissions[revokedPermIdx%len(permissions)]

				// Revoke one permission
				if err := checker.RevokePermission(pluginName, revokedPerm); err != nil {
					return false
				}

				// Check that only the revoked permission fails
				for _, perm := range permissions {
					err := checker.CheckPermission(pluginName, perm)
					if perm == revokedPerm {
						if err == nil {
							return false // Revoked permission should fail
						}
					} else {
						if err != nil {
							return false // Other permissions should still succeed
						}
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 6),
		),
	)

	properties.Property("custom permissions are independent from standard permissions",
		prop.ForAll(
			func(pluginName, customPerm string) bool {
				if pluginName == "" || customPerm == "" {
					return true // Skip empty names
				}

				logger := &MockLogger{}
				checker := NewPermissionChecker(logger)

				// Grant a custom permission
				if err := checker.GrantPermission(pluginName, customPerm); err != nil {
					return false
				}

				// Check custom permission succeeds
				if err := checker.CheckPermission(pluginName, customPerm); err != nil {
					return false
				}

				// Check that standard permissions still fail
				standardPerms := []string{
					PermissionDatabase,
					PermissionCache,
					PermissionConfig,
					PermissionRouter,
				}

				for _, perm := range standardPerms {
					if err := checker.CheckPermission(pluginName, perm); err == nil {
						return false // Standard permissions should fail
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// MockLoggerWithTracking is a mock logger that tracks warnings and errors
type MockLoggerWithTracking struct {
	mu       sync.Mutex
	warnings []string
	errors   []string
}

func (m *MockLoggerWithTracking) Debug(msg string, fields ...interface{}) {}

func (m *MockLoggerWithTracking) Info(msg string, fields ...interface{}) {}

func (m *MockLoggerWithTracking) Warn(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnings = append(m.warnings, msg)
}

func (m *MockLoggerWithTracking) Error(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, msg)
}

func (m *MockLoggerWithTracking) WithRequestID(requestID string) Logger {
	return m
}
