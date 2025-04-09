package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"

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

func root(c * gin.Context) {
    data := map[string]string{
        "message": "Hello, World!",
    }
    c.JSON(http.StatusOK, data)
}

func main() {
	apiKey := getAPIKey()

	r := gin.Default()

	r.Use(analytics.Analytics(apiKey))

	r.GET("/", root)
	r.Run(":8080")
}
