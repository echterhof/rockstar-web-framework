package pkg

import (
	"fmt"
	"sync"
)

// MiddlewarePosition defines when middleware should execute
type MiddlewarePosition int

const (
	// MiddlewarePositionBefore executes before the handler
	MiddlewarePositionBefore MiddlewarePosition = iota
	// MiddlewarePositionAfter executes after the handler
	MiddlewarePositionAfter
)

// MiddlewareConfig holds configuration for a middleware
type MiddlewareConfig struct {
	// Name is the middleware identifier
	Name string

	// Handler is the middleware function
	Handler MiddlewareFunc

	// Position determines when the middleware executes (before or after handler)
	Position MiddlewarePosition

	// Priority determines execution order (higher priority executes first for before, last for after)
	Priority int

	// Enabled determines if the middleware is active
	Enabled bool
}

// MiddlewareEngine manages middleware execution with configurable ordering
type MiddlewareEngine interface {
	// Register adds a middleware with configuration
	Register(config MiddlewareConfig) error

	// Unregister removes a middleware by name
	Unregister(name string) error

	// Enable enables a middleware by name
	Enable(name string) error

	// Disable disables a middleware by name
	Disable(name string) error

	// SetPriority changes the priority of a middleware
	SetPriority(name string, priority int) error

	// SetPosition changes the position of a middleware
	SetPosition(name string, position MiddlewarePosition) error

	// Execute runs all middleware and the handler
	Execute(ctx Context, handler HandlerFunc) error

	// List returns all registered middleware configurations
	List() []MiddlewareConfig

	// Clear removes all middleware
	Clear()
}

// middlewareEngine implements MiddlewareEngine
type middlewareEngine struct {
	middleware map[string]*MiddlewareConfig
	mu         sync.RWMutex
}

// NewMiddlewareEngine creates a new middleware engine
func NewMiddlewareEngine() MiddlewareEngine {
	return &middlewareEngine{
		middleware: make(map[string]*MiddlewareConfig),
	}
}

// Register adds a middleware with configuration
func (m *middlewareEngine) Register(config MiddlewareConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config.Name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	if config.Handler == nil {
		return fmt.Errorf("middleware handler cannot be nil")
	}

	// Check if middleware already exists
	if _, exists := m.middleware[config.Name]; exists {
		return fmt.Errorf("middleware %s already registered", config.Name)
	}

	// Store a copy of the config
	configCopy := config
	m.middleware[config.Name] = &configCopy

	return nil
}

// Unregister removes a middleware by name
func (m *middlewareEngine) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.middleware[name]; !exists {
		return fmt.Errorf("middleware %s not found", name)
	}

	delete(m.middleware, name)
	return nil
}

// Enable enables a middleware by name
func (m *middlewareEngine) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mw, exists := m.middleware[name]
	if !exists {
		return fmt.Errorf("middleware %s not found", name)
	}

	mw.Enabled = true
	return nil
}

// Disable disables a middleware by name
func (m *middlewareEngine) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mw, exists := m.middleware[name]
	if !exists {
		return fmt.Errorf("middleware %s not found", name)
	}

	mw.Enabled = false
	return nil
}

// SetPriority changes the priority of a middleware
func (m *middlewareEngine) SetPriority(name string, priority int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mw, exists := m.middleware[name]
	if !exists {
		return fmt.Errorf("middleware %s not found", name)
	}

	mw.Priority = priority
	return nil
}

// SetPosition changes the position of a middleware
func (m *middlewareEngine) SetPosition(name string, position MiddlewarePosition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mw, exists := m.middleware[name]
	if !exists {
		return fmt.Errorf("middleware %s not found", name)
	}

	mw.Position = position
	return nil
}

// Execute runs all middleware and the handler
func (m *middlewareEngine) Execute(ctx Context, handler HandlerFunc) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Separate middleware by position
	var beforeMiddleware []*MiddlewareConfig
	var afterMiddleware []*MiddlewareConfig

	for _, mw := range m.middleware {
		if !mw.Enabled {
			continue
		}

		if mw.Position == MiddlewarePositionBefore {
			beforeMiddleware = append(beforeMiddleware, mw)
		} else {
			afterMiddleware = append(afterMiddleware, mw)
		}
	}

	// Sort before middleware by priority (higher priority first)
	sortMiddlewareByPriority(beforeMiddleware, true)

	// Sort after middleware by priority (higher priority last)
	sortMiddlewareByPriority(afterMiddleware, false)

	// Build the execution chain
	// Start with the handler
	finalHandler := handler

	// Wrap with after middleware (in reverse order so higher priority executes last)
	for i := len(afterMiddleware) - 1; i >= 0; i-- {
		mw := afterMiddleware[i]
		next := finalHandler
		finalHandler = func(ctx Context) error {
			return mw.Handler(ctx, next)
		}
	}

	// Wrap with before middleware (in order so higher priority executes first)
	for i := len(beforeMiddleware) - 1; i >= 0; i-- {
		mw := beforeMiddleware[i]
		next := finalHandler
		finalHandler = func(ctx Context) error {
			return mw.Handler(ctx, next)
		}
	}

	// Execute the complete chain
	return finalHandler(ctx)
}

// List returns all registered middleware configurations
func (m *middlewareEngine) List() []MiddlewareConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]MiddlewareConfig, 0, len(m.middleware))
	for _, mw := range m.middleware {
		configs = append(configs, *mw)
	}

	return configs
}

// Clear removes all middleware
func (m *middlewareEngine) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middleware = make(map[string]*MiddlewareConfig)
}

// sortMiddlewareByPriority sorts middleware by priority
// If descending is true, higher priority comes first
// If descending is false, lower priority comes first
func sortMiddlewareByPriority(middleware []*MiddlewareConfig, descending bool) {
	// Simple bubble sort - sufficient for small middleware lists
	n := len(middleware)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			shouldSwap := false
			if descending {
				shouldSwap = middleware[j].Priority < middleware[j+1].Priority
			} else {
				shouldSwap = middleware[j].Priority > middleware[j+1].Priority
			}

			if shouldSwap {
				middleware[j], middleware[j+1] = middleware[j+1], middleware[j]
			}
		}
	}
}

// Common middleware helpers

// ChainMiddleware chains multiple middleware functions into a single middleware
func ChainMiddleware(middleware ...MiddlewareFunc) MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		// Build chain from the end
		handler := next
		for i := len(middleware) - 1; i >= 0; i-- {
			mw := middleware[i]
			currentHandler := handler
			handler = func(ctx Context) error {
				return mw(ctx, currentHandler)
			}
		}
		return handler(ctx)
	}
}

// SkipMiddleware creates a middleware that conditionally skips execution
func SkipMiddleware(condition func(ctx Context) bool, middleware MiddlewareFunc) MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		if condition(ctx) {
			return next(ctx)
		}
		return middleware(ctx, next)
	}
}

// RecoverMiddleware creates a middleware that recovers from panics
func RecoverMiddleware(handler func(ctx Context, recovered interface{}) error) MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		defer func() {
			if r := recover(); r != nil {
				if handler != nil {
					handler(ctx, r)
				}
			}
		}()
		return next(ctx)
	}
}
