package pkg

import (
	"regexp"
	"strings"
	"sync"
)

// router implements the RouterEngine interface
type router struct {
	routes     *[]*Route           // Pointer to shared routes slice
	hosts      *map[string]*router // Pointer to shared hosts map
	prefix     string              // Group prefix
	middleware []MiddlewareFunc    // Group middleware
	mu         *sync.RWMutex       // Pointer to shared mutex
}

// NewRouter creates a new router instance
func NewRouter() RouterEngine {
	routes := make([]*Route, 0)
	hosts := make(map[string]*router)
	mu := &sync.RWMutex{}

	return &router{
		routes: &routes,
		hosts:  &hosts,
		mu:     mu,
	}
}

// GET registers a GET route
func (r *router) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("GET", path, handler, middleware...)
}

// POST registers a POST route
func (r *router) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("POST", path, handler, middleware...)
}

// PUT registers a PUT route
func (r *router) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("PUT", path, handler, middleware...)
}

// DELETE registers a DELETE route
func (r *router) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("DELETE", path, handler, middleware...)
}

// PATCH registers a PATCH route
func (r *router) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("PATCH", path, handler, middleware...)
}

// HEAD registers a HEAD route
func (r *router) HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("HEAD", path, handler, middleware...)
}

// OPTIONS registers an OPTIONS route
func (r *router) OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	return r.Handle("OPTIONS", path, handler, middleware...)
}

// Handle registers a route with any HTTP method
func (r *router) Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Apply prefix if in a group
	fullPath := r.prefix + path

	// Combine group middleware with route middleware
	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	route := &Route{
		Method:     method,
		Path:       fullPath,
		Handler:    handler,
		Middleware: allMiddleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// Group creates a route group with a common prefix and middleware
func (r *router) Group(prefix string, middleware ...MiddlewareFunc) RouterEngine {
	// Create a new middleware slice to avoid modifying the parent
	groupMiddleware := make([]MiddlewareFunc, len(r.middleware))
	copy(groupMiddleware, r.middleware)
	groupMiddleware = append(groupMiddleware, middleware...)

	return &router{
		routes:     r.routes, // Share the same routes slice
		hosts:      r.hosts,  // Share the same hosts map
		prefix:     r.prefix + prefix,
		middleware: groupMiddleware,
		mu:         r.mu, // Share the same mutex
	}
}

// Host creates a host-specific router for multi-tenancy
func (r *router) Host(hostname string) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hostRouter, exists := (*r.hosts)[hostname]; exists {
		return hostRouter
	}

	routes := make([]*Route, 0)
	hosts := make(map[string]*router)
	mu := &sync.RWMutex{}

	hostRouter := &router{
		routes:     &routes,
		hosts:      &hosts,
		middleware: []MiddlewareFunc{},
		mu:         mu,
	}

	(*r.hosts)[hostname] = hostRouter
	return hostRouter
}

