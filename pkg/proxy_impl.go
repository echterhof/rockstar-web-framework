package pkg

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// proxyManager implements the ProxyManager interface
type proxyManager struct {
	backends   map[string]*Backend
	backendsMu sync.RWMutex

	loadBalancer   LoadBalancer
	circuitBreaker CircuitBreaker
	connectionPool ConnectionPool

	config  *ProxyConfig
	metrics *ProxyMetrics

	healthStatus map[string]*BackendHealth
	healthMu     sync.RWMutex

	cache CacheManager

	stopHealthCheck chan struct{}
	healthCheckWg   sync.WaitGroup
}

// NewProxyManager creates a new proxy manager with the given configuration
func NewProxyManager(config *ProxyConfig, cache CacheManager) ProxyManager {
	if config == nil {
		config = DefaultProxyConfig()
	}

	pm := &proxyManager{
		backends:        make(map[string]*Backend),
		config:          config,
		metrics:         &ProxyMetrics{BackendMetrics: make(map[string]*BackendMetrics)},
		healthStatus:    make(map[string]*BackendHealth),
		cache:           cache,
		stopHealthCheck: make(chan struct{}),
	}

	// Initialize load balancer
	pm.loadBalancer = NewRoundRobinLoadBalancer()

	// Initialize circuit breaker
	pm.circuitBreaker = NewCircuitBreaker(config)

	// Initialize connection pool
	pm.connectionPool = NewConnectionPool(config)

	// Start health checks if enabled
	if config.HealthCheckEnabled {
		pm.startHealthChecks()
	}

	return pm
}

// AddBackend adds a new backend server
func (pm *proxyManager) AddBackend(backend *Backend) error {
	if backend == nil {
		return errors.New("backend cannot be nil")
	}

	if backend.ID == "" {
		return errors.New("backend ID cannot be empty")
	}

	if backend.URL == nil {
		return errors.New("backend URL cannot be nil")
	}

	pm.backendsMu.Lock()
	defer pm.backendsMu.Unlock()

	backend.CreatedAt = time.Now()
	backend.UpdatedAt = time.Now()

	if backend.Weight <= 0 {
		backend.Weight = 1
	}

	pm.backends[backend.ID] = backend

	// Initialize health status
	pm.healthMu.Lock()
	pm.healthStatus[backend.ID] = &BackendHealth{
		BackendID: backend.ID,
		IsHealthy: true,
		LastCheck: time.Now(),
	}
	pm.healthMu.Unlock()

	// Initialize backend metrics
	pm.metrics.mu.Lock()
	pm.metrics.BackendMetrics[backend.ID] = &BackendMetrics{}
	pm.metrics.mu.Unlock()

	return nil
}

// RemoveBackend removes a backend server
func (pm *proxyManager) RemoveBackend(backendID string) error {
	pm.backendsMu.Lock()
	defer pm.backendsMu.Unlock()

	if _, exists := pm.backends[backendID]; !exists {
		return fmt.Errorf("backend %s not found", backendID)
	}

	delete(pm.backends, backendID)

	pm.healthMu.Lock()
	delete(pm.healthStatus, backendID)
	pm.healthMu.Unlock()

	return nil
}

// GetBackend retrieves a backend by ID
func (pm *proxyManager) GetBackend(backendID string) (*Backend, error) {
	pm.backendsMu.RLock()
	defer pm.backendsMu.RUnlock()

	backend, exists := pm.backends[backendID]
	if !exists {
		return nil, fmt.Errorf("backend %s not found", backendID)
	}

	return backend, nil
}

// ListBackends returns all backends
func (pm *proxyManager) ListBackends() []*Backend {
	pm.backendsMu.RLock()
	defer pm.backendsMu.RUnlock()

	backends := make([]*Backend, 0, len(pm.backends))
	for _, backend := range pm.backends {
		backends = append(backends, backend)
	}

	return backends
}

