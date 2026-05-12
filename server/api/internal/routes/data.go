package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
)

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
			log.Info("api key missing")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required in X-AUTH-TOKEN header."})
			return
		}
		if !database.ValidAPIKey(apiKey) {
			log.Info("api key invalid format")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key format. Expected UUID format."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: data access", apiKey))

		queries := getQueriesFromRequest(c)
		ctx := c.Request.Context()

		query, arguments := buildDataFetchQuery(apiKey, queries)
		rows, err := db.Pool.Query(ctx, query, arguments...)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: queries failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		defer rows.Close()

		if queries.compact {
			cols := [12]any{
				"ip_address", "path", "hostname", "user_agent", "method",
				"response_time", "status", "location", "user_id", "created_at", "referrer",
			}
			requests, skipped := buildRequestDataCompact(rows, cols)
			if skipped > 0 {
				log.Error(fmt.Sprintf("key=%s: skipped %d rows during data fetch", apiKey, skipped))
			}
			log.Info(fmt.Sprintf("key=%s: data access successful [%d]", apiKey, len(requests)-1))
			c.JSON(http.StatusOK, requests)
		} else {
			requests, skipped := buildRequestData(rows)
			if skipped > 0 {
				log.Error(fmt.Sprintf("key=%s: skipped %d rows during data fetch", apiKey, skipped))
			}
			log.Info(fmt.Sprintf("key=%s: data access successful [%d]", apiKey, len(requests)))
			c.JSON(http.StatusOK, requests)
		}

		// Close rows explicitly to release the connection before UpdateLastAccessed
		rows.Close()

		if err := db.UpdateLastAccessed(ctx, apiKey); err != nil {
			log.Error(fmt.Sprintf("key=%s: user last access update failed - %s", apiKey, err.Error()))
		}
	}
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

	const pageSize = 50_000
	offset := (queries.page - 1) * pageSize
	query.WriteString(fmt.Sprintf(" ORDER BY created_at LIMIT $%d OFFSET $%d;", len(arguments)+1, len(arguments)+2))
	arguments = append(arguments, pageSize, offset)

	return query.String(), arguments
}

func getQueriesFromRequest(c *gin.Context) DataFetchQueries {
	page := 1
	if pageQuery := c.Query("page"); pageQuery != "" {
		if p, err := strconv.Atoi(pageQuery); err == nil {
			page = p
		}
	}

	status, err := strconv.Atoi(c.Query("status"))
	if err != nil {
		status = 0
	}

	return DataFetchQueries{
		page:      page,
		compact:   c.Query("compact") == "true",
		date:      parseQueryDate(c.Query("date")),
		dateFrom:  parseQueryDate(c.Query("dateFrom")),
		dateTo:    parseQueryDate(c.Query("dateTo")),
		hostname:  c.Query("hostname"),
		ipAddress: c.Query("ip"),
		location:  c.Query("location"),
		status:    status,
		userID:    c.Query("userID"),
	}
}

func parseQueryDate(date string) time.Time {
	if date == "" {
		return time.Time{}
	}
	if d, err := time.Parse("2006-01-02", date); err == nil {
		return d
	}
	return time.Time{}
}
