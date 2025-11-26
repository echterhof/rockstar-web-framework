package pkg

import (
	"database/sql"
	"time"
)

// DatabaseManager defines the database interface supporting multiple database engines
type DatabaseManager interface {
	// Connection management
	Connect(config DatabaseConfig) error
	Close() error
	Ping() error
	Stats() DatabaseStats
	IsConnected() bool

	// Query execution
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Prepared statements
	Prepare(query string) (*sql.Stmt, error)

	// Transaction support
	Begin() (Transaction, error)
	BeginTx(opts *sql.TxOptions) (Transaction, error)

	// Framework-specific model operations
	SaveSession(session *Session) error
	LoadSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	CleanupExpiredSessions() error

	SaveAccessToken(token *AccessToken) error
	LoadAccessToken(tokenValue string) (*AccessToken, error)
	ValidateAccessToken(tokenValue string) (*AccessToken, error)
	DeleteAccessToken(tokenValue string) error
	CleanupExpiredTokens() error

	SaveTenant(tenant *Tenant) error
	LoadTenant(tenantID string) (*Tenant, error)
	LoadTenantByHost(hostname string) (*Tenant, error)

	SaveWorkloadMetrics(metrics *WorkloadMetrics) error
	GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)

	// Rate limiting support
	CheckRateLimit(key string, limit int, window time.Duration) (bool, error)
	IncrementRateLimit(key string, window time.Duration) error

	// Migration support
	Migrate() error
	CreateTables() error
	DropTables() error

	// Plugin system support
	InitializePluginTables() error

	// SQL loader support
	GetQuery(name string) (string, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Commit() error
	Rollback() error
}

// DatabaseConfig defines database connection configuration
type DatabaseConfig struct {
	// Driver specifies the database driver (mysql, postgres, mssql, sqlite).
	// Required, no default
	Driver string `json:"driver"`

	// Host is the database server hostname or IP address.
	// Default: "localhost"
	Host string `json:"host"`

	// Port is the database server port.
	// Default: driver-specific (postgres=5432, mysql=3306, mssql=1433, sqlite=0)
	Port int `json:"port"`

	// Database is the name of the database to connect to.
	// Required, no default
	Database string `json:"database"`

	// Username is the database user for authentication.
	// Required, no default
	Username string `json:"username"`

	// Password is the database password for authentication.
	// Required, no default
	Password string `json:"password"`

	// SSLMode specifies the SSL/TLS mode for the connection.
	// Default: ""
	SSLMode string `json:"ssl_mode"`

	// Charset specifies the character set for the connection.
	// Default: ""
	Charset string `json:"charset"`

	// Timezone specifies the timezone for the connection.
	// Default: ""
	Timezone string `json:"timezone"`

	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	// Default: 5 minutes
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`

	// MaxOpenConns is the maximum number of open connections to the database.
	// Default: 25
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns is the maximum number of idle connections in the pool.
	// Default: 5
	MaxIdleConns int `json:"max_idle_conns"`

	// Options provides driver-specific connection options.
	// Default: nil
	Options map[string]string `json:"options"`
}

// DatabaseStats provides database connection statistics
type DatabaseStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// Session represents a user session stored in database
type Session struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	Data      map[string]interface{} `json:"data" db:"data"`
	ExpiresAt time.Time              `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
}

// Tenant represents a tenant in multi-tenant architecture
type Tenant struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Hosts     []string               `json:"hosts" db:"hosts"`
	Config    map[string]interface{} `json:"config" db:"config"`
	IsActive  bool                   `json:"is_active" db:"is_active"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`

	// Resource limits
	MaxUsers    int   `json:"max_users" db:"max_users"`
	MaxStorage  int64 `json:"max_storage" db:"max_storage"`
	MaxRequests int64 `json:"max_requests" db:"max_requests"`
}

// AccessToken represents an API access token
type AccessToken struct {
	Token     string    `json:"token" db:"token"`
	UserID    string    `json:"user_id" db:"user_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Scopes    []string  `json:"scopes" db:"scopes"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WorkloadMetrics represents performance and usage metrics
type WorkloadMetrics struct {
	ID           int64     `json:"id" db:"id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	RequestID    string    `json:"request_id" db:"request_id"`
	Duration     int64     `json:"duration_ms" db:"duration_ms"`
	ContextSize  int64     `json:"context_size" db:"context_size"`
	MemoryUsage  int64     `json:"memory_usage" db:"memory_usage"`
	CPUUsage     float64   `json:"cpu_usage" db:"cpu_usage"`
	Path         string    `json:"path" db:"path"`
	Method       string    `json:"method" db:"method"`
	StatusCode   int       `json:"status_code" db:"status_code"`
	ResponseSize int64     `json:"response_size" db:"response_size"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
}
