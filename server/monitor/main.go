package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/database"
)

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

func getMethod(ping bool) string {
	if ping {
		return "HEAD"
	}
	return "GET"
}

func ping(client http.Client, url string, secure bool, ping bool) (int, time.Duration, error) {
	// Determine the method (HEAD, GET, etc.) based on whether 'ping' is true
	method := getMethod(ping)

	// Create a new HTTP request
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %v", err)
	}

	// Start the timer before sending the request
	start := time.Now()

	// Send the HTTP request
	response, err := client.Do(request)
	elapsed := time.Since(start)

	if err != nil {
		return 0, elapsed, fmt.Errorf("failed to ping URL %s: %v", url, err)
	}

	// Ensure the response body is closed after reading the response
	defer response.Body.Close()

	// Return the status code, response time, and no error
	return response.StatusCode, elapsed, nil
}

func getMonitoredURLs(conn *pgx.Conn) []MonitorRow {
	query := "SELECT * FROM monitor;"
	rows, err := conn.Query(context.Background(), query)
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

func deleteExpiredPings(conn *pgx.Conn) error {
	// Calculate the timestamp 60 days ago
	expiryTime := time.Now().Add(-60 * 24 * time.Hour).UTC()

	// Define the query with a parameter placeholder
	query := "DELETE FROM pings WHERE created_at < $1"

	// Execute the query with the expiry time as the parameter
	_, err := conn.Exec(context.Background(), query, expiryTime)
	if err != nil {
		return fmt.Errorf("failed to delete expired pings: %v", err)
	}

	return nil
}

func uploadPings(pings []PingsRow, conn *pgx.Conn) error {
	// Return early if there are no pings to insert
	if len(pings) == 0 {
		return nil
	}

	// Build the query with placeholders
	var query strings.Builder
	query.WriteString("INSERT INTO pings (api_key, url, response_time, status, created_at) VALUES ")

	// Collect query arguments and placeholders
	args := make([]interface{}, 0, len(pings)*5) // 5 fields per ping
	placeholderID := 1

	for i, ping := range pings {
		if i > 0 {
			query.WriteString(",")
		}

		// Add placeholders like ($1, $2, $3, $4, $5) for each ping
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", placeholderID, placeholderID+1, placeholderID+2, placeholderID+3, placeholderID+4))

		// Append the values to the args slice
		args = append(args, ping.APIKey, ping.URL, ping.ResponseTime, ping.Status, ping.CreatedAt.UTC())

		// Increment the placeholder ID by 5 for the next set of values
		placeholderID += 5
	}

	query.WriteString(";")

	// Execute the query with the collected arguments
	_, err := conn.Exec(context.Background(), query.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to upload pings: %v", err)
	}

	return nil
}

func shuffle(monitored []MonitorRow) {
	rand.Shuffle(len(monitored), func(i, j int) {
		monitored[i], monitored[j] = monitored[j], monitored[i]
	})
}

func pingMonitored(monitored []MonitorRow) []PingsRow {
	client := getClient()
	var wg sync.WaitGroup
	var mu sync.Mutex

	pings := make([]PingsRow, 0, len(monitored)) // Preallocate slice with expected length

	for _, m := range monitored {
		wg.Add(1)
		go func(m MonitorRow) {
			defer wg.Done() 

			status, elapsed, err := ping(client, m.URL, m.Secure, m.Ping)
			if err != nil {
				fmt.Println(err) 
				return
			}

			pingResult := PingsRow{
				APIKey:       m.APIKey,
				URL:          m.URL,
				ResponseTime: int(elapsed.Milliseconds()),
				Status:       status,
				CreatedAt:    time.Now(),
			}

			mu.Lock()
			pings = append(pings, pingResult)
			mu.Unlock()
		}(m)
	}

	wg.Wait()
	return pings
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
	database.LoadConfig()
	conn, err := database.NewConnection()
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	monitored := getMonitoredURLs(conn)
	// Shuffle URLs to ping to avoid a page looking consistently slow or fast
	// due to cold starts or caching
	shuffle(monitored)

	pings := pingMonitored(monitored)
	err = uploadPings(pings, conn)
	if err != nil {
		panic(err)
	}
	err = deleteExpiredPings(conn)
	if err != nil {
		panic(err)
	}
}
