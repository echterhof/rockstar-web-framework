package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
	// Plugin imports are automatically generated in plugin_imports.go
	// Run 'make discover-plugins' to regenerate plugin imports
)

var (
	addr          = flag.String("addr", ":8080", "Server address")
	configFile    = flag.String("config", "config.yaml", "Configuration file path")
	tlsCert       = flag.String("tls-cert", "", "TLS certificate file")
	tlsKey        = flag.String("tls-key", "", "TLS key file")
	enableQUIC    = flag.Bool("quic", false, "Enable QUIC protocol")
	enableHTTP2   = flag.Bool("http2", true, "Enable HTTP/2 protocol")
	dbDriver      = flag.String("db-driver", "sqlite", "Database driver (mysql, postgres, mssql, sqlite)")
	dbHost        = flag.String("db-host", "localhost", "Database host")
	dbPort        = flag.Int("db-port", 5432, "Database port")
	dbName        = flag.String("db-name", "rockstar.db", "Database name")
	dbUser        = flag.String("db-user", "", "Database username")
	dbPass        = flag.String("db-pass", "", "Database password")
	pluginDir     = flag.String("plugin-dir", "./plugins", "Plugin directory path")
	localesDir    = flag.String("locales-dir", "./locales", "Locales directory path")
	enableMetrics = flag.Bool("metrics", true, "Enable metrics endpoint")
	enablePprof   = flag.Bool("pprof", false, "Enable pprof debugging endpoints")
	logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	version       = flag.Bool("version", false, "Print version and exit")
	listPlugins   = flag.Bool("list-plugins", false, "List all registered compile-time plugins and exit")
)

const appVersion = "1.0.0"

