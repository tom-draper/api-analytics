package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

const (
	P1 PrivacyLevel = iota
	P2
	P3
)

type PrivacyLevel int

type RequestData struct {
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       int16  `json:"status"`
	Referrer     string `json:"referrer"`
	ResponseTime int16  `json:"response_time"`
	UserID       string `json:"user_id"`
	CreatedAt    string `json:"created_at"`
}

type Payload struct {
	APIKey       string        `json:"api_key"`
	Requests     []RequestData `json:"requests"`
	Framework    string        `json:"framework"`
	PrivacyLevel PrivacyLevel  `json:"privacy_level"`
}

type ProcessedRequest struct {
	Path         string
	Hostname     string
	IPAddress    *string
	UserHash     string
	Referrer     string
	Status       int16
	ResponseTime int16
	Method       int16
	Framework    int16
	Location     string
	UserID       string
	UserAgent    string
	CreatedAt    time.Time
	UserAgentID  int
}

func main() {
	if err := log.Init(); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer log.Close()

	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("application crashed: %v", r))
		}
	}()

	startTime := time.Now()
	log.Info("starting logger...")

	cfg, err := config.Load()
	if err != nil {
		log.Error(fmt.Sprintf("configuration error: %v", err))
		return
	}

	db, err := database.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Error(fmt.Sprintf("failed to initialize database: %v", err))
		return
	}
	defer db.Close()
	log.Info("database connection pool initialized")

	geoIPDB, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Error(fmt.Sprintf("failed to open GeoIP db: %v", err))
		log.Error("location data will be unavailable")
	}
	defer func() {
		if geoIPDB != nil {
			geoIPDB.Close()
		}
	}()

	cache := &Cache{
		userAgentMap: make(map[string]int),
		geoIPMap:     make(map[string]*geoIPEntry),
		maxSize:      10000,
	}

	count, err := preloadUserAgentCache(context.Background(), db, cache)
	if err != nil {
		log.Error(fmt.Sprintf("failed to preload user agent cache: %v", err))
	} else {
		log.Info(fmt.Sprintf("user agent cache preloaded: %d entries", count))
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(cors.Default())

	handler := logRequestHandler(db, geoIPDB, cache, cfg.RateLimit, cfg.MaxInsert)
	router.POST("/api/log-request", handler)
	router.POST("/api/requests", handler)
	router.GET("/api/health", checkHealth(db, startTime))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info(fmt.Sprintf("server listening on port %d", cfg.Port))
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info(fmt.Sprintf("received signal: %v, shutting down...", sig))
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("server failed: %v", err))
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("server forced shutdown: %v", err))
		return
	}

	log.Info("server exited")
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
