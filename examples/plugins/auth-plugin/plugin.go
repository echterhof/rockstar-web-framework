package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Authentication Plugin Example
// ============================================================================
//
// This plugin demonstrates:
// - Authentication hooks
// - Authentication middleware
// - Hook registration
// - Plugin configuration handling
// - Database and cache integration
//
// Requirements: 1.1, 1.2, 1.3, 2.1, 2.5, 7.2, 7.5
// ============================================================================

// AuthPlugin implements authentication functionality
type AuthPlugin struct {
	ctx pkg.PluginContext
	
	// Configuration
	jwtSecret            string
	tokenDuration        time.Duration
	requireAuth          bool
	excludedPaths        []string
	sessionCookieName    string
	sessionDuration      time.Duration
	enableCSRF           bool
	enableRateLimiting   bool
	maxLoginAttempts     int
	lockoutDuration      time.Duration
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

func (p *AuthPlugin) Name() string {
	return "auth-plugin"
}

func (p *AuthPlugin) Version() string {
	return "1.0.0"
}

func (p *AuthPlugin) Description() string {
	return "Authentication and authorization plugin with JWT and session support"
}

func (p *AuthPlugin) Author() string {
	return "Rockstar Framework Team"
}

func (p *AuthPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

func (p *AuthPlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing authentication plugin...\n", p.Name())
	
	p.ctx = ctx
	
	// Parse configuration
	if err := p.parseConfig(ctx.PluginConfig()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Register authentication hooks
	if err := p.registerHooks(); err != nil {
		return fmt.Errorf("failed to register hooks: %w", err)
	}
	
	// Register authentication middleware
	if err := p.registerMiddleware(); err != nil {
		return fmt.Errorf("failed to register middleware: %w", err)
	}
	
	// Initialize database tables if needed
	if err := p.initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

func (p *AuthPlugin) Start() error {
	fmt.Printf("[%s] Starting authentication plugin...\n", p.Name())
	
	// Start background tasks (e.g., token cleanup)
	go p.cleanupExpiredTokens()
	
	fmt.Printf("[%s] Plugin started\n", p.Name())
	return nil
}

func (p *AuthPlugin) Stop() error {
	fmt.Printf("[%s] Stopping authentication plugin...\n", p.Name())
	
	// Stop background tasks
	// In a real implementation, you would use a context or channel to signal shutdown
	
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

func (p *AuthPlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up authentication plugin...\n", p.Name())
	
	p.ctx = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

func (p *AuthPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"jwt_secret": map[string]interface{}{
			"type":        "string",
			"required":    true,
			"description": "Secret key for JWT token signing",
		},
		"token_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "2h",
			"description": "JWT token validity duration",
		},
		"require_auth": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Whether authentication is required by default",
		},
		"excluded_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"/health", "/metrics"},
			"description": "Paths that don't require authentication",
		},
		"session_cookie_name": map[string]interface{}{
			"type":        "string",
			"default":     "auth_session",
			"description": "Name of the session cookie",
		},
		"session_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "24h",
			"description": "Session validity duration",
		},
		"enable_csrf": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable CSRF protection",
		},
		"enable_rate_limiting": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable rate limiting for login attempts",
		},
		"max_login_attempts": map[string]interface{}{
			"type":        "int",
			"default":     5,
			"description": "Maximum login attempts before lockout",
		},
		"lockout_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "15m",
			"description": "Account lockout duration after max attempts",
		},
	}
}

func (p *AuthPlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	
	return p.parseConfig(config)
}

// ============================================================================
// Configuration Parsing
// ============================================================================

func (p *AuthPlugin) parseConfig(config map[string]interface{}) error {
	// JWT Secret (required)
	if secret, ok := config["jwt_secret"].(string); ok {
		p.jwtSecret = secret
	} else {
		return fmt.Errorf("jwt_secret is required")
	}
	
	// Token Duration
	if duration, ok := config["token_duration"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid token_duration: %w", err)
		}
		p.tokenDuration = d
	} else {
		p.tokenDuration = 2 * time.Hour
	}
	
	// Require Auth
	if requireAuth, ok := config["require_auth"].(bool); ok {
		p.requireAuth = requireAuth
	} else {
		p.requireAuth = true
	}
	
	// Excluded Paths
	if paths, ok := config["excluded_paths"].([]interface{}); ok {
		p.excludedPaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if pathStr, ok := path.(string); ok {
				p.excludedPaths = append(p.excludedPaths, pathStr)
			}
		}
	} else {
		p.excludedPaths = []string{"/health", "/metrics"}
	}
	
	// Session Cookie Name
	if cookieName, ok := config["session_cookie_name"].(string); ok {
		p.sessionCookieName = cookieName
	} else {
		p.sessionCookieName = "auth_session"
	}
	
	// Session Duration
	if duration, ok := config["session_duration"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid session_duration: %w", err)
		}
		p.sessionDuration = d
	} else {
		p.sessionDuration = 24 * time.Hour
	}
	
	// Enable CSRF
	if enableCSRF, ok := config["enable_csrf"].(bool); ok {
		p.enableCSRF = enableCSRF
	} else {
		p.enableCSRF = true
	}
	
	// Enable Rate Limiting
	if enableRL, ok := config["enable_rate_limiting"].(bool); ok {
		p.enableRateLimiting = enableRL
	} else {
		p.enableRateLimiting = true
	}
	
	// Max Login Attempts
	if maxAttempts, ok := config["max_login_attempts"].(int); ok {
		p.maxLoginAttempts = maxAttempts
	} else if maxAttemptsFloat, ok := config["max_login_attempts"].(float64); ok {
		p.maxLoginAttempts = int(maxAttemptsFloat)
	} else {
		p.maxLoginAttempts = 5
	}
	
	// Lockout Duration
	if duration, ok := config["lockout_duration"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid lockout_duration: %w", err)
		}
		p.lockoutDuration = d
	} else {
		p.lockoutDuration = 15 * time.Minute
	}
	
	return nil
}

