package main

import (
	"os"
	"server/api"

	"github.com/gin-contrib/cors"
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

	r.Use(cors.Default())

	api.RegisterRouter(r, supabase) // Register route
	app.Run(":8080")
}
