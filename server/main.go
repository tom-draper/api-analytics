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

	router := gin.Default()

	router.Use(api.CORSMiddleware())
	router.GET("/generate-api-key", api.GenAPIKeyHandler(supabase))
	router.POST("/log-request", api.LogRequestHandler(supabase))
	router.POST("/user-id", api.GetUserIDHandler(supabase))
	router.POST("/data", api.GetDataHandler(supabase))

	router.Run("localhost:8080")
}
