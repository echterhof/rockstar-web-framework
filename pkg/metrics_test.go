package pkg

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// mockDatabaseForMetrics is a mock database for testing metrics
type mockDatabaseForMetrics struct {
	savedMetrics    []*WorkloadMetrics
	shouldFailSave  bool
	shouldFailQuery bool
}

func (m *mockDatabaseForMetrics) SaveWorkloadMetrics(metrics *WorkloadMetrics) error {
	if m.shouldFailSave {
		return errors.New("database save failed")
	}
	m.savedMetrics = append(m.savedMetrics, metrics)
	return nil
}

func (m *mockDatabaseForMetrics) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	if m.shouldFailQuery {
		return nil, errors.New("database query failed")
	}

	var result []*WorkloadMetrics
	for _, metric := range m.savedMetrics {
		if metric.TenantID == tenantID &&
			!metric.Timestamp.Before(from) &&
			!metric.Timestamp.After(to) {
			result = append(result, metric)
		}
	}
	return result, nil
}

// Implement other DatabaseManager methods as no-ops
func (m *mockDatabaseForMetrics) Connect(config DatabaseConfig) error { return nil }
func (m *mockDatabaseForMetrics) Close() error                        { return nil }
func (m *mockDatabaseForMetrics) Ping() error                         { return nil }
func (m *mockDatabaseForMetrics) Stats() DatabaseStats                { return DatabaseStats{} }
func (m *mockDatabaseForMetrics) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *mockDatabaseForMetrics) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) Prepare(query string) (*sql.Stmt, error) { return nil, nil }
func (m *mockDatabaseForMetrics) Begin() (Transaction, error)             { return nil, nil }
func (m *mockDatabaseForMetrics) BeginTx(opts *sql.TxOptions) (Transaction, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) SaveSession(session *Session) error { return nil }
func (m *mockDatabaseForMetrics) LoadSession(sessionID string) (*Session, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) DeleteSession(sessionID string) error { return nil }
func (m *mockDatabaseForMetrics) CleanupExpiredSessions() error        { return nil }
func (m *mockDatabaseForMetrics) SaveAccessToken(token *AccessToken) error {
	return nil
}
func (m *mockDatabaseForMetrics) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) DeleteAccessToken(tokenValue string) error { return nil }
func (m *mockDatabaseForMetrics) CleanupExpiredTokens() error               { return nil }
func (m *mockDatabaseForMetrics) SaveTenant(tenant *Tenant) error           { return nil }
func (m *mockDatabaseForMetrics) LoadTenant(tenantID string) (*Tenant, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) LoadTenantByHost(hostname string) (*Tenant, error) {
	return nil, nil
}
func (m *mockDatabaseForMetrics) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
func (m *mockDatabaseForMetrics) IncrementRateLimit(key string, window time.Duration) error {
	return nil
}
func (m *mockDatabaseForMetrics) Migrate() error                { return nil }
func (m *mockDatabaseForMetrics) CreateTables() error           { return nil }
func (m *mockDatabaseForMetrics) DropTables() error             { return nil }
func (m *mockDatabaseForMetrics) InitializePluginTables() error { return nil }

// TestNewMetricsCollector tests creating a new metrics collector
func TestNewMetricsCollector(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db)

	if collector == nil {
		t.Fatal("Expected non-nil metrics collector")
	}
}

// TestMetricsCollectorStart tests starting metrics collection
func TestMetricsCollectorStart(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	requestID := "req-123"
	metrics := collector.Start(requestID)

	if metrics == nil {
		t.Fatal("Expected non-nil request metrics")
	}

	if metrics.RequestID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, metrics.RequestID)
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}
}

// TestRequestMetricsEnd tests ending metrics collection
func TestRequestMetricsEnd(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	metrics.End()

	if metrics.EndTime.IsZero() {
		t.Error("Expected non-zero end time")
	}

	if metrics.Duration == 0 {
		t.Error("Expected non-zero duration")
	}

	if metrics.Duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", metrics.Duration)
	}
}

// TestRequestMetricsSetContext tests setting context information
func TestRequestMetricsSetContext(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.SetContext("tenant-1", "user-1", "/api/users", "GET")

	if metrics.TenantID != "tenant-1" {
		t.Errorf("Expected tenant ID tenant-1, got %s", metrics.TenantID)
	}

	if metrics.UserID != "user-1" {
		t.Errorf("Expected user ID user-1, got %s", metrics.UserID)
	}

	if metrics.Path != "/api/users" {
		t.Errorf("Expected path /api/users, got %s", metrics.Path)
	}

	if metrics.Method != "GET" {
		t.Errorf("Expected method GET, got %s", metrics.Method)
	}
}

