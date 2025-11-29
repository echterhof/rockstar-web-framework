# WebSockets

## Overview

The Rockstar Web Framework provides comprehensive WebSocket support for building real-time, bidirectional communication applications. WebSockets enable persistent connections between clients and servers, allowing instant data exchange without the overhead of HTTP polling.

**When to use WebSockets:**
- Real-time chat applications
- Live notifications and updates
- Collaborative editing tools
- Live dashboards and monitoring
- Multiplayer games
- Streaming data applications

**Key benefits:**
- **Bidirectional Communication**: Full-duplex communication channel
- **Low Latency**: Instant message delivery without polling overhead
- **Efficient**: Single persistent connection vs. multiple HTTP requests
- **Authentication Support**: Secure WebSocket connections with token-based auth
- **Host-Based Routing**: Multi-tenant WebSocket support
- **Automatic Ping/Pong**: Built-in connection health monitoring

## Quick Start

Here's a minimal WebSocket example:

```go
package main

import (
    "log"
    "time"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            EnableHTTP1:  true,
        },
    }

    app, err := pkg.New(config)
    if err != nil {
        log.Fatal(err)
    }

    router := app.Router()

    // Register WebSocket route
    router.WebSocket("/ws", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()

        // Send welcome message
        err := conn.WriteMessage(pkg.TextMessage, []byte("Welcome to WebSocket!"))
        if err != nil {
            return err
        }

        // Read messages in a loop
        for {
            messageType, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }

            // Echo message back
            err = conn.WriteMessage(messageType, data)
            if err != nil {
                return err
            }
        }
    })

    log.Fatal(app.Listen(":8080"))
}
```

## Configuration

### WebSocket Upgrader

Customize the WebSocket upgrader:

```go
import "github.com/gorilla/websocket"

upgrader := websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // Allow specific origins
        origin := r.Header.Get("Origin")
        return origin == "https://example.com"
    },
}

// Set custom upgrader
wsServer := app.WebSocketServer()
wsServer.SetUpgrader(upgrader)
```


### Message Types

WebSocket supports different message types:

```go
const (
    TextMessage   = 1  // UTF-8 text message
    BinaryMessage = 2  // Binary data message
    CloseMessage  = 8  // Connection close message
    PingMessage   = 9  // Ping message
    PongMessage   = 10 // Pong message
)
```

## Usage

### Basic WebSocket Handler

Create a simple WebSocket handler:

```go
router.WebSocket("/ws/echo", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    for {
        // Read message
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            // Connection closed or error
            return err
        }

        // Process message
        response := processMessage(data)

        // Send response
        err = conn.WriteMessage(messageType, response)
        if err != nil {
            return err
        }
    }
})
```

### Sending Messages

Send different types of messages:

```go
// Send text message
err := conn.WriteMessage(pkg.TextMessage, []byte("Hello, WebSocket!"))

// Send binary message
binaryData := []byte{0x01, 0x02, 0x03}
err = conn.WriteMessage(pkg.BinaryMessage, binaryData)

// Send JSON
import "encoding/json"

data := map[string]interface{}{
    "type": "notification",
    "message": "New message received",
}
jsonData, _ := json.Marshal(data)
err = conn.WriteMessage(pkg.TextMessage, jsonData)
```

### Reading Messages

Read and process incoming messages:

```go
for {
    messageType, data, err := conn.ReadMessage()
    if err != nil {
        // Handle error (connection closed, timeout, etc.)
        log.Printf("Read error: %v", err)
        return err
    }

    switch messageType {
    case pkg.TextMessage:
        // Handle text message
        text := string(data)
        log.Printf("Received text: %s", text)

    case pkg.BinaryMessage:
        // Handle binary message
        log.Printf("Received binary data: %d bytes", len(data))

    case pkg.CloseMessage:
        // Connection closing
        return nil
    }
}
```

### Broadcasting Messages

Broadcast to multiple connections:

