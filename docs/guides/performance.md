# Performance Optimization

## Overview

The Rockstar Web Framework is designed for high performance with built-in optimizations for memory management, request handling, and resource pooling. This guide covers the framework's performance characteristics, optimization techniques, caching strategies, and best practices for building fast, scalable applications.

**Key performance features:**
- **Arena-Based Memory Management**: Request-scoped memory allocation
- **Connection Pooling**: Efficient database connection reuse
- **Multi-Level Caching**: Request, application, and distributed caching
- **Efficient Routing**: Fast route matching and parameter extraction
- **Resource Pooling**: Buffer and object reuse
- **Automatic Optimization**: Process-guided memory optimization
- **Protocol Support**: HTTP/1, HTTP/2, and QUIC for optimal performance

## Quick Start

Enable performance optimizations:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
    "time"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP2:     true,
            MaxConnections:  10000,
            ReadBufferSize:  8192,
            WriteBufferSize: 8192,
        },
        CacheConfig: pkg.CacheConfig{
            Type:       "memory",
            MaxSize:    100 * 1024 * 1024, // 100MB
            DefaultTTL: 5 * time.Minute,
        },
        MonitoringConfig: pkg.MonitoringConfig{
            EnableOptimization:   true,
            OptimizationInterval: 5 * time.Minute,
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Fatal(app.Listen(":8080"))
}
```

## Framework Performance Characteristics

### Request Handling

The framework is optimized for high-throughput request handling:

**Typical Performance Metrics:**
- **Throughput**: 50,000+ requests/second (simple endpoints)
- **Latency**: <1ms P50, <5ms P99 (without database)
- **Memory**: ~1KB per request (arena-based allocation)
- **Concurrency**: 10,000+ concurrent connections

**Performance Factors:**
- Handler complexity
- Database queries
- External API calls
- Response size
- Middleware chain length

### Memory Management

The framework uses arena-based memory management for request handling:

```go
// Request-scoped memory is automatically managed
router.GET("/api/users", func(ctx pkg.Context) error {
    // Memory allocated here is freed when request completes
    users := fetchUsers()
    
    // Request cache uses arena allocation
    cache := ctx.RequestCache()
    cache.Set("users", users)
    
    return ctx.JSON(200, users)
})
```

**Memory Benefits:**
- Reduced GC pressure
- Predictable memory usage
- Automatic cleanup after request
- No memory leaks from request handling

### Connection Pooling

Database connections are automatically pooled:

```go
config := pkg.DatabaseConfig{
    Driver:          "postgres",
    MaxOpenConns:    50,  // Maximum open connections
    MaxIdleConns:    10,  // Maximum idle connections
    ConnMaxLifetime: 10 * time.Minute,
}
```

**Pooling Benefits:**
- Reduced connection overhead
- Better resource utilization
- Automatic connection recycling
- Protection against connection exhaustion

## Caching Strategies

### Request-Level Caching

Cache data for the duration of a single request:

```go
router.GET("/api/dashboard", func(ctx pkg.Context) error {
    cache := ctx.RequestCache()
    
    // Check request cache first
    if user, exists := cache.Get("current_user"); exists {
        return ctx.JSON(200, user)
    }
    
    // Fetch and cache for this request
    user := fetchUser(ctx.User().ID)
    cache.Set("current_user", user)
    
    return ctx.JSON(200, user)
})
```

**Use Cases:**
- Avoiding duplicate database queries in a single request
- Sharing computed values across middleware
- Temporary data that doesn't need persistence

### Application-Level Caching

Cache data across multiple requests:

```go
router.GET("/api/config", func(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Check application cache
    if config, err := cache.Get("app_config"); err == nil {
        return ctx.JSON(200, config)
    }
    
    // Fetch and cache for 5 minutes
    config := loadConfig()
    cache.Set("app_config", config, 5*time.Minute)
    
    return ctx.JSON(200, config)
})
```

**Use Cases:**
- Configuration data
- Reference data (countries, categories)
- Computed results
- API responses

### Cache-Aside Pattern

Implement cache-aside (lazy loading):

```go
func getUser(ctx pkg.Context, userID string) (*User, error) {
    cache := ctx.Cache()
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // Try cache first
    if cached, err := cache.Get(cacheKey); err == nil {
        return cached.(*User), nil
    }
    
    // Cache miss - fetch from database
    db := ctx.Database()
    user, err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for 10 minutes
    cache.Set(cacheKey, user, 10*time.Minute)
    
    return user, nil
}
```

### Write-Through Caching

Update cache when data changes:

```go
router.PUT("/api/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return err
    }
    
    // Update database
    db := ctx.Database()
    if err := db.Exec("UPDATE users SET ... WHERE id = ?", userID); err != nil {
        return err
    }
    
    // Update cache
    cache := ctx.Cache()
    cacheKey := fmt.Sprintf("user:%s", userID)
    cache.Set(cacheKey, &user, 10*time.Minute)
    
    return ctx.JSON(200, user)
})
```

### Cache Invalidation

Invalidate cache entries when data changes:

```go
// Invalidate specific key
cache.Delete("user:123")

