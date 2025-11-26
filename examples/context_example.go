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
			Database: "context_example.db",
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    10 * 1024 * 1024, // 10 MB
			DefaultTTL: 5 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache, // Use cache instead of database for simplicity
			CookieName:      "rockstar_session",
			SessionLifetime: 24 * time.Hour,
			CleanupInterval: 15 * time.Minute, // Cleanup expired sessions every 15 minutes
			CookieSecure:    false,
			CookieHTTPOnly:  true,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        "./examples/locales",
			SupportedLocales:  []string{"en", "de"},
			FallbackToDefault: true,
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
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Demonstrate request data access
	router.GET("/api/users/:id", getUserHandler)

	// Demonstrate response methods
	router.GET("/api/json", jsonResponseHandler)
	router.GET("/api/xml", xmlResponseHandler)
	router.GET("/api/html", htmlResponseHandler)
	router.GET("/api/string", stringResponseHandler)

	// Demonstrate service access
	router.GET("/api/services", servicesHandler)

	// Demonstrate context control
	router.GET("/api/timeout", timeoutHandler)
	router.GET("/api/cancel", cancelHandler)

	// Demonstrate cookie and header management
	router.GET("/api/cookies", cookieHandler)
	router.GET("/api/headers", headerHandler)

	// Demonstrate form handling
	router.POST("/api/form", formHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Context Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Request data access (params, query, headers)")
	fmt.Println("  curl http://localhost:8080/api/users/123?sort=name&filter=active")
	fmt.Println()
	fmt.Println("  # Response methods")
	fmt.Println("  curl http://localhost:8080/api/json")
	fmt.Println("  curl http://localhost:8080/api/xml")
	fmt.Println("  curl http://localhost:8080/api/html")
	fmt.Println("  curl http://localhost:8080/api/string")
	fmt.Println()
	fmt.Println("  # Service access (DB, Cache, Session, Config, I18n)")
	fmt.Println("  curl http://localhost:8080/api/services")
	fmt.Println()
	fmt.Println("  # Context control (timeout, cancel)")
	fmt.Println("  curl http://localhost:8080/api/timeout")
	fmt.Println("  curl http://localhost:8080/api/cancel")
	fmt.Println()
	fmt.Println("  # Cookie and header management")
	fmt.Println("  curl -v http://localhost:8080/api/cookies")
	fmt.Println("  curl -v http://localhost:8080/api/headers")
	fmt.Println()
	fmt.Println("  # Form handling")
	fmt.Println("  curl -X POST http://localhost:8080/api/form -d 'username=john&email=john@example.com'")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - Request Data Access
// ============================================================================

// getUserHandler demonstrates accessing request data: params, query, headers, body
func getUserHandler(ctx pkg.Context) error {
	// Access route parameters
	userID := ctx.Params()["id"]

	// Access query parameters
	query := ctx.Query()
	sort := query["sort"]
	filter := query["filter"]

	// Access headers
	headers := ctx.Headers()
	contentType := ctx.GetHeader("Content-Type")
	userAgent := ctx.GetHeader("User-Agent")

	// Access request body (if any)
	body := ctx.Body()

	return ctx.JSON(200, map[string]interface{}{
		"message": "Request data accessed successfully",
		"data": map[string]interface{}{
			"params": map[string]string{
				"id": userID,
			},
			"query": map[string]string{
				"sort":   sort,
				"filter": filter,
			},
			"headers": map[string]string{
				"content_type": contentType,
				"user_agent":   userAgent,
				"all_headers":  fmt.Sprintf("%d headers received", len(headers)),
			},
			"body_length": len(body),
		},
	})
}

// ============================================================================
// Handler Functions - Response Methods
// ============================================================================

// jsonResponseHandler demonstrates JSON response
func jsonResponseHandler(ctx pkg.Context) error {
	data := map[string]interface{}{
		"message": "This is a JSON response",
		"status":  "success",
		"data": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
		},
	}
	return ctx.JSON(200, data)
}

// xmlResponseHandler demonstrates XML response
func xmlResponseHandler(ctx pkg.Context) error {
	type User struct {
		ID   int    `xml:"id"`
		Name string `xml:"name"`
	}
	user := User{ID: 123, Name: "John Doe"}
	return ctx.XML(200, user)
}

// htmlResponseHandler demonstrates HTML response
func htmlResponseHandler(ctx pkg.Context) error {
	html := `
<!DOCTYPE html>
<html>
<head><title>Context Example</title></head>
<body>
	<h1>HTML Response from Context</h1>
	<p>This demonstrates the HTML response method.</p>
</body>
</html>
`
	return ctx.HTML(200, "inline", html)
}

// stringResponseHandler demonstrates String response
func stringResponseHandler(ctx pkg.Context) error {
	return ctx.String(200, "This is a plain text response from the Context")
}

