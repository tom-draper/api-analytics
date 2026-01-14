package checkup

import (
	"context"
	"log"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/config"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

// Client provides methods for system checkup and monitoring
type Client struct {
	cfg          *config.Config
	db           *database.DB
	usageClient  *usage.Client
	printer      *message.Printer
	ctx          context.Context
}

// NewClient creates a new checkup client
func NewClient(ctx context.Context, cfg *config.Config, db *database.DB) *Client {
	return &Client{
		cfg:         cfg,
		db:          db,
		usageClient: usage.NewClient(db),
		printer:     message.NewPrinter(language.English),
		ctx:         ctx,
	}
}

// NewClientFromEnv creates a new checkup client from environment variables
func NewClientFromEnv(ctx context.Context, requiredFields ...string) (*Client, error) {
	var cfg *config.Config
	var err error

	if len(requiredFields) > 0 {
		cfg, err = config.LoadWithRequired(requiredFields...)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		return nil, err
	}

	// Database is required for checkup
	if cfg.PostgresURL == "" {
		log.Println("POSTGRES_URL environment variable not set")
		return nil, err
	}

	db, err := cfg.NewDatabase(ctx)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return nil, err
	}

	return NewClient(ctx, cfg, db), nil
}

// Close closes the database connection
func (c *Client) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

// handleError logs errors and returns them
func (c *Client) handleError(err error) error {
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	return err
}
