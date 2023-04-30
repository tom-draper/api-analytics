package usage

import (
	"testing"
)

func TestUsage(t *testing.T) {
	requests, err := WeeklyUsage()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestUsers(t *testing.T) {
	users, err := WeeklyUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestMonitors(t *testing.T) {
	monitors, err := WeeklyMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestTopUsers(t *testing.T) {
	users, err := TopUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestTotalRequests(t *testing.T) {
	requests, err := TotalRequests()
	if err != nil {
		t.Error(err)
	}
	if len(requests) == 0 {
		t.Error("no requests found")
	}
}

func TestDatabaseSize(t *testing.T) {
	size, err := DatabaseSize()
	if err != nil {
		t.Error(err)
	}
	if size == "" {
		t.Error("database size is blank")
	}
}

func TestDatabaseConnections(t *testing.T) {
	connections, err := DatabaseConnections()
	if err != nil {
		t.Error(connections)
	}
	if connections == 0 {
		t.Error("number of active database connections is 0")
	}
}
