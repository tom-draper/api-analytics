package routes

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimitRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const limit = 1024
	router := gin.New()
	router.POST("/", bodyLimit(limit), func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.String(413, "too large")
			return
		}
		c.String(200, "ok")
	})

	t.Run("within limit passes", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("a", 512)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("body within limit got %d, want 200", w.Code)
		}
	})

	t.Run("over limit is rejected", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader(strings.Repeat("a", limit*4)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != 413 {
			t.Errorf("oversized body got %d, want 413", w.Code)
		}
	})
}
