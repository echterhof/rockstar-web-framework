package pkg

import (
	"context"
	"os"
	"testing"
	"time"

	// Import SQLite driver for tests
	_ "github.com/mattn/go-sqlite3"
)

func TestFrameworkCreation(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		CacheConfig: CacheConfig{
			Type:       "memory",
			MaxSize:    10 * 1024 * 1024,
			DefaultTTL: 5 * time.Minute,
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
		SecurityConfig: SecurityConfig{
			XFrameOptions:  "DENY",
			MaxRequestSize: 1024 * 1024,
			RequestTimeout: 30 * time.Second,
		},
		MonitoringConfig: MonitoringConfig{
			EnableMetrics: true,
			MetricsPath:   "/metrics",
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	if app == nil {
		t.Fatal("Framework instance is nil")
	}

	// Test component access
	if app.Router() == nil {
		t.Error("Router is nil")
	}

	if app.Database() == nil {
		t.Error("Database is nil")
	}

	if app.Cache() == nil {
		t.Error("Cache is nil")
	}

	if app.Session() == nil {
		t.Error("Session is nil")
	}

	if app.Security() == nil {
		t.Error("Security is nil")
	}

	if app.Config() == nil {
		t.Error("Config is nil")
	}

	if app.I18n() == nil {
		t.Error("I18n is nil")
	}

	if app.Metrics() == nil {
		t.Error("Metrics is nil")
	}

	if app.Monitoring() == nil {
		t.Error("Monitoring is nil")
	}

	if app.Proxy() == nil {
		t.Error("Proxy is nil")
	}
}

func TestFrameworkMiddleware(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Add middleware
	_ = false // middlewareCalled placeholder
	app.Use(func(ctx Context, next HandlerFunc) error {
		// middlewareCalled = true
		return next(ctx)
	})

	// Verify middleware was added
	if len(app.globalMiddleware) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(app.globalMiddleware))
	}
}

func TestFrameworkLifecycleHooks(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Test startup hooks
	_ = false // startupCalled placeholder
	app.RegisterStartupHook(func(ctx context.Context) error {
		// startupCalled = true
		return nil
	})

	if len(app.startupHooks) != 1 {
		t.Errorf("Expected 1 startup hook, got %d", len(app.startupHooks))
	}

	// Test shutdown hooks
	_ = false // shutdownCalled placeholder
	app.RegisterShutdownHook(func(ctx context.Context) error {
		// shutdownCalled = true
		return nil
	})

	if len(app.shutdownHooks) != 1 {
		t.Errorf("Expected 1 shutdown hook, got %d", len(app.shutdownHooks))
	}
}

func TestFrameworkErrorHandler(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Set custom error handler
	_ = false // errorHandlerCalled placeholder
	app.SetErrorHandler(func(ctx Context, err error) error {
		// errorHandlerCalled = true
		return nil
	})

	if app.errorHandler == nil {
		t.Error("Error handler was not set")
	}
}

func TestFrameworkRouting(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Define routes
	router := app.Router()

	_ = false // handlerCalled placeholder
	router.GET("/test", func(ctx Context) error {
		// handlerCalled = true
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	// Verify route was registered
	routes := router.Routes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].Method != "GET" {
		t.Errorf("Expected GET method, got %s", routes[0].Method)
	}

	if routes[0].Path != "/test" {
		t.Errorf("Expected /test path, got %s", routes[0].Path)
	}
}

func TestFrameworkShutdown(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Register shutdown hook
	shutdownCalled := false
	app.RegisterShutdownHook(func(ctx context.Context) error {
		shutdownCalled = true
		return nil
	})

	// Test shutdown
	err = app.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !shutdownCalled {
		t.Error("Shutdown hook was not called")
	}

	if app.IsRunning() {
		t.Error("Framework should not be running after shutdown")
	}
}

func TestFrameworkIsRunning(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}

	// Initially not running
	if app.IsRunning() {
		t.Error("Framework should not be running initially")
	}
}

func BenchmarkFrameworkCreation(b *testing.B) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(config)
		if err != nil {
			b.Fatalf("Failed to create framework: %v", err)
		}
	}
}

func BenchmarkFrameworkRouteRegistration(b *testing.B) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	app, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}

	router := app.Router()
	handler := func(ctx Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.GET("/test", handler)
	}
}

