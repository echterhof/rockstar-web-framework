package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Blog Application - Complete Example
// ============================================================================
// This example demonstrates a complete blog application with:
// - User authentication and sessions
// - Post CRUD operations
// - Comment system
// - Database integration
// - Session management
// - Template rendering (JSON responses for API)
// - Middleware for authentication
// - Error handling
// ============================================================================

// ============================================================================
// Data Models
// ============================================================================

// Post represents a blog post
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  int       `json:"author_id"`
	Author    string    `json:"author,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment represents a comment on a blog post
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  int       `json:"author_id"`
	Author    string    `json:"author,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// BlogUser represents a user in the blog system
type BlogUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	CreatedAt time.Time `json:"created_at"`
}

// CreatePostRequest represents the request body for creating a post
type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdatePostRequest represents the request body for updating a post
type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateCommentRequest represents the request body for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ============================================================================
// In-Memory Storage
// ============================================================================
// In production, replace with actual database operations using ctx.DB()

var (
	posts         = make(map[int]*Post)
	comments      = make(map[int]*Comment)
	users         = make(map[int]*BlogUser)
	nextPostID    = 1
	nextCommentID = 1
	nextUserID    = 1
)

func main() {
	// Initialize sample data
	initSampleData()

	// ========================================================================
	// Configuration Setup
	// ========================================================================
	config := pkg.FrameworkConfig{
		// Server configuration
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  2 << 20, // 2 MB
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableMetrics:   true,
			MetricsPath:     "/metrics",
			ShutdownTimeout: 30 * time.Second,
		},
		// Database configuration - SQLite for simplicity
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "blog_application.db",
		},
		// Cache configuration - for caching frequently accessed data
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    50 * 1024 * 1024, // 50 MB
			DefaultTTL: 10 * time.Minute,
		},
		// Session configuration - manages user authentication sessions
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "blog_session",
			SessionLifetime: 24 * time.Hour,
			CookieSecure:    false,                                      // Set to true in production with HTTPS
			CookieHTTPOnly:  true,                                       // Prevent JavaScript access
			EncryptionKey:   []byte("ABSNTNZMSNENLMONABSNTNZMSNENLMON"), // 32 bytes for AES-256
			CleanupInterval: 10 * time.Minute,
		},
		// Security configuration
		SecurityConfig: pkg.SecurityConfig{
			XFrameOptions:    "SAMEORIGIN",
			EnableCSRF:       true,
			EnableXSSProtect: true,
			MaxRequestSize:   10 * 1024 * 1024, // 10 MB
			RequestTimeout:   30 * time.Second,
		},
		// Monitoring configuration
		MonitoringConfig: pkg.MonitoringConfig{
			EnableMetrics: true,
			MetricsPath:   "/metrics",
		},
	}

	// ========================================================================
	// Framework Initialization
	// ========================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ========================================================================
	// Lifecycle Hooks
	// ========================================================================
	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Blog application starting...")
		fmt.Printf("   Loaded %d users, %d posts, %d comments\n", len(users), len(posts), len(comments))
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Blog application shutting down...")
		fmt.Println("   Saving data and cleaning up resources...")
		return nil
	})

	// ========================================================================
	// Global Middleware
	// ========================================================================
	// Request logging middleware
	app.Use(loggingMiddleware)

	// Recovery middleware - recovers from panics
	app.Use(recoveryMiddleware)

	// CORS middleware - allows cross-origin requests
	app.Use(corsMiddleware)

	// ========================================================================
	// Route Registration
	// ========================================================================
	router := app.Router()

	// Home and documentation endpoints
	router.GET("/", homeHandler)
	router.GET("/health", healthHandler)
	router.GET("/api/docs", apiDocsHandler)

	// Authentication endpoints (public)
	router.POST("/api/auth/login", loginHandler)
	router.POST("/api/auth/logout", logoutHandler)
	router.GET("/api/auth/me", getCurrentUserHandler)

	// Public API routes - no authentication required
	publicAPI := router.Group("/api")
	publicAPI.GET("/posts", listPostsHandler)
	publicAPI.GET("/posts/:id", getPostHandler)
	publicAPI.GET("/posts/:id/comments", listCommentsHandler)
	publicAPI.GET("/users/:id", getUserHandler)

	// Authenticated API routes - require authentication
	authAPI := router.Group("/api", authMiddleware)
	authAPI.POST("/posts", createPostHandler)
	authAPI.PUT("/posts/:id", updatePostHandler)
	authAPI.DELETE("/posts/:id", deletePostHandler)
	authAPI.POST("/posts/:id/comments", createCommentHandler)
	authAPI.DELETE("/comments/:id", deleteCommentHandler)

	// ========================================================================
	// Server Startup
	// ========================================================================
	fmt.Println("🎸 Rockstar Blog Application")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server: http://localhost:8080")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  Home:        GET    /")
	fmt.Println("  Health:      GET    /health")
	fmt.Println("  API Docs:    GET    /api/docs")
	fmt.Println("  Login:       POST   /api/auth/login")
	fmt.Println("  Posts:       GET    /api/posts")
	fmt.Println("  Create Post: POST   /api/posts (auth required)")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Login")
	fmt.Println("  curl -X POST http://localhost:8080/api/auth/login \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"username\":\"john_doe\",\"password\":\"password123\"}'")
	fmt.Println()
	fmt.Println("  # List posts")
	fmt.Println("  curl http://localhost:8080/api/posts")
	fmt.Println()
	fmt.Println("  # Get specific post")
	fmt.Println("  curl http://localhost:8080/api/posts/1")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Start the server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Data Initialization
