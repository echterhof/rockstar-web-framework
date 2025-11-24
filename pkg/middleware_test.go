package pkg

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestMiddlewareEngineRegister tests middleware registration
func TestMiddlewareEngineRegister(t *testing.T) {
	engine := NewMiddlewareEngine()

	// Test successful registration
	config := MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Priority: 10,
		Enabled:  true,
	}

	err := engine.Register(config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = engine.Register(config)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}

	// Test registration with empty name
	invalidConfig := MiddlewareConfig{
		Name:    "",
		Handler: func(ctx Context, next HandlerFunc) error { return next(ctx) },
	}
	err = engine.Register(invalidConfig)
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	// Test registration with nil handler
	invalidConfig = MiddlewareConfig{
		Name:    "invalid",
		Handler: nil,
	}
	err = engine.Register(invalidConfig)
	if err == nil {
		t.Error("Expected error for nil handler, got nil")
	}
}

// TestMiddlewareEngineUnregister tests middleware unregistration
func TestMiddlewareEngineUnregister(t *testing.T) {
	engine := NewMiddlewareEngine()

	config := MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Enabled:  true,
	}

	engine.Register(config)

	// Test successful unregistration
	err := engine.Unregister("test-middleware")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test unregistering non-existent middleware
	err = engine.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent middleware, got nil")
	}
}

// TestMiddlewareEngineEnableDisable tests enabling and disabling middleware
func TestMiddlewareEngineEnableDisable(t *testing.T) {
	engine := NewMiddlewareEngine()

	config := MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Enabled:  true,
	}

	engine.Register(config)

	// Test disable
	err := engine.Disable("test-middleware")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test enable
	err = engine.Enable("test-middleware")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test enable non-existent middleware
	err = engine.Enable("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent middleware, got nil")
	}

	// Test disable non-existent middleware
	err = engine.Disable("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent middleware, got nil")
	}
}

// TestMiddlewareEngineSetPriority tests changing middleware priority
func TestMiddlewareEngineSetPriority(t *testing.T) {
	engine := NewMiddlewareEngine()

	config := MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Priority: 10,
		Enabled:  true,
	}

	engine.Register(config)

	// Test setting priority
	err := engine.SetPriority("test-middleware", 20)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify priority was changed
	list := engine.List()
	found := false
	for _, mw := range list {
		if mw.Name == "test-middleware" && mw.Priority == 20 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Priority was not updated correctly")
	}

	// Test setting priority for non-existent middleware
	err = engine.SetPriority("non-existent", 30)
	if err == nil {
		t.Error("Expected error for non-existent middleware, got nil")
	}
}

// TestMiddlewareEngineSetPosition tests changing middleware position
func TestMiddlewareEngineSetPosition(t *testing.T) {
	engine := NewMiddlewareEngine()

	config := MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Enabled:  true,
	}

	engine.Register(config)

	// Test setting position
	err := engine.SetPosition("test-middleware", MiddlewarePositionAfter)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify position was changed
	list := engine.List()
	found := false
	for _, mw := range list {
		if mw.Name == "test-middleware" && mw.Position == MiddlewarePositionAfter {
			found = true
			break
		}
	}
	if !found {
		t.Error("Position was not updated correctly")
	}

	// Test setting position for non-existent middleware
	err = engine.SetPosition("non-existent", MiddlewarePositionBefore)
	if err == nil {
		t.Error("Expected error for non-existent middleware, got nil")
	}
}

