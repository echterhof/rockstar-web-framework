package pkg

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

// TestParseJSON tests JSON parsing functionality
func TestParseJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		target   interface{}
		wantErr  bool
		validate func(t *testing.T, target interface{})
	}{
		{
			name:    "valid JSON object",
			data:    `{"name":"John","age":30}`,
			target:  &map[string]interface{}{},
			wantErr: false,
			validate: func(t *testing.T, target interface{}) {
				m := target.(*map[string]interface{})
				if (*m)["name"] != "John" {
					t.Errorf("expected name=John, got %v", (*m)["name"])
				}
			},
		},
		{
			name:    "valid JSON array",
			data:    `[1,2,3]`,
			target:  &[]int{},
			wantErr: false,
			validate: func(t *testing.T, target interface{}) {
				arr := target.(*[]int)
				if len(*arr) != 3 {
					t.Errorf("expected length=3, got %d", len(*arr))
				}
			},
		},
		{
			name:    "empty JSON",
			data:    ``,
			target:  &map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			target:  &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseJSON([]byte(tt.data), tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tt.target)
			}
		})
	}
}

// TestToJSON tests JSON serialization
func TestToJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
		check   func(t *testing.T, result []byte)
	}{
		{
			name:    "simple object",
			data:    map[string]string{"key": "value"},
			wantErr: false,
			check: func(t *testing.T, result []byte) {
				var m map[string]string
				if err := json.Unmarshal(result, &m); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
				}
				if m["key"] != "value" {
					t.Errorf("expected key=value, got %v", m["key"])
				}
			},
		},
		{
			name:    "array",
			data:    []int{1, 2, 3},
			wantErr: false,
			check: func(t *testing.T, result []byte) {
				var arr []int
				if err := json.Unmarshal(result, &arr); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
				}
				if len(arr) != 3 {
					t.Errorf("expected length=3, got %d", len(arr))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestRESTError tests REST error functionality
func TestRESTError(t *testing.T) {
	t.Run("create error", func(t *testing.T) {
		err := NewRESTError("TEST_ERROR", "Test message", 400)
		if err.Code != "TEST_ERROR" {
			t.Errorf("expected code=TEST_ERROR, got %s", err.Code)
		}
		if err.Message != "Test message" {
			t.Errorf("expected message='Test message', got %s", err.Message)
		}
		if err.Status != 400 {
			t.Errorf("expected status=400, got %d", err.Status)
		}
	})

	t.Run("error with details", func(t *testing.T) {
		err := NewRESTError("TEST_ERROR", "Test message", 400)
		details := map[string]interface{}{"field": "value"}
		err = err.WithDetails(details)

		if err.Details["field"] != "value" {
			t.Errorf("expected details field=value, got %v", err.Details["field"])
		}
	})

	t.Run("error string", func(t *testing.T) {
		err := NewRESTError("TEST_ERROR", "Test message", 400)
		expected := "TEST_ERROR: Test message"
		if err.Error() != expected {
			t.Errorf("expected error string=%s, got %s", expected, err.Error())
		}
	})
}

// TestCommonRESTErrors tests predefined REST errors
func TestCommonRESTErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *RESTError
		status int
		code   string
	}{
		{"bad request", ErrBadRequest, 400, "BAD_REQUEST"},
		{"unauthorized", ErrUnauthorized, 401, "UNAUTHORIZED"},
		{"forbidden", ErrForbidden, 403, "FORBIDDEN"},
		{"not found", ErrNotFound, 404, "NOT_FOUND"},
		{"method not allowed", ErrMethodNotAllowed, 405, "METHOD_NOT_ALLOWED"},
		{"conflict", ErrConflict, 409, "CONFLICT"},
		{"rate limit exceeded", ErrRateLimitExceeded, 429, "RATE_LIMIT_EXCEEDED"},
		{"internal server error", ErrInternalServerError, 500, "INTERNAL_SERVER_ERROR"},
		{"service unavailable", ErrServiceUnavailable, 503, "SERVICE_UNAVAILABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Status != tt.status {
				t.Errorf("expected status=%d, got %d", tt.status, tt.err.Status)
			}
			if tt.err.Code != tt.code {
				t.Errorf("expected code=%s, got %s", tt.code, tt.err.Code)
			}
		})
	}
}

