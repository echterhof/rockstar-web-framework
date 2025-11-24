package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Monitoring Example ===\n")

	// Create a simple in-memory database for metrics
	db := createMockDatabase()

	// Create metrics collector
	metrics := pkg.NewMetricsCollector(db)

	// Create logger
	logger := createMockLogger()

	// Configure monitoring with all features enabled
	config := pkg.MonitoringConfig{
		// Metrics endpoint configuration
		EnableMetrics: true,
		MetricsPath:   "/metrics",
		MetricsPort:   9090,

		// Pprof profiling configuration
		EnablePprof: true,
		PprofPath:   "/debug/pprof",
		PprofPort:   6060,

		// SNMP monitoring configuration
		EnableSNMP:    true,
		SNMPPort:      8161, // Using non-privileged port
		SNMPCommunity: "public",

		// Process optimization configuration
		EnableOptimization:   true,
		OptimizationInterval: 30 * time.Second,

		// Security configuration
		RequireAuth: false, // Disabled for demo
		AuthToken:   "demo-token",
	}

	// Create monitoring manager
	manager := pkg.NewMonitoringManager(config, metrics, db, logger)

	// Start monitoring
	fmt.Println("Starting monitoring manager...")
	if err := manager.Start(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("\nStopping monitoring manager...")
		manager.Stop()
	}()

	fmt.Println("✓ Monitoring manager started successfully\n")

	// Display available endpoints
	fmt.Println("Available Monitoring Endpoints:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Metrics:  http://localhost:9090/metrics")
	fmt.Println("🔍 Pprof:    http://localhost:6060/debug/pprof")
	fmt.Println("📡 SNMP:     http://localhost:8161/snmp")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Simulate some application activity
	fmt.Println("Simulating application activity...")
	simulateActivity(metrics)

	// Demonstrate metrics endpoint
	fmt.Println("\n1. Testing Metrics Endpoint")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	testMetricsEndpoint()

	// Demonstrate pprof
	fmt.Println("\n2. Testing Pprof Endpoint")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	testPprofEndpoint()

	// Demonstrate SNMP
	fmt.Println("\n3. Testing SNMP Endpoint")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	testSNMPEndpoint(manager)

	// Demonstrate process optimization
	fmt.Println("\n4. Testing Process Optimization")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	testOptimization(manager)

	// Demonstrate dynamic control
	fmt.Println("\n5. Testing Dynamic Control")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	testDynamicControl(manager)

	// Show pprof usage examples
	fmt.Println("\n6. Pprof Usage Examples")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	showPprofExamples()

	fmt.Println("\n✓ All monitoring features demonstrated successfully!")
	fmt.Println("\nPress Ctrl+C to exit...")

	// Keep running
	select {}
}

// simulateActivity simulates application activity by recording metrics
func simulateActivity(metrics pkg.MetricsCollector) {
	for i := 0; i < 10; i++ {
		// Start request metrics
		requestMetrics := metrics.Start(fmt.Sprintf("req-%d", i))
		requestMetrics.SetContext("tenant-1", "user-1", "/api/test", "GET")

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		// End metrics
		requestMetrics.SetResponse(200, 1024, nil)
		requestMetrics.End()

		// Record metrics
		metrics.Record(requestMetrics)

		// Record some custom metrics
		metrics.IncrementCounter("api.requests", map[string]string{
			"endpoint": "/api/test",
			"method":   "GET",
		})

		metrics.SetGauge("active.connections", float64(i+1), nil)
	}

	fmt.Println("✓ Simulated 10 API requests")
}

// testMetricsEndpoint tests the metrics HTTP endpoint
func testMetricsEndpoint() {
	resp, err := http.Get("http://localhost:9090/metrics")
	if err != nil {
		fmt.Printf("✗ Failed to get metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ Metrics endpoint is accessible")
		fmt.Println("  Status:", resp.StatusCode)
		fmt.Println("  Content-Type:", resp.Header.Get("Content-Type"))
		fmt.Println("  Try: curl http://localhost:9090/metrics | jq")
	} else {
		fmt.Printf("✗ Unexpected status code: %d\n", resp.StatusCode)
	}
}

