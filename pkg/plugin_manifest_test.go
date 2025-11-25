package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gopkg.in/yaml.v3"
)

// **Feature: plugin-system, Property 49: Manifest parsing**
// **Validates: Requirements 12.1**
// For any plugin directory containing a valid manifest file in YAML or JSON format,
// the system should successfully parse the manifest
func TestProperty_ManifestParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("valid YAML manifest parses successfully",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal to YAML
				data, err := yaml.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse from bytes
				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "yaml")
				if err != nil {
					return false
				}

				// Verify fields match
				if parsed.Name != name || parsed.Version != version ||
					parsed.Description != description || parsed.Author != author {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("valid JSON manifest parses successfully",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal to JSON
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse from bytes
				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify fields match
				if parsed.Name != name || parsed.Version != version ||
					parsed.Description != description || parsed.Author != author {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest file with .yaml extension parses as YAML",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal to YAML
				data, err := yaml.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Create temp file
				tmpDir := t.TempDir()
				manifestPath := filepath.Join(tmpDir, "plugin.yaml")
				err = os.WriteFile(manifestPath, data, 0644)
				if err != nil {
					return false
				}

				// Parse from file
				parser := NewManifestParser()
				parsed, err := parser.ParseManifest(manifestPath)
				if err != nil {
					return false
				}

				// Verify fields match
				if parsed.Name != name || parsed.Version != version ||
					parsed.Description != description || parsed.Author != author {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest file with .json extension parses as JSON",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal to JSON
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Create temp file
				tmpDir := t.TempDir()
				manifestPath := filepath.Join(tmpDir, "plugin.json")
				err = os.WriteFile(manifestPath, data, 0644)
				if err != nil {
					return false
				}

				// Parse from file
				parser := NewManifestParser()
				parsed, err := parser.ParseManifest(manifestPath)
				if err != nil {
					return false
				}

				// Verify fields match
				if parsed.Name != name || parsed.Version != version ||
					parsed.Description != description || parsed.Author != author {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 50: Manifest field extraction**
// **Validates: Requirements 12.2**
// For any valid manifest, the system should extract the name, version, description, and author fields
func TestProperty_ManifestFieldExtraction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all required fields are extracted correctly",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal and parse
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify all fields are extracted
				if parsed.Name != name {
					return false
				}
				if parsed.Version != version {
					return false
				}
				if parsed.Description != description {
					return false
				}
				if parsed.Author != author {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("framework version is extracted when present",
		prop.ForAll(
			func(name, version, description, author, frameworkVersion string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest with framework version
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Framework: FrameworkRequirement{
						Version: frameworkVersion,
					},
				}

				// Marshal and parse
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify framework version is extracted
				if parsed.Framework.Version != frameworkVersion {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 51: Dependency parsing**
// **Validates: Requirements 12.3**
// For any manifest declaring dependencies, the system should parse all dependency names and version constraints
func TestProperty_DependencyParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all dependencies are parsed correctly",
		prop.ForAll(
			func(name, version, description, author string, depCount int) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}
				// Limit dependency count
				if depCount < 0 || depCount > 10 {
					return true
				}

				// Create dependencies
				deps := make([]ManifestDependency, depCount)
				for i := 0; i < depCount; i++ {
					deps[i] = ManifestDependency{
						Name:     fmt.Sprintf("dep-%d", i),
						Version:  "1.0.0",
						Optional: i%2 == 0,
					}
				}

				// Create a valid manifest with dependencies
				manifest := PluginManifest{
					Name:         name,
					Version:      version,
					Description:  description,
					Author:       author,
					Dependencies: deps,
				}

				// Marshal and parse
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify all dependencies are parsed
				if len(parsed.Dependencies) != depCount {
					return false
				}

				for i, dep := range parsed.Dependencies {
					if dep.Name != fmt.Sprintf("dep-%d", i) {
						return false
					}
					if dep.Version != "1.0.0" {
						return false
					}
					if dep.Optional != (i%2 == 0) {
						return false
					}
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.IntRange(0, 10),
		),
	)

	properties.Property("dependency version constraints are preserved",
		prop.ForAll(
			func(name, version, description, author, depName, depVersion string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}
				if depName == "" || depVersion == "" {
					return true
				}

				// Create a valid manifest with one dependency
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Dependencies: []ManifestDependency{
						{
							Name:     depName,
							Version:  depVersion,
							Optional: false,
						},
					},
				}

				// Marshal and parse
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify dependency version constraint is preserved
				if len(parsed.Dependencies) != 1 {
					return false
				}
				if parsed.Dependencies[0].Name != depName {
					return false
				}
				if parsed.Dependencies[0].Version != depVersion {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("ToPluginDependencies converts manifest dependencies correctly",
		prop.ForAll(
			func(name, version, description, author, frameworkVersion string, depCount int) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}
				// Limit dependency count
				if depCount < 0 || depCount > 10 {
					return true
				}

				// Create dependencies
				deps := make([]ManifestDependency, depCount)
				for i := 0; i < depCount; i++ {
					deps[i] = ManifestDependency{
						Name:     fmt.Sprintf("dep-%d", i),
						Version:  "1.0.0",
						Optional: i%2 == 0,
					}
				}

				// Create a valid manifest with dependencies
				manifest := PluginManifest{
					Name:         name,
					Version:      version,
					Description:  description,
					Author:       author,
					Dependencies: deps,
					Framework: FrameworkRequirement{
						Version: frameworkVersion,
					},
				}

				parser := NewManifestParser()
				pluginDeps := parser.ToPluginDependencies(&manifest)

				// Verify conversion
				if len(pluginDeps) != depCount {
					return false
				}

				for i, dep := range pluginDeps {
					if dep.Name != fmt.Sprintf("dep-%d", i) {
						return false
					}
					if dep.Version != "1.0.0" {
						return false
					}
					if dep.Optional != (i%2 == 0) {
						return false
					}
					if dep.FrameworkVersion != frameworkVersion {
						return false
					}
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString(),
			gen.IntRange(0, 10),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 52: Permission parsing**
// **Validates: Requirements 12.4**
// For any manifest specifying permissions, the system should parse the complete permission list
func TestProperty_PermissionParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all permission flags are parsed correctly",
		prop.ForAll(
			func(name, version, description, author string,
				database, cache, router, config, filesystem, network, exec bool) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest with permissions
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Permissions: ManifestPermissions{
						Database:   database,
						Cache:      cache,
						Router:     router,
						Config:     config,
						Filesystem: filesystem,
						Network:    network,
						Exec:       exec,
					},
				}

				// Marshal and parse
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				parser := NewManifestParser()
				parsed, err := parser.ParseManifestFromBytes(data, "json")
				if err != nil {
					return false
				}

				// Verify all permission flags are parsed
				if parsed.Permissions.Database != database {
					return false
				}
				if parsed.Permissions.Cache != cache {
					return false
				}
				if parsed.Permissions.Router != router {
					return false
				}
				if parsed.Permissions.Config != config {
					return false
				}
				if parsed.Permissions.Filesystem != filesystem {
					return false
				}
				if parsed.Permissions.Network != network {
					return false
				}
				if parsed.Permissions.Exec != exec {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
		),
	)

	properties.Property("ToPluginPermissions converts manifest permissions correctly",
		prop.ForAll(
			func(name, version, description, author string,
				database, cache, router, config, filesystem, network, exec bool) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a valid manifest with permissions
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Permissions: ManifestPermissions{
						Database:   database,
						Cache:      cache,
						Router:     router,
						Config:     config,
						Filesystem: filesystem,
						Network:    network,
						Exec:       exec,
					},
				}

				parser := NewManifestParser()
				pluginPerms := parser.ToPluginPermissions(&manifest)

				// Verify conversion
				if pluginPerms.AllowDatabase != database {
					return false
				}
				if pluginPerms.AllowCache != cache {
					return false
				}
				if pluginPerms.AllowRouter != router {
					return false
				}
				if pluginPerms.AllowConfig != config {
					return false
				}
				if pluginPerms.AllowFileSystem != filesystem {
					return false
				}
				if pluginPerms.AllowNetwork != network {
					return false
				}
				if pluginPerms.AllowExec != exec {
					return false
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
			gen.Bool(),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 53: Invalid manifest rejection**
// **Validates: Requirements 12.5**
// For any invalid manifest file, the system should prevent plugin loading and log descriptive validation errors
func TestProperty_InvalidManifestRejection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("manifest with missing name is rejected",
		prop.ForAll(
			func(version, description, author string) bool {
				// Skip empty required fields (except name which we're testing)
				if version == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with missing name
				manifest := PluginManifest{
					Name:        "", // Missing name
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with missing version is rejected",
		prop.ForAll(
			func(name, description, author string) bool {
				// Skip empty required fields (except version which we're testing)
				if name == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with missing version
				manifest := PluginManifest{
					Name:        name,
					Version:     "", // Missing version
					Description: description,
					Author:      author,
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with missing description is rejected",
		prop.ForAll(
			func(name, version, author string) bool {
				// Skip empty required fields (except description which we're testing)
				if name == "" || version == "" || author == "" {
					return true
				}

				// Create a manifest with missing description
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: "", // Missing description
					Author:      author,
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with missing author is rejected",
		prop.ForAll(
			func(name, version, description string) bool {
				// Skip empty required fields (except author which we're testing)
				if name == "" || version == "" || description == "" {
					return true
				}

				// Create a manifest with missing author
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      "", // Missing author
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with invalid plugin name is rejected",
		prop.ForAll(
			func(version, description, author string) bool {
				// Skip empty required fields
				if version == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with invalid name (contains invalid characters)
				manifest := PluginManifest{
					Name:        "invalid@name!", // Invalid characters
					Version:     version,
					Description: description,
					Author:      author,
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with invalid version format is rejected",
		prop.ForAll(
			func(name, description, author string) bool {
				// Skip empty required fields
				if name == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with invalid version (no dots)
				manifest := PluginManifest{
					Name:        name,
					Version:     "invalid", // Invalid version format
					Description: description,
					Author:      author,
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with dependency missing name is rejected",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with dependency missing name
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Dependencies: []ManifestDependency{
						{
							Name:    "", // Missing name
							Version: "1.0.0",
						},
					},
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("manifest with dependency missing version is rejected",
		prop.ForAll(
			func(name, version, description, author string) bool {
				// Skip empty required fields
				if name == "" || version == "" || description == "" || author == "" {
					return true
				}

				// Create a manifest with dependency missing version
				manifest := PluginManifest{
					Name:        name,
					Version:     version,
					Description: description,
					Author:      author,
					Dependencies: []ManifestDependency{
						{
							Name:    "some-dep",
							Version: "", // Missing version
						},
					},
				}

				// Marshal
				data, err := json.Marshal(&manifest)
				if err != nil {
					return false
				}

				// Parse should fail
				parser := NewManifestParser()
				_, err = parser.ParseManifestFromBytes(data, "json")
				if err == nil {
					return false // Should have returned an error
				}

				return true
			},
			genValidPluginName(),
			genValidVersion(),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// Generator helpers

func genValidPluginName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		if s == "" {
			return false
		}
		// Only alphanumeric, hyphens, and underscores
		for _, ch := range s {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
				return false
			}
		}
		return true
	})
}

func genValidVersion() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("1.0.0"),
		gen.Const("2.1.3"),
		gen.Const("0.1.0"),
		gen.Const("10.20.30"),
		gen.Const("1.0.0-alpha"),
		gen.Const("2.0.0-beta.1"),
	)
}