// ============================================================================

func initSampleData() {
	// Create sample users
	users[1] = &BlogUser{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		Password:  "password123", // In production, use hashed passwords
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
	}
	users[2] = &BlogUser{
		ID:        2,
		Username:  "jane_smith",
		Email:     "jane@example.com",
		Password:  "password456",
		CreatedAt: time.Now().Add(-25 * 24 * time.Hour),
	}
	nextUserID = 3

	// Create sample posts
	posts[1] = &Post{
		ID:        1,
		Title:     "Welcome to Rockstar Blog",
		Content:   "This is the first post on our new blog platform built with Rockstar Web Framework! We're excited to share our journey with you.",
		AuthorID:  1,
		Author:    "john_doe",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	posts[2] = &Post{
		ID:        2,
		Title:     "Getting Started with Go Web Development",
		Content:   "Go is an amazing language for building web applications. In this post, we'll explore why Go is perfect for web development and how to get started.",
		AuthorID:  2,
		Author:    "jane_smith",
		CreatedAt: time.Now().Add(-12 * time.Hour),
		UpdatedAt: time.Now().Add(-12 * time.Hour),
	}
	posts[3] = &Post{
		ID:        3,
		Title:     "Building RESTful APIs with Rockstar",
		Content:   "Learn how to build robust RESTful APIs using the Rockstar Web Framework. We'll cover routing, middleware, and best practices.",
		AuthorID:  1,
		Author:    "john_doe",
		CreatedAt: time.Now().Add(-6 * time.Hour),
		UpdatedAt: time.Now().Add(-6 * time.Hour),
	}
	nextPostID = 4

	// Create sample comments
	comments[1] = &Comment{
		ID:        1,
		PostID:    1,
		AuthorID:  2,
		Author:    "jane_smith",
		Content:   "Great post! Looking forward to more content.",
		CreatedAt: time.Now().Add(-20 * time.Hour),
	}
	comments[2] = &Comment{
		ID:        2,
		PostID:    1,
		AuthorID:  1,
		Author:    "john_doe",
		Content:   "Thanks for the feedback!",
		CreatedAt: time.Now().Add(-19 * time.Hour),
	}
	comments[3] = &Comment{
		ID:        3,
		PostID:    2,
		AuthorID:  1,
		Author:    "john_doe",
		Content:   "Excellent introduction to Go web development!",
		CreatedAt: time.Now().Add(-10 * time.Hour),
	}
	nextCommentID = 4
}

// ============================================================================
// Middleware Functions
// ============================================================================

// loggingMiddleware logs all incoming requests
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()
	method := ctx.Request().Method
	path := ctx.Request().URL.Path

	fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), method, path)

	err := next(ctx)

	duration := time.Since(start)
	fmt.Printf("  ⏱️  Completed in %v\n", duration)

	return err
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Panic recovered: %v\n", r)
			ctx.JSON(500, map[string]interface{}{
				"error":   "Internal server error",
				"message": "An unexpected error occurred",
			})
		}
	}()
	return next(ctx)
}

// corsMiddleware adds CORS headers
func corsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if ctx.Request().Method == "OPTIONS" {
		return ctx.String(204, "")
	}

	return next(ctx)
}

// authMiddleware checks if user is authenticated
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// Get session manager
	sessionMgr := ctx.Session()
	if sessionMgr == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required - session manager not available",
		})
	}

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required - no valid session",
		})
	}

	// Check if user is logged in
	userID, exists := session.Data["user_id"]
	if !exists || userID == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required - not logged in",
		})
	}

	// Store user ID in context for handlers to use
	// In production, you might want to load the full user object
	return next(ctx)
}

