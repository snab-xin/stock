package tushare

import (
	"context"
	"log/slog"
	"time"

	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/ingest"
)

type Source struct {
	client *Client
}

func NewSource(client *Client) *Source {
	return &Source{client: client}
}

func (s *Source) ListInstruments(ctx context.Context) ([]domain.Instrument, error) {
	var all []domain.Instrument
	for _, status := range []string{"L", "D", "P"} {
		rows, err := s.client.Query(ctx, "stock_basic", map[string]string{
			"list_status": status,
		}, "ts_code,symbol,name,exchange,list_status,list_date,delist_date")
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, err := instrumentFromRow(row)
			if err != nil || it.TSCode == "" {
				continue
			}
			all = append(all, it)
		}
	}
	return all, nil
}

func (s *Source) ListCalendar(ctx context.Context, exchange string, start, end time.Time) ([]domain.TradeCalendar, error) {
	rows, err := s.client.Query(ctx, "trade_cal", map[string]string{
		"exchange":   exchange,
		"start_date": dates.YYYYMMDD(start),
		"end_date":   dates.YYYYMMDD(end),
	}, "exchange,cal_date,is_open,pretrade_date")
	if err != nil {
		return nil, err
	}
	out := make([]domain.TradeCalendar, 0, len(rows))
	for _, row := range rows {
		cal, ok := calendarFromRow(row)
		if !ok {
			continue
		}
		if cal.Exchange == "" {
			cal.Exchange = exchange
		}
		out = append(out, cal)
	}
	return out, nil
}

func (s *Source) ListDailyBars(ctx context.Context, q ingest.BarQuery) ([]domain.DailyBar, error) {
	dailyRows, err := s.client.Query(ctx, "daily", barParams(q), "ts_code,trade_date,open,high,low,close,pre_close,vol,amount")
	if err != nil {
		return nil, err
	}
	bars := make([]domain.DailyBar, 0, len(dailyRows))
	for _, row := range dailyRows {
		bar, ok := dailyFromRow(row)
		if ok {
			bars = append(bars, bar)
		}
	}

	if limitRows, err := s.client.Query(ctx, "stk_limit", barParams(q), "ts_code,trade_date,up_limit,down_limit"); err != nil {
		slog.Warn("stk_limit 跳过", "err", err)
	} else {
		applyLimits(bars, limitRows)
	}

	if basicRows, err := s.client.Query(ctx, "daily_basic", barParams(q), "ts_code,trade_date,turnover_rate"); err != nil {
		slog.Warn("daily_basic 跳过", "err", err)
	} else {
		applyTurnover(bars, basicRows)
	}
	return bars, nil
}

func (s *Source) ListAdjFactors(ctx context.Context, q ingest.BarQuery) ([]domain.AdjFactor, error) {
	rows, err := s.client.Query(ctx, "adj_factor", barParams(q), "ts_code,trade_date,adj_factor")
	if err != nil {
		return nil, err
	}
	out := make([]domain.AdjFactor, 0, len(rows))
	for _, row := range rows {
		if a, ok := adjFromRow(row); ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *Source) ListSuspensions(ctx context.Context, q ingest.BarQuery) ([]domain.Suspension, error) {
	params := map[string]string{}
	if q.TSCode != "" {
		params["ts_code"] = q.TSCode
	}
	if !q.TradeDate.IsZero() {
		params["trade_date"] = dates.YYYYMMDD(q.TradeDate)
	} else {
		if !q.Start.IsZero() {
			params["start_date"] = dates.YYYYMMDD(q.Start)
		}
		if !q.End.IsZero() {
			params["end_date"] = dates.YYYYMMDD(q.End)
		}
	}
	rows, err := s.client.Query(ctx, "suspend_d", params, "ts_code,trade_date,suspend_timing,suspend_type")
	if err != nil {
		return nil, err
	}
	out := make([]domain.Suspension, 0, len(rows))
	for _, row := range rows {
		if srow, ok := suspendFromRow(row); ok {
			out = append(out, srow)
		}
	}
	return out, nil
}

func barParams(q ingest.BarQuery) map[string]string {
	params := map[string]string{}
	if q.TSCode != "" {
		params["ts_code"] = q.TSCode
	}
	if !q.TradeDate.IsZero() {
		params["trade_date"] = dates.YYYYMMDD(q.TradeDate)
		return params
	}
	if !q.Start.IsZero() {
		params["start_date"] = dates.YYYYMMDD(q.Start)
	}
	if !q.End.IsZero() {
		params["end_date"] = dates.YYYYMMDD(q.End)
	}
	return params
}

func applyTurnover(bars []domain.DailyBar, rows []Row) {
	type key struct {
		code string
		day  string
	}
	idx := make(map[key]*float64, len(rows))
	for _, row := range rows {
		d, err := row.Date("trade_date")
		if err != nil {
			continue
		}
		idx[key{row.Str("ts_code"), d.Format("20060102")}] = row.FloatPtr("turnover_rate")
	}
	for i := range bars {
		if t, ok := idx[key{bars[i].TSCode, bars[i].TradeDate.Format("20060102")}]; ok {
			bars[i].Turnover = t
		}
	}
}
