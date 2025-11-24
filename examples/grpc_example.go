package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// UserService implements a simple gRPC user service
type UserService struct {
	users map[string]*User
}

// User represents a user entity
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
		return s.getUser(req)
	case "CreateUser":
		return s.createUser(req)
	case "UpdateUser":
		return s.updateUser(req)
	case "DeleteUser":
		return s.deleteUser(req)
	case "ListUsers":
		return s.listUsers(req)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// HandleStream handles streaming RPC calls
func (s *UserService) HandleStream(stream pkg.GRPCServerStream, method string) error {
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

func (s *UserService) getUser(req interface{}) (interface{}, error) {
	// In a real implementation, parse the request and fetch user
	return map[string]interface{}{
		"id":    "123",
		"name":  "John Doe",
		"email": "john@example.com",
	}, nil
}

func (s *UserService) createUser(req interface{}) (interface{}, error) {
	// In a real implementation, parse the request and create user
	return map[string]interface{}{
		"id":      "456",
		"name":    "Jane Smith",
		"email":   "jane@example.com",
		"created": true,
	}, nil
}

func (s *UserService) updateUser(req interface{}) (interface{}, error) {
	return map[string]interface{}{
		"id":      "123",
		"updated": true,
	}, nil
}

func (s *UserService) deleteUser(req interface{}) (interface{}, error) {
	return map[string]interface{}{
		"id":      "123",
		"deleted": true,
	}, nil
}

func (s *UserService) listUsers(req interface{}) (interface{}, error) {
	users := []map[string]interface{}{
		{"id": "123", "name": "John Doe", "email": "john@example.com"},
		{"id": "456", "name": "Jane Smith", "email": "jane@example.com"},
	}
	return map[string]interface{}{
		"users": users,
		"total": len(users),
	}, nil
}

func main() {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1: true,
			EnableHTTP2: true,
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

	// Get router
	router := app.Router()

	// Get database manager
	db := app.Database()

	// Create auth manager
	authManager := pkg.NewAuthManager(db, "my-secret-key", pkg.OAuth2Config{})

	// Create gRPC manager
	grpcManager := pkg.NewGRPCManager(router, db, authManager)

	// Add logging middleware
	grpcManager.Use(func(ctx pkg.Context, next pkg.GRPCHandler) error {
		start := time.Now()
		log.Printf("gRPC request started: %s", ctx.Request().URL)

		err := next(ctx)

		duration := time.Since(start)
		log.Printf("gRPC request completed: %s (duration: %v)", ctx.Request().URL, duration)

		return err
	})

	// Create user service
	userService := &UserService{
		users: make(map[string]*User),
	}

	// Register service without authentication (public endpoints)
	err = grpcManager.RegisterService(userService, pkg.GRPCConfig{
		RequireAuth: false,
		RateLimit: &pkg.GRPCRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "ip_address",
		},
	})
	if err != nil {
		log.Fatalf("Failed to register user service: %v", err)
	}

	// Create admin service with authentication
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

	// Start server
	log.Println("Starting gRPC server on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// AdminService implements admin operations
type AdminService struct{}

func (s *AdminService) ServiceName() string {
	return "AdminService"
}

func (s *AdminService) Methods() []string {
	return []string{"DeleteAllUsers", "GetSystemStats"}
}

func (s *AdminService) HandleUnary(ctx context.Context, method string, req interface{}) (interface{}, error) {
	switch method {
	case "DeleteAllUsers":
		return map[string]interface{}{
			"deleted": true,
			"count":   0,
		}, nil
	case "GetSystemStats":
		return map[string]interface{}{
			"users":    0,
			"requests": 0,
			"uptime":   "1h",
		}, nil
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
