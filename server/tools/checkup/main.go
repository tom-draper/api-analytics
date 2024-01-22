package main

import (
	"fmt"
	"os"
	"strings"

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

func displayAPITest() {
	printBanner("API")
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
}

func displayLoggerTest() {
	printBanner("Logger")
	err := monitor.TryLogRequests(false)
	fmt.Printf("/log-request (legacy)")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
	}

	err = monitor.TryLogRequests(true)
	fmt.Printf("/requests ")
	if err != nil {
		color.Red("offline")
		fmt.Println(err)
	} else {
		color.Green("online")
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
