package main

import (
	"os"
	api "server/api"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

func GetDBLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func main() {
	supabaseURL, supabaseKey := GetDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	app := gin.New()

	r := app.Group("/api") // Vercel - must be /api/xxx

	r.Use(api.CORSMiddleware())

	// r.GET("/generate-api-key", api.GenAPIKeyHandler(supabase))
	// r.POST("/log-request", api.LogRequestHandler(supabase))
	// r.GET("/user-id/:apiKey", api.GetUserIDHandler(supabase))
	// r.GET("/data/:userID", api.GetDataHandler(supabase))

	api.RegisterRouter(r, supabase) // Register route
	app.Run("localhost:8080")
}
