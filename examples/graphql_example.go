package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// mockDatabase implements a simple in-memory database for the example
type mockDatabase struct {
	rateLimits map[string]int
	mu         sync.RWMutex
}

func (m *mockDatabase) Connect(config pkg.DatabaseConfig) error                        { return nil }
func (m *mockDatabase) Close() error                                                   { return nil }
func (m *mockDatabase) Ping() error                                                    { return nil }
func (m *mockDatabase) Stats() pkg.DatabaseStats                                       { return pkg.DatabaseStats{} }
func (m *mockDatabase) Query(query string, args ...interface{}) (*sql.Rows, error)     { return nil, nil }
func (m *mockDatabase) QueryRow(query string, args ...interface{}) *sql.Row            { return nil }
func (m *mockDatabase) Exec(query string, args ...interface{}) (sql.Result, error)     { return nil, nil }
func (m *mockDatabase) Prepare(query string) (*sql.Stmt, error)                        { return nil, nil }
func (m *mockDatabase) Begin() (pkg.Transaction, error)                                { return nil, nil }
func (m *mockDatabase) BeginTx(opts *sql.TxOptions) (pkg.Transaction, error)           { return nil, nil }
func (m *mockDatabase) Save(model interface{}) error                                   { return nil }
func (m *mockDatabase) Find(model interface{}, conditions ...pkg.Condition) error      { return nil }
func (m *mockDatabase) FindAll(models interface{}, conditions ...pkg.Condition) error  { return nil }
func (m *mockDatabase) Delete(model interface{}) error                                 { return nil }
func (m *mockDatabase) Update(model interface{}, updates map[string]interface{}) error { return nil }
func (m *mockDatabase) SaveSession(session *pkg.Session) error                         { return nil }
func (m *mockDatabase) LoadSession(sessionID string) (*pkg.Session, error)             { return nil, nil }
func (m *mockDatabase) DeleteSession(sessionID string) error                           { return nil }
func (m *mockDatabase) CleanupExpiredSessions() error                                  { return nil }
func (m *mockDatabase) SaveAccessToken(token *pkg.AccessToken) error                   { return nil }
func (m *mockDatabase) LoadAccessToken(tokenValue string) (*pkg.AccessToken, error)    { return nil, nil }
func (m *mockDatabase) ValidateAccessToken(tokenValue string) (*pkg.AccessToken, error) {
	return nil, nil
}
func (m *mockDatabase) DeleteAccessToken(tokenValue string) error              { return nil }
func (m *mockDatabase) CleanupExpiredTokens() error                            { return nil }
func (m *mockDatabase) SaveTenant(tenant *pkg.Tenant) error                    { return nil }
func (m *mockDatabase) LoadTenant(tenantID string) (*pkg.Tenant, error)        { return nil, nil }
func (m *mockDatabase) LoadTenantByHost(hostname string) (*pkg.Tenant, error)  { return nil, nil }
func (m *mockDatabase) SaveWorkloadMetrics(metrics *pkg.WorkloadMetrics) error { return nil }
func (m *mockDatabase) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*pkg.WorkloadMetrics, error) {
	return nil, nil
}
func (m *mockDatabase) Migrate() error      { return nil }
func (m *mockDatabase) CreateTables() error { return nil }
func (m *mockDatabase) DropTables() error   { return nil }

func (m *mockDatabase) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := m.rateLimits[key]
	return count < limit, nil
}

func (m *mockDatabase) IncrementRateLimit(key string, window time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimits[key]++
	return nil
}

// SimpleGraphQLSchema implements a basic GraphQL schema for demonstration
type SimpleGraphQLSchema struct {
	users map[string]User
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Execute implements the GraphQLSchema interface
func (s *SimpleGraphQLSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
	// This is a simplified implementation for demonstration
	// In a real application, you would use a proper GraphQL library like graphql-go

	// Simple query parsing (for demo purposes only)
	if query == "{ users { id name email } }" {
		users := make([]User, 0, len(s.users))
		for _, user := range s.users {
			users = append(users, user)
		}
		return map[string]interface{}{
			"users": users,
		}, nil
	}

	if query == "{ hello }" {
		return map[string]interface{}{
			"hello": "world",
		}, nil
	}

	return nil, fmt.Errorf("unsupported query")
}

func main() {
	// Create router
	router := pkg.NewRouter()

	// Create database manager (using mock for example)
	db := &mockDatabase{
		rateLimits: make(map[string]int),
	}

	// Create auth manager
	authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

	// Create GraphQL manager
	graphqlManager := pkg.NewGraphQLManager(router, db, authManager)

	// Create a simple schema
	schema := &SimpleGraphQLSchema{
		users: map[string]User{
			"1": {ID: "1", Name: "John Doe", Email: "john@example.com"},
			"2": {ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
		},
	}

	// Register GraphQL endpoint with configuration
	config := pkg.GraphQLConfig{
		EnableIntrospection: true,
		EnablePlayground:    true,
		MaxRequestSize:      1024 * 1024, // 1MB
		Timeout:             30 * time.Second,
		RateLimit: &pkg.GraphQLRateLimitConfig{
			Limit:  100,
			Window: time.Minute,
			Key:    "ip_address",
		},
	}

	err := graphqlManager.RegisterSchema("/graphql", schema, config)
	if err != nil {
		log.Fatalf("Failed to register GraphQL schema: %v", err)
	}

	// Register authenticated GraphQL endpoint
	authenticatedConfig := pkg.GraphQLConfig{
		RequireAuth:         true,
		RequiredRoles:       []string{"user"},
		EnableIntrospection: false,
		EnablePlayground:    false,
		MaxRequestSize:      1024 * 1024,
		Timeout:             30 * time.Second,
		RateLimit: &pkg.GraphQLRateLimitConfig{
			Limit:  50,
			Window: time.Minute,
			Key:    "user_id",
		},
	}

	err = graphqlManager.RegisterSchema("/api/graphql", schema, authenticatedConfig)
	if err != nil {
		log.Fatalf("Failed to register authenticated GraphQL schema: %v", err)
	}

	// Create API group with middleware
	apiV1 := graphqlManager.Group("/api/v1")

	// Add custom middleware
	apiV1.Use(func(ctx pkg.Context, next pkg.GraphQLHandler) error {
		log.Printf("GraphQL request to: %s", ctx.Request().URL.Path)
		return next(ctx)
	})

	// Register schema in group
	err = apiV1.RegisterSchema("/graphql", schema, config)
	if err != nil {
		log.Fatalf("Failed to register GraphQL schema in group: %v", err)
	}

	// Create and configure server
	server := pkg.NewServer(pkg.ServerConfig{
		EnableHTTP1: true,
		EnableHTTP2: true,
	})

	server.SetRouter(router)
	server.EnableHTTP1().EnableHTTP2()

	// Print registered routes
	fmt.Println("Registered GraphQL endpoints:")
	for _, route := range router.Routes() {
		fmt.Printf("  %s %s\n", route.Method, route.Path)
	}

	fmt.Println("\nGraphQL endpoints:")
	fmt.Println("  Public GraphQL API: http://localhost:8080/graphql")
	fmt.Println("  GraphQL Playground: http://localhost:8080/graphql (GET)")
	fmt.Println("  Authenticated API: http://localhost:8080/api/graphql")
	fmt.Println("  API v1: http://localhost:8080/api/v1/graphql")

	fmt.Println("\nExample queries:")
	fmt.Println("  { hello }")
	fmt.Println("  { users { id name email } }")

	fmt.Println("\nStarting server on :8080...")

	// Start server
	if err := server.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
