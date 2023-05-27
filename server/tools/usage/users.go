package usage

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
)

func DailyUsersCount() (int, error) {
	return UsersCount(1)
}

func WeeklyUsersCount() (int, error) {
	return UsersCount(7)
}

func MonthlyUsersCount() (int, error) {
	return UsersCount(30)
}

func TotalUsersCount() (int, error) {
	return UsersCount(0)
}

func UsersCount(days int) (int, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT COUNT(*) FROM users"
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

func DailyUsers() ([]database.UserRow, error) {
	return Users(1)
}

func WeeklyUsers() ([]database.UserRow, error) {
	return Users(7)
}

func MonthlyUsers() ([]database.UserRow, error) {
	return Users(30)
}

func TotalUsers() ([]database.UserRow, error) {
	return Users(0)
}

func Users(days int) ([]database.UserRow, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, user_id, created_at FROM users"
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

type User struct {
	APIKey         string    `json:"api_key"`
	TotalRequests  int       `json:"total_requests"`
	DailyRequests  int       `json:"daily_requests"`
	WeeklyRequests int       `json:"weekly_requests"`
	CreatedAt      time.Time `json:"created_at"`
}

func DisplayUsers(users []User) {
	var colorPrintf func(format string, a ...interface{})
	for i, user := range users {
		if user.DailyRequests == 0 && user.WeeklyRequests == 0 {
			colorPrintf = color.Red
		} else if user.DailyRequests == 0 || user.WeeklyRequests == 0 {
			colorPrintf = color.Yellow
		} else {
			colorPrintf = color.Green
		}
		colorPrintf("[%d] %s %d (+%d / +%d) %s\n", i, user.APIKey, user.TotalRequests, user.DailyRequests, user.WeeklyRequests, user.CreatedAt.Format("2006-01-02"))
	}
}

func TopUsers(n int) ([]User, error) {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("SELECT requests.api_key, users.created_at, COUNT(*) AS total_requests FROM requests left join users on users.api_key = requests.api_key GROUP BY requests.api_key, users.created_at ORDER BY total_requests DESC LIMIT %d", n)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var users []User
	for rows.Next() {
		user := new(User)
		err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.TotalRequests)
		if err == nil {
			users = append(users, *user)
		}
	}
	db.Close()

	// Add daily requests for each top user
	dailyRequests, err := DailyUserRequests()
	if err != nil {
		return nil, err
	}
	for _, userRequests := range dailyRequests {
		for i, r := range users {
			if userRequests.APIKey == r.APIKey {
				users[i].DailyRequests = userRequests.Count
			}
		}
	}

	// Add weekly requests for each top user
	weeklyRequests, err := WeeklyUserRequests()
	if err != nil {
		return nil, err
	}
	for _, userRequests := range weeklyRequests {
		for i, r := range users {
			if userRequests.APIKey == r.APIKey {
				users[i].WeeklyRequests = userRequests.Count
			}
		}
	}

	return users, nil
}

type UserTime struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	Days      string    `json:"days"`
}

func DisplayUserTimes(users []UserTime) {
	format := "[%d] %s %s (%s)\n"
	for i, user := range users {
		if time.Since(user.CreatedAt) > time.Hour*24*30*6 {
			color.Red(format, i, user.APIKey, user.CreatedAt.Format("2006-01-02"), user.Days)
		} else if time.Since(user.CreatedAt) > time.Hour*24*30*3 {
			color.Yellow(format, i, user.APIKey, user.CreatedAt.Format("2006-01-02"), user.Days)
		} else {
			fmt.Printf(format, i, user.APIKey, user.CreatedAt.Format("2006-01-02"), user.Days)
		}
	}
}

func UnusedUsers() ([]UserTime, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM users u WHERE NOT EXISTS (SELECT FROM requests WHERE api_key = u.api_key) ORDER BY created_at;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var users []UserTime
	for rows.Next() {
		user := new(UserTime)
		err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.Days)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, nil
}

func SinceLastRequestUsers() ([]UserTime, error) {
	db := database.OpenDBConnection()
	defer db.Close()

	query := "SELECT api_key, created_at, (NOW() - created_at) AS days FROM (SELECT DISTINCT ON (api_key) api_key, created_at FROM requests ORDER BY api_key, created_at DESC) AS derived_table ORDER BY created_at;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var users []UserTime
	for rows.Next() {
		user := new(UserTime)
		err := rows.Scan(&user.APIKey, &user.CreatedAt, &user.Days)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, nil
}
