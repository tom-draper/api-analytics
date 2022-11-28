package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/chi"

	chi "github.com/go-chi/chi/v5"
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

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"message": "Hello World!"}`)
	w.Write(jsonData)
}

func main() {
	apiKey := getAPIKey()

	router := chi.NewRouter()

	router.Use(analytics.Analytics(apiKey))

	router.Get("/", root)
	http.ListenAndServe(":3000", router)
}
