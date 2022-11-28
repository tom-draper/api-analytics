package analytics

import (
	"fmt"
	"time"

	echo "github.com/labstack/echo/v4"
	core "github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(APIKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			elapsed := time.Since(start).Milliseconds()

			data := core.Data{
				APIKey:       APIKey,
				Hostname:     c.Request().Host,
				Path:         c.Request().URL.Path,
				UserAgent:    c.Request().UserAgent(),
				Method:       core.MethodMap[c.Request().Method],
				ResponseTime: elapsed,
				Status:       c.Response().Status,
				Framework:    2,
			}

			fmt.Println(data)

			go core.LogRequest(data)
			return err
		}
	}
}
