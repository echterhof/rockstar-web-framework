# Routing

## Overview

The Rockstar Web Framework provides a powerful and flexible routing system that supports all standard HTTP methods, path parameters, query strings, route groups, and middleware. The router is designed for high performance while maintaining an intuitive API.

Routes map HTTP requests to handler functions based on the request method and URL path. The framework supports parameter extraction, wildcard matching, regex patterns, and host-based routing for multi-tenancy.

## Quick Start

Here's a simple example of defining routes:

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
    
    // Define routes
    router.GET("/", homeHandler)
    router.GET("/users/:id", getUserHandler)
    router.POST("/users", createUserHandler)
    
    app.Listen(":8080")
}

func homeHandler(ctx pkg.Context) error {
    return ctx.String(200, "Welcome to Rockstar!")
}

func getUserHandler(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    return ctx.JSON(200, map[string]string{"user_id": userID})
}

func createUserHandler(ctx pkg.Context) error {
    // Handle user creation
    return ctx.JSON(201, map[string]string{"status": "created"})
}
```

## HTTP Methods

The router supports all standard HTTP methods with dedicated methods for each:

### GET

Used for retrieving resources:

```go
router.GET("/products", func(ctx pkg.Context) error {
    return ctx.JSON(200, map[string]string{"message": "List of products"})
})

router.GET("/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    return ctx.JSON(200, map[string]string{"product_id": id})
})
```

### POST

Used for creating new resources:

```go
router.POST("/products", func(ctx pkg.Context) error {
    var product map[string]interface{}
    if err := ctx.BindJSON(&product); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    // Create product
    return ctx.JSON(201, map[string]string{"status": "created"})
})
```

### PUT

Used for updating entire resources:

```go
router.PUT("/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    var product map[string]interface{}
    if err := ctx.BindJSON(&product); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    // Update product
    return ctx.JSON(200, map[string]string{"id": id, "status": "updated"})
})
```

### PATCH

Used for partial updates:

```go
router.PATCH("/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    var updates map[string]interface{}
    if err := ctx.BindJSON(&updates); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    
    // Apply partial update
    return ctx.JSON(200, map[string]string{"id": id, "status": "patched"})
})
```

### DELETE

Used for deleting resources:

```go
router.DELETE("/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    // Delete product
    return ctx.JSON(200, map[string]string{"id": id, "status": "deleted"})
})
```

### HEAD

Used for retrieving headers without body:

```go
router.HEAD("/products/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    // Check if product exists
    // Set headers but don't send body
    ctx.Response().Header().Set("X-Product-Exists", "true")
    return ctx.NoContent(200)
})
```

### OPTIONS

Used for CORS preflight requests:

```go
router.OPTIONS("/api/*", func(ctx pkg.Context) error {
    ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    return ctx.NoContent(204)
})
```

### Generic Handle Method

For custom or multiple methods:

```go
// Single custom method
router.Handle("CUSTOM", "/endpoint", customHandler)

// Multiple methods for same route
router.Handle("GET", "/resource", getHandler)
router.Handle("POST", "/resource", postHandler)
router.Handle("PUT", "/resource", putHandler)
```

## Path Parameters

Path parameters allow you to capture values from the URL path:

### Basic Parameters

Parameters are defined with a colon prefix:

```go
// Single parameter
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Params()["id"]
    return ctx.JSON(200, map[string]string{"user_id": userID})
})

// Multiple parameters
router.GET("/users/:userID/posts/:postID", func(ctx pkg.Context) error {
    userID := ctx.Params()["userID"]
    postID := ctx.Params()["postID"]
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
        "post_id": postID,
    })
})
```

### Wildcard Parameters

Wildcards capture the rest of the path:

```go
// Capture everything after /files/
router.GET("/files/*filepath", func(ctx pkg.Context) error {
    filepath := ctx.Params()["filepath"]
    // filepath could be "documents/report.pdf" for /files/documents/report.pdf
    return ctx.String(200, "File: "+filepath)
})

// Named wildcard
router.GET("/static/*path", func(ctx pkg.Context) error {
    path := ctx.Params()["path"]
    return ctx.String(200, "Static file: "+path)
})
```

### Regex Parameters

Use regex patterns for parameter validation:

```go
// Match only numeric IDs
router.GET("/users/:id([0-9]+)", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    // id is guaranteed to be numeric
    return ctx.JSON(200, map[string]string{"user_id": id})
})

