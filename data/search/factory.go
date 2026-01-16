package search

import (
	"fmt"
)

// AdapterFactory creates search adapters from data connections
type AdapterFactory func(conn any) (Adapter, error)

var (
	// Registry of adapter factories by engine type
	adapterFactories = make(map[Engine]AdapterFactory)
)

// RegisterAdapterFactory registers a factory for creating search adapters
// This is called by search driver packages in their init() functions
func RegisterAdapterFactory(engine Engine, factory AdapterFactory) {
	adapterFactories[engine] = factory
}

// GetAdapterFactory returns the factory for a given engine
func GetAdapterFactory(engine Engine) (AdapterFactory, error) {
	factory, ok := adapterFactories[engine]
	if !ok {
		return nil, fmt.Errorf("no adapter factory registered for engine: %s", engine)
	}
	return factory, nil
}

// GetRegisteredEngines returns list of engines with registered factories
func GetRegisteredEngines() []Engine {
	engines := make([]Engine, 0, len(adapterFactories))
	for engine := range adapterFactories {
		engines = append(engines, engine)
	}
	return engines
}
