package backtest

import (
	"fmt"
	"math"
	"time"

	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
)

type Trade struct {
	Date       time.Time `json:"date"`
	Side       string    `json:"side"`
	Price      float64   `json:"price"`
	Shares     int64     `json:"shares"`
	Amount     float64   `json:"amount"`
	Commission float64   `json:"commission"`
	StampTax   float64   `json:"stamp_tax"`
	Slippage   float64   `json:"slippage"`
	CashAfter  float64   `json:"cash_after"`
	Reason     string    `json:"reason"`
}

type EquityPoint struct {
	Date   time.Time `json:"date"`
	Cash   float64   `json:"cash"`
	Shares int64     `json:"shares"`
	Close  float64   `json:"close"`
	Equity float64   `json:"equity"`
	Target int       `json:"target"`
}

type Result struct {
	Code          string        `json:"code"`
	Name          string        `json:"name,omitempty"`
	Strategy      string        `json:"strategy"`
	Fast          int           `json:"fast"`
	Slow          int           `json:"slow"`
	InitCash      float64       `json:"init_cash"`
	EndEquity     float64       `json:"end_equity"`
	TotalReturn   float64       `json:"total_return"`
	MaxDrawdown   float64       `json:"max_drawdown"`
	AnnReturn     float64       `json:"ann_return"`
	BuyHoldReturn float64       `json:"buy_hold_return"`
	TradeCount    int           `json:"trade_count"`
	WinRate       float64       `json:"win_rate"`
	Commission    float64       `json:"commission"`
	StampTax      float64       `json:"stamp_tax"`
	SlippageCost  float64       `json:"slippage_cost"`
	BarCount      int           `json:"bar_count"`
	Start         string        `json:"start"`
	End           string        `json:"end"`
	Trades        []Trade       `json:"trades"`
	Equity        []EquityPoint `json:"equity"`
	Limitations   []string      `json:"limitations"`
}

func Align(raw, qfq []domain.DailyBar) (rawOut, qfqOut []domain.DailyBar) {
	qmap := make(map[string]domain.DailyBar, len(qfq))
	for _, b := range qfq {
		qmap[dates.ISO(b.TradeDate)] = b
	}
	for _, b := range raw {
		q, ok := qmap[dates.ISO(b.TradeDate)]
		if !ok {
			continue
		}
		rawOut = append(rawOut, b)
		qfqOut = append(qfqOut, q)
	}
	return rawOut, qfqOut
}

