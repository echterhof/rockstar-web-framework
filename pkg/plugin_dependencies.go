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
	graph   *DependencyGraph
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
		graph:   NewDependencyGraph(),
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

	// Add to dependency graph
	r.graph.AddNode(name, version, deps)

	// Add edges for dependencies
	for _, dep := range deps {
		r.graph.AddEdge(name, dep.Name)
	}
}

// GetDependencyGraph returns the dependency graph
func (r *DependencyResolver) GetDependencyGraph() *DependencyGraph {
	return r.graph
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

	// Validate all dependencies using the graph
	if err := r.graph.ValidateDependencies(); err != nil {
		return nil, err
	}

	// Perform topological sort to get load order
	return r.graph.TopologicalSort()
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
