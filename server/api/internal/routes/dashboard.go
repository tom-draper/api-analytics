package routes

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/api/internal/config"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
)

type DashboardData struct {
	UserAgents UserAgentsLookup `json:"user_agents"`
	// Each row has 11 positional columns, matching the dashboard's ColumnIndex
	// enum (IPAddress..Referrer, 0-10).
	Requests [][11]any `json:"requests"`
	// HasMore reports whether another page follows this one. It is computed
	// server-side from the full page (returned rows plus any skipped on scan
	// errors), so the dashboard never hardcodes the page size and cannot stop
	// early on a page that was full but had a row skipped.
	HasMore bool `json:"has_more"`
}

type UserAgentsLookup map[int]string

type DashboardRequestRow struct {
	Hostname     *string     `json:"hostname"`
	IPAddress    pgtype.CIDR `json:"ip_address"`
	Path         string      `json:"path"`
	UserAgent    *int        `json:"user_agent"`
	Referrer     *string     `json:"referrer"`
	Method       int16       `json:"method"`
	Status       int16       `json:"status"`
	ResponseTime int16       `json:"response_time"`
	Location     *string     `json:"location"`
	UserID       *string     `json:"user_id"`
	CreatedAt    time.Time   `json:"created_at"`
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
	Referrer     string    `json:"referrer"`
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
	Referrer     *string     `json:"referrer"`
	UserID       *string     `json:"user_id"`
	CreatedAt    time.Time   `json:"created_at"`
}

func fetchAndFormatRequestsPage(ctx context.Context, db *database.DB, apiKey string, page, pageSize int) (requests [][11]any, userAgentIDs map[int]struct{}, count, skipped int, err error) {
	query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
	offset := (page - 1) * pageSize
	rows, err := db.Pool.Query(ctx, query, apiKey, pageSize, offset)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer rows.Close()

	requests = make([][11]any, 0)
	userAgentIDs = make(map[int]struct{})
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
			skipped++
			continue
		}

		var ipAddress string
		if request.IPAddress.IPNet != nil {
			ipAddress = request.IPAddress.IPNet.IP.String()
		}

		requests = append(requests, [11]any{
			ipAddress,
			request.Path,
			getNullableString(request.Hostname),
			request.UserAgent,
			request.Method,
			request.ResponseTime,
			request.Status,
			getNullableString(request.Location),
			getNullableString(request.UserID),
			request.CreatedAt,
			getNullableString(request.Referrer),
		})

		if request.UserAgent != nil {
			userAgentIDs[*request.UserAgent] = struct{}{}
		}

		count++
	}

	return requests, userAgentIDs, count, skipped, rows.Err()
}

