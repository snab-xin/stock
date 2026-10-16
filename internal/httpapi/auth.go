package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/repo"
)

const sessionCookie = "ashare_session"
const sessionTTL = 14 * 24 * time.Hour

func (h *Handler) authStatus(c *gin.Context) {
	n, err := h.users.Count(c.Request.Context())
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	u, ok := h.currentUser(c)
	c.JSON(http.StatusOK, gin.H{
		"needs_setup": n == 0,
		"user":        userJSON(u, ok),
	})
}

func (h *Handler) register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求无效")
		return
	}
	username := strings.TrimSpace(req.Username)
	if len(username) < 2 || len(username) > 64 {
		abortJSON(c, http.StatusBadRequest, "用户名长度为 2–64")
		return
	}
	if len(req.Password) < 6 {
		abortJSON(c, http.StatusBadRequest, "密码至少 6 位")
		return
	}
	u, err := h.users.Create(c.Request.Context(), username, req.Password)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, "无法注册，用户名可能已存在")
		return
	}
	if err := h.startSession(c, u.ID); err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userJSON(u, true)})
}

func (h *Handler) login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求无效")
		return
	}
	u, err := h.users.GetByUsername(c.Request.Context(), req.Username)
	if err != nil || !repo.CheckPassword(u, req.Password) {
		abortJSON(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if u.Status == "locked" {
		abortJSON(c, http.StatusForbidden, "账号已锁定")
		return
	}
	if err := h.startSession(c, u.ID); err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userJSON(u, true)})
}

func (h *Handler) logout(c *gin.Context) {
	if token, err := c.Cookie(sessionCookie); err == nil && token != "" {
		_ = h.users.DeleteSession(c.Request.Context(), token)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) me(c *gin.Context) {
	u, ok := h.currentUser(c)
	if !ok {
		abortJSON(c, http.StatusUnauthorized, "未登录")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userJSON(u, true)})
}

func (h *Handler) requireUser(c *gin.Context) {
	u, ok := h.currentUser(c)
	if !ok {
		abortJSON(c, http.StatusUnauthorized, "未登录")
		c.Abort()
		return
	}
	c.Set("user", u)
	c.Next()
}

func (h *Handler) currentUser(c *gin.Context) (domain.User, bool) {
	token, err := c.Cookie(sessionCookie)
	if err != nil || token == "" || h.users == nil {
		return domain.User{}, false
	}
	u, err := h.users.UserByToken(c.Request.Context(), token)
	if err != nil {
		return domain.User{}, false
	}
	if u.Status == "locked" {
		return domain.User{}, false
	}
	return u, true
}

func userFromCtx(c *gin.Context) (domain.User, bool) {
	v, ok := c.Get("user")
	if !ok {
		return domain.User{}, false
	}
	u, ok := v.(domain.User)
	return u, ok
}

func (h *Handler) startSession(c *gin.Context, userID int64) error {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	token := hex.EncodeToString(buf)
	if err := h.users.CreateSession(c.Request.Context(), userID, token, sessionTTL); err != nil {
		return err
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, token, int(sessionTTL.Seconds()), "/", "", false, true)
	return nil
}

func userJSON(u domain.User, ok bool) any {
	if !ok {
		return nil
	}
	return gin.H{
		"id":       u.ID,
		"username": u.Username,
		"role":     u.Role,
		"status":   u.Status,
	}
}
