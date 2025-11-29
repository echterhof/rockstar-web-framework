---
title: "Router API"
description: "RouterEngine interface for HTTP routing and request handling"
category: "api"
tags: ["api", "router", "routing", "handlers", "middleware"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "framework.md"
  - "context.md"
  - "../guides/routing.md"
  - "../guides/middleware.md"
---

# Router API

## Overview

The `RouterEngine` interface provides the routing system for the Rockstar Web Framework. It handles HTTP request routing, path parameter extraction, route groups, middleware management, and multi-protocol support including WebSocket, GraphQL, gRPC, and SOAP.

The router uses a fluent API design, allowing method chaining for clean and readable route definitions. It supports path parameters, wildcards, host-based routing for multi-tenancy, and static file serving.

**Primary Use Cases:**
- Defining HTTP routes for all standard methods (GET, POST, PUT, DELETE, etc.)
- Creating route groups with shared prefixes and middleware
- Extracting path parameters and wildcards
- Host-based routing for multi-tenant applications
- Serving static files and directories
- WebSocket endpoint registration
- GraphQL, gRPC, and SOAP API integration

## Type Definition

### RouterEngine Interface

```go
type RouterEngine interface {
    // HTTP method routing
    GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
    OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine

    // Generic method routing
    Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine

    // Route groups
    Group(prefix string, middleware ...MiddlewareFunc) RouterEngine

    // Host-specific routing for multi-tenancy
    Host(hostname string) RouterEngine

    // Static file serving
    Static(prefix string, filesystem VirtualFS) RouterEngine
    StaticFile(path, filepath string) RouterEngine

    // WebSocket routing
    WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine

    // API protocol routing
    GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
    GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
    SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine

    // Middleware management
    Use(middleware ...MiddlewareFunc) RouterEngine

    // Route matching
    Match(method, path, host string) (*Route, map[string]string, bool)

    // Route information
    Routes() []*Route
}
```

### Supporting Types

```go
// HandlerFunc represents a request handler function
type HandlerFunc func(ctx Context) error

// MiddlewareFunc represents a middleware function
type MiddlewareFunc func(ctx Context, next HandlerFunc) error

// WebSocketHandler represents a WebSocket handler function
type WebSocketHandler func(ctx Context, conn WebSocketConnection) error

// Route represents a registered route
type Route struct {
    Method      string
    Path        string
    Handler     HandlerFunc
    Middleware  []MiddlewareFunc
    Host        string
    Name        string
    IsWebSocket bool
    IsStatic    bool

    // WebSocket-specific fields
    WebSocketHandler WebSocketHandler

    // API-specific fields
    GraphQLSchema GraphQLSchema
    GRPCService   GRPCService
    SOAPService   SOAPService
}
```

## HTTP Method Routing

### GET

```go
func GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP GET requests.

**Parameters**:
- `path` (string): URL path pattern (supports parameters and wildcards)
- `handler` (HandlerFunc): Handler function to execute for matching requests
- `middleware` (...MiddlewareFunc): Optional middleware functions to apply to this route

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
router := app.Router()

// Simple route
router.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome!")
})

// Route with path parameter
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"user_id": userID})
})

// Route with middleware
router.GET("/admin", adminHandler, authMiddleware, adminMiddleware)

// Route with wildcard
router.GET("/files/*filepath", func(ctx pkg.Context) error {
    filepath := ctx.Param("filepath")
    return ctx.String(200, "File: "+filepath)
})
```

**See Also**:
- [POST](#post)
- [Handle](#handle)
- [Routing Guide](../guides/routing.md)

### POST

```go
func POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP POST requests.

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Create user
router.POST("/users", func(ctx pkg.Context) error {
    var user User
    if err := json.Unmarshal(ctx.Body(), &user); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    // Save user to database
    db := ctx.DB()
    result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        user.Name, user.Email)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Database error"})
    }
    
    return ctx.JSON(201, map[string]string{"message": "User created"})
})

// Login endpoint
router.POST("/auth/login", loginHandler, rateLimitMiddleware)
```

### PUT

```go
func PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP PUT requests for updating resources.

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Update user
router.PUT("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    var updates User
    if err := json.Unmarshal(ctx.Body(), &updates); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    db := ctx.DB()
    _, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?",
        updates.Name, updates.Email, userID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Update failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "User updated"})
})
```

### DELETE

```go
func DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP DELETE requests for removing resources.

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Delete user
router.DELETE("/users/:id", func(ctx pkg.Context) error {
    if !ctx.IsAuthorized("users", "delete") {
        return ctx.JSON(403, map[string]string{"error": "Forbidden"})
    }
    
    userID := ctx.Param("id")
    db := ctx.DB()
    
    _, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Delete failed"})
    }
    
    return ctx.JSON(200, map[string]string{"message": "User deleted"})
}, authMiddleware)
```

### PATCH

```go
func PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP PATCH requests for partial resource updates.

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Partial update
router.PATCH("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    var updates map[string]interface{}
    if err := json.Unmarshal(ctx.Body(), &updates); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    // Build dynamic UPDATE query based on provided fields
    // ... implementation ...
    
    return ctx.JSON(200, map[string]string{"message": "User updated"})
})
```

### HEAD

```go
func HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP HEAD requests (like GET but without response body).

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Check if resource exists
router.HEAD("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    db := ctx.DB()
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
    
    if err != nil || !exists {
        return ctx.String(404, "")
    }
    
    return ctx.String(200, "")
})
```

### OPTIONS

```go
func OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route that handles HTTP OPTIONS requests for CORS preflight.

**Parameters**:
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// CORS preflight
router.OPTIONS("/api/*", func(ctx pkg.Context) error {
    ctx.SetHeader("Access-Control-Allow-Origin", "*")
    ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
    return ctx.String(204, "")
})
```

### Handle

```go
func Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a route for any HTTP method. This is the generic routing method used by all HTTP method-specific functions.

**Parameters**:
- `method` (string): HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, or "*" for all)
- `path` (string): URL path pattern
- `handler` (HandlerFunc): Handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Custom method
router.Handle("PROPFIND", "/webdav/*", webdavHandler)

// Match all methods
router.Handle("*", "/catch-all", func(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{
        "method": ctx.Request().Method,
        "path": ctx.Request().RequestURI,
    })
})

// Multiple routes with same handler
for _, method := range []string{"GET", "POST", "PUT"} {
    router.Handle(method, "/api/resource", resourceHandler)
}
```

## Route Groups

### Group

```go
func Group(prefix string, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Creates a route group with a common path prefix and shared middleware. Groups can be nested for hierarchical routing.

**Parameters**:
- `prefix` (string): Path prefix for all routes in the group
- `middleware` (...MiddlewareFunc): Middleware to apply to all routes in the group

**Returns**:
- `RouterEngine`: New router instance for the group

**Example**:
```go
router := app.Router()

// API v1 group
v1 := router.Group("/api/v1")
v1.GET("/users", listUsersHandler)
v1.POST("/users", createUserHandler)
v1.GET("/users/:id", getUserHandler)

// API v2 group with middleware
v2 := router.Group("/api/v2", apiKeyMiddleware, rateLimitMiddleware)
v2.GET("/users", listUsersV2Handler)
v2.POST("/users", createUserV2Handler)

// Admin group with authentication
admin := router.Group("/admin", authMiddleware, adminMiddleware)
admin.GET("/dashboard", dashboardHandler)
admin.GET("/users", adminUsersHandler)
admin.POST("/users/:id/ban", banUserHandler)

// Nested groups
api := router.Group("/api")
{
    v1 := api.Group("/v1")
    {
        users := v1.Group("/users")
        users.GET("", listUsersHandler)
        users.POST("", createUserHandler)
        users.GET("/:id", getUserHandler)
        users.PUT("/:id", updateUserHandler)
        users.DELETE("/:id", deleteUserHandler)
    }
}
```

**Note**: Middleware from parent groups is inherited by child groups and routes.

**See Also**:
- [Use](#use)
- [Middleware Guide](../guides/middleware.md)

## Multi-Tenancy Support

### Host

```go
func Host(hostname string) RouterEngine
```

**Description**: Creates a host-specific router for multi-tenant applications. Routes registered on this router will only match requests to the specified hostname.

**Parameters**:
- `hostname` (string): Hostname to match (e.g., "tenant1.example.com")

**Returns**:
- `RouterEngine`: New router instance for the specified host

**Example**:
```go
router := app.Router()

// Default routes (no host restriction)
router.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Main site")
})

// Tenant-specific routes
tenant1 := router.Host("tenant1.example.com")
tenant1.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Tenant 1")
})
tenant1.GET("/dashboard", tenant1DashboardHandler)

