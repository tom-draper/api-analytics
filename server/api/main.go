package main

import (
	"fmt"
	"net/http"
	"time"
	"fmt"

	"github.com/tom-draper/api-analytics/server/api/lib/log"
	"github.com/tom-draper/api-analytics/server/api/lib/routes"
	"github.com/tom-draper/api-analytics/server/api/lib/environment"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func getRateLimit() int {
	GetIntegerEnvVariable("RATE_LIMIT", 100)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.LogToFile(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.LogToFile("Starting api...")

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")

	r.Use(cors.Default())

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
