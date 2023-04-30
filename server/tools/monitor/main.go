package main

import (
	"io/ioutil"
	"net/http"

	"github.com/tom-draper/api-analytics/server/database"
)

func TestNewUser() bool {
	apiKey, err := createNewUser()
	if err != nil {
		return false
	}
	err = database.DeleteUser(apiKey)
	if err != nil {
		return false
	}
	return true
}

func createNewUser() (string, error) {
	response, err := http.Get("https://apianalytics-server.com/api/generate-api-key")
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	apiKey := sb[1 : len(sb)-1]
	return apiKey, nil
}

func main() {
	TestNewUser()
}
