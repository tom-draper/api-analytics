package analytics

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	PrivacyLevel int
	ServerURL    string
	GetPath      func(c *fiber.Ctx) string
	GetHostname  func(c *fiber.Ctx) string
	GetUserAgent func(c *fiber.Ctx) string
	GetIPAddress func(c *fiber.Ctx) string
	GetUserID    func(c *fiber.Ctx) string
}

func NewConfig() *Config {
	return &Config{
		PrivacyLevel: 0,
		ServerURL: core.DefaultServerURL,
		GetPath: GetPath,
		GetHostname: GetHostname,
		GetUserAgent: GetUserAgent,
		GetIPAddress: GetIPAddress,
		GetUserID: GetUserID,
	}
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
			UserID:       getUserID(c, config),
			CreatedAt:    start.Format(time.RFC3339),
		}

		core.LogRequest(apiKey, data, "Fiber", config.PrivacyLevel, config.ServerURL)

		return err
	}
}

func getHostname(c *fiber.Ctx, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(c)
	}
	return GetHostname(c)
}

func getPath(c *fiber.Ctx, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(c)
	}
	return GetPath(c)
}

func getUserAgent(c *fiber.Ctx, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return GetUserAgent(c)
}

func getIPAddress(c *fiber.Ctx, config *Config) string {
	if config.PrivacyLevel >= 2 {
		return ""
	}

	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return GetIPAddress(c)
}

func getUserID(c *fiber.Ctx, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return GetUserID(c)
}

func GetHostname(c *fiber.Ctx) string {
	return c.Hostname()
}

func GetPath(c *fiber.Ctx) string {
	return c.Path()
}

func GetUserAgent(c *fiber.Ctx) string {
	return string(c.Request().Header.UserAgent())
}

func GetIPAddress(c *fiber.Ctx) string {
	return c.IP()
}

func GetUserID(c *fiber.Ctx) string {
	return ""
}
