package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

type MonitorRow struct {
	APIKey    string    `json:"api_key"`
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

func HourlyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, hourly)
}

func DailyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, daily)
}

func WeeklyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, weekly)
}

func MonthlyMonitorsCount(ctx context.Context) (int, error) {
	return MonitorsCount(ctx, monthly)
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
	return Monitors(ctx, hourly)
}

func DailyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, daily)
}

func WeeklyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, weekly)
}

func MonthlyMonitors(ctx context.Context) ([]MonitorRow, error) {
	return Monitors(ctx, monthly)
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

func HourlyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return UserMonitors(ctx, hourly)
}

func DailyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return UserMonitors(ctx, daily)
}

func WeeklyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return UserMonitors(ctx, weekly)
}

func MonthlyUserMonitors(ctx context.Context) ([]UserCount, error) {
	return UserMonitors(ctx, monthly)
}

func TotalUserMonitors(ctx context.Context) ([]UserCount, error) {
	return UserMonitors(ctx, "")
}

func UserMonitors(ctx context.Context, interval string) ([]UserCount, error) {
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

	var monitors []UserCount
	for rows.Next() {
		var userMonitors UserCount
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
