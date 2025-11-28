package pkg

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestValidateInputLengths_URLLength(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxURLLength = 100

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test URL within limit
	req := &Request{
		RequestURI: "/api/users",
		Header:     make(http.Header),
		Query:      make(map[string]string),
		Form:       make(map[string]string),
	}
	ctx := &validationMockContext{request: req}

	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid URL, got: %v", err)
	}

	// Test URL exceeding limit
	req.RequestURI = "/" + strings.Repeat("a", 200)
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for URL exceeding limit")
	}
}

func TestValidateInputLengths_HeaderSize(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxHeaderSize = 100

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test header within limit
	req := &Request{
		RequestURI: "/api/users",
		Header:     make(http.Header),
		Query:      make(map[string]string),
		Form:       make(map[string]string),
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	ctx := &validationMockContext{request: req}

	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid header, got: %v", err)
	}

	// Test header exceeding limit
	req.Header.Set("X-Large-Header", strings.Repeat("a", 200))
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for header exceeding limit")
	}
}

func TestValidateInputLengths_QueryParams(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxQueryParams = 5
	config.MaxFormFieldSize = 100

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test query params within limit
	req := &Request{
		RequestURI: "/api/users",
		Header:     make(http.Header),
		Query: map[string]string{
			"page":  "1",
			"limit": "10",
		},
		Form: make(map[string]string),
	}
	ctx := &validationMockContext{request: req}

	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid query params, got: %v", err)
	}

	// Test too many query params
	req.Query = make(map[string]string)
	for i := 0; i < 10; i++ {
		req.Query[string(rune('a'+i))] = "value"
	}
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for too many query params")
	}

	// Test query param value too large
	req.Query = map[string]string{
		"data": strings.Repeat("a", 200),
	}
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for query param value too large")
	}
}

func TestValidateInputLengths_FormFields(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxFormFields = 5
	config.MaxFormFieldSize = 100

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test form fields within limit
	req := &Request{
		RequestURI: "/api/users",
		Header:     make(http.Header),
		Query:      make(map[string]string),
		Form: map[string]string{
			"name":  "John",
			"email": "john@example.com",
		},
	}
	ctx := &validationMockContext{request: req}

	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid form fields, got: %v", err)
	}

	// Test too many form fields
	req.Form = make(map[string]string)
	for i := 0; i < 10; i++ {
		req.Form[string(rune('a'+i))] = "value"
	}
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for too many form fields")
	}

	// Test form field value too large
	req.Form = map[string]string{
		"description": strings.Repeat("a", 200),
	}
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for form field value too large")
	}
}

func TestValidateInputLengths_FileName(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxFileNameLength = 50

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test filename within limit
	req := &Request{
		RequestURI: "/api/upload",
		Header:     make(http.Header),
		Query:      make(map[string]string),
		Form:       make(map[string]string),
		Files: map[string]*FormFile{
			"avatar": {
				Filename: "profile.jpg",
				Size:     1024,
			},
		},
	}
	ctx := &validationMockContext{request: req}

	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid filename, got: %v", err)
	}

	// Test filename exceeding limit
	req.Files["avatar"].Filename = strings.Repeat("a", 100) + ".jpg"
	err = sm.(*securityManagerImpl).ValidateInputLengths(ctx)
	if err == nil {
		t.Error("Expected error for filename exceeding limit")
	}
}

func TestValidateRequest_WithInputLengths(t *testing.T) {
	db := NewNoopDatabaseManager()
	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.MaxURLLength = 100
	config.MaxHeaderSize = 100

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Test valid request
	req := &Request{
		RequestURI: "/api/users",
		Header:     make(http.Header),
		Query:      make(map[string]string),
		Form:       make(map[string]string),
		RawBody:    []byte("test"),
		StartTime:  time.Now(),
	}
	req.Header.Set("Content-Type", "application/json")
	ctx := &validationMockContext{request: req}

	err = sm.ValidateRequest(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}

	// Test request with URL too long
	req.RequestURI = "/" + strings.Repeat("a", 200)
	err = sm.ValidateRequest(ctx)
	if err == nil {
		t.Error("Expected error for request with URL too long")
	}
}

// Mock context for validation testing
type validationMockContext struct {
	request *Request
	cookies map[string]*Cookie
}

func (m *validationMockContext) Request() *Request                           { return m.request }
func (m *validationMockContext) Response() ResponseWriter                    { return nil }
func (m *validationMockContext) Params() map[string]string                   { return nil }
func (m *validationMockContext) Param(name string) string                    { return "" }
func (m *validationMockContext) Query() map[string]string                    { return m.request.Query }
func (m *validationMockContext) Headers() map[string]string                  { return nil }
func (m *validationMockContext) Body() []byte                                { return nil }
func (m *validationMockContext) Session() SessionManager                     { return nil }
func (m *validationMockContext) User() *User                                 { return nil }
func (m *validationMockContext) Tenant() *Tenant                             { return nil }
func (m *validationMockContext) DB() DatabaseManager                         { return nil }
func (m *validationMockContext) Cache() CacheManager                         { return nil }
func (m *validationMockContext) Config() ConfigManager                       { return nil }
func (m *validationMockContext) I18n() I18nManager                           { return nil }
func (m *validationMockContext) Files() FileManager                          { return nil }
func (m *validationMockContext) Logger() Logger                              { return nil }
func (m *validationMockContext) Metrics() MetricsCollector                   { return nil }
func (m *validationMockContext) Context() context.Context                    { return nil }
func (m *validationMockContext) WithTimeout(timeout time.Duration) Context   { return m }
func (m *validationMockContext) WithCancel() (Context, context.CancelFunc)   { return m, func() {} }
func (m *validationMockContext) JSON(statusCode int, data interface{}) error { return nil }
func (m *validationMockContext) XML(statusCode int, data interface{}) error  { return nil }
func (m *validationMockContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}
func (m *validationMockContext) String(statusCode int, message string) error { return nil }
func (m *validationMockContext) Redirect(statusCode int, url string) error   { return nil }
func (m *validationMockContext) SetCookie(cookie *Cookie) error {
	m.cookies[cookie.Name] = cookie
	return nil
}
func (m *validationMockContext) GetCookie(name string) (*Cookie, error) {
	if cookie, ok := m.cookies[name]; ok {
		return cookie, nil
	}
	return nil, fmt.Errorf("cookie not found")
}
func (m *validationMockContext) SetHeader(key, value string)               {}
func (m *validationMockContext) GetHeader(key string) string               { return "" }
func (m *validationMockContext) Set(key string, value interface{})         {}
func (m *validationMockContext) Get(key string) (interface{}, bool)        { return nil, false }
func (m *validationMockContext) FormValue(key string) string               { return "" }
func (m *validationMockContext) FormFile(key string) (*FormFile, error)    { return nil, nil }
func (m *validationMockContext) IsAuthenticated() bool                     { return false }
func (m *validationMockContext) IsAuthorized(resource, action string) bool { return false }
