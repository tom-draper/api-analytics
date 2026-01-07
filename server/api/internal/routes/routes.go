package routes

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/api/internal/env"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		apiKey, err := db.CreateUser(ctx)
		if err != nil {
			log.Info(fmt.Sprintf("API key generation failed - %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key generation failed."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: API key generation successful", apiKey))

		c.JSON(http.StatusOK, apiKey)
	}
}

func getUserID(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string = c.Param("apiKey")
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		ctx := c.Request.Context()

		// Get user ID associated with API key
		userID, err := db.GetUserID(ctx, apiKey)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User ID fetch failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		c.JSON(http.StatusOK, userID)
	}
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

func getRequestsHandler(db *database.DB) gin.HandlerFunc {
	var pageSize int = getPageSize()
	var maxLoad int = getMaxLoad()

	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			log.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		var err error
		pageQuery := c.Query("page")
		log.Info(pageQuery)
		targetPage := 1
		if pageQuery != "" {
			targetPage, err = strconv.Atoi(pageQuery)
			if err != nil {
				log.Info(fmt.Sprintf("Failed to parse page number '%s' from query", pageQuery))
			}
		}

		var message string
		if targetPage == 0 {
			message = fmt.Sprintf("id=%s: Dashboard access", userID)
		} else {
			message = fmt.Sprintf("id=%s: Dashboard page %d access", userID, targetPage)
		}
		log.Info(message)

		ctx := c.Request.Context()
		// Fetch API key corresponding with user ID
		apiKey, err := db.GetAPIKey(ctx, userID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
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
			query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
			offset := (currentPage - 1) * pageSize
			rows, err := db.Pool.Query(ctx, query, apiKey, pageSize, offset)
			if err != nil {
				log.Info(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
				return
			}

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

				if len(requests) >= maxLoad {
					break
				}
			}
			rows.Close()

			currentPage++

			if targetPage != 0 || count+skipped < pageSize || count >= maxLoad {
				break
			}
		}

		userAgents, err := db.GetUserAgents(ctx, userAgentIDs)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User agent lookup failed."})
			return
		}

		body := DashboardData{
			UserAgents: userAgents,
			Requests:   requests,
		}

		gzipOutput, err := compressJSON(body)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
			return
		}

		// Update last accessed BEFORE sending response
		err = db.UpdateLastAccessed(ctx, apiKey)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
			// Don't return error to user, just log it
		}

		// Return API request data
		c.Writer.Header().Set("Accept-Encoding", "gzip")
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Data(http.StatusOK, "gzip", gzipOutput)

		// Record successful access
		if targetPage == 0 {
			message = fmt.Sprintf("key=%s: Dashboard access successful [%d]", apiKey, len(requests))
		} else {
			message = fmt.Sprintf("key=%s: Dashboard page %d access successful [%d]", apiKey, targetPage, len(requests))
		}
		log.Info(message)
	}
}

func getPaginatedRequestsHandler(db *database.DB) gin.HandlerFunc {
	var pageSize int = getPageSize()

	return func(c *gin.Context) {
		var userID string = c.Param("userID")
		if userID == "" {
			log.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		page, err := strconv.Atoi(c.Param("page"))
		if err != nil || page == 0 {
			log.Info("Invalid page number")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: Dashboard page %d access", userID, page))

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, userID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: No API key associated with user ID - %s", userID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		requests := [][12]any{}
		userAgentIDs := make(map[int]struct{})

		query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
		rows, err := db.Pool.Query(ctx, query, apiKey, pageSize, (page-1)*pageSize)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Invalid API key - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		request := new(DashboardRequestRow)
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

		userAgents, err := db.GetUserAgents(ctx, userAgentIDs)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User agent lookup failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User agent lookup failed."})
			return
		}

		body := DashboardData{
			UserAgents: userAgents,
			Requests:   requests,
		}

		gzipOutput, err := compressJSON(body)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Compression failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
			return
		}

		// Update last accessed before sending response
		err = db.UpdateLastAccessed(ctx, apiKey)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
		}

		c.Writer.Header().Set("Accept-Encoding", "gzip")
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Data(http.StatusOK, "gzip", gzipOutput)

		log.Info(fmt.Sprintf("key=%s: Dashboard page %d access successful [%d]", apiKey, page, len(requests)))
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

func getData(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := getAPIKeyFromHeader(c)
		if apiKey == "" {
			log.Info("API key empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: Data access", apiKey))

		// Get any queries from url
		queries := getQueriesFromRequest(c)

		ctx := c.Request.Context()

		// Fetch all API request data associated with this account
		query, arguments := buildDataFetchQuery(apiKey, queries)
		rows, err := db.Pool.Query(ctx, query, arguments...)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Queries failed - %s", apiKey, err.Error()))
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
			log.Info(fmt.Sprintf("key=%s: Data access successful [%d]", apiKey, len(requests)-1))
			c.JSON(http.StatusOK, requests)
		} else {
			requests := buildRequestData(rows)
			log.Info(fmt.Sprintf("key=%s: Data access successful [%d]", apiKey, len(requests)))
			c.JSON(http.StatusOK, requests)
		}

		rows.Close()

		err = db.UpdateLastAccessed(ctx, apiKey)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: User last access update failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}
	}
}

func getAPIKeyFromHeader(c *gin.Context) string {
	apiKey := c.GetHeader("X-AUTH-TOKEN")
	if apiKey == "" {
		// Check old (deprecated) identifier
		apiKey = c.GetHeader("API-Key")
		if apiKey == "" {
			return ""
		}
	}
	return apiKey
}

