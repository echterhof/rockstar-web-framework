package pkg

import (
	"io"
	"net/http"
)

// ResponseWriter provides methods for writing HTTP responses
type ResponseWriter interface {
	// Standard HTTP response methods
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)

	// Framework-specific methods
	WriteJSON(statusCode int, data interface{}) error
	WriteXML(statusCode int, data interface{}) error
	WriteHTML(statusCode int, template string, data interface{}) error
	WriteString(statusCode int, message string) error

	// Stream support
	WriteStream(statusCode int, contentType string, reader io.Reader) error

	// Cookie support
	SetCookie(cookie *Cookie) error

	// Header helpers
	SetHeader(key, value string)
	SetContentType(contentType string)

	// Status and size tracking
	Status() int
	Size() int64
	Written() bool

	// Response control
	Flush() error
	Close() error

	// Template support
	SetTemplateManager(tm TemplateManager)
}

// Response represents the HTTP response data
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Size       int64

	// Framework-specific fields
	Template string      // Template name for HTML responses
	Data     interface{} // Response data
	Cookies  []*Cookie   // Cookies to set

	// Streaming support
	Stream     io.Reader // Stream reader for large responses
	StreamType string    // Content type for streaming
}
