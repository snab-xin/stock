package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

type noteBody struct {
	TSCode     string `json:"ts_code"`
	TargetType string `json:"target_type"`
	TradeDate  string `json:"trade_date"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
}

func (h *Handler) noteList(c *gin.Context) {
	u, _ := userFromCtx(c)
	items, err := h.notes.List(c.Request.Context(), u.ID, c.Query("ts_code"), c.Query("target_type"), 100)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, noteJSON(it, false))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) noteMarks(c *gin.Context) {
	u, _ := userFromCtx(c)
	code := strings.TrimSpace(c.Query("ts_code"))
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少 ts_code")
		return
	}
	items, err := h.notes.Marks(c.Request.Context(), u.ID, code)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		out = append(out, gin.H{
			"id":         it.ID,
			"trade_date": dates.ISO(it.TradeDate),
			"summary":    it.Summary,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) noteGet(c *gin.Context) {
	u, _ := userFromCtx(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abortJSON(c, http.StatusBadRequest, "笔记 id 无效")
		return
	}
	it, err := h.notes.Get(c.Request.Context(), u.ID, id)
	if err == gorm.ErrRecordNotFound {
		abortJSON(c, http.StatusNotFound, "笔记不存在")
		return
	}
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, noteJSON(it, true))
}

func (h *Handler) noteCreate(c *gin.Context) {
	u, _ := userFromCtx(c)
	n, ok := bindNote(c, u.ID)
	if !ok {
		return
	}
	if err := h.notes.Create(c.Request.Context(), &n); err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, noteJSON(n, true))
}

func (h *Handler) noteUpdate(c *gin.Context) {
	u, _ := userFromCtx(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abortJSON(c, http.StatusBadRequest, "笔记 id 无效")
		return
	}
	if _, err := h.notes.Get(c.Request.Context(), u.ID, id); err == gorm.ErrRecordNotFound {
		abortJSON(c, http.StatusNotFound, "笔记不存在")
		return
	} else if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	n, ok := bindNote(c, u.ID)
	if !ok {
		return
	}
	n.ID = id
	if err := h.notes.Update(c.Request.Context(), &n); err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, noteJSON(n, true))
}

func (h *Handler) noteDelete(c *gin.Context) {
	u, _ := userFromCtx(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abortJSON(c, http.StatusBadRequest, "笔记 id 无效")
		return
	}
	if err := h.notes.Delete(c.Request.Context(), u.ID, id); err == gorm.ErrRecordNotFound {
		abortJSON(c, http.StatusNotFound, "笔记不存在")
		return
	} else if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func bindNote(c *gin.Context, userID int64) (domain.UserNote, bool) {
	var req noteBody
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求无效")
		return domain.UserNote{}, false
	}
	td, err := dates.Parse(req.TradeDate)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return domain.UserNote{}, false
	}
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		abortJSON(c, http.StatusBadRequest, "标题必填")
		return domain.UserNote{}, false
	}
	if len(summary) > 256 {
		abortJSON(c, http.StatusBadRequest, "标题过长")
		return domain.UserNote{}, false
	}
	code, kind, err := repo.NormalizeNoteTarget(req.TSCode, req.TargetType)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return domain.UserNote{}, false
	}
	return domain.UserNote{
		UserID:     userID,
		TSCode:     code,
		TargetType: kind,
		TradeDate:  td,
		Summary:    summary,
		Content:    req.Content,
	}, true
}

func noteJSON(it domain.UserNote, withContent bool) gin.H {
	out := gin.H{
		"id":          it.ID,
		"ts_code":     it.TSCode,
		"target_type": it.TargetType,
		"trade_date":  dates.ISO(it.TradeDate),
		"summary":     it.Summary,
		"updated_at":  it.UpdatedAt,
	}
	if withContent {
		out["content"] = it.Content
	}
	return out
}
