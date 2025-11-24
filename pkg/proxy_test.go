package pkg

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// TestNewProxyManager tests proxy manager creation
func TestNewProxyManager(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()

	pm := NewProxyManager(config, cache)
	if pm == nil {
		t.Fatal("Expected proxy manager to be created")
	}

	// Test with nil config
	pm2 := NewProxyManager(nil, cache)
	if pm2 == nil {
		t.Fatal("Expected proxy manager to be created with default config")
	}
}

// TestAddBackend tests adding backends
func TestAddBackend(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	backendURL, _ := url.Parse("http://backend1.example.com")
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		Weight:   1,
		IsActive: true,
	}

	err := pm.AddBackend(backend)
	if err != nil {
		t.Fatalf("Failed to add backend: %v", err)
	}

	// Verify backend was added
	retrieved, err := pm.GetBackend("backend1")
	if err != nil {
		t.Fatalf("Failed to get backend: %v", err)
	}

	if retrieved.ID != "backend1" {
		t.Errorf("Expected backend ID 'backend1', got '%s'", retrieved.ID)
	}

	// Test adding backend with nil
	err = pm.AddBackend(nil)
	if err == nil {
		t.Error("Expected error when adding nil backend")
	}

	// Test adding backend with empty ID
	err = pm.AddBackend(&Backend{URL: backendURL})
	if err == nil {
		t.Error("Expected error when adding backend with empty ID")
	}

	// Test adding backend with nil URL
	err = pm.AddBackend(&Backend{ID: "backend2"})
	if err == nil {
		t.Error("Expected error when adding backend with nil URL")
	}
}

// TestRemoveBackend tests removing backends
func TestRemoveBackend(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	backendURL, _ := url.Parse("http://backend1.example.com")
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}

	pm.AddBackend(backend)

	err := pm.RemoveBackend("backend1")
	if err != nil {
		t.Fatalf("Failed to remove backend: %v", err)
	}

	// Verify backend was removed
	_, err = pm.GetBackend("backend1")
	if err == nil {
		t.Error("Expected error when getting removed backend")
	}

	// Test removing non-existent backend
	err = pm.RemoveBackend("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent backend")
	}
}

// TestListBackends tests listing all backends
func TestListBackends(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	// Add multiple backends
	for i := 1; i <= 3; i++ {
		backendURL, _ := url.Parse(fmt.Sprintf("http://backend%d.example.com", i))
		backend := &Backend{
			ID:       fmt.Sprintf("backend%d", i),
			URL:      backendURL,
			IsActive: true,
		}
		pm.AddBackend(backend)
	}

	backends := pm.ListBackends()
	if len(backends) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(backends))
	}
}

// TestRoundRobinLoadBalancer tests round-robin load balancing
func TestRoundRobinLoadBalancer(t *testing.T) {
	lb := NewRoundRobinLoadBalancer()

	backends := []*Backend{
		{ID: "backend1", URL: mustParseURL("http://backend1.example.com"), IsActive: true},
		{ID: "backend2", URL: mustParseURL("http://backend2.example.com"), IsActive: true},
		{ID: "backend3", URL: mustParseURL("http://backend3.example.com"), IsActive: true},
	}

	// Test round-robin distribution
	selectedIDs := make(map[string]int)
	for i := 0; i < 9; i++ {
		backend, err := lb.SelectBackend(backends)
		if err != nil {
			t.Fatalf("Failed to select backend: %v", err)
		}
		selectedIDs[backend.ID]++
	}

	// Each backend should be selected 3 times
	for _, backend := range backends {
		count := selectedIDs[backend.ID]
		if count != 3 {
			t.Errorf("Expected backend %s to be selected 3 times, got %d", backend.ID, count)
		}
	}

	// Test with empty backends
	_, err := lb.SelectBackend([]*Backend{})
	if err == nil {
		t.Error("Expected error when selecting from empty backends")
	}
}

