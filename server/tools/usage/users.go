package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
)

type UserRow struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func HourlyUsersCount(ctx context.Context) (int, error) {
	return UsersCount(ctx, "1 hour")
}

func DailyUsersCount(ctx context.Context) (int, error) {
	return UsersCount(ctx, "24 hours")
}

func WeeklyUsersCount(ctx context.Context) (int, error) {
	return UsersCount(ctx, "7 days")
}

func MonthlyUsersCount(ctx context.Context) (int, error) {
	return UsersCount(ctx, "30 days")
}

func TotalUsersCount(ctx context.Context) (int, error) {
	return UsersCount(ctx, "")
}

func UsersCount(ctx context.Context, interval string) (int, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	var count int
	query := "SELECT COUNT(*) FROM users"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	err = conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func HourlyUsers(ctx context.Context) ([]UserRow, error) {
	return Users(ctx, "1 hour")
}

func DailyUsers(ctx context.Context) ([]UserRow, error) {
	return Users(ctx, "24 hours")
}

func WeeklyUsers(ctx context.Context) ([]UserRow, error) {
	return Users(ctx, "7 days")
}

func MonthlyUsers(ctx context.Context) ([]UserRow, error) {
	return Users(ctx, "30 days")
}

func TotalUsers(ctx context.Context) ([]UserRow, error) {
	return Users(ctx, "")
}

func Users(ctx context.Context, interval string) ([]UserRow, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, user_id, created_at FROM users"
	if interval != "" {
		query += fmt.Sprintf(" WHERE created_at >= NOW() - interval '%s'", interval)
	}

	rows, err := conn.Query(ctx, query)
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

type User struct {
	APIKey         string    `json:"api_key"`
	TotalRequests  int       `json:"total_requests"`
	DailyRequests  int       `json:"daily_requests"`
	WeeklyRequests int       `json:"weekly_requests"`
	CreatedAt      time.Time `json:"created_at"`
}

func (user User) Display(rank int) {
	var colorPrintf func(format string, a ...interface{})
	if user.DailyRequests == 0 && user.WeeklyRequests == 0 {
		colorPrintf = color.Red
	} else if user.DailyRequests == 0 || user.WeeklyRequests == 0 {
		colorPrintf = color.Yellow
	} else {
		colorPrintf = color.Green
	}
	colorPrintf("[%d] %s %d (+%d / +%d) %s\n", rank, user.APIKey, user.TotalRequests, user.DailyRequests, user.WeeklyRequests, user.CreatedAt.Format("2006-01-02 15:04:05"))
}

func DisplayUsers(users []User) {
	for i, user := range users {
		user.Display(i)
	}
}

func TopUsers(ctx context.Context, n int) ([]User, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := `
		SELECT requests.api_key, users.created_at, COUNT(*) AS total_requests,
		       COALESCE(SUM(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 ELSE 0 END), 0) AS daily_requests,
		       COALESCE(SUM(CASE WHEN DATE(created_at) >= CURRENT_DATE - INTERVAL '7 days' THEN 1 ELSE 0 END), 0) AS weekly_requests
		FROM requests
		LEFT JOIN users ON users.api_key = requests.api_key
		GROUP BY requests.api_key, users.created_at
		ORDER BY total_requests DESC
		LIMIT $1;`
	rows, err := conn.Query(ctx, query, n)
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

type UserTime struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	Days      string    `json:"days"`
}

func (user UserTime) Display(rank int) {
	format := "[%d] %s %s (%s)\n"
	timeFormat := "2006-01-02 15:04:05"
	daysSince := time.Since(user.CreatedAt)

	switch {
	case daysSince > time.Hour*24*30*6:
		color.Red(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	case daysSince > time.Hour*24*30*3:
		color.Yellow(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	default:
		fmt.Printf(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	}
}

func DisplayUserTimes(users []UserTime) {
	for i, user := range users {
		user.Display(i)
	}
}

func UnusedUsers(ctx context.Context) ([]UserTime, error) {
	usersRequests, err := UnusedUsersRequests(ctx)
	if err != nil {
		return nil, err
	}

	usersMonitors, err := UnusedUsersMonitors(ctx)
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

func UnusedUsersRequests(ctx context.Context) ([]UserTime, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM users u WHERE NOT EXISTS (SELECT FROM requests WHERE api_key = u.api_key) ORDER BY created_at;"
	rows, err := conn.Query(ctx, query)
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

func UnusedUsersMonitors(ctx context.Context) ([]UserTime, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM users u WHERE NOT EXISTS (SELECT FROM monitors WHERE api_key = u.api_key) ORDER BY created_at;"
	rows, err := conn.Query(ctx, query)
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

func SinceLastRequestUsers(ctx context.Context) ([]UserTime, error) {
	conn, err := database.NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM (SELECT DISTINCT ON (api_key) api_key, created_at FROM requests ORDER BY api_key, created_at DESC) AS derived_table ORDER BY created_at;"
	rows, err := conn.Query(ctx, query)
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
