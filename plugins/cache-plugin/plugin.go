package cacheplugin

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Cache Plugin (Compile-Time)
// ============================================================================
//
// This plugin demonstrates:
// - Caching middleware
// - Cache integration
// - Plugin storage usage
// - Cache key generation
// - Cache invalidation
// - Compile-time plugin registration
//
// Requirements: 1.1, 1.2, 1.3, 2.1, 7.2, 7.5, 14.1, 14.2
// ============================================================================

// init registers the plugin at compile time
func init() {
	pkg.RegisterPlugin("cache-plugin", func() pkg.Plugin {
		return &CachePlugin{}
	})
}

// CachePlugin implements HTTP response caching functionality
type CachePlugin struct {
	ctx pkg.PluginContext
	
	// Configuration
	enabled                bool
	cacheDuration          time.Duration
	maxCacheSize           string
	cacheMethods           []string
	includeQueryParams     bool
	includeHeaders         bool
	customKeyPrefix        string
	excludedPaths          []string
	excludedContentTypes   []string
	invalidateOnMutation   bool
	invalidationPatterns   []string
	compressionEnabled     bool
	compressionThreshold   string
	
	// Plugin storage for cache data
	storage pkg.PluginStorage
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

func (p *CachePlugin) Name() string {
	return "cache-plugin"
}

func (p *CachePlugin) Version() string {
	return "1.0.0"
}

func (p *CachePlugin) Description() string {
	return "HTTP response caching plugin with intelligent cache key generation"
}

func (p *CachePlugin) Author() string {
	return "Rockstar Framework Team"
}

func (p *CachePlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

func (p *CachePlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing cache plugin...\n", p.Name())
	
	p.ctx = ctx
	p.storage = ctx.PluginStorage()
	
	// Parse configuration
	if err := p.parseConfig(ctx.PluginConfig()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Register hooks
	if err := p.registerHooks(); err != nil {
		return fmt.Errorf("failed to register hooks: %w", err)
	}
	
	// Register middleware
	if err := p.registerMiddleware(); err != nil {
		return fmt.Errorf("failed to register middleware: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

func (p *CachePlugin) Start() error {
	fmt.Printf("[%s] Starting cache plugin...\n", p.Name())
	
	// Start background cache cleanup task
	go p.cleanupExpiredCache()
	
	fmt.Printf("[%s] Plugin started\n", p.Name())
	return nil
}

func (p *CachePlugin) Stop() error {
	fmt.Printf("[%s] Stopping cache plugin...\n", p.Name())
	
	// Stop background tasks
	// In a real implementation, use context or channel to signal shutdown
	
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

func (p *CachePlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up cache plugin...\n", p.Name())
	
	// Clear all cached data
	if p.storage != nil {
		if err := p.storage.Clear(); err != nil {
			fmt.Printf("[%s] Warning: Failed to clear cache: %v\n", p.Name(), err)
		}
	}
	
	p.ctx = nil
	p.storage = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

func (p *CachePlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"enabled": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable caching",
		},
		"cache_duration": map[string]interface{}{
			"type":        "duration",
			"default":     "10m",
			"description": "Default cache duration",
		},
		"cache_methods": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"GET", "HEAD"},
			"description": "HTTP methods to cache",
		},
		"include_query_params": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Include query parameters in cache key",
		},
		"custom_key_prefix": map[string]interface{}{
			"type":        "string",
			"default":     "api_cache",
			"description": "Custom prefix for cache keys",
		},
		"excluded_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"/admin", "/api/auth"},
			"description": "Paths to exclude from caching",
		},
		"invalidate_on_mutation": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Invalidate cache on POST/PUT/DELETE requests",
		},
		"compression_enabled": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable response compression in cache",
		},
	}
}

func (p *CachePlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	return p.parseConfig(config)
}

// ============================================================================
// Configuration Parsing
// ============================================================================

func (p *CachePlugin) parseConfig(config map[string]interface{}) error {
	// Enabled
	if val, ok := config["enabled"].(bool); ok {
		p.enabled = val
	} else {
		p.enabled = true
	}
	
	// Cache Duration
	if duration, ok := config["cache_duration"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid cache_duration: %w", err)
		}
		p.cacheDuration = d
	} else {
		p.cacheDuration = 10 * time.Minute
	}
	
	// Cache Methods
	if methods, ok := config["cache_methods"].([]interface{}); ok {
		p.cacheMethods = make([]string, 0, len(methods))
		for _, method := range methods {
			if methodStr, ok := method.(string); ok {
				p.cacheMethods = append(p.cacheMethods, methodStr)
			}
		}
	} else {
		p.cacheMethods = []string{"GET", "HEAD"}
	}
	
	// Include Query Params
	if val, ok := config["include_query_params"].(bool); ok {
		p.includeQueryParams = val
	} else {
		p.includeQueryParams = true
	}
	
	// Custom Key Prefix
	if prefix, ok := config["custom_key_prefix"].(string); ok {
		p.customKeyPrefix = prefix
	} else {
		p.customKeyPrefix = "api_cache"
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
		p.excludedPaths = []string{"/admin", "/api/auth"}
	}
	
	// Invalidate On Mutation
	if val, ok := config["invalidate_on_mutation"].(bool); ok {
		p.invalidateOnMutation = val
	} else {
		p.invalidateOnMutation = true
	}
	
	// Compression Enabled
	if val, ok := config["compression_enabled"].(bool); ok {
		p.compressionEnabled = val
	} else {
		p.compressionEnabled = true
	}
	
	return nil
}

// ============================================================================
// Hook Registration
// ============================================================================