// TestRequestMetricsSetResponse tests setting response information
func TestRequestMetricsSetResponse(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.SetResponse(200, 1024, nil)

	if metrics.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", metrics.StatusCode)
	}

	if metrics.ResponseSize != 1024 {
		t.Errorf("Expected response size 1024, got %d", metrics.ResponseSize)
	}

	if metrics.ErrorMessage != "" {
		t.Errorf("Expected empty error message, got %s", metrics.ErrorMessage)
	}
}

// TestRequestMetricsSetResponseWithError tests setting response with error
func TestRequestMetricsSetResponseWithError(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	err := errors.New("test error")
	metrics.SetResponse(500, 0, err)

	if metrics.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", metrics.StatusCode)
	}

	if metrics.ErrorMessage != "test error" {
		t.Errorf("Expected error message 'test error', got %s", metrics.ErrorMessage)
	}
}

// TestRequestMetricsSetContextSize tests setting context size
func TestRequestMetricsSetContextSize(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.SetContextSize(2048)

	if metrics.ContextSize != 2048 {
		t.Errorf("Expected context size 2048, got %d", metrics.ContextSize)
	}
}

// TestMetricsCollectorRecord tests recording metrics to database
func TestMetricsCollectorRecord(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.SetContext("tenant-1", "user-1", "/api/users", "GET")
	metrics.SetResponse(200, 1024, nil)
	metrics.SetContextSize(2048)
	time.Sleep(10 * time.Millisecond)
	metrics.End()

	err := collector.Record(metrics)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(db.savedMetrics) != 1 {
		t.Fatalf("Expected 1 saved metric, got %d", len(db.savedMetrics))
	}

	saved := db.savedMetrics[0]
	if saved.RequestID != "req-123" {
		t.Errorf("Expected request ID req-123, got %s", saved.RequestID)
	}

	if saved.TenantID != "tenant-1" {
		t.Errorf("Expected tenant ID tenant-1, got %s", saved.TenantID)
	}

	if saved.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", saved.StatusCode)
	}
}

// TestMetricsCollectorRecordWithoutDatabase tests recording without database
func TestMetricsCollectorRecordWithoutDatabase(t *testing.T) {
	collector := NewMetricsCollector(nil).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.End()

	// Should not error when database is nil
	err := collector.Record(metrics)
	if err != nil {
		t.Errorf("Expected no error with nil database, got %v", err)
	}
}

// TestMetricsCollectorRecordDatabaseError tests handling database errors
func TestMetricsCollectorRecordDatabaseError(t *testing.T) {
	db := &mockDatabaseForMetrics{shouldFailSave: true}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")
	metrics.End()

	err := collector.Record(metrics)
	if err == nil {
		t.Error("Expected error when database save fails")
	}
}

// TestMetricsCollectorGetMetrics tests retrieving metrics
func TestMetricsCollectorGetMetrics(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	// Record some metrics
	now := time.Now()
	for i := 0; i < 5; i++ {
		metrics := collector.Start("req-" + string(rune('0'+i)))
		metrics.SetContext("tenant-1", "user-1", "/api/test", "GET")
		metrics.End()
		collector.Record(metrics)
	}

	// Retrieve metrics
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)
	retrieved, err := collector.GetMetrics("tenant-1", from, to)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(retrieved) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(retrieved))
	}
}

// TestMetricsCollectorGetMetricsFiltering tests metrics filtering by tenant and time
func TestMetricsCollectorGetMetricsFiltering(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	now := time.Now()

	// Record metrics for different tenants
	m1 := collector.Start("req-1")
	m1.SetContext("tenant-1", "user-1", "/api/test", "GET")
	m1.End()
	collector.Record(m1)

	m2 := collector.Start("req-2")
	m2.SetContext("tenant-2", "user-2", "/api/test", "GET")
	m2.End()
	collector.Record(m2)

	// Retrieve metrics for tenant-1 only
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)
	retrieved, err := collector.GetMetrics("tenant-1", from, to)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 metric for tenant-1, got %d", len(retrieved))
	}

	if len(retrieved) > 0 && retrieved[0].TenantID != "tenant-1" {
		t.Errorf("Expected tenant-1, got %s", retrieved[0].TenantID)
	}
}