tenant2 := router.Host("tenant2.example.com")
tenant2.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Tenant 2")
})
tenant2.GET("/dashboard", tenant2DashboardHandler)

// Admin subdomain
admin := router.Host("admin.example.com")
admin.GET("/", adminHomeHandler, authMiddleware)
admin.GET("/users", adminUsersHandler, authMiddleware)
```

**See Also**:
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)

## Static File Serving

### Static

```go
func Static(prefix string, filesystem VirtualFS) RouterEngine
```

**Description**: Registers a route to serve static files from a virtual filesystem.

**Parameters**:
- `prefix` (string): URL path prefix for static files
- `filesystem` (VirtualFS): Virtual filesystem containing the files

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Serve files from ./public directory
fs := pkg.NewOSFileSystem("./public")
router.Static("/static", fs)
// Serves: /static/css/style.css -> ./public/css/style.css
//         /static/js/app.js -> ./public/js/app.js

// Serve files from embedded filesystem
//go:embed assets/*
var assetsFS embed.FS
router.Static("/assets", pkg.NewEmbedFileSystem(assetsFS, "assets"))

// Multiple static directories
router.Static("/images", pkg.NewOSFileSystem("./images"))
router.Static("/downloads", pkg.NewOSFileSystem("./downloads"))
```

**See Also**:
- [StaticFile](#staticfile)

### StaticFile

```go
func StaticFile(path, filepath string) RouterEngine
```

**Description**: Registers a route to serve a single static file.

**Parameters**:
- `path` (string): URL path for the route
- `filepath` (string): Filesystem path to the file

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Serve favicon
router.StaticFile("/favicon.ico", "./public/favicon.ico")

// Serve robots.txt
router.StaticFile("/robots.txt", "./public/robots.txt")

// Serve specific files
router.StaticFile("/sitemap.xml", "./public/sitemap.xml")
router.StaticFile("/.well-known/security.txt", "./public/security.txt")
```

## WebSocket Support

### WebSocket

```go
func WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a WebSocket endpoint. The handler receives a WebSocket connection for bidirectional communication.

**Parameters**:
- `path` (string): URL path for the WebSocket endpoint
- `handler` (WebSocketHandler): WebSocket handler function
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Chat WebSocket
router.WebSocket("/ws/chat", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    // Send welcome message
    conn.WriteMessage(websocket.TextMessage, []byte("Welcome to chat!"))
    
    // Read messages
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Echo message back
        if err := conn.WriteMessage(messageType, data); err != nil {
            return err
        }
    }
})

