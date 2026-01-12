package cleanup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage/requests"
	"github.com/tom-draper/api-analytics/server/tools/usage/users"
)

func DeleteOldestRequests(db *database.DB, apiKey string, count int) error {
	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err := db.Pool.Exec(context.Background(), query, apiKey, count)
	return err
}

func DeleteExpiredRequests(db *database.DB, requestsLimit int) error {
	usersList, err := requests.UserRequestsOverLimit(context.Background(), db, requestsLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch users over limit: %w", err)
	}

	log.Printf("%d users found\n", len(usersList))
	for _, user := range usersList {
		err = DeleteOldestRequests(db, user.APIKey, user.Count-requestsLimit)
		if err != nil {
			log.Printf("Error deleting requests for user %s: %v", user.APIKey, err)
			continue
		}
		log.Printf("%s: %d requests deleted\n", user.APIKey, user.Count-requestsLimit)
	}
	return nil
}

func DeleteExpiredUsers(db *database.DB, userExpiry time.Duration) error {
	if err := deleteExpiredUnusedUsers(db, userExpiry); err != nil {
		return err
	}
	if err := deleteExpiredRetiredUsers(db, userExpiry); err != nil {
		return err
	}
	return nil
}

func deleteExpiredUnusedUsers(db *database.DB, userExpiry time.Duration) error {
	usersList, err := users.UnusedUsers(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to fetch unused users: %w", err)
	}

	log.Printf("%d unused users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			DeleteUser(db, user.APIKey)
		}
	}
	return nil
}

func deleteExpiredRetiredUsers(db *database.DB, userExpiry time.Duration) error {
	usersList, err := users.SinceLastRequestUsers(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to fetch retired users: %w", err)
	}

	log.Printf("%d retired users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			DeleteUser(db, user.APIKey)
		}
	}
	return nil
}

func DeleteUser(db *database.DB, apiKey string) {
	if !confirmDeletion(apiKey) {
		log.Println("User deletion cancelled.")
		return
	}

	ctx := context.Background()
	deleteFromTables := []struct {
		name       string
		deleteFunc func(context.Context, string) error
	}{
		{"requests", db.DeleteRequests},
		{"monitors", db.DeleteMonitors},
		{"pings", db.DeletePings},
		{"users", db.DeleteUser},
	}

	for _, table := range deleteFromTables {
		err := table.deleteFunc(ctx, apiKey)
		if err != nil {
			log.Printf("Failed to delete user from '%s' table: %v", table.name, err)
			return
		}
		log.Printf("User deleted from table '%s'.\n", table.name)
	}

	log.Println("User deletion successful.")
}

func confirmDeletion(apiKey string) bool {
	fmt.Printf("Delete API key '%s' from the database? (Y/n): ", apiKey)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return false
	}
	response = strings.ToLower(response)
	return response == "y" || response == "yes"
}
