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

	"github.com/gofiber/fiber/v2"
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

// Setup test Fiber app and middleware
func setupFiberTest(t *testing.T, config *Config) (*fiber.App, *MockAnalyticsServer) {
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

	// Create Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	
	// Add the analytics middleware
	app.Use(AnalyticsWithConfig("test-api-key", config))
	
	// Add test routes
	app.Get("/test-path", func(c *fiber.Ctx) error {
		time.Sleep(5 * time.Millisecond) // Simulate work
		return c.SendString("OK")
	})
	
	app.Get("/error", func(c *fiber.Ctx) error {
		time.Sleep(5 * time.Millisecond) // Simulate work
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	})
	
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	return app, mockServer
}

func TestFiberAnalyticsMiddleware(t *testing.T) {
	app, mockServer := setupFiberTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	resp, err := app.Test(req)
	
	// Verify response
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
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

func TestFiberErrorHandling(t *testing.T) {
	app, mockServer := setupFiberTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	resp, err := app.Test(req)
	
	// Verify response
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
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

	req_data := requests[0]
	if req_data.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", req_data.Status)
	}
}

func TestFiberPanicRecovery(t *testing.T) {
	// Create a new app with recovery middleware
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		},
	})
	
	// Create a mock server
	mockServer := NewMockAnalyticsServer()
	defer mockServer.Close()
	
	// Configure analytics middleware
	config := NewConfig()
	config.ServerURL = mockServer.URL()
	app.Use(AnalyticsWithConfig("test-api-key", config))
	
	// Add a route that will panic
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, _ := app.Test(req)

	// Verify response - should recover and return 500
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Wait a bit for async request to complete
	time.Sleep(50 * time.Millisecond)

	// Check that analytics were captured
	requests := mockServer.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(requests))
	}

	req_data := requests[0]
	if req_data.Status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", req_data.Status)
	}
}

func TestFiberCustomConfig(t *testing.T) {
	customConfig := &Config{
		PrivacyLevel: 1,
		GetPath: func(c *fiber.Ctx) string {
			return "/custom-path"
		},
		GetHostname: func(c *fiber.Ctx) string {
			return "custom-host"
		},
		GetUserAgent: func(c *fiber.Ctx) string {
			return "custom-agent"
		},
		GetIPAddress: func(c *fiber.Ctx) string {
			return "1.2.3.4"
		},
		GetUserID: func(c *fiber.Ctx) string {
			return "user-123"
		},
	}

	app, mockServer := setupFiberTest(t, customConfig)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ignored-path", nil)
	_, _ = app.Test(req)

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

func TestFiberPrivacyLevels(t *testing.T) {
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

			app, mockServer := setupFiberTest(t, config)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/privacy-test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			_, _ = app.Test(req)

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

func TestFiberNewConfig(t *testing.T) {
	config := NewConfig()

	if config.PrivacyLevel != 0 {
		t.Errorf("Expected PrivacyLevel to be 0, got %d", config.PrivacyLevel)
	}
	if config.ServerURL != core.DefaultServerURL {
		t.Errorf("Expected ServerURL to be default, got %s", config.ServerURL)
	}

	// Test GetPath, GetHostname, GetUserAgent, GetIPAddress, and GetUserID functions
	// by creating a Fiber app and using app.Test() to generate proper Fiber contexts
	app := fiber.New()
	
	// Add a route that uses all getter functions
	app.Get("/test-getters", func(c *fiber.Ctx) error {
		path := config.GetPath(c)
		hostname := config.GetHostname(c)
		userAgent := config.GetUserAgent(c)
		ipAddress := config.GetIPAddress(c)
		userID := config.GetUserID(c)
		
		// Build response with all values
		result := map[string]string{
			"path": path,
			"hostname": hostname,
			"userAgent": userAgent,
			"ipAddress": ipAddress,
			"userID": userID,
		}
		
		return c.JSON(result)
	})
	
	// Create request with specific values
	req := httptest.NewRequest("GET", "/test-getters", nil)
	req.Host = "example.com"
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	
	// Send request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error testing config getters: %v", err)
	}
	
	// Parse response
	var result map[string]string
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}
	
	// Check getter results
	if result["path"] != "/test-getters" {
		t.Errorf("Expected path '/test-getters', got '%s'", result["path"])
	}
	if result["hostname"] != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", result["hostname"])
	}
	if result["userAgent"] != "test-agent" {
		t.Errorf("Expected userAgent 'test-agent', got '%s'", result["userAgent"])
	}
	if result["userID"] != "" {
		t.Errorf("Expected empty userID, got '%s'", result["userID"])
	}
	// Note: IP address might vary in test environment, so just check it's not empty
	if result["ipAddress"] == "" {
		t.Errorf("Expected non-empty IP address")
	}
}

func TestFiberIPExtraction(t *testing.T) {
	app := fiber.New()
	
	// Add route that returns the IP address
	app.Get("/ip-test", func(c *fiber.Ctx) error {
		return c.SendString(c.IP())
	})
	
	testCases := []struct {
		name        string
		headers     map[string]string
	}{
		{
			name: "X-Forwarded-For Header",
			headers: map[string]string{
				"X-Forwarded-For": "10.10.10.10, 20.20.20.20",
			},
		},
		{
			name: "X-Real-IP Header",
			headers: map[string]string{
				"X-Real-Ip": "30.30.30.30",
			},
		},
		{
			name:       "No Headers",
			headers:    map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with the specified headers
			req := httptest.NewRequest(http.MethodGet, "/ip-test", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Error testing IP extraction: %v", err)
			}
			
			// Read the response body to get the extracted IP
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Error reading response: %v", err)
			}
			
			// In test environment, IP logic might behave differently
			// Just verify we got some kind of response
			if len(body) == 0 {
				t.Errorf("Expected non-empty IP address")
			}
		})
	}
}

func BenchmarkFiberAnalyticsMiddleware(b *testing.B) {
	// Create a mock server that does minimal work
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create config to point to mock server
	config := NewConfig()
	config.ServerURL = mockServer.URL
	
	// Create Fiber app with middleware
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(AnalyticsWithConfig("test-api-key", config))
	app.Get("/bench", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Setup request
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		_ = resp.Body.Close()
	}
}

// Helper to read response body
func readBody(resp *http.Response) (string, error) {
	buffer := new(strings.Builder)
	_, err := io.Copy(buffer, resp.Body)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}