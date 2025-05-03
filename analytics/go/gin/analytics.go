package analytics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

type Config struct {
	PrivacyLevel int
	ServerURL    string
	GetPath      func(c *gin.Context) string
	GetHostname  func(c *gin.Context) string
	GetUserAgent func(c *gin.Context) string
	GetIPAddress func(c *gin.Context) string
	GetUserID    func(c *gin.Context) string
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

func Analytics(apiKey string) gin.HandlerFunc {
	return AnalyticsWithConfig(apiKey, NewConfig())
}

func AnalyticsWithConfig(apiKey string, config *Config) gin.HandlerFunc {
	client := core.NewClient(apiKey, "Gin", config.PrivacyLevel, config.ServerURL)
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
			UserID:       getUserID(c, config),
			CreatedAt:    start.Format(time.RFC3339),
		}

		client.LogRequest(data)
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
	return GetPath(c)
}

func getUserAgent(c *gin.Context, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(c)
	}
	return GetUserAgent(c)
}

func getIPAddress(c *gin.Context, config *Config) string {
	// IP address never sent to the server for privacy level 2 and above
	if config.PrivacyLevel >= 2 {
		return ""
	}

	if config.GetIPAddress != nil {
		return config.GetIPAddress(c)
	}
	return GetIPAddress(c)
}

func getUserID(c *gin.Context, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(c)
	}
	return GetUserID(c)
}

func GetHostname(c *gin.Context) string {
	return c.Request.Host
}

func GetPath(c *gin.Context) string {
	return c.Request.URL.Path
}

func GetUserAgent(c *gin.Context) string {
	return c.Request.UserAgent()
}

func GetIPAddress(c *gin.Context) string {
	return c.ClientIP()
}

func GetUserID(c *gin.Context) string {
	return ""
}
