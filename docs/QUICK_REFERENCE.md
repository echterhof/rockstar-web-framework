# Rockstar Web Framework - Quick Reference

**Version**: 1.0.0 | **Go**: 1.25.4 | **Updated**: November 25, 2025

## Installation

```bash
go get github.com/echterhof/rockstar-web-framework/pkg
```

## Basic Setup

```go
package main

import (
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
            EnableHTTP2:  true,
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    app.Router().GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]string{"message": "Hello!"})
    })
    
    log.Fatal(app.Listen(":8080"))
}
```

## Routing

### HTTP Methods

```go
router := app.Router()

router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
router.PATCH("/users/:id", patchUser)
```

### Route Parameters

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    return ctx.JSON(200, map[string]string{"id": id})
})
```

### Query Parameters

```go
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query()["q"]
    return ctx.JSON(200, map[string]string{"query": query})
})
```

### Route Groups

```go
api := router.Group("/api/v1", authMiddleware)
api.GET("/users", listUsers)
api.POST("/users", createUser)
```

### Host-Based Routing

```go
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHome)
```

## Context

### Request Data

```go
func handler(ctx pkg.Context) error {
    // URL parameters
    id := ctx.Params()["id"]
    
    // Query strings
    search := ctx.Query()["q"]
    
    // Headers
    auth := ctx.GetHeader("Authorization")
    
    // Body
    body := ctx.Body()
    
    // Form values
    username := ctx.FormValue("username")
    
    // Files
    file, err := ctx.FormFile("avatar")
    
    return nil
}
```

### Response Methods

```go
// JSON response
ctx.JSON(200, data)

// String response
ctx.String(200, "Hello")

// HTML response
ctx.HTML(200, "template.html", data)

// XML response
ctx.XML(200, data)

// Redirect
ctx.Redirect(302, "/login")
```

### Framework Services

```go
// Database
db := ctx.DB()

// Cache
cache := ctx.Cache()

// Session
session := ctx.Session()

// Config
config := ctx.Config()

// I18n
i18n := ctx.I18n()

// Logger
logger := ctx.Logger()

// Metrics
metrics := ctx.Metrics()

// Files
files := ctx.Files()
```

## Middleware

### Global Middleware

```go
app.Use(loggingMiddleware)
app.Use(recoveryMiddleware)
```

### Route Middleware

```go
router.GET("/admin", adminHandler, authMiddleware, adminMiddleware)
```

### Custom Middleware

```go
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    log.Printf("Request: %s %s", ctx.Request().Method, ctx.Request().Path)
    
    err := next(ctx)
    
    log.Printf("Completed in %v", time.Since(start))
    return err
}
```

## Database

### Configuration

```go
DatabaseConfig: pkg.DatabaseConfig{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "mydb",
    Username:        "user",
    Password:        "pass",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}
```

### Queries

```go
// Query
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)

// Execute
result, err := db.Exec("INSERT INTO users (name) VALUES (?)", name)

// Transaction
tx, err := db.Begin()
tx.Exec("INSERT INTO users ...")
tx.Commit()
```

## Caching

### Basic Operations

```go
cache := ctx.Cache()

// Set
cache.Set("key", value, 5*time.Minute)

// Get
value, err := cache.Get("key")

// Delete
cache.Delete("key")

// Exists
exists := cache.Exists("key")
```

### Request Cache

```go
reqCache := cache.GetRequestCache(requestID)
reqCache.Set("key", value)
value := reqCache.Get("key")
```

## Sessions

### Session Operations

```go
session := ctx.Session()

// Get
userID := session.Get("user_id")

// Set
session.Set("user_id", "123")

// Delete
session.Delete("user_id")

// Save
session.Save(ctx)

// Destroy
session.Destroy(ctx)
```

## Configuration

### Load Configuration

```go
config := pkg.NewConfigManager()
config.Load("config.yaml")

// Get values
appName := config.GetString("app_name")
port := config.GetInt("port")
debug := config.GetBool("debug")
timeout := config.GetDuration("timeout")
```

### Environment Variables

```bash
export ROCKSTAR_APP_NAME=MyApp
export ROCKSTAR_PORT=8080
```

```go
config.LoadFromEnv()
appName := config.GetString("app.name")
```

## Internationalization

### Setup

```go
I18nConfig: pkg.I18nConfig{
    DefaultLocale:     "en",
    LocalesDir:        "./locales",
    SupportedLocales:  []string{"en", "de", "fr"},
    FallbackToDefault: true,
}
```

### Usage

```go
i18n := ctx.I18n()

// Translate
message := i18n.Translate("welcome_message")

// With parameters
message := i18n.Translate("error.missing_field", map[string]interface{}{
    "field": "email",
})

// Change locale
i18n.SetLocale("de")
```

## Templates

### Setup

```go
tm := pkg.NewTemplateManager()
tm.LoadTemplates(os.DirFS("templates"), "*.html")
```

### Render

```go
ctx.Response().SetTemplateManager(tm)
return ctx.HTML(200, "index.html", map[string]interface{}{
    "Title": "Home",
    "User":  user,
})
```

## Error Handling

### Create Errors

```go
// Authentication error
err := pkg.NewAuthenticationError("Invalid credentials")

// Validation error
err := pkg.NewValidationError("Invalid email", "email")

// Not found error
err := pkg.NewNotFoundError("user")

// Rate limit error
err := pkg.NewRateLimitError("Too many requests", 100, "1m")
```

### Handle Errors

```go
func handler(ctx pkg.Context) error {
    if err := someOperation(); err != nil {
        return errorHandler.HandleError(ctx, err)
    }
    return nil
}
```

## REST API

### Register Routes

```go
restAPI := pkg.NewRESTAPIManager(router, db)

