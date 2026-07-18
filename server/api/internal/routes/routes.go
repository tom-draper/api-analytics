package routes

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tom-draper/api-analytics/server/api/internal/config"
	"github.com/tom-draper/api-analytics/server/database"
)

func RegisterRouter(r *gin.RouterGroup, db *database.DB, cfg *config.Config, startTime time.Time) {
	r.GET("/generate", genAPIKey(db))
	r.GET("/generate-api-key", genAPIKey(db))
	r.GET("/user-id/:apiKey", getUserID(db))
	r.GET("/reset-user-id/:apiKey", regenerateUserID(db))
	r.GET("/requests/:userID", getRequestsHandler(db, cfg))
	r.GET("/requests/:userID/:page", getPaginatedRequestsHandler(db, cfg))
	r.POST("/delete", deleteAccount(db))
	r.GET("/monitor/:userID", getUserMonitor(db))
	r.GET("/monitor/pings/:userID", getUserPings(db))
	r.POST("/monitor/add", addUserMonitor(db))
	r.POST("/monitor/delete", deleteUserMonitor(db))
	r.GET("/data", getData(db))
	r.GET("/health", checkHealth(db, startTime))
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

func compressJSON(data any) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	gzw := gzip.NewWriter(&buffer)
	if _, err = gzw.Write(body); err != nil {
		return nil, err
	}
	if err = gzw.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
