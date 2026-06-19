package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/server/api/lib/env"
	"github.com/tom-draper/api-analytics/server/api/lib/log"
	"github.com/tom-draper/api-analytics/server/api/lib/routes"
	"github.com/tom-draper/api-analytics/server/database"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func getRateLimit() uint {
	return uint(env.GetIntegerEnvVariable("RATE_LIMIT", 100))
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.LogToFile(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.LogToFile("Starting api...")

	if godotenv.Load(".env") != nil {
		log.LogToFile("Failed to load .env file.")
	}

	err := database.LoadConfig()
	if err != nil {
		log.LogToFile("Failed to load database configuration: " + err.Error())
		return
	}

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")

	// Matches cors.Default() (allow all origins) but also permits the auth header
	// so cross-origin dashboards (e.g. via ?source= or self-hosting) can call
	// authenticated endpoints such as /api/data.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-AUTH-TOKEN", "API-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Limit a single IP's request logs to 100 per second
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: getRateLimit(),
	})
	rateLimiter := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	app.Use(rateLimiter)

	routes.RegisterRouter(r)

	if err := app.Run(":3000"); err != nil {
		log.LogToFile(fmt.Sprintf("Failed to run server: %v", err))
	}
}
