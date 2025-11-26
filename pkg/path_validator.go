package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathValidator provides secure path validation to prevent directory traversal attacks
type PathValidator struct {
	baseDir string
}

// NewPathValidator creates a new path validator with a base directory
func NewPathValidator(baseDir string) *PathValidator {
	return &PathValidator{
		baseDir: filepath.Clean(baseDir),
	}
}

// ValidatePath validates that a path is safe and within the base directory
// This prevents directory traversal attacks (e.g., "../../../etc/passwd")
func (pv *PathValidator) ValidatePath(requestedPath string) error {
	// Clean the requested path to normalize it
	cleanPath := filepath.Clean(requestedPath)

	// Check for suspicious patterns
	if strings.Contains(cleanPath, "..") {
		return &FrameworkError{
			Code:       ErrCodePathTraversal,
			Message:    fmt.Sprintf("path traversal detected: %s", requestedPath),
			StatusCode: 400,
			I18nKey:    "error.security.path_traversal",
		}
	}

	// If path is absolute, it must start with base directory
	if filepath.IsAbs(cleanPath) {
		if !strings.HasPrefix(cleanPath, pv.baseDir) {
			return &FrameworkError{
				Code:       ErrCodePathTraversal,
				Message:    fmt.Sprintf("path outside base directory: %s", requestedPath),
				StatusCode: 400,
				I18nKey:    "error.security.path_outside_base",
			}
		}
		return nil
	}

	// For relative paths, join with base and validate
	fullPath := filepath.Join(pv.baseDir, cleanPath)
	fullPath = filepath.Clean(fullPath)

	// Ensure the resolved path is still within base directory
	if !strings.HasPrefix(fullPath, pv.baseDir) {
		return &FrameworkError{
			Code:       ErrCodePathTraversal,
			Message:    fmt.Sprintf("path resolves outside base directory: %s", requestedPath),
			StatusCode: 400,
			I18nKey:    "error.security.path_outside_base",
		}
	}

	return nil
}

// ResolvePath validates and resolves a path to its absolute form within the base directory
func (pv *PathValidator) ResolvePath(requestedPath string) (string, error) {
	// Validate first
	if err := pv.ValidatePath(requestedPath); err != nil {
		return "", err
	}

	// Clean the path
	cleanPath := filepath.Clean(requestedPath)

	// If already absolute and valid, return it
	if filepath.IsAbs(cleanPath) {
		return cleanPath, nil
	}

	// Join with base directory and clean
	fullPath := filepath.Join(pv.baseDir, cleanPath)
	return filepath.Clean(fullPath), nil
}

// ValidateAndResolvePath is a convenience method that validates and resolves in one call
func ValidateAndResolvePath(baseDir, requestedPath string) (string, error) {
	validator := NewPathValidator(baseDir)
	return validator.ResolvePath(requestedPath)
}

// IsPathSafe checks if a path is safe without returning an error
func IsPathSafe(baseDir, requestedPath string) bool {
	validator := NewPathValidator(baseDir)
	return validator.ValidatePath(requestedPath) == nil
}
