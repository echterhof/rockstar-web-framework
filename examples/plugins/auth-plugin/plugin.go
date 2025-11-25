package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/rockstar/pkg"
)

// AuthPlugin demonstrates authentication hooks and middleware
type AuthPlugin struct {
	ctx           pkg.PluginContext
	validTokens   map[string]string // token -> username
	tokenDuration time.Duration
}

// Name returns the plugin name
func (p *AuthPlugin) Name() string {
	return "auth-plugin"
}

// Version returns the plugin version
func (p *AuthPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *AuthPlugin) Description() string {
	return "Authentication plugin demonstrating hooks and middleware"
}

// Author returns the plugin author
func (p *AuthPlugin) Author() string {
	return "Rockstar Framework Team"
}

// Dependencies returns the plugin dependencies
func (p *AuthPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

// Initialize initializes the plugin with the provided context
func (p *AuthPlugin) Initialize(ctx pkg.PluginContext) error {
	p.ctx = ctx
	p.validTokens = make(map[string]string)
	
	// Get configuration
	config := ctx.PluginConfig()
	if duration, ok := config["token_duration"].(string); ok {
		if d, err := time.ParseDuration(duration); err == nil {
			p.tokenDuration = d
		}
	}
	
	// Default to 1 hour
	if p.tokenDuration == 0 {
		p.tokenDuration = time.Hour
	}
	
	// Register pre-request hook for authentication
	err := ctx.RegisterHook(pkg.HookTypePreRequest, 100, p.authenticateRequest)
	if err != nil {
		return fmt.Errorf("failed to register authentication hook: %w", err)
	}
	
	// Register authentication middleware
	err = ctx.RegisterMiddleware("auth-check", p.authMiddleware, 100, []string{})
	if err != nil {
		return fmt.Errorf("failed to register auth middleware: %w", err)
	}
	
	// Subscribe to login events
	err = ctx.SubscribeEvent("auth.login", p.handleLoginEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to login events: %w", err)
	}
	
	if logger := ctx.Logger(); logger != nil {
		logger.Info("Auth plugin initialized")
	}
	
	return nil
}

// Start starts the plugin
func (p *AuthPlugin) Start() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Auth plugin started")
	}
	
	// Publish startup event
	p.ctx.PublishEvent("auth.started", map[string]interface{}{
		"timestamp": time.Now(),
	})
	
	return nil
}

// Stop stops the plugin
func (p *AuthPlugin) Stop() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Auth plugin stopped")
	}
	
	// Publish shutdown event
	p.ctx.PublishEvent("auth.stopped", map[string]interface{}{
		"timestamp": time.Now(),
	})
	
	return nil
}

// Cleanup cleans up plugin resources
func (p *AuthPlugin) Cleanup() error {
	p.validTokens = nil
	
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Auth plugin cleaned up")
	}
	
	return nil
}

// ConfigSchema returns the configuration schema
func (p *AuthPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"token_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "1h",
			"description": "Duration for which authentication tokens are valid",
		},
		"require_auth": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Whether to require authentication for all requests",
		},
		"excluded_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []interface{}{"/health", "/public"},
			"description": "Paths that don't require authentication",
		},
	}
}

// OnConfigChange handles configuration changes
func (p *AuthPlugin) OnConfigChange(config map[string]interface{}) error {
	if duration, ok := config["token_duration"].(string); ok {
		if d, err := time.ParseDuration(duration); err == nil {
			p.tokenDuration = d
		}
	}
	
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Auth plugin configuration updated")
	}
	
	return nil
}

// authenticateRequest is a hook that runs before each request
func (p *AuthPlugin) authenticateRequest(ctx pkg.HookContext) error {
	reqCtx := ctx.Context()
	if reqCtx == nil {
		return nil
	}
	
	// Get the authorization header
	authHeader := reqCtx.GetHeader("Authorization")
	if authHeader == "" {
		// Check if this path is excluded
		config := p.ctx.PluginConfig()
		if excluded, ok := config["excluded_paths"].([]interface{}); ok {
			path := reqCtx.Path()
			for _, ex := range excluded {
				if exPath, ok := ex.(string); ok && strings.HasPrefix(path, exPath) {
					return nil
				}
			}
		}
		
		// Store authentication failure in context
		ctx.Set("auth_failed", true)
		ctx.Set("auth_reason", "missing authorization header")
		return nil
	}
	
	// Extract token from "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.Set("auth_failed", true)
		ctx.Set("auth_reason", "invalid authorization format")
		return nil
	}
	
	token := parts[1]
	
	// Validate token
	if username, ok := p.validTokens[token]; ok {
		ctx.Set("authenticated", true)
		ctx.Set("username", username)
		
		if logger := p.ctx.Logger(); logger != nil {
			logger.Info(fmt.Sprintf("User %s authenticated", username))
		}
	} else {
		ctx.Set("auth_failed", true)
		ctx.Set("auth_reason", "invalid token")
	}
	
	return nil
}

// authMiddleware is middleware that enforces authentication
func (p *AuthPlugin) authMiddleware(ctx pkg.Context) error {
	// This would typically check authentication and return 401 if not authenticated
	// For this example, we'll just log the request
	
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info(fmt.Sprintf("Auth middleware processing request to %s", ctx.Path()))
	}
	
	return nil
}

// handleLoginEvent handles login events from other plugins
func (p *AuthPlugin) handleLoginEvent(event pkg.Event) error {
	if data, ok := event.Data.(map[string]interface{}); ok {
		if username, ok := data["username"].(string); ok {
			if token, ok := data["token"].(string); ok {
				// Store the token
				p.validTokens[token] = username
				
				if logger := p.ctx.Logger(); logger != nil {
					logger.Info(fmt.Sprintf("Login event received for user %s", username))
				}
				
				// Publish authentication success event
				p.ctx.PublishEvent("auth.authenticated", map[string]interface{}{
					"username":  username,
					"timestamp": time.Now(),
				})
			}
		}
	}
	
	return nil
}

// NewPlugin creates a new instance of the plugin
func NewPlugin() pkg.Plugin {
	return &AuthPlugin{}
}
