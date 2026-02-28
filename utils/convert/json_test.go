package convert

import (
	"reflect"
	"strings"
	"testing"
)

type jsonSample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestToJSONAndFromJSON(t *testing.T) {
	empty, err := ToJSON(nil)
	if err != nil || empty != "" {
		t.Fatalf("unexpected nil JSON result: value=%q err=%v", empty, err)
	}

	source := jsonSample{Name: "alice", Age: 18}
	raw, err := ToJSON(source)
	if err != nil {
		t.Fatalf("ToJSON returned unexpected error: %v", err)
	}

	var target jsonSample
	if err := FromJSON(raw, &target); err != nil {
		t.Fatalf("FromJSON returned unexpected error: %v", err)
	}
	if target != source {
		t.Fatalf("unexpected FromJSON result: %#v", target)
	}

	if err := FromJSON("", &target); err != nil {
		t.Fatalf("expected empty JSON input to be ignored, got error: %v", err)
	}
	if err := FromJSON("{invalid", &target); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestMustJSONHelpers(t *testing.T) {
	if got := MustToJSON(map[string]int{"a": 1}); got == "" {
		t.Fatal("MustToJSON returned empty string for valid input")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected MustToJSON panic for unsupported value")
		}
	}()
	MustToJSON(func() {})
}

func TestMustFromJSONPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected MustFromJSON panic for invalid input")
		}
	}()
	MustFromJSON("{invalid", &jsonSample{})
}

func TestToJSONMapAndFromJSONMap(t *testing.T) {
	result, err := ToJSONMap(nil)
	if err != nil || result != nil {
		t.Fatalf("unexpected nil map conversion result: value=%#v err=%v", result, err)
	}

	original := map[string]any{"name": "alice"}
	mapped, err := ToJSONMap(original)
	if err != nil {
		t.Fatalf("ToJSONMap returned unexpected error: %v", err)
	}
	mapped["name"] = "bob"
	if original["name"] != "bob" {
		t.Fatal("expected direct map conversion to return same map instance")
	}

	sampleMap, err := ToJSONMap(jsonSample{Name: "alice", Age: 20})
	if err != nil {
		t.Fatalf("ToJSONMap struct conversion failed: %v", err)
	}
	if sampleMap["name"] != "alice" {
		t.Fatalf("unexpected struct conversion result: %#v", sampleMap)
	}

	if _, err := ToJSONMap(func() {}); err == nil {
		t.Fatal("expected ToJSONMap error for unsupported value")
	}

	var target jsonSample
	if err := FromJSONMap(nil, &target); err != nil {
		t.Fatalf("expected nil JSON map to be ignored, got error: %v", err)
	}
	if err := FromJSONMap(map[string]any{"name": "alice", "age": 21}, &target); err != nil {
		t.Fatalf("FromJSONMap returned unexpected error: %v", err)
	}
	if target.Name != "alice" || target.Age != 21 {
		t.Fatalf("unexpected FromJSONMap result: %#v", target)
	}
	if err := FromJSONMap(map[string]any{"bad": make(chan int)}, &target); err == nil {
		t.Fatal("expected FromJSONMap error for unsupported map value")
	}
}

func TestPrettyJSONAndStringHelpers(t *testing.T) {
	value, err := PrettyJSON(nil)
	if err != nil || value != "" {
		t.Fatalf("unexpected pretty JSON nil result: value=%q err=%v", value, err)
	}

	pretty, err := PrettyJSON(map[string]string{"name": "alice"})
	if err != nil {
		t.Fatalf("PrettyJSON returned unexpected error: %v", err)
	}
	if !strings.Contains(pretty, "\n") {
		t.Fatalf("expected indented output, got %q", pretty)
	}
	if _, err := PrettyJSON(func() {}); err == nil {
		t.Fatal("expected PrettyJSON error for unsupported value")
	}

	if got := StringifyJSON(map[string]int{"a": 1}); got == "" {
		t.Fatal("expected non-empty JSON string")
	}
	if got := StringifyJSON(func() {}); got != "" {
		t.Fatalf("expected empty string on stringify error, got %q", got)
	}
}

func TestParseJSONAndJSONUnmarshal(t *testing.T) {
	target := map[string]any{}
	if !ParseJSON(`{"a":1}`, &target) {
		t.Fatal("expected ParseJSON to succeed for valid JSON")
	}
	if !reflect.DeepEqual(target["a"], float64(1)) {
		t.Fatalf("unexpected parsed JSON value: %#v", target)
	}

	if ParseJSON("{invalid", &target) {
		t.Fatal("expected ParseJSON to fail for invalid JSON")
	}
	if !ParseJSON("", &target) {
		t.Fatal("expected empty ParseJSON input to return true")
	}

	target = map[string]any{}
	if !JSONUnmarshal(`{"b":2}`, &target) {
		t.Fatal("expected JSONUnmarshal to succeed for valid JSON")
	}
	if JSONUnmarshal("{invalid", &target) {
		t.Fatal("expected JSONUnmarshal to fail for invalid JSON")
	}
}
