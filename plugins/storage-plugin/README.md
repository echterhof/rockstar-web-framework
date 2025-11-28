# Storage Plugin

A compile-time plugin for the Rockstar Web Framework that provides file storage abstraction with support for local filesystem and S3-compatible backends.

## Features

- **Multiple Backends**: Support for local filesystem and S3-compatible storage
- **File Operations**: Store, retrieve, delete, list, and check file existence
- **Extension Filtering**: Restrict uploads to specific file types
- **Size Limiting**: Configure maximum file size
- **Service Export**: Export StorageService for use by other plugins
- **Public URLs**: Generate public URLs for stored files

## Configuration

### Local Storage

```yaml
plugins:
  storage-plugin:
    enabled: true
    config:
      storage_type: "local"
      local_base_path: "./storage"
      max_file_size: 10485760  # 10MB
      allowed_extensions:
        - ".jpg"
        - ".jpeg"
        - ".png"
        - ".gif"
        - ".pdf"
        - ".txt"
      public_url: "/storage"
    permissions:
      router: true
      filesystem: true
```

### S3 Storage

```yaml
plugins:
  storage-plugin:
    enabled: true
    config:
      storage_type: "s3"
      s3_bucket: "my-bucket"
      s3_region: "us-east-1"
      s3_access_key: "${AWS_ACCESS_KEY}"
      s3_secret_key: "${AWS_SECRET_KEY}"
      max_file_size: 10485760
      allowed_extensions:
        - ".jpg"
        - ".jpeg"
        - ".png"
      public_url: "/storage"
    permissions:
      router: true
      network: true
```

### S3-Compatible Services (MinIO, DigitalOcean Spaces, etc.)

```yaml
plugins:
  storage-plugin:
    enabled: true
    config:
      storage_type: "s3"
      s3_bucket: "my-bucket"
      s3_region: "us-east-1"
      s3_endpoint: "https://nyc3.digitaloceanspaces.com"
      s3_access_key: "${SPACES_ACCESS_KEY}"
      s3_secret_key: "${SPACES_SECRET_KEY}"
      max_file_size: 10485760
    permissions:
      router: true
      network: true
```

## Usage

### Using the Exported Service

Other plugins can import and use the StorageService:

```go
// Import the service
service, err := ctx.ImportService("storage-plugin", "StorageService")
if err != nil {
    return err
}
storageService := service.(*StorageService)

// Store a file
file, _ := os.Open("image.jpg")
defer file.Close()
err = storageService.Store("uploads/image.jpg", file)

// Retrieve a file
reader, err := storageService.Retrieve("uploads/image.jpg")
if err != nil {
    return err
}
defer reader.Close()

// Delete a file
err = storageService.Delete("uploads/image.jpg")

// Check if file exists
exists, err := storageService.Exists("uploads/image.jpg")

// List files with prefix
files, err := storageService.List("uploads/")

// Get public URL
url, err := storageService.GetURL("uploads/image.jpg")
```

### Creating Upload Endpoints

```go
// POST /api/upload - Upload a file
router.POST("/api/upload", func(ctx pkg.Context) error {
    // Get storage service
    service, _ := pluginCtx.ImportService("storage-plugin", "StorageService")
    storageService := service.(*StorageService)
    
    // Get file from request
    file, header, err := ctx.FormFile("file")
    if err != nil {
        return ctx.JSON(400, map[string]interface{}{"error": "No file provided"})
    }
    defer file.Close()
    
    // Generate unique filename
    filename := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), header.Filename)
    
    // Store file
    err = storageService.Store(filename, file)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{"error": err.Error()})
    }
    
    // Get public URL
    url, _ := storageService.GetURL(filename)
    
    return ctx.JSON(200, map[string]interface{}{
        "filename": filename,
        "url": url,
    })
})

// GET /api/files/:path - Download a file
router.GET("/api/files/:path", func(ctx pkg.Context) error {
    service, _ := pluginCtx.ImportService("storage-plugin", "StorageService")
    storageService := service.(*StorageService)
    
    path := ctx.Param("path")
    
    // Retrieve file
    reader, err := storageService.Retrieve(path)
    if err != nil {
        return ctx.JSON(404, map[string]interface{}{"error": "File not found"})
    }
    defer reader.Close()
    
    // Stream file to response
    return ctx.Stream(200, "application/octet-stream", reader)
})
```

## Storage Backends

### Local Filesystem

The local backend stores files in a directory on the server filesystem.

**Pros:**
- Simple setup
- No external dependencies
- Fast for small files
- No additional costs

**Cons:**
- Not suitable for distributed systems
- Limited by disk space
- No built-in redundancy

### S3-Compatible

The S3 backend stores files in Amazon S3 or S3-compatible services (MinIO, DigitalOcean Spaces, Backblaze B2, etc.).

**Pros:**
- Scalable and distributed
- Built-in redundancy
- Works across multiple servers
- CDN integration available

**Cons:**
- Requires external service
- Network latency
- Additional costs

## Security Considerations

1. **File Extension Validation**: Configure allowed extensions to prevent malicious uploads
2. **Size Limiting**: Set maximum file size to prevent DoS attacks
3. **Path Traversal**: The plugin sanitizes paths to prevent directory traversal
4. **Access Control**: Implement authentication/authorization in your upload endpoints
5. **Virus Scanning**: Consider integrating virus scanning for uploaded files
6. **HTTPS**: Always use HTTPS for file uploads and downloads

## Events

The plugin publishes the following events:

- `storage.file_uploaded` - When a file is successfully stored
- `storage.file_deleted` - When a file is deleted

## Best Practices

1. **Organize Files**: Use prefixes/folders to organize files (e.g., `uploads/2024/01/`)
2. **Unique Names**: Generate unique filenames to avoid collisions
3. **Cleanup**: Implement cleanup for temporary or expired files
4. **Monitoring**: Monitor storage usage and set up alerts
5. **Backups**: Regularly backup important files
6. **CDN**: Use a CDN for serving static files at scale

## Example: Image Upload with Validation

```go
func handleImageUpload(ctx pkg.Context, storageService *StorageService) error {
    // Get file
    file, header, err := ctx.FormFile("image")
    if err != nil {
        return ctx.JSON(400, map[string]interface{}{"error": "No image provided"})
    }
    defer file.Close()
    
    // Validate file size
    if header.Size > 5*1024*1024 { // 5MB
        return ctx.JSON(400, map[string]interface{}{"error": "File too large"})
    }
    
    // Validate file type
    ext := filepath.Ext(header.Filename)
    if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
        return ctx.JSON(400, map[string]interface{}{"error": "Invalid file type"})
    }
    
    // Generate unique filename
    filename := fmt.Sprintf("images/%s%s", uuid.New().String(), ext)
    
    // Store file
    err = storageService.Store(filename, file)
    if err != nil {
        return ctx.JSON(500, map[string]interface{}{"error": "Failed to store file"})
    }
    
    // Get URL
    url, _ := storageService.GetURL(filename)
    
    return ctx.JSON(200, map[string]interface{}{
        "filename": filename,
        "url": url,
    })
}
```

## Extending the Plugin

To add support for additional storage backends:

1. Implement the `StorageBackend` interface
2. Add configuration options for the new backend
3. Update `initializeBackend()` to handle the new type

```go
type MyCustomBackend struct {
    // Configuration fields
}

func (b *MyCustomBackend) Store(path string, data io.Reader) error {
    // Implementation
}

// Implement other StorageBackend methods...
```

## License

This plugin is part of the Rockstar Web Framework and is provided under the same license.
