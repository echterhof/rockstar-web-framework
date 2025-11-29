# Monitoring and Metrics

## Overview

The Rockstar Web Framework provides comprehensive monitoring and metrics capabilities to help you understand your application's performance, health, and resource usage. Built-in support for metrics collection, Prometheus integration, pprof profiling, and SNMP monitoring makes it easy to observe and optimize your applications in production.

**Key monitoring features:**
- **Metrics Collection**: Automatic request, response, and system metrics
- **Prometheus Integration**: Export metrics in Prometheus format
- **Pprof Profiling**: CPU, memory, and goroutine profiling
- **SNMP Support**: Network monitoring protocol support
- **Workload Metrics**: Detailed per-request performance tracking
- **Health Checks**: Application and component health monitoring
- **Process Optimization**: Automatic memory and GC optimization

## Quick Start

Enable monitoring with minimal configuration:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
)

func main() {
    config := pkg.FrameworkConfig{
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics: true,
            MetricsPort:   9090,
            EnablePprof:   true,
            PprofPort:     6060,
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Metrics available at http://localhost:9090/metrics
    // Pprof available at http://localhost:6060/debug/pprof
    
    log.Fatal(app.Listen(":8080"))
}
```

## Metrics Collection

### Automatic Request Metrics

The framework automatically collects metrics for every request:

```go
// Metrics are collected automatically for all requests
router := app.Router()

router.GET("/api/users", func(ctx pkg.Context) error {
    // Request metrics collected automatically:
    // - Duration
    // - Memory usage
    // - CPU usage
    // - Status code
    // - Response size
    
    return ctx.JSON(200, users)
})
```

**Automatically collected metrics:**
- Request duration (milliseconds)
- Memory allocation per request
- CPU usage percentage
- HTTP status code
- Response size (bytes)
- Request path and method
- Tenant ID (if multi-tenant)
- User ID (if authenticated)
- Error messages (if any)

### Custom Metrics

Record custom application metrics:

```go
router.GET("/api/process", func(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Increment a counter
    metrics.IncrementCounter("orders.processed", map[string]string{
        "type": "online",
    })
    
    // Set a gauge value
    metrics.SetGauge("queue.size", float64(queueSize), map[string]string{
        "queue": "orders",
    })
    
    // Record a histogram value
    metrics.RecordHistogram("order.value", orderValue, map[string]string{
        "currency": "USD",
    })
    
    // Record timing
    metrics.RecordTiming("database.query", queryDuration, map[string]string{
        "table": "orders",
    })
    
    return ctx.JSON(200, result)
})
```

### Timer Metrics

Use timers to measure operation duration:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Start a timer
    timer := metrics.StartTimer("data.processing", map[string]string{
        "operation": "fetch",
    })
    
    // Perform operation
    data := fetchData()
    
    // Stop timer and record duration
    duration := timer.Stop()
    
    log.Printf("Operation took %v", duration)
    
    return ctx.JSON(200, data)
})
```

### Workload Metrics

Access detailed workload metrics for analysis:

```go
// Get metrics for a specific time range
metricsCollector := app.Metrics()

from := time.Now().Add(-1 * time.Hour)
to := time.Now()

workloadMetrics, err := metricsCollector.GetMetrics("tenant-123", from, to)
if err != nil {
    log.Fatal(err)
}

for _, m := range workloadMetrics {
    log.Printf("Request %s: %dms, %d bytes memory",
        m.RequestID, m.Duration, m.MemoryUsage)
}
```

### Aggregated Metrics

Get aggregated metrics for performance analysis:

```go
// Get aggregated metrics
agg, err := metricsCollector.GetAggregatedMetrics("tenant-123", from, to)
if err != nil {
    log.Fatal(err)
}

log.Printf("Total Requests: %d", agg.TotalRequests)
log.Printf("Success Rate: %.2f%%", 
    float64(agg.SuccessfulRequests)/float64(agg.TotalRequests)*100)
log.Printf("Avg Duration: %v", agg.AvgDuration)
log.Printf("Avg Memory: %d bytes", agg.AvgMemoryUsage)
log.Printf("Requests/sec: %.2f", agg.RequestsPerSecond)
```

## Metrics Endpoint

### Enabling the Metrics Endpoint

```go
config := pkg.MonitoringConfig{
    EnableMetrics: true,
    MetricsPath:   "/metrics",  // Default
    MetricsPort:   9090,         // Default
}

app, _ := pkg.New(pkg.FrameworkConfig{
    MonitoringConfig: config,
})
```

Access metrics at `http://localhost:9090/metrics`

### Metrics Response Format

The metrics endpoint returns JSON with all collected metrics:

```json
{
  "counters": {
    "http.requests,method=GET,path=/api/users,status=200": 1523,
    "http.requests,method=POST,path=/api/orders,status=201": 342,
    "http.errors,error=database timeout": 5
  },
  "gauges": {
    "queue.size,queue=orders": 42,
    "system.memory.usage": 134217728,
    "system.cpu.usage": 23.5
  },
  "system": {
    "num_cpu": 8,
    "num_goroutine": 156,
    "memory_alloc": 134217728,
    "memory_total_alloc": 2147483648,
    "memory_sys": 268435456,
    "num_gc": 42,
    "gc_pause_total_ns": 1234567890,
    "last_gc_time": "2025-01-15T10:30:00Z"
  }
}
```

### Securing the Metrics Endpoint

Require authentication for metrics access:

```go
config := pkg.MonitoringConfig{
    EnableMetrics: true,
    MetricsPort:   9090,
    RequireAuth:   true,
    AuthToken:     "your-secret-token",
}
```

Access with authentication:

```bash
curl -H "Authorization: Bearer your-secret-token" \
    http://localhost:9090/metrics
```

## Prometheus Integration

### Exporting Prometheus Metrics

The framework supports Prometheus metric export:

```go
// Enable metrics endpoint
config := pkg.MonitoringConfig{
    EnableMetrics: true,
    MetricsPort:   9090,
}

app, _ := pkg.New(pkg.FrameworkConfig{
    MonitoringConfig: config,
})

// Get Prometheus-formatted metrics
router := app.Router()
router.GET("/prometheus", func(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    prometheusData, err := metrics.ExportPrometheus()
    if err != nil {
        return err
    }
    
    ctx.Response().SetHeader("Content-Type", "text/plain")
    return ctx.String(200, string(prometheusData))
})
```

### Prometheus Configuration

Configure Prometheus to scrape your application:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'rockstar-app'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9090']
        labels:
          app: 'my-rockstar-app'
          environment: 'production'
```

### Prometheus Metric Format

Metrics are exported in Prometheus text format:

```
# TYPE http_requests counter
http_requests,method=GET,path=/api/users,status=200 1523
http_requests,method=POST,path=/api/orders,status=201 342

# TYPE queue_size gauge
queue_size,queue=orders 42.000000

# TYPE system_memory_usage gauge
system_memory_usage 134217728.000000
```

### Common Prometheus Queries

Useful PromQL queries for your application:

```promql
# Request rate
rate(http_requests[5m])

# Error rate
rate(http_errors[5m])

# Average response time
avg(http_request_duration) by (path)

# 95th percentile response time
histogram_quantile(0.95, http_request_duration)

# Memory usage trend
system_memory_usage

# CPU usage
system_cpu_usage
```

## Profiling with Pprof

### Enabling Pprof

```go
config := pkg.MonitoringConfig{
    EnablePprof: true,
    PprofPath:   "/debug/pprof",  // Default
    PprofPort:   6060,             // Default
}

app, _ := pkg.New(pkg.FrameworkConfig{
    MonitoringConfig: config,
})
```

### Available Pprof Endpoints

Access profiling data at `http://localhost:6060/debug/pprof/`:

- `/debug/pprof/` - Index of available profiles
- `/debug/pprof/heap` - Memory allocation profile
- `/debug/pprof/goroutine` - Goroutine stack traces
- `/debug/pprof/threadcreate` - Thread creation profile
- `/debug/pprof/block` - Blocking profile
- `/debug/pprof/mutex` - Mutex contention profile
- `/debug/pprof/profile` - CPU profile (30-second sample)
- `/debug/pprof/trace` - Execution trace

### CPU Profiling

Capture a CPU profile:

```bash
# Capture 30-second CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with pprof
go tool pprof cpu.prof

# Interactive commands in pprof:
# top10 - Show top 10 functions by CPU time
# list <function> - Show source code for function
# web - Open visualization in browser
```

### Memory Profiling

Capture a memory profile:

```bash
# Capture heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze with pprof
go tool pprof heap.prof

# Show allocations
go tool pprof -alloc_space heap.prof

# Show in-use memory
go tool pprof -inuse_space heap.prof
```

### Goroutine Profiling

Analyze goroutine usage:

```bash
# Capture goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof

# Analyze
go tool pprof goroutine.prof

# Or view as text
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

### Blocking Profile

Identify blocking operations:

```bash
# Enable blocking profile in your application
import "runtime"

func main() {
    runtime.SetBlockProfileRate(1)
    
    // ... rest of application
}

# Capture blocking profile
curl http://localhost:6060/debug/pprof/block > block.prof

# Analyze
go tool pprof block.prof
```

### Mutex Contention Profile

Identify mutex contention:

```bash
# Enable mutex profile in your application
import "runtime"

func main() {
    runtime.SetMutexProfileFraction(1)
    
    // ... rest of application
}

# Capture mutex profile
curl http://localhost:6060/debug/pprof/mutex > mutex.prof

# Analyze
go tool pprof mutex.prof
```

## SNMP Monitoring

### Enabling SNMP

```go
config := pkg.MonitoringConfig{
    EnableSNMP:    true,
    SNMPPort:      161,      // Default
    SNMPCommunity: "public", // Default
}

app, _ := pkg.New(pkg.FrameworkConfig{
    MonitoringConfig: config,
})
```

### SNMP Data Structure

SNMP endpoint provides comprehensive monitoring data:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "system_info": {
    "num_cpu": 8,
    "num_goroutine": 156,
    "memory_alloc": 134217728,
    "memory_total_alloc": 2147483648,
    "memory_sys": 268435456,
    "num_gc": 42,
    "gc_pause_total_ns": 1234567890,
    "last_gc_time": "2025-01-15T10:29:55Z"
  },
  "workload_metrics": [
    {
      "timestamp": "2025-01-15T10:29:50Z",
      "tenant_id": "tenant-123",
      "request_id": "req-abc123",
      "duration": 45,
      "memory_usage": 1048576,
      "cpu_usage": 2.5,
      "path": "/api/users",
      "method": "GET",
      "status_code": 200
    }
  ],
  "aggregated_metrics": {
    "total_requests": 1523,
    "successful_requests": 1518,
    "failed_requests": 5,
    "avg_duration_ms": 42,
    "requests_per_second": 25.4
  }
}
```

### Querying SNMP Data

```bash
# Query SNMP endpoint
curl -H "X-SNMP-Community: public" \
    http://localhost:161/snmp
```

## Process Optimization

### Enabling Automatic Optimization

```go
config := pkg.MonitoringConfig{
    EnableOptimization:   true,
    OptimizationInterval: 5 * time.Minute, // Default
}

app, _ := pkg.New(pkg.FrameworkConfig{
    MonitoringConfig: config,
})
```

### Manual Optimization

Trigger optimization manually:

```go
monitoring := app.Monitoring()

// Perform immediate optimization
err := monitoring.OptimizeNow()
if err != nil {
    log.Printf("Optimization failed: %v", err)
}

// Get optimization statistics
stats := monitoring.GetOptimizationStats()
log.Printf("Last optimization: %v", stats.LastOptimization)
log.Printf("Memory freed: %d bytes", stats.MemoryFreed)
log.Printf("GC runs: %d", stats.GCRunsAfter-stats.GCRunsBefore)
```

### Optimization Statistics

```go
type OptimizationStats struct {
    LastOptimization  time.Time
    OptimizationCount int64
    GCRunsBefore      uint32
    GCRunsAfter       uint32
    MemoryBefore      uint64
    MemoryAfter       uint64
    MemoryFreed       uint64
}
```

## Health Checks

### Application Health Check

Implement a health check endpoint:

```go
router.GET("/health", func(ctx pkg.Context) error {
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
    }
    
    // Check database
    db := ctx.Database()
    if err := db.Ping(); err != nil {
        health["status"] = "unhealthy"
        health["database"] = "down"
        return ctx.JSON(503, health)
    }
    health["database"] = "up"
    
    // Check cache
    cache := ctx.Cache()
    if err := cache.Set("health_check", "ok", 1*time.Second); err != nil {
        health["cache"] = "degraded"
    } else {
        health["cache"] = "up"
    }
    
    // Add system metrics
    monitoring := ctx.Framework().Monitoring()
    if monitoring.IsRunning() {
        health["monitoring"] = "up"
    }
    
    return ctx.JSON(200, health)
})
```

### Detailed Health Check

```go
router.GET("/health/detailed", func(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Get system info
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "system": map[string]interface{}{
            "goroutines": runtime.NumGoroutine(),
            "memory_alloc": memStats.Alloc,
            "memory_sys": memStats.Sys,
            "num_gc": memStats.NumGC,
        },
    }
    
    // Check recent error rate
    from := time.Now().Add(-5 * time.Minute)
    to := time.Now()
    agg, err := metrics.GetAggregatedMetrics("", from, to)
    if err == nil {
        errorRate := float64(agg.FailedRequests) / float64(agg.TotalRequests) * 100
        health["error_rate"] = fmt.Sprintf("%.2f%%", errorRate)
        
        if errorRate > 5.0 {
            health["status"] = "degraded"
        }
    }
    
    return ctx.JSON(200, health)
})
```

### Kubernetes Liveness and Readiness Probes

```go
// Liveness probe - is the application running?
router.GET("/healthz", func(ctx pkg.Context) error {
    return ctx.String(200, "OK")
})

// Readiness probe - is the application ready to serve traffic?
router.GET("/readyz", func(ctx pkg.Context) error {
    // Check critical dependencies
    db := ctx.Database()
    if err := db.Ping(); err != nil {
        return ctx.String(503, "Database not ready")
    }
    
    return ctx.String(200, "Ready")
})
```

## Load Prediction

### Predicting Future Load

```go
metricsCollector := app.Metrics()

// Predict load for next hour based on last hour
prediction, err := metricsCollector.PredictLoad("tenant-123", 1*time.Hour)
if err != nil {
    log.Fatal(err)
}

log.Printf("Predicted requests: %d", prediction.PredictedRequests)
log.Printf("Predicted memory: %d bytes", prediction.PredictedMemory)
log.Printf("Predicted CPU: %.2f%%", prediction.PredictedCPU)
log.Printf("Confidence: %.2f%%", prediction.ConfidenceLevel*100)
```

### Using Predictions for Auto-Scaling

```go
// Check if scaling is needed
prediction, _ := metricsCollector.PredictLoad("", 15*time.Minute)

if prediction.PredictedCPU > 80.0 && prediction.ConfidenceLevel > 0.8 {
    log.Println("High CPU predicted, consider scaling up")
    // Trigger auto-scaling
}

if prediction.PredictedMemory > maxMemory && prediction.ConfidenceLevel > 0.8 {
    log.Println("High memory predicted, consider scaling up")
    // Trigger auto-scaling
}
```

## Monitoring Best Practices

### 1. Set Appropriate Metrics Intervals

```go
// Don't collect too frequently
config := pkg.MonitoringConfig{
    OptimizationInterval: 5 * time.Minute, // Not every second
}
```

### 2. Use Tags for Metric Dimensions

```go
// Good: Use tags for dimensions
metrics.IncrementCounter("api.requests", map[string]string{
    "endpoint": "/users",
    "method": "GET",
    "status": "200",
})

// Bad: Encode dimensions in metric name
metrics.IncrementCounter("api.requests.users.GET.200", nil)
```

### 3. Monitor What Matters

Focus on key metrics:
- **Request rate**: Requests per second
- **Error rate**: Percentage of failed requests
- **Response time**: P50, P95, P99 latencies
- **Resource usage**: CPU, memory, goroutines
- **Saturation**: Queue sizes, connection pools

### 4. Set Up Alerts

Configure alerts for critical metrics:

```yaml
# Prometheus alerting rules
groups:
  - name: rockstar_app
    rules:
      - alert: HighErrorRate
        expr: rate(http_errors[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          
      - alert: HighMemoryUsage
        expr: system_memory_usage > 1073741824  # 1GB
        for: 10m
        annotations:
          summary: "High memory usage"
          
      - alert: SlowResponses
        expr: avg(http_request_duration) > 1000  # 1 second
        for: 5m
        annotations:
          summary: "Slow response times"
```

### 5. Secure Monitoring Endpoints

```go
// Require authentication
config := pkg.MonitoringConfig{
    EnableMetrics: true,
    RequireAuth:   true,
    AuthToken:     os.Getenv("METRICS_TOKEN"),
}

// Or use firewall rules to restrict access
// Only allow access from monitoring systems
```

### 6. Regular Performance Reviews

Schedule regular reviews of metrics:
- Daily: Check error rates and response times
- Weekly: Review resource usage trends
- Monthly: Analyze capacity and scaling needs

## Troubleshooting

### Metrics Not Appearing

**Problem**: Metrics endpoint returns empty data

**Solutions**:
- Verify monitoring is enabled: `config.EnableMetrics = true`
- Check metrics port is accessible
- Ensure requests are being made to generate metrics
- Check logs for initialization errors

### High Memory Usage

**Problem**: Application memory usage keeps growing

**Solutions**:
```go
// Enable automatic optimization
config.EnableOptimization = true
config.OptimizationInterval = 5 * time.Minute

// Check for memory leaks with pprof
// curl http://localhost:6060/debug/pprof/heap > heap.prof
// go tool pprof -alloc_space heap.prof
```

### Pprof Not Accessible

**Problem**: Cannot access pprof endpoints

**Solutions**:
- Verify pprof is enabled: `config.EnablePprof = true`
- Check pprof port (default 6060)
- Ensure firewall allows access
- Check if port is already in use

### SNMP Connection Issues

**Problem**: Cannot connect to SNMP endpoint

**Solutions**:
- Verify SNMP is enabled: `config.EnableSNMP = true`
- Check SNMP port (default 161)
- Verify community string matches
- Check network connectivity

## See Also

- [Performance Guide](performance.md) - Performance optimization
- [Deployment Guide](deployment.md) - Production deployment
- [Configuration Guide](configuration.md) - Configuration options
- [API Reference: Monitoring](../api/monitoring.md) - Complete API documentation

