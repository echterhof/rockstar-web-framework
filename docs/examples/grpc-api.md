# gRPC API Example

The gRPC API example (`examples/grpc_example.go`) demonstrates how to build gRPC services with the Rockstar Web Framework. It showcases service registration, unary RPCs, authentication, and rate limiting.

## What This Example Demonstrates

- **gRPC service** implementation
- **Unary RPC** methods
- **Service registration** with the framework
- **Authentication** for protected services
- **Rate limiting** per service
- **Error handling** with gRPC status codes
- **Middleware** for logging and metrics

## Prerequisites

- Go 1.25 or higher
- HTTP/2 support (enabled by default)

## Setup Instructions

```bash
go run examples/grpc_example.go
```

The server will start on `http://localhost:8080`.

## Services

### UserService (Public)

| Method | Description |
|--------|-------------|
| GetUser | Retrieve user by ID |
| CreateUser | Create new user |
| UpdateUser | Update existing user |
| DeleteUser | Delete user |
| ListUsers | List all users |

### AdminService (Requires Authentication)

| Method | Description |
|--------|-------------|
| GetSystemStats | Get system statistics |
| ResetDatabase | Reset database (admin only) |

## Testing the API

### List Users

```bash
curl -X POST http://localhost:8080/UserService/ListUsers \
  -H 'Content-Type: application/json' \
  -d '{}'
```

### Get User

```bash
curl -X POST http://localhost:8080/UserService/GetUser \
  -H 'Content-Type: application/json' \
  -d '{"id": "1"}'
```

### Create User

```bash
curl -X POST http://localhost:8080/UserService/CreateUser \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Alice Johnson",
    "email": "alice@example.com",
    "role": "user"
  }'
```

### Update User

```bash
curl -X POST http://localhost:8080/UserService/UpdateUser \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "1",
    "name": "Alice Updated",
    "email": "alice.updated@example.com"
  }'
```

### Delete User

```bash
curl -X POST http://localhost:8080/UserService/DeleteUser \
  -H 'Content-Type: application/json' \
  -d '{"id": "1"}'
```

### Get System Stats (Admin)

```bash
curl -X POST http://localhost:8080/AdminService/GetSystemStats \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin_token>' \
  -d '{}'
```

## Code Walkthrough

### Service Implementation

```go
type UserService struct {
    users map[string]*User
}

func (s *UserService) ServiceName() string {
    return "UserService"
}

func (s *UserService) Methods() []string {
    return []string{"GetUser", "CreateUser", "UpdateUser", "DeleteUser", "ListUsers"}
}

func (s *UserService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
    switch method {
    case "GetUser":
        return s.getUser(ctx, req)
    case "CreateUser":
        return s.createUser(ctx, req)
    // ... other methods
    }
}
```

### Service Registration

```go
grpcManager := pkg.NewGRPCManager(router, db, authManager)

// Public service with rate limiting
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

// Protected admin service
err = grpcManager.RegisterService(adminService, pkg.GRPCConfig{
    RequireAuth:   true,
    RequiredRoles: []string{"admin"},
    RateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  50,
        Window: time.Minute,
        Key:    "user_id",
    },
})
```

### Middleware

```go
grpcManager.Use(func(ctx pkg.Context, next pkg.GRPCHandler) error {
    start := time.Now()
    log.Printf("gRPC request started: %s", ctx.Request().URL.Path)
    
    err := next(ctx)
    
    duration := time.Since(start)
    log.Printf("gRPC request completed (duration: %v)", duration)
    
    return err
})
```

## Production Considerations

### Use Protocol Buffers

Define services with `.proto` files:

```protobuf
syntax = "proto3";
package user;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
}
```

Generate Go code:

```bash
protoc --go_out=. --go-grpc_out=. user.proto
```

### Add TLS

Enable TLS for production:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP2: true,
        TLSCertFile: "cert.pem",
        TLSKeyFile:  "key.pem",
    },
}
```

### Implement Streaming

Add streaming support:

```go
func (s *UserService) HandleStream(stream pkg.GRPCServerStream, method string) error {
    switch method {
    case "StreamUsers":
        return s.streamUsers(stream)
    }
}

func (s *UserService) streamUsers(stream pkg.GRPCServerStream) error {
    for _, user := range s.users {
        if err := stream.Send(user); err != nil {
            return err
        }
    }
    return nil
}
```

## Related Documentation

- [API Styles Guide](../guides/api-styles.md) - REST, GraphQL, gRPC, SOAP
- [Protocols Guide](../guides/protocols.md) - HTTP/1, HTTP/2, QUIC
- [Security Guide](../guides/security.md) - Authentication and authorization

## Source Code

The complete source code is available at `examples/grpc_example.go` in the repository.
