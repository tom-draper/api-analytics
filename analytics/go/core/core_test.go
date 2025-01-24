package core

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines or comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}

	return scanner.Err()
}

func getTestClient(serverURL string) *Client {
	apiKey := os.Getenv("API_KEY")
	return NewClient(apiKey, "Gin", 0, serverURL)
}

func TestMain(m *testing.M) {
	// Load environment variables before running tests
	err := loadEnv(".env")
	if err != nil {
		// Log the error and exit if the .env file fails to load
		panic("Failed to load .env file: " + err.Error())
	}

	// Run the tests
	code := m.Run()

	// Perform any teardown tasks here if necessary

	// Exit with the test exit code
	os.Exit(code)
}

func TestNewClient(t *testing.T) {
	client := getTestClient("")
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	if client.endpointURL != defaultServerURL+"api/log-request" {
		t.Errorf("Expected endpointURL to be '%s', got '%s'", defaultServerURL+"api/log-request", client.endpointURL)
	}
}

func TestLogRequest(t *testing.T) {
	client := getTestClient("")
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	req := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	client.LogRequest(req)

	select {
	case r := <-client.requestChannel:
		if r != req {
			t.Errorf("Expected logged request to be %+v, got %+v", req, r)
		}
	default:
		t.Error("Expected request to be logged but channel was empty")
	}
}

func TestWorkerPushing(t *testing.T) {
	req1 := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user-1",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	req2 := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user-2",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	requests := []RequestData{}
	serverTriggerCount := 0
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle requests to /api/log-request
		if r.URL.Path != "/api/log-request" {
			t.Errorf("Expected to hit /api/log-request, but got %s", r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		mu.Lock()
		requests = append(requests, payload.Requests...)
		serverTriggerCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := getTestClient(ts.URL)
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	client.LogRequest(req1)
	client.LogRequest(req2)

	// Wait 1 minute for worker to begin pushing requests
	time.Sleep(time.Minute)
	// Allow worker to process requests
	time.Sleep(time.Second)

	client.Shutdown()

	if serverTriggerCount == 0 {
		t.Errorf("Server was not triggered")
	} else if serverTriggerCount != 1 {
		t.Errorf("Expected server to be triggered once, got %d", serverTriggerCount)
	}

	if len(requests) != 2 {
		t.Errorf("Expected 2 requests to be pushed, got %d", len(requests))
	}

	if requests[0].UserID == requests[1].UserID {
		t.Errorf("Expected requests to have different user IDs, got %s", requests[0].UserID)
	}

	if requests[0].UserID != req1.UserID || requests[1].UserID != req2.UserID {
		t.Errorf("Pushed requests do not match logged requests")
	}
}

func TestWorkerPushingMultiple(t *testing.T) {
	req1 := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user-1",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	req2 := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user-2",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	req3 := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "test-user-3",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	requests := []RequestData{}
	serverTriggerCount := 0
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle requests to /api/log-request
		if r.URL.Path != "/api/log-request" {
			t.Errorf("Expected to hit /api/log-request, but got %s", r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		mu.Lock()
		requests = append(requests, payload.Requests...)
		serverTriggerCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := getTestClient(ts.URL)
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	client.LogRequest(req1)
	client.LogRequest(req2)

	// Wait 1 minute for worker to begin pushing requests
	time.Sleep(time.Minute)
	// Allow worker to process requests
	time.Sleep(time.Second)

	client.LogRequest(req3)

	// Wait 1 minute for worker to begin pushing requests
	time.Sleep(time.Minute)
	// Allow worker to process requests
	time.Sleep(time.Second)

	client.Shutdown()

	if serverTriggerCount == 0 {
		t.Errorf("Server was not triggered")
	} else if serverTriggerCount != 2 {
		t.Errorf("Expected server to be triggered twice, got %d", serverTriggerCount)
	}

	if len(requests) != 3 {
		t.Errorf("Expected 2 requests to be pushed, got %d", len(requests))
	}

	if requests[0].UserID == requests[1].UserID || requests[1].UserID == requests[2].UserID || requests[0].UserID == requests[2].UserID {
		t.Errorf("Expected requests to have different user IDs, got %s, %s, %s", requests[0].UserID, requests[1].UserID, requests[2].UserID)
	}

	if requests[0].UserID != req1.UserID || requests[1].UserID != req2.UserID || requests[2].UserID != req3.UserID {
		t.Errorf("Pushed requests do not match logged requests")
	}
}

func TestClientShutdown(t *testing.T) {
	client := getTestClient("")
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	req := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "user-1",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
	client.LogRequest(req)

	client.Shutdown()

	time.Sleep(time.Second)

	select {
	case req, open := <-client.requestChannel:
		if open {
			t.Errorf("Expected request channel to be closed after shutdown %v", req)
		}
	default:
		// Channel closed as expected
	}
}

func TestPerformanceLogRequests(t *testing.T) {
	requests := []RequestData{}
	serverTriggerCount := 0
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle requests to /api/log-request
		if r.URL.Path != "/api/log-request" {
			t.Errorf("Expected to hit /api/log-request, but got %s", r.URL.Path)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		mu.Lock()
		requests = append(requests, payload.Requests...)
		serverTriggerCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := getTestClient(ts.URL)
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}
	defer client.Shutdown()

	req := RequestData{
		Hostname:     "localhost",
		IPAddress:    "127.0.0.1",
		Path:         "/test",
		UserAgent:    "test-agent",
		Method:       "GET",
		ResponseTime: 200,
		Status:       200,
		UserID:       "user-1",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	// Measure time to process all requests
	start := time.Now()

	// Generate and log thousands of requests
	numRequests := 10000
	for i := 0; i < numRequests; i++ {
		client.LogRequest(req)
	}

	loggingTime := time.Since(start)

	t.Logf("Processed %d requests in %s, 1 request per %s", numRequests, loggingTime, loggingTime/time.Duration(numRequests))

	if loggingTime > 1*time.Second {
		t.Errorf("Performance test took too long: %s", loggingTime)
	}

	// Wait 1 minute for worker to begin pushing requests
	time.Sleep(time.Minute)
	// Allow worker to process requests
	time.Sleep(time.Second)

	if serverTriggerCount == 0 {
		t.Errorf("Server was not triggered")
	} else if serverTriggerCount != 1 {
		t.Errorf("Expected server to be triggered once, got %d", serverTriggerCount)
	}

	if len(requests) != numRequests {
		t.Errorf("Expected %d requests to be pushed, got %d", numRequests, len(requests))
	}
}

func TestPerformanceShutdown(t *testing.T) {
	client := getTestClient("")
	if client == nil {
		t.Fatalf("Expected client to be initialized, got nil")
	}

	start := time.Now()
	client.Shutdown()
	shutdownTime := time.Since(start)

	t.Logf("Shutdown took %s", shutdownTime)

	if shutdownTime > 1*time.Second {
		t.Errorf("Shutdown took too long: %s", shutdownTime)
	}
}