# WebSocket Chat Example

The WebSocket example demonstrates real-time bidirectional communication using WebSockets with the Rockstar Web Framework. While the framework includes a simple echo server example in `full_featured_app.go`, this guide shows how to build a complete chat application with rooms, broadcasting, and connection management.

## What This Example Demonstrates

- **WebSocket connections** establishment and management
- **Real-time messaging** between clients
- **Message broadcasting** to multiple clients
- **Connection lifecycle** handling
- **Error handling** for WebSocket operations
- **Ping/pong** for connection health checks

## Prerequisites

- Go 1.25 or higher
- WebSocket client (browser, websocat, or wscat)

## Basic WebSocket Handler

From the full_featured_app.go example:

```go
func websocketHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    fmt.Println("WebSocket connection established")
    
    // Echo server - reads messages and sends them back
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            fmt.Printf("WebSocket read error: %v\n", err)
            return err
        }
        
        fmt.Printf("Received: %s\n", string(data))
        
        // Echo message back
        if err := conn.WriteMessage(messageType, data); err != nil {
            fmt.Printf("WebSocket write error: %v\n", err)
            return err
        }
    }
}
```

## Registering WebSocket Routes

```go
router := app.Router()
router.WebSocket("/ws", websocketHandler)
```

## Testing WebSocket Connections

### Using Browser JavaScript

```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <h1>WebSocket Echo Test</h1>
    <input type="text" id="message" placeholder="Enter message">
    <button onclick="sendMessage()">Send</button>
    <div id="output"></div>
    
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        
        ws.onopen = () => {
            console.log('Connected');
            addOutput('Connected to server');
        };
        
        ws.onmessage = (event) => {
            console.log('Received:', event.data);
            addOutput('Received: ' + event.data);
        };
        
        ws.onerror = (error) => {
            console.error('Error:', error);
            addOutput('Error: ' + error);
        };
        
        ws.onclose = () => {
            console.log('Disconnected');
            addOutput('Disconnected from server');
        };
        
        function sendMessage() {
            const input = document.getElementById('message');
            const message = input.value;
            ws.send(message);
            addOutput('Sent: ' + message);
            input.value = '';
        }
        
        function addOutput(text) {
            const output = document.getElementById('output');
            output.innerHTML += '<p>' + text + '</p>';
        }
    </script>
</body>
</html>
```

### Using websocat

```bash
# Install websocat
# macOS: brew install websocat
# Linux: cargo install websocat

# Connect to WebSocket
websocat ws://localhost:8080/ws

# Type messages and press Enter
Hello, WebSocket!
```

### Using wscat

```bash
# Install wscat
npm install -g wscat

# Connect to WebSocket
wscat -c ws://localhost:8080/ws

# Type messages
> Hello, WebSocket!
< Hello, WebSocket!
```

## Building a Chat Application

### Chat Room Manager

```go
type ChatRoom struct {
    clients   map[*pkg.WebSocketConnection]bool
    broadcast chan []byte
    register  chan *pkg.WebSocketConnection
    unregister chan *pkg.WebSocketConnection
    mu        sync.RWMutex
}

func NewChatRoom() *ChatRoom {
    room := &ChatRoom{
        clients:    make(map[*pkg.WebSocketConnection]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *pkg.WebSocketConnection),
        unregister: make(chan *pkg.WebSocketConnection),
    }
    
    go room.run()
    return room
}

func (r *ChatRoom) run() {
    for {
        select {
        case client := <-r.register:
            r.mu.Lock()
            r.clients[client] = true
            r.mu.Unlock()
            fmt.Printf("Client connected. Total: %d\n", len(r.clients))
            
        case client := <-r.unregister:
            r.mu.Lock()
            if _, ok := r.clients[client]; ok {
                delete(r.clients, client)
                client.Close()
            }
            r.mu.Unlock()
            fmt.Printf("Client disconnected. Total: %d\n", len(r.clients))
            
        case message := <-r.broadcast:
            r.mu.RLock()
            for client := range r.clients {
                err := client.WriteMessage(pkg.TextMessage, message)
                if err != nil {
                    fmt.Printf("Error broadcasting to client: %v\n", err)
                    r.unregister <- client
                }
            }
            r.mu.RUnlock()
        }
    }
}
```

### Chat Handler

