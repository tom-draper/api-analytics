package api

import (
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
		apiKey := result[0].APIKey
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
		var request Request
		if err := c.BindJSON(&request); err != nil {
			panic(err)
		}

		var result []interface{}
		if err := supabase.DB.From("Requests").Insert(request).Execute(&result); err != nil {
			panic(err)
		}

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
	}

	return gin.HandlerFunc(logRequest)
}
