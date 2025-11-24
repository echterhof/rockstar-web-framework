# Internationalization (i18n) Implementation

## Overview

The Rockstar Web Framework provides comprehensive internationalization (i18n) support for error messages, logging, and application content. The i18n system is built on YAML locale files and integrates seamlessly with Go's `slog` package for internationalized logging.

## Features

- **Multiple Locale Support**: Load and manage translations for multiple languages
- **YAML-Based Locale Files**: Easy-to-maintain locale files in YAML format
- **Parameter Interpolation**: Support for dynamic values in translations using `{{placeholder}}` syntax
- **Fallback Mechanism**: Automatic fallback to default locale when translations are missing
- **Internationalized Logging**: Built-in logger with i18n support using Go's `slog` package
- **Thread-Safe**: Concurrent access to translations is safe
- **Context Integration**: Access i18n through the framework's unified context

## Architecture

### Core Components

1. **I18nManager**: Main interface for managing translations and locales
2. **I18nLogger**: Wrapper around `slog.Logger` with i18n support
3. **Locale Files**: YAML files containing translations (e.g., `locales.en.yaml`, `locales.de.yaml`)

### Translation Key Format

Translations use dot-notation keys that map to nested YAML structures:

```yaml
error:
  authentication:
    failed: "Authentication failed"
```

This becomes the key: `error.authentication.failed`

## Configuration

### I18nConfig Structure

```go
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
```

### Creating an I18n Manager

```go
config := pkg.I18nConfig{
    DefaultLocale:     "en",
    LocalesDir:        "./locales",
    SupportedLocales:  []string{"en", "de", "fr"},
    FallbackToDefault: true,
}

manager, err := pkg.NewI18nManager(config)
if err != nil {
    log.Fatalf("Failed to create i18n manager: %v", err)
}
```

## Locale Files

### File Naming Convention

Locale files must follow this naming pattern:
- `locales.{locale}.yaml` or `locales.{locale}.yml`
- Examples: `locales.en.yaml`, `locales.de.yaml`, `locales.fr.yaml`

### File Structure

Locale files use nested YAML structure:

```yaml
# locales.en.yaml
error:
  authentication:
    failed: "Authentication failed"
    invalid_token: "Invalid authentication token"
  validation:
    missing_field: "Required field '{{field}}' is missing"
    
log:
  server:
    starting: "Server starting on {{address}}"
    stopped: "Server stopped"
    
message:
  welcome: "Welcome to Rockstar Web Framework"
```

### Parameter Placeholders

Use `{{parameter_name}}` syntax for dynamic values:

```yaml
error:
  rate_limit: "Rate limit exceeded: {{limit}} requests per {{window}}"
```

## Usage

### Basic Translation

```go
// Translate a key
message := manager.Translate("error.authentication.failed")
// Returns: "Authentication failed"

// Translate with parameters
message := manager.Translate("error.validation.missing_field", map[string]interface{}{
    "field": "email",
})
// Returns: "Required field 'email' is missing"
```

### Changing Locale

```go
// Set current locale
err := manager.SetLocale("de")
if err != nil {
    log.Printf("Failed to set locale: %v", err)
}

// Get current locale
currentLocale := manager.GetLocale()
```

### Checking Available Locales

```go
locales := manager.GetAvailableLocales()
// Returns: []string{"en", "de", "fr"}
```

### Adding Translations Programmatically

```go
err := manager.AddTranslation("en", "custom.key", "Custom message")
if err != nil {
    log.Printf("Failed to add translation: %v", err)
}
```

## Internationalized Logging

### Getting an I18n Logger

```go
logger := manager.GetLogger()
```

### Logging with I18n

```go
// Simple log message
logger.Info("log.server.starting", map[string]interface{}{
    "address": "localhost:8080",
})

// Log with additional attributes
attrs := []slog.Attr{
    slog.String("request_id", "req-12345"),
    slog.String("tenant_id", "tenant-abc"),
}
logger.InfoWithAttrs("log.request.completed", attrs, map[string]interface{}{
    "method": "GET",
    "path":   "/api/users",
    "status": 200,
})
```

### Log Levels

The I18nLogger supports all standard log levels:

```go
logger.Debug("log.debug.message")
logger.Info("log.info.message")
logger.Warn("log.warn.message")
logger.Error("log.error.message")
```

## Integration with Framework Errors

Framework errors support i18n through the `I18nKey` and `I18nParams` fields:

```go
// Create error with i18n support
err := pkg.NewValidationError("Validation failed", "email")
err.I18nKey = "error.validation.missing_field"
err.I18nParams = map[string]interface{}{"field": "email"}

// Translate error message
translatedMsg := manager.Translate(err.I18nKey, err.I18nParams)
```

## Context Integration

In request handlers, access i18n through the context:

```go
func MyHandler(ctx pkg.Context) error {
    // Get i18n manager
    i18n := ctx.I18n()
    
    // Translate messages
    message := i18n.Translate("message.welcome")
    
    // Use internationalized logger
    logger := i18n.GetLogger()
    logger.Info("log.request.received")
    
    return ctx.JSON(200, map[string]string{
        "message": message,
    })
}
```

