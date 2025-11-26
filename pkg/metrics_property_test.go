package pkg

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_MetricsCollectorCollectsMetricsInMemory tests Property 10:
// MetricsCollector collects metrics in memory
// **Feature: optional-database, Property 10: MetricsCollector collects metrics in memory**
// **Validates: Requirements 3.4**
func TestProperty_MetricsCollectorCollectsMetricsInMemory(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("MetricsCollector initializes with no-op database", prop.ForAll(
		func() bool {
			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)
			if mc == nil {
				t.Log("Metrics collector is nil")
				return false
			}

			// Verify it's using in-memory storage
			mcImpl, ok := mc.(*metricsCollectorImpl)
			if !ok {
				t.Log("Metrics collector is not the expected implementation")
				return false
			}

			// Check that in-memory storage is initialized
			if mcImpl.metricsStorage == nil {
				t.Log("Metrics storage is nil - should be in-memory")
				return false
			}

			return true
		},
	))

	properties.Property("Metrics recording works with in-memory storage", prop.ForAll(
		func(requestID string, tenantID string, path string) bool {
			// Skip empty values
			if requestID == "" || tenantID == "" || path == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Start metrics collection
			metrics := mc.Start(requestID)
			if metrics == nil {
				t.Log("Request metrics is nil")
				return false
			}

			// Set context
			metrics.SetContext(tenantID, "user-1", path, "GET")
			metrics.SetResponse(200, 1024, nil)

			// End metrics collection
			time.Sleep(1 * time.Millisecond)
			metrics.End()

			// Record metrics - should work without database
			err := mc.Record(metrics)
			if err != nil {
				t.Logf("Failed to record metrics: %v", err)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.RegexMatch(`/[a-z]+(/[a-z]+)?`),
	))

	properties.Property("Metrics retrieval works with in-memory storage", prop.ForAll(
		func(requestID string, tenantID string) bool {
			// Skip empty values
			if requestID == "" || tenantID == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Record a metric
			metrics := mc.Start(requestID)
			metrics.SetContext(tenantID, "user-1", "/api/test", "GET")
			metrics.SetResponse(200, 1024, nil)
			metrics.End()

			err := mc.Record(metrics)
			if err != nil {
				t.Logf("Failed to record metrics: %v", err)
				return false
			}

			// Retrieve metrics - should work without database
			now := time.Now()
			from := now.Add(-1 * time.Hour)
			to := now.Add(1 * time.Hour)

			retrieved, err := mc.GetMetrics(tenantID, from, to)
			if err != nil {
				t.Logf("Failed to retrieve metrics: %v", err)
				return false
			}

			// Should have at least one metric
			if len(retrieved) == 0 {
				t.Log("No metrics retrieved")
				return false
			}

			// Verify the metric matches
			found := false
			for _, m := range retrieved {
				if m.RequestID == requestID && m.TenantID == tenantID {
					found = true
					break
				}
			}

			if !found {
				t.Log("Recorded metric not found in retrieved metrics")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("Multiple metrics can be stored and retrieved", prop.ForAll(
		func(tenantID string, count uint8) bool {
			// Skip empty tenant ID and zero count
			if tenantID == "" || count == 0 {
				return true
			}

			// Limit count to reasonable number
			if count > 50 {
				count = 50
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Record multiple metrics
			for i := uint8(0); i < count; i++ {
				metrics := mc.Start("req-" + string(rune('0'+i)))
				metrics.SetContext(tenantID, "user-1", "/api/test", "GET")
				metrics.SetResponse(200, 1024, nil)
				metrics.End()

				err := mc.Record(metrics)
				if err != nil {
					t.Logf("Failed to record metric %d: %v", i, err)
					return false
				}
			}

			// Retrieve all metrics
			now := time.Now()
			from := now.Add(-1 * time.Hour)
			to := now.Add(1 * time.Hour)

			retrieved, err := mc.GetMetrics(tenantID, from, to)
			if err != nil {
				t.Logf("Failed to retrieve metrics: %v", err)
				return false
			}

			// Should have all recorded metrics
			if len(retrieved) != int(count) {
				t.Logf("Expected %d metrics, got %d", count, len(retrieved))
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.UInt8Range(1, 50),
	))

	properties.Property("Workload metrics recording works with in-memory storage", prop.ForAll(
		func(requestID string, tenantID string, duration int64) bool {
			// Skip empty values and negative duration
			if requestID == "" || tenantID == "" || duration < 0 {
				return true
			}

			// Limit duration to reasonable value
			if duration > 10000 {
				duration = 10000
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Create workload metrics
			wm := &WorkloadMetrics{
				Timestamp:   time.Now(),
				TenantID:    tenantID,
				RequestID:   requestID,
				Duration:    duration,
				MemoryUsage: 1024,
				CPUUsage:    50.0,
				Path:        "/api/test",
				Method:      "GET",
				StatusCode:  200,
			}

			// Record workload metrics - should work without database
			err := mc.RecordWorkloadMetrics(wm)
			if err != nil {
				t.Logf("Failed to record workload metrics: %v", err)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64Range(0, 10000),
	))

	properties.Property("Workload metrics retrieval works with in-memory storage", prop.ForAll(
		func(requestID string, tenantID string) bool {
			// Skip empty values
			if requestID == "" || tenantID == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Record workload metrics
			wm := &WorkloadMetrics{
				Timestamp:   time.Now(),
				TenantID:    tenantID,
				RequestID:   requestID,
				Duration:    100,
				MemoryUsage: 1024,
				CPUUsage:    50.0,
				Path:        "/api/test",
				Method:      "GET",
				StatusCode:  200,
			}

			err := mc.RecordWorkloadMetrics(wm)
			if err != nil {
				t.Logf("Failed to record workload metrics: %v", err)
				return false
			}

			// Retrieve workload metrics - should work without database
			now := time.Now()
			from := now.Add(-1 * time.Hour)
			to := now.Add(1 * time.Hour)

			retrieved, err := mc.GetWorkloadMetrics(tenantID, from, to)
			if err != nil {
				t.Logf("Failed to retrieve workload metrics: %v", err)
				return false
			}

			// Should have at least one metric
			if len(retrieved) == 0 {
				t.Log("No workload metrics retrieved")
				return false
			}

			// Verify the metric matches
			found := false
			for _, m := range retrieved {
				if m.RequestID == requestID && m.TenantID == tenantID {
					found = true
					break
				}
			}

			if !found {
				t.Log("Recorded workload metric not found in retrieved metrics")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.Property("Aggregated metrics work with in-memory storage", prop.ForAll(
		func(tenantID string, count uint8) bool {
			// Skip empty tenant ID and zero count
			if tenantID == "" || count == 0 {
				return true
			}

			// Limit count to reasonable number
			if count > 20 {
				count = 20
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Record multiple metrics
			for i := uint8(0); i < count; i++ {
				metrics := mc.Start("req-" + string(rune('0'+i)))
				metrics.SetContext(tenantID, "user-1", "/api/test", "GET")

				// Vary status codes
				statusCode := 200
				if i%3 == 0 {
					statusCode = 500
				}
				metrics.SetResponse(statusCode, 1024, nil)

				time.Sleep(1 * time.Millisecond)
				metrics.End()

				err := mc.Record(metrics)
				if err != nil {
					t.Logf("Failed to record metric %d: %v", i, err)
					return false
				}
			}

			// Get aggregated metrics - should work without database
			now := time.Now()
			from := now.Add(-1 * time.Hour)
			to := now.Add(1 * time.Hour)

			agg, err := mc.GetAggregatedMetrics(tenantID, from, to)
			if err != nil {
				t.Logf("Failed to get aggregated metrics: %v", err)
				return false
			}

			if agg == nil {
				t.Log("Aggregated metrics is nil")
				return false
			}

			// Verify total requests
			if agg.TotalRequests != int64(count) {
				t.Logf("Expected %d total requests, got %d", count, agg.TotalRequests)
				return false
			}

			// Verify successful and failed requests add up
			if agg.SuccessfulRequests+agg.FailedRequests != int64(count) {
				t.Logf("Successful (%d) + Failed (%d) != Total (%d)",
					agg.SuccessfulRequests, agg.FailedRequests, count)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.UInt8Range(1, 20),
	))

	properties.Property("Counter metrics work without database", prop.ForAll(
		func(name string, value int64) bool {
			// Skip empty name and negative values
			if name == "" || value < 0 {
				return true
			}

			// Limit value to reasonable range
			if value > 1000 {
				value = 1000
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Increment counter - should work without database
			err := mc.IncrementCounterBy(name, value, nil)
			if err != nil {
				t.Logf("Failed to increment counter: %v", err)
				return false
			}

			// Export metrics to verify
			exported, err := mc.Export()
			if err != nil {
				t.Logf("Failed to export metrics: %v", err)
				return false
			}

			counters, ok := exported["counters"].(map[string]int64)
			if !ok {
				t.Log("Counters not found in exported metrics")
				return false
			}

			if counters[name] != value {
				t.Logf("Expected counter value %d, got %d", value, counters[name])
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.Int64Range(0, 1000),
	))

	properties.Property("Gauge metrics work without database", prop.ForAll(
		func(name string, value float64) bool {
			// Skip empty name and invalid values
			if name == "" {
				return true
			}

			// Limit value to reasonable range
			if value < -1000 || value > 1000 {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create metrics collector
			mc := NewMetricsCollector(noopDB)

			// Set gauge - should work without database
			err := mc.SetGauge(name, value, nil)
			if err != nil {
				t.Logf("Failed to set gauge: %v", err)
				return false
			}

			// Export metrics to verify
			exported, err := mc.Export()
			if err != nil {
				t.Logf("Failed to export metrics: %v", err)
				return false
			}

			gauges, ok := exported["gauges"].(map[string]float64)
			if !ok {
				t.Log("Gauges not found in exported metrics")
				return false
			}

			if gauges[name] != value {
				t.Logf("Expected gauge value %f, got %f", value, gauges[name])
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.Float64Range(-1000, 1000),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