config := pkg.RESTRouteConfig{
    RateLimit: &pkg.RESTRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
}

restAPI.RegisterRoute("GET", "/api/users", handler, config)
```

## GraphQL

### Register Schema

```go
graphqlManager := pkg.NewGraphQLManager(router, db, authManager)

config := pkg.GraphQLConfig{
    EnableIntrospection: true,
    EnablePlayground:    true,
    RequireAuth:         true,
}

graphqlManager.RegisterSchema("/graphql", schema, config)
```

## gRPC

### Register Service

```go
grpcManager := pkg.NewGRPCManager(router, db, authManager)

config := pkg.GRPCConfig{
    RequireAuth: true,
    RateLimit: &pkg.GRPCRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
    },
}

grpcManager.RegisterService(service, config)
```

## SOAP

### Register Service

```go
soapManager := pkg.NewSOAPManager(router, db, authManager)

config := pkg.SOAPConfig{
    EnableWSDL:  true,
    ServiceName: "MyService",
    Namespace:   "http://example.com/soap",
}

soapManager.RegisterService("/soap/service", service, config)
```

## Monitoring

### Metrics

```go
MonitoringConfig: pkg.MonitoringConfig{
    EnableMetrics: true,
    MetricsPath:   "/metrics",
    EnablePprof:   true,
    PprofPath:     "/debug/pprof",
}
```

### Access Metrics

```bash
# Metrics endpoint
curl http://localhost:8080/metrics

# CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Heap profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

## Security

### Authentication

```go
SecurityConfig: pkg.SecurityConfig{
    EnableXFrameOptions: true,
    XFrameOptions:       "DENY",
    EnableCORS:          true,
    EnableCSRF:          true,
    EnableXSS:           true,
    MaxRequestSize:      10 * 1024 * 1024,
    RequestTimeout:      30 * time.Second,
}
```

### Check Authentication

```go
if !ctx.IsAuthenticated() {
    return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
}

user := ctx.User()
```

### Check Authorization

```go
if !ctx.IsAuthorized("resource", "action") {
    return ctx.JSON(403, map[string]string{"error": "Forbidden"})
}
```

## Multi-Server

### Register Hosts

```go
sm := pkg.NewServerManager()

hostConfig := pkg.HostConfig{
    Hostname:  "api.example.com",
    TenantID:  "tenant1",
    VirtualFS: fs,
}

sm.RegisterHost("api.example.com", hostConfig)
sm.RegisterTenant("tenant1", []string{"api.example.com"})
```

## Forward Proxy

### Setup Proxy

```go
proxyManager := pkg.NewProxyManager(config, cache)

backend := &pkg.Backend{
    ID:       "backend1",
    URL:      "http://localhost:8081",
    Weight:   1,
    IsActive: true,
}

proxyManager.AddBackend(backend)
```

## Lifecycle Hooks

### Startup Hook

```go
app.RegisterStartupHook(func(ctx context.Context) error {
    log.Println("Application starting...")
    return nil
})
```

### Shutdown Hook

```go
app.RegisterShutdownHook(func(ctx context.Context) error {
    log.Println("Application shutting down...")
    return nil
})
```

## Graceful Shutdown

```go
// Shutdown with timeout
app.Shutdown(30 * time.Second)
```

## Common Patterns

### API Handler

```go
func getUser(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    user, err := fetchUser(ctx.DB(), id)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "User not found"})
    }
    
    return ctx.JSON(200, user)
}
```

### Authenticated Handler

```go
func protectedHandler(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    user := ctx.User()
    // Process request
    
    return ctx.JSON(200, data)
}
```

### Cached Handler

```go
func cachedHandler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    data, err := cache.Get("key")
    if err == nil {
        return ctx.JSON(200, data)
    }
    
    data = fetchData()
    cache.Set("key", data, 5*time.Minute)
    
    return ctx.JSON(200, data)
}
```

## Testing

### Unit Test

```go
func TestHandler(t *testing.T) {
    ctx := createMockContext()
    
    err := handler(ctx)
    if err != nil {
        t.Errorf("Handler failed: %v", err)
    }
}
```

### Integration Test

```go
func TestAPI(t *testing.T) {
    app, _ := pkg.New(config)
    app.Router().GET("/test", testHandler)
    
    // Test the endpoint
    resp := makeRequest(app, "GET", "/test")
    if resp.StatusCode != 200 {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Deployment

### Build

```bash
go build -o app cmd/rockstar/main.go
```

### Run

```bash
./app -addr :8080 -config config.yaml
```

### Docker

```bash
docker build -t myapp .
docker run -p 8080:8080 myapp
```

## Troubleshooting

### Enable Debug Logging

```go
LoggerConfig: pkg.LoggerConfig{
    Level: "debug",
}
```

### Check Health

```bash
curl http://localhost:8080/health
```

### View Metrics

```bash
curl http://localhost:8080/metrics
```

## Resources

- **Documentation**: [docs/](.)
- **Examples**: [examples/](../examples/)
- **API Reference**: [API_REFERENCE.md](API_REFERENCE.md)
- **Getting Started**: [GETTING_STARTED.md](GETTING_STARTED.md)
- **Deployment**: [DEPLOYMENT.md](DEPLOYMENT.md)

---

**For detailed information, see the full documentation at [docs/DOCUMENTATION_INDEX.md](DOCUMENTATION_INDEX.md)**
