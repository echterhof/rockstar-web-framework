package pkg

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// soapManager implements the SOAPManager interface
// Requirements: 2.4
type soapManager struct {
	router      RouterEngine
	db          DatabaseManager
	authManager *AuthManager
	middleware  []SOAPMiddleware
	prefix      string
	rateLimiter *rateLimiter
}

// NewSOAPManager creates a new SOAP manager
// Requirements: 2.4
func NewSOAPManager(router RouterEngine, db DatabaseManager, authManager *AuthManager) SOAPManager {
	return &soapManager{
		router:      router,
		db:          db,
		authManager: authManager,
		middleware:  make([]SOAPMiddleware, 0),
		prefix:      "",
		rateLimiter: newRateLimiter(db),
	}
}

// RegisterService registers a SOAP service with configuration
// Requirements: 2.4
func (s *soapManager) RegisterService(path string, service SOAPService, config SOAPConfig) error {
	// Build full path with prefix
	fullPath := s.prefix + path

	// Wrap handler with SOAP middleware chain
	wrappedHandler := s.wrapHandler(service, config)

	// Convert to framework handler
	frameworkHandler := func(ctx Context) error {
		return wrappedHandler(ctx)
	}

	// Register POST endpoint for SOAP requests
	s.router.POST(fullPath, frameworkHandler)

	// Register GET endpoint for WSDL if enabled
	// Requirements: 2.4
	if config.EnableWSDL {
		wsdlPath := fullPath
		if config.WSDLPath != "" {
			wsdlPath = fullPath + config.WSDLPath
		}

		s.router.GET(wsdlPath, func(ctx Context) error {
			// Check if ?wsdl query parameter is present
			query := ctx.Query()
			if _, hasWSDL := query["wsdl"]; hasWSDL || config.WSDLPath != "" {
				return s.ServeWSDL(ctx, service)
			}

			// Otherwise, handle as regular SOAP request
			return wrappedHandler(ctx)
		})
	}

	return nil
}

// wrapHandler wraps a SOAP service with middleware and configuration
func (s *soapManager) wrapHandler(service SOAPService, config SOAPConfig) SOAPHandler {
	// Create the base handler that executes SOAP operations
	handler := func(ctx Context) error {
		// Parse SOAP request
		req, err := s.parseRequest(ctx)
		if err != nil {
			return s.sendFaultResponse(ctx, SOAP11, ErrSOAPInvalidRequest.WithDetail(err.Error()))
		}

		// Execute operation
		result, err := service.Execute(req.Action, req.Body)
		if err != nil {
			return s.sendFaultResponse(ctx, req.Version, NewSOAPFault(FaultCodeServer, err.Error()))
		}

		// Build SOAP response
		response := &SOAPResponse{
			Version: req.Version,
			Body:    result,
		}

		// Send response
		return s.sendResponse(ctx, response)
	}

	// Apply rate limiting middleware if configured
	// Requirements: 2.6
	if config.RateLimit != nil {
		handler = s.rateLimitMiddleware(config.RateLimit, handler)
	}

	if config.GlobalRateLimit != nil {
		handler = s.globalRateLimitMiddleware(config.GlobalRateLimit, handler)
	}

	// Apply authentication middleware if required
	// Requirements: 2.5
	if config.RequireAuth {
		handler = s.authMiddleware(config, handler)
	}

	// Apply request validation middleware
	if config.MaxRequestSize > 0 || config.Timeout > 0 {
		handler = s.validationMiddleware(config, handler)
	}

	// Apply CORS middleware if configured
	if config.CORS != nil {
		handler = s.corsMiddleware(config.CORS, handler)
	}

	// Apply custom middleware in order
	for i := len(s.middleware) - 1; i >= 0; i-- {
		mw := s.middleware[i]
		currentHandler := handler
		handler = func(ctx Context) error {
			return mw(ctx, currentHandler)
		}
	}

	return handler
}

