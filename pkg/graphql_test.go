package pkg

import (
	"encoding/json"
	"testing"
	"time"
)

// mockGraphQLSchema implements GraphQLSchema for testing
type mockGraphQLSchema struct {
	executeFunc func(query string, variables map[string]interface{}) (interface{}, error)
}

func (m *mockGraphQLSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(query, variables)
	}
	return map[string]interface{}{
		"hello": "world",
	}, nil
}

// TestNewGraphQLManager tests GraphQL manager creation
func TestNewGraphQLManager(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	if manager == nil {
		t.Fatal("Expected GraphQL manager to be created")
	}
}

// TestGraphQLManagerRegisterSchema tests schema registration
// Requirements: 2.1
func TestGraphQLManagerRegisterSchema(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	schema := &mockGraphQLSchema{}
	config := GraphQLConfig{
		EnableIntrospection: true,
		EnablePlayground:    true,
	}

	err := manager.RegisterSchema("/graphql", schema, config)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Verify routes were registered
	routes := router.Routes()
	if len(routes) < 2 {
		t.Errorf("Expected at least 2 routes (POST and GET), got %d", len(routes))
	}

	// Check POST route
	foundPost := false
	foundGet := false
	for _, route := range routes {
		if route.Path == "/graphql" && route.Method == "POST" {
			foundPost = true
		}
		if route.Path == "/graphql" && route.Method == "GET" {
			foundGet = true
		}
	}

	if !foundPost {
		t.Error("Expected POST route to be registered")
	}
	if !foundGet {
		t.Error("Expected GET route to be registered for playground")
	}
}

// TestGraphQLManagerWithAuthentication tests authentication integration
// Requirements: 2.5
func TestGraphQLManagerWithAuthentication(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	schema := &mockGraphQLSchema{}
	config := GraphQLConfig{
		RequireAuth:    true,
		RequiredRoles:  []string{"admin"},
		RequiredScopes: []string{"read:data"},
	}

	err := manager.RegisterSchema("/graphql", schema, config)
	if err != nil {
		t.Fatalf("Failed to register schema with auth: %v", err)
	}

	// Verify route was registered
	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatal("Expected route to be registered")
	}
}

// TestGraphQLManagerWithRateLimiting tests rate limiting integration
// Requirements: 2.6
func TestGraphQLManagerWithRateLimiting(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	schema := &mockGraphQLSchema{}
	config := GraphQLConfig{
		RateLimit: &GraphQLRateLimitConfig{
			Limit:  10,
			Window: time.Minute,
			Key:    "ip_address",
		},
		GlobalRateLimit: &GraphQLRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "tenant_id",
		},
	}

	err := manager.RegisterSchema("/graphql", schema, config)
	if err != nil {
		t.Fatalf("Failed to register schema with rate limiting: %v", err)
	}

	// Verify route was registered
	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatal("Expected route to be registered")
	}
}

// TestGraphQLManagerUseMiddleware tests middleware support
func TestGraphQLManagerUseMiddleware(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	middleware := func(ctx Context, next GraphQLHandler) error {
		return next(ctx)
	}

	manager.Use(middleware)

	// Register a schema to trigger middleware
	schema := &mockGraphQLSchema{}
	config := GraphQLConfig{}

	err := manager.RegisterSchema("/graphql", schema, config)
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}
}

// TestGraphQLManagerGroup tests route grouping
func TestGraphQLManagerGroup(t *testing.T) {
	router := NewRouter()
	manager := NewGraphQLManager(router, nil, nil)

	// Create a group with prefix
	apiGroup := manager.Group("/api/v1")

	schema := &mockGraphQLSchema{}
	config := GraphQLConfig{}

	err := apiGroup.RegisterSchema("/graphql", schema, config)
	if err != nil {
		t.Fatalf("Failed to register schema in group: %v", err)
	}

	// Verify route was registered with prefix
	routes := router.Routes()
	foundWithPrefix := false
	for _, route := range routes {
		if route.Path == "/api/v1/graphql" {
			foundWithPrefix = true
			break
		}
	}

	if !foundWithPrefix {
		t.Error("Expected route to be registered with group prefix")
	}
}

