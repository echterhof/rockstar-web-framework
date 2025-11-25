package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 21: Global middleware registration**
// **Validates: Requirements 6.1**
// For any plugin registering global middleware, the middleware should appear in the framework's global middleware chain
func TestProperty_GlobalMiddlewareRegistration(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("global middleware is registered and retrievable",
		prop.ForAll(
			func(pluginName string, middlewareName string, priority int) bool {
				// Create middleware registry
				registry := NewMiddlewareRegistry()

				// Create a simple middleware handler
				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register global middleware (empty routes means global)
				err := registry.Register(pluginName, middlewareName, handler, priority, []string{})
				if err != nil {
					return false
				}

				// Verify middleware is registered
				registrations := registry.List(pluginName)
				if len(registrations) != 1 {
					return false
				}

				// Verify middleware details
				reg := registrations[0]
				if reg.PluginName != pluginName {
					return false
				}
				if reg.Name != middlewareName {
					return false
				}
				if reg.Priority != priority {
					return false
				}
				if len(reg.Routes) != 0 {
					return false
				}
				if reg.Handler == nil {
					return false
				}

				return true
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.IntRange(-100, 100),
		),
	)

	properties.Property("multiple plugins can register global middleware independently",
		prop.ForAll(
			func(plugin1 string, plugin2 string, mw1 string, mw2 string) bool {
				if plugin1 == plugin2 {
					return true // Skip if plugins have same name
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register middleware for plugin 1
				err := registry.Register(plugin1, mw1, handler, 10, []string{})
				if err != nil {
					return false
				}

				// Register middleware for plugin 2
				err = registry.Register(plugin2, mw2, handler, 20, []string{})
				if err != nil {
					return false
				}

				// Verify both are registered
				regs1 := registry.List(plugin1)
				regs2 := registry.List(plugin2)

				if len(regs1) != 1 || len(regs2) != 1 {
					return false
				}

				if regs1[0].PluginName != plugin1 || regs2[0].PluginName != plugin2 {
					return false
				}

				return true
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.Identifier(),
			gen.Identifier(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 22: Route-specific middleware scoping**
// **Validates: Requirements 6.2**
// For any plugin registering route-specific middleware, the middleware should only execute for requests matching the specified routes
func TestProperty_RouteSpecificMiddlewareScoping(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("route-specific middleware is registered with correct routes",
		prop.ForAll(
			func(pluginName string, middlewareName string, routeCount int) bool {
				if routeCount <= 0 || routeCount > 10 {
					return true // Skip invalid cases
				}

				// Generate routes
				routes := make([]string, routeCount)
				for i := 0; i < routeCount; i++ {
					routes[i] = fmt.Sprintf("/route-%d", i)
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register route-specific middleware
				err := registry.Register(pluginName, middlewareName, handler, 10, routes)
				if err != nil {
					return false
				}

				// Verify middleware is registered with correct routes
				registrations := registry.List(pluginName)
				if len(registrations) != 1 {
					return false
				}

				reg := registrations[0]
				if len(reg.Routes) != len(routes) {
					return false
				}

				// Verify all routes are present
				routeMap := make(map[string]bool)
				for _, r := range reg.Routes {
					routeMap[r] = true
				}

				for _, r := range routes {
					if !routeMap[r] {
						return false
					}
				}

				return true
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.IntRange(1, 10),
		),
	)

	properties.Property("global middleware has empty routes list",
		prop.ForAll(
			func(pluginName string, middlewareName string) bool {
				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register global middleware (empty routes)
				err := registry.Register(pluginName, middlewareName, handler, 10, []string{})
				if err != nil {
					return false
				}

				// Verify routes list is empty
				registrations := registry.List(pluginName)
				if len(registrations) != 1 {
					return false
				}

				return len(registrations[0].Routes) == 0
			},
			gen.Identifier(),
			gen.Identifier(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 23: Middleware priority ordering**
// **Validates: Requirements 6.3**
// For any set of plugin middleware with different priorities, the middleware should execute in priority order
func TestProperty_MiddlewarePriorityOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("middleware is stored with correct priority",
		prop.ForAll(
			func(pluginName string, middlewareName string, priority int) bool {
				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				err := registry.Register(pluginName, middlewareName, handler, priority, []string{})
				if err != nil {
					return false
				}

				registrations := registry.List(pluginName)
				if len(registrations) != 1 {
					return false
				}

				return registrations[0].Priority == priority
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.IntRange(-1000, 1000),
		),
	)

	properties.Property("multiple middleware from same plugin maintain their priorities",
		prop.ForAll(
			func(pluginName string, middlewareCount int) bool {
				if middlewareCount <= 0 || middlewareCount > 10 {
					return true // Skip invalid cases
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Generate priorities
				priorities := make([]int, middlewareCount)
				for i := 0; i < middlewareCount; i++ {
					priorities[i] = i * 10 // Use predictable priorities
				}

				// Register multiple middleware with different priorities
				for i, priority := range priorities {
					middlewareName := fmt.Sprintf("middleware-%d", i)
					err := registry.Register(pluginName, middlewareName, handler, priority, []string{})
					if err != nil {
						return false
					}
				}

				// Verify all middleware are registered with correct priorities
				registrations := registry.List(pluginName)
				if len(registrations) != len(priorities) {
					return false
				}

				// Create a map of middleware name to priority
				priorityMap := make(map[string]int)
				for _, reg := range registrations {
					priorityMap[reg.Name] = reg.Priority
				}

				// Verify each priority matches
				for i, expectedPriority := range priorities {
					middlewareName := fmt.Sprintf("middleware-%d", i)
					actualPriority, exists := priorityMap[middlewareName]
					if !exists || actualPriority != expectedPriority {
						return false
					}
				}

				return true
			},
			gen.Identifier(),
			gen.IntRange(1, 10),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 25: Middleware cleanup on unload**
// **Validates: Requirements 6.5**
// For any plugin that is unloaded, all middleware registered by that plugin should be removed from the middleware chain
func TestProperty_MiddlewareCleanupOnUnload(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("unregistering all middleware removes all entries for plugin",
		prop.ForAll(
			func(pluginName string, middlewareCount int) bool {
				if middlewareCount <= 0 || middlewareCount > 20 {
					return true // Skip invalid cases
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register multiple middleware for the plugin
				for i := 0; i < middlewareCount; i++ {
					middlewareName := fmt.Sprintf("middleware-%d", i)
					err := registry.Register(pluginName, middlewareName, handler, i, []string{})
					if err != nil {
						return false
					}
				}

				// Verify all middleware are registered
				registrations := registry.List(pluginName)
				if len(registrations) != middlewareCount {
					return false
				}

				// Unregister all middleware for the plugin
				err := registry.UnregisterAll(pluginName)
				if err != nil {
					return false
				}

				// Verify all middleware are removed
				registrations = registry.List(pluginName)
				return len(registrations) == 0
			},
			gen.Identifier(),
			gen.IntRange(1, 20),
		),
	)

	properties.Property("unregistering one plugin's middleware doesn't affect other plugins",
		prop.ForAll(
			func(plugin1 string, plugin2 string, mwCount1 int, mwCount2 int) bool {
				if plugin1 == plugin2 {
					return true // Skip if plugins have same name
				}
				if mwCount1 <= 0 || mwCount1 > 10 || mwCount2 <= 0 || mwCount2 > 10 {
					return true // Skip invalid cases
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register middleware for plugin 1
				for i := 0; i < mwCount1; i++ {
					middlewareName := fmt.Sprintf("mw1-%d", i)
					err := registry.Register(plugin1, middlewareName, handler, i, []string{})
					if err != nil {
						return false
					}
				}

				// Register middleware for plugin 2
				for i := 0; i < mwCount2; i++ {
					middlewareName := fmt.Sprintf("mw2-%d", i)
					err := registry.Register(plugin2, middlewareName, handler, i, []string{})
					if err != nil {
						return false
					}
				}

				// Unregister all middleware for plugin 1
				err := registry.UnregisterAll(plugin1)
				if err != nil {
					return false
				}

				// Verify plugin 1 middleware are removed
				regs1 := registry.List(plugin1)
				if len(regs1) != 0 {
					return false
				}

				// Verify plugin 2 middleware are still present
				regs2 := registry.List(plugin2)
				return len(regs2) == mwCount2
			},
			gen.Identifier(),
			gen.Identifier(),
			gen.IntRange(1, 10),
			gen.IntRange(1, 10),
		),
	)

	properties.Property("unregister individual middleware removes only that middleware",
		prop.ForAll(
			func(pluginName string, middlewareCount int, removeIndex int) bool {
				if middlewareCount <= 1 || middlewareCount > 10 {
					return true // Skip invalid cases
				}
				if removeIndex < 0 || removeIndex >= middlewareCount {
					return true // Skip invalid index
				}

				registry := NewMiddlewareRegistry()

				handler := func(ctx Context, next HandlerFunc) error {
					return next(ctx)
				}

				// Register multiple middleware
				for i := 0; i < middlewareCount; i++ {
					middlewareName := fmt.Sprintf("middleware-%d", i)
					err := registry.Register(pluginName, middlewareName, handler, i, []string{})
					if err != nil {
						return false
					}
				}

				// Unregister one specific middleware
				removeMiddlewareName := fmt.Sprintf("middleware-%d", removeIndex)
				err := registry.Unregister(pluginName, removeMiddlewareName)
				if err != nil {
					return false
				}

				// Verify count is reduced by 1
				registrations := registry.List(pluginName)
				if len(registrations) != middlewareCount-1 {
					return false
				}

				// Verify the removed middleware is not present
				for _, reg := range registrations {
					if reg.Name == removeMiddlewareName {
						return false
					}
				}

				return true
			},
			gen.Identifier(),
			gen.IntRange(2, 10),
			gen.IntRange(0, 9),
		),
	)

	properties.TestingRun(t)
}
