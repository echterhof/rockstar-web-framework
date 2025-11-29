---
title: "Performance Tuning Guide"
description: "Comprehensive guide to optimizing Rockstar Web Framework applications for production workloads"
category: "guide"
tags: ["performance", "optimization", "tuning", "production"]
version: "1.0.0"
last_updated: "2025-11-29"
---

# Performance Tuning Guide

## Overview

This guide provides comprehensive strategies for optimizing Rockstar Web Framework applications for production workloads. Performance tuning involves balancing resource utilization, throughput, latency, and scalability based on your specific application requirements.

**Key Topics:**
- Connection pool tuning for databases and HTTP clients
- Cache strategy selection and optimization
- Memory management for large files and high-throughput scenarios
- Arena-based memory allocation optimization
- Workload-specific tuning recommendations

## Benchmarking Methodology

Before optimizing, establish baseline metrics to measure improvements.

### Setting Up Benchmarks

```go
// tests/benchmark_test.go
package tests

import (
    "testing"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func BenchmarkHandler(b *testing.B) {
    // Setup
    config := pkg.FrameworkConfig{
        // Your configuration
    }
    app, _ := pkg.New(config)
    
    // Benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Execute operation
    }
}
```

### Key Metrics to Track

1. **Throughput**: Requests per second (RPS)
2. **Latency**: Response time (p50, p95, p99)
3. **Memory**: Heap allocation, GC pressure
4. **CPU**: CPU utilization percentage
5. **Connections**: Active database/HTTP connections
6. **Cache**: Hit rate, miss rate

### Profiling Tools

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Trace analysis
go test -trace=trace.out -bench=.
go tool trace trace.out
```

## Connection Pool Tuning

Connection pooling is critical for database and HTTP client performance. Proper tuning prevents connection exhaustion while minimizing resource waste.

### Database Connection Pool Concepts

Connection pools maintain a set of reusable database connections, avoiding the overhead of creating new connections for each request.

**Pool States:**
- **Open Connections**: Total connections (in-use + idle)
- **In-Use Connections**: Currently executing queries
- **Idle Connections**: Available for reuse
- **Waiting Requests**: Requests waiting for available connections

### Database Connection Pool Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:   "postgres",
    Host:     "db.example.com",
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    
    // Connection pool settings
    MaxOpenConns:    25,                  // Maximum total connections
    MaxIdleConns:    5,                   // Maximum idle connections
    ConnMaxLifetime: 5 * time.Minute,     // Maximum connection lifetime
}
```

### Connection Pool Sizing Strategies

#### Formula-Based Sizing

**Basic Formula:**
```
MaxOpenConns = (Number of CPU cores × 2) + Effective spindle count
```

For cloud databases or SSDs, use:
```
MaxOpenConns = (Number of CPU cores × 4)
```

**Example Calculations:**

```go
// 4-core server with SSD
MaxOpenConns: 4 * 4 = 16

// 8-core server with SSD
MaxOpenConns: 8 * 4 = 32

// High-concurrency API server (16 cores)
MaxOpenConns: 16 * 4 = 64
```

#### Workload-Based Sizing

**Low Concurrency (< 100 concurrent requests):**
```go
DatabaseConfig: pkg.DatabaseConfig{
    MaxOpenConns:    10,
    MaxIdleConns:    2,
    ConnMaxLifetime: 5 * time.Minute,
}
```

**Medium Concurrency (100-1000 concurrent requests):**
```go
DatabaseConfig: pkg.DatabaseConfig{
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}
```

**High Concurrency (> 1000 concurrent requests):**
```go
DatabaseConfig: pkg.DatabaseConfig{
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 3 * time.Minute,
}
```

### MaxIdleConns Tuning

Keep idle connections ready for burst traffic while avoiding resource waste.

**Guidelines:**
- Set to 20-40% of `MaxOpenConns`
- Minimum of 2 for small applications
- Increase for applications with burst traffic patterns

```go
// Steady traffic
MaxIdleConns: MaxOpenConns / 5  // 20%

// Burst traffic
MaxIdleConns: MaxOpenConns / 3  // 33%

// Highly variable traffic
MaxIdleConns: MaxOpenConns / 2  // 50%
```

### ConnMaxLifetime Tuning

Recycle connections periodically to prevent stale connections and load balancer timeouts.

**Guidelines:**
- Set below database server timeout (typically 8 hours)
- Set below load balancer timeout
- Shorter lifetimes for cloud databases with connection limits

```go
// Standard configuration
ConnMaxLifetime: 5 * time.Minute

// Cloud database with connection limits
ConnMaxLifetime: 3 * time.Minute

// Stable on-premise database
ConnMaxLifetime: 10 * time.Minute

// Behind load balancer with 5-minute timeout
ConnMaxLifetime: 4 * time.Minute
```

### Monitoring Connection Pool Health

```go
func monitorDatabasePool(ctx pkg.Context) error {
    db := ctx.DB()
    stats := db.Stats()
    
    // Log statistics
    ctx.Logger().Info("Database pool stats",
        "open", stats.OpenConnections,
        "in_use", stats.InUse,
        "idle", stats.Idle,
        "wait_count", stats.WaitCount,
        "wait_duration", stats.WaitDuration,
    )
    
    // Alert on high wait counts
    if stats.WaitCount > 1000 {
        ctx.Logger().Warn("High connection wait count - consider increasing MaxOpenConns")
    }
    
    // Alert on low utilization
    utilizationRate := float64(stats.InUse) / float64(stats.OpenConnections)
    if utilizationRate < 0.2 && stats.OpenConnections > 10 {
        ctx.Logger().Warn("Low connection utilization - consider reducing MaxOpenConns")
    }
    
    return ctx.JSON(200, stats)
}
```

### Connection Pool Adjustment Strategy

