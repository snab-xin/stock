package httpapi

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/applog"
)

func requestLogger() gin.HandlerFunc {
	log := applog.With("component", "http")
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.Request.URL.Path
		status := c.Writer.Status()
		elapsed := time.Since(start)
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"elapsed_ms", applog.DurationMS(elapsed),
		}
		if q := c.Request.URL.RawQuery; q != "" {
			attrs = append(attrs, "query", q)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "err", c.Errors.String())
		}

		switch {
		case status >= 500:
			log.Error("request", attrs...)
		case status >= 400:
			log.Warn("request", attrs...)
		case path == "/api/health":
			log.Debug("request", attrs...)
		default:
			log.Info("request", attrs...)
		}
	}
}

// recoveryLogger 把 panic 细节留给 Gin Recovery，同时保证有结构化入口。
func recoveryLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic",
			"component", "http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
		)
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
	})
}
