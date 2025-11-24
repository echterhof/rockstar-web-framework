package pkg

import (
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Helper function to create a test security manager
func createTestSecurityManager(t *testing.T) SecurityManager {
	db := NewMockDatabaseManager()

	config := DefaultSecurityConfig()
	// Generate a valid 32-byte encryption key
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	return sm
}

// Helper function to create a test context
func createTestContext(t *testing.T, req *Request) Context {
	w := httptest.NewRecorder()
	baseCtx := context.Background()

	// Create a simple response writer wrapper
	respWriter := &testResponseWriter{
		recorder: w,
		headers:  make(http.Header),
	}

	// Re-create context with response writer
	return &contextImpl{
		request:  req,
		response: respWriter,
		ctx:      baseCtx,
		params:   make(map[string]string),
		query:    make(map[string]string),
		headers:  make(map[string]string),
	}
}

type testResponseWriter struct {
	recorder   *httptest.ResponseRecorder
	headers    http.Header
	statusCode int
}

func (w *testResponseWriter) Header() http.Header {
	return w.headers
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	return w.recorder.Write(b)
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.recorder.WriteHeader(statusCode)
}

func (w *testResponseWriter) WriteJSON(statusCode int, data interface{}) error {
	return nil
}

func (w *testResponseWriter) WriteXML(statusCode int, data interface{}) error {
	return nil
}

func (w *testResponseWriter) WriteHTML(statusCode int, template string, data interface{}) error {
	return nil
}

func (w *testResponseWriter) WriteString(statusCode int, message string) error {
	w.statusCode = statusCode
	return nil
}

func (w *testResponseWriter) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	return nil
}

func (w *testResponseWriter) SetCookie(cookie *Cookie) error {
	return nil
}

func (w *testResponseWriter) SetHeader(key, value string) {
	w.headers.Set(key, value)
}

func (w *testResponseWriter) SetContentType(contentType string) {
	w.headers.Set("Content-Type", contentType)
}

func (w *testResponseWriter) Status() int {
	return w.statusCode
}

func (w *testResponseWriter) Size() int64 {
	return 0
}

func (w *testResponseWriter) Written() bool {
	return w.statusCode != 0
}

func (w *testResponseWriter) Flush() error {
	return nil
}

func (w *testResponseWriter) Close() error {
	return nil
}

func (w *testResponseWriter) SetTemplateManager(tm TemplateManager) {
	// No-op for mock
}

// Test request size validation
func TestValidateRequestSize_WithinLimit(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Header:  http.Header{},
		RawBody: []byte("test body"),
	}
	req.Header.Set("Content-Length", "9")

	ctx := createTestContext(t, req)

	err := sm.ValidateRequestSize(ctx, 1024)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateRequestSize_ExceedsLimit(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Header:  http.Header{},
		RawBody: make([]byte, 2000),
	}
	req.Header.Set("Content-Length", "2000")

	ctx := createTestContext(t, req)

	err := sm.ValidateRequestSize(ctx, 1024)
	if err == nil {
		t.Error("Expected error for request exceeding size limit")
	}

	frameworkErr, ok := err.(*FrameworkError)
	if !ok {
		t.Errorf("Expected FrameworkError, got: %T", err)
	}

	if frameworkErr.Code != ErrCodeRequestTooLarge {
		t.Errorf("Expected error code %s, got: %s", ErrCodeRequestTooLarge, frameworkErr.Code)
	}
}

// Test request timeout validation
func TestValidateRequestTimeout_NotExpired(t *testing.T) {
	sm := createTestSecurityManager(t)

	baseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &Request{}
	ctx := &contextImpl{
		request: req,
		ctx:     baseCtx,
	}

	err := sm.ValidateRequestTimeout(ctx, 10*time.Second)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateRequestTimeout_Expired(t *testing.T) {
	sm := createTestSecurityManager(t)

	baseCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Wait for timeout

	req := &Request{}
	ctx := &contextImpl{
		request: req,
		ctx:     baseCtx,
	}

	err := sm.ValidateRequestTimeout(ctx, 1*time.Second)
	if err == nil {
		t.Error("Expected error for expired timeout")
	}
}

// Test bogus data detection
func TestValidateBogusData_Clean(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		RequestURI: "/api/users",
		Header:     http.Header{},
		RawBody:    []byte("clean data"),
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateBogusData(ctx)
	if err != nil {
		t.Errorf("Expected no error for clean data, got: %v", err)
	}
}

func TestValidateBogusData_NullBytesInURI(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		RequestURI: "/api/users\x00malicious",
		Header:     http.Header{},
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateBogusData(ctx)
	if err == nil {
		t.Error("Expected error for null bytes in URI")
	}

	frameworkErr, ok := err.(*FrameworkError)
	if !ok {
		t.Errorf("Expected FrameworkError, got: %T", err)
	}

	if frameworkErr.Code != ErrCodeBogusData {
		t.Errorf("Expected error code %s, got: %s", ErrCodeBogusData, frameworkErr.Code)
	}
}

func TestValidateBogusData_NullBytesInHeader(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		RequestURI: "/api/users",
		Header:     http.Header{},
	}
	req.Header.Set("X-Custom", "value\x00malicious")

	ctx := createTestContext(t, req)

	err := sm.ValidateBogusData(ctx)
	if err == nil {
		t.Error("Expected error for null bytes in header")
	}
}

