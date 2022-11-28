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

func root(c echo.Context) {
	jsonData := []byte(`{"message": "Hello World!"}`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	apiKey := getAPIKey()

	e := echo.New()

	e.Use(analytics.Analytics(apiKey))

	e.GET("/", root)

	e.Logger.Fatal(e.Start(":8080"))
}
