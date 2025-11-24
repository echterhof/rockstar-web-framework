# Forward Proxy Implementation

## Overview

The Rockstar Web Framework includes a comprehensive forward proxy system that enables load balancing, request distribution, circuit breaking, caching, and connection pooling. This implementation satisfies Requirements 15.1-15.6.

## Features

### 1. Backend Management (Requirement 15.1)

The proxy manager supports adding, removing, and managing multiple backend servers:

```go
// Create proxy manager
proxyManager := pkg.NewProxyManager(config, cache)

// Add backend
backend := &pkg.Backend{
    ID:       "backend1",
    URL:      backendURL,
    Weight:   1,
    IsActive: true,
}
proxyManager.AddBackend(backend)

// Remove backend
proxyManager.RemoveBackend("backend1")

// List all backends
backends := proxyManager.ListBackends()
```

### 2. Round-Robin Load Balancing (Requirement 15.2)

The proxy implements round-robin distribution to backend servers:

```go
// Round-robin is the default load balancer
lb := pkg.NewRoundRobinLoadBalancer()
proxyManager.SetLoadBalancer(lb)

// Requests are automatically distributed across backends
response, err := proxyManager.Forward(ctx, request)
```

The round-robin algorithm ensures even distribution of requests across all available backends.

### 3. Circuit Breakers for DNS Failures (Requirement 15.3)

Circuit breakers protect against cascading failures:

```go
// Configure circuit breaker
config := pkg.DefaultProxyConfig()
config.CircuitBreakerEnabled = true
config.CircuitBreakerThreshold = 5  // Open after 5 failures
config.CircuitBreakerTimeout = 30 * time.Second

// Circuit breaker automatically manages backend health
cb := proxyManager.GetCircuitBreaker()
state := cb.GetState("backend1")  // closed, open, or half_open
```

Circuit breaker states:
- **Closed**: Normal operation, requests flow through
- **Open**: Too many failures, requests are rejected
- **Half-Open**: Testing if backend has recovered

### 4. Caching Strategies (Requirement 15.4)

The proxy supports response caching for GET requests:

```go
// Enable caching
config.CacheEnabled = true
config.CacheTTL = 5 * time.Minute
config.CacheMaxSize = 100 * 1024 * 1024  // 100MB

// Successful GET responses are automatically cached
// Cache key format: "proxy:{backendID}:{method}:{uri}"
```

Cache features:
- Automatic caching of successful GET responses (status 200)
- Configurable TTL per cache entry
- Cache hit/miss metrics tracking
- Integration with framework cache manager

### 5. Retry Strategies (Requirement 15.5)

The proxy implements intelligent retry logic with exponential backoff:

```go
// Configure retry behavior
config.MaxRetries = 3
config.RetryDelay = 100 * time.Millisecond
config.RetryBackoff = true  // Exponential backoff

// Retries are automatic on failure
// Backoff delays: 100ms, 200ms, 400ms, etc.
```

Retry features:
- Configurable maximum retry attempts
- Optional exponential backoff
- Automatic backend selection for each retry
- Retry metrics tracking

### 6. Connection Pooling (Requirement 15.6)

The proxy maintains connection pools for each backend:

```go
// Configure connection pooling
config.MaxConnectionsPerBackend = 100
config.ConnectionTimeout = 10 * time.Second
config.IdleConnTimeout = 90 * time.Second

// Get pool statistics
pool := proxyManager.GetConnectionPool()
stats := pool.Stats()
```

Connection pool features:
- Per-backend connection pooling
- Automatic connection reuse
- Configurable connection limits
- Idle connection timeout
- Connection statistics

## Architecture

### Components

1. **ProxyManager**: Main interface for proxy operations
2. **LoadBalancer**: Selects backend for each request
3. **CircuitBreaker**: Protects against failing backends
4. **ConnectionPool**: Manages HTTP client connections
5. **HealthChecker**: Monitors backend health

### Request Flow

