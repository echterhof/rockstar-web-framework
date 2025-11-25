# Getting Started with Rockstar Web Framework

This guide will help you get started with the Rockstar Web Framework, from installation to building your first application.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Your First Application](#your-first-application)
4. [Understanding the Framework](#understanding-the-framework)
5. [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25 or higher**: [Download Go](https://golang.org/dl/)
- **Database** (optional): MySQL, PostgreSQL, MSSQL, or SQLite
- **Redis** (optional): For distributed caching

## Installation

### 1. Create a New Go Project

```bash
mkdir my-rockstar-app
cd my-rockstar-app
go mod init my-rockstar-app
```

### 2. Install the Framework

```bash
go get github.com/echterhof/rockstar-web-framework/pkg
```

### 3. Install Database Driver (Optional)

```bash
# For PostgreSQL
go get github.com/lib/pq

# For MySQL
go get github.com/go-sql-driver/mysql

# For SQLite
go get github.com/mattn/go-sqlite3
```

## Your First Application

### Step 1: Create main.go

Create a file named `main.go` with the following content:

```go
package main

import (
    "log"
    "time"
    "github.com/rockstar-framework/pkg"
)

func main() {
    // Create framework configuration
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
            EnableHTTP2:  true,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "app.db",
        },
    }
    
    // Create framework instance
    app, err := pkg.New(config)
    if err != nil {
        log.Fatalf("Failed to create framework: %v", err)
    }
    
    // Get router
    router := app.Router()
    
    // Define a simple route
    router.GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]interface{}{
            "message": "Welcome to Rockstar! 🎸",
            "status":  "running",
        })
    })
    
    // Start server
    log.Println("Server starting on :8080")
    if err := app.Listen(":8080"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Step 2: Run Your Application

```bash
go run main.go
```

### Step 3: Test Your Application

Open your browser or use curl:

```bash
curl http://localhost:8080/
```

You should see:

```json
{
  "message": "Welcome to Rockstar! 🎸",
  "status": "running"
}
```

## Understanding the Framework

### Framework Architecture

The Rockstar Web Framework follows a modular architecture:

```
┌─────────────────────────────────────┐
│         Your Application            │
├─────────────────────────────────────┤
│  Routes │ Middleware │ Handlers     │
├─────────────────────────────────────┤
│         Framework Core              │
│  Context │ Router │ Security        │
├─────────────────────────────────────┤
│         Protocol Layer              │
│  HTTP/1 │ HTTP/2 │ QUIC │ WebSocket│
├─────────────────────────────────────┤
│          Data Layer                 │
│  Database │ Cache │ Session         │
└─────────────────────────────────────┘
```

### Key Concepts

#### 1. Framework Instance

The `Framework` struct is your application's entry point:

```go
app, err := pkg.New(config)
```

It wires together all components:
- Server manager
- Router
- Database
- Cache
- Session manager
- Security manager
- Monitoring
- And more...

#### 2. Context

The `Context` interface is the heart of request handling:

```go
func handler(ctx pkg.Context) error {
    // Access request data
    params := ctx.Params()      // URL parameters
    query := ctx.Query()        // Query strings
    headers := ctx.Headers()    // HTTP headers
    body := ctx.Body()          // Request body
    
    // Access framework services
    db := ctx.DB()              // Database
    cache := ctx.Cache()        // Cache
    session := ctx.Session()    // Session
    config := ctx.Config()      // Configuration
    i18n := ctx.I18n()          // Internationalization
    
    // Send response
    return ctx.JSON(200, data)
}
```

#### 3. Router

The router maps URLs to handlers:

```go
router := app.Router()

// HTTP methods
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)

// URL parameters
router.GET("/users/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    return ctx.JSON(200, map[string]string{"id": id})
})

// Query parameters
router.GET("/search", func(ctx pkg.Context) error {
    query := ctx.Query()["q"]
    return ctx.JSON(200, map[string]string{"query": query})
})
```

#### 4. Middleware

Middleware processes requests before/after handlers:

```go
// Global middleware
app.Use(loggingMiddleware)

// Route-specific middleware
router.GET("/admin", adminHandler, authMiddleware)

// Middleware function
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    log.Printf("Request: %s %s", ctx.Request().Method, ctx.Request().Path)
    return next(ctx)
}
```

## Building a Real Application

Let's build a simple TODO API:

```go
package main

import (
    "log"
    "time"
    "github.com/rockstar-framework/pkg"
)

type Todo struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

var todos = []Todo{
    {ID: 1, Title: "Learn Rockstar Framework", Completed: false},
    {ID: 2, Title: "Build an API", Completed: false},
}

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "todos.db",
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    router := app.Router()
    
    // List all todos
    router.GET("/todos", func(ctx pkg.Context) error {
        return ctx.JSON(200, todos)
    })
    
    // Get a specific todo
    router.GET("/todos/:id", func(ctx pkg.Context) error {
        id := ctx.Params()["id"]
        for _, todo := range todos {
            if todo.ID == id {
                return ctx.JSON(200, todo)
            }
        }
        return ctx.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    })
    
    // Create a new todo
    router.POST("/todos", func(ctx pkg.Context) error {
        // In production, parse body properly
        newTodo := Todo{
            ID:        len(todos) + 1,
            Title:     "New Todo",
            Completed: false,
        }
        todos = append(todos, newTodo)
        return ctx.JSON(201, newTodo)
    })
    
    // Update a todo
    router.PUT("/todos/:id", func(ctx pkg.Context) error {
        id := ctx.Params()["id"]
        for i, todo := range todos {
            if todo.ID == id {
                todos[i].Completed = !todos[i].Completed
                return ctx.JSON(200, todos[i])
            }
        }
        return ctx.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    })
    
    // Delete a todo
    router.DELETE("/todos/:id", func(ctx pkg.Context) error {
        id := ctx.Params()["id"]
        for i, todo := range todos {
            if todo.ID == id {
                todos = append(todos[:i], todos[i+1:]...)
                return ctx.JSON(204, nil)
            }
        }
        return ctx.JSON(404, map[string]string{
            "error": "Todo not found",
        })
    })
    
    log.Println("TODO API starting on :8080")
    log.Fatal(app.Listen(":8080"))
}
```

### Test the API

```bash
# List todos
curl http://localhost:8080/todos

