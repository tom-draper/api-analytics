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

	"github.com/gin-gonic/gin"
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

// Setup test Gin engine and middleware
func setupGinTest(t *testing.T, config *Config) (*gin.Engine, *MockAnalyticsServer) {
	// Disable Gin debug output during tests
	gin.SetMode(gin.TestMode)

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

	// Create Gin engine
	r := gin.New()
	
	// Add the analytics middleware
	r.Use(AnalyticsWithConfig("test-api-key", config))
	
	// Add test routes
	r.GET("/test-path", func(c *gin.Context) {
		time.Sleep(5 * time.Millisecond) // Simulate work
		c.String(http.StatusOK, "OK")
	})
	
	r.GET("/error", func(c *gin.Context) {
		time.Sleep(5 * time.Millisecond) // Simulate work
		c.String(http.StatusInternalServerError, "Internal Server Error")
	})
	
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	return r, mockServer
}

func TestGinAnalyticsMiddleware(t *testing.T) {
	r, mockServer := setupGinTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(rec, req)

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

func TestGinErrorHandling(t *testing.T) {
	r, mockServer := setupGinTest(t, nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(rec, req)

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

func TestGinPanicRecovery(t *testing.T) {
	// Create a new engine with the Recovery middleware
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	
	// Create a mock server
	mockServer := NewMockAnalyticsServer()
	defer mockServer.Close()
	
	// Configure analytics middleware
	config := NewConfig()
	config.ServerURL = mockServer.URL()
	r.Use(AnalyticsWithConfig("test-api-key", config))
	
	// Add a route that will panic
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(rec, req)

	// Verify response - should recover and return 500
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
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

func TestGinCustomConfig(t *testing.T) {
	customConfig := &Config{
		PrivacyLevel: 1,
		GetPath: func(c *gin.Context) string {
			return "/custom-path"
		},
		GetHostname: func(c *gin.Context) string {
			return "custom-host"
		},
		GetUserAgent: func(c *gin.Context) string {
			return "custom-agent"
		},
		GetIPAddress: func(c *gin.Context) string {
			return "1.2.3.4"
		},
		GetUserID: func(c *gin.Context) string {
			return "user-123"
		},
	}

	r, mockServer := setupGinTest(t, customConfig)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ignored-path", nil)
	rec := httptest.NewRecorder()
	
	// Serve the request
	r.ServeHTTP(rec, req)

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

func TestGinPrivacyLevels(t *testing.T) {
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

			r, mockServer := setupGinTest(t, config)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/privacy-test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1")
			rec := httptest.NewRecorder()
			
			// Serve the request
			r.ServeHTTP(rec, req)

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

func TestGinNewConfig(t *testing.T) {
	config := NewConfig()

	if config.PrivacyLevel != 0 {
		t.Errorf("Expected PrivacyLevel to be 0, got %d", config.PrivacyLevel)
	}
	if config.ServerURL != core.DefaultServerURL {
		t.Errorf("Expected ServerURL to be default, got %s", config.ServerURL)
	}

	// Create mock Gin context for testing functions
	gin.SetMode(gin.TestMode)
	_, _ = gin.CreateTestContext(httptest.NewRecorder())
	req, _ := http.NewRequest("GET", "http://example.com/path", nil)
	req.Header.Set("User-Agent", "test-agent")
	
	// Set up a Gin context for testing
	engine := gin.New()
	c := gin.CreateTestContextOnly(httptest.NewRecorder(), engine)
	c.Request = req
	
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
	// Note: ClientIP will return empty string in this test context
	if userID := config.GetUserID(c); userID != "" {
		t.Errorf("Expected empty userID, got '%s'", userID)
	}
}

func TestGinClientIPExtraction(t *testing.T) {
	config := NewConfig()
	r, mockServer := setupGinTest(t, config)

	testCases := []struct {
		name      string
		headers   map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For Header",
			headers: map[string]string{
				"X-Forwarded-For": "10.10.10.10, 20.20.20.20",
			},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "10.10.10.10",
		},
		{
			name: "X-Real-IP Header",
			headers: map[string]string{
				"X-Real-Ip": "30.30.30.30",
			},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "30.30.30.30",
		},
		{
			name:       "Remote Addr Only",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with the specified headers
			req := httptest.NewRequest(http.MethodGet, "/ip-test", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tc.remoteAddr
			rec := httptest.NewRecorder()
			
			// Serve the request
			r.ServeHTTP(rec, req)

			// Wait a bit for async request to complete
			time.Sleep(50 * time.Millisecond)

			// Check IP extraction
			requests := mockServer.GetRequests()
			// Find the latest request for this test case
			var req_data core.RequestData
			for i := len(requests) - 1; i >= 0; i-- {
				if strings.Contains(requests[i].Path, "/ip-test") {
					req_data = requests[i]
					break
				}
			}
			
			if req_data.IPAddress != tc.expectedIP {
				t.Errorf("Expected IP '%s', got '%s'", tc.expectedIP, req_data.IPAddress)
			}
		})
	}
}

func BenchmarkGinAnalyticsMiddleware(b *testing.B) {
	// Disable Gin debug output during benchmarks
	gin.SetMode(gin.ReleaseMode)
	
	// Create a mock server that does minimal work
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Create config to point to mock server
	config := NewConfig()
	config.ServerURL = mockServer.URL
	
	// Create Gin engine with middleware
	r := gin.New()
	r.Use(AnalyticsWithConfig("test-api-key", config))
	r.GET("/bench", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup request
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)
	rec := httptest.NewRecorder()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rec, req)
	}
}