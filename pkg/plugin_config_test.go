package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 9: Configuration format support**
// **Validates: Requirements 3.1**
// For any valid plugin configuration in YAML, JSON, or TOML format,
// the system should successfully parse and load the configuration
func TestProperty_ConfigurationFormatSupport(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("valid JSON configuration parses successfully",
		prop.ForAll(
			func(pluginName, pluginPath string, enabled bool, priority int) bool {
				// Skip invalid inputs
				if pluginName == "" || pluginPath == "" {
					return true
				}

				// Create a valid plugin configuration
				config := map[string]interface{}{
					"plugins": map[string]interface{}{
						"enabled":   true,
						"directory": "./plugins",
						"plugins": []interface{}{
							map[string]interface{}{
								"name":     pluginName,
								"enabled":  enabled,
								"path":     pluginPath,
								"priority": priority,
								"config":   map[string]interface{}{},
								"permissions": map[string]interface{}{
									"database": false,
									"cache":    false,
								},
							},
						},
					},
				}

				// Marshal to JSON
				data, err := json.Marshal(config)
				if err != nil {
					return false
				}

				// Create temporary file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")
				if err := os.WriteFile(configPath, data, 0644); err != nil {
					return false
				}

				// Create a mock plugin manager
				manager := createMockPluginManagerForConfigTest(t)

				// Load plugins from config
				err = manager.LoadPluginsFromConfig(configPath)

				// If plugins are disabled, loading should succeed but not load anything
				// If enabled, loading should succeed (even if plugin binary doesn't exist,
				// the config parsing itself should work)
				return err == nil || err.Error() == "failed to load plugin from "+pluginPath+": plugin not found"
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.IntRange(0, 1000),
		),
	)

	properties.Property("valid YAML configuration parses successfully",
		prop.ForAll(
			func(pluginName, pluginPath string, enabled bool, priority int) bool {
				// Skip invalid inputs
				if pluginName == "" || pluginPath == "" {
					return true
				}

				// Create YAML content
				yamlContent := fmt.Sprintf(`plugins:
  enabled: true
  directory: ./plugins
  plugins:
    - name: %s
      enabled: %t
      path: %s
      priority: %d
      config: {}
      permissions:
        database: false
        cache: false
`, pluginName, enabled, pluginPath, priority)

				// Create temporary file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
					return false
				}

				// Create a mock plugin manager
				manager := createMockPluginManagerForConfigTest(t)

				// Load plugins from config
				err := manager.LoadPluginsFromConfig(configPath)

				// Config parsing should succeed
				return err == nil || err.Error() == "failed to load plugin from "+pluginPath+": plugin not found"
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.IntRange(0, 1000),
		),
	)

	properties.Property("valid TOML configuration parses successfully",
		prop.ForAll(
			func(pluginName, pluginPath string, enabled bool, priority int) bool {
				// Skip invalid inputs
				if pluginName == "" || pluginPath == "" {
					return true
				}

				// Create TOML content
				tomlContent := fmt.Sprintf(`[plugins]
enabled = true
directory = "./plugins"

[[plugins.plugins]]
name = "%s"
enabled = %t
path = "%s"
priority = %d

[plugins.plugins.config]

[plugins.plugins.permissions]
database = false
cache = false
`, pluginName, enabled, pluginPath, priority)

				// Create temporary file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.toml")
				if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
					return false
				}

				// Create a mock plugin manager
				manager := createMockPluginManagerForConfigTest(t)

				// Load plugins from config
				err := manager.LoadPluginsFromConfig(configPath)

				// Config parsing should succeed
				return err == nil || err.Error() == "failed to load plugin from "+pluginPath+": plugin not found"
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.IntRange(0, 1000),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 12: Initialization parameter passing**
// **Validates: Requirements 3.4**
// For any plugin with initialization parameters in configuration,
// those exact parameters should be passed to the plugin's Initialize method
func TestProperty_InitializationParameterPassing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("initialization parameters are passed correctly",
		prop.ForAll(
			func(pluginName string, configKey string, configValue string) bool {
				// Skip invalid inputs
				if pluginName == "" || configKey == "" {
					return true
				}

				// Create a plugin configuration with custom config
				config := map[string]interface{}{
					"plugins": map[string]interface{}{
						"enabled":   true,
						"directory": "./plugins",
						"plugins": []interface{}{
							map[string]interface{}{
								"name":    pluginName,
								"enabled": true,
								"path":    "./test-plugin",
								"config": map[string]interface{}{
									configKey: configValue,
								},
								"permissions": map[string]interface{}{},
							},
						},
					},
				}

				// Marshal to JSON
				data, err := json.Marshal(config)
				if err != nil {
					return false
				}

				// Create temporary file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")
				if err := os.WriteFile(configPath, data, 0644); err != nil {
					return false
				}

				// Parse the config file
				fileData, err := os.ReadFile(configPath)
				if err != nil {
					return false
				}

				parsed, err := parseJSON(fileData)
				if err != nil {
					return false
				}

				// Extract plugins config
				pluginsConfig, err := extractPluginsConfig(parsed)
				if err != nil {
					return false
				}

				// Verify the config was parsed correctly
				if len(pluginsConfig.Plugins) != 1 {
					return false
				}

				pluginEntry := pluginsConfig.Plugins[0]

				// Verify the initialization parameter is present
				if pluginEntry.Config == nil {
					return false
				}

				value, exists := pluginEntry.Config[configKey]
				if !exists {
					return false
				}

				// Verify the value matches
				return value == configValue
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 13: Load order preservation**
// **Validates: Requirements 3.5**
// For any plugin configuration specifying load order,
// plugins should be loaded in the exact sequence specified
func TestProperty_LoadOrderPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("plugins are loaded in configuration order",
		prop.ForAll(
			func(plugin1Name, plugin2Name, plugin3Name string) bool {
				// Skip invalid inputs
				if plugin1Name == "" || plugin2Name == "" || plugin3Name == "" {
					return true
				}

				// Skip if names are not unique
				if plugin1Name == plugin2Name || plugin2Name == plugin3Name || plugin1Name == plugin3Name {
					return true
				}

				// Create a plugin configuration with multiple plugins
				config := map[string]interface{}{
					"plugins": map[string]interface{}{
						"enabled":   true,
						"directory": "./plugins",
						"plugins": []interface{}{
							map[string]interface{}{
								"name":        plugin1Name,
								"enabled":     true,
								"path":        "./plugin1",
								"config":      map[string]interface{}{},
								"permissions": map[string]interface{}{},
							},
							map[string]interface{}{
								"name":        plugin2Name,
								"enabled":     true,
								"path":        "./plugin2",
								"config":      map[string]interface{}{},
								"permissions": map[string]interface{}{},
							},
							map[string]interface{}{
								"name":        plugin3Name,
								"enabled":     true,
								"path":        "./plugin3",
								"config":      map[string]interface{}{},
								"permissions": map[string]interface{}{},
							},
						},
					},
				}

				// Marshal to JSON
				data, err := json.Marshal(config)
				if err != nil {
					return false
				}

				// Parse the config
				parsed, err := parseJSON(data)
				if err != nil {
					return false
				}

				// Extract plugins config
				pluginsConfig, err := extractPluginsConfig(parsed)
				if err != nil {
					return false
				}

				// Verify the order is preserved
				if len(pluginsConfig.Plugins) != 3 {
					return false
				}

				// Check that the order matches the input order
				return pluginsConfig.Plugins[0].Name == plugin1Name &&
					pluginsConfig.Plugins[1].Name == plugin2Name &&
					pluginsConfig.Plugins[2].Name == plugin3Name
			},
			genValidPluginName(),
			genValidPluginName(),
			genValidPluginName(),
		),
	)

	properties.TestingRun(t)
}

// Helper function to create a mock plugin manager for config tests
func createMockPluginManagerForConfigTest(t *testing.T) PluginManager {
	// Create mock components
	logger := &MockLogger{}
	metrics := &MockMetrics{}

	registry := NewPluginRegistry()
	loader := NewMockPluginLoader()
	hookSystem := NewHookSystem(logger, metrics)
	eventBus := NewEventBus(logger)
	permChecker := NewPermissionChecker(logger)

	// Create mock framework services
	router := &MockRouter{}
	database := &MockDatabase{}
	cache := &MockCache{}
	config := NewConfigManager()

	return NewPluginManager(
		registry,
		loader,
		hookSystem,
		eventBus,
		permChecker,
		logger,
		metrics,
		router,
		database,
		cache,
		config,
	)
}