1. **Start Conservative**: Begin with lower values
2. **Monitor Wait Counts**: If `WaitCount` increases rapidly, increase `MaxOpenConns`
3. **Check Database Limits**: Ensure pool size doesn't exceed database connection limit
4. **Monitor Idle Connections**: Adjust `MaxIdleConns` based on idle connection count
5. **Test Under Load**: Use load testing to validate configuration

### HTTP Client Connection Pool

For applications making outbound HTTP requests, configure HTTP client pools:

```go
// Configure HTTP client with connection pooling
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,              // Total idle connections
        MaxIdleConnsPerHost: 10,               // Idle connections per host
        MaxConnsPerHost:     50,               // Total connections per host
        IdleConnTimeout:     90 * time.Second, // Idle connection timeout
        DisableKeepAlives:   false,            // Enable keep-alive
    },
    Timeout: 30 * time.Second,
}
```

**Sizing Guidelines:**

```go
// Low-volume API calls (< 10 RPS)
MaxIdleConnsPerHost: 2
MaxConnsPerHost:     10

// Medium-volume API calls (10-100 RPS)
MaxIdleConnsPerHost: 10
MaxConnsPerHost:     50

// High-volume API calls (> 100 RPS)
MaxIdleConnsPerHost: 20
MaxConnsPerHost:     100
```


## Cache Strategy Selection

Effective caching dramatically improves performance by reducing database queries and expensive computations. Choose the right cache strategy based on your data access patterns.

### Cache Types

#### Request-Level Cache

Isolated cache storage for each request, automatically cleaned up after request completion.

**Use Cases:**
- Expensive computations used multiple times in a request
- Preventing duplicate database queries within a request
- Temporary data that doesn't need persistence

**Configuration:**
```go
// Request cache is automatically available
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    requestCache := cache.GetRequestCache(ctx.Request().ID)
    
    // Use request cache
    requestCache.Set("computed_value", expensiveComputation())
    
    return nil
}
```

**Performance Characteristics:**
- **Overhead**: Very low (in-memory map)
- **Lifetime**: Single request
- **Concurrency**: Request-isolated (no locking needed)
- **Size**: Limited by request context size

#### Application-Level Cache (Memory)

Shared in-memory cache across all requests.

**Use Cases:**
- Frequently accessed data
- Data shared across requests
- Session data
- Configuration values

**Configuration:**
```go
CacheConfig: pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    100 * 1024 * 1024,  // 100 MB
    DefaultTTL: 5 * time.Minute,
}
```

**Performance Characteristics:**
- **Overhead**: Low (in-memory with mutex locking)
- **Lifetime**: Until expiration or eviction
- **Concurrency**: Thread-safe with RWMutex
- **Size**: Configurable with `MaxSize`

#### Distributed Cache

External cache system (Redis, Memcached) for multi-server deployments.

**Use Cases:**
- Multi-server deployments
- Large cache sizes
- Cache persistence across restarts
- Shared cache across services

**Configuration:**
```go
// Implement DistributedCache interface
type RedisCache struct {
    client *redis.Client
}

// Create cache manager with distributed backend
cache := pkg.NewCacheManagerWithDistributed(config, redisCache)
```

**Performance Characteristics:**
- **Overhead**: Medium (network latency)
- **Lifetime**: Configurable persistence
- **Concurrency**: Handled by cache server
- **Size**: Large (limited by cache server)

### Cache Key Design Patterns

Well-designed cache keys improve hit rates and simplify invalidation.

#### Hierarchical Keys

Use colon-separated hierarchical structure:

```go
// Good: Hierarchical and descriptive
"user:123:profile"
"user:123:permissions"
"product:456:inventory"
"tenant:789:config"

// Bad: Flat and ambiguous
"user123profile"
"prod456inv"
```

#### Versioned Keys

Include version numbers for cache invalidation:

```go
// Version in key
cacheKey := fmt.Sprintf("user:%s:v2", userID)

// Version in configuration
const CACHE_VERSION = "v3"
cacheKey := fmt.Sprintf("user:%s:%s", userID, CACHE_VERSION)
```

#### Parameterized Keys

Include query parameters in cache keys:

```go
// Include pagination parameters
cacheKey := fmt.Sprintf("users:page:%d:size:%d", page, pageSize)

// Include filter parameters
cacheKey := fmt.Sprintf("products:category:%s:sort:%s", category, sortBy)

// Include tenant context
cacheKey := fmt.Sprintf("tenant:%s:users:active", tenantID)
```

### Cache Invalidation Strategies

#### Time-Based Expiration (TTL)

Set appropriate TTLs based on data volatility:

```go
// Frequently changing data - short TTL
cache.Set("stock:price:AAPL", price, 30*time.Second)

// Moderately changing data - medium TTL
cache.Set("user:123:profile", user, 5*time.Minute)

// Rarely changing data - long TTL
cache.Set("config:app:settings", config, 1*time.Hour)

// Static data - no expiration
cache.Set("country:list", countries, 0)
```

**TTL Selection Guidelines:**

| Data Type | Volatility | Recommended TTL |
|-----------|-----------|-----------------|
| Real-time data | Very high | 10-30 seconds |
| User sessions | High | 5-15 minutes |
| User profiles | Medium | 5-30 minutes |
| Product catalogs | Low | 30-60 minutes |
| Configuration | Very low | 1-24 hours |
| Static content | None | No expiration |

#### Write-Through Invalidation

Update cache immediately when data changes:

```go
func updateUserHandler(ctx pkg.Context) error {
    db := ctx.DB()
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    // Update database
    err := db.Exec("UPDATE users SET name = ? WHERE id = ?", name, userID)
    if err != nil {
        return err
    }
    
    // Update cache immediately
    user, _ := loadUser(db, userID)
    cache.Set(fmt.Sprintf("user:%s", userID), user, 5*time.Minute)
    
    return ctx.JSON(200, user)
}
```

