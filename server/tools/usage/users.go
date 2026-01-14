package usage

import (
	"context"
	"fmt"
)

// User queries

// UsersCount returns the count of users within the given interval
func (c *Client) UsersCount(ctx context.Context, interval string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	err := c.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// HourlyUsersCount returns the count of users in the last hour
func (c *Client) HourlyUsersCount(ctx context.Context) (int, error) {
	return c.UsersCount(ctx, Hourly)
}

// DailyUsersCount returns the count of users in the last 24 hours
func (c *Client) DailyUsersCount(ctx context.Context) (int, error) {
	return c.UsersCount(ctx, Daily)
}

// WeeklyUsersCount returns the count of users in the last week
func (c *Client) WeeklyUsersCount(ctx context.Context) (int, error) {
	return c.UsersCount(ctx, Weekly)
}

// MonthlyUsersCount returns the count of users in the last month
func (c *Client) MonthlyUsersCount(ctx context.Context) (int, error) {
	return c.UsersCount(ctx, Monthly)
}

// TotalUsersCount returns the total count of all users
func (c *Client) TotalUsersCount(ctx context.Context) (int, error) {
	return c.UsersCount(ctx, "")
}

// Users returns users within the given interval
func (c *Client) Users(ctx context.Context, interval string) ([]UserRow, error) {
	query := "SELECT api_key, user_id, created_at FROM users"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserRow
	for rows.Next() {
		var user UserRow
		if err := rows.Scan(&user.APIKey, &user.UserID, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// HourlyUsers returns users from the last hour
func (c *Client) HourlyUsers(ctx context.Context) ([]UserRow, error) {
	return c.Users(ctx, Hourly)
}

// DailyUsers returns users from the last 24 hours
func (c *Client) DailyUsers(ctx context.Context) ([]UserRow, error) {
	return c.Users(ctx, Daily)
}

// WeeklyUsers returns users from the last week
func (c *Client) WeeklyUsers(ctx context.Context) ([]UserRow, error) {
	return c.Users(ctx, Weekly)
}

// MonthlyUsers returns users from the last month
func (c *Client) MonthlyUsers(ctx context.Context) ([]UserRow, error) {
	return c.Users(ctx, Monthly)
}

// TotalUsers returns all users
func (c *Client) TotalUsers(ctx context.Context) ([]UserRow, error) {
	return c.Users(ctx, "")
}

// TopUsers returns the top N users by request count
func (c *Client) TopUsers(ctx context.Context, n int) ([]User, error) {
	query := `
		SELECT requests.api_key, users.created_at, COUNT(*) AS total_requests,
		       COALESCE(SUM(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 ELSE 0 END), 0) AS daily_requests,
		       COALESCE(SUM(CASE WHEN DATE(created_at) >= CURRENT_DATE - INTERVAL '7 days' THEN 1 ELSE 0 END), 0) AS weekly_requests
		FROM requests
		LEFT JOIN users ON users.api_key = requests.api_key
		GROUP BY requests.api_key, users.created_at
		ORDER BY total_requests DESC
		LIMIT $1;`
	rows, err := c.db.Pool.Query(ctx, query, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.TotalRequests, &user.DailyRequests, &user.WeeklyRequests); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UnusedUsers returns users with no requests or monitors
func (c *Client) UnusedUsers(ctx context.Context) ([]UserTime, error) {
	usersRequests, err := c.UnusedUsersRequests(ctx)
	if err != nil {
		return nil, err
	}

	usersMonitors, err := c.UnusedUsersMonitors(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]UserTime, 0)
	userMap := make(map[string]bool)
	for _, user := range usersMonitors {
		userMap[user.APIKey] = true
	}
	for _, userRequests := range usersRequests {
		if userMap[userRequests.APIKey] {
			users = append(users, userRequests)
		}
	}

	return users, nil
}

// UnusedUsersRequests returns users with no requests
func (c *Client) UnusedUsersRequests(ctx context.Context) ([]UserTime, error) {
	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM users u WHERE NOT EXISTS (SELECT FROM requests WHERE api_key = u.api_key) ORDER BY created_at;"
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserTime
	for rows.Next() {
		var user UserTime
		if err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.Days); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UnusedUsersMonitors returns users with no monitors
func (c *Client) UnusedUsersMonitors(ctx context.Context) ([]UserTime, error) {
	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM users u WHERE NOT EXISTS (SELECT FROM monitors WHERE api_key = u.api_key) ORDER BY created_at;"
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserTime
	for rows.Next() {
		var user UserTime
		if err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.Days); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// SinceLastRequestUsers returns users sorted by their last request time
func (c *Client) SinceLastRequestUsers(ctx context.Context) ([]UserTime, error) {
	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM (SELECT DISTINCT ON (api_key) api_key, created_at FROM requests ORDER BY api_key, created_at DESC) AS derived_table ORDER BY created_at;"
	rows, err := c.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserTime
	for rows.Next() {
		var user UserTime
		if err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.Days); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
