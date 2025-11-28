package pkg

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestSecurityConfig_DefaultsSecure tests that default security config is secure
func TestSecurityConfig_DefaultsSecure(t *testing.T) {
	config := DefaultSecurityConfig()

	// CORS should not have wildcard by default
	if len(config.AllowedOrigins) > 0 {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				t.Error("Default config should not allow wildcard CORS origin")
			}
		}
	}

	// HSTS should be enabled by default
	if !config.EnableHSTS {
		t.Error("HSTS should be enabled by default")
	}

	// HSTS max age should be at least 1 year
	if config.HSTSMaxAge < 31536000 {
		t.Errorf("HSTS max age should be at least 1 year, got %d seconds", config.HSTSMaxAge)
	}

	// HSTS should include subdomains by default
	if !config.HSTSIncludeSubdomains {
		t.Error("HSTS should include subdomains by default")
	}

	// Production mode should be disabled by default (must be explicitly enabled)
	if config.ProductionMode {
		t.Error("Production mode should be disabled by default")
	}
}

// TestSecurityManager_HSTSHeader tests HSTS header generation
func TestSecurityManager_HSTSHeader(t *testing.T) {
	config := DefaultSecurityConfig()
	config.EncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	config.JWTSecret = "test-secret"

	sm, err := NewSecurityManager(NewNoopDatabaseManager(), config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	ctx := &mockSecurityContext{
		headers: make(map[string]string),
	}

	err = sm.SetSecurityHeaders(ctx)
	if err != nil {
		t.Fatalf("Failed to set security headers: %v", err)
	}

	// Check HSTS header
	hstsHeader := ctx.headers["Strict-Transport-Security"]
	if hstsHeader == "" {
		t.Error("HSTS header should be set")
	}

	// Should contain max-age
	if !strings.Contains(hstsHeader, "max-age=31536000") {
		t.Errorf("HSTS header should contain max-age=31536000, got: %s", hstsHeader)
	}

	// Should contain includeSubDomains
	if !strings.Contains(hstsHeader, "includeSubDomains") {
		t.Errorf("HSTS header should contain includeSubDomains, got: %s", hstsHeader)
	}
}

// TestSecurityManager_HSTSDisabled tests that HSTS can be disabled
func TestSecurityManager_HSTSDisabled(t *testing.T) {
	config := DefaultSecurityConfig()
	config.EncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	config.JWTSecret = "test-secret"
	config.EnableHSTS = false

	sm, err := NewSecurityManager(NewNoopDatabaseManager(), config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	ctx := &mockSecurityContext{
		headers: make(map[string]string),
	}

	err = sm.SetSecurityHeaders(ctx)
	if err != nil {
		t.Fatalf("Failed to set security headers: %v", err)
	}

	// HSTS header should not be set
	hstsHeader := ctx.headers["Strict-Transport-Security"]
	if hstsHeader != "" {
		t.Errorf("HSTS header should not be set when disabled, got: %s", hstsHeader)
	}
}

// TestSecurityManager_ProductionErrorMode tests production error mode
func TestSecurityManager_ProductionErrorMode(t *testing.T) {
	config := DefaultSecurityConfig()
	config.EncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	config.JWTSecret = "test-secret"
	config.ProductionMode = true

	sm, err := NewSecurityManager(NewNoopDatabaseManager(), config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	smImpl := sm.(*securityManagerImpl)

	// Test with FrameworkError
	originalErr := &FrameworkError{
		Code:       "TEST_ERROR",
		Message:    "Detailed error message with sensitive info",
		StatusCode: 500,
		I18nKey:    "error.test",
		Details: map[string]interface{}{
			"sensitive": "data",
		},
		UserID:   "user123",
		TenantID: "tenant456",
	}

	safeErr := smImpl.SafeError(originalErr)
	fwErr, ok := safeErr.(*FrameworkError)
	if !ok {
		t.Fatal("SafeError should return FrameworkError")
	}

	// Should preserve code and status
	if fwErr.Code != originalErr.Code {
		t.Errorf("Expected code %s, got %s", originalErr.Code, fwErr.Code)
	}

	// Should use i18n key instead of detailed message
	if fwErr.Message != originalErr.I18nKey {
		t.Errorf("Expected message to be i18n key %s, got %s", originalErr.I18nKey, fwErr.Message)
	}

	// Should not include sensitive details
	if fwErr.Details != nil {
		t.Error("Details should be nil in production mode")
	}

	if fwErr.UserID != "" {
		t.Error("UserID should be empty in production mode")
	}

	if fwErr.TenantID != "" {
		t.Error("TenantID should be empty in production mode")
	}
}

// TestSecurityManager_DevelopmentErrorMode tests development error mode
func TestSecurityManager_DevelopmentErrorMode(t *testing.T) {
	config := DefaultSecurityConfig()
	config.EncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	config.JWTSecret = "test-secret"
	config.ProductionMode = false

	sm, err := NewSecurityManager(NewNoopDatabaseManager(), config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	smImpl := sm.(*securityManagerImpl)

	// Test with FrameworkError
	originalErr := &FrameworkError{
		Code:       "TEST_ERROR",
		Message:    "Detailed error message",
		StatusCode: 500,
		I18nKey:    "error.test",
		Details: map[string]interface{}{
			"debug": "info",
		},
	}

	safeErr := smImpl.SafeError(originalErr)

	// In development mode, should return original error
	if safeErr != originalErr {
		t.Error("In development mode, SafeError should return original error")
	}
}

// TestMockSecurityContext_InterfaceCompliance verifies mockSecurityContext implements Context interface
func TestMockSecurityContext_InterfaceCompliance(t *testing.T) {
	// Verify that mockSecurityContext implements Context interface
	var _ Context = (*mockSecurityContext)(nil)

	// Create instance and test Body() method
	ctx := &mockSecurityContext{
		headers: make(map[string]string),
	}

	// Test Body() method returns expected value (nil in this case)
	body := ctx.Body()
	if body != nil {
		t.Errorf("Expected Body() to return nil, got %v", body)
	}

	// Verify other key methods work without panicking
	if ctx.Params() != nil {
		t.Error("Expected Params() to return nil")
	}

	if ctx.Param("test") != "" {
		t.Error("Expected Param() to return empty string")
	}

	if ctx.Query() != nil {
		t.Error("Expected Query() to return nil")
	}

	// Verify Headers() returns the initialized map
	headers := ctx.Headers()
	if headers == nil {
		t.Error("Expected Headers() to return non-nil map")
	}
}

// mockSecurityContext for testing
type mockSecurityContext struct {
	headers map[string]string
}

func (m *mockSecurityContext) SetHeader(key, value string) {
	m.headers[key] = value
}

func (m *mockSecurityContext) Request() *Request                           { return nil }
func (m *mockSecurityContext) Response() ResponseWriter                    { return nil }
func (m *mockSecurityContext) Params() map[string]string                   { return nil }
func (m *mockSecurityContext) Param(name string) string                    { return "" }
func (m *mockSecurityContext) Query() map[string]string                    { return nil }
func (m *mockSecurityContext) Headers() map[string]string                  { return m.headers }
func (m *mockSecurityContext) Body() []byte                                { return nil }
func (m *mockSecurityContext) Session() SessionManager                     { return nil }
func (m *mockSecurityContext) User() *User                                 { return nil }
func (m *mockSecurityContext) Tenant() *Tenant                             { return nil }
func (m *mockSecurityContext) DB() DatabaseManager                         { return nil }
func (m *mockSecurityContext) Cache() CacheManager                         { return nil }
func (m *mockSecurityContext) Config() ConfigManager                       { return nil }
func (m *mockSecurityContext) I18n() I18nManager                           { return nil }
func (m *mockSecurityContext) Files() FileManager                          { return nil }
func (m *mockSecurityContext) Logger() Logger                              { return nil }
func (m *mockSecurityContext) Metrics() MetricsCollector                   { return nil }
func (m *mockSecurityContext) Context() context.Context                    { return context.Background() }
func (m *mockSecurityContext) WithTimeout(timeout time.Duration) Context   { return m }
func (m *mockSecurityContext) WithCancel() (Context, context.CancelFunc)   { return m, func() {} }
func (m *mockSecurityContext) JSON(statusCode int, data interface{}) error { return nil }
func (m *mockSecurityContext) XML(statusCode int, data interface{}) error  { return nil }
func (m *mockSecurityContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}
func (m *mockSecurityContext) String(statusCode int, message string) error { return nil }
func (m *mockSecurityContext) Redirect(statusCode int, url string) error   { return nil }
func (m *mockSecurityContext) SetCookie(cookie *Cookie) error              { return nil }
func (m *mockSecurityContext) GetCookie(name string) (*Cookie, error)      { return nil, nil }
func (m *mockSecurityContext) GetHeader(key string) string                 { return "" }
func (m *mockSecurityContext) Set(key string, value interface{})           {}
func (m *mockSecurityContext) Get(key string) (interface{}, bool)          { return nil, false }
func (m *mockSecurityContext) FormValue(key string) string                 { return "" }
func (m *mockSecurityContext) FormFile(key string) (*FormFile, error)      { return nil, nil }
func (m *mockSecurityContext) IsAuthenticated() bool                       { return false }
func (m *mockSecurityContext) IsAuthorized(resource, action string) bool   { return false }