// Forward forwards a request to a backend server
func (pm *proxyManager) Forward(ctx Context, request *Request) (*Response, error) {
	startTime := time.Now()

	// Get available backends
	availableBackends := pm.getAvailableBackends()
	if len(availableBackends) == 0 {
		return nil, errors.New("no available backends")
	}

	// Try forwarding with retries
	var lastErr error
	for attempt := 0; attempt <= pm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Record retry
			atomic.AddInt64(&pm.metrics.TotalRetries, 1)

			// Apply retry delay with optional backoff
			delay := pm.config.RetryDelay
			if pm.config.RetryBackoff {
				delay = time.Duration(math.Pow(2, float64(attempt-1))) * pm.config.RetryDelay
			}
			time.Sleep(delay)
		}

		// Select backend using load balancer
		backend, err := pm.loadBalancer.SelectBackend(availableBackends)
		if err != nil {
			lastErr = err
			continue
		}

		// Check circuit breaker
		if pm.circuitBreaker.IsOpen(backend.ID) {
			lastErr = fmt.Errorf("circuit breaker open for backend %s", backend.ID)
			continue
		}

		// Try to get from cache if enabled
		if pm.config.CacheEnabled && request.Method == "GET" {
			cacheKey := pm.getCacheKey(backend.ID, request)
			if cached, err := pm.cache.Get(cacheKey); err == nil && cached != nil {
				atomic.AddInt64(&pm.metrics.CacheHits, 1)
				if response, ok := cached.(*Response); ok {
					return response, nil
				}
			}
			atomic.AddInt64(&pm.metrics.CacheMisses, 1)
		}

		// Forward request
		response, err := pm.forwardToBackend(ctx, backend, request)

		responseTime := time.Since(startTime)

		if err != nil {
			// Record failure
			pm.circuitBreaker.RecordFailure(backend.ID)
			pm.loadBalancer.UpdateBackend(backend.ID, false, responseTime)
			pm.updateBackendHealth(backend.ID, false, err.Error())
			pm.recordMetrics(backend.ID, false, responseTime)

			lastErr = err
			continue
		}

		// Record success
		pm.circuitBreaker.RecordSuccess(backend.ID)
		pm.loadBalancer.UpdateBackend(backend.ID, true, responseTime)
		pm.updateBackendHealth(backend.ID, true, "")
		pm.recordMetrics(backend.ID, true, responseTime)

		// Cache successful GET responses
		if pm.config.CacheEnabled && request.Method == "GET" && response.StatusCode == 200 {
			cacheKey := pm.getCacheKey(backend.ID, request)
			pm.cache.Set(cacheKey, response, pm.config.CacheTTL)
		}

		return response, nil
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// ForwardHTTP forwards a standard HTTP request
func (pm *proxyManager) ForwardHTTP(ctx context.Context, req *http.Request) (*http.Response, error) {
	availableBackends := pm.getAvailableBackends()
	if len(availableBackends) == 0 {
		return nil, errors.New("no available backends")
	}

	backend, err := pm.loadBalancer.SelectBackend(availableBackends)
	if err != nil {
		return nil, err
	}

	// Get connection from pool
	client, err := pm.connectionPool.GetConnection(backend.ID)
	if err != nil {
		return nil, err
	}

	// Update request URL to point to backend
	req.URL.Scheme = backend.URL.Scheme
	req.URL.Host = backend.URL.Host

	// Forward request
	resp, err := client.Do(req)
	if err != nil {
		pm.circuitBreaker.RecordFailure(backend.ID)
		return nil, err
	}

	pm.circuitBreaker.RecordSuccess(backend.ID)
	return resp, nil
}

