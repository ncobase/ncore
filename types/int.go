package types

import (
	"fmt"
	"math"
	"strconv"
)

// Int64ToInt converts an int64 to an int
func Int64ToInt(n int64) (int, error) {
	// Check for potential overflow
	if n > math.MaxInt || n < math.MinInt {
		return 0, fmt.Errorf("int64 value %d out of int range", n)
	}
	return int(n), nil
}

// IntToInt64 converts an int to an int64
func IntToInt64(n int) int64 {
	return int64(n)
}

// IntToString converts an int to a string
func IntToString(n int) string {
	return fmt.Sprintf("%d", n)
}

// Int64ToString converts an int64 to a string
func Int64ToString(n int64) string {
	return fmt.Sprintf("%d", n)
}

// StringToInt converts a string to an int
func StringToInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// StringToInt64 converts a string to an int64
func StringToInt64(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// StringToInt32 converts a string to an int32
func StringToInt32(s string) (int32, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}
