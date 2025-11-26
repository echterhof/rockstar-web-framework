# Context Guide

## Overview

The `Context` interface is the heart of the Rockstar Web Framework. It provides a unified, request-scoped interface to access all framework features, eliminating global state and enabling clean dependency injection. Every handler receives a Context instance that encapsulates the request, response, and all framework services.

## Core Philosophy

- **Request-Scoped**: Each request gets its own Context instance
- **No Global State**: All services accessed through Context, not global variables
- **Dependency Injection**: Framework components injected into Context at creation
- **Type-Safe**: Strongly-typed interface with compile-time guarantees
- **Testable**: Easy to mock and test handlers in isolation

## Context Interface

```go
type Context interface {
    // Request and Response data
    Request() *Request
    Response() ResponseWriter
    Params() map[string]string
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
}
```

## Basic Usage

### Simple Handler

```go
func HelloHandler(ctx pkg.Context) error {
    name := ctx.Query()["name"]
    if name == "" {
        name = "World"
    }
    
    return ctx.JSON(200, map[string]string{
        "message": fmt.Sprintf("Hello, %s!", name),
    })
}
```

### Accessing Request Data

```go
func RequestDataHandler(ctx pkg.Context) error {
    // Route parameters (e.g., /users/:id)
    userID := ctx.Params()["id"]
    
    // Query parameters (e.g., ?page=1&limit=10)
    page := ctx.Query()["page"]
    limit := ctx.Query()["limit"]
    
    // Request headers
    authHeader := ctx.GetHeader("Authorization")
    contentType := ctx.GetHeader("Content-Type")
    
    // Request body
    body := ctx.Body()
    
    // Full request object
    req := ctx.Request()
    method := req.Method
    path := req.Path
    
    return ctx.JSON(200, map[string]interface{}{
        "userID":      userID,
        "page":        page,
        "limit":       limit,
        "auth":        authHeader,
        "contentType": contentType,
        "method":      method,
        "path":        path,
    })
}
```

## Response Methods

### JSON Response

```go
func JSONHandler(ctx pkg.Context) error {
    data := map[string]interface{}{
        "status": "success",
        "data": []string{"item1", "item2", "item3"},
    }
    return ctx.JSON(200, data)
}
```

### XML Response

```go
func XMLHandler(ctx pkg.Context) error {
    type Response struct {
        XMLName xml.Name `xml:"response"`
        Status  string   `xml:"status"`
        Message string   `xml:"message"`
    }
    
    return ctx.XML(200, Response{
        Status:  "success",
        Message: "Operation completed",
    })
}
```

### HTML Response

```go
func HTMLHandler(ctx pkg.Context) error {
    data := map[string]interface{}{
        "Title": "Welcome",
        "User":  ctx.User(),
    }
    return ctx.HTML(200, "index.html", data)
}
```

### String Response

```go
func StringHandler(ctx pkg.Context) error {
    return ctx.String(200, "Plain text response")
}
```

### Redirect

```go
func RedirectHandler(ctx pkg.Context) error {
    return ctx.Redirect(302, "/new-location")
}
```

## Database Access

### Basic Query

```go
func GetUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    var user User
    err := ctx.DB().QueryRow(
        "SELECT id, name, email FROM users WHERE id = ?",
        userID,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return ctx.JSON(404, map[string]string{
            "error": "User not found",
        })
    }
    
    return ctx.JSON(200, user)
}
```

### Transaction

```go
func CreateOrderHandler(ctx pkg.Context) error {
    tx, err := ctx.DB().Begin()
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to start transaction"})
    }
    defer tx.Rollback()
    
    // Insert order
    result, err := tx.Exec("INSERT INTO orders (user_id, total) VALUES (?, ?)", 
        userID, total)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to create order"})
    }
    
    orderID, _ := result.LastInsertId()
    
    // Insert order items
    for _, item := range items {
        _, err = tx.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
            orderID, item.ProductID, item.Quantity)
        if err != nil {
            return ctx.JSON(500, map[string]string{"error": "Failed to add items"})
        }
    }
    
    if err := tx.Commit(); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to commit transaction"})
    }
    
    return ctx.JSON(201, map[string]interface{}{
        "orderID": orderID,
    })
}
```

## Cache Operations

### Get and Set

```go
func CachedDataHandler(ctx pkg.Context) error {
    cacheKey := "user_stats"
    
    // Try to get from cache
    if cached, err := ctx.Cache().Get(cacheKey); err == nil {
        return ctx.JSON(200, map[string]interface{}{
            "data":   cached,
            "cached": true,
        })
    }
    
    // Cache miss - fetch from database
    stats := fetchUserStats(ctx)
    
    // Store in cache for 5 minutes
    ctx.Cache().Set(cacheKey, stats, 5*time.Minute)
    
    return ctx.JSON(200, map[string]interface{}{
        "data":   stats,
        "cached": false,
    })
}
```

