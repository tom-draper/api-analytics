package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/cleanup/internal/cleanup"
)

func main() {
	usersPtr := flag.Bool("users", false, "Delete expired users")
	targetUserPtr := flag.String("target-user", "", "Specify an API key for account deletion")
	helpPtr := flag.Bool("help", false, "Display help")

	requestsLimitPtr := flag.Int("requests-limit", 1_500_000, "Maximum number of requests per user before old requests are deleted")
	userExpiryPtr := flag.Duration("user-expiry", time.Hour*24*30*6, "Duration after which unused or retired users are deleted")

	flag.Parse()

	if *helpPtr {
		fmt.Printf("Cleanup - A command-line tool to delete expired users and requests.\n\nOptions:\n")
		flag.PrintDefaults()
		return
	}

	conn, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	if *targetUserPtr != "" {
		cleanup.DeleteUser(conn, *targetUserPtr)
		return
	}

	if *usersPtr {
		if err := cleanup.DeleteExpiredUsers(conn, *userExpiryPtr); err != nil {
			log.Printf("Error deleting expired users: %v", err)
		}
	}

	if err := cleanup.DeleteExpiredRequests(conn, *requestsLimitPtr); err != nil {
		log.Printf("Error deleting expired requests: %v", err)
	}
}