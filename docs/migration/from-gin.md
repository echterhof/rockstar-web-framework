# Migrating from Gin to Rockstar

This guide helps you migrate your application from the Gin web framework to Rockstar. While both frameworks share similar concepts, Rockstar provides additional enterprise features like multi-tenancy, plugin system, and built-in security.

## Overview

Gin is a lightweight, high-performance HTTP web framework. Rockstar builds on similar principles but adds enterprise-grade features including multi-protocol support, comprehensive security, and extensibility through plugins.

## Concept Mapping

| Gin Concept | Rockstar Equivalent | Notes |
|-------------|---------------------|-------|
| `gin.Engine` | `pkg.Framework` | Main application instance |
| `gin.Context` | `pkg.Context` | Request/response context |
| `gin.HandlerFunc` | `pkg.HandlerFunc` | Handler function signature |
| `gin.RouterGroup` | Route groups via `Router()` | Similar grouping functionality |
| `gin.H` | `map[string]interface{}` | JSON response helper |
| Middleware | Middleware via `Use()` | Similar middleware pattern |
| `c.Param()` | `c.Param()` | Path parameter extraction |
| `c.Query()` | `c.QueryParam()` | Query parameter extraction |
| `c.Bind()` | `c.Bind()` | Request body binding |
| `c.JSON()` | `c.JSON()` | JSON response |
| `c.String()` | `c.String()` | String response |
| `c.HTML()` | `c.HTML()` | HTML response |
| `c.Redirect()` | `c.Redirect()` | HTTP redirect |
| `c.File()` | `c.File()` | File serving |
| `c.Set()` / `c.Get()` | `c.Set()` / `c.Get()` | Context value storage |
| `gin.BasicAuth()` | `SecurityManager` | Built-in auth system |
| Custom validators | Built-in validation | More comprehensive |
| Manual session handling | `SessionManager` | Built-in session support |
| Manual database setup | `DatabaseManager` | Built-in database abstraction |

## Key Differences

### 1. Initialization

**Gin:**
```go
import "github.com/gin-gonic/gin"

router := gin.Default()
// or
router := gin.New()
router.Use(gin.Logger(), gin.Recovery())
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

**Gin:**
```go
router.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

**Rockstar:**
```go
router := app.Router()

router.GET("/users/:id", func(c pkg.Context) error {
    id := c.Param("id")
    return c.JSON(200, map[string]interface{}{"id": id})
})

router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

**Key Difference:** Rockstar handlers return `error` for consistent error handling.

### 3. Route Groups

**Gin:**
```go
v1 := router.Group("/api/v1")
{
    v1.GET("/users", listUsers)
    v1.POST("/users", createUser)
}

admin := router.Group("/admin")
admin.Use(AuthRequired())
{
    admin.GET("/dashboard", dashboard)
}
```

**Rockstar:**
```go
router := app.Router()

// API v1 routes
router.GET("/api/v1/users", listUsers)
router.POST("/api/v1/users", createUser)

// Admin routes with middleware
router.Use("/admin/*", AuthRequired())
router.GET("/admin/dashboard", dashboard)
```

### 4. Middleware

**Gin:**
```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before request
        start := time.Now()
        
        c.Next()
        
        // After request
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    }
}

