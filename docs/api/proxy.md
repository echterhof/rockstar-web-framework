---
title: "Proxy API"
description: "Proxy manager interface for load balancing and request forwarding"
category: "api"
tags: ["api", "proxy", "load-balancing", "forwarding"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
---

# Proxy API

## Overview

The `ProxyManager` interface provides forward proxy capabilities for the Rockstar Web Framework, including load balancing, circuit breakers, connection pooling, and health checking. It enables building reverse proxies and API gateways with advanced routing and failover capabilities.

**Primary Use Cases:**
- Reverse proxy implementation
- Load balancing across backend servers
- API gateway functionality
- Service mesh integration
- Failover and circuit breaking
- Connection pooling and reuse

## Interface Definition

```go
type ProxyManager interface {
    // Backend management
    AddBackend(backend *Backend) error
    RemoveBackend(backendID string) error
    GetBackend(backendID string) (*Backend, error)
    ListBackends() []*Backend

    // Request forwarding
    Forward(ctx Context, request *Request) (*Response, error)
    ForwardHTTP(ctx context.Context, req *http.Request) (*http.Response, error)

    // Load balancing
    SetLoadBalancer(lb LoadBalancer) error
    GetLoadBalancer() LoadBalancer

    // Circuit breaker
    SetCircuitBreaker(cb CircuitBreaker) error
    GetCircuitBreaker() CircuitBreaker

    // Connection pooling
    SetConnectionPool(pool ConnectionPool) error
    GetConnectionPool() ConnectionPool

    // Health checking
    HealthCheck() error
    GetHealthStatus() map[string]*BackendHealth

    // Metrics
    GetMetrics() *ProxyMetrics
    ResetMetrics()
}
```

## Backend Management

### AddBackend

```go
func AddBackend(backend *Backend) error
```

**Description**: Adds a backend server to the proxy pool.

**Parameters**:
- `backend` (*Backend): Backend configuration

**Returns**:
- `error`: Error if addition fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    backend := &pkg.Backend{
        ID:     "backend-1",
        URL:    parseURL("http://localhost:8081"),
        Weight: 10,
        IsActive: true,
        HealthCheckPath: "/health",
        HealthCheckInterval: 10 * time.Second,
    }
    
    if err := proxy.AddBackend(backend); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to add backend"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Backend added"})
}
```

### RemoveBackend

```go
func RemoveBackend(backendID string) error
```

**Description**: Removes a backend server from the proxy pool.

**Parameters**:
- `backendID` (string): Backend identifier

**Returns**:
- `error`: Error if removal fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    backendID := ctx.Param("id")
    
    if err := proxy.RemoveBackend(backendID); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to remove backend"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Backend removed"})
}
```

### GetBackend

```go
func GetBackend(backendID string) (*Backend, error)
```

**Description**: Retrieves a backend configuration by ID.

**Parameters**:
- `backendID` (string): Backend identifier

**Returns**:
- `*Backend`: Backend configuration
- `error`: Error if backend not found

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    backendID := ctx.Param("id")
    
    backend, err := proxy.GetBackend(backendID)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Backend not found"})
    }
    
    return ctx.JSON(200, backend)
}
```

### ListBackends

```go
func ListBackends() []*Backend
```

**Description**: Returns all registered backends.

**Returns**:
- `[]*Backend`: List of backends

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    backends := proxy.ListBackends()
    
    return ctx.JSON(200, map[string]interface{}{
        "count":    len(backends),
        "backends": backends,
    })
}
```

## Request Forwarding

### Forward

```go
func Forward(ctx Context, request *Request) (*Response, error)
```

**Description**: Forwards a request to a backend server using the configured load balancer.

**Parameters**:
- `ctx` (Context): Request context
- `request` (*Request): Request to forward

**Returns**:
- `*Response`: Backend response
- `error`: Error if forwarding fails

**Example**:
```go
func proxyHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    // Forward request to backend
    response, err := proxy.Forward(ctx, ctx.Request())
    if err != nil {
        return ctx.JSON(502, map[string]string{"error": "Bad gateway"})
    }
    
    // Return backend response
    return ctx.JSON(response.StatusCode, response.Body)
}
```

### ForwardHTTP

```go
func ForwardHTTP(ctx context.Context, req *http.Request) (*http.Response, error)
```

**Description**: Forwards a standard HTTP request to a backend server.

**Parameters**:
- `ctx` (context.Context): Go context
- `req` (*http.Request): HTTP request to forward

**Returns**:
- `*http.Response`: HTTP response from backend
- `error`: Error if forwarding fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    // Create HTTP request
    req, _ := http.NewRequest("GET", "/api/users", nil)
    
    // Forward request
    resp, err := proxy.ForwardHTTP(ctx.Context(), req)
    if err != nil {
        return ctx.JSON(502, map[string]string{"error": "Bad gateway"})
    }
    defer resp.Body.Close()
    
    // Read response
    body, _ := ioutil.ReadAll(resp.Body)
    return ctx.String(resp.StatusCode, string(body))
}
```

## Load Balancing

### SetLoadBalancer

```go
func SetLoadBalancer(lb LoadBalancer) error
```

**Description**: Sets the load balancing strategy.

**Parameters**:
- `lb` (LoadBalancer): Load balancer implementation

**Returns**:
- `error`: Error if setting fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    // Set round-robin load balancer
    lb := pkg.NewRoundRobinLoadBalancer()
    if err := proxy.SetLoadBalancer(lb); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set load balancer"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Load balancer updated"})
}
```

### GetLoadBalancer

```go
func GetLoadBalancer() LoadBalancer
```

**Description**: Returns the current load balancer.

