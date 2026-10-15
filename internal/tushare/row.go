package tushare

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/stock/ashare/internal/dates"
)

type Row map[string]string

func newRows(fields []string, items [][]any) []Row {
	out := make([]Row, 0, len(items))
	for _, item := range items {
		row := make(Row, len(fields))
		for i, field := range fields {
			if i >= len(item) {
				break
			}
			row[field] = cellString(item[i])
		}
		out = append(out, row)
	}
	return out
}

func cellString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(x)
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return ""
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func (r Row) Str(key string) string {
	return strings.TrimSpace(r[key])
}

func (r Row) Float(key string) (float64, bool) {
	s := r.Str(key)
	if s == "" || s == "nan" || s == "NaN" || s == "None" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (r Row) FloatPtr(key string) *float64 {
	v, ok := r.Float(key)
	if !ok {
		return nil
	}
	return &v
}

func (r Row) Date(key string) (time.Time, error) {
	s := r.Str(key)
	if s == "" {
		return time.Time{}, fmt.Errorf("空日期")
	}
	if t, err := dates.ParseYYYYMMDD(s); err == nil {
		return t, nil
	}
	return dates.Parse(s)
}

func (r Row) DatePtr(key string) *time.Time {
	t, err := r.Date(key)
	if err != nil {
		return nil
	}
	return &t
}

func (r Row) Bool01(key string) bool {
	s := r.Str(key)
	return s == "1" || strings.EqualFold(s, "true")
}

func (r Row) Int(key string) (int, bool) {
	s := r.Str(key)
	if s == "" || s == "nan" || s == "NaN" || s == "None" {
		return 0, false
	}
	if strings.Contains(s, ".") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return int(math.Round(f)), true
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (r Row) IntPtr(key string) *int {
	v, ok := r.Int(key)
	if !ok {
		return nil
	}
	return &v
}
