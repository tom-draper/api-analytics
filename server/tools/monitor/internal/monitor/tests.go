package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TryNewUser tests creating and deleting a new user
func (c *Client) TryNewUser() error {
	body, err := c.makeGetRequest(c.apiBaseURL+"generate-api-key", "")
	if err != nil {
		return err
	}
	sb := string(body)
	apiKey := sb[1 : len(sb)-1]
	if len(apiKey) != UUIDLength {
		return ErrInvalidUUID
	}

	if c.db == nil {
		return nil
	}

	ctx := context.Background()
	err = c.db.DeleteUser(ctx, apiKey)
	if err != nil {
		return err
	}
	return nil
}

// TryFetchData tests fetching API data
func (c *Client) TryFetchData() error {
	body, err := c.makeGetRequest(c.apiBaseURL+"data", c.monitorAPIKey)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

// TryFetchDashboardData tests fetching dashboard data
func (c *Client) TryFetchDashboardData() error {
	body, err := c.makeGetRequest(c.apiBaseURL+"requests/"+c.monitorUserID, "")
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

// TryFetchUserID tests fetching user ID
func (c *Client) TryFetchUserID() error {
	body, err := c.makeGetRequest(c.apiBaseURL+"user-id/"+c.monitorAPIKey, "")
	if err != nil {
		return err
	}
	sb := string(body)
	userID := sb[1 : len(sb)-1]
	if len(userID) != UUIDLength {
		return ErrInvalidUUID
	}
	return nil
}

// TryFetchMonitorPings tests fetching monitor pings
func (c *Client) TryFetchMonitorPings() error {
	body, err := c.makeGetRequest(c.apiBaseURL+"monitor/pings/"+c.monitorUserID, "")
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

// TryLogRequests tests logging requests
func (c *Client) TryLogRequests(legacy bool) error {
	postBody, err := json.Marshal(map[string]interface{}{
		"api_key":   c.monitorAPIKey,
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

	response, err := http.Post(c.apiBaseURL+endpoint, "application/json", bytes.NewBuffer(postBody))
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

// RunServiceChecks checks the status of all services
func (c *Client) RunServiceChecks() ServiceStatus {
	return ServiceStatus{
		API:         !ServiceDown("api"),
		Logger:      !ServiceDown("logger"),
		Nginx:       !ServiceDown("nginx"),
		PostgresSQL: !ServiceDown("postgresql"),
	}
}

// RunAPITests runs all API endpoint tests
func (c *Client) RunAPITests() APITestStatus {
	return APITestStatus{
		NewUser:            c.TryNewUser(),
		FetchDashboardData: c.TryFetchDashboardData(),
		FetchData:          c.TryFetchData(),
		FetchUserID:        c.TryFetchUserID(),
		FetchMonitorPings:  c.TryFetchMonitorPings(),
		LogRequests:        c.TryLogRequests(false),
	}
}
