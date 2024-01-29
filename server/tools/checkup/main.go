package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/lib"
	"github.com/tom-draper/api-analytics/server/tools/usage"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func printBanner(text string) {
	fmt.Printf("---- %s %s\n", text, strings.Repeat("-", 35-len(text)))
}

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

func displayCheckup() {
	displayServicesTest()
	displayAPITest()
	displayLoggerTest()

	displayDatabaseStats()
	displayLastHour()
	displayLast24Hours()
	displayLastWeek()
	displayTotal()
}

func displayServicesTest() {
	printBanner("Services")

	testService("api")
	testService("logger")
	testService("nginx")
	testService("postgresql")
}

func testService(service string) {
	down := monitor.ServiceDown(service)
	fmt.Printf("%s: ", service)
	if down {
		color.Red("offline")
	} else {
		color.Green("online")
	}
}

func displayAPITest() {
	printBanner("API")

	testAPIEndpoint("/generate-api-key", monitor.TryNewUser)
	testAPIEndpoint("/requests/<user-id>", monitor.TryFetchDashboardData)
	testAPIEndpoint("/data", monitor.TryFetchData)
	testAPIEndpoint("/user-id/<api-key>", monitor.TryFetchUserID)
	testAPIEndpoint("/monitor/pings/<user-id>", monitor.TryFetchMonitorPings)
}

func testAPIEndpoint(endpoint string, testEndpoint func() error) {
	start := time.Now()
	err := testEndpoint()
	fmt.Printf("%s ", endpoint)
	if err != nil {
		color.New(color.FgRed).Printf("offline")
		fmt.Printf("\n%s\n", err.Error())
	} else {
		color.New(color.FgGreen).Printf("online")
		fmt.Printf(" %s\n", time.Since(start))
	}
}

func displayLoggerTest() {
	printBanner("Logger")

	testLoggerEndpoint("/log-request", monitor.TryLogRequests, true)
	testLoggerEndpoint("/requests", monitor.TryLogRequests, false)
}

func testLoggerEndpoint(endpoint string, testEndpoint func(legacy bool) error, legacy bool) {
	start := time.Now()
	err := testEndpoint(legacy)
	fmt.Printf("%s ", endpoint)
	if err != nil {
		color.New(color.FgRed).Printf("offline")
		fmt.Printf("\n%s\n", err.Error())
	} else {
		color.New(color.FgGreen).Printf("online")
		fmt.Printf(" %s\n", time.Since(start))
	}
}

func displayDatabaseStats() {
	p := message.NewPrinter(language.English)
	printBanner("Database")
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

func displayLastHour() {
	p := message.NewPrinter(language.English)
	printBanner("Last Hour")
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

func displayLast24Hours() {
	p := message.NewPrinter(language.English)
	printBanner("Last 24 Hours")
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

func displayLastWeek() {
	p := message.NewPrinter(language.English)
	printBanner("Last Week")
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

func displayUsersCheckup() {
	displayTopUsers()
	displayUnusedUsers()
	displayUsersSinceLastRequest()
}

func displayTotal() {
	p := message.NewPrinter(language.English)
	printBanner("Total")
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

func displayTopUsers() {
	printBanner("Top Users")
	topUsers, err := usage.TopUsers(10)
	if err != nil {
		panic(err)
	}
	usage.DisplayUsers(topUsers)
}

func displayUnusedUsers() {
	printBanner("Unused Users")
	unusedUsers, err := usage.UnusedUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(unusedUsers)
}

func displayUsersSinceLastRequest() {
	printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := usage.SinceLastRequestUsers()
	if err != nil {
		panic(err)
	}
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func displayMonitorsCheckup() {
	displayMonitors()
}

func displayMonitors() {
	printBanner("Monitors")
	monitors, err := usage.TotalMonitors()
	if err != nil {
		panic(err)
	}
	for i, monitor := range monitors {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}

type Options struct {
	email    bool
	users    bool
	monitors bool
	help     bool
}

func getOptions() Options {
	options := Options{}
	for _, arg := range os.Args {
		if arg == "--email" {
			options.email = true
		} else if arg == "--users" {
			options.users = true
		} else if arg == "--monitors" {
			options.monitors = true
		} else if arg == "--help" {
			options.help = true
		}
	}
	return options
}

func displayHelp() {
	fmt.Printf("Checkup - A command-line tool for checking resource usage.\n\nOptions:\n`--users` show user account usage\n`--monitors` show monitor usage\n`--email` email the summary instead of printing to console\n`--help` to display help\n")
}

func main() {
	options := getOptions()
	if options.help {
		displayHelp()
	} else if options.email {
		emailCheckup()
	} else if options.users {
		displayUsersCheckup()
	} else if options.monitors {
		displayMonitorsCheckup()
	} else {
		displayCheckup()
	}
}
