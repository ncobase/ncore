package types

import (
	"errors"
	"strings"
	"time"

	"github.com/ncobase/common/validator"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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
)

// ParseLocalTime parse to local time
func ParseLocalTime(str string) (t time.Time, err error) {
	location := time.Now().Location()
	for _, format := range timeFormats {
		t, err = time.ParseInLocation(format, str, location)
		if err == nil {
			return
		}
	}
	err = errors.New("Can't parse string as time: " + str)
	return
}

// FormatTime format time
func FormatTime(t time.Time, layout string) string {
	if validator.IsEmpty(layout) {
		return ""
	}
	patterns := []string{
		// 年
		"yyyy", "2006", // 完整表示的年份
		"yy", "06", // 2 位数字表示的年份
		// 月
		"MM", "01", // 数字表示的月份，有前导零
		// 日
		"dd", "02", // 月份中的第几天，有前导零的 2 位数字
		// 小时
		"hh", "03", // 12 小时格式，有前导零
		"HH", "15", // 24 小时格式，有前导零
		// 分钟
		"mm", "04", // 有前导零的分钟数
		// 秒
		"ss", "05", // 秒数，有前导零
	}
	replacer := strings.NewReplacer(patterns...)
	layout = replacer.Replace(layout)
	return t.Format(layout)
}

// UnixSecToTime unix sec to time
func UnixSecToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// ToPBTimestamp convert time.Time to pb.Timestamp
func ToPBTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// PtrToPBTimestamp convert *time.Time to *timestamppb.Timestamp
func PtrToPBTimestamp(t *time.Time) *timestamppb.Timestamp {
	if validator.IsNil(t) {
		return nil
	}
	return timestamppb.New(*t)
}
