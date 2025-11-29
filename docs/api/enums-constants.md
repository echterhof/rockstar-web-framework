# Enums and Constants Reference

Complete reference for all enumerations and constants in the Rockstar Web Framework.

## Overview

This document provides formal documentation for all enum types and constant values used throughout the framework. Enums provide type-safe values for configuration and state management.

## Table of Contents

- [Pipeline Enums](#pipeline-enums)
- [Plugin System Enums](#plugin-system-enums)
- [Middleware Enums](#middleware-enums)
- [Session Enums](#session-enums)
- [Security Enums](#security-enums)
- [Protocol Enums](#protocol-enums)
- [Error Code Constants](#error-code-constants)
- [Permission Constants](#permission-constants)

---

## Pipeline Enums

### PipelineResult

```go
type PipelineResult int
```

Represents the result of a pipeline execution, controlling flow to the next stage.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `PipelineResultContinue` | 0 | Continue to the next pipeline stage or handler |
| `PipelineResultClose` | 1 | Stop pipeline execution and close connection |
| `PipelineResultChain` | 2 | Chain to another pipeline (specified in config) |
| `PipelineResultView` | 3 | Execute a view handler |

**Usage**:
```go
func validationStage(ctx pkg.Context) (pkg.PipelineResult, error) {
    if !isValid(ctx) {
        // Stop pipeline on validation failure
        return pkg.PipelineResultClose, pkg.NewValidationError("Invalid input", "body")
    }
    
    // Continue to next stage
    return pkg.PipelineResultContinue, nil
}

func processingStage(ctx pkg.Context) (pkg.PipelineResult, error) {
    result := process(ctx)
    
    // Chain to another pipeline for further processing
    if needsMoreProcessing(result) {
        return pkg.PipelineResultChain, nil
    }
    
    // Execute view to render response
    return pkg.PipelineResultView, nil
}
```

**Best Practices**:
- Use `PipelineResultContinue` for normal flow
- Use `PipelineResultClose` for fatal errors or early termination
- Use `PipelineResultChain` for complex multi-stage workflows
- Use `PipelineResultView` when ready to render response

---

## Plugin System Enums

### PluginStatus

```go
type PluginStatus string
```

Represents the current lifecycle state of a plugin.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `PluginStatusUnloaded` | "unloaded" | Plugin not yet loaded into memory |
| `PluginStatusLoading` | "loading" | Plugin is being loaded |
| `PluginStatusInitialized` | "initialized" | Plugin loaded and initialized |
| `PluginStatusRunning` | "running" | Plugin is actively running |
| `PluginStatusStopped` | "stopped" | Plugin has been stopped |
| `PluginStatusError` | "error" | Plugin encountered an error |

**State Transitions**:
```
Unloaded → Loading → Initialized → Running → Stopped
                ↓         ↓           ↓
              Error     Error       Error
```

**Usage**:
```go
func checkPluginStatus(manager pkg.PluginManager, name string) error {
    info, err := manager.GetPluginInfo(name)
    if err != nil {
        return err
    }
    
    switch info.Status {
    case pkg.PluginStatusRunning:
        fmt.Println("Plugin is running")
    case pkg.PluginStatusError:
        fmt.Println("Plugin has errors")
    case pkg.PluginStatusStopped:
        fmt.Println("Plugin is stopped")
    default:
        fmt.Println("Plugin status:", info.Status)
    }
    
    return nil
}
```

**Best Practices**:
- Check status before plugin operations
- Handle `PluginStatusError` gracefully
- Monitor status transitions for debugging
- Log status changes for audit trail

---

### HookType

```go
type HookType string
```

Defines the type of lifecycle hook for plugin event handling.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `HookTypeStartup` | "startup" | Called when plugin starts |
| `HookTypeShutdown` | "shutdown" | Called when plugin stops |
| `HookTypePreRequest` | "pre_request" | Called before request processing |
| `HookTypePostRequest` | "post_request" | Called after request processing |
| `HookTypePreResponse` | "pre_response" | Called before response sent |
| `HookTypePostResponse` | "post_response" | Called after response sent |
| `HookTypeError` | "error" | Called when error occurs |

**Execution Order**:
```
Startup
  ↓
PreRequest → PostRequest → PreResponse → PostResponse
                ↓
              Error (if error occurs)
  ↓
Shutdown
```

**Usage**:
```go
func (p *MyPlugin) Initialize(ctx pkg.PluginContext) error {
    // Register startup hook
    ctx.RegisterHook(pkg.HookTypeStartup, 100, func(hctx pkg.HookContext) error {
        p.logger.Info("Plugin starting up")
        return p.initializeResources()
    })
    
    // Register request logging hook
    ctx.RegisterHook(pkg.HookTypePreRequest, 50, func(hctx pkg.HookContext) error {
        reqCtx := hctx.Context()
        p.logger.Info("Request received", "path", reqCtx.Request().Path)
        return nil
    })
    
    // Register error handling hook
    ctx.RegisterHook(pkg.HookTypeError, 200, func(hctx pkg.HookContext) error {
        p.logger.Error("Error occurred in request")
        return nil
    })
    
    return nil
}
```

**Priority Guidelines**:
- 0-99: Low priority (logging, metrics)
- 100-199: Normal priority (business logic)
- 200-299: High priority (security, validation)
- 300+: Critical priority (authentication)

---

## Middleware Enums

### MiddlewarePosition

```go
type MiddlewarePosition int
```

Defines when middleware should execute relative to the handler.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `MiddlewarePositionBefore` | 0 | Execute before the handler |
| `MiddlewarePositionAfter` | 1 | Execute after the handler |

**Execution Flow**:
```
Before Middleware (Priority: High → Low)
         ↓
      Handler
         ↓
After Middleware (Priority: Low → High)
```

**Usage**:
```go
// Authentication middleware (before handler)
authConfig := pkg.MiddlewareConfig{
    Name: "auth",
    Handler: authMiddleware,
    Position: pkg.MiddlewarePositionBefore,
    Priority: 100,
    Enabled: true,
}

// Logging middleware (after handler)
logConfig := pkg.MiddlewareConfig{
    Name: "logging",
    Handler: loggingMiddleware,
    Position: pkg.MiddlewarePositionAfter,
    Priority: 50,
    Enabled: true,
}

engine.Register(authConfig)
engine.Register(logConfig)
```

**Best Practices**:
- Use `Before` for authentication, validation, rate limiting
- Use `After` for logging, metrics, cleanup
- Higher priority executes first for `Before`
- Higher priority executes last for `After`

---

## Session Enums

### SessionStorageType

```go
type SessionStorageType string
```

Defines the backend storage type for session data.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `SessionStorageDatabase` | "database" | Store sessions in database (persistent) |
| `SessionStorageCache` | "cache" | Store sessions in cache (fast, volatile) |
| `SessionStorageFilesystem` | "filesystem" | Store sessions in files (persistent) |

**Comparison**:

| Storage Type | Persistence | Speed | Scalability | Use Case |
|--------------|-------------|-------|-------------|----------|
| Database | High | Medium | High | Production, multi-server |
| Cache | Low | High | High | High-performance, temporary |
| Filesystem | High | Low | Low | Development, single-server |

**Usage**:
```go
// Production: Database storage
config := pkg.SessionConfig{
    StorageType: pkg.SessionStorageDatabase,
    SessionLifetime: 24 * time.Hour,
}

// High-performance: Cache storage
config := pkg.SessionConfig{
    StorageType: pkg.SessionStorageCache,
    SessionLifetime: 1 * time.Hour,
}

// Development: Filesystem storage
config := pkg.SessionConfig{
    StorageType: pkg.SessionStorageFilesystem,
    FilesystemPath: "./sessions",
    SessionLifetime: 24 * time.Hour,
}
```

**Best Practices**:
- Use `Database` for production with multiple servers
- Use `Cache` for high-traffic applications with short sessions
- Use `Filesystem` only for development/testing
- Consider session lifetime based on storage type

---

## Security Enums

### PasswordHashAlgorithm

```go
type PasswordHashAlgorithm string
```

Represents the hashing algorithm for password storage.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `AlgorithmBcrypt` | "bcrypt" | Bcrypt algorithm (recommended for most cases) |
| `AlgorithmArgon2id` | "argon2id" | Argon2id algorithm (high-security applications) |

**Comparison**:

| Algorithm | Security | Speed | Memory | Use Case |
|-----------|----------|-------|--------|----------|
| Bcrypt | High | Medium | Low | General purpose, web apps |
| Argon2id | Very High | Slow | High | High-security, sensitive data |

**Usage**:
```go
// Bcrypt (recommended for most applications)
hasher := pkg.NewBcryptHasher(12) // cost factor
hash, err := hasher.Hash("password123")

// Argon2id (high-security applications)
params := pkg.DefaultArgon2Params()
hasher := pkg.NewArgon2Hasher(params)
hash, err := hasher.Hash("password123")

// Verify password
valid, err := hasher.Verify("password123", hash)
```

**Best Practices**:
- Use `Bcrypt` for most web applications
- Use `Argon2id` for high-security requirements
- Adjust cost/parameters based on hardware
- Regularly review and update parameters

---

## Protocol Enums

### SOAPVersion

```go
type SOAPVersion int
```

Represents SOAP protocol version.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `SOAP11` | 11 | SOAP 1.1 protocol |
| `SOAP12` | 12 | SOAP 1.2 protocol |

**Differences**:

| Feature | SOAP 1.1 | SOAP 1.2 |
|---------|----------|----------|
| Namespace | `http://schemas.xmlsoap.org/soap/envelope/` | `http://www.w3.org/2003/05/soap-envelope` |
| Fault Codes | `Client`, `Server` | `Sender`, `Receiver` |
| HTTP Binding | Required | Optional |
| Error Handling | Basic | Enhanced |

**Usage**:
```go
// SOAP 1.1 service
service := &MySOAPService{
    version: pkg.SOAP11,
}

// SOAP 1.2 service
service := &MySOAPService{
    version: pkg.SOAP12,
}

// Version detection
func detectSOAPVersion(envelope []byte) pkg.SOAPVersion {
    if bytes.Contains(envelope, []byte("soap-envelope")) {
        return pkg.SOAP12
    }
    return pkg.SOAP11
}
```

---

### SOAPFaultCode

```go
type SOAPFaultCode string
```

Standard SOAP fault codes for error responses.

**SOAP 1.1 Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `FaultCodeVersionMismatch` | "VersionMismatch" | SOAP version not supported |
| `FaultCodeMustUnderstand` | "MustUnderstand" | Required header not understood |
| `FaultCodeClient` | "Client" | Client error (invalid request) |
| `FaultCodeServer` | "Server" | Server error (processing failed) |

**SOAP 1.2 Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `FaultCodeSender` | "Sender" | Sender error (replaces Client) |
| `FaultCodeReceiver` | "Receiver" | Receiver error (replaces Server) |
| `FaultCodeDataEncodingUnknown` | "DataEncodingUnknown" | Unknown data encoding |

**Usage**:
```go
// Create SOAP fault
fault := pkg.NewSOAPFault(pkg.FaultCodeClient, "Invalid request format")

// With additional details
fault := pkg.NewSOAPFault(pkg.FaultCodeServer, "Database connection failed").
    WithActor("http://example.com/service").
    WithDetail("Connection timeout after 30 seconds")

// Return fault in response
return ctx.XML(500, &pkg.SOAPEnvelope{
    Body: pkg.SOAPBody{
        Fault: fault,
    },
})
```

---

### GRPCStatusCode

```go
type GRPCStatusCode int
```

Standard gRPC status codes for RPC responses.

**Values**:

| Constant | Value | HTTP Equivalent | Description |
|----------|-------|-----------------|-------------|
| `GRPCStatusOK` | 0 | 200 | Success |
| `GRPCStatusCanceled` | 1 | 499 | Operation canceled |
| `GRPCStatusUnknown` | 2 | 500 | Unknown error |
| `GRPCStatusInvalidArgument` | 3 | 400 | Invalid argument |
| `GRPCStatusDeadlineExceeded` | 4 | 504 | Deadline exceeded |
| `GRPCStatusNotFound` | 5 | 404 | Not found |
| `GRPCStatusAlreadyExists` | 6 | 409 | Already exists |
| `GRPCStatusPermissionDenied` | 7 | 403 | Permission denied |
| `GRPCStatusResourceExhausted` | 8 | 429 | Resource exhausted |
| `GRPCStatusFailedPrecondition` | 9 | 400 | Failed precondition |
| `GRPCStatusAborted` | 10 | 409 | Aborted |
| `GRPCStatusOutOfRange` | 11 | 400 | Out of range |
| `GRPCStatusUnimplemented` | 12 | 501 | Not implemented |
| `GRPCStatusInternal` | 13 | 500 | Internal error |
| `GRPCStatusUnavailable` | 14 | 503 | Service unavailable |
| `GRPCStatusDataLoss` | 15 | 500 | Data loss |
| `GRPCStatusUnauthenticated` | 16 | 401 | Unauthenticated |

**Usage**:
```go
// Return gRPC error
func getUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := fetchUser(req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, pkg.NewGRPCError(pkg.GRPCStatusNotFound, "User not found")
        }
        return nil, pkg.NewGRPCError(pkg.GRPCStatusInternal, "Failed to fetch user")
    }
    return user, nil
}

// With details
err := pkg.NewGRPCError(pkg.GRPCStatusInvalidArgument, "Invalid user ID").
    WithDetails(map[string]interface{}{
        "field": "id",
        "reason": "must be positive integer",
    })
```

**Best Practices**:
- Use appropriate status codes for errors
- Provide detailed error messages
- Include error details for debugging
- Map to HTTP status codes when needed

---

### CircuitState

```go
type CircuitState string
```

Represents the state of a circuit breaker for fault tolerance.

**Values**:

| Constant | Value | Description |
|----------|-------|-------------|
| `CircuitStateClosed` | "closed" | Normal operation, requests allowed |
| `CircuitStateOpen` | "open" | Failing, requests rejected immediately |
| `CircuitStateHalfOpen` | "half_open" | Testing if service recovered |

**State Transitions**:
```
Closed ──(failures exceed threshold)──> Open
  ↑                                       ↓
  │                              (timeout expires)
  │                                       ↓
  └──(success in half-open)──── HalfOpen
```

**Usage**:
```go
// Check circuit state before request
func makeRequest(ctx pkg.Context, backendID string) error {
    cb := ctx.Proxy().GetCircuitBreaker()
    
    state := cb.GetState(backendID)
    switch state {
    case pkg.CircuitStateOpen:
        return pkg.NewServiceUnavailableError("Circuit breaker open")
        
    case pkg.CircuitStateHalfOpen:
        // Allow limited requests to test recovery
        if !shouldAllowTestRequest() {
            return pkg.NewServiceUnavailableError("Circuit breaker testing")
        }
    }
    
    // Make request
    err := performRequest(backendID)
    if err != nil {
        cb.RecordFailure(backendID)
        return err
    }
    
    cb.RecordSuccess(backendID)
    return nil
}
```

**Configuration**:
```go
config := pkg.ProxyConfig{
    CircuitBreakerEnabled: true,
    CircuitBreakerThreshold: 5,              // Open after 5 failures
    CircuitBreakerTimeout: 60 * time.Second, // Try half-open after 60s
}
```

---

## Error Code Constants

### Authentication Error Codes

```go
const (
    ErrCodeAuthenticationFailed = "AUTH_FAILED"
    ErrCodeInvalidToken         = "INVALID_TOKEN"
    ErrCodeTokenExpired         = "TOKEN_EXPIRED"
    ErrCodeUnauthorized         = "UNAUTHORIZED"
)
```

**Usage**: See [Error Codes Reference](error-codes.md) for complete documentation.

---

### Authorization Error Codes

```go
const (
    ErrCodeForbidden           = "FORBIDDEN"
    ErrCodeInsufficientRoles   = "INSUFFICIENT_ROLES"
    ErrCodeInsufficientActions = "INSUFFICIENT_ACTIONS"
    ErrCodeInsufficientScopes  = "INSUFFICIENT_SCOPES"
)
```

---

### Validation Error Codes

```go
const (
    ErrCodeValidationFailed = "VALIDATION_FAILED"
    ErrCodeInvalidInput     = "INVALID_INPUT"
    ErrCodeMissingField     = "MISSING_FIELD"
    ErrCodeInvalidFormat    = "INVALID_FORMAT"
    ErrCodeFileTooLarge     = "FILE_TOO_LARGE"
    ErrCodeInvalidFileType  = "INVALID_FILE_TYPE"
)
```

---

### Security Error Codes

```go
const (
    ErrCodeCSRFTokenInvalid     = "CSRF_TOKEN_INVALID"
    ErrCodeXSSDetected          = "XSS_DETECTED"
    ErrCodeSQLInjectionDetected = "SQL_INJECTION_DETECTED"
    ErrCodePathTraversal        = "PATH_TRAVERSAL"
    ErrCodeRegexTimeout         = "REGEX_TIMEOUT"
)
```

---

### Database Error Codes

```go
const (
    ErrCodeDatabaseConnection   = "DATABASE_CONNECTION"
    ErrCodeDatabaseQuery        = "DATABASE_QUERY"
    ErrCodeDatabaseTransaction  = "DATABASE_TRANSACTION"
    ErrCodeRecordNotFound       = "RECORD_NOT_FOUND"
    ErrCodeDuplicateRecord      = "DUPLICATE_RECORD"
    ErrCodeNoDatabaseConfigured = "NO_DATABASE_CONFIGURED"
)
```

---

## Permission Constants

### Standard Plugin Permissions

```go
const (
    PermissionDatabase   = "database"
    PermissionCache      = "cache"
    PermissionConfig     = "config"
    PermissionRouter     = "router"
    PermissionFileSystem = "filesystem"
    PermissionNetwork    = "network"
    PermissionExec       = "exec"
)
```

**Usage**:
```go
// Check permission before operation
func (p *MyPlugin) accessDatabase(ctx pkg.PluginContext) error {
    if err := ctx.CheckPermission(p.Name(), pkg.PermissionDatabase); err != nil {
        return err
    }
    
    // Access database
    return ctx.Database().Query("SELECT * FROM users")
}

// Request permissions in manifest
manifest := pkg.PluginManifest{
    Name: "my-plugin",
    Permissions: []string{
        pkg.PermissionDatabase,
        pkg.PermissionCache,
        pkg.PermissionNetwork,
    },
}
```

**Permission Descriptions**:

| Permission | Description | Risk Level |
|------------|-------------|------------|
| `database` | Access database operations | High |
| `cache` | Access cache operations | Medium |
| `config` | Read configuration | Medium |
| `router` | Modify routing | High |
| `filesystem` | File system access | High |
| `network` | Network requests | High |
| `exec` | Execute commands | Critical |

---

## Configuration Schema Types

### ConfigSchemaType

```go
type ConfigSchemaType string

const (
    ConfigTypeString   ConfigSchemaType = "string"
    ConfigTypeInt      ConfigSchemaType = "int"
    ConfigTypeBool     ConfigSchemaType = "bool"
    ConfigTypeFloat    ConfigSchemaType = "float"
    ConfigTypeArray    ConfigSchemaType = "array"
    ConfigTypeObject   ConfigSchemaType = "object"
)
```

**Usage**:
```go
// Define plugin configuration schema
func (p *MyPlugin) ConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "api_key": map[string]interface{}{
            "type": pkg.ConfigTypeString,
            "required": true,
            "description": "API key for external service",
        },
        "timeout": map[string]interface{}{
            "type": pkg.ConfigTypeInt,
            "default": 30,
            "description": "Request timeout in seconds",
        },
        "enabled": map[string]interface{}{
            "type": pkg.ConfigTypeBool,
            "default": true,
            "description": "Enable plugin functionality",
        },
    }
}
```

---

## Best Practices

### 1. Type Safety

Always use enum constants instead of raw values:

```go
// Good
status := pkg.PluginStatusRunning

// Bad
status := "running"
```

### 2. Exhaustive Switch Statements

Handle all enum values:

```go
switch status {
case pkg.PluginStatusRunning:
    // Handle running
case pkg.PluginStatusStopped:
    // Handle stopped
case pkg.PluginStatusError:
    // Handle error
default:
    // Handle unexpected values
    return fmt.Errorf("unknown status: %s", status)
}
```

### 3. Validation

Validate enum values from external sources:

```go
func validateStorageType(t string) (pkg.SessionStorageType, error) {
    switch pkg.SessionStorageType(t) {
    case pkg.SessionStorageDatabase,
         pkg.SessionStorageCache,
         pkg.SessionStorageFilesystem:
        return pkg.SessionStorageType(t), nil
    default:
        return "", fmt.Errorf("invalid storage type: %s", t)
    }
}
```

### 4. Documentation

Document enum usage in configuration:

```go
// SessionConfig with documented enum values
type SessionConfig struct {
    // StorageType specifies the session storage backend.
    // Valid values: "database", "cache", "filesystem"
    // Default: "database"
    StorageType SessionStorageType `json:"storage_type"`
}
```

---

## See Also

- [Error Codes Reference](error-codes.md) - Complete error code catalog
- [Function Types Reference](function-types.md) - Function type signatures
- [Types Reference](types-reference.md) - Data types and structs
- [Configuration Reference](../CONFIGURATION_REFERENCE.md) - Configuration options

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0  
**Total Enums**: 10+  
**Total Constants**: 50+
