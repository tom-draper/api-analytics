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

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/lib/log"
	"github.com/tom-draper/api-analytics/server/logger/lib/ratelimit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/oschwald/geoip2-golang"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.LogToFile(fmt.Sprintf("Application crashed: %v", err))
		}
	}()

	log.LogToFile("Starting logger...")

	err := database.LoadConfig()
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
	P1 PrivacyLevel = iota // Client IP address stored, location inferred from IP and stored
	P2                     // Location inferred from IP and stored, client IP address discarded
	P3                     // Client IP address never be sent to server, optional custom user ID field is the only user identification
)

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

func getCountryCode(IPAddress string) string {
	if IPAddress == "" {
		return ""
	}
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		return ""
	}
	defer db.Close()

	ip := net.ParseIP(IPAddress)
	if ip == nil {
		return ""
	}

	record, err := db.Country(ip)
	if err != nil {
		return ""
	}

	return record.Country.IsoCode
}

func storeNewUserAgents(userAgents map[string]struct{}) error {
	var query strings.Builder
	query.WriteString("INSERT INTO user_agents (user_agent) VALUES ")
	arguments := make([]any, len(userAgents))
	i := 0

	for userAgent := range userAgents {
		query.WriteString(fmt.Sprintf("($%d)", i+1))
		if i < len(userAgents)-1 {
			query.WriteString(",")
		}
		arguments[i] = userAgent
		i++
	}

	query.WriteString(" ON CONFLICT (user_agent) DO NOTHING;")

	conn, err := database.NewConnection()
	if err != nil {
		log.LogToFile(err.Error())
		return err
	}
	_, err = conn.Exec(context.Background(), query.String(), arguments...)
	conn.Close(context.Background())
	if err != nil {
		log.LogToFile(err.Error())
		return err
	}

	return nil
}

func getUserAgentIDs(userAgents map[string]struct{}) (map[string]int, error) {
	if len(userAgents) == 0 {
		return make(map[string]int), nil
	}

	var query strings.Builder
	query.WriteString("SELECT user_agent, id FROM user_agents WHERE user_agent IN (")
	arguments := make([]any, len(userAgents))
	i := 0

	for userAgent := range userAgents {
		query.WriteString(fmt.Sprintf("$%d", i+1))
		if i < len(userAgents)-1 {
			query.WriteString(",")
		}
		arguments[i] = userAgent
		i++
	}
	query.WriteString(");")

	ids := make(map[string]int)
	conn, err := database.NewConnection()
	if err != nil {
		log.LogToFile(err.Error())
		return ids, err
	}
	rows, err := conn.Query(context.Background(), query.String(), arguments...)
	conn.Close(context.Background())
	if err != nil {
		log.LogToFile(err.Error())
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var userAgent string
		var id int
		err := rows.Scan(&userAgent, &id)
		if err != nil {
			log.LogToFile(err.Error())
			continue
		}
		ids[userAgent] = id
	}

	if err := rows.Err(); err != nil {
		log.LogToFile(err.Error())
		return ids, err // Return error if rows iteration failed
	}

	return ids, nil
}

// CreateVisitorID generates a hash from an IP address and user agent string
func getUserHash(ipAddress string, userAgent string) string {
	if ipAddress == "" && userAgent == "" {
		return ""
	}

	// Combine the IP address and user agent with a separator
	combined := strings.TrimSpace(ipAddress) + "|" + strings.TrimSpace(userAgent)

	// Create SHA-256 hash
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hashBytes := hasher.Sum(nil)

	// Convert to hex string
	hashString := hex.EncodeToString(hashBytes)

	// Return the full hash or a shorter version if needed
	return hashString[:32] // Return first 32 chars (16 bytes) of the hash
}

