//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    50 * 1024 * 1024, // 50 MB
			DefaultTTL: 5 * time.Minute,
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Basic cache operations
	router.GET("/api/cache/set", setCacheHandler)
	router.GET("/api/cache/get/:key", getCacheHandler)
	router.GET("/api/cache/delete/:key", deleteCacheHandler)
	router.GET("/api/cache/exists/:key", existsCacheHandler)

	// TTL management
	router.GET("/api/cache/ttl/:key", getTTLHandler)
	router.GET("/api/cache/expire/:key", expireCacheHandler)

	// Multiple operations
	router.GET("/api/cache/set-multiple", setMultipleHandler)
	router.GET("/api/cache/get-multiple", getMultipleHandler)

	// Numeric operations
	router.GET("/api/cache/increment/:key", incrementHandler)
	router.GET("/api/cache/decrement/:key", decrementHandler)

	// Cache invalidation strategies
	router.GET("/api/cache/invalidate", invalidateHandler)
	router.GET("/api/cache/clear", clearCacheHandler)

	// Request-level caching
	router.GET("/api/cache/request-cache", requestCacheHandler)

	// Cache statistics
	router.GET("/api/cache/stats", cacheStatsHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Cache Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Basic cache operations")
	fmt.Println("  curl http://localhost:8080/api/cache/set")
	fmt.Println("  curl http://localhost:8080/api/cache/get/user:123")
	fmt.Println("  curl http://localhost:8080/api/cache/exists/user:123")
	fmt.Println("  curl http://localhost:8080/api/cache/delete/user:123")
	fmt.Println()
	fmt.Println("  # TTL management")
	fmt.Println("  curl http://localhost:8080/api/cache/ttl/user:123")
	fmt.Println("  curl http://localhost:8080/api/cache/expire/user:123")
	fmt.Println()
	fmt.Println("  # Multiple operations")
	fmt.Println("  curl http://localhost:8080/api/cache/set-multiple")
	fmt.Println("  curl http://localhost:8080/api/cache/get-multiple")
	fmt.Println()
	fmt.Println("  # Numeric operations")
	fmt.Println("  curl http://localhost:8080/api/cache/increment/counter")
	fmt.Println("  curl http://localhost:8080/api/cache/decrement/counter")
	fmt.Println()
	fmt.Println("  # Cache invalidation")
	fmt.Println("  curl http://localhost:8080/api/cache/invalidate")
	fmt.Println("  curl http://localhost:8080/api/cache/clear")
	fmt.Println()
	fmt.Println("  # Request-level caching")
	fmt.Println("  curl http://localhost:8080/api/cache/request-cache")
	fmt.Println()
	fmt.Println("  # Cache statistics")
	fmt.Println("  curl http://localhost:8080/api/cache/stats")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - Basic Cache Operations
// ============================================================================

// setCacheHandler demonstrates setting values in cache with TTL
func setCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	// Set a simple value with 1 minute TTL
	err := cache.Set("user:123", map[string]interface{}{
		"id":    123,
		"name":  "John Doe",
		"email": "john@example.com",
	}, 1*time.Minute)

	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set cache",
		})
	}

	// Set a value with default TTL (from config: 5 minutes)
	err = cache.Set("product:456", map[string]interface{}{
		"id":    456,
		"name":  "Guitar",
		"price": 599.99,
	}, 0) // 0 means use default TTL

	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set cache with default TTL",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache values set successfully",
		"keys": []string{
			"user:123 (TTL: 1 minute)",
			"product:456 (TTL: default 5 minutes)",
		},
	})
}

// getCacheHandler demonstrates retrieving values from cache
func getCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	value, err := cache.Get(key)
	if err != nil {
		if err == pkg.ErrCacheKeyNotFound {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "Cache key not found",
				"key":     key,
				"message": "The key does not exist or has expired",
			})
		}
		if err == pkg.ErrCacheExpired {
			return ctx.JSON(410, map[string]interface{}{
				"error":   "Cache entry expired",
				"key":     key,
				"message": "The key existed but has expired",
			})
		}
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to get cache",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache value retrieved successfully",
		"key":     key,
		"value":   value,
	})
}

// deleteCacheHandler demonstrates deleting values from cache
func deleteCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	err := cache.Delete(key)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to delete cache",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache value deleted successfully",
		"key":     key,
	})
}

// existsCacheHandler demonstrates checking if a key exists in cache
func existsCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	exists := cache.Exists(key)

	return ctx.JSON(200, map[string]interface{}{
		"key":    key,
		"exists": exists,
	})
}

// ============================================================================
// Handler Functions - TTL Management
// ============================================================================

// getTTLHandler demonstrates getting the TTL of a cache entry
func getTTLHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	ttl, err := cache.TTL(key)
	if err != nil {
		if err == pkg.ErrCacheKeyNotFound {
			return ctx.JSON(404, map[string]interface{}{
				"error": "Cache key not found",
				"key":   key,
			})
		}
		if err == pkg.ErrCacheExpired {
			return ctx.JSON(410, map[string]interface{}{
				"error": "Cache entry expired",
				"key":   key,
			})
		}
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to get TTL",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"key":     key,
		"ttl":     ttl.String(),
		"seconds": int(ttl.Seconds()),
	})
}

