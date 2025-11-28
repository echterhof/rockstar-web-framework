package pkg

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewPluginMetrics(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	if metrics.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", metrics.Name)
	}

	if metrics.Status != PluginStatusUnloaded {
		t.Errorf("Expected status %s, got %s", PluginStatusUnloaded, metrics.Status)
	}

	if metrics.HookExecutions == nil {
		t.Error("Expected HookExecutions map to be initialized")
	}

	if metrics.HookDurations == nil {
		t.Error("Expected HookDurations map to be initialized")
	}
}

func TestPluginMetrics_RecordInit(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Test successful initialization
	duration := 100 * time.Millisecond
	metrics.RecordInit(duration, nil)

	if metrics.InitDuration != duration {
		t.Errorf("Expected init duration %v, got %v", duration, metrics.InitDuration)
	}

	if metrics.Status != PluginStatusInitialized {
		t.Errorf("Expected status %s, got %s", PluginStatusInitialized, metrics.Status)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", metrics.ErrorCount)
	}

	// Test failed initialization
	metrics2 := NewPluginMetrics("test-plugin-2")
	err := errors.New("init failed")
	metrics2.RecordInit(duration, err)

	if metrics2.Status != PluginStatusError {
		t.Errorf("Expected status %s, got %s", PluginStatusError, metrics2.Status)
	}

	if metrics2.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metrics2.ErrorCount)
	}

	if metrics2.LastError != err {
		t.Errorf("Expected last error to be set")
	}
}

func TestPluginMetrics_RecordStart(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Test successful start
	duration := 50 * time.Millisecond
	metrics.RecordStart(duration, nil)

	if metrics.StartDuration != duration {
		t.Errorf("Expected start duration %v, got %v", duration, metrics.StartDuration)
	}

	if metrics.Status != PluginStatusRunning {
		t.Errorf("Expected status %s, got %s", PluginStatusRunning, metrics.Status)
	}

	// Test failed start
	metrics2 := NewPluginMetrics("test-plugin-2")
	err := errors.New("start failed")
	metrics2.RecordStart(duration, err)

	if metrics2.Status != PluginStatusError {
		t.Errorf("Expected status %s, got %s", PluginStatusError, metrics2.Status)
	}

	if metrics2.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metrics2.ErrorCount)
	}
}

func TestPluginMetrics_RecordStop(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Test successful stop
	metrics.RecordStop(nil)

	if metrics.Status != PluginStatusStopped {
		t.Errorf("Expected status %s, got %s", PluginStatusStopped, metrics.Status)
	}

	// Test failed stop
	metrics2 := NewPluginMetrics("test-plugin-2")
	err := errors.New("stop failed")
	metrics2.RecordStop(err)

	if metrics2.Status != PluginStatusError {
		t.Errorf("Expected status %s, got %s", PluginStatusError, metrics2.Status)
	}

	if metrics2.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metrics2.ErrorCount)
	}
}

func TestPluginMetrics_RecordError(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	err1 := errors.New("error 1")
	metrics.RecordError(err1)

	if metrics.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metrics.ErrorCount)
	}

	if metrics.LastError != err1 {
		t.Errorf("Expected last error to be err1")
	}

	err2 := errors.New("error 2")
	metrics.RecordError(err2)

	if metrics.ErrorCount != 2 {
		t.Errorf("Expected error count 2, got %d", metrics.ErrorCount)
	}

	if metrics.LastError != err2 {
		t.Errorf("Expected last error to be err2")
	}
}

func TestPluginMetrics_RecordHookExecution(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Record first execution
	duration1 := 10 * time.Millisecond
	metrics.RecordHookExecution(HookTypePreRequest, duration1)

	if metrics.HookExecutions[HookTypePreRequest] != 1 {
		t.Errorf("Expected 1 execution, got %d", metrics.HookExecutions[HookTypePreRequest])
	}

	if metrics.HookDurations[HookTypePreRequest] != duration1 {
		t.Errorf("Expected duration %v, got %v", duration1, metrics.HookDurations[HookTypePreRequest])
	}

	// Record second execution
	duration2 := 15 * time.Millisecond
	metrics.RecordHookExecution(HookTypePreRequest, duration2)

	if metrics.HookExecutions[HookTypePreRequest] != 2 {
		t.Errorf("Expected 2 executions, got %d", metrics.HookExecutions[HookTypePreRequest])
	}

	expectedTotal := duration1 + duration2
	if metrics.HookDurations[HookTypePreRequest] != expectedTotal {
		t.Errorf("Expected total duration %v, got %v", expectedTotal, metrics.HookDurations[HookTypePreRequest])
	}

	// Record different hook type
	metrics.RecordHookExecution(HookTypePostRequest, duration1)

	if metrics.HookExecutions[HookTypePostRequest] != 1 {
		t.Errorf("Expected 1 execution for PostRequest, got %d", metrics.HookExecutions[HookTypePostRequest])
	}
}

