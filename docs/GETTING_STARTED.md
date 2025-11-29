---
title: "Getting Started"
description: "Quick start guide for building your first application with Rockstar"
category: "guide"
tags: ["getting-started", "tutorial", "quickstart"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "INSTALLATION.md"
  - "guides/configuration.md"
  - "guides/routing.md"
---

# Getting Started

Welcome to the Rockstar Web Framework! This tutorial will guide you through building your first web application. By the end, you'll understand the core concepts and be ready to build production-ready applications.

## Prerequisites

Before starting, make sure you have:

- Go 1.25 or higher installed ([Installation Guide](INSTALLATION.md))
- Basic familiarity with Go programming
- A text editor or IDE

## Your First Application: Hello World

Let's build a minimal web application that responds to HTTP requests.

### Step 1: Create a New Project

```bash
mkdir hello-rockstar
cd hello-rockstar
go mod init hello-rockstar
```

### Step 2: Install Rockstar

```bash
go get github.com/echterhof/rockstar-web-framework/pkg
```

### Step 3: Create Your Application

Create a file named `main.go`:

```go
package main

import (
    "log"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Create a minimal configuration
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1: true,  // Enable HTTP/1.1 support
        },
    }

    // Initialize the framework
    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create framework: %v", err)
    }

    // Get the router
    router := app.Router()

    // Define a route handler
    router.GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]interface{}{
            "message": "Hello, Rockstar! 🎸",
        })
    })

    // Start the server
    log.Println("Server starting on http://localhost:8080")
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Step 4: Run Your Application

```bash
go run main.go
```

You should see:
```
Server starting on http://localhost:8080
```

### Step 5: Test Your Application

Open a new terminal and test your endpoint:

```bash
curl http://localhost:8080/
```

You should see:
```json
{"message":"Hello, Rockstar! 🎸"}
```

Congratulations! You've built your first Rockstar application! 🎉

## Understanding the Core Concepts

Now let's break down what's happening in your application and explore the key concepts.

### 1. Framework Configuration

The `FrameworkConfig` struct controls how your application behaves:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,  // Enable HTTP/1.1
        EnableHTTP2: true,  // Enable HTTP/2 (optional)
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "sqlite",
        Database: "myapp.db",
    },
    // ... more configuration options
}
```

**Key configuration areas:**
- **ServerConfig**: HTTP protocols, timeouts, connection limits
- **DatabaseConfig**: Database connection settings
- **CacheConfig**: Caching behavior
- **SessionConfig**: Session management and security

Most settings have sensible defaults, so you only need to specify what you want to change.

[Learn more about configuration →](guides/configuration.md)

### 2. The Framework Instance

The `Framework` is the central object that manages your application:

```go
app, err := pkg.New(config)
```

The framework provides access to:
- **Router**: Route registration and HTTP handling
- **Database**: Database operations
- **Cache**: Caching operations
- **Security**: Authentication and authorization
- **Sessions**: Session management
- **Plugins**: Plugin system

### 3. The Router

The `Router` handles HTTP routing and maps URLs to handler functions:

```go
router := app.Router()

// Register routes for different HTTP methods
router.GET("/users", listUsers)           // GET request
router.POST("/users", createUser)         // POST request
router.PUT("/users/:id", updateUser)      // PUT request with parameter
router.DELETE("/users/:id", deleteUser)   // DELETE request
```

**Supported HTTP methods:**
- `GET` - Retrieve resources
- `POST` - Create resources
- `PUT` - Update resources
- `DELETE` - Delete resources
- `PATCH` - Partial updates
- `HEAD` - Headers only
- `OPTIONS` - Supported methods

[Learn more about routing →](guides/routing.md)

### 4. The Context

The `Context` is passed to every handler and provides access to the request, response, and framework services:

```go
func myHandler(ctx pkg.Context) error {
    // Access request information
    method := ctx.Request().Method
    path := ctx.Request().URL.Path
    
    // Get URL parameters
    userID := ctx.Params()["id"]
    
    // Get query parameters
    page := ctx.Query("page")
    
    // Access framework services
    db := ctx.DB()           // Database
    cache := ctx.Cache()     // Cache
    session := ctx.Session() // Session
    
    // Send responses
    return ctx.JSON(200, map[string]interface{}{
        "status": "success",
    })
}
```

**Context provides:**
- Request data (headers, body, parameters)
- Response methods (JSON, HTML, file, etc.)
- Framework services (database, cache, session)
- Request-scoped values

[Learn more about Context →](guides/context.md)

### 5. Handler Functions

Handlers are functions that process HTTP requests:

```go
func myHandler(ctx pkg.Context) error {
    // Process the request
    // Return a response or an error
    return ctx.JSON(200, data)
}
```

**Handler signature:**
```go
type HandlerFunc func(Context) error
```

Handlers receive a `Context` and return an `error`. If an error is returned, the framework's error handler processes it.

## Building a More Complete Application

Let's expand our Hello World into a simple user API with multiple routes and features.

### Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Configuration
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            EnableHTTP1: true,
            EnableHTTP2: true,
        },
    }

    // Initialize framework
    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create framework: %v", err)
    }

    // Add middleware for logging
    app.Use(loggingMiddleware)

    // Get router and register routes
    router := app.Router()
    
    // Home endpoint
    router.GET("/", homeHandler)
    
    // User API endpoints
    router.GET("/users", listUsersHandler)
    router.GET("/users/:id", getUserHandler)
    router.POST("/users", createUserHandler)
    router.PUT("/users/:id", updateUserHandler)
    router.DELETE("/users/:id", deleteUserHandler)

    // Start server
    log.Println("🎸 Server starting on http://localhost:8080")
    log.Println("Try: curl http://localhost:8080/")
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

