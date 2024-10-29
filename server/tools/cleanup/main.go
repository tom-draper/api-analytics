package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"log"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

const requestsLimit int = 1_500_000
const userExpiry time.Duration = time.Hour * 24 * 30 * 6

func deleteOldestRequests(apiKey string, count int) error {
	conn, err := database.NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err = conn.Exec(context.Background(), query, apiKey, count)
	return err
}

func deleteExpiredRequests() {
	users, err := usage.UserRequestsOverLimit(context.Background(), requestsLimit)
	if err != nil {
		log.Fatalf("Failed to fetch users over limit: %v", err) // Use log for error handling
	}

	log.Printf("%d users found\n", len(users))
	for _, user := range users {
		err = deleteOldestRequests(user.APIKey, user.Count-requestsLimit)
		if err != nil {
			log.Printf("Error deleting requests for user %s: %v", user.APIKey, err)
			continue // Don't panic, just log the error and continue
		}
		log.Printf("%s: %d requests deleted\n", user.APIKey, user.Count-requestsLimit)
	}
}

func deleteExpiredUsers() {
	deleteExpiredUnusedUsers()
	deleteExpiredRetiredUsers()
}

func deleteExpiredUnusedUsers() {
	users, err := usage.UnusedUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch unused users: %v", err)
	}

	log.Printf("%d unused users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			deleteUser(user.APIKey)
		}
	}
}

func deleteExpiredRetiredUsers() {
	users, err := usage.SinceLastRequestUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch retired users: %v", err)
	}

	log.Printf("%d retired users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			deleteUser(user.APIKey)
		}
	}
}

func deleteUser(apiKey string) {
	fmt.Printf("Delete API key '%s' from the database? (Y/n): ", apiKey)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return
	}
	response = strings.ToLower(response)
	if response != "y" && response != "yes" {
		log.Println("User deletion cancelled.")
		return
	}

	err = database.DeleteUser(apiKey)
	if err != nil {
		log.Printf("Failed to delete user from 'users' table: %v", err)
		return
	}
	log.Println("User deleted from table 'users'.")

	err = database.DeleteRequests(apiKey)
	if err != nil {
		log.Printf("Failed to delete user from 'requests' table: %v", err)
		return
	}
	log.Println("User deleted from table 'requests'.")

	err = database.DeleteMonitors(apiKey)
	if err != nil {
		log.Printf("Failed to delete user from 'monitors' table: %v", err)
		return
	}
	log.Println("User deleted from table 'monitors'.")

	err = database.DeletePings(apiKey)
	if err != nil {
		log.Printf("Failed to delete user from 'pings' table: %v", err)
		return
	}
	log.Println("User deleted from table 'pings'.")

	log.Println("User deletion successful.")
}

type Options struct {
	users      bool
	targetUser string
	help       bool
}

func getOptions() Options {
	options := Options{}
	for i, arg := range os.Args {
		switch arg {
		case "--users":
			options.users = true
		case "--help":
			options.help = true
		case "--target-user":
			if i+1 < len(os.Args) {
				options.targetUser = os.Args[i+1]
			}
		}
	}
	return options
}

func displayHelp() {
	fmt.Printf("Cleanup - A command-line tool to delete expired users and requests.\n\nOptions:\n`--users` to delete expired users\n`--target-user` to specify an API key for account deletion\n`--help` to display help\n")
}

func main() {
	options := getOptions()
	if options.help {
		displayHelp()
		return
	}
	if options.targetUser != "" {
		deleteUser(options.targetUser)
		return
	}

	if options.users {
		deleteExpiredUsers()
	}
	deleteExpiredRequests()
}
