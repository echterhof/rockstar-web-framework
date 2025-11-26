# Multi-Server Architecture Implementation

## Overview

The Rockstar Web Framework supports running multiple server instances within a single process, enabling efficient resource utilization and flexible deployment architectures. This implementation provides comprehensive multi-server management, host registration, tenant support, and port reuse capabilities.

## Features

- **Multiple Servers in One Process**: Run multiple HTTP/HTTPS/QUIC servers simultaneously
- **Host-Based Routing**: Route requests based on hostname for multi-tenancy
- **Tenant Management**: Associate multiple hosts with tenants for organizational isolation
- **Port Reuse**: Support multiple hosts on the same port with different configurations
- **Graceful Shutdown**: Coordinated shutdown of all servers with timeout support
- **Health Monitoring**: Built-in health checks for all managed servers

## Architecture

### ServerManager

The `ServerManager` is the central component that manages multiple server instances:

```go
type ServerManager interface {
    // Server creation
    NewServer(config ServerConfig) Server
    
    // Multi-server management
    StartAll() error
    StopAll() error
    GracefulShutdown(timeout time.Duration) error
    
    // Host and tenant management
    RegisterHost(hostname string, config HostConfig) error
    RegisterTenant(tenantID string, hosts []string) error
    UnregisterHost(hostname string) error
    UnregisterTenant(tenantID string) error
    
    // Server information
    GetServer(addr string) (Server, bool)
    GetServers() []Server
    GetServerCount() int
    
    // Health checks
    HealthCheck() error
    IsHealthy() bool
}
```

### Key Components

1. **Server Registry**: Maintains a map of address to server instances
2. **Host Registry**: Tracks host configurations and their associated tenants
3. **Tenant Registry**: Manages tenant information and host associations
4. **Port Bindings**: Tracks which addresses are bound to which ports for reuse

## Usage Examples

### Basic Multi-Server Setup

```go
package main

import (
    "log"
    "time"
    "your-framework/pkg"
)

func main() {
    // Create server manager
    sm := pkg.NewServerManager()
    
    // Configure first server (HTTP)
    config1 := pkg.ServerConfig{
        EnableHTTP1:    true,
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   30 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    
    server1 := sm.NewServer(config1)
    server1.SetRouter(createRouter1())
    
    // Configure second server (HTTPS)
    config2 := pkg.ServerConfig{
        EnableHTTP1:    true,
        EnableHTTP2:    true,
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   30 * time.Second,
    }
    
    server2 := sm.NewServer(config2)
    server2.SetRouter(createRouter2())
    
    // Start servers
    go server1.Listen(":8080")
    go server2.ListenTLS(":8443", "cert.pem", "key.pem")
    
    // Wait for shutdown signal
    // ... (signal handling code)
    
    // Graceful shutdown
    if err := sm.GracefulShutdown(30 * time.Second); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

### Multi-Tenant Configuration

```go
package main

import (
    "log"
    "your-framework/pkg"
)

