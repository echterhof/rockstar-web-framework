package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Example: Using SessionManager with Database Storage
	// Session storage implementation is in pkg/session_impl.go
	fmt.Println("=== Session Manager Example ===\n")

	// 1. Setup Database
	db := pkg.NewDatabaseManager()
	dbConfig := pkg.DatabaseConfig{
		Driver:          "sqlite3",
		Database:        "./sessions.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	if err := db.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// 2. Configure Session Manager
	encryptionKey := make([]byte, 32) // AES-256 requires 32 bytes
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	sessionConfig := &pkg.SessionConfig{
		StorageType:     pkg.SessionStorageDatabase,
		CookieName:      "app_session",
		CookiePath:      "/",
		CookieSecure:    true,
		CookieHTTPOnly:  true,
		CookieSameSite:  "Lax",
		SessionLifetime: 24 * time.Hour,
		EncryptionKey:   encryptionKey,
		CleanupInterval: 1 * time.Hour,
	}

	// 3. Create Session Manager
	sessionManager, err := pkg.NewSessionManager(sessionConfig, db, nil)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}

	// 4. Create a mock context for demonstration
	ctx := newMockContext()

	// 5. Create a new session
	fmt.Println("Creating new session...")
	session, err := sessionManager.Create(ctx)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	fmt.Printf("✓ Session created: ID=%s\n", session.ID)
	fmt.Printf("  Expires at: %s\n\n", session.ExpiresAt.Format(time.RFC3339))

	// 6. Store data in session
	fmt.Println("Storing data in session...")
	session.Data["username"] = "john_doe"
	session.Data["email"] = "john@example.com"
	session.Data["role"] = "admin"
	session.Data["login_time"] = time.Now()

	if err := sessionManager.Save(ctx, session); err != nil {
		log.Fatalf("Failed to save session: %v", err)
	}
	fmt.Println("✓ Session data saved\n")

	// 7. Load session from storage
	fmt.Println("Loading session from database...")
	loadedSession, err := sessionManager.Load(ctx, session.ID)
	if err != nil {
		log.Fatalf("Failed to load session: %v", err)
	}
	fmt.Printf("✓ Session loaded: ID=%s\n", loadedSession.ID)
	fmt.Printf("  Username: %v\n", loadedSession.Data["username"])
	fmt.Printf("  Email: %v\n", loadedSession.Data["email"])
	fmt.Printf("  Role: %v\n\n", loadedSession.Data["role"])

	// 8. Update session data using Set method
	fmt.Println("Updating session data...")
	if err := sessionManager.Set(session.ID, "last_activity", time.Now()); err != nil {
		log.Fatalf("Failed to set session data: %v", err)
	}
	fmt.Println("✓ Session updated\n")

	// 9. Get specific value from session
	fmt.Println("Retrieving specific session value...")
	username, err := sessionManager.Get(session.ID, "username")
	if err != nil {
		log.Fatalf("Failed to get session value: %v", err)
	}
	fmt.Printf("✓ Retrieved username: %v\n\n", username)

	// 10. Check session validity
	fmt.Println("Checking session validity...")
	isValid := sessionManager.IsValid(session.ID)
	isExpired := sessionManager.IsExpired(session.ID)
	fmt.Printf("✓ Session valid: %v\n", isValid)
	fmt.Printf("  Session expired: %v\n\n", isExpired)

	// 11. Refresh session expiration
	fmt.Println("Refreshing session...")
	if err := sessionManager.Refresh(ctx, session.ID); err != nil {
		log.Fatalf("Failed to refresh session: %v", err)
	}
	refreshedSession, _ := sessionManager.Load(ctx, session.ID)
	fmt.Printf("✓ Session refreshed\n")
	fmt.Printf("  New expiration: %s\n\n", refreshedSession.ExpiresAt.Format(time.RFC3339))

	// 12. Delete specific key from session
	fmt.Println("Deleting session key...")
	if err := sessionManager.Delete(session.ID, "email"); err != nil {
		log.Fatalf("Failed to delete session key: %v", err)
	}
	fmt.Println("✓ Email key deleted from session\n")

	// 13. Demonstrate cookie operations
	fmt.Println("Setting session cookie...")
	if err := sessionManager.SetCookie(ctx, session); err != nil {
		log.Fatalf("Failed to set session cookie: %v", err)
	}
	fmt.Println("✓ Session cookie set (encrypted)\n")

	fmt.Println("Retrieving session from cookie...")
	cookieSession, err := sessionManager.GetSessionFromCookie(ctx)
	if err != nil {
		log.Fatalf("Failed to get session from cookie: %v", err)
	}
	fmt.Printf("✓ Session retrieved from cookie: ID=%s\n\n", cookieSession.ID)

	// 14. Create multiple sessions for cleanup demo
	fmt.Println("Creating additional sessions for cleanup demo...")
	for i := 0; i < 3; i++ {
		tempSession, _ := sessionManager.Create(ctx)
		fmt.Printf("  Created session %d: %s\n", i+1, tempSession.ID)
	}
	fmt.Println()

	// 15. Cleanup expired sessions
	fmt.Println("Cleaning up expired sessions...")
	if err := sessionManager.CleanupExpired(); err != nil {
		log.Fatalf("Failed to cleanup sessions: %v", err)
	}
	fmt.Println("✓ Expired sessions cleaned up\n")

	// 16. Clear all data from session
	fmt.Println("Clearing session data...")
	if err := sessionManager.Clear(session.ID); err != nil {
		log.Fatalf("Failed to clear session: %v", err)
	}
	clearedSession, _ := sessionManager.Load(ctx, session.ID)
	fmt.Printf("✓ Session data cleared (data count: %d)\n\n", len(clearedSession.Data))

	// 17. Destroy session
	fmt.Println("Destroying session...")
	if err := sessionManager.Destroy(ctx, session.ID); err != nil {
		log.Fatalf("Failed to destroy session: %v", err)
	}
	fmt.Println("✓ Session destroyed\n")

	// 18. Verify session is gone
	fmt.Println("Verifying session deletion...")
	_, err = sessionManager.Load(ctx, session.ID)
	if err != nil {
		fmt.Printf("✓ Session no longer exists: %v\n\n", err)
	}

	// 19. Multi-tenant session example
	fmt.Println("=== Multi-Tenant Session Example ===\n")
	demonstrateMultiTenantSessions(sessionManager, db)

	fmt.Println("=== Session Manager Example Complete ===")
}

