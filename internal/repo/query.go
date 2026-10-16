package repo

import (
	"context"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

func (r *InstrumentRepo) GetByTSCode(ctx context.Context, tsCode string) (domain.Instrument, error) {
	var items []domain.Instrument
	if err := r.db.WithContext(ctx).Select(instrumentColumns).Where("ts_code = ?", tsCode).Limit(1).Find(&items).Error; err != nil {
		return domain.Instrument{}, err
	}
	if len(items) == 0 {
		return domain.Instrument{}, gorm.ErrRecordNotFound
	}
	return items[0], nil
}

func (r *InstrumentRepo) FindBySymbol(ctx context.Context, symbol string) ([]domain.Instrument, error) {
	var items []domain.Instrument
	err := r.db.WithContext(ctx).Select(instrumentColumns).Where("symbol = ?", symbol).Find(&items).Error
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
		Select(instrumentColumns).
		Where("ts_code ILIKE ? OR symbol ILIKE ? OR name ILIKE ?", like, like, like).
		Order("ts_code").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *MarketRepo) ListOpenDays(ctx context.Context, exchange string, start, end time.Time) ([]time.Time, error) {
	var rows []domain.TradeCalendar
	q := r.db.WithContext(ctx).Select("cal_date").Where("exchange = ? AND is_open = ?", exchange, true)
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

// BarPeriod 日 / 周 / 月；周月来自库内聚合视图。
type BarPeriod string

const (
	PeriodDay   BarPeriod = "day"
	PeriodWeek  BarPeriod = "week"
	PeriodMonth BarPeriod = "month"
)

func NormalizeBarPeriod(s string) BarPeriod {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "week", "w", "1w":
		return PeriodWeek
	case "month", "m", "1m":
		return PeriodMonth
	default:
		return PeriodDay
	}
}

func barTable(qfq bool, period BarPeriod) string {
	switch period {
	case PeriodWeek:
		if qfq {
			return "daily_bar_week_qfq"
		}
		return "daily_bar_week"
	case PeriodMonth:
		if qfq {
			return "daily_bar_month_qfq"
		}
		return "daily_bar_month"
	default:
		if qfq {
			return "daily_bar_qfq"
		}
		return "daily_bar"
	}
}

func (r *MarketRepo) ListDailyBars(ctx context.Context, tsCode string, start, end time.Time, limit int, qfq bool, period BarPeriod) ([]domain.DailyBar, error) {
	table := barTable(qfq, period)
	q := r.db.WithContext(ctx).Table(table).Select(dailyBarColumns).Where("ts_code = ?", tsCode)
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

// ListDailyBarsBefore 取 before（不含）之前最近 limit 根，按日期升序返回。
func (r *MarketRepo) ListDailyBarsBefore(ctx context.Context, tsCode string, before time.Time, limit int, qfq bool, period BarPeriod) ([]domain.DailyBar, error) {
	if before.IsZero() {
		return nil, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 250
	}
	table := barTable(qfq, period)
	var rows []domain.DailyBar
	err := r.db.WithContext(ctx).Table(table).
		Select(dailyBarColumns).
		Where("ts_code = ? AND trade_date < ?", tsCode, before).
		Order("trade_date DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	reverseBars(rows)
	return rows, nil
}

func reverseBars(rows []domain.DailyBar) {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
}
