package httpapi

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

func NewMux(gdb *gorm.DB, cfg config.Config, runner *ingest.Runner) http.Handler {
	h := &Handler{
		cfg:         cfg,
		instruments: repo.NewInstrumentRepo(gdb),
		market:      repo.NewMarketRepo(gdb),
		runs:        repo.NewIngestRunRepo(gdb),
		runner:      runner,
	}
	return newEngine(h)
}

func newEngine(h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(recoveryLogger(), requestLogger())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:    []string{"Content-Type"},
	}))
	_ = r.SetTrustedProxies(nil)

	api := r.Group("/api")
	{
		api.GET("/health", h.health)
		api.GET("/config", h.configInfo)
		api.GET("/instruments", h.instrumentsSearch)
		api.GET("/bars", h.bars)
		api.GET("/gaps", h.gaps)
		api.GET("/ingest-runs", h.ingestRuns)
		api.GET("/ingest/status", h.ingestStatus)
		api.POST("/ingest", h.ingestStart)
		api.POST("/backtest", h.backtest)
	}
	return r
}
