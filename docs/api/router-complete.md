# Router API Reference

## Overview

The Router is the core routing engine of the Rockstar Web Framework. It maps HTTP requests to handlers based on method, path, and host. Features include route parameters, middleware, route groups, multi-tenancy, static file serving, WebSocket support, and multi-protocol APIs (GraphQL, gRPC, SOAP).

## RouterEngine Interface

Main interface for routing operations.

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
    
    // Host-specific routing
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

## Handler Types

### HandlerFunc

Function signature for request handlers.

```go
type HandlerFunc func(ctx Context) error
```

**Example:**
```go
handler := func(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{"status": "ok"})
}
```

### MiddlewareFunc

Function signature for middleware.

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

**Example:**
```go
middleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
    log.Println("Before handler")
    err := next(ctx)
    log.Println("After handler")
    return err
}
```

### WebSocketHandler

Function signature for WebSocket handlers.

```go
type WebSocketHandler func(ctx Context, conn WebSocketConnection) error
```

## Route Type

Represents a registered route.

```go
type Route struct {
    Method      string
    Path        string
    Handler     HandlerFunc
    Middleware  []MiddlewareFunc
    Host        string
    Name        string
    IsWebSocket bool
    IsStatic    bool
    
    // Protocol-specific fields
    WebSocketHandler WebSocketHandler
    GraphQLSchema    GraphQLSchema
    GRPCService      GRPCService
    SOAPService      SOAPService
}
```

## Factory Function

### NewRouter()

Creates a new router instance.

**Signature:**
```go
func NewRouter() RouterEngine
```

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
router := pkg.NewRouter()
```

## HTTP Method Routing

### GET()

Registers a GET route.

**Signature:**
```go
GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `path` - Route path (supports parameters and wildcards)
- `handler` - Handler function
- `middleware` - Optional middleware functions

**Returns:**
- `RouterEngine` - Router instance (for chaining)

**Example:**
```go
router.GET("/users", func(ctx pkg.Context) error {
    return ctx.JSON(200, []string{"user1", "user2"})
})

// With middleware
router.GET("/admin", adminHandler, authMiddleware, logMiddleware)
```

### POST()

Registers a POST route.

**Signature:**
```go
POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.POST("/users", func(ctx pkg.Context) error {
    // Create user
    return ctx.JSON(201, map[string]string{"status": "created"})
})
```

### PUT()

Registers a PUT route.

**Signature:**
```go
PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.PUT("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    // Update user
    return ctx.JSON(200, map[string]string{"id": userID})
})
```

### DELETE()

Registers a DELETE route.

**Signature:**
```go
DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.DELETE("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    // Delete user
    return ctx.JSON(204, nil)
})
```

### PATCH()

Registers a PATCH route.

**Signature:**
```go
PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.PATCH("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    // Partially update user
    return ctx.JSON(200, map[string]string{"id": userID})
})
```

### HEAD()

Registers a HEAD route.

**Signature:**
```go
HEAD(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.HEAD("/users/:id", func(ctx pkg.Context) error {
    // Return headers only
    ctx.SetHeader("X-User-Exists", "true")
    return nil
})
```

### OPTIONS()

Registers an OPTIONS route.

**Signature:**
```go
OPTIONS(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.OPTIONS("/api/*", func(ctx pkg.Context) error {
    ctx.SetHeader("Allow", "GET, POST, PUT, DELETE")
    return ctx.String(200, "")
})
```

### Handle()

Registers a route with any HTTP method.