#### Pattern-Based Invalidation

Invalidate multiple related cache entries:

```go
func updateUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    userID := ctx.Params()["id"]
    
    // Update user
    // ...
    
    // Invalidate all user-related caches
    cache.Invalidate(fmt.Sprintf("user:%s:*", userID))
    
    return ctx.JSON(200, user)
}
```

#### Tag-Based Invalidation

Group related cache entries with tags:

```go
func cacheUserData(ctx pkg.Context, user User) error {
    cache := ctx.Cache()
    
    // Cache with tags
    tags := []string{
        "user",
        fmt.Sprintf("tenant:%s", user.TenantID),
        fmt.Sprintf("role:%s", user.Role),
    }
    
    cacheKey := fmt.Sprintf("user:%s", user.ID)
    return cache.SetWithTags(cacheKey, user, 5*time.Minute, tags)
}

func invalidateTenantData(ctx pkg.Context, tenantID string) error {
    cache := ctx.Cache()
    
    // Invalidate all tenant-related caches
    return cache.InvalidateTag(fmt.Sprintf("tenant:%s", tenantID))
}
```

### Cache Hit Rate Optimization

#### Measuring Cache Hit Rate

```go
type CacheMetrics struct {
    Hits   int64
    Misses int64
    mu     sync.Mutex
}

func (m *CacheMetrics) RecordHit() {
    m.mu.Lock()
    m.Hits++
    m.mu.Unlock()
}

func (m *CacheMetrics) RecordMiss() {
    m.mu.Lock()
    m.Misses++
    m.mu.Unlock()
}

func (m *CacheMetrics) HitRate() float64 {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    total := m.Hits + m.Misses
    if total == 0 {
        return 0
    }
    return float64(m.Hits) / float64(total)
}

// Use in handler
func getUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    metrics := ctx.Get("cache_metrics").(*CacheMetrics)
    
    value, err := cache.Get(key)
    if err == nil {
        metrics.RecordHit()
        return ctx.JSON(200, value)
    }
    
    metrics.RecordMiss()
    // Load from database
    // ...
}
```

#### Improving Hit Rates

**1. Cache Warming**

Pre-populate cache with frequently accessed data:

```go
func warmCache(cache pkg.CacheManager, db pkg.DatabaseManager) error {
    // Load popular items
    rows, err := db.Query("SELECT id, data FROM products ORDER BY views DESC LIMIT 100")
    if err != nil {
        return err
    }
    defer rows.Close()
    
    items := make(map[string]interface{})
    for rows.Next() {
        var id string
        var data interface{}
        rows.Scan(&id, &data)
        items[fmt.Sprintf("product:%s", id)] = data
    }
    
    return cache.SetMultiple(items, 10*time.Minute)
}
```

**2. Predictive Caching**

Cache related data proactively:

```go
func getUserHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    userID := ctx.Params()["id"]
    
    // Get user
    user := getUser(cache, db, userID)
    
    // Proactively cache related data
    go func() {
        // Cache user's recent posts
        posts := getUserPosts(db, userID)
        cache.Set(fmt.Sprintf("user:%s:posts", userID), posts, 5*time.Minute)
        
        // Cache user's permissions
        perms := getUserPermissions(db, userID)
        cache.Set(fmt.Sprintf("user:%s:permissions", userID), perms, 5*time.Minute)
    }()
    
    return ctx.JSON(200, user)
}
```

**3. Longer TTLs for Stable Data**

Increase TTL for data that changes infrequently:

```go
// Analyze data change frequency
func determineTTL(dataType string) time.Duration {
    switch dataType {
    case "user_profile":
        return 10 * time.Minute  // Users update profiles occasionally
    case "product_catalog":
        return 1 * time.Hour     // Products change daily
    case "site_config":
        return 24 * time.Hour    // Config changes rarely
    default:
        return 5 * time.Minute
    }
}
```

### Cache Sizing Recommendations

#### Memory-Based Sizing

Calculate cache size based on available memory:

```go
// Reserve 20-30% of available memory for cache
availableMemory := 4 * 1024 * 1024 * 1024  // 4 GB
cacheSize := int64(float64(availableMemory) * 0.25)  // 1 GB

CacheConfig: pkg.CacheConfig{
    MaxSize: cacheSize,
}
```

#### Item-Based Sizing

Calculate based on expected cache entries:

```go
// Estimate cache size
avgItemSize := 10 * 1024  // 10 KB per item
expectedItems := 10000     // 10,000 items
cacheSize := int64(avgItemSize * expectedItems)  // 100 MB

CacheConfig: pkg.CacheConfig{
    MaxSize: cacheSize,
}
```

#### Workload-Based Sizing

**Low Traffic (< 100 RPS):**
```go
CacheConfig: pkg.CacheConfig{
    MaxSize:    50 * 1024 * 1024,  // 50 MB
    DefaultTTL: 5 * time.Minute,
}
```

**Medium Traffic (100-1000 RPS):**
```go
CacheConfig: pkg.CacheConfig{
    MaxSize:    200 * 1024 * 1024,  // 200 MB
    DefaultTTL: 5 * time.Minute,
}
```

**High Traffic (> 1000 RPS):**
```go
CacheConfig: pkg.CacheConfig{
    MaxSize:    1024 * 1024 * 1024,  // 1 GB
    DefaultTTL: 10 * time.Minute,
}
```

### Cache Strategy Decision Matrix

| Scenario | Cache Type | TTL | Invalidation |
|----------|-----------|-----|--------------|
| User sessions | Memory | 15 min | Time-based |
| Product catalog | Memory/Distributed | 1 hour | Write-through |
| Real-time prices | Memory | 30 sec | Time-based |
| User profiles | Memory/Distributed | 5 min | Write-through |
| Static content | Memory | No expiration | Manual |
| API responses | Request | N/A | Automatic |
| Search results | Memory | 5 min | Pattern-based |
| Multi-tenant data | Distributed | 10 min | Tag-based |


