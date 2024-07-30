package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

const requestsLimit int = 1_500_000
const userExpiry time.Duration = time.Hour * 24 * 30 * 6

func deleteOldestRequests(apiKey string, count int) error {
	conn := database.NewConnection()
	defer conn.Close(context.Background())

	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err := conn.Exec(context.Background(), query, apiKey, count)
	if err != nil {
		return err
	}
	return nil
}

func deleteExpiredRequests() {
	users, err := usage.UserRequestsOverLimit(requestsLimit)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d users found\n", len(users))
	for _, user := range users {
		err = deleteOldestRequests(user.APIKey, user.Count-requestsLimit)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %d requests deleted\n", user.APIKey, user.Count-requestsLimit)
	}
}

func deleteExpiredUsers() {
	deleteExpiredUnusedUsers()
	deleteExpiredRetiredUsers()
}

func deleteExpiredUnusedUsers() {
	users, err := usage.UnusedUsers()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			deleteUser(user.APIKey)
			fmt.Printf("%s: unused user expired", user.APIKey)
		}
	}
}

func deleteExpiredRetiredUsers() {
	users, err := usage.SinceLastRequestUsers()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			deleteUser(user.APIKey)
			fmt.Printf("%s: retired user expired", user.APIKey)
		}
	}
}

func deleteUser(apiKey string) {
	fmt.Println("Delete API key '%s' from the database? Y/n", apiKey)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	response = strings.ToLower(response)
	if response != "y" && response != "yes" {
		fmt.Println("User deletion cancelled.")
		return
	}

	err = database.DeleteUser(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'users'.")
	err = database.DeleteRequests(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'requests'.")
	err = database.DeleteMonitors(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'monitors'.")
	err = database.DeletePings(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'pings'.")

	fmt.Println("User deletion successful.")
}

type Options struct {
	users      bool
	targetUser string
	help       bool
}

func getOptions() Options {
	options := Options{}
	for i, arg := range os.Args {
		if arg == "--users" {
			options.users = true
		} else if arg == "--help" {
			options.help = true
		} else if i > 0 && os.Args[i-1] == "--target-user" {
			options.targetUser = arg
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
	} else if options.targetUser != "" {
		deleteUser(options.targetUser)
	} else {
		if options.users {
			deleteExpiredUsers()
		}
		deleteExpiredRequests()
	}
}
