---
title: "Debugging Guide"
description: "Debugging techniques and tools for Rockstar applications"
category: "troubleshooting"
tags: ["troubleshooting", "debugging", "profiling"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "common-errors.md"
  - "faq.md"
  - "../guides/monitoring.md"
---

# Debugging Guide

This guide provides comprehensive debugging techniques and tools for diagnosing and resolving issues in Rockstar Web Framework applications.

## Table of Contents

- [Logging Strategies](#logging-strategies)
- [Performance Profiling](#performance-profiling)
- [Database Debugging](#database-debugging)
- [Plugin Debugging](#plugin-debugging)
- [Network Debugging](#network-debugging)
- [Memory Debugging](#memory-debugging)
- [Concurrency Debugging](#concurrency-debugging)
- [Testing and Validation](#testing-and-validation)
- [Production Debugging](#production-debugging)

## Logging Strategies

### Enable Debug Logging

The first step in debugging is enabling detailed logging:

```go
config := pkg.FrameworkConfig{
    LogLevel: "debug",  // Options: debug, info, warn, error
    LogFormat: "json",  // Options: json, text
    LogOutput: os.Stdout,
}

app, err := pkg.New(config)
```

### Structured Logging

Use structured logging for better searchability:

```go
router.GET("/api/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Log with structured fields
    ctx.Logger().Info("Fetching user",
        "user_id", userID,
        "request_id", ctx.RequestID(),
        "tenant_id", ctx.TenantID(),
    )
    
    user, err := db.GetUser(userID)
    if err != nil {
        ctx.Logger().Error("Failed to fetch user",
            "user_id", userID,
            "error", err.Error(),
        )
        return err
    }
    
    return ctx.JSON(200, user)
})
```

### Log Levels

Use appropriate log levels:

```go
// DEBUG: Detailed information for diagnosing problems
ctx.Logger().Debug("Processing request", "path", ctx.Path())

// INFO: General informational messages
ctx.Logger().Info("User logged in", "user_id", userID)

// WARN: Warning messages for potentially harmful situations
ctx.Logger().Warn("Rate limit approaching", "requests", count)

// ERROR: Error messages for error events
ctx.Logger().Error("Database connection failed", "error", err)
```

### Request Logging

Enable request logging to track all incoming requests:

```go
config := pkg.FrameworkConfig{
    LogRequests: true,
    LogRequestBody: true,  // Be careful with sensitive data
    LogResponseBody: false, // Usually too verbose
}
```

### Custom Logger

Implement a custom logger for specific needs:

```go
type CustomLogger struct {
    *log.Logger
}

func (l *CustomLogger) Debug(msg string, fields ...interface{}) {
    l.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *CustomLogger) Info(msg string, fields ...interface{}) {
    l.Printf("[INFO] %s %v", msg, fields)
}

// Use custom logger
config := pkg.FrameworkConfig{
    Logger: &CustomLogger{
        Logger: log.New(os.Stdout, "", log.LstdFlags),
    },
}
```

### Log Filtering

Filter logs by component or request:

```bash
# Filter by log level
cat app.log | jq 'select(.level == "error")'

# Filter by component
cat app.log | jq 'select(.component == "database")'

# Filter by request ID
cat app.log | jq 'select(.request_id == "abc123")'

# Filter by tenant
cat app.log | jq 'select(.tenant_id == "tenant1")'
```

## Performance Profiling

### Enable Pprof

Enable Go's built-in profiling tools:

```go
import _ "net/http/pprof"

config := pkg.FrameworkConfig{
    MonitoringConfig: pkg.MonitoringConfig{
        EnablePprof: true,
        PprofPort: 6060,
    },
}
```

Access profiling endpoints at `http://localhost:6060/debug/pprof/`

### CPU Profiling

Profile CPU usage to find performance bottlenecks:

```bash
# Capture 30 seconds of CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with pprof
go tool pprof cpu.prof

# Interactive commands in pprof:
# top10 - Show top 10 functions by CPU time
# list functionName - Show source code with timing
# web - Generate visual graph (requires graphviz)
```

**Example pprof session:**

```bash
$ go tool pprof cpu.prof
(pprof) top10
Showing nodes accounting for 2.5s, 83.33% of 3s total
      flat  flat%   sum%        cum   cum%
     0.8s 26.67% 26.67%      0.8s 26.67%  runtime.mallocgc
     0.5s 16.67% 43.33%      0.5s 16.67%  database/sql.(*DB).query
     0.4s 13.33% 56.67%      0.4s 13.33%  encoding/json.Marshal
     
(pprof) list handlerFunction
Total: 3s
ROUTINE ======================== handlerFunction
     0.3s      1.2s (flat, cum) 40.00% of Total
         .          .     10:func handlerFunction(ctx pkg.Context) error {
     0.1s      0.2s     11:    data := fetchData()
     0.2s      1.0s     12:    return ctx.JSON(200, data)
         .          .     13:}
```

### Memory Profiling

Profile memory usage to find memory leaks:

```bash
# Capture heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze memory usage
go tool pprof heap.prof

# Commands:
# top10 - Top memory consumers
# list functionName - Source with allocations
# inuse_space - Currently allocated memory
# alloc_space - Total allocated memory
```

**Find memory leaks:**

```bash
# Take two heap snapshots
curl http://localhost:6060/debug/pprof/heap > heap1.prof
# ... wait and generate load ...
curl http://localhost:6060/debug/pprof/heap > heap2.prof

# Compare snapshots
go tool pprof -base heap1.prof heap2.prof
```

### Goroutine Profiling

Debug goroutine leaks:

```bash
# View goroutine stack traces
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Analyze goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

**Check for goroutine leaks:**

```go
// Monitor goroutine count
router.GET("/debug/goroutines", func(ctx pkg.Context) error {
    count := runtime.NumGoroutine()
    return ctx.JSON(200, map[string]int{
        "goroutines": count,
    })
})
```

### Block Profiling

Profile blocking operations:

```go
import "runtime"

func init() {
    // Enable block profiling
    runtime.SetBlockProfileRate(1)
}
```

```bash
# Capture block profile
curl http://localhost:6060/debug/pprof/block > block.prof

# Analyze blocking operations
go tool pprof block.prof
```

### Mutex Profiling

Profile mutex contention:

```go
import "runtime"

func init() {
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)
}
```

```bash
# Capture mutex profile
curl http://localhost:6060/debug/pprof/mutex > mutex.prof

# Analyze mutex contention
go tool pprof mutex.prof
```

### Trace Analysis

Use execution tracing for detailed analysis:

```bash
# Capture 5 seconds of trace
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out

# View trace in browser
go tool trace trace.out
```

The trace viewer shows:
- Goroutine execution timeline
- Network blocking
- Synchronization blocking
- System calls
- GC events

## Database Debugging

### Enable Query Logging

Log all database queries:

```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        Driver: "postgres",
        // ... connection details
        LogQueries: true,
        LogSlowQueries: true,
        SlowQueryThreshold: 100 * time.Millisecond,
    },
}
```

### Analyze Slow Queries

Identify and optimize slow queries:

```go
// Log query execution time
start := time.Now()
result, err := db.Query("SELECT * FROM users WHERE active = ?", true)
duration := time.Since(start)

if duration > 100*time.Millisecond {
    ctx.Logger().Warn("Slow query detected",
        "query", "SELECT * FROM users",
        "duration_ms", duration.Milliseconds(),
    )
}
```

### Connection Pool Monitoring

Monitor database connection pool:

```go
router.GET("/debug/db", func(ctx pkg.Context) error {
    db := ctx.DB()
    stats := db.Stats()
    
    return ctx.JSON(200, map[string]interface{}{
        "open_connections": stats.OpenConnections,
        "in_use": stats.InUse,
        "idle": stats.Idle,
        "wait_count": stats.WaitCount,
        "wait_duration": stats.WaitDuration,
        "max_idle_closed": stats.MaxIdleClosed,
        "max_lifetime_closed": stats.MaxLifetimeClosed,
    })
})
```

### Transaction Debugging

Debug transaction issues:

```go
func processOrder(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Start transaction with logging
    tx, err := db.Begin()
    if err != nil {
        ctx.Logger().Error("Failed to start transaction", "error", err)
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            ctx.Logger().Error("Transaction rolled back due to panic", "panic", p)
            panic(p)
        } else if err != nil {
            tx.Rollback()
            ctx.Logger().Warn("Transaction rolled back", "error", err)
        } else {
            err = tx.Commit()
            if err != nil {
                ctx.Logger().Error("Failed to commit transaction", "error", err)
            } else {
                ctx.Logger().Info("Transaction committed successfully")
            }
        }
    }()
    
    // Perform operations
    // ...
    
    return nil
}
```

### Query Explain Plans

Analyze query execution plans:

```go
// PostgreSQL
result, err := db.Query("EXPLAIN ANALYZE SELECT * FROM users WHERE email = ?", email)

// MySQL
result, err := db.Query("EXPLAIN SELECT * FROM users WHERE email = ?", email)

// Log the explain plan
for result.Next() {
    var plan string
    result.Scan(&plan)
    ctx.Logger().Debug("Query plan", "plan", plan)
}
```

## Plugin Debugging

### Enable Plugin Logging

Debug plugin loading and execution:

```go
config := pkg.FrameworkConfig{
    Plugins: &pkg.PluginConfig{
        Directory: "./plugins",
        Enabled: true,
        LogLevel: "debug",
    },
}
```

### Plugin Lifecycle Debugging

Track plugin lifecycle events:

```go
// In your plugin
func (p *MyPlugin) Init(ctx pkg.PluginContext) error {
    ctx.Logger().Info("Plugin initializing", "plugin", p.Name())
    
    // Debug plugin configuration
    config := ctx.Config()
    ctx.Logger().Debug("Plugin config", "config", config)
    
    return nil
}

func (p *MyPlugin) Start(ctx pkg.PluginContext) error {
    ctx.Logger().Info("Plugin starting", "plugin", p.Name())
    return nil
}

func (p *MyPlugin) Stop(ctx pkg.PluginContext) error {
    ctx.Logger().Info("Plugin stopping", "plugin", p.Name())
    return nil
}
```

### Plugin Dependency Debugging

Debug plugin dependency resolution:

```go
// List loaded plugins
router.GET("/debug/plugins", func(ctx pkg.Context) error {
    plugins := app.Plugins().List()
    
    pluginInfo := make([]map[string]interface{}, 0)
    for _, plugin := range plugins {
        info := map[string]interface{}{
            "name": plugin.Name(),
            "version": plugin.Version(),
            "status": plugin.Status(),
            "dependencies": plugin.Dependencies(),
        }
        pluginInfo = append(pluginInfo, info)
    }
    
    return ctx.JSON(200, pluginInfo)
})
```

### Plugin Error Isolation

Isolate plugin errors:

```go
// Plugin errors are isolated by default
// Check plugin health
router.GET("/debug/plugin/:name/health", func(ctx pkg.Context) error {
    pluginName := ctx.Param("name")
    plugin := app.Plugins().Get(pluginName)
    
    if plugin == nil {
        return ctx.JSON(404, map[string]string{"error": "Plugin not found"})
    }
    
    health := plugin.Health()
    return ctx.JSON(200, health)
})
```

## Network Debugging

### Request/Response Logging

Log complete request and response details:

```go
func debugMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Log request
    ctx.Logger().Debug("Incoming request",
        "method", ctx.Method(),
        "path", ctx.Path(),
        "headers", ctx.Headers(),
        "query", ctx.QueryParams(),
        "remote_addr", ctx.RemoteAddr(),
    )
    
    // Execute handler
    err := next(ctx)
    
    // Log response
    ctx.Logger().Debug("Outgoing response",
        "status", ctx.StatusCode(),
        "error", err,
    )
    
    return err
}

app.Use(debugMiddleware)
```

### Network Tracing

Use tcpdump or Wireshark for network-level debugging:

```bash
# Capture HTTP traffic
sudo tcpdump -i any -A 'tcp port 8080'

# Capture to file for Wireshark
sudo tcpdump -i any -w capture.pcap 'tcp port 8080'
```

### TLS Debugging

Debug TLS/SSL issues:

```go
import "crypto/tls"

config := pkg.FrameworkConfig{
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile: "key.pem",
        // Enable TLS debugging
        Config: &tls.Config{
            MinVersion: tls.VersionTLS12,
            // Log TLS handshake
            InsecureSkipVerify: false,
        },
    },
}
```

**Verify TLS configuration:**

```bash
# Test TLS connection
openssl s_client -connect localhost:8443 -tls1_2

# Check certificate
openssl x509 -in cert.pem -text -noout

# Verify certificate chain
openssl verify -CAfile ca.pem cert.pem
```

### WebSocket Debugging

Debug WebSocket connections:

```go
router.WebSocket("/ws", func(ctx pkg.Context) error {
    conn, err := ctx.UpgradeWebSocket()
    if err != nil {
        ctx.Logger().Error("WebSocket upgrade failed", "error", err)
        return err
    }
    defer conn.Close()
    
    ctx.Logger().Info("WebSocket connected",
        "remote_addr", ctx.RemoteAddr(),
        "user_id", ctx.UserID(),
    )
    
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            ctx.Logger().Warn("WebSocket read error", "error", err)
            break
        }
        
        ctx.Logger().Debug("WebSocket message received",
            "type", messageType,
            "size", len(message),
        )
        
        // Process message
        // ...
    }
    
    ctx.Logger().Info("WebSocket disconnected")
    return nil
})
```

## Memory Debugging

### Memory Leak Detection

Detect memory leaks using heap profiling:

```bash
# Take baseline heap snapshot
curl http://localhost:6060/debug/pprof/heap > heap_baseline.prof

# Generate load and wait
# ...

# Take second snapshot
curl http://localhost:6060/debug/pprof/heap > heap_after.prof

# Compare to find leaks
go tool pprof -base heap_baseline.prof heap_after.prof
```

### Memory Statistics

Monitor memory usage:

```go
import "runtime"

router.GET("/debug/memory", func(ctx pkg.Context) error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return ctx.JSON(200, map[string]interface{}{
        "alloc_mb": m.Alloc / 1024 / 1024,
        "total_alloc_mb": m.TotalAlloc / 1024 / 1024,
        "sys_mb": m.Sys / 1024 / 1024,
        "num_gc": m.NumGC,
        "gc_cpu_fraction": m.GCCPUFraction,
    })
})
```

### Force Garbage Collection

Force GC for testing:

```go
import "runtime"

router.POST("/debug/gc", func(ctx pkg.Context) error {
    runtime.GC()
    return ctx.JSON(200, map[string]string{"status": "gc triggered"})
})
```

### Memory Profiling in Tests

Profile memory in tests:

```go
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    // Run test code
    for i := 0; i < 1000; i++ {
        // ... test operations
    }
    
    runtime.ReadMemStats(&m2)
    
    allocated := m2.TotalAlloc - m1.TotalAlloc
    t.Logf("Memory allocated: %d bytes", allocated)
    
    if allocated > 10*1024*1024 { // 10MB
        t.Errorf("Excessive memory allocation: %d bytes", allocated)
    }
}
```

## Concurrency Debugging

### Race Detector

Use Go's race detector to find data races:

```bash
# Run with race detector
go run -race main.go

