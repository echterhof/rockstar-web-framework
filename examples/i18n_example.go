//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Configuration with locale file loading
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        "./examples/locales", // Load locale files from this directory
			SupportedLocales:  []string{"en", "de"},
			FallbackToDefault: true,
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get router
	router := app.Router()

	// Route 1: Get welcome message in current locale
	router.GET("/welcome", func(ctx pkg.Context) error {
		// Get i18n manager from context
		i18n := ctx.I18n()

		// Translate welcome message
		message := i18n.Translate("message.welcome")

		return ctx.JSON(200, map[string]interface{}{
			"message": message,
			"locale":  i18n.GetLanguage(),
		})
	})

	// Route 2: Get welcome message in specific locale
	router.GET("/welcome/:locale", func(ctx pkg.Context) error {
		locale := ctx.Params()["locale"]
		i18n := ctx.I18n()

		// Translate with specific locale
		message := i18n.Translate("message.welcome")

		return ctx.JSON(200, map[string]interface{}{
			"message": message,
			"locale":  locale,
		})
	})

	// Route 3: Demonstrate translation with parameters
	router.GET("/error/missing-field/:field", func(ctx pkg.Context) error {
		field := ctx.Params()["field"]
		i18n := ctx.I18n()

		// Translate error message with parameter
		message := i18n.Translate("error.validation.missing_field", map[string]interface{}{
			"field": field,
		})

		return ctx.JSON(400, map[string]interface{}{
			"error":  message,
			"locale": i18n.GetLanguage(),
		})
	})

	// Route 4: Demonstrate rate limit message with multiple parameters
	router.GET("/error/rate-limit", func(ctx pkg.Context) error {
		i18n := ctx.I18n()

		// Translate with multiple parameters
		message := i18n.Translate("error.request.rate_limit_exceeded", map[string]interface{}{
			"limit":  100,
			"window": "minute",
		})

		return ctx.JSON(429, map[string]interface{}{
			"error":  message,
			"locale": i18n.GetLanguage(),
		})
	})

	// Route 5: Switch locale dynamically
	router.POST("/locale/:locale", func(ctx pkg.Context) error {
		locale := ctx.Params()["locale"]
		i18n := ctx.I18n()

		// Set language
		if err := i18n.SetLanguage(locale); err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": fmt.Sprintf("Unsupported locale: %s", locale),
			})
		}

		// Return confirmation in new locale
		message := i18n.Translate("message.success.updated")

		return ctx.JSON(200, map[string]interface{}{
			"message":    message,
			"locale":     i18n.GetLanguage(),
			"locale_set": locale,
		})
	})

	// Route 6: Get all supported locales
	router.GET("/locales", func(ctx pkg.Context) error {
		i18n := ctx.I18n()

		return ctx.JSON(200, map[string]interface{}{
			"supported_locales": i18n.GetSupportedLanguages(),
			"current_locale":    i18n.GetLanguage(),
		})
	})

	// Route 7: Demonstrate fallback behavior
	router.GET("/fallback/:locale", func(ctx pkg.Context) error {
		locale := ctx.Params()["locale"]
		i18n := ctx.I18n()

		// Try to translate a key that might not exist in the requested locale
		// This will fallback to the default locale (English)
		// First set the language, then translate
		originalLang := i18n.GetLanguage()
		i18n.SetLanguage(locale)
		message := i18n.Translate("message.welcome")
		i18n.SetLanguage(originalLang)

		// Check if the locale is actually supported
		supportedLangs := i18n.GetSupportedLanguages()
		isSupported := false
		for _, lang := range supportedLangs {
			if lang == locale {
				isSupported = true
				break
			}
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":          message,
			"requested_locale": locale,
			"locale_supported": isSupported,
			"fallback_used":    !isSupported,
		})
	})

	// Route 8: Demonstrate all error categories
	router.GET("/errors/:category", func(ctx pkg.Context) error {
		category := ctx.Params()["category"]
		i18n := ctx.I18n()

		var errors []map[string]string

		switch category {
		case "authentication":
			errors = []map[string]string{
				{"key": "error.authentication.failed", "message": i18n.Translate("error.authentication.failed")},
				{"key": "error.authentication.invalid_token", "message": i18n.Translate("error.authentication.invalid_token")},
				{"key": "error.authentication.token_expired", "message": i18n.Translate("error.authentication.token_expired")},
				{"key": "error.authentication.unauthorized", "message": i18n.Translate("error.authentication.unauthorized")},
			}
		case "validation":
			errors = []map[string]string{
				{"key": "error.validation.failed", "message": i18n.Translate("error.validation.failed")},
				{"key": "error.validation.invalid_input", "message": i18n.Translate("error.validation.invalid_input")},
			}
		case "database":
			errors = []map[string]string{
				{"key": "error.database.connection", "message": i18n.Translate("error.database.connection")},
				{"key": "error.database.record_not_found", "message": i18n.Translate("error.database.record_not_found")},
			}
		default:
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid category. Use: authentication, validation, or database",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"category": category,
			"locale":   i18n.GetLanguage(),
			"errors":   errors,
		})
	})

	// Route 9: Demonstrate log messages
	router.GET("/logs", func(ctx pkg.Context) error {
		i18n := ctx.I18n()

		logs := []map[string]string{
			{
				"key":     "log.server.starting",
				"message": i18n.Translate("log.server.starting", map[string]interface{}{"address": "localhost:8080"}),
			},
			{
				"key": "log.request.completed",
				"message": i18n.Translate("log.request.completed", map[string]interface{}{
					"method":   "GET",
					"path":     "/api/users",
					"status":   200,
					"duration": 45,
				}),
			},
			{
				"key":     "log.cache.hit",
				"message": i18n.Translate("log.cache.hit", map[string]interface{}{"key": "user:123"}),
			},
		}

		return ctx.JSON(200, map[string]interface{}{
			"locale": i18n.GetLanguage(),
			"logs":   logs,
		})
	})

	// Route 10: Demonstrate locale detection from Accept-Language header
	router.GET("/auto-locale", func(ctx pkg.Context) error {
		// Get Accept-Language header
		acceptLang := ctx.GetHeader("Accept-Language")
		i18n := ctx.I18n()

		// Simple locale detection (in production, use a proper parser)
		detectedLocale := "en" // default
		if len(acceptLang) >= 2 {
			lang := acceptLang[:2]
			// Check if language is supported
			supportedLangs := i18n.GetSupportedLanguages()
			for _, supported := range supportedLangs {
				if supported == lang {
					detectedLocale = lang
					break
				}
			}
		}

		// Get message in detected locale
		originalLang := i18n.GetLanguage()
		i18n.SetLanguage(detectedLocale)
		message := i18n.Translate("message.welcome")
		i18n.SetLanguage(originalLang)

		return ctx.JSON(200, map[string]interface{}{
			"accept_language": acceptLang,
			"detected_locale": detectedLocale,
			"message":         message,
		})
	})

	// Startup message
	fmt.Printf("🎸 Rockstar Web Framework - Internationalization Example\n")
	fmt.Printf("=========================================================\n\n")
	fmt.Printf("Listening on :8080\n\n")
	fmt.Printf("Available endpoints:\n")
	fmt.Printf("  GET  /welcome                    - Get welcome message in current locale\n")
	fmt.Printf("  GET  /welcome/:locale            - Get welcome message in specific locale (en, de)\n")
	fmt.Printf("  GET  /error/missing-field/:field - Demonstrate parameterized error message\n")
	fmt.Printf("  GET  /error/rate-limit           - Demonstrate multi-parameter error message\n")
	fmt.Printf("  POST /locale/:locale             - Switch current locale\n")
	fmt.Printf("  GET  /locales                    - Get all supported locales\n")
	fmt.Printf("  GET  /fallback/:locale           - Demonstrate fallback behavior\n")
	fmt.Printf("  GET  /errors/:category           - Get error messages by category\n")
	fmt.Printf("  GET  /logs                       - Demonstrate log message translation\n")
	fmt.Printf("  GET  /auto-locale                - Demonstrate Accept-Language header detection\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  curl http://localhost:8080/welcome\n")
	fmt.Printf("  curl http://localhost:8080/welcome/de\n")
	fmt.Printf("  curl http://localhost:8080/error/missing-field/email\n")
	fmt.Printf("  curl -X POST http://localhost:8080/locale/de\n")
	fmt.Printf("  curl http://localhost:8080/locales\n")
	fmt.Printf("  curl http://localhost:8080/fallback/fr\n")
	fmt.Printf("  curl http://localhost:8080/errors/authentication\n")
	fmt.Printf("  curl -H \"Accept-Language: de\" http://localhost:8080/auto-locale\n\n")

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
