package usage

import (
	"fmt"
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
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT COUNT(*) FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s';", interval)
	} else {
		query += ";"
	}
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}

	var count int
	rows.Next()
	err = rows.Scan(&count)
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
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" where created_at >= NOW() - interval '%s';", interval)
	} else {
		query += ";"
	}
	rows, err := db.Query(query)
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
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s' GROUP BY api_key ORDER BY count DESC", interval)
	}
	query += " GROUP BY api_key ORDER BY count;"
	rows, err := db.Query(query)
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
