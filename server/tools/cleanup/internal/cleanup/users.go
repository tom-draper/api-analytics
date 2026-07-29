package cleanup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// DeleteExpiredUsers removes users who have been unused or retired past the
// expiry duration. When force is false it only logs what it would delete.
func (c *Client) DeleteExpiredUsers(userExpiry time.Duration, force bool) error {
	if err := c.deleteExpiredUnusedUsers(userExpiry, force); err != nil {
		return err
	}
	if err := c.deleteExpiredRetiredUsers(userExpiry, force); err != nil {
		return err
	}
	return nil
}

// deleteExpiredUnusedUsers removes users who never made requests and are past expiry
func (c *Client) deleteExpiredUnusedUsers(userExpiry time.Duration, force bool) error {
	usersList, err := c.usageClient.UnusedUsers(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch unused users: %w", err)
	}

	log.Printf("%d unused users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			c.deleteExpiredUser(user.APIKey, force)
		}
	}
	return nil
}

// deleteExpiredRetiredUsers removes users who haven't made requests in a while and are past expiry
func (c *Client) deleteExpiredRetiredUsers(userExpiry time.Duration, force bool) error {
	usersList, err := c.usageClient.SinceLastRequestUsers(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch retired users: %w", err)
	}

	log.Printf("%d retired users found\n", len(usersList))
	for _, user := range usersList {
		if time.Since(user.CreatedAt) > userExpiry {
			c.deleteExpiredUser(user.APIKey, force)
		}
	}
	return nil
}

// deleteExpiredUser deletes one expired user in batch mode. Without force it is
// a dry run (log only) so an unattended cron run never deletes by surprise, and
// it never prompts — a per-user prompt would make batch cleanup impossible
// interactively and silently cancel every deletion when stdin is not a terminal.
func (c *Client) deleteExpiredUser(apiKey string, force bool) {
	if !force {
		log.Printf("Would delete user %s (re-run with --yes to apply)", apiKey)
		return
	}
	c.deleteUser(apiKey)
}

// DeleteUser removes a single targeted user, prompting for confirmation unless
// force is set.
func (c *Client) DeleteUser(apiKey string, force bool) {
	if !force && !c.confirmDeletion(apiKey) {
		log.Println("User deletion cancelled.")
		return
	}
	c.deleteUser(apiKey)
}

// deleteUser removes a user and all their associated data from every table,
// without prompting.
func (c *Client) deleteUser(apiKey string) {
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

	log.Printf("User %s deletion successful.\n", apiKey)
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
