package tests

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// mockContext is a mock implementation of Context for testing
type mockContext struct {
	user    *pkg.User
	tenant  *pkg.Tenant
	request *pkg.Request
	cookies map[string]*pkg.Cookie
	headers map[string]string
	body    []byte
	params  map[string]string
	query   map[string]string
}

func newMockContext() *mockContext {
	return &mockContext{
		cookies: make(map[string]*pkg.Cookie),
		headers: make(map[string]string),
		params:  make(map[string]string),
		query:   make(map[string]string),
	}
}

func (m *mockContext) Request() *pkg.Request                         { return m.request }
func (m *mockContext) Response() pkg.ResponseWriter                  { return nil }
func (m *mockContext) Params() map[string]string                     { return m.params }
func (m *mockContext) Query() map[string]string                      { return m.query }
func (m *mockContext) Headers() map[string]string                    { return m.headers }
func (m *mockContext) Body() []byte                                  { return m.body }
func (m *mockContext) Session() pkg.SessionManager                   { return nil }
func (m *mockContext) User() *pkg.User                               { return m.user }
func (m *mockContext) Tenant() *pkg.Tenant                           { return m.tenant }
func (m *mockContext) DB() pkg.DatabaseManager                       { return nil }
func (m *mockContext) Cache() pkg.CacheManager                       { return nil }
func (m *mockContext) Config() pkg.ConfigManager                     { return nil }
func (m *mockContext) I18n() pkg.I18nManager                         { return nil }
func (m *mockContext) Files() pkg.FileManager                        { return nil }
func (m *mockContext) Logger() pkg.Logger                            { return nil }
func (m *mockContext) Metrics() pkg.MetricsCollector                 { return nil }
func (m *mockContext) Context() context.Context                      { return context.Background() }
func (m *mockContext) WithTimeout(timeout time.Duration) pkg.Context { return m }
func (m *mockContext) WithCancel() (pkg.Context, context.CancelFunc) {
	return m, func() {}
}
func (m *mockContext) JSON(statusCode int, data interface{}) error { return nil }
func (m *mockContext) XML(statusCode int, data interface{}) error  { return nil }
func (m *mockContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}
func (m *mockContext) String(statusCode int, message string) error { return nil }
func (m *mockContext) Redirect(statusCode int, url string) error   { return nil }

func (m *mockContext) SetCookie(cookie *pkg.Cookie) error {
	if m.cookies == nil {
		m.cookies = make(map[string]*pkg.Cookie)
	}
	m.cookies[cookie.Name] = cookie
	return nil
}

func (m *mockContext) GetCookie(name string) (*pkg.Cookie, error) {
	if m.cookies == nil {
		return nil, errors.New("cookie not found")
	}
	cookie, exists := m.cookies[name]
	if !exists {
		return nil, errors.New("cookie not found")
	}
	return cookie, nil
}

func (m *mockContext) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockContext) GetHeader(key string) string {
	if m.headers == nil {
		return ""
	}
	return m.headers[key]
}

func (m *mockContext) FormValue(key string) string                { return "" }
func (m *mockContext) FormFile(key string) (*pkg.FormFile, error) { return nil, nil }
func (m *mockContext) IsAuthenticated() bool                      { return m.user != nil }
func (m *mockContext) IsAuthorized(resource, action string) bool  { return false }

// mockVirtualFS is a mock implementation of VirtualFS for testing
type mockVirtualFS struct{}

func (m *mockVirtualFS) Open(name string) (http.File, error) {
	return nil, nil
}

func (m *mockVirtualFS) Exists(name string) bool {
	return true
}

// Mock database manager for integration tests
type testMockDB struct {
	accessTokens map[string]*pkg.AccessToken
	sessions     map[string]*pkg.Session
	tenants      map[string]*pkg.Tenant
	metrics      []*pkg.WorkloadMetrics
}

func newTestMockDB() *testMockDB {
	return &testMockDB{
		accessTokens: make(map[string]*pkg.AccessToken),
		sessions:     make(map[string]*pkg.Session),
		tenants:      make(map[string]*pkg.Tenant),
		metrics:      make([]*pkg.WorkloadMetrics, 0),
	}
}

func (m *testMockDB) Connect(config pkg.DatabaseConfig) error { return nil }
func (m *testMockDB) Close() error                            { return nil }
func (m *testMockDB) Ping() error                             { return nil }
func (m *testMockDB) Stats() pkg.DatabaseStats                { return pkg.DatabaseStats{} }

func (m *testMockDB) SaveAccessToken(token *pkg.AccessToken) error {
	m.accessTokens[token.Token] = token
	return nil
}

func (m *testMockDB) LoadAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	token, exists := m.accessTokens[tokenValue]
	if !exists {
		return nil, errors.New("token not found")
	}
	return token, nil
}

func (m *testMockDB) ValidateAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	token, err := m.LoadAccessToken(tokenValue)
	if err != nil {
		return nil, err
	}
	if token.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	return token, nil
}

func (m *testMockDB) DeleteAccessToken(tokenValue string) error {
	delete(m.accessTokens, tokenValue)
	return nil
}

func (m *testMockDB) CleanupExpiredTokens() error { return nil }

