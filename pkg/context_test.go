package pkg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	// Create a test request
	req := &Request{
		Method: "GET",
		URL: &url.URL{
			Path:     "/test",
			RawQuery: "param1=value1&param2=value2",
		},
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer token123"},
		},
		Params: map[string]string{
			"id": "123",
		},
		Query: map[string]string{
			"filter": "active",
		},
		ID:       "req-123",
		TenantID: "tenant-1",
		UserID:   "user-1",
	}

	// Create a test response writer
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)

	// Create context
	baseCtx := context.Background()
	ctx := NewContext(req, resp, baseCtx)

	// Test basic request/response access
	if ctx.Request() != req {
		t.Error("Request() should return the original request")
	}

	if ctx.Response() != resp {
		t.Error("Response() should return the original response writer")
	}

	if ctx.Context() != baseCtx {
		t.Error("Context() should return the base context")
	}
}

func TestContextParams(t *testing.T) {
	req := &Request{
		Params: map[string]string{
			"id":   "123",
			"name": "test",
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	params := ctx.Params()

	if params["id"] != "123" {
		t.Errorf("Expected param 'id' to be '123', got '%s'", params["id"])
	}

	if params["name"] != "test" {
		t.Errorf("Expected param 'name' to be 'test', got '%s'", params["name"])
	}
}

func TestContextQuery(t *testing.T) {
	u, _ := url.Parse("/test?param1=value1&param2=value2")
	req := &Request{
		URL: u,
		Query: map[string]string{
			"filter": "active",
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	query := ctx.Query()

	// Should have both URL query params and pre-set query params
	if query["param1"] != "value1" {
		t.Errorf("Expected query 'param1' to be 'value1', got '%s'", query["param1"])
	}

	if query["filter"] != "active" {
		t.Errorf("Expected query 'filter' to be 'active', got '%s'", query["filter"])
	}
}

func TestContextHeaders(t *testing.T) {
	req := &Request{
		Header: http.Header{
			"Content-Type":    []string{"application/json"},
			"Authorization":   []string{"Bearer token123"},
			"X-Custom-Header": []string{"custom-value"},
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	headers := ctx.Headers()

	// Headers should be normalized to lowercase
	if headers["content-type"] != "application/json" {
		t.Errorf("Expected header 'content-type' to be 'application/json', got '%s'", headers["content-type"])
	}

	if headers["authorization"] != "Bearer token123" {
		t.Errorf("Expected header 'authorization' to be 'Bearer token123', got '%s'", headers["authorization"])
	}

	if headers["x-custom-header"] != "custom-value" {
		t.Errorf("Expected header 'x-custom-header' to be 'custom-value', got '%s'", headers["x-custom-header"])
	}
}

func TestContextGetHeader(t *testing.T) {
	req := &Request{
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer token123"},
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Test case-insensitive header access
	if ctx.GetHeader("content-type") != "application/json" {
		t.Error("GetHeader should be case-insensitive")
	}

	if ctx.GetHeader("Content-Type") != "application/json" {
		t.Error("GetHeader should be case-insensitive")
	}

	if ctx.GetHeader("AUTHORIZATION") != "Bearer token123" {
		t.Error("GetHeader should be case-insensitive")
	}
}

func TestContextWithTimeout(t *testing.T) {
	req := &Request{ID: "test-req"}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Create context with timeout
	timeoutCtx := ctx.WithTimeout(5 * time.Second)

	// Should have the same request
	if timeoutCtx.Request().ID != "test-req" {
		t.Error("WithTimeout should preserve request data")
	}

	// Context should have timeout
	deadline, ok := timeoutCtx.Context().Deadline()
	if !ok {
		t.Error("WithTimeout should set a deadline")
	}

	if time.Until(deadline) > 6*time.Second {
		t.Error("Timeout should be approximately 5 seconds")
	}
}

func TestContextWithCancel(t *testing.T) {
	req := &Request{ID: "test-req"}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Create context with cancel
	cancelCtx, cancel := ctx.WithCancel()

	// Should have the same request
	if cancelCtx.Request().ID != "test-req" {
		t.Error("WithCancel should preserve request data")
	}

	// Context should not be cancelled initially
	select {
	case <-cancelCtx.Context().Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Expected
	}

	// Cancel the context
	cancel()

	// Context should now be cancelled
	select {
	case <-cancelCtx.Context().Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after calling cancel()")
	}
}

func TestContextResponseHelpers(t *testing.T) {
	req := &Request{}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Test JSON response
	data := map[string]string{"message": "hello"}
	err := ctx.JSON(200, data)
	if err != nil {
		t.Errorf("JSON response failed: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "hello") {
		t.Error("Response body should contain 'hello'")
	}
}

func TestContextSetGetCookie(t *testing.T) {
	req := &Request{
		Header: http.Header{
			"Cookie": []string{"session=abc123; user=john"},
		},
	}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Test getting cookie
	cookie, err := ctx.GetCookie("session")
	if err != nil {
		t.Errorf("Failed to get cookie: %v", err)
	}

	if cookie.Name != "session" || cookie.Value != "abc123" {
		t.Errorf("Expected cookie session=abc123, got %s=%s", cookie.Name, cookie.Value)
	}

	// Test setting cookie
	newCookie := &Cookie{
		Name:  "new_session",
		Value: "xyz789",
		Path:  "/",
	}

	err = ctx.SetCookie(newCookie)
	if err != nil {
		t.Errorf("Failed to set cookie: %v", err)
	}
}

func TestContextFormValue(t *testing.T) {
	req := &Request{
		Form: map[string]string{
			"username": "john",
			"email":    "john@example.com",
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	if ctx.FormValue("username") != "john" {
		t.Error("FormValue should return correct form value")
	}

	if ctx.FormValue("email") != "john@example.com" {
		t.Error("FormValue should return correct form value")
	}

	if ctx.FormValue("nonexistent") != "" {
		t.Error("FormValue should return empty string for non-existent fields")
	}
}

func TestContextAuthentication(t *testing.T) {
	req := &Request{}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Initially not authenticated
	if ctx.IsAuthenticated() {
		t.Error("Context should not be authenticated initially")
	}

	// Set user (simulating authentication middleware)
	user := &User{
		ID:       "user-123",
		Username: "john",
	}

	if ctxImpl, ok := ctx.(*contextImpl); ok {
		ctxImpl.SetUser(user)
	}

	// Now should be authenticated
	if !ctx.IsAuthenticated() {
		t.Error("Context should be authenticated after setting user")
	}

	if ctx.User().ID != "user-123" {
		t.Error("User ID should match")
	}
}

func TestContextHelperMethods(t *testing.T) {
	req := &Request{
		Params: map[string]string{
			"id":    "123",
			"count": "45",
		},
		Query: map[string]string{
			"page":   "2",
			"active": "true",
		},
	}

	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	ctxImpl := ctx.(*contextImpl)

	// Test GetParam
	if ctxImpl.GetParam("id") != "123" {
		t.Error("GetParam should return correct parameter")
	}

	// Test GetParamInt
	count, err := ctxImpl.GetParamInt("count")
	if err != nil || count != 45 {
		t.Error("GetParamInt should return correct integer parameter")
	}

	// Test GetQueryParam
	if ctxImpl.GetQueryParam("page") != "2" {
		t.Error("GetQueryParam should return correct query parameter")
	}

	// Test GetQueryParamBool
	active, err := ctxImpl.GetQueryParamBool("active")
	if err != nil || !active {
		t.Error("GetQueryParamBool should return correct boolean parameter")
	}
}
