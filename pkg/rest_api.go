package pkg

import (
	"encoding/json"
	"fmt"
	"time"
)

// RESTAPIManager defines the REST API interface
type RESTAPIManager interface {
	// Route registration with rate limiting
	RegisterRoute(method, path string, handler RESTHandler, config RESTRouteConfig) error

	// Rate limiting
	CheckRateLimit(ctx Context, resource string) error
	CheckGlobalRateLimit(ctx Context) error

	// Request/Response handling
	ParseJSONRequest(ctx Context, target interface{}) error
	SendJSONResponse(ctx Context, statusCode int, data interface{}) error
	SendErrorResponse(ctx Context, statusCode int, message string, details map[string]interface{}) error

	// Middleware support
	Use(middleware RESTMiddleware) RESTAPIManager

	// Route groups
	Group(prefix string, middleware ...RESTMiddleware) RESTAPIManager
}

// RESTHandler represents a REST API handler function
type RESTHandler func(ctx Context) error

// RESTMiddleware represents REST API middleware
type RESTMiddleware func(ctx Context, next RESTHandler) error

// RESTRouteConfig defines configuration for a REST route
type RESTRouteConfig struct {
	// Rate limiting configuration
	RateLimit       *RESTRateLimitConfig
	GlobalRateLimit *RESTRateLimitConfig

	// Authentication and authorization
	RequireAuth    bool
	RequiredScopes []string

	// Request validation
	MaxRequestSize int64
	Timeout        time.Duration

	// Response configuration
	CacheControl string
	CORS         *CORSConfig
}

// RESTRateLimitConfig defines rate limiting configuration for REST APIs
type RESTRateLimitConfig struct {
	Limit  int           // Maximum number of requests
	Window time.Duration // Time window for the limit
	Key    string        // Rate limit key (e.g., "user_id", "ip_address", "tenant_id")
}

// RESTError represents a REST API error response
type RESTError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Status  int                    `json:"-"`
}

// Error implements the error interface
func (e *RESTError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// RESTResponse represents a standard REST API response
type RESTResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *RESTError  `json:"error,omitempty"`
	Meta    *RESTMeta   `json:"meta,omitempty"`
}

// RESTMeta represents metadata for REST API responses
type RESTMeta struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id"`
	Version    string                 `json:"version,omitempty"`
	Pagination *Pagination            `json:"pagination,omitempty"`
	RateLimit  *RateLimitInfo         `json:"rate_limit,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// RateLimitInfo represents rate limit information in responses
type RateLimitInfo struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"` // Unix timestamp
}

// NewRESTError creates a new REST API error
func NewRESTError(code, message string, status int) *RESTError {
	return &RESTError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// WithDetails adds details to a REST error
func (e *RESTError) WithDetails(details map[string]interface{}) *RESTError {
	e.Details = details
	return e
}

// Common REST API errors
var (
	ErrBadRequest          = NewRESTError("BAD_REQUEST", "Invalid request", 400)
	ErrUnauthorized        = NewRESTError("UNAUTHORIZED", "Authentication required", 401)
	ErrForbidden           = NewRESTError("FORBIDDEN", "Access denied", 403)
	ErrNotFound            = NewRESTError("NOT_FOUND", "Resource not found", 404)
	ErrMethodNotAllowed    = NewRESTError("METHOD_NOT_ALLOWED", "Method not allowed", 405)
	ErrConflict            = NewRESTError("CONFLICT", "Resource conflict", 409)
	ErrRateLimitExceeded   = NewRESTError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", 429)
	ErrInternalServerError = NewRESTError("INTERNAL_SERVER_ERROR", "Internal server error", 500)
	ErrServiceUnavailable  = NewRESTError("SERVICE_UNAVAILABLE", "Service unavailable", 503)
)

// Helper functions for JSON handling

// ParseJSON parses JSON from request body into target
func ParseJSON(data []byte, target interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty request body")
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// ToJSON converts data to JSON bytes
func ToJSON(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return jsonData, nil
}

// ToJSONIndent converts data to indented JSON bytes
func ToJSONIndent(data interface{}) ([]byte, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return jsonData, nil
}
