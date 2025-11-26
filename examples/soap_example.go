//go:build ignore

package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// 🎸 SOAP Service Example
// This example demonstrates SOAP service implementation with the Rockstar Web Framework
// Features: SOAP 1.1/1.2 support, WSDL generation, XML request/response handling, error handling

// User represents a user entity
type User struct {
	XMLName   xml.Name `xml:"User"`
	ID        string   `xml:"ID"`
	Name      string   `xml:"Name"`
	Email     string   `xml:"Email"`
	Role      string   `xml:"Role"`
	CreatedAt string   `xml:"CreatedAt"`
}

// GetUserRequest represents a GetUser SOAP request
type GetUserRequest struct {
	XMLName xml.Name `xml:"GetUser"`
	UserID  string   `xml:"UserID"`
}

// GetUserResponse represents a GetUser SOAP response
type GetUserResponse struct {
	XMLName xml.Name `xml:"GetUserResponse"`
	User    *User    `xml:"User"`
}

// CreateUserRequest represents a CreateUser SOAP request
type CreateUserRequest struct {
	XMLName xml.Name `xml:"CreateUser"`
	Name    string   `xml:"Name"`
	Email   string   `xml:"Email"`
	Role    string   `xml:"Role"`
}

// CreateUserResponse represents a CreateUser SOAP response
type CreateUserResponse struct {
	XMLName xml.Name `xml:"CreateUserResponse"`
	User    *User    `xml:"User"`
}

// ListUsersRequest represents a ListUsers SOAP request
type ListUsersRequest struct {
	XMLName xml.Name `xml:"ListUsers"`
}

// ListUsersResponse represents a ListUsers SOAP response
type ListUsersResponse struct {
	XMLName xml.Name `xml:"ListUsersResponse"`
	Users   []*User  `xml:"Users>User"`
	Total   int      `xml:"Total"`
}

// DeleteUserRequest represents a DeleteUser SOAP request
type DeleteUserRequest struct {
	XMLName xml.Name `xml:"DeleteUser"`
	UserID  string   `xml:"UserID"`
}

// DeleteUserResponse represents a DeleteUser SOAP response
type DeleteUserResponse struct {
	XMLName xml.Name `xml:"DeleteUserResponse"`
	Success bool     `xml:"Success"`
	UserID  string   `xml:"UserID"`
}

// In-memory storage for demonstration
var (
	users     = make(map[string]*User)
	userIDSeq = 1
)

// UserService implements a SOAP service for user management
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

// WSDL returns the WSDL document for the service
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
		{
			Name:        "CreateUser",
			InputType:   "CreateUserRequest",
			OutputType:  "CreateUserResponse",
			Description: "Creates a new user",
		},
		{
			Name:        "ListUsers",
			InputType:   "ListUsersRequest",
			OutputType:  "ListUsersResponse",
			Description: "Lists all users",
		},
		{
			Name:        "DeleteUser",
			InputType:   "DeleteUserRequest",
			OutputType:  "DeleteUserResponse",
			Description: "Deletes a user by ID",
		},
	}

	return pkg.GenerateWSDL(config, "http://localhost:8080/soap/user", operations)
}

// Execute handles SOAP operations
func (s *UserService) Execute(action string, body []byte) ([]byte, error) {
	switch action {
	case "GetUser":
		return s.handleGetUser(body)
	case "CreateUser":
		return s.handleCreateUser(body)
	case "ListUsers":
		return s.handleListUsers(body)
	case "DeleteUser":
		return s.handleDeleteUser(body)
	default:
		return nil, fmt.Errorf("unknown operation: %s", action)
	}
}

// handleGetUser handles the GetUser operation
func (s *UserService) handleGetUser(body []byte) ([]byte, error) {
	var req GetUserRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, exists := s.users[req.UserID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserID)
	}

	response := GetUserResponse{
		User: user,
	}

	return xml.MarshalIndent(response, "", "  ")
}