```go
type ChatRoom struct {
    connections map[*pkg.WebSocketConnection]bool
    broadcast   chan []byte
    register    chan *pkg.WebSocketConnection
    unregister  chan *pkg.WebSocketConnection
    mu          sync.RWMutex
}

func NewChatRoom() *ChatRoom {
    room := &ChatRoom{
        connections: make(map[*pkg.WebSocketConnection]bool),
        broadcast:   make(chan []byte),
        register:    make(chan *pkg.WebSocketConnection),
        unregister:  make(chan *pkg.WebSocketConnection),
    }
    go room.run()
    return room
}

func (r *ChatRoom) run() {
    for {
        select {
        case conn := <-r.register:
            r.mu.Lock()
            r.connections[conn] = true
            r.mu.Unlock()

        case conn := <-r.unregister:
            r.mu.Lock()
            if _, ok := r.connections[conn]; ok {
                delete(r.connections, conn)
                conn.Close()
            }
            r.mu.Unlock()

        case message := <-r.broadcast:
            r.mu.RLock()
            for conn := range r.connections {
                err := conn.WriteMessage(pkg.TextMessage, message)
                if err != nil {
                    r.unregister <- conn
                }
            }
            r.mu.RUnlock()
        }
    }
}

// Usage
chatRoom := NewChatRoom()

router.WebSocket("/ws/chat", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer func() {
        chatRoom.unregister <- &conn
    }()

    chatRoom.register <- &conn

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        // Broadcast to all connections
        chatRoom.broadcast <- data
    }
})
```


### Connection Management

Manage WebSocket connections:

```go
router.WebSocket("/ws/managed", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Get connection info
    remoteAddr := conn.RemoteAddr()
    localAddr := conn.LocalAddr()
    log.Printf("Connection from %s to %s", remoteAddr, localAddr)

    // Set up cleanup
    defer func() {
        log.Printf("Closing connection from %s", remoteAddr)
        conn.Close()
    }()

    // Handle messages
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        // Process and respond
        response := processMessage(data)
        err = conn.WriteMessage(messageType, response)
        if err != nil {
            return err
        }
    }
})
```

### Authenticated WebSockets

Secure WebSocket connections with authentication:

```go
// Register authenticated WebSocket route
router.WebSocketAuthenticated("/ws/secure", 
    func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        defer conn.Close()

        // User is authenticated - access user info
        userID := ctx.Request().UserID
        tenantID := ctx.Request().TenantID

        log.Printf("Authenticated WebSocket: user=%s, tenant=%s", userID, tenantID)

        // Send personalized welcome
        welcome := fmt.Sprintf("Welcome, user %s!", userID)
        conn.WriteMessage(pkg.TextMessage, []byte(welcome))

        // Handle messages
        for {
            _, data, err := conn.ReadMessage()
            if err != nil {
                return err
            }

            // Process authenticated request
            response := processAuthenticatedMessage(userID, tenantID, data)
            conn.WriteMessage(pkg.TextMessage, response)
        }
    },
    []string{"chat:read", "chat:write"}, // Required scopes
)
```

Client authentication with token:

```javascript
// JavaScript client
const token = "your-access-token";

// Option 1: Token in query parameter
const ws = new WebSocket(`ws://localhost:8080/ws/secure?token=${token}`);

// Option 2: Token in subprotocol (requires custom upgrader)
const ws = new WebSocket('ws://localhost:8080/ws/secure', [token]);
```

### Host-Based WebSocket Routing

Support multi-tenant WebSockets:

```go
wsServer := app.WebSocketServer()

// Register WebSocket for specific host
wsServer.RegisterHostRoute("tenant1.example.com", "/ws/chat",
    func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        tenantID := ctx.Request().TenantID
        log.Printf("WebSocket for tenant: %s", tenantID)

        // Tenant-specific logic
        return handleTenantWebSocket(tenantID, conn)
    },
    []pkg.MiddlewareFunc{},
)

// Register authenticated host route
wsServer.RegisterAuthenticatedHostRoute("tenant2.example.com", "/ws/secure",
    func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        // Authenticated and tenant-scoped
        return handleSecureWebSocket(ctx, conn)
    },
    []pkg.MiddlewareFunc{},
    []string{"websocket:access"},
)
```

## Real-Time Communication Patterns

### Request-Response Pattern

Implement request-response over WebSocket:

```go
type Message struct {
    ID      string                 `json:"id"`
    Type    string                 `json:"type"`
    Payload map[string]interface{} `json:"payload"`
}

