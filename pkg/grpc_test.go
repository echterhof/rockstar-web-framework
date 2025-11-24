package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// mockGRPCService implements GRPCService for testing
type mockGRPCService struct {
	name    string
	methods []string
}

func (m *mockGRPCService) ServiceName() string {
	return m.name
}

func (m *mockGRPCService) Methods() []string {
	return m.methods
}

// mockGRPCServiceExtended implements GRPCServiceExtended for testing
type mockGRPCServiceExtended struct {
	mockGRPCService
	handleUnaryFunc  func(ctx context.Context, method string, req interface{}) (interface{}, error)
	handleStreamFunc func(stream GRPCServerStream, method string) error
}

func (m *mockGRPCServiceExtended) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
	if m.handleUnaryFunc != nil {
		return m.handleUnaryFunc(ctx, method, req)
	}
	return map[string]string{"result": "success"}, nil
}

func (m *mockGRPCServiceExtended) HandleStream(stream GRPCServerStream, method string) error {
	if m.handleStreamFunc != nil {
		return m.handleStreamFunc(stream, method)
	}
	return nil
}

func (m *mockGRPCServiceExtended) GetMethodDescriptor(method string) *GRPCMethodDescriptor {
	return &GRPCMethodDescriptor{
		Name:           method,
		IsClientStream: false,
		IsServerStream: false,
	}
}

// Helper function to create a test context
func newGRPCTestContext(method, path string, body []byte) Context {
	req := &Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Header:     http.Header{},
		RemoteAddr: "127.0.0.1",
		RawBody:    body,
		ID:         "test-req",
		Params:     make(map[string]string),
		Query:      make(map[string]string),
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	return NewContext(req, resp, context.Background())
}

// mockGRPCDB is a simple mock database for testing
type mockGRPCDB struct {
	rateLimits map[string]int
}

func newMockGRPCDB() *mockGRPCDB {
	return &mockGRPCDB{
		rateLimits: make(map[string]int),
	}
}

func (m *mockGRPCDB) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	count, exists := m.rateLimits[key]
	if !exists {
		return true, nil
	}
	return count < limit, nil
}
func (m *mockGRPCDB) IncrementRateLimit(key string, window time.Duration) error {
	m.rateLimits[key]++
	return nil
}
func (m *mockGRPCDB) Connect(config DatabaseConfig) error                         { return nil }
func (m *mockGRPCDB) Close() error                                                { return nil }
func (m *mockGRPCDB) Ping() error                                                 { return nil }
func (m *mockGRPCDB) Stats() DatabaseStats                                        { return DatabaseStats{} }
func (m *mockGRPCDB) Query(query string, args ...interface{}) (*sql.Rows, error)  { return nil, nil }
func (m *mockGRPCDB) QueryRow(query string, args ...interface{}) *sql.Row         { return nil }
func (m *mockGRPCDB) Exec(query string, args ...interface{}) (sql.Result, error)  { return nil, nil }
func (m *mockGRPCDB) Prepare(query string) (*sql.Stmt, error)                     { return nil, nil }
func (m *mockGRPCDB) Begin() (Transaction, error)                                 { return nil, nil }
func (m *mockGRPCDB) BeginTx(opts *sql.TxOptions) (Transaction, error)            { return nil, nil }
func (m *mockGRPCDB) SaveSession(session *Session) error                          { return nil }
func (m *mockGRPCDB) LoadSession(sessionID string) (*Session, error)              { return nil, nil }
func (m *mockGRPCDB) DeleteSession(sessionID string) error                        { return nil }
func (m *mockGRPCDB) CleanupExpiredSessions() error                               { return nil }
func (m *mockGRPCDB) SaveAccessToken(token *AccessToken) error                    { return nil }
func (m *mockGRPCDB) LoadAccessToken(tokenValue string) (*AccessToken, error)     { return nil, nil }
func (m *mockGRPCDB) ValidateAccessToken(tokenValue string) (*AccessToken, error) { return nil, nil }
func (m *mockGRPCDB) DeleteAccessToken(tokenValue string) error                   { return nil }
func (m *mockGRPCDB) CleanupExpiredTokens() error                                 { return nil }
func (m *mockGRPCDB) SaveTenant(tenant *Tenant) error                             { return nil }
func (m *mockGRPCDB) LoadTenant(tenantID string) (*Tenant, error)                 { return nil, nil }
func (m *mockGRPCDB) LoadTenantByHost(hostname string) (*Tenant, error)           { return nil, nil }
func (m *mockGRPCDB) SaveWorkloadMetrics(metrics *WorkloadMetrics) error          { return nil }
func (m *mockGRPCDB) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *mockGRPCDB) Migrate() error      { return nil }
func (m *mockGRPCDB) CreateTables() error { return nil }
func (m *mockGRPCDB) DropTables() error   { return nil }

