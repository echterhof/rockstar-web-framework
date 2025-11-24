package pkg

import (
	"encoding/json"
	"testing"
	"time"
)

// TestSessionModel tests the Session model structure and JSON marshaling
func TestSessionModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	session := &Session{
		ID:        "session-123",
		UserID:    "user-456",
		TenantID:  "tenant-789",
		Data:      map[string]interface{}{"key": "value", "count": 42},
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Session
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal session: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, unmarshaled.ID)
	}
	if unmarshaled.UserID != session.UserID {
		t.Errorf("Expected UserID %s, got %s", session.UserID, unmarshaled.UserID)
	}
	if unmarshaled.TenantID != session.TenantID {
		t.Errorf("Expected TenantID %s, got %s", session.TenantID, unmarshaled.TenantID)
	}
	if unmarshaled.IPAddress != session.IPAddress {
		t.Errorf("Expected IPAddress %s, got %s", session.IPAddress, unmarshaled.IPAddress)
	}
	if unmarshaled.UserAgent != session.UserAgent {
		t.Errorf("Expected UserAgent %s, got %s", session.UserAgent, unmarshaled.UserAgent)
	}

	// Verify nested data
	if val, ok := unmarshaled.Data["key"].(string); !ok || val != "value" {
		t.Errorf("Expected Data['key'] to be 'value', got %v", unmarshaled.Data["key"])
	}
	if val, ok := unmarshaled.Data["count"].(float64); !ok || val != 42 {
		t.Errorf("Expected Data['count'] to be 42, got %v", unmarshaled.Data["count"])
	}
}

// TestAccessTokenModel tests the AccessToken model structure and JSON marshaling
func TestAccessTokenModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	token := &AccessToken{
		Token:     "token-abc123",
		UserID:    "user-456",
		TenantID:  "tenant-789",
		Scopes:    []string{"read", "write", "admin"},
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Failed to marshal access token: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled AccessToken
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal access token: %v", err)
	}

	// Verify fields
	if unmarshaled.Token != token.Token {
		t.Errorf("Expected Token %s, got %s", token.Token, unmarshaled.Token)
	}
	if unmarshaled.UserID != token.UserID {
		t.Errorf("Expected UserID %s, got %s", token.UserID, unmarshaled.UserID)
	}
	if unmarshaled.TenantID != token.TenantID {
		t.Errorf("Expected TenantID %s, got %s", token.TenantID, unmarshaled.TenantID)
	}

	// Verify scopes
	if len(unmarshaled.Scopes) != len(token.Scopes) {
		t.Errorf("Expected %d scopes, got %d", len(token.Scopes), len(unmarshaled.Scopes))
	}
	for i, scope := range token.Scopes {
		if unmarshaled.Scopes[i] != scope {
			t.Errorf("Expected scope[%d] to be %s, got %s", i, scope, unmarshaled.Scopes[i])
		}
	}
}

