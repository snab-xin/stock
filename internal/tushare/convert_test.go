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
		"change":     "0.2",
		"pct_chg":    "1.96",
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
	if bar.Change == nil || *bar.Change != 0.2 {
		t.Fatal("change")
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

func TestInstrumentFromRowAllFields(t *testing.T) {
	it, err := instrumentFromRow(Row{
		"ts_code":      "000001.SZ",
		"symbol":       "000001",
		"name":         "平安银行",
		"area":         "深圳",
		"industry":     "银行",
		"fullname":     "平安银行股份有限公司",
		"enname":       "Ping An Bank",
		"cnspell":      "PAYH",
		"market":       "主板",
		"exchange":     "SZSE",
		"curr_type":    "CNY",
		"list_status":  "L",
		"list_date":    "19910403",
		"is_hs":        "S",
		"act_name":     "深圳市投资控股",
		"act_ent_type": "地方国有",
	})
	if err != nil || it.TSCode != "000001.SZ" || it.Area != "深圳" || it.ListStatus != "L" || it.IsHS != "S" {
		t.Fatalf("%+v err=%v", it, err)
	}
}

func TestStkLimitAndDailyBasicFromRow(t *testing.T) {
	lim, ok := stkLimitFromRow(Row{
		"ts_code":    "000001.SZ",
		"trade_date": "20260105",
		"up_limit":   "11.22",
		"down_limit": "9.18",
		"asset_type": "STK",
		"exchange":   "SZSE",
	})
	if !ok || lim.UpLimit == nil || *lim.UpLimit != 11.22 || lim.AssetType != "STK" {
		t.Fatalf("%+v", lim)
	}
	b, ok := dailyBasicFromRow(Row{
		"ts_code":       "000001.SZ",
		"trade_date":    "20260105",
		"turnover_rate": "1.2",
		"pe":            "8.5",
		"total_share":   "1000",
		"limit_status":  "1",
	})
	if !ok || b.PE == nil || *b.PE != 8.5 || b.LimitStatus == nil || *b.LimitStatus != 1 {
		t.Fatalf("%+v", b)
	}
}

func TestIndexAndETFFromRow(t *testing.T) {
	ib, ok := indexBasicFromRow(Row{
		"ts_code":    "000300.SH",
		"name":       "沪深300",
		"market":     "SSE",
		"desc":       "规模指数",
		"base_point": "1000",
	})
	if !ok || ib.Description != "规模指数" || ib.BasePoint == nil {
		t.Fatalf("%+v", ib)
	}
	etf, ok := etfBasicFromRow(Row{
		"ts_code":     "510330.SH",
		"extname":     "沪深300ETF华夏",
		"list_status": "L",
		"exchange":    "SH",
		"etf_type":    "境内",
	})
	if !ok || etf.Extname == "" || etf.ETFType != "境内" {
		t.Fatalf("%+v", etf)
	}
	sh, ok := etfSHConsFromRow(Row{
		"trade_date": "20260615",
		"ts_code":    "517030.SH",
		"con_code":   "000001.SZ",
		"qty":        "1100",
		"cpr":        "-",
	})
	if !ok || sh.Qty == nil || *sh.Qty != 1100 || sh.CPR != nil {
		t.Fatalf("%+v", sh)
	}
}