// Match specific patterns
router.GET("/posts/:slug([a-z0-9-]+)", func(ctx pkg.Context) error {
    slug := ctx.Params()["slug"]
    // slug matches lowercase letters, numbers, and hyphens
    return ctx.JSON(200, map[string]string{"slug": slug})
})

// Match date format
router.GET("/archive/:date([0-9]{4}-[0-9]{2}-[0-9]{2})", func(ctx pkg.Context) error {
    date := ctx.Params()["date"]
    // date matches YYYY-MM-DD format
    return ctx.JSON(200, map[string]string{"date": date})
})
```

## Query Strings

Access query string parameters using the `Query()` method:

```go
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query()
    
    // Get individual parameters
    searchTerm := query["q"]
    page := query["page"]
    limit := query["limit"]
    
    // Handle missing parameters
    if searchTerm == "" {
        return ctx.JSON(400, map[string]string{"error": "Missing search term"})
    }
    
    // Use default values
    if page == "" {
        page = "1"
    }
    if limit == "" {
        limit = "10"
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "query": searchTerm,
        "page":  page,
        "limit": limit,
    })
})

// Example request: GET /search?q=golang&page=2&limit=20
```

### Multiple Values

Handle query parameters with multiple values:

```go
router.GET("/filter", func(ctx pkg.Context) error {
    // Get all query parameters
    query := ctx.Query()
    
    // For multiple values, parse the query string manually
    // or use ctx.Request().URL.Query()
    rawQuery := ctx.Request().URL.Query()
    
    // Get array of values
    tags := rawQuery["tag"] // []string{"go", "web", "framework"}
    
    return ctx.JSON(200, map[string]interface{}{
        "tags": tags,
    })
})

// Example request: GET /filter?tag=go&tag=web&tag=framework
```

## Route Groups

Route groups allow you to organize routes with common prefixes and middleware:

### Basic Groups

```go
// Create an API group
api := router.Group("/api")

// All routes in this group have /api prefix
api.GET("/users", listUsersHandler)       // /api/users
api.GET("/users/:id", getUserHandler)     // /api/users/:id
api.POST("/users", createUserHandler)     // /api/users
api.PUT("/users/:id", updateUserHandler)  // /api/users/:id
api.DELETE("/users/:id", deleteUserHandler) // /api/users/:id
```

### Nested Groups

```go
// Create nested groups
api := router.Group("/api")
v1 := api.Group("/v1")
v2 := api.Group("/v2")

// v1 routes
v1.GET("/users", listUsersV1Handler)     // /api/v1/users
v1.GET("/products", listProductsV1Handler) // /api/v1/products

// v2 routes
v2.GET("/users", listUsersV2Handler)     // /api/v2/users
v2.GET("/products", listProductsV2Handler) // /api/v2/products
```

### Groups with Middleware

Apply middleware to all routes in a group:

```go
// Create group with authentication middleware
admin := router.Group("/admin", authMiddleware, adminMiddleware)

// All routes in this group require authentication and admin privileges
admin.GET("/dashboard", dashboardHandler)
admin.GET("/users", adminUsersHandler)
admin.POST("/settings", updateSettingsHandler)

// Middleware functions
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check authentication
    token := ctx.Request().Header.Get("Authorization")
    if token == "" {
        return ctx.JSON(401, map[string]string{"error": "Unauthorized"})
    }
    
    // Continue to next middleware/handler
    return next(ctx)
}

func adminMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check admin privileges
    // ... admin check logic ...
    return next(ctx)
}
```

## Middleware

Middleware functions execute before route handlers and can modify requests, responses, or control flow:

### Route-Specific Middleware

Apply middleware to individual routes:

```go
// Single middleware
router.GET("/protected", protectedHandler, authMiddleware)

// Multiple middleware (executed in order)
router.GET("/admin", adminHandler, authMiddleware, adminMiddleware, loggingMiddleware)
```

### Global Middleware

Apply middleware to all routes:

```go
// Add global middleware
router.Use(loggingMiddleware)
router.Use(corsMiddleware)
router.Use(recoveryMiddleware)

