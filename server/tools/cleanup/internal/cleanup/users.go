package cleanup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// DeleteExpiredUsers removes users who have been unused or retired past the expiry duration
func (c *Client) DeleteExpiredUsers(userExpiry time.Duration) error {
	if err := c.deleteExpiredUnusedUsers(userExpiry); err != nil {
		return err
	}
	if err := c.deleteExpiredRetiredUsers(userExpiry); err != nil {
		return err
	}
	return nil
}

// deleteExpiredUnusedUsers removes users who never made requests and are past expiry
func (c *Client) deleteExpiredUnusedUsers(userExpiry time.Duration) error {
	usersList, err := c.usageClient.UnusedUsers(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch unused users: %w", err)
	}

	log.Printf("%d unused users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			c.DeleteUser(user.APIKey)
		}
	}
	return nil
}

// deleteExpiredRetiredUsers removes users who haven't made requests in a while and are past expiry
func (c *Client) deleteExpiredRetiredUsers(userExpiry time.Duration) error {
	usersList, err := c.usageClient.SinceLastRequestUsers(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch retired users: %w", err)
	}

	log.Printf("%d retired users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			c.DeleteUser(user.APIKey)
		}
	}
	return nil
}

// DeleteUser removes a user and all their associated data from all tables
func (c *Client) DeleteUser(apiKey string) {
	if !c.confirmDeletion(apiKey) {
		log.Println("User deletion cancelled.")
		return
	}

	ctx := context.Background()
	deleteFromTables := []struct {
		name       string
		deleteFunc func(context.Context, string) error
	}{
		{"requests", c.db.DeleteRequests},
		{"monitors", c.db.DeleteMonitors},
		{"pings", c.db.DeletePings},
		{"users", c.db.DeleteUser},
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

// confirmDeletion prompts for confirmation before deleting a user
func (c *Client) confirmDeletion(apiKey string) bool {
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