## Memory Management for Large Files

Handling large files efficiently is critical for preventing memory exhaustion and maintaining application performance. The framework provides multiple strategies for memory-efficient file operations.

### Streaming vs Buffering

#### Buffering Approach

Loads entire file into memory before processing.

**Advantages:**
- Simple implementation
- Fast random access
- Easy to manipulate

**Disadvantages:**
- High memory usage
- Risk of OOM for large files
- Slow for large files

**Use When:**
- Files are small (< 10 MB)
- Need random access to file content
- Memory is abundant

```go
func uploadSmallFileHandler(ctx pkg.Context) error {
    // Read entire file into memory
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    // Read all content
    content, err := io.ReadAll(file.File)
    if err != nil {
        return err
    }
    
    // Process content
    processContent(content)
    
    return ctx.JSON(200, map[string]string{"message": "File uploaded"})
}
```

#### Streaming Approach

Processes file in chunks without loading entirely into memory.

**Advantages:**
- Constant memory usage
- Handles files of any size
- Better performance for large files

**Disadvantages:**
- More complex implementation
- No random access
- Requires streaming-compatible operations

**Use When:**
- Files are large (> 10 MB)
- Memory is limited
- Processing can be done sequentially

```go
func uploadLargeFileHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    // Create destination file
    dst, err := os.Create(filepath.Join("/uploads", file.Filename))
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // Stream copy (processes in chunks)
    _, err = io.Copy(dst, file.File)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{"message": "File uploaded"})
}
```

### Chunked Upload Patterns

#### Basic Chunked Upload

```go
func chunkedUploadHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    // Create destination
    dst, err := os.Create(filepath.Join("/uploads", file.Filename))
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // Process in 32KB chunks
    buffer := make([]byte, 32*1024)
    for {
        n, err := file.File.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        // Write chunk
        if _, err := dst.Write(buffer[:n]); err != nil {
            return err
        }
    }
    
    return ctx.JSON(200, map[string]string{"message": "File uploaded"})
}
```

#### Chunked Upload with Progress Tracking

```go
type UploadProgress struct {
    BytesUploaded int64
    TotalBytes    int64
    Percentage    float64
}

func chunkedUploadWithProgressHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    // Get file size
    fileInfo, _ := file.File.Stat()
    totalSize := fileInfo.Size()
    
    dst, err := os.Create(filepath.Join("/uploads", file.Filename))
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // Track progress
    var bytesUploaded int64
    buffer := make([]byte, 32*1024)
    
    for {
        n, err := file.File.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        // Write chunk
        if _, err := dst.Write(buffer[:n]); err != nil {
            return err
        }
        
        // Update progress
        bytesUploaded += int64(n)
        progress := UploadProgress{
            BytesUploaded: bytesUploaded,
            TotalBytes:    totalSize,
            Percentage:    float64(bytesUploaded) / float64(totalSize) * 100,
        }
        
        // Log or broadcast progress
        ctx.Logger().Info("Upload progress", "percentage", progress.Percentage)
    }
    
    return ctx.JSON(200, map[string]string{"message": "File uploaded"})
}
```

### Chunked Download Patterns

#### Streaming Download

```go
func downloadLargeFileHandler(ctx pkg.Context) error {
    filename := ctx.Params()["filename"]
    filepath := filepath.Join("/uploads", filename)
    
    // Open file
    file, err := os.Open(filepath)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "File not found"})
    }
    defer file.Close()
    
    // Get file info
    fileInfo, err := file.Stat()
    if err != nil {
        return err
    }
    
    // Set headers
    ctx.SetHeader("Content-Type", "application/octet-stream")
    ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    ctx.SetHeader("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
    
    // Stream file to response
    _, err = io.Copy(ctx.Response(), file)
    return err
}
```

#### Range Request Support

Support partial downloads for resumable transfers:

```go
func downloadWithRangeHandler(ctx pkg.Context) error {
    filename := ctx.Params()["filename"]
    filepath := filepath.Join("/uploads", filename)
    
    file, err := os.Open(filepath)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "File not found"})
    }
    defer file.Close()
    
    fileInfo, err := file.Stat()
    if err != nil {
        return err
    }
    fileSize := fileInfo.Size()
    
    // Parse Range header
    rangeHeader := ctx.GetHeader("Range")
    if rangeHeader == "" {
        // No range requested - send entire file
        ctx.SetHeader("Content-Length", fmt.Sprintf("%d", fileSize))
        io.Copy(ctx.Response(), file)
        return nil
    }
    
    // Parse range (simplified - production should use proper parser)
    var start, end int64
    fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
    
    if end == 0 || end >= fileSize {
        end = fileSize - 1
    }
    
    // Seek to start position
    file.Seek(start, 0)
    
    // Set headers for partial content
    ctx.Response().WriteHeader(206)  // Partial Content
    ctx.SetHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
    ctx.SetHeader("Content-Length", fmt.Sprintf("%d", end-start+1))
    
    // Copy range
    io.CopyN(ctx.Response(), file, end-start+1)
    return nil
}
```

### Temporary File Management

Use temporary files for processing large uploads:

```go
func processLargeUploadHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    // Create temporary file
    tmpFile, err := os.CreateTemp("", "upload-*.tmp")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name())  // Clean up
    defer tmpFile.Close()
    
    // Stream to temporary file
    _, err = io.Copy(tmpFile, file.File)
    if err != nil {
        return err
    }
    
    // Process temporary file
    tmpFile.Seek(0, 0)  // Reset to beginning
    result, err := processFile(tmpFile)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, result)
}
```

### Memory Limit Configuration

Set limits to prevent memory exhaustion:

