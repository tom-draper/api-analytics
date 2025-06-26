package cleanup

import (
	"context"
	"strings"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

func DeleteOldestRequests(conn *pgx.Conn, apiKey string, count int) error {
	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err := conn.Exec(context.Background(), query, apiKey, count)
	return err
}

func DeleteExpiredRequests(conn *pgx.Conn, requestsLimit int) error {
	users, err := usage.UserRequestsOverLimit(context.Background(), requestsLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch users over limit: %w", err)
	}

	log.Printf("%d users found\n", len(users))
	for _, user := range users {
		err = DeleteOldestRequests(conn, user.APIKey, user.Count-requestsLimit)
		if err != nil {
			log.Printf("Error deleting requests for user %s: %v", user.APIKey, err)
			continue
		}
		log.Printf("%s: %d requests deleted\n", user.APIKey, user.Count-requestsLimit)
	}
	return nil
}

func DeleteExpiredUsers(conn *pgx.Conn, userExpiry time.Duration) error {
	if err := deleteExpiredUnusedUsers(conn, userExpiry); err != nil {
		return err
	}
	if err := deleteExpiredRetiredUsers(conn, userExpiry); err != nil {
		return err
	}
	return nil
}

func deleteExpiredUnusedUsers(conn *pgx.Conn, userExpiry time.Duration) error {
	users, err := usage.UnusedUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch unused users: %w", err)
	}

	log.Printf("%d unused users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			DeleteUser(conn, user.APIKey)
		}
	}
	return nil
}

func deleteExpiredRetiredUsers(conn *pgx.Conn, userExpiry time.Duration) error {
	users, err := usage.SinceLastRequestUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch retired users: %w", err)
	}

	log.Printf("%d retired users found\n", len(users))
	for _, user := range users {
		if time.Since(user.CreatedAt) > userExpiry {
			DeleteUser(conn, user.APIKey)
		}
	}
	return nil
}

func DeleteUser(conn *pgx.Conn, apiKey string) {
	if !confirmDeletion(apiKey) {
		log.Println("User deletion cancelled.")
		return
	}

	deleteFromTables := []struct {
		name       string
		deleteFunc func(string) error
	}{
		{"users", database.DeleteUser},
		{"requests", database.DeleteRequests},
		{"monitors", database.DeleteMonitors},
		{"pings", database.DeletePings},
	}

	for _, table := range deleteFromTables {
		err := table.deleteFunc(apiKey)
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
