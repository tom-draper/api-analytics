# Chi Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/chi
```

```go
package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/chi"

	chi "github.com/go-chi/chi/v5"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"message": "Hello World!"}`)
	w.Write(jsonData)
}

func main() {
	router := chi.NewRouter()

	router.Use(analytics.Analytics(<api_key>))

	router.GET("/", root)
	router.Run("localhost:8080")
}
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.
