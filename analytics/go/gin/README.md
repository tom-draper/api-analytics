# Gin Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/gin
```

```go
package main

import (
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
	"net/http"

	"github.com/gin-gonic/gin"
)

func root(c *gin.Context) {
	jsonData := []byte(`{"message": "Hello World!"}`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := gin.Default()
	
	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Run("localhost:8080")
}
```

### 3. View your analytics

Your API will now log and store incoming request data on all valid routes. Your logged data can be viewed using two methods: through visualizations and stats on our dashboard, or accessed directly via our data API.

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard. We recommend generating a new API key for each additional API you want analytics for.

#### Dashboard

Head to https://my-api-analytics.vercel.app/dashboard and paste in your API key to access your dashboard.

Demo: https://my-api-analytics.vercel.app/dashboard/demo

![Dashboard](https://user-images.githubusercontent.com/41476809/208440202-966a6930-3d2e-40c5-afc7-2fd0107d6b4f.png)

#### Data API

Logged data for all requests can be accessed via our API. Simply send a GET request to `https://api-analytics-server.vercel.app/api/data` with your API key set as `API-Key` in headers.

```py
import requests

headers = {
 "API-Key": <api_key>
}

response = requests.get("https://api-analytics-server.vercel.app/api/data", headers=headers)
print(response.json())
```

## Monitoring (coming soon)

Opt-in active API monitoring is coming soon. Our servers will regularly ping your API endpoints to monitor uptime and response time. Optional email alerts to notify you when your endpoints are down can be subscribed to.

![Monitoring](https://user-images.githubusercontent.com/41476809/208298759-f937b668-2d86-43a2-b615-6b7f0b2bc20c.png)

## Data and Security

All data is stored securely in compliance with The EU General Data Protection Regulation (GDPR).

For any given request to your API, data recorded is limited to:
 - Path requested by client
 - Client IP address
 - Client operating system
 - Client browser
 - Request method (GET, POST, PUT, etc.)
 - Time of request
 - Status code
 - Response time
 - API hostname
 - API framework (FastAPI, Flask, Express etc.)

Data collected is only ever used to populate your analytics dashboard. Your data is anonymous, with the API key the only link between you and you API's analytics. Should you lose your API key, you will have no method to access your API analytics. Inactive API keys (> 1 year) and its associated API request data may be deleted.

### Delete Data

At any time, you can delete all stored data associated with your API key by going to https://my-api-analytics.vercel.app/delete and entering your API key.
