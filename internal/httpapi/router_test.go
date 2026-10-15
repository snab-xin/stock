package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	r := newEngine(&Handler{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "{\"status\":\"ok\"}" && got != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body=%q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	r := newEngine(&Handler{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/backtest", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatal("missing CORS origin")
	}
}
