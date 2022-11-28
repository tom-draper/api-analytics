package analytics

import (
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/analytics/go/core"
)

func Analytics(APIKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			elapsed := time.Since(start).Milliseconds()

			data := core.Data{
				APIKey:       APIKey,
				Hostname:     r.Host,
				Path:         r.URL.Path,
				UserAgent:    r.UserAgent(),
				Method:       core.MethodMap[r.Method],
				ResponseTime: elapsed,
				Status:       r.Response.StatusCode,
				Framework:    7,
			}

			fmt.PrinlN(data)

			go core.LogRequest(data)
		})
	}
}
