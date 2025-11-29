# Getting Started Example

The Getting Started example (`examples/getting_started.go`) is the perfect introduction to the Rockstar Web Framework. It demonstrates the core concepts you need to build web applications, including routing, middleware, handlers, lifecycle hooks, and error handling.

## What This Example Demonstrates

- **Framework initialization** with configuration
- **Secure session management** with encryption keys
- **Lifecycle hooks** for startup and shutdown
- **Global middleware** for logging and recovery
- **Custom error handling** for consistent error responses
- **RESTful routing** with all HTTP methods
- **Path parameters** for dynamic routes
- **JSON responses** for API endpoints
- **Request handling** patterns

## Prerequisites

- Go 1.25 or higher
- SQLite (included with the framework)

## Setup Instructions

### 1. Generate Encryption Key

The example requires a session encryption key for secure session management. Generate one using:

```bash
go run examples/generate_keys.go
```

This will output a secure 32-byte hex-encoded key. Save this key!

### 2. Set Environment Variable

Set the encryption key as an environment variable:

```bash
# On Linux/macOS
export SESSION_ENCRYPTION_KEY=<your_generated_key>

# On Windows (PowerShell)
$env:SESSION_ENCRYPTION_KEY="<your_generated_key>"

# On Windows (CMD)
set SESSION_ENCRYPTION_KEY=<your_generated_key>
```

**Note**: If you don't set this variable, the example will generate a temporary key for development. This is **insecure for production** but convenient for testing.

### 3. Run the Example

```bash
go run examples/getting_started.go
```

The server will start on `http://localhost:8080`.

## Testing the Endpoints

Once the server is running, try these commands:

```bash
# Home endpoint - returns welcome message
curl http://localhost:8080/

# Hello endpoint with path parameter
curl http://localhost:8080/hello/World

# Create user (POST)
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# Get user (GET with parameter)
curl http://localhost:8080/api/users/123

# Update user (PUT)
curl -X PUT http://localhost:8080/api/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com"}'

# Delete user (DELETE)
curl -X DELETE http://localhost:8080/api/users/123
```

## Code Walkthrough

### Configuration Setup

The example starts by creating a `FrameworkConfig` with minimal settings. Most values use sensible defaults:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,
        EnableHTTP2: true,
        // ReadTimeout, WriteTimeout, etc. use defaults
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "getting_started.db",
        // Connection pool settings use defaults
    },
    SessionConfig: pkg.SessionConfig{
        EncryptionKey: encryptionKey, // 32 bytes for AES-256
        CookieSecure:  false,         // Set true in production
    },
}
```

**Key Points**:
- Only specify configuration values you need to change
- Defaults are production-ready for most use cases
- Session encryption requires a 32-byte key for AES-256
- Set `CookieSecure: true` when using HTTPS in production

### Framework Initialization

Create the framework instance with your configuration:

```go
app, err := pkg.New(config)
if err != nil {
    log.Fatalf("Failed to create framework: %v", err)
}
```

The framework initializes all managers (database, cache, session, etc.) based on your configuration.

### Lifecycle Hooks

Register hooks that run during server startup and shutdown:

```go
// Startup hook - runs when server starts
app.RegisterStartupHook(func(ctx context.Context) error {
    fmt.Println("🚀 Server starting up...")
    // Initialize resources, load data, etc.
    return nil
})

// Shutdown hook - runs during graceful shutdown
app.RegisterShutdownHook(func(ctx context.Context) error {
    fmt.Println("👋 Server shutting down...")
    // Clean up resources, close connections, etc.
    return nil
})
```

**Use Cases**:
- **Startup**: Initialize database connections, load configuration, start background workers
- **Shutdown**: Close connections, flush caches, save state, stop workers

### Global Middleware

Add middleware that runs for every request:

```go
// Logging middleware - logs all requests
app.Use(loggingMiddleware)

// Recovery middleware - recovers from panics
app.Use(recoveryMiddleware)
```

Middleware functions have this signature:

```go
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Before handler
    start := time.Now()
    
    // Call next handler
    err := next(ctx)
    
    // After handler
    duration := time.Since(start)
    fmt.Printf("Request completed in %v\n", duration)
    
    return err
}
```

**Middleware Order**: Middleware executes in the order registered. The example uses:
1. Logging (first) - logs all requests
2. Recovery (second) - catches panics from handlers

### Custom Error Handler

Set a custom error handler for consistent error responses:

```go
app.SetErrorHandler(customErrorHandler)

func customErrorHandler(ctx pkg.Context, err error) error {
    fmt.Printf("❌ Error: %v\n", err)
    return ctx.JSON(500, map[string]interface{}{
        "error":   "Internal server error",
        "message": err.Error(),
    })
}
```

This handler is called whenever a handler returns an error.

### Route Registration

Register routes using the router:

```go
router := app.Router()

// GET route - simple endpoint
router.GET("/", homeHandler)

// GET route with path parameter
router.GET("/hello/:name", helloHandler)

// RESTful routes
router.POST("/api/users", createUserHandler)
router.GET("/api/users/:id", getUserHandler)
router.PUT("/api/users/:id", updateUserHandler)
router.DELETE("/api/users/:id", deleteUserHandler)
```

**Path Parameters**: Use `:name` syntax for dynamic segments. Access them via `ctx.Params()["name"]`.

### Handler Functions

Handlers process requests and return responses:

```go
func homeHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]interface{}{
        "message": "Welcome to Rockstar! 🎸",
        "version": "1.0.0",
    })
}

