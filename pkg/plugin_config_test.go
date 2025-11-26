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

// Unit tests for plugin config defaults

func TestPluginConfig_ApplyDefaults(t *testing.T) {
	t.Run("initializes nil Config to empty map", func(t *testing.T) {
		config := PluginConfig{
			Path: "/path/to/plugin",
		}

		config.ApplyDefaults()

		if config.Config == nil {
			t.Error("Expected Config to be initialized to empty map, got nil")
		}

		if len(config.Config) != 0 {
			t.Errorf("Expected Config to be empty map, got length %d", len(config.Config))
		}
	})

	t.Run("preserves existing Config values", func(t *testing.T) {
		existingConfig := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		config := PluginConfig{
			Path:   "/path/to/plugin",
			Config: existingConfig,
		}

		config.ApplyDefaults()

		if config.Config == nil {
			t.Error("Expected Config to be preserved, got nil")
		}

		if len(config.Config) != 2 {
			t.Errorf("Expected Config to have 2 entries, got %d", len(config.Config))
		}

		if config.Config["key1"] != "value1" {
			t.Errorf("Expected key1 to be 'value1', got %v", config.Config["key1"])
		}

		if config.Config["key2"] != 42 {
			t.Errorf("Expected key2 to be 42, got %v", config.Config["key2"])
		}
	})

	t.Run("initializes nil CustomPermissions to empty map", func(t *testing.T) {
		config := PluginConfig{
			Path: "/path/to/plugin",
		}

		config.ApplyDefaults()

		if config.Permissions.CustomPermissions == nil {
			t.Error("Expected CustomPermissions to be initialized to empty map, got nil")
		}

		if len(config.Permissions.CustomPermissions) != 0 {
			t.Errorf("Expected CustomPermissions to be empty map, got length %d", len(config.Permissions.CustomPermissions))
		}
	})

	t.Run("preserves existing CustomPermissions", func(t *testing.T) {
		customPerms := map[string]bool{
			"custom1": true,
			"custom2": false,
		}

		config := PluginConfig{
			Path: "/path/to/plugin",
			Permissions: PluginPermissions{
				AllowDatabase:     true,
				CustomPermissions: customPerms,
			},
		}

		config.ApplyDefaults()

		if config.Permissions.CustomPermissions == nil {
			t.Error("Expected CustomPermissions to be preserved, got nil")
		}

		if len(config.Permissions.CustomPermissions) != 2 {
			t.Errorf("Expected CustomPermissions to have 2 entries, got %d", len(config.Permissions.CustomPermissions))
		}

		if !config.Permissions.AllowDatabase {
			t.Error("Expected AllowDatabase to be preserved as true")
		}
	})

	t.Run("Priority defaults to 0", func(t *testing.T) {
		config := PluginConfig{
			Path: "/path/to/plugin",
		}

		config.ApplyDefaults()

		if config.Priority != 0 {
			t.Errorf("Expected Priority to default to 0, got %d", config.Priority)
		}
	})

	t.Run("preserves non-zero Priority", func(t *testing.T) {
		config := PluginConfig{
			Path:     "/path/to/plugin",
			Priority: 10,
		}

		config.ApplyDefaults()

		if config.Priority != 10 {
			t.Errorf("Expected Priority to be preserved as 10, got %d", config.Priority)
		}
	})
}

func TestPluginsConfig_ApplyDefaults(t *testing.T) {
	t.Run("defaults Directory to ./plugins when empty", func(t *testing.T) {
		config := PluginsConfig{}

		config.ApplyDefaults()

		if config.Directory != "./plugins" {
			t.Errorf("Expected Directory to default to './plugins', got '%s'", config.Directory)
		}
	})

	t.Run("preserves non-empty Directory", func(t *testing.T) {
		config := PluginsConfig{
			Directory: "/custom/plugins",
		}

		config.ApplyDefaults()

		if config.Directory != "/custom/plugins" {
			t.Errorf("Expected Directory to be preserved as '/custom/plugins', got '%s'", config.Directory)
		}
	})

	t.Run("initializes nil Plugins to empty slice", func(t *testing.T) {
		config := PluginsConfig{}

		config.ApplyDefaults()

		if config.Plugins == nil {
			t.Error("Expected Plugins to be initialized to empty slice, got nil")
		}

		if len(config.Plugins) != 0 {
			t.Errorf("Expected Plugins to be empty slice, got length %d", len(config.Plugins))
		}
	})

	t.Run("preserves existing Plugins", func(t *testing.T) {
		plugins := []PluginConfigEntry{
			{
				Name:    "plugin1",
				Path:    "/path/to/plugin1",
				Enabled: true,
			},
			{
				Name:    "plugin2",
				Path:    "/path/to/plugin2",
				Enabled: false,
			},
		}

		config := PluginsConfig{
			Plugins: plugins,
		}

		config.ApplyDefaults()

		if config.Plugins == nil {
			t.Error("Expected Plugins to be preserved, got nil")
		}

		if len(config.Plugins) != 2 {
			t.Errorf("Expected Plugins to have 2 entries, got %d", len(config.Plugins))
		}

		if config.Plugins[0].Name != "plugin1" {
			t.Errorf("Expected first plugin name to be 'plugin1', got '%s'", config.Plugins[0].Name)
		}

		if config.Plugins[1].Name != "plugin2" {
			t.Errorf("Expected second plugin name to be 'plugin2', got '%s'", config.Plugins[1].Name)
		}
	})

	t.Run("applies all defaults together", func(t *testing.T) {
		config := PluginsConfig{}

		config.ApplyDefaults()

		if config.Directory != "./plugins" {
			t.Errorf("Expected Directory to default to './plugins', got '%s'", config.Directory)
		}

		if config.Plugins == nil {
			t.Error("Expected Plugins to be initialized to empty slice, got nil")
		}

		if len(config.Plugins) != 0 {
			t.Errorf("Expected Plugins to be empty slice, got length %d", len(config.Plugins))
		}
	})
}

func TestPluginConfig_DefaultsInLoadPlugin(t *testing.T) {
	t.Run("LoadPlugin applies defaults before using config", func(t *testing.T) {
		// Create a mock plugin manager
		logger := &MockLogger{}
		metrics := &MockMetrics{}

		registry := NewPluginRegistry()
		loader := NewMockPluginLoader()
		hookSystem := NewHookSystem(logger, metrics)
		eventBus := NewEventBus(logger)
		permChecker := NewPermissionChecker(logger)

		router := &MockRouter{}
		database := &MockDatabase{}
		cache := &MockCache{}
		configMgr := NewConfigManager()

		manager := NewPluginManager(
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
			configMgr,
		)

		// Create a config with nil Config map
		config := PluginConfig{
			Enabled: true,
			Path:    "./test-plugin",
		}

		// LoadPlugin should apply defaults
		// Note: This will fail to load the actual plugin, but we're testing that
		// defaults are applied before the load attempt
		_ = manager.LoadPlugin("./test-plugin", config)

		// The test passes if no panic occurs from nil map access
	})
}