// TestRESTRateLimitConfig tests rate limit configuration
func TestRESTRateLimitConfig(t *testing.T) {
	config := &RESTRateLimitConfig{
		Limit:  100,
		Window: time.Minute,
		Key:    "user_id",
	}

	if config.Limit != 100 {
		t.Errorf("expected limit=100, got %d", config.Limit)
	}
	if config.Window != time.Minute {
		t.Errorf("expected window=1m, got %v", config.Window)
	}
	if config.Key != "user_id" {
		t.Errorf("expected key=user_id, got %s", config.Key)
	}
}

// TestRESTRouteConfig tests REST route configuration
func TestRESTRouteConfig(t *testing.T) {
	config := RESTRouteConfig{
		RateLimit: &RESTRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "user_id",
		},
		RequireAuth:    true,
		RequiredScopes: []string{"read", "write"},
		MaxRequestSize: 1024 * 1024, // 1MB
		Timeout:        30 * time.Second,
	}

	if config.RateLimit.Limit != 100 {
		t.Errorf("expected rate limit=100, got %d", config.RateLimit.Limit)
	}
	if !config.RequireAuth {
		t.Error("expected RequireAuth=true")
	}
	if len(config.RequiredScopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(config.RequiredScopes))
	}
	if config.MaxRequestSize != 1024*1024 {
		t.Errorf("expected max request size=1MB, got %d", config.MaxRequestSize)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout=30s, got %v", config.Timeout)
	}
}

// TestRESTResponse tests REST response structure
func TestRESTResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		response := RESTResponse{
			Success: true,
			Data:    map[string]string{"key": "value"},
			Meta: &RESTMeta{
				Timestamp: time.Now(),
				RequestID: "req-123",
			},
		}

		if !response.Success {
			t.Error("expected success=true")
		}
		if response.Error != nil {
			t.Error("expected no error in success response")
		}
		if response.Meta.RequestID != "req-123" {
			t.Errorf("expected request ID=req-123, got %s", response.Meta.RequestID)
		}
	})

	t.Run("error response", func(t *testing.T) {
		response := RESTResponse{
			Success: false,
			Error: &RESTError{
				Code:    "TEST_ERROR",
				Message: "Test error",
				Status:  400,
			},
			Meta: &RESTMeta{
				Timestamp: time.Now(),
				RequestID: "req-456",
			},
		}

		if response.Success {
			t.Error("expected success=false")
		}
		if response.Error == nil {
			t.Error("expected error in error response")
		}
		if response.Error.Code != "TEST_ERROR" {
			t.Errorf("expected error code=TEST_ERROR, got %s", response.Error.Code)
		}
	})
}

// TestPagination tests pagination structure
func TestPagination(t *testing.T) {
	pagination := &Pagination{
		Page:       2,
		PerPage:    10,
		Total:      100,
		TotalPages: 10,
	}

	if pagination.Page != 2 {
		t.Errorf("expected page=2, got %d", pagination.Page)
	}
	if pagination.PerPage != 10 {
		t.Errorf("expected per_page=10, got %d", pagination.PerPage)
	}
	if pagination.Total != 100 {
		t.Errorf("expected total=100, got %d", pagination.Total)
	}
	if pagination.TotalPages != 10 {
		t.Errorf("expected total_pages=10, got %d", pagination.TotalPages)
	}
}

// TestRateLimitInfo tests rate limit info structure
func TestRateLimitInfo(t *testing.T) {
	now := time.Now().Unix()
	info := &RateLimitInfo{
		Limit:     100,
		Remaining: 50,
		Reset:     now + 60,
	}

	if info.Limit != 100 {
		t.Errorf("expected limit=100, got %d", info.Limit)
	}
	if info.Remaining != 50 {
		t.Errorf("expected remaining=50, got %d", info.Remaining)
	}
	if info.Reset != now+60 {
		t.Errorf("expected reset=%d, got %d", now+60, info.Reset)
	}
}

