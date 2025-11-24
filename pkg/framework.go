package pkg

import (
	"context"
	"fmt"
	"time"
)

// Framework is the main framework struct that wires all components together
type Framework struct {
	// Core components
	serverManager ServerManager
	router        RouterEngine
	security      SecurityManager

	// Data layer
	database DatabaseManager
	cache    CacheManager
	session  SessionManager

	// Configuration and i18n
	config ConfigManager
	i18n   I18nManager

	// Monitoring and metrics
	metrics    MetricsCollector
	monitoring MonitoringManager

	// Proxy
	proxy ProxyManager

	// Middleware
	globalMiddleware []MiddlewareFunc

	// Error handling
	errorHandler func(ctx Context, err error) error

	// Lifecycle hooks
	shutdownHooks []func(ctx context.Context) error
	startupHooks  []func(ctx context.Context) error

	// State
	isRunning bool
}

// FrameworkConfig holds the complete framework configuration
type FrameworkConfig struct {
	// Server configuration
	ServerConfig ServerConfig

	// Database configuration
	DatabaseConfig DatabaseConfig

	// Cache configuration
	CacheConfig CacheConfig

	// Session configuration
	SessionConfig SessionConfig

	// Configuration file paths
	ConfigFiles []string

	// i18n configuration
	I18nConfig I18nConfig

	// Security configuration
	SecurityConfig SecurityConfig

	// Monitoring configuration
	MonitoringConfig MonitoringConfig

	// Proxy configuration
	ProxyConfig ProxyConfig
}

// New creates a new Framework instance with the given configuration
func New(config FrameworkConfig) (*Framework, error) {
	f := &Framework{
		globalMiddleware: make([]MiddlewareFunc, 0),
		shutdownHooks:    make([]func(ctx context.Context) error, 0),
		startupHooks:     make([]func(ctx context.Context) error, 0),
	}

	// Initialize configuration manager
	f.config = NewConfigManager()

	// Load configuration files if specified
	for _, configFile := range config.ConfigFiles {
		if err := f.config.Load(configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	}

	// Initialize i18n manager
	i18nMgr, err := NewI18nManager(config.I18nConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create i18n manager: %w", err)
	}
	f.i18n = i18nMgr

	// Initialize database manager
	dbMgr := NewDatabaseManager()
	if err := dbMgr.Connect(config.DatabaseConfig); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	f.database = dbMgr

	// Initialize cache manager with configuration
	f.cache = NewCacheManager(config.CacheConfig)

	// Initialize session manager
	sessionMgr, err := NewSessionManager(&config.SessionConfig, f.database, f.cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}
	f.session = sessionMgr

	// Initialize security manager
	securityMgr, err := NewSecurityManager(f.database, config.SecurityConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}
	f.security = securityMgr

	// Initialize metrics collector
	metricsMgr := NewMetricsCollector(f.database)
	f.metrics = metricsMgr

	// Initialize monitoring manager
	logger := NewLogger(nil)
	monitoringMgr := NewMonitoringManager(config.MonitoringConfig, metricsMgr, f.database, logger)
	f.monitoring = monitoringMgr

	// Initialize proxy manager
	proxyMgr := NewProxyManager(&config.ProxyConfig, f.cache)
	f.proxy = proxyMgr

	// Initialize router
	router := NewRouter()
	f.router = router

	// Initialize server manager
	serverMgr := NewServerManager()
	f.serverManager = serverMgr

	return f, nil
}

// Router returns the framework's router for route registration
func (f *Framework) Router() RouterEngine {
	return f.router
}

// Use adds global middleware to the framework
func (f *Framework) Use(middleware ...MiddlewareFunc) {
	f.globalMiddleware = append(f.globalMiddleware, middleware...)
}

// SetErrorHandler sets a custom error handler
func (f *Framework) SetErrorHandler(handler func(ctx Context, err error) error) {
	f.errorHandler = handler
}

// RegisterShutdownHook registers a function to be called during graceful shutdown
func (f *Framework) RegisterShutdownHook(hook func(ctx context.Context) error) {
	f.shutdownHooks = append(f.shutdownHooks, hook)
}

// RegisterStartupHook registers a function to be called during startup
func (f *Framework) RegisterStartupHook(hook func(ctx context.Context) error) {
	f.startupHooks = append(f.startupHooks, hook)
}

// Listen starts the framework server on the specified address
func (f *Framework) Listen(addr string) error {
	return f.ListenWithConfig(addr, ServerConfig{})
}

// ListenTLS starts the framework server with TLS on the specified address
func (f *Framework) ListenTLS(addr, certFile, keyFile string) error {
	config := ServerConfig{}
	server := f.serverManager.NewServer(config)

	// Set router and middleware
	server.SetRouter(f.router)
	server.SetMiddleware(f.globalMiddleware...)

	if f.errorHandler != nil {
		server.SetErrorHandler(f.errorHandler)
	}

	// Register shutdown hooks
	for _, hook := range f.shutdownHooks {
		server.RegisterShutdownHook(hook)
	}

	// Run startup hooks
	ctx := context.Background()
	for _, hook := range f.startupHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("startup hook failed: %w", err)
		}
	}

	f.isRunning = true
	return server.ListenTLS(addr, certFile, keyFile)
}

