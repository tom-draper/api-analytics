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

		data := core.Data{
			APIKey:       apiKey,
			Hostname:     c.Hostname(),
			Path:         c.Path(),
			UserAgent:    string(c.Request().Header.UserAgent()),
			Method:       c.Method(),
			Status:       c.Response().StatusCode(),
			Framework:    "Fiber",
			ResponseTime: time.Since(start).Milliseconds(),
		}

		go core.LogRequest(data)

		return err
	}
}
