package checkup

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	monitor "github.com/tom-draper/api-analytics/server/tools/monitor/pkg"
)

func (c *Client) printBanner(text string) {
	fmt.Printf("---- %s %s\n", text, strings.Repeat("-", 35-len(text)))
}

// DisplayServicesTest displays the status of all services
func (c *Client) DisplayServicesTest() {
	c.printBanner("Services")
	services := []string{"api", "logger", "nginx", "postgresql"}
	for _, service := range services {
		c.testService(service)
	}
}

func (c *Client) testService(service string) {
	down := monitor.ServiceDown(service)
	fmt.Printf("%s: ", service)
	if down {
		color.Red("offline")
	} else {
		color.Green("online")
	}
}

// DisplayAPITest displays the status of API endpoints
func (c *Client) DisplayAPITest() {
	c.printBanner("API")

	if !c.cfg.HasAPIConfig() {
		log.Println("API test environment variables not set")
		return
	}

	endpoints := []struct {
		endpoint string
		testFunc func() error
	}{
		{"/generate-api-key", func() error {
			return monitor.TryNewUserWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID, c.db)
		}},
		{"/requests/<user-id>", func() error {
			return monitor.TryFetchDashboardDataWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID)
		}},
		{"/data", func() error {
			return monitor.TryFetchDataWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID)
		}},
		{"/user-id/<api-key>", func() error {
			return monitor.TryFetchUserIDWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID)
		}},
		{"/monitor/pings/<user-id>", func() error {
			return monitor.TryFetchMonitorPingsWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID)
		}},
	}
	for _, ep := range endpoints {
		c.testAPIEndpoint(ep.endpoint, ep.testFunc)
	}
}

func (c *Client) testAPIEndpoint(endpoint string, testEndpoint func() error) {
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

// DisplayLoggerTest displays the status of logger endpoints
func (c *Client) DisplayLoggerTest() {
	c.printBanner("Logger")

	if !c.cfg.HasAPIConfig() {
		log.Println("Logger test environment variables not set")
		return
	}

	c.testLoggerEndpoint("/log-request", func(legacy bool) error {
		return monitor.TryLogRequestsWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID, legacy)
	}, true)
	c.testLoggerEndpoint("/requests", func(legacy bool) error {
		return monitor.TryLogRequestsWithParams(c.cfg.APIBaseURL, c.cfg.MonitorAPIKey, c.cfg.MonitorUserID, legacy)
	}, false)
}

func (c *Client) testLoggerEndpoint(endpoint string, testEndpoint func(legacy bool) error, legacy bool) {
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
