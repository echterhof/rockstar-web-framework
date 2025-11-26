//go:build !test
// +build !test

package pkg

import (
	"path/filepath"
	"testing"
	"time"
)

// Helper function to create a test database configuration
func createTestDBConfig(dbPath string) DatabaseConfig {
	return DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbPath,
		Options: map[string]string{
			"sql_dir": "../sql",
		},
	}
}

// TestIntegration_SQLiteSessionOperations tests session operations with real SQLite database
func TestIntegration_SQLiteSessionOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sessions.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test session save
	session := &Session{
		ID:        "test_session_123",
		UserID:    "user_456",
		TenantID:  "tenant_789",
		Data:      map[string]interface{}{"key": "value", "number": float64(42)},
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "Test Agent",
	}

	if err := dm.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test session load
	loadedSession, err := dm.LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify session data
	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, loadedSession.ID)
	}
	if loadedSession.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, loadedSession.UserID)
	}
	if loadedSession.TenantID != session.TenantID {
		t.Errorf("Expected tenant ID %s, got %s", session.TenantID, loadedSession.TenantID)
	}

	// Test session update
	session.Data["updated"] = "yes"
	session.UpdatedAt = time.Now()
	if err := dm.SaveSession(session); err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Verify update
	updatedSession, err := dm.LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load updated session: %v", err)
	}
	if updatedSession.Data["updated"] != "yes" {
		t.Errorf("Expected updated data, got %v", updatedSession.Data)
	}

	// Test session deletion
	if err := dm.DeleteSession(session.ID); err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify deletion
	_, err = dm.LoadSession(session.ID)
	if err == nil {
		t.Error("Expected error when loading deleted session")
	}
}

// TestIntegration_SQLiteTokenOperations tests token operations with real SQLite database
func TestIntegration_SQLiteTokenOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_tokens.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test token save
	token := &AccessToken{
		Token:     "test_token_abc123",
		UserID:    "user_456",
		TenantID:  "tenant_789",
		Scopes:    []string{"read", "write", "admin"},
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if err := dm.SaveAccessToken(token); err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Test token load
	loadedToken, err := dm.LoadAccessToken(token.Token)
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	// Verify token data
	if loadedToken.Token != token.Token {
		t.Errorf("Expected token %s, got %s", token.Token, loadedToken.Token)
	}
	if loadedToken.UserID != token.UserID {
		t.Errorf("Expected user ID %s, got %s", token.UserID, loadedToken.UserID)
	}
	if len(loadedToken.Scopes) != len(token.Scopes) {
		t.Errorf("Expected %d scopes, got %d", len(token.Scopes), len(loadedToken.Scopes))
	}

	// Test token validation
	validToken, err := dm.ValidateAccessToken(token.Token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if validToken.Token != token.Token {
		t.Errorf("Expected validated token %s, got %s", token.Token, validToken.Token)
	}

	// Test token deletion
	if err := dm.DeleteAccessToken(token.Token); err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify deletion
	_, err = dm.LoadAccessToken(token.Token)
	if err == nil {
		t.Error("Expected error when loading deleted token")
	}
}

// TestIntegration_SQLiteTenantOperations tests tenant operations with real SQLite database
func TestIntegration_SQLiteTenantOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_tenants.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test tenant save
	tenant := &Tenant{
		ID:          "tenant_123",
		Name:        "Test Tenant",
		Hosts:       []string{"example.com", "test.example.com"},
		Config:      map[string]interface{}{"theme": "dark", "max_users": float64(100)},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MaxUsers:    1000,
		MaxStorage:  1024 * 1024 * 1024, // 1GB
		MaxRequests: 10000,
	}

	if err := dm.SaveTenant(tenant); err != nil {
		t.Fatalf("Failed to save tenant: %v", err)
	}

	// Test tenant load
	loadedTenant, err := dm.LoadTenant(tenant.ID)
	if err != nil {
		t.Fatalf("Failed to load tenant: %v", err)
	}

	// Verify tenant data
	if loadedTenant.ID != tenant.ID {
		t.Errorf("Expected tenant ID %s, got %s", tenant.ID, loadedTenant.ID)
	}
	if loadedTenant.Name != tenant.Name {
		t.Errorf("Expected tenant name %s, got %s", tenant.Name, loadedTenant.Name)
	}
	if len(loadedTenant.Hosts) != len(tenant.Hosts) {
		t.Errorf("Expected %d hosts, got %d", len(tenant.Hosts), len(loadedTenant.Hosts))
	}

	// Test load tenant by host
	loadedByHost, err := dm.LoadTenantByHost("example.com")
	if err != nil {
		t.Fatalf("Failed to load tenant by host: %v", err)
	}
	if loadedByHost.ID != tenant.ID {
		t.Errorf("Expected tenant ID %s, got %s", tenant.ID, loadedByHost.ID)
	}
}

