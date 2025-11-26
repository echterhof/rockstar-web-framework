package pkg

import (
	"os"
	"testing"
	"time"
)

// TestTenantOperations_WithSQLLoader tests tenant operations using the SQL loader
func TestTenantOperations_WithSQLLoader(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_tenant_operations.db"
	defer os.Remove(dbFile)

	// Create database manager
	db := NewDatabaseManager()

	// Connect to SQLite database
	config := DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbFile,
		Options: map[string]string{
			"sql_dir": "../sql", // SQL directory is in project root, one level up from pkg/
		},
	}

	if err := db.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create the tenants table manually for this test
	// We'll use the SQL loader to get the create table query
	createTableQuery := `CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		hosts TEXT NOT NULL,
		config TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		max_users INTEGER NOT NULL DEFAULT 0,
		max_storage INTEGER NOT NULL DEFAULT 0,
		max_requests INTEGER NOT NULL DEFAULT 0
	)`

	if _, err := db.Exec(createTableQuery); err != nil {
		t.Fatalf("Failed to create tenants table: %v", err)
	}

	// Test SaveTenant
	tenant := &Tenant{
		ID:          "tenant_test_123",
		Name:        "Test Tenant",
		Hosts:       []string{"example.com", "test.example.com"},
		Config:      map[string]interface{}{"theme": "dark", "max_users": 100},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MaxUsers:    1000,
		MaxStorage:  1024 * 1024 * 1024, // 1GB
		MaxRequests: 10000,
	}

	if err := db.SaveTenant(tenant); err != nil {
		t.Fatalf("SaveTenant() failed: %v", err)
	}

	// Test LoadTenant
	loadedTenant, err := db.LoadTenant(tenant.ID)
	if err != nil {
		t.Fatalf("LoadTenant() failed: %v", err)
	}

	if loadedTenant.ID != tenant.ID {
		t.Errorf("Expected tenant ID %s, got %s", tenant.ID, loadedTenant.ID)
	}
	if loadedTenant.Name != tenant.Name {
		t.Errorf("Expected tenant name %s, got %s", tenant.Name, loadedTenant.Name)
	}
	if len(loadedTenant.Hosts) != len(tenant.Hosts) {
		t.Errorf("Expected %d hosts, got %d", len(tenant.Hosts), len(loadedTenant.Hosts))
	}

	// Test LoadTenantByHost
	loadedByHost, err := db.LoadTenantByHost("example.com")
	if err != nil {
		t.Fatalf("LoadTenantByHost() failed: %v", err)
	}

	if loadedByHost.ID != tenant.ID {
		t.Errorf("Expected tenant ID %s from LoadTenantByHost, got %s", tenant.ID, loadedByHost.ID)
	}
}
