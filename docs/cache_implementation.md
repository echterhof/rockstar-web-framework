# Cache Implementation

## Overview

The Rockstar Web Framework provides a comprehensive caching system with support for both in-memory and distributed caching. The cache system includes request-specific caching using arena-like memory management for optimal performance.

## Features

- **In-Memory Caching**: Fast local cache with TTL support
- **Request-Specific Cache**: Arena-like memory management for request-scoped data
- **Distributed Cache Support**: Optional integration with distributed cache backends
- **TTL Management**: Automatic expiration of cache entries
- **Batch Operations**: Efficient multi-key operations
- **Pattern-Based Invalidation**: Wildcard pattern matching for cache invalidation
- **Tag-Based Invalidation**: Group cache entries by tags for bulk invalidation
- **Atomic Operations**: Thread-safe increment/decrement operations
- **Type Preservation**: Stores any Go type without serialization overhead

## Architecture

### CacheManager Interface

The `CacheManager` interface provides the main caching API:

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
    
    // Request-specific cache
    GetRequestCache(requestID string) RequestCache
    ClearRequestCache(requestID string) error
    
    // Invalidation
    Invalidate(pattern string) error
    InvalidateTag(tag string) error
}
```

### RequestCache Interface

Request-specific caching for per-request data:

```go
type RequestCache interface {
    Get(key string) interface{}
    Set(key string, value interface{})
    Delete(key string)
    Clear()
    Size() int64
    Keys() []string
}
```

## Usage

### Basic Cache Operations

```go
// Create cache manager
cache := pkg.NewCacheManager()

// Set a value
cache.Set("user:123", userData, 0) // No expiration

// Set with TTL
cache.Set("session:abc", sessionData, 30*time.Minute)

// Get a value
value, err := cache.Get("user:123")
if err == pkg.ErrCacheKeyNotFound {
    // Key doesn't exist
}

// Check existence
if cache.Exists("user:123") {
    // Key exists and is not expired
}

// Delete a value
cache.Delete("user:123")
```

### TTL Management

```go
// Set with expiration
cache.Set("temp:data", "value", 5*time.Minute)

// Check remaining TTL
ttl, err := cache.TTL("temp:data")
if err == nil {
    fmt.Printf("Expires in: %v\n", ttl)
}

// Update expiration
cache.Expire("temp:data", 10*time.Minute)
```

### Batch Operations

```go
// Set multiple values at once
items := map[string]interface{}{
    "product:1": product1,
    "product:2": product2,
    "product:3": product3,
}
cache.SetMultiple(items, 1*time.Hour)

// Get multiple values
keys := []string{"product:1", "product:2", "product:3"}
results, err := cache.GetMultiple(keys)

// Delete multiple values
cache.DeleteMultiple(keys)
```

### Atomic Operations

```go
// Increment counter
views, err := cache.Increment("page:views", 1)

// Decrement counter
remaining, err := cache.Decrement("quota:remaining", 1)

// Initialize counter if not exists
cache.Increment("new:counter", 10) // Sets to 10
```

### Request-Specific Cache

```go
// In a request handler
func handleRequest(ctx pkg.Context) error {
    requestID := ctx.Request().ID
    reqCache := ctx.Cache().GetRequestCache(requestID)
    
    // Check cache first
    if cached := reqCache.Get("parsed_data"); cached != nil {
        return processData(cached)
    }
    
    // Parse and cache
    data := parseRequestData(ctx)
    reqCache.Set("parsed_data", data)
    
    // Cache is automatically cleaned up after request
    defer ctx.Cache().ClearRequestCache(requestID)
    
    return processData(data)
}
```

### Pattern-Based Invalidation

```go
// Cache user data
cache.Set("user:1:profile", profile1, 0)
cache.Set("user:1:settings", settings1, 0)
cache.Set("user:2:profile", profile2, 0)

// Invalidate all user:1 data
cache.Invalidate("user:1:*")

