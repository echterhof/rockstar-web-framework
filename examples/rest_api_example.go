package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// Example demonstrating REST API support in the Rockstar Web Framework
func main() {
	// Create a new router
	router := pkg.NewRouter()

	// Create a mock database for rate limiting
	db := &mockDatabase{
		rateLimits: make(map[string]int),
	}

	// Create REST API manager
	restAPI := pkg.NewRESTAPIManager(router, db)

	// Create a new server
	server := pkg.NewServer(pkg.ServerConfig{
		EnableHTTP1: true,
		EnableHTTP2: true,
	})

	// Set the router on the server
	server.SetRouter(router)

	// Enable HTTP/1 and HTTP/2
	server.EnableHTTP1().EnableHTTP2()

	// Configure rate limiting for API routes
	apiConfig := pkg.RESTRouteConfig{
		RateLimit: &pkg.RESTRateLimitConfig{
			Limit:  100,          // 100 requests
			Window: time.Minute,  // per minute
			Key:    "ip_address", // per IP address
		},
		RequireAuth:    false,       // Set to true to require authentication
		MaxRequestSize: 1024 * 1024, // 1MB max request size
		Timeout:        30 * time.Second,
		CORS: &pkg.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	// Register REST API routes

	// GET /api/users - List users
	restAPI.RegisterRoute("GET", "/api/users", func(ctx pkg.Context) error {
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		}
		return restAPI.SendJSONResponse(ctx, 200, users)
	}, apiConfig)

	// GET /api/users/:id - Get user by ID
	restAPI.RegisterRoute("GET", "/api/users/:id", func(ctx pkg.Context) error {
		userID := ctx.Params()["id"]
		user := map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		}
		return restAPI.SendJSONResponse(ctx, 200, user)
	}, apiConfig)

	// POST /api/users - Create user
	restAPI.RegisterRoute("POST", "/api/users", func(ctx pkg.Context) error {
		var newUser map[string]interface{}
		if err := restAPI.ParseJSONRequest(ctx, &newUser); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Validate required fields
		if newUser["name"] == nil || newUser["email"] == nil {
			return restAPI.SendErrorResponse(ctx, 400, "Missing required fields", map[string]interface{}{
				"required": []string{"name", "email"},
			})
		}

		// Create user (mock)
		newUser["id"] = 3
		newUser["created_at"] = time.Now()

		return restAPI.SendJSONResponse(ctx, 201, newUser)
	}, apiConfig)

	// PUT /api/users/:id - Update user
	restAPI.RegisterRoute("PUT", "/api/users/:id", func(ctx pkg.Context) error {
		userID := ctx.Params()["id"]

		var updates map[string]interface{}
		if err := restAPI.ParseJSONRequest(ctx, &updates); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Update user (mock)
		updatedUser := map[string]interface{}{
			"id":         userID,
			"name":       updates["name"],
			"email":      updates["email"],
			"updated_at": time.Now(),
		}

		return restAPI.SendJSONResponse(ctx, 200, updatedUser)
	}, apiConfig)

	// DELETE /api/users/:id - Delete user
	restAPI.RegisterRoute("DELETE", "/api/users/:id", func(ctx pkg.Context) error {
		userID := ctx.Params()["id"]

		result := map[string]interface{}{
			"deleted": true,
			"id":      userID,
		}

		return restAPI.SendJSONResponse(ctx, 200, result)
	}, apiConfig)

	// Create API v2 group with stricter rate limiting
	apiV2 := restAPI.Group("/api/v2")

	v2Config := pkg.RESTRouteConfig{
		RateLimit: &pkg.RESTRateLimitConfig{
			Limit:  50,          // 50 requests
			Window: time.Minute, // per minute
			Key:    "user_id",   // per user
		},
		RequireAuth:    true,       // Require authentication for v2
		MaxRequestSize: 512 * 1024, // 512KB max request size
		Timeout:        15 * time.Second,
	}

	// Register v2 routes
	apiV2.RegisterRoute("GET", "/users", func(ctx pkg.Context) error {
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com", "version": "v2"},
		}
		return restAPI.SendJSONResponse(ctx, 200, users)
	}, v2Config)

	// Health check endpoint (no rate limiting)
	restAPI.RegisterRoute("GET", "/health", func(ctx pkg.Context) error {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		}
		return restAPI.SendJSONResponse(ctx, 200, health)
	}, pkg.RESTRouteConfig{}) // Empty config = no rate limiting

	// Start the server
	fmt.Println("Starting REST API server on http://localhost:8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /api/users")
	fmt.Println("  GET    /api/users/:id")
	fmt.Println("  POST   /api/users")
	fmt.Println("  PUT    /api/users/:id")
	fmt.Println("  DELETE /api/users/:id")
	fmt.Println("  GET    /api/v2/users (requires auth)")
	fmt.Println("  GET    /health")
	fmt.Println()
	fmt.Println("Rate limits:")
	fmt.Println("  API v1: 100 requests/minute per IP")
	fmt.Println("  API v2: 50 requests/minute per user")

	if err := server.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

