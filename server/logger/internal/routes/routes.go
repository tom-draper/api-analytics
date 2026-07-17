package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"
)

func RegisterRouter(r *gin.RouterGroup, db *database.DB, geoIPDB *geoip2.Reader, cfg *config.Config, startTime time.Time) {
	cache := newCache(10000)

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

	h := logRequestHandler(db, geoIPDB, cache, cfg.RateLimit, cfg.MaxInsert)
	r.POST("/log-request", ipLimiter, h)
	r.POST("/requests", ipLimiter, h)
	r.GET("/health", checkHealth(db, startTime))
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

func checkHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		uptime := int(time.Since(startTime).Seconds())

		if err := db.CheckConnection(ctx); err != nil {
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