# Get a specific todo
curl http://localhost:8080/todos/1

# Create a todo
curl -X POST http://localhost:8080/todos

# Update a todo
curl -X PUT http://localhost:8080/todos/1

# Delete a todo
curl -X DELETE http://localhost:8080/todos/1
```

## Next Steps

Now that you have a basic understanding, explore these topics:

### 1. Add Middleware

Learn about logging, authentication, and custom middleware:
- [Middleware Guide](middleware_implementation.md)

### 2. Database Integration

Connect to a real database and perform CRUD operations:
- [Database Guide](database_implementation.md)

### 3. Session Management

Implement user sessions and authentication:
- [Session Guide](session_implementation.md)

### 4. Security Features

Add authentication, authorization, and security headers:
- [Security Guide](security_implementation.md)

### 5. Multi-Protocol APIs

Build REST, GraphQL, gRPC, and SOAP APIs:
- [REST API Guide](rest_api_implementation.md)
- [GraphQL Guide](graphql_implementation.md)
- [gRPC Guide](grpc_implementation.md)

### 6. Plugin System

Extend the framework with plugins:
- [Plugin System Overview](PLUGIN_SYSTEM.md)
- [Plugin Development Guide](PLUGIN_DEVELOPMENT.md)

### 7. Advanced Features

Explore advanced features:
- [Multi-Tenancy](multi_server_implementation.md)
- [Caching](cache_implementation.md)
- [Internationalization](i18n_implementation.md)
- [Monitoring](monitoring_implementation.md)
- [Forward Proxy](proxy_implementation.md)

## Common Patterns

### Error Handling

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    id := ctx.Params()["id"]
    
    user, err := fetchUser(id)
    if err != nil {
        return ctx.JSON(500, map[string]string{
            "error": "Failed to fetch user",
        })
    }
    
    if user == nil {
        return ctx.JSON(404, map[string]string{
            "error": "User not found",
        })
    }
    
    return ctx.JSON(200, user)
})
```

### Request Validation

```go
router.POST("/users", func(ctx pkg.Context) error {
    // Validate required fields
    name := ctx.FormValue("name")
    if name == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Name is required",
        })
    }
    
    // Create user
    user := createUser(name)
    return ctx.JSON(201, user)
})
```

### Using Database

```go
router.GET("/users", func(ctx pkg.Context) error {
    db := ctx.DB()
    
    rows, err := db.Query("SELECT id, name, email FROM users")
    if err != nil {
        return ctx.JSON(500, map[string]string{
            "error": "Database error",
        })
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        rows.Scan(&user.ID, &user.Name, &user.Email)
        users = append(users, user)
    }
    
    return ctx.JSON(200, users)
})
```

## Troubleshooting

### Port Already in Use

If you see "address already in use", change the port:

```go
app.Listen(":8081")  // Use a different port
```

### Database Connection Failed

Check your database configuration:

```go
DatabaseConfig{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "mydb",
    Username: "user",
    Password: "pass",
}
```

### Import Errors

Make sure you've installed all dependencies:

```bash
go mod tidy
```

## Getting Help

