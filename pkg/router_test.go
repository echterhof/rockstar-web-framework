package pkg

import (
	"net/http"
	"testing"
)

// TestRouterBasicRouting tests basic route registration and matching
func TestRouterBasicRouting(t *testing.T) {
	router := NewRouter()

	// Register routes
	router.GET("/users", func(ctx Context) error {
		return nil
	})

	router.POST("/users", func(ctx Context) error {
		return nil
	})

	router.PUT("/users/:id", func(ctx Context) error {
		return nil
	})

	// Test GET route
	route, params, found := router.Match("GET", "/users", "")
	if !found {
		t.Error("Expected to find GET /users route")
	}
	if route.Method != "GET" || route.Path != "/users" {
		t.Errorf("Expected GET /users, got %s %s", route.Method, route.Path)
	}
	if len(params) != 0 {
		t.Errorf("Expected no params, got %v", params)
	}

	// Test POST route
	route, params, found = router.Match("POST", "/users", "")
	if !found {
		t.Error("Expected to find POST /users route")
	}
	if route.Method != "POST" {
		t.Errorf("Expected POST method, got %s", route.Method)
	}

	// Test PUT route with parameter
	route, params, found = router.Match("PUT", "/users/123", "")
	if !found {
		t.Error("Expected to find PUT /users/:id route")
	}
	if route.Method != "PUT" {
		t.Errorf("Expected PUT method, got %s", route.Method)
	}
	if params["id"] != "123" {
		t.Errorf("Expected id=123, got %v", params)
	}

	// Test non-existent route
	_, _, found = router.Match("DELETE", "/users", "")
	if found {
		t.Error("Expected not to find DELETE /users route")
	}
}

// TestRouterDynamicParameters tests dynamic path parameter extraction
func TestRouterDynamicParameters(t *testing.T) {
	router := NewRouter()

	router.GET("/users/:id", func(ctx Context) error {
		return nil
	})

	router.GET("/users/:id/posts/:postId", func(ctx Context) error {
		return nil
	})

	// Test single parameter
	_, params, found := router.Match("GET", "/users/42", "")
	if !found {
		t.Error("Expected to find route")
	}
	if params["id"] != "42" {
		t.Errorf("Expected id=42, got %s", params["id"])
	}

	// Test multiple parameters
	_, params, found = router.Match("GET", "/users/42/posts/99", "")
	if !found {
		t.Error("Expected to find route")
	}
	if params["id"] != "42" {
		t.Errorf("Expected id=42, got %s", params["id"])
	}
	if params["postId"] != "99" {
		t.Errorf("Expected postId=99, got %s", params["postId"])
	}
}

// TestRouterWildcard tests wildcard path matching
func TestRouterWildcard(t *testing.T) {
	router := NewRouter()

	router.GET("/static/*filepath", func(ctx Context) error {
		return nil
	})

	// Test wildcard matching
	_, params, found := router.Match("GET", "/static/css/style.css", "")
	if !found {
		t.Error("Expected to find wildcard route")
	}
	if params["filepath"] != "css/style.css" {
		t.Errorf("Expected filepath=css/style.css, got %s", params["filepath"])
	}

	// Test deeper path
	_, params, found = router.Match("GET", "/static/js/lib/jquery.min.js", "")
	if !found {
		t.Error("Expected to find wildcard route")
	}
	if params["filepath"] != "js/lib/jquery.min.js" {
		t.Errorf("Expected filepath=js/lib/jquery.min.js, got %s", params["filepath"])
	}
}

