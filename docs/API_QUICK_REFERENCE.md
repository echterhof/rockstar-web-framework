# API Quick Reference Card

Quick reference for the most commonly used APIs in the Rockstar Web Framework.

## Framework Initialization

```go
import "github.com/echterhof/rockstar-web-framework/pkg"

config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        Address: ":8080",
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver: "sqlite",
        Database: "app.db",
    },
}

app, err := pkg.New(config)
```

[Full Documentation →](api/framework.md)

## Routing

```go
// Basic routes
app.Router().GET("/", handler)
app.Router().POST("/users", createUser)
app.Router().PUT("/users/:id", updateUser)
app.Router().DELETE("/users/:id", deleteUser)

// Route groups
api := app.Router().Group("/api")
api.GET("/users", listUsers)
api.POST("/users", createUser)

// Middleware
app.Router().Use(authMiddleware)
app.Router().GET("/admin", adminHandler, adminMiddleware)
```

[Full Documentation →](api/router.md)

## Context (Request/Response)

```go
func handler(ctx pkg.Context) error {
    // Request data
    id := ctx.Param("id")
    name := ctx.Query()["name"]
    body := ctx.Body()
    
    // Services
    db := ctx.DB()
    cache := ctx.Cache()
    
    // Response
    return ctx.JSON(200, map[string]string{
        "message": "Success",
    })
}
```

[Full Documentation →](api/context.md)

## Database

```go
// Query
rows, err := ctx.DB().Query("SELECT * FROM users WHERE id = ?", id)

// Execute
result, err := ctx.DB().Exec("INSERT INTO users (name) VALUES (?)", name)

// Transaction
tx, err := ctx.DB().Begin()
tx.Exec("INSERT INTO users (name) VALUES (?)", name)
tx.Commit()

// Framework models
session, err := ctx.DB().LoadSession(sessionID)
token, err := ctx.DB().ValidateAccessToken(tokenValue)
```

[Full Documentation →](api/database.md)

## Cache

```go
// Set with TTL
ctx.Cache().Set("key", value, 5*time.Minute)

// Get
value, err := ctx.Cache().Get("key")

// Delete
ctx.Cache().Delete("key")

// Tags
ctx.Cache().SetWithTags("key", value, 5*time.Minute, []string{"users"})
ctx.Cache().InvalidateByTag("users")
```

[Full Documentation →](api/cache.md)

## Sessions

```go
// Create session
session, err := ctx.Session().Create(ctx)

// Load session
session, err := ctx.Session().Load(ctx, sessionID)

// Set data
session.Set("user_id", userID)

// Get data
userID := session.Get("user_id")

// Save
ctx.Session().Save(ctx, session)

// Destroy
ctx.Session().Destroy(ctx, sessionID)
```

[Full Documentation →](api/session.md)

## Security

```go
// OAuth2
user, err := ctx.Security().AuthenticateOAuth2(token)

// JWT
user, err := ctx.Security().AuthenticateJWT(token)

// Authorization
allowed := ctx.Security().Authorize(userID, "users", "read")

// RBAC
allowed := ctx.Security().AuthorizeRole(userID, "admin")

// Password hashing
hash, err := ctx.Security().HashPassword(password)
valid := ctx.Security().VerifyPassword(password, hash)
```

[Full Documentation →](api/security.md)

## Forms & Validation ⭐

```go
// Parse form
parser := pkg.NewFormParser()
parser.ParseForm(ctx)

// Get values
name := parser.GetFormValue(ctx, "name")
email := parser.GetFormValue(ctx, "email")

// File upload
file, err := parser.GetFormFile(ctx, "avatar")

// Validate
validator := pkg.NewFormValidator()
rules := map[string][]pkg.ValidationRule{
    "email": {pkg.RequiredRule{}, pkg.EmailRule{}},
    "name": {pkg.RequiredRule{}, pkg.MinLengthRule{Min: 2}},
}
errors := validator.ValidateForm(ctx, rules)
```

[Full Documentation →](api/forms-validation.md)

## Error Handling ⭐

```go
// Framework errors
return pkg.NewValidationError("Invalid input")
return pkg.NewAuthenticationError("Invalid token")
return pkg.NewAuthorizationError("Access denied")
return pkg.NewNotFoundError("User")

// Custom errors
err := pkg.NewFrameworkError("CUSTOM_ERROR", "Message", 400)
err.WithCause(originalErr)
err.WithI18n("errors.custom", map[string]interface{}{"field": "email"})

// Error handler
app.SetErrorHandler(&CustomErrorHandler{})
```

[Full Documentation →](api/errors-recovery.md)

## Middleware ⭐

```go
// Register middleware
engine := app.MiddlewareEngine()
engine.Register(pkg.MiddlewareConfig{
    Name: "auth",
    Priority: 100,
    Routes: []string{"/api/*"},
    Handler: authMiddleware,
})

// Enable/disable
engine.Disable("auth")
engine.Enable("auth")

// Set priority
engine.SetPriority("cors", 10)
```

[Full Documentation →](api/pipeline-middleware-engine.md)

## Pipelines ⭐

```go
// Register pipeline
engine := app.PipelineEngine()
engine.Register(pkg.PipelineConfig{
    Name: "process-order",
    Stages: []pkg.PipelineStage{
        {Name: "validate", Handler: validateOrder},
        {Name: "payment", Handler: processPayment},
        {Name: "create", Handler: createOrder},
    },
})

// Execute
engine.Execute(ctx, "process-order")

// Chain
engine.ExecuteChain(ctx, []string{"validate", "process", "notify"})

// Parallel
engine.ExecuteParallel(ctx, []string{"email", "sms", "push"})
```

