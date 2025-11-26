//go:build test
// +build test

package pkg

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// mockDatabaseManager implements DatabaseManager for testing without requiring CGO
type mockDatabaseManager struct {
	connected  bool
	sessions   map[string]*Session
	tokens     map[string]*AccessToken
	tenants    map[string]*Tenant
	metrics    []*WorkloadMetrics
	rateLimits map[string][]time.Time
}

func newMockDatabaseManager() *mockDatabaseManager {
	return &mockDatabaseManager{
		sessions:   make(map[string]*Session),
		tokens:     make(map[string]*AccessToken),
		tenants:    make(map[string]*Tenant),
		metrics:    make([]*WorkloadMetrics, 0),
		rateLimits: make(map[string][]time.Time),
	}
}

// NewMockDatabaseManager creates a new mock database manager for testing (exported)
func NewMockDatabaseManager() DatabaseManager {
	return newMockDatabaseManager()
}

// Mock implementation of DatabaseManager interface for testing

func (m *mockDatabaseManager) Connect(config DatabaseConfig) error {
	m.connected = true
	return nil
}

func (m *mockDatabaseManager) Close() error {
	m.connected = false
	return nil
}

func (m *mockDatabaseManager) Ping() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *mockDatabaseManager) Stats() DatabaseStats {
	return DatabaseStats{OpenConnections: 1}
}

func (m *mockDatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	return nil, fmt.Errorf("mock query not implemented")
}

func (m *mockDatabaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}

func (m *mockDatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	return &mockResult{rowsAffected: 1}, nil
}

func (m *mockDatabaseManager) Prepare(query string) (*sql.Stmt, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	return nil, fmt.Errorf("mock prepare not implemented")
}

func (m *mockDatabaseManager) Begin() (Transaction, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	return &mockTransaction{}, nil
}

func (m *mockDatabaseManager) BeginTx(opts *sql.TxOptions) (Transaction, error) {
	return m.Begin()
}

// Session operations
func (m *mockDatabaseManager) SaveSession(session *Session) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockDatabaseManager) LoadSession(sessionID string) (*Session, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session expired")
	}
	return session, nil
}

func (m *mockDatabaseManager) DeleteSession(sessionID string) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockDatabaseManager) CleanupExpiredSessions() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	now := time.Now()
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(now) {
			delete(m.sessions, id)
		}
	}
	return nil
}

// Token operations
func (m *mockDatabaseManager) SaveAccessToken(token *AccessToken) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockDatabaseManager) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	token, exists := m.tokens[tokenValue]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

func (m *mockDatabaseManager) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	token, exists := m.tokens[tokenValue]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	if token.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}
	return token, nil
}

func (m *mockDatabaseManager) DeleteAccessToken(tokenValue string) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	delete(m.tokens, tokenValue)
	return nil
}

func (m *mockDatabaseManager) CleanupExpiredTokens() error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	now := time.Now()
	for token, tokenData := range m.tokens {
		if tokenData.ExpiresAt.Before(now) {
			delete(m.tokens, token)
		}
	}
	return nil
}

// Tenant operations
func (m *mockDatabaseManager) SaveTenant(tenant *Tenant) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *mockDatabaseManager) LoadTenant(tenantID string) (*Tenant, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, nil
}

func (m *mockDatabaseManager) LoadTenantByHost(hostname string) (*Tenant, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	for _, tenant := range m.tenants {
		for _, host := range tenant.Hosts {
			if host == hostname {
				return tenant, nil
			}
		}
	}
	return nil, fmt.Errorf("tenant not found for host: %s", hostname)
}

// Metrics operations
func (m *mockDatabaseManager) SaveWorkloadMetrics(metrics *WorkloadMetrics) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.metrics = append(m.metrics, metrics)
	return nil
}

func (m *mockDatabaseManager) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	var result []*WorkloadMetrics
	for _, metric := range m.metrics {
		if metric.TenantID == tenantID &&
			metric.Timestamp.After(from) &&
			metric.Timestamp.Before(to) {
			result = append(result, metric)
		}
	}
	return result, nil
}

// Rate limiting operations
func (m *mockDatabaseManager) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	if !m.connected {
		return false, fmt.Errorf("not connected")
	}

	now := time.Now()
	windowStart := now.Add(-window)

	// Clean up old entries
	var validEntries []time.Time
	for _, timestamp := range m.rateLimits[key] {
		if timestamp.After(windowStart) {
			validEntries = append(validEntries, timestamp)
		}
	}
	m.rateLimits[key] = validEntries

	return len(validEntries) >= limit, nil
}

func (m *mockDatabaseManager) IncrementRateLimit(key string, window time.Duration) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}

	now := time.Now()
	m.rateLimits[key] = append(m.rateLimits[key], now)
	return nil
}

