package ingest

import (
	"math"

	"github.com/stock/ashare/internal/domain"
)

func MarkPaused(bars []domain.DailyBar, suspensions []domain.Suspension) {
	type key struct {
		code string
		day  string
	}
	paused := make(map[key]struct{}, len(suspensions))
	for _, s := range suspensions {
		if s.SuspendType == "R" {
			continue
		}
		paused[key{s.TSCode, s.TradeDate.Format("20060102")}] = struct{}{}
	}
	for i := range bars {
		if _, ok := paused[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]; ok {
			bars[i].IsPaused = true
		}
	}
}

func ApplyLimits(bars []domain.DailyBar, limits []domain.StkLimit) {
	type key struct {
		code string
		day  string
	}
	idx := make(map[key]domain.StkLimit, len(limits))
	for _, row := range limits {
		idx[key{row.TSCode, row.TradeDate.Format("20060102")}] = row
	}
	for i := range bars {
		row, ok := idx[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]
		if !ok {
			continue
		}
		if row.UpLimit != nil {
			up := *row.UpLimit
			bars[i].HighLimit = &up
			v := closeHitsLimit(bars[i].Close, up, true)
			bars[i].IsLimitUp = &v
		}
		if row.DownLimit != nil {
			down := *row.DownLimit
			bars[i].LowLimit = &down
			v := closeHitsLimit(bars[i].Close, down, false)
			bars[i].IsLimitDown = &v
		}
	}
}

func ApplyTurnover(bars []domain.DailyBar, basics []domain.DailyBasic) {
	type key struct {
		code string
		day  string
	}
	idx := make(map[key]*float64, len(basics))
	for _, row := range basics {
		idx[key{row.TSCode, row.TradeDate.Format("20060102")}] = row.TurnoverRate
	}
	for i := range bars {
		if t, ok := idx[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]; ok {
			bars[i].Turnover = t
		}
	}
}

func closeHitsLimit(close, limit float64, up bool) bool {
	c := round2(close)
	l := round2(limit)
	if l == 0 {
		return false
	}
	if up {
		return c >= l
	}
	return c <= l
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
