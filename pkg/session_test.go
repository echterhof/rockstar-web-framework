package pkg

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Mock implementations for testing

type mockContext struct {
	user            *User
	tenant          *Tenant
	request         *Request
	cookies         map[string]*Cookie
	headers         map[string]string
	isAuthenticated bool
}

func (m *mockContext) Request() *Request                                            { return m.request }
func (m *mockContext) Response() ResponseWriter                                     { return nil }
func (m *mockContext) Params() map[string]string                                    { return nil }
func (m *mockContext) Param(name string) string                                     { return "" }
func (m *mockContext) Query() map[string]string                                     { return nil }
func (m *mockContext) Headers() map[string]string                                   { return m.headers }
func (m *mockContext) Body() []byte                                                 { return nil }
func (m *mockContext) Session() SessionManager                                      { return nil }
func (m *mockContext) User() *User                                                  { return m.user }
func (m *mockContext) Tenant() *Tenant                                              { return m.tenant }
func (m *mockContext) DB() DatabaseManager                                          { return nil }
func (m *mockContext) Cache() CacheManager                                          { return nil }
func (m *mockContext) Config() ConfigManager                                        { return nil }
func (m *mockContext) I18n() I18nManager                                            { return nil }
func (m *mockContext) Files() FileManager                                           { return nil }
func (m *mockContext) Logger() Logger                                               { return nil }
func (m *mockContext) Metrics() MetricsCollector                                    { return nil }
func (m *mockContext) Context() context.Context                                     { return context.Background() }
func (m *mockContext) WithTimeout(timeout time.Duration) Context                    { return m }
func (m *mockContext) WithCancel() (Context, context.CancelFunc)                    { return m, func() {} }
func (m *mockContext) JSON(statusCode int, data interface{}) error                  { return nil }
func (m *mockContext) XML(statusCode int, data interface{}) error                   { return nil }
func (m *mockContext) HTML(statusCode int, template string, data interface{}) error { return nil }
func (m *mockContext) String(statusCode int, message string) error                  { return nil }
func (m *mockContext) Redirect(statusCode int, url string) error                    { return nil }

func (m *mockContext) SetCookie(cookie *Cookie) error {
	if m.cookies == nil {
		m.cookies = make(map[string]*Cookie)
	}
	m.cookies[cookie.Name] = cookie
	return nil
}

func (m *mockContext) GetCookie(name string) (*Cookie, error) {
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

func (m *mockContext) Set(key string, value interface{})      {}
func (m *mockContext) Get(key string) (interface{}, bool)     { return nil, false }
func (m *mockContext) FormValue(key string) string            { return "" }
func (m *mockContext) FormFile(key string) (*FormFile, error) { return nil, nil }
func (m *mockContext) IsAuthenticated() bool {
	if m.isAuthenticated {
		return true
	}
	return m.user != nil
}
func (m *mockContext) IsAuthorized(resource, action string) bool { return false }

type mockSessionCacheManager struct {
	data map[string]interface{}
}

func newMockSessionCacheManager() *mockSessionCacheManager {
	return &mockSessionCacheManager{
		data: make(map[string]interface{}),
	}
}

func (m *mockSessionCacheManager) Get(key string) (interface{}, error) {
	value, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *mockSessionCacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockSessionCacheManager) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockSessionCacheManager) Exists(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m *mockSessionCacheManager) Clear() error {
	m.data = make(map[string]interface{})
	return nil
}

// Implement other required methods with no-op
func (m *mockSessionCacheManager) GetMultiple(keys []string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockSessionCacheManager) SetMultiple(items map[string]interface{}, ttl time.Duration) error {
	return nil
}
func (m *mockSessionCacheManager) DeleteMultiple(keys []string) error { return nil }
func (m *mockSessionCacheManager) Increment(key string, delta int64) (int64, error) {
	return 0, nil
}
func (m *mockSessionCacheManager) Decrement(key string, delta int64) (int64, error) {
	return 0, nil
}
func (m *mockSessionCacheManager) Expire(key string, ttl time.Duration) error { return nil }
func (m *mockSessionCacheManager) TTL(key string) (time.Duration, error)      { return 0, nil }
func (m *mockSessionCacheManager) GetRequestCache(requestID string) RequestCache {
	return nil
}
func (m *mockSessionCacheManager) ClearRequestCache(requestID string) error { return nil }
func (m *mockSessionCacheManager) Invalidate(pattern string) error          { return nil }
func (m *mockSessionCacheManager) InvalidateTag(tag string) error           { return nil }

// Helper function to generate encryption key
func generateEncryptionKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

