package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 16: Framework version validation**
// **Validates: Requirements 5.1**
// For any plugin declaring a framework version requirement, the system should verify
// the current framework version satisfies the semantic version constraint before loading
func TestProperty_FrameworkVersionValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("valid framework version constraints are accepted",
		prop.ForAll(
			func(constraint string) bool {
				resolver := NewDependencyResolver()

				// Test with valid constraints that should pass
				err := resolver.ValidateFrameworkVersion(constraint)

				// Should not error for valid constraints that match current version
				return err == nil
			},
			gen.OneConstOf(
				">=1.0.0",
				">=0.1.0",
				">=1.0.0,<2.0.0",
				"1.0.0",
				">=1.0.0,<=1.0.0",
				"", // Empty constraint means any version
			),
		),
	)

	properties.Property("incompatible framework version constraints are rejected",
		prop.ForAll(
			func(constraint string) bool {
				resolver := NewDependencyResolver()

				// Test with constraints that should fail
				err := resolver.ValidateFrameworkVersion(constraint)

				// Should error for incompatible constraints
				return err != nil
			},
			gen.OneConstOf(
				">=2.0.0",
				"<1.0.0",
				">1.0.0",
				">=1.1.0",
				"2.0.0",
			),
		),
	)

	properties.Property("empty framework version constraint accepts any version",
		prop.ForAll(
			func() bool {
				resolver := NewDependencyResolver()

				// Empty constraint should always pass
				err := resolver.ValidateFrameworkVersion("")

				return err == nil
			},
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 17: Dependency load ordering**
// **Validates: Requirements 5.2**
// For any plugin with dependencies, all dependencies should be loaded and initialized
// before the dependent plugin is initialized
func TestProperty_DependencyLoadOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("dependencies appear before dependents in load order",
		prop.ForAll(
			func(pluginCount int) bool {
				// Limit plugin count
				if pluginCount < 2 || pluginCount > 10 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create a chain of dependencies: plugin-0 <- plugin-1 <- plugin-2 <- ...
				for i := 0; i < pluginCount; i++ {
					name := fmt.Sprintf("plugin-%d", i)
					var deps []PluginDependency

					// Each plugin depends on the previous one
					if i > 0 {
						deps = []PluginDependency{
							{
								Name:    fmt.Sprintf("plugin-%d", i-1),
								Version: "1.0.0",
							},
						}
					}

					resolver.AddPlugin(name, "1.0.0", deps, nil)
				}

				// Resolve dependencies
				loadOrder, err := resolver.ResolveDependencies()
				if err != nil {
					return false
				}

				// Verify load order: dependencies should come before dependents
				pluginIndex := make(map[string]int)
				for i, name := range loadOrder {
					pluginIndex[name] = i
				}

				// Check that each plugin comes after its dependencies
				for i := 1; i < pluginCount; i++ {
					pluginName := fmt.Sprintf("plugin-%d", i)
					depName := fmt.Sprintf("plugin-%d", i-1)

					if pluginIndex[depName] >= pluginIndex[pluginName] {
						return false // Dependency should come before dependent
					}
				}

				return true
			},
			gen.IntRange(2, 10),
		),
	)

	properties.Property("plugins with no dependencies can be loaded in any order",
		prop.ForAll(
			func(pluginCount int) bool {
				// Limit plugin count
				if pluginCount < 1 || pluginCount > 10 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create plugins with no dependencies
				for i := 0; i < pluginCount; i++ {
					name := fmt.Sprintf("plugin-%d", i)
					resolver.AddPlugin(name, "1.0.0", []PluginDependency{}, nil)
				}

				// Resolve dependencies
				loadOrder, err := resolver.ResolveDependencies()
				if err != nil {
					return false
				}

				// Should return all plugins
				if len(loadOrder) != pluginCount {
					return false
				}

				// All plugins should be present
				pluginSet := make(map[string]bool)
				for _, name := range loadOrder {
					pluginSet[name] = true
				}

				for i := 0; i < pluginCount; i++ {
					if !pluginSet[fmt.Sprintf("plugin-%d", i)] {
						return false
					}
				}

				return true
			},
			gen.IntRange(1, 10),
		),
	)

	properties.Property("complex dependency graphs are resolved correctly",
		prop.ForAll(
			func() bool {
				resolver := NewDependencyResolver()

				// Create a diamond dependency:
				//     A
				//    / \
				//   B   C
				//    \ /
				//     D

				resolver.AddPlugin("A", "1.0.0", []PluginDependency{}, nil)
				resolver.AddPlugin("B", "1.0.0", []PluginDependency{
					{Name: "A", Version: "1.0.0"},
				}, nil)
				resolver.AddPlugin("C", "1.0.0", []PluginDependency{
					{Name: "A", Version: "1.0.0"},
				}, nil)
				resolver.AddPlugin("D", "1.0.0", []PluginDependency{
					{Name: "B", Version: "1.0.0"},
					{Name: "C", Version: "1.0.0"},
				}, nil)

				// Resolve dependencies
				loadOrder, err := resolver.ResolveDependencies()
				if err != nil {
					return false
				}

				// Build index
				pluginIndex := make(map[string]int)
				for i, name := range loadOrder {
					pluginIndex[name] = i
				}

				// Verify constraints
				// A must come before B and C
				if pluginIndex["A"] >= pluginIndex["B"] || pluginIndex["A"] >= pluginIndex["C"] {
					return false
				}

				// B and C must come before D
				if pluginIndex["B"] >= pluginIndex["D"] || pluginIndex["C"] >= pluginIndex["D"] {
					return false
				}

				return true
			},
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 18: Missing dependency detection**
// **Validates: Requirements 5.3**
// For any plugin with a missing required dependency, the system should prevent loading
// and log an error identifying the missing dependency
func TestProperty_MissingDependencyDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("missing required dependency is detected",
		prop.ForAll(
			func(pluginName, depName string) bool {
				// Skip empty names
				if pluginName == "" || depName == "" {
					return true
				}
				// Ensure they're different
				if pluginName == depName {
					return true
				}

				resolver := NewDependencyResolver()

				// Add plugin with dependency on non-existent plugin
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{
						Name:     depName,
						Version:  "1.0.0",
						Optional: false,
					},
				}, nil)

				// Resolve should fail
				_, err := resolver.ResolveDependencies()

				// Should return an error about missing dependency
				return err != nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("missing optional dependency does not prevent loading",
		prop.ForAll(
			func(pluginName, depName string) bool {
				// Skip empty names
				if pluginName == "" || depName == "" {
					return true
				}
				// Ensure they're different
				if pluginName == depName {
					return true
				}

				resolver := NewDependencyResolver()

				// Add plugin with optional dependency on non-existent plugin
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{
						Name:     depName,
						Version:  "1.0.0",
						Optional: true,
					},
				}, nil)

				// Resolve should succeed
				loadOrder, err := resolver.ResolveDependencies()

				// Should not error and should include the plugin
				if err != nil {
					return false
				}

				if len(loadOrder) != 1 || loadOrder[0] != pluginName {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("all plugins present means no missing dependencies",
		prop.ForAll(
			func(pluginCount int) bool {
				// Limit plugin count
				if pluginCount < 2 || pluginCount > 10 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create plugins where each depends on the previous
				for i := 0; i < pluginCount; i++ {
					name := fmt.Sprintf("plugin-%d", i)
					var deps []PluginDependency

					if i > 0 {
						deps = []PluginDependency{
							{
								Name:    fmt.Sprintf("plugin-%d", i-1),
								Version: "1.0.0",
							},
						}
					}

					resolver.AddPlugin(name, "1.0.0", deps, nil)
				}

				// Resolve should succeed
				_, err := resolver.ResolveDependencies()

				return err == nil
			},
			gen.IntRange(2, 10),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 19: Incompatible dependency version detection**
// **Validates: Requirements 5.4**
// For any plugin with a dependency version mismatch, the system should prevent loading
// and log the version incompatibility
func TestProperty_VersionIncompatibilityDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("incompatible version constraint is detected",
		prop.ForAll(
			func(pluginName, depName string) bool {
				// Skip empty names
				if pluginName == "" || depName == "" {
					return true
				}
				// Ensure they're different
				if pluginName == depName {
					return true
				}

				resolver := NewDependencyResolver()

				// Add dependency with version 1.0.0
				resolver.AddPlugin(depName, "1.0.0", []PluginDependency{}, nil)

				// Add plugin requiring version >= 2.0.0
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{
						Name:    depName,
						Version: ">=2.0.0",
					},
				}, nil)

				// Resolve should fail
				_, err := resolver.ResolveDependencies()

				// Should return an error about version incompatibility
				return err != nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("compatible version constraint is accepted",
		prop.ForAll(
			func(pluginName, depName string) bool {
				// Skip empty names
				if pluginName == "" || depName == "" {
					return true
				}
				// Ensure they're different
				if pluginName == depName {
					return true
				}

				resolver := NewDependencyResolver()

				// Add dependency with version 1.5.0
				resolver.AddPlugin(depName, "1.5.0", []PluginDependency{}, nil)

				// Add plugin requiring version >= 1.0.0, < 2.0.0
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{
						Name:    depName,
						Version: ">=1.0.0,<2.0.0",
					},
				}, nil)

				// Resolve should succeed
				_, err := resolver.ResolveDependencies()

				return err == nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("exact version match is accepted",
		prop.ForAll(
			func(pluginName, depName string) bool {
				// Skip empty names
				if pluginName == "" || depName == "" {
					return true
				}
				// Ensure they're different
				if pluginName == depName {
					return true
				}

				resolver := NewDependencyResolver()

				// Add dependency with version 1.2.3
				resolver.AddPlugin(depName, "1.2.3", []PluginDependency{}, nil)

				// Add plugin requiring exact version 1.2.3
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{
						Name:    depName,
						Version: "1.2.3",
					},
				}, nil)

				// Resolve should succeed
				_, err := resolver.ResolveDependencies()

				return err == nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 20: Circular dependency detection**
// **Validates: Requirements 5.5**
// For any set of plugins with circular dependencies, the system should detect the cycle
// and prevent loading all plugins in the cycle
func TestProperty_CircularDependencyDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("simple circular dependency is detected",
		prop.ForAll(
			func(plugin1, plugin2 string) bool {
				// Skip empty names
				if plugin1 == "" || plugin2 == "" {
					return true
				}
				// Ensure they're different
				if plugin1 == plugin2 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create circular dependency: plugin1 -> plugin2 -> plugin1
				resolver.AddPlugin(plugin1, "1.0.0", []PluginDependency{
					{Name: plugin2, Version: "1.0.0"},
				}, nil)
				resolver.AddPlugin(plugin2, "1.0.0", []PluginDependency{
					{Name: plugin1, Version: "1.0.0"},
				}, nil)

				// Resolve should fail
				_, err := resolver.ResolveDependencies()

				// Should return an error about circular dependency
				return err != nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("self-dependency is detected",
		prop.ForAll(
			func(pluginName string) bool {
				// Skip empty names
				if pluginName == "" {
					return true
				}

				resolver := NewDependencyResolver()

				// Create self-dependency
				resolver.AddPlugin(pluginName, "1.0.0", []PluginDependency{
					{Name: pluginName, Version: "1.0.0"},
				}, nil)

				// Resolve should fail
				_, err := resolver.ResolveDependencies()

				// Should return an error about circular dependency
				return err != nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("three-way circular dependency is detected",
		prop.ForAll(
			func(plugin1, plugin2, plugin3 string) bool {
				// Skip empty names
				if plugin1 == "" || plugin2 == "" || plugin3 == "" {
					return true
				}
				// Ensure they're all different
				if plugin1 == plugin2 || plugin1 == plugin3 || plugin2 == plugin3 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create circular dependency: plugin1 -> plugin2 -> plugin3 -> plugin1
				resolver.AddPlugin(plugin1, "1.0.0", []PluginDependency{
					{Name: plugin2, Version: "1.0.0"},
				}, nil)
				resolver.AddPlugin(plugin2, "1.0.0", []PluginDependency{
					{Name: plugin3, Version: "1.0.0"},
				}, nil)
				resolver.AddPlugin(plugin3, "1.0.0", []PluginDependency{
					{Name: plugin1, Version: "1.0.0"},
				}, nil)

				// Resolve should fail
				_, err := resolver.ResolveDependencies()

				// Should return an error about circular dependency
				return err != nil
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("acyclic dependency graph is accepted",
		prop.ForAll(
			func(pluginCount int) bool {
				// Limit plugin count
				if pluginCount < 2 || pluginCount > 10 {
					return true
				}

				resolver := NewDependencyResolver()

				// Create acyclic chain: plugin-0 <- plugin-1 <- plugin-2 <- ...
				for i := 0; i < pluginCount; i++ {
					name := fmt.Sprintf("plugin-%d", i)
					var deps []PluginDependency

					if i > 0 {
						deps = []PluginDependency{
							{
								Name:    fmt.Sprintf("plugin-%d", i-1),
								Version: "1.0.0",
							},
						}
					}

					resolver.AddPlugin(name, "1.0.0", deps, nil)
				}

				// Resolve should succeed
				_, err := resolver.ResolveDependencies()

				return err == nil
			},
			gen.IntRange(2, 10),
		),
	)

	properties.TestingRun(t)
}