**Signature:**
```go
Handle(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `method` - HTTP method (GET, POST, etc.)
- `path` - Route path
- `handler` - Handler function
- `middleware` - Optional middleware functions

**Example:**
```go
router.Handle("GET", "/users", usersHandler)
router.Handle("CUSTOM", "/special", specialHandler)
```

## Route Groups

### Group()

Creates a route group with a common prefix and middleware.

**Signature:**
```go
Group(prefix string, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `prefix` - Path prefix for all routes in the group
- `middleware` - Middleware applied to all routes in the group

**Returns:**
- `RouterEngine` - New router instance for the group

**Example:**
```go
// API v1 group
v1 := router.Group("/api/v1", authMiddleware)
v1.GET("/users", listUsers)
v1.POST("/users", createUser)

// API v2 group
v2 := router.Group("/api/v2", authMiddleware, rateLimitMiddleware)
v2.GET("/users", listUsersV2)
v2.POST("/users", createUserV2)

// Nested groups
admin := router.Group("/admin", authMiddleware)
adminUsers := admin.Group("/users", adminMiddleware)
adminUsers.GET("/", listAdminUsers)
adminUsers.DELETE("/:id", deleteUser)
```

## Multi-Tenancy

### Host()

Creates a host-specific router for multi-tenancy.

**Signature:**
```go
Host(hostname string) RouterEngine
```

**Parameters:**
- `hostname` - Host name for tenant-specific routing

**Returns:**
- `RouterEngine` - Host-specific router instance

**Example:**
```go
// Tenant 1 routes
tenant1 := router.Host("tenant1.example.com")
tenant1.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Tenant 1")
})

// Tenant 2 routes
tenant2 := router.Host("tenant2.example.com")
tenant2.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Tenant 2")
})

// Default routes (no host specified)
router.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Default")
})
```

## Static File Serving

### Static()

Serves static files from a virtual filesystem.

**Signature:**
```go
Static(prefix string, filesystem VirtualFS) RouterEngine
```

**Parameters:**
- `prefix` - URL prefix for static files
- `filesystem` - Virtual filesystem containing static files

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
//go:embed static/*
var staticFS embed.FS

router.Static("/static", staticFS)
// Serves files from /static/css/style.css, /static/js/app.js, etc.
```

### StaticFile()

Serves a single static file.

**Signature:**
```go
StaticFile(path, filepath string) RouterEngine
```

**Parameters:**
- `path` - URL path
- `filepath` - File path on disk

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
router.StaticFile("/favicon.ico", "./static/favicon.ico")
router.StaticFile("/robots.txt", "./static/robots.txt")
```

## WebSocket Routing

### WebSocket()

Registers a WebSocket endpoint.

**Signature:**
```go
WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `path` - WebSocket endpoint path
- `handler` - WebSocket handler function
- `middleware` - Optional middleware functions

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
router.WebSocket("/ws/chat", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Echo back
        conn.WriteMessage(messageType, data)
    }
})
```

## API Protocol Routing

### GraphQL()

Registers a GraphQL endpoint.

**Signature:**
```go
GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `path` - GraphQL endpoint path
- `schema` - GraphQL schema
- `middleware` - Optional middleware functions

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
schema := createGraphQLSchema()
router.GraphQL("/graphql", schema, authMiddleware)
```

### GRPC()

Registers a gRPC service.

**Signature:**
```go
GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `service` - gRPC service implementation
- `middleware` - Optional middleware functions

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
userService := NewUserGRPCService()
router.GRPC(userService, authMiddleware)
```

### SOAP()

Registers a SOAP service.

**Signature:**
```go
SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `path` - SOAP endpoint path
- `service` - SOAP service implementation
- `middleware` - Optional middleware functions

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
soapService := NewUserSOAPService()
router.SOAP("/soap/users", soapService, authMiddleware)
```

## Middleware Management

### Use()

Adds middleware to the router.

**Signature:**
```go
Use(middleware ...MiddlewareFunc) RouterEngine
```

**Parameters:**
- `middleware` - Middleware functions to add

**Returns:**
- `RouterEngine` - Router instance

**Example:**
```go
// Global middleware
router.Use(loggingMiddleware)
router.Use(recoveryMiddleware)
router.Use(corsMiddleware)

// Multiple middleware at once
router.Use(
    loggingMiddleware,
    recoveryMiddleware,
    corsMiddleware,
)
```

## Route Matching

### Match()

Finds a route that matches the given method, path, and host.

**Signature:**
```go
Match(method, path, host string) (*Route, map[string]string, bool)
```

**Parameters:**
- `method` - HTTP method
- `path` - Request path
- `host` - Request host

**Returns:**
- `*Route` - Matched route
- `map[string]string` - Extracted path parameters
- `bool` - true if route found, false otherwise

**Example:**
```go
route, params, found := router.Match("GET", "/users/123", "")
if found {
    userID := params["id"]  // "123"
    // Execute route handler
}
```

### Routes()

Returns all registered routes.

**Signature:**
```go
Routes() []*Route
```

**Returns:**
- `[]*Route` - Array of all routes

**Example:**
```go
routes := router.Routes()
for _, route := range routes {
    fmt.Printf("%s %s\n", route.Method, route.Path)
}
```

## Path Parameters

### Simple Parameters

Use `:name` syntax for path parameters.

**Example:**
```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": userID})
})

// Matches: /users/123, /users/abc
// Extracts: id="123", id="abc"
```

### Multiple Parameters

**Example:**
```go
router.GET("/users/:userID/posts/:postID", func(ctx pkg.Context) error {
    userID := ctx.Param("userID")
    postID := ctx.Param("postID")
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
        "post_id": postID,
    })
})

// Matches: /users/123/posts/456
// Extracts: userID="123", postID="456"
```

### Wildcard Parameters

Use `*name` syntax for wildcard parameters that match multiple segments.

**Example:**
```go
router.GET("/files/*filepath", func(ctx pkg.Context) error {
    filepath := ctx.Param("filepath")
    return ctx.String(200, "File: "+filepath)
})

// Matches: /files/docs/readme.md
// Extracts: filepath="docs/readme.md"
```

### Regex Parameters

Use `:name(regex)` syntax for regex-validated parameters.

**Example:**
```go
router.GET("/users/:id([0-9]+)", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": userID})
})

// Matches: /users/123
// Does NOT match: /users/abc
```

## Complete Examples

### Basic REST API

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    router := app.Router()
    
    // List users
    router.GET("/users", func(ctx pkg.Context) error {
        users := []string{"Alice", "Bob"}
        return ctx.JSON(200, users)
    })
    
    // Get user
    router.GET("/users/:id", func(ctx pkg.Context) error {
        userID := ctx.Param("id")
        return ctx.JSON(200, map[string]string{"id": userID})
    })
    
    // Create user
    router.POST("/users", func(ctx pkg.Context) error {
        return ctx.JSON(201, map[string]string{"status": "created"})
    })
    
    // Update user
    router.PUT("/users/:id", func(ctx pkg.Context) error {
        userID := ctx.Param("id")
        return ctx.JSON(200, map[string]string{"id": userID})
    })
    
    // Delete user
    router.DELETE("/users/:id", func(ctx pkg.Context) error {
        return ctx.JSON(204, nil)
    })
    
    app.Listen(":8080")
}
```

### Route Groups

```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    router := app.Router()
    
    // Public routes
    router.GET("/", homeHandler)
    router.GET("/about", aboutHandler)
    
    // API v1
    v1 := router.Group("/api/v1")
    v1.GET("/users", listUsersV1)
    v1.POST("/users", createUserV1)
    
    // API v2 with auth
    v2 := router.Group("/api/v2", authMiddleware)
    v2.GET("/users", listUsersV2)
    v2.POST("/users", createUserV2)
    
    // Admin routes
    admin := router.Group("/admin", authMiddleware, adminMiddleware)
    admin.GET("/dashboard", dashboardHandler)
    admin.GET("/users", adminUsersHandler)
    
    app.Listen(":8080")
}
```

### Multi-Tenant Application

```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    router := app.Router()
    
    // Tenant 1
    tenant1 := router.Host("tenant1.example.com")
    tenant1.GET("/", func(ctx pkg.Context) error {
        return ctx.String(200, "Tenant 1 Home")
    })
    tenant1.GET("/dashboard", tenant1DashboardHandler)
    
    // Tenant 2
    tenant2 := router.Host("tenant2.example.com")
    tenant2.GET("/", func(ctx pkg.Context) error {
        return ctx.String(200, "Tenant 2 Home")
    })
    tenant2.GET("/dashboard", tenant2DashboardHandler)
    
    // Default (no host match)
    router.GET("/", func(ctx pkg.Context) error {
        return ctx.String(200, "Default Home")
    })
    
    app.Listen(":8080")
}
```

### Static Files and WebSocket

```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    router := app.Router()
    
    // Static files
    //go:embed static/*
    var staticFS embed.FS
    router.Static("/static", staticFS)
    router.StaticFile("/favicon.ico", "./static/favicon.ico")
    
    // WebSocket
    router.WebSocket("/ws/chat", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }
            
            // Broadcast to all clients
            conn.WriteMessage(messageType, data)
        }
    })
    
    // Regular routes
    router.GET("/", homeHandler)
    
    app.Listen(":8080")
}
```

## Best Practices

1. **Use Method-Specific Functions:** Use GET(), POST(), etc. instead of Handle() for clarity
2. **Group Related Routes:** Use Group() to organize related routes
3. **Apply Middleware Strategically:** Apply middleware at the appropriate level (global, group, or route)
4. **Use Path Parameters:** Use path parameters instead of query parameters for resource identifiers
5. **RESTful Design:** Follow REST conventions for API design
6. **Validate Parameters:** Always validate path parameters before use
7. **Handle Errors:** Return appropriate error responses
8. **Use Host Routing:** Use Host() for multi-tenant applications
9. **Static File Optimization:** Use embedded filesystems for static files
10. **Document Routes:** Document all routes and their parameters

## Security Considerations

1. **Input Validation:** Validate all path parameters and query parameters
2. **Authentication:** Use middleware for authentication on protected routes
3. **Authorization:** Check permissions before executing handlers
4. **Rate Limiting:** Implement rate limiting on public endpoints
5. **CORS:** Configure CORS properly for API endpoints
6. **Path Traversal:** Validate file paths in static file serving
7. **Method Validation:** Only allow appropriate HTTP methods
8. **Host Validation:** Validate host headers in multi-tenant applications

## See Also

- [Context API](context.md) - Context interface
- [Middleware API](middleware-pipeline.md) - Middleware system
- [WebSocket API](websockets.md) - WebSocket support
- [Routing Guide](../guides/routing.md) - Routing patterns
- [Middleware Guide](../guides/middleware.md) - Middleware patterns
