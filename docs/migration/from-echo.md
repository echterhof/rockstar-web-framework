# Migrating from Echo to Rockstar

This guide helps you migrate your application from the Echo web framework to Rockstar. Both frameworks share similar design philosophies around performance and simplicity, but Rockstar adds enterprise features like multi-tenancy, plugin system, and comprehensive security.

## Overview

Echo is a high-performance, minimalist web framework for Go. Rockstar builds on similar principles but provides a more comprehensive feature set including multi-protocol support, built-in database abstraction, and enterprise-grade security out of the box.

## Concept Mapping

| Echo Concept | Rockstar Equivalent | Notes |
|-------------|---------------------|-------|
| `echo.Echo` | `pkg.Framework` | Main application instance |
| `echo.Context` | `pkg.Context` | Request/response context |
| `echo.HandlerFunc` | `pkg.HandlerFunc` | Handler function signature |
| `echo.Group` | Route groups via `Router()` | Similar grouping functionality |
| `echo.Map` | `map[string]interface{}` | JSON response helper |
| Middleware | Middleware via `Use()` | Similar middleware pattern |
| `c.Param()` | `c.Param()` | Path parameter extraction |
| `c.QueryParam()` | `c.QueryParam()` | Query parameter extraction |
| `c.Bind()` | `c.Bind()` | Request body binding |
| `c.JSON()` | `c.JSON()` | JSON response |
| `c.String()` | `c.String()` | String response |
| `c.HTML()` | `c.HTML()` | HTML response |
| `c.Redirect()` | `c.Redirect()` | HTTP redirect |
| `c.File()` | `c.File()` | File serving |
| `c.Set()` / `c.Get()` | `c.Set()` / `c.Get()` | Context value storage |
| `echo.Validator` | Built-in validation | More comprehensive |
| Manual session handling | `SessionManager` | Built-in session support |
| Manual database setup | `DatabaseManager` | Built-in database abstraction |

## Key Differences

### 1. Initialization

**Echo:**
```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())
```

**Rockstar:**
```go
import "github.com/echterhof/rockstar-web-framework/pkg"

config := &pkg.Config{
    ServerAddress: ":8080",
    Environment:   "development",
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
defer app.Shutdown()
```

### 2. Routing

**Echo:**
```go
e.GET("/users/:id", getUser)
e.POST("/users", createUser)
e.PUT("/users/:id", updateUser)
e.DELETE("/users/:id", deleteUser)
```

**Rockstar:**
```go
router := app.Router()

router.GET("/users/:id", getUser)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

**Note:** The routing syntax is nearly identical between Echo and Rockstar.

### 3. Route Groups

**Echo:**
```go
api := e.Group("/api")
api.Use(middleware.Logger())

v1 := api.Group("/v1")
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

admin := e.Group("/admin")
admin.Use(AuthMiddleware)
admin.GET("/dashboard", dashboard)
```

**Rockstar:**
```go
router := app.Router()

// API routes with middleware
router.Use("/api/*", LoggerMiddleware())

// V1 routes
router.GET("/api/v1/users", listUsers)
router.POST("/api/v1/users", createUser)

// Admin routes with auth
router.Use("/admin/*", AuthMiddleware())
router.GET("/admin/dashboard", dashboard)
```

### 4. Middleware

**Echo:**
```go
func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Before request
        start := time.Now()
        
        err := next(c)
        
        // After request
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
        
        return err
    }
}

e.Use(LoggerMiddleware)
```

**Rockstar:**
```go
func LoggerMiddleware() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            // Before request
            start := time.Now()
            
            err := next(c)
            
            // After request
            duration := time.Since(start)
            log.Printf("Request took %v", duration)
            
            return err
        }
    }
}

app.Use(LoggerMiddleware())
```

**Key Difference:** Rockstar middleware returns a function that creates the middleware, allowing for configuration.

### 5. Request Handling

**Echo:**
```go
func getUser(c echo.Context) error {
    id := c.Param("id")
    name := c.QueryParam("name")
    
    var user User
    if err := c.Bind(&user); err != nil {
        return c.JSON(400, map[string]interface{}{"error": err.Error()})
    }
    
    return c.JSON(200, user)
}
```

**Rockstar:**
```go
func getUser(c pkg.Context) error {
    id := c.Param("id")
    name := c.QueryParam("name")
    
    var user User
    if err := c.Bind(&user); err != nil {
        return c.JSON(400, map[string]interface{}{"error": err.Error()})
    }
    
    return c.JSON(200, user)
}
```

**Note:** Handler signatures are identical - both return `error`.

### 6. Error Handling

**Echo:**
```go
func handler(c echo.Context) error {
    if err := doSomething(); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }
    return c.JSON(200, map[string]interface{}{"status": "ok"})
}

// Custom error handler
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := http.StatusInternalServerError
    message := err.Error()
    
    if he, ok := err.(*echo.HTTPError); ok {
        code = he.Code
        message = he.Message.(string)
    }
    
    c.JSON(code, map[string]interface{}{"error": message})
}
```

**Rockstar:**
```go
func handler(c pkg.Context) error {
    if err := doSomething(); err != nil {
        return err // Framework handles error response
    }
    return c.JSON(200, map[string]interface{}{"status": "ok"})
}

