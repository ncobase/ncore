package convert

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseLocalTime(t *testing.T) {
	parsed, err := ParseLocalTime("2026-02-28")
	if err != nil {
		t.Fatalf("ParseLocalTime returned unexpected error: %v", err)
	}
	if parsed.Year() != 2026 || parsed.Month() != time.February || parsed.Day() != 28 {
		t.Fatalf("unexpected parsed date: %v", parsed)
	}

	if _, err := ParseLocalTime("not-time"); err == nil {
		t.Fatal("expected parse error for invalid time value")
	}
}

func TestUnixAndProtobufTimeConversions(t *testing.T) {
	base := UnixSecToTime(100)
	if base.Unix() != 100 {
		t.Fatalf("unexpected unix second conversion: %v", base)
	}

	pb := ToPBTimestamp(base)
	if pb.GetSeconds() != 100 {
		t.Fatalf("unexpected protobuf timestamp seconds: %d", pb.GetSeconds())
	}

	if PtrToPBTimestamp(nil) != nil {
		t.Fatal("expected nil pointer conversion to return nil")
	}
	ptr := PtrToPBTimestamp(&base)
	if ptr == nil || ptr.GetSeconds() != 100 {
		t.Fatalf("unexpected protobuf pointer conversion result: %#v", ptr)
	}
}

func TestAdjustToEndOfDay(t *testing.T) {
	input := "2026-02-28T09:10:11Z"
	millis, err := AdjustToEndOfDay(input)
	if err != nil {
		t.Fatalf("AdjustToEndOfDay string conversion failed: %v", err)
	}

	parsedInput, _ := time.Parse(time.RFC3339, input)
	inputLocal := parsedInput.Local()
	got := time.UnixMilli(millis).Local()
	if got.Year() != inputLocal.Year() || got.Month() != inputLocal.Month() || got.Day() != inputLocal.Day() {
		t.Fatalf("unexpected adjusted date: %v", got)
	}
	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Fatalf("expected end-of-day time, got %v", got)
	}

	local := time.Date(2026, 2, 28, 8, 0, 0, 0, time.Now().Location())
	millis, err = AdjustToEndOfDay(&local)
	if err != nil {
		t.Fatalf("AdjustToEndOfDay pointer conversion failed: %v", err)
	}
	got = time.UnixMilli(millis).Local()
	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Fatalf("expected pointer conversion end-of-day time, got %v", got)
	}

	if _, err := AdjustToEndOfDay(123); err == nil {
		t.Fatal("expected invalid type error")
	}
	if _, err := AdjustToEndOfDay("not-rfc3339"); err == nil {
		t.Fatal("expected invalid RFC3339 parse error")
	}
}

func TestFormatAndUnixMilliHelpers(t *testing.T) {
	if FormatTime(nil) != nil {
		t.Fatal("expected FormatTime nil input to return nil")
	}

	local := time.Date(2026, 2, 28, 8, 9, 10, 0, time.Now().Location())
	formatted := FormatTime(&local)
	if formatted == nil || *formatted != "2026-02-28 08:09:10" {
		t.Fatalf("unexpected default format result: %#v", formatted)
	}

	custom := FormatTime(&local, "2006/01/02")
	if custom == nil || *custom != "2026/02/28" {
		t.Fatalf("unexpected custom format result: %#v", custom)
	}

	if UnixMilliToString(nil) != nil {
		t.Fatal("expected UnixMilliToString nil input to return nil")
	}

	millis := local.UnixMilli()
	milliFormatted := UnixMilliToString(&millis)
	if milliFormatted == nil || *milliFormatted != "2026-02-28 08:09:10" {
		t.Fatalf("unexpected unix-milli default format result: %#v", milliFormatted)
	}

	milliCustom := UnixMilliToString(&millis, "2006/01/02")
	if milliCustom == nil || *milliCustom != "2026/02/28" {
		t.Fatalf("unexpected unix-milli custom format result: %#v", milliCustom)
	}

	if UnixMilliToTime(nil) != nil {
		t.Fatal("expected UnixMilliToTime nil input to return nil")
	}
	convertedTime := UnixMilliToTime(&millis)
	if convertedTime == nil || convertedTime.UnixMilli() != millis {
		t.Fatalf("unexpected UnixMilliToTime result: %#v", convertedTime)
	}
}

func TestToUnixMilliAndParseTimePtr(t *testing.T) {
	if got := ToUnixMilli(float64(123)); got != 123 {
		t.Fatalf("unexpected float conversion result: %d", got)
	}

	if got := ToUnixMilli(json.Number("456")); got != 456 {
		t.Fatalf("unexpected json.Number conversion result: %d", got)
	}
	if got := ToUnixMilli(json.Number("invalid")); got != 0 {
		t.Fatalf("expected invalid json.Number conversion to return 0, got %d", got)
	}
	if got := ToUnixMilli("invalid"); got != 0 {
		t.Fatalf("expected unsupported conversion to return 0, got %d", got)
	}

	parsed := ParseTimePtr("2026-02-28")
	if parsed == nil || parsed.Year() != 2026 {
		t.Fatalf("unexpected ParseTimePtr result: %#v", parsed)
	}
	if ParseTimePtr("invalid-time") != nil {
		t.Fatal("expected invalid ParseTimePtr input to return nil")
	}
}