**Returns**:
- `LoadBalancer`: Current load balancer

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    lb := proxy.GetLoadBalancer()
    
    return ctx.JSON(200, map[string]interface{}{
        "type": lb.Type(),
    })
}
```

## Circuit Breaker

### SetCircuitBreaker

```go
func SetCircuitBreaker(cb CircuitBreaker) error
```

**Description**: Sets the circuit breaker for fault tolerance.

**Parameters**:
- `cb` (CircuitBreaker): Circuit breaker implementation

**Returns**:
- `error`: Error if setting fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    // Set circuit breaker
    cb := pkg.NewCircuitBreaker(5, 30*time.Second)
    if err := proxy.SetCircuitBreaker(cb); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to set circuit breaker"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Circuit breaker updated"})
}
```

### GetCircuitBreaker

```go
func GetCircuitBreaker() CircuitBreaker
```

**Description**: Returns the current circuit breaker.

**Returns**:
- `CircuitBreaker`: Current circuit breaker

**Example**:
```go
func handler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    cb := proxy.GetCircuitBreaker()
    
    // Check circuit state for a backend
    state := cb.GetState("backend-1")
    
    return ctx.JSON(200, map[string]interface{}{
        "state": state,
    })
}
```

## Health Checking

### HealthCheck

```go
func HealthCheck() error
```

**Description**: Performs health checks on all backends.

**Returns**:
- `error`: Error if health check fails

**Example**:
```go
func healthCheckHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    
    if err := proxy.HealthCheck(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Health check failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "All backends healthy"})
}
```

### GetHealthStatus

```go
func GetHealthStatus() map[string]*BackendHealth
```

**Description**: Returns health status for all backends.

**Returns**:
- `map[string]*BackendHealth`: Map of backend ID to health status

**Example**:
```go
func statusHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    health := proxy.GetHealthStatus()
    
    return ctx.JSON(200, health)
}
```

## Metrics

### GetMetrics

```go
func GetMetrics() *ProxyMetrics
```

**Description**: Returns proxy performance metrics.

**Returns**:
- `*ProxyMetrics`: Proxy metrics

**Example**:
```go
func metricsHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    metrics := proxy.GetMetrics()
    
    return ctx.JSON(200, map[string]interface{}{
        "total_requests":      metrics.TotalRequests,
        "successful_requests": metrics.SuccessfulRequests,
        "failed_requests":     metrics.FailedRequests,
        "avg_response_time":   metrics.AverageResponseTime,
    })
}
```

### ResetMetrics

```go
func ResetMetrics()
```

**Description**: Resets all proxy metrics to zero.

**Example**:
```go
func resetHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    proxy.ResetMetrics()
    
    return ctx.JSON(200, map[string]string{"message": "Metrics reset"})
}
```

## Complete Example

Here's a complete example demonstrating proxy setup with load balancing:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
    "net/url"
)

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    proxy := app.Proxy()

    // Add backend servers
    addBackends(proxy)

    // Configure load balancer
    lb := pkg.NewRoundRobinLoadBalancer()
    proxy.SetLoadBalancer(lb)

    // Configure circuit breaker
    cb := pkg.NewCircuitBreaker(5, 30*time.Second)
    proxy.SetCircuitBreaker(cb)

    // Register routes
    router := app.Router()
    router.GET("/api/*", proxyHandler)
    router.GET("/proxy/status", statusHandler)
    router.GET("/proxy/metrics", metricsHandler)

    app.Listen(":8080")
}

func addBackends(proxy pkg.ProxyManager) {
    backends := []string{
        "http://localhost:8081",
        "http://localhost:8082",
        "http://localhost:8083",
    }

    for i, urlStr := range backends {
        u, _ := url.Parse(urlStr)
        backend := &pkg.Backend{
            ID:                  fmt.Sprintf("backend-%d", i+1),
            URL:                 u,
            Weight:              10,
            IsActive:            true,
            HealthCheckPath:     "/health",
            HealthCheckInterval: 10 * time.Second,
            HealthCheckTimeout:  5 * time.Second,
        }

        if err := proxy.AddBackend(backend); err != nil {
            log.Printf("Failed to add backend: %v", err)
        }
    }
}

func proxyHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()

    // Forward request to backend
    response, err := proxy.Forward(ctx, ctx.Request())
    if err != nil {
        return ctx.JSON(502, map[string]string{
            "error": "Bad gateway",
        })
    }

    return ctx.JSON(response.StatusCode, response.Body)
}

func statusHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()

    backends := proxy.ListBackends()
    health := proxy.GetHealthStatus()

    return ctx.JSON(200, map[string]interface{}{
        "backends": backends,
        "health":   health,
    })
}

func metricsHandler(ctx pkg.Context) error {
    proxy := ctx.Framework().Proxy()
    metrics := proxy.GetMetrics()

    return ctx.JSON(200, metrics)
}
```

## Best Practices

### Backend Configuration

1. **Health checks**: Always configure health checks
2. **Timeouts**: Set appropriate timeouts for backend calls
3. **Weights**: Use weights for traffic distribution
4. **Metadata**: Store backend metadata for routing decisions

### Load Balancing

1. **Choose appropriate strategy**: Round-robin, weighted, least connections
2. **Monitor distribution**: Track requests per backend
3. **Dynamic weights**: Adjust weights based on performance

### Circuit Breaking

1. **Set thresholds**: Configure failure thresholds appropriately
2. **Recovery time**: Allow sufficient time for recovery
3. **Fallback**: Implement fallback responses

### Connection Pooling

1. **Pool size**: Configure based on expected load
2. **Idle timeout**: Clean up idle connections
3. **Max connections**: Prevent resource exhaustion

## See Also

- [Framework API](framework.md)
- [Context API](context.md)