router.WebSocket("/ws/rpc", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        var msg Message
        if err := json.Unmarshal(data, &msg); err != nil {
            continue
        }

        // Process request
        var response Message
        switch msg.Type {
        case "get_user":
            response = Message{
                ID:   msg.ID,
                Type: "user_data",
                Payload: map[string]interface{}{
                    "user": getUserData(msg.Payload["user_id"]),
                },
            }
        case "update_status":
            updateStatus(msg.Payload)
            response = Message{
                ID:   msg.ID,
                Type: "status_updated",
                Payload: map[string]interface{}{
                    "success": true,
                },
            }
        }

        // Send response
        responseData, _ := json.Marshal(response)
        conn.WriteMessage(pkg.TextMessage, responseData)
    }
})
```

### Publish-Subscribe Pattern

Implement pub/sub over WebSocket:

```go
type PubSub struct {
    subscribers map[string]map[*pkg.WebSocketConnection]bool
    mu          sync.RWMutex
}

func NewPubSub() *PubSub {
    return &PubSub{
        subscribers: make(map[string]map[*pkg.WebSocketConnection]bool),
    }
}

func (ps *PubSub) Subscribe(topic string, conn *pkg.WebSocketConnection) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.subscribers[topic] == nil {
        ps.subscribers[topic] = make(map[*pkg.WebSocketConnection]bool)
    }
    ps.subscribers[topic][conn] = true
}

func (ps *PubSub) Unsubscribe(topic string, conn *pkg.WebSocketConnection) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if subs, ok := ps.subscribers[topic]; ok {
        delete(subs, conn)
    }
}

func (ps *PubSub) Publish(topic string, message []byte) {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    if subs, ok := ps.subscribers[topic]; ok {
        for conn := range subs {
            conn.WriteMessage(pkg.TextMessage, message)
        }
    }
}

// Usage
pubsub := NewPubSub()

router.WebSocket("/ws/pubsub", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        var msg map[string]interface{}
        json.Unmarshal(data, &msg)

        switch msg["action"] {
        case "subscribe":
            topic := msg["topic"].(string)
            pubsub.Subscribe(topic, &conn)

        case "unsubscribe":
            topic := msg["topic"].(string)
            pubsub.Unsubscribe(topic, &conn)

        case "publish":
            topic := msg["topic"].(string)
            message := msg["message"].(string)
            pubsub.Publish(topic, []byte(message))
        }
    }
})
```


### Streaming Data Pattern

Stream continuous data to clients:

```go
router.WebSocket("/ws/stream", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Generate and send data
            data := map[string]interface{}{
                "timestamp": time.Now().Unix(),
                "value":     generateValue(),
            }

            jsonData, _ := json.Marshal(data)
            err := conn.WriteMessage(pkg.TextMessage, jsonData)
            if err != nil {
                return err
            }

        case <-ctx.Request().Context().Done():
            // Request cancelled
            return nil
        }
    }
})
```

## Integration

### WebSockets with Middleware

Apply middleware to WebSocket routes:

```go
// Logging middleware
func wsLoggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    log.Printf("WebSocket connection from %s", ctx.Request().RemoteAddr)
    return next(ctx)
}

// Rate limiting middleware
func wsRateLimitMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    // Check rate limit
    if !checkRateLimit(ctx.Request().RemoteAddr) {
        return ctx.JSON(429, map[string]interface{}{
            "error": "Rate limit exceeded",
        })
    }
    return next(ctx)
}

