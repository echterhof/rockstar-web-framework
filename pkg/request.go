package pkg

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// Request represents an HTTP request with additional framework features
type Request struct {
	// Standard HTTP request fields
	Method     string
	URL        *url.URL
	Proto      string
	Header     http.Header
	Body       io.ReadCloser
	Host       string
	RemoteAddr string
	RequestURI string

	// Framework-specific fields
	ID        string               // Unique request ID
	TenantID  string               // Multi-tenancy support
	UserID    string               // Authenticated user ID
	StartTime time.Time            // Request start time
	Params    map[string]string    // Route parameters
	Query     map[string]string    // Query parameters
	Form      map[string]string    // Form data
	Files     map[string]*FormFile // Uploaded files

	// Security context
	AccessToken string // API access token
	SessionID   string // Session identifier

	// Protocol information
	IsWebSocket bool   // WebSocket upgrade request
	Protocol    string // HTTP/1, HTTP/2, QUIC, WebSocket

	// Raw request data
	RawBody []byte // Cached body content
}

// FormFile represents an uploaded file
type FormFile struct {
	Filename string
	Header   map[string][]string
	Size     int64
	Content  []byte
}

// Cookie represents an HTTP cookie
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite

	// Framework extensions
	Encrypted bool // Whether cookie value is encrypted
}
