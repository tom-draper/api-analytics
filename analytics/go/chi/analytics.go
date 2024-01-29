package analytics

import (
	"net"
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/analytics/go/core"
)

// http.ResponseWriter wrapper to store status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func createResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

type Config struct {
	GetPath      func(r *http.Request) string
	GetHostname  func(r *http.Request) string
	GetUserAgent func(r *http.Request) string
	GetIPAddress func(r *http.Request) string
	GetUserID    func(r *http.Request) string
	PrivacyLevel int
}

func Analytics(apiKey string) func(next http.Handler) http.Handler {
	return AnalyticsWithConfig(apiKey, &Config{})
}

func AnalyticsWithConfig(apiKey string, config *Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			ctx := r.Context()
			rw := createResponseWriter(w) // Wrap to store status code

			start := time.Now()
			next.ServeHTTP(rw, r.WithContext(ctx))

			data := core.RequestData{
				Hostname:     getHostname(r, config),
				IPAddress:    getIPAddress(r, config),
				Path:         getPath(r, config),
				UserAgent:    getUserAgent(r, config),
				Method:       r.Method,
				Status:       rw.status,
				ResponseTime: time.Since(start).Milliseconds(),
				UserID:       getUserID(r, config),
				CreatedAt:    start.Format(time.RFC3339),
			}

			core.LogRequest(apiKey, data, "Chi", config.PrivacyLevel)
		})
	}
}

func getHostname(r *http.Request, config *Config) string {
	if config.GetHostname != nil {
		return config.GetHostname(r)
	}
	return r.Host
}

func getPath(r *http.Request, config *Config) string {
	if config.GetPath != nil {
		return config.GetPath(r)
	}
	return r.URL.Path
}

func getUserAgent(r *http.Request, config *Config) string {
	if config.GetUserAgent != nil {
		return config.GetUserAgent(r)
	}
	return r.UserAgent()
}

func getIPAddress(r *http.Request, config *Config) string {
	if config.PrivacyLevel >= 2 {
		return ""
	}

	if config.GetIPAddress != nil {
		return config.GetIPAddress(r)
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func getUserID(r *http.Request, config *Config) string {
	if config.GetUserID != nil {
		return config.GetUserID(r)
	}
	return ""
}
