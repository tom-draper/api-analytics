package usage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

type RequestRow struct {
	RequestID    int            `json:"request_id"`
	APIKey       string         `json:"api_key"`
	Path         string         `json:"path"`
	Hostname     sql.NullString `json:"hostname"`
	IPAddress    sql.NullString `json:"ip_address"`
	Location     sql.NullString `json:"location"`
	UserAgentID  sql.NullInt64  `json:"user_agent_id"`
	Method       int16          `json:"method"`
	Status       int16          `json:"status"`
	ResponseTime int16          `json:"response_time"`
	Framework    int16          `json:"framework"`
	CreatedAt    time.Time      `json:"created_at"`
}

func HourlyRequestsCount(ctx context.Context) (int, error) {
	return RequestsCount(ctx, hourly)
}

func DailyRequestsCount(ctx context.Context) (int, error) {
	return RequestsCount(ctx, daily)
}

func WeeklyRequestsCount(ctx context.Context) (int, error) {
	return RequestsCount(ctx, weekly)
}

func MonthlyRequestsCount(ctx context.Context) (int, error) {
	return RequestsCount(ctx, monthly)
}

func TotalRequestsCount(ctx context.Context) (int, error) {
	return RequestsCount(ctx, "")
}

func RequestsCount(ctx context.Context, interval string) (int, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	var count int
	query := "SELECT COUNT(*) FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	err = conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func HourlyRequests(ctx context.Context) ([]RequestRow, error) {
	return Requests(ctx, hourly)
}

func DailyRequests(ctx context.Context) ([]RequestRow, error) {
	return Requests(ctx, daily)
}

func WeeklyRequests(ctx context.Context) ([]RequestRow, error) {
	return Requests(ctx, weekly)
}

func MonthlyRequests(ctx context.Context) ([]RequestRow, error) {
	return Requests(ctx, monthly)
}

func TotalRequests(ctx context.Context) ([]RequestRow, error) {
	return Requests(ctx, "")
}

func Requests(ctx context.Context, interval string) ([]RequestRow, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT request_id, api_key, path, hostname, ip_address, location, user_agent_id, method, status, response_time, framework, created_at FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += ";"

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []RequestRow
	for rows.Next() {
		var request RequestRow
		if err := rows.Scan(&request.RequestID, &request.APIKey, &request.Path, &request.Hostname, &request.IPAddress, &request.Location, &request.UserAgentID, &request.Method, &request.Status, &request.ResponseTime, &request.Framework, &request.CreatedAt); err != nil {
			return nil, err // Return the error if scanning fails
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err // Check for any errors during iteration
	}

	return requests, nil
}

func HourlyUserRequests(ctx context.Context) ([]UserCount, error) {
	return UserRequests(ctx, hourly)
}

func DailyUserRequests(ctx context.Context) ([]UserCount, error) {
	return UserRequests(ctx, daily)
}

func WeeklyUserRequests(ctx context.Context) ([]UserCount, error) {
	return UserRequests(ctx, weekly)
}

func MonthlyUserRequests(ctx context.Context) ([]UserCount, error) {
	return UserRequests(ctx, monthly)
}

func TotalUserRequests(ctx context.Context) ([]UserCount, error) {
	return UserRequests(ctx, "")
}

func UserRequests(ctx context.Context, interval string) ([]UserCount, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, COUNT(*) as count FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " GROUP BY api_key ORDER BY count;"

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []UserCount
	for rows.Next() {
		var userRequests UserCount
		if err := rows.Scan(&userRequests.APIKey, &userRequests.Count); err != nil {
			return nil, err
		}
		requests = append(requests, userRequests)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func UserRequestsOverLimit(ctx context.Context, limit int) ([]UserCount, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, COUNT(*) as count FROM requests GROUP BY api_key HAVING COUNT(*) > $1 ORDER BY count;"
	rows, err := conn.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []UserCount
	for rows.Next() {
		var userRequests UserCount
		if err := rows.Scan(&userRequests.APIKey, &userRequests.Count); err != nil {
			return nil, err
		}
		requests = append(requests, userRequests)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

type RequestsColumnSizes struct {
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

func (r RequestsColumnSizes) Display() {
	fmt.Printf("request_id: %s\napi_key: %s\npath: %s\nhostname: %s\nip_address: %s\nlocation: %s\nuser_agent: %s\nmethod: %s\nstatus: %s\nresponse_time: %s\nframework: %s\ncreated_at: %s\n",
		r.RequestID, r.APIKey, r.Path, r.Hostname, r.IPAddress, r.Location, r.UserAgent, r.Method, r.Status, r.ResponseTime, r.Framework, r.CreatedAt)
}

func RequestsColumnSize(ctx context.Context) (RequestsColumnSizes, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return RequestsColumnSizes{}, err
	}
	defer conn.Close(ctx)

	var size RequestsColumnSizes
	query := `
		SELECT 
			pg_size_pretty(sum(pg_column_size(request_id))) AS request_id, 
			pg_size_pretty(sum(pg_column_size(api_key))) AS api_key, 
			pg_size_pretty(sum(pg_column_size(path))) AS path, 
			pg_size_pretty(sum(pg_column_size(hostname))) AS hostname, 
			pg_size_pretty(sum(pg_column_size(ip_address))) AS ip_address, 
			pg_size_pretty(sum(pg_column_size(location))) AS location, 
			pg_size_pretty(sum(pg_column_size(user_agent_id))) AS user_agent_id, 
			pg_size_pretty(sum(pg_column_size(method))) AS method, 
			pg_size_pretty(sum(pg_column_size(status))) AS status, 
			pg_size_pretty(sum(pg_column_size(response_time))) AS response_time, 
			pg_size_pretty(sum(pg_column_size(framework))) AS framework, 
			pg_size_pretty(sum(pg_column_size(created_at))) AS created_at 
		FROM requests;`

	err = conn.QueryRow(ctx, query).Scan(&size.RequestID, &size.APIKey, &size.Path, &size.Hostname, &size.IPAddress, &size.Location, &size.UserAgent, &size.Method, &size.Status, &size.ResponseTime, &size.Framework, &size.CreatedAt)
	if err != nil {
		return RequestsColumnSizes{}, err
	}

	return size, nil
}

type ColumnValueCount[T any] struct {
	Value T
	Count int
}

func ColumnValuesCount[T any](ctx context.Context, column string) ([]ColumnValueCount[T], error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM requests GROUP BY %s ORDER BY count DESC;", column, column)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []ColumnValueCount[T]
	for rows.Next() {
		var count ColumnValueCount[T]
		if err := rows.Scan(&count.Value, &count.Count); err != nil {
			return nil, err
		}
		counts = append(counts, count)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

func TopFrameworks(ctx context.Context) ([]ColumnValueCount[int], error) {
	return ColumnValuesCount[int](ctx, "framework")
}

func TopUserAgents(ctx context.Context) ([]ColumnValueCount[string], error) {
	return ColumnValuesCount[string](ctx, "user_agent")
}

func TopIPAddresses(ctx context.Context) ([]ColumnValueCount[string], error) {
	return ColumnValuesCount[string](ctx, "ip_address")
}

func TopLocations(ctx context.Context) ([]ColumnValueCount[string], error) {
	return ColumnValuesCount[string](ctx, "location")
}

func AvgResponseTime(ctx context.Context) (float64, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return 0.0, err
	}
	defer conn.Close(ctx)

	var avg float64
	query := "SELECT AVG(response_time) FROM requests;"
	err = conn.QueryRow(ctx, query).Scan(&avg)
	if err != nil {
		return 0.0, err
	}

	return avg, nil
}
