package usage

import (
	"context"
	"fmt"
)

// Request queries

// RequestsCount returns the count of requests within the given interval
func (c *Client) RequestsCount(ctx context.Context, interval string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	err := c.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// HourlyRequestsCount returns the count of requests in the last hour
func (c *Client) HourlyRequestsCount(ctx context.Context) (int, error) {
	return c.RequestsCount(ctx, Hourly)
}

// DailyRequestsCount returns the count of requests in the last 24 hours
func (c *Client) DailyRequestsCount(ctx context.Context) (int, error) {
	return c.RequestsCount(ctx, Daily)
}

// WeeklyRequestsCount returns the count of requests in the last week
func (c *Client) WeeklyRequestsCount(ctx context.Context) (int, error) {
	return c.RequestsCount(ctx, Weekly)
}

// MonthlyRequestsCount returns the count of requests in the last month
func (c *Client) MonthlyRequestsCount(ctx context.Context) (int, error) {
	return c.RequestsCount(ctx, Monthly)
}

// TotalRequestsCount returns the total count of all requests
func (c *Client) TotalRequestsCount(ctx context.Context) (int, error) {
	return c.RequestsCount(ctx, "")
}

// Requests returns requests within the given interval
func (c *Client) Requests(ctx context.Context, interval string) ([]RequestRow, error) {
	query := "SELECT request_id, api_key, path, hostname, ip_address, location, user_agent_id, method, status, response_time, framework, created_at FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += ";"

	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []RequestRow
	for rows.Next() {
		var request RequestRow
		if err := rows.Scan(&request.RequestID, &request.APIKey, &request.Path, &request.Hostname, &request.IPAddress, &request.Location, &request.UserAgentID, &request.Method, &request.Status, &request.ResponseTime, &request.Framework, &request.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// HourlyRequests returns requests from the last hour
func (c *Client) HourlyRequests(ctx context.Context) ([]RequestRow, error) {
	return c.Requests(ctx, Hourly)
}

// DailyRequests returns requests from the last 24 hours
func (c *Client) DailyRequests(ctx context.Context) ([]RequestRow, error) {
	return c.Requests(ctx, Daily)
}

// WeeklyRequests returns requests from the last week
func (c *Client) WeeklyRequests(ctx context.Context) ([]RequestRow, error) {
	return c.Requests(ctx, Weekly)
}

// MonthlyRequests returns requests from the last month
func (c *Client) MonthlyRequests(ctx context.Context) ([]RequestRow, error) {
	return c.Requests(ctx, Monthly)
}

// TotalRequests returns all requests
func (c *Client) TotalRequests(ctx context.Context) ([]RequestRow, error) {
	return c.Requests(ctx, "")
}

// UserRequests returns request counts grouped by user within the given interval
func (c *Client) UserRequests(ctx context.Context, interval string) ([]UserCount, error) {
	query := "SELECT api_key, COUNT(*) as count FROM requests"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}
	query += " GROUP BY api_key ORDER BY count;"

	rows, err := c.db.Pool.Query(ctx, query)
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

// HourlyUserRequests returns user request counts for the last hour
func (c *Client) HourlyUserRequests(ctx context.Context) ([]UserCount, error) {
	return c.UserRequests(ctx, Hourly)
}

// DailyUserRequests returns user request counts for the last 24 hours
func (c *Client) DailyUserRequests(ctx context.Context) ([]UserCount, error) {
	return c.UserRequests(ctx, Daily)
}

// WeeklyUserRequests returns user request counts for the last week
func (c *Client) WeeklyUserRequests(ctx context.Context) ([]UserCount, error) {
	return c.UserRequests(ctx, Weekly)
}

// MonthlyUserRequests returns user request counts for the last month
func (c *Client) MonthlyUserRequests(ctx context.Context) ([]UserCount, error) {
	return c.UserRequests(ctx, Monthly)
}

// TotalUserRequests returns user request counts for all time
func (c *Client) TotalUserRequests(ctx context.Context) ([]UserCount, error) {
	return c.UserRequests(ctx, "")
}

// UserRequestsOverLimit returns users with request counts over the given limit
func (c *Client) UserRequestsOverLimit(ctx context.Context, limit int) ([]UserCount, error) {
	query := "SELECT api_key, COUNT(*) as count FROM requests GROUP BY api_key HAVING COUNT(*) > $1 ORDER BY count;"
	rows, err := c.db.Pool.Query(ctx, query, limit)
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

// RequestsColumnSize returns the size of each column in the requests table
func (c *Client) RequestsColumnSize(ctx context.Context) (RequestsColumnSizes, error) {
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

	err := c.db.Pool.QueryRow(ctx, query).Scan(&size.RequestID, &size.APIKey, &size.Path, &size.Hostname, &size.IPAddress, &size.Location, &size.UserAgent, &size.Method, &size.Status, &size.ResponseTime, &size.Framework, &size.CreatedAt)
	if err != nil {
		return RequestsColumnSizes{}, err
	}

	return size, nil
}

// ColumnValuesCount returns counts of values in a specific column
func (c *Client) ColumnValuesCount(ctx context.Context, column string) ([]ColumnValueCount[any], error) {
	query := fmt.Sprintf("SELECT %s, COUNT(*) AS count FROM requests GROUP BY %s ORDER BY count DESC;", column, column)
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []ColumnValueCount[any]
	for rows.Next() {
		var count ColumnValueCount[any]
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

// TopFrameworks returns the most used frameworks
func (c *Client) TopFrameworks(ctx context.Context) ([]ColumnValueCount[any], error) {
	return c.ColumnValuesCount(ctx, "framework")
}

// TopUserAgents returns the most used user agents
func (c *Client) TopUserAgents(ctx context.Context) ([]ColumnValueCount[any], error) {
	return c.ColumnValuesCount(ctx, "user_agent")
}

// TopIPAddresses returns the most common IP addresses
func (c *Client) TopIPAddresses(ctx context.Context) ([]ColumnValueCount[any], error) {
	return c.ColumnValuesCount(ctx, "ip_address")
}

// TopLocations returns the most common locations
func (c *Client) TopLocations(ctx context.Context) ([]ColumnValueCount[any], error) {
	return c.ColumnValuesCount(ctx, "location")
}

// AvgResponseTime returns the average response time of all requests
func (c *Client) AvgResponseTime(ctx context.Context) (float64, error) {
	var avg float64
	query := "SELECT AVG(response_time) FROM requests;"
	err := c.db.Pool.QueryRow(ctx, query).Scan(&avg)
	if err != nil {
		return 0.0, err
	}

	return avg, nil
}
