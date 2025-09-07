package types

import (
	"errors"
	"sort"
)

// Order represents sorting direction.
type Order string

const (
	Ascending  Order = "asc"  // Ascending order
	Descending Order = "desc" // Descending order
)

// Criterion represents a single sorting criterion.
type Criterion struct {
	Field string `json:"field"` // Field to sort by
	Order Order  `json:"order"` // Sort direction
}

// MultiCriteria supports multi-field sorting.
type MultiCriteria struct {
	Criteria []Criterion `json:"criteria"` // List of sorting criteria
}

// Sortable represents a sortable dataset interface.
type Sortable interface {
	Sort(criteria MultiCriteria) error // Sort method to be implemented
}

// DynamicSorter provides a generic implementation for sorting based on criteria.
type DynamicSorter struct {
	Data   []map[string]any                                     // Dataset to be sorted
	Getter func(item map[string]any, field string) (any, error) // Field value getter
}

// Sort sorts the dataset based on the given MultiCriteria.
func (ds *DynamicSorter) Sort(criteria MultiCriteria) error {
	if ds.Getter == nil {
		return errors.New("getter function is not defined")
	}

	// Sort the data based on the criteria
	sort.SliceStable(ds.Data, func(i, j int) bool {
		for _, c := range criteria.Criteria {
			val1, err1 := ds.Getter(ds.Data[i], c.Field)
			val2, err2 := ds.Getter(ds.Data[j], c.Field)
			if err1 != nil || err2 != nil {
				continue // Skip this field if there's an error
			}

			// Compare values based on the order
			comparison := CompareValues(val1, val2)
			if c.Order == Descending {
				comparison = -comparison
			}

			if comparison != 0 {
				return comparison < 0
			}
		}
		return false // Default to equal
	})

	return nil
}

// CompareValues compares two values and returns -1, 0, or 1.
// Supports int and string types.
// Returns -1 if a < b.
// Returns 1 if a > b.
// Returns 0 if a == b or types are not comparable.
func CompareValues(a, b any) int {
	switch aVal := a.(type) {
	case int:
		bVal, ok := b.(int)
		if ok {
			return CompareInt(aVal, bVal)
		}
	case string:
		bVal, ok := b.(string)
		if ok {
			return CompareString(aVal, bVal)
		}
	}
	// Consider them equal if types don't match or are unsupported
	return 0
}

// CompareInt compares two integers.
// Returns -1 if a < b.
// Returns 1 if a > b.
// Returns 0 if a == b.
func CompareInt(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// CompareString compares two strings lexicographically.
// Returns -1 if a < b.
// Returns 1 if a > b.
// Returns 0 if a == b.
func CompareString(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}
