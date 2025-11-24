# SOAP Implementation

## Overview

The Rockstar Web Framework provides comprehensive SOAP (Simple Object Access Protocol) support, enabling developers to build and expose SOAP web services with built-in authentication, rate limiting, and WSDL generation.

**Requirements Implemented:**
- **2.4**: SOAP protocol support with WSDL generation
- **2.5**: Authentication and authorization integration
- **2.6**: Rate limiting per resource and globally

## Features

### Core SOAP Features

1. **SOAP 1.1 and 1.2 Support**
   - Automatic version detection from content type and envelope
   - Unified handling of both SOAP versions
   - Proper fault handling for each version

2. **WSDL Generation**
   - Automatic WSDL document generation
   - Support for multiple operations
   - Configurable service metadata (namespace, port name, etc.)
   - WSDL served at `?wsdl` query parameter

3. **Authentication Integration**
   - OAuth2 and JWT authentication support
   - Role-based access control (RBAC)
   - Scope-based authorization
   - Access token validation

4. **Rate Limiting**
   - Per-resource rate limiting
   - Global rate limiting
   - Database-backed rate limit storage
   - Configurable limits and time windows
   - Multiple key types (user_id, tenant_id, ip_address)

5. **Request Validation**
   - Request size limits
   - Timeout configuration
   - SOAP envelope validation
   - Automatic fault responses for invalid requests

6. **CORS Support**
   - Configurable CORS policies
   - Preflight request handling
   - Origin, method, and header validation

## Architecture

### SOAP Manager

The `SOAPManager` interface provides the main API for SOAP service management:

```go
type SOAPManager interface {
    // Service registration
    RegisterService(path string, service SOAPService, config SOAPConfig) error
    
    // WSDL support
    ServeWSDL(ctx Context, service SOAPService) error
    
    // Rate limiting
    CheckRateLimit(ctx Context, resource string) error
    CheckGlobalRateLimit(ctx Context) error
    
    // Middleware support
    Use(middleware SOAPMiddleware) SOAPManager
    
    // Service groups
    Group(prefix string, middleware ...SOAPMiddleware) SOAPManager
}
```

### SOAP Service Interface

Services must implement the `SOAPService` interface:

```go
type SOAPService interface {
    ServiceName() string
    WSDL() (string, error)
    Execute(action string, body []byte) ([]byte, error)
}
```

### SOAP Configuration

The `SOAPConfig` struct provides comprehensive configuration options:

```go
type SOAPConfig struct {
    // Rate limiting
    RateLimit       *SOAPRateLimitConfig
    GlobalRateLimit *SOAPRateLimitConfig
    
    // Authentication
    RequireAuth    bool
    RequiredScopes []string
    RequiredRoles  []string
    
    // Request validation
    MaxRequestSize int64
    Timeout        time.Duration
    
    // SOAP-specific settings
    EnableWSDL      bool
    WSDLPath        string
    Namespace       string
    ServiceName     string
    PortName        string
    
    // CORS
    CORS *CORSConfig
}
```

## Usage Examples

### Basic SOAP Service

```go
// Define your service
type CalculatorService struct{}

func (s *CalculatorService) ServiceName() string {
    return "CalculatorService"
}

func (s *CalculatorService) WSDL() (string, error) {
    config := pkg.SOAPConfig{
        ServiceName: "CalculatorService",
        Namespace:   "http://example.com/calculator",
        PortName:    "CalculatorPort",
    }
    
    operations := []pkg.WSDLOperation{
        {Name: "Add", InputType: "AddRequest", OutputType: "AddResponse"},
        {Name: "Subtract", InputType: "SubtractRequest", OutputType: "SubtractResponse"},
    }
    
    return pkg.GenerateWSDL(config, "http://localhost:8080/soap/calc", operations)
}

func (s *CalculatorService) Execute(action string, body []byte) ([]byte, error) {
    switch action {
    case "Add":
        return s.handleAdd(body)
    case "Subtract":
        return s.handleSubtract(body)
    default:
        return nil, fmt.Errorf("unknown operation: %s", action)
    }
}

// Register the service
router := pkg.NewRouterEngine()
db := pkg.NewMockDatabase()
authManager := pkg.NewAuthManager(db)

soapManager := pkg.NewSOAPManager(router, db, authManager)

config := pkg.SOAPConfig{
    EnableWSDL:  true,
    ServiceName: "CalculatorService",
    Namespace:   "http://example.com/calculator",
    PortName:    "CalculatorPort",
}

soapManager.RegisterService("/soap/calc", &CalculatorService{}, config)
```

