package repo

// 查询时显式列名，避免 SELECT *。

var instrumentColumns = []string{
	"id", "ts_code", "symbol", "name", "area", "industry", "fullname", "enname",
	"cnspell", "market", "exchange", "curr_type", "list_status", "list_date",
	"delist_date", "is_hs", "act_name", "act_ent_type", "asset_type", "is_st",
	"created_at", "updated_at",
}

const instrumentSQLCols = `i.id, i.ts_code, i.symbol, i.name, i.area, i.industry, i.fullname, i.enname,
	i.cnspell, i.market, i.exchange, i.curr_type, i.list_status, i.list_date,
	i.delist_date, i.is_hs, i.act_name, i.act_ent_type, i.asset_type, i.is_st,
	i.created_at, i.updated_at`

var dailyBarColumns = []string{
	"ts_code", "trade_date", "open", "high", "low", "close", "pre_close",
	"change", "pct_chg", "volume", "amount", "ah_vol", "ah_amount", "turnover",
	"high_limit", "low_limit", "is_limit_up", "is_limit_down", "is_paused",
}

var calendarColumns = []string{
	"id", "exchange", "cal_date", "is_open", "pretrade_date", "created_at", "updated_at",
}

var ingestRunColumns = []string{
	"id", "job_name", "started_at", "finished_at", "status", "message",
	"rows_affected", "params", "created_at", "updated_at",
}

const cbBasicSQLCols = `b.id, b.ts_code, b.bond_full_name, b.bond_short_name, b.cb_code, b.cb_type,
	b.stk_code, b.stk_short_name, b.maturity, b.par, b.issue_price, b.issue_size, b.remain_size,
	b.value_date, b.maturity_date, b.rate_type, b.coupon_rate, b.add_rate, b.pay_per_year,
	b.list_date, b.delist_date, b.exchange, b.conv_start_date, b.conv_end_date, b.conv_stop_date,
	b.first_conv_price, b.conv_price, b.rate_clause, b.put_clause, b.maturity_call_price,
	b.call_clause, b.reset_clause, b.conv_clause, b.guarantor, b.guarantee_type,
	b.issue_rating, b.newest_rating, b.rating_comp, b.created_at, b.updated_at`

var cbBasicColumns = []string{
	"id", "ts_code", "bond_full_name", "bond_short_name", "cb_code", "cb_type",
	"stk_code", "stk_short_name", "maturity", "par", "issue_price", "issue_size", "remain_size",
	"value_date", "maturity_date", "rate_type", "coupon_rate", "add_rate", "pay_per_year",
	"list_date", "delist_date", "exchange", "conv_start_date", "conv_end_date", "conv_stop_date",
	"first_conv_price", "conv_price", "rate_clause", "put_clause", "maturity_call_price",
	"call_clause", "reset_clause", "conv_clause", "guarantor", "guarantee_type",
	"issue_rating", "newest_rating", "rating_comp", "created_at", "updated_at",
}

var cbHolderColumns = []string{
	"id", "ts_code", "end_date", "holder_rank", "holder_name", "hold_amount", "hold_ratio", "type",
	"created_at", "updated_at",
}

var userColumns = []string{
	"id", "username", "password_hash", "email", "phone", "role", "status", "created_at", "updated_at",
}

var noteColumns = []string{
	"id", "user_id", "ts_code", "target_type", "trade_date", "summary", "content", "created_at", "updated_at",
}
