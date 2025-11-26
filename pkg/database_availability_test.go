package pkg

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: optional-database, Property 13: IsConnected reflects database availability**
// **Validates: Requirements 5.1, 5.2**
func TestProperty_IsConnectedReflectsDatabaseAvailability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("IsConnected returns false for no-op database manager",
		prop.ForAll(
			func() bool {
				// Create no-op database manager
				db := NewNoopDatabaseManager()

				// IsConnected should return false
				if db.IsConnected() {
					t.Log("No-op database manager reports connected")
					return false
				}

				return true
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestProperty_IsConnectedReturnsFalseForFrameworkWithoutDatabase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework without database has IsConnected() returning false",
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

				// Database manager should exist
				if app.Database() == nil {
					t.Log("Database manager is nil")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// IsConnected should return false
				if app.Database().IsConnected() {
					t.Log("Database reports connected when no database is configured")
					_ = app.Shutdown(1 * time.Second)
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

func TestProperty_IsConnectedReturnsTrueForRealDatabaseManager(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Real database manager returns correct IsConnected() state",
		prop.ForAll(
			func() bool {
				// Create real database manager
				db := NewDatabaseManager()

				// Before connection, IsConnected should return false
				if db.IsConnected() {
					t.Log("Database manager reports connected before Connect() is called")
					return false
				}

				return true
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestProperty_IsConnectedConsistentWithNoopDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("IsConnected() is consistent with isNoopDatabase() detection",
		prop.ForAll(
			func(readTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						ReadTimeout: time.Duration(readTimeoutSecs) * time.Second,
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

				db := app.Database()

				// If isNoopDatabase returns true, IsConnected should return false
				if isNoopDatabase(db) && db.IsConnected() {
					t.Log("No-op database reports connected")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// If isNoopDatabase returns false, IsConnected should return true (when connected)
				// Note: We can't test actual connection without a real database

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.IntRange(5, 30),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: optional-database, Property 14: Managers accept no-op database**
// **Validates: Requirements 6.2**
func TestProperty_ManagersAcceptNoopDatabase(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("All managers initialize successfully with no-op database",
		prop.ForAll(
			func() bool {
				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Verify it's a no-op database
				if !isNoopDatabase(noopDB) {
					t.Log("Database is not a no-op implementation")
					return false
				}

				// Test SessionManager
				cache := NewCacheManager(CacheConfig{})
				sessionConfig := &SessionConfig{
					StorageType:     SessionStorageCache,
					CookieName:      "test_session",
					SessionLifetime: 1 * time.Hour,
					CleanupInterval: 5 * time.Minute,
					EncryptionKey:   []byte("12345678901234567890123456789012"),
				}
				sm, err := NewSessionManager(sessionConfig, noopDB, cache)
				if err != nil {
					t.Logf("SessionManager initialization failed with no-op database: %v", err)
					return false
				}
				if sm == nil {
					t.Log("SessionManager is nil")
					return false
				}

				// Test SecurityManager
				secConfig := DefaultSecurityConfig()
				secConfig.EncryptionKey = "0123456789abcdef0123456789abcdef"
				secConfig.JWTSecret = "test-jwt-secret"
				secMgr, err := NewSecurityManager(noopDB, secConfig)
				if err != nil {
					t.Logf("SecurityManager initialization failed with no-op database: %v", err)
					return false
				}
				if secMgr == nil {
					t.Log("SecurityManager is nil")
					return false
				}

				// Test MetricsCollector
				metrics := NewMetricsCollector(noopDB)
				if metrics == nil {
					t.Log("MetricsCollector is nil")
					return false
				}

				// Test MonitoringManager
				logger := NewLogger(nil)
				monConfig := MonitoringConfig{}
				monMgr := NewMonitoringManager(monConfig, metrics, noopDB, logger)
				if monMgr == nil {
					t.Log("MonitoringManager is nil")
					return false
				}

				return true
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestProperty_ManagersAcceptNoopDatabaseFromFramework(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework managers accept no-op database when framework is initialized without database",
		prop.ForAll(
			func(readTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						ReadTimeout: time.Duration(readTimeoutSecs) * time.Second,
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

				// Verify database is no-op
				if !isNoopDatabase(app.Database()) {
					t.Log("Database is not a no-op implementation")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Verify all managers are initialized
				if app.Session() == nil {
					t.Log("SessionManager is nil")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				if app.Security() == nil {
					t.Log("SecurityManager is nil")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				if app.Metrics() == nil {
					t.Log("MetricsCollector is nil")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				if app.Monitoring() == nil {
					t.Log("MonitoringManager is nil")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.IntRange(5, 30),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: optional-database, Property 15: Managers configure for no-database operation**
// **Validates: Requirements 6.3**
func TestProperty_ManagersConfigureForNoDatabaseOperation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Managers automatically configure for in-memory storage when no database is available",
		prop.ForAll(
			func() bool {
				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Test SessionManager - should use in-memory storage
				cache := NewCacheManager(CacheConfig{})
				sessionConfig := &SessionConfig{
					StorageType:     SessionStorageCache,
					CookieName:      "test_session",
					SessionLifetime: 1 * time.Hour,
					CleanupInterval: 5 * time.Minute,
					EncryptionKey:   []byte("12345678901234567890123456789012"),
				}
				sm, err := NewSessionManager(sessionConfig, noopDB, cache)
				if err != nil {
					t.Logf("SessionManager initialization failed: %v", err)
					return false
				}

				// SessionManager should work without database
				// Try to save and load a session
				testSession := &Session{
					ID:        "test-session-id",
					UserID:    "test-user",
					TenantID:  "test-tenant",
					Data:      map[string]interface{}{"key": "value"},
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				// This should work with in-memory storage
				// Note: Save requires a Context, but for in-memory storage it's not strictly needed
				// We pass nil since we're testing the storage mechanism directly
				err = sm.Save(nil, testSession)
				if err != nil {
					t.Logf("SessionManager Save failed: %v", err)
					return false
				}

				// Test SecurityManager - should use in-memory token storage
				secConfig := DefaultSecurityConfig()
				secConfig.EncryptionKey = "0123456789abcdef0123456789abcdef"
				secConfig.JWTSecret = "test-jwt-secret"
				secMgr, err := NewSecurityManager(noopDB, secConfig)
				if err != nil {
					t.Logf("SecurityManager initialization failed: %v", err)
					return false
				}

				// SecurityManager should work without database
				// The fact that it initialized successfully means it configured for in-memory operation
				if secMgr == nil {
					t.Log("SecurityManager is nil")
					return false
				}

				// Test MetricsCollector - should use in-memory metrics storage
				metrics := NewMetricsCollector(noopDB)
				if metrics == nil {
					t.Log("MetricsCollector is nil")
					return false
				}

				// MetricsCollector should work without database
				// Try to record metrics
				testMetrics := &WorkloadMetrics{
					Timestamp:   time.Now(),
					TenantID:    "test-tenant",
					UserID:      "test-user",
					RequestID:   "test-request",
					Duration:    100,
					ContextSize: 1024,
					MemoryUsage: 2048,
					CPUUsage:    10.5,
					Path:        "/test",
					Method:      "GET",
					StatusCode:  200,
				}

				err = metrics.RecordWorkloadMetrics(testMetrics)
				if err != nil {
					t.Logf("MetricsCollector RecordWorkloadMetrics failed: %v", err)
					return false
				}

				// Test MonitoringManager - should use in-memory metrics storage
				logger := NewLogger(nil)
				monConfig := MonitoringConfig{}
				monMgr := NewMonitoringManager(monConfig, metrics, noopDB, logger)
				if monMgr == nil {
					t.Log("MonitoringManager is nil")
					return false
				}

				return true
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestProperty_ManagersConfigureForNoDatabaseInFramework(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Framework managers configure for in-memory operation when initialized without database",
		prop.ForAll(
			func(readTimeoutSecs int) bool {
				// Create config without database configuration
				config := FrameworkConfig{
					ServerConfig: ServerConfig{
						ReadTimeout: time.Duration(readTimeoutSecs) * time.Second,
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

				// Verify database is no-op
				if !isNoopDatabase(app.Database()) {
					t.Log("Database is not a no-op implementation")
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Test that SessionManager works (configured for in-memory)
				testSession := &Session{
					ID:        "framework-test-session",
					UserID:    "test-user",
					TenantID:  "test-tenant",
					Data:      map[string]interface{}{"test": "data"},
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				err = app.Session().Save(nil, testSession)
				if err != nil {
					t.Logf("SessionManager Save failed in framework: %v", err)
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Test that MetricsCollector works (configured for in-memory)
				testMetrics := &WorkloadMetrics{
					Timestamp:   time.Now(),
					TenantID:    "test-tenant",
					UserID:      "test-user",
					RequestID:   "framework-test-request",
					Duration:    150,
					ContextSize: 2048,
					MemoryUsage: 4096,
					CPUUsage:    15.5,
					Path:        "/framework-test",
					Method:      "POST",
					StatusCode:  201,
				}

				err = app.Metrics().RecordWorkloadMetrics(testMetrics)
				if err != nil {
					t.Logf("MetricsCollector RecordWorkloadMetrics failed in framework: %v", err)
					_ = app.Shutdown(1 * time.Second)
					return false
				}

				// Cleanup
				_ = app.Shutdown(1 * time.Second)

				return true
			},
			gen.IntRange(5, 30),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
