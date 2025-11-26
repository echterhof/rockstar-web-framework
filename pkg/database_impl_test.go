//go:build test
// +build test

package pkg

import (
	"fmt"
	"strings"
)

// Test stub for databaseManager to test DSN building without CGO dependencies
type databaseManager struct{}

// NewDatabaseManager creates a new mock database manager for testing
func NewDatabaseManager() DatabaseManager {
	return newMockDatabaseManager()
}

// buildDSN constructs the data source name for different database drivers
func (dm *databaseManager) buildDSN(config DatabaseConfig) (string, error) {
	switch config.Driver {
	case "mysql":
		return dm.buildMySQLDSN(config), nil
	case "postgres":
		return dm.buildPostgresDSN(config), nil
	case "mssql", "sqlserver":
		return dm.buildMSSQLDSN(config), nil
	case "sqlite3", "sqlite":
		return dm.buildSQLiteDSN(config), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// buildMySQLDSN builds MySQL connection string
func (dm *databaseManager) buildMySQLDSN(config DatabaseConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	params := []string{}
	if config.Charset != "" {
		params = append(params, "charset="+config.Charset)
	}
	if config.Timezone != "" {
		params = append(params, "loc="+config.Timezone)
	}

	// Add custom options
	for key, value := range config.Options {
		params = append(params, key+"="+value)
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// buildPostgresDSN builds PostgreSQL connection string
func (dm *databaseManager) buildPostgresDSN(config DatabaseConfig) string {
	params := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("port=%d", config.Port),
		fmt.Sprintf("user=%s", config.Username),
		fmt.Sprintf("password=%s", config.Password),
		fmt.Sprintf("dbname=%s", config.Database),
	}

	if config.SSLMode != "" {
		params = append(params, "sslmode="+config.SSLMode)
	}
	if config.Timezone != "" {
		params = append(params, "timezone="+config.Timezone)
	}

	// Add custom options
	for key, value := range config.Options {
		params = append(params, key+"="+value)
	}

	return strings.Join(params, " ")
}

// buildMSSQLDSN builds MSSQL connection string
func (dm *databaseManager) buildMSSQLDSN(config DatabaseConfig) string {
	return fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s",
		config.Host, config.Port, config.Database, config.Username, config.Password)
}

// buildSQLiteDSN builds SQLite connection string with required pragmas
func (dm *databaseManager) buildSQLiteDSN(config DatabaseConfig) string {
	dsn := config.Database

	// Append SQLite-specific parameters for optimal performance and correctness
	params := []string{
		"_journal_mode=WAL",  // Enable Write-Ahead Logging for better concurrency
		"_foreign_keys=ON",   // Enable foreign key constraint enforcement
		"_busy_timeout=5000", // Wait up to 5 seconds when database is locked
	}

	// Add custom options from config
	for key, value := range config.Options {
		params = append(params, key+"="+value)
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
