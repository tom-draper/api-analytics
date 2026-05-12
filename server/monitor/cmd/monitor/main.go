package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/monitor/internal/config"
	"github.com/tom-draper/api-analytics/server/monitor/internal/log"
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

func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := log.Init(); err != nil {
		log.Error(fmt.Sprintf("failed to initialize log file: %v", err))
	}

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("configuration error: %v", err))
	}

	// Initialize database connection pool
	db, err := database.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create database connection pool: %v", err))
	}
	defer db.Close()
	log.Info("database connection pool initialized")

	monitored, err := getMonitoredURLs(ctx, db)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to fetch monitored URLs: %v", err))
	}
	if len(monitored) == 0 {
		log.Info("no monitored URLs found")
		return
	}

	// Shuffle URLs to ping to avoid a page looking consistently slow or fast
	// due to cold starts or caching
	shuffle(monitored)

	pings := pingMonitored(monitored)
	log.Info(fmt.Sprintf("completed %d pings", len(pings)))

	err = uploadPings(ctx, db, pings)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to upload pings: %v", err))
	}
	log.Info(fmt.Sprintf("uploaded %d pings", len(pings)))

	err = deleteExpiredPings(ctx, db)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to delete expired pings: %v", err))
	}
	log.Info("deleted expired pings")
}

func getMonitoredURLs(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	query := "SELECT * FROM monitor;"
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored URLs: %w", err)
	}
	defer rows.Close()

	// Read monitors into list to return
	monitors := make([]MonitorRow, 0)
	for rows.Next() {
		monitor := new(MonitorRow)
		err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt)
		if err == nil {
			monitors = append(monitors, *monitor)
		} else {
			log.Error(fmt.Sprintf("failed to scan monitor row: %v", err))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monitored URLs: %w", err)
	}

	return monitors, nil
}

func uploadPings(ctx context.Context, db *database.DB, pings []PingsRow) error {
	// Return early if there are no pings to insert
	if len(pings) == 0 {
		return nil
	}

	// Use COPY protocol for efficient bulk insert
	_, err := db.Pool.CopyFrom(
		ctx,
		pgx.Identifier{"pings"},
		[]string{"api_key", "url", "response_time", "status", "created_at"},
		pgx.CopyFromSlice(len(pings), func(i int) ([]any, error) {
			ping := pings[i]
			return []any{
				ping.APIKey,
				ping.URL,
				ping.ResponseTime,
				ping.Status,
				ping.CreatedAt.UTC(),
			}, nil
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to upload pings: %v", err)
	}

	return nil
}

func deleteExpiredPings(ctx context.Context, db *database.DB) error {
	// Calculate the timestamp 60 days ago
	expiryTime := time.Now().Add(-60 * 24 * time.Hour).UTC()

	// Define the query with a parameter placeholder
	query := "DELETE FROM pings WHERE created_at < $1;"

	// Execute the query with the expiry time as the parameter
	_, err := db.Pool.Exec(ctx, query, expiryTime)
	if err != nil {
		return fmt.Errorf("failed to delete expired pings: %v", err)
	}

	return nil
}

func pingMonitored(monitored []MonitorRow) []PingsRow {
	client := getClient()
	var wg sync.WaitGroup
	var mu sync.Mutex

	pings := make([]PingsRow, 0, len(monitored))

	for _, m := range monitored {
		wg.Add(1)
		go func(m MonitorRow) {
			defer wg.Done()

			status, elapsed, err := ping(client, m.URL, m.Secure, m.Ping)
			if err != nil {
				log.Error(err.Error())
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

func getMethod(ping bool) string {
	if ping {
		return "HEAD"
	}
	return "GET"
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

func shuffle(monitored []MonitorRow) {
	rand.Shuffle(len(monitored), func(i, j int) {
		monitored[i], monitored[j] = monitored[j], monitored[i]
	})
}
