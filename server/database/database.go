package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
)

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
	UserAgent    sql.NullString `json:"user_agent"`
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
	username, password := getDBLogin()
	database := os.Getenv("POSTGRES_DATABASE")
	return openDBConnection(database, username, password)
}

func OpenDBConnectionNamed(database string) *sql.DB {
	username, password := getDBLogin()
	return openDBConnection(database, username, password)
}

func CreateUsersTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE users (user_id UUID NOT NULL, api_key UUID NOT NULL, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	return err
}

func CreateRequestsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS requests;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE requests (request_id INTEGER, api_key UUID NOT NULL, path VARCHAR(255) NOT NULL, hostname VARCHAR(255), ip_address CIDR, location CHAR(2), user_agent VARCHAR(255), method SMALLINT NOT NULL, status SMALLINT NOT NULL, response_time SMALLINT NOT NULL, framework SMALLINT NOT NULL, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	return err
}

func CreateMonitorTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS monitor;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE monitor (api_key UUID NOT NULL, url VARCHAR(255) NOT NULL, secure BOOLEAN, PING BOOLEAN, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	return err
}

func CreatePingsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS pings;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE pings (api_key UUID NOT NULL, url VARCHAR(255) NOT NULL, response_time INTEGER, status SMALLINT, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY (api_key));")
	return err
}

func DeleteUser(apiKey string) error {
	db := OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM user WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func DeleteRequests(apiKey string) error {
	db := OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM requests WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func DeleteMonitors(apiKey string) error {
	db := OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM monitor WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func DeletePings(apiKey string) error {
	db := OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM pings WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}
