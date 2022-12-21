package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

func GetDBLogin() (string, string) {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	return supabaseURL, supabaseKey
}

func Ping(client http.Client, domain string, secure bool, method string) (int, time.Duration, error) {
	var url string
	if !secure {
		url = "http://" + domain
	} else {
		url = "https://" + domain
	}

	if method != "GET" && method != "HEAD" {
		return 0, time.Duration(0), errors.New("invalid method")
	}

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
	Method    bool      `json:"method"`
	Secure    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func getMonitoredURLs(supabase *supa.Client) {
	// Fetch all API request data associated with this account
	var result []MonitorRow
	err := supabase.DB.From("Monitor").Select("*").Execute(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func main() {
	supabaseURL, supabaseKey := GetDBLogin()
	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	getMonitoredURLs(supabase)

	dialer := net.Dialer{Timeout: 2 * time.Second}
	var client = http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}

	res, elapsed, err := Ping(client, "example.com", true, "GET")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res, elapsed)
	}
}
