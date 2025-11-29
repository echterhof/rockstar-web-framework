---
title: "Monitoring API"
description: "Monitoring manager interface for health checks, profiling, and system monitoring"
category: "api"
tags: ["api", "monitoring", "health", "profiling", "observability"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "metrics.md"
  - "../guides/monitoring.md"
---

# Monitoring API

## Overview

The `MonitoringManager` interface provides comprehensive monitoring capabilities for the Rockstar Web Framework, including metrics endpoints, pprof profiling, SNMP support, and automatic process optimization.

The monitoring manager enables observability through multiple channels and helps maintain optimal application performance.

**Primary Use Cases:**
- Exposing metrics for Prometheus/Grafana
- Performance profiling with pprof
- SNMP monitoring integration
- Automatic memory optimization
- Health check endpoints
- System resource monitoring

## Interface Definition

```go
type MonitoringManager interface {
    // Lifecycle
    Start() error
    Stop() error
    IsRunning() bool

    // Metrics endpoint
    EnableMetricsEndpoint(path string) error
    DisableMetricsEndpoint() error
    GetMetricsHandler() http.HandlerFunc

    // Pprof support
    EnablePprof(path string) error
    DisablePprof() error
    GetPprofHandlers() map[string]http.HandlerFunc

    // SNMP support
    EnableSNMP(port int, community string) error
    DisableSNMP() error
    GetSNMPData() (*SNMPData, error)

    // Process optimization
    EnableOptimization() error
    DisableOptimization() error
    OptimizeNow() error
    GetOptimizationStats() *OptimizationStats

    // Configuration
    SetConfig(config MonitoringConfig) error
    GetConfig() MonitoringConfig
}
```

## Configuration

### MonitoringConfig

```go
type MonitoringConfig struct {
    // Metrics endpoint configuration
    EnableMetrics bool   // Default: false
    MetricsPath   string // Default: ""
    MetricsPort   int    // Default: 9090

    // Pprof configuration
    EnablePprof bool   // Default: false
    PprofPath   string // Default: ""
    PprofPort   int    // Default: 6060

    // SNMP configuration
    EnableSNMP    bool   // Default: false
    SNMPPort      int    // Default: 161
    SNMPCommunity string // Default: "public"

    // Process optimization
    EnableOptimization   bool          // Default: false
    OptimizationInterval time.Duration // Default: 5 minutes

    // Security
    RequireAuth bool   // Default: false
    AuthToken   string // Default: ""
}
```

**Example**:
```go
config := pkg.MonitoringConfig{
    EnableMetrics:        true,
    MetricsPath:          "/metrics",
    MetricsPort:          9090,
    EnablePprof:          true,
    PprofPath:            "/debug/pprof",
    PprofPort:            6060,
    EnableOptimization:   true,
    OptimizationInterval: 5 * time.Minute,
    RequireAuth:          true,
    AuthToken:            "your-secret-token",
}
```

## Lifecycle Management

### Start

```go
func Start() error
```

**Description**: Starts the monitoring manager and all enabled monitoring features.

**Returns**:
- `error`: Error if startup fails

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics: true,
            EnablePprof:   true,
        },
    })

    monitoring := app.Monitoring()
    
    // Start monitoring
    if err := monitoring.Start(); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }

    app.Listen(":8080")
}
```

### Stop

```go
func Stop() error
```

**Description**: Stops the monitoring manager and all monitoring services.

**Returns**:
- `error`: Error if shutdown fails

**Example**:
```go
func shutdown(app *pkg.Framework) {
    monitoring := app.Monitoring()
    
    if err := monitoring.Stop(); err != nil {
        log.Printf("Error stopping monitoring: %v", err)
    }
}
```

### IsRunning

```go
func IsRunning() bool
```

**Description**: Returns whether the monitoring manager is currently running.

**Returns**:
- `bool`: true if running, false otherwise

**Example**:
```go
func statusHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    return ctx.JSON(200, map[string]interface{}{
        "monitoring_active": monitoring.IsRunning(),
    })
}
```

## Metrics Endpoint

### EnableMetricsEndpoint

```go
func EnableMetricsEndpoint(path string) error
```

**Description**: Enables the HTTP metrics endpoint for Prometheus scraping.

**Parameters**:
- `path` (string): HTTP path for metrics endpoint (e.g., "/metrics")

**Returns**:
- `error`: Error if enabling fails

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    monitoring := app.Monitoring()
    
    // Enable metrics endpoint
    if err := monitoring.EnableMetricsEndpoint("/metrics"); err != nil {
        log.Fatalf("Failed to enable metrics: %v", err)
    }
    
    // Metrics available at http://localhost:9090/metrics
    
    app.Listen(":8080")
}
```

### DisableMetricsEndpoint

```go
func DisableMetricsEndpoint() error
```

**Description**: Disables the metrics endpoint.

