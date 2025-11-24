package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Internationalization Example ===\n")

	// Example 1: Basic i18n setup
	fmt.Println("1. Basic I18n Setup")
	basicExample()

	// Example 2: Loading locale files
	fmt.Println("\n2. Loading Locale Files")
	localeFileExample()

	// Example 3: Translation with parameters
	fmt.Println("\n3. Translation with Parameters")
	parameterExample()

	// Example 4: Internationalized logging
	fmt.Println("\n4. Internationalized Logging")
	loggingExample()

	// Example 5: Error messages with i18n
	fmt.Println("\n5. Error Messages with I18n")
	errorExample()

	// Example 6: Multiple locales
	fmt.Println("\n6. Multiple Locales")
	multiLocaleExample()
}

func basicExample() {
	// Create i18n manager with basic configuration
	config := pkg.I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add some translations manually
	manager.LoadLocale("en", map[string]interface{}{
		"greeting": "Hello, World!",
		"farewell": "Goodbye!",
	})

	// Translate keys
	fmt.Printf("  English greeting: %s\n", manager.Translate("greeting"))
	fmt.Printf("  English farewell: %s\n", manager.Translate("farewell"))
}

func localeFileExample() {
	// Create i18n manager that loads locale files from a directory
	config := pkg.I18nConfig{
		DefaultLocale:     "en",
		LocalesDir:        "./examples", // Directory containing locales.en.yaml, locales.de.yaml
		FallbackToDefault: true,
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Printf("  Note: Could not load locale files: %v\n", err)
		log.Println("  Make sure locales.en.yaml and locales.de.yaml exist in ./examples/")
		return
	}

	// Get available locales
	locales := manager.GetSupportedLanguages()
	fmt.Printf("  Available locales: %v\n", locales)

	// Translate error messages
	fmt.Printf("  Auth error (EN): %s\n", manager.Translate("error.authentication.failed"))

	// Switch to German
	if err := manager.SetLanguage("de"); err == nil {
		fmt.Printf("  Auth error (DE): %s\n", manager.Translate("error.authentication.failed"))
	}
}

func parameterExample() {
	config := pkg.I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translations with placeholders
	manager.LoadLocale("en", map[string]interface{}{
		"greeting": map[string]interface{}{
			"user": "Hello, {{name}}! Welcome back.",
		},
		"rate": map[string]interface{}{
			"limit": "Rate limit exceeded: {{limit}} requests per {{window}}",
		},
	})

	// Translate with parameters
	greeting := manager.Translate("greeting.user", map[string]interface{}{
		"name": "Alice",
	})
	fmt.Printf("  Personalized greeting: %s\n", greeting)

	rateLimit := manager.Translate("rate.limit", map[string]interface{}{
		"limit":  100,
		"window": "minute",
	})
	fmt.Printf("  Rate limit message: %s\n", rateLimit)
}

func loggingExample() {
	config := pkg.I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add log message translations
	manager.LoadLocale("en", map[string]interface{}{
		"log": map[string]interface{}{
			"server": map[string]interface{}{
				"starting": "Server starting on {{address}}",
			},
			"request": map[string]interface{}{
				"completed": "Request completed: {{method}} {{path}} - {{status}}",
			},
		},
	})

	// Get internationalized logger
	logger := pkg.NewI18nLogger(manager, nil)

	// Log messages with i18n
	fmt.Println("  Logging with i18n:")
	logger.Info("log.server.starting", map[string]interface{}{
		"address": "localhost:8080",
	})

	logger.Info("log.request.completed", map[string]interface{}{
		"method": "GET",
		"path":   "/api/users",
		"status": 200,
	})

	// Log with additional attributes
	attrs := []slog.Attr{
		slog.String("request_id", "req-12345"),
		slog.String("tenant_id", "tenant-abc"),
	}
	logger.InfoWithAttrs("log.request.completed", attrs, map[string]interface{}{
		"method": "POST",
		"path":   "/api/users",
		"status": 201,
	})
}