// Real-time notifications
router.WebSocket("/ws/notifications", notificationHandler, authMiddleware)

// Live data stream
router.WebSocket("/ws/metrics", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            metrics := collectMetrics()
            data, _ := json.Marshal(metrics)
            if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
                return err
            }
        case <-ctx.Context().Done():
            return nil
        }
    }
})
```

**See Also**:
- [WebSocket Guide](../guides/websockets.md)

## API Protocol Support

### GraphQL

```go
func GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a GraphQL endpoint with the specified schema.

**Parameters**:
- `path` (string): URL path for the GraphQL endpoint
- `schema` (GraphQLSchema): GraphQL schema implementation
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Create GraphQL schema
schema := createGraphQLSchema()

// Register GraphQL endpoint
router.GraphQL("/graphql", schema)

// GraphQL with authentication
router.GraphQL("/api/graphql", schema, authMiddleware)

// Multiple GraphQL endpoints
router.GraphQL("/graphql/public", publicSchema)
router.GraphQL("/graphql/admin", adminSchema, authMiddleware, adminMiddleware)
```

**See Also**:
- [API Styles Guide](../guides/api-styles.md#graphql)

### GRPC

```go
func GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a gRPC service. The service is automatically mounted at `/grpc/{ServiceName}`.

**Parameters**:
- `service` (GRPCService): gRPC service implementation
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Register gRPC service
userService := NewUserGRPCService()
router.GRPC(userService)

