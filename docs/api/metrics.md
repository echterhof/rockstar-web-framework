---
title: "Metrics API"
description: "Metrics collector interface for application monitoring and performance tracking"
category: "api"
tags: ["api", "metrics", "monitoring", "performance", "observability"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "monitoring.md"
  - "../guides/monitoring.md"
---

# Metrics API

## Overview

The `MetricsCollector` interface provides comprehensive metrics collection and analysis for the Rockstar Web Framework. It tracks request performance, resource usage, and custom application metrics, enabling detailed monitoring and performance optimization.

The metrics collector supports counters, gauges, histograms, timers, and workload metrics with support for tags and aggregation.

**Primary Use Cases:**
- Request performance monitoring
- Resource usage tracking (CPU, memory)
- Custom business metrics
- Load prediction and capacity planning
- Performance analysis and optimization

## Interface Definition

```go
type MetricsCollector interface {
    // Request metrics
    Start(requestID string) *RequestMetrics
    Record(metrics *RequestMetrics) error
    RecordRequest(ctx Context, duration time.Duration, statusCode int) error
    RecordError(ctx Context, err error) error

    // Workload metrics
    RecordWorkloadMetrics(metrics *WorkloadMetrics) error
    GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)
    GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error)
    PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error)

    // Counter metrics
    IncrementCounter(name string, tags map[string]string) error
    IncrementCounterBy(name string, value int64, tags map[string]string) error

    // Gauge metrics
    SetGauge(name string, value float64, tags map[string]string) error
    IncrementGauge(name string, value float64, tags map[string]string) error
    DecrementGauge(name string, value float64, tags map[string]string) error

    // Histogram and timing metrics
    RecordHistogram(name string, value float64, tags map[string]string) error
    RecordTiming(name string, duration time.Duration, tags map[string]string) error
    StartTimer(name string, tags map[string]string) Timer

    // System metrics
    RecordMemoryUsage(usage int64) error
    RecordCPUUsage(usage float64) error
    RecordCustomMetric(name string, value interface{}, tags map[string]string) error

    // Export
    Export() (map[string]interface{}, error)
    ExportPrometheus() ([]byte, error)
}
```

## Request Metrics

### Start

```go
func Start(requestID string) *RequestMetrics
```

**Description**: Begins metrics collection for a request. Captures initial memory and CPU state.

**Parameters**:
- `requestID` (string): Unique request identifier

**Returns**:
- `*RequestMetrics`: Request metrics tracker

**Example**:
```go
func middleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    metrics := ctx.Metrics()
    
    // Start tracking request
    requestMetrics := metrics.Start(ctx.Request().ID)
    
    // Set context information
    requestMetrics.SetContext(
        ctx.Tenant().ID,
        ctx.User().ID,
        ctx.Request().RequestURI,
        ctx.Request().Method,
    )
    
    // Execute handler
    err := next(ctx)
    
    // Record response
    statusCode := 200
    if err != nil {
        statusCode = 500
    }
    requestMetrics.SetResponse(statusCode, 0, err)
    
    // Finalize and record
    requestMetrics.End()
    metrics.Record(requestMetrics)
    
    return err
}
```

### Record

```go
func Record(metrics *RequestMetrics) error
```

**Description**: Saves collected request metrics to storage.

**Parameters**:
- `metrics` (*RequestMetrics): Request metrics to record

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    requestMetrics := metrics.Start(ctx.Request().ID)
    
    // Process request...
    
    requestMetrics.End()
    if err := metrics.Record(requestMetrics); err != nil {
        // Log error but don't fail request
        ctx.Logger().Error("Failed to record metrics", "error", err)
    }
    
    return ctx.JSON(200, result)
}
```

### RecordRequest

```go
func RecordRequest(ctx Context, duration time.Duration, statusCode int) error
```

**Description**: Records request metrics with duration and status code. Automatically extracts tenant and method information from context.

**Parameters**:
- `ctx` (Context): Request context
- `duration` (time.Duration): Request duration
- `statusCode` (int): HTTP status code

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func middleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    err := next(ctx)
    
    duration := time.Since(start)
    statusCode := 200
    if err != nil {
        statusCode = 500
    }
    
    // Record request metrics
    ctx.Metrics().RecordRequest(ctx, duration, statusCode)
    
    return err
}
```

