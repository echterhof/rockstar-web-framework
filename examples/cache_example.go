package main

import (
	"fmt"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Cache Example ===\n")

	// Create a new cache manager
	cache := pkg.NewCacheManager(pkg.CacheConfig{})

	// Example 1: Basic cache operations
	fmt.Println("1. Basic Cache Operations:")
	cache.Set("user:123", map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"role":  "admin",
	}, 0)

	user, err := cache.Get("user:123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Retrieved user: %v\n", user)
	}

	// Example 2: Cache with TTL
	fmt.Println("\n2. Cache with TTL:")
	cache.Set("session:abc123", "session-data", 5*time.Second)

	ttl, _ := cache.TTL("session:abc123")
	fmt.Printf("Session TTL: %v\n", ttl)

	// Example 3: Increment/Decrement operations
	fmt.Println("\n3. Counter Operations:")
	cache.Increment("page:views", 1)
	cache.Increment("page:views", 5)
	views, _ := cache.Get("page:views")
	fmt.Printf("Page views: %v\n", views)

	cache.Decrement("page:views", 2)
	views, _ = cache.Get("page:views")
	fmt.Printf("Page views after decrement: %v\n", views)

	// Example 4: Batch operations
	fmt.Println("\n4. Batch Operations:")
	items := map[string]interface{}{
		"product:1": map[string]interface{}{"name": "Laptop", "price": 999.99},
		"product:2": map[string]interface{}{"name": "Mouse", "price": 29.99},
		"product:3": map[string]interface{}{"name": "Keyboard", "price": 79.99},
	}
	cache.SetMultiple(items, 0)

	products, _ := cache.GetMultiple([]string{"product:1", "product:2", "product:3"})
	fmt.Printf("Retrieved %d products\n", len(products))
	for key, value := range products {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Example 5: Pattern-based invalidation
	fmt.Println("\n5. Pattern-based Invalidation:")
	cache.Set("cache:user:1", "data1", 0)
	cache.Set("cache:user:2", "data2", 0)
	cache.Set("cache:product:1", "data3", 0)

	fmt.Println("Before invalidation:")
	fmt.Printf("  cache:user:1 exists: %v\n", cache.Exists("cache:user:1"))
	fmt.Printf("  cache:user:2 exists: %v\n", cache.Exists("cache:user:2"))
	fmt.Printf("  cache:product:1 exists: %v\n", cache.Exists("cache:product:1"))

	cache.Invalidate("cache:user:*")

	fmt.Println("After invalidating cache:user:*:")
	fmt.Printf("  cache:user:1 exists: %v\n", cache.Exists("cache:user:1"))
	fmt.Printf("  cache:user:2 exists: %v\n", cache.Exists("cache:user:2"))
	fmt.Printf("  cache:product:1 exists: %v\n", cache.Exists("cache:product:1"))

	// Example 6: Request-specific cache
	fmt.Println("\n6. Request-specific Cache:")
	reqCache := cache.GetRequestCache("request-xyz789")

	// Store request-specific data
	reqCache.Set("parsed_body", map[string]interface{}{
		"action": "create",
		"data":   "some data",
	})
	reqCache.Set("user_permissions", []string{"read", "write", "delete"})
	reqCache.Set("computed_value", 42)

	fmt.Printf("Request cache size: %d bytes\n", reqCache.Size())
	fmt.Printf("Request cache keys: %v\n", reqCache.Keys())

	// Retrieve from request cache
	parsedBody := reqCache.Get("parsed_body")
	fmt.Printf("Parsed body: %v\n", parsedBody)

	// Clear request cache when done
	cache.ClearRequestCache("request-xyz789")
	fmt.Println("Request cache cleared")

	// Example 7: Cache expiration
	fmt.Println("\n7. Cache Expiration:")
	cache.Set("temp:data", "temporary", 2*time.Second)
	fmt.Printf("temp:data exists: %v\n", cache.Exists("temp:data"))

	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	fmt.Printf("temp:data exists after expiration: %v\n", cache.Exists("temp:data"))

	// Example 8: Using cache in a handler context
	fmt.Println("\n8. Cache in Handler Context:")
	demonstrateCacheInHandler(cache)

	fmt.Println("\n=== Cache Example Complete ===")
}

// demonstrateCacheInHandler shows how to use cache in a request handler
func demonstrateCacheInHandler(cache pkg.CacheManager) {
	// Simulate a request ID
	requestID := "req-12345"

	// Get request-specific cache
	reqCache := cache.GetRequestCache(requestID)

	// Check if we have cached data
	cachedResult := reqCache.Get("expensive_computation")
	if cachedResult != nil {
		fmt.Printf("Using cached result: %v\n", cachedResult)
		return
	}

	// Simulate expensive computation
	fmt.Println("Performing expensive computation...")
	result := computeExpensiveOperation()

	// Cache the result for this request
	reqCache.Set("expensive_computation", result)
	fmt.Printf("Computed and cached result: %v\n", result)

	// Also cache globally for other requests
	cache.Set("global:expensive_computation", result, 5*time.Minute)

	// Clean up request cache when done
	defer cache.ClearRequestCache(requestID)
}

func computeExpensiveOperation() int {
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	return 42
}
