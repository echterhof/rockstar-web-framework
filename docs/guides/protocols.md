# Multi-Protocol Support

## Overview

The Rockstar Web Framework provides comprehensive support for multiple network protocols, allowing you to build modern, high-performance applications that leverage the latest protocol technologies. The framework supports HTTP/1.1, HTTP/2, and QUIC (HTTP/3), each with distinct characteristics and use cases.

This guide explains how to configure and use each protocol, understand their features and limitations, and optimize performance for your specific needs.

## Supported Protocols

### HTTP/1.1

HTTP/1.1 is the traditional HTTP protocol that has been the backbone of the web for decades. It's widely supported and well-understood, making it a reliable choice for most applications.

**Key Features:**
- Universal browser and client support
- Simple request/response model
- Connection keep-alive for reusing connections
- Chunked transfer encoding for streaming

**Limitations:**
- Head-of-line blocking (one request at a time per connection)
- Multiple connections needed for parallelism
- Text-based protocol with higher overhead
- No built-in server push

### HTTP/2

HTTP/2 is a binary protocol that improves upon HTTP/1.1 with multiplexing, header compression, and server push capabilities. It requires TLS in most implementations.

**Key Features:**
- Request/response multiplexing over a single connection
- Header compression (HPACK) reduces overhead
- Server push for proactive resource delivery
- Stream prioritization for better resource management
- Binary framing for efficient parsing

**Limitations:**
- Requires TLS in browsers (though h2c exists for cleartext)
- Head-of-line blocking at TCP level
- More complex than HTTP/1.1
- Not all clients support it

### QUIC (HTTP/3)

QUIC is a modern transport protocol built on UDP that provides HTTP/3 support. It eliminates head-of-line blocking and offers improved performance, especially on lossy networks.

**Key Features:**
- Built on UDP instead of TCP
- No head-of-line blocking between streams
- 0-RTT connection establishment for repeat connections
- Built-in TLS 1.3 encryption
- Connection migration (seamless network switching)
- Better performance on lossy networks

**Limitations:**
- Newer protocol with less universal support
- Requires UDP, which may be blocked by some firewalls
- Higher CPU usage for encryption/decryption
- Limited debugging tools compared to HTTP/1.1

## Configuration

### Basic Protocol Configuration

Enable protocols in your framework configuration:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        // Enable HTTP/1.1
        EnableHTTP1: true,
        
        // Enable HTTP/2
        EnableHTTP2: true,
        
        // Enable QUIC/HTTP3
        EnableQUIC: true,
        
        // Timeouts
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
```

### HTTP/1.1 Configuration

HTTP/1.1 can run with or without TLS:

```go
// HTTP/1.1 without TLS (development only)
if err := app.Listen(":8080"); err != nil {
    log.Fatal(err)
}

// HTTP/1.1 with TLS (production)
if err := app.ListenTLS(":8443", "cert.pem", "key.pem"); err != nil {
    log.Fatal(err)
}
```

**Configuration Options:**

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1:    true,
        ReadTimeout:    30 * time.Second,  // Max time to read request
        WriteTimeout:   30 * time.Second,  // Max time to write response
        IdleTimeout:    120 * time.Second, // Max idle time for keep-alive
        MaxHeaderBytes: 1 << 20,           // 1 MB max header size
    },
}
```

### HTTP/2 Configuration

HTTP/2 requires TLS in browsers but supports cleartext (h2c) for server-to-server communication:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,  // Enable HTTP/1.1 for fallback
        EnableHTTP2: true,  // Enable HTTP/2
        
        // HTTP/2 specific settings
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

// Start with TLS (HTTP/2 will be negotiated via ALPN)
if err := app.ListenTLS(":8443", "cert.pem", "key.pem"); err != nil {
    log.Fatal(err)
}
```

**HTTP/2 Cleartext (h2c):**

For server-to-server communication without TLS:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: false, // Disable HTTP/1.1
        EnableHTTP2: true,  // Enable HTTP/2 only
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

// Start without TLS (h2c mode)
if err := app.Listen(":8080"); err != nil {
    log.Fatal(err)
}
```

### QUIC/HTTP3 Configuration

QUIC always requires TLS and runs on UDP:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableQUIC:  true,
        IdleTimeout: 120 * time.Second, // QUIC idle timeout
    },
}

