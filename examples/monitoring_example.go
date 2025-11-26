//go:build ignore

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// This example demonstrates monitoring and profiling features in the Rockstar Web Framework.
// It shows how to collect metrics, enable pprof profiling, create health check endpoints,
// and implement custom metrics for application monitoring.

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Monitoring Example\n")

	// Configure the framework with minimal settings - defaults work great!
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1: true,
			EnableHTTP2: true,
			// ReadTimeout defaults to 30s
			// WriteTimeout defaults to 30s
		},
		// Enable database for metrics storage
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "monitoring_example.db",
			// Host defaults to "localhost"
			// MaxOpenConns defaults to 25
			// MaxIdleConns defaults to 5
			// ConnMaxLifetime defaults to 5 minutes
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get metrics collector
	metrics := app.Metrics()

	// Configure monitoring - only specify what you want to enable
	monitoringConfig := pkg.MonitoringConfig{
		// Metrics endpoint - exposes application metrics
		EnableMetrics: true,
		MetricsPath:   "/metrics",
		// MetricsPort defaults to 9090

		// Pprof profiling - enables Go profiling endpoints
		EnablePprof: true,
		PprofPath:   "/debug/pprof",
		// PprofPort defaults to 6060

		// Process optimization - periodic garbage collection
		EnableOptimization:   true,
		OptimizationInterval: 30 * time.Second, // Override default of 5 minutes for demo
		// OptimizationInterval defaults to 5 minutes

		// Security - disable auth for demo purposes
		RequireAuth: false,
		// SNMPPort defaults to 161
		// SNMPCommunity defaults to "public"
	}

	// Create logger
	logger := pkg.NewLogger(nil)

	// Create and start monitoring manager
	monitor := pkg.NewMonitoringManager(monitoringConfig, metrics, app.Database(), logger)
	if err := monitor.Start(); err != nil {
		log.Fatalf("Failed to start monitoring: %v", err)
	}
	defer monitor.Stop()

	fmt.Println("✅ Monitoring started successfully\n")
	fmt.Println("Available Endpoints:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Metrics:     http://localhost:9090/metrics")
	fmt.Println("🔍 Pprof:       http://localhost:6060/debug/pprof")
	fmt.Println("❤️  Health:      http://localhost:8080/health")
	fmt.Println("📈 App Metrics: http://localhost:8080/app/metrics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Get router
	router := app.Router()

	// 1. Health Check Endpoint
	// Standard health check that returns application status
	router.GET("/health", func(ctx pkg.Context) error {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
			"version":   "1.0.0",
		}

		// Check database connectivity
		if err := app.Database().Ping(); err != nil {
			health["status"] = "unhealthy"
			health["database"] = "disconnected"
			return ctx.JSON(503, health)
		}
		health["database"] = "connected"

		return ctx.JSON(200, health)
	})

	// 2. Detailed Health Check with Component Status
	router.GET("/health/detailed", func(ctx pkg.Context) error {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"components": map[string]interface{}{
				"database": checkDatabase(app.Database()),
				"cache":    checkCache(app.Cache()),
				"monitor":  checkMonitor(monitor),
			},
			"metrics": map[string]interface{}{
				"goroutines": getGoroutineCount(),
				"memory":     getMemoryUsage(),
			},
		}

		return ctx.JSON(200, health)
	})

	// 3. Application Metrics Endpoint
	// Custom endpoint for application-specific metrics
	router.GET("/app/metrics", func(ctx pkg.Context) error {
		// Export all collected metrics
		metricsData, err := metrics.Export()
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to export metrics",
			})
		}

		return ctx.JSON(200, metricsData)
	})

	// 4. Custom Metrics Example
	// Endpoint that demonstrates recording custom metrics
	router.GET("/api/data", func(ctx pkg.Context) error {
		// Start timing this request
		timer := metrics.StartTimer("api.data.duration", map[string]string{
			"endpoint": "/api/data",
		})
		defer timer.Stop()

		// Increment request counter
		metrics.IncrementCounter("api.data.requests", map[string]string{
			"method": "GET",
		})

		// Simulate some work
		time.Sleep(50 * time.Millisecond)

		// Record custom business metric
		metrics.SetGauge("api.data.items", 42, nil)

		return ctx.JSON(200, map[string]interface{}{
			"message": "Data retrieved successfully",
			"items":   42,
		})
	})

	// 5. Error Tracking Example
	// Endpoint that demonstrates error metric recording
	router.GET("/api/error", func(ctx pkg.Context) error {
		// Increment error counter
		metrics.IncrementCounter("api.errors", map[string]string{
			"endpoint": "/api/error",
			"type":     "simulated",
		})

		return ctx.JSON(500, map[string]interface{}{
			"error": "Simulated error for monitoring demo",
		})
	})

	// 6. Workload Metrics Example
	// Middleware to automatically collect request metrics
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		// Generate request ID
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

		// Start request metrics
		requestMetrics := metrics.Start(requestID)
		requestMetrics.SetContext(
			"default-tenant",
			"demo-user",
			ctx.Request().RequestURI,
			ctx.Request().Method,
		)

		// Execute handler
		err := next(ctx)

		// Record response metrics
		statusCode := 200
		if err != nil {
			statusCode = 500
		}
		requestMetrics.SetResponse(statusCode, 0, err)
		requestMetrics.End()

		// Save to database
		metrics.Record(requestMetrics)

		return err
	})

	// 7. Optimization Status Endpoint
	router.GET("/monitor/optimization", func(ctx pkg.Context) error {
		stats := monitor.GetOptimizationStats()
		return ctx.JSON(200, stats)
	})

	// 8. Trigger Manual Optimization
	router.POST("/monitor/optimize", func(ctx pkg.Context) error {
		if err := monitor.OptimizeNow(); err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		stats := monitor.GetOptimizationStats()
		return ctx.JSON(200, map[string]interface{}{
			"message": "Optimization completed",
			"stats":   stats,
		})
	})

	// Start the application server
	fmt.Println("Starting application server on :8080...")
	fmt.Println("\nTry these commands:")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println("  curl http://localhost:8080/health/detailed")
	fmt.Println("  curl http://localhost:8080/app/metrics")
	fmt.Println("  curl http://localhost:9090/metrics")
	fmt.Println("  curl http://localhost:6060/debug/pprof/")
	fmt.Println("\nPprof profiling commands:")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
	fmt.Println()

	// Simulate some background activity
	go simulateActivity(metrics)

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