// TestRouterHostBasedRouting tests host-specific routing for multi-tenancy
func TestRouterHostBasedRouting(t *testing.T) {
	router := NewRouter()

	// Register routes for different hosts
	router.Host("api.example.com").GET("/users", func(ctx Context) error {
		return nil
	})

	router.Host("admin.example.com").GET("/users", func(ctx Context) error {
		return nil
	})

	// Default route (no host)
	router.GET("/users", func(ctx Context) error {
		return nil
	})

	// Test host-specific route
	_, _, found := router.Match("GET", "/users", "api.example.com")
	if !found {
		t.Error("Expected to find route for api.example.com")
	}

	// Test different host
	_, _, found = router.Match("GET", "/users", "admin.example.com")
	if !found {
		t.Error("Expected to find route for admin.example.com")
	}

	// Test default route (no host match)
	_, _, found = router.Match("GET", "/users", "other.example.com")
	if !found {
		t.Error("Expected to find default route")
	}

	// Test with empty host
	_, _, found = router.Match("GET", "/users", "")
	if !found {
		t.Error("Expected to find default route with empty host")
	}
}

// TestRouterGroups tests route grouping with prefixes
func TestRouterGroups(t *testing.T) {
	router := NewRouter()

	// Create API v1 group
	v1 := router.Group("/api/v1")
	v1.GET("/users", func(ctx Context) error {
		return nil
	})
	v1.POST("/users", func(ctx Context) error {
		return nil
	})

	// Create API v2 group
	v2 := router.Group("/api/v2")
	v2.GET("/users", func(ctx Context) error {
		return nil
	})

	// Test v1 routes
	route, _, found := router.Match("GET", "/api/v1/users", "")
	if !found {
		t.Error("Expected to find /api/v1/users route")
	}
	if route.Path != "/api/v1/users" {
		t.Errorf("Expected path /api/v1/users, got %s", route.Path)
	}

	// Test v2 routes
	route, _, found = router.Match("GET", "/api/v2/users", "")
	if !found {
		t.Error("Expected to find /api/v2/users route")
	}
	if route.Path != "/api/v2/users" {
		t.Errorf("Expected path /api/v2/users, got %s", route.Path)
	}

	// Test nested groups
	admin := router.Group("/admin")
	users := admin.Group("/users")
	users.GET("/:id", func(ctx Context) error {
		return nil
	})

	route, params, found := router.Match("GET", "/admin/users/123", "")
	if !found {
		t.Error("Expected to find nested group route")
	}
	if params["id"] != "123" {
		t.Errorf("Expected id=123, got %s", params["id"])
	}
}

// TestRouterMiddleware tests middleware registration
func TestRouterMiddleware(t *testing.T) {
	router := NewRouter()

	middleware1 := func(ctx Context, next HandlerFunc) error {
		return next(ctx)
	}

	middleware2 := func(ctx Context, next HandlerFunc) error {
		return next(ctx)
	}

	// Register route with middleware
	router.GET("/users", func(ctx Context) error {
		return nil
	}, middleware1, middleware2)

	route, _, found := router.Match("GET", "/users", "")
	if !found {
		t.Error("Expected to find route")
	}

	if len(route.Middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(route.Middleware))
	}

	// Test group middleware
	group := router.Group("/api", middleware1)
	group.GET("/users", func(ctx Context) error {
		return nil
	}, middleware2)

	route, _, found = router.Match("GET", "/api/users", "")
	if !found {
		t.Error("Expected to find route")
	}

	if len(route.Middleware) != 2 {
		t.Errorf("Expected 2 middleware (group + route), got %d", len(route.Middleware))
	}
}

// TestRouterAllHTTPMethods tests all HTTP method helpers
func TestRouterAllHTTPMethods(t *testing.T) {
	router := NewRouter()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		switch method {
		case "GET":
			router.GET("/test", func(ctx Context) error { return nil })
		case "POST":
			router.POST("/test", func(ctx Context) error { return nil })
		case "PUT":
			router.PUT("/test", func(ctx Context) error { return nil })
		case "DELETE":
			router.DELETE("/test", func(ctx Context) error { return nil })
		case "PATCH":
			router.PATCH("/test", func(ctx Context) error { return nil })
		case "HEAD":
			router.HEAD("/test", func(ctx Context) error { return nil })
		case "OPTIONS":
			router.OPTIONS("/test", func(ctx Context) error { return nil })
		}
	}

	// Verify all methods are registered
	for _, method := range methods {
		route, _, found := router.Match(method, "/test", "")
		if !found {
			t.Errorf("Expected to find %s /test route", method)
		}
		if route.Method != method {
			t.Errorf("Expected method %s, got %s", method, route.Method)
		}
	}
}

