package pkg

import (
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/config"
	"github.com/tom-draper/api-analytics/server/tools/monitor/internal/monitor"
)

// Client is the public monitor client
type Client = monitor.Client

// NewClient creates a new monitor client
func NewClient(apiBaseURL, monitorAPIKey, monitorUserID string, db *database.DB) *Client {
	return monitor.NewClient(apiBaseURL, monitorAPIKey, monitorUserID, db)
}

// NewClientFromConfig creates a new monitor client from config
func NewClientFromConfig(cfg *config.Config, db *database.DB) (*Client, error) {
	return monitor.NewClientFromConfig(cfg, db)
}

// ServiceDown checks if a service is down
func ServiceDown(service string) bool {
	return monitor.ServiceDown(service)
}

// Standalone test functions for backward compatibility with checkup tool

func TryNewUserWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string, db *database.DB) error {
	return monitor.TryNewUser(apiBaseURL, monitorAPIKey, monitorUserID, db)
}

func TryFetchDashboardDataWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchDashboardData(apiBaseURL, monitorAPIKey, monitorUserID)
}

func TryFetchDataWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchData(apiBaseURL, monitorAPIKey, monitorUserID)
}

func TryFetchUserIDWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchUserID(apiBaseURL, monitorAPIKey, monitorUserID)
}

func TryFetchMonitorPingsWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchMonitorPings(apiBaseURL, monitorAPIKey, monitorUserID)
}

func TryLogRequestsWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string, legacy bool) error {
	return monitor.TryLogRequests(apiBaseURL, monitorAPIKey, monitorUserID, legacy)
}
