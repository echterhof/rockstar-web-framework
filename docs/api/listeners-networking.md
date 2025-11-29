# Listeners and Networking API

Complete API reference for network listeners, connection management, and low-level networking in the Rockstar Web Framework.

## Overview

The framework provides flexible network listener management with:
- Custom listener implementations
- Platform-specific optimizations (Unix socket reuse, Windows SO_REUSEADDR)
- Prefork support for multi-process scaling
- Connection pooling and management
- Protocol-specific listeners (HTTP/1, HTTP/2, QUIC)

## Listener Interface

Represents a network listener that accepts connections.

```go
type Listener interface {
    Accept() (net.Conn, error)
    Close() error
    Addr() net.Addr
}
```

### Methods

#### Accept

```go
Accept() (net.Conn, error)
```

Accepts the next incoming connection.

**Returns:**
- `net.Conn`: Accepted connection
- `error`: Error if accept fails

**Example:**
```go
listener, err := net.Listen("tcp", ":8080")
if err != nil {
    panic(err)
}

for {
    conn, err := listener.Accept()
    if err != nil {
        log.Printf("Accept error: %v", err)
        continue
    }
    
    go handleConnection(conn)
}
```

#### Close

```go
Close() error
```

Closes the listener and stops accepting connections.

**Returns:**
- `error`: Error if close fails

#### Addr

```go
Addr() net.Addr
```

Returns the listener's network address.

**Returns:**
- `net.Addr`: Network address

## NetworkClient Interface

Provides HTTP client functionality for plugins and internal use.

```go
type NetworkClient interface {
    Get(url string, headers map[string]string) ([]byte, error)
    Post(url string, body []byte, headers map[string]string) ([]byte, error)
    Put(url string, body []byte, headers map[string]string) ([]byte, error)
    Delete(url string, headers map[string]string) ([]byte, error)
    Do(req *http.Request) (*http.Response, error)
}
```

### Methods

#### Get

```go
Get(url string, headers map[string]string) ([]byte, error)
```

Performs an HTTP GET request.

**Parameters:**
- `url`: Target URL
- `headers`: Request headers

**Returns:**
- `[]byte`: Response body
- `error`: Error if request fails

**Example:**
```go
client := app.NetworkClient()

headers := map[string]string{
    "Authorization": "Bearer token123",
    "Accept": "application/json",
}

body, err := client.Get("https://api.example.com/users", headers)
if err != nil {
    return err
}

var users []User
json.Unmarshal(body, &users)
```

#### Post

```go
Post(url string, body []byte, headers map[string]string) ([]byte, error)
```

Performs an HTTP POST request.

**Parameters:**
- `url`: Target URL
- `body`: Request body
- `headers`: Request headers

**Returns:**
- `[]byte`: Response body
- `error`: Error if request fails

**Example:**
```go
data := map[string]string{
    "name": "John Doe",
    "email": "john@example.com",
}
body, _ := json.Marshal(data)

headers := map[string]string{
    "Content-Type": "application/json",
}

response, err := client.Post("https://api.example.com/users", body, headers)
```

#### Put

```go
Put(url string, body []byte, headers map[string]string) ([]byte, error)
```

Performs an HTTP PUT request.

#### Delete

```go
Delete(url string, headers map[string]string) ([]byte, error)
```

Performs an HTTP DELETE request.

#### Do

```go
Do(req *http.Request) (*http.Response, error)
```

Executes a custom HTTP request with full control.

**Parameters:**
- `req`: HTTP request

**Returns:**
- `*http.Response`: HTTP response
- `error`: Error if request fails

**Example:**
```go
req, _ := http.NewRequest("PATCH", "https://api.example.com/users/123", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer token")

resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()

body, _ := io.ReadAll(resp.Body)
```

## Platform-Specific Listeners

### Unix Listener (Linux, macOS)

Optimized listener for Unix-based systems with SO_REUSEPORT support.

**File:** `pkg/listener_unix.go`

**Features:**
- Socket reuse for load balancing across processes
- Prefork support for multi-process scaling
- Optimized for high-concurrency workloads

**Configuration:**
```go
config := pkg.ServerConfig{
    EnablePrefork: true,
    PreforkWorkers: 4, // Number of worker processes
}
```

### Windows Listener

Optimized listener for Windows with SO_REUSEADDR support.

**File:** `pkg/listener_windows.go`

**Features:**
- Address reuse for rapid restart
- Windows-specific socket options
- Compatible with Windows Server

## Prefork Support

Multi-process scaling for improved performance on multi-core systems.

### Configuration

```go
config := pkg.ServerConfig{
    EnablePrefork: true,
    PreforkWorkers: runtime.NumCPU(), // One worker per CPU core
}

app, _ := pkg.New(pkg.FrameworkConfig{
    ServerConfig: config,
})
```

### How It Works

1. Master process creates listener
2. Forks N worker processes
3. Each worker accepts connections independently
4. OS kernel load-balances connections across workers

### Benefits

- Better CPU utilization on multi-core systems
- Reduced lock contention
- Improved throughput for CPU-bound workloads
- Graceful restart support

### Limitations

- Unix/Linux only (not supported on Windows)
- Increased memory usage (N processes)
- More complex debugging

## Connection Pooling

### ConnectionPool Interface

Manages connection pools for backend services.

```go
type ConnectionPool interface {
    GetConnection(backendID string) (*http.Client, error)
    ReleaseConnection(backendID string, client *http.Client) error
    Close() error
    Stats() ConnectionPoolStats
}
```

