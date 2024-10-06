package usage

import (
	"context"
	"testing"
)

func TestUsers(t *testing.T) {
	users, err := Users(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestDailyUsers(t *testing.T) {
	users, err := DailyUsers(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestWeeklyUsers(t *testing.T) {
	users, err := WeeklyUsers(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestMonthlyUsers(t *testing.T) {
	users, err := MonthlyUsers(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestUsersCount(t *testing.T) {
	users, err := UsersCount(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestDailyUsersCount(t *testing.T) {
	users, err := DailyUsersCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestWeeklyUsersCount(t *testing.T) {
	users, err := WeeklyUsersCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestMonthlyUsersCount(t *testing.T) {
	users, err := MonthlyUsersCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if users == 0 {
		t.Error("no users found")
	}
}

func TestTopUsers(t *testing.T) {
	users, err := TopUsers(context.Background(), 5)
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestUnusedUsers(t *testing.T) {
	users, err := UnusedUsers(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}

func TestSinceLastRequestUsers(t *testing.T) {
	users, err := SinceLastRequestUsers(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(users) == 0 {
		t.Error("no users found")
	}
}
