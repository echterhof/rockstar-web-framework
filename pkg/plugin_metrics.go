package pkg

import (
	"fmt"
	"sync"
	"time"
)

// PluginMetrics tracks metrics for a plugin
type PluginMetrics struct {
	Name            string
	Status          PluginStatus
	InitDuration    time.Duration
	StartDuration   time.Duration
	ErrorCount      int64
	LastError       error
	LastErrorAt     time.Time
	HookExecutions  map[HookType]int64
	HookDurations   map[HookType]time.Duration
	EventsPublished int64
	EventsReceived  int64
	ServiceCalls    int64
	mu              sync.RWMutex
}

// NewPluginMetrics creates a new PluginMetrics instance
func NewPluginMetrics(name string) *PluginMetrics {
	return &PluginMetrics{
		Name:           name,
		Status:         PluginStatusUnloaded,
		HookExecutions: make(map[HookType]int64),
		HookDurations:  make(map[HookType]time.Duration),
	}
}

// RecordInit records initialization metrics
func (pm *PluginMetrics) RecordInit(duration time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.InitDuration = duration
	if err != nil {
		pm.ErrorCount++
		pm.LastError = err
		pm.LastErrorAt = time.Now()
		pm.Status = PluginStatusError
	} else {
		pm.Status = PluginStatusInitialized
	}
}

// RecordStart records start metrics
func (pm *PluginMetrics) RecordStart(duration time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.StartDuration = duration
	if err != nil {
		pm.ErrorCount++
		pm.LastError = err
		pm.LastErrorAt = time.Now()
		pm.Status = PluginStatusError
	} else {
		pm.Status = PluginStatusRunning
	}
}

// RecordStop records stop metrics
func (pm *PluginMetrics) RecordStop(err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if err != nil {
		pm.ErrorCount++
		pm.LastError = err
		pm.LastErrorAt = time.Now()
		pm.Status = PluginStatusError
	} else {
		pm.Status = PluginStatusStopped
	}
}

// RecordError records an error
func (pm *PluginMetrics) RecordError(err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ErrorCount++
	pm.LastError = err
	pm.LastErrorAt = time.Now()
}

// RecordHookExecution records hook execution metrics
func (pm *PluginMetrics) RecordHookExecution(hookType HookType, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.HookExecutions[hookType]++
	pm.HookDurations[hookType] += duration
}

// RecordEventPublished increments the events published counter
func (pm *PluginMetrics) RecordEventPublished() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.EventsPublished++
}

// RecordEventReceived increments the events received counter
func (pm *PluginMetrics) RecordEventReceived() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.EventsReceived++
}

// RecordServiceCall increments the service calls counter
func (pm *PluginMetrics) RecordServiceCall() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ServiceCalls++
}

// SetStatus sets the plugin status
func (pm *PluginMetrics) SetStatus(status PluginStatus) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.Status = status
}

// GetStatus returns the current plugin status
func (pm *PluginMetrics) GetStatus() PluginStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.Status
}

// GetErrorCount returns the error count
func (pm *PluginMetrics) GetErrorCount() int64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.ErrorCount
}

// GetLastError returns the last error and timestamp
func (pm *PluginMetrics) GetLastError() (error, time.Time) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.LastError, pm.LastErrorAt
}

// GetHookMetrics returns hook execution metrics
func (pm *PluginMetrics) GetHookMetrics() map[HookType]HookMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make(map[HookType]HookMetrics)
	for hookType, count := range pm.HookExecutions {
		totalDuration := pm.HookDurations[hookType]
		avgDuration := time.Duration(0)
		if count > 0 {
			avgDuration = totalDuration / time.Duration(count)
		}

		metrics[hookType] = HookMetrics{
			ExecutionCount:  count,
			TotalDuration:   totalDuration,
			AverageDuration: avgDuration,
			ErrorCount:      0, // Hook-specific errors would need separate tracking
		}
	}

	return metrics
}

// Clone creates a copy of the metrics for safe reading
func (pm *PluginMetrics) Clone() *PluginMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	clone := &PluginMetrics{
		Name:            pm.Name,
		Status:          pm.Status,
		InitDuration:    pm.InitDuration,
		StartDuration:   pm.StartDuration,
		ErrorCount:      pm.ErrorCount,
		LastError:       pm.LastError,
		LastErrorAt:     pm.LastErrorAt,
		EventsPublished: pm.EventsPublished,
		EventsReceived:  pm.EventsReceived,
		ServiceCalls:    pm.ServiceCalls,
		HookExecutions:  make(map[HookType]int64),
		HookDurations:   make(map[HookType]time.Duration),
	}

	for k, v := range pm.HookExecutions {
		clone.HookExecutions[k] = v
	}
	for k, v := range pm.HookDurations {
		clone.HookDurations[k] = v
	}

	return clone
}