// Migration operations
func (m *mockDatabaseManager) Migrate() error {
	return nil
}

func (m *mockDatabaseManager) CreateTables() error {
	return nil
}

func (m *mockDatabaseManager) DropTables() error {
	return nil
}

func (m *mockDatabaseManager) InitializePluginTables() error {
	return nil
}

func (m *mockDatabaseManager) GetQuery(name string) (string, error) {
	return "", nil
}

// Mock result and transaction types
type mockResult struct {
	rowsAffected int64
}

func (r *mockResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *mockResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type mockTransaction struct{}

func (t *mockTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("mock transaction query not implemented")
}

func (t *mockTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}

func (t *mockTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return &mockResult{rowsAffected: 1}, nil
}

func (t *mockTransaction) Prepare(query string) (*sql.Stmt, error) {
	return nil, fmt.Errorf("mock transaction prepare not implemented")
}

func (t *mockTransaction) Commit() error {
	return nil
}

func (t *mockTransaction) Rollback() error {
	return nil
}

// Test functions

func TestMockDatabaseManager_BasicOperations(t *testing.T) {
	dm := newMockDatabaseManager()

	// Test connection
	config := DatabaseConfig{
		Driver:   "mock",
		Database: "test",
	}

	if err := dm.Connect(config); err != nil {
		t.Errorf("Connect() failed: %v", err)
	}

	// Test Ping
	if err := dm.Ping(); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}

	// Test Stats
	stats := dm.Stats()
	if stats.OpenConnections != 1 {
		t.Errorf("Expected 1 open connection, got %d", stats.OpenConnections)
	}

	// Test basic exec operation
	result, err := dm.Exec("INSERT INTO test_table (name) VALUES (?)", "test")
	if err != nil {
		t.Errorf("Exec() failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Errorf("RowsAffected() failed: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Test Close
	if err := dm.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestMockDatabaseManager_SessionOperations(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	// Test session save and load
	session := &Session{
		ID:        "test_session_123",
		UserID:    "user_456",
		TenantID:  "tenant_789",
		Data:      map[string]interface{}{"key": "value", "number": 42},
		ExpiresAt: time.Now().Add(time.Hour),
		IPAddress: "192.168.1.1",
		UserAgent: "Test Agent",
	}

	// Save session
	if err := dm.SaveSession(session); err != nil {
		t.Errorf("SaveSession() failed: %v", err)
	}

	// Load session
	loadedSession, err := dm.LoadSession(session.ID)
	if err != nil {
		t.Errorf("LoadSession() failed: %v", err)
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
	if loadedSession.Data["key"] != "value" {
		t.Errorf("Expected data key 'value', got %v", loadedSession.Data["key"])
	}

	// Test session deletion
	if err := dm.DeleteSession(session.ID); err != nil {
		t.Errorf("DeleteSession() failed: %v", err)
	}

	// Verify session is deleted
	_, err = dm.LoadSession(session.ID)
	if err == nil {
		t.Error("Expected error when loading deleted session")
	}
}

func TestMockDatabaseManager_AccessTokenOperations(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	// Test access token save and load
	token := &AccessToken{
		Token:     "test_token_abc123",
		UserID:    "user_456",
		TenantID:  "tenant_789",
		Scopes:    []string{"read", "write", "admin"},
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Save token
	if err := dm.SaveAccessToken(token); err != nil {
		t.Errorf("SaveAccessToken() failed: %v", err)
	}

	// Load token
	loadedToken, err := dm.LoadAccessToken(token.Token)
	if err != nil {
		t.Errorf("LoadAccessToken() failed: %v", err)
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
		t.Errorf("ValidateAccessToken() failed: %v", err)
	}
	if validToken.Token != token.Token {
		t.Errorf("Expected validated token %s, got %s", token.Token, validToken.Token)
	}

	// Test token deletion
	if err := dm.DeleteAccessToken(token.Token); err != nil {
		t.Errorf("DeleteAccessToken() failed: %v", err)
	}

	// Verify token is deleted
	_, err = dm.LoadAccessToken(token.Token)
	if err == nil {
		t.Error("Expected error when loading deleted token")
	}
}

func TestMockDatabaseManager_TenantOperations(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	// Test tenant save and load
	tenant := &Tenant{
		ID:          "tenant_123",
		Name:        "Test Tenant",
		Hosts:       []string{"example.com", "test.example.com"},
		Config:      map[string]interface{}{"theme": "dark", "max_users": 100},
		IsActive:    true,
		MaxUsers:    1000,
		MaxStorage:  1024 * 1024 * 1024, // 1GB
		MaxRequests: 10000,
	}

	// Save tenant
	if err := dm.SaveTenant(tenant); err != nil {
		t.Errorf("SaveTenant() failed: %v", err)
	}

	// Load tenant
	loadedTenant, err := dm.LoadTenant(tenant.ID)
	if err != nil {
		t.Errorf("LoadTenant() failed: %v", err)
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
	if loadedTenant.MaxUsers != tenant.MaxUsers {
		t.Errorf("Expected max users %d, got %d", tenant.MaxUsers, loadedTenant.MaxUsers)
	}

	// Test load tenant by host
	loadedByHost, err := dm.LoadTenantByHost("example.com")
	if err != nil {
		t.Errorf("LoadTenantByHost() failed: %v", err)
	}
	if loadedByHost.ID != tenant.ID {
		t.Errorf("Expected tenant ID %s, got %s", tenant.ID, loadedByHost.ID)
	}
}

func TestMockDatabaseManager_WorkloadMetrics(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	// Test workload metrics save
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

	// Save metrics
	if err := dm.SaveWorkloadMetrics(metrics); err != nil {
		t.Errorf("SaveWorkloadMetrics() failed: %v", err)
	}

	// Get metrics
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)

	loadedMetrics, err := dm.GetWorkloadMetrics(metrics.TenantID, from, to)
	if err != nil {
		t.Errorf("GetWorkloadMetrics() failed: %v", err)
	}

	if len(loadedMetrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(loadedMetrics))
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

func TestMockDatabaseManager_RateLimiting(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	key := "test_rate_limit"
	limit := 3
	window := time.Minute

	// Test initial rate limit check (should be false)
	exceeded, err := dm.CheckRateLimit(key, limit, window)
	if err != nil {
		t.Errorf("CheckRateLimit() failed: %v", err)
	}
	if exceeded {
		t.Error("Expected rate limit not exceeded initially")
	}

	// Increment rate limit multiple times
	for i := 0; i < limit; i++ {
		if err := dm.IncrementRateLimit(key, window); err != nil {
			t.Errorf("IncrementRateLimit() failed: %v", err)
		}
	}

	// Check if rate limit is now exceeded
	exceeded, err = dm.CheckRateLimit(key, limit, window)
	if err != nil {
		t.Errorf("CheckRateLimit() after increments failed: %v", err)
	}
	if !exceeded {
		t.Error("Expected rate limit to be exceeded")
	}
}

func TestMockDatabaseManager_CleanupOperations(t *testing.T) {
	dm := newMockDatabaseManager()
	dm.Connect(DatabaseConfig{Driver: "mock"})
	defer dm.Close()

	// Create expired session
	expiredSession := &Session{
		ID:        "expired_session",
		UserID:    "user_123",
		TenantID:  "tenant_456",
		Data:      map[string]interface{}{"test": "data"},
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
	}

	if err := dm.SaveSession(expiredSession); err != nil {
		t.Errorf("SaveSession() for expired session failed: %v", err)
	}

	// Create expired token
	expiredToken := &AccessToken{
		Token:     "expired_token",
		UserID:    "user_123",
		TenantID:  "tenant_456",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
	}

	if err := dm.SaveAccessToken(expiredToken); err != nil {
		t.Errorf("SaveAccessToken() for expired token failed: %v", err)
	}

	// Test cleanup operations
	if err := dm.CleanupExpiredSessions(); err != nil {
		t.Errorf("CleanupExpiredSessions() failed: %v", err)
	}

	if err := dm.CleanupExpiredTokens(); err != nil {
		t.Errorf("CleanupExpiredTokens() failed: %v", err)
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

func TestDatabaseManager_DSNBuilding(t *testing.T) {
	// Create a real database manager instance to test DSN building
	dm := &databaseManager{}

	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "MySQL DSN",
			config: DatabaseConfig{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				Charset:  "utf8mb4",
			},
			expected: "user:pass@tcp(localhost:3306)/testdb?charset=utf8mb4",
		},
		{
			name: "PostgreSQL DSN",
			config: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable",
		},
		{
			name: "SQLite DSN",
			config: DatabaseConfig{
				Driver:   "sqlite3",
				Database: "/path/to/db.sqlite",
			},
			expected: "/path/to/db.sqlite?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000",
		},
		{
			name: "MSSQL DSN",
			config: DatabaseConfig{
				Driver:   "mssql",
				Host:     "localhost",
				Port:     1433,
				Database: "testdb",
				Username: "user",
				Password: "pass",
			},
			expected: "server=localhost;port=1433;database=testdb;user id=user;password=pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := dm.buildDSN(tt.config)
			if err != nil {
				t.Errorf("buildDSN() failed: %v", err)
				return
			}

			if dsn != tt.expected {
				t.Errorf("buildDSN() = %v, expected %v", dsn, tt.expected)
			}
		})
	}

	// Test invalid driver
	invalidConfig := DatabaseConfig{
		Driver:   "invalid_driver",
		Database: "test.db",
	}

	_, err := dm.buildDSN(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid driver")
	}
}
