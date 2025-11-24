package pkg

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// i18nManagerImpl provides internationalization support for the framework
// This is the internal implementation of the I18nManager interface defined in managers.go

// PluralRule is a function that determines the plural form index based on count
// Returns 0 for singular (one), 1 for plural (other), and potentially more for complex languages
type PluralRule func(count int) int

// PluralRules manages plural rules for different locales
type PluralRules struct {
	rules map[string]PluralRule
}

// NewPluralRules creates a new plural rules engine with default rules
func NewPluralRules() *PluralRules {
	pr := &PluralRules{
		rules: make(map[string]PluralRule),
	}

	// Register default rules for English and German
	pr.RegisterRule("en", EnglishPluralRule)
	pr.RegisterRule("de", GermanPluralRule)

	return pr
}

// RegisterRule registers a plural rule for a specific locale
func (pr *PluralRules) RegisterRule(locale string, rule PluralRule) {
	pr.rules[locale] = rule
}

// GetRule returns the plural rule for a locale, or nil if not found
func (pr *PluralRules) GetRule(locale string) PluralRule {
	return pr.rules[locale]
}

// EnglishPluralRule implements English plural rules (2 forms: one, other)
// Form 0 (one): count == 1
// Form 1 (other): everything else
func EnglishPluralRule(count int) int {
	if count == 1 {
		return 0 // "one" form
	}
	return 1 // "other" form
}

// GermanPluralRule implements German plural rules (2 forms: one, other)
// Form 0 (one): count == 1
// Form 1 (other): everything else
func GermanPluralRule(count int) int {
	if count == 1 {
		return 0 // "one" form
	}
	return 1 // "other" form
}

// I18nConfig holds configuration for the i18n manager
type I18nConfig struct {
	// DefaultLocale is the fallback locale when translation is not found
	DefaultLocale string

	// LocalesDir is the directory containing locale files
	LocalesDir string

	// SupportedLocales is the list of supported locales
	SupportedLocales []string

	// FallbackToDefault determines if missing translations should fall back to default locale
	FallbackToDefault bool
}

// i18nManagerImpl implements the I18nManager interface
type i18nManagerImpl struct {
	config        I18nConfig
	currentLocale string
	translations  map[string]map[string]string // locale -> key -> translation
	pluralRules   *PluralRules                 // plural rules engine
	mu            sync.RWMutex
	logger        *I18nLogger
}

// NewI18nManager creates a new internationalization manager
func NewI18nManager(config I18nConfig) (I18nManager, error) {
	if config.DefaultLocale == "" {
		config.DefaultLocale = "en"
	}

	if config.FallbackToDefault == false {
		config.FallbackToDefault = true
	}

	manager := &i18nManagerImpl{
		config:        config,
		currentLocale: config.DefaultLocale,
		translations:  make(map[string]map[string]string),
		pluralRules:   NewPluralRules(),
	}

	// Initialize logger
	manager.logger = NewI18nLogger(manager, slog.Default())

	// Load locale files if LocalesDir is specified
	if config.LocalesDir != "" {
		if err := manager.loadLocalesFromDir(config.LocalesDir); err != nil {
			return nil, fmt.Errorf("failed to load locales: %w", err)
		}
	}

	return manager, nil
}

// loadLocalesFromDir loads all locale files from a directory
func (m *i18nManagerImpl) loadLocalesFromDir(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("locales directory does not exist: %s", dir)
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	// Load each locale file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Check if file matches pattern: locales.{locale}.yaml or locales.{locale}.yml
		if !strings.HasPrefix(filename, "locales.") {
			continue
		}

		ext := filepath.Ext(filename)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		// Extract locale from filename
		// locales.en.yaml -> en
		// locales.de.yaml -> de
		parts := strings.Split(filename, ".")
		if len(parts) < 3 {
			continue
		}

		locale := parts[1]

		// Load the file
		filePath := filepath.Join(dir, filename)
		if err := m.LoadLocaleFromFile(locale, filePath); err != nil {
			return fmt.Errorf("failed to load locale file %s: %w", filename, err)
		}
	}

	return nil
}

// LoadLocaleFromFile loads a locale file from disk
func (m *i18nManagerImpl) LoadLocaleFromFile(locale, filepath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read locale file: %w", err)
	}

	// Parse YAML
	var translations map[string]interface{}
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Flatten nested structure
	flattened := make(map[string]string)
	flattenMap("", translations, flattened)

	// Store translations
	if m.translations[locale] == nil {
		m.translations[locale] = make(map[string]string)
	}

	for key, value := range flattened {
		m.translations[locale][key] = value
	}

	return nil
}

