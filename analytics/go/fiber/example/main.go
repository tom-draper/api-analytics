package main

import (
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"

	"github.com/gofiber/fiber/v2"
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

func root(c *fiber.Ctx) error {
	jsonData := []byte(`{"message": "Hello World!"}`)
	return c.SendString(string(jsonData))
}

func main() {
	apiKey := getAPIKey()

	app := fiber.New()

	app.Use(analytics.Analytics(apiKey))

	app.Get("/", root)
	app.Listen(":8080")
}
