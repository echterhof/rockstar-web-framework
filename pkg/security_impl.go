package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// securityManagerImpl implements the SecurityManager interface
type securityManagerImpl struct {
	db            DatabaseManager
	config        SecurityConfig
	csrfTokens    map[string]time.Time // token -> expiration time
	encryptionKey []byte
	jwtSecret     []byte
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	MaxRequestSize   int64         // Maximum request size in bytes
	RequestTimeout   time.Duration // Request timeout duration
	CSRFTokenExpiry  time.Duration // CSRF token expiry duration
	EncryptionKey    string        // Encryption key for cookies
	JWTSecret        string        // JWT secret key
	XFrameOptions    string        // X-Frame-Options header value
	EnableXSSProtect bool          // Enable XSS protection
	EnableCSRF       bool          // Enable CSRF protection
	AllowedOrigins   []string      // Allowed origins for CORS
}

// NewSecurityManager creates a new security manager instance
func NewSecurityManager(db DatabaseManager, config SecurityConfig) (SecurityManager, error) {
	// Decode encryption key
	encKey, err := hex.DecodeString(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	// Decode JWT secret
	jwtSecret := []byte(config.JWTSecret)

	return &securityManagerImpl{
		db:            db,
		config:        config,
		csrfTokens:    make(map[string]time.Time),
		encryptionKey: encKey,
		jwtSecret:     jwtSecret,
	}, nil
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		MaxRequestSize:   10 * 1024 * 1024, // 10 MB
		RequestTimeout:   30 * time.Second,
		CSRFTokenExpiry:  24 * time.Hour,
		XFrameOptions:    "SAMEORIGIN",
		EnableXSSProtect: true,
		EnableCSRF:       true,
		AllowedOrigins:   []string{"*"},
	}
}

// Request validation methods

// ValidateRequest performs comprehensive request validation
func (s *securityManagerImpl) ValidateRequest(ctx Context) error {
	// Validate request size
	if err := s.ValidateRequestSize(ctx, s.config.MaxRequestSize); err != nil {
		return err
	}

	// Validate request timeout
	if err := s.ValidateRequestTimeout(ctx, s.config.RequestTimeout); err != nil {
		return err
	}

	// Validate bogus data
	if err := s.ValidateBogusData(ctx); err != nil {
		return err
	}

	return nil
}

// ValidateRequestSize validates that request size doesn't exceed the limit
func (s *securityManagerImpl) ValidateRequestSize(ctx Context, maxSize int64) error {
	req := ctx.Request()
	if req == nil {
		return nil
	}

	// Check content length header
	contentLength := req.Header.Get("Content-Length")
	if contentLength != "" {
		var size int64
		fmt.Sscanf(contentLength, "%d", &size)
		if size > maxSize {
			return &FrameworkError{
				Code:       ErrCodeRequestTooLarge,
				Message:    fmt.Sprintf("request size %d exceeds maximum allowed size %d", size, maxSize),
				StatusCode: http.StatusRequestEntityTooLarge,
				I18nKey:    "error.request.too_large",
			}
		}
	}

	// Check actual body size if available
	if req.RawBody != nil && int64(len(req.RawBody)) > maxSize {
		return &FrameworkError{
			Code:       ErrCodeRequestTooLarge,
			Message:    fmt.Sprintf("request body size %d exceeds maximum allowed size %d", len(req.RawBody), maxSize),
			StatusCode: http.StatusRequestEntityTooLarge,
			I18nKey:    "error.request.too_large",
		}
	}

	return nil
}

// ValidateRequestTimeout validates request timeout and cancels if exceeded
func (s *securityManagerImpl) ValidateRequestTimeout(ctx Context, timeout time.Duration) error {
	baseCtx := ctx.Context()
	if baseCtx == nil {
		return nil
	}

	// Check if context has deadline
	deadline, ok := baseCtx.Deadline()
	if !ok {
		// No deadline set, context will handle timeout naturally
		return nil
	}

	// Check if deadline has passed
	if time.Now().After(deadline) {
		return &FrameworkError{
			Code:       ErrCodeRequestTimeout,
			Message:    "request timeout exceeded",
			StatusCode: http.StatusRequestTimeout,
			I18nKey:    "error.request.timeout",
		}
	}

	// Check if context is already cancelled
	select {
	case <-baseCtx.Done():
		return &FrameworkError{
			Code:       ErrCodeRequestTimeout,
			Message:    "request cancelled or timeout exceeded",
			StatusCode: http.StatusRequestTimeout,
			I18nKey:    "error.request.timeout",
		}
	default:
		return nil
	}
}

