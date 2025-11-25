package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/rockstar/pkg"
)

// CachePlugin demonstrates caching middleware
type CachePlugin struct {
	ctx           pkg.PluginContext
	cacheDuration time.Duration
	cacheHits     int64
	cacheMisses   int64
	cacheEnabled  bool
}

// Name returns the plugin name
func (p *CachePlugin) Name() string {
	return "cache-plugin"
}

// Version returns the plugin version
func (p *CachePlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *CachePlugin) Description() string {
	return "Cache plugin demonstrating caching middleware"
}

// Author returns the plugin author
func (p *CachePlugin) Author() string {
	return "Rockstar Framework Team"
}

// Dependencies returns the plugin dependencies
func (p *CachePlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

// Initialize initializes the plugin with the provided context
func (p *CachePlugin) Initialize(ctx pkg.PluginContext) error {
	p.ctx = ctx
	
	// Get configuration
	config := ctx.PluginConfig()
	if duration, ok := config["cache_duration"].(string); ok {
		if d, err := time.ParseDuration(duration); err == nil {
			p.cacheDuration = d
		}
	}
	
	// Default to 5 minutes
	if p.cacheDuration == 0 {
		p.cacheDuration = 5 * time.Minute
	}
	
	if enabled, ok := config["enabled"].(bool); ok {
		p.cacheEnabled = enabled
	} else {
		p.cacheEnabled = true
	}
	
	// Register caching middleware
	if p.cacheEnabled {
		err := ctx.RegisterMiddleware("cache", p.cacheMiddleware, 50, []string{})
		if err != nil {
			return fmt.Errorf("failed to register cache middleware: %w", err)
		}
	}
	
	// Register pre-response hook to cache responses
	err := ctx.RegisterHook(pkg.HookTypePreResponse, 50, p.cacheResponse)
	if err != nil {
		return fmt.Errorf("failed to register cache response hook: %w", err)
	}
	
	// Subscribe to cache invalidation events
	err = ctx.SubscribeEvent("cache.invalidate", p.handleInvalidation)
	if err != nil {
		return fmt.Errorf("failed to subscribe to cache invalidation events: %w", err)
	}
	
	// Export cache service
	err = ctx.ExportService("CacheService", &CacheService{plugin: p})
	if err != nil {
		return fmt.Errorf("failed to export cache service: %w", err)
	}
	
	if logger := ctx.Logger(); logger != nil {
		logger.Info("Cache plugin initialized")
	}
	
	return nil
}

// Start starts the plugin
func (p *CachePlugin) Start() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Cache plugin started")
	}
	return nil
}

// Stop stops the plugin
func (p *CachePlugin) Stop() error {
	if logger := p.ctx.Logger(); logger != nil {
		hitRate := float64(0)
		total := p.cacheHits + p.cacheMisses
		if total > 0 {
			hitRate = float64(p.cacheHits) / float64(total) * 100
		}
		logger.Info(fmt.Sprintf("Cache plugin stopped. Hits: %d, Misses: %d, Hit Rate: %.2f%%", 
			p.cacheHits, p.cacheMisses, hitRate))
	}
	return nil
}

// Cleanup cleans up plugin resources
func (p *CachePlugin) Cleanup() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Cache plugin cleaned up")
	}
	return nil
}

// ConfigSchema returns the configuration schema
func (p *CachePlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"cache_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "5m",
			"description": "Duration for which responses are cached",
		},
		"enabled": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Whether caching is enabled",
		},
		"cache_methods": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []interface{}{"GET"},
			"description": "HTTP methods to cache",
		},
		"excluded_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []interface{}{"/admin", "/api/auth"},
			"description": "Paths to exclude from caching",
		},
	}
}

// OnConfigChange handles configuration changes
func (p *CachePlugin) OnConfigChange(config map[string]interface{}) error {
	if duration, ok := config["cache_duration"].(string); ok {
		if d, err := time.ParseDuration(duration); err == nil {
			p.cacheDuration = d
		}
	}
	
	if enabled, ok := config["enabled"].(bool); ok {
		p.cacheEnabled = enabled
	}
	
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Cache plugin configuration updated")
	}
	
	return nil
}

