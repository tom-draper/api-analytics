package main

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/database"
)

func TotalRequests() {
	db := database.OpenDBConnection()

	query := "SELECT COUNT(*) AS count FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var count int
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		panic(err)
	}

	fmt.Println(count)
}

type UserRequestCount struct {
	APIKey string
	Count  string
}

func TopUsers() {
	db := database.OpenDBConnection()

	query := "SELECT api_key, COUNT(*) AS count FROM requests GROUP BY api_key ORDER BY count DESC;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var requests []UserRequestCount
	for rows.Next() {
		request := new(UserRequestCount)
		err := rows.Scan(&request.APIKey, &request.Count)
		if err == nil {
			requests = append(requests, *request)
		}
	}

	fmt.Println(requests)
}

func NewUsers() {
	db := database.OpenDBConnection()

	query := "SELECT * FROM users WHERE created_at >= NOW() - interval '7 days';"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var users []database.UserRow
	for rows.Next() {
		user := new(database.UserRow)
		err := rows.Scan(&user.APIKey, &user.UserID, &user.CreatedAt)
		if err == nil {
			users = append(users, *user)
		}
	}

	fmt.Println(users)
}

func DailyUsage() {
	Usage(1)
}

func Usage(days int) {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("SELECT api_key, COUNT(*) AS count FROM requests where created_at >= NOW() - interval '%d day' GROUP BY api_key ORDER BY count DESC;", days)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var requests []UserRequestCount
	for rows.Next() {
		request := new(UserRequestCount)
		err := rows.Scan(&request.APIKey, &request.Count)
		if err == nil {
			requests = append(requests, *request)
		}
	}

	fmt.Println(requests)
}
