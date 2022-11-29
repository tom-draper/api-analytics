package analytics

import (
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			elapsed := time.Since(start).Milliseconds()

			data := core.Data{
				APIKey:       apiKey,
				Hostname:     c.Request().Host,
				Path:         c.Request().URL.Path,
				UserAgent:    c.Request().UserAgent(),
				Method:       core.MethodMap[c.Request().Method],
				ResponseTime: elapsed,
				Status:       c.Response().Status,
				Framework:    3,
			}

			go core.LogRequest(data)
			return err
		}
	}
}
