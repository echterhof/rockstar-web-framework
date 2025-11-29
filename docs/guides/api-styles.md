# API Styles

## Overview

The Rockstar Web Framework supports multiple API architectural styles, allowing you to build APIs that match your specific requirements and client needs. Whether you're building a traditional REST API, a modern GraphQL service, a high-performance gRPC API, or maintaining legacy SOAP services, the framework provides comprehensive support for each style.

This guide covers the setup, configuration, and best practices for each API style, along with examples and error handling strategies.

## Supported API Styles

### REST (Representational State Transfer)

REST is the most common API style, using HTTP methods and URLs to represent resources and operations.

**Best for:**
- Public APIs with broad client support
- CRUD operations on resources
- Caching and HTTP semantics
- Simple, stateless interactions

### GraphQL

GraphQL provides a query language for APIs, allowing clients to request exactly the data they need.

**Best for:**
- Complex data requirements
- Mobile applications with bandwidth constraints
- Reducing over-fetching and under-fetching
- Rapid frontend development

### gRPC

gRPC is a high-performance RPC framework using Protocol Buffers for serialization.

**Best for:**
- Microservice communication
- High-performance requirements
- Strongly-typed contracts
- Streaming data

### SOAP

SOAP is an XML-based protocol with formal contracts defined in WSDL.

**Best for:**
- Enterprise integrations
- Legacy system compatibility
- Formal contracts and validation
- Transaction support

## REST API

### Setup and Configuration

Create a REST API with the framework:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        EnableHTTP1:  true,
        EnableHTTP2:  true,
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

router := app.Router()
db := app.Database()

// Create REST API manager
restAPI := pkg.NewRESTAPIManager(router, db)
```

### Route Configuration

Configure REST routes with rate limiting and CORS:

```go
apiConfig := pkg.RESTRouteConfig{
    RateLimit: &pkg.RESTRateLimitConfig{
        Limit:  100,         // 100 requests
        Window: time.Minute, // per minute
        Key:    "ip_address",
    },
    MaxRequestSize: 1024 * 1024, // 1MB max request size
    Timeout:        30 * time.Second,
    CORS: &pkg.CORSConfig{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: false,
        MaxAge:           3600,
    },
}
```

### RESTful Patterns

Implement standard REST patterns:

```go
// List resources with pagination
restAPI.RegisterRoute("GET", "/api/products", func(ctx pkg.Context) error {
    query := ctx.Query()
    
    // Parse pagination
    page := 1
    perPage := 10
    if pageStr := query["page"]; pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }
    
    // Fetch data
    products, total, err := fetchProducts(page, perPage)
    if err != nil {
        return restAPI.SendErrorResponse(ctx, 500, "Failed to fetch products", nil)
    }
    
    // Return with pagination metadata
    return restAPI.SendJSONResponse(ctx, 200, map[string]interface{}{
        "products": products,
        "pagination": map[string]interface{}{
            "page":        page,
            "per_page":    perPage,
            "total":       total,
            "total_pages": (total + perPage - 1) / perPage,
        },
    })
}, apiConfig)

// Get single resource
restAPI.RegisterRoute("GET", "/api/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    product, err := fetchProduct(id)
    if err != nil {
        return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
            "id": id,
        })
    }
    
    return restAPI.SendJSONResponse(ctx, 200, product)
}, apiConfig)

// Create resource
restAPI.RegisterRoute("POST", "/api/products", func(ctx pkg.Context) error {
    var input struct {
        Name  string  `json:"name"`
        Price float64 `json:"price"`
    }
    
    if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
        return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", nil)
    }
    
    // Validate
    if input.Name == "" {
        return restAPI.SendErrorResponse(ctx, 400, "Name is required", nil)
    }
    
    // Create
    product, err := createProduct(input.Name, input.Price)
    if err != nil {
        return restAPI.SendErrorResponse(ctx, 500, "Failed to create product", nil)
    }
    
    return restAPI.SendJSONResponse(ctx, 201, product)
}, apiConfig)

// Update resource (full update)
restAPI.RegisterRoute("PUT", "/api/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    var input struct {
        Name  string  `json:"name"`
        Price float64 `json:"price"`
    }
    
    if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
        return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", nil)
    }
    
    product, err := updateProduct(id, input.Name, input.Price)
    if err != nil {
        return restAPI.SendErrorResponse(ctx, 404, "Product not found", nil)
    }
    
    return restAPI.SendJSONResponse(ctx, 200, product)
}, apiConfig)

