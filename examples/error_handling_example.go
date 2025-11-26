//go:build ignore

package main

import (
	"errors"
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
			Database: "error_handling_example.db",
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
	// Custom Error Handler
	// ============================================================================
	// Create error handler with i18n support
	errorHandler := pkg.NewErrorHandler(pkg.ErrorHandlerConfig{
		I18n:         app.I18n(),
		Logger:       nil, // Would use actual logger in production
		IncludeStack: true,
		LogErrors:    true,
	})

	// Set custom error handler for the framework
	app.SetErrorHandler(func(ctx pkg.Context, err error) error {
		return errorHandler.HandleError(ctx, err)
	})

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Demonstrate different error types
	router.GET("/api/validation-error", validationErrorHandler)
	router.GET("/api/auth-error", authErrorHandler)
	router.GET("/api/not-found-error", notFoundErrorHandler)
	router.GET("/api/rate-limit-error", rateLimitErrorHandler)
	router.GET("/api/database-error", databaseErrorHandler)
	router.GET("/api/internal-error", internalErrorHandler)

	// Demonstrate internationalized errors
	router.GET("/api/i18n-error", i18nErrorHandler)

	// Demonstrate error logging and monitoring
	router.GET("/api/logged-error", loggedErrorHandler)

	// Demonstrate panic recovery
	router.GET("/api/panic", panicHandler)

	// Demonstrate custom error response formatting
	router.GET("/api/custom-format", customFormatHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Error Handling Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Error Handling Features Demonstrated:")
	fmt.Println("  ✓ Custom error handlers")
	fmt.Println("  ✓ Error response formatting")
	fmt.Println("  ✓ Internationalized error messages")
	fmt.Println("  ✓ Error logging and monitoring")
	fmt.Println("  ✓ Panic recovery")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Validation error")
	fmt.Println("  curl http://localhost:8080/api/validation-error")
	fmt.Println()
	fmt.Println("  # Authentication error")
	fmt.Println("  curl http://localhost:8080/api/auth-error")
	fmt.Println()
	fmt.Println("  # Not found error")
	fmt.Println("  curl http://localhost:8080/api/not-found-error")
	fmt.Println()
	fmt.Println("  # Rate limit error")
	fmt.Println("  curl http://localhost:8080/api/rate-limit-error")
	fmt.Println()
	fmt.Println("  # Database error")
	fmt.Println("  curl http://localhost:8080/api/database-error")
	fmt.Println()
	fmt.Println("  # Internal error")
	fmt.Println("  curl http://localhost:8080/api/internal-error")
	fmt.Println()
	fmt.Println("  # Internationalized error (English)")
	fmt.Println("  curl http://localhost:8080/api/i18n-error")
	fmt.Println()
	fmt.Println("  # Internationalized error (German)")
	fmt.Println("  curl -H 'Accept-Language: de' http://localhost:8080/api/i18n-error")
	fmt.Println()
	fmt.Println("  # Logged error (check console)")
	fmt.Println("  curl http://localhost:8080/api/logged-error")
	fmt.Println()
	fmt.Println("  # Panic recovery")
	fmt.Println("  curl http://localhost:8080/api/panic")
	fmt.Println()
	fmt.Println("  # Custom error format")
	fmt.Println("  curl http://localhost:8080/api/custom-format")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - Different Error Types
// ============================================================================

// validationErrorHandler demonstrates validation errors
func validationErrorHandler(ctx pkg.Context) error {
	// Create a validation error for missing email field
	err := pkg.NewValidationError("Invalid email format", "email")
	err.WithDetails(map[string]interface{}{
		"provided": "invalid-email",
		"expected": "user@example.com",
	})

	// Return the error - it will be handled by the custom error handler
	return err
}

// authErrorHandler demonstrates authentication errors
func authErrorHandler(ctx pkg.Context) error {
	// Create an authentication error
	err := pkg.NewAuthenticationError("Invalid credentials")
	err.WithContext(
		ctx.Request().ID,
		ctx.Request().TenantID,
		ctx.Request().UserID,
		ctx.Request().RequestURI,
		ctx.Request().Method,
	)

	return err
}

// notFoundErrorHandler demonstrates not found errors
func notFoundErrorHandler(ctx pkg.Context) error {
	// Create a not found error
	return pkg.NewNotFoundError("user")
}

// rateLimitErrorHandler demonstrates rate limit errors
func rateLimitErrorHandler(ctx pkg.Context) error {
	// Create a rate limit error
	return pkg.NewRateLimitError("Too many requests", 100, "1 minute")
}

// databaseErrorHandler demonstrates database errors
func databaseErrorHandler(ctx pkg.Context) error {
	// Simulate a database error
	dbErr := errors.New("connection timeout")

	// Create a database error and chain the cause
	err := pkg.NewDatabaseError("Failed to query database", "SELECT")
	err.WithCause(dbErr)

	return err
}

// internalErrorHandler demonstrates internal server errors
func internalErrorHandler(ctx pkg.Context) error {
	// Create an internal error
	err := pkg.NewInternalError("Unexpected error occurred")
	err.WithDetails(map[string]interface{}{
		"component": "payment-processor",
		"operation": "charge",
	})

	return err
}

// ============================================================================
// Handler Functions - Internationalized Errors
// ============================================================================

// i18nErrorHandler demonstrates internationalized error messages
func i18nErrorHandler(ctx pkg.Context) error {
	// Get i18n manager from context
	i18n := ctx.I18n()

	// Check Accept-Language header to determine language
	acceptLang := ctx.GetHeader("Accept-Language")
	if acceptLang == "de" {
		i18n.SetLanguage("de")
	} else {
		i18n.SetLanguage("en")
	}

	// Create an error with i18n support
	err := pkg.NewAuthenticationError("Invalid credentials")

	// The error message will be translated by the error handler
	return err
}

// ============================================================================
// Handler Functions - Error Logging and Monitoring
// ============================================================================

// loggedErrorHandler demonstrates error logging
func loggedErrorHandler(ctx pkg.Context) error {
	// Get logger from context
	logger := ctx.Logger()

	// Log some context before the error
	logger.Info("Processing request", "endpoint", "/api/logged-error")

	// Create an error
	err := pkg.NewInternalError("This error will be logged")
	err.WithContext(
		ctx.Request().ID,
		ctx.Request().TenantID,
		ctx.Request().UserID,
		ctx.Request().RequestURI,
		ctx.Request().Method,
	)

	// Log the error explicitly
	logger.Error("Error occurred", "code", err.Code, "message", err.Message)

	// Return the error - it will also be logged by the error handler
	return err
}

// ============================================================================
// Handler Functions - Panic Recovery
// ============================================================================

// panicHandler demonstrates panic recovery
func panicHandler(ctx pkg.Context) error {
	// This will cause a panic that should be recovered by the recovery middleware
	panic("simulated panic for demonstration")
}

// ============================================================================
// Handler Functions - Custom Error Formatting
// ============================================================================

// customFormatHandler demonstrates custom error response formatting
func customFormatHandler(ctx pkg.Context) error {
	// Create multiple errors to show in response
	errors := []map[string]interface{}{
		{
			"field":   "username",
			"message": "Username is required",
			"code":    "MISSING_FIELD",
		},
		{
			"field":   "email",
			"message": "Invalid email format",
			"code":    "INVALID_FORMAT",
		},
	}

	// Return a custom formatted error response
	return ctx.JSON(400, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "VALIDATION_FAILED",
			"message": "Multiple validation errors occurred",
			"errors":  errors,
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"path":      ctx.Request().RequestURI,
	})
}
