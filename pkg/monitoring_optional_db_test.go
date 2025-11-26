package pkg

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: optional-database, Property 9: MonitoringManager collects metrics in memory**
// **Validates: Requirements 3.3**
func TestProperty_MonitoringManagerCollectsMetricsInMemory(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("MonitoringManager collects metrics in memory without database",
		prop.ForAll(
			func(tenantID, userID, requestID, path, method string, duration, contextSize, memoryUsage int64, cpuUsage float64, statusCode int) bool {
				// Skip invalid inputs
				if duration < 0 || contextSize < 0 || memoryUsage < 0 || cpuUsage < 0 || statusCode < 100 || statusCode > 599 {
					return true
				}

				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Create metrics collector with no-op database
				metricsCollector := NewMetricsCollector(noopDB)

				// Create logger
				logger := NewLogger(nil)

				// Create monitoring manager with no-op database
				monitoringConfig := MonitoringConfig{}
				monitoringManager := NewMonitoringManager(monitoringConfig, metricsCollector, noopDB, logger)

				// Verify monitoring manager was created
				if monitoringManager == nil {
					t.Log("MonitoringManager is nil")
					return false
				}

				// Verify it's using in-memory storage by checking the implementation
				monitoringImpl, ok := monitoringManager.(*monitoringManagerImpl)
				if !ok {
					t.Log("MonitoringManager is not the expected implementation type")
					return false
				}

				// Verify in-memory storage was initialized
				if monitoringImpl.metricsStorage == nil {
					t.Log("In-memory metrics storage was not initialized")
					return false
				}

				// Create a workload metric
				metric := &WorkloadMetrics{
					Timestamp:    time.Now(),
					TenantID:     tenantID,
					UserID:       userID,
					RequestID:    requestID,
					Duration:     duration,
					ContextSize:  contextSize,
					MemoryUsage:  memoryUsage,
					CPUUsage:     cpuUsage,
					Path:         path,
					Method:       method,
					StatusCode:   statusCode,
					ResponseSize: 0,
					ErrorMessage: "",
				}

				// Save metric to in-memory storage
				err := monitoringImpl.metricsStorage.Save(metric)
				if err != nil {
					t.Logf("Failed to save metric to in-memory storage: %v", err)
					return false
				}

				// Query metrics from in-memory storage
				from := time.Now().Add(-1 * time.Hour)
				to := time.Now().Add(1 * time.Hour)
				metrics, err := monitoringImpl.metricsStorage.Query(tenantID, from, to)
				if err != nil {
					t.Logf("Failed to query metrics from in-memory storage: %v", err)
					return false
				}

				// Verify we got at least one metric back
				if len(metrics) == 0 {
					t.Log("No metrics returned from in-memory storage")
					return false
				}

				// Verify the metric we saved is in the results
				found := false
				for _, m := range metrics {
					if m.RequestID == requestID && m.TenantID == tenantID {
						found = true
						// Verify the data matches
						if m.Duration != duration || m.StatusCode != statusCode {
							t.Log("Metric data doesn't match what was saved")
							return false
						}
						break
					}
				}

				if !found {
					t.Log("Saved metric not found in query results")
					return false
				}

				return true
			},
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
			gen.AlphaString(),
			gen.OneConstOf("GET", "POST", "PUT", "DELETE", "PATCH"),
			gen.Int64Range(0, 10000),
			gen.Int64Range(0, 1000000),
			gen.Int64Range(0, 1000000),
			gen.Float64Range(0, 100),
			gen.IntRange(200, 599),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that MonitoringManager works without attempting database operations
func TestProperty_MonitoringManagerNoDatabaseOperations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("MonitoringManager doesn't attempt database operations with no-op database",
		prop.ForAll(
			func(enableMetrics, enablePprof, enableSNMP, enableOptimization bool) bool {
				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Create metrics collector with no-op database
				metricsCollector := NewMetricsCollector(noopDB)

				// Create logger
				logger := NewLogger(nil)

				// Create monitoring config with various features enabled
				monitoringConfig := MonitoringConfig{
					EnableMetrics:      enableMetrics,
					EnablePprof:        enablePprof,
					EnableSNMP:         enableSNMP,
					EnableOptimization: enableOptimization,
				}

				// Create monitoring manager with no-op database
				monitoringManager := NewMonitoringManager(monitoringConfig, metricsCollector, noopDB, logger)

				// Verify monitoring manager was created
				if monitoringManager == nil {
					t.Log("MonitoringManager is nil")
					return false
				}

				// Verify it's using in-memory storage
				monitoringImpl, ok := monitoringManager.(*monitoringManagerImpl)
				if !ok {
					t.Log("MonitoringManager is not the expected implementation type")
					return false
				}

				if monitoringImpl.metricsStorage == nil {
					t.Log("In-memory metrics storage was not initialized")
					return false
				}

				// Verify database is no-op
				if !isNoopDatabase(monitoringImpl.db) {
					t.Log("Database is not no-op")
					return false
				}

				// Try to get SNMP data (should work without database)
				snmpData, err := monitoringManager.GetSNMPData()
				if err != nil {
					t.Logf("GetSNMPData failed: %v", err)
					return false
				}

				// Verify SNMP data was returned
				if snmpData == nil {
					t.Log("SNMP data is nil")
					return false
				}

				// Verify system info is present
				if snmpData.SystemInfo == nil {
					t.Log("System info is nil")
					return false
				}

				return true
			},
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that metrics persist in memory across multiple operations
func TestProperty_MonitoringManagerMetricsPersistInMemory(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Metrics persist in memory across multiple save operations",
		prop.ForAll(
			func(numMetrics int) bool {
				// Skip invalid inputs
				if numMetrics < 1 || numMetrics > 100 {
					return true
				}

				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Create metrics collector with no-op database
				metricsCollector := NewMetricsCollector(noopDB)

				// Create logger
				logger := NewLogger(nil)

				// Create monitoring manager with no-op database
				monitoringConfig := MonitoringConfig{}
				monitoringManager := NewMonitoringManager(monitoringConfig, metricsCollector, noopDB, logger)

				// Get implementation
				monitoringImpl, ok := monitoringManager.(*monitoringManagerImpl)
				if !ok {
					t.Log("MonitoringManager is not the expected implementation type")
					return false
				}

				// Save multiple metrics
				for i := 0; i < numMetrics; i++ {
					metric := &WorkloadMetrics{
						Timestamp:   time.Now(),
						TenantID:    "test-tenant",
						RequestID:   string(rune('A' + i)),
						Duration:    int64(i * 100),
						StatusCode:  200,
						MemoryUsage: int64(i * 1000),
					}

					err := monitoringImpl.metricsStorage.Save(metric)
					if err != nil {
						t.Logf("Failed to save metric %d: %v", i, err)
						return false
					}
				}

				// Query all metrics
				from := time.Now().Add(-1 * time.Hour)
				to := time.Now().Add(1 * time.Hour)
				metrics, err := monitoringImpl.metricsStorage.Query("test-tenant", from, to)
				if err != nil {
					t.Logf("Failed to query metrics: %v", err)
					return false
				}

				// Verify we got all metrics back
				if len(metrics) != numMetrics {
					t.Logf("Expected %d metrics, got %d", numMetrics, len(metrics))
					return false
				}

				return true
			},
			gen.IntRange(1, 100),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that MonitoringManager filters metrics by tenant correctly
func TestProperty_MonitoringManagerFiltersByTenant(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("MonitoringManager filters metrics by tenant ID correctly",
		prop.ForAll(
			func(tenant1, tenant2 string) bool {
				// Skip if tenants are the same or empty
				if tenant1 == "" || tenant2 == "" || tenant1 == tenant2 {
					return true
				}

				// Create no-op database manager
				noopDB := NewNoopDatabaseManager()

				// Create metrics collector with no-op database
				metricsCollector := NewMetricsCollector(noopDB)

				// Create logger
				logger := NewLogger(nil)

				// Create monitoring manager with no-op database
				monitoringConfig := MonitoringConfig{}
				monitoringManager := NewMonitoringManager(monitoringConfig, metricsCollector, noopDB, logger)

				// Get implementation
				monitoringImpl, ok := monitoringManager.(*monitoringManagerImpl)
				if !ok {
					t.Log("MonitoringManager is not the expected implementation type")
					return false
				}

				// Save metrics for tenant1
				for i := 0; i < 5; i++ {
					metric := &WorkloadMetrics{
						Timestamp:  time.Now(),
						TenantID:   tenant1,
						RequestID:  string(rune('A' + i)),
						Duration:   int64(i * 100),
						StatusCode: 200,
					}
					_ = monitoringImpl.metricsStorage.Save(metric)
				}

				// Save metrics for tenant2
				for i := 0; i < 3; i++ {
					metric := &WorkloadMetrics{
						Timestamp:  time.Now(),
						TenantID:   tenant2,
						RequestID:  string(rune('X' + i)),
						Duration:   int64(i * 100),
						StatusCode: 200,
					}
					_ = monitoringImpl.metricsStorage.Save(metric)
				}

				// Query metrics for tenant1
				from := time.Now().Add(-1 * time.Hour)
				to := time.Now().Add(1 * time.Hour)
				metrics1, err := monitoringImpl.metricsStorage.Query(tenant1, from, to)
				if err != nil {
					t.Logf("Failed to query metrics for tenant1: %v", err)
					return false
				}

				// Verify we got only tenant1 metrics
				if len(metrics1) != 5 {
					t.Logf("Expected 5 metrics for tenant1, got %d", len(metrics1))
					return false
				}

				for _, m := range metrics1 {
					if m.TenantID != tenant1 {
						t.Logf("Got metric for wrong tenant: %s", m.TenantID)
						return false
					}
				}

				// Query metrics for tenant2
				metrics2, err := monitoringImpl.metricsStorage.Query(tenant2, from, to)
				if err != nil {
					t.Logf("Failed to query metrics for tenant2: %v", err)
					return false
				}

				// Verify we got only tenant2 metrics
				if len(metrics2) != 3 {
					t.Logf("Expected 3 metrics for tenant2, got %d", len(metrics2))
					return false
				}

				for _, m := range metrics2 {
					if m.TenantID != tenant2 {
						t.Logf("Got metric for wrong tenant: %s", m.TenantID)
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
			gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
