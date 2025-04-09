package analytics

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
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

// Setup test Echo instance and middleware
func setupEchoTest(t *testing.T, config *Config) (*echo.Echo, *MockAnalyticsServer) {
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

	// Create Echo instance
	e := echo.New()

	// Add the analytics middleware
	e.Use(AnalyticsWithConfig("test-api-key", config))

	// Add test routes
	e.GET("/test-path", func(c echo.Context) error {
		time.Sleep(5 * time.Millisecond) // Simulate work
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/error", func(c echo.Context) error {
		time.Sleep(5 * time.Millisecond) // Simulate work
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	})

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	return e, mockServer
}

func TestEchoAnalyticsMiddleware(t *testing.T) {
	e, mockServer := setupEchoTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check that analytics were captured
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req_data := requests[0]
	if req_data.Path != "/test-path" {
		t.Errorf("Expected path '/test-path', got '%s'", req_data.Path)
	}
	if req_data.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req_data.Method)
	}
	if req_data.Status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", req_data.Status)
	}
	if req_data.ResponseTime <= 0 {
		t.Errorf("Expected response time > 0, got %d", req_data.ResponseTime)
	}
}

func TestEchoErrorHandling(t *testing.T) {
	e, mockServer := setupEchoTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check that analytics were captured with correct status
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req_data := requests[0]
	if req_data.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", req_data.Status)
	}
}

func TestEchoCustomConfig(t *testing.T) {
	customConfig := &Config{
		PrivacyLevel: 1,
		GetPath: func(c echo.Context) string {
			return "/custom-path"
		},
		GetHostname: func(c echo.Context) string {
			return "custom-host"
		},
		GetUserAgent: func(c echo.Context) string {
			return "custom-agent"
		},
		GetIPAddress: func(c echo.Context) string {
			return "1.2.3.4"
		},
		GetUserID: func(c echo.Context) string {
			return "user-123"
		},
	}

	e, mockServer := setupEchoTest(t, customConfig)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ignored-path", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check custom values were used
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req_data := requests[0]
	if req_data.Path != "/custom-path" {
		t.Errorf("Expected path '/custom-path', got '%s'", req_data.Path)
	}
	if req_data.Hostname != "custom-host" {
		t.Errorf("Expected hostname 'custom-host', got '%s'", req_data.Hostname)
	}
	if req_data.UserAgent != "custom-agent" {
		t.Errorf("Expected user agent 'custom-agent', got '%s'", req_data.UserAgent)
	}
	if req_data.IPAddress != "1.2.3.4" {
		t.Errorf("Expected IP '1.2.3.4', got '%s'", req_data.IPAddress)
	}
	if req_data.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", req_data.UserID)
	}
}

func TestEchoPrivacyLevels(t *testing.T) {
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

			e, mockServer := setupEchoTest(t, config)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/privacy-test", nil)
			rec := httptest.NewRecorder()

			// Add a IP for testing
			req.RemoteAddr = "192.168.1.1:12345"

			// Serve the request
			e.ServeHTTP(rec, req)

			// Wait a bit for async request to complete
			time.Sleep(50 * time.Millisecond)

			// Check privacy settings were respected
			requests := mockServer.GetRequests()
			if len(requests) != 1 {
				t.Fatalf("Expected 1 logged request, got %d", len(requests))
			}

			req_data := requests[0]
			if tt.expectIP && req_data.IPAddress == "" {
				t.Errorf("Expected IP address to be recorded at privacy level %d", tt.privacyLevel)
			}
			if !tt.expectIP && req_data.IPAddress != "" {
				t.Errorf("Expected no IP address at privacy level %d, got '%s'",
					tt.privacyLevel, req_data.IPAddress)
			}
		})
	}
}

func TestEchoNewConfig(t *testing.T) {
	config := NewConfig()

	if config.PrivacyLevel != 0 {
		t.Errorf("Expected PrivacyLevel to be 0, got %d", config.PrivacyLevel)
	}
	if config.ServerURL != core.DefaultServerURL {
		t.Errorf("Expected ServerURL to be default, got %s", config.ServerURL)
	}

	// Create mock Echo context for testing functions
	e := echo.New()
	req := httptest.NewRequest("GET", "http://example.com/path", nil)
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test function pointers are set correctly by using them
	if path := config.GetPath(c); path != "/path" {
		t.Errorf("Expected path '/path', got '%s'", path)
	}
	if hostname := config.GetHostname(c); hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", hostname)
	}
	if userAgent := config.GetUserAgent(c); userAgent != "test-agent" {
		t.Errorf("Expected userAgent 'test-agent', got '%s'", userAgent)
	}
	// Note: Echo's RealIP will default to an empty string in tests
	if userID := config.GetUserID(c); userID != "" {
		t.Errorf("Expected empty userID, got '%s'", userID)
	}
}

func TestEchoRealIPExtraction(t *testing.T) {
	config := NewConfig()
	e, mockServer := setupEchoTest(t, config)

	// Create a test request with various IP headers
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	req.Header.Set("X-Real-IP", "10.10.10.10")
	req.Header.Set("X-Forwarded-For", "20.20.20.20, 30.30.30.30")
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check IP extraction based on Echo's RealIP implementation
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	// Echo's RealIP prioritizes X-Real-IP if present
	req_data := requests[0]
	expectedIP := "10.10.10.10" // Echo should prioritize X-Real-IP
	if req_data.IPAddress != expectedIP {
		t.Errorf("Expected IP '%s', got '%s'", expectedIP, req_data.IPAddress)
	}
}

func BenchmarkEchoAnalyticsMiddleware(b *testing.B) {
	// Create a mock server that does minimal work
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create config to point to mock server
	config := NewConfig()
	config.ServerURL = mockServer.URL

	// Create Echo instance with middleware
	e := echo.New()
	e.Use(AnalyticsWithConfig("test-api-key", config))
	e.GET("/bench", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Setup request
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}
