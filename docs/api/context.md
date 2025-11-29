# Context API Reference

## Overview

The `Context` interface is the primary way to interact with the Rockstar Web Framework during request handling. It provides access to request data, response methods, framework services, and utilities. Every handler function receives a Context instance.

## Interface Definition

```go
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
```

## Request Data Methods

### Request()

Returns the underlying Request object containing all request information.

**Signature:**
```go
Request() *Request
```

**Returns:**
- `*Request` - The request object

**Example:**
```go
router.GET("/info", func(ctx pkg.Context) error {
    req := ctx.Request()
    return ctx.JSON(200, map[string]interface{}{
        "method":      req.Method,
        "url":         req.URL.String(),
        "remote_addr": req.RemoteAddr,
        "host":        req.Host,
    })
})
```

### Response()

Returns the ResponseWriter for advanced response manipulation.

**Signature:**
```go
Response() ResponseWriter
```

**Returns:**
- `ResponseWriter` - The response writer interface

**Example:**
```go
router.GET("/stream", func(ctx pkg.Context) error {
    resp := ctx.Response()
    resp.SetHeader("Content-Type", "text/event-stream")
    resp.WriteHeader(200)
    // Stream data...
    return nil
})
```

### Params()

Returns all route parameters as a map.

**Signature:**
```go
Params() map[string]string
```

**Returns:**
- `map[string]string` - Map of parameter names to values

**Example:**
```go
// Route: /users/:id/posts/:postID
router.GET("/users/:id/posts/:postID", func(ctx pkg.Context) error {
    params := ctx.Params()
    return ctx.JSON(200, params)
    // {"id": "123", "postID": "456"}
})
```

### Param()

Returns a single route parameter by name.

**Signature:**
```go
Param(name string) string
```

**Parameters:**
- `name` - Parameter name from the route definition

**Returns:**
- `string` - Parameter value, or empty string if not found

**Example:**
```go
// Route: /users/:id
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    return ctx.String(200, "User ID: " + userID)
})
```

### Query()

Returns all query string parameters as a map.

**Signature:**
```go
Query() map[string]string
```

**Returns:**
- `map[string]string` - Map of query parameter names to values

**Example:**
```go
// Request: GET /search?q=golang&page=2
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query()
    searchTerm := query["q"]
    page := query["page"]
    return ctx.JSON(200, map[string]interface{}{
        "query": searchTerm,
        "page":  page,
    })
})
```

### Headers()

Returns all request headers as a map with lowercase keys.

**Signature:**
```go
Headers() map[string]string
```

**Returns:**
- `map[string]string` - Map of header names (lowercase) to values

**Example:**
```go
router.GET("/headers", func(ctx pkg.Context) error {
    headers := ctx.Headers()
    return ctx.JSON(200, headers)
})
```

### Body()

Returns the raw request body as bytes.

**Signature:**
```go
Body() []byte
```

**Returns:**
- `[]byte` - Raw request body

**Example:**
```go
router.POST("/webhook", func(ctx pkg.Context) error {
    body := ctx.Body()
    // Process raw body
    return ctx.String(200, "Received")
})
```

## Service Access Methods

### Session()

Returns the SessionManager for session operations.

**Signature:**
```go
Session() SessionManager
```

**Returns:**
- `SessionManager` - Session management interface

**Example:**
```go
router.POST("/cart/add", func(ctx pkg.Context) error {
    session := ctx.Session()
    sess, _ := session.GetSessionFromCookie(ctx)
    session.Set(sess.ID, "cart_item", "product_123")
    return ctx.JSON(200, map[string]string{"status": "added"})
})
```

### User()

Returns the authenticated user, or nil if not authenticated.

**Signature:**
```go
User() *User
```

**Returns:**
- `*User` - Authenticated user object, or nil

**Example:**
```go
router.GET("/profile", func(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    return ctx.JSON(200, user)
})
```

### Tenant()

Returns the current tenant in multi-tenant applications, or nil if not set.

**Signature:**
```go
Tenant() *Tenant
```

