package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool for optimal performance
	config.MaxConns = 25                              // Maximum concurrent connections
	config.MinConns = 5                               // Minimum idle connections (kept warm)
	config.MaxConnLifetime = 1 * time.Hour            // Recycle connections after 1 hour
	config.MaxConnIdleTime = 10 * time.Minute         // Close idle connections after 10 minutes
	config.HealthCheckPeriod = 1 * time.Minute        // Check connection health every minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