// TestGetAggregatedMetrics tests aggregated metrics calculation
func TestGetAggregatedMetrics(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	now := time.Now()

	// Record multiple metrics with different characteristics
	for i := 0; i < 10; i++ {
		metrics := collector.Start("req-" + string(rune('0'+i)))
		metrics.SetContext("tenant-1", "user-1", "/api/test", "GET")

		// Vary status codes
		statusCode := 200
		if i%3 == 0 {
			statusCode = 500
		}
		metrics.SetResponse(statusCode, int64(1000+i*100), nil)

		time.Sleep(1 * time.Millisecond)
		metrics.End()
		collector.Record(metrics)
	}

	// Get aggregated metrics
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)
	agg, err := collector.GetAggregatedMetrics("tenant-1", from, to)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if agg.TotalRequests != 10 {
		t.Errorf("Expected 10 total requests, got %d", agg.TotalRequests)
	}

	// 4 requests should have status 500 (i=0,3,6,9)
	if agg.FailedRequests != 4 {
		t.Errorf("Expected 4 failed requests, got %d", agg.FailedRequests)
	}

	// 6 requests should have status 200
	if agg.SuccessfulRequests != 6 {
		t.Errorf("Expected 6 successful requests, got %d", agg.SuccessfulRequests)
	}

	if agg.AvgDuration == 0 {
		t.Error("Expected non-zero average duration")
	}

	if agg.TotalResponseSize == 0 {
		t.Error("Expected non-zero total response size")
	}
}

// TestGetAggregatedMetricsEmpty tests aggregated metrics with no data
func TestGetAggregatedMetricsEmpty(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	from := time.Now().Add(-1 * time.Hour)
	to := time.Now().Add(1 * time.Hour)

	agg, err := collector.GetAggregatedMetrics("tenant-1", from, to)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if agg.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", agg.TotalRequests)
	}
}

// TestPredictLoad tests load prediction
func TestPredictLoad(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	// Record historical metrics
	for i := 0; i < 100; i++ {
		metrics := collector.Start("req-" + string(rune('0'+i)))
		metrics.SetContext("tenant-1", "user-1", "/api/test", "GET")
		metrics.SetResponse(200, 1000, nil)
		metrics.End()
		collector.Record(metrics)
	}

	// Predict load for next hour
	prediction, err := collector.PredictLoad("tenant-1", 1*time.Hour)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if prediction.TenantID != "tenant-1" {
		t.Errorf("Expected tenant-1, got %s", prediction.TenantID)
	}

	if prediction.PredictedRequests == 0 {
		t.Error("Expected non-zero predicted requests")
	}

	if prediction.BasedOnDataPoints != 100 {
		t.Errorf("Expected 100 data points, got %d", prediction.BasedOnDataPoints)
	}

	// With 100 data points, confidence should be at least 0.60
	if prediction.ConfidenceLevel < 0.60 {
		t.Errorf("Expected confidence level >= 0.60, got %f", prediction.ConfidenceLevel)
	}
}

// TestPredictLoadLowConfidence tests load prediction with few data points
func TestPredictLoadLowConfidence(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	// Record only 5 metrics
	for i := 0; i < 5; i++ {
		metrics := collector.Start("req-" + string(rune('0'+i)))
		metrics.SetContext("tenant-1", "user-1", "/api/test", "GET")
		metrics.End()
		collector.Record(metrics)
	}

	prediction, err := collector.PredictLoad("tenant-1", 1*time.Hour)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// With only 5 data points, confidence should be 0.30
	if prediction.ConfidenceLevel != 0.30 {
		t.Errorf("Expected confidence level 0.30, got %f", prediction.ConfidenceLevel)
	}
}

// TestGetCurrentMemoryUsage tests getting current memory usage
func TestGetCurrentMemoryUsage(t *testing.T) {
	usage := GetCurrentMemoryUsage()

	if usage <= 0 {
		t.Error("Expected positive memory usage")
	}
}

// TestGetCurrentCPUUsage tests getting current CPU usage
func TestGetCurrentCPUUsage(t *testing.T) {
	usage := GetCurrentCPUUsage()

	if usage < 0 {
		t.Error("Expected non-negative CPU usage")
	}
}

// TestMemoryTracking tests that memory usage is tracked correctly
func TestMemoryTracking(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	metrics := collector.Start("req-123")

	// Allocate some memory
	_ = make([]byte, 1024*1024) // 1MB

	metrics.End()

	// Memory usage should be tracked (though exact value depends on GC)
	// We just verify it's non-negative
	if metrics.MemoryUsage < 0 {
		t.Errorf("Expected non-negative memory usage, got %d", metrics.MemoryUsage)
	}
}

// TestConcurrentMetricsCollection tests concurrent metrics collection
func TestConcurrentMetricsCollection(t *testing.T) {
	db := &mockDatabaseForMetrics{}
	collector := NewMetricsCollector(db).(*metricsCollectorImpl)

	done := make(chan bool)

	// Start multiple concurrent requests
	for i := 0; i < 10; i++ {
		go func(id int) {
			metrics := collector.Start("req-" + string(rune('0'+id)))
			metrics.SetContext("tenant-1", "user-1", "/api/test", "GET")
			time.Sleep(1 * time.Millisecond)
			metrics.End()
			collector.Record(metrics)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if len(db.savedMetrics) != 10 {
		t.Errorf("Expected 10 saved metrics, got %d", len(db.savedMetrics))
	}
}
