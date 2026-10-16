package ingest

import (
	"testing"
	"time"

	"github.com/stock/ashare/internal/domain"
)

func TestApplyLimitsAndTurnover(t *testing.T) {
	d := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	up, down := 11.22, 9.18
	bars := []domain.DailyBar{{
		TSCode: "000001.SZ", TradeDate: d, Close: 11.22,
	}}
	ApplyLimits(bars, []domain.StkLimit{{
		TSCode: "000001.SZ", TradeDate: d, UpLimit: &up, DownLimit: &down,
	}})
	if bars[0].HighLimit == nil || *bars[0].IsLimitUp != true {
		t.Fatalf("%+v", bars[0])
	}
	tr := 1.25
	ApplyTurnover(bars, []domain.DailyBasic{{
		TSCode: "000001.SZ", TradeDate: d, TurnoverRate: &tr,
	}})
	if bars[0].Turnover == nil || *bars[0].Turnover != 1.25 {
		t.Fatal("turnover")
	}
}