// TestNewGRPCManager tests creating a new gRPC manager
// Requirements: 2.3
func TestNewGRPCManager(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewGRPCManager(router, db, authManager)

	if manager == nil {
		t.Fatal("Expected non-nil gRPC manager")
	}
}

// TestRegisterService tests registering a gRPC service
// Requirements: 2.3
func TestRegisterService(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser", "CreateUser"},
		},
	}

	config := GRPCConfig{
		RequireAuth: false,
	}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Verify routes were registered
	routes := router.Routes()
	if len(routes) < 2 {
		t.Errorf("Expected at least 2 routes, got %d", len(routes))
	}
}

// TestRegisterServiceDuplicate tests registering a duplicate service
// Requirements: 2.3
func TestRegisterServiceDuplicate(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser"},
		},
	}

	config := GRPCConfig{}

	// Register first time
	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Try to register again
	err = manager.RegisterService(service, config)
	if err == nil {
		t.Error("Expected error when registering duplicate service")
	}
}

// TestGRPCRateLimiting tests rate limiting for gRPC
// Requirements: 2.6
func TestGRPCRateLimiting(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser"},
		},
	}

	config := GRPCConfig{
		RateLimit: &GRPCRateLimitConfig{
			Limit:  2,
			Window: time.Minute,
			Key:    "ip_address",
		},
	}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create test context
	ctx := newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "123"}`))

	// First request should succeed
	route, _, found := router.Match("POST", "/TestService/GetUser", "")
	if !found {
		t.Fatal("Route not found")
	}

	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("First request failed: %v", err)
	}

	// Second request should succeed
	ctx = newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "456"}`))

	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("Second request failed: %v", err)
	}

	// Third request should be rate limited
	ctx = newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "789"}`))

	err = route.Handler(ctx)
	if err == nil {
		t.Error("Expected rate limit error on third request")
	}
}

// TestGRPCAuthentication tests authentication for gRPC
// Requirements: 2.5
func TestGRPCAuthentication(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser"},
		},
	}

	config := GRPCConfig{
		RequireAuth: true,
	}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create test context without authentication
	ctx := newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "123"}`))

	route, _, found := router.Match("POST", "/TestService/GetUser", "")
	if !found {
		t.Fatal("Route not found")
	}

	// Request without auth should fail
	err = route.Handler(ctx)
	if err == nil {
		t.Error("Expected authentication error")
	}

	// Create authenticated context
	ctx = newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "123"}`))
	if ctxImpl, ok := ctx.(*contextImpl); ok {
		ctxImpl.SetUser(&User{
			ID:    "user123",
			Email: "test@example.com",
			Roles: []string{"user"},
		})
	}

	// Request with auth should succeed
	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("Authenticated request failed: %v", err)
	}
}

// TestGRPCAuthorizationRoles tests role-based authorization for gRPC
// Requirements: 2.5
func TestGRPCAuthorizationRoles(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"AdminAction"},
		},
	}

	config := GRPCConfig{
		RequireAuth:   true,
		RequiredRoles: []string{"admin"},
	}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create context with user role
	ctx := newGRPCTestContext("POST", "/TestService/AdminAction", []byte(`{"action": "delete"}`))
	if ctxImpl, ok := ctx.(*contextImpl); ok {
		ctxImpl.SetUser(&User{
			ID:    "user123",
			Email: "test@example.com",
			Roles: []string{"user"},
		})
	}

	route, _, found := router.Match("POST", "/TestService/AdminAction", "")
	if !found {
		t.Fatal("Route not found")
	}

	// Request without admin role should fail
	err = route.Handler(ctx)
	if err == nil {
		t.Error("Expected authorization error for non-admin user")
	}

	// Create context with admin role
	ctx = newGRPCTestContext("POST", "/TestService/AdminAction", []byte(`{"action": "delete"}`))
	if ctxImpl, ok := ctx.(*contextImpl); ok {
		ctxImpl.SetUser(&User{
			ID:    "admin123",
			Email: "admin@example.com",
			Roles: []string{"admin"},
		})
	}

	// Request with admin role should succeed
	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("Admin request failed: %v", err)
	}
}

// TestGRPCRequestValidation tests request size validation
// Requirements: 2.3
func TestGRPCRequestValidation(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"Upload"},
		},
	}

	config := GRPCConfig{
		MaxRequestSize: 100, // 100 bytes max
	}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create context with large request
	ctx := newGRPCTestContext("POST", "/TestService/Upload", make([]byte, 200))

	route, _, found := router.Match("POST", "/TestService/Upload", "")
	if !found {
		t.Fatal("Route not found")
	}

	// Large request should fail
	err = route.Handler(ctx)
	if err == nil {
		t.Error("Expected validation error for large request")
	}

	// Create context with small request
	ctx = newGRPCTestContext("POST", "/TestService/Upload", make([]byte, 50))

	// Small request should succeed
	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("Small request failed: %v", err)
	}
}

// TestGRPCMiddleware tests custom middleware
// Requirements: 2.3
func TestGRPCMiddleware(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	// Add custom middleware
	middlewareCalled := false
	manager.Use(func(ctx Context, next GRPCHandler) error {
		middlewareCalled = true
		return next(ctx)
	})

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser"},
		},
	}

	config := GRPCConfig{}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create test context
	ctx := newGRPCTestContext("POST", "/TestService/GetUser", []byte(`{"id": "123"}`))

	route, _, found := router.Match("POST", "/TestService/GetUser", "")
	if !found {
		t.Fatal("Route not found")
	}

	err = route.Handler(ctx)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
}

// TestGRPCGroup tests service grouping
// Requirements: 2.3
func TestGRPCGroup(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	// Create a group with prefix
	group := manager.Group("/api/v1")

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"GetUser"},
		},
	}

	config := GRPCConfig{}

	err := group.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service in group: %v", err)
	}

	// Verify route was registered with prefix
	route, _, found := router.Match("POST", "/api/v1/TestService/GetUser", "")
	if !found {
		t.Error("Route with prefix not found")
	}

	if route == nil {
		t.Error("Expected non-nil route")
	}
}

// TestGRPCErrorHandling tests error response handling
// Requirements: 2.3
func TestGRPCErrorHandling(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	service := &mockGRPCServiceExtended{
		mockGRPCService: mockGRPCService{
			name:    "TestService",
			methods: []string{"FailingMethod"},
		},
		handleUnaryFunc: func(ctx context.Context, method string, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("intentional error")
		},
	}

	config := GRPCConfig{}

	err := manager.RegisterService(service, config)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create test context
	ctx := newGRPCTestContext("POST", "/TestService/FailingMethod", []byte(`{"data": "test"}`))

	route, _, found := router.Match("POST", "/TestService/FailingMethod", "")
	if !found {
		t.Fatal("Route not found")
	}

	err = route.Handler(ctx)
	if err == nil {
		t.Error("Expected error from failing method")
	}
}

// TestGRPCStatusCodeMapping tests gRPC to HTTP status code mapping
// Requirements: 2.3
func TestGRPCStatusCodeMapping(t *testing.T) {
	tests := []struct {
		grpcCode   GRPCStatusCode
		httpStatus int
	}{
		{GRPCStatusOK, 200},
		{GRPCStatusInvalidArgument, 400},
		{GRPCStatusUnauthenticated, 401},
		{GRPCStatusPermissionDenied, 403},
		{GRPCStatusNotFound, 404},
		{GRPCStatusResourceExhausted, 429},
		{GRPCStatusInternal, 500},
		{GRPCStatusUnavailable, 503},
	}

	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager).(*grpcManager)

	for _, tt := range tests {
		httpStatus := manager.grpcStatusToHTTP(tt.grpcCode)
		if httpStatus != tt.httpStatus {
			t.Errorf("grpcStatusToHTTP(%d) = %d, want %d", tt.grpcCode, httpStatus, tt.httpStatus)
		}
	}
}

// TestCheckRateLimit tests the CheckRateLimit method
// Requirements: 2.6
func TestCheckRateLimit(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	ctx := newGRPCTestContext("POST", "/test", []byte(`{}`))

	// First check should succeed
	err := manager.CheckRateLimit(ctx, "test_resource")
	if err != nil {
		t.Errorf("First rate limit check failed: %v", err)
	}
}

// TestCheckGlobalRateLimit tests the CheckGlobalRateLimit method
// Requirements: 2.6
func TestCheckGlobalRateLimit(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	ctx := newGRPCTestContext("POST", "/test", []byte(`{}`))

	// First check should succeed
	err := manager.CheckGlobalRateLimit(ctx)
	if err != nil {
		t.Errorf("First global rate limit check failed: %v", err)
	}
}

// TestGRPCServerLifecycle tests server start/stop
// Requirements: 2.3
func TestGRPCServerLifecycle(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	// Start server
	err := manager.Start(":50051")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Try to start again (should fail)
	err = manager.Start(":50051")
	if err == nil {
		t.Error("Expected error when starting already running server")
	}

	// Stop server
	err = manager.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Try to stop again (should fail)
	err = manager.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped server")
	}
}

// TestGRPCGracefulStop tests graceful shutdown
// Requirements: 2.3
func TestGRPCGracefulStop(t *testing.T) {
	router := NewRouter()
	db := newMockGRPCDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	manager := NewGRPCManager(router, db, authManager)

	// Start server
	err := manager.Start(":50051")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Graceful stop
	err = manager.GracefulStop(5 * time.Second)
	if err != nil {
		t.Errorf("Graceful stop failed: %v", err)
	}
}
