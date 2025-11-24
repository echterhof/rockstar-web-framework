package pkg

import (
	"fmt"
	"strings"
	"time"
)

// graphqlManager implements the GraphQLManager interface
type graphqlManager struct {
	router      RouterEngine
	db          DatabaseManager
	authManager *AuthManager
	middleware  []GraphQLMiddleware
	prefix      string
	rateLimiter *rateLimiter
}

// NewGraphQLManager creates a new GraphQL manager
func NewGraphQLManager(router RouterEngine, db DatabaseManager, authManager *AuthManager) GraphQLManager {
	return &graphqlManager{
		router:      router,
		db:          db,
		authManager: authManager,
		middleware:  make([]GraphQLMiddleware, 0),
		prefix:      "",
		rateLimiter: newRateLimiter(db),
	}
}

// RegisterSchema registers a GraphQL schema with configuration
// Requirements: 2.1
func (g *graphqlManager) RegisterSchema(path string, schema GraphQLSchema, config GraphQLConfig) error {
	// Build full path with prefix
	fullPath := g.prefix + path

	// Wrap handler with GraphQL middleware chain
	wrappedHandler := g.wrapHandler(schema, config)

	// Convert to framework handler
	frameworkHandler := func(ctx Context) error {
		return wrappedHandler(ctx)
	}

	// Register POST endpoint for GraphQL queries
	g.router.POST(fullPath, frameworkHandler)

	// Register GET endpoint for introspection and playground
	if config.EnableIntrospection || config.EnablePlayground {
		g.router.GET(fullPath, func(ctx Context) error {
			// If playground is enabled, serve playground HTML
			if config.EnablePlayground {
				return g.servePlayground(ctx, fullPath)
			}

			// Otherwise, handle GET queries (for introspection)
			return wrappedHandler(ctx)
		})
	}

	return nil
}

