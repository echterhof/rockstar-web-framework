package main

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// UserService implements a SOAP service for user management
type UserService struct {
	users map[string]*User
}

// User represents a user entity
type User struct {
	XMLName xml.Name `xml:"User"`
	ID      string   `xml:"ID"`
	Name    string   `xml:"Name"`
	Email   string   `xml:"Email"`
	Created string   `xml:"Created"`
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
}

// CreateUserResponse represents a CreateUser SOAP response
type CreateUserResponse struct {
	XMLName xml.Name `xml:"CreateUserResponse"`
	User    *User    `xml:"User"`
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
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

	// Generate user ID
	userID := fmt.Sprintf("user_%d", len(s.users)+1)

	// Create user
	user := &User{
		ID:      userID,
		Name:    req.Name,
		Email:   req.Email,
		Created: time.Now().Format(time.RFC3339),
	}

	s.users[userID] = user

	response := CreateUserResponse{
		User: user,
	}

	return xml.MarshalIndent(response, "", "  ")
}

// handleListUsers handles the ListUsers operation
func (s *UserService) handleListUsers(body []byte) ([]byte, error) {
	type ListUsersResponse struct {
		XMLName xml.Name `xml:"ListUsersResponse"`
		Users   []*User  `xml:"Users>User"`
	}

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	response := ListUsersResponse{
		Users: users,
	}

	return xml.MarshalIndent(response, "", "  ")
}

func main() {
	// Note: This example shows the SOAP service structure
	// In a real application, you would initialize the framework components properly

	// For demonstration purposes, we'll show the service structure
	// Actual initialization would be:
	// router := pkg.NewRouter()
	// db := pkg.NewDatabase(config)
	// authManager := pkg.NewAuthManager(db, "secret", oauth2Config)
	// soapManager := pkg.NewSOAPManager(router, db, authManager)

	fmt.Println("SOAP Service Example")
	fmt.Println("====================")
	fmt.Println()

	// Create user service
	userService := NewUserService()

	// Add some sample users
	userService.users["user_1"] = &User{
		ID:      "user_1",
		Name:    "John Doe",
		Email:   "john@example.com",
		Created: time.Now().Format(time.RFC3339),
	}
	userService.users["user_2"] = &User{
		ID:      "user_2",
		Name:    "Jane Smith",
		Email:   "jane@example.com",
		Created: time.Now().Format(time.RFC3339),
	}

	// Show service configuration
	fmt.Println("Service Configuration:")
	fmt.Println("----------------------")
	config := pkg.SOAPConfig{
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

		// Global rate limiting
		GlobalRateLimit: &pkg.SOAPRateLimitConfig{
			Limit:  1000,
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

	// Show registration (in real app, you would call soapManager.RegisterService)
	fmt.Printf("  Service Name: %s\n", config.ServiceName)
	fmt.Printf("  Namespace: %s\n", config.Namespace)
	fmt.Printf("  WSDL Enabled: %v\n", config.EnableWSDL)
	fmt.Printf("  Rate Limit: %d requests per %v\n", config.RateLimit.Limit, config.RateLimit.Window)
	fmt.Println()

	fmt.Println("SOAP service would be registered at:")
	fmt.Println("  Service endpoint: http://localhost:8080/soap/user")
	fmt.Println("  WSDL endpoint: http://localhost:8080/soap/user?wsdl")
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
      <UserID>user_1</UserID>
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
      <Name>Bob Johnson</Name>
      <Email>bob@example.com</Email>
    </CreateUser>
  </soap:Body>
</soap:Envelope>
`)

	// In a real application, you would start the server here
	// server := pkg.NewServer(router)
	// server.Listen(":8080")
}