// parseRequest parses a SOAP request from the context
func (s *soapManager) parseRequest(ctx Context) (*SOAPRequest, error) {
	body := ctx.Body()
	if len(body) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	// Detect SOAP version
	contentType := ctx.GetHeader("Content-Type")
	version := DetectSOAPVersion(contentType, body)

	// Parse SOAP envelope
	env, detectedVersion, err := ParseSOAPEnvelope(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SOAP envelope: %w", err)
	}

	// Use detected version from envelope if different
	if detectedVersion != version {
		version = detectedVersion
	}

	// Check for fault in request
	if env.Body.Fault != nil {
		return nil, fmt.Errorf("SOAP fault in request: %s", env.Body.Fault.String)
	}

	// Extract SOAPAction header
	action := ctx.GetHeader("SOAPAction")
	action = strings.Trim(action, "\"")

	// Build request
	req := &SOAPRequest{
		Version:    version,
		Action:     action,
		Body:       env.Body.Content,
		RemoteAddr: ctx.Request().RemoteAddr,
	}

	if env.Header != nil {
		req.Header = env.Header.Content
	}

	return req, nil
}

// rateLimitMiddleware applies rate limiting per resource
// Requirements: 2.6
func (s *soapManager) rateLimitMiddleware(config *SOAPRateLimitConfig, next SOAPHandler) SOAPHandler {
	return func(ctx Context) error {
		// Build rate limit key
		key := s.buildRateLimitKey(ctx, config.Key)

		// Check rate limit
		allowed, err := s.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return s.sendFaultResponse(ctx, SOAP11, NewSOAPFault(FaultCodeServer, "Rate limit check failed").WithDetail(err.Error()))
		}

		if !allowed {
			return s.sendFaultResponse(ctx, SOAP11, ErrSOAPRateLimit.WithDetail(
				fmt.Sprintf("limit: %d, window: %s", config.Limit, config.Window.String()),
			))
		}

		// Increment rate limit counter
		if err := s.rateLimiter.Increment(key, config.Window); err != nil {
			// Log error but don't fail the request
			if ctx.Logger() != nil {
				ctx.Logger().Error("Failed to increment rate limit", "error", err)
			}
		}

		return next(ctx)
	}
}

// globalRateLimitMiddleware applies global rate limiting
// Requirements: 2.6
func (s *soapManager) globalRateLimitMiddleware(config *SOAPRateLimitConfig, next SOAPHandler) SOAPHandler {
	return func(ctx Context) error {
		key := "global:" + s.buildRateLimitKey(ctx, config.Key)

		allowed, err := s.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return s.sendFaultResponse(ctx, SOAP11, NewSOAPFault(FaultCodeServer, "Global rate limit check failed").WithDetail(err.Error()))
		}

		if !allowed {
			return s.sendFaultResponse(ctx, SOAP11, ErrSOAPRateLimit.WithDetail(
				fmt.Sprintf("global limit: %d, window: %s", config.Limit, config.Window.String()),
			))
		}

		if err := s.rateLimiter.Increment(key, config.Window); err != nil {
			if ctx.Logger() != nil {
				ctx.Logger().Error("Failed to increment global rate limit", "error", err)
			}
		}

		return next(ctx)
	}
}

// authMiddleware validates authentication and authorization
// Requirements: 2.5
func (s *soapManager) authMiddleware(config SOAPConfig, next SOAPHandler) SOAPHandler {
	return func(ctx Context) error {
		if !ctx.IsAuthenticated() {
			return s.sendFaultResponse(ctx, SOAP11, ErrSOAPUnauthenticated)
		}

		user := ctx.User()
		if user == nil {
			return s.sendFaultResponse(ctx, SOAP11, ErrSOAPUnauthenticated.WithDetail("user not found"))
		}

		// Check required roles if specified
		if len(config.RequiredRoles) > 0 && s.authManager != nil {
			if err := s.authManager.AuthorizeRoles(user, config.RequiredRoles); err != nil {
				return s.sendFaultResponse(ctx, SOAP11, ErrSOAPPermissionDenied.WithDetail(
					fmt.Sprintf("required_roles: %v, user_roles: %v", config.RequiredRoles, user.Roles),
				))
			}
		}

		// Check required scopes if specified
		if len(config.RequiredScopes) > 0 {
			hasScope := false
			for _, scope := range config.RequiredScopes {
				for _, userScope := range user.Roles {
					if userScope == scope {
						hasScope = true
						break
					}
				}
				if hasScope {
					break
				}
			}

			if !hasScope {
				return s.sendFaultResponse(ctx, SOAP11, ErrSOAPPermissionDenied.WithDetail(
					fmt.Sprintf("required_scopes: %v, user_scopes: %v", config.RequiredScopes, user.Roles),
				))
			}
		}

		return next(ctx)
	}
}

