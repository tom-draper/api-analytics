package options

import (
	"flag"
	"fmt"
	"time"
)

type Options struct {
	Users         *bool
	TargetUser    *string
	Help          *bool
	RequestsLimit *int
	UserExpiry    *time.Duration
	Yes           *bool
}

func GetOptions() Options {
	options := Options{
		Users:         flag.Bool("users", false, "Delete expired users"),
		TargetUser:    flag.String("target-user", "", "Specify an API key for account deletion"),
		Help:          flag.Bool("help", false, "Display help"),
		RequestsLimit: flag.Int("requests-limit", 1_500_000, "Maximum number of requests per user before old requests are deleted"),
		UserExpiry:    flag.Duration("user-expiry", time.Hour*24*30*6, "Duration after which unused or retired users are deleted"),
		Yes:           flag.Bool("yes", false, "Skip confirmation prompts; required to actually delete users in batch (non-interactive) mode"),
	}
	flag.Parse()
	return options
}

func DisplayHelp() {
	fmt.Printf("Cleanup - A command-line tool to delete expired users and requests.\n\nUsage:\n  cleanup [options]\n\nOptions:\n")
	flag.PrintDefaults()
}
