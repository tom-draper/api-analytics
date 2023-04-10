package main

import (
	"github.com/tom-draper/api-analytics/server/database"
)

func getAllUsers() []database.UserRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM users;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var users []database.UserRow
	for rows.Next() {
		var user database.UserRow
		err := rows.Scan(&user.UserID, &user.APIKey, &user.CreatedAt)
		if err == nil {
			users = append(users, user)
		}
	}

	return users
}

func getAllRequests() []database.RequestRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var requests []database.RequestRow
	for rows.Next() {
		var request database.RequestRow
		err := rows.Scan(&request.RequestID, &request.APIKey, &request.Path, &request.Hostname, &request.IPAddress, &request.Location, &request.UserAgent, &request.Method, &request.Status, &request.ResponseTime, &request.Framework, &request.CreatedAt)
		if err == nil {
			requests = append(requests, request)
		}
	}

	return requests
}

func getAllMonitors() []database.MonitorRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM monitor;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var monitors []database.MonitorRow
	for rows.Next() {
		var monitor database.MonitorRow
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, monitor)
		}
	}

	return monitors
}

func getAllPings() []database.PingsRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM pings;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var pings []database.PingsRow
	for rows.Next() {
		var ping database.PingsRow
		err := rows.Scan(&ping.APIKey, &ping.URL, &ping.ResponseTime, &ping.Status, &ping.CreatedAt)
		if err == nil {
			pings = append(pings, ping)
		}
	}

	return pings
}

func BackupDatabase() {
	requests := getAllRequests()
	users := getAllUsers()
	monitors := getAllMonitors()
	pings := getAllPings()
}

func RestoreFromBackup() {

}
