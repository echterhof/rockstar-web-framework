# GraphQL Implementation

## Overview

The Rockstar Web Framework provides comprehensive GraphQL support with built-in authentication, authorization, rate limiting, and middleware capabilities. The implementation follows the framework's design principles of modularity, performance, and security.

## Features

- **Schema Registration**: Register GraphQL schemas with flexible configuration
- **Authentication Integration**: OAuth2, JWT, and access token authentication
- **Authorization**: Role-based and scope-based access control
- **Rate Limiting**: Per-resource and global rate limiting with database storage
- **Middleware Support**: Custom middleware for request/response processing
- **CORS Support**: Built-in CORS configuration
- **GraphQL Playground**: Optional GraphQL Playground for development
- **Introspection**: Optional schema introspection support
- **Request Validation**: Size limits, timeouts, and complexity controls
- **Route Grouping**: Organize endpoints with prefixes and shared middleware

## Requirements Satisfied

- **Requirement 2.1**: GraphQL protocol access support
- **Requirement 2.5**: Access token validation with database storage
- **Requirement 2.6**: Rate limiting per resource and globally with database storage

## Architecture

### Core Components

1. **GraphQLManager**: Main interface for GraphQL functionality
2. **GraphQLSchema**: Interface for schema implementations
3. **GraphQLMiddleware**: Middleware for request processing
4. **GraphQLConfig**: Configuration for endpoints

### Request Flow

```
Client Request
    ↓
CORS Middleware (if configured)
    ↓
Rate Limiting (if configured)
    ↓
Authentication (if required)
    ↓
Authorization (if required)
    ↓
Request Validation
    ↓
Custom Middleware
    ↓
Schema Execution
    ↓
Response
```

## Usage

### Basic Setup

```go
import "github.com/rockstar-web-framework/framework/pkg"

// Create router and dependencies
router := pkg.NewRouter()
db := pkg.NewMockDatabaseManager()
authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

// Create GraphQL manager
graphqlManager := pkg.NewGraphQLManager(router, db, authManager)
```

### Implementing a GraphQL Schema

```go
type MySchema struct {
    // Your schema data
}

// Implement the GraphQLSchema interface
func (s *MySchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    // Parse and execute the query
    // Return the result
    return result, nil
}
```

### Registering a Schema

```go
schema := &MySchema{}

config := pkg.GraphQLConfig{
    EnableIntrospection: true,
    EnablePlayground:    true,
    MaxRequestSize:      1024 * 1024, // 1MB
    Timeout:             30 * time.Second,
}

err := graphqlManager.RegisterSchema("/graphql", schema, config)
```

### Authentication and Authorization

```go
config := pkg.GraphQLConfig{
    RequireAuth:    true,
    RequiredRoles:  []string{"admin", "user"},
    RequiredScopes: []string{"read:data", "write:data"},
}

err := graphqlManager.RegisterSchema("/api/graphql", schema, config)
```

### Rate Limiting

```go
config := pkg.GraphQLConfig{
    RateLimit: &pkg.GraphQLRateLimitConfig{
        Limit:  100,           // 100 requests
        Window: time.Minute,   // per minute
        Key:    "user_id",     // per user
    },
    GlobalRateLimit: &pkg.GraphQLRateLimitConfig{
        Limit:  1000,
        Window: time.Minute,
        Key:    "tenant_id",
    },
}
```

### Custom Middleware

```go
// Add logging middleware
graphqlManager.Use(func(ctx pkg.Context, next pkg.GraphQLHandler) error {
    log.Printf("GraphQL request: %s", ctx.Request().URL.Path)
    err := next(ctx)
    log.Printf("GraphQL response: %v", err)
    return err
})

// Add timing middleware
graphqlManager.Use(func(ctx pkg.Context, next pkg.GraphQLHandler) error {
    start := time.Now()
    err := next(ctx)
    duration := time.Since(start)
    log.Printf("Query took: %v", duration)
    return err
})
```

### Route Grouping

```go
// Create API v1 group
apiV1 := graphqlManager.Group("/api/v1")

// Add group-specific middleware
apiV1.Use(func(ctx pkg.Context, next pkg.GraphQLHandler) error {
    // Group middleware logic
    return next(ctx)
})

// Register schema in group
err := apiV1.RegisterSchema("/graphql", schema, config)
// This creates endpoint at /api/v1/graphql
```

### CORS Configuration

```go
config := pkg.GraphQLConfig{
    CORS: &pkg.CORSConfig{
        AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
        AllowMethods:     []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           3600,
    },
}
```

## GraphQL Request Format

### POST Request

```json
{
  "query": "query GetUser($id: ID!) { user(id: $id) { id name email } }",
  "variables": {
    "id": "123"
  },
  "operationName": "GetUser"
}
```

### GET Request (for introspection)

