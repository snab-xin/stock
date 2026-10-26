package httpapi

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/indicator"
	"github.com/stock/ashare/internal/pkg/dates"
	"github.com/stock/ashare/internal/repo"
)

func (h *Handler) bars(c *gin.Context) {
	code, inst, ok := h.resolveOrAbort(c)
	if !ok {
		return
	}
	qfq := c.Query("qfq") == "1" || c.Query("qfq") == "true"
	period := repo.NormalizeBarPeriod(c.Query("period"))
	limit := parseLimit(c.Query("limit"), 0)

	var rows []domain.DailyBar
	var err error

	if beforeStr := strings.TrimSpace(c.Query("before")); beforeStr != "" {
		before, perr := dates.Parse(beforeStr)
		if perr != nil {
			abortJSON(c, http.StatusBadRequest, perr.Error())
			return
		}
		if limit <= 0 {
			limit = 250
		}
		rows, err = h.market.ListDailyBarsBefore(c.Request.Context(), code, before, limit, qfq, period)
	} else {
		start, end, perr := parseRange(c)
		if perr != nil {
			abortJSON(c, http.StatusBadRequest, perr.Error())
			return
		}
		if limit <= 0 && start.IsZero() && end.IsZero() {
			limit = 250
		}
		rows, err = h.market.ListDailyBars(c.Request.Context(), code, start, end, limit, qfq, period)
	}
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]barDTO, 0, len(rows))
	closes := make([]float64, len(rows))
	for i, b := range rows {
		out = append(out, toBarDTO(b))
		closes[i] = b.Close
	}
	c.JSON(http.StatusOK, gin.H{
		"instrument": inst,
		"qfq":        qfq,
		"period":     string(period),
		"bars":       out,
		"indicators": indicatorsJSON(closes),
		"has_more":   limit > 0 && len(rows) >= limit,
	})
}

func (h *Handler) indicators(c *gin.Context) {
	var req struct {
		Closes []float64 `json:"closes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "需要 closes 数组")
		return
	}
	if len(req.Closes) > 5000 {
		abortJSON(c, http.StatusBadRequest, "closes 过长（最多 5000）")
		return
	}
	c.JSON(http.StatusOK, indicatorsJSON(req.Closes))
}

func indicatorsJSON(closes []float64) gin.H {
	macd := indicator.MACDDefault(closes)
	return gin.H{
		"rsi":       floatOrNullSlice(indicator.RSI14(closes)),
		"macd_dif":  floatOrNullSlice(macd.DIF),
		"macd_dea":  floatOrNullSlice(macd.DEA),
		"macd_hist": floatOrNullSlice(macd.HIST),
	}
}

// floatOrNullSlice 把 NaN/Inf 编成 JSON null，其余保留为 number。
func floatOrNullSlice(xs []float64) []any {
	out := make([]any, len(xs))
	for i, v := range xs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			out[i] = nil
			continue
		}
		out[i] = v
	}
	return out
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
	rows, err := h.market.ListDailyBars(c.Request.Context(), code, start, end, 0, false, repo.PeriodDay)
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
