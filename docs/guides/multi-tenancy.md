# Multi-Tenancy

## Overview

Multi-tenancy allows a single instance of the Rockstar Web Framework to serve multiple tenants (customers, organizations, or isolated user groups) with complete data isolation and tenant-specific configuration. Each tenant operates independently with its own database, cache, and configuration while sharing the same application codebase.

**When to use multi-tenancy:**
- Building SaaS applications serving multiple customers
- Creating white-label solutions with tenant-specific branding
- Implementing organization-based access control
- Requiring strict data isolation between customer groups

**Key benefits:**
- **Data Isolation**: Complete separation of tenant data at the database and cache level
- **Host-Based Routing**: Automatic tenant resolution from request hostname
- **Tenant-Specific Configuration**: Custom settings per tenant
- **Resource Efficiency**: Single application instance serves all tenants
- **Scalability**: Add new tenants without code changes

## Quick Start

Here's a minimal example of setting up multi-tenancy:

```go
package main

import (
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Create framework with multi-tenancy enabled
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver: "sqlite",
            DSN:    "tenants.db",
        },
    }

    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // Register tenants with their hostnames
    serverMgr := app.ServerManager()
    
    // Tenant 1: Acme Corp
    serverMgr.RegisterHost("acme.example.com", pkg.HostConfig{
        TenantID: "tenant-acme",
    })
    serverMgr.RegisterTenant("tenant-acme", []string{"acme.example.com"})
    
    // Tenant 2: Globex Inc
    serverMgr.RegisterHost("globex.example.com", pkg.HostConfig{
        TenantID: "tenant-globex",
    })
    serverMgr.RegisterTenant("tenant-globex", []string{"globex.example.com"})

    // Define routes - automatically tenant-aware
    router := app.Router()
    router.GET("/api/data", func(ctx pkg.Context) error {
        // Get tenant ID from context
        tenantID := ctx.Request().TenantID
        
        // Database queries are automatically scoped to tenant
        db := ctx.Database()
        // ... tenant-specific database operations
        
        return ctx.JSON(200, map[string]interface{}{
            "tenant": tenantID,
            "message": "Tenant-specific data",
        })
    })

    log.Fatal(app.Listen(":8080"))
}
```

## Configuration

### Framework Configuration

Enable multi-tenancy in your framework configuration:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        EnableHTTP1:  true,
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver: "sqlite",
        DSN:    "app.db",
        // Multi-tenant database configuration
        MaxOpenConns: 25,
        MaxIdleConns: 5,
    },
    CacheConfig: pkg.CacheConfig{
        DefaultTTL: 5 * time.Minute,
        // Cache is automatically tenant-scoped
    },
}
```

### Host Configuration

Register hosts with tenant associations:

```go
serverMgr := app.ServerManager()

// Register a host
err := serverMgr.RegisterHost("tenant1.example.com", pkg.HostConfig{
    TenantID: "tenant-1",
    // Optional: tenant-specific TLS configuration
    TLSConfig: &tls.Config{
        // ... tenant-specific TLS settings
    },
})
```

### Tenant Registration

Register tenants with their associated hosts:

```go
// Register tenant with one or more hosts
err := serverMgr.RegisterTenant("tenant-1", []string{
    "tenant1.example.com",
    "www.tenant1.example.com",
    "app.tenant1.example.com",
})

