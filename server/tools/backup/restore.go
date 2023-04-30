package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

func unzipBackup(dirname string) {
	archive, err := zip.OpenReader(dirname + ".zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		if f.FileInfo().IsDir() {
			os.MkdirAll(f.Name, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func readTable(dirname string, table string) []RestoreRow {
	rows := []RestoreRow{}
	filepath.Walk(filepath.Join(dirname, table), func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".csv") {
			userRows, err := parseRows(info.Name(), table)
			if err != nil {
				panic(err)
			}
			rows = append(rows, userRows...)
		}
		return nil
	})
	return rows
}

type RestoreRow interface {
	Parse([]string) error
}

type RestoreRows interface {
	Upload(*sql.DB) error
}

type UserRows []UserRow

func (r *UserRow) Parse(row []string) error {
	r.UserID = row[0]
	r.APIKey = row[1]
	createdAt, err := time.Parse(row[2], time.RFC3339)
	r.CreatedAt = createdAt
	return err
}

func (r UserRows) Upload(db *sql.DB) error {
	return insertUserData(db, r)
}

func (r *RequestRow) Parse(row []string) error {
	var err error
	if r.RequestID, err = strconv.Atoi(row[0]); err != nil {
		return err
	}
	r.APIKey = row[1]
	r.Path = row[2]
	r.Hostname = sql.NullString{row[3], row[3] != ""}
	r.IPAddress = sql.NullString{row[4], row[4] != ""}
	r.Location = sql.NullString{row[5], row[5] != ""}
	r.UserAgent = sql.NullString{row[6], row[6] != ""}
	method, err := strconv.ParseInt(row[7], 10, 16)
	if err != nil {
		return err
	}
	r.Method = int16(method)
	status, err := strconv.ParseInt(row[8], 10, 16)
	if err != nil {
		return err
	}
	r.Status = int16(status)
	responseTime, err := strconv.ParseInt(row[9], 10, 16)
	if err != nil {
		return err
	}
	r.ResponseTime = int16(responseTime)
	framework, err := strconv.ParseInt(row[10], 10, 16)
	if err != nil {
		return err
	}
	r.Framework = int16(framework)
	createdAt, err := time.Parse(row[11], time.RFC3339)
	r.CreatedAt = createdAt
	return err
}

func (r *MonitorRow) Parse(row []string) error {
	r.APIKey = row[0]
	r.URL = row[1]
	secure, err := strconv.ParseBool(row[2])
	if err != nil {
		return err
	}
	r.Secure = secure
	ping, err := strconv.ParseBool(row[3])
	if err != nil {
		return err
	}
	r.Ping = ping
	createdAt, err := time.Parse(row[4], time.RFC3339)
	r.CreatedAt = createdAt
	return err
}

func (r *PingsRow) Parse(row []string) error {
	r.APIKey = row[0]
	r.URL = row[1]
	responseTime, err := strconv.Atoi(row[2])
	if err != nil {
		return err
	}
	r.ResponseTime = responseTime
	status, err := strconv.Atoi(row[3])
	if err != nil {
		return err
	}
	r.Status = status
	createdAt, err := time.Parse(row[4], time.RFC3339)
	r.CreatedAt = createdAt
	return err
}

func parseRows(file string, table string) ([]RestoreRow, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	rows := []RestoreRow{}
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return rows, err
		}

		var r RestoreRow
		switch table {
		case "requests":
			r = &RequestRow{}
		case "users":
			r = &UserRow{}
		case "monitor":
			r = &MonitorRow{}
		case "pings":
			r = &PingsRow{}
		}
		err = r.Parse(row)
		if err != nil {
			panic(err)
		}
		rows = append(rows, r)
	}
	return rows, nil
}

func insertUserData(db *sql.DB, rows []UserRow) error {
	if len(rows) == 0 {
		return nil
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO users (api_key, user_id, created_at) VALUES")
	for i, user := range rows {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', '%s')", user.APIKey, user.UserID, user.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err := db.Query(query.String())
	return err
}

func insertRequestsData(db *sql.DB, rows []RequestRow) error {
	blockSize := 1000
	if len(rows) == 0 {
		return nil
	}

	var query bytes.Buffer
	for i, request := range rows {
		if i%blockSize == 0 {
			query = bytes.Buffer{}
			query.WriteString("INSERT INTO requests (api_key, method, created_at, path, status, response_time, framework, user_agent, hostname, ip_address, location) VALUES")
		}
		if i%blockSize > 0 {
			query.WriteString(",")
		}

		fmtIPAddress := nullInsertString(request.IPAddress)
		fmtHostname := nullInsertString(request.Hostname)
		fmtLocation := nullInsertString(request.Location)
		fmtUserAgent := nullInsertString(request.UserAgent)

		query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, %s, %s, %s)", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, fmtUserAgent, fmtHostname, fmtIPAddress, fmtLocation))

		if (i+1)%blockSize == 0 || i == len(rows)-1 {
			query.WriteString(";")

			fmt.Println("Write to database")

			db = database.OpenDBConnection()
			_, err := db.Query(query.String())
			db.Close()
			if err != nil {
				return err
			}
			time.Sleep(8 * time.Second) // Pause to avoid exceeding memory space
		}
	}
	return nil
}

func nullInsertString(value sql.NullString) string {
	if value.String == "" {
		return "NULL"
	} else {
		return fmt.Sprintf("'%s'", value.String)
	}
}

func insertMonitorData(db *sql.DB, rows []MonitorRow) error {
	if len(rows) == 0 {
		return nil
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES")
	for i, monitor := range rows {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %t, %t, '%s')", monitor.APIKey, monitor.URL, monitor.Secure, monitor.Ping, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err := db.Query(query.String())
	return err
}

func insertPingsData(db *sql.DB, rows []PingsRow) error {
	if len(rows) == 0 {
		return nil
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, monitor := range rows {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", monitor.APIKey, monitor.URL, monitor.ResponseTime, monitor.Status, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err := db.Query(query.String())
	return err
}

func RestoreUsers(dirname string, dbName string) {
	rows := readTable(dirname, "users")
	db := database.OpenDBConnectionNamed(dbName)
	database.CreateUsersTable(db)
	insertRequestsData(db, rows)
	switch rows.(type) {
	case UserRows:
		rows.Upload(db)
	}
}

func RestoreRequests(dirname string, dbName string) {
	rows := readTable(dirname, "requests")
	db := database.OpenDBConnectionNamed(dbName)
	database.CreateRequestsTable(db)
	database.InsertRequestsData(db, rows)
}

func RestoreMonitor(dirname string, dbName string) {
	rows := readTable(dirname, "monitor")
	db := database.OpenDBConnectionNamed(dbName)
	database.CreateMonitorTable(db)
	database.InsertMonitorData(db, rows)
}

func RestorePings(dirname string, dbName string) {
	rows := readTable(dirname, "pings")
	db := database.OpenDBConnectionNamed(dbName)
	database.CreatePingsTable(db)
	switch rows.(type) {
	case []PingsRow:
		database.InsertPingsData(db, rows)
	}
}

func Restore(dirname string, dbName string) {
	// Database with dbName assumed already exists
	unzipBackup(dirname)
	RestoreRequests(dirname, dbName)
	RestoreUsers(dirname, dbName)
	RestoreMonitor(dirname, dbName)
	RestorePings(dirname, dbName)
}
