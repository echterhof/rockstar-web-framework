package pkg

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestCookieManager_SetCookie tests setting a cookie
func TestCookieManager_SetCookie(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create cookie manager
	config := DefaultCookieConfig()
	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Set a cookie
	cookie := &Cookie{
		Name:  "test_cookie",
		Value: "test_value",
	}

	err = cm.SetCookie(ctx, cookie)
	if err != nil {
		t.Fatalf("Failed to set cookie: %v", err)
	}

	// Verify cookie was set in response
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set in response")
	}

	found := false
	for _, c := range cookies {
		if c.Name == "test_cookie" && c.Value == "test_value" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Cookie not found in response")
	}
}

// TestCookieManager_GetCookie tests retrieving a cookie
func TestCookieManager_GetCookie(t *testing.T) {
	// Create test context with cookie
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "test_cookie",
		Value: "test_value",
	})

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := &contextImpl{
		request:  request,
		response: resp,
		httpReq:  req,
		params:   make(map[string]string),
		query:    make(map[string]string),
		headers:  make(map[string]string),
	}

	// Create cookie manager
	config := DefaultCookieConfig()
	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Get the cookie
	cookie, err := cm.GetCookie(ctx, "test_cookie")
	if err != nil {
		t.Fatalf("Failed to get cookie: %v", err)
	}

	if cookie.Name != "test_cookie" {
		t.Errorf("Expected cookie name 'test_cookie', got '%s'", cookie.Name)
	}

	if cookie.Value != "test_value" {
		t.Errorf("Expected cookie value 'test_value', got '%s'", cookie.Value)
	}
}

// TestCookieManager_GetAllCookies tests retrieving all cookies
func TestCookieManager_GetAllCookies(t *testing.T) {
	// Create test context with multiple cookies
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "cookie1", Value: "value1"})
	req.AddCookie(&http.Cookie{Name: "cookie2", Value: "value2"})
	req.AddCookie(&http.Cookie{Name: "cookie3", Value: "value3"})

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := &contextImpl{
		request:  request,
		response: resp,
		httpReq:  req,
		params:   make(map[string]string),
		query:    make(map[string]string),
		headers:  make(map[string]string),
	}

	// Create cookie manager
	config := DefaultCookieConfig()
	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Get all cookies
	cookies, err := cm.GetAllCookies(ctx)
	if err != nil {
		t.Fatalf("Failed to get all cookies: %v", err)
	}

	if len(cookies) != 3 {
		t.Errorf("Expected 3 cookies, got %d", len(cookies))
	}

	// Verify cookie names
	names := make(map[string]bool)
	for _, cookie := range cookies {
		names[cookie.Name] = true
	}

	if !names["cookie1"] || !names["cookie2"] || !names["cookie3"] {
		t.Error("Not all cookies were retrieved")
	}
}

// TestCookieManager_DeleteCookie tests deleting a cookie
func TestCookieManager_DeleteCookie(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create cookie manager
	config := DefaultCookieConfig()
	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Delete a cookie
	err = cm.DeleteCookie(ctx, "test_cookie")
	if err != nil {
		t.Fatalf("Failed to delete cookie: %v", err)
	}

	// Verify cookie was set with MaxAge=-1
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set in response")
	}

	found := false
	for _, c := range cookies {
		if c.Name == "test_cookie" && c.MaxAge == -1 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Delete cookie not found in response")
	}
}

// TestCookieManager_EncryptDecrypt tests cookie encryption and decryption
func TestCookieManager_EncryptDecrypt(t *testing.T) {
	// Generate encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create cookie manager with encryption
	config := &CookieConfig{
		EncryptionKey: key,
		DefaultPath:   "/",
	}

	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Test encryption and decryption
	originalValue := "sensitive_data_12345"

	encrypted, err := cm.EncryptValue(originalValue)
	if err != nil {
		t.Fatalf("Failed to encrypt value: %v", err)
	}

	if encrypted == originalValue {
		t.Error("Encrypted value should be different from original")
	}

	decrypted, err := cm.DecryptValue(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt value: %v", err)
	}

	if decrypted != originalValue {
		t.Errorf("Expected decrypted value '%s', got '%s'", originalValue, decrypted)
	}
}

