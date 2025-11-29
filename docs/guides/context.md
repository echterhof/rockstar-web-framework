# Context

## Overview

The Context is the heart of request handling in the Rockstar Web Framework. It provides a unified interface to access everything you need when processing a request: request data, response methods, framework services, and utilities. Every handler function receives a Context, making it the primary way you interact with the framework.

The Context eliminates global state and provides dependency injection, making your code testable and maintainable. It also enables request-scoped data sharing between middleware and handlers.

## Quick Start

Here's a basic example using Context:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
    }
    
    app, _ := pkg.New(config)
    router := app.Router()
    
    // Handler receives Context
    router.GET("/users/:id", func(ctx pkg.Context) error {
        // Access route parameters
        userID := ctx.Param("id")
        
        // Access query parameters
        query := ctx.Query()
        sort := query["sort"]
        
        // Access database
        db := ctx.DB()
        
        // Send JSON response
        return ctx.JSON(200, map[string]interface{}{
            "user_id": userID,
            "sort":    sort,
        })
    })
    
    app.Listen(":8080")
}
```

## Request Data Access

### Route Parameters

Access path parameters defined in routes:

```go
// Route: /users/:id/posts/:postID
router.GET("/users/:id/posts/:postID", func(ctx pkg.Context) error {
    // Get all parameters as map
    params := ctx.Params()
    userID := params["id"]
    postID := params["postID"]
    
    // Or get individual parameter
    userID := ctx.Param("id")
    postID := ctx.Param("postID")
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
        "post_id": postID,
    })
})
```

### Query Parameters

Access URL query string parameters:

```go
// Request: GET /search?q=golang&page=2&limit=10
router.GET("/search", func(ctx pkg.Context) error {
    // Get all query parameters as map
    query := ctx.Query()
    searchTerm := query["q"]
    page := query["page"]
    limit := query["limit"]
    
    // Handle missing parameters
    if searchTerm == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Missing search term",
        })
    }
    
    // Use default values
    if page == "" {
        page = "1"
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "query": searchTerm,
        "page":  page,
        "limit": limit,
    })
})
```

### Request Headers

Access HTTP headers:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Get all headers as map
    headers := ctx.Headers()
    
    // Get specific header
    contentType := ctx.GetHeader("Content-Type")
    authorization := ctx.GetHeader("Authorization")
    userAgent := ctx.GetHeader("User-Agent")
    
    // Headers are case-insensitive
    auth1 := ctx.GetHeader("Authorization")
    auth2 := ctx.GetHeader("authorization")  // Same result
    
    return ctx.JSON(200, map[string]interface{}{
        "content_type":  contentType,
        "authorization": authorization,
        "user_agent":    userAgent,
        "total_headers": len(headers),
    })
})
```

### Request Body

Access the raw request body:

```go
router.POST("/api/data", func(ctx pkg.Context) error {
    // Get raw body as bytes
    body := ctx.Body()
    
    // Process body
    bodyLength := len(body)
    
    return ctx.JSON(200, map[string]interface{}{
        "body_length": bodyLength,
    })
})
```

### Request Object

Access the full request object:

```go
router.GET("/api/info", func(ctx pkg.Context) error {
    req := ctx.Request()
    
    return ctx.JSON(200, map[string]interface{}{
        "method":      req.Method,
        "url":         req.URL.String(),
        "remote_addr": req.RemoteAddr,
        "host":        req.Host,
    })
})
```

## Data Binding

### JSON Binding

Parse JSON request bodies:

```go
type CreateUserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

router.POST("/users", func(ctx pkg.Context) error {
    var req CreateUserRequest
    
    // Bind JSON body to struct
    if err := ctx.BindJSON(&req); err != nil {
        return ctx.JSON(400, map[string]string{
            "error": "Invalid JSON",
        })
    }
    
    // Validate
    if req.Name == "" || req.Email == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Missing required fields",
        })
    }
    
    // Process request
    return ctx.JSON(201, map[string]interface{}{
        "message": "User created",
        "name":    req.Name,
        "email":   req.Email,
    })
})
```

### Form Data

Access form data from POST requests:

```go
router.POST("/login", func(ctx pkg.Context) error {
    // Get form values
    username := ctx.FormValue("username")
    password := ctx.FormValue("password")
    
    if username == "" || password == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Missing credentials",
        })
    }
    
    // Authenticate user
    // ...
    
    return ctx.JSON(200, map[string]string{
        "message": "Login successful",
    })
})
```

### File Uploads

Handle file uploads:

