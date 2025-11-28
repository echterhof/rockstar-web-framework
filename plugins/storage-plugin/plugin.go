package storageplugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Storage Plugin (Compile-Time)
// ============================================================================
//
// This plugin demonstrates:
// - File storage abstraction (local filesystem and S3-compatible)
// - Service export for other plugins
// - Multiple storage backend support
// - File upload/download handling
// - Compile-time plugin registration
//
// Requirements: 9.4, 14.1, 14.2, 14.3
// ============================================================================

// init registers the plugin at compile time
func init() {
	pkg.RegisterPlugin("storage-plugin", func() pkg.Plugin {
		return &StoragePlugin{}
	})
}

// StoragePlugin implements file storage functionality
type StoragePlugin struct {
	ctx pkg.PluginContext
	
	// Configuration
	storageType       string // "local" or "s3"
	localBasePath     string
	s3Bucket          string
	s3Region          string
	s3Endpoint        string
	s3AccessKey       string
	s3SecretKey       string
	maxFileSize       int64
	allowedExtensions []string
	publicURL         string
	
	// Storage backend
	backend StorageBackend
}

// StorageBackend is the interface for storage implementations
type StorageBackend interface {
	Store(path string, data io.Reader) error
	Retrieve(path string) (io.ReadCloser, error)
	Delete(path string) error
	Exists(path string) (bool, error)
	List(prefix string) ([]string, error)
	GetURL(path string) (string, error)
}

// StorageService is exported for other plugins to use
type StorageService struct {
	plugin *StoragePlugin
}

// Store saves a file to storage
func (s *StorageService) Store(path string, data io.Reader) error {
	return s.plugin.store(path, data)
}

// Retrieve gets a file from storage
func (s *StorageService) Retrieve(path string) (io.ReadCloser, error) {
	return s.plugin.retrieve(path)
}

// Delete removes a file from storage
func (s *StorageService) Delete(path string) error {
	return s.plugin.delete(path)
}

// Exists checks if a file exists in storage
func (s *StorageService) Exists(path string) (bool, error) {
	return s.plugin.exists(path)
}

// List returns files matching a prefix
func (s *StorageService) List(prefix string) ([]string, error) {
	return s.plugin.list(prefix)
}

