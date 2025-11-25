package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

// Unit tests for manifest parser basic functionality

func TestManifestParser_ParseYAML(t *testing.T) {
	yamlContent := `
name: test-plugin
version: 1.0.0
description: A test plugin
author: Test Author
framework:
  version: ">=1.0.0"
dependencies:
  - name: dep1
    version: ">=1.0.0"
    optional: false
permissions:
  database: true
  cache: true
  router: false
config:
  api_key:
    type: string
    required: true
    description: API key
`

	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "plugin.yaml")
	err := os.WriteFile(manifestPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}

	parser := NewManifestParser()
	manifest, err := parser.ParseManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Verify basic fields
	if manifest.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", manifest.Name)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", manifest.Version)
	}
	if manifest.Description != "A test plugin" {
		t.Errorf("Expected description 'A test plugin', got '%s'", manifest.Description)
	}
	if manifest.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", manifest.Author)
	}

	// Verify framework version
	if manifest.Framework.Version != ">=1.0.0" {
		t.Errorf("Expected framework version '>=1.0.0', got '%s'", manifest.Framework.Version)
	}

	// Verify dependencies
	if len(manifest.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(manifest.Dependencies))
	} else {
		dep := manifest.Dependencies[0]
		if dep.Name != "dep1" {
			t.Errorf("Expected dependency name 'dep1', got '%s'", dep.Name)
		}
		if dep.Version != ">=1.0.0" {
			t.Errorf("Expected dependency version '>=1.0.0', got '%s'", dep.Version)
		}
		if dep.Optional {
			t.Error("Expected dependency to be required, got optional")
		}
	}

	// Verify permissions
	if !manifest.Permissions.Database {
		t.Error("Expected database permission to be true")
	}
	if !manifest.Permissions.Cache {
		t.Error("Expected cache permission to be true")
	}
	if manifest.Permissions.Router {
		t.Error("Expected router permission to be false")
	}

	// Verify config schema
	if len(manifest.Config) != 1 {
		t.Errorf("Expected 1 config field, got %d", len(manifest.Config))
	} else {
		apiKey, exists := manifest.Config["api_key"]
		if !exists {
			t.Error("Expected 'api_key' config field to exist")
		} else {
			if apiKey.Type != "string" {
				t.Errorf("Expected api_key type 'string', got '%s'", apiKey.Type)
			}
			if !apiKey.Required {
				t.Error("Expected api_key to be required")
			}
			if apiKey.Description != "API key" {
				t.Errorf("Expected api_key description 'API key', got '%s'", apiKey.Description)
			}
		}
	}
}

func TestManifestParser_ParseJSON(t *testing.T) {
	jsonContent := `{
  "name": "json-plugin",
  "version": "2.0.0",
  "description": "A JSON plugin",
  "author": "JSON Author",
  "permissions": {
    "database": false,
    "cache": true
  }
}`

	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "plugin.json")
	err := os.WriteFile(manifestPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test manifest: %v", err)
	}

	parser := NewManifestParser()
	manifest, err := parser.ParseManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	if manifest.Name != "json-plugin" {
		t.Errorf("Expected name 'json-plugin', got '%s'", manifest.Name)
	}
	if manifest.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", manifest.Version)
	}
	if manifest.Permissions.Database {
		t.Error("Expected database permission to be false")
	}
	if !manifest.Permissions.Cache {
		t.Error("Expected cache permission to be true")
	}
}

func TestManifestParser_InvalidManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "missing name",
			content: `{
				"version": "1.0.0",
				"description": "Test",
				"author": "Author"
			}`,
			wantErr: true,
		},
		{
			name: "missing version",
			content: `{
				"name": "test",
				"description": "Test",
				"author": "Author"
			}`,
			wantErr: true,
		},
		{
			name: "invalid version format",
			content: `{
				"name": "test",
				"version": "invalid",
				"description": "Test",
				"author": "Author"
			}`,
			wantErr: true,
		},
		{
			name: "invalid plugin name",
			content: `{
				"name": "invalid@name!",
				"version": "1.0.0",
				"description": "Test",
				"author": "Author"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewManifestParser()
			_, err := parser.ParseManifestFromBytes([]byte(tt.content), "json")
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifestFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManifestParser_ToPluginDependencies(t *testing.T) {
	manifest := &PluginManifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test",
		Author:      "Author",
		Framework: FrameworkRequirement{
			Version: ">=1.0.0",
		},
		Dependencies: []ManifestDependency{
			{
				Name:     "dep1",
				Version:  ">=1.0.0",
				Optional: false,
			},
			{
				Name:     "dep2",
				Version:  ">=2.0.0",
				Optional: true,
			},
		},
	}

	parser := NewManifestParser()
	deps := parser.ToPluginDependencies(manifest)

	if len(deps) != 2 {
		t.Fatalf("Expected 2 dependencies, got %d", len(deps))
	}

	// Check first dependency
	if deps[0].Name != "dep1" {
		t.Errorf("Expected dep1, got %s", deps[0].Name)
	}
	if deps[0].Version != ">=1.0.0" {
		t.Errorf("Expected >=1.0.0, got %s", deps[0].Version)
	}
	if deps[0].Optional {
		t.Error("Expected dep1 to be required")
	}
	if deps[0].FrameworkVersion != ">=1.0.0" {
		t.Errorf("Expected framework version >=1.0.0, got %s", deps[0].FrameworkVersion)
	}

	// Check second dependency
	if deps[1].Name != "dep2" {
		t.Errorf("Expected dep2, got %s", deps[1].Name)
	}
	if !deps[1].Optional {
		t.Error("Expected dep2 to be optional")
	}
}

func TestManifestParser_ToPluginPermissions(t *testing.T) {
	manifest := &PluginManifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test",
		Author:      "Author",
		Permissions: ManifestPermissions{
			Database:   true,
			Cache:      true,
			Router:     false,
			Config:     true,
			Filesystem: false,
			Network:    false,
			Exec:       false,
		},
	}

	parser := NewManifestParser()
	perms := parser.ToPluginPermissions(manifest)

	if !perms.AllowDatabase {
		t.Error("Expected database permission to be true")
	}
	if !perms.AllowCache {
		t.Error("Expected cache permission to be true")
	}
	if perms.AllowRouter {
		t.Error("Expected router permission to be false")
	}
	if !perms.AllowConfig {
		t.Error("Expected config permission to be true")
	}
	if perms.AllowFileSystem {
		t.Error("Expected filesystem permission to be false")
	}
	if perms.AllowNetwork {
		t.Error("Expected network permission to be false")
	}
	if perms.AllowExec {
		t.Error("Expected exec permission to be false")
	}
}
