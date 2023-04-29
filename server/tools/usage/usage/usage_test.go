package usage

import (
	"testing"
)

func TestUsage(t *testing.T) {
	requests, err := WeeklyUsage()
	if err != nil {
		t.Error(err)
	}
}

func TestUsers(t *testing.T) {
	users, err := WeeklyUsers()
	if err != nil {
		t.Error(err)
	}
}

func TestMonitors(t *testing.T) {
	monitors, err := WeeklyMonitors()
	if err != nil {
		t.Error(err)
	}
}

func TestTopUsers(t *testing.T) {
	users, err := TopUsers()
	if err != nil {
		t.Error(err)
	}
}

func TestTotalRequests(t *testing.T) {
	requests, err := TotalRequests()
	if err != nil {
		t.Error(err)
	}
}

func TestDatabaseSize(t *testing.T) {
	size, err := DatabaseSize()
	if err != nil {
		t.Error(err)
	}
}
