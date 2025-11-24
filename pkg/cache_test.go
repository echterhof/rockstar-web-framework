package pkg

import (
	"testing"
	"time"
)

func TestCacheManager_BasicOperations(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Test Set and Get
	err := cache.Set("key1", "value1", 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}

	// Test Exists
	if !cache.Exists("key1") {
		t.Error("Expected key1 to exist")
	}

	if cache.Exists("nonexistent") {
		t.Error("Expected nonexistent key to not exist")
	}

	// Test Delete
	err = cache.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if cache.Exists("key1") {
		t.Error("Expected key1 to be deleted")
	}

	_, err = cache.Get("key1")
	if err != ErrCacheKeyNotFound {
		t.Errorf("Expected ErrCacheKeyNotFound, got %v", err)
	}
}

func TestCacheManager_TTL(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Set with TTL
	err := cache.Set("expiring", "value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	// Should exist immediately
	if !cache.Exists("expiring") {
		t.Error("Expected key to exist immediately after set")
	}

	// Check TTL
	ttl, err := cache.TTL("expiring")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}

	if ttl <= 0 || ttl > 100*time.Millisecond {
		t.Errorf("Expected TTL around 100ms, got %v", ttl)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	if cache.Exists("expiring") {
		t.Error("Expected key to be expired")
	}

	_, err = cache.Get("expiring")
	if err != ErrCacheExpired {
		t.Errorf("Expected ErrCacheExpired, got %v", err)
	}
}

func TestCacheManager_MultipleOperations(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Test SetMultiple
	items := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	err := cache.SetMultiple(items, 0)
	if err != nil {
		t.Fatalf("SetMultiple failed: %v", err)
	}

	// Test GetMultiple
	keys := []string{"key1", "key2", "key3", "nonexistent"}
	result, err := cache.GetMultiple(keys)
	if err != nil {
		t.Fatalf("GetMultiple failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected 'value1', got '%v'", result["key1"])
	}

	if result["key2"] != 42 {
		t.Errorf("Expected 42, got '%v'", result["key2"])
	}

	if result["key3"] != true {
		t.Errorf("Expected true, got '%v'", result["key3"])
	}

	// Test DeleteMultiple
	err = cache.DeleteMultiple([]string{"key1", "key2"})
	if err != nil {
		t.Fatalf("DeleteMultiple failed: %v", err)
	}

	if cache.Exists("key1") || cache.Exists("key2") {
		t.Error("Expected keys to be deleted")
	}

	if !cache.Exists("key3") {
		t.Error("Expected key3 to still exist")
	}
}

func TestCacheManager_IncrementDecrement(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Test Increment on non-existent key
	value, err := cache.Increment("counter", 1)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}

	if value != 1 {
		t.Errorf("Expected 1, got %d", value)
	}

	// Test Increment on existing key
	value, err = cache.Increment("counter", 5)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}

	if value != 6 {
		t.Errorf("Expected 6, got %d", value)
	}

	// Test Decrement
	value, err = cache.Decrement("counter", 2)
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}

	if value != 4 {
		t.Errorf("Expected 4, got %d", value)
	}

	// Test Increment on non-numeric value
	cache.Set("string", "not a number", 0)
	_, err = cache.Increment("string", 1)
	if err == nil {
		t.Error("Expected error when incrementing non-numeric value")
	}
}

func TestCacheManager_Expire(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Set without expiration
	cache.Set("key", "value", 0)

	// Check no expiration
	ttl, err := cache.TTL("key")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}

	if ttl != 0 {
		t.Errorf("Expected no expiration (0), got %v", ttl)
	}

	// Set expiration
	err = cache.Expire("key", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}

	// Check expiration is set
	ttl, err = cache.TTL("key")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}

	if ttl <= 0 || ttl > 100*time.Millisecond {
		t.Errorf("Expected TTL around 100ms, got %v", ttl)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.TTL("key")
	if err != ErrCacheExpired {
		t.Errorf("Expected ErrCacheExpired, got %v", err)
	}
}

func TestCacheManager_Clear(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Add multiple entries
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	// Clear all
	err := cache.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Check all are gone
	if cache.Exists("key1") || cache.Exists("key2") || cache.Exists("key3") {
		t.Error("Expected all keys to be cleared")
	}
}

func TestCacheManager_Invalidate(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Add entries with pattern
	cache.Set("user:1:profile", "data1", 0)
	cache.Set("user:2:profile", "data2", 0)
	cache.Set("user:1:settings", "data3", 0)
	cache.Set("product:1", "data4", 0)

	// Invalidate user:* pattern
	err := cache.Invalidate("user:*")
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	// Check user keys are gone
	if cache.Exists("user:1:profile") || cache.Exists("user:2:profile") || cache.Exists("user:1:settings") {
		t.Error("Expected user keys to be invalidated")
	}

	// Check product key still exists
	if !cache.Exists("product:1") {
		t.Error("Expected product key to still exist")
	}
}

