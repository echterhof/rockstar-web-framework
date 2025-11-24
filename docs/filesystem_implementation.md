# Virtual File System Implementation

## Overview

The Rockstar Web Framework provides a comprehensive virtual file system (VFS) implementation that supports both OS-based and in-memory file systems. The VFS enables per-host file serving, static file hosting, and flexible file operations through the framework's context.

## Features

- **Multiple Filesystem Types**: Support for OS-based and in-memory file systems
- **Per-Host Virtual Filesystems**: Different hosts can serve different file systems
- **Security**: Built-in directory traversal protection
- **HTTP Compatibility**: Implements `http.File` interface for seamless integration
- **Thread-Safe**: Concurrent access support with proper locking
- **Static File Serving**: Integrated with router for static file routes

## Architecture

### VirtualFS Interface

The `VirtualFS` interface is the core abstraction for all file system implementations:

```go
type VirtualFS interface {
    Open(name string) (http.File, error)
    Exists(name string) bool
}
```

### Implementations

#### 1. OS File System

The `osFileSystem` implementation provides access to the operating system's file system:

```go
fs := NewOSFileSystem("/path/to/root")
```

**Features:**
- Serves files from a specified root directory
- Directory traversal protection
- Standard file system operations

**Security:**
- Automatically prevents access outside the root directory
- Returns `ErrInvalidPath` for directory traversal attempts

#### 2. Memory File System

The `memoryFileSystem` implementation provides an in-memory file system:

```go
fs := NewMemoryFileSystem()
memFS := fs.(*memoryFileSystem)
memFS.AddFile("/index.html", []byte("<html>Hello</html>"))
memFS.AddDir("/assets")
```

**Features:**
- Fast in-memory file storage
- No disk I/O overhead
- Perfect for testing and temporary files
- Support for both files and directories

### FileManager Interface

The `FileManager` interface provides high-level file operations:

```go
type FileManager interface {
    Write(path string, data []byte) error
    Read(path string) ([]byte, error)
    Delete(path string) error
    Exists(path string) bool
    MkDir(path string) error
    ReadDir(path string) ([]fs.FileInfo, error)
    VirtualFS() VirtualFS
}
```

## Usage Examples

### Basic File Operations

```go
// Create a file manager with memory filesystem
vfs := NewMemoryFileSystem()
fm := NewFileManager(vfs)

// Write a file
err := fm.Write("/config.json", []byte(`{"key": "value"}`))

// Read a file
data, err := fm.Read("/config.json")

// Check if file exists
if fm.Exists("/config.json") {
    // File exists
}

// Delete a file
err := fm.Delete("/config.json")

// Create a directory
err := fm.MkDir("/uploads")
```

### Static File Serving

```go
// Create a router
router := NewRouter()

// Create a filesystem
fs := NewOSFileSystem("./public")

// Register static file route
router.Static("/static", fs)

// Now files in ./public are accessible at /static/*
// Example: ./public/css/style.css -> /static/css/style.css
```

### Per-Host Virtual Filesystems

```go
// Create different filesystems for different hosts
fs1 := NewMemoryFileSystem().(*memoryFileSystem)
fs1.AddFile("/index.html", []byte("<html>Host 1</html>"))

fs2 := NewMemoryFileSystem().(*memoryFileSystem)
fs2.AddFile("/index.html", []byte("<html>Host 2</html>"))

// Create router with host-specific routes
router := NewRouter()
router.Host("host1.com").Static("/", fs1)
router.Host("host2.com").Static("/", fs2)

// host1.com/index.html serves "Host 1"
// host2.com/index.html serves "Host 2"
```

### Host Filesystem Manager

```go
// Create a host filesystem manager
manager := NewHostFileSystemManager()

// Register filesystems for different hosts
fs1 := NewOSFileSystem("./sites/host1")
fs2 := NewOSFileSystem("./sites/host2")

manager.RegisterHost("host1.com", fs1)
manager.RegisterHost("host2.com", fs2)

// Retrieve filesystem for a specific host
if fs, exists := manager.GetFileSystem("host1.com"); exists {
    file, err := fs.Open("/index.html")
    // ...
}

// Unregister a host
manager.UnregisterHost("host1.com")
```

### Using FileManager in Context

