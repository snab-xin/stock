package repo

import (
	"context"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type MarketRepo struct {
	db *gorm.DB
}

func NewMarketRepo(db *gorm.DB) *MarketRepo {
	return &MarketRepo{db: db}
}

func (r *MarketRepo) UpsertCalendar(ctx context.Context, rows []domain.TradeCalendar) error {
	return upsert(ctx, r.db, &rows, []string{"exchange", "cal_date"}, []string{
		"is_open", "pretrade_date",
	})
}

func (r *MarketRepo) UpsertDailyBars(ctx context.Context, rows []domain.DailyBar) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"open", "high", "low", "close", "pre_close", "change", "pct_chg",
		"volume", "amount", "ah_vol", "ah_amount", "turnover",
		"high_limit", "low_limit", "is_limit_up", "is_limit_down", "is_paused",
	})
}

func (r *MarketRepo) UpsertStkLimits(ctx context.Context, rows []domain.StkLimit) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"pre_close", "up_limit", "down_limit", "asset_type", "exchange",
	})
}

func (r *MarketRepo) UpsertDailyBasics(ctx context.Context, rows []domain.DailyBasic) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"close", "turnover_rate", "turnover_rate_f", "volume_ratio",
		"pe", "pe_ttm", "pb", "ps", "ps_ttm", "dv_ratio", "dv_ttm",
		"total_share", "float_share", "free_share", "total_mv", "circ_mv", "limit_status",
	})
}

func (r *MarketRepo) UpsertAdjFactors(ctx context.Context, rows []domain.AdjFactor) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"adj_factor",
	})
}

func (r *MarketRepo) UpsertSuspensions(ctx context.Context, rows []domain.Suspension) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"suspend_timing", "suspend_type",
	})
}