// All routes will use these middleware
router.GET("/", homeHandler)
router.GET("/api/users", usersHandler)
```

### Middleware Examples

**Logging Middleware:**

```go
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Log request
    log.Printf("Request: %s %s", ctx.Request().Method, ctx.Request().URL.Path)
    
    // Call next handler
    err := next(ctx)
    
    // Log response
    duration := time.Since(start)
    log.Printf("Response: %d (took %v)", ctx.Response().Status(), duration)
    
    return err
}
```

**Authentication Middleware:**

```go
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Get token from header
    token := ctx.Request().Header.Get("Authorization")
    
    if token == "" {
        return ctx.JSON(401, map[string]string{"error": "Missing authorization token"})
    }
    
    // Validate token
    userID, err := validateToken(token)
    if err != nil {
        return ctx.JSON(401, map[string]string{"error": "Invalid token"})
    }
    
    // Store user ID in context
    ctx.Set("user_id", userID)
    
    // Continue to next handler
    return next(ctx)
}
```

**CORS Middleware:**

```go
func corsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Set CORS headers
    ctx.Response().Header().Set("Access-Control-Allow-Origin", "*")
    ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    // Handle preflight requests
    if ctx.Request().Method == "OPTIONS" {
        return ctx.NoContent(204)
    }
    
    // Continue to next handler
    return next(ctx)
}
```

**Rate Limiting Middleware:**

```go
func rateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Get client IP
    clientIP := ctx.ClientIP()
    
    // Check rate limit
    if isRateLimited(clientIP) {
        return ctx.JSON(429, map[string]string{
            "error": "Too many requests",
        })
    }
    
    // Continue to next handler
    return next(ctx)
}
```

## Static File Serving

Serve static files and directories:

### Serve Directory

```go
// Serve files from ./public directory at /static/*
router.Static("/static", pkg.NewFilesystem("./public"))

// Access files:
// /static/css/style.css -> ./public/css/style.css
// /static/js/app.js -> ./public/js/app.js
// /static/images/logo.png -> ./public/images/logo.png
```

### Serve Single File

```go
// Serve a single file
router.StaticFile("/favicon.ico", "./public/favicon.ico")
router.StaticFile("/robots.txt", "./public/robots.txt")
```

## Host-Based Routing

Route requests based on the hostname for multi-tenancy:

```go
// Default routes (no specific host)
router.GET("/", defaultHomeHandler)

// Routes for specific host
tenant1 := router.Host("tenant1.example.com")
tenant1.GET("/", tenant1HomeHandler)
tenant1.GET("/dashboard", tenant1DashboardHandler)

// Routes for another host
tenant2 := router.Host("tenant2.example.com")
tenant2.GET("/", tenant2HomeHandler)
tenant2.GET("/dashboard", tenant2DashboardHandler)

// Wildcard host matching
api := router.Host("*.api.example.com")
api.GET("/status", apiStatusHandler)
```

## Common Routing Patterns

### RESTful API Routes

```go
// Resource-based routing
api := router.Group("/api/v1")

// Users resource
users := api.Group("/users")
users.GET("", listUsers)           // GET /api/v1/users
users.POST("", createUser)         // POST /api/v1/users
users.GET("/:id", getUser)         // GET /api/v1/users/:id
users.PUT("/:id", updateUser)      // PUT /api/v1/users/:id
users.PATCH("/:id", patchUser)     // PATCH /api/v1/users/:id
users.DELETE("/:id", deleteUser)   // DELETE /api/v1/users/:id

// Nested resources
users.GET("/:id/posts", getUserPosts)           // GET /api/v1/users/:id/posts
users.POST("/:id/posts", createUserPost)        // POST /api/v1/users/:id/posts
users.GET("/:id/posts/:postID", getUserPost)    // GET /api/v1/users/:id/posts/:postID
```

### Pagination Routes

```go
router.GET("/products", func(ctx pkg.Context) error {
    query := ctx.Query()
    
    // Parse pagination parameters
    page := query["page"]
    if page == "" {
        page = "1"
    }
    
    perPage := query["per_page"]
    if perPage == "" {
        perPage = "10"
    }
    
    // Fetch paginated results
    products := fetchProducts(page, perPage)
    
    return ctx.JSON(200, map[string]interface{}{
        "products": products,
        "page":     page,
        "per_page": perPage,
    })
})

