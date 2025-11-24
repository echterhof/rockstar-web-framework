package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Initialize i18n manager with error translations
	i18nConfig := pkg.I18nConfig{
		DefaultLocale:     "en",
		LocalesDir:        "./examples",
		SupportedLocales:  []string{"en", "de"},
		FallbackToDefault: true,
	}

	i18nManager, err := pkg.NewI18nManager(i18nConfig)
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	// Create error handler with i18n support
	errorHandler := pkg.NewErrorHandler(pkg.ErrorHandlerConfig{
		I18n:         i18nManager,
		Logger:       nil, // Would use actual logger in production
		IncludeStack: true,
		LogErrors:    true,
	})

	fmt.Println("=== Rockstar Web Framework - Error Handling Examples ===\n")

	// Example 1: Authentication error
	fmt.Println("1. Authentication Error (English):")
	authErr := pkg.NewAuthenticationError("Invalid credentials")
	authErr.WithContext("req-123", "tenant-1", "user-1", "/api/login", "POST")

	if i18nManager != nil {
		translatedMsg := i18nManager.Translate(authErr.I18nKey, authErr.I18nParams)
		fmt.Printf("   Code: %s\n", authErr.Code)
		fmt.Printf("   Message: %s\n", translatedMsg)
		fmt.Printf("   Status: %d\n\n", authErr.StatusCode)
	}

	// Example 2: Same error in German
	fmt.Println("2. Authentication Error (German):")
	i18nManager.SetLanguage("de")
	translatedMsg := i18nManager.Translate(authErr.I18nKey, authErr.I18nParams)
	fmt.Printf("   Code: %s\n", authErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Status: %d\n\n", authErr.StatusCode)
	i18nManager.SetLanguage("en") // Reset to English

	// Example 3: Validation error with field details
	fmt.Println("3. Validation Error:")
	validationErr := pkg.NewValidationError("Invalid email format", "email")
	validationErr.WithDetails(map[string]interface{}{
		"provided": "invalid-email",
		"expected": "user@example.com",
	})

	translatedMsg = i18nManager.Translate(validationErr.I18nKey, validationErr.I18nParams)
	fmt.Printf("   Code: %s\n", validationErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", validationErr.Details)

	// Example 4: Rate limit error with parameters
	fmt.Println("4. Rate Limit Error:")
	rateLimitErr := pkg.NewRateLimitError("Too many requests", 100, "1m")
	translatedMsg = i18nManager.Translate(rateLimitErr.I18nKey, rateLimitErr.I18nParams)
	fmt.Printf("   Code: %s\n", rateLimitErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", rateLimitErr.Details)

	// Example 5: Request timeout error
	fmt.Println("5. Request Timeout Error:")
	timeoutErr := pkg.NewRequestTimeoutError(30 * time.Second)
	translatedMsg = i18nManager.Translate(timeoutErr.I18nKey, timeoutErr.I18nParams)
	fmt.Printf("   Code: %s\n", timeoutErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Status: %d\n\n", timeoutErr.StatusCode)

	// Example 6: Missing field error
	fmt.Println("6. Missing Field Error:")
	missingFieldErr := pkg.NewMissingFieldError("username")
	translatedMsg = i18nManager.Translate(missingFieldErr.I18nKey, missingFieldErr.I18nParams)
	fmt.Printf("   Code: %s\n", missingFieldErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", missingFieldErr.Details)

	// Example 7: Invalid format error
	fmt.Println("7. Invalid Format Error:")
	formatErr := pkg.NewInvalidFormatError("date", "YYYY-MM-DD")
	translatedMsg = i18nManager.Translate(formatErr.I18nKey, formatErr.I18nParams)
	fmt.Printf("   Code: %s\n", formatErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", formatErr.Details)

	// Example 8: Wrapping generic errors
	fmt.Println("8. Wrapped Generic Error:")
	genericErr := errors.New("database connection failed")
	wrappedErr := pkg.WrapError(genericErr, pkg.ErrCodeDatabaseConnection, http.StatusInternalServerError)
	fmt.Printf("   Code: %s\n", wrappedErr.Code)
	fmt.Printf("   Message: %s\n", wrappedErr.Message)
	fmt.Printf("   Cause: %v\n\n", wrappedErr.Cause)

	// Example 9: Error chaining
	fmt.Println("9. Error Chaining:")
	baseErr := errors.New("connection refused")
	chainedErr := pkg.NewDatabaseError("Failed to connect", "CONNECT").WithCause(baseErr)
	fmt.Printf("   Code: %s\n", chainedErr.Code)
	fmt.Printf("   Message: %s\n", chainedErr.Message)
	fmt.Printf("   Cause: %v\n", chainedErr.Cause)
	fmt.Printf("   Unwrapped: %v\n\n", chainedErr.Unwrap())

	// Example 10: Error response formatting
	fmt.Println("10. Error Response Formatting:")
	response := errorHandler.FormatError(nil, validationErr)
	fmt.Printf("   Response: %+v\n\n", response)

	// Example 11: Not found error
	fmt.Println("11. Not Found Error:")
	notFoundErr := pkg.NewNotFoundError("user")
	translatedMsg = i18nManager.Translate(notFoundErr.I18nKey, notFoundErr.I18nParams)
	fmt.Printf("   Code: %s\n", notFoundErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Status: %d\n\n", notFoundErr.StatusCode)

	// Example 12: Method not allowed error
	fmt.Println("12. Method Not Allowed Error:")
	methodErr := pkg.NewMethodNotAllowedError("DELETE", []string{"GET", "POST"})
	translatedMsg = i18nManager.Translate(methodErr.I18nKey, methodErr.I18nParams)
	fmt.Printf("   Code: %s\n", methodErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", methodErr.Details)

	// Example 13: Configuration error
	fmt.Println("13. Configuration Error:")
	configErr := pkg.NewConfigurationError("database.host", "missing required value")
	translatedMsg = i18nManager.Translate(configErr.I18nKey, configErr.I18nParams)
	fmt.Printf("   Code: %s\n", configErr.Code)
	fmt.Printf("   Message: %s\n", translatedMsg)
	fmt.Printf("   Details: %v\n\n", configErr.Details)

	// Example 14: Recovery from panic
	fmt.Println("14. Panic Recovery:")
	recoveryHandler := pkg.NewRecoveryHandler(nil)

	func() {
		defer func() {
			if r := recover(); r != nil {
				err := recoveryHandler.Recover(nil, r)
				if fwErr, ok := pkg.GetFrameworkError(err); ok {
					fmt.Printf("   Recovered from panic\n")
					fmt.Printf("   Code: %s\n", fwErr.Code)
					fmt.Printf("   Message: %s\n", fwErr.Message)
					fmt.Printf("   Details: %v\n\n", fwErr.Details)
				}
			}
		}()

		panic("simulated panic")
	}()

	// Example 15: Error type checking
	fmt.Println("15. Error Type Checking:")
	testErr := pkg.NewInternalError("test")
	fmt.Printf("   IsFrameworkError: %v\n", pkg.IsFrameworkError(testErr))

	genericTestErr := errors.New("generic")
	fmt.Printf("   IsFrameworkError (generic): %v\n", pkg.IsFrameworkError(genericTestErr))

	if fwErr, ok := pkg.GetFrameworkError(testErr); ok {
		fmt.Printf("   Extracted FrameworkError: %s\n\n", fwErr.Code)
	}

	fmt.Println("=== Examples Complete ===")
}
