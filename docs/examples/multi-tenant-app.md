# Multi-Tenant Application Example

The Multi-Tenant Application example demonstrates how to build applications that serve multiple tenants with isolated data and configuration using the Rockstar Web Framework. This pattern is shown in the `full_featured_app.go` example.

## What This Example Demonstrates

- **Host-based routing** for tenant identification
- **Tenant isolation** with separate configurations
- **Per-tenant rate limiting** for resource control
- **Tenant-specific databases** and caches
- **Tenant context** access in handlers
- **Multi-tenancy configuration** patterns

## Prerequisites

- Go 1.25 or higher
- Understanding of DNS/host routing

## Multi-Tenancy Configuration

### Host-Based Configuration

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        // Multi-tenancy with host-specific configs
        HostConfigs: map[string]*pkg.HostConfig{
            "api.example.com": {
                Hostname: "api.example.com",
                TenantID: "tenant-1",
                RateLimits: &pkg.RateLimitConfig{
                    Enabled:           true,
                    RequestsPerSecond: 100,
                    BurstSize:         20,
                    Storage:           "database",
                },
            },
            "admin.example.com": {
                Hostname: "admin.example.com",
                TenantID: "tenant-admin",
                RateLimits: &pkg.RateLimitConfig{
                    Enabled:           true,
                    RequestsPerSecond: 50,
                    BurstSize:         10,
                    Storage:           "database",
                },
            },
        },
    },
}
```

## Host-Based Routing

### Register Tenant-Specific Routes

```go
router := app.Router()

// API tenant routes
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHomeHandler)
apiHost.GET("/products", apiProductsHandler)
apiHost.POST("/orders", apiOrdersHandler)

// Admin tenant routes
adminHost := router.Host("admin.example.com")
adminHost.GET("/", adminHomeHandler)
adminHost.GET("/users", adminUsersHandler)
adminHost.GET("/reports", adminReportsHandler)
```

### Tenant-Specific Handlers

```go
func apiHomeHandler(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "API Host - api.example.com",
        "tenant":  tenant,
        "features": []string{
            "Product catalog",
            "Order management",
            "Inventory tracking",
        },
    })
}

func adminHomeHandler(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Admin Host - admin.example.com",
        "tenant":  tenant,
        "features": []string{
            "User management",
            "System reports",
            "Configuration",
        },
    })
}
```

## Testing Multi-Tenant Applications

### Using Host Header

```bash
# Access API tenant
curl -H "Host: api.example.com" http://localhost:8080/

# Access Admin tenant
curl -H "Host: admin.example.com" http://localhost:8080/
```

### Using /etc/hosts

Add entries to `/etc/hosts`:

```
127.0.0.1 api.example.com
127.0.0.1 admin.example.com
```

Then access directly:

```bash
curl http://api.example.com:8080/
curl http://admin.example.com:8080/
```

## Tenant Isolation Patterns

### Tenant-Specific Databases

```go
func getTenantDatabase(ctx pkg.Context) (pkg.DatabaseManager, error) {
    tenant := ctx.Tenant()
    if tenant == nil {
        return nil, fmt.Errorf("no tenant context")
    }
    
    // Get tenant-specific database connection
    dbConfig := pkg.DatabaseConfig{
        Driver:   "postgres",
        Host:     "localhost",
        Database: fmt.Sprintf("tenant_%s", tenant.ID),
        Username: tenant.DBUser,
        Password: tenant.DBPassword,
    }
    
    return pkg.NewDatabaseManager(dbConfig)
}

func productHandler(ctx pkg.Context) error {
    db, err := getTenantDatabase(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Database error",
        })
    }
    
    // Query tenant-specific database
    products, err := db.Query("SELECT * FROM products")
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Query error",
        })
    }
    
    return ctx.JSON(200, products)
}
```

### Tenant-Specific Caching

```go
func getCachedData(ctx pkg.Context, key string) (interface{}, error) {
    tenant := ctx.Tenant()
    if tenant == nil {
        return nil, fmt.Errorf("no tenant context")
    }
    
    // Prefix cache key with tenant ID
    tenantKey := fmt.Sprintf("tenant:%s:%s", tenant.ID, key)
    
    cache := ctx.Cache()
    return cache.Get(tenantKey)
}

func setCachedData(ctx pkg.Context, key string, value interface{}, ttl time.Duration) error {
    tenant := ctx.Tenant()
    if tenant == nil {
        return fmt.Errorf("no tenant context")
    }
    
    tenantKey := fmt.Sprintf("tenant:%s:%s", tenant.ID, key)
    
    cache := ctx.Cache()
    return cache.Set(tenantKey, value, ttl)
}
```

### Tenant-Specific Configuration

```go
type TenantConfig struct {
    ID              string
    Name            string
    Features        []string
    MaxUsers        int
    MaxStorage      int64
    AllowedDomains  []string
    CustomSettings  map[string]interface{}
}

func getTenantConfig(ctx pkg.Context) (*TenantConfig, error) {
    tenant := ctx.Tenant()
    if tenant == nil {
        return nil, fmt.Errorf("no tenant context")
    }
    
    // Load tenant configuration from database or cache
    config := &TenantConfig{
        ID:   tenant.ID,
        Name: tenant.Name,
        // ... load other settings
    }
    
    return config, nil
}

