package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"

	"github.com/tom-draper/api-analytics/server/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
)

func main() {
	db := database.OpenDBConnection()

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.Use(cors.Default())

	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 5,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	app.Use(mw)

	app.POST("/api/log-request", logRequestHandler(db))

	app.Run(":8000")
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
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
	default:
		return -1, fmt.Errorf("error: invalid framework")
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

func logRequestHandler(db *sql.DB) gin.HandlerFunc {
	logRequest := func(c *gin.Context) {
		// Collect API request data sent via POST request
		var payload Payload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
			return
		}

		if len(payload.Requests) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Payload empty."})
			return
		} else if payload.APIKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		} else {
			var query strings.Builder
			query.WriteString("INSERT INTO requests (api_key, path, hostname, ip_address, user_agent, status, response_time, method, framework, location, created_at) VALUES")
			inserted := 0
			for _, request := range payload.Requests {
				if inserted > 1000 {
					break
				}

				location := getCountryCode(request.IPAddress)

				method, err := methodMap(request.Method)
				if err != nil {
					continue
				}

				framework, err := frameworkMap(payload.Framework)
				if err != nil {
					continue
				}

				userAgent := request.UserAgent
				if len(userAgent) > 255 {
					userAgent = userAgent[:255]
				}
				if !database.SanitizeUserAgent(userAgent) {
					continue
				}

				if !database.SanitizeHostname(request.Hostname) {
					continue
				}

				if !database.SanitizePath(request.Path) {
					continue
				}

				// Convert to NULL or '<value>' for SQL query
				fmtHostname := fmtNullableString(request.Hostname)
				fmtIPAddress := fmtNullableString(request.IPAddress)
				fmtUserAgent := fmtNullableString(userAgent)
				fmtLocation := fmtNullableString(location)

				if inserted > 0 {
					query.WriteString(",")
				}
				query.WriteString(
					fmt.Sprintf("('%s','%s',%s,%s,%s,%d,%d,%d,%d,%s,'%s')",
						payload.APIKey,
						request.Path,
						fmtHostname,
						fmtIPAddress,
						fmtUserAgent,
						request.Status,
						request.ResponseTime,
						method,
						framework,
						fmtLocation,
						request.CreatedAt,
					),
				)
				inserted += 1
			}

			if inserted == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request data."})
				return
			}

			query.WriteString(";")

			// Insert logged requests into database
			_, err := db.Query(query.String())
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid data."})
				return
			}

			// Return success response
			c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "API request logged successfully."})
		}
	}

	return gin.HandlerFunc(logRequest)
}
