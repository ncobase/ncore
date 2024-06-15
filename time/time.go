package time

import (
	"encoding/json"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

// Format format time to string
func Format(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(timeLayout)
	return &s
}

// UnixMilliToString timestamp to string
func UnixMilliToString(t *int64) *string {
	if t == nil {
		return nil
	}
	s := UnixMilliToTime(t).Format(timeLayout)
	return &s
}

// UnixMilliToTime timestamp to time.Time
func UnixMilliToTime(i *int64) *time.Time {
	if i == nil {
		return nil
	}
	t := time.UnixMilli(*i)
	return &t
}

func ToUnixMilli(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	default:
		return 0
	}
}
