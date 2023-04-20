package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
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

func createUserTable(dbName string) {
	db := database.OpenDBConnection()
	_, err := db.Exec("DROP TABLE IF EXISTS users;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE users (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at timestamptz, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createRequestsTable(dbName string) {
	db := database.OpenDBConnection()
	_, err := db.Exec("DROP TABLE IF EXISTS requests;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE requests (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at timestamptz, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createMonitorTable(dbName string) {
	db := database.OpenDBConnection()
	_, err := db.Exec("DROP TABLE IF EXISTS monitor;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE monitor (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at timestamptz, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func createPingsTable(dbName string) {
	db := database.OpenDBConnection()
	_, err := db.Exec("DROP TABLE IF EXISTS pings;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE pings (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at timestamptz, PRIMARY KEY (api_key));")
	if err != nil {
		panic(err)
	}
}

func RestoreUser(dirname string, dbName string) {
	rows := readTable(dirname, "users")
	createUserTable(dbName)
}

func RestoreRequests(dirname string, dbName string) {
	rows := readTable(dirname, "requests")
	createRequestsTable(dbName)
}

func RestoreMonitor(dirname string, dbName string) {
	rows := readTable(dirname, "monitor")
	createMonitorTable(dbName)
}

func RestorePings(dirname string, dbName string) {
	rows := readTable(dirname, "pings")
	createPingsTable(dbName)
}

func Restore(dirname string, dbName string) {
	// Database with dbName assumed already exists
	unzipBackup(dirname)
	RestoreRequests(dirname, dbName)
	RestoreUser(dirname, dbName)
	RestoreMonitor(dirname, dbName)
	RestorePings(dirname, dbName)
}
