package pkg

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_NoopDatabaseOperationsReturnDescriptiveErrors tests Property 5
// **Feature: optional-database, Property 5: No-op database operations return descriptive errors**
// **Validates: Requirements 2.1**
func TestProperty_NoopDatabaseOperationsReturnDescriptiveErrors(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("All no-op database operations return ErrNoDatabaseConfigured",
		prop.ForAll(
			func(query string, sessionID string, tokenValue string, tenantID string, hostname string, key string, limit int) bool {
				db := NewNoopDatabaseManager()

				// Test query operations
				_, err1 := db.Query(query)
				if err1 != ErrNoDatabaseConfigured {
					return false
				}

				_, err2 := db.Exec(query)
				if err2 != ErrNoDatabaseConfigured {
					return false
				}

				_, err3 := db.Prepare(query)
				if err3 != ErrNoDatabaseConfigured {
					return false
				}

				// Test transaction operations
				_, err4 := db.Begin()
				if err4 != ErrNoDatabaseConfigured {
					return false
				}

				_, err5 := db.BeginTx(nil)
				if err5 != ErrNoDatabaseConfigured {
					return false
				}

				// Test connection operations
				err6 := db.Connect(DatabaseConfig{})
				if err6 != ErrNoDatabaseConfigured {
					return false
				}

				err7 := db.Close()
				if err7 != ErrNoDatabaseConfigured {
					return false
				}

				err8 := db.Ping()
				if err8 != ErrNoDatabaseConfigured {
					return false
				}

				// Test session operations
				err9 := db.SaveSession(&Session{ID: sessionID})
				if err9 != ErrNoDatabaseConfigured {
					return false
				}

				_, err10 := db.LoadSession(sessionID)
				if err10 != ErrNoDatabaseConfigured {
					return false
				}

				err11 := db.DeleteSession(sessionID)
				if err11 != ErrNoDatabaseConfigured {
					return false
				}

				err12 := db.CleanupExpiredSessions()
				if err12 != ErrNoDatabaseConfigured {
					return false
				}

				// Test token operations
				err13 := db.SaveAccessToken(&AccessToken{Token: tokenValue})
				if err13 != ErrNoDatabaseConfigured {
					return false
				}

				_, err14 := db.LoadAccessToken(tokenValue)
				if err14 != ErrNoDatabaseConfigured {
					return false
				}

				_, err15 := db.ValidateAccessToken(tokenValue)
				if err15 != ErrNoDatabaseConfigured {
					return false
				}

				err16 := db.DeleteAccessToken(tokenValue)
				if err16 != ErrNoDatabaseConfigured {
					return false
				}

				err17 := db.CleanupExpiredTokens()
				if err17 != ErrNoDatabaseConfigured {
					return false
				}

				// Test tenant operations
				err18 := db.SaveTenant(&Tenant{ID: tenantID})
				if err18 != ErrNoDatabaseConfigured {
					return false
				}

				_, err19 := db.LoadTenant(tenantID)
				if err19 != ErrNoDatabaseConfigured {
					return false
				}

				_, err20 := db.LoadTenantByHost(hostname)
				if err20 != ErrNoDatabaseConfigured {
					return false
				}

				// Test metrics operations
				err21 := db.SaveWorkloadMetrics(&WorkloadMetrics{TenantID: tenantID})
				if err21 != ErrNoDatabaseConfigured {
					return false
				}

				_, err22 := db.GetWorkloadMetrics(tenantID, time.Now(), time.Now())
				if err22 != ErrNoDatabaseConfigured {
					return false
				}

				// Test rate limiting operations
				_, err23 := db.CheckRateLimit(key, limit, time.Minute)
				if err23 != ErrNoDatabaseConfigured {
					return false
				}

				err24 := db.IncrementRateLimit(key, time.Minute)
				if err24 != ErrNoDatabaseConfigured {
					return false
				}

				// Test migration operations
				err25 := db.Migrate()
				if err25 != ErrNoDatabaseConfigured {
					return false
				}

				err26 := db.CreateTables()
				if err26 != ErrNoDatabaseConfigured {
					return false
				}

				err27 := db.DropTables()
				if err27 != ErrNoDatabaseConfigured {
					return false
				}

				err28 := db.InitializePluginTables()
				if err28 != ErrNoDatabaseConfigured {
					return false
				}

				// Test SQL loader operations
				_, err29 := db.GetQuery("test_query")
				if err29 != ErrNoDatabaseConfigured {
					return false
				}

				return true
			},
			gen.AnyString(),
			gen.AnyString(),
			gen.AnyString(),
			gen.AnyString(),
			gen.AnyString(),
			gen.AnyString(),
			gen.IntRange(1, 1000),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_DatabaseErrorMessagesIncludeConfigurationGuidance tests Property 6
// **Feature: optional-database, Property 6: Database error messages include configuration guidance**
// **Validates: Requirements 2.5**
func TestProperty_DatabaseErrorMessagesIncludeConfigurationGuidance(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Error messages contain configuration guidance",
		prop.ForAll(
			func() bool {
				// Check that ErrNoDatabaseConfigured contains helpful guidance
				errMsg := ErrNoDatabaseConfigured.Error()

				// Should mention database configuration
				containsDatabase := dbTestStringContains(errMsg, "database") || dbTestStringContains(errMsg, "Database")
				if !containsDatabase {
					return false
				}

				// Should mention configuration or setup
				containsConfig := dbTestStringContains(errMsg, "config") || dbTestStringContains(errMsg, "Config") ||
					dbTestStringContains(errMsg, "configure") || dbTestStringContains(errMsg, "setup")
				if !containsConfig {
					return false
				}

				// Should provide actionable guidance (credentials, driver, etc.)
				containsGuidance := dbTestStringContains(errMsg, "Driver") || dbTestStringContains(errMsg, "credentials") ||
					dbTestStringContains(errMsg, "documentation") || dbTestStringContains(errMsg, "docs")
				if !containsGuidance {
					return false
				}

				return true
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestNoopDatabaseManager_IsConnected tests that IsConnected returns false for no-op manager
func TestNoopDatabaseManager_IsConnected(t *testing.T) {
	db := NewNoopDatabaseManager()
	if db.IsConnected() {
		t.Error("Expected IsConnected() to return false for no-op database manager")
	}
}

// TestNoopDatabaseManager_Stats tests that Stats returns empty stats
func TestNoopDatabaseManager_Stats(t *testing.T) {
	db := NewNoopDatabaseManager()
	stats := db.Stats()

	if stats.OpenConnections != 0 || stats.InUse != 0 || stats.Idle != 0 {
		t.Error("Expected Stats() to return empty stats for no-op database manager")
	}
}

// TestNoopDatabaseManager_QueryRow tests that QueryRow returns an empty row
func TestNoopDatabaseManager_QueryRow(t *testing.T) {
	db := NewNoopDatabaseManager()
	row := db.QueryRow("SELECT * FROM test")

	// QueryRow returns an empty sql.Row which will fail when scanned
	// We just verify it doesn't panic
	if row == nil {
		t.Error("Expected QueryRow to return a non-nil row")
	}
}

// TestRealDatabaseManager_IsConnected tests that IsConnected returns true for real manager
func TestRealDatabaseManager_IsConnected(t *testing.T) {
	db := NewDatabaseManager()

	// Before connection, should return false
	if db.IsConnected() {
		t.Error("Expected IsConnected() to return false before connection")
	}

	// After successful connection, should return true
	// Note: We'll test this in integration tests with actual database
}

// Helper function to check if a string contains a substring
func dbTestStringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
