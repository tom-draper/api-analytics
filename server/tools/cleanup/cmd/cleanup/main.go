package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: could not load .env file")
	}

	// Initialize database connection pool
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_URL is not set in the environment")
	}

	db, err := database.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}
	defer db.Close()
	log.Println("Database connection pool initialized")

	if *targetUserPtr != "" {
		cleanup.DeleteUser(db, *targetUserPtr)
		return
	}

	if *usersPtr {
		if err := cleanup.DeleteExpiredUsers(db, *userExpiryPtr); err != nil {
			log.Printf("Error deleting expired users: %v", err)
		}
	}

	if err := cleanup.DeleteExpiredRequests(db, *requestsLimitPtr); err != nil {
		log.Printf("Error deleting expired requests: %v", err)
	}
}