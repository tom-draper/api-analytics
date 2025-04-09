package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
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
    data := map[string]string{
        "message": "Hello, World!",
    }
    return c.JSON(http.StatusOK, data)
}

func main() {
	apiKey := getAPIKey()

	e := echo.New()

	e.Use(analytics.Analytics(apiKey))

	e.GET("/", root)
	e.Start(":8080")
}
