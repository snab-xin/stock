package ingest

import "github.com/stock/ashare/internal/domain"

func MarkPaused(bars []domain.DailyBar, suspensions []domain.Suspension) {
	type key struct {
		code string
		day  string
	}
	paused := make(map[key]struct{}, len(suspensions))
	for _, s := range suspensions {
		paused[key{s.TSCode, s.TradeDate.Format("20060102")}] = struct{}{}
	}
	for i := range bars {
		if _, ok := paused[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]; ok {
			bars[i].IsPaused = true
		}
	}
}
