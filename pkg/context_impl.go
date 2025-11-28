package pkg

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// contextImpl is a basic implementation of Context interface
type contextImpl struct {
	request  *Request
	response ResponseWriter
	httpReq  *http.Request
	params   map[string]string
	query    map[string]string
	headers  map[string]string
	ctx      context.Context

	// Managers (will be nil for basic implementation)
	session SessionManager
	db      DatabaseManager
	cache   CacheManager
	config  ConfigManager
	i18n    I18nManager
	files   FileManager
	logger  Logger
	metrics MetricsCollector

	// User context
	user   *User
	tenant *Tenant

	// Plugin context extension storage (request-scoped)
	values map[string]interface{}
}

// Request returns the request object
func (c *contextImpl) Request() *Request {
	return c.request
}

// Response returns the response writer
func (c *contextImpl) Response() ResponseWriter {
	return c.response
}

// Params returns route parameters
func (c *contextImpl) Params() map[string]string {
	return c.params
}

// Param returns a route parameter by name
func (c *contextImpl) Param(name string) string {
	return c.GetParam(name)
}

// Query returns query parameters
func (c *contextImpl) Query() map[string]string {
	// Merge URL query params if not already done
	if c.request != nil && c.request.URL != nil {
		urlQuery := c.request.URL.Query()
		for key, values := range urlQuery {
			if len(values) > 0 {
				// Only add if not already present
				if _, exists := c.query[key]; !exists {
					c.query[key] = values[0]
				}
			}
		}
	}
	return c.query
}

// Headers returns request headers
func (c *contextImpl) Headers() map[string]string {
	if len(c.headers) == 0 && c.request != nil && c.request.Header != nil {
		// Lazy load headers with lowercase keys
		c.headers = make(map[string]string)
		for key, values := range c.request.Header {
			if len(values) > 0 {
				// Normalize to lowercase for case-insensitive access
				c.headers[strings.ToLower(key)] = values[0]
			}
		}
	}
	return c.headers
}

// Body returns the request body
func (c *contextImpl) Body() []byte {
	if c.request != nil {
		return c.request.RawBody
	}
	return nil
}

// Session returns the session manager
func (c *contextImpl) Session() SessionManager {
	return c.session
}

// User returns the authenticated user
func (c *contextImpl) User() *User {
	return c.user
}

// Tenant returns the tenant
func (c *contextImpl) Tenant() *Tenant {
	return c.tenant
}

// DB returns the database manager
func (c *contextImpl) DB() DatabaseManager {
	return c.db
}

// Cache returns the cache manager
func (c *contextImpl) Cache() CacheManager {
	return c.cache
}

// Config returns the configuration manager
func (c *contextImpl) Config() ConfigManager {
	return c.config
}

// I18n returns the internationalization manager
func (c *contextImpl) I18n() I18nManager {
	return c.i18n
}

// Files returns the file manager
func (c *contextImpl) Files() FileManager {
	return c.files
}

// Logger returns the logger
func (c *contextImpl) Logger() Logger {
	return c.logger
}

// Metrics returns the metrics collector
func (c *contextImpl) Metrics() MetricsCollector {
	return c.metrics
}

// Context returns the underlying context
func (c *contextImpl) Context() context.Context {
	if c.ctx == nil {
		if c.httpReq != nil {
			c.ctx = c.httpReq.Context()
		} else {
			c.ctx = context.Background()
		}
	}
	return c.ctx
}

// WithTimeout creates a context with timeout
func (c *contextImpl) WithTimeout(timeout time.Duration) Context {
	ctx, _ := context.WithTimeout(c.Context(), timeout)
	newCtx := *c
	newCtx.ctx = ctx
	// Share the same values map (same request)
	newCtx.values = c.values
	return &newCtx
}

// WithCancel creates a context with cancel
func (c *contextImpl) WithCancel() (Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context())
	newCtx := *c
	newCtx.ctx = ctx
	// Share the same values map (same request)
	newCtx.values = c.values
	return &newCtx, cancel
}

// JSON writes JSON response
func (c *contextImpl) JSON(statusCode int, data interface{}) error {
	return c.response.WriteJSON(statusCode, data)
}

// XML writes XML response
func (c *contextImpl) XML(statusCode int, data interface{}) error {
	return c.response.WriteXML(statusCode, data)
}

// HTML writes HTML response
func (c *contextImpl) HTML(statusCode int, template string, data interface{}) error {
	return c.response.WriteHTML(statusCode, template, data)
}

// String writes string response
func (c *contextImpl) String(statusCode int, message string) error {
	return c.response.WriteString(statusCode, message)
}

// Redirect sends a redirect response
func (c *contextImpl) Redirect(statusCode int, url string) error {
	c.response.SetHeader("Location", url)
	c.response.WriteHeader(statusCode)
	return nil
}

// SetCookie sets a cookie
func (c *contextImpl) SetCookie(cookie *Cookie) error {
	return c.response.SetCookie(cookie)
}

