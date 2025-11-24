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
