package analytics

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	GetPath      func(c *fiber.Ctx) string
	GetHostname  func(c *fiber.Ctx) string
	GetUserAgent func(c *fiber.Ctx) string
	GetIPAddress func(c *fiber.Ctx) string
	GetUserID    func(c *fiber.Ctx) string
}

func Analytics(apiKey string) func(c *fiber.Ctx) error {
	return AnalyticsWithConfig(apiKey, &Config{})
}

func AnalyticsWithConfig(apiKey string, config *Config) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		data := core.RequestData{
			Hostname:     getHostname(c, config),
			Path:         getPath(c, config),
			IPAddress:    getIPAddress(c, config),
			UserAgent:    getUserAgent(c, config),
			Method:       c.Method(),
			Status:       c.Response().StatusCode(),
			ResponseTime: time.Since(start).Milliseconds(),
			CreatedAt:    start.Format(time.RFC3339),
		}

		core.LogRequest(apiKey, data, "Fiber")

		return err
	}
}

func getHostname(c *fiber.Ctx, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(c)
	}
	return c.Hostname()
}

func getPath(c *fiber.Ctx, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(c)
	}
	return c.Path()
}

func getUserAgent(c *fiber.Ctx, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return string(c.Request().Header.UserAgent())
}

func getIPAddress(c *fiber.Ctx, config *Config) string {
	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return c.IP()
}

func getUserID(c *fiber.Ctx, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return ""
}
