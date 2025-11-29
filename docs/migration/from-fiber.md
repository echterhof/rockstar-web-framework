# Migrating from Fiber to Rockstar

This guide helps you migrate your application from the Fiber web framework to Rockstar. While Fiber is inspired by Express.js and focuses on extreme performance, Rockstar provides similar performance with additional enterprise features like multi-tenancy, plugin system, and comprehensive security.

## Overview

Fiber is an Express-inspired web framework built on top of Fasthttp, designed for extreme performance. Rockstar provides comparable performance using the standard library while adding enterprise-grade features including multi-protocol support, built-in database abstraction, and comprehensive security out of the box.

## Concept Mapping

| Fiber Concept | Rockstar Equivalent | Notes |
|-------------|---------------------|-------|
| `fiber.App` | `pkg.Framework` | Main application instance |
| `fiber.Ctx` | `pkg.Context` | Request/response context |
| `fiber.Handler` | `pkg.HandlerFunc` | Handler function signature |
| `fiber.Router` | Route groups via `Router()` | Similar grouping functionality |
| `fiber.Map` | `map[string]interface{}` | JSON response helper |
| Middleware | Middleware via `Use()` | Similar middleware pattern |
| `c.Params()` | `c.Param()` | Path parameter extraction |
| `c.Query()` | `c.QueryParam()` | Query parameter extraction |
| `c.BodyParser()` | `c.Bind()` | Request body binding |
| `c.JSON()` | `c.JSON()` | JSON response |
| `c.SendString()` | `c.String()` | String response |
| `c.Render()` | `c.Render()` | Template rendering |
| `c.Redirect()` | `c.Redirect()` | HTTP redirect |
| `c.SendFile()` | `c.File()` | File serving |
| `c.Locals()` | `c.Set()` / `c.Get()` | Context value storage |
| Manual session handling | `SessionManager` | Built-in session support |
| Manual database setup | `DatabaseManager` | Built-in database abstraction |

## Key Differences

### 1. Initialization

**Fiber:**
```go
import "github.com/gofiber/fiber/v2"

app := fiber.New()
// or with config
app := fiber.New(fiber.Config{
    Prefork: true,
    ServerHeader: "Fiber",
})
```

**Rockstar:**
```go
import "github.com/echterhof/rockstar-web-framework/pkg"

config := &pkg.Config{
    ServerAddress: ":8080",
    Environment:   "development",
    Prefork:       true,
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
defer app.Shutdown()
```

### 2. Routing

**Fiber:**
```go
app.Get("/users/:id", getUser)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
```

**Rockstar:**
```go
router := app.Router()

router.GET("/users/:id", getUser)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

**Key Difference:** Rockstar uses uppercase HTTP method names (GET vs Get).

### 3. Handler Signature

**Fiber:**
```go
func handler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "message": "Hello",
    })
}
```

**Rockstar:**
```go
func handler(c pkg.Context) error {
    return c.JSON(200, map[string]interface{}{
        "message": "Hello",
    })
}
```

**Key Differences:**
- Rockstar uses interface `pkg.Context` instead of pointer `*fiber.Ctx`
- Rockstar requires explicit status code in JSON responses
- Use `map[string]interface{}` instead of `fiber.Map`

### 4. Route Groups

**Fiber:**
```go
api := app.Group("/api")
api.Use(logger)

v1 := api.Group("/v1")
v1.Get("/users", listUsers)
v1.Post("/users", createUser)

admin := app.Group("/admin")
admin.Use(authMiddleware)
admin.Get("/dashboard", dashboard)
```

**Rockstar:**
```go
router := app.Router()

// API routes with middleware
router.Use("/api/*", logger())

// V1 routes
router.GET("/api/v1/users", listUsers)
router.POST("/api/v1/users", createUser)

// Admin routes with auth
router.Use("/admin/*", authMiddleware())
router.GET("/admin/dashboard", dashboard)
```

### 5. Middleware

**Fiber:**
```go
func logger(c *fiber.Ctx) error {
    start := time.Now()
    
    err := c.Next()
    
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
    
    return err
}

app.Use(logger)
```

**Rockstar:**
```go
func logger() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            start := time.Now()
            
            err := next(c)
            
            duration := time.Since(start)
            log.Printf("Request took %v", duration)
            
            return err
        }
    }
}

