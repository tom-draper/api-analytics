package usage

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
)

func TotalRequests() (int, error) {
	db := database.OpenDBConnection()

	query := "SELECT COUNT(*) AS count FROM requests;"
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

type UserCount struct {
	APIKey string
	Count  string
}

func TopUsers() ([]UserCount, error) {
	db := database.OpenDBConnection()

	query := "SELECT api_key, COUNT(*) AS count FROM requests GROUP BY api_key ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var requests []UserCount
	for rows.Next() {
		request := new(UserCount)
		err := rows.Scan(&request.APIKey, &request.Count)
		if err == nil {
			requests = append(requests, *request)
		}
	}

	return requests, nil
}

func DailyTotalUsers() (int, error) {
	return TotalUsers(1)
}

func WeeklyTotalUsers() (int, error) {
	return TotalUsers(7)
}

func MonthlyTotalUsers() (int, error) {
	return TotalUsers(30)
}

func TotalUsers(days int) (int, error) {
	db := database.OpenDBConnection()

	query := "SELECT count(*) FROM users"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day';", days)
	}
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}

	var users int
	rows.Next()
	err = rows.Scan(&users)
	if err != nil {
		return 0, err
	}

	return users, nil
}

func DailyUsers() ([]database.UserRow, error) {
	return Users(1)
}

func WeeklyUsers() ([]database.UserRow, error) {
	return Users(7)
}

func MonthlyUsers() ([]database.UserRow, error) {
	return Users(30)
}

func Users(days int) ([]database.UserRow, error) {
	db := database.OpenDBConnection()

	query := "SELECT * FROM users"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day';", days)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var users []database.UserRow
	for rows.Next() {
		user := new(database.UserRow)
		err := rows.Scan(&user.APIKey, &user.UserID, &user.CreatedAt)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, nil
}

func DailyUsage() ([]UserCount, error) {
	return Usage(1)
}

func WeeklyUsage() ([]UserCount, error) {
	return Usage(7)
}

func MonthlyUsage() ([]UserCount, error) {
	return Usage(30)
}

func Usage(days int) ([]UserCount, error) {
	db := database.OpenDBConnection()

	query := "SELECT api_key, created_at FROM requests"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day';", days)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var requests []UserCount
	for rows.Next() {
		userRequests := new(UserCount)
		err := rows.Scan(&userRequests.APIKey, &userRequests.Count)
		if err == nil {
			requests = append(requests, *userRequests)
		}
	}

	return requests, nil
}

func DailyMonitors() ([]UserCount, error) {
	return Monitors(1)
}

func WeeklyMonitors() ([]UserCount, error) {
	return Monitors(7)
}

func MonthlyMonitors() ([]UserCount, error) {
	return Monitors(30)
}

func Monitors(days int) ([]UserCount, error) {
	db := database.OpenDBConnection()

	var query string
	if days == 1 {
		query = "SELECT api_key, COUNT(*) AS count FROM monitor where created_at >= NOW() - interval '24 hours' GROUP BY api_key ORDER BY count DESC;"
	} else {
		query = fmt.Sprintf("SELECT api_key, COUNT(*) AS count FROM monitor where created_at >= NOW() - interval '%d day' GROUP BY api_key ORDER BY count DESC;", days)
	}
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

func DatabaseSize() (string, error) {
	db := database.OpenDBConnection()

	query := "SELECT pg_size_pretty(pg_total_relation_size('requests'));"
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}

	var size string
	rows.Next()
	err = rows.Scan(&size)
	return size, err
}

func DatabaseConnections() (int, error) {
	db := database.OpenDBConnection()

	query := "SELECT count(*) from pg_stat_activity;"
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}

	var connections int
	rows.Next()
	err = rows.Scan(&connections)
	return connections, err
}
