package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SQLLoader interface for loading and managing SQL statements from external files
type SQLLoader interface {
	// GetQuery returns the SQL statement for the given query name
	GetQuery(name string) (string, error)

	// HasQuery checks if a query exists
	HasQuery(name string) bool

	// LoadAll loads all SQL files for the configured driver
	LoadAll() error

	// GetDriver returns the current database driver name
	GetDriver() string
}

// sqlLoader implements the SQLLoader interface
type sqlLoader struct {
	driver  string
	queries map[string]string
	sqlDir  string
}

// NewSQLLoader creates a new SQL loader instance
// driver: database driver name (sqlite, mysql, postgres, mssql)
// sqlDir: base directory containing SQL files (e.g., "./sql")
func NewSQLLoader(driver string, sqlDir string) (SQLLoader, error) {
	// Validate driver
	validDrivers := map[string]bool{
		"sqlite":    true,
		"sqlite3":   true,
		"mysql":     true,
		"postgres":  true,
		"mssql":     true,
		"sqlserver": true,
	}

	// Normalize driver name
	normalizedDriver := driver
	if driver == "sqlite3" {
		normalizedDriver = "sqlite"
	} else if driver == "sqlserver" {
		normalizedDriver = "mssql"
	}

	if !validDrivers[driver] {
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	loader := &sqlLoader{
		driver:  normalizedDriver,
		queries: make(map[string]string),
		sqlDir:  sqlDir,
	}

	return loader, nil
}

// LoadAll loads all SQL files from the driver-specific directory
func (sl *sqlLoader) LoadAll() error {
	// Validate driver name to prevent path traversal
	validator := NewPathValidator(sl.sqlDir)
	driverDir, err := validator.ResolvePath(sl.driver)
	if err != nil {
		return fmt.Errorf("invalid driver name: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(driverDir); os.IsNotExist(err) {
		return fmt.Errorf("SQL directory not found: %s", driverDir)
	}

	// Read all .sql files from the directory
	entries, err := os.ReadDir(driverDir)
	if err != nil {
		return fmt.Errorf("failed to read SQL directory %s: %w", driverDir, err)
	}

	// Load each SQL file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .sql files
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Validate filename to prevent path traversal
		filePath, err := validator.ResolvePath(filepath.Join(sl.driver, entry.Name()))
		if err != nil {
			return fmt.Errorf("invalid SQL filename: %w", err)
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read SQL file %s: %w", filePath, err)
		}

		// Extract query name from filename (remove .sql extension)
		queryName := strings.TrimSuffix(entry.Name(), ".sql")

		// Store the query
		sl.queries[queryName] = string(content)
	}

	return nil
}

// GetQuery returns the SQL statement for the given query name
func (sl *sqlLoader) GetQuery(name string) (string, error) {
	query, exists := sl.queries[name]
	if !exists {
		return "", fmt.Errorf("query not found: %s", name)
	}
	return query, nil
}

// HasQuery checks if a query exists
func (sl *sqlLoader) HasQuery(name string) bool {
	_, exists := sl.queries[name]
	return exists
}

// GetDriver returns the current database driver name
func (sl *sqlLoader) GetDriver() string {
	return sl.driver
}