// TestParseGraphQLRequest tests GraphQL request parsing
func TestParseGraphQLRequest(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantError bool
	}{
		{
			name:      "Valid request",
			json:      `{"query": "{ hello }"}`,
			wantError: false,
		},
		{
			name:      "Valid request with variables",
			json:      `{"query": "query($id: ID!) { user(id: $id) { name } }", "variables": {"id": "123"}}`,
			wantError: false,
		},
		{
			name:      "Valid request with operation name",
			json:      `{"query": "query GetUser { user { name } }", "operationName": "GetUser"}`,
			wantError: false,
		},
		{
			name:      "Empty body",
			json:      "",
			wantError: true,
		},
		{
			name:      "Invalid JSON",
			json:      `{invalid}`,
			wantError: true,
		},
		{
			name:      "Missing query",
			json:      `{"variables": {}}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseGraphQLRequest([]byte(tt.json))

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if req == nil {
					t.Error("Expected request to be parsed")
				}
			}
		})
	}
}

// TestGraphQLError tests GraphQL error creation and methods
func TestGraphQLError(t *testing.T) {
	err := NewGraphQLError("Test error")

	if err.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got '%s'", err.Message)
	}

	// Test WithPath
	err = err.WithPath("user", "name")
	if len(err.Path) != 2 {
		t.Errorf("Expected path length 2, got %d", len(err.Path))
	}

	// Test WithLocation
	err = err.WithLocation(10, 5)
	if len(err.Locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(err.Locations))
	}
	if err.Locations[0].Line != 10 || err.Locations[0].Column != 5 {
		t.Errorf("Expected location (10, 5), got (%d, %d)", err.Locations[0].Line, err.Locations[0].Column)
	}

	// Test WithExtensions
	err = err.WithExtensions(map[string]interface{}{
		"code": "VALIDATION_ERROR",
	})
	if err.Extensions["code"] != "VALIDATION_ERROR" {
		t.Error("Expected extension to be set")
	}
}

// TestToGraphQLJSON tests GraphQL response serialization
func TestToGraphQLJSON(t *testing.T) {
	response := &GraphQLResponse{
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "123",
				"name": "John Doe",
			},
		},
	}

	jsonData, err := ToGraphQLJSON(response)
	if err != nil {
		t.Fatalf("Failed to serialize response: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	// Verify data is present
	if parsed["data"] == nil {
		t.Error("Expected data field in response")
	}
}

// TestToGraphQLJSONIndent tests indented GraphQL response serialization
func TestToGraphQLJSONIndent(t *testing.T) {
	response := &GraphQLResponse{
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "123",
				"name": "John Doe",
			},
		},
	}

	jsonData, err := ToGraphQLJSONIndent(response)
	if err != nil {
		t.Fatalf("Failed to serialize response: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	// Verify it's indented (contains newlines)
	if len(jsonData) > 0 && !containsNewline(jsonData) {
		t.Error("Expected indented JSON to contain newlines")
	}
}

func containsNewline(data []byte) bool {
	for _, b := range data {
		if b == '\n' {
			return true
		}
	}
	return false
}

// TestGraphQLConfig tests GraphQL configuration
func TestGraphQLConfig(t *testing.T) {
	config := GraphQLConfig{
		RequireAuth:         true,
		RequiredScopes:      []string{"read:data", "write:data"},
		RequiredRoles:       []string{"admin"},
		MaxRequestSize:      1024 * 1024, // 1MB
		MaxQueryDepth:       10,
		MaxComplexity:       1000,
		Timeout:             30 * time.Second,
		EnableIntrospection: true,
		EnablePlayground:    true,
		RateLimit: &GraphQLRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "user_id",
		},
	}

	if !config.RequireAuth {
		t.Error("Expected RequireAuth to be true")
	}

	if len(config.RequiredScopes) != 2 {
		t.Errorf("Expected 2 required scopes, got %d", len(config.RequiredScopes))
	}

	if config.RateLimit == nil {
		t.Error("Expected rate limit config to be set")
	}

	if config.RateLimit.Limit != 100 {
		t.Errorf("Expected rate limit of 100, got %d", config.RateLimit.Limit)
	}
}

// TestGraphQLResponse tests GraphQL response structure
func TestGraphQLResponse(t *testing.T) {
	response := &GraphQLResponse{
		Data: map[string]interface{}{
			"hello": "world",
		},
		Errors: []GraphQLError{
			NewGraphQLError("Test error"),
		},
		Extensions: &GraphQLExtensions{
			Timestamp:  time.Now(),
			RequestID:  "test-123",
			Complexity: 50,
		},
	}

	if response.Data == nil {
		t.Error("Expected data to be set")
	}

	if len(response.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(response.Errors))
	}

	if response.Extensions == nil {
		t.Error("Expected extensions to be set")
	}

	if response.Extensions.RequestID != "test-123" {
		t.Errorf("Expected request ID 'test-123', got '%s'", response.Extensions.RequestID)
	}
}
