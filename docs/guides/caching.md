# Caching

## Overview

The Rockstar Web Framework provides a flexible caching system that supports both application-level and request-level caching. The cache manager offers in-memory caching with support for TTL (time-to-live), pattern-based invalidation, and tag-based cache management.

**Key Features:**
- In-memory caching with configurable size limits
- TTL-based expiration
- Request-level caching for per-request data
- Pattern-based cache invalidation
- Tag-based cache management
- Numeric operations (increment/decrement)
- Batch operations (get/set multiple)
- Distributed cache support (extensible)

## Configuration

### Basic Configuration

Configure caching in your `FrameworkConfig`:

```go
config := pkg.FrameworkConfig{
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    50 * 1024 * 1024,  // 50 MB
        DefaultTTL: 5 * time.Minute,
    },
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Type` | string | "memory" | Cache backend type ("memory" or "distributed") |
| `MaxSize` | int64 | 0 (unlimited) | Maximum cache size in bytes |
| `DefaultTTL` | duration | 0 (no expiration) | Default time-to-live for cache entries |

### Configuration Examples

**Development (Unlimited Cache):**
```go
CacheConfig: pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    0,  // Unlimited
    DefaultTTL: 0,  // No expiration
}
```

**Production (Limited Cache with TTL):**
```go
CacheConfig: pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    100 * 1024 * 1024,  // 100 MB
    DefaultTTL: 10 * time.Minute,
}
```

**High-Performance (Short TTL):**
```go
CacheConfig: pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    500 * 1024 * 1024,  // 500 MB
    DefaultTTL: 1 * time.Minute,
}
```

## Accessing the Cache

Access the cache manager through the context in your handlers:

```go
func myHandler(ctx pkg.Context) error {
    // Get cache manager
    cache := ctx.Cache()
    
    // Use cache operations
    // ...
    
    return nil
}
```

## Basic Cache Operations

### Set

Store a value in the cache with a TTL:

```go
func cacheUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    user := User{
        ID:    123,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    // Set with 5 minute TTL
    err := cache.Set("user:123", user, 5*time.Minute)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "User cached"})
}
```

**Using Default TTL:**
```go
// Set with default TTL from config (0 means use default)
err := cache.Set("user:123", user, 0)
```

**No Expiration:**
```go
// Set without expiration (if DefaultTTL is 0)
err := cache.Set("user:123", user, 0)
```

### Get

Retrieve a value from the cache:

```go
func getUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    // Try to get from cache first
    cacheKey := fmt.Sprintf("user:%s", userID)
    value, err := cache.Get(cacheKey)
    
    if err == nil {
        // Cache hit
        user := value.(User)
        return ctx.JSON(200, user)
    }
    
    if err == pkg.ErrCacheKeyNotFound {
        // Cache miss - fetch from database
        user, err := fetchUserFromDB(userID)
        if err != nil {
            return err
        }
        
        // Store in cache for next time
        cache.Set(cacheKey, user, 5*time.Minute)
        
        return ctx.JSON(200, user)
    }
    
    if err == pkg.ErrCacheExpired {
        // Entry expired - fetch fresh data
        user, err := fetchUserFromDB(userID)
        if err != nil {
            return err
        }
        
        cache.Set(cacheKey, user, 5*time.Minute)
        return ctx.JSON(200, user)
    }
    
    return err
}
```

### Delete

Remove a value from the cache:

```go
func deleteUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    // Delete from database
    err := deleteUserFromDB(userID)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%s", userID)
    cache.Delete(cacheKey)
    
    return ctx.JSON(200, map[string]string{"message": "User deleted"})
}
```

### Exists

Check if a key exists in the cache:

```go
func checkCacheHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    key := ctx.Params()["key"]
    
    exists := cache.Exists(key)
    
    return ctx.JSON(200, map[string]interface{}{
        "key":    key,
        "exists": exists,
    })
}
```

### Clear

Remove all entries from the cache:

```go
func clearCacheHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    err := cache.Clear()
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "Cache cleared"})
}
```

## TTL Management

### Get TTL

Check the remaining time-to-live for a cache entry:

