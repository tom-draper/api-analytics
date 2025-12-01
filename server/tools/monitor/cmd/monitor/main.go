package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/internal/monitor"
)

func buildEmailBody(serviceStatus monitor.ServiceStatus, apiTestStatus monitor.APITestStatus) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Failure detected at %v\n", time.Now()))

	if !serviceStatus.API {
		body.WriteString("Service api down\n")
	}
	if !serviceStatus.Logger {
		body.WriteString("Service logger down\n")
	}
	if !serviceStatus.Nginx {
		body.WriteString("Service nginx down\n")
	}
	if !serviceStatus.PostgresSQL {
		body.WriteString("Service postgresql down\n")
	}

	if apiTestStatus.NewUser != nil {
		body.WriteString(fmt.Sprintf("Error when creating new user: %s\n", apiTestStatus.NewUser.Error()))
	}
	if apiTestStatus.FetchDashboardData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching dashboard data: %s\n", apiTestStatus.FetchDashboardData.Error()))
	}
	if apiTestStatus.FetchData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching API data: %s\n", apiTestStatus.FetchData.Error()))
	}
	if apiTestStatus.FetchUserID != nil {
		body.WriteString(fmt.Sprintf("Error when fetching user ID: %s\n", apiTestStatus.FetchUserID.Error()))
	}
	if apiTestStatus.FetchMonitorPings != nil {
		body.WriteString(fmt.Sprintf("Error when fetching monitor pings: %s\n", apiTestStatus.FetchMonitorPings.Error()))
	}
	if apiTestStatus.LogRequests != nil {
		body.WriteString(fmt.Sprintf("Error when logging requests: %s\n", apiTestStatus.LogRequests.Error()))
	}

	return body.String()
}

func main() {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		log.Fatalf("API_BASE_URL environment variable not set")
	}
	monitorAPIKey := os.Getenv("MONITOR_API_KEY")
	if monitorAPIKey == "" {
		log.Fatalf("MONITOR_API_KEY environment variable not set")
	}
	monitorUserID := os.Getenv("MONITOR_USER_ID")
	if monitorUserID == "" {
		log.Fatalf("MONITOR_USER_ID environment variable not set")
	}

	serviceStatus := monitor.ServiceStatus{
		API:    !monitor.ServiceDown("api"),
		Logger: !monitor.ServiceDown("logger"),
		Nginx:  !monitor.ServiceDown("nginx"), PostgresSQL: !monitor.ServiceDown("postgreql"),
	}
	apiTestStatus := monitor.APITestStatus{
		NewUser:            monitor.TryNewUser(apiBaseURL, monitorAPIKey, monitorUserID),
		FetchDashboardData: monitor.TryFetchDashboardData(apiBaseURL, monitorAPIKey, monitorUserID),
		FetchData:          monitor.TryFetchData(apiBaseURL, monitorAPIKey, monitorUserID),
		FetchUserID:        monitor.TryFetchUserID(apiBaseURL, monitorAPIKey, monitorUserID),
		FetchMonitorPings:  monitor.TryFetchMonitorPings(apiBaseURL, monitorAPIKey, monitorUserID),
		LogRequests:        monitor.TryLogRequests(apiBaseURL, monitorAPIKey, monitorUserID, false),
	}
	if serviceStatus.ServiceDown() || apiTestStatus.TestFailed() {
		client, err := email.NewClientFromEnv()
		if err != nil {
			log.Fatalf("Error creating email client: %v", err)
		}
		address := os.Getenv("EMAIL_ADDRESS")
		if address == "" {
			log.Fatalf("EMAIL_ADDRESS environment variable not set")
		}
		body := buildEmailBody(serviceStatus, apiTestStatus)
		err = client.Send(email.Message{To: []string{address}, From: "", Subject: "API Analytics", Body: body})
		if err != nil {
			log.Fatalf("Error sending email: %v", err)
		}
	}
}
