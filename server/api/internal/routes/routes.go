package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/server/api/internal/config"
	"github.com/tom-draper/api-analytics/server/database"
)

// maxJSONBody caps request bodies on the POST routes. Every accepted body holds
// only an API key/user ID/URL, so 64 KiB is ample while stopping an unbounded
// read into memory.
const maxJSONBody = 64 * 1024

func RegisterRouter(r *gin.RouterGroup, db *database.DB, cfg *config.Config, startTime time.Time) {
	body := bodyLimit(maxJSONBody)

	r.GET("/generate", genAPIKey(db))
	r.GET("/generate-api-key", genAPIKey(db))
	r.POST("/user-id", body, getUserIDFromBody(db))
	// Deprecated: the API key in the URL can leak through access logs. New
	// callers should use POST /user-id with {"api_key": "..."} instead.
	r.GET("/user-id/:apiKey", getUserID(db))
	r.POST("/regenerate-user-id", body, regenerateUserIDFromBody(db))
	// Deprecated: the key in the URL leaks into logs. Use POST /regenerate-user-id.
	r.GET("/reset-user-id/:apiKey", regenerateUserID(db))
	r.GET("/requests/:userID", getRequestsHandler(db, cfg))
	r.GET("/requests/:userID/:page", getPaginatedRequestsHandler(db, cfg))
	r.POST("/delete", body, deleteAccount(db))
	r.GET("/monitor/:userID", getUserMonitor(db))
	r.GET("/monitor/pings/:userID", getUserPings(db))
	r.POST("/monitor/add", body, addUserMonitor(db))
	r.POST("/monitor/delete", body, deleteUserMonitor(db))
	r.GET("/data", getData(db))
	r.GET("/health", checkHealth(db, startTime))
}

// bodyLimit caps the request body so an oversized POST cannot be read into
// memory; once the limit is exceeded the bound reader errors and the handler's
// JSON bind fails, returning 400.
func bodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// requireAPIKeyParam extracts and validates the :apiKey route param.
// Returns the key and true on success, or writes a 400 response and returns false.
func requireAPIKeyParam(c *gin.Context) (string, bool) {
	apiKey := strings.TrimSpace(c.Param("apiKey"))
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "API key required."})
		return "", false
	}
	if !database.ValidAPIKey(apiKey) {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Invalid API key format. Expected UUID format."})
		return "", false
	}
	return apiKey, true
}

func getNullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func getAPIKeyFromHeader(c *gin.Context) string {
	apiKey := c.GetHeader("X-AUTH-TOKEN")
	if apiKey == "" {
		// Check old (deprecated) identifier
		apiKey = c.GetHeader("API-Key")
		if apiKey == "" {
			return ""
		}
	}
	// Clean up API key (users sometimes provide it in quotes)
	return strings.TrimSpace(strings.ReplaceAll(apiKey, "\"", ""))
}
