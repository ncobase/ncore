package utils

import (
	"reflect"
	"testing"
)

func TestFindAndContains(t *testing.T) {
	items := []string{"a", "b", "c"}

	if got := Find(items, "b"); got != 1 {
		t.Fatalf("Find() expected index 1, got %d", got)
	}
	if got := FindID(items, "c"); got != 2 {
		t.Fatalf("FindID() expected index 2, got %d", got)
	}
	if got := Find(items, "x"); got != len(items) {
		t.Fatalf("Find() expected fallback index %d, got %d", len(items), got)
	}
	if !Contains(items, "a") {
		t.Fatal("Contains() expected true for existing value")
	}
	if Contains(items, "x") {
		t.Fatal("Contains() expected false for non-existing value")
	}
}

func TestRemoveDuplicates(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	got := RemoveDuplicates(input)
	want := []string{"a", "b", "c", "d"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RemoveDuplicates() expected %v, got %v", want, got)
	}
}

func TestFirstUpperFirstLower(t *testing.T) {
	if got := FirstUpper("hello"); got != "Hello" {
		t.Fatalf("FirstUpper() expected Hello, got %q", got)
	}
	if got := FirstLower("Hello"); got != "hello" {
		t.Fatalf("FirstLower() expected hello, got %q", got)
	}
	if got := FirstUpper(""); got != "" {
		t.Fatalf("FirstUpper() expected empty string, got %q", got)
	}
	if got := FirstLower(""); got != "" {
		t.Fatalf("FirstLower() expected empty string, got %q", got)
	}
}
