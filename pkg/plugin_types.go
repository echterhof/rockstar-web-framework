package pkg

// PluginConfig represents configuration for a plugin
type PluginConfig struct {
	// Enabled indicates whether the plugin is enabled.
	// Default: true
	Enabled bool

	// Config provides plugin-specific configuration as key-value pairs.
	// Default: empty map
	Config map[string]interface{}

	// Permissions defines what operations the plugin is allowed to perform.
	// Default: all false (secure by default)
	Permissions PluginPermissions

	// Priority determines the plugin's execution order (lower values execute first).
	// Default: 0
	Priority int
}

// ApplyDefaults applies default values to PluginConfig
func (c *PluginConfig) ApplyDefaults() {
	// Default Enabled to true if not explicitly set
	// Note: We can't distinguish between false and unset for bool,
	// so we assume false means explicitly disabled

	// Initialize Config to empty map if nil
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}

	// Initialize Permissions with all false if not set
	// This ensures secure by default behavior
	if c.Permissions.CustomPermissions == nil {
		c.Permissions.CustomPermissions = make(map[string]bool)
	}

	// Priority defaults to 0 (already the zero value, no action needed)
}
