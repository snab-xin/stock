package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/repo"
)

type Service struct {
	Source      MarketSource
	Instruments *repo.InstrumentRepo
	Market      *repo.MarketRepo
	CB          *repo.CBRepo
	Runs        *repo.IngestRunRepo
	log         *slog.Logger
}

func (s *Service) logger() *slog.Logger {
	if s.log != nil {
		return s.log
	}
	s.log = applog.With("component", "ingest")
	return s.log
}

func (s *Service) withRun(ctx context.Context, job string, fn func() (int64, error)) error {
	run, err := s.Runs.Start(ctx, job)
	if err != nil {
		return err
	}
	n, err := fn()
	status, msg := "success", ""
	if err != nil {
		status = "failed"
		msg = err.Error()
		_ = s.Runs.Finish(ctx, run, status, msg, n)
		return err
	}
	return s.Runs.Finish(ctx, run, status, msg, n)
}

func (s *Service) SyncInstruments(ctx context.Context) error {
	return s.withRun(ctx, "instruments", func() (int64, error) {
		items, err := s.Source.ListInstruments(ctx)
		if err != nil {
			return 0, err
		}
		if err := s.Instruments.Upsert(ctx, items); err != nil {
			return 0, err
		}
		s.logger().Info("instruments upserted", "rows", len(items))
		return int64(len(items)), nil
	})
}

func (s *Service) SyncCalendar(ctx context.Context, start, end time.Time) error {
	return s.withRun(ctx, "calendar", func() (int64, error) {
		var total int64
		var sse []domain.TradeCalendar
		for _, ex := range []string{"SSE", "SZSE"} {
			rows, err := s.Source.ListCalendar(ctx, ex, start, end)
			if err != nil {
				return total, err
			}
			if err := s.Market.UpsertCalendar(ctx, rows); err != nil {
				return total, err
			}
			s.logger().Info("calendar upserted", "exchange", ex, "rows", len(rows))
			total += int64(len(rows))
			if ex == "SSE" {
				sse = rows
			}
		}
		if len(sse) > 0 {
			bse := copyCalendar(sse, "BSE")
			if err := s.Market.UpsertCalendar(ctx, bse); err != nil {
				return total, err
			}
			s.logger().Info("calendar upserted", "exchange", "BSE", "rows", len(bse), "copied_from", "SSE")
			total += int64(len(bse))
		}
		return total, nil
	})
}

func (s *Service) SyncDaily(ctx context.Context, start, end time.Time, codes []string) error {
	return s.withRun(ctx, "daily", func() (int64, error) {
		if start.IsZero() || end.IsZero() {
			return 0, fmt.Errorf("ingest daily 需要 -start 和 -end")
		}
		if end.Before(start) {
			return 0, fmt.Errorf("结束日期早于开始日期")
		}

		var total int64
		if len(codes) > 0 {
			for _, code := range codes {
				n, err := s.syncCodeRange(ctx, code, start, end)
				if err != nil {
					return total, err
				}
				total += n
			}
			return total, nil
		}

		days, err := s.Market.ListOpenDays(ctx, "SSE", start, end)
		if err != nil {
			return 0, err
		}
		if len(days) == 0 {
			return 0, fmt.Errorf("区间内没有开市日，请先 ingest calendar")
		}

		susp, err := s.Source.ListSuspensions(ctx, BarQuery{Start: start, End: end})
		if err != nil {
			s.logger().Warn("suspension skipped", "err", err)
			susp = nil
		} else if err := s.Market.UpsertSuspensions(ctx, susp); err != nil {
			return 0, err
		}

		for _, d := range days {
			q := BarQuery{TradeDate: d}
			bars, err := s.Source.ListDailyBars(ctx, q)
			if err != nil {
				return total, fmt.Errorf("%s 日线: %w", dates.ISO(d), err)
			}
			MarkPaused(bars, susp)
			if err := s.Market.UpsertDailyBars(ctx, bars); err != nil {
				return total, err
			}
			adjs, err := s.Source.ListAdjFactors(ctx, q)
			if err != nil {
				return total, fmt.Errorf("%s 复权因子: %w", dates.ISO(d), err)
			}
			if err := s.Market.UpsertAdjFactors(ctx, adjs); err != nil {
				return total, err
			}
			s.logger().Info("daily upserted",
				"trade_date", dates.ISO(d),
				"bars", len(bars),
				"adj", len(adjs),
			)
			total += int64(len(bars) + len(adjs))
		}
		return total, nil
	})
}

func (s *Service) SyncAll(ctx context.Context, start, end time.Time, codes []string) error {
	if err := s.SyncCalendar(ctx, start, end); err != nil {
		return err
	}
	if err := s.SyncInstruments(ctx); err != nil {
		return err
	}
	return s.SyncDaily(ctx, start, end, codes)
}

