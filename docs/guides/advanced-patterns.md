---
title: "Advanced Patterns"
description: "Sophisticated usage patterns for the Rockstar Web Framework"
category: "guide"
tags: ["advanced", "patterns", "http2", "websockets", "circuit-breaker", "plugins", "context"]
version: "1.0.0"
last_updated: "2025-11-29"
---

# Advanced Patterns

## Overview

This guide covers sophisticated usage patterns that leverage the full power of the Rockstar Web Framework. These patterns combine multiple framework features to solve complex real-world problems in production applications.

**Who should read this guide:**
- Experienced developers building high-performance applications
- Teams implementing advanced architectural patterns
- Developers optimizing for specific use cases (real-time, high-throughput, etc.)
- Anyone looking to leverage HTTP/2, WebSocket, or plugin capabilities

**What you'll learn:**
- HTTP/2 server push for performance optimization
- WebSocket connection hijacking for custom protocols
- Circuit breaker implementation for resilience
- Plugin service sharing for extensibility
- Request context cancellation for resource management
- Combining patterns for complex applications

## Table of Contents

1. [HTTP/2 Server Push](#http2-server-push)
2. [WebSocket Connection Hijacking](#websocket-connection-hijacking)
3. [Circuit Breaker Implementation](#circuit-breaker-implementation)
4. [Plugin Service Sharing](#plugin-service-sharing)
5. [Request Context Cancellation](#request-context-cancellation)
6. [Combining Patterns](#combining-patterns)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## HTTP/2 Server Push

### Concept

HTTP/2 server push allows the server to proactively send resources to the client before they're requested. This eliminates round-trip latency for critical resources like CSS, JavaScript, and images that the server knows the client will need.

**When to use:**
- Pushing critical CSS and JavaScript files
- Preloading fonts and images for faster page rendering
- Sending API responses with related data
- Optimizing single-page application (SPA) initial load

**Performance benefits:**
- Eliminates round-trip time for predictable resources
- Reduces time to first paint and time to interactive
- Improves perceived performance for users
- Reduces bandwidth waste compared to inlining

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP2: true,
        },
    }
    
    app, _ := pkg.New(config)
    router := app.Router()
    
    router.GET("/", func(ctx pkg.Context) error {
        // Push critical resources before sending HTML
        pusher := ctx.Response()
        
        // Push CSS file
        if err := pusher.Push("/static/styles.css", nil); err != nil {
            ctx.Logger().Warn("Failed to push CSS: " + err.Error())
        }
        
        // Push JavaScript file
        if err := pusher.Push("/static/app.js", nil); err != nil {
            ctx.Logger().Warn("Failed to push JS: " + err.Error())
        }
        
        // Send HTML response
        return ctx.HTML(200, "index.html", nil)
    })
    
    app.Listen(":8080")
}
```


### Advanced Push with Options

Control push behavior with PushOptions:

```go
router.GET("/dashboard", func(ctx pkg.Context) error {
    pusher := ctx.Response()
    
    // Push with custom headers
    pushOpts := &http.PushOptions{
        Method: "GET",
        Header: http.Header{
            "Accept-Encoding": []string{"gzip"},
        },
    }
    
    // Push multiple resources
    resources := []string{
        "/static/dashboard.css",
        "/static/dashboard.js",
        "/static/chart-lib.js",
        "/api/dashboard/data",
    }
    
    for _, resource := range resources {
        if err := pusher.Push(resource, pushOpts); err != nil {
            // Log but don't fail - push is an optimization
            ctx.Logger().Debug("Push failed for " + resource + ": " + err.Error())
        }
    }
    
    return ctx.HTML(200, "dashboard.html", nil)
})
```

### Conditional Push

Only push resources when beneficial:

```go
func smartPushMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check if client supports HTTP/2
    if ctx.Request().ProtoMajor < 2 {
        return next(ctx)
    }
    
    // Check if resources are already cached
    ifNoneMatch := ctx.GetHeader("If-None-Match")
    if ifNoneMatch != "" {
        // Client has cached version, skip push
        return next(ctx)
    }
    
    // Store pusher in context for handlers to use
    ctx.Set("http2_push_enabled", true)
    
    return next(ctx)
}

router.GET("/app", func(ctx pkg.Context) error {
    // Only push if middleware enabled it
    if enabled, ok := ctx.Get("http2_push_enabled"); ok && enabled.(bool) {
        pusher := ctx.Response()
        pusher.Push("/static/app.css", nil)
        pusher.Push("/static/app.js", nil)
    }
    
    return ctx.HTML(200, "app.html", nil)
}, smartPushMiddleware)
```


### Use Cases

**Single Page Application (SPA) Initial Load:**

```go
router.GET("/spa", func(ctx pkg.Context) error {
    pusher := ctx.Response()
    
    // Push all critical SPA resources
    criticalResources := []string{
        "/static/vendor.js",      // Third-party libraries
        "/static/app.bundle.js",  // Application code
        "/static/app.css",        // Styles
        "/static/fonts/main.woff2", // Web fonts
        "/api/config",            // Initial configuration
    }
    
    for _, resource := range criticalResources {
        pusher.Push(resource, nil)
    }
    
    return ctx.HTML(200, "spa.html", nil)
})
```

**API Response with Related Data:**

```go
router.GET("/api/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    pusher := ctx.Response()
    
    // Push related resources the client will likely request
    pusher.Push("/api/users/"+userID+"/posts", nil)
    pusher.Push("/api/users/"+userID+"/profile-image", nil)
    
    // Return main user data
    user := getUserData(userID)
    return ctx.JSON(200, user)
})
```

### Performance Considerations

**Do:**
- Push only critical resources needed for initial render
- Limit pushes to 5-10 resources per page
- Push resources that are always needed
- Monitor push effectiveness with metrics

**Don't:**
- Push resources that might already be cached
- Push large files (>100KB) - let client request them
- Push resources conditionally needed
- Push without measuring impact

### Common Mistakes

**Mistake 1: Pushing too many resources**
```go
// BAD: Pushing everything
for i := 0; i < 50; i++ {
    pusher.Push(fmt.Sprintf("/static/image%d.jpg", i), nil)
}
```

**Mistake 2: Not handling push errors**
```go
// BAD: Ignoring errors
pusher.Push("/static/app.js", nil) // Error ignored

// GOOD: Log errors for debugging
if err := pusher.Push("/static/app.js", nil); err != nil {
    ctx.Logger().Debug("Push failed: " + err.Error())
}
```

**Mistake 3: Pushing without HTTP/2 check**
```go
// BAD: Assuming HTTP/2
pusher.Push("/static/app.js", nil)

// GOOD: Check protocol version
if ctx.Request().ProtoMajor >= 2 {
    pusher.Push("/static/app.js", nil)
}
```

---

## WebSocket Connection Hijacking

### Concept

Connection hijacking allows you to take over the underlying TCP connection from the HTTP handler. This is useful for implementing custom protocols, raw WebSocket handling, or bidirectional streaming that doesn't fit the standard HTTP request/response model.

**When to use:**
- Implementing custom binary protocols
- Low-level WebSocket control
- Tunneling other protocols over HTTP
- Building proxy servers
- Custom streaming protocols

**Important:** The framework provides high-level WebSocket support via `router.WebSocket()`. Use hijacking only when you need low-level control.

### Basic Hijacking

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    router := app.Router()
    
    router.GET("/hijack", func(ctx pkg.Context) error {
        // Hijack the connection
        conn, rw, err := ctx.Response().Hijack()
        if err != nil {
            return ctx.JSON(500, map[string]string{
                "error": "Failed to hijack connection",
            })
        }
        defer conn.Close()
        
        // Send HTTP response manually
        response := "HTTP/1.1 200 OK\r\n" +
                   "Content-Type: text/plain\r\n" +
                   "\r\n" +
                   "Connection hijacked!\r\n"
        
        rw.WriteString(response)
        rw.Flush()
        
        // Now use raw connection for custom protocol
        for {
            line, err := rw.ReadString('\n')
            if err != nil {
                break
            }
            
            // Echo back
            rw.WriteString("Echo: " + line)
            rw.Flush()
        }
        
        return nil
    })
    
    app.Listen(":8080")
}
```


### Custom Binary Protocol

Implement a custom binary protocol over HTTP:

```go
router.GET("/binary-protocol", func(ctx pkg.Context) error {
    conn, rw, err := ctx.Response().Hijack()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Send upgrade response
    response := "HTTP/1.1 101 Switching Protocols\r\n" +
               "Upgrade: custom-protocol\r\n" +
               "Connection: Upgrade\r\n" +
               "\r\n"
    
    rw.WriteString(response)
    rw.Flush()
    
    // Custom binary protocol handler
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            break
        }
        
        // Process binary data
        data := buffer[:n]
        processed := processCustomProtocol(data)
        
        // Send binary response
        conn.Write(processed)
    }
    
    return nil
})

func processCustomProtocol(data []byte) []byte {
    // Implement your custom protocol logic
    // Example: simple echo with length prefix
    result := make([]byte, len(data)+4)
    result[0] = byte(len(data) >> 24)
    result[1] = byte(len(data) >> 16)
    result[2] = byte(len(data) >> 8)
    result[3] = byte(len(data))
    copy(result[4:], data)
    return result
}
```

### TCP Proxy Pattern

Build a TCP proxy using connection hijacking:

```go
router.GET("/proxy/:target", func(ctx pkg.Context) error {
    target := ctx.Param("target")
    
    // Hijack client connection
    clientConn, clientRW, err := ctx.Response().Hijack()
    if err != nil {
        return err
    }
    defer clientConn.Close()
    
    // Connect to target server
    targetConn, err := net.Dial("tcp", target+":80")
    if err != nil {
        clientRW.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
        clientRW.Flush()
        return nil
    }
    defer targetConn.Close()
    
    // Send success response
    clientRW.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
    clientRW.Flush()
    
    // Bidirectional copy
    done := make(chan error, 2)
    
    // Client to target
    go func() {
        _, err := io.Copy(targetConn, clientConn)
        done <- err
    }()
    
    // Target to client
    go func() {
        _, err := io.Copy(clientConn, targetConn)
        done <- err
    }()
    
    // Wait for either direction to complete
    <-done
    
    return nil
})
```


### Connection Management Best Practices

**Set Timeouts:**

```go
router.GET("/hijack-safe", func(ctx pkg.Context) error {
    conn, rw, err := ctx.Response().Hijack()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Set read/write deadlines
    conn.SetDeadline(time.Now().Add(30 * time.Second))
    
    // Send response
    rw.WriteString("HTTP/1.1 200 OK\r\n\r\n")
    rw.Flush()
    
    // Handle connection with timeout protection
    buffer := make([]byte, 1024)
    for {
        conn.SetReadDeadline(time.Now().Add(10 * time.Second))
        n, err := conn.Read(buffer)
        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                ctx.Logger().Info("Connection timeout")
            }
            break
        }
        
        // Process data
        conn.Write(buffer[:n])
    }
    
    return nil
})
```

**Resource Cleanup:**

```go
func hijackWithCleanup(ctx pkg.Context) error {
    conn, rw, err := ctx.Response().Hijack()
    if err != nil {
        return err
    }
    
    // Ensure cleanup happens
    defer func() {
        conn.Close()
        ctx.Logger().Info("Connection closed")
    }()
    
    // Track connection for monitoring
    ctx.Metrics().Increment("hijacked_connections")
    defer ctx.Metrics().Decrement("hijacked_connections")
    
    // Handle connection
    rw.WriteString("HTTP/1.1 200 OK\r\n\r\n")
    rw.Flush()
    
    // ... connection handling logic ...
    
    return nil
}
```

### Security Considerations

**Validate Before Hijacking:**

```go
func secureHijack(ctx pkg.Context) error {
    // Verify authentication
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Check authorization
    if !ctx.IsAuthorized("hijack", "connect") {
        return ctx.JSON(403, map[string]string{"error": "Forbidden"})
    }
    
    // Validate upgrade header
    upgrade := ctx.GetHeader("Upgrade")
    if upgrade != "custom-protocol" {
        return ctx.JSON(400, map[string]string{"error": "Invalid upgrade"})
    }
    
    // Now safe to hijack
    conn, rw, err := ctx.Response().Hijack()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // ... handle connection ...
    
    return nil
}
```

### Common Mistakes

**Mistake 1: Not closing the connection**
```go
// BAD: Connection leak
conn, rw, _ := ctx.Response().Hijack()
// ... use connection but never close

// GOOD: Always defer close
conn, rw, _ := ctx.Response().Hijack()
defer conn.Close()
```

**Mistake 2: Writing after hijack without manual response**
```go
// BAD: Framework can't write after hijack
ctx.Response().Hijack()
return ctx.JSON(200, data) // This won't work!

// GOOD: Write manually
conn, rw, _ := ctx.Response().Hijack()
rw.WriteString("HTTP/1.1 200 OK\r\n\r\n")
rw.Flush()
```

---

## Circuit Breaker Implementation

### Concept

A circuit breaker prevents cascading failures by stopping requests to a failing service. It has three states: Closed (normal), Open (failing), and Half-Open (testing recovery).

**When to use:**
- Calling external APIs or microservices
- Database operations that might fail
- Any operation with potential for cascading failures
- Protecting against slow or unresponsive dependencies

**Benefits:**
- Prevents resource exhaustion
- Fails fast instead of waiting for timeouts
- Allows failing services time to recover
- Improves overall system resilience

### State Transitions

```
Closed → Open: After threshold failures
Open → Half-Open: After timeout period
Half-Open → Closed: After successful requests
Half-Open → Open: If requests still failing
```

### Basic Implementation

```go
package main

import (
    "errors"
    "sync"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
    mu          sync.RWMutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        "closed",
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    
    // Check if we should transition from open to half-open
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = "half-open"
            cb.failures = 0
        } else {
            cb.mu.Unlock()
            return errors.New("circuit breaker is open")
        }
    }
    
    cb.mu.Unlock()
    
    // Execute the function
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        
        return err
    }
    
    // Success - reset if in half-open state
    if cb.state == "half-open" {
        cb.state = "closed"
        cb.failures = 0
    }
    
    return nil
}

func (cb *CircuitBreaker) State() string {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}
```


### Using Circuit Breaker in Handlers

```go
func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Create circuit breaker for external API
    apiBreaker := NewCircuitBreaker(5, 30*time.Second)
    
    router.GET("/api/external-data", func(ctx pkg.Context) error {
        var result interface{}
        
        // Call external API through circuit breaker
        err := apiBreaker.Call(func() error {
            var callErr error
            result, callErr = callExternalAPI()
            return callErr
        })
        
        if err != nil {
            if err.Error() == "circuit breaker is open" {
                return ctx.JSON(503, map[string]string{
                    "error": "Service temporarily unavailable",
                    "retry_after": "30",
                })
            }
            
            return ctx.JSON(500, map[string]string{
                "error": "External API error",
            })
        }
        
        return ctx.JSON(200, result)
    })
    
    app.Listen(":8080")
}

func callExternalAPI() (interface{}, error) {
    // Simulate external API call
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return nil, errors.New("API returned error")
    }
    
    var data interface{}
    json.NewDecoder(resp.Body).Decode(&data)
    return data, nil
}
```

### Circuit Breaker Middleware

Create reusable middleware with circuit breaker:

```go
func CircuitBreakerMiddleware(breaker *CircuitBreaker) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        // Check circuit breaker state
        if breaker.State() == "open" {
            return ctx.JSON(503, map[string]interface{}{
                "error": "Service temporarily unavailable",
                "state": "circuit_breaker_open",
            })
        }
        
        // Execute handler through circuit breaker
        err := breaker.Call(func() error {
            return next(ctx)
        })
        
        return err
    }
}

// Usage
externalAPIBreaker := NewCircuitBreaker(5, 30*time.Second)

router.GET("/api/users", getUsersHandler, 
    CircuitBreakerMiddleware(externalAPIBreaker))
```


### Advanced Circuit Breaker with Metrics

Integrate with framework metrics:

```go
type MetricsCircuitBreaker struct {
    *CircuitBreaker
    metrics pkg.MetricsCollector
    name    string
}

func NewMetricsCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration, metrics pkg.MetricsCollector) *MetricsCircuitBreaker {
    return &MetricsCircuitBreaker{
        CircuitBreaker: NewCircuitBreaker(maxFailures, resetTimeout),
        metrics:        metrics,
        name:           name,
    }
}

func (mcb *MetricsCircuitBreaker) Call(fn func() error) error {
    // Record state before call
    mcb.metrics.Gauge("circuit_breaker."+mcb.name+".state", mcb.stateValue())
    
    // Execute through circuit breaker
    err := mcb.CircuitBreaker.Call(fn)
    
    // Record metrics
    if err != nil {
        if err.Error() == "circuit breaker is open" {
            mcb.metrics.Increment("circuit_breaker." + mcb.name + ".rejected")
        } else {
            mcb.metrics.Increment("circuit_breaker." + mcb.name + ".failures")
        }
    } else {
        mcb.metrics.Increment("circuit_breaker." + mcb.name + ".successes")
    }
    
    // Record state after call
    mcb.metrics.Gauge("circuit_breaker."+mcb.name+".state", mcb.stateValue())
    
    return err
}

func (mcb *MetricsCircuitBreaker) stateValue() float64 {
    switch mcb.State() {
    case "closed":
        return 0
    case "half-open":
        return 1
    case "open":
        return 2
    default:
        return -1
    }
}

// Usage
router.GET("/api/data", func(ctx pkg.Context) error {
    breaker := NewMetricsCircuitBreaker(
        "external_api",
        5,
        30*time.Second,
        ctx.Metrics(),
    )
    
    var result interface{}
    err := breaker.Call(func() error {
        var callErr error
        result, callErr = callExternalAPI()
        return callErr
    })
    
    if err != nil {
        return ctx.JSON(503, map[string]string{"error": err.Error()})
    }
    
    return ctx.JSON(200, result)
})
```

### Configuration and Tuning

**Failure Threshold:**
- Low traffic: 3-5 failures
- High traffic: 10-20 failures
- Critical services: 2-3 failures

**Reset Timeout:**
- Fast recovery services: 10-30 seconds
- Slow recovery services: 60-120 seconds
- Database connections: 30-60 seconds

**Example Configuration:**

```go
type CircuitBreakerConfig struct {
    MaxFailures  int
    ResetTimeout time.Duration
    HalfOpenMax  int // Max requests in half-open state
}

var configs = map[string]CircuitBreakerConfig{
    "external_api": {
        MaxFailures:  5,
        ResetTimeout: 30 * time.Second,
        HalfOpenMax:  3,
    },
    "database": {
        MaxFailures:  3,
        ResetTimeout: 60 * time.Second,
        HalfOpenMax:  1,
    },
    "cache": {
        MaxFailures:  10,
        ResetTimeout: 10 * time.Second,
        HalfOpenMax:  5,
    },
}
```

---

## Plugin Service Sharing

### Concept

The plugin system allows plugins to export services that other plugins can import, enabling cross-plugin communication and code reuse without tight coupling.

**When to use:**
- Sharing functionality between plugins
- Building plugin ecosystems
- Creating plugin dependencies
- Implementing plugin-based architectures

**Benefits:**
- Loose coupling between plugins
- Reusable services across plugins
- Dynamic service discovery
- Version-aware service contracts

### Basic Service Export and Import

**Exporting a Service:**

```go
// In authentication plugin
package main

import "github.com/echterhof/rockstar-web-framework/pkg"

type AuthService interface {
    ValidateToken(token string) (string, error)
    GenerateToken(userID string) (string, error)
}

type authServiceImpl struct {
    secret string
}

func (s *authServiceImpl) ValidateToken(token string) (string, error) {
    // Validate JWT token
    userID, err := validateJWT(token, s.secret)
    return userID, err
}

func (s *authServiceImpl) GenerateToken(userID string) (string, error) {
    // Generate JWT token
    token, err := generateJWT(userID, s.secret)
    return token, err
}

type AuthPlugin struct{}

func (p *AuthPlugin) Name() string { return "auth-plugin" }
func (p *AuthPlugin) Version() string { return "1.0.0" }
func (p *AuthPlugin) Description() string { return "Authentication service" }
func (p *AuthPlugin) Author() string { return "Your Name" }

func (p *AuthPlugin) Initialize(ctx pkg.PluginContext) error {
    // Create service instance
    authService := &authServiceImpl{
        secret: "your-secret-key",
    }
    
    // Export service for other plugins
    err := ctx.ExportService("auth", authService)
    if err != nil {
        return err
    }
    
    ctx.Logger().Info("Auth service exported")
    return nil
}

func (p *AuthPlugin) Shutdown(ctx pkg.PluginContext) error {
    return nil
}
```


**Importing a Service:**

```go
// In user management plugin
package main

import "github.com/echterhof/rockstar-web-framework/pkg"

type UserPlugin struct {
    authService AuthService
}

func (p *UserPlugin) Name() string { return "user-plugin" }
func (p *UserPlugin) Version() string { return "1.0.0" }
func (p *UserPlugin) Description() string { return "User management" }
func (p *UserPlugin) Author() string { return "Your Name" }

func (p *UserPlugin) Initialize(ctx pkg.PluginContext) error {
    // Import auth service from auth plugin
    service, err := ctx.ImportService("auth-plugin", "auth")
    if err != nil {
        return err
    }
    
    // Type assert to expected interface
    authService, ok := service.(AuthService)
    if !ok {
        return fmt.Errorf("invalid auth service type")
    }
    
    p.authService = authService
    
    // Register routes that use the auth service
    router := ctx.Router()
    router.GET("/api/users/me", p.getCurrentUser)
    
    ctx.Logger().Info("User plugin initialized with auth service")
    return nil
}

func (p *UserPlugin) getCurrentUser(ctx pkg.Context) error {
    // Get token from header
    token := ctx.GetHeader("Authorization")
    
    // Use imported auth service
    userID, err := p.authService.ValidateToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid token"})
    }
    
    // Get user data
    user := getUserData(userID)
    return ctx.JSON(200, user)
}

func (p *UserPlugin) Shutdown(ctx pkg.PluginContext) error {
    return nil
}
```

### Service Versioning

Handle service version compatibility:

```go
type VersionedAuthService interface {
    Version() string
    ValidateToken(token string) (string, error)
    GenerateToken(userID string) (string, error)
}

type authServiceV2 struct {
    secret string
}

func (s *authServiceV2) Version() string {
    return "2.0.0"
}

func (s *authServiceV2) ValidateToken(token string) (string, error) {
    // V2 implementation with additional features
    return validateJWTv2(token, s.secret)
}

func (s *authServiceV2) GenerateToken(userID string) (string, error) {
    // V2 implementation with additional claims
    return generateJWTv2(userID, s.secret)
}

// In importing plugin
func (p *UserPlugin) Initialize(ctx pkg.PluginContext) error {
    service, err := ctx.ImportService("auth-plugin", "auth")
    if err != nil {
        return err
    }
    
    // Check version
    if versioned, ok := service.(VersionedAuthService); ok {
        version := versioned.Version()
        ctx.Logger().Info("Using auth service version: " + version)
        
        // Handle version-specific logic
        if version >= "2.0.0" {
            // Use V2 features
        }
    }
    
    p.authService = service.(AuthService)
    return nil
}
```


### Service Discovery Pattern

Dynamically discover available services:

```go
type ServiceDiscoveryPlugin struct {
    services map[string]interface{}
}

func (p *ServiceDiscoveryPlugin) Initialize(ctx pkg.PluginContext) error {
    p.services = make(map[string]interface{})
    
    // Try to import known services
    knownServices := []struct {
        plugin  string
        service string
    }{
        {"auth-plugin", "auth"},
        {"cache-plugin", "cache"},
        {"storage-plugin", "storage"},
    }
    
    for _, svc := range knownServices {
        service, err := ctx.ImportService(svc.plugin, svc.service)
        if err != nil {
            ctx.Logger().Warn(fmt.Sprintf("Service %s.%s not available: %v", 
                svc.plugin, svc.service, err))
            continue
        }
        
        key := svc.plugin + "." + svc.service
        p.services[key] = service
        ctx.Logger().Info("Discovered service: " + key)
    }
    
    return nil
}

func (p *ServiceDiscoveryPlugin) GetService(name string) (interface{}, bool) {
    service, ok := p.services[name]
    return service, ok
}
```

### Dependency Management

Declare and validate plugin dependencies:

```go
type PluginWithDependencies struct {
    dependencies []string
}

func (p *PluginWithDependencies) Initialize(ctx pkg.PluginContext) error {
    // Define required dependencies
    p.dependencies = []string{
        "auth-plugin.auth",
        "cache-plugin.cache",
    }
    
    // Validate all dependencies are available
    for _, dep := range p.dependencies {
        parts := strings.Split(dep, ".")
        if len(parts) != 2 {
            return fmt.Errorf("invalid dependency format: %s", dep)
        }
        
        _, err := ctx.ImportService(parts[0], parts[1])
        if err != nil {
            return fmt.Errorf("missing required dependency %s: %w", dep, err)
        }
    }
    
    ctx.Logger().Info("All dependencies satisfied")
    return nil
}
```

### Service Registry Pattern

Create a centralized service registry:

```go
type ServiceRegistryPlugin struct {
    registry map[string]ServiceInfo
    mu       sync.RWMutex
}

type ServiceInfo struct {
    Plugin      string
    Service     string
    Version     string
    Description string
    Instance    interface{}
}

func (p *ServiceRegistryPlugin) Initialize(ctx pkg.PluginContext) error {
    p.registry = make(map[string]ServiceInfo)
    
    // Export registry service
    ctx.ExportService("registry", p)
    
    // Register API endpoint
    router := ctx.Router()
    router.GET("/api/services", p.listServices)
    
    return nil
}

func (p *ServiceRegistryPlugin) RegisterService(info ServiceInfo) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    key := info.Plugin + "." + info.Service
    p.registry[key] = info
    return nil
}

func (p *ServiceRegistryPlugin) listServices(ctx pkg.Context) error {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    services := make([]ServiceInfo, 0, len(p.registry))
    for _, info := range p.registry {
        services = append(services, info)
    }
    
    return ctx.JSON(200, services)
}
```

### Best Practices

**1. Define Clear Service Interfaces:**

```go
// Good: Clear interface contract
type CacheService interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
}

// Bad: Vague interface
type Service interface {
    Do(params ...interface{}) interface{}
}
```

**2. Handle Missing Services Gracefully:**

```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Try to import optional service
    service, err := ctx.ImportService("optional-plugin", "service")
    if err != nil {
        ctx.Logger().Warn("Optional service not available, using fallback")
        p.useDefaultImplementation()
    } else {
        p.service = service
    }
    
    return nil
}
```

**3. Version Your Service Interfaces:**

```go
type CacheServiceV1 interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}

type CacheServiceV2 interface {
    CacheServiceV1
    SetWithTTL(key string, value interface{}, ttl time.Duration) error
    GetMulti(keys []string) (map[string]interface{}, error)
}
```

---

## Request Context Cancellation

### Concept

Context cancellation allows you to stop long-running operations when the client disconnects or a timeout occurs. This prevents wasting resources on operations whose results won't be used.

**When to use:**
- Long-running database queries
- External API calls
- File processing operations
- Streaming responses
- Any operation that might outlive the request

**Benefits:**
- Resource efficiency
- Faster failure detection
- Better user experience
- Prevents resource leaks

### Basic Cancellation Handling

```go
package main

import (
    "context"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    router := app.Router()
    
    router.GET("/api/long-task", func(ctx pkg.Context) error {
        // Get the request context
        reqCtx := ctx.Context()
        
        // Simulate long-running operation
        for i := 0; i < 100; i++ {
            // Check if request was cancelled
            select {
            case <-reqCtx.Done():
                // Client disconnected or timeout occurred
                ctx.Logger().Info("Request cancelled")
                return reqCtx.Err()
            default:
                // Continue processing
            }
            
            // Do work
            time.Sleep(100 * time.Millisecond)
        }
        
        return ctx.JSON(200, map[string]string{
            "status": "completed",
        })
    })
    
    app.Listen(":8080")
}
```

### Using Context with Timeouts

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Create context with timeout
    timeoutCtx, cancel := ctx.WithTimeout(5 * time.Second)
    defer cancel()
    
    // Channel for result
    result := make(chan interface{}, 1)
    errChan := make(chan error, 1)
    
    // Start operation in goroutine
    go func() {
        data, err := fetchDataFromDatabase(timeoutCtx)
        if err != nil {
            errChan <- err
            return
        }
        result <- data
    }()
    
    // Wait for result or timeout
    select {
    case data := <-result:
        return ctx.JSON(200, data)
    case err := <-errChan:
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    case <-timeoutCtx.Done():
        return ctx.JSON(504, map[string]string{
            "error": "Request timeout",
        })
    }
})

