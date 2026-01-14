package monitor

import (
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/config"
)

// Client provides methods for monitoring API Analytics services
type Client struct {
	apiBaseURL    string
	monitorAPIKey string
	monitorUserID string
	db            *database.DB
	httpClient    *http.Client
}

// NewClient creates a new monitor client with the provided configuration
func NewClient(apiBaseURL, monitorAPIKey, monitorUserID string, db *database.DB) *Client {
	return &Client{
		apiBaseURL:    apiBaseURL,
		monitorAPIKey: monitorAPIKey,
		monitorUserID: monitorUserID,
		db:            db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewClientFromConfig creates a new monitor client from config
func NewClientFromConfig(cfg *config.Config, db *database.DB) (*Client, error) {
	if cfg.APIBaseURL == "" {
		return nil, ErrMissingAPIBaseURL
	}
	if cfg.MonitorAPIKey == "" {
		return nil, ErrMissingMonitorAPIKey
	}
	if cfg.MonitorUserID == "" {
		return nil, ErrMissingMonitorUserID
	}

	return NewClient(cfg.APIBaseURL, cfg.MonitorAPIKey, cfg.MonitorUserID, db), nil
}
