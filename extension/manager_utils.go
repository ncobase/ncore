package extension

import (
	"fmt"
	"ncobase/common/config"
	"sort"
)

// // validateConfig validates the configuration
// func validateConfig(conf *config.Config) error {
// 	if conf == nil {
// 		return fmt.Errorf("configuration cannot be nil")
// 	}
// 	if conf.Extension == nil {
// 		return fmt.Errorf("extension configuration cannot be nil")
// 	}
// 	if conf.Extension.Path == "" {
// 		return fmt.Errorf("extension path is required")
// 	}
// 	return nil
// }

// initializePlugin initializes a single plugin
func (m *Manager) initializePlugin(c *Wrapper) error {
	if err := c.Instance.PreInit(); err != nil {
		return fmt.Errorf("failed pre-initialization: %v", err)
	}
	if err := c.Instance.Init(m.conf, m); err != nil {
		return fmt.Errorf("failed initialization: %v", err)
	}
	if err := c.Instance.PostInit(); err != nil {
		return fmt.Errorf("failed post-initialization: %v", err)
	}
	return nil
}

// checkDependencies checks if all dependencies are loaded
func (m *Manager) checkDependencies() error {
	for name, extension := range m.extensions {
		for _, dep := range extension.Instance.Dependencies() {
			if _, ok := m.extensions[dep]; !ok {
				return fmt.Errorf("extension '%s' depends on '%s', which is not available", name, dep)
			}
		}
	}
	return nil
}

// getInitOrder returns the initialization order based on dependencies
//
// noDeps - modules with no dependencies, first to initialize
// withDeps - modules with dependencies
// special - special modules that should be ordered last
func getInitOrder(extensions map[string]*Wrapper) ([]string, error) {
	var noDeps, withDeps, special []string
	// Exclude these modules from dependency check, adjust as needed
	var specialModules []string
	// Example of how to add modules dynamically
	// specialModules = append(specialModules, "relation", "relations", "linker", "linkers")
	specialSet := make(map[string]bool)
	for _, m := range specialModules {
		specialSet[m] = true
	}

	dependencies := make(map[string]map[string]bool)
	initialized := make(map[string]bool)

	// Collect all available extension names
	availableExtensions := make(map[string]bool)
	for name := range extensions {
		availableExtensions[name] = true
	}

	// analyze dependencies, classify modules into noDeps and withDeps
	for name, extension := range extensions {
		if specialSet[name] {
			special = append(special, name)
			continue
		}

		deps := make(map[string]bool)

		// Check if all dependencies exist
		for _, dep := range extension.Metadata.Dependencies {
			if !specialSet[dep] {
				// Check if dependency exists in extensions
				if !availableExtensions[dep] {
					return nil, fmt.Errorf("extension '%s' depends on '%s' which does not exist", name, dep)
				}
				deps[dep] = true
			}
		}

		if len(deps) == 0 {
			noDeps = append(noDeps, name)
			initialized[name] = true
		} else {
			withDeps = append(withDeps, name)
			dependencies[name] = deps
		}
	}

	// sort noDeps modules
	sort.Strings(noDeps)

	// sort withDeps modules
	var order []string
	order = append(order, noDeps...)

	for len(withDeps) > 0 {
		progress := false
		remainingDeps := withDeps[:0]

		for _, name := range withDeps {
			canInitialize := true
			for dep := range dependencies[name] {
				if !initialized[dep] {
					canInitialize = false
					break
				}
			}

			if canInitialize {
				order = append(order, name)
				initialized[name] = true
				progress = true
			} else {
				remainingDeps = append(remainingDeps, name)
			}
		}

		if !progress {
			return nil, fmt.Errorf("cyclic dependency detected")
		}

		withDeps = remainingDeps
	}

	// add special modules
	order = append(order, special...)

	return order, nil
}

// shouldLoadPlugin returns true if the plugin should be loaded
func (m *Manager) shouldLoadPlugin(name string) bool {
	fc := m.conf.Extension

	if len(fc.Includes) > 0 {
		for _, include := range fc.Includes {
			if include == name {
				return true
			}
		}
		return false
	}

	if len(fc.Excludes) > 0 {
		for _, exclude := range fc.Excludes {
			if exclude == name {
				return false
			}
		}
	}

	return true
}

// isIncludePluginMode returns true if the mode is "c2hlbgo"
func isIncludePluginMode(conf *config.Config) bool {
	return conf.Extension.Mode == "c2hlbgo"
}
