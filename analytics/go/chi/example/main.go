package main

import (
	"encoding/json"
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
    data := map[string]string{
        "message": "Hello, World!",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    err := json.NewEncoder(w).Encode(data)
    if err != nil {
        http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
        return
    }
}

func main() {
	apiKey := getAPIKey()

	r := chi.NewRouter()

	r.Use(analytics.Analytics(apiKey))

	r.Get("/", root)
	http.ListenAndServe(":8080", r)
}
