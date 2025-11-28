package pkg

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Mock implementations for testing

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {}
func (m *MockLogger) Info(msg string, fields ...interface{})  {}
func (m *MockLogger) Warn(msg string, fields ...interface{})  {}
func (m *MockLogger) Error(msg string, fields ...interface{}) {}
func (m *MockLogger) WithRequestID(requestID string) Logger   { return m }

type MockMetrics struct{}

func (m *MockMetrics) Start(requestID string) *RequestMetrics { return nil }
func (m *MockMetrics) Record(metrics *RequestMetrics) error   { return nil }
func (m *MockMetrics) RecordRequest(ctx Context, duration time.Duration, statusCode int) error {
	return nil
}
func (m *MockMetrics) RecordError(ctx Context, err error) error { return nil }
func (m *MockMetrics) GetMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *MockMetrics) GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error) {
	return nil, nil
}
func (m *MockMetrics) PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error) {
	return nil, nil
}
func (m *MockMetrics) RecordWorkloadMetrics(metrics *WorkloadMetrics) error { return nil }
func (m *MockMetrics) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *MockMetrics) IncrementCounter(name string, tags map[string]string) error { return nil }
func (m *MockMetrics) IncrementCounterBy(name string, value int64, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) SetGauge(name string, value float64, tags map[string]string) error { return nil }
func (m *MockMetrics) IncrementGauge(name string, value float64, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) DecrementGauge(name string, value float64, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) RecordHistogram(name string, value float64, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) RecordTiming(name string, duration time.Duration, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) StartTimer(name string, tags map[string]string) Timer { return nil }
func (m *MockMetrics) RecordMemoryUsage(usage int64) error                  { return nil }
func (m *MockMetrics) RecordCPUUsage(usage float64) error                   { return nil }
func (m *MockMetrics) RecordCustomMetric(name string, value interface{}, tags map[string]string) error {
	return nil
}
func (m *MockMetrics) Export() (map[string]interface{}, error) { return nil, nil }
func (m *MockMetrics) ExportPrometheus() ([]byte, error)       { return nil, nil }

type MockRouter struct{}

func (m *MockRouter) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) Group(prefix string, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) Host(hostname string) RouterEngine {
	return m
}
func (m *MockRouter) Static(prefix string, filesystem VirtualFS) RouterEngine {
	return m
}
func (m *MockRouter) StaticFile(path, filepath string) RouterEngine {
	return m
}
func (m *MockRouter) WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) Use(middleware ...MiddlewareFunc) RouterEngine {
	return m
}
func (m *MockRouter) Match(method, path, host string) (*Route, map[string]string, bool) {
	return nil, nil, false
}
func (m *MockRouter) Routes() []*Route {
	return nil
}

type MockDatabase struct {
	initCalled bool
}

func (m *MockDatabase) Connect(config DatabaseConfig) error { return nil }
func (m *MockDatabase) Close() error                        { return nil }
func (m *MockDatabase) Ping() error                         { return nil }
func (m *MockDatabase) Stats() DatabaseStats                { return DatabaseStats{} }
func (m *MockDatabase) IsConnected() bool                   { return true }
func (m *MockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDatabase) QueryRow(query string, args ...interface{}) *sql.Row { return nil }
func (m *MockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *MockDatabase) Prepare(query string) (*sql.Stmt, error)          { return nil, nil }
func (m *MockDatabase) Begin() (Transaction, error)                      { return nil, nil }
func (m *MockDatabase) BeginTx(opts *sql.TxOptions) (Transaction, error) { return nil, nil }
func (m *MockDatabase) SaveSession(session *Session) error               { return nil }
func (m *MockDatabase) LoadSession(sessionID string) (*Session, error)   { return nil, nil }
func (m *MockDatabase) DeleteSession(sessionID string) error             { return nil }
func (m *MockDatabase) CleanupExpiredSessions() error                    { return nil }
func (m *MockDatabase) SaveAccessToken(token *AccessToken) error         { return nil }
func (m *MockDatabase) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, nil
}
func (m *MockDatabase) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, nil
}
func (m *MockDatabase) DeleteAccessToken(tokenValue string) error { return nil }
func (m *MockDatabase) CleanupExpiredTokens() error               { return nil }
func (m *MockDatabase) SaveTenant(tenant *Tenant) error           { return nil }
func (m *MockDatabase) LoadTenant(tenantID string) (*Tenant, error) {
	return nil, nil
}
func (m *MockDatabase) LoadTenantByHost(hostname string) (*Tenant, error) {
	return nil, nil
}
func (m *MockDatabase) SaveWorkloadMetrics(metrics *WorkloadMetrics) error { return nil }
func (m *MockDatabase) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *MockDatabase) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return false, nil
}
func (m *MockDatabase) IncrementRateLimit(key string, window time.Duration) error { return nil }
func (m *MockDatabase) Migrate() error                                            { return nil }
func (m *MockDatabase) CreateTables() error                                       { return nil }
func (m *MockDatabase) DropTables() error                                         { return nil }
func (m *MockDatabase) InitializePluginTables() error {
	m.initCalled = true
	return nil
}
func (m *MockDatabase) GetQuery(name string) (string, error) {
	return "", nil
}

