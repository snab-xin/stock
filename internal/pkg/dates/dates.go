package dates

import (
	"fmt"
	"strings"
	"time"
)

func Parse(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("日期为空")
	}
	for _, layout := range []string{"20060102", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析日期 %q，请用 20060102 或 2006-01-02", s)
}

func YYYYMMDD(t time.Time) string {
	return t.Format("20060102")
}

func ISO(t time.Time) string {
	return t.Format("2006-01-02")
}

func ParseYYYYMMDD(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("日期为空")
	}
	return time.ParseInLocation("20060102", s, time.Local)
}
