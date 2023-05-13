package main

import (
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/monitor"
)

type ServiceStatus struct {
	api        bool
	logger     bool
	nginx      bool
	postgresql bool
}

func (s ServiceStatus) ServiceDown() bool {
	return s.api && s.logger && s.nginx && s.postgresql
}

type APITestStatus struct {
	newUser            error
	fetchDashboardData error
	fetchData          error
}

func (s APITestStatus) TestFailed() bool {
	return s.newUser != nil || s.fetchDashboardData != nil || s.fetchData != nil
}

func main() {
	serviceStatus := ServiceStatus{
		api:        !monitor.ServiceDown("api"),
		logger:     !monitor.ServiceDown("logger"),
		nginx:      !monitor.ServiceDown("nginx"),
		postgresql: !monitor.ServiceDown("postgreql"),
	}
	apiTestStatus := APITestStatus{
		newUser:            monitor.TryNewUser(),
		fetchDashboardData: monitor.TryFetchDashboardData(),
		fetchData:          monitor.TryFetchData(),
	}
	if serviceStatus.ServiceDown() || apiTestStatus.TestFailed() {
		address := email.GetEmailAddress()
		body := email.buildEmailBody(serviceStatus, apiTestStatus)
		err := email.SendEmail("Failure detected at API Analytics", body, address)
		if err != nil {
			panic(err)
		}
	}
}
