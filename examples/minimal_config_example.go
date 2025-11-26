package main

import (
	"fmt"
	"log"

	"github.com/echterhof/rockstar-web-framework/pkg"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// This example demonstrates minimal framework configuration using defaults.
// The Rockstar Web Framework provides sensible defaults for all configuration
// values, so you only need to specify what you want to change!

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Minimal Configuration Example\n")

	// ============================================================================
	// Example 1: Absolute Minimal Configuration
	// ============================================================================
	// This is the smallest possible configuration - everything else uses defaults!
	fmt.Println("Example 1: Absolute Minimal Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	minimalConfig := pkg.FrameworkConfig{
		// Only specify what you absolutely need
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "minimal.db",
			// Host defaults to "localhost"
			// Port defaults to 0 for sqlite (driver-specific defaults apply)
			// MaxOpenConns defaults to 25
			// MaxIdleConns defaults to 5
			// ConnMaxLifetime defaults to 5 minutes
		},
		// SessionConfig uses all defaults:
		//   EncryptionKey: random 32 bytes (WARNING: sessions won't persist across restarts)
		//   CookieName: "rockstar_session"
		//   CookiePath: "/"
		//   SessionLifetime: 24 hours
		//   CleanupInterval: 1 hour
		//   CookieHTTPOnly: true
		//   CookieSecure: true
		//   CookieSameSite: "Lax"
		//   FilesystemPath: "./sessions"
		// ServerConfig uses all defaults:
		//   ReadTimeout: 30s
		//   WriteTimeout: 30s
		//   IdleTimeout: 120s
		//   MaxHeaderBytes: 1MB
		//   MaxConnections: 10000
		//   MaxRequestSize: 10MB
		//   ShutdownTimeout: 30s
		//   ReadBufferSize: 4096
		//   WriteBufferSize: 4096
		//
		// CacheConfig uses all defaults:
		//   Type: "memory"
		//   MaxSize: 0 (unlimited)
		//   DefaultTTL: 0 (no expiration)
	}

	app1, err := pkg.New(minimalConfig)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	fmt.Println("✅ Framework created with minimal configuration!")
	fmt.Println("   All unspecified values use sensible defaults")
	fmt.Println()

	// ============================================================================
	// Example 2: Configuration with Some Overrides
	// ============================================================================
	// Specify only the values you want to change from defaults
	fmt.Println("Example 2: Configuration with Selective Overrides")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	customConfig := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1: true,
			EnableHTTP2: true,
			// Override only the values you need to change
			MaxConnections: 5000, // Default is 10000
			// All other values use defaults
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:       "sqlite",
			Database:     "custom.db",
			MaxOpenConns: 50, // Default is 25
			// Host defaults to "localhost"
			// Port defaults to 0 for sqlite
			// MaxIdleConns defaults to 5
			// ConnMaxLifetime defaults to 5 minutes
		},
		CacheConfig: pkg.CacheConfig{
			// Type defaults to "memory"
			MaxSize: 100 * 1024 * 1024, // 100MB (default is unlimited)
			// DefaultTTL defaults to 0 (no expiration)
		},
		SessionConfig: pkg.SessionConfig{
			CookieSecure: false, // Override for development (default is true)
			// EncryptionKey defaults to random 32 bytes
			// All other values use defaults
		},
	}

	app2, err := pkg.New(customConfig)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	fmt.Println("✅ Framework created with custom overrides!")
	fmt.Println("   MaxConnections: 5000 (overridden from default 10000)")
	fmt.Println("   MaxOpenConns: 50 (overridden from default 25)")
	fmt.Println("   CacheMaxSize: 100MB (overridden from default unlimited)")
	fmt.Println("   All other values use defaults")
	fmt.Println()

	// ============================================================================
	// Example 3: Monitoring Configuration with Defaults
	// ============================================================================
	fmt.Println("Example 3: Monitoring Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	monitoringConfig := pkg.MonitoringConfig{
		EnableMetrics: true,
		EnablePprof:   true,
		// MetricsPort defaults to 9090
		// PprofPort defaults to 6060
		// SNMPPort defaults to 161
		// SNMPCommunity defaults to "public"
		// OptimizationInterval defaults to 5 minutes
	}

	logger := pkg.NewLogger(nil)
	monitor := pkg.NewMonitoringManager(monitoringConfig, app1.Metrics(), app1.Database(), logger)
	if err := monitor.Start(); err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
	}
	defer monitor.Stop()

	fmt.Println("✅ Monitoring started with defaults!")
	fmt.Println("   Metrics endpoint: http://localhost:9090/metrics (default port)")
	fmt.Println("   Pprof endpoint: http://localhost:6060/debug/pprof (default port)")
	fmt.Println("   Optimization runs every 5 minutes (default interval)")
	fmt.Println()

	// ============================================================================
	// Default Values Summary
	// ============================================================================
	fmt.Println("Default Values Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("ServerConfig:")
	fmt.Println("  ReadTimeout: 30s")
	fmt.Println("  WriteTimeout: 30s")
	fmt.Println("  IdleTimeout: 120s")
	fmt.Println("  MaxHeaderBytes: 1MB")
	fmt.Println("  MaxConnections: 10000")
	fmt.Println("  MaxRequestSize: 10MB")
	fmt.Println("  ShutdownTimeout: 30s")
	fmt.Println("  ReadBufferSize: 4096")
	fmt.Println("  WriteBufferSize: 4096")
	fmt.Println()
	fmt.Println("DatabaseConfig:")
	fmt.Println("  Host: \"localhost\"")
	fmt.Println("  Port: driver-specific (postgres=5432, mysql=3306, mssql=1433, sqlite=0)")
	fmt.Println("  MaxOpenConns: 25")
	fmt.Println("  MaxIdleConns: 5")
	fmt.Println("  ConnMaxLifetime: 5 minutes")
	fmt.Println()
	fmt.Println("CacheConfig:")
	fmt.Println("  Type: \"memory\"")
	fmt.Println("  MaxSize: 0 (unlimited)")
	fmt.Println("  DefaultTTL: 0 (no expiration)")
	fmt.Println()
	fmt.Println("SessionConfig:")
	fmt.Println("  EncryptionKey: random 32 bytes (WARNING: sessions won't persist across restarts)")
	fmt.Println("  CookieName: \"rockstar_session\"")
	fmt.Println("  CookiePath: \"/\"")
	fmt.Println("  SessionLifetime: 24 hours")
	fmt.Println("  CleanupInterval: 1 hour")
	fmt.Println("  FilesystemPath: \"./sessions\"")
	fmt.Println("  CookieHTTPOnly: true")
	fmt.Println("  CookieSecure: true")
	fmt.Println("  CookieSameSite: \"Lax\"")
	fmt.Println()
	fmt.Println("MonitoringConfig:")
	fmt.Println("  MetricsPort: 9090")
	fmt.Println("  PprofPort: 6060")
	fmt.Println("  SNMPPort: 161")
	fmt.Println("  SNMPCommunity: \"public\"")
	fmt.Println("  OptimizationInterval: 5 minutes")
	fmt.Println()
	fmt.Println("PluginConfig:")
	fmt.Println("  Enabled: true")
	fmt.Println("  Priority: 0")
	fmt.Println("  Config: {} (empty map)")
	fmt.Println("  Permissions: all false (secure by default)")
	fmt.Println()
	fmt.Println("PluginsConfig:")
	fmt.Println("  Enabled: true")
	fmt.Println("  Directory: \"./plugins\"")
	fmt.Println("  Plugins: [] (empty slice)")
	fmt.Println()

	// Clean up
	_ = app1
	_ = app2

	fmt.Println("✅ All examples completed successfully!")
}
