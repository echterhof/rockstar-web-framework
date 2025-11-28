package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// mockPlugin is a simple plugin implementation for testing
type mockPlugin struct {
	name         string
	version      string
	description  string
	author       string
	dependencies []PluginDependency
	configSchema map[string]interface{}
}

func (p *mockPlugin) Name() string        { return p.name }
func (p *mockPlugin) Version() string     { return p.version }
func (p *mockPlugin) Description() string { return p.description }
func (p *mockPlugin) Author() string      { return p.author }
func (p *mockPlugin) Dependencies() []PluginDependency {
	return p.dependencies
}
func (p *mockPlugin) Initialize(ctx PluginContext) error                 { return nil }
func (p *mockPlugin) Start() error                                       { return nil }
func (p *mockPlugin) Stop() error                                        { return nil }
func (p *mockPlugin) Cleanup() error                                     { return nil }
func (p *mockPlugin) ConfigSchema() map[string]interface{}               { return p.configSchema }
func (p *mockPlugin) OnConfigChange(config map[string]interface{}) error { return nil }

// TestProperty_PluginDiscoveryCompleteness tests Property 1:
// Plugin Discovery Completeness
// **Feature: compile-time-plugins, Property 1: Plugin Discovery Completeness**
// **Validates: Requirements 1.2, 3.2**
func TestProperty_PluginDiscoveryCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all registered plugins are discoverable", prop.ForAll(
		func(pluginNames []string) bool {
			// Create a fresh registry for this test
			registry := newCompileTimeRegistry()

			// Register all plugins
			for _, name := range pluginNames {
				factory := func(n string) PluginFactory {
					return func() Plugin {
						return &mockPlugin{
							name:    n,
							version: "1.0.0",
						}
					}
				}(name)
				registry.Register(name, factory)
			}

			// Get list of registered plugins
			registered := registry.ListPlugins()

			// Create a map for quick lookup
			registeredMap := make(map[string]bool)
			for _, name := range registered {
				registeredMap[name] = true
			}

			// Verify all original plugins are in the registered list
			for _, name := range pluginNames {
				if !registeredMap[name] {
					t.Logf("Plugin %s was registered but not found in ListPlugins()", name)
					return false
				}
			}

			// Verify the count matches
			if len(registered) != len(pluginNames) {
				t.Logf("Expected %d plugins, but found %d", len(pluginNames), len(registered))
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique plugin names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return true
		}),
	))

	properties.Property("registered plugins can be created", prop.ForAll(
		func(pluginNames []string) bool {
			// Create a fresh registry for this test
			registry := newCompileTimeRegistry()

			// Register all plugins
			for _, name := range pluginNames {
				factory := func(n string) PluginFactory {
					return func() Plugin {
						return &mockPlugin{
							name:    n,
							version: "1.0.0",
						}
					}
				}(name)
				registry.Register(name, factory)
			}

			// Try to create each plugin
			for _, name := range pluginNames {
				factory, ok := registry.GetFactory(name)
				if !ok {
					t.Logf("Plugin %s was registered but factory not found", name)
					return false
				}

				plugin := factory()
				if plugin == nil {
					t.Logf("Factory for plugin %s returned nil", name)
					return false
				}

				if plugin.Name() != name {
					t.Logf("Plugin name mismatch: expected %s, got %s", name, plugin.Name())
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique plugin names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return true
		}),
	))

	properties.Property("registry is thread-safe for concurrent registration", prop.ForAll(
		func(pluginNames []string) bool {
			// Create a fresh registry for this test
			registry := newCompileTimeRegistry()

			// Register plugins concurrently
			done := make(chan bool, len(pluginNames))
			for _, name := range pluginNames {
				go func(n string) {
					factory := func(pluginName string) PluginFactory {
						return func() Plugin {
							return &mockPlugin{
								name:    pluginName,
								version: "1.0.0",
							}
						}
					}(n)
					registry.Register(n, factory)
					done <- true
				}(name)
			}

			// Wait for all registrations to complete
			for i := 0; i < len(pluginNames); i++ {
				<-done
			}

			// Verify all plugins are registered
			registered := registry.ListPlugins()
			if len(registered) != len(pluginNames) {
				t.Logf("Expected %d plugins after concurrent registration, got %d",
					len(pluginNames), len(registered))
				return false
			}

			return true
		},
		gen.SliceOfN(20, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique plugin names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return true
		}),
	))

	properties.Property("GetFactory returns false for unregistered plugins", prop.ForAll(
		func(registeredNames []string, unregisteredName string) bool {
			// Create a fresh registry for this test
			registry := newCompileTimeRegistry()

			// Register plugins
			for _, name := range registeredNames {
				factory := func(n string) PluginFactory {
					return func() Plugin {
						return &mockPlugin{
							name:    n,
							version: "1.0.0",
						}
					}
				}(name)
				registry.Register(name, factory)
			}

			// Check if unregistered name is actually unregistered
			for _, name := range registeredNames {
				if name == unregisteredName {
					// Skip this test case if the unregistered name is actually registered
					return true
				}
			}

			// Try to get factory for unregistered plugin
			_, ok := registry.GetFactory(unregisteredName)
			if ok {
				t.Logf("GetFactory returned true for unregistered plugin %s", unregisteredName)
				return false
			}

			return true
		},
		gen.SliceOfN(10, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique plugin names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return true
		}),
		gen.Const("unregistered_plugin_xyz_12345"),
	))

	properties.Property("global registry functions work correctly", prop.ForAll(
		func(pluginNames []string, testID int) bool {
			// Note: We can't fully test the global registry in isolation
			// because it's shared across tests, but we can test the API

			// Create unique names using the test ID
			uniqueNames := make([]string, len(pluginNames))
			for i, name := range pluginNames {
				uniqueNames[i] = fmt.Sprintf("test_%d_%s", testID, name)
			}

			// Register plugins using the global API
			for _, uniqueName := range uniqueNames {
				RegisterPlugin(uniqueName, func(n string) PluginFactory {
					return func() Plugin {
						return &mockPlugin{
							name:    n,
							version: "1.0.0",
						}
					}
				}(uniqueName))
			}

			// Verify we can retrieve the registered plugins
			allPlugins := GetRegisteredPlugins()
			if len(allPlugins) == 0 {
				t.Logf("GetRegisteredPlugins returned empty list")
				return false
			}

			// Try to create plugins using the global API
			for _, uniqueName := range uniqueNames {
				plugin, err := CreatePlugin(uniqueName)
				if err != nil {
					t.Logf("CreatePlugin failed for %s: %v", uniqueName, err)
					return false
				}
				if plugin == nil {
					t.Logf("CreatePlugin returned nil for %s", uniqueName)
					return false
				}
			}

			return true
		},
		gen.SliceOfN(5, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique plugin names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}),
		gen.Int(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
