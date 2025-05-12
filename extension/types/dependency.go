package types

// DependencyType defines the type of dependency
type DependencyType string

const (
	// StrongDependency indicates a required dependency
	StrongDependency DependencyType = "strong"
	// WeakDependency indicates an optional dependency
	WeakDependency DependencyType = "weak"
)

// DependencyEntry represents a dependency with its type
type DependencyEntry struct {
	Name string
	Type DependencyType
}

// GetStrongDependencies filters and returns only strong dependencies
func GetStrongDependencies(deps []DependencyEntry) []string {
	var result []string
	for _, dep := range deps {
		if dep.Type == StrongDependency {
			result = append(result, dep.Name)
		}
	}
	return result
}

// GetWeakDependencies filters and returns only weak dependencies
func GetWeakDependencies(deps []DependencyEntry) []string {
	var result []string
	for _, dep := range deps {
		if dep.Type == WeakDependency {
			result = append(result, dep.Name)
		}
	}
	return result
}
