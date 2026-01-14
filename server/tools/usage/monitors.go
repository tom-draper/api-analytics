package usage

import (
	"context"
	"fmt"
)

// Monitor queries

// MonitorsCount returns the count of monitors within the given interval
func (c *Client) MonitorsCount(ctx context.Context, interval string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	err := c.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// HourlyMonitorsCount returns the count of monitors in the last hour
func (c *Client) HourlyMonitorsCount(ctx context.Context) (int, error) {
	return c.MonitorsCount(ctx, Hourly)
}

// DailyMonitorsCount returns the count of monitors in the last 24 hours
func (c *Client) DailyMonitorsCount(ctx context.Context) (int, error) {
	return c.MonitorsCount(ctx, Daily)
}

// WeeklyMonitorsCount returns the count of monitors in the last week
func (c *Client) WeeklyMonitorsCount(ctx context.Context) (int, error) {
	return c.MonitorsCount(ctx, Weekly)
}

// MonthlyMonitorsCount returns the count of monitors in the last month
func (c *Client) MonthlyMonitorsCount(ctx context.Context) (int, error) {
	return c.MonitorsCount(ctx, Monthly)
}

// TotalMonitorsCount returns the total count of all monitors
func (c *Client) TotalMonitorsCount(ctx context.Context) (int, error) {
	return c.MonitorsCount(ctx, "")
}

// Monitors returns monitors within the given interval
func (c *Client) Monitors(ctx context.Context, interval string) ([]MonitorRow, error) {
	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " ORDER BY created_at;"
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []MonitorRow
	for rows.Next() {
		var monitor MonitorRow
		if err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}

// HourlyMonitors returns monitors from the last hour
func (c *Client) HourlyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return c.Monitors(ctx, Hourly)
}

// DailyMonitors returns monitors from the last 24 hours
func (c *Client) DailyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return c.Monitors(ctx, Daily)
}

// WeeklyMonitors returns monitors from the last week
func (c *Client) WeeklyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return c.Monitors(ctx, Weekly)
}

// MonthlyMonitors returns monitors from the last month
func (c *Client) MonthlyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return c.Monitors(ctx, Monthly)
}

// TotalMonitors returns all monitors
func (c *Client) TotalMonitors(ctx context.Context) ([]MonitorRow, error) {
	return c.Monitors(ctx, "")
}

// UserMonitors returns monitor counts grouped by user within the given interval
func (c *Client) UserMonitors(ctx context.Context, interval string) ([]UserCount, error) {
	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " GROUP BY api_key ORDER BY count DESC;"
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []UserCount
	for rows.Next() {
		var userMonitors UserCount
		if err := rows.Scan(&userMonitors.APIKey, &userMonitors.Count); err != nil {
			return nil, err
		}
		monitors = append(monitors, userMonitors)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}

// HourlyUserMonitors returns user monitor counts for the last hour
func (c *Client) HourlyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return c.UserMonitors(ctx, Hourly)
}

// DailyUserMonitors returns user monitor counts for the last 24 hours
func (c *Client) DailyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return c.UserMonitors(ctx, Daily)
}

// WeeklyUserMonitors returns user monitor counts for the last week
func (c *Client) WeeklyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return c.UserMonitors(ctx, Weekly)
}

// MonthlyUserMonitors returns user monitor counts for the last month
func (c *Client) MonthlyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return c.UserMonitors(ctx, Monthly)
}

// TotalUserMonitors returns user monitor counts for all time
func (c *Client) TotalUserMonitors(ctx context.Context) ([]UserCount, error) {
	return c.UserMonitors(ctx, "")
}