// wrapHandler wraps a GraphQL schema with middleware and configuration
func (g *graphqlManager) wrapHandler(schema GraphQLSchema, config GraphQLConfig) GraphQLHandler {
	// Create the base handler that executes GraphQL queries
	handler := func(ctx Context) error {
		// Parse GraphQL request
		req, err := g.parseRequest(ctx)
		if err != nil {
			return g.sendErrorResponse(ctx, []GraphQLError{
				NewGraphQLError(fmt.Sprintf("Invalid request: %s", err.Error())),
			})
		}

		// Execute query using the basic GraphQLSchema interface
		result, err := schema.Execute(req.Query, req.Variables)
		if err != nil {
			return g.sendErrorResponse(ctx, []GraphQLError{
				NewGraphQLError(fmt.Sprintf("Execution error: %s", err.Error())),
			})
		}

		// Build GraphQL response
		response := &GraphQLResponse{
			Data: result,
			Extensions: &GraphQLExtensions{
				Timestamp: time.Now(),
				RequestID: ctx.Request().ID,
			},
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
	if config.MaxRequestSize > 0 || config.Timeout > 0 || config.MaxQueryDepth > 0 || config.MaxComplexity > 0 {
		handler = g.validationMiddleware(config, handler)
	}

	// Apply CORS middleware if configured
	if config.CORS != nil {
		handler = g.corsMiddleware(config.CORS, handler)
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

// parseRequest parses a GraphQL request from the context
func (g *graphqlManager) parseRequest(ctx Context) (*GraphQLRequest, error) {
	method := ctx.Request().Method

	if method == "POST" {
		// Parse JSON body
		return ParseGraphQLRequest(ctx.Body())
	} else if method == "GET" {
		// Parse query parameters
		query := ctx.Query()
		req := &GraphQLRequest{
			Query:         query["query"],
			OperationName: query["operationName"],
		}

		// Parse variables if present
		if varsStr := query["variables"]; varsStr != "" {
			// Variables should be JSON-encoded
			// For simplicity, we'll skip parsing here
			// In a real implementation, you'd parse the JSON
		}

		if req.Query == "" {
			return nil, fmt.Errorf("query parameter is required")
		}

		return req, nil
	}

	return nil, fmt.Errorf("unsupported HTTP method: %s", method)
}

// rateLimitMiddleware applies rate limiting per resource
// Requirements: 2.6
func (g *graphqlManager) rateLimitMiddleware(config *GraphQLRateLimitConfig, next GraphQLHandler) GraphQLHandler {
	return func(ctx Context) error {
		// Build rate limit key
		key := g.buildRateLimitKey(ctx, config.Key)

		// Check rate limit
		allowed, err := g.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return g.sendErrorResponse(ctx, []GraphQLError{
				NewGraphQLError("Rate limit check failed").WithExtensions(map[string]interface{}{
					"error": err.Error(),
				}),
			})
		}

		if !allowed {
			return g.sendErrorResponse(ctx, []GraphQLError{
				ErrGraphQLRateLimit.WithExtensions(map[string]interface{}{
					"limit":  config.Limit,
					"window": config.Window.String(),
				}),
			})
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
func (g *graphqlManager) globalRateLimitMiddleware(config *GraphQLRateLimitConfig, next GraphQLHandler) GraphQLHandler {
	return func(ctx Context) error {
		key := "global:" + g.buildRateLimitKey(ctx, config.Key)

		allowed, err := g.rateLimiter.Check(key, config.Limit, config.Window)
		if err != nil {
			return g.sendErrorResponse(ctx, []GraphQLError{
				NewGraphQLError("Global rate limit check failed").WithExtensions(map[string]interface{}{
					"error": err.Error(),
				}),
			})
		}

		if !allowed {
			return g.sendErrorResponse(ctx, []GraphQLError{
				ErrGraphQLRateLimit.WithExtensions(map[string]interface{}{
					"limit":  config.Limit,
					"window": config.Window.String(),
					"scope":  "global",
				}),
			})
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
func (g *graphqlManager) authMiddleware(config GraphQLConfig, next GraphQLHandler) GraphQLHandler {
	return func(ctx Context) error {
		if !ctx.IsAuthenticated() {
			return g.sendErrorResponse(ctx, []GraphQLError{
				ErrGraphQLAuthentication,
			})
		}

		user := ctx.User()
		if user == nil {
			return g.sendErrorResponse(ctx, []GraphQLError{
				ErrGraphQLAuthentication.WithExtensions(map[string]interface{}{
					"reason": "user not found",
				}),
			})
		}

		// Check required roles if specified
		if len(config.RequiredRoles) > 0 && g.authManager != nil {
			if err := g.authManager.AuthorizeRoles(user, config.RequiredRoles); err != nil {
				return g.sendErrorResponse(ctx, []GraphQLError{
					ErrGraphQLAuthorization.WithExtensions(map[string]interface{}{
						"required_roles": config.RequiredRoles,
						"user_roles":     user.Roles,
					}),
				})
			}
		}

		// Check required scopes if specified
		if len(config.RequiredScopes) > 0 {
			// Scopes are typically stored in user.Roles for access tokens
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
				return g.sendErrorResponse(ctx, []GraphQLError{
					ErrGraphQLAuthorization.WithExtensions(map[string]interface{}{
						"required_scopes": config.RequiredScopes,
						"user_scopes":     user.Roles,
					}),
				})
			}
		}

		return next(ctx)
	}
}

// validationMiddleware validates request size, timeout, and query complexity
func (g *graphqlManager) validationMiddleware(config GraphQLConfig, next GraphQLHandler) GraphQLHandler {
	return func(ctx Context) error {
		// Validate request size
		if config.MaxRequestSize > 0 {
			body := ctx.Body()
			if int64(len(body)) > config.MaxRequestSize {
				return g.sendErrorResponse(ctx, []GraphQLError{
					NewGraphQLError("Request entity too large").WithExtensions(map[string]interface{}{
						"max_size":     config.MaxRequestSize,
						"request_size": len(body),
					}),
				})
			}
		}

		// Apply timeout if configured
		if config.Timeout > 0 {
			timeoutCtx := ctx.WithTimeout(config.Timeout)
			return next(timeoutCtx)
		}

		// Note: Query depth and complexity validation would require parsing the query
		// This is typically done by the GraphQL schema implementation
		// We're leaving this as a placeholder for now

		return next(ctx)
	}
}

// corsMiddleware applies CORS headers
func (g *graphqlManager) corsMiddleware(config *CORSConfig, next GraphQLHandler) GraphQLHandler {
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
func (g *graphqlManager) buildRateLimitKey(ctx Context, keyType string) string {
	switch keyType {
	case "user_id":
		if user := ctx.User(); user != nil {
			return fmt.Sprintf("graphql:user:%s", user.ID)
		}
		return "graphql:user:anonymous"
	case "tenant_id":
		if tenant := ctx.Tenant(); tenant != nil {
			return fmt.Sprintf("graphql:tenant:%s", tenant.ID)
		}
		return "graphql:tenant:default"
	case "ip_address":
		return fmt.Sprintf("graphql:ip:%s", ctx.Request().RemoteAddr)
	default:
		// Default to IP address
		return fmt.Sprintf("graphql:ip:%s", ctx.Request().RemoteAddr)
	}
}

// ExecuteQuery executes a GraphQL query
func (g *graphqlManager) ExecuteQuery(ctx Context, query string, variables map[string]interface{}, operationName string) (*GraphQLResponse, error) {
	// This is a helper method that can be used programmatically
	// The actual execution is handled by the registered schema
	return nil, fmt.Errorf("ExecuteQuery should be called on a specific schema")
}

// CheckRateLimit checks rate limit for a specific resource
// Requirements: 2.6
func (g *graphqlManager) CheckRateLimit(ctx Context, resource string) error {
	// Default rate limit: 100 requests per minute
	key := fmt.Sprintf("graphql:resource:%s:%s", resource, ctx.Request().RemoteAddr)
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
func (g *graphqlManager) CheckGlobalRateLimit(ctx Context) error {
	// Default global rate limit: 1000 requests per minute
	key := fmt.Sprintf("graphql:global:%s", ctx.Request().RemoteAddr)
	allowed, err := g.rateLimiter.Check(key, 1000, time.Minute)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("global rate limit exceeded")
	}

	return g.rateLimiter.Increment(key, time.Minute)
}

// sendErrorResponse sends a GraphQL error response
func (g *graphqlManager) sendErrorResponse(ctx Context, errors []GraphQLError) error {
	response := &GraphQLResponse{
		Errors: errors,
		Extensions: &GraphQLExtensions{
			Timestamp: time.Now(),
			RequestID: ctx.Request().ID,
		},
	}

	return ctx.JSON(400, response)
}

// servePlayground serves the GraphQL playground HTML
func (g *graphqlManager) servePlayground(ctx Context, endpoint string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root"></div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '%s'
      })
    })
  </script>
</body>
</html>`, endpoint)

	ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
	return ctx.String(200, html)
}

// Use adds middleware to the GraphQL manager
func (g *graphqlManager) Use(middleware GraphQLMiddleware) GraphQLManager {
	g.middleware = append(g.middleware, middleware)
	return g
}

// Group creates a new GraphQL manager with a prefix
func (g *graphqlManager) Group(prefix string, middleware ...GraphQLMiddleware) GraphQLManager {
	newManager := &graphqlManager{
		router:      g.router,
		db:          g.db,
		authManager: g.authManager,
		middleware:  append(g.middleware, middleware...),
		prefix:      g.prefix + prefix,
		rateLimiter: g.rateLimiter,
	}
	return newManager
}
