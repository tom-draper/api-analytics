package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
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
		c.JSON(http.StatusOK, apiKey)
	}

	return gin.HandlerFunc(genAPIKey)
}

type RequestData struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
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
	IPAddress    string `json:"ip_address"`
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
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		fmt.Println(requestData)

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
				IPAddress:    requestData.IPAddress,
				UserAgent:    requestData.UserAgent,
				Status:       requestData.Status,
				ResponseTime: requestData.ResponseTime,
				Method:       method,
				Framework:    framework,
			}
			fmt.Println(request)
			// Insert request data into database
			// var result []interface{}
			// err = supabase.DB.From("Requests").Insert(request).Execute(&result)
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
	// Get user ID associated with API key
	getUserID := func(c *gin.Context) {
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
		c.JSON(http.StatusOK, userID)
	}

	return gin.HandlerFunc(getUserID)
}

type RequestRow struct {
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	CreatedAt    time.Time `json:"created_at"`
}

func GetUserRequestsHandler(supabase *supa.Client) gin.HandlerFunc {
	getUserRequests := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch all API request data associated with this account
		var result []struct {
			Requests []RequestRow `json:"Requests"`
			APIKey   string       `json:"api_key"`
		}
		err := supabase.DB.From("Users").Select("api_key, Requests!inner(hostname, ip_address, path, user_agent, method, status, response_time, created_at)").Eq("user_id", userID).Execute(&result)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, result[0].Requests)
	}

	return gin.HandlerFunc(getUserRequests)
}

func GetDataHandler(supabase *supa.Client) gin.HandlerFunc {
	getData := func(c *gin.Context) {
		apiKey := c.GetHeader("API-Key")

		// Fetch all API request data associated with this account
		var result []RequestRow
		err := supabase.DB.From("Requests").Select("hostname", "ip_address", "path", "user_agent", "method", "created_at", "response_time", "framework", "status").Eq("api_key", apiKey).Execute(&result)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(getData)
}

func DeleteDataHandler(supabase *supa.Client) gin.HandlerFunc {
	deleteData := func(c *gin.Context) {
		apiKey := c.Param("apiKey")

		// Delete all API request data associated with this account
		var requestResult []RequestRow
		err := supabase.DB.From("Requests").Delete().Eq("api_key", apiKey).Execute(&requestResult)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// // Delete user account record
		var userResult []User
		err = supabase.DB.From("Users").Delete().Eq("api_key", apiKey).Execute(&userResult)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
	}

	return gin.HandlerFunc(deleteData)
}

type MonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func GetUserMonitorHandler(supabase *supa.Client) gin.HandlerFunc {
	getUserMonitor := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch all ping data associated with this account
		var result []struct {
			Monitor []MonitorRow `json:"monitor"`
			APIKey  string       `json:"api_key"`
		}
		err := supabase.DB.From("Users").Select("api_key, Monitor!inner(url, secure, ping, created_at)").Eq("user_id", userID).Execute(&result)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, result[0])
	}

	return gin.HandlerFunc(getUserMonitor)
}

type MonitorInsertRow struct {
	APIKey string `json:"api_key"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

func InsertUserMonitorHandler(supabase *supa.Client) gin.HandlerFunc {
	insertUserMonitor := func(c *gin.Context) {
		var monitor MonitorInsertRow
		if err := c.BindJSON(&monitor); err != nil {
			panic(err)
		}

		if monitor.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			// Insert request data into database
			var result []interface{}
			err := supabase.DB.From("Monitor").Insert(monitor).Execute(&result)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Return success response
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor added successfully."})
		}
	}

	return gin.HandlerFunc(insertUserMonitor)
}

type PingsRow struct {
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func GetUserPingsHandler(supabase *supa.Client) gin.HandlerFunc {
	getData := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch all ping data associated with this account
		var result []struct {
			Pings  []PingsRow `json:"pings"`
			APIKey string     `json:"api_key"`
		}
		err := supabase.DB.From("Users").Select("api_key, Pings!inner(url, response_time, status, created_at)").Eq("user_id", userID).Execute(&result)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, result[0])
	}

	return gin.HandlerFunc(getData)
}

func RegisterRouter(r *gin.RouterGroup, supabase *supa.Client) {
	r.GET("/generate-api-key", GenAPIKeyHandler(supabase))
	r.POST("/log-request", LogRequestHandler(supabase))
	r.GET("/user-id/:apiKey", GetUserIDHandler(supabase))
	r.GET("/requests/:userID", GetUserRequestsHandler(supabase))
	r.GET("/delete/:apiKey", DeleteDataHandler(supabase))
	r.GET("/pings/:userID", GetUserPingsHandler(supabase))
	r.GET("/monitor/:userID", GetUserMonitorHandler(supabase))
	r.POST("/add-monitor", InsertUserMonitorHandler(supabase))
	r.GET("/data", GetDataHandler(supabase))
}

func getDBLogin() (string, string) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func init() {
	supabaseURL, supabaseKey := getDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	gin.SetMode(gin.ReleaseMode)
	app = gin.New()

	// Add rate limiter
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 10000,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})

	r := app.Group("/api", mw) // Vercel - must be /api/xxx

	r.Use(cors.Default())
	RegisterRouter(r, supabase)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
