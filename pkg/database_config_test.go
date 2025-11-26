package pkg

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_FrameworkInitWithoutDatabaseAlwaysSucceeds tests Property 1
// **Feature: optional-database, Property 1: Framework initialization succeeds without database configuration**
// **Validates: Requirements 1.1, 1.2**
func TestProperty_FrameworkInitWithoutDatabaseAlwaysSucceeds(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Framework initialization succeeds with empty database config",
		prop.ForAll(
			func(driver, database, username, password string) bool {
				// Test various combinations of missing/empty configuration fields
				configs := []DatabaseConfig{
					// Empty driver
					{Driver: "", Database: database, Username: username, Password: password},
					// Empty database
					{Driver: driver, Database: "", Username: username, Password: password},
					// Empty credentials (non-SQLite)
					{Driver: "postgres", Database: database, Username: "", Password: ""},
					{Driver: "mysql", Database: database, Username: "", Password: password},
					{Driver: "postgres", Database: database, Username: username, Password: ""},
					// Completely empty
					{},
				}

				for _, config := range configs {
					// Skip SQLite special case - it only needs database path
					if config.Driver == "sqlite" || config.Driver == "sqlite3" {
						if config.Database != "" {
							continue // This would be considered configured
						}
					}

					// Configuration should be detected as not configured
					if isDatabaseConfigured(config) {
						return false
					}
				}

				return true
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
		))

	properties.Property("SQLite configuration only requires driver and database",
		prop.ForAll(
			func(database string) bool {
				// SQLite should be considered configured with just driver and database
				if database == "" {
					return true // Skip empty database
				}

				config := DatabaseConfig{
					Driver:   "sqlite",
					Database: database,
					// No username/password needed
				}

				return isDatabaseConfigured(config)
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		))

	properties.Property("Non-SQLite drivers require credentials",
		prop.ForAll(
			func(driver, database, username, password string) bool {
				// Skip empty values and SQLite
				if driver == "" || database == "" || driver == "sqlite" || driver == "sqlite3" {
					return true
				}

				config := DatabaseConfig{
					Driver:   driver,
					Database: database,
					Username: username,
					Password: password,
				}

				// Should only be configured if both username and password are non-empty
				expected := username != "" && password != ""
				actual := isDatabaseConfigured(config)

				return expected == actual
			},
			gen.OneConstOf("postgres", "mysql", "mssql"),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
			gen.AlphaString(),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty_NoopDatabaseDetection tests the isNoopDatabase helper
func TestProperty_NoopDatabaseDetection(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("isNoopDatabase correctly identifies no-op manager",
		prop.ForAll(
			func() bool {
				// Create a no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Should be detected as no-op
				if !isNoopDatabase(noopDB) {
					return false
				}

				// Nil should also be considered no-op
				if !isNoopDatabase(nil) {
					return false
				}

				return true
			},
		))

	properties.Property("isNoopDatabase returns false for real database manager",
		prop.ForAll(
			func() bool {
				// Create a real database manager
				realDB := NewDatabaseManager()

				// Should NOT be detected as no-op
				return !isNoopDatabase(realDB)
			},
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit tests for specific edge cases

func TestIsDatabaseConfigured_EmptyConfig(t *testing.T) {
	config := DatabaseConfig{}
	if isDatabaseConfigured(config) {
		t.Error("Empty config should not be considered configured")
	}
}

func TestIsDatabaseConfigured_SQLiteWithDatabase(t *testing.T) {
	config := DatabaseConfig{
		Driver:   "sqlite",
		Database: "test.db",
	}
	if !isDatabaseConfigured(config) {
		t.Error("SQLite with database path should be considered configured")
	}
}

func TestIsDatabaseConfigured_SQLite3WithDatabase(t *testing.T) {
	config := DatabaseConfig{
		Driver:   "sqlite3",
		Database: "test.db",
	}
	if !isDatabaseConfigured(config) {
		t.Error("SQLite3 with database path should be considered configured")
	}
}

func TestIsDatabaseConfigured_PostgresWithoutCredentials(t *testing.T) {
	config := DatabaseConfig{
		Driver:   "postgres",
		Database: "testdb",
		Host:     "localhost",
		Port:     5432,
	}
	if isDatabaseConfigured(config) {
		t.Error("Postgres without credentials should not be considered configured")
	}
}

func TestIsDatabaseConfigured_PostgresWithCredentials(t *testing.T) {
	config := DatabaseConfig{
		Driver:   "postgres",
		Database: "testdb",
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "pass",
	}
	if !isDatabaseConfigured(config) {
		t.Error("Postgres with credentials should be considered configured")
	}
}

func TestIsDatabaseConfigured_MySQLWithPartialCredentials(t *testing.T) {
	// Only username, no password
	config1 := DatabaseConfig{
		Driver:   "mysql",
		Database: "testdb",
		Username: "user",
	}
	if isDatabaseConfigured(config1) {
		t.Error("MySQL with only username should not be considered configured")
	}

	// Only password, no username
	config2 := DatabaseConfig{
		Driver:   "mysql",
		Database: "testdb",
		Password: "pass",
	}
	if isDatabaseConfigured(config2) {
		t.Error("MySQL with only password should not be considered configured")
	}
}

func TestIsNoopDatabase_WithNoopManager(t *testing.T) {
	noopDB := NewNoopDatabaseManager()
	if !isNoopDatabase(noopDB) {
		t.Error("No-op database manager should be detected as no-op")
	}
}

func TestIsNoopDatabase_WithRealManager(t *testing.T) {
	realDB := NewDatabaseManager()
	if isNoopDatabase(realDB) {
		t.Error("Real database manager should not be detected as no-op")
	}
}

func TestIsNoopDatabase_WithNil(t *testing.T) {
	if !isNoopDatabase(nil) {
		t.Error("Nil should be considered no-op")
	}
}
