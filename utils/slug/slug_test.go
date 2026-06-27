package slug

import "testing"

func TestUnicode(t *testing.T) {
	got := Unicode("Hello 世界")
	if got == "" {
		t.Fatal("Unicode() should not return empty slug")
	}
	if got != "hello-shi-jie" {
		t.Fatalf("Unicode() expected %q, got %q", "hello-shi-jie", got)
	}
}