## Best Practices

### 1. Organize Translation Keys

Use a hierarchical structure for translation keys:

```
error.{category}.{specific_error}
log.{component}.{event}
message.{context}.{message_type}
```

### 2. Use Descriptive Keys

Keys should be self-documenting:

```yaml
# Good
error.authentication.token_expired: "Token has expired"

# Avoid
err.auth.1: "Token has expired"
```

### 3. Consistent Parameter Names

Use consistent parameter names across translations:

```yaml
error.validation.missing_field: "Required field '{{field}}' is missing"
error.validation.invalid_format: "Invalid format for field '{{field}}'"
```

### 4. Provide Fallback Translations

Always provide translations in the default locale (usually English):

```yaml
# locales.en.yaml - Always complete
error:
  all:
    possible:
      errors: "..."

# locales.de.yaml - Can be partial, will fallback to English
error:
  some:
    errors: "..."
```

### 5. Test All Locales

Ensure all locale files are valid YAML and contain expected keys:

```go
func TestAllLocalesHaveRequiredKeys(t *testing.T) {
    requiredKeys := []string{
        "error.authentication.failed",
        "error.validation.failed",
        "log.server.starting",
    }
    
    locales := []string{"en", "de", "fr"}
    
    for _, locale := range locales {
        manager.SetLocale(locale)
        for _, key := range requiredKeys {
            if !manager.HasKey(key) {
                t.Errorf("Locale %s missing key: %s", locale, key)
            }
        }
    }
}
```

## Performance Considerations

### 1. Locale File Loading

Locale files are loaded once during initialization. For production:

```go
// Load all locales at startup
config := pkg.I18nConfig{
    DefaultLocale: "en",
    LocalesDir:    "./locales",
}
manager, _ := pkg.NewI18nManager(config)
```

### 2. Translation Caching

Translations are cached in memory. No disk I/O occurs during translation.

### 3. Thread Safety

The i18n manager uses read-write locks for thread-safe access:
- Multiple goroutines can read translations concurrently
- Write operations (adding translations, changing locale) are serialized

## Error Handling

### Missing Translations

When a translation is not found:

1. If `FallbackToDefault` is true, tries the default locale
2. If still not found, returns the key itself
3. No error is thrown - graceful degradation

```go
// Key doesn't exist
result := manager.Translate("nonexistent.key")
// Returns: "nonexistent.key"
```

### Invalid Locale Files

If a locale file has invalid YAML:

```go
err := manager.LoadLocaleFile("en", "invalid.yaml")
// Returns error with details
```

### Unsupported Locale

If trying to set an unsupported locale:

```go
err := manager.SetLocale("unsupported")
// Returns: error "unsupported locale: unsupported"
```

## Examples

### Complete Example

See `examples/i18n_example.go` for a comprehensive example demonstrating:
- Basic i18n setup
- Loading locale files
- Translation with parameters
- Internationalized logging
- Error messages with i18n
- Multiple locale support

### Running the Example

```bash
go run examples/i18n_example.go
```

## Requirements Satisfied

This implementation satisfies the following requirements:

- **5.1**: Error messages in configured languages
- **5.2**: Internationalized log messages using slog package
- **5.3**: Language definitions loaded from application configuration
- **5.4**: i18n functionality accessible through context
- **5.5**: YAML locale file support (locales.de.yaml, locales.en.yaml)

## API Reference

### I18nManager Interface

```go
type I18nManager interface {
    Translate(key string, params ...map[string]interface{}) string
    TranslateWithLocale(locale, key string, params ...map[string]interface{}) string
    SetLocale(locale string) error
    GetLocale() string
    GetAvailableLocales() []string
    HasKey(key string) bool
    LoadLocaleFile(locale, filepath string) error
    AddTranslation(locale, key, value string) error
    GetLogger() *I18nLogger
}
```

### I18nLogger Methods

```go
type I18nLogger struct {
    Info(key string, params ...map[string]interface{})
    InfoWithAttrs(key string, attrs []slog.Attr, params ...map[string]interface{})
    Warn(key string, params ...map[string]interface{})
    WarnWithAttrs(key string, attrs []slog.Attr, params ...map[string]interface{})
    Error(key string, params ...map[string]interface{})
    ErrorWithAttrs(key string, attrs []slog.Attr, params ...map[string]interface{})
    Debug(key string, params ...map[string]interface{})
    DebugWithAttrs(key string, attrs []slog.Attr, params ...map[string]interface{})
    WithLocale(locale string) *I18nLogger
    GetUnderlyingLogger() *slog.Logger
}
```

## Testing

Run the i18n tests:

```bash
go test -v ./pkg -run TestI18n
```

Run all tests including i18n:

```bash
go test -v ./pkg
```

## Future Enhancements

Potential future improvements:

1. **Pluralization Support**: Handle singular/plural forms
2. **Date/Time Formatting**: Locale-specific date and time formatting
3. **Number Formatting**: Locale-specific number formatting
4. **Hot Reload**: Reload locale files without restarting
5. **Translation Validation**: CLI tool to validate locale files
6. **Missing Translation Detection**: Development mode to detect missing translations
