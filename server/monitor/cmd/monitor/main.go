package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/monitor/internal/alert"
	"github.com/tom-draper/api-analytics/server/monitor/internal/config"
	"github.com/tom-draper/api-analytics/server/monitor/internal/log"
)

// pingResult is the outcome of pinging one monitored URL. A non-nil err means
// the URL was unreachable, in which case no ping row is stored.
type pingResult struct {
	monitor MonitorRow
	status  int
	elapsed time.Duration
	err     error
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

	// Email alerts are optional and off by default. A nil sender (no
	// EMAIL_PROVIDER) or disabled config yields a no-op alerter.
	sender, err := email.NewFromEnv()
	if err != nil {
		log.Fatal(fmt.Sprintf("email configuration error: %v", err))
	}
	alerter := alert.New(cfg.AlertsEnabled, sender, cfg.AlertEmail)
	if alerter.Enabled() {
		log.Info(fmt.Sprintf("email alerts enabled, notifying %s", cfg.AlertEmail))
	} else if cfg.AlertsEnabled {
		log.Info("email alerts requested but no email provider is configured; alerts disabled")
	}

	log.Info(fmt.Sprintf("monitor running every %d minutes", cfg.Interval))

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup, then on each interval
	runCycle(ctx, db, alerter)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runCycle(ctx, db, alerter)
		case <-quit:
			log.Info("shutting down...")
			return
		}
	}
}

func runCycle(ctx context.Context, db *database.DB, alerter *alert.Alerter) {
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

	results := pingMonitored(monitored)
	pings := successfulPings(results)
	log.Info(fmt.Sprintf("completed %d pings", len(pings)))

	// Alert on any up/down transitions before touching the database, so a DB
	// error does not suppress downtime notifications.
	alerter.Evaluate(alertResults(results))

	if err = uploadPings(ctx, db, pings); err != nil {
		log.Error(fmt.Sprintf("failed to upload pings: %v", err))
		return
	}
	log.Info(fmt.Sprintf("uploaded %d pings", len(pings)))

	if err = deleteExpiredPings(ctx, db); err != nil {
		log.Error(fmt.Sprintf("failed to delete expired pings: %v", err))
	}
}

// successfulPings turns the reachable results into rows for upload.
func successfulPings(results []pingResult) []PingsRow {
	pings := make([]PingsRow, 0, len(results))
	for _, r := range results {
		if r.err != nil {
			continue
		}
		pings = append(pings, PingsRow{
			APIKey:       r.monitor.APIKey,
			URL:          r.monitor.URL,
			ResponseTime: int(r.elapsed.Milliseconds()),
			Status:       r.status,
			CreatedAt:    time.Now(),
		})
	}
	return pings
}

// alertResults adapts ping results for the alerter.
func alertResults(results []pingResult) []alert.Result {
	out := make([]alert.Result, 0, len(results))
	for _, r := range results {
		out = append(out, alert.Result{
			APIKey: r.monitor.APIKey,
			URL:    r.monitor.URL,
			Status: r.status,
			Err:    r.err,
		})
	}
	return out
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

func pingMonitored(monitored []MonitorRow) []pingResult {
	client := getClient()
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]pingResult, 0, len(monitored))

	for _, m := range monitored {
		wg.Add(1)
		go func(m MonitorRow) {
			defer wg.Done()

			status, elapsed, err := ping(client, m.URL, m.Secure, m.Ping)
			if err != nil {
				// An unreachable URL is recorded as a result (for alerting) but
				// not stored as a ping row.
				log.Error(err.Error())
			}

			mu.Lock()
			results = append(results, pingResult{
				monitor: m,
				status:  status,
				elapsed: elapsed,
				err:     err,
			})
			mu.Unlock()
		}(m)
	}

	wg.Wait()
	return results
}

func ping(client http.Client, url string, secure bool, ping bool) (int, time.Duration, error) {
	method := getMethod(ping)
	url = applyScheme(url, secure)

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

// applyScheme ensures url carries an http/https scheme so it can be pinged.
// A url that already specifies a scheme is used as-is (the dashboard bakes the
// scheme in when a monitor is added); a scheme-less url — which a direct API
// caller can store — falls back to the monitor's secure flag, selecting https
// when secure and http otherwise.
func applyScheme(url string, secure bool) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	if secure {
		return "https://" + url
	}
	return "http://" + url
}

func getClient() http.Client {
	dialer := net.Dialer{
		Timeout: 2 * time.Second,
		// Control runs after DNS resolution, on the address actually being
		// dialed, and rejects non-public targets. Because it sees the resolved
		// IP it also defeats DNS rebinding between when a monitor is added and
		// when it is pinged.
		Control: guardPrivateAddress,
	}
	return http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
	}
}

// errPrivateAddress is returned when a monitored URL resolves to a non-public
// address, so the monitor pinger cannot be abused to reach internal services.
var errPrivateAddress = errors.New("refusing to connect to non-public address")

// guardPrivateAddress is a net.Dialer Control function that blocks connections
// to any address that is not a routable public IP: loopback, link-local
// (including the 169.254.169.254 cloud metadata endpoint), private (RFC 1918
// and IPv6 ULA), carrier-grade NAT, multicast and unspecified addresses.
func guardPrivateAddress(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid dial address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("unresolved dial address %q", host)
	}
	if !database.IsPublicIP(ip) {
		return fmt.Errorf("%w: %s", errPrivateAddress, ip)
	}
	return nil
}

func shuffle(monitored []MonitorRow) {
	rand.Shuffle(len(monitored), func(i, j int) {
		monitored[i], monitored[j] = monitored[j], monitored[i]
	})
}
