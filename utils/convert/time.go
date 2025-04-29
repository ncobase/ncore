package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Supported time formats for parsing
var timeFormats = []string{
	"2006-01",
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006.01",
	"2006.01.02",
	"2006.01.02 15:04:05",
	"2006/01",
	"2006/01/02",
	"2006/01/02 15:04:05",
	"200601",
	"20060102",
	"20060102150405",
	"2006-01-02T15:04:05Z",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
}

const (
	// DefaultLayout24h default 24h layout
	DefaultLayout24h = "yyyy-MM-dd HH:mm:ss"
	// DefaultLayout12h default 12h layout
	DefaultLayout12h = "yyyy-MM-dd hh:mm:ss"
	// Default time layout for formatting
	timeLayout = "2006-01-02 15:04:05"
)

// ParseLocalTime parses string to local time by trying multiple formats
func ParseLocalTime(str string) (t time.Time, err error) {
	location := time.Now().Location()
	for _, format := range timeFormats {
		t, err = time.ParseInLocation(format, str, location)
		if err == nil {
			return
		}
	}
	err = errors.New("cannot parse string as time: " + str)
	return
}

// UnixSecToTime converts unix seconds to time.Time
func UnixSecToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// ToPBTimestamp converts time.Time to protobuf Timestamp
func ToPBTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// PtrToPBTimestamp converts *time.Time to *timestamppb.Timestamp
func PtrToPBTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// AdjustToEndOfDay adjusts the given time to the end of the day (23:59:59)
func AdjustToEndOfDay(value any) (int64, error) {
	var adjustedTime time.Time

	switch v := value.(type) {
	case string:
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		localTime := parsedTime.Local()
		adjustedTime = time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 23, 59, 59, 0, localTime.Location())
	case *time.Time:
		if v != nil {
			localTime := v.Local()
			adjustedTime = time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 23, 59, 59, 0, localTime.Location())
		}
	default:
		return 0, fmt.Errorf("invalid type for time adjustment")
	}

	return adjustedTime.UnixMilli(), nil
}

// FormatTime formats time to string with optional layout
func FormatTime(t *time.Time, layout ...string) *string {
	if t == nil {
		return nil
	}
	l := timeLayout
	if len(layout) > 0 && layout[0] != "" {
		l = layout[0]
	}
	s := t.Format(l)
	return &s
}

// UnixMilliToString converts unix milliseconds timestamp to string
func UnixMilliToString(t *int64, layout ...string) *string {
	if t == nil {
		return nil
	}
	l := timeLayout
	if len(layout) > 0 && layout[0] != "" {
		l = layout[0]
	}
	s := UnixMilliToTime(t).Format(l)
	return &s
}

// UnixMilliToTime converts unix milliseconds timestamp to time.Time
func UnixMilliToTime(i *int64) *time.Time {
	if i == nil {
		return nil
	}
	t := time.UnixMilli(*i)
	return &t
}

// ToUnixMilli converts various types to unix milliseconds timestamp
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