func buildDataFetchQuery(apiKey string, queries DataFetchQueries) (string, []any) {
	var query strings.Builder
	query.WriteString("SELECT r.ip_address, r.path, r.hostname, u.user_agent, r.method, r.response_time, r.status, r.location, r.user_id, r.created_at, r.referrer FROM requests r JOIN user_agents u ON r.user_agent_id = u.id WHERE api_key = $1")

	arguments := []any{apiKey}

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

	// Use parameterized query for pagination as well
	const pageSize = 50_000
	offset := (queries.page - 1) * pageSize
	query.WriteString(fmt.Sprintf(" ORDER BY created_at LIMIT $%d OFFSET $%d;", len(arguments)+1, len(arguments)+2))
	arguments = append(arguments, pageSize, offset)

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

func deleteData(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.Param("apiKey")
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
		}

		ctx := c.Request.Context()

		err := db.DeleteUserAccount(ctx, apiKey)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Data deletion failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
	}
}

type MonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		ctx := c.Request.Context()

		// Retreive monitors created by this user
		query := "SELECT url, secure, ping, monitor.created_at FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
		rows, err := db.Pool.Query(ctx, query, userID)
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
}

type Monitor struct {
	UserID string `json:"user_id"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

func addUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var monitor Monitor
		err := c.BindJSON(&monitor)
		if err != nil {
			log.Info("Invalid monitor to add")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if monitor.UserID == "" {
			log.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: Add monitor", monitor.UserID))

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, monitor.UserID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", monitor.UserID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Use a transaction to prevent race conditions
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to start transaction - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}
		defer tx.Rollback(ctx)

		// Check monitor count within transaction
		var monitorCount int
		query := "SELECT count(*) FROM monitor WHERE api_key = $1 FOR UPDATE;"
		err = tx.QueryRow(ctx, query, apiKey).Scan(&monitorCount)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to get monitor count - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		if monitorCount >= 3 {
			log.Info(fmt.Sprintf("key=%s: Monitor limit reached [%d]", apiKey, monitorCount))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Monitor limit reached."})
			return
		}

		// Try to insert, handle conflict with ON CONFLICT
		query = "INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (api_key, url) DO NOTHING RETURNING url"
		var insertedURL string
		err = tx.QueryRow(ctx, query, apiKey, monitor.URL, monitor.Secure, monitor.Ping).Scan(&insertedURL)
		if err == pgx.ErrNoRows {
			log.Info(fmt.Sprintf("key=%s: Monitor already exists", apiKey))
			c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "message": "Monitor already exists."})
			return
		}
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to create new monitor - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to commit transaction - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: Monitor '%s' created successfully", apiKey, monitor.URL))
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor created successfully."})
	}
}

func deleteUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			UserID string `json:"user_id"`
			URL    string `json:"url"`
		}
		err := c.BindJSON(&body)
		if err != nil {
			log.Info(fmt.Sprintf("Invalid monitor to delete - %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if body.UserID == "" {
			log.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: Delete monitor", body.UserID))

		ctx := c.Request.Context()

		// Get API key associated with this user ID
		apiKey, err := db.GetAPIKey(ctx, body.UserID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: Invalid monitor user ID - %s", body.UserID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Delete URL monitor
		err = db.DeleteURLMonitor(ctx, apiKey, body.URL)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to delete monitor - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		// Delete all pings associated with this URL
		err = db.DeleteURLPings(ctx, apiKey, body.URL)
		if err != nil {
			log.Info(fmt.Sprintf("key=%s: Failed to delete pings - %s", apiKey, err.Error()))
			// Continue even if ping deletion fails
		}

		log.Info(fmt.Sprintf("key=%s: Monitor '%s' deleted successfully", apiKey, body.URL))

		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Monitor deleted successfully."})
	}
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

func getUserPings(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID string = c.Param("userID")
		if userID == "" {
			log.Info("User ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: Monitor access", userID))

		ctx := c.Request.Context()

		// Fetch user ID corresponding with API key
		query := "SELECT url FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
		rows, err := db.Pool.Query(ctx, query, userID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: Monitor access failed - %s", userID, err.Error()))
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
		rows, err = db.Pool.Query(ctx, query, userID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: Ping access failed - %s", userID, err.Error()))
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
		err = db.UpdateLastAccessedByUserID(ctx, userID)
		if err != nil {
			log.Info(fmt.Sprintf("id=%s: User last access update failed - %s", userID, err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: Monitor access successful [%d]", userID, len(monitors)))

		c.JSON(http.StatusOK, monitors)
	}
}

func checkHealth(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if err := db.CheckConnection(ctx); err != nil {
			log.Info(fmt.Sprintf("Health check failed: %v", err))
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
}

func RegisterRouter(r *gin.RouterGroup, db *database.DB) {
	r.GET("/generate", genAPIKey(db))
	r.GET("/generate-api-key", genAPIKey(db))
	r.GET("/user-id/:apiKey", getUserID(db))
	r.GET("/requests/:userID", getRequestsHandler(db))
	r.GET("/requests/:userID/:page", getPaginatedRequestsHandler(db))
	r.GET("/delete/:apiKey", deleteData(db))
	r.GET("/monitor/:userID", getUserMonitor(db))
	r.GET("/monitor/pings/:userID", getUserPings(db))
	r.POST("/monitor/add", addUserMonitor(db))
	r.POST("/monitor/delete", deleteUserMonitor(db))
	r.GET("/data", getData(db))
	r.GET("/health", checkHealth(db))
}
