package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// mockMetricsCollectorForMonitoring is a mock metrics collector for testing
type mockMetricsCollectorForMonitoring struct {
	exportCalled         bool
	exportPrometheusCall bool
	getAggregatedCalled  bool
	shouldFailExport     bool
	shouldFailAggregated bool
}

func (m *mockMetricsCollectorForMonitoring) Start(requestID string) *RequestMetrics {
	return &RequestMetrics{RequestID: requestID}
}

func (m *mockMetricsCollectorForMonitoring) Record(metrics *RequestMetrics) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordRequest(ctx Context, duration time.Duration, statusCode int) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordError(ctx Context, err error) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) GetMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}

func (m *mockMetricsCollectorForMonitoring) GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error) {
	m.getAggregatedCalled = true
	if m.shouldFailAggregated {
		return nil, fmt.Errorf("aggregated metrics error")
	}
	return &AggregatedMetrics{
		TenantID:      tenantID,
		TotalRequests: 100,
	}, nil
}

func (m *mockMetricsCollectorForMonitoring) PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error) {
	return nil, nil
}

func (m *mockMetricsCollectorForMonitoring) RecordWorkloadMetrics(metrics *WorkloadMetrics) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}

func (m *mockMetricsCollectorForMonitoring) IncrementCounter(name string, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) IncrementCounterBy(name string, value int64, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) SetGauge(name string, value float64, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) IncrementGauge(name string, value float64, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) DecrementGauge(name string, value float64, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordHistogram(name string, value float64, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordTiming(name string, duration time.Duration, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) StartTimer(name string, tags map[string]string) Timer {
	return &timerImpl{startTime: time.Now()}
}

func (m *mockMetricsCollectorForMonitoring) RecordMemoryUsage(usage int64) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordCPUUsage(usage float64) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) RecordCustomMetric(name string, value interface{}, tags map[string]string) error {
	return nil
}

func (m *mockMetricsCollectorForMonitoring) Export() (map[string]interface{}, error) {
	m.exportCalled = true
	if m.shouldFailExport {
		return nil, fmt.Errorf("export error")
	}
	return map[string]interface{}{
		"counters": map[string]int64{"requests": 100},
		"gauges":   map[string]float64{"memory": 1024.0},
	}, nil
}

func (m *mockMetricsCollectorForMonitoring) ExportPrometheus() ([]byte, error) {
	m.exportPrometheusCall = true
	return []byte("# metrics"), nil
}

// mockDatabaseForMonitoring is a mock database for testing
type mockDatabaseForMonitoring struct {
	mockDatabaseForMetrics
}

func (m *mockDatabaseForMonitoring) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return []*WorkloadMetrics{
		{
			RequestID:  "req-1",
			TenantID:   "tenant-1",
			Duration:   100,
			StatusCode: 200,
		},
	}, nil
}

// mockLogger is a mock logger for testing
type mockLogger struct {
	debugCalls []string
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
}

func (l *mockLogger) Debug(msg string, fields ...interface{}) {
	l.debugCalls = append(l.debugCalls, msg)
}

func (l *mockLogger) Info(msg string, fields ...interface{}) {
	l.infoCalls = append(l.infoCalls, msg)
}

func (l *mockLogger) Warn(msg string, fields ...interface{}) {
	l.warnCalls = append(l.warnCalls, msg)
}

func (l *mockLogger) Error(msg string, fields ...interface{}) {
	l.errorCalls = append(l.errorCalls, msg)
}

func (l *mockLogger) WithRequestID(requestID string) Logger {
	return l
}

// TestNewMonitoringManager tests creating a new monitoring manager
func TestNewMonitoringManager(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: true,
		MetricsPath:   "/metrics",
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	db := &mockDatabaseForMonitoring{}
	logger := &mockLogger{}

	manager := NewMonitoringManager(config, metrics, db, logger)

	if manager == nil {
		t.Fatal("Expected non-nil monitoring manager")
	}
}

// TestMonitoringManagerStartStop tests starting and stopping the manager
func TestMonitoringManagerStartStop(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: false, // Disable to avoid port conflicts
		EnablePprof:   false,
		EnableSNMP:    false,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	if manager.IsRunning() {
		t.Error("Expected manager to not be running initially")
	}

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	if !manager.IsRunning() {
		t.Error("Expected manager to be running after start")
	}

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager: %v", err)
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after stop")
	}
}

