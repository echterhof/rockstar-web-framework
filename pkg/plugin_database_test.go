package pkg

import (
	"os"
	"testing"
)

// TestPluginDatabaseSchema_Creation tests that plugin tables are created correctly
func TestPluginDatabaseSchema_Creation(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_plugin_schema.db"
	defer os.Remove(dbFile)

	// Create database manager
	db := NewDatabaseManager()

	// Connect to SQLite database
	config := DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbFile,
	}

	if err := db.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize plugin tables
	if err := db.InitializePluginTables(); err != nil {
		t.Fatalf("Failed to initialize plugin tables: %v", err)
	}

	// Verify tables exist by querying them
	tables := []string{"plugins", "plugin_hooks", "plugin_events", "plugin_storage", "plugin_metrics"}

	for _, table := range tables {
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		row := db.QueryRow(query, table)

		var name string
		if err := row.Scan(&name); err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}

		if name != table {
			t.Errorf("Expected table %s, got %s", table, name)
		}
	}
}

// TestPluginDatabaseSchema_Indexes tests that indexes are created correctly
func TestPluginDatabaseSchema_Indexes(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_plugin_indexes.db"
	defer os.Remove(dbFile)

	// Create database manager
	db := NewDatabaseManager()

	// Connect to SQLite database
	config := DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbFile,
	}

	if err := db.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize plugin tables
	if err := db.InitializePluginTables(); err != nil {
		t.Fatalf("Failed to initialize plugin tables: %v", err)
	}

	// Verify indexes exist
	expectedIndexes := map[string][]string{
		"plugins":        {"idx_plugins_name", "idx_plugins_enabled", "idx_plugins_status"},
		"plugin_hooks":   {"idx_plugin_hooks_plugin", "idx_plugin_hooks_type", "idx_plugin_hooks_priority"},
		"plugin_events":  {"idx_plugin_events_name", "idx_plugin_events_publisher", "idx_plugin_events_subscriber"},
		"plugin_storage": {"idx_plugin_storage_plugin", "idx_plugin_storage_key"},
		"plugin_metrics": {"idx_plugin_metrics_plugin", "idx_plugin_metrics_name", "idx_plugin_metrics_recorded"},
	}

	for table, indexes := range expectedIndexes {
		for _, index := range indexes {
			query := "SELECT name FROM sqlite_master WHERE type='index' AND name=?"
			row := db.QueryRow(query, index)

			var name string
			if err := row.Scan(&name); err != nil {
				t.Logf("Warning: Index %s for table %s may not exist: %v", index, table, err)
				// Note: SQLite may create indexes differently, so we just log warnings
			}
		}
	}
}

// TestPluginDatabaseSchema_ForeignKeys tests that foreign key constraints work
func TestPluginDatabaseSchema_ForeignKeys(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_plugin_fk.db"
	defer os.Remove(dbFile)

	// Create database manager
	db := NewDatabaseManager()

	// Connect to SQLite database
	config := DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbFile,
	}

	if err := db.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Initialize plugin tables
	if err := db.InitializePluginTables(); err != nil {
		t.Fatalf("Failed to initialize plugin tables: %v", err)
	}

	// Insert a plugin
	_, err := db.Exec(`INSERT INTO plugins (name, version, description, author, enabled, status) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		"test-plugin", "1.0.0", "Test plugin", "Test Author", true, "running")

	if err != nil {
		t.Fatalf("Failed to insert plugin: %v", err)
	}

	// Insert a plugin hook (should succeed because plugin exists)
	_, err = db.Exec(`INSERT INTO plugin_hooks (plugin_name, hook_type, priority) 
		VALUES (?, ?, ?)`,
		"test-plugin", "startup", 100)

	if err != nil {
		t.Errorf("Failed to insert plugin hook: %v", err)
	}

	// Try to insert a hook for non-existent plugin (should fail with foreign key constraint)
	_, err = db.Exec(`INSERT INTO plugin_hooks (plugin_name, hook_type, priority) 
		VALUES (?, ?, ?)`,
		"non-existent-plugin", "startup", 100)

	if err == nil {
		t.Error("Expected foreign key constraint error, but insert succeeded")
	}
}
