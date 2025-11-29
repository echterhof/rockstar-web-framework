# Migration Guides

This directory contains guides for migrating to Rockstar Web Framework from other popular Go web frameworks, as well as upgrading between Rockstar versions.

## Available Migration Guides

### Migrating from Other Frameworks

- **[From Gin](from-gin.md)** - Migrate from the Gin web framework
  - Concept mapping between Gin and Rockstar
  - Code translation examples
  - Step-by-step migration checklist
  - Common pitfalls and solutions

- **[From Echo](from-echo.md)** - Migrate from the Echo web framework
  - Concept mapping between Echo and Rockstar
  - Code translation examples
  - Step-by-step migration checklist
  - Common pitfalls and solutions

- **[From Fiber](from-fiber.md)** - Migrate from the Fiber web framework
  - Concept mapping between Fiber and Rockstar
  - Code translation examples
  - Step-by-step migration checklist
  - Common pitfalls and solutions

### Upgrading Rockstar

- **[Upgrading Guide](upgrading.md)** - Upgrade between Rockstar versions
  - Version compatibility information
  - Breaking changes documentation
  - Deprecation notices
  - Upgrade procedures and best practices

## Quick Comparison

### Framework Philosophy

| Framework | Philosophy | Rockstar Advantage |
|-----------|-----------|-------------------|
| **Gin** | Lightweight, high-performance | Adds enterprise features while maintaining performance |
| **Echo** | High-performance, minimalist | Provides comprehensive built-in features |
| **Fiber** | Express-inspired, extreme performance | Standard library compatibility + enterprise features |

### Key Differences

#### Handler Signatures

**Gin:**
```go
func handler(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Hello"})
}
```

**Echo:**
```go
func handler(c echo.Context) error {
    return c.JSON(200, map[string]interface{}{"message": "Hello"})
}
```

**Fiber:**
```go
func handler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Hello"})
}
```

**Rockstar:**
```go
func handler(c pkg.Context) error {
    return c.JSON(200, map[string]interface{}{"message": "Hello"})
}
```

#### Initialization

**Gin:**
```go
router := gin.Default()
router.Run(":8080")
```

**Echo:**
```go
e := echo.New()
e.Start(":8080")
```

**Fiber:**
```go
app := fiber.New()
app.Listen(":8080")
```

**Rockstar:**
```go
config := &pkg.Config{ServerAddress: ":8080"}
app, _ := pkg.New(config)
defer app.Shutdown()
app.Start()
```

## Migration Complexity

### Easy Migration (1-2 days)

- **From Echo**: Very similar design philosophy and handler signatures
- Minimal code changes required
- Mostly initialization and import updates

### Moderate Migration (2-5 days)

- **From Gin**: Similar concepts but different error handling
- Handler signature changes needed
- Middleware pattern updates required

### More Complex Migration (5-10 days)

- **From Fiber**: Different API design and method names
- More extensive code changes needed
- Status code handling differences

## Why Migrate to Rockstar?

### Built-in Enterprise Features

Unlike other frameworks that require third-party libraries, Rockstar includes:

- **Database Integration**: Built-in support for MySQL, PostgreSQL, SQLite, MSSQL
- **Session Management**: Encrypted sessions with multiple storage backends
- **Caching**: Built-in caching with multiple backends
- **Security**: OAuth2, JWT, RBAC, CSRF/XSS protection out of the box
- **Multi-Tenancy**: Host-based routing with tenant isolation
- **Internationalization**: YAML-based i18n support
- **Plugin System**: Extensible architecture

### Multi-Protocol Support

- HTTP/1.1
- HTTP/2 with server push
- QUIC (HTTP/3)
- WebSocket

### Multi-API Support

- RESTful APIs
- GraphQL
- gRPC
- SOAP

### Performance

- Zero-allocation routing for common cases
- Arena-based memory management
- Connection pooling
- Request-level caching
- Efficient middleware chain

### Developer Experience

- Comprehensive documentation
- Working examples for all features
- Context-driven architecture (no globals)
- Consistent error handling
- Type-safe interfaces

## Migration Support

### Resources

- **Documentation**: Complete [documentation](../README.md) for all features
- **Examples**: Working [examples](../../examples/) demonstrating features
- **API Reference**: Detailed [API documentation](../api/README.md)
- **Guides**: Feature-specific [guides](../guides/README.md)

### Getting Help

- **GitHub Issues**: Report migration issues
- **Community**: Join community channels for support
- **Examples**: Reference example applications
- **Documentation**: Search comprehensive docs

## Migration Best Practices

### Before Migration

1. **Understand Your Application**: Document current architecture and dependencies
2. **Review Documentation**: Read relevant migration guide thoroughly
3. **Set Up Test Environment**: Create isolated environment for testing
4. **Backup**: Backup your current codebase
5. **Plan Timeline**: Allocate appropriate time based on complexity

### During Migration

1. **Incremental Approach**: Migrate one component at a time
2. **Test Frequently**: Test after each change
3. **Keep Notes**: Document issues and solutions
4. **Use Version Control**: Commit frequently with clear messages
5. **Leverage Features**: Take advantage of built-in features

### After Migration

1. **Comprehensive Testing**: Test all functionality thoroughly
2. **Performance Testing**: Benchmark against previous version
3. **Code Review**: Review migrated code for best practices
4. **Documentation**: Update internal documentation
5. **Monitor**: Monitor application closely after deployment

## Common Migration Patterns

### Database Integration

**Before (Manual):**
```go
db, _ := sql.Open("mysql", dsn)
defer db.Close()

// Pass db around or use global
```

**After (Rockstar):**
```go
config := &pkg.Config{
    DatabaseDriver: "mysql",
    DatabaseDSN:    dsn,
}
app, _ := pkg.New(config)

// Access via context
func handler(c pkg.Context) error {
    db := c.Database()
    // Use database
}
```

### Session Management

**Before (Manual):**
```go
store := sessions.NewCookieStore([]byte("secret"))
session, _ := store.Get(r, "session-name")
```

**After (Rockstar):**
```go
func handler(c pkg.Context) error {
    session := c.Session()
    session.Set("key", "value")
}
```

### Authentication

**Before (Manual JWT):**
```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte("secret"))
```

**After (Rockstar):**
```go
func handler(c pkg.Context) error {
    security := c.Security()
    token, _ := security.GenerateJWT(userID, claims)
}
```

## Next Steps

1. **Choose Your Migration Guide**: Select the guide for your current framework
2. **Review Requirements**: Check system requirements and dependencies
3. **Set Up Environment**: Create test environment
4. **Follow Guide**: Follow step-by-step migration instructions
5. **Test Thoroughly**: Test all functionality
6. **Deploy**: Deploy to production with monitoring

## Additional Resources

- [Getting Started Guide](../GETTING_STARTED.md)
- [Configuration Guide](../guides/configuration.md)
- [Architecture Overview](../architecture/overview.md)
- [API Reference](../api/README.md)
- [Examples](../../examples/)
- [Troubleshooting](../troubleshooting/README.md)

## Contributing

Help improve migration guides:

- Report migration issues
- Share migration experiences
- Suggest improvements
- Add examples
- Update documentation

See [Contributing Guide](../CONTRIBUTING.md) for details.