**Returns:**
- `*Tenant` - Current tenant object, or nil

**Example:**
```go
router.GET("/data", func(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    if tenant == nil {
        return ctx.JSON(400, map[string]string{"error": "No tenant"})
    }
    return ctx.JSON(200, map[string]interface{}{
        "tenant": tenant.Name,
    })
})
```

### DB()

Returns the DatabaseManager for database operations.

**Signature:**
```go
DB() DatabaseManager
```

**Returns:**
- `DatabaseManager` - Database management interface

**Example:**
```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    db := ctx.DB()
    userID := ctx.Param("id")
    
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "Not found"})
    }
    return ctx.JSON(200, user)
})
```

### Cache()

Returns the CacheManager for caching operations.

**Signature:**
```go
Cache() CacheManager
```

**Returns:**
- `CacheManager` - Cache management interface

**Example:**
```go
router.GET("/products/:id", func(ctx pkg.Context) error {
    cache := ctx.Cache()
    productID := ctx.Param("id")
    cacheKey := "product:" + productID
    
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Fetch from database and cache
    // ...
    return ctx.JSON(200, product)
})
```

### Config()

Returns the ConfigManager for accessing configuration.

**Signature:**
```go
Config() ConfigManager
```

**Returns:**
- `ConfigManager` - Configuration management interface

**Example:**
```go
router.GET("/settings", func(ctx pkg.Context) error {
    config := ctx.Config()
    appName := config.GetString("app_name")
    maxSize := config.GetInt("max_upload_size")
    return ctx.JSON(200, map[string]interface{}{
        "app_name": appName,
        "max_size": maxSize,
    })
})
```

### I18n()

Returns the I18nManager for internationalization.

**Signature:**
```go
I18n() I18nManager
```

**Returns:**
- `I18nManager` - Internationalization management interface

**Example:**
```go
router.GET("/welcome", func(ctx pkg.Context) error {
    i18n := ctx.I18n()
    welcome := i18n.Translate("welcome.message")
    return ctx.String(200, welcome)
})
```

### Files()

Returns the FileManager for file operations.

**Signature:**
```go
Files() FileManager
```

**Returns:**
- `FileManager` - File management interface

**Example:**
```go
router.POST("/upload", func(ctx pkg.Context) error {
    files := ctx.Files()
    file, _ := ctx.FormFile("file")
    err := files.SaveUploadedFile(ctx, file.Filename, "/uploads/"+file.Filename)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Upload failed"})
    }
    return ctx.JSON(200, map[string]string{"status": "uploaded"})
})
```

### Logger()

Returns the Logger for logging operations.

**Signature:**
```go
Logger() Logger
```

**Returns:**
- `Logger` - Logger interface

