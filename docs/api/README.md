# API Reference

This section provides comprehensive API documentation for all public interfaces, types, and functions in the Rockstar Web Framework.

## Quick Reference

| Component | Description | Documentation |
|-----------|-------------|---------------|
| [Framework](framework.md) | Main framework struct and initialization | Core framework API |
| [Context](context.md) | Unified request context interface | Request/response handling |
| [Router](router.md) | Routing engine and route management | HTTP routing and handlers |
| [Server](server.md) | Server management and configuration | Server lifecycle and protocols |
| [Database](database.md) | Database abstraction layer | Multi-database support |
| [Cache](cache.md) | Caching system | In-memory and distributed caching |
| [Session](session.md) | Session management | User session handling |
| [Security](security.md) | Authentication and authorization | OAuth2, JWT, RBAC |
| [I18n](i18n.md) | Internationalization | Multi-language support |
| [Metrics](metrics.md) | Metrics collection | Performance monitoring |
| [Monitoring](monitoring.md) | System monitoring | Health checks and profiling |
| [Proxy](proxy.md) | Forward proxy and load balancing | Backend management |
| [Plugins](plugins.md) | Plugin system | Extensibility framework |
| [Forms & Validation](forms-validation.md) | Form parsing and validation | File uploads and validation rules |
| [Errors & Recovery](errors-recovery.md) | Error handling and panic recovery | Structured errors and logging |
| [Pipeline & Middleware](pipeline-middleware-engine.md) | Advanced middleware and pipelines | Execution control and orchestration |
| [Listeners & Networking](listeners-networking.md) | Network listeners and connections | Low-level networking and pooling |

## API Organization

### Core Components

The framework is organized around a set of manager interfaces that provide access to different subsystems:

- **Framework**: The main entry point that wires all components together
- **Context**: Provides unified access to request/response data and framework services
- **Router**: Handles HTTP routing, middleware, and protocol-specific endpoints
- **Server**: Manages server lifecycle, protocols (HTTP/1, HTTP/2, QUIC), and connections

### Data Layer

Components for data persistence and caching:

- **Database**: Multi-database support (MySQL, PostgreSQL, SQLite, MSSQL)
- **Cache**: In-memory and distributed caching with TTL and tag-based invalidation
- **Session**: Flexible session storage (database, cache, filesystem, in-memory)

### Security

Security-related interfaces:

- **Security**: Authentication (OAuth2, JWT), authorization (RBAC), input validation
- **CSRF/XSS Protection**: Built-in protection against common vulnerabilities
- **Rate Limiting**: Request rate limiting and throttling

### Application Features

Higher-level application features:

- **I18n**: Internationalization with YAML locale files and pluralization
- **Metrics**: Request metrics, workload tracking, and performance analysis
- **Monitoring**: System monitoring, health checks, pprof profiling, SNMP support
- **Proxy**: Forward proxy with load balancing, circuit breakers, and health checks

### Extensibility

Plugin system for extending the framework:

- **Plugins**: Plugin lifecycle, hooks, events, and service registry
- **Plugin Context**: Isolated access to framework services with permission control
- **Event Bus**: Inter-plugin communication

## Common Patterns

### Manager Pattern

Most framework services follow the Manager pattern:

```go
type XxxManager interface {
    // Core operations
    Operation1() error
    Operation2(param Type) (Result, error)
    
    // Configuration
    SetConfig(config XxxConfig) error
    GetConfig() XxxConfig
}
```

Managers are accessed through the Framework or Context:

```go
// Through Framework
db := app.Database()
cache := app.Cache()

// Through Context (in handlers)
func handler(ctx pkg.Context) error {
    db := ctx.DB()
    cache := ctx.Cache()
    // ...
}
```

### Configuration Pattern

All major components use configuration structs with defaults:

```go
type XxxConfig struct {
    // Field with default value
    Field1 string // Default: "value"
    
    // Field with no default (required)
    Field2 string // Required, no default
}

// Apply defaults
config.ApplyDefaults()
```

### Context-Driven Architecture

The Context interface provides unified access to all framework services, eliminating global state:

```go
func handler(ctx pkg.Context) error {
    // Request data
    params := ctx.Params()
    body := ctx.Body()
    
    // Framework services
    db := ctx.DB()
    cache := ctx.Cache()
    i18n := ctx.I18n()
    
    // Response
    return ctx.JSON(200, data)
}
```

## Interface Hierarchy

```
Framework
├── Router (RouterEngine)
├── Server (ServerManager)
├── Database (DatabaseManager)
├── Cache (CacheManager)
├── Session (SessionManager)
├── Security (SecurityManager)
├── Config (ConfigManager)
├── I18n (I18nManager)
├── Metrics (MetricsCollector)
├── Monitoring (MonitoringManager)
├── Proxy (ProxyManager)
└── Plugins (PluginManager)

Context
├── Request/Response
├── All Manager interfaces
└── Helper methods
```

## Type Conventions

### Error Handling

Most methods return `error` as the last return value:

```go
result, err := manager.Operation()
if err != nil {
    return err
}
```

Framework-specific errors implement `FrameworkError`:

```go
type FrameworkError struct {
    Code       string
    Message    string
    StatusCode int
    I18nKey    string
    I18nParams map[string]interface{}
}
```

### Handler Functions

Handler functions follow this signature:

```go
type HandlerFunc func(ctx Context) error
```

Middleware functions wrap handlers:

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

### Configuration Structs

Configuration structs use struct tags for JSON serialization:

```go
type Config struct {
    Field string `json:"field"`
}
```

## Version Compatibility

This API reference documents version 1.0.0 of the Rockstar Web Framework.

- **Stability**: All documented interfaces are considered stable
- **Deprecation**: Deprecated features are clearly marked
- **Breaking Changes**: See [CHANGELOG](../CHANGELOG.md) for breaking changes between versions

## Getting Help

- **Guides**: See [Guides](../guides/README.md) for feature-specific documentation
- **Examples**: See [Examples](../examples/README.md) for working code examples
- **Troubleshooting**: See [Troubleshooting](../troubleshooting/README.md) for common issues
- **Migration**: See [Migration Guides](../migration/README.md) for upgrading from other frameworks

## Next Steps

- Start with the [Framework API](framework.md) to understand initialization
- Learn about [Context API](context.md) for request handling
- Explore [Router API](router.md) for routing and middleware
- Review [Database API](database.md) for data persistence