func fetchDataFromDatabase(ctx context.Context) (interface{}, error) {
    // Use context in database query
    db := getDB()
    
    rows, err := db.QueryContext(ctx, "SELECT * FROM large_table")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []map[string]interface{}
    for rows.Next() {
        // Check cancellation periodically
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // Process row
        var data map[string]interface{}
        // ... scan row ...
        results = append(results, data)
    }
    
    return results, nil
}
```


### Cancellation Middleware

Use the built-in cancellation middleware:

```go
import "github.com/echterhof/rockstar-web-framework/pkg"

func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Add cancellation middleware globally
    router.Use(pkg.CancellationMiddleware())
    
    router.GET("/api/task", func(ctx pkg.Context) error {
        // Middleware will handle cancellation automatically
        result := performLongTask(ctx.Context())
        return ctx.JSON(200, result)
    })
    
    app.Listen(":8080")
}
```

### Graceful Shutdown Pattern

Handle graceful shutdown with context:

```go
router.GET("/api/batch-process", func(ctx pkg.Context) error {
    items := []string{"item1", "item2", "item3", "item4", "item5"}
    results := make([]string, 0, len(items))
    
    for _, item := range items {
        // Check cancellation before each item
        select {
        case <-ctx.Context().Done():
            // Return partial results
            return ctx.JSON(200, map[string]interface{}{
                "status":  "partial",
                "results": results,
                "message": "Request cancelled, returning partial results",
            })
        default:
        }
        
        // Process item
        result := processItem(item)
        results = append(results, result)
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "status":  "complete",
        "results": results,
    })
})
```

### Streaming with Cancellation

Handle cancellation in streaming responses:

```go
router.GET("/api/stream", func(ctx pkg.Context) error {
    // Set headers for streaming
    ctx.SetHeader("Content-Type", "text/event-stream")
    ctx.SetHeader("Cache-Control", "no-cache")
    ctx.SetHeader("Connection", "keep-alive")
    
    // Get flusher
    flusher, ok := ctx.Response().(http.Flusher)
    if !ok {
        return ctx.JSON(500, map[string]string{
            "error": "Streaming not supported",
        })
    }
    
    // Stream data with cancellation support
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Context().Done():
            // Client disconnected
            ctx.Logger().Info("Stream cancelled by client")
            return nil
            
        case <-ticker.C:
            // Send data
            data := fmt.Sprintf("data: %s\n\n", time.Now().Format(time.RFC3339))
            ctx.Response().Write([]byte(data))
            flusher.Flush()
        }
    }
})
```


### External API Calls with Cancellation

Propagate cancellation to external services:

```go
router.GET("/api/proxy", func(ctx pkg.Context) error {
    // Create HTTP client request with context
    req, err := http.NewRequestWithContext(
        ctx.Context(),
        "GET",
        "https://api.example.com/data",
        nil,
    )
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    }
    
    // Make request - will be cancelled if context is cancelled
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        // Check if error was due to cancellation
        if pkg.IsCancellationError(err) {
            return ctx.JSON(499, map[string]string{
                "error": "Request cancelled",
            })
        }
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    }
    defer resp.Body.Close()
    
    // Read and return response
    var data interface{}
    json.NewDecoder(resp.Body).Decode(&data)
    return ctx.JSON(200, data)
})
```

### Database Queries with Cancellation

Use context with database operations:

```go
router.GET("/api/users", func(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Query with context - will be cancelled if request is cancelled
    rows, err := db.QueryContext(
        ctx.Context(),
        "SELECT id, name, email FROM users WHERE active = ?",
        true,
    )
    if err != nil {
        if pkg.IsCancellationError(err) {
            return ctx.JSON(499, map[string]string{
                "error": "Query cancelled",
            })
        }
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    }
    defer rows.Close()
    
    users := []User{}
    for rows.Next() {
        // Check cancellation during iteration
        select {
        case <-ctx.Context().Done():
            return ctx.JSON(499, map[string]string{
                "error": "Request cancelled during processing",
            })
        default:
        }
        
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            return ctx.JSON(500, map[string]string{"error": err.Error()})
        }
        users = append(users, user)
    }
    
    return ctx.JSON(200, users)
})
```

### Timeout Strategies

Different timeout strategies for different scenarios:

```go
// Short timeout for fast operations
router.GET("/api/quick", func(ctx pkg.Context) error {
    timeoutCtx, cancel := ctx.WithTimeout(1 * time.Second)
    defer cancel()
    
    result := quickOperation(timeoutCtx)
    return ctx.JSON(200, result)
})

