# Function Types Reference

Complete reference for all function type signatures in the Rockstar Web Framework.

## Overview

The framework uses function types extensively for handlers, middleware, callbacks, and hooks. This document provides detailed documentation for all function type signatures.

## Table of Contents

- [Core Handler Types](#core-handler-types)
- [Middleware Types](#middleware-types)
- [Pipeline Types](#pipeline-types)
- [Protocol Handler Types](#protocol-handler-types)
- [Plugin System Types](#plugin-system-types)
- [Template Types](#template-types)
- [Callback Types](#callback-types)

---

## Core Handler Types

### HandlerFunc

```go
type HandlerFunc func(ctx Context) error
```

The primary handler function type for HTTP routes.

**Parameters**:
- `ctx`: Request context providing access to all framework features

**Returns**:
- `error`: Error if handler fails (automatically converted to HTTP error response)

**Usage**:
```go
func myHandler(ctx pkg.Context) error {
    user := ctx.User()
    return ctx.JSON(200, map[string]interface{}{
        "message": "Hello " + user.Username,
    })
}

app.Router().GET("/profile", myHandler)
```

**Best Practices**:
- Return structured errors using `pkg.NewFrameworkError()`
- Use context methods for responses (`ctx.JSON()`, `ctx.HTML()`, etc.)
- Keep handlers focused on single responsibility
- Extract business logic to separate functions

**Common Patterns**:
```go
// Error handling
func handler(ctx pkg.Context) error {
    data, err := fetchData(ctx)
    if err != nil {
        return pkg.NewDatabaseError("Failed to fetch data", "query")
    }
    return ctx.JSON(200, data)
}

// Validation
func createHandler(ctx pkg.Context) error {
    var input CreateRequest
    if err := json.Unmarshal(ctx.Body(), &input); err != nil {
        return pkg.NewValidationError("Invalid JSON", "body")
    }
    
    if input.Name == "" {
        return pkg.NewMissingFieldError("name")
    }
    
    return ctx.JSON(201, input)
}

// Authentication check
func protectedHandler(ctx pkg.Context) error {
    if !ctx.IsAuthenticated() {
        return pkg.NewAuthenticationError("Login required")
    }
    return ctx.JSON(200, map[string]string{"status": "ok"})
}
```

---

### MiddlewareFunc

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

Middleware function type for request processing pipeline.

**Parameters**:
- `ctx`: Request context
- `next`: Next handler in the chain (must be called to continue)

**Returns**:
- `error`: Error if middleware fails

**Usage**:
```go
func loggingMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    start := time.Now()
    
    // Before handler
    ctx.Logger().Info("request started", "path", ctx.Request().Path)
    
    // Call next handler
    err := next(ctx)
    
    // After handler
    duration := time.Since(start)
    ctx.Logger().Info("request completed", 
        "path", ctx.Request().Path,
        "duration", duration,
    )
    
    return err
}

app.Router().GET("/api/users", handler, loggingMiddleware)
```

**Best Practices**:
- Always call `next(ctx)` unless intentionally stopping the chain
- Handle errors from `next(ctx)` appropriately
- Use defer for cleanup operations
- Keep middleware focused and composable

**Common Patterns**:
```go
// Authentication middleware
func authMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        return pkg.NewAuthenticationError("Missing token")
    }
    
    user, err := validateToken(token)
    if err != nil {
        return pkg.NewAuthenticationError("Invalid token")
    }
    
    ctx.Set("user", user)
    return next(ctx)
}

// Rate limiting middleware
func rateLimitMiddleware(limit int) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        key := "ratelimit:" + ctx.Request().RemoteAddr
        
        allowed, err := ctx.DB().CheckRateLimit(key, limit, time.Minute)
        if err != nil {
            return err
        }
        
        if !allowed {
            return pkg.NewRateLimitError("Rate limit exceeded", limit, "1m")
        }
        
        ctx.DB().IncrementRateLimit(key, time.Minute)
        return next(ctx)
    }
}

// Error recovery middleware
func recoveryMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    defer func() {
        if r := recover(); r != nil {
            ctx.Logger().Error("panic recovered", "panic", r)
            ctx.JSON(500, map[string]interface{}{
                "error": "Internal server error",
            })
        }
    }()
    
    return next(ctx)
}

// CORS middleware
func corsMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    ctx.SetHeader("Access-Control-Allow-Origin", "*")
    ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if ctx.Request().Method == "OPTIONS" {
        return ctx.String(204, "")
    }
    
    return next(ctx)
}
```

---

## Middleware Types

### ConditionFunc

```go
type ConditionFunc func(ctx Context) bool
```

Function type for conditional middleware execution.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `bool`: True if middleware should execute, false to skip

**Usage**:
```go
// Only apply to authenticated users
condition := func(ctx pkg.Context) bool {
    return ctx.IsAuthenticated()
}

config := pkg.MiddlewareConfig{
    Name: "premium-features",
    Handler: premiumMiddleware,
    Condition: condition,
}
```

**Common Patterns**:
```go
// Path-based condition
func pathMatches(pattern string) pkg.ConditionFunc {
    return func(ctx pkg.Context) bool {
        matched, _ := filepath.Match(pattern, ctx.Request().Path)
        return matched
    }
}

// Role-based condition
func hasRole(role string) pkg.ConditionFunc {
    return func(ctx pkg.Context) bool {
        user := ctx.User()
        if user == nil {
            return false
        }
        for _, r := range user.Roles {
            if r == role {
                return true
            }
        }
        return false
    }
}

// Time-based condition
func duringBusinessHours() pkg.ConditionFunc {
    return func(ctx pkg.Context) bool {
        hour := time.Now().Hour()
        return hour >= 9 && hour < 17
    }
}
```

---

## Pipeline Types

### PipelineFunc

```go
type PipelineFunc func(ctx Context) (PipelineResult, error)
```

Function type for pipeline stage processing.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `PipelineResult`: Control flow directive (Continue, Close, Retry)
- `error`: Error if stage fails

**PipelineResult Values**:
```go
const (
    PipelineResultContinue PipelineResult = iota // Continue to next stage
    PipelineResultClose                          // Stop pipeline execution
    PipelineResultRetry                          // Retry current stage
)
```

**Usage**:
```go
// Validation stage
func validateInput(ctx pkg.Context) (pkg.PipelineResult, error) {
    var input map[string]interface{}
    if err := json.Unmarshal(ctx.Body(), &input); err != nil {
        return pkg.PipelineResultClose, pkg.NewValidationError("Invalid JSON", "body")
    }
    
    if input["email"] == "" {
        return pkg.PipelineResultClose, pkg.NewMissingFieldError("email")
    }
    
    ctx.Set("validated_input", input)
    return pkg.PipelineResultContinue, nil
}

// Processing stage
func processData(ctx pkg.Context) (pkg.PipelineResult, error) {
    input := ctx.Get("validated_input").(map[string]interface{})
    
    result, err := performProcessing(input)
    if err != nil {
        // Retry on transient errors
        if isTransientError(err) {
            return pkg.PipelineResultRetry, err
        }
        return pkg.PipelineResultClose, err
    }
    
    ctx.Set("result", result)
    return pkg.PipelineResultContinue, nil
}

// Register pipeline
pipeline := pkg.PipelineConfig{
    Name: "data-processing",
    Stages: []pkg.PipelineStage{
        {Name: "validate", Handler: validateInput, Required: true},
        {Name: "process", Handler: processData, Retry: 3},
    },
}
```

**Best Practices**:
- Return `PipelineResultClose` for fatal errors
- Use `PipelineResultRetry` for transient failures
- Store intermediate results in context
- Keep stages focused and testable

**Common Patterns**:
```go
// Conditional stage
func conditionalStage(ctx pkg.Context) (pkg.PipelineResult, error) {
    if ctx.Get("skip_processing") != nil {
        return pkg.PipelineResultContinue, nil
    }
    
    // Perform processing
    return pkg.PipelineResultContinue, nil
}

// Parallel processing stage
func parallelStage(ctx pkg.Context) (pkg.PipelineResult, error) {
    tasks := []func() error{
        func() error { return task1(ctx) },
        func() error { return task2(ctx) },
        func() error { return task3(ctx) },
    }
    
    var wg sync.WaitGroup
    errors := make(chan error, len(tasks))
    
    for _, task := range tasks {
        wg.Add(1)
        go func(t func() error) {
            defer wg.Done()
            if err := t(); err != nil {
                errors <- err
            }
        }(task)
    }
    
    wg.Wait()
    close(errors)
    
    if len(errors) > 0 {
        return pkg.PipelineResultClose, <-errors
    }
    
    return pkg.PipelineResultContinue, nil
}
```

---

## Protocol Handler Types

### WebSocketHandler

```go
type WebSocketHandler func(ctx Context, conn WebSocketConnection) error
```

Handler function for WebSocket connections.

**Parameters**:
- `ctx`: Request context
- `conn`: WebSocket connection interface

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func chatHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    defer conn.Close()
    
    // Send welcome message
    conn.WriteMessage(websocket.TextMessage, []byte("Welcome to chat!"))
    
    // Message loop
    for {
        msgType, data, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                return nil
            }
            return err
        }
        
        // Echo message back
        if err := conn.WriteMessage(msgType, data); err != nil {
            return err
        }
    }
}

app.Router().WebSocket("/chat", chatHandler)
```

**Best Practices**:
- Always close connection with `defer conn.Close()`
- Handle close errors gracefully
- Implement ping/pong for keepalive
- Use goroutines for concurrent read/write

**Common Patterns**:
```go
// Broadcast pattern
type Hub struct {
    clients map[*pkg.WebSocketConnection]bool
    broadcast chan []byte
    mu sync.RWMutex
}

func (h *Hub) wsHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    h.mu.Lock()
    h.clients[&conn] = true
    h.mu.Unlock()
    
    defer func() {
        h.mu.Lock()
        delete(h.clients, &conn)
        h.mu.Unlock()
        conn.Close()
    }()
    
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        h.broadcast <- msg
    }
}

// Authenticated WebSocket
func authWSHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
    if !ctx.IsAuthenticated() {
        conn.WriteMessage(websocket.TextMessage, []byte("Authentication required"))
        conn.Close()
        return pkg.NewAuthenticationError("Not authenticated")
    }
    
    userID := ctx.User().ID
    // Handle authenticated connection
    return handleUserConnection(userID, conn)
}
```

---

### RESTHandler

```go
type RESTHandler func(ctx Context) (interface{}, error)
```

Handler function for REST API endpoints.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `interface{}`: Response data (automatically serialized to JSON)
- `error`: Error if handler fails

**Usage**:
```go
func listUsers(ctx pkg.Context) (interface{}, error) {
    users, err := getUsersFromDB(ctx)
    if err != nil {
        return nil, err
    }
    
    return users, nil
}

app.REST().GET("/api/users", listUsers, pkg.RESTRouteConfig{
    RateLimit: 100,
    RequireAuth: true,
})
```

**Common Patterns**:
```go
// CRUD operations
func createUser(ctx pkg.Context) (interface{}, error) {
    var user User
    if err := json.Unmarshal(ctx.Body(), &user); err != nil {
        return nil, pkg.NewValidationError("Invalid input", "body")
    }
    
    if err := ctx.DB().SaveUser(&user); err != nil {
        return nil, err
    }
    
    return user, nil
}

func getUser(ctx pkg.Context) (interface{}, error) {
    id := ctx.Param("id")
    user, err := ctx.DB().LoadUser(id)
    if err != nil {
        return nil, pkg.NewNotFoundError("User")
    }
    return user, nil
}

func updateUser(ctx pkg.Context) (interface{}, error) {
    id := ctx.Param("id")
    var updates User
    if err := json.Unmarshal(ctx.Body(), &updates); err != nil {
        return nil, pkg.NewValidationError("Invalid input", "body")
    }
    
    updates.ID = id
    if err := ctx.DB().UpdateUser(&updates); err != nil {
        return nil, err
    }
    
    return updates, nil
}

func deleteUser(ctx pkg.Context) (interface{}, error) {
    id := ctx.Param("id")
    if err := ctx.DB().DeleteUser(id); err != nil {
        return nil, err
    }
    return map[string]string{"status": "deleted"}, nil
}
```

---

### GraphQLHandler

```go
type GraphQLHandler func(ctx Context) error
```

Handler function for GraphQL endpoints.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func graphqlHandler(ctx pkg.Context) error {
    var request struct {
        Query     string                 `json:"query"`
        Variables map[string]interface{} `json:"variables"`
    }
    
    if err := json.Unmarshal(ctx.Body(), &request); err != nil {
        return pkg.NewValidationError("Invalid GraphQL request", "body")
    }
    
    result, err := schema.Execute(request.Query, request.Variables)
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, result)
}
```

---

### GRPCHandler

```go
type GRPCHandler func(ctx Context) error
```

Handler function for gRPC endpoints.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func grpcHandler(ctx pkg.Context) error {
    // gRPC request handling
    return nil
}
```

---

### GRPCUnaryHandler

```go
type GRPCUnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)
```

Handler function for unary gRPC calls.

**Parameters**:
- `ctx`: Go context
- `req`: Request message

**Returns**:
- `interface{}`: Response message
- `error`: Error if handler fails

**Usage**:
```go
func getUser(ctx context.Context, req interface{}) (interface{}, error) {
    userReq := req.(*pb.GetUserRequest)
    
    user, err := fetchUser(userReq.Id)
    if err != nil {
        return nil, err
    }
    
    return &pb.User{
        Id: user.ID,
        Name: user.Name,
    }, nil
}
```

---

### GRPCStreamHandler

```go
type GRPCStreamHandler func(srv interface{}, stream GRPCServerStream) error
```

Handler function for streaming gRPC calls.

**Parameters**:
- `srv`: Service implementation
- `stream`: Server stream interface

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func streamUsers(srv interface{}, stream pkg.GRPCServerStream) error {
    users, err := getAllUsers()
    if err != nil {
        return err
    }
    
    for _, user := range users {
        if err := stream.Send(&pb.User{
            Id: user.ID,
            Name: user.Name,
        }); err != nil {
            return err
        }
    }
    
    return nil
}
```

