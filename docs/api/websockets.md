# WebSocket API Reference

## Overview

The Rockstar Web Framework provides comprehensive WebSocket support with authentication, routing, middleware, and connection management. Built on top of the Gorilla WebSocket library with additional framework integration.

## WebSocket Connection

### WebSocketConnection Interface

Interface for WebSocket connection operations.

```go
type WebSocketConnection interface {
    ReadMessage() (messageType int, data []byte, err error)
    WriteMessage(messageType int, data []byte) error
    Close() error
    RemoteAddr() string
    LocalAddr() string
}
```

### Message Types

Constants for WebSocket message types.

```go
const (
    TextMessage   = 1  // Text data message
    BinaryMessage = 2  // Binary data message
    CloseMessage  = 8  // Close control message
    PingMessage   = 9  // Ping control message
    PongMessage   = 10 // Pong control message
)
```

### ReadMessage()

Reads a message from the WebSocket connection.

**Signature:**
```go
ReadMessage() (messageType int, data []byte, err error)
```

**Returns:**
- `messageType` - Message type (TextMessage, BinaryMessage, etc.)
- `data` - Message data
- `error` - Error if read fails or connection closed

**Example:**
```go
handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Process message
        fmt.Printf("Received: %s\n", string(data))
        
        // Echo back
        conn.WriteMessage(messageType, data)
    }
}
```

### WriteMessage()

Writes a message to the WebSocket connection.

**Signature:**
```go
WriteMessage(messageType int, data []byte) error
```

**Parameters:**
- `messageType` - Message type (TextMessage or BinaryMessage)
- `data` - Message data

**Returns:**
- `error` - Error if write fails or connection closed

**Example:**
```go
handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Send text message
    err := conn.WriteMessage(pkg.TextMessage, []byte("Hello, client!"))
    if err != nil {
        return err
    }
    
    // Send binary message
    binaryData := []byte{0x01, 0x02, 0x03}
    err = conn.WriteMessage(pkg.BinaryMessage, binaryData)
    if err != nil {
        return err
    }
    
    return nil
}
```

### Close()

Closes the WebSocket connection gracefully.

**Signature:**
```go
Close() error
```

**Returns:**
- `error` - Error if close fails

**Example:**
```go
handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    // Handle messages
    // ...
    
    return nil
}
```

### RemoteAddr()

Returns the remote network address.

**Signature:**
```go
RemoteAddr() string
```

**Returns:**
- `string` - Remote address (e.g., "192.168.1.100:54321")

**Example:**
```go
handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    remoteAddr := conn.RemoteAddr()
    log.Printf("Connection from: %s", remoteAddr)
    
    return nil
}
```

### LocalAddr()

Returns the local network address.

**Signature:**
```go
LocalAddr() string
```

**Returns:**
- `string` - Local address

## WebSocket Server

### WebSocketServer

Manages WebSocket connections and routing.

```go
type WebSocketServer struct {
    // private fields
}
```

### NewWebSocketServer()

Creates a new WebSocket server.

**Signature:**
```go
func NewWebSocketServer(authManager *AuthManager) *WebSocketServer
```

**Parameters:**
- `authManager` - Authentication manager (can be nil if auth not needed)

**Returns:**
- `*WebSocketServer` - WebSocket server instance

**Example:**
```go
authManager := pkg.NewAuthManager(/* config */)
wsServer := pkg.NewWebSocketServer(authManager)
```

### SetUpgrader()

Sets a custom WebSocket upgrader configuration.

**Signature:**
```go
SetUpgrader(upgrader websocket.Upgrader)
```

**Parameters:**
- `upgrader` - Custom WebSocket upgrader

**Example:**
```go
wsServer := pkg.NewWebSocketServer(nil)

upgrader := websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        // Custom origin check
        origin := r.Header.Get("Origin")
        return origin == "https://yourdomain.com"
    },
}

wsServer.SetUpgrader(upgrader)
```

### RegisterRoute()

Registers a WebSocket route.