**Example:**
```go
router.POST("/api/data", func(ctx pkg.Context) error {
    logger := ctx.Logger()
    logger.Info("Processing request", "endpoint", "/api/data")
    
    // Process request
    
    logger.Debug("Request completed")
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### Metrics()

Returns the MetricsCollector for recording metrics.

**Signature:**
```go
Metrics() MetricsCollector
```

**Returns:**
- `MetricsCollector` - Metrics collection interface

**Example:**
```go
router.GET("/api/expensive", func(ctx pkg.Context) error {
    start := time.Now()
    metrics := ctx.Metrics()
    
    // Expensive operation
    
    duration := time.Since(start)
    metrics.RecordTiming("expensive_op", duration, nil)
    return ctx.JSON(200, map[string]string{"status": "done"})
})
```

## Context Control Methods

### Context()

Returns the underlying Go context.Context for use with context-aware libraries.

**Signature:**
```go
Context() context.Context
```

**Returns:**
- `context.Context` - Standard Go context

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    stdCtx := ctx.Context()
    
    // Check for cancellation
    select {
    case <-stdCtx.Done():
        return stdCtx.Err()
    default:
        // Continue processing
    }
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

### WithTimeout()

Creates a new Context with a timeout. The new context shares the same request data but has an independent timeout.

**Signature:**
```go
WithTimeout(timeout time.Duration) Context
```

**Parameters:**
- `timeout` - Duration after which the context will be cancelled

**Returns:**
- `Context` - New context with timeout

**Example:**
```go
router.GET("/api/slow", func(ctx pkg.Context) error {
    timeoutCtx := ctx.WithTimeout(5 * time.Second)
    
    done := make(chan error, 1)
    go func() {
        // Long operation
        time.Sleep(3 * time.Second)
        done <- nil
    }()
    
    select {
    case err := <-done:
        return ctx.JSON(200, map[string]string{"status": "completed"})
    case <-timeoutCtx.Context().Done():
        return ctx.JSON(504, map[string]string{"error": "timeout"})
    }
})
```

### WithCancel()

Creates a new Context with cancellation capability. Returns the new context and a cancel function.

**Signature:**
```go
WithCancel() (Context, context.CancelFunc)
```

**Returns:**
- `Context` - New cancellable context
- `context.CancelFunc` - Function to cancel the context

**Example:**
```go
router.GET("/api/cancellable", func(ctx pkg.Context) error {
    cancelCtx, cancel := ctx.WithCancel()
    defer cancel()
    
    done := make(chan error, 1)
    go func() {
        select {
        case <-cancelCtx.Context().Done():
            done <- cancelCtx.Context().Err()
            return
        default:
            // Process
        }
        done <- nil
    }()
    
    err := <-done
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "cancelled"})
    }
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

## Response Methods

### JSON()

Sends a JSON response with the specified status code.

**Signature:**
```go
JSON(statusCode int, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code (e.g., 200, 404, 500)
- `data` - Data to serialize as JSON

**Returns:**
- `error` - Error if serialization or writing fails

**Example:**
```go
router.GET("/users", func(ctx pkg.Context) error {
    users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
    return ctx.JSON(200, map[string]interface{}{
        "users": users,
        "total": len(users),
    })
})
```

### XML()

Sends an XML response with the specified status code.

**Signature:**
```go
XML(statusCode int, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `data` - Data to serialize as XML

**Returns:**
- `error` - Error if serialization or writing fails

**Example:**
```go
type User struct {
    ID   int    `xml:"id"`
    Name string `xml:"name"`
}

router.GET("/users.xml", func(ctx pkg.Context) error {
    user := User{ID: 1, Name: "Alice"}
    return ctx.XML(200, user)
})
```

### HTML()

Sends an HTML response using a template.

**Signature:**
```go
HTML(statusCode int, template string, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `template` - Template name or inline HTML
- `data` - Data to pass to the template

**Returns:**
- `error` - Error if template rendering or writing fails

**Example:**
```go
router.GET("/", func(ctx pkg.Context) error {
    return ctx.HTML(200, "index.html", map[string]interface{}{
        "Title": "Welcome",
        "User":  "Alice",
    })
})
```

### String()

Sends a plain text response.

**Signature:**
```go
String(statusCode int, message string) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `message` - Text message to send

**Returns:**
- `error` - Error if writing fails

**Example:**
```go
router.GET("/health", func(ctx pkg.Context) error {
    return ctx.String(200, "OK")
})
```

### Redirect()

Sends a redirect response to the specified URL.

**Signature:**
```go
Redirect(statusCode int, url string) error
```

**Parameters:**
- `statusCode` - HTTP redirect status code (301, 302, 303, 307, 308)
- `url` - Target URL

**Returns:**
- `error` - Error if writing fails

**Example:**
```go
router.GET("/old-path", func(ctx pkg.Context) error {
    return ctx.Redirect(301, "/new-path")
})

router.GET("/login-required", func(ctx pkg.Context) error {
    return ctx.Redirect(302, "/login")
})
```

## Cookie Methods

### SetCookie()

Sets a cookie in the response.

**Signature:**
```go
SetCookie(cookie *Cookie) error
```

**Parameters:**
- `cookie` - Cookie object to set

**Returns:**
- `error` - Error if setting cookie fails

**Example:**
```go
router.POST("/login", func(ctx pkg.Context) error {
    cookie := &pkg.Cookie{
        Name:     "session_id",
        Value:    "abc123",
        Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    ctx.SetCookie(cookie)
    return ctx.JSON(200, map[string]string{"status": "logged in"})
})
```

### GetCookie()

Retrieves a cookie from the request by name.

**Signature:**
```go
GetCookie(name string) (*Cookie, error)
```

**Parameters:**
- `name` - Cookie name

**Returns:**
- `*Cookie` - Cookie object if found
- `error` - http.ErrNoCookie if not found

**Example:**
```go
router.GET("/profile", func(ctx pkg.Context) error {
    cookie, err := ctx.GetCookie("session_id")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    sessionID := cookie.Value
    // Validate session
    return ctx.JSON(200, map[string]interface{}{"session": sessionID})
})
```

## Header Methods

### SetHeader()

Sets a response header.

**Signature:**
```go
SetHeader(key, value string)
```

**Parameters:**
- `key` - Header name
- `value` - Header value

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    ctx.SetHeader("X-Custom-Header", "CustomValue")
    ctx.SetHeader("Cache-Control", "no-cache")
    return ctx.JSON(200, map[string]string{"data": "value"})
})
```

### GetHeader()

Retrieves a request header by name (case-insensitive).

**Signature:**
```go
GetHeader(key string) string
```

**Parameters:**
- `key` - Header name

**Returns:**
- `string` - Header value, or empty string if not found

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    contentType := ctx.GetHeader("Content-Type")
    auth := ctx.GetHeader("Authorization")
    return ctx.JSON(200, map[string]interface{}{
        "content_type": contentType,
        "has_auth":     auth != "",
    })
})
```

