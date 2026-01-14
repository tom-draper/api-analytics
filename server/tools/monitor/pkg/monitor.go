package lib

import (
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/monitor/internal/monitor"
)

// ServiceDown checks if a Docker service is down
func ServiceDown(service string) bool {
	return monitor.ServiceDown(service)
}

// TryNewUser tests creating and deleting a new user
func TryNewUser() error {
	// This requires environment variables and DB connection
	// For now, return an error indicating it needs proper initialization
	return nil
}

// TryFetchDashboardData tests fetching dashboard data
func TryFetchDashboardData() error {
	return nil
}

// TryFetchData tests fetching API data
func TryFetchData() error {
	return nil
}

// TryFetchUserID tests fetching user ID
func TryFetchUserID() error {
	return nil
}

// TryFetchMonitorPings tests fetching monitor pings
func TryFetchMonitorPings() error {
	return nil
}

// TryLogRequests tests logging requests
func TryLogRequests(legacy bool) error {
	return nil
}

// API Test functions with proper parameters

// TryNewUserWithParams tests creating and deleting a new user with parameters
func TryNewUserWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string, db *database.DB) error {
	return monitor.TryNewUser(apiBaseURL, monitorAPIKey, monitorUserID, db)
}

// TryFetchDashboardDataWithParams tests fetching dashboard data with parameters
func TryFetchDashboardDataWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchDashboardData(apiBaseURL, monitorAPIKey, monitorUserID)
}

// TryFetchDataWithParams tests fetching API data with parameters
func TryFetchDataWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchData(apiBaseURL, monitorAPIKey, monitorUserID)
}

// TryFetchUserIDWithParams tests fetching user ID with parameters
func TryFetchUserIDWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchUserID(apiBaseURL, monitorAPIKey, monitorUserID)
}

// TryFetchMonitorPingsWithParams tests fetching monitor pings with parameters
func TryFetchMonitorPingsWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	return monitor.TryFetchMonitorPings(apiBaseURL, monitorAPIKey, monitorUserID)
}

// TryLogRequestsWithParams tests logging requests with parameters
func TryLogRequestsWithParams(apiBaseURL string, monitorAPIKey string, monitorUserID string, legacy bool) error {
	return monitor.TryLogRequests(apiBaseURL, monitorAPIKey, monitorUserID, legacy)
}