func main() {
    sm := pkg.NewServerManager()
    
    // Register hosts for Tenant 1
    host1Config := pkg.HostConfig{
        Hostname:  "app1.tenant1.com",
        TenantID:  "tenant1",
        VirtualFS: createVirtualFS("tenant1/app1"),
        RateLimits: &pkg.RateLimitConfig{
            Enabled:           true,
            RequestsPerSecond: 100,
            BurstSize:         20,
        },
    }
    
    host2Config := pkg.HostConfig{
        Hostname:  "app2.tenant1.com",
        TenantID:  "tenant1",
        VirtualFS: createVirtualFS("tenant1/app2"),
    }
    
    // Register hosts
    if err := sm.RegisterHost("app1.tenant1.com", host1Config); err != nil {
        log.Fatal(err)
    }
    if err := sm.RegisterHost("app2.tenant1.com", host2Config); err != nil {
        log.Fatal(err)
    }
    
    // Register tenant
    if err := sm.RegisterTenant("tenant1", []string{
        "app1.tenant1.com",
        "app2.tenant1.com",
    }); err != nil {
        log.Fatal(err)
    }
    
    // Create server with host configurations
    config := pkg.ServerConfig{
        EnableHTTP1: true,
        EnableHTTP2: true,
        HostConfigs: map[string]*pkg.HostConfig{
            "app1.tenant1.com": &host1Config,
            "app2.tenant1.com": &host2Config,
        },
    }
    
    server := sm.NewServer(config)
    
    // Set up host-specific routing
    router := pkg.NewRouter()
    
    // Routes for app1.tenant1.com
    app1Router := router.Host("app1.tenant1.com")
    app1Router.GET("/", handleApp1Home)
    app1Router.GET("/api/data", handleApp1Data)
    
    // Routes for app2.tenant1.com
    app2Router := router.Host("app2.tenant1.com")
    app2Router.GET("/", handleApp2Home)
    app2Router.GET("/api/users", handleApp2Users)
    
    server.SetRouter(router)
    
    // Start server
    if err := server.Listen(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### Port Reuse Example

```go
package main

import (
    "log"
    "your-framework/pkg"
)

func main() {
    sm := pkg.NewServerManager()
    
    // Create multiple servers on the same port but different hosts
    // This is useful for virtual hosting scenarios
    
    config := pkg.ServerConfig{
        EnableHTTP1: true,
        HostConfigs: map[string]*pkg.HostConfig{
            "site1.example.com": {
                Hostname:  "site1.example.com",
                VirtualFS: createVirtualFS("site1"),
            },
            "site2.example.com": {
                Hostname:  "site2.example.com",
                VirtualFS: createVirtualFS("site2"),
            },
        },
    }
    
    server := sm.NewServer(config)
    
    // Set up routing for both hosts
    router := pkg.NewRouter()
    
    site1Router := router.Host("site1.example.com")
    site1Router.GET("/", handleSite1)
    
    site2Router := router.Host("site2.example.com")
    site2Router.GET("/", handleSite2)
    
    server.SetRouter(router)
    
    // Both hosts served on port 8080
    if err := server.Listen(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### Health Monitoring

```go
package main

import (
    "log"
    "time"
    "your-framework/pkg"
)

func main() {
    sm := pkg.NewServerManager()
    
    // ... (server setup code)
    
    // Start health monitoring
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            if err := sm.HealthCheck(); err != nil {
                log.Printf("Health check failed: %v", err)
                // Take corrective action
            } else {
                log.Println("All servers healthy")
            }
        }
    }()
    
    // Check if system is healthy
    if sm.IsHealthy() {
        log.Println("System is healthy")
    }
}
```

### Dynamic Server Management

```go
package main

import (
    "log"
    "your-framework/pkg"
)

