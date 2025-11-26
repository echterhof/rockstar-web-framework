package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// 🎸 GraphQL API Example
// This example demonstrates GraphQL API implementation with the Rockstar Web Framework
// Features: Schema definition, queries, mutations, error handling

// User represents a user in our system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Post represents a blog post
type Post struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

// In-memory storage for demonstration
var (
	users     = make(map[string]*User)
	posts     = make(map[string]*Post)
	userIDSeq = 1
	postIDSeq = 1
)

// BlogSchema implements a GraphQL schema for a blog
type BlogSchema struct{}

// Execute implements the GraphQLSchema interface
// This is a simplified GraphQL implementation for demonstration
// In production, use a proper GraphQL library like graphql-go or gqlgen
func (s *BlogSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
	// Trim whitespace and normalize query
	query = strings.TrimSpace(query)

	// Parse and execute query
	// This is a very basic parser for demonstration purposes
	// A real implementation would use a proper GraphQL parser

	// Query: { users { id name email role } }
	if strings.Contains(query, "users") && strings.Contains(query, "{") {
		return s.queryUsers()
	}

	// Query: { user(id: "1") { id name email role } }
	if strings.Contains(query, "user(") {
		id := extractID(query, "user")
		return s.queryUser(id)
	}

	// Query: { posts { id title content author_id published } }
	if strings.Contains(query, "posts") && strings.Contains(query, "{") {
		return s.queryPosts()
	}

	// Query: { post(id: "1") { id title content author_id published } }
	if strings.Contains(query, "post(") {
		id := extractID(query, "post")
		return s.queryPost(id)
	}

	// Mutation: createUser(name: "...", email: "...", role: "...")
	if strings.Contains(query, "createUser") {
		name := extractStringArg(query, "name")
		email := extractStringArg(query, "email")
		role := extractStringArg(query, "role")
		return s.mutationCreateUser(name, email, role)
	}

	// Mutation: createPost(title: "...", content: "...", author_id: "...")
	if strings.Contains(query, "createPost") {
		title := extractStringArg(query, "title")
		content := extractStringArg(query, "content")
		authorID := extractStringArg(query, "author_id")
		return s.mutationCreatePost(title, content, authorID)
	}

	// Mutation: updatePost(id: "...", published: true)
	if strings.Contains(query, "updatePost") {
		id := extractStringArg(query, "id")
		published := strings.Contains(query, "published: true")
		return s.mutationUpdatePost(id, published)
	}

	// Mutation: deletePost(id: "...")
	if strings.Contains(query, "deletePost") {
		id := extractStringArg(query, "id")
		return s.mutationDeletePost(id)
	}

	// Introspection query
	if strings.Contains(query, "__schema") || strings.Contains(query, "__type") {
		return s.introspection()
	}

	return nil, fmt.Errorf("unsupported query: %s", query)
}

// Query resolvers

func (s *BlogSchema) queryUsers() (interface{}, error) {
	userList := make([]*User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	return map[string]interface{}{
		"users": userList,
	}, nil
}

func (s *BlogSchema) queryUser(id string) (interface{}, error) {
	user, exists := users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return map[string]interface{}{
		"user": user,
	}, nil
}

func (s *BlogSchema) queryPosts() (interface{}, error) {
	postList := make([]*Post, 0, len(posts))
	for _, post := range posts {
		postList = append(postList, post)
	}
	return map[string]interface{}{
		"posts": postList,
	}, nil
}

func (s *BlogSchema) queryPost(id string) (interface{}, error) {
	post, exists := posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found: %s", id)
	}
	return map[string]interface{}{
		"post": post,
	}, nil
}

// Mutation resolvers

func (s *BlogSchema) mutationCreateUser(name, email, role string) (interface{}, error) {
	if name == "" || email == "" {
		return nil, fmt.Errorf("name and email are required")
	}
	if role == "" {
		role = "user"
	}

	id := fmt.Sprintf("%d", userIDSeq)
	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
	}
	users[id] = user
	userIDSeq++

	return map[string]interface{}{
		"createUser": user,
	}, nil
}

func (s *BlogSchema) mutationCreatePost(title, content, authorID string) (interface{}, error) {
	if title == "" || content == "" || authorID == "" {
		return nil, fmt.Errorf("title, content, and author_id are required")
	}

	// Verify author exists
	if _, exists := users[authorID]; !exists {
		return nil, fmt.Errorf("author not found: %s", authorID)
	}

	id := fmt.Sprintf("%d", postIDSeq)
	post := &Post{
		ID:        id,
		Title:     title,
		Content:   content,
		AuthorID:  authorID,
		Published: false,
		CreatedAt: time.Now(),
	}
	posts[id] = post
	postIDSeq++

	return map[string]interface{}{
		"createPost": post,
	}, nil
}

func (s *BlogSchema) mutationUpdatePost(id string, published bool) (interface{}, error) {
	post, exists := posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found: %s", id)
	}

	post.Published = published

	return map[string]interface{}{
		"updatePost": post,
	}, nil
}

func (s *BlogSchema) mutationDeletePost(id string) (interface{}, error) {
	_, exists := posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found: %s", id)
	}

	delete(posts, id)

	return map[string]interface{}{
		"deletePost": map[string]interface{}{
			"success": true,
			"id":      id,
		},
	}, nil
}

