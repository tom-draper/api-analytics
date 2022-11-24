package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	supa "github.com/nedpals/supabase-go"
)

var (
	app *gin.Engine
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func registerRouter(r *gin.RouterGroup, supabase *supa.Client) {
	r.GET("/generate-api-key", GenAPIKeyHandler(supabase))
	r.POST("/log-request", LogRequestHandler(supabase))
}

func getDBLogin() (string, string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func init() {
	supabaseURL, supabaseKey := getDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	app = gin.New()

	r := app.Group("/api") // Vercel - must be /api/xxx

	r.Use(CORSMiddleware())
	registerRouter(r, supabase) // Register route
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