```
GET /graphql?query={hello}
```

## GraphQL Response Format

### Success Response

```json
{
  "data": {
    "user": {
      "id": "123",
      "name": "John Doe",
      "email": "john@example.com"
    }
  },
  "extensions": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req-123"
  }
}
```

### Error Response

```json
{
  "errors": [
    {
      "message": "User not found",
      "path": ["user"],
      "locations": [
        {
          "line": 1,
          "column": 10
        }
      ],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ],
  "extensions": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req-123"
  }
}
```

## Configuration Options

### GraphQLConfig

| Field | Type | Description |
|-------|------|-------------|
| `RequireAuth` | bool | Require authentication for this endpoint |
| `RequiredRoles` | []string | Required user roles for authorization |
| `RequiredScopes` | []string | Required access token scopes |
| `MaxRequestSize` | int64 | Maximum request body size in bytes |
| `MaxQueryDepth` | int | Maximum query depth (requires schema support) |
| `MaxComplexity` | int | Maximum query complexity (requires schema support) |
| `Timeout` | time.Duration | Request timeout |
| `EnableIntrospection` | bool | Enable schema introspection |
| `EnablePlayground` | bool | Enable GraphQL Playground UI |
| `RateLimit` | *GraphQLRateLimitConfig | Per-resource rate limiting |
| `GlobalRateLimit` | *GraphQLRateLimitConfig | Global rate limiting |
| `CORS` | *CORSConfig | CORS configuration |

### GraphQLRateLimitConfig

| Field | Type | Description |
|-------|------|-------------|
| `Limit` | int | Maximum number of requests |
| `Window` | time.Duration | Time window for the limit |
| `Key` | string | Rate limit key type: "user_id", "tenant_id", "ip_address" |

## Error Handling

The framework provides standard GraphQL errors:

- `ErrGraphQLSyntax`: Syntax error in query
- `ErrGraphQLValidation`: Validation error
- `ErrGraphQLExecution`: Execution error
- `ErrGraphQLAuthentication`: Authentication required
- `ErrGraphQLAuthorization`: Insufficient permissions
- `ErrGraphQLRateLimit`: Rate limit exceeded
- `ErrGraphQLComplexity`: Query complexity exceeds limit
- `ErrGraphQLDepth`: Query depth exceeds limit
- `ErrGraphQLTimeout`: Query execution timeout

## Best Practices

1. **Use Authentication**: Always require authentication for sensitive data
2. **Implement Rate Limiting**: Protect your API from abuse
3. **Set Request Limits**: Configure `MaxRequestSize` and `Timeout`
4. **Disable Introspection in Production**: Set `EnableIntrospection: false`
5. **Disable Playground in Production**: Set `EnablePlayground: false`
6. **Use CORS Properly**: Configure allowed origins explicitly
7. **Implement Query Complexity**: Prevent expensive queries
8. **Use Middleware**: Add logging, monitoring, and custom logic
9. **Group Related Endpoints**: Use route groups for organization
10. **Handle Errors Gracefully**: Return meaningful error messages

## Integration with Other Framework Features

### Database Integration

Rate limiting data is automatically stored in the configured database:

```go
db := pkg.NewDatabaseManager()
db.Connect(pkg.DatabaseConfig{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    Username: "user",
    Password: "pass",
})

graphqlManager := pkg.NewGraphQLManager(router, db, authManager)
```

### Session Management

Access user session data in your schema:

```go
func (s *MySchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    // Access session through context
    session := ctx.Session()
    userID := session.Get("user_id")
    
    // Use session data in query execution
    return result, nil
}
```

### Multi-Tenancy

GraphQL endpoints support multi-tenancy through the framework's tenant system:

```go
// Rate limiting per tenant
config := pkg.GraphQLConfig{
    RateLimit: &pkg.GraphQLRateLimitConfig{
        Key: "tenant_id",
        Limit: 1000,
        Window: time.Hour,
    },
}
```

## Performance Considerations

1. **Connection Pooling**: Database connections are pooled automatically
2. **Rate Limit Caching**: Rate limit counters are cached for performance
3. **Middleware Ordering**: Middleware is executed in registration order
4. **Request Validation**: Early validation prevents unnecessary processing
5. **Context Reuse**: Context objects are pooled and reused

## Testing

The framework includes comprehensive tests for GraphQL functionality:

```bash
go test -v -run TestGraphQL ./pkg
```

## Example Application

See `examples/graphql_example.go` for a complete working example.

## Future Enhancements

- Subscription support via WebSockets
- Automatic schema generation from Go types
- Query complexity analysis
- Persistent query support
- Batch query support
- DataLoader integration for N+1 query prevention

## References

- [GraphQL Specification](https://spec.graphql.org/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- Framework Requirements: 2.1, 2.5, 2.6