**Signature:**
```go
RegisterRoute(path string, handler WebSocketHandler, middleware []MiddlewareFunc)
```

**Parameters:**
- `path` - WebSocket endpoint path
- `handler` - WebSocket handler function
- `middleware` - Array of middleware functions

**Example:**
```go
wsServer := pkg.NewWebSocketServer(nil)

handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Echo back
        conn.WriteMessage(messageType, data)
    }
}

wsServer.RegisterRoute("/ws/echo", handler, nil)
```

### RegisterHostRoute()

Registers a WebSocket route for a specific host (multi-tenancy).

**Signature:**
```go
RegisterHostRoute(host, path string, handler WebSocketHandler, middleware []MiddlewareFunc)
```

**Parameters:**
- `host` - Host name
- `path` - WebSocket endpoint path
- `handler` - WebSocket handler function
- `middleware` - Array of middleware functions

**Example:**
```go
wsServer := pkg.NewWebSocketServer(nil)

handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Handle messages for tenant1
    return nil
}

wsServer.RegisterHostRoute("tenant1.example.com", "/ws/chat", handler, nil)
```

### RegisterAuthenticatedRoute()

Registers a WebSocket route that requires authentication.

**Signature:**
```go
RegisterAuthenticatedRoute(path string, handler WebSocketHandler, 
    middleware []MiddlewareFunc, allowedScopes []string)
```

**Parameters:**
- `path` - WebSocket endpoint path
- `handler` - WebSocket handler function
- `middleware` - Array of middleware functions
- `allowedScopes` - Required OAuth scopes

**Example:**
```go
authManager := pkg.NewAuthManager(/* config */)
wsServer := pkg.NewWebSocketServer(authManager)

handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // User is authenticated
    userID := ctx.Request().UserID
    log.Printf("Authenticated user: %s", userID)
    
    return nil
}

wsServer.RegisterAuthenticatedRoute("/ws/private", handler, nil, []string{"chat:read", "chat:write"})
```

### RegisterAuthenticatedHostRoute()

Registers an authenticated WebSocket route for a specific host.

**Signature:**
```go
RegisterAuthenticatedHostRoute(host, path string, handler WebSocketHandler,
    middleware []MiddlewareFunc, allowedScopes []string)
```

**Parameters:**
- `host` - Host name
- `path` - WebSocket endpoint path
- `handler` - WebSocket handler function
- `middleware` - Array of middleware functions
- `allowedScopes` - Required OAuth scopes

**Example:**
```go
authManager := pkg.NewAuthManager(/* config */)
wsServer := pkg.NewWebSocketServer(authManager)

handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    tenantID := ctx.Request().TenantID
    userID := ctx.Request().UserID
    
    // Handle authenticated tenant-specific WebSocket
    return nil
}

wsServer.RegisterAuthenticatedHostRoute("tenant1.example.com", "/ws/admin", 
    handler, nil, []string{"admin:access"})
```

### HandleUpgrade()

Handles WebSocket upgrade requests. Called by the router.

**Signature:**
```go
HandleUpgrade(ctx Context) error
```

**Parameters:**
- `ctx` - Request context

**Returns:**
- `error` - Error if upgrade fails

**Example:**
```go
// Typically called by router
router.GET("/ws/chat", func(ctx pkg.Context) error {
    return wsServer.HandleUpgrade(ctx)
})
```

### GetConnections()

Returns the number of active WebSocket connections.

**Signature:**
```go
GetConnections() int
```

**Returns:**
- `int` - Number of active connections

## WebSocket Handler

### WebSocketHandler Type

Function signature for WebSocket handlers.

```go
type WebSocketHandler func(ctx Context, conn WebSocketConnection) error
```

**Parameters:**
- `ctx` - Request context
- `conn` - WebSocket connection

**Returns:**
- `error` - Error if handling fails

## Complete Examples

