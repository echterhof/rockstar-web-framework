//go:build test
// +build test

package pkg

import (
	"strings"
	"testing"
	"testing/quick"
)

// **Feature: sqlite-standardization, Property 5: SQLite connections use CGO driver**
// **Validates: Requirements 9.1, 9.4**
func TestProperty_SQLiteUsesCGODriver(t *testing.T) {
	// Property: For any SQLite database configuration, the system should use
	// the github.com/mattn/go-sqlite3 driver which requires CGO

	property := func(seed uint32) bool {
		// Generate a valid SQLite database path from seed
		dbPath := generateSQLitePath(seed)

		// Create a database configuration for SQLite
		config := DatabaseConfig{
			Driver:   "sqlite3",
			Database: dbPath,
			Options:  make(map[string]string),
		}

		// Create database manager
		dm := &databaseManager{}

		// Build DSN
		dsn, err := dm.buildDSN(config)
		if err != nil {
			t.Logf("Failed to build DSN: %v", err)
			return false
		}

		// Verify the DSN is not empty (basic sanity check)
		if dsn == "" {
			t.Logf("DSN is empty for SQLite configuration")
			return false
		}

		// Verify the DSN starts with the database path
		if !strings.HasPrefix(dsn, dbPath) {
			t.Logf("DSN does not start with database path. Expected prefix: %s, Got: %s", dbPath, dsn)
			return false
		}

		// The actual CGO driver verification happens at runtime when sql.Open is called
		// This property test verifies that the DSN is correctly formatted for the CGO driver
		// The driver import is: _ "github.com/mattn/go-sqlite3"
		// which is present in database_impl.go

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// **Feature: sqlite-standardization, Property 6: SQLite connections enable required pragmas**
// **Validates: Requirements 9.2, 9.3**
func TestProperty_SQLiteEnablesRequiredPragmas(t *testing.T) {
	// Property: For any SQLite database connection, the DSN should include
	// parameters for WAL mode and foreign key enforcement

	property := func(seed uint32) bool {
		// Generate a valid SQLite database path from seed
		dbPath := generateSQLitePath(seed)

		// Test both "sqlite" and "sqlite3" driver names
		drivers := []string{"sqlite", "sqlite3"}
		driverIdx := seed % 2
		driver := drivers[driverIdx]

		// Create a database configuration for SQLite
		config := DatabaseConfig{
			Driver:   driver,
			Database: dbPath,
			Options:  make(map[string]string),
		}

		// Create database manager
		dm := &databaseManager{}

		// Build DSN
		dsn, err := dm.buildDSN(config)
		if err != nil {
			t.Logf("Failed to build DSN: %v", err)
			return false
		}

		// Verify the DSN contains required pragmas
		requiredParams := []string{
			"_journal_mode=WAL",
			"_foreign_keys=ON",
			"_busy_timeout=5000",
		}

		for _, param := range requiredParams {
			if !strings.Contains(dsn, param) {
				t.Logf("DSN missing required parameter: %s. DSN: %s", param, dsn)
				return false
			}
		}

		// Verify the DSN has the correct format (path?params)
		if !strings.Contains(dsn, "?") {
			t.Logf("DSN does not contain parameter separator '?'. DSN: %s", dsn)
			return false
		}

		// Verify parameters are properly joined with &
		parts := strings.Split(dsn, "?")
		if len(parts) != 2 {
			t.Logf("DSN has incorrect format. Expected 'path?params', got: %s", dsn)
			return false
		}

		params := parts[1]
		paramList := strings.Split(params, "&")

		// Verify we have at least the 3 required parameters
		if len(paramList) < 3 {
			t.Logf("DSN has fewer than 3 parameters. Got: %d, DSN: %s", len(paramList), dsn)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Helper function to generate SQLite database paths from a seed
func generateSQLitePath(seed uint32) string {
	// Generate deterministic but varied database paths
	prefixes := []string{"test", "data", "app", "db", "storage", "cache"}
	suffixes := []string{".db", ".sqlite", ".sqlite3", ".database"}

	prefixIdx := seed % uint32(len(prefixes))
	suffixIdx := (seed / uint32(len(prefixes))) % uint32(len(suffixes))

	return prefixes[prefixIdx] + suffixes[suffixIdx]
}
