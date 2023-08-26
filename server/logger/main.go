package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/lib"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

func main() {
	logToFile("Starting logger...")

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.Use(cors.Default())

	rateLimiter := lib.RateLimiter{}

	app.POST("/api/log-request", logRequestHandler(rateLimiter))

	app.Run(":8000")
}

type RequestData struct {
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       int16  `json:"status"`
	ResponseTime int16  `json:"response_time"`
	CreatedAt    string `json:"created_at"`
}

type Payload struct {
	APIKey    string        `json:"api_key"`
	Requests  []RequestData `json:"requests"`
	Framework string        `json:"framework"`
}

func methodMap(method string) (int16, error) {
	switch method {
	case "GET":
		return 0, nil
	case "POST":
		return 1, nil
	case "PUT":
		return 2, nil
	case "PATCH":
		return 3, nil
	case "DELETE":
		return 4, nil
	case "OPTIONS":
		return 5, nil
	case "CONNECT":
		return 6, nil
	case "HEAD":
		return 7, nil
	case "TRACE":
		return 8, nil
	default:
		return -1, fmt.Errorf("invalid method")
	}
}

func frameworkMap(framework string) (int16, error) {
	switch framework {
	case "FastAPI":
		return 0, nil
	case "Flask":
		return 1, nil
	case "Gin":
		return 2, nil
	case "Echo":
		return 3, nil
	case "Express":
		return 4, nil
	case "Fastify":
		return 5, nil
	case "Koa":
		return 6, nil
	case "Chi":
		return 7, nil
	case "Fiber":
		return 8, nil
	case "Actix":
		return 9, nil
	case "Axum":
		return 10, nil
	case "Tornado":
		return 11, nil
	case "Django":
		return 12, nil
	case "Rails":
		return 13, nil
	case "Laravel":
		return 14, nil
	case "Sinatra":
		return 15, nil
	case "Rocket":
		return 16, nil
	default:
		return -1, fmt.Errorf("invalid framework")
	}
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
	location := record.Country.IsoCode
	return location
}

func fmtNullableString(value string) string {
	if value == "" {
		return "NULL"
	} else {
		return fmt.Sprintf("'%s'", value)
	}
}

type RequestErrors struct {
	method    int
	framework int
	userAgent int
	hostname  int
	path      int
}

func logRequest(c *gin.Context) {
	// Collect API request data sent via POST request
	var payload Payload
	if err := c.BindJSON(&payload); err != nil {
		logErrorToFile(c.ClientIP(), "", "Invalid request data.")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
		return
	} else if payload.APIKey == "" {
		logErrorToFile(c.ClientIP(), payload.APIKey, "API key required.")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
		return
	} else if len(payload.Requests) == 0 {
		logErrorToFile(c.ClientIP(), payload.APIKey, "Payload contains no logged requests.")
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Payload contains no logged requests."})
		return
	}

	var query strings.Builder
	arguments := make([]any, 0)
	query.WriteString("INSERT INTO requests (api_key, path, hostname, ip_address, user_agent, status, response_time, method, framework, location, created_at) VALUES")
	inserted := 0
	requestErrors := RequestErrors{}
	for _, request := range payload.Requests {
		// Temporary 1000 request per minute limit
		if inserted > 1000 {
			break
		}

		location := getCountryCode(request.IPAddress)

		method, err := methodMap(request.Method)
		if err != nil {
			requestErrors.method += 1
			continue
		}

		framework, err := frameworkMap(payload.Framework)
		if err != nil {
			requestErrors.framework += 1
			continue
		}

		userAgent := request.UserAgent
		if len(userAgent) > 255 {
			userAgent = userAgent[:255]
		}
		if !database.SanitizeUserAgent(userAgent) {
			requestErrors.userAgent += 1
			continue
		} else if !database.SanitizeHostname(request.Hostname) {
			requestErrors.hostname += 1
			continue
		} else if !database.SanitizePath(request.Path) {
			requestErrors.path += 1
			continue
		}

		if inserted > 0 {
			query.WriteString(",")
		}
		query.WriteString("(?,?,?,?,?,?,?,?,?,?,?)")
		arguments = append(
			arguments,
			payload.APIKey,
			request.Path,
			request.Hostname,
			request.IPAddress,
			userAgent,
			request.Status,
			request.ResponseTime,
			method,
			framework,
			location,
			request.CreatedAt)
		inserted += 1
	}

	// Record in log file for debugging purposes
	logRequestsToFile(c.ClientIP(), payload.APIKey, inserted, len(payload.Requests), requestErrors)

	// If no valid logged requests received
	if inserted == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
		return
	}

	query.WriteString(";")

	// Insert logged requests into database
	db := database.OpenDBConnection()
	_, err := db.Query(query.String(), arguments...)
	db.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
}

