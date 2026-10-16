package tushare

import (
	"context"
	"fmt"

	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/ingest"
)

var indexMarkets = []string{"MSCI", "CSI", "SSE", "SZSE", "CICC", "SW", "OTH"}

func (s *Source) ListIndexBasic(ctx context.Context) ([]domain.IndexBasic, error) {
	// 指数基本信息 https://tushare.pro/document/2?doc_id=94
	var all []domain.IndexBasic
	seen := map[string]struct{}{}
	for _, market := range indexMarkets {
		rows, err := s.client.Query(ctx, "index_basic", map[string]string{
			"market": market,
		}, indexBasicFields)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, ok := indexBasicFromRow(row)
			if !ok {
				continue
			}
			if _, dup := seen[it.TSCode]; dup {
				continue
			}
			seen[it.TSCode] = struct{}{}
			all = append(all, it)
		}
	}
	return all, nil
}

func (s *Source) ListIndexDaily(ctx context.Context, q ingest.BarQuery) ([]domain.IndexDaily, error) {
	// 指数日线行情 https://tushare.pro/document/2?doc_id=95
	if q.TSCode == "" {
		return nil, fmt.Errorf("index_daily 需要 ts_code")
	}
	rows, err := s.client.Query(ctx, "index_daily", barParams(q), indexDailyFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.IndexDaily, 0, len(rows))
	for _, row := range rows {
		if it, ok := indexDailyFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListIndexClassify(ctx context.Context) ([]domain.IndexClassify, error) {
	// 申万行业分类 https://tushare.pro/document/2?doc_id=181
	var all []domain.IndexClassify
	for _, src := range []string{"SW2014", "SW2021"} {
		rows, err := s.client.Query(ctx, "index_classify", map[string]string{
			"src": src,
		}, indexClassifyFields)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, ok := indexClassifyFromRow(row)
			if !ok {
				continue
			}
			it.Src = src
			all = append(all, it)
		}
	}
	return all, nil
}

func (s *Source) ListSWDaily(ctx context.Context, q ingest.BarQuery) ([]domain.SWDaily, error) {
	// 申万日线行情 https://tushare.pro/document/2?doc_id=327
	rows, err := s.client.Query(ctx, "sw_daily", barParams(q), swDailyFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SWDaily, 0, len(rows))
	for _, row := range rows {
		if it, ok := swDailyFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func indexBasicFromRow(row Row) (domain.IndexBasic, bool) {
	tsCode := row.Str("ts_code")
	if tsCode == "" {
		return domain.IndexBasic{}, false
	}
	return domain.IndexBasic{
		TSCode:      tsCode,
		Name:        row.Str("name"),
		Fullname:    row.Str("fullname"),
		Market:      row.Str("market"),
		Publisher:   row.Str("publisher"),
		IndexType:   row.Str("index_type"),
		Category:    row.Str("category"),
		BaseDate:    row.DatePtr("base_date"),
		BasePoint:   row.FloatPtr("base_point"),
		ListDate:    row.DatePtr("list_date"),
		WeightRule:  row.Str("weight_rule"),
		Description: row.Str("desc"),
		ExpDate:     row.DatePtr("exp_date"),
	}, true
}

func indexDailyFromRow(row Row) (domain.IndexDaily, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.IndexDaily{}, false
	}
	return domain.IndexDaily{
		TSCode:    tsCode,
		TradeDate: d,
		Close:     row.FloatPtr("close"),
		Open:      row.FloatPtr("open"),
		High:      row.FloatPtr("high"),
		Low:       row.FloatPtr("low"),
		PreClose:  row.FloatPtr("pre_close"),
		Change:    row.FloatPtr("change"),
		PctChg:    row.FloatPtr("pct_chg"),
		Vol:       row.FloatPtr("vol"),
		Amount:    row.FloatPtr("amount"),
	}, true
}

func indexClassifyFromRow(row Row) (domain.IndexClassify, bool) {
	indexCode := row.Str("index_code")
	if indexCode == "" {
		return domain.IndexClassify{}, false
	}
	return domain.IndexClassify{
		IndexCode:    indexCode,
		IndustryName: row.Str("industry_name"),
		ParentCode:   row.Str("parent_code"),
		Level:        row.Str("level"),
		IndustryCode: row.Str("industry_code"),
		IsPub:        row.Str("is_pub"),
		Src:          row.Str("src"),
	}, true
}

func swDailyFromRow(row Row) (domain.SWDaily, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.SWDaily{}, false
	}
	return domain.SWDaily{
		TSCode:    tsCode,
		TradeDate: d,
		Name:      row.Str("name"),
		Open:      row.FloatPtr("open"),
		Low:       row.FloatPtr("low"),
		High:      row.FloatPtr("high"),
		Close:     row.FloatPtr("close"),
		Change:    row.FloatPtr("change"),
		PctChange: row.FloatPtr("pct_change"),
		Vol:       row.FloatPtr("vol"),
		Amount:    row.FloatPtr("amount"),
		PE:        row.FloatPtr("pe"),
		PB:        row.FloatPtr("pb"),
		FloatMV:   row.FloatPtr("float_mv"),
		TotalMV:   row.FloatPtr("total_mv"),
	}, true
}