// TestIntegration_SQLiteMetricsOperations tests metrics operations with real SQLite database
func TestIntegration_SQLiteMetricsOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_metrics.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Test metrics save
	metrics := &WorkloadMetrics{
		Timestamp:    time.Now(),
		TenantID:     "tenant_123",
		UserID:       "user_456",
		RequestID:    "req_789",
		Duration:     150,
		ContextSize:  1024,
		MemoryUsage:  2048,
		CPUUsage:     0.75,
		Path:         "/api/test",
		Method:       "GET",
		StatusCode:   200,
		ResponseSize: 512,
		ErrorMessage: "",
	}

	if err := dm.SaveWorkloadMetrics(metrics); err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Test metrics retrieval
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)

	loadedMetrics, err := dm.GetWorkloadMetrics(metrics.TenantID, from, to)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if len(loadedMetrics) == 0 {
		t.Error("Expected at least one metric")
	}

	if len(loadedMetrics) > 0 {
		metric := loadedMetrics[0]
		if metric.TenantID != metrics.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", metrics.TenantID, metric.TenantID)
		}
		if metric.Duration != metrics.Duration {
			t.Errorf("Expected duration %d, got %d", metrics.Duration, metric.Duration)
		}
	}
}

// TestIntegration_SQLiteRateLimitingOperations tests rate limiting with real SQLite database
func TestIntegration_SQLiteRateLimitingOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ratelimit.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	key := "test_rate_limit"
	limit := 3
	window := time.Minute

	// Test initial rate limit check (should be false)
	exceeded, err := dm.CheckRateLimit(key, limit, window)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}
	if exceeded {
		t.Error("Expected rate limit not exceeded initially")
	}

	// Increment rate limit multiple times
	for i := 0; i < limit; i++ {
		if err := dm.IncrementRateLimit(key, window); err != nil {
			t.Fatalf("Failed to increment rate limit: %v", err)
		}
	}

	// Check if rate limit is now exceeded
	exceeded, err = dm.CheckRateLimit(key, limit, window)
	if err != nil {
		t.Fatalf("Failed to check rate limit after increments: %v", err)
	}
	if !exceeded {
		t.Error("Expected rate limit to be exceeded")
	}
}

// TestIntegration_SQLiteCleanupOperations tests cleanup operations with real SQLite database
func TestIntegration_SQLiteCleanupOperations(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cleanup.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Create tables
	if err := dm.CreateTables(); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Create expired session
	expiredSession := &Session{
		ID:        "expired_session",
		UserID:    "user_123",
		TenantID:  "tenant_456",
		Data:      map[string]interface{}{"test": "data"},
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour),
		IPAddress: "127.0.0.1",
		UserAgent: "Test",
	}

	if err := dm.SaveSession(expiredSession); err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

	// Create expired token
	expiredToken := &AccessToken{
		Token:     "expired_token",
		UserID:    "user_123",
		TenantID:  "tenant_456",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	if err := dm.SaveAccessToken(expiredToken); err != nil {
		t.Fatalf("Failed to save expired token: %v", err)
	}

	// Test cleanup operations
	if err := dm.CleanupExpiredSessions(); err != nil {
		t.Fatalf("Failed to cleanup expired sessions: %v", err)
	}

	if err := dm.CleanupExpiredTokens(); err != nil {
		t.Fatalf("Failed to cleanup expired tokens: %v", err)
	}

	// Verify expired items were cleaned up
	_, err := dm.LoadSession(expiredSession.ID)
	if err == nil {
		t.Error("Expected error when loading expired session after cleanup")
	}

	_, err = dm.LoadAccessToken(expiredToken.Token)
	if err == nil {
		t.Error("Expected error when loading expired token after cleanup")
	}
}

// TestIntegration_SQLitePragmas verifies SQLite is configured with correct pragmas
func TestIntegration_SQLitePragmas(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_pragmas.db")

	// Create database manager
	dm := NewDatabaseManager()
	config := createTestDBConfig(dbPath)

	// Connect to database
	if err := dm.Connect(config); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dm.Close()

	// Query pragma settings
	var journalMode string
	err := dm.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	if journalMode != "wal" && journalMode != "WAL" {
		t.Errorf("Expected journal_mode to be WAL, got %s", journalMode)
	}

	var foreignKeys int
	err = dm.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys: %v", err)
	}

	if foreignKeys != 1 {
		t.Errorf("Expected foreign_keys to be enabled (1), got %d", foreignKeys)
	}
}
