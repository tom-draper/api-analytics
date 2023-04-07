package main

import (
	"io"
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	app *gin.Engine
)

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}
func init() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")
	r.Use(cors.Default())

	// Add rate limiter
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 5,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	r.Use(mw)

	r.GET("/data", getDataHandler)

	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
}

func getDataHandler() []byte {
	resp, err := http.Get("https://apianalytics-server.com/api/data")
	if err != nil {
		return []byte{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body
}