// Partial update
restAPI.RegisterRoute("PATCH", "/api/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    var input map[string]interface{}
    if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
        return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", nil)
    }
    
    product, err := patchProduct(id, input)
    if err != nil {
        return restAPI.SendErrorResponse(ctx, 404, "Product not found", nil)
    }
    
    return restAPI.SendJSONResponse(ctx, 200, product)
}, apiConfig)

// Delete resource
restAPI.RegisterRoute("DELETE", "/api/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    if err := deleteProduct(id); err != nil {
        return restAPI.SendErrorResponse(ctx, 404, "Product not found", nil)
    }
    
    return restAPI.SendJSONResponse(ctx, 200, map[string]interface{}{
        "deleted": true,
        "id":      id,
    })
}, apiConfig)
```

### REST Error Handling

Implement consistent error responses:

```go
// Standard error response format
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Code    int                    `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Use the REST API manager's error response helper
return restAPI.SendErrorResponse(ctx, 400, "Validation failed", map[string]interface{}{
    "field": "email",
    "issue": "invalid format",
})
```

### REST Best Practices

**Use proper HTTP methods:**
- GET: Retrieve resources (idempotent, cacheable)
- POST: Create resources
- PUT: Full update (idempotent)
- PATCH: Partial update
- DELETE: Remove resources (idempotent)

**Use appropriate status codes:**
- 200: Success
- 201: Created
- 204: No Content
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 409: Conflict
- 422: Unprocessable Entity
- 500: Internal Server Error

**Version your API:**
```go
// URL versioning
restAPI.RegisterRoute("GET", "/api/v1/products", handlerV1, apiConfig)
restAPI.RegisterRoute("GET", "/api/v2/products", handlerV2, apiConfig)

// Header versioning
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    version := ctx.GetHeader("API-Version")
    if version == "" {
        version = "v1"
    }
    ctx.Set("api_version", version)
    return next(ctx)
})
```

## GraphQL API

### Setup and Configuration

Create a GraphQL API:

```go
router := app.Router()
db := app.Database()
authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

// Create GraphQL manager
graphqlManager := pkg.NewGraphQLManager(router, db, authManager)

// Configure GraphQL endpoint
graphqlConfig := pkg.GraphQLConfig{
    EnableIntrospection: true,
    EnablePlayground:    true,
    MaxRequestSize:      1024 * 1024, // 1MB
    Timeout:             30 * time.Second,
    RateLimit: &pkg.GraphQLRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
}
```

### Schema Definition

Define your GraphQL schema:

```go
type BlogSchema struct{}

func (s *BlogSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    // Parse and execute query
    // This is simplified - use a proper GraphQL library in production
    
    if strings.Contains(query, "users") {
        return s.queryUsers()
    }
    
    if strings.Contains(query, "createUser") {
        name := extractArg(query, "name")
        email := extractArg(query, "email")
        return s.mutationCreateUser(name, email)
    }
    
    return nil, fmt.Errorf("unsupported query")
}

func (s *BlogSchema) queryUsers() (interface{}, error) {
    users := fetchAllUsers()
    return map[string]interface{}{
        "users": users,
    }, nil
}

func (s *BlogSchema) mutationCreateUser(name, email string) (interface{}, error) {
    if name == "" || email == "" {
        return nil, fmt.Errorf("name and email are required")
    }
    
    user := createUser(name, email)
    return map[string]interface{}{
        "createUser": user,
    }, nil
}
```

### Register Schema

Register the schema with the GraphQL manager:

```go
schema := &BlogSchema{}
err = graphqlManager.RegisterSchema("/graphql", schema, graphqlConfig)
if err != nil {
    log.Fatal(err)
}
```

### GraphQL Queries

Example queries:

```graphql
# List all users
{
  users {
    id
    name
    email
    role
  }
}

# Get user by ID
{
  user(id: "1") {
    id
    name
    email
    posts {
      id
      title
    }
  }
}

# Create user (mutation)
mutation {
  createUser(name: "Alice", email: "alice@example.com") {
    id
    name
    email
  }
}
```

### GraphQL Error Handling

Return errors in GraphQL format:

```go
func (s *BlogSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    result, err := s.executeQuery(query, variables)
    if err != nil {
        // Return GraphQL error format
        return map[string]interface{}{
            "data":   nil,
            "errors": []map[string]interface{}{
                {
                    "message": err.Error(),
                    "path":    []string{"query"},
                },
            },
        }, nil
    }
    
    return map[string]interface{}{
        "data": result,
    }, nil
}
```