// Invalidate by pattern
cache.Invalidate("user:*")

// Invalidate by tag
cache.SetWithTags("product:123", product, 10*time.Minute, []string{"products", "category:electronics"})
cache.InvalidateTag("products") // Invalidates all products
```

### Multi-Level Caching

Combine request and application caching:

```go
func getExpensiveData(ctx pkg.Context, key string) (interface{}, error) {
    // Level 1: Request cache (fastest)
    reqCache := ctx.RequestCache()
    if data, exists := reqCache.Get(key); exists {
        return data, nil
    }
    
    // Level 2: Application cache
    appCache := ctx.Cache()
    if data, err := appCache.Get(key); err == nil {
        reqCache.Set(key, data) // Promote to request cache
        return data, nil
    }
    
    // Level 3: Database (slowest)
    data := fetchFromDatabase(key)
    
    // Populate caches
    appCache.Set(key, data, 10*time.Minute)
    reqCache.Set(key, data)
    
    return data, nil
}
```

## Database Optimization

### Connection Pool Tuning

Optimize connection pool settings:

```go
config := pkg.DatabaseConfig{
    Driver:   "postgres",
    Host:     "db.example.com",
    Database: "myapp",
    
    // Connection pool settings
    MaxOpenConns:    50,              // Max concurrent connections
    MaxIdleConns:    10,              // Idle connections to keep
    ConnMaxLifetime: 10 * time.Minute, // Recycle connections
}
```

**Tuning Guidelines:**
- `MaxOpenConns`: Set to ~2x CPU cores for CPU-bound workloads
- `MaxIdleConns`: Set to ~25% of MaxOpenConns
- `ConnMaxLifetime`: 5-15 minutes to prevent stale connections

### Query Optimization

Use prepared statements and efficient queries:

```go
// Good: Use prepared statements
db := ctx.Database()
users, err := db.Query(
    "SELECT id, name, email FROM users WHERE status = ? LIMIT ?",
    "active", 100,
)

// Good: Select only needed columns
user, err := db.QueryRow(
    "SELECT id, name FROM users WHERE id = ?",
    userID,
)

// Bad: Select all columns when not needed
user, err := db.QueryRow(
    "SELECT * FROM users WHERE id = ?",
    userID,
)
```

### Batch Operations

Batch database operations for better performance:

```go
// Good: Batch insert
db := ctx.Database()
tx, _ := db.Begin()

stmt, _ := tx.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
for _, user := range users {
    stmt.Exec(user.Name, user.Email)
}

tx.Commit()

// Bad: Individual inserts
for _, user := range users {
    db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", user.Name, user.Email)
}
```

### Index Usage

Ensure proper database indexing:

```sql
-- Index frequently queried columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_created_at ON orders(created_at);

-- Composite indexes for multi-column queries
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
```

## Memory Optimization

### Request Memory Management

The framework automatically manages request memory:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Memory allocated here is automatically freed
    // when the request completes
    
    data := make([]byte, 1024*1024) // 1MB allocation
    // ... use data
    
    return ctx.JSON(200, result)
    // Memory freed here automatically
})
```

### Avoid Memory Leaks

Common memory leak patterns to avoid:

```go
// Bad: Goroutine leak
router.GET("/api/process", func(ctx pkg.Context) error {
    go func() {
        // This goroutine may never exit
        for {
            doWork()
        }
    }()
    return ctx.JSON(200, "started")
})

// Good: Controlled goroutine lifecycle
router.GET("/api/process", func(ctx pkg.Context) error {
    done := make(chan struct{})
    go func() {
        defer close(done)
        for {
            select {
            case <-done:
                return
            default:
                doWork()
            }
        }
    }()
    
    // Cleanup when request context is done
    go func() {
        <-ctx.Request().Context().Done()
        close(done)
    }()
    
    return ctx.JSON(200, "started")
})
```

### Buffer Pooling

Reuse buffers to reduce allocations:

```go
import "sync"

var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

router.GET("/api/data", func(ctx pkg.Context) error {
    // Get buffer from pool
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // Use buffer
    n := copy(buf, data)
    
    return ctx.Bytes(200, buf[:n])
})
```

### Automatic Optimization

