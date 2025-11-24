# gRPC Implementation

## Overview

The Rockstar Web Framework provides comprehensive gRPC support with built-in authentication, authorization, and rate limiting. The implementation follows the same patterns as GraphQL and REST API support, providing a consistent developer experience across all API protocols.

## Requirements

This implementation satisfies the following requirements:
- **Requirement 2.3**: Support for gRPC protocol access
- **Requirement 2.5**: API access validation with access tokens stored in database
- **Requirement 2.6**: Rate limiting per resource or globally with database storage

## Features

### Core Features
- Service registration with configuration
- Unary and streaming RPC support
- Method-based routing
- Middleware support
- Service grouping with prefixes

### Authentication & Authorization
- OAuth2 and JWT authentication integration
- Role-based access control (RBAC)
- Scope-based authorization
- Access token validation

### Rate Limiting
- Per-resource rate limiting
- Global rate limiting
- Configurable limits and time windows
- Database-backed rate limit storage
- Multiple key types (user_id, tenant_id, ip_address)

### Request Validation
- Request size limits
- Response size limits
- Timeout configuration
- Automatic validation middleware

### Error Handling
- Standard gRPC status codes
- Detailed error messages
- Error details support
- Automatic HTTP status code mapping

## Architecture

### Components

1. **GRPCManager**: Main interface for gRPC functionality
2. **GRPCService**: Interface for implementing gRPC services
3. **GRPCMiddleware**: Middleware system for request processing
4. **GRPCConfig**: Configuration for services and endpoints

### Request Flow

```
Client Request
    ↓
HTTP/2 POST to /ServiceName/MethodName
    ↓
Router matches route
    ↓
Middleware chain (custom → validation → auth → rate limiting)
    ↓
Service handler execution
    ↓
Response serialization
    ↓
Client Response
```

## Usage

### Basic Service Implementation

```go
// Define your service
type UserService struct {
    users map[string]*User
}

// Implement GRPCService interface
func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) Methods() []string {
    return []string{"GetUser", "CreateUser", "UpdateUser"}
}

// Implement GRPCServiceExtended for handling calls
func (s *UserService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
    switch method {
    case "GetUser":
        return s.getUser(req)
    case "CreateUser":
        return s.createUser(req)
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}
```

### Service Registration

```go
// Create gRPC manager
grpcManager := pkg.NewGRPCManager(router, db, authManager)

// Register service with configuration
err := grpcManager.RegisterService(userService, pkg.GRPCConfig{
    RequireAuth: false,
    RateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
})
```

### Authentication & Authorization

```go
// Register service with authentication
err := grpcManager.RegisterService(adminService, pkg.GRPCConfig{
    RequireAuth:   true,
    RequiredRoles: []string{"admin"},
    RequiredScopes: []string{"admin:write"},
})
```

### Rate Limiting

```go
// Per-resource rate limiting
config := pkg.GRPCConfig{
    RateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  100,           // 100 requests
        Window: time.Minute,   // per minute
        Key:    "user_id",     // per user
    },
}

// Global rate limiting
config := pkg.GRPCConfig{
    GlobalRateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  1000,
        Window: time.Minute,
        Key:    "ip_address",
    },
}
```

### Custom Middleware

```go
// Add logging middleware
grpcManager.Use(func(ctx pkg.Context, next pkg.GRPCHandler) error {
    start := time.Now()
    log.Printf("Request started: %s", ctx.Request().Path)
    
    err := next(ctx)
    
    duration := time.Since(start)
    log.Printf("Request completed: %s (duration: %v)", ctx.Request().Path, duration)
    
    return err
})
```

### Service Groups

```go
// Create a group with prefix and middleware
apiV1 := grpcManager.Group("/api/v1")

// Register services in the group
apiV1.RegisterService(userService, config)
apiV1.RegisterService(productService, config)
```

### Request Validation

```go
config := pkg.GRPCConfig{
    MaxRequestSize:  1024 * 1024,  // 1MB max request
    MaxResponseSize: 5 * 1024 * 1024,  // 5MB max response
    Timeout:         30 * time.Second,
}
```

## Configuration Options

### GRPCConfig

```go
type GRPCConfig struct {
    // Rate limiting
    RateLimit       *GRPCRateLimitConfig
    GlobalRateLimit *GRPCRateLimitConfig
    
    // Authentication and authorization
    RequireAuth    bool
    RequiredScopes []string
    RequiredRoles  []string
    
    // Request validation
    MaxRequestSize  int64
    MaxResponseSize int64
    Timeout         time.Duration
    
    // Connection settings
    MaxConcurrentStreams uint32
    KeepAlive           *GRPCKeepAliveConfig
    
    // TLS configuration
    TLSCertFile string
    TLSKeyFile  string
}
```

