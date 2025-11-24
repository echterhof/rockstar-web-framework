package pkg

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FrameworkError represents a framework-specific error with internationalization support
type FrameworkError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Cause      error                  `json:"-"`

	// Internationalization support
	I18nKey    string                 `json:"-"`
	I18nParams map[string]interface{} `json:"-"`

	// Request context
	RequestID string `json:"request_id,omitempty"`
	TenantID  string `json:"tenant_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Path      string `json:"path,omitempty"`
	Method    string `json:"method,omitempty"`
}

// Error implements the error interface
func (e *FrameworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *FrameworkError) Unwrap() error {
	return e.Cause
}

// WithCause adds a cause error
func (e *FrameworkError) WithCause(cause error) *FrameworkError {
	e.Cause = cause
	return e
}

// WithDetails adds additional details
func (e *FrameworkError) WithDetails(details map[string]interface{}) *FrameworkError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithContext adds request context information
func (e *FrameworkError) WithContext(requestID, tenantID, userID, path, method string) *FrameworkError {
	e.RequestID = requestID
	e.TenantID = tenantID
	e.UserID = userID
	e.Path = path
	e.Method = method
	return e
}

// Common framework error codes and constructors
const (
	// Authentication errors
	ErrCodeAuthenticationFailed = "AUTH_FAILED"
	ErrCodeInvalidToken         = "INVALID_TOKEN"
	ErrCodeTokenExpired         = "TOKEN_EXPIRED"
	ErrCodeUnauthorized         = "UNAUTHORIZED"

	// Authorization errors
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeInsufficientRoles   = "INSUFFICIENT_ROLES"
	ErrCodeInsufficientActions = "INSUFFICIENT_ACTIONS"

	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeFileTooLarge     = "FILE_TOO_LARGE"
	ErrCodeInvalidFileType  = "INVALID_FILE_TYPE"

	// Request errors
	ErrCodeRequestTooLarge   = "REQUEST_TOO_LARGE"
	ErrCodeRequestTimeout    = "REQUEST_TIMEOUT"
	ErrCodeBogusData         = "BOGUS_DATA"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	// Security errors
	ErrCodeCSRFTokenInvalid     = "CSRF_TOKEN_INVALID"
	ErrCodeXSSDetected          = "XSS_DETECTED"
	ErrCodeSQLInjectionDetected = "SQL_INJECTION_DETECTED"

	// Database errors
	ErrCodeDatabaseConnection  = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery       = "DATABASE_QUERY"
	ErrCodeDatabaseTransaction = "DATABASE_TRANSACTION"
	ErrCodeRecordNotFound      = "RECORD_NOT_FOUND"
	ErrCodeDuplicateRecord     = "DUPLICATE_RECORD"

	// Session errors
	ErrCodeSessionNotFound = "SESSION_NOT_FOUND"
	ErrCodeSessionExpired  = "SESSION_EXPIRED"
	ErrCodeSessionInvalid  = "SESSION_INVALID"

	// Configuration errors
	ErrCodeConfigurationError   = "CONFIGURATION_ERROR"
	ErrCodeMissingConfiguration = "MISSING_CONFIGURATION"

	// Server errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeNotImplemented     = "NOT_IMPLEMENTED"

	// Multi-tenancy errors
	ErrCodeTenantNotFound      = "TENANT_NOT_FOUND"
	ErrCodeTenantInactive      = "TENANT_INACTIVE"
	ErrCodeTenantLimitExceeded = "TENANT_LIMIT_EXCEEDED"

	// WebSocket errors
	ErrCodeWebSocketUpgradeFailed    = "WEBSOCKET_UPGRADE_FAILED"
	ErrCodeWebSocketConnectionClosed = "WEBSOCKET_CONNECTION_CLOSED"
	ErrCodeWebSocketAuthRequired     = "WEBSOCKET_AUTH_REQUIRED"
	ErrCodeWebSocketInvalidMessage   = "WEBSOCKET_INVALID_MESSAGE"

	// Routing errors
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	ErrCodeAuthorizationFailed = "AUTHORIZATION_FAILED"
)

// Error constructors for common scenarios

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeAuthenticationFailed,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		I18nKey:    "error.authentication.failed",
	}
}

// NewAuthorizationError creates an authorization error
func NewAuthorizationError(message string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		I18nKey:    "error.authorization.forbidden",
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, field string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeValidationFailed,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		I18nKey:    "error.validation.failed",
		Details:    map[string]interface{}{"field": field},
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, limit int, window string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeRateLimitExceeded,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
		I18nKey:    "error.rate_limit.exceeded",
		Details: map[string]interface{}{
			"limit":  limit,
			"window": window,
		},
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, operation string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeDatabaseQuery,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		I18nKey:    "error.database.query",
		Details:    map[string]interface{}{"operation": operation},
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		I18nKey:    "error.internal.server",
	}
}

// NewTenantError creates a tenant-related error
func NewTenantError(code, message string, statusCode int) *FrameworkError {
	return &FrameworkError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		I18nKey:    fmt.Sprintf("error.tenant.%s", code),
	}
}

// ErrorHandler defines the interface for handling framework errors
type ErrorHandler interface {
	HandleError(ctx Context, err error) error
	HandlePanic(ctx Context, recovered interface{}) error
	FormatError(ctx Context, err *FrameworkError) interface{}
}

// RecoveryHandler defines the interface for panic recovery
type RecoveryHandler interface {
	Recover(ctx Context, recovered interface{}) error
	ShouldRecover(recovered interface{}) bool
}

// errorHandlerImpl implements the ErrorHandler interface
type errorHandlerImpl struct {
	i18n         I18nManager
	logger       Logger
	includeStack bool
	logErrors    bool
	recovery     RecoveryHandler
}

// ErrorHandlerConfig configures the error handler
type ErrorHandlerConfig struct {
	I18n         I18nManager
	Logger       Logger
	IncludeStack bool
	LogErrors    bool
	Recovery     RecoveryHandler
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(config ErrorHandlerConfig) ErrorHandler {
	if config.Recovery == nil {
		config.Recovery = NewRecoveryHandler(config.Logger)
	}

	return &errorHandlerImpl{
		i18n:         config.I18n,
		logger:       config.Logger,
		includeStack: config.IncludeStack,
		logErrors:    config.LogErrors,
		recovery:     config.Recovery,
	}
}

// HandleError handles framework errors with internationalization
func (h *errorHandlerImpl) HandleError(ctx Context, err error) error {
	if err == nil {
		return nil
	}

	// Convert to FrameworkError if not already
	var fwErr *FrameworkError
	if fe, ok := err.(*FrameworkError); ok {
		fwErr = fe
	} else {
		// Wrap generic error
		fwErr = NewInternalError(err.Error()).WithCause(err)
	}

	// Add context information if available
	if ctx != nil {
		req := ctx.Request()
		if req != nil {
			fwErr.WithContext(
				req.ID,
				req.TenantID,
				req.UserID,
				req.RequestURI,
				req.Method,
			)
		}
	}

	// Log error if enabled
	if h.logErrors && h.logger != nil {
		h.logError(ctx, fwErr)
	}

	// Translate error message if i18n is available
	if h.i18n != nil && fwErr.I18nKey != "" {
		fwErr.Message = h.i18n.Translate(fwErr.I18nKey, fwErr.I18nParams)
	}

	// Send error response
	if ctx != nil {
		return ctx.JSON(fwErr.StatusCode, h.FormatError(ctx, fwErr))
	}

	return fwErr
}

// HandlePanic handles panic recovery
func (h *errorHandlerImpl) HandlePanic(ctx Context, recovered interface{}) error {
	if recovered == nil {
		return nil
	}

	// Use recovery handler
	if h.recovery != nil {
		return h.recovery.Recover(ctx, recovered)
	}

	// Default panic handling
	err := NewInternalError("panic recovered")
	err.Details = map[string]interface{}{
		"panic": recovered,
	}

	return h.HandleError(ctx, err)
}

// FormatError formats an error for response
func (h *errorHandlerImpl) FormatError(ctx Context, err *FrameworkError) interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
		},
	}

	// Add details if present
	if len(err.Details) > 0 {
		response["error"].(map[string]interface{})["details"] = err.Details
	}

	// Add request context in development mode
	if ctx != nil && ctx.Config() != nil && ctx.Config().IsDevelopment() {
		if err.RequestID != "" {
			response["error"].(map[string]interface{})["request_id"] = err.RequestID
		}
		if err.Path != "" {
			response["error"].(map[string]interface{})["path"] = err.Path
		}
		if err.Method != "" {
			response["error"].(map[string]interface{})["method"] = err.Method
		}

		// Include cause in development
		if err.Cause != nil {
			response["error"].(map[string]interface{})["cause"] = err.Cause.Error()
		}
	}

	return response
}

// logError logs an error with context
func (h *errorHandlerImpl) logError(ctx Context, err *FrameworkError) {
	if h.logger == nil {
		return
	}

	fields := []interface{}{
		"code", err.Code,
		"status", err.StatusCode,
	}

	if err.RequestID != "" {
		fields = append(fields, "request_id", err.RequestID)
	}
	if err.TenantID != "" {
		fields = append(fields, "tenant_id", err.TenantID)
	}
	if err.UserID != "" {
		fields = append(fields, "user_id", err.UserID)
	}
	if err.Path != "" {
		fields = append(fields, "path", err.Path)
	}
	if err.Method != "" {
		fields = append(fields, "method", err.Method)
	}
	if err.Cause != nil {
		fields = append(fields, "cause", err.Cause.Error())
	}

	h.logger.Error(err.Message, fields...)
}

// recoveryHandlerImpl implements the RecoveryHandler interface
type recoveryHandlerImpl struct {
	logger Logger
}

// NewRecoveryHandler creates a new recovery handler
func NewRecoveryHandler(logger Logger) RecoveryHandler {
	return &recoveryHandlerImpl{
		logger: logger,
	}
}

// Recover handles panic recovery
func (r *recoveryHandlerImpl) Recover(ctx Context, recovered interface{}) error {
	if recovered == nil {
		return nil
	}

	// Log the panic
	if r.logger != nil {
		r.logger.Error("panic recovered", "panic", recovered)
	}

	// Create error
	err := NewInternalError("internal server error")
	err.Details = map[string]interface{}{
		"recovered": fmt.Sprintf("%v", recovered),
	}

	// Add context information
	if ctx != nil {
		req := ctx.Request()
		if req != nil {
			err.WithContext(req.ID, req.TenantID, req.UserID, req.RequestURI, req.Method)
		}

		// Send error response
		return ctx.JSON(err.StatusCode, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    err.Code,
				"message": err.Message,
			},
		})
	}

	return err
}

// ShouldRecover determines if a panic should be recovered
func (r *recoveryHandlerImpl) ShouldRecover(recovered interface{}) bool {
	// Recover all panics by default
	return true
}

// RecoveryMiddleware creates a middleware that recovers from panics
func RecoveryMiddleware(handler ErrorHandler) MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		defer func() {
			if recovered := recover(); recovered != nil {
				handler.HandlePanic(ctx, recovered)
			}
		}()

		return next(ctx)
	}
}

// ErrorMiddleware creates a middleware that handles errors
func ErrorMiddleware(handler ErrorHandler) MiddlewareFunc {
	return func(ctx Context, next HandlerFunc) error {
		err := next(ctx)
		if err != nil {
			return handler.HandleError(ctx, err)
		}
		return nil
	}
}

// ValidationError creates a validation error with field information
func ValidationError(field, message string) *FrameworkError {
	return NewValidationError(message, field)
}

// WrapError wraps a generic error into a FrameworkError
func WrapError(err error, code string, statusCode int) *FrameworkError {
	if err == nil {
		return nil
	}

	// If already a FrameworkError, return as-is
	if fe, ok := err.(*FrameworkError); ok {
		return fe
	}

	return &FrameworkError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: statusCode,
		Cause:      err,
	}
}

// IsFrameworkError checks if an error is a FrameworkError
func IsFrameworkError(err error) bool {
	_, ok := err.(*FrameworkError)
	return ok
}

// GetFrameworkError extracts a FrameworkError from an error
func GetFrameworkError(err error) (*FrameworkError, bool) {
	fe, ok := err.(*FrameworkError)
	return fe, ok
}

// ErrorResponse creates a standardized error response
func ErrorResponse(code, message string, details map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if len(details) > 0 {
		response["error"].(map[string]interface{})["details"] = details
	}

	return response
}

// Additional error constructors for specific scenarios

// NewRequestTooLargeError creates a request too large error
func NewRequestTooLargeError(maxSize int64) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeRequestTooLarge,
		Message:    "request size exceeds maximum allowed",
		StatusCode: http.StatusRequestEntityTooLarge,
		I18nKey:    "error.request.too_large",
		I18nParams: map[string]interface{}{"max_size": maxSize},
		Details:    map[string]interface{}{"max_size": maxSize},
	}
}

// NewRequestTimeoutError creates a request timeout error
func NewRequestTimeoutError(timeout time.Duration) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeRequestTimeout,
		Message:    "request timeout exceeded",
		StatusCode: http.StatusRequestTimeout,
		I18nKey:    "error.request.timeout",
		I18nParams: map[string]interface{}{"timeout": timeout.String()},
		Details:    map[string]interface{}{"timeout": timeout.String()},
	}
}

// NewBogusDataError creates a bogus data error
func NewBogusDataError(reason string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeBogusData,
		Message:    "invalid or malformed data detected",
		StatusCode: http.StatusBadRequest,
		I18nKey:    "error.request.bogus_data",
		I18nParams: map[string]interface{}{"reason": reason},
		Details:    map[string]interface{}{"reason": reason},
	}
}

// NewMissingFieldError creates a missing field error
func NewMissingFieldError(field string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeMissingField,
		Message:    fmt.Sprintf("required field '%s' is missing", field),
		StatusCode: http.StatusBadRequest,
		I18nKey:    "error.validation.missing_field",
		I18nParams: map[string]interface{}{"field": field},
		Details:    map[string]interface{}{"field": field},
	}
}

// NewInvalidFormatError creates an invalid format error
func NewInvalidFormatError(field, expectedFormat string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeInvalidFormat,
		Message:    fmt.Sprintf("field '%s' has invalid format, expected: %s", field, expectedFormat),
		StatusCode: http.StatusBadRequest,
		I18nKey:    "error.validation.invalid_format",
		I18nParams: map[string]interface{}{"field": field, "format": expectedFormat},
		Details:    map[string]interface{}{"field": field, "expected_format": expectedFormat},
	}
}

// NewSessionError creates a session-related error
func NewSessionError(code, message string) *FrameworkError {
	statusCode := http.StatusUnauthorized
	if code == ErrCodeSessionExpired {
		statusCode = http.StatusUnauthorized
	}

	return &FrameworkError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		I18nKey:    fmt.Sprintf("error.session.%s", strings.ToLower(strings.TrimPrefix(code, "SESSION_"))),
	}
}

// NewWebSocketError creates a WebSocket-related error
func NewWebSocketError(code, message string) *FrameworkError {
	return &FrameworkError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		I18nKey:    fmt.Sprintf("error.websocket.%s", strings.ToLower(strings.TrimPrefix(code, "WEBSOCKET_"))),
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		I18nKey:    "error.not_found",
		I18nParams: map[string]interface{}{"resource": resource},
		Details:    map[string]interface{}{"resource": resource},
	}
}

// NewMethodNotAllowedError creates a method not allowed error
func NewMethodNotAllowedError(method string, allowedMethods []string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeMethodNotAllowed,
		Message:    fmt.Sprintf("method %s not allowed", method),
		StatusCode: http.StatusMethodNotAllowed,
		I18nKey:    "error.method_not_allowed",
		I18nParams: map[string]interface{}{"method": method, "allowed": allowedMethods},
		Details:    map[string]interface{}{"method": method, "allowed_methods": allowedMethods},
	}
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(key, reason string) *FrameworkError {
	return &FrameworkError{
		Code:       ErrCodeConfigurationError,
		Message:    fmt.Sprintf("configuration error for '%s': %s", key, reason),
		StatusCode: http.StatusInternalServerError,
		I18nKey:    "error.configuration.error",
		I18nParams: map[string]interface{}{"key": key, "reason": reason},
		Details:    map[string]interface{}{"key": key, "reason": reason},
	}
}
