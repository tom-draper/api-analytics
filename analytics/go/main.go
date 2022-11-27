package main

import (
	analytics "analytics/analytics"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getAPIKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	apiKey := os.Getenv("API_KEY")
	return apiKey
}

func root(c *gin.Context) {
	jsonData := []byte(`{"message": "Hello World!"}`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	apiKey := getAPIKey()

	router := gin.Default()

	router.Use(analytics.Analytics(apiKey))

	router.GET("/", root)
	router.Run("localhost:8080")
}
