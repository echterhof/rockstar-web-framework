package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PluginConfigFile represents the structure of a plugin configuration file
type PluginConfigFile struct {
	Plugins PluginsConfig `json:"plugins" yaml:"plugins" toml:"plugins"`
}

// PluginsConfig represents the plugins section of the configuration
type PluginsConfig struct {
	// Enabled indicates whether the plugin system is enabled globally.
	// Default: true
	Enabled bool `json:"enabled" yaml:"enabled" toml:"enabled"`

	// Directory is the base directory for plugin files.
	// Default: "./plugins"
	Directory string `json:"directory" yaml:"directory" toml:"directory"`

	// Plugins is the list of plugin configuration entries.
	// Default: empty slice
	Plugins []PluginConfigEntry `json:"plugins" yaml:"plugins" toml:"plugins"`
}

// ApplyDefaults applies default values to PluginsConfig
func (c *PluginsConfig) ApplyDefaults() {
	// Default Enabled to true if not explicitly set
	// Note: We can't distinguish between false and unset for bool,
	// so we assume false means explicitly disabled

	// Default Directory to "./plugins" if empty
	if c.Directory == "" {
		c.Directory = "./plugins"
	}

	// Initialize Plugins to empty slice if nil
	if c.Plugins == nil {
		c.Plugins = []PluginConfigEntry{}
	}
}

// PluginConfigEntry represents a single plugin configuration entry
type PluginConfigEntry struct {
	// Name is the unique identifier for the plugin.
	// Required, no default
	Name string `json:"name" yaml:"name" toml:"name"`

	// Enabled indicates whether this plugin is enabled.
	// Default: true
	Enabled bool `json:"enabled" yaml:"enabled" toml:"enabled"`

	// Path is the file path to the plugin binary.
	// Required, no default
	Path string `json:"path" yaml:"path" toml:"path"`

	// Priority determines the plugin's execution order (lower values execute first).
	// Default: 0
	Priority int `json:"priority" yaml:"priority" toml:"priority"`

	// Config provides plugin-specific configuration as key-value pairs.
	// Default: empty map
	Config map[string]interface{} `json:"config" yaml:"config" toml:"config"`

	// Permissions defines what operations the plugin is allowed to perform.
	// Default: all false (secure by default)
	Permissions PluginPermissions `json:"permissions" yaml:"permissions" toml:"permissions"`
}

// LoadPluginsFromConfig loads plugins from a configuration file
func (m *pluginManagerImpl) LoadPluginsFromConfig(configPath string) error {
	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(configPath))

	var parsed map[string]interface{}
	switch ext {
	case ".json":
		parsed, err = parseJSON(data)
	case ".toml":
		parsed, err = parseTOML(data)
	case ".yaml", ".yml":
		parsed, err = parseYAML(data)
	default:
		return fmt.Errorf("unsupported config format: %s (supported: .json, .yaml, .yml, .toml)", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract plugins configuration
	pluginsConfig, err := extractPluginsConfig(parsed)
	if err != nil {
		return fmt.Errorf("failed to extract plugins config: %w", err)
	}

	// Apply defaults to plugins configuration
	pluginsConfig.ApplyDefaults()

	// Check if plugins are enabled globally
	if !pluginsConfig.Enabled {
		if m.logger != nil {
			m.logger.Info("Plugin system is disabled in configuration")
		}
		return nil
	}

	// Load plugins in the order specified in the configuration
	for _, pluginEntry := range pluginsConfig.Plugins {
		// Skip disabled plugins
		if !pluginEntry.Enabled {
			if m.logger != nil {
				m.logger.Info(fmt.Sprintf("Skipping disabled plugin: %s", pluginEntry.Name))
			}
			continue
		}

		// Resolve plugin path
		pluginPath := pluginEntry.Path
		if !filepath.IsAbs(pluginPath) {
			// If path is relative, resolve it relative to the config directory or plugins directory
			if pluginsConfig.Directory != "" {
				pluginPath = filepath.Join(pluginsConfig.Directory, pluginPath)
			} else {
				configDir := filepath.Dir(configPath)
				pluginPath = filepath.Join(configDir, pluginPath)
			}
		}

		// Create plugin config
		config := PluginConfig{
			Enabled:     pluginEntry.Enabled,
			Path:        pluginPath,
			Config:      pluginEntry.Config,
			Permissions: pluginEntry.Permissions,
			Priority:    pluginEntry.Priority,
		}

		// Apply defaults to plugin config
		config.ApplyDefaults()

		// Load the plugin
		if err := m.LoadPlugin(pluginPath, config); err != nil {
			if m.logger != nil {
				m.logger.Error(fmt.Sprintf("Failed to load plugin %s from %s: %v", pluginEntry.Name, pluginPath, err))
			}
			// Continue loading other plugins
			continue
		}

		if m.logger != nil {
			m.logger.Info(fmt.Sprintf("Loaded plugin %s from configuration", pluginEntry.Name))
		}
	}

	return nil
}

// extractPluginsConfig extracts the plugins configuration from parsed data
func extractPluginsConfig(parsed map[string]interface{}) (*PluginsConfig, error) {
	// Look for "plugins" key
	pluginsData, ok := parsed["plugins"]
	if !ok {
		return nil, fmt.Errorf("no 'plugins' section found in configuration")
	}

	pluginsMap, ok := pluginsData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'plugins' section is not a valid object")
	}

	config := &PluginsConfig{
		Enabled: true, // Default to enabled
		Plugins: []PluginConfigEntry{},
	}

	// Extract enabled flag
	if enabled, ok := pluginsMap["enabled"]; ok {
		if enabledBool, ok := enabled.(bool); ok {
			config.Enabled = enabledBool
		}
	}

	// Extract directory
	if directory, ok := pluginsMap["directory"]; ok {
		if dirStr, ok := directory.(string); ok {
			config.Directory = dirStr
		}
	}

	// Extract plugins list
	if pluginsList, ok := pluginsMap["plugins"]; ok {
		// Try to convert to []interface{} first
		pluginsArray, ok := pluginsList.([]interface{})
		if !ok {
			// TOML library might return []map[string]interface{} directly
			// Try to convert it
			if typedArray, ok := pluginsList.([]map[string]interface{}); ok {
				// Convert to []interface{}
				pluginsArray = make([]interface{}, len(typedArray))
				for i, v := range typedArray {
					pluginsArray[i] = v
				}
			} else {
				return nil, fmt.Errorf("'plugins.plugins' is not a valid array (type: %T)", pluginsList)
			}
		}

		for i, pluginData := range pluginsArray {
			pluginMap, ok := pluginData.(map[string]interface{})
			if !ok {
				// Skip invalid entries
				continue
			}

			entry, err := parsePluginEntry(pluginMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse plugin entry %d: %w", i, err)
			}

			config.Plugins = append(config.Plugins, entry)
		}
	}

	return config, nil
}

