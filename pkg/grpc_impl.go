package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// grpcManager implements the GRPCManager interface
// Requirements: 2.3
type grpcManager struct {
	router      RouterEngine
	db          DatabaseManager
	authManager *AuthManager
	middleware  []GRPCMiddleware
	prefix      string
	rateLimiter *rateLimiter
	services    map[string]GRPCService
	configs     map[string]GRPCConfig
	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
}

// NewGRPCManager creates a new gRPC manager
// Requirements: 2.3
func NewGRPCManager(router RouterEngine, db DatabaseManager, authManager *AuthManager) GRPCManager {
	return &grpcManager{
		router:      router,
		db:          db,
		authManager: authManager,
		middleware:  make([]GRPCMiddleware, 0),
		prefix:      "",
		rateLimiter: newRateLimiter(db),
		services:    make(map[string]GRPCService),
		configs:     make(map[string]GRPCConfig),
		stopChan:    make(chan struct{}),
	}
}

// RegisterService registers a gRPC service with configuration
// Requirements: 2.3
func (g *grpcManager) RegisterService(service GRPCService, config GRPCConfig) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	serviceName := service.ServiceName()
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Check if service is already registered
	if _, exists := g.services[serviceName]; exists {
		return fmt.Errorf("service %s is already registered", serviceName)
	}

	// Store service and config
	g.services[serviceName] = service
	g.configs[serviceName] = config

	// Register routes for each method in the service
	for _, method := range service.Methods() {
		fullMethod := fmt.Sprintf("%s/%s/%s", g.prefix, serviceName, method)

		// Create handler for this method
		handler := g.createMethodHandler(service, method, config)

		// Register with router using POST (gRPC uses HTTP/2 POST)
		g.router.POST(fullMethod, handler)
	}

	return nil
}

// createMethodHandler creates a handler for a specific gRPC method
func (g *grpcManager) createMethodHandler(service GRPCService, method string, config GRPCConfig) HandlerFunc {
	return func(ctx Context) error {
		// Wrap with middleware chain
		handler := g.wrapHandler(service, method, config)
		return handler(ctx)
	}
}

// wrapHandler wraps a gRPC method with middleware and configuration
func (g *grpcManager) wrapHandler(service GRPCService, method string, config GRPCConfig) GRPCHandler {
	// Create the base handler that executes gRPC methods
	handler := func(ctx Context) error {
		// Parse gRPC request
		req, err := g.parseRequest(ctx, service, method)
		if err != nil {
			return g.sendErrorResponse(ctx, ErrGRPCInvalidArgument.WithDetails(err.Error()))
		}

		// Execute method
		var result interface{}
		if extService, ok := service.(GRPCServiceExtended); ok {
			result, err = extService.HandleUnary(context.Background(), method, req.Message)
		} else {
			// Fallback: service doesn't implement extended interface
			err = fmt.Errorf("service does not support unary calls")
		}

		if err != nil {
			return g.sendErrorResponse(ctx, NewGRPCError(GRPCStatusInternal, err.Error()))
		}

		// Build gRPC response
		response := &GRPCResponse{
			Message:  result,
			Metadata: make(map[string]string),
			Trailer:  make(map[string]string),
		}

		// Send response
		return ctx.JSON(200, response)
	}

	// Apply rate limiting middleware if configured
	// Requirements: 2.6
	if config.RateLimit != nil {
		handler = g.rateLimitMiddleware(config.RateLimit, handler)
	}

	if config.GlobalRateLimit != nil {
		handler = g.globalRateLimitMiddleware(config.GlobalRateLimit, handler)
	}

	// Apply authentication middleware if required
	// Requirements: 2.5
	if config.RequireAuth {
		handler = g.authMiddleware(config, handler)
	}

	// Apply request validation middleware
	if config.MaxRequestSize > 0 || config.Timeout > 0 {
		handler = g.validationMiddleware(config, handler)
	}

	// Apply custom middleware in order
	for i := len(g.middleware) - 1; i >= 0; i-- {
		mw := g.middleware[i]
		currentHandler := handler
		handler = func(ctx Context) error {
			return mw(ctx, currentHandler)
		}
	}

	return handler
}

