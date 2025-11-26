package pkg

import (
	"errors"
	"sync"
	"time"
)

// inMemoryMetricsStorage implements metrics storage in memory
type inMemoryMetricsStorage struct {
	mu      sync.RWMutex
	metrics []*WorkloadMetrics
	nextID  int64
}

// newInMemoryMetricsStorage creates a new in-memory metrics storage instance
func newInMemoryMetricsStorage() *inMemoryMetricsStorage {
	return &inMemoryMetricsStorage{
		metrics: make([]*WorkloadMetrics, 0),
		nextID:  1,
	}
}

// Save saves a workload metric to memory
func (s *inMemoryMetricsStorage) Save(metric *WorkloadMetrics) error {
	if metric == nil {
		return errors.New("metric is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy of the metric to avoid external modifications
	metricCopy := &WorkloadMetrics{
		ID:           s.nextID,
		Timestamp:    metric.Timestamp,
		TenantID:     metric.TenantID,
		UserID:       metric.UserID,
		RequestID:    metric.RequestID,
		Duration:     metric.Duration,
		ContextSize:  metric.ContextSize,
		MemoryUsage:  metric.MemoryUsage,
		CPUUsage:     metric.CPUUsage,
		Path:         metric.Path,
		Method:       metric.Method,
		StatusCode:   metric.StatusCode,
		ResponseSize: metric.ResponseSize,
		ErrorMessage: metric.ErrorMessage,
	}

	s.nextID++
	s.metrics = append(s.metrics, metricCopy)
	return nil
}

// Query retrieves metrics from memory based on tenant ID and time range
func (s *inMemoryMetricsStorage) Query(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*WorkloadMetrics, 0)

	for _, metric := range s.metrics {
		// Filter by tenant ID (empty string means all tenants)
		if tenantID != "" && metric.TenantID != tenantID {
			continue
		}

		// Filter by time range
		if metric.Timestamp.Before(from) || metric.Timestamp.After(to) {
			continue
		}

		// Create a copy to avoid external modifications
		metricCopy := &WorkloadMetrics{
			ID:           metric.ID,
			Timestamp:    metric.Timestamp,
			TenantID:     metric.TenantID,
			UserID:       metric.UserID,
			RequestID:    metric.RequestID,
			Duration:     metric.Duration,
			ContextSize:  metric.ContextSize,
			MemoryUsage:  metric.MemoryUsage,
			CPUUsage:     metric.CPUUsage,
			Path:         metric.Path,
			Method:       metric.Method,
			StatusCode:   metric.StatusCode,
			ResponseSize: metric.ResponseSize,
			ErrorMessage: metric.ErrorMessage,
		}

		result = append(result, metricCopy)
	}

	return result, nil
}

// Count returns the number of metrics in memory (useful for testing)
func (s *inMemoryMetricsStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.metrics)
}

// Clear removes all metrics from memory (useful for testing)
func (s *inMemoryMetricsStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = make([]*WorkloadMetrics, 0)
	s.nextID = 1
	return nil
}
