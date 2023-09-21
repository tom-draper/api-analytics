package usage

import (
	"testing"
)

func TestMonitors(t *testing.T) {
	monitors, err := Monitors("")
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestDailyMonitors(t *testing.T) {
	monitors, err := DailyMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestWeeklyMonitors(t *testing.T) {
	monitors, err := WeeklyMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestMonthlyMonitors(t *testing.T) {
	monitors, err := MonthlyMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestMonitorsCount(t *testing.T) {
	monitors, err := MonitorsCount("")
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		t.Error("no monitors found")
	}
}

func TestDailyMonitorsCount(t *testing.T) {
	monitors, err := DailyMonitorsCount()
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		t.Error("no monitors found")
	}
}

func TestWeeklyMonitorsCount(t *testing.T) {
	monitors, err := WeeklyMonitorsCount()
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		t.Error("no monitors found")
	}
}

func TestMonthlyMonitorsCount(t *testing.T) {
	monitors, err := MonthlyMonitorsCount()
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		t.Error("no monitors found")
	}
}

func TestUserMonitors(t *testing.T) {
	monitors, err := UserMonitors("")
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestUserMonitorsCount(t *testing.T) {
	monitors, err := DailyUserMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestWeeklyUserMonitors(t *testing.T) {
	monitors, err := WeeklyUserMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestMonthlyUserMonitors(t *testing.T) {
	monitors, err := MonthlyUserMonitors()
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}
