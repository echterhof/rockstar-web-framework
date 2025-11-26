package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// 🎸 gRPC Service Example
// This example demonstrates gRPC service implementation with the Rockstar Web Framework
// Features: Service registration, unary RPCs, streaming support, error handling
//
// Note: This example shows the gRPC service structure using the framework's gRPC support.
// In a production environment, you would typically define your service using Protocol Buffers (.proto files)
// and generate the Go code using protoc. For this example, we're using a simplified approach.
//
// Example .proto file structure:
//
// syntax = "proto3";
// package user;
//
// service UserService {
//   rpc GetUser(GetUserRequest) returns (GetUserResponse);
//   rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
//   rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
//   rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
// }
//
// message GetUserRequest {
//   string id = 1;
// }
//
// message GetUserResponse {
//   string id = 1;
//   string name = 2;
//   string email = 3;
//   string role = 4;
// }

// User represents a user entity
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// In-memory storage for demonstration
var (
	users     = make(map[string]*User)
	userIDSeq = 1
)

// UserService implements a gRPC user service
type UserService struct {
	users map[string]*User
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: users,
	}
}

// ServiceName returns the service name
func (s *UserService) ServiceName() string {
	return "UserService"
}

// Methods returns the list of available methods
func (s *UserService) Methods() []string {
	return []string{"GetUser", "CreateUser", "UpdateUser", "DeleteUser", "ListUsers"}
}

// HandleUnary handles unary RPC calls
func (s *UserService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
	switch method {
	case "GetUser":
		return s.getUser(ctx, req)
	case "CreateUser":
		return s.createUser(ctx, req)
	case "UpdateUser":
		return s.updateUser(ctx, req)
	case "DeleteUser":
		return s.deleteUser(ctx, req)
	case "ListUsers":
		return s.listUsers(ctx, req)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// HandleStream handles streaming RPC calls
func (s *UserService) HandleStream(stream pkg.GRPCServerStream, method string) error {
	// Streaming not implemented in this example
	// In a real implementation, you would handle client/server/bidirectional streaming here
	return fmt.Errorf("streaming not implemented for method: %s", method)
}

// GetMethodDescriptor returns method descriptor
func (s *UserService) GetMethodDescriptor(method string) *pkg.GRPCMethodDescriptor {
	return &pkg.GRPCMethodDescriptor{
		Name:           method,
		IsClientStream: false,
		IsServerStream: false,
	}
}

// RPC method implementations

func (s *UserService) getUser(ctx context.Context, req interface{}) (interface{}, error) {
	// In a real implementation, req would be a protobuf message
	// For this example, we'll parse it as a map
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	id, ok := reqMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid user id")
	}

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}, nil
}

func (s *UserService) createUser(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	name, _ := reqMap["name"].(string)
	email, _ := reqMap["email"].(string)
	role, _ := reqMap["role"].(string)

	if name == "" || email == "" {
		return nil, fmt.Errorf("name and email are required")
	}

	if role == "" {
		role = "user"
	}

	id := fmt.Sprintf("%d", userIDSeq)
	user := &User{
		ID:    id,
		Name:  name,
		Email: email,
		Role:  role,
	}
	s.users[id] = user
	userIDSeq++

	return map[string]interface{}{
		"id":      user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"role":    user.Role,
		"created": true,
	}, nil
}