// cacheMiddleware checks cache before processing request
func (p *CachePlugin) cacheMiddleware(ctx pkg.Context) error {
	if !p.cacheEnabled {
		return nil
	}
	
	// Only cache GET requests
	if ctx.Method() != "GET" {
		return nil
	}
	
	// Check if path is excluded
	config := p.ctx.PluginConfig()
	if excluded, ok := config["excluded_paths"].([]interface{}); ok {
		path := ctx.Path()
		for _, ex := range excluded {
			if exPath, ok := ex.(string); ok && strings.HasPrefix(path, exPath) {
				return nil
			}
		}
	}
	
	// Generate cache key
	cacheKey := p.generateCacheKey(ctx)
	
	// Try to get from cache
	cache := p.ctx.Cache()
	if cache != nil {
		if cachedData, err := cache.Get(cacheKey); err == nil && cachedData != nil {
			p.cacheHits++
			
			if logger := p.ctx.Logger(); logger != nil {
				logger.Info(fmt.Sprintf("Cache hit for %s", ctx.Path()))
			}
			
			// Record metrics
			if metrics := p.ctx.Metrics(); metrics != nil {
				metrics.IncrementCounter("cache_plugin.hits", 1)
			}
			
			// In a real implementation, we would return the cached response here
			// For this example, we just log it
		} else {
			p.cacheMisses++
			
			if logger := p.ctx.Logger(); logger != nil {
				logger.Info(fmt.Sprintf("Cache miss for %s", ctx.Path()))
			}
			
			// Record metrics
			if metrics := p.ctx.Metrics(); metrics != nil {
				metrics.IncrementCounter("cache_plugin.misses", 1)
			}
		}
	}
	
	return nil
}

// cacheResponse caches the response before it's sent
func (p *CachePlugin) cacheResponse(ctx pkg.HookContext) error {
	if !p.cacheEnabled {
		return nil
	}
	
	reqCtx := ctx.Context()
	if reqCtx == nil {
		return nil
	}
	
	// Only cache GET requests with 200 status
	if reqCtx.Method() != "GET" || reqCtx.Status() != 200 {
		return nil
	}
	
	// Check if path is excluded
	config := p.ctx.PluginConfig()
	if excluded, ok := config["excluded_paths"].([]interface{}); ok {
		path := reqCtx.Path()
		for _, ex := range excluded {
			if exPath, ok := ex.(string); ok && strings.HasPrefix(path, exPath) {
				return nil
			}
		}
	}
	
	// Generate cache key
	cacheKey := p.generateCacheKey(reqCtx)
	
	// Store in cache
	cache := p.ctx.Cache()
	if cache != nil {
		// In a real implementation, we would cache the actual response
		// For this example, we just cache a placeholder
		cacheData := map[string]interface{}{
			"path":      reqCtx.Path(),
			"timestamp": time.Now(),
		}
		
		if err := cache.Set(cacheKey, cacheData, p.cacheDuration); err != nil {
			if logger := p.ctx.Logger(); logger != nil {
				logger.Error(fmt.Sprintf("Failed to cache response: %v", err))
			}
		} else {
			if logger := p.ctx.Logger(); logger != nil {
				logger.Info(fmt.Sprintf("Cached response for %s", reqCtx.Path()))
			}
		}
	}
	
	return nil
}

// handleInvalidation handles cache invalidation events
func (p *CachePlugin) handleInvalidation(event pkg.Event) error {
	if data, ok := event.Data.(map[string]interface{}); ok {
		if pattern, ok := data["pattern"].(string); ok {
			if logger := p.ctx.Logger(); logger != nil {
				logger.Info(fmt.Sprintf("Cache invalidation requested for pattern: %s", pattern))
			}
			
			// In a real implementation, we would invalidate matching cache entries
			// For this example, we just log it
		}
	}
	
	return nil
}

// generateCacheKey generates a cache key for a request
func (p *CachePlugin) generateCacheKey(ctx pkg.Context) string {
	// Create a key based on method, path, and query parameters
	key := fmt.Sprintf("%s:%s:%s", ctx.Method(), ctx.Path(), ctx.Query(""))
	
	// Hash the key for consistent length
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("cache_plugin:%x", hash)
}

// CacheService is an exported service for other plugins
type CacheService struct {
	plugin *CachePlugin
}

// GetHitRate returns the cache hit rate
func (s *CacheService) GetHitRate() float64 {
	total := s.plugin.cacheHits + s.plugin.cacheMisses
	if total == 0 {
		return 0
	}
	return float64(s.plugin.cacheHits) / float64(total) * 100
}

// GetStats returns cache statistics
func (s *CacheService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"hits":     s.plugin.cacheHits,
		"misses":   s.plugin.cacheMisses,
		"hit_rate": s.GetHitRate(),
	}
}

// NewPlugin creates a new instance of the plugin
func NewPlugin() pkg.Plugin {
	return &CachePlugin{}
}
