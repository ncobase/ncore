package convert

import (
	"reflect"
	"strings"
	"testing"
)

func TestConvertValue(t *testing.T) {
	got, err := ConvertValue("string", "value")
	if err != nil || got != "value" {
		t.Fatalf("unexpected string conversion result: value=%#v err=%v", got, err)
	}

	got, err = ConvertValue("number", "2")
	if err != nil || got != int64(2) {
		t.Fatalf("unexpected integer number conversion result: value=%#v err=%v", got, err)
	}

	got, err = ConvertValue("number", "2.5")
	if err != nil || got != 2.5 {
		t.Fatalf("unexpected float number conversion result: value=%#v err=%v", got, err)
	}

	got, err = ConvertValue("boolean", "yes")
	if err != nil || got != true {
		t.Fatalf("unexpected boolean conversion result: value=%#v err=%v", got, err)
	}

	got, err = ConvertValue("object", `{"a":1}`)
	if err != nil {
		t.Fatalf("unexpected object conversion error: %v", err)
	}
	if _, ok := got.(map[string]any); !ok {
		t.Fatalf("unexpected object conversion type: %T", got)
	}

	got, err = ConvertValue("array", `[1,2]`)
	if err != nil {
		t.Fatalf("unexpected array conversion error: %v", err)
	}
	if _, ok := got.([]any); !ok {
		t.Fatalf("unexpected array conversion type: %T", got)
	}

	if _, err := ConvertValue("number", "not-a-number"); err == nil {
		t.Fatal("expected invalid number conversion error")
	}
	if _, err := ConvertValue("boolean", "maybe"); err == nil {
		t.Fatal("expected invalid boolean conversion error")
	}
	if _, err := ConvertValue("object", "{invalid"); err == nil {
		t.Fatal("expected invalid object conversion error")
	}
	if _, err := ConvertValue("array", "[invalid"); err == nil {
		t.Fatal("expected invalid array conversion error")
	}
	if _, err := ConvertValue("unknown", "x"); err == nil {
		t.Fatal("expected unsupported type conversion error")
	}
}

func TestToString(t *testing.T) {
	if got := ToString(nil); got != "" {
		t.Fatalf("unexpected nil ToString result: %q", got)
	}
	if got := ToString(true); got != "true" {
		t.Fatalf("unexpected bool ToString result: %q", got)
	}
	if got := ToString(12); got != "12" {
		t.Fatalf("unexpected int ToString result: %q", got)
	}
	if got := ToString(int64(13)); got != "13" {
		t.Fatalf("unexpected int64 ToString result: %q", got)
	}
	if got := ToString(1.5); got != "1.5" {
		t.Fatalf("unexpected float ToString result: %q", got)
	}
	if got := ToString(map[string]any{"a": 1}); !strings.Contains(got, `"a":1`) {
		t.Fatalf("unexpected map ToString result: %q", got)
	}
	if got := ToString([]any{"a", 1}); !strings.Contains(got, `"a"`) {
		t.Fatalf("unexpected array ToString result: %q", got)
	}

	custom := struct {
		Name string
	}{Name: "alice"}
	if got := ToString(custom); !strings.Contains(got, "alice") {
		t.Fatalf("unexpected custom ToString result: %q", got)
	}
}

func TestToInt(t *testing.T) {
	cases := []struct {
		input any
		want  int64
	}{
		{1, 1},
		{int64(2), 2},
		{3.9, 3},
		{"4", 4},
		{true, 1},
		{false, 0},
	}

	for _, c := range cases {
		got, err := ToInt(c.input)
		if err != nil || got != c.want {
			t.Fatalf("unexpected ToInt result for %T: value=%d err=%v", c.input, got, err)
		}
	}

	if _, err := ToInt("not-int"); err == nil {
		t.Fatal("expected string parsing error for ToInt")
	}
	if _, err := ToInt(struct{}{}); err == nil {
		t.Fatal("expected unsupported type error for ToInt")
	}
}

func TestToFloat(t *testing.T) {
	cases := []struct {
		input any
		want  float64
	}{
		{1.5, 1.5},
		{1, 1},
		{int64(2), 2},
		{"3.5", 3.5},
		{true, 1},
		{false, 0},
	}

	for _, c := range cases {
		got, err := ToFloat(c.input)
		if err != nil || got != c.want {
			t.Fatalf("unexpected ToFloat result for %T: value=%f err=%v", c.input, got, err)
		}
	}

	if _, err := ToFloat("not-float"); err == nil {
		t.Fatal("expected string parsing error for ToFloat")
	}
	if _, err := ToFloat(struct{}{}); err == nil {
		t.Fatal("expected unsupported type error for ToFloat")
	}
}