type MockCache struct{}

func (m *MockCache) Get(key string) (interface{}, error) { return nil, nil }
func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) error {
	return nil
}
func (m *MockCache) Delete(key string) error               { return nil }
func (m *MockCache) Exists(key string) bool                { return false }
func (m *MockCache) TTL(key string) (time.Duration, error) { return 0, nil }
func (m *MockCache) SetMultiple(items map[string]interface{}, ttl time.Duration) error {
	return nil
}
func (m *MockCache) GetMultiple(keys []string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockCache) DeleteMultiple(keys []string) error               { return nil }
func (m *MockCache) Increment(key string, delta int64) (int64, error) { return 0, nil }
func (m *MockCache) Decrement(key string, delta int64) (int64, error) { return 0, nil }
func (m *MockCache) Expire(key string, ttl time.Duration) error       { return nil }
func (m *MockCache) Clear() error                                     { return nil }
func (m *MockCache) Invalidate(pattern string) error                  { return nil }
func (m *MockCache) GetRequestCache(requestID string) RequestCache    { return nil }
func (m *MockCache) ClearRequestCache(requestID string) error         { return nil }

type MockConfig struct{}

func (m *MockConfig) Load(configPath string) error         { return nil }
func (m *MockConfig) LoadFromEnv() error                   { return nil }
func (m *MockConfig) Reload() error                        { return nil }
func (m *MockConfig) GetString(key string) string          { return "" }
func (m *MockConfig) GetInt(key string) int                { return 0 }
func (m *MockConfig) GetInt64(key string) int64            { return 0 }
func (m *MockConfig) GetFloat64(key string) float64        { return 0 }
func (m *MockConfig) GetBool(key string) bool              { return false }
func (m *MockConfig) GetDuration(key string) time.Duration { return 0 }
func (m *MockConfig) GetStringSlice(key string) []string   { return nil }
func (m *MockConfig) GetWithDefault(key string, defaultValue interface{}) interface{} {
	return defaultValue
}
func (m *MockConfig) GetStringWithDefault(key, defaultValue string) string { return defaultValue }
func (m *MockConfig) GetIntWithDefault(key string, defaultValue int) int   { return defaultValue }
func (m *MockConfig) GetBoolWithDefault(key string, defaultValue bool) bool {
	return defaultValue
}
func (m *MockConfig) GetEnv() string               { return "" }
func (m *MockConfig) Sub(key string) ConfigManager { return m }
func (m *MockConfig) IsSet(key string) bool        { return false }
func (m *MockConfig) IsProduction() bool           { return false }
func (m *MockConfig) IsDevelopment() bool          { return false }
func (m *MockConfig) IsTest() bool                 { return false }
func (m *MockConfig) Validate() error              { return nil }
func (m *MockConfig) Watch(callback func()) error  { return nil }
func (m *MockConfig) StopWatching() error          { return nil }

type MockFileManager struct{}

func (m *MockFileManager) Read(path string) ([]byte, error)     { return nil, nil }
func (m *MockFileManager) Write(path string, data []byte) error { return nil }
func (m *MockFileManager) Delete(path string) error             { return nil }
func (m *MockFileManager) Exists(path string) bool              { return false }
func (m *MockFileManager) CreateDir(path string) error          { return nil }
func (m *MockFileManager) SaveUploadedFile(ctx Context, filename string, destPath string) error {
	return nil
}

type MockNetworkClient struct{}

