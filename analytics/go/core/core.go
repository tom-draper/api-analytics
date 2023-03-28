package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var requests []Data
var last_posted time.Time = time.Now()

type Data struct {
	APIKey       string    `json:"api_key"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	Path         string    `json:"path"`
	UserAgent    string    `json:"user_agent"`
	Method       string    `json:"method"`
	ResponseTime int64     `json:"response_time"`
	Status       int       `json:"status"`
	Framework    string    `json:"framework"`
	CreatedAt    time.Time `json:"created_at"`
}

func Post(data []Data) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	http.Post("http://213.168.248.206/api/log-request", "application/json", bytes.NewBuffer(reqBody))
}

func LogRequest(data Data) {
	now := time.Now()
	requests = append(requests, data)
	if time.Since(last_posted) > time.Minute {
		go Post(requests)
		requests = nil
		last_posted = now
	}
}
