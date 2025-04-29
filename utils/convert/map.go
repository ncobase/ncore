package convert

import (
	"sync"
)

// ToSyncMap takes a regular map[string]T and returns a pointer to a sync.Map
// The function copies all the key-value pairs from the regular map into the sync.Map
func ToSyncMap[T any](m map[string]T) *sync.Map {
	sm := &sync.Map{}
	for key, value := range m {
		sm.Store(key, value)
	}
	return sm
}

// FromSyncMap takes a sync.Map and converts it into a regular map[string]T
// The function iterates over the sync.Map and retrieves each key-value pair,
// returning them as a standard Go map
func FromSyncMap[T any](sm *sync.Map) map[string]T {
	m := make(map[string]T)
	sm.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(T); ok {
				m[k] = v
			}
		}
		return true
	})
	return m
}