// TestCookieManager_SetEncryptedCookie tests setting an encrypted cookie
func TestCookieManager_SetEncryptedCookie(t *testing.T) {
	// Generate encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create cookie manager with encryption
	config := &CookieConfig{
		EncryptionKey: key,
		DefaultPath:   "/",
	}

	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Set encrypted cookie
	cookie := &Cookie{
		Name:  "encrypted_cookie",
		Value: "sensitive_value",
	}

	err = cm.SetEncryptedCookie(ctx, cookie)
	if err != nil {
		t.Fatalf("Failed to set encrypted cookie: %v", err)
	}

	// Verify cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set in response")
	}

	found := false
	for _, c := range cookies {
		if c.Name == "encrypted_cookie" {
			found = true
			// Value should be encrypted (different from original)
			if c.Value == "sensitive_value" {
				t.Error("Cookie value should be encrypted")
			}
			break
		}
	}

	if !found {
		t.Error("Encrypted cookie not found in response")
	}
}

// TestCookieManager_GetEncryptedCookie tests retrieving an encrypted cookie
func TestCookieManager_GetEncryptedCookie(t *testing.T) {
	// Generate encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create cookie manager
	config := &CookieConfig{
		EncryptionKey: key,
		DefaultPath:   "/",
	}

	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Encrypt a value
	originalValue := "sensitive_data"
	encryptedValue, err := cm.EncryptValue(originalValue)
	if err != nil {
		t.Fatalf("Failed to encrypt value: %v", err)
	}

	// Create test context with encrypted cookie
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "encrypted_cookie",
		Value: encryptedValue,
	})

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := &contextImpl{
		request:  request,
		response: resp,
		httpReq:  req,
		params:   make(map[string]string),
		query:    make(map[string]string),
		headers:  make(map[string]string),
	}

	// Get and decrypt the cookie
	cookie, err := cm.GetEncryptedCookie(ctx, "encrypted_cookie")
	if err != nil {
		t.Fatalf("Failed to get encrypted cookie: %v", err)
	}

	if cookie.Value != originalValue {
		t.Errorf("Expected decrypted value '%s', got '%s'", originalValue, cookie.Value)
	}
}

// TestCookieManager_ApplyDefaults tests default configuration application
func TestCookieManager_ApplyDefaults(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create cookie manager with custom defaults
	config := &CookieConfig{
		DefaultPath:     "/custom",
		DefaultDomain:   "example.com",
		DefaultSecure:   true,
		DefaultHTTPOnly: true,
		DefaultSameSite: http.SameSiteStrictMode,
		DefaultMaxAge:   3600,
	}

	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Set a cookie without specifying defaults
	cookie := &Cookie{
		Name:  "test_cookie",
		Value: "test_value",
	}

	err = cm.SetCookie(ctx, cookie)
	if err != nil {
		t.Fatalf("Failed to set cookie: %v", err)
	}

	// Verify defaults were applied
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set in response")
	}

	c := cookies[0]
	if c.Path != "/custom" {
		t.Errorf("Expected path '/custom', got '%s'", c.Path)
	}

	if c.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", c.Domain)
	}

	if !c.Secure {
		t.Error("Expected Secure flag to be true")
	}

	if !c.HttpOnly {
		t.Error("Expected HttpOnly flag to be true")
	}

	if c.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite Strict, got %v", c.SameSite)
	}
}

