package pkg

import (
	"sync"
)

// requestCacheImpl implements RequestCache interface
// This provides request-specific caching with arena-like memory management
type requestCacheImpl struct {
	requestID string
	data      map[string]interface{}
	mu        sync.RWMutex
	size      int64
}

// newRequestCache creates a new request-specific cache
func newRequestCache(requestID string) *requestCacheImpl {
	return &requestCacheImpl{
		requestID: requestID,
		data:      make(map[string]interface{}),
		size:      0,
	}
}

// Get retrieves a value from the request cache
func (r *requestCacheImpl) Get(key string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.data[key]
}

// Set stores a value in the request cache
func (r *requestCacheImpl) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Estimate size increase (simplified)
	if _, exists := r.data[key]; !exists {
		r.size += int64(len(key)) + estimateSize(value)
	}

	r.data[key] = value
}

// Delete removes a value from the request cache
func (r *requestCacheImpl) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if value, exists := r.data[key]; exists {
		r.size -= int64(len(key)) + estimateSize(value)
		delete(r.data, key)
	}
}

// Clear removes all values from the request cache
func (r *requestCacheImpl) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = make(map[string]interface{})
	r.size = 0
}

// Size returns the estimated size of the cache in bytes
func (r *requestCacheImpl) Size() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.size
}

// Keys returns all keys in the cache
func (r *requestCacheImpl) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.data))
	for key := range r.data {
		keys = append(keys, key)
	}

	return keys
}

// estimateSize provides a rough estimate of value size in bytes
func estimateSize(value interface{}) int64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int8, int16, int32, int64:
		return 8
	case uint, uint8, uint16, uint32, uint64:
		return 8
	case float32, float64:
		return 8
	case bool:
		return 1
	case map[string]interface{}:
		size := int64(0)
		for k, val := range v {
			size += int64(len(k)) + estimateSize(val)
		}
		return size
	case []interface{}:
		size := int64(0)
		for _, val := range v {
			size += estimateSize(val)
		}
		return size
	default:
		// Default estimate for unknown types
		return 64
	}
}
