package tushare

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/ingest"
)

const cbBasicFields = "ts_code,bond_full_name,bond_short_name,cb_code,cb_type,stk_code,stk_short_name," +
	"maturity,par,issue_price,issue_size,remain_size,value_date,maturity_date,rate_type,coupon_rate,add_rate," +
	"pay_per_year,list_date,delist_date,exchange,conv_start_date,conv_end_date,conv_stop_date," +
	"first_conv_price,conv_price,rate_clause,put_clause,maturity_call_price,call_clause,reset_clause," +
	"conv_clause,guarantor,guarantee_type,issue_rating,newest_rating,rating_comp"

const cbHolderFields = "ts_code,end_date,holder_rank,holder_name,hold_amount,hold_ratio"

const cbDailyFields = "ts_code,trade_date,pre_close,open,high,low,close,change,pct_chg,vol,amount," +
	"bond_value,bond_over_rate,cb_value,cb_over_rate"

func (s *Source) ListCBBasic(ctx context.Context) ([]domain.CBBasic, error) {
	const page = 2000
	var all []domain.CBBasic
	for offset := 0; ; offset += page {
		rows, err := s.client.Query(ctx, "cb_basic", map[string]string{
			"offset": strconv.Itoa(offset),
			"limit":  strconv.Itoa(page),
		}, cbBasicFields)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, ok := cbBasicFromRow(row)
			if ok {
				all = append(all, it)
			}
		}
		if len(rows) < page {
			break
		}
	}
	return all, nil
}

func (s *Source) ListCBTopHolders(ctx context.Context, q ingest.CBHolderQuery) ([]domain.CBTopHolder, error) {
	if len(q.TSCodes) == 0 {
		return nil, fmt.Errorf("top10_cb_holders 需要至少一只转债代码")
	}
	params := map[string]string{
		"ts_code": strings.Join(q.TSCodes, ","),
	}
	if !q.Period.IsZero() {
		params["period"] = dates.YYYYMMDD(q.Period)
	}
	if !q.Start.IsZero() {
		params["start_date"] = dates.YYYYMMDD(q.Start)
	}
	if !q.End.IsZero() {
		params["end_date"] = dates.YYYYMMDD(q.End)
	}
	if q.Period.IsZero() && q.Start.IsZero() && q.End.IsZero() {
		return nil, fmt.Errorf("top10_cb_holders 需要 period 或 start/end")
	}

	rows, err := s.client.Query(ctx, "top10_cb_holders", params, cbHolderFields)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CBTopHolder, 0, len(rows))
	for _, row := range rows {
		it, ok := cbTopHolderFromRow(row)
		if ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Source) ListCBDaily(ctx context.Context, q ingest.CBDailyQuery) ([]domain.CBDaily, error) {
	params := map[string]string{}
	if q.TSCode != "" {
		params["ts_code"] = q.TSCode
	}
	if !q.TradeDate.IsZero() {
		params["trade_date"] = dates.YYYYMMDD(q.TradeDate)
	}
	if !q.Start.IsZero() {
		params["start_date"] = dates.YYYYMMDD(q.Start)
	}
	if !q.End.IsZero() {
		params["end_date"] = dates.YYYYMMDD(q.End)
	}
	if q.TradeDate.IsZero() && q.Start.IsZero() && q.End.IsZero() {
		return nil, fmt.Errorf("cb_daily 需要 trade_date 或 start/end")
	}

	const page = 2000
	var all []domain.CBDaily
	for offset := 0; ; offset += page {
		p := map[string]string{}
		for k, v := range params {
			p[k] = v
		}
		p["offset"] = strconv.Itoa(offset)
		p["limit"] = strconv.Itoa(page)
		rows, err := s.client.Query(ctx, "cb_daily", p, cbDailyFields)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			it, ok := cbDailyFromRow(row)
			if ok {
				all = append(all, it)
			}
		}
		if len(rows) < page {
			break
		}
	}
	return all, nil
}

