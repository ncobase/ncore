package utils

import (
	"testing"

	"github.com/ncobase/ncore/config"
)

func TestGetHost(t *testing.T) {
	conf := &config.Config{
		Protocol: "https",
		Port:     8443,
	}

	if got := GetHost(conf, "example.com"); got != "https://example.com:8443" {
		t.Fatalf("GetHost() expected https://example.com:8443, got %q", got)
	}
	if got := GetHost(conf, "example.com", 9443); got != "https://example.com:9443" {
		t.Fatalf("GetHost() with override port expected https://example.com:9443, got %q", got)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := map[string]string{
		"https://example.com/path": "example.com",
		"http://example.com:8080":  "example.com",
		"example.com/test":         "example.com",
	}

	for input, want := range tests {
		if got := ExtractDomain(input); got != want {
			t.Fatalf("ExtractDomain(%q) expected %q, got %q", input, want, got)
		}
	}
}

func TestURLAndDomainValidation(t *testing.T) {
	if !IsValidURL("https://example.com/path") {
		t.Fatal("IsValidURL() expected true for valid https URL")
	}
	if IsValidURL("ftp://example.com") {
		t.Fatal("IsValidURL() expected false for unsupported scheme")
	}

	if !IsValidDomain("sub.example.com") {
		t.Fatal("IsValidDomain() expected true for valid domain")
	}
	if IsValidDomain("-invalid.example.com") {
		t.Fatal("IsValidDomain() expected false for invalid domain label")
	}
}
