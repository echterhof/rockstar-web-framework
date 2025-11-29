---
title: "Cache API"
description: "Cache manager interface for caching operations"
category: "api"
tags: ["api", "cache", "performance", "storage"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "../guides/caching.md"
---

# Cache API

## Overview

The `CacheManager` interface provides a comprehensive caching system for the Rockstar Web Framework. It supports both in-memory and distributed caching, with features including TTL-based expiration, tag-based invalidation, request-scoped caching, and atomic operations.

The cache manager is accessible through the Context in request handlers or directly from the Framework instance for application-level caching.

**Primary Use Cases:**
- Caching database query results
- Storing computed values to avoid recalculation
- Session data caching
- API response caching
- Request-scoped temporary data storage

## Interface Definition

```go
type CacheManager interface {
    // Basic operations
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Exists(key string) bool
    Clear() error

    // Batch operations
    GetMultiple(keys []string) (map[string]interface{}, error)
    SetMultiple(items map[string]interface{}, ttl time.Duration) error
    DeleteMultiple(keys []string) error

    // Atomic operations
    Increment(key string, delta int64) (int64, error)
    Decrement(key string, delta int64) (int64, error)

    // TTL management
    Expire(key string, ttl time.Duration) error
    TTL(key string) (time.Duration, error)

    // Request-scoped caching
    GetRequestCache(requestID string) RequestCache
    ClearRequestCache(requestID string) error

    // Pattern-based operations
    Invalidate(pattern string) error
    InvalidateTag(tag string) error
    SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error

    // Maintenance
    CleanupExpired() int
}
```

## Configuration

### CacheConfig

```go
type CacheConfig struct {
    // Type specifies the cache backend type.
    // Supported values: "memory", "distributed"
    // Default: "memory"
    Type string

    // MaxSize specifies the maximum cache size in bytes.
    // 0 means no limit.
    // Default: 0 (unlimited)
    MaxSize int64

    // DefaultTTL specifies the default time-to-live for cache entries.
    // 0 means no expiration.
    // Default: 0 (no expiration)
    DefaultTTL time.Duration
}
```

**Example**:
```go
config := pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    100 * 1024 * 1024, // 100MB
    DefaultTTL: 5 * time.Minute,
}
```

## Basic Operations

### Get

```go
func Get(key string) (interface{}, error)
```

**Description**: Retrieves a value from the cache by key.

**Parameters**:
- `key` (string): Cache key

**Returns**:
- `interface{}`: Cached value if found
- `error`: `ErrCacheKeyNotFound` if key doesn't exist, `ErrCacheExpired` if entry expired

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    value, err := cache.Get("user:123")
    if err == pkg.ErrCacheKeyNotFound {
        // Cache miss - fetch from database
        user, err := ctx.DB().QueryRow("SELECT * FROM users WHERE id = ?", 123)
        if err != nil {
            return ctx.JSON(500, map[string]string{"error": "Database error"})
        }
        
        // Store in cache for future requests
        cache.Set("user:123", user, 5*time.Minute)
        return ctx.JSON(200, user)
    }
    
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Cache error"})
    }
    
    // Cache hit
    return ctx.JSON(200, value)
}
```

### Set

```go
func Set(key string, value interface{}, ttl time.Duration) error
```

**Description**: Stores a value in the cache with an optional TTL (time-to-live).

**Parameters**:
- `key` (string): Cache key
- `value` (interface{}): Value to cache (must be serializable)
- `ttl` (time.Duration): Time-to-live (0 for no expiration, uses DefaultTTL if configured)

**Returns**:
- `error`: Error if caching fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Cache for 10 minutes
    err := cache.Set("config:app", appConfig, 10*time.Minute)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to cache"})
    }
    
    // Cache permanently (no expiration)
    err = cache.Set("static:data", staticData, 0)
    
    return ctx.JSON(200, map[string]string{"message": "Cached successfully"})
}
```

### Delete

```go
func Delete(key string) error
```

**Description**: Removes a value from the cache.

**Parameters**:
- `key` (string): Cache key to delete

**Returns**:
- `error`: Error if deletion fails

**Example**:
```go
func updateUserHandler(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Update user in database
    err := ctx.DB().Exec("UPDATE users SET name = ? WHERE id = ?", newName, userID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Update failed"})
    }
    
    // Invalidate cache
    cache := ctx.Cache()
    cache.Delete(fmt.Sprintf("user:%s", userID))
    
    return ctx.JSON(200, map[string]string{"message": "User updated"})
}
```

### Exists

```go
func Exists(key string) bool
```

**Description**: Checks if a key exists in the cache and is not expired.

**Parameters**:
- `key` (string): Cache key to check

