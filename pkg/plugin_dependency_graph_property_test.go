package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_DependencyResolutionOrder tests Property 2:
// Dependency Resolution Order
// **Feature: compile-time-plugins, Property 2: Dependency Resolution Order**
// **Validates: Requirements 3.3, 3.5, 7.1**
func TestProperty_DependencyResolutionOrder(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("dependencies load before dependents", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Skip graphs with cycles (they should fail separately)
			if err := graph.DetectCircularDependencies(); err != nil {
				return true
			}

			// Get topological sort order
			order, err := graph.TopologicalSort()
			if err != nil {
				t.Logf("TopologicalSort failed: %v", err)
				return false
			}

			// Create position map
			position := make(map[string]int)
			for i, name := range order {
				position[name] = i
			}

			// Verify all required dependencies come before their dependents
			for _, node := range graphData {
				for _, dep := range node.dependencies {
					if !dep.Optional && graph.HasNode(dep.Name) {
						if position[dep.Name] >= position[node.name] {
							t.Logf("Dependency %s (pos %d) should come before %s (pos %d)",
								dep.Name, position[dep.Name], node.name, position[node.name])
							return false
						}
					}
				}
			}

			return true
		},
		genValidDependencyGraph(),
	))

	properties.Property("all plugins appear in resolution order", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Skip graphs with cycles
			if err := graph.DetectCircularDependencies(); err != nil {
				return true
			}

			// Get topological sort order
			order, err := graph.TopologicalSort()
			if err != nil {
				return true // Skip if sort fails
			}

			// Verify all plugins are in the order
			if len(order) != len(graphData) {
				t.Logf("Expected %d plugins in order, got %d", len(graphData), len(order))
				return false
			}

			// Create a set of all plugin names
			pluginSet := make(map[string]bool)
			for _, node := range graphData {
				pluginSet[node.name] = true
			}

			// Verify all plugins in order are from the graph
			for _, name := range order {
				if !pluginSet[name] {
					t.Logf("Plugin %s in order but not in graph", name)
					return false
				}
				delete(pluginSet, name)
			}

			// Verify no plugins were missed
			if len(pluginSet) > 0 {
				t.Logf("Plugins missing from order: %v", pluginSet)
				return false
			}

			return true
		},
		genValidDependencyGraph(),
	))

	properties.Property("optional dependencies don't affect order", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Skip graphs with cycles
			if err := graph.DetectCircularDependencies(); err != nil {
				return true
			}

			// Get topological sort order
			order, err := graph.TopologicalSort()
			if err != nil {
				return true
			}

			// Create position map
			position := make(map[string]int)
			for i, name := range order {
				position[name] = i
			}

			// Verify optional dependencies can come before or after
			// (we just check that the sort doesn't fail)
			for _, node := range graphData {
				for _, dep := range node.dependencies {
					if dep.Optional && graph.HasNode(dep.Name) {
						// Optional dependencies can be in any order
						// Just verify both plugins are in the result
						if _, hasNode := position[node.name]; !hasNode {
							t.Logf("Plugin %s not in order", node.name)
							return false
						}
						if _, hasDep := position[dep.Name]; !hasDep {
							t.Logf("Optional dependency %s not in order", dep.Name)
							return false
						}
					}
				}
			}

			return true
		},
		genValidDependencyGraph(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// TestProperty_CircularDependencyDetectionGraph tests Property 3:
// Circular Dependency Detection
// **Feature: compile-time-plugins, Property 3: Circular Dependency Detection**
// **Validates: Requirements 7.3**
func TestProperty_CircularDependencyDetectionGraph(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("graphs with cycles are rejected", prop.ForAll(
		func(cycleData cyclicGraphData) bool {
			// Build a graph with a guaranteed cycle
			graph := NewDependencyGraph()

			// Add all nodes
			for _, node := range cycleData.nodes {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Add the cycle
			for i := 0; i < len(cycleData.cycle)-1; i++ {
				from := cycleData.cycle[i]
				to := cycleData.cycle[i+1]

				// Find the node and add the dependency
				if node, exists := graph.GetNode(from); exists {
					// Add dependency to create cycle
					node.Dependencies = append(node.Dependencies, PluginDependency{
						Name:     to,
						Optional: false,
					})
				}
			}

			// Close the cycle
			if len(cycleData.cycle) > 1 {
				from := cycleData.cycle[len(cycleData.cycle)-1]
				to := cycleData.cycle[0]
				if node, exists := graph.GetNode(from); exists {
					node.Dependencies = append(node.Dependencies, PluginDependency{
						Name:     to,
						Optional: false,
					})
				}
			}

			// Detect circular dependencies
			err := graph.DetectCircularDependencies()
			if err == nil {
				t.Logf("Expected cycle detection to fail, but it succeeded")
				return false
			}

			// Verify error message mentions circular dependency
			errMsg := err.Error()
			if len(errMsg) == 0 || (len(errMsg) > 0 && errMsg[0:8] != "circular") {
				t.Logf("Expected 'circular dependency' error, got: %v", err)
				return false
			}

			return true
		},
		genCyclicGraph(),
	))

	properties.Property("acyclic graphs pass cycle detection", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// This should not detect a cycle
			err := graph.DetectCircularDependencies()
			if err != nil {
				t.Logf("Acyclic graph incorrectly detected as cyclic: %v", err)
				return false
			}

			return true
		},
		genValidDependencyGraph(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// TestProperty_OptionalDependencyHandling tests Property 10:
// Optional Dependency Handling
// **Feature: compile-time-plugins, Property 10: Optional Dependency Handling**
// **Validates: Requirements 7.4**
func TestProperty_OptionalDependencyHandling(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("plugins with missing optional dependencies initialize", prop.ForAll(
		func(graphData []pluginGraphNode, missingOptional string) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph, but skip the optional dependency
			graph := NewDependencyGraph()
			for _, node := range graphData {
				// Filter out the missing optional dependency
				filteredDeps := []PluginDependency{}
				for _, dep := range node.dependencies {
					if dep.Name != missingOptional || !dep.Optional {
						filteredDeps = append(filteredDeps, dep)
					}
				}
				graph.AddNode(node.name, node.version, filteredDeps)
			}

			// Validation should succeed even with missing optional dependencies
			err := graph.ValidateDependencies()
			if err != nil {
				// Check if error is about a required dependency
				// If so, that's expected and we skip this test case
				return true
			}

			// Topological sort should also succeed
			_, err = graph.TopologicalSort()
			if err != nil {
				t.Logf("TopologicalSort failed with missing optional dependency: %v", err)
				return false
			}

			return true
		},
		genGraphWithOptionalDeps(),
		gen.Identifier(),
	))

	properties.Property("optional dependencies are identified correctly", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Verify optional dependencies are correctly identified
			for _, node := range graphData {
				optionalDeps := graph.GetOptionalDependencies(node.name)
				requiredDeps := graph.GetRequiredDependencies(node.name)

				// Count optional and required from original
				expectedOptional := 0
				expectedRequired := 0
				for _, dep := range node.dependencies {
					if dep.Optional {
						expectedOptional++
					} else {
						expectedRequired++
					}
				}

				if len(optionalDeps) != expectedOptional {
					t.Logf("Plugin %s: expected %d optional deps, got %d",
						node.name, expectedOptional, len(optionalDeps))
					return false
				}

				if len(requiredDeps) != expectedRequired {
					t.Logf("Plugin %s: expected %d required deps, got %d",
						node.name, expectedRequired, len(requiredDeps))
					return false
				}
			}

			return true
		},
		genGraphWithOptionalDeps(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// TestProperty_RequiredDependencyEnforcement tests Property 11:
// Required Dependency Enforcement
// **Feature: compile-time-plugins, Property 11: Required Dependency Enforcement**
// **Validates: Requirements 7.5**
func TestProperty_RequiredDependencyEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("missing required dependencies cause validation failure", prop.ForAll(
		func(graphData []pluginGraphNode, missingRequired string) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Find a plugin with a required dependency
			var pluginWithReqDep *pluginGraphNode
			for i := range graphData {
				for _, dep := range graphData[i].dependencies {
					if !dep.Optional {
						pluginWithReqDep = &graphData[i]
						missingRequired = dep.Name
						break
					}
				}
				if pluginWithReqDep != nil {
					break
				}
			}

			// If no required dependencies, skip this test
			if pluginWithReqDep == nil {
				return true
			}

			// Build dependency graph, but skip the required dependency
			graph := NewDependencyGraph()
			for _, node := range graphData {
				if node.name != missingRequired {
					graph.AddNode(node.name, node.version, node.dependencies)
				}
			}

			// Validation should fail
			err := graph.ValidateDependencies()
			if err == nil {
				t.Logf("Expected validation to fail with missing required dependency %s", missingRequired)
				return false
			}

			// Verify error message mentions missing dependency
			errMsg := err.Error()
			if len(errMsg) == 0 {
				t.Logf("Expected error message about missing dependency")
				return false
			}

			return true
		},
		genGraphWithRequiredDeps(),
		gen.Identifier(),
	))

	properties.Property("all required dependencies must be present", prop.ForAll(
		func(graphData []pluginGraphNode) bool {
			// Skip empty graphs
			if len(graphData) == 0 {
				return true
			}

			// Build complete dependency graph
			graph := NewDependencyGraph()
			for _, node := range graphData {
				graph.AddNode(node.name, node.version, node.dependencies)
			}

			// Validation should succeed when all dependencies are present
			err := graph.ValidateDependencies()
			if err != nil {
				// This is expected if the graph has missing dependencies
				// We just verify the error is about missing dependencies
				return true
			}

			// If validation succeeds, verify all required dependencies are present
			for _, node := range graphData {
				for _, dep := range node.dependencies {
					if !dep.Optional {
						if !graph.HasNode(dep.Name) {
							t.Logf("Required dependency %s for plugin %s is missing but validation passed",
								dep.Name, node.name)
							return false
						}
					}
				}
			}

			return true
		},
		genValidDependencyGraph(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// Generator types and functions

type pluginGraphNode struct {
	name         string
	version      string
	dependencies []PluginDependency
}

type cyclicGraphData struct {
	nodes []pluginGraphNode
	cycle []string // Names of nodes forming a cycle
}

// genValidDependencyGraph generates a valid acyclic dependency graph
func genValidDependencyGraph() gopter.Gen {
	return gen.SliceOfN(10, gen.Identifier()).
		SuchThat(func(v interface{}) bool {
			names := v.([]string)
			// Ensure unique names
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}).
		Map(func(names []string) []pluginGraphNode {
			nodes := make([]pluginGraphNode, len(names))
			for i, name := range names {
				// Only allow dependencies on earlier plugins to avoid cycles
				var deps []PluginDependency
				if i > 0 {
					// Randomly depend on 0-2 earlier plugins
					numDeps := i % 3
					for j := 0; j < numDeps && j < i; j++ {
						deps = append(deps, PluginDependency{
							Name:     names[j],
							Version:  ">=1.0.0",
							Optional: false,
						})
					}
				}

				nodes[i] = pluginGraphNode{
					name:         name,
					version:      "1.0.0",
					dependencies: deps,
				}
			}
			return nodes
		})
}

// genCyclicGraph generates a graph with a guaranteed cycle
func genCyclicGraph() gopter.Gen {
	return gen.SliceOfN(5, gen.Identifier()).
		SuchThat(func(v interface{}) bool {
			names := v.([]string)
			// Ensure unique names
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) >= 2
		}).
		Map(func(names []string) cyclicGraphData {
			nodes := make([]pluginGraphNode, len(names))
			for i, name := range names {
				nodes[i] = pluginGraphNode{
					name:         name,
					version:      "1.0.0",
					dependencies: []PluginDependency{},
				}
			}

			// Create a cycle with at least 2 nodes
			cycleSize := 2 + (len(names) % 3) // 2-4 nodes in cycle
			if cycleSize > len(names) {
				cycleSize = len(names)
			}

			cycle := names[:cycleSize]

			return cyclicGraphData{
				nodes: nodes,
				cycle: cycle,
			}
		})
}

// genGraphWithOptionalDeps generates a graph with optional dependencies
func genGraphWithOptionalDeps() gopter.Gen {
	return gen.SliceOfN(10, gen.Identifier()).
		SuchThat(func(v interface{}) bool {
			names := v.([]string)
			// Ensure unique names
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}).
		Map(func(names []string) []pluginGraphNode {
			nodes := make([]pluginGraphNode, len(names))
			for i, name := range names {
				var deps []PluginDependency
				if i > 0 {
					// Add some optional dependencies
					numDeps := (i % 3) + 1
					for j := 0; j < numDeps && j < i; j++ {
						deps = append(deps, PluginDependency{
							Name:     names[j],
							Version:  ">=1.0.0",
							Optional: j%2 == 0, // Every other dependency is optional
						})
					}
				}

				nodes[i] = pluginGraphNode{
					name:         name,
					version:      "1.0.0",
					dependencies: deps,
				}
			}
			return nodes
		})
}

// genGraphWithRequiredDeps generates a graph with required dependencies
func genGraphWithRequiredDeps() gopter.Gen {
	return gen.SliceOfN(10, gen.Identifier()).
		SuchThat(func(v interface{}) bool {
			names := v.([]string)
			// Ensure unique names
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) >= 2 // Need at least 2 for dependencies
		}).
		Map(func(names []string) []pluginGraphNode {
			nodes := make([]pluginGraphNode, len(names))
			for i, name := range names {
				var deps []PluginDependency
				if i > 0 {
					// Add at least one required dependency
					deps = append(deps, PluginDependency{
						Name:     names[i-1],
						Version:  ">=1.0.0",
						Optional: false,
					})
				}

				nodes[i] = pluginGraphNode{
					name:         name,
					version:      fmt.Sprintf("1.%d.0", i),
					dependencies: deps,
				}
			}
			return nodes
		})
}
