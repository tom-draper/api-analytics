package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/oschwald/geoip2-golang"
)

const (
	P1 PrivacyLevel = iota
	P2
	P3
)

type PrivacyLevel int

// Cache holds shared caches with their mutexes
// These need to be shared across all requests for efficient lookups
type Cache struct {
	userAgentMap map[string]int
	userAgentMu  sync.RWMutex
	geoIPMap     map[string]*geoIPEntry
	geoIPMu      sync.RWMutex
	maxSize      int
}

type geoIPEntry struct {
	countryCode string
	lastAccess  int64 // Unix timestamp for LRU eviction
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

// Processed request for batch insertion
type ProcessedRequest struct {
	APIKey       string
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
	CreatedAt    string
	UserAgentID  int
}

func main() {
	// Initialize logging first
	if err := log.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Close()

	defer func() {
		if err := recover(); err != nil {
			log.LogToFile(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.LogToFile("Starting logger...")

	// Load and validate configuration
	cfg, err := config.LoadAndValidate()
	if err != nil {
		log.LogToFile(fmt.Sprintf("Configuration error: %v", err))
		return
	}

	// Initialize database connection pool
	db, err := database.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to create database connection pool: %v", err))
		return
	}
	defer db.Close()
	log.LogToFile("Database connection pool initialized")

	// Initialize GeoIP database once
	geoIPDB, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to open GeoIP database: %v", err))
	}
	defer func() {
		if geoIPDB != nil {
			geoIPDB.Close()
		}
	}()

	// Create shared cache
	cache := &Cache{
		userAgentMap: make(map[string]int),
		geoIPMap:     make(map[string]*geoIPEntry),
		maxSize:      10000,
	}

	// Preload user agent cache
	err = preloadUserAgentCache(context.Background(), db, cache)
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to preload user agent cache: %v", err))
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(cors.Default())

	// Pass dependencies to handler factories
	handler := logRequestHandler(db, geoIPDB, cache, cfg.MaxInsert)
	router.POST("/api/log-request", handler)
	router.POST("/api/requests", handler) // Preferred
	router.GET("/api/health", checkHealth(db))

	if err := router.Run(":8000"); err != nil {
		log.LogToFile(fmt.Sprintf("Failed to run server: %v", err))
	}
}

func checkHealth(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		err := db.CheckConnection(ctx)
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
}

func logRequestHandler(db *database.DB, geoIPDB *geoip2.Reader, cache *Cache, maxInsert int) gin.HandlerFunc {
	var rateLimiter = ratelimit.NewRateLimiter()

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
			location := getCountryCode(geoIPDB, cache, request.IPAddress)
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

		ctx := c.Request.Context()

		// Get user agent IDs
		userAgentIDs, err := ensureUserAgentIDs(ctx, db, cache, userAgents)
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
		_, err = db.Pool.CopyFrom(
			ctx,
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

// Preload user agent cache at startup
func preloadUserAgentCache(ctx context.Context, db *database.DB, cache *Cache) error {
	rows, err := db.Pool.Query(ctx, "SELECT user_agent, id FROM user_agents LIMIT 50000")
	if err != nil {
		return err
	}
	defer rows.Close()

	cache.userAgentMu.Lock()
	defer cache.userAgentMu.Unlock()

	for rows.Next() {
		var userAgent string
		var id int
		err := rows.Scan(&userAgent, &id)
		if err != nil {
			continue
		}
		cache.userAgentMap[userAgent] = id
	}

	return rows.Err()
}

// Efficient batch user agent insertion and ID retrieval
func ensureUserAgentIDs(ctx context.Context, db *database.DB, cache *Cache, userAgents []string) (map[string]int, error) {
	if len(userAgents) == 0 {
		return make(map[string]int), nil
	}

	result := make(map[string]int)
	newUserAgents := make([]string, 0)

	// Check cache first
	cache.userAgentMu.RLock()
	for _, ua := range userAgents {
		if id, exists := cache.userAgentMap[ua]; exists {
			result[ua] = id
		} else {
			newUserAgents = append(newUserAgents, ua)
		}
	}
	cache.userAgentMu.RUnlock()

	if len(newUserAgents) == 0 {
		return result, nil
	}

	// Use COPY for bulk insert of new user agents
	// First, try to insert new user agents using COPY
	_, err := db.Pool.CopyFrom(
		ctx,
		pgx.Identifier{"user_agents"},
		[]string{"user_agent"},
		pgx.CopyFromSlice(len(newUserAgents), func(i int) ([]any, error) {
			return []any{newUserAgents[i]}, nil
		}),
	)
	if err != nil {
		// Fallback to individual inserts with ON CONFLICT
		for _, ua := range newUserAgents {
			_, err := db.Pool.Exec(ctx,
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
		rows, err := db.Pool.Query(ctx, query, newUserAgents)
		if err != nil {
			return result, err
		}
		defer rows.Close()

		cache.userAgentMu.Lock()
		for rows.Next() {
			var userAgent string
			var id int
			err := rows.Scan(&userAgent, &id)
			if err != nil {
				continue
			}
			result[userAgent] = id
			// Update cache
			if len(cache.userAgentMap) < cache.maxSize {
				cache.userAgentMap[userAgent] = id
			}
		}
		cache.userAgentMu.Unlock()
	}

	return result, nil
}

// Cached country code lookup with LRU eviction
func getCountryCode(geoIPDB *geoip2.Reader, cache *Cache, ipAddress string) string {
	if ipAddress == "" || geoIPDB == nil {
		return ""
	}

	now := time.Now().Unix()

	// Check cache first (read lock)
	cache.geoIPMu.RLock()
	if entry, exists := cache.geoIPMap[ipAddress]; exists {
		countryCode := entry.countryCode
		cache.geoIPMu.RUnlock()

		// Update access time (write lock)
		cache.geoIPMu.Lock()
		entry.lastAccess = now
		cache.geoIPMu.Unlock()

		return countryCode
	}
	cache.geoIPMu.RUnlock()

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ""
	}

	record, err := geoIPDB.Country(ip)
	if err != nil {
		return ""
	}

	countryCode := record.Country.IsoCode

	// Cache the result with LRU eviction
	cache.geoIPMu.Lock()
	if len(cache.geoIPMap) >= cache.maxSize {
		// LRU eviction: remove least recently used entries
		evictLRUEntries(cache)
	}
	cache.geoIPMap[ipAddress] = &geoIPEntry{
		countryCode: countryCode,
		lastAccess:  now,
	}
	cache.geoIPMu.Unlock()

	return countryCode
}

// evictLRUEntries removes the least recently used entries from the GeoIP cache
// Caller must hold cache.geoIPMu write lock
func evictLRUEntries(cache *Cache) {
	// Find entries older than 1 hour
	cutoff := time.Now().Unix() - 3600
	var toDelete []string

	for ip, entry := range cache.geoIPMap {
		if entry.lastAccess < cutoff {
			toDelete = append(toDelete, ip)
		}
	}

	// If we found old entries, delete them
	if len(toDelete) > 0 {
		for _, ip := range toDelete {
			delete(cache.geoIPMap, ip)
		}
	}

	// If still over capacity after removing old entries, remove oldest entries
	if len(cache.geoIPMap) >= cache.maxSize {
		// Build slice of entries with their IPs for sorting
		type entry struct {
			ip         string
			lastAccess int64
		}
		entries := make([]entry, 0, len(cache.geoIPMap))
		for ip, e := range cache.geoIPMap {
			entries = append(entries, entry{ip: ip, lastAccess: e.lastAccess})
		}

		// Sort by access time (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].lastAccess < entries[j].lastAccess
		})

		// Remove oldest 25% of entries
		removeCount := len(entries) / 4
		if removeCount < 1 {
			removeCount = 1
		}

		for i := 0; i < removeCount && i < len(entries); i++ {
			delete(cache.geoIPMap, entries[i].ip)
		}
	}
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
