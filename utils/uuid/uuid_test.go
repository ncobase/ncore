package uuid

import "testing"

func TestUUIDGeneration(t *testing.T) {
	id := New()
	if id.String() == "" {
		t.Fatal("New() should return a non-empty UUID")
	}

	idStr := NewString()
	if len(idStr) != 36 {
		t.Fatalf("NewString() expected UUID string length 36, got %d", len(idStr))
	}
}

func TestShortUUID(t *testing.T) {
	short := ShortUUID(New())
	if len(short) != 16 {
		t.Fatalf("ShortUUID() expected length 16, got %d", len(short))
	}
	for _, r := range short {
		if !isBase64Alphabet(r) {
			t.Fatalf("ShortUUID() returned unexpected character %q", string(r))
		}
	}
}

func isBase64Alphabet(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	return r == '+' || r == '-'
}
