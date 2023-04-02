package analytics

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			data := core.Data{
				Hostname:     c.Request().Host,
				IPAddress:    c.RealIP(),
				Path:         c.Request().URL.Path,
				UserAgent:    c.Request().UserAgent(),
				Method:       c.Request().Method,
				Status:       c.Response().Status,
				ResponseTime: time.Since(start).Milliseconds(),
				CreatedAt:    start.Format(time.RFC3339),
			}

			core.LogRequest(apiKey, data, "Echo")
			return err
		}
	}
}
