package pkg

// HandlerFunc represents a request handler function
type HandlerFunc func(ctx Context) error

// MiddlewareFunc represents a middleware function
type MiddlewareFunc func(ctx Context, next HandlerFunc) error

// WebSocketHandler represents a WebSocket handler function
type WebSocketHandler func(ctx Context, conn WebSocketConnection) error

// RouterEngine defines the routing interface for the framework
type RouterEngine interface {
	// HTTP method routing
	GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
	OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine

	// Generic method routing
	Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine

	// Route groups
	Group(prefix string, middleware ...MiddlewareFunc) RouterEngine

	// Host-specific routing for multi-tenancy
	Host(hostname string) RouterEngine

	// Static file serving
	Static(prefix string, filesystem VirtualFS) RouterEngine
	StaticFile(path, filepath string) RouterEngine

	// WebSocket routing
	WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine

	// API protocol routing
	GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
	GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
	SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine

	// Middleware management
	Use(middleware ...MiddlewareFunc) RouterEngine

	// Route matching
	Match(method, path, host string) (*Route, map[string]string, bool)

	// Route information
	Routes() []*Route
}

// Route represents a registered route
type Route struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Middleware  []MiddlewareFunc
	Host        string
	Name        string
	IsWebSocket bool
	IsStatic    bool

	// WebSocket-specific fields
	WebSocketHandler WebSocketHandler

	// API-specific fields
	GraphQLSchema GraphQLSchema
	GRPCService   GRPCService
	SOAPService   SOAPService
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection interface {
	ReadMessage() (messageType int, data []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
	RemoteAddr() string
	LocalAddr() string
}

// GraphQLSchema represents a GraphQL schema
type GraphQLSchema interface {
	Execute(query string, variables map[string]interface{}) (interface{}, error)
}

// GRPCService represents a gRPC service
type GRPCService interface {
	ServiceName() string
	Methods() []string
}

// SOAPService represents a SOAP service
type SOAPService interface {
	ServiceName() string
	WSDL() (string, error)
	Execute(action string, body []byte) ([]byte, error)
}
