package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/backtest"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

type Handler struct {
	cfg         config.Config
	instruments *repo.InstrumentRepo
	market      *repo.MarketRepo
	cb          *repo.CBRepo
	runs        *repo.IngestRunRepo
	runner      *ingest.Runner
}

func abortJSON(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) configInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tushare_rpm":       h.cfg.TushareRPM,
		"tushare_apis":      h.cfg.TushareAPIs,
		"tushare_url":       h.cfg.TushareURL,
		"http_addr":         h.cfg.HTTPAddr,
		"has_tushare_token": h.cfg.TushareToken != "",
		"backtest":          h.cfg.Backtest,
		"log": gin.H{
			"level":  h.cfg.Log.Level,
			"format": h.cfg.Log.Format,
		},
	})
}

func (h *Handler) instrumentsSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	items, err := h.instruments.Search(c.Request.Context(), q, 20)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]instrumentDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toInstrumentDTO(it))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) bars(c *gin.Context) {
	code, inst, ok := h.resolveOrAbort(c)
	if !ok {
		return
	}
	start, end, err := parseRange(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	qfq := c.Query("qfq") == "1" || c.Query("qfq") == "true"
	limit := 0
	if start.IsZero() && end.IsZero() {
		limit = 250
	}
	rows, err := h.market.ListDailyBars(c.Request.Context(), code, start, end, limit, qfq)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]barDTO, 0, len(rows))
	for _, b := range rows {
		out = append(out, toBarDTO(b))
	}
	c.JSON(http.StatusOK, gin.H{
		"instrument": inst,
		"qfq":        qfq,
		"bars":       out,
	})
}

func (h *Handler) gaps(c *gin.Context) {
	code, inst, ok := h.resolveOrAbort(c)
	if !ok {
		return
	}
	start, end, err := parseRange(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.market.ListDailyBars(c.Request.Context(), code, start, end, 0, false)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{
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
	openDays, err := h.market.ListOpenDays(c.Request.Context(), ex, start, end)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	have := make(map[string]struct{}, len(rows))
	for _, b := range rows {
		have[dates.ISO(b.TradeDate)] = struct{}{}
	}
	missing := make([]string, 0)
	for _, d := range openDays {
		key := dates.ISO(d)
		if _, ok := have[key]; !ok {
			missing = append(missing, key)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"instrument": inst,
		"missing":    missing,
		"bar_count":  len(rows),
		"open_days":  len(openDays),
	})
}

func (h *Handler) ingestRuns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	rows, err := h.runs.ListRecent(c.Request.Context(), limit)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]runDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRunDTO(row))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ingestStatus(c *gin.Context) {
	if h.runner == nil {
		c.JSON(http.StatusOK, ingest.JobStatus{})
		return
	}
	c.JSON(http.StatusOK, h.runner.Status())
}

func (h *Handler) ingestStart(c *gin.Context) {
	if h.runner == nil {
		abortJSON(c, http.StatusServiceUnavailable, "采集服务未启用（检查 TUSHARE_TOKEN）")
		return
	}
	var req ingest.JobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求体无效 JSON")
		return
	}
	if err := h.runner.Start(c.Request.Context(), req); err != nil {
		abortJSON(c, http.StatusConflict, err.Error())
		return
	}
	c.JSON(http.StatusAccepted, h.runner.Status())
}

func (h *Handler) backtest(c *gin.Context) {
	var req backtestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求体无效 JSON")
		return
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少 code")
		return
	}

	p := backtest.Params{
		Fast:           req.Fast,
		Slow:           req.Slow,
		InitCash:       req.Cash,
		CommissionRate: h.cfg.Backtest.CommissionRate,
		CommissionMin:  h.cfg.Backtest.CommissionMin,
		StampTaxRate:   h.cfg.Backtest.StampTaxRate,
		Slippage:       h.cfg.Backtest.Slippage,
		LotSize:        h.cfg.Backtest.LotSize,
	}
	if p.Fast <= 0 {
		p.Fast = h.cfg.Backtest.Fast
	}
	if p.Slow <= 0 {
		p.Slow = h.cfg.Backtest.Slow
	}
	if p.InitCash <= 0 {
		p.InitCash = h.cfg.Backtest.InitCash
	}
	if s := strings.TrimSpace(req.Start); s != "" {
		t, err := dates.Parse(s)
		if err != nil {
			abortJSON(c, http.StatusBadRequest, err.Error())
			return
		}
		p.TradeFrom = t
	}
	if s := strings.TrimSpace(req.End); s != "" {
		t, err := dates.Parse(s)
		if err != nil {
			abortJSON(c, http.StatusBadRequest, err.Error())
			return
		}
		p.TradeTo = t
	}

	res, err := backtest.LoadAndRun(c.Request.Context(), h.instruments, h.market, code, p)
	if err != nil {
		msg := err.Error()
		status := http.StatusInternalServerError
		if strings.Contains(msg, "没有") || strings.Contains(msg, "必须") || strings.Contains(msg, "不足") || strings.Contains(msg, "无法解析") {
			status = http.StatusBadRequest
		}
		abortJSON(c, status, msg)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) resolveOrAbort(c *gin.Context) (string, instrumentDTO, bool) {
	raw := strings.TrimSpace(c.Query("code"))
	if raw == "" {
		abortJSON(c, http.StatusBadRequest, "缺少 code")
		return "", instrumentDTO{}, false
	}
	code, err := ingest.ResolveTSCode(c.Request.Context(), h.instruments, raw)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return "", instrumentDTO{}, false
	}
	inst, err := h.instruments.GetByTSCode(c.Request.Context(), code)
	if err == gorm.ErrRecordNotFound {
		return code, instrumentDTO{TSCode: code}, true
	}
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return "", instrumentDTO{}, false
	}
	return code, toInstrumentDTO(inst), true
}

func parseRange(c *gin.Context) (start, end time.Time, err error) {
	if s := c.Query("start"); s != "" {
		start, err = dates.Parse(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if s := c.Query("end"); s != "" {
		end, err = dates.Parse(s)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	return start, end, nil
}
