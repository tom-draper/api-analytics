package usage

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
)

func HourlyRequestsCount() (int, error) {
	return RequestsCount("1 hour")
}

func DailyRequestsCount() (int, error) {
	return RequestsCount("24 hours")
}

func WeeklyRequestsCount() (int, error) {
	return RequestsCount("7 days")
}

func MonthlyRequestsCount() (int, error) {
	return RequestsCount("30 days")
}

func TotalRequestsCount() (int, error) {
	return RequestsCount("")
}

func RequestsCount(interval string) (int, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT COUNT(*) FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += ";"
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

func HourlyRequests() ([]database.RequestRow, error) {
	return Requests("1 hour")
}

func DailyRequests() ([]database.RequestRow, error) {
	return Requests("24 hours")
}

func WeeklyRequests() ([]database.RequestRow, error) {
	return Requests("7 days")
}

func MonthlyRequests() ([]database.RequestRow, error) {
	return Requests("30 days")
}

func TotalRequests() ([]database.RequestRow, error) {
	return Requests("")
}

func Requests(interval string) ([]database.RequestRow, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT request_id, api_key, path, hostname, ip_address, location, user_agent, method, status, response_time, framework, created_at FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += ";"

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

func HourlyUserRequests() ([]UserCount, error) {
	return UserRequests("1 hour")
}

func DailyUserRequests() ([]UserCount, error) {
	return UserRequests("24 hours")
}

func WeeklyUserRequests() ([]UserCount, error) {
	return UserRequests("7 days")
}

func MonthlyUserRequests() ([]UserCount, error) {
	return UserRequests("30 days")
}

func TotalUserRequests() ([]UserCount, error) {
	return UserRequests("")
}

func UserRequests(interval string) ([]UserCount, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, COUNT(*) as count FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
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

func UserRequestsOverLimit(limit int) ([]UserCount, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM (SELECT api_key, COUNT(*) as count FROM requests GROUP BY api_key) as derived_table WHERE count > %d ORDER BY count;", limit)
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

type requestsColumnSize struct {
	RequestID    string `json:"request_id"`
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	Location     string `json:"location"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       string `json:"status"`
	ResponseTime string `json:"response_time"`
	Framework    string `json:"framework"`
	CreatedAt    string `json:"created_at"`
}

func (r requestsColumnSize) Display() {
	fmt.Printf("request_id: %s\napi_key: %s\npath: %s\nhostname: %s\nip_address: %s\nlocation: %s\nuser_agent: %s\nmethod: %s\nstatus: %s\nresponse_time: %s\nframework: %s\ncreated_at: %s\n", r.RequestID, r.APIKey, r.Path, r.Hostname, r.IPAddress, r.Location, r.UserAgent, r.Method, r.Status, r.ResponseTime, r.Framework, r.CreatedAt)
}

func RequestsColumnSize() (requestsColumnSize, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT pg_size_pretty(sum(pg_column_size(request_id))) AS request_id, pg_size_pretty(sum(pg_column_size(api_key))) AS api_key, pg_size_pretty(sum(pg_column_size(path))) AS path, pg_size_pretty(sum(pg_column_size(hostname))) AS hostname, pg_size_pretty(sum(pg_column_size(ip_address))) AS ip_address, pg_size_pretty(sum(pg_column_size(location))) AS location, pg_size_pretty(sum(pg_column_size(user_agent))) AS user_agent, pg_size_pretty(sum(pg_column_size(method))) AS method, pg_size_pretty(sum(pg_column_size(status))) AS status, pg_size_pretty(sum(pg_column_size(response_time))) AS response_time, pg_size_pretty(sum(pg_column_size(framework))) AS framework, pg_size_pretty(sum(pg_column_size(created_at))) AS created_at FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		return requestsColumnSize{}, err
	}

	var size requestsColumnSize
	rows.Next()
	err = rows.Scan(&size.RequestID, &size.APIKey, &size.Path, &size.Hostname, &size.IPAddress, &size.Location, &size.UserAgent, &size.Method, &size.Status, &size.ResponseTime, &size.Framework, &size.CreatedAt)
	return size, err
}

func columnValuesCount[T string | int](column string) ([]struct {
	Value T
	Count int
}, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM requests GROUP BY %s ORDER BY count DESC;", column, column)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var count []struct {
		Value T
		Count int
	}
	for rows.Next() {
		var userAgent struct {
			Value T
			Count int
		}
		err := rows.Scan(&userAgent.Value, &userAgent.Count)
		if err == nil {
			count = append(count, userAgent)
		}
	}

	return count, nil
}

func TopFrameworks() ([]struct {
	Value int
	Count int
}, error) {
	return columnValuesCount[int]("framework")
}

func TopUserAgents() ([]struct {
	Value string
	Count int
}, error) {
	return columnValuesCount[string]("user_agent")
}

func TopIPAddresses() ([]struct {
	Value string
	Count int
}, error) {
	return columnValuesCount[string]("ip_address")
}

func TopLocations() ([]struct {
	Value string
	Count int
}, error) {
	return columnValuesCount[string]("location")
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