### RecordError

```go
func RecordError(ctx Context, err error) error
```

**Description**: Records error metrics for a request.

**Parameters**:
- `ctx` (Context): Request context
- `err` (error): Error that occurred

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    result, err := processRequest(ctx)
    if err != nil {
        // Record error metric
        ctx.Metrics().RecordError(ctx, err)
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    }
    
    return ctx.JSON(200, result)
}
```

## Workload Metrics

### RecordWorkloadMetrics

```go
func RecordWorkloadMetrics(metrics *WorkloadMetrics) error
```

**Description**: Records detailed workload metrics including resource usage and performance data.

**Parameters**:
- `metrics` (*WorkloadMetrics): Workload metrics to record

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    workloadMetrics := &pkg.WorkloadMetrics{
        Timestamp:    time.Now(),
        TenantID:     ctx.Tenant().ID,
        UserID:       ctx.User().ID,
        RequestID:    ctx.Request().ID,
        Duration:     100, // milliseconds
        ContextSize:  1024,
        MemoryUsage:  2048,
        CPUUsage:     15.5,
        Path:         ctx.Request().RequestURI,
        Method:       ctx.Request().Method,
        StatusCode:   200,
        ResponseSize: 4096,
    }
    
    ctx.Metrics().RecordWorkloadMetrics(workloadMetrics)
    
    return ctx.JSON(200, result)
}
```

### GetWorkloadMetrics

```go
func GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)
```

**Description**: Retrieves workload metrics for a tenant within a time range.

**Parameters**:
- `tenantID` (string): Tenant identifier (empty for all tenants)
- `from` (time.Time): Start time
- `to` (time.Time): End time

**Returns**:
- `[]*WorkloadMetrics`: List of workload metrics
- `error`: Error if retrieval fails

**Example**:
```go
func metricsHandler(ctx pkg.Context) error {
    tenantID := ctx.Param("tenant_id")
    from := time.Now().Add(-24 * time.Hour)
    to := time.Now()
    
    metrics, err := ctx.Metrics().GetWorkloadMetrics(tenantID, from, to)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to get metrics"})
    }
    
    return ctx.JSON(200, metrics)
}
```

### GetAggregatedMetrics

```go
func GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error)
```

**Description**: Returns aggregated metrics with statistics like averages, totals, and request counts.

**Parameters**:
- `tenantID` (string): Tenant identifier
- `from` (time.Time): Start time
- `to` (time.Time): End time

**Returns**:
- `*AggregatedMetrics`: Aggregated metrics
- `error`: Error if aggregation fails

**Example**:
```go
func statsHandler(ctx pkg.Context) error {
    tenantID := ctx.Tenant().ID
    from := time.Now().Add(-1 * time.Hour)
    to := time.Now()
    
    agg, err := ctx.Metrics().GetAggregatedMetrics(tenantID, from, to)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to get stats"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "total_requests":       agg.TotalRequests,
        "successful_requests":  agg.SuccessfulRequests,
        "failed_requests":      agg.FailedRequests,
        "avg_duration_ms":      agg.AvgDuration.Milliseconds(),
        "requests_per_second":  agg.RequestsPerSecond,
        "avg_memory_usage":     agg.AvgMemoryUsage,
        "avg_cpu_usage":        agg.AvgCPUUsage,
    })
}
```

### PredictLoad

```go
func PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error)
```

**Description**: Predicts future load based on historical data using simple statistical analysis.

**Parameters**:
- `tenantID` (string): Tenant identifier
- `duration` (time.Duration): Duration to look back for historical data

**Returns**:
- `*LoadPrediction`: Load prediction with confidence level
- `error`: Error if prediction fails