### Cache Invalidation

```go
func UpdateUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // Update user in database
    err := updateUser(ctx, userID, userData)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Update failed"})
    }
    
    // Invalidate cache
    ctx.Cache().Delete(fmt.Sprintf("user:%s", userID))
    ctx.Cache().Delete("user_list")
    
    return ctx.JSON(200, map[string]string{"status": "updated"})
}
```

## Session Management

### Reading Session Data

```go
func ProfileHandler(ctx pkg.Context) error {
    session := ctx.Session()
    
    userID, err := session.Get("user_id")
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    preferences, _ := session.Get("preferences")
    
    return ctx.JSON(200, map[string]interface{}{
        "userID":      userID,
        "preferences": preferences,
    })
}
```

### Writing Session Data

```go
func LoginHandler(ctx pkg.Context) error {
    // Authenticate user
    user := authenticateUser(ctx)
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid credentials"})
    }
    
    session := ctx.Session()
    session.Set("user_id", user.ID)
    session.Set("login_time", time.Now())
    
    return ctx.JSON(200, map[string]string{
        "status": "logged in",
    })
}
```

### Destroying Session

```go
func LogoutHandler(ctx pkg.Context) error {
    ctx.Session().Destroy()
    return ctx.JSON(200, map[string]string{
        "status": "logged out",
    })
}
```

## Authentication and Authorization

### Checking Authentication

```go
func ProtectedHandler(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{
            "error": "Authentication required",
        })
    }
    
    user := ctx.User()
    return ctx.JSON(200, map[string]interface{}{
        "message": fmt.Sprintf("Welcome, %s!", user.Username),
        "user":    user,
    })
}
```

### Checking Authorization

```go
func AdminHandler(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Authentication required"})
    }
    
    if !ctx.IsAuthorized("admin", "access") {
        return ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
    }
    
    // Admin-only logic
    return ctx.JSON(200, map[string]string{"status": "admin access granted"})
}
```

### Accessing User Information

```go
func UserInfoHandler(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "id":       user.ID,
        "username": user.Username,
        "email":    user.Email,
        "roles":    user.Roles,
    })
}
```

## Multi-Tenancy

### Accessing Tenant Information

```go
func TenantDataHandler(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    if tenant == nil {
        return ctx.JSON(400, map[string]string{"error": "No tenant context"})
    }
    
    // Query tenant-specific data
    data := ctx.DB().Query(
        "SELECT * FROM products WHERE tenant_id = ?",
        tenant.ID,
    )
    
    return ctx.JSON(200, map[string]interface{}{
        "tenant": tenant.Name,
        "data":   data,
    })
}
```

## Internationalization

### Translating Messages

```go
func I18nHandler(ctx pkg.Context) error {
    i18n := ctx.I18n()
    
    // Simple translation
    greeting := i18n.Translate("greeting")
    
    // Translation with parameters
    welcome := i18n.TranslateWithParams("welcome_user", map[string]interface{}{
        "name": ctx.User().Username,
    })
    
    // Plural translation
    itemCount := 5
    itemsMsg := i18n.TranslatePlural("items_count", itemCount, map[string]interface{}{
        "count": itemCount,
    })
    
    return ctx.JSON(200, map[string]string{
        "greeting": greeting,
        "welcome":  welcome,
        "items":    itemsMsg,
    })
}
```

## Configuration Access

### Reading Configuration

```go
func ConfigHandler(ctx pkg.Context) error {
    config := ctx.Config()
    
    // Get configuration values
    apiKey := config.GetString("api.key")
    maxConnections := config.GetInt("database.max_connections")
    enableCache := config.GetBool("cache.enabled")
    
    return ctx.JSON(200, map[string]interface{}{
        "apiKey":         apiKey,
        "maxConnections": maxConnections,
        "cacheEnabled":   enableCache,
    })
}
```

## File Operations

### Handling File Uploads

```go
func UploadHandler(ctx pkg.Context) error {
    file, err := ctx.FormFile("upload")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "No file uploaded"})
    }
    
    // Save file
    savedPath, err := ctx.Files().Save(file, "uploads/")
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save file"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
        "path":     savedPath,
    })
}
```

### Serving Files

```go
func DownloadHandler(ctx pkg.Context) error {
    filename := ctx.Params()["filename"]
    
    filePath := fmt.Sprintf("uploads/%s", filename)
    return ctx.Files().Serve(ctx.Response(), filePath)
}
```

## Context Control

### Timeout Context

