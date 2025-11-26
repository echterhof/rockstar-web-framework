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
