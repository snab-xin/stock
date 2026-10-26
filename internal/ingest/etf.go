package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/stock/ashare/internal/pkg/dates"
	"github.com/stock/ashare/internal/domain"
)

func (s *Service) SyncETFBasic(ctx context.Context) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_basic", runArgs{}, func() (int64, error) {
		items, err := s.Source.ListETFBasic(ctx)
		if err != nil {
			return 0, err
		}
		if err := s.ETF.UpsertBasic(ctx, items); err != nil {
			return 0, err
		}
		if s.Instruments != nil {
			inst := make([]domain.Instrument, 0, len(items))
			for _, it := range items {
				inst = append(inst, instrumentFromETF(it))
			}
			if err := s.Instruments.Upsert(ctx, inst); err != nil {
				return 0, err
			}
		}
		s.logger().Info("etf_basic upserted", "rows", len(items))
		return int64(len(items)), nil
	})
}

func (s *Service) SyncETFDaily(ctx context.Context, start, end time.Time, codes []string) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_daily", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		return s.syncETFByDay(ctx, start, end, codes, "etf_daily",
			func(ctx context.Context, q BarQuery) (int, error) {
				rows, err := s.Source.ListETFDaily(ctx, q)
				if err != nil {
					return 0, err
				}
				if err := s.ETF.UpsertDaily(ctx, rows); err != nil {
					return 0, err
				}
				return len(rows), nil
			})
	})
}

func (s *Service) SyncETFAdj(ctx context.Context, start, end time.Time, codes []string) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_adj", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		return s.syncETFByDay(ctx, start, end, codes, "etf_adj",
			func(ctx context.Context, q BarQuery) (int, error) {
				rows, err := s.Source.ListETFAdj(ctx, q)
				if err != nil {
					return 0, err
				}
				if err := s.ETF.UpsertAdj(ctx, rows); err != nil {
					return 0, err
				}
				return len(rows), nil
			})
	})
}

func (s *Service) SyncETFShareSize(ctx context.Context, start, end time.Time, codes []string) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_share_size", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		return s.syncETFByDay(ctx, start, end, codes, "etf_share_size",
			func(ctx context.Context, q BarQuery) (int, error) {
				rows, err := s.Source.ListETFShareSize(ctx, q)
				if err != nil {
					return 0, err
				}
				if err := s.ETF.UpsertShareSize(ctx, rows); err != nil {
					return 0, err
				}
				return len(rows), nil
			})
	})
}

func (s *Service) SyncETFSHCons(ctx context.Context, start, end time.Time, codes []string) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_sh_cons", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		return s.syncETFByDay(ctx, start, end, codes, "etf_sh_cons",
			func(ctx context.Context, q BarQuery) (int, error) {
				rows, err := s.Source.ListETFSHCons(ctx, q)
				if err != nil {
					return 0, err
				}
				if err := s.ETF.UpsertSHCons(ctx, rows); err != nil {
					return 0, err
				}
				return len(rows), nil
			})
	})
}

func (s *Service) SyncETFSZCons(ctx context.Context, start, end time.Time, codes []string) error {
	if s.ETF == nil {
		return fmt.Errorf("未配置 ETF 仓储")
	}
	return s.withRun(ctx, "etf_sz_cons", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		return s.syncETFByDay(ctx, start, end, codes, "etf_sz_cons",
			func(ctx context.Context, q BarQuery) (int, error) {
				rows, err := s.Source.ListETFSZCons(ctx, q)
				if err != nil {
					return 0, err
				}
				if err := s.ETF.UpsertSZCons(ctx, rows); err != nil {
					return 0, err
				}
				return len(rows), nil
			})
	})
}

func (s *Service) syncETFByDay(ctx context.Context, start, end time.Time, codes []string, job string, fn func(context.Context, BarQuery) (int, error)) (int64, error) {
	if start.IsZero() || end.IsZero() {
		return 0, fmt.Errorf("%s 需要 start 和 end", job)
	}
	if len(codes) > 0 {
		var total int64
		for _, code := range codes {
			n, err := fn(ctx, BarQuery{TSCode: code, Start: start, End: end})
			if err != nil {
				return total, err
			}
			total += int64(n)
			s.logger().Info(job+" upserted", "ts_code", code, "rows", n)
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
	var total int64
	for _, d := range days {
		n, err := fn(ctx, BarQuery{TradeDate: d})
		if err != nil {
			return total, fmt.Errorf("%s %s: %w", job, dates.ISO(d), err)
		}
		total += int64(n)
		s.logger().Info(job+" upserted", "trade_date", dates.ISO(d), "rows", n)
	}
	return total, nil
}

func instrumentFromETF(it domain.ETFBasic) domain.Instrument {
	symbol := it.TSCode
	if i := len(it.TSCode); i > 3 && it.TSCode[i-3] == '.' {
		symbol = it.TSCode[:i-3]
	}
	name := it.Extname
	if name == "" {
		name = it.Csname
	}
	ex := it.Exchange
	switch ex {
	case "SH":
		ex = "SSE"
	case "SZ":
		ex = "SZSE"
	}
	return domain.Instrument{
		TSCode:     it.TSCode,
		Symbol:     symbol,
		Name:       name,
		Fullname:   it.Cname,
		Exchange:   ex,
		ListStatus: it.ListStatus,
		ListDate:   it.ListDate,
		AssetType:  "etf",
	}
}