func featureHandler(ctx pkg.Context) error {
    config, err := getTenantConfig(ctx)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{
            "error": "Configuration error",
        })
    }
    
    // Check if feature is enabled for tenant
    if !hasFeature(config.Features, "advanced_analytics") {
        return ctx.JSON(403, map[string]interface{}{
            "error": "Feature not available for your plan",
        })
    }
    
    // Feature is available, proceed
    return ctx.JSON(200, map[string]interface{}{
        "message": "Feature enabled",
    })
}
```

## Tenant Identification Strategies

### 1. Host-Based (Subdomain)

```
api.example.com → Tenant: api
admin.example.com → Tenant: admin
customer1.example.com → Tenant: customer1
```

**Pros**: Clear separation, easy to configure
**Cons**: Requires DNS configuration

### 2. Path-Based

```
example.com/api → Tenant: api
example.com/admin → Tenant: admin
example.com/customer1 → Tenant: customer1
```

**Pros**: Single domain, easier DNS
**Cons**: More complex routing

### 3. Header-Based

```
X-Tenant-ID: tenant-1
X-Tenant-ID: tenant-2
```

**Pros**: Flexible, no DNS needed
**Cons**: Requires client configuration

### 4. Token-Based

Extract tenant from JWT or API key:

```go
func extractTenantFromToken(ctx pkg.Context) (*Tenant, error) {
    token := ctx.GetHeader("Authorization")
    claims, err := parseJWT(token)
    if err != nil {
        return nil, err
    }
    
    tenantID := claims["tenant_id"].(string)
    return loadTenant(tenantID)
}
```

## Security Considerations

### Tenant Data Isolation

```go
// Middleware to enforce tenant isolation
func tenantIsolationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    tenant := ctx.Tenant()
    if tenant == nil {
        return ctx.JSON(400, map[string]interface{}{
            "error": "Tenant not identified",
        })
    }
    
    // Set tenant context for all database queries
    ctx.Set("tenant_id", tenant.ID)
    
    return next(ctx)
}

// Apply to all routes
app.Use(tenantIsolationMiddleware)
```

### Cross-Tenant Access Prevention

```go
func getResource(ctx pkg.Context, resourceID string) error {
    tenant := ctx.Tenant()
    
    // Query with tenant filter
    var resource Resource
    err := ctx.DB().QueryRow(
        "SELECT * FROM resources WHERE id = ? AND tenant_id = ?",
        resourceID, tenant.ID,
    ).Scan(&resource)
    
    if err != nil {
        return ctx.JSON(404, map[string]interface{}{
            "error": "Resource not found",
        })
    }
    
    return ctx.JSON(200, resource)
}
```

### Tenant-Specific Rate Limiting

```go
// Rate limit per tenant
RateLimits: &pkg.RateLimitConfig{
    Enabled:           true,
    RequestsPerSecond: 100,
    BurstSize:         20,
    Storage:           "database",
    Key:               "tenant_id", // Rate limit by tenant
}
```

## Production Patterns

### Tenant Onboarding

```go
func createTenant(name, domain string) (*Tenant, error) {
    // 1. Create tenant record
    tenant := &Tenant{
        ID:     generateTenantID(),
        Name:   name,
        Domain: domain,
        Status: "active",
    }
    
    // 2. Create tenant database
    err := createTenantDatabase(tenant.ID)
    if err != nil {
        return nil, err
    }
    
    // 3. Initialize tenant schema
    err = initializeTenantSchema(tenant.ID)
    if err != nil {
        return nil, err
    }
    
    // 4. Create default admin user
    err = createTenantAdmin(tenant.ID)
    if err != nil {
        return nil, err
    }
    
    // 5. Save tenant configuration
    err = saveTenantConfig(tenant)
    if err != nil {
        return nil, err
    }
    
    return tenant, nil
}
```

### Tenant Migration

```go
func migrateTenant(tenantID string, newVersion int) error {
    // 1. Backup tenant data
    err := backupTenantData(tenantID)
    if err != nil {
        return err
    }
    
    // 2. Run migrations
    err = runTenantMigrations(tenantID, newVersion)
    if err != nil {
        // Rollback on error
        restoreTenantData(tenantID)
        return err
    }
    
    // 3. Update tenant version
    err = updateTenantVersion(tenantID, newVersion)
    if err != nil {
        return err
    }
    
    return nil
}
```

### Tenant Monitoring

```go
func monitorTenantUsage(ctx pkg.Context) error {
    tenant := ctx.Tenant()
    
    metrics := map[string]interface{}{
        "requests_today":    getTenantRequestCount(tenant.ID),
        "storage_used":      getTenantStorageUsage(tenant.ID),
        "active_users":      getTenantActiveUsers(tenant.ID),
        "api_calls_today":   getTenantAPICallCount(tenant.ID),
    }
    
    // Check limits
    if metrics["storage_used"].(int64) > tenant.MaxStorage {
        return ctx.JSON(429, map[string]interface{}{
            "error": "Storage limit exceeded",
        })
    }
    
    return ctx.JSON(200, metrics)
}
```

## Common Issues

### "Tenant not found"

**Solution**: Verify host configuration and DNS settings

### "Cross-tenant data leak"

**Solution**: Always filter queries by tenant_id

### "Rate limit shared across tenants"

**Solution**: Use tenant-specific rate limit keys

## Related Documentation

- [Multi-Tenancy Guide](../guides/multi-tenancy.md) - Multi-tenant patterns
- [Security Guide](../guides/security.md) - Security best practices
- [Configuration Guide](../guides/configuration.md) - Configuration options
- [Full Featured App](full-featured-app.md) - Complete example

## Source Code

Multi-tenancy examples are available in `examples/full_featured_app.go` in the repository.
