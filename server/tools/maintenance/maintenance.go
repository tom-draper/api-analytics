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

	err = database.DeleteUser(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'users'.")
	err = database.DeleteRequests(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'requests'.")
	err = database.DeleteMonitors(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'monitors'.")
	err = database.DeletePings(apiKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("User from table 'pings'.")

	fmt.Println("User deletion successful.")
}