// GetURL returns the public URL for a file
func (s *StorageService) GetURL(path string) (string, error) {
	return s.plugin.getURL(path)
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

func (p *StoragePlugin) Name() string {
	return "storage-plugin"
}

func (p *StoragePlugin) Version() string {
	return "1.0.0"
}

func (p *StoragePlugin) Description() string {
	return "File storage plugin with local filesystem and S3-compatible backends"
}

func (p *StoragePlugin) Author() string {
	return "Rockstar Framework Team"
}

func (p *StoragePlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

func (p *StoragePlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing storage plugin...\n", p.Name())
	
	p.ctx = ctx
	
	// Parse configuration
	if err := p.parseConfig(ctx.PluginConfig()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Initialize storage backend
	if err := p.initializeBackend(); err != nil {
		return fmt.Errorf("failed to initialize storage backend: %w", err)
	}
	
	// Register routes
	if err := p.registerRoutes(); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}
	
	// Export storage service
	if err := p.exportService(); err != nil {
		return fmt.Errorf("failed to export service: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

func (p *StoragePlugin) Start() error {
	fmt.Printf("[%s] Starting storage plugin...\n", p.Name())
	fmt.Printf("[%s] Storage type: %s\n", p.Name(), p.storageType)
	fmt.Printf("[%s] Plugin started\n", p.Name())
	return nil
}

func (p *StoragePlugin) Stop() error {
	fmt.Printf("[%s] Stopping storage plugin...\n", p.Name())
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

func (p *StoragePlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up storage plugin...\n", p.Name())
	
	p.ctx = nil
	p.backend = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

func (p *StoragePlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"storage_type": map[string]interface{}{
			"type":        "string",
			"default":     "local",
			"description": "Storage backend type (local or s3)",
		},
		"local_base_path": map[string]interface{}{
			"type":        "string",
			"default":     "./storage",
			"description": "Base path for local filesystem storage",
		},
		"s3_bucket": map[string]interface{}{
			"type":        "string",
			"description": "S3 bucket name",
		},
		"s3_region": map[string]interface{}{
			"type":        "string",
			"default":     "us-east-1",
			"description": "S3 region",
		},
		"s3_endpoint": map[string]interface{}{
			"type":        "string",
			"description": "S3 endpoint (for S3-compatible services)",
		},
		"s3_access_key": map[string]interface{}{
			"type":        "string",
			"description": "S3 access key",
		},
		"s3_secret_key": map[string]interface{}{
			"type":        "string",
			"description": "S3 secret key",
		},
		"max_file_size": map[string]interface{}{
			"type":        "int",
			"default":     10485760, // 10MB
			"description": "Maximum file size in bytes",
		},
		"allowed_extensions": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt"},
			"description": "Allowed file extensions",
		},
		"public_url": map[string]interface{}{
			"type":        "string",
			"default":     "/storage",
			"description": "Public URL prefix for accessing files",
		},
	}
}

func (p *StoragePlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	return p.parseConfig(config)
}

// ============================================================================
// Configuration Parsing
// ============================================================================

func (p *StoragePlugin) parseConfig(config map[string]interface{}) error {
	// Storage Type
	if val, ok := config["storage_type"].(string); ok {
		p.storageType = val
	} else {
		p.storageType = "local"
	}
	
	// Local Base Path
	if val, ok := config["local_base_path"].(string); ok {
		p.localBasePath = val
	} else {
		p.localBasePath = "./storage"
	}
	
	// S3 Configuration
	if val, ok := config["s3_bucket"].(string); ok {
		p.s3Bucket = val
	}
	if val, ok := config["s3_region"].(string); ok {
		p.s3Region = val
	} else {
		p.s3Region = "us-east-1"
	}
	if val, ok := config["s3_endpoint"].(string); ok {
		p.s3Endpoint = val
	}
	if val, ok := config["s3_access_key"].(string); ok {
		p.s3AccessKey = val
	}
	if val, ok := config["s3_secret_key"].(string); ok {
		p.s3SecretKey = val
	}
	
	// Max File Size
	if val, ok := config["max_file_size"].(int64); ok {
		p.maxFileSize = val
	} else if val, ok := config["max_file_size"].(float64); ok {
		p.maxFileSize = int64(val)
	} else {
		p.maxFileSize = 10485760 // 10MB
	}
	
	// Allowed Extensions
	if exts, ok := config["allowed_extensions"].([]interface{}); ok {
		p.allowedExtensions = make([]string, 0, len(exts))
		for _, ext := range exts {
			if extStr, ok := ext.(string); ok {
				p.allowedExtensions = append(p.allowedExtensions, extStr)
			}
		}
	} else {
		p.allowedExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt"}
	}
	
	// Public URL
	if val, ok := config["public_url"].(string); ok {
		p.publicURL = val
	} else {
		p.publicURL = "/storage"
	}
	
	return nil
}

// ============================================================================
// Backend Initialization
// ============================================================================

func (p *StoragePlugin) initializeBackend() error {
	switch p.storageType {
	case "local":
		p.backend = &LocalStorageBackend{
			basePath: p.localBasePath,
		}
		// Create base directory if it doesn't exist
		if err := os.MkdirAll(p.localBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
		fmt.Printf("[%s] Initialized local storage backend at %s\n", p.Name(), p.localBasePath)
	case "s3":
		// In a real implementation, initialize S3 client here
		p.backend = &S3StorageBackend{
			bucket:    p.s3Bucket,
			region:    p.s3Region,
			endpoint:  p.s3Endpoint,
			accessKey: p.s3AccessKey,
			secretKey: p.s3SecretKey,
		}
		fmt.Printf("[%s] Initialized S3 storage backend (bucket: %s)\n", p.Name(), p.s3Bucket)
	default:
		return fmt.Errorf("unsupported storage type: %s", p.storageType)
	}
	
	return nil
}

// ============================================================================
// Route Registration
// ============================================================================

func (p *StoragePlugin) registerRoutes() error {
	// Register upload endpoint
	// In a real implementation, register routes with the router
	fmt.Printf("[%s] Registered storage routes\n", p.Name())
	return nil
}

// ============================================================================
// Service Export
// ============================================================================

func (p *StoragePlugin) exportService() error {
	service := &StorageService{plugin: p}
	
	err := p.ctx.ExportService("StorageService", service)
	if err != nil {
		return fmt.Errorf("failed to export StorageService: %w", err)
	}
	
	fmt.Printf("[%s] Exported StorageService\n", p.Name())
	return nil
}

// ============================================================================
// Storage Operations
// ============================================================================

func (p *StoragePlugin) store(path string, data io.Reader) error {
	// Validate file extension
	if !p.isAllowedExtension(path) {
		return fmt.Errorf("file extension not allowed")
	}
	
	// Store using backend
	if err := p.backend.Store(path, data); err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}
	
	fmt.Printf("[%s] Stored file: %s\n", p.Name(), path)
	return nil
}

func (p *StoragePlugin) retrieve(path string) (io.ReadCloser, error) {
	return p.backend.Retrieve(path)
}

func (p *StoragePlugin) delete(path string) error {
	if err := p.backend.Delete(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	fmt.Printf("[%s] Deleted file: %s\n", p.Name(), path)
	return nil
}

func (p *StoragePlugin) exists(path string) (bool, error) {
	return p.backend.Exists(path)
}

func (p *StoragePlugin) list(prefix string) ([]string, error) {
	return p.backend.List(prefix)
}

func (p *StoragePlugin) getURL(path string) (string, error) {
	return p.backend.GetURL(path)
}

func (p *StoragePlugin) isAllowedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range p.allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// ============================================================================
// Local Storage Backend
// ============================================================================

type LocalStorageBackend struct {
	basePath string
}

func (b *LocalStorageBackend) Store(path string, data io.Reader) error {
	fullPath := filepath.Join(b.basePath, path)
	
	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Copy data
	_, err = io.Copy(file, data)
	return err
}

func (b *LocalStorageBackend) Retrieve(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(b.basePath, path)
	return os.Open(fullPath)
}

func (b *LocalStorageBackend) Delete(path string) error {
	fullPath := filepath.Join(b.basePath, path)
	return os.Remove(fullPath)
}

func (b *LocalStorageBackend) Exists(path string) (bool, error) {
	fullPath := filepath.Join(b.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (b *LocalStorageBackend) List(prefix string) ([]string, error) {
	fullPath := filepath.Join(b.basePath, prefix)
	var files []string
	
	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(b.basePath, path)
			files = append(files, relPath)
		}
		return nil
	})
	
	return files, err
}

func (b *LocalStorageBackend) GetURL(path string) (string, error) {
	return "/storage/" + path, nil
}

// ============================================================================
// S3 Storage Backend (Stub)
// ============================================================================

type S3StorageBackend struct {
	bucket    string
	region    string
	endpoint  string
	accessKey string
	secretKey string
}

func (b *S3StorageBackend) Store(path string, data io.Reader) error {
	// In a real implementation, use AWS SDK to upload to S3
	fmt.Printf("[S3] Would store file: %s to bucket: %s\n", path, b.bucket)
	return nil
}

func (b *S3StorageBackend) Retrieve(path string) (io.ReadCloser, error) {
	// In a real implementation, use AWS SDK to download from S3
	return nil, fmt.Errorf("S3 retrieve not implemented")
}

func (b *S3StorageBackend) Delete(path string) error {
	// In a real implementation, use AWS SDK to delete from S3
	fmt.Printf("[S3] Would delete file: %s from bucket: %s\n", path, b.bucket)
	return nil
}

func (b *S3StorageBackend) Exists(path string) (bool, error) {
	// In a real implementation, use AWS SDK to check if object exists
	return false, nil
}

func (b *S3StorageBackend) List(prefix string) ([]string, error) {
	// In a real implementation, use AWS SDK to list objects
	return []string{}, nil
}

func (b *S3StorageBackend) GetURL(path string) (string, error) {
	// In a real implementation, generate S3 URL or signed URL
	if b.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", b.endpoint, b.bucket, path), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", b.bucket, b.region, path), nil
}
