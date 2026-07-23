package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tom-draper/api-analytics/server/api/internal/config"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/api/internal/routes"
	"github.com/tom-draper/api-analytics/server/database"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	log.Info("starting api...")

	// Load and validate configuration
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

	app := setupRouter(db, cfg, startTime)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: app,
		// Bound every stage of a connection so a slow or idle client cannot hold
		// resources open indefinitely (Slowloris). ReadTimeout is generous because
		// the dashboard full-load response can take a while to stream.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
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

func setupRouter(db *database.DB, cfg *config.Config, startTime time.Time) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	// Only believe X-Forwarded-For from trusted proxies; otherwise c.ClientIP()
	// (the rate-limit key and log source) would trust a client-spoofable header,
	// letting an attacker rotate fake IPs to bypass the per-IP rate limit.
	if err := app.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Error(fmt.Sprintf("failed to set trusted proxies: %v", err))
	}

	// Middleware must be registered before the route group is created. A group
	// copies the engine's handler chain as it stands at that moment, so
	// anything added afterwards never runs for the group's routes.
	//
	// CORS goes first so that it answers preflight requests and aborts before
	// they consume rate limit budget. Registering it on the engine rather than
	// the group also means it covers requests that match no route, which is
	// what a preflight is until the browser is told the real method is allowed.
	app.Use(corsMiddleware(cfg.AllowedOrigins))

	// Limit a single IP's request logs per second
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: cfg.RateLimit,
	})
	ratelimiter := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      rateLimitKey,
	})
	app.Use(ratelimiter)

	r := app.Group("/api")

	routes.RegisterRouter(r, db, cfg, startTime)

	return app
}

// corsMiddleware allows all origins when allowedOrigins is empty (the historical
// default) and otherwise restricts CORS to the configured origins, so a hosted
// deployment can lock the dashboard API down to its own front-end.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		return cors.Default()
	}
	c := cors.DefaultConfig()
	c.AllowOrigins = allowedOrigins
	return cors.New(c)
}

func rateLimitKey(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}
