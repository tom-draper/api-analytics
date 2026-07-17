package monitor

import (
	"github.com/tom-draper/api-analytics/server/database"
)

// Standalone wrapper functions for backward compatibility with pkg exports

func TryNewUser(apiBaseURL string, monitorAPIKey string, monitorUserID string, db *database.DB) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, db)
	return client.TryNewUser()
}

func TryFetchData(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, nil)
	return client.TryFetchData()
}

func TryFetchDashboardData(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, nil)
	return client.TryFetchDashboardData()
}

func TryFetchUserID(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, nil)
	return client.TryFetchUserID()
}

func TryFetchMonitorPings(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, nil)
	return client.TryFetchMonitorPings()
}

func TryLogRequests(apiBaseURL string, monitorAPIKey string, monitorUserID string, legacy bool) error {
	client := NewClient(apiBaseURL, monitorAPIKey, monitorUserID, nil)
	return client.TryLogRequests(legacy)
}