```go
// In a handler function
func myHandler(ctx Context) error {
    fm := ctx.Files()
    
    // Write uploaded file
    fileData := ctx.Body()
    err := fm.Write("/uploads/file.txt", fileData)
    if err != nil {
        return err
    }
    
    // Read configuration
    config, err := fm.Read("/config.json")
    if err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]string{
        "status": "success",
        "config": string(config),
    })
}
```

## Multi-Tenancy Support

The virtual file system is designed to support multi-tenancy scenarios:

```go
// Create tenant-specific filesystems
tenantFS := make(map[string]VirtualFS)
tenantFS["tenant1"] = NewOSFileSystem("./tenants/tenant1")
tenantFS["tenant2"] = NewOSFileSystem("./tenants/tenant2")

// In middleware, set the appropriate filesystem based on tenant
func tenantMiddleware(ctx Context, next HandlerFunc) error {
    tenant := ctx.Tenant()
    if tenant != nil {
        if fs, ok := tenantFS[tenant.ID]; ok {
            // Set tenant-specific filesystem
            // (This would require extending the context implementation)
        }
    }
    return next(ctx)
}
```

## Security Considerations

### Directory Traversal Protection

Both OS and memory file systems include built-in protection against directory traversal attacks:

```go
fs := NewOSFileSystem("/var/www")

// This will return ErrInvalidPath
file, err := fs.Open("../../../etc/passwd")
```

### Path Cleaning

All paths are automatically cleaned using `filepath.Clean()` to normalize paths and remove relative components.

### Root Confinement

OS file systems are confined to their root directory. Any attempt to access files outside the root will fail with `ErrInvalidPath`.

## Performance Considerations

### Memory File System

- **Pros**: Extremely fast, no disk I/O, perfect for caching
- **Cons**: Limited by available RAM, data lost on restart
- **Use Cases**: Testing, temporary files, frequently accessed static assets

### OS File System

- **Pros**: Persistent storage, can handle large files
- **Cons**: Slower due to disk I/O
- **Use Cases**: Production static file serving, user uploads

### Concurrent Access

Both implementations use `sync.RWMutex` for thread-safe operations:
- Multiple concurrent reads are allowed
- Writes are exclusive and block other operations

## Error Handling

The virtual file system defines specific error types:

```go
var (
    ErrFileNotFound = errors.New("file not found")
    ErrInvalidPath  = errors.New("invalid path")
    ErrIsDirectory  = errors.New("is a directory")
)
```

Example error handling:

```go
file, err := fs.Open("/nonexistent.txt")
if err == ErrFileNotFound {
    // Handle file not found
} else if err == ErrInvalidPath {
    // Handle invalid path (security issue)
} else if err != nil {
    // Handle other errors
}
```

## Testing

The virtual file system includes comprehensive unit tests covering:

- OS file system operations
- Memory file system operations
- Per-host filesystem management
- Static file serving through router
- Concurrent access patterns
- Security (directory traversal prevention)

Run tests:

```bash
go test -v ./pkg -run TestOSFileSystem
go test -v ./pkg -run TestMemoryFileSystem
go test -v ./pkg -run TestHostFileSystemManager
go test -v ./pkg -run TestStaticFileServing
```

## Integration with Framework

The virtual file system integrates seamlessly with other framework components:

1. **Router**: Static file routes use VirtualFS for serving files
2. **Context**: FileManager accessible via `ctx.Files()`
3. **Server**: Host-specific filesystems configured in `HostConfig`
4. **Multi-tenancy**: Different tenants can have isolated file systems

## Best Practices

1. **Use Memory FS for Testing**: Memory file systems are perfect for unit tests
2. **Separate Static Assets**: Use different filesystems for different asset types
3. **Per-Host Isolation**: Always use per-host filesystems in multi-tenant scenarios
4. **Error Handling**: Always check for `ErrInvalidPath` to detect security issues
5. **Path Normalization**: Let the framework handle path cleaning automatically
6. **Concurrent Access**: The framework handles locking, but be aware of performance implications

## Future Enhancements

Potential future improvements:

- Caching layer for frequently accessed files
- Compression support (gzip, brotli)
- Content-Type detection
- ETag support for caching
- Range request support for large files
- Virtual filesystem composition (layered filesystems)
