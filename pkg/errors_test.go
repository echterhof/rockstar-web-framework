package pkg

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

// TestFrameworkError tests the FrameworkError structure
func TestFrameworkError(t *testing.T) {
	t.Run("Error method returns formatted message", func(t *testing.T) {
		err := &FrameworkError{
			Code:    "TEST_ERROR",
			Message: "test error message",
		}

		expected := "TEST_ERROR: test error message"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error method includes cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &FrameworkError{
			Code:    "TEST_ERROR",
			Message: "test error message",
			Cause:   cause,
		}

		if !errors.Is(err, cause) {
			t.Error("expected error to wrap cause")
		}
	})

	t.Run("WithCause adds cause error", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewInternalError("test").WithCause(cause)

		if err.Cause != cause {
			t.Error("expected cause to be set")
		}
	})

	t.Run("WithDetails adds details", func(t *testing.T) {
		err := NewInternalError("test").WithDetails(map[string]interface{}{
			"key": "value",
		})

		if err.Details["key"] != "value" {
			t.Error("expected details to be set")
		}
	})

	t.Run("WithContext adds context information", func(t *testing.T) {
		err := NewInternalError("test").WithContext(
			"req-123",
			"tenant-456",
			"user-789",
			"/api/test",
			"GET",
		)

		if err.RequestID != "req-123" {
			t.Errorf("expected request ID 'req-123', got %q", err.RequestID)
		}
		if err.TenantID != "tenant-456" {
			t.Errorf("expected tenant ID 'tenant-456', got %q", err.TenantID)
		}
		if err.UserID != "user-789" {
			t.Errorf("expected user ID 'user-789', got %q", err.UserID)
		}
		if err.Path != "/api/test" {
			t.Errorf("expected path '/api/test', got %q", err.Path)
		}
		if err.Method != "GET" {
			t.Errorf("expected method 'GET', got %q", err.Method)
		}
	})
}

// TestErrorConstructors tests error constructor functions
func TestErrorConstructors(t *testing.T) {
	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError("invalid credentials")

		if err.Code != ErrCodeAuthenticationFailed {
			t.Errorf("expected code %q, got %q", ErrCodeAuthenticationFailed, err.Code)
		}
		if err.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, err.StatusCode)
		}
		if err.I18nKey != "error.authentication.failed" {
			t.Errorf("expected i18n key 'error.authentication.failed', got %q", err.I18nKey)
		}
	})

	t.Run("NewAuthorizationError", func(t *testing.T) {
		err := NewAuthorizationError("insufficient permissions")

		if err.Code != ErrCodeForbidden {
			t.Errorf("expected code %q, got %q", ErrCodeForbidden, err.Code)
		}
		if err.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, err.StatusCode)
		}
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("invalid email", "email")

		if err.Code != ErrCodeValidationFailed {
			t.Errorf("expected code %q, got %q", ErrCodeValidationFailed, err.Code)
		}
		if err.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
		}
		if err.Details["field"] != "email" {
			t.Error("expected field detail to be set")
		}
	})

	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError("rate limit exceeded", 100, "1m")

		if err.Code != ErrCodeRateLimitExceeded {
			t.Errorf("expected code %q, got %q", ErrCodeRateLimitExceeded, err.Code)
		}
		if err.StatusCode != http.StatusTooManyRequests {
			t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, err.StatusCode)
		}
		if err.Details["limit"] != 100 {
			t.Error("expected limit detail to be set")
		}
	})

	t.Run("NewDatabaseError", func(t *testing.T) {
		err := NewDatabaseError("query failed", "SELECT")

		if err.Code != ErrCodeDatabaseQuery {
			t.Errorf("expected code %q, got %q", ErrCodeDatabaseQuery, err.Code)
		}
		if err.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.StatusCode)
		}
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("something went wrong")

		if err.Code != ErrCodeInternalError {
			t.Errorf("expected code %q, got %q", ErrCodeInternalError, err.Code)
		}
		if err.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.StatusCode)
		}
	})

	t.Run("NewRequestTooLargeError", func(t *testing.T) {
		err := NewRequestTooLargeError(1024)

		if err.Code != ErrCodeRequestTooLarge {
			t.Errorf("expected code %q, got %q", ErrCodeRequestTooLarge, err.Code)
		}
		if err.Details["max_size"] != int64(1024) {
			t.Error("expected max_size detail to be set")
		}
	})

	t.Run("NewRequestTimeoutError", func(t *testing.T) {
		err := NewRequestTimeoutError(30 * time.Second)

		if err.Code != ErrCodeRequestTimeout {
			t.Errorf("expected code %q, got %q", ErrCodeRequestTimeout, err.Code)
		}
	})

	t.Run("NewBogusDataError", func(t *testing.T) {
		err := NewBogusDataError("invalid format")

		if err.Code != ErrCodeBogusData {
			t.Errorf("expected code %q, got %q", ErrCodeBogusData, err.Code)
		}
		if err.Details["reason"] != "invalid format" {
			t.Error("expected reason detail to be set")
		}
	})

	t.Run("NewMissingFieldError", func(t *testing.T) {
		err := NewMissingFieldError("username")

		if err.Code != ErrCodeMissingField {
			t.Errorf("expected code %q, got %q", ErrCodeMissingField, err.Code)
		}
		if err.Details["field"] != "username" {
			t.Error("expected field detail to be set")
		}
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("user")

		if err.Code != ErrCodeNotFound {
			t.Errorf("expected code %q, got %q", ErrCodeNotFound, err.Code)
		}
		if err.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, err.StatusCode)
		}
	})
}

