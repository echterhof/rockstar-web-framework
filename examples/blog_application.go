package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// Blog application demonstrating a real-world use case
// Features: Posts, Comments, Users, Authentication, Sessions

// Models
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  int       `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  int       `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type BlogUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// In-memory storage (replace with database in production)
var (
	posts          = make(map[int]*Post)
	comments       = make(map[int]*Comment)
	users          = make(map[int]*BlogUser)
	nextPostID     = 1
	nextCommentID  = 1
	nextBlogUserID = 1
)

func main() {
	// Initialize sample data
	initSampleData()

	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  2 << 20,
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableMetrics:   true,
			MetricsPath:     "/metrics",
			ShutdownTimeout: 30 * time.Second,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "blog.db",
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    50 * 1024 * 1024,
			DefaultTTL: 10 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "blog_session",
			SessionLifetime: 24 * time.Hour,
			CookieSecure:    false,
			CookieHTTPOnly:  true,
			EncryptionKey:   []byte("ABSNTNZMSNENLMONABSNTNZMSNENLMON"),
			CleanupInterval: 10 * time.Minute,
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			LocalesDir:       "./locales",
			SupportedLocales: []string{"en", "de"},
		},
		SecurityConfig: pkg.SecurityConfig{
			XFrameOptions:    "SAMEORIGIN",
			EnableCSRF:       true,
			EnableXSSProtect: true,
			MaxRequestSize:   10 * 1024 * 1024,
			RequestTimeout:   30 * time.Second,
		},
		MonitoringConfig: pkg.MonitoringConfig{
			EnableMetrics: true,
			MetricsPath:   "/metrics",
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Setup lifecycle hooks
	setupLifecycleHooks(app)

	// Setup middleware
	setupMiddleware(app)

	// Setup routes
	setupRoutes(app)

	// Setup graceful shutdown
	setupGracefulShutdown(app)

	// Start server
	fmt.Println("🎸 Rockstar Blog Application")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("Server: http://localhost:8080")
	fmt.Println("API Docs: http://localhost:8080/api/docs")
	fmt.Println("Health: http://localhost:8080/health")
	fmt.Println("=" + string(make([]byte, 50)))

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initSampleData() {
	// Create sample users
	users[1] = &BlogUser{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
	}
	users[2] = &BlogUser{
		ID:        2,
		Username:  "jane_smith",
		Email:     "jane@example.com",
		CreatedAt: time.Now(),
	}
	nextBlogUserID = 3

	// Create sample posts
	posts[1] = &Post{
		ID:        1,
		Title:     "Welcome to Rockstar Blog",
		Content:   "This is the first post on our new blog platform built with Rockstar Web Framework!",
		AuthorID:  1,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	posts[2] = &Post{
		ID:        2,
		Title:     "Getting Started with Go",
		Content:   "Go is an amazing language for building web applications. Here's why...",
		AuthorID:  2,
		CreatedAt: time.Now().Add(-12 * time.Hour),
		UpdatedAt: time.Now().Add(-12 * time.Hour),
	}
	nextPostID = 3

	// Create sample comments
	comments[1] = &Comment{
		ID:        1,
		PostID:    1,
		AuthorID:  2,
		Content:   "Great post! Looking forward to more content.",
		CreatedAt: time.Now().Add(-20 * time.Hour),
	}
	comments[2] = &Comment{
		ID:        2,
		PostID:    1,
		AuthorID:  1,
		Content:   "Thanks for the feedback!",
		CreatedAt: time.Now().Add(-19 * time.Hour),
	}
	nextCommentID = 3
}

func setupLifecycleHooks(app *pkg.Framework) {
	app.RegisterStartupHook(func(ctx context.Context) error {
		log.Println("✓ Blog application starting...")
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		log.Printf("✓ Loaded %d posts, %d comments, %d users", len(posts), len(comments), len(users))
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		log.Println("✓ Saving data...")
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		log.Println("✓ Blog application stopped")
		return nil
	})
}

func setupMiddleware(app *pkg.Framework) {
	// Request logging
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		start := time.Now()
		log.Printf("[%s] %s %s", ctx.Request().Method, ctx.Request().URL.Path, ctx.Request().RemoteAddr)
		err := next(ctx)
		log.Printf("  ⏱️  %v", time.Since(start))
		return err
	})

	// Recovery middleware
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ Panic: %v", r)
				ctx.JSON(500, map[string]interface{}{
					"error": "Internal server error",
				})
			}
		}()
		return next(ctx)
	})

	// CORS middleware
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.Request().Method == "OPTIONS" {
			return ctx.String(204, "")
		}

		return next(ctx)
	})
}

func setupRoutes(app *pkg.Framework) {
	router := app.Router()

	// Home
	router.GET("/", homeHandler)

	// Health check
	router.GET("/health", healthHandler)

	// API documentation
	router.GET("/api/docs", apiDocsHandler)

	// Public API routes
	api := router.Group("/api")

	// Posts
	api.GET("/posts", listPostsHandler)
	api.GET("/posts/:id", getPostHandler)

	// Comments
	api.GET("/posts/:id/comments", listCommentsHandler)

	// Users
	api.GET("/users/:id", getUserHandler)

	// Authenticated routes (simplified - no real auth for demo)
	authAPI := router.Group("/api", simpleAuthMiddleware)
	authAPI.POST("/posts", createPostHandler)
	authAPI.PUT("/posts/:id", updatePostHandler)
	authAPI.DELETE("/posts/:id", deletePostHandler)
	authAPI.POST("/posts/:id/comments", createCommentHandler)
	authAPI.DELETE("/comments/:id", deleteCommentHandler)
}