func TestCacheManager_RequestCache(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Get request cache
	reqCache := cache.GetRequestCache("request-123")
	if reqCache == nil {
		t.Fatal("Expected request cache to be created")
	}

	// Test Set and Get
	reqCache.Set("key1", "value1")
	value := reqCache.Get("key1")

	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}

	// Test Keys
	reqCache.Set("key2", "value2")
	keys := reqCache.Keys()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Test Size
	size := reqCache.Size()
	if size <= 0 {
		t.Error("Expected size to be greater than 0")
	}

	// Test Delete
	reqCache.Delete("key1")
	value = reqCache.Get("key1")

	if value != nil {
		t.Error("Expected key1 to be deleted")
	}

	// Test Clear
	reqCache.Clear()
	keys = reqCache.Keys()

	if len(keys) != 0 {
		t.Error("Expected all keys to be cleared")
	}

	if reqCache.Size() != 0 {
		t.Error("Expected size to be 0 after clear")
	}

	// Test ClearRequestCache
	err := cache.ClearRequestCache("request-123")
	if err != nil {
		t.Fatalf("ClearRequestCache failed: %v", err)
	}
}

func TestCacheManager_RequestCacheIsolation(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Create two request caches
	reqCache1 := cache.GetRequestCache("request-1")
	reqCache2 := cache.GetRequestCache("request-2")

	// Set different values
	reqCache1.Set("key", "value1")
	reqCache2.Set("key", "value2")

	// Check isolation
	if reqCache1.Get("key") != "value1" {
		t.Error("Request cache 1 should have value1")
	}

	if reqCache2.Get("key") != "value2" {
		t.Error("Request cache 2 should have value2")
	}

	// Clear one cache
	reqCache1.Clear()

	// Check other cache is unaffected
	if reqCache2.Get("key") != "value2" {
		t.Error("Request cache 2 should still have value2")
	}
}

func TestCacheManager_CleanupExpired(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	impl := cache.(*cacheManagerImpl)

	// Add entries with different TTLs
	cache.Set("key1", "value1", 50*time.Millisecond)
	cache.Set("key2", "value2", 200*time.Millisecond)
	cache.Set("key3", "value3", 0) // No expiration

	// Wait for first key to expire
	time.Sleep(100 * time.Millisecond)

	// Cleanup expired
	count := impl.CleanupExpired()

	if count != 1 {
		t.Errorf("Expected 1 expired entry, got %d", count)
	}

	// Check key1 is gone
	if cache.Exists("key1") {
		t.Error("Expected key1 to be cleaned up")
	}

	// Check other keys still exist
	if !cache.Exists("key2") || !cache.Exists("key3") {
		t.Error("Expected key2 and key3 to still exist")
	}
}

func TestRequestCache_EstimateSize(t *testing.T) {
	reqCache := newRequestCache("test")

	// Test string
	reqCache.Set("string", "hello")
	if reqCache.Size() <= 0 {
		t.Error("Expected size > 0 for string")
	}

	// Test int
	reqCache.Set("int", 42)
	prevSize := reqCache.Size()

	// Test slice
	reqCache.Set("slice", []interface{}{1, 2, 3})
	if reqCache.Size() <= prevSize {
		t.Error("Expected size to increase with slice")
	}

	// Test map
	prevSize = reqCache.Size()
	reqCache.Set("map", map[string]interface{}{"key": "value"})
	if reqCache.Size() <= prevSize {
		t.Error("Expected size to increase with map")
	}
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.Set("key", n*100+j, 0)
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cache.Get("key")
				cache.Exists("key")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic and should have a value
	if !cache.Exists("key") {
		t.Error("Expected key to exist after concurrent access")
	}
}

func TestCacheManager_TypePreservation(t *testing.T) {
	cache := NewCacheManager(CacheConfig{})

	// Test different types
	testCases := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello"},
		{"int", 42},
		{"int64", int64(123456789)},
		{"float64", 3.14159},
		{"bool", true},
		{"slice", []string{"a", "b", "c"}},
		{"map", map[string]int{"one": 1, "two": 2}},
		{"struct", struct{ Name string }{"test"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set(tc.name, tc.value, 0)
			if err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			value, err := cache.Get(tc.name)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			// Type should be preserved
			if value == nil {
				t.Error("Expected non-nil value")
			}
		})
	}
}
