package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

type UserRow struct {
	database.UserRow
}

type RequestRow struct {
	database.RequestRow
}

type MonitorRow struct {
	database.MonitorRow
}

type PingsRow struct {
	database.PingsRow
}

func makeBackupDirectory() string {
	dirname := fmt.Sprintf("backup-%s", time.Now().Format("2006-01-02 15:04:05"))
	if err := os.Mkdir(dirname, os.ModeDir); err != nil {
		panic(err)
	}

	return dirname
}

func getAllUsers() []UserRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM users;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var users []UserRow
	for rows.Next() {
		var user UserRow
		err := rows.Scan(&user.UserID, &user.APIKey, &user.CreatedAt)
		if err == nil {
			users = append(users, user)
		}
	}

	return users
}

func getAllRequests() []RequestRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM requests;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var requests []RequestRow
	for rows.Next() {
		var request RequestRow
		err := rows.Scan(&request.RequestID, &request.APIKey, &request.Path, &request.Hostname, &request.IPAddress, &request.Location, &request.UserAgent, &request.Method, &request.Status, &request.ResponseTime, &request.Framework, &request.CreatedAt)
		if err == nil {
			requests = append(requests, request)
		}
	}

	return requests
}

func getAllMonitors() []MonitorRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM monitor;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var monitors []MonitorRow
	for rows.Next() {
		var monitor MonitorRow
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, monitor)
		}
	}

	return monitors
}

func getAllPings() []PingsRow {
	db := database.OpenDBConnection()

	query := "SELECT * FROM pings;"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	var pings []PingsRow
	for rows.Next() {
		var ping PingsRow
		err := rows.Scan(&ping.APIKey, &ping.URL, &ping.ResponseTime, &ping.Status, &ping.CreatedAt)
		if err == nil {
			pings = append(pings, ping)
		}
	}

	return pings
}

type Row interface {
	GetAPIKey() string
}

func (row UserRow) GetAPIKey() string {
	return row.APIKey
}

func (row RequestRow) GetAPIKey() string {
	return row.APIKey
}

func (row MonitorRow) GetAPIKey() string {
	return row.APIKey
}

func (row PingsRow) GetAPIKey() string {
	return row.APIKey
}

func GroupByUser[T Row](rows []T) map[string][]T {
	data := make(map[string][]T)
	for _, row := range rows {
		apiKey := row.GetAPIKey()
		if _, ok := data[apiKey]; ok {
			data[apiKey] = make([]T, 0)
		}
		data[apiKey] = append(data[apiKey], row)
	}
	return data
}

func makeUsersBackup(dirname string, users map[string][]UserRow) {
	if err := os.Mkdir(fmt.Sprintf("%s/users", dirname), os.ModeDir); err != nil {
		panic(err)
	}

	for apiKey, rows := range users {
		file, err := os.Create(fmt.Sprintf("%s/users/%s.csv", dirname, apiKey))
		defer file.Close()
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(file)
		defer w.Flush()

		data := [][]string{{"user_id", "api_key", "created_at"}}
		for _, row := range rows {
			row := []string{row.UserID, row.APIKey, row.CreatedAt.Format(time.RFC3339)}
			data = append(data, row)
		}
		w.WriteAll(data)
	}
}

func makeRequestsBackup(dirname string, requests map[string][]RequestRow) {
	if err := os.Mkdir(fmt.Sprintf("%s/requests", dirname), os.ModeDir); err != nil {
		panic(err)
	}

	for apiKey, rows := range requests {
		file, err := os.Create(fmt.Sprintf("%s/requests/%s.csv", dirname, apiKey))
		defer file.Close()
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(file)
		defer w.Flush()

		data := [][]string{{"request_id", "api_key", "path", "hostname", "ip_address", "location", "user_agent", "method", "status", "response_time", "framework", "created_at"}}
		for _, row := range rows {
			row := []string{strconv.Itoa(row.RequestID), row.APIKey, row.Path, row.Hostname, row.IPAddress, row.Location, row.UserAgent, strconv.FormatInt(int64(row.Method), 10), strconv.FormatInt(int64(row.Status), 10), strconv.FormatInt(int64(row.ResponseTime), 10), strconv.FormatInt(int64(row.Framework), 10), row.CreatedAt.Format(time.RFC3339)}
			data = append(data, row)
		}
		w.WriteAll(data)
	}
}

func makeMonitorsBackup(dirname string, monitors map[string][]MonitorRow) {
	if err := os.Mkdir(fmt.Sprintf("%s/monitors", dirname), os.ModeDir); err != nil {
		panic(err)
	}

	for apiKey, rows := range monitors {
		file, err := os.Create(fmt.Sprintf("%s/monitors/%s.csv", dirname, apiKey))
		defer file.Close()
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(file)
		defer w.Flush()

		data := [][]string{{"api_key", "url", "secure", "ping", "created_at"}}
		for _, row := range rows {
			row := []string{row.APIKey, row.URL, strconv.FormatBool(row.Secure), strconv.FormatBool(row.Ping), row.CreatedAt.Format(time.RFC3339)}
			data = append(data, row)
		}
		w.WriteAll(data)
	}
}

func makePingsBackup(dirname string, pings map[string][]PingsRow) {
	if err := os.Mkdir(fmt.Sprintf("%s/pings", dirname), os.ModeDir); err != nil {
		panic(err)
	}

	for apiKey, rows := range pings {
		file, err := os.Create(fmt.Sprintf("%s/pings/%s.csv", dirname, apiKey))
		defer file.Close()
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(file)
		defer w.Flush()

		data := [][]string{{"api_key", "url", "response_time", "status", "created_at"}}
		for _, row := range rows {
			row := []string{row.APIKey, row.URL, strconv.Itoa(row.ResponseTime), strconv.Itoa(row.Status), row.CreatedAt.Format(time.RFC3339)}
			data = append(data, row)
		}
		w.WriteAll(data)
	}
}

func BackupDatabase() {
	dirname := makeBackupDirectory()

	requests := getAllRequests()
	groupedRequests := GroupByUser(requests)
	fmt.Println(groupedRequests)

	users := getAllUsers()
	monitors := getAllMonitors()
	pings := getAllPings()
}

func RestoreFromBackup() {

}