if err != nil {
    log.Fatal(err)
}
```

## Usage

### Host-Based Routing

The framework automatically resolves tenants from the request hostname:

```go
router.GET("/api/users", func(ctx pkg.Context) error {
    // Tenant ID is automatically extracted from hostname
    tenantID := ctx.Request().TenantID
    
    if tenantID == "" {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Tenant not found",
        })
    }
    
    // Use tenant ID for business logic
    return ctx.JSON(200, map[string]interface{}{
        "tenant": tenantID,
        "users": getUsersForTenant(tenantID),
    })
})
```

### Tenant-Specific Database Access

Database operations are automatically scoped to the current tenant:

```go
router.POST("/api/documents", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    db := ctx.Database()
    
    // All database operations are tenant-scoped
    // The framework automatically adds tenant_id to queries
    result, err := db.Exec(`
        INSERT INTO documents (tenant_id, title, content)
        VALUES (?, ?, ?)
    `, tenantID, title, content)
    
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Failed to create document",
        })
    }
    
    return ctx.JSON(201, map[string]interface{}{
        "id": result.LastInsertId(),
    })
})
```

### Tenant-Specific Cache Access

Cache operations are automatically namespaced by tenant:

```go
router.GET("/api/profile", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    cache := ctx.Cache()
    
    // Cache keys are automatically prefixed with tenant ID
    cacheKey := "user:profile:" + userID
    
    // Check cache (tenant-scoped)
    if cached, err := cache.Get(cacheKey); err == nil {
        return ctx.JSON(200, cached)
    }
    
    // Fetch from database
    profile := fetchUserProfile(tenantID, userID)
    
    // Store in cache (tenant-scoped)
    cache.Set(cacheKey, profile, 10*time.Minute)
    
    return ctx.JSON(200, profile)
})
```

### Retrieving Tenant Information

Access tenant configuration and metadata:

```go
router.GET("/api/tenant/info", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    serverMgr := app.ServerManager()
    
    // Get tenant information
    tenantInfo, exists := serverMgr.GetTenantInfo(tenantID)
    if !exists {
        return ctx.JSON(404, map[string]interface{}{
            "error": "Tenant not found",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "tenant_id": tenantInfo.TenantID,
        "hosts": tenantInfo.Hosts,
        "config": tenantInfo.Config,
    })
})
```

### Dynamic Tenant Registration

Add tenants at runtime:

```go
router.POST("/admin/tenants", func(ctx pkg.Context) error {
    var req struct {
        TenantID string   `json:"tenant_id"`
        Hosts    []string `json:"hosts"`
    }
    
    if err := ctx.BindJSON(&req); err != nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Invalid request",
        })
    }
    
    serverMgr := app.ServerManager()
    
    // Register hosts
    for _, host := range req.Hosts {
        err := serverMgr.RegisterHost(host, pkg.HostConfig{
            TenantID: req.TenantID,
        })
        if err != nil {
            return ctx.JSON(500, map[string]interface{}{
                "error": "Failed to register host",
            })
        }
    }
    
    // Register tenant
    err := serverMgr.RegisterTenant(req.TenantID, req.Hosts)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Failed to register tenant",
        })
    }
    
    return ctx.JSON(201, map[string]interface{}{
        "message": "Tenant registered successfully",
        "tenant_id": req.TenantID,
    })
})
```

## Integration

### Multi-Tenancy with Authentication

Combine tenant isolation with user authentication:

```go
router.GET("/api/secure/data", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    userID := ctx.Request().UserID
    
    // Verify user belongs to tenant
    if !userBelongsToTenant(userID, tenantID) {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Access denied",
        })
    }
    
    // Fetch tenant-specific data for user
    data := fetchUserData(tenantID, userID)
    
    return ctx.JSON(200, data)
})
```

### Multi-Tenancy with Sessions

Sessions are automatically tenant-scoped:

```go
router.POST("/login", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    
    // Authenticate user
    user := authenticateUser(username, password, tenantID)
    
    // Create tenant-scoped session
    session, err := ctx.Session().Create(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Failed to create session",
        })
    }
    
    session.Set("user_id", user.ID)
    session.Set("tenant_id", tenantID)
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Login successful",
    })
})
```

### Multi-Tenancy with Metrics

Track metrics per tenant:

```go
router.GET("/api/data", func(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    metrics := ctx.Metrics()
    
    // Record tenant-specific metrics
    metrics.IncrementCounter("api.requests", map[string]string{
        "tenant": tenantID,
        "endpoint": "/api/data",
    })
    
    // ... handle request
    
    return ctx.JSON(200, data)
})
```

## Best Practices

### Data Isolation

Always include tenant_id in database tables:

```sql
CREATE TABLE documents (
    id INTEGER PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id)
);
```

Always filter queries by tenant:

```go
// Good: Tenant-scoped query
rows, err := db.Query(`
    SELECT * FROM documents 
    WHERE tenant_id = ? AND user_id = ?
`, tenantID, userID)

// Bad: Missing tenant filter (security risk!)
rows, err := db.Query(`
    SELECT * FROM documents 
    WHERE user_id = ?
`, userID)
```

### Tenant Validation

Always validate tenant exists and is active:

```go
func ValidateTenant(ctx pkg.Context) error {
    tenantID := ctx.Request().TenantID
    
    if tenantID == "" {
        return &pkg.FrameworkError{
            Code: pkg.ErrCodeTenantNotFound,
            Message: "Tenant not found",
            StatusCode: 404,
        }
    }
    
    // Check if tenant is active
    if !isTenantActive(tenantID) {
        return &pkg.FrameworkError{
            Code: pkg.ErrCodeTenantInactive,
            Message: "Tenant is inactive",
            StatusCode: 403,
        }
    }
    
    return nil
}

