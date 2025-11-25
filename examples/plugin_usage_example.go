package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/rockstar/pkg"
)

// This example demonstrates how to use the plugin system with the example plugins

func main() {
	fmt.Println("Rockstar Framework - Plugin Usage Example")
	fmt.Println("==========================================")

	// Create a basic framework configuration
	config := &pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		LogLevel: "info",
	}

	// Create the framework instance
	framework, err := pkg.NewFramework(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	fmt.Println("\n1. Loading plugins...")

	// Load minimal plugin
	err = framework.PluginManager().LoadPlugin(
		"./examples/plugins/minimal-plugin",
		pkg.PluginConfig{
			Enabled: true,
			Config: map[string]interface{}{
				"message": "Hello from minimal plugin!",
			},
			Permissions: pkg.PluginPermissions{},
			Priority:    100,
		},
	)
	if err != nil {
		log.Printf("Failed to load minimal plugin: %v", err)
	} else {
		fmt.Println("   ✓ Minimal plugin loaded")
	}

	// Load logging plugin
	err = framework.PluginManager().LoadPlugin(
		"./examples/plugins/logging-plugin",
		pkg.PluginConfig{
			Enabled: true,
			Config: map[string]interface{}{
				"log_requests":  true,
				"log_responses": true,
				"log_headers":   false,
				"log_body":      false,
			},
			Permissions: pkg.PluginPermissions{},
			Priority:    50,
		},
	)
	if err != nil {
		log.Printf("Failed to load logging plugin: %v", err)
	} else {
		fmt.Println("   ✓ Logging plugin loaded")
	}

	// Load auth plugin
	err = framework.PluginManager().LoadPlugin(
		"./examples/plugins/auth-plugin",
		pkg.PluginConfig{
			Enabled: true,
			Config: map[string]interface{}{
				"token_duration": "2h",
				"require_auth":   true,
				"excluded_paths": []interface{}{"/health", "/public"},
			},
			Permissions: pkg.PluginPermissions{
				AllowRouter: true,
			},
			Priority: 200,
		},
	)
	if err != nil {
		log.Printf("Failed to load auth plugin: %v", err)
	} else {
		fmt.Println("   ✓ Auth plugin loaded")
	}

	// Load cache plugin
	err = framework.PluginManager().LoadPlugin(
		"./examples/plugins/cache-plugin",
		pkg.PluginConfig{
			Enabled: true,
			Config: map[string]interface{}{
				"cache_duration": "10m",
				"enabled":        true,
				"cache_methods":  []interface{}{"GET"},
				"excluded_paths": []interface{}{"/admin", "/api/auth"},
			},
			Permissions: pkg.PluginPermissions{
				AllowCache:  true,
				AllowRouter: true,
			},
			Priority: 150,
		},
	)
	if err != nil {
		log.Printf("Failed to load cache plugin: %v", err)
	} else {
		fmt.Println("   ✓ Cache plugin loaded")
	}

	fmt.Println("\n2. Initializing all plugins...")

	// Initialize all plugins
	if err := framework.PluginManager().InitializeAll(); err != nil {
		log.Fatalf("Failed to initialize plugins: %v", err)
	}
	fmt.Println("   ✓ All plugins initialized")

	fmt.Println("\n3. Starting all plugins...")

	// Start all plugins
	if err := framework.PluginManager().StartAll(); err != nil {
		log.Fatalf("Failed to start plugins: %v", err)
	}
	fmt.Println("   ✓ All plugins started")

	fmt.Println("\n4. Listing loaded plugins...")

	// List all loaded plugins
	plugins := framework.PluginManager().ListPlugins()
	for _, plugin := range plugins {
		fmt.Printf("   - %s v%s (%s)\n", plugin.Name, plugin.Version, plugin.Status)
	}

	fmt.Println("\n5. Checking plugin health...")

	// Get health status for all plugins
	healthMap := framework.PluginManager().GetAllHealth()
	for name, health := range healthMap {
		fmt.Printf("   - %s: %s (errors: %d)\n", name, health.Status, health.ErrorCount)
	}

	fmt.Println("\n6. Demonstrating inter-plugin communication...")

	// Get the logging plugin and access its exported service
	loggingPlugin, err := framework.PluginManager().GetPlugin("logging-plugin")
	if err == nil {
		fmt.Println("   ✓ Logging plugin found")
		// In a real scenario, you would import the service here
	}

	// Get the cache plugin and access its exported service
	cachePlugin, err := framework.PluginManager().GetPlugin("cache-plugin")
	if err == nil {
		fmt.Println("   ✓ Cache plugin found")
		// In a real scenario, you would import the service here
	}

	fmt.Println("\n7. Simulating plugin hot reload...")

	// Hot reload the minimal plugin
	if err := framework.PluginManager().ReloadPlugin("minimal-plugin"); err != nil {
		log.Printf("Failed to reload minimal plugin: %v", err)
	} else {
		fmt.Println("   ✓ Minimal plugin reloaded")
	}

	// Wait a bit to see plugin activity
	fmt.Println("\n8. Running for 5 seconds...")
	time.Sleep(5 * time.Second)

	fmt.Println("\n9. Stopping all plugins...")

	// Stop all plugins
	if err := framework.PluginManager().StopAll(); err != nil {
		log.Fatalf("Failed to stop plugins: %v", err)
	}
	fmt.Println("   ✓ All plugins stopped")

	fmt.Println("\n10. Unloading plugins...")

	// Unload each plugin
	for _, plugin := range plugins {
		if err := framework.PluginManager().UnloadPlugin(plugin.Name); err != nil {
			log.Printf("Failed to unload %s: %v", plugin.Name, err)
		} else {
			fmt.Printf("   ✓ %s unloaded\n", plugin.Name)
		}
	}

	fmt.Println("\n==========================================")
	fmt.Println("Plugin usage example completed!")
}
