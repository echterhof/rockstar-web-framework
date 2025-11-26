//go:build !test
// +build !test

package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

// databaseManager implements the DatabaseManager interface
type databaseManager struct {
	db        *sql.DB
	config    DatabaseConfig
	mutex     sync.RWMutex
	sqlLoader SQLLoader
}

// transaction implements the Transaction interface
type transaction struct {
	tx *sql.Tx
}

// NewDatabaseManager creates a new database manager instance
func NewDatabaseManager() DatabaseManager {
	return &databaseManager{}
}

// Connect establishes a database connection based on the provided configuration
func (dm *databaseManager) Connect(config DatabaseConfig) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Close existing connection if any
	if dm.db != nil {
		dm.db.Close()
	}

	// Build connection string based on driver
	dsn, err := dm.buildDSN(config)
	if err != nil {
		return fmt.Errorf("failed to build DSN: %w", err)
	}

	// Open database connection
	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	dm.db = db
	dm.config = config

	// Initialize SQL loader
	// Skip SQL loader for mock driver (used in tests)
	if config.Driver != "mock" {
		sqlDir := "./sql"
		if config.Options != nil {
			if customDir, ok := config.Options["sql_dir"]; ok {
				sqlDir = customDir
			}
		}

		loader, err := NewSQLLoader(config.Driver, sqlDir)
		if err != nil {
			return fmt.Errorf("failed to create SQL loader: %w", err)
		}

		if err := loader.LoadAll(); err != nil {
			return fmt.Errorf("failed to load SQL files: %w", err)
		}

		dm.sqlLoader = loader
	}

	return nil
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
	case "mock":
		// Mock driver for testing - return empty DSN
		return ":memory:", nil
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

// Close closes the database connection
func (dm *databaseManager) Close() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.db != nil {
		err := dm.db.Close()
		dm.db = nil
		return err
	}
	return nil
}

// Ping tests the database connection
func (dm *databaseManager) Ping() error {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return fmt.Errorf("database connection not established")
	}
	return dm.db.Ping()
}

// Stats returns database connection statistics
func (dm *databaseManager) Stats() DatabaseStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return DatabaseStats{}
	}

	stats := dm.db.Stats()
	return DatabaseStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

// Query executes a query that returns rows
func (dm *databaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return dm.db.Query(query, args...)
}

// QueryRow executes a query that returns at most one row
func (dm *databaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		// Return a row that will return an error when scanned
		return &sql.Row{}
	}
	return dm.db.QueryRow(query, args...)
}

// Exec executes a query without returning any rows
func (dm *databaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return dm.db.Exec(query, args...)
}

// Prepare creates a prepared statement
func (dm *databaseManager) Prepare(query string) (*sql.Stmt, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return dm.db.Prepare(query)
}

// Begin starts a transaction
func (dm *databaseManager) Begin() (Transaction, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	tx, err := dm.db.Begin()
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx}, nil
}

// BeginTx starts a transaction with options
func (dm *databaseManager) BeginTx(opts *sql.TxOptions) (Transaction, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if dm.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	tx, err := dm.db.BeginTx(nil, opts)
	if err != nil {
		return nil, err
	}

	return &transaction{tx: tx}, nil
}

// Transaction implementation
func (t *transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

func (t *transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

func (t *transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

func (t *transaction) Prepare(query string) (*sql.Stmt, error) {
	return t.tx.Prepare(query)
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}

// Framework-specific model operations
// Note: Session methods are implemented in session_impl.go

// SaveAccessToken saves an access token to the database
func (dm *databaseManager) SaveAccessToken(token *AccessToken) error {
	scopesJSON, err := json.Marshal(token.Scopes)
	if err != nil {
		return fmt.Errorf("failed to marshal token scopes: %w", err)
	}

	query, err := dm.sqlLoader.GetQuery("save_token")
	if err != nil {
		return fmt.Errorf("failed to load save_token query: %w", err)
	}

	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	_, err = dm.Exec(query,
		token.Token, token.UserID, token.TenantID, string(scopesJSON),
		token.ExpiresAt, token.CreatedAt)

	return err
}

// LoadAccessToken loads an access token from the database
func (dm *databaseManager) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	query, err := dm.sqlLoader.GetQuery("load_token")
	if err != nil {
		return nil, fmt.Errorf("failed to load load_token query: %w", err)
	}

	row := dm.QueryRow(query, tokenValue)

	token := &AccessToken{}
	var scopesJSON string

	err = row.Scan(&token.Token, &token.UserID, &token.TenantID, &scopesJSON,
		&token.ExpiresAt, &token.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("access token not found")
		}
		return nil, fmt.Errorf("failed to load access token: %w", err)
	}

	if err := json.Unmarshal([]byte(scopesJSON), &token.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token scopes: %w", err)
	}

	return token, nil
}

// ValidateAccessToken validates an access token and returns it if valid
func (dm *databaseManager) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	query, err := dm.sqlLoader.GetQuery("validate_token")
	if err != nil {
		return nil, fmt.Errorf("failed to load validate_token query: %w", err)
	}

	row := dm.QueryRow(query, tokenValue, time.Now())

	token := &AccessToken{}
	var scopesJSON string

	err = row.Scan(&token.Token, &token.UserID, &token.TenantID, &scopesJSON,
		&token.ExpiresAt, &token.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("access token not found or expired")
		}
		return nil, fmt.Errorf("failed to validate access token: %w", err)
	}

	if err := json.Unmarshal([]byte(scopesJSON), &token.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token scopes: %w", err)
	}

	return token, nil
}

