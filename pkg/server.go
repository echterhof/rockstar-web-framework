package pkg

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

// Server represents an HTTP server instance
type Server interface {
	// Server lifecycle
	Listen(addr string) error
	ListenTLS(addr, certFile, keyFile string) error
	ListenQUIC(addr, certFile, keyFile string) error
	Shutdown(ctx context.Context) error
	Close() error

	// Protocol support
	EnableHTTP1() Server
	EnableHTTP2() Server
	EnableQUIC() Server

	// Configuration
	SetConfig(config ServerConfig) Server
	SetMiddleware(middleware ...MiddlewareFunc) Server
	SetRouter(router RouterEngine) Server
	SetErrorHandler(handler func(ctx Context, err error) error) Server

	// Server information
	Addr() string
	IsRunning() bool
	Protocol() string

	// Graceful shutdown
	RegisterShutdownHook(hook func(ctx context.Context) error)
	GracefulShutdown(timeout time.Duration) error
}

// ServerConfig holds server configuration
type ServerConfig struct {
	// Network configuration
	// ReadTimeout is the maximum duration for reading the entire request.
	// Default: 30 seconds
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// Default: 30 seconds
	WriteTimeout time.Duration

	// HSTS (HTTP Strict Transport Security) configuration
	EnableHSTS            bool          // Enable HSTS header
	HSTSMaxAge            time.Duration // HSTS max-age in seconds (default: 1 year)
	HSTSIncludeSubdomains bool          // Include subdomains in HSTS
	HSTSPreload           bool          // Enable HSTS preload

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// Default: 120 seconds
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header.
	// Default: 1048576 (1 MB)
	MaxHeaderBytes int

	// Protocol configuration
	// EnableHTTP1 enables HTTP/1.1 protocol support.
	// Default: false
	EnableHTTP1 bool

	// EnableHTTP2 enables HTTP/2 protocol support.
	// Default: false
	EnableHTTP2 bool

	// EnableQUIC enables QUIC protocol support.
	// Default: false
	EnableQUIC bool

	// TLS configuration
	// TLSConfig provides TLS configuration for secure connections.
	// Default: nil
	TLSConfig *tls.Config

	// Connection limits
	// MaxConnections is the maximum number of concurrent connections.
	// Default: 10000
	MaxConnections int

	// MaxRequestSize is the maximum size of a request body in bytes.
	// Default: 10485760 (10 MB)
	MaxRequestSize int64

	// Graceful shutdown
	// ShutdownTimeout is the maximum duration to wait for graceful shutdown.
	// Default: 30 seconds
	ShutdownTimeout time.Duration

	// Performance tuning
	// ReadBufferSize is the size of the read buffer in bytes.
	// Default: 4096
	ReadBufferSize int

	// WriteBufferSize is the size of the write buffer in bytes.
	// Default: 4096
	WriteBufferSize int

	// Multi-tenancy
	// HostConfigs provides host-specific configuration for multi-tenancy.
	// Default: nil
	HostConfigs map[string]*HostConfig

	// Monitoring
	// EnableMetrics enables the metrics endpoint.
	// Default: false
	EnableMetrics bool

	// MetricsPath is the path for the metrics endpoint.
	// Default: ""
	MetricsPath string

	// EnablePprof enables pprof profiling endpoints.
	// Default: false
	EnablePprof bool

	// PprofPath is the path for pprof endpoints.
	// Default: ""
	PprofPath string

	// Platform-specific options
	// ListenerConfig provides platform-specific listener configuration.
	// Default: nil
	ListenerConfig *ListenerConfig

	// EnablePrefork enables prefork mode for multi-process servers.
	// Default: false
	EnablePrefork bool

	// PreforkWorkers is the number of worker processes in prefork mode.
	// Default: 0
	PreforkWorkers int
}

// HostConfig holds host-specific configuration
type HostConfig struct {
	Hostname       string
	TenantID       string
	VirtualFS      VirtualFS
	Middleware     []MiddlewareFunc
	RateLimits     *RateLimitConfig
	SecurityConfig *ServerSecurityConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	BurstSize         int
	Storage           string // "memory", "database", "redis"
}

// ServerSecurityConfig holds security configuration for the server
type ServerSecurityConfig struct {
	EnableXFrameOptions bool
	XFrameOptions       string
	EnableCORS          bool
	CORSConfig          *CORSConfig
	EnableCSRF          bool
	EnableXSS           bool
	MaxRequestSize      int64
	RequestTimeout      time.Duration
}

// ServerManager manages multiple server instances
type ServerManager interface {
	// Server creation
	NewServer(config ServerConfig) Server

	// Multi-server management
	StartAll() error
	StopAll() error
	GracefulShutdown(timeout time.Duration) error

	// Host and tenant management
	RegisterHost(hostname string, config HostConfig) error
	RegisterTenant(tenantID string, hosts []string) error
	UnregisterHost(hostname string) error
	UnregisterTenant(tenantID string) error
	GetHostConfig(hostname string) (*HostConfig, bool)
	GetTenantInfo(tenantID string) (*TenantInfo, bool)

	// Server information
	GetServer(addr string) (Server, bool)
	GetServers() []Server
	GetServerCount() int

	// Health checks
	HealthCheck() error
	IsHealthy() bool
}

// Listener represents a network listener
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
}

// TenantInfo holds tenant information
type TenantInfo struct {
	TenantID string
	Hosts    []string
	Config   map[string]interface{}
}