// parsePluginEntry parses a single plugin configuration entry
func parsePluginEntry(data map[string]interface{}) (PluginConfigEntry, error) {
	entry := PluginConfigEntry{
		Enabled:     true, // Default to enabled
		Config:      make(map[string]interface{}),
		Permissions: PluginPermissions{},
	}

	// Extract name (required)
	if name, ok := data["name"]; ok {
		if nameStr, ok := name.(string); ok {
			entry.Name = nameStr
		} else {
			return entry, fmt.Errorf("plugin name must be a string")
		}
	} else {
		return entry, fmt.Errorf("plugin name is required")
	}

	// Extract enabled flag
	if enabled, ok := data["enabled"]; ok {
		if enabledBool, ok := enabled.(bool); ok {
			entry.Enabled = enabledBool
		}
	}

	// Extract path (required)
	if path, ok := data["path"]; ok {
		if pathStr, ok := path.(string); ok {
			entry.Path = pathStr
		} else {
			return entry, fmt.Errorf("plugin path must be a string")
		}
	} else {
		return entry, fmt.Errorf("plugin path is required")
	}

	// Extract priority
	if priority, ok := data["priority"]; ok {
		switch p := priority.(type) {
		case int:
			entry.Priority = p
		case int64:
			entry.Priority = int(p)
		case float64:
			entry.Priority = int(p)
		}
	}

	// Extract config
	if config, ok := data["config"]; ok {
		if configMap, ok := config.(map[string]interface{}); ok {
			entry.Config = configMap
		}
	}

	// Extract permissions
	if permissions, ok := data["permissions"]; ok {
		if permMap, ok := permissions.(map[string]interface{}); ok {
			entry.Permissions = parsePermissions(permMap)
		}
	}

	return entry, nil
}

// parsePermissions parses plugin permissions from configuration
func parsePermissions(data map[string]interface{}) PluginPermissions {
	perms := PluginPermissions{
		CustomPermissions: make(map[string]bool),
	}

	if database, ok := data["database"]; ok {
		if dbBool, ok := database.(bool); ok {
			perms.AllowDatabase = dbBool
		}
	}

	if cache, ok := data["cache"]; ok {
		if cacheBool, ok := cache.(bool); ok {
			perms.AllowCache = cacheBool
		}
	}

	if config, ok := data["config"]; ok {
		if configBool, ok := config.(bool); ok {
			perms.AllowConfig = configBool
		}
	}

	if router, ok := data["router"]; ok {
		if routerBool, ok := router.(bool); ok {
			perms.AllowRouter = routerBool
		}
	}

	if filesystem, ok := data["filesystem"]; ok {
		if fsBool, ok := filesystem.(bool); ok {
			perms.AllowFileSystem = fsBool
		}
	}

	if network, ok := data["network"]; ok {
		if netBool, ok := network.(bool); ok {
			perms.AllowNetwork = netBool
		}
	}

	if exec, ok := data["exec"]; ok {
		if execBool, ok := exec.(bool); ok {
			perms.AllowExec = execBool
		}
	}

	// Extract custom permissions
	for key, value := range data {
		if key != "database" && key != "cache" && key != "config" &&
			key != "router" && key != "filesystem" && key != "network" && key != "exec" {
			if boolVal, ok := value.(bool); ok {
				perms.CustomPermissions[key] = boolVal
			}
		}
	}

	return perms
}