// flattenMap flattens a nested map structure into dot-notation keys
// e.g., {"error": {"auth": {"failed": "Authentication failed"}}} -> "error.auth.failed": "Authentication failed"
func flattenMap(prefix string, data map[string]interface{}, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]interface{}:
			flattenMap(fullKey, v, result)
		default:
			// Convert to string for other types
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

// Translate translates a key to the current locale
func (m *i18nManagerImpl) Translate(key string, params ...interface{}) string {
	return m.TranslateWithLang(m.currentLocale, key, params...)
}

// TranslateWithLang translates a key to a specific locale
func (m *i18nManagerImpl) TranslateWithLang(locale, key string, params ...interface{}) string {
	// Convert params to map format
	paramMap := make(map[string]interface{})
	if len(params) > 0 {
		if pm, ok := params[0].(map[string]interface{}); ok {
			paramMap = pm
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try to get translation for requested locale
	if translations, ok := m.translations[locale]; ok {
		if translation, ok := translations[key]; ok {
			return m.interpolate(translation, paramMap)
		}
	}

	// Fallback to default locale if enabled
	if m.config.FallbackToDefault && locale != m.config.DefaultLocale {
		if translations, ok := m.translations[m.config.DefaultLocale]; ok {
			if translation, ok := translations[key]; ok {
				return m.interpolate(translation, paramMap)
			}
		}
	}

	// Return key if no translation found
	return key
}

// interpolate replaces placeholders in a translation with parameter values
// Supports {{param}} syntax
func (m *i18nManagerImpl) interpolate(translation string, paramMap map[string]interface{}) string {
	result := translation
	for key, value := range paramMap {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// SetLanguage sets the current locale for this manager
func (m *i18nManagerImpl) SetLanguage(locale string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if locale is supported
	if len(m.config.SupportedLocales) > 0 {
		supported := false
		for _, l := range m.config.SupportedLocales {
			if l == locale {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("unsupported locale: %s", locale)
		}
	}

	m.currentLocale = locale
	return nil
}

// GetLanguage returns the current locale
func (m *i18nManagerImpl) GetLanguage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentLocale
}

// GetSupportedLanguages returns all available locales
func (m *i18nManagerImpl) GetSupportedLanguages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locales := make([]string, 0, len(m.translations))
	for locale := range m.translations {
		locales = append(locales, locale)
	}
	return locales
}

// IsLanguageSupported checks if a language is supported
func (m *i18nManagerImpl) IsLanguageSupported(lang string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.translations[lang]
	return exists
}

// GetDefaultLanguage returns the default language
func (m *i18nManagerImpl) GetDefaultLanguage() string {
	return m.config.DefaultLocale
}

// hasKey checks if a translation key exists in the current locale
func (m *i18nManagerImpl) hasKey(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if translations, ok := m.translations[m.currentLocale]; ok {
		_, exists := translations[key]
		return exists
	}

	return false
}

// LoadLocale adds translations for a specific locale
func (m *i18nManagerImpl) LoadLocale(locale string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.translations[locale] == nil {
		m.translations[locale] = make(map[string]string)
	}

	// Flatten nested structure
	flattened := make(map[string]string)
	flattenMap("", data, flattened)

	for key, value := range flattened {
		m.translations[locale][key] = value
	}

	return nil
}

// addTranslation adds a single translation for a specific locale
func (m *i18nManagerImpl) addTranslation(locale, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.translations[locale] == nil {
		m.translations[locale] = make(map[string]string)
	}

	m.translations[locale][key] = value
	return nil
}

// TranslatePlural translates a key with plural support
func (m *i18nManagerImpl) TranslatePlural(key string, count int, params ...interface{}) string {
	return m.TranslatePluralWithLang(m.currentLocale, key, count, params...)
}

// TranslatePluralWithLang translates a key with plural support for a specific locale
func (m *i18nManagerImpl) TranslatePluralWithLang(locale, key string, count int, params ...interface{}) string {
	// Convert params to map format and add count
	paramMap := make(map[string]interface{})
	if len(params) > 0 {
		if pm, ok := params[0].(map[string]interface{}); ok {
			paramMap = pm
		}
	}
	// Always add count to parameters for interpolation
	paramMap["count"] = count

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get plural rule for the locale
	rule := m.pluralRules.GetRule(locale)
	if rule == nil {
		// If no rule found, try default locale
		if locale != m.config.DefaultLocale {
			rule = m.pluralRules.GetRule(m.config.DefaultLocale)
		}
		// If still no rule, use English as ultimate fallback
		if rule == nil {
			rule = EnglishPluralRule
		}
	}

	// Calculate plural form index
	formIndex := rule(count)

	// Build translation key with plural form suffix
	var pluralKey string
	switch formIndex {
	case 0:
		pluralKey = key + ".one"
	case 1:
		pluralKey = key + ".other"
	default:
		// For languages with more than 2 forms (future support)
		pluralKey = fmt.Sprintf("%s.form%d", key, formIndex)
	}

	// Try to get translation for requested locale with plural form
	if translations, ok := m.translations[locale]; ok {
		if translation, ok := translations[pluralKey]; ok {
			return m.interpolate(translation, paramMap)
		}
	}

	// Fallback to default locale if enabled
	if m.config.FallbackToDefault && locale != m.config.DefaultLocale {
		if translations, ok := m.translations[m.config.DefaultLocale]; ok {
			if translation, ok := translations[pluralKey]; ok {
				return m.interpolate(translation, paramMap)
			}
		}
	}

	// If no plural form found, try the base key without suffix
	if translations, ok := m.translations[locale]; ok {
		if translation, ok := translations[key]; ok {
			return m.interpolate(translation, paramMap)
		}
	}

	// Final fallback to default locale base key
	if m.config.FallbackToDefault && locale != m.config.DefaultLocale {
		if translations, ok := m.translations[m.config.DefaultLocale]; ok {
			if translation, ok := translations[key]; ok {
				return m.interpolate(translation, paramMap)
			}
		}
	}

	// Return key if no translation found
	return key
}

// TranslateError translates a framework error
func (m *i18nManagerImpl) TranslateError(err *FrameworkError) string {
	if err.I18nKey != "" {
		return m.Translate(err.I18nKey, err.I18nParams)
	}
	return err.Message
}

// TranslateErrorWithLang translates a framework error with a specific language
func (m *i18nManagerImpl) TranslateErrorWithLang(lang string, err *FrameworkError) string {
	if err.I18nKey != "" {
		return m.TranslateWithLang(lang, err.I18nKey, err.I18nParams)
	}
	return err.Message
}

// ReloadLocales reloads all locale files
func (m *i18nManagerImpl) ReloadLocales() error {
	if m.config.LocalesDir != "" {
		return m.loadLocalesFromDir(m.config.LocalesDir)
	}
	return nil
}

// TranslateForContext translates a key using the context's language preference
func (m *i18nManagerImpl) TranslateForContext(ctx Context, key string, params ...interface{}) string {
	// Use the context's i18n manager if available, otherwise use current locale
	return m.Translate(key, params...)
}

// getLogger returns a logger with i18n support
func (m *i18nManagerImpl) getLogger() *I18nLogger {
	return m.logger
}

// I18nLogger wraps slog.Logger with internationalization support
type I18nLogger struct {
	i18n   I18nManager
	logger *slog.Logger
}

// NewI18nLogger creates a new internationalized logger
func NewI18nLogger(i18n I18nManager, logger *slog.Logger) *I18nLogger {
	if logger == nil {
		logger = slog.Default()
	}

	return &I18nLogger{
		i18n:   i18n,
		logger: logger,
	}
}

// Info logs an info message with i18n support
func (l *I18nLogger) Info(key string, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.Info(message)
}

// InfoWithAttrs logs an info message with i18n support and additional attributes
func (l *I18nLogger) InfoWithAttrs(key string, attrs []slog.Attr, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.LogAttrs(nil, slog.LevelInfo, message, attrs...)
}

// Warn logs a warning message with i18n support
func (l *I18nLogger) Warn(key string, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.Warn(message)
}

// WarnWithAttrs logs a warning message with i18n support and additional attributes
func (l *I18nLogger) WarnWithAttrs(key string, attrs []slog.Attr, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.LogAttrs(nil, slog.LevelWarn, message, attrs...)
}

// Error logs an error message with i18n support
func (l *I18nLogger) Error(key string, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.Error(message)
}

// ErrorWithAttrs logs an error message with i18n support and additional attributes
func (l *I18nLogger) ErrorWithAttrs(key string, attrs []slog.Attr, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.LogAttrs(nil, slog.LevelError, message, attrs...)
}

// Debug logs a debug message with i18n support
func (l *I18nLogger) Debug(key string, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.Debug(message)
}

// DebugWithAttrs logs a debug message with i18n support and additional attributes
func (l *I18nLogger) DebugWithAttrs(key string, attrs []slog.Attr, params ...interface{}) {
	message := l.i18n.Translate(key, params...)
	l.logger.LogAttrs(nil, slog.LevelDebug, message, attrs...)
}

// WithLocale creates a new logger with a specific locale
func (l *I18nLogger) WithLocale(locale string) *I18nLogger {
	// Create a new i18n manager with the specified locale
	newI18n := l.i18n
	if err := newI18n.SetLanguage(locale); err == nil {
		return &I18nLogger{
			i18n:   newI18n,
			logger: l.logger,
		}
	}
	return l
}

// GetUnderlyingLogger returns the underlying slog.Logger
func (l *I18nLogger) GetUnderlyingLogger() *slog.Logger {
	return l.logger
}
