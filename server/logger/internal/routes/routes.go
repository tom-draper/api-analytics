package routes

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"
)

// maxRequestBytes is a generous JSON upper bound for a single logged request
// (all string fields are capped at 255 bytes). The ingest body limit is derived
// from it and MaxInsert so the whole payload cannot be unmarshalled into
// unbounded memory before the per-request cap is applied.
const maxRequestBytes = 4096

// bodyLimitBytes bounds the ingest request body to what MaxInsert requests could
// plausibly occupy, plus headroom for the payload envelope.
func bodyLimitBytes(maxInsert int) int64 {
	return int64(maxInsert)*maxRequestBytes + 64*1024
}

func RegisterRouter(r *gin.RouterGroup, db *database.DB, geoIPDB *geoip2.Reader, cfg *config.Config, startTime time.Time) {
	// geoIP is a bounded LRU of recent IPs; the user-agent set is small and
	// long-lived, so it gets a much larger cap and is fully preloaded.
	cache := newCache(10000, 50000)

	// Keep the user_agents id sequence ahead of max(id) so inserts cannot hit a
	// primary-key collision after a restore or backfill.
	if err := reseedUserAgentSequence(context.Background(), db); err != nil {
		log.Error(fmt.Sprintf("failed to reseed user agent sequence: %v", err))
	}

	count, err := preloadUserAgentCache(context.Background(), db, cache)
	if err != nil {
		log.Error(fmt.Sprintf("failed to preload user agent cache: %v", err))
	} else {
		log.Info(fmt.Sprintf("user agent cache preloaded: %d entries", count))
	}

	// Rate limit ingest by source IP so a flood of unknown API keys is bounded
	// before it reaches the per-key limiter and the key-existence lookup. The
	// health check is left unthrottled so monitoring is unaffected.
	ipLimiter := ipRateLimit(cfg.IPRateLimit)
	bodyLimiter := bodyLimit(bodyLimitBytes(cfg.MaxInsert))

	h := logRequestHandler(db, geoIPDB, cache, cfg.RateLimit, cfg.MaxInsert, cfg.HashSecret)
	r.POST("/log-request", ipLimiter, bodyLimiter, h)
	r.POST("/requests", ipLimiter, bodyLimiter, h)
	r.GET("/health", checkHealth(db, startTime))
}

// bodyLimit caps the request body so an oversized payload cannot be read into
// memory. Once the limit is exceeded the bound reader errors, so the handler's
// JSON bind fails and the request is rejected with 400.
func bodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// ipRateLimit returns middleware that limits requests per source IP to
// requestsPerMinute, keyed by client IP.
func ipRateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := ratelimit.NewRateLimiter(requestsPerMinute)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if limiter.RateLimited(ip) {
			msg := "Too many requests."
			log.LogClientError(ip, "", msg)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"status": http.StatusTooManyRequests, "message": msg})
			return
		}
		c.Next()
	}
}

// healthTTL bounds how often the health check actually hits the database, so an
// unthrottled flood of /health requests cannot contend with ingest for the
// connection pool.
const healthTTL = time.Second

func checkHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	var (
		mu        sync.Mutex
		lastCheck time.Time
		lastErr   error
		checked   bool
	)

	return func(c *gin.Context) {
		uptime := int(time.Since(startTime).Seconds())

		mu.Lock()
		if !checked || time.Since(lastCheck) > healthTTL {
			lastErr = db.CheckConnection(c.Request.Context())
			lastCheck = time.Now()
			checked = true
		}
		err := lastErr
		mu.Unlock()

		if err != nil {
			log.Error(fmt.Sprintf("health check failed: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"health":         "unhealthy",
				"uptime_seconds": uptime,
				"database":       "unreachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"health":         "healthy",
			"uptime_seconds": uptime,
			"database":       "connected",
		})
	}
}
