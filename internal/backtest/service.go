package backtest

import (
	"context"
	"fmt"

	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
)

func LoadAndRun(ctx context.Context, instruments *repo.InstrumentRepo, market *repo.MarketRepo, input string, p Params) (Result, error) {
	p = p.withDefaults()
	code, err := ingest.ResolveTSCode(ctx, instruments, input)
	if err != nil {
		return Result{}, err
	}
	p.Code = code

	fetchStart := p.TradeFrom
	if !p.TradeFrom.IsZero() {
		fetchStart = p.TradeFrom.AddDate(0, 0, -(p.Slow*3 + 20))
	}

	raw, err := market.ListDailyBars(ctx, code, fetchStart, p.TradeTo, 0, false)
	if err != nil {
		return Result{}, err
	}
	qfq, err := market.ListDailyBars(ctx, code, fetchStart, p.TradeTo, 0, true)
	if err != nil {
		return Result{}, err
	}
	if len(raw) == 0 {
		return Result{}, fmt.Errorf("%s 没有未复权日线，请先 ingest daily", code)
	}
	if len(qfq) == 0 {
		return Result{}, fmt.Errorf("%s 没有前复权日线（需要 daily_bar 与 adj_factor）", code)
	}

	res, err := Run(raw, qfq, p)
	if err != nil {
		return Result{}, err
	}
	if inst, err := instruments.GetByTSCode(ctx, code); err == nil {
		res.Name = inst.Name
	}
	return res, nil
}
