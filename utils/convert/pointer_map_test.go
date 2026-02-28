package convert

import (
	"reflect"
	"sync"
	"testing"
)

func TestPointerHelpers(t *testing.T) {
	ptr := ToPointer("value")
	if ptr == nil || *ptr != "value" {
		t.Fatalf("ToPointer returned unexpected value: %#v", ptr)
	}

	value := ToValue(ptr)
	if value != "value" {
		t.Fatalf("ToValue returned unexpected value: %q", value)
	}
}

func TestMapHelpers(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
	}

	sm := ToSyncMap(input)
	if sm == nil {
		t.Fatal("ToSyncMap returned nil")
	}

	got := FromSyncMap[int](sm)
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("unexpected map conversion result: %#v", got)
	}
}

func TestFromSyncMap_IgnoresUnexpectedTypes(t *testing.T) {
	sm := &sync.Map{}
	sm.Store("ok", 1)
	sm.Store(2, 2)
	sm.Store("bad", "not-int")

	got := FromSyncMap[int](sm)
	expected := map[string]int{"ok": 1}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected conversion result: got=%#v want=%#v", got, expected)
	}
}