# Test with race detector
go test -race ./...

# Build with race detector
go build -race -o app main.go
```

**Example race detection:**

```go
// This code has a race condition
var counter int

func increment() {
    counter++ // Race: concurrent access
}

// Run with -race flag to detect
```

### Deadlock Detection

Detect deadlocks:

```go
import "time"

func detectDeadlock() {
    timeout := time.After(30 * time.Second)
    
    select {
    case <-done:
        // Operation completed
    case <-timeout:
        // Potential deadlock
        log.Fatal("Operation timed out - possible deadlock")
    }
}
```

### Goroutine Leak Detection

Detect goroutine leaks in tests:

```go
func TestNoGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // Run test code
    // ...
    
    // Wait for goroutines to finish
    time.Sleep(100 * time.Millisecond)
    
    after := runtime.NumGoroutine()
    
    if after > before {
        t.Errorf("Goroutine leak detected: %d -> %d", before, after)
    }
}
```

## Testing and Validation

### Unit Testing with Debugging

Add debugging to unit tests:

```go
func TestHandler(t *testing.T) {
    // Enable debug logging in tests
    config := pkg.FrameworkConfig{
        LogLevel: "debug",
        LogOutput: os.Stdout,
    }
    
    app, err := pkg.New(config)
    if err != nil {
        t.Fatal(err)
    }
    
    // Test with logging
    // ...
}
```

### Integration Testing

Debug integration tests:

```go
func TestIntegration(t *testing.T) {
    // Start test server
    app := setupTestApp(t)
    
    // Enable request logging
    app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
        t.Logf("Request: %s %s", ctx.Method(), ctx.Path())
        err := next(ctx)
        t.Logf("Response: %d", ctx.StatusCode())
        return err
    })
    
    // Run tests
    // ...
}
```

### Benchmark Debugging

Debug performance issues in benchmarks:

```go
func BenchmarkHandler(b *testing.B) {
    app := setupApp()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
    
    b.StopTimer()
    
    // Report custom metrics
    b.ReportMetric(float64(allocations), "allocs/op")
}
```

## Production Debugging

### Safe Production Debugging

Debug production issues safely:

```go
// Enable debug endpoints only for admin users
router.GET("/debug/*", func(ctx pkg.Context) error {
    if !ctx.IsAdmin() {
        return pkg.NewAuthorizationError("Admin access required")
    }
    
    // Serve debug information
    return next(ctx)
}, pkg.RequireRole("admin"))
```

### Remote Debugging

Enable remote debugging securely:

```go
import "net/http/pprof"