func (s *UserService) updateUser(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	id, _ := reqMap["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	// Update fields if provided
	if name, ok := reqMap["name"].(string); ok && name != "" {
		user.Name = name
	}
	if email, ok := reqMap["email"].(string); ok && email != "" {
		user.Email = email
	}
	if role, ok := reqMap["role"].(string); ok && role != "" {
		user.Role = role
	}

	return map[string]interface{}{
		"id":      user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"role":    user.Role,
		"updated": true,
	}, nil
}

func (s *UserService) deleteUser(ctx context.Context, req interface{}) (interface{}, error) {
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	id, _ := reqMap["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}

	_, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	delete(s.users, id)

	return map[string]interface{}{
		"id":      id,
		"deleted": true,
	}, nil
}

func (s *UserService) listUsers(ctx context.Context, req interface{}) (interface{}, error) {
	userList := make([]map[string]interface{}, 0, len(s.users))
	for _, user := range s.users {
		userList = append(userList, map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		})
	}

	return map[string]interface{}{
		"users": userList,
		"total": len(userList),
	}, nil
}

// AdminService implements admin operations
type AdminService struct{}

func (s *AdminService) ServiceName() string {
	return "AdminService"
}

func (s *AdminService) Methods() []string {
	return []string{"GetSystemStats", "ResetDatabase"}
}

func (s *AdminService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
	switch method {
	case "GetSystemStats":
		return s.getSystemStats(ctx, req)
	case "ResetDatabase":
		return s.resetDatabase(ctx, req)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

func (s *AdminService) HandleStream(stream pkg.GRPCServerStream, method string) error {
	return fmt.Errorf("streaming not implemented")
}

func (s *AdminService) GetMethodDescriptor(method string) *pkg.GRPCMethodDescriptor {
	return &pkg.GRPCMethodDescriptor{
		Name:           method,
		IsClientStream: false,
		IsServerStream: false,
	}
}

func (s *AdminService) getSystemStats(ctx context.Context, req interface{}) (interface{}, error) {
	return map[string]interface{}{
		"total_users": len(users),
		"uptime":      "1h",
		"version":     "1.0.0",
	}, nil
}

func (s *AdminService) resetDatabase(ctx context.Context, req interface{}) (interface{}, error) {
	// Clear all users
	count := len(users)
	users = make(map[string]*User)
	userIDSeq = 1

	return map[string]interface{}{
		"deleted": count,
		"success": true,
	}, nil
}

func main() {
	// Initialize sample data
	initSampleData()

	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true, // gRPC requires HTTP/2
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		SecurityConfig: pkg.SecurityConfig{
			JWTSecret: "my-secret-key",
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get router and database
	router := app.Router()
	db := app.Database()

	// Create auth manager
	authManager := pkg.NewAuthManager(db, "my-secret-key", pkg.OAuth2Config{})

	// Create gRPC manager
	grpcManager := pkg.NewGRPCManager(router, db, authManager)

	// Add logging middleware
	grpcManager.Use(func(ctx pkg.Context, next pkg.GRPCHandler) error {
		start := time.Now()
		log.Printf("gRPC request started: %s", ctx.Request().URL.Path)

		err := next(ctx)

		duration := time.Since(start)
		log.Printf("gRPC request completed: %s (duration: %v)", ctx.Request().URL.Path, duration)

		return err
	})

	// Create and register user service (public, with rate limiting)
	userService := NewUserService()
	err = grpcManager.RegisterService(userService, pkg.GRPCConfig{
		RequireAuth: false,
		RateLimit: &pkg.GRPCRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "ip_address",
		},
		MaxRequestSize: 1024 * 1024, // 1MB
		Timeout:        30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to register user service: %v", err)
	}

	// Create and register admin service (requires authentication)
	adminService := &AdminService{}
	err = grpcManager.RegisterService(adminService, pkg.GRPCConfig{
		RequireAuth:   true,
		RequiredRoles: []string{"admin"},
		RateLimit: &pkg.GRPCRateLimitConfig{
			Limit:  50,
			Window: time.Minute,
			Key:    "user_id",
		},
		MaxRequestSize: 1024 * 1024, // 1MB
		Timeout:        30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to register admin service: %v", err)
	}

	// Print startup information
	fmt.Println("🎸 gRPC Service Example")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("Server listening on http://localhost:8080")
	fmt.Println()
	fmt.Println("gRPC Services:")
	fmt.Println("  UserService (public)")
	fmt.Println("    - GetUser")
	fmt.Println("    - CreateUser")
	fmt.Println("    - UpdateUser")
	fmt.Println("    - DeleteUser")
	fmt.Println("    - ListUsers")
	fmt.Println()
	fmt.Println("  AdminService (requires authentication)")
	fmt.Println("    - GetSystemStats")
	fmt.Println("    - ResetDatabase")
	fmt.Println()
	fmt.Println("Service endpoints:")
	fmt.Println("  POST /UserService/GetUser")
	fmt.Println("  POST /UserService/CreateUser")
	fmt.Println("  POST /UserService/UpdateUser")
	fmt.Println("  POST /UserService/DeleteUser")
	fmt.Println("  POST /UserService/ListUsers")
	fmt.Println("  POST /AdminService/GetSystemStats")
	fmt.Println("  POST /AdminService/ResetDatabase")
	fmt.Println()
	fmt.Println("Example request (using curl):")
	fmt.Println("  curl -X POST http://localhost:8080/UserService/ListUsers \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{}'")
	fmt.Println()
	fmt.Println("  curl -X POST http://localhost:8080/UserService/GetUser \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"id\": \"1\"}'")
	fmt.Println()
	fmt.Println("  curl -X POST http://localhost:8080/UserService/CreateUser \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"name\": \"Alice\", \"email\": \"alice@example.com\", \"role\": \"user\"}'")
	fmt.Println()
	fmt.Println("Note: In production, use Protocol Buffers (.proto files) to define your services")
	fmt.Println("      and generate type-safe client/server code using protoc.")
	fmt.Println()
	fmt.Println("Rate limits:")
	fmt.Println("  UserService: 100 requests/minute per IP")
	fmt.Println("  AdminService: 50 requests/minute per user (requires authentication)")
	fmt.Println()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initSampleData initializes sample data
func initSampleData() {
	users["1"] = &User{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
		Role:  "admin",
	}
	userIDSeq = 2

	users["2"] = &User{
		ID:    "2",
		Name:  "Jane Smith",
		Email: "jane@example.com",
		Role:  "user",
	}
	userIDSeq = 3
}
