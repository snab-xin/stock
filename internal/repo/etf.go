package repo

import (
	"context"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type ETFRepo struct {
	db *gorm.DB
}

func NewETFRepo(db *gorm.DB) *ETFRepo {
	return &ETFRepo{db: db}
}

func (r *ETFRepo) UpsertBasic(ctx context.Context, items []domain.ETFBasic) error {
	return upsert(ctx, r.db, &items, []string{"ts_code"}, []string{
		"csname", "extname", "cname", "index_code", "index_name",
		"setup_date", "list_date", "list_status", "exchange",
		"mgr_name", "custod_name", "mgt_fee", "etf_type",
	})
}

func (r *ETFRepo) UpsertDaily(ctx context.Context, items []domain.ETFDaily) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"open", "high", "low", "close", "pre_close", "change", "pct_chg", "vol", "amount",
	})
}

func (r *ETFRepo) UpsertAdj(ctx context.Context, items []domain.ETFAdj) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"adj_factor",
	})
}

func (r *ETFRepo) UpsertShareSize(ctx context.Context, items []domain.ETFShareSize) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"etf_name", "total_share", "total_size", "nav", "close", "exchange",
	})
}

func (r *ETFRepo) UpsertSHCons(ctx context.Context, items []domain.ETFSHCons) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date", "con_code"}, []string{
		"con_name", "qty", "sub_flag", "cpr", "rdr", "sca", "exchange",
	})
}

func (r *ETFRepo) UpsertSZCons(ctx context.Context, items []domain.ETFSZCons) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date", "con_code"}, []string{
		"con_name", "qty", "sub_flag", "cpr", "rdr", "sub_cc", "red_cc", "exchange",
	})
}

func (r *ETFRepo) ListTSCodes(ctx context.Context) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&domain.ETFBasic{}).
		Order("ts_code").
		Pluck("ts_code", &codes).Error
	return codes, err
}
