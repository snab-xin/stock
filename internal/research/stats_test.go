package research

import (
	"testing"
	"time"

	"github.com/stock/ashare/internal/domain"
)

func TestFromBars(t *testing.T) {
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := []domain.DailyBar{
		{TradeDate: day, Close: 10},
		{TradeDate: day.AddDate(0, 0, 1), Close: 12},
		{TradeDate: day.AddDate(0, 0, 2), Close: 9},
		{TradeDate: day.AddDate(0, 0, 3), Close: 11},
	}
	s, err := FromBars(bars)
	if err != nil {
		t.Fatal(err)
	}
	if s.N != 4 || mathAbs(s.TotalReturn-0.1) > 1e-9 {
		t.Fatalf("%+v", s)
	}
	wantDD := 9.0/12.0 - 1
	if mathAbs(s.MaxDrawdown-wantDD) > 1e-9 {
		t.Fatalf("dd=%v want=%v", s.MaxDrawdown, wantDD)
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