// ============================================================================
// Handler Functions - Home and Documentation
// ============================================================================

func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Welcome to Rockstar Blog! 🎸",
		"version": "1.0.0",
		"features": []string{
			"User authentication",
			"Post management",
			"Comment system",
			"Session management",
			"RESTful API",
		},
		"endpoints": map[string]string{
			"docs":   "/api/docs",
			"health": "/health",
			"posts":  "/api/posts",
			"login":  "/api/auth/login",
		},
	})
}

func healthHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
		"stats": map[string]int{
			"users":    len(users),
			"posts":    len(posts),
			"comments": len(comments),
		},
	})
}

func apiDocsHandler(ctx pkg.Context) error {
	docs := map[string]interface{}{
		"title":       "Rockstar Blog API",
		"version":     "1.0.0",
		"description": "Complete blog application API with authentication, posts, and comments",
		"endpoints": []map[string]string{
			// Authentication
			{"method": "POST", "path": "/api/auth/login", "description": "Login with username and password", "auth": "no"},
			{"method": "POST", "path": "/api/auth/logout", "description": "Logout current user", "auth": "yes"},
			{"method": "GET", "path": "/api/auth/me", "description": "Get current user information", "auth": "yes"},
			// Posts
			{"method": "GET", "path": "/api/posts", "description": "List all posts", "auth": "no"},
			{"method": "GET", "path": "/api/posts/:id", "description": "Get a specific post", "auth": "no"},
			{"method": "POST", "path": "/api/posts", "description": "Create a new post", "auth": "yes"},
			{"method": "PUT", "path": "/api/posts/:id", "description": "Update a post", "auth": "yes"},
			{"method": "DELETE", "path": "/api/posts/:id", "description": "Delete a post", "auth": "yes"},
			// Comments
			{"method": "GET", "path": "/api/posts/:id/comments", "description": "List comments for a post", "auth": "no"},
			{"method": "POST", "path": "/api/posts/:id/comments", "description": "Add a comment to a post", "auth": "yes"},
			{"method": "DELETE", "path": "/api/comments/:id", "description": "Delete a comment", "auth": "yes"},
			// Users
			{"method": "GET", "path": "/api/users/:id", "description": "Get user information", "auth": "no"},
		},
	}
	return ctx.JSON(200, docs)
}

// ============================================================================
// Handler Functions - Authentication
// ============================================================================

func loginHandler(ctx pkg.Context) error {
	var req LoginRequest

	// Parse request body
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Username and password are required",
		})
	}

	// Find user by username
	var foundUser *BlogUser
	for _, user := range users {
		if user.Username == req.Username {
			foundUser = user
			break
		}
	}

	if foundUser == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Invalid username or password",
		})
	}

	// Check password (in production, use bcrypt or similar)
	if foundUser.Password != req.Password {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Invalid username or password",
		})
	}

	// Create session
	sessionMgr := ctx.Session()
	if sessionMgr == nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Session manager not available",
		})
	}

	session, err := sessionMgr.Create(ctx)
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to create session",
		})
	}

	// Store user data in session
	session.Data["user_id"] = foundUser.ID
	session.Data["username"] = foundUser.Username

	// Save the session
	if err := sessionMgr.Save(ctx, session); err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to save session",
		})
	}

	// Set session cookie
	if err := sessionMgr.SetCookie(ctx, session); err != nil {
		fmt.Printf("Error setting session cookie: %v\n", err)
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set session cookie",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       foundUser.ID,
			"username": foundUser.Username,
			"email":    foundUser.Email,
		},
	})
}

func logoutHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()
	if sessionMgr != nil {
		// Get session from cookie
		session, err := sessionMgr.GetSessionFromCookie(ctx)
		if err == nil {
			// Destroy the session
			if err := sessionMgr.Destroy(ctx, session.ID); err != nil {
				fmt.Printf("Error destroying session: %v\n", err)
			}
		}
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Logout successful",
	})
}

func getCurrentUserHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()
	if sessionMgr == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Not authenticated",
		})
	}

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Not authenticated",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists || userID == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Not authenticated",
		})
	}

	// Find user
	id, ok := userID.(int)
	if !ok {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Invalid session data",
		})
	}

	user, exists := users[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "User not found",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// ============================================================================
// Handler Functions - Posts
// ============================================================================

func listPostsHandler(ctx pkg.Context) error {
	// Convert map to slice for JSON response
	postList := make([]*Post, 0, len(posts))
	for _, post := range posts {
		postList = append(postList, post)
	}

	return ctx.JSON(200, map[string]interface{}{
		"posts": postList,
		"total": len(postList),
	})
}

func getPostHandler(ctx pkg.Context) error {
	idStr := ctx.Params()["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid post ID",
		})
	}

	post, exists := posts[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Post not found",
		})
	}

	return ctx.JSON(200, post)
}