// user:1:profile and user:1:settings are removed
// user:2:profile remains
```

### Tag-Based Invalidation

```go
// Set with tags (using extended API)
cacheImpl := cache.(*pkg.CacheManagerImpl)
cacheImpl.SetWithTags("article:123", article, 0, []string{"articles", "user:456"})
cacheImpl.SetWithTags("article:124", article2, 0, []string{"articles", "user:456"})

// Invalidate all articles by user:456
cache.InvalidateTag("user:456")
```

## Integration with Context

The cache manager is accessible through the request context:

```go
func myHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Use cache
    value, err := cache.Get("key")
    if err != nil {
        // Compute value
        value = computeValue()
        cache.Set("key", value, 5*time.Minute)
    }
    
    return ctx.JSON(200, value)
}
```

## Request-Specific Cache Benefits

The request-specific cache provides several benefits:

1. **Automatic Cleanup**: Cache is cleared when request completes
2. **Memory Efficiency**: Arena-like allocation reduces GC pressure
3. **Isolation**: Each request has its own cache namespace
4. **Performance**: No serialization overhead for in-request data

### Use Cases

- Caching parsed request data
- Storing intermediate computation results
- Caching database query results within a request
- Storing user permissions for the request duration

## Distributed Cache Support

The cache manager supports optional distributed cache backends:

```go
// Create with distributed cache
distributed := NewRedisCache(redisClient)
cache := pkg.NewCacheManagerWithDistributed(distributed)

// Operations automatically sync to distributed cache
cache.Set("key", "value", 0) // Stored in both local and distributed cache
```

### Distributed Cache Interface

```go
type DistributedCache interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
    GetMultiple(keys []string) (map[string]interface{}, error)
    SetMultiple(items map[string]interface{}, ttl time.Duration) error
    DeleteMultiple(keys []string) error
    Invalidate(pattern string) error
    InvalidateTag(tag string) error
}
```

## Performance Considerations

### Memory Management

- Request caches use arena-like allocation for efficient memory usage
- Expired entries are lazily cleaned up on access
- Manual cleanup available via `CleanupExpired()` method

### Concurrency

- All operations are thread-safe using read-write locks
- Batch operations minimize lock contention
- Request caches are isolated per request

### Best Practices

1. **Use TTL**: Always set appropriate TTL for cache entries
2. **Request Cache**: Use request cache for request-scoped data
3. **Batch Operations**: Use batch operations for multiple keys
4. **Pattern Invalidation**: Use patterns for bulk invalidation
5. **Monitor Size**: Monitor request cache size to prevent memory issues

## Error Handling

The cache system defines specific errors:

```go
var (
    ErrCacheKeyNotFound = errors.New("cache key not found")
    ErrCacheExpired     = errors.New("cache entry expired")
)
```

Handle errors appropriately:

```go
value, err := cache.Get("key")
switch err {
case nil:
    // Use value
case pkg.ErrCacheKeyNotFound:
    // Key doesn't exist, compute value
case pkg.ErrCacheExpired:
    // Key expired, recompute value
default:
    // Other error
}
```

## Testing

The cache implementation includes comprehensive unit tests:

```bash
go test -v -run TestCacheManager ./pkg
```

Test coverage includes:
- Basic operations (Get, Set, Delete, Exists)
- TTL and expiration
- Batch operations
- Atomic operations (Increment, Decrement)
- Request-specific cache
- Pattern-based invalidation
- Concurrent access
- Type preservation

## Example Application

See `examples/cache_example.go` for a complete example demonstrating:
- Basic cache operations
- TTL management
- Counter operations
- Batch operations
- Pattern-based invalidation
- Request-specific cache
- Cache expiration
- Integration with handlers

Run the example:

```bash
go run examples/cache_example.go
```

## Configuration

Cache behavior can be configured through the cache manager:

```go
// Create cache manager
cache := pkg.NewCacheManager()

// Periodic cleanup of expired entries
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        count := cache.(*pkg.CacheManagerImpl).CleanupExpired()
        log.Printf("Cleaned up %d expired entries", count)
    }
}()
```

## Requirements Satisfied

This implementation satisfies the following requirements:

- Request-specific cache using arena package
- Cache access through context

The cache system provides high-performance caching with both local and distributed support, integrated seamlessly with the framework's context system.
