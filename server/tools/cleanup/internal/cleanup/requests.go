package cleanup

import (
	"context"
	"fmt"
	"log"
)

// DeleteOldestRequests deletes the oldest N requests for a specific user
func (c *Client) DeleteOldestRequests(apiKey string, count int) error {
	query := "DELETE FROM requests WHERE request_id = any(array(SELECT request_id from requests WHERE api_key = $1 ORDER BY created_at LIMIT $2));"

	_, err := c.db.Pool.Exec(context.Background(), query, apiKey, count)
	return err
}

// DeleteExpiredRequests removes old requests for users over the limit
func (c *Client) DeleteExpiredRequests(requestsLimit int) error {
	usersList, err := c.usageClient.UserRequestsOverLimit(c.ctx, requestsLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch users over limit: %w", err)
	}

	log.Printf("%d users found over request limit\n", len(usersList))
	for _, user := range usersList {
		deleteCount := user.Count - requestsLimit
		err = c.DeleteOldestRequests(user.APIKey, deleteCount)
		if err != nil {
			log.Printf("Error deleting requests for user %s: %v", user.APIKey, err)
			continue
		}
		log.Printf("%s: %d requests deleted\n", user.APIKey, deleteCount)
	}
	return nil
}