// Test form validation
func TestValidateFormData_AllValid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"age":      "25",
		},
	}

	ctx := createTestContext(t, req)

	rules := ValidationRules{
		Required: []string{"username", "email"},
		Types: map[string]string{
			"email": "email",
			"age":   "int",
		},
		Lengths: map[string]LengthRule{
			"username": {Min: 3, Max: 20},
		},
	}

	err := sm.ValidateFormData(ctx, rules)
	if err != nil {
		t.Errorf("Expected no error for valid form data, got: %v", err)
	}
}

func TestValidateFormData_MissingRequired(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"username": "testuser",
		},
	}

	ctx := createTestContext(t, req)

	rules := ValidationRules{
		Required: []string{"username", "email"},
	}

	err := sm.ValidateFormData(ctx, rules)
	if err == nil {
		t.Error("Expected error for missing required field")
	}
}

func TestValidateFormData_InvalidEmail(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"email": "invalid-email",
		},
	}

	ctx := createTestContext(t, req)

	rules := ValidationRules{
		Types: map[string]string{
			"email": "email",
		},
	}

	err := sm.ValidateFormData(ctx, rules)
	if err == nil {
		t.Error("Expected error for invalid email")
	}
}

func TestValidateFormData_LengthViolation(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"username": "ab", // Too short
		},
	}

	ctx := createTestContext(t, req)

	rules := ValidationRules{
		Lengths: map[string]LengthRule{
			"username": {Min: 3, Max: 20},
		},
	}

	err := sm.ValidateFormData(ctx, rules)
	if err == nil {
		t.Error("Expected error for length violation")
	}
}

// Test file upload validation
func TestValidateFileUpload_Valid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{
			"avatar": {
				Filename: "profile.jpg",
				Size:     1024,
				Header: map[string][]string{
					"Content-Type": {"image/jpeg"},
				},
			},
		},
	}

	ctx := createTestContext(t, req)

	rules := FileValidationRules{
		MaxSize:      10 * 1024,
		AllowedTypes: []string{"image/jpeg", "image/png"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png"},
	}

	err := sm.ValidateFileUpload(ctx, rules)
	if err != nil {
		t.Errorf("Expected no error for valid file, got: %v", err)
	}
}

func TestValidateFileUpload_FileTooLarge(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{
			"avatar": {
				Filename: "profile.jpg",
				Size:     20 * 1024,
			},
		},
	}

	ctx := createTestContext(t, req)

	rules := FileValidationRules{
		MaxSize: 10 * 1024,
	}

	err := sm.ValidateFileUpload(ctx, rules)
	if err == nil {
		t.Error("Expected error for file too large")
	}

	frameworkErr, ok := err.(*FrameworkError)
	if !ok {
		t.Errorf("Expected FrameworkError, got: %T", err)
	}

	if frameworkErr.Code != ErrCodeFileTooLarge {
		t.Errorf("Expected error code %s, got: %s", ErrCodeFileTooLarge, frameworkErr.Code)
	}
}

func TestValidateFileUpload_InvalidExtension(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{
			"document": {
				Filename: "malicious.exe",
				Size:     1024,
			},
		},
	}

	ctx := createTestContext(t, req)

	rules := FileValidationRules{
		AllowedExts: []string{".pdf", ".doc", ".docx"},
	}

	err := sm.ValidateFileUpload(ctx, rules)
	if err == nil {
		t.Error("Expected error for invalid file extension")
	}
}

func TestValidateFileUpload_MissingRequired(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{},
	}

	ctx := createTestContext(t, req)

	rules := FileValidationRules{
		Required: []string{"avatar"},
	}

	err := sm.ValidateFileUpload(ctx, rules)
	if err == nil {
		t.Error("Expected error for missing required file")
	}
}

