package pkg

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestNewI18nManager(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.GetLanguage() != "en" {
		t.Errorf("Expected default locale 'en', got '%s'", manager.GetLanguage())
	}
}

func TestI18nManager_LoadLocale(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translation
	data := map[string]interface{}{
		"test": map[string]interface{}{
			"key": "Test Value",
		},
	}

	err = manager.LoadLocale("en", data)
	if err != nil {
		t.Fatalf("Failed to load locale: %v", err)
	}

	// Verify translation
	result := manager.Translate("test.key")
	if result != "Test Value" {
		t.Errorf("Expected 'Test Value', got '%s'", result)
	}
}

func TestI18nManager_Translate(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translations
	manager.LoadLocale("en", map[string]interface{}{"greeting": "Hello"})
	manager.LoadLocale("de", map[string]interface{}{"greeting": "Hallo"})

	// Test English translation
	result := manager.Translate("greeting")
	if result != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", result)
	}

	// Change locale to German
	manager.SetLanguage("de")

	// Test German translation
	result = manager.Translate("greeting")
	if result != "Hallo" {
		t.Errorf("Expected 'Hallo', got '%s'", result)
	}
}

func TestI18nManager_TranslateWithParams(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translation with placeholders
	manager.LoadLocale("en", map[string]interface{}{
		"greeting": map[string]interface{}{
			"user": "Hello, {{name}}!",
		},
	})

	// Translate with parameters
	result := manager.Translate("greeting.user", map[string]interface{}{
		"name": "Alice",
	})

	if result != "Hello, Alice!" {
		t.Errorf("Expected 'Hello, Alice!', got '%s'", result)
	}
}

func TestI18nManager_TranslateWithMultipleParams(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translation with multiple placeholders
	manager.LoadLocale("en", map[string]interface{}{
		"error": map[string]interface{}{
			"rate_limit": "Rate limit exceeded: {{limit}} requests per {{window}}",
		},
	})

	// Translate with parameters
	result := manager.Translate("error.rate_limit", map[string]interface{}{
		"limit":  100,
		"window": "minute",
	})

	expected := "Rate limit exceeded: 100 requests per minute"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestI18nManager_FallbackToDefault(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translation only in English
	manager.LoadLocale("en", map[string]interface{}{
		"test": map[string]interface{}{
			"key": "English Value",
		},
	})

	// Set locale to German (which doesn't have this translation)
	manager.SetLanguage("de")

	// Should fallback to English
	result := manager.Translate("test.key")
	if result != "English Value" {
		t.Errorf("Expected fallback to 'English Value', got '%s'", result)
	}
}

func TestI18nManager_MissingKey(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Try to translate a key that doesn't exist
	result := manager.Translate("nonexistent.key")

	// Should return the key itself
	if result != "nonexistent.key" {
		t.Errorf("Expected key 'nonexistent.key', got '%s'", result)
	}
}

func TestI18nManager_GetSupportedLanguages(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Add translations for multiple locales
	manager.LoadLocale("en", map[string]interface{}{"test": map[string]interface{}{"key": "English"}})
	manager.LoadLocale("de", map[string]interface{}{"test": map[string]interface{}{"key": "German"}})
	manager.LoadLocale("fr", map[string]interface{}{"test": map[string]interface{}{"key": "French"}})

	// Get available locales
	locales := manager.GetSupportedLanguages()

	if len(locales) != 3 {
		t.Errorf("Expected 3 locales, got %d", len(locales))
	}

	// Check if all locales are present
	localeMap := make(map[string]bool)
	for _, locale := range locales {
		localeMap[locale] = true
	}

	if !localeMap["en"] || !localeMap["de"] || !localeMap["fr"] {
		t.Error("Expected all locales to be present")
	}
}

func TestI18nManager_SetLanguage(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:    "en",
		SupportedLocales: []string{"en", "de", "fr"},
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Set valid locale
	err = manager.SetLanguage("de")
	if err != nil {
		t.Errorf("Failed to set valid locale: %v", err)
	}

	if manager.GetLanguage() != "de" {
		t.Errorf("Expected locale 'de', got '%s'", manager.GetLanguage())
	}

	// Try to set invalid locale
	err = manager.SetLanguage("es")
	if err == nil {
		t.Error("Expected error when setting unsupported locale")
	}
}

func TestI18nManager_LoadLocaleFromFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test locale file
	localeFile := filepath.Join(tmpDir, "locales.en.yaml")
	content := `
error:
  authentication:
    failed: "Authentication failed"
  validation:
    missing_field: "Required field '{{field}}' is missing"

log:
  server:
    starting: "Server starting on {{address}}"
`

	err := os.WriteFile(localeFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test locale file: %v", err)
	}

	// Create i18n manager
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load locale file
	err = manager.LoadLocaleFromFile("en", localeFile)
	if err != nil {
		t.Fatalf("Failed to load locale file: %v", err)
	}

	// Test translations
	result := manager.Translate("error.authentication.failed")
	if result != "Authentication failed" {
		t.Errorf("Expected 'Authentication failed', got '%s'", result)
	}

	// Test translation with parameters
	result = manager.Translate("error.validation.missing_field", map[string]interface{}{
		"field": "email",
	})
	if result != "Required field 'email' is missing" {
		t.Errorf("Expected 'Required field 'email' is missing', got '%s'", result)
	}
}

