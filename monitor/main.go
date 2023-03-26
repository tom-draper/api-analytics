package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
)

func getDBLogin() (string, string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	database := os.Getenv("POSTGRES_DATABASE")
	username := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	return database, username, password
}

func OpenDBConnection() *sql.DB {
	database, username, password := getDBLogin()
	args := fmt.Sprintf("host=%s port=%d dbname=%s user='%s' password=%s sslmode=%s", "localhost", 5432, database, username, password, "disable")

	db, err := sql.Open("postgres", args)
	if err != nil {
		panic(err)
	}

	return db
}

func getURL(domain string, secure bool) string {
	var url string
	if secure {
		url = "https://" + domain
	} else {
		url = "http://" + domain
	}
	return url
}

func getMethod(ping bool) string {
	var method string
	if ping {
		method = "HEAD"
	} else {
		method = "GET"
	}
	return method
}

func ping(client http.Client, domain string, secure bool, ping bool) (int, time.Duration, error) {
	url := getURL(domain, secure)
	method := getMethod(ping)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, time.Duration(0), err
	}

	// Make request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, time.Duration(0), err
	}
	elapsed := time.Since(start)

	resp.Body.Close()

	return resp.StatusCode, elapsed, nil
}

type MonitorRow struct {
	APIKey    string    `json:"api_key"`
	URL       string    `json:"url" `
	Ping      bool      `json:"ping"`
	Secure    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func getMonitoredURLs(db *sql.DB) []MonitorRow {
	query := fmt.Sprintf("SELECT * FROM monitor;")
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	// Read monitors into list to return
	monitors := make([]MonitorRow, 0)
	for rows.Next() {
		monitor := new(MonitorRow)
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, *monitor)
		}
	}

	return monitors
}

type PingRow struct {
	APIKey       string    `json:"api_key"`
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func deleteOldPings(db *sql.DB) {
	query := fmt.Sprintf("DELETE FROM pings WHERE created_at < '%s';", time.Now().Add(-60*24*time.Hour).UTC().Format("2006-01-02T15:04:05-0700"))
	_, err := db.Query(query)
	if err != nil {
		panic(err)
	}
}

func uploadPings(pings []PingRow, db *sql.DB) {
	var query bytes.Buffer
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES")
	for i, ping := range pings {
		if i > 0 {
			query.WriteString(",")
		}
		query.WriteString(fmt.Sprintf(" ('%s', '%s', %d, %d, '%s')", ping.APIKey, ping.URL, ping.ResponseTime, ping.Status, ping.CreatedAt.UTC().Format("2006-01-02T15:04:05-0700")))
	}
	query.WriteString(";")

	_, err := db.Query(query.String())
	if err != nil {
		panic(err)
	}
}

func pingMonitored(monitored []MonitorRow, client http.Client, db *sql.DB) {
	var pings []PingRow
	for _, m := range monitored {
		status, elapsed, err := ping(client, m.URL, m.Secure, m.Ping)
		if err != nil {
			fmt.Println(err)
		}
		ping := PingRow{
			APIKey:       m.APIKey,
			URL:          m.URL,
			ResponseTime: int(elapsed.Milliseconds()),
			Status:       status,
			CreatedAt:    time.Now(),
		}
		pings = append(pings, ping)
	}

	uploadPings(pings, db)
	deleteOldPings(db)
}

func getClient() http.Client {
	dialer := net.Dialer{Timeout: 2 * time.Second}
	var client = http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
	return client
}

func main() {
	db := OpenDBConnection()

	monitored := getMonitoredURLs(db)

	client := getClient()
	pingMonitored(monitored, client, db)
}
