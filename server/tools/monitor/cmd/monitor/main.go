package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/config"
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
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadWithRequired("API_BASE_URL", "MONITOR_API_KEY", "MONITOR_USER_ID", "POSTGRES_URL", "EMAIL_ADDRESS")
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize database connection
	db, err := cfg.NewDatabase(ctx)
	if err != nil {
		log.Printf("Warning: could not connect to database: %v", err)
		db = nil
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// Create monitor client
	monitorClient, err := monitor.NewClientFromConfig(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create monitor client: %v", err)
	}

	// Run checks
	serviceStatus := monitorClient.RunServiceChecks()
	apiTestStatus := monitorClient.RunAPITests()

	// Send email alert if any failures detected
	if serviceStatus.ServiceDown() || apiTestStatus.TestFailed() {
		emailClient, err := email.NewClientFromEnv()
		if err != nil {
			log.Fatalf("Error creating email client: %v", err)
		}

		body := buildEmailBody(serviceStatus, apiTestStatus)
		err = emailClient.Send(email.Message{
			To:      []string{cfg.EmailAddress},
			From:    "",
			Subject: "API Analytics",
			Body:    body,
		})
		if err != nil {
			log.Fatalf("Error sending email: %v", err)
		}
	}
}