### GRPCRateLimitConfig

```go
type GRPCRateLimitConfig struct {
    Limit  int           // Maximum number of requests
    Window time.Duration // Time window for the limit
    Key    string        // Rate limit key: "user_id", "ip_address", "tenant_id"
}
```

## Error Handling

### gRPC Status Codes

The implementation supports all standard gRPC status codes:

- `GRPCStatusOK` (0): Success
- `GRPCStatusInvalidArgument` (3): Invalid request
- `GRPCStatusUnauthenticated` (16): Authentication required
- `GRPCStatusPermissionDenied` (7): Authorization failed
- `GRPCStatusNotFound` (5): Resource not found
- `GRPCStatusResourceExhausted` (8): Rate limit exceeded
- `GRPCStatusInternal` (13): Internal server error
- `GRPCStatusUnavailable` (14): Service unavailable

### Error Response Format

```go
type GRPCResponse struct {
    Message  interface{}
    Metadata map[string]string
    Trailer  map[string]string
    Error    *GRPCError
}

type GRPCError struct {
    Code    GRPCStatusCode
    Message string
    Details []interface{}
}
```

### Creating Custom Errors

```go
// Create error with details
err := pkg.NewGRPCError(pkg.GRPCStatusInvalidArgument, "Invalid user ID").
    WithDetails("user_id must be a valid UUID")

// Return error from handler
return nil, err
```

## HTTP/2 Integration

gRPC services are exposed over HTTP/2 using POST requests:

- **Endpoint Format**: `POST /ServiceName/MethodName`
- **Content-Type**: `application/grpc` or `application/json`
- **Headers**: Standard HTTP/2 headers plus gRPC metadata

### Example Request

```
POST /UserService/GetUser HTTP/2
Content-Type: application/json
Authorization: Bearer <token>

{"id": "123"}
```

### Example Response

```json
{
  "message": {
    "id": "123",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "metadata": {},
  "trailer": {}
}
```

## Testing

### Unit Tests

The implementation includes comprehensive unit tests covering:

- Service registration
- Authentication and authorization
- Rate limiting
- Request validation
- Middleware execution
- Error handling
- Server lifecycle

### Running Tests

```bash
go test -v ./pkg -run TestGRPC
```

## Performance Considerations

### Rate Limiting

- Rate limits are stored in the database for distributed deployments
- In-memory caching can be added for improved performance
- Rate limit checks are non-blocking

### Connection Management

- HTTP/2 connection reuse
- Keep-alive configuration support
- Graceful shutdown support

### Middleware

- Middleware is executed in order
- Early termination on authentication/authorization failures
- Minimal overhead for authenticated requests

## Best Practices

1. **Service Design**
   - Keep services focused and cohesive
   - Use clear, descriptive method names
   - Document expected request/response formats

2. **Authentication**
   - Always require authentication for sensitive operations
   - Use role-based access control for fine-grained permissions
   - Validate access tokens on every request

3. **Rate Limiting**
   - Set appropriate limits based on service capacity
   - Use different limits for different user tiers
   - Monitor rate limit metrics

4. **Error Handling**
   - Return appropriate gRPC status codes
   - Include helpful error messages
   - Add details for debugging

5. **Middleware**
   - Keep middleware focused and reusable
   - Order middleware appropriately
   - Handle errors gracefully

## Integration with Other Components

### Database Integration

- Access tokens stored in database
- Rate limits stored in database
- Session management integration

### Authentication Integration

- OAuth2 support via AuthManager
- JWT validation
- Role-based authorization

### Router Integration

- Automatic route registration
- Host-based routing support
- Middleware integration

## Future Enhancements

Potential improvements for future versions:

1. **Protocol Buffers Support**
   - Native protobuf serialization/deserialization
   - Schema validation
   - Code generation

2. **Streaming Support**
   - Client streaming
   - Server streaming
   - Bidirectional streaming

3. **Service Discovery**
   - Automatic service registration
   - Health checks
   - Load balancing

4. **Observability**
   - Request tracing
   - Metrics collection
   - Performance monitoring

## Comparison with Other Protocols

| Feature | gRPC | GraphQL | REST |
|---------|------|---------|------|
| Protocol | HTTP/2 | HTTP/1.1+ | HTTP/1.1+ |
| Serialization | Protobuf/JSON | JSON | JSON |
| Type Safety | Strong | Strong | Weak |
| Streaming | Yes | Subscriptions | SSE |
| Browser Support | Limited | Full | Full |
| Performance | High | Medium | Medium |

## References

- [gRPC Official Documentation](https://grpc.io/docs/)
- [HTTP/2 Specification](https://http2.github.io/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
