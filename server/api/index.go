package api

import (
	"net/http"
	"os"
	"time"

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

type User struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func GenAPIKeyHandler(supabase *supa.Client) gin.HandlerFunc {
	genAPIKey := func(c *gin.Context) {
		var row struct{} // Insert empty row, use default values
		var result []User
		err := supabase.DB.From("Users").Insert(row).Execute(&result)
		if err != nil {
			panic(err)
		}
		apiKey := result[0].APIKey
		c.JSON(200, gin.H{"status": 200, "api-key": apiKey})
	}

	return gin.HandlerFunc(genAPIKey)
}

type Request struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       int16  `json:"method"`
	StatusCode   int16  `json:"status_code"`
	ResponseTime int16  `json:"response_time"`
	Framework    int16  `json:"framework"`
}

func LogRequestHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		var request Request
		if err := c.BindJSON(&request); err != nil {
			panic(err)
		}

		var result []interface{}
		if err := supabase.DB.From("Requests").Insert(request).Execute(&result); err != nil {
			panic(err)
		}

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
	}

	return gin.HandlerFunc(logRequest)
}

func registerRouter(r *gin.RouterGroup, supabase *supa.Client) {
	r.GET("/gen-api-key", GenAPIKeyHandler(supabase))
	r.POST("/request", LogRequestHandler(supabase))
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
