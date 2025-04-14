package utils

import (
	"regexp"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// ValidateName validates a name
func ValidateName(name string) bool {
	return nameRegex.MatchString(name)
}
