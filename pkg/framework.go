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

	// Plugin system
	pluginManager PluginManager

	// File and network
	fileManager   FileManager
	networkClient NetworkClient

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

	// Plugin configuration
	PluginConfigPath string
	EnablePlugins    bool

	// File system configuration
	FileSystemRoot string // Root directory for file operations (defaults to current directory)
}

// New creates a new Framework instance with the given configuration
func New(config FrameworkConfig) (*Framework, error) {
	f := &Framework{
		globalMiddleware: make([]MiddlewareFunc, 0),
		shutdownHooks:    make([]func(ctx context.Context) error, 0),
		startupHooks:     make([]func(ctx context.Context) error, 0),
	}

	// Apply defaults to all configuration structures before using them
	config.ServerConfig.ApplyDefaults()
	config.DatabaseConfig.ApplyDefaults()
	config.CacheConfig.ApplyDefaults()
	config.SessionConfig.ApplyDefaults()
	config.MonitoringConfig.ApplyDefaults()

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

	// Initialize database manager conditionally
	var dbMgr DatabaseManager
	if isDatabaseConfigured(config.DatabaseConfig) {
		dbMgr = NewDatabaseManager()
		if err := dbMgr.Connect(config.DatabaseConfig); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		fmt.Println("INFO: Framework initialized with database connection")
	} else {
		dbMgr = NewNoopDatabaseManager()
		fmt.Println("INFO: Framework initialized without database. Database-dependent features will use in-memory storage.")
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

	// Initialize file manager
	fsRoot := config.FileSystemRoot
	if fsRoot == "" {
		fsRoot = "."
	}
	vfs := NewOSFileSystem(fsRoot)
	f.fileManager = NewFileManager(vfs)

	// Initialize network client
	f.networkClient = NewNetworkClient()

	// Initialize plugin system if enabled
	if config.EnablePlugins {
		// Create plugin system components
		logger := NewLogger(nil)
		pluginRegistry := NewPluginRegistry()
		hookSystem := NewHookSystem(logger, metricsMgr)
		eventBus := NewEventBus(logger)
		permissionChecker := NewPermissionChecker(logger)

		// Create plugin manager
		pluginMgr := NewPluginManager(
			pluginRegistry,
			hookSystem,
			eventBus,
			permissionChecker,
			logger,
			metricsMgr,
			router,
			dbMgr,
			f.cache,
			f.config,
			f.fileManager,
			f.networkClient,
		)
		f.pluginManager = pluginMgr

		// Discover and initialize plugins
		if err := pluginMgr.DiscoverPlugins(); err != nil {
			return nil, fmt.Errorf("failed to discover plugins: %w", err)
		}

		// Resolve dependencies and initialize plugins
		if err := pluginMgr.InitializeAll(); err != nil {
			return nil, fmt.Errorf("failed to initialize plugins: %w", err)
		}
	}

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

// executeShutdownHooks executes all shutdown hooks including plugin hooks
func (f *Framework) executeShutdownHooks(ctx context.Context) error {
	// Execute plugin shutdown hooks first if plugin system is enabled
	if f.pluginManager != nil {
		// Execute plugin shutdown hooks through the hook system
		if pm, ok := f.pluginManager.(*pluginManagerImpl); ok {
			if pm.hookSystem != nil {
				// Create a minimal context for hook execution
				hookCtx := &shutdownHookContext{ctx: ctx}
				if err := pm.hookSystem.ExecuteHooks(HookTypeShutdown, hookCtx); err != nil {
					// Log error but continue with shutdown
					fmt.Printf("plugin shutdown hooks failed: %v\n", err)
				}
			}
		}

		// Stop all plugins
		if err := f.pluginManager.StopAll(); err != nil {
			// Log error but continue with shutdown
			fmt.Printf("failed to stop plugins: %v\n", err)
		}
	}

	// Execute framework shutdown hooks
	for _, hook := range f.shutdownHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("shutdown hook failed: %w", err)
		}
	}

	return nil
}

// RegisterStartupHook registers a function to be called during startup
func (f *Framework) RegisterStartupHook(hook func(ctx context.Context) error) {
	f.startupHooks = append(f.startupHooks, hook)
}

