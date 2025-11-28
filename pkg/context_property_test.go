package pkg

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ParamRetrievalConsistency tests Property 1:
// Param retrieval consistency
// **Feature: test-fixes, Property 1: Param retrieval consistency**
// **Validates: Requirements 1.1, 1.2**
func TestProperty_ParamRetrievalConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Param(name) returns same value as Params()[name]", prop.ForAll(
		func(params map[string]string, key string) bool {
			// Create a test request with the generated params
			req := &Request{
				Params: params,
			}

			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Get value using Param method
			paramValue := ctx.Param(key)

			// Get value using Params map
			mapValue := ctx.Params()[key]

			// They should be equal
			if paramValue != mapValue {
				t.Logf("Param(%s) returned '%s', but Params()[%s] is '%s'",
					key, paramValue, key, mapValue)
				return false
			}

			return true
		},
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
		gen.AlphaString(),
	))

	properties.Property("Param returns empty string for non-existent parameter", prop.ForAll(
		func(params map[string]string, nonExistentKey string) bool {
			// Ensure the key doesn't exist in params
			if _, exists := params[nonExistentKey]; exists {
				// Skip this test case if the key happens to exist
				return true
			}

			// Create a test request with the generated params
			req := &Request{
				Params: params,
			}

			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Get value for non-existent key
			paramValue := ctx.Param(nonExistentKey)

			// Should return empty string
			if paramValue != "" {
				t.Logf("Param(%s) should return empty string for non-existent key, got '%s'",
					nonExistentKey, paramValue)
				return false
			}

			return true
		},
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
		gen.Const("non_existent_key_12345"),
	))

	properties.Property("Param works with empty params map", prop.ForAll(
		func(key string) bool {
			// Create a test request with empty params
			req := &Request{
				Params: map[string]string{},
			}

			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Get value for any key
			paramValue := ctx.Param(key)

			// Should return empty string
			if paramValue != "" {
				t.Logf("Param(%s) should return empty string for empty params, got '%s'",
					key, paramValue)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("Param works with nil params map", prop.ForAll(
		func(key string) bool {
			// Create a test request with nil params
			req := &Request{
				Params: nil,
			}

			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Get value for any key
			paramValue := ctx.Param(key)

			// Should return empty string
			if paramValue != "" {
				t.Logf("Param(%s) should return empty string for nil params, got '%s'",
					key, paramValue)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// TestProperty_ContextValueIsolation tests Property 13:
// Context Value Isolation
// **Feature: compile-time-plugins, Property 13: Context Value Isolation**
// **Validates: Requirements 10.1, 10.4, 10.6**
func TestProperty_ContextValueIsolation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("values set in one request context are not visible in another", prop.ForAll(
		func(key1, key2 string, value1, value2 int) bool {
			// Skip if keys are empty or identical
			if key1 == "" || key2 == "" || key1 == key2 {
				return true
			}

			// Create two separate request contexts (simulating two different requests)
			req1 := &Request{
				Params: make(map[string]string),
			}
			req2 := &Request{
				Params: make(map[string]string),
			}

			w1 := httptest.NewRecorder()
			w2 := httptest.NewRecorder()
			resp1 := NewResponseWriter(w1)
			resp2 := NewResponseWriter(w2)

			ctx1 := NewContext(req1, resp1, context.Background())
			ctx2 := NewContext(req2, resp2, context.Background())

			// Set values in context 1
			ctx1.Set(key1, value1)
			ctx1.Set(key2, value2)

			// Verify values are present in context 1
			val1, ok1 := ctx1.Get(key1)
			if !ok1 || val1 != value1 {
				t.Logf("Context 1: expected to find key '%s' with value %d, got ok=%v, value=%v",
					key1, value1, ok1, val1)
				return false
			}

			val2, ok2 := ctx1.Get(key2)
			if !ok2 || val2 != value2 {
				t.Logf("Context 1: expected to find key '%s' with value %d, got ok=%v, value=%v",
					key2, value2, ok2, val2)
				return false
			}

			// Verify values are NOT present in context 2 (isolation)
			val1_ctx2, ok1_ctx2 := ctx2.Get(key1)
			if ok1_ctx2 {
				t.Logf("Context 2: should not find key '%s' from context 1, but found value=%v",
					key1, val1_ctx2)
				return false
			}

			val2_ctx2, ok2_ctx2 := ctx2.Get(key2)
			if ok2_ctx2 {
				t.Logf("Context 2: should not find key '%s' from context 1, but found value=%v",
					key2, val2_ctx2)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("multiple plugins can set different keys without conflicts", prop.ForAll(
		func(keys []string, values []int) bool {
			// Ensure we have at least 2 keys
			if len(keys) < 2 || len(values) < 2 {
				return true // Skip
			}

			// Create a single request context
			req := &Request{
				Params: make(map[string]string),
			}
			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Simulate multiple plugins setting values
			keyValueMap := make(map[string]int)
			for i := 0; i < len(keys) && i < len(values); i++ {
				ctx.Set(keys[i], values[i])
				keyValueMap[keys[i]] = values[i]
			}

			// Verify all values are retrievable
			for key, expectedValue := range keyValueMap {
				val, ok := ctx.Get(key)
				if !ok {
					t.Logf("Expected to find key '%s', but it was not found", key)
					return false
				}
				if val != expectedValue {
					t.Logf("Expected key '%s' to have value %d, got %v",
						key, expectedValue, val)
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.AlphaString()),
		gen.SliceOf(gen.Int()),
	))

	properties.Property("values are accessible within the same request across derived contexts", prop.ForAll(
		func(key string, value int) bool {
			// Create a request context
			req := &Request{
				Params: make(map[string]string),
			}
			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Set a value in the original context
			ctx.Set(key, value)

			// Create derived contexts (same request)
			ctxWithTimeout := ctx.WithTimeout(1000000000) // 1 second
			ctxWithCancel, cancel := ctx.WithCancel()
			defer cancel()

			// Verify value is accessible in derived contexts
			val1, ok1 := ctxWithTimeout.Get(key)
			if !ok1 || val1 != value {
				t.Logf("WithTimeout context: expected to find key '%s' with value %d, got ok=%v, value=%v",
					key, value, ok1, val1)
				return false
			}

			val2, ok2 := ctxWithCancel.Get(key)
			if !ok2 || val2 != value {
				t.Logf("WithCancel context: expected to find key '%s' with value %d, got ok=%v, value=%v",
					key, value, ok2, val2)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.Int(),
	))

	properties.Property("Get returns false for non-existent keys", prop.ForAll(
		func(existingKeys []string, nonExistentKey string) bool {
			// Create a request context
			req := &Request{
				Params: make(map[string]string),
			}
			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Set some existing keys
			for i, key := range existingKeys {
				ctx.Set(key, i)
			}

			// Ensure nonExistentKey is not in existingKeys
			for _, key := range existingKeys {
				if key == nonExistentKey {
					return true // Skip this test case
				}
			}

			// Try to get a non-existent key
			val, ok := ctx.Get(nonExistentKey)
			if ok {
				t.Logf("Expected Get('%s') to return false, but got true with value=%v",
					nonExistentKey, val)
				return false
			}

			if val != nil {
				t.Logf("Expected Get('%s') to return nil value, but got %v",
					nonExistentKey, val)
				return false
			}

			return true
		},
		gen.SliceOf(gen.AlphaString()),
		gen.Const("non_existent_key_xyz_12345"),
	))

	properties.Property("overwriting a key updates the value", prop.ForAll(
		func(key string, value1, value2 int) bool {
			// Ensure values are different
			if value1 == value2 {
				return true // Skip
			}

			// Create a request context
			req := &Request{
				Params: make(map[string]string),
			}
			w := httptest.NewRecorder()
			resp := NewResponseWriter(w)
			ctx := NewContext(req, resp, context.Background())

			// Set initial value
			ctx.Set(key, value1)

			// Verify initial value
			val, ok := ctx.Get(key)
			if !ok || val != value1 {
				t.Logf("Expected initial value %d for key '%s', got ok=%v, value=%v",
					value1, key, ok, val)
				return false
			}

			// Overwrite with new value
			ctx.Set(key, value2)

			// Verify new value
			val, ok = ctx.Get(key)
			if !ok || val != value2 {
				t.Logf("Expected updated value %d for key '%s', got ok=%v, value=%v",
					value2, key, ok, val)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.Int(),
		gen.Int(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