func TestToBool(t *testing.T) {
	cases := []struct {
		input any
		want  bool
	}{
		{true, true},
		{1, true},
		{0, false},
		{int64(2), true},
		{float64(0), false},
		{"yes", true},
		{"off", false},
		{"", false},
	}

	for _, c := range cases {
		got, err := ToBool(c.input)
		if err != nil || got != c.want {
			t.Fatalf("unexpected ToBool result for %T: value=%v err=%v", c.input, got, err)
		}
	}

	if _, err := ToBool("maybe"); err == nil {
		t.Fatal("expected invalid boolean value error for ToBool")
	}
	if _, err := ToBool(struct{}{}); err == nil {
		t.Fatal("expected unsupported type error for ToBool")
	}
}

func TestToObjectAndToArray(t *testing.T) {
	sourceObj := map[string]any{"a": 1}
	gotObj, err := ToObject(sourceObj)
	if err != nil || !reflect.DeepEqual(gotObj, sourceObj) {
		t.Fatalf("unexpected ToObject map result: value=%#v err=%v", gotObj, err)
	}

	gotObj, err = ToObject(`{"a":1}`)
	if err != nil || gotObj["a"] != float64(1) {
		t.Fatalf("unexpected ToObject string result: value=%#v err=%v", gotObj, err)
	}

	if _, err := ToObject("{invalid"); err == nil {
		t.Fatal("expected invalid JSON object error")
	}
	if _, err := ToObject([]int{1, 2}); err == nil {
		t.Fatal("expected unsupported ToObject type error")
	}

	sourceArr := []any{"a", 1}
	gotArr, err := ToArray(sourceArr)
	if err != nil || !reflect.DeepEqual(gotArr, sourceArr) {
		t.Fatalf("unexpected ToArray array result: value=%#v err=%v", gotArr, err)
	}

	gotArr, err = ToArray(`[1,"a"]`)
	if err != nil || len(gotArr) != 2 {
		t.Fatalf("unexpected ToArray string result: value=%#v err=%v", gotArr, err)
	}

	if _, err := ToArray("[invalid"); err == nil {
		t.Fatal("expected invalid JSON array error")
	}
	if _, err := ToArray(map[string]int{"a": 1}); err == nil {
		t.Fatal("expected unsupported ToArray type error")
	}
}

func TestTypeValidationAndStringArrayConversions(t *testing.T) {
	if !IsValidType("string") || !IsValidType("array") {
		t.Fatal("expected supported type validation to return true")
	}
	if IsValidType("custom") {
		t.Fatal("expected unsupported type validation to return false")
	}

	if ToStringArray(nil) != nil {
		t.Fatal("expected nil input to return nil output")
	}

	strs := []string{"a", "b"}
	gotStrs := ToStringArray(strs)
	if !reflect.DeepEqual(gotStrs, strs) {
		t.Fatalf("unexpected []string conversion: %#v", gotStrs)
	}
	strs[0] = "changed"
	if gotStrs[0] != "a" {
		t.Fatal("expected []string conversion to return a copy")
	}

	if !reflect.DeepEqual(ToStringArray([]int{1, 2}), []string{"1", "2"}) {
		t.Fatal("unexpected []int conversion")
	}
	if !reflect.DeepEqual(ToStringArray([]int64{1, 2}), []string{"1", "2"}) {
		t.Fatal("unexpected []int64 conversion")
	}
	if !reflect.DeepEqual(ToStringArray([]float64{1.5, 2}), []string{"1.5", "2"}) {
		t.Fatal("unexpected []float64 conversion")
	}
	if !reflect.DeepEqual(ToStringArray([]bool{true, false}), []string{"true", "false"}) {
		t.Fatal("unexpected []bool conversion")
	}
	if !reflect.DeepEqual(ToStringArray([]any{1, "a"}), []string{"1", "a"}) {
		t.Fatal("unexpected []any conversion")
	}
	if len(ToStringArray([]chan int{make(chan int)})) != 0 {
		t.Fatal("expected unsupported array conversion fallback to empty slice")
	}
}

