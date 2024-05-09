package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

var requests []RequestData
var lastPosted time.Time = time.Now()

const DefaultServerURL string = "https://www.apianalytics-server.com/"

type Payload struct {
	APIKey       string        `json:"api_key"`
	Requests     []RequestData `json:"requests"`
	Framework    string        `json:"framework"`
	PrivacyLevel int           `json:"privacy_level"`
}

type RequestData struct {
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	ResponseTime int64  `json:"response_time"`
	Status       int    `json:"status"`
	UserID       string `json:"user_id"`
	CreatedAt    string `json:"created_at"`
}

func getServerEndpoint(serverURL string) {
	if serverURL = "" {
		return DefaultServerUrl + "api/log-request"
	}
	if serverURL.HasSuffix("/") {
		return serverURL + "api/log-request"
	}
	return serverURL + "/api/log-request"
}

func postRequest(apiKey string, requests []RequestData, framework string, privacyLevel int, serverURL string) {
	data := Payload{
		APIKey:       apiKey,
		Requests:     requests,
		Framework:    framework,
		PrivacyLevel: privacyLevel,
	}
	body, err := json.Marshal(data)
	if err == nil {
		url := getServerEndpoint(serverURL)
		http.Post(url, "application/json", bytes.NewBuffer(body))
	}
}

func LogRequest(apiKey string, request RequestData, framework string, privacyLevel int, serverURL string) {
	if apiKey == "" {
		return
	}
	now := time.Now()
	requests = append(requests, request)
	if time.Since(lastPosted) > time.Minute {
		go postRequest(apiKey, requests, framework, privacyLevel, serverURL)
		requests = nil
		lastPosted = now
	}
}
