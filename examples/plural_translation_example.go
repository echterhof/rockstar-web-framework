package main

import (
	"fmt"
	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create i18n manager
	config := pkg.I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := pkg.NewI18nManager(config)
	if err != nil {
		panic(err)
	}

	// Load English plural translations
	manager.LoadLocale("en", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "You have {{count}} item",
			"other": "You have {{count}} items",
		},
		"files": map[string]interface{}{
			"one":   "{{count}} file selected",
			"other": "{{count}} files selected",
		},
	})

	// Load German plural translations
	manager.LoadLocale("de", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "Sie haben {{count}} Element",
			"other": "Sie haben {{count}} Elemente",
		},
		"files": map[string]interface{}{
			"one":   "{{count}} Datei ausgewählt",
			"other": "{{count}} Dateien ausgewählt",
		},
	})

	// Demonstrate English plural translations
	fmt.Println("=== English Translations ===")
	manager.SetLanguage("en")
	fmt.Println(manager.TranslatePlural("items", 0))
	fmt.Println(manager.TranslatePlural("items", 1))
	fmt.Println(manager.TranslatePlural("items", 5))
	fmt.Println(manager.TranslatePlural("files", 1))
	fmt.Println(manager.TranslatePlural("files", 10))

	// Demonstrate German plural translations
	fmt.Println("\n=== German Translations ===")
	manager.SetLanguage("de")
	fmt.Println(manager.TranslatePlural("items", 0))
	fmt.Println(manager.TranslatePlural("items", 1))
	fmt.Println(manager.TranslatePlural("items", 5))
	fmt.Println(manager.TranslatePlural("files", 1))
	fmt.Println(manager.TranslatePlural("files", 10))

	// Demonstrate fallback behavior
	fmt.Println("\n=== Fallback Behavior ===")
	manager.SetLanguage("fr") // French not loaded, should fallback to English
	fmt.Println(manager.TranslatePlural("items", 3))
}
