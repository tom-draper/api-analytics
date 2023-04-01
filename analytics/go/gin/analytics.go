package analytics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		data := core.Data{
			APIKey:       apiKey,
			Hostname:     c.Request.Host,
			IPAddress:    c.ClientIP(),
			Path:         c.Request.URL.Path,
			UserAgent:    c.Request.UserAgent(),
			Method:       c.Request.Method,
			Status:       c.Writer.Status(),
			Framework:    "Gin",
			ResponseTime: time.Since(start).Milliseconds(),
			CreatedAt:    start.Format(time.RFC3339),
		}

		core.LogRequest(data)
	}
}
