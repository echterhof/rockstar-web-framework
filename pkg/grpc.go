package pkg

import (
	"context"
	"fmt"
	"time"
)

// GRPCManager defines the gRPC interface
// Requirements: 2.3
type GRPCManager interface {
	// Service registration
	RegisterService(service GRPCService, config GRPCConfig) error

	// Rate limiting
	// Requirements: 2.6
	CheckRateLimit(ctx Context, resource string) error
	CheckGlobalRateLimit(ctx Context) error

	// Middleware support
	Use(middleware GRPCMiddleware) GRPCManager

	// Service groups
	Group(prefix string, middleware ...GRPCMiddleware) GRPCManager

	// Server lifecycle
	Start(addr string) error
	Stop() error
	GracefulStop(timeout time.Duration) error
}

// GRPCMiddleware represents gRPC middleware
type GRPCMiddleware func(ctx Context, next GRPCHandler) error

// GRPCHandler represents a gRPC handler function
type GRPCHandler func(ctx Context) error

// GRPCConfig defines configuration for a gRPC service
type GRPCConfig struct {
	// Rate limiting configuration
	// Requirements: 2.6
	RateLimit       *GRPCRateLimitConfig
	GlobalRateLimit *GRPCRateLimitConfig

	// Authentication and authorization
	// Requirements: 2.5
	RequireAuth    bool
	RequiredScopes []string
	RequiredRoles  []string

	// Request validation
	MaxRequestSize  int64
	MaxResponseSize int64
	Timeout         time.Duration

	// Connection settings
	MaxConcurrentStreams uint32
	KeepAlive            *GRPCKeepAliveConfig

	// TLS configuration
	TLSCertFile string
	TLSKeyFile  string

	// Interceptors
	UnaryInterceptors  []GRPCUnaryInterceptor
	StreamInterceptors []GRPCStreamInterceptor
}

// GRPCRateLimitConfig defines rate limiting configuration for gRPC
// Requirements: 2.6
type GRPCRateLimitConfig struct {
	Limit  int           // Maximum number of requests
	Window time.Duration // Time window for the limit
	Key    string        // Rate limit key (e.g., "user_id", "ip_address", "tenant_id")
}

// GRPCKeepAliveConfig defines keep-alive configuration
type GRPCKeepAliveConfig struct {
	Time    time.Duration // Time between keep-alive pings
	Timeout time.Duration // Timeout for keep-alive ping response
}

// GRPCUnaryInterceptor represents a unary RPC interceptor
type GRPCUnaryInterceptor func(ctx context.Context, req interface{}, info *GRPCUnaryServerInfo, handler GRPCUnaryHandler) (interface{}, error)

// GRPCStreamInterceptor represents a streaming RPC interceptor
type GRPCStreamInterceptor func(srv interface{}, ss GRPCServerStream, info *GRPCStreamServerInfo, handler GRPCStreamHandler) error

// GRPCUnaryHandler represents a unary RPC handler
type GRPCUnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// GRPCStreamHandler represents a streaming RPC handler
type GRPCStreamHandler func(srv interface{}, stream GRPCServerStream) error

// GRPCUnaryServerInfo contains information about a unary RPC
type GRPCUnaryServerInfo struct {
	FullMethod string
	Server     interface{}
}

// GRPCStreamServerInfo contains information about a streaming RPC
type GRPCStreamServerInfo struct {
	FullMethod     string
	IsClientStream bool
	IsServerStream bool
}