func TestFromStringArray(t *testing.T) {
	result, err := FromStringArray(nil, "string")
	if err != nil || result != nil {
		t.Fatalf("unexpected nil conversion result: value=%#v err=%v", result, err)
	}

	result, err = FromStringArray([]string{"a", "b"}, "string")
	if err != nil || !reflect.DeepEqual(result, []string{"a", "b"}) {
		t.Fatalf("unexpected string array conversion result: value=%#v err=%v", result, err)
	}

	result, err = FromStringArray([]string{"1", "2"}, "int")
	if err != nil || !reflect.DeepEqual(result, []int{1, 2}) {
		t.Fatalf("unexpected int array conversion result: value=%#v err=%v", result, err)
	}
	if _, err := FromStringArray([]string{"x"}, "int"); err == nil {
		t.Fatal("expected int array conversion error")
	}

	result, err = FromStringArray([]string{"1", "2"}, "int64")
	if err != nil || !reflect.DeepEqual(result, []int64{1, 2}) {
		t.Fatalf("unexpected int64 array conversion result: value=%#v err=%v", result, err)
	}
	if _, err := FromStringArray([]string{"x"}, "int64"); err == nil {
		t.Fatal("expected int64 array conversion error")
	}

	result, err = FromStringArray([]string{"1.5", "2"}, "float64")
	if err != nil || !reflect.DeepEqual(result, []float64{1.5, 2}) {
		t.Fatalf("unexpected float64 array conversion result: value=%#v err=%v", result, err)
	}
	if _, err := FromStringArray([]string{"x"}, "float64"); err == nil {
		t.Fatal("expected float64 array conversion error")
	}

	result, err = FromStringArray([]string{"true", "0"}, "bool")
	if err != nil || !reflect.DeepEqual(result, []bool{true, false}) {
		t.Fatalf("unexpected bool array conversion result: value=%#v err=%v", result, err)
	}
	if _, err := FromStringArray([]string{"maybe"}, "bool"); err == nil {
		t.Fatal("expected bool array conversion error")
	}

	if _, err := FromStringArray([]string{"1"}, "unknown"); err == nil {
		t.Fatal("expected unsupported target type error")
	}
}

func TestToTypedArray(t *testing.T) {
	stringValues, err := ToTypedArray[string]([]string{"a", "b"})
	if err != nil || !reflect.DeepEqual(stringValues, []string{"a", "b"}) {
		t.Fatalf("unexpected typed string conversion: value=%#v err=%v", stringValues, err)
	}

	intValues, err := ToTypedArray[int]([]string{"1", "2"})
	if err != nil || !reflect.DeepEqual(intValues, []int{1, 2}) {
		t.Fatalf("unexpected typed int conversion: value=%#v err=%v", intValues, err)
	}
	if _, err := ToTypedArray[int]([]string{"x"}); err == nil {
		t.Fatal("expected typed int conversion error")
	}

	int64Values, err := ToTypedArray[int64]([]string{"1", "2"})
	if err != nil || !reflect.DeepEqual(int64Values, []int64{1, 2}) {
		t.Fatalf("unexpected typed int64 conversion: value=%#v err=%v", int64Values, err)
	}
	if _, err := ToTypedArray[int64]([]string{"x"}); err == nil {
		t.Fatal("expected typed int64 conversion error")
	}

	floatValues, err := ToTypedArray[float64]([]string{"1.5", "2"})
	if err != nil || !reflect.DeepEqual(floatValues, []float64{1.5, 2}) {
		t.Fatalf("unexpected typed float64 conversion: value=%#v err=%v", floatValues, err)
	}
	if _, err := ToTypedArray[float64]([]string{"x"}); err == nil {
		t.Fatal("expected typed float64 conversion error")
	}

	boolValues, err := ToTypedArray[bool]([]string{"true", "0"})
	if err != nil || !reflect.DeepEqual(boolValues, []bool{true, false}) {
		t.Fatalf("unexpected typed bool conversion: value=%#v err=%v", boolValues, err)
	}
	if _, err := ToTypedArray[bool]([]string{"maybe"}); err == nil {
		t.Fatal("expected typed bool conversion error")
	}

	if _, err := ToTypedArray[uint]([]string{"1"}); err == nil {
		t.Fatal("expected unsupported typed array conversion error")
	}
}