func (m *testMockDB) SaveSession(session *pkg.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *testMockDB) LoadSession(sessionID string) (*pkg.Session, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (m *testMockDB) DeleteSession(sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func (m *testMockDB) CleanupExpiredSessions() error { return nil }

func (m *testMockDB) SaveTenant(tenant *pkg.Tenant) error {
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *testMockDB) LoadTenant(tenantID string) (*pkg.Tenant, error) {
	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, errors.New("tenant not found")
	}
	return tenant, nil
}

func (m *testMockDB) LoadTenantByHost(hostname string) (*pkg.Tenant, error) {
	for _, tenant := range m.tenants {
		for _, host := range tenant.Hosts {
			if host == hostname {
				return tenant, nil
			}
		}
	}
	return nil, errors.New("tenant not found")
}

func (m *testMockDB) SaveWorkloadMetrics(metrics *pkg.WorkloadMetrics) error {
	m.metrics = append(m.metrics, metrics)
	return nil
}

func (m *testMockDB) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*pkg.WorkloadMetrics, error) {
	result := make([]*pkg.WorkloadMetrics, 0)
	for _, metric := range m.metrics {
		if metric.TenantID == tenantID && metric.Timestamp.After(from) && metric.Timestamp.Before(to) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (m *testMockDB) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}

func (m *testMockDB) IncrementRateLimit(key string, window time.Duration) error {
	return nil
}

// Stub implementations for other DatabaseManager methods
func (m *testMockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *testMockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *testMockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *testMockDB) Prepare(query string) (*sql.Stmt, error) { return nil, nil }
func (m *testMockDB) Begin() (pkg.Transaction, error)         { return nil, nil }
func (m *testMockDB) BeginTx(opts *sql.TxOptions) (pkg.Transaction, error) {
	return nil, nil
}

// Note: Save, Find, FindAll, Delete, Update are not part of DatabaseManager interface
// They were removed as they're not used by the framework
func (m *testMockDB) Migrate() error                { return nil }
func (m *testMockDB) CreateTables() error           { return nil }
func (m *testMockDB) DropTables() error             { return nil }
func (m *testMockDB) InitializePluginTables() error { return nil }

// Helper functions for creating test data

func createTestUser(id, username, tenantID string, roles, actions []string) *pkg.User {
	return &pkg.User{
		ID:         id,
		Username:   username,
		Email:      username + "@example.com",
		Roles:      roles,
		Actions:    actions,
		TenantID:   tenantID,
		Metadata:   make(map[string]interface{}),
		AuthMethod: "test",
		AuthTime:   time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
}

func createTestTenant(id, name string, hosts []string) *pkg.Tenant {
	return &pkg.Tenant{
		ID:       id,
		Name:     name,
		Hosts:    hosts,
		Config:   make(map[string]interface{}),
		IsActive: true,
	}
}

func createTestAccessToken(token, userID, tenantID string, scopes []string, expiresIn time.Duration) *pkg.AccessToken {
	return &pkg.AccessToken{
		Token:     token,
		UserID:    userID,
		TenantID:  tenantID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
	}
}

func createTestSession(id, userID, tenantID string) *pkg.Session {
	return &pkg.Session{
		ID:        id,
		UserID:    userID,
		TenantID:  tenantID,
		Data:      make(map[string]interface{}),
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent/1.0",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Test server configuration helpers

func createTestServerConfig() pkg.ServerConfig {
	return pkg.ServerConfig{
		EnableHTTP1:     true,
		EnableHTTP2:     false,
		EnableQUIC:      false,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1 << 20,
		ShutdownTimeout: 10 * time.Second,
	}
}

func createTestSessionConfig() *pkg.SessionConfig {
	config := pkg.DefaultSessionConfig()
	config.EncryptionKey = make([]byte, 32)
	config.StorageType = pkg.SessionStorageDatabase
	config.SessionLifetime = 1 * time.Hour
	return config
}

// Assertion helpers

func assertEqual(t interface {
	Errorf(format string, args ...interface{})
}, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

func assertNotNil(t interface {
	Errorf(format string, args ...interface{})
}, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: expected non-nil value", message)
	}
}

func assertNil(t interface {
	Errorf(format string, args ...interface{})
}, value interface{}, message string) {
	if value != nil {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

func assertNoError(t interface {
	Fatalf(format string, args ...interface{})
}, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

func assertError(t interface {
	Errorf(format string, args ...interface{})
}, err error, message string) {
	if err == nil {
		t.Errorf("%s: expected error, got nil", message)
	}
}

func assertTrue(t interface {
	Errorf(format string, args ...interface{})
}, condition bool, message string) {
	if !condition {
		t.Errorf("%s: expected true, got false", message)
	}
}

func assertFalse(t interface {
	Errorf(format string, args ...interface{})
}, condition bool, message string) {
	if condition {
		t.Errorf("%s: expected false, got true", message)
	}
}

// HTTP test helpers

func makeGetRequest(url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

func makePostRequest(url string, body []byte, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

// Concurrency test helpers

func runConcurrent(count int, fn func(int)) {
	done := make(chan bool, count)

	for i := 0; i < count; i++ {
		go func(id int) {
			fn(id)
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}
}

func runConcurrentWithErrors(count int, fn func(int) error) []error {
	done := make(chan bool, count)
	errors := make(chan error, count)

	for i := 0; i < count; i++ {
		go func(id int) {
			if err := fn(id); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}

	close(errors)

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	return errs
}

// Timing helpers

func measureExecutionTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

func waitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// Cleanup helpers

func cleanupServer(server pkg.Server) {
	if server != nil && server.IsRunning() {
		server.Close()
	}
}

func cleanupServers(servers []pkg.Server) {
	for _, server := range servers {
		cleanupServer(server)
	}
}
