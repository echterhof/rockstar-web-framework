# API Reference

Complete API reference for the Rockstar Web Framework.

## Table of Contents

1. [Framework](#framework)
2. [Context](#context)
3. [Router](#router)
4. [Server](#server)
5. [Database](#database)
6. [Cache](#cache)
7. [Session](#session)
8. [Security](#security)
9. [Configuration](#configuration)
10. [Internationalization](#internationalization)

## Framework

The main framework struct that wires all components together.

### Creating a Framework Instance

```go
func New(config FrameworkConfig) (*Framework, error)
```

Creates a new framework instance with the given configuration.

**Parameters:**
- `config`: FrameworkConfig - Complete framework configuration

**Returns:**
- `*Framework`: Framework instance
- `error`: Error if initialization fails

**Example:**
```go
app, err := pkg.New(config)
if err != nil {
    log.Fatal(err)
}
```

### Framework Methods

#### Router()

```go
func (f *Framework) Router() RouterEngine
```

Returns the framework's router for route registration.

#### Use()

```go
func (f *Framework) Use(middleware ...MiddlewareFunc)
```

Adds global middleware to the framework.

**Parameters:**
- `middleware`: One or more middleware functions

**Example:**
```go
app.Use(loggingMiddleware, recoveryMiddleware)
```

#### Listen()

```go
func (f *Framework) Listen(addr string) error
```

Starts the framework server on the specified address.

**Parameters:**
- `addr`: Server address (e.g., ":8080")

**Returns:**
- `error`: Error if server fails to start

#### ListenTLS()

```go
func (f *Framework) ListenTLS(addr, certFile, keyFile string) error
```

Starts the framework server with TLS.

**Parameters:**
- `addr`: Server address
- `certFile`: Path to TLS certificate file
- `keyFile`: Path to TLS key file

#### ListenQUIC()

```go
func (f *Framework) ListenQUIC(addr, certFile, keyFile string) error
```

Starts the framework server with QUIC protocol.

#### Shutdown()

```go
func (f *Framework) Shutdown(timeout time.Duration) error
```

Gracefully shuts down the framework.

**Parameters:**
- `timeout`: Maximum time to wait for shutdown

#### RegisterStartupHook()

```go
func (f *Framework) RegisterStartupHook(hook func(ctx context.Context) error)
```

Registers a function to be called during startup.

#### RegisterShutdownHook()

```go
func (f *Framework) RegisterShutdownHook(hook func(ctx context.Context) error)
```

Registers a function to be called during graceful shutdown.

## Context

The unified request context providing access to all framework features.

### Request Data Methods

#### Request()

```go
func (ctx Context) Request() *Request
```

Returns the HTTP request object.

#### Response()

```go
func (ctx Context) Response() ResponseWriter
```

Returns the response writer.

#### Params()

```go
func (ctx Context) Params() map[string]string
```

Returns URL parameters (e.g., `/users/:id`).

**Example:**
```go
id := ctx.Params()["id"]
```

#### Query()

```go
func (ctx Context) Query() map[string]string
```

Returns query string parameters.

**Example:**
```go
search := ctx.Query()["q"]
```

#### Headers()

```go
func (ctx Context) Headers() map[string]string
```

Returns request headers.

#### Body()

```go
func (ctx Context) Body() []byte
```

Returns the request body as bytes.

### Framework Services

#### DB()

```go
func (ctx Context) DB() DatabaseManager
```

Returns the database manager.

#### Cache()

```go
func (ctx Context) Cache() CacheManager
```

Returns the cache manager.

#### Session()

```go
func (ctx Context) Session() SessionManager
```

Returns the session manager.

#### Config()

```go
func (ctx Context) Config() ConfigManager
```

Returns the configuration manager.

#### I18n()

```go
func (ctx Context) I18n() I18nManager
```

Returns the internationalization manager.

#### Logger()

```go
func (ctx Context) Logger() Logger
```

Returns the logger instance.

#### Metrics()

```go
func (ctx Context) Metrics() MetricsCollector
```

Returns the metrics collector.

### Response Methods

#### JSON()

```go
func (ctx Context) JSON(statusCode int, data interface{}) error
```

Sends a JSON response.

**Example:**
```go
return ctx.JSON(200, map[string]string{
    "message": "Success",
})
```

#### XML()

```go
func (ctx Context) XML(statusCode int, data interface{}) error
```

Sends an XML response.

#### HTML()

```go
func (ctx Context) HTML(statusCode int, template string, data interface{}) error
```

Renders an HTML template.

#### String()

```go
func (ctx Context) String(statusCode int, message string) error
```

Sends a plain text response.

#### Redirect()

```go
func (ctx Context) Redirect(statusCode int, url string) error
```

Redirects to another URL.

### Cookie Methods

#### SetCookie()

```go
func (ctx Context) SetCookie(cookie *Cookie) error
```

Sets a cookie in the response.

#### GetCookie()

```go
func (ctx Context) GetCookie(name string) (*Cookie, error)
```

Gets a cookie from the request.

### Header Methods

#### SetHeader()

```go
func (ctx Context) SetHeader(key, value string)
```

Sets a response header.

#### GetHeader()

```go
func (ctx Context) GetHeader(key string) string
```

Gets a request header.

### Form Methods

#### FormValue()

```go
func (ctx Context) FormValue(key string) string
```

Gets a form value from POST data.

#### FormFile()

```go
func (ctx Context) FormFile(key string) (*FormFile, error)
```

Gets an uploaded file from the request.

### Authentication Methods

#### User()

```go
func (ctx Context) User() *User
```

Returns the authenticated user.

#### Tenant()

```go
func (ctx Context) Tenant() *Tenant
```

Returns the current tenant (multi-tenancy).

#### IsAuthenticated()

```go
func (ctx Context) IsAuthenticated() bool
```

Checks if the user is authenticated.

#### IsAuthorized()

```go
func (ctx Context) IsAuthorized(resource, action string) bool
```

Checks if the user is authorized for an action.

## Router

The routing engine for mapping URLs to handlers.

### HTTP Method Routes

```go
func (r RouterEngine) GET(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) POST(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) PUT(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) DELETE(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) PATCH(path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouterEngine
```

**Example:**
```go
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

### Route Groups

```go
func (r RouterEngine) Group(prefix string, middleware ...MiddlewareFunc) RouterEngine
```

Creates a route group with a common prefix and middleware.

**Example:**
```go
api := router.Group("/api/v1", authMiddleware)
api.GET("/profile", getProfile)
api.POST("/logout", logout)
```

### Host-Based Routing

```go
func (r RouterEngine) Host(hostname string) RouterEngine
```

Creates routes for a specific hostname (multi-tenancy).

**Example:**
```go
apiHost := router.Host("api.example.com")
apiHost.GET("/", apiHome)
```

### Static Files

```go
func (r RouterEngine) Static(prefix string, filesystem VirtualFS) RouterEngine
```

Serves static files from a filesystem.

**Example:**
```go
router.Static("/static", fileSystem)
```

### WebSocket

```go
func (r RouterEngine) WebSocket(path string, handler WebSocketHandler, middleware ...MiddlewareFunc) RouterEngine
```

Registers a WebSocket handler.

**Example:**
```go
router.WebSocket("/ws", wsHandler)
```

### Multi-Protocol APIs

```go
func (r RouterEngine) GraphQL(path string, schema GraphQLSchema, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) GRPC(service GRPCService, middleware ...MiddlewareFunc) RouterEngine
func (r RouterEngine) SOAP(path string, service SOAPService, middleware ...MiddlewareFunc) RouterEngine
```

## Database

Database management interface.

### Connection Methods

```go
func (db DatabaseManager) Connect(config DatabaseConfig) error
func (db DatabaseManager) Close() error
```

### Query Methods

```go
func (db DatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error)
func (db DatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error)
```

**Example:**
```go
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
```

### Transaction Methods

```go
func (db DatabaseManager) Begin() (Transaction, error)
```

**Example:**
```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Perform operations
tx.Exec("INSERT INTO users ...")

// Commit
tx.Commit()
```

## Cache

Cache management interface.

### Cache Methods

```go
func (c CacheManager) Get(key string) (interface{}, bool)
func (c CacheManager) Set(key string, value interface{}, ttl time.Duration) error
func (c CacheManager) Delete(key string) error
func (c CacheManager) Clear() error
```

**Example:**
```go
// Get from cache
data, found := cache.Get("user:123")
if !found {
    // Fetch from database
    data = fetchUser(123)
    // Store in cache
    cache.Set("user:123", data, 5*time.Minute)
}
```

## Session

Session management interface.

### Session Methods

```go
func (s SessionManager) Create(ctx Context) (*Session, error)
func (s SessionManager) Load(ctx Context, sessionID string) (*Session, error)
func (s SessionManager) Save(ctx Context, session *Session) error
func (s SessionManager) Destroy(ctx Context, sessionID string) error
```

### Session Data Methods

```go
func (s SessionManager) Get(key string) interface{}
func (s SessionManager) Set(key string, value interface{})
func (s SessionManager) Delete(key string)
```

**Example:**
```go
session := ctx.Session()

// Get data
userID := session.Get("user_id")

// Set data
session.Set("user_id", "123")

// Save session
session.Save(ctx)
```

## Security

Security management interface.

### Authentication Methods

```go
func (s SecurityManager) AuthenticateOAuth2(token string) (*User, error)
func (s SecurityManager) AuthenticateJWT(token string) (*User, error)
```

### Authorization Methods

```go
func (s SecurityManager) Authorize(user *User, resource string, action string) bool
```

### Validation Methods

```go
func (s SecurityManager) ValidateRequest(ctx Context) error
func (s SecurityManager) ValidateFormData(ctx Context, rules ValidationRules) error
```

## Configuration

Configuration management interface.

### Configuration Methods

```go
func (c ConfigManager) LoadFile(path string) error
func (c ConfigManager) Get(key string) interface{}
func (c ConfigManager) GetString(key string) string
func (c ConfigManager) GetInt(key string) int
func (c ConfigManager) GetBool(key string) bool
```

**Example:**
```go
config := ctx.Config()
dbHost := config.GetString("database.host")
dbPort := config.GetInt("database.port")
```

## Internationalization

Internationalization (i18n) interface.

### I18n Methods

```go
func (i I18nManager) Translate(key string, params ...interface{}) string
func (i I18nManager) SetLocale(locale string) error
func (i I18nManager) GetLocale() string
```

**Example:**
```go
i18n := ctx.I18n()
message := i18n.Translate("welcome_message", userName)
```

## Types

### FrameworkConfig

```go
type FrameworkConfig struct {
    ServerConfig     ServerConfig
    DatabaseConfig   DatabaseConfig
    CacheConfig      CacheConfig
    SessionConfig    SessionConfig
    ConfigFiles      []string
    I18nConfig       I18nConfig
    SecurityConfig   SecurityConfig
    MonitoringConfig MonitoringConfig
    ProxyConfig      ProxyConfig
}
```

### ServerConfig

```go
type ServerConfig struct {
    ReadTimeout      time.Duration
    WriteTimeout     time.Duration
    IdleTimeout      time.Duration
    MaxHeaderBytes   int
    EnableHTTP1      bool
    EnableHTTP2      bool
    EnableQUIC       bool
    EnableMetrics    bool
    MetricsPath      string
    EnablePprof      bool
    PprofPath        string
    ShutdownTimeout  time.Duration
}
```

### DatabaseConfig

```go
type DatabaseConfig struct {
    Driver          string
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}
```

### HandlerFunc

```go
type HandlerFunc func(ctx Context) error
```

Handler function signature for route handlers.

### MiddlewareFunc

```go
type MiddlewareFunc func(ctx Context, next HandlerFunc) error
```

Middleware function signature.

---

For more detailed examples, see the [examples/](../examples/) directory.