// Test expected form values validation
func TestValidateExpectedFormValues_AllPresent(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"username": "testuser",
			"password": "secret",
		},
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateExpectedFormValues(ctx, []string{"username", "password"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateExpectedFormValues_Missing(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Form: map[string]string{
			"username": "testuser",
		},
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateExpectedFormValues(ctx, []string{"username", "password"})
	if err == nil {
		t.Error("Expected error for missing expected field")
	}
}

// Test expected files validation
func TestValidateExpectedFiles_AllPresent(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{
			"avatar": {Filename: "profile.jpg"},
			"cover":  {Filename: "cover.jpg"},
		},
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateExpectedFiles(ctx, []string{"avatar", "cover"})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateExpectedFiles_Missing(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Files: map[string]*FormFile{
			"avatar": {Filename: "profile.jpg"},
		},
	}

	ctx := createTestContext(t, req)

	err := sm.ValidateExpectedFiles(ctx, []string{"avatar", "cover"})
	if err == nil {
		t.Error("Expected error for missing expected file")
	}
}

// Test security headers
func TestSetSecurityHeaders(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.SetSecurityHeaders(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that headers were set
	respWriter := ctx.Response().(*testResponseWriter)

	if respWriter.headers.Get("X-Frame-Options") == "" {
		t.Error("Expected X-Frame-Options header to be set")
	}

	if respWriter.headers.Get("X-Content-Type-Options") == "" {
		t.Error("Expected X-Content-Type-Options header to be set")
	}
}

func TestSetXFrameOptions_Valid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.SetXFrameOptions(ctx, "DENY")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	respWriter := ctx.Response().(*testResponseWriter)
	if respWriter.headers.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options to be DENY, got: %s", respWriter.headers.Get("X-Frame-Options"))
	}
}

func TestSetXFrameOptions_Invalid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.SetXFrameOptions(ctx, "INVALID")
	if err == nil {
		t.Error("Expected error for invalid X-Frame-Options value")
	}
}

// Test CORS
func TestEnableCORS_AllowedOrigin(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://example.com")

	ctx := createTestContext(t, req)

	config := CORSConfig{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET", "POST"},
	}

	err := sm.EnableCORS(ctx, config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	respWriter := ctx.Response().(*testResponseWriter)
	if respWriter.headers.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("Expected CORS header to be set")
	}
}

func TestEnableCORS_DisallowedOrigin(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Header: http.Header{},
	}
	req.Header.Set("Origin", "https://malicious.com")

	ctx := createTestContext(t, req)

	config := CORSConfig{
		AllowOrigins: []string{"https://example.com"},
	}

	err := sm.EnableCORS(ctx, config)
	if err == nil {
		t.Error("Expected error for disallowed origin")
	}
}

// Test XSS protection
func TestEnableXSSProtection(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.EnableXSSProtection(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	respWriter := ctx.Response().(*testResponseWriter)
	if respWriter.headers.Get("X-XSS-Protection") == "" {
		t.Error("Expected X-XSS-Protection header to be set")
	}
}

// Test CSRF protection
func TestEnableCSRFProtection(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	token, err := sm.EnableCSRFProtection(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty CSRF token")
	}
}

func TestValidateCSRFToken_Valid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		Header: http.Header{},
	}
	ctx := createTestContext(t, req)

	// Generate token
	token, err := sm.EnableCSRFProtection(ctx)
	if err != nil {
		t.Fatalf("Failed to generate CSRF token: %v", err)
	}

	// Set cookie manually for test
	req.Header.Set("Cookie", "csrf_token="+token)

	// Validate token
	err = sm.ValidateCSRFToken(ctx, token)
	if err != nil {
		t.Errorf("Expected no error for valid token, got: %v", err)
	}
}

