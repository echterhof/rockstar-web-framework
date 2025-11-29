---
title: "Frequently Asked Questions"
description: "Common questions about the Rockstar Web Framework"
category: "troubleshooting"
tags: ["troubleshooting", "faq", "questions"]
version: "1.0.0"
last_updated: "2025-11-28"
related:
  - "debugging.md"
  - "common-errors.md"
---

# Frequently Asked Questions

Common questions and answers about the Rockstar Web Framework.

## Table of Contents

- [General Questions](#general-questions)
- [Configuration](#configuration)
- [Database](#database)
- [Routing](#routing)
- [Security](#security)
- [Performance](#performance)
- [Plugins](#plugins)
- [Multi-Tenancy](#multi-tenancy)
- [Protocols](#protocols)
- [Deployment](#deployment)
- [Migration](#migration)

## General Questions

### What is the Rockstar Web Framework?

Rockstar is a high-performance, enterprise-grade Go web framework designed for building production-ready applications with multi-protocol support. It provides a batteries-included approach with built-in support for databases, caching, sessions, security, internationalization, multi-tenancy, and plugins.

**Key features:**
- Multi-protocol support (HTTP/1, HTTP/2, QUIC, WebSocket)
- Multiple API styles (REST, GraphQL, gRPC, SOAP)
- Enterprise security (OAuth2, JWT, RBAC)
- Multi-tenancy with tenant isolation
- Extensible plugin system
- Comprehensive monitoring and metrics

**See also:**
- [Getting Started Guide](../GETTING_STARTED.md)
- [Architecture Overview](../architecture/overview.md)

### Why choose Rockstar over other Go frameworks?

Rockstar provides a complete, integrated solution rather than requiring you to wire together multiple libraries:

**Compared to Gin/Echo/Fiber:**
- Built-in multi-tenancy support
- Comprehensive plugin system
- Enterprise security features out of the box
- Multi-protocol support (HTTP/2, QUIC)
- Integrated monitoring and metrics
- Database abstraction with multiple drivers

**See also:**
- [Migration from Gin](../migration/from-gin.md)
- [Migration from Echo](../migration/from-echo.md)
- [Migration from Fiber](../migration/from-fiber.md)

### What are the system requirements?

**Minimum requirements:**
- Go 1.25 or higher
- 512MB RAM (1GB+ recommended)
- Linux, macOS, or Windows

**Optional dependencies:**
- Database server (PostgreSQL, MySQL, SQLite, or MSSQL)
- Redis (for distributed caching)
- TLS certificates (for HTTPS/HTTP2/QUIC)

**See also:**
- [Installation Guide](../INSTALLATION.md)

### Is Rockstar production-ready?

Yes! Rockstar v1.0.0 is production-ready and includes:
- Comprehensive test coverage
- Battle-tested in production environments
- Graceful shutdown and error recovery
- Performance optimizations
- Security best practices
- Monitoring and observability

**See also:**
- [Deployment Guide](../guides/deployment.md)
- [Performance Guide](../guides/performance.md)

### How do I get help?

**Self-Service Resources:**
1. **Documentation**: Check the [documentation](../README.md) first
2. **FAQ**: Review this FAQ for common questions
3. **Troubleshooting**: See [Common Errors](common-errors.md) and [Debugging Guide](debugging.md)
4. **Examples**: Review [working examples](../examples/README.md)

**Community Support:**

5. **GitHub Issues** (Recommended for questions and bug reports)
   - [Ask questions or report issues](https://github.com/echterhof/rockstar-web-framework/issues)
   - Get help from the community
   - Share experiences and use cases
   - Discuss feature ideas

6. **GitHub Issues** (For bugs and feature requests)
   - [Search existing issues](https://github.com/echterhof/rockstar-web-framework/issues)
   - [Report bugs](https://github.com/echterhof/rockstar-web-framework/issues/new?template=bug_report.md)
   - [Request features](https://github.com/echterhof/rockstar-web-framework/issues/new?template=feature_request.md)

**When Creating an Issue:**
- Include framework and Go versions
- Provide minimal reproduction example
- Share complete error messages
- Describe what you've tried
- Sanitize sensitive information

**Response Times:**
- Community support: Best effort, typically 24-48 hours
- Bug reports: Prioritized by severity
- Feature requests: Reviewed during planning

**See also:**
- [Troubleshooting Guide](README.md)
- [Contributing Guidelines](../CONTRIBUTING.md)

## Configuration

### How do I configure the framework?

Configuration can be provided through code, files, or environment variables:

**Code configuration:**
```go
config := pkg.FrameworkConfig{
    Port: 8080,
    Host: "0.0.0.0",
    LogLevel: "info",
}

app, err := pkg.New(config)
```

**File configuration:**
```go
config, err := pkg.LoadConfig("config.yaml")
app, err := pkg.New(config)
```

**Environment variables:**
```bash
export ROCKSTAR_PORT=8080
export ROCKSTAR_LOG_LEVEL=info
```

**See also:**
- [Configuration Guide](../guides/configuration.md)

### What configuration formats are supported?

Rockstar supports multiple configuration formats:
- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- TOML (`.toml`)
- INI (`.ini`)

**See also:**
- [Configuration Guide](../guides/configuration.md)

### How do I use environment-specific configurations?

Use different configuration files for each environment:

```go
env := os.Getenv("ENVIRONMENT")
if env == "" {
    env = "development"
}

configFile := fmt.Sprintf("config.%s.yaml", env)
config, err := pkg.LoadConfig(configFile)
```

**See also:**
- [Configuration Guide](../guides/configuration.md)
- [Deployment Guide](../guides/deployment.md)

### Can I validate my configuration?

Yes, use the built-in validation:

```go
config := pkg.FrameworkConfig{
    // ... your configuration
}

if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

**See also:**
- [Configuration Guide](../guides/configuration.md)

## Database

### Which databases are supported?

Rockstar supports multiple database systems:
- **PostgreSQL** - Recommended for production
- **MySQL/MariaDB** - Wide compatibility
- **SQLite** - Great for development and small deployments
- **Microsoft SQL Server** - Enterprise environments

**See also:**
- [Database Guide](../guides/database.md)

### Do I need a database?

No, the database is optional. If you don't configure a database, the framework will use a no-op implementation that safely handles database calls without errors.

```go
// Without database
config := pkg.FrameworkConfig{
    // No Database field
}

// Database features will be disabled
```

**See also:**
- [Database Guide](../guides/database.md)
- [Optional Database Testing](../guides/database.md#optional-database)

### How do I handle database migrations?

Rockstar doesn't include a built-in migration tool, but integrates well with popular migration tools:

**Using golang-migrate:**
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost/db" up
```

**Using goose:**
```bash
goose -dir ./migrations postgres "user=user password=pass dbname=db" up
```

**See also:**
- [Database Guide](../guides/database.md)

### How do I use transactions?

Use the database manager's transaction support:

```go
func handler(ctx pkg.Context) error {
    db := ctx.DB()
    
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed
    
    // Perform operations
    if err := tx.Exec("INSERT INTO users ..."); err != nil {
        return err
    }
    
    // Commit transaction
    return tx.Commit()
}
```

**See also:**
- [Database Guide](../guides/database.md)

### How do I optimize database performance?

**Connection pooling:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        MaxOpenConns: 25,
        MaxIdleConns: 5,
        ConnMaxLifetime: 5 * time.Minute,
    },
}
```

**Query optimization:**
- Use prepared statements
- Add appropriate indexes
- Use EXPLAIN to analyze queries
- Enable query logging to find slow queries

**See also:**
- [Database Guide](../guides/database.md)
- [Performance Guide](../guides/performance.md)

## Routing

### How do I define routes?

Use the router to define routes:

```go
router := app.Router()

// Simple route
router.GET("/users", listUsers)

// Route with parameter
router.GET("/users/:id", getUser)

// Multiple methods
router.Match([]string{"GET", "POST"}, "/api", handler)

// Route group
api := router.Group("/api")
api.GET("/users", listUsers)
api.POST("/users", createUser)
```

**See also:**
- [Routing Guide](../guides/routing.md)

### How do I extract path parameters?

Use the Context methods:

```go
router.GET("/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    
    // Or with type conversion
    id, err := ctx.ParamInt("id")
    if err != nil {
        return pkg.NewValidationError("Invalid user ID", "id")
    }
    
    return ctx.JSON(200, user)
})
```

**See also:**
- [Routing Guide](../guides/routing.md)
- [Context API](../api/context.md)

### How do I handle query parameters?

Use Context query methods:

```go
router.GET("/search", func(ctx pkg.Context) error {
    // Single value
    query := ctx.Query("q")
    
    // With default
    page := ctx.QueryDefault("page", "1")
    
    // As integer
    limit, err := ctx.QueryInt("limit")
    
    // All query parameters
    params := ctx.QueryParams()
    
    return ctx.JSON(200, results)
})
```

**See also:**
- [Routing Guide](../guides/routing.md)

### How do I add middleware?

Add middleware globally or per-route:

```go
// Global middleware
app.Use(loggingMiddleware)
app.Use(authMiddleware)

// Per-route middleware
router.GET("/admin", adminHandler, requireAdmin)

// Group middleware
admin := router.Group("/admin", requireAdmin)
admin.GET("/users", listUsers)
```

**See also:**
- [Middleware Guide](../guides/middleware.md)

### What's the middleware execution order?

Middleware executes in the order it's registered:

```go
app.Use(middleware1) // Executes first
app.Use(middleware2) // Executes second
app.Use(middleware3) // Executes third

// Then the handler executes
```

**See also:**
- [Middleware Guide](../guides/middleware.md)

## Security

### How do I implement authentication?

Rockstar supports multiple authentication methods:

**JWT Authentication:**
```go
router.POST("/login", func(ctx pkg.Context) error {
    // Validate credentials
    user := authenticateUser(username, password)
    
    // Generate JWT
    token, err := ctx.Security().GenerateJWT(user.ID, claims)
    
    return ctx.JSON(200, map[string]string{"token": token})
})

// Protected route
router.GET("/profile", profileHandler, requireAuth)
```

**OAuth2:**
```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        OAuth2: &pkg.OAuth2Config{
            Provider: "google",
            ClientID: os.Getenv("OAUTH_CLIENT_ID"),
            ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
        },
    },
}
```

**See also:**
- [Security Guide](../guides/security.md)

### How do I implement authorization?

Use Role-Based Access Control (RBAC):

```go
// Define roles
security := app.Security()
security.DefineRole("admin", []string{
    "users:read",
    "users:write",
    "users:delete",
})

// Check permissions
router.GET("/admin/users", func(ctx pkg.Context) error {
    if !ctx.Security().HasRole(ctx, "admin") {
        return pkg.NewAuthorizationError("Admin role required")
    }
    
    // Handle request
    return ctx.JSON(200, users)
})

// Or use middleware
router.GET("/admin/users", handler, pkg.RequireRole("admin"))
```

**See also:**
- [Security Guide](../guides/security.md)

### How do I protect against CSRF?

CSRF protection is built-in:

```go
config := pkg.FrameworkConfig{
    Security: &pkg.SecurityConfig{
        EnableCSRF: true,
    },
}

// In templates
<form method="POST">
    <input type="hidden" name="_csrf" value="{{ .CSRFToken }}">
    <!-- form fields -->
</form>

// In AJAX
fetch('/api/endpoint', {
    headers: {
        'X-CSRF-Token': csrfToken
    }
})
```

**See also:**
- [Security Guide](../guides/security.md)

### How do I enable HTTPS?

Configure TLS:

```go
config := pkg.FrameworkConfig{
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile: "key.pem",
    },
}

app, err := pkg.New(config)
app.Listen(":443")
```

**Generate certificates for development:**
```bash
go run examples/generate_keys.go
```

**See also:**
- [Security Guide](../guides/security.md)
- [Protocols Guide](../guides/protocols.md)

## Performance

### How fast is Rockstar?

Rockstar is designed for high performance:
- Handles 50,000+ requests/second on modern hardware
- Low memory footprint with arena-based allocation
- Efficient routing with minimal overhead
- Connection pooling and caching built-in

**See also:**
- [Performance Guide](../guides/performance.md)
- Benchmarks in `tests/benchmark_test.go`

### How do I optimize performance?

**Enable caching:**
```go
config := pkg.FrameworkConfig{
    Cache: &pkg.CacheConfig{
        Enabled: true,
        DefaultTTL: 5 * time.Minute,
    },
}
```

**Connection pooling:**
```go
config := pkg.FrameworkConfig{
    Database: &pkg.DatabaseConfig{
        MaxOpenConns: 25,
        MaxIdleConns: 5,
    },
}
```

**Use HTTP/2:**
```go
config := pkg.FrameworkConfig{
    EnableHTTP2: true,
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile: "key.pem",
    },
}
```

**See also:**
- [Performance Guide](../guides/performance.md)
- [Caching Guide](../guides/caching.md)

### How do I profile my application?

Enable pprof:

```go
config := pkg.FrameworkConfig{
    MonitoringConfig: pkg.MonitoringConfig{
        EnablePprof: true,
        PprofPort: 6060,
    },
}
```

Access at `http://localhost:6060/debug/pprof/`

**See also:**
- [Debugging Guide](debugging.md#performance-profiling)
- [Monitoring Guide](../guides/monitoring.md)

### What's the memory usage?

Rockstar is memory-efficient:
- Base framework: ~10-20MB
- Per-request overhead: ~1-2KB
- Arena-based allocation reduces GC pressure
- Configurable memory limits

**See also:**
- [Performance Guide](../guides/performance.md)

## Plugins

### What is the plugin system?

The plugin system allows you to extend the framework with custom functionality:
- Hot reload support
- Dependency management
- Event-driven architecture
- Isolated execution
- Permission-based access control

**See also:**
- [Plugin Development Guide](../guides/plugins.md)

### How do I create a plugin?

Create a plugin using the template:

```go
package main

import "github.com/echterhof/rockstar-web-framework/pkg"

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Init(ctx pkg.PluginContext) error {
    // Initialize plugin
    return nil
}

func (p *MyPlugin) Start(ctx pkg.PluginContext) error {
    // Start plugin
    return nil
}

func (p *MyPlugin) Stop(ctx pkg.PluginContext) error {
    // Stop plugin
    return nil
}

var Plugin MyPlugin
```

**Build the plugin:**
```bash
go build -buildmode=plugin -o my-plugin.so plugin.go
```

**See also:**
- [Plugin Development Guide](../guides/plugins.md)
- [Plugin API](../api/plugins.md)

### How do I load plugins?

Configure plugin loading:

```go
config := pkg.FrameworkConfig{
    Plugins: &pkg.PluginConfig{
        Directory: "./plugins",
        Enabled: true,
        AutoLoad: true,
    },
}
```

**See also:**
- [Plugin Development Guide](../guides/plugins.md)

### Can plugins access the database?

Yes, plugins have controlled access to framework services:

```go
func (p *MyPlugin) Init(ctx pkg.PluginContext) error {
    // Request database permission
    if ctx.HasPermission("database:read") {
        db := ctx.DB()
        // Use database
    }
    
    return nil
}
```

**See also:**
- [Plugin Development Guide](../guides/plugins.md)

## Multi-Tenancy

### What is multi-tenancy?

Multi-tenancy allows a single application instance to serve multiple tenants (customers/organizations) with data isolation and per-tenant configuration.

**See also:**
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)

### How do I enable multi-tenancy?

Configure multi-tenancy:

```go
config := pkg.FrameworkConfig{
    MultiTenancy: &pkg.MultiTenancyConfig{
        Enabled: true,
        HostBased: true,
    },
}

// Register tenants
app.RegisterTenant(&pkg.Tenant{
    ID: "tenant1",
    Host: "tenant1.example.com",
    Active: true,
})
```

**See also:**
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)

### How does tenant isolation work?

Tenants are isolated at multiple levels:
- Separate database connections
- Isolated cache namespaces
- Separate session storage
- Per-tenant configuration
- Host-based routing

**See also:**
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)

### Can tenants have different databases?

Yes, configure per-tenant databases:

```go
app.RegisterTenant(&pkg.Tenant{
    ID: "tenant1",
    Database: &pkg.DatabaseConfig{
        Driver: "postgres",
        Database: "tenant1_db",
    },
})
```

**See also:**
- [Multi-Tenancy Guide](../guides/multi-tenancy.md)

## Protocols

### Which protocols are supported?

Rockstar supports multiple protocols:
- **HTTP/1.1** - Standard HTTP
- **HTTP/2** - Multiplexing and server push
- **QUIC** - UDP-based protocol (HTTP/3)
- **WebSocket** - Real-time bidirectional communication

**See also:**
- [Protocols Guide](../guides/protocols.md)

### How do I enable HTTP/2?

HTTP/2 requires TLS:

```go
config := pkg.FrameworkConfig{
    EnableHTTP2: true,
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile: "key.pem",
    },
}
```

**See also:**
- [Protocols Guide](../guides/protocols.md)

### How do I enable QUIC?

Configure QUIC (HTTP/3):

```go
config := pkg.FrameworkConfig{
    EnableQUIC: true,
    QUICPort: 443,
    TLS: &pkg.TLSConfig{
        CertFile: "cert.pem",
        KeyFile: "key.pem",
    },
}
```

**See also:**
- [Protocols Guide](../guides/protocols.md)

### How do I use WebSockets?

Define WebSocket routes:

```go
router.WebSocket("/ws", func(ctx pkg.Context) error {
    conn, err := ctx.UpgradeWebSocket()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Process message
        conn.WriteMessage(messageType, response)
    }
    
    return nil
})
```

**See also:**
- [WebSockets Guide](../guides/websockets.md)

## Deployment

### How do I deploy to production?

**Docker deployment:**
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app main.go

FROM alpine:latest
COPY --from=builder /app/app /app
EXPOSE 8080
CMD ["/app"]
```

**See also:**
- [Deployment Guide](../guides/deployment.md)

### How do I handle graceful shutdown?

Graceful shutdown is built-in:

```go
app, err := pkg.New(config)

// Graceful shutdown with 30 second timeout
app.GracefulShutdown(30 * time.Second)

// Start server
app.Listen(":8080")
```

**See also:**
- [Deployment Guide](../guides/deployment.md)

### How do I monitor production applications?

Enable monitoring:

```go
config := pkg.FrameworkConfig{
    MonitoringConfig: pkg.MonitoringConfig{
        EnableMetrics: true,
        MetricsPort: 9090,
    },
}
```

Integrate with Prometheus:
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'rockstar'
    static_configs:
      - targets: ['localhost:9090']
```

**See also:**
- [Monitoring Guide](../guides/monitoring.md)
- [Deployment Guide](../guides/deployment.md)

## Migration

### How do I migrate from Gin?

Rockstar's API is similar to Gin with enhancements:

**Gin:**
```go
r := gin.Default()
r.GET("/users/:id", func(c *gin.Context) {
    c.JSON(200, user)
})
```

**Rockstar:**
```go
app, _ := pkg.New(config)
router := app.Router()
router.GET("/users/:id", func(ctx pkg.Context) error {
    return ctx.JSON(200, user)
})
```

**See also:**
- [Migration from Gin](../migration/from-gin.md)

### How do I migrate from Echo?

Echo and Rockstar have similar patterns:

**Echo:**
```go
e := echo.New()
e.GET("/users/:id", func(c echo.Context) error {
    return c.JSON(200, user)
})
```

**Rockstar:**
```go
app, _ := pkg.New(config)
router := app.Router()
router.GET("/users/:id", func(ctx pkg.Context) error {
    return ctx.JSON(200, user)
})
```

**See also:**
- [Migration from Echo](../migration/from-echo.md)

### How do I migrate from Fiber?

Fiber and Rockstar have similar APIs:

**Fiber:**
```go
app := fiber.New()
app.Get("/users/:id", func(c *fiber.Ctx) error {
    return c.JSON(user)
})
```

**Rockstar:**
```go
app, _ := pkg.New(config)
router := app.Router()
router.GET("/users/:id", func(ctx pkg.Context) error {
    return ctx.JSON(200, user)
})
```

**See also:**
- [Migration from Fiber](../migration/from-fiber.md)

## Common Misconceptions

### "I need to use all features"

**False.** Rockstar is modular - use only what you need:
- Database is optional
- Plugins are optional
- Multi-tenancy is optional
- Most features can be disabled

### "It's too complex for simple applications"

**False.** Rockstar can be as simple as you need:

```go
app, _ := pkg.New(pkg.FrameworkConfig{})
router := app.Router()
router.GET("/", func(ctx pkg.Context) error {
    return ctx.String(200, "Hello World")
})
app.Listen(":8080")
```

### "Performance requires sacrificing features"

**False.** Rockstar is designed for both performance and features:
- High performance by default
- Features don't significantly impact performance
- Unused features have minimal overhead

### "I can't customize the framework"

**False.** Rockstar is highly customizable:
- Plugin system for extensions
- Custom middleware
- Custom managers
- Configuration flexibility

## Still Have Questions?

If your question isn't answered here:

1. **Search the documentation**: [Documentation Home](../README.md)
2. **Check troubleshooting**: [Common Errors](common-errors.md)
3. **Review examples**: [Examples](../examples/README.md)
4. **Search GitHub issues**: [Issues](https://github.com/echterhof/rockstar-web-framework/issues)
5. **Ask a question**: Open a GitHub issue or discussion

## Navigation

- [← Back to Debugging](debugging.md)
- [← Back to Troubleshooting](README.md)
- [Common Errors →](common-errors.md)