// Static registers a static file serving route
func (r *router) Static(prefix string, filesystem VirtualFS) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + prefix

	route := &Route{
		Method:   "GET",
		Path:     fullPath + "/*filepath",
		IsStatic: true,
		Handler: func(ctx Context) error {
			// Extract filepath from params
			filepath := ctx.Params()["filepath"]
			if filepath == "" {
				filepath = "index.html"
			}

			// Open file from virtual filesystem
			file, err := filesystem.Open(filepath)
			if err != nil {
				return err
			}
			defer file.Close()

			// Serve the file
			// This is a simplified implementation
			// Real implementation would handle content types, caching, etc.
			return ctx.Response().WriteStream(200, "application/octet-stream", file)
		},
		Middleware: r.middleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// StaticFile registers a single static file route
func (r *router) StaticFile(path, filepath string) RouterEngine {
	// This would be implemented with actual file system access
	// For now, we'll create a placeholder
	return r.Handle("GET", path, func(ctx Context) error {
		return ctx.String(200, "Static file: "+filepath)
	})
}

// WebSocket registers a WebSocket route
func (r *router) WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	route := &Route{
		Method:           "GET",
		Path:             fullPath,
		IsWebSocket:      true,
		WebSocketHandler: handler,
		Handler: func(ctx Context) error {
			// The actual WebSocket upgrade will be handled by the WebSocketServer
			// This handler is a placeholder that indicates WebSocket support
			return nil
		},
		Middleware: allMiddleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// GraphQL registers a GraphQL endpoint
func (r *router) GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	route := &Route{
		Method:        "POST",
		Path:          fullPath,
		GraphQLSchema: schema,
		Handler: func(ctx Context) error {
			// GraphQL query execution would happen here
			return ctx.JSON(200, map[string]string{"status": "graphql"})
		},
		Middleware: allMiddleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// GRPC registers a gRPC service
func (r *router) GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	route := &Route{
		Method:      "POST",
		Path:        "/grpc/" + service.ServiceName(),
		GRPCService: service,
		Handler: func(ctx Context) error {
			// gRPC handling would happen here
			return ctx.JSON(200, map[string]string{"status": "grpc"})
		},
		Middleware: allMiddleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// SOAP registers a SOAP service
func (r *router) SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	allMiddleware := append([]MiddlewareFunc{}, r.middleware...)
	allMiddleware = append(allMiddleware, middleware...)

	route := &Route{
		Method:      "POST",
		Path:        fullPath,
		SOAPService: service,
		Handler: func(ctx Context) error {
			// SOAP handling would happen here
			return ctx.JSON(200, map[string]string{"status": "soap"})
		},
		Middleware: allMiddleware,
	}

	*r.routes = append(*r.routes, route)
	return r
}

// Use adds middleware to the router
func (r *router) Use(middleware ...MiddlewareFunc) RouterEngine {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middleware = append(r.middleware, middleware...)
	return r
}

// Match finds a route that matches the given method, path, and host
func (r *router) Match(method, path, host string) (*Route, map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First, check if there's a host-specific router
	if host != "" {
		if hostRouter, exists := (*r.hosts)[host]; exists {
			if route, params, found := hostRouter.Match(method, path, ""); found {
				return route, params, true
			}
		}
	}

	// Then check routes in this router
	for _, route := range *r.routes {
		if route.Method != method && route.Method != "*" {
			continue
		}

		// Check if path matches
		params, matches := matchPath(route.Path, path)
		if matches {
			return route, params, true
		}
	}

	return nil, nil, false
}

// Routes returns all registered routes
func (r *router) Routes() []*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*Route, len(*r.routes))
	copy(routes, *r.routes)

	// Add routes from host-specific routers
	for host, hostRouter := range *r.hosts {
		hostRoutes := hostRouter.Routes()
		for _, route := range hostRoutes {
			routeCopy := *route
			routeCopy.Host = host
			routes = append(routes, &routeCopy)
		}
	}

	return routes
}

// matchPath matches a route path pattern against an actual path
// Returns extracted parameters and whether the path matches
func matchPath(pattern, path string) (map[string]string, bool) {
	params := make(map[string]string)

	// Handle exact match
	if pattern == path {
		return params, true
	}

	// Split pattern and path into segments
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Handle wildcard at the end
	if len(patternParts) > 0 && strings.HasPrefix(patternParts[len(patternParts)-1], "*") {
		if len(pathParts) < len(patternParts)-1 {
			return nil, false
		}

		// Match all parts before wildcard
		for i := 0; i < len(patternParts)-1; i++ {
			if !matchSegment(patternParts[i], pathParts[i], params) {
				return nil, false
			}
		}

		// Capture wildcard
		wildcardName := strings.TrimPrefix(patternParts[len(patternParts)-1], "*")
		if wildcardName == "" {
			wildcardName = "wildcard"
		}
		params[wildcardName] = strings.Join(pathParts[len(patternParts)-1:], "/")
		return params, true
	}

	// Must have same number of segments
	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	// Match each segment
	for i := 0; i < len(patternParts); i++ {
		if !matchSegment(patternParts[i], pathParts[i], params) {
			return nil, false
		}
	}

	return params, true
}

// matchSegment matches a single path segment
func matchSegment(pattern, segment string, params map[string]string) bool {
	// Parameter segment (starts with :)
	if strings.HasPrefix(pattern, ":") {
		paramName := strings.TrimPrefix(pattern, ":")
		params[paramName] = segment
		return true
	}

	// Regex segment (contains regex pattern)
	if strings.Contains(pattern, "(") && strings.Contains(pattern, ")") {
		// Extract parameter name and regex
		parts := strings.SplitN(pattern, "(", 2)
		paramName := strings.TrimPrefix(parts[0], ":")
		regexPattern := strings.TrimSuffix(parts[1], ")")

		// Compile and match regex
		re, err := regexp.Compile("^" + regexPattern + "$")
		if err != nil {
			return false
		}

		if re.MatchString(segment) {
			if paramName != "" {
				params[paramName] = segment
			}
			return true
		}
		return false
	}

	// Exact match
	return pattern == segment
}