// Custom error handling via middleware
func ErrorHandler() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            err := next(c)
            if err != nil {
                // Custom error handling logic
                return c.JSON(500, map[string]interface{}{"error": err.Error()})
            }
            return nil
        }
    }
}
```

### 7. Static Files

**Echo:**
```go
e.Static("/static", "public")
e.File("/favicon.ico", "favicon.ico")
```

**Rockstar:**
```go
router := app.Router()
router.Static("/static", "public")
router.GET("/favicon.ico", func(c pkg.Context) error {
    return c.File("favicon.ico")
})
```

### 8. Template Rendering

**Echo:**
```go
import "html/template"

t := &Template{
    templates: template.Must(template.ParseGlob("templates/*.html")),
}
e.Renderer = t

func handler(c echo.Context) error {
    return c.Render(200, "index.html", map[string]interface{}{
        "title": "Home",
    })
}
```

**Rockstar:**
```go
// Configure templates in config
config := &pkg.Config{
    TemplateDir: "templates",
}

func handler(c pkg.Context) error {
    return c.Render(200, "index.html", map[string]interface{}{
        "title": "Home",
    })
}
```

### 9. Validation

**Echo:**
```go
import "github.com/go-playground/validator/v10"

type CustomValidator struct {
    validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

e.Validator = &CustomValidator{validator: validator.New()}

type User struct {
    Email string `json:"email" validate:"required,email"`
}

func handler(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return err
    }
    if err := c.Validate(&user); err != nil {
        return err
    }
    return c.JSON(200, user)
}
```

**Rockstar:**
```go
type User struct {
    Email string `json:"email" validate:"required,email"`
}

func handler(c pkg.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return err
    }
    // Validation happens automatically during Bind
    return c.JSON(200, user)
}
```

## Additional Features in Rockstar

### Database Integration

Rockstar includes built-in database support:

```go
config := &pkg.Config{
    DatabaseDriver: "postgres",
    DatabaseDSN:    "postgres://user:pass@localhost/dbname?sslmode=disable",
}

app, _ := pkg.New(config)

// Access database through context
func handler(c pkg.Context) error {
    db := c.Database()
    var users []User
    err := db.Query("SELECT * FROM users", &users)
    if err != nil {
        return err
    }
    return c.JSON(200, users)
}
```

### Session Management

Built-in session support:

```go
func handler(c pkg.Context) error {
    session := c.Session()
    
    // Set session value
    session.Set("user_id", 123)
    
    // Get session value
    userID := session.Get("user_id")
    
    return c.JSON(200, map[string]interface{}{"user_id": userID})
}
```

### Caching

Built-in caching:

```go
func handler(c pkg.Context) error {
    cache := c.Cache()
    
    // Try to get from cache
    var data interface{}
    if err := cache.Get("key", &data); err == nil {
        return c.JSON(200, data)
    }
    
    // Fetch and cache
    data = fetchData()
    cache.Set("key", data, 5*time.Minute)
    
    return c.JSON(200, data)
}
```

### Security Features

Built-in authentication and authorization:

```go
func handler(c pkg.Context) error {
    security := c.Security()
    
    // Validate JWT token
    token := c.Header("Authorization")
    claims, err := security.ValidateJWT(token)
    if err != nil {
        return c.JSON(401, map[string]interface{}{"error": "Unauthorized"})
    }
    
    // Check permissions
    if !security.HasPermission(claims.UserID, "admin") {
        return c.JSON(403, map[string]interface{}{"error": "Forbidden"})
    }
    
    return c.JSON(200, map[string]interface{}{"status": "ok"})
}
```

### Multi-Tenancy

Built-in multi-tenant support:

```go
config := &pkg.Config{
    MultiTenancy: true,
}

func handler(c pkg.Context) error {
    tenant := c.Tenant()
    
    // Tenant-specific database
    db := c.Database()
    // Automatically scoped to current tenant
    
    return c.JSON(200, map[string]interface{}{"tenant": tenant.ID})
}
```

### Internationalization

Built-in i18n support:

```go
func handler(c pkg.Context) error {
    i18n := c.I18n()
    
    message := i18n.Translate("welcome.message")
    
    return c.JSON(200, map[string]interface{}{"message": message})
}
```

### Multi-Protocol Support

Support for HTTP/2, QUIC, and WebSocket:

```go
config := &pkg.Config{
    ServerAddress: ":8080",
    EnableHTTP2:   true,
    EnableQUIC:    true,
    TLSCertFile:   "cert.pem",
    TLSKeyFile:    "key.pem",
}

