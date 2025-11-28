package pkg

import (
	"context"
	"net/http/httptest"
	"testing"
)

// TestContextSetGet tests basic Set and Get functionality
func TestContextSetGet(t *testing.T) {
	req := &Request{
		Params: make(map[string]string),
	}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Test setting and getting a value
	ctx.Set("key1", "value1")
	val, ok := ctx.Get("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test getting non-existent key
	val, ok = ctx.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
	if val != nil {
		t.Errorf("Expected nil value for nonexistent key, got %v", val)
	}
}

// TestContextIsolation tests that contexts are isolated
func TestContextIsolation(t *testing.T) {
	req1 := &Request{Params: make(map[string]string)}
	req2 := &Request{Params: make(map[string]string)}

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	resp1 := NewResponseWriter(w1)
	resp2 := NewResponseWriter(w2)

	ctx1 := NewContext(req1, resp1, context.Background())
	ctx2 := NewContext(req2, resp2, context.Background())

	// Set value in ctx1
	ctx1.Set("key", "value1")

	// Verify it's in ctx1
	val, ok := ctx1.Get("key")
	if !ok || val != "value1" {
		t.Error("Expected to find key in ctx1")
	}

	// Verify it's NOT in ctx2
	val, ok = ctx2.Get("key")
	if ok {
		t.Errorf("Expected not to find key in ctx2, but found %v", val)
	}
}

// TestContextDerivedContexts tests that derived contexts share values
func TestContextDerivedContexts(t *testing.T) {
	req := &Request{Params: make(map[string]string)}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Set value in original context
	ctx.Set("key", "value")

	// Create derived contexts
	ctxWithTimeout := ctx.WithTimeout(1000000000) // 1 second
	ctxWithCancel, cancel := ctx.WithCancel()
	defer cancel()

	// Verify value is accessible in derived contexts
	val, ok := ctxWithTimeout.Get("key")
	if !ok || val != "value" {
		t.Error("Expected to find key in WithTimeout context")
	}

	val, ok = ctxWithCancel.Get("key")
	if !ok || val != "value" {
		t.Error("Expected to find key in WithCancel context")
	}

	// Set value in derived context
	ctxWithTimeout.Set("key2", "value2")

	// Verify it's accessible in original context (same request)
	val, ok = ctx.Get("key2")
	if !ok || val != "value2" {
		t.Error("Expected to find key2 in original context")
	}
}

// TestContextOverwrite tests overwriting values
func TestContextOverwrite(t *testing.T) {
	req := &Request{Params: make(map[string]string)}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Set initial value
	ctx.Set("key", "value1")

	// Verify initial value
	val, ok := ctx.Get("key")
	if !ok || val != "value1" {
		t.Error("Expected to find initial value")
	}

	// Overwrite
	ctx.Set("key", "value2")

	// Verify new value
	val, ok = ctx.Get("key")
	if !ok || val != "value2" {
		t.Error("Expected to find updated value")
	}
}

// TestContextMultipleKeys tests multiple keys
func TestContextMultipleKeys(t *testing.T) {
	req := &Request{Params: make(map[string]string)}
	w := httptest.NewRecorder()
	resp := NewResponseWriter(w)
	ctx := NewContext(req, resp, context.Background())

	// Set multiple values
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.Set("key3", true)

	// Verify all values
	val1, ok1 := ctx.Get("key1")
	if !ok1 || val1 != "value1" {
		t.Error("Expected to find key1")
	}

	val2, ok2 := ctx.Get("key2")
	if !ok2 || val2 != 42 {
		t.Error("Expected to find key2")
	}

	val3, ok3 := ctx.Get("key3")
	if !ok3 || val3 != true {
		t.Error("Expected to find key3")
	}
}