// TestTenantModel tests the Tenant model structure and JSON marshaling
func TestTenantModel(t *testing.T) {
	now := time.Now()

	tenant := &Tenant{
		ID:          "tenant-123",
		Name:        "Test Tenant",
		Hosts:       []string{"example.com", "www.example.com"},
		Config:      map[string]interface{}{"theme": "dark", "max_connections": 100},
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		MaxUsers:    1000,
		MaxStorage:  1073741824, // 1GB
		MaxRequests: 10000,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(tenant)
	if err != nil {
		t.Fatalf("Failed to marshal tenant: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Tenant
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal tenant: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != tenant.ID {
		t.Errorf("Expected ID %s, got %s", tenant.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != tenant.Name {
		t.Errorf("Expected Name %s, got %s", tenant.Name, unmarshaled.Name)
	}
	if unmarshaled.IsActive != tenant.IsActive {
		t.Errorf("Expected IsActive %v, got %v", tenant.IsActive, unmarshaled.IsActive)
	}
	if unmarshaled.MaxUsers != tenant.MaxUsers {
		t.Errorf("Expected MaxUsers %d, got %d", tenant.MaxUsers, unmarshaled.MaxUsers)
	}
	if unmarshaled.MaxStorage != tenant.MaxStorage {
		t.Errorf("Expected MaxStorage %d, got %d", tenant.MaxStorage, unmarshaled.MaxStorage)
	}
	if unmarshaled.MaxRequests != tenant.MaxRequests {
		t.Errorf("Expected MaxRequests %d, got %d", tenant.MaxRequests, unmarshaled.MaxRequests)
	}

	// Verify hosts
	if len(unmarshaled.Hosts) != len(tenant.Hosts) {
		t.Errorf("Expected %d hosts, got %d", len(tenant.Hosts), len(unmarshaled.Hosts))
	}
	for i, host := range tenant.Hosts {
		if unmarshaled.Hosts[i] != host {
			t.Errorf("Expected host[%d] to be %s, got %s", i, host, unmarshaled.Hosts[i])
		}
	}

	// Verify config
	if val, ok := unmarshaled.Config["theme"].(string); !ok || val != "dark" {
		t.Errorf("Expected Config['theme'] to be 'dark', got %v", unmarshaled.Config["theme"])
	}
	if val, ok := unmarshaled.Config["max_connections"].(float64); !ok || val != 100 {
		t.Errorf("Expected Config['max_connections'] to be 100, got %v", unmarshaled.Config["max_connections"])
	}
}

// TestWorkloadMetricsModel tests the WorkloadMetrics model structure and JSON marshaling
func TestWorkloadMetricsModel(t *testing.T) {
	now := time.Now()

	metrics := &WorkloadMetrics{
		ID:           1,
		Timestamp:    now,
		TenantID:     "tenant-123",
		UserID:       "user-456",
		RequestID:    "req-789",
		Duration:     150,
		ContextSize:  2048,
		MemoryUsage:  4096000,
		CPUUsage:     45.5,
		Path:         "/api/users",
		Method:       "GET",
		StatusCode:   200,
		ResponseSize: 1024,
		ErrorMessage: "",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal workload metrics: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled WorkloadMetrics
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal workload metrics: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != metrics.ID {
		t.Errorf("Expected ID %d, got %d", metrics.ID, unmarshaled.ID)
	}
	if unmarshaled.TenantID != metrics.TenantID {
		t.Errorf("Expected TenantID %s, got %s", metrics.TenantID, unmarshaled.TenantID)
	}
	if unmarshaled.UserID != metrics.UserID {
		t.Errorf("Expected UserID %s, got %s", metrics.UserID, unmarshaled.UserID)
	}
	if unmarshaled.RequestID != metrics.RequestID {
		t.Errorf("Expected RequestID %s, got %s", metrics.RequestID, unmarshaled.RequestID)
	}
	if unmarshaled.Duration != metrics.Duration {
		t.Errorf("Expected Duration %d, got %d", metrics.Duration, unmarshaled.Duration)
	}
	if unmarshaled.ContextSize != metrics.ContextSize {
		t.Errorf("Expected ContextSize %d, got %d", metrics.ContextSize, unmarshaled.ContextSize)
	}
	if unmarshaled.MemoryUsage != metrics.MemoryUsage {
		t.Errorf("Expected MemoryUsage %d, got %d", metrics.MemoryUsage, unmarshaled.MemoryUsage)
	}
	if unmarshaled.CPUUsage != metrics.CPUUsage {
		t.Errorf("Expected CPUUsage %f, got %f", metrics.CPUUsage, unmarshaled.CPUUsage)
	}
	if unmarshaled.Path != metrics.Path {
		t.Errorf("Expected Path %s, got %s", metrics.Path, unmarshaled.Path)
	}
	if unmarshaled.Method != metrics.Method {
		t.Errorf("Expected Method %s, got %s", metrics.Method, unmarshaled.Method)
	}
	if unmarshaled.StatusCode != metrics.StatusCode {
		t.Errorf("Expected StatusCode %d, got %d", metrics.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.ResponseSize != metrics.ResponseSize {
		t.Errorf("Expected ResponseSize %d, got %d", metrics.ResponseSize, unmarshaled.ResponseSize)
	}
	if unmarshaled.ErrorMessage != metrics.ErrorMessage {
		t.Errorf("Expected ErrorMessage %s, got %s", metrics.ErrorMessage, unmarshaled.ErrorMessage)
	}
}

// TestSessionDataTypes tests that Session.Data can handle various data types
func TestSessionDataTypes(t *testing.T) {
	session := &Session{
		ID: "test-session",
		Data: map[string]interface{}{
			"string": "value",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"array":  []interface{}{1, 2, 3},
			"nested": map[string]interface{}{"key": "nested-value"},
		},
	}

	// Marshal and unmarshal
	jsonData, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session with various data types: %v", err)
	}

	var unmarshaled Session
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal session with various data types: %v", err)
	}

	// Verify all data types are preserved
	if unmarshaled.Data["string"] != "value" {
		t.Errorf("String value not preserved")
	}
	if unmarshaled.Data["bool"] != true {
		t.Errorf("Bool value not preserved")
	}

	// Numbers are unmarshaled as float64 in JSON
	if val, ok := unmarshaled.Data["int"].(float64); !ok || val != 42 {
		t.Errorf("Int value not preserved correctly")
	}
	if val, ok := unmarshaled.Data["float"].(float64); !ok || val != 3.14 {
		t.Errorf("Float value not preserved correctly")
	}
}

// TestTenantConfigTypes tests that Tenant.Config can handle various data types
func TestTenantConfigTypes(t *testing.T) {
	tenant := &Tenant{
		ID:   "test-tenant",
		Name: "Test",
		Config: map[string]interface{}{
			"string": "value",
			"int":    100,
			"float":  99.99,
			"bool":   false,
			"array":  []interface{}{"a", "b", "c"},
		},
	}

	// Marshal and unmarshal
	jsonData, err := json.Marshal(tenant)
	if err != nil {
		t.Fatalf("Failed to marshal tenant with various config types: %v", err)
	}

	var unmarshaled Tenant
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal tenant with various config types: %v", err)
	}

	// Verify all config types are preserved
	if unmarshaled.Config["string"] != "value" {
		t.Errorf("String config not preserved")
	}
	if unmarshaled.Config["bool"] != false {
		t.Errorf("Bool config not preserved")
	}
}

// TestAccessTokenEmptyScopes tests AccessToken with empty scopes
func TestAccessTokenEmptyScopes(t *testing.T) {
	token := &AccessToken{
		Token:     "token-123",
		UserID:    "user-456",
		TenantID:  "tenant-789",
		Scopes:    []string{},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Failed to marshal token with empty scopes: %v", err)
	}

	var unmarshaled AccessToken
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal token with empty scopes: %v", err)
	}

	if unmarshaled.Scopes == nil {
		t.Errorf("Expected empty scopes array, got nil")
	}
	if len(unmarshaled.Scopes) != 0 {
		t.Errorf("Expected 0 scopes, got %d", len(unmarshaled.Scopes))
	}
}

// TestWorkloadMetricsWithError tests WorkloadMetrics with error message
func TestWorkloadMetricsWithError(t *testing.T) {
	metrics := &WorkloadMetrics{
		ID:           1,
		Timestamp:    time.Now(),
		TenantID:     "tenant-123",
		UserID:       "user-456",
		RequestID:    "req-789",
		Duration:     500,
		Path:         "/api/error",
		Method:       "POST",
		StatusCode:   500,
		ErrorMessage: "Internal server error: database connection failed",
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal metrics with error: %v", err)
	}

	var unmarshaled WorkloadMetrics
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal metrics with error: %v", err)
	}

	if unmarshaled.ErrorMessage != metrics.ErrorMessage {
		t.Errorf("Expected ErrorMessage %s, got %s", metrics.ErrorMessage, unmarshaled.ErrorMessage)
	}
	if unmarshaled.StatusCode != 500 {
		t.Errorf("Expected StatusCode 500, got %d", unmarshaled.StatusCode)
	}
}
