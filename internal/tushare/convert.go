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
		Area:       row.Str("area"),
		Industry:   row.Str("industry"),
		Fullname:   row.Str("fullname"),
		Enname:     row.Str("enname"),
		Cnspell:    row.Str("cnspell"),
		Market:     row.Str("market"),
		Exchange:   ex,
		CurrType:   row.Str("curr_type"),
		ListStatus: status,
		ListDate:   row.DatePtr("list_date"),
		DelistDate: row.DatePtr("delist_date"),
		IsHS:       row.Str("is_hs"),
		ActName:    row.Str("act_name"),
		ActEntType: row.Str("act_ent_type"),
		AssetType:  "stock",
		IsST:       isSTName(name),
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
	ahVol, _ := row.Float("ah_vol")
	ahAmount, _ := row.Float("ah_amount")
	bar := domain.DailyBar{
		TSCode:    tsCode,
		TradeDate: d,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePx,
		PreClose:  row.FloatPtr("pre_close"),
		Change:    row.FloatPtr("change"),
		PctChg:    row.FloatPtr("pct_chg"),
		Volume:    tushareVolToShares(vol),
		Amount:    tushareAmountToYuan(amount),
		AHVolume:  tushareVolToShares(ahVol),
		AHAmount:  tushareAmountToYuan(ahAmount),
	}
	bar.IsPaused = bar.Volume == 0
	return bar, true
}

func stkLimitFromRow(row Row) (domain.StkLimit, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.StkLimit{}, false
	}
	return domain.StkLimit{
		TradeDate: d,
		TSCode:    tsCode,
		PreClose:  row.FloatPtr("pre_close"),
		UpLimit:   row.FloatPtr("up_limit"),
		DownLimit: row.FloatPtr("down_limit"),
		AssetType: row.Str("asset_type"),
		Exchange:  row.Str("exchange"),
	}, true
}

func dailyBasicFromRow(row Row) (domain.DailyBasic, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.DailyBasic{}, false
	}
	return domain.DailyBasic{
		TSCode:        tsCode,
		TradeDate:     d,
		Close:         row.FloatPtr("close"),
		TurnoverRate:  row.FloatPtr("turnover_rate"),
		TurnoverRateF: row.FloatPtr("turnover_rate_f"),
		VolumeRatio:   row.FloatPtr("volume_ratio"),
		PE:            row.FloatPtr("pe"),
		PETTM:         row.FloatPtr("pe_ttm"),
		PB:            row.FloatPtr("pb"),
		PS:            row.FloatPtr("ps"),
		PSTTM:         row.FloatPtr("ps_ttm"),
		DVRatio:       row.FloatPtr("dv_ratio"),
		DVTTM:         row.FloatPtr("dv_ttm"),
		TotalShare:    row.FloatPtr("total_share"),
		FloatShare:    row.FloatPtr("float_share"),
		FreeShare:     row.FloatPtr("free_share"),
		TotalMV:       row.FloatPtr("total_mv"),
		CircMV:        row.FloatPtr("circ_mv"),
		LimitStatus:   row.IntPtr("limit_status"),
	}, true
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
		TSCode:        tsCode,
		TradeDate:     d,
		SuspendTiming: row.Str("suspend_timing"),
		SuspendType:   row.Str("suspend_type"),
	}, true
}