func TestPluginMetrics_RecordEvents(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	metrics.RecordEventPublished()
	metrics.RecordEventPublished()

	if metrics.EventsPublished != 2 {
		t.Errorf("Expected 2 events published, got %d", metrics.EventsPublished)
	}

	metrics.RecordEventReceived()

	if metrics.EventsReceived != 1 {
		t.Errorf("Expected 1 event received, got %d", metrics.EventsReceived)
	}
}

func TestPluginMetrics_RecordServiceCall(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	metrics.RecordServiceCall()
	metrics.RecordServiceCall()
	metrics.RecordServiceCall()

	if metrics.ServiceCalls != 3 {
		t.Errorf("Expected 3 service calls, got %d", metrics.ServiceCalls)
	}
}

func TestPluginMetrics_GetHookMetrics(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Record some hook executions
	metrics.RecordHookExecution(HookTypePreRequest, 10*time.Millisecond)
	metrics.RecordHookExecution(HookTypePreRequest, 20*time.Millisecond)
	metrics.RecordHookExecution(HookTypePostRequest, 15*time.Millisecond)

	hookMetrics := metrics.GetHookMetrics()

	// Check PreRequest metrics
	preReqMetrics, exists := hookMetrics[HookTypePreRequest]
	if !exists {
		t.Fatal("Expected PreRequest metrics to exist")
	}

	if preReqMetrics.ExecutionCount != 2 {
		t.Errorf("Expected 2 executions, got %d", preReqMetrics.ExecutionCount)
	}

	expectedTotal := 30 * time.Millisecond
	if preReqMetrics.TotalDuration != expectedTotal {
		t.Errorf("Expected total duration %v, got %v", expectedTotal, preReqMetrics.TotalDuration)
	}

	expectedAvg := 15 * time.Millisecond
	if preReqMetrics.AverageDuration != expectedAvg {
		t.Errorf("Expected average duration %v, got %v", expectedAvg, preReqMetrics.AverageDuration)
	}

	// Check PostRequest metrics
	postReqMetrics, exists := hookMetrics[HookTypePostRequest]
	if !exists {
		t.Fatal("Expected PostRequest metrics to exist")
	}

	if postReqMetrics.ExecutionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", postReqMetrics.ExecutionCount)
	}
}

func TestPluginMetrics_Clone(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")
	metrics.RecordInit(100*time.Millisecond, nil)
	metrics.RecordHookExecution(HookTypePreRequest, 10*time.Millisecond)
	metrics.RecordEventPublished()

	clone := metrics.Clone()

	// Verify clone has same values
	if clone.Name != metrics.Name {
		t.Errorf("Clone name mismatch")
	}

	if clone.Status != metrics.Status {
		t.Errorf("Clone status mismatch")
	}

	if clone.InitDuration != metrics.InitDuration {
		t.Errorf("Clone init duration mismatch")
	}

	if clone.EventsPublished != metrics.EventsPublished {
		t.Errorf("Clone events published mismatch")
	}

	// Verify modifying clone doesn't affect original
	clone.RecordEventPublished()

	if metrics.EventsPublished == clone.EventsPublished {
		t.Error("Modifying clone affected original")
	}
}

func TestPluginMetrics_ExportPrometheus(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")
	metrics.RecordInit(100*time.Millisecond, nil)
	metrics.RecordStart(50*time.Millisecond, nil)
	metrics.RecordHookExecution(HookTypePreRequest, 10*time.Millisecond)
	metrics.RecordEventPublished()
	metrics.RecordServiceCall()

	output := metrics.ExportPrometheus()

	// Check that output contains expected metrics
	expectedMetrics := []string{
		"rockstar_plugin_status",
		"rockstar_plugin_errors_total",
		"rockstar_plugin_init_duration_seconds",
		"rockstar_plugin_start_duration_seconds",
		"rockstar_plugin_hook_executions_total",
		"rockstar_plugin_hook_duration_seconds",
		"rockstar_plugin_events_published_total",
		"rockstar_plugin_events_received_total",
		"rockstar_plugin_service_calls_total",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("Expected output to contain metric '%s'", metric)
		}
	}

	// Check that plugin name is included
	if !strings.Contains(output, "test-plugin") {
		t.Error("Expected output to contain plugin name")
	}
}