// Medium timeout for normal operations
router.GET("/api/normal", func(ctx pkg.Context) error {
    timeoutCtx, cancel := ctx.WithTimeout(10 * time.Second)
    defer cancel()
    
    result := normalOperation(timeoutCtx)
    return ctx.JSON(200, result)
})

// Long timeout for batch operations
router.POST("/api/batch", func(ctx pkg.Context) error {
    timeoutCtx, cancel := ctx.WithTimeout(60 * time.Second)
    defer cancel()
    
    result := batchOperation(timeoutCtx)
    return ctx.JSON(200, result)
})

// No timeout for streaming
router.GET("/api/stream", func(ctx pkg.Context) error {
    // Use request context directly - no additional timeout
    return streamData(ctx.Context(), ctx.Response())
})
```

### Best Practices

**1. Always check context cancellation in loops:**

```go
// Good
for i := 0; i < 1000; i++ {
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    default:
    }
    processItem(i)
}

// Bad - no cancellation check
for i := 0; i < 1000; i++ {
    processItem(i)
}
```

**2. Propagate context to all operations:**

```go
// Good
db.QueryContext(ctx.Context(), query)
http.NewRequestWithContext(ctx.Context(), method, url, body)

// Bad - no context propagation
db.Query(query)
http.NewRequest(method, url, body)
```

**3. Set appropriate timeouts:**

```go
// Good - specific timeouts
quickCtx, _ := ctx.WithTimeout(1 * time.Second)
normalCtx, _ := ctx.WithTimeout(10 * time.Second)

