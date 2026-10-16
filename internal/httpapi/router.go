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
		cb:          repo.NewCBRepo(gdb),
		runs:        repo.NewIngestRunRepo(gdb),
		users:       repo.NewUserRepo(gdb),
		notes:       repo.NewNoteRepo(gdb),
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
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Content-Type"},
	}))
	_ = r.SetTrustedProxies(nil)

	api := r.Group("/api")
	{
		api.GET("/health", h.health)
		api.GET("/auth/status", h.authStatus)
		api.POST("/auth/register", h.register)
		api.POST("/auth/login", h.login)
		api.POST("/auth/logout", h.logout)
		api.GET("/auth/me", h.me)
		api.GET("/config", h.configInfo)

		notes := api.Group("/notes", h.requireUser)
		notes.GET("", h.noteList)
		notes.GET("/marks", h.noteMarks)
		notes.GET("/:id", h.noteGet)
		notes.POST("", h.noteCreate)
		notes.PUT("/:id", h.noteUpdate)
		notes.DELETE("/:id", h.noteDelete)

		api.GET("/instruments", h.instrumentsSearch)
		api.GET("/stocks", h.stockList)
		api.GET("/stocks/:code/neighbors", h.stockNeighbors)
		api.GET("/bars", h.bars)
		api.POST("/indicators", h.indicators)
		api.GET("/gaps", h.gaps)
		api.GET("/ingest-runs", h.ingestRuns)
		api.GET("/ingest/status", h.ingestStatus)
		api.POST("/ingest", h.ingestStart)
		api.POST("/backtest", h.backtest)
		api.GET("/cb", h.cbList)
		api.GET("/cb/:code/neighbors", h.cbNeighbors)
		api.GET("/cb/:code", h.cbDetail)
		api.PUT("/cb/:code/holders", h.cbReplaceHolders)
	}
	return r
}
