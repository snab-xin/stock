package tushare

import (
	"context"
	"time"

	"github.com/stock/ashare/internal/pkg/dates"
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
	// 股票基本信息 https://tushare.pro/document/2?doc_id=25
	for _, status := range []string{"L", "D", "P", "G", "UN"} {
		rows, err := s.client.Query(ctx, "stock_basic", map[string]string{
			"list_status": status,
		}, stockBasicFields)
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
	// 交易日历 https://tushare.pro/document/2?doc_id=26
	rows, err := s.client.Query(ctx, "trade_cal", map[string]string{
		"exchange":   exchange,
		"start_date": dates.YYYYMMDD(start),
		"end_date":   dates.YYYYMMDD(end),
	}, tradeCalFields)
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
	// 日线行情 https://tushare.pro/document/2?doc_id=27
	dailyRows, err := s.client.Query(ctx, "daily", barParams(q), dailyFields)
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
	return bars, nil
}

func (s *Source) ListStkLimits(ctx context.Context, q ingest.BarQuery) ([]domain.StkLimit, error) {
	// 涨跌停价格 https://tushare.pro/document/2?doc_id=183
	rows, err := s.client.Query(ctx, "stk_limit", barParams(q), stkLimitFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.StkLimit, 0, len(rows))
	for _, row := range rows {
		if it, ok := stkLimitFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListDailyBasics(ctx context.Context, q ingest.BarQuery) ([]domain.DailyBasic, error) {
	// 每日指标 https://tushare.pro/document/2?doc_id=32
	rows, err := s.client.Query(ctx, "daily_basic", barParams(q), dailyBasicFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DailyBasic, 0, len(rows))
	for _, row := range rows {
		if it, ok := dailyBasicFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListAdjFactors(ctx context.Context, q ingest.BarQuery) ([]domain.AdjFactor, error) {
	// 复权因子 https://tushare.pro/document/2?doc_id=28
	rows, err := s.client.Query(ctx, "adj_factor", barParams(q), adjFactorFields)
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
	// 停牌复牌信息 https://tushare.pro/document/2?doc_id=214
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
	rows, err := s.client.QueryPages(ctx, "suspend_d", params, suspendFields, 5000)
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