// Mock implementations for testing

type mockRouter struct {
	routes map[string]map[string]HandlerFunc
}

func newMockRouter() *mockRouter {
	return &mockRouter{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

func (m *mockRouter) addRoute(method, path string, handler HandlerFunc) {
	if m.routes[method] == nil {
		m.routes[method] = make(map[string]HandlerFunc)
	}
	m.routes[method][path] = handler
}

func (m *mockRouter) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("GET", path, handler)
	return m
}

func (m *mockRouter) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("POST", path, handler)
	return m
}

func (m *mockRouter) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("PUT", path, handler)
	return m
}

func (m *mockRouter) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("DELETE", path, handler)
	return m
}

func (m *mockRouter) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("PATCH", path, handler)
	return m
}

func (m *mockRouter) HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("HEAD", path, handler)
	return m
}

func (m *mockRouter) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute("OPTIONS", path, handler)
	return m
}

func (m *mockRouter) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	m.addRoute(method, path, handler)
	return m
}

func (m *mockRouter) Group(prefix string, middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) Host(hostname string) RouterEngine {
	return m
}

func (m *mockRouter) Static(prefix string, filesystem VirtualFS) RouterEngine {
	return m
}

func (m *mockRouter) StaticFile(path, filepath string) RouterEngine {
	return m
}

func (m *mockRouter) WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) Use(middleware ...MiddlewareFunc) RouterEngine {
	return m
}

func (m *mockRouter) Match(method, path, host string) (*Route, map[string]string, bool) {
	return nil, nil, false
}

func (m *mockRouter) Routes() []*Route {
	return nil
}

type mockDB struct {
	rateLimits map[string]int
}

func newMockDB() *mockDB {
	return &mockDB{
		rateLimits: make(map[string]int),
	}
}

func (m *mockDB) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	count, exists := m.rateLimits[key]
	if !exists {
		return true, nil
	}
	return count < limit, nil
}

func (m *mockDB) IncrementRateLimit(key string, window time.Duration) error {
	m.rateLimits[key]++
	return nil
}

// Implement other DatabaseManager methods as no-ops
func (m *mockDB) Connect(config DatabaseConfig) error                            { return nil }
func (m *mockDB) Close() error                                                   { return nil }
func (m *mockDB) Ping() error                                                    { return nil }
func (m *mockDB) Stats() DatabaseStats                                           { return DatabaseStats{} }
func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error)     { return nil, nil }
func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row            { return nil }
func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error)     { return nil, nil }
func (m *mockDB) Prepare(query string) (*sql.Stmt, error)                        { return nil, nil }
func (m *mockDB) Begin() (Transaction, error)                                    { return nil, nil }
func (m *mockDB) BeginTx(opts *sql.TxOptions) (Transaction, error)               { return nil, nil }
func (m *mockDB) Save(model interface{}) error                                   { return nil }
func (m *mockDB) Find(model interface{}, conditions ...Condition) error          { return nil }
func (m *mockDB) FindAll(models interface{}, conditions ...Condition) error      { return nil }
func (m *mockDB) Delete(model interface{}) error                                 { return nil }
func (m *mockDB) Update(model interface{}, updates map[string]interface{}) error { return nil }
func (m *mockDB) SaveSession(session *Session) error                             { return nil }
func (m *mockDB) LoadSession(sessionID string) (*Session, error)                 { return nil, nil }
func (m *mockDB) DeleteSession(sessionID string) error                           { return nil }
func (m *mockDB) CleanupExpiredSessions() error                                  { return nil }
func (m *mockDB) SaveAccessToken(token *AccessToken) error                       { return nil }
func (m *mockDB) LoadAccessToken(tokenValue string) (*AccessToken, error)        { return nil, nil }
func (m *mockDB) ValidateAccessToken(tokenValue string) (*AccessToken, error)    { return nil, nil }
func (m *mockDB) DeleteAccessToken(tokenValue string) error                      { return nil }
func (m *mockDB) CleanupExpiredTokens() error                                    { return nil }
func (m *mockDB) SaveTenant(tenant *Tenant) error                                { return nil }
func (m *mockDB) LoadTenant(tenantID string) (*Tenant, error)                    { return nil, nil }
func (m *mockDB) LoadTenantByHost(hostname string) (*Tenant, error)              { return nil, nil }
func (m *mockDB) SaveWorkloadMetrics(metrics *WorkloadMetrics) error             { return nil }
func (m *mockDB) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *mockDB) Migrate() error      { return nil }
func (m *mockDB) CreateTables() error { return nil }
func (m *mockDB) DropTables() error   { return nil }

