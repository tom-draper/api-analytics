package analytics

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	GetPath      func(c echo.Context) string
	GetHostname  func(c echo.Context) string
	GetUserAgent func(c echo.Context) string
	GetIPAddress func(c echo.Context) string
	GetUserID    func(c echo.Context) string
}

func Analytics(apiKey string) echo.MiddlewareFunc {
	return AnalyticsWithConfig(apiKey, &Config{})
}

func AnalyticsWithConfig(apiKey string, config *Config) echo.MiddlewareFunc {
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
				CreatedAt:    start.Format(time.RFC3339),
			}

			core.LogRequest(apiKey, data, "Echo")
			return err
		}
	}
}

func getHostname(c echo.Context, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(c)
	}
	return c.Request().Host
}

func getPath(c echo.Context, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(c)
	}
	return c.Request().URL.Path
}

func getUserAgent(c echo.Context, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return c.Request().UserAgent()
}

func getIPAddress(c echo.Context, config *Config) string {
	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return c.RealIP()
}

func getUserID(c echo.Context, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return ""
}