router.Use(Logger())
```

**Rockstar:**
```go
func Logger() pkg.MiddlewareFunc {
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

app.Use(Logger())
```

### 5. Request Handling

**Gin:**
```go
func getUser(c *gin.Context) {
    id := c.Param("id")
    name := c.Query("name")
    
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, user)
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

### 6. Error Handling

**Gin:**
```go
func handler(c *gin.Context) {
    if err := doSomething(); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "ok"})
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

// Or with custom error handling
func handler(c pkg.Context) error {
    if err := doSomething(); err != nil {
        return c.JSON(500, map[string]interface{}{"error": err.Error()})
    }
    return c.JSON(200, map[string]interface{}{"status": "ok"})
}
```

### 7. Static Files

**Gin:**
```go
router.Static("/static", "./public")
router.StaticFile("/favicon.ico", "./favicon.ico")
router.StaticFS("/assets", http.Dir("./assets"))
```

**Rockstar:**
```go
router := app.Router()
router.Static("/static", "./public")
router.GET("/favicon.ico", func(c pkg.Context) error {
    return c.File("./favicon.ico")
})
```

### 8. Template Rendering

**Gin:**
```go
router.LoadHTMLGlob("templates/*")

func handler(c *gin.Context) {
    c.HTML(200, "index.html", gin.H{
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

## Additional Features in Rockstar

### Database Integration

Rockstar includes built-in database support:

```go
config := &pkg.Config{
    DatabaseDriver: "mysql",
    DatabaseDSN:    "user:pass@tcp(localhost:3306)/dbname",
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

## Migration Checklist

### Phase 1: Setup

- [ ] Install Rockstar framework: `go get github.com/echterhof/rockstar-web-framework/pkg`
- [ ] Create configuration struct with your settings
- [ ] Initialize framework with `pkg.New(config)`
- [ ] Update imports from `gin` to `pkg`

### Phase 2: Update Handlers

- [ ] Change handler signature from `func(c *gin.Context)` to `func(c pkg.Context) error`
- [ ] Update all handlers to return `error`
- [ ] Replace `gin.H` with `map[string]interface{}`
- [ ] Update `c.Query()` calls to `c.QueryParam()`
- [ ] Ensure all response methods return their result

### Phase 3: Update Routing

- [ ] Replace `gin.Engine` with `pkg.Framework`
- [ ] Update route registration to use `app.Router()`
- [ ] Convert route groups to use path prefixes or middleware
- [ ] Update middleware to new signature

### Phase 4: Update Middleware

- [ ] Convert middleware from `gin.HandlerFunc` to `pkg.MiddlewareFunc`
- [ ] Update middleware to wrap `next` handler and return error
- [ ] Replace `c.Next()` with `next(c)` call
- [ ] Update middleware registration

### Phase 5: Leverage New Features

- [ ] Replace manual database code with `DatabaseManager`
- [ ] Replace manual session handling with `SessionManager`
- [ ] Add caching where appropriate using `CacheManager`
- [ ] Implement authentication using `SecurityManager`
- [ ] Add i18n support if needed

### Phase 6: Testing

- [ ] Update tests to use Rockstar test helpers
- [ ] Test all routes and handlers
- [ ] Test middleware chain
- [ ] Test error handling
- [ ] Load test for performance comparison

### Phase 7: Deployment

- [ ] Update deployment configuration
- [ ] Configure production settings
- [ ] Set up monitoring and metrics
- [ ] Deploy and monitor

## Common Pitfalls

### 1. Forgetting to Return Errors

**Wrong:**
```go
func handler(c pkg.Context) error {
    c.JSON(200, data) // Missing return
}
```

**Correct:**
```go
func handler(c pkg.Context) error {
    return c.JSON(200, data)
}
```

### 2. Not Handling Framework Initialization Errors

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

### 3. Not Calling Shutdown

**Wrong:**
```go
app, _ := pkg.New(config)
app.Start()
```

**Correct:**
```go
app, _ := pkg.New(config)
defer app.Shutdown()
app.Start()
```

### 4. Incorrect Middleware Signature

**Wrong:**
```go
func middleware(c pkg.Context) error {
    // This is a handler, not middleware
    return next(c)
}
```

**Correct:**
```go
func middleware() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            // Middleware logic
            return next(c)
        }
    }
}
```

## Performance Considerations

Rockstar is designed for high performance similar to Gin:

- Zero-allocation routing for common cases
- Efficient memory management with arena allocators
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

### Before (Gin)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    router := gin.Default()
    
    router.GET("/users/:id", getUser)
    router.POST("/users", createUser)
    
    router.Run(":8080")
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    user := User{ID: 1, Name: "John"}
    c.JSON(http.StatusOK, user)
}

func createUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, user)
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

The migration is straightforward with minimal code changes, while gaining access to enterprise features like built-in database support, sessions, caching, and security.
