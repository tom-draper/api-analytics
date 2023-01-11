package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

func getDBLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
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

func getMonitoredURLs(supabase *supa.Client) []MonitorRow {
	// Fetch all API request data associated with this account
	var result []MonitorRow
	err := supabase.DB.From("Monitor").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}
	return result
}

type PingRow struct {
	APIKey       string    `json:"api_key"`
	URL          string    `json:"url"`
	ResponseTime int       `json:"response_time"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func deleteOldPings(supabase *supa.Client) {
	var result interface{}
	err := supabase.DB.From("Pings").Delete().Lt("created_at", time.Now().Add(-60*24*time.Hour).Format("2006-01-02 15:04:05")).Execute(&result)
	if err != nil {
		fmt.Println(err)
	}
}

func uploadPings(pings []PingRow, supabase *supa.Client) {
	var result interface{}
	err := supabase.DB.From("Pings").Insert(pings).Execute(&result)
	if err != nil {
		fmt.Println(err)
	}
}

func pingMonitored(monitored []MonitorRow, client http.Client, supabase *supa.Client) {
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
	uploadPings(pings, supabase)
	deleteOldPings(supabase)
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
	supabaseURL, supabaseKey := getDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	monitored := getMonitoredURLs(supabase)

	client := getClient()
	pingMonitored(monitored, client, supabase)
}
