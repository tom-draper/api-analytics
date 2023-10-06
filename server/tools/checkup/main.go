package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/lib"
	"github.com/tom-draper/api-analytics/server/tools/usage"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func emailBody(users []database.UserRow, requests []usage.UserCount, monitors []usage.UserCount, size string, connections int) string {
	return fmt.Sprintf("%d new users\n%d requests\n%d monitors\nDatabase size: %s\nActive database connections: %d", len(users), len(requests), len(monitors), size, connections)
}

func emailCheckup() {
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

func printCheckup() {
	printServicesTest()
	printAPITest()

	printDatabaseStats()
	printLastHour()
	printLast24Hours()
	printLastWeek()
	printTotal()
}

func printUsersCheckup() {
	printTopUsers()
	printUnusedUsers()
	printUsersSinceLastRequest()
}

func printServicesTest() {
	fmt.Println("---- Services ------------------------")
	apiDown := monitor.ServiceDown("api")
	fmt.Printf("api: ")
	if apiDown {
		color.Red("offline")
	} else {
		color.Green("online")
	}
	loggerDown := monitor.ServiceDown("logger")
	fmt.Printf("logger: ")
	if loggerDown {
		color.Red("offline")
	} else {
		color.Green("online")
	}
	nginxDown := monitor.ServiceDown("nginx")
	fmt.Printf("nginx: ")
	if nginxDown {
		color.Red("offline")
	} else {
		color.Green("online")
	}
	postgresqlDown := monitor.ServiceDown("postgresql")
	fmt.Printf("postgresql: ")
	if postgresqlDown {
		color.Red("offline")
	} else {
		color.Green("online")
	}
}

func printAPITest() {
	fmt.Println("---- API -----------------------------")
	var err error
	err = monitor.TryNewUser()
	fmt.Printf("/generate-api-key ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}
	err = monitor.TryFetchDashboardData()
	fmt.Printf("/requests/<user-id> ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}
	err = monitor.TryFetchData()
	fmt.Printf("/data ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}
	err = monitor.TryFetchUserID()
	fmt.Printf("/user-id/<api-key> ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}
	err = monitor.TryFetchMonitorPings()
	fmt.Printf("/monitor/pings/<user-id> ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}
	// err = monitor.TryLogRequests()
	// fmt.Printf("/log-request ")
	// if err != nil {
	// 	color.Red("offline")
	// 	fmt.Println(err)
	// } else {
	// 	color.Green("online")
	// }
}

func printDatabaseStats() {
	p := message.NewPrinter(language.English)
	p.Println("---- Database ------------------------")
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
	// columnSize, err := usage.RequestsColumnSize()
	// if err != nil {
	// 	panic(err)
	// }
	// columnSize.Display()
}

func printLastHour() {
	p := message.NewPrinter(language.English)
	p.Println("---- Last hour -----------------------")
	dailyUsers, err := usage.HourlyUsersCount()
	if err != nil {
		panic(err)
	}
	p.Println("Users:", dailyUsers)
	dailyRequests, err := usage.HourlyRequestsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Requests:", dailyRequests)
	dailyMonitors, err := usage.HourlyMonitorsCount()
	if err != nil {
		panic(err)
	}
	p.Println("Monitors:", dailyMonitors)
}

func printLast24Hours() {
	p := message.NewPrinter(language.English)
	p.Println("---- Last 24-hours -------------------")
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
}

func printLastWeek() {
	p := message.NewPrinter(language.English)
	p.Println("---- Last week -----------------------")
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
}

func printTotal() {
	p := message.NewPrinter(language.English)
	p.Println("---- Total ---------------------------")
	totalUsers, err := usage.UsersCount("")
	if err != nil {
		panic(err)
	}
	p.Println("Users:", totalUsers)
	totalRequests, err := usage.RequestsCount("")
	if err != nil {
		panic(err)
	}
	p.Println("Requests:", totalRequests)
	totalMonitors, err := usage.MonitorsCount("")
	if err != nil {
		panic(err)
	}
	p.Println("Monitors:", totalMonitors)
}

func printTopUsers() {
	fmt.Println("---- Top Users -----------------------")
	topUsers, err := usage.TopUsers(10)
	if err != nil {
		panic(err)
	}
	usage.DisplayUsers(topUsers)
}

func printUnusedUsers() {
	fmt.Println("---- Unused Users --------------------")
	unusedUsers, err := usage.UnusedUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(unusedUsers)
}

func printUsersSinceLastRequest() {
	fmt.Println("---- Users Since Last Request --------")
	sinceLastRequestUsers, err := usage.SinceLastRequestUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--email" {
		emailCheckup()
	} else if len(os.Args) > 1 && os.Args[1] == "--users" {
		printUsersCheckup()
	} else {
		printCheckup()
	}
}
