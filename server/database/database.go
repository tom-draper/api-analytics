package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
)

var dbUsername string
var dbPassword string
var dbName string

type UserRow struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

type RequestRow struct {
	RequestID    int            `json:"request_id"`
	APIKey       string         `json:"api_key"`
	Path         string         `json:"path"`
	Hostname     sql.NullString `json:"hostname"`
	IPAddress    sql.NullString `json:"ip_address"`
	Location     sql.NullString `json:"location"`
	UserAgentID  sql.NullInt64  `json:"user_agent_id"`
	Method       int16          `json:"method"`
	Status       int16          `json:"status"`
	ResponseTime int16          `json:"response_time"`
	Framework    int16          `json:"framework"`
	CreatedAt    time.Time      `json:"created_at"`
}

type MonitorRow struct {
	APIKey    string    `json:"api_key"`
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

type PingsRow struct {
	APIKey       string    `json:"api_key"`
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserAgentsRow struct {
	UserAgentID int64  `json:"user_agent_id"`
	UserAgent   string `json:"user_agent"`
}

func getDBLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	username := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	return username, password
}

func openDBConnection(database string, username string, password string) *sql.DB {
	args := fmt.Sprintf("host=%s port=%d dbname=%s user='%s' password=%s sslmode=%s", "localhost", 5432, database, username, password, "disable")

	db, err := sql.Open("postgres", args)
	if err != nil {
		panic(err)
	}

	return db
}

func OpenDBConnection() *sql.DB {
	if dbUsername == "" || dbPassword == "" || dbName == "" {
		username, password := getDBLogin()
		database := os.Getenv("POSTGRES_DATABASE")
		dbUsername = username
		dbPassword = password
		dbName = database
		return openDBConnection(database, username, password)
	} else {
		return openDBConnection(dbName, dbUsername, dbPassword)
	}
}

func OpenDBConnectionNamed(database string) *sql.DB {
	if dbUsername == "" || dbPassword == "" {
		username, password := getDBLogin()
		return openDBConnection(database, username, password)
	} else {
		return openDBConnection(database, dbUsername, dbPassword)
	}
}

func DeleteUser(apiKey string) error {
	db := OpenDBConnection()
	defer db.Close()

	query := "DELETE FROM users WHERE api_key = $1;"
	_, err := db.Query(query, apiKey)
	return err
}

func DeleteRequests(apiKey string) error {
	db := OpenDBConnection()
	defer db.Close()

	query := "DELETE FROM requests WHERE api_key = $1;"
	_, err := db.Query(query, apiKey)
	return err
}

func DeleteMonitors(apiKey string) error {
	db := OpenDBConnection()
	defer db.Close()

	query := "DELETE FROM monitor WHERE api_key = $1;"
	_, err := db.Query(query, apiKey)
	return err
}

func DeletePings(apiKey string) error {
	db := OpenDBConnection()
	defer db.Close()

	query := "DELETE FROM pings WHERE api_key = $1;"
	_, err := db.Query(query, apiKey)
	return err
}
