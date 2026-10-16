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
		"symbol", "name", "area", "industry", "fullname", "enname", "cnspell",
		"market", "exchange", "curr_type", "list_status",
		"list_date", "delist_date", "is_hs", "act_name", "act_ent_type",
		"asset_type", "is_st",
	})
}