Enable automatic memory optimization:

```go
config := pkg.MonitoringConfig{
    EnableOptimization:   true,
    OptimizationInterval: 5 * time.Minute,
}

// Or trigger manually
monitoring := app.Monitoring()
monitoring.OptimizeNow()
```

## Response Optimization

### Compression

Enable response compression:

```go
// Compression middleware (if available)
router.Use(compressionMiddleware)

router.GET("/api/large-data", func(ctx pkg.Context) error {
    // Response will be automatically compressed
    return ctx.JSON(200, largeDataset)
})
```

### Streaming Responses

Stream large responses:

```go
router.GET("/api/export", func(ctx pkg.Context) error {
    ctx.Response().SetHeader("Content-Type", "text/csv")
    ctx.Response().SetHeader("Transfer-Encoding", "chunked")
    
    writer := ctx.Response().Writer()
    
    // Stream data in chunks
    for row := range fetchRows() {
        fmt.Fprintf(writer, "%s\n", row)
        if flusher, ok := writer.(http.Flusher); ok {
            flusher.Flush()
        }
    }
    
    return nil
})
```

### Response Caching

Cache entire responses:

```go
func cacheMiddleware(ctx pkg.Context) error {
    cache := ctx.Cache()
    cacheKey := fmt.Sprintf("response:%s", ctx.Request().RequestURI)
    
    // Check cache
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Continue to handler
    if err := ctx.Next(); err != nil {
        return err
    }
    
    // Cache response (implement response capture)
    // cache.Set(cacheKey, response, 5*time.Minute)
    
    return nil
}
```

## Protocol Optimization

### HTTP/2 Support

Enable HTTP/2 for better performance:

```go
config := pkg.ServerConfig{
    EnableHTTP2: true,
    TLSEnabled:  true, // Required for HTTP/2
    TLSCertFile: "cert.pem",
    TLSKeyFile:  "key.pem",
}
```

**HTTP/2 Benefits:**
- Multiplexing: Multiple requests over single connection
- Header compression: Reduced overhead
- Server push: Proactive resource delivery
- Binary protocol: More efficient parsing

### QUIC Support

Enable QUIC for low-latency connections:

```go
config := pkg.ServerConfig{
    EnableQUIC:  true,
    TLSEnabled:  true,
    TLSCertFile: "cert.pem",
    TLSKeyFile:  "key.pem",
}
```

**QUIC Benefits:**
- 0-RTT connection establishment
- Better handling of packet loss
- Connection migration
- Improved mobile performance

### Keep-Alive Tuning

Optimize keep-alive settings:

```go
config := pkg.ServerConfig{
    IdleTimeout:    120 * time.Second, // Keep connections alive
    ReadTimeout:    30 * time.Second,
    WriteTimeout:   30 * time.Second,
    MaxConnections: 10000,
}
```

## Concurrency Optimization

### Goroutine Management

Use goroutines efficiently:

```go
// Good: Limited concurrency
router.POST("/api/batch", func(ctx pkg.Context) error {
    var items []Item
    ctx.BindJSON(&items)
    
    // Process with limited concurrency
    sem := make(chan struct{}, 10) // Max 10 concurrent
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        sem <- struct{}{} // Acquire
        
        go func(item Item) {
            defer wg.Done()
            defer func() { <-sem }() // Release
            
            processItem(item)
        }(item)
    }
    
    wg.Wait()
    return ctx.JSON(200, "completed")
})
```

### Worker Pools

Use worker pools for background tasks:

```go
type WorkerPool struct {
    tasks chan func()
    wg    sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        tasks: make(chan func(), 100),
    }
    
    for i := 0; i < workers; i++ {
        pool.wg.Add(1)
        go pool.worker()
    }
    
    return pool
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for task := range p.tasks {
        task()
    }
}

func (p *WorkerPool) Submit(task func()) {
    p.tasks <- task
}

func (p *WorkerPool) Close() {
    close(p.tasks)
    p.wg.Wait()
}
```

## Benchmarking

### Built-in Benchmarks

Run framework benchmarks:

```bash
# Run all benchmarks
cd tests
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkRouting -benchmem

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Custom Benchmarks

Write benchmarks for your handlers:

```go
func BenchmarkUserHandler(b *testing.B) {
    app, _ := pkg.New(pkg.DefaultConfig())
    router := app.Router()
    
    router.GET("/api/users/:id", userHandler)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        req := httptest.NewRequest("GET", "/api/users/123", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
    }
}
```

### Load Testing

Use load testing tools:

```bash
# Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/users

# wrk
wrk -t12 -c400 -d30s http://localhost:8080/api/users

