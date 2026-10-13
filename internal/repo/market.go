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
		"open", "high", "low", "close", "pre_close",
		"volume", "amount", "turnover", "high_limit", "low_limit",
		"is_limit_up", "is_limit_down", "is_paused",
	})
}

func (r *MarketRepo) UpsertAdjFactors(ctx context.Context, rows []domain.AdjFactor) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"adj_factor",
	})
}

func (r *MarketRepo) UpsertSuspensions(ctx context.Context, rows []domain.Suspension) error {
	return upsert(ctx, r.db, &rows, []string{"ts_code", "trade_date"}, []string{
		"suspend_type", "reason",
	})
}

func (r *MarketRepo) UpsertIndexConstituents(ctx context.Context, rows []domain.IndexConstituent) error {
	return upsert(ctx, r.db, &rows, []string{"index_code", "ts_code", "trade_date"}, []string{
		"weight",
	})
}