func main() {
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("Rockstar Web Framework v%s\n", appVersion)
		os.Exit(0)
	}

	// Handle list-plugins flag
	if *listPlugins {
		listRegisteredPlugins()
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Validate flags
	if err := validateFlags(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create configuration
	config := createConfig()

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Load plugins if directory exists
	if err := loadPlugins(app); err != nil {
		log.Printf("Warning: Failed to load plugins: %v", err)
	}

	// Setup lifecycle hooks
	setupHooks(app)

	// Setup routes
	setupRoutes(app)

	// Setup graceful shutdown
	setupGracefulShutdown(app)

	// Start server
	log.Printf("Starting server on %s", *addr)
	if err := startServer(app); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func listRegisteredPlugins() {
	fmt.Println("Registered Compile-Time Plugins:")
	fmt.Println("================================")

	plugins := pkg.GetRegisteredPlugins()

	if len(plugins) == 0 {
		fmt.Println("No plugins registered.")
		fmt.Println("\nTo add plugins:")
		fmt.Println("  1. Place plugin packages in the plugins/ directory")
		fmt.Println("  2. Import them in cmd/rockstar/main.go")
		fmt.Println("  3. Rebuild the application with 'make build'")
		return
	}

	fmt.Printf("\nTotal plugins: %d\n\n", len(plugins))

	for i, name := range plugins {
		fmt.Printf("%d. %s\n", i+1, name)

		// Try to create plugin instance to get more details
		plugin, err := pkg.CreatePlugin(name)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		fmt.Printf("   Version:     %s\n", plugin.Version())
		fmt.Printf("   Description: %s\n", plugin.Description())
		fmt.Printf("   Author:      %s\n", plugin.Author())

		deps := plugin.Dependencies()
		if len(deps) > 0 {
			fmt.Printf("   Dependencies:\n")
			for _, dep := range deps {
				optional := ""
				if dep.Optional {
					optional = " (optional)"
				}
				fmt.Printf("     - %s %s%s\n", dep.Name, dep.Version, optional)
			}
		}

		fmt.Println()
	}
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   🎸  ROCKSTAR WEB FRAMEWORK  🎸                         ║
║                                                           ║
║   High-Performance Enterprise Go Web Framework            ║
║   Version %-48s║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Printf(banner, appVersion)
}

func validateFlags() error {
	// Validate TLS configuration
	if (*tlsCert != "" && *tlsKey == "") || (*tlsCert == "" && *tlsKey != "") {
		return fmt.Errorf("both tls-cert and tls-key must be provided together")
	}

	// Validate QUIC requires TLS
	if *enableQUIC && (*tlsCert == "" || *tlsKey == "") {
		return fmt.Errorf("QUIC protocol requires TLS certificate and key")
	}

	// Validate database port
	if *dbPort < 1 || *dbPort > 65535 {
		return fmt.Errorf("invalid database port: %d", *dbPort)
	}

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[*logLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", *logLevel)
	}

	return nil
}

func createConfig() pkg.FrameworkConfig {
	// Generate encryption key for session (32 bytes for AES-256)
	encryptionKey := make([]byte, 32)
	for i := range encryptionKey {
		encryptionKey[i] = byte(i)
	}

	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  2 << 20,
			EnableHTTP1:     true,
			EnableHTTP2:     *enableHTTP2,
			EnableQUIC:      *enableQUIC,
			EnableMetrics:   *enableMetrics,
			MetricsPath:     "/metrics",
			EnablePprof:     *enablePprof,
			PprofPath:       "/debug/pprof",
			ShutdownTimeout: 30 * time.Second,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:          *dbDriver,
			Host:            *dbHost,
			Port:            *dbPort,
			Database:        *dbName,
			Username:        *dbUser,
			Password:        *dbPass,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    100 * 1024 * 1024,
			DefaultTTL: 5 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "rockstar_session",
			CookiePath:      "/",
			CookieDomain:    "",
			CookieSecure:    *tlsCert != "",
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			SessionLifetime: 24 * time.Hour,
			EncryptionKey:   encryptionKey,
			FilesystemPath:  "./sessions",
			CleanupInterval: 1 * time.Hour,
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        *localesDir,
			SupportedLocales:  []string{"en", "de", "es", "fr"},
			FallbackToDefault: true,
		},
		SecurityConfig: pkg.SecurityConfig{
			MaxRequestSize:   10 * 1024 * 1024,
			RequestTimeout:   30 * time.Second,
			CSRFTokenExpiry:  24 * time.Hour,
			EncryptionKey:    "",
			JWTSecret:        "",
			XFrameOptions:    "DENY",
			EnableXSSProtect: true,
			EnableCSRF:       true,
			AllowedOrigins:   []string{"*"},
		},
		MonitoringConfig: pkg.MonitoringConfig{
			EnableMetrics: *enableMetrics,
			MetricsPath:   "/metrics",
			EnablePprof:   *enablePprof,
			PprofPath:     "/debug/pprof",
		},
	}

	// Load config file if specified and exists
	if *configFile != "" {
		if _, err := os.Stat(*configFile); err == nil {
			config.ConfigFiles = []string{*configFile}
			log.Printf("Loading configuration from: %s", *configFile)
		} else {
			log.Printf("Config file not found: %s (using defaults)", *configFile)
		}
	}

	return config
}

func loadPlugins(app *pkg.Framework) error {
	// With compile-time plugins, they are already registered via init()
	// This function now just reports what plugins are available

	plugins := pkg.GetRegisteredPlugins()

	if len(plugins) == 0 {
		log.Println("No compile-time plugins registered")
		log.Println("To add plugins, place them in plugins/ directory and rebuild")
		return nil
	}

	log.Printf("Registered compile-time plugins: %d", len(plugins))
	for _, name := range plugins {
		plugin, err := pkg.CreatePlugin(name)
		if err != nil {
			log.Printf("  ✗ %s (error: %v)", name, err)
			continue
		}
		log.Printf("  ✓ %s v%s", name, plugin.Version())
	}

	// The framework's plugin manager will handle initialization
	// based on configuration

	return nil
}

