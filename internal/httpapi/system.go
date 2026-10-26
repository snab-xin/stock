package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