func logRequestHandler() gin.HandlerFunc {
	var rateLimiter = ratelimit.RateLimiter{}

	var maxInsert = getMaxInsert()

	var methodID = map[string]int16{
		"GET":     0,
		"POST":    1,
		"PUT":     2,
		"PATCH":   3,
		"DELETE":  4,
		"OPTIONS": 5,
		"CONNECT": 6,
		"HEAD":    7,
		"TRACE":   8,
	}

	var frameworkID = map[string]int16{
		"FastAPI":      0,
		"Flask":        1,
		"Gin":          2,
		"Echo":         3,
		"Express":      4,
		"Fastify":      5,
		"Koa":          6,
		"Chi":          7,
		"Fiber":        8,
		"Actix":        9,
		"Axum":         10,
		"Tornado":      11,
		"Django":       12,
		"Rails":        13,
		"Laravel":      14,
		"Sinatra":      15,
		"Rocket":       16,
		"ASP.NET Core": 17,
		"Hono":         18,
	}

	return func(c *gin.Context) {
		var payload Payload
		err := c.BindJSON(&payload)
		if err != nil {
			msg := fmt.Sprintf("Invalid request data.\n%s\nRequest body: %s", err.Error(), "body")
			log.LogErrorToFile(c.ClientIP(), "", msg)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": msg})
			return
		}

		if payload.APIKey == "" {
			msg := "API key requied."
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

		var query strings.Builder
		query.WriteString("INSERT INTO requests (api_key, path, hostname, ip_address, user_hash, status, response_time, method, framework, location, user_id, created_at, user_agent_id) VALUES ")
		arguments := make([]any, 0)
		inserted := 0
		userAgents := make([]string, 0)
		uniqueUserAgents := map[string]struct{}{}
		for _, request := range payload.Requests {
			if inserted >= maxInsert {
				break
			}

			location := getCountryCode(request.IPAddress)

			userHash := getUserHash(request.IPAddress, request.UserAgent)

			if payload.PrivacyLevel > P1 {
				// Client IP address discarded for privacy level P2 and P3
				request.IPAddress = ""
			}
			var ipAddress any
			if request.IPAddress == "" {
				ipAddress = nil
			} else {
				ipAddress = request.IPAddress
			}

			method, ok := methodID[request.Method]
			if !ok {
				continue
			}

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

			// If not at final row in query, separate with comma
			if inserted > 0 && inserted < maxInsert && inserted < len(payload.Requests) {
				query.WriteString(",")
			}

			// Register user agent to be stored in the database
			if _, ok := uniqueUserAgents[request.UserAgent]; !ok {
				uniqueUserAgents[request.UserAgent] = struct{}{}
			}
			// Temp store for user agents in each row for conversion to user agent IDs
			userAgents = append(userAgents, request.UserAgent)

			numArgs := len(arguments)
			query.WriteString(
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
					numArgs+1, // api_key
					numArgs+2, // path
					numArgs+3, // hostname
					numArgs+4, // ip_address
					numArgs+5, // user_hash
					numArgs+6, // referrer
					numArgs+7, // status
					numArgs+8, // response_time
					numArgs+9, // method
					numArgs+10, // framework
					numArgs+11, // location
					numArgs+12, // user_id
					numArgs+13, // created_at
					numArgs+14, // user_agent_id
				),
			)
			arguments = append(
				arguments,
				payload.APIKey,
				request.Path,
				request.Hostname,
				ipAddress,
				userHash,
				request.Referrer,
				request.Status,
				request.ResponseTime,
				method,
				framework,
				location,
				request.UserID,
				request.CreatedAt,
				0, // Placeholder for user_agent_id
			)
			inserted += 1
		}

		// If no valid logged requests received
		if inserted == 0 {
			log.LogToFile("No rows inserted.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		query.WriteString(";")

		// Store any new user agents found
		_ = storeNewUserAgents(uniqueUserAgents)
		// Get associated user IDs for user agents
		userAgentIDs, _ := getUserAgentIDs(uniqueUserAgents)
		// Insert user agent IDs into arguments
		for i, userAgent := range userAgents {
			if id, ok := userAgentIDs[userAgent]; ok {
				arguments[(i*14)+13] = id
			}
		}

		// Insert logged requests into database
		conn, err := database.NewConnection()
		if err != nil {
			log.LogToFile(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database connection failed."})
			return
		}
		_, err = conn.Exec(context.Background(), query.String(), arguments...)
		if err != nil {
			log.LogToFile(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
			return
		}

		// Return success response
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API requests logged successfully."})

		// Record in log file for debugging
		log.LogRequestsToFile(payload.APIKey, inserted, len(payload.Requests))
	}
}