```go
func TimeoutHandler(ctx pkg.Context) error {
    // Create a context with 5-second timeout
    timeoutCtx := ctx.WithTimeout(5 * time.Second)
    
    // Check if context has deadline
    deadline, hasDeadline := timeoutCtx.Context().Deadline()
    
    // Use timeout context for operations
    result := performLongOperation(timeoutCtx)
    
    return ctx.JSON(200, map[string]interface{}{
        "result":      result,
        "hasDeadline": hasDeadline,
        "deadline":    deadline,
    })
}
```

### Cancellable Context

```go
func CancellableHandler(ctx pkg.Context) error {
    // Create a cancellable context
    cancelCtx, cancel := ctx.WithCancel()
    defer cancel()
    
    // Start background operation
    go func() {
        select {
        case <-cancelCtx.Context().Done():
            // Context was cancelled
            ctx.Logger().Info("Operation cancelled")
            return
        case <-time.After(10 * time.Second):
            // Operation completed
            ctx.Logger().Info("Operation completed")
        }
    }()
    
    return ctx.JSON(200, map[string]string{
        "status": "operation started",
    })
}
```

### Checking Cancellation

```go
func LongRunningHandler(ctx pkg.Context) error {
    for i := 0; i < 100; i++ {
        // Check for cancellation
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        default:
        }
        
        // Do work
        processItem(i)
    }
    
    return ctx.JSON(200, map[string]string{"status": "completed"})
}
```

## Logging

### Basic Logging

```go
func LoggingHandler(ctx pkg.Context) error {
    logger := ctx.Logger()
    
    logger.Info("Processing request")
    logger.Debug("Request details", "path", ctx.Request().Path)
    logger.Warn("Deprecated endpoint used")
    
    if err := someOperation(); err != nil {
        logger.Error("Operation failed", "error", err)
        return ctx.JSON(500, map[string]string{"error": "Internal error"})
    }
    
    return ctx.JSON(200, map[string]string{"status": "success"})
}
```

## Metrics Collection

### Recording Metrics

```go
func MetricsHandler(ctx pkg.Context) error {
    metrics := ctx.Metrics()
    
    // Increment counter
    metrics.IncrementCounter("api.requests")
    
    // Record timing
    start := time.Now()
    result := performOperation()
    metrics.RecordTiming("operation.duration", time.Since(start))
    
    // Record gauge
    metrics.RecordGauge("active.connections", getActiveConnections())
    
    return ctx.JSON(200, result)
}
```

## Cookie Management

### Setting Cookies

```go
func SetCookieHandler(ctx pkg.Context) error {
    cookie := &pkg.Cookie{
        Name:     "session_token",
        Value:    generateToken(),
        Path:     "/",
        MaxAge:   3600,
        Secure:   true,
        HttpOnly: true,
        SameSite: http.SameSiteStrictMode,
    }
    
    ctx.SetCookie(cookie)
    
    return ctx.JSON(200, map[string]string{
        "status": "cookie set",
    })
}
```

### Reading Cookies

```go
func GetCookieHandler(ctx pkg.Context) error {
    cookie, err := ctx.GetCookie("session_token")
    if err != nil {
        return ctx.JSON(401, map[string]string{
            "error": "No session cookie",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "token":  cookie.Value,
        "maxAge": cookie.MaxAge,
    })
}
```

## Form Handling

### Processing Form Data

```go
func FormHandler(ctx pkg.Context) error {
    // Get individual form values
    username := ctx.FormValue("username")
    email := ctx.FormValue("email")
    
    // Validate
    if username == "" || email == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Missing required fields",
        })
    }
    
    // Process form data
    user := createUser(username, email)
    
    return ctx.JSON(201, user)
}
```

## Best Practices

### 1. Always Check for Nil

```go
func SafeHandler(ctx pkg.Context) error {
    user := ctx.User()
    if user == nil {
        return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
    }
    
    tenant := ctx.Tenant()
    if tenant == nil {
        return ctx.JSON(400, map[string]string{"error": "No tenant context"})
    }
    
    // Safe to use user and tenant
    return ctx.JSON(200, map[string]interface{}{
        "user":   user,
        "tenant": tenant,
    })
}
```

### 2. Use Context for Cancellation

```go
func CancellationAwareHandler(ctx pkg.Context) error {
    results := make(chan Result)
    errors := make(chan error)
    
    go func() {
        result, err := performDatabaseQuery(ctx.Context())
        if err != nil {
            errors <- err
            return
        }
        results <- result
    }()
    
    select {
    case <-ctx.Context().Done():
        return ctx.Context().Err()
    case err := <-errors:
        return ctx.JSON(500, map[string]string{"error": err.Error()})
    case result := <-results:
        return ctx.JSON(200, result)
    }
}
```