// introspection returns schema introspection data
func (s *BlogSchema) introspection() (interface{}, error) {
	return map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []map[string]interface{}{
				{
					"name": "User",
					"fields": []map[string]string{
						{"name": "id", "type": "String!"},
						{"name": "name", "type": "String!"},
						{"name": "email", "type": "String!"},
						{"name": "role", "type": "String!"},
						{"name": "created_at", "type": "String!"},
					},
				},
				{
					"name": "Post",
					"fields": []map[string]string{
						{"name": "id", "type": "String!"},
						{"name": "title", "type": "String!"},
						{"name": "content", "type": "String!"},
						{"name": "author_id", "type": "String!"},
						{"name": "published", "type": "Boolean!"},
						{"name": "created_at", "type": "String!"},
					},
				},
			},
		},
	}, nil
}

// Helper functions for parsing GraphQL queries
// Note: These are simplified parsers for demonstration
// Use a proper GraphQL library in production

func extractID(query, operation string) string {
	// Extract id from: operation(id: "123")
	start := strings.Index(query, operation+"(")
	if start == -1 {
		return ""
	}
	start += len(operation) + 1

	idStart := strings.Index(query[start:], "id:")
	if idStart == -1 {
		return ""
	}
	idStart += start + 3

	// Find the value between quotes
	quoteStart := strings.Index(query[idStart:], "\"")
	if quoteStart == -1 {
		return ""
	}
	quoteStart += idStart + 1

	quoteEnd := strings.Index(query[quoteStart:], "\"")
	if quoteEnd == -1 {
		return ""
	}

	return query[quoteStart : quoteStart+quoteEnd]
}

func extractStringArg(query, argName string) string {
	// Extract argument from: argName: "value"
	start := strings.Index(query, argName+":")
	if start == -1 {
		return ""
	}
	start += len(argName) + 1

	// Find the value between quotes
	quoteStart := strings.Index(query[start:], "\"")
	if quoteStart == -1 {
		return ""
	}
	quoteStart += start + 1

	quoteEnd := strings.Index(query[quoteStart:], "\"")
	if quoteEnd == -1 {
		return ""
	}

	return query[quoteStart : quoteStart+quoteEnd]
}

func main() {
	// Initialize sample data
	initSampleData()

	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get router and database
	router := app.Router()
	db := app.Database()

	// Create auth manager (for authenticated endpoints)
	authManager := pkg.NewAuthManager(db, "secret-key", pkg.OAuth2Config{})

	// Create GraphQL manager
	graphqlManager := pkg.NewGraphQLManager(router, db, authManager)

	// Create schema
	schema := &BlogSchema{}

	// Configure public GraphQL endpoint with playground
	publicConfig := pkg.GraphQLConfig{
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

	// Register public GraphQL endpoint
	err = graphqlManager.RegisterSchema("/graphql", schema, publicConfig)
	if err != nil {
		log.Fatalf("Failed to register GraphQL schema: %v", err)
	}

	// Print startup information
	fmt.Println("🎸 GraphQL API Example")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Server listening on http://localhost:8080")
	fmt.Println()
	fmt.Println("GraphQL endpoints:")
	fmt.Println("  POST http://localhost:8080/graphql       - GraphQL API")
	fmt.Println("  GET  http://localhost:8080/graphql       - GraphQL Playground")
	fmt.Println()
	fmt.Println("Example queries:")
	fmt.Println()
	fmt.Println("  # List all users")
	fmt.Println("  { users { id name email role } }")
	fmt.Println()
	fmt.Println("  # Get user by ID")
	fmt.Println("  { user(id: \"1\") { id name email role } }")
	fmt.Println()
	fmt.Println("  # List all posts")
	fmt.Println("  { posts { id title content author_id published } }")
	fmt.Println()
	fmt.Println("  # Get post by ID")
	fmt.Println("  { post(id: \"1\") { id title content author_id published } }")
	fmt.Println()
	fmt.Println("Example mutations:")
	fmt.Println()
	fmt.Println("  # Create user")
	fmt.Println("  mutation { createUser(name: \"Alice\", email: \"alice@example.com\", role: \"admin\") { id name email } }")
	fmt.Println()
	fmt.Println("  # Create post")
	fmt.Println("  mutation { createPost(title: \"My Post\", content: \"Hello World\", author_id: \"1\") { id title } }")
	fmt.Println()
	fmt.Println("  # Publish post")
	fmt.Println("  mutation { updatePost(id: \"1\", published: true) { id published } }")
	fmt.Println()
	fmt.Println("  # Delete post")
	fmt.Println("  mutation { deletePost(id: \"1\") { success id } }")
	fmt.Println()
	fmt.Println("Try it with curl:")
	fmt.Println("  curl -X POST http://localhost:8080/graphql \\")
	fmt.Println("       -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"query\": \"{ users { id name email } }\"}'")
	fmt.Println()
	fmt.Println("Or open http://localhost:8080/graphql in your browser for the GraphQL Playground")
	fmt.Println()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// initSampleData initializes sample data
func initSampleData() {
	// Create sample users
	users["1"] = &User{
		ID:        "1",
		Name:      "John Doe",
		Email:     "john@example.com",
		Role:      "admin",
		CreatedAt: time.Now(),
	}
	userIDSeq = 2

	users["2"] = &User{
		ID:        "2",
		Name:      "Jane Smith",
		Email:     "jane@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
	}
	userIDSeq = 3

	// Create sample posts
	posts["1"] = &Post{
		ID:        "1",
		Title:     "Getting Started with GraphQL",
		Content:   "GraphQL is a query language for APIs...",
		AuthorID:  "1",
		Published: true,
		CreatedAt: time.Now(),
	}
	postIDSeq = 2

	posts["2"] = &Post{
		ID:        "2",
		Title:     "Building Web APIs",
		Content:   "Learn how to build modern web APIs...",
		AuthorID:  "2",
		Published: false,
		CreatedAt: time.Now(),
	}
	postIDSeq = 3
}