// DeleteAccessToken deletes an access token from the database
func (dm *databaseManager) DeleteAccessToken(tokenValue string) error {
	query, err := dm.sqlLoader.GetQuery("delete_token")
	if err != nil {
		return fmt.Errorf("failed to load delete_token query: %w", err)
	}
	_, err = dm.Exec(query, tokenValue)
	return err
}

// CleanupExpiredTokens removes expired access tokens from the database
func (dm *databaseManager) CleanupExpiredTokens() error {
	query, err := dm.sqlLoader.GetQuery("cleanup_expired_tokens")
	if err != nil {
		return fmt.Errorf("failed to load cleanup_expired_tokens query: %w", err)
	}
	_, err = dm.Exec(query, time.Now())
	return err
}

// SaveTenant saves a tenant to the database
func (dm *databaseManager) SaveTenant(tenant *Tenant) error {
	hostsJSON, err := json.Marshal(tenant.Hosts)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant hosts: %w", err)
	}

	configJSON, err := json.Marshal(tenant.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant config: %w", err)
	}

	query, err := dm.sqlLoader.GetQuery("save_tenant")
	if err != nil {
		return fmt.Errorf("failed to load save_tenant query: %w", err)
	}

	now := time.Now()
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	tenant.UpdatedAt = now

	_, err = dm.Exec(query,
		tenant.ID, tenant.Name, string(hostsJSON), string(configJSON),
		tenant.IsActive, tenant.CreatedAt, tenant.UpdatedAt,
		tenant.MaxUsers, tenant.MaxStorage, tenant.MaxRequests)

	return err
}

// LoadTenant loads a tenant by ID from the database
func (dm *databaseManager) LoadTenant(tenantID string) (*Tenant, error) {
	query, err := dm.sqlLoader.GetQuery("load_tenant")
	if err != nil {
		return nil, fmt.Errorf("failed to load load_tenant query: %w", err)
	}

	row := dm.QueryRow(query, tenantID)

	tenant := &Tenant{}
	var hostsJSON, configJSON string

	err = row.Scan(&tenant.ID, &tenant.Name, &hostsJSON, &configJSON,
		&tenant.IsActive, &tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.MaxUsers, &tenant.MaxStorage, &tenant.MaxRequests)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to load tenant: %w", err)
	}

	if err := json.Unmarshal([]byte(hostsJSON), &tenant.Hosts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant hosts: %w", err)
	}

	if err := json.Unmarshal([]byte(configJSON), &tenant.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant config: %w", err)
	}

	return tenant, nil
}

// LoadTenantByHost loads a tenant by hostname from the database
func (dm *databaseManager) LoadTenantByHost(hostname string) (*Tenant, error) {
	query, err := dm.sqlLoader.GetQuery("load_tenant_by_host")
	if err != nil {
		return nil, fmt.Errorf("failed to load load_tenant_by_host query: %w", err)
	}

	// Pass hostname directly - json_each.value returns raw string values
	row := dm.QueryRow(query, hostname)

	tenant := &Tenant{}
	var hostsJSON, configJSON string

	err = row.Scan(&tenant.ID, &tenant.Name, &hostsJSON, &configJSON,
		&tenant.IsActive, &tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.MaxUsers, &tenant.MaxStorage, &tenant.MaxRequests)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found for host: %s", hostname)
		}
		return nil, fmt.Errorf("failed to load tenant by host: %w", err)
	}

	if err := json.Unmarshal([]byte(hostsJSON), &tenant.Hosts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant hosts: %w", err)
	}

	if err := json.Unmarshal([]byte(configJSON), &tenant.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant config: %w", err)
	}

	return tenant, nil
}

// SaveWorkloadMetrics saves workload metrics to the database
func (dm *databaseManager) SaveWorkloadMetrics(metrics *WorkloadMetrics) error {
	query, err := dm.sqlLoader.GetQuery("save_workload_metrics")
	if err != nil {
		return fmt.Errorf("failed to load save_workload_metrics query: %w", err)
	}

	_, err = dm.Exec(query,
		metrics.Timestamp, metrics.TenantID, metrics.UserID, metrics.RequestID,
		metrics.Duration, metrics.ContextSize, metrics.MemoryUsage, metrics.CPUUsage,
		metrics.Path, metrics.Method, metrics.StatusCode, metrics.ResponseSize, metrics.ErrorMessage)

	return err
}

