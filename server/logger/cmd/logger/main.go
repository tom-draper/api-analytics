package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/config"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
	"github.com/tom-draper/api-analytics/server/logger/internal/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	UserAgent    string
	CreatedAt    time.Time
	UserAgentID  int
}

func main() {
	// Initialize logging first
	if err := log.Init(); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer log.Close()

	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("application crashed: %v", r))
		}
	}()

	log.Info("starting logger...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Error(fmt.Sprintf("configuration error: %v", err))
		return
	}

	// DB init
	db, err := database.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Error(fmt.Sprintf("failed to initialize database: %v", err))
		return
	}
	defer db.Close()
	log.Info("database connection pool initialized")

	// GeoIP
	geoIPDB, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Error(fmt.Sprintf("failed to open GeoIP db: %v", err))
		log.Error("location data will be unavailable")
	}
	defer func() {
		if geoIPDB != nil {
			geoIPDB.Close()
		}
	}()

	// Cache
	cache := &Cache{
		userAgentMap: make(map[string]int),
		geoIPMap:     make(map[string]*geoIPEntry),
		maxSize:      10000,
	}

	// Preload cache
	count, err := preloadUserAgentCache(context.Background(), db, cache)
	if err != nil {
		log.Error(fmt.Sprintf("failed to preload user agent cache: %v", err))
	} else {
		log.Info(fmt.Sprintf("user agent cache preloaded: %d entries", count))
	}

	// Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(cors.Default())

	handler := logRequestHandler(db, geoIPDB, cache, cfg.RateLimit, cfg.MaxInsert)

	router.POST("/api/log-request", handler)
	router.POST("/api/requests", handler)
	router.GET("/api/health", checkHealth(db))

	// HTTP server (IMPORTANT CHANGE)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server async
	serverErr := make(chan error, 1)
	go func() {
		log.Info(fmt.Sprintf("server listening on port %d", cfg.Port))
		serverErr <- srv.ListenAndServe()
	}()

	// Signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info(fmt.Sprintf("received signal: %v, shutting down...", sig))

	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("server failed: %v", err))
			return
		}
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("server forced shutdown: %v", err))
		return
	}

	log.Info("server exited")
}

func checkHealth(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		err := db.CheckConnection(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("health check failed: %v", err))
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

func logRequestHandler(db *database.DB, geoIPDB *geoip2.Reader, cache *Cache, rateLimit int, maxInsert int) gin.HandlerFunc {
	var rateLimiter = ratelimit.NewRateLimiter(rateLimit)

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
			msg := "Unsupported API framework."
			log.LogClientError(c.ClientIP(), payload.APIKey, msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

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

			// Process IP address based on privacy level
			var ipAddress *string
			if payload.PrivacyLevel <= P1 && request.IPAddress != "" {
				ipAddress = &request.IPAddress
			}

			// Get location and user hash
			location := getCountryCode(geoIPDB, cache, request.IPAddress)
			userHash := getUserHash(request.IPAddress, request.UserAgent)

			createdAt, err := time.Parse(time.RFC3339Nano, request.CreatedAt)
			if err != nil {
				createdAt = time.Now().UTC()
			}

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
				UserAgent:    request.UserAgent,
				CreatedAt:    createdAt,
			})
		}

		if len(validRequests) == 0 {
			log.Info(fmt.Sprintf("key=%s: no valid requests to insert (received %d)", payload.APIKey, len(payload.Requests)))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		ctx := c.Request.Context()

		// Get user agent IDs
		userAgentIDs, err := ensureUserAgentIDs(ctx, db, cache, userAgents)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to ensure user agent IDs: %v", payload.APIKey, err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		// Set user agent IDs
		for i := range validRequests {
			if id, exists := userAgentIDs[validRequests[i].UserAgent]; exists {
				validRequests[i].UserAgentID = id
			}
		}

		// Use COPY for bulk insert
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
					req.APIKey, req.Path, req.Hostname, req.IPAddress, req.UserHash,
					req.Referrer, req.Status, req.ResponseTime, req.Method, req.Framework,
					req.Location, req.UserID, req.CreatedAt, req.UserAgentID,
				}, nil
			}),
		)

		if err != nil {
			log.Error(fmt.Sprintf(
				"COPY FAILED: api_key=%s request_count=%d error=%v",
				payload.APIKey,
				len(validRequests),
				err,
			))

			// Classify likely COPY vs DB failure
			if pgErr, ok := err.(*pgconn.PgError); ok {
				log.Error(fmt.Sprintf(
					"POSTGRES ERROR DETAIL: code=%s message=%s detail=%s hint=%s",
					pgErr.Code,
					pgErr.Message,
					pgErr.Detail,
					pgErr.Hint,
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

// Preload user agent cache at startup
func preloadUserAgentCache(ctx context.Context, db *database.DB, cache *Cache) (int, error) {
	rows, err := db.Pool.Query(ctx, "SELECT user_agent, id FROM user_agents LIMIT 50000")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cache.userAgentMu.Lock()
	defer cache.userAgentMu.Unlock()

	for rows.Next() {
		var userAgent string
		var id int
		if err := rows.Scan(&userAgent, &id); err != nil {
			continue
		}
		cache.userAgentMap[userAgent] = id
	}

	return len(cache.userAgentMap), rows.Err()
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
				log.Error(fmt.Sprintf("failed to insert user agent: %v", err))
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

		if err := rows.Err(); err != nil {
			return result, err
		}
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
	cache.geoIPMu.Lock()
	if entry, exists := cache.geoIPMap[ipAddress]; exists {
		entry.lastAccess = now
		countryCode := entry.countryCode
		cache.geoIPMu.Unlock()
		return countryCode
	}
	cache.geoIPMu.Unlock()

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
		removeCount := max(len(entries)/4, 1)

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
