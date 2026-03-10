package core

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

const DefaultServerURL string = "https://www.apianalytics-server.com/"

type Client struct {
	apiKey       string
	framework    string
	privacyLevel int
	endpointURL  string

	requestChannel chan RequestData
	done           chan struct{}
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
		log.Println("Failed to create new API Analytics client: API key is required")
		return nil
	}

	getEndpointURL := func(serverURL string) string {
		if serverURL == "" {
			return DefaultServerURL + "api/log-request"
		}
		if serverURL[len(serverURL)-1] == '/' {
			return serverURL + "api/log-request"
		}
		return serverURL + "/api/log-request"
	}

	client := &Client{
		apiKey:         apiKey,
		framework:      framework,
		privacyLevel:   privacyLevel,
		endpointURL:    getEndpointURL(serverURL),
		requestChannel: make(chan RequestData, 1000),
		done:           make(chan struct{}),
	}

	go client.worker()
	return client
}

func (c *Client) LogRequest(request RequestData) {
	if c == nil || c.apiKey == "" {
		return
	}

	select {
	case c.requestChannel <- request:
	default:
		log.Println("API Analytics: request buffer full, dropping request")
	}
}

func (c *Client) worker() {
	var requests []RequestData
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case request := <-c.requestChannel:
			requests = append(requests, request)

		case <-ticker.C:
			// Push any logged requests periodically
			if len(requests) > 0 {
				c.pushRequests(requests)
				requests = nil
			}

		case <-c.done:
			// Drain any requests still sitting in the channel
			for len(c.requestChannel) > 0 {
				requests = append(requests, <-c.requestChannel)
			}
			if len(requests) > 0 {
				c.pushRequests(requests)
			}
			return
		}
	}
}

func (c *Client) pushRequests(requests []RequestData) {
	data := Payload{
		APIKey:       c.apiKey,
		Requests:     requests,
		Framework:    c.framework,
		PrivacyLevel: c.privacyLevel,
	}
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to send requests: %v", err)
		return
	}
	resp, err := httpClient.Post(c.endpointURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to send requests: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server responded with status: %d", resp.StatusCode)
	}
}

func (c *Client) Shutdown() {
	if c == nil {
		return
	}
	close(c.done)
}
