package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

// FrameworkVersion is the current version of the Rockstar Web Framework
const FrameworkVersion = "1.0.0"

// DependencyResolver handles plugin dependency resolution
type DependencyResolver struct {
	plugins map[string]*PluginWithManifest
}

// PluginWithManifest wraps a plugin with its manifest information
type PluginWithManifest struct {
	Name         string
	Version      string
	Dependencies []PluginDependency
	Manifest     *PluginManifest
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		plugins: make(map[string]*PluginWithManifest),
	}
}

// AddPlugin adds a plugin to the resolver
func (r *DependencyResolver) AddPlugin(name, version string, deps []PluginDependency, manifest *PluginManifest) {
	r.plugins[name] = &PluginWithManifest{
		Name:         name,
		Version:      version,
		Dependencies: deps,
		Manifest:     manifest,
	}
}

// ValidateFrameworkVersion validates that a plugin's framework version requirement is satisfied
func (r *DependencyResolver) ValidateFrameworkVersion(frameworkVersionConstraint string) error {
	if frameworkVersionConstraint == "" {
		// No constraint means any version is acceptable
		return nil
	}

	satisfied, err := satisfiesVersionConstraint(FrameworkVersion, frameworkVersionConstraint)
	if err != nil {
		return fmt.Errorf("invalid framework version constraint '%s': %w", frameworkVersionConstraint, err)
	}

	if !satisfied {
		return fmt.Errorf("framework version %s does not satisfy constraint %s", FrameworkVersion, frameworkVersionConstraint)
	}

	return nil
}

// ResolveDependencies resolves all dependencies and returns the load order
func (r *DependencyResolver) ResolveDependencies() ([]string, error) {
	// First, validate all framework version requirements
	for _, plugin := range r.plugins {
		for _, dep := range plugin.Dependencies {
			if dep.FrameworkVersion != "" {
				if err := r.ValidateFrameworkVersion(dep.FrameworkVersion); err != nil {
					return nil, fmt.Errorf("plugin %s: %w", plugin.Name, err)
				}
			}
		}
		// Also check the manifest framework version if available
		if plugin.Manifest != nil && plugin.Manifest.Framework.Version != "" {
			if err := r.ValidateFrameworkVersion(plugin.Manifest.Framework.Version); err != nil {
				return nil, fmt.Errorf("plugin %s: %w", plugin.Name, err)
			}
		}
	}

	// Check for missing dependencies
	if err := r.checkMissingDependencies(); err != nil {
		return nil, err
	}

	// Check for version incompatibilities
	if err := r.checkVersionCompatibility(); err != nil {
		return nil, err
	}

	// Check for circular dependencies
	if err := r.checkCircularDependencies(); err != nil {
		return nil, err
	}

	// Build dependency graph and determine load order
	return r.buildLoadOrder()
}

// checkMissingDependencies checks if all required dependencies are present
func (r *DependencyResolver) checkMissingDependencies() error {
	for _, plugin := range r.plugins {
		for _, dep := range plugin.Dependencies {
			if dep.Optional {
				continue
			}
			if _, exists := r.plugins[dep.Name]; !exists {
				return fmt.Errorf("plugin %s: missing required dependency '%s'", plugin.Name, dep.Name)
			}
		}
	}
	return nil
}

// checkVersionCompatibility checks if dependency versions are compatible
func (r *DependencyResolver) checkVersionCompatibility() error {
	for _, plugin := range r.plugins {
		for _, dep := range plugin.Dependencies {
			depPlugin, exists := r.plugins[dep.Name]
			if !exists {
				// Skip if optional
				if dep.Optional {
					continue
				}
				// This should have been caught by checkMissingDependencies
				continue
			}

			// Check version constraint
			satisfied, err := satisfiesVersionConstraint(depPlugin.Version, dep.Version)
			if err != nil {
				return fmt.Errorf("plugin %s: invalid version constraint '%s' for dependency '%s': %w",
					plugin.Name, dep.Version, dep.Name, err)
			}

			if !satisfied {
				return fmt.Errorf("plugin %s: dependency '%s' version %s does not satisfy constraint %s",
					plugin.Name, dep.Name, depPlugin.Version, dep.Version)
			}
		}
	}
	return nil
}

// checkCircularDependencies detects circular dependencies
func (r *DependencyResolver) checkCircularDependencies() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range r.plugins {
		if err := r.detectCycle(name, visited, recStack, []string{name}); err != nil {
			return err
		}
	}

	return nil
}

// detectCycle performs DFS to detect cycles
func (r *DependencyResolver) detectCycle(pluginName string, visited, recStack map[string]bool, path []string) error {
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

	plugin, exists := r.plugins[pluginName]
	if exists {
		for _, dep := range plugin.Dependencies {
			if _, depExists := r.plugins[dep.Name]; depExists {
				newPath := append([]string{}, path...)
				newPath = append(newPath, dep.Name)
				if err := r.detectCycle(dep.Name, visited, recStack, newPath); err != nil {
					return err
				}
			}
		}
	}

	recStack[pluginName] = false
	return nil
}

