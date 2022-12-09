package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	supa "github.com/nedpals/supabase-go"
)

var (
	app *gin.Engine
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

type RequestData struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       int16  `json:"status"`
	ResponseTime int16  `json:"response_time"`
	Framework    string `json:"framework"`
}

type RequestInsert struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	UserAgent    string `json:"user_agent"`
	Method       int16  `json:"method"`
	Status       int16  `json:"status"`
	ResponseTime int16  `json:"response_time"`
	Framework    int16  `json:"framework"`
}

func methodMap(method string) (int16, error) {
	switch method {
	case "GET":
		return 0, nil
	case "POST":
		return 1, nil
	case "PUT":
		return 2, nil
	case "PATCH":
		return 3, nil
	case "DELETE":
		return 4, nil
	case "OPTIONS":
		return 5, nil
	case "CONNECT":
		return 6, nil
	case "HEAD":
		return 7, nil
	case "TRACE":
		return 8, nil
	default:
		return -1, fmt.Errorf("error: invalid method")
	}
}

func frameworkMap(framework string) (int16, error) {
	switch framework {
	case "FastAPI":
		return 0, nil
	case "Flask":
		return 1, nil
	case "Gin":
		return 2, nil
	case "Echo":
		return 3, nil
	case "Express":
		return 4, nil
	case "Fastify":
		return 5, nil
	case "Koa":
		return 6, nil
	case "Chi":
		return 7, nil
	case "Fiber":
		return 8, nil
	case "Actix":
		return 9, nil
	case "Axum":
		return 10, nil
	case "Tornado":
		return 11, nil
	case "Django":
		return 12, nil
	case "Rails":
		return 13, nil
	case "Laravel":
		return 14, nil
	default:
		return -1, fmt.Errorf("error: invalid framework")
	}
}

func LogRequestHandler(supabase *supa.Client) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect API request data sent via POST request
		var requestData RequestData
		if err := c.BindJSON(&requestData); err != nil {
			panic(err)
		}

		if requestData.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			method, err := methodMap(requestData.Method)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid method."})
				return
			}
			framework, err := frameworkMap(requestData.Framework)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid framework."})
				return
			}
			request := RequestInsert{
				APIKey:       requestData.APIKey,
				Path:         requestData.Path,
				Hostname:     requestData.Hostname,
				UserAgent:    requestData.UserAgent,
				Status:       requestData.Status,
				ResponseTime: requestData.ResponseTime,
				Method:       method,
				Framework:    framework,
			}
			// Insert request data into database
			var result []interface{}
			err = supabase.DB.From("Requests").Insert(request).Execute(&result)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Return success response
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
		}
	}

	return gin.HandlerFunc(logRequest)
}

func GetUserIDHandler(supabase *supa.Client) gin.HandlerFunc {
	getUserID := func(c *gin.Context) {
		// Collect API key sent via POST request
		apiKey := c.Param("apiKey")

		// Fetch user ID corresponding with API key
		var result []struct {
			UserID string `json:"user_id"`
		}
		err := supabase.DB.From("Users").Select("user_id").Eq("api_key", apiKey).Execute(&result)
		if err != nil {
			panic(err)
		}

		userID := result[0].UserID

		// Return user ID
		c.JSON(200, gin.H{"value": userID})
	}

	return gin.HandlerFunc(getUserID)
}

type RequestRow struct {
	APIKey       string    `json:"api_key"`
	RequestID    int16     `json:"request_id" `
	Hostname     string    `json:"hostname"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	Framework    int16     `json:"framework"`
	CreatedAt    time.Time `json:"created_at"`
}

func GetDataHandler(supabase *supa.Client) gin.HandlerFunc {
	getData := func(c *gin.Context) {
		// Collect user ID sent via POST request
		userID := c.Param("userID")

		// Fetch all API request data associated with this account
		var result []struct {
			Requests []RequestRow `json:"Requests"`
			APIKey   string       `json:"api_key"`
		}
		err := supabase.DB.From("Users").Select("api_key, Requests!inner(*)").Eq("user_id", userID).Execute(&result)
		if err != nil {
			panic(err)
		}

		// Return API request data
		c.JSON(200, gin.H{"value": result[0].Requests})
	}

	return gin.HandlerFunc(getData)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RegisterRouter(r *gin.RouterGroup, supabase *supa.Client) {
	r.GET("/generate-api-key", GenAPIKeyHandler(supabase))
	r.POST("/log-request", LogRequestHandler(supabase))
	r.GET("/user-id/:apiKey", GetUserIDHandler(supabase))
	r.GET("/data/:userID", GetDataHandler(supabase))
}

func getDBLogin() (string, string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func init() {
	supabaseURL, supabaseKey := getDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	app = gin.New()

	r := app.Group("/api") // Vercel - must be /api/xxx

	// r.Use(CORSMiddleware())
	r.Use(cors.Default())
	RegisterRouter(r, supabase) // Register route
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