// TestRouterRoutesList tests retrieving all registered routes
func TestRouterRoutesList(t *testing.T) {
	router := NewRouter()

	router.GET("/users", func(ctx Context) error { return nil })
	router.POST("/users", func(ctx Context) error { return nil })
	router.GET("/posts", func(ctx Context) error { return nil })

	routes := router.Routes()
	if len(routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(routes))
	}
}

// TestRouterHostRoutesList tests retrieving routes including host-specific ones
func TestRouterHostRoutesList(t *testing.T) {
	router := NewRouter()

	router.GET("/users", func(ctx Context) error { return nil })
	router.Host("api.example.com").GET("/users", func(ctx Context) error { return nil })
	router.Host("admin.example.com").GET("/users", func(ctx Context) error { return nil })

	routes := router.Routes()

	// Should have 3 routes total (1 default + 2 host-specific)
	if len(routes) < 3 {
		t.Errorf("Expected at least 3 routes, got %d", len(routes))
	}

	// Check that host-specific routes have Host field set
	hostRouteCount := 0
	for _, route := range routes {
		if route.Host != "" {
			hostRouteCount++
		}
	}

	if hostRouteCount != 2 {
		t.Errorf("Expected 2 host-specific routes, got %d", hostRouteCount)
	}
}

// TestRouterComplexPaths tests complex path patterns
func TestRouterComplexPaths(t *testing.T) {
	router := NewRouter()

	// Register complex routes
	router.GET("/users/:userId/posts/:postId/comments/:commentId", func(ctx Context) error {
		return nil
	})

	_, params, found := router.Match("GET", "/users/1/posts/2/comments/3", "")
	if !found {
		t.Error("Expected to find complex route")
	}

	if params["userId"] != "1" {
		t.Errorf("Expected userId=1, got %s", params["userId"])
	}
	if params["postId"] != "2" {
		t.Errorf("Expected postId=2, got %s", params["postId"])
	}
	if params["commentId"] != "3" {
		t.Errorf("Expected commentId=3, got %s", params["commentId"])
	}
}

// TestRouterStaticRoutes tests static file serving routes
func TestRouterStaticRoutes(t *testing.T) {
	router := NewRouter()

	// Mock virtual filesystem
	mockFS := &mockVirtualFS{}

	router.Static("/static", mockFS)

	route, params, found := router.Match("GET", "/static/css/style.css", "")
	if !found {
		t.Error("Expected to find static route")
	}

	if !route.IsStatic {
		t.Error("Expected route to be marked as static")
	}

	if params["filepath"] != "css/style.css" {
		t.Errorf("Expected filepath=css/style.css, got %s", params["filepath"])
	}
}

// TestRouterWebSocket tests WebSocket route registration
func TestRouterWebSocket(t *testing.T) {
	router := NewRouter()

	router.WebSocket("/ws", func(ctx Context, conn WebSocketConnection) error {
		return nil
	})

	route, _, found := router.Match("GET", "/ws", "")
	if !found {
		t.Error("Expected to find WebSocket route")
	}

	if !route.IsWebSocket {
		t.Error("Expected route to be marked as WebSocket")
	}
}

// TestRouterConcurrency tests concurrent route registration and matching
func TestRouterConcurrency(t *testing.T) {
	router := NewRouter()

	// Register routes concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			router.GET("/test", func(ctx Context) error {
				return nil
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Match routes concurrently
	for i := 0; i < 10; i++ {
		go func() {
			_, _, _ = router.Match("GET", "/test", "")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// mockVirtualFS is a mock implementation of VirtualFS for testing
type mockVirtualFS struct{}

func (m *mockVirtualFS) Open(name string) (http.File, error) {
	return nil, nil
}

func (m *mockVirtualFS) Exists(name string) bool {
	return true
}