func (p *CachePlugin) registerHooks() error {
	// Register pre-request hook for cache lookup
	err := p.ctx.RegisterHook(pkg.HookTypePreRequest, 150, func(hookCtx pkg.HookContext) error {
		if !p.enabled {
			return nil
		}
		
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		// Check if request is cacheable
		if !p.isCacheable(reqCtx) {
			return nil
		}
		
		// Generate cache key
		cacheKey := p.generateCacheKey(reqCtx)
		
		// Try to get from cache
		if cached, err := p.storage.Get(cacheKey); err == nil && cached != nil {
			// Cache hit - store in context for middleware to use
			hookCtx.Set("cache_hit", true)
			hookCtx.Set("cached_response", cached)
			fmt.Printf("[%s] Cache hit: %s\n", p.Name(), cacheKey)
		}
		
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register pre-request hook: %w", err)
	}
	
	// Register post-request hook for cache storage
	err = p.ctx.RegisterHook(pkg.HookTypePostRequest, 150, func(hookCtx pkg.HookContext) error {
		if !p.enabled {
			return nil
		}
		
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		// Check if request is cacheable
		if !p.isCacheable(reqCtx) {
			return nil
		}
		
		// Check if this was a cache hit (don't re-cache)
		if cacheHit := hookCtx.Get("cache_hit"); cacheHit != nil {
			if hit, ok := cacheHit.(bool); ok && hit {
				return nil
			}
		}
		
		// Generate cache key
		cacheKey := p.generateCacheKey(reqCtx)
		
		// Store response in cache
		// In a real implementation, capture the response data
		cacheData := map[string]interface{}{
			"timestamp": time.Now(),
			"expires":   time.Now().Add(p.cacheDuration),
		}
		
		if err := p.storage.Set(cacheKey, cacheData); err != nil {
			fmt.Printf("[%s] Failed to cache response: %v\n", p.Name(), err)
		} else {
			fmt.Printf("[%s] Cached response: %s\n", p.Name(), cacheKey)
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

func (p *CachePlugin) registerMiddleware() error {
	// Register caching middleware
	cacheMiddleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		if !p.enabled {
			return next(ctx)
		}
		
		// Check if request is cacheable
		if !p.isCacheable(ctx) {
			return next(ctx)
		}
		
		// Generate cache key
		cacheKey := p.generateCacheKey(ctx)
		
		// Try to get from cache
		if cached, err := p.storage.Get(cacheKey); err == nil && cached != nil {
			// Check if cache entry is expired
			if cacheMap, ok := cached.(map[string]interface{}); ok {
				if expires, ok := cacheMap["expires"].(time.Time); ok {
					if time.Now().Before(expires) {
						// Cache hit - return cached response
						fmt.Printf("[%s] Serving from cache: %s\n", p.Name(), cacheKey)
						return ctx.JSON(200, cacheMap)
					}
				}
			}
		}
		
		// Cache miss - call next handler
		err := next(ctx)
		
		// Handle cache invalidation on mutations
		req := ctx.Request()
		if req != nil && req.URL != nil && p.invalidateOnMutation && p.isMutationMethod(req.Method) {
			p.invalidateCache(req.URL.Path)
		}
		
		return err
	}
	
	// Register the middleware
	err := p.ctx.RegisterMiddleware("cache", cacheMiddleware, 150, []string{})
	if err != nil {
		return fmt.Errorf("failed to register cache middleware: %w", err)
	}
	
	fmt.Printf("[%s] Registered cache middleware\n", p.Name())
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (p *CachePlugin) isCacheable(ctx pkg.Context) bool {
	req := ctx.Request()
	if req == nil || req.URL == nil {
		return false
	}
	
	// Check if method is cacheable
	method := req.Method
	cacheable := false
	for _, m := range p.cacheMethods {
		if m == method {
			cacheable = true
			break
		}
	}
	if !cacheable {
		return false
	}
	
	// Check if path is excluded
	path := req.URL.Path
	for _, excluded := range p.excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return false
		}
	}
	
	return true
}

func (p *CachePlugin) isMutationMethod(method string) bool {
	return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}

func (p *CachePlugin) generateCacheKey(ctx pkg.Context) string {
	req := ctx.Request()
	if req == nil || req.URL == nil {
		return p.customKeyPrefix
	}
	
	// Build cache key from request components
	parts := []string{
		p.customKeyPrefix,
		req.Method,
		req.URL.Path,
	}
	
	// Include query parameters if configured
	if p.includeQueryParams {
		// In a real implementation, extract and sort query parameters
		parts = append(parts, "query_params")
	}
	
	// Generate hash of the key components
	key := strings.Join(parts, ":")
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%s:%x", p.customKeyPrefix, hash)
}

func (p *CachePlugin) invalidateCache(path string) {
	// Invalidate cache entries matching the path
	// In a real implementation, iterate through cache keys and remove matching entries
	fmt.Printf("[%s] Invalidating cache for path: %s\n", p.Name(), path)
	
	// For demonstration, we'll just log the invalidation
	// A real implementation would:
	// 1. List all cache keys
	// 2. Match against invalidation patterns
	// 3. Delete matching entries
}

func (p *CachePlugin) cleanupExpiredCache() {
	// Background task to clean up expired cache entries
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		fmt.Printf("[%s] Cleaning up expired cache entries...\n", p.Name())
		
		if p.storage != nil {
			keys, err := p.storage.List()
			if err != nil {
				continue
			}
			
			for _, key := range keys {
				cached, err := p.storage.Get(key)
				if err != nil {
					continue
				}
				
				if cacheMap, ok := cached.(map[string]interface{}); ok {
					if expires, ok := cacheMap["expires"].(time.Time); ok {
						if time.Now().After(expires) {
							p.storage.Delete(key)
							fmt.Printf("[%s] Deleted expired cache entry: %s\n", p.Name(), key)
						}
					}
				}
			}
		}
	}
}
