package lib

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	_ "github.com/lib/pq"
	"github.com/oschwald/geoip2-golang"

	"github.com/gin-gonic/gin"
)

type User struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func genAPIKeyHandler(db *sql.DB) gin.HandlerFunc {
	genAPIKey := func(c *gin.Context) {
		// Fetch all API request data associated with this account
		query := fmt.Sprintf("INSERT INTO users (api_key, user_id, created_at) VALUES (gen_random_uuid(), gen_random_uuid(), NOW()) RETURNING api_key;")
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
			return
		}

		// Get API key auto generated from new row insertion
		rows.Next()
		var apiKey string
		err = rows.Scan(&apiKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
			return
		}

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

type RequestRow struct {
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	Location     string `json:"location"`
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

func getCountryCode(IPAddress string) (string, error) {
	var location string
	if IPAddress != "" {
		db, err := geoip2.Open("GeoLite2-Country.mmdb")
		if err != nil {
			return location, err
		}
		defer db.Close()

		ip := net.ParseIP(IPAddress)
		record, err := db.Country(ip)
		if err != nil {
			return location, err
		}
		location = record.Country.IsoCode
	}
	return location, nil
}

func logRequestHandler(db *sql.DB) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect API request data sent via POST request
		var requestData RequestData
		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			fmt.Println(err)
			return
		}

		if requestData.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			location, _ := getCountryCode(requestData.IPAddress)

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

			request := RequestRow{
				APIKey:       requestData.APIKey,
				Path:         requestData.Path,
				Hostname:     requestData.Hostname,
				IPAddress:    requestData.IPAddress,
				UserAgent:    requestData.UserAgent,
				Status:       requestData.Status,
				ResponseTime: requestData.ResponseTime,
				Method:       method,
				Framework:    framework,
				Location:     location,
			}

			// Insert request data into database
			query := fmt.Sprintf("INSERT INTO requests (api_key, path, hostname, ip_address, user_agent, status, response_time, method, framework, location, created_at) VALUES ('%s', '%s', '%s', '%s', '%s', %d, %d, %d, %d, '%s', NOW());", request.APIKey, request.Path, request.Hostname, request.IPAddress, request.UserAgent, request.Status, request.ResponseTime, request.Method, request.Framework, request.Location)
			fmt.Println(query)
			_, err = db.Query(query)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				fmt.Println(err)
				return
			}

			// Return success response
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
		}
	}

	return gin.HandlerFunc(logRequest)
}

func getUserIDHandler(db *sql.DB) gin.HandlerFunc {
	// Get user ID associated with API key
	getUserID := func(c *gin.Context) {
		apiKey := c.Param("apiKey")

		// Fetch user ID corresponding with API key
		query := fmt.Sprintf("SELECT user_id FROM users WHERE api_key = '%s';", apiKey)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// API key is primary key so assumed only one row returned
		rows.Next()
		var userID string
		err = rows.Scan(&userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// Return user ID
		c.JSON(http.StatusOK, userID)
	}

	return gin.HandlerFunc(getUserID)
}

type PublicRequestRow struct {
	Hostname     string         `json:"hostname"`
	IPAddress    sql.NullString `json:"ip_address"`
	Path         string         `json:"path"`
	UserAgent    sql.NullString `json:"user_agent"`
	Method       int16          `json:"method"`
	Status       int16          `json:"status"`
	ResponseTime int16          `json:"response_time"`
	Location     string         `json:"location"`
	CreatedAt    time.Time      `json:"created_at"`
}

func getUserRequestsHandler(db *sql.DB) gin.HandlerFunc {
	getUserRequests := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch user ID corresponding with API key
		query := fmt.Sprintf("SELECT ip_address, path, user_agent, method, response_time, status, location, requests.created_at FROM requests INNER JOIN users ON users.api_key = requests.api_key WHERE users.user_id = '%s';", userID)
		fmt.Println(query)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Read data into compact list of lists to return
		cols := []interface{}{"ip_address", "path", "user_agent", "method", "response_time", "status", "location", "created_at"}
		requests := buildRequestDataCompact(rows, cols)

		// Return API request data
		c.JSON(http.StatusOK, requests)
	}

	return gin.HandlerFunc(getUserRequests)
}

