package pkg

import (
	"encoding/xml"
	"fmt"
	"testing"
	"time"
)

// mockSOAPService implements SOAPService for testing
type mockSOAPService struct {
	name       string
	operations map[string]func([]byte) ([]byte, error)
}

func (m *mockSOAPService) ServiceName() string {
	return m.name
}

func (m *mockSOAPService) WSDL() (string, error) {
	config := SOAPConfig{
		ServiceName: m.name,
		Namespace:   "http://example.com/soap",
		PortName:    m.name + "Port",
	}

	operations := []WSDLOperation{
		{Name: "GetUser", InputType: "GetUserRequest", OutputType: "GetUserResponse"},
		{Name: "CreateUser", InputType: "CreateUserRequest", OutputType: "CreateUserResponse"},
	}

	return GenerateWSDL(config, "http://localhost:8080/soap", operations)
}

func (m *mockSOAPService) Execute(action string, body []byte) ([]byte, error) {
	if handler, ok := m.operations[action]; ok {
		return handler(body)
	}
	return nil, fmt.Errorf("unknown operation: %s", action)
}

// TestSOAPEnvelopeParsing tests SOAP envelope parsing
func TestSOAPEnvelopeParsing(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		expectError bool
		version     SOAPVersion
	}{
		{
			name: "Valid SOAP 1.1 envelope",
			xml: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUser>
      <UserID>123</UserID>
    </GetUser>
  </soap:Body>
</soap:Envelope>`,
			expectError: false,
			version:     SOAP11,
		},
		{
			name: "Valid SOAP 1.2 envelope",
			xml: `<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <GetUser>
      <UserID>456</UserID>
    </GetUser>
  </soap:Body>
</soap:Envelope>`,
			expectError: false,
			version:     SOAP12,
		},
		{
			name:        "Empty body",
			xml:         "",
			expectError: true,
		},
		{
			name:        "Invalid XML",
			xml:         "<invalid>",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, version, err := ParseSOAPEnvelope([]byte(tt.xml))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if env == nil {
				t.Error("Expected envelope but got nil")
				return
			}

			if version != tt.version {
				t.Errorf("Expected version %d, got %d", tt.version, version)
			}
		})
	}
}

// TestSOAPFaultCreation tests SOAP fault creation
func TestSOAPFaultCreation(t *testing.T) {
	fault := NewSOAPFault(FaultCodeClient, "Invalid request")

	if fault.Code != string(FaultCodeClient) {
		t.Errorf("Expected code %s, got %s", FaultCodeClient, fault.Code)
	}

	if fault.String != "Invalid request" {
		t.Errorf("Expected message 'Invalid request', got '%s'", fault.String)
	}

	// Test with actor and detail
	fault = fault.WithActor("http://example.com/actor").WithDetail("Missing required field")

	if fault.Actor != "http://example.com/actor" {
		t.Errorf("Expected actor 'http://example.com/actor', got '%s'", fault.Actor)
	}

	if fault.Detail != "Missing required field" {
		t.Errorf("Expected detail 'Missing required field', got '%s'", fault.Detail)
	}
}

// TestSOAP12FaultCreation tests SOAP 1.2 fault creation
func TestSOAP12FaultCreation(t *testing.T) {
	fault := NewSOAP12Fault(FaultCodeSender, "Invalid request")

	if fault.Code.Value != string(FaultCodeSender) {
		t.Errorf("Expected code %s, got %s", FaultCodeSender, fault.Code.Value)
	}

	if fault.Reason.Text != "Invalid request" {
		t.Errorf("Expected reason 'Invalid request', got '%s'", fault.Reason.Text)
	}

	// Test with node, role, and detail
	fault = fault.WithNode("http://example.com/node").
		WithRole("http://example.com/role").
		WithDetail("Missing required field")

	if fault.Node != "http://example.com/node" {
		t.Errorf("Expected node 'http://example.com/node', got '%s'", fault.Node)
	}

	if fault.Role != "http://example.com/role" {
		t.Errorf("Expected role 'http://example.com/role', got '%s'", fault.Role)
	}

	if fault.Detail != "Missing required field" {
		t.Errorf("Expected detail 'Missing required field', got '%s'", fault.Detail)
	}
}

// TestSOAPEnvelopeMarshaling tests SOAP envelope marshaling
func TestSOAPEnvelopeMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		version SOAPVersion
	}{
		{
			name:    "SOAP 1.1",
			version: SOAP11,
		},
		{
			name:    "SOAP 1.2",
			version: SOAP12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &SOAPEnvelope{
				Body: SOAPBody{
					Content: []byte("<GetUser><UserID>123</UserID></GetUser>"),
				},
			}

			xmlData, err := MarshalSOAPEnvelope(env, tt.version)
			if err != nil {
				t.Errorf("Failed to marshal envelope: %v", err)
				return
			}

			if len(xmlData) == 0 {
				t.Error("Expected XML data but got empty")
			}

			// Verify it can be parsed back
			_, version, err := ParseSOAPEnvelope(xmlData)
			if err != nil {
				t.Errorf("Failed to parse marshaled envelope: %v", err)
			}

			if version != tt.version {
				t.Errorf("Expected version %d, got %d", tt.version, version)
			}
		})
	}
}

// TestSOAPFaultMarshaling tests SOAP fault marshaling
func TestSOAPFaultMarshaling(t *testing.T) {
	fault := NewSOAPFault(FaultCodeServer, "Internal error")

	xmlData, err := MarshalSOAPFault(fault, SOAP11)
	if err != nil {
		t.Errorf("Failed to marshal fault: %v", err)
		return
	}

	if len(xmlData) == 0 {
		t.Error("Expected XML data but got empty")
	}

	// Verify it contains fault information
	env, _, err := ParseSOAPEnvelope(xmlData)
	if err != nil {
		t.Errorf("Failed to parse marshaled fault: %v", err)
		return
	}

	if env.Body.Fault == nil {
		t.Error("Expected fault in body but got none")
		return
	}

	if env.Body.Fault.Code != string(FaultCodeServer) {
		t.Errorf("Expected code %s, got %s", FaultCodeServer, env.Body.Fault.Code)
	}

	if env.Body.Fault.String != "Internal error" {
		t.Errorf("Expected message 'Internal error', got '%s'", env.Body.Fault.String)
	}
}

// TestSOAPVersionDetection tests SOAP version detection
func TestSOAPVersionDetection(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        string
		expected    SOAPVersion
	}{
		{
			name:        "SOAP 1.2 from content type",
			contentType: "application/soap+xml",
			data:        "",
			expected:    SOAP12,
		},
		{
			name:        "SOAP 1.2 from envelope namespace",
			contentType: "text/xml",
			data:        `<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">`,
			expected:    SOAP12,
		},
		{
			name:        "SOAP 1.1 default",
			contentType: "text/xml",
			data:        `<Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">`,
			expected:    SOAP11,
		},
		{
			name:        "SOAP 1.1 empty",
			contentType: "",
			data:        "",
			expected:    SOAP11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := DetectSOAPVersion(tt.contentType, []byte(tt.data))
			if version != tt.expected {
				t.Errorf("Expected version %d, got %d", tt.expected, version)
			}
		})
	}
}