func createPostHandler(ctx pkg.Context) error {
	var req CreatePostRequest

	// Parse request body
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Title == "" || req.Content == "" {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Title and content are required",
		})
	}

	// Get user ID from session
	sessionMgr := ctx.Session()
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	authorID, ok := userID.(int)
	if !ok {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Invalid session data",
		})
	}

	// Get username for the post
	user, exists := users[authorID]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "User not found",
		})
	}

	// Create new post
	post := &Post{
		ID:        nextPostID,
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  authorID,
		Author:    user.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	posts[nextPostID] = post
	nextPostID++

	return ctx.JSON(201, post)
}

func updatePostHandler(ctx pkg.Context) error {
	idStr := ctx.Params()["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid post ID",
		})
	}

	post, exists := posts[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Post not found",
		})
	}

	// Check if user owns the post
	sessionMgr := ctx.Session()
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	authorID, ok := userID.(int)
	if !ok || post.AuthorID != authorID {
		return ctx.JSON(403, map[string]interface{}{
			"error": "You don't have permission to update this post",
		})
	}

	var req UpdatePostRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Update post
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	post.UpdatedAt = time.Now()

	return ctx.JSON(200, post)
}

func deletePostHandler(ctx pkg.Context) error {
	idStr := ctx.Params()["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid post ID",
		})
	}

	post, exists := posts[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Post not found",
		})
	}

	// Check if user owns the post
	sessionMgr := ctx.Session()
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	authorID, ok := userID.(int)
	if !ok || post.AuthorID != authorID {
		return ctx.JSON(403, map[string]interface{}{
			"error": "You don't have permission to delete this post",
		})
	}

	// Delete post and its comments
	delete(posts, id)
	for commentID, comment := range comments {
		if comment.PostID == id {
			delete(comments, commentID)
		}
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Post deleted successfully",
	})
}

// ============================================================================
// Handler Functions - Comments
// ============================================================================

func listCommentsHandler(ctx pkg.Context) error {
	postIDStr := ctx.Params()["id"]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid post ID",
		})
	}

	// Check if post exists
	if _, exists := posts[postID]; !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Post not found",
		})
	}

	// Get comments for this post
	commentList := make([]*Comment, 0)
	for _, comment := range comments {
		if comment.PostID == postID {
			commentList = append(commentList, comment)
		}
	}

	return ctx.JSON(200, map[string]interface{}{
		"comments": commentList,
		"total":    len(commentList),
	})
}

func createCommentHandler(ctx pkg.Context) error {
	postIDStr := ctx.Params()["id"]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid post ID",
		})
	}

	// Check if post exists
	if _, exists := posts[postID]; !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Post not found",
		})
	}

	var req CreateCommentRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Content == "" {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Content is required",
		})
	}

	// Get user ID from session
	sessionMgr := ctx.Session()
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	authorID, ok := userID.(int)
	if !ok {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Invalid session data",
		})
	}

	// Get username for the comment
	user, exists := users[authorID]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "User not found",
		})
	}

	// Create new comment
	comment := &Comment{
		ID:        nextCommentID,
		PostID:    postID,
		AuthorID:  authorID,
		Author:    user.Username,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	comments[nextCommentID] = comment
	nextCommentID++

	return ctx.JSON(201, comment)
}

func deleteCommentHandler(ctx pkg.Context) error {
	idStr := ctx.Params()["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid comment ID",
		})
	}

	comment, exists := comments[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Comment not found",
		})
	}

	// Check if user owns the comment
	sessionMgr := ctx.Session()
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	userID, exists := session.Data["user_id"]
	if !exists {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	authorID, ok := userID.(int)
	if !ok || comment.AuthorID != authorID {
		return ctx.JSON(403, map[string]interface{}{
			"error": "You don't have permission to delete this comment",
		})
	}

	delete(comments, id)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Comment deleted successfully",
	})
}

// ============================================================================
// Handler Functions - Users
// ============================================================================

func getUserHandler(ctx pkg.Context) error {
	idStr := ctx.Params()["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(400, map[string]interface{}{
			"error": "Invalid user ID",
		})
	}

	user, exists := users[id]
	if !exists {
		return ctx.JSON(404, map[string]interface{}{
			"error": "User not found",
		})
	}

	// Return user without password
	return ctx.JSON(200, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}