```go
router.POST("/upload", func(ctx pkg.Context) error {
    // Get uploaded file
    file, err := ctx.FormFile("file")
    if err != nil {
        return ctx.JSON(400, map[string]string{
            "error": "No file uploaded",
        })
    }
    
    // Access file information
    filename := file.Filename
    size := file.Size
    contentType := file.Header.Get("Content-Type")
    
    // Save file using file manager
    files := ctx.Files()
    destPath := "/uploads/" + filename
    if err := files.SaveUploadedFile(ctx, filename, destPath); err != nil {
        return ctx.JSON(500, map[string]string{
            "error": "Failed to save file",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message":      "File uploaded successfully",
        "filename":     filename,
        "size":         size,
        "content_type": contentType,
        "path":         destPath,
    })
})
```

## Response Methods

### JSON Response

Send JSON responses:

```go
router.GET("/api/users", func(ctx pkg.Context) error {
    users := []map[string]interface{}{
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"},
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "users": users,
        "total": len(users),
    })
})
```

### XML Response

Send XML responses:

```go
type User struct {
    ID   int    `xml:"id"`
    Name string `xml:"name"`
}

router.GET("/api/users.xml", func(ctx pkg.Context) error {
    user := User{ID: 1, Name: "Alice"}
    return ctx.XML(200, user)
})
```

### HTML Response

Send HTML responses:

```go
router.GET("/", func(ctx pkg.Context) error {
    html := `
<!DOCTYPE html>
<html>
<head><title>Welcome</title></head>
<body>
    <h1>Welcome to Rockstar!</h1>
</body>
</html>
`
    return ctx.HTML(200, "inline", html)
})
```

### String Response

Send plain text responses:

```go
router.GET("/health", func(ctx pkg.Context) error {
    return ctx.String(200, "OK")
})
```

### Redirect

Send redirect responses:

```go
router.GET("/old-path", func(ctx pkg.Context) error {
    return ctx.Redirect(301, "/new-path")
})

router.GET("/login-required", func(ctx pkg.Context) error {
    return ctx.Redirect(302, "/login")
})
```

### No Content

Send responses without body:

```go
router.DELETE("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Delete user
    // ...
    
    return ctx.NoContent(204)
})
```

## Response Headers and Cookies

### Setting Headers

