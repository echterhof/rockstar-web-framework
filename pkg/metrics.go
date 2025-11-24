package pkg

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// RequestMetrics tracks metrics for a single request
type RequestMetrics struct {
	RequestID    string
	TenantID     string
	UserID       string
	Path         string
	Method       string
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	ContextSize  int64
	MemoryUsage  int64
	CPUUsage     float64
	StatusCode   int
	ResponseSize int64
	ErrorMessage string

	// Internal tracking
	startMemStats runtime.MemStats
	endMemStats   runtime.MemStats
	mu            sync.Mutex
}

// AggregatedMetrics provides aggregated metrics for analysis
type AggregatedMetrics struct {
	TenantID           string        `json:"tenant_id"`
	From               time.Time     `json:"from"`
	To                 time.Time     `json:"to"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AvgDuration        time.Duration `json:"avg_duration_ms"`
	MinDuration        time.Duration `json:"min_duration_ms"`
	MaxDuration        time.Duration `json:"max_duration_ms"`
	AvgMemoryUsage     int64         `json:"avg_memory_usage"`
	AvgCPUUsage        float64       `json:"avg_cpu_usage"`
	TotalResponseSize  int64         `json:"total_response_size"`
	RequestsPerSecond  float64       `json:"requests_per_second"`
}

// LoadPrediction provides predicted load metrics
type LoadPrediction struct {
	TenantID          string    `json:"tenant_id"`
	PredictionTime    time.Time `json:"prediction_time"`
	PredictedRequests int64     `json:"predicted_requests"`
	PredictedMemory   int64     `json:"predicted_memory"`
	PredictedCPU      float64   `json:"predicted_cpu"`
	ConfidenceLevel   float64   `json:"confidence_level"`
	BasedOnDataPoints int       `json:"based_on_data_points"`
}

// metricsCollectorImpl implements MetricsCollector
type metricsCollectorImpl struct {
	db       DatabaseManager
	counters map[string]int64
	gauges   map[string]float64
	mu       sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db DatabaseManager) MetricsCollector {
	return &metricsCollectorImpl{
		db:       db,
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

// Start begins metrics collection for a request
func (m *metricsCollectorImpl) Start(requestID string) *RequestMetrics {
	metrics := &RequestMetrics{
		RequestID: requestID,
		StartTime: time.Now(),
	}

	// Capture initial memory stats
	runtime.ReadMemStats(&metrics.startMemStats)

	return metrics
}

// End finalizes metrics collection for a request
func (rm *RequestMetrics) End() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.EndTime = time.Now()
	rm.Duration = rm.EndTime.Sub(rm.StartTime)

	// Capture final memory stats
	runtime.ReadMemStats(&rm.endMemStats)

	// Calculate memory usage (allocated during request)
	rm.MemoryUsage = int64(rm.endMemStats.Alloc - rm.startMemStats.Alloc)

	// Calculate CPU usage (approximate based on duration and goroutines)
	// This is a simplified calculation - in production, you'd use more sophisticated methods
	numCPU := float64(runtime.NumCPU())
	if rm.Duration.Seconds() > 0 {
		// Estimate CPU usage based on time spent
		rm.CPUUsage = (float64(rm.Duration.Microseconds()) / 1000000.0) * 100.0 / numCPU
		if rm.CPUUsage > 100.0 {
			rm.CPUUsage = 100.0
		}
	}
}

// SetContext sets context-related metrics
func (rm *RequestMetrics) SetContext(tenantID, userID, path, method string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.TenantID = tenantID
	rm.UserID = userID
	rm.Path = path
	rm.Method = method
}

// SetResponse sets response-related metrics
func (rm *RequestMetrics) SetResponse(statusCode int, responseSize int64, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.StatusCode = statusCode
	rm.ResponseSize = responseSize
	if err != nil {
		rm.ErrorMessage = err.Error()
	}
}

// SetContextSize sets the context size in bytes
func (rm *RequestMetrics) SetContextSize(size int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.ContextSize = size
}

// Record saves the collected metrics to the database
func (m *metricsCollectorImpl) Record(metrics *RequestMetrics) error {
	if m.db == nil {
		// If no database is configured, silently skip recording
		return nil
	}

	// Ensure metrics are finalized
	if metrics.EndTime.IsZero() {
		metrics.End()
	}

	workloadMetrics := &WorkloadMetrics{
		Timestamp:    metrics.StartTime,
		TenantID:     metrics.TenantID,
		UserID:       metrics.UserID,
		RequestID:    metrics.RequestID,
		Duration:     metrics.Duration.Milliseconds(),
		ContextSize:  metrics.ContextSize,
		MemoryUsage:  metrics.MemoryUsage,
		CPUUsage:     metrics.CPUUsage,
		Path:         metrics.Path,
		Method:       metrics.Method,
		StatusCode:   metrics.StatusCode,
		ResponseSize: metrics.ResponseSize,
		ErrorMessage: metrics.ErrorMessage,
	}

	return m.db.SaveWorkloadMetrics(workloadMetrics)
}

// GetMetrics retrieves metrics for a tenant within a time range
func (m *metricsCollectorImpl) GetMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	if m.db == nil {
		return nil, nil
	}

	return m.db.GetWorkloadMetrics(tenantID, from, to)
}

// GetAggregatedMetrics returns aggregated metrics for analysis
func (m *metricsCollectorImpl) GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error) {
	metrics, err := m.GetMetrics(tenantID, from, to)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return &AggregatedMetrics{
			TenantID: tenantID,
			From:     from,
			To:       to,
		}, nil
	}

	agg := &AggregatedMetrics{
		TenantID:      tenantID,
		From:          from,
		To:            to,
		TotalRequests: int64(len(metrics)),
		MinDuration:   time.Duration(metrics[0].Duration) * time.Millisecond,
		MaxDuration:   time.Duration(metrics[0].Duration) * time.Millisecond,
	}

	var totalDuration int64
	var totalMemory int64
	var totalCPU float64
	var totalResponseSize int64

	for _, m := range metrics {
		duration := time.Duration(m.Duration) * time.Millisecond

		// Count successes and failures
		if m.StatusCode >= 200 && m.StatusCode < 400 {
			agg.SuccessfulRequests++
		} else {
			agg.FailedRequests++
		}

		// Accumulate totals
		totalDuration += m.Duration
		totalMemory += m.MemoryUsage
		totalCPU += m.CPUUsage
		totalResponseSize += m.ResponseSize

		// Track min/max duration
		if duration < agg.MinDuration {
			agg.MinDuration = duration
		}
		if duration > agg.MaxDuration {
			agg.MaxDuration = duration
		}
	}

	// Calculate averages
	count := int64(len(metrics))
	agg.AvgDuration = time.Duration(totalDuration/count) * time.Millisecond
	agg.AvgMemoryUsage = totalMemory / count
	agg.AvgCPUUsage = totalCPU / float64(count)
	agg.TotalResponseSize = totalResponseSize

	// Calculate requests per second
	timeRange := to.Sub(from).Seconds()
	if timeRange > 0 {
		agg.RequestsPerSecond = float64(agg.TotalRequests) / timeRange
	}

	return agg, nil
}

// PredictLoad predicts future load based on historical data
func (m *metricsCollectorImpl) PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error) {
	// Get historical data from the same time period in the past
	now := time.Now()
	from := now.Add(-duration)

	agg, err := m.GetAggregatedMetrics(tenantID, from, now)
	if err != nil {
		return nil, err
	}

	// Simple prediction based on historical averages
	// In production, you'd use more sophisticated algorithms (moving averages, ML models, etc.)
	prediction := &LoadPrediction{
		TenantID:          tenantID,
		PredictionTime:    now.Add(duration),
		PredictedRequests: agg.TotalRequests,
		PredictedMemory:   agg.AvgMemoryUsage * agg.TotalRequests,
		PredictedCPU:      agg.AvgCPUUsage,
		BasedOnDataPoints: int(agg.TotalRequests),
	}

	// Calculate confidence level based on data points
	// More data points = higher confidence
	if prediction.BasedOnDataPoints > 1000 {
		prediction.ConfidenceLevel = 0.95
	} else if prediction.BasedOnDataPoints > 100 {
		prediction.ConfidenceLevel = 0.80
	} else if prediction.BasedOnDataPoints > 10 {
		prediction.ConfidenceLevel = 0.60
	} else {
		prediction.ConfidenceLevel = 0.30
	}

	return prediction, nil
}

// GetCurrentMemoryUsage returns current memory usage in bytes
func GetCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// GetCurrentCPUUsage returns approximate current CPU usage percentage
func GetCurrentCPUUsage() float64 {
	// This is a simplified implementation
	// In production, you'd use more sophisticated CPU monitoring
	return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10.0
}

// IncrementCounter increments a counter metric
func (m *metricsCollectorImpl) IncrementCounter(name string, tags map[string]string) error {
	return m.IncrementCounterBy(name, 1, tags)
}

// IncrementCounterBy increments a counter metric by a specific value
func (m *metricsCollectorImpl) IncrementCounterBy(name string, value int64, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, tags)
	m.counters[key] += value
	return nil
}

// SetGauge sets a gauge metric value
func (m *metricsCollectorImpl) SetGauge(name string, value float64, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, tags)
	m.gauges[key] = value
	return nil
}

// IncrementGauge increments a gauge metric
func (m *metricsCollectorImpl) IncrementGauge(name string, value float64, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, tags)
	m.gauges[key] += value
	return nil
}

// DecrementGauge decrements a gauge metric
func (m *metricsCollectorImpl) DecrementGauge(name string, value float64, tags map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, tags)
	m.gauges[key] -= value
	return nil
}

// RecordHistogram records a histogram metric
func (m *metricsCollectorImpl) RecordHistogram(name string, value float64, tags map[string]string) error {
	// For simplicity, we'll treat histograms as gauges
	return m.SetGauge(name, value, tags)
}

// RecordTiming records a timing metric
func (m *metricsCollectorImpl) RecordTiming(name string, duration time.Duration, tags map[string]string) error {
	return m.RecordHistogram(name, float64(duration.Milliseconds()), tags)
}

// StartTimer starts a timer for a metric
func (m *metricsCollectorImpl) StartTimer(name string, tags map[string]string) Timer {
	return &timerImpl{
		name:      name,
		tags:      tags,
		startTime: time.Now(),
		collector: m,
	}
}

// RecordRequest records request metrics
func (m *metricsCollectorImpl) RecordRequest(ctx Context, duration time.Duration, statusCode int) error {
	tags := map[string]string{
		"method": ctx.Request().Method,
		"path":   ctx.Request().RequestURI,
		"status": fmt.Sprintf("%d", statusCode),
	}

	if ctx.Tenant() != nil {
		tags["tenant"] = ctx.Tenant().ID
	}

	m.IncrementCounter("http.requests", tags)
	m.RecordTiming("http.request.duration", duration, tags)

	return nil
}

// RecordError records error metrics
func (m *metricsCollectorImpl) RecordError(ctx Context, err error) error {
	tags := map[string]string{
		"error": err.Error(),
	}

	if ctx.Tenant() != nil {
		tags["tenant"] = ctx.Tenant().ID
	}

	return m.IncrementCounter("http.errors", tags)
}

// RecordWorkloadMetrics records workload metrics to the database
func (m *metricsCollectorImpl) RecordWorkloadMetrics(metrics *WorkloadMetrics) error {
	if m.db == nil {
		return nil
	}
	return m.db.SaveWorkloadMetrics(metrics)
}

// GetWorkloadMetrics retrieves workload metrics from the database
func (m *metricsCollectorImpl) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	if m.db == nil {
		return nil, nil
	}
	return m.db.GetWorkloadMetrics(tenantID, from, to)
}

// RecordMemoryUsage records memory usage metrics
func (m *metricsCollectorImpl) RecordMemoryUsage(usage int64) error {
	return m.SetGauge("system.memory.usage", float64(usage), nil)
}

// RecordCPUUsage records CPU usage metrics
func (m *metricsCollectorImpl) RecordCPUUsage(usage float64) error {
	return m.SetGauge("system.cpu.usage", usage, nil)
}

// RecordCustomMetric records a custom metric
func (m *metricsCollectorImpl) RecordCustomMetric(name string, value interface{}, tags map[string]string) error {
	switch v := value.(type) {
	case int:
		return m.IncrementCounterBy(name, int64(v), tags)
	case int64:
		return m.IncrementCounterBy(name, v, tags)
	case float64:
		return m.SetGauge(name, v, tags)
	case float32:
		return m.SetGauge(name, float64(v), tags)
	default:
		return fmt.Errorf("unsupported metric value type: %T", value)
	}
}

// Export exports all metrics
func (m *metricsCollectorImpl) Export() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})

	counters := make(map[string]int64)
	for k, v := range m.counters {
		counters[k] = v
	}
	result["counters"] = counters

	gauges := make(map[string]float64)
	for k, v := range m.gauges {
		gauges[k] = v
	}
	result["gauges"] = gauges

	return result, nil
}

// ExportPrometheus exports metrics in Prometheus format
func (m *metricsCollectorImpl) ExportPrometheus() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var output string

	// Export counters
	for name, value := range m.counters {
		output += fmt.Sprintf("# TYPE %s counter\n", name)
		output += fmt.Sprintf("%s %d\n", name, value)
	}

	// Export gauges
	for name, value := range m.gauges {
		output += fmt.Sprintf("# TYPE %s gauge\n", name)
		output += fmt.Sprintf("%s %f\n", name, value)
	}

	return []byte(output), nil
}

// buildKey builds a metric key from name and tags
func (m *metricsCollectorImpl) buildKey(name string, tags map[string]string) string {
	if len(tags) == 0 {
		return name
	}

	key := name
	for k, v := range tags {
		key += fmt.Sprintf(",%s=%s", k, v)
	}
	return key
}

// timerImpl implements Timer
type timerImpl struct {
	name      string
	tags      map[string]string
	startTime time.Time
	collector *metricsCollectorImpl
}

// Stop stops the timer and records the duration
func (t *timerImpl) Stop() time.Duration {
	duration := time.Since(t.startTime)
	t.collector.RecordTiming(t.name, duration, t.tags)
	return duration
}

// Elapsed returns the elapsed time without stopping the timer
func (t *timerImpl) Elapsed() time.Duration {
	return time.Since(t.startTime)
}
