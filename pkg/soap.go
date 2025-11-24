package pkg

import (
	"encoding/xml"
	"fmt"
	"time"
)

// SOAPManager defines the SOAP interface
// Requirements: 2.4
type SOAPManager interface {
	// Service registration
	RegisterService(path string, service SOAPService, config SOAPConfig) error

	// WSDL support
	// Requirements: 2.4
	ServeWSDL(ctx Context, service SOAPService) error

	// Rate limiting
	// Requirements: 2.6
	CheckRateLimit(ctx Context, resource string) error
	CheckGlobalRateLimit(ctx Context) error

	// Middleware support
	Use(middleware SOAPMiddleware) SOAPManager

	// Service groups
	Group(prefix string, middleware ...SOAPMiddleware) SOAPManager
}

// SOAPMiddleware represents SOAP middleware
type SOAPMiddleware func(ctx Context, next SOAPHandler) error

// SOAPHandler represents a SOAP handler function
type SOAPHandler func(ctx Context) error

// SOAPConfig defines configuration for a SOAP service
type SOAPConfig struct {
	// Rate limiting configuration
	// Requirements: 2.6
	RateLimit       *SOAPRateLimitConfig
	GlobalRateLimit *SOAPRateLimitConfig

	// Authentication and authorization
	// Requirements: 2.5
	RequireAuth    bool
	RequiredScopes []string
	RequiredRoles  []string

	// Request validation
	MaxRequestSize int64
	Timeout        time.Duration

	// SOAP-specific settings
	// Requirements: 2.4
	EnableWSDL  bool
	WSDLPath    string // Path to serve WSDL (default: ?wsdl)
	Namespace   string
	ServiceName string
	PortName    string

	// Response configuration
	CORS *CORSConfig
}

// SOAPRateLimitConfig defines rate limiting configuration for SOAP
// Requirements: 2.6
type SOAPRateLimitConfig struct {
	Limit  int           // Maximum number of requests
	Window time.Duration // Time window for the limit
	Key    string        // Rate limit key (e.g., "user_id", "ip_address", "tenant_id")
}

// SOAPEnvelope represents a SOAP 1.1/1.2 envelope
type SOAPEnvelope struct {
	XMLName xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  *SOAPHeader `xml:"Header,omitempty"`
	Body    SOAPBody    `xml:"Body"`
}

// SOAPHeader represents SOAP header
type SOAPHeader struct {
	Content []byte `xml:",innerxml"`
}

// SOAPBody represents SOAP body
type SOAPBody struct {
	Content []byte     `xml:",innerxml"`
	Fault   *SOAPFault `xml:"Fault,omitempty"`
}

// SOAPFault represents a SOAP fault
type SOAPFault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`
	Code    string   `xml:"faultcode"`
	String  string   `xml:"faultstring"`
	Actor   string   `xml:"faultactor,omitempty"`
	Detail  string   `xml:"detail,omitempty"`
}

// SOAP12Envelope represents a SOAP 1.2 envelope
type SOAP12Envelope struct {
	XMLName xml.Name      `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *SOAP12Header `xml:"Header,omitempty"`
	Body    SOAP12Body    `xml:"Body"`
}

// SOAP12Header represents SOAP 1.2 header
type SOAP12Header struct {
	Content []byte `xml:",innerxml"`
}

// SOAP12Body represents SOAP 1.2 body
type SOAP12Body struct {
	Content []byte       `xml:",innerxml"`
	Fault   *SOAP12Fault `xml:"Fault,omitempty"`
}