---

### SOAPHandler

```go
type SOAPHandler func(ctx Context) error
```

Handler function for SOAP endpoints.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func soapHandler(ctx pkg.Context) error {
    // Parse SOAP envelope
    var envelope pkg.SOAPEnvelope
    if err := xml.Unmarshal(ctx.Body(), &envelope); err != nil {
        return pkg.NewValidationError("Invalid SOAP request", "body")
    }
    
    // Process request
    response := processSOAPRequest(envelope)
    
    return ctx.XML(200, response)
}
```

---

## Plugin System Types

### HookHandler

```go
type HookHandler func(ctx HookContext) error
```

Handler function for plugin lifecycle hooks.

**Parameters**:
- `ctx`: Hook context providing hook metadata and data passing

**Returns**:
- `error`: Error if hook fails

**Usage**:
```go
func onStartup(ctx pkg.HookContext) error {
    ctx.Get("logger").(pkg.Logger).Info("Plugin starting up")
    
    // Initialize plugin resources
    if err := initializeResources(); err != nil {
        return err
    }
    
    return nil
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    return ctx.RegisterHook(pkg.HookTypeStartup, 100, onStartup)
}
```

**Hook Types**:
- `HookTypeStartup`: Called when plugin starts
- `HookTypeShutdown`: Called when plugin stops
- `HookTypePreRequest`: Called before request processing
- `HookTypePostRequest`: Called after request processing
- `HookTypePreResponse`: Called before response sent
- `HookTypePostResponse`: Called after response sent
- `HookTypeError`: Called when error occurs

**Common Patterns**:
```go
// Request logging hook
func logRequest(ctx pkg.HookContext) error {
    reqCtx := ctx.Context()
    if reqCtx == nil {
        return nil
    }
    
    ctx.Get("logger").(pkg.Logger).Info("request received",
        "path", reqCtx.Request().Path,
        "method", reqCtx.Request().Method,
    )
    
    return nil
}

