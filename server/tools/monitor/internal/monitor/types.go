package monitor

import "errors"

const UUIDLength = 36

// Errors
var (
	ErrMissingAPIBaseURL     = errors.New("API_BASE_URL not set")
	ErrMissingMonitorAPIKey  = errors.New("MONITOR_API_KEY not set")
	ErrMissingMonitorUserID  = errors.New("MONITOR_USER_ID not set")
	ErrInvalidUUID           = errors.New("uuid value returned is invalid")
)

// Request represents a logged API request
type Request struct {
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       int16  `json:"status"`
	ResponseTime int16  `json:"response_time"`
	CreatedAt    string `json:"created_at"`
}

// ServiceStatus tracks the status of all services
type ServiceStatus struct {
	API         bool
	Logger      bool
	Nginx       bool
	PostgresSQL bool
}

// ServiceDown returns true if any service is down
func (s ServiceStatus) ServiceDown() bool {
	return !s.API || !s.Logger || !s.Nginx || !s.PostgresSQL
}

// APITestStatus tracks the results of API endpoint tests
type APITestStatus struct {
	NewUser            error
	FetchDashboardData error
	FetchData          error
	FetchUserID        error
	FetchMonitorPings  error
	LogRequests        error
}

// TestFailed returns true if any API test failed
func (s APITestStatus) TestFailed() bool {
	return s.NewUser != nil || s.FetchDashboardData != nil || s.FetchData != nil ||
		s.FetchUserID != nil || s.FetchMonitorPings != nil || s.LogRequests != nil
}
