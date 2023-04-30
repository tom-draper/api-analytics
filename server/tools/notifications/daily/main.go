package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

func emailBody(users []database.UserRow, usage []usage.UserCount) string {
	return fmt.Sprintf("%d new users\n%d requests", len(users), len(usage))
}

func main() {
	users, err := usage.DailyUsers()
	if err != nil {
		panic(err)
	}
	usage, err := usage.DailyUsage()
	if err != nil {
		panic(err)
	}
	monitors, err := usage.DailyMonitors()
	if err != nil {
		panic(err)
	}
	size, err := usage.DatabaseSize()
	if err != nil {
		panic(err)
	}
	connections, err := usage.DatabaseConnections()
	if err != nil {
		panic(err)
	}
	body := emailBody(users, usage)
	address := email.GetEmailAddress()
	email.SendEmail("API Analytics", body, address)
}
