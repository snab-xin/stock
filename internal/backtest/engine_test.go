package backtest

import (
	"math"
	"testing"
	"time"

	"github.com/stock/ashare/internal/domain"
)

func day(i int) time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
}

func barsFromCloses(closes []float64) []domain.DailyBar {
	out := make([]domain.DailyBar, len(closes))
	for i, c := range closes {
		out[i] = domain.DailyBar{
			TSCode:    "000001.SZ",
			TradeDate: day(i),
			Open:      c,
			High:      c,
			Low:       c,
			Close:     c,
		}
	}
	return out
}

func TestSMA(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	got := SMA(xs, 2)
	if !math.IsNaN(got[0]) {
		t.Fatal("first should be NaN")
	}
	if math.Abs(got[1]-1.5) > 1e-9 || math.Abs(got[2]-2.5) > 1e-9 {
		t.Fatalf("%v", got)
	}
}

func TestGoldenCrossBuysNextOpen(t *testing.T) {
	// slow=3, fast=2：前几天横盘，随后上涨穿过。
	closes := []float64{10, 10, 10, 10, 11, 12, 13, 14}
	raw := barsFromCloses(closes)
	p := DefaultParams()
	p.Fast, p.Slow = 2, 3
	p.Slippage = 0
	p.CommissionRate = 0
	p.CommissionMin = 0
	p.StampTaxRate = 0
	p.InitCash = 10000
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	var buy *Trade
	for i := range res.Trades {
		if res.Trades[i].Side == "buy" {
			buy = &res.Trades[i]
			break
		}
	}
	if buy == nil {
		t.Fatalf("应有买入: %+v", res.Trades)
	}
	// i=4 收盘快线上穿，i=5 开盘买入。
	if !buy.Date.Equal(day(5)) {
		t.Fatalf("买入日 got=%s want=%s trades=%+v", buy.Date.Format("2006-01-02"), day(5).Format("2006-01-02"), res.Trades)
	}
}

func TestLimitUpBlocksBuy(t *testing.T) {
	closes := []float64{10, 10, 10, 11, 12, 13, 14, 15}
	raw := barsFromCloses(closes)
	tru := true
	for i := range raw {
		raw[i].IsLimitUp = &tru
	}
	p := DefaultParams()
	p.Fast, p.Slow = 2, 3
	p.CommissionMin = 0
	p.CommissionRate = 0
	p.Slippage = 0
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range res.Trades {
		if tr.Side == "buy" {
			t.Fatalf("涨停不应买入: %+v", tr)
		}
	}
}

func TestCashConstraint(t *testing.T) {
	closes := make([]float64, 8)
	for i := range closes {
		closes[i] = 100
	}
	closes[6], closes[7] = 120, 130
	raw := barsFromCloses(closes)
	p := DefaultParams()
	p.Fast, p.Slow = 2, 3
	p.InitCash = 150
	p.LotSize = 100
	p.CommissionMin = 0
	p.CommissionRate = 0
	p.Slippage = 0
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range res.Trades {
		if tr.Side == "buy" {
			t.Fatalf("现金不够 100 股: %+v", tr)
		}
	}
}

func TestSellCostHandCheck(t *testing.T) {
	// 一手买入再卖出，核对佣金和印花税。
	p := DefaultParams()
	p.Fast, p.Slow = 1, 2
	p.InitCash = 1_000_000
	p.LotSize = 100
	p.Slippage = 0
	p.CommissionRate = 0.00025
	p.CommissionMin = 5
	p.StampTaxRate = 0.0005

	closes := []float64{10, 9, 9, 11, 12, 12, 8, 8, 8}
	raw := barsFromCloses(closes)
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	var sell *Trade
	for i := range res.Trades {
		if res.Trades[i].Side == "sell" {
			sell = &res.Trades[i]
		}
	}
	if sell == nil {
		t.Fatalf("应有卖出: %+v", res.Trades)
	}
	wantStamp := round2(sell.Amount * 0.0005)
	if math.Abs(sell.StampTax-wantStamp) > 0.011 {
		t.Fatalf("印花税 got=%v want=%v amount=%v", sell.StampTax, wantStamp, sell.Amount)
	}
}

func TestPausedBlocks(t *testing.T) {
	closes := []float64{10, 10, 11, 12, 13, 14, 15, 16}
	raw := barsFromCloses(closes)
	for i := range raw {
		raw[i].IsPaused = true
	}
	p := DefaultParams()
	p.Fast, p.Slow = 2, 3
	p.CommissionMin, p.CommissionRate, p.Slippage = 0, 0, 0
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range res.Trades {
		if tr.Side == "buy" || tr.Side == "sell" {
			t.Fatalf("停牌不应成交: %+v", tr)
		}
	}
}

func TestOneRoundTripCosts(t *testing.T) {
	closes := []float64{10, 10, 12, 12, 8, 8}
	raw := barsFromCloses(closes)
	p := DefaultParams()
	p.Fast, p.Slow = 1, 2
	p.InitCash = 1300
	p.LotSize = 100
	p.Slippage = 0
	p.CommissionRate = 0.00025
	p.CommissionMin = 5
	p.StampTaxRate = 0.0005
	res, err := Run(raw, raw, p)
	if err != nil {
		t.Fatal(err)
	}
	var buy, sell *Trade
	for i := range res.Trades {
		switch res.Trades[i].Side {
		case "buy":
			buy = &res.Trades[i]
		case "sell":
			sell = &res.Trades[i]
		}
	}
	if buy == nil || sell == nil {
		t.Fatalf("应有买卖: %+v", res.Trades)
	}
	if buy.Shares != 100 || math.Abs(buy.Amount-1200) > 1e-6 || math.Abs(buy.Commission-5) > 1e-6 {
		t.Fatalf("买入核对失败: %+v", buy)
	}
	if sell.Shares != 100 || math.Abs(sell.Amount-800) > 1e-6 || math.Abs(sell.Commission-5) > 1e-6 {
		t.Fatalf("卖出核对失败: %+v", sell)
	}
	if math.Abs(sell.StampTax-0.4) > 0.011 {
		t.Fatalf("印花税 got=%v want=0.4", sell.StampTax)
	}
	if !sell.Date.After(buy.Date) {
		t.Fatalf("T+1 卖出应晚于买入: buy=%s sell=%s", buy.Date, sell.Date)
	}
}

func TestMaxAffordableShares(t *testing.T) {
	sh := maxAffordableShares(100_000, 10, 0.00025, 5, 100)
	if sh <= 0 || sh%100 != 0 {
		t.Fatalf("shares=%d", sh)
	}
	cost := 10 * float64(sh)
	comm := commission(cost, 0.00025, 5)
	if cost+comm > 100_000+1e-6 {
		t.Fatalf("超现金 %v+%v", cost, comm)
	}
}