// ValidateBogusData detects and validates bogus or malformed data
func (s *securityManagerImpl) ValidateBogusData(ctx Context) error {
	req := ctx.Request()
	if req == nil {
		return nil
	}

	// Check for null bytes in URL
	if strings.Contains(req.RequestURI, "\x00") {
		return &FrameworkError{
			Code:       ErrCodeBogusData,
			Message:    "null bytes detected in request URI",
			StatusCode: http.StatusBadRequest,
			I18nKey:    "error.request.bogus_data",
		}
	}

	// Check for suspicious patterns in headers
	for key, values := range req.Header {
		for _, value := range values {
			if strings.Contains(value, "\x00") {
				return &FrameworkError{
					Code:       ErrCodeBogusData,
					Message:    fmt.Sprintf("null bytes detected in header: %s", key),
					StatusCode: http.StatusBadRequest,
					I18nKey:    "error.request.bogus_data",
				}
			}
		}
	}

	// Check body for suspicious patterns if available
	if req.RawBody != nil {
		body := string(req.RawBody)

		// Check for excessive null bytes
		nullCount := strings.Count(body, "\x00")
		if nullCount > 10 {
			return &FrameworkError{
				Code:       ErrCodeBogusData,
				Message:    "excessive null bytes detected in request body",
				StatusCode: http.StatusBadRequest,
				I18nKey:    "error.request.bogus_data",
			}
		}
	}

	return nil
}

// Form and file validation methods

// ValidateFormData validates form data against provided rules
func (s *securityManagerImpl) ValidateFormData(ctx Context, rules ValidationRules) error {
	req := ctx.Request()
	if req == nil || req.Form == nil {
		return fmt.Errorf("no form data available")
	}

	// Check required fields
	for _, field := range rules.Required {
		value, exists := req.Form[field]
		if !exists || value == "" {
			return NewValidationError(fmt.Sprintf("required field missing: %s", field), field)
		}
	}

	// Validate field types
	for field, expectedType := range rules.Types {
		value, exists := req.Form[field]
		if !exists {
			continue
		}

		if err := s.validateFieldType(value, expectedType, field); err != nil {
			return err
		}
	}

	// Validate field lengths
	for field, lengthRule := range rules.Lengths {
		value, exists := req.Form[field]
		if !exists {
			continue
		}

		length := len(value)
		if lengthRule.Min > 0 && length < lengthRule.Min {
			return NewValidationError(
				fmt.Sprintf("field %s length %d is less than minimum %d", field, length, lengthRule.Min),
				field,
			)
		}
		if lengthRule.Max > 0 && length > lengthRule.Max {
			return NewValidationError(
				fmt.Sprintf("field %s length %d exceeds maximum %d", field, length, lengthRule.Max),
				field,
			)
		}
	}

	// Validate patterns
	for field, pattern := range rules.Patterns {
		value, exists := req.Form[field]
		if !exists {
			continue
		}

		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			return fmt.Errorf("invalid regex pattern for field %s: %w", field, err)
		}
		if !matched {
			return NewValidationError(
				fmt.Sprintf("field %s does not match required pattern", field),
				field,
			)
		}
	}

	// Run custom validators
	for field, validator := range rules.Custom {
		value, exists := req.Form[field]
		if !exists {
			continue
		}

		if err := validator(value); err != nil {
			return NewValidationError(
				fmt.Sprintf("field %s failed custom validation: %v", field, err),
				field,
			)
		}
	}

	return nil
}

