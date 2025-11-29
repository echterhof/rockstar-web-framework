# SOAP API Example

The SOAP API example (`examples/soap_example.go`) demonstrates how to build SOAP web services with the Rockstar Web Framework. It showcases SOAP 1.1/1.2 support, WSDL generation, XML request/response handling, and error handling.

## What This Example Demonstrates

- **SOAP service** implementation
- **WSDL generation** for service discovery
- **XML request/response** handling
- **SOAP operations** (GetUser, CreateUser, ListUsers, DeleteUser)
- **Rate limiting** for API protection
- **CORS configuration** for cross-origin requests
- **Error handling** with SOAP faults

## Prerequisites

- Go 1.25 or higher

## Setup Instructions

```bash
go run examples/soap_example.go
```

The server will start on `http://localhost:8080`.

## SOAP Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/soap/user` | SOAP service endpoint |
| GET | `/soap/user?wsdl` | WSDL document |

## Available Operations

- **GetUser** - Retrieve user by ID
- **CreateUser** - Create new user
- **ListUsers** - List all users
- **DeleteUser** - Delete user by ID

## Testing the API

### View WSDL

```bash
curl http://localhost:8080/soap/user?wsdl
```

### GetUser Operation

```bash
curl -X POST http://localhost:8080/soap/user \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SOAPAction: "GetUser"' \
  -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUser>
      <UserID>1</UserID>
    </GetUser>
  </soap:Body>
</soap:Envelope>'
```

**Response**:
```xml
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetUserResponse>
      <User>
        <ID>1</ID>
        <Name>John Doe</Name>
        <Email>john@example.com</Email>
        <Role>admin</Role>
        <CreatedAt>2025-01-15T10:00:00Z</CreatedAt>
      </User>
    </GetUserResponse>
  </soap:Body>
</soap:Envelope>
```

### CreateUser Operation

```bash
curl -X POST http://localhost:8080/soap/user \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SOAPAction: "CreateUser"' \
  -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <CreateUser>
      <Name>Alice Johnson</Name>
      <Email>alice@example.com</Email>
      <Role>user</Role>
    </CreateUser>
  </soap:Body>
</soap:Envelope>'
```

### ListUsers Operation

```bash
curl -X POST http://localhost:8080/soap/user \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SOAPAction: "ListUsers"' \
  -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <ListUsers/>
  </soap:Body>
</soap:Envelope>'
```

### DeleteUser Operation

```bash
curl -X POST http://localhost:8080/soap/user \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SOAPAction: "DeleteUser"' \
  -d '<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <DeleteUser>
      <UserID>1</UserID>
    </DeleteUser>
  </soap:Body>
</soap:Envelope>'
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
        // ... more operations
    }
    
    return pkg.GenerateWSDL(config, "http://localhost:8080/soap/user", operations)
}

func (s *UserService) Execute(action string, body []byte) ([]byte, error) {
    switch action {
    case "GetUser":
        return s.handleGetUser(body)
    case "CreateUser":
        return s.handleCreateUser(body)
    // ... other operations
    }
}
```

### Request/Response Types

```go
type GetUserRequest struct {
    XMLName xml.Name `xml:"GetUser"`
    UserID  string   `xml:"UserID"`
}

type GetUserResponse struct {
    XMLName xml.Name `xml:"GetUserResponse"`
    User    *User    `xml:"User"`
}

type User struct {
    XMLName   xml.Name `xml:"User"`
    ID        string   `xml:"ID"`
    Name      string   `xml:"Name"`
    Email     string   `xml:"Email"`
    Role      string   `xml:"Role"`
    CreatedAt string   `xml:"CreatedAt"`
}
```

### Operation Handler

```go
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
```

### Service Registration

```go
soapManager := pkg.NewSOAPManager(router, db, authManager)

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
    
    MaxRequestSize: 1024 * 1024, // 1MB
    Timeout:        30 * time.Second,
    
    CORS: &pkg.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"POST", "OPTIONS"},
        AllowHeaders: []string{"Content-Type", "SOAPAction"},
    },
}

err = soapManager.RegisterService("/soap/user", userService, soapConfig)
```

## SOAP Fault Handling

When an error occurs, the service returns a SOAP fault:

```xml
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <soap:Fault>
      <faultcode>soap:Server</faultcode>
      <faultstring>user not found: 999</faultstring>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>
```

## WSDL Document

The WSDL document describes the service:

```xml
<?xml version="1.0"?>
<definitions xmlns="http://schemas.xmlsoap.org/wsdl/"
             xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/"
             xmlns:tns="http://example.com/soap/user"
             targetNamespace="http://example.com/soap/user">
  
  <types>
    <!-- XML Schema definitions -->
  </types>
  
  <message name="GetUserRequest">
    <part name="parameters" element="tns:GetUser"/>
  </message>
  
  <message name="GetUserResponse">
    <part name="parameters" element="tns:GetUserResponse"/>
  </message>
  
  <portType name="UserServicePortType">
    <operation name="GetUser">
      <input message="tns:GetUserRequest"/>
      <output message="tns:GetUserResponse"/>
    </operation>
  </portType>
  
  <binding name="UserServiceBinding" type="tns:UserServicePortType">
    <soap:binding transport="http://schemas.xmlsoap.org/soap/http"/>
    <operation name="GetUser">
      <soap:operation soapAction="GetUser"/>
      <input>
        <soap:body use="literal"/>
      </input>
      <output>
        <soap:body use="literal"/>
      </output>
    </operation>
  </binding>
  
  <service name="UserService">
    <port name="UserServicePort" binding="tns:UserServiceBinding">
      <soap:address location="http://localhost:8080/soap/user"/>
    </port>
  </service>
</definitions>
```

## Production Considerations

### Add Authentication

Protect SOAP operations:

```go
soapConfig := pkg.SOAPConfig{
    RequireAuth:   true,
    RequiredRoles: []string{"admin"},
    // ... other config
}
```

### Validate XML Schema

Add XML schema validation:

```go
func (s *UserService) ValidateRequest(body []byte) error {
    // Validate against XML schema
    return validateXMLSchema(body, userSchema)
}
```

### Add WS-Security

Implement WS-Security for message-level security:

```xml
<soap:Header>
  <wsse:Security xmlns:wsse="...">
    <wsse:UsernameToken>
      <wsse:Username>user</wsse:Username>
      <wsse:Password Type="...">password</wsse:Password>
    </wsse:UsernameToken>
  </wsse:Security>
</soap:Header>
```

### Use SOAP Client Libraries

For clients, use SOAP libraries:

```go
// Go SOAP client
import "github.com/tiaguinho/gosoap"

client, err := gosoap.SoapClient("http://localhost:8080/soap/user?wsdl")
params := gosoap.Params{
    "UserID": "1",
}
res, err := client.Call("GetUser", params)
```

## Common Issues

### "Invalid SOAP envelope"

**Solution**: Ensure proper XML structure with soap:Envelope and soap:Body

### "Missing SOAPAction header"

**Solution**: Include SOAPAction header in request

### "XML parsing error"

**Solution**: Validate XML syntax and encoding

## Related Documentation

- [API Styles Guide](../guides/api-styles.md) - REST, GraphQL, gRPC, SOAP
- [Security Guide](../guides/security.md) - Authentication and authorization

## Source Code

The complete source code is available at `examples/soap_example.go` in the repository.
