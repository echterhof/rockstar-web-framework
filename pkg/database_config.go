package pkg

// isDatabaseConfigured checks if a database configuration is complete and valid
// Returns true if the configuration has all required fields for a database connection
func isDatabaseConfigured(config DatabaseConfig) bool {
	// Empty driver means no database configured
	if config.Driver == "" {
		return false
	}

	// Empty database name means no database configured
	if config.Database == "" {
		return false
	}

	// SQLite only needs driver and database path
	if config.Driver == "sqlite" || config.Driver == "sqlite3" {
		return true
	}

	// Other drivers need credentials (username and password)
	// Both must be non-empty for a valid configuration
	return config.Username != "" && config.Password != ""
}

// isNoopDatabase checks if a DatabaseManager is a no-op implementation
// Returns true if the manager is the no-op type that returns errors for all operations
func isNoopDatabase(db DatabaseManager) bool {
	if db == nil {
		return true
	}
	_, ok := db.(*noopDatabaseManager)
	return ok
}
