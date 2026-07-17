package routes

import (
	"context"
	"fmt"
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
	Requests   [][12]any        `json:"requests"`
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

func fetchAndFormatRequestsPage(ctx context.Context, db *database.DB, apiKey string, page, pageSize int) (requests [][12]any, userAgentIDs map[int]struct{}, count, skipped int, err error) {
	query := "SELECT ip_address, path, hostname, user_agent_id, method, response_time, status, location, user_id, created_at, referrer FROM requests WHERE api_key = $1 ORDER BY created_at LIMIT $2 OFFSET $3;"
	offset := (page - 1) * pageSize
	rows, err := db.Pool.Query(ctx, query, apiKey, pageSize, offset)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer rows.Close()

	requests = make([][12]any, 0)
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

		requests = append(requests, [12]any{
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

func sendDashboardResponse(c *gin.Context, db *database.DB, ctx context.Context, apiKey string, requests [][12]any, userAgentIDs map[int]struct{}) error {
	userAgents, err := db.GetUserAgents(ctx, userAgentIDs)
	if err != nil {
		log.Error(fmt.Sprintf("key=%s: user agent lookup failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
		return err
	}

	gzipOutput, err := compressJSON(DashboardData{UserAgents: userAgents, Requests: requests})
	if err != nil {
		log.Error(fmt.Sprintf("key=%s: compression failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Compression failed."})
		return err
	}

	if err := db.UpdateLastAccessed(ctx, apiKey); err != nil {
		log.Error(fmt.Sprintf("key=%s: user last access update failed - %s", apiKey, err.Error()))
	}

	c.Writer.Header().Set("Accept-Encoding", "gzip")
	c.Writer.Header().Set("Content-Encoding", "gzip")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Data(http.StatusOK, "gzip", gzipOutput)
	return nil
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
		// the full load.
		targetPage := 1
		if pageQuery := c.Query("page"); pageQuery != "" {
			page, err := strconv.Atoi(pageQuery)
			if err != nil {
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

		allRequests := make([][12]any, 0)
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

		if err := sendDashboardResponse(c, db, ctx, apiKey, allRequests, allUserAgentIDs); err != nil {
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

		page, err := strconv.Atoi(c.Param("page"))
		if err != nil || page == 0 {
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

		requests, userAgentIDs, _, skipped, err := fetchAndFormatRequestsPage(ctx, db, apiKey, page, cfg.PageSize)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to fetch requests - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		if skipped > 0 {
			log.Error(fmt.Sprintf("key=%s: skipped %d rows on page %d", apiKey, skipped, page))
		}

		if err := sendDashboardResponse(c, db, ctx, apiKey, requests, userAgentIDs); err != nil {
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
