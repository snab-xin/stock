package repo

import (
	"context"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type IndexRepo struct {
	db *gorm.DB
}

func NewIndexRepo(db *gorm.DB) *IndexRepo {
	return &IndexRepo{db: db}
}

func (r *IndexRepo) UpsertBasic(ctx context.Context, items []domain.IndexBasic) error {
	return upsert(ctx, r.db, &items, []string{"ts_code"}, []string{
		"name", "fullname", "market", "publisher", "index_type", "category",
		"base_date", "base_point", "list_date", "weight_rule", "description", "exp_date",
	})
}

func (r *IndexRepo) UpsertDaily(ctx context.Context, items []domain.IndexDaily) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"close", "open", "high", "low", "pre_close", "change", "pct_chg", "vol", "amount",
	})
}

func (r *IndexRepo) UpsertClassify(ctx context.Context, items []domain.IndexClassify) error {
	return upsert(ctx, r.db, &items, []string{"index_code", "src"}, []string{
		"industry_name", "parent_code", "level", "industry_code", "is_pub",
	})
}

func (r *IndexRepo) UpsertSWDaily(ctx context.Context, items []domain.SWDaily) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"name", "open", "low", "high", "close", "change", "pct_change",
		"vol", "amount", "pe", "pb", "float_mv", "total_mv",
	})
}

func (r *IndexRepo) ListTSCodes(ctx context.Context) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&domain.IndexBasic{}).
		Order("ts_code").
		Pluck("ts_code", &codes).Error
	return codes, err
}
