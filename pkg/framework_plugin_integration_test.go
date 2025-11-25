package pkg

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestFrameworkPluginIntegration tests the integration of the plugin system with the framework
func TestFrameworkPluginIntegration(t *testing.T) {
	t.Run("Framework initializes with plugin system disabled", func(t *testing.T) {
		config := FrameworkConfig{
			DatabaseConfig: DatabaseConfig{
				Driver:   "sqlite3",
				Database: ":memory:",
			},
			CacheConfig: CacheConfig{
				DefaultTTL: 5 * time.Minute,
			},
			SessionConfig: SessionConfig{
				CookieName:      "session_id",
				SessionLifetime: 3600 * time.Second,
				EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
				CleanupInterval: 10 * time.Minute,
			},
			EnablePlugins: false,
		}

		framework, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create framework: %v", err)
		}

		if framework.PluginManager() != nil {
			t.Error("Expected plugin manager to be nil when plugins are disabled")
		}
	})

	t.Run("Framework initializes with plugin system enabled", func(t *testing.T) {
		config := FrameworkConfig{
			DatabaseConfig: DatabaseConfig{
				Driver:   "sqlite3",
				Database: ":memory:",
			},
			CacheConfig: CacheConfig{
				DefaultTTL: 5 * time.Minute,
			},
			SessionConfig: SessionConfig{
				CookieName:      "session_id",
				SessionLifetime: 3600 * time.Second,
				EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
				CleanupInterval: 10 * time.Minute,
			},
			EnablePlugins: true,
		}

		framework, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create framework: %v", err)
		}

		if framework.PluginManager() == nil {
			t.Error("Expected plugin manager to be initialized when plugins are enabled")
		}
	})

	t.Run("Framework startup hooks execute plugin hooks", func(t *testing.T) {
		config := FrameworkConfig{
			DatabaseConfig: DatabaseConfig{
				Driver:   "sqlite3",
				Database: ":memory:",
			},
			CacheConfig: CacheConfig{
				DefaultTTL: 5 * time.Minute,
			},
			SessionConfig: SessionConfig{
				CookieName:      "session_id",
				SessionLifetime: 3600 * time.Second,
				EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
				CleanupInterval: 10 * time.Minute,
			},
			EnablePlugins: true,
		}

		framework, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create framework: %v", err)
		}

		// Create a test plugin
		plugin := &LifecycleTrackingPlugin{
			name:           "test-plugin",
			version:        "1.0.0",
			mu:             &sync.Mutex{},
			lifecycleSteps: []string{},
			shouldFail:     false,
		}

		// Load the plugin
		pluginConfig := PluginConfig{
			Enabled: true,
			Path:    "/test/plugin",
			Config:  make(map[string]interface{}),
			Permissions: PluginPermissions{
				AllowRouter: true,
			},
		}

		// Manually add plugin to manager for testing
		if pm, ok := framework.PluginManager().(*pluginManagerImpl); ok {
			entry := &pluginEntry{
				plugin:   plugin,
				config:   pluginConfig,
				status:   PluginStatusLoading,
				loadTime: time.Now(),
				enabled:  true,
			}
			pm.mu.Lock()
			pm.plugins[plugin.Name()] = entry
			pm.mu.Unlock()

			// Initialize the plugin
			if err := pm.initializePlugin(entry); err != nil {
				t.Fatalf("Failed to initialize plugin: %v", err)
			}

			// Register a startup hook
			hookCalled := false
			pm.hookSystem.RegisterHook(plugin.Name(), HookTypeStartup, 100, func(ctx HookContext) error {
				hookCalled = true
				return nil
			})

			// Execute startup hooks
			ctx := context.Background()
			if err := framework.executeStartupHooks(ctx); err != nil {
				t.Fatalf("Failed to execute startup hooks: %v", err)
			}

			if !hookCalled {
				t.Error("Expected startup hook to be called")
			}
		}
	})

	t.Run("Framework shutdown hooks execute plugin hooks", func(t *testing.T) {
		config := FrameworkConfig{
			DatabaseConfig: DatabaseConfig{
				Driver:   "sqlite3",
				Database: ":memory:",
			},
			CacheConfig: CacheConfig{
				DefaultTTL: 5 * time.Minute,
			},
			SessionConfig: SessionConfig{
				CookieName:      "session_id",
				SessionLifetime: 3600 * time.Second,
				EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
				CleanupInterval: 10 * time.Minute,
			},
			EnablePlugins: true,
		}

		framework, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create framework: %v", err)
		}

		// Create a test plugin
		plugin := &LifecycleTrackingPlugin{
			name:           "test-plugin",
			version:        "1.0.0",
			mu:             &sync.Mutex{},
			lifecycleSteps: []string{},
			shouldFail:     false,
		}

		// Load the plugin
		pluginConfig := PluginConfig{
			Enabled: true,
			Path:    "/test/plugin",
			Config:  make(map[string]interface{}),
			Permissions: PluginPermissions{
				AllowRouter: true,
			},
		}

		// Manually add plugin to manager for testing
		if pm, ok := framework.PluginManager().(*pluginManagerImpl); ok {
			entry := &pluginEntry{
				plugin:   plugin,
				config:   pluginConfig,
				status:   PluginStatusRunning,
				loadTime: time.Now(),
				enabled:  true,
			}
			pm.mu.Lock()
			pm.plugins[plugin.Name()] = entry
			pm.mu.Unlock()

			// Register a shutdown hook
			hookCalled := false
			pm.hookSystem.RegisterHook(plugin.Name(), HookTypeShutdown, 100, func(ctx HookContext) error {
				hookCalled = true
				return nil
			})

			// Execute shutdown hooks
			ctx := context.Background()
			if err := framework.executeShutdownHooks(ctx); err != nil {
				t.Fatalf("Failed to execute shutdown hooks: %v", err)
			}

			if !hookCalled {
				t.Error("Expected shutdown hook to be called")
			}
		}
	})

	t.Run("Framework exposes plugin manager", func(t *testing.T) {
		config := FrameworkConfig{
			DatabaseConfig: DatabaseConfig{
				Driver:   "sqlite3",
				Database: ":memory:",
			},
			CacheConfig: CacheConfig{
				DefaultTTL: 5 * time.Minute,
			},
			SessionConfig: SessionConfig{
				CookieName:      "session_id",
				SessionLifetime: 3600 * time.Second,
				EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
				CleanupInterval: 10 * time.Minute,
			},
			EnablePlugins: true,
		}

		framework, err := New(config)
		if err != nil {
			t.Fatalf("Failed to create framework: %v", err)
		}

		pluginManager := framework.PluginManager()
		if pluginManager == nil {
			t.Fatal("Expected plugin manager to be available")
		}

		// Verify we can call plugin manager methods
		plugins := pluginManager.ListPlugins()
		if plugins == nil {
			t.Error("Expected plugins list to be non-nil")
		}
	})
}
