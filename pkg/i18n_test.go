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