// SOAP12Fault represents a SOAP 1.2 fault
type SOAP12Fault struct {
	XMLName xml.Name          `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
	Code    SOAP12FaultCode   `xml:"Code"`
	Reason  SOAP12FaultReason `xml:"Reason"`
	Node    string            `xml:"Node,omitempty"`
	Role    string            `xml:"Role,omitempty"`
	Detail  string            `xml:"Detail,omitempty"`
}

// SOAP12FaultCode represents SOAP 1.2 fault code
type SOAP12FaultCode struct {
	Value   string              `xml:"Value"`
	Subcode *SOAP12FaultSubcode `xml:"Subcode,omitempty"`
}

// SOAP12FaultSubcode represents SOAP 1.2 fault subcode
type SOAP12FaultSubcode struct {
	Value string `xml:"Value"`
}

// SOAP12FaultReason represents SOAP 1.2 fault reason
type SOAP12FaultReason struct {
	Text string `xml:"Text"`
}

// SOAPRequest represents a parsed SOAP request
type SOAPRequest struct {
	Version    SOAPVersion
	Action     string
	Body       []byte
	Header     []byte
	RemoteAddr string
}

// SOAPResponse represents a SOAP response
type SOAPResponse struct {
	Version SOAPVersion
	Body    []byte
	Header  []byte
	Fault   *SOAPFault
}

// SOAPVersion represents SOAP protocol version
type SOAPVersion int

const (
	SOAP11 SOAPVersion = 11
	SOAP12 SOAPVersion = 12
)

// SOAPFaultCode represents standard SOAP fault codes
type SOAPFaultCode string

const (
	// SOAP 1.1 fault codes
	FaultCodeVersionMismatch SOAPFaultCode = "VersionMismatch"
	FaultCodeMustUnderstand  SOAPFaultCode = "MustUnderstand"
	FaultCodeClient          SOAPFaultCode = "Client"
	FaultCodeServer          SOAPFaultCode = "Server"

	// SOAP 1.2 fault codes
	FaultCodeSender              SOAPFaultCode = "Sender"
	FaultCodeReceiver            SOAPFaultCode = "Receiver"
	FaultCodeDataEncodingUnknown SOAPFaultCode = "DataEncodingUnknown"
)

// NewSOAPFault creates a new SOAP 1.1 fault
func NewSOAPFault(code SOAPFaultCode, message string) *SOAPFault {
	return &SOAPFault{
		Code:   string(code),
		String: message,
	}
}

// WithActor adds an actor to a SOAP fault
func (f *SOAPFault) WithActor(actor string) *SOAPFault {
	f.Actor = actor
	return f
}

// WithDetail adds detail to a SOAP fault
func (f *SOAPFault) WithDetail(detail string) *SOAPFault {
	f.Detail = detail
	return f
}

// NewSOAP12Fault creates a new SOAP 1.2 fault
func NewSOAP12Fault(code SOAPFaultCode, message string) *SOAP12Fault {
	return &SOAP12Fault{
		Code: SOAP12FaultCode{
			Value: string(code),
		},
		Reason: SOAP12FaultReason{
			Text: message,
		},
	}
}

// WithNode adds a node to a SOAP 1.2 fault
func (f *SOAP12Fault) WithNode(node string) *SOAP12Fault {
	f.Node = node
	return f
}

// WithRole adds a role to a SOAP 1.2 fault
func (f *SOAP12Fault) WithRole(role string) *SOAP12Fault {
	f.Role = role
	return f
}

// WithDetail adds detail to a SOAP 1.2 fault
func (f *SOAP12Fault) WithDetail(detail string) *SOAP12Fault {
	f.Detail = detail
	return f
}

// Common SOAP errors
var (
	ErrSOAPInvalidRequest     = NewSOAPFault(FaultCodeClient, "Invalid SOAP request")
	ErrSOAPUnauthenticated    = NewSOAPFault(FaultCodeClient, "Authentication required")
	ErrSOAPPermissionDenied   = NewSOAPFault(FaultCodeClient, "Permission denied")
	ErrSOAPNotFound           = NewSOAPFault(FaultCodeClient, "Service or method not found")
	ErrSOAPRateLimit          = NewSOAPFault(FaultCodeClient, "Rate limit exceeded")
	ErrSOAPInternalError      = NewSOAPFault(FaultCodeServer, "Internal server error")
	ErrSOAPServiceUnavailable = NewSOAPFault(FaultCodeServer, "Service unavailable")
	ErrSOAPTimeout            = NewSOAPFault(FaultCodeServer, "Request timeout")
)

// Helper functions for SOAP envelope handling

// ParseSOAPEnvelope parses a SOAP envelope from XML data
func ParseSOAPEnvelope(data []byte) (*SOAPEnvelope, SOAPVersion, error) {
	if len(data) == 0 {
		return nil, SOAP11, fmt.Errorf("empty request body")
	}

	// Try SOAP 1.1 first
	var env11 SOAPEnvelope
	if err := xml.Unmarshal(data, &env11); err == nil {
		return &env11, SOAP11, nil
	}

	// Try SOAP 1.2
	var env12 SOAP12Envelope
	if err := xml.Unmarshal(data, &env12); err == nil {
		// Convert SOAP 1.2 to SOAP 1.1 format for unified handling
		env := &SOAPEnvelope{
			Body: SOAPBody{
				Content: env12.Body.Content,
			},
		}
		if env12.Header != nil {
			env.Header = &SOAPHeader{
				Content: env12.Header.Content,
			}
		}
		if env12.Body.Fault != nil {
			env.Body.Fault = &SOAPFault{
				Code:   env12.Body.Fault.Code.Value,
				String: env12.Body.Fault.Reason.Text,
				Detail: env12.Body.Fault.Detail,
			}
		}
		return env, SOAP12, nil
	}

	return nil, SOAP11, fmt.Errorf("invalid SOAP envelope")
}

// MarshalSOAPEnvelope marshals a SOAP envelope to XML
func MarshalSOAPEnvelope(env *SOAPEnvelope, version SOAPVersion) ([]byte, error) {
	if version == SOAP12 {
		// Convert to SOAP 1.2 format
		env12 := &SOAP12Envelope{
			Body: SOAP12Body{
				Content: env.Body.Content,
			},
		}
		if env.Header != nil {
			env12.Header = &SOAP12Header{
				Content: env.Header.Content,
			}
		}
		if env.Body.Fault != nil {
			env12.Body.Fault = &SOAP12Fault{
				Code: SOAP12FaultCode{
					Value: env.Body.Fault.Code,
				},
				Reason: SOAP12FaultReason{
					Text: env.Body.Fault.String,
				},
				Detail: env.Body.Fault.Detail,
			}
		}
		return xml.MarshalIndent(env12, "", "  ")
	}

	// SOAP 1.1
	return xml.MarshalIndent(env, "", "  ")
}

// MarshalSOAPFault marshals a SOAP fault to XML
func MarshalSOAPFault(fault *SOAPFault, version SOAPVersion) ([]byte, error) {
	env := &SOAPEnvelope{
		Body: SOAPBody{
			Fault: fault,
		},
	}
	return MarshalSOAPEnvelope(env, version)
}

// DetectSOAPVersion detects SOAP version from content type or envelope
func DetectSOAPVersion(contentType string, data []byte) SOAPVersion {
	// Check content type first
	if contentType == "application/soap+xml" {
		return SOAP12
	}

	// Try to detect from envelope namespace
	if len(data) > 0 {
		dataStr := string(data)
		if containsSOAPNamespace(dataStr, "http://www.w3.org/2003/05/soap-envelope") {
			return SOAP12
		}
	}

	// Default to SOAP 1.1
	return SOAP11
}

// containsSOAPNamespace checks if a string contains a SOAP namespace substring
func containsSOAPNamespace(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSOAPMiddle(s, substr)))
}

func containsSOAPMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