// ============================================================================
// Handler Functions - Service Access
// ============================================================================

// servicesHandler demonstrates accessing framework services
func servicesHandler(ctx pkg.Context) error {
	// Access database manager
	db := ctx.DB()
	dbConnected := db != nil

	// Access cache manager
	cache := ctx.Cache()
	cacheAvailable := cache != nil

	// Access session manager
	session := ctx.Session()
	sessionAvailable := session != nil

	// Access config manager
	config := ctx.Config()
	configAvailable := config != nil

	// Access i18n manager
	i18n := ctx.I18n()
	i18nAvailable := i18n != nil

	// Access logger
	logger := ctx.Logger()
	if logger != nil {
		logger.Info("Services accessed", "endpoint", "/api/services")
	}

	// Access metrics collector
	metrics := ctx.Metrics()
	metricsAvailable := metrics != nil

	return ctx.JSON(200, map[string]interface{}{
		"message": "All services accessed successfully",
		"services": map[string]bool{
			"database": dbConnected,
			"cache":    cacheAvailable,
			"session":  sessionAvailable,
			"config":   configAvailable,
			"i18n":     i18nAvailable,
			"logger":   logger != nil,
			"metrics":  metricsAvailable,
		},
	})
}

// ============================================================================
// Handler Functions - Context Control
// ============================================================================

// timeoutHandler demonstrates WithTimeout context control
func timeoutHandler(ctx pkg.Context) error {
	// Create a context with timeout
	timeoutCtx := ctx.WithTimeout(5 * time.Second)

	// Check if context has deadline
	deadline, hasDeadline := timeoutCtx.Context().Deadline()

	return ctx.JSON(200, map[string]interface{}{
		"message":      "Timeout context created",
		"has_deadline": hasDeadline,
		"deadline":     deadline.Format(time.RFC3339),
		"timeout":      "5 seconds",
	})
}

// cancelHandler demonstrates WithCancel context control
func cancelHandler(ctx pkg.Context) error {
	// Create a cancellable context
	cancelCtx, cancel := ctx.WithCancel()

	// Check context before cancellation
	beforeCancel := cancelCtx.Context().Err() == nil

	// Cancel the context
	cancel()

	// Check context after cancellation
	afterCancel := cancelCtx.Context().Err() != nil

	return ctx.JSON(200, map[string]interface{}{
		"message":       "Cancel context demonstrated",
		"before_cancel": beforeCancel,
		"after_cancel":  afterCancel,
	})
}

// ============================================================================
// Handler Functions - Cookie and Header Management
// ============================================================================

// cookieHandler demonstrates cookie management
func cookieHandler(ctx pkg.Context) error {
	// Set a cookie
	cookie := &pkg.Cookie{
		Name:     "session_id",
		Value:    "abc123xyz",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
	}
	ctx.SetCookie(cookie)

	// Try to get a cookie (if sent by client)
	existingCookie, err := ctx.GetCookie("session_id")
	var cookieValue string
	if err == nil && existingCookie != nil {
		cookieValue = existingCookie.Value
	} else {
		cookieValue = "not found (will be set in response)"
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":        "Cookie operations demonstrated",
		"cookie_set":     "session_id=abc123xyz",
		"cookie_read":    cookieValue,
		"check_response": "Check Set-Cookie header in response",
	})
}

// headerHandler demonstrates header management
func headerHandler(ctx pkg.Context) error {
	// Set custom headers
	ctx.SetHeader("X-Custom-Header", "CustomValue")
	ctx.SetHeader("X-Request-ID", "req-12345")
	ctx.SetHeader("X-API-Version", "1.0")

	// Get headers from request
	userAgent := ctx.GetHeader("User-Agent")
	acceptHeader := ctx.GetHeader("Accept")

	return ctx.JSON(200, map[string]interface{}{
		"message": "Header operations demonstrated",
		"request_headers": map[string]string{
			"user_agent": userAgent,
			"accept":     acceptHeader,
		},
		"response_headers": map[string]string{
			"X-Custom-Header": "CustomValue",
			"X-Request-ID":    "req-12345",
			"X-API-Version":   "1.0",
		},
		"check_response": "Check response headers with curl -v",
	})
}

// ============================================================================
// Handler Functions - Form Handling
// ============================================================================

// formHandler demonstrates form data handling
func formHandler(ctx pkg.Context) error {
	// Access form values
	username := ctx.FormValue("username")
	email := ctx.FormValue("email")

	// Validate form data
	if username == "" || email == "" {
		return ctx.JSON(400, map[string]interface{}{
			"error":   "Missing required fields",
			"message": "Both username and email are required",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Form data processed successfully",
		"data": map[string]string{
			"username": username,
			"email":    email,
		},
	})
}
