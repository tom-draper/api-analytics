package usage

import (
	"log"
	"testing"
)

func TestRequests(t *testing.T) {
	requests, err := Requests("")
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestDailyRequests(t *testing.T) {
	requests, err := DailyRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestWeeklyRequests(t *testing.T) {
	requests, err := WeeklyRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestMonthlyRequests(t *testing.T) {
	requests, err := MonthlyRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestRequestsCount(t *testing.T) {
	requests, err := RequestsCount("")
	if err != nil {
		t.Error(err)
	}
	if requests == 0 {
		t.Error("no requests found")
	}
}

func TestDailyRequestsCount(t *testing.T) {
	requests, err := DailyRequestsCount()
	if err != nil {
		t.Error(err)
	}
	if requests == 0 {
		t.Error("no requests found")
	}
}

func TestWeeklyRequestsCount(t *testing.T) {
	requests, err := WeeklyRequestsCount()
	if err != nil {
		t.Error(err)
	}
	if requests == 0 {
		t.Error("no requests found")
	}
}

func TestMonthlyRequestsCount(t *testing.T) {
	requests, err := MonthlyRequestsCount()
	if err != nil {
		t.Error(err)
	}
	if requests == 0 {
		t.Error("no requests found")
	}
}

func TestUserRequests(t *testing.T) {
	requests, err := UserRequests("")
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestUserRequestsCount(t *testing.T) {
	requests, err := DailyUserRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestWeeklyUserRequests(t *testing.T) {
	requests, err := WeeklyUserRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestMonthlyUserRequests(t *testing.T) {
	requests, err := MonthlyUserRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestTopFrameworks(t *testing.T) {
	frameworks, err := TopFrameworks()
	if err != nil {
		t.Error(err)
	}
	if len(frameworks) == 0 {
		t.Error("no frameworks found")
	}
	for _, framework := range frameworks {
		log.Println(framework.Value, framework.Count)
	}
}

func TestTopUserAgents(t *testing.T) {
	userAgents, err := TopUserAgents()
	if err != nil {
		t.Error(err)
	}
	if len(userAgents) == 0 {
		t.Error("no user agents found")
	}
	for _, userAgent := range userAgents {
		log.Println(userAgent.Value, userAgent.Count)
	}
}

func TestTopIPAddresses(t *testing.T) {
	ipAddresses, err := TopIPAddresses()
	if err != nil {
		t.Error(err)
	}
	if len(ipAddresses) == 0 {
		t.Error("no IP addresses found")
	}
	for _, ipAddress := range ipAddresses {
		log.Println(ipAddress.Value, ipAddress.Count)
	}
}

func TestLocations(t *testing.T) {
	locations, err := TopLocations()
	if err != nil {
		t.Error(err)
	}
	if len(locations) == 0 {
		t.Error("no locations found")
	}
	for _, location := range locations {
		log.Println(location.Value, location.Count)
	}
}

func TestAvgResponseTime(t *testing.T) {
	avgResponseTime, err := AvgResponseTime()
	if err != nil {
		t.Error(err)
	}
	if avgResponseTime == 0 {
		t.Error("average reponse time is 0")
	}
	log.Println("avg response time:", avgResponseTime)
}
