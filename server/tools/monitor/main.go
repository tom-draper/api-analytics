package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/monitor"
)

func buildEmailBody(serviceStatus monitor.ServiceStatus, apiTestStatus monitor.APITestStatus) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Failure detected at %v\n", time.Now()))

	if !serviceStatus.api {
		body.WriteString("Service api down\n")
	}
	if !serviceStatus.logger {
		body.WriteString("Service logger down\n")
	}
	if !serviceStatus.nginx {
		body.WriteString("Service nginx down\n")
	}
	if !serviceStatus.postgresql {
		body.WriteString("Service postgresql down\n")
	}

	if apiTestStatus.newUser != nil {
		body.WriteString(fmt.Sprintf("Error when creating new user: %s\n", apiTestStatus.newUser.Error()))
	}
	if apiTestStatus.fetchDashboardData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching dashboard data: %s\n", apiTestStatus.fetchDashboardData.Error()))
	}
	if apiTestStatus.fetchData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching API data: %s\n", apiTestStatus.fetchData.Error()))
	}

	return body.String()
}

func main() {
	serviceStatus := monitor.ServiceStatus{
		api:        !monitor.ServiceDown("api"),
		logger:     !monitor.ServiceDown("logger"),
		nginx:      !monitor.ServiceDown("nginx"),
		postgresql: !monitor.ServiceDown("postgreql"),
	}
	apiTestStatus := monitor.APITestStatus{
		newUser:            monitor.TryNewUser(),
		fetchDashboardData: monitor.TryFetchDashboardData(),
		fetchData:          monitor.TryFetchData(),
	}
	if serviceStatus.ServiceDown() || apiTestStatus.TestFailed() {
		address := email.GetEmailAddress()
		body := buildEmailBody(serviceStatus, apiTestStatus)
		err := email.SendEmail("Failure detected at API Analytics", body, address)
		if err != nil {
			panic(err)
		}
	}
}
