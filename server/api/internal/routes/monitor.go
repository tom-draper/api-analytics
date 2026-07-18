package routes

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
)

// validMonitorURL rejects a monitored URL that is empty, longer than its
// varchar(255) column, or contains control characters. Control characters (in
// particular CR/LF) are rejected so a URL cannot inject headers into the
// downtime alert emails that embed it.
func validMonitorURL(url string) bool {
	if url == "" || len(url) > 255 {
		return false
	}
	for _, r := range url {
		if r == 0 || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// internalMonitorTarget reports whether rawURL points at an address that must
// not be monitored, giving the user immediate feedback instead of a monitor
// that only ever reports "down". It rejects localhost and literal non-public
// IPs (loopback, private, link-local metadata, etc). Hostnames are not resolved
// here — that would add DNS latency to the request path — so a hostname that
// resolves to an internal address is caught authoritatively by the pinger's
// dialer guard, which re-checks the resolved IP at connection time.
func internalMonitorTarget(rawURL string) bool {
	// Ensure a scheme so url.Parse populates Host rather than Path.
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && !database.IsPublicIP(ip) {
		return true
	}
	return false
}

type MonitorRow struct {
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

type Monitor struct {
	UserID string `json:"user_id"`
	URL    string `json:"url"`
	Secure bool   `json:"secure"`
	Ping   bool   `json:"ping"`
}

type MonitorPing struct {
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func getUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		ctx := c.Request.Context()

		query := "SELECT url, secure, ping, monitor.created_at FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;"
		rows, err := db.Pool.Query(ctx, query, userID)
		if err != nil {
			log.Error(fmt.Sprintf("id=%s: failed to fetch monitors - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Failed to fetch monitors."})
			return
		}
		defer rows.Close()

		monitors := make([]MonitorRow, 0)
		for rows.Next() {
			var monitor MonitorRow
			if err := rows.Scan(&monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt); err == nil {
				monitors = append(monitors, monitor)
			}
		}
		if err := rows.Err(); err != nil {
			log.Error(fmt.Sprintf("id=%s: monitor scan error - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		c.JSON(http.StatusOK, monitors)
	}
}

func addUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var monitor Monitor
		if err := c.ShouldBindJSON(&monitor); err != nil {
			log.Info("invalid monitor to add")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if monitor.UserID == "" {
			log.Info("user ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		}

		if !validMonitorURL(monitor.URL) {
			log.Info("invalid monitor URL")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid monitor URL."})
			return
		}

		if internalMonitorTarget(monitor.URL) {
			log.Info("internal monitor URL rejected")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Monitor URL must point to a public address."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: add monitor", monitor.UserID))

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, monitor.UserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "User ID not found."})
			} else {
				log.Error(fmt.Sprintf("id=%s: monitor user ID lookup failed - %s", monitor.UserID, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			}
			return
		}

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to start transaction - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}
		defer tx.Rollback(ctx)

		var monitorCount int
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM monitor WHERE api_key = $1 FOR UPDATE;", apiKey).Scan(&monitorCount); err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to get monitor count - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		if monitorCount >= 3 {
			log.Info(fmt.Sprintf("key=%s: monitor limit reached [%d]", apiKey, monitorCount))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "monitor limit reached."})
			return
		}

		var insertedURL string
		err = tx.QueryRow(ctx,
			"INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT (api_key, url) DO NOTHING RETURNING url",
			apiKey, monitor.URL, monitor.Secure, monitor.Ping,
		).Scan(&insertedURL)
		if err == pgx.ErrNoRows {
			log.Info(fmt.Sprintf("key=%s: monitor already exists", apiKey))
			c.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "message": "monitor already exists."})
			return
		}
		if err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to create new monitor - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		if err = tx.Commit(ctx); err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to commit transaction - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: monitor '%s' created successfully", apiKey, monitor.URL))
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "New monitor created successfully."})
	}
}

func deleteUserMonitor(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			UserID string `json:"user_id"`
			URL    string `json:"url"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			log.Info(fmt.Sprintf("invalid monitor to delete - %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		if body.UserID == "" {
			log.Info("user ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "User ID required."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: delete monitor", body.UserID))

		ctx := c.Request.Context()

		apiKey, err := db.GetAPIKey(ctx, body.UserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "User ID not found."})
			} else {
				log.Error(fmt.Sprintf("id=%s: monitor user ID lookup failed - %s", body.UserID, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			}
			return
		}

		if err = db.DeleteURLMonitor(ctx, apiKey, body.URL); err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to delete monitor - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Internal server error."})
			return
		}

		if err = db.DeleteURLPings(ctx, apiKey, body.URL); err != nil {
			log.Error(fmt.Sprintf("key=%s: failed to delete pings - %s", apiKey, err.Error()))
			// Continue even if ping deletion fails
		}

		log.Info(fmt.Sprintf("key=%s: monitor '%s' deleted successfully", apiKey, body.URL))
		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Monitor deleted successfully."})
	}
}

func getUserPings(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			log.Info("user ID empty")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid user ID."})
			return
		}

		log.Info(fmt.Sprintf("id=%s: monitor access", userID))

		ctx := c.Request.Context()

		rows, err := db.Pool.Query(ctx,
			"SELECT url FROM monitor INNER JOIN users ON users.api_key = monitor.api_key WHERE users.user_id = $1;",
			userID,
		)
		if err != nil {
			log.Error(fmt.Sprintf("id=%s: monitor access failed - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		monitors := make(map[string][]MonitorPing)
		for rows.Next() {
			var url string
			if err := rows.Scan(&url); err == nil {
				monitors[url] = make([]MonitorPing, 0)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			log.Error(fmt.Sprintf("id=%s: monitor scan error - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		rows.Close()

		rows, err = db.Pool.Query(ctx,
			"SELECT url, response_time, status, pings.created_at FROM pings INNER JOIN users ON users.api_key = pings.api_key WHERE users.user_id = $1;",
			userID,
		)
		if err != nil {
			log.Error(fmt.Sprintf("id=%s: ping access failed - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var url string
			var ping MonitorPing
			if err := rows.Scan(&url, &ping.ResponseTime, &ping.Status, &ping.CreatedAt); err == nil {
				if val, ok := monitors[url]; ok {
					monitors[url] = append(val, ping)
				}
			}
		}
		if err := rows.Err(); err != nil {
			log.Error(fmt.Sprintf("id=%s: ping scan error - %s", userID, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		if err = db.UpdateLastAccessedByUserID(ctx, userID); err != nil {
			log.Error(fmt.Sprintf("id=%s: user last access update failed - %s", userID, err.Error()))
		}

		log.Info(fmt.Sprintf("id=%s: monitor access successful [%d]", userID, len(monitors)))
		c.JSON(http.StatusOK, monitors)
	}
}
