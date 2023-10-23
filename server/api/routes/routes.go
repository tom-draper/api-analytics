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
	"github.com/tom-draper/api-analytics/server/api/lib/log"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(c *gin.Context) {
	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch all API request data associated with this account
	query := "INSERT INTO users (api_key, user_id, created_at, last_accessed) VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) RETURNING api_key;"

	rows, err := db.Query(query)
	if err != nil {
		log.LogToFile(fmt.Sprintf("API key generation failed - %w", err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
		return
	}

	// Get API key auto generated from new row insertion
	rows.Next()
	var apiKey string
	if err = rows.Scan(&apiKey); err != nil {
		log.LogToFile(fmt.Sprintf("Failed to access generated key - %w", err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: API key generation successful", apiKey))

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
	query := "SELECT user_id FROM users WHERE api_key = $1;"
	rows, err := db.Query(query, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// API key is primary key so assumed only one row returned
	rows.Next()
	var userID string
	if err = rows.Scan(&userID); err != nil {
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
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Dashboard access", userID))

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch API key corresponding with user_id
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	rows, err := db.Query(query, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %w", userID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	rows.Next()
	var apiKey string
	if err = rows.Scan(&apiKey); err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %w", userID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	cols := []any{"ip_address", "path", "hostname", "user_agent", "method", "response_time", "status", "location", "created_at"}
	requests := [][]any{cols}
	pageSize := 500000
	minTimestamp := time.Time{}

	// Read paginated requests data
	for {
		// Fetch user ID corresponding with API key
		// Left table join was originally used but often exceeded postgresql working memory limit with large numbers of requests
		query = fmt.Sprintf("SELECT ip_address, path, hostname, user_agent, method, response_time, status, location, created_at FROM requests WHERE api_key = $1 AND created_at >= $2 ORDER BY created_at LIMIT %d;", pageSize)
		rows, err = db.Query(query, apiKey, minTimestamp)
		if err != nil {
			log.LogToFile(fmt.Sprintf("key=%s: Invalid API key - %w", apiKey, err))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// First value in list holds column names
		var request PublicRequestRow
		var count int
		for rows.Next() {
			err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt)
			if err == nil {
				if request.Location.String == "  " {
					request.Location.String = ""
				}
				requests = append(requests, []any{request.IPAddress.String, request.Path, request.Hostname.String, request.UserAgent.String, request.Method, request.ResponseTime, request.Status, request.Location.String, request.CreatedAt})
			}
			count++
		}
		// If haven't reached page size, there are no more rows to read
		if count < pageSize {
			break
		}
		// Save the final row's timestamp to know where next page begins
		lastIdx := len(requests) - 1
		lastTimestamp := requests[lastIdx][len(requests[lastIdx])-1].(time.Time)
		minTimestamp = lastTimestamp
	}

	gzipOutput, err := compressJSON(requests)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Compression failed - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
		return
	}

	// Record access
	if err = updateLastAccessed(db, apiKey); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User last access update failed - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: Dashboard access successful (%d)", apiKey, len(requests)-1))

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
	if _, err = gzw.Write(body); err != nil {
		return nil, err
	}

	if err = gzw.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func updateLastAccessedByUserID(db *sql.DB, userID string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE user_id = $1;"
	_, err := db.Query(query, userID)
	return err
}

func updateLastAccessed(db *sql.DB, apiKey string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE api_key = $1;"
	_, err := db.Query(query, apiKey)
	return err
}

func buildRequestDataCompact(rows *sql.Rows, cols []any) [][]any {
	// First value in list holds column names
	requests := [][]any{cols}
	// request := new(PublicRequestRow) // Reused to avoid repeated memory allocation
	var request PublicRequestRow
	for rows.Next() {
		if err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt); err == nil {
			if request.Location.String == "  " {
				request.Location.String = ""
			}
			requests = append(requests, []any{request.IPAddress.String, request.Path, request.Hostname.String, request.UserAgent.String, request.Method, request.ResponseTime, request.Status, request.Location.String, request.CreatedAt})
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
			log.LogToFile("API key empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}
	}

	log.LogToFile(fmt.Sprintf("key=%s: Data access", apiKey))

	// Get any queries from url
	queries := getQueriesFromRequest(c)

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch all API request data associated with this account
	query, arguments := buildDataFetchQuery(apiKey, queries)
	rows, err := db.Query(query, arguments...)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Queries failed - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	if err := updateLastAccessed(db, apiKey); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User last access update failed - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Read data into list of objects to return
	if queries.compact {
		cols := []interface{}{"ip_address", "path", "hostname", "user_agent", "method", "response_time", "status", "location", "created_at"}
		requests := buildRequestDataCompact(rows, cols)
		log.LogToFile(fmt.Sprintf("key=%s: Data access successful (%d)", apiKey, len(requests)-1))
		c.JSON(http.StatusOK, requests)
	} else {
		requests := buildRequestData(rows)
		log.LogToFile(fmt.Sprintf("key=%s: Data access successful (%d)", apiKey, len(requests)-1))
		c.JSON(http.StatusOK, requests)
	}
}

func buildDataFetchQuery(apiKey string, queries DataFetchQueries) (string, []any) {
	var query strings.Builder
	query.WriteString("SELECT ip_address, path, hostname, user_agent, method, response_time, status, location, created_at FROM requests WHERE api_key = $1")

	arguments := []any{apiKey}

	// Providing a single date takes priority over range with dateFrom and dateTo
	if database.SanitizeDate(queries.date) {
		query.WriteString(fmt.Sprintf(" and created_at >= $%d and created_at < date $%d + interval '1 days'", len(arguments)+1, len(arguments)+2))
		arguments = append(arguments, queries.date.Format("2006-01-02"), queries.date.Format("2006-01-02"))
	} else {
		if database.SanitizeDate(queries.dateFrom) {
			query.WriteString(fmt.Sprintf(" and created_at >= $%d", len(arguments)+1))
			arguments = append(arguments, queries.dateFrom.Format("2006-01-02"))
		}
		if database.SanitizeDate(queries.dateTo) {
			query.WriteString(fmt.Sprintf(" and created_at <= $%d", len(arguments)+1))
			arguments = append(arguments, queries.dateTo.Format("2006-01-02"))
		}
	}

	if database.SanitizeIPAddress(queries.ipAddress) {
		query.WriteString(fmt.Sprintf(" and ip_address = $%d", len(arguments)+1))
		arguments = append(arguments, queries.ipAddress)
	}
	if database.SanitizeLocation(queries.location) {
		query.WriteString(fmt.Sprintf(" and location = $%d", len(arguments)+1))
		arguments = append(arguments, queries.location)
	}
	if database.SanitizeStatus(queries.status) {
		query.WriteString(fmt.Sprintf(" and status = $%d", len(arguments)+1))
		arguments = append(arguments, queries.status)
	}
	if database.SanitizeString(queries.hostname) {
		query.WriteString(fmt.Sprintf(" and hostname = $%d", len(arguments)+1))
		arguments = append(arguments, queries.hostname)
	}

	query.WriteString(" LIMIT 700000;")
	return query.String(), arguments
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
	if d, err := time.Parse("2006-01-02", date); err == nil {
		return d
	}
	return time.Time{}
}

func parseQueryDateTime(date string) time.Time {
	if date == "" {
		return time.Time{}
	}

	// Try parse date time
	if d, err := time.Parse("2006-01-02 15:04:05", date); err == nil {
		return d
	}

	// Try parse date
	if d, err := time.Parse("2006-01-02", date); err == nil {
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
		var request PublicRequestRow
		if err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.CreatedAt); err == nil {
			r := PublicRequestData{
				IPAddress:    request.IPAddress.String,
				Path:         request.Path,
				Hostname:     request.Hostname.String,
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
	query := "DELETE FROM requests WHERE api_key = $1;"
	if _, err := db.Query(query, apiKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserAccount(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete user account record
	query := "DELETE FROM users WHERE api_key = $1;"
	if _, err := db.Query(query, apiKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserMonitors(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete all user's monitored urls
	query := "DELETE FROM monitor WHERE api_key = $1;"
	if _, err := db.Query(query, apiKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserPings(apiKey string, c *gin.Context, db *sql.DB) error {
	// Delete all user's recorded pings to all monitored urls
	query := "DELETE FROM pings WHERE api_key = $1;"
	if _, err := db.Query(query, apiKey); err != nil {
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

	if err := deleteUserRequests(apiKey, c, db); err != nil {
		return
	} else if err := deleteUserAccount(apiKey, c, db); err != nil {
		return
	} else if err := deleteUserMonitors(apiKey, c, db); err != nil {
		return
	} else if err := deleteUserPings(apiKey, c, db); err != nil {
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
	query := "SELECT url, secure, ping, monitor.created_at FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Read monitors into list to return
	monitors := make([]PublicMonitorRow, 0)
	for rows.Next() {
		var monitor PublicMonitorRow
		err := rows.Scan(&monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, monitor)
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
		log.LogToFile("Invalid monitor to add")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if monitor.UserID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Add monitor", monitor.UserID))

	db := database.OpenDBConnection()
	defer db.Close()

	// Get API key from user ID
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	rows, err := db.Query(query, monitor.UserID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Invalid monitor user ID - %w", monitor.UserID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	rows.Next()
	var apiKey string
	err = rows.Scan(&apiKey)
	if err != nil || apiKey == "" {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %w", monitor.UserID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Check if monitor already exists
	query = "SELECT count(*) FROM monitor WHERE api_key = $1 AND url = $2;"
	rows, err = db.Query(query, apiKey, monitor.URL)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to get monitor count - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to read monitor count - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	if count == 1 {
		log.LogToFile(fmt.Sprintf("key=%s: Monitor already exists", apiKey))
		c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "message": "Monitor already exists."})
		return
	}

	// Get monitor count
	query = fmt.Sprintf("SELECT count(*) FROM monitor WHERE api_key = '%s';", apiKey)
	rows, err = db.Query(query)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to get monitor count - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	rows.Next()
	var monitorCount int
	if err := rows.Scan(&monitorCount); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to read monitor count - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	// Check if existing monitors already at max limit
	if monitorCount >= 3 {
		log.LogToFile(fmt.Sprintf("key=%s: Monitor limit reached (%d)", apiKey, monitorCount))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Monitor limit reached."})
		return
	}

	// Insert new monitor into database
	query = "INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ($1, $2, $3, $4, NOW())"
	if _, err := db.Query(query, apiKey, monitor.URL, monitor.Secure, monitor.Ping); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to create new monitor - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: Monitor '%s' created successfully", apiKey, monitor.URL))

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor created successfully."})
}

func deleteMonitor(apiKey string, url string, c *gin.Context, db *sql.DB) error {
	// Delete user's monitor to this specific url
	query := "DELETE FROM monitor WHERE api_key = $1 AND url = $2;"
	_, err := db.Query(query, apiKey, url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return err
	}
	return nil
}

func deletePings(apiKey string, url string, c *gin.Context, db *sql.DB) error {
	// Delete user's recorded pings to monitored url
	query := "DELETE FROM pings WHERE api_key = $1 AND url = $2;"
	_, err := db.Query(query, apiKey, url)
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
		log.LogToFile(fmt.Sprintf("Invalid monitor to delete - %w", err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if body.UserID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Delete monitor", body.UserID))

	db := database.OpenDBConnection()
	defer db.Close()

	// Get API key from user ID
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	rows, err := db.Query(query, body.UserID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Invalid monitor user ID - %w", body.UserID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	rows.Next()
	var apiKey string
	if err := rows.Scan(&apiKey); err != nil || apiKey == "" {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %w", body.UserID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Delete monitor from database
	if err := deleteMonitor(apiKey, body.URL, c, db); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to delete monitor - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	// Delete recorded pings from database for this monitor
	if err := deletePings(apiKey, body.URL, c, db); err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to delete pings - %w", apiKey, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: Monitor '%s' deleted successfully", apiKey, body.URL))

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Monitor deleted successfully."})
}

type PublicPingsRow struct {
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type MonitorPing struct {
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func getUserPings(c *gin.Context) {
	var userID string = c.Param("userID")
	if userID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Monitor access", userID))

	db := database.OpenDBConnection()
	defer db.Close()

	// Fetch user ID corresponding with API key
	query := "SELECT url FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
	rows, err := db.Query(query, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Monitor access failed - %w", userID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Initialise monitored URLs
	monitors := make(map[string][]MonitorPing)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err == nil {
			monitors[url] = make([]MonitorPing, 0)
		}
	}

	// Fetch user ID corresponding with API key
	query = "SELECT url, response_time, status, pings.created_at FROM pings INNER JOIN users ON users.api_key = pings.api_key WHERE users.user_id = $1;"
	rows, err = db.Query(query, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Ping access failed - %w", userID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Read pings into list to return
	for rows.Next() {
		var url string
		var ping MonitorPing
		if err := rows.Scan(&url, &ping.ResponseTime, &ping.Status, &ping.CreatedAt); err == nil {
			if val, ok := monitors[url]; ok {
				monitors[url] = append(val, ping)
			}
		}
	}

	// Record access
	if err := updateLastAccessedByUserID(db, userID); err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: User last access update failed - %w", userID, err))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Monitor access successful (%d)", userID, len(monitors)))

	// Return API request data
	c.JSON(http.StatusOK, monitors)
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
