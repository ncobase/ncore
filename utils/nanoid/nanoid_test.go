package nanoid

import (
	"strings"
	"testing"
)

func TestGeneratorsLengthAndCharset(t *testing.T) {
	if got := Must(); len(got) != 16 {
		t.Fatalf("Must() expected default length 16, got %d", len(got))
	}
	if got := Must(8); len(got) != 8 {
		t.Fatalf("Must(8) expected length 8, got %d", len(got))
	}

	letters := String(12)
	if len(letters) != 12 || !isOnlyLetters(letters) {
		t.Fatalf("String(12) expected 12 letters, got %q", letters)
	}

	lower := Lower(10)
	if len(lower) != 10 || strings.ToLower(lower) != lower {
		t.Fatalf("Lower(10) expected lowercase letters, got %q", lower)
	}

	upper := Upper(10)
	if len(upper) != 10 || strings.ToUpper(upper) != upper {
		t.Fatalf("Upper(10) expected uppercase letters, got %q", upper)
	}

	number := Number(14)
	if len(number) != 14 || !isOnlyDigits(number) {
		t.Fatalf("Number(14) expected digits only, got %q", number)
	}
}

func TestPrimaryKeyHelpers(t *testing.T) {
	generate := PrimaryKey()
	id := generate()
	if !IsPrimaryKey(id) {
		t.Fatalf("IsPrimaryKey() expected true for generated key %q", id)
	}
	if IsPrimaryKey(id + "!") {
		t.Fatal("IsPrimaryKey() expected false for invalid characters")
	}
	if IsPrimaryKey("short") {
		t.Fatal("IsPrimaryKey() expected false for invalid length")
	}
}

func TestIsAlphanumeric(t *testing.T) {
	if !IsAlphanumeric("abc123XYZ") {
		t.Fatal("IsAlphanumeric() expected true for alphanumeric input")
	}
	if IsAlphanumeric("abc-123") {
		t.Fatal("IsAlphanumeric() expected false for non-alphanumeric input")
	}
}

func isOnlyDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isOnlyLetters(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}