// Authentication hook
func authHook(ctx pkg.HookContext) error {
    reqCtx := ctx.Context()
    if reqCtx == nil {
        return nil
    }
    
    if !reqCtx.IsAuthenticated() {
        ctx.Skip() // Skip remaining hooks
        return pkg.NewAuthenticationError("Authentication required")
    }
    
    return nil
}

// Metrics collection hook
func metricsHook(ctx pkg.HookContext) error {
    reqCtx := ctx.Context()
    if reqCtx == nil {
        return nil
    }
    
    start := ctx.Get("start_time").(time.Time)
    duration := time.Since(start)
    
    reqCtx.Metrics().RecordTiming("request_duration", duration, nil)
    
    return nil
}
```

---

### EventHandler

```go
type EventHandler func(event Event) error
```

Handler function for plugin events.

**Parameters**:
- `event`: Event containing name, data, source, and timestamp

**Returns**:
- `error`: Error if handler fails

**Usage**:
```go
func onUserCreated(event pkg.Event) error {
    user := event.Data.(*User)
    
    // Send welcome email
    if err := sendWelcomeEmail(user); err != nil {
        return err
    }
    
    return nil
}

func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    return ctx.SubscribeEvent("user.created", onUserCreated)
}
```

**Common Patterns**:
```go
// Type-safe event handling
func handleTypedEvent(event pkg.Event) error {
    switch event.Name {
    case "user.created":
        user, ok := event.Data.(*User)
        if !ok {
            return fmt.Errorf("invalid event data type")
        }
        return handleUserCreated(user)
        
    case "user.updated":
        user, ok := event.Data.(*User)
        if !ok {
            return fmt.Errorf("invalid event data type")
        }
        return handleUserUpdated(user)
        
    default:
        return fmt.Errorf("unknown event: %s", event.Name)
    }
}