### SOAP Service with Authentication

```go
config := pkg.SOAPConfig{
    EnableWSDL:     true,
    RequireAuth:    true,
    RequiredScopes: []string{"soap:read", "soap:write"},
    RequiredRoles:  []string{"admin", "user"},
    ServiceName:    "SecureService",
    Namespace:      "http://example.com/secure",
    PortName:       "SecurePort",
}

soapManager.RegisterService("/soap/secure", service, config)
```

### SOAP Service with Rate Limiting

```go
config := pkg.SOAPConfig{
    EnableWSDL: true,
    
    // Per-resource rate limiting
    RateLimit: &pkg.SOAPRateLimitConfig{
        Limit:  100,              // 100 requests
        Window: time.Minute,      // per minute
        Key:    "user_id",        // per user
    },
    
    // Global rate limiting
    GlobalRateLimit: &pkg.SOAPRateLimitConfig{
        Limit:  1000,             // 1000 requests
        Window: time.Minute,      // per minute
        Key:    "ip_address",     // per IP
    },
    
    ServiceName: "RateLimitedService",
    Namespace:   "http://example.com/ratelimited",
    PortName:    "RateLimitedPort",
}

soapManager.RegisterService("/soap/ratelimited", service, config)
```

### SOAP Service with Request Validation

```go
config := pkg.SOAPConfig{
    EnableWSDL:     true,
    MaxRequestSize: 1024 * 1024,  // 1MB limit
    Timeout:        30 * time.Second,
    ServiceName:    "ValidatedService",
    Namespace:      "http://example.com/validated",
    PortName:       "ValidatedPort",
}

soapManager.RegisterService("/soap/validated", service, config)
```

### SOAP Service with CORS

```go
config := pkg.SOAPConfig{
    EnableWSDL: true,
    CORS: &pkg.CORSConfig{
        AllowOrigins:     []string{"http://example.com", "https://example.com"},
        AllowMethods:     []string{"POST", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "SOAPAction"},
        AllowCredentials: true,
        MaxAge:           3600,
    },
    ServiceName: "CORSService",
    Namespace:   "http://example.com/cors",
    PortName:    "CORSPort",
}

soapManager.RegisterService("/soap/cors", service, config)
```

### Using Middleware

```go
// Add custom middleware
soapManager.Use(func(ctx pkg.Context, next pkg.SOAPHandler) error {
    // Pre-processing
    fmt.Println("Before SOAP request")
    
    // Call next handler
    err := next(ctx)
    
    // Post-processing
    fmt.Println("After SOAP request")
    
    return err
})

// Register service with middleware
soapManager.RegisterService("/soap/service", service, config)
```

### Using Service Groups

```go
// Create a group with prefix and middleware
apiV1 := soapManager.Group("/api/v1", loggingMiddleware, authMiddleware)

// Register services in the group
apiV1.RegisterService("/soap/users", userService, userConfig)
apiV1.RegisterService("/soap/orders", orderService, orderConfig)

// Services will be available at:
// - /api/v1/soap/users
// - /api/v1/soap/orders
```

## SOAP Request/Response Format

### SOAP 1.1 Request

```xml
POST /soap/service HTTP/1.1
Host: localhost:8080
Content-Type: text/xml; charset=utf-8
SOAPAction: "OperationName"

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <!-- Optional header elements -->
  </soap:Header>
  <soap:Body>
    <OperationName>
      <Parameter1>Value1</Parameter1>
      <Parameter2>Value2</Parameter2>
    </OperationName>
  </soap:Body>
</soap:Envelope>
```

### SOAP 1.1 Response

```xml
HTTP/1.1 200 OK
Content-Type: text/xml; charset=utf-8

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <OperationNameResponse>
      <Result>Success</Result>
    </OperationNameResponse>
  </soap:Body>
</soap:Envelope>
```

### SOAP 1.1 Fault

```xml
HTTP/1.1 500 Internal Server Error
Content-Type: text/xml; charset=utf-8

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Client</faultcode>
      <faultstring>Invalid request</faultstring>
      <faultactor>http://example.com/actor</faultactor>
      <detail>Missing required parameter</detail>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>
```

### SOAP 1.2 Request