**Returns**:
- `error`: Error if disabling fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    if err := monitoring.DisableMetricsEndpoint(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to disable metrics"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Metrics disabled"})
}
```

### GetMetricsHandler

```go
func GetMetricsHandler() http.HandlerFunc
```

**Description**: Returns the HTTP handler for the metrics endpoint.

**Returns**:
- `http.HandlerFunc`: Metrics handler

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    monitoring := app.Monitoring()
    
    // Get metrics handler and register on main router
    router := app.Router()
    metricsHandler := monitoring.GetMetricsHandler()
    
    router.GET("/custom/metrics", func(ctx pkg.Context) error {
        metricsHandler(ctx.Response(), ctx.Request())
        return nil
    })
    
    app.Listen(":8080")
}
```

## Pprof Support

### EnablePprof

```go
func EnablePprof(path string) error
```

**Description**: Enables pprof profiling endpoints for performance analysis.

**Parameters**:
- `path` (string): Base path for pprof endpoints (e.g., "/debug/pprof")

**Returns**:
- `error`: Error if enabling fails

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    monitoring := app.Monitoring()
    
    // Enable pprof
    if err := monitoring.EnablePprof("/debug/pprof"); err != nil {
        log.Fatalf("Failed to enable pprof: %v", err)
    }
    
    // Pprof available at:
    // http://localhost:6060/debug/pprof/
    // http://localhost:6060/debug/pprof/heap
    // http://localhost:6060/debug/pprof/goroutine
    // http://localhost:6060/debug/pprof/profile
    
    app.Listen(":8080")
}
```

### DisablePprof

```go
func DisablePprof() error
```

**Description**: Disables pprof profiling endpoints.

**Returns**:
- `error`: Error if disabling fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    if err := monitoring.DisablePprof(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to disable pprof"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Pprof disabled"})
}
```

### GetPprofHandlers

```go
func GetPprofHandlers() map[string]http.HandlerFunc
```

**Description**: Returns all pprof HTTP handlers.

**Returns**:
- `map[string]http.HandlerFunc`: Map of path to handler

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    monitoring := app.Monitoring()
    
    // Get pprof handlers
    pprofHandlers := monitoring.GetPprofHandlers()
    
    // Register on custom paths
    router := app.Router()
    for path, handler := range pprofHandlers {
        customPath := "/profiling" + path
        router.GET(customPath, func(ctx pkg.Context) error {
            handler(ctx.Response(), ctx.Request())
            return nil
        })
    }
    
    app.Listen(":8080")
}
```

## SNMP Support

### EnableSNMP

```go
func EnableSNMP(port int, community string) error
```

**Description**: Enables SNMP monitoring support.

**Parameters**:
- `port` (int): SNMP port (typically 161)
- `community` (string): SNMP community string for authentication

**Returns**:
- `error`: Error if enabling fails

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    monitoring := app.Monitoring()
    
    // Enable SNMP
    if err := monitoring.EnableSNMP(161, "public"); err != nil {
        log.Fatalf("Failed to enable SNMP: %v", err)
    }
    
    app.Listen(":8080")
}
```

### DisableSNMP

```go
func DisableSNMP() error
```

**Description**: Disables SNMP monitoring.

**Returns**:
- `error`: Error if disabling fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    if err := monitoring.DisableSNMP(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to disable SNMP"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "SNMP disabled"})
}
```

### GetSNMPData

```go
func GetSNMPData() (*SNMPData, error)
```

**Description**: Returns current SNMP monitoring data including system info and metrics.

**Returns**:
- `*SNMPData`: SNMP data structure
- `error`: Error if retrieval fails

**Example**:
```go
func snmpHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    data, err := monitoring.GetSNMPData()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to get SNMP data"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "timestamp":   data.Timestamp,
        "system_info": data.SystemInfo,
        "metrics":     data.WorkloadMetrics,
    })
}
```

## Process Optimization

### EnableOptimization

```go
func EnableOptimization() error
```

**Description**: Enables automatic process optimization including periodic garbage collection.

**Returns**:
- `error`: Error if enabling fails

**Example**:
```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{
        MonitoringConfig: pkg.MonitoringConfig{
            EnableOptimization:   true,
            OptimizationInterval: 5 * time.Minute,
        },
    })
    
    monitoring := app.Monitoring()
    monitoring.Start()
    
    app.Listen(":8080")
}
```

### DisableOptimization

```go
func DisableOptimization() error
```

**Description**: Disables automatic process optimization.

**Returns**:
- `error`: Error if disabling fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    if err := monitoring.DisableOptimization(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to disable optimization"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Optimization disabled"})
}
```

### OptimizeNow

```go
func OptimizeNow() error
```

**Description**: Performs immediate process optimization (forces garbage collection).

**Returns**:
- `error`: Error if optimization fails

**Example**:
```go
func optimizeHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    if err := monitoring.OptimizeNow(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Optimization failed"})
    }
    
    stats := monitoring.GetOptimizationStats()
    
    return ctx.JSON(200, map[string]interface{}{
        "memory_freed": stats.MemoryFreed,
        "gc_runs":      stats.GCRunsAfter - stats.GCRunsBefore,
    })
}
```

### GetOptimizationStats

```go
func GetOptimizationStats() *OptimizationStats
```

