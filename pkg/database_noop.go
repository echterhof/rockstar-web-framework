package pkg

import (
	"database/sql"
	"time"
)

// noopDatabaseManager implements DatabaseManager with no-op operations
// All methods return ErrNoDatabaseConfigured to indicate no database is configured
type noopDatabaseManager struct{}

// NewNoopDatabaseManager creates a new no-op database manager
func NewNoopDatabaseManager() DatabaseManager {
	return &noopDatabaseManager{}
}

// Connection management methods

func (n *noopDatabaseManager) Connect(config DatabaseConfig) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) Close() error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) Ping() error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) Stats() DatabaseStats {
	return DatabaseStats{}
}

func (n *noopDatabaseManager) IsConnected() bool {
	return false
}

// Query execution methods

func (n *noopDatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}

func (n *noopDatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, ErrNoDatabaseConfigured
}

// Prepared statements

func (n *noopDatabaseManager) Prepare(query string) (*sql.Stmt, error) {
	return nil, ErrNoDatabaseConfigured
}

// Transaction support

func (n *noopDatabaseManager) Begin() (Transaction, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) BeginTx(opts *sql.TxOptions) (Transaction, error) {
	return nil, ErrNoDatabaseConfigured
}

// Session operations

func (n *noopDatabaseManager) SaveSession(session *Session) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) LoadSession(sessionID string) (*Session, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) DeleteSession(sessionID string) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) CleanupExpiredSessions() error {
	return ErrNoDatabaseConfigured
}

// Access token operations

func (n *noopDatabaseManager) SaveAccessToken(token *AccessToken) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) DeleteAccessToken(tokenValue string) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) CleanupExpiredTokens() error {
	return ErrNoDatabaseConfigured
}

// Tenant operations

func (n *noopDatabaseManager) SaveTenant(tenant *Tenant) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) LoadTenant(tenantID string) (*Tenant, error) {
	return nil, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) LoadTenantByHost(hostname string) (*Tenant, error) {
	return nil, ErrNoDatabaseConfigured
}

// Workload metrics operations

func (n *noopDatabaseManager) SaveWorkloadMetrics(metrics *WorkloadMetrics) error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, ErrNoDatabaseConfigured
}

// Rate limiting operations

func (n *noopDatabaseManager) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return false, ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) IncrementRateLimit(key string, window time.Duration) error {
	return ErrNoDatabaseConfigured
}

// Migration support

func (n *noopDatabaseManager) Migrate() error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) CreateTables() error {
	return ErrNoDatabaseConfigured
}

func (n *noopDatabaseManager) DropTables() error {
	return ErrNoDatabaseConfigured
}

// Plugin system support

func (n *noopDatabaseManager) InitializePluginTables() error {
	return ErrNoDatabaseConfigured
}

// SQL loader support

func (n *noopDatabaseManager) GetQuery(name string) (string, error) {
	return "", ErrNoDatabaseConfigured
}