// Apply middleware to WebSocket route
router.WebSocket("/ws/chat", 
    func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
        // WebSocket handler
        return handleChat(ctx, conn)
    },
    wsLoggingMiddleware,
    wsRateLimitMiddleware,
)
```

### WebSockets with Sessions

Use sessions with WebSocket connections:

```go
router.WebSocket("/ws/session", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    // Get session
    session, err := ctx.Session().GetSessionFromCookie(ctx)
    if err != nil {
        return err
    }

    // Get user from session
    userID := session.Get("user_id")
    if userID == nil {
        return fmt.Errorf("not authenticated")
    }

    log.Printf("WebSocket for user: %v", userID)

    // Handle messages
    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        // Process with session context
        response := processWithSession(session, data)
        conn.WriteMessage(pkg.TextMessage, response)
    }
})
```

### WebSockets with Database

Access database from WebSocket handlers:

```go
router.WebSocket("/ws/data", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    db := ctx.Database()

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        var query map[string]interface{}
        json.Unmarshal(data, &query)

        // Query database
        rows, err := db.Query("SELECT * FROM items WHERE category = ?", 
            query["category"])
        if err != nil {
            continue
        }

        // Send results
        var items []Item
        for rows.Next() {
            var item Item
            rows.Scan(&item.ID, &item.Name, &item.Category)
            items = append(items, item)
        }
        rows.Close()

        response, _ := json.Marshal(items)
        conn.WriteMessage(pkg.TextMessage, response)
    }
})
```

## Best Practices

### Error Handling

Handle errors gracefully:

```go
router.WebSocket("/ws/robust", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic: %v", r)
        }
        conn.Close()
    }()

    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            // Check if connection closed normally
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                log.Println("Connection closed normally")
                return nil
            }

            // Log unexpected errors
            log.Printf("Read error: %v", err)
            return err
        }

        // Process with error handling
        response, err := processMessage(data)
        if err != nil {
            // Send error message to client
            errorMsg := map[string]interface{}{
                "error": err.Error(),
            }
            errorData, _ := json.Marshal(errorMsg)
            conn.WriteMessage(pkg.TextMessage, errorData)
            continue
        }

        // Send successful response
        err = conn.WriteMessage(messageType, response)
        if err != nil {
            log.Printf("Write error: %v", err)
            return err
        }
    }
})
```

### Connection Timeouts

Implement connection timeouts:

```go
router.WebSocket("/ws/timeout", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    // Set read deadline
    timeout := 60 * time.Second
    deadline := time.Now().Add(timeout)

    for {
        // Update deadline before each read
        conn.SetReadDeadline(deadline)

        messageType, data, err := conn.ReadMessage()
        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                log.Println("Connection timeout")
            }
            return err
        }

        // Reset deadline on successful read
        deadline = time.Now().Add(timeout)

        // Process message
        response := processMessage(data)
        conn.WriteMessage(messageType, response)
    }
})
```

### Message Validation

Validate incoming messages:

```go
type MessageValidator struct {
    MaxSize int
}

func (v *MessageValidator) Validate(data []byte) error {
    if len(data) > v.MaxSize {
        return fmt.Errorf("message too large: %d bytes (max: %d)", 
            len(data), v.MaxSize)
    }

    // Validate JSON structure
    var msg map[string]interface{}
    if err := json.Unmarshal(data, &msg); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    // Validate required fields
    if _, ok := msg["type"]; !ok {
        return fmt.Errorf("missing required field: type")
    }

    return nil
}

validator := &MessageValidator{MaxSize: 1024 * 1024} // 1MB

router.WebSocket("/ws/validated", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        // Validate message
        if err := validator.Validate(data); err != nil {
            errorMsg := map[string]interface{}{
                "error": err.Error(),
            }
            errorData, _ := json.Marshal(errorMsg)
            conn.WriteMessage(pkg.TextMessage, errorData)
            continue
        }

        // Process valid message
        response := processMessage(data)
        conn.WriteMessage(pkg.TextMessage, response)
    }
})
```


### Resource Management

Manage resources efficiently:

```go
type ConnectionPool struct {
    connections map[string]*pkg.WebSocketConnection
    mu          sync.RWMutex
    maxConns    int
}

func NewConnectionPool(maxConns int) *ConnectionPool {
    return &ConnectionPool{
        connections: make(map[string]*pkg.WebSocketConnection),
        maxConns:    maxConns,
    }
}

func (cp *ConnectionPool) Add(id string, conn *pkg.WebSocketConnection) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if len(cp.connections) >= cp.maxConns {
        return fmt.Errorf("connection pool full")
    }

    cp.connections[id] = conn
    return nil
}

func (cp *ConnectionPool) Remove(id string) {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if conn, ok := cp.connections[id]; ok {
        conn.Close()
        delete(cp.connections, id)
    }
}

func (cp *ConnectionPool) Count() int {
    cp.mu.RLock()
    defer cp.mu.RUnlock()
    return len(cp.connections)
}

// Usage
pool := NewConnectionPool(1000)

router.WebSocket("/ws/pooled", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    connID := generateConnectionID()

    // Add to pool
    if err := pool.Add(connID, &conn); err != nil {
        return err
    }

    defer pool.Remove(connID)

    // Handle messages
    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }

        response := processMessage(data)
        conn.WriteMessage(pkg.TextMessage, response)
    }
})
```

### Monitoring and Metrics

Track WebSocket metrics:

```go
type WebSocketMetrics struct {
    ActiveConnections int64
    TotalMessages     int64
    TotalErrors       int64
    mu                sync.RWMutex
}

func (m *WebSocketMetrics) IncrementConnections() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ActiveConnections++
}

func (m *WebSocketMetrics) DecrementConnections() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ActiveConnections--
}