func TestValidateCSRFToken_Empty(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.ValidateCSRFToken(ctx, "")
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

func TestValidateCSRFToken_Invalid(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{}
	ctx := createTestContext(t, req)

	err := sm.ValidateCSRFToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// Test input validation
func TestValidateInput_Clean(t *testing.T) {
	sm := createTestSecurityManager(t)

	rules := InputValidationRules{
		MaxLength: 100,
		AllowHTML: false,
	}

	err := sm.ValidateInput("clean input text", rules)
	if err != nil {
		t.Errorf("Expected no error for clean input, got: %v", err)
	}
}

func TestValidateInput_TooLong(t *testing.T) {
	sm := createTestSecurityManager(t)

	rules := InputValidationRules{
		MaxLength: 10,
	}

	err := sm.ValidateInput("this is a very long input text", rules)
	if err == nil {
		t.Error("Expected error for input exceeding max length")
	}
}

func TestValidateInput_HTMLNotAllowed(t *testing.T) {
	sm := createTestSecurityManager(t)

	rules := InputValidationRules{
		AllowHTML: false,
	}

	err := sm.ValidateInput("<script>alert('xss')</script>", rules)
	if err == nil {
		t.Error("Expected error for HTML when not allowed")
	}
}

func TestValidateInput_SQLInjection(t *testing.T) {
	sm := createTestSecurityManager(t)

	rules := InputValidationRules{}

	err := sm.ValidateInput("' OR '1'='1", rules)
	if err == nil {
		t.Error("Expected error for SQL injection pattern")
	}
}

func TestValidateInput_Pattern(t *testing.T) {
	sm := createTestSecurityManager(t)

	rules := InputValidationRules{
		Pattern: `^[a-zA-Z0-9]+$`,
	}

	err := sm.ValidateInput("validInput123", rules)
	if err != nil {
		t.Errorf("Expected no error for valid pattern, got: %v", err)
	}

	err = sm.ValidateInput("invalid input!", rules)
	if err == nil {
		t.Error("Expected error for input not matching pattern")
	}
}

// Test input sanitization
func TestSanitizeInput(t *testing.T) {
	sm := createTestSecurityManager(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML escaping",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "Null byte removal",
			input:    "test\x00data",
			expected: "testdata",
		},
		{
			name:     "Whitespace trimming",
			input:    "  test data  ",
			expected: "test data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Test cookie encryption
func TestEncryptDecryptCookie(t *testing.T) {
	sm := createTestSecurityManager(t)

	original := "sensitive-cookie-value"

	encrypted, err := sm.EncryptCookie(original)
	if err != nil {
		t.Fatalf("Failed to encrypt cookie: %v", err)
	}

	if encrypted == original {
		t.Error("Encrypted value should be different from original")
	}

	decrypted, err := sm.DecryptCookie(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt cookie: %v", err)
	}

	if decrypted != original {
		t.Errorf("Expected decrypted value %q, got %q", original, decrypted)
	}
}

func TestDecryptCookie_Invalid(t *testing.T) {
	sm := createTestSecurityManager(t)

	_, err := sm.DecryptCookie("invalid-encrypted-data")
	if err == nil {
		t.Error("Expected error for invalid encrypted data")
	}
}

// Test comprehensive request validation
func TestValidateRequest_AllChecks(t *testing.T) {
	sm := createTestSecurityManager(t)

	req := &Request{
		RequestURI: "/api/users",
		Header:     http.Header{},
		RawBody:    []byte("valid request body"),
	}
	req.Header.Set("Content-Length", "18")

	ctx := createTestContext(t, req)

	err := sm.ValidateRequest(ctx)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}
}

// Test helper functions
func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("Expected contains to return true for existing item")
	}

	if contains(slice, "orange") {
		t.Error("Expected contains to return false for non-existing item")
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"document.pdf", ".pdf"},
		{"image.jpg", ".jpg"},
		{"archive.tar.gz", ".gz"},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getFileExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestContainsHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"<div>test</div>", true},
		{"<script>alert('xss')</script>", true},
		{"plain text", false},
		{"text with < and > but not tags", true}, // This is detected as HTML for safety
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsHTML(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContainsSQLInjection(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"' OR '1'='1", true},
		{"'; DROP TABLE users", true},
		{"UNION SELECT * FROM passwords", true},
		{"normal query text", false},
		{"user@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsSQLInjection(tt.input)
			if result != tt.expected {
				t.Errorf("For input %q: expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

// Test CSRF token cleanup
func TestCleanupExpiredCSRFTokens(t *testing.T) {
	db := NewMockDatabaseManager()

	config := DefaultSecurityConfig()
	config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
	config.JWTSecret = "test-jwt-secret"
	config.CSRFTokenExpiry = 1 * time.Millisecond

	sm, err := NewSecurityManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	smImpl := sm.(*securityManagerImpl)

	// Add some tokens
	smImpl.csrfTokens["token1"] = time.Now().Add(-2 * time.Millisecond) // Expired
	smImpl.csrfTokens["token2"] = time.Now().Add(1 * time.Hour)         // Valid

	smImpl.CleanupExpiredCSRFTokens()

	if _, exists := smImpl.csrfTokens["token1"]; exists {
		t.Error("Expected expired token to be removed")
	}

	if _, exists := smImpl.csrfTokens["token2"]; !exists {
		t.Error("Expected valid token to remain")
	}
}
