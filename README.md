# Rockstar Web Framework 🎸

A high-performance, enterprise-grade Go web framework with multi-protocol support, internationalization, and advanced security features.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🌟 Features

### Core Features
- **Multi-Protocol Support**: HTTP/1, HTTP/2, QUIC, WebSocket
- **Multi-API Support**: REST, GraphQL, gRPC, SOAP
- **Context-Driven Architecture**: Unified context for all request/response operations
- **High Performance**: Built with performance in mind using Go's arena package for memory management

### Security
- **Authentication**: OAuth2, JWT support
- **Authorization**: Role-based and action-based access control
- **Built-in Protection**: XSS, CSRF, CORS, X-Frame-Options
- **Request Validation**: Size limits, timeouts, bogus data detection
- **Encrypted Sessions**: AES-encrypted session cookies

### Enterprise Features
- **Multi-Tenancy**: Host-based routing with tenant isolation
- **Session Management**: Database, cache, or filesystem storage
- **Internationalization**: Multi-language support with YAML locale files
- **Database Support**: MySQL, PostgreSQL, MSSQL, SQLite
- **Caching**: Request-level and distributed caching
- **Forward Proxy**: Built-in load balancing and circuit breakers

### Developer Experience
- **Middleware System**: Flexible, configurable middleware pipeline
- **Pipeline Processing**: Parallel data processing with goroutines
- **Template Engine**: Go template support with context passing
- **Error Handling**: Internationalized error messages
- **Monitoring**: Built-in metrics, pprof, SNMP support
- **Graceful Shutdown**: Context-based graceful shutdown

## 📦 Installation

```bash
go get github.com/echterhof/rockstar-web-framework/pkg
```

## 🚀 Quick Start

```go
package main

import (
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
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
        log.Fatal(err)
    }
    
    // Define routes
    app.Router().GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]interface{}{
            "message": "Welcome to Rockstar! 🎸",
        })
    })
    
    // Start server
    log.Fatal(app.Listen(":8080"))
}
```

## 📚 Documentation

### Quick Links

- **[Quick Reference](docs/QUICK_REFERENCE.md)** - Fast lookup for common tasks
- **[Documentation Index](docs/DOCUMENTATION_INDEX.md)** - Complete documentation catalog
- **[Getting Started](docs/GETTING_STARTED.md)** - Step-by-step tutorial
- **[API Reference](docs/API_REFERENCE.md)** - Complete API documentation
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment
- **[Changelog](docs/CHANGELOG.md)** - Documentation updates

### Project Structure

```
├── cmd/                    # Main applications
├── internal/              # Private application and library code
├── pkg/                   # Framework library code
│   ├── framework.go       # Main framework struct
│   ├── server.go          # Server implementation
│   ├── router.go          # Routing engine
│   ├── context.go         # Request context
│   ├── security.go        # Security features
│   ├── database.go        # Database management
│   ├── session.go         # Session management
│   ├── cache.go           # Caching system
│   └── ...                # Other components
├── examples/              # Example applications
│   ├── getting_started.go # Basic example
│   ├── full_featured_app.go # Comprehensive example
│   └── ...                # Feature-specific examples
├── tests/                 # Integration and load tests
├── docs/                  # Detailed documentation
└── README.md              # This file
```

### Core Concepts

#### Framework Instance

The `Framework` struct is the main entry point that wires all components together:

```go
app, err := pkg.New(config)
```

#### Context

The `Context` interface provides unified access to all request/response data and framework features:

```go
func handler(ctx pkg.Context) error {
    // Request data
    params := ctx.Params()
    query := ctx.Query()
    headers := ctx.Headers()
    
    // Framework services
    db := ctx.DB()
    cache := ctx.Cache()
    session := ctx.Session()
    
    // Response
    return ctx.JSON(200, data)
}
```

#### Routing

The router supports multiple routing patterns:

```go
router := app.Router()

// Basic routes
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.GET("/users/:id", getUser)

// Route groups
api := router.Group("/api/v1", authMiddleware)
api.GET("/profile", getProfile)

// Host-based routing (multi-tenancy)
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHome)

// Static files
router.Static("/static", fileSystem)

// WebSocket
router.WebSocket("/ws", wsHandler)
```

#### Middleware

Middleware can be applied globally or per-route:

```go
// Global middleware
app.Use(loggingMiddleware)
app.Use(recoveryMiddleware)

// Route-specific middleware
router.GET("/admin", adminHandler, authMiddleware, adminMiddleware)

// Group middleware
admin := router.Group("/admin", authMiddleware, adminMiddleware)
```

