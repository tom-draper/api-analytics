package analytics

import (
	"fmt"
	"net/http"
	"time"
)

func Analytics(APIKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			fmt.Println(ctx)
			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			elapsed := time.Since(start).Milliseconds()

			fmt.Println(elapsed)

			// data := core.Data{
			// 	APIKey:       APIKey,
			// 	Hostname:     r.Host,
			// 	Path:         r.URL.Path,
			// 	UserAgent:    r.UserAgent(),
			// 	Method:       core.MethodMap[r.Method],
			// 	ResponseTime: elapsed,
			// 	Status:       r.Response.StatusCode,
			// 	Framework:    7,
			// }

			// fmt.Println(data)

			// go core.LogRequest(data)
		})
	}
}