// TestCircuitBreaker tests circuit breaker functionality
func TestCircuitBreaker(t *testing.T) {
	config := DefaultProxyConfig()
	config.CircuitBreakerThreshold = 3

	cb := NewCircuitBreaker(config)

	backendID := "backend1"

	// Initially closed
	if cb.IsOpen(backendID) {
		t.Error("Circuit should be closed initially")
	}

	if cb.GetState(backendID) != CircuitStateClosed {
		t.Error("Circuit state should be closed initially")
	}

	// Record failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure(backendID)
	}

	if !cb.IsOpen(backendID) {
		t.Error("Circuit should be open after threshold failures")
	}

	if cb.GetState(backendID) != CircuitStateOpen {
		t.Error("Circuit state should be open")
	}

	// Record success should not close circuit immediately
	cb.RecordSuccess(backendID)
	if !cb.IsOpen(backendID) {
		t.Error("Circuit should still be open after single success")
	}

	// Reset circuit
	cb.Reset(backendID)
	if cb.IsOpen(backendID) {
		t.Error("Circuit should be closed after reset")
	}
}

// TestConnectionPool tests connection pooling
func TestConnectionPool(t *testing.T) {
	config := DefaultProxyConfig()
	pool := NewConnectionPool(config)

	backendID := "backend1"

	// Get connection
	client1, err := pool.GetConnection(backendID)
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	if client1 == nil {
		t.Fatal("Expected client to be created")
	}

	// Get same connection again
	client2, err := pool.GetConnection(backendID)
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	if client1 != client2 {
		t.Error("Expected same client instance from pool")
	}

	// Get stats
	stats := pool.Stats()
	if stats.TotalConnections != 1 {
		t.Errorf("Expected 1 connection, got %d", stats.TotalConnections)
	}

	// Close pool
	err = pool.Close()
	if err != nil {
		t.Fatalf("Failed to close pool: %v", err)
	}
}

// TestProxyForward tests request forwarding
func TestProxyForward(t *testing.T) {
	// Create test backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Backend response"))
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()
	config.MaxRetries = 1
	pm := NewProxyManager(config, cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}
	pm.AddBackend(backend)

	// Create test request
	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	// Create mock context
	ctx := &mockContext{}

	// Forward request
	response, err := pm.Forward(ctx, request)
	if err != nil {
		t.Fatalf("Failed to forward request: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if string(response.Body) != "Backend response" {
		t.Errorf("Expected 'Backend response', got '%s'", string(response.Body))
	}

	// Check metrics
	metrics := pm.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}
}

