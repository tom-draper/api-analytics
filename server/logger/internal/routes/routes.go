package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
)

func RegisterRouter(r *gin.RouterGroup, db *database.DB, geoIPDB *geoip2.Reader, cfg *config.Config, startTime time.Time) {
	cache := newCache(10000)

	count, err := preloadUserAgentCache(context.Background(), db, cache)
	if err != nil {
		log.Error(fmt.Sprintf("failed to preload user agent cache: %v", err))
	} else {
		log.Info(fmt.Sprintf("user agent cache preloaded: %d entries", count))
	}

	h := logRequestHandler(db, geoIPDB, cache, cfg.RateLimit, cfg.MaxInsert)
	r.POST("/log-request", h)
	r.POST("/requests", h)
	r.GET("/health", checkHealth(db, startTime))
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