```
Client Request
    ↓
ProxyManager.Forward()
    ↓
Get Available Backends
    ↓
Load Balancer Selection
    ↓
Circuit Breaker Check
    ↓
Cache Check (GET only)
    ↓
Forward to Backend
    ↓
Record Metrics
    ↓
Update Health Status
    ↓
Cache Response (if applicable)
    ↓
Return Response
```

## Configuration

### Default Configuration

```go
config := pkg.DefaultProxyConfig()
// Returns:
// - LoadBalancerType: "round_robin"
// - CircuitBreakerEnabled: true
// - CircuitBreakerThreshold: 5
// - CircuitBreakerTimeout: 30s
// - MaxConnectionsPerBackend: 100
// - ConnectionTimeout: 10s
// - MaxRetries: 3
// - RetryDelay: 100ms
// - CacheEnabled: true
// - CacheTTL: 5m
// - HealthCheckEnabled: true
// - HealthCheckInterval: 10s
// - RequestTimeout: 30s
```

### Custom Configuration

```go
config := &pkg.ProxyConfig{
    LoadBalancerType:           "round_robin",
    CircuitBreakerEnabled:      true,
    CircuitBreakerThreshold:    10,
    CircuitBreakerTimeout:      60 * time.Second,
    MaxConnectionsPerBackend:   200,
    ConnectionTimeout:          15 * time.Second,
    MaxRetries:                 5,
    RetryDelay:                 200 * time.Millisecond,
    RetryBackoff:               true,
    CacheEnabled:               true,
    CacheTTL:                   10 * time.Minute,
    HealthCheckEnabled:         true,
    HealthCheckInterval:        30 * time.Second,
    RequestTimeout:             60 * time.Second,
}
```

## Health Checking

### Automatic Health Checks

```go
// Health checks run automatically when enabled
config.HealthCheckEnabled = true
config.HealthCheckInterval = 10 * time.Second
config.HealthCheckPath = "/health"

// Get health status
healthStatus := proxyManager.GetHealthStatus()
for backendID, health := range healthStatus {
    fmt.Printf("%s: %v (consecutive fails: %d)\n",
        backendID, health.IsHealthy, health.ConsecutiveFails)
}
```

### Manual Health Checks

```go
// Trigger manual health check
err := proxyManager.HealthCheck()

// Check specific backend health
health := healthStatus["backend1"]
if health.IsHealthy {
    fmt.Printf("Backend healthy, response time: %s\n", health.ResponseTime)
} else {
    fmt.Printf("Backend unhealthy: %s\n", health.ErrorMessage)
}
```

## Metrics

### Proxy Metrics

```go
metrics := proxyManager.GetMetrics()

// Overall metrics
fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
fmt.Printf("Successful: %d\n", metrics.SuccessfulRequests)
fmt.Printf("Failed: %d\n", metrics.FailedRequests)
fmt.Printf("Average Response Time: %s\n", metrics.AverageResponseTime)

// Cache metrics
fmt.Printf("Cache Hits: %d\n", metrics.CacheHits)
fmt.Printf("Cache Misses: %d\n", metrics.CacheMisses)

// Retry metrics
fmt.Printf("Total Retries: %d\n", metrics.TotalRetries)
```

### Per-Backend Metrics

```go
for backendID, backendMetrics := range metrics.BackendMetrics {
    fmt.Printf("Backend %s:\n", backendID)
    fmt.Printf("  Requests: %d\n", backendMetrics.Requests)
    fmt.Printf("  Success Rate: %.2f%%\n",
        float64(backendMetrics.SuccessfulRequests) / float64(backendMetrics.Requests) * 100)
    fmt.Printf("  Avg Response Time: %s\n", backendMetrics.AverageResponseTime)
}
```

## Integration with Framework

### Server Integration

