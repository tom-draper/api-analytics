package analytics

import (
	"strings"
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
		ServerURL:    core.DefaultServerURL,
		GetPath:      GetPath,
		GetHostname:  GetHostname,
		GetUserAgent: GetUserAgent,
		GetIPAddress: GetIPAddress,
		GetUserID:    GetUserID,
	}
}

func Analytics(apiKey string) func(c *fiber.Ctx) error {
	return AnalyticsWithConfig(apiKey, NewConfig())
}

func AnalyticsWithConfig(apiKey string, config *Config) func(c *fiber.Ctx) error {
	middleware, _ := AnalyticsWithClient(apiKey, config)
	return middleware
}

// AnalyticsWithClient is like AnalyticsWithConfig but also returns the underlying
// client. Call client.Shutdown() on graceful exit to flush buffered requests: the
// client batches and flushes roughly once a minute, so the final batch is lost on
// exit otherwise. The client is nil when apiKey is empty; Shutdown handles that.
func AnalyticsWithClient(apiKey string, config *Config) (func(c *fiber.Ctx) error, *core.Client) {
	client := core.NewClient(apiKey, "Fiber", config.PrivacyLevel, config.ServerURL)
	middleware := func(c *fiber.Ctx) error {
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

		client.LogRequest(data)

		return err
	}
	return middleware, client
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
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if fwd := c.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
	}
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.IP()
}

func GetUserID(c *fiber.Ctx) string {
	return ""
}