// demonstrateMultiTenantSessions shows session isolation between tenants
func demonstrateMultiTenantSessions(sm pkg.SessionManager, db pkg.DatabaseManager) {
	// Create two tenants
	tenant1 := &pkg.Tenant{
		ID:       "tenant-1",
		Name:     "Acme Corp",
		Hosts:    []string{"acme.example.com"},
		IsActive: true,
		Config:   make(map[string]interface{}),
	}

	tenant2 := &pkg.Tenant{
		ID:       "tenant-2",
		Name:     "TechStart Inc",
		Hosts:    []string{"techstart.example.com"},
		IsActive: true,
		Config:   make(map[string]interface{}),
	}

	db.SaveTenant(tenant1)
	db.SaveTenant(tenant2)

	// Create contexts for each tenant
	ctx1 := newMockContextWithTenant(tenant1)
	ctx2 := newMockContextWithTenant(tenant2)

	// Create sessions for each tenant
	session1, _ := sm.Create(ctx1)
	session2, _ := sm.Create(ctx2)

	fmt.Printf("Created session for %s: %s\n", tenant1.Name, session1.ID)
	fmt.Printf("Created session for %s: %s\n", tenant2.Name, session2.ID)

	// Store tenant-specific data
	session1.Data["company"] = "Acme Corp"
	session1.Data["plan"] = "enterprise"
	sm.Save(ctx1, session1)

	session2.Data["company"] = "TechStart Inc"
	session2.Data["plan"] = "startup"
	sm.Save(ctx2, session2)

	// Load and verify isolation
	loaded1, _ := sm.Load(ctx1, session1.ID)
	loaded2, _ := sm.Load(ctx2, session2.ID)

	fmt.Printf("\nTenant 1 session data: %v\n", loaded1.Data)
	fmt.Printf("Tenant 2 session data: %v\n\n", loaded2.Data)
}

// Mock context implementation for demonstration
type mockContext struct {
	cookies map[string]*pkg.Cookie
	tenant  *pkg.Tenant
	user    *pkg.User
}

func newMockContext() *mockContext {
	return &mockContext{
		cookies: make(map[string]*pkg.Cookie),
	}
}

func newMockContextWithTenant(tenant *pkg.Tenant) *mockContext {
	return &mockContext{
		cookies: make(map[string]*pkg.Cookie),
		tenant:  tenant,
	}
}

func (m *mockContext) Request() *pkg.Request                         { return nil }
func (m *mockContext) Response() pkg.ResponseWriter                  { return nil }
func (m *mockContext) Params() map[string]string                     { return nil }
func (m *mockContext) Query() map[string]string                      { return nil }
func (m *mockContext) Headers() map[string]string                    { return nil }
func (m *mockContext) Body() []byte                                  { return nil }
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
func (m *mockContext) SetHeader(key, value string)                 {}
func (m *mockContext) GetHeader(key string) string                 { return "" }
func (m *mockContext) FormValue(key string) string                 { return "" }
func (m *mockContext) FormFile(key string) (*pkg.FormFile, error)  { return nil, nil }
func (m *mockContext) IsAuthenticated() bool                       { return false }
func (m *mockContext) IsAuthorized(resource, action string) bool   { return false }
func (m *mockContext) SetCookie(cookie *pkg.Cookie) error {
	m.cookies[cookie.Name] = cookie
	return nil
}
func (m *mockContext) GetCookie(name string) (*pkg.Cookie, error) {
	if cookie, exists := m.cookies[name]; exists {
		return cookie, nil
	}
	return nil, fmt.Errorf("cookie not found: %s", name)
}
func (m *mockContext) DeleteCookie(name string) {
	delete(m.cookies, name)
}
