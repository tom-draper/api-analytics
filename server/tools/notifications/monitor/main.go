package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"

	"github.com/joho/godotenv"
)

var url string = "https://apianalytics-server.com/api/"

func TestNewUser() error {
	apiKey, err := createNewUser()
	if err != nil {
		return err
	}
	err = database.DeleteUser(apiKey)
	if err != nil {
		return err
	}
	return nil
}

func createNewUser() (string, error) {
	response, err := http.Get(url + "generate-api-key")
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("status code: %d", response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	apiKey := sb[1 : len(sb)-1]
	return apiKey, nil
}

func TestFetchData() error {
	client := http.Client{}
	apiKey := getAPIKey()
	request, err := http.NewRequest("GET", url+"data", nil)
	if err != nil {
		return err
	}
	request.Header = http.Header{
		"Content-Type": {"application/json"},
		"X-Auth-Token": {apiKey},
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("status code: %d", response.StatusCode))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	return nil
}

func TestFetchDashboardData() error {
	userID := getUserID()
	response, err := http.Get(url + "requests/" + userID)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("status code: %d", response.StatusCode))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	return nil
}

func getAPIKey() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	apiKey := os.Getenv("MONITOR_API_KEY")
	return apiKey
}

func getUserID() string {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	userID := os.Getenv("MONITOR_USER_ID")
	return userID
}

func getEmailBody(newUser error, fetchDashboardData error, fetchData error) string {
	var body strings.Builder
	if newUser != nil {
		body.WriteString(fmt.Sprintf("Error when creating new user: %s\n", newUser.Error()))
	}
	if fetchDashboardData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching dashboard data: %s\n", fetchDashboardData.Error()))
	}
	if fetchData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching API data: %s\n", fetchData.Error()))
	}
	return body.String()
}

func main() {
	newUserSuccessful := TestNewUser()
	fetchDashboardDataSuccessful := TestFetchDashboardData()
	fetchDataSuccessful := TestFetchData()
	address := email.GetEmailAddress()
	body := getEmailBody(newUserSuccessful, fetchDashboardDataSuccessful, fetchDataSuccessful)
	email.SendEmail("Error at API Analytics", body, address)
}
