# Configuration Reference

Complete reference for all configuration options in the Rockstar Web Framework.

## Table of Contents

1. [Framework Configuration](#framework-configuration)
2. [Server Configuration](#server-configuration)
3. [Database Configuration](#database-configuration)
4. [Cache Configuration](#cache-configuration)
5. [Session Configuration](#session-configuration)
6. [Security Configuration](#security-configuration)
7. [Monitoring Configuration](#monitoring-configuration)
8. [Proxy Configuration](#proxy-configuration)
9. [I18n Configuration](#i18n-configuration)
10. [Plugin Configuration](#plugin-configuration)
11. [Listener Configuration](#listener-configuration)
12. [Cookie Configuration](#cookie-configuration)
13. [OAuth2 Configuration](#oauth2-configuration)
14. [Configuration File Formats](#configuration-file-formats)
15. [Environment Variables](#environment-variables)
16. [Configuration Examples](#configuration-examples)

---

## Framework Configuration

The main `FrameworkConfig` struct holds the complete framework configuration.

```go
type FrameworkConfig struct {
    ServerConfig       ServerConfig
    DatabaseConfig     DatabaseConfig
    CacheConfig        CacheConfig
    SessionConfig      SessionConfig
    SecurityConfig     SecurityConfig
    MonitoringConfig   MonitoringConfig
    ProxyConfig        ProxyConfig
    I18nConfig         I18nConfig
    ConfigFiles        []string
    PluginConfigPath   string
    EnablePlugins      bool
    FileSystemRoot     string
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ServerConfig` | `ServerConfig` | See [Server Configuration](#server-configuration) | HTTP server settings |
| `DatabaseConfig` | `DatabaseConfig` | See [Database Configuration](#database-configuration) | Database connection settings |
| `CacheConfig` | `CacheConfig` | See [Cache Configuration](#cache-configuration) | Caching system settings |
| `SessionConfig` | `SessionConfig` | See [Session Configuration](#session-configuration) | Session management settings |
| `SecurityConfig` | `SecurityConfig` | See [Security Configuration](#security-configuration) | Security features settings |
| `MonitoringConfig` | `MonitoringConfig` | See [Monitoring Configuration](#monitoring-configuration) | Monitoring and metrics settings |
| `ProxyConfig` | `ProxyConfig` | See [Proxy Configuration](#proxy-configuration) | Proxy and load balancing settings |
| `I18nConfig` | `I18nConfig` | See [I18n Configuration](#i18n-configuration) | Internationalization settings |
| `ConfigFiles` | `[]string` | `[]` | List of configuration file paths to load |
| `PluginConfigPath` | `string` | `""` | Path to plugin configuration file |
| `EnablePlugins` | `bool` | `false` | Enable the plugin system |
| `FileSystemRoot` | `string` | `"."` | Root directory for file operations |

### Example

```go
config := pkg.FrameworkConfig{
    ServerConfig: pkg.ServerConfig{
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    },
    DatabaseConfig: pkg.DatabaseConfig{
        Driver:   "postgres",
        Host:     "localhost",
        Port:     5432,
        Database: "myapp",
        Username: "user",
        Password: "pass",
    },
    EnablePlugins: true,
}

app, err := pkg.New(config)
```

---

## Server Configuration

Controls HTTP server behavior, protocols, timeouts, and connection limits.

```go
type ServerConfig struct {
    // Timeouts
    ReadTimeout         time.Duration
    WriteTimeout        time.Duration
    IdleTimeout         time.Duration
    ShutdownTimeout     time.Duration
    
    // Protocols
    EnableHTTP1         bool
    EnableHTTP2         bool
    EnableQUIC          bool
    
    // TLS/Security
    TLSConfig           *tls.Config
    EnableHSTS          bool
    HSTSMaxAge          time.Duration
    HSTSIncludeSubdomains bool
    HSTSPreload         bool
    
    // Limits
    MaxHeaderBytes      int
    MaxConnections      int
    MaxRequestSize      int64
    
    // Performance
    ReadBufferSize      int
    WriteBufferSize     int
    
    // Multi-tenancy
    HostConfigs         map[string]*HostConfig
    
    // Monitoring
    EnableMetrics       bool
    MetricsPath         string
    EnablePprof         bool
    PprofPath           string
    
    // Platform-specific
    ListenerConfig      *ListenerConfig
    EnablePrefork       bool
    PreforkWorkers      int
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **Timeouts** | | | |
| `ReadTimeout` | `time.Duration` | `30s` | Maximum duration for reading entire request |
| `WriteTimeout` | `time.Duration` | `30s` | Maximum duration before timing out writes |
| `IdleTimeout` | `time.Duration` | `120s` | Maximum time to wait for next request with keep-alives |
| `ShutdownTimeout` | `time.Duration` | `30s` | Maximum duration to wait for graceful shutdown |
| **Protocols** | | | |
| `EnableHTTP1` | `bool` | `false` | Enable HTTP/1.1 protocol support |
| `EnableHTTP2` | `bool` | `false` | Enable HTTP/2 protocol support |
| `EnableQUIC` | `bool` | `false` | Enable QUIC protocol support |
| **TLS/Security** | | | |
| `TLSConfig` | `*tls.Config` | `nil` | TLS configuration for secure connections |
| `EnableHSTS` | `bool` | `false` | Enable HTTP Strict Transport Security |
| `HSTSMaxAge` | `time.Duration` | `0` | HSTS max-age in seconds (default: 1 year) |
| `HSTSIncludeSubdomains` | `bool` | `false` | Include subdomains in HSTS |
| `HSTSPreload` | `bool` | `false` | Enable HSTS preload |
| **Limits** | | | |
| `MaxHeaderBytes` | `int` | `1048576` (1 MB) | Maximum bytes for request header |
| `MaxConnections` | `int` | `10000` | Maximum concurrent connections |
| `MaxRequestSize` | `int64` | `10485760` (10 MB) | Maximum request body size in bytes |
| **Performance** | | | |
| `ReadBufferSize` | `int` | `4096` | Size of read buffer in bytes |
| `WriteBufferSize` | `int` | `4096` | Size of write buffer in bytes |
| **Multi-tenancy** | | | |
| `HostConfigs` | `map[string]*HostConfig` | `nil` | Host-specific configurations |
| **Monitoring** | | | |
| `EnableMetrics` | `bool` | `false` | Enable metrics endpoint |
| `MetricsPath` | `string` | `""` | Path for metrics endpoint |
| `EnablePprof` | `bool` | `false` | Enable pprof profiling endpoints |
| `PprofPath` | `string` | `""` | Path for pprof endpoints |
| **Platform-specific** | | | |
| `ListenerConfig` | `*ListenerConfig` | `nil` | Platform-specific listener configuration |
| `EnablePrefork` | `bool` | `false` | Enable prefork mode for multi-process servers |
| `PreforkWorkers` | `int` | `0` | Number of worker processes in prefork mode |

### HostConfig

Configuration for multi-tenant host-specific settings.

```go
type HostConfig struct {
    Hostname       string
    TenantID       string
    VirtualFS      VirtualFS
    Middleware     []MiddlewareFunc
    RateLimits     *RateLimitConfig
    SecurityConfig *ServerSecurityConfig
}
```

### RateLimitConfig

```go
type RateLimitConfig struct {
    Enabled           bool
    RequestsPerSecond int
    BurstSize         int
    Storage           string  // "memory", "database", "redis"
}
```

### Example

```go
config := pkg.ServerConfig{
    ReadTimeout:    30 * time.Second,
    WriteTimeout:   30 * time.Second,
    IdleTimeout:    120 * time.Second,
    MaxConnections: 10000,
    EnableHTTP2:    true,
    EnableHSTS:     true,
    HSTSMaxAge:     365 * 24 * time.Hour,
}
```

---

## Database Configuration

Configures database connections for MySQL, PostgreSQL, MSSQL, and SQLite.

```go
type DatabaseConfig struct {
    Driver          string
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    SSLMode         string
    Charset         string
    Timezone        string
    ConnMaxLifetime time.Duration
    MaxOpenConns    int
    MaxIdleConns    int
    Options         map[string]string
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Driver` | `string` | **Required** | Database driver: `mysql`, `postgres`, `mssql`, `sqlite` |
| `Host` | `string` | `"localhost"` | Database server hostname or IP address |
| `Port` | `int` | Driver-specific | Database server port (postgres=5432, mysql=3306, mssql=1433, sqlite=0) |
| `Database` | `string` | **Required** | Name of the database to connect to |
| `Username` | `string` | **Required** | Database user for authentication |
| `Password` | `string` | **Required** | Database password for authentication |
| `SSLMode` | `string` | `""` | SSL/TLS mode for the connection |
| `Charset` | `string` | `""` | Character set for the connection |
| `Timezone` | `string` | `""` | Timezone for the connection |
| `ConnMaxLifetime` | `time.Duration` | `5m` | Maximum time a connection may be reused |
| `MaxOpenConns` | `int` | `25` | Maximum number of open connections |
| `MaxIdleConns` | `int` | `5` | Maximum number of idle connections in pool |
| `Options` | `map[string]string` | `nil` | Driver-specific connection options |

### Example

```go
// PostgreSQL
config := pkg.DatabaseConfig{
    Driver:          "postgres",
    Host:            "localhost",
    Port:            5432,
    Database:        "myapp",
    Username:        "dbuser",
    Password:        "dbpass",
    SSLMode:         "require",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
}

// MySQL
config := pkg.DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "dbuser",
    Password: "dbpass",
    Charset:  "utf8mb4",
}

// SQLite
config := pkg.DatabaseConfig{
    Driver:   "sqlite",
    Database: "./myapp.db",
}
```

---

## Cache Configuration

Configures the caching system with memory or distributed backends.

```go
type CacheConfig struct {
    Type       string
    MaxSize    int64
    DefaultTTL time.Duration
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Type` | `string` | `"memory"` | Cache backend type: `memory`, `distributed` |
| `MaxSize` | `int64` | `0` (unlimited) | Maximum cache size in bytes (0 = unlimited) |
| `DefaultTTL` | `time.Duration` | `0` (no expiration) | Default time-to-live for cache entries (0 = no expiration) |

### Example

```go
config := pkg.CacheConfig{
    Type:       "memory",
    MaxSize:    100 * 1024 * 1024, // 100 MB
    DefaultTTL: 5 * time.Minute,
}
```

---

## Session Configuration

Configures session management with database, cache, or filesystem storage.

```go
type SessionConfig struct {
    StorageType     SessionStorageType
    CookieName      string
    CookiePath      string
    CookieDomain    string
    CookieSecure    bool
    CookieHTTPOnly  bool
    CookieSameSite  string
    SessionLifetime time.Duration
    EncryptionKey   []byte
    FilesystemPath  string
    CleanupInterval time.Duration
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `StorageType` | `SessionStorageType` | `"database"` | Storage backend: `database`, `cache`, `filesystem` |
| `CookieName` | `string` | `"rockstar_session"` | Name of the session cookie |
| `CookiePath` | `string` | `"/"` | Path scope for the session cookie |
| `CookieDomain` | `string` | `""` | Domain scope for the session cookie |
| `CookieSecure` | `bool` | `true` | Cookie only sent over HTTPS |
| `CookieHTTPOnly` | `bool` | `true` | Cookie inaccessible to JavaScript |
| `CookieSameSite` | `string` | `"Lax"` | SameSite attribute: `Strict`, `Lax`, `None` |
| `SessionLifetime` | `time.Duration` | `24h` | Duration before a session expires |
| `EncryptionKey` | `[]byte` | **Required** | AES-256 key (32 bytes) for encrypting session data |
| `FilesystemPath` | `string` | `"./sessions"` | Directory path for filesystem-based storage |
| `CleanupInterval` | `time.Duration` | `1h` | Interval for cleaning up expired sessions |

### Example

```go
// Generate encryption key (32 bytes for AES-256)
encryptionKey := make([]byte, 32)
rand.Read(encryptionKey)

config := pkg.SessionConfig{
    StorageType:     pkg.SessionStorageDatabase,
    CookieName:      "my_session",
    CookieSecure:    true,
    CookieHTTPOnly:  true,
    SessionLifetime: 24 * time.Hour,
    EncryptionKey:   encryptionKey,
    CleanupInterval: 1 * time.Hour,
}
```

---

## Security Configuration

Configures security features including CSRF, XSS protection, and input validation.

```go
type SecurityConfig struct {
    MaxRequestSize        int64
    RequestTimeout        time.Duration
    CSRFTokenExpiry       time.Duration
    EncryptionKey         string
    JWTSecret             string
    XFrameOptions         string
    EnableXSSProtect      bool
    EnableCSRF            bool
    AllowedOrigins        []string
    EnableHSTS            bool
    HSTSMaxAge            int
    HSTSIncludeSubdomains bool
    HSTSPreload           bool
    ProductionMode        bool
    MaxHeaderSize         int
    MaxURLLength          int
    MaxFormFieldSize      int
    MaxFormFields         int
    MaxFileNameLength     int
    MaxCookieSize         int
    MaxQueryParams        int
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **Request Limits** | | | |
| `MaxRequestSize` | `int64` | `10485760` (10 MB) | Maximum request size in bytes |
| `RequestTimeout` | `time.Duration` | `30s` | Request timeout duration |
| **CSRF Protection** | | | |
| `CSRFTokenExpiry` | `time.Duration` | `24h` | CSRF token expiry duration |
| `EnableCSRF` | `bool` | `true` | Enable CSRF protection |
| **Encryption** | | | |
| `EncryptionKey` | `string` | `""` | Hex-encoded encryption key for cookies |
| `JWTSecret` | `string` | `""` | JWT secret key |
| **Headers** | | | |
| `XFrameOptions` | `string` | `"SAMEORIGIN"` | X-Frame-Options header value |
| `EnableXSSProtect` | `bool` | `true` | Enable XSS protection headers |
| **CORS** | | | |
| `AllowedOrigins` | `[]string` | `[]` | Allowed origins for CORS (empty = no wildcard) |
| **HSTS** | | | |
| `EnableHSTS` | `bool` | `true` | Enable HTTP Strict Transport Security |
| `HSTSMaxAge` | `int` | `31536000` (1 year) | HSTS max-age in seconds |
| `HSTSIncludeSubdomains` | `bool` | `true` | Include subdomains in HSTS |
| `HSTSPreload` | `bool` | `false` | Enable HSTS preload |
| **Production** | | | |
| `ProductionMode` | `bool` | `false` | Hide sensitive error details in production |
| **Input Limits** | | | |
| `MaxHeaderSize` | `int` | `8192` (8 KB) | Maximum size of a single header value |
| `MaxURLLength` | `int` | `2048` | Maximum URL length |
| `MaxFormFieldSize` | `int` | `1048576` (1 MB) | Maximum size of a single form field |
| `MaxFormFields` | `int` | `1000` | Maximum number of form fields |
| `MaxFileNameLength` | `int` | `255` | Maximum filename length |
| `MaxCookieSize` | `int` | `4096` (4 KB) | Maximum cookie size |
| `MaxQueryParams` | `int` | `100` | Maximum number of query parameters |

### Example

```go
config := pkg.SecurityConfig{
    MaxRequestSize:   10 * 1024 * 1024, // 10 MB
    RequestTimeout:   30 * time.Second,
    EnableCSRF:       true,
    EnableXSSProtect: true,
    EnableHSTS:       true,
    HSTSMaxAge:       31536000, // 1 year
    ProductionMode:   true,
    AllowedOrigins:   []string{"https://example.com"},
}
```

---

## Monitoring Configuration

Configures monitoring, metrics, profiling, and process optimization.

```go
type MonitoringConfig struct {
    EnableMetrics        bool
    MetricsPath          string
    MetricsPort          int
    EnablePprof          bool
    PprofPath            string
    PprofPort            int
    EnableSNMP           bool
    SNMPPort             int
    SNMPCommunity        string
    EnableOptimization   bool
    OptimizationInterval time.Duration
    RequireAuth          bool
    AuthToken            string
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **Metrics** | | | |
| `EnableMetrics` | `bool` | `false` | Enable metrics HTTP endpoint |
| `MetricsPath` | `string` | `""` | HTTP path for metrics endpoint |
| `MetricsPort` | `int` | `9090` | Port number for metrics HTTP server |
| **Profiling** | | | |
| `EnablePprof` | `bool` | `false` | Enable pprof profiling endpoints |
| `PprofPath` | `string` | `""` | HTTP path for pprof endpoints |
| `PprofPort` | `int` | `6060` | Port number for pprof HTTP server |
| **SNMP** | | | |
| `EnableSNMP` | `bool` | `false` | Enable SNMP monitoring support |
| `SNMPPort` | `int` | `161` | Port number for SNMP server |
| `SNMPCommunity` | `string` | `"public"` | SNMP community string for authentication |
| **Optimization** | | | |
| `EnableOptimization` | `bool` | `false` | Enable automatic process optimization |
| `OptimizationInterval` | `time.Duration` | `5m` | Interval between automatic optimizations |
| **Security** | | | |
| `RequireAuth` | `bool` | `false` | Require authentication for monitoring endpoints |
| `AuthToken` | `string` | `""` | Bearer token for authentication |

### Example

```go
config := pkg.MonitoringConfig{
    EnableMetrics:        true,
    MetricsPort:          9090,
    EnablePprof:          true,
    PprofPort:            6060,
    EnableOptimization:   true,
    OptimizationInterval: 5 * time.Minute,
    RequireAuth:          true,
    AuthToken:            "secret-token",
}
```

---

## Proxy Configuration

Configures forward proxy, load balancing, and circuit breaker settings.

```go
type ProxyConfig struct {
    LoadBalancerType           string
    CircuitBreakerEnabled      bool
    CircuitBreakerThreshold    int
    CircuitBreakerTimeout      time.Duration
    CircuitBreakerResetTimeout time.Duration
    MaxConnectionsPerBackend   int
    ConnectionTimeout          time.Duration
    IdleConnTimeout            time.Duration
    MaxRetries                 int
    RetryDelay                 time.Duration
    RetryBackoff               bool
    CacheEnabled               bool
    CacheTTL                   time.Duration
    CacheMaxSize               int64
    HealthCheckEnabled         bool
    HealthCheckInterval        time.Duration
    HealthCheckTimeout         time.Duration
    HealthCheckPath            string
    RequestTimeout             time.Duration
    DNSCacheEnabled            bool
    DNSCacheTTL                time.Duration
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **Load Balancing** | | | |
| `LoadBalancerType` | `string` | `"round_robin"` | Strategy: `round_robin`, `weighted_round_robin`, `least_connections` |
| **Circuit Breaker** | | | |
| `CircuitBreakerEnabled` | `bool` | `true` | Enable circuit breaker pattern |
| `CircuitBreakerThreshold` | `int` | `5` | Failures before opening circuit |
| `CircuitBreakerTimeout` | `time.Duration` | `30s` | Time before trying half-open state |
| `CircuitBreakerResetTimeout` | `time.Duration` | `60s` | Time before resetting counters |
| **Connection Pool** | | | |
| `MaxConnectionsPerBackend` | `int` | `100` | Maximum connections per backend |
| `ConnectionTimeout` | `time.Duration` | `10s` | Connection timeout |
| `IdleConnTimeout` | `time.Duration` | `90s` | Idle connection timeout |
| **Retry** | | | |
| `MaxRetries` | `int` | `3` | Maximum retry attempts |
| `RetryDelay` | `time.Duration` | `100ms` | Delay between retries |
| `RetryBackoff` | `bool` | `true` | Use exponential backoff |
| **Cache** | | | |
| `CacheEnabled` | `bool` | `true` | Enable response caching |
| `CacheTTL` | `time.Duration` | `5m` | Cache time-to-live |
| `CacheMaxSize` | `int64` | `104857600` (100 MB) | Maximum cache size in bytes |
| **Health Check** | | | |
| `HealthCheckEnabled` | `bool` | `true` | Enable backend health checks |
| `HealthCheckInterval` | `time.Duration` | `10s` | Interval between health checks |
| `HealthCheckTimeout` | `time.Duration` | `5s` | Health check timeout |
| `HealthCheckPath` | `string` | `"/health"` | Health check endpoint path |
| **Request** | | | |
| `RequestTimeout` | `time.Duration` | `30s` | Request timeout |
| **DNS** | | | |
| `DNSCacheEnabled` | `bool` | `true` | Enable DNS caching |
| `DNSCacheTTL` | `time.Duration` | `5m` | DNS cache time-to-live |

### Example

```go
config := pkg.ProxyConfig{
    LoadBalancerType:        "round_robin",
    CircuitBreakerEnabled:   true,
    CircuitBreakerThreshold: 5,
    MaxRetries:              3,
    RetryBackoff:            true,
    CacheEnabled:            true,
    CacheTTL:                5 * time.Minute,
    HealthCheckEnabled:      true,
    HealthCheckInterval:     10 * time.Second,
}
```

---

## I18n Configuration

Configures internationalization and localization support.

```go
type I18nConfig struct {
    DefaultLocale     string
    LocalesDir        string
    SupportedLocales  []string
    FallbackToDefault bool
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DefaultLocale` | `string` | `"en"` | Fallback locale when translation not found |
| `LocalesDir` | `string` | `""` | Directory containing locale files |
| `SupportedLocales` | `[]string` | `[]` | List of supported locales |
| `FallbackToDefault` | `bool` | `true` | Fall back to default locale for missing translations |

### Locale File Format

Locale files should be named `locales.{locale}.yaml` (e.g., `locales.en.yaml`, `locales.de.yaml`).

```yaml
# locales.en.yaml
welcome: "Welcome"
error:
  auth:
    failed: "Authentication failed"
  validation:
    required: "{{field}} is required"
```

### Example

```go
config := pkg.I18nConfig{
    DefaultLocale:     "en",
    LocalesDir:        "./locales",
    SupportedLocales:  []string{"en", "de", "fr"},
    FallbackToDefault: true,
}
```

---

## Plugin Configuration

Configures the plugin system and individual plugins.

```go
type PluginsConfig struct {
    Enabled   bool
    Directory string
    Plugins   []PluginConfigEntry
}

type PluginConfigEntry struct {
    Name        string
    Enabled     bool
    Path        string
    Priority    int
    Config      map[string]interface{}
    Permissions PluginPermissions
}

type PluginPermissions struct {
    AllowDatabase      bool
    AllowCache         bool
    AllowConfig        bool
    AllowRouter        bool
    AllowFileSystem    bool
    AllowNetwork       bool
    AllowExec          bool
    CustomPermissions  map[string]bool
}
```

### Fields

#### PluginsConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Enabled` | `bool` | `true` | Enable plugin system globally |
| `Directory` | `string` | `"./plugins"` | Base directory for plugin files |
| `Plugins` | `[]PluginConfigEntry` | `[]` | List of plugin configurations |

#### PluginConfigEntry

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Name` | `string` | **Required** | Unique identifier for the plugin |
| `Enabled` | `bool` | `true` | Enable this plugin |
| `Path` | `string` | **Required** | File path to the plugin binary |
| `Priority` | `int` | `0` | Execution order (lower values execute first) |
| `Config` | `map[string]interface{}` | `{}` | Plugin-specific configuration |
| `Permissions` | `PluginPermissions` | All `false` | Plugin permissions |

#### PluginPermissions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `AllowDatabase` | `bool` | `false` | Allow database access |
| `AllowCache` | `bool` | `false` | Allow cache access |
| `AllowConfig` | `bool` | `false` | Allow configuration access |
| `AllowRouter` | `bool` | `false` | Allow router modifications |
| `AllowFileSystem` | `bool` | `false` | Allow file system access |
| `AllowNetwork` | `bool` | `false` | Allow network access |
| `AllowExec` | `bool` | `false` | Allow command execution |
| `CustomPermissions` | `map[string]bool` | `{}` | Custom permission flags |

### Example (YAML)

```yaml
plugins:
  enabled: true
  directory: "./plugins"
  plugins:
    - name: "auth-plugin"
      enabled: true
      path: "./plugins/auth.so"
      priority: 10
      config:
        provider: "oauth2"
        client_id: "abc123"
      permissions:
        database: true
        cache: true
        router: true
    - name: "logging-plugin"
      enabled: true
      path: "./plugins/logging.so"
      priority: 5
      permissions:
        filesystem: true
```

---

## Listener Configuration

Platform-specific listener configuration for advanced networking.

```go
type ListenerConfig struct {
    Network         string
    Address         string
    EnablePrefork   bool
    PreforkWorkers  int
    ReusePort       bool
    ReuseAddr       bool
    ReadBuffer      int
    WriteBuffer     int
    PlatformOptions map[string]interface{}
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Network` | `string` | `"tcp"` | Network type: `tcp`, `tcp4`, `tcp6`, `unix` |
| `Address` | `string` | **Required** | Address to listen on |
| `EnablePrefork` | `bool` | `false` | Enable prefork mode |
| `PreforkWorkers` | `int` | `0` (NumCPU) | Number of worker processes |
| `ReusePort` | `bool` | `false` | Enable SO_REUSEPORT (Linux/BSD) |
| `ReuseAddr` | `bool` | `false` | Enable SO_REUSEADDR |
| `ReadBuffer` | `int` | `0` | Read buffer size |
| `WriteBuffer` | `int` | `0` | Write buffer size |
| `PlatformOptions` | `map[string]interface{}` | `nil` | Platform-specific options |

### Example

```go
config := pkg.ListenerConfig{
    Network:        "tcp",
    Address:        ":8080",
    EnablePrefork:  true,
    PreforkWorkers: 4,
    ReusePort:      true,
}
```

---

## Cookie Configuration

Configures cookie management and encryption.

```go
type CookieConfig struct {
    EncryptionKey   []byte
    DefaultPath     string
    DefaultDomain   string
    DefaultSecure   bool
    DefaultHTTPOnly bool
    DefaultSameSite http.SameSite
    DefaultMaxAge   int
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `EncryptionKey` | `[]byte` | `nil` | AES-256 key (32 bytes) for cookie encryption |
| `DefaultPath` | `string` | `"/"` | Default cookie path |
| `DefaultDomain` | `string` | `""` | Default cookie domain |
| `DefaultSecure` | `bool` | `true` | Default secure flag |
| `DefaultHTTPOnly` | `bool` | `true` | Default HTTP-only flag |
| `DefaultSameSite` | `http.SameSite` | `http.SameSiteLaxMode` | Default SameSite policy |
| `DefaultMaxAge` | `int` | `86400` (24 hours) | Default max age in seconds |

### Example

```go
// Generate encryption key
encryptionKey := make([]byte, 32)
rand.Read(encryptionKey)

config := pkg.CookieConfig{
    EncryptionKey:   encryptionKey,
    DefaultPath:     "/",
    DefaultSecure:   true,
    DefaultHTTPOnly: true,
    DefaultSameSite: http.SameSiteStrictMode,
    DefaultMaxAge:   86400,
}
```

---

## OAuth2 Configuration

Configures OAuth2 authentication.

```go
type OAuth2Config struct {
    ClientID     string
    ClientSecret string
    TokenURL     string
    AuthURL      string
    RedirectURL  string
    Scopes       []string
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ClientID` | `string` | **Required** | OAuth2 client ID |
| `ClientSecret` | `string` | **Required** | OAuth2 client secret |
| `TokenURL` | `string` | **Required** | Token endpoint URL |
| `AuthURL` | `string` | **Required** | Authorization endpoint URL |
| `RedirectURL` | `string` | **Required** | Redirect URL after authentication |
| `Scopes` | `[]string` | `[]` | Requested OAuth2 scopes |

### Example

```go
config := pkg.OAuth2Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    TokenURL:     "https://provider.com/oauth/token",
    AuthURL:      "https://provider.com/oauth/authorize",
    RedirectURL:  "https://yourapp.com/callback",
    Scopes:       []string{"read", "write"},
}
```

---

## Configuration File Formats

The framework supports multiple configuration file formats.

### Supported Formats

- **JSON** (`.json`)
- **YAML** (`.yaml`, `.yml`)
- **TOML** (`.toml`)
- **INI** (`.ini`)

### Loading Configuration Files

```go
config := pkg.FrameworkConfig{
    ConfigFiles: []string{
        "config/app.yaml",
        "config/database.yaml",
        "config/security.yaml",
    },
}

app, err := pkg.New(config)
```

### JSON Example

```json
{
  "server": {
    "read_timeout": "30s",
    "write_timeout": "30s",
    "max_connections": 10000
  },
  "database": {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432,
    "database": "myapp",
    "username": "user",
    "password": "pass"
  }
}
```

### YAML Example

```yaml
server:
  read_timeout: 30s
  write_timeout: 30s
  max_connections: 10000

database:
  driver: postgres
  host: localhost
  port: 5432
  database: myapp
  username: user
  password: pass
```

### TOML Example

```toml
[server]
read_timeout = "30s"
write_timeout = "30s"
max_connections = 10000

[database]
driver = "postgres"
host = "localhost"
port = 5432
database = "myapp"
username = "user"
password = "pass"
```

---

## Environment Variables

Configuration can be loaded from environment variables with the `ROCKSTAR_` prefix.

### Variable Naming

Environment variables use underscores for nested configuration:

- `ROCKSTAR_APP_NAME` → `app.name`
- `ROCKSTAR_DATABASE_HOST` → `database.host`
- `ROCKSTAR_DATABASE_PORT` → `database.port`

### Example

```bash
export ROCKSTAR_APP_NAME="MyApp"
export ROCKSTAR_PORT="8080"
export ROCKSTAR_DATABASE_HOST="localhost"
export ROCKSTAR_DATABASE_PORT="5432"
export ROCKSTAR_DATABASE_NAME="myapp"
export ROCKSTAR_CACHE_ENABLED="true"
export ROCKSTAR_CACHE_TTL="3600"
```

### Loading from Environment

```go
config := pkg.NewConfigManager()
err := config.LoadFromEnv()
if err != nil {
    log.Fatal(err)
}

// Access values
appName := config.GetString("app.name")
port := config.GetInt("port")
dbHost := config.GetString("database.host")
```

---

## Configuration Examples

This section provides complete, production-ready configuration examples for different environments and use cases.

### Production Configuration

A complete production-ready configuration optimized for security, performance, and reliability.

```go
package main

import (
    "crypto/rand"
    "crypto/tls"
    "log"
    "time"
    
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Generate encryption keys (in production, load from secure storage)
    sessionKey := make([]byte, 32)
    rand.Read(sessionKey)
    
    cookieKey := make([]byte, 32)
    rand.Read(cookieKey)
    
    // TLS configuration for HTTPS
    tlsConfig := &tls.Config{
        MinVersion:               tls.VersionTLS12,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        },
    }
    
    config := pkg.FrameworkConfig{
        // Server configuration - optimized for production workloads
        ServerConfig: pkg.ServerConfig{
            // Timeouts - prevent resource exhaustion
            ReadTimeout:     30 * time.Second,  // Time to read request headers and body
            WriteTimeout:    30 * time.Second,  // Time to write response
            IdleTimeout:     120 * time.Second, // Keep-alive timeout
            ShutdownTimeout: 30 * time.Second,  // Graceful shutdown timeout
            
            // Protocol support
            EnableHTTP1: true,  // Enable HTTP/1.1 for compatibility
            EnableHTTP2: true,  // Enable HTTP/2 for performance
            EnableQUIC:  false, // QUIC disabled by default (enable if needed)
            
            // TLS/Security
            TLSConfig:             tlsConfig,
            EnableHSTS:            true,
            HSTSMaxAge:            365 * 24 * time.Hour, // 1 year
            HSTSIncludeSubdomains: true,
            HSTSPreload:           true,
            
            // Connection limits - prevent resource exhaustion
            MaxHeaderBytes:  1048576,         // 1 MB max header size
            MaxConnections:  10000,           // Max concurrent connections
            MaxRequestSize:  10 * 1024 * 1024, // 10 MB max request body
            
            // Performance tuning
            ReadBufferSize:  8192, // 8 KB read buffer
            WriteBufferSize: 8192, // 8 KB write buffer
            
            // Monitoring
            EnableMetrics: true,
            MetricsPath:   "/metrics",
            EnablePprof:   false, // Disable pprof in production
        },
        
        // Database configuration - production PostgreSQL
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "postgres",
            Host:     "db.production.example.com",
            Port:     5432,
            Database: "myapp_production",
            Username: "myapp_user",      // Load from environment
            Password: "secure_password", // Load from environment/secrets manager
            SSLMode:  "require",         // Require SSL for database connections
            
            // Connection pool tuning
            MaxOpenConns:    50,              // Max open connections
            MaxIdleConns:    10,              // Max idle connections in pool
            ConnMaxLifetime: 5 * time.Minute, // Recycle connections every 5 minutes
            
            // Additional options
            Options: map[string]string{
                "application_name": "myapp",
                "connect_timeout":  "10",
            },
        },
        
        // Cache configuration - memory cache for performance
        CacheConfig: pkg.CacheConfig{
            Type:       "memory",
            MaxSize:    500 * 1024 * 1024, // 500 MB cache
            DefaultTTL: 5 * time.Minute,   // 5 minute default TTL
        },
        
        // Session configuration - database-backed sessions
        SessionConfig: pkg.SessionConfig{
            StorageType:     pkg.SessionStorageDatabase,
            CookieName:      "app_session",
            CookiePath:      "/",
            CookieSecure:    true,              // HTTPS only
            CookieHTTPOnly:  true,              // No JavaScript access
            CookieSameSite:  "Strict",          // CSRF protection
            SessionLifetime: 24 * time.Hour,    // 24 hour sessions
            EncryptionKey:   sessionKey,        // AES-256 encryption
            CleanupInterval: 1 * time.Hour,     // Clean expired sessions hourly
        },
        
        // Security configuration - maximum security
        SecurityConfig: pkg.SecurityConfig{
            // Request limits
            MaxRequestSize: 10 * 1024 * 1024, // 10 MB
            RequestTimeout: 30 * time.Second,
            
            // CSRF protection
            EnableCSRF:      true,
            CSRFTokenExpiry: 24 * time.Hour,
            
            // Encryption
            EncryptionKey: "hex_encoded_32_byte_key", // Load from environment
            JWTSecret:     "jwt_secret_key",          // Load from environment
            
            // Security headers
            XFrameOptions:    "DENY",
            EnableXSSProtect: true,
            
            // CORS
            AllowedOrigins: []string{
                "https://myapp.example.com",
                "https://www.myapp.example.com",
            },
            
            // HSTS
            EnableHSTS:            true,
            HSTSMaxAge:            31536000, // 1 year in seconds
            HSTSIncludeSubdomains: true,
            HSTSPreload:           true,
            
            // Production mode - hide error details
            ProductionMode: true,
            
            // Input validation limits
            MaxHeaderSize:     8192,  // 8 KB
            MaxURLLength:      2048,  // 2 KB
            MaxFormFieldSize:  1048576, // 1 MB
            MaxFormFields:     1000,
            MaxFileNameLength: 255,
            MaxCookieSize:     4096, // 4 KB
            MaxQueryParams:    100,
        },
        
        // Monitoring configuration - production monitoring
        MonitoringConfig: pkg.MonitoringConfig{
            // Metrics
            EnableMetrics: true,
            MetricsPath:   "/metrics",
            MetricsPort:   9090,
            
            // Profiling - disabled in production
            EnablePprof: false,
            
            // Process optimization
            EnableOptimization:   true,
            OptimizationInterval: 5 * time.Minute,
            
            // Authentication for monitoring endpoints
            RequireAuth: true,
            AuthToken:   "monitoring_bearer_token", // Load from environment
        },
        
        // Proxy configuration - for load balancing
        ProxyConfig: pkg.ProxyConfig{
            LoadBalancerType: "round_robin",
            
            // Circuit breaker for fault tolerance
            CircuitBreakerEnabled:      true,
            CircuitBreakerThreshold:    5,
            CircuitBreakerTimeout:      30 * time.Second,
            CircuitBreakerResetTimeout: 60 * time.Second,
            
            // Connection pooling
            MaxConnectionsPerBackend: 100,
            ConnectionTimeout:        10 * time.Second,
            IdleConnTimeout:          90 * time.Second,
            
            // Retry logic
            MaxRetries:   3,
            RetryDelay:   100 * time.Millisecond,
            RetryBackoff: true,
            
            // Response caching
            CacheEnabled: true,
            CacheTTL:     5 * time.Minute,
            CacheMaxSize: 100 * 1024 * 1024, // 100 MB
            
            // Health checks
            HealthCheckEnabled:  true,
            HealthCheckInterval: 10 * time.Second,
            HealthCheckTimeout:  5 * time.Second,
            HealthCheckPath:     "/health",
            
            // Request timeout
            RequestTimeout: 30 * time.Second,
            
            // DNS caching
            DNSCacheEnabled: true,
            DNSCacheTTL:     5 * time.Minute,
        },
        
        // I18n configuration
        I18nConfig: pkg.I18nConfig{
            DefaultLocale:     "en",
            LocalesDir:        "./locales",
            SupportedLocales:  []string{"en", "de", "fr", "es", "ja"},
            FallbackToDefault: true,
        },
        
        // Plugin configuration
        EnablePlugins:    true,
        PluginConfigPath: "./config/plugins.yaml",
        
        // File system
        FileSystemRoot: "/var/app/data",
    }
    
    // Create framework instance
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal("Failed to initialize framework:", err)
    }
    
    // Start server with TLS
    log.Println("Starting production server on :443")
    log.Fatal(app.ListenTLS(":443", "/etc/ssl/certs/server.crt", "/etc/ssl/private/server.key"))
}
```

**Production Configuration (YAML)**

```yaml
# config/production.yaml
# Production-ready configuration with security and performance optimizations

server:
  # Timeouts - prevent resource exhaustion
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  shutdown_timeout: 30s
  
  # Protocol support
  enable_http1: true
  enable_http2: true
  enable_quic: false
  
  # TLS/Security
  enable_hsts: true
  hsts_max_age: 31536000  # 1 year
  hsts_include_subdomains: true
  hsts_preload: true
  
  # Connection limits
  max_header_bytes: 1048576  # 1 MB
  max_connections: 10000
  max_request_size: 10485760  # 10 MB
  
  # Performance tuning
  read_buffer_size: 8192
  write_buffer_size: 8192
  
  # Monitoring
  enable_metrics: true
  metrics_path: /metrics
  enable_pprof: false

database:
  driver: postgres
  host: db.production.example.com
  port: 5432
  database: myapp_production
  username: ${DB_USER}
  password: ${DB_PASSWORD}
  ssl_mode: require
  
  # Connection pool tuning
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 5m
  
  options:
    application_name: myapp
    connect_timeout: "10"

cache:
  type: memory
  max_size: 524288000  # 500 MB
  default_ttl: 5m

session:
  storage_type: database
  cookie_name: app_session
  cookie_path: /
  cookie_secure: true
  cookie_http_only: true
  cookie_same_site: Strict
  session_lifetime: 24h
  cleanup_interval: 1h

security:
  max_request_size: 10485760
  request_timeout: 30s
  enable_csrf: true
  csrf_token_expiry: 24h
  x_frame_options: DENY
  enable_xss_protect: true
  enable_hsts: true
  hsts_max_age: 31536000
  hsts_include_subdomains: true
  hsts_preload: true
  production_mode: true
  allowed_origins:
    - https://myapp.example.com
    - https://www.myapp.example.com
  max_header_size: 8192
  max_url_length: 2048
  max_form_field_size: 1048576
  max_form_fields: 1000
  max_file_name_length: 255
  max_cookie_size: 4096
  max_query_params: 100

monitoring:
  enable_metrics: true
  metrics_path: /metrics
  metrics_port: 9090
  enable_pprof: false
  enable_optimization: true
  optimization_interval: 5m
  require_auth: true
  auth_token: ${MONITORING_TOKEN}

proxy:
  load_balancer_type: round_robin
  circuit_breaker_enabled: true
  circuit_breaker_threshold: 5
  circuit_breaker_timeout: 30s
  circuit_breaker_reset_timeout: 60s
  max_connections_per_backend: 100
  connection_timeout: 10s
  idle_conn_timeout: 90s
  max_retries: 3
  retry_delay: 100ms
  retry_backoff: true
  cache_enabled: true
  cache_ttl: 5m
  cache_max_size: 104857600  # 100 MB
  health_check_enabled: true
  health_check_interval: 10s
  health_check_timeout: 5s
  health_check_path: /health
  request_timeout: 30s
  dns_cache_enabled: true
  dns_cache_ttl: 5m

i18n:
  default_locale: en
  locales_dir: ./locales
  supported_locales:
    - en
    - de
    - fr
    - es
    - ja
  fallback_to_default: true

plugins:
  enabled: true
  directory: ./plugins

filesystem_root: /var/app/data
```

### Development Configuration

A developer-friendly configuration with debugging enabled and relaxed security for local development.

```go
package main

import (
    "crypto/rand"
    "log"
    "time"
    
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Generate encryption keys
    sessionKey := make([]byte, 32)
    rand.Read(sessionKey)
    
    config := pkg.FrameworkConfig{
        // Server configuration - relaxed for development
        ServerConfig: pkg.ServerConfig{
            // Longer timeouts for debugging
            ReadTimeout:     60 * time.Second,
            WriteTimeout:    60 * time.Second,
            IdleTimeout:     180 * time.Second,
            ShutdownTimeout: 10 * time.Second,
            
            // HTTP/1.1 only for simplicity
            EnableHTTP1: true,
            EnableHTTP2: false,
            EnableQUIC:  false,
            
            // No TLS in development
            EnableHSTS: false,
            
            // Lower limits for local development
            MaxConnections: 100,
            MaxRequestSize: 50 * 1024 * 1024, // 50 MB for testing large uploads
            
            // Enable profiling for debugging
            EnableMetrics: true,
            MetricsPath:   "/debug/metrics",
            EnablePprof:   true,
            PprofPath:     "/debug/pprof",
        },
        
        // Database configuration - local SQLite
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: "./dev.db",
            
            // Minimal connection pool for SQLite
            MaxOpenConns:    1,
            MaxIdleConns:    1,
            ConnMaxLifetime: 0, // No limit
        },
        
        // Cache configuration - small memory cache
        CacheConfig: pkg.CacheConfig{
            Type:       "memory",
            MaxSize:    10 * 1024 * 1024, // 10 MB
            DefaultTTL: 1 * time.Minute,  // Short TTL for testing
        },
        
        // Session configuration - filesystem for easy inspection
        SessionConfig: pkg.SessionConfig{
            StorageType:     pkg.SessionStorageFilesystem,
            FilesystemPath:  "./sessions",
            CookieName:      "dev_session",
            CookieSecure:    false,             // Allow HTTP
            CookieHTTPOnly:  true,
            CookieSameSite:  "Lax",
            SessionLifetime: 8 * time.Hour,     // Work day
            EncryptionKey:   sessionKey,
            CleanupInterval: 30 * time.Minute,
        },
        
        // Security configuration - relaxed for development
        SecurityConfig: pkg.SecurityConfig{
            MaxRequestSize: 50 * 1024 * 1024,
            RequestTimeout: 60 * time.Second,
            
            // Disable CSRF for easier API testing
            EnableCSRF: false,
            
            // Enable XSS protection
            EnableXSSProtect: true,
            
            // No HSTS in development
            EnableHSTS: false,
            
            // Development mode - show detailed errors
            ProductionMode: false,
            
            // Allow all origins for CORS
            AllowedOrigins: []string{"*"},
            
            // Relaxed input limits
            MaxFormFields: 10000,
        },
        
        // Monitoring configuration - full debugging
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics: true,
            MetricsPath:   "/debug/metrics",
            MetricsPort:   9090,
            
            // Enable pprof for profiling
            EnablePprof: true,
            PprofPath:   "/debug/pprof",
            PprofPort:   6060,
            
            // Disable optimization in development
            EnableOptimization: false,
            
            // No auth required
            RequireAuth: false,
        },
        
        // I18n configuration
        I18nConfig: pkg.I18nConfig{
            DefaultLocale:     "en",
            LocalesDir:        "./locales",
            SupportedLocales:  []string{"en"},
            FallbackToDefault: true,
        },
        
        // Plugin configuration
        EnablePlugins:    true,
        PluginConfigPath: "./config/plugins.dev.yaml",
        
        // File system
        FileSystemRoot: "./data",
    }
    
    // Create framework instance
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal("Failed to initialize framework:", err)
    }
    
    // Start server on HTTP
    log.Println("Starting development server on :8080")
    log.Println("Metrics: http://localhost:9090/debug/metrics")
    log.Println("Pprof: http://localhost:6060/debug/pprof")
    log.Fatal(app.Listen(":8080"))
}
```

**Development Configuration (YAML)**

```yaml
# config/development.yaml
# Developer-friendly configuration with debugging and relaxed security

server:
  # Longer timeouts for debugging
  read_timeout: 60s
  write_timeout: 60s
  idle_timeout: 180s
  shutdown_timeout: 10s
  
  # HTTP/1.1 only
  enable_http1: true
  enable_http2: false
  enable_quic: false
  
  # No TLS
  enable_hsts: false
  
  # Lower limits
  max_connections: 100
  max_request_size: 52428800  # 50 MB
  
  # Enable debugging
  enable_metrics: true
  metrics_path: /debug/metrics
  enable_pprof: true
  pprof_path: /debug/pprof

database:
  driver: sqlite
  database: ./dev.db
  max_open_conns: 1
  max_idle_conns: 1

cache:
  type: memory
  max_size: 10485760  # 10 MB
  default_ttl: 1m

session:
  storage_type: filesystem
  filesystem_path: ./sessions
  cookie_name: dev_session
  cookie_secure: false
  cookie_http_only: true
  cookie_same_site: Lax
  session_lifetime: 8h
  cleanup_interval: 30m

security:
  max_request_size: 52428800
  request_timeout: 60s
  enable_csrf: false
  enable_xss_protect: true
  enable_hsts: false
  production_mode: false
  allowed_origins:
    - "*"
  max_form_fields: 10000

monitoring:
  enable_metrics: true
  metrics_path: /debug/metrics
  metrics_port: 9090
  enable_pprof: true
  pprof_path: /debug/pprof
  pprof_port: 6060
  enable_optimization: false
  require_auth: false

i18n:
  default_locale: en
  locales_dir: ./locales
  supported_locales:
    - en
  fallback_to_default: true

plugins:
  enabled: true
  directory: ./plugins

filesystem_root: ./data
```

### Testing Configuration

A test-optimized configuration using in-memory storage and fast timeouts for automated testing.

```go
package main

import (
    "crypto/rand"
    "log"
    "time"
    
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Generate encryption keys
    sessionKey := make([]byte, 32)
    rand.Read(sessionKey)
    
    config := pkg.FrameworkConfig{
        // Server configuration - fast timeouts for tests
        ServerConfig: pkg.ServerConfig{
            // Short timeouts for fast test execution
            ReadTimeout:     5 * time.Second,
            WriteTimeout:    5 * time.Second,
            IdleTimeout:     10 * time.Second,
            ShutdownTimeout: 5 * time.Second,
            
            // HTTP/1.1 only for test simplicity
            EnableHTTP1: true,
            EnableHTTP2: false,
            EnableQUIC:  false,
            
            // No TLS in tests
            EnableHSTS: false,
            
            // Minimal limits for tests
            MaxConnections: 10,
            MaxRequestSize: 1 * 1024 * 1024, // 1 MB
            
            // Disable monitoring in tests
            EnableMetrics: false,
            EnablePprof:   false,
        },
        
        // Database configuration - in-memory SQLite
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "sqlite",
            Database: ":memory:", // In-memory database for tests
            
            // Single connection for SQLite in-memory
            MaxOpenConns:    1,
            MaxIdleConns:    1,
            ConnMaxLifetime: 0,
        },
        
        // Cache configuration - small in-memory cache
        CacheConfig: pkg.CacheConfig{
            Type:       "memory",
            MaxSize:    1 * 1024 * 1024, // 1 MB
            DefaultTTL: 10 * time.Second, // Short TTL for tests
        },
        
        // Session configuration - memory storage for tests
        SessionConfig: pkg.SessionConfig{
            StorageType:     pkg.SessionStorageCache, // Use cache (memory)
            CookieName:      "test_session",
            CookieSecure:    false,
            CookieHTTPOnly:  true,
            CookieSameSite:  "Lax",
            SessionLifetime: 1 * time.Hour,
            EncryptionKey:   sessionKey,
            CleanupInterval: 5 * time.Minute,
        },
        
        // Security configuration - minimal for tests
        SecurityConfig: pkg.SecurityConfig{
            MaxRequestSize: 1 * 1024 * 1024,
            RequestTimeout: 5 * time.Second,
            
            // Disable CSRF for easier testing
            EnableCSRF: false,
            
            // Disable XSS protection for testing
            EnableXSSProtect: false,
            
            // No HSTS in tests
            EnableHSTS: false,
            
            // Test mode - show all errors
            ProductionMode: false,
            
            // Allow all origins
            AllowedOrigins: []string{"*"},
            
            // Minimal limits
            MaxFormFields: 100,
        },
        
        // Monitoring configuration - disabled for tests
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics:      false,
            EnablePprof:        false,
            EnableOptimization: false,
            RequireAuth:        false,
        },
        
        // Proxy configuration - disabled for tests
        ProxyConfig: pkg.ProxyConfig{
            CircuitBreakerEnabled: false,
            CacheEnabled:          false,
            HealthCheckEnabled:    false,
        },
        
        // I18n configuration - minimal
        I18nConfig: pkg.I18nConfig{
            DefaultLocale:     "en",
            LocalesDir:        "./test/locales",
            SupportedLocales:  []string{"en"},
            FallbackToDefault: true,
        },
        
        // Plugin configuration - disabled for tests
        EnablePlugins: false,
        
        // File system - temporary directory
        FileSystemRoot: "./test/data",
    }
    
    // Create framework instance
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal("Failed to initialize framework:", err)
    }
    
    // Start server for integration tests
    log.Println("Starting test server on :8080")
    log.Fatal(app.Listen(":8080"))
}
```

**Testing Configuration (YAML)**

```yaml
# config/testing.yaml
# Test-optimized configuration with in-memory storage and fast timeouts

server:
  # Short timeouts for fast tests
  read_timeout: 5s
  write_timeout: 5s
  idle_timeout: 10s
  shutdown_timeout: 5s
  
  # HTTP/1.1 only
  enable_http1: true
  enable_http2: false
  enable_quic: false
  
  # No TLS
  enable_hsts: false
  
  # Minimal limits
  max_connections: 10
  max_request_size: 1048576  # 1 MB
  
  # Disable monitoring
  enable_metrics: false
  enable_pprof: false

database:
  driver: sqlite
  database: ":memory:"  # In-memory database
  max_open_conns: 1
  max_idle_conns: 1

cache:
  type: memory
  max_size: 1048576  # 1 MB
  default_ttl: 10s

session:
  storage_type: cache
  cookie_name: test_session
  cookie_secure: false
  cookie_http_only: true
  cookie_same_site: Lax
  session_lifetime: 1h
  cleanup_interval: 5m

security:
  max_request_size: 1048576
  request_timeout: 5s
  enable_csrf: false
  enable_xss_protect: false
  enable_hsts: false
  production_mode: false
  allowed_origins:
    - "*"
  max_form_fields: 100

monitoring:
  enable_metrics: false
  enable_pprof: false
  enable_optimization: false
  require_auth: false

proxy:
  circuit_breaker_enabled: false
  cache_enabled: false
  health_check_enabled: false

i18n:
  default_locale: en
  locales_dir: ./test/locales
  supported_locales:
    - en
  fallback_to_default: true

plugins:
  enabled: false

filesystem_root: ./test/data
```

### Multi-Tenant Configuration

A multi-tenant configuration with host-based routing and tenant isolation.

```go
package main

import (
    "crypto/rand"
    "crypto/tls"
    "log"
    "time"
    
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
    // Generate encryption keys
    sessionKey := make([]byte, 32)
    rand.Read(sessionKey)
    
    // TLS configuration
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
    }
    
    config := pkg.FrameworkConfig{
        // Server configuration with multi-tenant support
        ServerConfig: pkg.ServerConfig{
            ReadTimeout:     30 * time.Second,
            WriteTimeout:    30 * time.Second,
            IdleTimeout:     120 * time.Second,
            ShutdownTimeout: 30 * time.Second,
            
            EnableHTTP1: true,
            EnableHTTP2: true,
            
            TLSConfig:  tlsConfig,
            EnableHSTS: true,
            HSTSMaxAge: 365 * 24 * time.Hour,
            
            MaxConnections: 10000,
            MaxRequestSize: 10 * 1024 * 1024,
            
            // Host-specific configurations for multi-tenancy
            HostConfigs: map[string]*pkg.HostConfig{
                "tenant1.example.com": {
                    Hostname: "tenant1.example.com",
                    TenantID: "tenant1",
                    RateLimits: &pkg.RateLimitConfig{
                        Enabled:           true,
                        RequestsPerSecond: 100,
                        BurstSize:         200,
                        Storage:           "database",
                    },
                },
                "tenant2.example.com": {
                    Hostname: "tenant2.example.com",
                    TenantID: "tenant2",
                    RateLimits: &pkg.RateLimitConfig{
                        Enabled:           true,
                        RequestsPerSecond: 50,
                        BurstSize:         100,
                        Storage:           "database",
                    },
                },
                "premium.example.com": {
                    Hostname: "premium.example.com",
                    TenantID: "premium",
                    RateLimits: &pkg.RateLimitConfig{
                        Enabled:           true,
                        RequestsPerSecond: 1000,
                        BurstSize:         2000,
                        Storage:           "database",
                    },
                },
            },
            
            EnableMetrics: true,
            MetricsPath:   "/metrics",
        },
        
        // Database configuration - shared database with tenant isolation
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "postgres",
            Host:     "db.example.com",
            Port:     5432,
            Database: "multitenant_app",
            Username: "app_user",
            Password: "secure_password",
            SSLMode:  "require",
            
            // Larger connection pool for multi-tenant workload
            MaxOpenConns:    100,
            MaxIdleConns:    20,
            ConnMaxLifetime: 5 * time.Minute,
            
            Options: map[string]string{
                "application_name": "multitenant_app",
            },
        },
        
        // Cache configuration - larger cache for multiple tenants
        CacheConfig: pkg.CacheConfig{
            Type:       "memory",
            MaxSize:    1024 * 1024 * 1024, // 1 GB
            DefaultTTL: 5 * time.Minute,
        },
        
        // Session configuration - tenant-isolated sessions
        SessionConfig: pkg.SessionConfig{
            StorageType:     pkg.SessionStorageDatabase,
            CookieName:      "mt_session",
            CookiePath:      "/",
            CookieSecure:    true,
            CookieHTTPOnly:  true,
            CookieSameSite:  "Strict",
            SessionLifetime: 24 * time.Hour,
            EncryptionKey:   sessionKey,
            CleanupInterval: 1 * time.Hour,
        },
        
        // Security configuration
        SecurityConfig: pkg.SecurityConfig{
            MaxRequestSize:   10 * 1024 * 1024,
            RequestTimeout:   30 * time.Second,
            EnableCSRF:       true,
            CSRFTokenExpiry:  24 * time.Hour,
            EnableXSSProtect: true,
            EnableHSTS:       true,
            HSTSMaxAge:       31536000,
            ProductionMode:   true,
            
            // Allow all tenant domains
            AllowedOrigins: []string{
                "https://tenant1.example.com",
                "https://tenant2.example.com",
                "https://premium.example.com",
            },
        },
        
        // Monitoring configuration
        MonitoringConfig: pkg.MonitoringConfig{
            EnableMetrics:        true,
            MetricsPort:          9090,
            EnableOptimization:   true,
            OptimizationInterval: 5 * time.Minute,
            RequireAuth:          true,
            AuthToken:            "monitoring_token",
        },
        
        // I18n configuration - support multiple languages
        I18nConfig: pkg.I18nConfig{
            DefaultLocale:     "en",
            LocalesDir:        "./locales",
            SupportedLocales:  []string{"en", "de", "fr", "es", "ja", "zh"},
            FallbackToDefault: true,
        },
        
        // Plugin configuration
        EnablePlugins:    true,
        PluginConfigPath: "./config/plugins.yaml",
        
        // File system - tenant-isolated storage
        FileSystemRoot: "/var/app/tenants",
    }
    
    // Create framework instance
    app, err := pkg.New(config)
    if err != nil {
        log.Fatal("Failed to initialize framework:", err)
    }
    
    // Add tenant-specific middleware
    app.Router().Use(func(ctx pkg.Context) error {
        // Extract tenant ID from hostname
        hostname := ctx.Request().Host
        tenantID := extractTenantID(hostname)
        
        // Set tenant context
        ctx.Set("tenant_id", tenantID)
        
        // Log tenant request
        log.Printf("Request from tenant: %s", tenantID)
        
        return ctx.Next()
    })
    
    // Start server
    log.Println("Starting multi-tenant server on :443")
    log.Fatal(app.ListenTLS(":443", "/etc/ssl/certs/server.crt", "/etc/ssl/private/server.key"))
}

func extractTenantID(hostname string) string {
    // Extract tenant ID from hostname
    // Example: tenant1.example.com -> tenant1
    // Implement your tenant extraction logic here
    return "default"
}
```

**Multi-Tenant Configuration (YAML)**

```yaml
# config/multitenant.yaml
# Multi-tenant configuration with host-based routing and tenant isolation

server:
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  shutdown_timeout: 30s
  
  enable_http1: true
  enable_http2: true
  
  enable_hsts: true
  hsts_max_age: 31536000
  hsts_include_subdomains: true
  
  max_connections: 10000
  max_request_size: 10485760
  
  # Host-specific configurations for each tenant
  host_configs:
    tenant1.example.com:
      hostname: tenant1.example.com
      tenant_id: tenant1
      rate_limits:
        enabled: true
        requests_per_second: 100
        burst_size: 200
        storage: database
    
    tenant2.example.com:
      hostname: tenant2.example.com
      tenant_id: tenant2
      rate_limits:
        enabled: true
        requests_per_second: 50
        burst_size: 100
        storage: database
    
    premium.example.com:
      hostname: premium.example.com
      tenant_id: premium
      rate_limits:
        enabled: true
        requests_per_second: 1000
        burst_size: 2000
        storage: database
  
  enable_metrics: true
  metrics_path: /metrics

database:
  driver: postgres
  host: db.example.com
  port: 5432
  database: multitenant_app
  username: ${DB_USER}
  password: ${DB_PASSWORD}
  ssl_mode: require
  
  # Larger connection pool for multi-tenant workload
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 5m
  
  options:
    application_name: multitenant_app

cache:
  type: memory
  max_size: 1073741824  # 1 GB
  default_ttl: 5m

session:
  storage_type: database
  cookie_name: mt_session
  cookie_path: /
  cookie_secure: true
  cookie_http_only: true
  cookie_same_site: Strict
  session_lifetime: 24h
  cleanup_interval: 1h

security:
  max_request_size: 10485760
  request_timeout: 30s
  enable_csrf: true
  csrf_token_expiry: 24h
  enable_xss_protect: true
  enable_hsts: true
  hsts_max_age: 31536000
  production_mode: true
  allowed_origins:
    - https://tenant1.example.com
    - https://tenant2.example.com
    - https://premium.example.com

monitoring:
  enable_metrics: true
  metrics_port: 9090
  enable_optimization: true
  optimization_interval: 5m
  require_auth: true
  auth_token: ${MONITORING_TOKEN}

i18n:
  default_locale: en
  locales_dir: ./locales
  supported_locales:
    - en
    - de
    - fr
    - es
    - ja
    - zh
  fallback_to_default: true

plugins:
  enabled: true
  directory: ./plugins

filesystem_root: /var/app/tenants
```

### Configuration Validation and Troubleshooting

This section covers common configuration errors, validation approaches, and troubleshooting strategies.

#### Common Configuration Errors

**1. Database Connection Failures**

**Symptoms:**
- Application fails to start with "connection refused" error
- Timeout errors when connecting to database
- SSL/TLS handshake failures

**Common Causes:**
- Incorrect host or port
- Database server not running
- Firewall blocking connection
- Invalid credentials
- SSL mode mismatch

**Solutions:**
```yaml
# Verify database configuration
database:
  driver: postgres
  host: localhost  # Check hostname is correct
  port: 5432       # Verify port number
  database: myapp  # Ensure database exists
  username: user   # Check credentials
  password: pass
  ssl_mode: disable  # Try disable first, then require

# Test connection with psql
# psql -h localhost -p 5432 -U user -d myapp

# Check if database server is running
# systemctl status postgresql  # Linux
# brew services list           # macOS
```

**2. Port Already in Use**

**Symptoms:**
- "bind: address already in use" error
- Application fails to start

**Common Causes:**
- Another process using the same port
- Previous instance not properly shut down
- Port conflict with system services

**Solutions:**
```bash
# Find process using port 8080
# Linux/macOS:
lsof -i :8080
netstat -tulpn | grep 8080

# Windows:
netstat -ano | findstr :8080

# Kill the process or change port
server:
  # Use a different port
  listen_address: :8081
```

**3. Session/Cookie Encryption Key Errors**

**Symptoms:**
- "invalid key size" error
- Sessions not persisting
- Cookie decryption failures

**Common Causes:**
- Encryption key not 32 bytes (AES-256)
- Key not properly generated
- Key changed between restarts

**Solutions:**
```go
// Generate proper 32-byte key
sessionKey := make([]byte, 32)
_, err := rand.Read(sessionKey)
if err != nil {
    log.Fatal("Failed to generate session key:", err)
}

// Store key securely (environment variable or secrets manager)
// Don't regenerate on every restart!

// In YAML, use hex-encoded key
session:
  encryption_key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
```

**4. TLS Certificate Errors**

**Symptoms:**
- "certificate signed by unknown authority"
- "x509: certificate has expired"
- Browser shows security warnings

**Common Causes:**
- Self-signed certificate
- Expired certificate
- Certificate path incorrect
- Certificate doesn't match hostname

**Solutions:**
```bash
# Generate self-signed certificate for development
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Verify certificate
openssl x509 -in cert.pem -text -noout

# Check certificate expiration
openssl x509 -in cert.pem -noout -enddate

# Use Let's Encrypt for production
# certbot certonly --standalone -d example.com
```

**5. Memory/Resource Exhaustion**

**Symptoms:**
- Out of memory errors
- Slow performance
- Connection pool exhausted
- Cache eviction storms

**Common Causes:**
- Connection pool too large
- Cache size too large
- Too many concurrent connections
- Memory leaks

**Solutions:**
```yaml
# Tune connection pool
database:
  max_open_conns: 25  # Reduce if memory constrained
  max_idle_conns: 5   # Keep lower than max_open_conns
  conn_max_lifetime: 5m

# Limit cache size
cache:
  max_size: 104857600  # 100 MB, adjust based on available memory

# Limit concurrent connections
server:
  max_connections: 1000  # Reduce if memory constrained
  max_request_size: 10485760  # 10 MB
```

**6. CORS Configuration Issues**

**Symptoms:**
- Browser blocks requests with CORS error
- "Access-Control-Allow-Origin" header missing
- Preflight requests failing

**Common Causes:**
- Missing allowed origins
- Wildcard not working as expected
- Credentials mode mismatch

**Solutions:**
```yaml
security:
  # Specific origins (recommended for production)
  allowed_origins:
    - https://app.example.com
    - https://www.example.com
  
  # Allow all origins (development only)
  allowed_origins:
    - "*"
  
  # For credentials, must specify exact origin
  allowed_origins:
    - https://app.example.com  # Not "*"
```

#### Configuration Validation Checklist

Use this checklist before deploying:

**Pre-Deployment Validation:**

- [ ] Database connection tested and working
- [ ] All required environment variables set
- [ ] Encryption keys generated and stored securely (32 bytes for AES-256)
- [ ] TLS certificates valid and not expired
- [ ] Port numbers not conflicting with other services
- [ ] File system paths exist and have correct permissions
- [ ] Connection pool sizes appropriate for workload
- [ ] Timeouts configured appropriately
- [ ] Security settings enabled (CSRF, XSS, HSTS)
- [ ] Production mode enabled for production deployments
- [ ] Monitoring and metrics configured
- [ ] Log levels appropriate for environment
- [ ] Rate limiting configured if needed
- [ ] CORS origins properly configured
- [ ] Session storage configured and tested
- [ ] Cache size appropriate for available memory

**Production-Specific Validation:**

- [ ] HTTPS enabled with valid certificates
- [ ] HSTS enabled with appropriate max-age
- [ ] CSRF protection enabled
- [ ] Production mode enabled (hides error details)
- [ ] Debug endpoints (pprof) disabled
- [ ] Secrets loaded from secure storage (not hardcoded)
- [ ] Database SSL/TLS enabled
- [ ] Connection pools sized for production load
- [ ] Monitoring and alerting configured
- [ ] Backup and recovery procedures tested

#### Configuration Testing Approaches

**1. Unit Testing Configuration**

```go
func TestConfigValidation(t *testing.T) {
    config := pkg.FrameworkConfig{
        ServerConfig: pkg.ServerConfig{
            ReadTimeout: 30 * time.Second,
        },
        DatabaseConfig: pkg.DatabaseConfig{
            Driver:   "postgres",
            Host:     "localhost",
            Database: "test",
        },
    }
    
    // Apply defaults
    config.ServerConfig.ApplyDefaults()
    config.DatabaseConfig.ApplyDefaults()
    
    // Validate configuration
    err := config.Validate()
    if err != nil {
        t.Fatalf("Configuration validation failed: %v", err)
    }
}
```

**2. Integration Testing**

```go
func TestDatabaseConnection(t *testing.T) {
    config := loadTestConfig()
    
    app, err := pkg.New(config)
    if err != nil {
        t.Fatalf("Failed to create app: %v", err)
    }
    
    // Test database connection
    db := app.Database()
    err = db.Ping()
    if err != nil {
        t.Fatalf("Database connection failed: %v", err)
    }
}
```

**3. Configuration Dry-Run**

```go
// Add a --validate flag to your application
func main() {
    validateOnly := flag.Bool("validate", false, "Validate configuration and exit")
    flag.Parse()
    
    config := loadConfig()
    
    if *validateOnly {
        err := config.Validate()
        if err != nil {
            log.Fatal("Configuration validation failed:", err)
        }
        log.Println("Configuration is valid")
        os.Exit(0)
    }
    
    // Normal startup
    app, err := pkg.New(config)
    // ...
}
```

**4. Environment-Specific Testing**

```bash
# Test configuration loading
./myapp --validate --config=config/production.yaml

# Test with environment variables
export DB_HOST=localhost
export DB_PORT=5432
./myapp --validate

# Test database connection
./myapp --test-db --config=config/production.yaml
```

#### Troubleshooting Decision Tree

```
Configuration Issue?
│
├─ Application won't start?
│  ├─ Port conflict? → Check with lsof/netstat, change port
│  ├─ Database error? → Verify connection settings, test with CLI
│  ├─ TLS error? → Check certificate paths and validity
│  └─ Permission error? → Check file/directory permissions
│
├─ Application starts but behaves incorrectly?
│  ├─ Sessions not working? → Check encryption key (32 bytes)
│  ├─ CORS errors? → Verify allowed_origins configuration
│  ├─ Slow performance? → Check connection pool sizes, cache config
│  └─ Memory issues? → Reduce pool sizes, cache size, max connections
│
├─ Security warnings?
│  ├─ HTTPS issues? → Verify TLS configuration and certificates
│  ├─ HSTS warnings? → Check HSTS settings and max-age
│  └─ CSRF errors? → Verify CSRF is enabled and tokens valid
│
└─ Monitoring not working?
   ├─ Metrics not available? → Check enable_metrics and metrics_port
   ├─ Can't access endpoints? → Verify require_auth and auth_token
   └─ No data showing? → Check monitoring_interval and storage
```

#### Best Practices for Configuration Management

1. **Use Environment-Specific Files**
   - Separate files for dev, staging, production
   - Base configuration with environment overrides
   - Never commit secrets to version control

2. **Validate Early**
   - Validate configuration on startup
   - Fail fast with clear error messages
   - Use --validate flag for pre-deployment checks

3. **Use Secrets Management**
   - Store secrets in environment variables or secrets manager
   - Never hardcode passwords or keys
   - Rotate secrets regularly

4. **Document Your Configuration**
   - Comment all non-obvious settings
   - Document why specific values were chosen
   - Keep configuration documentation up to date

5. **Test Configuration Changes**
   - Test in non-production environment first
   - Use configuration validation tools
   - Monitor after deployment for issues

6. **Monitor Configuration**
   - Log configuration on startup (without secrets)
   - Alert on configuration changes
   - Track configuration versions

---

## Configuration Best Practices

### 1. Use Environment-Specific Files

Separate configuration files for different environments:

```
config/
├── base.yaml          # Common settings
├── development.yaml   # Development overrides
├── staging.yaml       # Staging overrides
└── production.yaml    # Production overrides
```

### 2. Secure Sensitive Data

Never commit sensitive data to version control:

```yaml
# Use environment variables for secrets
database:
  username: ${DB_USER}
  password: ${DB_PASSWORD}

security:
  jwt_secret: ${JWT_SECRET}
```

### 3. Apply Defaults

The framework automatically applies sensible defaults:

```go
config.ServerConfig.ApplyDefaults()
config.DatabaseConfig.ApplyDefaults()
config.CacheConfig.ApplyDefaults()
```

### 4. Validate Configuration

Always validate configuration before starting:

```go
config := pkg.NewConfigManager()
err := config.Load("config/app.yaml")
if err != nil {
    log.Fatal(err)
}

err = config.Validate()
if err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

### 5. Use Configuration Watching

Automatically reload configuration on changes:

```go
config := pkg.NewConfigManager()
config.Load("config/app.yaml")

// Watch for changes
config.Watch(func() {
    log.Println("Configuration reloaded")
    // Update application settings
})
```

---

## See Also

- [Getting Started Guide](GETTING_STARTED.md)
- [API Reference](API_REFERENCE.md)
- [Security Guide](SECURITY.md)
- [Plugin Development](PLUGIN_DEVELOPMENT.md)
- [Examples](../examples/)
