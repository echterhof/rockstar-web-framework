# Examples

This directory contains comprehensive documentation for all example applications included with the Rockstar Web Framework. Each example demonstrates specific features and patterns, with complete code walkthroughs and setup instructions.

## Quick Reference

| Example | Complexity | Features Demonstrated | Documentation |
|---------|-----------|----------------------|---------------|
| [Getting Started](getting-started.md) | ⭐ Beginner | Basic routing, middleware, handlers, lifecycle hooks | Complete tutorial |
| [REST API](rest-api.md) | ⭐⭐ Intermediate | RESTful patterns, CRUD operations, pagination, rate limiting | Full guide |
| [Full Featured App](full-featured-app.md) | ⭐⭐⭐ Advanced | All framework features, production patterns | Comprehensive walkthrough |
| [GraphQL API](graphql-api.md) | ⭐⭐ Intermediate | GraphQL integration, schema definition, resolvers | Complete guide |
| [gRPC API](grpc-api.md) | ⭐⭐ Intermediate | gRPC service, protocol buffers, streaming | Full tutorial |
| [WebSocket Chat](websocket-chat.md) | ⭐⭐ Intermediate | WebSocket connections, real-time messaging | Complete example |
| [SOAP API](soap-api.md) | ⭐⭐ Intermediate | SOAP service, WSDL, XML handling | Full guide |
| [Multi-Tenant App](multi-tenant-app.md) | ⭐⭐⭐ Advanced | Host-based routing, tenant isolation, per-tenant config | Comprehensive guide |
| [Secure App](secure-app.md) | ⭐⭐⭐ Advanced | Authentication, authorization, RBAC, session security | Security patterns |

## Examples by Feature Area

### Core Concepts
- **[Getting Started](getting-started.md)** - Your first Rockstar application
  - Basic routing and handlers
  - Middleware usage
  - Request/response handling
  - Lifecycle hooks
  - Error handling

### API Styles
- **[REST API](rest-api.md)** - RESTful API design
  - CRUD operations
  - Request validation
  - Pagination and filtering
  - Rate limiting
  - CORS configuration

- **[GraphQL API](graphql-api.md)** - GraphQL integration
  - Schema definition
  - Query and mutation resolvers
  - GraphQL playground
  - Error handling

- **[gRPC API](grpc-api.md)** - gRPC service implementation
  - Protocol buffer definitions
  - Service implementation
  - Unary and streaming RPCs
  - Error handling

- **[SOAP API](soap-api.md)** - SOAP web service
  - WSDL generation
  - XML request/response handling
  - SOAP envelope processing

### Real-Time Communication
- **[WebSocket Chat](websocket-chat.md)** - WebSocket implementation
  - Connection management
  - Message broadcasting
  - Room-based chat
  - Connection lifecycle

### Enterprise Features
- **[Full Featured App](full-featured-app.md)** - Production-ready application
  - Database integration
  - Session management
  - Authentication and authorization
  - Caching strategies
  - Internationalization
  - Multi-protocol support
  - Monitoring and profiling
  - Graceful shutdown

- **[Multi-Tenant App](multi-tenant-app.md)** - Multi-tenancy patterns
  - Host-based routing
  - Tenant isolation
  - Per-tenant configuration
  - Tenant-specific databases

- **[Secure App](secure-app.md)** - Security best practices
  - OAuth2 and JWT authentication
  - Role-based access control (RBAC)
  - CSRF and XSS protection
  - Encrypted sessions
  - Security headers

## Examples by Complexity

### Beginner (⭐)
Start here if you're new to the framework:
1. [Getting Started](getting-started.md) - Learn the basics

### Intermediate (⭐⭐)
Build on the basics with specific features:
1. [REST API](rest-api.md) - RESTful patterns
2. [GraphQL API](graphql-api.md) - GraphQL integration
3. [gRPC API](grpc-api.md) - gRPC services
4. [WebSocket Chat](websocket-chat.md) - Real-time communication
5. [SOAP API](soap-api.md) - SOAP services

### Advanced (⭐⭐⭐)
Comprehensive examples for production applications:
1. [Full Featured App](full-featured-app.md) - All features combined
2. [Multi-Tenant App](multi-tenant-app.md) - Multi-tenancy patterns
3. [Secure App](secure-app.md) - Security patterns

## Running the Examples

All examples are located in the `examples/` directory of the repository. Each example is a standalone Go program that can be run directly.

### Prerequisites

- Go 1.25 or higher
- SQLite (for database examples)
- Optional: PostgreSQL, MySQL, or MSSQL for database examples

### Basic Usage

```bash
# Run an example directly
go run examples/getting_started.go

# Build and run an example
go build -o app examples/getting_started.go
./app
```

### Environment Variables

Some examples require environment variables:

```bash
# Session encryption key (required for session examples)
export SESSION_ENCRYPTION_KEY=$(go run examples/generate_keys.go)

# Database configuration (for database examples)
export DB_DRIVER=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=myapp
export DB_USER=myuser
export DB_PASSWORD=mypassword
```

### Generating Encryption Keys

For examples that use sessions or encryption:

```bash
# Generate a secure encryption key
go run examples/generate_keys.go

# Use the generated key
export SESSION_ENCRYPTION_KEY=<generated_key>
```

## Example Code Structure

Each example follows a consistent structure:

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // 1. Configuration setup
    config := pkg.FrameworkConfig{
        // ... configuration options
    }

    // 2. Framework initialization
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Middleware registration
    app.Use(middleware)

    // 4. Route registration
    router := app.Router()
    router.GET("/", handler)

    // 5. Server startup
    app.Listen(":8080")
}
```

## Learning Path

We recommend following this learning path:

1. **Start with basics**: Read [Getting Started](getting-started.md) to understand core concepts
2. **Choose your API style**: Pick the API style that matches your needs (REST, GraphQL, gRPC, SOAP)
3. **Add features**: Explore specific features like WebSockets, caching, or i18n
4. **Study production patterns**: Review [Full Featured App](full-featured-app.md) for production best practices
5. **Implement security**: Learn security patterns from [Secure App](secure-app.md)
6. **Scale with multi-tenancy**: If needed, study [Multi-Tenant App](multi-tenant-app.md)

## Additional Resources

- [API Reference](../api/README.md) - Complete API documentation
- [Feature Guides](../guides/README.md) - In-depth feature documentation
- [Architecture](../architecture/README.md) - Framework design and patterns
- [Getting Started Guide](../GETTING_STARTED.md) - Quick start tutorial

## Contributing Examples

If you've built something interesting with Rockstar and want to share it as an example:

1. Create a standalone example file in `examples/`
2. Add comprehensive comments explaining the code
3. Create documentation in `docs/examples/`
4. Update this README with a link to your example
5. Submit a pull request

See [CONTRIBUTING.md](../CONTRIBUTING.md) for more details.

## Getting Help

- Check the [FAQ](../troubleshooting/faq.md) for common questions
- Review [Troubleshooting Guide](../troubleshooting/README.md) for common issues
- Open an issue on GitHub for bugs or feature requests
- Join our community for discussions and support
