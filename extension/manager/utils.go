package manager

import (
	"fmt"
	"sort"

	"github.com/ncobase/ncore/extension/types"
)

// getInitOrder returns the initialization order based on dependencies
func getInitOrder(extensions map[string]*types.Wrapper, dependencyGraph map[string][]string) ([]string, error) {
	var noDeps, withDeps []string
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

	return order, nil
}