func (m *WebSocketMetrics) IncrementMessages() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.TotalMessages++
}

func (m *WebSocketMetrics) IncrementErrors() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.TotalErrors++
}

metrics := &WebSocketMetrics{}

router.WebSocket("/ws/monitored", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    metrics.IncrementConnections()
    defer metrics.DecrementConnections()

    defer conn.Close()

    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            metrics.IncrementErrors()
            return err
        }

        metrics.IncrementMessages()

        response := processMessage(data)
        err = conn.WriteMessage(pkg.TextMessage, response)
        if err != nil {
            metrics.IncrementErrors()
            return err
        }
    }
})

// Expose metrics endpoint
router.GET("/metrics/websocket", func(ctx pkg.Context) error {
    metrics.mu.RLock()
    defer metrics.mu.RUnlock()

    return ctx.JSON(200, map[string]interface{}{
        "active_connections": metrics.ActiveConnections,
        "total_messages":     metrics.TotalMessages,
        "total_errors":       metrics.TotalErrors,
    })
})
```

## Client Examples

### JavaScript Client

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/chat');

// Connection opened
ws.addEventListener('open', (event) => {
    console.log('Connected to WebSocket');
    ws.send('Hello Server!');
});

// Listen for messages
ws.addEventListener('message', (event) => {
    console.log('Message from server:', event.data);
    
    // Parse JSON if needed
    try {
        const data = JSON.parse(event.data);
        console.log('Parsed data:', data);
    } catch (e) {
        console.log('Text message:', event.data);
    }
});

// Connection closed
ws.addEventListener('close', (event) => {
    console.log('Disconnected from WebSocket');
});

// Error handling
ws.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
});

// Send JSON message
function sendMessage(type, payload) {
    const message = JSON.stringify({
        type: type,
        payload: payload
    });
    ws.send(message);
}

// Close connection
function disconnect() {
    ws.close();
}
```

### Go Client

```go
package main

import (
    "log"
    "github.com/gorilla/websocket"
)

func main() {
    // Connect to WebSocket
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws/chat", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Send message
    err = conn.WriteMessage(websocket.TextMessage, []byte("Hello Server!"))
    if err != nil {
        log.Fatal(err)
    }

    // Read messages
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            return
        }

        log.Printf("Received: %s", string(data))

        // Echo back
        err = conn.WriteMessage(messageType, data)
        if err != nil {
            log.Println("Write error:", err)
            return
        }
    }
}
```

## API Reference

See [Context API](../api/context.md) for WebSocket-related methods and [Framework API](../api/framework.md) for WebSocket configuration.

## Examples

See [WebSocket Chat Example](../examples/websocket-chat.md) for a complete working implementation.

## Troubleshooting

### Connection Upgrade Failed

**Problem**: WebSocket upgrade returns 400 or 403 error

**Solutions**:
- Verify `Upgrade: websocket` header is present
- Check `Connection: Upgrade` header
- Verify WebSocket route is registered correctly
- Check CORS/origin validation in upgrader
- Ensure HTTP/1.1 is enabled

### Messages Not Received

**Problem**: Client doesn't receive messages from server

**Solutions**:
- Check for write errors in server logs
- Verify connection is still open
- Check message type matches (text vs binary)
- Ensure no blocking operations in handler
- Verify client is listening for messages

### Connection Drops Frequently

**Problem**: WebSocket connections close unexpectedly

**Solutions**:
- Implement ping/pong heartbeat (built-in)
- Check for network timeouts
- Verify no proxy/firewall interference
- Handle connection errors gracefully
- Implement reconnection logic on client

### High Memory Usage

**Problem**: Server memory grows with WebSocket connections

**Solutions**:
- Implement connection pooling with limits
- Close connections properly in defer statements
- Clean up resources in connection handlers
- Monitor active connection count
- Implement connection timeouts

### Authentication Failures

**Problem**: Authenticated WebSocket routes return 401

**Solutions**:
- Verify token is passed correctly (query param or header)
- Check token format: `Bearer <token>` in Authorization header
- Ensure token is valid and not expired
- Verify required scopes are granted
- Check authentication manager is configured

## Related Documentation

- [Routing Guide](routing.md) - WebSocket routing
- [Security Guide](security.md) - Securing WebSocket connections
- [Middleware Guide](middleware.md) - WebSocket middleware
- [Architecture Overview](../architecture/overview.md) - Architecture patterns
