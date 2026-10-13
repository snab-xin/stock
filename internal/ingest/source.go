package ingest

import (
	"context"
	"time"

	"github.com/stock/ashare/internal/domain"
)

// BarQuery 描述一段日线类数据。TSCode 为空表示全市场；TradeDate 非零时按交易日取截面。
type BarQuery struct {
	TSCode    string
	TradeDate time.Time
	Start     time.Time
	End       time.Time
}

type MarketSource interface {
	ListInstruments(ctx context.Context) ([]domain.Instrument, error)
	ListCalendar(ctx context.Context, exchange string, start, end time.Time) ([]domain.TradeCalendar, error)
	ListDailyBars(ctx context.Context, q BarQuery) ([]domain.DailyBar, error)
	ListAdjFactors(ctx context.Context, q BarQuery) ([]domain.AdjFactor, error)
	ListSuspensions(ctx context.Context, q BarQuery) ([]domain.Suspension, error)
}
