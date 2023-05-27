package usage

import (
	"fmt"
	"github.com/tom-draper/api-analytics/server/database"
)

func DailyMonitorsCount() (int, error) {
	return MonitorsCount(1)
}

func WeeklyMonitorsCount() (int, error) {
	return MonitorsCount(7)
}

func MonthlyMonitorsCount() (int, error) {
	return MonitorsCount(30)
}

func TotalMonitorsCount() (int, error) {
	return MonitorsCount(0)
}

func MonitorsCount(days int) (int, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT COUNT(*) FROM monitor"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else if days > 1 {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day';", days)
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

func DailyMonitors() ([]database.MonitorRow, error) {
	return Monitors(1)
}

func WeeklyMonitors() ([]database.MonitorRow, error) {
	return Monitors(7)
}

func MonthlyMonitors() ([]database.MonitorRow, error) {
	return Monitors(30)
}

func TotalMonitors() ([]database.MonitorRow, error) {
	return Monitors(0)
}

func Monitors(days int) ([]database.MonitorRow, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, url, secure, ping, created_at FROM monitor"
	if days == 1 {
		query += " where created_at >= NOW() - interval '24 hours';"
	} else if days > 1 {
		query += fmt.Sprintf(" where created_at >= NOW() - interval '%d day';", days)
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

func DailyUserMonitors() ([]UserCount, error) {
	return UserMonitors(1)
}

func WeeklyUserMonitors() ([]UserCount, error) {
	return UserMonitors(7)
}

func MonthlyUserMonitors() ([]UserCount, error) {
	return UserMonitors(30)
}

func TotalUserMonitors() ([]UserCount, error) {
	return UserMonitors(0)
}

func UserMonitors(days int) ([]UserCount, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, COUNT(*) AS count FROM monitor"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours' GROUP BY api_key ORDER BY count DESC"
	} else if days > 1 {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day' GROUP BY api_key ORDER BY count DESC", days)
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