// TestHeaderManager_SetGetHeader tests setting and getting headers
func TestHeaderManager_SetGetHeader(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Custom-Header", "custom_value")

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Get request header
	value := hm.GetHeader(ctx, "X-Custom-Header")
	if value != "custom_value" {
		t.Errorf("Expected header value 'custom_value', got '%s'", value)
	}

	// Set response header
	err := hm.SetHeader(ctx, "X-Response-Header", "response_value")
	if err != nil {
		t.Fatalf("Failed to set header: %v", err)
	}

	// Verify response header was set
	if w.Header().Get("X-Response-Header") != "response_value" {
		t.Error("Response header not set correctly")
	}
}

// TestHeaderManager_GetAllHeaders tests retrieving all headers
func TestHeaderManager_GetAllHeaders(t *testing.T) {
	// Create test context with multiple headers
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Header-1", "value1")
	req.Header.Set("X-Header-2", "value2")
	req.Header.Set("User-Agent", "test-agent")

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Get all headers
	headers := hm.GetAllHeaders(ctx)

	if len(headers) == 0 {
		t.Fatal("No headers retrieved")
	}

	// Verify headers (note: keys are lowercase in context)
	if headers["x-header-1"] != "value1" {
		t.Error("Header X-Header-1 not found or incorrect")
	}

	if headers["x-header-2"] != "value2" {
		t.Error("Header X-Header-2 not found or incorrect")
	}

	if headers["user-agent"] != "test-agent" {
		t.Error("Header User-Agent not found or incorrect")
	}
}

// TestHeaderManager_SetHeaders tests setting multiple headers
func TestHeaderManager_SetHeaders(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Set multiple headers
	headers := map[string]string{
		"X-Header-1": "value1",
		"X-Header-2": "value2",
		"X-Header-3": "value3",
	}

	err := hm.SetHeaders(ctx, headers)
	if err != nil {
		t.Fatalf("Failed to set headers: %v", err)
	}

	// Verify all headers were set
	for key, expectedValue := range headers {
		actualValue := w.Header().Get(key)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s='%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}

// TestHeaderManager_CommonHelpers tests common header helper methods
func TestHeaderManager_CommonHelpers(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://example.com")
	req.Header.Set("Content-Type", "application/json")

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Test GetAuthorization
	auth := hm.GetAuthorization(ctx)
	if auth != "Bearer token123" {
		t.Errorf("Expected Authorization 'Bearer token123', got '%s'", auth)
	}

	// Test GetUserAgent
	ua := hm.GetUserAgent(ctx)
	if ua != "Mozilla/5.0" {
		t.Errorf("Expected User-Agent 'Mozilla/5.0', got '%s'", ua)
	}

	// Test GetReferer
	referer := hm.GetReferer(ctx)
	if referer != "https://example.com" {
		t.Errorf("Expected Referer 'https://example.com', got '%s'", referer)
	}

	// Test GetContentType
	ct := hm.GetContentType(ctx)
	if ct != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", ct)
	}

	// Test SetContentType
	err := hm.SetContentType(ctx, "text/html")
	if err != nil {
		t.Fatalf("Failed to set Content-Type: %v", err)
	}

	if w.Header().Get("Content-Type") != "text/html" {
		t.Error("Content-Type not set correctly")
	}

	// Test SetCacheControl
	err = hm.SetCacheControl(ctx, "no-cache")
	if err != nil {
		t.Fatalf("Failed to set Cache-Control: %v", err)
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Cache-Control not set correctly")
	}

	// Test SetLocation
	err = hm.SetLocation(ctx, "https://redirect.com")
	if err != nil {
		t.Fatalf("Failed to set Location: %v", err)
	}

	if w.Header().Get("Location") != "https://redirect.com" {
		t.Error("Location not set correctly")
	}
}

// TestHeaderManager_HasHeader tests checking header existence
func TestHeaderManager_HasHeader(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Existing-Header", "value")

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Test existing header
	if !hm.HasHeader(ctx, "X-Existing-Header") {
		t.Error("Expected header to exist")
	}

	// Test non-existing header
	if hm.HasHeader(ctx, "X-Non-Existing-Header") {
		t.Error("Expected header to not exist")
	}
}

// TestHeaderManager_DeleteHeader tests deleting a header
func TestHeaderManager_DeleteHeader(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Create header manager
	hm := NewHeaderManager()

	// Set a header
	err := hm.SetHeader(ctx, "X-Test-Header", "value")
	if err != nil {
		t.Fatalf("Failed to set header: %v", err)
	}

	// Verify header was set
	if w.Header().Get("X-Test-Header") != "value" {
		t.Error("Header not set correctly")
	}

	// Delete the header
	err = hm.DeleteHeader(ctx, "X-Test-Header")
	if err != nil {
		t.Fatalf("Failed to delete header: %v", err)
	}

	// Verify header was deleted
	if w.Header().Get("X-Test-Header") != "" {
		t.Error("Header not deleted")
	}
}

// TestContext_CookieIntegration tests cookie operations through context
func TestContext_CookieIntegration(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "abc123",
	})

	request := &Request{
		Method: "GET",
		URL:    req.URL,
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := &contextImpl{
		request:  request,
		response: resp,
		httpReq:  req,
		params:   make(map[string]string),
		query:    make(map[string]string),
		headers:  make(map[string]string),
	}

	// Get cookie through context
	cookie, err := ctx.GetCookie("session_id")
	if err != nil {
		t.Fatalf("Failed to get cookie: %v", err)
	}

	if cookie.Value != "abc123" {
		t.Errorf("Expected cookie value 'abc123', got '%s'", cookie.Value)
	}

	// Set cookie through context
	newCookie := &Cookie{
		Name:     "new_cookie",
		Value:    "new_value",
		Path:     "/",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
	}

	err = ctx.SetCookie(newCookie)
	if err != nil {
		t.Fatalf("Failed to set cookie: %v", err)
	}

	// Verify cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "new_cookie" && c.Value == "new_value" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Cookie not set through context")
	}
}

