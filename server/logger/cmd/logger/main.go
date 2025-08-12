package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/oschwald/geoip2-golang"
	"github.com/jackc/pgx/v5"
)

// Global connection pool and caches
var (
	geoIPDB           *geoip2.Reader
	userAgentCache    = make(map[string]int)
	userAgentCacheMu  sync.RWMutex
	geoIPCacheMap     = make(map[string]string)
	geoIPCacheMu      sync.RWMutex
	maxCacheSize      = 10000
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.LogToFile(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.LogToFile("Starting logger...")

	// Initialize GeoIP database once
	var err error
	geoIPDB, err = geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to open GeoIP database: %v", err))
	}
	defer func() {
		if geoIPDB != nil {
			geoIPDB.Close()
		}
	}()

	// Preload user agent cache
	err = preloadUserAgentCache()
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to preload user agent cache: %v", err))
	}

	err = database.LoadConfig()
	if err != nil {
		log.LogToFile("Failed to load database configuration: " + err.Error())
		return
	}

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.Use(cors.Default())

	handler := logRequestHandler()
	app.POST("/api/log-request", handler)
	app.POST("/api/requests", handler)
	app.GET("/api/health", checkHealth)

	if err := app.Run(":8000"); err != nil {
		log.LogToFile(fmt.Sprintf("Failed to run server: %v", err))
	}
}

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

type PrivacyLevel int

const (
	P1 PrivacyLevel = iota
	P2
	P3
)

// Processed request for batch insertion
type ProcessedRequest struct {
	APIKey        string
	Path          string
	Hostname      string
	IPAddress     *string
	UserHash      string
	Referrer      string
	Status        int16
	ResponseTime  int16
	Method        int16
	Framework     int16
	Location      string
	UserID        string
	CreatedAt     string
	UserAgentID   int
}