```go
var chatRoom = NewChatRoom()

func chatHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Register client
    chatRoom.register <- conn
    defer func() {
        chatRoom.unregister <- conn
    }()
    
    // Read messages from client
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                fmt.Printf("WebSocket error: %v\n", err)
            }
            break
        }
        
        if messageType == pkg.TextMessage {
            // Broadcast message to all clients
            chatRoom.broadcast <- data
        }
    }
    
    return nil
}
```

### Message Types

```go
type ChatMessage struct {
    Type      string    `json:"type"`      // "join", "leave", "message"
    Username  string    `json:"username"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}

func handleChatMessage(conn pkg.WebSocketConnection, data []byte) error {
    var msg ChatMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return err
    }
    
    msg.Timestamp = time.Now()
    
    switch msg.Type {
    case "join":
        return handleJoin(conn, msg)
    case "leave":
        return handleLeave(conn, msg)
    case "message":
        return handleMessage(conn, msg)
    default:
        return fmt.Errorf("unknown message type: %s", msg.Type)
    }
}
```

## Connection Management

### Ping/Pong for Health Checks

```go
func websocketHandlerWithPing(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Set up ping/pong handlers
    conn.SetPongHandler(func(string) error {
        conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    // Start ping ticker
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    done := make(chan struct{})
    
    // Read messages
    go func() {
        defer close(done)
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                return
            }
            // Handle message
            handleMessage(message)
        }
    }()
    
    // Send pings
    for {
        select {
        case <-ticker.C:
            if err := conn.WriteMessage(pkg.PingMessage, nil); err != nil {
                return err
            }
        case <-done:
            return nil
        }
    }
}
```

### Graceful Disconnection

```go
func closeConnection(conn pkg.WebSocketConnection) {
    // Send close message
    conn.WriteMessage(
        pkg.CloseMessage,
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
    )
    
    // Wait briefly for client to close
    time.Sleep(time.Second)
    
    // Force close
    conn.Close()
}
```

## Production Considerations

### Authentication

Authenticate WebSocket connections:

```go
func authenticatedWebSocketHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    // Check authentication token from query parameter or header
    token := ctx.Query()["token"]
    if token == "" {
        conn.WriteMessage(pkg.CloseMessage, 
            websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Unauthorized"))
        return fmt.Errorf("unauthorized")
    }
    
    user, err := validateToken(token)
    if err != nil {
        return err
    }
    
    // Store user in connection context
    // ... handle authenticated connection
}
```

### Rate Limiting

Limit message rate per connection:

```go
type RateLimiter struct {
    tokens   int
    capacity int
    refill   time.Duration
    mu       sync.Mutex
}

func (rl *RateLimiter) Allow() bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    if rl.tokens > 0 {
        rl.tokens--
        return true
    }
    return false
}

func websocketHandlerWithRateLimit(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    limiter := &RateLimiter{
        tokens:   10,
        capacity: 10,
        refill:   time.Second,
    }
    
    // Refill tokens
    go func() {
        ticker := time.NewTicker(limiter.refill)
        defer ticker.Stop()
        for range ticker.C {
            limiter.mu.Lock()
            if limiter.tokens < limiter.capacity {
                limiter.tokens++
            }
            limiter.mu.Unlock()
        }
    }()
    
    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        if !limiter.Allow() {
            conn.WriteMessage(pkg.TextMessage, []byte("Rate limit exceeded"))
            continue
        }
        
        // Handle message
        handleMessage(data)
    }
}
```

### Message Validation

Validate message size and content:

```go
const maxMessageSize = 512 * 1024 // 512 KB

func websocketHandlerWithValidation(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    conn.SetReadLimit(maxMessageSize)
    
    for {
        messageType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        
        // Validate message type
        if messageType != pkg.TextMessage && messageType != pkg.BinaryMessage {
            continue
        }
        
        // Validate message content
        if !isValidMessage(data) {
            conn.WriteMessage(pkg.TextMessage, []byte("Invalid message"))
            continue
        }
        
        // Handle valid message
        handleMessage(data)
    }
}
```

## Common Issues

### "Connection closed unexpectedly"

**Solution**: Implement ping/pong for connection health checks

### "Message too large"

**Solution**: Set appropriate read limit with `conn.SetReadLimit()`

### "Too many connections"

**Solution**: Implement connection limits and cleanup

## Related Documentation

- [WebSocket Guide](../guides/websockets.md) - WebSocket patterns and best practices
- [Full Featured App](full-featured-app.md) - Complete application example
- [Security Guide](../guides/security.md) - Authentication and authorization

## Source Code

WebSocket examples are available in `examples/full_featured_app.go` in the repository.
