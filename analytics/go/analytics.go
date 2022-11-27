package analytics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type data struct {
	APIKey       string `json:"api_key"`
	Hostname     string `json:"hostname"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       int    `json:"method"`
	ResponseTime int64  `json:"response_time"`
	Status       int    `json:"status"`
	Framework    int8   `json:"framework"`
}

var methodMap = map[string]int{
	"GET":     0,
	"POST":    1,
	"PUT":     2,
	"PATCH":   3,
	"DELETE":  4,
	"OPTIONS": 5,
	"CONNECT": 6,
	"HEAD":    7,
	"TRACE":   8,
}

func logRequest(data data) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		print(err)
	}
	http.Post("https://api-analytics-server.vercel.app/api/log-request", "application/json", bytes.NewBuffer(reqBody))
}

func Analytics(APIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		elapsed := time.Since(start).Milliseconds()

		data := data{
			APIKey:       APIKey,
			Hostname:     c.Request.Host,
			Path:         c.Request.URL.Path,
			UserAgent:    c.Request.UserAgent(),
			Method:       methodMap[c.Request.Method],
			ResponseTime: elapsed,
			Status:       c.Writer.Status(),
			Framework:    2,
		}

		go logRequest(data)
	}
}