// validateFieldType validates a field value against expected type
func (s *securityManagerImpl) validateFieldType(value, expectedType, field string) error {
	switch expectedType {
	case "string":
		// Any value is valid as string
		return nil
	case "int":
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err != nil {
			return NewValidationError(fmt.Sprintf("field %s is not a valid integer", field), field)
		}
	case "float":
		var f float64
		if _, err := fmt.Sscanf(value, "%f", &f); err != nil {
			return NewValidationError(fmt.Sprintf("field %s is not a valid float", field), field)
		}
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(value) {
			return NewValidationError(fmt.Sprintf("field %s is not a valid email", field), field)
		}
	case "url":
		urlRegex := regexp.MustCompile(`^https?://[^\s]+$`)
		if !urlRegex.MatchString(value) {
			return NewValidationError(fmt.Sprintf("field %s is not a valid URL", field), field)
		}
	case "bool":
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return NewValidationError(fmt.Sprintf("field %s is not a valid boolean", field), field)
		}
	default:
		return fmt.Errorf("unknown field type: %s", expectedType)
	}

	return nil
}

// ValidateFileUpload validates uploaded files against provided rules
func (s *securityManagerImpl) ValidateFileUpload(ctx Context, rules FileValidationRules) error {
	req := ctx.Request()
	if req == nil {
		return fmt.Errorf("no request available")
	}

	// Check required files
	for _, field := range rules.Required {
		_, exists := req.Files[field]
		if !exists {
			return NewValidationError(fmt.Sprintf("required file missing: %s", field), field)
		}
	}

	// Validate each uploaded file
	fileCount := 0
	for field, file := range req.Files {
		fileCount++

		// Check max files limit
		if rules.MaxFiles > 0 && fileCount > rules.MaxFiles {
			return &FrameworkError{
				Code:       ErrCodeValidationFailed,
				Message:    fmt.Sprintf("too many files uploaded, maximum is %d", rules.MaxFiles),
				StatusCode: http.StatusBadRequest,
				I18nKey:    "error.validation.too_many_files",
			}
		}

		// Check file size
		if rules.MaxSize > 0 && file.Size > rules.MaxSize {
			return &FrameworkError{
				Code:       ErrCodeFileTooLarge,
				Message:    fmt.Sprintf("file %s size %d exceeds maximum %d", field, file.Size, rules.MaxSize),
				StatusCode: http.StatusRequestEntityTooLarge,
				I18nKey:    "error.validation.file_too_large",
				Details:    map[string]interface{}{"field": field, "size": file.Size, "max": rules.MaxSize},
			}
		}

		// Check file extension
		if len(rules.AllowedExts) > 0 {
			ext := getFileExtension(file.Filename)
			if !contains(rules.AllowedExts, ext) {
				return &FrameworkError{
					Code:       ErrCodeInvalidFileType,
					Message:    fmt.Sprintf("file %s has invalid extension %s", field, ext),
					StatusCode: http.StatusBadRequest,
					I18nKey:    "error.validation.invalid_file_type",
					Details:    map[string]interface{}{"field": field, "extension": ext},
				}
			}
		}

		// Check MIME type from header
		if len(rules.AllowedTypes) > 0 {
			contentType := ""
			if ct, ok := file.Header["Content-Type"]; ok && len(ct) > 0 {
				contentType = ct[0]
			}

			if contentType != "" && !contains(rules.AllowedTypes, contentType) {
				return &FrameworkError{
					Code:       ErrCodeInvalidFileType,
					Message:    fmt.Sprintf("file %s has invalid MIME type %s", field, contentType),
					StatusCode: http.StatusBadRequest,
					I18nKey:    "error.validation.invalid_mime_type",
					Details:    map[string]interface{}{"field": field, "mime_type": contentType},
				}
			}
		}
	}

	return nil
}

// ValidateExpectedFormValues validates that expected form fields are present
func (s *securityManagerImpl) ValidateExpectedFormValues(ctx Context, expectedFields []string) error {
	req := ctx.Request()
	if req == nil || req.Form == nil {
		return &FrameworkError{
			Code:       ErrCodeValidationFailed,
			Message:    "no form data available",
			StatusCode: http.StatusBadRequest,
			I18nKey:    "error.validation.no_form_data",
		}
	}

	for _, field := range expectedFields {
		if _, exists := req.Form[field]; !exists {
			return NewValidationError(fmt.Sprintf("expected form field missing: %s", field), field)
		}
	}

	return nil
}