// Async event processing
func asyncEventHandler(event pkg.Event) error {
    go func() {
        if err := processEvent(event); err != nil {
            log.Printf("Error processing event: %v", err)
        }
    }()
    return nil
}

// Event filtering
func filteredEventHandler(event pkg.Event) error {
    // Only process events from specific source
    if event.Source != "user-service" {
        return nil
    }
    
    return processEvent(event)
}
```

---

## Template Types

### ViewFunc

```go
type ViewFunc func(ctx Context) ResponseFunc
```

Function that returns a response function for template rendering.

**Parameters**:
- `ctx`: Request context

**Returns**:
- `ResponseFunc`: Function that writes the response

**Usage**:
```go
func userProfileView(ctx pkg.Context) pkg.ResponseFunc {
    user := ctx.User()
    
    return func() error {
        return ctx.HTML(200, "profile.html", map[string]interface{}{
            "user": user,
        })
    }
}
```

---

### ResponseFunc

```go
type ResponseFunc func() error
```

Function that writes an HTTP response.

**Returns**:
- `error`: Error if response writing fails

**Usage**:
```go
func createResponse() pkg.ResponseFunc {
    return func() error {
        // Write response
        return nil
    }
}
```

---

## Callback Types

### ErrorHandlerCallback

```go
type ErrorHandlerCallback func(ctx Context, err error) error
```

Callback function for custom error handling.

**Parameters**:
- `ctx`: Request context
- `err`: Error that occurred

**Returns**:
- `error`: Transformed or wrapped error

**Usage**:
```go
func customErrorHandler(ctx pkg.Context, err error) error {
    // Log error
    ctx.Logger().Error("request failed", "error", err)
    
    // Transform error
    if dbErr, ok := err.(*DatabaseError); ok {
        return pkg.NewInternalError("Database temporarily unavailable")
    }
    
    return err
}
```

---

### RecoveryCallback

```go
type RecoveryCallback func(ctx Context, recovered interface{}) error
```

Callback function for panic recovery.

**Parameters**:
- `ctx`: Request context
- `recovered`: Recovered panic value

**Returns**:
- `error`: Error to return

**Usage**:
```go
func panicRecovery(ctx pkg.Context, recovered interface{}) error {
    ctx.Logger().Error("panic recovered",
        "panic", recovered,
        "stack", string(debug.Stack()),
    )
    
    return pkg.NewInternalError("Internal server error")
}
```

---

### ConfigChangeCallback

```go
type ConfigChangeCallback func()
```

Callback function for configuration changes.

**Usage**:
```go
func onConfigChange() {
    log.Println("Configuration reloaded")
    // Reinitialize components
}

