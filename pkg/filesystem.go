package pkg

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// ErrFileNotFound is returned when a file is not found in the virtual filesystem
	ErrFileNotFound = errors.New("file not found")
	// ErrInvalidPath is returned when a path is invalid
	ErrInvalidPath = errors.New("invalid path")
	// ErrIsDirectory is returned when trying to read a directory as a file
	ErrIsDirectory = errors.New("is a directory")
)

// VirtualFS interface is defined in managers.go
// This file provides implementations of VirtualFS

// osFileSystem implements VirtualFS using the OS filesystem
type osFileSystem struct {
	root string
	mu   sync.RWMutex
}

// NewOSFileSystem creates a new OS-based virtual filesystem
func NewOSFileSystem(root string) VirtualFS {
	return &osFileSystem{
		root: filepath.Clean(root),
	}
}

// Open opens a file from the OS filesystem
func (fs *osFileSystem) Open(name string) (http.File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Clean and validate the path
	name = filepath.Clean("/" + name)
	fullPath := filepath.Join(fs.root, name)

	// Ensure the path is within the root directory (prevent directory traversal)
	if !strings.HasPrefix(fullPath, fs.root) {
		return nil, ErrInvalidPath
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return file, nil
}

// Exists checks if a file exists in the OS filesystem
func (fs *osFileSystem) Exists(name string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	name = filepath.Clean("/" + name)
	fullPath := filepath.Join(fs.root, name)

	if !strings.HasPrefix(fullPath, fs.root) {
		return false
	}

	_, err := os.Stat(fullPath)
	return err == nil
}

// MemoryFileSystem implements VirtualFS using in-memory storage
type MemoryFileSystem struct {
	files map[string]*memoryFile
	mu    sync.RWMutex
}

// memoryFile represents a file in memory
type memoryFile struct {
	name    string
	data    []byte
	modTime time.Time
	isDir   bool
	files   map[string]*memoryFile // For directories
}

// NewMemoryFileSystem creates a new in-memory virtual filesystem
func NewMemoryFileSystem() VirtualFS {
	return &MemoryFileSystem{
		files: make(map[string]*memoryFile),
	}
}

// Open opens a file from the memory filesystem
func (fs *MemoryFileSystem) Open(name string) (http.File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	name = path.Clean("/" + name)

	file, exists := fs.files[name]
	if !exists {
		return nil, ErrFileNotFound
	}

	return &memoryHTTPFile{
		file:   file,
		reader: strings.NewReader(string(file.data)),
	}, nil
}

// Exists checks if a file exists in the memory filesystem
func (fs *MemoryFileSystem) Exists(name string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	name = path.Clean("/" + name)
	_, exists := fs.files[name]
	return exists
}

// AddFile adds a file to the memory filesystem
func (fs *MemoryFileSystem) AddFile(name string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	name = path.Clean("/" + name)

	fs.files[name] = &memoryFile{
		name:    name,
		data:    data,
		modTime: time.Now(),
		isDir:   false,
	}

	return nil
}

// AddDir adds a directory to the memory filesystem
func (fs *MemoryFileSystem) AddDir(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	name = path.Clean("/" + name)

	fs.files[name] = &memoryFile{
		name:    name,
		modTime: time.Now(),
		isDir:   true,
		files:   make(map[string]*memoryFile),
	}

	return nil
}

// memoryHTTPFile implements http.File for in-memory files
type memoryHTTPFile struct {
	file   *memoryFile
	reader *strings.Reader
	offset int64
}

// Read reads from the memory file
func (f *memoryHTTPFile) Read(p []byte) (int, error) {
	if f.file.isDir {
		return 0, ErrIsDirectory
	}
	return f.reader.Read(p)
}

// Seek seeks within the memory file
func (f *memoryHTTPFile) Seek(offset int64, whence int) (int64, error) {
	if f.file.isDir {
		return 0, ErrIsDirectory
	}
	return f.reader.Seek(offset, whence)
}

// Close closes the memory file (no-op for memory files)
func (f *memoryHTTPFile) Close() error {
	return nil
}

// Readdir reads directory entries
func (f *memoryHTTPFile) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.file.isDir {
		return nil, errors.New("not a directory")
	}

	// Convert memory files to FileInfo
	infos := make([]fs.FileInfo, 0, len(f.file.files))
	for _, file := range f.file.files {
		infos = append(infos, &memoryFileInfo{file: file})
	}

	if count <= 0 {
		return infos, nil
	}

	if count > len(infos) {
		count = len(infos)
	}

	return infos[:count], nil
}

