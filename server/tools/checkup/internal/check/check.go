package check

import (
	"context"
	"log"
	"os"

	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/display"
	"github.com/tom-draper/api-analytics/server/tools/usage"
)

func EmailCheckup(ctx context.Context) {
	users, err := usage.DailyUsers(context.Background())
	if display.HandleError(err) != nil { return }

	requests, err := usage.DailyUserRequests(context.Background())
	if display.HandleError(err) != nil { return }

	monitors, err := usage.DailyUserMonitors(context.Background())
	if display.HandleError(err) != nil { return }

	size, err := usage.TableSize(context.Background(), "requests")
	if display.HandleError(err) != nil { return }

	connections, err := usage.DatabaseConnections(context.Background())
	if display.HandleError(err) != nil { return }

	client, err := email.NewClientFromEnv()
	if err != nil {
		log.Printf("Error creating email client: %v\n", err)
		return
	}
	address := os.Getenv("EMAIL_ADDRESS")
	body := EmailBody(users, requests, monitors, size, connections)
	client.Send(email.Message{To: []string{address}, From: "", Subject: "API Analytics", Body: body})
}