config.Watch(onConfigChange)
```

---

## Best Practices

### 1. Error Handling

Always return structured errors:
```go
func handler(ctx pkg.Context) error {
    if err := operation(); err != nil {
        return pkg.NewInternalError("Operation failed").WithCause(err)
    }
    return nil
}
```

### 2. Context Usage

Use context for data passing:
```go
func middleware(ctx pkg.Context, next pkg.HandlerFunc) error {
    ctx.Set("start_time", time.Now())
    return next(ctx)
}

func handler(ctx pkg.Context) error {
    start := ctx.Get("start_time").(time.Time)
    duration := time.Since(start)
    return ctx.JSON(200, map[string]interface{}{"duration": duration})
}
```

### 3. Resource Cleanup

Use defer for cleanup:
```go
func handler(ctx pkg.Context) error {
    conn, err := openConnection()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    // Use connection
    return nil
}
```

### 4. Composability

Create reusable function factories:
```go
func requireRole(role string) pkg.MiddlewareFunc {
    return func(ctx pkg.Context, next pkg.HandlerFunc) error {
        if !hasRole(ctx.User(), role) {
            return pkg.NewAuthorizationError("Insufficient permissions")
        }
        return next(ctx)
    }
}

app.Router().GET("/admin", handler, requireRole("admin"))
```

### 5. Type Safety

Use type assertions safely:
```go
func handler(ctx pkg.Context) error {
    value := ctx.Get("key")
    if value == nil {
        return pkg.NewInternalError("Missing value")
    }
    
    typed, ok := value.(string)
    if !ok {
        return pkg.NewInternalError("Invalid value type")
    }
    
    return ctx.String(200, typed)
}
```

---

## See Also

- [Context API](context.md) - Request context interface
- [Router API](router.md) - Routing and handlers
- [Middleware Engine API](pipeline-middleware-engine.md) - Middleware system
- [Plugin API](plugins.md) - Plugin system
- [Error Codes](error-codes.md) - Error handling

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0  
**Total Function Types**: 15+