// TestProxyForwardWithRetry tests request forwarding with retry
func TestProxyForwardWithRetry(t *testing.T) {
	attempts := 0
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()
	config.MaxRetries = 3
	config.RetryDelay = 10 * time.Millisecond
	pm := NewProxyManager(config, cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}
	pm.AddBackend(backend)

	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	ctx := &mockContext{}

	response, err := pm.Forward(ctx, request)
	if err != nil {
		t.Fatalf("Failed to forward request: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attempts)
	}

	// Check retry metrics
	metrics := pm.GetMetrics()
	if metrics.TotalRetries == 0 {
		t.Error("Expected retries to be recorded")
	}
}

// TestProxyHealthCheck tests health checking
func TestProxyHealthCheck(t *testing.T) {
	healthCheckCalled := false
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			healthCheckCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()
	config.HealthCheckEnabled = false // Disable automatic checks
	pm := NewProxyManager(config, cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:              "backend1",
		URL:             backendURL,
		IsActive:        true,
		HealthCheckPath: "/health",
	}
	pm.AddBackend(backend)

	// Perform manual health check
	err := pm.HealthCheck()
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if !healthCheckCalled {
		t.Error("Health check endpoint was not called")
	}

	// Check health status
	healthStatus := pm.GetHealthStatus()
	if len(healthStatus) != 1 {
		t.Errorf("Expected 1 health status, got %d", len(healthStatus))
	}

	status := healthStatus["backend1"]
	if status == nil {
		t.Fatal("Expected health status for backend1")
	}

	if !status.IsHealthy {
		t.Error("Expected backend to be healthy")
	}
}

// TestProxyCaching tests response caching
func TestProxyCaching(t *testing.T) {
	requestCount := 0
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Response %d", requestCount)))
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()
	config.CacheEnabled = true
	config.CacheTTL = 1 * time.Second
	pm := NewProxyManager(config, cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}
	pm.AddBackend(backend)

	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	ctx := &mockContext{}

	// First request - should hit backend
	response1, err := pm.Forward(ctx, request)
	if err != nil {
		t.Fatalf("Failed to forward request: %v", err)
	}

	// Second request - should hit cache
	response2, err := pm.Forward(ctx, request)
	if err != nil {
		t.Fatalf("Failed to forward request: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("Expected 1 backend request, got %d", requestCount)
	}

	if string(response1.Body) != string(response2.Body) {
		t.Error("Expected cached response to match original")
	}

	// Check cache metrics
	metrics := pm.GetMetrics()
	if metrics.CacheHits == 0 {
		t.Error("Expected cache hits to be recorded")
	}
}

// TestProxyMetrics tests metrics collection
func TestProxyMetrics(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	config := DefaultProxyConfig()
	config.CacheEnabled = false // Disable caching for accurate metrics
	pm := NewProxyManager(config, cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}
	pm.AddBackend(backend)

	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	ctx := &mockContext{}

	// Make multiple requests
	for i := 0; i < 5; i++ {
		pm.Forward(ctx, request)
	}

	metrics := pm.GetMetrics()
	if metrics.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 5 {
		t.Errorf("Expected 5 successful requests, got %d", metrics.SuccessfulRequests)
	}

	if metrics.AverageResponseTime == 0 {
		t.Error("Expected average response time to be calculated")
	}

	// Check backend metrics
	backendMetrics := metrics.BackendMetrics["backend1"]
	if backendMetrics == nil {
		t.Fatal("Expected backend metrics")
	}

	if backendMetrics.Requests != 5 {
		t.Errorf("Expected 5 backend requests, got %d", backendMetrics.Requests)
	}

	// Reset metrics
	pm.ResetMetrics()
	metrics = pm.GetMetrics()
	if metrics.TotalRequests != 0 {
		t.Error("Expected metrics to be reset")
	}
}

// TestProxyConcurrency tests concurrent request handling
func TestProxyConcurrency(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	backendURL, _ := url.Parse(backendServer.URL)
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: true,
	}
	pm.AddBackend(backend)

	reqURL, _ := url.Parse("http://example.com/test")

	var wg sync.WaitGroup
	concurrentRequests := 10

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			request := &Request{
				Method:     "GET",
				URL:        reqURL,
				Header:     make(http.Header),
				RequestURI: "/test",
			}

			ctx := &mockContext{}
			_, err := pm.Forward(ctx, request)
			if err != nil {
				t.Errorf("Failed to forward request: %v", err)
			}
		}()
	}

	wg.Wait()

	metrics := pm.GetMetrics()
	if metrics.TotalRequests != int64(concurrentRequests) {
		t.Errorf("Expected %d total requests, got %d", concurrentRequests, metrics.TotalRequests)
	}
}

// TestProxyNoBackends tests behavior with no backends
func TestProxyNoBackends(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	ctx := &mockContext{}

	_, err := pm.Forward(ctx, request)
	if err == nil {
		t.Error("Expected error when forwarding with no backends")
	}
}

// TestProxyInactiveBackend tests that inactive backends are not used
func TestProxyInactiveBackend(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	pm := NewProxyManager(DefaultProxyConfig(), cache)

	backendURL, _ := url.Parse("http://backend1.example.com")
	backend := &Backend{
		ID:       "backend1",
		URL:      backendURL,
		IsActive: false, // Inactive
	}
	pm.AddBackend(backend)

	reqURL, _ := url.Parse("http://example.com/test")
	request := &Request{
		Method:     "GET",
		URL:        reqURL,
		Header:     make(http.Header),
		RequestURI: "/test",
	}

	ctx := &mockContext{}

	_, err := pm.Forward(ctx, request)
	if err == nil {
		t.Error("Expected error when forwarding to inactive backend")
	}
}

// Helper functions

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
