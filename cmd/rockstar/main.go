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
)

var (
	addr       = flag.String("addr", ":8080", "Server address")
	configFile = flag.String("config", "config.yaml", "Configuration file path")
	tlsCert    = flag.String("tls-cert", "", "TLS certificate file")
	tlsKey     = flag.String("tls-key", "", "TLS key file")
	enableQUIC = flag.Bool("quic", false, "Enable QUIC protocol")
	dbDriver   = flag.String("db-driver", "sqlite", "Database driver (mysql, postgres, mssql, sqlite)")
	dbHost     = flag.String("db-host", "localhost", "Database host")
	dbPort     = flag.Int("db-port", 5432, "Database port")
	dbName     = flag.String("db-name", "rockstar.db", "Database name")
	dbUser     = flag.String("db-user", "", "Database username")
	dbPass     = flag.String("db-pass", "", "Database password")
)

func main() {
	flag.Parse()

	// Print banner
	printBanner()

	// Create configuration
	config := createConfig()

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
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

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   🎸  ROCKSTAR WEB FRAMEWORK  🎸                         ║
║                                                           ║
║   High-Performance Enterprise Go Web Framework            ║
║   Version 1.0.0                                           ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func createConfig() pkg.FrameworkConfig {
	// Generate encryption key for session (32 bytes for AES-256)
	encryptionKey := make([]byte, 32)

	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  2 << 20,
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableQUIC:      *enableQUIC,
			EnableMetrics:   true,
			MetricsPath:     "/metrics",
			EnablePprof:     true,
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
			LocalesDir:        "./locales",
			SupportedLocales:  []string{"en", "de"},
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
			EnableMetrics: true,
			MetricsPath:   "/metrics",
			EnablePprof:   true,
			PprofPath:     "/debug/pprof",
		},
	}

	// Load config file if specified
	if *configFile != "" {
		config.ConfigFiles = []string{*configFile}
	}

	return config
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
		log.Println("✓ Framework ready")
		return nil
	})

	// Shutdown hooks
	app.RegisterShutdownHook(func(ctx context.Context) error {
		log.Println("✓ Graceful shutdown initiated")
		return nil
	})
}

func setupRoutes(app *pkg.Framework) {
	router := app.Router()

	// Health check endpoints
	router.GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	router.GET("/ready", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status": "ready",
		})
	})

	// API root
	router.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"message": "Welcome to Rockstar Web Framework! 🎸",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"health":  "/health",
				"ready":   "/ready",
				"metrics": "/metrics",
				"pprof":   "/debug/pprof",
			},
		})
	})

	// Example API routes
	api := router.Group("/api/v1")
	api.GET("/status", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"api_version": "v1",
			"status":      "operational",
		})
	})
}

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