func (m *MockNetworkClient) Get(url string, headers map[string]string) ([]byte, error) {
	return nil, nil
}
func (m *MockNetworkClient) Post(url string, body []byte, headers map[string]string) ([]byte, error) {
	return nil, nil
}
func (m *MockNetworkClient) Put(url string, body []byte, headers map[string]string) ([]byte, error) {
	return nil, nil
}
func (m *MockNetworkClient) Delete(url string, headers map[string]string) ([]byte, error) {
	return nil, nil
}
func (m *MockNetworkClient) Do(method, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	return nil, 0, nil
}

// LifecycleTrackingPlugin tracks lifecycle method calls for testing
type LifecycleTrackingPlugin struct {
	name           string
	version        string
	lifecycleSteps []string
	mu             *sync.Mutex
	shouldFail     bool
}

func (p *LifecycleTrackingPlugin) Name() string        { return p.name }
func (p *LifecycleTrackingPlugin) Version() string     { return p.version }
func (p *LifecycleTrackingPlugin) Description() string { return "Lifecycle tracking plugin" }
func (p *LifecycleTrackingPlugin) Author() string      { return "Test" }
func (p *LifecycleTrackingPlugin) Dependencies() []PluginDependency {
	return []PluginDependency{}
}

func (p *LifecycleTrackingPlugin) Initialize(ctx PluginContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lifecycleSteps = append(p.lifecycleSteps, "Initialize")
	if p.shouldFail {
		return fmt.Errorf("intentional initialization failure")
	}
	return nil
}

func (p *LifecycleTrackingPlugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lifecycleSteps = append(p.lifecycleSteps, "Start")
	if p.shouldFail {
		return fmt.Errorf("intentional start failure")
	}
	return nil
}

func (p *LifecycleTrackingPlugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lifecycleSteps = append(p.lifecycleSteps, "Stop")
	return nil
}

func (p *LifecycleTrackingPlugin) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lifecycleSteps = append(p.lifecycleSteps, "Cleanup")
	return nil
}

func (p *LifecycleTrackingPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *LifecycleTrackingPlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}

// TestPlugin is a simple test plugin with configurable behavior
type TestPlugin struct {
	name         string
	version      string
	initPanic    bool
	startPanic   bool
	stopPanic    bool
	cleanupPanic bool
	initError    bool
	startError   bool
	stopError    bool
	cleanupError bool
}

func (p *TestPlugin) Name() string        { return p.name }
func (p *TestPlugin) Version() string     { return p.version }
func (p *TestPlugin) Description() string { return "Test plugin" }
func (p *TestPlugin) Author() string      { return "Test" }

func (p *TestPlugin) Dependencies() []PluginDependency {
	return []PluginDependency{}
}

func (p *TestPlugin) Initialize(ctx PluginContext) error {
	if p.initPanic {
		panic("intentional panic during initialization")
	}
	if p.initError {
		return fmt.Errorf("intentional error during initialization")
	}
	return nil
}

func (p *TestPlugin) Start() error {
	if p.startPanic {
		panic("intentional panic during start")
	}
	if p.startError {
		return fmt.Errorf("intentional error during start")
	}
	return nil
}

func (p *TestPlugin) Stop() error {
	if p.stopPanic {
		panic("intentional panic during stop")
	}
	if p.stopError {
		return fmt.Errorf("intentional error during stop")
	}
	return nil
}

func (p *TestPlugin) Cleanup() error {
	if p.cleanupPanic {
		panic("intentional panic during cleanup")
	}
	if p.cleanupError {
		return fmt.Errorf("intentional error during cleanup")
	}
	return nil
}

func (p *TestPlugin) ConfigSchema() map[string]interface{} {
	return make(map[string]interface{})
}

func (p *TestPlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}

// NewTestLogger creates a logger for testing
func NewTestLogger() Logger {
	return &MockLogger{}
}

// NewTestMetrics creates a metrics collector for testing
func NewTestMetrics() MetricsCollector {
	return &MockMetrics{}
}

// ============================================================================
// Mock PluginContext Implementation
// ============================================================================

// MockPluginContext is a comprehensive mock implementation of PluginContext
// for testing plugins in isolation
type MockPluginContext struct {
	// Framework services
	router     RouterEngine
	logger     Logger
	metrics    MetricsCollector
	database   DatabaseManager
	cache      CacheManager
	config     ConfigManager
	fileSystem FileManager
	network    NetworkClient

	// Plugin-specific data
	pluginConfig  map[string]interface{}
	pluginStorage PluginStorage

	// Systems
	hookSystem         HookSystem
	eventBus           EventBus
	serviceRegistry    ServiceRegistry
	middlewareRegistry MiddlewareRegistry

	// Permissions
	permissions       PluginPermissions
	permissionChecker PermissionChecker

	// Tracking for testing
	RegisteredHooks      []HookRegistration
	PublishedEvents      []EventPublication
	SubscribedEvents     []EventSubscription
	ExportedServices     map[string]interface{}
	ImportedServices     []ServiceImport
	RegisteredMiddleware []MiddlewareRegistration
	mu                   sync.RWMutex
}

