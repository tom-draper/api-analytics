package main

import (
	"fmt"
	"github.com/tom-draper/api-analytics/server/database"
	"strings"
)

func DeleteUser(apiKey string) {
	fmt.Println("Delete API key '%s' from the database? Y/n", apiKey)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	response = strings.ToLower(response)
	if response != "y" && response != "yes" {
		fmt.Println("User deletion cancelled.")
		return
	}

	err = deleteUser(apiKey)
	if err != nil {
		panic(err)
	}
	err = deleteRequests(apiKey)
	if err != nil {
		panic(err)
	}
	err = deleteMonitors(apiKey)
	if err != nil {
		panic(err)
	}
	err = deletePings(apiKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("User deletion successful.")
}

func deleteUser(apiKey string) error {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM user WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func deleteRequests(apiKey string) error {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM requests WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func deleteMonitors(apiKey string) error {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM monitor WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}

func deletePings(apiKey string) error {
	db := database.OpenDBConnection()

	query := fmt.Sprintf("DELETE FROM pings WHERE api_key = '%s';", apiKey)
	_, err := db.Query(query)
	return err
}