// Bad - one size fits all
ctx, _ := ctx.WithTimeout(30 * time.Second)
```

---

## Combining Patterns

### Real-World Scenario: Resilient API Gateway

Combine multiple patterns for a production-ready API gateway:

```go
package main

import (
    "context"
    "net/http"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP2: true,
        },
    }
    
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Global middleware
    router.Use(pkg.CancellationMiddleware())
    
    // Create circuit breakers for different services
    userServiceBreaker := NewMetricsCircuitBreaker("user_service", 5, 30*time.Second, app.Metrics())
    orderServiceBreaker := NewMetricsCircuitBreaker("order_service", 5, 30*time.Second, app.Metrics())
    
    // API Gateway endpoint
    router.GET("/api/dashboard", func(ctx pkg.Context) error {
        // HTTP/2 Server Push for static resources
        if ctx.Request().ProtoMajor >= 2 {
            pusher := ctx.Response()
            pusher.Push("/static/dashboard.css", nil)
            pusher.Push("/static/dashboard.js", nil)
        }
        
        // Create timeout context
        timeoutCtx, cancel := ctx.WithTimeout(5 * time.Second)
        defer cancel()
        
        // Fetch data from multiple services with circuit breakers
        userChan := make(chan interface{}, 1)
        orderChan := make(chan interface{}, 1)
        errChan := make(chan error, 2)
        
        // Fetch user data
        go func() {
            err := userServiceBreaker.Call(func() error {
                data, err := fetchUserData(timeoutCtx)
                if err != nil {
                    return err
                }
                userChan <- data
                return nil
            })
            if err != nil {
                errChan <- err
            }
        }()
        
        // Fetch order data
        go func() {
            err := orderServiceBreaker.Call(func() error {
                data, err := fetchOrderData(timeoutCtx)
                if err != nil {
                    return err
                }
                orderChan <- data
                return nil
            })
            if err != nil {
                errChan <- err
            }
        }()
        
        // Collect results
        var userData, orderData interface{}
        errors := []string{}
        
        for i := 0; i < 2; i++ {
            select {
            case userData = <-userChan:
            case orderData = <-orderChan:
            case err := <-errChan:
                errors = append(errors, err.Error())
            case <-timeoutCtx.Done():
                return ctx.JSON(504, map[string]interface{}{
                    "error": "Request timeout",
                })
            }
        }
        
        // Return response with partial data if needed
        response := map[string]interface{}{
            "user":   userData,
            "orders": orderData,
        }
        
        if len(errors) > 0 {
            response["errors"] = errors
            response["partial"] = true
        }
        
        return ctx.JSON(200, response)
    })
    
    app.Listen(":8080")
}

