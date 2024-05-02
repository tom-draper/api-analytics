package routes

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/api/lib/log"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(c *gin.Context) {
	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch all API request data associated with this account
	query := "INSERT INTO users (api_key, user_id, created_at, last_accessed) VALUES (gen_random_uuid(), gen_random_uuid(), NOW(), NOW()) RETURNING api_key;"

	var apiKey string
	err := connection.QueryRow(context.Background(), query).Scan(&apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("API key generation failed - %s", err.Error()))
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch user ID corresponding with API key
	var userID string
	query := "SELECT user_id FROM users WHERE api_key = $1;"
	err := connection.QueryRow(context.Background(), query, apiKey).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Return user ID
	c.JSON(http.StatusOK, userID)
}

type DashboardData struct {
	UserAgents UserAgentsLookup `json:"user_agents"`
	Requests   [][10]any        `json:"requests"`
}

type UserAgentsLookup map[int]string

type DashboardRequestRow struct {
	Hostname     *string     `json:"hostname"` // Nullable
	IPAddress    pgtype.CIDR `json:"ip_address"`
	Path         string      `json:"path"`
	UserAgent    *int        `json:"user_agent"` // Nullable
	Method       int16       `json:"method"`
	Status       int16       `json:"status"`
	ResponseTime int16       `json:"response_time"`
	Location     *string     `json:"location"` // Nullable
	UserID       *string     `json:"user_id"` // Nullable, custom user identifier field specific to each API service
	CreatedAt    time.Time   `json:"created_at"`
}

