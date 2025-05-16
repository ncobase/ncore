package registry

import (
	"fmt"
	"sort"
	"sync"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// ExtensionEntry represents an entry in the registry
type ExtensionEntry struct {
	Instance         ext.Interface
	Group            string
	WeakDependencies []string // Optional dependencies that won't block initialization
}

var (
	// extensionRegistry stores all registered extensions
	extensionRegistry = make(map[string]ExtensionEntry)
	// mutex protects concurrent access to registry
	mutex = &sync.RWMutex{}
)

// Register registers a single extension
func Register(extension ext.Interface) {
	RegisterToGroup(extension, "")
}

// RegisterToGroup registers an extension to a specific group
func RegisterToGroup(extension ext.Interface, group string) {
	mutex.Lock()
	defer mutex.Unlock()

	name := extension.Name()
	extensionRegistry[name] = ExtensionEntry{
		Instance: extension,
		Group:    group,
	}
}

// RegisterWithWeakDeps registers an extension with weak dependencies
func RegisterWithWeakDeps(extension ext.Interface, weakDeps []string) {
	RegisterToGroupWithWeakDeps(extension, "", weakDeps)
}

// RegisterToGroupWithWeakDeps registers an extension to a group with weak dependencies
func RegisterToGroupWithWeakDeps(extension ext.Interface, group string, weakDeps []string) {
	mutex.Lock()
	defer mutex.Unlock()

	name := extension.Name()
	extensionRegistry[name] = ExtensionEntry{
		Instance:         extension,
		Group:            group,
		WeakDependencies: weakDeps,
	}
}

// GetExtensions returns all registered extensions
func GetExtensions() map[string]ExtensionEntry {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make(map[string]ExtensionEntry)
	for k, v := range extensionRegistry {
		result[k] = v
	}
	return result
}

// GetExtensionsByGroup returns extensions in a specific group
func GetExtensionsByGroup(groupName string) map[string]ExtensionEntry {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make(map[string]ExtensionEntry)
	for k, v := range extensionRegistry {
		if v.Group == groupName {
			result[k] = v
		}
	}
	return result
}

// detectCircularDependencies checks for circular dependencies in the graph
func detectCircularDependencies(graph map[string][]string) (bool, []string) {
	// Visited nodes in the DFS traversal
	visited := make(map[string]int) // 0=not visited, 1=visiting, 2=visited

	// Stack to track path
	var path []string
	var cycles []string

	// Helper function for DFS traversal
	var dfs func(node string) bool
	dfs = func(node string) bool {
		// Set node as being visited
		visited[node] = 1
		path = append(path, node)

		// Visit each dependency
		for _, dep := range graph[node] {
			// If not visited yet
			if visited[dep] == 0 {
				if dfs(dep) {
					return true // Cycle detected in a deeper level
				}
			} else if visited[dep] == 1 {
				// Found a cycle
				cycles = append(cycles, fmt.Sprintf("Cycle detected: %s -> %s", node, dep))
				return true
			}
		}

		// Remove node from path and mark as fully visited
		path = path[:len(path)-1]
		visited[node] = 2
		return false
	}

	// Run DFS from each unvisited node
	for node := range graph {
		if visited[node] == 0 {
			if dfs(node) {
				return true, cycles
			}
		}
	}

	return false, nil
}

// resolveDependencyGraph resolves circular dependencies by marking certain dependencies as weak
func resolveDependencyGraph(_ map[string]ext.Interface, graph map[string][]string) map[string][]string {
	hasCycles, cycleInfo := detectCircularDependencies(graph)
	if !hasCycles {
		return graph
	}

	logger.Infof(nil, "Circular dependencies detected, attempting to resolve: %v", cycleInfo)

	// Create a copy of the graph to work with
	resolvedGraph := make(map[string][]string)
	for node, deps := range graph {
		depsCopy := make([]string, len(deps))
		copy(depsCopy, deps)
		resolvedGraph[node] = depsCopy
	}

	// Sort nodes by number of dependencies to prioritize "core" modules
	type nodeInfo struct {
		name string
		deps int
	}
	nodes := make([]nodeInfo, 0, len(resolvedGraph))
	for node, deps := range resolvedGraph {
		nodes = append(nodes, nodeInfo{node, len(deps)})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].deps < nodes[j].deps
	})

	// Define core modules that others depend on
	coreModules := map[string]bool{}

	// Find non-core modules in cycles and break them
	for _, ext := range extensionRegistry {
		for node, deps := range resolvedGraph {
			// Skip core modules, we don't want to modify their dependencies
			if coreModules[node] {
				continue
			}

			// Check for circular dependencies
			for _, dep := range deps {
				// If the dependency is this module, remove it
				if dep == ext.Instance.Name() {
					// Check if this dependency is in the weak dependencies list
					weakDeps := ext.WeakDependencies
					isWeak := false
					for _, weakDep := range weakDeps {
						if weakDep == dep {
							isWeak = true
							break
						}
					}

					// If it's a weak dependency, we can safely break the cycle
					if isWeak {
						newDeps := make([]string, 0)
						for _, d := range resolvedGraph[node] {
							if d != dep {
								newDeps = append(newDeps, d)
							}
						}
						resolvedGraph[node] = newDeps
						logger.Infof(nil, "Breaking circular dependency: removed %s from %s dependencies", dep, node)
					}
				}
			}
		}
	}

	// Final check - if we still have cycles, take more aggressive measures
	hasCycles, _ = detectCircularDependencies(resolvedGraph)
	if hasCycles {
		logger.Warnf(nil, "Still have circular dependencies after initial resolution, taking more aggressive measures")

		// Break all cycles by prioritizing modules with fewer dependencies
		for i, node := range nodes {
			for j := i + 1; j < len(nodes); j++ {
				otherNode := nodes[j]

				// Check if these two nodes depend on each other
				dependsOnOther := false
				for _, dep := range resolvedGraph[node.name] {
					if dep == otherNode.name {
						dependsOnOther = true
						break
					}
				}

				otherDependsOn := false
				for _, dep := range resolvedGraph[otherNode.name] {
					if dep == node.name {
						otherDependsOn = true
						break
					}
				}

				// If they depend on each other, break the cycle
				if dependsOnOther && otherDependsOn {
					// Remove the dependency from the node with more dependencies
					if node.deps <= otherNode.deps {
						newDeps := make([]string, 0)
						for _, dep := range resolvedGraph[otherNode.name] {
							if dep != node.name {
								newDeps = append(newDeps, dep)
							}
						}
						resolvedGraph[otherNode.name] = newDeps
						logger.Infof(nil, "Breaking cyclical dependency: removed %s from %s dependencies", node.name, otherNode.name)
					} else {
						newDeps := make([]string, 0)
						for _, dep := range resolvedGraph[node.name] {
							if dep != otherNode.name {
								newDeps = append(newDeps, dep)
							}
						}
						resolvedGraph[node.name] = newDeps
						logger.Infof(nil, "Breaking cyclical dependency: removed %s from %s dependencies", otherNode.name, node.name)
					}
				}
			}
		}
	}

	return resolvedGraph
}

// GetExtensionsAndDependencies returns all extensions and their dependency graph
func GetExtensionsAndDependencies() (map[string]ext.Interface, map[string][]string) {
	mutex.RLock()
	defer mutex.RUnlock()

	instances := make(map[string]ext.Interface)
	dependencyGraph := make(map[string][]string)

	// First collect all extension instances
	for name, entry := range extensionRegistry {
		instances[name] = entry.Instance
	}

	// Then build the dependency graph
	for name, entry := range extensionRegistry {
		// Add strong dependencies
		deps := make([]string, 0)
		deps = append(deps, entry.Instance.Dependencies()...)

		// Process weak dependencies - only add those that exist
		for _, weakDep := range entry.WeakDependencies {
			if _, exists := instances[weakDep]; exists {
				deps = append(deps, weakDep)
			}
		}

		dependencyGraph[name] = deps
	}

	// Resolve circular dependencies
	resolvedGraph := resolveDependencyGraph(instances, dependencyGraph)

	return instances, resolvedGraph
}

// ClearRegistry clears the registry (mainly for testing)
func ClearRegistry() {
	mutex.Lock()
	defer mutex.Unlock()
	extensionRegistry = make(map[string]ExtensionEntry)
}
