package convert

import (
	"math"
	"strconv"
	"testing"
)

func TestNumericConversions(t *testing.T) {
	gotInt, err := Int64ToInt(42)
	if err != nil {
		t.Fatalf("Int64ToInt returned unexpected error: %v", err)
	}
	if gotInt != 42 {
		t.Fatalf("unexpected int conversion result: %d", gotInt)
	}

	if strconv.IntSize == 32 {
		if _, err := Int64ToInt(math.MaxInt64); err == nil {
			t.Fatal("expected overflow error on 32-bit platform")
		}
	}

	if IntToInt64(7) != int64(7) {
		t.Fatal("IntToInt64 returned unexpected value")
	}
	if IntToString(7) != "7" {
		t.Fatal("IntToString returned unexpected value")
	}
	if Int64ToString(9) != "9" {
		t.Fatal("Int64ToString returned unexpected value")
	}
	if Float64ToString(1.25) != "1.25" {
		t.Fatal("Float64ToString returned unexpected value")
	}

	parsedInt, err := StringToInt("11")
	if err != nil || parsedInt != 11 {
		t.Fatalf("StringToInt failed: value=%d err=%v", parsedInt, err)
	}
	if _, err := StringToInt("not-int"); err == nil {
		t.Fatal("expected StringToInt error for invalid input")
	}

	parsedInt64, err := StringToInt64("12")
	if err != nil || parsedInt64 != 12 {
		t.Fatalf("StringToInt64 failed: value=%d err=%v", parsedInt64, err)
	}
	if _, err := StringToInt64("not-int64"); err == nil {
		t.Fatal("expected StringToInt64 error for invalid input")
	}

	parsedFloat, err := StringToFloat64("3.14")
	if err != nil || parsedFloat != 3.14 {
		t.Fatalf("StringToFloat64 failed: value=%f err=%v", parsedFloat, err)
	}
	if _, err := StringToFloat64("not-float"); err == nil {
		t.Fatal("expected StringToFloat64 error for invalid input")
	}

	parsedInt32, err := StringToInt32("13")
	if err != nil || parsedInt32 != int32(13) {
		t.Fatalf("StringToInt32 failed: value=%d err=%v", parsedInt32, err)
	}
	if _, err := StringToInt32("not-int32"); err == nil {
		t.Fatal("expected StringToInt32 error for invalid input")
	}
}

func TestMinMax(t *testing.T) {
	if Max(3, 2) != 3 {
		t.Fatal("Max returned unexpected result")
	}
	if Min(3, 2) != 2 {
		t.Fatal("Min returned unexpected result")
	}
}
