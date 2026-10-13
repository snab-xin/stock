package tushare

import (
	"testing"
	"time"
)

func TestExchangeFromTSCode(t *testing.T) {
	if exchangeFromTSCode("000001.SZ") != "SZSE" {
		t.Fatal()
	}
	if exchangeFromTSCode("600000.SH") != "SSE" {
		t.Fatal()
	}
	if exchangeFromTSCode("430047.BJ") != "BSE" {
		t.Fatal()
	}
}

func TestVolumeAmountConvert(t *testing.T) {
	if tushareVolToShares(123.4) != 12340 {
		t.Fatalf("got %d", tushareVolToShares(123.4))
	}
	if tushareAmountToYuan(56.7) != 56700 {
		t.Fatalf("got %v", tushareAmountToYuan(56.7))
	}
}

func TestLimitHit(t *testing.T) {
	if !limitHit(11.00, 11.00, true) {
		t.Fatal("涨停应命中")
	}
	if limitHit(10.99, 11.00, true) {
		t.Fatal("未涨停")
	}
	if !limitHit(9.00, 9.00, false) {
		t.Fatal("跌停应命中")
	}
}

func TestDailyFromRow(t *testing.T) {
	bar, ok := dailyFromRow(Row{
		"ts_code":    "000001.SZ",
		"trade_date": "20260105",
		"open":       "10.1",
		"high":       "10.5",
		"low":        "10.0",
		"close":      "10.4",
		"pre_close":  "10.2",
		"vol":        "1000",
		"amount":     "200",
		"up_limit":   "11.22",
		"down_limit": "9.18",
	})
	if !ok {
		t.Fatal("parse")
	}
	if bar.Volume != 100000 {
		t.Fatalf("volume=%d", bar.Volume)
	}
	if bar.Amount != 200000 {
		t.Fatalf("amount=%v", bar.Amount)
	}
	if bar.TradeDate.Format("20060102") != "20260105" {
		t.Fatalf("date=%v", bar.TradeDate)
	}
	if bar.IsPaused {
		t.Fatal("paused")
	}
	if bar.HighLimit == nil || *bar.HighLimit != 11.22 {
		t.Fatal("high limit")
	}
}

func TestIsSTName(t *testing.T) {
	if !isSTName("*ST宁科") || !isSTName("ST华微") {
		t.Fatal()
	}
	if isSTName("平安银行") {
		t.Fatal()
	}
}

func TestCalendarFromRow(t *testing.T) {
	row, ok := calendarFromRow(Row{
		"exchange":      "SSE",
		"cal_date":      "20260105",
		"is_open":       "1",
		"pretrade_date": "20251231",
	})
	if !ok || !row.IsOpen {
		t.Fatal()
	}
	if row.PretradeDate == nil || row.PretradeDate.Format(time.DateOnly) != "2025-12-31" {
		t.Fatalf("pre=%v", row.PretradeDate)
	}
}