```go
func getTTLHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    key := ctx.Params()["key"]
    
    ttl, err := cache.TTL(key)
    if err == pkg.ErrCacheKeyNotFound {
        return ctx.JSON(404, map[string]string{"error": "Key not found"})
    }
    if err == pkg.ErrCacheExpired {
        return ctx.JSON(410, map[string]string{"error": "Key expired"})
    }
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "key":     key,
        "ttl":     ttl.String(),
        "seconds": int(ttl.Seconds()),
    })
}
```

### Set TTL

Update the TTL for an existing cache entry:

```go
func extendCacheHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    key := ctx.Params()["key"]
    
    // Extend TTL to 10 minutes
    err := cache.Expire(key, 10*time.Minute)
    if err == pkg.ErrCacheKeyNotFound {
        return ctx.JSON(404, map[string]string{"error": "Key not found"})
    }
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "TTL updated"})
}
```

**Remove Expiration:**
```go
// Set TTL to 0 to remove expiration
err := cache.Expire(key, 0)
```

## Batch Operations

### Set Multiple

Store multiple values at once:

```go
func cacheUsersHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    items := map[string]interface{}{
        "user:100": User{ID: 100, Name: "Alice"},
        "user:101": User{ID: 101, Name: "Bob"},
        "user:102": User{ID: 102, Name: "Charlie"},
    }
    
    // Set all with 5 minute TTL
    err := cache.SetMultiple(items, 5*time.Minute)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Users cached",
        "count":   len(items),
    })
}
```

### Get Multiple

Retrieve multiple values at once:

```go
func getUsersHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    keys := []string{"user:100", "user:101", "user:102"}
    values, err := cache.GetMultiple(keys)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Users retrieved",
        "count":   len(values),
        "users":   values,
    })
}
```

### Delete Multiple

Remove multiple values at once:

```go
func deleteUsersHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    keys := []string{"user:100", "user:101", "user:102"}
    err := cache.DeleteMultiple(keys)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Users deleted from cache",
        "count":   len(keys),
    })
}
```

## Numeric Operations

### Increment

Increment a numeric cache value:

```go
func incrementCounterHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Increment page view counter
    newValue, err := cache.Increment("page:views", 1)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Counter incremented",
        "value":   newValue,
    })
}
```

**Increment by Custom Amount:**
```go
// Increment by 10
newValue, err := cache.Increment("score", 10)
```

### Decrement

Decrement a numeric cache value:

```go
func decrementInventoryHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    productID := ctx.Params()["id"]
    
    // Decrement inventory count
    key := fmt.Sprintf("inventory:%s", productID)
    newValue, err := cache.Decrement(key, 1)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message":   "Inventory decremented",
        "remaining": newValue,
    })
}
```

## Cache Invalidation Strategies

### Pattern-Based Invalidation

Invalidate cache entries matching a pattern:

```go
func updateUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    // Update user in database
    err := updateUserInDB(userID, userData)
    if err != nil {
        return err
    }
    
    // Invalidate all user-related cache entries
    cache.Invalidate("user:*")
    
    return ctx.JSON(200, map[string]string{"message": "User updated"})
}
```

**Pattern Examples:**
```go
// Invalidate all user cache entries
cache.Invalidate("user:*")

// Invalidate all product cache entries
cache.Invalidate("product:*")

// Invalidate specific pattern
cache.Invalidate("session:tenant-123:*")

// Invalidate everything (same as Clear)
cache.Invalidate("*")
```

### Tag-Based Invalidation

Use tags to group related cache entries:

```go
func cacheWithTagsHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Cache user data with tags
    user := User{ID: 123, Name: "John", TenantID: "tenant-456"}
    tags := []string{"user", "tenant-456", "active-users"}
    
    err := cache.SetWithTags("user:123", user, 5*time.Minute, tags)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "User cached with tags"})
}

func invalidateByTagHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Invalidate all cache entries tagged with "tenant-456"
    err := cache.InvalidateTag("tenant-456")
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{
        "message": "Cache invalidated by tag",
    })
}
```

## Request-Level Caching

Request-level caching provides isolated cache storage for each request, automatically cleaned up after the request completes.

### Basic Usage

