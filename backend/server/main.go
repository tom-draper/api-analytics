package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"os"

	supa "github.com/nedpals/supabase-go"
)

func main() {
	supabaseURL, supabaseKey := getDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	router := gin.Default()
	router.GET("/gen-api-key", genAPIKeyHandler(supabase))
	router.POST("/request", logRequestHandler(supabase))
	router.Run("localhost:8080")
}

func getDBLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

type User struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func genAPIKeyHandler(supabase *supa.Client) gin.HandlerFunc {
	genAPIKey := func(c *gin.Context) {
		var row struct{} // Insert empty row, use default values
		var result []User
		err := supabase.DB.From("Users").Insert(row).Execute(&result)
		if err != nil {
			panic(err)
		}
		apiKey := result[0].APIKey
		fmt.Println(apiKey)
		c.JSON(200, gin.H{"status": 200, "api-key": apiKey})
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
}

func logRequestHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		var request Request
		if err := c.BindJSON(&request); err != nil {
			panic(err)
		}
		fmt.Println(request)

		var result []interface{}
		err := supabase.DB.From("Requests").Insert(request).Execute(&result)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})

	}

	return gin.HandlerFunc(logRequest)
}