**Returns**:
- `bool`: true if key exists and is not expired, false otherwise

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    if cache.Exists("user:123") {
        // Use cached value
        value, _ := cache.Get("user:123")
        return ctx.JSON(200, value)
    }
    
    // Fetch from database
    user, _ := fetchUser(123)
    cache.Set("user:123", user, 5*time.Minute)
    return ctx.JSON(200, user)
}
```

### Clear

```go
func Clear() error
```

**Description**: Removes all entries from the cache.

**Returns**:
- `error`: Error if clearing fails

**Example**:
```go
func clearCacheHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    if err := cache.Clear(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to clear cache"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cache cleared"})
}
```

## Batch Operations

### GetMultiple

```go
func GetMultiple(keys []string) (map[string]interface{}, error)
```

**Description**: Retrieves multiple values from the cache in a single operation.

**Parameters**:
- `keys` ([]string): List of cache keys to retrieve

**Returns**:
- `map[string]interface{}`: Map of found keys to their values (missing keys are omitted)
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    keys := []string{"user:1", "user:2", "user:3"}
    results, err := cache.GetMultiple(keys)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Cache error"})
    }
    
    // Check which keys were found
    users := make([]interface{}, 0)
    for _, key := range keys {
        if value, found := results[key]; found {
            users = append(users, value)
        }
    }
    
    return ctx.JSON(200, users)
}
```

### SetMultiple

```go
func SetMultiple(items map[string]interface{}, ttl time.Duration) error
```

**Description**: Stores multiple key-value pairs in the cache with the same TTL.

**Parameters**:
- `items` (map[string]interface{}): Map of keys to values
- `ttl` (time.Duration): Time-to-live for all items

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    items := map[string]interface{}{
        "user:1": user1,
        "user:2": user2,
        "user:3": user3,
    }
    
    err := cache.SetMultiple(items, 10*time.Minute)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to cache"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cached successfully"})
}
```

### DeleteMultiple

```go
func DeleteMultiple(keys []string) error
```

**Description**: Removes multiple keys from the cache in a single operation.

**Parameters**:
- `keys` ([]string): List of cache keys to delete

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    keys := []string{"user:1", "user:2", "user:3"}
    err := cache.DeleteMultiple(keys)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to delete"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Deleted successfully"})
}
```

## Atomic Operations

### Increment

```go
func Increment(key string, delta int64) (int64, error)
```

**Description**: Atomically increments a numeric value in the cache. If the key doesn't exist, it's initialized with the delta value.

**Parameters**:
- `key` (string): Cache key
- `delta` (int64): Amount to increment by

**Returns**:
- `int64`: New value after increment
- `error`: Error if value is not numeric or operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Increment page view counter
    newCount, err := cache.Increment("page:views", 1)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to increment"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "views": newCount,
    })
}
```

### Decrement

```go
func Decrement(key string, delta int64) (int64, error)
```

**Description**: Atomically decrements a numeric value in the cache.

**Parameters**:
- `key` (string): Cache key
- `delta` (int64): Amount to decrement by

**Returns**:
- `int64`: New value after decrement
- `error`: Error if value is not numeric or operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Decrement available seats
    remaining, err := cache.Decrement("seats:available", 1)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to decrement"})
    }
    
    if remaining < 0 {
        // Rollback
        cache.Increment("seats:available", 1)
        return ctx.JSON(400, map[string]string{"error": "No seats available"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "remaining": remaining,
    })
}
```

## TTL Management

### Expire

```go
func Expire(key string, ttl time.Duration) error
```

**Description**: Sets or updates the TTL for an existing cache entry.

**Parameters**:
- `key` (string): Cache key
- `ttl` (time.Duration): New time-to-live (0 to remove expiration)

**Returns**:
- `error`: `ErrCacheKeyNotFound` if key doesn't exist, or other error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Extend cache lifetime
    err := cache.Expire("user:123", 30*time.Minute)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to update TTL"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "TTL updated"})
}
```

### TTL

```go
func TTL(key string) (time.Duration, error)
```

**Description**: Returns the remaining time-to-live for a cache entry.

**Parameters**:
- `key` (string): Cache key

**Returns**:
- `time.Duration`: Remaining TTL (0 if no expiration set)
- `error`: `ErrCacheKeyNotFound` if key doesn't exist, `ErrCacheExpired` if already expired

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    ttl, err := cache.TTL("user:123")
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Key not found"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "ttl_seconds": ttl.Seconds(),
    })
}
```

## Request-Scoped Caching

### GetRequestCache

```go
func GetRequestCache(requestID string) RequestCache
```

**Description**: Returns a request-specific cache that is automatically cleaned up after the request completes. Useful for storing temporary data during request processing.

**Parameters**:
- `requestID` (string): Unique request identifier

**Returns**:
- `RequestCache`: Request-scoped cache instance

**Example**:
```go
func middleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    cache := ctx.Cache()
    requestCache := cache.GetRequestCache(ctx.Request().ID)
    
    // Store request-scoped data
    requestCache.Set("start_time", time.Now())
    
    err := next(ctx)
    
    // Access request-scoped data
    if startTime, ok := requestCache.Get("start_time"); ok {
        duration := time.Since(startTime.(time.Time))
        ctx.SetHeader("X-Response-Time", duration.String())
    }
    
    return err
}
```

### ClearRequestCache

```go
func ClearRequestCache(requestID string) error
```