// TestRESTAPIManager tests the REST API manager
func TestRESTAPIManager(t *testing.T) {
	router := newMockRouter()
	db := newMockDB()
	manager := NewRESTAPIManager(router, db)

	t.Run("register GET route", func(t *testing.T) {
		handler := func(ctx Context) error {
			return nil
		}

		config := RESTRouteConfig{}
		err := manager.RegisterRoute("GET", "/test", handler, config)
		if err != nil {
			t.Errorf("failed to register route: %v", err)
		}

		if router.routes["GET"]["/test"] == nil {
			t.Error("route not registered")
		}
	})

	t.Run("register POST route", func(t *testing.T) {
		handler := func(ctx Context) error {
			return nil
		}

		config := RESTRouteConfig{}
		err := manager.RegisterRoute("POST", "/test", handler, config)
		if err != nil {
			t.Errorf("failed to register route: %v", err)
		}

		if router.routes["POST"]["/test"] == nil {
			t.Error("route not registered")
		}
	})

	t.Run("register unsupported method", func(t *testing.T) {
		handler := func(ctx Context) error {
			return nil
		}

		config := RESTRouteConfig{}
		err := manager.RegisterRoute("INVALID", "/test", handler, config)
		if err == nil {
			t.Error("expected error for unsupported method")
		}
	})

	t.Run("create group", func(t *testing.T) {
		group := manager.Group("/api/v1")
		if group == nil {
			t.Error("failed to create group")
		}
	})
}

// TestRateLimiter tests the rate limiter
func TestRateLimiter(t *testing.T) {
	db := newMockDB()
	limiter := newRateLimiter(db)

	t.Run("check rate limit - allowed", func(t *testing.T) {
		allowed, err := limiter.Check("test-key", 10, time.Minute)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("expected request to be allowed")
		}
	})

	t.Run("increment and check", func(t *testing.T) {
		key := "test-key-2"
		limit := 3

		// Should allow first 3 requests
		for i := 0; i < limit; i++ {
			allowed, err := limiter.Check(key, limit, time.Minute)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !allowed {
				t.Errorf("request %d should be allowed", i+1)
			}

			err = limiter.Increment(key, time.Minute)
			if err != nil {
				t.Errorf("failed to increment: %v", err)
			}
		}

		// 4th request should be denied
		allowed, err := limiter.Check(key, limit, time.Minute)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if allowed {
			t.Error("request should be denied after limit reached")
		}
	})

	t.Run("nil database", func(t *testing.T) {
		limiter := newRateLimiter(nil)
		allowed, err := limiter.Check("test-key", 10, time.Minute)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("should allow all requests when database is nil")
		}
	})
}

// Benchmark tests
func BenchmarkParseJSON(b *testing.B) {
	data := []byte(`{"name":"John","age":30,"email":"john@example.com"}`)
	target := &map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseJSON(data, target)
	}
}

func BenchmarkToJSON(b *testing.B) {
	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToJSON(data)
	}
}

func BenchmarkRESTError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := NewRESTError("TEST_ERROR", "Test message", 400)
		_ = err.WithDetails(map[string]interface{}{"field": "value"})
	}
}
