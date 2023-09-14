package usage

import (
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
	requests, err := UserRequests(0)
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

func TestUserAgents(t *testing.T) {
	_, err := UserAgents()
	if err != nil {
		t.Error(err)
	}
}

func TestIPAddresses(t *testing.T) {
	_, err := IPAddresses()
	if err != nil {
		t.Error(err)
	}
}

func TestLocations(t *testing.T) {
	_, err := Locations()
	if err != nil {
		t.Error(err)
	}
}

func TestAvgResponseTime(t *testing.T) {
	_, err := AvgResponseTime()
	if err != nil {
		t.Error(err)
	}
}
