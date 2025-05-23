package utils

import "strings"

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

// RemoveDuplicates removes duplicate elements from a.
func RemoveDuplicates(a []string) []string {
	seen := make(map[string]struct{}, len(a))
	j := 0
	for _, v := range a {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		a[j] = v
		j++
	}
	// Create a new slice with the exact length
	result := make([]string, j)
	copy(result, a[:j])
	return result
}

// FirstUpper returns a copy of the string s with its first letter capitalized.
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower returns a copy of the string s with its first letter lowercased.
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
