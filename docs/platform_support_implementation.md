# Cross-Platform Support Implementation

## Overview

The Rockstar Web Framework provides comprehensive cross-platform support with platform-specific optimizations for AIX, Unix-like systems (Linux, BSD, macOS), and Windows. The implementation includes advanced features like prefork support and socket option tuning.

## Supported Platforms

### Operating Systems
- **Linux**: Full support with prefork and SO_REUSEPORT
- **Windows**: Full support with prefork
- **macOS (Darwin)**: Full support with SO_REUSEPORT
- **FreeBSD**: Full support with SO_REUSEPORT
- **NetBSD**: Full support with SO_REUSEPORT
- **OpenBSD**: Full support with SO_REUSEPORT
- **DragonFly BSD**: Full support with SO_REUSEPORT
- **AIX**: Full support

### Platform-Specific Features

| Feature | Linux | Windows | macOS | BSD | AIX |
|---------|-------|---------|-------|-----|-----|
| Prefork | ✓ | ✓ | ✗ | ✗ | ✗ |
| SO_REUSEPORT | ✓ | ✗ | ✓ | ✓ | ✗ |
| SO_REUSEADDR | ✓ | ✓ | ✓ | ✓ | ✓ |
| Custom Buffers | ✓ | ✓ | ✓ | ✓ | ✓ |

## Listener Configuration

### Basic Listener

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReuseAddr: true,
}

listener, err := pkg.CreateListener(config)
if err != nil {
    log.Fatal(err)
}
```

### Listener with SO_REUSEPORT

SO_REUSEPORT allows multiple processes to bind to the same port, enabling kernel-level load balancing.

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReusePort: true,  // Only works on Linux, macOS, BSD
    ReuseAddr: true,
}

listener, err := pkg.CreateListener(config)
if err != nil {
    log.Fatal(err)
}
```

### Listener with Custom Buffer Sizes

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReadBuffer: 65536,   // 64KB read buffer
    WriteBuffer: 65536,  // 64KB write buffer
}

listener, err := pkg.CreateListener(config)
if err != nil {
    log.Fatal(err)
}
```

## Prefork Support

Prefork creates multiple worker processes that share the same listening socket, improving performance on multi-core systems.

### Supported Platforms
- Linux
- Windows

### Basic Prefork Configuration

```go
config := pkg.ServerConfig{
    EnableHTTP1: true,
    EnablePrefork: true,
    PreforkWorkers: 4,  // Number of worker processes
}

server := pkg.NewServer(config)
router := pkg.NewRouter()

router.GET("/", func(ctx pkg.Context) error {
    return ctx.String(http.StatusOK, "Hello from worker!")
})

server.SetRouter(router)
server.Listen("0.0.0.0:8080")
```

### Automatic Worker Count

If `PreforkWorkers` is not specified, the framework automatically uses the number of CPU cores:

```go
config := pkg.ServerConfig{
    EnableHTTP1: true,
    EnablePrefork: true,
    // PreforkWorkers defaults to runtime.NumCPU()
}
```

### Prefork with Custom Listener

```go
config := pkg.ServerConfig{
    EnableHTTP1: true,
    ListenerConfig: &pkg.ListenerConfig{
        Network: "tcp",
        EnablePrefork: true,
        PreforkWorkers: 8,
        ReusePort: true,  // Recommended with prefork on Linux
        ReuseAddr: true,
    },
}

server := pkg.NewServer(config)
```

## Platform Detection

### Get Platform Information

```go
info := pkg.GetPlatformInfo()

fmt.Printf("OS: %s\n", info.OS)
fmt.Printf("Architecture: %s\n", info.Arch)
fmt.Printf("CPU Cores: %d\n", info.NumCPU)
fmt.Printf("Supports Prefork: %v\n", info.SupportsPrefork)
fmt.Printf("Supports SO_REUSEPORT: %v\n", info.SupportsReusePort)
```

### Conditional Feature Usage

```go
config := pkg.ServerConfig{
    EnableHTTP1: true,
}

// Enable prefork only on supported platforms
info := pkg.GetPlatformInfo()
if info.SupportsPrefork {
    config.EnablePrefork = true
    config.PreforkWorkers = info.NumCPU
}

// Enable SO_REUSEPORT only on supported platforms
if info.SupportsReusePort {
    config.ListenerConfig = &pkg.ListenerConfig{
        ReusePort: true,
        ReuseAddr: true,
    }
}

server := pkg.NewServer(config)
```

## Horizontal Scaling with Forward Proxy

The framework's forward proxy is designed to work seamlessly with platform-specific features for horizontal scaling.

### Single Server with Prefork

```go
// Backend server with prefork
backendConfig := pkg.ServerConfig{
    EnableHTTP1: true,
    EnablePrefork: true,
    PreforkWorkers: 4,
}

backend := pkg.NewServer(backendConfig)
// ... configure backend routes
backend.Listen("127.0.0.1:8081")
```

### Multiple Servers with SO_REUSEPORT

```go
// Multiple servers on same port (Linux, macOS, BSD)
for i := 0; i < 4; i++ {
    config := pkg.ServerConfig{
        EnableHTTP1: true,
        ListenerConfig: &pkg.ListenerConfig{
            ReusePort: true,
            ReuseAddr: true,
        },
    }
    
    server := pkg.NewServer(config)
    // ... configure routes
    go server.Listen("0.0.0.0:8080")
}
```

### Forward Proxy with Multiple Backends

```go
// Proxy server
proxyConfig := pkg.ServerConfig{
    EnableHTTP1: true,
}

