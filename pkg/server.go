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
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int

	// Protocol configuration
	EnableHTTP1 bool
	EnableHTTP2 bool
	EnableQUIC  bool

	// TLS configuration
	TLSConfig *tls.Config

	// Connection limits
	MaxConnections int
	MaxRequestSize int64

	// Graceful shutdown
	ShutdownTimeout time.Duration

	// Performance tuning
	ReadBufferSize  int
	WriteBufferSize int

	// Multi-tenancy
	HostConfigs map[string]*HostConfig

	// Monitoring
	EnableMetrics bool
	MetricsPath   string
	EnablePprof   bool
	PprofPath     string

	// Platform-specific options
	ListenerConfig *ListenerConfig
	EnablePrefork  bool
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
