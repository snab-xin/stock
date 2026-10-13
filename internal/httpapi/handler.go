package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

type Handler struct {
	cfg         config.Config
	instruments *repo.InstrumentRepo
	market      *repo.MarketRepo
	runs        *repo.IngestRunRepo
	runner      *ingest.Runner
}

func NewMux(gdb *gorm.DB, cfg config.Config, runner *ingest.Runner) http.Handler {
	h := &Handler{
		cfg:         cfg,
		instruments: repo.NewInstrumentRepo(gdb),
		market:      repo.NewMarketRepo(gdb),
		runs:        repo.NewIngestRunRepo(gdb),
		runner:      runner,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/config", h.configInfo)
	mux.HandleFunc("GET /api/instruments", h.instrumentsSearch)
	mux.HandleFunc("GET /api/bars", h.bars)
	mux.HandleFunc("GET /api/gaps", h.gaps)
	mux.HandleFunc("GET /api/ingest-runs", h.ingestRuns)
	mux.HandleFunc("GET /api/ingest/status", h.ingestStatus)
	mux.HandleFunc("POST /api/ingest", h.ingestStart)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) configInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"tushare_rpm":     h.cfg.TushareRPM,
		"tushare_url":     h.cfg.TushareURL,
		"http_addr":       h.cfg.HTTPAddr,
		"has_tushare_token": h.cfg.TushareToken != "",
	})
}

func (h *Handler) instrumentsSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.instruments.Search(r.Context(), q, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]instrumentDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toInstrumentDTO(it))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) bars(w http.ResponseWriter, r *http.Request) {
	code, inst, err := h.resolve(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	start, end, err := parseRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	qfq := r.URL.Query().Get("qfq") == "1" || r.URL.Query().Get("qfq") == "true"
	limit := 0
	if start.IsZero() && end.IsZero() {
		limit = 250
	}
	rows, err := h.market.ListDailyBars(r.Context(), code, start, end, limit, qfq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]barDTO, 0, len(rows))
	for _, b := range rows {
		out = append(out, toBarDTO(b))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instrument": inst,
		"qfq":        qfq,
		"bars":       out,
	})
}

func (h *Handler) gaps(w http.ResponseWriter, r *http.Request) {
	code, inst, err := h.resolve(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	start, end, err := parseRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.market.ListDailyBars(r.Context(), code, start, end, 0, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(rows) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"instrument": inst,
			"missing":    []string{},
			"bar_count":  0,
		})
		return
	}
	if start.IsZero() {
		start = rows[0].TradeDate
	}
	if end.IsZero() {
		end = rows[len(rows)-1].TradeDate
	}
	ex := inst.Exchange
	if ex == "" {
		ex = "SSE"
	}
	openDays, err := h.market.ListOpenDays(r.Context(), ex, start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	have := make(map[string]struct{}, len(rows))
	for _, b := range rows {
		have[dates.ISO(b.TradeDate)] = struct{}{}
	}
	var missing []string
	for _, d := range openDays {
		key := dates.ISO(d)
		if _, ok := have[key]; !ok {
			missing = append(missing, key)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instrument": inst,
		"missing":    missing,
		"bar_count":  len(rows),
		"open_days":  len(openDays),
	})
}

func (h *Handler) ingestRuns(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.runs.ListRecent(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]runDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRunDTO(row))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) ingestStatus(w http.ResponseWriter, _ *http.Request) {
	if h.runner == nil {
		writeJSON(w, http.StatusOK, ingest.JobStatus{})
		return
	}
	writeJSON(w, http.StatusOK, h.runner.Status())
}

func (h *Handler) ingestStart(w http.ResponseWriter, r *http.Request) {
	if h.runner == nil {
		writeError(w, http.StatusServiceUnavailable, "采集服务未启用（检查 TUSHARE_TOKEN）")
		return
	}
	var req ingest.JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体无效 JSON")
		return
	}
	if err := h.runner.Start(r.Context(), req); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, h.runner.Status())
}

func (h *Handler) resolve(r *http.Request) (string, instrumentDTO, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("code"))
	if raw == "" {
		return "", instrumentDTO{}, errBad("缺少 code")
	}
	code, err := ingest.ResolveTSCode(r.Context(), h.instruments, raw)
	if err != nil {
		return "", instrumentDTO{}, err
	}
	inst, err := h.instruments.GetByTSCode(r.Context(), code)
	if err == gorm.ErrRecordNotFound {
		return code, instrumentDTO{TSCode: code}, nil
	}
	if err != nil {
		return "", instrumentDTO{}, err
	}
	return code, toInstrumentDTO(inst), nil
}

func parseRange(r *http.Request) (start, end time.Time, err error) {
	if s := r.URL.Query().Get("start"); s != "" {
		start, err = dates.Parse(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if s := r.URL.Query().Get("end"); s != "" {
		end, err = dates.Parse(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	return start, end, nil
}

type badRequest string

func (e badRequest) Error() string { return string(e) }

func errBad(msg string) error { return badRequest(msg) }

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
	ID           int64  `json:"id"`
	JobName      string `json:"job_name"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	RowsAffected int64  `json:"rows_affected"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at,omitempty"`
}

func toInstrumentDTO(it domain.Instrument) instrumentDTO {
	return instrumentDTO{
		TSCode:   it.TSCode,
		Symbol:   it.Symbol,
		Name:     it.Name,
		Exchange: it.Exchange,
		IsST:     it.IsST,
		Status:   it.Status,
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
	dto := runDTO{
		ID:           row.ID,
		JobName:      row.JobName,
		Status:       row.Status,
		Message:      row.Message,
		RowsAffected: row.RowsAffected,
		StartedAt:    row.StartedAt.Format(time.RFC3339),
	}
	if row.FinishedAt != nil {
		dto.FinishedAt = row.FinishedAt.Format(time.RFC3339)
	}
	return dto
}
