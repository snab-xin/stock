package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