func fetchUserData(ctx context.Context) (interface{}, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://user-service/api/user", nil)
    client := &http.Client{Timeout: 3 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var data interface{}
    json.NewDecoder(resp.Body).Decode(&data)
    return data, nil
}

func fetchOrderData(ctx context.Context) (interface{}, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://order-service/api/orders", nil)
    client := &http.Client{Timeout: 3 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var data interface{}
    json.NewDecoder(resp.Body).Decode(&data)
    return data, nil
}
```


### Real-Time Monitoring Dashboard

Combine WebSocket, context cancellation, and plugin services:

```go
// Monitoring plugin that exports metrics service
type MonitoringPlugin struct {
    metrics *MetricsService
}

func (p *MonitoringPlugin) Initialize(ctx pkg.PluginContext) error {
    p.metrics = NewMetricsService()
    
    // Export metrics service
    ctx.ExportService("metrics", p.metrics)
    
    // Register WebSocket endpoint
    router := ctx.Router()
    router.WebSocket("/ws/metrics", p.handleMetricsStream)
    
    return nil
}

func (p *MonitoringPlugin) handleMetricsStream(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    // Create cancellable context
    streamCtx, cancel := ctx.WithCancel()
    defer cancel()
    
    // Send metrics every second
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-streamCtx.Done():
            // Client disconnected
            return nil
            
        case <-ticker.C:
            // Collect current metrics
            metrics := p.metrics.GetCurrent()
            
            // Send to client
            data, _ := json.Marshal(metrics)
            err := conn.WriteMessage(pkg.TextMessage, data)
            if err != nil {
                return err
            }
        }
    }
}

