package tushare

import (
	"context"

	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/ingest"
)

func (s *Source) ListETFBasic(ctx context.Context) ([]domain.ETFBasic, error) {
	// etf基本信息 https://tushare.pro/document/2?doc_id=385
	var all []domain.ETFBasic
	seen := map[string]struct{}{}
	for _, status := range []string{"L", "D", "P"} {
		rows, err := s.client.Query(ctx, "etf_basic", map[string]string{
			"list_status": status,
		}, etfBasicFields)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, ok := etfBasicFromRow(row)
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

func (s *Source) ListETFDaily(ctx context.Context, q ingest.BarQuery) ([]domain.ETFDaily, error) {
	// etf日线行情 https://tushare.pro/document/2?doc_id=127
	rows, err := s.client.Query(ctx, "fund_daily", barParams(q), fundDailyFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ETFDaily, 0, len(rows))
	for _, row := range rows {
		if it, ok := etfDailyFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListETFAdj(ctx context.Context, q ingest.BarQuery) ([]domain.ETFAdj, error) {
	// etf基金复权因子 https://tushare.pro/document/2?doc_id=199
	rows, err := s.client.QueryPages(ctx, "fund_adj", barParams(q), fundAdjFields, 2000)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ETFAdj, 0, len(rows))
	for _, row := range rows {
		if it, ok := etfAdjFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListETFShareSize(ctx context.Context, q ingest.BarQuery) ([]domain.ETFShareSize, error) {
	// ETF份额规模 https://tushare.pro/document/2?doc_id=408
	rows, err := s.client.QueryPages(ctx, "etf_share_size", barParams(q), etfShareSizeFields, 5000)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ETFShareSize, 0, len(rows))
	for _, row := range rows {
		if it, ok := etfShareSizeFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListETFSHCons(ctx context.Context, q ingest.BarQuery) ([]domain.ETFSHCons, error) {
	// ETF每日持仓组合(沪市） https://tushare.pro/document/2?doc_id=471
	rows, err := s.client.QueryPages(ctx, "etf_sh_cons", etfConsParams(q), etfSHConsFields, 3000)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ETFSHCons, 0, len(rows))
	for _, row := range rows {
		if it, ok := etfSHConsFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListETFSZCons(ctx context.Context, q ingest.BarQuery) ([]domain.ETFSZCons, error) {
	// ETF每日持仓组合(深市） https://tushare.pro/document/2?doc_id=472
	rows, err := s.client.QueryPages(ctx, "etf_sz_cons", etfConsParams(q), etfSZConsFields, 3000)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ETFSZCons, 0, len(rows))
	for _, row := range rows {
		if it, ok := etfSZConsFromRow(row); ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func etfConsParams(q ingest.BarQuery) map[string]string {
	return barParams(q)
}

func etfBasicFromRow(row Row) (domain.ETFBasic, bool) {
	tsCode := row.Str("ts_code")
	if tsCode == "" {
		return domain.ETFBasic{}, false
	}
	return domain.ETFBasic{
		TSCode:     tsCode,
		Csname:     row.Str("csname"),
		Extname:    row.Str("extname"),
		Cname:      row.Str("cname"),
		IndexCode:  row.Str("index_code"),
		IndexName:  row.Str("index_name"),
		SetupDate:  row.DatePtr("setup_date"),
		ListDate:   row.DatePtr("list_date"),
		ListStatus: row.Str("list_status"),
		Exchange:   row.Str("exchange"),
		MgrName:    row.Str("mgr_name"),
		CustodName: row.Str("custod_name"),
		MgtFee:     row.FloatPtr("mgt_fee"),
		ETFType:    row.Str("etf_type"),
	}, true
}

func etfDailyFromRow(row Row) (domain.ETFDaily, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.ETFDaily{}, false
	}
	return domain.ETFDaily{
		TSCode:    tsCode,
		TradeDate: d,
		Open:      row.FloatPtr("open"),
		High:      row.FloatPtr("high"),
		Low:       row.FloatPtr("low"),
		Close:     row.FloatPtr("close"),
		PreClose:  row.FloatPtr("pre_close"),
		Change:    row.FloatPtr("change"),
		PctChg:    row.FloatPtr("pct_chg"),
		Vol:       row.FloatPtr("vol"),
		Amount:    row.FloatPtr("amount"),
	}, true
}

func etfAdjFromRow(row Row) (domain.ETFAdj, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	f, ok := row.Float("adj_factor")
	if err != nil || tsCode == "" || !ok {
		return domain.ETFAdj{}, false
	}
	return domain.ETFAdj{TSCode: tsCode, TradeDate: d, Factor: f}, true
}

func etfShareSizeFromRow(row Row) (domain.ETFShareSize, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.ETFShareSize{}, false
	}
	return domain.ETFShareSize{
		TradeDate:  d,
		TSCode:     tsCode,
		ETFName:    row.Str("etf_name"),
		TotalShare: row.FloatPtr("total_share"),
		TotalSize:  row.FloatPtr("total_size"),
		NAV:        row.FloatPtr("nav"),
		Close:      row.FloatPtr("close"),
		Exchange:   row.Str("exchange"),
	}, true
}

func etfSHConsFromRow(row Row) (domain.ETFSHCons, bool) {
	tsCode := row.Str("ts_code")
	conCode := row.Str("con_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" || conCode == "" {
		return domain.ETFSHCons{}, false
	}
	return domain.ETFSHCons{
		TradeDate: d,
		TSCode:    tsCode,
		ConCode:   conCode,
		ConName:   row.Str("con_name"),
		Qty:       row.Int64Ptr("qty"),
		SubFlag:   row.Str("sub_flag"),
		CPR:       row.FloatPtr("cpr"),
		RDR:       row.FloatPtr("rdr"),
		SCA:       row.FloatPtr("sca"),
		Exchange:  row.Str("exchange"),
	}, true
}

func etfSZConsFromRow(row Row) (domain.ETFSZCons, bool) {
	tsCode := row.Str("ts_code")
	conCode := row.Str("con_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" || conCode == "" {
		return domain.ETFSZCons{}, false
	}
	return domain.ETFSZCons{
		TradeDate: d,
		TSCode:    tsCode,
		ConCode:   conCode,
		ConName:   row.Str("con_name"),
		Qty:       row.Int64Ptr("qty"),
		SubFlag:   row.Str("sub_flag"),
		CPR:       row.FloatPtr("cpr"),
		RDR:       row.FloatPtr("rdr"),
		SubCC:     row.FloatPtr("sub_cc"),
		RedCC:     row.FloatPtr("red_cc"),
		Exchange:  row.Str("exchange"),
	}, true
}
