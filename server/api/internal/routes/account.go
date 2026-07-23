package routes

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
	"github.com/tom-draper/api-analytics/server/database"
)

func genAPIKey(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		apiKey, err := db.CreateUser(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("api key generation failed - %s", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "API key generation failed."})
			return
		}

		log.Info(fmt.Sprintf("key=%s: API key generation successful", apiKey))
		c.JSON(http.StatusOK, apiKey)
	}
}

func getUserID(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := requireAPIKeyParam(c)
		if !ok {
			return
		}

		ctx := c.Request.Context()

		userID, err := db.GetUserID(ctx, apiKey)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "API key not found."})
			} else {
				log.Error(fmt.Sprintf("key=%s: user ID fetch failed - %s", apiKey, err.Error()))
				c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			}
			return
		}

		c.JSON(http.StatusOK, userID)
	}
}

// deleteAccount permanently deletes an account and every request, monitor and
// ping belonging to it. The API key is taken from the request body rather than
// the URL so that it stays out of access logs, proxy logs and browser history,
// and so that fetching a URL cannot delete an account.
func deleteAccount(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			APIKey string `json:"api_key"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			log.Info(fmt.Sprintf("invalid account deletion request - %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		// Users sometimes provide the key in quotes.
		apiKey := strings.TrimSpace(strings.ReplaceAll(body.APIKey, "\"", ""))
		if apiKey == "" {
			log.Info("api key missing")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		}
		if !database.ValidAPIKey(apiKey) {
			log.Info("api key invalid format")
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key format. Expected UUID format."})
			return
		}

		deleteAccountByAPIKey(c, db, apiKey)
	}
}

func deleteAccountByAPIKey(c *gin.Context, db *database.DB, apiKey string) {
	ctx := c.Request.Context()

	if err := db.DeleteUserAccount(ctx, apiKey); err != nil {
		log.Error(fmt.Sprintf("key=%s: data deletion failed - %s", apiKey, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
		return
	}

	log.Info(fmt.Sprintf("key=%s: account data deleted successfully", apiKey))
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
}

// regenerateUserIDFromBody regenerates a user's dashboard ID, taking the API key
// from the POST body so it stays out of URLs and logs and needs no custom
// header. Sending no header keeps this a CORS "simple request", avoiding a
// preflight the API does not answer — the same approach as deleteAccount.
func regenerateUserIDFromBody(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			APIKey string `json:"api_key"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			log.Info(fmt.Sprintf("invalid user ID regeneration request - %s", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid request body."})
			return
		}

		// Users sometimes provide the key in quotes.
		apiKey := strings.TrimSpace(strings.ReplaceAll(body.APIKey, "\"", ""))
		if apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
			return
		}
		if !database.ValidAPIKey(apiKey) {
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key format. Expected UUID format."})
			return
		}

		regenerateUserIDByAPIKey(c, db, apiKey)
	}
}

// regenerateUserID is the deprecated GET form, kept for existing clients. Prefer
// POST /regenerate-user-id, which keeps the key out of the URL.
func regenerateUserID(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := requireAPIKeyParam(c)
		if !ok {
			return
		}
		regenerateUserIDByAPIKey(c, db, apiKey)
	}
}

func regenerateUserIDByAPIKey(c *gin.Context, db *database.DB, apiKey string) {
	ctx := c.Request.Context()

	userID, err := db.RegenerateUserID(ctx, apiKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "API key not found."})
		} else {
			log.Error(fmt.Sprintf("key=%s: user ID regeneration failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
		}
		return
	}

	log.Info(fmt.Sprintf("key=%s: user ID regenerated successfully", apiKey))
	c.JSON(http.StatusOK, userID)
}

// healthTTL bounds how often the health check actually pings the database, so a
// flood of /health requests cannot contend with real traffic for the pool.
const healthTTL = time.Second

func checkHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	var (
		mu        sync.Mutex
		lastCheck time.Time
		lastErr   error
		checked   bool
	)

	return func(c *gin.Context) {
		uptime := int(time.Since(startTime).Seconds())

		mu.Lock()
		if !checked || time.Since(lastCheck) > healthTTL {
			lastErr = db.CheckConnection(c.Request.Context())
			lastCheck = time.Now()
			checked = true
		}
		err := lastErr
		mu.Unlock()

		if err != nil {
			log.Error(fmt.Sprintf("health check failed: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"health":         "unhealthy",
				"uptime_seconds": uptime,
				"database":       "unreachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"health":         "healthy",
			"uptime_seconds": uptime,
			"database":       "connected",
		})
	}
}
