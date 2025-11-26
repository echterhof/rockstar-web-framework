//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Plugin Manifest Parser Test Utility
// ============================================================================
// This utility demonstrates:
// - Parsing plugin manifests from YAML and JSON files
// - Validating manifest structure and required fields
// - Displaying manifest information
// - Error handling for invalid manifests
// ============================================================================

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Plugin Manifest Parser Test")
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Create manifest parser
	parser := pkg.NewManifestParser()

	// Test files to parse
	testFiles := []string{
		"examples/plugin-manifest-example.yaml",
		"examples/plugin-manifest-example.json",
	}

	successCount := 0
	failCount := 0

	// Parse each test file
	for _, filePath := range testFiles {
		fmt.Printf("Testing: %s\n", filePath)
		fmt.Println("-" + "--------------------------------------------------------")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("  ❌ File not found: %s\n", filePath)
			fmt.Println()
			failCount++
			continue
		}

		// Parse the manifest
		manifest, err := parser.ParseManifest(filePath)
		if err != nil {
			fmt.Printf("  ❌ Error parsing manifest: %v\n", err)
			fmt.Println()
			failCount++
			continue
		}

		// Validate the manifest
		if err := validateManifest(manifest); err != nil {
			fmt.Printf("  ❌ Validation failed: %v\n", err)
			fmt.Println()
			failCount++
			continue
		}

		// Display manifest information
		displayManifest(manifest)
		fmt.Printf("  ✓ Successfully parsed and validated\n")
		fmt.Println()
		successCount++
	}

	// Print summary
	fmt.Println("=" + "==========================================================")
	fmt.Printf("Results: %d passed, %d failed\n", successCount, failCount)

	if failCount > 0 {
		os.Exit(1)
	}
}

// ============================================================================
// Validation Functions
// ============================================================================

// validateManifest validates that a manifest has all required fields
func validateManifest(manifest *pkg.PluginManifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Check required fields
	if manifest.Name == "" {
		return fmt.Errorf("name is required")
	}

	if manifest.Version == "" {
		return fmt.Errorf("version is required")
	}

	if manifest.Description == "" {
		return fmt.Errorf("description is required")
	}

	if manifest.Author == "" {
		return fmt.Errorf("author is required")
	}

	// Validate framework requirement
	if manifest.Framework.Version == "" {
		return fmt.Errorf("framework version is required")
	}

	// Validate dependencies
	for i, dep := range manifest.Dependencies {
		if dep.Name == "" {
			return fmt.Errorf("dependency %d: name is required", i)
		}
		if dep.Version == "" {
			return fmt.Errorf("dependency %d: version is required", i)
		}
	}

	// Validate hooks
	for i, hook := range manifest.Hooks {
		if hook.Type == "" {
			return fmt.Errorf("hook %d: type is required", i)
		}
	}

	// Validate exports
	for i, export := range manifest.Exports {
		if export.Name == "" {
			return fmt.Errorf("export %d: name is required", i)
		}
	}

	return nil
}

// ============================================================================
// Display Functions
// ============================================================================

// displayManifest displays detailed information about a manifest
func displayManifest(manifest *pkg.PluginManifest) {
	// Basic information
	fmt.Printf("  Plugin: %s v%s\n", manifest.Name, manifest.Version)
	fmt.Printf("  Description: %s\n", manifest.Description)
	fmt.Printf("  Author: %s\n", manifest.Author)
	fmt.Printf("  Framework: %s\n", manifest.Framework.Version)

	// Dependencies
	if len(manifest.Dependencies) > 0 {
		fmt.Printf("  Dependencies:\n")
		for _, dep := range manifest.Dependencies {
			optional := ""
			if dep.Optional {
				optional = " (optional)"
			}
			fmt.Printf("    - %s v%s%s\n", dep.Name, dep.Version, optional)
		}
	} else {
		fmt.Printf("  Dependencies: none\n")
	}

	// Permissions
	fmt.Printf("  Permissions:\n")
	fmt.Printf("    - Database: %v\n", manifest.Permissions.Database)
	fmt.Printf("    - Cache: %v\n", manifest.Permissions.Cache)
	fmt.Printf("    - Router: %v\n", manifest.Permissions.Router)
	fmt.Printf("    - Config: %v\n", manifest.Permissions.Config)
	fmt.Printf("    - Filesystem: %v\n", manifest.Permissions.Filesystem)
	fmt.Printf("    - Network: %v\n", manifest.Permissions.Network)
	fmt.Printf("    - Exec: %v\n", manifest.Permissions.Exec)

	// Configuration fields
	if len(manifest.Config) > 0 {
		fmt.Printf("  Configuration Fields:\n")
		for name, field := range manifest.Config {
			required := ""
			if field.Required {
				required = " (required)"
			}
			fmt.Printf("    - %s: %s%s\n", name, field.Type, required)
			if field.Description != "" {
				fmt.Printf("      %s\n", field.Description)
			}
		}
	} else {
		fmt.Printf("  Configuration Fields: none\n")
	}

	// Hooks
	if len(manifest.Hooks) > 0 {
		fmt.Printf("  Hooks:\n")
		for _, hook := range manifest.Hooks {
			fmt.Printf("    - %s (priority: %d)\n", hook.Type, hook.Priority)
		}
	} else {
		fmt.Printf("  Hooks: none\n")
	}

	// Events
	if len(manifest.Events.Publishes) > 0 || len(manifest.Events.Subscribes) > 0 {
		fmt.Printf("  Events:\n")
		if len(manifest.Events.Publishes) > 0 {
			fmt.Printf("    Publishes:\n")
			for _, event := range manifest.Events.Publishes {
				fmt.Printf("      - %s\n", event)
			}
		}
		if len(manifest.Events.Subscribes) > 0 {
			fmt.Printf("    Subscribes:\n")
			for _, event := range manifest.Events.Subscribes {
				fmt.Printf("      - %s\n", event)
			}
		}
	} else {
		fmt.Printf("  Events: none\n")
	}

	// Exports
	if len(manifest.Exports) > 0 {
		fmt.Printf("  Exports:\n")
		for _, export := range manifest.Exports {
			fmt.Printf("    - %s: %s\n", export.Name, export.Description)
		}
	} else {
		fmt.Printf("  Exports: none\n")
	}
}

// ============================================================================
// Error Handling Examples
// ============================================================================

// testInvalidManifests demonstrates error handling for invalid manifests
func testInvalidManifests() {
	fmt.Println("\nTesting error handling with invalid manifests:")
	fmt.Println("-" + "--------------------------------------------------------")

	parser := pkg.NewManifestParser()

	// Test non-existent file
	_, err := parser.ParseManifest("non-existent-file.yaml")
	if err != nil {
		fmt.Printf("  ✓ Correctly handled non-existent file: %v\n", err)
	} else {
		log.Printf("  ❌ Should have failed for non-existent file\n")
	}

	// Test invalid file extension
	_, err = parser.ParseManifest("test.txt")
	if err != nil {
		fmt.Printf("  ✓ Correctly handled invalid file extension: %v\n", err)
	} else {
		log.Printf("  ❌ Should have failed for invalid file extension\n")
	}
}
