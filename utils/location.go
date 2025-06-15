package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ncobase/ncore/config"
)

// GetHost constructs the URL based on the given space and config, with an optional port.
func GetHost(conf *config.Config, space string, ports ...int) string {
	port := getPort(conf, ports...)
	return buildURL(conf.Protocol, space, port)
}

// getPort retrieves the port number from the config or the optional ports parameter.
func getPort(conf *config.Config, ports ...int) string {
	if len(ports) > 0 {
		return strconv.Itoa(ports[0])
	} else if conf.Port != 0 {
		return strconv.Itoa(conf.Port)
	}
	return ""
}

// buildURL constructs the URL string based on the protocol, space, and optional port.
func buildURL(protocol, space, port string) string {
	if port != "" {
		return fmt.Sprintf("%v://%v:%v", protocol, space, port)
	}
	return fmt.Sprintf("%v://%v", protocol, space)
}

// ExtractDomain extracts the domain from a URL string.
func ExtractDomain(url string) string {
	// Remove protocol prefix if exists
	for _, prefix := range []string{"http://", "https://"} {
		if len(url) > len(prefix) && url[:len(prefix)] == prefix {
			url = url[len(prefix):]
			break
		}
	}

	// Find domain part (up to first / or : if exists)
	end := len(url)
	for i, c := range url {
		if c == '/' || c == ':' {
			end = i
			break
		}
	}

	return url[:end]
}

// IsValidURL checks if a string is a valid URL.
func IsValidURL(url string) bool {
	// Check if URL starts with valid protocol
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// Check minimum length
	if len(url) < 10 { // http://a.b
		return false
	}

	// Must contain domain after protocol
	domain := ExtractDomain(url)
	return IsValidDomain(domain)
}

// IsValidDomain checks if a string is a valid domain.
func IsValidDomain(domain string) bool {
	// Check if empty
	if domain == "" {
		return false
	}

	// Check length constraints
	if len(domain) > 255 {
		return false
	}

	// Split into labels
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}

	// Validate each label
	for _, label := range labels {
		// Check label length
		if len(label) == 0 || len(label) > 63 {
			return false
		}

		// Check characters
		for _, c := range label {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}

		// Check first and last characters
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
	}

	return true
}