```go
func expensiveOperationHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Get request-specific cache
    requestID := ctx.Request().ID
    requestCache := cache.GetRequestCache(requestID)
    
    // Check if we've already computed this
    if cached := requestCache.Get("computed_result"); cached != nil {
        return ctx.JSON(200, map[string]interface{}{
            "result": cached,
            "cached": true,
        })
    }
    
    // Perform expensive computation
    result := performExpensiveComputation()
    
    // Store in request cache
    requestCache.Set("computed_result", result)
    
    return ctx.JSON(200, map[string]interface{}{
        "result": result,
        "cached": false,
    })
}
```

### Request Cache Operations

```go
func requestCacheExampleHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    requestCache := cache.GetRequestCache(ctx.Request().ID)
    
    // Set values
    requestCache.Set("user_id", 123)
    requestCache.Set("permissions", []string{"read", "write"})
    
    // Get values
    userID := requestCache.Get("user_id")
    permissions := requestCache.Get("permissions")
    
    // Get cache statistics
    size := requestCache.Size()
    keys := requestCache.Keys()
    
    // Delete specific key
    requestCache.Delete("user_id")
    
    // Clear all request cache
    requestCache.Clear()
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id":     userID,
        "permissions": permissions,
        "cache_size":  size,
        "cache_keys":  keys,
    })
}
```

### Automatic Cleanup

Request caches are automatically cleaned up after the request completes:

```go
// In middleware or framework internals
defer func() {
    cache.ClearRequestCache(requestID)
}()
```

## Common Caching Patterns

### Cache-Aside (Lazy Loading)

Load data on demand and cache it:

```go
func getUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // Try cache first
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Cache miss - load from database
    user, err := loadUserFromDB(db, userID)
    if err != nil {
        return err
    }
    
    // Store in cache
    cache.Set(cacheKey, user, 5*time.Minute)
    
    return ctx.JSON(200, user)
}
```

### Write-Through

Update cache when writing to database:

```go
func updateUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return err
    }
    
    // Update database
    err := updateUserInDB(db, userID, user)
    if err != nil {
        return err
    }
    
    // Update cache
    cacheKey := fmt.Sprintf("user:%s", userID)
    cache.Set(cacheKey, user, 5*time.Minute)
    
    return ctx.JSON(200, user)
}
```

### Write-Behind (Write-Back)

Cache writes and persist asynchronously:

```go
func updateUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return err
    }
    
    // Update cache immediately
    cacheKey := fmt.Sprintf("user:%s", userID)
    cache.Set(cacheKey, user, 5*time.Minute)
    
    // Queue for async database update
    queueDatabaseUpdate(userID, user)
    
    return ctx.JSON(200, user)
}
```

### Cache Warming

Pre-populate cache with frequently accessed data:

```go
func warmCacheHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    
    // Load frequently accessed users
    users, err := loadPopularUsers(db)
    if err != nil {
        return err
    }
    
    // Cache them
    items := make(map[string]interface{})
    for _, user := range users {
        key := fmt.Sprintf("user:%d", user.ID)
        items[key] = user
    }
    
    err = cache.SetMultiple(items, 10*time.Minute)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Cache warmed",
        "count":   len(items),
    })
}
```

### Refresh-Ahead

Refresh cache before expiration:

```go
func getUserWithRefreshHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // Get from cache
    cached, err := cache.Get(cacheKey)
    if err == nil {
        // Check TTL
        ttl, _ := cache.TTL(cacheKey)
        
        // Refresh if TTL is low (< 1 minute)
        if ttl < 1*time.Minute {
            go func() {
                user, err := loadUserFromDB(db, userID)
                if err == nil {
                    cache.Set(cacheKey, user, 5*time.Minute)
                }
            }()
        }
        
        return ctx.JSON(200, cached)
    }
    
    // Cache miss - load from database
    user, err := loadUserFromDB(db, userID)
    if err != nil {
        return err
    }
    
    cache.Set(cacheKey, user, 5*time.Minute)
    return ctx.JSON(200, user)
}
```

## Best Practices

### 1. Use Appropriate TTLs

Choose TTLs based on data volatility:

```go
// Frequently changing data - short TTL
cache.Set("stock:price", price, 30*time.Second)

// Moderately changing data - medium TTL
cache.Set("user:profile", user, 5*time.Minute)

// Rarely changing data - long TTL
cache.Set("config:settings", config, 1*time.Hour)

// Static data - no expiration
cache.Set("country:list", countries, 0)
```

