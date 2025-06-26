package monitors

import (
	"context"
	"fmt"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

type MonitorRow struct {
	APIKey    string    `json:"api_key"`
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func HourlyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, usage.Hourly)
}

func DailyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, usage.Daily)
}

func WeeklyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, usage.Weekly)
}

func MonthlyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, usage.Monthly)
}

func TotalMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, "")
}

func MonitorsCount(ctx context.Context, interval string) (int, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	var count int
	query := "SELECT COUNT(*) FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	err = conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func HourlyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, usage.Hourly)
}

func DailyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, usage.Daily)
}

func WeeklyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, usage.Weekly)
}

func MonthlyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, usage.Monthly)
}

func TotalMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, "")
}

func Monitors(ctx context.Context, interval string) ([]MonitorRow, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " ORDER BY created_at;"
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []MonitorRow
	for rows.Next() {
		var monitor MonitorRow
		if err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt); err != nil {
			return nil, err // Return error instead of silently continuing
		}
		monitors = append(monitors, monitor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}

func HourlyUserMonitors(ctx context.Context) ([]usage.UserCount, error) {
	return UserMonitors(ctx, usage.Hourly)
}

func DailyUserMonitors(ctx context.Context) ([]usage.UserCount, error) {
	return UserMonitors(ctx, usage.Daily)
}

func WeeklyUserMonitors(ctx context.Context) ([]usage.UserCount, error) {
	return UserMonitors(ctx, usage.Weekly)
}

func MonthlyUserMonitors(ctx context.Context) ([]usage.UserCount, error) {
	return UserMonitors(ctx, usage.Monthly)
}

func TotalUserMonitors(ctx context.Context) ([]usage.UserCount, error) {
	return UserMonitors(ctx, "")
}

func UserMonitors(ctx context.Context, interval string) ([]usage.UserCount, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " GROUP BY api_key ORDER BY count DESC;"
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []usage.UserCount
	for rows.Next() {
		var userMonitors usage.UserCount
		if err := rows.Scan(&userMonitors.APIKey, &userMonitors.Count); err != nil {
			return nil, err // Return error instead of silently continuing
		}
		monitors = append(monitors, userMonitors)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return monitors, nil
}
