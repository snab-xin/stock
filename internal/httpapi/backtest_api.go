package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/backtest"
	"github.com/stock/ashare/internal/pkg/dates"
)

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
