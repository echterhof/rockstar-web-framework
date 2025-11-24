package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1 MB
			EnableHTTP1:    true,
			EnableHTTP2:    true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Host:     "",
			Port:     0,
			Database: "rockstar.db",
			Username: "",
			Password: "",
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    100 * 1024 * 1024, // 100 MB
			DefaultTTL: 5 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "rockstar_session",
			SessionLifetime: 24 * time.Hour,
			CookieSecure:    false, // Set to true in production with HTTPS
			CookieHTTPOnly:  true,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        "./examples",
			SupportedLocales:  []string{"en", "de"},
			FallbackToDefault: true,
		},
		SecurityConfig: pkg.SecurityConfig{
			XFrameOptions:    "DENY",
			EnableXSSProtect: true,
			EnableCSRF:       true,
			MaxRequestSize:   10 * 1024 * 1024, // 10 MB
			RequestTimeout:   30 * time.Second,
			CSRFTokenExpiry:  24 * time.Hour,
			EncryptionKey:    "0123456789abcdef0123456789abcdef", // 32 hex chars for 16 bytes
			JWTSecret:        "your-secret-key-here",
			AllowedOrigins:   []string{"*"},
		},
		MonitoringConfig: pkg.MonitoringConfig{
			EnableMetrics: true,
			MetricsPath:   "/metrics",
			EnablePprof:   true,
			PprofPath:     "/debug/pprof",
		},
	}

	// Create new framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Register startup hook
	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Rockstar Web Framework starting up...")
		return nil
	})

	// Register shutdown hook
	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Rockstar Web Framework shutting down...")
		return nil
	})

	// Add global middleware
	app.Use(loggingMiddleware)
	app.Use(recoveryMiddleware)

	// Set custom error handler
	app.SetErrorHandler(customErrorHandler)

	// Get router
	router := app.Router()

	// Define routes
	router.GET("/", homeHandler)
	router.GET("/hello/:name", helloHandler)
	router.POST("/api/users", createUserHandler)
	router.GET("/api/users/:id", getUserHandler)

	// API group with authentication middleware
	api := router.Group("/api/v1", authMiddleware)
	api.GET("/profile", profileHandler)
	api.POST("/logout", logoutHandler)

	// Static files
	router.Static("/static", pkg.NewOSFileSystem("./static"))

	// Start server
	fmt.Println("🎸 Starting Rockstar Web Framework on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Middleware examples

func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	start := time.Now()

	// Log request
	fmt.Printf("[%s] %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		ctx.Request().Method,
		ctx.Request().RequestURI,
	)

	// Call next handler
	err := next(ctx)

	// Log response time
	duration := time.Since(start)
	fmt.Printf("  ⏱️  Completed in %v\n", duration)

	return err
}

func recoveryMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Panic recovered: %v\n", r)
			ctx.JSON(500, map[string]interface{}{
				"error": "Internal server error",
			})
		}
	}()

	return next(ctx)
}

func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	// Check for authentication token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Unauthorized",
		})
	}

	// Validate token (simplified example)
	// In production, use ctx.Security().AuthenticateJWT(token)

	return next(ctx)
}

// Handler examples

func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Welcome to Rockstar Web Framework! 🎸",
		"version": "1.0.0",
		"docs":    "https://github.com/echterhof/rockstar-web-framework",
	})
}

func helloHandler(ctx pkg.Context) error {
	name := ctx.Params()["name"]
	return ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s! 👋", name),
	})
}

func createUserHandler(ctx pkg.Context) error {
	// Parse request body
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// In production, parse body properly
	// For now, return success

	return ctx.JSON(201, map[string]interface{}{
		"message": "User created successfully",
		"user":    user,
	})
}

func getUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]

	// In production, fetch from database
	// user, err := ctx.DB().FindUserByID(userID)

	return ctx.JSON(200, map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	})
}

func profileHandler(ctx pkg.Context) error {
	// Get authenticated user
	user := ctx.User()
	if user == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Not authenticated",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"user": user,
	})
}

func logoutHandler(ctx pkg.Context) error {
	// Get session manager and destroy session
	sessionMgr := ctx.Session()
	if sessionMgr != nil {
		// Get session from cookie
		session, err := sessionMgr.GetSessionFromCookie(ctx)
		if err == nil && session != nil {
			sessionMgr.Destroy(ctx, session.ID)
		}
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

func customErrorHandler(ctx pkg.Context, err error) error {
	fmt.Printf("❌ Error: %v\n", err)

	return ctx.JSON(500, map[string]interface{}{
		"error": err.Error(),
	})
}