// EventPublication tracks published events
type EventPublication struct {
	Event string
	Data  interface{}
}

// EventSubscription tracks event subscriptions
type EventSubscription struct {
	Event   string
	Handler EventHandler
}

// ServiceImport tracks service imports
type ServiceImport struct {
	PluginName  string
	ServiceName string
}

// NewMockPluginContext creates a new mock plugin context with default implementations
func NewMockPluginContext() *MockPluginContext {
	return &MockPluginContext{
		router:               &MockRouter{},
		logger:               &MockLogger{},
		metrics:              &MockMetrics{},
		database:             &MockDatabase{},
		cache:                &MockCache{},
		config:               &MockConfig{},
		fileSystem:           &MockFileManager{},
		network:              &MockNetworkClient{},
		pluginConfig:         make(map[string]interface{}),
		pluginStorage:        NewMockPluginStorage(),
		hookSystem:           NewMockHookSystem(),
		eventBus:             NewMockEventBus(),
		serviceRegistry:      NewServiceRegistry(),
		middlewareRegistry:   NewMiddlewareRegistry(),
		permissions:          PluginPermissions{},
		permissionChecker:    nil,
		RegisteredHooks:      []HookRegistration{},
		PublishedEvents:      []EventPublication{},
		SubscribedEvents:     []EventSubscription{},
		ExportedServices:     make(map[string]interface{}),
		ImportedServices:     []ServiceImport{},
		RegisteredMiddleware: []MiddlewareRegistration{},
	}
}

// NewMockPluginContextWithPermissions creates a mock context with specific permissions
func NewMockPluginContextWithPermissions(permissions PluginPermissions) *MockPluginContext {
	ctx := NewMockPluginContext()
	ctx.permissions = permissions
	checker := NewMockPermissionChecker(permissions)
	if mockChecker, ok := checker.(*MockPermissionChecker); ok {
		ctx.permissionChecker = mockChecker
	}
	return ctx
}

// NewMockPluginContextWithConfig creates a mock context with specific configuration
func NewMockPluginContextWithConfig(config map[string]interface{}) *MockPluginContext {
	ctx := NewMockPluginContext()
	ctx.pluginConfig = config
	return ctx
}

// Router returns the mock router
func (m *MockPluginContext) Router() RouterEngine {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "router"); err != nil {
			return &MockRouter{} // Return no-op router
		}
	}
	return m.router
}

// Logger returns the mock logger
func (m *MockPluginContext) Logger() Logger {
	return m.logger
}

// Metrics returns the mock metrics collector
func (m *MockPluginContext) Metrics() MetricsCollector {
	return m.metrics
}

// Database returns the mock database manager
func (m *MockPluginContext) Database() DatabaseManager {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "database"); err != nil {
			return &MockDatabase{} // Return no-op database
		}
	}
	return m.database
}

// Cache returns the mock cache manager
func (m *MockPluginContext) Cache() CacheManager {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "cache"); err != nil {
			return &MockCache{} // Return no-op cache
		}
	}
	return m.cache
}

// Config returns the mock config manager
func (m *MockPluginContext) Config() ConfigManager {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "config"); err != nil {
			return &MockConfig{} // Return no-op config
		}
	}
	return m.config
}

// FileSystem returns the mock file manager
func (m *MockPluginContext) FileSystem() FileManager {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "filesystem"); err != nil {
			return &MockFileManager{} // Return no-op file manager
		}
	}
	return m.fileSystem
}

// Network returns the mock network client
func (m *MockPluginContext) Network() NetworkClient {
	if m.permissionChecker != nil {
		if err := m.permissionChecker.CheckPermission("test-plugin", "network"); err != nil {
			return &MockNetworkClient{} // Return no-op network client
		}
	}
	return m.network
}

// PluginConfig returns the plugin-specific configuration
func (m *MockPluginContext) PluginConfig() map[string]interface{} {
	return m.pluginConfig
}

