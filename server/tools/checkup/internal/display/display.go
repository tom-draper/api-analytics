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
	"github.com/tom-draper/api-analytics/server/tools/usage"
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

	client := usage.NewClient(db)

	DisplayServicesTest()
	DisplayAPITest(db)
	DisplayLoggerTest()
	DisplayDatabaseStats(ctx, p, client)
	DisplayLastHour(ctx, p, client)
	DisplayLast24Hours(ctx, p, client)
	DisplayLastWeek(ctx, p, client)
	DisplayTotal(ctx, p, client)
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

func DisplayDatabaseStats(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Database")

	connections, err := client.DatabaseConnections(ctx)
	if HandleError(err) != nil {
		return
	}
	p.Println("Active database connections:", connections)

	size, err := client.TableSize(ctx, "requests")
	if HandleError(err) != nil {
		return
	}
	p.Println("Database size:", size)
}

type timePeriodStat struct {
	label string
	count func(context.Context, *usage.Client) (int, error)
}

func displayTimePeriodStats(ctx context.Context, p *message.Printer, client *usage.Client, bannerText string, stats []timePeriodStat) {
	printBanner(bannerText)
	for _, stat := range stats {
		count, err := stat.count(ctx, client)
		if HandleError(err) != nil {
			return
		}
		p.Println(stat.label+":", count)
	}
}

func DisplayLastHour(ctx context.Context, p *message.Printer, client *usage.Client) {
	hourlyStats := []timePeriodStat{
		{"Users", func(ctx context.Context, c *usage.Client) (int, error) { return c.HourlyUsersCount(ctx) }},
		{"Requests", func(ctx context.Context, c *usage.Client) (int, error) { return c.HourlyRequestsCount(ctx) }},
		{"Monitors", func(ctx context.Context, c *usage.Client) (int, error) { return c.HourlyMonitorsCount(ctx) }},
	}
	displayTimePeriodStats(ctx, p, client, "Last Hour", hourlyStats)
}

func DisplayLast24Hours(ctx context.Context, p *message.Printer, client *usage.Client) {
	dailyStats := []timePeriodStat{
		{"Users", func(ctx context.Context, c *usage.Client) (int, error) { return c.DailyUsersCount(ctx) }},
		{"Requests", func(ctx context.Context, c *usage.Client) (int, error) { return c.DailyRequestsCount(ctx) }},
		{"Monitors", func(ctx context.Context, c *usage.Client) (int, error) { return c.DailyMonitorsCount(ctx) }},
	}
	displayTimePeriodStats(ctx, p, client, "Last 24 Hours", dailyStats)
}

func DisplayLastWeek(ctx context.Context, p *message.Printer, client *usage.Client) {
	weeklyStats := []timePeriodStat{
		{"Users", func(ctx context.Context, c *usage.Client) (int, error) { return c.WeeklyUsersCount(ctx) }},
		{"Requests", func(ctx context.Context, c *usage.Client) (int, error) { return c.WeeklyRequestsCount(ctx) }},
		{"Monitors", func(ctx context.Context, c *usage.Client) (int, error) { return c.WeeklyMonitorsCount(ctx) }},
	}
	displayTimePeriodStats(ctx, p, client, "Last Week", weeklyStats)
}

func DisplayDatabaseCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	client := usage.NewClient(db)

	DisplayDatabaseStats(ctx, p, client)
	totalRequests, err := client.RequestsCount(ctx, "")
	if HandleError(err) != nil {
		return
	}
	p.Println("Requests:", totalRequests)
	DisplayDatabaseTableStats(ctx, client)
}

func DisplayDatabaseTableStats(ctx context.Context, client *usage.Client) {
	printBanner("Requests Fields")
	columnSize, err := client.RequestsColumnSize(ctx)
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

	client := usage.NewClient(db)

	DisplayTopUsers(ctx, p, client)
	DisplayUnusedUsers(ctx, p, client)
	DisplayUsersSinceLastRequest(ctx, p, client)
}

func DisplayTotal(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Total")

	totalStats := []struct {
		label string
		count func(context.Context, *usage.Client, string) (int, error)
	}{
		{"Users", func(ctx context.Context, c *usage.Client, _ string) (int, error) { return c.UsersCount(ctx, "") }},
		{"Requests", func(ctx context.Context, c *usage.Client, _ string) (int, error) { return c.RequestsCount(ctx, "") }},
		{"Monitors", func(ctx context.Context, c *usage.Client, _ string) (int, error) { return c.MonitorsCount(ctx, "") }},
	}

	for _, stat := range totalStats {
		count, err := stat.count(ctx, client, "")
		if HandleError(err) != nil {
			return
		}
		p.Println(stat.label+":", count)
	}
}

func DisplayTopUsers(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Top Users")
	topUsers, err := client.TopUsers(ctx, 10)
	if HandleError(err) != nil {
		return
	}
	usage.DisplayUsers(topUsers)
}

func DisplayUnusedUsers(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Unused Users")
	unusedUsers, err := client.UnusedUsers(ctx)
	if HandleError(err) != nil {
		return
	}
	usage.DisplayUserTimes(unusedUsers)
}

func DisplayUsersSinceLastRequest(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := client.SinceLastRequestUsers(ctx)
	if HandleError(err) != nil {
		return
	}
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func DisplayMonitorsCheckup(ctx context.Context, p *message.Printer) {
	db := initDatabase(ctx)
	if db == nil {
		return
	}
	defer db.Close()

	client := usage.NewClient(db)

	DisplayMonitors(ctx, p, client)
}

func DisplayMonitors(ctx context.Context, p *message.Printer, client *usage.Client) {
	printBanner("Monitors")
	monitorsList, err := client.TotalMonitors(ctx)
	if HandleError(err) != nil {
		return
	}

	for i, monitor := range monitorsList {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}