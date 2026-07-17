package routes

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

		// Pages are 1-based: anything below 1 becomes a negative OFFSET, which
		// Postgres rejects.
		if queries.page < 1 {
			log.Info(fmt.Sprintf("key=%s: invalid page number %d", apiKey, queries.page))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
			return
		}

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
			cols := [compactCols]any{
				"ip_address", "path", "hostname", "user_agent", "method",
				"response_time", "status", "location", "user_id", "created_at", "referrer",
			}
			requests, skipped := buildRequestDataCompact(rows, cols)
			rows.Close()
			if skipped > 0 {
				log.Error(fmt.Sprintf("key=%s: skipped %d rows during data fetch", apiKey, skipped))
			}
			if err := rows.Err(); err != nil {
				log.Error(fmt.Sprintf("key=%s: rows error during data fetch - %s", apiKey, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
				return
			}
			log.Info(fmt.Sprintf("key=%s: data access successful [%d]", apiKey, len(requests)-1))
			c.JSON(http.StatusOK, requests)
		} else {
			requests, skipped := buildRequestData(rows)
			rows.Close()
			if skipped > 0 {
				log.Error(fmt.Sprintf("key=%s: skipped %d rows during data fetch", apiKey, skipped))
			}
			if err := rows.Err(); err != nil {
				log.Error(fmt.Sprintf("key=%s: rows error during data fetch - %s", apiKey, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
				return
			}
			log.Info(fmt.Sprintf("key=%s: data access successful [%d]", apiKey, len(requests)))
			c.JSON(http.StatusOK, requests)
		}

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
			// created_at is a timestamp, so a bare date compares against midnight.
			// Use < dateTo + 1 day to include the whole dateTo day, matching the
			// single-date branch above.
			query.WriteString(fmt.Sprintf(" and r.created_at < date $%d + interval '1 days'", len(arguments)+1))
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

// normalizeQueryKey reduces a parameter name to a case- and separator-
// insensitive form, so that userId, userID, user_id and userid all resolve to
// the same filter. An unrecognized name applies no filter and returns a wider
// result set than asked for, which is silent and easy to miss, so names are
// matched leniently.
func normalizeQueryKey(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if r == '_' || r == '-' {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// queryLookup indexes a request's query parameters by normalized name.
type queryLookup map[string]string

func newQueryLookup(c *gin.Context) queryLookup {
	values := c.Request.URL.Query()

	// Sort so that a request supplying two spellings of one filter (say user_id
	// and userId) always resolves to the same value rather than to whichever
	// map iteration reached first.
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	lookup := make(queryLookup, len(names))
	for _, name := range names {
		value := values.Get(name)
		if value == "" {
			continue
		}
		if key := normalizeQueryKey(name); key != "" {
			if _, exists := lookup[key]; !exists {
				lookup[key] = value
			}
		}
	}
	return lookup
}

// get returns the value of the first of names that is present.
func (q queryLookup) get(names ...string) string {
	for _, name := range names {
		if value, exists := q[normalizeQueryKey(name)]; exists {
			return value
		}
	}
	return ""
}

func getQueriesFromRequest(c *gin.Context) DataFetchQueries {
	query := newQueryLookup(c)

	page := 1
	if pageQuery := query.get("page"); pageQuery != "" {
		// A present but unparseable page is rejected (page 0 fails the page < 1
		// check in getData) rather than silently serving page 1, matching the
		// /requests routes.
		p, err := strconv.Atoi(pageQuery)
		if err != nil {
			p = 0
		}
		page = p
	}

	status, err := strconv.Atoi(query.get("status"))
	if err != nil {
		status = 0
	}

	return DataFetchQueries{
		page:      page,
		compact:   query.get("compact") == "true",
		date:      parseQueryDate(query.get("date")),
		dateFrom:  parseQueryDate(query.get("dateFrom")),
		dateTo:    parseQueryDate(query.get("dateTo")),
		hostname:  query.get("hostname"),
		ipAddress: query.get("ipAddress", "ip"),
		location:  query.get("location"),
		status:    status,
		userID:    query.get("userId"),
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

// compactCols is the number of columns in a compact response row, matching the
// header written by getData and the column list selected by buildDataFetchQuery.
const compactCols = 11

// buildRequestDataCompact formats rows as positional arrays, prefixed by cols
// as a header row. It scans into RequestRow, whose UserAgent is a string: the
// compact query joins user_agents and selects the user_agent text column, not
// the user_agent_id integer that DashboardRequestRow carries.
func buildRequestDataCompact(rows pgx.Rows, cols [compactCols]any) ([][compactCols]any, int) {
	requests := [][compactCols]any{cols}
	skipped := 0
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
			var ipAddress string
			if request.IPAddress.IPNet != nil {
				ipAddress = request.IPAddress.IPNet.IP.String()
			}
			requests = append(requests, [compactCols]any{
				ipAddress,
				request.Path,
				getNullableString(request.Hostname),
				getNullableString(request.UserAgent),
				request.Method,
				request.ResponseTime,
				request.Status,
				getNullableString(request.Location),
				getNullableString(request.UserID),
				request.CreatedAt,
				getNullableString(request.Referrer),
			})
		} else {
			skipped++
		}
	}
	return requests, skipped
}