// TestMiddlewareEngineExecuteOrder tests middleware execution order
func TestMiddlewareEngineExecuteOrder(t *testing.T) {
	engine := NewMiddlewareEngine()

	// Track execution order
	var executionOrder []string

	// Register before middleware with different priorities
	engine.Register(MiddlewareConfig{
		Name:     "before-high",
		Position: MiddlewarePositionBefore,
		Priority: 100,
		Enabled:  true,
		Handler: func(ctx Context, next HandlerFunc) error {
			executionOrder = append(executionOrder, "before-high-start")
			err := next(ctx)
			executionOrder = append(executionOrder, "before-high-end")
			return err
		},
	})

	engine.Register(MiddlewareConfig{
		Name:     "before-low",
		Position: MiddlewarePositionBefore,
		Priority: 10,
		Enabled:  true,
		Handler: func(ctx Context, next HandlerFunc) error {
			executionOrder = append(executionOrder, "before-low-start")
			err := next(ctx)
			executionOrder = append(executionOrder, "before-low-end")
			return err
		},
	})

	// Register after middleware with different priorities
	engine.Register(MiddlewareConfig{
		Name:     "after-high",
		Position: MiddlewarePositionAfter,
		Priority: 100,
		Enabled:  true,
		Handler: func(ctx Context, next HandlerFunc) error {
			executionOrder = append(executionOrder, "after-high-start")
			err := next(ctx)
			executionOrder = append(executionOrder, "after-high-end")
			return err
		},
	})

	engine.Register(MiddlewareConfig{
		Name:     "after-low",
		Position: MiddlewarePositionAfter,
		Priority: 10,
		Enabled:  true,
		Handler: func(ctx Context, next HandlerFunc) error {
			executionOrder = append(executionOrder, "after-low-start")
			err := next(ctx)
			executionOrder = append(executionOrder, "after-low-end")
			return err
		},
	})

	// Create a test handler
	handler := func(ctx Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	// Execute
	ctx := &mockMiddlewareContext{}
	err := engine.Execute(ctx, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify execution order
	// Expected: before-high-start, before-low-start, after-low-start, after-high-start, handler, after-high-end, after-low-end, before-low-end, before-high-end
	expectedOrder := []string{
		"before-high-start",
		"before-low-start",
		"after-low-start",
		"after-high-start",
		"handler",
		"after-high-end",
		"after-low-end",
		"before-low-end",
		"before-high-end",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) {
			t.Errorf("Missing execution step at index %d: expected %s", i, expected)
			continue
		}
		if executionOrder[i] != expected {
			t.Errorf("At index %d: expected %s, got %s", i, expected, executionOrder[i])
		}
	}
}

// TestMiddlewareEngineExecuteDisabled tests that disabled middleware is not executed
func TestMiddlewareEngineExecuteDisabled(t *testing.T) {
	engine := NewMiddlewareEngine()

	var executed bool

	engine.Register(MiddlewareConfig{
		Name:     "disabled-middleware",
		Position: MiddlewarePositionBefore,
		Enabled:  false,
		Handler: func(ctx Context, next HandlerFunc) error {
			executed = true
			return next(ctx)
		},
	})

	handler := func(ctx Context) error {
		return nil
	}

	ctx := &mockMiddlewareContext{}
	err := engine.Execute(ctx, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if executed {
		t.Error("Disabled middleware should not be executed")
	}
}

// TestMiddlewareEngineExecuteError tests error propagation
func TestMiddlewareEngineExecuteError(t *testing.T) {
	engine := NewMiddlewareEngine()

	expectedError := errors.New("middleware error")

	engine.Register(MiddlewareConfig{
		Name:     "error-middleware",
		Position: MiddlewarePositionBefore,
		Enabled:  true,
		Handler: func(ctx Context, next HandlerFunc) error {
			return expectedError
		},
	})

	handler := func(ctx Context) error {
		return nil
	}

	ctx := &mockMiddlewareContext{}
	err := engine.Execute(ctx, handler)
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

// TestMiddlewareEngineList tests listing middleware
func TestMiddlewareEngineList(t *testing.T) {
	engine := NewMiddlewareEngine()

	// Register multiple middleware
	for i := 0; i < 3; i++ {
		engine.Register(MiddlewareConfig{
			Name:     fmt.Sprintf("middleware-%d", i),
			Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
			Position: MiddlewarePositionBefore,
			Enabled:  true,
		})
	}

	list := engine.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 middleware, got %d", len(list))
	}
}

// TestMiddlewareEngineClear tests clearing all middleware
func TestMiddlewareEngineClear(t *testing.T) {
	engine := NewMiddlewareEngine()

	// Register middleware
	engine.Register(MiddlewareConfig{
		Name:     "test-middleware",
		Handler:  func(ctx Context, next HandlerFunc) error { return next(ctx) },
		Position: MiddlewarePositionBefore,
		Enabled:  true,
	})

	// Clear
	engine.Clear()

	list := engine.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 middleware after clear, got %d", len(list))
	}
}

// TestChainMiddleware tests chaining multiple middleware
func TestChainMiddleware(t *testing.T) {
	var executionOrder []string

	mw1 := func(ctx Context, next HandlerFunc) error {
		executionOrder = append(executionOrder, "mw1-start")
		err := next(ctx)
		executionOrder = append(executionOrder, "mw1-end")
		return err
	}

	mw2 := func(ctx Context, next HandlerFunc) error {
		executionOrder = append(executionOrder, "mw2-start")
		err := next(ctx)
		executionOrder = append(executionOrder, "mw2-end")
		return err
	}

	chained := ChainMiddleware(mw1, mw2)

	handler := func(ctx Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	ctx := &mockMiddlewareContext{}
	err := chained(ctx, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedOrder := []string{"mw1-start", "mw2-start", "handler", "mw2-end", "mw1-end"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("At index %d: expected %s, got %s", i, expected, executionOrder[i])
		}
	}
}

// TestSkipMiddleware tests conditional middleware execution
func TestSkipMiddleware(t *testing.T) {
	var executed bool

	mw := func(ctx Context, next HandlerFunc) error {
		executed = true
		return next(ctx)
	}

	// Test with skip condition true
	skipMw := SkipMiddleware(func(ctx Context) bool { return true }, mw)
	handler := func(ctx Context) error { return nil }

	ctx := &mockMiddlewareContext{}
	err := skipMw(ctx, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if executed {
		t.Error("Middleware should have been skipped")
	}

	// Test with skip condition false
	executed = false
	skipMw = SkipMiddleware(func(ctx Context) bool { return false }, mw)
	err = skipMw(ctx, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !executed {
		t.Error("Middleware should have been executed")
	}
}

// TestRecoverMiddleware tests panic recovery
func TestRecoverMiddleware(t *testing.T) {
	var recovered interface{}

	recoverMw := RecoverMiddleware(func(ctx Context, r interface{}) error {
		recovered = r
		return nil
	})

	handler := func(ctx Context) error {
		panic("test panic")
	}

	ctx := &mockMiddlewareContext{}
	recoverMw(ctx, handler)

	if recovered == nil {
		t.Error("Expected panic to be recovered")
	}

	if recovered != "test panic" {
		t.Errorf("Expected recovered value 'test panic', got %v", recovered)
	}
}

// mockMiddlewareContext is a minimal Context implementation for testing
type mockMiddlewareContext struct{}

func (m *mockMiddlewareContext) Request() *Request                           { return nil }
func (m *mockMiddlewareContext) Response() ResponseWriter                    { return nil }
func (m *mockMiddlewareContext) Params() map[string]string                   { return nil }
func (m *mockMiddlewareContext) Query() map[string]string                    { return nil }
func (m *mockMiddlewareContext) Headers() map[string]string                  { return nil }
func (m *mockMiddlewareContext) Body() []byte                                { return nil }
func (m *mockMiddlewareContext) Session() SessionManager                     { return nil }
func (m *mockMiddlewareContext) User() *User                                 { return nil }
func (m *mockMiddlewareContext) Tenant() *Tenant                             { return nil }
func (m *mockMiddlewareContext) DB() DatabaseManager                         { return nil }
func (m *mockMiddlewareContext) Cache() CacheManager                         { return nil }
func (m *mockMiddlewareContext) Config() ConfigManager                       { return nil }
func (m *mockMiddlewareContext) I18n() I18nManager                           { return nil }
func (m *mockMiddlewareContext) Files() FileManager                          { return nil }
func (m *mockMiddlewareContext) Logger() Logger                              { return nil }
func (m *mockMiddlewareContext) Metrics() MetricsCollector                   { return nil }
func (m *mockMiddlewareContext) Context() context.Context                    { return nil }
func (m *mockMiddlewareContext) WithTimeout(timeout time.Duration) Context   { return m }
func (m *mockMiddlewareContext) WithCancel() (Context, context.CancelFunc)   { return m, func() {} }
func (m *mockMiddlewareContext) JSON(statusCode int, data interface{}) error { return nil }
func (m *mockMiddlewareContext) XML(statusCode int, data interface{}) error  { return nil }
func (m *mockMiddlewareContext) HTML(statusCode int, template string, data interface{}) error {
	return nil
}
func (m *mockMiddlewareContext) String(statusCode int, message string) error { return nil }
func (m *mockMiddlewareContext) Redirect(statusCode int, url string) error   { return nil }
func (m *mockMiddlewareContext) SetCookie(cookie *Cookie) error              { return nil }
func (m *mockMiddlewareContext) GetCookie(name string) (*Cookie, error)      { return nil, nil }
func (m *mockMiddlewareContext) SetHeader(key, value string)                 {}
func (m *mockMiddlewareContext) GetHeader(key string) string                 { return "" }
func (m *mockMiddlewareContext) FormValue(key string) string                 { return "" }
func (m *mockMiddlewareContext) FormFile(key string) (*FormFile, error)      { return nil, nil }
func (m *mockMiddlewareContext) IsAuthenticated() bool                       { return false }
func (m *mockMiddlewareContext) IsAuthorized(resource, action string) bool   { return false }
