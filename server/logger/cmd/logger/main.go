package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/routes"
)

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

	app := setupRouter(db, geoIPDB, cfg, startTime)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: app,
		// Bound every stage of a connection so a slow or idle client cannot hold
		// resources open indefinitely (Slowloris). ReadTimeout covers the whole
		// request including the body; WriteTimeout covers handler + response.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
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

func setupRouter(db *database.DB, geoIPDB *geoip2.Reader, cfg *config.Config, startTime time.Time) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	// Only believe X-Forwarded-For from trusted proxies; otherwise c.ClientIP()
	// (the per-IP rate-limit key and client-error log source) would trust a
	// client-spoofable header, letting an attacker rotate fake IPs to bypass the
	// ingest rate limit.
	if err := app.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Error(fmt.Sprintf("failed to set trusted proxies: %v", err))
	}

	app.Use(corsMiddleware(cfg.AllowedOrigins))

	r := app.Group("/api")
	routes.RegisterRouter(r, db, geoIPDB, cfg, startTime)

	return app
}

// corsMiddleware allows all origins when allowedOrigins is empty (the historical
// default) and otherwise restricts CORS to the configured origins, matching the
// API service so both can be locked down with the same CORS_ALLOWED_ORIGINS.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		return cors.Default()
	}
	c := cors.DefaultConfig()
	c.AllowOrigins = allowedOrigins
	return cors.New(c)
}