func TestI18nManager_LoadLocalesFromDir(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create English locale file
	enFile := filepath.Join(tmpDir, "locales.en.yaml")
	enContent := `
greeting: "Hello"
farewell: "Goodbye"
`
	err := os.WriteFile(enFile, []byte(enContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create English locale file: %v", err)
	}

	// Create German locale file
	deFile := filepath.Join(tmpDir, "locales.de.yaml")
	deContent := `
greeting: "Hallo"
farewell: "Auf Wiedersehen"
`
	err = os.WriteFile(deFile, []byte(deContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create German locale file: %v", err)
	}

	// Create i18n manager with LocalesDir
	config := I18nConfig{
		DefaultLocale: "en",
		LocalesDir:    tmpDir,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Test English translations
	result := manager.Translate("greeting")
	if result != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", result)
	}

	// Switch to German
	manager.SetLanguage("de")

	// Test German translations
	result = manager.Translate("greeting")
	if result != "Hallo" {
		t.Errorf("Expected 'Hallo', got '%s'", result)
	}

	result = manager.Translate("farewell")
	if result != "Auf Wiedersehen" {
		t.Errorf("Expected 'Auf Wiedersehen', got '%s'", result)
	}
}

func TestNewI18nLogger(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Create logger with custom slog.Logger
	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := NewI18nLogger(manager, customLogger)

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	if logger.GetUnderlyingLogger() != customLogger {
		t.Error("Expected custom logger to be used")
	}
}

func TestNewI18nLogger_NilLogger(t *testing.T) {
	config := I18nConfig{
		DefaultLocale: "en",
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Create logger with nil slog.Logger (should use default)
	logger := NewI18nLogger(manager, nil)

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	if logger.GetUnderlyingLogger() == nil {
		t.Error("Expected default logger to be used")
	}
}

func TestStandardLogger(t *testing.T) {
	logger := NewLogger(nil)

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Test all methods (should not panic)
	logger.Info("Test info")
	logger.Warn("Test warn")
	logger.Error("Test error")
	logger.Debug("Test debug")
}

func TestStandardLogger_WithRequestID(t *testing.T) {
	logger := NewLogger(nil)

	contextLogger := logger.WithRequestID("req-123")

	if contextLogger == nil {
		t.Fatal("Expected non-nil context logger")
	}

	// Test logging with context (should not panic)
	contextLogger.Info("Test with context")
}

// Property-Based Tests for Plural Translation Support

// **Feature: todo-implementations, Property 9: Plural form selection**
// **Validates: Requirements 3.1**
func TestProperty_PluralFormSelection(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations for English
	manager.LoadLocale("en", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "{{count}} item",
			"other": "{{count}} items",
		},
	})

	// Load plural translations for German
	manager.LoadLocale("de", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "{{count}} Element",
			"other": "{{count}} Elemente",
		},
	})

	// Test English plural form selection
	testCases := []struct {
		locale   string
		count    int
		expected string
	}{
		// English tests
		{"en", 0, "0 items"},
		{"en", 1, "1 item"},
		{"en", 2, "2 items"},
		{"en", 5, "5 items"},
		{"en", 100, "100 items"},
		{"en", -1, "-1 items"},
		// German tests
		{"de", 0, "0 Elemente"},
		{"de", 1, "1 Element"},
		{"de", 2, "2 Elemente"},
		{"de", 5, "5 Elemente"},
		{"de", 100, "100 Elemente"},
	}

	for _, tc := range testCases {
		manager.SetLanguage(tc.locale)
		result := manager.TranslatePlural("items", tc.count)
		if result != tc.expected {
			t.Errorf("Locale %s, count %d: expected '%s', got '%s'", tc.locale, tc.count, tc.expected, result)
		}
	}
}