// Separate debug server
go func() {
    mux := http.NewServeMux()
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    // Only listen on localhost
    log.Fatal(http.ListenAndServe("localhost:6060", mux))
}()
```

### Production Logging

Configure production-safe logging:

```go
config := pkg.FrameworkConfig{
    LogLevel: "info",  // Don't use debug in production
    LogFormat: "json",
    LogOutput: logFile,
    LogErrors: true,
    LogRequests: true,
    LogRequestBody: false,  // Don't log sensitive data
    LogResponseBody: false,
}
```

### Error Tracking

Integrate with error tracking services:

```go
import "github.com/getsentry/sentry-go"

func init() {
    sentry.Init(sentry.ClientOptions{
        Dsn: os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENVIRONMENT"),
    })
}

// Error middleware
func errorTrackingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    err := next(ctx)
    
    if err != nil {
        sentry.CaptureException(err)
    }
    
    return err
}
```

## Debugging Tools

### Delve Debugger

Use Delve for interactive debugging:

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug application
dlv debug main.go

# Debug with arguments
dlv debug main.go -- --config=config.yaml

# Attach to running process
dlv attach <pid>
```

**Delve commands:**

```bash
# Set breakpoint
(dlv) break main.handler

# Continue execution
(dlv) continue

# Step through code
(dlv) next
(dlv) step

# Print variables
(dlv) print variableName

# List goroutines
(dlv) goroutines

# Switch goroutine
(dlv) goroutine <id>
```

