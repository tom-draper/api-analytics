package routes

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimitBytesScalesWithMaxInsert(t *testing.T) {
	// Larger batch limits must permit larger bodies, and every limit keeps the
	// envelope headroom.
	small := bodyLimitBytes(1)
	large := bodyLimitBytes(2000)
	if large <= small {
		t.Errorf("bodyLimitBytes(2000)=%d not greater than bodyLimitBytes(1)=%d", large, small)
	}
	if small <= 64*1024 {
		t.Errorf("bodyLimitBytes(1)=%d dropped the envelope headroom", small)
	}
}

func TestBodyLimitRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const limit = 1024
	router := gin.New()
	router.POST("/", bodyLimit(limit), func(c *gin.Context) {
		// Reading past the limit must fail rather than buffer the whole body.
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
