package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/database"
)

const url string = "https://apianalytics-server.com/api/"

func TryNewUser() error {
	response, err := http.Get(url + "generate-api-key")
	if err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	sb := string(body)
	apiKey := sb[1 : len(sb)-1]
	if len(apiKey) != 36 {
		return fmt.Errorf("uuid value returned is invalid")
	}

	err = database.DeleteUser(apiKey)
	if err != nil {
		return err
	}
	return nil
}

func TryFetchData() error {
	client := http.Client{}
	apiKey := getTestAPIKey()
	request, err := http.NewRequest("GET", url+"data", nil)
	if err != nil {
		return err
	}

	request.Header = http.Header{
		"X-AUTH-TOKEN": {apiKey},
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

func TryFetchDashboardData() error {
	userID := getTestUserID()
	response, err := http.Get(url + "requests/" + userID)
	if err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

func TryFetchUserID() error {
	apiKey := getTestAPIKey()
	response, err := http.Get(url + "user-id/" + apiKey)
	if err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	sb := string(body)
	userID := sb[1 : len(sb)-1]
	if len(userID) != 36 {
		return fmt.Errorf("uuid value returned is invalid")
	}
	return nil
}

func TryFetchMonitorPings() error {
	userID := getTestUserID()
	response, err := http.Get(url + "monitor/pings/" + userID)
	if err != nil {
		return err
	} else if response.StatusCode != 200 {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
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

func TryLogRequests(legacy bool) error {
	apiKey := getTestAPIKey()

	postBody, err := json.Marshal(map[string]interface{}{
		"api_key":   apiKey,
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

	response, err := http.Post(url+endpoint, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("status code: %d\n%s", response.StatusCode, body)
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	return err
}

func getTestAPIKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	apiKey := os.Getenv("MONITOR_API_KEY")
	return apiKey
}

func getTestUserID() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	userID := os.Getenv("MONITOR_USER_ID")
	return userID
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
}

func (s APITestStatus) TestFailed() bool {
	return s.NewUser != nil || s.FetchDashboardData != nil || s.FetchData != nil
}
