package main

import (
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	p := message.NewPrinter(language.English)

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
	columnSize := usage.RequestsColumnSize()
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