func (s *Service) syncCodeRange(ctx context.Context, code string, start, end time.Time) (int64, error) {
	q := BarQuery{TSCode: code, Start: start, End: end}
	bars, err := s.Source.ListDailyBars(ctx, q)
	if err != nil {
		return 0, err
	}
	susp, err := s.Source.ListSuspensions(ctx, q)
	if err != nil {
		susp = nil
	} else if err := s.Market.UpsertSuspensions(ctx, susp); err != nil {
		return 0, err
	}
	MarkPaused(bars, susp)
	if err := s.Market.UpsertDailyBars(ctx, bars); err != nil {
		return 0, err
	}
	adjs, err := s.Source.ListAdjFactors(ctx, q)
	if err != nil {
		return 0, err
	}
	if err := s.Market.UpsertAdjFactors(ctx, adjs); err != nil {
		return 0, err
	}
	s.logger().Info("daily upserted",
		"ts_code", code,
		"bars", len(bars),
		"adj", len(adjs),
	)
	return int64(len(bars) + len(adjs)), nil
}

func (s *Service) SyncCBBasic(ctx context.Context) error {
	if s.CB == nil {
		return fmt.Errorf("未配置可转债仓储")
	}
	return s.withRun(ctx, "cb_basic", func() (int64, error) {
		items, err := s.Source.ListCBBasic(ctx)
		if err != nil {
			return 0, err
		}
		if err := s.CB.UpsertBasic(ctx, items); err != nil {
			return 0, err
		}
		s.logger().Info("cb_basic upserted", "rows", len(items))
		return int64(len(items)), nil
	})
}

func (s *Service) SyncCBDaily(ctx context.Context, start, end, tradeDate time.Time, codes []string) error {
	if s.CB == nil {
		return fmt.Errorf("未配置可转债仓储")
	}
	return s.withRun(ctx, "cb_daily", func() (int64, error) {
		var total int64
		if !tradeDate.IsZero() {
			rows, err := s.Source.ListCBDaily(ctx, CBDailyQuery{TradeDate: tradeDate})
			if err != nil {
				return 0, err
			}
			if err := s.CB.UpsertDaily(ctx, rows); err != nil {
				return 0, err
			}
			s.logger().Info("cb_daily upserted", "trade_date", dates.ISO(tradeDate), "rows", len(rows))
			return int64(len(rows)), nil
		}
		if start.IsZero() || end.IsZero() {
			return 0, fmt.Errorf("cb_daily 需要 trade_date，或 start 与 end")
		}
		if len(codes) == 0 {
			// 按交易日截面拉全市场（与股票 daily 类似）
			for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
				rows, err := s.Source.ListCBDaily(ctx, CBDailyQuery{TradeDate: d})
				if err != nil {
					return total, err
				}
				if len(rows) == 0 {
					continue
				}
				if err := s.CB.UpsertDaily(ctx, rows); err != nil {
					return total, err
				}
				total += int64(len(rows))
				s.logger().Info("cb_daily upserted", "trade_date", dates.ISO(d), "rows", len(rows))
			}
			return total, nil
		}
		for _, code := range codes {
			rows, err := s.Source.ListCBDaily(ctx, CBDailyQuery{TSCode: code, Start: start, End: end})
			if err != nil {
				return total, err
			}
			if err := s.CB.UpsertDaily(ctx, rows); err != nil {
				return total, err
			}
			total += int64(len(rows))
			s.logger().Info("cb_daily upserted", "ts_code", code, "rows", len(rows))
		}
		return total, nil
	})
}

func (s *Service) SyncCBTopHolders(ctx context.Context, period, start, end time.Time, codes []string) error {
	if s.CB == nil {
		return fmt.Errorf("未配置可转债仓储")
	}
	return s.withRun(ctx, "cb_holders", func() (int64, error) {
		if len(codes) == 0 {
			var err error
			codes, err = s.CB.ListTSCodes(ctx)
			if err != nil {
				return 0, err
			}
			if len(codes) == 0 {
				return 0, fmt.Errorf("库中没有可转债，请先 ingest cb-basic；未指定 -code 时按库内全部转债同步")
			}
			s.logger().Info("cb_holders using all bonds in db", "count", len(codes))
		}
		const batch = 20
		var total int64
		for i := 0; i < len(codes); i += batch {
			j := i + batch
			if j > len(codes) {
				j = len(codes)
			}
			rows, err := s.Source.ListCBTopHolders(ctx, CBHolderQuery{
				TSCodes: codes[i:j],
				Period:  period,
				Start:   start,
				End:     end,
			})
			if err != nil {
				return total, err
			}
			if err := s.CB.UpsertTopHolders(ctx, rows); err != nil {
				return total, err
			}
			total += int64(len(rows))
			s.logger().Info("cb_holders upserted",
				"batch_from", codes[i],
				"batch_size", j-i,
				"rows", len(rows),
			)
		}
		return total, nil
	})
}

func copyCalendar(rows []domain.TradeCalendar, exchange string) []domain.TradeCalendar {
	out := make([]domain.TradeCalendar, len(rows))
	for i, row := range rows {
		row.ID = 0
		row.CreatedAt = time.Time{}
		row.UpdatedAt = time.Time{}
		row.Exchange = exchange
		out[i] = row
	}
	return out
}