// forwardToBackend forwards a request to a specific backend
func (pm *proxyManager) forwardToBackend(ctx Context, backend *Backend, request *Request) (*Response, error) {
	// Create HTTP request
	targetURL := *backend.URL
	targetURL.Path = request.URL.Path
	targetURL.RawQuery = request.URL.RawQuery

	httpReq, err := http.NewRequest(request.Method, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range request.Header {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	// Set timeout
	httpCtx, cancel := context.WithTimeout(context.Background(), pm.config.RequestTimeout)
	defer cancel()
	httpReq = httpReq.WithContext(httpCtx)

	// Get connection from pool
	client, err := pm.connectionPool.GetConnection(backend.ID)
	if err != nil {
		return nil, err
	}

	// Execute request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// Create response
	response := &Response{
		StatusCode: httpResp.StatusCode,
		Header:     httpResp.Header,
		Body:       body,
		Size:       int64(len(body)),
	}

	// Treat 5xx status codes as errors for retry logic
	if httpResp.StatusCode >= 500 {
		return response, fmt.Errorf("backend returned error status: %d", httpResp.StatusCode)
	}

	return response, nil
}

// getAvailableBackends returns backends that are active and healthy
func (pm *proxyManager) getAvailableBackends() []*Backend {
	pm.backendsMu.RLock()
	defer pm.backendsMu.RUnlock()

	pm.healthMu.RLock()
	defer pm.healthMu.RUnlock()

	available := make([]*Backend, 0)
	for _, backend := range pm.backends {
		if !backend.IsActive {
			continue
		}

		// Check circuit breaker
		if pm.circuitBreaker.IsOpen(backend.ID) {
			continue
		}

		// Check health status
		if health, exists := pm.healthStatus[backend.ID]; exists && !health.IsHealthy {
			continue
		}

		available = append(available, backend)
	}

	return available
}

// SetLoadBalancer sets the load balancer
func (pm *proxyManager) SetLoadBalancer(lb LoadBalancer) error {
	if lb == nil {
		return errors.New("load balancer cannot be nil")
	}
	pm.loadBalancer = lb
	return nil
}

// GetLoadBalancer returns the current load balancer
func (pm *proxyManager) GetLoadBalancer() LoadBalancer {
	return pm.loadBalancer
}

// SetCircuitBreaker sets the circuit breaker
func (pm *proxyManager) SetCircuitBreaker(cb CircuitBreaker) error {
	if cb == nil {
		return errors.New("circuit breaker cannot be nil")
	}
	pm.circuitBreaker = cb
	return nil
}

// GetCircuitBreaker returns the current circuit breaker
func (pm *proxyManager) GetCircuitBreaker() CircuitBreaker {
	return pm.circuitBreaker
}

// SetConnectionPool sets the connection pool
func (pm *proxyManager) SetConnectionPool(pool ConnectionPool) error {
	if pool == nil {
		return errors.New("connection pool cannot be nil")
	}
	pm.connectionPool = pool
	return nil
}

// GetConnectionPool returns the current connection pool
func (pm *proxyManager) GetConnectionPool() ConnectionPool {
	return pm.connectionPool
}

// HealthCheck performs health checks on all backends
func (pm *proxyManager) HealthCheck() error {
	pm.backendsMu.RLock()
	backends := make([]*Backend, 0, len(pm.backends))
	for _, backend := range pm.backends {
		backends = append(backends, backend)
	}
	pm.backendsMu.RUnlock()

	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *Backend) {
			defer wg.Done()
			pm.checkBackendHealth(b)
		}(backend)
	}

	wg.Wait()
	return nil
}

// checkBackendHealth checks the health of a single backend
func (pm *proxyManager) checkBackendHealth(backend *Backend) {
	startTime := time.Now()

	healthPath := backend.HealthCheckPath
	if healthPath == "" {
		healthPath = pm.config.HealthCheckPath
	}

	healthURL := *backend.URL
	healthURL.Path = healthPath

	client := &http.Client{
		Timeout: pm.config.HealthCheckTimeout,
	}

	resp, err := client.Get(healthURL.String())
	responseTime := time.Since(startTime)

	isHealthy := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300

	if resp != nil {
		resp.Body.Close()
	}

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	pm.healthMu.Lock()
	health := pm.healthStatus[backend.ID]
	if health == nil {
		health = &BackendHealth{BackendID: backend.ID}
		pm.healthStatus[backend.ID] = health
	}

	health.LastCheck = time.Now()
	health.ResponseTime = responseTime
	health.IsHealthy = isHealthy
	health.ErrorMessage = errorMsg

	if isHealthy {
		health.LastSuccess = time.Now()
		health.ConsecutiveFails = 0
	} else {
		health.LastFailure = time.Now()
		health.ConsecutiveFails++
	}
	pm.healthMu.Unlock()
}

// startHealthChecks starts periodic health checks
func (pm *proxyManager) startHealthChecks() {
	pm.healthCheckWg.Add(1)
	go func() {
		defer pm.healthCheckWg.Done()

		ticker := time.NewTicker(pm.config.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pm.HealthCheck()
			case <-pm.stopHealthCheck:
				return
			}
		}
	}()
}

// GetHealthStatus returns health status for all backends
func (pm *proxyManager) GetHealthStatus() map[string]*BackendHealth {
	pm.healthMu.RLock()
	defer pm.healthMu.RUnlock()

	status := make(map[string]*BackendHealth)
	for id, health := range pm.healthStatus {
		status[id] = health
	}

	return status
}

