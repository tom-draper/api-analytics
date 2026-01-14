package main

import (
	"context"
	"log"

	"github.com/tom-draper/api-analytics/server/tools/checkup/cmd/checkup/options"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/checkup"
)

func main() {
	opts := options.GetOptions()
	if *opts.Help {
		options.DisplayHelp()
		return
	}

	ctx := context.Background()

	// Determine required config fields based on options
	var requiredFields []string
	if *opts.Email {
		requiredFields = []string{"POSTGRES_URL", "EMAIL_ADDRESS"}
	} else {
		requiredFields = []string{"POSTGRES_URL"}
	}

	// Create checkup client
	client, err := checkup.NewClientFromEnv(ctx, requiredFields...)
	if err != nil {
		log.Fatalf("Failed to initialize checkup client: %v", err)
	}
	defer client.Close()

	// Run the appropriate checkup
	if *opts.Email {
		if err := client.SendEmailCheckup(); err != nil {
			log.Fatalf("Failed to send email checkup: %v", err)
		}
	} else if *opts.Users {
		client.DisplayUsersCheckup()
	} else if *opts.Monitors {
		client.DisplayMonitorsCheckup()
	} else if *opts.Database {
		client.DisplayDatabaseCheckup()
	} else {
		client.DisplayCheckup()
	}
}