Set response headers:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Set custom headers
    ctx.SetHeader("X-Custom-Header", "CustomValue")
    ctx.SetHeader("X-Request-ID", "req-12345")
    ctx.SetHeader("Cache-Control", "no-cache")
    
    return ctx.JSON(200, map[string]string{
        "message": "Headers set",
    })
})
```

### Setting Cookies

Set cookies in responses:

```go
router.POST("/login", func(ctx pkg.Context) error {
    // Authenticate user
    // ...
    
    // Set session cookie
    cookie := &pkg.Cookie{
        Name:     "session_id",
        Value:    "abc123xyz",
        Path:     "/",
        MaxAge:   3600,        // 1 hour
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    
    ctx.SetCookie(cookie)
    
    return ctx.JSON(200, map[string]string{
        "message": "Login successful",
    })
})
```

### Getting Cookies

Read cookies from requests:

```go
router.GET("/profile", func(ctx pkg.Context) error {
    // Get session cookie
    cookie, err := ctx.GetCookie("session_id")
    if err != nil {
        return ctx.JSON(401, map[string]string{
            "error": "Not authenticated",
        })
    }
    
    sessionID := cookie.Value
    
    // Validate session
    // ...
    
    return ctx.JSON(200, map[string]interface{}{
        "session_id": sessionID,
    })
})
```

## Framework Services

### Database Access

Access the database manager:

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Get database manager
    db := ctx.DB()
    
    // Query database
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    if err != nil {
        return ctx.JSON(404, map[string]string{
            "error": "User not found",
        })
    }
    
    return ctx.JSON(200, user)
})
```

### Cache Access

Access the cache manager:

```go
router.GET("/products/:id", func(ctx pkg.Context) error {
    productID := ctx.Param("id")
    cacheKey := "product:" + productID
    
    // Get cache manager
    cache := ctx.Cache()
    
    // Try to get from cache
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Fetch from database
    db := ctx.DB()
    var product Product
    err := db.QueryRow("SELECT * FROM products WHERE id = ?", productID).Scan(&product)
    if err != nil {
        return ctx.JSON(404, map[string]string{
            "error": "Product not found",
        })
    }
    
    // Store in cache
    cache.Set(cacheKey, product, 5*time.Minute)
    
    return ctx.JSON(200, product)
})
```

### Session Management

Access the session manager:

```go
router.POST("/cart/add", func(ctx pkg.Context) error {
    // Get session manager
    session := ctx.Session()
    
    // Get or create session
    sess, err := session.GetSessionFromCookie(ctx)
    if err != nil {
        sess, err = session.Create(ctx)
        if err != nil {
            return ctx.JSON(500, map[string]string{
                "error": "Failed to create session",
            })
        }
    }
    
    // Store data in session
    productID := ctx.FormValue("product_id")
    session.Set(sess.ID, "cart_item", productID)
    
    return ctx.JSON(200, map[string]string{
        "message": "Item added to cart",
    })
})
```

### Configuration Access

Access configuration:

```go
router.GET("/api/config", func(ctx pkg.Context) error {
    // Get config manager
    config := ctx.Config()
    
    // Read configuration values
    appName := config.GetString("app_name")
    maxUploadSize := config.GetInt("max_upload_size")
    debugMode := config.GetBool("debug")
    
    return ctx.JSON(200, map[string]interface{}{
        "app_name":        appName,
        "max_upload_size": maxUploadSize,
        "debug_mode":      debugMode,
    })
})
```

### Internationalization

Access i18n for translations:

```go
router.GET("/welcome", func(ctx pkg.Context) error {
    // Get i18n manager
    i18n := ctx.I18n()
    
    // Get user's language preference
    lang := ctx.GetHeader("Accept-Language")
    if lang != "" {
        i18n.SetLanguage(lang)
    }
    
    // Translate messages
    welcome := i18n.Translate("welcome.message")
    greeting := i18n.Translate("welcome.greeting", "John")
    
    return ctx.JSON(200, map[string]interface{}{
        "welcome":  welcome,
        "greeting": greeting,
    })
})
```

### Logging

Access the logger:

```go
router.POST("/api/data", func(ctx pkg.Context) error {
    // Get logger
    logger := ctx.Logger()
    
    // Log messages
    logger.Info("Processing data request", "endpoint", "/api/data")
    
    // Process request
    // ...
    
    logger.Debug("Data processed successfully", "count", 10)
    
    return ctx.JSON(200, map[string]string{
        "message": "Data processed",
    })
})
```

### Metrics

Access metrics collector:

```go
router.GET("/api/expensive", func(ctx pkg.Context) error {
    start := time.Now()
    
    // Get metrics collector
    metrics := ctx.Metrics()
    
    // Process expensive operation
    // ...
    
    // Record timing
    duration := time.Since(start)
    metrics.RecordTiming("expensive_operation", duration, map[string]string{
        "endpoint": "/api/expensive",
    })
    
    return ctx.JSON(200, map[string]string{
        "message": "Operation completed",
    })
})
```

## Context Control

### Timeout Context

Create a context with timeout:

```go
router.GET("/api/slow", func(ctx pkg.Context) error {
    // Create context with 5-second timeout
    timeoutCtx := ctx.WithTimeout(5 * time.Second)
    
    // Use timeout context for operations
    done := make(chan error, 1)
    go func() {
        // Long-running operation
        time.Sleep(3 * time.Second)
        done <- nil
    }()
    
    // Wait for completion or timeout
    select {
    case err := <-done:
        if err != nil {
            return ctx.JSON(500, map[string]string{
                "error": "Operation failed",
            })
        }
        return ctx.JSON(200, map[string]string{
            "message": "Operation completed",
        })
    case <-timeoutCtx.Context().Done():
        return ctx.JSON(504, map[string]string{
            "error": "Operation timed out",
        })
    }
})
```

### Cancellable Context

Create a cancellable context:

```go
router.GET("/api/cancellable", func(ctx pkg.Context) error {
    // Create cancellable context
    cancelCtx, cancel := ctx.WithCancel()
    defer cancel()
    
    // Start operation
    done := make(chan error, 1)
    go func() {
        // Check for cancellation
        select {
        case <-cancelCtx.Context().Done():
            done <- cancelCtx.Context().Err()
            return
        default:
            // Continue processing
        }
        
        // Long operation
        time.Sleep(2 * time.Second)
        done <- nil
    }()
    
    // Wait for completion
    err := <-done
    if err != nil {
        return ctx.JSON(500, map[string]string{
            "error": "Operation cancelled",
        })
    }
    
    return ctx.JSON(200, map[string]string{
        "message": "Operation completed",
    })
})
```

### Standard Context

Access the underlying Go context:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Get standard context
    stdCtx := ctx.Context()
    
    // Check if request was cancelled
    select {
    case <-stdCtx.Done():
        return stdCtx.Err()
    default:
        // Continue processing
    }
    
    // Use with context-aware libraries
    // result, err := someLibrary.DoWork(stdCtx, data)
    
    return ctx.JSON(200, map[string]string{
        "message": "Data processed",
    })
})
```

## Request-Scoped Data

### Storing Data

Store custom data in context for sharing between middleware and handlers:

```go
// Middleware sets user data
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    
    // Validate token and get user
    user, err := validateToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{
            "error": "Unauthorized",
        })
    }
    
    // Store user in context
    ctx.Set("user", user)
    ctx.Set("user_id", user.ID)
    
    return next(ctx)
}