// executeStartupHooks executes all startup hooks including plugin hooks
func (f *Framework) executeStartupHooks(ctx context.Context) error {
	// Execute framework startup hooks first
	for _, hook := range f.startupHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("startup hook failed: %w", err)
		}
	}

	// Execute plugin startup hooks if plugin system is enabled
	if f.pluginManager != nil {
		// Start all plugins
		if err := f.pluginManager.StartAll(); err != nil {
			return fmt.Errorf("failed to start plugins: %w", err)
		}

		// Execute plugin startup hooks through the hook system
		if pm, ok := f.pluginManager.(*pluginManagerImpl); ok {
			if pm.hookSystem != nil {
				// Create a minimal context for hook execution
				hookCtx := &startupHookContext{ctx: ctx}
				if err := pm.hookSystem.ExecuteHooks(HookTypeStartup, hookCtx); err != nil {
					return fmt.Errorf("plugin startup hooks failed: %w", err)
				}
			}
		}
	}

	return nil
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
	if err := f.executeStartupHooks(ctx); err != nil {
		return err
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
	if err := f.executeStartupHooks(ctx); err != nil {
		return err
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

		// Set hook system if plugin manager is available
		if f.pluginManager != nil {
			if pm, ok := f.pluginManager.(*pluginManagerImpl); ok {
				if pm.hookSystem != nil {
					httpServer.SetHookSystem(pm.hookSystem)
				}
			}
		}
	}

	// Register shutdown hooks
	for _, hook := range f.shutdownHooks {
		server.RegisterShutdownHook(hook)
	}

	// Run startup hooks
	ctx := context.Background()
	if err := f.executeStartupHooks(ctx); err != nil {
		return err
	}

	f.isRunning = true
	return server.Listen(addr)
}

