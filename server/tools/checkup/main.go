package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/monitor"
	"github.com/tom-draper/api-analytics/server/tools/usage"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func emailBody(users []database.UserRow, requests []usage.UserCount, monitors []usage.UserCount, size string, connections int) string {
	return fmt.Sprintf("%d new users\n%d requests\n%d monitors\nDatabase size: %s\nActive database connections: %d", len(users), len(requests), len(monitors), size, connections)
}

func email_checkup() {
	users, err := usage.DailyUsers()
	if err != nil {
		panic(err)
	}
	requests, err := usage.DailyUserRequests()
	if err != nil {
		panic(err)
	}
	monitors, err := usage.DailyUserMonitors()
	if err != nil {
		panic(err)
	}
	size, err := usage.TableSize("requests")
	if err != nil {
		panic(err)
	}
	connections, err := usage.DatabaseConnections()
	if err != nil {
		panic(err)
	}
	body := emailBody(users, requests, monitors, size, connections)
	address := email.GetEmailAddress()
	email.SendEmail("API Analytics", body, address)
}

func print_checkup() {
	p := message.NewPrinter(language.English)

	apiDown := monitor.ServiceDown("api")
	if apiDown {
		color.Red("api: offline")
	} else {
		color.Green("api: live")
	}
	loggerDown := monitor.ServiceDown("logger")
	if loggerDown {
		color.Red("logger: offline")
	} else {
		color.Green("logger: live")
	}

	p.Println("---- Database --------------------")
	connections, err := usage.DatabaseConnections()
	if err != nil {
		panic(err)
	}
	p.Println("Active database connections:", connections)
	size, err := usage.TableSize("requests")
	if err != nil {
		panic(err)
	}
	p.Println("Database size:", size)
	columnSize, err := usage.RequestsColumnSize()
	if err != nil {
		panic(err)
	}
	columnSize.Display()

	p.Println("---- Last 24-hours ----------------")
	dailyUsers, err := usage.DailyUsersCount()
	if err != nil {
		panic(err)
	}
	p.Println("Users:", dailyUsers)
	dailyRequests, err := usage.DailyRequestsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Requests:", dailyRequests)
	dailyMonitors, err := usage.DailyMonitorsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Monitors:", dailyMonitors)

	p.Println("---- Last week --------------------")
	weeklyUsers, err := usage.WeeklyUsersCount()
	if err != nil {
		panic(err)
	}
	p.Println("Users:", weeklyUsers)
	weeklyRequests, err := usage.WeeklyRequestsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Requests:", weeklyRequests)
	weeklyMonitors, err := usage.WeeklyMonitorsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Monitors:", weeklyMonitors)

	p.Println("---- Total ------------------------")
	totalUsers, err := usage.UsersCount(0)
	if err != nil {
		panic(err)
	}
	p.Println("Users:", totalUsers)
	totalRequests, err := usage.RequestsCount(0)
	if err != nil {
		panic(err)
	}
	p.Println("Requests:", totalRequests)
	totalMonitors, err := usage.MonitorsCount(0)
	if err != nil {
		panic(err)
	}
	p.Println("Monitors:", totalMonitors)

	p.Println("---- Top Users --------------------")
	topUsers, err := usage.TopUsers(10)
	if err != nil {
		panic(err)
	}
	usage.DisplayUsers(topUsers)

	p.Println("---- Unused Users -----------------")
	unusedUsers, err := usage.UnusedUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(unusedUsers)

	p.Println("---- Users Since Last Request -----")
	sinceLastRequestUsers, err := usage.SinceLastRequestUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--email" {
		email_checkup()
	} else {
		print_checkup()
	}
}
