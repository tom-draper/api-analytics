package usage

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Common errors
var (
	ErrPostgresURLNotSet = fmt.Errorf("POSTGRES_URL environment variable not set")
)

// Time intervals for queries
const (
	Hourly  = "1 hour"
	Daily   = "24 hours"
	Weekly  = "7 days"
	Monthly = "30 days"
)

var printer = message.NewPrinter(language.English)

// UserCount represents a count of items per user
type UserCount struct {
	APIKey string
	Count  int
}

// Display prints the user count
func (u UserCount) Display() {
	printer.Printf("%s: %d\n", u.APIKey, u.Count)
}

// DisplayUserCounts prints a list of user counts
func DisplayUserCounts(counts []UserCount) {
	for _, c := range counts {
		c.Display()
	}
}

// UserRow represents a user record
type UserRow struct {
	UserID    string    `json:"user_id"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a user with request statistics
type User struct {
	APIKey         string    `json:"api_key"`
	TotalRequests  int       `json:"total_requests"`
	DailyRequests  int       `json:"daily_requests"`
	WeeklyRequests int       `json:"weekly_requests"`
	CreatedAt      time.Time `json:"created_at"`
}

// Display prints the user with request stats
func (user User) Display(rank int) {
	var colorPrintf func(format string, a ...interface{})
	if user.DailyRequests == 0 && user.WeeklyRequests == 0 {
		colorPrintf = color.Red
	} else if user.DailyRequests == 0 || user.WeeklyRequests == 0 {
		colorPrintf = color.Yellow
	} else {
		colorPrintf = color.Green
	}
	colorPrintf("[%d] %s %d (+%d / +%d) %s\n", rank, user.APIKey, user.TotalRequests, user.DailyRequests, user.WeeklyRequests, user.CreatedAt.Format("2006-01-02 15:04:05"))
}

// DisplayUsers prints a list of users
func DisplayUsers(users []User) {
	for i, user := range users {
		user.Display(i)
	}
}

// UserTime represents a user with a time-related field
type UserTime struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	Days      string    `json:"days"`
}

// Display prints the user with time information
func (user UserTime) Display(rank int) {
	format := "[%d] %s %s (%s)\n"
	timeFormat := "2006-01-02 15:04:05"
	daysSince := time.Since(user.CreatedAt)

	switch {
	case daysSince > time.Hour*24*30*6:
		color.Red(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	case daysSince > time.Hour*24*30*3:
		color.Yellow(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	default:
		fmt.Printf(format, rank, user.APIKey, user.CreatedAt.Format(timeFormat), user.Days)
	}
}

// DisplayUserTimes prints a list of users with time info
func DisplayUserTimes(users []UserTime) {
	for i, user := range users {
		user.Display(i)
	}
}

// MonitorRow represents a monitor record
type MonitorRow struct {
	APIKey    string    `json:"api_key"`
	URL       string    `json:"url"`
	Secure    bool      `json:"secure"`
	Ping      bool      `json:"ping"`
	CreatedAt time.Time `json:"created_at"`
}

// RequestRow represents a request record
type RequestRow struct {
	RequestID    int            `json:"request_id"`
	APIKey       string         `json:"api_key"`
	Path         string         `json:"path"`
	Hostname     interface{}    `json:"hostname"`
	IPAddress    interface{}    `json:"ip_address"`
	Location     interface{}    `json:"location"`
	UserAgentID  interface{}    `json:"user_agent_id"`
	Method       int16          `json:"method"`
	Status       int16          `json:"status"`
	ResponseTime int16          `json:"response_time"`
	Framework    int16          `json:"framework"`
	CreatedAt    time.Time      `json:"created_at"`
}

// RequestsColumnSizes represents the size of each column in the requests table
type RequestsColumnSizes struct {
	RequestID    string `json:"request_id"`
	APIKey       string `json:"api_key"`
	Path         string `json:"path"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	Location     string `json:"location"`
	UserAgent    string `json:"user_agent"`
	Method       string `json:"method"`
	Status       string `json:"status"`
	ResponseTime string `json:"response_time"`
	Framework    string `json:"framework"`
	CreatedAt    string `json:"created_at"`
}

// Display prints the column sizes
func (r RequestsColumnSizes) Display() {
	fmt.Printf("request_id: %s\napi_key: %s\npath: %s\nhostname: %s\nip_address: %s\nlocation: %s\nuser_agent: %s\nmethod: %s\nstatus: %s\nresponse_time: %s\nframework: %s\ncreated_at: %s\n",
		r.RequestID, r.APIKey, r.Path, r.Hostname, r.IPAddress, r.Location, r.UserAgent, r.Method, r.Status, r.ResponseTime, r.Framework, r.CreatedAt)
}

// ColumnValueCount represents a count of values in a column
type ColumnValueCount[T any] struct {
	Value T
	Count int
}