// Shutdown gracefully shuts down the framework
func (f *Framework) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Run shutdown hooks (including plugin hooks)
	if err := f.executeShutdownHooks(ctx); err != nil {
		return fmt.Errorf("shutdown hooks failed: %w", err)
	}

	// Shutdown all servers if running
	if f.isRunning {
		if err := f.serverManager.GracefulShutdown(timeout); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	// Close database connections only if connected
	if f.database != nil && f.database.IsConnected() {
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

// PluginManager returns the framework's plugin manager
func (f *Framework) PluginManager() PluginManager {
	return f.pluginManager
}

// FileManager returns the framework's file manager
func (f *Framework) FileManager() FileManager {
	return f.fileManager
}

// NetworkClient returns the framework's network client
func (f *Framework) NetworkClient() NetworkClient {
	return f.networkClient
}

// LoadPlugin is deprecated for compile-time plugins
// Plugins are now discovered automatically at startup
// This method is kept for backward compatibility but does nothing
func (f *Framework) LoadPlugin(path string) error {
	// Check if plugin system is enabled
	if f.pluginManager == nil {
		return fmt.Errorf("plugin system is not enabled")
	}

	// With compile-time plugins, this is a no-op
	// Plugins are registered via init() functions
	return nil
}

// startupHookContext is a minimal context implementation for startup hooks
type startupHookContext struct {
	ctx context.Context
}

func (c *startupHookContext) Context() context.Context {
	return c.ctx
}

func (c *startupHookContext) Request() *Request {
	return nil
}

func (c *startupHookContext) Response() ResponseWriter {
	return nil
}

func (c *startupHookContext) Params() map[string]string {
	return nil
}

func (c *startupHookContext) Param(name string) string {
	return ""
}

func (c *startupHookContext) Query() map[string]string {
	return nil
}

func (c *startupHookContext) Headers() map[string]string {
	return nil
}

func (c *startupHookContext) Body() []byte {
	return nil
}

func (c *startupHookContext) Session() SessionManager {
	return nil
}

func (c *startupHookContext) User() *User {
	return nil
}

func (c *startupHookContext) Tenant() *Tenant {
	return nil
}

func (c *startupHookContext) DB() DatabaseManager {
	return nil
}

func (c *startupHookContext) Cache() CacheManager {
	return nil
}

func (c *startupHookContext) Config() ConfigManager {
	return nil
}

func (c *startupHookContext) I18n() I18nManager {
	return nil
}

func (c *startupHookContext) Files() FileManager {
	return nil
}

func (c *startupHookContext) Logger() Logger {
	return nil
}

func (c *startupHookContext) Metrics() MetricsCollector {
	return nil
}

func (c *startupHookContext) WithTimeout(timeout time.Duration) Context {
	return c
}

func (c *startupHookContext) WithCancel() (Context, context.CancelFunc) {
	return c, func() {}
}

func (c *startupHookContext) JSON(statusCode int, data interface{}) error {
	return nil
}

func (c *startupHookContext) XML(statusCode int, data interface{}) error {
	return nil
}

func (c *startupHookContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}

func (c *startupHookContext) String(statusCode int, message string) error {
	return nil
}

func (c *startupHookContext) Redirect(statusCode int, url string) error {
	return nil
}

func (c *startupHookContext) SetCookie(cookie *Cookie) error {
	return nil
}

func (c *startupHookContext) GetCookie(name string) (*Cookie, error) {
	return nil, nil
}

func (c *startupHookContext) SetHeader(key, value string) {}

func (c *startupHookContext) GetHeader(key string) string {
	return ""
}

func (c *startupHookContext) FormValue(key string) string {
	return ""
}

func (c *startupHookContext) FormFile(key string) (*FormFile, error) {
	return nil, nil
}

func (c *startupHookContext) IsAuthenticated() bool {
	return false
}

func (c *startupHookContext) IsAuthorized(resource, action string) bool {
	return false
}

func (c *startupHookContext) Set(key string, value interface{}) {
	// No-op for startup hooks
}

func (c *startupHookContext) Get(key string) (interface{}, bool) {
	return nil, false
}

// shutdownHookContext is a minimal context implementation for shutdown hooks
type shutdownHookContext struct {
	ctx context.Context
}

func (c *shutdownHookContext) Context() context.Context {
	return c.ctx
}

func (c *shutdownHookContext) Request() *Request {
	return nil
}

func (c *shutdownHookContext) Response() ResponseWriter {
	return nil
}

func (c *shutdownHookContext) Params() map[string]string {
	return nil
}

func (c *shutdownHookContext) Param(name string) string {
	return ""
}

func (c *shutdownHookContext) Query() map[string]string {
	return nil
}

func (c *shutdownHookContext) Headers() map[string]string {
	return nil
}

func (c *shutdownHookContext) Body() []byte {
	return nil
}

func (c *shutdownHookContext) Session() SessionManager {
	return nil
}

func (c *shutdownHookContext) User() *User {
	return nil
}

func (c *shutdownHookContext) Tenant() *Tenant {
	return nil
}

func (c *shutdownHookContext) DB() DatabaseManager {
	return nil
}

func (c *shutdownHookContext) Cache() CacheManager {
	return nil
}

func (c *shutdownHookContext) Config() ConfigManager {
	return nil
}

func (c *shutdownHookContext) I18n() I18nManager {
	return nil
}

func (c *shutdownHookContext) Files() FileManager {
	return nil
}

func (c *shutdownHookContext) Logger() Logger {
	return nil
}

func (c *shutdownHookContext) Metrics() MetricsCollector {
	return nil
}

func (c *shutdownHookContext) WithTimeout(timeout time.Duration) Context {
	return c
}

func (c *shutdownHookContext) WithCancel() (Context, context.CancelFunc) {
	return c, func() {}
}

func (c *shutdownHookContext) JSON(statusCode int, data interface{}) error {
	return nil
}

func (c *shutdownHookContext) XML(statusCode int, data interface{}) error {
	return nil
}

func (c *shutdownHookContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}

func (c *shutdownHookContext) String(statusCode int, message string) error {
	return nil
}

func (c *shutdownHookContext) Redirect(statusCode int, url string) error {
	return nil
}

func (c *shutdownHookContext) SetCookie(cookie *Cookie) error {
	return nil
}

func (c *shutdownHookContext) GetCookie(name string) (*Cookie, error) {
	return nil, nil
}

func (c *shutdownHookContext) SetHeader(key, value string) {}

func (c *shutdownHookContext) GetHeader(key string) string {
	return ""
}

func (c *shutdownHookContext) FormValue(key string) string {
	return ""
}

func (c *shutdownHookContext) FormFile(key string) (*FormFile, error) {
	return nil, nil
}

func (c *shutdownHookContext) IsAuthenticated() bool {
	return false
}

func (c *shutdownHookContext) IsAuthorized(resource, action string) bool {
	return false
}

func (c *shutdownHookContext) Set(key string, value interface{}) {
	// No-op for shutdown hooks
}

func (c *shutdownHookContext) Get(key string) (interface{}, bool) {
	return nil, false
}
