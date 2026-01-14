package usage

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/database"
)

// Client wraps a database connection and provides query methods
type Client struct {
	db *database.DB
}

// NewClient creates a new usage client with the given database connection
func NewClient(db *database.DB) *Client {
	return &Client{db: db}
}

// NewClientFromEnv creates a new client by connecting to the database
// using POSTGRES_URL from environment variables
func NewClientFromEnv(ctx context.Context) (*Client, error) {
	// Try to load .env file (not fatal if it doesn't exist)
	godotenv.Load(".env")

	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		return nil, ErrPostgresURLNotSet
	}

	db, err := database.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

// DB returns the underlying database connection
func (c *Client) DB() *database.DB {
	return c.db
}