// PluginStorage returns the mock plugin storage
func (m *MockPluginContext) PluginStorage() PluginStorage {
	return m.pluginStorage
}

// RegisterHook registers a hook and tracks it for testing
func (m *MockPluginContext) RegisterHook(hookType HookType, priority int, handler HookHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	registration := HookRegistration{
		PluginName: "test-plugin",
		HookType:   hookType,
		Priority:   priority,
		Handler:    handler,
	}
	m.RegisteredHooks = append(m.RegisteredHooks, registration)

	if m.hookSystem != nil {
		return m.hookSystem.RegisterHook("test-plugin", hookType, priority, handler)
	}
	return nil
}

// PublishEvent publishes an event and tracks it for testing
func (m *MockPluginContext) PublishEvent(event string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	publication := EventPublication{
		Event: event,
		Data:  data,
	}
	m.PublishedEvents = append(m.PublishedEvents, publication)

	if m.eventBus != nil {
		return m.eventBus.Publish(event, data)
	}
	return nil
}

// SubscribeEvent subscribes to an event and tracks it for testing
func (m *MockPluginContext) SubscribeEvent(event string, handler EventHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscription := EventSubscription{
		Event:   event,
		Handler: handler,
	}
	m.SubscribedEvents = append(m.SubscribedEvents, subscription)

	if m.eventBus != nil {
		return m.eventBus.Subscribe("test-plugin", event, handler)
	}
	return nil
}

// ExportService exports a service and tracks it for testing
func (m *MockPluginContext) ExportService(name string, service interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExportedServices[name] = service

	if m.serviceRegistry != nil {
		return m.serviceRegistry.Export("test-plugin", name, service)
	}
	return nil
}

// ImportService imports a service and tracks it for testing
func (m *MockPluginContext) ImportService(pluginName, serviceName string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	importRecord := ServiceImport{
		PluginName:  pluginName,
		ServiceName: serviceName,
	}
	m.ImportedServices = append(m.ImportedServices, importRecord)

	if m.serviceRegistry != nil {
		return m.serviceRegistry.Import(pluginName, serviceName)
	}
	return nil, nil
}

// RegisterMiddleware registers middleware and tracks it for testing
func (m *MockPluginContext) RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	registration := MiddlewareRegistration{
		PluginName: "test-plugin",
		Name:       name,
		Handler:    handler,
		Priority:   priority,
		Routes:     routes,
	}
	m.RegisteredMiddleware = append(m.RegisteredMiddleware, registration)

	if m.middlewareRegistry != nil {
		return m.middlewareRegistry.Register("test-plugin", name, handler, priority, routes)
	}
	return nil
}

// UnregisterMiddleware unregisters middleware
func (m *MockPluginContext) UnregisterMiddleware(name string) error {
	if m.middlewareRegistry != nil {
		return m.middlewareRegistry.Unregister("test-plugin", name)
	}
	return nil
}

// SetRouter sets a custom router for testing
func (m *MockPluginContext) SetRouter(router RouterEngine) {
	m.router = router
}

// SetLogger sets a custom logger for testing
func (m *MockPluginContext) SetLogger(logger Logger) {
	m.logger = logger
}

// SetMetrics sets a custom metrics collector for testing
func (m *MockPluginContext) SetMetrics(metrics MetricsCollector) {
	m.metrics = metrics
}

// SetDatabase sets a custom database manager for testing
func (m *MockPluginContext) SetDatabase(database DatabaseManager) {
	m.database = database
}

// SetCache sets a custom cache manager for testing
func (m *MockPluginContext) SetCache(cache CacheManager) {
	m.cache = cache
}

// SetConfig sets a custom config manager for testing
func (m *MockPluginContext) SetConfig(config ConfigManager) {
	m.config = config
}

// SetFileSystem sets a custom file manager for testing
func (m *MockPluginContext) SetFileSystem(fileSystem FileManager) {
	m.fileSystem = fileSystem
}

// SetNetwork sets a custom network client for testing
func (m *MockPluginContext) SetNetwork(network NetworkClient) {
	m.network = network
}

// ============================================================================
// Mock Supporting Systems
// ============================================================================

// MockPluginStorage is a simple in-memory plugin storage for testing
type MockPluginStorage struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewMockPluginStorage creates a new mock plugin storage
func NewMockPluginStorage() *MockPluginStorage {
	return &MockPluginStorage{
		data: make(map[string]interface{}),
	}
}

