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

func TryLogRequests() error {
	apiKey := getTestAPIKey()

	postBody, err := json.Marshal(map[string]interface{}{
		"api_key":   apiKey,
		"framework": "FastAPI",
		"requests": []Request{
			Request{
				Path:         "/v1/test",
				Hostname:     "api-analytics.com",
				IPAddress:    "192.168.0.1",
				UserAgent:    "test",
				Method:       "GET",
				Status:       200,
				ResponseTime: 10,
				CreatedAt:    time.Now().Format(time.RFC3339),
			},
			Request{
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

	response, err := http.Post(url+"log-request", "application/json", bytes.NewBuffer(postBody))
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
	api        bool
	logger     bool
	nginx      bool
	postgresql bool
}

func (s ServiceStatus) ServiceDown() bool {
	return s.api && s.logger && s.nginx && s.postgresql
}

type APITestStatus struct {
	newUser            error
	fetchDashboardData error
	fetchData          error
}

func (s APITestStatus) TestFailed() bool {
	return s.newUser != nil || s.fetchDashboardData != nil || s.fetchData != nil
}