**Description**: Returns statistics about process optimizations.

**Returns**:
- `*OptimizationStats`: Optimization statistics

**Example**:
```go
func statsHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    stats := monitoring.GetOptimizationStats()
    
    return ctx.JSON(200, map[string]interface{}{
        "last_optimization":  stats.LastOptimization,
        "optimization_count": stats.OptimizationCount,
        "memory_freed":       stats.MemoryFreed,
        "memory_before":      stats.MemoryBefore,
        "memory_after":       stats.MemoryAfter,
    })
}
```

## Configuration Management

### SetConfig

```go
func SetConfig(config MonitoringConfig) error
```

**Description**: Updates the monitoring configuration.

**Parameters**:
- `config` (MonitoringConfig): New configuration

**Returns**:
- `error`: Error if update fails

**Example**:
```go
func updateConfigHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    newConfig := pkg.MonitoringConfig{
        EnableMetrics:      true,
        MetricsPort:        9091,
        EnableOptimization: true,
    }
    
    if err := monitoring.SetConfig(newConfig); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to update config"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Config updated"})
}
```

### GetConfig

```go
func GetConfig() MonitoringConfig
```

**Description**: Returns the current monitoring configuration.

**Returns**:
- `MonitoringConfig`: Current configuration

**Example**:
```go
func configHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    config := monitoring.GetConfig()
    
    return ctx.JSON(200, config)
}
```

## Complete Example

Here's a complete example demonstrating comprehensive monitoring setup:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
    "time"
)

func main() {
    // Configure monitoring
    config := pkg.FrameworkConfig{
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics:        true,
            MetricsPath:          "/metrics",
            MetricsPort:          9090,
            EnablePprof:          true,
            PprofPath:            "/debug/pprof",
            PprofPort:            6060,
            EnableOptimization:   true,
            OptimizationInterval: 5 * time.Minute,
            RequireAuth:          true,
            AuthToken:            "secret-token",
        },
    }

    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }

    monitoring := app.Monitoring()

    // Start monitoring
    if err := monitoring.Start(); err != nil {
        log.Fatalf("Failed to start monitoring: %v", err)
    }

    // Register monitoring endpoints
    router := app.Router()
    router.GET("/health", healthHandler)
    router.GET("/monitoring/status", monitoringStatusHandler)
    router.POST("/monitoring/optimize", optimizeHandler)
    router.GET("/monitoring/stats", statsHandler)

    // Graceful shutdown
    app.RegisterShutdownHook(func(ctx context.Context) error {
        return monitoring.Stop()
    })

    log.Println("Server starting...")
    log.Println("Metrics: http://localhost:9090/metrics")
    log.Println("Pprof: http://localhost:6060/debug/pprof/")
    
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

func healthHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    return ctx.JSON(200, map[string]interface{}{
        "status":             "healthy",
        "monitoring_active":  monitoring.IsRunning(),
        "timestamp":          time.Now(),
    })
}

func monitoringStatusHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    config := monitoring.GetConfig()
    
    return ctx.JSON(200, map[string]interface{}{
        "running":             monitoring.IsRunning(),
        "metrics_enabled":     config.EnableMetrics,
        "pprof_enabled":       config.EnablePprof,
        "optimization_enabled": config.EnableOptimization,
    })
}

func optimizeHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    
    // Perform optimization
    if err := monitoring.OptimizeNow(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Optimization failed"})
    }
    
    // Get stats
    stats := monitoring.GetOptimizationStats()
    
    return ctx.JSON(200, map[string]interface{}{
        "message":      "Optimization completed",
        "memory_freed": stats.MemoryFreed,
        "gc_runs":      stats.GCRunsAfter - stats.GCRunsBefore,
    })
}

func statsHandler(ctx pkg.Context) error {
    monitoring := ctx.Framework().Monitoring()
    stats := monitoring.GetOptimizationStats()
    
    // Get SNMP data if enabled
    var snmpData *pkg.SNMPData
    if monitoring.GetConfig().EnableSNMP {
        snmpData, _ = monitoring.GetSNMPData()
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "optimization": stats,
        "snmp":         snmpData,
    })
}
```

## Best Practices

### Security

1. **Protect monitoring endpoints**: Use authentication for production
2. **Separate ports**: Run monitoring on different ports from main application
3. **Firewall rules**: Restrict access to monitoring ports
4. **Disable in production**: Consider disabling pprof in production

### Performance

1. **Optimize intervals**: Balance monitoring frequency with overhead
2. **Use separate servers**: Run metrics/pprof on dedicated ports
3. **Monitor the monitor**: Track monitoring overhead itself

### Integration

1. **Prometheus**: Use standard metrics endpoint for scraping
2. **Grafana**: Create dashboards from exported metrics
3. **Alerting**: Set up alerts based on metrics thresholds

## See Also

- [Framework API](framework.md)
- [Metrics API](metrics.md)
- [Monitoring Guide](../guides/monitoring.md)
- [Performance Guide](../guides/performance.md)
- [Deployment Guide](../guides/deployment.md)
