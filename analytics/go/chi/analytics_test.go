package analytics

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tom-draper/api-analytics/analytics/go/core"
)

// MockAnalyticsServer creates a test server that captures analytics requests
type MockAnalyticsServer struct {
	server   *httptest.Server
	requests []core.RequestData
	mu       sync.Mutex
}

func NewMockAnalyticsServer() *MockAnalyticsServer {
	mock := &MockAnalyticsServer{
		requests: []core.RequestData{},
	}

	// Create the mock server to receive analytics data
	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the incoming request
		if r.Method == "POST" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			var data core.RequestData
			if err := json.Unmarshal(body, &data); err == nil {
				mock.mu.Lock()
				mock.requests = append(mock.requests, data)
				mock.mu.Unlock()
			}
		}
		w.WriteHeader(http.StatusOK)
	}))

	return mock
}

func (m *MockAnalyticsServer) Close() {
	m.server.Close()
}

func (m *MockAnalyticsServer) URL() string {
	return m.server.URL
}

func (m *MockAnalyticsServer) GetRequests() []core.RequestData {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]core.RequestData, len(m.requests))
	copy(result, m.requests)
	return result
}

// Setup test server and middleware
func setupTest(t *testing.T, config *Config) (*httptest.Server, *MockAnalyticsServer) {
	// Create mock analytics server
	mockServer := NewMockAnalyticsServer()
	t.Cleanup(func() {
		mockServer.Close()
	})

	// Apply mock server URL to config
	if config == nil {
		config = NewConfig()
	}
	config.ServerURL = mockServer.URL()

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond) // Simulate work
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create middleware
	handler := AnalyticsWithConfig("test-api-key", config)(testHandler)

	// Create test server
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	return server, mockServer
}

func TestAnalyticsMiddleware(t *testing.T) {
	server, mockServer := setupTest(t, nil)

	// Make a request
	resp, err := http.Get(server.URL + "/test-path")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check that analytics were captured
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req := requests[0]
	if !strings.HasSuffix(req.Path, "/test-path") {
		t.Errorf("Expected path to end with '/test-path', got '%s'", req.Path)
	}
	if req.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req.Method)
	}
	if req.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", req.Status)
	}
	if req.ResponseTime <= 0 {
		t.Errorf("Expected response time > 0, got %d", req.ResponseTime)
	}
}

func TestErrorHandling(t *testing.T) {
	server, mockServer := setupTest(t, nil)

	// Make a request to error path
	resp, err := http.Get(server.URL + "/error")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check that analytics were captured with correct status
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req := requests[0]
	if req.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", req.Status)
	}
}

func TestCustomConfig(t *testing.T) {
	customConfig := &Config{
		PrivacyLevel: 1,
		GetPath: func(r *http.Request) string {
			return "/custom-path"
		},
		GetHostname: func(r *http.Request) string {
			return "custom-host"
		},
		GetUserAgent: func(r *http.Request) string {
			return "custom-agent"
		},
		GetIPAddress: func(r *http.Request) string {
			return "1.2.3.4"
		},
		GetUserID: func(r *http.Request) string {
			return "user-123"
		},
	}

	server, mockServer := setupTest(t, customConfig)

	// Make a request
	resp, err := http.Get(server.URL + "/ignored-path")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check custom values were used
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req := requests[0]
	if req.Path != "/custom-path" {
		t.Errorf("Expected path '/custom-path', got '%s'", req.Path)
	}
	if req.Hostname != "custom-host" {
		t.Errorf("Expected hostname 'custom-host', got '%s'", req.Hostname)
	}
	if req.UserAgent != "custom-agent" {
		t.Errorf("Expected user agent 'custom-agent', got '%s'", req.UserAgent)
	}
	if req.IPAddress != "1.2.3.4" {
		t.Errorf("Expected IP '1.2.3.4', got '%s'", req.IPAddress)
	}
	if req.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", req.UserID)
	}
}

func TestPrivacyLevels(t *testing.T) {
	tests := []struct {
		name         string
		privacyLevel int
		expectIP     bool
	}{
		{"PrivacyLevel0", 0, true},
		{"PrivacyLevel1", 1, true},
		{"PrivacyLevel2", 2, false},
		{"PrivacyLevel3", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				PrivacyLevel: tt.privacyLevel,
			}

			server, mockServer := setupTest(t, config)

			// Make a request
			resp, err := http.Get(server.URL + "/privacy-test")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Wait a bit for async request to complete
			time.Sleep(50 * time.Millisecond)

			// Check privacy settings were respected
			requests := mockServer.GetRequests()
			if len(requests) != 1 {
				t.Fatalf("Expected 1 logged request, got %d", len(requests))
			}

			req := requests[0]
			if tt.expectIP && req.IPAddress == "" {
				t.Errorf("Expected IP address to be recorded at privacy level %d", tt.privacyLevel)
			}
			if !tt.expectIP && req.IPAddress != "" {
				t.Errorf("Expected no IP address at privacy level %d, got '%s'",
					tt.privacyLevel, req.IPAddress)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config.PrivacyLevel != 0 {
		t.Errorf("Expected PrivacyLevel to be 0, got %d", config.PrivacyLevel)
	}
	if config.ServerURL != core.DefaultServerURL {
		t.Errorf("Expected ServerURL to be default, got %s", config.ServerURL)
	}

	// Test function pointers are set correctly by using them
	req := httptest.NewRequest("GET", "http://example.com/path", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "192.168.1.1:1234"

	if path := config.GetPath(req); path != "/path" {
		t.Errorf("Expected path '/path', got '%s'", path)
	}
	if hostname := config.GetHostname(req); hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", hostname)
	}
	if userAgent := config.GetUserAgent(req); userAgent != "test-agent" {
		t.Errorf("Expected userAgent 'test-agent', got '%s'", userAgent)
	}
	if ipAddress := config.GetIPAddress(req); ipAddress != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", ipAddress)
	}
	if userID := config.GetUserID(req); userID != "" {
		t.Errorf("Expected empty userID, got '%s'", userID)
	}
}

func TestPanicRecovery(t *testing.T) {
	// Create config for mock server
	config := NewConfig()
	mockServer := NewMockAnalyticsServer()
	defer mockServer.Close()
	config.ServerURL = mockServer.URL()

	// Create panic handler
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Apply middleware
	handler := AnalyticsWithConfig("test-api-key", config)(panicHandler)

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response - should recover and return 500
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestResponseWriter(t *testing.T) {
	// Test that calling WriteHeader multiple times only sets status once
	rw := createResponseWriter(httptest.NewRecorder())

	rw.WriteHeader(http.StatusOK)
	if rw.status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rw.status)
	}

	// Try to change status
	rw.WriteHeader(http.StatusBadRequest)
	if rw.status != http.StatusOK {
		t.Errorf("Expected status to remain 200, got %d", rw.status)
	}
}

func BenchmarkAnalyticsMiddleware(b *testing.B) {
	// Create a mock server that does minimal work
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create config to point to mock server
	config := NewConfig()
	config.ServerURL = mockServer.URL

	// Create test handler with middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := AnalyticsWithConfig("test-api-key", config)(testHandler)

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup request
	req, _ := http.NewRequest("GET", server.URL, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
	}
}