// ValidateExpectedFiles validates that expected files are present
func (s *securityManagerImpl) ValidateExpectedFiles(ctx Context, expectedFiles []string) error {
	req := ctx.Request()
	if req == nil || req.Files == nil {
		return &FrameworkError{
			Code:       ErrCodeValidationFailed,
			Message:    "no files available",
			StatusCode: http.StatusBadRequest,
			I18nKey:    "error.validation.no_files",
		}
	}

	for _, field := range expectedFiles {
		if _, exists := req.Files[field]; !exists {
			return NewValidationError(fmt.Sprintf("expected file missing: %s", field), field)
		}
	}

	return nil
}

// Security headers and protection methods

// SetSecurityHeaders sets all security headers
func (s *securityManagerImpl) SetSecurityHeaders(ctx Context) error {
	// Set X-Frame-Options
	if err := s.SetXFrameOptions(ctx, s.config.XFrameOptions); err != nil {
		return err
	}

	// Enable XSS protection
	if s.config.EnableXSSProtect {
		if err := s.EnableXSSProtection(ctx); err != nil {
			return err
		}
	}

	// Set other security headers
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
	ctx.SetHeader("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	return nil
}

// SetXFrameOptions sets the X-Frame-Options header
func (s *securityManagerImpl) SetXFrameOptions(ctx Context, option string) error {
	validOptions := []string{"DENY", "SAMEORIGIN"}
	if !contains(validOptions, option) && !strings.HasPrefix(option, "ALLOW-FROM ") {
		return fmt.Errorf("invalid X-Frame-Options value: %s", option)
	}

	ctx.SetHeader("X-Frame-Options", option)
	return nil
}

// EnableCORS enables CORS with provided configuration
func (s *securityManagerImpl) EnableCORS(ctx Context, config CORSConfig) error {
	req := ctx.Request()
	origin := req.Header.Get("Origin")

	// Check if origin is allowed
	allowedOrigin := ""
	if contains(config.AllowOrigins, "*") {
		allowedOrigin = "*"
	} else if contains(config.AllowOrigins, origin) {
		allowedOrigin = origin
	}

	if allowedOrigin == "" {
		return &FrameworkError{
			Code:       ErrCodeForbidden,
			Message:    "origin not allowed",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.cors.origin_not_allowed",
		}
	}

	// Set CORS headers
	ctx.SetHeader("Access-Control-Allow-Origin", allowedOrigin)

	if len(config.AllowMethods) > 0 {
		ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
	}

	if len(config.AllowHeaders) > 0 {
		ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
	}

	if len(config.ExposeHeaders) > 0 {
		ctx.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
	}

	if config.AllowCredentials {
		ctx.SetHeader("Access-Control-Allow-Credentials", "true")
	}

	if config.MaxAge > 0 {
		ctx.SetHeader("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
	}

	return nil
}

// EnableXSSProtection enables XSS protection headers
func (s *securityManagerImpl) EnableXSSProtection(ctx Context) error {
	ctx.SetHeader("X-XSS-Protection", "1; mode=block")
	ctx.SetHeader("Content-Security-Policy", "default-src 'self'")
	return nil
}

// EnableCSRFProtection generates and returns a CSRF token
func (s *securityManagerImpl) EnableCSRFProtection(ctx Context) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store token with expiration
	s.csrfTokens[token] = time.Now().Add(s.config.CSRFTokenExpiry)

	// Set token in cookie
	cookie := &Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(s.config.CSRFTokenExpiry.Seconds()),
	}

	if err := ctx.SetCookie(cookie); err != nil {
		return "", fmt.Errorf("failed to set CSRF cookie: %w", err)
	}

	return token, nil
}

// ValidateCSRFToken validates a CSRF token
func (s *securityManagerImpl) ValidateCSRFToken(ctx Context, token string) error {
	if token == "" {
		return &FrameworkError{
			Code:       ErrCodeCSRFTokenInvalid,
			Message:    "CSRF token is empty",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.csrf.token_empty",
		}
	}

	// Check if token exists and is not expired
	expiry, exists := s.csrfTokens[token]
	if !exists {
		return &FrameworkError{
			Code:       ErrCodeCSRFTokenInvalid,
			Message:    "CSRF token not found",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.csrf.token_invalid",
		}
	}

	if time.Now().After(expiry) {
		delete(s.csrfTokens, token)
		return &FrameworkError{
			Code:       ErrCodeCSRFTokenInvalid,
			Message:    "CSRF token expired",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.csrf.token_expired",
		}
	}

	// Get token from cookie
	cookie, err := ctx.GetCookie("csrf_token")
	if err != nil {
		return &FrameworkError{
			Code:       ErrCodeCSRFTokenInvalid,
			Message:    "CSRF cookie not found",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.csrf.cookie_not_found",
		}
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) != 1 {
		return &FrameworkError{
			Code:       ErrCodeCSRFTokenInvalid,
			Message:    "CSRF token mismatch",
			StatusCode: http.StatusForbidden,
			I18nKey:    "error.csrf.token_mismatch",
		}
	}

	return nil
}

