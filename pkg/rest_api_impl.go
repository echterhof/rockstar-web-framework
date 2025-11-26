package pkg

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// restAPIManager implements the RESTAPIManager interface
type restAPIManager struct {
	router      RouterEngine
	db          DatabaseManager
	authManager *AuthManager
	middleware  []RESTMiddleware
	prefix      string
	rateLimiter *rateLimiter
}

// NewRESTAPIManager creates a new REST API manager
func NewRESTAPIManager(router RouterEngine, db DatabaseManager) RESTAPIManager {
	return &restAPIManager{
		router:      router,
		db:          db,
		authManager: nil, // Will be set later if needed
		middleware:  make([]RESTMiddleware, 0),
		prefix:      "",
		rateLimiter: newRateLimiter(db),
	}
}

// NewRESTAPIManagerWithAuth creates a new REST API manager with AuthManager
func NewRESTAPIManagerWithAuth(router RouterEngine, db DatabaseManager, authManager *AuthManager) RESTAPIManager {
	return &restAPIManager{
		router:      router,
		db:          db,
		authManager: authManager,
		middleware:  make([]RESTMiddleware, 0),
		prefix:      "",
		rateLimiter: newRateLimiter(db),
	}
}

// RegisterRoute registers a REST API route with configuration
func (r *restAPIManager) RegisterRoute(method, path string, handler RESTHandler, config RESTRouteConfig) error {
	// Build full path with prefix
	fullPath := r.prefix + path

	// Wrap handler with REST middleware chain
	wrappedHandler := r.wrapHandler(handler, config)

	// Convert to framework handler
	frameworkHandler := func(ctx Context) error {
		return wrappedHandler(ctx)
	}

	// Register with router based on method
	switch strings.ToUpper(method) {
	case "GET":
		r.router.GET(fullPath, frameworkHandler)
	case "POST":
		r.router.POST(fullPath, frameworkHandler)
	case "PUT":
		r.router.PUT(fullPath, frameworkHandler)
	case "DELETE":
		r.router.DELETE(fullPath, frameworkHandler)
	case "PATCH":
		r.router.PATCH(fullPath, frameworkHandler)
	case "HEAD":
		r.router.HEAD(fullPath, frameworkHandler)
	case "OPTIONS":
		r.router.OPTIONS(fullPath, frameworkHandler)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	return nil
}

// wrapHandler wraps a REST handler with middleware and configuration
func (r *restAPIManager) wrapHandler(handler RESTHandler, config RESTRouteConfig) RESTHandler {
	// Start with the original handler
	wrapped := handler

	// Apply rate limiting middleware if configured
	if config.RateLimit != nil {
		wrapped = r.rateLimitMiddleware(config.RateLimit, wrapped)
	}

	if config.GlobalRateLimit != nil {
		wrapped = r.globalRateLimitMiddleware(config.GlobalRateLimit, wrapped)
	}

	// Apply authentication middleware if required
	if config.RequireAuth {
		wrapped = r.authMiddleware(config.RequiredScopes, wrapped)
	}

	// Apply request validation middleware
	if config.MaxRequestSize > 0 || config.Timeout > 0 {
		wrapped = r.validationMiddleware(config, wrapped)
	}

	// Apply CORS middleware if configured
	if config.CORS != nil {
		wrapped = r.corsMiddleware(config.CORS, wrapped)
	}

	// Apply custom middleware in order
	for i := len(r.middleware) - 1; i >= 0; i-- {
		mw := r.middleware[i]
		currentHandler := wrapped
		wrapped = func(ctx Context) error {
			return mw(ctx, currentHandler)
		}
	}

	return wrapped
}

// rateLimitMiddleware applies rate limiting per resource
func (r *restAPIManager) rateLimitMiddleware(config *RESTRateLimitConfig, next RESTHandler) RESTHandler {
	return func(ctx Context) error {
		// Build rate limit key
		key := r.buildRateLimitKey(ctx, config.Key)

		// Check rate limit
		allowed, err := r.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return r.SendErrorResponse(ctx, 500, "Rate limit check failed", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if !allowed {
			return r.SendErrorResponse(ctx, 429, "Rate limit exceeded", map[string]interface{}{
				"limit":  config.Limit,
				"window": config.Window.String(),
			})
		}

		// Increment rate limit counter
		if err := r.rateLimiter.Increment(key, config.Window); err != nil {
			// Log error but don't fail the request
			if ctx.Logger() != nil {
				ctx.Logger().Error("Failed to increment rate limit", "error", err)
			}
		}

		return next(ctx)
	}
}

// globalRateLimitMiddleware applies global rate limiting
func (r *restAPIManager) globalRateLimitMiddleware(config *RESTRateLimitConfig, next RESTHandler) RESTHandler {
	return func(ctx Context) error {
		key := "global:" + r.buildRateLimitKey(ctx, config.Key)

		allowed, err := r.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return r.SendErrorResponse(ctx, 500, "Global rate limit check failed", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if !allowed {
			return r.SendErrorResponse(ctx, 429, "Global rate limit exceeded", map[string]interface{}{
				"limit":  config.Limit,
				"window": config.Window.String(),
			})
		}

		if err := r.rateLimiter.Increment(key, config.Window); err != nil {
			if ctx.Logger() != nil {
				ctx.Logger().Error("Failed to increment global rate limit", "error", err)
			}
		}

		return next(ctx)
	}
}

// authMiddleware validates authentication and authorization
func (r *restAPIManager) authMiddleware(requiredScopes []string, next RESTHandler) RESTHandler {
	return func(ctx Context) error {
		if !ctx.IsAuthenticated() {
			r.SendErrorResponse(ctx, 401, "Authentication required", nil)
			return NewAuthenticationError("Authentication required")
		}

		// Check required scopes if specified
		if len(requiredScopes) > 0 {
			user := ctx.User()
			if user == nil {
				r.SendErrorResponse(ctx, 401, "User not found", nil)
				return NewAuthorizationError("User not found")
			}

			// Use AuthManager to check scopes if available
			if r.authManager != nil {
				err := r.authManager.AuthorizeAllScopes(user, requiredScopes)
				if err != nil {
					// Extract error details for response
					if frameworkErr, ok := err.(*FrameworkError); ok {
						r.SendErrorResponse(ctx, frameworkErr.StatusCode, frameworkErr.Message, frameworkErr.Details)
					} else {
						r.SendErrorResponse(ctx, 403, "Insufficient scopes", nil)
					}
					return err
				}
			} else {
				// Fallback: manual scope checking if AuthManager not available
				if err := r.checkScopesManually(ctx, user, requiredScopes); err != nil {
					return err
				}
			}
		}

		return next(ctx)
	}
}

// checkScopesManually performs manual scope checking when AuthManager is not available
func (r *restAPIManager) checkScopesManually(ctx Context, user *User, requiredScopes []string) error {
	// Check for wildcard scope
	for _, scope := range user.Scopes {
		if scope == "*" {
			return nil
		}
	}

	// Check if user has all required scopes
	missingScopes := []string{}
	for _, requiredScope := range requiredScopes {
		found := false
		for _, userScope := range user.Scopes {
			// Exact match or hierarchical match
			if userScope == requiredScope || strings.HasPrefix(requiredScope, userScope+":") {
				found = true
				break
			}
		}
		if !found {
			missingScopes = append(missingScopes, requiredScope)
		}
	}

	if len(missingScopes) > 0 {
		r.SendErrorResponse(ctx, 403, "Insufficient scopes", map[string]interface{}{
			"required_scopes": requiredScopes,
			"missing_scopes":  missingScopes,
			"user_scopes":     user.Scopes,
		})
		return &FrameworkError{
			Code:       ErrCodeInsufficientScopes,
			Message:    "Insufficient scopes",
			StatusCode: 403,
			Details: map[string]interface{}{
				"required_scopes": requiredScopes,
				"missing_scopes":  missingScopes,
				"user_scopes":     user.Scopes,
			},
		}
	}

	return nil
}

// validationMiddleware validates request size and timeout
func (r *restAPIManager) validationMiddleware(config RESTRouteConfig, next RESTHandler) RESTHandler {
	return func(ctx Context) error {
		// Validate request size
		if config.MaxRequestSize > 0 {
			body := ctx.Body()
			if int64(len(body)) > config.MaxRequestSize {
				return r.SendErrorResponse(ctx, 413, "Request entity too large", map[string]interface{}{
					"max_size":     config.MaxRequestSize,
					"request_size": len(body),
				})
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
func (r *restAPIManager) corsMiddleware(config *CORSConfig, next RESTHandler) RESTHandler {
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
func (r *restAPIManager) buildRateLimitKey(ctx Context, keyType string) string {
	switch keyType {
	case "user_id":
		if user := ctx.User(); user != nil {
			return fmt.Sprintf("user:%s", user.ID)
		}
		return "user:anonymous"
	case "tenant_id":
		if tenant := ctx.Tenant(); tenant != nil {
			return fmt.Sprintf("tenant:%s", tenant.ID)
		}
		return "tenant:default"
	case "ip_address":
		return fmt.Sprintf("ip:%s", ctx.Request().RemoteAddr)
	default:
		// Default to IP address
		return fmt.Sprintf("ip:%s", ctx.Request().RemoteAddr)
	}
}

// CheckRateLimit checks rate limit for a specific resource
func (r *restAPIManager) CheckRateLimit(ctx Context, resource string) error {
	// Default rate limit: 100 requests per minute
	key := fmt.Sprintf("resource:%s:%s", resource, ctx.Request().RemoteAddr)
	allowed, err := r.rateLimiter.Check(key, 100, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return ErrRateLimitExceeded
	}

	return r.rateLimiter.Increment(key, time.Minute)
}

// CheckGlobalRateLimit checks global rate limit
func (r *restAPIManager) CheckGlobalRateLimit(ctx Context) error {
	// Default global rate limit: 1000 requests per minute
	key := fmt.Sprintf("global:%s", ctx.Request().RemoteAddr)
	allowed, err := r.rateLimiter.Check(key, 1000, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return ErrRateLimitExceeded
	}

	return r.rateLimiter.Increment(key, time.Minute)
}

// ParseJSONRequest parses JSON from request body
func (r *restAPIManager) ParseJSONRequest(ctx Context, target interface{}) error {
	body := ctx.Body()
	if len(body) == 0 {
		return fmt.Errorf("empty request body")
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// SendJSONResponse sends a JSON response
func (r *restAPIManager) SendJSONResponse(ctx Context, statusCode int, data interface{}) error {
	response := RESTResponse{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
		Meta: &RESTMeta{
			Timestamp: time.Now(),
			RequestID: ctx.Request().ID,
		},
	}

	return ctx.JSON(statusCode, response)
}

// SendErrorResponse sends an error response
func (r *restAPIManager) SendErrorResponse(ctx Context, statusCode int, message string, details map[string]interface{}) error {
	response := RESTResponse{
		Success: false,
		Error: &RESTError{
			Code:    fmt.Sprintf("ERROR_%d", statusCode),
			Message: message,
			Details: details,
			Status:  statusCode,
		},
		Meta: &RESTMeta{
			Timestamp: time.Now(),
			RequestID: ctx.Request().ID,
		},
	}

	return ctx.JSON(statusCode, response)
}

// Use adds middleware to the REST API manager
func (r *restAPIManager) Use(middleware RESTMiddleware) RESTAPIManager {
	r.middleware = append(r.middleware, middleware)
	return r
}

// Group creates a new REST API manager with a prefix
func (r *restAPIManager) Group(prefix string, middleware ...RESTMiddleware) RESTAPIManager {
	newManager := &restAPIManager{
		router:      r.router,
		db:          r.db,
		authManager: r.authManager,
		middleware:  append(r.middleware, middleware...),
		prefix:      r.prefix + prefix,
		rateLimiter: r.rateLimiter,
	}
	return newManager
}

// rateLimiter handles rate limiting with database storage
type rateLimiter struct {
	db DatabaseManager
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(db DatabaseManager) *rateLimiter {
	return &rateLimiter{db: db}
}

// Check checks if a request is within rate limits
func (rl *rateLimiter) Check(key string, limit int, window time.Duration) (bool, error) {
	if rl.db == nil {
		// If no database, allow all requests
		return true, nil
	}

	exceeded, err := rl.db.CheckRateLimit(key, limit, window)
	if err != nil {
		return false, err
	}
	// Return true if NOT exceeded (allowed), false if exceeded (not allowed)
	return !exceeded, nil
}

// Increment increments the rate limit counter
func (rl *rateLimiter) Increment(key string, window time.Duration) error {
	if rl.db == nil {
		return nil
	}

	return rl.db.IncrementRateLimit(key, window)
}
