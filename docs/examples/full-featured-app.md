# Full Featured Application Example

The Full Featured Application example (`examples/full_featured_app.go`) is a comprehensive demonstration of all major features in the Rockstar Web Framework. This production-ready example shows how to build enterprise-grade applications with authentication, multi-tenancy, monitoring, and more.

## What This Example Demonstrates

- **Complete configuration** for production environments
- **Database integration** with PostgreSQL
- **Session management** with database-backed storage
- **Authentication and authorization** with JWT and RBAC
- **Caching strategies** for performance
- **Internationalization (i18n)** for multi-language support
- **Multi-tenancy** with host-based routing
- **WebSocket support** for real-time communication
- **Monitoring and profiling** with Prometheus metrics and pprof
- **Rate limiting** per tenant
- **Graceful shutdown** with cleanup hooks
- **Security features** (CSRF, XSS protection, secure headers)
- **Proxy configuration** with circuit breakers
- **Lifecycle hooks** for startup and shutdown
- **Global middleware** for logging, metrics, and recovery

## Prerequisites

- Go 1.25 or higher
- PostgreSQL database (or modify config for SQLite/MySQL/MSSQL)
- Optional: Prometheus for metrics collection

## Setup Instructions

### 1. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE rockstar_db;
CREATE USER rockstar_user WITH PASSWORD 'rockstar_pass';
GRANT ALL PRIVILEGES ON DATABASE rockstar_db TO rockstar_user;
```

### 2. Environment Variables

Set required environment variables:

```bash
# Database configuration
export DB_DRIVER=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=rockstar_db
export DB_USER=rockstar_user
export DB_PASSWORD=rockstar_pass

# Security keys
export SESSION_ENCRYPTION_KEY=$(go run examples/generate_keys.go)
export JWT_SECRET=your-secret-jwt-key
```

### 3. Run the Example

```bash
go run examples/full_featured_app.go
```

The server will start on `http://localhost:8080`.

## Available Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Home page with feature list |
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check (all services) |
| GET | `/metrics` | Prometheus metrics |
| GET | `/debug/pprof` | Profiling index |

### API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/status` | API status | No |
| GET | `/api/v1/profile` | User profile | Yes |
| PUT | `/api/v1/profile` | Update profile | Yes |
| POST | `/api/v1/logout` | Logout | Yes |

### Admin Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/users` | List users | Admin |
| POST | `/admin/users` | Create user | Admin |
| GET | `/admin/users/:id` | Get user | Admin |
| PUT | `/admin/users/:id` | Update user | Admin |
| DELETE | `/admin/users/:id` | Delete user | Admin |

### WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws` | WebSocket echo server |

### Multi-Tenant Endpoints

Use the `Host` header to access tenant-specific endpoints:

```bash
# API tenant
curl -H "Host: api.example.com" http://localhost:8080/

# Admin tenant
curl -H "Host: admin.example.com" http://localhost:8080/
```

## Testing the Application

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Readiness check (checks all services)
curl http://localhost:8080/ready
```

### API Endpoints

```bash
# Get API status (no auth required)
curl http://localhost:8080/api/v1/status

# Get profile (requires authentication)
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/profile

# Update profile
curl -X PUT -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe"}' \
  http://localhost:8080/api/v1/profile
```

### Admin Endpoints

```bash
# List users (requires admin role)
curl -H "Authorization: Bearer <admin_token>" \
  http://localhost:8080/admin/users

# Create user
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com"}' \
  http://localhost:8080/admin/users
```

### WebSocket

```bash
# Connect with websocat
websocat ws://localhost:8080/ws

# Or use JavaScript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => console.log(event.data);
ws.send('Hello!');
```

### Metrics

```bash
# View Prometheus metrics
curl http://localhost:8080/metrics
```

### Profiling

Access profiling endpoints in your browser:

- `http://localhost:8080/debug/pprof` - Profiling index
- `http://localhost:8080/debug/pprof/heap` - Memory allocation
- `http://localhost:8080/debug/pprof/goroutine` - Goroutine stacks
- `http://localhost:8080/debug/pprof/profile` - 30s CPU profile

## Code Walkthrough

### Comprehensive Configuration