// updateBackendHealth updates backend health status
func (pm *proxyManager) updateBackendHealth(backendID string, isHealthy bool, errorMsg string) {
	pm.healthMu.Lock()
	defer pm.healthMu.Unlock()

	health := pm.healthStatus[backendID]
	if health == nil {
		health = &BackendHealth{BackendID: backendID}
		pm.healthStatus[backendID] = health
	}

	health.IsHealthy = isHealthy
	health.LastCheck = time.Now()

	if isHealthy {
		health.LastSuccess = time.Now()
		health.ConsecutiveFails = 0
		health.ErrorMessage = ""
	} else {
		health.LastFailure = time.Now()
		health.ConsecutiveFails++
		health.ErrorMessage = errorMsg
	}
}

// recordMetrics records request metrics
func (pm *proxyManager) recordMetrics(backendID string, success bool, responseTime time.Duration) {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	atomic.AddInt64(&pm.metrics.TotalRequests, 1)

	if success {
		atomic.AddInt64(&pm.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&pm.metrics.FailedRequests, 1)
	}

	pm.metrics.TotalResponseTime += responseTime
	if pm.metrics.TotalRequests > 0 {
		pm.metrics.AverageResponseTime = pm.metrics.TotalResponseTime / time.Duration(pm.metrics.TotalRequests)
	}

	// Update backend metrics
	backendMetrics := pm.metrics.BackendMetrics[backendID]
	if backendMetrics == nil {
		backendMetrics = &BackendMetrics{}
		pm.metrics.BackendMetrics[backendID] = backendMetrics
	}

	backendMetrics.Requests++
	if success {
		backendMetrics.SuccessfulRequests++
	} else {
		backendMetrics.FailedRequests++
	}
	backendMetrics.TotalResponseTime += responseTime
	if backendMetrics.Requests > 0 {
		backendMetrics.AverageResponseTime = backendMetrics.TotalResponseTime / time.Duration(backendMetrics.Requests)
	}
	backendMetrics.LastRequestTime = time.Now()
}

// GetMetrics returns proxy metrics
func (pm *proxyManager) GetMetrics() *ProxyMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &ProxyMetrics{
		TotalRequests:       atomic.LoadInt64(&pm.metrics.TotalRequests),
		SuccessfulRequests:  atomic.LoadInt64(&pm.metrics.SuccessfulRequests),
		FailedRequests:      atomic.LoadInt64(&pm.metrics.FailedRequests),
		TotalResponseTime:   pm.metrics.TotalResponseTime,
		AverageResponseTime: pm.metrics.AverageResponseTime,
		CacheHits:           atomic.LoadInt64(&pm.metrics.CacheHits),
		CacheMisses:         atomic.LoadInt64(&pm.metrics.CacheMisses),
		TotalRetries:        atomic.LoadInt64(&pm.metrics.TotalRetries),
		BackendMetrics:      make(map[string]*BackendMetrics),
	}

	for id, bm := range pm.metrics.BackendMetrics {
		metrics.BackendMetrics[id] = &BackendMetrics{
			Requests:            bm.Requests,
			SuccessfulRequests:  bm.SuccessfulRequests,
			FailedRequests:      bm.FailedRequests,
			TotalResponseTime:   bm.TotalResponseTime,
			AverageResponseTime: bm.AverageResponseTime,
			LastRequestTime:     bm.LastRequestTime,
		}
	}

	return metrics
}

// ResetMetrics resets all metrics
func (pm *proxyManager) ResetMetrics() {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	atomic.StoreInt64(&pm.metrics.TotalRequests, 0)
	atomic.StoreInt64(&pm.metrics.SuccessfulRequests, 0)
	atomic.StoreInt64(&pm.metrics.FailedRequests, 0)
	pm.metrics.TotalResponseTime = 0
	pm.metrics.AverageResponseTime = 0
	atomic.StoreInt64(&pm.metrics.CacheHits, 0)
	atomic.StoreInt64(&pm.metrics.CacheMisses, 0)
	atomic.StoreInt64(&pm.metrics.TotalRetries, 0)
	pm.metrics.BackendMetrics = make(map[string]*BackendMetrics)
}

// getCacheKey generates a cache key for a request
func (pm *proxyManager) getCacheKey(backendID string, request *Request) string {
	return fmt.Sprintf("proxy:%s:%s:%s", backendID, request.Method, request.RequestURI)
}

// RoundRobinLoadBalancer implements round-robin load balancing
type roundRobinLoadBalancer struct {
	current uint64
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer() LoadBalancer {
	return &roundRobinLoadBalancer{}
}

// SelectBackend selects the next backend using round-robin
func (lb *roundRobinLoadBalancer) SelectBackend(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, errors.New("no backends available")
	}

	index := atomic.AddUint64(&lb.current, 1) % uint64(len(backends))
	return backends[index], nil
}