// Use as middleware
router.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    if err := ValidateTenant(ctx); err != nil {
        return err
    }
    return next(ctx)
})
```

### Cache Key Namespacing

Use consistent cache key patterns:

```go
// Good: Clear tenant-scoped cache keys
cacheKey := fmt.Sprintf("tenant:%s:user:%s:profile", tenantID, userID)

// Good: Use helper function
func TenantCacheKey(tenantID, resource, id string) string {
    return fmt.Sprintf("tenant:%s:%s:%s", tenantID, resource, id)
}
```

### Tenant Configuration

Store tenant-specific configuration:

```go
type TenantConfig struct {
    Features map[string]bool
    Limits   map[string]int
    Branding map[string]string
}

func GetTenantConfig(tenantID string) (*TenantConfig, error) {
    serverMgr := app.ServerManager()
    tenantInfo, exists := serverMgr.GetTenantInfo(tenantID)
    if !exists {
        return nil, errors.New("tenant not found")
    }
    
    // Parse tenant configuration
    config := &TenantConfig{}
    // ... load from tenantInfo.Config
    
    return config, nil
}
```

### Security Considerations

1. **Always validate tenant ownership**: Never trust client-provided tenant IDs
2. **Use tenant-scoped queries**: Always include tenant_id in WHERE clauses
3. **Isolate tenant data**: Use separate database schemas or strict filtering
4. **Audit tenant access**: Log all cross-tenant access attempts
5. **Rate limit per tenant**: Prevent one tenant from affecting others

```go
// Middleware to enforce tenant isolation
func TenantIsolationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    requestTenantID := ctx.Request().TenantID
    
    // If user is authenticated, verify tenant match
    if userID := ctx.Request().UserID; userID != "" {
        userTenantID := getUserTenantID(userID)
        if userTenantID != requestTenantID {
            // Log security event
            log.Warn("Cross-tenant access attempt", 
                "user", userID,
                "user_tenant", userTenantID,
                "request_tenant", requestTenantID)
            
            return ctx.JSON(403, map[string]interface{}{
                "error": "Access denied",
            })
        }
    }
    
    return next(ctx)
}
```

### Performance Optimization

1. **Connection pooling**: Configure appropriate pool sizes per tenant load
2. **Cache tenant metadata**: Avoid repeated tenant lookups
3. **Index tenant_id columns**: Ensure fast tenant-scoped queries
4. **Monitor per-tenant metrics**: Identify resource-intensive tenants

```go
// Cache tenant configuration
var tenantConfigCache = make(map[string]*TenantConfig)
var tenantConfigMutex sync.RWMutex

func GetCachedTenantConfig(tenantID string) (*TenantConfig, error) {
    tenantConfigMutex.RLock()
    if config, exists := tenantConfigCache[tenantID]; exists {
        tenantConfigMutex.RUnlock()
        return config, nil
    }
    tenantConfigMutex.RUnlock()
    
    // Load and cache
    config, err := LoadTenantConfig(tenantID)
    if err != nil {
        return nil, err
    }
    
    tenantConfigMutex.Lock()
    tenantConfigCache[tenantID] = config
    tenantConfigMutex.Unlock()
    
    return config, nil
}
```

## API Reference

See [ServerManager API](../api/server.md) for complete multi-tenancy API documentation.

## Examples

See [Multi-Tenant Application Example](../examples/multi-tenant-app.md) for a complete working implementation.

## Troubleshooting

### Tenant Not Found

**Problem**: Requests return "Tenant not found" error

**Solutions**:
- Verify host is registered: `serverMgr.GetHostConfig(hostname)`
- Check tenant registration: `serverMgr.GetTenantInfo(tenantID)`
- Ensure DNS points to your application
- Verify Host header in requests

### Cross-Tenant Data Leakage

**Problem**: Users see data from other tenants

**Solutions**:
- Always include tenant_id in WHERE clauses
- Use tenant isolation middleware
- Audit database queries for missing tenant filters
- Enable query logging to identify issues

### Performance Issues

**Problem**: Slow queries or high memory usage

**Solutions**:
- Add indexes on tenant_id columns
- Implement tenant-specific connection pools
- Cache tenant configuration
- Monitor per-tenant resource usage
- Consider tenant sharding for large deployments

## Related Documentation

- [Database Guide](database.md) - Database operations with multi-tenancy
- [Caching Guide](caching.md) - Tenant-scoped caching
- [Security Guide](security.md) - Securing multi-tenant applications
- [Configuration Guide](configuration.md) - Tenant-specific configuration
