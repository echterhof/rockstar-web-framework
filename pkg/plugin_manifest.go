package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PluginManifest represents the parsed plugin manifest
type PluginManifest struct {
	Name         string                 `json:"name" yaml:"name"`
	Version      string                 `json:"version" yaml:"version"`
	Description  string                 `json:"description" yaml:"description"`
	Author       string                 `json:"author" yaml:"author"`
	Framework    FrameworkRequirement   `json:"framework" yaml:"framework"`
	Dependencies []ManifestDependency   `json:"dependencies" yaml:"dependencies"`
	Permissions  ManifestPermissions    `json:"permissions" yaml:"permissions"`
	Config       map[string]ConfigField `json:"config" yaml:"config"`
	Hooks        []ManifestHook         `json:"hooks" yaml:"hooks"`
	Events       ManifestEvents         `json:"events" yaml:"events"`
	Exports      []ManifestExport       `json:"exports" yaml:"exports"`
}

// FrameworkRequirement specifies framework version requirements
type FrameworkRequirement struct {
	Version string `json:"version" yaml:"version"`
}

// ManifestDependency represents a plugin dependency in the manifest
type ManifestDependency struct {
	Name     string `json:"name" yaml:"name"`
	Version  string `json:"version" yaml:"version"`
	Optional bool   `json:"optional" yaml:"optional"`
}

// ManifestPermissions represents permissions in the manifest
type ManifestPermissions struct {
	Database   bool `json:"database" yaml:"database"`
	Cache      bool `json:"cache" yaml:"cache"`
	Router     bool `json:"router" yaml:"router"`
	Config     bool `json:"config" yaml:"config"`
	Filesystem bool `json:"filesystem" yaml:"filesystem"`
	Network    bool `json:"network" yaml:"network"`
	Exec       bool `json:"exec" yaml:"exec"`
}

// ConfigField represents a configuration field schema
type ConfigField struct {
	Type        string      `json:"type" yaml:"type"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default" yaml:"default"`
	Description string      `json:"description" yaml:"description"`
	Items       string      `json:"items,omitempty" yaml:"items,omitempty"` // For array types
}

// ManifestHook represents a hook declaration in the manifest
type ManifestHook struct {
	Type     string `json:"type" yaml:"type"`
	Priority int    `json:"priority" yaml:"priority"`
}

// ManifestEvents represents event subscriptions and publications
type ManifestEvents struct {
	Publishes  []string `json:"publishes" yaml:"publishes"`
	Subscribes []string `json:"subscribes" yaml:"subscribes"`
}

// ManifestExport represents an exported service
type ManifestExport struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// ManifestParser handles parsing of plugin manifests
type ManifestParser struct{}

// NewManifestParser creates a new manifest parser
func NewManifestParser() *ManifestParser {
	return &ManifestParser{}
}

// ParseManifest parses a plugin manifest from a file path
func (p *ManifestParser) ParseManifest(path string) (*PluginManifest, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(path))

	var manifest PluginManifest

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse YAML manifest: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse JSON manifest: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported manifest format: %s (supported: .yaml, .yml, .json)", ext)
	}

	// Validate the manifest
	if err := p.ValidateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// ParseManifestFromBytes parses a plugin manifest from raw bytes
func (p *ManifestParser) ParseManifestFromBytes(data []byte, format string) (*PluginManifest, error) {
	var manifest PluginManifest

	switch strings.ToLower(format) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse YAML manifest: %w", err)
		}
	case "json":
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse JSON manifest: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported manifest format: %s (supported: yaml, yml, json)", format)
	}

	// Validate the manifest
	if err := p.ValidateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// ValidateManifest validates a parsed manifest
func (p *ManifestParser) ValidateManifest(manifest *PluginManifest) error {
	var errors []string

	// Required fields
	if manifest.Name == "" {
		errors = append(errors, "name is required")
	}
	if manifest.Version == "" {
		errors = append(errors, "version is required")
	}
	if manifest.Description == "" {
		errors = append(errors, "description is required")
	}
	if manifest.Author == "" {
		errors = append(errors, "author is required")
	}

	// Validate name format (alphanumeric, hyphens, underscores)
	if manifest.Name != "" && !isValidPluginName(manifest.Name) {
		errors = append(errors, "name must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Validate version format (basic semantic version check)
	if manifest.Version != "" && !isValidVersion(manifest.Version) {
		errors = append(errors, "version must be in semantic version format (e.g., 1.0.0)")
	}

	// Validate dependencies
	for i, dep := range manifest.Dependencies {
		if dep.Name == "" {
			errors = append(errors, fmt.Sprintf("dependency %d: name is required", i))
		}
		if dep.Version == "" {
			errors = append(errors, fmt.Sprintf("dependency %d: version is required", i))
		}
	}

	// Validate hooks
	for i, hook := range manifest.Hooks {
		if hook.Type == "" {
			errors = append(errors, fmt.Sprintf("hook %d: type is required", i))
		} else if !isValidHookType(HookType(hook.Type)) {
			errors = append(errors, fmt.Sprintf("hook %d: invalid hook type '%s'", i, hook.Type))
		}
	}

	// Validate config schema
	for key, field := range manifest.Config {
		if field.Type == "" {
			errors = append(errors, fmt.Sprintf("config field '%s': type is required", key))
		} else if !isValidConfigType(field.Type) {
			errors = append(errors, fmt.Sprintf("config field '%s': invalid type '%s'", key, field.Type))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("manifest validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ToPluginDependencies converts manifest dependencies to PluginDependency slice
func (p *ManifestParser) ToPluginDependencies(manifest *PluginManifest) []PluginDependency {
	deps := make([]PluginDependency, len(manifest.Dependencies))
	for i, dep := range manifest.Dependencies {
		deps[i] = PluginDependency{
			Name:             dep.Name,
			Version:          dep.Version,
			Optional:         dep.Optional,
			FrameworkVersion: manifest.Framework.Version,
		}
	}
	return deps
}

// ToPluginPermissions converts manifest permissions to PluginPermissions
func (p *ManifestParser) ToPluginPermissions(manifest *PluginManifest) PluginPermissions {
	return PluginPermissions{
		AllowDatabase:   manifest.Permissions.Database,
		AllowCache:      manifest.Permissions.Cache,
		AllowConfig:     manifest.Permissions.Config,
		AllowRouter:     manifest.Permissions.Router,
		AllowFileSystem: manifest.Permissions.Filesystem,
		AllowNetwork:    manifest.Permissions.Network,
		AllowExec:       manifest.Permissions.Exec,
	}
}

// Helper functions

func isValidPluginName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}

func isValidVersion(version string) bool {
	// Basic semantic version check: should have at least one dot
	// More sophisticated validation can be added later
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check if part contains only digits (basic check)
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				// Allow version suffixes like "1.0.0-alpha"
				if ch == '-' {
					return true
				}
				return false
			}
		}
	}
	return true
}

func isValidConfigType(configType string) bool {
	validTypes := []string{
		"string", "int", "float", "bool", "duration", "array", "object",
	}
	for _, valid := range validTypes {
		if configType == valid {
			return true
		}
	}
	return false
}