// **Feature: todo-implementations, Property 10: Language-specific plural rules**
// **Validates: Requirements 3.2**
func TestProperty_LanguageSpecificPluralRules(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations
	manager.LoadLocale("en", map[string]interface{}{
		"books": map[string]interface{}{
			"one":   "{{count}} book",
			"other": "{{count}} books",
		},
	})

	manager.LoadLocale("de", map[string]interface{}{
		"books": map[string]interface{}{
			"one":   "{{count}} Buch",
			"other": "{{count}} Bücher",
		},
	})

	// Test that English and German both use the same rule (count == 1 for singular)
	// but apply it correctly to their respective translations

	// English: 1 should use "one" form
	manager.SetLanguage("en")
	result := manager.TranslatePlural("books", 1)
	if result != "1 book" {
		t.Errorf("English: expected '1 book', got '%s'", result)
	}

	// German: 1 should use "one" form
	manager.SetLanguage("de")
	result = manager.TranslatePlural("books", 1)
	if result != "1 Buch" {
		t.Errorf("German: expected '1 Buch', got '%s'", result)
	}

	// English: 2 should use "other" form
	manager.SetLanguage("en")
	result = manager.TranslatePlural("books", 2)
	if result != "2 books" {
		t.Errorf("English: expected '2 books', got '%s'", result)
	}

	// German: 2 should use "other" form
	manager.SetLanguage("de")
	result = manager.TranslatePlural("books", 2)
	if result != "2 Bücher" {
		t.Errorf("German: expected '2 Bücher', got '%s'", result)
	}

	// Test with 0 - both languages should use "other" form
	manager.SetLanguage("en")
	result = manager.TranslatePlural("books", 0)
	if result != "0 books" {
		t.Errorf("English: expected '0 books', got '%s'", result)
	}

	manager.SetLanguage("de")
	result = manager.TranslatePlural("books", 0)
	if result != "0 Bücher" {
		t.Errorf("German: expected '0 Bücher', got '%s'", result)
	}
}

// **Feature: todo-implementations, Property 11: Plural fallback behavior**
// **Validates: Requirements 3.3**
func TestProperty_PluralFallback(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations only in English (default locale)
	manager.LoadLocale("en", map[string]interface{}{
		"messages": map[string]interface{}{
			"one":   "{{count}} message",
			"other": "{{count}} messages",
		},
	})

	// Set locale to German (which doesn't have this translation)
	manager.SetLanguage("de")

	// Should fallback to English default locale
	result := manager.TranslatePlural("messages", 1)
	if result != "1 message" {
		t.Errorf("Expected fallback to '1 message', got '%s'", result)
	}

	result = manager.TranslatePlural("messages", 5)
	if result != "5 messages" {
		t.Errorf("Expected fallback to '5 messages', got '%s'", result)
	}

	// Test with missing key entirely - should return the key
	result = manager.TranslatePlural("nonexistent", 1)
	if result != "nonexistent" {
		t.Errorf("Expected key 'nonexistent', got '%s'", result)
	}
}