// ListenQUIC starts the framework server with QUIC on the specified address
func (f *Framework) ListenQUIC(addr, certFile, keyFile string) error {
	config := ServerConfig{
		EnableQUIC: true,
	}
	server := f.serverManager.NewServer(config)

	// Set router and middleware
	server.SetRouter(f.router)
	server.SetMiddleware(f.globalMiddleware...)

	if f.errorHandler != nil {
		server.SetErrorHandler(f.errorHandler)
	}

	// Register shutdown hooks
	for _, hook := range f.shutdownHooks {
		server.RegisterShutdownHook(hook)
	}

	// Run startup hooks
	ctx := context.Background()
	for _, hook := range f.startupHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("startup hook failed: %w", err)
		}
	}

	f.isRunning = true
	return server.ListenQUIC(addr, certFile, keyFile)
}

// ListenWithConfig starts the framework server with custom configuration
func (f *Framework) ListenWithConfig(addr string, config ServerConfig) error {
	server := f.serverManager.NewServer(config)

	// Set router and middleware
	server.SetRouter(f.router)
	server.SetMiddleware(f.globalMiddleware...)

	if f.errorHandler != nil {
		server.SetErrorHandler(f.errorHandler)
	}

	// Set managers for context creation
	logger := NewLogger(nil)
	if httpServer, ok := server.(*httpServer); ok {
		httpServer.SetManagers(logger, f.metrics, f.session, f.database, f.cache, f.config, f.i18n, f.security)
	}

	// Register shutdown hooks
	for _, hook := range f.shutdownHooks {
		server.RegisterShutdownHook(hook)
	}

	// Run startup hooks
	ctx := context.Background()
	for _, hook := range f.startupHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("startup hook failed: %w", err)
		}
	}

	f.isRunning = true
	return server.Listen(addr)
}

// Shutdown gracefully shuts down the framework
func (f *Framework) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Run shutdown hooks
	for _, hook := range f.shutdownHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("shutdown hook failed: %w", err)
		}
	}

	// Shutdown all servers if running
	if f.isRunning {
		if err := f.serverManager.GracefulShutdown(timeout); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	// Close database connections
	if f.database != nil {
		if err := f.database.Close(); err != nil {
			return fmt.Errorf("database close failed: %w", err)
		}
	}

	// Cleanup cache
	if f.cache != nil {
		f.cache.Clear()
	}

	// Cleanup expired sessions
	if f.session != nil {
		f.session.CleanupExpired()
	}

	f.isRunning = false
	return nil
}

// IsRunning returns whether the framework is currently running
func (f *Framework) IsRunning() bool {
	return f.isRunning
}

// ServerManager returns the framework's server manager
func (f *Framework) ServerManager() ServerManager {
	return f.serverManager
}

// Database returns the framework's database manager
func (f *Framework) Database() DatabaseManager {
	return f.database
}

// Cache returns the framework's cache manager
func (f *Framework) Cache() CacheManager {
	return f.cache
}

// Session returns the framework's session manager
func (f *Framework) Session() SessionManager {
	return f.session
}

// Security returns the framework's security manager
func (f *Framework) Security() SecurityManager {
	return f.security
}

// Config returns the framework's configuration manager
func (f *Framework) Config() ConfigManager {
	return f.config
}

// I18n returns the framework's i18n manager
func (f *Framework) I18n() I18nManager {
	return f.i18n
}

// Metrics returns the framework's metrics collector
func (f *Framework) Metrics() MetricsCollector {
	return f.metrics
}

// Monitoring returns the framework's monitoring manager
func (f *Framework) Monitoring() MonitoringManager {
	return f.monitoring
}

// Proxy returns the framework's proxy manager
func (f *Framework) Proxy() ProxyManager {
	return f.proxy
}
