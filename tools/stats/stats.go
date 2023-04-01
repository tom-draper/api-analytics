package main

import "github.com/tom-draper/api-analytics/server/lib/database"

func UserUsage() {
	db := database.OpenDBConnection()

	query := "SELECT *, COUNT(*) OVER(PARTITION BY api_key) count FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		monitor := new(MonitorRow)
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, *monitor)
		}
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
