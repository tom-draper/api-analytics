package routes

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(c *gin.Context) {
	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch all API request data associated with this account
	query := "INSERT INTO users (api_key, user_id, created_at, last_accessed) VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) RETURNING api_key;"

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
		return
	}
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

func getUserID(c *gin.Context) {
	// Get user ID associated with API key
	var apiKey string = c.Param("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch user ID corresponding with API key
	query := fmt.Sprintf("SELECT user_id FROM users WHERE api_key = '%s';", apiKey)
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}
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

type PublicRequestRow struct {
	Hostname     sql.NullString `json:"hostname"`
	IPAddress    sql.NullString `json:"ip_address"`
	Path         string         `json:"path"`
	UserAgent    sql.NullString `json:"user_agent"`
	Method       int16          `json:"method"`
	Status       int16          `json:"status"`
	ResponseTime int16          `json:"response_time"`
	Location     sql.NullString `json:"location"`
	CreatedAt    time.Time      `json:"created_at"`
}

func getUserRequests(c *gin.Context) {
	var userID string = c.Param("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch API key corresponding with user_id
	query := fmt.Sprintf("SELECT api_key FROM users WHERE user_id = '%s';", userID)
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
	rows.Next()
	var apiKey string
	err = rows.Scan(&apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Fetch user ID corresponding with API key
	// Left table join was originally used but often exceeded postgresql working memory limit with large numbers of requests
	query = fmt.Sprintf("SELECT ip_address, path, user_agent, method, response_time, status, location, created_at FROM requests WHERE api_key = '%s' LIMIT 1000000;", apiKey)
	rows, err = db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	err = updateLastAccessedByUserID(db, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Read data into compact list of lists to return
	cols := []any{"ip_address", "path", "user_agent", "method", "response_time", "status", "location", "created_at"}
	requests := buildRequestDataCompact(rows, cols)

	gzipOutput, err := compressJSON(requests)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
		return
	}

	// Return API request data
	c.Writer.Header().Set("Accept-Encoding", "gzip")
	c.Writer.Header().Set("Content-Encoding", "gzip")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Data(http.StatusOK, "gzip", gzipOutput)
}

func compressJSON(data any) ([]byte, error) {
	// Convert data to []byte
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Compress using gzip
	var buffer bytes.Buffer
	gzw := gzip.NewWriter(&buffer)
	_, err = gzw.Write(body)
	if err != nil {
		return nil, err
	}

	err = gzw.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func updateLastAccessedByUserID(db *sql.DB, userID string) error {
	query := fmt.Sprintf("UPDATE users SET last_accessed = NOW() WHERE user_id = '%s';", userID)
	_, err := db.Query(query)
	return err
}

func updateLastAccessedByAPIKey(db *sql.DB, apiKey string) error {
	query := fmt.Sprintf("UPDATE users SET last_accessed = NOW() WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func buildRequestDataCompact(rows *sql.Rows, cols []any) [][]any {
	// First value in list holds column names
	requests := [][]any{cols}
	request := new(PublicRequestRow) // Reused to avoid repeated memory allocation
	for rows.Next() {
		err := rows.Scan(&request.IPAddress, &request.Path, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
		if request.Location.String == "  " {
			request.Location.String = ""
		}
		if err == nil {
			requests = append(requests, []any{request.IPAddress.String, request.Path, request.UserAgent.String, request.Method, request.ResponseTime, request.Status, request.Location.String, request.CreatedAt})
		}
	}
	return requests
}

type DataFetchQueries struct {
	compact   bool
	date      time.Time
	dateFrom  time.Time
	dateTo    time.Time
	hostname  string
	ipAddress string
	location  string
	status    int
}

func getData(c *gin.Context) {
	apiKey := c.GetHeader("X-AUTH-TOKEN")
	if apiKey == "" {
		// Check old (deprecated) identifier
		apiKey = c.GetHeader("API-Key")
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}
	}

	// Get any queries from url
	queries := getQueriesFromRequest(c)

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch all API request data associated with this account
	query := buildDataFetchQuery(apiKey, queries)
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	err = updateLastAccessedByAPIKey(db, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Read data into list of objects to return
	if queries.compact {
		cols := []interface{}{"hostname", "ip_address", "path", "user_agent", "method", "response_time", "status", "location", "created_at"}
		requests := buildRequestDataCompact(rows, cols)
		c.JSON(http.StatusOK, requests)
	} else {
		requests := buildRequestData(rows)
		c.JSON(http.StatusOK, requests)
	}
}

func buildDataFetchQuery(apiKey string, queries DataFetchQueries) string {
	var query strings.Builder
	query.WriteString(fmt.Sprintf("SELECT hostname, ip_address, path, user_agent, method, response_time, status, location, created_at FROM requests WHERE api_key = '%s'", apiKey))

	// Providing a single date takes priority over range with dateFrom and dateTo
	if database.SanitizeDate(queries.date) {
		query.WriteString(fmt.Sprintf(" and created_at >= '%s' and created_at < date '%s' + interval '1 days'", queries.date.Format("2006-01-02"), queries.date.Format("2006-01-02")))
	} else {
		if database.SanitizeDate(queries.dateFrom) {
			query.WriteString(fmt.Sprintf(" and created_at >= '%s'", queries.dateFrom.Format("2006-01-02")))
		}
		if database.SanitizeDate(queries.dateTo) {
			query.WriteString(fmt.Sprintf(" and created_at <= '%s'", queries.dateTo.Format("2006-01-02")))
		}
	}

	if database.SanitizeIPAddress(queries.ipAddress) {
		query.WriteString(fmt.Sprintf(" and ip_address = '%s'", queries.ipAddress))
	}

	if database.SanitizeLocation(queries.location) {
		query.WriteString(fmt.Sprintf(" and location = '%s'", queries.location))
	}

	if database.SanitizeStatus(queries.status) {
		query.WriteString(fmt.Sprintf(" and status = %d", queries.status))
	}

	query.WriteString("LIMIT 1000000;")
	return query.String()
}

func getQueriesFromRequest(c *gin.Context) DataFetchQueries {
	compactQuery := c.Query("compact")
	dateQuery := c.Query("date")
	dateFromQuery := c.Query("dateFrom")
	dateToQuery := c.Query("dateTo")
	hostname := c.Query("hostname")
	ipAddressQuery := c.Query("ip")
	locationQuery := c.Query("location")
	statusQuery := c.Query("status")

	date := parseQueryDate(dateQuery)
	dateFrom := parseQueryDate(dateFromQuery)
	dateTo := parseQueryDate(dateToQuery)
	status, err := strconv.Atoi(statusQuery)
	if err != nil {
		status = 0
	}

	queries := DataFetchQueries{
		compactQuery == "true",
		date,
		dateFrom,
		dateTo,
		hostname,
		ipAddressQuery,
		locationQuery,
		status,
	}
	return queries
}

func parseQueryDate(date string) time.Time {
	if date == "" {
		return time.Time{}
	}

	// Try parse date
	d, err := time.Parse("2006-01-02", date)
	if err == nil {
		return d
	}
	return time.Time{}
}

func parseQueryDateTime(date string) time.Time {
	if date == "" {
		return time.Time{}
	}

	// Try parse date time
	d, err := time.Parse("2006-01-02 15:04:05", date)
	if err == nil {
		return d
	}

	// Try parse date
	d, err = time.Parse("2006-01-02", date)
	if err == nil {
		return d
	}
	return time.Time{}
}

type PublicRequestData struct {
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	Location     string    `json:"location"`
	CreatedAt    time.Time `json:"created_at"`
}

func buildRequestData(rows *sql.Rows) []PublicRequestData {
	requests := make([]PublicRequestData, 0)
	for rows.Next() {
		request := new(PublicRequestRow)
		err := rows.Scan(&request.Hostname, &request.IPAddress, &request.Path, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
		if err == nil {
			r := PublicRequestData{
				Hostname:     request.Hostname.String,
				IPAddress:    request.IPAddress.String,
				Path:         request.Path,
				UserAgent:    request.UserAgent.String,
				Method:       request.Method,
				Status:       request.Status,
				ResponseTime: request.ResponseTime,
				Location:     request.Location.String,
				CreatedAt:    request.CreatedAt,
			}
			requests = append(requests, r)
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

func deleteData(c *gin.Context) {
	apiKey := c.Param("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
	}

	db := database.OpenDBConnection()
	defer db.Close()

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

type PublicMonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserMonitor(c *gin.Context) {
	var userID string = c.Param("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	db := database.OpenDBConnection()
	defer db.Close()

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

type Monitor struct {
	UserID string `json:"user_id"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

func addUserMonitor(c *gin.Context) {
	var monitor Monitor
	if err := c.BindJSON(&monitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if monitor.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	db := database.OpenDBConnection()
	defer db.Close()

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

	// Get monitor count
	query = fmt.Sprintf("SELECT count(*) FROM monitor WHERE api_key = '%s';", apiKey)
	rows, err = db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	if count > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Monitor limit reached."})
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

func deleteUserMonitor(c *gin.Context) {
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
	}

	db := database.OpenDBConnection()
	defer db.Close()

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

type PublicPingsRow struct {
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func getUserPings(c *gin.Context) {
	var userID string = c.Param("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	db := database.OpenDBConnection()
	defer db.Close()

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

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/generate-api-key", genAPIKey)
	r.GET("/user-id/:apiKey", getUserID)
	r.GET("/requests/:userID", getUserRequests)
	r.GET("/delete/:apiKey", deleteData)
	r.GET("/monitor/pings/:userID", getUserPings)
	r.POST("/monitor/add", addUserMonitor)
	r.POST("/monitor/delete", deleteUserMonitor)
	r.GET("/data", getData)
}
