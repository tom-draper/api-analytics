package main

import "github.com/tom-draper/api-analytics/server/lib/database"

func UserUsage() {
	db := database.OpenDBConnection()

	query := "SELECT *, COUNT(*) OVER(PARTITION BY api_key) count FROM requests;"
	_, err := db.Query(query)
	if err != nil {
		panic(err)
	}
}

func TopUsers() {

}

func DeleteUser() {

}

func DeleteUserRequests() {

}

func DailyUsage() {

}