// validationMiddleware validates request size and timeout
func (s *soapManager) validationMiddleware(config SOAPConfig, next SOAPHandler) SOAPHandler {
	return func(ctx Context) error {
		// Validate request size
		if config.MaxRequestSize > 0 {
			body := ctx.Body()
			if int64(len(body)) > config.MaxRequestSize {
				return s.sendFaultResponse(ctx, SOAP11, ErrSOAPInvalidRequest.WithDetail(
					fmt.Sprintf("request too large: %d bytes (max: %d)", len(body), config.MaxRequestSize),
				))
			}
		}

		// Apply timeout if configured
		if config.Timeout > 0 {
			timeoutCtx := ctx.WithTimeout(config.Timeout)
			return next(timeoutCtx)
		}

		return next(ctx)
	}
}

// corsMiddleware applies CORS headers
func (s *soapManager) corsMiddleware(config *CORSConfig, next SOAPHandler) SOAPHandler {
	return func(ctx Context) error {
		// Set CORS headers
		if len(config.AllowOrigins) > 0 {
			origin := ctx.GetHeader("Origin")
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					ctx.SetHeader("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}

		if len(config.AllowMethods) > 0 {
			ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			ctx.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			ctx.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			ctx.SetHeader("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight requests
		if ctx.Request().Method == "OPTIONS" {
			return ctx.String(204, "")
		}

		return next(ctx)
	}
}

// buildRateLimitKey builds a rate limit key based on the key type
func (s *soapManager) buildRateLimitKey(ctx Context, keyType string) string {
	switch keyType {
	case "user_id":
		if user := ctx.User(); user != nil {
			return fmt.Sprintf("soap:user:%s", user.ID)
		}
		return "soap:user:anonymous"
	case "tenant_id":
		if tenant := ctx.Tenant(); tenant != nil {
			return fmt.Sprintf("soap:tenant:%s", tenant.ID)
		}
		return "soap:tenant:default"
	case "ip_address":
		return fmt.Sprintf("soap:ip:%s", ctx.Request().RemoteAddr)
	default:
		// Default to IP address
		return fmt.Sprintf("soap:ip:%s", ctx.Request().RemoteAddr)
	}
}

// CheckRateLimit checks rate limit for a specific resource
// Requirements: 2.6
func (s *soapManager) CheckRateLimit(ctx Context, resource string) error {
	// Default rate limit: 100 requests per minute
	key := fmt.Sprintf("soap:resource:%s:%s", resource, ctx.Request().RemoteAddr)
	allowed, err := s.rateLimiter.Check(key, 100, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("rate limit exceeded for resource: %s", resource)
	}

	return s.rateLimiter.Increment(key, time.Minute)
}

// CheckGlobalRateLimit checks global rate limit
// Requirements: 2.6
func (s *soapManager) CheckGlobalRateLimit(ctx Context) error {
	// Default global rate limit: 1000 requests per minute
	key := fmt.Sprintf("soap:global:%s", ctx.Request().RemoteAddr)
	allowed, err := s.rateLimiter.Check(key, 1000, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("global rate limit exceeded")
	}

	return s.rateLimiter.Increment(key, time.Minute)
}

// sendResponse sends a SOAP response
func (s *soapManager) sendResponse(ctx Context, response *SOAPResponse) error {
	// Build SOAP envelope
	env := &SOAPEnvelope{
		Body: SOAPBody{
			Content: response.Body,
		},
	}

	if response.Header != nil {
		env.Header = &SOAPHeader{
			Content: response.Header,
		}
	}

	// Marshal to XML
	xmlData, err := MarshalSOAPEnvelope(env, response.Version)
	if err != nil {
		return s.sendFaultResponse(ctx, response.Version, NewSOAPFault(FaultCodeServer, "Failed to marshal response"))
	}

	// Set content type based on SOAP version
	if response.Version == SOAP12 {
		ctx.SetHeader("Content-Type", "application/soap+xml; charset=utf-8")
	} else {
		ctx.SetHeader("Content-Type", "text/xml; charset=utf-8")
	}

	return ctx.String(200, string(xmlData))
}

// sendFaultResponse sends a SOAP fault response
func (s *soapManager) sendFaultResponse(ctx Context, version SOAPVersion, fault *SOAPFault) error {
	// Marshal fault to XML
	xmlData, err := MarshalSOAPFault(fault, version)
	if err != nil {
		// Fallback to simple error response
		return ctx.String(500, "Internal Server Error")
	}

	// Set content type based on SOAP version
	if version == SOAP12 {
		ctx.SetHeader("Content-Type", "application/soap+xml; charset=utf-8")
	} else {
		ctx.SetHeader("Content-Type", "text/xml; charset=utf-8")
	}

	// SOAP faults are returned with 500 status code
	return ctx.String(500, string(xmlData))
}

// ServeWSDL serves the WSDL document for a SOAP service
// Requirements: 2.4
func (s *soapManager) ServeWSDL(ctx Context, service SOAPService) error {
	// Get WSDL from service
	wsdl, err := service.WSDL()
	if err != nil {
		return ctx.String(500, fmt.Sprintf("Failed to generate WSDL: %s", err.Error()))
	}

	// Set content type and write response
	ctx.Response().SetContentType("text/xml; charset=utf-8")
	ctx.Response().WriteHeader(200)
	_, err = ctx.Response().Write([]byte(wsdl))
	return err
}

// Use adds middleware to the SOAP manager
func (s *soapManager) Use(middleware SOAPMiddleware) SOAPManager {
	s.middleware = append(s.middleware, middleware)
	return s
}

// Group creates a new SOAP manager with a prefix
func (s *soapManager) Group(prefix string, middleware ...SOAPMiddleware) SOAPManager {
	newManager := &soapManager{
		router:      s.router,
		db:          s.db,
		authManager: s.authManager,
		middleware:  append(s.middleware, middleware...),
		prefix:      s.prefix + prefix,
		rateLimiter: s.rateLimiter,
	}
	return newManager
}

// GenerateWSDL generates a basic WSDL document for a SOAP service
// Requirements: 2.4
func GenerateWSDL(config SOAPConfig, endpoint string, operations []WSDLOperation) (string, error) {
	wsdl := &WSDLDefinitions{
		Name:            config.ServiceName,
		TargetNamespace: config.Namespace,
		Xmlns:           "http://schemas.xmlsoap.org/wsdl/",
		XmlnsSoap:       "http://schemas.xmlsoap.org/wsdl/soap/",
		XmlnsXsd:        "http://www.w3.org/2001/XMLSchema",
		XmlnsTns:        config.Namespace,
	}

	// Add port type
	wsdl.PortType = WSDLPortType{
		Name:       config.PortName,
		Operations: make([]WSDLPortOperation, len(operations)),
	}

	for i, op := range operations {
		wsdl.PortType.Operations[i] = WSDLPortOperation{
			Name:   op.Name,
			Input:  WSDLMessage{Message: "tns:" + op.Name + "Request"},
			Output: WSDLMessage{Message: "tns:" + op.Name + "Response"},
		}
	}

	// Add binding
	wsdl.Binding = WSDLBinding{
		Name: config.ServiceName + "Binding",
		Type: "tns:" + config.PortName,
		SoapBinding: WSDLSoapBinding{
			Style:     "document",
			Transport: "http://schemas.xmlsoap.org/soap/http",
		},
		Operations: make([]WSDLBindingOperation, len(operations)),
	}

	for i, op := range operations {
		wsdl.Binding.Operations[i] = WSDLBindingOperation{
			Name: op.Name,
			SoapOperation: WSDLSoapOperation{
				SoapAction: config.Namespace + "/" + op.Name,
			},
			Input: WSDLSoapBody{
				Use: "literal",
			},
			Output: WSDLSoapBody{
				Use: "literal",
			},
		}
	}

	// Add service
	wsdl.Service = WSDLService{
		Name: config.ServiceName,
		Port: WSDLPort{
			Name:    config.ServiceName + "Port",
			Binding: "tns:" + config.ServiceName + "Binding",
			Address: WSDLSoapAddress{
				Location: endpoint,
			},
		},
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(wsdl, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal WSDL: %w", err)
	}

	return xml.Header + string(xmlData), nil
}

// WSDLDefinitions represents a WSDL document
type WSDLDefinitions struct {
	XMLName         xml.Name     `xml:"definitions"`
	Name            string       `xml:"name,attr"`
	TargetNamespace string       `xml:"targetNamespace,attr"`
	Xmlns           string       `xml:"xmlns,attr"`
	XmlnsSoap       string       `xml:"xmlns:soap,attr"`
	XmlnsXsd        string       `xml:"xmlns:xsd,attr"`
	XmlnsTns        string       `xml:"xmlns:tns,attr"`
	PortType        WSDLPortType `xml:"portType"`
	Binding         WSDLBinding  `xml:"binding"`
	Service         WSDLService  `xml:"service"`
}

// WSDLPortType represents a WSDL port type
type WSDLPortType struct {
	Name       string              `xml:"name,attr"`
	Operations []WSDLPortOperation `xml:"operation"`
}

// WSDLPortOperation represents a WSDL port operation
type WSDLPortOperation struct {
	Name   string      `xml:"name,attr"`
	Input  WSDLMessage `xml:"input"`
	Output WSDLMessage `xml:"output"`
}

// WSDLMessage represents a WSDL message reference
type WSDLMessage struct {
	Message string `xml:"message,attr"`
}

// WSDLBinding represents a WSDL binding
type WSDLBinding struct {
	Name        string                 `xml:"name,attr"`
	Type        string                 `xml:"type,attr"`
	SoapBinding WSDLSoapBinding        `xml:"soap:binding"`
	Operations  []WSDLBindingOperation `xml:"operation"`
}

// WSDLSoapBinding represents SOAP binding configuration
type WSDLSoapBinding struct {
	Style     string `xml:"style,attr"`
	Transport string `xml:"transport,attr"`
}

// WSDLBindingOperation represents a WSDL binding operation
type WSDLBindingOperation struct {
	Name          string            `xml:"name,attr"`
	SoapOperation WSDLSoapOperation `xml:"soap:operation"`
	Input         WSDLSoapBody      `xml:"input>soap:body"`
	Output        WSDLSoapBody      `xml:"output>soap:body"`
}

// WSDLSoapOperation represents SOAP operation configuration
type WSDLSoapOperation struct {
	SoapAction string `xml:"soapAction,attr"`
}

// WSDLSoapBody represents SOAP body configuration
type WSDLSoapBody struct {
	Use string `xml:"use,attr"`
}

// WSDLService represents a WSDL service
type WSDLService struct {
	Name string   `xml:"name,attr"`
	Port WSDLPort `xml:"port"`
}

// WSDLPort represents a WSDL port
type WSDLPort struct {
	Name    string          `xml:"name,attr"`
	Binding string          `xml:"binding,attr"`
	Address WSDLSoapAddress `xml:"soap:address"`
}

// WSDLSoapAddress represents SOAP address configuration
type WSDLSoapAddress struct {
	Location string `xml:"location,attr"`
}

// WSDLOperation represents an operation for WSDL generation
type WSDLOperation struct {
	Name        string
	InputType   string
	OutputType  string
	Description string
}
