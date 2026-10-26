package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/stock/ashare/internal/pkg/dates"
	"github.com/stock/ashare/internal/domain"
)

func (s *Service) SyncIndexBasic(ctx context.Context) error {
	if s.Index == nil {
		return fmt.Errorf("未配置指数仓储")
	}
	return s.withRun(ctx, "index_basic", runArgs{}, func() (int64, error) {
		items, err := s.Source.ListIndexBasic(ctx)
		if err != nil {
			return 0, err
		}
		if err := s.Index.UpsertBasic(ctx, items); err != nil {
			return 0, err
		}
		if s.Instruments != nil {
			inst := make([]domain.Instrument, 0, len(items))
			for _, it := range items {
				inst = append(inst, instrumentFromIndex(it))
			}
			if err := s.Instruments.Upsert(ctx, inst); err != nil {
				return 0, err
			}
		}
		s.logger().Info("index_basic upserted", "rows", len(items))
		return int64(len(items)), nil
	})
}

func (s *Service) SyncIndexClassify(ctx context.Context) error {
	if s.Index == nil {
		return fmt.Errorf("未配置指数仓储")
	}
	return s.withRun(ctx, "index_classify", runArgs{}, func() (int64, error) {
		items, err := s.Source.ListIndexClassify(ctx)
		if err != nil {
			return 0, err
		}
		if err := s.Index.UpsertClassify(ctx, items); err != nil {
			return 0, err
		}
		s.logger().Info("index_classify upserted", "rows", len(items))
		return int64(len(items)), nil
	})
}

func (s *Service) SyncIndexDaily(ctx context.Context, start, end time.Time, codes []string) error {
	if s.Index == nil {
		return fmt.Errorf("未配置指数仓储")
	}
	return s.withRun(ctx, "index_daily", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		if start.IsZero() || end.IsZero() {
			return 0, fmt.Errorf("index_daily 需要 start 和 end")
		}
		if len(codes) == 0 {
			var err error
			codes, err = s.Index.ListTSCodes(ctx)
			if err != nil {
				return 0, err
			}
			if len(codes) == 0 {
				return 0, fmt.Errorf("库中没有指数，请先 ingest index-basic")
			}
			s.logger().Info("index_daily using all indices in db", "count", len(codes))
		}
		var total int64
		for _, code := range codes {
			rows, err := s.Source.ListIndexDaily(ctx, BarQuery{TSCode: code, Start: start, End: end})
			if err != nil {
				return total, fmt.Errorf("%s: %w", code, err)
			}
			if err := s.Index.UpsertDaily(ctx, rows); err != nil {
				return total, err
			}
			total += int64(len(rows))
			s.logger().Info("index_daily upserted", "ts_code", code, "rows", len(rows))
		}
		return total, nil
	})
}

func (s *Service) SyncSWDaily(ctx context.Context, start, end time.Time, codes []string) error {
	if s.Index == nil {
		return fmt.Errorf("未配置指数仓储")
	}
	return s.withRun(ctx, "sw_daily", runArgs{Start: start, End: end, Codes: codes}, func() (int64, error) {
		if start.IsZero() || end.IsZero() {
			return 0, fmt.Errorf("sw_daily 需要 start 和 end")
		}
		if len(codes) > 0 {
			var total int64
			for _, code := range codes {
				rows, err := s.Source.ListSWDaily(ctx, BarQuery{TSCode: code, Start: start, End: end})
				if err != nil {
					return total, err
				}
				if err := s.Index.UpsertSWDaily(ctx, rows); err != nil {
					return total, err
				}
				total += int64(len(rows))
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
			rows, err := s.Source.ListSWDaily(ctx, BarQuery{TradeDate: d})
			if err != nil {
				return total, fmt.Errorf("%s: %w", dates.ISO(d), err)
			}
			if err := s.Index.UpsertSWDaily(ctx, rows); err != nil {
				return total, err
			}
			total += int64(len(rows))
			s.logger().Info("sw_daily upserted", "trade_date", dates.ISO(d), "rows", len(rows))
		}
		return total, nil
	})
}

func instrumentFromIndex(it domain.IndexBasic) domain.Instrument {
	symbol := it.TSCode
	if i := len(it.TSCode); i > 3 && it.TSCode[i-3] == '.' {
		symbol = it.TSCode[:i-3]
	}
	status := "L"
	if it.ExpDate != nil {
		status = "D"
	}
	return domain.Instrument{
		TSCode:     it.TSCode,
		Symbol:     symbol,
		Name:       it.Name,
		Fullname:   it.Fullname,
		Market:     it.Market,
		Exchange:   it.Market,
		ListStatus: status,
		ListDate:   it.ListDate,
		DelistDate: it.ExpDate,
		AssetType:  "index",
	}
}
