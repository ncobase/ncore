package manager

import (
	"fmt"
	"github.com/ncobase/ncore/ext/core"
	"sort"
)

// getInitOrder returns the initialization order based on dependencies
//
// noDeps - modules with no dependencies, first to initialize
// withDeps - modules with dependencies
// special - special modules that should be ordered last
func getInitOrder(extensions map[string]*core.Wrapper) ([]string, error) {
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
	for name, ext := range extensions {
		if specialSet[name] {
			special = append(special, name)
			continue
		}

		deps := make(map[string]bool)

		// Check if all dependencies exist
		for _, dep := range ext.Metadata.Dependencies {
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
