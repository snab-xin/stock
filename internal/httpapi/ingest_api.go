package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/ingest"
)

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

func (h *Handler) nextIngestPeriod(c *gin.Context) {
	jobName := c.Query("job")
	if jobName == "" {
		abortJSON(c, http.StatusBadRequest, "缺少 job")
	}
	jobReq, err := ingest.ResolveLatestRunParams(c.Request.Context(), h.runs, jobName)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	if jobReq == nil {
		abortJSON(c, http.StatusNotFound, "没有找到最近一次的采集任务")
		return
	}
	err = jobReq.NextPeriod()
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, jobReq)
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