// TestErrorHandler tests the error handler implementation
func TestErrorHandler(t *testing.T) {
	t.Run("HandleError with nil error returns nil", func(t *testing.T) {
		handler := NewErrorHandler(ErrorHandlerConfig{})

		err := handler.HandleError(nil, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("HandleError converts generic error to FrameworkError", func(t *testing.T) {
		handler := NewErrorHandler(ErrorHandlerConfig{})

		genericErr := errors.New("generic error")
		err := handler.HandleError(nil, genericErr)

		fwErr, ok := err.(*FrameworkError)
		if !ok {
			t.Fatal("expected FrameworkError")
		}

		if fwErr.Code != ErrCodeInternalError {
			t.Errorf("expected code %q, got %q", ErrCodeInternalError, fwErr.Code)
		}
		if fwErr.Cause != genericErr {
			t.Error("expected cause to be set")
		}
	})

	t.Run("HandleError preserves FrameworkError", func(t *testing.T) {
		handler := NewErrorHandler(ErrorHandlerConfig{})

		originalErr := NewAuthenticationError("test")
		err := handler.HandleError(nil, originalErr)

		if err != originalErr {
			t.Error("expected original error to be preserved")
		}
	})

	t.Run("FormatError creates proper response structure", func(t *testing.T) {
		handler := NewErrorHandler(ErrorHandlerConfig{})

		fwErr := NewValidationError("invalid input", "email")
		response := handler.FormatError(nil, fwErr)

		respMap, ok := response.(map[string]interface{})
		if !ok {
			t.Fatal("expected map response")
		}

		errorMap, ok := respMap["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error map")
		}

		if errorMap["code"] != fwErr.Code {
			t.Error("expected code in response")
		}
		if errorMap["message"] != fwErr.Message {
			t.Error("expected message in response")
		}
		if errorMap["details"] == nil {
			t.Error("expected details in response")
		}
	})
}

// TestRecoveryHandler tests the recovery handler implementation
func TestRecoveryHandler(t *testing.T) {
	t.Run("Recover handles nil panic", func(t *testing.T) {
		handler := NewRecoveryHandler(nil)

		err := handler.Recover(nil, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Recover creates error from panic", func(t *testing.T) {
		handler := NewRecoveryHandler(nil)

		err := handler.Recover(nil, "panic message")
		if err == nil {
			t.Fatal("expected error")
		}

		fwErr, ok := err.(*FrameworkError)
		if !ok {
			t.Fatal("expected FrameworkError")
		}

		if fwErr.Code != ErrCodeInternalError {
			t.Errorf("expected code %q, got %q", ErrCodeInternalError, fwErr.Code)
		}
	})

	t.Run("ShouldRecover returns true by default", func(t *testing.T) {
		handler := NewRecoveryHandler(nil)

		if !handler.ShouldRecover("anything") {
			t.Error("expected ShouldRecover to return true")
		}
	})
}

// TestWrapError tests error wrapping functionality
func TestWrapError(t *testing.T) {
	t.Run("WrapError with nil returns nil", func(t *testing.T) {
		err := WrapError(nil, "TEST", 500)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("WrapError preserves FrameworkError", func(t *testing.T) {
		original := NewInternalError("test")
		wrapped := WrapError(original, "TEST", 500)

		if wrapped != original {
			t.Error("expected original error to be preserved")
		}
	})

	t.Run("WrapError wraps generic error", func(t *testing.T) {
		genericErr := errors.New("generic error")
		wrapped := WrapError(genericErr, "TEST_CODE", http.StatusBadRequest)

		if wrapped.Code != "TEST_CODE" {
			t.Errorf("expected code 'TEST_CODE', got %q", wrapped.Code)
		}
		if wrapped.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, wrapped.StatusCode)
		}
		if wrapped.Cause != genericErr {
			t.Error("expected cause to be set")
		}
	})
}

// TestIsFrameworkError tests error type checking
func TestIsFrameworkError(t *testing.T) {
	t.Run("IsFrameworkError returns true for FrameworkError", func(t *testing.T) {
		err := NewInternalError("test")
		if !IsFrameworkError(err) {
			t.Error("expected IsFrameworkError to return true")
		}
	})

	t.Run("IsFrameworkError returns false for generic error", func(t *testing.T) {
		err := errors.New("generic error")
		if IsFrameworkError(err) {
			t.Error("expected IsFrameworkError to return false")
		}
	})
}

// TestGetFrameworkError tests error extraction
func TestGetFrameworkError(t *testing.T) {
	t.Run("GetFrameworkError extracts FrameworkError", func(t *testing.T) {
		original := NewInternalError("test")
		extracted, ok := GetFrameworkError(original)

		if !ok {
			t.Fatal("expected ok to be true")
		}
		if extracted != original {
			t.Error("expected extracted error to match original")
		}
	})

	t.Run("GetFrameworkError returns false for generic error", func(t *testing.T) {
		err := errors.New("generic error")
		_, ok := GetFrameworkError(err)

		if ok {
			t.Error("expected ok to be false")
		}
	})
}

// TestErrorResponse tests error response formatting
func TestErrorResponse(t *testing.T) {
	t.Run("ErrorResponse creates proper structure", func(t *testing.T) {
		response := ErrorResponse("TEST_CODE", "test message", map[string]interface{}{
			"field": "value",
		})

		errorMap, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error map")
		}

		if errorMap["code"] != "TEST_CODE" {
			t.Error("expected code in response")
		}
		if errorMap["message"] != "test message" {
			t.Error("expected message in response")
		}
		if errorMap["details"] == nil {
			t.Error("expected details in response")
		}
	})

	t.Run("ErrorResponse without details", func(t *testing.T) {
		response := ErrorResponse("TEST_CODE", "test message", nil)

		errorMap, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error map")
		}

		if errorMap["details"] != nil {
			t.Error("expected no details in response")
		}
	})
}

// TestErrorUnwrap tests error unwrapping
func TestErrorUnwrap(t *testing.T) {
	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewInternalError("test").WithCause(cause)

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Error("expected unwrapped error to be cause")
		}
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := NewInternalError("test")

		unwrapped := err.Unwrap()
		if unwrapped != nil {
			t.Error("expected unwrapped error to be nil")
		}
	})
}