// Dashboard plugin that imports metrics service
type DashboardPlugin struct {
    metricsService *MetricsService
}

func (p *DashboardPlugin) Initialize(ctx pkg.PluginContext) error {
    // Import metrics service
    service, err := ctx.ImportService("monitoring-plugin", "metrics")
    if err != nil {
        return err
    }
    
    p.metricsService = service.(*MetricsService)
    
    // Register dashboard endpoint with HTTP/2 push
    router := ctx.Router()
    router.GET("/dashboard", p.serveDashboard)
    
    return nil
}

func (p *DashboardPlugin) serveDashboard(ctx pkg.Context) error {
    // Push dashboard resources
    if ctx.Request().ProtoMajor >= 2 {
        pusher := ctx.Response()
        pusher.Push("/static/dashboard.css", nil)
        pusher.Push("/static/dashboard.js", nil)
        pusher.Push("/static/chart.js", nil)
    }
    
    // Get initial metrics from service
    initialMetrics := p.metricsService.GetCurrent()
    
    return ctx.HTML(200, "dashboard.html", map[string]interface{}{
        "metrics": initialMetrics,
    })
}
```


### Microservices Communication Hub

Combine circuit breakers, service registry, and context cancellation:

```go
type ServiceHub struct {
    services map[string]*ServiceClient
    mu       sync.RWMutex
}

