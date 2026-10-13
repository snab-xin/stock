package repo

import (
	"context"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type InstrumentRepo struct {
	db *gorm.DB
}

func NewInstrumentRepo(db *gorm.DB) *InstrumentRepo {
	return &InstrumentRepo{db: db}
}

func (r *InstrumentRepo) Upsert(ctx context.Context, items []domain.Instrument) error {
	return upsert(ctx, r.db, &items, []string{"ts_code"}, []string{
		"symbol", "name", "exchange", "asset_type",
		"list_date", "delist_date", "is_st", "status",
	})
}
