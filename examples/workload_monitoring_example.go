package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// This example demonstrates workload monitoring and performance tracking in the Rockstar Web Framework.
// It shows how to track request metrics, monitor system resources, analyze performance patterns,
// and visualize workload data for capacity planning and optimization.

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Workload Monitoring Example\n")

	// Configure framework with database for metrics storage
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "workload_monitoring.db",
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Initialize database tables
	if err := app.Database().CreateTables(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	// Get metrics collector
	metrics := app.Metrics()

	fmt.Println("1. Recording Workload Metrics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	recordWorkloadMetrics(metrics)

	fmt.Println("\n2. Tracking System Resources")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trackSystemResources(metrics)

	fmt.Println("\n3. Using Counters and Gauges")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	useCountersAndGauges(metrics)

	fmt.Println("\n4. Timing Operations")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	timingOperations(metrics)

	fmt.Println("\n5. Retrieving Historical Metrics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	retrieveHistoricalMetrics(metrics)

	fmt.Println("\n6. Aggregated Metrics Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	analyzeAggregatedMetrics(metrics)

	fmt.Println("\n7. Load Prediction")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	predictLoad(metrics)

	fmt.Println("\n8. Exporting Metrics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	exportMetrics(metrics)

	// Set up HTTP endpoints for metrics visualization
	router := app.Router()

	// Metrics dashboard endpoint
	router.GET("/metrics/dashboard", func(ctx pkg.Context) error {
		from := time.Now().Add(-1 * time.Hour)
		to := time.Now()

		agg, err := metrics.GetAggregatedMetrics("", from, to)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"period":  "last_hour",
			"from":    from.Format(time.RFC3339),
			"to":      to.Format(time.RFC3339),
			"metrics": agg,
			"system":  getSystemMetrics(),
		})
	})

	// Performance report endpoint
	router.GET("/metrics/performance", func(ctx pkg.Context) error {
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()

		workloadMetrics, err := metrics.GetWorkloadMetrics("", from, to)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		report := generatePerformanceReport(workloadMetrics)
		return ctx.JSON(200, report)
	})

	fmt.Println("\n✅ Workload monitoring examples completed!")
	fmt.Println("\nStarting HTTP server on :8080...")
	fmt.Println("Try: curl http://localhost:8080/metrics/dashboard")
	fmt.Println("Try: curl http://localhost:8080/metrics/performance")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// recordWorkloadMetrics demonstrates recording detailed workload metrics
func recordWorkloadMetrics(metrics pkg.MetricsCollector) {
	// Simulate recording metrics for multiple requests
	for i := 0; i < 5; i++ {
		requestID := fmt.Sprintf("req-%d", i+1)

		// Start metrics collection
		reqMetrics := metrics.Start(requestID)

		// Set request context
		reqMetrics.SetContext(
			"tenant-1",
			fmt.Sprintf("user-%d", i+1),
			"/api/data",
			"GET",
		)

		// Simulate request processing
		time.Sleep(time.Duration(10+i*5) * time.Millisecond)

		// Set response information
		reqMetrics.SetResponse(200, int64(1024*(i+1)), nil)

		// Finalize metrics
		reqMetrics.End()

		// Record to database
		if err := metrics.Record(reqMetrics); err != nil {
			log.Printf("Error recording metrics: %v", err)
		}

		fmt.Printf("  ✓ Recorded metrics for %s (duration: %v)\n",
			requestID, reqMetrics.Duration)
	}
}

// trackSystemResources demonstrates tracking system-level resources
func trackSystemResources(metrics pkg.MetricsCollector) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Record memory metrics
	memUsage := int64(m.Alloc)
	if err := metrics.RecordMemoryUsage(memUsage); err != nil {
		log.Printf("Error recording memory: %v", err)
	}

	fmt.Printf("  Memory Allocated: %d bytes (%.2f MB)\n",
		memUsage, float64(memUsage)/(1024*1024))
	fmt.Printf("  Total Allocated: %d bytes (%.2f MB)\n",
		m.TotalAlloc, float64(m.TotalAlloc)/(1024*1024))
	fmt.Printf("  System Memory: %d bytes (%.2f MB)\n",
		m.Sys, float64(m.Sys)/(1024*1024))
	fmt.Printf("  GC Runs: %d\n", m.NumGC)
	fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())

	// Record CPU usage (simplified calculation)
	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10.0
	if cpuUsage > 100.0 {
		cpuUsage = 100.0
	}

	if err := metrics.RecordCPUUsage(cpuUsage); err != nil {
		log.Printf("Error recording CPU: %v", err)
	}

	fmt.Printf("  Estimated CPU Usage: %.2f%%\n", cpuUsage)
}

// useCountersAndGauges demonstrates using counters and gauges for metrics
func useCountersAndGauges(metrics pkg.MetricsCollector) {
	// Increment request counter
	tags := map[string]string{
		"endpoint": "/api/users",
		"method":   "GET",
		"status":   "200",
	}

	if err := metrics.IncrementCounter("http.requests", tags); err != nil {
		log.Printf("Error incrementing counter: %v", err)
	}
	fmt.Println("  ✓ Incremented http.requests counter")

	// Increment by specific value
	if err := metrics.IncrementCounterBy("http.bytes_sent", 2048, tags); err != nil {
		log.Printf("Error incrementing counter: %v", err)
	}
	fmt.Println("  ✓ Incremented http.bytes_sent by 2048")

	// Set gauge for active connections
	if err := metrics.SetGauge("connections.active", 42.0, nil); err != nil {
		log.Printf("Error setting gauge: %v", err)
	}
	fmt.Println("  ✓ Set connections.active gauge to 42")

	// Increment gauge
	if err := metrics.IncrementGauge("connections.active", 5.0, nil); err != nil {
		log.Printf("Error incrementing gauge: %v", err)
	}
	fmt.Println("  ✓ Incremented connections.active by 5 (now 47)")

	// Decrement gauge
	if err := metrics.DecrementGauge("connections.active", 3.0, nil); err != nil {
		log.Printf("Error decrementing gauge: %v", err)
	}
	fmt.Println("  ✓ Decremented connections.active by 3 (now 44)")

	// Record histogram value
	if err := metrics.RecordHistogram("response.size", 1024.0, tags); err != nil {
		log.Printf("Error recording histogram: %v", err)
	}
	fmt.Println("  ✓ Recorded response.size histogram value")
}

// timingOperations demonstrates timing operations with metrics
func timingOperations(metrics pkg.MetricsCollector) {
	// Example 1: Manual timing
	start := time.Now()
	time.Sleep(25 * time.Millisecond)
	duration := time.Since(start)

	tags := map[string]string{"operation": "database_query"}
	if err := metrics.RecordTiming("operation.duration", duration, tags); err != nil {
		log.Printf("Error recording timing: %v", err)
	}
	fmt.Printf("  ✓ Manual timing recorded: %v\n", duration)

	// Example 2: Using timer
	timer := metrics.StartTimer("api.request.duration", map[string]string{
		"endpoint": "/api/data",
	})

	// Simulate API request
	time.Sleep(50 * time.Millisecond)

	elapsed := timer.Stop()
	fmt.Printf("  ✓ Timer recorded: %v\n", elapsed)

	// Example 3: Multiple timers
	dbTimer := metrics.StartTimer("database.query", map[string]string{"query": "SELECT"})
	time.Sleep(15 * time.Millisecond)
	dbElapsed := dbTimer.Stop()

	cacheTimer := metrics.StartTimer("cache.lookup", map[string]string{"key": "user:123"})
	time.Sleep(5 * time.Millisecond)
	cacheElapsed := cacheTimer.Stop()

	fmt.Printf("  ✓ Database query: %v\n", dbElapsed)
	fmt.Printf("  ✓ Cache lookup: %v\n", cacheElapsed)
}

// retrieveHistoricalMetrics demonstrates retrieving historical workload metrics
func retrieveHistoricalMetrics(metrics pkg.MetricsCollector) {
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()

	workloadMetrics, err := metrics.GetWorkloadMetrics("tenant-1", from, to)
	if err != nil {
		log.Printf("Error retrieving metrics: %v", err)
		return
	}

	fmt.Printf("  Retrieved %d metrics for tenant-1\n", len(workloadMetrics))

	if len(workloadMetrics) > 0 {
		fmt.Println("\n  Sample metrics:")
		for i, m := range workloadMetrics {
			if i >= 3 {
				break
			}
			fmt.Printf("    %d. %s %s - %dms (status: %d)\n",
				i+1, m.Method, m.Path, m.Duration, m.StatusCode)
		}
	}
}

// analyzeAggregatedMetrics demonstrates analyzing aggregated metrics
func analyzeAggregatedMetrics(metrics pkg.MetricsCollector) {
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()

	agg, err := metrics.GetAggregatedMetrics("tenant-1", from, to)
	if err != nil {
		log.Printf("Error getting aggregated metrics: %v", err)
		return
	}

	fmt.Printf("  Period: %s to %s\n", from.Format("15:04:05"), to.Format("15:04:05"))
	fmt.Printf("  Total Requests: %d\n", agg.TotalRequests)
	fmt.Printf("  Successful: %d (%.1f%%)\n",
		agg.SuccessfulRequests,
		float64(agg.SuccessfulRequests)/float64(agg.TotalRequests)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n",
		agg.FailedRequests,
		float64(agg.FailedRequests)/float64(agg.TotalRequests)*100)
	fmt.Printf("  Avg Duration: %v\n", agg.AvgDuration)
	fmt.Printf("  Min Duration: %v\n", agg.MinDuration)
	fmt.Printf("  Max Duration: %v\n", agg.MaxDuration)
	fmt.Printf("  Avg Memory: %.2f MB\n", float64(agg.AvgMemoryUsage)/(1024*1024))
	fmt.Printf("  Avg CPU: %.2f%%\n", agg.AvgCPUUsage)
	fmt.Printf("  Requests/sec: %.2f\n", agg.RequestsPerSecond)
}

// predictLoad demonstrates load prediction based on historical data
func predictLoad(metrics pkg.MetricsCollector) {
	duration := 1 * time.Hour

	prediction, err := metrics.PredictLoad("tenant-1", duration)
	if err != nil {
		log.Printf("Error predicting load: %v", err)
		return
	}

	fmt.Printf("  Prediction for next %v:\n", duration)
	fmt.Printf("  Expected Requests: %d\n", prediction.PredictedRequests)
	fmt.Printf("  Expected Memory: %.2f MB\n", float64(prediction.PredictedMemory)/(1024*1024))
	fmt.Printf("  Expected CPU: %.2f%%\n", prediction.PredictedCPU)
	fmt.Printf("  Confidence: %.0f%%\n", prediction.ConfidenceLevel*100)
	fmt.Printf("  Based on %d data points\n", prediction.BasedOnDataPoints)
}

// exportMetrics demonstrates exporting metrics in different formats
func exportMetrics(metrics pkg.MetricsCollector) {
	// Export as map
	data, err := metrics.Export()
	if err != nil {
		log.Printf("Error exporting metrics: %v", err)
		return
	}

	fmt.Println("  JSON Format:")
	if counters, ok := data["counters"].(map[string]int64); ok {
		fmt.Printf("    Counters: %d entries\n", len(counters))
		count := 0
		for name, value := range counters {
			if count >= 3 {
				break
			}
			fmt.Printf("      %s: %d\n", name, value)
			count++
		}
	}

	if gauges, ok := data["gauges"].(map[string]float64); ok {
		fmt.Printf("    Gauges: %d entries\n", len(gauges))
		count := 0
		for name, value := range gauges {
			if count >= 3 {
				break
			}
			fmt.Printf("      %s: %.2f\n", name, value)
			count++
		}
	}

	// Export in Prometheus format
	prometheusData, err := metrics.ExportPrometheus()
	if err != nil {
		log.Printf("Error exporting Prometheus: %v", err)
		return
	}

	fmt.Println("\n  Prometheus Format (sample):")
	lines := string(prometheusData)
	if len(lines) > 200 {
		fmt.Printf("    %s...\n", lines[:200])
	} else {
		fmt.Printf("    %s\n", lines)
	}
}

// getSystemMetrics returns current system metrics
func getSystemMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc_bytes":       m.Alloc,
			"total_alloc_bytes": m.TotalAlloc,
			"sys_bytes":         m.Sys,
			"num_gc":            m.NumGC,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"num_cpu":    runtime.NumCPU(),
		},
	}
}

// generatePerformanceReport generates a performance report from workload metrics
func generatePerformanceReport(metrics []*pkg.WorkloadMetrics) map[string]interface{} {
	if len(metrics) == 0 {
		return map[string]interface{}{
			"message": "No metrics available",
		}
	}

	var totalDuration int64
	var totalMemory int64
	var totalCPU float64
	statusCodes := make(map[int]int)
	paths := make(map[string]int)

	for _, m := range metrics {
		totalDuration += m.Duration
		totalMemory += m.MemoryUsage
		totalCPU += m.CPUUsage
		statusCodes[m.StatusCode]++
		paths[m.Path]++
	}

	count := int64(len(metrics))

	return map[string]interface{}{
		"total_requests":   count,
		"avg_duration_ms":  totalDuration / count,
		"avg_memory_bytes": totalMemory / count,
		"avg_cpu_percent":  totalCPU / float64(count),
		"status_codes":     statusCodes,
		"top_paths":        paths,
	}
}
