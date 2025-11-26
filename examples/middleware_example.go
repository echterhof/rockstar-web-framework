//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite3",
			Database: "middleware_example.db",
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    10 * 1024 * 1024,
			DefaultTTL: 5 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache, // Use cache instead of database
			CookieName:      "rockstar_session",
			SessionLifetime: 24 * time.Hour,
			CleanupInterval: 15 * time.Minute,
			CookieSecure:    false,
			CookieHTTPOnly:  true,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ============================================================================
	// Global Middleware Registration
	// ============================================================================
	// Global middleware applies to all routes
	// Middleware executes in the order they are registered

	// 1. Logging middleware - logs all requests
	app.Use(loggingMiddleware)

	// 2. Recovery middleware - recovers from panics
	app.Use(recoveryMiddleware)

	// 3. CORS middleware - adds CORS headers
	app.Use(corsMiddleware)

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Route without additional middleware - uses only global middleware
	router.GET("/api/public", publicHandler)

	// Route with single middleware - authentication required
	router.GET("/api/protected", protectedHandler, authMiddleware)

	// Route with multiple middleware - authentication + rate limiting
	router.GET("/api/admin", adminHandler, authMiddleware, rateLimitMiddleware)

	// ============================================================================
	// Route Groups with Shared Middleware
	// ============================================================================
	// Create a route group for API v1 with shared middleware
	apiV1 := router.Group("/api/v1", apiVersionMiddleware("v1"))

	// All routes in this group will have the apiVersionMiddleware applied
	apiV1.GET("/users", usersHandler)
	apiV1.GET("/posts", postsHandler)

	// Create a route group for admin routes with authentication
	adminGroup := router.Group("/admin", authMiddleware, adminCheckMiddleware)

	// All routes in this group require authentication and admin privileges
	adminGroup.GET("/dashboard", dashboardHandler)
	adminGroup.GET("/settings", settingsHandler)

	// ============================================================================
	// Middleware Chaining Example
	// ============================================================================
	// Demonstrate chaining multiple middleware functions
	chainedMw := pkg.ChainMiddleware(
		timingMiddleware,
		validationMiddleware,
	)

	router.POST("/api/data", dataHandler, chainedMw)

	// ============================================================================
	// Conditional Middleware Example
	// ============================================================================
	// Skip authentication for health check endpoint
	skipAuthMw := pkg.SkipMiddleware(
		func(ctx pkg.Context) bool {
			return ctx.Request().URL.Path == "/health"
		},
		authMiddleware,
	)

	router.GET("/health", healthHandler, skipAuthMw)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Middleware Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Middleware Patterns Demonstrated:")
	fmt.Println("  ✓ Global middleware (logging, recovery, CORS)")
	fmt.Println("  ✓ Route-specific middleware (auth, rate limiting)")
	fmt.Println("  ✓ Middleware groups (API versioning, admin routes)")
	fmt.Println("  ✓ Middleware chaining (timing + validation)")
	fmt.Println("  ✓ Conditional middleware (skip auth for health)")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Public endpoint (global middleware only)")
	fmt.Println("  curl http://localhost:8080/api/public")
	fmt.Println()
	fmt.Println("  # Protected endpoint (requires auth)")
	fmt.Println("  curl http://localhost:8080/api/protected")
	fmt.Println("  curl -H 'Authorization: Bearer token123' http://localhost:8080/api/protected")
	fmt.Println()
	fmt.Println("  # Admin endpoint (requires auth + rate limit)")
	fmt.Println("  curl -H 'Authorization: Bearer token123' http://localhost:8080/api/admin")
	fmt.Println()
	fmt.Println("  # Route group endpoints")
	fmt.Println("  curl http://localhost:8080/api/v1/users")
	fmt.Println("  curl -H 'Authorization: Bearer token123' http://localhost:8080/admin/dashboard")
	fmt.Println()
	fmt.Println("  # Chained middleware")
	fmt.Println("  curl -X POST http://localhost:8080/api/data -d '{\"value\":\"test\"}'")
	fmt.Println()
	fmt.Println("  # Health check (skips auth)")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Global Middleware Functions
// ============================================================================

// loggingMiddleware logs all incoming requests
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()

	fmt.Printf("[%s] %s %s\n",
		time.Now().Format("15:04:05"),
		ctx.Request().Method,
		ctx.Request().URL.Path,
	)

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

// corsMiddleware adds CORS headers to responses
func corsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// Set CORS headers
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if ctx.Request().Method == "OPTIONS" {
		return ctx.String(204, "")
	}

	return next(ctx)
}

// ============================================================================
// Route-Specific Middleware Functions
// ============================================================================

// authMiddleware checks for authentication
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		return ctx.JSON(401, map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Authorization header required",
		})
	}

	// In production, validate the token here
	fmt.Println("  🔐 Authentication successful")

	return next(ctx)
}

// rateLimitMiddleware implements simple rate limiting
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// In production, implement actual rate limiting logic
	fmt.Println("  ⏳ Rate limit check passed")

	return next(ctx)
}

// adminCheckMiddleware verifies admin privileges
func adminCheckMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// In production, check user roles/permissions
	fmt.Println("  👑 Admin privileges verified")

	return next(ctx)
}

// apiVersionMiddleware adds API version information
func apiVersionMiddleware(version string) pkg.MiddlewareFunc {
	return func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-API-Version", version)
		fmt.Printf("  📌 API Version: %s\n", version)
		return next(ctx)
	}
}

// timingMiddleware measures handler execution time
func timingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()
	err := next(ctx)
	duration := time.Since(start)

	ctx.SetHeader("X-Response-Time", duration.String())
	fmt.Printf("  ⏱️  Handler execution: %v\n", duration)

	return err
}

// validationMiddleware validates request data
func validationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// In production, implement actual validation logic
	fmt.Println("  ✅ Request validation passed")

	return next(ctx)
}

// ============================================================================
// Handler Functions
// ============================================================================

// publicHandler handles public endpoint
func publicHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Public endpoint - no authentication required",
		"middleware": []string{"logging", "recovery", "cors"},
	})
}

// protectedHandler handles protected endpoint
func protectedHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Protected endpoint - authentication required",
		"middleware": []string{"logging", "recovery", "cors", "auth"},
	})
}

// adminHandler handles admin endpoint
func adminHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Admin endpoint - authentication and rate limiting applied",
		"middleware": []string{"logging", "recovery", "cors", "auth", "rateLimit"},
	})
}

// usersHandler handles users endpoint in API v1 group
func usersHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Users endpoint in API v1",
		"middleware": []string{"logging", "recovery", "cors", "apiVersion"},
	})
}

// postsHandler handles posts endpoint in API v1 group
func postsHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Posts endpoint in API v1",
		"middleware": []string{"logging", "recovery", "cors", "apiVersion"},
	})
}

// dashboardHandler handles admin dashboard
func dashboardHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Admin dashboard",
		"middleware": []string{"logging", "recovery", "cors", "auth", "adminCheck"},
	})
}

// settingsHandler handles admin settings
func settingsHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Admin settings",
		"middleware": []string{"logging", "recovery", "cors", "auth", "adminCheck"},
	})
}

// dataHandler handles data endpoint with chained middleware
func dataHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message":    "Data endpoint with chained middleware",
		"middleware": []string{"logging", "recovery", "cors", "timing", "validation"},
	})
}

// healthHandler handles health check endpoint
func healthHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status":     "healthy",
		"message":    "Health check endpoint - auth skipped",
		"middleware": []string{"logging", "recovery", "cors"},
	})
}
