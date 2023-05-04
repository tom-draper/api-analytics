package usage

import (
	"fmt"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	p := message.NewPrinter(language.English)
	for _, c := range users {
		p.Printf("%s: %d (+%d -- +%d) %s\n", c.APIKey, c.TotalRequests, c.DailyRequests, c.WeeklyRequests, c.CreatedAt.Format("2006-01-02"))
	}
}

func TopUsers(n int) ([]User, error) {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("SELECT requests.api_key, users.created_at, COUNT(*) AS total_requests FROM requests left join users on users.api_key = requests.api_key GROUP BY requests.api_key, users.created_at ORDER BY total_requests DESC LIMIT %d", n)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var requests []User
	for rows.Next() {
		request := new(User)
		err := rows.Scan(&request.APIKey, &request.CreatedAt, &request.TotalRequests)
		if err == nil {
			requests = append(requests, *request)
		}
	}
	db.Close()

	// Add daily requests for each top user
	dailyRequests, err := DailyUserRequests()
	if err != nil {
		return nil, err
	}
	for _, userRequests := range dailyRequests {
		for i, r := range requests {
			if userRequests.APIKey == r.APIKey {
				requests[i].DailyRequests = userRequests.Count
			}
		}
	}

	// Add weekly requests for each top user
	weeklyRequests, err := WeeklyUserRequests()
	if err != nil {
		return nil, err
	}
	for _, userRequests := range weeklyRequests {
		for i, r := range requests {
			if userRequests.APIKey == r.APIKey {
				requests[i].WeeklyRequests = userRequests.Count
			}
		}
	}

	return requests, nil
}
