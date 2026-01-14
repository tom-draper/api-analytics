package display

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tom-draper/api-analytics/server/database"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/pkg"
	"github.com/tom-draper/api-analytics/server/tools/usage/monitors"
	"github.com/tom-draper/api-analytics/server/tools/usage/requests"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
	"github.com/tom-draper/api-analytics/server/tools/usage/users"
	"golang.org/x/text/message"
)

func printBanner(text string) {
	fmt.Printf("---- %s %s\n", text, strings.Repeat("-", 35-len(text)))
}

func HandleError(err error) error {
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	return err
}

func DisplayCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	DisplayServicesTest()
	DisplayAPITest(db)
	DisplayLoggerTest()
	DisplayDatabaseStats(ctx, p, db)
	DisplayLastHour(ctx, p, db)
	DisplayLast24Hours(ctx, p, db)
	DisplayLastWeek(ctx, p, db)
	DisplayTotal(ctx, p, db)
}

func initDatabase(ctx context.Context) *database.DB {
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Println("POSTGRES_URL environment variable not set")
		return nil
	}

	db, err := database.New(ctx, dbURL)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return nil
	}

	return db
}

func DisplayServicesTest() {
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

func DisplayAPITest(db *database.DB) {
	printBanner("API")
	apiBaseURL := os.Getenv("API_BASE_URL")
	monitorAPIKey := os.Getenv("MONITOR_API_KEY")
	monitorUserID := os.Getenv("MONITOR_USER_ID")

	if apiBaseURL == "" || monitorAPIKey == "" || monitorUserID == "" {
		log.Println("API test environment variables not set")
		return
	}

	endpoints := []struct {
		endpoint string
		testFunc func() error
	}{
		{"/generate-api-key", func() error { return monitor.TryNewUserWithParams(apiBaseURL, monitorAPIKey, monitorUserID, db) }},
		{"/requests/<user-id>", func() error { return monitor.TryFetchDashboardDataWithParams(apiBaseURL, monitorAPIKey, monitorUserID) }},
		{"/data", func() error { return monitor.TryFetchDataWithParams(apiBaseURL, monitorAPIKey, monitorUserID) }},
		{"/user-id/<api-key>", func() error { return monitor.TryFetchUserIDWithParams(apiBaseURL, monitorAPIKey, monitorUserID) }},
		{"/monitor/pings/<user-id>", func() error { return monitor.TryFetchMonitorPingsWithParams(apiBaseURL, monitorAPIKey, monitorUserID) }},
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

func DisplayLoggerTest() {
	printBanner("Logger")
	apiBaseURL := os.Getenv("API_BASE_URL")
	monitorAPIKey := os.Getenv("MONITOR_API_KEY")
	monitorUserID := os.Getenv("MONITOR_USER_ID")

	if apiBaseURL == "" || monitorAPIKey == "" || monitorUserID == "" {
		log.Println("Logger test environment variables not set")
		return
	}

	testLoggerEndpoint("/log-request", func(legacy bool) error {
		return monitor.TryLogRequestsWithParams(apiBaseURL, monitorAPIKey, monitorUserID, legacy)
	}, true)
	testLoggerEndpoint("/requests", func(legacy bool) error {
		return monitor.TryLogRequestsWithParams(apiBaseURL, monitorAPIKey, monitorUserID, legacy)
	}, false)
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

func DisplayDatabaseStats(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Database")

	connections, err := usage.DatabaseConnections(ctx, db)
	if HandleError(err) != nil {
		return
	}
	p.Println("Active database connections:", connections)

	size, err := usage.TableSize(ctx, db, "requests")
	if HandleError(err) != nil {
		return
	}
	p.Println("Database size:", size)
}

type timePeriodStat struct {
	label string
	count func(context.Context, *database.DB) (int, error)
}

func displayTimePeriodStats(ctx context.Context, p *message.Printer, db *database.DB, bannerText string, stats []timePeriodStat) {
	printBanner(bannerText)
	for _, stat := range stats {
		count, err := stat.count(ctx, db)
		if HandleError(err) != nil {
			return
		}
		p.Println(stat.label+":", count)
	}
}

func DisplayLastHour(ctx context.Context, p *message.Printer, db *database.DB) {
	hourlyStats := []timePeriodStat{
		{"Users", users.HourlyUsersCount},
		{"Requests", requests.HourlyRequestsCount},
		{"Monitors", monitors.HourlyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, db, "Last Hour", hourlyStats)
}

func DisplayLast24Hours(ctx context.Context, p *message.Printer, db *database.DB) {
	dailyStats := []timePeriodStat{
		{"Users", users.DailyUsersCount},
		{"Requests", requests.DailyRequestsCount},
		{"Monitors", monitors.DailyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, db, "Last 24 Hours", dailyStats)
}

func DisplayLastWeek(ctx context.Context, p *message.Printer, db *database.DB) {
	weeklyStats := []timePeriodStat{
		{"Users", users.WeeklyUsersCount},
		{"Requests", requests.WeeklyRequestsCount},
		{"Monitors", monitors.WeeklyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, db, "Last Week", weeklyStats)
}

func DisplayDatabaseCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	DisplayDatabaseStats(ctx, p, db)
	totalRequests, err := requests.RequestsCount(ctx, db, "")
	if HandleError(err) != nil {
		return
	}
	p.Println("Requests:", totalRequests)
	DisplayDatabaseTableStats(ctx, db)
}

func DisplayDatabaseTableStats(ctx context.Context, db *database.DB) {
	printBanner("Requests Fields")
	columnSize, err := requests.RequestsColumnSize(ctx, db)
	if HandleError(err) != nil {
		return
	}
	columnSize.Display()
}

func DisplayUsersCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	DisplayTopUsers(ctx, p, db)
	DisplayUnusedUsers(ctx, p, db)
	DisplayUsersSinceLastRequest(ctx, p, db)
}

func DisplayTotal(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Total")

	totalStats := []struct {
		label string
		count func(context.Context, *database.DB, string) (int, error)
	}{
		{"Users", users.UsersCount},
		{"Requests", requests.RequestsCount},
		{"Monitors", monitors.MonitorsCount},
	}

	for _, stat := range totalStats {
		count, err := stat.count(ctx, db, "")
		if HandleError(err) != nil {
			return
		}
		p.Println(stat.label+":", count)
	}
}

func DisplayTopUsers(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Top Users")
	topUsers, err := users.TopUsers(ctx, db, 10)
	if HandleError(err) != nil {
		return
	}
	users.DisplayUsers(topUsers)
}

func DisplayUnusedUsers(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Unused Users")
	unusedUsers, err := users.UnusedUsers(ctx, db)
	if HandleError(err) != nil {
		return
	}
	users.DisplayUserTimes(unusedUsers)
}

func DisplayUsersSinceLastRequest(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := users.SinceLastRequestUsers(ctx, db)
	if HandleError(err) != nil {
		return
	}
	users.DisplayUserTimes(sinceLastRequestUsers)
}

func DisplayMonitorsCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	DisplayMonitors(ctx, p, db)
}

func DisplayMonitors(ctx context.Context, p *message.Printer, db *database.DB) {
	printBanner("Monitors")
	monitorsList, err := monitors.TotalMonitors(ctx, db)
	if HandleError(err) != nil {
		return
	}

	for i, monitor := range monitorsList {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}