func errorExample() {
	config := pkg.I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add error message translations
	manager.LoadLocale("en", map[string]interface{}{
		"error": map[string]interface{}{
			"validation": map[string]interface{}{
				"missing_field": "Required field '{{field}}' is missing",
			},
			"database": map[string]interface{}{
				"query": "Database query failed: {{operation}}",
			},
		},
	})

	// Create framework errors with i18n
	validationError := pkg.NewValidationError("Validation failed", "email")
	validationError.I18nKey = "error.validation.missing_field"
	validationError.I18nParams = map[string]interface{}{"field": "email"}

	// Translate error message
	translatedMsg := manager.Translate(validationError.I18nKey, validationError.I18nParams)
	fmt.Printf("  Validation error: %s\n", translatedMsg)

	// Database error example
	dbError := pkg.NewDatabaseError("Query failed", "SELECT")
	dbError.I18nKey = "error.database.query"
	dbError.I18nParams = map[string]interface{}{"operation": "SELECT"}

	translatedMsg = manager.Translate(dbError.I18nKey, dbError.I18nParams)
	fmt.Printf("  Database error: %s\n", translatedMsg)
}

func multiLocaleExample() {
	config := pkg.I18nConfig{
		DefaultLocale:     "en",
		SupportedLocales:  []string{"en", "de", "fr"},
		FallbackToDefault: true,
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		log.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translations for multiple locales
	manager.LoadLocale("en", map[string]interface{}{
		"welcome": "Welcome to Rockstar Web Framework",
	})
	manager.LoadLocale("de", map[string]interface{}{
		"welcome": "Willkommen beim Rockstar Web Framework",
	})
	manager.LoadLocale("fr", map[string]interface{}{
		"welcome": "Bienvenue dans le Rockstar Web Framework",
	})

	// Demonstrate translation in different locales
	locales := []string{"en", "de", "fr"}
	for _, locale := range locales {
		manager.SetLanguage(locale)
		fmt.Printf("  [%s] %s\n", locale, manager.Translate("welcome"))
	}

	// Demonstrate fallback behavior
	fmt.Println("\n  Fallback behavior:")
	manager.SetLanguage("en")
	manager.LoadLocale("en", map[string]interface{}{
		"only": map[string]interface{}{
			"english": "This is only in English",
		},
	})

	// Try to get translation in German (should fallback to English)
	manager.SetLanguage("de")
	fmt.Printf("  [de] %s (fallback to en)\n", manager.Translate("only.english"))
}

// Example of integrating i18n with a web handler
func exampleHandler() {
	// This would typically be called within a request handler
	config := pkg.I18nConfig{
		DefaultLocale: "en",
		LocalesDir:    "./locales",
	}

	manager, _ := pkg.NewI18nManager(config)

	// Detect user's preferred language from request headers
	// (In real code, you'd get this from Accept-Language header)
	userLocale := "de"
	manager.SetLanguage(userLocale)

	// Use i18n for response messages
	successMsg := manager.Translate("message.success.created")
	fmt.Printf("\n  API Response: %s\n", successMsg)

	// Use i18n for error messages
	errorMsg := manager.Translate("error.validation.failed")
	fmt.Printf("  API Error: %s\n", errorMsg)
}

// Example of using i18n with the context
func exampleWithContext() {
	fmt.Println("\n7. I18n with Context")

	// In a real application, the i18n manager would be accessible through the context
	// ctx.I18n().Translate("key")
	// ctx.Logger().Info("log.key", params)

	fmt.Println("  In a real handler, you would access i18n through:")
	fmt.Println("    - ctx.I18n().Translate(\"error.auth.failed\")")
	fmt.Println("    - ctx.Logger().Info(\"log.request.completed\", params)")
}

func init() {
	// Configure slog for better output
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}
