//go:build ignore

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	// ============================================================================
	// Security: Load Encryption Key from Environment
	// ============================================================================
	// IMPORTANT: Never hardcode encryption keys in your code!
	// Generate a secure key using: go run examples/generate_keys.go
	// Then set it as an environment variable: export SESSION_ENCRYPTION_KEY=your_key_here

	encryptionKeyHex := os.Getenv("SESSION_ENCRYPTION_KEY")
	if encryptionKeyHex == "" {
		log.Println("WARNING: SESSION_ENCRYPTION_KEY not set, generating temporary key")
		log.Println("This is INSECURE for production! Generate a key with: go run examples/generate_keys.go")

		// Generate temporary key for development only
		tempKey, err := pkg.GenerateEncryptionKeyHex(32)
		if err != nil {
			log.Fatalf("Failed to generate temporary key: %v", err)
		}
		encryptionKeyHex = tempKey
		log.Printf("Temporary key (save this): %s", tempKey)
	}

	// Decode hex key to bytes
	encryptionKey, err := hex.DecodeString(encryptionKeyHex)
	if err != nil {
		log.Fatalf("Invalid encryption key format (must be hex): %v", err)
	}

	if len(encryptionKey) != 32 {
		log.Fatalf("Encryption key must be 32 bytes for AES-256, got %d bytes", len(encryptionKey))
	}

	// ============================================================================
	// Configuration Setup
	// ============================================================================
	// Create framework configuration with minimal settings.
	// Most values use sensible defaults - only specify what you need to change!
	config := pkg.FrameworkConfig{
		// Server configuration - most values use defaults
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1: true, // Enable HTTP/1.1 support
			EnableHTTP2: true, // Enable HTTP/2 support
			// ReadTimeout defaults to 30s
			// WriteTimeout defaults to 30s
			// IdleTimeout defaults to 120s
			// MaxHeaderBytes defaults to 1MB
			// MaxConnections defaults to 10000
			// MaxRequestSize defaults to 10MB
			// ShutdownTimeout defaults to 30s
			// ReadBufferSize defaults to 4096
			// WriteBufferSize defaults to 4096
		},
		// Database configuration - only required fields specified
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "getting_started.db",
			// Host defaults to "localhost"
			// Port defaults to 0 for sqlite (driver-specific: postgres=5432, mysql=3306, mssql=1433)
			// MaxOpenConns defaults to 25
			// MaxIdleConns defaults to 5
			// ConnMaxLifetime defaults to 5 minutes
		},
		// Cache configuration - all defaults work great for most use cases
		CacheConfig: pkg.CacheConfig{
			// Type defaults to "memory"
			// MaxSize defaults to 0 (unlimited)
			// DefaultTTL defaults to 0 (no expiration)
		},
		// Session configuration - encryption key loaded from environment
		SessionConfig: pkg.SessionConfig{
			EncryptionKey: encryptionKey, // 32 bytes for AES-256 (loaded from env)
			CookieSecure:  false,         // Set to true in production with HTTPS
			// CookieName defaults to "rockstar_session"
			// CookiePath defaults to "/"
			// SessionLifetime defaults to 24 hours
			// CleanupInterval defaults to 1 hour
			// CookieHTTPOnly defaults to true
			// CookieSameSite defaults to "Lax"
			// FilesystemPath defaults to "./sessions"
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	// Create new framework instance with the configuration
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ============================================================================
	// Lifecycle Hooks
	// ============================================================================
	// Register startup hook - called when the server starts
	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Server starting up...")
		fmt.Println("   Initializing database connections...")
		fmt.Println("   Loading configuration...")
		fmt.Println("   Ready to accept requests!")
		return nil
	})

	// Register shutdown hook - called during graceful shutdown
	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Server shutting down...")
		fmt.Println("   Closing database connections...")
		fmt.Println("   Cleaning up resources...")
		fmt.Println("   Goodbye!")
		return nil
	})

	// ============================================================================
	// Global Middleware
	// ============================================================================
	// Add logging middleware - logs all requests
	app.Use(loggingMiddleware)

	// Add recovery middleware - recovers from panics
	app.Use(recoveryMiddleware)

	// ============================================================================
	// Custom Error Handler
	// ============================================================================
	// Set custom error handler for all errors
	app.SetErrorHandler(customErrorHandler)

	// ============================================================================
	// Route Registration
	// ============================================================================
	// Get the router instance
	router := app.Router()

	// GET route - simple welcome endpoint
	router.GET("/", homeHandler)

	// GET route with path parameter - demonstrates URL parameters
	router.GET("/hello/:name", helloHandler)

	// POST route - demonstrates creating resources
	router.POST("/api/users", createUserHandler)

	// GET route with parameter - demonstrates retrieving resources
	router.GET("/api/users/:id", getUserHandler)

	// PUT route - demonstrates updating resources
	router.PUT("/api/users/:id", updateUserHandler)

	// DELETE route - demonstrates deleting resources
	router.DELETE("/api/users/:id", deleteUserHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	// Print startup information
	fmt.Println("🎸 Rockstar Web Framework - Getting Started Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/hello/World")
	fmt.Println("  curl -X POST http://localhost:8080/api/users -d '{\"name\":\"John\",\"email\":\"john@example.com\"}'")
	fmt.Println("  curl http://localhost:8080/api/users/123")
	fmt.Println("  curl -X PUT http://localhost:8080/api/users/123 -d '{\"name\":\"Jane\",\"email\":\"jane@example.com\"}'")
	fmt.Println("  curl -X DELETE http://localhost:8080/api/users/123")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Start the server - this blocks until shutdown
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Middleware Functions
// ============================================================================

// loggingMiddleware logs all incoming requests and their response times
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()

	// Log request details
	fmt.Printf("[%s] %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		ctx.Request().Method,
		ctx.Request().RequestURI,
	)

	// Call the next handler in the chain
	err := next(ctx)

	// Log response time
	duration := time.Since(start)
	fmt.Printf("  ⏱️  Completed in %v\n", duration)

	return err
}

