# REST API Implementation

## Overview

This document describes the REST API support implementation for the Rockstar Web Framework, completed as part of Task 14.

## Requirements Satisfied

- **Requirement 2.2**: REST protocol access support
- **Requirement 2.6**: Rate limiting per resource or globally
- **Requirement 2.7**: Rate limiting storage in database (MySQL, PostgreSQL, MSSQL, SQLite)

## Components Implemented

### 1. REST API Manager (`pkg/rest_api.go`)

Core interfaces and types for REST API functionality:

- **RESTAPIManager**: Main interface for REST API management
  - Route registration with configuration
  - Rate limiting (per-resource and global)
  - JSON request/response handling
  - Middleware support
  - Route grouping

- **RESTHandler**: Handler function type for REST endpoints
- **RESTMiddleware**: Middleware function type for REST APIs
- **RESTRouteConfig**: Configuration for REST routes including:
  - Rate limiting configuration
  - Authentication requirements
  - Request validation (size, timeout)
  - CORS configuration

- **RESTRateLimitConfig**: Rate limiting configuration
  - Limit: Maximum number of requests
  - Window: Time window for the limit
  - Key: Rate limit key (user_id, ip_address, tenant_id)

- **RESTError**: Structured error responses
- **RESTResponse**: Standard response format with success/error/meta
- **Pagination**: Pagination information for list endpoints
- **RateLimitInfo**: Rate limit information in responses

### 2. REST API Implementation (`pkg/rest_api_impl.go`)

Implementation of the REST API manager:

- **restAPIManager**: Concrete implementation
  - Route registration for all HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
  - Middleware chain execution
  - Rate limiting with database storage
  - Authentication and authorization middleware
  - Request validation middleware
  - CORS middleware
  - Route grouping with prefix support

- **rateLimiter**: Database-backed rate limiter
  - Check rate limits
  - Increment counters
  - Support for multiple rate limit keys

### 3. Unit Tests (`pkg/rest_api_test.go`)

Comprehensive test coverage:

- JSON parsing and serialization tests
- REST error handling tests
- Rate limit configuration tests
- REST route configuration tests
- Response structure tests
- Pagination tests
- Rate limit info tests
- REST API manager tests
- Rate limiter tests with database integration
- Mock implementations for testing

All tests pass successfully.

### 4. Example Application (`examples/rest_api_example.go`)

Demonstration of REST API features:

- User CRUD operations (Create, Read, Update, Delete)
- Rate limiting per IP address
- API versioning with different rate limits
- CORS configuration
- JSON request/response handling
- Error handling
- Health check endpoint

## Features

### JSON Request/Response Handling

- Automatic JSON parsing from request body
- Structured JSON responses with success/error/meta
- Support for custom response metadata
- Pagination support
- Rate limit information in responses

### Rate Limiting

- Per-resource rate limiting
- Global rate limiting
- Database-backed storage (MySQL, PostgreSQL, MSSQL, SQLite)
- Configurable limits and time windows
- Multiple rate limit keys:
  - user_id: Per authenticated user
  - tenant_id: Per tenant (multi-tenancy)
  - ip_address: Per client IP

### Middleware System

- Authentication middleware
- Authorization middleware with scope checking
- Request validation (size, timeout)
- CORS middleware
- Rate limiting middleware
- Custom middleware support

### Route Configuration

Each route can be configured with:
- Rate limiting (per-resource and global)
- Authentication requirements
- Required scopes for authorization
- Maximum request size
- Request timeout
- CORS settings
- Cache control headers

### Route Grouping

- Create route groups with common prefix
- Apply middleware to entire groups
- Different configurations per group
- Support for API versioning

### Error Handling

Predefined error types:
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 405 Method Not Allowed
- 409 Conflict
- 429 Rate Limit Exceeded
- 500 Internal Server Error
- 503 Service Unavailable

Custom errors with:
- Error code
- Error message
- Additional details
- HTTP status code

## Usage Example

```go
// Create REST API manager
restAPI := pkg.NewRESTAPIManager(router, db)

// Configure route
config := pkg.RESTRouteConfig{
    RateLimit: &pkg.RESTRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
    RequireAuth:    false,
    MaxRequestSize: 1024 * 1024,
    Timeout:        30 * time.Second,
    CORS: &pkg.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST"},
    },
}

// Register route
restAPI.RegisterRoute("GET", "/api/users", func(ctx pkg.Context) error {
    users := []User{...}
    return restAPI.SendJSONResponse(ctx, 200, users)
}, config)

// Create API group
apiV2 := restAPI.Group("/api/v2")
apiV2.RegisterRoute("GET", "/users", handler, config)
```

## Testing

All unit tests pass:
- 11 test suites
- 30+ individual test cases
- Mock implementations for router and database
- Benchmark tests for performance

Run tests:
```bash
go test -v ./pkg -run "TestREST|TestParseJSON|TestToJSON|TestRateLimiter"
```

## Integration with Framework

The REST API implementation integrates seamlessly with:
- Router engine for route registration
- Database manager for rate limiting storage
- Context for request/response handling
- Security manager for authentication/authorization
- Session manager for user sessions
- Multi-tenancy support

## Performance Considerations

- Efficient JSON parsing and serialization
- Database-backed rate limiting for distributed deployments
- Middleware chain optimization
- Connection pooling for database operations
- Minimal memory allocations

## Future Enhancements

Potential improvements:
- Response caching
- Request/response compression
- API documentation generation (OpenAPI/Swagger)
- Request/response logging
- Metrics collection
- Circuit breaker for external dependencies
- Request throttling
- API key management

## Conclusion

The REST API implementation provides a complete, production-ready solution for building REST APIs with the Rockstar Web Framework. It includes all essential features like rate limiting, authentication, CORS, and comprehensive error handling, while maintaining high performance and ease of use.
