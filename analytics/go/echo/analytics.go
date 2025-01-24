package analytics

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	PrivacyLevel int
	ServerURL    string
	GetPath      func(c echo.Context) string
	GetHostname  func(c echo.Context) string
	GetUserAgent func(c echo.Context) string
	GetIPAddress func(c echo.Context) string
	GetUserID    func(c echo.Context) string
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

func Analytics(apiKey string) echo.MiddlewareFunc {
	return AnalyticsWithConfig(apiKey, NewConfig())
}

func AnalyticsWithConfig(apiKey string, config *Config) echo.MiddlewareFunc {
	client := core.NewClient(apiKey, "Echo", config.PrivacyLevel, config.ServerURL)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			data := core.RequestData{
				Hostname:     getHostname(c, config),
				IPAddress:    getIPAddress(c, config),
				Path:         getPath(c, config),
				UserAgent:    getUserAgent(c, config),
				Method:       c.Request().Method,
				Status:       c.Response().Status,
				ResponseTime: time.Since(start).Milliseconds(),
				UserID:       getUserID(c, config),
				CreatedAt:    start.Format(time.RFC3339),
			}

			client.LogRequest(data)
			return err
		}
	}
}

func getHostname(c echo.Context, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(c)
	}
	return GetHostname(c)
}

func getPath(c echo.Context, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(c)
	}
	return GetPath(c)
}

func getUserAgent(c echo.Context, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return GetUserAgent(c)
}

func getIPAddress(c echo.Context, config *Config) string {
	// IP address never sent to the server for privacy level 2 and above
	if config.PrivacyLevel >= 2 {
		return ""
	}

	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return GetIPAddress(c)
}

func getUserID(c echo.Context, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return GetUserID(c)
}

func GetHostname(c echo.Context) string {
	return c.Request().Host
}

func GetPath(c echo.Context) string {
	return c.Request().URL.Path
}

func GetUserAgent(c echo.Context) string {
	return c.Request().UserAgent()
}

func GetIPAddress(c echo.Context) string {
	return c.RealIP()
}

func GetUserID(c echo.Context) string {
	return ""
}