// gRPC with middleware
productService := NewProductGRPCService()
router.GRPC(productService, loggingMiddleware, metricsMiddleware)

// Multiple gRPC services
router.GRPC(userService)
router.GRPC(productService)
router.GRPC(orderService)
```

**See Also**:
- [API Styles Guide](../guides/api-styles.md#grpc)

### SOAP

```go
func SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Registers a SOAP service endpoint with WSDL support.

**Parameters**:
- `path` (string): URL path for the SOAP endpoint
- `service` (SOAPService): SOAP service implementation
- `middleware` (...MiddlewareFunc): Optional middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
// Register SOAP service
soapService := NewUserSOAPService()
router.SOAP("/soap/users", soapService)

// SOAP with authentication
router.SOAP("/soap/admin", adminSOAPService, authMiddleware)

// WSDL is automatically available at {path}?wsdl
// Example: /soap/users?wsdl
```

**See Also**:
- [API Styles Guide](../guides/api-styles.md#soap)

## Middleware Management

### Use

```go
func Use(middleware ...MiddlewareFunc) RouterEngine
```

**Description**: Adds middleware to the router. Middleware added with Use applies to all routes registered after the Use call.

**Parameters**:
- `middleware` (...MiddlewareFunc): One or more middleware functions

**Returns**:
- `RouterEngine`: The router instance for method chaining

**Example**:
```go
router := app.Router()

// Global middleware (applies to all routes)
router.Use(loggingMiddleware)
router.Use(recoveryMiddleware)
router.Use(corsMiddleware)

// Routes registered after Use will have the middleware
router.GET("/", homeHandler)
router.GET("/about", aboutHandler)

// Add more middleware
router.Use(authMiddleware)

// These routes have all middleware including auth
router.GET("/profile", profileHandler)
router.GET("/settings", settingsHandler)

// Group with additional middleware
admin := router.Group("/admin")
admin.Use(adminMiddleware)
admin.GET("/dashboard", dashboardHandler)
```

**Middleware Execution Order**:
1. Global middleware (from Use)
2. Group middleware (from Group)
3. Route-specific middleware (from route registration)
4. Handler

**See Also**:
- [Middleware Guide](../guides/middleware.md)

## Route Matching and Inspection

### Match

```go
func Match(method, path, host string) (*Route, map[string]string, bool)
```

**Description**: Finds a route that matches the given method, path, and host. Returns the route, extracted parameters, and whether a match was found.

**Parameters**:
- `method` (string): HTTP method
- `path` (string): Request path
- `host` (string): Request hostname (optional)

**Returns**:
- `*Route`: Matched route, or nil if no match
- `map[string]string`: Extracted path parameters
- `bool`: true if a route was matched, false otherwise

**Example**:
```go
router := app.Router()
router.GET("/users/:id", getUserHandler)

// Match a route
route, params, found := router.Match("GET", "/users/123", "")
if found {
    fmt.Printf("Route: %s\n", route.Path)        // "/users/:id"
    fmt.Printf("User ID: %s\n", params["id"])    // "123"
}

// No match
route, params, found = router.Match("POST", "/users/123", "")
// found = false
```

**Note**: This method is primarily used internally by the framework but can be useful for testing or custom routing logic.

### Routes

```go
func Routes() []*Route
```

**Description**: Returns all registered routes including routes from host-specific routers.

**Returns**:
- `[]*Route`: Slice of all registered routes

**Example**:
```go
router := app.Router()
router.GET("/", homeHandler)
router.POST("/users", createUserHandler)
router.GET("/users/:id", getUserHandler)

// Get all routes
routes := router.Routes()
for _, route := range routes {
    fmt.Printf("%s %s\n", route.Method, route.Path)
}
// Output:
// GET /
// POST /users
// GET /users/:id

