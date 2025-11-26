package pkg

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrCacheKeyNotFound is returned when a cache key is not found
	ErrCacheKeyNotFound = errors.New("cache key not found")
	// ErrCacheExpired is returned when a cache entry has expired
	ErrCacheExpired = errors.New("cache entry expired")
)

// CacheConfig holds configuration options for the cache system
type CacheConfig struct {
	// Type specifies the cache backend type.
	// Supported values: "memory", "distributed"
	// Default: "memory"
	Type string

	// MaxSize specifies the maximum cache size in bytes.
	// 0 means no limit. Negative values are normalized to 0.
	// Default: 0 (unlimited)
	MaxSize int64

	// DefaultTTL specifies the default time-to-live for cache entries.
	// 0 means no expiration. Negative values are normalized to 0.
	// Default: 0 (no expiration)
	DefaultTTL time.Duration
}

// cacheEntry represents a single cache entry with expiration
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
	tags      []string
}

// isExpired checks if the cache entry has expired
func (e *cacheEntry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// cacheManagerImpl implements the CacheManager interface
type cacheManagerImpl struct {
	// In-memory cache storage
	cache map[string]*cacheEntry
	mu    sync.RWMutex

	// Request-specific caches using arena-like approach
	requestCaches map[string]*requestCacheImpl
	requestMu     sync.RWMutex

	// Tag-based invalidation
	tagIndex map[string][]string // tag -> []keys
	tagMu    sync.RWMutex

	// Distributed cache support (optional)
	distributed DistributedCache

	// Configuration
	config CacheConfig
}

// NewCacheManager creates a new cache manager instance with the given configuration
func NewCacheManager(config CacheConfig) CacheManager {
	// Apply default values for zero values
	config.ApplyDefaults()

	return &cacheManagerImpl{
		cache:         make(map[string]*cacheEntry),
		requestCaches: make(map[string]*requestCacheImpl),
		tagIndex:      make(map[string][]string),
		config:        config,
	}
}

// NewCacheManagerWithDistributed creates a cache manager with distributed cache support
func NewCacheManagerWithDistributed(config CacheConfig, distributed DistributedCache) CacheManager {
	// Apply default values for zero values
	config.ApplyDefaults()
	// Override Type to "distributed" when using distributed cache
	if config.Type == "memory" {
		config.Type = "distributed"
	}

	return &cacheManagerImpl{
		cache:         make(map[string]*cacheEntry),
		requestCaches: make(map[string]*requestCacheImpl),
		tagIndex:      make(map[string][]string),
		distributed:   distributed,
		config:        config,
	}
}

// Get retrieves a value from cache
func (c *cacheManagerImpl) Get(key string) (interface{}, error) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		// Try distributed cache if available
		if c.distributed != nil {
			return c.distributed.Get(key)
		}
		return nil, ErrCacheKeyNotFound
	}

	if entry.isExpired() {
		// Return expired error before cleanup
		err := ErrCacheExpired
		// Clean up expired entry
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		return nil, err
	}

	return entry.value, nil
}

// Set stores a value in cache with TTL
func (c *cacheManagerImpl) Set(key string, value interface{}, ttl time.Duration) error {
	entry := &cacheEntry{
		value: value,
	}

	// Use DefaultTTL if ttl is 0 and DefaultTTL is configured
	effectiveTTL := ttl
	if ttl == 0 && c.config.DefaultTTL > 0 {
		effectiveTTL = c.config.DefaultTTL
	}

	if effectiveTTL > 0 {
		entry.expiresAt = time.Now().Add(effectiveTTL)
	}

	c.mu.Lock()
	c.cache[key] = entry
	c.mu.Unlock()

	// Also set in distributed cache if available
	if c.distributed != nil {
		return c.distributed.Set(key, value, effectiveTTL)
	}

	return nil
}

// Delete removes a value from cache
func (c *cacheManagerImpl) Delete(key string) error {
	c.mu.Lock()
	delete(c.cache, key)
	c.mu.Unlock()

	// Also delete from distributed cache if available
	if c.distributed != nil {
		return c.distributed.Delete(key)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *cacheManagerImpl) Exists(key string) bool {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	if entry.isExpired() {
		return false
	}

	return true
}

// Clear removes all entries from cache
func (c *cacheManagerImpl) Clear() error {
	c.mu.Lock()
	c.cache = make(map[string]*cacheEntry)
	c.mu.Unlock()

	c.tagMu.Lock()
	c.tagIndex = make(map[string][]string)
	c.tagMu.Unlock()

	// Also clear distributed cache if available
	if c.distributed != nil {
		return c.distributed.Clear()
	}

	return nil
}

// GetMultiple retrieves multiple values from cache
func (c *cacheManagerImpl) GetMultiple(keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	c.mu.RLock()
	for _, key := range keys {
		if entry, exists := c.cache[key]; exists && !entry.isExpired() {
			result[key] = entry.value
		}
	}
	c.mu.RUnlock()

	return result, nil
}

// SetMultiple stores multiple values in cache
func (c *cacheManagerImpl) SetMultiple(items map[string]interface{}, ttl time.Duration) error {
	// Use DefaultTTL if ttl is 0 and DefaultTTL is configured
	effectiveTTL := ttl
	if ttl == 0 && c.config.DefaultTTL > 0 {
		effectiveTTL = c.config.DefaultTTL
	}

	var expiresAt time.Time
	if effectiveTTL > 0 {
		expiresAt = time.Now().Add(effectiveTTL)
	}

	c.mu.Lock()
	for key, value := range items {
		c.cache[key] = &cacheEntry{
			value:     value,
			expiresAt: expiresAt,
		}
	}
	c.mu.Unlock()

	// Also set in distributed cache if available
	if c.distributed != nil {
		return c.distributed.SetMultiple(items, effectiveTTL)
	}

	return nil
}

// DeleteMultiple removes multiple values from cache
func (c *cacheManagerImpl) DeleteMultiple(keys []string) error {
	c.mu.Lock()
	for _, key := range keys {
		delete(c.cache, key)
	}
	c.mu.Unlock()

	// Also delete from distributed cache if available
	if c.distributed != nil {
		return c.distributed.DeleteMultiple(keys)
	}

	return nil
}

// Increment increments a numeric value in cache
func (c *cacheManagerImpl) Increment(key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		// Initialize with delta
		c.cache[key] = &cacheEntry{
			value: delta,
		}
		return delta, nil
	}

	if entry.isExpired() {
		// Reset with delta
		entry.value = delta
		entry.expiresAt = time.Time{}
		return delta, nil
	}

	// Try to increment existing value
	switch v := entry.value.(type) {
	case int64:
		newValue := v + delta
		entry.value = newValue
		return newValue, nil
	case int:
		newValue := int64(v) + delta
		entry.value = newValue
		return newValue, nil
	default:
		return 0, fmt.Errorf("cannot increment non-numeric value")
	}
}