// WebSocket handler
router.GET("/ws", func(c pkg.Context) error {
    ws := c.WebSocket()
    defer ws.Close()
    
    for {
        var msg string
        if err := ws.ReadJSON(&msg); err != nil {
            break
        }
        ws.WriteJSON(map[string]interface{}{"echo": msg})
    }
    return nil
})
```

## Migration Checklist

### Phase 1: Setup

- [ ] Install Rockstar framework: `go get github.com/echterhof/rockstar-web-framework/pkg`
- [ ] Create configuration struct with your settings
- [ ] Initialize framework with `pkg.New(config)`
- [ ] Update imports from `echo` to `pkg`

### Phase 2: Update Application Structure

- [ ] Replace `echo.New()` with `pkg.New(config)`
- [ ] Update `e.Start()` to `app.Start()`
- [ ] Add `defer app.Shutdown()` for cleanup
- [ ] Update router access from `e` to `app.Router()`

### Phase 3: Update Handlers

- [ ] Change handler signature from `echo.Context` to `pkg.Context`
- [ ] Update `echo.Map` to `map[string]interface{}`
- [ ] Verify all handlers return `error`
- [ ] Update context method calls if needed

### Phase 4: Update Routing

- [ ] Update route registration to use `app.Router()`
- [ ] Convert route groups to use path prefixes or middleware
- [ ] Update static file serving

### Phase 5: Update Middleware

- [ ] Convert middleware to return `pkg.MiddlewareFunc`
- [ ] Update middleware registration from `e.Use()` to `app.Use()`
- [ ] Test middleware chain execution

### Phase 6: Update Error Handling

- [ ] Remove `echo.HTTPError` usage
- [ ] Update custom error handler if needed
- [ ] Test error responses

### Phase 7: Leverage New Features

- [ ] Replace manual database code with `DatabaseManager`
- [ ] Replace manual session handling with `SessionManager`
- [ ] Add caching where appropriate using `CacheManager`
- [ ] Implement authentication using `SecurityManager`
- [ ] Add i18n support if needed

### Phase 8: Testing

- [ ] Update tests to use Rockstar test helpers
- [ ] Test all routes and handlers
- [ ] Test middleware chain
- [ ] Test error handling
- [ ] Load test for performance comparison

### Phase 9: Deployment

- [ ] Update deployment configuration
- [ ] Configure production settings
- [ ] Set up monitoring and metrics
- [ ] Deploy and monitor

## Common Pitfalls

### 1. Forgetting Shutdown

**Wrong:**
```go
app, _ := pkg.New(config)
app.Start()
```

**Correct:**
```go
app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
defer app.Shutdown()
app.Start()
```

### 2. Not Handling Initialization Errors

**Wrong:**
```go
app, _ := pkg.New(config) // Ignoring error
```

**Correct:**
```go
app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
```

### 3. Incorrect Middleware Signature

**Wrong:**
```go
func middleware(next pkg.HandlerFunc) pkg.HandlerFunc {
    // Missing wrapper function
    return func(c pkg.Context) error {
        return next(c)
    }
}
```

**Correct:**
```go
func middleware() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            return next(c)
        }
    }
}
```

### 4. Using Echo-Specific Features

Remove dependencies on Echo-specific packages:
- `echo.HTTPError` → Standard error handling
- `echo.Validator` → Built-in validation
- `echo.Binder` → Built-in binding

## Performance Considerations

Both Echo and Rockstar are designed for high performance:

- Similar routing performance
- Efficient memory management
- Connection pooling for databases
- Built-in caching reduces external dependencies
- Request-level context prevents memory leaks

Benchmark your application after migration to ensure performance meets your requirements.

## Getting Help

- **Documentation**: See the [complete documentation](../README.md)
- **Examples**: Check the [examples directory](../../examples/)
- **Issues**: Report issues on GitHub
- **Community**: Join our community channels

## Next Steps

After completing the migration:

1. Review the [Configuration Guide](../guides/configuration.md) for advanced settings
2. Explore [Security Features](../guides/security.md) to enhance your application
3. Consider adding [Multi-Tenancy](../guides/multi-tenancy.md) if applicable
4. Set up [Monitoring](../guides/monitoring.md) for production
5. Review [Performance Tuning](../guides/performance.md) guide

## Example: Complete Migration

### Before (Echo)

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "net/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    e := echo.New()
    
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    e.GET("/users/:id", getUser)
    e.POST("/users", createUser)
    
    e.Start(":8080")
}

func getUser(c echo.Context) error {
    id := c.Param("id")
    user := User{ID: 1, Name: "John"}
    return c.JSON(http.StatusOK, user)
}

func createUser(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, user)
}
```

### After (Rockstar)

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
    "net/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    config := &pkg.Config{
        ServerAddress: ":8080",
        Environment:   "development",
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer app.Shutdown()
    
    router := app.Router()
    router.GET("/users/:id", getUser)
    router.POST("/users", createUser)
    
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}

func getUser(c pkg.Context) error {
    id := c.Param("id")
    user := User{ID: 1, Name: "John"}
    return c.JSON(http.StatusOK, user)
}

func createUser(c pkg.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, user)
}
```

The migration from Echo is very straightforward due to similar design philosophies. The main changes are in initialization and accessing the router, while handler code remains largely unchanged.
