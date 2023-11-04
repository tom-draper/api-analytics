package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

const requestsLimit int = 5_000_000
const userExpiry time.Duration = time.Hour * 24 * 30 * 6

func deleteOldestRequests(apiKey string, count int) error {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err := db.Query(query, apiKey, count)
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
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			err := database.DeleteUser(user.APIKey)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s: unused user expired", user.APIKey)
		}
	}
}

func deleteExpiredRetiredUsers() {
	users, err := usage.SinceLastRequestUsers()
	if err != nil {
		panic(err)
	}
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			err := database.DeleteUser(user.APIKey)
			if err != nil {
				panic(err)
			}
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

func main() {
	deleteExpiredUsers()
	deleteExpiredRequests()
}
