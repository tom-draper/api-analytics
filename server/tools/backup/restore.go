package main

import (
	"archive/zip"
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

func readTable(dirname string, table string) []Row {
	rows := []Row{}
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

type Row interface {
	Parse([]string) error
}

type UserRow struct {
	database.UserRow
}

func (r *UserRow) Parse(row []string) error {
	r.UserID = row[0]
	r.APIKey = row[1]
	createdAt, err := time.Parse(row[2], time.RFC3339)
	r.CreatedAt = createdAt
	return err
}

type RequestRow struct {
	database.RequestRow
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

type MonitorRow struct {
	database.MonitorRow
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

type PingsRow struct {
	database.PingsRow
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

func parseRows(file string, table string) ([]Row, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	rows := []Row{}
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return rows, err
		}

		var r Row
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

func createUsersTable(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS users;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE users (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createRequestsTable(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS requests;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE requests (request_id INTEGER, api_key UUID NOT NULL, path VARCHAR(255) NOT NULL, hostname VARCHAR(255), ip_address CIDR, location CHAR(2), user_agent VARCHAR(255), method SMALLINT NOT NULL, status SMALLINT NOT NULL, response_time SMALLINT NOT NULL, framework SMALLINT NOT NULL, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createMonitorTable(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS monitor;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE monitor (api_key UUID NOT NULL, url VARCHAR(255) NOT NULL, secure BOOLEAN, PING BOOLEAN, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createPingsTable(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS pings;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE pings (api_key UUID NOT NULL, url VARCHAR(255) NOT NULL, response_time INTEGER, status SMALLINT, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func insertUserData(db *sql.DB, rows []UserRow) {
	if len(rows) == 0 {
		return
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO users (api_key, user_id, created_at) VALUES")
	for i, user := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', '%s')", user.APIKey, user.UserID, user.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func insertRequestData(db *sql.DB, rows []UserRow) {
	blockSize := 1000
	if len(rows) == 0 {
		return
	}

	var query bytes.Buffer
	for i, request := range result {
		if i % blockSize == 0 {
			query = bytes.Buffer{}
			query.WriteString("INSERT INTO requests (api_key, method, created_at, path, status, response_time, framework, user_agent, hostname, ip_address, location) VALUES")
		}
		if i % blockSize > 0 {
			query.WriteString(",")
		} 

		fmtIPAddress = nullInsertString(request.IPAddress)
		fmtHostname = nullInsertString(request.Hostname)
		fmtLocation = nullInsertString(request.Location)
		fmtUserAgent = nullInsertString(request.UserAgent)

		query.WriteString(fmt.Sprintf(" ('%s', %d, '%s', '%s', %d, %d, %d, %s, %s, %s)", request.APIKey, request.Method, request.CreatedAt.UTC().Format(time.RFC3339), request.Path, request.Status, request.ResponseTime, request.Framework, fmtUserAgent, fmtHostname, fmtIPAddress, fmtLocation))

		if (i + 1) % blockSize == 0 || i == len(result) - 1 {
			query.WriteString(";")

			fmt.Println("Write to database")

			db = database.OpenDBConnection()
			_, err := db.Query(query.String())
			if err != nil {
				panic(err)
			}
			db.Close()
			time.Sleep(8 * time.Second)  // Pause to avoid exceeding memory space
		}
	}
}

func nullInsertString(value sql.NullString) string {
	if value.String == "" {
		return "NULL"
	} else {
		return fmt.Sprintf("'%s'", value.String)
	}
}

func insertMonitorData(db *sql.DB, rows []MonitorRow) {
	if len(rows) == 0 {
		return
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO monitor (api_key, url, secure, ping, created_at) VALUES")
	for i, user := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %t, %t, '%s')", monitor.APIKey, monitor.URL, monitor.Secure, monitor.Ping, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func insertPingsData(db *sql.DB, rows []PingsRow) {
	if len(rows) == 0 {
		return
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, monitor := range result {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", monitor.APIKey, monitor.URL, monitor.ResponseTime, monitor.Status, monitor.CreatedAt.UTC().Format(time.RFC3339)))
	}
	query.WriteString(";")

	_, err = db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func RestoreUsers(dirname string, dbName string) {
	rows := readTable(dirname, "users")
	db := database.OpenDBConnectionNamed(dbName)
	createUsersTable(db)
	insertUserData(db, rows)
}

func RestoreRequests(dirname string, dbName string) {
	rows := readTable(dirname, "requests")
	db := database.OpenDBConnectionNamed(dbName)
	createRequestsTable(db)
	insertRequestsData(db, rows)
}

func RestoreMonitor(dirname string, dbName string) {
	rows := readTable(dirname, "monitor")
	db := database.OpenDBConnectionNamed(dbName)
	createMonitorTable(db)
	insertMonitorData(db, rows)
}

func RestorePings(dirname string, dbName string) {
	rows := readTable(dirname, "pings")
	db := database.OpenDBConnectionNamed(dbName)
	createPingsTable(db)
	insertPingsData(db, rows)
}

func Restore(dirname string, dbName string) {
	// Database with dbName assumed already exists
	unzipBackup(dirname)
	RestoreRequests(dirname, dbName)
	RestoreUsers(dirname, dbName)
	RestoreMonitor(dirname, dbName)
	RestorePings(dirname, dbName)
}