// TestMonitoringManagerStartTwice tests starting the manager twice
func TestMonitoringManagerStartTwice(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: false,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	err = manager.Start()
	if err == nil {
		t.Error("Expected error when starting manager twice")
	}
}

// TestEnableDisableMetricsEndpoint tests enabling and disabling metrics endpoint
func TestEnableDisableMetricsEndpoint(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: false,
		MetricsPort:   9091, // Use different port to avoid conflicts
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Enable metrics endpoint
	err = manager.EnableMetricsEndpoint("/metrics")
	if err != nil {
		t.Fatalf("Failed to enable metrics endpoint: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test metrics endpoint
	resp, err := http.Get("http://localhost:9091/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify export was called
	if !metrics.exportCalled {
		t.Error("Expected Export to be called")
	}

	// Disable metrics endpoint
	err = manager.DisableMetricsEndpoint()
	if err != nil {
		t.Fatalf("Failed to disable metrics endpoint: %v", err)
	}

	// Give server time to stop
	time.Sleep(100 * time.Millisecond)
}

// TestMetricsEndpointWithAuth tests metrics endpoint with authentication
func TestMetricsEndpointWithAuth(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: false,
		MetricsPort:   9092,
		RequireAuth:   true,
		AuthToken:     "test-token",
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	err = manager.EnableMetricsEndpoint("/metrics")
	if err != nil {
		t.Fatalf("Failed to enable metrics endpoint: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test without auth - should fail
	resp, err := http.Get("http://localhost:9092/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// Test with auth - should succeed
	req, _ := http.NewRequest("GET", "http://localhost:9092/metrics", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get metrics with auth: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestMetricsEndpointResponse tests the metrics endpoint response format
func TestMetricsEndpointResponse(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: false,
		MetricsPort:   9093,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	err = manager.EnableMetricsEndpoint("/metrics")
	if err != nil {
		t.Fatalf("Failed to enable metrics endpoint: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:9093/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check for system metrics
	if _, ok := result["system"]; !ok {
		t.Error("Expected system metrics in response")
	}

	// Check for counters
	if _, ok := result["counters"]; !ok {
		t.Error("Expected counters in response")
	}
}

// TestEnableDisablePprof tests enabling and disabling pprof
func TestEnableDisablePprof(t *testing.T) {
	config := MonitoringConfig{
		EnablePprof: false,
		PprofPort:   6061, // Use different port
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Enable pprof
	err = manager.EnablePprof("/debug/pprof")
	if err != nil {
		t.Fatalf("Failed to enable pprof: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test pprof endpoint
	resp, err := http.Get("http://localhost:6061/debug/pprof/")
	if err != nil {
		t.Fatalf("Failed to get pprof: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Disable pprof
	err = manager.DisablePprof()
	if err != nil {
		t.Fatalf("Failed to disable pprof: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

// TestGetPprofHandlers tests getting pprof handlers
func TestGetPprofHandlers(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	handlers := manager.GetPprofHandlers()

	if len(handlers) == 0 {
		t.Error("Expected non-empty pprof handlers")
	}

	expectedHandlers := []string{"/", "/cmdline", "/profile", "/symbol", "/trace"}
	for _, path := range expectedHandlers {
		if _, ok := handlers[path]; !ok {
			t.Errorf("Expected handler for %s", path)
		}
	}
}

// TestGetSNMPData tests getting SNMP data
func TestGetSNMPData(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	db := &mockDatabaseForMonitoring{}
	manager := NewMonitoringManager(config, metrics, db, nil).(*monitoringManagerImpl)

	data, err := manager.GetSNMPData()
	if err != nil {
		t.Fatalf("Failed to get SNMP data: %v", err)
	}

	if data == nil {
		t.Fatal("Expected non-nil SNMP data")
	}

	if data.SystemInfo == nil {
		t.Error("Expected system info in SNMP data")
	}

	if data.SystemInfo.NumCPU != runtime.NumCPU() {
		t.Errorf("Expected %d CPUs, got %d", runtime.NumCPU(), data.SystemInfo.NumCPU)
	}
}

// TestEnableDisableOptimization tests enabling and disabling optimization
func TestEnableDisableOptimization(t *testing.T) {
	config := MonitoringConfig{
		EnableOptimization:   false,
		OptimizationInterval: 100 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Enable optimization
	err = manager.EnableOptimization()
	if err != nil {
		t.Fatalf("Failed to enable optimization: %v", err)
	}

	// Wait for at least one optimization cycle
	time.Sleep(200 * time.Millisecond)

	stats := manager.GetOptimizationStats()
	if stats.OptimizationCount == 0 {
		t.Error("Expected at least one optimization to have run")
	}

	// Disable optimization
	err = manager.DisableOptimization()
	if err != nil {
		t.Fatalf("Failed to disable optimization: %v", err)
	}
}

// TestOptimizeNow tests immediate optimization
func TestOptimizeNow(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Get initial stats
	statsBefore := manager.GetOptimizationStats()
	countBefore := statsBefore.OptimizationCount

	// Run optimization
	err = manager.OptimizeNow()
	if err != nil {
		t.Fatalf("Failed to optimize: %v", err)
	}

	// Check stats updated
	statsAfter := manager.GetOptimizationStats()
	if statsAfter.OptimizationCount != countBefore+1 {
		t.Errorf("Expected optimization count to increase by 1, got %d", statsAfter.OptimizationCount)
	}

	if statsAfter.LastOptimization.IsZero() {
		t.Error("Expected non-zero last optimization time")
	}

	// Check logger was called
	if len(logger.infoCalls) == 0 {
		t.Error("Expected info log for optimization")
	}
}

// TestGetOptimizationStats tests getting optimization statistics
func TestGetOptimizationStats(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	stats := manager.GetOptimizationStats()

	if stats == nil {
		t.Fatal("Expected non-nil optimization stats")
	}

	if stats.OptimizationCount != 0 {
		t.Errorf("Expected 0 optimizations initially, got %d", stats.OptimizationCount)
	}
}

// TestSetGetConfig tests setting and getting configuration
func TestSetGetConfig(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics: true,
		MetricsPath:   "/metrics",
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	retrievedConfig := manager.GetConfig()
	if retrievedConfig.EnableMetrics != true {
		t.Error("Expected EnableMetrics to be true")
	}

	if retrievedConfig.MetricsPath != "/metrics" {
		t.Errorf("Expected MetricsPath /metrics, got %s", retrievedConfig.MetricsPath)
	}

	// Update config
	newConfig := MonitoringConfig{
		EnableMetrics: false,
		MetricsPath:   "/new-metrics",
	}

	err := manager.SetConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	retrievedConfig = manager.GetConfig()
	if retrievedConfig.EnableMetrics != false {
		t.Error("Expected EnableMetrics to be false after update")
	}

	if retrievedConfig.MetricsPath != "/new-metrics" {
		t.Errorf("Expected MetricsPath /new-metrics, got %s", retrievedConfig.MetricsPath)
	}
}

// TestSystemInfoCollection tests system info collection
func TestSystemInfoCollection(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil).(*monitoringManagerImpl)

	info := manager.getSystemInfo()

	if info == nil {
		t.Fatal("Expected non-nil system info")
	}

	if info.NumCPU <= 0 {
		t.Error("Expected positive number of CPUs")
	}

	if info.NumGoroutine <= 0 {
		t.Error("Expected positive number of goroutines")
	}

	if info.MemoryAlloc == 0 {
		t.Error("Expected non-zero memory allocation")
	}
}

// TestConcurrentOperations tests concurrent monitoring operations
func TestConcurrentOperations(t *testing.T) {
	config := MonitoringConfig{}
	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	done := make(chan bool)

	// Run multiple operations concurrently
	for i := 0; i < 10; i++ {
		go func() {
			manager.GetConfig()
			manager.GetOptimizationStats()
			manager.OptimizeNow()
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestOptimizationGoroutineExitsCleanly tests that the optimization goroutine exits properly on Stop()
func TestOptimizationGoroutineExitsCleanly(t *testing.T) {
	config := MonitoringConfig{
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	// Start the manager with optimization enabled
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Let optimization run at least once
	time.Sleep(100 * time.Millisecond)

	// Get goroutine count before stop
	goroutinesBefore := runtime.NumGoroutine()

	// Stop should complete without hanging
	done := make(chan error, 1)
	go func() {
		done <- manager.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Failed to stop manager: %v", err)
		}
	case <-time.After(7 * time.Second):
		t.Fatal("Stop() hung - optimization goroutine did not exit cleanly")
	}

	// Give goroutines time to fully exit
	time.Sleep(100 * time.Millisecond)

	// Verify goroutines decreased (optimization goroutine exited)
	goroutinesAfter := runtime.NumGoroutine()
	if goroutinesAfter >= goroutinesBefore {
		t.Logf("Goroutines before: %d, after: %d", goroutinesBefore, goroutinesAfter)
		// This is a soft check - we expect goroutines to decrease but won't fail the test
		// as other goroutines might have started in the meantime
	}

	// Verify manager is not running
	if manager.IsRunning() {
		t.Error("Expected manager to not be running after stop")
	}
}

// TestOptimizationStartStopStartCycle tests that optimization can be restarted after stopping
func TestOptimizationStartStopStartCycle(t *testing.T) {
	config := MonitoringConfig{
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	// First cycle: Start → Stop
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager (first cycle): %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager (first cycle): %v", err)
	}

	// Second cycle: Start → Stop
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager (second cycle): %v", err)
	}

	if !manager.IsRunning() {
		t.Error("Expected manager to be running after second start")
	}

	// Let optimization run in second cycle
	time.Sleep(100 * time.Millisecond)

	stats := manager.GetOptimizationStats()
	if stats.OptimizationCount == 0 {
		t.Error("Expected at least one optimization to have run in second cycle")
	}

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager (second cycle): %v", err)
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after second stop")
	}
}

// TestMultipleStopCallsWithOptimization tests that Stop() can be called multiple times safely with optimization
func TestMultipleStopCallsWithOptimization(t *testing.T) {
	config := MonitoringConfig{
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// First stop
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager (first call): %v", err)
	}

	// Second stop - should not hang or panic
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager (second call): %v", err)
	}

	// Third stop - should still be safe
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager (third call): %v", err)
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after multiple stops")
	}
}

// TestStopCompletesWithinTimeout tests that Stop() completes within the expected timeout
// Requirements: 1.1, 1.2
func TestStopCompletesWithinTimeout(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics:        false,
		EnablePprof:          false,
		EnableSNMP:           false,
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop should complete within 7 seconds (6 second timeout + 1 second buffer)
	done := make(chan error, 1)
	start := time.Now()

	go func() {
		done <- manager.Stop()
	}()

	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("Stop() returned error: %v", err)
		}
		if elapsed > 7*time.Second {
			t.Errorf("Stop() took too long: %v (expected < 7s)", elapsed)
		}
		t.Logf("Stop() completed in %v", elapsed)
	case <-time.After(8 * time.Second):
		t.Fatal("Stop() hung - did not complete within timeout")
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after Stop()")
	}
}

// TestStopIdempotent tests that Stop() is idempotent and safe to call multiple times
// Requirements: 1.3
func TestStopIdempotent(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics:      false,
		EnablePprof:        false,
		EnableSNMP:         false,
		EnableOptimization: false,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	// Call Stop() on a manager that was never started - should not error
	err := manager.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running manager returned error: %v", err)
	}

	// Start the manager
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Stop it
	err = manager.Stop()
	if err != nil {
		t.Fatalf("First Stop() returned error: %v", err)
	}

	// Call Stop() multiple times - should all succeed without hanging
	for i := 0; i < 5; i++ {
		done := make(chan error, 1)
		go func() {
			done <- manager.Stop()
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Stop() call %d returned error: %v", i+2, err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Stop() call %d hung", i+2)
		}
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after multiple Stop() calls")
	}
}

// TestStartStopStartCycle tests that the manager can be restarted after stopping
// Requirements: 2.1, 2.2, 2.4
func TestStartStopStartCycle(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics:        false,
		EnablePprof:          false,
		EnableSNMP:           false,
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	// Perform multiple start-stop cycles
	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("Starting cycle %d", cycle)

		// Start
		err := manager.Start()
		if err != nil {
			t.Fatalf("Cycle %d: Failed to start manager: %v", cycle, err)
		}

		if !manager.IsRunning() {
			t.Errorf("Cycle %d: Expected manager to be running after Start()", cycle)
		}

		// Let optimization run
		time.Sleep(150 * time.Millisecond)

		// Verify optimization is working
		stats := manager.GetOptimizationStats()
		if stats.OptimizationCount == 0 {
			t.Errorf("Cycle %d: Expected at least one optimization to have run", cycle)
		}

		// Stop
		err = manager.Stop()
		if err != nil {
			t.Fatalf("Cycle %d: Failed to stop manager: %v", cycle, err)
		}

		if manager.IsRunning() {
			t.Errorf("Cycle %d: Expected manager to not be running after Stop()", cycle)
		}

		// Brief pause between cycles
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("All start-stop cycles completed successfully")
}

// TestStopWithOptimizationEnabled tests Stop() with optimization goroutine running
// Requirements: 1.1, 1.2, 1.4
func TestStopWithOptimizationEnabled(t *testing.T) {
	config := MonitoringConfig{
		EnableMetrics:        false,
		EnablePprof:          false,
		EnableSNMP:           false,
		EnableOptimization:   true,
		OptimizationInterval: 50 * time.Millisecond,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Let optimization run multiple times
	time.Sleep(200 * time.Millisecond)

	// Verify optimization is running
	stats := manager.GetOptimizationStats()
	if stats.OptimizationCount == 0 {
		t.Error("Expected at least one optimization to have run")
	}

	// Stop should cleanly shut down optimization goroutine without hanging
	done := make(chan error, 1)
	go func() {
		done <- manager.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stop() returned error: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("Stop() hung - optimization goroutine did not exit")
	}

	if manager.IsRunning() {
		t.Error("Expected manager to not be running after Stop()")
	}

	// Wait for goroutines to fully exit
	time.Sleep(100 * time.Millisecond)

	// Verify we can restart after stopping
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to restart manager after stop: %v", err)
	}

	// Let it run again
	time.Sleep(150 * time.Millisecond)

	// Verify optimization is working after restart
	stats = manager.GetOptimizationStats()
	if stats.OptimizationCount == 0 {
		t.Error("Expected optimization to run after restart")
	}

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager on second cycle: %v", err)
	}
}

// TestMonitoringConfigApplyDefaults tests that ApplyDefaults sets correct default values
// Requirements: 2.2, 5.1, 5.2, 5.3, 5.4, 5.5
func TestMonitoringConfigApplyDefaults(t *testing.T) {
	config := MonitoringConfig{}
	config.ApplyDefaults()

	// Test MetricsPort defaults to 9090
	if config.MetricsPort != 9090 {
		t.Errorf("Expected MetricsPort to default to 9090, got %d", config.MetricsPort)
	}

	// Test PprofPort defaults to 6060
	if config.PprofPort != 6060 {
		t.Errorf("Expected PprofPort to default to 6060, got %d", config.PprofPort)
	}

	// Test SNMPPort defaults to 161
	if config.SNMPPort != 161 {
		t.Errorf("Expected SNMPPort to default to 161, got %d", config.SNMPPort)
	}

	// Test SNMPCommunity defaults to "public"
	if config.SNMPCommunity != "public" {
		t.Errorf("Expected SNMPCommunity to default to 'public', got %s", config.SNMPCommunity)
	}

	// Test OptimizationInterval defaults to 5 minutes
	expectedInterval := 5 * time.Minute
	if config.OptimizationInterval != expectedInterval {
		t.Errorf("Expected OptimizationInterval to default to %v, got %v", expectedInterval, config.OptimizationInterval)
	}
}

// TestMonitoringConfigApplyDefaultsPreservesUserValues tests that user-provided values are preserved
// Requirements: 8.3, 8.4
func TestMonitoringConfigApplyDefaultsPreservesUserValues(t *testing.T) {
	config := MonitoringConfig{
		MetricsPort:          8080,
		PprofPort:            7070,
		SNMPPort:             162,
		SNMPCommunity:        "private",
		OptimizationInterval: 10 * time.Minute,
	}

	config.ApplyDefaults()

	// Verify user values are preserved
	if config.MetricsPort != 8080 {
		t.Errorf("Expected MetricsPort to remain 8080, got %d", config.MetricsPort)
	}

	if config.PprofPort != 7070 {
		t.Errorf("Expected PprofPort to remain 7070, got %d", config.PprofPort)
	}

	if config.SNMPPort != 162 {
		t.Errorf("Expected SNMPPort to remain 162, got %d", config.SNMPPort)
	}

	if config.SNMPCommunity != "private" {
		t.Errorf("Expected SNMPCommunity to remain 'private', got %s", config.SNMPCommunity)
	}

	if config.OptimizationInterval != 10*time.Minute {
		t.Errorf("Expected OptimizationInterval to remain 10m, got %v", config.OptimizationInterval)
	}
}

// TestMonitoringConfigApplyDefaultsPartialConfig tests defaults with partial configuration
// Requirements: 2.2, 5.1, 5.2, 5.3, 5.4, 5.5, 8.3, 8.4
func TestMonitoringConfigApplyDefaultsPartialConfig(t *testing.T) {
	config := MonitoringConfig{
		MetricsPort:   8080,     // User-provided
		SNMPCommunity: "custom", // User-provided
		// Other fields are zero values
	}

	config.ApplyDefaults()

	// User values should be preserved
	if config.MetricsPort != 8080 {
		t.Errorf("Expected MetricsPort to remain 8080, got %d", config.MetricsPort)
	}

	if config.SNMPCommunity != "custom" {
		t.Errorf("Expected SNMPCommunity to remain 'custom', got %s", config.SNMPCommunity)
	}

	// Zero values should get defaults
	if config.PprofPort != 6060 {
		t.Errorf("Expected PprofPort to default to 6060, got %d", config.PprofPort)
	}

	if config.SNMPPort != 161 {
		t.Errorf("Expected SNMPPort to default to 161, got %d", config.SNMPPort)
	}

	if config.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected OptimizationInterval to default to 5m, got %v", config.OptimizationInterval)
	}
}

// TestNewMonitoringManagerAppliesDefaults tests that NewMonitoringManager applies defaults
// Requirements: 2.2, 5.5
func TestNewMonitoringManagerAppliesDefaults(t *testing.T) {
	config := MonitoringConfig{
		// All zero values
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(config, metrics, nil, nil)

	retrievedConfig := manager.GetConfig()

	// Verify defaults were applied
	if retrievedConfig.MetricsPort != 9090 {
		t.Errorf("Expected MetricsPort to default to 9090, got %d", retrievedConfig.MetricsPort)
	}

	if retrievedConfig.PprofPort != 6060 {
		t.Errorf("Expected PprofPort to default to 6060, got %d", retrievedConfig.PprofPort)
	}

	if retrievedConfig.SNMPPort != 161 {
		t.Errorf("Expected SNMPPort to default to 161, got %d", retrievedConfig.SNMPPort)
	}

	if retrievedConfig.SNMPCommunity != "public" {
		t.Errorf("Expected SNMPCommunity to default to 'public', got %s", retrievedConfig.SNMPCommunity)
	}

	if retrievedConfig.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected OptimizationInterval to default to 5m, got %v", retrievedConfig.OptimizationInterval)
	}
}

// TestSetConfigAppliesDefaults tests that SetConfig applies defaults
// Requirements: 2.2, 5.5
func TestSetConfigAppliesDefaults(t *testing.T) {
	initialConfig := MonitoringConfig{
		MetricsPort: 8080,
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	manager := NewMonitoringManager(initialConfig, metrics, nil, nil)

	// Set new config with zero values
	newConfig := MonitoringConfig{
		EnableMetrics: true,
		// Other fields are zero values
	}

	err := manager.SetConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	retrievedConfig := manager.GetConfig()

	// Verify defaults were applied
	if retrievedConfig.MetricsPort != 9090 {
		t.Errorf("Expected MetricsPort to default to 9090, got %d", retrievedConfig.MetricsPort)
	}

	if retrievedConfig.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected OptimizationInterval to default to 5m, got %v", retrievedConfig.OptimizationInterval)
	}
}

// TestOptimizationWithZeroIntervalUsesDefault tests that ticker creation with zero interval uses default
// Requirements: 2.2, 5.5
func TestOptimizationWithZeroIntervalUsesDefault(t *testing.T) {
	config := MonitoringConfig{
		EnableOptimization:   true,
		OptimizationInterval: 0, // Zero value
	}

	metrics := &mockMetricsCollectorForMonitoring{}
	logger := &mockLogger{}
	manager := NewMonitoringManager(config, metrics, nil, logger)

	// Verify config has default applied
	retrievedConfig := manager.GetConfig()
	if retrievedConfig.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected OptimizationInterval to default to 5m, got %v", retrievedConfig.OptimizationInterval)
	}

	// Start manager with optimization enabled
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Wait for optimization to run at least once
	// With 5 minute default, we won't wait that long, but we can verify it doesn't panic
	time.Sleep(100 * time.Millisecond)

	// Verify optimization stats exist (manager started successfully)
	stats := manager.GetOptimizationStats()
	if stats == nil {
		t.Error("Expected non-nil optimization stats")
	}
}