// Stat returns file information
func (f *memoryHTTPFile) Stat() (fs.FileInfo, error) {
	return &memoryFileInfo{file: f.file}, nil
}

// memoryFileInfo implements fs.FileInfo for memory files
type memoryFileInfo struct {
	file *memoryFile
}

func (fi *memoryFileInfo) Name() string       { return filepath.Base(fi.file.name) }
func (fi *memoryFileInfo) Size() int64        { return int64(len(fi.file.data)) }
func (fi *memoryFileInfo) Mode() fs.FileMode  { return 0644 }
func (fi *memoryFileInfo) ModTime() time.Time { return fi.file.modTime }
func (fi *memoryFileInfo) IsDir() bool        { return fi.file.isDir }
func (fi *memoryFileInfo) Sys() interface{}   { return nil }

// hostFileSystemManager manages per-host virtual filesystems
type hostFileSystemManager struct {
	hostFS map[string]VirtualFS
	mu     sync.RWMutex
}

// NewHostFileSystemManager creates a new host filesystem manager
func NewHostFileSystemManager() *hostFileSystemManager {
	return &hostFileSystemManager{
		hostFS: make(map[string]VirtualFS),
	}
}

// RegisterHost registers a virtual filesystem for a specific host
func (m *hostFileSystemManager) RegisterHost(hostname string, fs VirtualFS) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hostFS[hostname] = fs
}

// UnregisterHost removes a host's virtual filesystem
func (m *hostFileSystemManager) UnregisterHost(hostname string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.hostFS, hostname)
}

// GetFileSystem returns the virtual filesystem for a specific host
func (m *hostFileSystemManager) GetFileSystem(hostname string) (VirtualFS, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fs, exists := m.hostFS[hostname]
	return fs, exists
}

// fileManager implements FileManager interface
type fileManager struct {
	vfs    VirtualFS
	hostFS *hostFileSystemManager
	mu     sync.RWMutex
}

// NewFileManager creates a new file manager with a virtual filesystem
func NewFileManager(vfs VirtualFS) FileManager {
	return &fileManager{
		vfs:    vfs,
		hostFS: NewHostFileSystemManager(),
	}
}

// Read reads data from a file
func (fm *fileManager) Read(path string) ([]byte, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	file, err := fm.vfs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// Write writes data to a file
func (fm *fileManager) Write(path string, data []byte) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// For memory filesystem, we can write directly
	if memFS, ok := fm.vfs.(*MemoryFileSystem); ok {
		return memFS.AddFile(path, data)
	}

	// For OS filesystem, write to disk
	if osFS, ok := fm.vfs.(*osFileSystem); ok {
		fullPath := filepath.Join(osFS.root, filepath.Clean("/"+path))
		if !strings.HasPrefix(fullPath, osFS.root) {
			return ErrInvalidPath
		}
		return os.WriteFile(fullPath, data, 0644)
	}

	return errors.New("write not supported for this filesystem type")
}

// Append appends data to a file
func (fm *fileManager) Append(path string, data []byte) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// For memory filesystem
	if memFS, ok := fm.vfs.(*MemoryFileSystem); ok {
		// Read existing data
		existing, _ := fm.Read(path)
		newData := append(existing, data...)
		return memFS.AddFile(path, newData)
	}

	// For OS filesystem
	if osFS, ok := fm.vfs.(*osFileSystem); ok {
		fullPath := filepath.Join(osFS.root, filepath.Clean("/"+path))
		if !strings.HasPrefix(fullPath, osFS.root) {
			return ErrInvalidPath
		}
		file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.Write(data)
		return err
	}

	return errors.New("append not supported for this filesystem type")
}