```go
server := pkg.NewServer()

// Create proxy manager
proxyManager := pkg.NewProxyManager(config, cache)

// Add backends
proxyManager.AddBackend(backend1)
proxyManager.AddBackend(backend2)

// Add proxy route
server.Router().GET("/api/*", func(ctx pkg.Context) error {
    request := ctx.Request()
    
    // Forward to backend
    response, err := proxyManager.Forward(ctx, request)
    if err != nil {
        return ctx.String(502, "Bad Gateway")
    }
    
    // Write response
    writer := ctx.Response()
    writer.WriteHeader(response.StatusCode)
    for key, values := range response.Header {
        for _, value := range values {
            writer.Header().Add(key, value)
        }
    }
    writer.Write(response.Body)
    
    return nil
})
```

### Middleware Integration

```go
// Create proxy middleware
func ProxyMiddleware(proxyManager pkg.ProxyManager) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        // Check if request should be proxied
        if shouldProxy(ctx.Request()) {
            response, err := proxyManager.Forward(ctx, ctx.Request())
            if err != nil {
                return err
            }
            
            writer := ctx.Response()
            writer.WriteHeader(response.StatusCode)
            writer.Write(response.Body)
            return nil
        }
        
        return next(ctx)
    }
}

// Use middleware
server.Router().Use(ProxyMiddleware(proxyManager))
```

## Best Practices

### 1. Backend Configuration

- Set appropriate health check intervals based on backend stability
- Use weighted backends for capacity-based load balancing
- Configure realistic timeouts based on backend response times
- Monitor backend health status regularly

### 2. Circuit Breaker Tuning

- Set threshold based on acceptable failure rate
- Configure timeout to allow backends time to recover
- Monitor circuit breaker state transitions
- Implement alerting for open circuits

### 3. Caching Strategy

- Enable caching for read-heavy workloads
- Set appropriate TTL based on data freshness requirements
- Monitor cache hit rates
- Consider cache size limits for memory management

### 4. Retry Configuration

- Limit retries to avoid cascading delays
- Use exponential backoff to reduce backend load
- Consider idempotency when enabling retries
- Monitor retry rates for performance issues

### 5. Connection Pooling

- Set pool size based on expected concurrency
- Configure appropriate idle timeouts
- Monitor pool utilization
- Adjust limits based on backend capacity

## Error Handling

### Common Errors

```go
response, err := proxyManager.Forward(ctx, request)
if err != nil {
    switch {
    case errors.Is(err, pkg.ErrNoBackendsAvailable):
        // No healthy backends available
        return ctx.String(503, "Service Unavailable")
        
    case errors.Is(err, pkg.ErrCircuitBreakerOpen):
        // Circuit breaker is open
        return ctx.String(503, "Service Temporarily Unavailable")
        
    case errors.Is(err, pkg.ErrRequestTimeout):
        // Request timed out
        return ctx.String(504, "Gateway Timeout")
        
    default:
        // Other proxy errors
        return ctx.String(502, "Bad Gateway")
    }
}
```

## Performance Considerations

### Optimization Tips

1. **Connection Reuse**: Connection pooling automatically reuses connections
2. **Caching**: Enable caching for frequently accessed resources
3. **Health Checks**: Balance check frequency with overhead
4. **Timeouts**: Set appropriate timeouts to prevent hanging requests
5. **Metrics**: Monitor metrics to identify bottlenecks

### Scalability

The proxy system is designed for high concurrency:
- Thread-safe backend management
- Lock-free metrics collection using atomics
- Efficient connection pooling
- Concurrent health checking
- Minimal allocation in hot paths

## Testing

See `pkg/proxy_test.go` for comprehensive unit tests covering:
- Backend management
- Load balancing
- Circuit breakers
- Connection pooling
- Request forwarding
- Retry logic
- Caching
- Health checking
- Metrics collection
- Concurrent request handling

## Requirements Satisfied

- ✅ **15.1**: Forward proxy functionality with backend management
- ✅ **15.2**: Round-robin distribution to backend servers
- ✅ **15.3**: Circuit breakers for DNS and backend failures
- ✅ **15.4**: Caching strategies for GET requests
- ✅ **15.5**: Retry strategies with exponential backoff
- ✅ **15.6**: Connection pooling for efficient resource usage