app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}

// Start QUIC server (UDP-based)
if err := app.ListenQUIC(":4433", "cert.pem", "key.pem"); err != nil {
    log.Fatal(err)
}
```

**Multi-Protocol Setup:**

Run HTTP/2 and QUIC simultaneously for protocol negotiation:

```go
// Start HTTP/2 server for initial connection
go func() {
    if err := app.ListenTLS(":8443", "cert.pem", "key.pem"); err != nil {
        log.Printf("HTTP/2 server error: %v", err)
    }
}()

// Add Alt-Svc header to advertise HTTP/3
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    ctx.SetHeader("Alt-Svc", `h3=":4433"; ma=86400`)
    return next(ctx)
})

// Start QUIC server
if err := app.ListenQUIC(":4433", "cert.pem", "key.pem"); err != nil {
    log.Fatal(err)
}
```

## Protocol Selection and Fallback

### Client-Driven Negotiation

Browsers and clients automatically negotiate the best protocol:

1. **Initial Connection:** Client connects via HTTP/1.1 or HTTP/2
2. **Protocol Discovery:** Server sends `Alt-Svc` header advertising HTTP/3
3. **Upgrade:** Client attempts HTTP/3 connection on subsequent requests
4. **Fallback:** If HTTP/3 fails, client falls back to HTTP/2 or HTTP/1.1

### Server Configuration for Fallback

```go
// Enable all protocols for maximum compatibility
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,  // Fallback for old clients
        EnableHTTP2: true,  // Modern browsers
        EnableQUIC:  true,  // Cutting-edge performance
    },
}
```

### Protocol Detection

Detect which protocol a request is using:

```go
router.GET("/api/info", func(ctx pkg.Context) error {
    protocol := ctx.Request().Protocol
    if protocol == "" {
        protocol = ctx.Request().Proto
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "protocol": protocol,
        "message":  "Request received",
    })
})
```

## Protocol-Specific Features

### HTTP/2 Request Cancellation

HTTP/2 supports stream cancellation, allowing clients to cancel requests:

```go
// Add cancellation middleware
app.Use(pkg.CancellationMiddleware())

router.GET("/api/long-running", func(ctx pkg.Context) error {
    for i := 0; i < 100; i++ {
        // Check for cancellation
        select {
        case <-ctx.Context().Done():
            log.Println("Request cancelled by client")
            return ctx.Context().Err()
        default:
        }
        
        // Do work
        time.Sleep(100 * time.Millisecond)
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "status": "completed",
    })
})
```

### HTTP/2 Server Push

HTTP/2 allows servers to push resources proactively:

```go
router.GET("/", func(ctx pkg.Context) error {
    // Push CSS and JS before sending HTML
    // Note: Server push is being deprecated in favor of HTTP/3
    // and early hints (103 status code)
    
    return ctx.HTML(200, "<html>...</html>")
})
```

### QUIC Connection Migration

QUIC supports seamless network switching (e.g., WiFi to cellular):

```go
// QUIC handles connection migration automatically
// No special configuration needed
router.GET("/api/data", func(ctx pkg.Context) error {
    // This handler works the same regardless of network changes
    return ctx.JSON(200, map[string]interface{}{
        "data": "Connection migration handled automatically",
    })
})
```

### QUIC 0-RTT

QUIC supports 0-RTT for repeat connections:

```go
// 0-RTT is handled automatically by the QUIC implementation
// Clients can send data in the first packet for repeat connections
// No server-side configuration needed
```

## Performance Tuning

### Buffer Sizes

Adjust buffer sizes for different protocols:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadBufferSize:  4096,  // 4 KB read buffer
        WriteBufferSize: 4096,  // 4 KB write buffer
    },
}
```

### Connection Limits

Control concurrent connections:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        MaxConnections: 10000,  // Max concurrent connections
    },
}
```

### Timeout Configuration

Set appropriate timeouts for each protocol:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        // HTTP/1.1 and HTTP/2
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
        
        // Graceful shutdown
        ShutdownTimeout: 30 * time.Second,
    },
}
```

### QUIC-Specific Tuning

```go
// QUIC configuration is handled internally
// Key settings:
// - IdleTimeout: How long to keep idle connections
// - KeepAlivePeriod: How often to send keep-alive packets
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        IdleTimeout: 120 * time.Second,  // QUIC idle timeout
    },
}
```

