package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 10: Path resolution**
// **Validates: Requirements 3.2**
// For any plugin path (absolute or relative), the system should correctly resolve
// the path and locate the plugin
func TestProperty_PathResolution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Create a temporary base directory for testing
	tempDir := t.TempDir()
	logger := &testLogger{}
	loader := NewPluginLoader(tempDir, logger)

	properties.Property("absolute paths within base directory are resolved correctly",
		prop.ForAll(
			func(pathComponents []string) bool {
				// Skip empty path components
				if len(pathComponents) == 0 {
					return true
				}

				// Filter out empty components and clean them
				validComponents := make([]string, 0, len(pathComponents))
				for _, comp := range pathComponents {
					cleaned := strings.TrimSpace(comp)
					// Skip empty, ".", "..", and components with path separators
					if cleaned != "" && cleaned != "." && cleaned != ".." &&
						!strings.Contains(cleaned, string(filepath.Separator)) &&
						!strings.Contains(cleaned, "/") &&
						!strings.Contains(cleaned, "\\") {
						validComponents = append(validComponents, cleaned)
					}
				}

				if len(validComponents) == 0 {
					return true
				}

				// Build an absolute path within the base directory
				relPath := filepath.Join(validComponents...)
				absPath := filepath.Join(tempDir, relPath)

				// Resolve the path
				resolved, err := loader.ResolvePath(absPath)
				if err != nil {
					return false
				}

				// The resolved path should be absolute
				if !filepath.IsAbs(resolved) {
					return false
				}

				// The resolved path should be clean (no redundant separators, etc.)
				if resolved != filepath.Clean(absPath) {
					return false
				}

				return true
			},
			gen.SliceOf(genPathComponent()),
		),
	)

	properties.Property("relative paths are resolved relative to base directory",
		prop.ForAll(
			func(pathComponents []string) bool {
				// Skip empty path components
				if len(pathComponents) == 0 {
					return true
				}

				// Filter out empty components and clean them
				validComponents := make([]string, 0, len(pathComponents))
				for _, comp := range pathComponents {
					cleaned := strings.TrimSpace(comp)
					// Skip empty, ".", "..", and components with path separators
					if cleaned != "" && cleaned != "." && cleaned != ".." &&
						!strings.Contains(cleaned, string(filepath.Separator)) &&
						!strings.Contains(cleaned, "/") &&
						!strings.Contains(cleaned, "\\") {
						validComponents = append(validComponents, cleaned)
					}
				}

				if len(validComponents) == 0 {
					return true
				}

				// Build a relative path
				relPath := filepath.Join(validComponents...)

				// Resolve the path
				resolved, err := loader.ResolvePath(relPath)
				if err != nil {
					return false
				}

				// The resolved path should be absolute
				if !filepath.IsAbs(resolved) {
					return false
				}

				// The resolved path should start with the base directory
				expectedPath := filepath.Join(tempDir, relPath)
				if resolved != filepath.Clean(expectedPath) {
					return false
				}

				return true
			},
			gen.SliceOf(genPathComponent()),
		),
	)

	properties.Property("empty path returns error",
		prop.ForAll(
			func() bool {
				_, err := loader.ResolvePath("")
				return err != nil
			},
		),
	)

	properties.Property("path resolution is idempotent",
		prop.ForAll(
			func(pathComponents []string) bool {
				// Skip empty path components
				if len(pathComponents) == 0 {
					return true
				}

				// Filter out empty components and clean them
				validComponents := make([]string, 0, len(pathComponents))
				for _, comp := range pathComponents {
					cleaned := strings.TrimSpace(comp)
					if cleaned != "" && cleaned != "." && cleaned != ".." &&
						!strings.Contains(cleaned, string(filepath.Separator)) &&
						!strings.Contains(cleaned, "/") &&
						!strings.Contains(cleaned, "\\") {
						validComponents = append(validComponents, cleaned)
					}
				}

				if len(validComponents) == 0 {
					return true
				}

				// Build a relative path
				relPath := filepath.Join(validComponents...)

				// Resolve the path twice
				resolved1, err1 := loader.ResolvePath(relPath)
				if err1 != nil {
					return false
				}

				resolved2, err2 := loader.ResolvePath(relPath)
				if err2 != nil {
					return false
				}

				// Both resolutions should produce the same result
				return resolved1 == resolved2
			},
			gen.SliceOf(genPathComponent()),
		),
	)

	properties.TestingRun(t)
}

// genPathComponent generates valid path components
func genPathComponent() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		// Must be non-empty and not contain path separators
		return s != "" &&
			!strings.Contains(s, "/") &&
			!strings.Contains(s, "\\") &&
			!strings.Contains(s, string(filepath.Separator)) &&
			s != "." &&
			s != ".."
	})
}

// Test basic loader functionality
func TestPluginLoader_Load(t *testing.T) {
	tempDir := t.TempDir()
	logger := &testLogger{}
	loader := NewPluginLoader(tempDir, logger)

	// Create a dummy plugin binary
	pluginPath := filepath.Join(tempDir, "test-plugin")
	if runtime.GOOS == "windows" {
		pluginPath += ".exe"
	}

	// Create the file
	f, err := os.Create(pluginPath)
	if err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}
	f.Close()

	// Make it executable on Unix-like systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(pluginPath, 0755); err != nil {
			t.Fatalf("Failed to make plugin executable: %v", err)
		}
	}

	// Test loading with relative path
	config := PluginConfig{
		Enabled: true,
		Path:    filepath.Base(pluginPath),
	}

	plugin, err := loader.Load(filepath.Base(pluginPath), config)
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin to be non-nil")
	}

	// Test unloading
	if err := loader.Unload(plugin); err != nil {
		t.Fatalf("Failed to unload plugin: %v", err)
	}
}

func TestPluginLoader_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	logger := &testLogger{}
	loader := NewPluginLoader(tempDir, logger)

	config := PluginConfig{
		Enabled: true,
		Path:    "nonexistent-plugin",
	}

	_, err := loader.Load("nonexistent-plugin", config)
	if err == nil {
		t.Fatal("Expected error when loading non-existent plugin")
	}
}

func TestPluginLoader_ResolvePath(t *testing.T) {
	tempDir := t.TempDir()
	logger := &testLogger{}
	loader := NewPluginLoader(tempDir, logger)

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "relative path",
			path:        "plugins/test-plugin",
			expectError: false,
		},
		{
			name:        "absolute path",
			path:        filepath.Join(tempDir, "test-plugin"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := loader.ResolvePath(tt.path)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !filepath.IsAbs(resolved) {
					t.Errorf("Expected absolute path, got: %s", resolved)
				}
			}
		})
	}
}

// testLogger is a simple logger implementation for testing
type testLogger struct {
	messages []string
}

func (l *testLogger) Info(msg string, fields ...interface{}) {
	l.messages = append(l.messages, "INFO: "+msg)
}

func (l *testLogger) Warn(msg string, fields ...interface{}) {
	l.messages = append(l.messages, "WARN: "+msg)
}

func (l *testLogger) Error(msg string, fields ...interface{}) {
	l.messages = append(l.messages, "ERROR: "+msg)
}

func (l *testLogger) Debug(msg string, fields ...interface{}) {
	l.messages = append(l.messages, "DEBUG: "+msg)
}

func (l *testLogger) WithRequestID(requestID string) Logger {
	return l
}