// Middleware: Logs all requests
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Log incoming request
    log.Printf("[%s] %s %s",
        ctx.Request().Method,
        ctx.Request().URL.Path,
        ctx.Request().RemoteAddr,
    )
    
    // Call next handler
    err := next(ctx)
    
    // Log completion time
    duration := time.Since(start)
    log.Printf("  Completed in %v", duration)
    
    return err
}

// Handler: Home page
func homeHandler(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]interface{}{
        "message": "Welcome to the User API! 🎸",
        "version": "1.0.0",
        "endpoints": []string{
            "GET /users - List all users",
            "GET /users/:id - Get a specific user",
            "POST /users - Create a new user",
            "PUT /users/:id - Update a user",
            "DELETE /users/:id - Delete a user",
        },
    })
}

// Handler: List all users
func listUsersHandler(ctx pkg.Context) error {
    // In a real app, fetch from database using ctx.DB()
    users := []map[string]interface{}{
        {"id": "1", "name": "Alice", "email": "alice@example.com"},
        {"id": "2", "name": "Bob", "email": "bob@example.com"},
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "users": users,
        "count": len(users),
    })
}

// Handler: Get a specific user
func getUserHandler(ctx pkg.Context) error {
    // Extract URL parameter
    userID := ctx.Params()["id"]
    
    // In a real app, query database using ctx.DB()
    user := map[string]interface{}{
        "id":    userID,
        "name":  "Alice",
        "email": "alice@example.com",
    }
    
    return ctx.JSON(200, user)
}

// Handler: Create a new user
func createUserHandler(ctx pkg.Context) error {
    // In a real app:
    // 1. Parse request body: ctx.BindJSON(&user)
    // 2. Validate input
    // 3. Save to database: ctx.DB().Exec(...)
    // 4. Return created resource
    
    return ctx.JSON(201, map[string]interface{}{
        "message": "User created successfully",
        "id":      "3",
    })
}

// Handler: Update a user
func updateUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // In a real app:
    // 1. Parse request body
    // 2. Validate input
    // 3. Update in database
    // 4. Return updated resource
    
    return ctx.JSON(200, map[string]interface{}{
        "message": fmt.Sprintf("User %s updated successfully", userID),
    })
}

// Handler: Delete a user
func deleteUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // In a real app:
    // 1. Check if user exists
    // 2. Delete from database
    // 3. Return confirmation
    
    return ctx.JSON(200, map[string]interface{}{
        "message": fmt.Sprintf("User %s deleted successfully", userID),
    })
}
```

### Testing the Complete Application

```bash
# List all users
curl http://localhost:8080/users

# Get a specific user
curl http://localhost:8080/users/1

# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'

# Update a user
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.new@example.com"}'

# Delete a user
curl -X DELETE http://localhost:8080/users/1
```

## Key Concepts Explained

### Middleware

Middleware functions run before your handlers and can:
- Log requests
- Authenticate users
- Validate input
- Handle errors
- Modify requests/responses

```go
func myMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Code before handler
    log.Println("Before handler")
    
    // Call the next handler
    err := next(ctx)
    
    // Code after handler
    log.Println("After handler")
    
    return err
}