**Example**:
```go
func predictionHandler(ctx pkg.Context) error {
    tenantID := ctx.Tenant().ID
    
    // Predict based on last 7 days
    prediction, err := ctx.Metrics().PredictLoad(tenantID, 7*24*time.Hour)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to predict load"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "predicted_requests": prediction.PredictedRequests,
        "predicted_memory":   prediction.PredictedMemory,
        "predicted_cpu":      prediction.PredictedCPU,
        "confidence":         prediction.ConfidenceLevel,
        "data_points":        prediction.BasedOnDataPoints,
    })
}
```

## Counter Metrics

### IncrementCounter

```go
func IncrementCounter(name string, tags map[string]string) error
```

**Description**: Increments a counter metric by 1.

**Parameters**:
- `name` (string): Counter name
- `tags` (map[string]string): Optional tags for grouping

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Increment page view counter
    metrics.IncrementCounter("page.views", map[string]string{
        "page": "/home",
    })
    
    // Increment API call counter
    metrics.IncrementCounter("api.calls", map[string]string{
        "endpoint": "/api/users",
        "method":   "GET",
    })
    
    return ctx.JSON(200, result)
}
```

### IncrementCounterBy

```go
func IncrementCounterBy(name string, value int64, tags map[string]string) error
```

**Description**: Increments a counter metric by a specific value.

**Parameters**:
- `name` (string): Counter name
- `value` (int64): Amount to increment by
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Increment by batch size
    batchSize := int64(100)
    metrics.IncrementCounterBy("items.processed", batchSize, map[string]string{
        "type": "batch",
    })
    
    return ctx.JSON(200, result)
}
```

## Gauge Metrics

### SetGauge

```go
func SetGauge(name string, value float64, tags map[string]string) error
```

**Description**: Sets a gauge metric to a specific value.

**Parameters**:
- `name` (string): Gauge name
- `value` (float64): Value to set
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Set current queue size
    queueSize := float64(len(queue))
    metrics.SetGauge("queue.size", queueSize, map[string]string{
        "queue": "processing",
    })
    
    // Set temperature reading
    metrics.SetGauge("sensor.temperature", 23.5, map[string]string{
        "sensor": "room1",
    })
    
    return ctx.JSON(200, result)
}
```

### IncrementGauge

```go
func IncrementGauge(name string, value float64, tags map[string]string) error
```

**Description**: Increments a gauge metric by a value.

**Parameters**:
- `name` (string): Gauge name
- `value` (float64): Amount to increment by
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Increment active connections
    metrics.IncrementGauge("connections.active", 1.0, nil)
    
    return ctx.JSON(200, result)
}
```

### DecrementGauge

```go
func DecrementGauge(name string, value float64, tags map[string]string) error
```

**Description**: Decrements a gauge metric by a value.

**Parameters**:
- `name` (string): Gauge name
- `value` (float64): Amount to decrement by
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Decrement active connections
    metrics.DecrementGauge("connections.active", 1.0, nil)
    
    return ctx.JSON(200, result)
}
```

## Histogram and Timing Metrics

### RecordHistogram

```go
func RecordHistogram(name string, value float64, tags map[string]string) error
```

**Description**: Records a value in a histogram metric for distribution analysis.

**Parameters**:
- `name` (string): Histogram name
- `value` (float64): Value to record
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Record response size
    responseSize := float64(len(responseData))
    metrics.RecordHistogram("response.size", responseSize, map[string]string{
        "endpoint": "/api/users",
    })
    
    return ctx.JSON(200, result)
}
```

### RecordTiming

```go
func RecordTiming(name string, duration time.Duration, tags map[string]string) error
```

**Description**: Records a timing/duration metric.