func helloHandler(ctx pkg.Context) error {
    // Extract path parameter
    name := ctx.Params()["name"]
    
    return ctx.JSON(200, map[string]interface{}{
        "message": fmt.Sprintf("Hello, %s! 👋", name),
    })
}
```

**Handler Patterns**:
- Return `ctx.JSON()` for JSON responses
- Return `ctx.String()` for text responses
- Return `ctx.HTML()` for HTML responses
- Return `error` to trigger error handler
- Access request data via `ctx.Request()`
- Access services via `ctx.DB()`, `ctx.Cache()`, etc.

### CRUD Operations

The example shows typical CRUD patterns:

```go
// CREATE - POST /api/users
func createUserHandler(ctx pkg.Context) error {
    // 1. Parse request body
    // 2. Validate input
    // 3. Save to database: ctx.DB().Exec(...)
    // 4. Return created resource
    return ctx.JSON(201, map[string]interface{}{
        "message": "User created",
        "id":      "123",
    })
}

// READ - GET /api/users/:id
func getUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    // 1. Query database: ctx.DB().Query(...)
    // 2. Handle not found
    // 3. Return user data
    return ctx.JSON(200, map[string]interface{}{
        "id": userID,
    })
}

// UPDATE - PUT /api/users/:id
func updateUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    // 1. Parse request body
    // 2. Validate input
    // 3. Update database: ctx.DB().Exec(...)
    // 4. Return updated resource
    return ctx.JSON(200, map[string]interface{}{
        "message": "User updated",
    })
}

// DELETE - DELETE /api/users/:id
func deleteUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    // 1. Check if exists
    // 2. Delete from database: ctx.DB().Exec(...)
    // 3. Return confirmation
    return ctx.JSON(200, map[string]interface{}{
        "message": "User deleted",
    })
}
```

### Server Startup

Start the server and listen for requests:

```go
if err := app.Listen(":8080"); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

This blocks until the server shuts down. The server handles:
- Graceful shutdown on SIGINT/SIGTERM
- Connection draining
- Shutdown hook execution

## Key Concepts Explained

### Context

The `Context` interface is your gateway to all framework features:

```go
func handler(ctx pkg.Context) error {
    // Request data
    method := ctx.Request().Method
    params := ctx.Params()
    query := ctx.Query()
    
    // Services
    db := ctx.DB()
    cache := ctx.Cache()
    session := ctx.Session()
    
    // Response
    return ctx.JSON(200, data)
}
```

### Middleware Chain

Middleware forms a chain where each middleware can:
1. Process the request before the handler
2. Call `next(ctx)` to continue the chain
3. Process the response after the handler

```
Request → Logging → Recovery → Handler → Recovery → Logging → Response
```

### Error Handling

Errors flow through the middleware chain:
1. Handler returns error
2. Middleware can handle or pass through
3. Error handler provides final response

### Configuration Defaults

The framework provides sensible defaults for all configuration:
- **Server**: 30s timeouts, 10000 max connections, 10MB max request size
- **Database**: 25 max connections, 5 idle connections, 5min connection lifetime
- **Cache**: Unlimited size, no expiration
- **Session**: 24h lifetime, 1h cleanup interval

Only override what you need!

## Production Considerations

When moving to production:

1. **Use environment variables** for sensitive data:
   ```go
   encryptionKey := os.Getenv("SESSION_ENCRYPTION_KEY")
   dbPassword := os.Getenv("DB_PASSWORD")
   ```

2. **Enable HTTPS** and secure cookies:
   ```go
   SessionConfig: pkg.SessionConfig{
       CookieSecure: true,  // Require HTTPS
       CookieHTTPOnly: true, // Prevent JavaScript access
       CookieSameSite: "Strict",
   }
   ```

3. **Use a production database** (PostgreSQL, MySQL, MSSQL):
   ```go
   DatabaseConfig: pkg.DatabaseConfig{
       Driver:   "postgres",
       Host:     "db.example.com",
       Database: "production_db",
       // ... other settings
   }
   ```

4. **Add proper error handling**:
   ```go
   // Parse and validate input
   // Handle database errors
   // Return appropriate status codes
   ```

5. **Implement authentication and authorization**:
   ```go
   // Use ctx.Security() for auth
   // Check permissions before operations
   ```

## Next Steps

After understanding this example:

1. **Explore API styles**: Try [REST API Example](rest-api.md) for more advanced patterns
2. **Add features**: Learn about caching, sessions, i18n from other examples
3. **Study production patterns**: Review [Full Featured App](full-featured-app.md)
4. **Read the guides**: Check [Feature Guides](../guides/README.md) for in-depth documentation

## Common Issues

### "SESSION_ENCRYPTION_KEY not set"

**Solution**: Generate and set the encryption key:
```bash
export SESSION_ENCRYPTION_KEY=$(go run examples/generate_keys.go)
```

### "Failed to create framework"

**Solution**: Check your configuration, especially database settings. For SQLite, ensure the directory is writable.

### "Address already in use"

**Solution**: Another process is using port 8080. Either stop that process or change the port:
```go
app.Listen(":8081")
```

## Related Documentation

- [Configuration Guide](../guides/configuration.md) - All configuration options
- [Routing Guide](../guides/routing.md) - Advanced routing patterns
- [Middleware Guide](../guides/middleware.md) - Creating custom middleware
- [Context API](../api/context.md) - Complete Context interface reference
- [Framework API](../api/framework.md) - Framework methods and lifecycle

## Source Code

The complete source code for this example is available at `examples/getting_started.go` in the repository.
