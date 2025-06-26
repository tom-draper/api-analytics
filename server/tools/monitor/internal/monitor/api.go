package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

const UUIDLength = 36

func makeGetRequest(url string, apiKey string) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		request.Header = http.Header{
			"X-AUTH-TOKEN": {apiKey},
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func TryNewUser(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	body, err := makeGetRequest(apiBaseURL+"generate-api-key", "")
	if err != nil {
		return err
	}
	sb := string(body)
	apiKey := sb[1 : len(sb)-1]
	if len(apiKey) != UUIDLength {
		return fmt.Errorf("uuid value returned is invalid")
	}

	err = database.DeleteUser(apiKey)
	if err != nil {
		return err
	}
	return nil
}

func TryFetchData(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	body, err := makeGetRequest(apiBaseURL+"data", monitorAPIKey)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

func TryFetchDashboardData(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	body, err := makeGetRequest(apiBaseURL+"requests/"+monitorUserID, "")
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

func TryFetchUserID(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	body, err := makeGetRequest(apiBaseURL+"user-id/"+monitorAPIKey, "")
	if err != nil {
		return err
	}
	sb := string(body)
	userID := sb[1 : len(sb)-1]
	if len(userID) != UUIDLength {
		return fmt.Errorf("uuid value returned is invalid")
	}
	return nil
}

func TryFetchMonitorPings(apiBaseURL string, monitorAPIKey string, monitorUserID string) error {
	body, err := makeGetRequest(apiBaseURL+"monitor/pings/"+monitorUserID, "")
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

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

func TryLogRequests(apiBaseURL string, monitorAPIKey string, monitorUserID string, legacy bool) error {
	postBody, err := json.Marshal(map[string]interface{}{
		"api_key":   monitorAPIKey,
		"framework": "FastAPI",
		"requests": []Request{
			{
				Path:         "/v1/test",
				Hostname:     "api-analytics.com",
				IPAddress:    "192.168.0.1",
				UserAgent:    "test",
				Method:       "GET",
				Status:       200,
				ResponseTime: 10,
				CreatedAt:    time.Now().Format(time.RFC3339),
			},
			{
				Path:         "/v1/test",
				Hostname:     "api-analytics.com",
				IPAddress:    "192.168.0.1",
				UserAgent:    "test",
				Method:       "POST",
				Status:       201,
				ResponseTime: 20,
				CreatedAt:    time.Now().Format(time.RFC3339),
			},
		},
	})
	if err != nil {
		return err
	}

	var endpoint string
	if legacy {
		endpoint = "log-request"
	} else {
		endpoint = "requests"
	}

	response, err := http.Post(apiBaseURL+endpoint, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("status code: %d\n%s", response.StatusCode, body)
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

type ServiceStatus struct {
	API         bool
	Logger      bool
	Nginx       bool
	PostgresSQL bool
}

func (s ServiceStatus) ServiceDown() bool {
	return s.API && s.Logger && s.Nginx && s.PostgresSQL
}

type APITestStatus struct {
	NewUser            error
	FetchDashboardData error
	FetchData          error
	FetchUserID        error
	FetchMonitorPings  error
	LogRequests        error
}

func (s APITestStatus) TestFailed() bool {
	return s.NewUser != nil || s.FetchDashboardData != nil || s.FetchData != nil || s.FetchUserID != nil || s.FetchMonitorPings != nil || s.LogRequests != nil
}
