package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// configManager implements the ConfigManager interface
type configManager struct {
	mu            sync.RWMutex
	data          map[string]interface{}
	configPath    string
	env           string
	watchCallback func()
	stopWatch     chan struct{}
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() ConfigManager {
	return &configManager{
		data:      make(map[string]interface{}),
		env:       getEnvironment(),
		stopWatch: make(chan struct{}),
	}
}

// Load loads configuration from a file (supports INI, TOML, YAML)
func (c *configManager) Load(configPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configPath = configPath

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(configPath))

	var parsed map[string]interface{}
	switch ext {
	case ".ini":
		parsed, err = parseINI(data)
	case ".toml":
		parsed, err = parseTOML(data)
	case ".yaml", ".yml":
		parsed, err = parseYAML(data)
	default:
		return fmt.Errorf("unsupported config format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	c.data = parsed
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (c *configManager) LoadFromEnv() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Load all environment variables with a specific prefix
	prefix := "ROCKSTAR_"
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := pair[1]

		// Only process variables with our prefix
		if strings.HasPrefix(key, prefix) {
			// Remove prefix and convert to lowercase with dots
			configKey := strings.ToLower(strings.TrimPrefix(key, prefix))
			configKey = strings.ReplaceAll(configKey, "_", ".")

			// Try to parse as different types
			c.data[configKey] = parseValue(value)
		}
	}

	return nil
}

// Reload reloads the configuration from the original source
func (c *configManager) Reload() error {
	if c.configPath == "" {
		return fmt.Errorf("no config path set")
	}
	return c.Load(c.configPath)
}

// Get retrieves a configuration value
func (c *configManager) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getNestedValue(key)
}

// GetString retrieves a string configuration value
func (c *configManager) GetString(key string) string {
	val := c.Get(key)
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

// GetInt retrieves an integer configuration value
func (c *configManager) GetInt(key string) int {
	val := c.Get(key)
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

// GetInt64 retrieves an int64 configuration value
func (c *configManager) GetInt64(key string) int64 {
	val := c.Get(key)
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		return 0
	}
}

// GetFloat64 retrieves a float64 configuration value
func (c *configManager) GetFloat64(key string) float64 {
	val := c.Get(key)
	if val == nil {
		return 0.0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0.0
	}
}

// GetBool retrieves a boolean configuration value
func (c *configManager) GetBool(key string) bool {
	val := c.Get(key)
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)
		return b
	case int, int64:
		return v != 0
	default:
		return false
	}
}

// GetDuration retrieves a duration configuration value
func (c *configManager) GetDuration(key string) time.Duration {
	val := c.Get(key)
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case time.Duration:
		return v
	case string:
		d, _ := time.ParseDuration(v)
		return d
	case int64:
		return time.Duration(v)
	case int:
		return time.Duration(v)
	default:
		return 0
	}
}

// GetStringSlice retrieves a string slice configuration value
func (c *configManager) GetStringSlice(key string) []string {
	val := c.Get(key)
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case string:
		// Split by comma
		return strings.Split(v, ",")
	default:
		return nil
	}
}

// GetWithDefault retrieves a value with a default fallback
func (c *configManager) GetWithDefault(key string, defaultValue interface{}) interface{} {
	val := c.Get(key)
	if val == nil {
		return defaultValue
	}
	return val
}

// GetStringWithDefault retrieves a string with a default fallback
func (c *configManager) GetStringWithDefault(key string, defaultValue string) string {
	val := c.GetString(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetIntWithDefault retrieves an int with a default fallback
func (c *configManager) GetIntWithDefault(key string, defaultValue int) int {
	if !c.IsSet(key) {
		return defaultValue
	}
	return c.GetInt(key)
}

// GetBoolWithDefault retrieves a bool with a default fallback
func (c *configManager) GetBoolWithDefault(key string, defaultValue bool) bool {
	if !c.IsSet(key) {
		return defaultValue
	}
	return c.GetBool(key)
}

// Validate validates the configuration
func (c *configManager) Validate() error {
	// Basic validation - can be extended
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.data) == 0 {
		return fmt.Errorf("configuration is empty")
	}

	return nil
}

// IsSet checks if a configuration key is set
func (c *configManager) IsSet(key string) bool {
	return c.Get(key) != nil
}

// GetEnv returns the current environment
func (c *configManager) GetEnv() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.env
}

// IsProduction checks if running in production
func (c *configManager) IsProduction() bool {
	return c.GetEnv() == "production"
}

// IsDevelopment checks if running in development
func (c *configManager) IsDevelopment() bool {
	return c.GetEnv() == "development"
}

// IsTest checks if running in test mode
func (c *configManager) IsTest() bool {
	return c.GetEnv() == "test"
}

// Sub returns a sub-configuration manager for a nested key
func (c *configManager) Sub(key string) ConfigManager {
	val := c.Get(key)
	if val == nil {
		return NewConfigManager()
	}

	subData, ok := val.(map[string]interface{})
	if !ok {
		return NewConfigManager()
	}

	sub := &configManager{
		data: subData,
		env:  c.env,
	}

	return sub
}

// Watch watches for configuration changes
func (c *configManager) Watch(callback func()) error {
	c.mu.Lock()
	c.watchCallback = callback
	c.mu.Unlock()

	// Start watching in a goroutine
	go c.watchLoop()

	return nil
}

// StopWatching stops watching for configuration changes
func (c *configManager) StopWatching() error {
	close(c.stopWatch)
	return nil
}

// Helper methods

func (c *configManager) getNestedValue(key string) interface{} {
	// First try direct key lookup (for flat keys with dots)
	if val, ok := c.data[key]; ok {
		return val
	}

	// Then try nested lookup
	parts := strings.Split(key, ".")
	current := c.data

	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return nil
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return val
		}

		// Otherwise, continue traversing
		nested, ok := val.(map[string]interface{})
		if !ok {
			return nil
		}
		current = nested
	}

	return nil
}

func (c *configManager) watchLoop() {
	if c.configPath == "" {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time
	info, err := os.Stat(c.configPath)
	if err == nil {
		lastModTime = info.ModTime()
	}

	for {
		select {
		case <-c.stopWatch:
			return
		case <-ticker.C:
			info, err := os.Stat(c.configPath)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				lastModTime = info.ModTime()
				if err := c.Reload(); err == nil {
					c.mu.RLock()
					callback := c.watchCallback
					c.mu.RUnlock()

					if callback != nil {
						callback()
					}
				}
			}
		}
	}
}

func getEnvironment() string {
	env := os.Getenv("ROCKSTAR_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func parseValue(value string) interface{} {
	// Try to parse as bool
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}

	// Try to parse as int
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Try to parse as duration
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}

	// Return as string
	return value
}