```go
// Limit request body size
FrameworkConfig: pkg.FrameworkConfig{
    MaxRequestBodySize: 100 * 1024 * 1024,  // 100 MB
}

// Limit multipart form memory
func uploadHandler(ctx pkg.Context) error {
    // Parse multipart form with memory limit
    err := ctx.Request().ParseMultipartForm(32 << 20)  // 32 MB in memory
    if err != nil {
        return err
    }
    
    // Files larger than 32 MB will be stored in temp files
    file, _, err := ctx.Request().FormFile("file")
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Process file
    // ...
}
```

### Memory-Efficient File Processing Examples

#### CSV Processing

```go
func processLargeCSVHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("csv")
    if err != nil {
        return err
    }
    
    // Create CSV reader (streams data)
    reader := csv.NewReader(file.File)
    
    var count int
    for {
        // Read one record at a time
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        // Process record
        processRecord(record)
        count++
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "CSV processed",
        "records": count,
    })
}
```

#### JSON Streaming

```go
func processLargeJSONHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("json")
    if err != nil {
        return err
    }
    
    // Create JSON decoder (streams data)
    decoder := json.NewDecoder(file.File)
    
    // Read opening bracket
    decoder.Token()
    
    var count int
    for decoder.More() {
        var item map[string]interface{}
        if err := decoder.Decode(&item); err != nil {
            return err
        }
        
        // Process item
        processItem(item)
        count++
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "JSON processed",
        "items":   count,
    })
}
```

#### Image Processing

```go
func processLargeImageHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("image")
    if err != nil {
        return err
    }
    
    // Decode image (streaming)
    img, format, err := image.Decode(file.File)
    if err != nil {
        return err
    }
    
    // Process image (e.g., resize)
    resized := resize(img, 800, 600)
    
    // Create temporary file for output
    tmpFile, err := os.CreateTemp("", "processed-*."+format)
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()
    
    // Encode to temporary file
    switch format {
    case "jpeg":
        jpeg.Encode(tmpFile, resized, nil)
    case "png":
        png.Encode(tmpFile, resized)
    }
    
    // Stream result back
    tmpFile.Seek(0, 0)
    ctx.SetHeader("Content-Type", "image/"+format)
    io.Copy(ctx.Response(), tmpFile)
    
    return nil
}
```

### Best Practices for Large File Handling

1. **Always Use Streaming for Files > 10 MB**
2. **Set Appropriate Memory Limits**
3. **Use Temporary Files for Processing**
4. **Implement Progress Tracking for User Feedback**
5. **Support Range Requests for Downloads**
6. **Clean Up Temporary Files Promptly**
7. **Validate File Size Before Processing**
8. **Use Appropriate Chunk Sizes (32-64 KB)**


## Arena-Based Memory Management

Arena allocation is a memory management technique where memory is allocated from a pre-allocated pool and freed all at once. The Rockstar Web Framework uses arena-based allocation for request contexts to reduce GC pressure and improve performance.

### Arena Allocation Concepts

**Traditional Allocation:**
- Each allocation triggers individual malloc/free
- Garbage collector must track each object
- High GC overhead for many small allocations

**Arena Allocation:**
- Pre-allocate large memory block (arena)
- Allocate from arena sequentially
- Free entire arena at once
- Reduced GC pressure

**Benefits:**
- Faster allocation (no malloc overhead)
- Reduced GC pressure (fewer objects to track)
- Better cache locality
- Predictable memory usage

**Trade-offs:**
- Memory held until arena is freed
- Not suitable for long-lived objects
- Requires careful lifetime management

### Framework's Arena Usage

The framework uses arenas for request contexts:

```go
// Simplified internal implementation
type requestContext struct {
    arena     *arena.Arena
    params    map[string]string
    headers   map[string]string
    cache     map[string]interface{}
    // ... other fields
}

func newRequestContext() *requestContext {
    // Allocate arena for this request
    arena := arena.New(8192)  // 8 KB initial size
    
    return &requestContext{
        arena:   arena,
        params:  make(map[string]string),
        headers: make(map[string]string),
        cache:   make(map[string]interface{}),
    }
}

func (ctx *requestContext) cleanup() {
    // Free entire arena at once
    ctx.arena.Free()
}
```

### Request Context Size Optimization

Minimize request context size to reduce memory usage and improve performance.

#### Measuring Context Size

```go
func measureContextSizeHandler(ctx pkg.Context) error {
    // Get context size (if available)
    size := ctx.Get("context_size")
    
    return ctx.JSON(200, map[string]interface{}{
        "context_size_bytes": size,
    })
}
```

#### Optimization Strategies

**1. Avoid Storing Large Objects in Context**

```go
// Bad: Storing large data in context
func badHandler(ctx pkg.Context) error {
    largeData := loadLargeDataset()  // 10 MB
    ctx.Set("large_data", largeData)  // Stored in arena
    
    // Context now holds 10 MB until request completes
    return ctx.JSON(200, result)
}

// Good: Use references or load on-demand
func goodHandler(ctx pkg.Context) error {
    dataID := loadDatasetID()
    ctx.Set("data_id", dataID)  // Only store ID
    
    // Load data only when needed
    if needData {
        data := loadDatasetByID(dataID)
        processData(data)
    }
    
    return ctx.JSON(200, result)
}
```

**2. Use Request Cache for Temporary Data**

```go
// Use request cache for temporary computations
func optimizedHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    requestCache := cache.GetRequestCache(ctx.Request().ID)
    
    // Store temporary data in request cache
    // (automatically cleaned up after request)
    requestCache.Set("temp_result", expensiveComputation())
    
    return ctx.JSON(200, result)
}
```

**3. Clear Unused Context Data**

```go
func longRunningHandler(ctx pkg.Context) error {
    // Store temporary data
    ctx.Set("temp_data", someData)
    
    // Use data
    processData(ctx.Get("temp_data"))
    
    // Clear when no longer needed
    ctx.Set("temp_data", nil)
    
    // Continue with rest of handler
    return ctx.JSON(200, result)
}
```

