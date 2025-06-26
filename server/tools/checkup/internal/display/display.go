package display

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/lib"
	"github.com/tom-draper/api-analytics/server/tools/usage"
	"golang.org/x/text/message"
	"time"
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
	DisplayServicesTest()
	DisplayAPITest()
	DisplayLoggerTest()
	DisplayDatabaseStats(ctx, p)
	DisplayLastHour(ctx, p)
	DisplayLast24Hours(ctx, p)
	DisplayLastWeek(ctx, p)
	DisplayTotal(ctx, p)
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

func DisplayAPITest() {
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

func DisplayLoggerTest() {
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

func DisplayDatabaseStats(ctx context.Context, p *message.Printer) {
	printBanner("Database")

	connections, err := usage.DatabaseConnections(ctx)
	if HandleError(err) != nil { return }
	p.Println("Active database connections:", connections)

	size, err := usage.TableSize(ctx, "requests")
	if HandleError(err) != nil { return }
	p.Println("Database size:", size)
}

type timePeriodStat struct {
	label string
	count func(context.Context) (int, error)
}

func displayTimePeriodStats(ctx context.Context, p *message.Printer, bannerText string, stats []timePeriodStat) {
	printBanner(bannerText)
	for _, stat := range stats {
		count, err := stat.count(ctx)
		if HandleError(err) != nil { return }
		p.Println(stat.label+":", count)
	}
}

func DisplayLastHour(ctx context.Context, p *message.Printer) {
	hourlyStats := []timePeriodStat{
		{"Users", usage.HourlyUsersCount},
		{"Requests", usage.HourlyRequestsCount},
		{"Monitors", usage.HourlyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, "Last Hour", hourlyStats)
}

func DisplayLast24Hours(ctx context.Context, p *message.Printer) {
	dailyStats := []timePeriodStat{
		{"Users", usage.DailyUsersCount},
		{"Requests", usage.DailyRequestsCount},
		{"Monitors", usage.DailyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, "Last 24 Hours", dailyStats)
}

func DisplayLastWeek(ctx context.Context, p *message.Printer) {
	weeklyStats := []timePeriodStat{
		{"Users", usage.WeeklyUsersCount},
		{"Requests", usage.WeeklyRequestsCount},
		{"Monitors", usage.WeeklyMonitorsCount},
	}
	displayTimePeriodStats(ctx, p, "Last Week", weeklyStats)
}

func DisplayDatabaseCheckup(ctx context.Context, p *message.Printer) {
	DisplayDatabaseStats(ctx, p)
	totalRequests, err := usage.RequestsCount(ctx, "")
	if HandleError(err) != nil { return }
	p.Println("Requests:", totalRequests)
	DisplayDatabaseTableStats(ctx)
}

func DisplayDatabaseTableStats(ctx context.Context) {
	printBanner("Requests Fields")
	columnSize, err := usage.RequestsColumnSize(ctx)
	if HandleError(err) != nil { return }
	columnSize.Display()
}

func DisplayUsersCheckup(ctx context.Context, p *message.Printer) {
	DisplayTopUsers(ctx, p)
	DisplayUnusedUsers(ctx, p)
	DisplayUsersSinceLastRequest(ctx, p)
}

func DisplayTotal(ctx context.Context, p *message.Printer) {
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
		count, err := stat.count(ctx, "")
		if HandleError(err) != nil { return }
		p.Println(stat.label+":", count)
	}
}

func DisplayTopUsers(ctx context.Context, p *message.Printer) {
	printBanner("Top Users")
	topUsers, err := usage.TopUsers(ctx, 10)
	if HandleError(err) != nil { return }
	usage.DisplayUsers(topUsers)
}

func DisplayUnusedUsers(ctx context.Context, p *message.Printer) {
	printBanner("Unused Users")
	unusedUsers, err := usage.UnusedUsers(ctx)
	if HandleError(err) != nil { return }
	usage.DisplayUserTimes(unusedUsers)
}

func DisplayUsersSinceLastRequest(ctx context.Context, p *message.Printer) {
	printBanner("Users Since Last Request")
	sinceLastRequestUsers, err := usage.SinceLastRequestUsers(ctx)
	if HandleError(err) != nil { return }
	usage.DisplayUserTimes(sinceLastRequestUsers)
}

func DisplayMonitorsCheckup(ctx context.Context, p *message.Printer) {
	DisplayMonitors(ctx, p)
}

func DisplayMonitors(ctx context.Context, p *message.Printer) {
	printBanner("Monitors")
	monitors, err := usage.TotalMonitors(ctx)
	if HandleError(err) != nil { return }

	for i, monitor := range monitors {
		fmt.Printf("[%d] %s %s %s\n", i, monitor.APIKey, monitor.CreatedAt.Format("2006-01-02 15:04:05"), monitor.URL)
	}
}