func getRequests(c *gin.Context) {
	var userID string = c.Param("userID")
	if userID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Dashboard access", userID))

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch API key corresponding with user ID
	apiKey, err := getUserAPIKey(connection, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	requests := [][10]any{}
	pageSize := 1_000_000
	maxRequests := pageSize   // Temporary limit to prevent memory issues
	pageMarker := time.Time{} // Start with min time to capture first page
	userAgentIDs := make(map[int]struct{})

	// Read paginated requests data
	for {
		// Left table join was originally used but often exceeded postgresql working memory limit with large numbers of requests
		query = "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at FROM requests WHERE api_key = $1 AND created_at >= $2 ORDER BY created_at LIMIT $3;"
		rows, err := connection.Query(context.Background(), query, apiKey, pageMarker, pageSize)
		if err != nil {
			log.LogToFile(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// First value in list holds column names
		request := new(DashboardRequestRow)
		var count int
		for rows.Next() {
			err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.UserID, &request.CreatedAt)
			if err == nil {
				var ip string
				if request.IPAddress.IPNet != nil {
					ip = request.IPAddress.IPNet.IP.String()
				}
				hostname := getNullableString(request.Hostname)
				location := getNullableString(request.Location)
				userID := getNullableString(request.UserID)
				requests = append(requests, [10]any{ip, request.Path, hostname, request.UserAgent, request.Method, request.ResponseTime, request.Status, location, userID, request.CreatedAt})
				if request.UserAgent != nil {
					if _, ok := userAgentIDs[*request.UserAgent]; !ok {
						userAgentIDs[*request.UserAgent] = struct{}{}
					}
				}
			}
			count++
			if count >= maxRequests {
				break
			}
		}
		// If haven't reached page size, there are no more rows to read
		if count <= pageSize || count >= maxRequests {
			rows.Close()
			break
		}
		// Save the final row's timestamp to know where next page begins
		lastIdx := len(requests) - 1
		lastTimestamp := requests[lastIdx][len(requests[lastIdx])-1].(time.Time)
		pageMarker = lastTimestamp
		rows.Close()
	}

	// Convert user agent IDs to names
	userAgents, err := getUserAgents(connection, userAgentIDs)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User agent lookup failed."})
		return
	}

	body := DashboardData{
		UserAgents: userAgents,
		Requests:   requests,
	}

	gzipOutput, err := compressJSON(body)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
		return
	}

	// Return API request data
	c.Writer.Header().Set("Accept-Encoding", "gzip")
	c.Writer.Header().Set("Content-Encoding", "gzip")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Data(http.StatusOK, "gzip", gzipOutput)

	log.LogToFile(fmt.Sprintf("key=%s: Dashboard access successful (%d)", apiKey, len(requests)))

	// Record user dashboard access
	err = updateLastAccessed(connection, apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
}

func getPaginatedRequests(c *gin.Context) {
	var userID string = c.Param("userID")
	if userID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	page, err = strconv.Atoi(c.Param("page"))
	if err != nil || page == 0 {
		log.LogToFile("Invalid page number")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Dashboard page %d access", userID, page))

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch API key corresponding with user ID
	apiKey, err := getUserAPIKey(connection, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	const pageSize int = 500_000
	requests := [][10]any{}
	userAgentIDs := make(map[int]struct{})

	query = "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
	rows, err := connection.Query(context.Background(), query, apiKey, pageSize, page*pageSize)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
	request := new(DashboardRequestRow)  // Reuseable request struct
	for rows.Next() {
		err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.UserID, &request.CreatedAt)
		if err != nil {
			continue
		}

		var ip string
		if request.IPAddress.IPNet != nil {
			ip = request.IPAddress.IPNet.IP.String()
		}
		hostname := getNullableString(request.Hostname)
		location := getNullableString(request.Location)
		userID := getNullableString(request.UserID)
		requests = append(requests, [10]any{ip, request.Path, hostname, request.UserAgent, request.Method, request.ResponseTime, request.Status, location, userID, request.CreatedAt})
		if request.UserAgent != nil {
			if _, ok := userAgentIDs[*request.UserAgent]; !ok {
				userAgentIDs[*request.UserAgent] = struct{}{}
			}
		}
	}
	rows.Close()

	// Convert user agent IDs to names
	userAgents, err := getUserAgents(connection, userAgentIDs)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User agent lookup failed."})
		return
	}

	// Store user agents in separate lookup table to reduce data transfer size
	body := DashboardData{
		UserAgents: userAgents,
		Requests:   requests,
	}

	gzipOutput, err := compressJSON(body)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
		return
	}

	// Return API request data
	c.Writer.Header().Set("Accept-Encoding", "gzip")
	c.Writer.Header().Set("Content-Encoding", "gzip")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Data(http.StatusOK, "gzip", gzipOutput)

	log.LogToFile(fmt.Sprintf("key=%s: Dashboard page %d access successful (%d)", apiKey, page, len(requests)))

	// Record user dashboard access
	err = updateLastAccessed(connection, apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
}

func getUserAPIKey(connection *pgx.Conn, userID string) (string, error) {
	// Avoiding table join due to memory limitations
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	err := connection.QueryRow(context.Background(), query, userID).Scan(&apiKey)
	return apiKey, err
}

func getUserAgents(connection *pgx.Conn, userAgentIDs map[int]struct{}) (map[int]string, error) {
	// Convert user agent int IDs to equivalent strings
	userAgents := make(map[int]string)
	if len(userAgentIDs) == 0 {
		return userAgents, nil
	}

	var userAgentsQuery strings.Builder
	userAgentsQuery.WriteString("SELECT id, user_agent FROM user_agents WHERE id IN (")
	arguments := []any{}
	var i int
	for id := range userAgentIDs {
		userAgentsQuery.WriteString("$" + strconv.Itoa(i+1))
		arguments = append(arguments, id)
		if i < len(userAgentIDs)-1 {
			userAgentsQuery.WriteString(",")
		}
		i++
	}
	userAgentsQuery.WriteString(");")
	rows, err := connection.Query(context.Background(), userAgentsQuery.String(), arguments...)
	if err != nil {
		return userAgent, err
	}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err == nil {
			userAgents[id] = name
		}
	}
	rows.Close()

	return userAgents, nil
}

func getNullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

func updateLastAccessedByUserID(connection *pgx.Conn, userID string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE user_id = $1;"
	_, err := connection.Exec(context.Background(), query, userID)
	return err
}

func updateLastAccessed(connection *pgx.Conn, apiKey string) error {
	query := "UPDATE users SET last_accessed = NOW() WHERE api_key = $1;"
	_, err := connection.Exec(context.Background(), query, apiKey)
	return err
}

