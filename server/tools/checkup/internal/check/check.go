package check

import (
	"context"
	"log"
	"os"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/display"
	"github.com/tom-draper/api-analytics/server/tools/usage/monitors"
	"github.com/tom-draper/api-analytics/server/tools/usage/requests"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
	"github.com/tom-draper/api-analytics/server/tools/usage/users"
)

func EmailCheckup(ctx context.Context) {
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Println("POSTGRES_URL environment variable not set")
		return
	}

	db, err := database.New(ctx, dbURL)
	if err != nil {
		log.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	usersList, err := users.DailyUsers(ctx, db)
	if display.HandleError(err) != nil {
		return
	}

	requestsList, err := requests.DailyUserRequests(ctx, db)
	if display.HandleError(err) != nil {
		return
	}

	monitorsList, err := monitors.DailyUserMonitors(ctx, db)
	if display.HandleError(err) != nil {
		return
	}

	size, err := usage.TableSize(ctx, db, "requests")
	if display.HandleError(err) != nil {
		return
	}

	connections, err := usage.DatabaseConnections(ctx, db)
	if display.HandleError(err) != nil {
		return
	}

	client, err := email.NewClientFromEnv()
	if err != nil {
		log.Printf("Error creating email client: %v\n", err)
		return
	}
	address := os.Getenv("EMAIL_ADDRESS")
	body := EmailBody(usersList, requestsList, monitorsList, size, connections)
	client.Send(email.Message{To: []string{address}, From: "", Subject: "API Analytics", Body: body})
}