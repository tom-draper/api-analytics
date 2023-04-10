package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"net/http"
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

	r := app.Group("/api")
	r.Use(cors.Default())
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 2,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	r.Use(mw)

	r.POST("/log-request", logRequestHandler(db))

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
		return -1, fmt.Errorf("error: invalid method")
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

func getCountryCode(IPAddress string) (string, error) {
	var location string
	if IPAddress != "" {
		db, err := geoip2.Open("GeoLite2-Country.mmdb")
		if err != nil {
			return location, err
		}
		defer db.Close()

		ip := net.ParseIP(IPAddress)
		record, err := db.Country(ip)
		if err != nil {
			return location, err
		}
		location = record.Country.IsoCode
	}
	return location, nil
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
			var query bytes.Buffer
			query.WriteString("INSERT INTO requests (api_key, path, hostname, ip_address, user_agent, status, response_time, method, framework, location, created_at) VALUES")
			for i, request := range payload.Requests {
				if i > 1000 {
					break
				}

				location, _ := getCountryCode(request.IPAddress)

				method, err := methodMap(request.Method)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid method."})
					return
				}

				framework, err := frameworkMap(payload.Framework)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid framework."})
					return
				}

				userAgent := request.UserAgent
				if len(userAgent) > 255 {
					userAgent = userAgent[:255]
				}

				fmtHostname := fmtNullableString(request.Hostname)
				fmtIPAddress := fmtNullableString(request.IPAddress)
				fmtUserAgent := fmtNullableString(userAgent)
				fmtLocation := fmtNullableString(location)

				if i > 0 {
					query.WriteString(",")
				}
				query.WriteString(
					fmt.Sprintf("('%s', '%s', %s, %s, %s, %d, %d, %d, %d, %s, '%s')",
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