func Run(raw, qfq []domain.DailyBar, p Params) (Result, error) {
	p = p.withDefaults()
	if p.Fast >= p.Slow {
		return Result{}, fmt.Errorf("fast(%d) 必须小于 slow(%d)", p.Fast, p.Slow)
	}
	raw, qfq = Align(raw, qfq)
	if len(raw) < p.Slow+2 {
		return Result{}, fmt.Errorf("对齐后日线不足：%d 根，双均线 %d/%d 至少需要 %d 根", len(raw), p.Fast, p.Slow, p.Slow+2)
	}

	closes := make([]float64, len(qfq))
	for i, b := range qfq {
		closes[i] = b.Close
	}
	fast := SMA(closes, p.Fast)
	slow := SMA(closes, p.Slow)

	cash := p.InitCash
	var shares int64
	boughtOn := -1
	target := 0
	var trades []Trade
	var equity []EquityPoint
	var commSum, stampSum, slipSum float64
	lastSkip := ""

	inWindow := func(t time.Time) bool {
		if !p.TradeFrom.IsZero() && dates.ISO(t) < dates.ISO(p.TradeFrom) {
			return false
		}
		if !p.TradeTo.IsZero() && dates.ISO(t) > dates.ISO(p.TradeTo) {
			return false
		}
		return true
	}

	for i := 0; i < len(raw); i++ {
		bar := raw[i]
		if inWindow(bar.TradeDate) {
			if target == 1 && shares == 0 {
				t, ok := tryBuy(bar, &cash, &shares, p)
				if ok {
					boughtOn = i
					commSum += t.Commission
					slipSum += t.Slippage
					lastSkip = ""
					trades = append(trades, t)
				} else {
					lastSkip = appendSkip(&trades, t, lastSkip)
				}
			} else if target == 0 && shares > 0 {
				if boughtOn == i {
					lastSkip = appendSkip(&trades, Trade{
						Date:   bar.TradeDate,
						Side:   "skip",
						Reason: "T+1 当日买入不可卖",
					}, lastSkip)
				} else if t, ok := trySell(bar, &cash, &shares, p); ok {
					commSum += t.Commission
					stampSum += t.StampTax
					slipSum += t.Slippage
					boughtOn = -1
					lastSkip = ""
					trades = append(trades, t)
				} else {
					lastSkip = appendSkip(&trades, t, lastSkip)
				}
			}
		}

		eq := cash + float64(shares)*bar.Close
		equity = append(equity, EquityPoint{
			Date:   bar.TradeDate,
			Cash:   round2(cash),
			Shares: shares,
			Close:  bar.Close,
			Equity: round2(eq),
			Target: target,
		})

		if i >= p.Slow-1 && !math.IsNaN(fast[i]) && !math.IsNaN(slow[i]) {
			if fast[i] > slow[i] {
				target = 1
			} else {
				target = 0
			}
		}
	}

	endEq := p.InitCash
	if len(equity) > 0 {
		endEq = equity[len(equity)-1].Equity
	}

	fills := 0
	wins := 0
	rounds := 0
	var pendingBuy float64
	hasBuy := false
	for _, t := range trades {
		switch t.Side {
		case "buy":
			fills++
			pendingBuy = t.Amount + t.Commission
			hasBuy = true
		case "sell":
			fills++
			if hasBuy {
				rounds++
				net := t.Amount - t.Commission - t.StampTax
				if net > pendingBuy {
					wins++
				}
				hasBuy = false
			}
		}
	}
	winRate := 0.0
	if rounds > 0 {
		winRate = float64(wins) / float64(rounds)
	}

	bh := 0.0
	first := -1
	last := -1
	for i, b := range raw {
		if inWindow(b.TradeDate) {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first >= 0 && last >= 0 && qfq[first].Close > 0 {
		bh = qfq[last].Close/qfq[first].Close - 1
	}

	start, end := "", ""
	if first >= 0 {
		start = dates.ISO(raw[first].TradeDate)
		end = dates.ISO(raw[last].TradeDate)
	}

	nDays := 0
	if first >= 0 {
		nDays = last - first
	}
	ann := 0.0
	total := 0.0
	if p.InitCash > 0 {
		total = endEq/p.InitCash - 1
	}
	if nDays > 0 {
		ann = math.Pow(1+total, 252/float64(nDays)) - 1
	}

	return Result{
		Code:          p.Code,
		Strategy:      "dual_ma",
		Fast:          p.Fast,
		Slow:          p.Slow,
		InitCash:      p.InitCash,
		EndEquity:     endEq,
		TotalReturn:   total,
		MaxDrawdown:   maxDrawdown(equity),
		AnnReturn:     ann,
		BuyHoldReturn: bh,
		TradeCount:    fills,
		WinRate:       winRate,
		Commission:    round2(commSum),
		StampTax:      round2(stampSum),
		SlippageCost:  round2(slipSum),
		BarCount:      len(raw),
		Start:         start,
		End:           end,
		Trades:        trades,
		Equity:        equity,
		Limitations:   KnownLimitations(),
	}, nil
}

func appendSkip(dst *[]Trade, t Trade, last string) string {
	if t.Reason == "" || t.Reason == last {
		return last
	}
	*dst = append(*dst, t)
	return t.Reason
}

func tryBuy(bar domain.DailyBar, cash *float64, shares *int64, p Params) (Trade, bool) {
	t := Trade{Date: bar.TradeDate, Side: "skip"}
	if reason := blockReason(bar, "buy"); reason != "" {
		t.Reason = reason
		return t, false
	}
	px := buyFillPrice(bar.Open, p.Slippage)
	sh := maxAffordableShares(*cash, px, p.CommissionRate, p.CommissionMin, p.LotSize)
	if sh <= 0 {
		t.Reason = "现金不足"
		return t, false
	}
	amt := px * float64(sh)
	comm := commission(amt, p.CommissionRate, p.CommissionMin)
	*cash -= amt + comm
	*shares = sh
	slip := (px - bar.Open) * float64(sh)
	return Trade{
		Date:       bar.TradeDate,
		Side:       "buy",
		Price:      round2(px),
		Shares:     sh,
		Amount:     round2(amt),
		Commission: round2(comm),
		Slippage:   round2(slip),
		CashAfter:  round2(*cash),
		Reason:     "均线多头，T+1 开盘买入",
	}, true
}

func trySell(bar domain.DailyBar, cash *float64, shares *int64, p Params) (Trade, bool) {
	t := Trade{Date: bar.TradeDate, Side: "skip"}
	if reason := blockReason(bar, "sell"); reason != "" {
		t.Reason = reason
		return t, false
	}
	px := sellFillPrice(bar.Open, p.Slippage)
	sh := *shares
	amt := px * float64(sh)
	comm := commission(amt, p.CommissionRate, p.CommissionMin)
	stamp := amt * p.StampTaxRate
	*cash += amt - comm - stamp
	*shares = 0
	slip := (bar.Open - px) * float64(sh)
	return Trade{
		Date:       bar.TradeDate,
		Side:       "sell",
		Price:      round2(px),
		Shares:     sh,
		Amount:     round2(amt),
		Commission: round2(comm),
		StampTax:   round2(stamp),
		Slippage:   round2(slip),
		CashAfter:  round2(*cash),
		Reason:     "均线空头，T+1 开盘卖出",
	}, true
}

func blockReason(bar domain.DailyBar, side string) string {
	if bar.IsPaused {
		return "停牌不成交"
	}
	if bar.Open <= 0 {
		return "开盘价无效"
	}
	switch side {
	case "buy":
		if bar.IsLimitUp != nil && *bar.IsLimitUp {
			return "涨停无法买入"
		}
		if bar.HighLimit != nil && bar.Open >= *bar.HighLimit-1e-6 {
			return "开盘触及涨停价"
		}
	case "sell":
		if bar.IsLimitDown != nil && *bar.IsLimitDown {
			return "跌停无法卖出"
		}
		if bar.LowLimit != nil && bar.Open <= *bar.LowLimit+1e-6 {
			return "开盘触及跌停价"
		}
	}
	return ""
}

func maxDrawdown(eq []EquityPoint) float64 {
	if len(eq) == 0 {
		return 0
	}
	peak := eq[0].Equity
	dd := 0.0
	for _, p := range eq {
		if p.Equity > peak {
			peak = p.Equity
		}
		if peak > 0 {
			d := p.Equity/peak - 1
			if d < dd {
				dd = d
			}
		}
	}
	return dd
}