// Handler retrieves user data
router.GET("/profile", func(ctx pkg.Context) error {
    // Get user from context
    user, ok := ctx.Get("user")
    if !ok {
        return ctx.JSON(401, map[string]string{
            "error": "Not authenticated",
        })
    }
    
    return ctx.JSON(200, user)
}, authMiddleware)
```

### Retrieving Data

Retrieve data stored in context:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    // Get with type assertion
    if userID, ok := ctx.Get("user_id"); ok {
        id := userID.(string)
        // Use user ID
    }
    
    // Check if key exists
    if _, exists := ctx.Get("tenant_id"); exists {
        // Tenant-specific logic
    }
    
    return ctx.JSON(200, map[string]string{
        "message": "Data retrieved",
    })
})
```

## Authentication and Authorization

### Check Authentication

Check if user is authenticated:

```go
router.GET("/dashboard", func(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return ctx.Redirect(302, "/login")
    }
    
    user := ctx.User()
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Welcome to dashboard",
        "user":    user.Name,
    })
})
```

### Check Authorization

Check if user is authorized:

```go
router.DELETE("/posts/:id", func(ctx pkg.Context) error {
    postID := ctx.Param("id")
    
    // Check if user can delete posts
    if !ctx.IsAuthorized("posts", "delete") {
        return ctx.JSON(403, map[string]string{
            "error": "Insufficient permissions",
        })
    }
    
    // Delete post
    // ...
    
    return ctx.JSON(200, map[string]string{
        "message": "Post deleted",
    })
})
```

### Access User

Access authenticated user:

```go
router.GET("/profile", func(ctx pkg.Context) error {
    user := ctx.User()
    
    if user == nil {
        return ctx.JSON(401, map[string]string{
            "error": "Not authenticated",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
    })
})
```

### Access Tenant

Access tenant in multi-tenant applications:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    
    if tenant == nil {
        return ctx.JSON(400, map[string]string{
            "error": "No tenant specified",
        })
    }
    
    // Use tenant-specific database
    db := ctx.DB()
    // Query with tenant filter
    
    return ctx.JSON(200, map[string]interface{}{
        "tenant": tenant.Name,
    })
})
```

## Best Practices

1. **Always Use Context:** Never use global variables. Access everything through the context.

2. **Don't Store Context:** Don't store context in structs or pass it between goroutines. Create new contexts when needed.

3. **Use Type Assertions Safely:** When retrieving data with `Get()`, always check the type assertion.

4. **Set Data Early:** Store shared data in context as early as possible (in middleware) so handlers can access it.

5. **Use Descriptive Keys:** Use clear, descriptive keys when storing data in context.

6. **Handle Missing Data:** Always check if data exists before using it.

7. **Return Errors:** Return errors from handlers instead of panicking.

8. **Use Appropriate Response Methods:** Use `JSON()` for APIs, `HTML()` for web pages, etc.

9. **Set Headers Before Body:** Set all headers before writing the response body.

10. **Check Authentication Early:** Check authentication at the beginning of handlers that require it.

## Troubleshooting

### Context Data Not Available

**Problem:** Data stored in context is not available in handlers.

**Solutions:**
- Ensure middleware runs before the handler
- Check that `ctx.Set()` is called before `ctx.Get()`
- Verify the key name matches exactly
- Ensure middleware calls `next(ctx)`

### Type Assertion Panic

**Problem:** Panic when retrieving data from context.

**Solutions:**
- Always check the boolean return value from `Get()`
- Use type assertions safely with comma-ok idiom
- Provide default values for missing data

```go
// Wrong - can panic
userID := ctx.Get("user_id").(string)

// Right - safe
if userID, ok := ctx.Get("user_id"); ok {
    id := userID.(string)
    // Use id
}
```

### Response Already Written

**Problem:** Error about response already being written.

**Solutions:**
- Only call one response method per request
- Don't write response in middleware and handler
- Check that middleware doesn't write response before calling `next()`

### Headers Not Set

**Problem:** Response headers are not being set.

**Solutions:**
- Set headers before writing response body
- Ensure `SetHeader()` is called before `JSON()`, `String()`, etc.
- Check that middleware isn't overwriting headers

## See Also

- [Routing Guide](routing.md) - Route definition and parameters
- [Middleware Guide](middleware.md) - Middleware and data sharing
- [Database Guide](database.md) - Database operations
- [Security Guide](security.md) - Authentication and authorization
- [API Reference: Context](../api/context.md) - Complete Context API
- [Getting Started Example](../examples/getting-started.md) - Complete working example