// Decrement decrements a numeric value in cache
func (c *cacheManagerImpl) Decrement(key string, delta int64) (int64, error) {
	return c.Increment(key, -delta)
}

// Expire sets a new TTL for a cache entry
func (c *cacheManagerImpl) Expire(key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return ErrCacheKeyNotFound
	}

	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	} else {
		entry.expiresAt = time.Time{}
	}

	return nil
}

// TTL returns the time-to-live for a cache entry
func (c *cacheManagerImpl) TTL(key string) (time.Duration, error) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		return 0, ErrCacheKeyNotFound
	}

	if entry.expiresAt.IsZero() {
		return 0, nil // No expiration
	}

	ttl := time.Until(entry.expiresAt)
	if ttl < 0 {
		return 0, ErrCacheExpired
	}

	return ttl, nil
}

// GetRequestCache returns a request-specific cache
func (c *cacheManagerImpl) GetRequestCache(requestID string) RequestCache {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	cache, exists := c.requestCaches[requestID]
	if !exists {
		cache = newRequestCache(requestID)
		c.requestCaches[requestID] = cache
	}

	return cache
}

// ClearRequestCache removes a request-specific cache
func (c *cacheManagerImpl) ClearRequestCache(requestID string) error {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	if cache, exists := c.requestCaches[requestID]; exists {
		cache.Clear()
		delete(c.requestCaches, requestID)
	}

	return nil
}

// Invalidate removes all cache entries matching a pattern
func (c *cacheManagerImpl) Invalidate(pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple pattern matching (supports * wildcard)
	keysToDelete := []string{}
	for key := range c.cache {
		if matchPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.cache, key)
	}

	// Also invalidate in distributed cache if available
	if c.distributed != nil {
		return c.distributed.Invalidate(pattern)
	}

	return nil
}

// InvalidateTag removes all cache entries with a specific tag
func (c *cacheManagerImpl) InvalidateTag(tag string) error {
	c.tagMu.RLock()
	keys, exists := c.tagIndex[tag]
	c.tagMu.RUnlock()

	if !exists {
		return nil
	}

	// Delete all keys with this tag
	c.mu.Lock()
	for _, key := range keys {
		delete(c.cache, key)
	}
	c.mu.Unlock()

	// Remove tag from index
	c.tagMu.Lock()
	delete(c.tagIndex, tag)
	c.tagMu.Unlock()

	// Also invalidate in distributed cache if available
	if c.distributed != nil {
		return c.distributed.InvalidateTag(tag)
	}

	return nil
}

// SetWithTags stores a value with tags for invalidation
func (c *cacheManagerImpl) SetWithTags(key string, value interface{}, ttl time.Duration, tags []string) error {
	entry := &cacheEntry{
		value: value,
		tags:  tags,
	}

	// Use DefaultTTL if ttl is 0 and DefaultTTL is configured
	effectiveTTL := ttl
	if ttl == 0 && c.config.DefaultTTL > 0 {
		effectiveTTL = c.config.DefaultTTL
	}

	if effectiveTTL > 0 {
		entry.expiresAt = time.Now().Add(effectiveTTL)
	}

	c.mu.Lock()
	c.cache[key] = entry
	c.mu.Unlock()

	// Update tag index
	c.tagMu.Lock()
	for _, tag := range tags {
		c.tagIndex[tag] = append(c.tagIndex[tag], key)
	}
	c.tagMu.Unlock()

	return nil
}

// matchPattern performs simple wildcard pattern matching
func matchPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple implementation: only supports * at the end
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(str) >= len(prefix) && str[:len(prefix)] == prefix
	}

	return str == pattern
}

// CleanupExpired removes all expired entries from cache
func (c *cacheManagerImpl) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.cache {
		if entry.isExpired() {
			delete(c.cache, key)
			count++
		}
	}

	return count
}

// DistributedCache defines the interface for distributed caching backends
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