// GetCookie gets a cookie by name
func (c *contextImpl) GetCookie(name string) (*Cookie, error) {
	// Try to get from httpReq first
	if c.httpReq != nil {
		httpCookie, err := c.httpReq.Cookie(name)
		if err == nil {
			return &Cookie{
				Name:     httpCookie.Name,
				Value:    httpCookie.Value,
				Path:     httpCookie.Path,
				Domain:   httpCookie.Domain,
				Expires:  httpCookie.Expires,
				MaxAge:   httpCookie.MaxAge,
				Secure:   httpCookie.Secure,
				HttpOnly: httpCookie.HttpOnly,
				SameSite: httpCookie.SameSite,
			}, nil
		}
	}

	// Fallback: parse from request header
	if c.request != nil && c.request.Header != nil {
		cookieHeader := c.request.Header.Get("Cookie")
		if cookieHeader != "" {
			// Simple cookie parsing
			cookies := strings.Split(cookieHeader, ";")
			for _, cookie := range cookies {
				parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
				if len(parts) == 2 && parts[0] == name {
					return &Cookie{
						Name:  parts[0],
						Value: parts[1],
					}, nil
				}
			}
		}
	}

	return nil, http.ErrNoCookie
}

// SetHeader sets a response header
func (c *contextImpl) SetHeader(key, value string) {
	c.response.SetHeader(key, value)
}

// GetHeader gets a request header
func (c *contextImpl) GetHeader(key string) string {
	if c.request != nil {
		return c.request.Header.Get(key)
	}
	return ""
}

// FormValue gets a form value
func (c *contextImpl) FormValue(key string) string {
	// Ensure form is parsed
	if c.request != nil {
		if c.request.Form == nil {
			parser := NewFormParser()
			parser.ParseForm(c.request)
		}
		if c.request.Form != nil {
			return c.request.Form[key]
		}
	}
	return ""
}

// FormFile gets an uploaded file
func (c *contextImpl) FormFile(key string) (*FormFile, error) {
	// Ensure form is parsed
	if c.request != nil {
		if c.request.Files == nil {
			parser := NewFormParser()
			parser.ParseMultipartForm(c.request, 32<<20) // 32MB default
		}
		if c.request.Files != nil {
			if file, ok := c.request.Files[key]; ok {
				return file, nil
			}
		}
	}
	return nil, http.ErrMissingFile
}

// IsAuthenticated checks if user is authenticated
func (c *contextImpl) IsAuthenticated() bool {
	return c.user != nil
}

// IsAuthorized checks if user is authorized for resource and action
func (c *contextImpl) IsAuthorized(resource, action string) bool {
	// Basic implementation - would integrate with SecurityManager
	return c.user != nil
}

// NewContext creates a new context instance
func NewContext(req *Request, resp ResponseWriter, ctx context.Context) Context {
	return &contextImpl{
		request:  req,
		response: resp,
		ctx:      ctx,
		params:   req.Params,
		query:    req.Query,
		headers:  make(map[string]string),
		values:   make(map[string]interface{}),
	}
}

// SetUser sets the authenticated user (for testing and middleware)
func (c *contextImpl) SetUser(user *User) {
	c.user = user
}

// SetTenant sets the tenant (for testing and middleware)
func (c *contextImpl) SetTenant(tenant *Tenant) {
	c.tenant = tenant
}

// SetCache sets the cache manager (for testing and initialization)
func (c *contextImpl) SetCache(cache CacheManager) {
	c.cache = cache
}

// SetDB sets the database manager (for testing and initialization)
func (c *contextImpl) SetDB(db DatabaseManager) {
	c.db = db
}

// SetSession sets the session manager (for testing and initialization)
func (c *contextImpl) SetSession(session SessionManager) {
	c.session = session
}

// SetConfig sets the config manager (for testing and initialization)
func (c *contextImpl) SetConfig(config ConfigManager) {
	c.config = config
}

// SetI18n sets the i18n manager (for testing and initialization)
func (c *contextImpl) SetI18n(i18n I18nManager) {
	c.i18n = i18n
}

// SetFiles sets the file manager (for testing and initialization)
func (c *contextImpl) SetFiles(files FileManager) {
	c.files = files
}

// SetLogger sets the logger (for testing and initialization)
func (c *contextImpl) SetLogger(logger Logger) {
	c.logger = logger
}

// SetMetrics sets the metrics collector (for testing and initialization)
func (c *contextImpl) SetMetrics(metrics MetricsCollector) {
	c.metrics = metrics
}

// GetParam gets a route parameter by name
func (c *contextImpl) GetParam(name string) string {
	if c.params != nil {
		return c.params[name]
	}
	return ""
}

// GetParamInt gets a route parameter as integer
func (c *contextImpl) GetParamInt(name string) (int, error) {
	value := c.GetParam(name)
	if value == "" {
		return 0, nil
	}
	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	return result, err
}

// GetQueryParam gets a query parameter by name
func (c *contextImpl) GetQueryParam(name string) string {
	if c.query != nil {
		return c.query[name]
	}
	return ""
}

// GetQueryParamBool gets a query parameter as boolean
func (c *contextImpl) GetQueryParamBool(name string) (bool, error) {
	value := c.GetQueryParam(name)
	if value == "" {
		return false, nil
	}
	if value == "true" || value == "1" || value == "yes" {
		return true, nil
	}
	if value == "false" || value == "0" || value == "no" {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean value: %s", value)
}

// Set stores a custom value in the request context
// This allows plugins to share data with handlers and other plugins
func (c *contextImpl) Set(key string, value interface{}) {
	if c.values == nil {
		c.values = make(map[string]interface{})
	}
	c.values[key] = value
}

// Get retrieves a custom value from the request context
// Returns the value and true if found, nil and false otherwise
func (c *contextImpl) Get(key string) (interface{}, bool) {
	if c.values == nil {
		return nil, false
	}
	value, ok := c.values[key]
	return value, ok
}