// expireCacheHandler demonstrates setting a new TTL for a cache entry
func expireCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	// Set new TTL to 30 seconds
	err := cache.Expire(key, 30*time.Second)
	if err != nil {
		if err == pkg.ErrCacheKeyNotFound {
			return ctx.JSON(404, map[string]interface{}{
				"error": "Cache key not found",
				"key":   key,
			})
		}
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set expiration",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache expiration updated",
		"key":     key,
		"new_ttl": "30 seconds",
	})
}

// ============================================================================
// Handler Functions - Multiple Operations
// ============================================================================

// setMultipleHandler demonstrates setting multiple cache values at once
func setMultipleHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	items := map[string]interface{}{
		"user:100": map[string]interface{}{"id": 100, "name": "Alice"},
		"user:101": map[string]interface{}{"id": 101, "name": "Bob"},
		"user:102": map[string]interface{}{"id": 102, "name": "Charlie"},
	}

	err := cache.SetMultiple(items, 2*time.Minute)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set multiple cache values",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Multiple cache values set successfully",
		"count":   len(items),
		"keys":    []string{"user:100", "user:101", "user:102"},
		"ttl":     "2 minutes",
	})
}

// getMultipleHandler demonstrates getting multiple cache values at once
func getMultipleHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	keys := []string{"user:100", "user:101", "user:102"}
	values, err := cache.GetMultiple(keys)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to get multiple cache values",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Multiple cache values retrieved",
		"count":   len(values),
		"values":  values,
	})
}

// ============================================================================
// Handler Functions - Numeric Operations
// ============================================================================

// incrementHandler demonstrates incrementing a numeric cache value
func incrementHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	// Increment by 1
	newValue, err := cache.Increment(key, 1)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to increment cache value",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":   "Cache value incremented",
		"key":       key,
		"new_value": newValue,
	})
}

// decrementHandler demonstrates decrementing a numeric cache value
func decrementHandler(ctx pkg.Context) error {
	cache := ctx.Cache()
	key := ctx.Params()["key"]

	// Decrement by 1
	newValue, err := cache.Decrement(key, 1)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to decrement cache value",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":   "Cache value decremented",
		"key":       key,
		"new_value": newValue,
	})
}

// ============================================================================
// Handler Functions - Cache Invalidation Strategies
// ============================================================================

// invalidateHandler demonstrates pattern-based cache invalidation
func invalidateHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	// Invalidate all user cache entries using wildcard pattern
	err := cache.Invalidate("user:*")
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to invalidate cache",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache invalidated successfully",
		"pattern": "user:*",
		"note":    "All keys matching 'user:*' have been removed",
	})
}

// clearCacheHandler demonstrates clearing all cache entries
func clearCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	err := cache.Clear()
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to clear cache",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "All cache entries cleared successfully",
		"warning": "This removes ALL cached data",
	})
}

// ============================================================================
// Handler Functions - Request-Level Caching
// ============================================================================

// requestCacheHandler demonstrates request-specific caching
func requestCacheHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	// Get request-specific cache (isolated per request)
	requestID := ctx.Request().ID
	requestCache := cache.GetRequestCache(requestID)

	// Store data in request cache
	requestCache.Set("computed_value", "expensive_calculation_result")
	requestCache.Set("user_data", map[string]interface{}{
		"id":   123,
		"name": "Request User",
	})

	// Retrieve data from request cache
	computedValue := requestCache.Get("computed_value")
	userData := requestCache.Get("user_data")

	// Get cache statistics
	cacheSize := requestCache.Size()
	cacheKeys := requestCache.Keys()

	return ctx.JSON(200, map[string]interface{}{
		"message":        "Request-level cache demonstrated",
		"request_id":     requestID,
		"computed_value": computedValue,
		"user_data":      userData,
		"cache_size":     cacheSize,
		"cache_keys":     cacheKeys,
		"note":           "Request cache is automatically cleared after the request completes",
	})
}

// ============================================================================
// Handler Functions - Cache Statistics
// ============================================================================

// cacheStatsHandler demonstrates cache statistics and monitoring
func cacheStatsHandler(ctx pkg.Context) error {
	cache := ctx.Cache()

	// Set some test data
	cache.Set("stat:test1", "value1", 1*time.Minute)
	cache.Set("stat:test2", "value2", 2*time.Minute)
	cache.Set("stat:test3", "value3", 3*time.Minute)

	// Check existence
	exists1 := cache.Exists("stat:test1")
	exists2 := cache.Exists("stat:test2")
	exists3 := cache.Exists("stat:test3")

	// Get TTLs
	ttl1, _ := cache.TTL("stat:test1")
	ttl2, _ := cache.TTL("stat:test2")
	ttl3, _ := cache.TTL("stat:test3")

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache statistics",
		"entries": []map[string]interface{}{
			{
				"key":    "stat:test1",
				"exists": exists1,
				"ttl":    ttl1.String(),
			},
			{
				"key":    "stat:test2",
				"exists": exists2,
				"ttl":    ttl2.String(),
			},
			{
				"key":    "stat:test3",
				"exists": exists3,
				"ttl":    ttl3.String(),
			},
		},
		"note": "These are sample cache entries for demonstration",
	})
}