func TestPluginMetricsCollector_GetOrCreate(t *testing.T) {
	collector := NewPluginMetricsCollector()

	// Get or create first time
	metrics1 := collector.GetOrCreate("plugin1")
	if metrics1 == nil {
		t.Fatal("Expected metrics to be created")
	}

	if metrics1.Name != "plugin1" {
		t.Errorf("Expected name 'plugin1', got '%s'", metrics1.Name)
	}

	// Get or create second time (should return same instance)
	metrics2 := collector.GetOrCreate("plugin1")
	if metrics1 != metrics2 {
		t.Error("Expected same metrics instance to be returned")
	}
}

func TestPluginMetricsCollector_Get(t *testing.T) {
	collector := NewPluginMetricsCollector()

	// Get non-existent plugin
	metrics := collector.Get("nonexistent")
	if metrics != nil {
		t.Error("Expected nil for non-existent plugin")
	}

	// Create and get
	collector.GetOrCreate("plugin1")
	metrics = collector.Get("plugin1")
	if metrics == nil {
		t.Error("Expected metrics to exist")
	}
}

func TestPluginMetricsCollector_GetAll(t *testing.T) {
	collector := NewPluginMetricsCollector()

	// Create multiple plugins
	collector.GetOrCreate("plugin1")
	collector.GetOrCreate("plugin2")
	collector.GetOrCreate("plugin3")

	all := collector.GetAll()

	if len(all) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(all))
	}

	// Verify all plugins are present
	if _, exists := all["plugin1"]; !exists {
		t.Error("Expected plugin1 to exist")
	}

	if _, exists := all["plugin2"]; !exists {
		t.Error("Expected plugin2 to exist")
	}

	if _, exists := all["plugin3"]; !exists {
		t.Error("Expected plugin3 to exist")
	}

	// Verify returned metrics are clones (modifying shouldn't affect collector)
	all["plugin1"].RecordEventPublished()
	original := collector.Get("plugin1")
	if original.EventsPublished != 0 {
		t.Error("Modifying returned metrics affected collector")
	}
}

func TestPluginMetricsCollector_Remove(t *testing.T) {
	collector := NewPluginMetricsCollector()

	collector.GetOrCreate("plugin1")
	collector.Remove("plugin1")

	metrics := collector.Get("plugin1")
	if metrics != nil {
		t.Error("Expected plugin to be removed")
	}
}

func TestPluginMetricsCollector_ExportPrometheus(t *testing.T) {
	collector := NewPluginMetricsCollector()

	// Create multiple plugins with metrics
	metrics1 := collector.GetOrCreate("plugin1")
	metrics1.RecordInit(100*time.Millisecond, nil)
	metrics1.RecordStart(50*time.Millisecond, nil)

	metrics2 := collector.GetOrCreate("plugin2")
	metrics2.RecordInit(200*time.Millisecond, nil)

	output := collector.ExportPrometheus()

	// Check for metric type definitions
	if !strings.Contains(output, "# HELP") {
		t.Error("Expected output to contain HELP comments")
	}

	if !strings.Contains(output, "# TYPE") {
		t.Error("Expected output to contain TYPE comments")
	}

	// Check for both plugins
	if !strings.Contains(output, "plugin1") {
		t.Error("Expected output to contain plugin1")
	}

	if !strings.Contains(output, "plugin2") {
		t.Error("Expected output to contain plugin2")
	}

	// Check for metric names
	if !strings.Contains(output, "rockstar_plugin_status") {
		t.Error("Expected output to contain status metric")
	}
}

func TestPluginMetrics_ConcurrentAccess(t *testing.T) {
	metrics := NewPluginMetrics("test-plugin")

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordHookExecution(HookTypePreRequest, time.Millisecond)
				metrics.RecordEventPublished()
				metrics.RecordServiceCall()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counts
	if metrics.HookExecutions[HookTypePreRequest] != 1000 {
		t.Errorf("Expected 1000 hook executions, got %d", metrics.HookExecutions[HookTypePreRequest])
	}

	if metrics.EventsPublished != 1000 {
		t.Errorf("Expected 1000 events published, got %d", metrics.EventsPublished)
	}

	if metrics.ServiceCalls != 1000 {
		t.Errorf("Expected 1000 service calls, got %d", metrics.ServiceCalls)
	}
}