func checkHealth(c *gin.Context) {
	connection, err := database.NewConnection()
	if err != nil {
		log.LogToFile(fmt.Sprintf("Health check failed: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}
	defer connection.Close(context.Background())

	err = connection.Ping(context.Background())
	if err != nil {
		log.LogToFile(fmt.Sprintf("Health check failed: %v", err))
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

func getMaxInsert() int {
	const defaultValue int = 2000

	err := godotenv.Load(".env")
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to load .env file. Using default value MAX_INSERT=%d.", defaultValue))
		return defaultValue
	}

	value := os.Getenv("MAX_INSERT")
	if value == "" {
		log.LogToFile(fmt.Sprintf("MAX_INSERT environment variable is blank. Using default value MAX_INSERT=%d.", defaultValue))
		return defaultValue
	}

	maxInsert, err := strconv.Atoi(value)
	if err != nil {
		log.LogToFile(fmt.Sprintf("MAX_INSERT environment variable is not an integer. Using default value MAX_INSERT=%d.", defaultValue))
		return defaultValue
	}

	return maxInsert
}

// Cached country code lookup with LRU-style eviction
func getCountryCode(IPAddress string) string {
	if IPAddress == "" || geoIPDB == nil {
		return ""
	}

	// Check cache first
	geoIPCacheMu.RLock()
	if code, exists := geoIPCacheMap[IPAddress]; exists {
		geoIPCacheMu.RUnlock()
		return code
	}
	geoIPCacheMu.RUnlock()

	ip := net.ParseIP(IPAddress)
	if ip == nil {
		return ""
	}

	record, err := geoIPDB.Country(ip)
	if err != nil {
		return ""
	}

	countryCode := record.Country.IsoCode

	// Cache the result with simple size limit
	geoIPCacheMu.Lock()
	if len(geoIPCacheMap) >= maxCacheSize {
		// Simple eviction: clear half the cache
		for k := range geoIPCacheMap {
			delete(geoIPCacheMap, k)
			if len(geoIPCacheMap) <= maxCacheSize/2 {
				break
			}
		}
	}
	geoIPCacheMap[IPAddress] = countryCode
	geoIPCacheMu.Unlock()

	return countryCode
}

// Preload user agent cache at startup
func preloadUserAgentCache() error {
	conn, err := database.NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT user_agent, id FROM user_agents LIMIT 50000")
	if err != nil {
		return err
	}
	defer rows.Close()

	userAgentCacheMu.Lock()
	defer userAgentCacheMu.Unlock()

	for rows.Next() {
		var userAgent string
		var id int
		err := rows.Scan(&userAgent, &id)
		if err != nil {
			continue
		}
		userAgentCache[userAgent] = id
	}

	return rows.Err()
}

// Efficient batch user agent insertion and ID retrieval
func ensureUserAgentIDs(userAgents []string) (map[string]int, error) {
	if len(userAgents) == 0 {
		return make(map[string]int), nil
	}

	result := make(map[string]int)
	newUserAgents := make([]string, 0)

	// Check cache first
	userAgentCacheMu.RLock()
	for _, ua := range userAgents {
		if id, exists := userAgentCache[ua]; exists {
			result[ua] = id
		} else {
			newUserAgents = append(newUserAgents, ua)
		}
	}
	userAgentCacheMu.RUnlock()

	if len(newUserAgents) == 0 {
		return result, nil
	}

	// Use COPY for bulk insert of new user agents
	conn, err := database.NewConnection()
	if err != nil {
		return result, err
	}
	defer conn.Close(context.Background())

	// First, try to insert new user agents using COPY
	_, err = conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"user_agents"},
		[]string{"user_agent"},
		pgx.CopyFromSlice(len(newUserAgents), func(i int) ([]any, error) {
			return []any{newUserAgents[i]}, nil
		}),
	)
	if err != nil {
		// Fallback to individual inserts with ON CONFLICT
		for _, ua := range newUserAgents {
			_, err := conn.Exec(context.Background(),
				"INSERT INTO user_agents (user_agent) VALUES ($1) ON CONFLICT (user_agent) DO NOTHING",
				ua)
			if err != nil {
				log.LogToFile(fmt.Sprintf("Failed to insert user agent: %v", err))
			}
		}
	}

	// Get IDs for new user agents
	if len(newUserAgents) > 0 {
		query := "SELECT user_agent, id FROM user_agents WHERE user_agent = ANY($1)"
		rows, err := conn.Query(context.Background(), query, newUserAgents)
		if err != nil {
			return result, err
		}
		defer rows.Close()

		userAgentCacheMu.Lock()
		for rows.Next() {
			var userAgent string
			var id int
			err := rows.Scan(&userAgent, &id)
			if err != nil {
				continue
			}
			result[userAgent] = id
			// Update cache
			if len(userAgentCache) < maxCacheSize {
				userAgentCache[userAgent] = id
			}
		}
		userAgentCacheMu.Unlock()
	}

	return result, nil
}

func getUserHash(ipAddress string, userAgent string) string {
	if ipAddress == "" && userAgent == "" {
		return ""
	}

	combined := strings.TrimSpace(ipAddress) + "|" + strings.TrimSpace(userAgent)
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)[:32]
}

func logRequestHandler() gin.HandlerFunc {
	var rateLimiter = ratelimit.RateLimiter{}
	var maxInsert = getMaxInsert()

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

	return func(c *gin.Context) {
		var payload Payload
		err := c.BindJSON(&payload)
		if err != nil {
			msg := fmt.Sprintf("Invalid request data: %s", err.Error())
			log.LogErrorToFile(c.ClientIP(), "", msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		if payload.APIKey == "" {
			msg := "API key required."
			log.LogErrorToFile(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		if rateLimiter.RateLimited(payload.APIKey) {
			msg := "Too many requests."
			log.LogErrorToFile(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusTooManyRequests, gin.H{"status": http.StatusTooManyRequests, "message": msg})
			return
		}

		if len(payload.Requests) == 0 {
			msg := "Payload contains no logged requests."
			log.LogErrorToFile(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		framework, ok := frameworkID[payload.Framework]
		if !ok {
			msg := "Unsupported API framework."
			log.LogErrorToFile(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		payload.APIKey = strings.ReplaceAll(payload.APIKey, "\"", "")

		// Process and validate requests
		var validRequests []ProcessedRequest
		var userAgents []string
		uniqueUserAgents := make(map[string]bool)

		for _, request := range payload.Requests {
			if len(validRequests) >= maxInsert {
				break
			}

			// Validate method
			method, ok := methodID[request.Method]
			if !ok {
				continue
			}

			// Validate and truncate fields
			if len(request.Referrer) > 255 {
				request.Referrer = request.Referrer[:255]
			}
			if !database.ValidString(request.Referrer) {
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
			if !database.ValidUserID(request.UserID) {
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

			if request.ResponseTime < 0 {
				continue
			}

			// Process IP address based on privacy level
			var ipAddress *string
			if payload.PrivacyLevel <= P1 && request.IPAddress != "" {
				ipAddress = &request.IPAddress
			}

			// Get location and user hash
			location := getCountryCode(request.IPAddress)
			userHash := getUserHash(request.IPAddress, request.UserAgent)

			// Collect unique user agents
			if !uniqueUserAgents[request.UserAgent] {
				uniqueUserAgents[request.UserAgent] = true
				userAgents = append(userAgents, request.UserAgent)
			}

			validRequests = append(validRequests, ProcessedRequest{
				APIKey:       payload.APIKey,
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
				CreatedAt:    request.CreatedAt,
			})
		}

		if len(validRequests) == 0 {
			log.LogToFile("No valid requests to insert.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		// Get user agent IDs
		userAgentIDs, err := ensureUserAgentIDs(userAgents)
		if err != nil {
			log.LogToFile(fmt.Sprintf("Failed to ensure user agent IDs: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		// Set user agent IDs
		for i := range validRequests {
			if id, exists := userAgentIDs[payload.Requests[i].UserAgent]; exists {
				validRequests[i].UserAgentID = id
			}
		}

		// Use COPY for bulk insert
		conn, err := database.NewConnection()
		if err != nil {
			log.LogToFile(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database connection failed."})
			return
		}
		defer conn.Close(context.Background())

		_, err = conn.CopyFrom(
			context.Background(),
			pgx.Identifier{"requests"},
			[]string{"api_key", "path", "hostname", "ip_address", "user_hash", "referrer", "status", "response_time", "method", "framework", "location", "user_id", "created_at", "user_agent_id"},
			pgx.CopyFromSlice(len(validRequests), func(i int) ([]any, error) {
				req := validRequests[i]
				return []any{
					req.APIKey, req.Path, req.Hostname, req.IPAddress, req.UserHash,
					req.Referrer, req.Status, req.ResponseTime, req.Method, req.Framework,
					req.Location, req.UserID, req.CreatedAt, req.UserAgentID,
				}, nil
			}),
		)

		if err != nil {
			log.LogToFile(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Database insert failed."})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API requests logged successfully."})
		log.LogRequestsToFile(payload.APIKey, len(validRequests), len(payload.Requests))
	}
}