// Test session creation
func TestSessionManager_Create(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{
		user:    &User{ID: "user123"},
		tenant:  &Tenant{ID: "tenant456"},
		request: &Request{RemoteAddr: "192.168.1.1"},
		headers: map[string]string{"User-Agent": "TestAgent/1.0"},
	}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	if session.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", session.UserID)
	}

	if session.TenantID != "tenant456" {
		t.Errorf("Expected TenantID 'tenant456', got '%s'", session.TenantID)
	}

	if session.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IPAddress '192.168.1.1', got '%s'", session.IPAddress)
	}

	if session.Data == nil {
		t.Error("Session data should be initialized")
	}
}

// Test session loading
func TestSessionManager_Load(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	// Create a session
	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Load the session
	loadedSession, err := sm.Load(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID '%s', got '%s'", session.ID, loadedSession.ID)
	}
}

// Test session save
func TestSessionManager_Save(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Modify session data
	session.Data["key1"] = "value1"
	session.Data["key2"] = 42

	err = sm.Save(ctx, session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load and verify
	loadedSession, err := sm.Load(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.Data["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got '%v'", loadedSession.Data["key1"])
	}

	if loadedSession.Data["key2"] != 42 {
		t.Errorf("Expected key2=42, got '%v'", loadedSession.Data["key2"])
	}
}

// Test session destruction
func TestSessionManager_Destroy(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Destroy the session
	err = sm.Destroy(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to destroy session: %v", err)
	}

	// Try to load destroyed session
	_, err = sm.Load(ctx, session.ID)
	if err == nil {
		t.Error("Expected error when loading destroyed session")
	}
}

// Test session data operations
func TestSessionManager_DataOperations(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test Set
	err = sm.Set(session.ID, "testKey", "testValue")
	if err != nil {
		t.Fatalf("Failed to set session data: %v", err)
	}

	// Test Get
	value, err := sm.Get(session.ID, "testKey")
	if err != nil {
		t.Fatalf("Failed to get session data: %v", err)
	}

	if value != "testValue" {
		t.Errorf("Expected 'testValue', got '%v'", value)
	}

	// Test Delete
	err = sm.Delete(session.ID, "testKey")
	if err != nil {
		t.Fatalf("Failed to delete session data: %v", err)
	}

	_, err = sm.Get(session.ID, "testKey")
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	// Test Clear
	sm.Set(session.ID, "key1", "value1")
	sm.Set(session.ID, "key2", "value2")

	err = sm.Clear(session.ID)
	if err != nil {
		t.Fatalf("Failed to clear session data: %v", err)
	}

	loadedSession, _ := sm.Load(ctx, session.ID)
	if len(loadedSession.Data) != 0 {
		t.Errorf("Expected empty session data, got %d items", len(loadedSession.Data))
	}
}

// Test encrypted cookie operations
func TestSessionManager_EncryptedCookie(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{
		cookies: make(map[string]*Cookie),
	}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Set cookie
	err = sm.SetCookie(ctx, session)
	if err != nil {
		t.Fatalf("Failed to set cookie: %v", err)
	}

	// Verify cookie was set
	cookie, err := ctx.GetCookie(config.CookieName)
	if err != nil {
		t.Fatalf("Failed to get cookie: %v", err)
	}

	if cookie.Value == session.ID {
		t.Error("Cookie value should be encrypted, not plain session ID")
	}

	// Get session from cookie
	retrievedSession, err := sm.GetSessionFromCookie(ctx)
	if err != nil {
		t.Fatalf("Failed to get session from cookie: %v", err)
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Expected session ID '%s', got '%s'", session.ID, retrievedSession.ID)
	}
}

// Test session expiration
func TestSessionManager_Expiration(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase
	config.SessionLifetime = 100 * time.Millisecond

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Session should be valid initially
	if !sm.IsValid(session.ID) {
		t.Error("Session should be valid")
	}

	if sm.IsExpired(session.ID) {
		t.Error("Session should not be expired")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Session should be expired now
	if sm.IsValid(session.ID) {
		t.Error("Session should not be valid after expiration")
	}

	if !sm.IsExpired(session.ID) {
		t.Error("Session should be expired")
	}

	// Loading expired session should fail
	_, err = sm.Load(ctx, session.ID)
	if err == nil {
		t.Error("Expected error when loading expired session")
	}
}

// Test session refresh
func TestSessionManager_Refresh(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase
	config.SessionLifetime = 200 * time.Millisecond

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	originalExpiry := session.ExpiresAt

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Refresh the session
	err = sm.Refresh(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to refresh session: %v", err)
	}

	// Load and check expiry
	refreshedSession, err := sm.Load(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to load refreshed session: %v", err)
	}

	if !refreshedSession.ExpiresAt.After(originalExpiry) {
		t.Error("Refreshed session should have later expiry time")
	}
}

// Test cache storage backend
func TestSessionManager_CacheStorage(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageCache

	cache := newMockSessionCacheManager()
	sm, err := NewSessionManager(config, nil, cache)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session is in cache
	if !cache.Exists(sessionKey(session.ID)) {
		t.Error("Session should be stored in cache")
	}

	// Load from cache
	loadedSession, err := sm.Load(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to load session from cache: %v", err)
	}

	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID '%s', got '%s'", session.ID, loadedSession.ID)
	}
}

// Test filesystem storage backend
func TestSessionManager_FilesystemStorage(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "rockstar_sessions_test")
	defer os.RemoveAll(tempDir)

	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageFilesystem
	config.FilesystemPath = tempDir

	sm, err := NewSessionManager(config, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	session, err := sm.Create(ctx)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session file exists
	sessionFile := filepath.Join(tempDir, session.ID+".json")
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Error("Session file should exist")
	}

	// Load from filesystem
	loadedSession, err := sm.Load(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to load session from filesystem: %v", err)
	}

	if loadedSession.ID != session.ID {
		t.Errorf("Expected session ID '%s', got '%s'", session.ID, loadedSession.ID)
	}

	// Destroy and verify file is deleted
	err = sm.Destroy(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to destroy session: %v", err)
	}

	if _, err := os.Stat(sessionFile); !os.IsNotExist(err) {
		t.Error("Session file should be deleted")
	}
}