## TLS Configuration

### Certificate Setup

All protocols can use TLS, and QUIC requires it:

```go
// Load certificates
cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
if err != nil {
    log.Fatal(err)
}

// Configure TLS
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion:   tls.VersionTLS12,
    NextProtos:   []string{"h2", "http/1.1"}, // ALPN for HTTP/2
}

config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        TLSConfig: tlsConfig,
    },
}
```

### HSTS (HTTP Strict Transport Security)

Enable HSTS for HTTPS connections:

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHSTS:            true,
        HSTSMaxAge:            365 * 24 * time.Hour, // 1 year
        HSTSIncludeSubdomains: true,
        HSTSPreload:           true,
    },
}
```

### Generating Self-Signed Certificates

For development:

```bash
# Using OpenSSL
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
    -days 365 -nodes -subj '/CN=localhost'

# Using the framework's script
go run examples/generate_keys.go
```

## Best Practices

### Protocol Selection

**Use HTTP/1.1 when:**
- Maximum compatibility is required
- Clients don't support newer protocols
- Simple request/response patterns
- Debugging with standard tools

**Use HTTP/2 when:**
- Building modern web applications
- Need multiplexing for multiple resources
- Want header compression benefits
- Clients support it (most modern browsers)

**Use QUIC when:**
- Need best performance on mobile networks
- Want 0-RTT connection establishment
- Building real-time applications
- Clients support it (Chrome, Firefox, Safari)

### Development vs Production

**Development:**
```go
// Simple HTTP/1.1 for easy debugging
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,
    },
}

if err := app.Listen(":8080"); err != nil {
    log.Fatal(err)
}
```

**Production:**
```go
// All protocols with TLS
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        EnableHTTP1: true,
        EnableHTTP2: true,
        EnableQUIC:  true,
        EnableHSTS:  true,
    },
}

// HTTP/2 server
go func() {
    if err := app.ListenTLS(":443", "cert.pem", "key.pem"); err != nil {
        log.Fatal(err)
    }
}()

// QUIC server
if err := app.ListenQUIC(":443", "cert.pem", "key.pem"); err != nil {
    log.Fatal(err)
}
```

### Monitoring Protocol Usage

Track which protocols clients are using:

```go
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    protocol := ctx.Request().Protocol
    if protocol == "" {
        protocol = ctx.Request().Proto
    }
    
    // Log or track protocol usage
    log.Printf("Request via %s: %s %s", 
        protocol, 
        ctx.Request().Method, 
        ctx.Request().URL.Path)
    
    return next(ctx)
})
```

## Troubleshooting

### HTTP/2 Not Working

**Check ALPN negotiation:**
```go
// Ensure NextProtos is set
tlsConfig := &tls.Config{
    NextProtos: []string{"h2", "http/1.1"},
}
```

**Verify TLS version:**
```go
// HTTP/2 requires TLS 1.2+
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
}
```

### QUIC Connection Failures

**Check UDP port:**
```bash
# Ensure UDP port is open
sudo netstat -ulnp | grep 4433
```

**Verify certificates:**
```bash
# Test certificate validity
openssl s_client -connect localhost:4433
```

**Check firewall:**
```bash
# Ensure UDP traffic is allowed
sudo ufw allow 4433/udp
```

### Performance Issues

**Monitor connection counts:**
```go
// Add metrics middleware
app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    err := next(ctx)
    duration := time.Since(start)
    
    log.Printf("Request took %v via %s", 
        duration, 
        ctx.Request().Protocol)
    
    return err
})
```

**Adjust buffer sizes:**
```go
// Increase for high-throughput scenarios
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadBufferSize:  8192,
        WriteBufferSize: 8192,
    },
}
```

## Complete Example

See the complete examples in the `examples/` directory:

- `examples/quic_server_example.go` - QUIC/HTTP3 server
- `examples/http2_cancellation_example.go` - HTTP/2 features
- `examples/multi_server_example.go` - Multiple protocols

## See Also

- [Server API Reference](../api/server.md) - Server setup and configuration
- [Security Guide](security.md) - TLS and security best practices
- [Performance Guide](performance.md) - Performance optimization
- [Deployment Guide](deployment.md) - Production deployment