func (s *MockPluginStorage) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *MockPluginStorage) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

func (s *MockPluginStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *MockPluginStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *MockPluginStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
	return nil
}

// MockHookSystem is a simple hook system for testing
type MockHookSystem struct {
	hooks map[string][]HookRegistration
	mu    sync.RWMutex
}

// NewMockHookSystem creates a new mock hook system
func NewMockHookSystem() *MockHookSystem {
	return &MockHookSystem{
		hooks: make(map[string][]HookRegistration),
	}
}

func (h *MockHookSystem) RegisterHook(pluginName string, hookType HookType, priority int, handler HookHandler) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := string(hookType)
	registration := HookRegistration{
		PluginName: pluginName,
		HookType:   hookType,
		Priority:   priority,
		Handler:    handler,
	}
	h.hooks[key] = append(h.hooks[key], registration)
	return nil
}

func (h *MockHookSystem) UnregisterHook(pluginName string, hookType HookType) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := string(hookType)
	hooks := h.hooks[key]
	filtered := make([]HookRegistration, 0)
	for _, hook := range hooks {
		if hook.PluginName != pluginName {
			filtered = append(filtered, hook)
		}
	}
	h.hooks[key] = filtered
	return nil
}

func (h *MockHookSystem) UnregisterAll(pluginName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key, hooks := range h.hooks {
		filtered := make([]HookRegistration, 0)
		for _, hook := range hooks {
			if hook.PluginName != pluginName {
				filtered = append(filtered, hook)
			}
		}
		h.hooks[key] = filtered
	}
	return nil
}

func (h *MockHookSystem) ExecuteHooks(hookType HookType, ctx Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	key := string(hookType)
	hooks := h.hooks[key]

	// Create a simple hook context
	hookCtx := &mockHookContext{
		ctx:      ctx,
		hookType: hookType,
		data:     make(map[string]interface{}),
	}

	for _, hook := range hooks {
		if hookCtx.IsSkipped() {
			break
		}
		if err := hook.Handler(hookCtx); err != nil {
			return err
		}
	}
	return nil
}

// mockHookContext is a simple implementation of HookContext for testing
type mockHookContext struct {
	ctx        Context
	hookType   HookType
	pluginName string
	data       map[string]interface{}
	skipped    bool
	mu         sync.RWMutex
}

func (h *mockHookContext) Context() Context {
	return h.ctx
}

func (h *mockHookContext) HookType() HookType {
	return h.hookType
}

func (h *mockHookContext) PluginName() string {
	return h.pluginName
}

func (h *mockHookContext) Set(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data[key] = value
}

func (h *mockHookContext) Get(key string) interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.data[key]
}

func (h *mockHookContext) Skip() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.skipped = true
}

func (h *mockHookContext) IsSkipped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.skipped
}

func (h *MockHookSystem) ListHooks(hookType HookType) []HookRegistration {
	h.mu.RLock()
	defer h.mu.RUnlock()

	key := string(hookType)
	return h.hooks[key]
}

// MockEventBus is a simple event bus for testing
type MockEventBus struct {
	subscriptions map[string]map[string]EventHandler // event -> pluginName -> handler
	mu            sync.RWMutex
}

// NewMockEventBus creates a new mock event bus
func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		subscriptions: make(map[string]map[string]EventHandler),
	}
}

func (e *MockEventBus) Publish(event string, data interface{}) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	handlers, exists := e.subscriptions[event]
	if !exists {
		return nil
	}

	evt := Event{
		Name:      event,
		Data:      data,
		Source:    "test",
		Timestamp: time.Now(),
	}

	for _, handler := range handlers {
		if err := handler(evt); err != nil {
			return err
		}
	}
	return nil
}

func (e *MockEventBus) Subscribe(pluginName, event string, handler EventHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.subscriptions[event]; !exists {
		e.subscriptions[event] = make(map[string]EventHandler)
	}
	e.subscriptions[event][pluginName] = handler
	return nil
}

func (e *MockEventBus) Unsubscribe(pluginName, event string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if handlers, exists := e.subscriptions[event]; exists {
		delete(handlers, pluginName)
		if len(handlers) == 0 {
			delete(e.subscriptions, event)
		}
	}
	return nil
}

func (e *MockEventBus) UnregisterAll(pluginName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for event, handlers := range e.subscriptions {
		delete(handlers, pluginName)
		if len(handlers) == 0 {
			delete(e.subscriptions, event)
		}
	}
	return nil
}