// Input validation and sanitization methods

// ValidateInput validates input against provided rules
func (s *securityManagerImpl) ValidateInput(input string, rules InputValidationRules) error {
	// Check max length
	if rules.MaxLength > 0 && len(input) > rules.MaxLength {
		return &FrameworkError{
			Code:       ErrCodeValidationFailed,
			Message:    fmt.Sprintf("input length %d exceeds maximum %d", len(input), rules.MaxLength),
			StatusCode: http.StatusBadRequest,
			I18nKey:    "error.validation.input_too_long",
		}
	}

	// Check pattern
	if rules.Pattern != "" {
		matched, err := regexp.MatchString(rules.Pattern, input)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		if !matched {
			return &FrameworkError{
				Code:       ErrCodeInvalidFormat,
				Message:    "input does not match required pattern",
				StatusCode: http.StatusBadRequest,
				I18nKey:    "error.validation.pattern_mismatch",
			}
		}
	}

	// Check for HTML if not allowed
	if !rules.AllowHTML {
		if containsHTML(input) {
			return &FrameworkError{
				Code:       ErrCodeXSSDetected,
				Message:    "HTML content not allowed",
				StatusCode: http.StatusBadRequest,
				I18nKey:    "error.validation.html_not_allowed",
			}
		}
	}

	// Check for SQL injection patterns
	if containsSQLInjection(input) {
		return &FrameworkError{
			Code:       ErrCodeSQLInjectionDetected,
			Message:    "potential SQL injection detected",
			StatusCode: http.StatusBadRequest,
			I18nKey:    "error.validation.sql_injection",
		}
	}

	return nil
}

// SanitizeInput sanitizes input by removing or escaping dangerous content
func (s *securityManagerImpl) SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Escape HTML entities
	input = html.EscapeString(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	return input
}

// Cookie encryption methods

// EncryptCookie encrypts a cookie value using AES
func (s *securityManagerImpl) EncryptCookie(value string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCookie decrypts an encrypted cookie value
func (s *securityManagerImpl) DecryptCookie(encryptedValue string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode cookie: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cookie: %w", err)
	}

	return string(plaintext), nil
}

// Rate limiting methods

// CheckRateLimit checks rate limit for a specific resource
func (s *securityManagerImpl) CheckRateLimit(ctx Context, resource string) error {
	// Get client identifier (IP address or user ID)
	clientID := s.getClientIdentifier(ctx)

	// Create rate limit key for this resource
	rateLimitKey := fmt.Sprintf("ratelimit:%s:%s", clientID, resource)

	// Default rate limit: 100 requests per minute
	limit := 100
	window := time.Minute

	// Check if rate limit is exceeded
	allowed, err := s.db.CheckRateLimit(rateLimitKey, limit, window)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if !allowed {
		return &FrameworkError{
			Code:       ErrCodeRateLimitExceeded,
			Message:    fmt.Sprintf("rate limit exceeded for resource: %s", resource),
			StatusCode: http.StatusTooManyRequests,
			I18nKey:    "error.rate_limit.exceeded",
			Details:    map[string]interface{}{"resource": resource, "limit": limit, "window": window.String()},
		}
	}

	// Increment rate limit counter
	if err := s.db.IncrementRateLimit(rateLimitKey, window); err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return nil
}