### Memory Profiling

Profile memory usage to identify optimization opportunities:

```go
// Enable memory profiling
import _ "net/http/pprof"

func main() {
    // Start pprof server
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
    
    // Start application
    app, _ := pkg.New(config)
    app.Start()
}
```

**Analyze Memory Profile:**

```bash
# Capture heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze with pprof
go tool pprof heap.prof

# Commands in pprof:
# top - Show top memory consumers
# list <function> - Show source code with allocations
# web - Generate visual graph
```

### Optimization Examples

#### Example 1: Minimize Context Allocations

```go
// Before: Multiple allocations
func beforeHandler(ctx pkg.Context) error {
    user := loadUser(ctx.Params()["id"])
    ctx.Set("user", user)
    
    permissions := loadPermissions(user.ID)
    ctx.Set("permissions", permissions)
    
    settings := loadSettings(user.ID)
    ctx.Set("settings", settings)
    
    // Process...
    return ctx.JSON(200, result)
}

// After: Single allocation with struct
type RequestData struct {
    User        *User
    Permissions []string
    Settings    map[string]interface{}
}

func afterHandler(ctx pkg.Context) error {
    data := &RequestData{
        User:        loadUser(ctx.Params()["id"]),
        Permissions: loadPermissions(user.ID),
        Settings:    loadSettings(user.ID),
    }
    ctx.Set("request_data", data)
    
    // Process...
    return ctx.JSON(200, result)
}
```

#### Example 2: Reuse Buffers

```go
// Buffer pool for reuse
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func efficientHandler(ctx pkg.Context) error {
    // Get buffer from pool
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // Use buffer for processing
    n, err := ctx.Request().Body.Read(buffer)
    if err != nil && err != io.EOF {
        return err
    }
    
    // Process buffer[:n]
    result := processBuffer(buffer[:n])
    
    return ctx.JSON(200, result)
}
```

#### Example 3: Lazy Loading

```go
// Lazy load expensive data only when needed
func lazyLoadHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // Store loader function instead of data
    ctx.Set("user_loader", func() *User {
        return loadUser(userID)
    })
    
    // Load only if needed
    if needUserData {
        loader := ctx.Get("user_loader").(func() *User)
        user := loader()
        processUser(user)
    }
    
    return ctx.JSON(200, result)
}
```

### Memory Management Best Practices

1. **Keep Request Context Small**
   - Store only essential data
   - Use references instead of large objects
   - Clear data when no longer needed

2. **Use Request Cache for Temporary Data**
   - Automatically cleaned up after request
   - Isolated per request
   - No manual cleanup needed

3. **Reuse Buffers with sync.Pool**
   - Reduce allocations
   - Lower GC pressure
   - Better performance

4. **Profile Memory Usage**
   - Identify memory hotspots
   - Measure optimization impact
   - Monitor production memory usage

5. **Avoid Memory Leaks**
   - Close resources properly
   - Clear goroutine references
   - Use defer for cleanup

### Monitoring Memory Usage

```go
func memoryStatsHandler(ctx pkg.Context) error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return ctx.JSON(200, map[string]interface{}{
        "alloc_mb":        m.Alloc / 1024 / 1024,
        "total_alloc_mb":  m.TotalAlloc / 1024 / 1024,
        "sys_mb":          m.Sys / 1024 / 1024,
        "num_gc":          m.NumGC,
        "gc_pause_ns":     m.PauseNs[(m.NumGC+255)%256],
        "heap_objects":    m.HeapObjects,
        "heap_alloc_mb":   m.HeapAlloc / 1024 / 1024,
        "heap_sys_mb":     m.HeapSys / 1024 / 1024,
        "heap_idle_mb":    m.HeapIdle / 1024 / 1024,
        "heap_in_use_mb":  m.HeapInuse / 1024 / 1024,
    })
}
```

### GC Tuning

Adjust garbage collector behavior for your workload:

```go
import "runtime/debug"

func init() {
    // Set GC target percentage
    // Lower = more frequent GC, less memory
    // Higher = less frequent GC, more memory
    debug.SetGCPercent(100)  // Default
    
    // For low-latency applications
    debug.SetGCPercent(50)   // More frequent GC
    
    // For high-throughput applications
    debug.SetGCPercent(200)  // Less frequent GC
}
```

**Environment Variable:**
```bash
# Set GC target percentage
export GOGC=100  # Default

# Disable GC (not recommended for production)
export GOGC=off
```


## Workload-Specific Recommendations

Different application workloads require different optimization strategies. This section provides tuning recommendations for common workload patterns.

### High-Throughput APIs

Applications serving many requests per second with low latency requirements.

#### Characteristics
- High request rate (> 1000 RPS)
- Small request/response payloads
- Short-lived requests
- CPU-bound operations

#### Configuration

```go
config := pkg.FrameworkConfig{
    // Database: Larger connection pool
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:          "postgres",
        MaxOpenConns:    50,   // Higher for concurrency
        MaxIdleConns:    15,   // Keep connections ready
        ConnMaxLifetime: 3 * time.Minute,
    },
    
    // Cache: Aggressive caching
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    500 * 1024 * 1024,  // 500 MB
        DefaultTTL: 5 * time.Minute,
    },
    
    // Server: Optimize for throughput
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

#### Optimization Strategies

**1. Aggressive Caching**

```go
func highThroughputHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    cacheKey := generateCacheKey(ctx.Request())
    
    // Try cache first
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Process request
    result := processRequest(ctx)
    
    // Cache aggressively
    cache.Set(cacheKey, result, 5*time.Minute)
    
    return ctx.JSON(200, result)
}
```

**2. Connection Pool Optimization**

```go
// Monitor and adjust based on metrics
func monitorPoolHealth() {
    stats := db.Stats()
    
    // If wait count is high, increase MaxOpenConns
    if stats.WaitCount > 1000 {
        // Increase pool size
    }
    
    // If utilization is low, decrease MaxOpenConns
    utilization := float64(stats.InUse) / float64(stats.OpenConnections)
    if utilization < 0.3 {
        // Decrease pool size
    }
}
```

**3. Minimize Allocations**

```go
// Use sync.Pool for frequently allocated objects
var responsePool = sync.Pool{
    New: func() interface{} {
        return &Response{}
    },
}

