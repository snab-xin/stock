package httpapi

import (
	"encoding/json"
	"time"

	"github.com/stock/ashare/internal/pkg/dates"
	"github.com/stock/ashare/internal/domain"
)

type instrumentDTO struct {
	TSCode   string `json:"ts_code"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	IsST     bool   `json:"is_st"`
	Status   string `json:"status"`
}

type barDTO struct {
	TradeDate   string   `json:"trade_date"`
	Open        float64  `json:"open"`
	High        float64  `json:"high"`
	Low         float64  `json:"low"`
	Close       float64  `json:"close"`
	Volume      int64    `json:"volume"`
	Amount      float64  `json:"amount"`
	IsLimitUp   *bool    `json:"is_limit_up"`
	IsLimitDown *bool    `json:"is_limit_down"`
	IsPaused    bool     `json:"is_paused"`
	HighLimit   *float64 `json:"high_limit"`
	LowLimit    *float64 `json:"low_limit"`
}

type runDTO struct {
	ID           int64           `json:"id"`
	JobName      string          `json:"job_name"`
	Status       string          `json:"status"`
	Message      string          `json:"message"`
	RowsAffected int64           `json:"rows_affected"`
	StartedAt    string          `json:"started_at"`
	FinishedAt   string          `json:"finished_at,omitempty"`
	Params       json.RawMessage `json:"params"`
}

type backtestRequest struct {
	Code  string  `json:"code"`
	Start string  `json:"start"`
	End   string  `json:"end"`
	Fast  int     `json:"fast"`
	Slow  int     `json:"slow"`
	Cash  float64 `json:"cash"`
}

func toInstrumentDTO(it domain.Instrument) instrumentDTO {
	return instrumentDTO{
		TSCode:   it.TSCode,
		Symbol:   it.Symbol,
		Name:     it.Name,
		Exchange: it.Exchange,
		IsST:     it.IsST,
		Status:   it.ListStatus,
	}
}

func toBarDTO(b domain.DailyBar) barDTO {
	return barDTO{
		TradeDate:   dates.ISO(b.TradeDate),
		Open:        b.Open,
		High:        b.High,
		Low:         b.Low,
		Close:       b.Close,
		Volume:      b.Volume,
		Amount:      b.Amount,
		IsLimitUp:   b.IsLimitUp,
		IsLimitDown: b.IsLimitDown,
		IsPaused:    b.IsPaused,
		HighLimit:   b.HighLimit,
		LowLimit:    b.LowLimit,
	}
}

func toRunDTO(row domain.IngestRun) runDTO {
	params := json.RawMessage(row.Params)
	if len(params) == 0 {
		params = json.RawMessage("{}")
	}
	dto := runDTO{
		ID:           row.ID,
		JobName:      row.JobName,
		Status:       row.Status,
		Message:      row.Message,
		RowsAffected: row.RowsAffected,
		StartedAt:    row.StartedAt.Format(time.RFC3339),
		Params:       params,
	}
	if row.FinishedAt != nil {
		dto.FinishedAt = row.FinishedAt.Format(time.RFC3339)
	}
	return dto
}
