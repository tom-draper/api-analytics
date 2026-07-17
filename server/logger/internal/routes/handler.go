package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"
)

const (
	P1 PrivacyLevel = iota
	P2
	P3
)

type PrivacyLevel int

type RequestData struct {
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       int16  `json:"status"`
	Referrer     string `json:"referrer"`
	ResponseTime int16  `json:"response_time"`
	UserID       string `json:"user_id"`
	CreatedAt    string `json:"created_at"`
}

type Payload struct {
	APIKey       string        `json:"api_key"`
	Requests     []RequestData `json:"requests"`
	Framework    string        `json:"framework"`
	PrivacyLevel PrivacyLevel  `json:"privacy_level"`
}

type ProcessedRequest struct {
	Path         string
	Hostname     string
	IPAddress    *string
	UserHash     string
	Referrer     string
	Status       int16
	ResponseTime int16
	Method       int16
	Framework    int16
	Location     string
	UserID       string
	UserAgent    string
	CreatedAt    time.Time
	UserAgentID  int
}

const frameworkOther int16 = 255

var methodID = map[string]int16{
	"GET": 0, "POST": 1, "PUT": 2, "PATCH": 3, "DELETE": 4,
	"OPTIONS": 5, "CONNECT": 6, "HEAD": 7, "TRACE": 8,
}

var frameworkID = map[string]int16{
	"FastAPI": 0, "Flask": 1, "Gin": 2, "Echo": 3, "Express": 4,
	"Fastify": 5, "Koa": 6, "Chi": 7, "Fiber": 8, "Actix": 9,
	"Axum": 10, "Tornado": 11, "Django": 12, "Rails": 13, "Laravel": 14,
	"Sinatra": 15, "Rocket": 16, "ASP.NET Core": 17, "Hono": 18,
}

func logRequestHandler(db *database.DB, geoIPDB *geoip2.Reader, cache *Cache, rateLimit int, maxInsert int) gin.HandlerFunc {
	rateLimiter := ratelimit.NewRateLimiter(rateLimit)

	return func(c *gin.Context) {
		var payload Payload
		if err := c.ShouldBindJSON(&payload); err != nil {
			msg := fmt.Sprintf("Invalid request data: %s", err.Error())
			log.LogClientError(c.ClientIP(), "", msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		// Clean up API key (users sometimes provide it in quotes)
		payload.APIKey = strings.TrimSpace(strings.ReplaceAll(payload.APIKey, "\"", ""))

		if payload.APIKey == "" {
			msg := "API key required."
			log.LogClientError(c.ClientIP(), "", msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		if !database.ValidAPIKey(payload.APIKey) {
			msg := "Invalid API key format. Expected UUID format."
			log.LogClientError(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		ctx := c.Request.Context()

		var exists bool
		if err := db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE api_key = $1)", payload.APIKey).Scan(&exists); err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to verify API key: %v", payload.APIKey, err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		if !exists {
			msg := "API key not found."
			log.LogClientError(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": msg})
			return
		}

		if rateLimiter.RateLimited(payload.APIKey) {
			msg := "Too many requests."
			log.LogClientError(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusTooManyRequests, gin.H{"status": http.StatusTooManyRequests, "message": msg})
			return
		}

		if len(payload.Requests) == 0 {
			msg := "Payload contains no logged requests."
			log.LogClientError(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		framework, ok := frameworkID[payload.Framework]
		if !ok {
			framework = frameworkOther
		}

		validRequests := make([]ProcessedRequest, 0, len(payload.Requests))
		userAgents := make([]string, 0, len(payload.Requests))
		uniqueUserAgents := make(map[string]struct{})

		for _, request := range payload.Requests {
			if len(validRequests) >= maxInsert {
				break
			}

			method, ok := methodID[request.Method]
			if !ok {
				continue
			}

			if len(request.UserAgent) > 255 {
				request.UserAgent = request.UserAgent[:255]
			}
			if !database.ValidUserAgent(request.UserAgent) {
				continue
			}

			if len(request.UserID) > 255 {
				request.UserID = request.UserID[:255]
			}
			if request.UserID != "" && !database.ValidUserID(request.UserID) {
				continue
			}

			if len(request.Hostname) > 255 {
				request.Hostname = request.Hostname[:255]
			}
			if !database.ValidHostname(request.Hostname) {
				continue
			}

			if len(request.Path) > 255 {
				request.Path = request.Path[:255]
			}
			if !database.ValidPath(request.Path) {
				continue
			}

			if len(request.Referrer) > 255 {
				request.Referrer = request.Referrer[:255]
			}
			if request.Referrer != "" && !database.ValidString(request.Referrer) {
				continue
			}

			if request.ResponseTime < 0 {
				continue
			}

			var ipAddress *string
			if payload.PrivacyLevel <= P1 && request.IPAddress != "" {
				ipAddress = &request.IPAddress
			}

			location := getCountryCode(geoIPDB, cache, request.IPAddress)
			userHash := getUserHash(request.IPAddress, request.UserAgent)

			createdAt, err := time.Parse(time.RFC3339Nano, request.CreatedAt)
			if err != nil {
				createdAt = time.Now().UTC()
			}

			if _, seen := uniqueUserAgents[request.UserAgent]; !seen {
				uniqueUserAgents[request.UserAgent] = struct{}{}
				userAgents = append(userAgents, request.UserAgent)
			}

			validRequests = append(validRequests, ProcessedRequest{
				Path:         request.Path,
				Hostname:     request.Hostname,
				IPAddress:    ipAddress,
				UserHash:     userHash,
				Referrer:     request.Referrer,
				Status:       request.Status,
				ResponseTime: request.ResponseTime,
				Method:       method,
				Framework:    framework,
				Location:     location,
				UserID:       request.UserID,
				UserAgent:    request.UserAgent,
				CreatedAt:    createdAt,
			})
		}

		if len(validRequests) == 0 {
			log.Info(fmt.Sprintf("key=%s: no valid requests to insert (received %d)", payload.APIKey, len(payload.Requests)))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		userAgentIDs, err := ensureUserAgentIDs(ctx, db, cache, userAgents)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to ensure user agent IDs: %v", payload.APIKey, err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		for i := range validRequests {
			if id, exists := userAgentIDs[validRequests[i].UserAgent]; exists {
				validRequests[i].UserAgentID = id
			}
		}

		_, err = db.Pool.CopyFrom(
			ctx,
			pgx.Identifier{"requests"},
			[]string{
				"api_key", "path", "hostname", "ip_address", "user_hash",
				"referrer", "status", "response_time", "method", "framework",
				"location", "user_id", "created_at", "user_agent_id",
			},
			pgx.CopyFromSlice(len(validRequests), func(i int) ([]any, error) {
				req := validRequests[i]
				return []any{
					payload.APIKey, req.Path, req.Hostname, req.IPAddress, req.UserHash,
					req.Referrer, req.Status, req.ResponseTime, req.Method, req.Framework,
					req.Location, req.UserID, req.CreatedAt, req.UserAgentID,
				}, nil
			}),
		)

		if err != nil {
			log.Error(fmt.Sprintf(
				"copy failed: api_key=%s request_count=%d error=%v",
				payload.APIKey, len(validRequests), err,
			))
			if pgErr, ok := err.(*pgconn.PgError); ok {
				log.Error(fmt.Sprintf(
					"postgres error: code=%s message=%s detail=%s hint=%s",
					pgErr.Code, pgErr.Message, pgErr.Detail, pgErr.Hint,
				))
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Database insert failed.",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API requests logged successfully."})
		log.LogRequestsInserted(payload.APIKey, len(validRequests), len(payload.Requests))
	}
}
