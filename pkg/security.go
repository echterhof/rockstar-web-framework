package pkg

import (
	"time"
)

// SecurityManager defines the security interface for authentication, authorization, and validation
type SecurityManager interface {
	// Authentication methods
	AuthenticateOAuth2(token string) (*User, error)
	AuthenticateJWT(token string) (*User, error)
	AuthenticateAccessToken(token string) (*AccessToken, error)

	// Authorization methods
	Authorize(user *User, resource string, action string) bool
	AuthorizeRole(user *User, role string) bool
	AuthorizeAction(user *User, action string) bool

	// Request validation
	ValidateRequest(ctx Context) error
	ValidateRequestSize(ctx Context, maxSize int64) error
	ValidateRequestTimeout(ctx Context, timeout time.Duration) error
	ValidateBogusData(ctx Context) error

	// Form and file validation
	ValidateFormData(ctx Context, rules ValidationRules) error
	ValidateFileUpload(ctx Context, rules FileValidationRules) error
	ValidateExpectedFormValues(ctx Context, expectedFields []string) error
	ValidateExpectedFiles(ctx Context, expectedFiles []string) error

	// Security headers and protection
	SetSecurityHeaders(ctx Context) error
	SetXFrameOptions(ctx Context, option string) error
	EnableCORS(ctx Context, config CORSConfig) error
	EnableXSSProtection(ctx Context) error
	EnableCSRFProtection(ctx Context) (string, error)
	ValidateCSRFToken(ctx Context, token string) error

	// Input validation
	ValidateInput(input string, rules InputValidationRules) error
	SanitizeInput(input string) string

	// Rate limiting
	CheckRateLimit(ctx Context, resource string) error
	CheckGlobalRateLimit(ctx Context) error

	// Cookie encryption
	EncryptCookie(value string) (string, error)
	DecryptCookie(encryptedValue string) (string, error)
}

// ValidationRules defines form validation rules
type ValidationRules struct {
	Required []string                   `json:"required"`
	Types    map[string]string          `json:"types"`    // field -> type (string, int, email, etc.)
	Lengths  map[string]LengthRule      `json:"lengths"`  // field -> length constraints
	Patterns map[string]string          `json:"patterns"` // field -> regex pattern
	Custom   map[string]CustomValidator `json:"-"`        // field -> custom validator
}

// LengthRule defines length constraints for a field
type LengthRule struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// CustomValidator is a function type for custom validation
type CustomValidator func(value interface{}) error

// FileValidationRules defines file upload validation rules
type FileValidationRules struct {
	MaxSize      int64    `json:"max_size"`      // Maximum file size in bytes
	AllowedTypes []string `json:"allowed_types"` // Allowed MIME types
	AllowedExts  []string `json:"allowed_exts"`  // Allowed file extensions
	Required     []string `json:"required"`      // Required file fields
	MaxFiles     int      `json:"max_files"`     // Maximum number of files
}

// InputValidationRules defines input validation and sanitization rules
type InputValidationRules struct {
	AllowHTML      bool     `json:"allow_html"`
	AllowedTags    []string `json:"allowed_tags"`
	MaxLength      int      `json:"max_length"`
	Pattern        string   `json:"pattern"` // Regex pattern
	StripTags      bool     `json:"strip_tags"`
	EscapeHTML     bool     `json:"escape_html"`
	TrimWhitespace bool     `json:"trim_whitespace"`
}

// CORSConfig defines CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}