### 2. Use Consistent Key Naming

Establish a key naming convention:

```go
// Good: Consistent, hierarchical naming
"user:123"
"user:123:profile"
"user:123:permissions"
"product:456"
"product:456:inventory"

// Bad: Inconsistent naming
"user_123"
"123_user"
"userProfile123"
```

### 3. Handle Cache Misses Gracefully

Always have a fallback for cache misses:

```go
value, err := cache.Get(key)
if err != nil {
    // Fallback to database
    value, err = loadFromDatabase(key)
    if err != nil {
        return err
    }
    // Repopulate cache
    cache.Set(key, value, ttl)
}
```

### 4. Invalidate on Updates

Always invalidate cache when data changes:

```go
// Update database
err := updateUser(db, userID, user)
if err != nil {
    return err
}

// Invalidate cache
cache.Delete(fmt.Sprintf("user:%s", userID))
// Or invalidate related caches
cache.Invalidate(fmt.Sprintf("user:%s:*", userID))
```

### 5. Use Request Cache for Expensive Operations

Cache expensive computations within a request:

```go
requestCache := cache.GetRequestCache(ctx.Request().ID)

// Check request cache first
if result := requestCache.Get("expensive_calc"); result != nil {
    return result
}

// Perform calculation
result := expensiveCalculation()

// Cache for this request
requestCache.Set("expensive_calc", result)
```

### 6. Monitor Cache Performance

Track cache hit rates and adjust strategies:

```go
func monitorCacheHandler(ctx pkg.Context) error {
    // Track hits and misses
    _, err := cache.Get(key)
    if err == nil {
        metrics.IncrementCacheHits()
    } else {
        metrics.IncrementCacheMisses()
    }
    
    // Calculate hit rate
    hitRate := float64(hits) / float64(hits + misses)
    
    return ctx.JSON(200, map[string]interface{}{
        "hit_rate": hitRate,
    })
}
```

### 7. Use Tags for Related Data

Group related cache entries with tags:

```go
// Cache with tags
cache.SetWithTags("user:123", user, 5*time.Minute, 
    []string{"user", "tenant-456", "active"})

// Invalidate all tenant data at once
cache.InvalidateTag("tenant-456")
```

### 8. Implement Cache Stampede Protection

Prevent multiple requests from loading the same data:

```go
var mu sync.Mutex
var loading = make(map[string]bool)

func getWithStampedeProtection(key string) (interface{}, error) {
    // Try cache first
    if value, err := cache.Get(key); err == nil {
        return value, nil
    }
    
    // Check if already loading
    mu.Lock()
    if loading[key] {
        mu.Unlock()
        time.Sleep(100 * time.Millisecond)
        return cache.Get(key)
    }
    loading[key] = true
    mu.Unlock()
    
    // Load data
    value, err := loadFromDatabase(key)
    if err != nil {
        mu.Lock()
        delete(loading, key)
        mu.Unlock()
        return nil, err
    }
    
    // Cache it
    cache.Set(key, value, 5*time.Minute)
    
    mu.Lock()
    delete(loading, key)
    mu.Unlock()
    
    return value, nil
}
```

## Troubleshooting

### High Memory Usage

**Problem:** Cache consuming too much memory

**Solutions:**
- Set `MaxSize` limit in configuration
- Reduce `DefaultTTL` to expire entries faster
- Implement cache eviction policy
- Use more aggressive invalidation

### Cache Misses

**Problem:** Low cache hit rate

**Solutions:**
- Increase TTL for stable data
- Implement cache warming
- Use refresh-ahead strategy
- Review key naming consistency

### Stale Data

**Problem:** Cache serving outdated data

**Solutions:**
- Reduce TTL
- Implement proper invalidation on updates
- Use write-through caching
- Add cache versioning

### Cache Stampede

**Problem:** Multiple requests loading same data simultaneously

**Solutions:**
- Implement stampede protection
- Use request-level caching
- Add mutex locks for critical keys
- Implement probabilistic early expiration

## See Also

- [Configuration Guide](configuration.md) - Cache configuration options
- [Context Guide](context.md) - Accessing cache through context
- [Performance Guide](performance.md) - Performance optimization with caching
- [API Reference: Cache](../api/cache.md) - Complete API documentation
