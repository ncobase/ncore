package validator

import "net/url"

// IsURL checks if a string is a valid URL.
func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Host != ""
}