// TestValidationError tests validation error helper
func TestValidationError(t *testing.T) {
	t.Run("ValidationError creates proper error", func(t *testing.T) {
		err := ValidationError("email", "invalid email format")

		if err.Code != ErrCodeValidationFailed {
			t.Errorf("expected code %q, got %q", ErrCodeValidationFailed, err.Code)
		}
		if err.Details["field"] != "email" {
			t.Error("expected field detail to be set")
		}
	})
}

// TestNewInvalidFormatError tests invalid format error
func TestNewInvalidFormatError(t *testing.T) {
	t.Run("NewInvalidFormatError creates proper error", func(t *testing.T) {
		err := NewInvalidFormatError("date", "YYYY-MM-DD")

		if err.Code != ErrCodeInvalidFormat {
			t.Errorf("expected code %q, got %q", ErrCodeInvalidFormat, err.Code)
		}
		if err.Details["field"] != "date" {
			t.Error("expected field detail to be set")
		}
		if err.Details["expected_format"] != "YYYY-MM-DD" {
			t.Error("expected format detail to be set")
		}
	})
}

// TestNewMethodNotAllowedError tests method not allowed error
func TestNewMethodNotAllowedError(t *testing.T) {
	t.Run("NewMethodNotAllowedError creates proper error", func(t *testing.T) {
		err := NewMethodNotAllowedError("DELETE", []string{"GET", "POST"})

		if err.Code != ErrCodeMethodNotAllowed {
			t.Errorf("expected code %q, got %q", ErrCodeMethodNotAllowed, err.Code)
		}
		if err.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, err.StatusCode)
		}
		if err.Details["method"] != "DELETE" {
			t.Error("expected method detail to be set")
		}
	})
}

// TestNewConfigurationError tests configuration error
func TestNewConfigurationError(t *testing.T) {
	t.Run("NewConfigurationError creates proper error", func(t *testing.T) {
		err := NewConfigurationError("database.host", "missing required value")

		if err.Code != ErrCodeConfigurationError {
			t.Errorf("expected code %q, got %q", ErrCodeConfigurationError, err.Code)
		}
		if err.Details["key"] != "database.host" {
			t.Error("expected key detail to be set")
		}
		if err.Details["reason"] != "missing required value" {
			t.Error("expected reason detail to be set")
		}
	})
}