// ExportPrometheus exports metrics in Prometheus format
func (pm *PluginMetrics) ExportPrometheus() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var output string

	// Plugin status (1=running, 0=stopped, -1=error)
	statusValue := 0
	switch pm.Status {
	case PluginStatusRunning:
		statusValue = 1
	case PluginStatusError:
		statusValue = -1
	}
	output += fmt.Sprintf("rockstar_plugin_status{plugin=\"%s\"} %d\n", pm.Name, statusValue)

	// Error count
	output += fmt.Sprintf("rockstar_plugin_errors_total{plugin=\"%s\"} %d\n", pm.Name, pm.ErrorCount)

	// Init duration
	output += fmt.Sprintf("rockstar_plugin_init_duration_seconds{plugin=\"%s\"} %f\n", pm.Name, pm.InitDuration.Seconds())

	// Start duration
	output += fmt.Sprintf("rockstar_plugin_start_duration_seconds{plugin=\"%s\"} %f\n", pm.Name, pm.StartDuration.Seconds())

	// Hook executions
	for hookType, count := range pm.HookExecutions {
		output += fmt.Sprintf("rockstar_plugin_hook_executions_total{plugin=\"%s\",hook=\"%s\"} %d\n", pm.Name, hookType, count)
	}

	// Hook durations
	for hookType, duration := range pm.HookDurations {
		output += fmt.Sprintf("rockstar_plugin_hook_duration_seconds{plugin=\"%s\",hook=\"%s\"} %f\n", pm.Name, hookType, duration.Seconds())
	}

	// Events
	output += fmt.Sprintf("rockstar_plugin_events_published_total{plugin=\"%s\"} %d\n", pm.Name, pm.EventsPublished)
	output += fmt.Sprintf("rockstar_plugin_events_received_total{plugin=\"%s\"} %d\n", pm.Name, pm.EventsReceived)

	// Service calls
	output += fmt.Sprintf("rockstar_plugin_service_calls_total{plugin=\"%s\"} %d\n", pm.Name, pm.ServiceCalls)

	return output
}

// PluginMetricsCollector manages metrics for all plugins
type PluginMetricsCollector struct {
	metrics map[string]*PluginMetrics
	mu      sync.RWMutex
}

// NewPluginMetricsCollector creates a new PluginMetricsCollector
func NewPluginMetricsCollector() *PluginMetricsCollector {
	return &PluginMetricsCollector{
		metrics: make(map[string]*PluginMetrics),
	}
}

// GetOrCreate gets or creates metrics for a plugin
func (pmc *PluginMetricsCollector) GetOrCreate(pluginName string) *PluginMetrics {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	if metrics, exists := pmc.metrics[pluginName]; exists {
		return metrics
	}

	metrics := NewPluginMetrics(pluginName)
	pmc.metrics[pluginName] = metrics
	return metrics
}

// Get gets metrics for a plugin
func (pmc *PluginMetricsCollector) Get(pluginName string) *PluginMetrics {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	return pmc.metrics[pluginName]
}

// GetAll returns all plugin metrics
func (pmc *PluginMetricsCollector) GetAll() map[string]*PluginMetrics {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	result := make(map[string]*PluginMetrics)
	for name, metrics := range pmc.metrics {
		result[name] = metrics.Clone()
	}

	return result
}

// Remove removes metrics for a plugin
func (pmc *PluginMetricsCollector) Remove(pluginName string) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	delete(pmc.metrics, pluginName)
}

// ExportPrometheus exports all plugin metrics in Prometheus format
func (pmc *PluginMetricsCollector) ExportPrometheus() string {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	var output string

	// Add metric type definitions
	output += "# HELP rockstar_plugin_status Plugin status (1=running, 0=stopped, -1=error)\n"
	output += "# TYPE rockstar_plugin_status gauge\n"

	output += "# HELP rockstar_plugin_errors_total Total plugin errors\n"
	output += "# TYPE rockstar_plugin_errors_total counter\n"

	output += "# HELP rockstar_plugin_init_duration_seconds Plugin initialization duration\n"
	output += "# TYPE rockstar_plugin_init_duration_seconds gauge\n"

	output += "# HELP rockstar_plugin_start_duration_seconds Plugin start duration\n"
	output += "# TYPE rockstar_plugin_start_duration_seconds gauge\n"

	output += "# HELP rockstar_plugin_hook_executions_total Total plugin hook executions\n"
	output += "# TYPE rockstar_plugin_hook_executions_total counter\n"

	output += "# HELP rockstar_plugin_hook_duration_seconds Total plugin hook duration\n"
	output += "# TYPE rockstar_plugin_hook_duration_seconds counter\n"

	output += "# HELP rockstar_plugin_events_published_total Total events published by plugin\n"
	output += "# TYPE rockstar_plugin_events_published_total counter\n"

	output += "# HELP rockstar_plugin_events_received_total Total events received by plugin\n"
	output += "# TYPE rockstar_plugin_events_received_total counter\n"

	output += "# HELP rockstar_plugin_service_calls_total Total service calls made by plugin\n"
	output += "# TYPE rockstar_plugin_service_calls_total counter\n"

	// Export metrics for each plugin
	for _, metrics := range pmc.metrics {
		output += metrics.ExportPrometheus()
	}

	return output
}