### 3. Leverage Response Helpers

```go
// Good - uses helper
func GoodHandler(ctx pkg.Context) error {
    return ctx.JSON(200, data)
}

// Avoid - manual response writing
func AvoidHandler(ctx pkg.Context) error {
    resp := ctx.Response()
    resp.SetHeader("Content-Type", "application/json")
    resp.WriteHeader(200)
    json.NewEncoder(resp).Encode(data)
    return nil
}
```

### 4. Chain Context Operations

```go
func ChainedHandler(ctx pkg.Context) error {
    // Create timeout context
    timeoutCtx := ctx.WithTimeout(5 * time.Second)
    
    // Use it for database query
    var result Result
    err := timeoutCtx.DB().QueryRow("SELECT ...").Scan(&result)
    if err != nil {
        timeoutCtx.Logger().Error("Query failed", "error", err)
        return timeoutCtx.JSON(500, map[string]string{"error": "Query failed"})
    }
    
    return timeoutCtx.JSON(200, result)
}
```

### 5. Don't Store Context

```go
// Bad - storing context
type Handler struct {
    ctx pkg.Context  // Don't do this!
}

// Good - pass context as parameter
type Handler struct {
    db DatabaseManager
}

func (h *Handler) Handle(ctx pkg.Context) error {
    // Use context directly in handler
    return ctx.JSON(200, data)
}
```

## Testing with Context

### Creating Mock Context

```go
func TestHandler(t *testing.T) {
    // Create mock request and response
    req := &pkg.Request{
        Method: "GET",
        Path:   "/test",
        Params: map[string]string{"id": "123"},
    }
    resp := &mockResponseWriter{}
    
    // Create context
    ctx := pkg.NewContext(req, resp, context.Background())
    
    // Call handler
    err := MyHandler(ctx)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### Injecting Mock Services

```go
func TestWithMockDB(t *testing.T) {
    ctx := pkg.NewContext(req, resp, context.Background())
    
    // Inject mock database
    mockDB := &mockDatabaseManager{}
    ctx.(*pkg.contextImpl).SetDB(mockDB)
    
    // Test handler
    err := DatabaseHandler(ctx)
    
    // Verify mock was called
    assert.True(t, mockDB.QueryCalled)
}
```

## Performance Considerations

### 1. Reuse Context Methods

```go
// Good - call once
func EfficientHandler(ctx pkg.Context) error {
    db := ctx.DB()
    cache := ctx.Cache()
    
    // Use db and cache multiple times
    data1 := db.Query("...")
    data2 := db.Query("...")
    cache.Set("key", data1, time.Minute)
}

// Less efficient - repeated calls
func InefficientHandler(ctx pkg.Context) error {
    data1 := ctx.DB().Query("...")
    data2 := ctx.DB().Query("...")
    ctx.Cache().Set("key", data1, time.Minute)
}
```

### 2. Lazy Loading

The Context implementation uses lazy loading for headers and query parameters, so they're only parsed when accessed.

### 3. Context Copying

When using `WithTimeout()` or `WithCancel()`, a shallow copy of the context is created. This is efficient but means you should use the returned context for subsequent operations.

## Common Patterns

### Middleware Pattern

```go
func AuthMiddleware(next pkg.HandlerFunc) pkg.HandlerFunc {
    return func(ctx pkg.Context) error {
        token := ctx.GetHeader("Authorization")
        if token == "" {
            return ctx.JSON(401, map[string]string{"error": "No token"})
        }
        
        user := validateToken(token)
        if user == nil {
            return ctx.JSON(401, map[string]string{"error": "Invalid token"})
        }
        
        // Set user in context
        ctx.(*pkg.contextImpl).SetUser(user)
        
        return next(ctx)
    }
}
```

### Repository Pattern

```go
type UserRepository struct{}

func (r *UserRepository) GetByID(ctx pkg.Context, id string) (*User, error) {
    var user User
    err := ctx.DB().QueryRow(
        "SELECT id, name, email FROM users WHERE id = ?",
        id,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    return &user, err
}

func UserHandler(ctx pkg.Context) error {
    repo := &UserRepository{}
    user, err := repo.GetByID(ctx, ctx.Params()["id"])
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return ctx.JSON(200, user)
}
```

## Summary

The Context interface is your gateway to all framework features. It provides:

- Request/response access
- Database and cache operations
- Session management
- Authentication and authorization
- Configuration and i18n
- File operations
- Logging and metrics
- Context control (timeouts, cancellation)

By centralizing all framework services through Context, the framework eliminates global state, improves testability, and provides a clean, consistent API for building web applications.