func buildRequestDataCompact(rows *sql.Rows, cols []interface{}) [][]interface{} {
	// First value in list holds column names
	requests := [][]interface{}{cols}
	request := new(PublicRequestRow) // Reused to avoid repeated memory allocation
	for rows.Next() {
		err := rows.Scan(&request.IPAddress, &request.Path, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
		if request.Location == "  " {
			request.Location = ""
		}
		if err == nil {
			requests = append(requests, []interface{}{request.IPAddress.String, request.Path, request.UserAgent.String, request.Method, request.ResponseTime, request.Status, request.Location, request.CreatedAt})
		}
	}
	return requests
}

func getDataHandler(db *sql.DB) gin.HandlerFunc {
	getData := func(c *gin.Context) {
		apiKey := c.GetHeader("X-AUTH-TOKEN")
		if apiKey == "" {
			// Check old (deprecated) identifier
			apiKey = c.GetHeader("API-Key")
		}

		// Fetch all API request data associated with this account
		query := fmt.Sprintf("SELECT hostname, ip_address, path, user_agent, method, response_time, status, location, created_at FROM requests WHERE api_key = '%s';", apiKey)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		// Read data into list of objects to return
		requests := buildRequestData(rows)

		// Return API request data
		c.JSON(http.StatusOK, requests)
	}

	return gin.HandlerFunc(getData)
}

func buildRequestData(rows *sql.Rows) []PublicRequestRow {
	requests := make([]PublicRequestRow, 0)
	for rows.Next() {
		request := new(PublicRequestRow)
		err := rows.Scan(&request.Hostname, &request.IPAddress, &request.Path, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
		if err == nil {
			requests = append(requests, *request)
		}
	}
	return requests
}

func deleteUserRequests(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete all user's API request data
	query := fmt.Sprintf("DELETE FROM requests WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserAccount(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete user account record
	query := fmt.Sprintf("DELETE FROM users WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserMonitors(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete all user's monitored urls
	query := fmt.Sprintf("DELETE FROM monitor WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserPings(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete all user's recorded pings to all monitored urls
	query := fmt.Sprintf("DELETE FROM pings WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteDataHandler(db *sql.DB) gin.HandlerFunc {
	deleteData := func(c *gin.Context) {
		apiKey := c.Param("apiKey")

		var err error
		err = deleteUserRequests(apiKey, c, db)
		if err != nil {
			return
		}
		err = deleteUserAccount(apiKey, c, db)
		if err != nil {
			return
		}
		err = deleteUserMonitors(apiKey, c, db)
		if err != nil {
			return
		}
		err = deleteUserPings(apiKey, c, db)
		if err != nil {
			return
		}

		// Return API request data
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
	}

	return gin.HandlerFunc(deleteData)
}

type PublicMonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserMonitorHandler(db *sql.DB) gin.HandlerFunc {
	getUserMonitor := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch user ID corresponding with API key
		query := fmt.Sprintf("SELECT url, secure, ping, monitor.created_at FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = '%s';", userID)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Read monitors into list to return
		monitors := make([]PublicMonitorRow, 0)
		for rows.Next() {
			monitor := new(PublicMonitorRow)
			err := rows.Scan(&monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
			if err == nil {
				monitors = append(monitors, *monitor)
			}
		}

		// Return API request data
		c.JSON(http.StatusOK, monitors)
	}

	return gin.HandlerFunc(getUserMonitor)
}

type MonitorRow struct {
	APIKey string `json:"api_key"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

func insertUserMonitorHandler(db *sql.DB) gin.HandlerFunc {
	insertUserMonitor := func(c *gin.Context) {
		var monitor MonitorRow
		if err := c.BindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if monitor.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			query := fmt.Sprintf("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ('%s', '%s', %t, %t, NOW())", monitor.APIKey, monitor.URL, monitor.Secure, monitor.Ping)
			_, err := db.Query(query)
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

func deleteMonitor(apiKey string, url string, c *gin.Context, db *sql.DB) error {
	// Delete user's monitor to this specific url
	query := fmt.Sprintf("DELETE FROM monitor WHERE api_key = '%s' AND url = '%s';", apiKey, url)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return err
	}
	return nil
}

func deletePings(apiKey string, url string, c *gin.Context, db *sql.DB) error {
	// Delete user's recorded pings to monitored url
	query := fmt.Sprintf("DELETE FROM pings WHERE api_key = '%s' AND url = '%s';", apiKey, url)
	_, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return err
	}
	return nil
}

func deleteUserMonitorHandler(db *sql.DB) gin.HandlerFunc {
	deleteUserMonitor := func(c *gin.Context) {
		var body struct {
			APIKey string `json:"api_key"`
			URL    string `json:"url"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if body.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			var err error
			err = deleteMonitor(body.APIKey, body.URL, c, db)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}
			err = deletePings(body.APIKey, body.URL, c, db)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Return success response
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Monitor deleted successfully."})
		}
	}

	return gin.HandlerFunc(deleteUserMonitor)
}

type PublicPingsRow struct {
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func getUserPingsHandler(db *sql.DB) gin.HandlerFunc {
	getData := func(c *gin.Context) {
		userID := c.Param("userID")

		// Fetch user ID corresponding with API key
		query := fmt.Sprintf("SELECT url, response_time, status, pings.created_at FROM pings INNER JOIN users ON users.api_key = pings.api_key WHERE users.user_id = '%s';", userID)
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Read pings into list to return
		pings := make([]PublicPingsRow, 0)
		for rows.Next() {
			ping := new(PublicPingsRow)
			err := rows.Scan(&ping.URL, &ping.ResponseTime, &ping.Status, &ping.CreatedAt)
			if err == nil {
				pings = append(pings, *ping)
			}
		}

		// Return API request data
		c.JSON(http.StatusOK, pings)
	}

	return gin.HandlerFunc(getData)
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func RegisterRouter(r *gin.RouterGroup, db *sql.DB) {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 10,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	r.POST("/log-request", mw, logRequestHandler(db))
	r.GET("/generate-api-key", genAPIKeyHandler(db))
	r.GET("/user-id/:apiKey", getUserIDHandler(db))
	r.GET("/requests/:userID", getUserRequestsHandler(db))
	r.GET("/delete/:apiKey", deleteDataHandler(db))
	r.GET("/monitor/pings/:userID", getUserPingsHandler(db))
	r.POST("/monitor/add", insertUserMonitorHandler(db))
	r.POST("/monitor/delete", deleteUserMonitorHandler(db))
	r.GET("/data", getDataHandler(db))
}
