package pkg

import (
	"context"
	"time"
)

// Context represents the unified request context providing access to all framework features
type Context interface {
	// Request and Response data
	Request() *Request
	Response() ResponseWriter
	Params() map[string]string
	Param(name string) string
	Query() map[string]string
	Headers() map[string]string
	Body() []byte

	// Session and authentication
	Session() SessionManager
	User() *User
	Tenant() *Tenant

	// Database and cache
	DB() DatabaseManager
	Cache() CacheManager

	// Configuration and internationalization
	Config() ConfigManager
	I18n() I18nManager

	// File operations
	Files() FileManager

	// Utilities
	Logger() Logger
	Metrics() MetricsCollector

	// Context control
	Context() context.Context
	WithTimeout(timeout time.Duration) Context
	WithCancel() (Context, context.CancelFunc)

	// Response helpers
	JSON(statusCode int, data interface{}) error
	XML(statusCode int, data interface{}) error
	HTML(statusCode int, template string, data interface{}) error
	String(statusCode int, message string) error
	Redirect(statusCode int, url string) error

	// Cookie management
	SetCookie(cookie *Cookie) error
	GetCookie(name string) (*Cookie, error)

	// Header management
	SetHeader(key, value string)
	GetHeader(key string) string

	// Form and file handling
	FormValue(key string) string
	FormFile(key string) (*FormFile, error)

	// Security
	IsAuthenticated() bool
	IsAuthorized(resource, action string) bool

	// Context extension (for plugins)
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
}