var startTime = time.Now()

// simulateActivity generates background metrics for demonstration
func simulateActivity(metrics pkg.MetricsCollector) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := 0
	for range ticker.C {
		counter++

		// Simulate various metrics
		metrics.SetGauge("app.active_users", float64(10+counter%5), nil)
		metrics.SetGauge("app.queue_size", float64(counter%20), nil)
		metrics.IncrementCounter("app.background_tasks", map[string]string{
			"type": "cleanup",
		})

		// Simulate memory and CPU metrics
		metrics.RecordMemoryUsage(pkg.GetCurrentMemoryUsage())
		metrics.RecordCPUUsage(pkg.GetCurrentCPUUsage())
	}
}

// checkDatabase checks database health
func checkDatabase(db pkg.DatabaseManager) map[string]interface{} {
	if err := db.Ping(); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	stats := db.Stats()
	return map[string]interface{}{
		"status":           "healthy",
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
	}
}

// checkCache checks cache health
func checkCache(cache pkg.CacheManager) map[string]interface{} {
	// Try to set and get a test value
	testKey := "health_check"
	testValue := "ok"

	if err := cache.Set(testKey, testValue, 1*time.Second); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	retrieved, err := cache.Get(testKey)
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	if retrieved != testValue {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "cache value mismatch",
		}
	}

	cache.Delete(testKey)

	return map[string]interface{}{
		"status": "healthy",
	}
}

// checkMonitor checks monitoring manager health
func checkMonitor(monitor pkg.MonitoringManager) map[string]interface{} {
	if !monitor.IsRunning() {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "monitoring manager not running",
		}
	}

	return map[string]interface{}{
		"status": "healthy",
		"config": monitor.GetConfig(),
	}
}

// getGoroutineCount returns the current number of goroutines
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

// getMemoryUsage returns current memory usage information
func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_bytes":       m.Alloc,
		"total_alloc_bytes": m.TotalAlloc,
		"sys_bytes":         m.Sys,
		"num_gc":            m.NumGC,
		"goroutines":        runtime.NumGoroutine(),
	}
}
