package check

import (
	"context"
	"log"
	"os"

	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/email"
	"github.com/tom-draper/api-analytics/server/tools/checkup/internal/display"
	"github.com/tom-draper/api-analytics/server/tools/usage"
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

	usageClient := usage.NewClient(db)

	usersList, err := usageClient.DailyUsers(ctx)
	if display.HandleError(err) != nil {
		return
	}

	requestsList, err := usageClient.DailyUserRequests(ctx)
	if display.HandleError(err) != nil {
		return
	}

	monitorsList, err := usageClient.DailyUserMonitors(ctx)
	if display.HandleError(err) != nil {
		return
	}

	size, err := usageClient.TableSize(ctx, "requests")
	if display.HandleError(err) != nil {
		return
	}

	connections, err := usageClient.DatabaseConnections(ctx)
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