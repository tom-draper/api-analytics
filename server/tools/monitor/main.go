package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/lib"
)

func buildEmailBody(serviceStatus lib.ServiceStatus, apiTestStatus lib.APITestStatus) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Failure detected at %v\n", time.Now()))

	if !serviceStatus.API {
		body.WriteString("Service api down\n")
	}
	if !serviceStatus.Logger {
		body.WriteString("Service logger down\n")
	}
	if !serviceStatus.Nginx {
		body.WriteString("Service nginx down\n")
	}
	if !serviceStatus.PostgresSQL {
		body.WriteString("Service postgresql down\n")
	}

	if apiTestStatus.NewUser != nil {
		body.WriteString(fmt.Sprintf("Error when creating new user: %s\n", apiTestStatus.NewUser.Error()))
	}
	if apiTestStatus.FetchDashboardData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching dashboard data: %s\n", apiTestStatus.FetchDashboardData.Error()))
	}
	if apiTestStatus.FetchData != nil {
		body.WriteString(fmt.Sprintf("Error when fetching API data: %s\n", apiTestStatus.FetchData.Error()))
	}

	return body.String()
}

func main() {
	serviceStatus := lib.ServiceStatus{
		API:         !lib.ServiceDown("api"),
		Logger:      !lib.ServiceDown("logger"),
		Nginx:       !lib.ServiceDown("nginx"),
		PostgresSQL: !lib.ServiceDown("postgreql"),
	}
	apiTestStatus := lib.APITestStatus{
		NewUser:            lib.TryNewUser(),
		FetchDashboardData: lib.TryFetchDashboardData(),
		FetchData:          lib.TryFetchData(),
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
