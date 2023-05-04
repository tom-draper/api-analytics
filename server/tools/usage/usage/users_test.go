package usage

import (
	"testing"
)

func TestUsers(t *testing.T) {
	users, err := Users(0)
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestDailyUsers(t *testing.T) {
	users, err := DailyUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestWeeklyUsers(t *testing.T) {
	users, err := WeeklyUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestMonthlyUsers(t *testing.T) {
	users, err := MonthlyUsers()
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestUsersCount(t *testing.T) {
	users, err := UsersCount(0)
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestDailyUsersCount(t *testing.T) {
	users, err := DailyUsersCount()
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestWeeklyUsersCount(t *testing.T) {
	users, err := WeeklyUsersCount()
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestMonthlyUsersCount(t *testing.T) {
	users, err := MonthlyUsersCount()
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestTopUsers(t *testing.T) {
	users, err := TopUsers(5)
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}
