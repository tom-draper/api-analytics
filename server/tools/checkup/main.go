package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"log"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/email"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/lib"
	"github.com/tom-draper/api-analytics/server/tools/usage"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func printBanner(text string) {
	fmt.Printf("---- %s %s\n", text, strings.Repeat("-", 35-len(text)))
}

func emailBody(users []usage.UserRow, requests []usage.UserCount, monitors []usage.UserCount, size string, connections int) string {
	return fmt.Sprintf("%d new users\n%d requests\n%d monitors\nDatabase size: %s\nActive database connections: %d",
		len(users), len(requests), len(monitors), size, connections)
}

func emailCheckup(ctx context.Context) {
	users, err := usage.DailyUsers(context.Background())
	handleError(err)

	requests, err := usage.DailyUserRequests(context.Background())
	handleError(err)

	monitors, err := usage.DailyUserMonitors(context.Background())
	handleError(err)

	size, err := usage.TableSize(context.Background(), "requests")
	handleError(err)

	connections, err := usage.DatabaseConnections(context.Background())
	handleError(err)

	body := emailBody(users, requests, monitors, size, connections)
	address := email.GetEmailAddress()
	email.SendEmail("API Analytics", body, address)
}

func handleError(err error) {
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
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
	services := []string{"api", "logger", "nginx", "postgresql"}
	for _, service := range services {
		testService(service)
	}
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
	endpoints := []struct {
		endpoint string
		testFunc func() error
	}{
		{"/generate-api-key", monitor.TryNewUser},
		{"/requests/<user-id>", monitor.TryFetchDashboardData},
		{"/data", monitor.TryFetchData},
		{"/user-id/<api-key>", monitor.TryFetchUserID},
		{"/monitor/pings/<user-id>", monitor.TryFetchMonitorPings},
	}
	for _, ep := range endpoints {
		testAPIEndpoint(ep.endpoint, ep.testFunc)
	}
}

func testAPIEndpoint(endpoint string, testEndpoint func() error) {
	start := time.Now()
	err := testEndpoint()
	fmt.Printf("%s ", endpoint)
	if err != nil {
		color.New(color.FgRed).Printf("offline\n")
		fmt.Printf("\n%s\n", err.Error())
	} else {
		color.New(color.FgGreen).Printf("online\n")
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
		color.New(color.FgRed).Printf("offline\n")
		fmt.Printf("\n%s\n", err.Error())
	} else {
		color.New(color.FgGreen).Printf("online\n")
		fmt.Printf(" %s\n", time.Since(start))
	}
}

func displayDatabaseStats() {
	p := message.NewPrinter(language.English)
	printBanner("Database")

	connections, err := usage.DatabaseConnections(context.Background())
	handleError(err)
	p.Println("Active database connections:", connections)

	size, err := usage.TableSize(context.Background(), "requests")
	handleError(err)
	p.Println("Database size:", size)
}

func displayLastHour() {
	p := message.NewPrinter(language.English)
	printBanner("Last Hour")

	hourlyStats := []struct {
		label string
		count func(context.Context) (int, error)
	}{
		{"Users", usage.HourlyUsersCount},
		{"Requests", usage.HourlyRequestsCount},
		{"Monitors", usage.HourlyMonitorsCount},
	}

	for _, stat := range hourlyStats {
		count, err := stat.count(context.Background())
		handleError(err)
		p.Println(stat.label+":", count)
	}
}

func displayLast24Hours() {
	p := message.NewPrinter(language.English)
	printBanner("Last 24 Hours")

	dailyStats := []struct {
		label string
		count func(context.Context) (int, error)
	}{
		{"Users", usage.DailyUsersCount},
		{"Requests", usage.DailyRequestsCount},
		{"Monitors", usage.DailyMonitorsCount},
	}

	for _, stat := range dailyStats {
		count, err := stat.count(context.Background())
		handleError(err)
		p.Println(stat.label+":", count)
	}
}

func displayLastWeek() {
	p := message.NewPrinter(language.English)
	printBanner("Last Week")

	weeklyStats := []struct {
		label string
		count func(context.Context) (int, error)
	}{
		{"Users", usage.WeeklyUsersCount},
		{"Requests", usage.WeeklyRequestsCount},
		{"Monitors", usage.WeeklyMonitorsCount},
	}

	for _, stat := range weeklyStats {
		count, err := stat.count(context.Background())
		handleError(err)
		p.Println(stat.label+":", count)
	}
}

func displayDatabaseCheckup() {
	displayDatabaseStats()
	p := message.NewPrinter(language.English)
	totalRequests, err := usage.RequestsCount(context.Background(), "")
	if err != nil {
		handleError(err)
	}
	p.Println("Requests:", totalRequests)
	displayDatabaseTableStats()
}

func displayDatabaseTableStats() {
	printBanner("Requests Fields")
	columnSize, err := usage.RequestsColumnSize(context.Background())
	if err != nil {
		handleError(err)
	}
	columnSize.Display()
}

func displayUsersCheckup() {
	displayTopUsers()
	displayUnusedUsers()
	displayUsersSinceLastRequest()
}

func displayTotal() {
	p := message.NewPrinter(language.English)
	printBanner("Total")

	totalStats := []struct {
		label string
		count func(context.Context, string) (int, error)
	}{
		{"Users", usage.UsersCount},
		{"Requests", usage.RequestsCount},
		{"Monitors", usage.MonitorsCount},
	}

	for _, stat := range totalStats {
		count, err := stat.count(context.Background(), "")
		handleError(err)
		p.Println(stat.label+":", count)
	}
}

func displayTopUsers() {
	printBanner("Top Users")
	topUsers, err := usage.TopUsers(context.Background(), 10)
	handleError(err)
	usage.DisplayUsers(topUsers)
}

func displayUnusedUsers() {
	printBanner("Unused Users")
	unusedUsers, err := usage.UnusedUsers(context.Background())
	handleError(err)
	usage.DisplayUserTimes(unusedUsers)
}

func displayUsersSinceLastRequest() {
	printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := usage.SinceLastRequestUsers(context.Background())
	handleError(err)
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func displayMonitorsCheckup() {
	displayMonitors()
}

func displayMonitors() {
	printBanner("Monitors")
	monitors, err := usage.TotalMonitors(context.Background())
	handleError(err)

	for i, monitor := range monitors {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}

type Options struct {
	email    bool
	users    bool
	monitors bool
	database bool
	help     bool
}

func getOptions() Options {
	options := Options{}
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--email":
			options.email = true
		case "--users":
			options.users = true
		case "--monitors":
			options.monitors = true
		case "--database":
			options.database = true
		case "--help":
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
		return
	}

	ctx := context.Background()

	if options.email {
		emailCheckup(ctx)
	} else if options.users {
		displayUsersCheckup()
	} else if options.monitors {
		displayMonitorsCheckup()
	} else if options.database {
		displayDatabaseCheckup()
	} else {
		displayCheckup()
	}
}
