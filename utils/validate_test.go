package utils

import "testing"

func TestValidateName(t *testing.T) {
	valid := []string{"sample", "sample_1", "SampleName", "sample-name"}
	for _, name := range valid {
		if !ValidateName(name) {
			t.Fatalf("ValidateName() expected true for %q", name)
		}
	}

	invalid := []string{"", "1sample", "sample name", "sample.name", "@sample"}
	for _, name := range invalid {
		if ValidateName(name) {
			t.Fatalf("ValidateName() expected false for %q", name)
		}
	}
}
