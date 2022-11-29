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
		elapsed := time.Since(start).Milliseconds()

		data := core.Data{
			APIKey:       apiKey,
			Hostname:     c.Request.Host,
			Path:         c.Request.URL.Path,
			UserAgent:    c.Request.UserAgent(),
			Method:       core.MethodMap[c.Request.Method],
			ResponseTime: elapsed,
			Status:       c.Writer.Status(),
			Framework:    2,
		}

		go core.LogRequest(data)
	}
}