func cbDailyFromRow(row Row) (domain.CBDaily, bool) {
	tsCode := row.Str("ts_code")
	d, err := row.Date("trade_date")
	if err != nil || tsCode == "" {
		return domain.CBDaily{}, false
	}
	return domain.CBDaily{
		TSCode:       tsCode,
		TradeDate:    d,
		PreClose:     row.FloatPtr("pre_close"),
		Open:         row.FloatPtr("open"),
		High:         row.FloatPtr("high"),
		Low:          row.FloatPtr("low"),
		Close:        row.FloatPtr("close"),
		Change:       row.FloatPtr("change"),
		PctChg:       row.FloatPtr("pct_chg"),
		Volume:       row.FloatPtr("vol"),
		Amount:       row.FloatPtr("amount"),
		BondValue:    row.FloatPtr("bond_value"),
		BondOverRate: row.FloatPtr("bond_over_rate"),
		CBValue:      row.FloatPtr("cb_value"),
		CBOverRate:   row.FloatPtr("cb_over_rate"),
	}, true
}

func cbBasicFromRow(row Row) (domain.CBBasic, bool) {
	tsCode := row.Str("ts_code")
	if tsCode == "" {
		return domain.CBBasic{}, false
	}
	ex := row.Str("exchange")
	if ex == "" {
		ex = exchangeFromTSCode(tsCode)
	}
	return domain.CBBasic{
		TSCode:            tsCode,
		BondFullName:      row.Str("bond_full_name"),
		BondShortName:     row.Str("bond_short_name"),
		CBCode:            row.Str("cb_code"),
		CBType:            row.Str("cb_type"),
		StkCode:           row.Str("stk_code"),
		StkShortName:      row.Str("stk_short_name"),
		Maturity:          row.FloatPtr("maturity"),
		Par:               row.FloatPtr("par"),
		IssuePrice:        row.FloatPtr("issue_price"),
		IssueSize:         row.FloatPtr("issue_size"),
		RemainSize:        row.FloatPtr("remain_size"),
		ValueDate:         row.DatePtr("value_date"),
		MaturityDate:      row.DatePtr("maturity_date"),
		RateType:          row.Str("rate_type"),
		CouponRate:        row.FloatPtr("coupon_rate"),
		AddRate:           row.FloatPtr("add_rate"),
		PayPerYear:        row.IntPtr("pay_per_year"),
		ListDate:          row.DatePtr("list_date"),
		DelistDate:        row.DatePtr("delist_date"),
		Exchange:          ex,
		ConvStartDate:     row.DatePtr("conv_start_date"),
		ConvEndDate:       row.DatePtr("conv_end_date"),
		ConvStopDate:      row.DatePtr("conv_stop_date"),
		FirstConvPrice:    row.FloatPtr("first_conv_price"),
		ConvPrice:         row.FloatPtr("conv_price"),
		RateClause:        row.Str("rate_clause"),
		PutClause:         row.Str("put_clause"),
		MaturityCallPrice: row.FloatPtr("maturity_call_price"),
		CallClause:        row.Str("call_clause"),
		ResetClause:       row.Str("reset_clause"),
		ConvClause:        row.Str("conv_clause"),
		Guarantor:         row.Str("guarantor"),
		GuaranteeType:     row.Str("guarantee_type"),
		IssueRating:       row.Str("issue_rating"),
		NewestRating:      row.Str("newest_rating"),
		RatingComp:        row.Str("rating_comp"),
	}, true
}

func cbTopHolderFromRow(row Row) (domain.CBTopHolder, bool) {
	tsCode := row.Str("ts_code")
	end, err := row.Date("end_date")
	rank, ok := row.Int("holder_rank")
	if err != nil || tsCode == "" || !ok {
		return domain.CBTopHolder{}, false
	}
	return domain.CBTopHolder{
		TSCode:     tsCode,
		EndDate:    end,
		HolderRank: rank,
		HolderName: row.Str("holder_name"),
		HoldAmount: row.FloatPtr("hold_amount"),
		HoldRatio:  row.FloatPtr("hold_ratio"),
	}, true
}
