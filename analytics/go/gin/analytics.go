package analytics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	GetPath      func(c *gin.Context) string
	GetHostname  func(c *gin.Context) string
	GetUserAgent func(c *gin.Context) string
	GetIPAddress func(c *gin.Context) string
	GetUserID    func(c *gin.Context) string
}

func Analytics(apiKey string) gin.HandlerFunc {
	return AnalyticsWithConfig(apiKey, &Config{})
}

func AnalyticsWithConfig(apiKey string, config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		data := core.RequestData{
			Hostname:     getHostname(c, config),
			IPAddress:    getIPAddress(c, config),
			Path:         getPath(c, config),
			UserAgent:    getUserAgent(c, config),
			Method:       c.Request.Method,
			Status:       c.Writer.Status(),
			ResponseTime: time.Since(start).Milliseconds(),
			CreatedAt:    start.Format(time.RFC3339),
		}

		core.LogRequest(apiKey, data, "Gin")
	}
}

func getHostname(c *gin.Context, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(c)
	}
	return c.Request.Host
}

func getPath(c *gin.Context, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(c)
	}
	return c.Request.URL.Path
}

func getUserAgent(c *gin.Context, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return c.Request.UserAgent()
}

func getIPAddress(c *gin.Context, config *Config) string {
	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return c.ClientIP()
}

func getUserID(c *gin.Context, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return ""
}