// Test cleanup of expired sessions
func TestSessionManager_CleanupExpired(t *testing.T) {
	config := DefaultSessionConfig()
	config.EncryptionKey = generateEncryptionKey()
	config.StorageType = SessionStorageDatabase
	config.SessionLifetime = 100 * time.Millisecond

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	ctx := &mockContext{}

	// Create multiple sessions
	session1, _ := sm.Create(ctx)
	session2, _ := sm.Create(ctx)
	session3, _ := sm.Create(ctx)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	err = sm.CleanupExpired()
	if err != nil {
		t.Fatalf("Failed to cleanup expired sessions: %v", err)
	}

	// All sessions should be cleaned up
	mockDB := db.(*MockDatabaseManager)
	if len(mockDB.sessions) != 0 {
		t.Errorf("Expected 0 sessions after cleanup, got %d", len(mockDB.sessions))
	}

	// Verify sessions cannot be loaded
	_, err = sm.Load(ctx, session1.ID)
	if err == nil {
		t.Error("Expected error when loading cleaned up session")
	}

	_, err = sm.Load(ctx, session2.ID)
	if err == nil {
		t.Error("Expected error when loading cleaned up session")
	}

	_, err = sm.Load(ctx, session3.ID)
	if err == nil {
		t.Error("Expected error when loading cleaned up session")
	}
}

// Property-based tests

// **Feature: optional-database, Property 7: SessionManager uses in-memory storage without database**
// **Validates: Requirements 3.1**
// For any SessionManager initialized with a no-op DatabaseManager, session operations (save, load, delete)
// should succeed using in-memory storage without attempting database operations
func TestProperty_SessionManagerInMemoryStorage(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("SessionManager uses in-memory storage when no database is available",
		prop.ForAll(
			func(userID string, tenantID string, dataKey string, dataValue string) bool {
				// Create session manager with no-op database
				config := DefaultSessionConfig()
				config.EncryptionKey = generateEncryptionKey()

				noopDB := NewNoopDatabaseManager()
				sm, err := NewSessionManager(config, noopDB, nil)
				if err != nil {
					return false
				}

				ctx := &mockContext{
					user:   &User{ID: userID},
					tenant: &Tenant{ID: tenantID},
				}

				// Create a session
				session, err := sm.Create(ctx)
				if err != nil {
					return false
				}

				// Verify session was created
				if session.ID == "" {
					return false
				}

				// Set data in session
				session.Data[dataKey] = dataValue
				err = sm.Save(ctx, session)
				if err != nil {
					return false
				}

				// Load session and verify data
				loadedSession, err := sm.Load(ctx, session.ID)
				if err != nil {
					return false
				}

				if loadedSession.Data[dataKey] != dataValue {
					return false
				}

				// Delete session
				err = sm.Destroy(ctx, session.ID)
				if err != nil {
					return false
				}

				// Verify session is deleted
				_, err = sm.Load(ctx, session.ID)
				if err == nil {
					return false // Should return error for deleted session
				}

				return true
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.Identifier(),
			gen.AnyString(),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