// Register middleware globally
app.Use(myMiddleware)
```

[Learn more about middleware →](guides/middleware.md)

### URL Parameters

Extract dynamic values from URLs:

```go
// Route with parameter
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Params()["id"]  // Extract :id parameter
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
    })
})
```

### Query Parameters

Access URL query strings:

```go
// URL: /search?q=golang&page=2
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query("q")      // "golang"
    page := ctx.Query("page")    // "2"
    
    return ctx.JSON(200, map[string]interface{}{
        "query": query,
        "page":  page,
    })
})
```

### Request Body Parsing

Parse JSON request bodies:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

router.POST("/users", func(ctx pkg.Context) error {
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid JSON",
        })
    }
    
    // Use the parsed user data
    return ctx.JSON(201, user)
})
```

### Response Types

Send different response types:

```go
// JSON response
ctx.JSON(200, data)

// HTML response
ctx.HTML(200, "<h1>Hello</h1>")

// Plain text
ctx.String(200, "Hello, World!")

// File download
ctx.File("/path/to/file.pdf")

// Redirect
ctx.Redirect(302, "/new-location")
```

## Common Patterns

### Error Handling

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    
    // Query database
    user, err := getUserFromDB(ctx.DB(), userID)
    if err != nil {
        // Return error - framework's error handler will process it
        return err
    }
    
    return ctx.JSON(200, user)
})

// Set custom error handler
app.SetErrorHandler(func(ctx pkg.Context, err error) error {
    log.Printf("Error: %v", err)
    return ctx.JSON(500, map[string]interface{}{
        "error": err.Error(),
    })
})
```

### Route Groups

Organize related routes:

```go
// API v1 routes
v1 := router.Group("/api/v1")
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

// API v2 routes
v2 := router.Group("/api/v2")
v2.GET("/users", listUsersV2)
v2.POST("/users", createUserV2)
```

### Lifecycle Hooks

Run code during application lifecycle:

```go
// Startup hook - runs when server starts
app.RegisterStartupHook(func(ctx context.Context) error {
    log.Println("Server starting...")
    // Initialize resources
    return nil
})

// Shutdown hook - runs during graceful shutdown
app.RegisterShutdownHook(func(ctx context.Context) error {
    log.Println("Server shutting down...")
    // Cleanup resources
    return nil
})
```

## Next Steps

Now that you understand the basics, explore these topics to build more advanced applications:

### Core Features
- [Configuration Guide](guides/configuration.md) - Learn about all configuration options
- [Routing Guide](guides/routing.md) - Advanced routing patterns and techniques
- [Middleware Guide](guides/middleware.md) - Create and use middleware effectively
- [Context Guide](guides/context.md) - Master the Context interface

### Data & Storage
- [Database Guide](guides/database.md) - Work with databases (PostgreSQL, MySQL, SQLite, MSSQL)
- [Caching Guide](guides/caching.md) - Implement caching strategies
- [Sessions Guide](guides/sessions.md) - Manage user sessions

### Security
- [Security Guide](guides/security.md) - Authentication, authorization, and security best practices

### Advanced Features
- [Multi-Tenancy Guide](guides/multi-tenancy.md) - Build multi-tenant applications
- [Internationalization Guide](guides/i18n.md) - Support multiple languages
- [Plugin System Guide](guides/plugins.md) - Extend the framework with plugins
- [WebSockets Guide](guides/websockets.md) - Real-time communication

### Protocols & APIs
- [Protocols Guide](guides/protocols.md) - HTTP/1, HTTP/2, and QUIC
- [API Styles Guide](guides/api-styles.md) - REST, GraphQL, gRPC, and SOAP

### Operations
- [Monitoring Guide](guides/monitoring.md) - Metrics and observability
- [Performance Guide](guides/performance.md) - Optimize your application
- [Deployment Guide](guides/deployment.md) - Deploy to production

### Examples
- [Complete Examples](examples/README.md) - Explore full example applications
- [REST API Example](examples/rest-api.md) - Build a complete REST API
- [Full-Featured App](examples/full-featured-app.md) - See all features in action

## Getting Help

- **Documentation**: Browse the [complete documentation](README.md)
- **Examples**: Check out the [examples directory](../examples/)
- **Troubleshooting**: See [common errors and solutions](troubleshooting/common-errors.md)
- **FAQ**: Read the [frequently asked questions](troubleshooting/faq.md)

## Navigation

- [← Back to Installation](INSTALLATION.md)
- [Next: Configuration Guide →](guides/configuration.md)
