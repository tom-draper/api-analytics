package main

import (
	"context"
	"log"

	"github.com/tom-draper/api-analytics/server/tools/cleanup/cmd/cleanup/options"
	"github.com/tom-draper/api-analytics/server/tools/cleanup/internal/cleanup"
)

func main() {
	opts := options.GetOptions()
	if *opts.Help {
		options.DisplayHelp()
		return
	}

	ctx := context.Background()

	// Create cleanup client
	client, err := cleanup.NewClientFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize cleanup client: %v", err)
	}
	defer client.Close()

	// Handle target user deletion
	if *opts.TargetUser != "" {
		client.DeleteUser(*opts.TargetUser)
		return
	}

	// Run cleanup
	if err := client.DeleteExpiredData(*opts.RequestsLimit, *opts.UserExpiry, *opts.Users); err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}
}