## Form and File Methods

### FormValue()

Retrieves a form value by key from POST/PUT request body.

**Signature:**
```go
FormValue(key string) string
```

**Parameters:**
- `key` - Form field name

**Returns:**
- `string` - Form value, or empty string if not found

**Example:**
```go
router.POST("/login", func(ctx pkg.Context) error {
    username := ctx.FormValue("username")
    password := ctx.FormValue("password")
    
    if username == "" || password == "" {
        return ctx.JSON(400, map[string]string{"error": "Missing credentials"})
    }
    
    // Authenticate
    return ctx.JSON(200, map[string]string{"status": "logged in"})
})
```

### FormFile()

Retrieves an uploaded file by form field name.

**Signature:**
```go
FormFile(key string) (*FormFile, error)
```

**Parameters:**
- `key` - Form field name for the file input

**Returns:**
- `*FormFile` - Uploaded file object
- `error` - http.ErrMissingFile if not found

**Example:**
```go
router.POST("/upload", func(ctx pkg.Context) error {
    file, err := ctx.FormFile("file")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "No file uploaded"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
    })
})
```

## Security Methods

### IsAuthenticated()

Checks if the current request has an authenticated user.

**Signature:**
```go
IsAuthenticated() bool
```

**Returns:**
- `bool` - true if user is authenticated, false otherwise

**Example:**
```go
router.GET("/dashboard", func(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return ctx.Redirect(302, "/login")
    }
    
    user := ctx.User()
    return ctx.JSON(200, map[string]interface{}{
        "user": user.Name,
    })
})
```

### IsAuthorized()

Checks if the current user is authorized to perform an action on a resource.

**Signature:**
```go
IsAuthorized(resource, action string) bool
```

**Parameters:**
- `resource` - Resource name (e.g., "posts", "users")
- `action` - Action name (e.g., "read", "write", "delete")

**Returns:**
- `bool` - true if authorized, false otherwise

**Example:**
```go
router.DELETE("/posts/:id", func(ctx pkg.Context) error {
    if !ctx.IsAuthorized("posts", "delete") {
        return ctx.JSON(403, map[string]string{"error": "Forbidden"})
    }
    
    postID := ctx.Param("id")
    // Delete post
    return ctx.JSON(200, map[string]string{"status": "deleted"})
})
```

## Context Extension Methods

### Set()

