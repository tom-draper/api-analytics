package usage

import (
	"context"

	"github.com/tom-draper/api-analytics/server/database"
)

func HourlyMonitorsCount() (int, error) {
	return MonitorsCount("1 hour")
}

func DailyMonitorsCount() (int, error) {
	return MonitorsCount("24 hours")
}

func WeeklyMonitorsCount() (int, error) {
	return MonitorsCount("7 days")
}

func MonthlyMonitorsCount() (int, error) {
	return MonitorsCount("30 days")
}

func TotalMonitorsCount() (int, error) {
	return MonitorsCount("")
}

func MonitorsCount(interval string) (int, error) {
	conn := database.NewConnection()
	defer conn.Close(context.Background())

	var count int
	query := "SELECT COUNT(*) FROM monitor"
	if interval != "" {
		query += " WHERE created_at >= NOW() - interval $1"
	}
	query += ";"
	err := conn.QueryRow(context.Background(), query, interval).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func HourlyMonitors() ([]database.MonitorRow, error) {
	return Monitors("1 hour")
}

func DailyMonitors() ([]database.MonitorRow, error) {
	return Monitors("24 hours")
}

func WeeklyMonitors() ([]database.MonitorRow, error) {
	return Monitors("7 days")
}

func MonthlyMonitors() ([]database.MonitorRow, error) {
	return Monitors("30 days")
}

func TotalMonitors() ([]database.MonitorRow, error) {
	return Monitors("")
}

func Monitors(interval string) ([]database.MonitorRow, error) {
	conn := database.NewConnection()
	defer conn.Close(context.Background())

	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if interval != "" {
		query += " WHERE created_at >= NOW() - interval $1"
	}
	query += "ORDER BY created_at;"
	rows, err := conn.Query(context.Background(), query, interval)
	if err != nil {
		return nil, err
	}

	var monitors []database.MonitorRow
	for rows.Next() {
		var monitor database.MonitorRow
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, monitor)
		}
	}

	return monitors, nil
}

func HourlyUserMonitors() ([]UserCount, error) {
	return UserMonitors("1 hour")
}

func DailyUserMonitors() ([]UserCount, error) {
	return UserMonitors("24 hours")
}

func WeeklyUserMonitors() ([]UserCount, error) {
	return UserMonitors("7 days")
}

func MonthlyUserMonitors() ([]UserCount, error) {
	return UserMonitors("30 days")
}

func TotalUserMonitors() ([]UserCount, error) {
	return UserMonitors("")
}

func UserMonitors(interval string) ([]UserCount, error) {
	conn := database.NewConnection()
	defer conn.Close(context.Background())

	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if interval != "" {
		query += " WHERE created_at >= NOW() - interval $1"
	}
	query += " GROUP BY api_key ORDER BY count DESC;"
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var monitors []UserCount
	for rows.Next() {
		userMonitors := new(UserCount)
		err := rows.Scan(&userMonitors.APIKey, &userMonitors.Count)
		if err == nil {
			monitors = append(monitors, *userMonitors)
		}
	}

	return monitors, nil
}
