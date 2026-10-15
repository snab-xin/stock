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

// CBHolderQuery 可转债十大持有人查询。Period 与 Start/End 二选一（或同时给 Start/End 做区间）。
type CBHolderQuery struct {
	TSCodes []string
	Period  time.Time
	Start   time.Time
	End     time.Time
}

type MarketSource interface {
	ListInstruments(ctx context.Context) ([]domain.Instrument, error)
	ListCalendar(ctx context.Context, exchange string, start, end time.Time) ([]domain.TradeCalendar, error)
	ListDailyBars(ctx context.Context, q BarQuery) ([]domain.DailyBar, error)
	ListAdjFactors(ctx context.Context, q BarQuery) ([]domain.AdjFactor, error)
	ListSuspensions(ctx context.Context, q BarQuery) ([]domain.Suspension, error)
	ListCBBasic(ctx context.Context) ([]domain.CBBasic, error)
	ListCBTopHolders(ctx context.Context, q CBHolderQuery) ([]domain.CBTopHolder, error)
	ListCBDaily(ctx context.Context, q CBDailyQuery) ([]domain.CBDaily, error)
}

// CBDailyQuery 可转债日线。TradeDate 非空按截面；否则用 Start/End，可带 TSCode。
type CBDailyQuery struct {
	TSCode    string
	TradeDate time.Time
	Start     time.Time
	End       time.Time
}
