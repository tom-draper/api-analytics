package analytics

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(apiKey string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start).Milliseconds()

		data := core.Data{
			APIKey:       apiKey,
			Hostname:     c.Hostname(),
			Path:         c.Path(),
			UserAgent:    string(c.Request().Header.UserAgent()),
			Method:       core.MethodMap[c.Method()],
			ResponseTime: elapsed,
			Status:       c.Response().StatusCode(),
			Framework:    8,
		}

		go core.LogRequest(data)

		return err
	}
}
