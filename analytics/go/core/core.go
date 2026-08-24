package core

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
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
	finished       chan struct{}
	shutdownOnce   sync.Once
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
		finished:       make(chan struct{}),
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
	// Signal Shutdown() callers once the final flush has completed.
	defer close(c.finished)

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
				// Keep a failed batch in memory for the next scheduled attempt.
				// Clearing it unconditionally makes temporary logger outages silently
				// lose every request collected during the previous interval.
				if c.pushRequests(requests) {
					requests = nil
				}
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

// pushRequests returns true only when the logger accepted the batch. Callers
// retain a false-returning batch and retry it later.
func (c *Client) pushRequests(requests []RequestData) bool {
	data := Payload{
		APIKey:       c.apiKey,
		Requests:     requests,
		Framework:    c.framework,
		PrivacyLevel: c.privacyLevel,
	}
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to send requests: %v", err)
		return false
	}
	resp, err := httpClient.Post(c.endpointURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to send requests: %v", err)
		return false
	}
	defer resp.Body.Close()
	// The logging endpoint returns 201 Created on success; treat any 2xx as
	// success so a normal upload is not logged as an error.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Server responded with status: %d", resp.StatusCode)
		return false
	}
	return true
}

// Shutdown flushes any buffered requests and blocks until that final upload has
// completed, so callers can drain the client on graceful exit without losing the
// last batch. It is safe to call more than once and from multiple goroutines.
func (c *Client) Shutdown() {
	if c == nil {
		return
	}
	c.shutdownOnce.Do(func() {
		close(c.done)
	})
	<-c.finished
}