type ServiceClient struct {
    name    string
    baseURL string
    breaker *CircuitBreaker
    client  *http.Client
}

func NewServiceHub() *ServiceHub {
    return &ServiceHub{
        services: make(map[string]*ServiceClient),
    }
}

func (h *ServiceHub) RegisterService(name, baseURL string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    h.services[name] = &ServiceClient{
        name:    name,
        baseURL: baseURL,
        breaker: NewCircuitBreaker(5, 30*time.Second),
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (h *ServiceHub) Call(ctx context.Context, serviceName, path string) (interface{}, error) {
    h.mu.RLock()
    service, ok := h.services[serviceName]
    h.mu.RUnlock()
    
    if !ok {
        return nil, fmt.Errorf("service %s not registered", serviceName)
    }
    
    var result interface{}
    
    // Call through circuit breaker
    err := service.breaker.Call(func() error {
        // Create request with context
        req, err := http.NewRequestWithContext(
            ctx,
            "GET",
            service.baseURL+path,
            nil,
        )
        if err != nil {
            return err
        }
        
        // Make request
        resp, err := service.client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != 200 {
            return fmt.Errorf("service returned %d", resp.StatusCode)
        }
        
        // Decode response
        return json.NewDecoder(resp.Body).Decode(&result)
    })
    
    return result, err
}

// Usage in handler
router.GET("/api/aggregate", func(ctx pkg.Context) error {
    hub := NewServiceHub()
    hub.RegisterService("users", "http://user-service")
    hub.RegisterService("products", "http://product-service")
    hub.RegisterService("orders", "http://order-service")
    
    // Create timeout context
    timeoutCtx, cancel := ctx.WithTimeout(5 * time.Second)
    defer cancel()
    
    // Call multiple services concurrently
    type result struct {
        name string
        data interface{}
        err  error
    }
    
    results := make(chan result, 3)
    
    services := []struct {
        name string
        path string
    }{
        {"users", "/api/users"},
        {"products", "/api/products"},
        {"orders", "/api/orders"},
    }
    
    for _, svc := range services {
        go func(name, path string) {
            data, err := hub.Call(timeoutCtx, name, path)
            results <- result{name: name, data: data, err: err}
        }(svc.name, svc.path)
    }
    
    // Collect results
    response := make(map[string]interface{})
    errors := []string{}
    
    for i := 0; i < len(services); i++ {
        select {
        case res := <-results:
            if res.err != nil {
                errors = append(errors, fmt.Sprintf("%s: %v", res.name, res.err))
            } else {
                response[res.name] = res.data
            }
        case <-timeoutCtx.Done():
            return ctx.JSON(504, map[string]interface{}{
                "error": "Request timeout",
            })
        }
    }
    
    if len(errors) > 0 {
        response["errors"] = errors
    }
    
    return ctx.JSON(200, response)
})
```


### Pattern Interaction Considerations

**HTTP/2 Push + Context Cancellation:**
- Push resources before starting long operations
- Don't push if request is already cancelled
- Push failures shouldn't block the main response

**Circuit Breaker + Context Cancellation:**
- Circuit breaker should respect context timeouts
- Cancelled requests shouldn't count as failures
- Use separate timeouts for circuit breaker and context

**WebSocket + Circuit Breaker:**
- Apply circuit breaker to WebSocket message handlers
- Don't break the circuit on client disconnects
- Monitor WebSocket-specific metrics separately

**Plugin Services + All Patterns:**
- Services can use circuit breakers internally
- Export services that support context cancellation
- Document service timeout expectations

---

## Best Practices

### General Guidelines

1. **Start Simple**: Don't over-engineer. Add patterns as needed.

2. **Monitor Everything**: Track metrics for all patterns:
   - HTTP/2 push hit rates
   - Circuit breaker state changes
   - Context cancellation frequency
   - Service call latencies

3. **Test Failure Scenarios**: Test each pattern under failure conditions:
   - Network failures
   - Timeouts
   - Service unavailability
   - High load

4. **Document Behavior**: Document how patterns interact in your system.

5. **Use Appropriate Timeouts**: Different operations need different timeouts.

### Performance Optimization

**HTTP/2 Push:**
- Measure push effectiveness with metrics
- A/B test push vs. no-push
- Monitor cache hit rates

**Circuit Breakers:**
- Tune thresholds based on actual traffic
- Use different settings for different services
- Monitor false positives

**Context Cancellation:**
- Check cancellation in tight loops
- Propagate context to all I/O operations
- Set realistic timeouts

**Plugin Services:**
- Cache service lookups
- Use connection pooling
- Implement service health checks

### Security Considerations

1. **Validate Before Hijacking**: Always authenticate before hijacking connections.

2. **Limit Resource Usage**: Set timeouts and limits on all operations.

3. **Sanitize Service Data**: Don't trust data from plugin services.

4. **Monitor Suspicious Activity**: Track unusual patterns in metrics.

---

## Troubleshooting

### HTTP/2 Push Issues

**Problem: Push not working**

Solutions:
- Verify HTTP/2 is enabled in config
- Check client supports HTTP/2
- Ensure resources exist at push paths
- Check for HTTPS requirement (HTTP/2 often requires TLS)

**Problem: Push hurting performance**

Solutions:
- Reduce number of pushed resources
- Only push critical resources
- Check if resources are already cached
- Measure push effectiveness with metrics

### WebSocket Hijacking Issues

**Problem: Hijack fails**

Solutions:
- Verify server supports hijacking
- Check for proxy/load balancer interference
- Ensure HTTP/1.1 is being used
- Verify no middleware is writing response before hijack

**Problem: Connection drops immediately**

Solutions:
- Set appropriate timeouts
- Handle errors in read/write loops
- Ensure proper cleanup with defer
- Check for network issues

### Circuit Breaker Issues

**Problem: Circuit opens too frequently**

Solutions:
- Increase failure threshold
- Adjust timeout period
- Check for false positives
- Monitor actual service health

**Problem: Circuit stays open too long**

Solutions:
- Reduce reset timeout
- Implement health checks
- Add manual circuit reset endpoint
- Monitor service recovery time

### Context Cancellation Issues

**Problem: Operations not cancelling**

Solutions:
- Ensure context is propagated to all operations
- Check for blocking operations without cancellation checks
- Use context-aware APIs (QueryContext, NewRequestWithContext)
- Add cancellation checks in loops

**Problem: Too many cancellations**

Solutions:
- Increase timeouts
- Optimize slow operations
- Check for client-side issues
- Monitor cancellation reasons

### Plugin Service Issues

**Problem: Service import fails**

Solutions:
- Verify exporting plugin is loaded
- Check service name spelling
- Ensure plugin is initialized before import
- Check plugin load order

**Problem: Service version mismatch**

Solutions:
- Implement version checking
- Use interface versioning
- Document breaking changes
- Provide migration guides

---

## See Also

- [Middleware Guide](middleware.md) - Middleware patterns and best practices
- [WebSockets Guide](websockets.md) - WebSocket implementation details
- [Plugin System Guide](plugins.md) - Plugin development and architecture
- [Performance Guide](performance.md) - Performance optimization techniques
- [Monitoring Guide](monitoring.md) - Monitoring and metrics collection
- [API Reference: Context](../api/context.md) - Context API documentation
- [API Reference: Request/Response](../api/request-response.md) - Request and response types
- [Examples](../examples/README.md) - Complete working examples

