package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func genAPIKeyHandler(db *sql.DB) gin.HandlerFunc {
	genAPIKey := func(c *gin.Context) {
		// Fetch all API request data associated with this account
		query := "INSERT INTO users (api_key, user_id, created_at) VALUES (gen_random_uuid(), gen_random_uuid(), NOW()) RETURNING api_key;"
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
		err := rows.Scan(&request.Hostname, &request.IPAddress.String, &request.Path, &request.UserAgent.String, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
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

type Monitor struct {
	UserID string `json:"user_id"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

func insertUserMonitorHandler(db *sql.DB) gin.HandlerFunc {
	insertUserMonitor := func(c *gin.Context) {
		var monitor Monitor
		if err := c.BindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if monitor.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		} else {
			// Get API key from user ID
			query := fmt.Sprintf("SELECT api_key FROM users WHERE user_id = '%s';", monitor.UserID)
			rows, err := db.Query(query)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}
			rows.Next()
			var apiKey string
			err = rows.Scan(&apiKey)
			if err != nil || apiKey == "" {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Insert new monitor into database
			query = fmt.Sprintf("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ('%s', '%s', %t, %t, NOW())", apiKey, monitor.URL, monitor.Secure, monitor.Ping)
			_, err = db.Query(query)
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
			UserID string `json:"user_id"`
			URL    string `json:"url"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if body.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		} else {
			// Get API key from user ID
			query := fmt.Sprintf("SELECT api_key FROM users WHERE user_id = '%s';", body.UserID)
			rows, err := db.Query(query)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}
			rows.Next()
			var apiKey string
			err = rows.Scan(&apiKey)
			if err != nil || apiKey == "" {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Delete monitor from database
			err = deleteMonitor(apiKey, body.URL, c, db)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}
			// Delete recorded pings from database for this monitor
			err = deletePings(apiKey, body.URL, c, db)
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

func RegisterRouter(r *gin.RouterGroup, db *sql.DB) {
	r.GET("/generate-api-key", genAPIKeyHandler(db))
	r.GET("/user-id/:apiKey", getUserIDHandler(db))
	r.GET("/requests/:userID", getUserRequestsHandler(db))
	r.GET("/delete/:apiKey", deleteDataHandler(db))
	r.GET("/monitor/pings/:userID", getUserPingsHandler(db))
	r.POST("/monitor/add", insertUserMonitorHandler(db))
	r.POST("/monitor/delete", deleteUserMonitorHandler(db))
	r.GET("/data", getDataHandler(db))
}
