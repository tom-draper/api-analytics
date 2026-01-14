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
	// Initialise logger first
	if err := log.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Close()

	defer func() {
		if err := recover(); err != nil {
			log.Info(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.Info("Starting api...")

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Info(fmt.Sprintf("Configuration error: %v", err))
		return
	}

	db, err := database.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Info("Failed to initialize database: " + err.Error())
		return
	}
	defer db.Close()

	app := setupRouter(db, cfg)

	port := cfg.Port
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info(fmt.Sprintf("listen: %s\n", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Info(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	log.Info("Server exiting")
}

func setupRouter(db *database.DB, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")

	r.Use(cors.Default())

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

	routes.RegisterRouter(r, db, cfg)

	return app
}

func rateLimitKey(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}