// GetWorkloadMetrics retrieves workload metrics for a tenant within a time range
func (dm *databaseManager) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	query, err := dm.sqlLoader.GetQuery("get_workload_metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to load get_workload_metrics query: %w", err)
	}

	rows, err := dm.Query(query, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query workload metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*WorkloadMetrics
	for rows.Next() {
		metric := &WorkloadMetrics{}
		err := rows.Scan(&metric.ID, &metric.Timestamp, &metric.TenantID, &metric.UserID, &metric.RequestID,
			&metric.Duration, &metric.ContextSize, &metric.MemoryUsage, &metric.CPUUsage,
			&metric.Path, &metric.Method, &metric.StatusCode, &metric.ResponseSize, &metric.ErrorMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workload metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// CheckRateLimit checks if a rate limit has been exceeded
func (dm *databaseManager) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	windowStart := time.Now().Add(-window)

	query, err := dm.sqlLoader.GetQuery("check_rate_limit")
	if err != nil {
		return false, fmt.Errorf("failed to load check_rate_limit query: %w", err)
	}

	row := dm.QueryRow(query, key, windowStart)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return count >= limit, nil
}

// IncrementRateLimit increments the rate limit counter for a key
func (dm *databaseManager) IncrementRateLimit(key string, window time.Duration) error {
	query, err := dm.sqlLoader.GetQuery("increment_rate_limit")
	if err != nil {
		return fmt.Errorf("failed to load increment_rate_limit query: %w", err)
	}

	_, err = dm.Exec(query, key, time.Now())
	if err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	// Clean up old entries
	cleanupQuery, err := dm.sqlLoader.GetQuery("cleanup_rate_limits")
	if err != nil {
		return fmt.Errorf("failed to load cleanup_rate_limits query: %w", err)
	}

	_, err = dm.Exec(cleanupQuery, time.Now().Add(-window))

	return err
}

// Migration support methods

// Migrate runs database migrations
func (dm *databaseManager) Migrate() error {
	return dm.CreateTables()
}

// CreateTables creates all required database tables
func (dm *databaseManager) CreateTables() error {
	// List of all table creation queries
	tableQueries := []string{
		"create_sessions_table",
		"create_tokens_table",
		"create_tenants_table",
		"create_workload_metrics_table",
		"create_rate_limits_table",
		"create_plugins_table",
		"create_plugin_hooks_table",
		"create_plugin_events_table",
		"create_plugin_storage_table",
		"create_plugin_metrics_table",
	}

	// Create each table using SQL loader
	for _, queryName := range tableQueries {
		schema, err := dm.sqlLoader.GetQuery(queryName)
		if err != nil {
			return fmt.Errorf("failed to load query %s: %w", queryName, err)
		}

		if _, err := dm.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table %s: %w", queryName, err)
		}
	}

	// Create indexes
	indexQueries := []string{
		"index_sessions",
		"index_tokens",
		"index_tenants",
		"index_metrics",
		"index_rate_limits",
		"index_plugins",
		"index_plugin_hooks",
		"index_plugin_events",
		"index_plugin_storage",
		"index_plugin_metrics",
	}

	for _, queryName := range indexQueries {
		indexSQL, err := dm.sqlLoader.GetQuery(queryName)
		if err != nil {
			return fmt.Errorf("failed to load query %s: %w", queryName, err)
		}

		if _, err := dm.Exec(indexSQL); err != nil {
			// Log but don't fail if index already exists
			// Some databases may error on duplicate index creation
		}
	}

	return nil
}

// DropTables drops all framework tables
func (dm *databaseManager) DropTables() error {
	tables := []string{
		"plugin_metrics", "plugin_storage", "plugin_events", "plugin_hooks", "plugins",
		"workload_metrics", "rate_limits", "access_tokens", "sessions", "tenants",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := dm.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// InitializePluginTables creates plugin system tables
func (dm *databaseManager) InitializePluginTables() error {
	// List of plugin table creation queries
	pluginTableQueries := []string{
		"create_plugins_table",
		"create_plugin_hooks_table",
		"create_plugin_events_table",
		"create_plugin_storage_table",
		"create_plugin_metrics_table",
	}

	// Create tables using SQL loader
	for _, queryName := range pluginTableQueries {
		schema, err := dm.sqlLoader.GetQuery(queryName)
		if err != nil {
			return fmt.Errorf("failed to load query %s: %w", queryName, err)
		}

		if _, err := dm.Exec(schema); err != nil {
			return fmt.Errorf("failed to create plugin table %s: %w", queryName, err)
		}
	}

	// Create indexes using SQL loader
	pluginIndexQueries := []string{
		"index_plugins",
		"index_plugin_hooks",
		"index_plugin_events",
		"index_plugin_storage",
		"index_plugin_metrics",
	}

	for _, queryName := range pluginIndexQueries {
		indexSQL, err := dm.sqlLoader.GetQuery(queryName)
		if err != nil {
			return fmt.Errorf("failed to load query %s: %w", queryName, err)
		}

		if _, err := dm.Exec(indexSQL); err != nil {
			// Log but don't fail if index already exists
			// Some databases may error on duplicate index creation
		}
	}

	return nil
}

// GetQuery retrieves a SQL query by name from the SQL loader
func (dm *databaseManager) GetQuery(name string) (string, error) {
	if dm.sqlLoader == nil {
		return "", fmt.Errorf("SQL loader not initialized")
	}
	return dm.sqlLoader.GetQuery(name)
}
