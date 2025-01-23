package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"sync"
)

const defaultServerURL string = "https://www.apianalytics-server.com/"

type Client struct {
	apiKey       string
	framework    string
	privacyLevel int
	endpointURL  string

	mu			 sync.Mutex
	requests     []RequestData
	lastPush     time.Time
}


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

func NewClient(apiKey string, framework string, privacyLevel int, serverURL string) *Client {
	if apiKey == "" {
		return nil
	}

	getEndpointURL := func(serverURL string) string {
		if serverURL == "" {
			return defaultServerURL + "api/log-request"
		}
		if serverURL[len(serverURL)-1] == '/' {
			return serverURL + "api/log-request"
		}
		return serverURL + "/api/log-request"
	}

	return &Client{
		apiKey:       apiKey,
		framework:    framework,
		privacyLevel: privacyLevel,
		endpointURL:    getEndpointURL(serverURL),
		lastPush:     time.Now(),
	}
}

func (c *Client) LogRequest(request RequestData) {
	if c == nil || c.apiKey == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.requests = append(c.requests, request)
	if time.Since(c.lastPush) > time.Minute {
		c.pushRequests()
	}
}

func (c *Client) pushRequests() {
	if len(c.requests) == 0 {
		return
	}

	requestsCopy := make([]RequestData, len(c.requests))
	copy(requestsCopy, c.requests) 

	go c.post(requestsCopy)
	c.requests = nil
	c.lastPush = time.Now()
}

func (c *Client) post(requests []RequestData) {
	data := Payload{
		APIKey:       c.apiKey,
		Requests:     requests,
		Framework:    c.framework,
		PrivacyLevel: c.privacyLevel,
	}
	body, err := json.Marshal(data)
	if err == nil {
		http.Post(c.endpointURL, "application/json", bytes.NewBuffer(body))
	}
}
