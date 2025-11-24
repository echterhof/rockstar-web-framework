package pkg

import (
	"context"
	"testing"
	"time"
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