### Simple Echo Server

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    
    wsServer := pkg.NewWebSocketServer(nil)
    
    // Echo handler
    echoHandler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }
            
            // Echo back
            err = conn.WriteMessage(messageType, data)
            if err != nil {
                return err
            }
        }
    }
    
    wsServer.RegisterRoute("/ws/echo", echoHandler, nil)
    
    router := app.Router()
    router.GET("/ws/echo", func(ctx pkg.Context) error {
        return wsServer.HandleUpgrade(ctx)
    })
    
    app.Listen(":8080")
}
```

### Chat Application

```go
type ChatRoom struct {
    clients   map[*pkg.WebSocketConnection]bool
    broadcast chan []byte
    mu        sync.Mutex
}

func NewChatRoom() *ChatRoom {
    room := &ChatRoom{
        clients:   make(map[*pkg.WebSocketConnection]bool),
        broadcast: make(chan []byte, 256),
    }
    
    go room.run()
    return room
}

func (room *ChatRoom) run() {
    for {
        message := <-room.broadcast
        
        room.mu.Lock()
        for client := range room.clients {
            err := client.WriteMessage(pkg.TextMessage, message)
            if err != nil {
                client.Close()
                delete(room.clients, client)
            }
        }
        room.mu.Unlock()
    }
}

func (room *ChatRoom) Handler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Register client
    room.mu.Lock()
    room.clients[&conn] = true
    room.mu.Unlock()
    
    defer func() {
        room.mu.Lock()
        delete(room.clients, &conn)
        room.mu.Unlock()
        conn.Close()
    }()
    
    // Read messages
    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Broadcast to all clients
        room.broadcast <- data
    }
}

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    wsServer := pkg.NewWebSocketServer(nil)
    
    chatRoom := NewChatRoom()
    wsServer.RegisterRoute("/ws/chat", chatRoom.Handler, nil)
    
    router := app.Router()
    router.GET("/ws/chat", func(ctx pkg.Context) error {
        return wsServer.HandleUpgrade(ctx)
    })
    
    app.Listen(":8080")
}
```

### Authenticated WebSocket

```go
func main() {
    config := pkg.FrameworkConfig{
        SecurityConfig: pkg.SecurityConfig{
            JWTSecret: "your-secret-key",
        },
    }
    app, _ := pkg.New(config)
    
    authManager := pkg.NewAuthManager(/* config */)
    wsServer := pkg.NewWebSocketServer(authManager)
    
    // Private chat handler
    privateHandler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()
        
        userID := ctx.Request().UserID
        log.Printf("User %s connected", userID)
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }
            
            // Process authenticated message
            message := fmt.Sprintf("User %s: %s", userID, string(data))
            conn.WriteMessage(messageType, []byte(message))
        }
    }
    
    // Requires authentication and chat:write scope
    wsServer.RegisterAuthenticatedRoute("/ws/private", privateHandler, nil, 
        []string{"chat:write"})
    
    router := app.Router()
    router.GET("/ws/private", func(ctx pkg.Context) error {
        return wsServer.HandleUpgrade(ctx)
    })
    
    app.Listen(":8080")
}
```

### Multi-Tenant WebSocket

```go
func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    wsServer := pkg.NewWebSocketServer(nil)
    
    // Tenant-specific handler
    tenantHandler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()
        
        tenantID := ctx.Request().TenantID
        log.Printf("Connection for tenant: %s", tenantID)
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }
            
            // Process tenant-specific message
            message := fmt.Sprintf("[%s] %s", tenantID, string(data))
            conn.WriteMessage(messageType, []byte(message))
        }
    }
    
    // Register for different tenants
    wsServer.RegisterHostRoute("tenant1.example.com", "/ws/chat", tenantHandler, nil)
    wsServer.RegisterHostRoute("tenant2.example.com", "/ws/chat", tenantHandler, nil)
    
    router := app.Router()
    router.GET("/ws/chat", func(ctx pkg.Context) error {
        return wsServer.HandleUpgrade(ctx)
    })
    
    app.Listen(":8080")
}
```

### WebSocket with Middleware

```go
// Logging middleware
func wsLoggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    log.Printf("WebSocket connection started: %s", ctx.Request().RemoteAddr)
    
    err := next(ctx)
    
    duration := time.Since(start)
    log.Printf("WebSocket connection ended: %s (duration: %v)", 
        ctx.Request().RemoteAddr, duration)
    
    return err
}