// testPprofEndpoint tests the pprof HTTP endpoint
func testPprofEndpoint() {
	resp, err := http.Get("http://localhost:6060/debug/pprof/")
	if err != nil {
		fmt.Printf("✗ Failed to get pprof: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✓ Pprof endpoint is accessible")
		fmt.Println("  Status:", resp.StatusCode)
		fmt.Println("  Available profiles:")
		fmt.Println("    - heap: http://localhost:6060/debug/pprof/heap")
		fmt.Println("    - goroutine: http://localhost:6060/debug/pprof/goroutine")
		fmt.Println("    - profile: http://localhost:6060/debug/pprof/profile")
	} else {
		fmt.Printf("✗ Unexpected status code: %d\n", resp.StatusCode)
	}
}

// testSNMPEndpoint tests the SNMP endpoint
func testSNMPEndpoint(manager pkg.MonitoringManager) {
	data, err := manager.GetSNMPData()
	if err != nil {
		fmt.Printf("✗ Failed to get SNMP data: %v\n", err)
		return
	}

	fmt.Println("✓ SNMP data retrieved successfully")
	fmt.Printf("  Timestamp: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Printf("  CPUs: %d\n", data.SystemInfo.NumCPU)
	fmt.Printf("  Goroutines: %d\n", data.SystemInfo.NumGoroutine)
	fmt.Printf("  Memory Allocated: %d bytes\n", data.SystemInfo.MemoryAlloc)
	fmt.Printf("  GC Runs: %d\n", data.SystemInfo.NumGC)
}

// testOptimization tests process optimization
func testOptimization(manager pkg.MonitoringManager) {
	// Get stats before optimization
	statsBefore := manager.GetOptimizationStats()
	fmt.Printf("Before optimization:\n")
	fmt.Printf("  Optimization count: %d\n", statsBefore.OptimizationCount)

	// Run optimization
	fmt.Println("\nRunning optimization...")
	err := manager.OptimizeNow()
	if err != nil {
		fmt.Printf("✗ Optimization failed: %v\n", err)
		return
	}

	// Get stats after optimization
	statsAfter := manager.GetOptimizationStats()
	fmt.Printf("\n✓ Optimization completed\n")
	fmt.Printf("  Optimization count: %d\n", statsAfter.OptimizationCount)
	fmt.Printf("  Memory before: %d bytes\n", statsAfter.MemoryBefore)
	fmt.Printf("  Memory after: %d bytes\n", statsAfter.MemoryAfter)
	fmt.Printf("  Memory freed: %d bytes\n", statsAfter.MemoryFreed)
	fmt.Printf("  GC runs: %d\n", statsAfter.GCRunsAfter-statsAfter.GCRunsBefore)
}

// testDynamicControl demonstrates dynamic enable/disable of features
func testDynamicControl(manager pkg.MonitoringManager) {
	fmt.Println("Testing dynamic control of monitoring features...")

	// Test disabling and re-enabling metrics
	fmt.Println("\n  Disabling metrics endpoint...")
	if err := manager.DisableMetricsEndpoint(); err != nil {
		fmt.Printf("  ✗ Failed to disable metrics: %v\n", err)
	} else {
		fmt.Println("  ✓ Metrics endpoint disabled")
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("  Re-enabling metrics endpoint...")
	if err := manager.EnableMetricsEndpoint("/metrics"); err != nil {
		fmt.Printf("  ✗ Failed to enable metrics: %v\n", err)
	} else {
		fmt.Println("  ✓ Metrics endpoint re-enabled")
	}

	// Test configuration update
	fmt.Println("\n  Updating configuration...")
	config := manager.GetConfig()
	config.OptimizationInterval = 1 * time.Minute
	if err := manager.SetConfig(config); err != nil {
		fmt.Printf("  ✗ Failed to update config: %v\n", err)
	} else {
		fmt.Println("  ✓ Configuration updated")
	}
}

// showPprofExamples shows example commands for using pprof
func showPprofExamples() {
	fmt.Println("Command-line examples for profiling:")
	fmt.Println()
	fmt.Println("  # CPU Profile (30 seconds)")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile")
	fmt.Println()
	fmt.Println("  # Heap Profile")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println()
	fmt.Println("  # Goroutine Profile")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
	fmt.Println()
	fmt.Println("  # View goroutines in browser")
	fmt.Println("  curl http://localhost:6060/debug/pprof/goroutine?debug=1")
	fmt.Println()
	fmt.Println("  # Generate heap graph")
	fmt.Println("  go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap")
}

// Mock implementations for the example

type mockDatabase struct{}

func createMockDatabase() pkg.DatabaseManager {
	return &mockDatabase{}
}

func (m *mockDatabase) SaveWorkloadMetrics(metrics *pkg.WorkloadMetrics) error {
	return nil
}

func (m *mockDatabase) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*pkg.WorkloadMetrics, error) {
	return []*pkg.WorkloadMetrics{
		{
			RequestID:  "req-1",
			TenantID:   "tenant-1",
			Duration:   100,
			StatusCode: 200,
		},
	}, nil
}

// Implement other required methods as no-ops
func (m *mockDatabase) Connect(config pkg.DatabaseConfig) error { return nil }
func (m *mockDatabase) Close() error                            { return nil }
func (m *mockDatabase) Ping() error                             { return nil }
func (m *mockDatabase) Stats() pkg.DatabaseStats                { return pkg.DatabaseStats{} }
func (m *mockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockDatabase) QueryRow(query string, args ...interface{}) *sql.Row { return nil }
func (m *mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockDatabase) Prepare(query string) (*sql.Stmt, error) { return nil, nil }
func (m *mockDatabase) Begin() (pkg.Transaction, error)         { return nil, nil }
func (m *mockDatabase) BeginTx(opts *sql.TxOptions) (pkg.Transaction, error) {
	return nil, nil
}
func (m *mockDatabase) Save(model interface{}) error { return nil }
func (m *mockDatabase) Find(model interface{}, conditions ...pkg.Condition) error {
	return nil
}
func (m *mockDatabase) FindAll(models interface{}, conditions ...pkg.Condition) error {
	return nil
}
func (m *mockDatabase) Delete(model interface{}) error { return nil }
func (m *mockDatabase) Update(model interface{}, updates map[string]interface{}) error {
	return nil
}
func (m *mockDatabase) SaveSession(session *pkg.Session) error { return nil }
func (m *mockDatabase) LoadSession(sessionID string) (*pkg.Session, error) {
	return nil, nil
}
func (m *mockDatabase) DeleteSession(sessionID string) error { return nil }
func (m *mockDatabase) CleanupExpiredSessions() error        { return nil }
func (m *mockDatabase) SaveAccessToken(token *pkg.AccessToken) error {
	return nil
}
func (m *mockDatabase) LoadAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	return nil, nil
}
func (m *mockDatabase) ValidateAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	return nil, nil
}
func (m *mockDatabase) DeleteAccessToken(tokenValue string) error { return nil }
func (m *mockDatabase) CleanupExpiredTokens() error               { return nil }
func (m *mockDatabase) SaveTenant(tenant *pkg.Tenant) error       { return nil }
func (m *mockDatabase) LoadTenant(tenantID string) (*pkg.Tenant, error) {
	return nil, nil
}
func (m *mockDatabase) LoadTenantByHost(hostname string) (*pkg.Tenant, error) {
	return nil, nil
}
func (m *mockDatabase) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
func (m *mockDatabase) IncrementRateLimit(key string, window time.Duration) error {
	return nil
}
func (m *mockDatabase) Migrate() error      { return nil }
func (m *mockDatabase) CreateTables() error { return nil }
func (m *mockDatabase) DropTables() error   { return nil }

type mockLogger struct{}

func createMockLogger() pkg.Logger {
	return &mockLogger{}
}

func (l *mockLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s\n", msg)
}

func (l *mockLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s\n", msg)
}

func (l *mockLogger) Warn(msg string, fields ...interface{}) {
	fmt.Printf("[WARN] %s\n", msg)
}

func (l *mockLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s\n", msg)
}

func (l *mockLogger) WithRequestID(requestID string) pkg.Logger {
	return l
}