// parseRequest parses a gRPC request from the context
func (g *grpcManager) parseRequest(ctx Context, service GRPCService, method string) (*GRPCRequest, error) {
	// Extract metadata from headers
	metadata := make(map[string]string)
	headers := ctx.Request().Header
	for key, values := range headers {
		if len(values) > 0 {
			metadata[key] = values[0]
		}
	}

	// Parse request body
	body := ctx.Body()
	if len(body) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	// For now, we'll store the raw body as the message
	// In a real implementation, this would be protobuf-decoded
	req := &GRPCRequest{
		Service:    service.ServiceName(),
		Method:     method,
		Message:    body,
		Metadata:   metadata,
		RemoteAddr: ctx.Request().RemoteAddr,
	}

	return req, nil
}

// rateLimitMiddleware applies rate limiting per resource
// Requirements: 2.6
func (g *grpcManager) rateLimitMiddleware(config *GRPCRateLimitConfig, next GRPCHandler) GRPCHandler {
	return func(ctx Context) error {
		// Build rate limit key
		key := g.buildRateLimitKey(ctx, config.Key)

		// Check rate limit
		allowed, err := g.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return g.sendErrorResponse(ctx, NewGRPCError(GRPCStatusInternal, "Rate limit check failed").WithDetails(err.Error()))
		}

		if !allowed {
			return g.sendErrorResponse(ctx, ErrGRPCResourceExhausted.WithDetails(
				fmt.Sprintf("limit: %d, window: %s", config.Limit, config.Window.String()),
			))
		}

		// Increment rate limit counter
		if err := g.rateLimiter.Increment(key, config.Window); err != nil {
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
func (g *grpcManager) globalRateLimitMiddleware(config *GRPCRateLimitConfig, next GRPCHandler) GRPCHandler {
	return func(ctx Context) error {
		key := "global:" + g.buildRateLimitKey(ctx, config.Key)

		allowed, err := g.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return g.sendErrorResponse(ctx, NewGRPCError(GRPCStatusInternal, "Global rate limit check failed").WithDetails(err.Error()))
		}

		if !allowed {
			return g.sendErrorResponse(ctx, ErrGRPCResourceExhausted.WithDetails(
				fmt.Sprintf("global limit: %d, window: %s", config.Limit, config.Window.String()),
			))
		}

		if err := g.rateLimiter.Increment(key, config.Window); err != nil {
			if ctx.Logger() != nil {
				ctx.Logger().Error("Failed to increment global rate limit", "error", err)
			}
		}

		return next(ctx)
	}
}

// authMiddleware validates authentication and authorization
// Requirements: 2.5
func (g *grpcManager) authMiddleware(config GRPCConfig, next GRPCHandler) GRPCHandler {
	return func(ctx Context) error {
		if !ctx.IsAuthenticated() {
			return g.sendErrorResponse(ctx, ErrGRPCUnauthenticated)
		}

		user := ctx.User()
		if user == nil {
			return g.sendErrorResponse(ctx, ErrGRPCUnauthenticated.WithDetails("user not found"))
		}

		// Check required roles if specified
		if len(config.RequiredRoles) > 0 && g.authManager != nil {
			if err := g.authManager.AuthorizeRoles(user, config.RequiredRoles); err != nil {
				return g.sendErrorResponse(ctx, ErrGRPCPermissionDenied.WithDetails(
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
				return g.sendErrorResponse(ctx, ErrGRPCPermissionDenied.WithDetails(
					fmt.Sprintf("required_scopes: %v, user_scopes: %v", config.RequiredScopes, user.Roles),
				))
			}
		}

		return next(ctx)
	}
}

// validationMiddleware validates request size and timeout
func (g *grpcManager) validationMiddleware(config GRPCConfig, next GRPCHandler) GRPCHandler {
	return func(ctx Context) error {
		// Validate request size
		if config.MaxRequestSize > 0 {
			body := ctx.Body()
			if int64(len(body)) > config.MaxRequestSize {
				return g.sendErrorResponse(ctx, ErrGRPCInvalidArgument.WithDetails(
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

// buildRateLimitKey builds a rate limit key based on the key type
func (g *grpcManager) buildRateLimitKey(ctx Context, keyType string) string {
	switch keyType {
	case "user_id":
		if user := ctx.User(); user != nil {
			return fmt.Sprintf("grpc:user:%s", user.ID)
		}
		return "grpc:user:anonymous"
	case "tenant_id":
		if tenant := ctx.Tenant(); tenant != nil {
			return fmt.Sprintf("grpc:tenant:%s", tenant.ID)
		}
		return "grpc:tenant:default"
	case "ip_address":
		return fmt.Sprintf("grpc:ip:%s", ctx.Request().RemoteAddr)
	default:
		// Default to IP address
		return fmt.Sprintf("grpc:ip:%s", ctx.Request().RemoteAddr)
	}
}

// CheckRateLimit checks rate limit for a specific resource
// Requirements: 2.6
func (g *grpcManager) CheckRateLimit(ctx Context, resource string) error {
	// Default rate limit: 100 requests per minute
	key := fmt.Sprintf("grpc:resource:%s:%s", resource, ctx.Request().RemoteAddr)
	allowed, err := g.rateLimiter.Check(key, 100, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("rate limit exceeded for resource: %s", resource)
	}

	return g.rateLimiter.Increment(key, time.Minute)
}

// CheckGlobalRateLimit checks global rate limit
// Requirements: 2.6
func (g *grpcManager) CheckGlobalRateLimit(ctx Context) error {
	// Default global rate limit: 1000 requests per minute
	key := fmt.Sprintf("grpc:global:%s", ctx.Request().RemoteAddr)
	allowed, err := g.rateLimiter.Check(key, 1000, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("global rate limit exceeded")
	}

	return g.rateLimiter.Increment(key, time.Minute)
}

// sendErrorResponse sends a gRPC error response
func (g *grpcManager) sendErrorResponse(ctx Context, grpcErr *GRPCError) error {
	response := &GRPCResponse{
		Error:    grpcErr,
		Metadata: make(map[string]string),
		Trailer:  make(map[string]string),
	}

	// Map gRPC status codes to HTTP status codes
	httpStatus := g.grpcStatusToHTTP(grpcErr.Code)

	// Send the JSON response
	ctx.JSON(httpStatus, response)

	// Return the gRPC error so it can be handled by error handlers
	return grpcErr
}

// grpcStatusToHTTP maps gRPC status codes to HTTP status codes
func (g *grpcManager) grpcStatusToHTTP(code GRPCStatusCode) int {
	switch code {
	case GRPCStatusOK:
		return 200
	case GRPCStatusCanceled:
		return 499 // Client Closed Request
	case GRPCStatusUnknown:
		return 500
	case GRPCStatusInvalidArgument:
		return 400
	case GRPCStatusDeadlineExceeded:
		return 504
	case GRPCStatusNotFound:
		return 404
	case GRPCStatusAlreadyExists:
		return 409
	case GRPCStatusPermissionDenied:
		return 403
	case GRPCStatusResourceExhausted:
		return 429
	case GRPCStatusFailedPrecondition:
		return 400
	case GRPCStatusAborted:
		return 409
	case GRPCStatusOutOfRange:
		return 400
	case GRPCStatusUnimplemented:
		return 501
	case GRPCStatusInternal:
		return 500
	case GRPCStatusUnavailable:
		return 503
	case GRPCStatusDataLoss:
		return 500
	case GRPCStatusUnauthenticated:
		return 401
	default:
		return 500
	}
}

// Use adds middleware to the gRPC manager
func (g *grpcManager) Use(middleware GRPCMiddleware) GRPCManager {
	g.middleware = append(g.middleware, middleware)
	return g
}

// Group creates a new gRPC manager with a prefix
func (g *grpcManager) Group(prefix string, middleware ...GRPCMiddleware) GRPCManager {
	newManager := &grpcManager{
		router:      g.router,
		db:          g.db,
		authManager: g.authManager,
		middleware:  append(g.middleware, middleware...),
		prefix:      g.prefix + prefix,
		rateLimiter: g.rateLimiter,
		services:    g.services,
		configs:     g.configs,
		mu:          sync.RWMutex{},
		stopChan:    make(chan struct{}),
	}
	return newManager
}

// Start starts the gRPC server
// Requirements: 2.3
func (g *grpcManager) Start(addr string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.running {
		return fmt.Errorf("gRPC server is already running")
	}

	g.running = true

	// In a real implementation, this would start a gRPC server
	// For now, we just mark it as running since routes are registered with the HTTP router

	return nil
}

// Stop stops the gRPC server immediately
func (g *grpcManager) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running {
		return fmt.Errorf("gRPC server is not running")
	}

	close(g.stopChan)
	g.running = false

	return nil
}

// GracefulStop gracefully stops the gRPC server with a timeout
func (g *grpcManager) GracefulStop(timeout time.Duration) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.running {
		return fmt.Errorf("gRPC server is not running")
	}

	// Create a timer for the timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Signal shutdown
	close(g.stopChan)

	// Wait for timeout or immediate shutdown
	select {
	case <-timer.C:
		// Timeout reached, force shutdown
		g.running = false
		return fmt.Errorf("graceful shutdown timeout exceeded")
	default:
		// Immediate shutdown
		g.running = false
		return nil
	}
}