// **Feature: todo-implementations, Property 12: Count interpolation**
// **Validates: Requirements 3.5**
func TestProperty_CountInterpolation(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations with {{count}} placeholder
	manager.LoadLocale("en", map[string]interface{}{
		"files": map[string]interface{}{
			"one":   "{{count}} file selected",
			"other": "{{count}} files selected",
		},
	})

	// Test various counts to ensure {{count}} is properly interpolated
	testCases := []struct {
		count    int
		expected string
	}{
		{0, "0 files selected"},
		{1, "1 file selected"},
		{2, "2 files selected"},
		{10, "10 files selected"},
		{100, "100 files selected"},
		{1000, "1000 files selected"},
		{-5, "-5 files selected"},
	}

	for _, tc := range testCases {
		result := manager.TranslatePlural("files", tc.count)
		if result != tc.expected {
			t.Errorf("Count %d: expected '%s', got '%s'", tc.count, tc.expected, result)
		}
	}

	// Test with additional parameters alongside count
	manager.LoadLocale("en", map[string]interface{}{
		"downloads": map[string]interface{}{
			"one":   "{{count}} file ({{size}} MB)",
			"other": "{{count}} files ({{size}} MB)",
		},
	})

	result := manager.TranslatePlural("downloads", 1, map[string]interface{}{
		"size": 5,
	})
	if result != "1 file (5 MB)" {
		t.Errorf("Expected '1 file (5 MB)', got '%s'", result)
	}

	result = manager.TranslatePlural("downloads", 3, map[string]interface{}{
		"size": 15,
	})
	if result != "3 files (15 MB)" {
		t.Errorf("Expected '3 files (15 MB)', got '%s'", result)
	}
}

// Unit Tests for Plural Translation Support

func TestI18nManager_TranslatePlural_English(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load English plural translations
	manager.LoadLocale("en", map[string]interface{}{
		"apples": map[string]interface{}{
			"one":   "{{count}} apple",
			"other": "{{count}} apples",
		},
	})

	// Test singular form (count = 1)
	result := manager.TranslatePlural("apples", 1)
	if result != "1 apple" {
		t.Errorf("Expected '1 apple', got '%s'", result)
	}

	// Test plural form (count = 0)
	result = manager.TranslatePlural("apples", 0)
	if result != "0 apples" {
		t.Errorf("Expected '0 apples', got '%s'", result)
	}

	// Test plural form (count = 2)
	result = manager.TranslatePlural("apples", 2)
	if result != "2 apples" {
		t.Errorf("Expected '2 apples', got '%s'", result)
	}

	// Test plural form (count = 100)
	result = manager.TranslatePlural("apples", 100)
	if result != "100 apples" {
		t.Errorf("Expected '100 apples', got '%s'", result)
	}
}

func TestI18nManager_TranslatePlural_German(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load German plural translations
	manager.LoadLocale("de", map[string]interface{}{
		"autos": map[string]interface{}{
			"one":   "{{count}} Auto",
			"other": "{{count}} Autos",
		},
	})

	manager.SetLanguage("de")

	// Test singular form (count = 1)
	result := manager.TranslatePlural("autos", 1)
	if result != "1 Auto" {
		t.Errorf("Expected '1 Auto', got '%s'", result)
	}

	// Test plural form (count = 0)
	result = manager.TranslatePlural("autos", 0)
	if result != "0 Autos" {
		t.Errorf("Expected '0 Autos', got '%s'", result)
	}

	// Test plural form (count = 5)
	result = manager.TranslatePlural("autos", 5)
	if result != "5 Autos" {
		t.Errorf("Expected '5 Autos', got '%s'", result)
	}
}

