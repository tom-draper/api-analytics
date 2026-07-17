package routes

import (
	"fmt"
	"net/http"
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

func deleteData(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := requireAPIKeyParam(c)
		if !ok {
			return
		}

		ctx := c.Request.Context()

		if err := db.DeleteUserAccount(ctx, apiKey); err != nil {
			log.Error(fmt.Sprintf("key=%s: data deletion failed - %s", apiKey, err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Database error."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "Account data deleted successfully."})
	}
}

func regenerateUserID(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := requireAPIKeyParam(c)
		if !ok {
			return
		}

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
}

func checkHealth(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		uptime := int(time.Since(startTime).Seconds())

		if err := db.CheckConnection(ctx); err != nil {
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