func efficientHandler(ctx pkg.Context) error {
    resp := responsePool.Get().(*Response)
    defer responsePool.Put(resp)
    
    // Use response object
    resp.Data = processRequest(ctx)
    
    return ctx.JSON(200, resp)
}
```

**4. Batch Database Operations**

```go
func batchProcessHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Collect IDs
    ids := extractIDs(ctx.Request())
    
    // Single query instead of N queries
    query := "SELECT * FROM users WHERE id IN (?)"
    rows, err := db.Query(query, ids)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    // Process results
    users := scanUsers(rows)
    
    return ctx.JSON(200, users)
}
```

### WebSocket-Heavy Applications

Applications with many concurrent WebSocket connections.

#### Characteristics
- Long-lived connections
- Bidirectional communication
- Real-time updates
- Memory-intensive

#### Configuration

```go
config := pkg.FrameworkConfig{
    // Database: Moderate connection pool
    DatabaseConfig: pkg.DatabaseConfig{
        MaxOpenConns:    25,   // Moderate for background tasks
        MaxIdleConns:    5,
        ConnMaxLifetime: 5 * time.Minute,
    },
    
    // Cache: Request-level caching
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    200 * 1024 * 1024,  // 200 MB
        DefaultTTL: 10 * time.Minute,
    },
    
    // Server: Long timeouts for WebSockets
    ReadTimeout:  0,  // No timeout for WebSocket
    WriteTimeout: 0,  // No timeout for WebSocket
    IdleTimeout:  300 * time.Second,  // 5 minutes
}
```

#### Optimization Strategies

**1. Connection Management**

```go
type ConnectionManager struct {
    connections map[string]*websocket.Conn
    mu          sync.RWMutex
}

func (cm *ConnectionManager) Add(id string, conn *websocket.Conn) {
    cm.mu.Lock()
    cm.connections[id] = conn
    cm.mu.Unlock()
}

func (cm *ConnectionManager) Remove(id string) {
    cm.mu.Lock()
    delete(cm.connections, id)
    cm.mu.Unlock()
}

func (cm *ConnectionManager) Broadcast(message []byte) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    for _, conn := range cm.connections {
        conn.WriteMessage(websocket.TextMessage, message)
    }
}
```

**2. Message Buffering**

```go
func websocketHandler(ctx pkg.Context) error {
    conn, err := ctx.UpgradeWebSocket()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Buffered channel for outgoing messages
    send := make(chan []byte, 256)
    
    // Writer goroutine
    go func() {
        for message := range send {
            conn.WriteMessage(websocket.TextMessage, message)
        }
    }()
    
    // Reader loop
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Process message
        response := processMessage(message)
        
        // Send response (non-blocking)
        select {
        case send <- response:
        default:
            // Buffer full - drop message or handle
        }
    }
    
    close(send)
    return nil
}
```

**3. Memory Management**

```go
// Limit concurrent connections
var (
    maxConnections = 10000
    currentConns   int32
)

func websocketHandler(ctx pkg.Context) error {
    // Check connection limit
    if atomic.LoadInt32(&currentConns) >= int32(maxConnections) {
        return ctx.JSON(503, map[string]string{
            "error": "Connection limit reached",
        })
    }
    
    atomic.AddInt32(&currentConns, 1)
    defer atomic.AddInt32(&currentConns, -1)
    
    // Handle WebSocket connection
    // ...
}
```

**4. Heartbeat/Ping-Pong**

```go
func websocketWithHeartbeat(ctx pkg.Context) error {
    conn, err := ctx.UpgradeWebSocket()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Set ping handler
    conn.SetPingHandler(func(appData string) error {
        conn.WriteMessage(websocket.PongMessage, []byte(appData))
        return nil
    })
    
    // Send periodic pings
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    go func() {
        for range ticker.C {
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }()
    
    // Handle messages
    // ...
}
```

### Database-Intensive Applications

Applications with heavy database usage and complex queries.

#### Characteristics
- Frequent database queries
- Complex joins and aggregations
- Data-intensive operations
- I/O-bound workload

#### Configuration

```go
config := pkg.FrameworkConfig{
    // Database: Optimized connection pool
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:          "postgres",
        MaxOpenConns:    40,   // Higher for database operations
        MaxIdleConns:    10,   // Keep connections ready
        ConnMaxLifetime: 5 * time.Minute,
    },
    
    // Cache: Aggressive database result caching
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    1024 * 1024 * 1024,  // 1 GB
        DefaultTTL: 10 * time.Minute,
    },
    
    // Server: Longer timeouts for slow queries
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
}
```

#### Optimization Strategies

**1. Query Result Caching**

```go
func cachedQueryHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    db := ctx.DB()
    
    cacheKey := "users:active:page:1"
    
    // Try cache first
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Execute query
    rows, err := db.Query("SELECT * FROM users WHERE active = true LIMIT 100")
    if err != nil {
        return err
    }
    defer rows.Close()
    
    users := scanUsers(rows)
    
    // Cache results
    cache.Set(cacheKey, users, 10*time.Minute)
    
    return ctx.JSON(200, users)
}
```

**2. Prepared Statement Reuse**

```go
// Cache prepared statements
var stmtCache = make(map[string]*sql.Stmt)
var stmtMu sync.RWMutex