func setupGracefulShutdown(app *pkg.Framework) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Shutdown signal received")

		if err := app.Shutdown(30 * time.Second); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		os.Exit(0)
	}()
}

// Handlers

func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Welcome to Rockstar Blog! 🎸",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"posts":   "/api/posts",
			"users":   "/api/users/:id",
			"docs":    "/api/docs",
			"health":  "/health",
			"metrics": "/metrics",
		},
	})
}

func healthHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
		"stats": map[string]int{
			"posts":    len(posts),
			"comments": len(comments),
			"users":    len(users),
		},
	})
}

func apiDocsHandler(ctx pkg.Context) error {
	docs := map[string]interface{}{
		"title":   "Rockstar Blog API",
		"version": "1.0.0",
		"endpoints": []map[string]string{
			{"method": "GET", "path": "/api/posts", "description": "List all posts"},
			{"method": "GET", "path": "/api/posts/:id", "description": "Get a specific post"},
			{"method": "POST", "path": "/api/posts", "description": "Create a new post (auth required)"},
			{"method": "PUT", "path": "/api/posts/:id", "description": "Update a post (auth required)"},
			{"method": "DELETE", "path": "/api/posts/:id", "description": "Delete a post (auth required)"},
			{"method": "GET", "path": "/api/posts/:id/comments", "description": "List comments for a post"},
			{"method": "POST", "path": "/api/posts/:id/comments", "description": "Add a comment (auth required)"},
			{"method": "GET", "path": "/api/users/:id", "description": "Get user information"},
		},
	}
	return ctx.JSON(200, docs)
}

func listPostsHandler(ctx pkg.Context) error {
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
	id := ctx.Params()["id"]

	for _, post := range posts {
		if fmt.Sprintf("%d", post.ID) == id {
			return ctx.JSON(200, post)
		}
	}

	return ctx.JSON(404, map[string]string{
		"error": "Post not found",
	})
}

func createPostHandler(ctx pkg.Context) error {
	// In production, parse body properly
	post := &Post{
		ID:        nextPostID,
		Title:     "New Post",
		Content:   "Post content here...",
		AuthorID:  1, // From authenticated user
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	posts[nextPostID] = post
	nextPostID++

	return ctx.JSON(201, post)
}

func updatePostHandler(ctx pkg.Context) error {
	id := ctx.Params()["id"]

	for _, post := range posts {
		if fmt.Sprintf("%d", post.ID) == id {
			post.UpdatedAt = time.Now()
			// In production, update fields from request body
			return ctx.JSON(200, post)
		}
	}

	return ctx.JSON(404, map[string]string{
		"error": "Post not found",
	})
}

func deletePostHandler(ctx pkg.Context) error {
	id := ctx.Params()["id"]

	for postID, post := range posts {
		if fmt.Sprintf("%d", post.ID) == id {
			delete(posts, postID)
			return ctx.JSON(204, nil)
		}
	}

	return ctx.JSON(404, map[string]string{
		"error": "Post not found",
	})
}

func listCommentsHandler(ctx pkg.Context) error {
	postID := ctx.Params()["id"]

	commentList := make([]*Comment, 0)
	for _, comment := range comments {
		if fmt.Sprintf("%d", comment.PostID) == postID {
			commentList = append(commentList, comment)
		}
	}

	return ctx.JSON(200, map[string]interface{}{
		"comments": commentList,
		"total":    len(commentList),
	})
}

func createCommentHandler(ctx pkg.Context) error {
	postID := ctx.Params()["id"]

	comment := &Comment{
		ID:        nextCommentID,
		PostID:    0, // Parse from postID
		AuthorID:  1, // From authenticated user
		Content:   "Comment content here...",
		CreatedAt: time.Now(),
	}

	// Verify post exists
	found := false
	for _, post := range posts {
		if fmt.Sprintf("%d", post.ID) == postID {
			found = true
			comment.PostID = post.ID
			break
		}
	}

	if !found {
		return ctx.JSON(404, map[string]string{
			"error": "Post not found",
		})
	}

	comments[nextCommentID] = comment
	nextCommentID++

	return ctx.JSON(201, comment)
}

func deleteCommentHandler(ctx pkg.Context) error {
	id := ctx.Params()["id"]

	for commentID, comment := range comments {
		if fmt.Sprintf("%d", comment.ID) == id {
			delete(comments, commentID)
			return ctx.JSON(204, nil)
		}
	}

	return ctx.JSON(404, map[string]string{
		"error": "Comment not found",
	})
}

func getUserHandler(ctx pkg.Context) error {
	id := ctx.Params()["id"]

	for _, user := range users {
		if fmt.Sprintf("%d", user.ID) == id {
			return ctx.JSON(200, user)
		}
	}

	return ctx.JSON(404, map[string]string{
		"error": "User not found",
	})
}

// Middleware

func simpleAuthMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// Simplified auth check - in production use real authentication
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return ctx.JSON(401, map[string]string{
			"error": "Authentication required",
		})
	}

	// In production: validate token, load user, etc.

	return next(ctx)
}