```xml
POST /soap/service HTTP/1.1
Host: localhost:8080
Content-Type: application/soap+xml; charset=utf-8

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Header>
    <!-- Optional header elements -->
  </soap:Header>
  <soap:Body>
    <OperationName>
      <Parameter1>Value1</Parameter1>
      <Parameter2>Value2</Parameter2>
    </OperationName>
  </soap:Body>
</soap:Envelope>
```

### SOAP 1.2 Fault

```xml
HTTP/1.1 500 Internal Server Error
Content-Type: application/soap+xml; charset=utf-8

<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <soap:Fault>
      <soap:Code>
        <soap:Value>soap:Sender</soap:Value>
      </soap:Code>
      <soap:Reason>
        <soap:Text>Invalid request</soap:Text>
      </soap:Reason>
      <soap:Detail>Missing required parameter</soap:Detail>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>
```

## WSDL Access

WSDL documents are automatically generated and served when `EnableWSDL` is set to `true`:

```
GET /soap/service?wsdl HTTP/1.1
Host: localhost:8080
```

Response:
```xml
<?xml version="1.0"?>
<definitions name="ServiceName" 
             targetNamespace="http://example.com/soap"
             xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:xsd="http://www.w3.org/2001/XMLSchema"
             xmlns:tns="http://example.com/soap">
  
  <portType name="ServicePort">
    <operation name="OperationName">
      <input message="tns:OperationNameRequest"/>
      <output message="tns:OperationNameResponse"/>
    </operation>
  </portType>
  
  <binding name="ServiceBinding" type="tns:ServicePort">
    <soap:binding style="document" transport="http://schemas.xmlsoap.org/soap/http"/>
    <operation name="OperationName">
      <soap:operation soapAction="http://example.com/soap/OperationName"/>
      <input><soap:body use="literal"/></input>
      <output><soap:body use="literal"/></output>
    </operation>
  </binding>
  
  <service name="ServiceName">
    <port name="ServicePort" binding="tns:ServiceBinding">
      <soap:address location="http://localhost:8080/soap/service"/>
    </port>
  </service>
  
</definitions>
```

## Error Handling

The framework provides standard SOAP faults for common errors:

- `ErrSOAPInvalidRequest`: Invalid SOAP request (Client fault)
- `ErrSOAPUnauthenticated`: Authentication required (Client fault)
- `ErrSOAPPermissionDenied`: Permission denied (Client fault)
- `ErrSOAPNotFound`: Service or method not found (Client fault)
- `ErrSOAPRateLimit`: Rate limit exceeded (Client fault)
- `ErrSOAPInternalError`: Internal server error (Server fault)
- `ErrSOAPServiceUnavailable`: Service unavailable (Server fault)
- `ErrSOAPTimeout`: Request timeout (Server fault)

## Best Practices

1. **Use WSDL**: Always enable WSDL generation for better client integration
2. **Implement Rate Limiting**: Protect your services with appropriate rate limits
3. **Validate Input**: Use request size limits and timeouts to prevent abuse
4. **Secure Services**: Enable authentication for sensitive operations
5. **Handle Errors Gracefully**: Return proper SOAP faults with meaningful messages
6. **Version Your Services**: Use service groups to manage API versions
7. **Document Operations**: Provide clear descriptions in WSDL operations
8. **Test Both Versions**: Ensure your service works with SOAP 1.1 and 1.2

## Testing

The framework includes comprehensive unit tests for SOAP functionality:

```bash
go test -v ./pkg -run TestSOAP
```

Test coverage includes:
- SOAP envelope parsing (1.1 and 1.2)
- Fault creation and marshaling
- Version detection
- Service registration
- Rate limiting
- Authentication
- WSDL generation
- Middleware functionality
- Request validation
- CORS support

## Performance Considerations

1. **Connection Pooling**: The framework uses database connection pooling for rate limiting
2. **Memory Management**: SOAP envelopes are parsed efficiently with minimal allocations
3. **Caching**: WSDL documents can be cached to reduce generation overhead
4. **Concurrent Requests**: The framework handles concurrent SOAP requests safely

## See Also

- [REST API Implementation](rest_api_implementation.md)
- [GraphQL Implementation](graphql_implementation.md)
- [gRPC Implementation](grpc_implementation.md)
- [Authentication System](../pkg/auth.go)
- [Rate Limiting](../pkg/database.go)
