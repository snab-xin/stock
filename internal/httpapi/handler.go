package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/pkg/dates"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

type Handler struct {
	cfg         config.Config
	instruments *repo.InstrumentRepo
	market      *repo.MarketRepo
	cb          *repo.CBRepo
	runs        *repo.IngestRunRepo
	users       *repo.UserRepo
	notes       *repo.NoteRepo
	runner      *ingest.Runner
}

func abortJSON(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func parseLimit(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	if n > 1000 {
		return 1000
	}
	return n
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