[Full Documentation →](api/pipeline-middleware-engine.md)

## I18n

```go
// Translate
message := ctx.I18n().Translate("welcome.message")

// With parameters
message := ctx.I18n().Translate("welcome.user", "name", userName)

// Pluralization
message := ctx.I18n().TranslatePlural("items.count", itemCount)

// Set language
ctx.I18n().SetLanguage("de")
```

[Full Documentation →](api/i18n.md)

## Metrics

```go
// Start request metrics
metrics := ctx.Metrics().Start(requestID)

// Record
ctx.Metrics().Record("api.requests", 1)
ctx.Metrics().RecordDuration("api.latency", duration)

// Get metrics
stats := ctx.Metrics().GetMetrics()
```

[Full Documentation →](api/metrics.md)

## Monitoring

```go
// Enable metrics endpoint
app.Monitoring().EnableMetricsEndpoint("/metrics")

// Enable profiling
app.Monitoring().EnableProfiling("/debug/pprof")

// Health check
app.Monitoring().RegisterHealthCheck("database", func() error {
    return db.Ping()
})
```

[Full Documentation →](api/monitoring.md)

## WebSockets

```go
// WebSocket handler
app.Router().WebSocket("/ws", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    for {
        msgType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Process message
        response := processMessage(data)
        
        // Send response
        conn.WriteMessage(msgType, response)
    }
})
```

[Full Documentation →](api/websockets.md)

## Plugins

```go
// Load plugins
app.Plugins().DiscoverPlugins()
app.Plugins().LoadPluginsFromConfig("plugins.yaml")
app.Plugins().InitializeAll()
app.Plugins().StartAll()

// Get plugin
plugin, err := app.Plugins().GetPlugin("my-plugin")

// Plugin context
ctx.PluginContext().Router().GET("/plugin-route", handler)
ctx.PluginContext().Database().Query("SELECT ...")
```

[Full Documentation →](api/plugins.md)

## Configuration

```go
// Load config
app.Config().Load("config.yaml")
app.Config().LoadFromEnv()

// Get values
host := app.Config().GetString("server.host")
port := app.Config().GetInt("server.port")
enabled := app.Config().GetBool("features.cache")

// Set values
app.Config().SetString("server.host", "localhost")
```

[Full Documentation →](api/configuration.md)

## Logging

```go
// Log levels
ctx.Logger().Debug("Debug message", "key", value)
ctx.Logger().Info("Info message", "key", value)
ctx.Logger().Warn("Warning message", "key", value)
ctx.Logger().Error("Error message", "error", err)

// With request ID
ctx.Logger().WithRequestID(requestID).Info("Request processed")
```

[Full Documentation →](api/utilities.md)

## Templates

```go
// Load templates
app.Templates().LoadTemplates(os.DirFS("templates"), "*.html")

// Render
return ctx.HTML(200, "index.html", map[string]interface{}{
    "title": "Home",
    "user": user,
})
```

[Full Documentation →](api/templates-responses.md)

## Response Helpers

```go
// JSON
ctx.JSON(200, data)

// XML
ctx.XML(200, data)

// HTML
ctx.HTML(200, "template.html", data)

// String
ctx.String(200, "Hello, World!")

// Redirect
ctx.Redirect(302, "/login")

// File
ctx.File("path/to/file.pdf")
```

[Full Documentation →](api/context.md)

## Common Patterns

### REST API Handler

```go
func createUser(ctx pkg.Context) error {
    // Parse and validate
    var user User
    if err := json.Unmarshal(ctx.Body(), &user); err != nil {
        return pkg.NewValidationError("Invalid JSON")
    }
    
    // Save to database
    result, err := ctx.DB().Exec(
        "INSERT INTO users (name, email) VALUES (?, ?)",
        user.Name, user.Email,
    )
    if err != nil {
        return pkg.NewFrameworkError("DB_ERROR", "Failed to create user", 500)
    }
    
    // Return response
    return ctx.JSON(201, map[string]interface{}{
        "id": result.LastInsertId(),
        "message": "User created",
    })
}
```

### Authenticated Handler

```go
func protectedHandler(ctx pkg.Context) error {
    // Check authentication
    if !ctx.IsAuthenticated() {
        return pkg.NewAuthenticationError("Authentication required")
    }
    
    // Check authorization
    if !ctx.IsAuthorized("resource", "action") {
        return pkg.NewAuthorizationError("Access denied")
    }
    
    // Process request
    user := ctx.User()
    return ctx.JSON(200, user)
}
```

### Cached Handler

```go
func expensiveHandler(ctx pkg.Context) error {
    cacheKey := "expensive_data"
    
    // Try cache
    if cached, err := ctx.Cache().Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Compute
    data := computeExpensiveData()
    
    // Cache for 5 minutes
    ctx.Cache().Set(cacheKey, data, 5*time.Minute)
    
    return ctx.JSON(200, data)
}
```

## See Also

- **[Complete API Reference](api/README.md)** - All APIs with detailed documentation
- **[API Documentation Summary](API_DOCUMENTATION_SUMMARY.md)** - Organized by category
- **[User Guides](guides/README.md)** - Feature-specific guides
- **[Examples](examples/README.md)** - Working code examples

---

**Version**: 1.0.0 | **Coverage**: 100% (61/61 interfaces)