The example uses a complete configuration covering all framework features:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadTimeout:     15 * time.Second,
        WriteTimeout:    15 * time.Second,
        IdleTimeout:     120 * time.Second,
        MaxHeaderBytes:  2 << 20, // 2 MB
        EnableHTTP1:     true,
        EnableHTTP2:     true,
        EnableMetrics:   true,
        MetricsPath:     "/metrics",
        EnablePprof:     true,
        PprofPath:       "/debug/pprof",
        ShutdownTimeout: 30 * time.Second,
        // Multi-tenancy with host-specific configs
        HostConfigs: map[string]*pkg.HostConfig{
            "api.example.com": {
                Hostname: "api.example.com",
                TenantID: "tenant-1",
                RateLimits: &pkg.RateLimitConfig{
                    Enabled:           true,
                    RequestsPerSecond: 100,
                    BurstSize:         20,
                },
            },
        },
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:          "postgres",
        Host:            "localhost",
        Port:            5432,
        Database:        "rockstar_db",
        Username:        "rockstar_user",
        Password:        "rockstar_pass",
        MaxOpenConns:    25,
        MaxIdleConns:    5,
        ConnMaxLifetime: 5 * time.Minute,
    },
    // ... more configuration
}
```

**Configuration Highlights**:
- Production-ready timeouts and limits
- Multi-tenancy with per-host rate limiting
- Database connection pooling
- Metrics and profiling enabled
- Security features configured

### Lifecycle Hooks

Register multiple hooks for startup and shutdown:

```go
// Startup hooks - executed in order
app.RegisterStartupHook(func(ctx context.Context) error {
    fmt.Println("🚀 Initializing database connections...")
    return nil
})

app.RegisterStartupHook(func(ctx context.Context) error {
    fmt.Println("🚀 Loading configuration...")
    return nil
})

app.RegisterStartupHook(func(ctx context.Context) error {
    fmt.Println("🚀 Starting background workers...")
    return nil
})

// Shutdown hooks - executed in order
app.RegisterShutdownHook(func(ctx context.Context) error {
    fmt.Println("👋 Stopping background workers...")
    return nil
})

app.RegisterShutdownHook(func(ctx context.Context) error {
    fmt.Println("👋 Flushing metrics...")
    return nil
})

app.RegisterShutdownHook(func(ctx context.Context) error {
    fmt.Println("👋 Closing connections...")
    return nil
})
```

**Hook Use Cases**:
- **Startup**: Initialize services, load data, start workers, warm caches
- **Shutdown**: Stop workers, flush buffers, close connections, save state

### Global Middleware Stack

The example implements a comprehensive middleware stack:

```go
// 1. Request ID - adds unique ID to each request
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    requestID := generateRequestID()
    ctx.SetHeader("X-Request-ID", requestID)
    return next(ctx)
})

// 2. Logging - logs all requests
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    logger := ctx.Logger()
    
    logger.Info("Request started",
        "method", ctx.Request().Method,
        "path", ctx.Request().URL.Path,
    )
    
    err := next(ctx)
    
    duration := time.Since(start)
    logger.Info("Request completed", "duration", duration)
    
    return err
})

// 3. Metrics - records request metrics
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    err := next(ctx)
    duration := time.Since(start)
    
    metrics := ctx.Metrics()
    if metrics != nil {
        statusCode := ctx.Response().Status()
        metrics.RecordRequest(ctx, duration, statusCode)
    }
    
    return err
})

// 4. Recovery - recovers from panics
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            logger := ctx.Logger()
            logger.Error("Panic recovered", "panic", r)
            ctx.JSON(500, map[string]interface{}{
                "error": "Internal server error",
            })
        }
    }()
    return next(ctx)
})
```

**Middleware Order**:
1. Request ID (first) - identifies requests
2. Logging - logs request/response
3. Metrics - records performance data
4. Recovery (last) - catches panics

### Authentication Middleware

Protect routes with authentication:

```go
func authenticationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Authentication required",
        })
    }
    
    // In production, validate token:
    // user, err := ctx.Security().AuthenticateJWT(token)
    // if err != nil {
    //     return ctx.JSON(401, map[string]interface{}{"error": "Invalid token"})
    // }
    
    return next(ctx)
}
```

### Authorization Middleware

Check permissions with authorization:

```go
func adminAuthorizationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]interface{}{
            "error": "Authentication required",
        })
    }
    
    // Check if user has admin role
    if !ctx.IsAuthorized("admin", "access") {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Insufficient permissions",
        })
    }
    
    return next(ctx)
}
```

### Route Groups

Organize routes with groups and middleware:

```go
// Public API routes
v1 := router.Group("/api/v1")
v1.GET("/status", apiStatusHandler)