func main() {
    sm := pkg.NewServerManager()
    
    // Add server dynamically
    config := pkg.ServerConfig{
        EnableHTTP1: true,
    }
    
    server := sm.NewServer(config)
    server.SetRouter(createRouter())
    
    // Add to manager
    if err := sm.(*pkg.serverManager).AddServer("localhost:8080", server); err != nil {
        log.Fatal(err)
    }
    
    // Start the server
    go server.Listen(":8080")
    
    // Later, remove server dynamically
    if err := sm.(*pkg.serverManager).RemoveServer("localhost:8080"); err != nil {
        log.Printf("Failed to remove server: %v", err)
    }
    
    // Get server information
    if srv, exists := sm.GetServer("localhost:8080"); exists {
        log.Printf("Server running: %v", srv.IsRunning())
        log.Printf("Server protocol: %s", srv.Protocol())
    }
    
    // Get all servers
    servers := sm.GetServers()
    log.Printf("Total servers: %d", len(servers))
}
```

## Configuration

### ServerConfig

```go
type ServerConfig struct {
    // Network configuration
    ReadTimeout       time.Duration
    WriteTimeout      time.Duration
    IdleTimeout       time.Duration
    MaxHeaderBytes    int
    
    // Protocol configuration
    EnableHTTP1       bool
    EnableHTTP2       bool
    EnableQUIC        bool
    
    // TLS configuration
    TLSConfig         *tls.Config
    
    // Connection limits
    MaxConnections    int
    MaxRequestSize    int64
    
    // Graceful shutdown
    ShutdownTimeout   time.Duration
    
    // Multi-tenancy
    HostConfigs       map[string]*HostConfig
}
```

### HostConfig

```go
type HostConfig struct {
    Hostname          string
    TenantID          string
    VirtualFS         VirtualFS
    Middleware        []MiddlewareFunc
    RateLimits        *RateLimitConfig
    SecurityConfig    *ServerSecurityConfig
}
```

## Best Practices

### 1. Resource Management

- Use graceful shutdown to ensure clean server termination
- Set appropriate timeouts for read, write, and idle operations
- Monitor server health regularly

### 2. Multi-Tenancy

- Always register hosts before registering tenants
- Use unique tenant IDs across the system
- Configure tenant-specific rate limits and security settings

### 3. Port Reuse

- Use host-based routing for multiple hosts on the same port
- Ensure proper DNS configuration for virtual hosting
- Test host resolution before deployment

### 4. Error Handling

- Always check errors when registering hosts and tenants
- Implement retry logic for transient failures
- Log all server lifecycle events

### 5. Performance

- Use connection pooling for database and cache connections
- Configure appropriate buffer sizes for your workload
- Enable HTTP/2 for better performance with modern clients

## Troubleshooting

### Server Won't Start

- Check if the port is already in use
- Verify TLS certificates are valid and accessible
- Ensure proper file permissions

### Host Registration Fails

- Verify hostname is not already registered
- Check that hostname format is valid
- Ensure no duplicate host configurations

### Tenant Registration Fails

- Verify all hosts are registered before registering tenant
- Check that tenant ID is unique
- Ensure at least one host is provided

### Health Check Fails

- Verify servers are actually running
- Check network connectivity to server addresses
- Review server logs for errors

## Performance Considerations

### Memory Usage

- Each server instance maintains its own connection pool
- Host configurations are shared across servers
- Use arena-based memory allocation for request handling

### CPU Usage

- Multiple servers can utilize multiple CPU cores
- Use prefork mode on Linux/Windows for better CPU utilization
- Monitor CPU usage per server instance

### Scalability

- Horizontal scaling via forward proxy
- Vertical scaling with multiple servers per process
- Load balancing across server instances

## Security Considerations

### Tenant Isolation

- Each tenant has isolated host configurations
- Rate limits are enforced per tenant
- Security settings are tenant-specific

### TLS Configuration

- Use strong cipher suites
- Enable HTTP/2 for better security
- Regularly update TLS certificates

### Access Control

- Implement authentication at the server level
- Use middleware for authorization checks
- Log all access attempts

## Monitoring and Metrics

### Server Metrics

- Request count per server
- Response times per server
- Error rates per server
- Active connections per server

### Host Metrics

- Requests per host
- Bandwidth usage per host
- Error rates per host

### Tenant Metrics

- Resource usage per tenant
- Request patterns per tenant
- Rate limit violations per tenant

## Integration with Other Components

### Router Integration

- Host-based routing for multi-tenancy
- Dynamic route registration per host
- Middleware chaining per host

### Database Integration

- Connection pooling per server
- Tenant-specific database connections
- Transaction management across servers

### Cache Integration

- Distributed caching support
- Cache invalidation across servers
- Tenant-specific cache namespaces

## Future Enhancements

- Dynamic server scaling based on load
- Automatic failover between servers
- Advanced load balancing strategies
- Real-time metrics dashboard
- Automated health recovery

## References

- Related Components: Router, Security, Database