// handleCreateUser handles the CreateUser operation
func (s *UserService) handleCreateUser(body []byte) ([]byte, error) {
	var req CreateUserRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.Name == "" || req.Email == "" {
		return nil, fmt.Errorf("name and email are required")
	}

	if req.Role == "" {
		req.Role = "user"
	}

	// Generate user ID
	userID := fmt.Sprintf("%d", userIDSeq)
	userIDSeq++

	// Create user
	user := &User{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	s.users[userID] = user

	response := CreateUserResponse{
		User: user,
	}

	return xml.MarshalIndent(response, "", "  ")
}

// handleListUsers handles the ListUsers operation
func (s *UserService) handleListUsers(body []byte) ([]byte, error) {
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	response := ListUsersResponse{
		Users: users,
		Total: len(users),
	}

	return xml.MarshalIndent(response, "", "  ")
}

// handleDeleteUser handles the DeleteUser operation
func (s *UserService) handleDeleteUser(body []byte) ([]byte, error) {
	var req DeleteUserRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	_, exists := s.users[req.UserID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", req.UserID)
	}

	delete(s.users, req.UserID)

	response := DeleteUserResponse{
		Success: true,
		UserID:  req.UserID,
	}

	return xml.MarshalIndent(response, "", "  ")
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
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
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
	authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

	// Create SOAP manager
	soapManager := pkg.NewSOAPManager(router, db, authManager)

	// Create user service
	userService := NewUserService()

	// Configure SOAP service
	soapConfig := pkg.SOAPConfig{
		// Enable WSDL at /soap/user?wsdl
		EnableWSDL:  true,
		ServiceName: "UserService",
		Namespace:   "http://example.com/soap/user",
		PortName:    "UserServicePort",

		// Rate limiting
		RateLimit: &pkg.SOAPRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "ip_address",
		},

		// Request validation
		MaxRequestSize: 1024 * 1024, // 1MB
		Timeout:        30 * time.Second,

		// CORS support
		CORS: &pkg.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"POST", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "SOAPAction"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	// Register SOAP service
	err = soapManager.RegisterService("/soap/user", userService, soapConfig)
	if err != nil {
		log.Fatalf("Failed to register SOAP service: %v", err)
	}

	// Print startup information
	fmt.Println("🎸 SOAP Service Example")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("Server listening on http://localhost:8080")
	fmt.Println()
	fmt.Println("SOAP endpoints:")
	fmt.Println("  POST http://localhost:8080/soap/user       - SOAP service endpoint")
	fmt.Println("  GET  http://localhost:8080/soap/user?wsdl  - WSDL document")
	fmt.Println()
	fmt.Println("Available operations:")
	fmt.Println("  - GetUser")
	fmt.Println("  - CreateUser")
	fmt.Println("  - ListUsers")
	fmt.Println("  - DeleteUser")
	fmt.Println()
	fmt.Println("Example SOAP request (GetUser):")
	fmt.Println(`
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
`)
	fmt.Println("Example SOAP request (CreateUser):")
	fmt.Println(`
POST /soap/user HTTP/1.1
Host: localhost:8080
Content-Type: text/xml; charset=utf-8
SOAPAction: "CreateUser"

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <CreateUser>
      <Name>Alice Johnson</Name>
      <Email>alice@example.com</Email>
      <Role>user</Role>
    </CreateUser>
  </soap:Body>
</soap:Envelope>
`)
	fmt.Println("Example SOAP request (ListUsers):")
	fmt.Println(`
POST /soap/user HTTP/1.1
Host: localhost:8080
Content-Type: text/xml; charset=utf-8
SOAPAction: "ListUsers"

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <ListUsers/>
  </soap:Body>
</soap:Envelope>
`)
	fmt.Println("Try it with curl:")
	fmt.Println(`  curl -X POST http://localhost:8080/soap/user \
       -H 'Content-Type: text/xml; charset=utf-8' \
       -H 'SOAPAction: "ListUsers"' \
       -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <ListUsers/>
  </soap:Body>
</soap:Envelope>'`)
	fmt.Println()
	fmt.Println("View WSDL:")
	fmt.Println("  curl http://localhost:8080/soap/user?wsdl")
	fmt.Println()
	fmt.Println("Rate limits:")
	fmt.Println("  100 requests/minute per IP")
	fmt.Println()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initSampleData initializes sample data
func initSampleData() {
	users["1"] = &User{
		ID:        "1",
		Name:      "John Doe",
		Email:     "john@example.com",
		Role:      "admin",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	userIDSeq = 2

	users["2"] = &User{
		ID:        "2",
		Name:      "Jane Smith",
		Email:     "jane@example.com",
		Role:      "user",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	userIDSeq = 3
}