// Authenticated API routes
authAPI := router.Group("/api/v1", authenticationMiddleware)
authAPI.GET("/profile", profileHandler)
authAPI.PUT("/profile", updateProfileHandler)
authAPI.POST("/logout", logoutHandler)

// Admin routes (requires auth + admin role)
admin := router.Group("/admin", authenticationMiddleware, adminAuthorizationMiddleware)
admin.GET("/users", listUsersHandler)
admin.POST("/users", createUserHandler)
admin.GET("/users/:id", getUserHandler)
admin.PUT("/users/:id", updateUserHandler)
admin.DELETE("/users/:id", deleteUserHandler)
```

### Multi-Tenancy

Configure host-specific routing:

```go
// API tenant
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHomeHandler)

// Admin tenant
adminHost := router.Host("admin.example.com")
adminHost.GET("/", adminHomeHandler)
```

Access tenant information in handlers:

```go
func apiHomeHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]interface{}{
        "message": "API Host - api.example.com",
        "tenant":  ctx.Tenant(),
    })
}
```

### WebSocket Handler

Implement WebSocket echo server:

```go
func websocketHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    fmt.Println("WebSocket connection established")
    
    // Echo server - reads and sends back messages
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            fmt.Printf("WebSocket read error: %v\n", err)
            return err
        }
        
        fmt.Printf("Received: %s\n", string(data))
        
        // Echo message back
        if err := conn.WriteMessage(messageType, data); err != nil {
            fmt.Printf("WebSocket write error: %v\n", err)
            return err
        }
    }
}
```

### Metrics Handler

Expose Prometheus-compatible metrics:

```go
func metricsHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    ctx.Response().SetHeader("Content-Type", "text/plain; version=0.0.4")
    ctx.Response().WriteHeader(200)
    
    if metrics != nil {
        metricsData, err := metrics.ExportPrometheus()
        if err == nil && len(metricsData) > 0 {
            ctx.Response().Write(metricsData)
            return nil
        }
    }
    
    // Fallback: show basic runtime metrics
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    output := fmt.Sprintf(`# HELP go_goroutines Number of goroutines
# TYPE go_goroutines gauge
go_goroutines %d

# HELP go_memstats_alloc_bytes Allocated bytes
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes %d
`, runtime.NumGoroutine(), m.Alloc)
    
    ctx.Response().Write([]byte(output))
    return nil
}
```

### Profiling Handlers

The example includes comprehensive profiling endpoints:

- **Heap Profile**: Memory allocation analysis
- **Goroutine Profile**: Goroutine stack traces
- **CPU Profile**: 30-second CPU profiling
- **Trace**: Execution trace for detailed analysis
- **Block Profile**: Blocking operations
- **Mutex Profile**: Mutex contention

Access these at `/debug/pprof/*` endpoints.

### Graceful Shutdown

Implement graceful shutdown with signal handling:

```go
func setupGracefulShutdown(app *pkg.Framework) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        fmt.Println("\n🛑 Shutdown signal received...")
        
        if err := app.Shutdown(30 * time.Second); err != nil {
            log.Printf("Error during shutdown: %v", err)
        }
        
        os.Exit(0)
    }()
}
```

**Shutdown Process**:
1. Receive SIGINT or SIGTERM signal
2. Stop accepting new connections
3. Wait for active requests to complete (up to timeout)
4. Execute shutdown hooks
5. Close all connections
6. Exit cleanly

## Production Features

### Security Configuration

The example includes comprehensive security settings:

```go
SecurityConfig: pkg.SecurityConfig{
    XFrameOptions:    "SAMEORIGIN",
    EnableXSSProtect: true,
    EnableCSRF:       true,
    MaxRequestSize:   20 * 1024 * 1024, // 20 MB
    RequestTimeout:   60 * time.Second,
    AllowedOrigins:   []string{"https://example.com"},
    EncryptionKey:    "...",
    JWTSecret:        "...",
    CSRFTokenExpiry:  24 * time.Hour,
}
```

### Monitoring Configuration

Enable comprehensive monitoring:

```go
MonitoringConfig: pkg.MonitoringConfig{
    EnableMetrics:        true,
    MetricsPath:          "/metrics",
    EnablePprof:          true,
    PprofPath:            "/debug/pprof",
    EnableSNMP:           true,
    SNMPPort:             161,
    EnableOptimization:   true,
    OptimizationInterval: 10 * time.Second,
}
```

### Proxy Configuration

Configure load balancing and circuit breakers:

```go
ProxyConfig: pkg.ProxyConfig{
    LoadBalancerType:         "round_robin",
    CircuitBreakerEnabled:    true,
    CircuitBreakerThreshold:  5,
    CircuitBreakerTimeout:    30 * time.Second,
    MaxRetries:               3,
    RetryDelay:               1 * time.Second,
    RetryBackoff:             true,
    HealthCheckEnabled:       true,
    HealthCheckInterval:      30 * time.Second,
    HealthCheckTimeout:       5 * time.Second,
    MaxConnectionsPerBackend: 100,
    ConnectionTimeout:        10 * time.Second,
    IdleConnTimeout:          90 * time.Second,
}
```

### Session Configuration

Use database-backed sessions for scalability:

```go
SessionConfig: pkg.SessionConfig{
    StorageType:     pkg.SessionStorageDatabase,
    CookieName:      "rockstar_session_id",
    SessionLifetime: 24 * time.Hour,
    CookieSecure:    true,
    CookieHTTPOnly:  true,
    CookieSameSite:  "Strict",
    EncryptionKey:   []byte("..."), // 32 bytes
    CleanupInterval: 10 * time.Minute,
}
```

## Monitoring and Observability

### Prometheus Metrics

The `/metrics` endpoint exposes:
- Request count and duration
- Response status codes
- Goroutine count
- Memory usage
- GC statistics
- Custom application metrics

### Profiling

Use pprof for performance analysis:

```bash
# Capture 30s CPU profile
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof

# Analyze heap
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# View goroutines
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

### Logging

Structured logging with context:

```go
logger := ctx.Logger()
logger.Info("User action",
    "user_id", user.ID,
    "action", "login",
    "ip", ctx.Request().RemoteAddr,
)
```

## Performance Optimization

### Caching

Implement caching for frequently accessed data:

```go
// Check cache first
cacheKey := fmt.Sprintf("user:%d", userID)
if cached, err := ctx.Cache().Get(cacheKey); err == nil {
    return ctx.JSON(200, cached)
}

// Fetch from database
user := fetchUser(userID)

// Store in cache
ctx.Cache().Set(cacheKey, user, 5*time.Minute)

return ctx.JSON(200, user)
```

### Connection Pooling

Configure database connection pooling:

```go
DatabaseConfig: pkg.DatabaseConfig{
    MaxOpenConns:    25,              // Max open connections
    MaxIdleConns:    5,               // Max idle connections
    ConnMaxLifetime: 5 * time.Minute, // Connection lifetime
}
```

### Rate Limiting

Protect against abuse with rate limiting:

```go
RateLimits: &pkg.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 100,
    BurstSize:         20,
    Storage:           "database",
}
```

## Deployment Considerations

### Environment Variables

Use environment variables for configuration:

```go
config := pkg.FrameworkConfig{
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   os.Getenv("DB_DRIVER"),
        Host:     os.Getenv("DB_HOST"),
        Port:     getEnvInt("DB_PORT", 5432),
        Database: os.Getenv("DB_NAME"),
        Username: os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
    },
}
```

### Docker Deployment

Create a Dockerfile:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app examples/full_featured_app.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
```

### Kubernetes Deployment

Deploy with health checks:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rockstar-app
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: rockstar-app:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

## Common Issues

### Database Connection Errors

**Solution**: Verify database credentials and network connectivity

### Port Already in Use

**Solution**: Change the port or stop the conflicting process

### High Memory Usage

**Solution**: Check for memory leaks using heap profiling, adjust cache sizes

### Slow Requests

**Solution**: Use CPU profiling to identify bottlenecks, add caching

## Next Steps

After understanding this example:

1. **Customize for your needs**: Adapt the configuration and features
2. **Add business logic**: Implement your application-specific handlers
3. **Deploy to production**: Use Docker/Kubernetes for deployment
4. **Monitor in production**: Set up Prometheus and alerting
5. **Scale horizontally**: Add more instances behind a load balancer

## Related Documentation

- [Configuration Guide](../guides/configuration.md) - All configuration options
- [Security Guide](../guides/security.md) - Security best practices
- [Monitoring Guide](../guides/monitoring.md) - Monitoring and profiling
- [Deployment Guide](../guides/deployment.md) - Production deployment
- [Multi-Tenancy Guide](../guides/multi-tenancy.md) - Multi-tenant patterns

## Source Code

The complete source code for this example is available at `examples/full_featured_app.go` in the repository.
