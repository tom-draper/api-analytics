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

func HourlyMonitorsCount(ctx context.Context, db *database.DB) (int, error) {
	return MonitorsCount(ctx, db, usage.Hourly)
}

func DailyMonitorsCount(ctx context.Context, db *database.DB) (int, error) {
	return MonitorsCount(ctx, db, usage.Daily)
}

func WeeklyMonitorsCount(ctx context.Context, db *database.DB) (int, error) {
	return MonitorsCount(ctx, db, usage.Weekly)
}

func MonthlyMonitorsCount(ctx context.Context, db *database.DB) (int, error) {
	return MonitorsCount(ctx, db, usage.Monthly)
}

func TotalMonitorsCount(ctx context.Context, db *database.DB) (int, error) {
	return MonitorsCount(ctx, db, "")
}

func MonitorsCount(ctx context.Context, db *database.DB, interval string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	err := db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func HourlyMonitors(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	return Monitors(ctx, db, usage.Hourly)
}

func DailyMonitors(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	return Monitors(ctx, db, usage.Daily)
}

func WeeklyMonitors(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	return Monitors(ctx, db, usage.Weekly)
}

func MonthlyMonitors(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	return Monitors(ctx, db, usage.Monthly)
}

func TotalMonitors(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	return Monitors(ctx, db, "")
}

func Monitors(ctx context.Context, db *database.DB, interval string) ([]MonitorRow, error) {
	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " ORDER BY created_at;"
	rows, err := db.Pool.Query(ctx, query)
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

func HourlyUserMonitors(ctx context.Context, db *database.DB) ([]usage.UserCount, error) {
	return UserMonitors(ctx, db, usage.Hourly)
}

func DailyUserMonitors(ctx context.Context, db *database.DB) ([]usage.UserCount, error) {
	return UserMonitors(ctx, db, usage.Daily)
}

func WeeklyUserMonitors(ctx context.Context, db *database.DB) ([]usage.UserCount, error) {
	return UserMonitors(ctx, db, usage.Weekly)
}

func MonthlyUserMonitors(ctx context.Context, db *database.DB) ([]usage.UserCount, error) {
	return UserMonitors(ctx, db, usage.Monthly)
}

func TotalUserMonitors(ctx context.Context, db *database.DB) ([]usage.UserCount, error) {
	return UserMonitors(ctx, db, "")
}

func UserMonitors(ctx context.Context, db *database.DB, interval string) ([]usage.UserCount, error) {
	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " GROUP BY api_key ORDER BY count DESC;"
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []usage.UserCount
	for rows.Next() {
		var userMonitors usage.UserCount
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
