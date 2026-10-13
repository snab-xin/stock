package repo

import (
	"context"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
)

func (r *InstrumentRepo) GetByTSCode(ctx context.Context, tsCode string) (domain.Instrument, error) {
	var it domain.Instrument
	err := r.db.WithContext(ctx).Where("ts_code = ?", tsCode).Take(&it).Error
	return it, err
}

func (r *InstrumentRepo) FindBySymbol(ctx context.Context, symbol string) ([]domain.Instrument, error) {
	var items []domain.Instrument
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Find(&items).Error
	return items, err
}

func (r *InstrumentRepo) Search(ctx context.Context, q string, limit int) ([]domain.Instrument, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.Instrument{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	like := "%" + q + "%"
	var items []domain.Instrument
	err := r.db.WithContext(ctx).
		Where("ts_code ILIKE ? OR symbol ILIKE ? OR name ILIKE ?", like, like, like).
		Order("ts_code").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *MarketRepo) ListOpenDays(ctx context.Context, exchange string, start, end time.Time) ([]time.Time, error) {
	var rows []domain.TradeCalendar
	q := r.db.WithContext(ctx).Where("exchange = ? AND is_open = ?", exchange, true)
	if !start.IsZero() {
		q = q.Where("cal_date >= ?", start)
	}
	if !end.IsZero() {
		q = q.Where("cal_date <= ?", end)
	}
	if err := q.Order("cal_date").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.CalDate)
	}
	return out, nil
}

func (r *MarketRepo) ListDailyBars(ctx context.Context, tsCode string, start, end time.Time, limit int, qfq bool) ([]domain.DailyBar, error) {
	table := "daily_bar"
	if qfq {
		table = "daily_bar_qfq"
	}
	q := r.db.WithContext(ctx).Table(table).Where("ts_code = ?", tsCode)
	if !start.IsZero() {
		q = q.Where("trade_date >= ?", start)
	}
	if !end.IsZero() {
		q = q.Where("trade_date <= ?", end)
	}
	var rows []domain.DailyBar
	if limit > 0 && start.IsZero() && end.IsZero() {
		if err := q.Order("trade_date DESC").Limit(limit).Find(&rows).Error; err != nil {
			return nil, err
		}
		reverseBars(rows)
		return rows, nil
	}
	if err := q.Order("trade_date").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func reverseBars(rows []domain.DailyBar) {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
}