**Description**: Manually clears a request-specific cache. Normally called automatically after request completion.

**Parameters**:
- `requestID` (string): Request identifier

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Clear request cache manually
    err := cache.ClearRequestCache(ctx.Request().ID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to clear"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cleared"})
}
```

## Pattern-Based Operations

### Invalidate

```go
func Invalidate(pattern string) error
```

**Description**: Removes all cache entries matching a pattern. Supports wildcard (*) at the end of the pattern.

**Parameters**:
- `pattern` (string): Pattern to match (e.g., "user:*" matches all keys starting with "user:")

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Invalidate all user caches
    err := cache.Invalidate("user:*")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to invalidate"})
    }
    
    // Invalidate specific pattern
    cache.Invalidate("session:tenant:123:*")
    
    return ctx.JSON(200, map[string]string{"message": "Cache invalidated"})
}
```

### InvalidateTag

```go
func InvalidateTag(tag string) error
```

**Description**: Removes all cache entries associated with a specific tag.

**Parameters**:
- `tag` (string): Tag name

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Invalidate all entries tagged with "products"
    err := cache.InvalidateTag("products")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to invalidate"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "Tagged cache invalidated"})
}
```

### SetWithTags

```go
func SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error
```

**Description**: Stores a value in the cache with associated tags for group invalidation.

**Parameters**:
- `key` (string): Cache key
- `value` (interface{}): Value to cache
- `ttl` (time.Duration): Time-to-live
- `tags` ([]string): List of tags to associate with this entry

**Returns**:
- `error`: Error if operation fails

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Cache product with tags
    err := cache.SetWithTags(
        "product:123",
        product,
        10*time.Minute,
        []string{"products", "category:electronics", "brand:acme"},
    )
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to cache"})
    }
    
    // Later, invalidate all products in a category
    cache.InvalidateTag("category:electronics")
    
    return ctx.JSON(200, map[string]string{"message": "Cached with tags"})
}
```

## Maintenance

### CleanupExpired

```go
func CleanupExpired() int
```

**Description**: Manually removes all expired entries from the cache. Returns the number of entries removed.

**Returns**:
- `int`: Number of expired entries removed

**Example**:
```go
func cleanupHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    count := cache.CleanupExpired()
    
    return ctx.JSON(200, map[string]interface{}{
        "removed": count,
    })
}
```

## Error Handling

The cache manager defines specific error types:

```go
var (
    ErrCacheKeyNotFound = errors.New("cache key not found")
    ErrCacheExpired     = errors.New("cache entry expired")
)
```

**Example**:
```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    value, err := cache.Get("key")
    if err == pkg.ErrCacheKeyNotFound {
        // Handle cache miss
        return ctx.JSON(404, map[string]string{"error": "Not found"})
    }
    if err == pkg.ErrCacheExpired {
        // Handle expired entry
        return ctx.JSON(410, map[string]string{"error": "Expired"})
    }
    if err != nil {
        // Handle other errors
        return ctx.JSON(500, map[string]string{"error": "Cache error"})
    }
    
    return ctx.JSON(200, value)
}
```

## Best Practices

### Cache Key Naming

Use a consistent naming convention for cache keys:

```go
// Good: Hierarchical, descriptive keys
"user:123"
"user:123:profile"
"session:abc:data"
"product:category:electronics:page:1"

// Bad: Unclear, inconsistent keys
"u123"
"data"
"temp"
```

### TTL Selection

Choose appropriate TTLs based on data volatility:

```go
// Frequently changing data: short TTL
cache.Set("stock:price", price, 30*time.Second)

// Moderately changing data: medium TTL
cache.Set("user:profile", profile, 5*time.Minute)

// Rarely changing data: long TTL
cache.Set("config:app", config, 1*time.Hour)

// Static data: no expiration
cache.Set("static:content", content, 0)
```

### Cache Invalidation

Always invalidate cache when underlying data changes:

```go
func updateUser(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Update database
    err := ctx.DB().Exec("UPDATE users SET ... WHERE id = ?", userID)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cache := ctx.Cache()
    cache.Delete(fmt.Sprintf("user:%s", userID))
    cache.Delete(fmt.Sprintf("user:%s:profile", userID))
    
    // Or use pattern-based invalidation
    cache.Invalidate(fmt.Sprintf("user:%s:*", userID))
    
    return ctx.JSON(200, map[string]string{"message": "Updated"})
}
```

### Error Handling

Always handle cache errors gracefully:

```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Try cache first
    value, err := cache.Get("key")
    if err == nil {
        return ctx.JSON(200, value)
    }
    
    // On cache miss or error, fetch from source
    value, err = fetchFromDatabase(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Database error"})
    }
    
    // Try to cache for next time (ignore errors)
    cache.Set("key", value, 5*time.Minute)
    
    return ctx.JSON(200, value)
}
```

## See Also

- [Framework API](framework.md)
- [Context API](context.md)
- [Caching Guide](../guides/caching.md)
- [Performance Guide](../guides/performance.md)
