//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Plugin Usage Example
// ============================================================================
//
// This example demonstrates:
// - Loading plugins from configuration files
// - Dynamic plugin loading at runtime
// - Plugin hot reloading without downtime
// - Plugin health monitoring
// - Plugin permissions management
//
// Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 3.1, 7.1, 7.5
// ============================================================================

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Plugin Usage Example")
	fmt.Println("================================================")
	fmt.Println()

	// ========================================================================
	// 1. Basic Framework Setup
	// ========================================================================
	fmt.Println("📦 Setting up framework...")

	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "plugin_example.db",
		},
	}

	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ========================================================================
	// 2. Loading Plugins from Configuration File
	// ========================================================================
	// The plugin system supports JSON, YAML, and TOML configuration formats
	fmt.Println("\n📂 Loading plugins from configuration...")

	// Load plugins from a configuration file
	// This demonstrates automatic plugin discovery and loading
	// Note: Plugin loading is configured during framework initialization
	// via FrameworkConfig.PluginConfigPath and FrameworkConfig.EnablePlugins

	pluginManager := app.PluginManager()
	if pluginManager != nil {
		fmt.Println("✓ Plugin system is enabled")

		// You can also load additional plugins at runtime
		configPath := "examples/plugin-config.yaml"
		if err := pluginManager.LoadPluginsFromConfig(configPath); err != nil {
			log.Printf("Warning: Failed to load plugins from config: %v", err)
			log.Println("Continuing without additional plugins...")
		} else {
			fmt.Printf("✓ Loaded plugins from %s\n", configPath)
		}
	} else {
		fmt.Println("⚠ Plugin system is not enabled")
		fmt.Println("  Enable it by setting EnablePlugins: true in FrameworkConfig")
	}

	// ========================================================================
	// 3. Dynamic Plugin Loading
	// ========================================================================
	// Plugins can also be loaded programmatically at runtime
	fmt.Println("\n🔌 Demonstrating dynamic plugin loading...")

	// Example: Dynamic plugin loading API
	// In a real scenario, you would use pluginManager.LoadPlugin()
	fmt.Println("✓ Dynamic plugin loading available via:")
	fmt.Println("  pluginManager.LoadPlugin(path, config)")
	fmt.Println("  - Load plugins at runtime")
	fmt.Println("  - Configure permissions and settings")
	fmt.Println("  - Initialize and start automatically")

	// ========================================================================
	// 4. Plugin Health Monitoring
	// ========================================================================
	// Monitor the health and performance of loaded plugins
	fmt.Println("\n🏥 Monitoring plugin health...")

	// Get health information for all plugins
	if pluginManager != nil {
		healthInfo := pluginManager.GetAllHealth()

		fmt.Printf("\n📊 Plugin Health Status:\n")
		fmt.Println("------------------------")

		if len(healthInfo) == 0 {
			fmt.Println("No plugins currently loaded")
		} else {
			for pluginName, health := range healthInfo {
				fmt.Printf("\nPlugin: %s\n", pluginName)
				fmt.Printf("  Status: %s\n", health.Status)
				fmt.Printf("  Error Count: %d\n", health.ErrorCount)

				if health.LastError != nil {
					fmt.Printf("  Last Error: %v\n", health.LastError)
					fmt.Printf("  Last Error At: %s\n", health.LastErrorAt.Format(time.RFC3339))
				}

				// Display hook metrics if available
				if len(health.HookMetrics) > 0 {
					fmt.Println("  Hook Metrics:")
					for hookName, metrics := range health.HookMetrics {
						fmt.Printf("    %s:\n", hookName)
						fmt.Printf("      Executions: %d\n", metrics.ExecutionCount)
						fmt.Printf("      Avg Duration: %v\n", metrics.AverageDuration)
						fmt.Printf("      Errors: %d\n", metrics.ErrorCount)
					}
				}
			}
		}
	}

	// ========================================================================
	// 5. List All Loaded Plugins
	// ========================================================================
	fmt.Println("\n📋 Listing all loaded plugins...")

	if pluginManager != nil {
		plugins := pluginManager.ListPlugins()

		if len(plugins) == 0 {
			fmt.Println("No plugins currently loaded")
		} else {
			fmt.Printf("\nLoaded Plugins (%d):\n", len(plugins))
			fmt.Println("-------------------")

			for _, info := range plugins {
				fmt.Printf("\n%s v%s\n", info.Name, info.Version)
				fmt.Printf("  Description: %s\n", info.Description)
				fmt.Printf("  Author: %s\n", info.Author)
				fmt.Printf("  Status: %s\n", info.Status)
				fmt.Printf("  Enabled: %t\n", info.Enabled)
				fmt.Printf("  Loaded At: %s\n", info.LoadTime.Format(time.RFC3339))
			}
		}
	}

	// ========================================================================
	// 6. Plugin Hot Reloading
	// ========================================================================
	// Demonstrate hot reloading of plugins without stopping the server
	fmt.Println("\n🔄 Demonstrating plugin hot reload capability...")

	// Hot reload allows updating plugins without downtime
	// The plugin manager handles:
	// - Graceful shutdown of the old plugin
	// - Loading the new plugin version
	// - Rollback on failure
	// - Request queuing during reload

	fmt.Println("✓ Hot reload API available via pluginManager.ReloadPlugin(name)")
	fmt.Println("  - Automatically handles graceful shutdown")
	fmt.Println("  - Queues requests during reload")
	fmt.Println("  - Rolls back on failure")

	// Example hot reload (commented out as it requires an actual plugin):
	// if pluginManager != nil {
	//     if err := pluginManager.ReloadPlugin("example-plugin"); err != nil {
	//         log.Printf("Failed to reload plugin: %v", err)
	//     } else {
	//         fmt.Println("✓ Plugin reloaded successfully")
	//     }
	// }

	// ========================================================================
	// 7. Plugin Permissions
	// ========================================================================
	// Demonstrate permission management for plugins
	fmt.Println("\n🔐 Plugin Permissions Management...")

	fmt.Println("\nAvailable Permissions:")
	fmt.Println("  - database: Access to DatabaseManager")
	fmt.Println("  - cache: Access to CacheManager")
	fmt.Println("  - router: Access to Router (register routes/middleware)")
	fmt.Println("  - config: Access to ConfigManager")
	fmt.Println("  - filesystem: Access to filesystem operations")
	fmt.Println("  - network: Access to network operations")
	fmt.Println("  - exec: Access to execute external commands")

	fmt.Println("\n✓ Permissions are enforced at runtime")
	fmt.Println("✓ Follow principle of least privilege")

	// ========================================================================
	// 8. Plugin Configuration Updates
	// ========================================================================
	// Demonstrate updating plugin configuration at runtime
	fmt.Println("\n⚙️  Runtime Configuration Updates...")

	// Plugins can receive configuration updates without reloading
	// The plugin's OnConfigChange() method is called
	fmt.Println("✓ Configuration updates via pluginManager.UpdatePluginConfig()")
	fmt.Println("  - Updates applied without restart")
	fmt.Println("  - Plugin validates new configuration")
	fmt.Println("  - Rollback on validation failure")

	// Example configuration update (commented out):
	// newConfig := map[string]interface{}{
	//     "timeout": "60s",
	//     "max_retries": 5,
	// }
	// if err := pluginManager.UpdatePluginConfig("example-plugin", newConfig); err != nil {
	//     log.Printf("Failed to update config: %v", err)
	// }

	// ========================================================================
	// 9. Setup Example Routes
	// ========================================================================
	fmt.Println("\n🛣️  Setting up example routes...")

	router := app.Router()

	// Health check endpoint
	router.GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Plugin status endpoint
	router.GET("/plugins", func(ctx pkg.Context) error {
		if pluginManager == nil {
			return ctx.JSON(200, map[string]interface{}{
				"plugins": []interface{}{},
			})
		}

		plugins := pluginManager.ListPlugins()
		return ctx.JSON(200, map[string]interface{}{
			"count":   len(plugins),
			"plugins": plugins,
		})
	})

	// Plugin health endpoint
	router.GET("/plugins/health", func(ctx pkg.Context) error {
		if pluginManager == nil {
			return ctx.JSON(200, map[string]interface{}{
				"health": map[string]interface{}{},
			})
		}

		health := pluginManager.GetAllHealth()
		return ctx.JSON(200, map[string]interface{}{
			"health": health,
		})
	})

	// Individual plugin health endpoint
	router.GET("/plugins/:name/health", func(ctx pkg.Context) error {
		params := ctx.Params()
		pluginName := params["name"]

		if pluginManager == nil {
			return ctx.JSON(404, map[string]interface{}{
				"error": "Plugin manager not available",
			})
		}

		health := pluginManager.GetPluginHealth(pluginName)
		return ctx.JSON(200, map[string]interface{}{
			"plugin": pluginName,
			"health": health,
		})
	})

	// Plugin reload endpoint (for demonstration)
	router.POST("/plugins/:name/reload", func(ctx pkg.Context) error {
		params := ctx.Params()
		pluginName := params["name"]

		if pluginManager == nil {
			return ctx.JSON(404, map[string]interface{}{
				"error": "Plugin manager not available",
			})
		}

		if err := pluginManager.ReloadPlugin(pluginName); err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": fmt.Sprintf("Failed to reload plugin: %v", err),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": fmt.Sprintf("Plugin %s reloaded successfully", pluginName),
		})
	})

	// ========================================================================
	// 10. Start Server
	// ========================================================================
	fmt.Println("\n🚀 Starting server...")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  http://localhost:8080/health")
	fmt.Println("  GET  http://localhost:8080/plugins")
	fmt.Println("  GET  http://localhost:8080/plugins/health")
	fmt.Println("  GET  http://localhost:8080/plugins/:name/health")
	fmt.Println("  POST http://localhost:8080/plugins/:name/reload")
	fmt.Println("\nExample commands:")
	fmt.Println("  curl http://localhost:8080/plugins")
	fmt.Println("  curl http://localhost:8080/plugins/health")
	fmt.Println("  curl -X POST http://localhost:8080/plugins/example-plugin/reload")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