app.Use(logger())
```

**Key Difference:** Rockstar middleware returns a function that wraps the next handler, while Fiber middleware calls `c.Next()`.

### 6. Request Handling

**Fiber:**
```go
func getUser(c *fiber.Ctx) error {
    id := c.Params("id")
    name := c.Query("name")
    
    var user User
    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(user)
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

**Key Differences:**
- `c.Params()` → `c.Param()`
- `c.Query()` → `c.QueryParam()`
- `c.BodyParser()` → `c.Bind()`
- `c.Status().JSON()` → `c.JSON(statusCode, data)`

### 7. Response Methods

**Fiber:**
```go
// JSON response
c.JSON(fiber.Map{"key": "value"})

// String response
c.SendString("Hello")

// Status code
c.Status(404).SendString("Not Found")

// Redirect
c.Redirect("/new-path")

// File
c.SendFile("./file.pdf")
```

**Rockstar:**
```go
// JSON response
c.JSON(200, map[string]interface{}{"key": "value"})

// String response
c.String(200, "Hello")

// Status code with string
c.String(404, "Not Found")

// Redirect
c.Redirect(302, "/new-path")

// File
c.File("./file.pdf")
```

**Key Difference:** Rockstar requires explicit status codes in most response methods.

### 8. Static Files

**Fiber:**
```go
app.Static("/", "./public")
app.Static("/static", "./assets")
```

**Rockstar:**
```go
router := app.Router()
router.Static("/", "./public")
router.Static("/static", "./assets")
```

### 9. Template Rendering

**Fiber:**
```go
import "github.com/gofiber/template/html"

engine := html.New("./views", ".html")
app := fiber.New(fiber.Config{
    Views: engine,
})

func handler(c *fiber.Ctx) error {
    return c.Render("index", fiber.Map{
        "Title": "Home",
    })
}
```

**Rockstar:**
```go
config := &pkg.Config{
    TemplateDir: "./views",
}

app, _ := pkg.New(config)

func handler(c pkg.Context) error {
    return c.Render(200, "index.html", map[string]interface{}{
        "Title": "Home",
    })
}
```

### 10. Context Values

**Fiber:**
```go
// Set value
c.Locals("user", user)

// Get value
user := c.Locals("user").(User)
```

**Rockstar:**
```go
// Set value
c.Set("user", user)

// Get value
user := c.Get("user").(User)
```

### 11. Cookies

**Fiber:**
```go
// Set cookie
c.Cookie(&fiber.Cookie{
    Name:  "token",
    Value: "abc123",
    MaxAge: 3600,
})

// Get cookie
token := c.Cookies("token")
```

**Rockstar:**
```go
// Set cookie
c.SetCookie("token", "abc123", 3600)

// Get cookie
token := c.Cookie("token")
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
- [ ] Update imports from `fiber` to `pkg`

### Phase 2: Update Application Structure

- [ ] Replace `fiber.New()` with `pkg.New(config)`
- [ ] Update `app.Listen()` to `app.Start()`
- [ ] Add `defer app.Shutdown()` for cleanup
- [ ] Update router access from `app` to `app.Router()`

### Phase 3: Update Handlers

- [ ] Change handler signature from `*fiber.Ctx` to `pkg.Context`
- [ ] Update `fiber.Map` to `map[string]interface{}`
- [ ] Add explicit status codes to response methods
- [ ] Update parameter extraction methods:
  - `c.Params()` → `c.Param()`
  - `c.Query()` → `c.QueryParam()`
  - `c.BodyParser()` → `c.Bind()`
- [ ] Update response methods:
  - `c.JSON()` → `c.JSON(statusCode, data)`
  - `c.SendString()` → `c.String(statusCode, text)`
  - `c.SendFile()` → `c.File(path)`

### Phase 4: Update Routing

- [ ] Update HTTP method names to uppercase (Get → GET)
- [ ] Update route registration to use `app.Router()`
- [ ] Convert route groups to use path prefixes or middleware
- [ ] Update static file serving

### Phase 5: Update Middleware

- [ ] Convert middleware to return `pkg.MiddlewareFunc`
- [ ] Replace `c.Next()` with `next(c)` call
- [ ] Update middleware registration from `app.Use()` to `app.Use()`
- [ ] Test middleware chain execution

### Phase 6: Update Context Operations

- [ ] Replace `c.Locals()` with `c.Set()` / `c.Get()`
- [ ] Update cookie operations
- [ ] Update header operations if needed

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

### 1. Missing Status Codes

**Wrong:**
```go
func handler(c pkg.Context) error {
    return c.JSON(data) // Missing status code
}
```

**Correct:**
```go
func handler(c pkg.Context) error {
    return c.JSON(200, data)
}
```

### 2. Using Pointer Context

**Wrong:**
```go
func handler(c *pkg.Context) error { // Wrong: pointer
    return c.JSON(200, data)
}
```

**Correct:**
```go
func handler(c pkg.Context) error { // Correct: interface
    return c.JSON(200, data)
}
```

### 3. Incorrect Middleware Pattern

**Wrong:**
```go
func middleware(c pkg.Context) error {
    // Before logic
    err := c.Next() // No c.Next() in Rockstar
    // After logic
    return err
}
```

**Correct:**
```go
func middleware() pkg.MiddlewareFunc {
    return func(next pkg.HandlerFunc) pkg.HandlerFunc {
        return func(c pkg.Context) error {
            // Before logic
            err := next(c)
            // After logic
            return err
        }
    }
}
```

### 4. Forgetting Shutdown

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

### 5. Using Fiber-Specific Packages

Remove dependencies on Fiber-specific packages:
- `fiber.Map` → `map[string]interface{}`
- `fiber.Cookie` → Use `c.SetCookie()` method
- Fiber middleware packages → Use Rockstar built-in features

## Performance Considerations

Fiber is built on Fasthttp for extreme performance. Rockstar uses the standard library but is still highly performant:

- Efficient routing with zero allocations for common cases
- Arena-based memory management
- Connection pooling for databases
- Built-in caching reduces external dependencies
- Request-level context prevents memory leaks

Benchmark your application after migration. While Fiber may have a slight edge in raw throughput, Rockstar's enterprise features and standard library compatibility often provide better overall value.

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

### Before (Fiber)

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "log"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    app := fiber.New()
    
    app.Get("/users/:id", getUser)
    app.Post("/users", createUser)
    
    log.Fatal(app.Listen(":8080"))
}

func getUser(c *fiber.Ctx) error {
    id := c.Params("id")
    user := User{ID: 1, Name: "John"}
    return c.JSON(user)
}

func createUser(c *fiber.Ctx) error {
    var user User
    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }
    return c.Status(201).JSON(user)
}
```

### After (Rockstar)

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "log"
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
    return c.JSON(200, user)
}

func createUser(c pkg.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return c.JSON(400, map[string]interface{}{"error": err.Error()})
    }
    return c.JSON(201, user)
}
```

The migration from Fiber requires more changes than from Echo or Gin, primarily around method names and explicit status codes, but the overall structure remains similar. The benefits include standard library compatibility and comprehensive enterprise features.
