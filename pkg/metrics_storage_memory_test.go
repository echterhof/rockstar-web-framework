package pkg

import (
	"testing"
	"time"
)

func TestInMemoryMetricsStorage_Save(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	metric := &WorkloadMetrics{
		Timestamp:    time.Now(),
		TenantID:     "tenant1",
		UserID:       "user1",
		RequestID:    "req1",
		Duration:     100,
		ContextSize:  1024,
		MemoryUsage:  2048,
		CPUUsage:     50.5,
		Path:         "/api/test",
		Method:       "GET",
		StatusCode:   200,
		ResponseSize: 512,
		ErrorMessage: "",
	}

	err := storage.Save(metric)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("Expected 1 metric, got %d", storage.Count())
	}
}

func TestInMemoryMetricsStorage_SaveNil(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	err := storage.Save(nil)
	if err == nil {
		t.Error("Expected error when saving nil metric")
	}
}

func TestInMemoryMetricsStorage_Query(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	now := time.Now()
	metrics := []*WorkloadMetrics{
		{
			Timestamp: now.Add(-2 * time.Hour),
			TenantID:  "tenant1",
			RequestID: "req1",
			Duration:  100,
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			TenantID:  "tenant1",
			RequestID: "req2",
			Duration:  200,
		},
		{
			Timestamp: now.Add(-30 * time.Minute),
			TenantID:  "tenant2",
			RequestID: "req3",
			Duration:  150,
		},
	}

	for _, m := range metrics {
		if err := storage.Save(m); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Query all metrics for tenant1
	from := now.Add(-3 * time.Hour)
	to := now
	results, err := storage.Query("tenant1", from, to)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 metrics for tenant1, got %d", len(results))
	}

	// Query all metrics (empty tenant ID)
	results, err = storage.Query("", from, to)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 metrics for all tenants, got %d", len(results))
	}
}

func TestInMemoryMetricsStorage_QueryTimeRange(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	now := time.Now()
	metrics := []*WorkloadMetrics{
		{
			Timestamp: now.Add(-2 * time.Hour),
			TenantID:  "tenant1",
			RequestID: "req1",
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			TenantID:  "tenant1",
			RequestID: "req2",
		},
		{
			Timestamp: now.Add(-30 * time.Minute),
			TenantID:  "tenant1",
			RequestID: "req3",
		},
	}

	for _, m := range metrics {
		if err := storage.Save(m); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Query only metrics from the last hour
	from := now.Add(-1*time.Hour - 1*time.Minute)
	to := now
	results, err := storage.Query("tenant1", from, to)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 metrics in time range, got %d", len(results))
	}
}

func TestInMemoryMetricsStorage_QueryEmptyResults(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now

	results, err := storage.Query("tenant1", from, to)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(results))
	}
}

func TestInMemoryMetricsStorage_Clear(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	// Add some metrics
	for i := 0; i < 5; i++ {
		metric := &WorkloadMetrics{
			Timestamp: time.Now(),
			TenantID:  "tenant1",
			RequestID: "req",
		}
		if err := storage.Save(metric); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	if storage.Count() != 5 {
		t.Errorf("Expected 5 metrics, got %d", storage.Count())
	}

	// Clear storage
	err := storage.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if storage.Count() != 0 {
		t.Errorf("Expected 0 metrics after clear, got %d", storage.Count())
	}
}

func TestInMemoryMetricsStorage_ConcurrentAccess(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			metric := &WorkloadMetrics{
				Timestamp: time.Now(),
				TenantID:  "tenant1",
				RequestID: "req",
			}
			storage.Save(metric)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if storage.Count() != 10 {
		t.Errorf("Expected 10 metrics after concurrent writes, got %d", storage.Count())
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			now := time.Now()
			storage.Query("tenant1", now.Add(-1*time.Hour), now)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestInMemoryMetricsStorage_IsolationBetweenQueries(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	metric := &WorkloadMetrics{
		Timestamp: time.Now(),
		TenantID:  "tenant1",
		RequestID: "req1",
		Duration:  100,
	}

	if err := storage.Save(metric); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Query and modify the result
	now := time.Now()
	results, err := storage.Query("tenant1", now.Add(-1*time.Hour), now)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Modify the returned metric
	results[0].Duration = 999

	// Query again and verify the original value is unchanged
	results2, err := storage.Query("tenant1", now.Add(-1*time.Hour), now)
	if err != nil {
		t.Fatalf("Second query failed: %v", err)
	}

	if len(results2) != 1 {
		t.Fatalf("Expected 1 result in second query, got %d", len(results2))
	}

	if results2[0].Duration != 100 {
		t.Errorf("Expected duration 100, got %d (modification leaked)", results2[0].Duration)
	}
}

func TestInMemoryMetricsStorage_AutoIncrementID(t *testing.T) {
	storage := newInMemoryMetricsStorage()

	// Save multiple metrics
	for i := 0; i < 3; i++ {
		metric := &WorkloadMetrics{
			Timestamp: time.Now(),
			TenantID:  "tenant1",
			RequestID: "req",
		}
		if err := storage.Save(metric); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Query all metrics
	now := time.Now()
	results, err := storage.Query("", now.Add(-1*time.Hour), now)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify IDs are auto-incremented
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		expectedID := int64(i + 1)
		if result.ID != expectedID {
			t.Errorf("Expected ID %d, got %d", expectedID, result.ID)
		}
	}
}
