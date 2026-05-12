package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	if err := log.Init(); err != nil {
		log.Error(fmt.Sprintf("failed to initialize log file: %v", err))
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("configuration error: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to initialize database: %v", err))
	}
	defer db.Close()
	log.Info("database connection pool initialized")

	log.Info(fmt.Sprintf("monitor running every %d minutes", cfg.Interval))

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup, then on each interval
	runCycle(ctx, db)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runCycle(ctx, db)
		case <-quit:
			log.Info("shutting down...")
			return
		}
	}
}

func runCycle(ctx context.Context, db *database.DB) {
	monitored, err := getMonitoredURLs(ctx, db)
	if err != nil {
		log.Error(fmt.Sprintf("failed to fetch monitored URLs: %v", err))
		return
	}
	if len(monitored) == 0 {
		log.Info("no monitored URLs found")
		return
	}

	shuffle(monitored)

	pings := pingMonitored(monitored)
	log.Info(fmt.Sprintf("completed %d pings", len(pings)))

	if err = uploadPings(ctx, db, pings); err != nil {
		log.Error(fmt.Sprintf("failed to upload pings: %v", err))
		return
	}
	log.Info(fmt.Sprintf("uploaded %d pings", len(pings)))

	if err = deleteExpiredPings(ctx, db); err != nil {
		log.Error(fmt.Sprintf("failed to delete expired pings: %v", err))
	}
}

func getMonitoredURLs(ctx context.Context, db *database.DB) ([]MonitorRow, error) {
	rows, err := db.Pool.Query(ctx, "SELECT * FROM monitor;")
	if err != nil {
		return nil, fmt.Errorf("failed to query monitored URLs: %w", err)
	}
	defer rows.Close()

	monitors := make([]MonitorRow, 0)
	for rows.Next() {
		monitor := new(MonitorRow)
		if err := rows.Scan(&monitor.APIKey, &monitor.URL, &monitor.Secure, &monitor.Ping, &monitor.CreatedAt); err == nil {
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
	if len(pings) == 0 {
		return nil
	}

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
	expiryTime := time.Now().Add(-60 * 24 * time.Hour).UTC()
	_, err := db.Pool.Exec(ctx, "DELETE FROM pings WHERE created_at < $1;", expiryTime)
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

			mu.Lock()
			pings = append(pings, PingsRow{
				APIKey:       m.APIKey,
				URL:          m.URL,
				ResponseTime: int(elapsed.Milliseconds()),
				Status:       status,
				CreatedAt:    time.Now(),
			})
			mu.Unlock()
		}(m)
	}

	wg.Wait()
	return pings
}

func ping(client http.Client, url string, secure bool, ping bool) (int, time.Duration, error) {
	method := getMethod(ping)

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %v", err)
	}

	start := time.Now()
	response, err := client.Do(request)
	elapsed := time.Since(start)

	if err != nil {
		return 0, elapsed, fmt.Errorf("failed to ping URL %s: %v", url, err)
	}
	defer response.Body.Close()

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
	return http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
}

func shuffle(monitored []MonitorRow) {
	rand.Shuffle(len(monitored), func(i, j int) {
		monitored[i], monitored[j] = monitored[j], monitored[i]
	})
}