func buildRequestDataCompact(rows pgx.Rows, cols [10]any) [][10]any {
	// First value in list holds column names
	requests := [][10]any{cols}
	var request DashboardRequestRow
	for rows.Next() {
		err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.UserID, &request.CreatedAt)
		if err == nil {
			requests = append(requests, [10]any{request.IPAddress, request.Path, request.Hostname, request.UserAgent, request.Method, request.ResponseTime, request.Status, request.Location, request.UserID, request.CreatedAt})
		}
	}
	return requests
}

type DataFetchQueries struct {
	page      int
	compact   bool
	date      time.Time
	dateFrom  time.Time
	dateTo    time.Time
	hostname  string
	ipAddress string
	location  string
	status    int
	userID    string
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch all API request data associated with this account
	query, arguments := buildDataFetchQuery(apiKey, queries)
	rows, err := connection.Query(context.Background(), query, arguments...)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Queries failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Read data into list of objects to return
	if queries.compact {
		cols := [10]any{"ip_address", "path", "hostname", "user_agent", "method", "response_time", "status", "location", "user_id", "created_at"}
		requests := buildRequestDataCompact(rows, cols)
		log.LogToFile(fmt.Sprintf("key=%s: Data access successful (%d)", apiKey, len(requests)-1))
		c.JSON(http.StatusOK, requests)
	} else {
		requests := buildRequestData(rows)
		log.LogToFile(fmt.Sprintf("key=%s: Data access successful (%d)", apiKey, len(requests)))
		c.JSON(http.StatusOK, requests)
	}

	rows.Close()

	err = updateLastAccessed(connection, apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}
}

func buildDataFetchQuery(apiKey string, queries DataFetchQueries) (string, []any) {
	var query strings.Builder
	query.WriteString("SELECT r.ip_address, r.path, r.hostname, u.user_agent, r.method, r.response_time, r.status, r.location, r.user_id, r.created_at FROM requests r JOIN user_agents u ON r.user_agent_id = u.id WHERE api_key = $1")

	arguments := []any{apiKey}

	// Providing a single date takes priority over range with dateFrom and dateTo
	if !queries.date.IsZero() && database.ValidDate(queries.date) {
		query.WriteString(fmt.Sprintf(" and r.created_at >= $%d and r.created_at < date $%d + interval '1 days'", len(arguments)+1, len(arguments)+2))
		arguments = append(arguments, queries.date.Format("2006-01-02"), queries.date.Format("2006-01-02"))
	} else {
		if !queries.dateFrom.IsZero() && database.ValidDate(queries.dateFrom) {
			query.WriteString(fmt.Sprintf(" and r.created_at >= $%d", len(arguments)+1))
			arguments = append(arguments, queries.dateFrom.Format("2006-01-02"))
		}
		if !queries.dateTo.IsZero() && database.ValidDate(queries.dateTo) {
			query.WriteString(fmt.Sprintf(" and r.created_at <= $%d", len(arguments)+1))
			arguments = append(arguments, queries.dateTo.Format("2006-01-02"))
		}
	}

	if queries.ipAddress != "" && database.ValidIPAddress(queries.ipAddress) {
		query.WriteString(fmt.Sprintf(" and r.ip_address = $%d", len(arguments)+1))
		arguments = append(arguments, queries.ipAddress)
	}

	if queries.location != "" && database.ValidLocation(queries.location) {
		query.WriteString(fmt.Sprintf(" and r.location = $%d", len(arguments)+1))
		arguments = append(arguments, queries.location)
	}

	if queries.status != 0 && database.ValidStatus(queries.status) {
		query.WriteString(fmt.Sprintf(" and r.status = $%d", len(arguments)+1))
		arguments = append(arguments, queries.status)
	}

	if queries.hostname != "" && database.ValidString(queries.hostname) {
		query.WriteString(fmt.Sprintf(" and r.hostname = $%d", len(arguments)+1))
		arguments = append(arguments, queries.hostname)
	}

	if queries.userID != "" && database.ValidString(queries.userID) {
		query.WriteString(fmt.Sprintf(" and r.user_id = $%d", len(arguments)+1))
		arguments = append(arguments, queries.userID)
	}

	const pageSize = 50_000
	offset := (queries.page - 1) * pageSize
	query.WriteString(fmt.Sprintf(" ORDER BY created_at LIMIT %d OFFSET %d;", pageSize, offset))
	return query.String(), arguments
}