func sendDashboardResponse(c *gin.Context, db *database.DB, ctx context.Context, apiKey string, requests [][11]any, userAgentIDs map[int]struct{}, hasMore bool) error {
	userAgents, err := db.GetUserAgents(ctx, userAgentIDs)
	if err != nil {
		log.Error(fmt.Sprintf("key=%s: user agent lookup failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
		return err
	}

	// Update last-accessed before streaming: once the body starts we have already
	// committed a 200 and can no longer fall back to an error status.
	if err := db.UpdateLastAccessed(ctx, apiKey); err != nil {
		log.Error(fmt.Sprintf("key=%s: user last access update failed - %s", apiKey, err.Error()))
	}

	c.Writer.Header().Set("Vary", "Accept-Encoding")
	c.Writer.Header().Set("Content-Encoding", "gzip")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Status(http.StatusOK)

	// Stream JSON straight through gzip into the response so a large full-load
	// result is never fully buffered as JSON bytes and again as a gzip buffer.
	if err := streamGzipJSON(c.Writer, DashboardData{UserAgents: userAgents, Requests: requests, HasMore: hasMore}); err != nil {
		// Status and headers are already sent, so this can only be logged.
		log.Error(fmt.Sprintf("key=%s: failed to write response - %s", apiKey, err.Error()))
		return err
	}
	return nil
}

// streamGzipJSON encodes data as gzip-compressed JSON directly to w, without
// buffering the whole payload as JSON bytes and again as a gzip buffer. It is
// the single place the dashboard response body is serialized.
func streamGzipJSON(w io.Writer, data any) error {
	gzw := gzip.NewWriter(w)
	if err := json.NewEncoder(gzw).Encode(data); err != nil {
		gzw.Close()
		return err
	}
	return gzw.Close()
}

func getRequestsHandler(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			log.Info("user ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Page 0 means load every page. strconv.Atoi also returns 0 on failure,
		// so an unparseable page must be rejected rather than fall through to
		// the full load. A negative page would become a negative OFFSET, which
		// Postgres rejects.
		targetPage := 1
		if pageQuery := c.Query("page"); pageQuery != "" {
			page, err := strconv.Atoi(pageQuery)
			if err != nil || page < 0 || page > maxPageNumber {
				log.Info(fmt.Sprintf("id=%s: failed to parse page number '%s' from query", userID, pageQuery))
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
				return
			}
			targetPage = page
		}

		if targetPage == 0 {
			log.Info(fmt.Sprintf("id=%s: dashboard access", userID))
		} else {
			log.Info(fmt.Sprintf("id=%s: dashboard page %d access", userID, targetPage))
		}

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "User ID not found."})
			} else {
				log.Error(fmt.Sprintf("id=%s: no API key associated with user ID - %s", userID, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			}
			return
		}

		allRequests := make([][11]any, 0)
		allUserAgentIDs := make(map[int]struct{})
		currentPage := 1
		if targetPage != 0 {
			currentPage = targetPage
		}

		for {
			pageRequests, pageUserAgentIDs, count, skipped, err := fetchAndFormatRequestsPage(ctx, db, apiKey, currentPage, cfg.PageSize)
			if err != nil {
				log.Error(fmt.Sprintf("key=%s: failed to fetch requests - %s", apiKey, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
				return
			}
			if skipped > 0 {
				log.Error(fmt.Sprintf("key=%s: skipped %d rows on page %d", apiKey, skipped, currentPage))
			}

			allRequests = append(allRequests, pageRequests...)
			for id := range pageUserAgentIDs {
				allUserAgentIDs[id] = struct{}{}
			}

			currentPage++

			if targetPage != 0 || count+skipped < cfg.PageSize {
				break
			}
			if len(allRequests) >= cfg.MaxLoad {
				log.Info(fmt.Sprintf("key=%s: results capped at max load [%d]", apiKey, cfg.MaxLoad))
				allRequests = allRequests[:cfg.MaxLoad]
				break
			}
		}

		// This handler returns the complete set it intends to (a full load, or a
		// single requested page capped at MaxLoad), so there is never a further
		// page to fetch through it.
		if err := sendDashboardResponse(c, db, ctx, apiKey, allRequests, allUserAgentIDs, false); err != nil {
			return
		}

		if targetPage == 0 {
			log.Info(fmt.Sprintf("key=%s: dashboard access successful [%d]", apiKey, len(allRequests)))
		} else {
			log.Info(fmt.Sprintf("key=%s: dashboard page %d access successful [%d]", apiKey, targetPage, len(allRequests)))
		}
	}
}

func getPaginatedRequestsHandler(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			log.Info("user ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		// Pages are 1-based here: anything below 1 becomes a negative OFFSET,
		// which Postgres rejects.
		page, err := strconv.Atoi(c.Param("page"))
		if err != nil || page < 1 || page > maxPageNumber {
			log.Info("invalid page number")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid page number."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: dashboard page %d access", userID, page))

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "User ID not found."})
			} else {
				log.Error(fmt.Sprintf("id=%s: no API key associated with user ID - %s", userID, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			}
			return
		}

		requests, userAgentIDs, count, skipped, err := fetchAndFormatRequestsPage(ctx, db, apiKey, page, cfg.PageSize)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to fetch requests - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		if skipped > 0 {
			log.Error(fmt.Sprintf("key=%s: skipped %d rows on page %d", apiKey, skipped, page))
		}

		// A full page (returned rows plus any skipped) means another page may
		// follow. Using count+skipped, not just the returned rows, avoids
		// stopping early when a row on a full page failed to scan.
		hasMore := count+skipped >= cfg.PageSize
		if err := sendDashboardResponse(c, db, ctx, apiKey, requests, userAgentIDs, hasMore); err != nil {
			return
		}

		log.Info(fmt.Sprintf("key=%s: dashboard page %d access successful [%d]", apiKey, page, len(requests)))
	}
}

func buildRequestData(rows pgx.Rows) ([]RequestData, int) {
	requests := make([]RequestData, 0)
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
			var ip string
			if request.IPAddress.IPNet != nil {
				ip = request.IPAddress.IPNet.IP.String()
			}
			requests = append(requests, RequestData{
				IPAddress:    ip,
				Path:         request.Path,
				Hostname:     getNullableString(request.Hostname),
				UserAgent:    getNullableString(request.UserAgent),
				Method:       request.Method,
				Status:       request.Status,
				ResponseTime: request.ResponseTime,
				Location:     getNullableString(request.Location),
				Referrer:     getNullableString(request.Referrer),
				UserID:       getNullableString(request.UserID),
				CreatedAt:    request.CreatedAt,
			})
		} else {
			skipped++
		}
	}
	return requests, skipped
}
