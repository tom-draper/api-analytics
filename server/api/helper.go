package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	supa "github.com/nedpals/supabase-go"
)

type User struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func GenAPIKeyHandler(supabase *supa.Client) gin.HandlerFunc {
	genAPIKey := func(c *gin.Context) {
		var row struct{} // Insert empty row, use default values
		var result []User
		err := supabase.DB.From("Users").Insert(row).Execute(&result)
		if err != nil {
			panic(err)
		}

		// Get API key auto generated from new row insertion
		apiKey := result[0].APIKey

		// Return API key
		c.JSON(200, gin.H{"value": apiKey})
	}

	return gin.HandlerFunc(genAPIKey)
}

type Request struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       int16  `json:"method"`
	StatusCode   int16  `json:"status_code"`
	ResponseTime int16  `json:"response_time"`
	Framework    int16  `json:"framework"`
}

func LogRequestHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect API request data sent via POST request
		var request Request
		if err := c.BindJSON(&request); err != nil {
			panic(err)
		}

		// Insert request data into database
		var result []interface{}
		err := supabase.DB.From("Requests").Insert(request).Execute(&result)
		if err != nil {
			panic(err)
		}

		// Return success response
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
	}

	return gin.HandlerFunc(logRequest)
}

func GetUserIDHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect API key sent via POST request
		var apiKey string
		if err := c.BindJSON(&apiKey); err != nil {
			panic(err)
		}

		// Fetch user ID corresponding with API key
		var result []struct {
			UserID string `json:"user_id"`
		}
		err := supabase.DB.From("Users").Select("user_id").Filter("api_key", "eq", apiKey).Execute(&result)
		if err != nil {
			panic(err)
		}

		userID := result[0].UserID

		// Return user ID
		c.JSON(200, gin.H{"value": userID})
	}

	return gin.HandlerFunc(logRequest)
}

type RequestRow struct {
	RequestID    int16  `json:"request_id" `
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       int16  `json:"method"`
	StatusCode   int16  `json:"status_code"`
	ResponseTime int16  `json:"response_time"`
	Framework    int16  `json:"framework"`
}

func GetDataHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect user ID sent via POST request
		var userID string
		if err := c.BindJSON(&userID); err != nil {
			panic(err)
		}

		fmt.Println(userID)

		// Fetch all API request data associated with this account
		var result []RequestRow
		err := supabase.DB.From("Requests").Select("*").Filter("user_id", "eq", userID).Execute(&result)
		if err != nil {
			panic(err)
		}

		fmt.Println(result)

		// Return API request data
		c.JSON(200, gin.H{"value": result})
	}

	return gin.HandlerFunc(logRequest)
}