func (e *MockEventBus) ListSubscriptions(event string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	handlers, exists := e.subscriptions[event]
	if !exists {
		return []string{}
	}

	plugins := make([]string, 0, len(handlers))
	for pluginName := range handlers {
		plugins = append(plugins, pluginName)
	}
	return plugins
}

// MockPermissionChecker checks permissions for testing
type MockPermissionChecker struct {
	permissions PluginPermissions
}

// NewMockPermissionChecker creates a new mock permission checker
func NewMockPermissionChecker(permissions PluginPermissions) PermissionChecker {
	return &MockPermissionChecker{
		permissions: permissions,
	}
}

func (m *MockPermissionChecker) CheckPermission(pluginName string, permission string) error {
	switch permission {
	case "database":
		if !m.permissions.AllowDatabase {
			return fmt.Errorf("permission denied: database access not allowed")
		}
	case "cache":
		if !m.permissions.AllowCache {
			return fmt.Errorf("permission denied: cache access not allowed")
		}
	case "config":
		if !m.permissions.AllowConfig {
			return fmt.Errorf("permission denied: config access not allowed")
		}
	case "router":
		if !m.permissions.AllowRouter {
			return fmt.Errorf("permission denied: router access not allowed")
		}
	case "filesystem":
		if !m.permissions.AllowFileSystem {
			return fmt.Errorf("permission denied: filesystem access not allowed")
		}
	case "network":
		if !m.permissions.AllowNetwork {
			return fmt.Errorf("permission denied: network access not allowed")
		}
	case "exec":
		if !m.permissions.AllowExec {
			return fmt.Errorf("permission denied: exec access not allowed")
		}
	default:
		if m.permissions.CustomPermissions != nil {
			if allowed, exists := m.permissions.CustomPermissions[permission]; exists && allowed {
				return nil
			}
		}
		return fmt.Errorf("permission denied: unknown permission %s", permission)
	}
	return nil
}

func (m *MockPermissionChecker) GrantPermission(pluginName string, permission string) error {
	return nil
}

func (m *MockPermissionChecker) RevokePermission(pluginName string, permission string) error {
	return nil
}

func (m *MockPermissionChecker) GetPermissions(pluginName string) PluginPermissions {
	return m.permissions
}

// ============================================================================
// Test Helper Functions
// ============================================================================

// AssertPluginInitialized checks if a plugin initialized successfully
func AssertPluginInitialized(t interface {
	Errorf(format string, args ...interface{})
}, plugin Plugin, ctx PluginContext) {
	if err := plugin.Initialize(ctx); err != nil {
		t.Errorf("Plugin initialization failed: %v", err)
	}
}

// AssertPluginLifecycle tests the complete plugin lifecycle
func AssertPluginLifecycle(t interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}, plugin Plugin, ctx PluginContext) {
	// Initialize
	if err := plugin.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start
	if err := plugin.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop
	if err := plugin.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Cleanup
	if err := plugin.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
}

// AssertHookRegistered checks if a hook was registered
func AssertHookRegistered(t interface {
	Errorf(format string, args ...interface{})
}, ctx *MockPluginContext, hookType HookType) {
	for _, hook := range ctx.RegisteredHooks {
		if hook.HookType == hookType {
			return
		}
	}
	t.Errorf("Hook %s was not registered", hookType)
}

// AssertEventPublished checks if an event was published
func AssertEventPublished(t interface {
	Errorf(format string, args ...interface{})
}, ctx *MockPluginContext, event string) {
	for _, pub := range ctx.PublishedEvents {
		if pub.Event == event {
			return
		}
	}
	t.Errorf("Event %s was not published", event)
}

// AssertServiceExported checks if a service was exported
func AssertServiceExported(t interface {
	Errorf(format string, args ...interface{})
}, ctx *MockPluginContext, serviceName string) {
	if _, exists := ctx.ExportedServices[serviceName]; !exists {
		t.Errorf("Service %s was not exported", serviceName)
	}
}

// AssertMiddlewareRegistered checks if middleware was registered
func AssertMiddlewareRegistered(t interface {
	Errorf(format string, args ...interface{})
}, ctx *MockPluginContext, middlewareName string) {
	for _, mw := range ctx.RegisteredMiddleware {
		if mw.Name == middlewareName {
			return
		}
	}
	t.Errorf("Middleware %s was not registered", middlewareName)
}