// UpdateBackend updates backend state (no-op for round-robin)
func (lb *roundRobinLoadBalancer) UpdateBackend(backendID string, success bool, responseTime time.Duration) {
	// No state to update for round-robin
}

// Type returns the load balancer type
func (lb *roundRobinLoadBalancer) Type() string {
	return "round_robin"
}

// circuitBreaker implements the CircuitBreaker interface
type circuitBreaker struct {
	states map[string]*circuitState
	mu     sync.RWMutex
	config *ProxyConfig
}

type circuitState struct {
	state            CircuitState
	failureCount     int
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	halfOpenAttempts int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *ProxyConfig) CircuitBreaker {
	return &circuitBreaker{
		states: make(map[string]*circuitState),
		config: config,
	}
}

// IsOpen checks if the circuit is open
func (cb *circuitBreaker) IsOpen(backendID string) bool {
	if !cb.config.CircuitBreakerEnabled {
		return false
	}

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state := cb.states[backendID]
	if state == nil {
		return false
	}

	// Check if we should transition from open to half-open
	if state.state == CircuitStateOpen {
		if time.Since(state.lastFailureTime) > cb.config.CircuitBreakerTimeout {
			state.state = CircuitStateHalfOpen
			state.halfOpenAttempts = 0
		}
	}

	return state.state == CircuitStateOpen
}

// RecordSuccess records a successful request
func (cb *circuitBreaker) RecordSuccess(backendID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[backendID]
	if state == nil {
		state = &circuitState{state: CircuitStateClosed}
		cb.states[backendID] = state
	}

	state.lastSuccessTime = time.Now()

	// Transition from half-open to closed after successful request
	if state.state == CircuitStateHalfOpen {
		state.state = CircuitStateClosed
		state.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *circuitBreaker) RecordFailure(backendID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[backendID]
	if state == nil {
		state = &circuitState{state: CircuitStateClosed}
		cb.states[backendID] = state
	}

	state.failureCount++
	state.lastFailureTime = time.Now()

	// Open circuit if threshold exceeded
	if state.failureCount >= cb.config.CircuitBreakerThreshold {
		state.state = CircuitStateOpen
	}

	// If in half-open state and failure occurs, go back to open
	if state.state == CircuitStateHalfOpen {
		state.state = CircuitStateOpen
	}
}

// Reset resets the circuit breaker state
func (cb *circuitBreaker) Reset(backendID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	delete(cb.states, backendID)
}

// GetState returns the current circuit state
func (cb *circuitBreaker) GetState(backendID string) CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	state := cb.states[backendID]
	if state == nil {
		return CircuitStateClosed
	}

	return state.state
}

// connectionPool implements the ConnectionPool interface
type connectionPool struct {
	clients map[string]*http.Client
	mu      sync.RWMutex
	config  *ProxyConfig
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *ProxyConfig) ConnectionPool {
	return &connectionPool{
		clients: make(map[string]*http.Client),
		config:  config,
	}
}

// GetConnection gets a connection for a backend
func (cp *connectionPool) GetConnection(backendID string) (*http.Client, error) {
	cp.mu.RLock()
	client, exists := cp.clients[backendID]
	cp.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := cp.clients[backendID]; exists {
		return client, nil
	}

	transport := &http.Transport{
		MaxIdleConns:        cp.config.MaxConnectionsPerBackend,
		MaxIdleConnsPerHost: cp.config.MaxConnectionsPerBackend,
		IdleConnTimeout:     cp.config.IdleConnTimeout,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		DialContext: (&net.Dialer{
			Timeout:   cp.config.ConnectionTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   cp.config.RequestTimeout,
	}

	cp.clients[backendID] = client
	return client, nil
}

// ReleaseConnection releases a connection back to the pool
func (cp *connectionPool) ReleaseConnection(backendID string, client *http.Client) {
	// HTTP client connections are managed automatically
	// This is a no-op for HTTP clients
}

// Close closes all connections
func (cp *connectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, client := range cp.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	cp.clients = make(map[string]*http.Client)
	return nil
}

// Stats returns connection pool statistics
func (cp *connectionPool) Stats() *PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	stats := &PoolStats{
		TotalConnections: len(cp.clients),
		PerBackend:       make(map[string]int),
	}

	for backendID := range cp.clients {
		stats.PerBackend[backendID] = 1
	}

	return stats
}