func getOrPrepareStmt(db pkg.DatabaseManager, query string) (*sql.Stmt, error) {
    stmtMu.RLock()
    stmt, exists := stmtCache[query]
    stmtMu.RUnlock()
    
    if exists {
        return stmt, nil
    }
    
    stmtMu.Lock()
    defer stmtMu.Unlock()
    
    // Double-check after acquiring write lock
    if stmt, exists := stmtCache[query]; exists {
        return stmt, nil
    }
    
    // Prepare statement
    stmt, err := db.Prepare(query)
    if err != nil {
        return nil, err
    }
    
    stmtCache[query] = stmt
    return stmt, nil
}
```

**3. Transaction Batching**

```go
func batchUpdateHandler(ctx pkg.Context) error {
    db := ctx.DB()
    
    // Begin transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Prepare statement once
    stmt, err := tx.Prepare("UPDATE users SET status = ? WHERE id = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    // Execute multiple updates
    for _, update := range updates {
        _, err := stmt.Exec(update.Status, update.ID)
        if err != nil {
            return err
        }
    }
    
    // Commit transaction
    return tx.Commit()
}
```

**4. Read Replicas**

```go
// Separate read and write connections
type DatabaseManager struct {
    writeDB pkg.DatabaseManager
    readDB  pkg.DatabaseManager
}

func (dm *DatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
    // Use read replica for queries
    return dm.readDB.Query(query, args...)
}

func (dm *DatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
    // Use primary for writes
    return dm.writeDB.Exec(query, args...)
}
```

### Mixed Workload Applications

Applications with diverse workload patterns.

#### Configuration

```go
config := pkg.FrameworkConfig{
    // Database: Balanced configuration
    DatabaseConfig: pkg.DatabaseConfig{
        MaxOpenConns:    30,
        MaxIdleConns:    8,
        ConnMaxLifetime: 5 * time.Minute,
    },
    
    // Cache: Moderate caching
    CacheConfig: pkg.CacheConfig{
        Type:       "memory",
        MaxSize:    300 * 1024 * 1024,  // 300 MB
        DefaultTTL: 5 * time.Minute,
    },
    
    // Server: Balanced timeouts
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
}
```

#### Optimization Strategies

**1. Route-Specific Optimization**

```go
// Fast path for simple requests
func fastPathHandler(ctx pkg.Context) error {
    // Minimal processing
    return ctx.JSON(200, simpleResponse)
}

// Slow path for complex requests
func slowPathHandler(ctx pkg.Context) error {
    // Complex processing with caching
    cache := ctx.Cache()
    db := ctx.DB()
    
    // Use caching and optimization strategies
    // ...
}
```

**2. Adaptive Connection Pooling**

```go
func adaptivePoolManager() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := db.Stats()
        
        // Adjust based on metrics
        if stats.WaitCount > 500 {
            // Increase pool size
            adjustPoolSize(+5)
        } else if stats.InUse < stats.OpenConnections/3 {
            // Decrease pool size
            adjustPoolSize(-5)
        }
    }
}
```

**3. Priority Queuing**

```go
type PriorityRequest struct {
    Priority int
    Handler  func(pkg.Context) error
}

func priorityMiddleware(next pkg.HandlerFunc) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        priority := determinePriority(ctx)
        
        if priority == HIGH_PRIORITY {
            // Process immediately
            return next(ctx)
        }
        
        // Queue low-priority requests
        queueRequest(ctx, next)
        return ctx.JSON(202, map[string]string{
            "message": "Request queued",
        })
    }
}
```

## Optimization Decision Matrix

Use this matrix to prioritize optimization efforts:

| Workload Type | Connection Pool | Cache Strategy | Memory Management | Priority |
|---------------|----------------|----------------|-------------------|----------|
| High-Throughput API | Large (50+) | Aggressive (500MB+) | Minimize allocations | High |
| WebSocket-Heavy | Moderate (25) | Request-level | Connection limits | High |
| Database-Intensive | Large (40+) | Query results (1GB+) | Prepared statements | High |
| Mixed Workload | Balanced (30) | Moderate (300MB) | Adaptive | Medium |
| Low Traffic | Small (10) | Minimal (50MB) | Standard | Low |

## Performance Testing Checklist

Before deploying optimizations:

- [ ] Establish baseline metrics
- [ ] Profile CPU and memory usage
- [ ] Load test with realistic traffic
- [ ] Monitor database connection pool
- [ ] Measure cache hit rates
- [ ] Test under peak load
- [ ] Verify GC behavior
- [ ] Check for memory leaks
- [ ] Validate error rates
- [ ] Monitor latency percentiles (p50, p95, p99)

## Troubleshooting Performance Issues

### High Latency

**Symptoms:**
- Slow response times
- High p95/p99 latencies

**Solutions:**
- Add caching
- Optimize database queries
- Increase connection pool
- Add database indexes
- Use prepared statements

### High Memory Usage

**Symptoms:**
- Increasing memory consumption
- Frequent GC pauses

**Solutions:**
- Reduce cache size
- Fix memory leaks
- Use streaming for large files
- Optimize context size
- Adjust GOGC

### Connection Pool Exhaustion

**Symptoms:**
- High wait counts
- Connection timeout errors

**Solutions:**
- Increase MaxOpenConns
- Reduce query execution time
- Check for connection leaks
- Optimize slow queries

### Low Cache Hit Rate

**Symptoms:**
- Cache hit rate < 50%
- High database load

**Solutions:**
- Increase cache TTL
- Improve cache key design
- Implement cache warming
- Review invalidation strategy

## See Also

- [Configuration Guide](configuration.md) - Configuration options
- [Caching Guide](caching.md) - Caching strategies
- [Database Guide](database.md) - Database optimization
- [Monitoring Guide](monitoring.md) - Performance monitoring
- [Platform Optimization Guide](platform-optimization.md) - OS-specific tuning