### GraphQL Best Practices

**Use proper GraphQL libraries:**
- `github.com/graphql-go/graphql` - GraphQL implementation
- `github.com/99designs/gqlgen` - Code generation from schema

**Implement DataLoader for N+1 queries:**
```go
// Batch load related data to avoid N+1 queries
type UserLoader struct {
    cache map[string]*User
}

func (l *UserLoader) Load(id string) (*User, error) {
    if user, ok := l.cache[id]; ok {
        return user, nil
    }
    // Batch load users
    return fetchUser(id)
}
```

**Add query complexity limits:**
```go
graphqlConfig := pkg.GraphQLConfig{
    MaxComplexity: 1000, // Limit query complexity
    MaxDepth:      10,   // Limit query depth
}
```

## gRPC API

### Setup and Configuration

Create a gRPC service:

```go
router := app.Router()
db := app.Database()
authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

// Create gRPC manager
grpcManager := pkg.NewGRPCManager(router, db, authManager)

// Add middleware
grpcManager.Use(func(ctx pkg.Context, next pkg.GRPCHandler) error {
    start := time.Now()
    err := next(ctx)
    log.Printf("gRPC call took %v", time.Since(start))
    return err
})
```

### Service Implementation

Implement a gRPC service:

```go
type UserService struct {
    users map[string]*User
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) Methods() []string {
    return []string{"GetUser", "CreateUser", "ListUsers"}
}

func (s *UserService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
    switch method {
    case "GetUser":
        return s.getUser(ctx, req)
    case "CreateUser":
        return s.createUser(ctx, req)
    case "ListUsers":
        return s.listUsers(ctx, req)
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}

func (s *UserService) getUser(ctx context.Context, req interface{}) (interface{}, error) {
    reqMap := req.(map[string]interface{})
    id := reqMap["id"].(string)
    
    user, exists := s.users[id]
    if !exists {
        return nil, fmt.Errorf("user not found: %s", id)
    }
    
    return map[string]interface{}{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
    }, nil
}
```

### Register Service

Register the service with configuration:

```go
userService := NewUserService()

err = grpcManager.RegisterService(userService, pkg.GRPCConfig{
    RequireAuth: false,
    RateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
    MaxRequestSize: 1024 * 1024,
    Timeout:        30 * time.Second,
})
if err != nil {
    log.Fatal(err)
}
```

### gRPC Error Handling

Return gRPC-style errors:

```go
func (s *UserService) getUser(ctx context.Context, req interface{}) (interface{}, error) {
    id := req.(map[string]interface{})["id"].(string)
    
    user, exists := s.users[id]
    if !exists {
        // Return error with gRPC status code
        return nil, status.Errorf(codes.NotFound, "user not found: %s", id)
    }
    
    return user, nil
}
```

### gRPC Best Practices

**Use Protocol Buffers:**
```protobuf
syntax = "proto3";

package user;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (stream User);
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

**Generate code:**
```bash
protoc --go_out=. --go-grpc_out=. user.proto
```

**Implement streaming:**
```go
func (s *UserService) HandleStream(stream pkg.GRPCServerStream, method string) error {
    switch method {
    case "ListUsers":
        for _, user := range s.users {
            if err := stream.Send(user); err != nil {
                return err
            }
        }
        return nil
    default:
        return fmt.Errorf("unknown streaming method: %s", method)
    }
}
```

## SOAP API

### Setup and Configuration

Create a SOAP service:

```go
router := app.Router()
db := app.Database()
authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

// Create SOAP manager
soapManager := pkg.NewSOAPManager(router, db, authManager)
```

### Service Implementation

Implement a SOAP service:

```go
type UserService struct {
    users map[string]*User
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) WSDL() (string, error) {
    config := pkg.SOAPConfig{
        ServiceName: "UserService",
        Namespace:   "http://example.com/soap/user",
        PortName:    "UserServicePort",
    }
    
    operations := []pkg.WSDLOperation{
        {
            Name:        "GetUser",
            InputType:   "GetUserRequest",
            OutputType:  "GetUserResponse",
            Description: "Retrieves a user by ID",
        },
    }
    
    return pkg.GenerateWSDL(config, "http://localhost:8080/soap/user", operations)
}