// ============================================================================
// Hook Registration
// ============================================================================

func (p *AuthPlugin) registerHooks() error {
	// Register pre-request hook for authentication
	err := p.ctx.RegisterHook(pkg.HookTypePreRequest, 200, func(hookCtx pkg.HookContext) error {
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		req := reqCtx.Request()
		if req == nil || req.URL == nil {
			return nil
		}
		path := req.URL.Path
		
		// Check if path is excluded
		if p.isPathExcluded(path) {
			return nil
		}
		
		// Perform authentication check
		if p.requireAuth {
			if !p.isAuthenticated(reqCtx) {
				// Set error in context for middleware to handle
				hookCtx.Set("auth_error", "authentication required")
			}
		}
		
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register pre-request hook: %w", err)
	}
	
	// Register post-request hook for logging
	err = p.ctx.RegisterHook(pkg.HookTypePostRequest, 200, func(hookCtx pkg.HookContext) error {
		reqCtx := hookCtx.Context()
		if reqCtx != nil {
			// Log authentication events
			// In a real implementation, check for authenticated user
			fmt.Printf("[%s] Post-request hook executed\n", p.Name())
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register post-request hook: %w", err)
	}
	
	return nil
}

// ============================================================================
// Middleware Registration
// ============================================================================

func (p *AuthPlugin) registerMiddleware() error {
	// Register authentication middleware
	authMiddleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		req := ctx.Request()
		if req == nil || req.URL == nil {
			return next(ctx)
		}
		path := req.URL.Path
		
		// Check if path is excluded
		if p.isPathExcluded(path) {
			return next(ctx)
		}
		
		// Check authentication
		if p.requireAuth && !p.isAuthenticated(ctx) {
			return ctx.JSON(401, map[string]interface{}{
				"error": "authentication required",
			})
		}
		
		return next(ctx)
	}
	
	// Register the middleware with high priority
	err := p.ctx.RegisterMiddleware("auth", authMiddleware, 200, []string{})
	if err != nil {
		return fmt.Errorf("failed to register auth middleware: %w", err)
	}
	
	fmt.Printf("[%s] Registered authentication middleware\n", p.Name())
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (p *AuthPlugin) isPathExcluded(path string) bool {
	for _, excluded := range p.excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}

func (p *AuthPlugin) isAuthenticated(ctx pkg.Context) bool {
	// Check for JWT token in Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// In a real implementation, validate the JWT token
		if p.validateToken(token) {
			return true
		}
	}
	
	// Check for session cookie
	cookie, err := ctx.GetCookie(p.sessionCookieName)
	if err == nil && cookie != nil && cookie.Value != "" {
		// In a real implementation, validate the session
		if p.validateSession(cookie.Value) {
			return true
		}
	}
	
	return false
}

func (p *AuthPlugin) validateToken(token string) bool {
	// Simplified token validation
	// In a real implementation, use a JWT library to validate the token
	return token != ""
}

func (p *AuthPlugin) validateSession(sessionID string) bool {
	// Simplified session validation
	// In a real implementation, check the session in the database or cache
	return sessionID != ""
}

func (p *AuthPlugin) initializeDatabase() error {
	// Initialize database tables for authentication
	// In a real implementation, create tables for users, sessions, tokens, etc.
	fmt.Printf("[%s] Database tables initialized\n", p.Name())
	return nil
}

func (p *AuthPlugin) cleanupExpiredTokens() {
	// Background task to clean up expired tokens
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// In a real implementation, delete expired tokens from the database
		fmt.Printf("[%s] Cleaning up expired tokens...\n", p.Name())
	}
}

// ============================================================================
// Plugin Entry Point
// ============================================================================

func NewPlugin() pkg.Plugin {
	return &AuthPlugin{}
}

func main() {
	// Plugin entry point
}