func main() {
    app, _ := pkg.New(pkg.FrameworkConfig{})
    wsServer := pkg.NewWebSocketServer(nil)
    
    handler := func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()
        
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }
            conn.WriteMessage(messageType, data)
        }
    }
    
    // Register with middleware
    middleware := []pkg.MiddlewareFunc{wsLoggingMiddleware}
    wsServer.RegisterRoute("/ws/echo", handler, middleware)
    
    router := app.Router()
    router.GET("/ws/echo", func(ctx pkg.Context) error {
        return wsServer.HandleUpgrade(ctx)
    })
    
    app.Listen(":8080")
}
```

## Client-Side Connection

### JavaScript Example

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/echo');

// Connection opened
ws.addEventListener('open', (event) => {
    console.log('Connected to WebSocket');
    ws.send('Hello Server!');
});

// Listen for messages
ws.addEventListener('message', (event) => {
    console.log('Message from server:', event.data);
});

// Connection closed
ws.addEventListener('close', (event) => {
    console.log('Disconnected from WebSocket');
});

// Error handling
ws.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
});
```

### Authenticated Connection

```javascript
// Get access token
const token = localStorage.getItem('access_token');

// Connect with token in query parameter
const ws = new WebSocket(`ws://localhost:8080/ws/private?token=${token}`);

// Or with token in header (if supported by client)
// Note: WebSocket API doesn't support custom headers directly
// Use query parameter or subprotocol for authentication
```

## Best Practices

1. **Always Close Connections:** Use `defer conn.Close()` in handlers
2. **Handle Errors:** Check errors from ReadMessage and WriteMessage
3. **Use Goroutines:** Handle each connection in a separate goroutine
4. **Implement Ping/Pong:** Use ping/pong for connection health checks
5. **Set Timeouts:** Configure read/write deadlines
6. **Validate Messages:** Validate and sanitize incoming messages
7. **Rate Limiting:** Implement rate limiting for message sending
8. **Authentication:** Use authenticated routes for sensitive data
9. **Origin Checking:** Configure CheckOrigin for security
10. **Connection Limits:** Limit concurrent connections per user/tenant

## Security Considerations

1. **Origin Validation:** Always validate WebSocket origin
2. **Authentication:** Use access tokens for authenticated connections
3. **Authorization:** Check scopes/permissions before processing messages
4. **Input Validation:** Validate all incoming messages
5. **Rate Limiting:** Prevent message flooding
6. **Connection Limits:** Limit connections per IP/user
7. **TLS/SSL:** Use WSS (WebSocket Secure) in production
8. **CSRF Protection:** Validate tokens to prevent CSRF
9. **Message Size:** Limit maximum message size
10. **Timeout Configuration:** Set appropriate timeouts

## Troubleshooting

### Connection Upgrade Fails

**Problem:** WebSocket upgrade returns 400 or 403.

**Solutions:**
- Check Origin header validation
- Verify authentication token
- Ensure correct upgrade headers
- Check firewall/proxy configuration

### Messages Not Received

**Problem:** Messages sent but not received.

**Solutions:**
- Check message type (Text vs Binary)
- Verify connection is still open
- Check for errors in write operations
- Ensure proper goroutine handling

### Connection Drops

**Problem:** Connections close unexpectedly.

**Solutions:**
- Implement ping/pong heartbeat
- Check network stability
- Increase timeout values
- Handle errors properly

## See Also

- [Context API](context.md) - Context interface
- [Router API](router.md) - Route registration
- [Security API](security.md) - Authentication and authorization
- [WebSocket Guide](../guides/websockets.md) - WebSocket usage guide
- [Examples](../examples/websocket-chat.md) - Complete WebSocket examples
