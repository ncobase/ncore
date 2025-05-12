package registry

import (
	"sync"

	ext "github.com/ncobase/ncore/extension/types"
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

	return instances, dependencyGraph
}

// ClearRegistry clears the registry (mainly for testing)
func ClearRegistry() {
	mutex.Lock()
	defer mutex.Unlock()
	extensionRegistry = make(map[string]ExtensionEntry)
}
