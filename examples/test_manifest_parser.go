package main

import (
	"fmt"
	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	parser := pkg.NewManifestParser()

	// Test YAML parsing
	manifest, err := parser.ParseManifest("examples/plugin-manifest-example.yaml")
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		return
	}
	fmt.Printf("✓ Successfully parsed YAML manifest: %s v%s\n", manifest.Name, manifest.Version)
	fmt.Printf("  Author: %s\n", manifest.Author)
	fmt.Printf("  Dependencies: %d\n", len(manifest.Dependencies))
	fmt.Printf("  Permissions: database=%v, cache=%v, router=%v\n",
		manifest.Permissions.Database, manifest.Permissions.Cache, manifest.Permissions.Router)

	// Test JSON parsing
	manifest2, err := parser.ParseManifest("examples/plugin-manifest-example.json")
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}
	fmt.Printf("\n✓ Successfully parsed JSON manifest: %s v%s\n", manifest2.Name, manifest2.Version)
	fmt.Printf("  Author: %s\n", manifest2.Author)
	fmt.Printf("  Dependencies: %d\n", len(manifest2.Dependencies))
	fmt.Printf("  Permissions: database=%v, cache=%v, router=%v\n",
		manifest2.Permissions.Database, manifest2.Permissions.Cache, manifest2.Permissions.Router)

	fmt.Println("\n✓ All manifest parsing tests passed!")
}
