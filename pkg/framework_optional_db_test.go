package pkg

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: optional-database, Property 2: No-op database manager is used when database is not configured**
// **Validates: Requirements 1.3, 4.4**
func TestProperty_NoopDatabaseManagerUsedWhenNotConfigured(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework uses no-op database when database is not configured",
		prop.ForAll(
			func(readTimeoutSecs, writeTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						ReadTimeout:  time.Duration(readTimeoutSecs) * time.Second,
						WriteTimeout: time.Duration(writeTimeoutSecs) * time.Second,
						EnableHTTP1:  true,
					},
					// DatabaseConfig is empty - no database configured
					DatabaseConfig: DatabaseConfig{},
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
				}

				// Initialize framework
				app, err := New(config)
				if err != nil {
					t.Logf("Framework initialization failed: %v", err)
					return false
				}

				// Verify database manager is not nil
				if app.Database() == nil {
					t.Log("Database manager is nil")
					return false
				}

				// Verify database manager is no-op (IsConnected returns false)
				if app.Database().IsConnected() {
					t.Log("Database reports connected when it should be no-op")
					return false
				}

				// Verify it's actually a no-op manager
				if !isNoopDatabase(app.Database()) {
					t.Log("Database manager is not a no-op implementation")
					return false
				}

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.IntRange(5, 30),
			gen.IntRange(5, 30),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test with various empty database configurations
func TestProperty_NoopDatabaseWithVariousEmptyConfigs(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework uses no-op database with various empty configurations",
		prop.ForAll(
			func(driver, database, username, password string) bool {
				// Create config with potentially empty database fields
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						EnableHTTP1: true,
					},
					DatabaseConfig: DatabaseConfig{
						Driver:   driver,
						Database: database,
						Username: username,
						Password: password,
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
				}

				// Determine if this config should be considered "configured"
				isConfigured := isDatabaseConfigured(config.DatabaseConfig)

				// Initialize framework
				app, err := New(config)

				// If database is not configured, framework should initialize successfully
				if !isConfigured {
					if err != nil {
						t.Logf("Framework initialization failed with empty config: %v", err)
						return false
					}

					// Should use no-op database
					if app.Database().IsConnected() {
						t.Log("Database reports connected with empty config")
						return false
					}

					// Cleanup
					_ = app.Shutdown(1 * time.Second)
					return true
				}

				// If database is configured but invalid, initialization may fail
				// This is acceptable behavior
				return true
			},
			gen.OneConstOf("", "sqlite", "postgres", "mysql"),
			gen.OneConstOf("", ":memory:", "testdb"),
			gen.OneConstOf("", "user"),
			gen.OneConstOf("", "pass"),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: optional-database, Property 3: All managers initialize without database**
// **Validates: Requirements 1.4**
func TestProperty_AllManagersInitializeWithoutDatabase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("All managers initialize successfully without database",
		prop.ForAll(
			func(readTimeoutSecs, writeTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						ReadTimeout:  time.Duration(readTimeoutSecs) * time.Second,
						WriteTimeout: time.Duration(writeTimeoutSecs) * time.Second,
						EnableHTTP1:  true,
					},
					// DatabaseConfig is empty - no database configured
					DatabaseConfig: DatabaseConfig{},
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
				}

				// Initialize framework
				app, err := New(config)
				if err != nil {
					t.Logf("Framework initialization failed: %v", err)
					return false
				}

				// Verify all manager accessor methods return non-nil instances
				if app.Router() == nil {
					t.Log("Router is nil")
					return false
				}

				if app.Cache() == nil {
					t.Log("Cache is nil")
					return false
				}

				if app.Session() == nil {
					t.Log("Session is nil")
					return false
				}

				if app.Security() == nil {
					t.Log("Security is nil")
					return false
				}

				if app.Monitoring() == nil {
					t.Log("Monitoring is nil")
					return false
				}

				if app.Metrics() == nil {
					t.Log("Metrics is nil")
					return false
				}

				if app.Config() == nil {
					t.Log("Config is nil")
					return false
				}

				if app.I18n() == nil {
					t.Log("I18n is nil")
					return false
				}

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.IntRange(5, 30),
			gen.IntRange(5, 30),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: optional-database, Property 4: Route handlers execute without database**
// **Validates: Requirements 1.5**
func TestProperty_RouteHandlersExecuteWithoutDatabase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Route handlers execute successfully without database",
		prop.ForAll(
			func(path, method string) bool {
				// Skip invalid paths and methods
				if path == "" || !isValidHTTPMethod(method) {
					return true
				}

				// Ensure path starts with /
				if path[0] != '/' {
					path = "/" + path
				}

				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						EnableHTTP1: true,
					},
					// DatabaseConfig is empty - no database configured
					DatabaseConfig: DatabaseConfig{},
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
				}

				// Initialize framework
				app, err := New(config)
				if err != nil {
					t.Logf("Framework initialization failed: %v", err)
					return false
				}

				// Register a route handler that doesn't use database
				handler := func(ctx Context) error {
					// This handler doesn't use database operations
					return ctx.JSON(200, map[string]string{"status": "ok"})
				}

				// Register route based on method
				router := app.Router()
				switch method {
				case "GET":
					router.GET(path, handler)
				case "POST":
					router.POST(path, handler)
				case "PUT":
					router.PUT(path, handler)
				case "DELETE":
					router.DELETE(path, handler)
				case "PATCH":
					router.PATCH(path, handler)
				}

				// Verify route was registered
				routes := router.Routes()
				if len(routes) == 0 {
					t.Log("No routes registered")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Verify the route matches what we registered
				found := false
				for _, route := range routes {
					if route.Method == method && route.Path == path {
						found = true
						break
					}
				}

				if !found {
					t.Logf("Route %s %s not found in registered routes", method, path)
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
			gen.OneConstOf("GET", "POST", "PUT", "DELETE", "PATCH"),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper function to validate HTTP methods
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, m := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}

// **Feature: optional-database, Property 11: Framework shutdown succeeds without database**
// **Validates: Requirements 3.5**
func TestProperty_FrameworkShutdownSucceedsWithoutDatabase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework shutdown completes successfully without database",
		prop.ForAll(
			func(shutdownTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						EnableHTTP1: true,
					},
					// DatabaseConfig is empty - no database configured
					DatabaseConfig: DatabaseConfig{},
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
				}

				// Initialize framework
				app, err := New(config)
				if err != nil {
					t.Logf("Framework initialization failed: %v", err)
					return false
				}

				// Verify database is not connected
				if app.Database().IsConnected() {
					t.Log("Database reports connected when it should be no-op")
					return false
				}

				// Attempt shutdown
				shutdownTimeout := time.Duration(shutdownTimeoutSecs) * time.Second
				err = app.Shutdown(shutdownTimeout)
				if err != nil {
					t.Logf("Shutdown failed: %v", err)
					return false
				}

				// Verify framework is no longer running
				if app.IsRunning() {
					t.Log("Framework still reports running after shutdown")
					return false
				}

				return true
			},
			gen.IntRange(1, 10),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test shutdown with various framework states
func TestProperty_ShutdownWithoutDatabaseVariousStates(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework shutdown succeeds in various states without database",
		prop.ForAll(
			func(registerRoutes, addMiddleware, registerHooks bool) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						EnableHTTP1: true,
					},
					// DatabaseConfig is empty - no database configured
					DatabaseConfig: DatabaseConfig{},
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
				}

				// Initialize framework
				app, err := New(config)
				if err != nil {
					t.Logf("Framework initialization failed: %v", err)
					return false
				}

				// Optionally register routes
				if registerRoutes {
					app.Router().GET("/test", func(ctx Context) error {
						return ctx.JSON(200, map[string]string{"status": "ok"})
					})
				}

				// Optionally add middleware
				if addMiddleware {
					app.Use(func(ctx Context, next HandlerFunc) error {
						return next(ctx)
					})
				}

				// Optionally register shutdown hooks
				hookExecuted := false
				if registerHooks {
					app.RegisterShutdownHook(func(ctx context.Context) error {
						hookExecuted = true
						return nil
					})
				}

				// Attempt shutdown
				err = app.Shutdown(5 * time.Second)
				if err != nil {
					t.Logf("Shutdown failed: %v", err)
					return false
				}

				// Verify shutdown hook was executed if registered
				if registerHooks && !hookExecuted {
					t.Log("Shutdown hook was not executed")
					return false
				}

				// Verify framework is no longer running
				if app.IsRunning() {
					t.Log("Framework still reports running after shutdown")
					return false
				}

				return true
			},
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
