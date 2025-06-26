package options

import (
	"flag"
	"fmt"
)

type Options struct {
	Email    *bool
	Users    *bool
	Monitors *bool
	Database *bool
	Help     *bool
}

func GetOptions() Options {
	options := Options{
		Email:    flag.Bool("email", false, "email the summary instead of printing to console"),
		Users:    flag.Bool("users", false, "show user account usage"),
		Monitors: flag.Bool("monitors", false, "show monitor usage"),		Database: flag.Bool("database", false, "show database usage"),
		Help:     flag.Bool("help", false, "display help"),
	}
	flag.Parse()
	return options
}

func DisplayHelp() {
	fmt.Printf("Checkup - A command-line tool for checking resource usage.\n\nUsage:\n  checkup [options]\n\nOptions:\n  --email     Email the summary instead of printing to console\n  --users     Show user account usage\n  --monitors  Show monitor usage\n  --database  Show database usage\n  --help      Display help information\n")
}