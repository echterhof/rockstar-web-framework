package pkg

import (
	"encoding/json"
	"fmt"
	"time"
)

// GraphQLManager defines the GraphQL interface
type GraphQLManager interface {
	// Schema registration
	RegisterSchema(path string, schema GraphQLSchema, config GraphQLConfig) error

	// Query execution
	ExecuteQuery(ctx Context, query string, variables map[string]interface{}, operationName string) (*GraphQLResponse, error)

	// Rate limiting
	CheckRateLimit(ctx Context, resource string) error
	CheckGlobalRateLimit(ctx Context) error

	// Middleware support
	Use(middleware GraphQLMiddleware) GraphQLManager

	// Schema groups
	Group(prefix string, middleware ...GraphQLMiddleware) GraphQLManager
}

// GraphQLMiddleware represents GraphQL middleware
type GraphQLMiddleware func(ctx Context, next GraphQLHandler) error

// GraphQLHandler represents a GraphQL handler function
type GraphQLHandler func(ctx Context) error

// GraphQLConfig defines configuration for a GraphQL endpoint
type GraphQLConfig struct {
	// Rate limiting configuration
	RateLimit       *GraphQLRateLimitConfig
	GlobalRateLimit *GraphQLRateLimitConfig

	// Authentication and authorization
	RequireAuth    bool
	RequiredScopes []string
	RequiredRoles  []string

	// Request validation
	MaxRequestSize int64
	MaxQueryDepth  int
	MaxComplexity  int
	Timeout        time.Duration

	// Response configuration
	CacheControl string
	CORS         *CORSConfig

	// GraphQL-specific settings
	EnableIntrospection bool
	EnablePlayground    bool
	PersistentQueries   bool
}

// GraphQLRateLimitConfig defines rate limiting configuration for GraphQL
type GraphQLRateLimitConfig struct {
	Limit  int           // Maximum number of requests
	Window time.Duration // Time window for the limit
	Key    string        // Rate limit key (e.g., "user_id", "ip_address", "tenant_id")
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}        `json:"data,omitempty"`
	Errors     []GraphQLError     `json:"errors,omitempty"`
	Extensions *GraphQLExtensions `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of an error in the query
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// GraphQLExtensions represents additional metadata in the response
type GraphQLExtensions struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id"`
	Complexity int                    `json:"complexity,omitempty"`
	RateLimit  *RateLimitInfo         `json:"rate_limit,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// GraphQLSchemaExtended represents an extended GraphQL schema interface
// Note: GraphQLSchema is defined in router.go for compatibility
type GraphQLSchemaExtended interface {
	GraphQLSchema

	// Validate a query
	Validate(query string) error

	// Get schema definition
	Schema() string

	// Introspection support
	Introspect() (interface{}, error)

	// Execute with context and operation name
	ExecuteWithContext(query string, variables map[string]interface{}, operationName string, ctx Context) (*GraphQLResponse, error)
}

// NewGraphQLError creates a new GraphQL error
func NewGraphQLError(message string) GraphQLError {
	return GraphQLError{
		Message: message,
	}
}

// WithPath adds a path to a GraphQL error
func (e GraphQLError) WithPath(path ...interface{}) GraphQLError {
	e.Path = path
	return e
}

// WithLocation adds a location to a GraphQL error
func (e GraphQLError) WithLocation(line, column int) GraphQLError {
	if e.Locations == nil {
		e.Locations = []GraphQLErrorLocation{}
	}
	e.Locations = append(e.Locations, GraphQLErrorLocation{
		Line:   line,
		Column: column,
	})
	return e
}

// WithExtensions adds extensions to a GraphQL error
func (e GraphQLError) WithExtensions(extensions map[string]interface{}) GraphQLError {
	e.Extensions = extensions
	return e
}

// Common GraphQL errors
var (
	ErrGraphQLSyntax         = NewGraphQLError("Syntax error in GraphQL query")
	ErrGraphQLValidation     = NewGraphQLError("Validation error in GraphQL query")
	ErrGraphQLExecution      = NewGraphQLError("Execution error in GraphQL query")
	ErrGraphQLAuthentication = NewGraphQLError("Authentication required")
	ErrGraphQLAuthorization  = NewGraphQLError("Insufficient permissions")
	ErrGraphQLRateLimit      = NewGraphQLError("Rate limit exceeded")
	ErrGraphQLComplexity     = NewGraphQLError("Query complexity exceeds limit")
	ErrGraphQLDepth          = NewGraphQLError("Query depth exceeds limit")
	ErrGraphQLTimeout        = NewGraphQLError("Query execution timeout")
)

// Helper functions for GraphQL request/response handling

// ParseGraphQLRequest parses a GraphQL request from JSON
func ParseGraphQLRequest(data []byte) (*GraphQLRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	var req GraphQLRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	return &req, nil
}

// ToGraphQLJSON converts a GraphQL response to JSON bytes
func ToGraphQLJSON(response *GraphQLResponse) ([]byte, error) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL response: %w", err)
	}
	return jsonData, nil
}

// ToGraphQLJSONIndent converts a GraphQL response to indented JSON bytes
func ToGraphQLJSONIndent(response *GraphQLResponse) ([]byte, error) {
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL response: %w", err)
	}
	return jsonData, nil
}