// TestContext_HeaderIntegration tests header operations through context
func TestContext_HeaderIntegration(t *testing.T) {
	// Create test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "req-123")

	request := &Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Header: req.Header,
	}

	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	// Get header through context
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got '%s'", requestID)
	}

	// Set header through context
	ctx.SetHeader("X-Response-ID", "resp-456")

	// Verify header was set
	if w.Header().Get("X-Response-ID") != "resp-456" {
		t.Error("Header not set through context")
	}
}

// TestCookieManager_ErrorHandling tests error handling
func TestCookieManager_ErrorHandling(t *testing.T) {
	config := DefaultCookieConfig()
	cm, err := NewCookieManager(config)
	if err != nil {
		t.Fatalf("Failed to create cookie manager: %v", err)
	}

	// Test with nil context
	err = cm.SetCookie(nil, &Cookie{Name: "test", Value: "value"})
	if err == nil {
		t.Error("Expected error with nil context")
	}

	// Test with nil cookie
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	request := &Request{Method: "GET", URL: req.URL, Header: req.Header}
	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	err = cm.SetCookie(ctx, nil)
	if err == nil {
		t.Error("Expected error with nil cookie")
	}

	// Test encryption without key
	cmNoKey, _ := NewCookieManager(&CookieConfig{})
	_, err = cmNoKey.EncryptValue("test")
	if err == nil {
		t.Error("Expected error when encrypting without key")
	}
}

// TestHeaderManager_ErrorHandling tests error handling
func TestHeaderManager_ErrorHandling(t *testing.T) {
	hm := NewHeaderManager()

	// Test with nil context
	err := hm.SetHeader(nil, "X-Test", "value")
	if err == nil {
		t.Error("Expected error with nil context")
	}

	// Test with empty key
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	request := &Request{Method: "GET", URL: req.URL, Header: req.Header}
	resp := NewResponseWriter(w)
	ctx := NewContext(request, resp, req.Context())

	err = hm.SetHeader(ctx, "", "value")
	if err == nil {
		t.Error("Expected error with empty header key")
	}
}
