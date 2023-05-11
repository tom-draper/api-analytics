package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"

	"github.com/joho/godotenv"
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
		api:        !ServiceDown("api"),
		logger:     !ServiceDown("logger"),
		nginx:      !ServiceDown("nginx"),
		postgresql: !ServiceDown("postgreql"),
	}
	apiTestStatus := APITestStatus{
		newUser:            TryNewUser(),
		fetchDashboardData: TryFetchDashboardData(),
		fetchData:          TryFetchData(),
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