// CheckGlobalRateLimit checks global rate limit
func (s *securityManagerImpl) CheckGlobalRateLimit(ctx Context) error {
	// Get client identifier (IP address or user ID)
	clientID := s.getClientIdentifier(ctx)

	// Create global rate limit key
	rateLimitKey := fmt.Sprintf("ratelimit:global:%s", clientID)

	// Default global rate limit: 1000 requests per hour
	limit := 1000
	window := time.Hour

	// Check if rate limit is exceeded
	allowed, err := s.db.CheckRateLimit(rateLimitKey, limit, window)
	if err != nil {
		return fmt.Errorf("failed to check global rate limit: %w", err)
	}

	if !allowed {
		return &FrameworkError{
			Code:       ErrCodeRateLimitExceeded,
			Message:    "global rate limit exceeded",
			StatusCode: http.StatusTooManyRequests,
			I18nKey:    "error.rate_limit.global_exceeded",
			Details:    map[string]interface{}{"limit": limit, "window": window.String()},
		}
	}

	// Increment rate limit counter
	if err := s.db.IncrementRateLimit(rateLimitKey, window); err != nil {
		return fmt.Errorf("failed to increment global rate limit: %w", err)
	}

	return nil
}

// getClientIdentifier extracts a unique identifier for the client
func (s *securityManagerImpl) getClientIdentifier(ctx Context) string {
	// Try to get authenticated user ID first
	user := ctx.User()
	if user != nil && user.ID != "" {
		return fmt.Sprintf("user:%s", user.ID)
	}

	// Fall back to IP address
	req := ctx.Request()
	if req != nil {
		// Check for X-Forwarded-For header (proxy/load balancer)
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			// Take the first IP in the chain
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				return fmt.Sprintf("ip:%s", strings.TrimSpace(ips[0]))
			}
		}

		// Check for X-Real-IP header
		if xri := req.Header.Get("X-Real-IP"); xri != "" {
			return fmt.Sprintf("ip:%s", xri)
		}

		// Use RemoteAddr as fallback
		if req.RemoteAddr != "" {
			// Remove port if present
			addr := req.RemoteAddr
			if idx := strings.LastIndex(addr, ":"); idx != -1 {
				addr = addr[:idx]
			}
			return fmt.Sprintf("ip:%s", addr)
		}
	}

	// Ultimate fallback
	return "unknown"
}

// Authentication methods (delegated to auth.go)

func (s *securityManagerImpl) AuthenticateOAuth2(token string) (*User, error) {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	return auth.AuthenticateOAuth2(token)
}

func (s *securityManagerImpl) AuthenticateJWT(token string) (*User, error) {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	return auth.AuthenticateJWT(token)
}

func (s *securityManagerImpl) AuthenticateAccessToken(token string) (*AccessToken, error) {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	return auth.AuthenticateAccessToken(token)
}

// Authorization methods (delegated to auth.go)

func (s *securityManagerImpl) Authorize(user *User, resource string, action string) bool {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	// The Authorize method in auth.go expects slices, so we wrap the strings
	err := auth.Authorize(user, []string{resource}, []string{action})
	return err == nil
}

func (s *securityManagerImpl) AuthorizeRole(user *User, role string) bool {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	err := auth.AuthorizeRole(user, role)
	return err == nil
}

func (s *securityManagerImpl) AuthorizeAction(user *User, action string) bool {
	auth := NewAuthManager(s.db, string(s.jwtSecret), OAuth2Config{})
	err := auth.AuthorizeAction(user, action)
	return err == nil
}

// Helper functions

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getFileExtension extracts file extension from filename
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

// containsHTML checks if input contains HTML tags
func containsHTML(input string) bool {
	htmlRegex := regexp.MustCompile(`<[^>]+>`)
	return htmlRegex.MatchString(input)
}

// containsSQLInjection checks for common SQL injection patterns
func containsSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)

	// Common SQL injection patterns
	patterns := []string{
		"' or '1'='1",
		"' or 1=1",
		"'; drop table",
		"'; delete from",
		"union select",
		"exec(",
		"execute(",
		"<script",
		"javascript:",
	}

	for _, pattern := range patterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}

// CleanupExpiredCSRFTokens removes expired CSRF tokens
func (s *securityManagerImpl) CleanupExpiredCSRFTokens() {
	now := time.Now()
	for token, expiry := range s.csrfTokens {
		if now.After(expiry) {
			delete(s.csrfTokens, token)
		}
	}
}