#### Multi-Protocol APIs

```go
// REST API
router.GET("/api/products", restHandler)

// GraphQL
router.GraphQL("/graphql", schema)

// gRPC
router.GRPC(grpcService)

// SOAP
router.SOAP("/soap", soapService)
```

### Configuration

#### Server Configuration

```go
ServerConfig{
    ReadTimeout:     10 * time.Second,
    WriteTimeout:    10 * time.Second,
    IdleTimeout:     60 * time.Second,
    MaxHeaderBytes:  1 << 20,
    EnableHTTP1:     true,
    EnableHTTP2:     true,
    EnableQUIC:      false,
    EnableMetrics:   true,
    MetricsPath:     "/metrics",
    EnablePprof:     true,
    PprofPath:       "/debug/pprof",
}
```

#### Database Configuration

```go
DatabaseConfig{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "myapp",
    Username:        "user",
    Password:        "pass",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}
```

#### Security Configuration

```go
SecurityConfig{
    EnableXFrameOptions: true,
    XFrameOptions:       "DENY",
    EnableCORS:          true,
    EnableCSRF:          true,
    EnableXSS:           true,
    MaxRequestSize:      10 * 1024 * 1024,
    RequestTimeout:      30 * time.Second,
}
```

### Examples

The `examples/` directory contains comprehensive examples:

- **getting_started.go**: Basic framework usage
- **full_featured_app.go**: All features demonstrated
- **rest_api_example.go**: REST API implementation
- **graphql_example.go**: GraphQL API implementation
- **grpc_example.go**: gRPC service implementation
- **multi_server_example.go**: Multi-tenancy setup
- **middleware_example.go**: Custom middleware
- **pipeline_example.go**: Data processing pipelines
- **cache_example.go**: Caching strategies
- **session_example.go**: Session management
- **i18n_example.go**: Internationalization
- **monitoring_example.go**: Metrics and profiling

### Advanced Features

#### Multi-Tenancy

```go
// Register hosts and tenants
app.ServerManager().RegisterHost("api.example.com", hostConfig)
app.ServerManager().RegisterTenant("tenant-1", []string{"api.example.com"})

// Host-specific routes
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHandler)
```

#### Session Management

```go
func handler(ctx pkg.Context) error {
    session := ctx.Session()
    
    // Get session data
    userID := session.Get("user_id")
    
    // Set session data
    session.Set("user_id", "123")
    
    // Save session
    session.Save(ctx)
    
    return ctx.JSON(200, data)
}
```

#### Internationalization

```go
// Load locales
I18nConfig{
    DefaultLocale: "en",
    LocalesPath:   "./locales",
    Locales:       []string{"en", "de", "fr"},
}

// Use in handlers
func handler(ctx pkg.Context) error {
    i18n := ctx.I18n()
    message := i18n.Translate("welcome_message")
    return ctx.String(200, message)
}
```

#### Caching

```go
func handler(ctx pkg.Context) error {
    cache := ctx.Cache()
    
    // Get from cache
    data, found := cache.Get("key")
    if found {
        return ctx.JSON(200, data)
    }
    
    // Fetch data
    data = fetchData()
    
    // Store in cache
    cache.Set("key", data, 5*time.Minute)
    
    return ctx.JSON(200, data)
}
```

#### Monitoring

```go
// Enable metrics
MonitoringConfig{
    EnableMetrics:   true,
    MetricsPath:     "/metrics",
    EnablePprof:     true,
    PprofPath:       "/debug/pprof",
}

// Access metrics
metrics := app.Metrics()
metrics.RecordRequest(method, path, duration)
```

## 🧪 Testing

```bash
# Run unit tests
go test ./pkg/...

# Run integration tests
go test ./tests/...

# Run benchmarks
go test -bench=. ./tests/...
```

## 📊 Performance

The Rockstar Web Framework is designed for high performance:

- Arena-based memory management for efficient garbage collection
- Connection pooling for databases and proxies
- Request-level caching
- Efficient routing with minimal allocations
- Goroutine pooling for concurrent request handling

Benchmark comparisons with GoFiber and Gin are available in `tests/benchmark_test.go`.

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines before submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

Built with Go and inspired by modern web frameworks while adding enterprise-grade features.

## 📞 Support

- Documentation: [docs/](docs/)
- Examples: [examples/](examples/)
- Issues: GitHub Issues
- Discussions: GitHub Discussions

---

Made with ❤️ and 🎸 by the Rockstar Framework Team