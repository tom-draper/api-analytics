package monitors

import (
	"context"
	"log"
	"testing"

	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

func TestMonitors(t *testing.T) {
	monitors, err := Monitors(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestDailyMonitors(t *testing.T) {
	monitors, err := DailyMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}

func TestWeeklyMonitors(t *testing.T) {
	monitors, err := WeeklyMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}

func TestMonthlyMonitors(t *testing.T) {
	monitors, err := MonthlyMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}

func TestMonitorsCount(t *testing.T) {
	monitors, err := MonitorsCount(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		t.Error("no monitors found")
	}
}

func TestDailyMonitorsCount(t *testing.T) {
	monitors, err := DailyMonitorsCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		log.Println("no monitors found")
	}
}

func TestWeeklyMonitorsCount(t *testing.T) {
	monitors, err := WeeklyMonitorsCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		log.Println("no monitors found")
	}
}

func TestMonthlyMonitorsCount(t *testing.T) {
	monitors, err := MonthlyMonitorsCount(context.Background())
	if err != nil {
		t.Error(err)
	}
	if monitors == 0 {
		log.Println("no monitors found")
	}
}

func TestUserMonitors(t *testing.T) {
	monitors, err := UserMonitors(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		t.Error("no monitors found")
	}
}

func TestUserCountDisplay(t *testing.T) {
	u := usage.UserCount{
		APIKey: "test-key",
		Count:  10,
	}
	u.Display()
}

func TestDailyUserMonitors(t *testing.T) {
	monitors, err := DailyUserMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}

func TestWeeklyUserMonitors(t *testing.T) {
	monitors, err := WeeklyUserMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}

func TestMonthlyUserMonitors(t *testing.T) {
	monitors, err := MonthlyUserMonitors(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(monitors) == 0 {
		log.Println("no monitors found")
	}
}
