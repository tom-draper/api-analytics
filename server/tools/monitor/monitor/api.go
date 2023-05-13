package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/database"
)

var url string = "https://apianalytics-server.com/api/"

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

func buildEmailBody(serviceStatus ServiceStatus, apiTestStatus APITestStatus) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Failure detected at %v\n", time.Now()))

	if !serviceStatus.api {
		body.WriteString("Service api down\n")
	}
	if !serviceStatus.logger {
		body.WriteString("Service logger down\n")
	}
	if !serviceStatus.nginx {
		body.WriteString("Service nginx down\n")
	}
	if !serviceStatus.postgresql {
		body.WriteString("Service postgresql down\n")
	}

	if apiTestStatus.newUser != nil {
		body.WriteString(fmt.Sprintf("Error when creating new user: %s\n", apiTestStatus.newUser.Error()))
	}
	if apiTestStatus.fetchDashboardData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching dashboard data: %s\n", apiTestStatus.fetchDashboardData.Error()))
	}
	if apiTestStatus.fetchData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching API data: %s\n", apiTestStatus.fetchData.Error()))
	}

	return body.String()
}