func logErrorToFile(ipAddress string, apiKey string, msg string) {
	text := fmt.Sprintf("%s %s :: %s", ipAddress, apiKey, msg)
	logToFile(text)
}

func logRequestsToFile(ipAddress string, apiKey string, inserted int, totalRequests int, requestErrors RequestErrors) {
	text := fmt.Sprintf("%s %s :: inserted=%d totalRequest=%d :: %d %d %d %d %d", ipAddress, apiKey, inserted, totalRequests, requestErrors.method, requestErrors.framework, requestErrors.userAgent, requestErrors.hostname, requestErrors.path)
	logToFile(text)
}

func logToFile(msg string) {
	f, err := os.OpenFile("./requests.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(msg)
}

func logRequestHandler(rateLimiter lib.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Collect API request data sent via POST request
		var payload Payload
		if err := c.BindJSON(&payload); err != nil {
			logErrorToFile(c.ClientIP(), "", "Invalid request data.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		} else if payload.APIKey == "" {
			logErrorToFile(c.ClientIP(), payload.APIKey, "API key required.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else if rateLimiter.RateLimited(payload.APIKey) {
			logErrorToFile(c.ClientIP(), payload.APIKey, "Too many requests.")
			c.String(http.StatusTooManyRequests, "Too many requests.")
			return
		} else if len(payload.Requests) == 0 {
			logErrorToFile(c.ClientIP(), payload.APIKey, "Payload contains no logged requests.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Payload contains no logged requests."})
			return
		}

		var query strings.Builder
		query.WriteString("INSERT INTO requests (api_key, path, hostname, ip_address, user_agent, status, response_time, method, framework, location, created_at) VALUES")
		arguments := make([]any, 0)
		inserted := 0
		requestErrors := RequestErrors{}
		for _, request := range payload.Requests {
			// Temporary 1000 request per minute limit
			if inserted > 1000 {
				break
			}

			location := getCountryCode(request.IPAddress)

			method, err := methodMap(request.Method)
			if err != nil {
				requestErrors.method += 1
				continue
			}

			framework, err := frameworkMap(payload.Framework)
			if err != nil {
				requestErrors.framework += 1
				continue
			}

			userAgent := request.UserAgent
			if len(userAgent) > 255 {
				userAgent = userAgent[:255]
			}
			if !database.SanitizeUserAgent(userAgent) {
				requestErrors.userAgent += 1
				continue
			} else if !database.SanitizeHostname(request.Hostname) {
				requestErrors.hostname += 1
				continue
			} else if !database.SanitizePath(request.Path) {
				requestErrors.path += 1
				continue
			}

			if inserted > 0 && inserted < 999 && inserted < len(payload.Requests)-1 {
				query.WriteString(",")
			}
			query.WriteString("(?,?,?,?,?,?,?,?,?,?,?)")
			arguments = append(
				arguments,
				payload.APIKey,
				request.Path,
				request.Hostname,
				request.IPAddress,
				userAgent,
				request.Status,
				request.ResponseTime,
				method,
				framework,
				location,
				request.CreatedAt)
			inserted += 1
		}

		// Record in log file for debugging purposes
		logRequestsToFile(c.ClientIP(), payload.APIKey, inserted, len(payload.Requests), requestErrors)

		// If no valid logged requests received
		if inserted == 0 {
			logToFile("No rows inserted.")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		query.WriteString(";")

		// Insert logged requests into database
		db := database.OpenDBConnection()
		_, err := db.Query(query.String(), arguments...)
		db.Close()
		if err != nil {
			logToFile(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
			return
		}

		// Return success response
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
	}
}