Stores a custom value in the request context for sharing between middleware and handlers.

**Signature:**
```go
Set(key string, value interface{})
```

**Parameters:**
- `key` - Unique key for the value
- `value` - Value to store (any type)

**Example:**
```go
// Middleware
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    user, _ := validateToken(token)
    ctx.Set("user", user)
    ctx.Set("user_id", user.ID)
    return next(ctx)
}

// Handler
router.GET("/profile", func(ctx pkg.Context) error {
    user, _ := ctx.Get("user")
    return ctx.JSON(200, user)
})
```

### Get()

Retrieves a custom value from the request context.

**Signature:**
```go
Get(key string) (interface{}, bool)
```

**Parameters:**
- `key` - Key of the value to retrieve

**Returns:**
- `interface{}` - The stored value
- `bool` - true if key exists, false otherwise

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    if userID, ok := ctx.Get("user_id"); ok {
        id := userID.(string)
        // Use user ID
    }
    
    if _, exists := ctx.Get("tenant_id"); exists {
        // Tenant-specific logic
    }
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

## Additional Methods (Implementation-Specific)

The following methods are available in the default implementation but not part of the public interface. They are used internally or for testing.

### GetParam()

Gets a route parameter by name (alias for Param).

**Signature:**
```go
GetParam(name string) string
```

### GetParamInt()

Gets a route parameter as an integer.

**Signature:**
```go
GetParamInt(name string) (int, error)
```

**Example:**
```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID, err := ctx.GetParamInt("id")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid ID"})
    }
    return ctx.JSON(200, map[string]interface{}{"user_id": userID})
})
```

### GetQueryParam()

Gets a query parameter by name.

**Signature:**
```go
GetQueryParam(name string) string
```

### GetQueryParamBool()

Gets a query parameter as a boolean.

**Signature:**
```go
GetQueryParamBool(name string) (bool, error)
```

**Example:**
```go
router.GET("/api/data", func(ctx pkg.Context) error {
    verbose, _ := ctx.GetQueryParamBool("verbose")
    if verbose {
        // Return detailed response
    }
    return ctx.JSON(200, map[string]string{"status": "ok"})
})
```

## Factory Function

### NewContext()

Creates a new Context instance. This is typically used internally by the framework.

**Signature:**
```go
func NewContext(req *Request, resp ResponseWriter, ctx context.Context) Context
```

**Parameters:**
- `req` - Request object
- `resp` - Response writer
- `ctx` - Go context

**Returns:**
- `Context` - New context instance

**Example:**
```go
// Internal use
req := &pkg.Request{Method: "GET", URL: url}
resp := pkg.NewResponseWriter(w)
ctx := pkg.NewContext(req, resp, context.Background())
```

## Best Practices

1. **Access Services Through Context:** Always use `ctx.DB()`, `ctx.Cache()`, etc. instead of global variables.

2. **Check Return Values:** Always check the boolean return from `Get()` before type assertion.

3. **Use Appropriate Response Methods:** Use `JSON()` for APIs, `HTML()` for web pages, `String()` for plain text.

4. **Set Headers Before Body:** Call `SetHeader()` before any response method that writes the body.

5. **Handle Errors:** Return errors from handlers instead of panicking.

6. **Share Data via Set/Get:** Use `Set()` and `Get()` to share data between middleware and handlers.

7. **Don't Store Context:** Don't store Context in structs or pass between goroutines. Create new contexts when needed.

8. **Use Timeouts:** Use `WithTimeout()` for operations that might hang.

9. **Check Authentication Early:** Call `IsAuthenticated()` at the start of protected handlers.

10. **Type Assert Safely:** Always use the comma-ok idiom when type asserting values from `Get()`.

## See Also

- [Context Guide](../guides/context.md) - Detailed usage guide with examples
- [Request API](request.md) - Request object reference
- [ResponseWriter API](response.md) - Response writer reference
- [Middleware Guide](../guides/middleware.md) - Middleware patterns
- [Database API](database.md) - Database operations
- [Cache API](cache.md) - Caching operations
