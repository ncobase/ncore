package manager

import (
	"fmt"
	"sort"

	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// getInitOrder returns the initialization order based on dependencies
//
// noDeps - modules with no dependencies, first to initialize
// withDeps - modules with dependencies
// special - special modules that should be ordered last
func getInitOrder(extensions map[string]*types.Wrapper, dependencyGraph map[string][]string) ([]string, error) {
	var noDeps, withDeps, special []string
	// Exclude these modules from dependency check, adjust as needed
	var specialModules []string
	// Example of how to add modules dynamically
	// specialModules = append(specialModules, "relation", "relations", "linker", "linkers")
	specialSet := make(map[string]bool)
	for _, m := range specialModules {
		specialSet[m] = true
	}

	initialized := make(map[string]bool)

	// If no dependency graph is provided, build one from extension metadata
	if dependencyGraph == nil {
		dependencyGraph = make(map[string][]string)
		for name, ext := range extensions {
			dependencyGraph[name] = ext.Instance.Dependencies()
		}
	}

	// Classify modules based on dependencies
	for name := range extensions {
		if specialSet[name] {
			special = append(special, name)
			continue
		}

		deps := dependencyGraph[name]
		if len(deps) == 0 {
			noDeps = append(noDeps, name)
			initialized[name] = true
		} else {
			withDeps = append(withDeps, name)
		}
	}

	// Sort modules with no dependencies
	sort.Strings(noDeps)

	// Build initialization order
	var order []string
	order = append(order, noDeps...)

	for len(withDeps) > 0 {
		progress := false
		remainingDeps := withDeps[:0]

		for _, name := range withDeps {
			canInitialize := true
			for _, dep := range dependencyGraph[name] {
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
			return nil, fmt.Errorf("cyclic dependency detected in extensions: %v", remainingDeps)
		}

		withDeps = remainingDeps
	}

	// Add special modules
	order = append(order, special...)

	logger.Debugf(nil, "Extension initialization order: %v", order)

	return order, nil
}
