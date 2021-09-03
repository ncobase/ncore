package util

import (
	"fmt"
	"reflect"
	"stone/common/conf"
)

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func FindID(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// GetDomain Get the run domain
func GetDomain(domain string) string {
	if domain == "localhost" {
		return fmt.Sprintf("%v://%v:%d", conf.G.Protocol, conf.G.Domain, conf.G.Port)
	}
	return fmt.Sprintf("%v://%v", conf.G.Protocol, conf.G.Domain)
}

// IsNil verify is nil
func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
