package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/config"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

// Client provides methods for cleaning up expired data
type Client struct {
	cfg         *config.Config
	db          *database.DB
	usageClient *usage.Client
	ctx         context.Context
}

// NewClient creates a new cleanup client
func NewClient(ctx context.Context, cfg *config.Config, db *database.DB) *Client {
	return &Client{
		cfg:         cfg,
		db:          db,
		usageClient: usage.NewClient(db),
		ctx:         ctx,
	}
}

// NewClientFromEnv creates a new cleanup client from environment variables
func NewClientFromEnv(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadWithRequired("POSTGRES_URL")
	if err != nil {
		return nil, err
	}

	db, err := cfg.NewDatabase(ctx)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return nil, err
	}

	log.Println("Database connection pool initialized")
	return NewClient(ctx, cfg, db), nil
}

// Close closes the database connection
func (c *Client) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

// DeleteExpiredData performs cleanup of both expired requests and users
func (c *Client) DeleteExpiredData(requestsLimit int, userExpiry time.Duration, deleteUsers bool) error {
	if deleteUsers {
		if err := c.DeleteExpiredUsers(userExpiry); err != nil {
			log.Printf("Error deleting expired users: %v", err)
			return err
		}
	}

	if err := c.DeleteExpiredRequests(requestsLimit); err != nil {
		log.Printf("Error deleting expired requests: %v", err)
		return err
	}

	return nil
}