// Useful for debugging or documentation generation
func printRoutes(router pkg.RouterEngine) {
    routes := router.Routes()
    fmt.Println("Registered Routes:")
    fmt.Println("==================")
    for _, route := range routes {
        middleware := len(route.Middleware)
        fmt.Printf("%-7s %-30s (%d middleware)\n", 
            route.Method, route.Path, middleware)
    }
}
```

## Path Parameter Patterns

The router supports several path parameter patterns:

### Named Parameters

```go
// Single parameter
router.GET("/users/:id", handler)
// Matches: /users/123
// Params: {"id": "123"}

// Multiple parameters
router.GET("/users/:userId/posts/:postId", handler)
// Matches: /users/123/posts/456
// Params: {"userId": "123", "postId": "456"}
```

### Wildcard Parameters

```go
// Wildcard (captures remaining path)
router.GET("/files/*filepath", handler)
// Matches: /files/docs/readme.md
// Params: {"filepath": "docs/readme.md"}

// Named wildcard
router.GET("/static/*path", handler)
// Matches: /static/css/style.css
// Params: {"path": "css/style.css"}
```

### Regex Parameters

```go
// Regex pattern
router.GET("/users/:id([0-9]+)", handler)
// Matches: /users/123
// Does not match: /users/abc

// Multiple regex parameters
router.GET("/posts/:year([0-9]{4})/:month([0-9]{2})", handler)
// Matches: /posts/2024/11
// Does not match: /posts/24/11
```

## Complete Example

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    app, _ := pkg.New(nil)
    router := app.Router()

    // Global middleware
    router.Use(loggingMiddleware, recoveryMiddleware)

    // Home page
    router.GET("/", homeHandler)

    // API v1
    v1 := router.Group("/api/v1")
    {
        // Public endpoints
        v1.POST("/auth/login", loginHandler)
        v1.POST("/auth/register", registerHandler)

        // Protected endpoints
        users := v1.Group("/users", authMiddleware)
        users.GET("", listUsersHandler)
        users.POST("", createUserHandler)
        users.GET("/:id", getUserHandler)
        users.PUT("/:id", updateUserHandler)
        users.DELETE("/:id", deleteUserHandler)

        // Posts
        posts := v1.Group("/posts", authMiddleware)
        posts.GET("", listPostsHandler)
        posts.POST("", createPostHandler)
        posts.GET("/:id", getPostHandler)
        posts.PUT("/:id", updatePostHandler)
        posts.DELETE("/:id", deletePostHandler)
    }

    // Admin panel
    admin := router.Group("/admin", authMiddleware, adminMiddleware)
    {
        admin.GET("/dashboard", dashboardHandler)
        admin.GET("/users", adminUsersHandler)
        admin.POST("/users/:id/ban", banUserHandler)
    }

    // Static files
    router.Static("/static", pkg.NewOSFileSystem("./public"))
    router.StaticFile("/favicon.ico", "./public/favicon.ico")

    // WebSocket
    router.WebSocket("/ws/chat", chatHandler)

    // GraphQL
    schema := createGraphQLSchema()
    router.GraphQL("/graphql", schema, authMiddleware)

    // Start server
    app.Start()
}
```

## Best Practices

1. **Use Route Groups**: Organize related routes into groups with shared prefixes and middleware
2. **Order Matters**: Register more specific routes before generic ones
3. **Middleware Placement**: Apply middleware at the appropriate level (global, group, or route)
4. **Parameter Validation**: Validate path parameters in handlers
5. **RESTful Design**: Follow REST conventions for resource-based APIs
6. **Error Handling**: Return appropriate HTTP status codes
7. **Documentation**: Document your routes for API consumers

## See Also

- [Context API](context.md) - Request handling and framework access
- [Framework API](framework.md) - Framework initialization and configuration
- [Routing Guide](../guides/routing.md) - Comprehensive routing guide
- [Middleware Guide](../guides/middleware.md) - Middleware patterns and examples
- [API Styles Guide](../guides/api-styles.md) - REST, GraphQL, gRPC, SOAP
