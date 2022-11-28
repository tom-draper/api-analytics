package analytics

import (
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(APIKey string) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			elapsed := time.Since(start).Milliseconds()

			data := core.Data{
				APIKey:       APIKey,
				Hostname:     ctx.Request.Host,
				Path:         ctx.Request.URL.Path,
				UserAgent:    ctx.Request.UserAgent(),
				Method:       core.MethodMap[ctx.Request.Method],
				ResponseTime: elapsed,
				Status:       ctx.Writer.Status(),
				Framework:    7,
			}

			go core.LogRequest(data)
		})
	}
}