### Methods

#### GetConnection

```go
GetConnection(backendID string) (*http.Client, error)
```

Gets a connection from the pool for a backend.

**Parameters:**
- `backendID`: Backend identifier

**Returns:**
- `*http.Client`: HTTP client
- `error`: Error if no connection available

#### ReleaseConnection

```go
ReleaseConnection(backendID string, client *http.Client) error
```

Returns a connection to the pool.

**Parameters:**
- `backendID`: Backend identifier
- `client`: HTTP client to return

**Returns:**
- `error`: Error if release fails

#### Stats

```go
Stats() ConnectionPoolStats
```

Returns connection pool statistics.

**Returns:**
- `ConnectionPoolStats`: Pool statistics

### ConnectionPoolStats

```go
type ConnectionPoolStats struct {
    TotalConnections  int
    ActiveConnections int
    IdleConnections   int
    WaitCount         int64
    WaitDuration      time.Duration
}
```

## Custom Listener Example

Create a custom listener with rate limiting:

```go
type RateLimitedListener struct {
    net.Listener
    limiter *rate.Limiter
}

func NewRateLimitedListener(addr string, rps int) (*RateLimitedListener, error) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }
    
    return &RateLimitedListener{
        Listener: listener,
        limiter: rate.NewLimiter(rate.Limit(rps), rps),
    }, nil
}

func (l *RateLimitedListener) Accept() (net.Conn, error) {
    // Wait for rate limiter
    if err := l.limiter.Wait(context.Background()); err != nil {
        return nil, err
    }
    
    return l.Listener.Accept()
}

// Use custom listener
listener, _ := NewRateLimitedListener(":8080", 1000)
server := app.Server().NewServer(pkg.ServerConfig{})
server.Serve(listener)
```

## Protocol-Specific Listeners

### HTTP/1.1 Listener

Standard TCP listener for HTTP/1.1:

```go
config := pkg.ServerConfig{
    Protocol: "http",
    Address: ":8080",
}

server := app.Server().NewServer(config)
server.Listen(":8080")
```

### HTTP/2 Listener

TLS-enabled listener for HTTP/2:

```go
config := pkg.ServerConfig{
    Protocol: "http2",
    Address: ":8443",
    TLSCertFile: "cert.pem",
    TLSKeyFile: "key.pem",
}

server := app.Server().NewServer(config)
server.ListenTLS(":8443", "cert.pem", "key.pem")
```

### QUIC Listener

UDP-based listener for QUIC protocol:

```go
config := pkg.ServerConfig{
    Protocol: "quic",
    Address: ":8443",
    TLSCertFile: "cert.pem",
    TLSKeyFile: "key.pem",
}

server := app.Server().NewServer(config)
server.ListenQUIC(":8443", "cert.pem", "key.pem")
```

## Network Configuration

### Timeouts

```go
config := pkg.ServerConfig{
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
    ReadHeaderTimeout: 5 * time.Second,
}
```

### Keep-Alive

```go
config := pkg.ServerConfig{
    EnableKeepAlive:  true,
    KeepAlivePeriod:  3 * time.Minute,
    MaxKeepAliveConns: 100,
}
```

### Buffer Sizes

```go
config := pkg.ServerConfig{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
}
```

## Complete Example

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
    "runtime"
    "time"
)

func main() {
    // Configure server with optimized networking
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            // Enable prefork on Unix systems
            EnablePrefork: true,
            PreforkWorkers: runtime.NumCPU(),
            
            // Timeouts
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            IdleTimeout:  120 * time.Second,
            
            // Keep-alive
            EnableKeepAlive: true,
            KeepAlivePeriod: 3 * time.Minute,
            
            // Buffer sizes
            ReadBufferSize:  8192,
            WriteBufferSize: 8192,
            
            // Connection limits
            MaxHeaderBytes: 1 << 20, // 1MB
        },
    }
    
    app, err := pkg.New(config)
    if err != nil {
        panic(err)
    }
    
    // Register routes
    app.Router().GET("/", func(ctx pkg.Context) error {
        return ctx.String(200, "Hello, World!")
    })
    
    // Start server
    if err := app.Start(":8080"); err != nil {
        panic(err)
    }
}
```

## Monitoring Connections

Track connection metrics:

```go
func connectionStatsHandler(ctx pkg.Context) error {
    stats := ctx.Server().Stats()
    
    return ctx.JSON(200, map[string]interface{}{
        "active_connections": stats.ActiveConnections,
        "total_connections": stats.TotalConnections,
        "idle_connections": stats.IdleConnections,
    })
}

app.Router().GET("/stats/connections", connectionStatsHandler)
```

## Best Practices

1. **Use prefork on Unix**: Enable prefork for better multi-core utilization
2. **Set appropriate timeouts**: Prevent resource exhaustion from slow clients
3. **Enable keep-alive**: Reduce connection overhead for repeated requests
4. **Monitor connections**: Track connection metrics for capacity planning
5. **Use connection pooling**: Reuse connections to backend services
6. **Configure buffer sizes**: Tune based on typical request/response sizes
7. **Handle errors gracefully**: Always handle Accept() errors properly
8. **Close connections**: Always close connections when done

## See Also

- [Server API](server.md) - Server management
- [Performance Guide](../guides/performance.md) - Performance tuning
- [Deployment Guide](../guides/deployment.md) - Production deployment
- [Monitoring API](monitoring.md) - Connection monitoring