// GRPCServerStream represents a server-side stream
type GRPCServerStream interface {
	SetHeader(key, value string) error
	SendHeader() error
	SetTrailer(key, value string)
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

// GRPCRequest represents a gRPC request
type GRPCRequest struct {
	Service    string
	Method     string
	Message    interface{}
	Metadata   map[string]string
	RemoteAddr string
}

// GRPCResponse represents a gRPC response
type GRPCResponse struct {
	Message  interface{}
	Metadata map[string]string
	Trailer  map[string]string
	Error    *GRPCError
}

// GRPCError represents a gRPC error
type GRPCError struct {
	Code    GRPCStatusCode
	Message string
	Details []interface{}
}

// GRPCStatusCode represents gRPC status codes
type GRPCStatusCode int

const (
	GRPCStatusOK                 GRPCStatusCode = 0
	GRPCStatusCanceled           GRPCStatusCode = 1
	GRPCStatusUnknown            GRPCStatusCode = 2
	GRPCStatusInvalidArgument    GRPCStatusCode = 3
	GRPCStatusDeadlineExceeded   GRPCStatusCode = 4
	GRPCStatusNotFound           GRPCStatusCode = 5
	GRPCStatusAlreadyExists      GRPCStatusCode = 6
	GRPCStatusPermissionDenied   GRPCStatusCode = 7
	GRPCStatusResourceExhausted  GRPCStatusCode = 8
	GRPCStatusFailedPrecondition GRPCStatusCode = 9
	GRPCStatusAborted            GRPCStatusCode = 10
	GRPCStatusOutOfRange         GRPCStatusCode = 11
	GRPCStatusUnimplemented      GRPCStatusCode = 12
	GRPCStatusInternal           GRPCStatusCode = 13
	GRPCStatusUnavailable        GRPCStatusCode = 14
	GRPCStatusDataLoss           GRPCStatusCode = 15
	GRPCStatusUnauthenticated    GRPCStatusCode = 16
)

// Error implements the error interface
func (e *GRPCError) Error() string {
	return fmt.Sprintf("gRPC error %d: %s", e.Code, e.Message)
}

// NewGRPCError creates a new gRPC error
func NewGRPCError(code GRPCStatusCode, message string) *GRPCError {
	return &GRPCError{
		Code:    code,
		Message: message,
		Details: make([]interface{}, 0),
	}
}

// WithDetails adds details to a gRPC error
func (e *GRPCError) WithDetails(details ...interface{}) *GRPCError {
	e.Details = append(e.Details, details...)
	return e
}

// Common gRPC errors
var (
	ErrGRPCInvalidArgument   = NewGRPCError(GRPCStatusInvalidArgument, "Invalid argument")
	ErrGRPCUnauthenticated   = NewGRPCError(GRPCStatusUnauthenticated, "Authentication required")
	ErrGRPCPermissionDenied  = NewGRPCError(GRPCStatusPermissionDenied, "Permission denied")
	ErrGRPCNotFound          = NewGRPCError(GRPCStatusNotFound, "Not found")
	ErrGRPCResourceExhausted = NewGRPCError(GRPCStatusResourceExhausted, "Rate limit exceeded")
	ErrGRPCInternal          = NewGRPCError(GRPCStatusInternal, "Internal server error")
	ErrGRPCUnavailable       = NewGRPCError(GRPCStatusUnavailable, "Service unavailable")
	ErrGRPCDeadlineExceeded  = NewGRPCError(GRPCStatusDeadlineExceeded, "Deadline exceeded")
)

// GRPCServiceExtended represents an extended gRPC service interface
type GRPCServiceExtended interface {
	GRPCService

	// Handle a unary RPC call
	HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error)

	// Handle a streaming RPC call
	HandleStream(stream GRPCServerStream, method string) error

	// Get method descriptor
	GetMethodDescriptor(method string) *GRPCMethodDescriptor
}

// GRPCMethodDescriptor describes a gRPC method
type GRPCMethodDescriptor struct {
	Name           string
	IsClientStream bool
	IsServerStream bool
	InputType      interface{}
	OutputType     interface{}
}

// Helper functions for gRPC metadata handling

// ExtractMetadata extracts metadata from context
func ExtractMetadata(ctx context.Context) map[string]string {
	// This would extract metadata from the gRPC context
	// For now, return empty map
	return make(map[string]string)
}

// InjectMetadata injects metadata into context
func InjectMetadata(ctx context.Context, metadata map[string]string) context.Context {
	// This would inject metadata into the gRPC context
	// For now, return the original context
	return ctx
}