func setupHooks(app *pkg.Framework) {
	// Startup hooks
	app.RegisterStartupHook(func(ctx context.Context) error {
		log.Println("✓ Database connections initialized")
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		log.Println("✓ Configuration loaded")
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		// Print configuration summary
		protocols := []string{"HTTP/1.1"}
		if *enableHTTP2 {
			protocols = append(protocols, "HTTP/2")
		}
		if *enableQUIC {
			protocols = append(protocols, "QUIC")
		}
		log.Printf("✓ Protocols enabled: %v", protocols)
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		log.Println("✓ Framework ready")
		return nil
	})

	// Shutdown hooks
	app.RegisterShutdownHook(func(ctx context.Context) error {
		log.Println("✓ Graceful shutdown initiated")
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		log.Println("✓ Cleaning up resources")
		return nil
	})
}

func setupRoutes(app *pkg.Framework) {
	router := app.Router()

	// Health check endpoints
	router.GET("/health", handleHealth)
	router.GET("/ready", handleReady)
	router.GET("/", handleRoot)

	// API routes
	api := router.Group("/api/v1")
	api.GET("/status", handleAPIStatus)
	api.GET("/info", handleAPIInfo)

	// Example CRUD endpoints
	api.GET("/items", handleListItems)
	api.GET("/items/:id", handleGetItem)
	api.POST("/items", handleCreateItem)
	api.PUT("/items/:id", handleUpdateItem)
	api.DELETE("/items/:id", handleDeleteItem)
}

func handleHealth(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
	})
}

func handleReady(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status": "ready",
	})
}

func handleRoot(ctx pkg.Context) error {
	endpoints := map[string]string{
		"health": "/health",
		"ready":  "/ready",
		"api":    "/api/v1",
	}

	if *enableMetrics {
		endpoints["metrics"] = "/metrics"
	}
	if *enablePprof {
		endpoints["pprof"] = "/debug/pprof"
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":   "Welcome to Rockstar Web Framework! 🎸",
		"version":   appVersion,
		"endpoints": endpoints,
	})
}

func handleAPIStatus(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"api_version": "v1",
		"status":      "operational",
		"timestamp":   time.Now().Unix(),
	})
}

func handleAPIInfo(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"framework": "Rockstar Web Framework",
		"version":   appVersion,
		"protocols": getEnabledProtocols(),
		"database":  *dbDriver,
	})
}

func handleListItems(ctx pkg.Context) error {
	// Example: Return mock items
	items := []map[string]interface{}{
		{"id": 1, "name": "Item 1", "description": "First item"},
		{"id": 2, "name": "Item 2", "description": "Second item"},
	}
	return ctx.JSON(200, map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func handleGetItem(ctx pkg.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(200, map[string]interface{}{
		"id":          id,
		"name":        "Sample Item",
		"description": "This is a sample item",
	})
}

func handleCreateItem(ctx pkg.Context) error {
	return ctx.JSON(201, map[string]interface{}{
		"message": "Item created successfully",
		"id":      "new-item-id",
	})
}

func handleUpdateItem(ctx pkg.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(200, map[string]interface{}{
		"message": "Item updated successfully",
		"id":      id,
	})
}

func handleDeleteItem(ctx pkg.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(200, map[string]interface{}{
		"message": "Item deleted successfully",
		"id":      id,
	})
}

func getEnabledProtocols() []string {
	protocols := []string{"HTTP/1.1"}
	if *enableHTTP2 {
		protocols = append(protocols, "HTTP/2")
	}
	if *enableQUIC {
		protocols = append(protocols, "QUIC")
	}
	return protocols
}

var startTime = time.Now()

func setupGracefulShutdown(app *pkg.Framework) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutdown signal received")

		if err := app.Shutdown(30 * time.Second); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		log.Println("👋 Server stopped")
		os.Exit(0)
	}()
}

func startServer(app *pkg.Framework) error {
	if *tlsCert != "" && *tlsKey != "" {
		log.Printf("Starting with TLS (cert: %s, key: %s)", *tlsCert, *tlsKey)
		if *enableQUIC {
			return app.ListenQUIC(*addr, *tlsCert, *tlsKey)
		}
		return app.ListenTLS(*addr, *tlsCert, *tlsKey)
	}

	return app.Listen(*addr)
}