// recoveryMiddleware recovers from panics and returns a 500 error
func recoveryMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Panic recovered: %v\n", r)
			// Return internal server error
			ctx.JSON(500, map[string]interface{}{
				"error":   "Internal server error",
				"message": "An unexpected error occurred",
			})
		}
	}()

	return next(ctx)
}

// ============================================================================
// Handler Functions
// ============================================================================

// homeHandler handles the root endpoint
func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Welcome to Rockstar Web Framework! 🎸",
		"version": "1.0.0",
		"docs":    "https://github.com/echterhof/rockstar-web-framework",
		"endpoints": []string{
			"GET /",
			"GET /hello/:name",
			"POST /api/users",
			"GET /api/users/:id",
			"PUT /api/users/:id",
			"DELETE /api/users/:id",
		},
	})
}

// helloHandler demonstrates path parameters
func helloHandler(ctx pkg.Context) error {
	// Extract the 'name' parameter from the URL path
	name := ctx.Params()["name"]

	return ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s! 👋", name),
		"path":    ctx.Request().RequestURI,
	})
}

// createUserHandler demonstrates POST request handling
func createUserHandler(ctx pkg.Context) error {
	// In a real application, you would:
	// 1. Parse the request body
	// 2. Validate the input
	// 3. Save to database using ctx.DB()
	// 4. Return the created resource

	return ctx.JSON(201, map[string]interface{}{
		"message": "User created successfully",
		"id":      "123",
		"note":    "In production, parse body and save to database",
	})
}

// getUserHandler demonstrates GET request with parameters
func getUserHandler(ctx pkg.Context) error {
	// Extract the user ID from the URL path
	userID := ctx.Params()["id"]

	// In a real application, you would:
	// 1. Query the database using ctx.DB()
	// 2. Handle not found errors
	// 3. Return the user data

	return ctx.JSON(200, map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
		"note":  "In production, fetch from database",
	})
}

// updateUserHandler demonstrates PUT request handling
func updateUserHandler(ctx pkg.Context) error {
	// Extract the user ID from the URL path
	userID := ctx.Params()["id"]

	// In a real application, you would:
	// 1. Parse the request body
	// 2. Validate the input
	// 3. Update in database using ctx.DB()
	// 4. Return the updated resource

	return ctx.JSON(200, map[string]interface{}{
		"message": "User updated successfully",
		"id":      userID,
		"note":    "In production, parse body and update in database",
	})
}

// deleteUserHandler demonstrates DELETE request handling
func deleteUserHandler(ctx pkg.Context) error {
	// Extract the user ID from the URL path
	userID := ctx.Params()["id"]

	// In a real application, you would:
	// 1. Check if the user exists
	// 2. Delete from database using ctx.DB()
	// 3. Return success confirmation

	return ctx.JSON(200, map[string]interface{}{
		"message": "User deleted successfully",
		"id":      userID,
		"note":    "In production, delete from database",
	})
}

// ============================================================================
// Error Handler
// ============================================================================

// customErrorHandler handles all errors that occur during request processing
func customErrorHandler(ctx pkg.Context, err error) error {
	// Log the error
	fmt.Printf("❌ Error: %v\n", err)

	// Return error response
	return ctx.JSON(500, map[string]interface{}{
		"error":   "Internal server error",
		"message": err.Error(),
	})
}
