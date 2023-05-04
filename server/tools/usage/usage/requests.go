package usage

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
)

func DailyRequestsCount() (int, error) {
	return RequestsCount(1)
}

func WeeklyRequestsCount() (int, error) {
	return RequestsCount(7)
}

func MonthlyRequestsCount() (int, error) {
	return RequestsCount(30)
}

func TotalRequestsCount() (int, error) {
	return RequestsCount(0)
}

func RequestsCount(days int) (int, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT COUNT(*) FROM requests"
	if days == 999 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else if days >= 1 {
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

func DailyRequests() ([]database.RequestRow, error) {
	return Requests(1)
}

func WeeklyRequests() ([]database.RequestRow, error) {
	return Requests(7)
}

func MonthlyRequests() ([]database.RequestRow, error) {
	return Requests(30)
}

func TotalRequests() ([]database.RequestRow, error) {
	return Requests(0)
}

func Requests(days int) ([]database.RequestRow, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT request_id, api_key, path, hostname, ip_address, location, user_agent, method, status, response_time, framework, created_at FROM requests"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours';"
	} else if days > 1 {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day';", days)
	} else {
		query += ";"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var requests []database.RequestRow
	for rows.Next() {
		var request database.RequestRow
		err := rows.Scan(&request.RequestID, &request.APIKey, &request.Path, &request.Hostname, &request.IPAddress, &request.Location, &request.UserAgent, &request.Method, &request.Status, &request.ResponseTime, &request.Framework, &request.CreatedAt)
		if err == nil {
			requests = append(requests, request)
		}
	}

	return requests, nil
}

func DailyUserRequests() ([]UserCount, error) {
	return UserRequests(1)
}

func WeeklyUserRequests() ([]UserCount, error) {
	return UserRequests(7)
}

func MonthlyUserRequests() ([]UserCount, error) {
	return UserRequests(30)
}

func TotalUserRequests() ([]UserCount, error) {
	return UserRequests(0)
}

func UserRequests(days int) ([]UserCount, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, COUNT(*) as count FROM requests"
	if days == 1 {
		query += " WHERE created_at >= NOW() - interval '24 hours'"
	} else if days > 1 {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%d day'", days)
	}
	query += " GROUP BY api_key ORDER BY count;"
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

type StringCount struct {
	Value string
	Count string
}

func columnStringValuesCount(column string) ([]StringCount, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM requests GROUP BY %s ORDER BY count DESC;", column, column)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var count []StringCount
	for rows.Next() {
		var userAgent StringCount
		err := rows.Scan(&userAgent.Value, &userAgent.Count)
		if err == nil {
			count = append(count, userAgent)
		}
	}

	return count, nil
}

func UserAgents() ([]StringCount, error) {
	return columnStringValuesCount("user_agent")
}

func IPAddresses() ([]StringCount, error) {
	return columnStringValuesCount("ip_address")
}

func Locations() ([]StringCount, error) {
	return columnStringValuesCount("location")
}

func AvgResponseTime() (float64, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT AVG(response_time) FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		return 0.0, err
	}

	var avg float64
	rows.Next()
	err = rows.Scan(&avg)
	if err != nil {
		return 0.0, err
	}

	return avg, nil
}
