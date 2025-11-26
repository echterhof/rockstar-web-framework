package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite3",
			Database: ":memory:",
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			SupportedLocales:  []string{"en", "de"},
			FallbackToDefault: true,
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get i18n manager and load plural translations
	i18n := app.I18n()

	// Load English plural translations
	i18n.LoadLocale("en", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "You have {{count}} item",
			"other": "You have {{count}} items",
		},
		"files": map[string]interface{}{
			"one":   "{{count}} file selected",
			"other": "{{count}} files selected",
		},
		"users": map[string]interface{}{
			"one":   "{{count}} user online",
			"other": "{{count}} users online",
		},
		"messages": map[string]interface{}{
			"one":   "You have {{count}} unread message",
			"other": "You have {{count}} unread messages",
		},
		"downloads": map[string]interface{}{
			"one":   "{{count}} download remaining",
			"other": "{{count}} downloads remaining",
		},
	})

	// Load German plural translations
	i18n.LoadLocale("de", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "Sie haben {{count}} Element",
			"other": "Sie haben {{count}} Elemente",
		},
		"files": map[string]interface{}{
			"one":   "{{count}} Datei ausgewählt",
			"other": "{{count}} Dateien ausgewählt",
		},
		"users": map[string]interface{}{
			"one":   "{{count}} Benutzer online",
			"other": "{{count}} Benutzer online",
		},
		"messages": map[string]interface{}{
			"one":   "Sie haben {{count}} ungelesene Nachricht",
			"other": "Sie haben {{count}} ungelesene Nachrichten",
		},
		"downloads": map[string]interface{}{
			"one":   "{{count}} Download verbleibend",
			"other": "{{count}} Downloads verbleibend",
		},
	})

	// Get router
	router := app.Router()

	// Route 1: Demonstrate plural translation with count parameter
	router.GET("/items/:count", func(ctx pkg.Context) error {
		countStr := ctx.Params()["count"]
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid count parameter",
			})
		}

		i18n := ctx.I18n()

		// Translate with plural support
		message := i18n.TranslatePlural("items", count)

		return ctx.JSON(200, map[string]interface{}{
			"count":   count,
			"message": message,
			"locale":  i18n.GetLanguage(),
		})
	})

	// Route 2: Demonstrate plural translation in specific locale
	router.GET("/items/:count/:locale", func(ctx pkg.Context) error {
		countStr := ctx.Params()["count"]
		locale := ctx.Params()["locale"]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid count parameter",
			})
		}

		i18n := ctx.I18n()

		// Translate with plural support in specific locale
		originalLang := i18n.GetLanguage()
		i18n.SetLanguage(locale)
		message := i18n.TranslatePlural("items", count)
		i18n.SetLanguage(originalLang)

		return ctx.JSON(200, map[string]interface{}{
			"count":   count,
			"message": message,
			"locale":  locale,
		})
	})

	// Route 3: Demonstrate multiple plural categories
	router.GET("/stats/:files/:users/:messages", func(ctx pkg.Context) error {
		filesCount, _ := strconv.Atoi(ctx.Params()["files"])
		usersCount, _ := strconv.Atoi(ctx.Params()["users"])
		messagesCount, _ := strconv.Atoi(ctx.Params()["messages"])

		i18n := ctx.I18n()

		stats := map[string]interface{}{
			"files": map[string]interface{}{
				"count":   filesCount,
				"message": i18n.TranslatePlural("files", filesCount),
			},
			"users": map[string]interface{}{
				"count":   usersCount,
				"message": i18n.TranslatePlural("users", usersCount),
			},
			"messages": map[string]interface{}{
				"count":   messagesCount,
				"message": i18n.TranslatePlural("messages", messagesCount),
			},
		}

		return ctx.JSON(200, map[string]interface{}{
			"locale": i18n.GetLanguage(),
			"stats":  stats,
		})
	})

	// Route 4: Compare plural forms across locales
	router.GET("/compare/:count", func(ctx pkg.Context) error {
		countStr := ctx.Params()["count"]
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid count parameter",
			})
		}

		i18n := ctx.I18n()

		// Get translations in all supported locales
		translations := make(map[string]string)
		originalLang := i18n.GetLanguage()
		for _, locale := range i18n.GetSupportedLanguages() {
			i18n.SetLanguage(locale)
			translations[locale] = i18n.TranslatePlural("items", count)
		}
		i18n.SetLanguage(originalLang)

		return ctx.JSON(200, map[string]interface{}{
			"count":        count,
			"translations": translations,
		})
	})

	// Route 5: Demonstrate edge cases (0, 1, many)
	router.GET("/edge-cases/:category", func(ctx pkg.Context) error {
		category := ctx.Params()["category"]
		i18n := ctx.I18n()

		// Test with 0, 1, and multiple counts
		testCounts := []int{0, 1, 2, 5, 10, 100}
		results := make([]map[string]interface{}, 0, len(testCounts))

		for _, count := range testCounts {
			results = append(results, map[string]interface{}{
				"count":   count,
				"message": i18n.TranslatePlural(category, count),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"category": category,
			"locale":   i18n.GetLanguage(),
			"results":  results,
		})
	})

	// Route 6: Demonstrate plural with additional parameters
	router.GET("/downloads/:count/:username", func(ctx pkg.Context) error {
		countStr := ctx.Params()["count"]
		username := ctx.Params()["username"]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid count parameter",
			})
		}

		i18n := ctx.I18n()

		// Translate with plural and additional parameters
		message := i18n.TranslatePlural("downloads", count, map[string]interface{}{
			"username": username,
		})

		return ctx.JSON(200, map[string]interface{}{
			"count":    count,
			"username": username,
			"message":  message,
			"locale":   i18n.GetLanguage(),
		})
	})

	// Route 7: Demonstrate fallback behavior for unsupported locale
	router.GET("/fallback/:count/:locale", func(ctx pkg.Context) error {
		countStr := ctx.Params()["count"]
		locale := ctx.Params()["locale"]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid count parameter",
			})
		}

		i18n := ctx.I18n()

		// Try to translate in unsupported locale (should fallback to default)
		originalLang := i18n.GetLanguage()
		i18n.SetLanguage(locale)
		message := i18n.TranslatePlural("items", count)
		i18n.SetLanguage(originalLang)

		// Check if locale is supported
		supportedLangs := i18n.GetSupportedLanguages()
		isSupported := false
		for _, lang := range supportedLangs {
			if lang == locale {
				isSupported = true
				break
			}
		}

		return ctx.JSON(200, map[string]interface{}{
			"count":            count,
			"requested_locale": locale,
			"locale_supported": isSupported,
			"message":          message,
			"fallback_used":    !isSupported,
		})
	})

	// Route 8: Interactive plural form tester
	router.GET("/test", func(ctx pkg.Context) error {
		i18n := ctx.I18n()

		// Generate test data for all categories
		categories := []string{"items", "files", "users", "messages", "downloads"}
		testData := make(map[string][]map[string]interface{})

		for _, category := range categories {
			categoryData := make([]map[string]interface{}, 0)
			for count := 0; count <= 5; count++ {
				categoryData = append(categoryData, map[string]interface{}{
					"count":   count,
					"message": i18n.TranslatePlural(category, count),
				})
			}
			testData[category] = categoryData
		}

		return ctx.JSON(200, map[string]interface{}{
			"locale":    i18n.GetLanguage(),
			"test_data": testData,
		})
	})

	// Startup message
	fmt.Printf("🎸 Rockstar Web Framework - Plural Translation Example\n")
	fmt.Printf("========================================================\n\n")
	fmt.Printf("Listening on :8080\n\n")
	fmt.Printf("Available endpoints:\n")
	fmt.Printf("  GET /items/:count                    - Get plural translation for item count\n")
	fmt.Printf("  GET /items/:count/:locale            - Get plural translation in specific locale\n")
	fmt.Printf("  GET /stats/:files/:users/:messages   - Get multiple plural translations\n")
	fmt.Printf("  GET /compare/:count                  - Compare plural forms across locales\n")
	fmt.Printf("  GET /edge-cases/:category            - Test edge cases (0, 1, 2, 5, 10, 100)\n")
	fmt.Printf("  GET /downloads/:count/:username      - Plural with additional parameters\n")
	fmt.Printf("  GET /fallback/:count/:locale         - Test fallback behavior\n")
	fmt.Printf("  GET /test                            - Interactive plural form tester\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  curl http://localhost:8080/items/1\n")
	fmt.Printf("  curl http://localhost:8080/items/5\n")
	fmt.Printf("  curl http://localhost:8080/items/3/de\n")
	fmt.Printf("  curl http://localhost:8080/stats/1/5/10\n")
	fmt.Printf("  curl http://localhost:8080/compare/1\n")
	fmt.Printf("  curl http://localhost:8080/edge-cases/files\n")
	fmt.Printf("  curl http://localhost:8080/downloads/3/alice\n")
	fmt.Printf("  curl http://localhost:8080/fallback/5/fr\n")
	fmt.Printf("  curl http://localhost:8080/test\n\n")
	fmt.Printf("Plural Rules:\n")
	fmt.Printf("  English: one (count == 1), other (count != 1)\n")
	fmt.Printf("  German:  one (count == 1), other (count != 1)\n\n")

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
