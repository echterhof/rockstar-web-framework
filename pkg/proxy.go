package pkg

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ProxyManager defines the forward proxy interface for load balancing and request distribution
type ProxyManager interface {
	// Backend management
	AddBackend(backend *Backend) error
	RemoveBackend(backendID string) error
	GetBackend(backendID string) (*Backend, error)
	ListBackends() []*Backend

	// Request forwarding
	Forward(ctx Context, request *Request) (*Response, error)
	ForwardHTTP(ctx context.Context, req *http.Request) (*http.Response, error)

	// Load balancing
	SetLoadBalancer(lb LoadBalancer) error
	GetLoadBalancer() LoadBalancer

	// Circuit breaker
	SetCircuitBreaker(cb CircuitBreaker) error
	GetCircuitBreaker() CircuitBreaker

	// Connection pooling
	SetConnectionPool(pool ConnectionPool) error
	GetConnectionPool() ConnectionPool

	// Health checking
	HealthCheck() error
	GetHealthStatus() map[string]*BackendHealth

	// Metrics
	GetMetrics() *ProxyMetrics
	ResetMetrics()
}

// Backend represents a backend server for proxy forwarding
type Backend struct {
	ID       string   `json:"id"`
	URL      *url.URL `json:"url"`
	Weight   int      `json:"weight"`    // Weight for weighted round-robin
	IsActive bool     `json:"is_active"` // Whether backend is currently active

	// Health check configuration
	HealthCheckPath     string        `json:"health_check_path"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`

	// Circuit breaker state
	FailureCount    int       `json:"failure_count"`
	LastFailureTime time.Time `json:"last_failure_time"`

	// Metadata
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata"`
}

// BackendHealth represents the health status of a backend
type BackendHealth struct {
	BackendID        string        `json:"backend_id"`
	IsHealthy        bool          `json:"is_healthy"`
	LastCheck        time.Time     `json:"last_check"`
	LastSuccess      time.Time     `json:"last_success"`
	LastFailure      time.Time     `json:"last_failure"`
	ConsecutiveFails int           `json:"consecutive_fails"`
	ResponseTime     time.Duration `json:"response_time"`
	ErrorMessage     string        `json:"error_message,omitempty"`
}

// LoadBalancer defines the interface for load balancing strategies
type LoadBalancer interface {
	// Select a backend for the next request
	SelectBackend(backends []*Backend) (*Backend, error)

	// Update backend state after request
	UpdateBackend(backendID string, success bool, responseTime time.Duration)

	// Get load balancer type
	Type() string
}

// CircuitBreaker defines the interface for circuit breaker pattern
type CircuitBreaker interface {
	// Check if circuit is open (failing)
	IsOpen(backendID string) bool

	// Record a success
	RecordSuccess(backendID string)

	// Record a failure
	RecordFailure(backendID string)

	// Reset circuit breaker state
	Reset(backendID string)

	// Get circuit state
	GetState(backendID string) CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"    // Normal operation
	CircuitStateOpen     CircuitState = "open"      // Failing, rejecting requests
	CircuitStateHalfOpen CircuitState = "half_open" // Testing if backend recovered
)

// ConnectionPool defines the interface for connection pooling
type ConnectionPool interface {
	// Get a connection for a backend
	GetConnection(backendID string) (*http.Client, error)

	// Release a connection back to the pool
	ReleaseConnection(backendID string, client *http.Client)

	// Close all connections
	Close() error

	// Get pool statistics
	Stats() *PoolStats
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	TotalConnections  int            `json:"total_connections"`
	ActiveConnections int            `json:"active_connections"`
	IdleConnections   int            `json:"idle_connections"`
	PerBackend        map[string]int `json:"per_backend"`
}

// ProxyMetrics represents proxy performance metrics
type ProxyMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	TotalResponseTime   time.Duration `json:"total_response_time"`
	AverageResponseTime time.Duration `json:"average_response_time"`

	// Per-backend metrics
	BackendMetrics map[string]*BackendMetrics `json:"backend_metrics"`

	// Cache metrics
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`

	// Retry metrics
	TotalRetries int64 `json:"total_retries"`

	mu sync.RWMutex
}

// BackendMetrics represents metrics for a specific backend
type BackendMetrics struct {
	Requests            int64         `json:"requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	TotalResponseTime   time.Duration `json:"total_response_time"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequestTime     time.Time     `json:"last_request_time"`
}

// ProxyConfig represents proxy configuration
type ProxyConfig struct {
	// Load balancing strategy
	LoadBalancerType string `json:"load_balancer_type"` // round_robin, weighted_round_robin, least_connections

	// Circuit breaker configuration
	CircuitBreakerEnabled      bool          `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold    int           `json:"circuit_breaker_threshold"`     // Failures before opening
	CircuitBreakerTimeout      time.Duration `json:"circuit_breaker_timeout"`       // Time before trying half-open
	CircuitBreakerResetTimeout time.Duration `json:"circuit_breaker_reset_timeout"` // Time before resetting counters

	// Connection pool configuration
	MaxConnectionsPerBackend int           `json:"max_connections_per_backend"`
	ConnectionTimeout        time.Duration `json:"connection_timeout"`
	IdleConnTimeout          time.Duration `json:"idle_conn_timeout"`

	// Retry configuration
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	RetryBackoff bool          `json:"retry_backoff"` // Exponential backoff

	// Cache configuration
	CacheEnabled bool          `json:"cache_enabled"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	CacheMaxSize int64         `json:"cache_max_size"`

	// Health check configuration
	HealthCheckEnabled  bool          `json:"health_check_enabled"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	HealthCheckPath     string        `json:"health_check_path"`

	// Request timeout
	RequestTimeout time.Duration `json:"request_timeout"`

	// DNS configuration
	DNSCacheEnabled bool          `json:"dns_cache_enabled"`
	DNSCacheTTL     time.Duration `json:"dns_cache_ttl"`
}

// DefaultProxyConfig returns default proxy configuration
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		LoadBalancerType:           "round_robin",
		CircuitBreakerEnabled:      true,
		CircuitBreakerThreshold:    5,
		CircuitBreakerTimeout:      30 * time.Second,
		CircuitBreakerResetTimeout: 60 * time.Second,
		MaxConnectionsPerBackend:   100,
		ConnectionTimeout:          10 * time.Second,
		IdleConnTimeout:            90 * time.Second,
		MaxRetries:                 3,
		RetryDelay:                 100 * time.Millisecond,
		RetryBackoff:               true,
		CacheEnabled:               true,
		CacheTTL:                   5 * time.Minute,
		CacheMaxSize:               100 * 1024 * 1024, // 100MB
		HealthCheckEnabled:         true,
		HealthCheckInterval:        10 * time.Second,
		HealthCheckTimeout:         5 * time.Second,
		HealthCheckPath:            "/health",
		RequestTimeout:             30 * time.Second,
		DNSCacheEnabled:            true,
		DNSCacheTTL:                5 * time.Minute,
	}
}