// Example: GET /products?page=2&per_page=20
```

### Search and Filter Routes

```go
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query()
    
    searchTerm := query["q"]
    category := query["category"]
    minPrice := query["min_price"]
    maxPrice := query["max_price"]
    sortBy := query["sort"]
    
    // Apply filters and search
    results := search(searchTerm, category, minPrice, maxPrice, sortBy)
    
    return ctx.JSON(200, map[string]interface{}{
        "results": results,
        "query":   searchTerm,
    })
})

// Example: GET /search?q=laptop&category=electronics&min_price=500&max_price=2000&sort=price_asc
```

### File Upload Routes

```go
router.POST("/upload", func(ctx pkg.Context) error {
    // Get uploaded file
    file, err := ctx.FormFile("file")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "No file uploaded"})
    }
    
    // Save file
    filename := "/uploads/" + file.Filename
    if err := ctx.SaveUploadedFile(file, filename); err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save file"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
        "path":     filename,
    })
})
```

## Route Information

Get information about registered routes:

```go
// Get all routes
routes := router.Routes()

for _, route := range routes {
    fmt.Printf("Method: %s, Path: %s, Host: %s\n", 
        route.Method, route.Path, route.Host)
}
```

## Best Practices

1. **Use Semantic HTTP Methods:** Use GET for reading, POST for creating, PUT for full updates, PATCH for partial updates, and DELETE for deletion.

2. **Group Related Routes:** Use route groups to organize related endpoints and apply common middleware.

3. **Use Path Parameters for Resources:** Use path parameters for resource identifiers (e.g., `/users/:id`) and query strings for filters and options.

4. **Apply Middleware Strategically:** Apply authentication and authorization middleware at the group level for protected routes.

5. **Version Your APIs:** Use route groups to version your APIs (e.g., `/api/v1`, `/api/v2`).

6. **Handle Errors Consistently:** Return consistent error responses across all routes.

7. **Document Your Routes:** Keep documentation up-to-date with your route definitions.

8. **Use Regex Sparingly:** Only use regex patterns when necessary, as they can impact performance.

9. **Validate Input Early:** Validate path parameters and query strings at the beginning of handlers.

10. **Keep Handlers Focused:** Each handler should do one thing well. Use middleware for cross-cutting concerns.

## Troubleshooting

### Route Not Matching

**Problem:** Requests are not matching expected routes.

**Solutions:**
- Check the HTTP method matches (GET vs POST, etc.)
- Verify the path pattern is correct
- Check for typos in path parameters
- Ensure middleware isn't blocking the request
- Use `router.Routes()` to list all registered routes

### Parameter Not Found

**Problem:** Path parameter returns empty string.

**Solutions:**
- Verify parameter name matches the route definition
- Check that the parameter is defined with `:` prefix
- Ensure the URL matches the route pattern

```go
// Route definition
router.GET("/users/:userID", handler)

// Handler
func handler(ctx pkg.Context) error {
    // Correct: matches route definition
    userID := ctx.Params()["userID"]
    
    // Wrong: parameter name doesn't match
    userID := ctx.Params()["id"] // Returns empty string
    
    return ctx.JSON(200, map[string]string{"user_id": userID})
}
```

### Middleware Not Executing

**Problem:** Middleware is not being called.

**Solutions:**
- Ensure middleware is added before route registration
- Check that middleware calls `next(ctx)`
- Verify middleware is added to the correct router/group
- Check for errors in middleware that prevent execution

### Route Conflicts

**Problem:** Multiple routes match the same path.

**Solutions:**
- Make routes more specific
- Use regex patterns to differentiate routes
- Order routes from most specific to least specific
- Use different HTTP methods for the same path

## See Also

- [Middleware Guide](middleware.md) - Detailed middleware documentation
- [Context Guide](context.md) - Request context and parameter handling
- [API Reference: Router](../api/router.md) - Complete router API
- [REST API Example](../examples/rest-api.md) - Complete REST API example
- [Getting Started](../GETTING_STARTED.md) - Basic framework setup