proxy := pkg.NewServer(proxyConfig)
proxyManager := pkg.NewProxyManager()

// Add multiple backend servers
proxyManager.AddBackend(pkg.Backend{
    URL: "http://127.0.0.1:8081",
})
proxyManager.AddBackend(pkg.Backend{
    URL: "http://127.0.0.1:8082",
})

// Configure round-robin load balancing
proxyManager.SetLoadBalancer(pkg.NewRoundRobinBalancer())

proxy.Listen("0.0.0.0:80")
```

## Performance Tuning

### Linux-Specific Optimizations

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReusePort: true,      // Kernel-level load balancing
    ReuseAddr: true,
    ReadBuffer: 262144,   // 256KB
    WriteBuffer: 262144,  // 256KB
    EnablePrefork: true,
    PreforkWorkers: runtime.NumCPU(),
}
```

### Windows-Specific Optimizations

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReuseAddr: true,      // Windows uses SO_REUSEADDR for similar functionality
    ReadBuffer: 262144,
    WriteBuffer: 262144,
    EnablePrefork: true,
    PreforkWorkers: runtime.NumCPU(),
}
```

### macOS/BSD-Specific Optimizations

```go
config := pkg.ListenerConfig{
    Network: "tcp",
    Address: "0.0.0.0:8080",
    ReusePort: true,      // Available on BSD systems
    ReuseAddr: true,
    ReadBuffer: 131072,   // 128KB (BSD may have lower limits)
    WriteBuffer: 131072,
}
```

## Best Practices

### 1. Use Prefork on Supported Platforms

Prefork provides significant performance improvements on multi-core systems:

```go
info := pkg.GetPlatformInfo()
if info.SupportsPrefork {
    config.EnablePrefork = true
    config.PreforkWorkers = info.NumCPU
}
```

### 2. Enable SO_REUSEPORT for Multiple Processes

When running multiple processes, use SO_REUSEPORT for kernel-level load balancing:

```go
if info.SupportsReusePort {
    config.ListenerConfig = &pkg.ListenerConfig{
        ReusePort: true,
        ReuseAddr: true,
    }
}
```

### 3. Tune Buffer Sizes

Adjust buffer sizes based on your workload:

```go
// For high-throughput applications
config.ReadBuffer = 262144   // 256KB
config.WriteBuffer = 262144

// For low-latency applications
config.ReadBuffer = 16384    // 16KB
config.WriteBuffer = 16384
```

### 4. Combine with Forward Proxy

Use the forward proxy for horizontal scaling across multiple machines:

```go
// Backend servers with prefork
for i := 0; i < numBackends; i++ {
    backend := pkg.NewServer(pkg.ServerConfig{
        EnablePrefork: true,
        PreforkWorkers: runtime.NumCPU(),
    })
    go backend.Listen(fmt.Sprintf("127.0.0.1:808%d", i))
}

// Proxy server
proxy := pkg.NewServer(pkg.ServerConfig{})
proxyManager := pkg.NewProxyManager()
// ... configure proxy
proxy.Listen("0.0.0.0:80")
```

## Platform-Specific Notes

### Linux
- Prefork is fully supported and recommended for production
- SO_REUSEPORT provides kernel-level load balancing
- Can combine prefork with SO_REUSEPORT for maximum performance

### Windows
- Prefork is supported using process spawning
- SO_REUSEADDR provides similar functionality to SO_REUSEPORT
- Process management differs from Unix systems

### macOS/BSD
- SO_REUSEPORT is supported for multiple processes
- Prefork is not implemented (use SO_REUSEPORT instead)
- Buffer size limits may be lower than Linux

### AIX
- Standard listener support
- No prefork or SO_REUSEPORT support
- Use forward proxy for horizontal scaling

## Testing

The framework includes comprehensive platform-specific tests:

```bash
# Run all tests
go test ./pkg/...

# Run platform-specific tests
go test ./pkg/ -run Platform

# Run prefork tests (Linux/Windows only)
go test ./pkg/ -run Prefork

# Run listener tests
go test ./pkg/ -run Listener
```

## Troubleshooting

### Prefork Not Working

Check if your platform supports prefork:

```go
info := pkg.GetPlatformInfo()
if !info.SupportsPrefork {
    log.Println("Prefork not supported on this platform")
}
```

### SO_REUSEPORT Errors

Check if your platform supports SO_REUSEPORT:

```go
info := pkg.GetPlatformInfo()
if !info.SupportsReusePort {
    log.Println("SO_REUSEPORT not supported on this platform")
}
```

### Port Already in Use

Ensure SO_REUSEADDR is enabled:

```go
config.ReuseAddr = true
```

## Examples

See the following example files:
- `examples/multi_server_example.go` - Multiple server instances
- `examples/proxy_example.go` - Forward proxy configuration

## Related Documentation

- [Multi-Server Implementation](multi_server_implementation.md)
- [Proxy Implementation](proxy_implementation.md)
- [Monitoring Implementation](monitoring_implementation.md)