**Parameters**:
- `name` (string): Timing metric name
- `duration` (time.Duration): Duration to record
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    start := time.Now()
    
    // Perform database query
    result, err := ctx.DB().Query("SELECT * FROM users")
    
    // Record query duration
    duration := time.Since(start)
    metrics.RecordTiming("db.query.duration", duration, map[string]string{
        "query": "select_users",
    })
    
    return ctx.JSON(200, result)
}
```

### StartTimer

```go
func StartTimer(name string, tags map[string]string) Timer
```

**Description**: Starts a timer that automatically records duration when stopped.

**Parameters**:
- `name` (string): Timer name
- `tags` (map[string]string): Optional tags

**Returns**:
- `Timer`: Timer instance

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Start timer
    timer := metrics.StartTimer("operation.duration", map[string]string{
        "operation": "process_data",
    })
    
    // Perform operation
    processData()
    
    // Stop timer and record duration
    duration := timer.Stop()
    
    return ctx.JSON(200, map[string]interface{}{
        "duration_ms": duration.Milliseconds(),
    })
}
```

## System Metrics

### RecordMemoryUsage

```go
func RecordMemoryUsage(usage int64) error
```

**Description**: Records current memory usage in bytes.

**Parameters**:
- `usage` (int64): Memory usage in bytes

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func monitoringHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Get current memory usage
    memUsage := pkg.GetCurrentMemoryUsage()
    metrics.RecordMemoryUsage(memUsage)
    
    return ctx.JSON(200, map[string]interface{}{
        "memory_bytes": memUsage,
    })
}
```

### RecordCPUUsage

```go
func RecordCPUUsage(usage float64) error
```

**Description**: Records current CPU usage percentage.

**Parameters**:
- `usage` (float64): CPU usage percentage (0-100)

**Returns**:
- `error`: Error if recording fails

**Example**:
```go
func monitoringHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Get current CPU usage
    cpuUsage := pkg.GetCurrentCPUUsage()
    metrics.RecordCPUUsage(cpuUsage)
    
    return ctx.JSON(200, map[string]interface{}{
        "cpu_percent": cpuUsage,
    })
}
```

### RecordCustomMetric

```go
func RecordCustomMetric(name string, value interface{}, tags map[string]string) error
```

**Description**: Records a custom metric with automatic type detection.

**Parameters**:
- `name` (string): Metric name
- `value` (interface{}): Metric value (int, int64, float32, float64)
- `tags` (map[string]string): Optional tags

**Returns**:
- `error`: Error if type is unsupported or recording fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Record various custom metrics
    metrics.RecordCustomMetric("business.revenue", 1250.50, map[string]string{
        "currency": "USD",
    })
    
    metrics.RecordCustomMetric("inventory.count", 500, map[string]string{
        "product": "widget",
    })
    
    return ctx.JSON(200, result)
}
```

## Export

### Export

```go
func Export() (map[string]interface{}, error)
```

**Description**: Exports all metrics as a map structure.

**Returns**:
- `map[string]interface{}`: Metrics data
- `error`: Error if export fails

**Example**:
```go
func metricsExportHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    data, err := metrics.Export()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to export metrics"})
    }
    
    return ctx.JSON(200, data)
}
```

### ExportPrometheus

```go
func ExportPrometheus() ([]byte, error)
```

**Description**: Exports metrics in Prometheus text format.

**Returns**:
- `[]byte`: Prometheus-formatted metrics
- `error`: Error if export fails

**Example**:
```go
func prometheusHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    data, err := metrics.ExportPrometheus()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to export"})
    }
    
    ctx.SetHeader("Content-Type", "text/plain; version=0.0.4")
    return ctx.String(200, string(data))
}
```

## Complete Example

Here's a complete example demonstrating comprehensive metrics collection:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "time"
)

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    
    // Add metrics middleware
    app.Use(metricsMiddleware)
    
    router := app.Router()
    router.GET("/", homeHandler)
    router.GET("/metrics", metricsHandler)
    router.GET("/metrics/prometheus", prometheusHandler)
    
    app.Listen(":8080")
}

func metricsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    metrics := ctx.Metrics()
    start := time.Now()
    
    // Start request tracking
    requestMetrics := metrics.Start(ctx.Request().ID)
    requestMetrics.SetContext(
        getTenantID(ctx),
        getUserID(ctx),
        ctx.Request().RequestURI,
        ctx.Request().Method,
    )
    
    // Increment request counter
    metrics.IncrementCounter("http.requests.total", map[string]string{
        "method": ctx.Request().Method,
        "path":   ctx.Request().RequestURI,
    })
    
    // Execute handler
    err := next(ctx)
    
    // Calculate duration
    duration := time.Since(start)
    
    // Determine status code
    statusCode := 200
    if err != nil {
        statusCode = 500
        metrics.RecordError(ctx, err)
    }
    
    // Record request metrics
    requestMetrics.SetResponse(statusCode, 0, err)
    requestMetrics.End()
    metrics.Record(requestMetrics)
    
    // Record timing
    metrics.RecordTiming("http.request.duration", duration, map[string]string{
        "method": ctx.Request().Method,
        "status": fmt.Sprintf("%d", statusCode),
    })
    
    // Record status code counter
    metrics.IncrementCounter("http.responses.total", map[string]string{
        "status": fmt.Sprintf("%d", statusCode),
    })
    
    return err
}

func homeHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Record custom business metric
    metrics.RecordCustomMetric("page.views", 1, map[string]string{
        "page": "home",
    })
    
    // Time database operation
    timer := metrics.StartTimer("db.query", map[string]string{
        "operation": "fetch_users",
    })
    
    users, err := ctx.DB().Query("SELECT * FROM users LIMIT 10")
    timer.Stop()
    
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Database error"})
    }
    
    return ctx.JSON(200, users)
}

func metricsHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Get aggregated metrics for last hour
    tenantID := getTenantID(ctx)
    from := time.Now().Add(-1 * time.Hour)
    to := time.Now()
    
    agg, err := metrics.GetAggregatedMetrics(tenantID, from, to)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to get metrics"})
    }
    
    // Get load prediction
    prediction, _ := metrics.PredictLoad(tenantID, 24*time.Hour)
    
    return ctx.JSON(200, map[string]interface{}{
        "current": agg,
        "prediction": prediction,
    })
}

func prometheusHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    data, err := metrics.ExportPrometheus()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Export failed"})
    }
    
    ctx.SetHeader("Content-Type", "text/plain; version=0.0.4")
    return ctx.String(200, string(data))
}

func getTenantID(ctx pkg.Context) string {
    if tenant := ctx.Tenant(); tenant != nil {
        return tenant.ID
    }
    return ""
}

func getUserID(ctx pkg.Context) string {
    if user := ctx.User(); user != nil {
        return user.ID
    }
    return ""
}
```

## Best Practices

### Metric Naming

Use hierarchical, descriptive names:

```go
// Good
"http.requests.total"
"db.query.duration"
"cache.hits"
"business.revenue"

// Bad
"requests"
"query"
"hits"
"rev"
```

### Tag Usage

Use tags for dimensions, not values:

```go
// Good
metrics.IncrementCounter("http.requests", map[string]string{
    "method": "GET",
    "status": "200",
})

// Bad - don't use high-cardinality values as tags
metrics.IncrementCounter("http.requests", map[string]string{
    "user_id": "12345",  // Too many unique values
    "timestamp": "...",   // Constantly changing
})
```

### Performance

Minimize metrics overhead:

```go
// Good - batch operations
timer := metrics.StartTimer("operation", tags)
// ... do work ...
timer.Stop()

// Bad - multiple separate calls
start := time.Now()
// ... do work ...
metrics.RecordTiming("operation", time.Since(start), tags)
metrics.IncrementCounter("operation.count", tags)
metrics.SetGauge("operation.last", float64(time.Now().Unix()), tags)
```

## See Also

- [Framework API](framework.md)
- [Context API](context.md)
- [Monitoring API](monitoring.md)
- [Monitoring Guide](../guides/monitoring.md)
- [Performance Guide](../guides/performance.md)
