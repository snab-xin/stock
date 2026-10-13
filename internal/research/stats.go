package research

import (
	"math"
	"time"

	"github.com/stock/ashare/internal/domain"
)

type SeriesStats struct {
	N           int
	Start       time.Time
	End         time.Time
	StartClose  float64
	EndClose    float64
	TotalReturn float64
	MaxDrawdown float64
	AnnReturn   float64
}

func FromBars(bars []domain.DailyBar) (SeriesStats, error) {
	var s SeriesStats
	if len(bars) == 0 {
		return s, nil
	}
	s.N = len(bars)
	s.Start = bars[0].TradeDate
	s.End = bars[len(bars)-1].TradeDate
	s.StartClose = bars[0].Close
	s.EndClose = bars[len(bars)-1].Close
	if s.StartClose != 0 {
		s.TotalReturn = s.EndClose/s.StartClose - 1
	}
	peak := bars[0].Close
	maxDD := 0.0
	for _, b := range bars {
		if b.Close > peak {
			peak = b.Close
		}
		if peak > 0 {
			dd := b.Close/peak - 1
			if dd < maxDD {
				maxDD = dd
			}
		}
	}
	s.MaxDrawdown = maxDD
	if s.N > 1 {
		s.AnnReturn = math.Pow(1+s.TotalReturn, 252/float64(s.N-1)) - 1
	}
	return s, nil
}