// buildLoadOrder builds the correct load order using topological sort
func (r *DependencyResolver) buildLoadOrder() ([]string, error) {
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize
	for name := range r.plugins {
		inDegree[name] = 0
		adjList[name] = []string{}
	}

	// Build adjacency list and in-degree count
	for name, plugin := range r.plugins {
		for _, dep := range plugin.Dependencies {
			if _, exists := r.plugins[dep.Name]; exists {
				adjList[dep.Name] = append(adjList[dep.Name], name)
				inDegree[name]++
			}
		}
	}

	// Topological sort using Kahn's algorithm
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

	if len(result) != len(r.plugins) {
		return nil, fmt.Errorf("failed to resolve dependencies: cycle detected")
	}

	return result, nil
}

// satisfiesVersionConstraint checks if a version satisfies a constraint
// Constraint format: ">=1.0.0", ">=1.0.0,<2.0.0", "1.0.0", etc.
func satisfiesVersionConstraint(version, constraint string) (bool, error) {
	if constraint == "" {
		return true, nil
	}

	// Parse version
	v, err := parseVersion(version)
	if err != nil {
		return false, fmt.Errorf("invalid version '%s': %w", version, err)
	}

	// Split constraint by comma for multiple constraints
	constraints := strings.Split(constraint, ",")
	for _, c := range constraints {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		satisfied, err := checkSingleConstraint(v, c)
		if err != nil {
			return false, err
		}
		if !satisfied {
			return false, nil
		}
	}

	return true, nil
}

// checkSingleConstraint checks a single version constraint
func checkSingleConstraint(version *semVersion, constraint string) (bool, error) {
	// Parse operator and version
	var op string
	var versionStr string

	if strings.HasPrefix(constraint, ">=") {
		op = ">="
		versionStr = strings.TrimSpace(constraint[2:])
	} else if strings.HasPrefix(constraint, "<=") {
		op = "<="
		versionStr = strings.TrimSpace(constraint[2:])
	} else if strings.HasPrefix(constraint, ">") {
		op = ">"
		versionStr = strings.TrimSpace(constraint[1:])
	} else if strings.HasPrefix(constraint, "<") {
		op = "<"
		versionStr = strings.TrimSpace(constraint[1:])
	} else if strings.HasPrefix(constraint, "=") {
		op = "="
		versionStr = strings.TrimSpace(constraint[1:])
	} else {
		// No operator means exact match
		op = "="
		versionStr = constraint
	}

	constraintVersion, err := parseVersion(versionStr)
	if err != nil {
		return false, fmt.Errorf("invalid constraint version '%s': %w", versionStr, err)
	}

	cmp := version.compare(constraintVersion)

	switch op {
	case ">=":
		return cmp >= 0, nil
	case "<=":
		return cmp <= 0, nil
	case ">":
		return cmp > 0, nil
	case "<":
		return cmp < 0, nil
	case "=":
		return cmp == 0, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}

// semVersion represents a semantic version
type semVersion struct {
	major int
	minor int
	patch int
	pre   string
}

// parseVersion parses a semantic version string
func parseVersion(version string) (*semVersion, error) {
	// Handle pre-release versions (e.g., "1.0.0-alpha")
	parts := strings.SplitN(version, "-", 2)
	versionPart := parts[0]
	var preRelease string
	if len(parts) > 1 {
		preRelease = parts[1]
	}

	// Split version into major.minor.patch
	versionParts := strings.Split(versionPart, ".")
	if len(versionParts) < 2 || len(versionParts) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", versionParts[0])
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	patch := 0
	if len(versionParts) == 3 {
		patch, err = strconv.Atoi(versionParts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", versionParts[2])
		}
	}

	return &semVersion{
		major: major,
		minor: minor,
		patch: patch,
		pre:   preRelease,
	}, nil
}

// compare compares two semantic versions
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v *semVersion) compare(other *semVersion) int {
	if v.major != other.major {
		if v.major < other.major {
			return -1
		}
		return 1
	}

	if v.minor != other.minor {
		if v.minor < other.minor {
			return -1
		}
		return 1
	}

	if v.patch != other.patch {
		if v.patch < other.patch {
			return -1
		}
		return 1
	}

	// If both have no pre-release, they're equal
	if v.pre == "" && other.pre == "" {
		return 0
	}

	// Version with pre-release is less than version without
	if v.pre != "" && other.pre == "" {
		return -1
	}
	if v.pre == "" && other.pre != "" {
		return 1
	}

	// Both have pre-release, compare lexicographically
	if v.pre < other.pre {
		return -1
	}
	if v.pre > other.pre {
		return 1
	}

	return 0
}