// Mock database for demonstration
type mockDatabase struct {
	rateLimits map[string]int
}

func (m *mockDatabase) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	count, exists := m.rateLimits[key]
	if !exists {
		return true, nil
	}
	return count < limit, nil
}

func (m *mockDatabase) IncrementRateLimit(key string, window time.Duration) error {
	m.rateLimits[key]++
	return nil
}

// Implement other required DatabaseManager methods as no-ops
func (m *mockDatabase) Connect(config pkg.DatabaseConfig) error { return nil }
func (m *mockDatabase) Close() error                            { return nil }
func (m *mockDatabase) Ping() error                             { return nil }
func (m *mockDatabase) Stats() pkg.DatabaseStats                { return pkg.DatabaseStats{} }
func (m *mockDatabase) Begin() (pkg.Transaction, error)         { return nil, nil }
func (m *mockDatabase) BeginTx(opts *sql.TxOptions) (pkg.Transaction, error) {
	return nil, nil
}
func (m *mockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockDatabase) QueryRow(query string, args ...interface{}) *sql.Row { return nil }
func (m *mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockDatabase) Prepare(query string) (*sql.Stmt, error)                   { return nil, nil }
func (m *mockDatabase) Save(model interface{}) error                              { return nil }
func (m *mockDatabase) Find(model interface{}, conditions ...pkg.Condition) error { return nil }
func (m *mockDatabase) FindAll(models interface{}, conditions ...pkg.Condition) error {
	return nil
}
func (m *mockDatabase) Delete(model interface{}) error { return nil }
func (m *mockDatabase) Update(model interface{}, updates map[string]interface{}) error {
	return nil
}
func (m *mockDatabase) SaveSession(session *pkg.Session) error             { return nil }
func (m *mockDatabase) LoadSession(sessionID string) (*pkg.Session, error) { return nil, nil }
func (m *mockDatabase) DeleteSession(sessionID string) error               { return nil }
func (m *mockDatabase) CleanupExpiredSessions() error                      { return nil }
func (m *mockDatabase) SaveAccessToken(token *pkg.AccessToken) error       { return nil }
func (m *mockDatabase) LoadAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	return nil, nil
}
func (m *mockDatabase) ValidateAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	return nil, nil
}
func (m *mockDatabase) DeleteAccessToken(tokenValue string) error       { return nil }
func (m *mockDatabase) CleanupExpiredTokens() error                     { return nil }
func (m *mockDatabase) SaveTenant(tenant *pkg.Tenant) error             { return nil }
func (m *mockDatabase) LoadTenant(tenantID string) (*pkg.Tenant, error) { return nil, nil }
func (m *mockDatabase) LoadTenantByHost(hostname string) (*pkg.Tenant, error) {
	return nil, nil
}
func (m *mockDatabase) SaveWorkloadMetrics(metrics *pkg.WorkloadMetrics) error { return nil }
func (m *mockDatabase) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*pkg.WorkloadMetrics, error) {
	return nil, nil
}
func (m *mockDatabase) Migrate() error                { return nil }
func (m *mockDatabase) CreateTables() error           { return nil }
func (m *mockDatabase) DropTables() error             { return nil }
func (m *mockDatabase) InitializePluginTables() error { return nil }