# hey
hey -n 10000 -c 100 http://localhost:8080/api/users
```

### Profiling

Profile your application:

```bash
# CPU profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Memory profiling
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Goroutine profiling
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

## Performance Best Practices

### 1. Use Appropriate Data Structures

```go
// Good: Use map for lookups
userMap := make(map[string]*User)
user := userMap[userID] // O(1)

// Bad: Use slice for lookups
for _, user := range users { // O(n)
    if user.ID == userID {
        return user
    }
}
```

### 2. Minimize Allocations

```go
// Good: Preallocate slices
users := make([]User, 0, expectedSize)

// Bad: Grow slice dynamically
var users []User
```

### 3. Avoid Unnecessary Conversions

```go
// Good: Work with bytes
data := []byte("hello")
processBytes(data)

// Bad: Convert unnecessarily
data := []byte("hello")
str := string(data)
processString(str)
```

### 4. Use Efficient JSON Encoding

```go
// Good: Stream JSON encoding
encoder := json.NewEncoder(w)
encoder.Encode(data)

// Less efficient: Marshal to bytes first
bytes, _ := json.Marshal(data)
w.Write(bytes)
```

### 5. Batch Database Operations

```go
// Good: Single query with IN clause
db.Query("SELECT * FROM users WHERE id IN (?, ?, ?)", id1, id2, id3)

// Bad: Multiple queries
db.Query("SELECT * FROM users WHERE id = ?", id1)
db.Query("SELECT * FROM users WHERE id = ?", id2)
db.Query("SELECT * FROM users WHERE id = ?", id3)
```

### 6. Use Context Timeouts

```go
// Good: Set timeouts
ctx, cancel := context.WithTimeout(ctx.Request().Context(), 5*time.Second)
defer cancel()

result, err := externalAPI.Call(ctx)
```

### 7. Cache Expensive Operations

```go
// Good: Cache computed results
cache := ctx.Cache()
if result, err := cache.Get("expensive_computation"); err == nil {
    return result
}

result := expensiveComputation()
cache.Set("expensive_computation", result, 10*time.Minute)
```

### 8. Optimize Middleware Chain

```go
// Good: Order middleware by frequency
router.Use(loggingMiddleware)    // Always runs
router.Use(authMiddleware)       // Most requests
router.Use(adminMiddleware)      // Few requests

// Bad: Expensive middleware first
router.Use(expensiveMiddleware)  // Runs for all requests
router.Use(loggingMiddleware)
```

## Performance Monitoring

### Key Metrics to Track

Monitor these performance metrics:

```go
// Request metrics
- Request rate (requests/second)
- Response time (P50, P95, P99)
- Error rate (errors/second)

// Resource metrics
- CPU usage (%)
- Memory usage (bytes)
- Goroutine count
- GC pause time

// Database metrics
- Query duration
- Connection pool usage
- Slow query count
```

### Setting Performance Targets

Define performance SLOs:

```yaml
# Service Level Objectives
availability: 99.9%
latency_p50: 50ms
latency_p95: 200ms
latency_p99: 500ms
error_rate: <0.1%
```

### Continuous Monitoring

Set up continuous monitoring:

```go
// Record metrics for every request
router.Use(func(ctx pkg.Context) error {
    start := time.Now()
    
    err := ctx.Next()
    
    duration := time.Since(start)
    metrics := ctx.Metrics()
    metrics.RecordRequest(ctx, duration, ctx.Response().StatusCode)
    
    return err
})
```

## Troubleshooting Performance Issues

### High Latency

**Symptoms**: Slow response times

**Solutions**:
- Profile with pprof to identify bottlenecks
- Check database query performance
- Review external API call timeouts
- Optimize expensive computations
- Add caching where appropriate

### High Memory Usage

**Symptoms**: Growing memory consumption

**Solutions**:
- Check for memory leaks with pprof heap profile
- Enable automatic optimization
- Review goroutine lifecycle
- Optimize data structures
- Implement proper cleanup

### High CPU Usage

**Symptoms**: Elevated CPU utilization

**Solutions**:
- Profile with pprof CPU profile
- Optimize hot code paths
- Reduce unnecessary computations
- Use caching for repeated operations
- Review algorithm complexity

### Connection Pool Exhaustion

**Symptoms**: Database connection errors

**Solutions**:
- Increase MaxOpenConns
- Reduce connection lifetime
- Fix connection leaks
- Optimize query performance
- Implement connection retry logic

## See Also

- [Monitoring Guide](monitoring.md) - Metrics and profiling
- [Caching Guide](caching.md) - Caching strategies
- [Database Guide](database.md) - Database optimization
- [Deployment Guide](deployment.md) - Production optimization

