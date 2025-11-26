# Rockstar Web Framework 🎸

A high-performance, enterprise-grade Go web framework with multi-protocol support, internationalization, and advanced security features.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🌟 Features

- **Multi-Protocol Support**: HTTP/1, HTTP/2, QUIC, WebSocket
- **Multi-API Support**: REST, GraphQL, gRPC, SOAP
- **Enterprise Security**: OAuth2, JWT, RBAC, CSRF/XSS protection, encrypted sessions
- **Multi-Tenancy**: Host-based routing with tenant isolation
- **Plugin System**: Extensible architecture with hot reload support
- **Internationalization**: Multi-language support with YAML locale files
- **Database Support**: MySQL, PostgreSQL, MSSQL, SQLite with connection pooling
- **Performance**: Arena-based memory management, request-level caching, efficient routing

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
    
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }
    
    app.Router().GET("/", func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]interface{}{
            "message": "Welcome to Rockstar! 🎸",
        })
    })
    
    log.Fatal(app.Listen(":8080"))
}
```

## 📚 Documentation

### Getting Started
- [Getting Started Guide](docs/GETTING_STARTED.md) - Step-by-step tutorial for new users
- [Quick Reference](docs/QUICK_REFERENCE.md) - Fast lookup for common tasks
- [Documentation Index](docs/DOCUMENTATION_INDEX.md) - Complete documentation catalog

### Core Documentation
- [API Reference](docs/API_REFERENCE.md) - Complete API documentation
- [Architecture](docs/ARCHITECTURE.md) - System design and patterns
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment strategies

### Feature Guides
- [Plugin System](docs/PLUGIN_SYSTEM.md) - Plugin architecture overview
- [Plugin Development](docs/PLUGIN_DEVELOPMENT.md) - Creating custom plugins
- [Framework Integration](docs/FRAMEWORK_INTEGRATION.md) - Integration patterns

### Implementation Guides
- [REST API](docs/rest_api_implementation.md) - RESTful API implementation
- [GraphQL](docs/graphql_implementation.md) - GraphQL API setup
- [gRPC](docs/grpc_implementation.md) - gRPC service implementation
- [SOAP](docs/soap_implementation.md) - SOAP service implementation
- [Session Management](docs/session_implementation.md) - Session handling
- [Caching](docs/cache_implementation.md) - Caching strategies
- [Internationalization](docs/i18n_implementation.md) - Multi-language support
- [Monitoring](docs/monitoring_implementation.md) - Metrics and profiling
- [Security](docs/error_handling_implementation.md) - Error handling and security

See [Documentation Index](docs/DOCUMENTATION_INDEX.md) for the complete list.

## 🏗️ Project Structure

```
rockstar-web-framework/
├── cmd/                    # Main applications and CLI tools
├── pkg/                   # Framework library code (single package)
├── examples/              # Example applications and usage demos
├── tests/                 # Integration and benchmark tests
├── docs/                  # Comprehensive documentation
├── sql/                   # Database-specific SQL queries
├── locales/               # Default locale files for i18n
├── plugins/               # Runtime plugin directory
└── scripts/               # Build and utility scripts
```

## 🔧 Development

```bash
# Build the framework
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run static analysis
make vet

# Run all checks
make check
```

## 📖 Examples

The `examples/` directory contains comprehensive examples demonstrating all framework features:

```bash
# Run the getting started example
go run examples/getting_started.go

# Run the full-featured application
go run examples/full_featured_app.go

# Run specific feature examples
go run examples/rest_api_example.go
go run examples/graphql_example.go
go run examples/plugin_usage_example.go
```

## 🧪 Testing

```bash
# Unit tests
go test ./pkg/...

# Integration tests
go test ./tests/...

# Benchmarks
go test -bench=. ./tests/...
```

## 📊 Performance

Designed for high performance with:
- Arena-based memory management
- Connection pooling
- Request-level caching
- Efficient routing with minimal allocations
- Goroutine pooling

Benchmark comparisons available in `tests/benchmark_test.go`.

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/CONTRIBUTING.md) before submitting pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Examples**: [examples/](examples/)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions

---

Made with ❤️ and 🎸 by the Rockstar Framework Team