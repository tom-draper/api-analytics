package main

import (
	"context"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/tom-draper/api-analytics/server/tools/checkup/cmd/checkup/options"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/check"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/display"
)

func main() {
	opts := options.GetOptions()
	if *opts.Help {
		options.DisplayHelp()
		return
	}

	ctx := context.Background()
	p := message.NewPrinter(language.English)

	if *opts.Email {
		check.EmailCheckup(ctx)
	} else if *opts.Users {
		display.DisplayUsersCheckup(ctx, p)
	} else if *opts.Monitors {
		display.DisplayMonitorsCheckup(ctx, p)
	} else if *opts.Database {
		display.DisplayDatabaseCheckup(ctx, p)
	} else {
		display.DisplayCheckup(ctx, p)
	}
}