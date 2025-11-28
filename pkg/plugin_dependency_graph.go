package pkg

import (
	"fmt"
	"strings"
)

// DependencyGraph represents the dependency relationships between plugins
type DependencyGraph struct {
	nodes map[string]*DependencyNode
	edges map[string][]string // adjacency list: plugin -> dependents
}

// DependencyNode represents a plugin node in the dependency graph
type DependencyNode struct {
	Name         string
	Version      string
	Dependencies []PluginDependency
	Optional     bool
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*DependencyNode),
		edges: make(map[string][]string),
	}
}

// AddNode adds a plugin node to the graph
func (g *DependencyGraph) AddNode(name, version string, deps []PluginDependency) {
	g.nodes[name] = &DependencyNode{
		Name:         name,
		Version:      version,
		Dependencies: deps,
		Optional:     false,
	}

	// Initialize edges if not present
	if _, exists := g.edges[name]; !exists {
		g.edges[name] = []string{}
	}
}

// AddEdge adds a dependency edge from dependent to dependency
func (g *DependencyGraph) AddEdge(dependent, dependency string) {
	if _, exists := g.edges[dependency]; !exists {
		g.edges[dependency] = []string{}
	}
	g.edges[dependency] = append(g.edges[dependency], dependent)
}

// GetNode retrieves a node by name
func (g *DependencyGraph) GetNode(name string) (*DependencyNode, bool) {
	node, exists := g.nodes[name]
	return node, exists
}

// HasNode checks if a node exists in the graph
func (g *DependencyGraph) HasNode(name string) bool {
	_, exists := g.nodes[name]
	return exists
}

// GetDependents returns all plugins that depend on the given plugin
func (g *DependencyGraph) GetDependents(name string) []string {
	if dependents, exists := g.edges[name]; exists {
		return dependents
	}
	return []string{}
}

// GetDependencies returns all dependencies of the given plugin
func (g *DependencyGraph) GetDependencies(name string) []PluginDependency {
	if node, exists := g.nodes[name]; exists {
		return node.Dependencies
	}
	return []PluginDependency{}
}

// TopologicalSort performs a topological sort on the dependency graph
// Returns the plugins in dependency order (dependencies before dependents)
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	// Check for circular dependencies first
	if err := g.DetectCircularDependencies(); err != nil {
		return nil, err
	}

	// Calculate in-degrees
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize
	for name := range g.nodes {
		inDegree[name] = 0
		adjList[name] = []string{}
	}

	// Build adjacency list and calculate in-degrees
	// For each plugin, add edges from its dependencies to it
	for name, node := range g.nodes {
		for _, dep := range node.Dependencies {
			// Only count required dependencies for topological sort
			if !dep.Optional && g.HasNode(dep.Name) {
				adjList[dep.Name] = append(adjList[dep.Name], name)
				inDegree[name]++
			}
		}
	}

	// Kahn's algorithm for topological sort
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we didn't process all nodes, there's a cycle
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("failed to resolve dependencies: cycle detected")
	}

	return result, nil
}

// DetectCircularDependencies checks for circular dependencies in the graph
func (g *DependencyGraph) DetectCircularDependencies() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range g.nodes {
		if err := g.detectCycle(name, visited, recStack, []string{name}); err != nil {
			return err
		}
	}

	return nil
}

// detectCycle performs DFS to detect cycles
func (g *DependencyGraph) detectCycle(pluginName string, visited, recStack map[string]bool, path []string) error {
	if recStack[pluginName] {
		// Found a cycle
		cycleStart := 0
		for i, name := range path {
			if name == pluginName {
				cycleStart = i
				break
			}
		}
		cycle := append(path[cycleStart:], pluginName)
		return fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
	}

	if visited[pluginName] {
		return nil
	}

	visited[pluginName] = true
	recStack[pluginName] = true

	node, exists := g.nodes[pluginName]
	if exists {
		for _, dep := range node.Dependencies {
			// Only check required dependencies for cycles
			if !dep.Optional && g.HasNode(dep.Name) {
				newPath := append([]string{}, path...)
				newPath = append(newPath, dep.Name)
				if err := g.detectCycle(dep.Name, visited, recStack, newPath); err != nil {
					return err
				}
			}
		}
	}

	recStack[pluginName] = false
	return nil
}

// ValidateDependencies validates all dependencies in the graph
func (g *DependencyGraph) ValidateDependencies() error {
	// Check for missing required dependencies
	for name, node := range g.nodes {
		for _, dep := range node.Dependencies {
			if dep.Optional {
				continue
			}
			if !g.HasNode(dep.Name) {
				return fmt.Errorf("plugin %s: missing required dependency '%s'", name, dep.Name)
			}
		}
	}

	// Check for version compatibility
	for name, node := range g.nodes {
		for _, dep := range node.Dependencies {
			depNode, exists := g.nodes[dep.Name]
			if !exists {
				// Skip if optional
				if dep.Optional {
					continue
				}
				// This should have been caught above
				continue
			}

			// Check version constraint
			satisfied, err := satisfiesVersionConstraint(depNode.Version, dep.Version)
			if err != nil {
				return fmt.Errorf("plugin %s: invalid version constraint '%s' for dependency '%s': %w",
					name, dep.Version, dep.Name, err)
			}

			if !satisfied {
				return fmt.Errorf("plugin %s: dependency '%s' version %s does not satisfy constraint %s",
					name, dep.Name, depNode.Version, dep.Version)
			}
		}
	}

	return nil
}

// GetOptionalDependencies returns all optional dependencies for a plugin
func (g *DependencyGraph) GetOptionalDependencies(name string) []PluginDependency {
	node, exists := g.nodes[name]
	if !exists {
		return []PluginDependency{}
	}

	var optional []PluginDependency
	for _, dep := range node.Dependencies {
		if dep.Optional {
			optional = append(optional, dep)
		}
	}
	return optional
}

// GetRequiredDependencies returns all required dependencies for a plugin
func (g *DependencyGraph) GetRequiredDependencies(name string) []PluginDependency {
	node, exists := g.nodes[name]
	if !exists {
		return []PluginDependency{}
	}

	var required []PluginDependency
	for _, dep := range node.Dependencies {
		if !dep.Optional {
			required = append(required, dep)
		}
	}
	return required
}

// Size returns the number of nodes in the graph
func (g *DependencyGraph) Size() int {
	return len(g.nodes)
}

// IsEmpty returns true if the graph has no nodes
func (g *DependencyGraph) IsEmpty() bool {
	return len(g.nodes) == 0
}
