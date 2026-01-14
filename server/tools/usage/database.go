package usage

import (
	"context"
	"fmt"
)

// Database utilities

// TableSize returns the size of the given table
func (c *Client) TableSize(ctx context.Context, table string) (string, error) {
	var size string
	query := fmt.Sprintf("SELECT pg_size_pretty(pg_total_relation_size('%s'));", table)
	err := c.db.Pool.QueryRow(ctx, query).Scan(&size)
	if err != nil {
		return "", err
	}

	return size, nil
}

// TableColumnSize returns the size and percentage of a specific column in a table
func (c *Client) TableColumnSize(ctx context.Context, table, column string) (totalSize, averageSize string, percentage float64, err error) {
	query := fmt.Sprintf(`
		SELECT pg_size_pretty(sum(pg_column_size(%s))) AS total_size,
		       pg_size_pretty(avg(pg_column_size(%s))) AS average_size,
		       sum(pg_column_size(%s)) * 100.0 / pg_total_relation_size('%s') AS percentage
		FROM %s;`, column, column, column, table, table)

	err = c.db.Pool.QueryRow(ctx, query).Scan(&totalSize, &averageSize, &percentage)
	if err != nil {
		return "", "", 0, err
	}

	return totalSize, averageSize, percentage, nil
}

// DatabaseConnections returns the number of active database connections
func (c *Client) DatabaseConnections(ctx context.Context) (int, error) {
	var connections int
	query := "SELECT count(*) FROM pg_stat_activity;"
	err := c.db.Pool.QueryRow(ctx, query).Scan(&connections)
	if err != nil {
		return 0, err
	}

	return connections, nil
}
