package main

import (
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/monitor/monitor"
)

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
		body := email.buildEmailBody(serviceStatus, apiTestStatus)
		err := email.SendEmail("Failure detected at API Analytics", body, address)
		if err != nil {
			panic(err)
		}
	}
}
