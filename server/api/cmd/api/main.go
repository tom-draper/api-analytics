package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tom-draper/api-analytics/server/api/internal/env"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/api/internal/routes"
	"github.com/tom-draper/api-analytics/server/database"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func rateLimitKey(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func getRateLimit() uint {
	return uint(env.GetIntegerEnvVariable("RATE_LIMIT", 100))
}

func setupRouter(db *database.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")

	r.Use(cors.Default())

	// Limit a single IP's request logs to 100 per second
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: getRateLimit(),
	})
	ratelimiter := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      rateLimitKey,
	})
	app.Use(ratelimiter)

	routes.RegisterRouter(r, db)

	return app
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Info(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.Init()
	log.Info("Starting api...")

	if err := env.LoadEnv(); err != nil {
		log.Info(err.Error())
	}

	pgURL := env.GetEnvVariable("POSTGRES_URL", "")

	db, err := database.New(context.Background(), pgURL)
	if err != nil {
		log.Info("Failed to initialize database: " + err.Error())
		return
	}

	app := setupRouter(db)

	port := env.GetIntegerEnvVariable("PORT", 3000)
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