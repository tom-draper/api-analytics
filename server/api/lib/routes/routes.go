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
	"github.com/tom-draper/api-analytics/server/api/lib/env"
	"github.com/tom-draper/api-analytics/server/api/lib/logging"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(c *gin.Context) {
	apiKey, err := database.CreateUser()
	if err != nil {
		logging.Info(fmt.Sprintf("API key generation failed - %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
		return
	}

	logging.Info(fmt.Sprintf("key=%s: API key generation successful", apiKey))

	c.JSON(http.StatusOK, apiKey)
}

func getUserID(c *gin.Context) {
	var apiKey string = c.Param("apiKey")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Get user ID associated with API key
	userID, err := database.GetUserID(apiKey)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: User ID fetch failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	c.JSON(http.StatusOK, userID)
}

type DashboardData struct {
	UserAgents UserAgentsLookup `json:"user_agents"`
	Requests   [][12]any        `json:"requests"`
}

type UserAgentsLookup map[int]string

type DashboardRequestRow struct {
	Hostname     *string     `json:"hostname"` // Nullable
	IPAddress    pgtype.CIDR `json:"ip_address"`
	Path         string      `json:"path"`
	UserAgent    *int        `json:"user_agent"` // Nullable
	Referrer     *string     `json:"referrer"`   // Nullable
	Method       int16       `json:"method"`
	Status       int16       `json:"status"`
	ResponseTime int16       `json:"response_time"`
	Location     *string     `json:"location"` // Nullable
	// UserHash     *string     `json:"user_hash"` // Nullable
	UserID    *string   `json:"user_id"` // Nullable, custom user identifier field specific to each API service
	CreatedAt time.Time `json:"created_at"`
}

func getMaxLoad() int {
	return env.GetIntegerEnvVariable("MAX_LOAD", 1_000_000)
}

func getPageSize() int {
	return env.GetIntegerEnvVariable("PAGE_SIZE", 250_000)
}

func getRequestsHandler() gin.HandlerFunc {
	var pageSize int = getPageSize()
	var maxLoad int = getMaxLoad()

	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			logging.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid targetPage ID."})
			return
		}

		var err error
		pageQuery := c.Query("page")
		logging.Info(pageQuery)
		targetPage := 1
		if pageQuery != "" {
			targetPage, err = strconv.Atoi(pageQuery)
			if err != nil {
				logging.Info(fmt.Sprintf("Failed to parse page number '%s' from query", pageQuery))
			}
		}

		var message string
		if targetPage == 0 {
			message = fmt.Sprintf("id=%s: Dashboard access", userID)
		} else {
			message = fmt.Sprintf("id=%s: Dashboard page %d access", userID, targetPage)
		}
		logging.Info(message)

		connection, err := database.NewConnection()
		if err != nil {
			logging.Info(fmt.Sprintf("id=%s: Dashboard access failed - %s", userID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}
		defer connection.Close(context.Background())

		// Fetch API key corresponding with user ID
		apiKey, err := database.GetAPIKeyWithConnection(context.Background(), connection, userID)
		if err != nil {
			logging.Info(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		requests := [][12]any{}
		userAgentIDs := make(map[int]struct{})
		currentPage := 1
		if targetPage != 0 {
			currentPage = targetPage
		}

		for {
			// Note: table joins currently avoided due to memory limitations
			query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
			offset := (currentPage - 1) * pageSize
			rows, err := connection.Query(context.Background(), query, apiKey, pageSize, offset)
			if err != nil {
				logging.Info(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
				return
			}

			// First value in the list will hold column names
			request := new(DashboardRequestRow)
			var count int
			var skipped int
			for rows.Next() {
				err = rows.Scan(
					&request.IPAddress, 
					&request.Path, 
					&request.Hostname, 
					&request.UserAgent, 
					&request.Method, 
					&request.ResponseTime, 
					&request.Status, 
					&request.Location, 
					&request.UserID, 
					&request.CreatedAt, 
					&request.Referrer,
				)
				if err != nil {
					skipped++
					continue
				}

				var ipAddress string
				if request.IPAddress.IPNet != nil {
					ipAddress = request.IPAddress.IPNet.IP.String()
				}
				hostname := getNullableString(request.Hostname)
				location := getNullableString(request.Location)
				referrer := getNullableString(request.Referrer)
				userID := getNullableString(request.UserID)
				requests = append(
					requests, [12]any{
						ipAddress, 
						request.Path, 
						hostname, 
						request.UserAgent, 
						request.Method, 
						request.ResponseTime, 
						request.Status, 
						location, 
						userID, 
						request.CreatedAt, 
						referrer,
					},
				)
				if request.UserAgent != nil {
					if _, ok := userAgentIDs[*request.UserAgent]; !ok {
						userAgentIDs[*request.UserAgent] = struct{}{}
					}
				}

				count++

				// Finish page read early if reached data limit
				if len(requests) >= maxLoad {
					break
				}
			}
			rows.Close()

			currentPage++

			// Finish data read if only needed one page, last page read was the final page available, or reached data limit
			if targetPage != 0 || count+skipped < pageSize || count >= maxLoad {
				break
			}
		}

		// Convert user agent IDs to names in-place
		userAgents, err := database.GetUserAgentsWithConnection(context.Background(), connection, userAgentIDs)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User agent lookup failed."})
			return
		}

		body := DashboardData{
			UserAgents: userAgents,
			Requests:   requests,
		}

		// Compress requests with gzip
		gzipOutput, err := compressJSON(body)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
			return
		}

		// Return API request data
		c.Writer.Header().Set("Accept-Encoding", "gzip")
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Data(http.StatusOK, "gzip", gzipOutput)

		// Record user dashboard access
		if targetPage == 0 {
			message = fmt.Sprintf("key=%s: Dashboard access successful [%d]", apiKey, len(requests))
		} else {
			message = fmt.Sprintf("key=%s: Dashboard page %d access successful [%d]", apiKey, targetPage, len(requests))
		}
		logging.Info(message)

		err = database.UpdateLastAccessedWithConnection(context.Background(), connection, apiKey)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}
	}
}

func getPaginatedRequestsHandler() gin.HandlerFunc {
	var pageSize int = getPageSize()

	return func(c *gin.Context) {
		var userID string = c.Param("userID")
		if userID == "" {
			logging.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		page, err := strconv.Atoi(c.Param("page"))
		if err != nil || page == 0 {
			logging.Info("Invalid page number")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
			return
		}

		logging.Info(fmt.Sprintf("id=%s: Dashboard page %d access", userID, page))

		connection, err := database.NewConnection()
		if err != nil {
			logging.Info(fmt.Sprintf("id=%s: Dashboard page %d access failed - %s", userID, page, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}
		defer connection.Close(context.Background())

		// Fetch API key corresponding with user ID
		apiKey, err := database.GetAPIKeyWithConnection(context.Background(), connection, userID)
		if err != nil {
			logging.Info(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		requests := [][12]any{}
		userAgentIDs := make(map[int]struct{})

		query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
		rows, err := connection.Query(context.Background(), query, apiKey, pageSize, (page-1)*pageSize)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}
		request := new(DashboardRequestRow) // Reuseable request struct
		for rows.Next() {
			err = rows.Scan(
				&request.IPAddress, 
				&request.Path, 
				&request.Hostname, 
				&request.UserAgent, 
				&request.Method, 
				&request.ResponseTime, 
				&request.Status, 
				&request.Location, 
				&request.UserID, 
				&request.CreatedAt, 
				&request.Referrer,
			)
			if err != nil {
				continue
			}

			var ipAddress string
			if request.IPAddress.IPNet != nil {
				ipAddress = request.IPAddress.IPNet.IP.String()
			}
			hostname := getNullableString(request.Hostname)
			location := getNullableString(request.Location)
			referrer := getNullableString(request.Referrer)
			userID := getNullableString(request.UserID)
			requests = append(
				requests, [12]any{
					ipAddress, 
					request.Path, 
					hostname, 
					request.UserAgent, 
					request.Method, 
					request.ResponseTime, 
					request.Status, 
					location, 
					userID, 
					request.CreatedAt, 
					referrer,
				},
			)
			if request.UserAgent != nil {
				if _, ok := userAgentIDs[*request.UserAgent]; !ok {
					userAgentIDs[*request.UserAgent] = struct{}{}
				}
			}
		}
		rows.Close()

		// Convert user agent IDs to names
		userAgents, err := database.GetUserAgentsWithConnection(context.Background(), connection, userAgentIDs)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
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
			logging.Info(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
			return
		}

		// Return API request data
		c.Writer.Header().Set("Accept-Encoding", "gzip")
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Data(http.StatusOK, "gzip", gzipOutput)

		logging.Info(fmt.Sprintf("key=%s: Dashboard page %d access successful [%d]", apiKey, page, len(requests)))

		// Record user dashboard access
		err = database.UpdateLastAccessedWithConnection(context.Background(), connection, apiKey)
		if err != nil {
			logging.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}
	}
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

func buildRequestDataCompact(rows pgx.Rows, cols [12]any) [][12]any {
	// First value in list holds column names
	requests := [][12]any{cols}
	var request DashboardRequestRow
	for rows.Next() {
		err := rows.Scan(
			&request.IPAddress, 
			&request.Path, 
			&request.Hostname, 
			&request.UserAgent, 
			&request.Method, 
			&request.ResponseTime, 
			&request.Status, 
			&request.Location, 
			&request.UserID, 
			&request.CreatedAt, 
			&request.Referrer,
			)
		if err == nil {
			requests = append(
				requests, [12]any{
					request.IPAddress, 
					request.Path, 
					request.Hostname, 
					request.UserAgent, 
					request.Method, 
					request.ResponseTime, 
					request.Status, 
					request.Location, 
					request.UserID, 
					request.CreatedAt, 
					request.Referrer,
				},
			)
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
			logging.Info("API key empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}
	}

	logging.Info(fmt.Sprintf("key=%s: Data access", apiKey))

	// Get any queries from url
	queries := getQueriesFromRequest(c)

	connection, err := database.NewConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Data access failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}
	defer connection.Close(context.Background())

	// Fetch all API request data associated with this account
	query, arguments := buildDataFetchQuery(apiKey, queries)
	rows, err := connection.Query(context.Background(), query, arguments...)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Queries failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	// Read data into list of objects to return
	if queries.compact {
		cols := [12]any{
			"ip_address", 
			"path", 
			"hostname", 
			"user_agent", 
			"method", 
			"response_time", 
			"status", 
			"location", 
			"user_id", 
			"created_at", 
			"referrer",
		}
		requests := buildRequestDataCompact(rows, cols)
		logging.Info(fmt.Sprintf("key=%s: Data access successful [%d]", apiKey, len(requests)-1))
		c.JSON(http.StatusOK, requests)
	} else {
		requests := buildRequestData(rows)
		logging.Info(fmt.Sprintf("key=%s: Data access successful [%d]", apiKey, len(requests)))
		c.JSON(http.StatusOK, requests)
	}

	rows.Close()

	err = database.UpdateLastAccessedWithConnection(context.Background(), connection, apiKey)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}
}

func buildDataFetchQuery(apiKey string, queries DataFetchQueries) (string, []any) {
	var query strings.Builder
	query.WriteString("SELECT r.ip_address, r.path, r.hostname, u.user_agent, r.method, r.response_time, r.status, r.location, r.user_id, r.created_at, r.referrer FROM requests r JOIN user_agents u ON r.user_agent_id = u.id WHERE api_key = $1")

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
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`
	// UserHash     string    `json:"user_hash"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	Location     string    `json:"location"`
	Referrer     string    `json:"referrer"`
	UserID       string    `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type RequestRow struct {
	Hostname  *string     `json:"hostname"`
	IPAddress pgtype.CIDR `json:"ip_address"`
	Path      string      `json:"path"`
	UserAgent *string     `json:"user_agent"`
	// UserHash     *string     `json:"user_hash"`
	Method       int16     `json:"method"`
	Status       int16     `json:"status"`
	ResponseTime int16     `json:"response_time"`
	Location     *string   `json:"location"`
	Referrer     *string   `json:"referrer"`
	UserID       *string   `json:"user_id"` // Custom user identifier field specific to each API service
	CreatedAt    time.Time `json:"created_at"`
}

func buildRequestData(rows pgx.Rows) []RequestData {
	requests := make([]RequestData, 0)
	var request RequestRow
	for rows.Next() {
		err := rows.Scan(
			&request.IPAddress, 
			&request.Path, 
			&request.Hostname, 
			&request.UserAgent, 
			&request.Method, 
			&request.ResponseTime, 
			&request.Status, 
			&request.Location, 
			&request.UserID, 
			&request.CreatedAt, 
			&request.Referrer,
		)
		if err == nil {
			var ip string
			if request.IPAddress.IPNet != nil {
				ip = request.IPAddress.IPNet.IP.String()
			}
			hostname := getNullableString(request.Hostname)
			userAgent := getNullableString(request.UserAgent)
			location := getNullableString(request.Location)
			referrer := getNullableString(request.Referrer)
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
				Referrer:     referrer,
				UserID:       userID,
				CreatedAt:    request.CreatedAt,
			})
		}
	}

	return requests
}

func deleteData(c *gin.Context) {
	apiKey := c.Param("apiKey")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
	}

	err := database.DeleteUserAccount(apiKey)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Data deletion failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
}

type MonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserMonitor(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	connection, err := database.NewConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Monitor access failed - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
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
		logging.Info("Invalid monitor to add")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if monitor.UserID == "" {
		logging.Info("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	logging.Info(fmt.Sprintf("id=%s: Add monitor", monitor.UserID))

	connection, err := database.NewConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Monitor creation failed - %s", monitor.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	defer connection.Close(context.Background())

	// Get API key from user ID
	apiKey, err := database.GetAPIKeyWithConnection(context.Background(), connection, monitor.UserID)
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", monitor.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Check if monitor already exists
	var count int
	query := "SELECT count(*) FROM monitor WHERE api_key = $1 AND url = $2;"
	err = connection.QueryRow(context.Background(), query, apiKey, monitor.URL).Scan(&count)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Failed to get monitor count - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	if count == 1 {
		logging.Info(fmt.Sprintf("key=%s: Monitor already exists", apiKey))
		c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "message": "Monitor already exists."})
		return
	}

	var monitorCount int
	query = "SELECT count(*) FROM monitor WHERE api_key = $1;"
	err = connection.QueryRow(context.Background(), query, apiKey).Scan(&monitorCount)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Failed to get monitor count - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	// Check if existing monitors already at max limit
	if monitorCount >= 3 {
		logging.Info(fmt.Sprintf("key=%s: Monitor limit reached [%d]", apiKey, monitorCount))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Monitor limit reached."})
		return
	}

	// Insert new monitor into database
	query = "INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ($1, $2, $3, $4, NOW())"
	_, err = connection.Exec(context.Background(), query, apiKey, monitor.URL, monitor.Secure, monitor.Ping)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Failed to create new monitor - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	logging.Info(fmt.Sprintf("key=%s: Monitor '%s' created successfully", apiKey, monitor.URL))

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor created successfully."})
}

func deleteUserMonitor(c *gin.Context) {
	var body struct {
		UserID string `json:"user_id"`
		URL    string `json:"url"`
	}
	err := c.BindJSON(&body)
	if err != nil {
		logging.Info(fmt.Sprintf("Invalid monitor to delete - %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
		return
	}

	if body.UserID == "" {
		logging.Info("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
		return
	}

	logging.Info(fmt.Sprintf("id=%s: Delete monitor", body.UserID))

	connection, err := database.NewConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Monitor deletion failed - %s", body.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	defer connection.Close(context.Background())

	// Get API key from user ID
	apiKey, err := database.GetAPIKeyWithConnection(context.Background(), connection, body.UserID)
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", body.UserID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Delete monitor from database
	err = database.DeleteURLMonitorWithConnection(context.Background(), connection, apiKey, body.URL)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Failed to delete monitor - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}
	// Delete recorded pings from database for this monitor
	err = database.DeleteURLPingsWithConnection(context.Background(), connection, apiKey, body.URL)
	if err != nil {
		logging.Info(fmt.Sprintf("key=%s: Failed to delete pings - %s", apiKey, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	logging.Info(fmt.Sprintf("key=%s: Monitor '%s' deleted successfully", apiKey, body.URL))

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
		logging.Info("User ID empty")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	logging.Info(fmt.Sprintf("id=%s: Monitor access", userID))

	connection, err := database.NewConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Monitor access failed - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}
	defer connection.Close(context.Background())

	// Fetch user ID corresponding with API key
	query := "SELECT url FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
	rows, err := connection.Query(context.Background(), query, userID)
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: Monitor access failed - %s", userID, err.Error()))
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
		logging.Info(fmt.Sprintf("id=%s: Ping access failed - %s", userID, err.Error()))
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
	err = database.UpdateLastAccessedByUserIDWithConnection(context.Background(), connection, userID)
	if err != nil {
		logging.Info(fmt.Sprintf("id=%s: User last access update failed - %s", userID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
		return
	}

	logging.Info(fmt.Sprintf("id=%s: Monitor access successful [%d]", userID, len(monitors)))

	c.JSON(http.StatusOK, monitors)
}

func checkHealth(c *gin.Context) {
	err := database.CheckConnection()
	if err != nil {
		logging.Info(fmt.Sprintf("Health check failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/generate", genAPIKey)
	r.GET("/generate-api-key", genAPIKey)
	r.GET("/user-id/:apiKey", getUserID)
	r.GET("/requests/:userID", getRequestsHandler())
	r.GET("/requests/:userID/:page", getPaginatedRequestsHandler())
	r.GET("/delete/:apiKey", deleteData)
	r.GET("/monitor/pings/:userID", getUserPings)
	r.POST("/monitor/add", addUserMonitor)
	r.POST("/monitor/delete", deleteUserMonitor)
	r.GET("/data", getData)
	r.GET("/health", checkHealth)
}
