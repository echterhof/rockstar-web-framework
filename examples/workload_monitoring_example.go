package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create a database manager (using mock for example)
	db := createMockDatabase()

	// Create metrics collector
	collector := pkg.NewMetricsCollector(db)

	// Example 1: Record workload metrics
	fmt.Println("=== Example 1: Record Workload Metrics ===")
	recordWorkloadMetrics(collector)

	// Example 2: Get workload metrics
	fmt.Println("\n=== Example 2: Get Workload Metrics ===")
	getWorkloadMetrics(collector)

	// Example 3: Monitor system resources
	fmt.Println("\n=== Example 3: Monitor System Resources ===")
	monitorSystemResources(collector)

	// Example 4: Use counters and gauges
	fmt.Println("\n=== Example 4: Use Counters and Gauges ===")
	useCountersAndGauges(collector)

	// Example 5: Export metrics
	fmt.Println("\n=== Example 5: Export Metrics ===")
	exportMetrics(collector)
}

// recordWorkloadMetrics demonstrates recording workload metrics
func recordWorkloadMetrics(collector pkg.MetricsCollector) {
	// Create a workload metric
	metric := &pkg.WorkloadMetrics{
		Timestamp:    time.Now(),
		TenantID:     "tenant-1",
		UserID:       "user-123",
		RequestID:    "req-12345",
		Duration:     150,     // milliseconds
		ContextSize:  2048,    // bytes
		MemoryUsage:  4096000, // bytes
		CPUUsage:     45.5,    // percentage
		Path:         "/api/users",
		Method:       "GET",
		StatusCode:   200,
		ResponseSize: 1024, // bytes
	}

	// Record the metric
	if err := collector.RecordWorkloadMetrics(metric); err != nil {
		log.Printf("Error recording workload metrics: %v", err)
		return
	}

	fmt.Printf("Workload metric recorded:\n")
	fmt.Printf("  Request ID: %s\n", metric.RequestID)
	fmt.Printf("  Duration: %d ms\n", metric.Duration)
	fmt.Printf("  Memory Usage: %d bytes (%.2f MB)\n", metric.MemoryUsage, float64(metric.MemoryUsage)/(1024*1024))
	fmt.Printf("  CPU Usage: %.2f%%\n", metric.CPUUsage)
	fmt.Printf("  Status Code: %d\n", metric.StatusCode)
}

// getWorkloadMetrics demonstrates retrieving workload metrics
func getWorkloadMetrics(collector pkg.MetricsCollector) {
	// Get metrics for the last hour
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()

	metrics, err := collector.GetWorkloadMetrics("tenant-1", from, to)
	if err != nil {
		log.Printf("Error getting workload metrics: %v", err)
		return
	}

	fmt.Printf("Retrieved %d workload metrics for tenant-1\n", len(metrics))

	if len(metrics) > 0 {
		fmt.Printf("\nSample metric:\n")
		m := metrics[0]
		fmt.Printf("  Request ID: %s\n", m.RequestID)
		fmt.Printf("  Path: %s\n", m.Path)
		fmt.Printf("  Method: %s\n", m.Method)
		fmt.Printf("  Duration: %d ms\n", m.Duration)
		fmt.Printf("  Status Code: %d\n", m.StatusCode)
	}
}

// monitorSystemResources demonstrates monitoring system-level resources
func monitorSystemResources(collector pkg.MetricsCollector) {
	// Get current memory usage
	memUsage := pkg.GetCurrentMemoryUsage()
	fmt.Printf("Current Memory Usage: %d bytes (%.2f MB)\n", memUsage, float64(memUsage)/(1024*1024))

	// Record memory usage metric
	if err := collector.RecordMemoryUsage(memUsage); err != nil {
		log.Printf("Error recording memory usage: %v", err)
	}

	// Get current CPU usage (approximate)
	cpuUsage := pkg.GetCurrentCPUUsage()
	fmt.Printf("Current CPU Usage: %.2f%%\n", cpuUsage)

	// Record CPU usage metric
	if err := collector.RecordCPUUsage(cpuUsage); err != nil {
		log.Printf("Error recording CPU usage: %v", err)
	}
}

// useCountersAndGauges demonstrates using counters and gauges
func useCountersAndGauges(collector pkg.MetricsCollector) {
	// Increment a counter
	tags := map[string]string{
		"endpoint": "/api/users",
		"method":   "GET",
	}

	if err := collector.IncrementCounter("http.requests", tags); err != nil {
		log.Printf("Error incrementing counter: %v", err)
	}
	fmt.Println("Incremented http.requests counter")

	// Set a gauge
	if err := collector.SetGauge("active.connections", 42.0, nil); err != nil {
		log.Printf("Error setting gauge: %v", err)
	}
	fmt.Println("Set active.connections gauge to 42")

	// Record a timing
	duration := 150 * time.Millisecond
	if err := collector.RecordTiming("request.duration", duration, tags); err != nil {
		log.Printf("Error recording timing: %v", err)
	}
	fmt.Printf("Recorded request duration: %v\n", duration)

	// Use a timer
	timer := collector.StartTimer("operation.duration", map[string]string{"operation": "database_query"})

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	elapsed := timer.Stop()
	fmt.Printf("Operation completed in: %v\n", elapsed)
}

// exportMetrics demonstrates exporting metrics
func exportMetrics(collector pkg.MetricsCollector) {
	// Export all metrics as a map
	metrics, err := collector.Export()
	if err != nil {
		log.Printf("Error exporting metrics: %v", err)
		return
	}

	fmt.Printf("Exported Metrics:\n")
	if counters, ok := metrics["counters"].(map[string]int64); ok {
		fmt.Printf("  Counters: %d\n", len(counters))
		for name, value := range counters {
			fmt.Printf("    %s: %d\n", name, value)
		}
	}
	if gauges, ok := metrics["gauges"].(map[string]float64); ok {
		fmt.Printf("  Gauges: %d\n", len(gauges))
		for name, value := range gauges {
			fmt.Printf("    %s: %.2f\n", name, value)
		}
	}

	// Export in Prometheus format
	fmt.Println("\nPrometheus Format:")
	prometheusData, err := collector.ExportPrometheus()
	if err != nil {
		log.Printf("Error exporting Prometheus metrics: %v", err)
		return
	}
	fmt.Println(string(prometheusData))
}

// createMockDatabase creates a mock database for the example
func createMockDatabase() pkg.DatabaseManager {
	// In a real application, you would create a real database connection
	// For this example, we'll return nil which will cause metrics to be skipped
	// when recording to the database
	return nil
}
