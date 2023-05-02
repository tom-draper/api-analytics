package main

import (
	"fmt"

	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

func main() {
	connections, err := usage.DatabaseConnections()
	if err != nil {
		panic(err)
	}
	fmt.Println("Active database connections:", connections)
	users, err := usage.DailyUsers()
	if err != nil {
		panic(err)
	}

	fmt.Println("New users:", len(users))
	requests, err := usage.DailyUsage()
	if err != nil {
		panic(err)
	}
	fmt.Println("Requests:", len(requests))
	monitors, err := usage.DailyMonitors()
	if err != nil {
		panic(err)
	}
	fmt.Println("New monitors:", len(monitors))
	totalUsers, err := usage.TotalUsers()
	if err != nil {
		panic(err)
	}
	fmt.Println("Total users:", totalUsers)
	totalRequests, err := usage.TotalRequests()
	if err != nil {
		panic(err)
	}
	fmt.Println("Total requests:", totalRequests)
	size, err := usage.DatabaseSize()
	if err != nil {
		panic(err)
	}
	fmt.Println("Database size:", size)
}