// Delete deletes a file
func (fm *fileManager) Delete(filePath string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// For memory filesystem
	if memFS, ok := fm.vfs.(*MemoryFileSystem); ok {
		memFS.mu.Lock()
		defer memFS.mu.Unlock()
		cleanPath := path.Clean("/" + filePath)
		delete(memFS.files, cleanPath)
		return nil
	}

	// For OS filesystem
	if osFS, ok := fm.vfs.(*osFileSystem); ok {
		fullPath := filepath.Join(osFS.root, filepath.Clean("/"+filePath))
		if !strings.HasPrefix(fullPath, osFS.root) {
			return ErrInvalidPath
		}
		return os.Remove(fullPath)
	}

	return errors.New("delete not supported for this filesystem type")
}

// Exists checks if a file exists
func (fm *fileManager) Exists(path string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.vfs.Exists(path)
}

// CreateDir creates a directory
func (fm *fileManager) CreateDir(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// For memory filesystem
	if memFS, ok := fm.vfs.(*MemoryFileSystem); ok {
		return memFS.AddDir(path)
	}

	// For OS filesystem
	if osFS, ok := fm.vfs.(*osFileSystem); ok {
		fullPath := filepath.Join(osFS.root, filepath.Clean("/"+path))
		if !strings.HasPrefix(fullPath, osFS.root) {
			return ErrInvalidPath
		}
		return os.MkdirAll(fullPath, 0755)
	}

	return errors.New("mkdir not supported for this filesystem type")
}

// RemoveDir removes a directory
func (fm *fileManager) RemoveDir(path string) error {
	return fm.Delete(path)
}

// ListDir lists directory entries
func (fm *fileManager) ListDir(path string) ([]string, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	file, err := fm.vfs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	infos, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names, nil
}

// Size returns the size of a file
func (fm *fileManager) Size(path string) (int64, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	file, err := fm.vfs.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

// ModTime returns the modification time of a file
func (fm *fileManager) ModTime(path string) (time.Time, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	file, err := fm.vfs.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return stat.ModTime(), nil
}

// IsDir checks if a path is a directory
func (fm *fileManager) IsDir(path string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	file, err := fm.vfs.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return false
	}

	return stat.IsDir()
}

// OpenReader opens a file for reading
func (fm *fileManager) OpenReader(path string) (io.ReadCloser, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.vfs.Open(path)
}

// OpenWriter opens a file for writing
func (fm *fileManager) OpenWriter(path string) (io.WriteCloser, error) {
	// For OS filesystem
	if osFS, ok := fm.vfs.(*osFileSystem); ok {
		fullPath := filepath.Join(osFS.root, filepath.Clean("/"+path))
		if !strings.HasPrefix(fullPath, osFS.root) {
			return nil, ErrInvalidPath
		}
		return os.Create(fullPath)
	}

	return nil, errors.New("OpenWriter not supported for this filesystem type")
}

// GetVirtualFS returns the virtual filesystem for a host
func (fm *fileManager) GetVirtualFS(host string) VirtualFS {
	fs, _ := fm.hostFS.GetFileSystem(host)
	return fs
}

// SetVirtualFS sets the virtual filesystem for a host
func (fm *fileManager) SetVirtualFS(host string, fs VirtualFS) error {
	fm.hostFS.RegisterHost(host, fs)
	return nil
}

// SaveUploadedFile saves an uploaded file
func (fm *fileManager) SaveUploadedFile(ctx Context, fieldName, destPath string) error {
	file, err := ctx.FormFile(fieldName)
	if err != nil {
		return err
	}

	return fm.Write(destPath, file.Content)
}

// GetUploadedFile gets an uploaded file
func (fm *fileManager) GetUploadedFile(ctx Context, fieldName string) (*FormFile, error) {
	return ctx.FormFile(fieldName)
}

// GetUploadedFiles gets multiple uploaded files
func (fm *fileManager) GetUploadedFiles(ctx Context, fieldName string) ([]*FormFile, error) {
	// This would need to be implemented in the context
	file, err := ctx.FormFile(fieldName)
	if err != nil {
		return nil, err
	}
	return []*FormFile{file}, nil
}

// CreateTempFile creates a temporary file
func (fm *fileManager) CreateTempFile(prefix string) (string, error) {
	file, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return file.Name(), nil
}

// CreateTempDir creates a temporary directory
func (fm *fileManager) CreateTempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// CleanupTempFiles cleans up temporary files
func (fm *fileManager) CleanupTempFiles() error {
	// This would need to track temp files
	return nil
}
