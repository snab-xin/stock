package tushare

import (
	"math"
	"strings"

	"github.com/stock/ashare/internal/domain"
)

func exchangeFromTSCode(tsCode string) string {
	switch {
	case strings.HasSuffix(tsCode, ".SH"):
		return "SSE"
	case strings.HasSuffix(tsCode, ".SZ"):
		return "SZSE"
	case strings.HasSuffix(tsCode, ".BJ"):
		return "BSE"
	default:
		return ""
	}
}

func isSTName(name string) bool {
	n := strings.ToUpper(name)
	return strings.Contains(n, "ST")
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func limitHit(close, limit float64, up bool) bool {
	if limit == 0 {
		return false
	}
	c, l := round2(close), round2(limit)
	if up {
		return c >= l
	}
	return c <= l
}

// tushareVolToShares：Tushare daily.vol 单位是手。
func tushareVolToShares(vol float64) int64 {
	return int64(math.Round(vol * 100))
}

// tushareAmountToYuan：Tushare daily.amount 单位是千元。
func tushareAmountToYuan(amount float64) float64 {
	return amount * 1000
}

func instrumentFromRow(row Row) (domain.Instrument, error) {
	tsCode := row.Str("ts_code")
	if tsCode == "" {
		return domain.Instrument{}, nil
	}
	status := row.Str("list_status")
	if status == "" {
		status = "L"
	}
	ex := row.Str("exchange")
	if ex == "" {
		ex = exchangeFromTSCode(tsCode)
	}
	name := row.Str("name")
	return domain.Instrument{
		TSCode:     tsCode,
		Symbol:     row.Str("symbol"),
		Name:       name,
		Exchange:   ex,
		AssetType:  "stock",
		ListDate:   row.DatePtr("list_date"),
		DelistDate: row.DatePtr("delist_date"),
		IsST:       isSTName(name),
		Status:     status,
	}, nil
}

func calendarFromRow(row Row) (domain.TradeCalendar, bool) {
	d, err := row.Date("cal_date")
	if err != nil {
		return domain.TradeCalendar{}, false
	}
	return domain.TradeCalendar{
		Exchange:     row.Str("exchange"),
		CalDate:      d,
		IsOpen:       row.Bool01("is_open"),
		PretradeDate: row.DatePtr("pretrade_date"),
	}, true
}

func dailyFromRow(row Row) (domain.DailyBar, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.DailyBar{}, false
	}
	open, _ := row.Float("open")
	high, _ := row.Float("high")
	low, _ := row.Float("low")
	closePx, _ := row.Float("close")
	vol, _ := row.Float("vol")
	amount, _ := row.Float("amount")
	bar := domain.DailyBar{
		TSCode:    tsCode,
		TradeDate: d,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePx,
		PreClose:  row.FloatPtr("pre_close"),
		Volume:    tushareVolToShares(vol),
		Amount:    tushareAmountToYuan(amount),
		Turnover:  row.FloatPtr("turnover_rate"),
	}
	if up, ok := row.Float("up_limit"); ok {
		bar.HighLimit = &up
		v := limitHit(closePx, up, true)
		bar.IsLimitUp = &v
	}
	if down, ok := row.Float("down_limit"); ok {
		bar.LowLimit = &down
		v := limitHit(closePx, down, false)
		bar.IsLimitDown = &v
	}
	bar.IsPaused = bar.Volume == 0
	return bar, true
}

func adjFromRow(row Row) (domain.AdjFactor, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	f, ok := row.Float("adj_factor")
	if err != nil || tsCode == "" || !ok {
		return domain.AdjFactor{}, false
	}
	return domain.AdjFactor{TSCode: tsCode, TradeDate: d, Factor: f}, true
}

func suspendFromRow(row Row) (domain.Suspension, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.Suspension{}, false
	}
	return domain.Suspension{
		TSCode:      tsCode,
		TradeDate:   d,
		SuspendType: row.Str("suspend_type"),
		Reason:      firstNonEmpty(row.Str("suspend_timing"), row.Str("reason")),
	}, true
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

func applyLimits(bars []domain.DailyBar, limits []Row) {
	type key struct {
		code string
		day  string
	}
	idx := make(map[key]Row, len(limits))
	for _, row := range limits {
		d, err := row.Date("trade_date")
		if err != nil {
			continue
		}
		idx[key{row.Str("ts_code"), d.Format("20060102")}] = row
	}
	for i := range bars {
		row, ok := idx[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]
		if !ok {
			continue
		}
		if up, ok := row.Float("up_limit"); ok {
			bars[i].HighLimit = &up
			v := limitHit(bars[i].Close, up, true)
			bars[i].IsLimitUp = &v
		}
		if down, ok := row.Float("down_limit"); ok {
			bars[i].LowLimit = &down
			v := limitHit(bars[i].Close, down, false)
			bars[i].IsLimitDown = &v
		}
	}
}