### VS Code Debugging

Configure VS Code for debugging:

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "env": {
                "LOG_LEVEL": "debug"
            },
            "args": []
        }
    ]
}
```

### GoLand Debugging

Use GoLand's built-in debugger:

1. Set breakpoints by clicking in the gutter
2. Right-click main.go → Debug 'go build main.go'
3. Use debug toolbar to step through code
4. Inspect variables in the Variables pane
5. Evaluate expressions in the Evaluate window

## Best Practices

### Debugging Checklist

When debugging an issue:

1. **Reproduce the issue** consistently
2. **Enable debug logging** for relevant components
3. **Isolate the problem** to a specific component
4. **Check logs** for error messages and stack traces
5. **Use profiling** if it's a performance issue
6. **Write a test** that reproduces the issue
7. **Fix the issue** and verify with the test
8. **Add logging** to prevent future issues

### Performance Debugging Workflow

1. **Measure baseline** performance
2. **Identify bottleneck** using profiling
3. **Optimize** the bottleneck
4. **Measure again** to verify improvement
5. **Repeat** until performance goals are met

### Production Debugging Guidelines

- **Never** enable debug logging in production by default
- **Always** sanitize logs to remove sensitive data
- **Use** structured logging for better searchability
- **Implement** proper error tracking
- **Monitor** key metrics continuously
- **Set up** alerts for critical issues

## Additional Resources

- [Monitoring Guide](../guides/monitoring.md) - Metrics and monitoring
- [Performance Guide](../guides/performance.md) - Performance optimization
- [Common Errors](common-errors.md) - Error messages and solutions
- [FAQ](faq.md) - Frequently asked questions

## Navigation

- [← Back to Common Errors](common-errors.md)
- [Next: FAQ →](faq.md)
- [← Back to Troubleshooting](README.md)