func TestI18nManager_TranslatePlural_EdgeCases(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations
	manager.LoadLocale("en", map[string]interface{}{
		"items": map[string]interface{}{
			"one":   "{{count}} item",
			"other": "{{count}} items",
		},
	})

	// Test with 0 (should use "other" form)
	result := manager.TranslatePlural("items", 0)
	if result != "0 items" {
		t.Errorf("Expected '0 items', got '%s'", result)
	}

	// Test with 1 (should use "one" form)
	result = manager.TranslatePlural("items", 1)
	if result != "1 item" {
		t.Errorf("Expected '1 item', got '%s'", result)
	}

	// Test with negative number (should use "other" form)
	result = manager.TranslatePlural("items", -5)
	if result != "-5 items" {
		t.Errorf("Expected '-5 items', got '%s'", result)
	}

	// Test with large number
	result = manager.TranslatePlural("items", 1000000)
	if result != "1000000 items" {
		t.Errorf("Expected '1000000 items', got '%s'", result)
	}
}

func TestI18nManager_TranslatePlural_MissingTranslation(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Don't load any translations

	// Test with missing key - should return the key itself
	result := manager.TranslatePlural("missing.key", 5)
	if result != "missing.key" {
		t.Errorf("Expected 'missing.key', got '%s'", result)
	}
}

func TestI18nManager_TranslatePlural_WithAdditionalParams(t *testing.T) {
	config := I18nConfig{
		DefaultLocale:     "en",
		FallbackToDefault: true,
	}

	manager, err := NewI18nManager(config)
	if err != nil {
		t.Fatalf("Failed to create i18n manager: %v", err)
	}

	// Load plural translations with additional parameters
	manager.LoadLocale("en", map[string]interface{}{
		"users": map[string]interface{}{
			"one":   "{{count}} user named {{name}}",
			"other": "{{count}} users including {{name}}",
		},
	})

	// Test with additional parameters
	result := manager.TranslatePlural("users", 1, map[string]interface{}{
		"name": "Alice",
	})
	if result != "1 user named Alice" {
		t.Errorf("Expected '1 user named Alice', got '%s'", result)
	}

	result = manager.TranslatePlural("users", 5, map[string]interface{}{
		"name": "Bob",
	})
	if result != "5 users including Bob" {
		t.Errorf("Expected '5 users including Bob', got '%s'", result)
	}
}

func TestPluralRules_EnglishRule(t *testing.T) {
	// Test English plural rule directly
	testCases := []struct {
		count    int
		expected int
	}{
		{0, 1},   // "other" form
		{1, 0},   // "one" form
		{2, 1},   // "other" form
		{5, 1},   // "other" form
		{100, 1}, // "other" form
		{-1, 1},  // "other" form
	}

	for _, tc := range testCases {
		result := EnglishPluralRule(tc.count)
		if result != tc.expected {
			t.Errorf("EnglishPluralRule(%d): expected %d, got %d", tc.count, tc.expected, result)
		}
	}
}

func TestPluralRules_GermanRule(t *testing.T) {
	// Test German plural rule directly
	testCases := []struct {
		count    int
		expected int
	}{
		{0, 1},   // "other" form
		{1, 0},   // "one" form
		{2, 1},   // "other" form
		{5, 1},   // "other" form
		{100, 1}, // "other" form
		{-1, 1},  // "other" form
	}

	for _, tc := range testCases {
		result := GermanPluralRule(tc.count)
		if result != tc.expected {
			t.Errorf("GermanPluralRule(%d): expected %d, got %d", tc.count, tc.expected, result)
		}
	}
}

func TestPluralRules_RegisterCustomRule(t *testing.T) {
	pr := NewPluralRules()

	// Register a custom rule for a fictional language
	customRule := func(count int) int {
		if count == 0 {
			return 0 // zero form
		} else if count == 1 {
			return 1 // one form
		}
		return 2 // other form
	}

	pr.RegisterRule("custom", customRule)

	// Verify the rule was registered
	rule := pr.GetRule("custom")
	if rule == nil {
		t.Fatal("Expected custom rule to be registered")
	}

	// Test the custom rule
	if rule(0) != 0 {
		t.Errorf("Expected 0 for count 0")
	}
	if rule(1) != 1 {
		t.Errorf("Expected 1 for count 1")
	}
	if rule(5) != 2 {
		t.Errorf("Expected 2 for count 5")
	}
}