func TestFrameworkWithConfigFiles(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/test_config.json"

	configContent := `{
		"app": {
			"name": "test-app",
			"version": "1.0.0"
		},
		"server": {
			"port": 8080,
			"host": "localhost"
		},
		"database": {
			"driver": "postgres",
			"host": "localhost"
		}
	}`

	err := writeFile(configPath, []byte(configContent))
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes
		},
		ConfigFiles: []string{configPath},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework with config files: %v", err)
	}

	if app == nil {
		t.Fatal("Framework instance is nil")
	}

	if app.Config() == nil {
		t.Fatal("Config manager is nil")
	}

	// Verify configuration was loaded
	appName := app.Config().GetString("app.name")
	if appName != "test-app" {
		t.Errorf("Expected app.name to be 'test-app', got '%s'", appName)
	}

	serverPort := app.Config().GetInt("server.port")
	if serverPort != 8080 {
		t.Errorf("Expected server.port to be 8080, got %d", serverPort)
	}

	dbDriver := app.Config().GetString("database.driver")
	if dbDriver != "postgres" {
		t.Errorf("Expected database.driver to be 'postgres', got '%s'", dbDriver)
	}
}

func TestFrameworkWithoutConfigFiles(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes
		},
		ConfigFiles: []string{}, // No config files
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework without config files: %v", err)
	}

	if app == nil {
		t.Fatal("Framework instance is nil")
	}

	if app.Config() == nil {
		t.Fatal("Config manager should not be nil even without config files")
	}

	// Config should be empty but functional
	nonExistentValue := app.Config().GetString("non.existent.key")
	if nonExistentValue != "" {
		t.Errorf("Expected empty string for non-existent key, got '%s'", nonExistentValue)
	}
}

func TestFrameworkWithInvalidConfigFile(t *testing.T) {
	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		ConfigFiles: []string{"/non/existent/config.json"},
	}

	app, err := New(config)
	if err == nil {
		t.Fatal("Expected error when loading non-existent config file")
	}

	if app != nil {
		t.Error("Framework instance should be nil when config loading fails")
	}
}

func TestFrameworkWithMultipleConfigFiles(t *testing.T) {
	// Create temporary config files
	tmpDir := t.TempDir()

	config1Path := tmpDir + "/config1.json"
	config1Content := `{
		"app": {
			"name": "test-app"
		},
		"feature1": {
			"enabled": true
		}
	}`

	config2Path := tmpDir + "/config2.yaml"
	config2Content := `
app:
  version: "2.0.0"
feature2:
  enabled: true
`

	err := writeFile(config1Path, []byte(config1Content))
	if err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}

	err = writeFile(config2Path, []byte(config2Content))
	if err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SessionConfig: SessionConfig{
			StorageType:     SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes
		},
		ConfigFiles: []string{config1Path, config2Path},
	}

	app, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create framework with multiple config files: %v", err)
	}

	if app == nil {
		t.Fatal("Framework instance is nil")
	}

	// Verify second config was loaded (it replaces the first)
	appVersion := app.Config().GetString("app.version")
	if appVersion != "2.0.0" {
		t.Errorf("Expected app.version to be '2.0.0', got '%s'", appVersion)
	}

	// feature1 should NOT be present since config2 replaced config1
	feature1Enabled := app.Config().GetBool("feature1.enabled")
	if feature1Enabled {
		t.Error("Expected feature1.enabled to be false (config was replaced)")
	}

	// feature2 should be present from config2
	feature2Enabled := app.Config().GetBool("feature2.enabled")
	if !feature2Enabled {
		t.Error("Expected feature2.enabled to be true")
	}
}

func TestFrameworkWithMalformedConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/malformed.json"

	// Invalid JSON
	configContent := `{
		"app": {
			"name": "test-app"
		// missing closing brace
	`

	err := writeFile(configPath, []byte(configContent))
	if err != nil {
		t.Fatalf("Failed to create malformed config file: %v", err)
	}

	config := FrameworkConfig{
		ServerConfig: ServerConfig{
			EnableHTTP1: true,
		},
		DatabaseConfig: DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		ConfigFiles: []string{configPath},
	}

	app, err := New(config)
	if err == nil {
		t.Fatal("Expected error when loading malformed config file")
	}

	if app != nil {
		t.Error("Framework instance should be nil when config parsing fails")
	}
}

// Helper function to write files
func writeFile(path string, data []byte) error {
	// Use os package to write file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
