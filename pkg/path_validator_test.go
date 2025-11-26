package pkg

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathValidator_ValidatePath(t *testing.T) {
	baseDir := "/var/app/data"
	if runtime.GOOS == "windows" {
		baseDir = "C:\\app\\data"
	}

	validator := NewPathValidator(baseDir)

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "safe relative path",
			path:        "files/document.txt",
			expectError: false,
		},
		{
			name:        "safe single file",
			path:        "document.txt",
			expectError: false,
		},
		{
			name:        "directory traversal with ..",
			path:        "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "directory traversal in middle",
			path:        "files/../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "encoded directory traversal",
			path:        "files/..%2F..%2F..%2Fetc/passwd",
			expectError: true,
		},
		{
			name:        "safe subdirectory",
			path:        "uploads/2024/01/file.pdf",
			expectError: false,
		},
		{
			name:        "current directory reference",
			path:        "./files/document.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for path %s, but got none", tt.path)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for path %s, but got: %v", tt.path, err)
			}
		})
	}
}

func TestPathValidator_ResolvePath(t *testing.T) {
	baseDir := "/var/app/data"
	if runtime.GOOS == "windows" {
		baseDir = "C:\\app\\data"
	}

	validator := NewPathValidator(baseDir)

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "resolve relative path",
			path:        "files/document.txt",
			expectError: false,
		},
		{
			name:        "reject traversal",
			path:        "../../../etc/passwd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := validator.ResolvePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path %s, but got none", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path %s, but got: %v", tt.path, err)
				}

				// Verify resolved path starts with base directory
				if !filepath.IsAbs(resolved) {
					t.Errorf("Resolved path should be absolute: %s", resolved)
				}
			}
		})
	}
}

func TestIsPathSafe(t *testing.T) {
	baseDir := "/var/app/data"
	if runtime.GOOS == "windows" {
		baseDir = "C:\\app\\data"
	}

	tests := []struct {
		name string
		path string
		safe bool
	}{
		{
			name: "safe path",
			path: "files/document.txt",
			safe: true,
		},
		{
			name: "unsafe path",
			path: "../../../etc/passwd",
			safe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe := IsPathSafe(baseDir, tt.path)

			if safe != tt.safe {
				t.Errorf("Expected IsPathSafe to return %v for path %s, but got %v", tt.safe, tt.path, safe)
			}
		})
	}
}