- **Documentation**: Check the [docs/](../docs/) directory
- **Examples**: See [examples/](../examples/) for working code
- **Issues**: Report bugs on GitHub Issues
- **Community**: Join discussions on GitHub Discussions

## Using Plugins

The Rockstar Framework includes a powerful plugin system for extending functionality.

### Loading Plugins from Configuration

```go
package main

import (
    "log"
    "time"
    "github.com/rockstar-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
        },
        PluginConfig: pkg.PluginConfig{
            Enabled:   true,
            Directory: "./plugins",
            Plugins: []pkg.PluginLoadConfig{
                {
                    Name:    "auth-plugin",
                    Enabled: true,
                    Path:    "./plugins/auth-plugin",
                    Config: map[string]interface{}{
                        "jwt_secret": "my-secret-key",
                    },
                    Permissions: pkg.PluginPermissions{
                        AllowDatabase: true,
                        AllowCache:    true,
                        AllowRouter:   true,
                    },
                },
                {
                    Name:    "logging-plugin",
                    Enabled: true,
                    Path:    "./plugins/logging-plugin",
                    Config: map[string]interface{}{
                        "log_level": "info",
                    },
                },
            },
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Plugins are automatically loaded and initialized
    
    router := app.Router()
    router.GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]string{
            "message": "Server with plugins running!",
        })
    })
    
    log.Fatal(app.Listen(":8080"))
}
```

### Loading Plugins Dynamically

```go
// Get plugin manager
pluginManager := app.PluginManager()

// Load a plugin at runtime
err := pluginManager.LoadPlugin("./plugins/my-plugin", pkg.PluginConfig{
    Enabled: true,
    Config: map[string]interface{}{
        "setting": "value",
    },
})

// Check if plugin is loaded
if pluginManager.IsLoaded("my-plugin") {
    log.Println("Plugin loaded successfully")
}

// Get plugin information
plugins := pluginManager.ListPlugins()
for _, info := range plugins {
    log.Printf("Plugin: %s v%s - Status: %s", 
        info.Name, info.Version, info.Status)
}
```

### Hot Reloading Plugins

```go
// Reload a plugin without restarting the server
err := pluginManager.ReloadPlugin("my-plugin")
if err != nil {
    log.Printf("Failed to reload plugin: %v", err)
}
```

### Monitoring Plugin Health

```go
// Get health status for a specific plugin
health := pluginManager.GetPluginHealth("my-plugin")
log.Printf("Plugin Status: %s", health.Status)
log.Printf("Error Count: %d", health.ErrorCount)

// Get health for all plugins
allHealth := pluginManager.GetAllHealth()
for name, health := range allHealth {
    log.Printf("%s: %s (errors: %d)", name, health.Status, health.ErrorCount)
}
```

### Example Plugins

The framework includes several example plugins in `examples/plugins/`:

- **minimal-plugin**: Basic plugin structure
- **auth-plugin**: Authentication hooks and middleware
- **logging-plugin**: Request/response logging
- **cache-plugin**: Caching middleware

To use an example plugin:

```bash
# Copy the plugin to your plugins directory
cp -r examples/plugins/auth-plugin ./plugins/

# Update your configuration to load it
# (see configuration example above)

# Run your application
go run main.go
```

### Creating Your Own Plugin

See the [Plugin Development Guide](PLUGIN_DEVELOPMENT.md) for detailed instructions on creating custom plugins.

Basic plugin structure:

```go
package main

import "github.com/rockstar-framework/pkg"

type MyPlugin struct {
    ctx pkg.PluginContext
}

func (p *MyPlugin) Name() string        { return "my-plugin" }
func (p *MyPlugin) Version() string     { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }
func (p *MyPlugin) Author() string      { return "Your Name" }

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    p.ctx = ctx
    
    // Register hooks, middleware, routes, etc.
    ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hctx pkg.HookContext) error {
        p.ctx.Logger().Info("Plugin initialized")
        return nil
    })
    
    return nil
}

func (p *MyPlugin) Start() error   { return nil }
func (p *MyPlugin) Stop() error    { return nil }
func (p *MyPlugin) Cleanup() error { return nil }

func (p *MyPlugin) Dependencies() []pkg.PluginDependency {
    return []pkg.PluginDependency{}
}

func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{}
}

func (p *MyPlugin) OnConfigChange(config map[string]interface{}) error {
    return nil
}

// Plugin entry point
func NewPlugin() pkg.Plugin {
    return &MyPlugin{}
}
```

## Summary

You've learned:
- ✅ How to install the framework
- ✅ How to create your first application
- ✅ Understanding of core concepts (Framework, Context, Router)
- ✅ How to build a simple API
- ✅ How to use the plugin system
- ✅ Common patterns and best practices

Continue learning by exploring the examples and documentation!

---

Happy coding with Rockstar! 🎸
