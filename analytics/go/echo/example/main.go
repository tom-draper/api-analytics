package main

import (
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
	analytics "github.com/tom-draper/api-analytics/analytics/go/echo"

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

func root(c echo.Context) error {
	jsonData := []byte(`{"message": "Hello World!"}`)
	return c.JSON(http.StatusOK, jsonData)
}

func main() {
	apiKey := getAPIKey()

	router := echo.New()

	router.Use(analytics.Analytics(apiKey))

	router.GET("/", root)
	router.Start(":8080")
}