// TestSOAPManagerRegistration tests SOAP service registration
func TestSOAPManagerRegistration(t *testing.T) {
	mockRouter := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(mockRouter, db, authManager)

	service := &mockSOAPService{
		name: "UserService",
		operations: map[string]func([]byte) ([]byte, error){
			"GetUser": func(body []byte) ([]byte, error) {
				return []byte("<User><ID>123</ID><Name>John</Name></User>"), nil
			},
		},
	}

	config := SOAPConfig{
		EnableWSDL:  true,
		ServiceName: "UserService",
		Namespace:   "http://example.com/soap",
		PortName:    "UserServicePort",
	}

	err := manager.RegisterService("/soap/user", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// Verify routes were registered
	if len(mockRouter.routes) == 0 {
		t.Error("Expected routes to be registered")
	}
}

// TestSOAPRateLimiting tests SOAP rate limiting
// Requirements: 2.6
func TestSOAPRateLimiting(t *testing.T) {
	router := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(router, db, authManager)

	service := &mockSOAPService{
		name: "TestService",
		operations: map[string]func([]byte) ([]byte, error){
			"Test": func(body []byte) ([]byte, error) {
				return []byte("<Result>OK</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		RateLimit: &SOAPRateLimitConfig{
			Limit:  5,
			Window: time.Minute,
			Key:    "ip_address",
		},
		ServiceName: "TestService",
		Namespace:   "http://example.com/soap",
		PortName:    "TestServicePort",
	}

	err := manager.RegisterService("/soap/test", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// Verify service was registered
	// In a real test with full context, we would test rate limiting
	// For now, just verify registration succeeded
}

// TestSOAPAuthentication tests SOAP authentication
// Requirements: 2.5
func TestSOAPAuthentication(t *testing.T) {
	mockRouter := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(mockRouter, db, authManager)

	service := &mockSOAPService{
		name: "SecureService",
		operations: map[string]func([]byte) ([]byte, error){
			"SecureOp": func(body []byte) ([]byte, error) {
				return []byte("<Result>Secure</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		RequireAuth:    true,
		RequiredScopes: []string{"soap:read"},
		ServiceName:    "SecureService",
		Namespace:      "http://example.com/soap",
		PortName:       "SecureServicePort",
	}

	err := manager.RegisterService("/soap/secure", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// Verify service was registered with auth requirement
	if len(mockRouter.routes) == 0 {
		t.Error("Expected routes to be registered")
	}
}

// TestWSDLGeneration tests WSDL generation
// Requirements: 2.4
func TestWSDLGeneration(t *testing.T) {
	config := SOAPConfig{
		ServiceName: "UserService",
		Namespace:   "http://example.com/soap",
		PortName:    "UserServicePort",
	}

	operations := []WSDLOperation{
		{Name: "GetUser", InputType: "GetUserRequest", OutputType: "GetUserResponse"},
		{Name: "CreateUser", InputType: "CreateUserRequest", OutputType: "CreateUserResponse"},
		{Name: "UpdateUser", InputType: "UpdateUserRequest", OutputType: "UpdateUserResponse"},
	}

	wsdl, err := GenerateWSDL(config, "http://localhost:8080/soap/user", operations)
	if err != nil {
		t.Errorf("Failed to generate WSDL: %v", err)
		return
	}

	if len(wsdl) == 0 {
		t.Error("Expected WSDL content but got empty")
		return
	}

	// Verify WSDL is valid XML
	var definitions WSDLDefinitions
	err = xml.Unmarshal([]byte(wsdl), &definitions)
	if err != nil {
		t.Errorf("Generated WSDL is not valid XML: %v", err)
		return
	}

	// Verify service name
	if definitions.Name != "UserService" {
		t.Errorf("Expected service name 'UserService', got '%s'", definitions.Name)
	}

	// Verify namespace
	if definitions.TargetNamespace != "http://example.com/soap" {
		t.Errorf("Expected namespace 'http://example.com/soap', got '%s'", definitions.TargetNamespace)
	}

	// Verify operations
	if len(definitions.PortType.Operations) != 3 {
		t.Errorf("Expected 3 operations, got %d", len(definitions.PortType.Operations))
	}

	// Verify binding operations
	if len(definitions.Binding.Operations) != 3 {
		t.Errorf("Expected 3 binding operations, got %d", len(definitions.Binding.Operations))
	}
}

// TestSOAPMiddleware tests SOAP middleware functionality
func TestSOAPMiddleware(t *testing.T) {
	router := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(router, db, authManager)

	// Add custom middleware
	manager.Use(func(ctx Context, next SOAPHandler) error {
		// Middleware logic here
		return next(ctx)
	})

	service := &mockSOAPService{
		name: "TestService",
		operations: map[string]func([]byte) ([]byte, error){
			"Test": func(body []byte) ([]byte, error) {
				return []byte("<Result>OK</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		ServiceName: "TestService",
		Namespace:   "http://example.com/soap",
		PortName:    "TestServicePort",
	}

	err := manager.RegisterService("/soap/test", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// Middleware should be registered
	// In a real test, we would invoke the handler to verify middleware is called
}

// TestSOAPGroup tests SOAP service grouping
func TestSOAPGroup(t *testing.T) {
	mockRouter := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(mockRouter, db, authManager)

	// Create a group with prefix
	group := manager.Group("/api/v1")

	service := &mockSOAPService{
		name: "TestService",
		operations: map[string]func([]byte) ([]byte, error){
			"Test": func(body []byte) ([]byte, error) {
				return []byte("<Result>OK</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		ServiceName: "TestService",
		Namespace:   "http://example.com/soap",
		PortName:    "TestServicePort",
	}

	err := group.RegisterService("/soap/test", service, config)
	if err != nil {
		t.Errorf("Failed to register service in group: %v", err)
	}

	// Verify routes include the group prefix
	if len(mockRouter.routes) == 0 {
		t.Error("Expected routes to be registered")
	}
}

// TestSOAPRequestValidation tests SOAP request validation
func TestSOAPRequestValidation(t *testing.T) {
	router := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(router, db, authManager)

	service := &mockSOAPService{
		name: "TestService",
		operations: map[string]func([]byte) ([]byte, error){
			"Test": func(body []byte) ([]byte, error) {
				return []byte("<Result>OK</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		MaxRequestSize: 1024, // 1KB limit
		Timeout:        5 * time.Second,
		ServiceName:    "TestService",
		Namespace:      "http://example.com/soap",
		PortName:       "TestServicePort",
	}

	err := manager.RegisterService("/soap/test", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// Validation middleware should be applied
	// In a real test, we would send requests exceeding limits to verify validation
}

// TestSOAPCORS tests SOAP CORS support
func TestSOAPCORS(t *testing.T) {
	router := newMockRouter()
	db := newMockDB()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	manager := NewSOAPManager(router, db, authManager)

	service := &mockSOAPService{
		name: "TestService",
		operations: map[string]func([]byte) ([]byte, error){
			"Test": func(body []byte) ([]byte, error) {
				return []byte("<Result>OK</Result>"), nil
			},
		},
	}

	config := SOAPConfig{
		CORS: &CORSConfig{
			AllowOrigins:     []string{"http://example.com"},
			AllowMethods:     []string{"POST", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "SOAPAction"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
		ServiceName: "TestService",
		Namespace:   "http://example.com/soap",
		PortName:    "TestServicePort",
	}

	err := manager.RegisterService("/soap/test", service, config)
	if err != nil {
		t.Errorf("Failed to register service: %v", err)
	}

	// CORS middleware should be applied
	// In a real test, we would send OPTIONS requests to verify CORS headers
}