func (s *UserService) Execute(action string, body []byte) ([]byte, error) {
    switch action {
    case "GetUser":
        return s.handleGetUser(body)
    default:
        return nil, fmt.Errorf("unknown operation: %s", action)
    }
}

func (s *UserService) handleGetUser(body []byte) ([]byte, error) {
    var req GetUserRequest
    if err := xml.Unmarshal(body, &req); err != nil {
        return nil, err
    }
    
    user, exists := s.users[req.UserID]
    if !exists {
        return nil, fmt.Errorf("user not found")
    }
    
    response := GetUserResponse{User: user}
    return xml.MarshalIndent(response, "", "  ")
}
```

### Register Service

Register the SOAP service:

```go
userService := NewUserService()

soapConfig := pkg.SOAPConfig{
    EnableWSDL:  true,
    ServiceName: "UserService",
    Namespace:   "http://example.com/soap/user",
    PortName:    "UserServicePort",
    RateLimit: &pkg.SOAPRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
    MaxRequestSize: 1024 * 1024,
    Timeout:        30 * time.Second,
}

err = soapManager.RegisterService("/soap/user", userService, soapConfig)
if err != nil {
    log.Fatal(err)
}
```

### SOAP Requests

Example SOAP request:

```xml
POST /soap/user HTTP/1.1
Host: localhost:8080
Content-Type: text/xml; charset=utf-8
SOAPAction: "GetUser"

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUser>
      <UserID>1</UserID>
    </GetUser>
  </soap:Body>
</soap:Envelope>
```

### SOAP Error Handling

Return SOAP faults:

```go
func (s *UserService) handleGetUser(body []byte) ([]byte, error) {
    var req GetUserRequest
    if err := xml.Unmarshal(body, &req); err != nil {
        // Return SOAP fault
        fault := pkg.SOAPFault{
            Code:   "Client",
            String: "Invalid request",
            Detail: err.Error(),
        }
        return xml.MarshalIndent(fault, "", "  ")
    }
    
    // ... rest of implementation
}
```

### SOAP Best Practices

**Provide WSDL:**
- Always provide WSDL at `?wsdl` endpoint
- Keep WSDL up to date with implementation

**Use proper namespaces:**
```go
config := pkg.SOAPConfig{
    Namespace: "http://example.com/soap/user",
}
```

**Validate against schema:**
- Use XML schema validation
- Return proper SOAP faults for validation errors

## Choosing an API Style

### Decision Matrix

| Requirement | REST | GraphQL | gRPC | SOAP |
|------------|------|---------|------|------|
| Public API | ✅ Best | ✅ Good | ❌ Poor | ❌ Poor |
| Mobile clients | ✅ Good | ✅ Best | ✅ Good | ❌ Poor |
| Microservices | ✅ Good | ❌ Poor | ✅ Best | ❌ Poor |
| Legacy systems | ✅ Good | ❌ Poor | ❌ Poor | ✅ Best |
| Real-time | ❌ Poor | ✅ Good | ✅ Best | ❌ Poor |
| Caching | ✅ Best | ❌ Poor | ❌ Poor | ❌ Poor |
| Type safety | ❌ Poor | ✅ Good | ✅ Best | ✅ Good |
| Learning curve | ✅ Easy | ⚠️ Medium | ⚠️ Medium | ❌ Hard |

### Hybrid Approaches

You can support multiple API styles simultaneously:

```go
// REST API
restAPI := pkg.NewRESTAPIManager(router, db)
restAPI.RegisterRoute("GET", "/api/products", restHandler, restConfig)

// GraphQL API
graphqlManager := pkg.NewGraphQLManager(router, db, authManager)
graphqlManager.RegisterSchema("/graphql", schema, graphqlConfig)

// gRPC API
grpcManager := pkg.NewGRPCManager(router, db, authManager)
grpcManager.RegisterService(grpcService, grpcConfig)
```

## Complete Examples

See the complete examples in the `examples/` directory:

- `examples/rest_api_example.go` - REST API with CRUD operations
- `examples/graphql_example.go` - GraphQL API with queries and mutations
- `examples/grpc_example.go` - gRPC service implementation
- `examples/soap_example.go` - SOAP service with WSDL

## See Also

- [Routing Guide](routing.md) - Route configuration
- [Security Guide](security.md) - API authentication and authorization
- [Database Guide](database.md) - Data persistence
- [Performance Guide](performance.md) - API optimization