func getQueriesFromRequest(c *gin.Context) DataFetchQueries {
	pageQuery := c.Query("page")
	compactQuery := c.Query("compact")
	dateQuery := c.Query("date")
	dateFromQuery := c.Query("dateFrom")
	dateToQuery := c.Query("dateTo")
	hostname := c.Query("hostname")
	ipAddressQuery := c.Query("ip")
	locationQuery := c.Query("location")
	statusQuery := c.Query("status")
	userIDQuery := c.Query("userID")

	date := parseQueryDate(dateQuery)
	dateFrom := parseQueryDate(dateFromQuery)
	dateTo := parseQueryDate(dateToQuery)
	status, err := strconv.Atoi(statusQuery)
	if err != nil {
		status = 0
	}
	page := 1
	if pageQuery != "" {
		p, err := strconv.Atoi(pageQuery)
		if err == nil {
			page = p
		}
	}

	queries := DataFetchQueries{
		page,
		compactQuery == "true",
		date,
		dateFrom,
		dateTo,
		hostname,
		ipAddressQuery,
		locationQuery,
		status,
		userIDQuery,
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

type RequestData struct {
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	Location     string    `json:"location"`
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type RequestRow struct {
	Hostname     *string     `json:"hostname"`
	IPAddress    pgtype.CIDR `json:"ip_address"`
	Path         string      `json:"path"`
	UserAgent    *string     `json:"user_agent"`
	Method       int16       `json:"method"`
	Status       int16       `json:"status"`
	ResponseTime int16       `json:"response_time"`
	Location     *string     `json:"location"`
	UserID       *string     `json:"user_id"` // Custom user identifier field specific to each API service
	CreatedAt    time.Time   `json:"created_at"`
}

func buildRequestData(rows pgx.Rows) []RequestData {
	requests := make([]RequestData, 0)
	var request RequestRow
	for rows.Next() {
		err := rows.Scan(&request.IPAddress, &request.Path, &request.Hostname, &request.UserAgent, &request.Method, &request.ResponseTime, &request.Status, &request.Location, &request.UserID, &request.CreatedAt)
		if err == nil {
			var ip string
			if request.IPAddress.IPNet != nil {
				ip = request.IPAddress.IPNet.IP.String()
			}
			hostname := getNullableString(request.Hostname)
			userAgent := getNullableString(request.UserAgent)
			location := getNullableString(request.Location)
			userID := getNullableString(request.UserID)
			requests = append(requests, RequestData{
				IPAddress:    ip,
				Path:         request.Path,
				Hostname:     hostname,
				UserAgent:    userAgent,
				Method:       request.Method,
				Status:       request.Status,
				ResponseTime: request.ResponseTime,
				Location:     location,
				UserID:       userID,
				CreatedAt:    request.CreatedAt,
			})
		}
	}
	return requests
}

func deleteUserRequests(apiKey string, c *gin.Context, connection *pgx.Conn) error {
	// Delete all user's API request data
	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err := connection.Exec(context.Background(), query, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserAccount(apiKey string, c *gin.Context, connection *pgx.Conn) error {
	// Delete user account record
	query := "DELETE FROM users WHERE api_key = $1;"
	_, err := connection.Exec(context.Background(), query, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserMonitors(apiKey string, c *gin.Context, connection *pgx.Conn) error {
	// Delete all user's monitored urls
	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err := connection.Exec(context.Background(), query, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return err
	}
	return nil
}

func deleteUserPings(apiKey string, c *gin.Context, connection *pgx.Conn) error {
	// Delete all user's recorded pings to all monitored urls
	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err := connection.Exec(context.Background(), query, apiKey)
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	if err := deleteUserRequests(apiKey, c, connection); err != nil {
		return
	} else if err := deleteUserAccount(apiKey, c, connection); err != nil {
		return
	} else if err := deleteUserMonitors(apiKey, c, connection); err != nil {
		return
	} else if err := deleteUserPings(apiKey, c, connection); err != nil {
		return
	}

	// Return API request data
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
}

type MonitorRow struct {
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Retreive monitors created by this user
	query := "SELECT url, secure, ping, monitor.created_at FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
	rows, err := connection.Query(context.Background(), query, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
	defer rows.Close()

	// Read monitors into list to return
	monitors := make([]MonitorRow, 0)
	for rows.Next() {
		var monitor MonitorRow
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
	err := c.BindJSON(&monitor)
	if err != nil {
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Get API key from user ID
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	err = connection.QueryRow(context.Background(), query, monitor.UserID).Scan(&apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", monitor.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Check if monitor already exists
	var count int
	query = "SELECT count(*) FROM monitor WHERE api_key = $1 AND url = $2;"
	err = connection.QueryRow(context.Background(), query, apiKey, monitor.URL).Scan(&count)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to get monitor count - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	if count == 1 {
		log.LogToFile(fmt.Sprintf("key=%s: Monitor already exists", apiKey))
		c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "message": "Monitor already exists."})
		return
	}

	// Get monitor count
	var monitorCount int
	query = "SELECT count(*) FROM monitor WHERE api_key = $1;"
	err = connection.QueryRow(context.Background(), query, apiKey).Scan(&monitorCount)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to get monitor count - %s", apiKey, err.Error()))
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
	_, err = connection.Exec(context.Background(), query, apiKey, monitor.URL, monitor.Secure, monitor.Ping)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to create new monitor - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: Monitor '%s' created successfully", apiKey, monitor.URL))

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor created successfully."})
}

func deleteMonitor(apiKey string, url string, c *gin.Context, connection *pgx.Conn) error {
	// Delete user's monitor to this specific url
	query := "DELETE FROM monitor WHERE api_key = $1 AND url = $2;"
	_, err := connection.Exec(context.Background(), query, apiKey, url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return err
	}
	return nil
}

func deletePings(apiKey string, url string, c *gin.Context, connection *pgx.Conn) error {
	// Delete user's recorded pings to monitored url
	query := "DELETE FROM pings WHERE api_key = $1 AND url = $2;"
	_, err := connection.Exec(context.Background(), query, apiKey, url)
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
	err := c.BindJSON(&body)
	if err != nil {
		log.LogToFile(fmt.Sprintf("Invalid monitor to delete - %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if body.UserID == "" {
		log.LogToFile("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	log.LogToFile(fmt.Sprintf("id=%s: Delete monitor", body.UserID))

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Get API key from user ID
	var apiKey string
	query := "SELECT api_key FROM users WHERE user_id = $1;"
	err = connection.QueryRow(context.Background(), query, body.UserID).Scan(&apiKey)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", body.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Delete monitor from database
	err = deleteMonitor(apiKey, body.URL, c, connection)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to delete monitor - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	// Delete recorded pings from database for this monitor
	err = deletePings(apiKey, body.URL, c, connection)
	if err != nil {
		log.LogToFile(fmt.Sprintf("key=%s: Failed to delete pings - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	log.LogToFile(fmt.Sprintf("key=%s: Monitor '%s' deleted successfully", apiKey, body.URL))

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Monitor deleted successfully."})
}

type PingsRow struct {
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

	connection := database.NewConnection()
	defer connection.Close(context.Background())

	// Fetch user ID corresponding with API key
	query := "SELECT url FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
	rows, err := connection.Query(context.Background(), query, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Monitor access failed - %s", userID, err.Error()))
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
	rows, err = connection.Query(context.Background(), query, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: Ping access failed - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	// Read pings into list to return
	for rows.Next() {
		var url string
		var ping MonitorPing
		err := rows.Scan(&url, &ping.ResponseTime, &ping.Status, &ping.CreatedAt)
		if err == nil {
			if val, ok := monitors[url]; ok {
				monitors[url] = append(val, ping)
			}
		}
	}

	// Record user pings access
	err = updateLastAccessedByUserID(connection, userID)
	if err != nil {
		log.LogToFile(fmt.Sprintf("id=%s: User last access update failed - %s", userID, err.Error()))
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
	r.GET("/requests/:userID", getRequests)
	r.GET("/requests/:userID/:page", getPaginatedRequests)
	r.GET("/delete/:apiKey", deleteData)
	r.GET("/monitor/pings/:userID", getUserPings)
	r.POST("/monitor/add", addUserMonitor)
	r.POST("/monitor/delete", deleteUserMonitor)
	r.GET("/data", getData)
}
